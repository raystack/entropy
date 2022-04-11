package server

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
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/odpf/salt/common"
	"github.com/odpf/salt/server"
	commonv1 "go.buf.build/odpf/gw/odpf/proton/odpf/common/v1"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	handlersv1 "github.com/odpf/entropy/internal/server/v1"
	"github.com/odpf/entropy/pkg/module"
	"github.com/odpf/entropy/pkg/provider"
	"github.com/odpf/entropy/pkg/resource"
	"github.com/odpf/entropy/pkg/version"
)

type Config struct {
	Host string `mapstructure:"host" default:""`
	Port int    `mapstructure:"port" default:"8080"`
}

func (cfg Config) addr() string { return fmt.Sprintf("%s:%d", cfg.Host, cfg.Port) }

func Serve(ctx context.Context, cfg Config, logger *zap.Logger, nr *newrelic.Application,
	resourceSvc resource.ServiceInterface, moduleSvc module.ServiceInterface, providerSvc provider.ServiceInterface,
) error {
	serverCfg := server.Config{
		Host: cfg.Host,
		Port: cfg.Port,
	}
	grpcOpts := server.WithMuxGRPCServerOptions(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		grpc_recovery.UnaryServerInterceptor(),
		grpc_ctxtags.UnaryServerInterceptor(),
		grpc_zap.UnaryServerInterceptor(logger),
		nrgrpc.UnaryServerInterceptor(nr),
	)))

	muxServer, err := server.NewMux(serverCfg, grpcOpts)
	if err != nil {
		return err
	}

	gw, err := server.NewGateway(serverCfg.Host, serverCfg.Port)
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
	muxServer.RegisterHandler("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "pong")
	}))

	muxServer.RegisterService(
		&commonv1.CommonService_ServiceDesc,
		common.New(version.GetVersionAndBuildInfo()),
	)
	muxServer.RegisterService(
		&entropyv1beta1.ResourceService_ServiceDesc,
		handlersv1.NewApiServer(resourceSvc, moduleSvc, providerSvc),
	)

	muxServer.RegisterService(
		&entropyv1beta1.ProviderService_ServiceDesc,
		handlersv1.NewApiServer(resourceSvc, moduleSvc, providerSvc),
	)

	logger.Info("starting server", zap.String("addr", cfg.addr()))
	return gracefulServe(ctx, muxServer)
}

func gracefulServe(ctx context.Context, mux *server.MuxServer) error {
	const gracefulTimeout = 30 * time.Second

	serverErrorChan := make(chan error)
	go func() {
		serverErrorChan <- mux.Serve()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), gracefulTimeout)
		defer shutdownCancel()
		mux.Shutdown(shutdownCtx)
		return nil

	case serverError := <-serverErrorChan:
		return serverError
	}
}
