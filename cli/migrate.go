package cli

import (
	"context"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/goto/entropy/pkg/logger"
)

func cmdMigrate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run DB migrations",
		Annotations: map[string]string{
			"group:other": "server",
		},
	}

	cmd.RunE = handleErr(func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig(cmd)
		if err != nil {
			return err
		}

		zapLog, err := logger.New(&cfg.Log)
		if err != nil {
			return err
		}

		return runMigrations(cmd.Context(), zapLog, cfg)
	})

	return cmd
}

func runMigrations(ctx context.Context, zapLog *zap.Logger, cfg Config) error {
	store := setupStorage(zapLog, cfg.PGConnStr, cfg.Syncer)
	return store.Migrate(ctx)
}
