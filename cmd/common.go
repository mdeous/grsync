package cmd

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/mdeous/grsync/internal/logger"
	"github.com/mdeous/grsync/pkg/ricoh/bt"
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
	scanSpinner := logger.StartSpinner(fmt.Sprintf("Scanning for Ricoh camera %s...", logger.Highlight(cameraNameVal)))
	camera, err = btClient.FindCamera(cameraNameVal, 10*time.Second)
	scanSpinner.Stop()
	if err != nil {
		return nil, fmt.Errorf("failed to find camera %s: %w", cameraNameVal, err)
	}
	logger.Success("Connected to camera %s at address %s", logger.Highlight(cameraNameVal), logger.Accent(camera.Address().String()))

	// Enable camera Wi-Fi hotspot
	ssid, passphrase, err := btClient.EnableWifi(camera)
	if err != nil {
		return camera, fmt.Errorf("failed to enable Wi-Fi: %w", err)
	}
	logger.Info("Wi-Fi hotspot information:")
	logger.SubDetail(1, "SSID: %s", logger.Highlight(ssid))
	logger.SubDetail(1, "Passphrase: %s", logger.Accent(passphrase))
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
		logger.Detail("Bluetooth connection closed")
	}
	logger.Success("All done, you can now disconnect from the camera's Wi-Fi hotspot.")
}

// waitForWifiConnection prompts the user to connect to the Wi-Fi and waits for an Enter key press.
func waitForWifiConnection() {
	promptStyle := "\033[1;36m" // Cyan bold
	resetStyle := "\033[0m"
	fmt.Printf("\n%s⏎ Press Enter to continue after connecting to the Wi-Fi hotspot...%s ", promptStyle, resetStyle)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	fmt.Println()
}
