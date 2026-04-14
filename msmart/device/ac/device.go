// Package ac provides air conditioner device implementation.
// This is a translation from msmart-ng Python library.
// Original file: msmart/device/AC/device.py
package ac

import (
	"context"
	"fmt"
	"log/slog"

	msmart "midea-go/pkg/msmart_ng_go"
)

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

// SwingMode represents swing mode enum
type SwingMode int

const (
	SwingModeOff       SwingMode = 0x0
	SwingModeVertical  SwingMode = 0xC
	SwingModeHorizontal SwingMode = 0x3
	SwingModeBoth      SwingMode = 0xF

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

// SwingAngle represents swing angle enum
type SwingAngle int

const (
	SwingAngleOff  SwingAngle = 0
	SwingAnglePos1 SwingAngle = 1
	SwingAnglePos2 SwingAngle = 25
	SwingAnglePos3 SwingAngle = 50
	SwingAnglePos4 SwingAngle = 75
	SwingAnglePos5 SwingAngle = 100

	SwingAngleDefault SwingAngle = SwingAngleOff
)

// String returns the string representation of SwingAngle
func (sa SwingAngle) String() string {
	switch sa {
	case SwingAngleOff:
		return "OFF"
	case SwingAnglePos1:
		return "POS_1"
	case SwingAnglePos2:
		return "POS_2"
	case SwingAnglePos3:
		return "POS_3"
	case SwingAnglePos4:
		return "POS_4"
	case SwingAnglePos5:
		return "POS_5"
	default:
		return fmt.Sprintf("SwingAngle(%d)", int(sa))
	}
}

// Values returns all valid SwingAngle values
func (SwingAngle) Values() []SwingAngle {
	return []SwingAngle{
		SwingAngleOff, SwingAnglePos1, SwingAnglePos2,
		SwingAnglePos3, SwingAnglePos4, SwingAnglePos5,
	}
}

// GetFromValue returns the SwingAngle for a given value, or the default if not found
func (SwingAngle) GetFromValue(value int) SwingAngle {
	for _, sa := range SwingAngle(0).Values() {
		if int(sa) == value {
			return sa
		}
	}
	return SwingAngleDefault
}

// CascadeMode represents cascade mode enum
type CascadeMode int

const (
	CascadeModeOff  CascadeMode = 0
	CascadeModeUp   CascadeMode = 1
	CascadeModeDown CascadeMode = 2

	CascadeModeDefault CascadeMode = CascadeModeOff
)

// String returns the string representation of CascadeMode
func (cm CascadeMode) String() string {
	switch cm {
	case CascadeModeOff:
		return "OFF"
	case CascadeModeUp:
		return "UP"
	case CascadeModeDown:
		return "DOWN"
	default:
		return fmt.Sprintf("CascadeMode(%d)", int(cm))
	}
}

// Values returns all valid CascadeMode values
func (CascadeMode) Values() []CascadeMode {
	return []CascadeMode{CascadeModeOff, CascadeModeUp, CascadeModeDown}
}

// GetFromValue returns the CascadeMode for a given value, or the default if not found
func (CascadeMode) GetFromValue(value int) CascadeMode {
	for _, cm := range CascadeMode(0).Values() {
		if int(cm) == value {
			return cm
		}
	}
	return CascadeModeDefault
}

// RateSelect represents rate select enum
type RateSelect int

const (
	RateSelectOff RateSelect = 100

	// 2 levels
	RateSelectGear50 RateSelect = 50
	RateSelectGear75 RateSelect = 75

	// 5 levels
	RateSelectLevel1 RateSelect = 1
	RateSelectLevel2 RateSelect = 20
	RateSelectLevel3 RateSelect = 40
	RateSelectLevel4 RateSelect = 60
	RateSelectLevel5 RateSelect = 80

	RateSelectDefault RateSelect = RateSelectOff
)

// String returns the string representation of RateSelect
func (rs RateSelect) String() string {
	switch rs {
	case RateSelectOff:
		return "OFF"
	case RateSelectGear50:
		return "GEAR_50"
	case RateSelectGear75:
		return "GEAR_75"
	case RateSelectLevel1:
		return "LEVEL_1"
	case RateSelectLevel2:
		return "LEVEL_2"
	case RateSelectLevel3:
		return "LEVEL_3"
	case RateSelectLevel4:
		return "LEVEL_4"
	case RateSelectLevel5:
		return "LEVEL_5"
	default:
		return fmt.Sprintf("RateSelect(%d)", int(rs))
	}
}

// Values returns all valid RateSelect values
func (RateSelect) Values() []RateSelect {
	return []RateSelect{
		RateSelectOff, RateSelectGear50, RateSelectGear75,
		RateSelectLevel1, RateSelectLevel2, RateSelectLevel3,
		RateSelectLevel4, RateSelectLevel5,
	}
}

// GetFromValue returns the RateSelect for a given value, or the default if not found
func (RateSelect) GetFromValue(value int) RateSelect {
	for _, rs := range RateSelect(0).Values() {
		if int(rs) == value {
			return rs
		}
	}
	return RateSelectDefault
}

// BreezeMode represents breeze mode enum
type BreezeMode int

const (
	BreezeModeOff        BreezeMode = 1
	BreezeModeBreezeAway BreezeMode = 2
	BreezeModeBreezeMild BreezeMode = 3
	BreezeModeBreezeless BreezeMode = 4

	BreezeModeDefault BreezeMode = BreezeModeOff
)

// String returns the string representation of BreezeMode
func (bm BreezeMode) String() string {
	switch bm {
	case BreezeModeOff:
		return "OFF"
	case BreezeModeBreezeAway:
		return "BREEZE_AWAY"
	case BreezeModeBreezeMild:
		return "BREEZE_MILD"
	case BreezeModeBreezeless:
		return "BREEZELESS"
	default:
		return fmt.Sprintf("BreezeMode(%d)", int(bm))
	}
}

// Values returns all valid BreezeMode values
func (BreezeMode) Values() []BreezeMode {
	return []BreezeMode{
		BreezeModeOff, BreezeModeBreezeAway, BreezeModeBreezeMild, BreezeModeBreezeless,
	}
}

// GetFromValue returns the BreezeMode for a given value, or the default if not found
func (BreezeMode) GetFromValue(value int) BreezeMode {
	for _, bm := range BreezeMode(0).Values() {
		if int(bm) == value {
			return bm
		}
	}
	return BreezeModeDefault
}

// AuxHeatMode represents aux heat mode enum
type AuxHeatMode int

const (
	AuxHeatModeOff     AuxHeatMode = 0
	AuxHeatModeAuxHeat AuxHeatMode = 1
	AuxHeatModeAuxOnly AuxHeatMode = 2

	AuxHeatModeDefault AuxHeatMode = AuxHeatModeOff
)

// String returns the string representation of AuxHeatMode
func (ahm AuxHeatMode) String() string {
	switch ahm {
	case AuxHeatModeOff:
		return "OFF"
	case AuxHeatModeAuxHeat:
		return "AUX_HEAT"
	case AuxHeatModeAuxOnly:
		return "AUX_ONLY"
	default:
		return fmt.Sprintf("AuxHeatMode(%d)", int(ahm))
	}
}

// Values returns all valid AuxHeatMode values
func (AuxHeatMode) Values() []AuxHeatMode {
	return []AuxHeatMode{AuxHeatModeOff, AuxHeatModeAuxHeat, AuxHeatModeAuxOnly}
}

// GetFromValue returns the AuxHeatMode for a given value, or the default if not found
func (AuxHeatMode) GetFromValue(value int) AuxHeatMode {
	for _, ahm := range AuxHeatMode(0).Values() {
		if int(ahm) == value {
			return ahm
		}
	}
	return AuxHeatModeDefault
}

// EnergyDataFormat represents energy data format enum
type EnergyDataFormat int

const (
	EnergyDataFormatBCD    EnergyDataFormat = 0
	EnergyDataFormatBinary EnergyDataFormat = 1
)

// String returns the string representation of EnergyDataFormat
func (edf EnergyDataFormat) String() string {
	switch edf {
	case EnergyDataFormatBCD:
		return "BCD"
	case EnergyDataFormatBinary:
		return "BINARY"
	default:
		return fmt.Sprintf("EnergyDataFormat(%d)", int(edf))
	}
}

// Capability represents device capability flags
// This is a translation of Python's Flag enum
type Capability int64

const (
	// Fan
	CapabilityCustomFanSpeed Capability = 1 << iota

	// Presets
	CapabilityEco
	CapabilityFreezeProtection
	CapabilityIECO
	CapabilityTurbo

	// UI
	CapabilityDisplayControl
	CapabilityEnergyStats
	CapabilityFilterReminder

	CapabilityHumidity
	CapabilityTargetHumidity

	// Swing
	CapabilitySwingHorizontalAngle
	CapabilitySwingVerticalAngle

	// Breeze control
	CapabilityBreezeAway
	CapabilityBreezeControl
	CapabilityBreezeless

	// Misc
	CapabilityCascade
	CapabilityJetCool
	CapabilityOutSilent
	CapabilityPurifier
	CapabilitySelfClean

	// Default capabilities
	CapabilityDefault Capability = CapabilityCustomFanSpeed |
		CapabilityEco | CapabilityTurbo | CapabilityFreezeProtection |
		CapabilityDisplayControl | CapabilityFilterReminder |
		CapabilityPurifier
)

// Has checks if a capability flag is set
func (c Capability) Has(flag Capability) bool {
	return c&flag != 0
}

// Set enables or disables a capability flag
func (c *Capability) Set(flag Capability, enable bool) {
	if enable {
		*c |= flag
	} else {
		*c &= ^flag
	}
}

// String returns the string representation of Capability
func (c Capability) String() string {
	var flags []string
	if c.Has(CapabilityCustomFanSpeed) {
		flags = append(flags, "CUSTOM_FAN_SPEED")
	}
	if c.Has(CapabilityEco) {
		flags = append(flags, "ECO")
	}
	if c.Has(CapabilityFreezeProtection) {
		flags = append(flags, "FREEZE_PROTECTION")
	}
	if c.Has(CapabilityIECO) {
		flags = append(flags, "IECO")
	}
	if c.Has(CapabilityTurbo) {
		flags = append(flags, "TURBO")
	}
	if c.Has(CapabilityDisplayControl) {
		flags = append(flags, "DISPLAY_CONTROL")
	}
	if c.Has(CapabilityEnergyStats) {
		flags = append(flags, "ENERGY_STATS")
	}
	if c.Has(CapabilityFilterReminder) {
		flags = append(flags, "FILTER_REMINDER")
	}
	if c.Has(CapabilityHumidity) {
		flags = append(flags, "HUMIDITY")
	}
	if c.Has(CapabilityTargetHumidity) {
		flags = append(flags, "TARGET_HUMIDITY")
	}
	if c.Has(CapabilitySwingHorizontalAngle) {
		flags = append(flags, "SWING_HORIZONTAL_ANGLE")
	}
	if c.Has(CapabilitySwingVerticalAngle) {
		flags = append(flags, "SWING_VERTICAL_ANGLE")
	}
	if c.Has(CapabilityBreezeAway) {
		flags = append(flags, "BREEZE_AWAY")
	}
	if c.Has(CapabilityBreezeControl) {
		flags = append(flags, "BREEZE_CONTROL")
	}
	if c.Has(CapabilityBreezeless) {
		flags = append(flags, "BREEZELESS")
	}
	if c.Has(CapabilityCascade) {
		flags = append(flags, "CASCADE")
	}
	if c.Has(CapabilityJetCool) {
		flags = append(flags, "JET_COOL")
	}
	if c.Has(CapabilityOutSilent) {
		flags = append(flags, "OUT_SILENT")
	}
	if c.Has(CapabilityPurifier) {
		flags = append(flags, "PURIFIER")
	}
	if c.Has(CapabilitySelfClean) {
		flags = append(flags, "SELF_CLEAN")
	}
	return fmt.Sprintf("%v", flags)
}

// AirConditioner represents an air conditioner device
// This is the main struct that translates Python's AirConditioner class
type AirConditioner struct {
	*msmart.Device

	// Basic controls
	beepOn           bool
	powerState       *bool
	targetTemperature float64
	targetHumidity   int

	operationalMode OperationalMode
	fanSpeed        interface{} // FanSpeed or int for custom speeds
	swingMode       SwingMode

	eco               bool
	turbo             bool
	freezeProtection bool
	sleep             bool

	fahrenheitUnit bool // Display temperature in Fahrenheit
	displayOn       *bool

	// Advanced controls
	followMe   bool
	purifier   bool
	ieco       bool
	flashCool  bool
	outSilent  bool

	horizontalSwingAngle SwingAngle
	verticalSwingAngle   SwingAngle
	cascadeMode          CascadeMode
	rateSelect           RateSelect
	breezeMode           BreezeMode
	auxMode              AuxHeatMode

	// Sensors
	indoorTemperature  *float64
	indoorHumidity     *int
	outdoorTemperature *float64

	filterAlert       *bool
	errorCode         *int
	selfCleanActive   *bool
	defrostActive     *bool
	outdoorFanSpeed   *int

	totalEnergyUsage   map[EnergyDataFormat]*float64
	currentEnergyUsage map[EnergyDataFormat]*float64
	realTimePowerUsage map[EnergyDataFormat]*float64
	useBinaryEnergy    bool // Deprecated

	// Capabilities
	minTargetTemperature float64
	maxTargetTemperature float64

	capabilities *msmart.CapabilityManager

	supportedOpModes    []OperationalMode
	supportedSwingModes []SwingMode
	supportedFanSpeeds  []interface{} // FanSpeed or int
	supportedRateSelects []RateSelect
	supportedAuxModes    []AuxHeatMode

	// Misc
	requestEnergyUsage  bool
	requestGroup5Data   bool

	// Supported properties
	supportedProperties map[PropertyId]bool
	updatedProperties   map[PropertyId]bool
}

// NewAirConditioner creates a new AirConditioner instance
// This is the Go equivalent of Python's __init__ method
func NewAirConditioner(ip string, port int, deviceID int, opts ...msmart.DeviceOption) *AirConditioner {
	ac := &AirConditioner{
		Device: msmart.NewDevice(ip, port, deviceID, msmart.DeviceTypeAirConditioner, opts...),

		// Basic controls
		beepOn:           false,
		powerState:       nil,
		targetTemperature: 17.0,
		targetHumidity:   40,

		operationalMode: OperationalModeAuto,
		fanSpeed:        FanSpeedAuto,
		swingMode:       SwingModeOff,

		eco:               false,
		turbo:             false,
		freezeProtection: false,
		sleep:             false,

		fahrenheitUnit: false,
		displayOn:       nil,

		// Advanced controls
		followMe:   false,
		purifier:   false,
		ieco:       false,
		flashCool:  false,
		outSilent:  false,

		horizontalSwingAngle: SwingAngleOff,
		verticalSwingAngle:   SwingAngleOff,
		cascadeMode:          CascadeModeOff,
		rateSelect:           RateSelectOff,
		breezeMode:           BreezeModeOff,
		auxMode:              AuxHeatModeOff,

		// Sensors
		indoorTemperature:  nil,
		indoorHumidity:     nil,
		outdoorTemperature: nil,

		filterAlert:       nil,
		errorCode:         nil,
		selfCleanActive:   nil,
		defrostActive:     nil,
		outdoorFanSpeed:   nil,

		totalEnergyUsage:   make(map[EnergyDataFormat]*float64),
		currentEnergyUsage: make(map[EnergyDataFormat]*float64),
		realTimePowerUsage: make(map[EnergyDataFormat]*float64),
		useBinaryEnergy:    false,

		// Capabilities
		minTargetTemperature: 16,
		maxTargetTemperature: 30,

		capabilities: msmart.NewCapabilityManager(int64(CapabilityDefault)),

		supportedOpModes:    OperationalMode(0).Values(),
		supportedSwingModes: SwingMode(0).Values(),
		supportedFanSpeeds:  make([]interface{}, 0),
		supportedRateSelects: []RateSelect{RateSelectOff},
		supportedAuxModes:    []AuxHeatMode{AuxHeatModeOff},

		// Misc
		requestEnergyUsage: false,
		requestGroup5Data:  false,

		// Supported properties
		supportedProperties: make(map[PropertyId]bool),
		updatedProperties:   make(map[PropertyId]bool),
	}

	// Initialize fan speeds
	for _, fs := range FanSpeed(0).Values() {
		ac.supportedFanSpeeds = append(ac.supportedFanSpeeds, fs)
	}

	return ac
}

// updateState updates the local state from a device state response
// This is the Go equivalent of Python's _update_state method
func (ac *AirConditioner) updateState(res ResponseInterface) {
	switch r := res.(type) {
	case *StateResponse:
		slog.Debug("State response payload from device", "id", ac.GetID(), "response", r)

		ac.powerState = r.PowerOn
		ac.targetTemperature = ptrToVal(r.TargetTemperature, ac.targetTemperature)
		ac.operationalMode = OperationalMode(0).GetFromValue(int(ptrToVal(r.OperationalMode, byte(ac.operationalMode))))

		if ac.supportsCustomFanSpeed() {
			// Attempt to use fan speed enum, but fallback to raw int if custom
			ac.fanSpeed = ptrToVal(r.FanSpeed, byte(FanSpeedAuto))
		} else {
			ac.fanSpeed = FanSpeed(0).GetFromValue(int(ptrToVal(r.FanSpeed, byte(FanSpeedAuto))))
		}

		ac.swingMode = SwingMode(0).GetFromValue(int(ptrToVal(r.SwingMode, byte(ac.swingMode))))

		ac.eco = ptrToVal(r.Eco, ac.eco)
		ac.turbo = ptrToVal(r.Turbo, ac.turbo)
		ac.freezeProtection = ptrToVal(r.FreezeProtection, ac.freezeProtection)
		ac.sleep = ptrToVal(r.Sleep, ac.sleep)

		ac.indoorTemperature = r.IndoorTemperature
		ac.outdoorTemperature = r.OutdoorTemperature

		ac.displayOn = r.DisplayOn
		if r.Fahrenheit != nil {
			ac.fahrenheitUnit = *r.Fahrenheit
		}

		ac.filterAlert = r.FilterAlert

		ac.followMe = ptrToVal(r.FollowMe, ac.followMe)
		ac.purifier = ptrToVal(r.Purifier, ac.purifier)

		ac.targetHumidity = int(ptrToVal(r.TargetHumidity, byte(ac.targetHumidity)))

		if r.IndependentAuxHeat != nil && *r.IndependentAuxHeat {
			ac.auxMode = AuxHeatModeAuxOnly
		} else if r.AuxHeat != nil && *r.AuxHeat {
			ac.auxMode = AuxHeatModeAuxHeat
		} else {
			ac.auxMode = AuxHeatModeOff
		}

		ac.errorCode = bytePtrToIntPtr(r.ErrorCode)

	case *PropertiesResponse:
		slog.Debug("Properties response payload from device", "id", ac.GetID(), "response", r)

		if angle := r.GetProperty(PropertyIdSwingLrAngle); angle != nil {
			ac.horizontalSwingAngle = SwingAngle(0).GetFromValue(toInt(angle))
		}

		if angle := r.GetProperty(PropertyIdSwingUdAngle); angle != nil {
			ac.verticalSwingAngle = SwingAngle(0).GetFromValue(toInt(angle))
		}

		if cascade := r.GetProperty(PropertyIdCascade); cascade != nil {
			ac.cascadeMode = CascadeMode(0).GetFromValue(toInt(cascade))
		}

		if value := r.GetProperty(PropertyIdSelfClean); value != nil {
			if b, ok := value.(bool); ok {
				ac.selfCleanActive = &b
			}
		}

		if rate := r.GetProperty(PropertyIdRateSelect); rate != nil {
			ac.rateSelect = RateSelect(0).GetFromValue(toInt(rate))
		}

		// Breeze control supersedes breeze away and breezeless
		if value := r.GetProperty(PropertyIdBreezeControl); value != nil {
			ac.breezeMode = BreezeMode(0).GetFromValue(toInt(value))
		} else {
			if value := r.GetProperty(PropertyIdBreezeAway); value != nil {
				if b, ok := value.(bool); ok && b {
					ac.breezeMode = BreezeModeBreezeAway
				} else {
					ac.breezeMode = BreezeModeOff
				}
			}

			if value := r.GetProperty(PropertyIdBreezeless); value != nil {
				if b, ok := value.(bool); ok && b {
					ac.breezeMode = BreezeModeBreezeless
				} else {
					ac.breezeMode = BreezeModeOff
				}
			}
		}

		if value := r.GetProperty(PropertyIdIECO); value != nil {
			if b, ok := value.(bool); ok {
				ac.ieco = b
			}
		}

		if value := r.GetProperty(PropertyIdJetCool); value != nil {
			if b, ok := value.(bool); ok {
				ac.flashCool = b
			}
		}

		if value := r.GetProperty(PropertyIdOutSilent); value != nil {
			if b, ok := value.(bool); ok {
				ac.outSilent = b
			}
		}

	case *EnergyUsageResponse:
		slog.Debug("Energy response payload from device", "id", ac.GetID(), "response", r)

		ac.totalEnergyUsage[EnergyDataFormatBCD] = r.TotalEnergy
		ac.totalEnergyUsage[EnergyDataFormatBinary] = r.TotalEnergyBinary
		ac.currentEnergyUsage[EnergyDataFormatBCD] = r.CurrentEnergy
		ac.currentEnergyUsage[EnergyDataFormatBinary] = r.CurrentEnergyBinary
		ac.realTimePowerUsage[EnergyDataFormatBCD] = r.RealTimePower
		ac.realTimePowerUsage[EnergyDataFormatBinary] = r.RealTimePowerBinary

	case *Group5Response:
		slog.Debug("Group 5 response payload from device", "id", ac.GetID(), "response", r)

		ac.indoorHumidity = bytePtrToIntPtr(r.Humidity)
		ac.outdoorFanSpeed = bytePtrToIntPtr(r.OutdoorFanSpeed)
		ac.defrostActive = r.Defrost

	default:
		slog.Debug("Ignored unknown response from device", "id", ac.GetID(), "response", res)
	}
}

// updateCapabilities updates device capabilities from a CapabilitiesResponse
// This is the Go equivalent of Python's _update_capabilities method
func (ac *AirConditioner) updateCapabilities(res *CapabilitiesResponse) {
	// Build list of supported operation modes
	opModes := []OperationalMode{OperationalModeFanOnly}
	if res.DryMode() {
		opModes = append(opModes, OperationalModeDry)
	}
	if res.CoolMode() {
		opModes = append(opModes, OperationalModeCool)
	}
	if res.HeatMode() {
		opModes = append(opModes, OperationalModeHeat)
	}
	if res.AutoMode() {
		opModes = append(opModes, OperationalModeAuto)
	}
	if res.TargetHumidity() {
		// Add SMART_DRY support if target humidity is supported
		opModes = append(opModes, OperationalModeSmartDry)
	}
	ac.supportedOpModes = opModes

	// Build list of supported swing modes
	swingModes := []SwingMode{SwingModeOff}
	if res.SwingHorizontal() {
		swingModes = append(swingModes, SwingModeHorizontal)
	}
	if res.SwingVertical() {
		swingModes = append(swingModes, SwingModeVertical)
	}
	if res.SwingBoth() {
		swingModes = append(swingModes, SwingModeBoth)
	}
	ac.supportedSwingModes = swingModes

	// Build list of supported fan speeds
	fanSpeeds := make([]interface{}, 0)
	if res.FanSilent() {
		fanSpeeds = append(fanSpeeds, FanSpeedSilent)
	}
	if res.FanLow() {
		fanSpeeds = append(fanSpeeds, FanSpeedLow)
	}
	if res.FanMedium() {
		fanSpeeds = append(fanSpeeds, FanSpeedMedium)
	}
	if res.FanHigh() {
		fanSpeeds = append(fanSpeeds, FanSpeedHigh)
	}
	if res.FanAuto() {
		fanSpeeds = append(fanSpeeds, FanSpeedAuto)
	}
	if res.FanCustom() {
		// Include additional MAX speed if custom speeds are supported
		fanSpeeds = append(fanSpeeds, FanSpeedMax)
	}
	ac.supportedFanSpeeds = fanSpeeds

	ac.capabilities.Set(int64(CapabilityCustomFanSpeed), res.FanCustom())

	ac.capabilities.Set(int64(CapabilityEco), res.Eco())
	ac.capabilities.Set(int64(CapabilityTurbo), res.Turbo())
	ac.capabilities.Set(int64(CapabilityFreezeProtection), res.FreezeProtection())

	ac.capabilities.Set(int64(CapabilityDisplayControl), res.DisplayControl())
	ac.capabilities.Set(int64(CapabilityFilterReminder), res.FilterReminder())

	ac.capabilities.Set(int64(CapabilityPurifier), res.Anion())

	// Build list of supported aux heating modes
	auxModes := []AuxHeatMode{AuxHeatModeOff}
	if res.AuxElectricHeat() || res.AuxHeatMode() {
		auxModes = append(auxModes, AuxHeatModeAuxHeat)
	}
	if res.AuxMode() {
		auxModes = append(auxModes, AuxHeatModeAuxOnly)
	}
	ac.supportedAuxModes = auxModes

	ac.minTargetTemperature = float64(res.MinTemperature())
	ac.maxTargetTemperature = float64(res.MaxTemperature())

	// Allow capabilities to enable energy usage requests, but not disable them
	// We've seen devices that claim no capability but return energy data
	ac.requestEnergyUsage = ac.requestEnergyUsage || res.EnergyStats()

	ac.capabilities.Set(int64(CapabilityHumidity), res.Humidity())
	ac.capabilities.Set(int64(CapabilityTargetHumidity), res.TargetHumidity())

	ac.capabilities.Set(int64(CapabilitySwingVerticalAngle), res.SwingVerticalAngle())
	ac.capabilities.Set(int64(CapabilitySwingHorizontalAngle), res.SwingHorizontalAngle())

	ac.capabilities.Set(int64(CapabilityCascade), res.Cascade())

	ac.capabilities.Set(int64(CapabilitySelfClean), res.SelfClean())

	// Add supported rate select levels
	if rates := res.RateSelectLevels(); rates != nil {
		if *rates > 2 {
			ac.supportedRateSelects = []RateSelect{
				RateSelectOff,
				RateSelectLevel5,
				RateSelectLevel4,
				RateSelectLevel3,
				RateSelectLevel2,
				RateSelectLevel1,
			}
		} else {
			ac.supportedRateSelects = []RateSelect{
				RateSelectOff,
				RateSelectGear75,
				RateSelectGear50,
			}
		}
	}

	// Breeze control supersedes breeze away and breezeless
	ac.capabilities.Set(int64(CapabilityBreezeControl), res.BreezeControl())
	if !res.BreezeControl() {
		ac.capabilities.Set(int64(CapabilityBreezeAway), res.BreezeAway())
		ac.capabilities.Set(int64(CapabilityBreezeless), res.Breezeless())
	}

	ac.capabilities.Set(int64(CapabilityIECO), res.Ieco())
	ac.capabilities.Set(int64(CapabilityJetCool), res.JetCool())

	ac.capabilities.Set(int64(CapabilityOutSilent), res.OutSilent())

	// Update supported properties from capabilities
	ac.updateSupportedProperties()
}

// updateSupportedProperties updates supported properties based on device capabilities
// This is the Go equivalent of Python's _update_supported_properties method
func (ac *AirConditioner) updateSupportedProperties() {
	// Map of capability flag to property ID
	capabilityMap := map[Capability]PropertyId{
		CapabilityBreezeAway:           PropertyIdBreezeAway,
		CapabilityBreezeControl:        PropertyIdBreezeControl,
		CapabilityBreezeless:           PropertyIdBreezeless,
		CapabilityCascade:              PropertyIdCascade,
		CapabilityIECO:                 PropertyIdIECO,
		CapabilityJetCool:              PropertyIdJetCool,
		CapabilityOutSilent:            PropertyIdOutSilent,
		CapabilitySelfClean:            PropertyIdSelfClean,
		CapabilitySwingHorizontalAngle: PropertyIdSwingLrAngle,
		CapabilitySwingVerticalAngle:   PropertyIdSwingUdAngle,
	}

	// Clear existing properties
	ac.supportedProperties = make(map[PropertyId]bool)

	// Test each capability
	caps := Capability(ac.capabilities.Flags())
	for cap, prop := range capabilityMap {
		if caps.Has(cap) {
			ac.supportedProperties[prop] = true
		}
	}

	// Rate select is a special case. It's property based but not controlled by a capability flag
	if len(ac.supportedRateSelects) > 1 {
		ac.supportedProperties[PropertyIdRateSelect] = true
	}
}

// sendCommandsGetResponses sends a list of commands and returns all valid responses
// This is the Go equivalent of Python's _send_commands_get_responses method
func (ac *AirConditioner) sendCommandsGetResponses(ctx context.Context, commands []CommandInterface) ([]ResponseInterface, error) {
	// TODO: Implement actual command sending
	// This is a placeholder for the translation

	return nil, fmt.Errorf("not implemented")
}

// GetCapabilities fetches the device capabilities
// This is the Go equivalent of Python's get_capabilities method
func (ac *AirConditioner) GetCapabilities(ctx context.Context) error {
	// TODO: Implement actual capabilities fetching
	// This is a placeholder for the translation

	return fmt.Errorf("not implemented")
}

// ToggleDisplay toggles the device display if the device supports it
// This is the Go equivalent of Python's toggle_display method
func (ac *AirConditioner) ToggleDisplay(ctx context.Context) error {
	if !ac.SupportsDisplayControl() {
		slog.Warn("Device is not capable of display control", "id", ac.GetID())
	}

	// TODO: Implement actual display toggle
	// This is a placeholder for the translation

	return fmt.Errorf("not implemented")
}

// StartSelfClean starts a self cleaning if the device supports it
// This is the Go equivalent of Python's start_self_clean method
func (ac *AirConditioner) StartSelfClean(ctx context.Context) error {
	// Start self clean via properties command
	return ac.applyProperties(ctx, map[PropertyId]interface{}{
		PropertyIdSelfClean: true,
	})
}

// Refresh refreshes the local copy of the device state by sending a GetState command
// This is the Go equivalent of Python's refresh method
func (ac *AirConditioner) Refresh(ctx context.Context) error {
	// TODO: Implement actual refresh
	// This is a placeholder for the translation

	return fmt.Errorf("not implemented")
}

// applyProperties applies the provided properties to the device
// This is the Go equivalent of Python's _apply_properties method
func (ac *AirConditioner) applyProperties(ctx context.Context, properties map[PropertyId]interface{}) error {
	// Warn if attempting to update a property that isn't supported
	for prop := range properties {
		if !ac.supportedProperties[prop] {
			slog.Warn("Device is not capable of property", "id", ac.GetID(), "property", prop)
		}
	}

	// Always add buzzer property
	properties[PropertyIdBuzzer] = ac.beepOn

	// TODO: Build command with properties and send

	return fmt.Errorf("not implemented")
}

// Apply applies the local state to the device
// This is the Go equivalent of Python's apply method
func (ac *AirConditioner) Apply(ctx context.Context) error {
	// Warn if trying to apply unsupported modes
	if !containsOpMode(ac.supportedOpModes, ac.operationalMode) {
		slog.Warn("Device is not capable of operational mode", "id", ac.GetID(), "mode", ac.operationalMode)
	}

	if !ac.supportsFanSpeed(ac.fanSpeed) && !ac.supportsCustomFanSpeed() {
		slog.Warn("Device is not capable of fan speed", "id", ac.GetID(), "speed", ac.fanSpeed)
	}

	if !containsSwingMode(ac.supportedSwingModes, ac.swingMode) {
		slog.Warn("Device is not capable of swing mode", "id", ac.GetID(), "mode", ac.swingMode)
	}

	if ac.turbo && !ac.SupportsTurbo() {
		slog.Warn("Device is not capable of turbo mode", "id", ac.GetID())
	}

	if ac.eco && !ac.SupportsEco() {
		slog.Warn("Device is not capable of eco mode", "id", ac.GetID())
	}

	if ac.freezeProtection && !ac.SupportsFreezeProtection() {
		slog.Warn("Device is not capable of freeze protection", "id", ac.GetID())
	}

	if ac.rateSelect != RateSelectOff && !containsRateSelect(ac.supportedRateSelects, ac.rateSelect) {
		slog.Warn("Device is not capable of rate select", "id", ac.GetID(), "rate", ac.rateSelect)
	}

	if ac.auxMode != AuxHeatModeOff && !containsAuxMode(ac.supportedAuxModes, ac.auxMode) {
		slog.Warn("Device is not capable of aux mode", "mode", ac.auxMode)
	}

	// TODO: Build and send SetStateCommand

	return fmt.Errorf("not implemented")
}

// OverrideCapabilities overrides device capabilities via serialized dict
// This is the Go equivalent of Python's override_capabilities method
func (ac *AirConditioner) OverrideCapabilities(overrides map[string]interface{}, merge bool) error {
	// TODO: Implement capability overrides

	return fmt.Errorf("not implemented")
}

// ============================================================================
// Property getters and setters
// ============================================================================

// Beep returns whether beep is enabled
func (ac *AirConditioner) Beep() bool {
	return ac.beepOn
}

// SetBeep sets whether beep is enabled
func (ac *AirConditioner) SetBeep(tone bool) {
	ac.beepOn = tone
}

// PowerState returns the power state
func (ac *AirConditioner) PowerState() *bool {
	return ac.powerState
}

// SetPowerState sets the power state
func (ac *AirConditioner) SetPowerState(state bool) {
	ac.powerState = &state
}

// Fahrenheit returns whether Fahrenheit unit is enabled
func (ac *AirConditioner) Fahrenheit() *bool {
	return &ac.fahrenheitUnit
}

// SetFahrenheit sets whether Fahrenheit unit is enabled
func (ac *AirConditioner) SetFahrenheit(enabled bool) {
	ac.fahrenheitUnit = enabled
}

// MinTargetTemperature returns the minimum target temperature
func (ac *AirConditioner) MinTargetTemperature() float64 {
	return ac.minTargetTemperature
}

// MaxTargetTemperature returns the maximum target temperature
func (ac *AirConditioner) MaxTargetTemperature() float64 {
	return ac.maxTargetTemperature
}

// TargetTemperature returns the target temperature
func (ac *AirConditioner) TargetTemperature() float64 {
	return ac.targetTemperature
}

// SetTargetTemperature sets the target temperature
func (ac *AirConditioner) SetTargetTemperature(temp float64) {
	ac.targetTemperature = temp
}

// IndoorTemperature returns the indoor temperature
func (ac *AirConditioner) IndoorTemperature() *float64 {
	return ac.indoorTemperature
}

// OutdoorTemperature returns the outdoor temperature
func (ac *AirConditioner) OutdoorTemperature() *float64 {
	return ac.outdoorTemperature
}

// SupportedOperationModes returns the supported operation modes
func (ac *AirConditioner) SupportedOperationModes() []OperationalMode {
	return ac.supportedOpModes
}

// OperationalMode returns the current operational mode
func (ac *AirConditioner) OperationalMode() OperationalMode {
	return ac.operationalMode
}

// SetOperationalMode sets the operational mode
func (ac *AirConditioner) SetOperationalMode(mode OperationalMode) {
	ac.operationalMode = mode
}

// SupportedFanSpeeds returns the supported fan speeds
func (ac *AirConditioner) SupportedFanSpeeds() []interface{} {
	return ac.supportedFanSpeeds
}

// supportsCustomFanSpeed returns whether custom fan speed is supported
func (ac *AirConditioner) supportsCustomFanSpeed() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityCustomFanSpeed)
}

