// Package ac provides air conditioner device implementation.
// This is a translation from msmart-ng Python library.
// Original file: msmart/device/AC/device.py
package ac

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	msmart "github.com/RelicOfTesla/midea-msmart/msmart"
	"github.com/RelicOfTesla/midea-msmart/msmart/device"
	"github.com/RelicOfTesla/midea-msmart/msmart/device/xc"
)

// SubmitMode defines how a property should be submitted
type SubmitMode int

const (
	SubmitModeFullState      SubmitMode = iota // Requires full state submission (SetStateCommand)
	SubmitModeSingleProperty                   // Supports single property submission (SetPropertiesCommand)
	SubmitModeRefreshFirst                     // Needs to refresh current state before submission
)

// propertyConfig defines the submission mode for each property
type propertyConfig struct {
	mode       SubmitMode
	propertyId PropertyId // Only valid for SubmitModeSingleProperty
}

// propertySubmitConfig is the global configuration map for property submission modes
// This determines how each property should be handled during Apply()
var propertySubmitConfig = map[string]propertyConfig{
	// Full state submission - main AC controls
	"power_state":          {mode: SubmitModeRefreshFirst}, // Need current state to avoid accidental power off
	"target_temperature":   {mode: SubmitModeFullState},
	"operational_mode":     {mode: SubmitModeFullState},
	"fan_speed":            {mode: SubmitModeFullState},
	"swing_mode":           {mode: SubmitModeFullState},
	"eco":                  {mode: SubmitModeFullState},
	"turbo":                {mode: SubmitModeFullState},
	"freeze_protection":    {mode: SubmitModeFullState},
	"sleep":                {mode: SubmitModeFullState},
	"fahrenheit":           {mode: SubmitModeFullState},
	"follow_me":            {mode: SubmitModeFullState},
	"purifier":             {mode: SubmitModeFullState},
	"target_humidity":      {mode: SubmitModeFullState},
	"aux_heat":             {mode: SubmitModeFullState},
	"independent_aux_heat": {mode: SubmitModeFullState},
	"beep_on":              {mode: SubmitModeFullState},

	// Single property submission - specific features
	"swing_ud_angle": {mode: SubmitModeSingleProperty, propertyId: PropertyIdSwingUdAngle},
	"swing_lr_angle": {mode: SubmitModeSingleProperty, propertyId: PropertyIdSwingLrAngle},
	"buzzer":         {mode: SubmitModeSingleProperty, propertyId: PropertyIdBuzzer},
	"breezeless":     {mode: SubmitModeSingleProperty, propertyId: PropertyIdBreezeless},
	"breeze_away":    {mode: SubmitModeSingleProperty, propertyId: PropertyIdBreezeAway},
	"rate_select":    {mode: SubmitModeSingleProperty, propertyId: PropertyIdRateSelect},
	"fresh_air":      {mode: SubmitModeSingleProperty, propertyId: PropertyIdFreshAir},
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
	return msmart.EnumFromInt(value, SwingAngle(0).Values(), SwingAngleDefault)
}

