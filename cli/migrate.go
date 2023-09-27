package cli

import (
	"context"

	"github.com/spf13/cobra"

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

		err = logger.Setup(&cfg.Log)
		if err != nil {
			return err
		}

		return runMigrations(cmd.Context(), cfg)
	})

	return cmd
}

func runMigrations(ctx context.Context, cfg Config) error {
	store := setupStorage(cfg.PGConnStr, cfg.Syncer)
	return store.Migrate(ctx)
}
