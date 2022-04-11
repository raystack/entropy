package main

import (
	"context"

	"github.com/odpf/salt/server"
	"github.com/spf13/cobra"

	entropyserver "github.com/odpf/entropy/internal/server"
	"github.com/odpf/entropy/module"
	"github.com/odpf/entropy/modules/firehose"
	"github.com/odpf/entropy/modules/log"
	"github.com/odpf/entropy/pkg/logger"
	"github.com/odpf/entropy/pkg/metric"
	"github.com/odpf/entropy/provider"
	"github.com/odpf/entropy/resource"
	"github.com/odpf/entropy/store/inmemory"
	"github.com/odpf/entropy/store/mongodb"
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
	ctx, cancelFunc := context.WithCancel(
		server.HandleSignals(context.Background()),
	)
	defer cancelFunc()

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

	resourceService := resource.NewService(resourceRepository)
	moduleService := module.NewService(moduleRepository)
	providerService := provider.NewService(providerRepository)

	return entropyserver.Serve(ctx, c.Service, loggerInstance, nr, resourceService, moduleService, providerService)
}
