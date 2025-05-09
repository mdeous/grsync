package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/mdeous/grsync/api"
	"github.com/mdeous/grsync/bt"
)

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
	flag.StringVar(&cameraName, "camera", "", "Name of the Ricoh camera to connect to (required)")
	flag.StringVar(&photosDestDir, "dest", currentDir, "Destination directory for downloaded photos")
	flag.Parse()

	// Check if required parameters are provided
	if cameraName == "" {
		fmt.Println("[!] Missing required parameter: -camera")
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
	fmt.Println("[+] Requesting device information...")
	props, err := api.GetDeviceInfo()
	if err != nil {
		fmt.Printf("[!] Failed to get device information: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[-]   Model: %s\n", props.Model)
	fmt.Printf("[-]   Firmware version: %s\n", props.FirmwareVersion)
	fmt.Printf("[-]   Battery: %d%%\n", props.Battery)

	fmt.Println("[+] Downloading photos...")
	photos, err := api.GetPhotos()
	if err != nil {
		fmt.Printf("[!] Failed to list photos on camera: %v\n", err)
		os.Exit(1)
	}
	for _, dir := range photos.Dirs {
		for _, filename := range dir.Files {
			photoPath := "/" + path.Join(dir.Name, filename)
			fmt.Printf("[-]   - %s... ", photoPath)
			destPath, err := api.DownloadPhoto(photoPath, photosDestDir)
			if err != nil {
				fmt.Printf("ERROR (%s)\n", err.Error())
				continue
			}
			fmt.Println("OK")
			fmt.Printf("[-]   --> %s\n", destPath)
		}
	}
}
