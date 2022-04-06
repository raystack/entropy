package main

import (
	"github.com/odpf/salt/cmdx"
	"github.com/spf13/cobra"

	v "github.com/odpf/entropy/pkg/version"
)

var rootCmd = &cobra.Command{
	Use: "entropy",
}

func main() {
	cmdx.SetHelp(rootCmd)

	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "migrate",
			Short: "Run DB migrations",
			RunE:  migrate,
		},
		&cobra.Command{
			Use:   "serve",
			Short: "Run server",
			RunE:  serve,
		},
		&cobra.Command{
			Use:   "version",
			Short: "Show version",
			RunE:  version,
		},
	)

	_ = rootCmd.Execute()
}

func version(cmd *cobra.Command, args []string) error {
	return v.Print()
}