// SupportsCustomFanSpeed returns whether custom fan speed is supported
func (ac *AirConditioner) SupportsCustomFanSpeed() bool {
	return ac.supportsCustomFanSpeed()
}

// FanSpeed returns the current fan speed
func (ac *AirConditioner) FanSpeed() interface{} {
	return ac.fanSpeed
}

// SetFanSpeed sets the fan speed
func (ac *AirConditioner) SetFanSpeed(speed interface{}) {
	// Convert float to int as needed
	if f, ok := speed.(float64); ok {
		speed = int(f)
	}
	ac.fanSpeed = speed
}

// supportsFanSpeed checks if a fan speed is supported
func (ac *AirConditioner) supportsFanSpeed(speed interface{}) bool {
	for _, s := range ac.supportedFanSpeeds {
		if s == speed {
			return true
		}
	}
	return false
}

// SupportsBreezeAway returns whether breeze away is supported
func (ac *AirConditioner) SupportsBreezeAway() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityBreezeAway) || caps.Has(CapabilityBreezeControl)
}

// BreezeAway returns whether breeze away is enabled
func (ac *AirConditioner) BreezeAway() *bool {
	result := ac.breezeMode == BreezeModeBreezeAway
	return &result
}

// SetBreezeAway sets whether breeze away is enabled
func (ac *AirConditioner) SetBreezeAway(enable bool) {
	if enable {
		ac.breezeMode = BreezeModeBreezeAway
	} else {
		ac.breezeMode = BreezeModeOff
	}

	caps := Capability(ac.capabilities.Flags())
	if caps.Has(CapabilityBreezeControl) {
		ac.updatedProperties[PropertyIdBreezeControl] = true
	} else {
		ac.updatedProperties[PropertyIdBreezeAway] = true
	}
}

