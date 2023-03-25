package cli

import (
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

		store := setupStorage(zapLog, cfg.PGConnStr, cfg.Syncer)
		moduleService := module.NewService(setupRegistry(zapLog), store)
		resourceService := core.New(store, moduleService, time.Now, zapLog)

		if migrate {
			if migrateErr := runMigrations(cmd.Context(), zapLog, cfg); migrateErr != nil {
				return migrateErr
			}
		}

		if spawnWorker {
			go func() {
				if runErr := resourceService.RunSyncer(cmd.Context(), cfg.Syncer.SyncInterval); runErr != nil {
					zapLog.Error("syncer exited with error", zap.Error(err))
				}
			}()
		}

		return entropyserver.Serve(cmd.Context(),
			cfg.Service.httpAddr(), cfg.Service.grpcAddr(),
			nrApp, zapLog, resourceService, moduleService,
		)
	})

	return cmd
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

func setupStorage(logger *zap.Logger, pgConStr string, syncCfg syncerConf) *postgres.Store {
	store, err := postgres.Open(pgConStr, syncCfg.RefreshInterval, syncCfg.ExtendLockBy)
	if err != nil {
		logger.Fatal("failed to connect to Postgres database",
			zap.Error(err), zap.String("conn_str", pgConStr))
	}
	return store
}
