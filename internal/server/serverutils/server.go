package serverutils

import (
	"context"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"

	"github.com/odpf/entropy/pkg/errors"
)

const grpcRequestMarker = "application/grpc"

// MultiplexHTTPAndGRPC returns an HTTP handler with grpcServer and httpHandler multiplexed.
// Dispatch occurs based on the content-type=application/grpc header.
func MultiplexHTTPAndGRPC(grpcServer *grpc.Server, httpServer http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), grpcRequestMarker) {
			grpcServer.ServeHTTP(w, r)
		} else {
			httpServer.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}

// GracefulServe starts an HTTP server with given http handler at addr and blocks until the
// server exits. A graceful shutdown is performed when the context is cancelled.
func GracefulServe(ctx context.Context, lg *zap.Logger, gracePeriod time.Duration, addr string, h http.Handler) error {
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
