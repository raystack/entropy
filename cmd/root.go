package cmd

import (
	"fmt"
	"os"

	"github.com/odpf/salt/cmdx"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "entropy",
}

// Execute runs the command line interface
func Execute() {
	cmdx.SetHelp(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
