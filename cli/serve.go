package cli

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/odpf/entropy/core"
	"github.com/odpf/entropy/core/module"
	entropyserver "github.com/odpf/entropy/internal/server"
	"github.com/odpf/entropy/internal/store/mongodb"
	"github.com/odpf/entropy/pkg/logger"
	"github.com/odpf/entropy/pkg/metric"
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
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	nr, err := metric.New(&c.NewRelic)
	if err != nil {
		return err
	}

	zapLog, err := logger.New(&c.Log)
	if err != nil {
		return err
	}

	mongoStore, err := mongodb.Connect(c.DB)
	if err != nil {
		return err
	}
	resourceRepository := mongodb.NewResourceRepository(mongoStore)

	moduleRegistry := module.NewRegistry()
	resourceService := core.New(resourceRepository, moduleRegistry, time.Now, zapLog)

	go func() {
		if err := resourceService.RunSync(ctx); err != nil {
			zapLog.Error("sync-loop exited with error", zap.Error(err))
		}
		zapLog.Info("sync-loop exited gracefully")
	}()

	return entropyserver.Serve(ctx, c.Service, zapLog, nr, resourceService)
}