// GetFromName returns the SwingAngle for a given name, or the default if not found
func (SwingAngle) GetFromName(name string) SwingAngle {
	switch name {
	case "OFF":
		return SwingAngleOff
	case "POS_1":
		return SwingAnglePos1
	case "POS_2":
		return SwingAnglePos2
	case "POS_3":
		return SwingAnglePos3
	case "POS_4":
		return SwingAnglePos4
	case "POS_5":
		return SwingAnglePos5
	default:
		return SwingAngleDefault
	}
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

// GetFromName returns the CascadeMode for a given name, or the default if not found
func (CascadeMode) GetFromName(name string) CascadeMode {
	switch name {
	case "OFF":
		return CascadeModeOff
	case "UP":
		return CascadeModeUp
	case "DOWN":
		return CascadeModeDown
	default:
		return CascadeModeDefault
	}
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

// GetFromName returns the RateSelect for a given name, or the default if not found
func (RateSelect) GetFromName(name string) RateSelect {
	switch name {
	case "OFF":
		return RateSelectOff
	case "GEAR_50":
		return RateSelectGear50
	case "GEAR_75":
		return RateSelectGear75
	case "LEVEL_1":
		return RateSelectLevel1
	case "LEVEL_2":
		return RateSelectLevel2
	case "LEVEL_3":
		return RateSelectLevel3
	case "LEVEL_4":
		return RateSelectLevel4
	case "LEVEL_5":
		return RateSelectLevel5
	default:
		return RateSelectDefault
	}
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

// GetFromName returns the BreezeMode for a given name, or the default if not found
func (BreezeMode) GetFromName(name string) BreezeMode {
	switch name {
	case "OFF":
		return BreezeModeOff
	case "BREEZE_AWAY":
		return BreezeModeBreezeAway
	case "BREEZE_MILD":
		return BreezeModeBreezeMild
	case "BREEZELESS":
		return BreezeModeBreezeless
	default:
		return BreezeModeDefault
	}
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

// GetFromName returns the AuxHeatMode for a given name, or the default if not found
func (AuxHeatMode) GetFromName(name string) AuxHeatMode {
	switch name {
	case "OFF":
		return AuxHeatModeOff
	case "AUX_HEAT":
		return AuxHeatModeAuxHeat
	case "AUX_ONLY":
		return AuxHeatModeAuxOnly
	default:
		return AuxHeatModeDefault
	}
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

// GetFromName returns the Capability flag for a given name, or 0 if not found
func (Capability) GetFromName(name string) Capability {
	switch name {
	case "CUSTOM_FAN_SPEED":
		return CapabilityCustomFanSpeed
	case "ECO":
		return CapabilityEco
	case "FREEZE_PROTECTION":
		return CapabilityFreezeProtection
	case "IECO":
		return CapabilityIECO
	case "TURBO":
		return CapabilityTurbo
	case "DISPLAY_CONTROL":
		return CapabilityDisplayControl
	case "ENERGY_STATS":
		return CapabilityEnergyStats
	case "FILTER_REMINDER":
		return CapabilityFilterReminder
	case "HUMIDITY":
		return CapabilityHumidity
	case "TARGET_HUMIDITY":
		return CapabilityTargetHumidity
	case "SWING_HORIZONTAL_ANGLE":
		return CapabilitySwingHorizontalAngle
	case "SWING_VERTICAL_ANGLE":
		return CapabilitySwingVerticalAngle
	case "BREEZE_AWAY":
		return CapabilityBreezeAway
	case "BREEZE_CONTROL":
		return CapabilityBreezeControl
	case "BREEZELESS":
		return CapabilityBreezeless
	case "CASCADE":
		return CapabilityCascade
	case "JET_COOL":
		return CapabilityJetCool
	case "OUT_SILENT":
		return CapabilityOutSilent
	case "PURIFIER":
		return CapabilityPurifier
	case "SELF_CLEAN":
		return CapabilitySelfClean
	default:
		return Capability(0)
	}
}

// AirConditioner represents an air conditioner device
// This is the main struct that translates Python's AirConditioner class
type AirConditioner struct {
	Device *msmart.DeviceBase

	// State management - unified KV maps
	// lastKvState stores current device state from Refresh()
	// pendingKvState stores user-set values that haven't been applied yet
	// lastKvProp stores current device properties from Refresh()
	// pendingKvProp stores property values that need to be updated
	lastKvState    map[StateId]any
	pendingKvState map[StateId]any
	lastKvProp     map[PropertyId]any
	pendingKvProp  map[PropertyId]any

	// Capabilities
	minTargetTemperature float64
	maxTargetTemperature float64

	capabilities *msmart.CapabilityManager

	supportedOpModes     []xc.OperationalMode
	supportedSwingModes  []xc.SwingMode
	supportedFanSpeeds   []interface{} // FanSpeed or int
	supportedRateSelects []RateSelect
	supportedAuxModes    []AuxHeatMode

	// Misc
	requestEnergyUsage bool
	requestGroup5Data  bool

	// Energy usage data
	totalEnergyUsage   map[EnergyDataFormat]*float64
	currentEnergyUsage map[EnergyDataFormat]*float64
	realTimePowerUsage map[EnergyDataFormat]*float64
	useBinaryEnergy    bool

	// Supported properties
	supportedProperties map[PropertyId]bool
	needsRefresh        bool // Flag to indicate if Refresh() is needed before Apply()
}

var _ device.Device = (*AirConditioner)(nil)
var _ device.DeviceAuthV3 = (*AirConditioner)(nil)
var _ AC = (*AirConditioner)(nil)

// NewAirConditioner creates a new AirConditioner instance
// This is the Go equivalent of Python's __init__ method
func NewAirConditioner(ip string, port int, deviceID string, opts ...msmart.DeviceOption) *AirConditioner {
	ac := &AirConditioner{
		Device: msmart.NewBaseDevice(ip, port, deviceID, msmart.DeviceTypeAirConditioner, opts...),

		// State management
		lastKvState:    make(map[StateId]any),
		pendingKvState: make(map[StateId]any),
		lastKvProp:     make(map[PropertyId]any),
		pendingKvProp:  make(map[PropertyId]any),

		// Default values
		minTargetTemperature: 17.0,
		maxTargetTemperature: 30.0,

		capabilities: msmart.NewCapabilityManager(0),

		supportedOpModes: []xc.OperationalMode{
			xc.OperationalModeAuto,
			xc.OperationalModeCool,
			xc.OperationalModeDry,
			xc.OperationalModeHeat,
			xc.OperationalModeFanOnly,
		},
		supportedSwingModes: []xc.SwingMode{
			xc.SwingModeOff,
		},
		supportedFanSpeeds: []interface{}{
			xc.FanSpeedAuto,
			xc.FanSpeedLow,
			xc.FanSpeedMedium,
			xc.FanSpeedHigh,
		},
		supportedRateSelects: []RateSelect{},
		supportedAuxModes:    []AuxHeatMode{},

		supportedProperties: make(map[PropertyId]bool),
		needsRefresh:        true,
	}

	// Initialize default lastState values
	ac.lastKvState[StateIdBeepOn] = false
	ac.lastKvState[StateIdTargetTemperature] = 17.0
	ac.lastKvState[StateIdTargetHumidity] = 40
	ac.lastKvState[StateIdOperationalMode] = xc.OperationalModeAuto
	ac.lastKvState[StateIdFanSpeed] = xc.FanSpeedAuto
	ac.lastKvState[StateIdSwingMode] = xc.SwingModeOff
	ac.lastKvState[StateIdEco] = false
	ac.lastKvState[StateIdTurbo] = false
	ac.lastKvState[StateIdFreezeProtection] = false
	ac.lastKvState[StateIdSleep] = false
	ac.lastKvState[StateIdFahrenheitUnit] = false
	ac.lastKvState[StateIdFollowMe] = false
	ac.lastKvState[StateIdPurifier] = false

	// Initialize property defaults in lastKvProp
	ac.lastKvProp[PropertyIdIECO] = false
	ac.lastKvProp[PropertyIdJetCool] = false
	ac.lastKvProp[PropertyIdOutSilent] = false
	ac.lastKvProp[PropertyIdSwingLrAngle] = byte(0)  // SwingAngleOff
	ac.lastKvProp[PropertyIdSwingUdAngle] = byte(0)  // SwingAngleOff
	ac.lastKvProp[PropertyIdCascade] = byte(0)       // CascadeModeOff
	ac.lastKvProp[PropertyIdRateSelect] = byte(100)  // RateSelectOff
	ac.lastKvProp[PropertyIdBreezeControl] = byte(1) // BreezeModeOff

	// Initialize energy usage maps
	ac.totalEnergyUsage = make(map[EnergyDataFormat]*float64)
	ac.currentEnergyUsage = make(map[EnergyDataFormat]*float64)
	ac.realTimePowerUsage = make(map[EnergyDataFormat]*float64)

	return ac
}

// getState retrieves a value from the state map
// If pendingState has the key, it returns that value instead (user-set value takes priority)
// _getState retrieves a value from the state map (internal use, no access control)
// For public API, use GetState which enforces read access control
func (ac *AirConditioner) _getState(key StateId, defaultValue interface{}) interface{} {
	// First check pendingState (user-set value takes priority)
	if val, ok := ac.pendingKvState[key]; ok {
		return val
	}
	// Then check lastState
	if val, ok := ac.lastKvState[key]; ok {
		return val
	}
	return defaultValue
}

// _setState sets a value in the pendingState map (internal use, no access control)
// For public API, use SetState which enforces write access control
func (ac *AirConditioner) _setState(key StateId, value interface{}) {
	ac.pendingKvState[key] = value
}

// _getProp retrieves a value from the property map (internal use)
// For public API, use GetProperty
func (ac *AirConditioner) _getProp(key PropertyId, defaultValue any) any {
	// First check pendingKvProp (user-set value takes priority)
	if val, ok := ac.pendingKvProp[key]; ok {
		return val
	}
	// Then check lastKvProp
	if val, ok := ac.lastKvProp[key]; ok {
		return val
	}
	return defaultValue
}

// _setProp sets a value in the pendingKvProp map (internal use)
// For public API, use SetProperty
func (ac *AirConditioner) _setProp(key PropertyId, value any) {
	ac.pendingKvProp[key] = value
}

// GetState retrieves a value from the state map (public API with read access control)
// Returns ErrStateNotReadable if the StateId is WriteOnly
func (ac *AirConditioner) GetState(key StateId) (interface{}, error) {
	if !key.IsReadable() {
		return nil, fmt.Errorf("state %d is not readable (WriteOnly)", key)
	}
	return ac._getState(key, nil), nil
}

// SetState sets a value in the pending state map (public API with write access control)
// Returns ErrStateNotWritable if the StateId is ReadOnly
func (ac *AirConditioner) SetState(key StateId, value interface{}) error {
	if !key.IsWritable() {
		return fmt.Errorf("state %d is not writable (ReadOnly)", key)
	}
	ac._setState(key, value)
	return nil
}

// GetProperty retrieves a value from the property map (public API)
func (ac *AirConditioner) GetProperty(key PropertyId) (any, error) {
	return ac._getProp(key, nil), nil
}

// SetProperty sets a value in the pending property map (public API)
func (ac *AirConditioner) SetProperty(key PropertyId, value any) error {
	ac._setProp(key, value)
	return nil
}

// commitState moves pendingState values to lastState and clears pendingState
// This is called after Apply() succeeds
func (ac *AirConditioner) commitState() {
	for k, v := range ac.pendingKvState {
		ac.lastKvState[k] = v
	}
	ac.pendingKvState = make(map[StateId]any)
	ac.pendingKvProp = make(map[PropertyId]any)
}

// updateStateFromResponse updates lastState map from a device response
func (ac *AirConditioner) updateStateFromResponse(key StateId, value interface{}) {
	ac.lastKvState[key] = value
}

func (ac *AirConditioner) updatePropFromResponse(key PropertyId, value any) {
	ac.lastKvProp[key] = value
}

// Get methods - read from state map

func (ac *AirConditioner) BeepOn() bool {
	return ac._getState(StateIdBeepOn, false).(bool)
}

func (ac *AirConditioner) PowerOn() *bool {
	if val := ac._getState(StateIdPowerOn, nil); val != nil {
		if b, ok := val.(bool); ok {
			return &b
		}
	}
	return nil
}

func (ac *AirConditioner) PowerState() bool {
	if val := ac.PowerOn(); val != nil {
		return *val
	}
	return false
}

func (ac *AirConditioner) TargetTemperature() float64 {
	return ac._getState(StateIdTargetTemperature, 17.0).(float64)
}

func (ac *AirConditioner) TargetHumidity() int {
	return ac._getState(StateIdTargetHumidity, 40).(int)
}

func (ac *AirConditioner) OperationalMode() xc.OperationalMode {
	return ac._getState(StateIdOperationalMode, xc.OperationalModeAuto).(xc.OperationalMode)
}

func (ac *AirConditioner) FanSpeed() xc.FanSpeed {
	return ac._getState(StateIdFanSpeed, xc.FanSpeedAuto).(xc.FanSpeed)
}

func (ac *AirConditioner) SwingMode() xc.SwingMode {
	return ac._getState(StateIdSwingMode, xc.SwingModeOff).(xc.SwingMode)
}

func (ac *AirConditioner) Eco() bool {
	return ac._getState(StateIdEco, false).(bool)
}

func (ac *AirConditioner) Turbo() bool {
	return ac._getState(StateIdTurbo, false).(bool)
}

func (ac *AirConditioner) FreezeProtection() bool {
	return ac._getState(StateIdFreezeProtection, false).(bool)
}

func (ac *AirConditioner) Sleep() bool {
	return ac._getState(StateIdSleep, false).(bool)
}

func (ac *AirConditioner) FahrenheitUnit() bool {
	return ac._getState(StateIdFahrenheitUnit, false).(bool)
}

func (ac *AirConditioner) DisplayOn() *bool {
	if val := ac._getState(StateIdDisplayOn, nil); val != nil {
		if b, ok := val.(bool); ok {
			return &b
		}
	}
	return nil
}

func (ac *AirConditioner) FollowMe() bool {
	return ac._getState(StateIdFollowMe, false).(bool)
}

func (ac *AirConditioner) Purifier() bool {
	return ac._getState(StateIdPurifier, false).(bool)
}

func (ac *AirConditioner) Ieco() bool {
	if val := ac._getProp(PropertyIdIECO, false); val != nil {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func (ac *AirConditioner) FlashCool() bool {
	if val := ac._getProp(PropertyIdJetCool, false); val != nil {
		if b, ok := val.(bool); ok {
			return b
		}
		if b, ok := val.(byte); ok {
			return b != 0
		}
	}
	return false
}

func (ac *AirConditioner) OutSilent() bool {
	if val := ac._getProp(PropertyIdOutSilent, false); val != nil {
		if b, ok := val.(bool); ok {
			return b
		}
		if b, ok := val.(byte); ok {
			return b != 0
		}
	}
	return false
}

func (ac *AirConditioner) HorizontalSwingAngle() SwingAngle {
	if val := ac._getProp(PropertyIdSwingLrAngle, byte(0)); val != nil {
		if b, ok := val.(byte); ok {
			return SwingAngle(b)
		}
	}
	return SwingAngleOff
}

func (ac *AirConditioner) VerticalSwingAngle() SwingAngle {
	if val := ac._getProp(PropertyIdSwingUdAngle, byte(0)); val != nil {
		if b, ok := val.(byte); ok {
			return SwingAngle(b)
		}
	}
	return SwingAngleOff
}

func (ac *AirConditioner) CascadeMode() CascadeMode {
	if val := ac._getProp(PropertyIdCascade, byte(0)); val != nil {
		if b, ok := val.(byte); ok {
			return CascadeMode(b)
		}
	}
	return CascadeModeOff
}

func (ac *AirConditioner) RateSelect() RateSelect {
	if val := ac._getProp(PropertyIdRateSelect, byte(100)); val != nil {
		if b, ok := val.(byte); ok {
			return RateSelect(b)
		}
	}
	return RateSelectOff
}

func (ac *AirConditioner) BreezeMode() BreezeMode {
	if val := ac._getProp(PropertyIdBreezeControl, byte(1)); val != nil {
		if b, ok := val.(byte); ok {
			return BreezeMode(b)
		}
	}
	return BreezeModeOff
}

func (ac *AirConditioner) AuxMode() AuxHeatMode {
	return ac._getState(StateIdAuxMode, AuxHeatModeOff).(AuxHeatMode)
}

func (ac *AirConditioner) IndoorTemperature() *float64 {
	if val := ac._getState(StateIdIndoorTemperature, nil); val != nil {
		if f, ok := val.(float64); ok {
			return &f
		}
	}
	return nil
}

func (ac *AirConditioner) IndoorHumidity() *int {
	if val := ac._getState(StateIdIndoorHumidity, nil); val != nil {
		if i, ok := val.(int); ok {
			return &i
		}
	}
	return nil
}

func (ac *AirConditioner) OutdoorTemperature() *float64 {
	if val := ac._getState(StateIdOutdoorTemperature, nil); val != nil {
		if f, ok := val.(float64); ok {
			return &f
		}
	}
	return nil
}

func (ac *AirConditioner) ErrorCode() *int {
	if val := ac._getState(StateIdErrorCode, nil); val != nil {
		if i, ok := val.(int); ok {
			return &i
		}
	}
	return nil
}

// Set methods - write to pendingState map

func (ac *AirConditioner) SetBeepOn(beep bool) {
	ac._setState(StateIdBeepOn, beep)
}

func (ac *AirConditioner) SetPowerOn(power bool) {
	ac._setState(StateIdPowerOn, power)
}

func (ac *AirConditioner) SetPowerState(power bool) {
	ac.SetPowerOn(power)
}

func (ac *AirConditioner) SetTargetTemperature(temp float64) {
	ac._setState(StateIdTargetTemperature, temp)
}

func (ac *AirConditioner) SetOperationalMode(mode xc.OperationalMode) {
	ac._setState(StateIdOperationalMode, mode)
}

func (ac *AirConditioner) SetFanSpeed(speed xc.FanSpeed) {
	ac._setState(StateIdFanSpeed, speed)
}

func (ac *AirConditioner) SetSwingMode(mode xc.SwingMode) {
	ac._setState(StateIdSwingMode, mode)
}

func (ac *AirConditioner) SetEco(eco bool) {
	ac._setState(StateIdEco, eco)
}

func (ac *AirConditioner) SetTurbo(turbo bool) {
	ac._setState(StateIdTurbo, turbo)
}

func (ac *AirConditioner) SetFreezeProtection(fp bool) {
	ac._setState(StateIdFreezeProtection, fp)
}

func (ac *AirConditioner) SetSleep(sleep bool) {
	ac._setState(StateIdSleep, sleep)
}

func (ac *AirConditioner) SetFahrenheitUnit(f bool) {
	ac._setState(StateIdFahrenheitUnit, f)
}

func (ac *AirConditioner) SetHorizontalSwingAngle(angle SwingAngle) {
	ac._setProp(PropertyIdSwingLrAngle, byte(angle))
}

func (ac *AirConditioner) SetVerticalSwingAngle(angle SwingAngle) {
	ac._setProp(PropertyIdSwingUdAngle, byte(angle))
}

func (ac *AirConditioner) SetCascadeMode(mode CascadeMode) {
	ac._setProp(PropertyIdCascade, byte(mode))
}

func (ac *AirConditioner) SetFlashCool(fc bool) {
	ac._setProp(PropertyIdJetCool, fc)
}

// updateState updates the local state from a device state response
// This is the Go equivalent of Python's _update_state method
func (ac *AirConditioner) updateState(res ResponseInterface) {
	switch r := res.(type) {
	case *StateResponse:
		slog.Debug("State response payload from device", "id", ac.GetID(), "response", fmt.Sprintf("%+v", r))

		// Convert to KV format
		kv := r.ToKv(nil)

		// Process special conversions
		for key, value := range kv.Values {
			switch key {
			case StateIdOperationalMode:
				if b, ok := value.(byte); ok {
					value = xc.OperationalMode(0).GetFromValue(int(b))
				}
			case StateIdFanSpeed:
				if b, ok := value.(byte); ok {
					if ac.supportsCustomFanSpeed() {
						value = xc.FanSpeed(b)
					} else {
						value = xc.FanSpeed(0).GetFromValue(int(b))
					}
				}
			case StateIdSwingMode:
				if b, ok := value.(byte); ok {
					value = xc.SwingMode(0).GetFromValue(int(b))
				}
			case StateIdEco, StateIdTurbo, StateIdFreezeProtection, StateIdSleep, StateIdFollowMe, StateIdPurifier:
				// Use current value as default if nil
				if value == nil {
					switch key {
					case StateIdEco:
						value = ac.Eco()
					case StateIdTurbo:
						value = ac.Turbo()
					case StateIdFreezeProtection:
						value = ac.FreezeProtection()
					case StateIdSleep:
						value = ac.Sleep()
					case StateIdFollowMe:
						value = ac.FollowMe()
					case StateIdPurifier:
						value = ac.Purifier()
					}
				}
			}
			ac.updateStateFromResponse(key, value)
		}

	case *PropertiesResponse:
		slog.Debug("Properties response payload from device", "id", ac.GetID(), "response", r)

		// Store all properties directly to lastKvProp
		for propId, value := range r.properties {
			ac.updatePropFromResponse(propId, value)
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

		if r.Humidity != nil {
			ac.updateStateFromResponse(StateIdIndoorHumidity, int(*r.Humidity))
		}
		if r.OutdoorFanSpeed != nil {
			ac.updateStateFromResponse(StateIdOutdoorFanSpeed, int(*r.OutdoorFanSpeed))
		}
		if r.Defrost != nil {
			ac.updateStateFromResponse(StateIdDefrostActive, *r.Defrost)
		}

	default:
		slog.Debug("Ignored unknown response from device", "id", ac.GetID(), "response", res)
	}
}

// updateCapabilities updates device capabilities from a CapabilitiesResponse
// This is the Go equivalent of Python's _update_capabilities method
func (ac *AirConditioner) updateCapabilities(res *CapabilitiesResponse) {
	// Build list of supported operation modes
	opModes := []xc.OperationalMode{xc.OperationalModeFanOnly}
	if res.DryMode() {
		opModes = append(opModes, xc.OperationalModeDry)
	}
	if res.CoolMode() {
		opModes = append(opModes, xc.OperationalModeCool)
	}
	if res.HeatMode() {
		opModes = append(opModes, xc.OperationalModeHeat)
	}
	if res.AutoMode() {
		opModes = append(opModes, xc.OperationalModeAuto)
	}
	if res.TargetHumidity() {
		// Add SMART_DRY support if target humidity is supported
		opModes = append(opModes, xc.OperationalModeSmartDry)
	}
	ac.supportedOpModes = opModes

	// Build list of supported swing modes
	swingModes := []xc.SwingMode{xc.SwingModeOff}
	if res.SwingHorizontal() {
		swingModes = append(swingModes, xc.SwingModeHorizontal)
	}
	if res.SwingVertical() {
		swingModes = append(swingModes, xc.SwingModeVertical)
	}
	if res.SwingBoth() {
		swingModes = append(swingModes, xc.SwingModeBoth)
	}
	ac.supportedSwingModes = swingModes

	// Build list of supported fan speeds
	fanSpeeds := make([]interface{}, 0)
	if res.FanSilent() {
		fanSpeeds = append(fanSpeeds, xc.FanSpeedSilent)
	}
	if res.FanLow() {
		fanSpeeds = append(fanSpeeds, xc.FanSpeedLow)
	}
	if res.FanMedium() {
		fanSpeeds = append(fanSpeeds, xc.FanSpeedMedium)
	}
	if res.FanHigh() {
		fanSpeeds = append(fanSpeeds, xc.FanSpeedHigh)
	}
	if res.FanAuto() {
		fanSpeeds = append(fanSpeeds, xc.FanSpeedAuto)
	}
	if res.FanCustom() {
		// Include additional MAX speed if custom speeds are supported
		fanSpeeds = append(fanSpeeds, xc.FanSpeedMax)
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
	var allResponses []ResponseInterface

	for _, cmd := range commands {
		// Get command bytes
		data := cmd.ToBytes()

		// Send command
		responses, err := ac.SendBytes(ctx, data)
		if err != nil {
			return nil, err
		}

		// Device is online if any response received
		ac.Device.SetOnline(len(responses) > 0)

		// Parse responses
		for _, respData := range responses {
			response, err := ConstructResponse(respData)
			if err != nil {
				slog.Debug("Failed to construct response", "error", err, "data", respData)
				continue
			}
			allResponses = append(allResponses, response)
		}
	}

	// Device is supported if online and any supported response is received
	ac.Device.SetSupported(ac.GetOnline() && len(allResponses) > 0)

	return allResponses, nil
}

// sendCommandGetResponseWithClass sends a command and returns the first response of the requested class
func (ac *AirConditioner) sendCommandGetResponseWithClass(ctx context.Context, command CommandInterface, responseClass ResponseId) (ResponseInterface, error) {
	responses, err := ac.sendCommandsGetResponses(ctx, []CommandInterface{command})
	if err != nil {
		return nil, err
	}

	for _, response := range responses {
		if response.ID() == byte(responseClass) {
			return response, nil
		}
		slog.Debug("Ignored response of unexpected type", "id", ac.GetID(), "response", response)
	}

	return nil, nil
}

// GetCapabilities fetches the device capabilities
// This is the Go equivalent of Python's get_capabilities method
func (ac *AirConditioner) GetCapabilities(ctx context.Context) error {
	// Send capabilities request and get a response
	cmd := NewGetCapabilitiesCommand(false)
	response, err := ac.sendCommandGetResponseWithClass(ctx, cmd, ResponseIdCapabilities)
	if err != nil {
		return fmt.Errorf("failed to query capabilities: %w", err)
	}
	if response == nil {
		return fmt.Errorf("failed to query capabilities from device %s", ac.GetID())
	}

	capsResponse, ok := response.(*CapabilitiesResponse)
	if !ok {
		return fmt.Errorf("unexpected response type for capabilities")
	}

	slog.Debug("Capabilities response payload from device", "id", ac.GetID(), "response", capsResponse)
	slog.Debug("Raw capabilities", "id", ac.GetID(), "capabilities", capsResponse.RawCapabilities())

	// Send 2nd capabilities request if needed
	if capsResponse.AdditionalCapabilities() {
		cmd := NewGetCapabilitiesCommand(true)
		additionalResponse, err := ac.sendCommandGetResponseWithClass(ctx, cmd, ResponseIdCapabilities)
		if err != nil {
			slog.Warn("Failed to query additional capabilities from device", "id", ac.GetID(), "error", err)
		} else if additionalResponse != nil {
			addCapsResponse, ok := additionalResponse.(*CapabilitiesResponse)
			if ok {
				slog.Debug("Additional capabilities response payload from device", "id", ac.GetID(), "response", addCapsResponse)
				// Merge additional capabilities
				capsResponse.Merge(addCapsResponse)
				slog.Debug("Merged raw capabilities", "id", ac.GetID(), "capabilities", capsResponse.RawCapabilities())
			}
		} else {
			slog.Warn("Failed to query additional capabilities from device", "id", ac.GetID())
		}
	}

	// Update device capabilities
	ac.updateCapabilities(capsResponse)

	return nil
}

// ToggleDisplay toggles the device display if the device supports it
// This is the Go equivalent of Python's toggle_display method
func (ac *AirConditioner) ToggleDisplay(ctx context.Context) error {
	if !ac.SupportsDisplayControl() {
		slog.Warn("Device is not capable of display control", "id", ac.GetID())
	}

	// Send the command and ignore all responses
	cmd := NewToggleDisplayCommand()
	cmd.BeepOn = ac.BeepOn()
	_, err := ac.sendCommandsGetResponses(ctx, []CommandInterface{cmd})
	if err != nil {
		return err
	}

	// Force a refresh to get the updated display state
	return ac.Refresh(ctx)
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
	var commands []CommandInterface

	// Always request state updates
	commands = append(commands, NewGetStateCommand())

	// Fetch power stats if supported
	if ac.requestEnergyUsage {
		commands = append(commands, NewGetEnergyUsageCommand())
	}

	// Request Group 5 data if humidity is supported or otherwise enabled
	if ac.SupportsHumidity() || ac.requestGroup5Data {
		commands = append(commands, NewGetGroup5Command())
	}

	// Update supported properties
	if len(ac.supportedProperties) > 0 {
		var props []PropertyId
		for prop := range ac.supportedProperties {
			props = append(props, prop)
		}
		commands = append(commands, NewGetPropertiesCommand(props))
	}

	// Send all commands and collect responses
	responses, err := ac.sendCommandsGetResponses(ctx, commands)
	if err != nil {
		return err
	}

	// Update state from responses
	for _, response := range responses {
		ac.updateState(response)
	}

	return nil
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
	properties[PropertyIdBuzzer] = ac.BeepOn()

	// Build command with properties
	cmd := NewSetPropertiesCommand(properties)

	// Send command and update state
	responses, err := ac.sendCommandsGetResponses(ctx, []CommandInterface{cmd})
	if err != nil {
		return err
	}

	for _, response := range responses {
		ac.updateState(response)
	}

	return nil
}

// Apply applies the local state to the device
// This is the Go equivalent of Python's apply method
func (ac *AirConditioner) Apply(ctx context.Context) error {
	// Warn if trying to apply unsupported modes
	if !msmart.Contains(ac.supportedOpModes, ac.OperationalMode()) {
		slog.Warn("Device is not capable of operational mode", "id", ac.GetID(), "mode", ac.OperationalMode())
	}

	if !ac.supportsFanSpeed(ac.FanSpeed()) && !ac.supportsCustomFanSpeed() {
		slog.Warn("Device is not capable of fan speed", "id", ac.GetID(), "speed", ac.FanSpeed())
	}

	if !msmart.Contains(ac.supportedSwingModes, ac.SwingMode()) {
		slog.Warn("Device is not capable of swing mode", "id", ac.GetID(), "mode", ac.SwingMode())
	}

	if ac.Turbo() && !ac.SupportsTurbo() {
		slog.Warn("Device is not capable of turbo mode", "id", ac.GetID())
	}

	if ac.Eco() && !ac.SupportsEco() {
		slog.Warn("Device is not capable of eco mode", "id", ac.GetID())
	}

	if ac.FreezeProtection() && !ac.SupportsFreezeProtection() {
		slog.Warn("Device is not capable of freeze protection", "id", ac.GetID())
	}

	if ac.RateSelect() != RateSelectOff && !msmart.Contains(ac.supportedRateSelects, ac.RateSelect()) {
		slog.Warn("Device is not capable of rate select", "id", ac.GetID(), "rate", ac.RateSelect())
	}

	if ac.AuxMode() != AuxHeatModeOff && !msmart.Contains(ac.supportedAuxModes, ac.AuxMode()) {
		slog.Warn("Device is not capable of aux mode", "mode", ac.AuxMode())
	}

	// Check if we need to refresh state before applying
	// This is needed for properties marked as SubmitModeRefreshFirst
	if ac.needsRefresh || ac.PowerOn() == nil {
		if err := ac.Refresh(ctx); err != nil {
			return err
		}
		ac.needsRefresh = false

		// No need to call applyPendingChanges() - pending state is already in pendingState map
	}

	// If powerState is still nil after refresh, return error
	// This prevents accidentally turning off the device
	if ac.PowerOn() == nil {
		return fmt.Errorf("failed to get device power state, cannot apply changes safely")
	}

	// Build SetStateCommand: Fill(Curr) + Fill(pending)
	cmd := NewSetStateCommand()

	// Fill 1: Apply current device state from lastKvState
	FillSetStateCommandFromMap(cmd, ac.lastKvState)

	// Fill 2: Apply user changes from pendingKvState (overrides current state)
	FillSetStateCommandFromMap(cmd, ac.pendingKvState)

	slog.Debug("发送设置命令", "targetTemp", cmd.TargetTemperature, "mode", cmd.OperationalMode, "fanSpeed", cmd.FanSpeed)
	slog.Info("发送设置命令", "targetTemp", cmd.TargetTemperature, "mode", cmd.OperationalMode, "fanSpeed", cmd.FanSpeed)

	// Process any state responses from the device
	responses, err := ac.sendCommandsGetResponses(ctx, []CommandInterface{cmd})
	if err != nil {
		return err
	}

	for _, response := range responses {
		ac.updateState(response)
	}

	// Done if no properties need updating
	if len(ac.pendingKvProp) == 0 {
		return nil
	}

	// Build property map from pending properties using propertyMap
	props := make(map[PropertyId]interface{})
	for prop := range ac.pendingKvProp {
		if val, ok := ac.getPropertyValue(prop); ok {
			props[prop] = val
		}
	}

	// Apply new properties
	err = ac.applyProperties(ctx, props)
	if err != nil {
		return err
	}

	// Reset pending properties set
	ac.pendingKvProp = make(map[PropertyId]any)

	// Clear pending state after successful apply
	ac.pendingKvState = make(map[StateId]any)

	return nil
}

// OverrideCapabilities overrides device capabilities via serialized dict
// This is the Go equivalent of Python's override_capabilities method
func (ac *AirConditioner) OverrideCapabilities(overrides map[string]interface{}, merge bool) error {
	// Get supported overrides from parent
	supportedOverrides := ac.Device.GetSupportedCapabilityOverrides()

	// Convert and apply each override
	for key, value := range overrides {
		// Check if override is allowed
		overrideInfo, exists := supportedOverrides[key]
		if !exists {
			return fmt.Errorf("unsupported capabilities override '%s'", key)
		}

		// Get target attribute and value type
		attrName := overrideInfo.AttrName
		valueType := overrideInfo.ValueType

		// Handle numeric overrides (float64)
		if valueType == reflect.TypeOf(float64(0)) {
			floatVal, ok := toFloat64(value)
			if !ok {
				return fmt.Errorf("'%s' must be a number", key)
			}
			ac.applyOverride(attrName, floatVal)
			continue
		}

		// Handle OperationalMode enum overrides
		if valueType == reflect.TypeOf(xc.OperationalMode(0)) {
			listVal, err := toOperationalModeList(value, merge, ac.supportedOpModes)
			if err != nil {
				return fmt.Errorf("'%s': %w", key, err)
			}
			ac.supportedOpModes = listVal
			continue
		}

		// Handle SwingMode enum overrides
		if valueType == reflect.TypeOf(xc.SwingMode(0)) {
			listVal, err := toSwingModeList(value, merge, ac.supportedSwingModes)
			if err != nil {
				return fmt.Errorf("'%s': %w", key, err)
			}
			ac.supportedSwingModes = listVal
			continue
		}

		// Handle FanSpeed enum overrides
		if valueType == reflect.TypeOf(xc.FanSpeed(0)) {
			listVal, err := toFanSpeedList(value, merge, ac.supportedFanSpeeds)
			if err != nil {
				return fmt.Errorf("'%s': %w", key, err)
			}
			ac.supportedFanSpeeds = listVal
			continue
		}

		// Handle AuxHeatMode enum overrides
		if valueType == reflect.TypeOf(AuxHeatMode(0)) {
			listVal, err := toAuxHeatModeList(value, merge, ac.supportedAuxModes)
			if err != nil {
				return fmt.Errorf("'%s': %w", key, err)
			}
			ac.supportedAuxModes = listVal
			continue
		}

		// Handle RateSelect enum overrides
		if valueType == reflect.TypeOf(RateSelect(0)) {
			listVal, err := toRateSelectList(value, merge, ac.supportedRateSelects)
			if err != nil {
				return fmt.Errorf("'%s': %w", key, err)
			}
			ac.supportedRateSelects = listVal
			continue
		}

		// Handle Capability flag overrides
		if valueType == reflect.TypeOf(Capability(0)) {
			flags, err := toCapabilityFlags(value, merge, Capability(ac.capabilities.Flags()))
			if err != nil {
				return fmt.Errorf("'%s': %w", key, err)
			}
			ac.capabilities.SetFlags(int64(flags))
			continue
		}
	}

	// Update supported properties from capabilities
	ac.updateSupportedProperties()

	return nil
}

// applyOverride applies a numeric override to a field
func (ac *AirConditioner) applyOverride(attrName string, value float64) {
	switch attrName {
	case "minTargetTemperature":
		ac.minTargetTemperature = value
	case "maxTargetTemperature":
		ac.maxTargetTemperature = value
	}
}

// toFloat64 attempts to convert a value to float64
func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	default:
		return 0, false
	}
}

// toOperationalModeList converts a value to a list of OperationalMode
func toOperationalModeList(value interface{}, merge bool, existing []xc.OperationalMode) ([]xc.OperationalMode, error) {
	list, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("must be a list")
	}

	var modes []xc.OperationalMode
	for _, v := range list {
		name, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value type in list")
		}
		mode, err := parseOperationalModeName(name)
		if err != nil {
			return nil, err
		}
		modes = append(modes, mode)
	}

	if merge {
		modeSet := make(map[xc.OperationalMode]bool)
		for _, m := range existing {
			modeSet[m] = true
		}
		for _, m := range modes {
			modeSet[m] = true
		}
		modes = make([]xc.OperationalMode, 0, len(modeSet))
		for m := range modeSet {
			modes = append(modes, m)
		}
	}

	return modes, nil
}

// toSwingModeList converts a value to a list of SwingMode
func toSwingModeList(value interface{}, merge bool, existing []xc.SwingMode) ([]xc.SwingMode, error) {
	list, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("must be a list")
	}

	var modes []xc.SwingMode
	for _, v := range list {
		name, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value type in list")
		}
		mode, err := parseSwingModeName(name)
		if err != nil {
			return nil, err
		}
		modes = append(modes, mode)
	}

	if merge {
		modeSet := make(map[xc.SwingMode]bool)
		for _, m := range existing {
			modeSet[m] = true
		}
		for _, m := range modes {
			modeSet[m] = true
		}
		modes = make([]xc.SwingMode, 0, len(modeSet))
		for m := range modeSet {
			modes = append(modes, m)
		}
	}

	return modes, nil
}

// toFanSpeedList converts a value to a list of fan speeds
func toFanSpeedList(value interface{}, merge bool, existing []interface{}) ([]interface{}, error) {
	list, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("must be a list")
	}

	var speeds []interface{}
	for _, v := range list {
		name, ok := v.(string)
		if !ok {
			// Allow raw numbers for custom fan speeds
			if num, ok := v.(float64); ok {
				speeds = append(speeds, int(num))
				continue
			}
			return nil, fmt.Errorf("invalid value type in list")
		}
		speed, err := parseFanSpeedName(name)
		if err != nil {
			return nil, err
		}
		speeds = append(speeds, speed)
	}

	if merge {
		speedSet := make(map[interface{}]bool)
		for _, s := range existing {
			speedSet[s] = true
		}
		for _, s := range speeds {
			speedSet[s] = true
		}
		speeds = make([]interface{}, 0, len(speedSet))
		for s := range speedSet {
			speeds = append(speeds, s)
		}
	}

	return speeds, nil
}

// toAuxHeatModeList converts a value to a list of AuxHeatMode
func toAuxHeatModeList(value interface{}, merge bool, existing []AuxHeatMode) ([]AuxHeatMode, error) {
	list, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("must be a list")
	}

	var modes []AuxHeatMode
	for _, v := range list {
		name, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value type in list")
		}
		mode, err := parseAuxHeatModeName(name)
		if err != nil {
			return nil, err
		}
		modes = append(modes, mode)
	}

	if merge {
		modeSet := make(map[AuxHeatMode]bool)
		for _, m := range existing {
			modeSet[m] = true
		}
		for _, m := range modes {
			modeSet[m] = true
		}
		modes = make([]AuxHeatMode, 0, len(modeSet))
		for m := range modeSet {
			modes = append(modes, m)
		}
	}

	return modes, nil
}

