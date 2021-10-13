package cmd

import (
	"github.com/odpf/entropy/app"
	"github.com/odpf/entropy/domain"
	"github.com/odpf/salt/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Run server",
		RunE:  serve,
	})
}

func serve(cmd *cobra.Command, args []string) error {
	var c domain.Config
	l := config.NewLoader(config.WithPath("./"))
	err := l.Load(&c)
	if err != nil {
		return err
	}
	return app.RunServer(&c)
}