// SupportsBreezeMild returns whether breeze mild is supported
func (ac *AirConditioner) SupportsBreezeMild() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityBreezeControl)
}

// BreezeMild returns whether breeze mild is enabled
func (ac *AirConditioner) BreezeMild() *bool {
	result := ac.breezeMode == BreezeModeBreezeMild
	return &result
}

// SetBreezeMild sets whether breeze mild is enabled
func (ac *AirConditioner) SetBreezeMild(enable bool) {
	if enable {
		ac.breezeMode = BreezeModeBreezeMild
	} else {
		ac.breezeMode = BreezeModeOff
	}
	ac.updatedProperties[PropertyIdBreezeControl] = true
}

// SupportsBreezeless returns whether breezeless is supported
func (ac *AirConditioner) SupportsBreezeless() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityBreezeless) || caps.Has(CapabilityBreezeControl)
}

// Breezeless returns whether breezeless is enabled
func (ac *AirConditioner) Breezeless() *bool {
	result := ac.breezeMode == BreezeModeBreezeless
	return &result
}

// SetBreezeless sets whether breezeless is enabled
func (ac *AirConditioner) SetBreezeless(enable bool) {
	if enable {
		ac.breezeMode = BreezeModeBreezeless
	} else {
		ac.breezeMode = BreezeModeOff
	}

	caps := Capability(ac.capabilities.Flags())
	if caps.Has(CapabilityBreezeControl) {
		ac.updatedProperties[PropertyIdBreezeControl] = true
	} else {
		ac.updatedProperties[PropertyIdBreezeless] = true
	}
}