// toRateSelectList converts a value to a list of RateSelect
func toRateSelectList(value interface{}, merge bool, existing []RateSelect) ([]RateSelect, error) {
	list, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("must be a list")
	}

	var rates []RateSelect
	for _, v := range list {
		name, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("invalid value type in list")
		}
		rate, err := parseRateSelectName(name)
		if err != nil {
			return nil, err
		}
		rates = append(rates, rate)
	}

	if merge {
		rateSet := make(map[RateSelect]bool)
		for _, r := range existing {
			rateSet[r] = true
		}
		for _, r := range rates {
			rateSet[r] = true
		}
		rates = make([]RateSelect, 0, len(rateSet))
		for r := range rateSet {
			rates = append(rates, r)
		}
	}

	return rates, nil
}

// toCapabilityFlags converts a value to Capability flags
func toCapabilityFlags(value interface{}, merge bool, existing Capability) (Capability, error) {
	list, ok := value.([]interface{})
	if !ok {
		return 0, fmt.Errorf("must be a list")
	}

	flags := Capability(0)
	for _, v := range list {
		name, ok := v.(string)
		if !ok {
			return 0, fmt.Errorf("invalid value type in list")
		}
		flag, err := parseCapabilityName(name)
		if err != nil {
			return 0, err
		}
		flags |= flag
	}

	if merge {
		flags |= existing
	}

	return flags, nil
}

