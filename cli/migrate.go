package cli

import (
	"github.com/spf13/cobra"

	"github.com/odpf/entropy/internal/store/mongodb"
)

const (
	resourceRepoName = "resources"
	providerRepoName = "providers"
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

	resourceRepository := mongodb.NewResourceRepository(mongoStore.Collection(resourceRepoName))
	if err = resourceRepository.Migrate(); err != nil {
		return err
	}

	providerRepository := mongodb.NewProviderRepository(mongoStore.Collection(providerRepoName))
	if err = providerRepository.Migrate(); err != nil {
		return err
	}

	return nil
}
