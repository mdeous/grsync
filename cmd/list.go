package cmd

import (
	"path"

	"github.com/mdeous/grsync/internal/logger"
	"github.com/mdeous/grsync/pkg/ricoh/api"
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
			logger.Fatal(err, "Failed to establish camera session")
		}

		waitForWifiConnection()

		fetchSpinner := logger.StartSpinner("Fetching photo list...")
		photos, err := api.GetPhotos()
		fetchSpinner.Stop()
		if err != nil {
			logger.Fatal(err, "Failed to list photos on camera")
		}

		if len(photos.Dirs) == 0 {
			logger.Detail("No photos found on the camera.")
			return
		}

		logger.Success("Photos on camera:")
		for _, dir := range photos.Dirs {
			for _, filename := range dir.Files {
				photoPath := path.Join(dir.Name, filename)
				logger.SubDetail(1, "%s", logger.Path("/"+photoPath))
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&cameraName, "camera", "c", "", "Name of the Ricoh camera to connect to")
	listCmd.MarkFlagRequired("camera")
}