// SupportedSwingModes returns the supported swing modes
func (ac *AirConditioner) SupportedSwingModes() []SwingMode {
	return ac.supportedSwingModes
}

// SwingMode returns the current swing mode
func (ac *AirConditioner) SwingMode() SwingMode {
	return ac.swingMode
}

// SetSwingMode sets the swing mode
func (ac *AirConditioner) SetSwingMode(mode SwingMode) {
	ac.swingMode = mode
}

// SupportsHorizontalSwingAngle returns whether horizontal swing angle is supported
func (ac *AirConditioner) SupportsHorizontalSwingAngle() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilitySwingHorizontalAngle)
}

// HorizontalSwingAngle returns the horizontal swing angle
func (ac *AirConditioner) HorizontalSwingAngle() SwingAngle {
	return ac.horizontalSwingAngle
}

// SetHorizontalSwingAngle sets the horizontal swing angle
func (ac *AirConditioner) SetHorizontalSwingAngle(angle SwingAngle) {
	ac.horizontalSwingAngle = angle
	ac.updatedProperties[PropertyIdSwingLrAngle] = true
}

// SupportsVerticalSwingAngle returns whether vertical swing angle is supported
func (ac *AirConditioner) SupportsVerticalSwingAngle() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilitySwingVerticalAngle)
}

