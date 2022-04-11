package main

import (
	"github.com/odpf/salt/cmdx"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "entropy",
	Long: `Entropy is a framework to safely and predictably create, change, 
and improve modern cloud applications and infrastructure using 
familiar languages, tools, and engineering practices.`,
}

func main() {
	rootCmd.PersistentFlags().StringP(configFlag, "c", "", "Override config file")
	rootCmd.AddCommand(
		cmdServe(),
		cmdMigrate(),
		cmdVersion(),
	)

	cmdx.SetHelp(rootCmd)
	_ = rootCmd.Execute()
}
