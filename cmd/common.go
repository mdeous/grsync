package cmd

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/mdeous/grsync/bt"
	"github.com/mdeous/grsync/internal/logger"
)

// establishCameraSession handles the initial Bluetooth connection and Wi-Fi enabling for the camera.
// It returns the connected Bluetooth device or an error if any step fails.
func establishCameraSession(cameraNameVal string) (camera bt.Device, err error) {
	btClient := bt.NewClient()

	// Enable Bluetooth adapter
	err = btClient.Enable()
	if err != nil {
		return nil, fmt.Errorf("failed to enable Bluetooth device: %w", err)
	}

	// Find Ricoh camera and connect
	logger.Info("Scanning for Ricoh camera %s...", cameraNameVal)
	camera, err = btClient.FindCamera(cameraNameVal, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to find camera %s: %w", cameraNameVal, err)
	}
	logger.SubDetail(1, "Connected to camera %s at address %s", cameraNameVal, camera.Address().String())

	// Enable camera Wi-Fi hotspot
	ssid, passphrase, err := btClient.EnableWifi(camera)
	if err != nil {
		return camera, fmt.Errorf("failed to enable Wi-Fi: %w", err)
	}
	logger.Info("Wi-Fi hotspot information:")
	logger.SubDetail(1, "SSID: %s", ssid)
	logger.SubDetail(1, "Passphrase: %s", passphrase)
	return camera, nil
}

// cameraDisconnect handles the disconnection from the camera and prints relevant messages.
func cameraDisconnect(camera bt.Device) {
	if camera == nil {
		return
	}
	logger.Info("Disconnecting from camera...")
	if err := camera.Disconnect(); err != nil {
		logger.Warn("Failed to disconnect from camera: %v", err)
	} else {
		logger.SubDetail(1, "Bluetooth connection closed")
	}
	logger.Info("All done, you can now disconnect from the camera's Wi-Fi hotspot.")
}

// waitForWifiConnection prompts the user to connect to the Wi-Fi and waits for an Enter key press.
func waitForWifiConnection() {
	fmt.Printf("%s[>] Press Enter to continue after connecting to the Wi-Fi hotspot...%s", logger.ColorBoldCyan, logger.ColorReset)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
}
