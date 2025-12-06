package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/mdeous/grsync/internal/logger"
	"github.com/mdeous/grsync/pkg/ricoh/bt"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Scan for available Ricoh GR cameras",
	Run: func(cmd *cobra.Command, args []string) {
		client := bt.NewClient()

		if err := client.Enable(); err != nil {
			logger.Fatal(err, "Failed to enable Bluetooth adapter")
		}

		scanSpinner := logger.StartSpinner("Scanning for available Ricoh cameras...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		results, err := client.Scan(ctx, "GR_")
		scanSpinner.Stop()
		if err != nil {
			logger.Fatal(err, "Scan failed")
		}

		if len(results) == 0 {
			logger.Detail("No cameras found.")
			return
		}

		logger.Success("Found %s camera(s):", logger.Number(fmt.Sprintf("%d", len(results))))
		for _, res := range results {
			logger.SubDetail(1, "%s (RSSI: %s)", logger.Highlight(res.Name), logger.Accent(fmt.Sprintf("%d", res.RSSI)))
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