// VerticalSwingAngle returns the vertical swing angle
func (ac *AirConditioner) VerticalSwingAngle() SwingAngle {
	return ac.verticalSwingAngle
}

// SetVerticalSwingAngle sets the vertical swing angle
func (ac *AirConditioner) SetVerticalSwingAngle(angle SwingAngle) {
	ac.verticalSwingAngle = angle
	ac.updatedProperties[PropertyIdSwingUdAngle] = true
}

// SupportsCascade returns whether cascade is supported
func (ac *AirConditioner) SupportsCascade() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityCascade)
}

// CascadeMode returns the cascade mode
func (ac *AirConditioner) CascadeMode() CascadeMode {
	return ac.cascadeMode
}

// SetCascadeMode sets the cascade mode
func (ac *AirConditioner) SetCascadeMode(mode CascadeMode) {
	ac.cascadeMode = mode
	ac.updatedProperties[PropertyIdCascade] = true
}

// SupportsEco returns whether eco mode is supported
func (ac *AirConditioner) SupportsEco() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityEco)
}

// Eco returns whether eco mode is enabled
func (ac *AirConditioner) Eco() bool {
	return ac.eco
}

// SetEco sets whether eco mode is enabled
func (ac *AirConditioner) SetEco(enabled bool) {
	ac.eco = enabled
}

