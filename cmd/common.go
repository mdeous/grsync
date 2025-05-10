package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/mdeous/grsync/bt"
	"tinygo.org/x/bluetooth"
)

// establishCameraSession handles the initial Bluetooth connection and Wi-Fi enabling for the camera.
// It returns the connected Bluetooth device or an error if any step fails.
func establishCameraSession(cameraNameVal string) (camera *bluetooth.Device, err error) {
	// Enable Bluetooth adapter
	err = bt.Adapter.Enable()
	if err != nil {
		return nil, fmt.Errorf("failed to enable Bluetooth device: %w", err)
	}

	// Find Ricoh camera and connect
	fmt.Printf("[+] Scanning for Ricoh camera %s...\n", cameraNameVal)
	camera, err = bt.FindCamera(cameraNameVal)
	if err != nil {
		return nil, fmt.Errorf("failed to find camera %s: %w", cameraNameVal, err)
	}
	fmt.Printf("[-]   Connected to camera %s at address %s\n", cameraNameVal, camera.Address.String())

	// Enable camera Wi-Fi hotspot
	ssid, passphrase, err := bt.EnableWifi(camera)
	if err != nil {
		return camera, fmt.Errorf("failed to enable Wi-Fi: %w", err)
	}
	fmt.Println("[+] Wi-Fi hotspot information:")
	fmt.Printf("[-]   SSID: %s\n", ssid)
	fmt.Printf("[-]   Passphrase: %s\n", passphrase)
	return camera, nil
}

// cameraDisconnect handles the disconnection from the camera and prints relevant messages.
func cameraDisconnect(camera *bluetooth.Device) {
	if camera == nil {
		return
	}
	fmt.Println("[+] Disconnecting from camera...")
	if err := camera.Disconnect(); err != nil {
		fmt.Printf("[!] Failed to disconnect from camera: %v\n", err)
	} else {
		fmt.Println("[-]   Bluetooth connection closed")
	}
	fmt.Println("[+] All done, you can now disconnect from the camera's Wi-Fi hotspot.")
}

// waitForWifiConnection prompts the user to connect to the Wi-Fi and waits for an Enter key press.
func waitForWifiConnection() {
	fmt.Print("[>] Press Enter to continue after connecting to the Wi-Fi hotspot...")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
}
