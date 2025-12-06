package cmd

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mdeous/grsync/internal/logger"
	"github.com/mdeous/grsync/pkg/ricoh/api"
	"github.com/spf13/cobra"
)

var (
	photosDestDir   string
	photoExtensions string
)

// parseExtensions takes a comma-separated list of extensions and returns a map of normalized extensions
func parseExtensions(extensionsInput string) map[string]bool {
	extensions := make(map[string]bool)

	// Handle special "all" value
	if strings.ToLower(strings.TrimSpace(extensionsInput)) == "all" {
		// Include all supported extensions
		extensions["jpg"] = true
		extensions["dng"] = true
		return extensions
	}

	for _, ext := range strings.Split(extensionsInput, ",") {
		// Normalize extension to lowercase without the dot
		ext = strings.ToLower(strings.TrimSpace(ext))
		if ext == "" {
			continue
		}
		// Remove leading dot if present
		ext = strings.TrimPrefix(ext, ".")

		// Only allow JPG and DNG extensions
		if ext == "jpg" || ext == "dng" {
			extensions[ext] = true
		}
	}
	return extensions
}

var syncCmd = &cobra.Command{
	Use:     "sync",
	Short:   "Synchronize photos from the Ricoh camera",
	Aliases: []string{"s"},
	Run: func(cmd *cobra.Command, args []string) {
		// Resolve and validate photosDestDir
		var err error
		photosDestDir, err = filepath.Abs(photosDestDir)
		if err != nil {
			logger.Fatal(err, "Failed to resolve destination directory path '%s'", photosDestDir)
		}

		// Ensure the destination directory exists, create if not.
		if _, err := os.Stat(photosDestDir); os.IsNotExist(err) {
			logger.Info("Destination directory '%s' does not exist, creating it.", photosDestDir)
			if err := os.MkdirAll(photosDestDir, 0755); err != nil {
				logger.Fatal(err, "Failed to create destination directory '%s'", photosDestDir)
			}
		} else if err != nil {
			logger.Fatal(err, "Error checking destination directory '%s'", photosDestDir)
		}

		// Parse photo extensions
		extensions := parseExtensions(photoExtensions)
		if len(extensions) == 0 {
			logger.Error(nil, "No valid photo extensions provided. Use 'jpg', 'dng', or 'all'.")
			cmd.Help()
			os.Exit(1)
		}

		// Establish camera session (BT connection and Wi-Fi)
		camera, err := establishCameraSession(cameraName)
		if camera != nil {
			defer cameraDisconnect(camera)
		}

		if err != nil {
			logger.Fatal(err, "Failed to establish camera session")
		}

		waitForWifiConnection()

		logger.Info("Fetching device information...")
		props, err := api.GetDeviceInfo()
		if err != nil {
			logger.Fatal(err, "Failed to get device information")
		}
		logger.SubDetail(1, "Model: %s", props.Model)
		logger.SubDetail(1, "Firmware version: %s", props.FirmwareVersion)
		logger.SubDetail(1, "Serial number: %s", props.SerialNumber)
		logger.SubDetail(1, "Battery level: %d%%", props.Battery)

		logger.Info("Fetching photo list...")
		photos, err := api.GetPhotos()
		if err != nil {
			logger.Fatal(err, "Failed to list photos on camera")
		}

		// Check if any photos were found
		if len(photos.Dirs) == 0 {
			logger.Detail("No photos found on the camera.")
			return
		}

		downloadCount := 0
		logger.Info("Downloading photos...")
		for _, dir := range photos.Dirs {
			for _, filename := range dir.Files {
				photoPath := path.Join(dir.Name, filename)
				// Check if the file extension is desired
				fileExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
				if !extensions[fileExt] {
					continue
				}

				logger.SubDetail(1, "Downloading %s...", photoPath)
				destPath, err := api.DownloadPhoto(photoPath, photosDestDir)
				if err != nil {
					if os.IsExist(err) {
						logger.SubDetail(2, "Skipping, file already exists: %s", destPath)
					} else {
						logger.SubWarn(2, "Failed to download %s: %v", photoPath, err)
					}
				} else {
					logger.SubDetail(2, "Saved to %s", destPath)
					downloadCount++
				}
			}
		}

		if downloadCount > 0 {
			logger.Info("Successfully downloaded %d photo(s) to %s", downloadCount, photosDestDir)
		} else {
			logger.Detail("No new photos to download.")
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().StringVarP(&photosDestDir, "dest", "d", ".", "Destination directory for photos")
	syncCmd.Flags().StringVarP(&photoExtensions, "extensions", "e", "all", "Comma-separated list of photo extensions to download (jpg, dng, all)")
	syncCmd.Flags().StringVarP(&cameraName, "camera", "c", "", "Name of the Ricoh camera to connect to")
	syncCmd.MarkFlagRequired("camera")
}
