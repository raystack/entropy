package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/odpf/entropy/internal/store/mongodb"
)

func cmdMigrate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run DB migrations",
		Annotations: map[string]string{
			"group:other": "server",
		},
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig(cmd)
		if err != nil {
			return err
		}

		return runMigrations(cmd.Context(), cfg.DB)
	}

	return cmd
}

func runMigrations(ctx context.Context, cfg mongodb.Config) error {
	mongoStore, err := mongodb.Connect(cfg)
	if err != nil {
		return err
	}

	resourceRepository := mongodb.NewResourceRepository(mongoStore)
	if err := resourceRepository.Migrate(ctx); err != nil {
		return err
	}

	return nil
}
