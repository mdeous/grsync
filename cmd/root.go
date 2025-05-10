package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "grsync",
	Short: "grsync is a tool to synchronize files from a Ricoh camera.",
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is given, print help
		if len(args) == 0 {
			cmd.Help()
			return
		}
	},
}

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