// SupportsIECO returns whether IECO is supported
func (ac *AirConditioner) SupportsIECO() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityIECO)
}

// IECO returns whether IECO is enabled
func (ac *AirConditioner) IECO() bool {
	return ac.ieco
}

// SetIECO sets whether IECO is enabled
func (ac *AirConditioner) SetIECO(enabled bool) {
	ac.ieco = enabled
	ac.updatedProperties[PropertyIdIECO] = true
}

// SupportsFlashCool returns whether flash cool is supported
func (ac *AirConditioner) SupportsFlashCool() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityJetCool)
}

// FlashCool returns whether flash cool is enabled
func (ac *AirConditioner) FlashCool() bool {
	return ac.flashCool
}

// SetFlashCool sets whether flash cool is enabled
func (ac *AirConditioner) SetFlashCool(enabled bool) {
	ac.flashCool = enabled
	ac.updatedProperties[PropertyIdJetCool] = true
}

// SupportsTurbo returns whether turbo mode is supported
func (ac *AirConditioner) SupportsTurbo() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityTurbo)
}

// Turbo returns whether turbo mode is enabled
func (ac *AirConditioner) Turbo() bool {
	return ac.turbo
}

// SetTurbo sets whether turbo mode is enabled
func (ac *AirConditioner) SetTurbo(enabled bool) {
	ac.turbo = enabled
}

// SupportsFreezeProtection returns whether freeze protection is supported
func (ac *AirConditioner) SupportsFreezeProtection() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityFreezeProtection)
}

// FreezeProtection returns whether freeze protection is enabled
func (ac *AirConditioner) FreezeProtection() bool {
	return ac.freezeProtection
}

// SetFreezeProtection sets whether freeze protection is enabled
func (ac *AirConditioner) SetFreezeProtection(enabled bool) {
	ac.freezeProtection = enabled
}

// Sleep returns whether sleep mode is enabled
func (ac *AirConditioner) Sleep() bool {
	return ac.sleep
}

// SetSleep sets whether sleep mode is enabled
func (ac *AirConditioner) SetSleep(enabled bool) {
	ac.sleep = enabled
}

// FollowMe returns whether follow me is enabled
func (ac *AirConditioner) FollowMe() bool {
	return ac.followMe
}

// SetFollowMe sets whether follow me is enabled
func (ac *AirConditioner) SetFollowMe(enabled bool) {
	ac.followMe = enabled
}

// SupportsPurifier returns whether purifier is supported
func (ac *AirConditioner) SupportsPurifier() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityPurifier)
}

// Purifier returns whether purifier is enabled
func (ac *AirConditioner) Purifier() bool {
	return ac.purifier
}

// SetPurifier sets whether purifier is enabled
func (ac *AirConditioner) SetPurifier(enabled bool) {
	ac.purifier = enabled
}

// SupportsDisplayControl returns whether display control is supported
func (ac *AirConditioner) SupportsDisplayControl() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityDisplayControl)
}

// DisplayOn returns whether display is on
func (ac *AirConditioner) DisplayOn() *bool {
	return ac.displayOn
}

// SupportsFilterReminder returns whether filter reminder is supported
func (ac *AirConditioner) SupportsFilterReminder() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityFilterReminder)
}

// FilterAlert returns whether filter alert is active
func (ac *AirConditioner) FilterAlert() *bool {
	return ac.filterAlert
}

// EnableEnergyUsageRequests returns whether energy usage requests are enabled
func (ac *AirConditioner) EnableEnergyUsageRequests() bool {
	return ac.requestEnergyUsage
}

// SetEnableEnergyUsageRequests sets whether energy usage requests are enabled
func (ac *AirConditioner) SetEnableEnergyUsageRequests(enable bool) {
	ac.requestEnergyUsage = enable
}

// GetTotalEnergyUsage returns the total energy usage
func (ac *AirConditioner) GetTotalEnergyUsage(format EnergyDataFormat) *float64 {
	return ac.totalEnergyUsage[format]
}

// GetCurrentEnergyUsage returns the current energy usage
func (ac *AirConditioner) GetCurrentEnergyUsage(format EnergyDataFormat) *float64 {
	return ac.currentEnergyUsage[format]
}

// GetRealTimePowerUsage returns the real time power usage
func (ac *AirConditioner) GetRealTimePowerUsage(format EnergyDataFormat) *float64 {
	return ac.realTimePowerUsage[format]
}

// SupportsHumidity returns whether humidity is supported
func (ac *AirConditioner) SupportsHumidity() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityHumidity)
}

// IndoorHumidity returns the indoor humidity
func (ac *AirConditioner) IndoorHumidity() *int {
	return ac.indoorHumidity
}

