package cmd

import (
	"github.com/mdeous/grsync/internal/logger"
	"github.com/spf13/cobra"
)

var cameraName string

var rootCmd = &cobra.Command{
	Use:   "grsync",
	Short: "Synchronize photos from a Ricoh camera.",
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is given, print help
		if len(args) == 0 {
			cmd.Help()
			return
		}
	},
}

func init() {
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal(err, "Failed to execute command")
	}
}