// Enum name parsing functions
func parseOperationalModeName(name string) (xc.OperationalMode, error) {
	switch name {
	case "AUTO":
		return xc.OperationalModeAuto, nil
	case "COOL":
		return xc.OperationalModeCool, nil
	case "DRY":
		return xc.OperationalModeDry, nil
	case "HEAT":
		return xc.OperationalModeHeat, nil
	case "FAN_ONLY":
		return xc.OperationalModeFanOnly, nil
	case "SMART_DRY":
		return xc.OperationalModeSmartDry, nil
	default:
		return 0, fmt.Errorf("invalid OperationalMode name: %s", name)
	}
}

func parseSwingModeName(name string) (xc.SwingMode, error) {
	switch name {
	case "OFF":
		return xc.SwingModeOff, nil
	case "VERTICAL":
		return xc.SwingModeVertical, nil
	case "HORIZONTAL":
		return xc.SwingModeHorizontal, nil
	case "BOTH":
		return xc.SwingModeBoth, nil
	default:
		return 0, fmt.Errorf("invalid SwingMode name: %s", name)
	}
}

func parseFanSpeedName(name string) (xc.FanSpeed, error) {
	switch name {
	case "AUTO":
		return xc.FanSpeedAuto, nil
	case "MAX":
		return xc.FanSpeedMax, nil
	case "HIGH":
		return xc.FanSpeedHigh, nil
	case "MEDIUM":
		return xc.FanSpeedMedium, nil
	case "LOW":
		return xc.FanSpeedLow, nil
	case "SILENT":
		return xc.FanSpeedSilent, nil
	default:
		return 0, fmt.Errorf("invalid FanSpeed name: %s", name)
	}
}

