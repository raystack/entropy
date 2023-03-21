package cli

import (
	"context"

	"github.com/goto/salt/cmdx"
	"github.com/spf13/cobra"

	"github.com/goto/entropy/cli/client"
)

var rootCmd = &cobra.Command{
	Use: "entropy",
	Long: `Entropy is a framework to safely and predictably create, change, 
and improve modern cloud applications and infrastructure using 
familiar languages, tools, and engineering practices.`,
}

func Execute(ctx context.Context) {
	rootCmd.PersistentFlags().StringP(configFlag, "c", "", "Override config file")
	rootCmd.AddCommand(
		cmdServe(),
		cmdMigrate(),
		cmdVersion(),
		cmdShowConfigs(),
		client.Command(),
	)

	cmdx.SetHelp(rootCmd)
	_ = rootCmd.ExecuteContext(ctx)
}
