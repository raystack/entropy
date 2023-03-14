package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	gorillamux "github.com/gorilla/mux"
	commonv1 "github.com/goto/entropy/proto/gotocompany/common/v1"
	entropyv1beta1 "github.com/goto/entropy/proto/gotocompany/entropy/v1beta1"
	"github.com/goto/salt/mux"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/newrelic/go-agent/v3/integrations/nrgorilla"
	"github.com/newrelic/go-agent/v3/integrations/nrgrpc"
	"github.com/newrelic/go-agent/v3/newrelic"
	"go.opencensus.io/plugin/ocgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"

	modulesv1 "github.com/goto/entropy/internal/server/v1/modules"
	resourcesv1 "github.com/goto/entropy/internal/server/v1/resources"
	"github.com/goto/entropy/pkg/common"
	"github.com/goto/entropy/pkg/version"
)

const defaultGracePeriod = 5 * time.Second

// Serve initialises all the gRPC+HTTP API routes, starts listening for requests at addr, and blocks until server exits.
// Server exits gracefully when context is cancelled.
func Serve(ctx context.Context, addr string, nrApp *newrelic.Application, logger *zap.Logger,
	resourceSvc resourcesv1.ResourceService, moduleSvc modulesv1.ModuleService,
) error {
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
	rpcHTTPGateway := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}))

	reflection.Register(grpcServer)

	commonServiceRPC := common.New(version.GetVersionAndBuildInfo())
	grpcServer.RegisterService(&commonv1.CommonService_ServiceDesc, commonServiceRPC)
	if err := commonv1.RegisterCommonServiceHandlerServer(ctx, rpcHTTPGateway, commonServiceRPC); err != nil {
		return err
	}

	resourceServiceRPC := resourcesv1.NewAPIServer(resourceSvc)
	grpcServer.RegisterService(&entropyv1beta1.ResourceService_ServiceDesc, resourceServiceRPC)
	if err := entropyv1beta1.RegisterResourceServiceHandlerServer(ctx, rpcHTTPGateway, resourceServiceRPC); err != nil {
		return err
	}

	moduleServiceRPC := modulesv1.NewAPIServer(moduleSvc)
	grpcServer.RegisterService(&entropyv1beta1.ModuleService_ServiceDesc, moduleServiceRPC)
	if err := entropyv1beta1.RegisterModuleServiceHandlerServer(ctx, rpcHTTPGateway, moduleServiceRPC); err != nil {
		return err
	}

	httpRouter := gorillamux.NewRouter()
	httpRouter.Use(nrgorilla.Middleware(nrApp))
	httpRouter.PathPrefix("/api/").Handler(http.StripPrefix("/api", rpcHTTPGateway))
	httpRouter.Handle("/ping", http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprintf(wr, "pong")
	}))

	httpRouter.Use(
		requestID(),
		withOpenCensus(),
		requestLogger(logger), // nolint
	)

	logger.Info("starting server", zap.String("addr", addr))
	return mux.Serve(ctx,
		mux.WithHTTPTarget(":8081", &http.Server{
			Handler: httpRouter,
		}),
		mux.WithGRPCTarget(addr, grpcServer),
		mux.WithGracePeriod(defaultGracePeriod),
	)
}
