package bt

import (
	"tinygo.org/x/bluetooth"
)

type Adapter interface {
	Enable() error
	Scan(callback func(adapter *bluetooth.Adapter, device bluetooth.ScanResult)) error
	StopScan() error
	Connect(address bluetooth.Address, params bluetooth.ConnectionParams) (Device, error)
}

// Wrapper to make bluetooth.Adapter satisfy our Adapter interface.
// tinygo's bluetooth.Adapter is a struct, so we use a thin wrapper to satisfy the interface.

type RealAdapter struct {
	adapter *bluetooth.Adapter
}

func (a *RealAdapter) Enable() error {
	return a.adapter.Enable()
}

func (a *RealAdapter) Scan(callback func(adapter *bluetooth.Adapter, device bluetooth.ScanResult)) error {
	return a.adapter.Scan(callback)
}

func (a *RealAdapter) StopScan() error {
	return a.adapter.StopScan()
}

func (a *RealAdapter) Connect(address bluetooth.Address, params bluetooth.ConnectionParams) (Device, error) {
	d, err := a.adapter.Connect(address, params)
	if err != nil {
		return nil, err
	}
	return &RealDevice{device: &d}, nil
}

type Device interface {
	DiscoverServices(uuids []bluetooth.UUID) ([]Service, error)
	Disconnect() error
	Address() bluetooth.Address
}

type RealDevice struct {
	device *bluetooth.Device
}

func (d *RealDevice) DiscoverServices(uuids []bluetooth.UUID) ([]Service, error) {
	services, err := d.device.DiscoverServices(uuids)
	if err != nil {
		return nil, err
	}
	// Convert []bluetooth.DeviceService to []Service
	wrapped := make([]Service, len(services))
	for i, s := range services {
		wrapped[i] = &RealService{service: &s}
	}
	return wrapped, nil
}

func (d *RealDevice) Disconnect() error {
	return d.device.Disconnect()
}

func (d *RealDevice) Address() bluetooth.Address {
	return d.device.Address
}

type Service interface {
	DiscoverCharacteristics(uuids []bluetooth.UUID) ([]Characteristic, error)
}

type RealService struct {
	service *bluetooth.DeviceService
}

func (s *RealService) DiscoverCharacteristics(uuids []bluetooth.UUID) ([]Characteristic, error) {
	chars, err := s.service.DiscoverCharacteristics(uuids)
	if err != nil {
		return nil, err
	}
	wrapped := make([]Characteristic, len(chars))
	for i, c := range chars {
		wrapped[i] = &RealCharacteristic{characteristic: &c}
	}
	return wrapped, nil
}

type Characteristic interface {
	Read(data []byte) (int, error)
	WriteWithoutResponse(data []byte) (int, error)
}

type RealCharacteristic struct {
	characteristic *bluetooth.DeviceCharacteristic
}

func (c *RealCharacteristic) Read(data []byte) (int, error) {
	return c.characteristic.Read(data)
}

func (c *RealCharacteristic) WriteWithoutResponse(data []byte) (int, error) {
	return c.characteristic.WriteWithoutResponse(data)
}
