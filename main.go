package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mdeous/grsync/api"
	"github.com/mdeous/grsync/bt"
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

func main() {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("[!] Failed to get current directory: %v\n", err)
		os.Exit(1)
	}

	// Parse command line arguments
	var cameraName string
	var photosDestDir string
	var photoExtensions string
	flag.StringVar(&cameraName, "camera", "", "Name of the Ricoh camera to connect to (required)")
	flag.StringVar(&photosDestDir, "dest", currentDir, "Destination directory for downloaded photos")
	flag.StringVar(&photoExtensions, "ext", "all", "Photo extensions to download (e.g., jpg,dng) or 'all' for all types")
	flag.Parse()

	// Check if required parameters are provided
	if cameraName == "" {
		fmt.Println("[!] Missing required parameter: -camera")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Parse photo extensions
	extensions := parseExtensions(photoExtensions)
	if len(extensions) == 0 {
		fmt.Println("[!] No valid photo extensions provided")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Enable Bluetooth adapter
	err = bt.Adapter.Enable()
	if err != nil {
		fmt.Printf("[!] Failed to enable Bluetooth device: %v\n", err)
		os.Exit(1)
	}

	// Find Ricoh camera and connect
	fmt.Printf("[+] Scanning for Ricoh camera %s...\n", cameraName)
	camera, err := bt.FindCamera(cameraName)
	if err != nil {
		fmt.Printf("[!] Failed to find camera: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[-]   Connected to camera %s at address %s\n", cameraName, camera.Address.String())

	// Defer disconnection to ensure it happens when the function exits
	defer func() {
		fmt.Println("[+] Disconnecting from camera...")
		if err := camera.Disconnect(); err != nil {
			fmt.Printf("[!] Failed to disconnect from camera: %v\n", err)
		} else {
			fmt.Println("[-]   Bluetooth connection closed")
		}
		fmt.Println("[+] All done, you can now disconnect from the camera's Wi-Fi hotspot.")
	}()

	// Enable camera Wi-Fi hotspot
	ssid, passphrase, err := bt.EnableWifi(camera)
	if err != nil {
		fmt.Printf("[!] Failed to enable Wi-Fi: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("[+] Wi-Fi hotspot information:")
	fmt.Printf("[-]   SSID: %s\n", ssid)
	fmt.Printf("[-]   Passphrase: %s\n", passphrase)

	// Wait for user to connect to Wi-Fi hotspot
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("[>] Press Enter to continue after connecting to the Wi-Fi hotspot...")
	scanner.Scan()

	// Get device information
	fmt.Println("[+] Fetching device information...")
	props, err := api.GetDeviceInfo()
	if err != nil {
		fmt.Printf("[!] Failed to get device information: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[-]   Model: %s\n", props.Model)
	fmt.Printf("[-]   Firmware version: %s\n", props.FirmwareVersion)
	fmt.Printf("[-]   Serial number: %s\n", props.SerialNumber)
	fmt.Printf("[-]   Battery: %d%%\n", props.Battery)

	fmt.Println("[+] Downloading photos...")
	photos, err := api.GetPhotos()
	if err != nil {
		fmt.Printf("[!] Failed to list photos on camera: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[-] Photo formats to download: %s\n", photoExtensions)
	downloadCount := 0
	skippedCount := 0
	errorCount := 0

	for _, dir := range photos.Dirs {
		for _, filename := range dir.Files {
			// Get file extension and normalize it
			ext := strings.ToLower(filepath.Ext(filename))
			ext = strings.TrimPrefix(ext, ".")

			// Skip files with extensions not in our filter
			if !extensions[ext] {
				skippedCount++
				continue
			}

			photoPath := "/" + path.Join(dir.Name, filename)
			fmt.Printf("[-]   - %s... ", photoPath)
			destPath, err := api.DownloadPhoto(photoPath, photosDestDir)
			if err != nil {
				fmt.Printf("ERROR (%s)\n", err.Error())
				errorCount++
				continue
			}
			fmt.Println("OK")
			fmt.Printf("[-]   --> %s\n", destPath)
			downloadCount++
		}
	}

	fmt.Println("[+] Download complete")
	fmt.Printf("[-]   Downloaded: %d\n", downloadCount)
	fmt.Printf("[-]   Skipped: %d\n", skippedCount)
	fmt.Printf("[-]   Errors: %d\n", errorCount)
}
