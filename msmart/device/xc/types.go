package xc

import (
	"fmt"

	"github.com/RelicOfTesla/midea-msmart/msmart/device"
)

// 空调设备
type XCDevice interface {
	device.SimpleDevice
	SetTargetTemperature(float64)
	SetOperationalMode(OperationalMode)

	// State queries
	TargetTemperature() float64
	IndoorTemperature() *float64
	OutdoorTemperature() *float64
	OperationalMode() OperationalMode
	FanSpeed() FanSpeed
	SwingMode() SwingMode
	Eco() bool
	Turbo() bool

	// control
	SetFanSpeed(FanSpeed)
	SetSwingMode(SwingMode)
}

// OperationalMode represents operational mode enum
type OperationalMode int

const (
	OperationalModeAuto     OperationalMode = 1
	OperationalModeCool     OperationalMode = 2
	OperationalModeDry      OperationalMode = 3
	OperationalModeHeat     OperationalMode = 4
	OperationalModeFanOnly  OperationalMode = 5
	OperationalModeSmartDry OperationalMode = 6

	OperationalModeDefault OperationalMode = OperationalModeFanOnly
)

// String returns the string representation of OperationalMode
func (om OperationalMode) String() string {
	switch om {
	case OperationalModeAuto:
		return "AUTO"
	case OperationalModeCool:
		return "COOL"
	case OperationalModeDry:
		return "DRY"
	case OperationalModeHeat:
		return "HEAT"
	case OperationalModeFanOnly:
		return "FAN_ONLY"
	case OperationalModeSmartDry:
		return "SMART_DRY"
	default:
		return fmt.Sprintf("OperationalMode(%d)", int(om))
	}
}

// Values returns all valid OperationalMode values
func (OperationalMode) Values() []OperationalMode {
	return []OperationalMode{
		OperationalModeAuto, OperationalModeCool, OperationalModeDry,
		OperationalModeHeat, OperationalModeFanOnly, OperationalModeSmartDry,
	}
}

// GetFromValue returns the OperationalMode for a given value, or the default if not found
func (OperationalMode) GetFromValue(value int) OperationalMode {
	for _, om := range OperationalMode(0).Values() {
		if int(om) == value {
			return om
		}
	}
	return OperationalModeDefault
}

// GetFromName returns the OperationalMode for a given name, or the default if not found
func (OperationalMode) GetFromName(name string) OperationalMode {
	switch name {
	case "AUTO":
		return OperationalModeAuto
	case "COOL":
		return OperationalModeCool
	case "DRY":
		return OperationalModeDry
	case "HEAT":
		return OperationalModeHeat
	case "FAN_ONLY":
		return OperationalModeFanOnly
	case "SMART_DRY":
		return OperationalModeSmartDry
	default:
		return OperationalModeDefault
	}
}

///////

// SwingMode represents swing mode enum
type SwingMode int

const (
	SwingModeOff        SwingMode = 0x0
	SwingModeVertical   SwingMode = 0xC
	SwingModeHorizontal SwingMode = 0x3
	SwingModeBoth       SwingMode = 0xF

	SwingModeDefault SwingMode = SwingModeOff
)

// String returns the string representation of SwingMode
func (sm SwingMode) String() string {
	switch sm {
	case SwingModeOff:
		return "OFF"
	case SwingModeVertical:
		return "VERTICAL"
	case SwingModeHorizontal:
		return "HORIZONTAL"
	case SwingModeBoth:
		return "BOTH"
	default:
		return fmt.Sprintf("SwingMode(%d)", int(sm))
	}
}

// Values returns all valid SwingMode values
func (SwingMode) Values() []SwingMode {
	return []SwingMode{
		SwingModeOff, SwingModeVertical, SwingModeHorizontal, SwingModeBoth,
	}
}

// GetFromValue returns the SwingMode for a given value, or the default if not found
func (SwingMode) GetFromValue(value int) SwingMode {
	for _, sm := range SwingMode(0).Values() {
		if int(sm) == value {
			return sm
		}
	}
	return SwingModeDefault
}

// GetFromName returns the SwingMode for a given name, or the default if not found
func (SwingMode) GetFromName(name string) SwingMode {
	switch name {
	case "OFF":
		return SwingModeOff
	case "VERTICAL":
		return SwingModeVertical
	case "HORIZONTAL":
		return SwingModeHorizontal
	case "BOTH":
		return SwingModeBoth
	default:
		return SwingModeDefault
	}
}

// FanSpeed represents fan speed enum
type FanSpeed int

const (
	FanSpeedAuto   FanSpeed = 102
	FanSpeedMax    FanSpeed = 100
	FanSpeedHigh   FanSpeed = 80
	FanSpeedMedium FanSpeed = 60
	FanSpeedLow    FanSpeed = 40
	FanSpeedSilent FanSpeed = 20

	FanSpeedDefault FanSpeed = FanSpeedAuto
)

// String returns the string representation of FanSpeed
func (fs FanSpeed) String() string {
	switch fs {
	case FanSpeedAuto:
		return "AUTO"
	case FanSpeedMax:
		return "MAX"
	case FanSpeedHigh:
		return "HIGH"
	case FanSpeedMedium:
		return "MEDIUM"
	case FanSpeedLow:
		return "LOW"
	case FanSpeedSilent:
		return "SILENT"
	default:
		return fmt.Sprintf("FanSpeed(%d)", int(fs))
	}
}

// Values returns all valid FanSpeed values
func (FanSpeed) Values() []FanSpeed {
	return []FanSpeed{
		FanSpeedAuto, FanSpeedMax, FanSpeedHigh,
		FanSpeedMedium, FanSpeedLow, FanSpeedSilent,
	}
}

// GetFromValue returns the FanSpeed for a given value, or the default if not found
func (FanSpeed) GetFromValue(value int) FanSpeed {
	for _, fs := range FanSpeed(0).Values() {
		if int(fs) == value {
			return fs
		}
	}
	return FanSpeedDefault
}

// GetFromName returns the FanSpeed for a given name, or the default if not found
func (FanSpeed) GetFromName(name string) FanSpeed {
	switch name {
	case "AUTO":
		return FanSpeedAuto
	case "MAX":
		return FanSpeedMax
	case "HIGH":
		return FanSpeedHigh
	case "MEDIUM":
		return FanSpeedMedium
	case "LOW":
		return FanSpeedLow
	case "SILENT":
		return FanSpeedSilent
	default:
		return FanSpeedDefault
	}
}