func parseAuxHeatModeName(name string) (AuxHeatMode, error) {
	switch name {
	case "OFF":
		return AuxHeatModeOff, nil
	case "AUX_HEAT":
		return AuxHeatModeAuxHeat, nil
	case "AUX_ONLY":
		return AuxHeatModeAuxOnly, nil
	default:
		return 0, fmt.Errorf("invalid AuxHeatMode name: %s", name)
	}
}

func parseRateSelectName(name string) (RateSelect, error) {
	switch name {
	case "OFF":
		return RateSelectOff, nil
	case "GEAR_50":
		return RateSelectGear50, nil
	case "GEAR_75":
		return RateSelectGear75, nil
	case "LEVEL_1":
		return RateSelectLevel1, nil
	case "LEVEL_2":
		return RateSelectLevel2, nil
	case "LEVEL_3":
		return RateSelectLevel3, nil
	case "LEVEL_4":
		return RateSelectLevel4, nil
	case "LEVEL_5":
		return RateSelectLevel5, nil
	default:
		return 0, fmt.Errorf("invalid RateSelect name: %s", name)
	}
}

func parseCapabilityName(name string) (Capability, error) {
	switch name {
	case "CUSTOM_FAN_SPEED":
		return CapabilityCustomFanSpeed, nil
	case "ECO":
		return CapabilityEco, nil
	case "FREEZE_PROTECTION":
		return CapabilityFreezeProtection, nil
	case "IECO":
		return CapabilityIECO, nil
	case "TURBO":
		return CapabilityTurbo, nil
	case "DISPLAY_CONTROL":
		return CapabilityDisplayControl, nil
	case "ENERGY_STATS":
		return CapabilityEnergyStats, nil
	case "FILTER_REMINDER":
		return CapabilityFilterReminder, nil
	case "HUMIDITY":
		return CapabilityHumidity, nil
	case "TARGET_HUMIDITY":
		return CapabilityTargetHumidity, nil
	case "SWING_HORIZONTAL_ANGLE":
		return CapabilitySwingHorizontalAngle, nil
	case "SWING_VERTICAL_ANGLE":
		return CapabilitySwingVerticalAngle, nil
	case "BREEZE_AWAY":
		return CapabilityBreezeAway, nil
	case "BREEZE_CONTROL":
		return CapabilityBreezeControl, nil
	case "BREEZELESS":
		return CapabilityBreezeless, nil
	case "CASCADE":
		return CapabilityCascade, nil
	case "JET_COOL":
		return CapabilityJetCool, nil
	case "OUT_SILENT":
		return CapabilityOutSilent, nil
	case "PURIFIER":
		return CapabilityPurifier, nil
	case "SELF_CLEAN":
		return CapabilitySelfClean, nil
	default:
		return 0, fmt.Errorf("invalid Capability name: %s", name)
	}
}

