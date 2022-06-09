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
	"github.com/odpf/entropy/internal/store/postgres"
	"github.com/odpf/entropy/modules/firehose"
	"github.com/odpf/entropy/modules/kubernetes"
	"github.com/odpf/entropy/pkg/logger"
	"github.com/odpf/entropy/pkg/metric"
	"github.com/odpf/entropy/pkg/worker"
	"github.com/odpf/entropy/pkg/worker/pgq"
)

func cmdServe() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serve",
		Short:   "Start gRPC & HTTP servers and optionally workers",
		Aliases: []string{"server", "start"},
		Annotations: map[string]string{
			"group:other": "server",
		},
	}

	var migrate bool
	cmd.Flags().BoolVar(&migrate, "migrate", false, "Run migrations before starting")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig(cmd)
		if err != nil {
			return err
		}

		zapLog, err := logger.New(&cfg.Log)
		if err != nil {
			return err
		}

		if migrate {
			if migrateErr := runMigrations(cmd.Context(), zapLog, cfg); migrateErr != nil {
				return migrateErr
			}
		}

		return runServer(cmd.Context(), zapLog, cfg)
	}
	return cmd
}

func runServer(baseCtx context.Context, zapLog *zap.Logger, cfg Config) error {
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	nr, err := metric.New(&cfg.NewRelic)
	if err != nil {
		return err
	}

	modules := []module.Descriptor{
		kubernetes.Module,
		firehose.Module,
	}

	store := setupStorage(zapLog, cfg.PGConnStr)
	asyncWorker := setupWorker(zapLog, cfg.Worker)
	moduleRegistry := setupRegistry(zapLog, modules...)

	service := core.New(store, moduleRegistry, asyncWorker, time.Now, zapLog)

	if err := asyncWorker.Register(core.JobKindSyncResource, service.HandleSyncJob); err != nil {
		return err
	}

	return entropyserver.Serve(ctx, cfg.Service, zapLog, nr, service)
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

func setupWorker(logger *zap.Logger, conf workerConf) *worker.Worker {
	pgQueue, err := pgq.Open(conf.QueueSpec, conf.QueueName)
	if err != nil {
		logger.Fatal("failed to init postgres job-queue", zap.Error(err))
	}

	asyncWorker, err := worker.New(pgQueue)
	if err != nil {
		logger.Fatal("failed to init worker instance", zap.Error(err))
	}

	return asyncWorker
}

func setupStorage(logger *zap.Logger, pgConStr string) *postgres.Store {
	store, err := postgres.Open(pgConStr)
	if err != nil {
		logger.Fatal("failed to connect to Postgres database",
			zap.Error(err), zap.String("conn_str", pgConStr))
	}
	return store
}
