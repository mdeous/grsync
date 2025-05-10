package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mdeous/grsync/api"
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
			fmt.Printf("[!] Failed to resolve destination directory path '%s': %v\n", photosDestDir, err)
			os.Exit(1)
		}

		// Ensure the destination directory exists, create if not.
		if _, err := os.Stat(photosDestDir); os.IsNotExist(err) {
			fmt.Printf("[+] Destination directory '%s' does not exist, creating it.\n", photosDestDir)
			if err := os.MkdirAll(photosDestDir, 0755); err != nil {
				fmt.Printf("[!] Failed to create destination directory '%s': %v\n", photosDestDir, err)
				os.Exit(1)
			}
		} else if err != nil {
			fmt.Printf("[!] Error checking destination directory '%s': %v\n", photosDestDir, err)
			os.Exit(1)
		}

		// Parse photo extensions
		extensions := parseExtensions(photoExtensions)
		if len(extensions) == 0 {
			fmt.Println("[!] No valid photo extensions provided. Use 'jpg', 'dng', or 'all'.")
			cmd.Help()
			os.Exit(1)
		}

		// Establish camera session (BT connection and Wi-Fi)
		camera, err := establishCameraSession(cameraName)
		if camera != nil {
			defer cameraDisconnect(camera)
		}

		if err != nil {
			fmt.Printf("[!] Failed to establish camera session: %v\n", err)
			os.Exit(1)
		}

		waitForWifiConnection()

		fmt.Println("[+] Fetching device information...")
		props, err := api.GetDeviceInfo()
		if err != nil {
			fmt.Printf("[!] Failed to get device information: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[-]   Model: %s\n", props.Model)
		fmt.Printf("[-]   Firmware version: %s\n", props.FirmwareVersion)
		fmt.Printf("[-]   Serial number: %s\n", props.SerialNumber)
		fmt.Printf("[-]   Battery level: %d%%\n", props.Battery)

		fmt.Println("[+] Fetching photo list...")
		photos, err := api.GetPhotos()
		if err != nil {
			fmt.Printf("[!] Failed to list photos on camera: %v\n", err)
			os.Exit(1)
		}

		// Check if any photos were found
		if len(photos.Dirs) == 0 {
			fmt.Println("[-] No photos found on the camera.")
			return
		}

		downloadCount := 0
		fmt.Println("[+] Downloading photos...")
		for _, dir := range photos.Dirs {
			for _, filename := range dir.Files {
				photoPath := path.Join(dir.Name, filename)
				// Check if the file extension is desired
				fileExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
				if !extensions[fileExt] {
					continue
				}

				fmt.Printf("[-]   Downloading %s...\n", photoPath)
				destPath, err := api.DownloadPhoto(photoPath, photosDestDir)
				if err != nil {
					if os.IsExist(err) {
						fmt.Printf("[-]     Skipping, file already exists: %s\n", destPath)
					} else {
						fmt.Printf("[!]     Failed to download %s: %v\n", photoPath, err)
					}
				} else {
					fmt.Printf("[-]     Saved to %s\n", destPath)
					downloadCount++
				}
			}
		}

		if downloadCount > 0 {
			fmt.Printf("[+] Successfully downloaded %d photo(s) to %s\n", downloadCount, photosDestDir)
		} else {
			fmt.Println("[-] No new photos to download.")
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().StringVarP(&photosDestDir, "dest", "d", ".", "Destination directory for photos")
	syncCmd.Flags().StringVarP(&photoExtensions, "extensions", "e", "all", "Comma-separated list of photo extensions to download (jpg, dng, all)")
}