// ============================================================================
// Property getters and setters
// ============================================================================

// MinTargetTemperature returns the minimum target temperature
func (ac *AirConditioner) MinTargetTemperature() float64 {
	return ac.minTargetTemperature
}

// MaxTargetTemperature returns the maximum target temperature
func (ac *AirConditioner) MaxTargetTemperature() float64 {
	return ac.maxTargetTemperature
}

// SupportedOperationModes returns the supported operation modes
func (ac *AirConditioner) SupportedOperationModes() []xc.OperationalMode {
	return ac.supportedOpModes
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

// SupportsBreezeMild returns whether breeze mild is supported
func (ac *AirConditioner) SupportsBreezeMild() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityBreezeControl)
}

// SupportsBreezeless returns whether breezeless is supported
func (ac *AirConditioner) SupportsBreezeless() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityBreezeless) || caps.Has(CapabilityBreezeControl)
}

// SupportedSwingModes returns the supported swing modes
func (ac *AirConditioner) SupportedSwingModes() []xc.SwingMode {
	return ac.supportedSwingModes
}

// SupportsHorizontalSwingAngle returns whether horizontal swing angle is supported
func (ac *AirConditioner) SupportsHorizontalSwingAngle() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilitySwingHorizontalAngle)
}

// SupportsVerticalSwingAngle returns whether vertical swing angle is supported
func (ac *AirConditioner) SupportsVerticalSwingAngle() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilitySwingVerticalAngle)
}

