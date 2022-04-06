package main

import (
	"github.com/odpf/salt/config"
	"github.com/spf13/cobra"

	"github.com/odpf/entropy/app"
)

func serve(cmd *cobra.Command, args []string) error {
	var c app.Config
	// TODO: load config from path using flag.
	l := config.NewLoader(config.WithPath("./"))
	err := l.Load(&c)
	if err != nil {
		return err
	}
	return app.RunServer(&c)
}
