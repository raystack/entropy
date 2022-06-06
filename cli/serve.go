package cli

import (
	"context"
	"reflect"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/odpf/entropy/core"
	"github.com/odpf/entropy/core/module"
	entropyserver "github.com/odpf/entropy/internal/server"
	"github.com/odpf/entropy/internal/store/mongodb"
	"github.com/odpf/entropy/modules/firehose"
	"github.com/odpf/entropy/modules/kubernetes"
	"github.com/odpf/entropy/pkg/logger"
	"github.com/odpf/entropy/pkg/metric"
)

func cmdServe() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serve",
		Short:   "Start gRPC & HTTP servers",
		Aliases: []string{"server", "start"},
		Annotations: map[string]string{
			"group:other": "server",
		},
	}

	var migrate, noSync bool
	cmd.Flags().BoolVar(&migrate, "migrate", false, "Run migrations before starting")
	cmd.Flags().BoolVar(&noSync, "no-sync", false, "Disable sync-loop")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig(cmd)
		if err != nil {
			return err
		}

		if migrate {
			if migrateErr := runMigrations(cmd.Context(), cfg.DB); migrateErr != nil {
				return migrateErr
			}
		}

		return runServer(cmd.Context(), cfg, noSync)
	}
	return cmd
}

func runServer(baseCtx context.Context, c Config, disableSync bool) error {
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	nr, err := metric.New(&c.NewRelic)
	if err != nil {
		return err
	}

	zapLog, err := logger.New(&c.Log)
	if err != nil {
		return err
	}

	moduleRegistry := setupRegistry(zapLog,
		kubernetes.Module,
		firehose.Module,
	)

	mongoStore, err := mongodb.Connect(c.DB)
	if err != nil {
		return err
	}
	resourceStore := mongodb.NewResourceStore(mongoStore)
	resourceService := core.New(resourceStore, moduleRegistry, time.Now, zapLog)

	if !disableSync {
		go func() {
			defer cancel()

			if err := resourceService.RunSync(ctx); err != nil {
				zapLog.Error("sync-loop exited with error", zap.Error(err))
			} else {
				zapLog.Info("sync-loop exited gracefully")
			}
		}()
	}

	return entropyserver.Serve(ctx, c.Service, zapLog, nr, resourceService)
}

func setupRegistry(logger *zap.Logger, modules ...module.Descriptor) *module.Registry {
	moduleRegistry := module.NewRegistry()
	for _, desc := range modules {
		if err := moduleRegistry.Register(desc); err != nil {
			logger.Fatal("failed to register module",
				zap.String("module_kind", desc.Kind),
				zap.String("go_type", reflect.TypeOf(desc.Module).String()),
				zap.Error(err),
			)
		}
	}
	return moduleRegistry
}
