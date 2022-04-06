package main

import (
	"github.com/odpf/salt/config"
	"github.com/spf13/cobra"

	"github.com/odpf/entropy/app"
)

func migrate(cmd *cobra.Command, args []string) error {
	var c app.Config
	l := config.NewLoader(config.WithPath("./"))
	err := l.Load(&c)
	if err != nil {
		return err
	}
	return app.RunMigrations(&c)
}
