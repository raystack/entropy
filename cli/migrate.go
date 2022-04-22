package cli

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/odpf/entropy/internal/store/mongodb"
)

func cmdMigrate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run DB migrations",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		cfg, err := loadConfig(cmd)
		if err != nil {
			return err
		}

		return runMigrations(ctx, cfg.DB)
	}

	return cmd
}

func runMigrations(ctx context.Context, cfg mongodb.Config) error {
	mongoStore, err := mongodb.Connect(cfg)
	if err != nil {
		return err
	}

	resourceRepository := mongodb.NewResourceRepository(mongoStore)
	if err = resourceRepository.Migrate(ctx); err != nil {
		return err
	}

	providerRepository := mongodb.NewProviderRepository(mongoStore)
	if err = providerRepository.Migrate(ctx); err != nil {
		return err
	}

	return nil
}
