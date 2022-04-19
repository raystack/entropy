package cli

import (
	"github.com/spf13/cobra"

	"github.com/odpf/entropy/internal/store/mongodb"
)

func cmdMigrate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run DB migrations",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig(cmd)
		if err != nil {
			return err
		}

		return runMigrations(cfg.DB)
	}

	return cmd
}

func runMigrations(cfg mongodb.DBConfig) error {
	mongoStore, err := mongodb.New(&cfg)
	if err != nil {
		return err
	}

	resourceRepository := mongodb.NewResourceRepository(mongoStore)
	if err = resourceRepository.Migrate(); err != nil {
		return err
	}

	providerRepository := mongodb.NewProviderRepository(mongoStore)
	if err = providerRepository.Migrate(); err != nil {
		return err
	}

	return nil
}
