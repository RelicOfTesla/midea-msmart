package msmart

import "github.com/RelicOfTesla/midea-msmart/msmart/device"

// Type aliases for device types
// These aliases provide backward compatibility for code that uses msmart.DeviceType, etc.
type DeviceType = device.DeviceType
type DeviceOption = device.DeviceOption

// Functional option wrappers
// These functions provide backward compatibility for code that uses msmart.WithName, etc.
var (
	WithName       = device.WithName
	WithVersion    = device.WithVersion
	WithSN         = device.WithSN
	WithDeviceID   = device.WithDeviceID
	WithDeviceAddr = device.WithDeviceAddr
	WithTokenKey   = device.WithTokenKey
	WithLocalKey   = device.WithLocalKey
)
