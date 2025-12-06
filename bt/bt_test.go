package bt

import (
	"errors"
	"testing"
	"time"

	"tinygo.org/x/bluetooth"
)

// -- Mocks --

type MockAdapter struct {
	EnableFunc   func() error
	ScanFunc     func(callback func(adapter *bluetooth.Adapter, device bluetooth.ScanResult)) error
	StopScanFunc func() error
	ConnectFunc  func(address bluetooth.Address, params bluetooth.ConnectionParams) (Device, error)
}

func (m *MockAdapter) Enable() error {
	if m.EnableFunc != nil {
		return m.EnableFunc()
	}
	return nil
}

func (m *MockAdapter) Scan(callback func(adapter *bluetooth.Adapter, device bluetooth.ScanResult)) error {
	if m.ScanFunc != nil {
		return m.ScanFunc(callback)
	}
	return nil
}

func (m *MockAdapter) StopScan() error {
	if m.StopScanFunc != nil {
		return m.StopScanFunc()
	}
	return nil
}

func (m *MockAdapter) Connect(address bluetooth.Address, params bluetooth.ConnectionParams) (Device, error) {
	if m.ConnectFunc != nil {
		return m.ConnectFunc(address, params)
	}
	return nil, errors.New("mock connect not implemented")
}

type MockDevice struct {
	DiscoverServicesFunc func(uuids []bluetooth.UUID) ([]Service, error)
	DisconnectFunc       func() error
	AddressFunc          func() bluetooth.Address
}

func (m *MockDevice) DiscoverServices(uuids []bluetooth.UUID) ([]Service, error) {
	if m.DiscoverServicesFunc != nil {
		return m.DiscoverServicesFunc(uuids)
	}
	return nil, nil
}

func (m *MockDevice) Disconnect() error {
	if m.DisconnectFunc != nil {
		return m.DisconnectFunc()
	}
	return nil
}

func (m *MockDevice) Address() bluetooth.Address {
	if m.AddressFunc != nil {
		return m.AddressFunc()
	}
	return bluetooth.Address{}
}

type MockService struct {
	DiscoverCharacteristicsFunc func(uuids []bluetooth.UUID) ([]Characteristic, error)
}

func (m *MockService) DiscoverCharacteristics(uuids []bluetooth.UUID) ([]Characteristic, error) {
	if m.DiscoverCharacteristicsFunc != nil {
		return m.DiscoverCharacteristicsFunc(uuids)
	}
	return nil, nil
}

type MockCharacteristic struct {
	ReadFunc                 func(data []byte) (int, error)
	WriteWithoutResponseFunc func(data []byte) (int, error)
}

func (m *MockCharacteristic) Read(data []byte) (int, error) {
	if m.ReadFunc != nil {
		return m.ReadFunc(data)
	}
	return 0, nil
}

func (m *MockCharacteristic) WriteWithoutResponse(data []byte) (int, error) {
	if m.WriteWithoutResponseFunc != nil {
		return m.WriteWithoutResponseFunc(data)
	}
	return 0, nil
}

type MockAdvertisementPayload struct {
	LocalNameFunc        func() string
	HasServiceUUIDFunc   func(uuid bluetooth.UUID) bool
	ServiceUUIDsFunc     func() []bluetooth.UUID
	BytesFunc            func() []byte
	ManufacturerDataFunc func() []bluetooth.ManufacturerDataElement
	ServiceDataFunc      func() []bluetooth.ServiceDataElement
}

func (m *MockAdvertisementPayload) LocalName() string {
	if m.LocalNameFunc != nil {
		return m.LocalNameFunc()
	}
	return ""
}

func (m *MockAdvertisementPayload) HasServiceUUID(uuid bluetooth.UUID) bool {
	if m.HasServiceUUIDFunc != nil {
		return m.HasServiceUUIDFunc(uuid)
	}
	return false
}

func (m *MockAdvertisementPayload) ServiceUUIDs() []bluetooth.UUID {
	if m.ServiceUUIDsFunc != nil {
		return m.ServiceUUIDsFunc()
	}
	return nil
}

func (m *MockAdvertisementPayload) Bytes() []byte {
	if m.BytesFunc != nil {
		return m.BytesFunc()
	}
	return nil
}

func (m *MockAdvertisementPayload) ManufacturerData() []bluetooth.ManufacturerDataElement {
	if m.ManufacturerDataFunc != nil {
		return m.ManufacturerDataFunc()
	}
	return nil
}

