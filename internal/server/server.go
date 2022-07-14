package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	gorillamux "github.com/gorilla/mux"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/newrelic/go-agent/v3/integrations/nrgorilla"
	"github.com/newrelic/go-agent/v3/integrations/nrgrpc"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/odpf/salt/common"
	commonv1 "go.buf.build/odpf/gw/odpf/proton/odpf/common/v1"
	entropyv1beta1 "go.buf.build/odpf/gwv/odpf/proton/odpf/entropy/v1beta1"
	"go.opencensus.io/plugin/ocgrpc"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	handlersv1 "github.com/odpf/entropy/internal/server/v1"
	"github.com/odpf/entropy/pkg/version"
)

const defaultGracePeriod = 5 * time.Second

func Serve(ctx context.Context, addr string, nrApp *newrelic.Application, logger *zap.Logger, resourceSvc handlersv1.ResourceService) error {
	grpcOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(logger),
			nrgrpc.UnaryServerInterceptor(nrApp),
		)),
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	}
	grpcServer := grpc.NewServer(grpcOpts...)
	rpcHTTPGateway := runtime.NewServeMux()
	reflection.Register(grpcServer)

	commonServiceRPC := common.New(version.GetVersionAndBuildInfo())
	grpcServer.RegisterService(&commonv1.CommonService_ServiceDesc, commonServiceRPC)
	if err := commonv1.RegisterCommonServiceHandlerServer(ctx, rpcHTTPGateway, commonServiceRPC); err != nil {
		return err
	}

	resourceServiceRPC := handlersv1.NewAPIServer(resourceSvc)
	grpcServer.RegisterService(&entropyv1beta1.ResourceService_ServiceDesc, resourceServiceRPC)
	if err := entropyv1beta1.RegisterResourceServiceHandlerServer(ctx, rpcHTTPGateway, resourceServiceRPC); err != nil {
		return err
	}

	httpRouter := gorillamux.NewRouter()
	httpRouter.Use(nrgorilla.Middleware(nrApp))
	httpRouter.PathPrefix("/api/").Handler(http.StripPrefix("/api", rpcHTTPGateway))
	httpRouter.Handle("/ping", http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprintf(wr, "pong")
	}))

	logger.Info("starting server", zap.String("addr", addr))

	httpRouter.Use(
		requestID(),
		withOpenCensus(),
		requestLogger(logger), // nolint
	)
	h := grpcHandlerFunc(grpcServer, httpRouter)
	return gracefulServe(ctx, logger, defaultGracePeriod, addr, h)
}

func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}

func gracefulServe(ctx context.Context, lg *zap.Logger, gracePeriod time.Duration, addr string, h http.Handler) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: h,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), gracePeriod)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil { // nolint
			lg.Error("graceful shutdown failed", zap.Error(err))
			return
		}
	}()

	if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
