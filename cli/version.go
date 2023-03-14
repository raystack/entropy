package cli

import (
	"github.com/spf13/cobra"

	v "github.com/goto/entropy/pkg/version"
)

func cmdVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.Print()
		},
	}
}
