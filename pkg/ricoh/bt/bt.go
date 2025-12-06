package bt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mdeous/grsync/internal/logger"
	"tinygo.org/x/bluetooth"
)

const (
	defaultScanTimeout = 30 * time.Second
	wifiStartupTime    = 2 * time.Second
	rxBuffSize         = 64
)

type WifiState int

const (
	WifiDisabled WifiState = iota
	WifiEnabled
	WifiUnknown
)

func (s WifiState) String() string {
	switch s {
	case WifiDisabled:
		return "disabled"
	case WifiEnabled:
		return "enabled"
	default:
		return "unknown"
	}
}

// Bluetooth UUIDs
var (
	wlanServiceUUID, _        = bluetooth.ParseUUID("F37F568F-9071-445D-A938-5441F2E82399")
	wlanNetworkCharUUID, _    = bluetooth.ParseUUID("9111CDD0-9F01-45C4-A2D4-E09E8FB0424D")
	wlanSSIDCharUUID, _       = bluetooth.ParseUUID("90638E5A-E77D-409D-B550-78F7E1CA5AB4")
	wlanPassphraseCharUUID, _ = bluetooth.ParseUUID("0F38279C-FE9E-461B-8596-81287E8C9A81")
)

type Client struct {
	adapter Adapter
}

func NewClient() *Client {
	return &Client{
		adapter: &RealAdapter{adapter: bluetooth.DefaultAdapter},
	}
}

// For testing purposes
func NewClientWithAdapter(a Adapter) *Client {
	return &Client{
		adapter: a,
	}
}

func (c *Client) Enable() error {
	return c.adapter.Enable()
}

type ScanResult struct {
	Name    string
	Address bluetooth.Address
	RSSI    int16
}

func (c *Client) Scan(ctx context.Context, prefix string) ([]ScanResult, error) {
	var results []ScanResult
	seen := make(map[string]bool)

	// The callback is called from the adapter scan loop.
	// For RealAdapter, it calls tinygo Scan.

	// Start a goroutine to stop the scan when the context is done
	go func() {
		<-ctx.Done()
		c.adapter.StopScan()
	}()

	err := c.adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		name := device.LocalName()
		addr := device.Address.String()

		if seen[addr] {
			return
		}

		if prefix != "" && !strings.HasPrefix(name, prefix) {
			return
		}

		results = append(results, ScanResult{
			Name:    name,
			Address: device.Address,
			RSSI:    device.RSSI,
		})
		seen[addr] = true
	})

	if err != nil {
		return nil, fmt.Errorf("failed to start scan: %w", err)
	}

	return results, nil
}

func (c *Client) FindCamera(name string, timeout time.Duration) (Device, error) {
	if timeout == 0 {
		timeout = defaultScanTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cameraAddress bluetooth.Address
	found := make(chan struct{})

	logger.Info("Scanning for camera: %s...", name)

	// Start a goroutine to stop the scan when the context is done or camera is found
	go func() {
		select {
		case <-found:
			// Camera found, scan already stopped
		case <-ctx.Done():
			c.adapter.StopScan()
		}
	}()

	// Capture the closure variables properly
	err := c.adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		if device.LocalName() == name {
			if device.AdvertisementPayload.HasServiceUUID(wlanServiceUUID) {
				cameraAddress = device.Address
				c.adapter.StopScan() // Use outer adapter reference
				close(found)
			}
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start scan: %w", err)
	}

	select {
	case <-found:
	case <-ctx.Done():
		return nil, fmt.Errorf("camera %s not found within timeout", name)
	}

	logger.Info("Found camera %s, connecting...", name)

	device, err := c.adapter.Connect(cameraAddress, bluetooth.ConnectionParams{})
	if err != nil {
		// Retry
		device, err = c.adapter.Connect(cameraAddress, bluetooth.ConnectionParams{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to camera %s: %w", name, err)
		}
	}

	return device, nil
}

func getService(device Device, svcUUID bluetooth.UUID) (Service, error) {
	services, err := device.DiscoverServices([]bluetooth.UUID{svcUUID})
	if err != nil {
		return nil, fmt.Errorf("failed to discover %s service: %w", svcUUID.String(), err)
	}
	if len(services) != 1 {
		return nil, fmt.Errorf("unexpected number of services found: %d", len(services))
	}
	return services[0], nil
}

func readCharacteristic(svc Service, charUUID bluetooth.UUID) (Characteristic, []byte, error) {
	characteristics, err := svc.DiscoverCharacteristics([]bluetooth.UUID{charUUID})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to discover %s characteristic: %w", charUUID.String(), err)
	}
	if len(characteristics) != 1 {
		return nil, nil, fmt.Errorf("unexpected number of characteristics found: %d", len(characteristics))
	}
	chr := characteristics[0]

	buffer := make([]byte, rxBuffSize)
	dataLen, err := chr.Read(buffer)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read characteristic %s data: %w", charUUID.String(), err)
	}

	return chr, buffer[:dataLen], nil
}

func getWifiStatus(svc Service) (Characteristic, WifiState, error) {
	chr, statusBytes, err := readCharacteristic(svc, wlanNetworkCharUUID)
	if err != nil {
		return nil, WifiUnknown, err
	}
	if len(statusBytes) != 1 {
		return nil, WifiUnknown, fmt.Errorf("unexpected Wi-Fi status length: %d", len(statusBytes))
	}
	status := WifiState(statusBytes[0])
	logger.Info("Wi-Fi hotspot status: %s", status)
	return chr, status, nil
}

func (c *Client) EnableWifi(device Device) (string, string, error) {
	svc, err := getService(device, wlanServiceUUID)
	if err != nil {
		return "", "", err
	}

	chr, status, err := getWifiStatus(svc)
	if err != nil {
		return "", "", err
	}

	if status == WifiDisabled {
		logger.Info("Enabling Wi-Fi hotspot...")
		if _, err := chr.WriteWithoutResponse([]byte{byte(WifiEnabled)}); err != nil {
			return "", "", fmt.Errorf("failed to write Wi-Fi status: %w", err)
		}

		time.Sleep(wifiStartupTime)

		_, newStatus, err := getWifiStatus(svc)
		if err != nil {
			return "", "", err
		}
		if newStatus != WifiEnabled {
			return "", "", fmt.Errorf("failed to enable Wi-Fi hotspot, current status: %s", newStatus)
		}
	}

	_, ssidBytes, err := readCharacteristic(svc, wlanSSIDCharUUID)
	if err != nil {
		return "", "", err
	}

	_, passBytes, err := readCharacteristic(svc, wlanPassphraseCharUUID)
	if err != nil {
		return "", "", err
	}

	return string(ssidBytes), string(passBytes), nil
}