// SupportsTargetHumidity returns whether target humidity is supported
func (ac *AirConditioner) SupportsTargetHumidity() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityTargetHumidity)
}

// TargetHumidity returns the target humidity
func (ac *AirConditioner) TargetHumidity() int {
	return ac.targetHumidity
}

// SetTargetHumidity sets the target humidity
func (ac *AirConditioner) SetTargetHumidity(humidity int) {
	ac.targetHumidity = humidity
}

// SupportsSelfClean returns whether self clean is supported
func (ac *AirConditioner) SupportsSelfClean() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilitySelfClean)
}

// SelfCleanActive returns whether self clean is active
func (ac *AirConditioner) SelfCleanActive() bool {
	if ac.selfCleanActive == nil {
		return false
	}
	return *ac.selfCleanActive
}

// SupportedRateSelects returns the supported rate selects
func (ac *AirConditioner) SupportedRateSelects() []RateSelect {
	return ac.supportedRateSelects
}

// RateSelect returns the current rate select
func (ac *AirConditioner) RateSelect() RateSelect {
	return ac.rateSelect
}

// SetRateSelect sets the rate select
func (ac *AirConditioner) SetRateSelect(rate RateSelect) {
	ac.rateSelect = rate
	ac.updatedProperties[PropertyIdRateSelect] = true
}

// SupportedAuxModes returns the supported aux modes
func (ac *AirConditioner) SupportedAuxModes() []AuxHeatMode {
	return ac.supportedAuxModes
}

// AuxMode returns the current aux mode
func (ac *AirConditioner) AuxMode() AuxHeatMode {
	return ac.auxMode
}

// SetAuxMode sets the aux mode
func (ac *AirConditioner) SetAuxMode(mode AuxHeatMode) {
	ac.auxMode = mode
}

// ErrorCode returns the error code
func (ac *AirConditioner) ErrorCode() *int {
	return ac.errorCode
}

// EnableGroup5DataRequests returns whether group 5 data requests are enabled
func (ac *AirConditioner) EnableGroup5DataRequests() bool {
	return ac.requestGroup5Data
}

// SetEnableGroup5DataRequests sets whether group 5 data requests are enabled
func (ac *AirConditioner) SetEnableGroup5DataRequests(enable bool) {
	ac.requestGroup5Data = enable
}

// DefrostActive returns whether defrost is active
func (ac *AirConditioner) DefrostActive() *bool {
	return ac.defrostActive
}

// OutdoorFanSpeed returns the outdoor fan speed
func (ac *AirConditioner) OutdoorFanSpeed() *int {
	return ac.outdoorFanSpeed
}

// SupportsOutSilent returns whether out silent is supported
func (ac *AirConditioner) SupportsOutSilent() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityOutSilent)
}

// OutSilent returns whether out silent is enabled
func (ac *AirConditioner) OutSilent() bool {
	return ac.outSilent
}

// SetOutSilent sets whether out silent is enabled
func (ac *AirConditioner) SetOutSilent(enabled bool) {
	ac.outSilent = enabled
	ac.updatedProperties[PropertyIdOutSilent] = true
}

// ToDict returns the device as a dictionary
// This is the Go equivalent of Python's to_dict method
func (ac *AirConditioner) ToDict() map[string]interface{} {
	result := ac.Device.ToDict()

	// Add AC-specific fields
	result["power"] = ac.powerState
	result["mode"] = ac.operationalMode
	result["fan_speed"] = ac.fanSpeed
	result["swing_mode"] = ac.swingMode
	result["horizontal_swing_angle"] = ac.horizontalSwingAngle
	result["vertical_swing_angle"] = ac.verticalSwingAngle
	result["cascade_mode"] = ac.cascadeMode
	result["target_temperature"] = ac.targetTemperature
	result["indoor_temperature"] = ac.indoorTemperature
	result["outdoor_temperature"] = ac.outdoorTemperature
	result["target_humidity"] = ac.targetHumidity
	result["indoor_humidity"] = ac.indoorHumidity
	result["eco"] = ac.eco
	result["turbo"] = ac.turbo
	result["freeze_protection"] = ac.freezeProtection
	result["sleep"] = ac.sleep
	result["display_on"] = ac.displayOn
	result["beep"] = ac.beepOn
	result["fahrenheit"] = ac.fahrenheitUnit
	result["filter_alert"] = ac.filterAlert
	result["follow_me"] = ac.followMe
	result["purifier"] = ac.purifier
	result["self_clean"] = ac.SelfCleanActive()
	result["total_energy_usage"] = ac.GetTotalEnergyUsage(EnergyDataFormatBCD)
	result["current_energy_usage"] = ac.GetCurrentEnergyUsage(EnergyDataFormatBCD)
	result["real_time_power_usage"] = ac.GetRealTimePowerUsage(EnergyDataFormatBCD)
	result["rate_select"] = ac.rateSelect
	result["aux_mode"] = ac.auxMode
	result["error_code"] = ac.errorCode
	result["defrost"] = ac.defrostActive
	result["out_silent"] = ac.outSilent

	return result
}

// CapabilitiesDict returns the device capabilities as a dictionary
// This is the Go equivalent of Python's capabilities_dict method
func (ac *AirConditioner) CapabilitiesDict() map[string]interface{} {
	caps := Capability(ac.capabilities.Flags())
	var flags []string
	if caps.Has(CapabilityCustomFanSpeed) {
		flags = append(flags, "CUSTOM_FAN_SPEED")
	}
	if caps.Has(CapabilityEco) {
		flags = append(flags, "ECO")
	}
	if caps.Has(CapabilityFreezeProtection) {
		flags = append(flags, "FREEZE_PROTECTION")
	}
	if caps.Has(CapabilityIECO) {
		flags = append(flags, "IECO")
	}
	if caps.Has(CapabilityTurbo) {
		flags = append(flags, "TURBO")
	}
	if caps.Has(CapabilityDisplayControl) {
		flags = append(flags, "DISPLAY_CONTROL")
	}
	if caps.Has(CapabilityEnergyStats) {
		flags = append(flags, "ENERGY_STATS")
	}
	if caps.Has(CapabilityFilterReminder) {
		flags = append(flags, "FILTER_REMINDER")
	}
	if caps.Has(CapabilityHumidity) {
		flags = append(flags, "HUMIDITY")
	}
	if caps.Has(CapabilityTargetHumidity) {
		flags = append(flags, "TARGET_HUMIDITY")
	}
	if caps.Has(CapabilitySwingHorizontalAngle) {
		flags = append(flags, "SWING_HORIZONTAL_ANGLE")
	}
	if caps.Has(CapabilitySwingVerticalAngle) {
		flags = append(flags, "SWING_VERTICAL_ANGLE")
	}
	if caps.Has(CapabilityBreezeAway) {
		flags = append(flags, "BREEZE_AWAY")
	}
	if caps.Has(CapabilityBreezeControl) {
		flags = append(flags, "BREEZE_CONTROL")
	}
	if caps.Has(CapabilityBreezeless) {
		flags = append(flags, "BREEZELESS")
	}
	if caps.Has(CapabilityCascade) {
		flags = append(flags, "CASCADE")
	}
	if caps.Has(CapabilityJetCool) {
		flags = append(flags, "JET_COOL")
	}
	if caps.Has(CapabilityOutSilent) {
		flags = append(flags, "OUT_SILENT")
	}
	if caps.Has(CapabilityPurifier) {
		flags = append(flags, "PURIFIER")
	}
	if caps.Has(CapabilitySelfClean) {
		flags = append(flags, "SELF_CLEAN")
	}

	return map[string]interface{}{
		"min_target_temperature":   ac.minTargetTemperature,
		"max_target_temperature":   ac.maxTargetTemperature,
		"supported_modes":          ac.supportedOpModes,
		"supported_swing_modes":    ac.supportedSwingModes,
		"supported_fan_speeds":     ac.supportedFanSpeeds,
		"supported_aux_modes":      ac.supportedAuxModes,
		"supported_rate_selects":   ac.supportedRateSelects,
		"additional_capabilities":  flags,
	}
}

