package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cameraName string

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

func init() {
	rootCmd.PersistentFlags().StringVarP(&cameraName, "camera", "c", "", "Name of the Ricoh camera to connect to")
	rootCmd.MarkPersistentFlagRequired("camera")
}

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
