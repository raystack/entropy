package app

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
	commonv1 "go.buf.build/odpf/gw/odpf/proton/odpf/common/v1"

	handlersv1 "github.com/odpf/entropy/api/handlers/v1"
	"github.com/odpf/entropy/domain"
	"github.com/odpf/entropy/logger"
	"github.com/odpf/entropy/metric"
	"github.com/odpf/entropy/service"
	"github.com/odpf/entropy/store"
	"github.com/odpf/entropy/version"
	"github.com/odpf/salt/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// RunServer runs the application server
func RunServer(c *domain.Config) error {
	ctx, cancelFunc := context.WithCancel(server.HandleSignals(context.Background()))
	defer cancelFunc()

	nr, err := metric.New(&c.NewRelic)
	if err != nil {
		return err
	}

	logger, err := logger.New(&c.Log)
	if err != nil {
		return err
	}

	store, err := store.New(&c.DB)
	if err != nil {
		return err
	}

	serviceContainer, err := service.Init(store)
	if err != nil {
		return err
	}

	_ = handlersv1.NewApiServer(serviceContainer)

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
	muxServer.SetGateway("/api", gw)

	muxServer.RegisterHandler("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong")
	}))

	muxServer.RegisterService(
		&commonv1.CommonService_ServiceDesc,
		common.New(version.GetVersionAndBuildInfo()),
	)

	logger.Info("starting server", zap.String("host", c.Service.Host), zap.Int("port", c.Service.Port))

	errorChan := make(chan error)

	go func() {
		errorChan <- muxServer.Serve()
	}()

	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*30)
	defer shutdownCancel()

	muxServer.Shutdown(shutdownCtx)
	return <-errorChan
}

func RunMigrations(c *domain.Config) error {
	store, err := store.New(&c.DB)
	if err != nil {
		return err
	}

	services, err := service.Init(store)
	if err != nil {
		return err
	}

	return services.MigrateAll(store)
}