// ============================================================================
// Deprecated methods (for backwards compatibility)
// ============================================================================

// SupportsEcoMode is deprecated. Use SupportsEco instead.
// Deprecated: Use SupportsEco instead.
func (ac *AirConditioner) SupportsEcoMode() bool {
	msmart.Deprecated("SupportsEcoMode", "SupportsEco", "")
	return ac.SupportsEco()
}

// EcoMode is deprecated. Use Eco instead.
// Deprecated: Use Eco instead.
func (ac *AirConditioner) EcoMode() bool {
	msmart.Deprecated("EcoMode", "Eco", "")
	return ac.Eco()
}

// SetEcoMode is deprecated. Use SetEco instead.
// Deprecated: Use SetEco instead.
func (ac *AirConditioner) SetEcoMode(enabled bool) {
	msmart.Deprecated("SetEcoMode", "SetEco", "")
	ac.SetEco(enabled)
}

// SupportsFreezeProtectionMode is deprecated. Use SupportsFreezeProtection instead.
// Deprecated: Use SupportsFreezeProtection instead.
func (ac *AirConditioner) SupportsFreezeProtectionMode() bool {
	msmart.Deprecated("SupportsFreezeProtectionMode", "SupportsFreezeProtection", "")
	return ac.SupportsFreezeProtection()
}

// FreezeProtectionMode is deprecated. Use FreezeProtection instead.
// Deprecated: Use FreezeProtection instead.
func (ac *AirConditioner) FreezeProtectionMode() bool {
	msmart.Deprecated("FreezeProtectionMode", "FreezeProtection", "")
	return ac.FreezeProtection()
}

// SetFreezeProtectionMode is deprecated. Use SetFreezeProtection instead.
// Deprecated: Use SetFreezeProtection instead.
func (ac *AirConditioner) SetFreezeProtectionMode(enabled bool) {
	msmart.Deprecated("SetFreezeProtectionMode", "SetFreezeProtection", "")
	ac.SetFreezeProtection(enabled)
}

// SleepMode is deprecated. Use Sleep instead.
// Deprecated: Use Sleep instead.
func (ac *AirConditioner) SleepMode() bool {
	msmart.Deprecated("SleepMode", "Sleep", "")
	return ac.Sleep()
}

// SetSleepMode is deprecated. Use SetSleep instead.
// Deprecated: Use SetSleep instead.
func (ac *AirConditioner) SetSleepMode(enabled bool) {
	msmart.Deprecated("SetSleepMode", "SetSleep", "")
	ac.SetSleep(enabled)
}

// SupportsTurboMode is deprecated. Use SupportsTurbo instead.
// Deprecated: Use SupportsTurbo instead.
func (ac *AirConditioner) SupportsTurboMode() bool {
	msmart.Deprecated("SupportsTurboMode", "SupportsTurbo", "")
	return ac.SupportsTurbo()
}

// TurboMode is deprecated. Use Turbo instead.
// Deprecated: Use Turbo instead.
func (ac *AirConditioner) TurboMode() bool {
	msmart.Deprecated("TurboMode", "Turbo", "")
	return ac.Turbo()
}

// SetTurboMode is deprecated. Use SetTurbo instead.
// Deprecated: Use SetTurbo instead.
func (ac *AirConditioner) SetTurboMode(enabled bool) {
	msmart.Deprecated("SetTurboMode", "SetTurbo", "")
	ac.SetTurbo(enabled)
}

// UseAlternateEnergyFormat is deprecated. Use format argument of Get*EnergyUsage methods instead.
// Deprecated: Use format argument of Get*EnergyUsage methods instead.
func (ac *AirConditioner) UseAlternateEnergyFormat() bool {
	msmart.Deprecated("", "", "Use format argument of Get*EnergyUsage methods instead.")
	return ac.useBinaryEnergy
}

// SetUseAlternateEnergyFormat is deprecated. Use format argument of Get*EnergyUsage methods instead.
// Deprecated: Use format argument of Get*EnergyUsage methods instead.
func (ac *AirConditioner) SetUseAlternateEnergyFormat(enable bool) {
	msmart.Deprecated("", "", "Use format argument of Get*EnergyUsage methods instead.")
	ac.useBinaryEnergy = enable
}

// TotalEnergyUsage is deprecated. Use GetTotalEnergyUsage instead.
// Deprecated: Use GetTotalEnergyUsage instead.
func (ac *AirConditioner) TotalEnergyUsage() *float64 {
	msmart.Deprecated("TotalEnergyUsage", "GetTotalEnergyUsage", "")
	format := EnergyDataFormatBCD
	if ac.useBinaryEnergy {
		format = EnergyDataFormatBinary
	}
	return ac.GetTotalEnergyUsage(format)
}

// CurrentEnergyUsage is deprecated. Use GetCurrentEnergyUsage instead.
// Deprecated: Use GetCurrentEnergyUsage instead.
func (ac *AirConditioner) CurrentEnergyUsage() *float64 {
	msmart.Deprecated("CurrentEnergyUsage", "GetCurrentEnergyUsage", "")
	format := EnergyDataFormatBCD
	if ac.useBinaryEnergy {
		format = EnergyDataFormatBinary
	}
	return ac.GetCurrentEnergyUsage(format)
}

// RealTimePowerUsage is deprecated. Use GetRealTimePowerUsage instead.
// Deprecated: Use GetRealTimePowerUsage instead.
func (ac *AirConditioner) RealTimePowerUsage() *float64 {
	msmart.Deprecated("RealTimePowerUsage", "GetRealTimePowerUsage", "")
	format := EnergyDataFormatBCD
	if ac.useBinaryEnergy {
		format = EnergyDataFormatBinary
	}
	return ac.GetRealTimePowerUsage(format)
}

// ============================================================================
// Helper functions
// ============================================================================

// ptrToVal returns the value of a pointer, or a default value if the pointer is nil
func ptrToVal[T any](ptr *T, def T) T {
	if ptr == nil {
		return def
	}
	return *ptr
}

// toInt converts an interface{} to int
func toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case int32:
		return int(val)
	case byte:
		return int(val)
	case float64:
		return int(val)
	default:
		return 0
	}
}

// containsOpMode checks if an operation mode is in a slice
func containsOpMode(modes []OperationalMode, mode OperationalMode) bool {
	for _, m := range modes {
		if m == mode {
			return true
		}
	}
	return false
}

// containsSwingMode checks if a swing mode is in a slice
func containsSwingMode(modes []SwingMode, mode SwingMode) bool {
	for _, m := range modes {
		if m == mode {
			return true
		}
	}
	return false
}

// containsRateSelect checks if a rate select is in a slice
func containsRateSelect(rates []RateSelect, rate RateSelect) bool {
	for _, r := range rates {
		if r == rate {
			return true
		}
	}
	return false
}

// containsAuxMode checks if an aux mode is in a slice
func containsAuxMode(modes []AuxHeatMode, mode AuxHeatMode) bool {
	for _, m := range modes {
		if m == mode {
			return true
		}
	}
	return false
}

// bytePtrToIntPtr converts a *byte to *int
func bytePtrToIntPtr(b *byte) *int {
	if b == nil {
		return nil
	}
	val := int(*b)
	return &val
}

// CommandInterface is an interface for all command types
type CommandInterface interface {
	ToBytes() []byte
}

// ResponseInterface is an interface for all response types
// Note: This is already defined in command.go, but we redeclare it here for clarity
// type ResponseInterface interface {
// 	ID() byte
// 	Payload() []byte
// 	String() string
// }
