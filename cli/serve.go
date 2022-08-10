package cli

import (
	"context"
	"reflect"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/odpf/entropy/core"
	"github.com/odpf/entropy/core/module"
	entropyserver "github.com/odpf/entropy/internal/server"
	"github.com/odpf/entropy/internal/store/postgres"
	"github.com/odpf/entropy/modules/firehose"
	"github.com/odpf/entropy/modules/kubernetes"
	"github.com/odpf/entropy/pkg/logger"
	"github.com/odpf/entropy/pkg/telemetry"
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

	var migrate, spawnWorker bool
	cmd.Flags().BoolVar(&migrate, "migrate", false, "Run migrations before starting")
	cmd.Flags().BoolVar(&spawnWorker, "worker", false, "Run worker threads as well")

	cmd.RunE = handleErr(func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig(cmd)
		if err != nil {
			return err
		}

		zapLog, err := logger.New(&cfg.Log)
		if err != nil {
			return err
		}

		telemetry.Init(cmd.Context(), cfg.Telemetry, zapLog)
		nrApp, err := newrelic.NewApplication(
			newrelic.ConfigAppName(cfg.Telemetry.ServiceName),
			newrelic.ConfigLicense(cfg.Telemetry.NewRelicAPIKey),
		)

		if migrate {
			if migrateErr := runMigrations(cmd.Context(), zapLog, cfg); migrateErr != nil {
				return migrateErr
			}
		}

		asyncWorker := setupWorker(zapLog, cfg.Worker)
		if spawnWorker {
			go func() {
				if runErr := asyncWorker.Run(cmd.Context()); runErr != nil {
					zapLog.Error("worker exited with error", zap.Error(err))
				}
			}()
		}

		return runServer(cmd.Context(), nrApp, zapLog, cfg, asyncWorker)
	})

	return cmd
}

func runServer(baseCtx context.Context, nrApp *newrelic.Application, zapLog *zap.Logger, cfg Config, asyncWorker *worker.Worker) error {
	ctx, cancel := context.WithCancel(baseCtx)
	defer cancel()

	modules := []module.Descriptor{
		kubernetes.Module,
		firehose.Module,
	}

	store := setupStorage(zapLog, cfg.PGConnStr)
	moduleRegistry := setupRegistry(zapLog, modules...)
	resourceService := core.New(store, moduleRegistry, asyncWorker, time.Now, zapLog)

	if err := asyncWorker.Register(core.JobKindSyncResource, resourceService.HandleSyncJob); err != nil {
		return err
	}

	if err := asyncWorker.Register(core.JobKindScheduledSyncResource, resourceService.HandleSyncJob); err != nil {
		return err
	}

	return entropyserver.Serve(ctx, cfg.Service.addr(), nrApp, zapLog, resourceService, moduleRegistry)
}

func setupRegistry(logger *zap.Logger, modules ...module.Descriptor) *module.Registry {
	moduleRegistry := module.NewRegistry(nil)
	for _, desc := range modules {
		if err := moduleRegistry.Register(desc); err != nil {
			logger.Fatal("failed to register module",
				zap.String("module_kind", desc.Kind),
				zap.String("go_type", reflect.TypeOf(desc.DriverFactory).String()),
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

	opts := []worker.Option{
		worker.WithLogger(logger.Named("worker")),
		worker.WithRunConfig(conf.Threads, conf.PollInterval),
	}

	asyncWorker, err := worker.New(pgQueue, opts...)
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