// SupportsCascade returns whether cascade is supported
func (ac *AirConditioner) SupportsCascade() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityCascade)
}

// SupportsEco returns whether eco mode is supported
func (ac *AirConditioner) SupportsEco() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityEco)
}

// SupportsIECO returns whether IECO is supported
func (ac *AirConditioner) SupportsIECO() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityIECO)
}

// SupportsFlashCool returns whether flash cool is supported
func (ac *AirConditioner) SupportsFlashCool() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityJetCool)
}

// SupportsTurbo returns whether turbo mode is supported
func (ac *AirConditioner) SupportsTurbo() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityTurbo)
}

// SupportsFreezeProtection returns whether freeze protection is supported
func (ac *AirConditioner) SupportsFreezeProtection() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityFreezeProtection)
}

// SupportsPurifier returns whether purifier is supported
func (ac *AirConditioner) SupportsPurifier() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityPurifier)
}

// SupportsDisplayControl returns whether display control is supported
func (ac *AirConditioner) SupportsDisplayControl() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityDisplayControl)
}

// SupportsFilterReminder returns whether filter reminder is supported
func (ac *AirConditioner) SupportsFilterReminder() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityFilterReminder)
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

// SupportsTargetHumidity returns whether target humidity is supported
func (ac *AirConditioner) SupportsTargetHumidity() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityTargetHumidity)
}

// SupportsSelfClean returns whether self clean is supported
func (ac *AirConditioner) SupportsSelfClean() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilitySelfClean)
}

// SupportedRateSelects returns the supported rate selects
func (ac *AirConditioner) SupportedRateSelects() []RateSelect {
	return ac.supportedRateSelects
}

// SupportedAuxModes returns the supported aux modes
func (ac *AirConditioner) SupportedAuxModes() []AuxHeatMode {
	return ac.supportedAuxModes
}

// EnableGroup5DataRequests returns whether group 5 data requests are enabled
func (ac *AirConditioner) EnableGroup5DataRequests() bool {
	return ac.requestGroup5Data
}

// SetEnableGroup5DataRequests sets whether group 5 data requests are enabled
func (ac *AirConditioner) SetEnableGroup5DataRequests(enable bool) {
	ac.requestGroup5Data = enable
}

// SupportsOutSilent returns whether out silent is supported
func (ac *AirConditioner) SupportsOutSilent() bool {
	caps := Capability(ac.capabilities.Flags())
	return caps.Has(CapabilityOutSilent)
}

