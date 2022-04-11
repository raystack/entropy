package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/newrelic/go-agent/v3/integrations/nrgrpc"
	"github.com/odpf/salt/common"
	"github.com/odpf/salt/server"
	"github.com/spf13/cobra"
	commonv1 "go.buf.build/odpf/gw/odpf/proton/odpf/common/v1"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	handlersv1 "github.com/odpf/entropy/api/handlers/v1"
	"github.com/odpf/entropy/modules/firehose"
	"github.com/odpf/entropy/modules/log"
	"github.com/odpf/entropy/pkg/logger"
	"github.com/odpf/entropy/pkg/metric"
	"github.com/odpf/entropy/pkg/module"
	"github.com/odpf/entropy/pkg/provider"
	"github.com/odpf/entropy/pkg/resource"
	"github.com/odpf/entropy/pkg/version"
	"github.com/odpf/entropy/store"
	"github.com/odpf/entropy/store/inmemory"
	"github.com/odpf/entropy/store/mongodb"
)

func cmdServe() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serve",
		Short:   "Start gRPC & HTTP servers",
		Aliases: []string{"server", "start"},
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig(cmd)
		if err != nil {
			return err
		}
		return runServer(cfg)
	}
	return cmd
}

func runServer(c Config) error {
	ctx, cancelFunc := context.WithCancel(
		server.HandleSignals(context.Background()),
	)
	defer cancelFunc()

	nr, err := metric.New(&c.NewRelic)
	if err != nil {
		return err
	}

	loggerInstance, err := logger.New(&c.Log)
	if err != nil {
		return err
	}

	mongoStore, err := mongodb.New(&c.DB)
	if err != nil {
		return err
	}

	resourceRepository := mongodb.NewResourceRepository(
		mongoStore.Collection(store.ResourceRepositoryName),
	)

	providerRepository := mongodb.NewProviderRepository(
		mongoStore.Collection(store.ProviderRepositoryName),
	)

	moduleRepository := inmemory.NewModuleRepository()

	err = moduleRepository.Register(log.New(loggerInstance))
	if err != nil {
		return err
	}

	err = moduleRepository.Register(firehose.New(providerRepository))
	if err != nil {
		return err
	}

	resourceService := resource.NewService(resourceRepository)
	moduleService := module.NewService(moduleRepository)
	providerService := provider.NewService(providerRepository)

	muxServer, err := server.NewMux(server.Config{
		Port: c.Service.Port,
		Host: c.Service.Host,
	}, server.WithMuxGRPCServerOptions(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		grpc_recovery.UnaryServerInterceptor(),
		grpc_ctxtags.UnaryServerInterceptor(),
		grpc_zap.UnaryServerInterceptor(loggerInstance),
		nrgrpc.UnaryServerInterceptor(nr),
	))))

	if err != nil {
		return err
	}

	gw, err := server.NewGateway(c.Service.Host, c.Service.Port)
	if err != nil {
		return err
	}

	err = gw.RegisterHandler(ctx, commonv1.RegisterCommonServiceHandlerFromEndpoint)
	if err != nil {
		return err
	}

	err = gw.RegisterHandler(ctx, entropyv1beta1.RegisterResourceServiceHandlerFromEndpoint)
	if err != nil {
		return err
	}

	err = gw.RegisterHandler(ctx, entropyv1beta1.RegisterProviderServiceHandlerFromEndpoint)
	if err != nil {
		return err
	}

	muxServer.SetGateway("/api", gw)

	muxServer.RegisterService(
		&commonv1.CommonService_ServiceDesc,
		common.New(version.GetVersionAndBuildInfo()),
	)
	muxServer.RegisterService(
		&entropyv1beta1.ResourceService_ServiceDesc,
		handlersv1.NewApiServer(resourceService, moduleService, providerService),
	)

	muxServer.RegisterService(
		&entropyv1beta1.ProviderService_ServiceDesc,
		handlersv1.NewApiServer(resourceService, moduleService, providerService),
	)

	muxServer.RegisterHandler("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "pong")
	}))

	loggerInstance.Info("starting server", zap.String("host", c.Service.Host), zap.Int("port", c.Service.Port))

	serverErrorChan := make(chan error)

	go func() {
		serverErrorChan <- muxServer.Serve()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*30)
		defer shutdownCancel()
		muxServer.Shutdown(shutdownCtx)
	case serverError := <-serverErrorChan:
		return serverError
	}
	return nil
}
