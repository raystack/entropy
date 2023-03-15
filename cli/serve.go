package cli

import (
	"context"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/goto/entropy/core"
	"github.com/goto/entropy/core/module"
	entropyserver "github.com/goto/entropy/internal/server"
	"github.com/goto/entropy/internal/store/postgres"
	"github.com/goto/entropy/modules"
	"github.com/goto/entropy/modules/firehose"
	"github.com/goto/entropy/modules/kubernetes"
	"github.com/goto/entropy/pkg/logger"
	"github.com/goto/entropy/pkg/telemetry"
	"github.com/goto/entropy/pkg/worker"
	"github.com/goto/entropy/pkg/worker/pgq"
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

	store := setupStorage(zapLog, cfg.PGConnStr)
	moduleService := module.NewService(setupRegistry(zapLog), store)
	resourceService := core.New(store, moduleService, asyncWorker, time.Now, zapLog)

	if err := asyncWorker.Register(core.JobKindSyncResource, resourceService.HandleSyncJob); err != nil {
		return err
	}

	if err := asyncWorker.Register(core.JobKindScheduledSyncResource, resourceService.HandleSyncJob); err != nil {
		return err
	}

	return entropyserver.Serve(ctx,
		cfg.Service.httpAddr(), cfg.Service.grpcAddr(),
		nrApp, zapLog, resourceService, moduleService,
	)
}

func setupRegistry(logger *zap.Logger) module.Registry {
	supported := []module.Descriptor{
		kubernetes.Module,
		firehose.Module,
	}

	registry := &modules.Registry{}
	for _, desc := range supported {
		if err := registry.Register(desc); err != nil {
			logger.Fatal("failed to register module",
				zap.String("module_kind", desc.Kind),
				zap.Error(err),
			)
		}
	}
	return registry
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