// ToDict returns the device state as a dictionary
// This is the Go equivalent of Python's to_dict method
func (ac *AirConditioner) ToDict() map[string]interface{} {
	// Start with base device info
	result := ac.Device.ToDict()

	// Basic state
	if powerOn := ac.PowerOn(); powerOn != nil {
		result["power_on"] = *powerOn
	}
	result["target_temperature"] = ac.TargetTemperature()
	result["operational_mode"] = ac.OperationalMode()
	result["fan_speed"] = ac.FanSpeed()
	result["swing_mode"] = ac.SwingMode()

	// Optional features
	result["eco"] = ac.Eco()
	result["turbo"] = ac.Turbo()
	result["freeze_protection"] = ac.FreezeProtection()
	result["sleep"] = ac.Sleep()
	result["fahrenheit"] = ac.FahrenheitUnit()
	result["follow_me"] = ac.FollowMe()
	result["purifier"] = ac.Purifier()

	// Temperatures
	if indoorTemp := ac.IndoorTemperature(); indoorTemp != nil {
		result["indoor_temperature"] = *indoorTemp
	}
	if outdoorTemp := ac.OutdoorTemperature(); outdoorTemp != nil {
		result["outdoor_temperature"] = *outdoorTemp
	}

	// Display
	if displayOn := ac.DisplayOn(); displayOn != nil {
		result["display_on"] = *displayOn
	}

	// Humidity
	if indoorHumidity := ac.IndoorHumidity(); indoorHumidity != nil {
		result["indoor_humidity"] = *indoorHumidity
	}
	result["target_humidity"] = ac.TargetHumidity()

	// Swing angles
	result["horizontal_swing_angle"] = ac.HorizontalSwingAngle()
	result["vertical_swing_angle"] = ac.VerticalSwingAngle()

	// Other features
	result["cascade_mode"] = ac.CascadeMode()
	result["rate_select"] = ac.RateSelect()
	result["breeze_mode"] = ac.BreezeMode()
	result["aux_mode"] = ac.AuxMode()

	// IECO and FlashCool
	result["ieco"] = ac.Ieco()
	result["flash_cool"] = ac.FlashCool()
	result["out_silent"] = ac.OutSilent()

	// Beep
	result["beep_on"] = ac.BeepOn()

	// Error code
	if errorCode := ac.ErrorCode(); errorCode != nil {
		result["error_code"] = *errorCode
	}

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
		"min_target_temperature":  ac.minTargetTemperature,
		"max_target_temperature":  ac.maxTargetTemperature,
		"supported_modes":         ac.supportedOpModes,
		"supported_swing_modes":   ac.supportedSwingModes,
		"supported_fan_speeds":    ac.supportedFanSpeeds,
		"supported_aux_modes":     ac.supportedAuxModes,
		"supported_rate_selects":  ac.supportedRateSelects,
		"additional_capabilities": flags,
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

// SleepMode is deprecated. Use Sleep instead.
// Deprecated: Use Sleep instead.
func (ac *AirConditioner) SleepMode() bool {
	msmart.Deprecated("SleepMode", "Sleep", "")
	return ac.Sleep()
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

//////////////////////////////////////

func (ac *AirConditioner) Authenticate(ctx context.Context, token device.Token, key device.Key) error {
	return ac.Device.Authenticate(ctx, token, key)
}

// AuthenticateFromPreset implements [device.Device].
func (ac *AirConditioner) AuthenticateFromPreset(ctx context.Context) (bool, error) {
	return ac.Device.AuthenticateFromPreset(ctx)
}

// GetID implements [device.Device].
func (ac *AirConditioner) GetID() string { return ac.Device.GetID() }

// GetIP implements [device.Device].
func (ac *AirConditioner) GetIP() string { return ac.Device.GetIP() }

// GetName implements [device.Device].
func (ac *AirConditioner) GetName() string { return ac.Device.GetName() }

// GetOnline implements [device.Device].
func (ac *AirConditioner) GetOnline() bool {
	return ac.Device.GetOnline()
}

// GetPort implements [device.Device].
func (ac *AirConditioner) GetPort() int {
	return ac.Device.GetPort()
}

// GetSN implements [device.Device].
func (ac *AirConditioner) GetSN() string {
	return ac.Device.GetSN()
}

// GetSupported implements [device.Device].
func (ac *AirConditioner) GetSupported() bool {
	return ac.Device.GetSupported()
}

// GetType implements [device.Device].
func (ac *AirConditioner) GetType() device.DeviceType {
	return ac.Device.GetType()
}

// GetVersion implements [device.Device].
func (ac *AirConditioner) GetVersion() int {
	return ac.Device.GetVersion()
}

// SendBytes implements [device.Device].
func (ac *AirConditioner) SendBytes(ctx context.Context, data []byte) ([][]byte, error) {
	return ac.Device.SendBytes(ctx, data)
}

// SetMaxConnectionLifetime implements [device.Device].
func (ac *AirConditioner) SetMaxConnectionLifetime(seconds *int) {
	ac.Device.SetMaxConnectionLifetime(seconds)
}

// GetKeyInfo implements [device.DeviceAuthV3].
func (ac *AirConditioner) GetKeyInfo() (device.Token, device.Key, device.LocalKey, time.Time) {
	return ac.Device.GetKeyInfo()
}

// IsAuthenticated implements [device.DeviceAuthV3].
func (ac *AirConditioner) IsAuthenticated() bool {
	return ac.Device.IsAuthenticated()
}

// ============================================================================
// Property Map
// ============================================================================

// propertyMap defines the mapping from PropertyId to a function that returns the property value
// This is the Go equivalent of Python's _PROPERTY_MAP
var propertyMap = map[PropertyId]func(*AirConditioner) interface{}{
	PropertyIdBreezeAway:    func(ac *AirConditioner) interface{} { return ac.BreezeMode() == BreezeModeBreezeAway },
	PropertyIdBreezeControl: func(ac *AirConditioner) interface{} { return ac.BreezeMode() },
	PropertyIdBreezeless:    func(ac *AirConditioner) interface{} { return ac.BreezeMode() == BreezeModeBreezeless },
	PropertyIdCascade:       func(ac *AirConditioner) interface{} { return ac.CascadeMode() },
	PropertyIdIECO:          func(ac *AirConditioner) interface{} { return ac.Ieco() },
	PropertyIdJetCool:       func(ac *AirConditioner) interface{} { return ac.FlashCool() },
	PropertyIdOutSilent:     func(ac *AirConditioner) interface{} { return ac.OutSilent() },
	PropertyIdRateSelect:    func(ac *AirConditioner) interface{} { return ac.RateSelect() },
	PropertyIdSwingLrAngle:  func(ac *AirConditioner) interface{} { return ac.HorizontalSwingAngle() },
	PropertyIdSwingUdAngle:  func(ac *AirConditioner) interface{} { return ac.VerticalSwingAngle() },
}

// getPropertyValue returns the current value of a property
// This is a helper method that uses propertyMap to get property values
func (ac *AirConditioner) getPropertyValue(prop PropertyId) (interface{}, bool) {
	fn, ok := propertyMap[prop]
	if !ok {
		return nil, false
	}
	return fn(ac), true
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

// SetBeep sets the beep state (alias for SetBeepOn to match interface)
func (ac *AirConditioner) SetBeep(beep bool) {
	ac.SetBeepOn(beep)
}

// SetFahrenheit sets the Fahrenheit unit display (alias for SetFahrenheitUnit to match interface)
func (ac *AirConditioner) SetFahrenheit(f bool) {
	ac.SetFahrenheitUnit(f)
}

// SetBreezeAway sets the breeze away mode
func (ac *AirConditioner) SetBreezeAway(enable bool) {
	if enable {
		ac._setProp(PropertyIdBreezeControl, byte(BreezeModeBreezeAway))
	} else {
		ac._setProp(PropertyIdBreezeControl, byte(BreezeModeOff))
	}
}

// SetBreezeMild sets the breeze mild mode
func (ac *AirConditioner) SetBreezeMild(enable bool) {
	if enable {
		ac._setProp(PropertyIdBreezeControl, byte(BreezeModeBreezeMild))
	} else {
		ac._setProp(PropertyIdBreezeControl, byte(BreezeModeOff))
	}
}

// SetBreezeless sets the breezeless mode
func (ac *AirConditioner) SetBreezeless(enable bool) {
	if enable {
		ac._setProp(PropertyIdBreezeControl, byte(BreezeModeBreezeless))
	} else {
		ac._setProp(PropertyIdBreezeControl, byte(BreezeModeOff))
	}
}

// SetIECO sets the IECO mode
func (ac *AirConditioner) SetIECO(enable bool) {
	ac._setProp(PropertyIdIECO, enable)
}
