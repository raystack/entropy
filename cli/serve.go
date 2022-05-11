package cli

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/odpf/entropy/core/resource"
	entropyserver "github.com/odpf/entropy/internal/server"
	"github.com/odpf/entropy/internal/store/inmemory"
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

	loggerInstance, err := logger.New(&c.Log)
	if err != nil {
		return err
	}

	mongoStore, err := mongodb.Connect(c.DB)
	if err != nil {
		return err
	}

	resourceRepository := mongodb.NewResourceRepository(mongoStore)
	moduleRepository := inmemory.NewModuleRepository()

	resourceService := resource.NewService(resourceRepository, moduleRepository, time.Now)

	return entropyserver.Serve(ctx, c.Service, loggerInstance, nr, resourceService)
}
