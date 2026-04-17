// Package device provides core device types, interfaces, and factory functions.
package device

import (
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// Types
// ============================================================================

// DeviceOption is a functional option for device configuration.
type DeviceOption func(*DeviceConfig)

// DeviceConfig holds device configuration options.
// This is used internally by the factory functions.
type DeviceConfig struct {
	DeviceID        *string
	DeviceIP        *string
	DevicePort      *int
	SN              *string
	Name            *string
	Version         *int
	PresetToken     Token
	PresetKey       Key
	LocalKey        LocalKey
	LocalKeyExpired *time.Time
}

func (d *DeviceConfig) GetLocalKeyExpired() time.Time {
	if d.LocalKeyExpired == nil {
		return time.Time{}
	}
	return *d.LocalKeyExpired
}

// DeviceFactory is a function type for creating device instances.
type DeviceFactory func(opts ...DeviceOption) Device

// ============================================================================
// Variables
// ============================================================================

var (
	// deviceFactories stores registered device type factories
	deviceFactories = make(map[DeviceType]DeviceFactory)
	factoriesMu     sync.RWMutex

	// FallbackNewDevice is called when no factory is registered for a device type.
	// This allows the msmart package to provide a fallback implementation.
	FallbackNewDevice = func(deviceType DeviceType, opts ...DeviceOption) Device {
		return nil
	}
)

// ============================================================================
// Functional Options
// ============================================================================

// WithDeviceAddr sets the device IP address and port.
func WithDeviceAddr(ip string, port int) DeviceOption {
	return func(cfg *DeviceConfig) {
		cfg.DeviceIP = &ip
		cfg.DevicePort = &port
	}
}

// WithDeviceID sets the device ID.
func WithDeviceID(deviceID string) DeviceOption {
	return func(cfg *DeviceConfig) {
		cfg.DeviceID = &deviceID
	}
}

// WithSN sets the serial number.
func WithSN(sn string) DeviceOption {
	return func(cfg *DeviceConfig) {
		cfg.SN = &sn
	}
}

// WithName sets the device name.
func WithName(name string) DeviceOption {
	return func(cfg *DeviceConfig) {
		cfg.Name = &name
	}
}

// WithVersion sets the device version.
func WithVersion(version int) DeviceOption {
	return func(cfg *DeviceConfig) {
		cfg.Version = &version
	}
}

// WithTokenKey sets the pre-set token and key for V3 devices.
func WithTokenKey(token Token, key Key) DeviceOption {
	return func(cfg *DeviceConfig) {
		cfg.PresetToken = token
		cfg.PresetKey = key
	}
}

func WithLocalKey(key LocalKey, expired time.Time) DeviceOption {
	return func(cfg *DeviceConfig) {
		cfg.LocalKey = key
		cfg.LocalKeyExpired = &expired
	}
}

// ============================================================================
// Factory Registry
// ============================================================================

// RegisterDeviceType registers a device type factory.
// This should be called in init() functions of device implementation packages.
func RegisterDeviceType(deviceType DeviceType, factory DeviceFactory) {
	factoriesMu.Lock()
	defer factoriesMu.Unlock()
	deviceFactories[deviceType] = factory
}

// GetDeviceFactory returns the factory for a device type.
// Returns an error if no factory is registered for the given device type.
func GetDeviceFactory(deviceType DeviceType) (DeviceFactory, error) {
	factoriesMu.RLock()
	defer factoriesMu.RUnlock()

	factory, ok := deviceFactories[deviceType]
	if !ok {
		return nil, fmt.Errorf("unsupported device type: %v (no factory registered)", deviceType)
	}
	return factory, nil
}

// IsDeviceTypeRegistered checks if a device type has a registered factory.
func IsDeviceTypeRegistered(deviceType DeviceType) bool {
	factoriesMu.RLock()
	defer factoriesMu.RUnlock()
	_, ok := deviceFactories[deviceType]
	return ok
}

// NewDeviceFromType creates a device instance based on the provided device type.
// If a factory is registered for the device type, it uses that factory.
// Otherwise, it falls back to FallbackNewDevice.
func NewDeviceFromType(deviceType DeviceType, opts ...DeviceOption) Device {
	factory, err := GetDeviceFactory(deviceType)
	if err == nil {
		return factory(opts...)
	}

	// Fallback to creating a generic device if no factory registered
	return FallbackNewDevice(deviceType, opts...)
}

// ApplyOptions applies the given options to a DeviceConfig and returns it.
// This is a helper function for factory implementations.
func ApplyOptions(opts ...DeviceOption) *DeviceConfig {
	cfg := &DeviceConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