func (m *MockAdvertisementPayload) ServiceData() []bluetooth.ServiceDataElement {
	if m.ServiceDataFunc != nil {
		return m.ServiceDataFunc()
	}
	return nil
}

// -- Tests --

func TestClient_FindCamera(t *testing.T) {
	mockAdapter := &MockAdapter{
		EnableFunc: func() error { return nil },
		ScanFunc: func(callback func(adapter *bluetooth.Adapter, device bluetooth.ScanResult)) error {
			// Simulate finding a camera
			go func() {
				// Wait a bit to simulate scan time
				time.Sleep(10 * time.Millisecond)

				payload := &MockAdvertisementPayload{
					LocalNameFunc: func() string { return "GR_Target" },
					HasServiceUUIDFunc: func(uuid bluetooth.UUID) bool {
						return uuid == wlanServiceUUID
					},
				}

				dev := bluetooth.ScanResult{
					AdvertisementPayload: payload,
					Address:              bluetooth.Address{MACAddress: bluetooth.MACAddress{MAC: [6]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}}},
					RSSI:                 -50,
				}
				callback(nil, dev)
			}()
			return nil
		},
		StopScanFunc: func() error { return nil },
		ConnectFunc: func(address bluetooth.Address, params bluetooth.ConnectionParams) (Device, error) {
			return &MockDevice{}, nil
		},
	}

	client := NewClientWithAdapter(mockAdapter)

	dev, err := client.FindCamera("GR_Target", 1*time.Second)
	if err != nil {
		t.Fatalf("FindCamera failed: %v", err)
	}
	if dev == nil {
		t.Fatal("Expected device, got nil")
	}
}

func TestClient_EnableWifi(t *testing.T) {
	// Setup mocks
	// 1. Discover generic service -> wlanService
	// 2. wlanService -> NetworkChar, SSIDChar, PassphraseChar
	// 3. NetworkChar Read -> WifiDisabled
	// 4. NetworkChar Write -> Enable
	// 5. NetworkChar Read -> WifiEnabled
	// 6. SSIDChar Read -> "test-ssid"
	// 7. PassphraseChar Read -> "test-pass"

	mockCharNetwork := &MockCharacteristic{
		ReadFunc: func(data []byte) (int, error) {
			// Simulate alternating state:
			// First read: Disabled (0)
			// Second read: Enabled (1)
			return 0, nil
		},
		WriteWithoutResponseFunc: func(data []byte) (int, error) {
			return len(data), nil
		},
	}

	readCount := 0
	mockCharNetwork.ReadFunc = func(data []byte) (int, error) {
		readCount++
		if readCount == 1 {
			data[0] = byte(WifiDisabled)
			return 1, nil
		}
		data[0] = byte(WifiEnabled)
		return 1, nil
	}

	mockCharSSID := &MockCharacteristic{
		ReadFunc: func(data []byte) (int, error) {
			copy(data, []byte("test-ssid"))
			return 9, nil
		},
	}

	mockCharPass := &MockCharacteristic{
		ReadFunc: func(data []byte) (int, error) {
			copy(data, []byte("test-pass"))
			return 9, nil
		},
	}

	mockService := &MockService{
		DiscoverCharacteristicsFunc: func(uuids []bluetooth.UUID) ([]Characteristic, error) {
			if len(uuids) != 1 {
				return nil, errors.New("expected 1 uuid")
			}
			uuid := uuids[0]
			if uuid == wlanNetworkCharUUID {
				return []Characteristic{mockCharNetwork}, nil
			}
			if uuid == wlanSSIDCharUUID {
				return []Characteristic{mockCharSSID}, nil
			}
			if uuid == wlanPassphraseCharUUID {
				return []Characteristic{mockCharPass}, nil
			}
			return nil, errors.New("unknown char")
		},
	}

	mockDevice := &MockDevice{
		DiscoverServicesFunc: func(uuids []bluetooth.UUID) ([]Service, error) {
			if len(uuids) == 1 && uuids[0] == wlanServiceUUID {
				return []Service{mockService}, nil
			}
			return nil, errors.New("service not found")
		},
	}

	client := NewClientWithAdapter(&MockAdapter{})
	ssid, pass, err := client.EnableWifi(mockDevice)
	if err != nil {
		t.Fatalf("EnableWifi failed: %v", err)
	}
	if ssid != "test-ssid" {
		t.Errorf("Expected ssid 'test-ssid', got '%s'", ssid)
	}
	if pass != "test-pass" {
		t.Errorf("Expected pass 'test-pass', got '%s'", pass)
	}
}
