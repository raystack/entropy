package app

import (
	"context"
	"fmt"
	"github.com/odpf/entropy/modules/log"
	"github.com/odpf/entropy/pkg/module"
	"github.com/odpf/entropy/pkg/resource"
	"github.com/odpf/entropy/store"
	"github.com/odpf/entropy/store/inmemory"
	"github.com/odpf/entropy/store/mongodb"
	"net/http"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"

	"github.com/newrelic/go-agent/v3/integrations/nrgrpc"

	"github.com/odpf/salt/common"
	commonv1 "go.buf.build/odpf/gw/odpf/proton/odpf/common/v1"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"

	handlersv1 "github.com/odpf/entropy/api/handlers/v1"
	"github.com/odpf/entropy/pkg/logger"
	"github.com/odpf/entropy/pkg/metric"
	"github.com/odpf/entropy/pkg/version"
	"github.com/odpf/salt/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Config contains the application configuration
type Config struct {
	Service  ServiceConfig         `mapstructure:"service"`
	DB       mongodb.DBConfig      `mapstructure:"db"`
	NewRelic metric.NewRelicConfig `mapstructure:"newrelic"`
	Log      logger.LogConfig      `mapstructure:"log"`
}

type ServiceConfig struct {
	Port int    `mapstructure:"port" default:"8080"`
	Host string `mapstructure:"host" default:""`
}

// RunServer runs the application server
func RunServer(c *Config) error {
	ctx, cancelFunc := context.WithCancel(
		server.HandleSignals(context.Background()),
	)
	defer cancelFunc()

	nr, err := metric.New(&c.NewRelic)
	if err != nil {
		return err
	}

	logger, err := logger.New(&c.Log)
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

	moduleRepository := inmemory.NewModuleRepository()
	err = moduleRepository.Register(&log.Module{})
	if err != nil {
		return err
	}

	resourceService := resource.NewService(resourceRepository)
	moduleService := module.NewService(resourceRepository, moduleRepository)

	muxServer, err := server.NewMux(server.Config{
		Port: c.Service.Port,
		Host: c.Service.Host,
	}, server.WithMuxGRPCServerOptions(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		grpc_recovery.UnaryServerInterceptor(),
		grpc_ctxtags.UnaryServerInterceptor(),
		grpc_zap.UnaryServerInterceptor(logger),
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

	muxServer.SetGateway("/api", gw)

	muxServer.RegisterService(
		&commonv1.CommonService_ServiceDesc,
		common.New(version.GetVersionAndBuildInfo()),
	)
	muxServer.RegisterService(
		&entropyv1beta1.ResourceService_ServiceDesc,
		handlersv1.NewApiServer(resourceService, moduleService),
	)

	muxServer.RegisterHandler("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong")
	}))

	logger.Info("starting server", zap.String("host", c.Service.Host), zap.Int("port", c.Service.Port))

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

func RunMigrations(c *Config) error {
	mongoStore, err := mongodb.New(&c.DB)
	if err != nil {
		return err
	}

	resourceRepository := mongodb.NewResourceRepository(
		mongoStore.Collection(store.ResourceRepositoryName),
	)

	return resourceRepository.Migrate()
}
