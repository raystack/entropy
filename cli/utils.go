package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type RunEFunc func(cmd *cobra.Command, args []string) error

func fatalExitf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
	os.Exit(1)
}

func handleErr(fn RunEFunc) RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		if err := fn(cmd, args); err != nil {
			fatalExitf(err.Error())
		}
		return nil
	}
}
