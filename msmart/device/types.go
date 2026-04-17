// Package device provides core device types and interfaces
package device

// DeviceType represents a device type
type DeviceType int

// Device type constants
const (
	DeviceTypeAirConditioner DeviceType = 0xAC
	DeviceTypeCommercialAC   DeviceType = 0xCC
)

// Token represents authentication token
type Token []byte

// Key represents encryption key
type Key []byte

type LocalKey []byte
