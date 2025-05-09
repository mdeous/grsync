package bt

import (
	"fmt"
	"time"

	"tinygo.org/x/bluetooth"
)

const scanMaxTime = 30 * time.Second
const wifiStartupTime = 2 * time.Second

var Adapter = bluetooth.DefaultAdapter
var CameraAddress bluetooth.Address
var WifiStatus = []string{
	"disabled",
	"enabled",
}

var (
	wlanServiceUUID, _        = bluetooth.ParseUUID("F37F568F-9071-445D-A938-5441F2E82399")
	wlanNetworkCharUUID, _    = bluetooth.ParseUUID("9111CDD0-9F01-45C4-A2D4-E09E8FB0424D")
	wlanSSIDCharUUID, _       = bluetooth.ParseUUID("90638E5A-E77D-409D-B550-78F7E1CA5AB4")
	wlanPassphraseCharUUID, _ = bluetooth.ParseUUID("0F38279C-FE9E-461B-8596-81287E8C9A81")
)

var cameraFound = false
var scanDone = make(chan struct{})
var rxBuffSize = 64

func stopScan(cameraFound bool) {
	close(scanDone)
	if !cameraFound {
		fmt.Printf("[!] Bluetooth scan timed out after %s, stopping scan\n", scanMaxTime)
	}
	if err := Adapter.StopScan(); err != nil {
		fmt.Printf("[!] Failed to stop Bluetooth scan: %v\n", err)
	}
	fmt.Println("[-]   Stopped scanning")
}

func scanCallback(deviceName string) func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
	return func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		// We already found the camera, do nothing with new advertisements
		if cameraFound {
			return
		}
		// Check if the device is our camera
		devAddr := device.Address.String()
		if device.AdvertisementPayload.HasServiceUUID(wlanServiceUUID) && device.LocalName() == deviceName {
			cameraFound = true
			fmt.Printf("[-] Found Ricoh camera %s with address %s (RSSI: %d)\n", deviceName, devAddr, device.RSSI)
			CameraAddress = device.Address
			// Camera found, no need to continue scanning
			stopScan(true)
		}
	}
}

func getService(device *bluetooth.Device, svcUUID bluetooth.UUID) (*bluetooth.DeviceService, error) {
	// Discover the requested service
	services, err := device.DiscoverServices([]bluetooth.UUID{svcUUID})
	if err != nil || len(services) != 1 {
		return nil, fmt.Errorf("failed to discover %s service: %v", svcUUID.String(), err)
	}
	svc := services[0]
	return &svc, nil
}

func readCharacteristic(svc *bluetooth.DeviceService, charUUID bluetooth.UUID) (*bluetooth.DeviceCharacteristic, []byte, error) {
	// Discover the requested characteristic
	characteristic, err := svc.DiscoverCharacteristics([]bluetooth.UUID{charUUID})
	if err != nil || len(characteristic) != 1 {
		return nil, nil, fmt.Errorf("failed to discover %s characteristic: %v", charUUID.String(), err)
	}
	chr := characteristic[0]

	// Read the characteristic value
	buffer := make([]byte, rxBuffSize)
	dataLen, err := chr.Read(buffer)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read characteristic %s data: %v", charUUID.String(), err)
	}

	// Return the relevant portion of the buffer
	return &chr, buffer[:dataLen], nil
}

func getWifiStatus(svc *bluetooth.DeviceService) (*bluetooth.DeviceCharacteristic, int, error) {
	chr, wlanStatus_bytes, err := readCharacteristic(svc, wlanNetworkCharUUID)
	if err != nil {
		return nil, 0, err
	}
	wlanStatusLen := len(wlanStatus_bytes)
	if wlanStatusLen != 1 {
		return nil, 0, fmt.Errorf("unexpected Wi-Fi status length: %d", wlanStatusLen)
	}
	wlanStatusValue := int(wlanStatus_bytes[0])
	fmt.Printf("[+] Wi-Fi hotspot status: %s\n", WifiStatus[wlanStatusValue])
	return chr, wlanStatusValue, nil
}

func FindCamera(name string) (*bluetooth.Device, error) {
	// Scan timeout handler
	go func() {
		for {
			select {
			case <-time.After(scanMaxTime):
				stopScan(false)
				return
			case <-scanDone:
				return
			}
		}
	}()
	// Start scanning for devices
	cb := scanCallback(name)
	if err := Adapter.Scan(cb); err != nil {
		return nil, fmt.Errorf("failed to start Bluetooth scan: %v", err)
	}
	// Wait for scan to finish
	<-scanDone
	if !cameraFound {
		return nil, fmt.Errorf("%s not found", name)
	}
	// Connect to the camera
	fmt.Println("[+] Connecting to camera...")
	cameraDevice, err := Adapter.Connect(CameraAddress, bluetooth.ConnectionParams{})
	if err != nil {
		// NOTE: for some reason the 1st connection sometimes fails and requires a retry
		cameraDevice, err = Adapter.Connect(CameraAddress, bluetooth.ConnectionParams{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to camera %s: %v", name, err)
		}
	}
	return &cameraDevice, nil
}

func EnableWifi(device *bluetooth.Device) (ssid string, passphrase string, err error) {
	// Discover the WLAN service
	svc, err := getService(device, wlanServiceUUID)
	if err != nil {
		return "", "", err
	}

	// Get Wi-Fi hotspot status
	chr, wlanStatusValue, err := getWifiStatus(svc)
	if err != nil {
		return "", "", err
	}
	if wlanStatusValue == 0 {
		fmt.Println("[+] Enabling Wi-Fi hotspot...")
		// Enable Wi-Fi hotspot
		wlanStatus_bytes := []byte{1}
		chr.WriteWithoutResponse(wlanStatus_bytes)
		time.Sleep(wifiStartupTime)
		// Check if Wi-Fi was successfully enabled
		_, _, err = getWifiStatus(svc)
		if err != nil {
			return "", "", err
		}
	}

	// Get Wi-Fi SSID
	_, ssid_b, err := readCharacteristic(svc, wlanSSIDCharUUID)
	if err != nil {
		return "", "", err
	}
	ssid = string(ssid_b)

	// Get Wi-Fi passphrase
	_, passphrase_b, err := readCharacteristic(svc, wlanPassphraseCharUUID)
	if err != nil {
		return "", "", err
	}
	passphrase = string(passphrase_b)

	return
}
