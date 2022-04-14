package cli

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/odpf/entropy/core/provider"
	"github.com/odpf/entropy/core/resource"
	entropyserver "github.com/odpf/entropy/internal/server"
	"github.com/odpf/entropy/internal/store/inmemory"
	"github.com/odpf/entropy/internal/store/mongodb"
	"github.com/odpf/entropy/pkg/logger"
	"github.com/odpf/entropy/pkg/metric"
	"github.com/odpf/entropy/plugins/modules/firehose"
	"github.com/odpf/entropy/plugins/modules/log"
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

	mongoStore, err := mongodb.New(&c.DB)
	if err != nil {
		return err
	}

	resourceRepository := mongodb.NewResourceRepository(mongoStore.Collection(resourceRepoName))
	providerRepository := mongodb.NewProviderRepository(mongoStore.Collection(providerRepoName))

	moduleRepository := inmemory.NewModuleRepository()
	err = moduleRepository.Register(log.New(loggerInstance))
	if err != nil {
		return err
	}

	err = moduleRepository.Register(firehose.New(providerRepository))
	if err != nil {
		return err
	}

	resourceService := resource.NewService(resourceRepository, moduleRepository)
	providerService := provider.NewService(providerRepository)

	return entropyserver.Serve(ctx, c.Service, loggerInstance, nr, resourceService, providerService)
}
