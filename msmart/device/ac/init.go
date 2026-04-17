package ac

import (
	"github.com/RelicOfTesla/midea-msmart/msmart/device"
)

func init() {
	device.RegisterDeviceType(device.DeviceTypeAirConditioner, func(opts ...device.DeviceOption) device.Device {
		// Apply options to get device configuration
		cfg := device.ApplyOptions(opts...)

		// Extract required parameters with defaults
		ip := ""
		if cfg.DeviceIP != nil {
			ip = *cfg.DeviceIP
		}

		port := 6444
		if cfg.DevicePort != nil {
			port = *cfg.DevicePort
		}

		deviceID := ""
		if cfg.DeviceID != nil {
			deviceID = *cfg.DeviceID
		}

		// Create AirConditioner using the factory function
		return NewAirConditioner(ip, port, deviceID, opts...)
	})
}
