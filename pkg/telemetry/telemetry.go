package telemetry

import (
	"context"
	"net/http"
	"net/http/pprof"

	"go.uber.org/zap"
)

type Config struct {
	// Debug sets the bind address for pprof & zpages server.
	Debug string `mapstructure:"debug_addr"`

	// OpenCensus trace & metrics configurations.
	EnableCPU        bool    `mapstructure:"enable_cpu"`
	EnableMemory     bool    `mapstructure:"enable_memory"`
	SamplingFraction float64 `mapstructure:"sampling_fraction"`

	// OpenCensus exporter configurations.
	ServiceName string `mapstructure:"service_name"`

	// NewRelic exporter.
	EnableNewrelic bool   `mapstructure:"enable_newrelic"`
	NewRelicAPIKey string `mapstructure:"newrelic_api_key"`

	// OpenTelemetry Agent exporter.
	EnableOtelAgent  bool   `mapstructure:"enable_otel_agent"`
	OpenTelAgentAddr string `mapstructure:"otel_agent_addr"`
}

// Init initialises OpenCensus based async-telemetry processes and
// returns (i.e., it does not block).
func Init(ctx context.Context, cfg Config, lg *zap.Logger) {
	mux := http.NewServeMux()
	mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	mux.Handle("/debug/pprof/block", pprof.Handler("block"))

	if err := setupOpenCensus(ctx, mux, cfg); err != nil {
		lg.Error("failed to setup OpenCensus", zap.Error(err))
	}

	if cfg.Debug != "" {
		go func() {
			if err := http.ListenAndServe(cfg.Debug, mux); err != nil {
				lg.Error("debug server exited due to error", zap.Error(err))
			}
		}()
	}
}
