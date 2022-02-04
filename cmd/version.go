package cmd

import (
	v "github.com/odpf/entropy/pkg/version"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version",
		RunE:  version,
	})
}

func version(cmd *cobra.Command, args []string) error {
	return v.Print()
}
