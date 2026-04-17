// Package ac provides air conditioner device interface and implementation
package ac

import (
	"github.com/RelicOfTesla/midea-msmart/msmart/device/xc"
)

// AC defines the interface for air conditioner devices
// This interface abstracts the AC-specific functionality
type AC interface {
	xc.XCDevice

	// Control settings
	SetEco(bool)
	SetTurbo(bool)
	SetBreezeAway(bool)
	SetBreezeMild(bool)
	SetBreezeless(bool)
	SetHorizontalSwingAngle(SwingAngle)
	SetVerticalSwingAngle(SwingAngle)
	SetCascadeMode(CascadeMode)
	SetIECO(bool)
	SetFlashCool(bool)
	SetBeep(bool)
	SetFahrenheit(bool)

	// Operations
	CapabilitiesDict() map[string]interface{}

	// Energy usage
	SetEnableEnergyUsageRequests(bool)
	GetRealTimePowerUsage(format EnergyDataFormat) *float64
	GetCurrentEnergyUsage(format EnergyDataFormat) *float64
	GetTotalEnergyUsage(format EnergyDataFormat) *float64

	IsAuthenticated() bool
}
