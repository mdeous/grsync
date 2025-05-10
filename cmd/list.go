package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/mdeous/grsync/api"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List photos on the Ricoh camera",
	Aliases: []string{"l"},
	Run: func(cmd *cobra.Command, args []string) {
		camera, err := establishCameraSession(cameraName)
		if camera != nil {
			defer cameraDisconnect(camera)
		}

		if err != nil {
			fmt.Printf("[!] Failed to establish camera session: %v\n", err)
			os.Exit(1)
		}

		waitForWifiConnection()

		fmt.Println("[+] Fetching photo list...")
		photos, err := api.GetPhotos()
		if err != nil {
			fmt.Printf("[!] Failed to list photos on camera: %v\n", err)
			os.Exit(1)
		}

		if len(photos.Dirs) == 0 {
			fmt.Println("[-] No photos found on the camera.")
			return
		}

		fmt.Println("[-] Photos on camera:")
		for _, dir := range photos.Dirs {
			for _, filename := range dir.Files {
				photoPath := path.Join(dir.Name, filename)
				fmt.Printf("  - /%s\n", photoPath)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.MarkFlagRequired("camera")
}
