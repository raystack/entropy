package cmd

import (
	"github.com/odpf/entropy/app"
	"github.com/odpf/entropy/domain"
	"github.com/odpf/salt/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "migrate",
		Short: "Run DB migrations",
		RunE:  migrate,
	})
}

func migrate(cmd *cobra.Command, args []string) error {
	var c domain.Config
	l := config.NewLoader(config.WithPath("./"))
	err := l.Load(&c)
	if err != nil {
		return err
	}
	return app.RunMigrations(&c)
}
