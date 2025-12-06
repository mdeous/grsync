package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mdeous/grsync/internal/logger"
	"github.com/mdeous/grsync/pkg/ricoh/api"
	"github.com/schollz/progressbar/v3"
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
			logger.Info("Destination directory %s does not exist, creating it.", logger.Path(photosDestDir))
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

		infoSpinner := logger.StartSpinner("Fetching device information...")
		props, err := api.GetDeviceInfo()
		infoSpinner.Stop()
		if err != nil {
			logger.Fatal(err, "Failed to get device information")
		}
		logger.Success("Device information:")
		logger.SubDetail(1, "Model: %s", logger.Highlight(props.Model))
		logger.SubDetail(1, "Firmware version: %s", logger.Accent(props.FirmwareVersion))
		logger.SubDetail(1, "Serial number: %s", logger.Accent(props.SerialNumber))
		logger.SubDetail(1, "Battery level: %s%%", logger.Number(fmt.Sprintf("%d", props.Battery)))

		listSpinner := logger.StartSpinner("Fetching photo list...")
		photos, err := api.GetPhotos()
		listSpinner.Stop()
		if err != nil {
			logger.Fatal(err, "Failed to list photos on camera")
		}

		// Check if any photos were found
		if len(photos.Dirs) == 0 {
			logger.Detail("No photos found on the camera.")
			return
		}

		// Count total files matching extensions
		totalFiles := 0
		for _, dir := range photos.Dirs {
			for _, filename := range dir.Files {
				fileExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
				if extensions[fileExt] {
					totalFiles++
				}
			}
		}

		if totalFiles == 0 {
			logger.Detail("No matching photos found on the camera.")
			return
		}

		downloadCount := 0
		skipCount := 0
		failCount := 0

		logger.Info("Downloading %s photo(s)...", logger.Number(fmt.Sprintf("%d", totalFiles)))

		// Create progress bar with modern styling
		bar := progressbar.NewOptions(totalFiles,
			progressbar.OptionSetDescription("Downloading"),
			progressbar.OptionSetWidth(40),
			progressbar.OptionShowCount(),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "=",
				SaucerHead:    ">",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}),
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionSetPredictTime(true),
		)

		for _, dir := range photos.Dirs {
			for _, filename := range dir.Files {
				photoPath := path.Join(dir.Name, filename)
				// Check if the file extension is desired
				fileExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
				if !extensions[fileExt] {
					continue
				}

				_, err := api.DownloadPhoto(photoPath, photosDestDir)
				if err != nil {
					if os.IsExist(err) {
						skipCount++
					} else {
						failCount++
					}
				} else {
					downloadCount++
				}
				bar.Add(1)
			}
		}

		fmt.Println() // Add newline after progress bar

		if downloadCount > 0 {
			logger.Success("Downloaded %s photo(s) to %s", logger.Number(fmt.Sprintf("%d", downloadCount)), logger.Path(photosDestDir))
		}
		if skipCount > 0 {
			logger.Detail("Skipped %s existing photo(s)", logger.Accent(fmt.Sprintf("%d", skipCount)))
		}
		if failCount > 0 {
			logger.Warn("Failed to download %s photo(s)", logger.Number(fmt.Sprintf("%d", failCount)))
		}
		if downloadCount == 0 && skipCount == 0 && failCount == 0 {
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
