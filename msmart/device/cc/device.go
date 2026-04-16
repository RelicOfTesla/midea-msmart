// Package cc provides commercial air conditioner device support.
// This is a translation from msmart-ng Python library.
// Original file: msmart/device/CC/device.py
package cc

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	msmart "github.com/RelicOfTesla/midea-msmart/msmart"
)

// Logger for the package
var logger = slog.Default()

// DefaultSendRetries is the default number of retries for sending commands
const DefaultSendRetries = 3

// FanSpeed represents fan speed levels.
// In Python: class FanSpeed(MideaIntEnum)
type FanSpeed byte

const (
	FanSpeedL1   FanSpeed = 0x01
	FanSpeedL2   FanSpeed = 0x02
	FanSpeedL3   FanSpeed = 0x03
	FanSpeedL4   FanSpeed = 0x04
	FanSpeedL5   FanSpeed = 0x05
	FanSpeedL6   FanSpeed = 0x06
	FanSpeedL7   FanSpeed = 0x07
	FanSpeedAuto FanSpeed = 0x08

	FanSpeedDefault FanSpeed = FanSpeedAuto
)

// Value returns the byte value of the fan speed.
func (f FanSpeed) Value() byte {
	return byte(f)
}

// String returns the string representation of the fan speed.
func (f FanSpeed) String() string {
	switch f {
	case FanSpeedL1:
		return "L1"
	case FanSpeedL2:
		return "L2"
	case FanSpeedL3:
		return "L3"
	case FanSpeedL4:
		return "L4"
	case FanSpeedL5:
		return "L5"
	case FanSpeedL6:
		return "L6"
	case FanSpeedL7:
		return "L7"
	case FanSpeedAuto:
		return "AUTO"
	default:
		return fmt.Sprintf("FanSpeed(%d)", byte(f))
	}
}

// List returns all fan speed values.
// In Python: @classmethod def list(cls) -> list[MideaIntEnum]
func FanSpeedList() []FanSpeed {
	return []FanSpeed{
		FanSpeedL1, FanSpeedL2, FanSpeedL3, FanSpeedL4,
		FanSpeedL5, FanSpeedL6, FanSpeedL7, FanSpeedAuto,
	}
}

// FanSpeedFromValue gets a FanSpeed from a value.
// In Python: @classmethod def get_from_value(cls, value: Optional[int], default: Optional[MideaIntEnum] = None)
func FanSpeedFromValue(value byte) FanSpeed {
	switch value {
	case 0x01:
		return FanSpeedL1
	case 0x02:
		return FanSpeedL2
	case 0x03:
		return FanSpeedL3
	case 0x04:
		return FanSpeedL4
	case 0x05:
		return FanSpeedL5
	case 0x06:
		return FanSpeedL6
	case 0x07:
		return FanSpeedL7
	case 0x08:
		return FanSpeedAuto
	default:
		logger.Debug("Unknown FanSpeed", "value", value)
		return FanSpeedDefault
	}
}

// OperationalMode represents AC operational modes.
// In Python: class OperationalMode(MideaIntEnum)
type OperationalMode byte

const (
	OperationalModeFan  OperationalMode = 0x01
	OperationalModeCool OperationalMode = 0x02
	OperationalModeHeat OperationalMode = 0x03
	OperationalModeAuto OperationalMode = 0x05
	OperationalModeDry  OperationalMode = 0x06

	OperationalModeDefault OperationalMode = OperationalModeFan
)

// Value returns the byte value of the operational mode.
func (m OperationalMode) Value() byte {
	return byte(m)
}

// String returns the string representation of the operational mode.
func (m OperationalMode) String() string {
	switch m {
	case OperationalModeFan:
		return "FAN"
	case OperationalModeCool:
		return "COOL"
	case OperationalModeHeat:
		return "HEAT"
	case OperationalModeAuto:
		return "AUTO"
	case OperationalModeDry:
		return "DRY"
	default:
		return fmt.Sprintf("OperationalMode(%d)", byte(m))
	}
}

// List returns all operational mode values.
// In Python: @classmethod def list(cls) -> list[MideaIntEnum]
func OperationalModeList() []OperationalMode {
	return []OperationalMode{
		OperationalModeFan, OperationalModeCool, OperationalModeHeat,
		OperationalModeAuto, OperationalModeDry,
	}
}

// OperationalModeFromValue gets an OperationalMode from a value.
func OperationalModeFromValue(value byte) OperationalMode {
	switch value {
	case 0x01:
		return OperationalModeFan
	case 0x02:
		return OperationalModeCool
	case 0x03:
		return OperationalModeHeat
	case 0x05:
		return OperationalModeAuto
	case 0x06:
		return OperationalModeDry
	default:
		logger.Debug("Unknown OperationalMode", "value", value)
		return OperationalModeDefault
	}
}

// SwingMode represents swing modes.
// In Python: class SwingMode(MideaIntEnum)
type SwingMode byte

const (
	SwingModeOff       SwingMode = 0x0
	SwingModeVertical  SwingMode = 0x1
	SwingModeHorizontal SwingMode = 0x2
	SwingModeBoth      SwingMode = 0x3

	SwingModeDefault SwingMode = SwingModeOff
)

// Value returns the byte value of the swing mode.
func (s SwingMode) Value() byte {
	return byte(s)
}

// String returns the string representation of the swing mode.
func (s SwingMode) String() string {
	switch s {
	case SwingModeOff:
		return "OFF"
	case SwingModeVertical:
		return "VERTICAL"
	case SwingModeHorizontal:
		return "HORIZONTAL"
	case SwingModeBoth:
		return "BOTH"
	default:
		return fmt.Sprintf("SwingMode(%d)", byte(s))
	}
}

// List returns all swing mode values.
func SwingModeList() []SwingMode {
	return []SwingMode{
		SwingModeOff, SwingModeVertical, SwingModeHorizontal, SwingModeBoth,
	}
}

// SwingModeFromValue gets a SwingMode from a value.
func SwingModeFromValue(value byte) SwingMode {
	switch value {
	case 0x0:
		return SwingModeOff
	case 0x1:
		return SwingModeVertical
	case 0x2:
		return SwingModeHorizontal
	case 0x3:
		return SwingModeBoth
	default:
		logger.Debug("Unknown SwingMode", "value", value)
		return SwingModeDefault
	}
}

// SwingAngle represents swing angles.
// In Python: class SwingAngle(MideaIntEnum)
type SwingAngle byte

const (
	SwingAngleClose SwingAngle = 0x00  // TODO unverified
	SwingAnglePos1  SwingAngle = 0x01
	SwingAnglePos2  SwingAngle = 0x02
	SwingAnglePos3  SwingAngle = 0x03
	SwingAnglePos4  SwingAngle = 0x04
	SwingAnglePos5  SwingAngle = 0x05
	SwingAngleAuto  SwingAngle = 0x06

	SwingAngleDefault SwingAngle = SwingAnglePos3
)

// Value returns the byte value of the swing angle.
func (s SwingAngle) Value() byte {
	return byte(s)
}

// String returns the string representation of the swing angle.
func (s SwingAngle) String() string {
	switch s {
	case SwingAngleClose:
		return "CLOSE"
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
	case SwingAngleAuto:
		return "AUTO"
	default:
		return fmt.Sprintf("SwingAngle(%d)", byte(s))
	}
}

// List returns all swing angle values.
func SwingAngleList() []SwingAngle {
	return []SwingAngle{
		SwingAngleClose, SwingAnglePos1, SwingAnglePos2, SwingAnglePos3,
		SwingAnglePos4, SwingAnglePos5, SwingAngleAuto,
	}
}

// SwingAngleFromValue gets a SwingAngle from a value.
func SwingAngleFromValue(value byte) SwingAngle {
	switch value {
	case 0x00:
		return SwingAngleClose
	case 0x01:
		return SwingAnglePos1
	case 0x02:
		return SwingAnglePos2
	case 0x03:
		return SwingAnglePos3
	case 0x04:
		return SwingAnglePos4
	case 0x05:
		return SwingAnglePos5
	case 0x06:
		return SwingAngleAuto
	default:
		logger.Debug("Unknown SwingAngle", "value", value)
		return SwingAngleDefault
	}
}

// PurifierMode represents purifier modes.
// In Python: class PurifierMode(MideaIntEnum)
type PurifierMode byte

const (
	PurifierModeAuto PurifierMode = 0x00
	PurifierModeOn   PurifierMode = 0x01
	PurifierModeOff  PurifierMode = 0x02

	PurifierModeDefault PurifierMode = PurifierModeOff
)

// Value returns the byte value of the purifier mode.
func (p PurifierMode) Value() byte {
	return byte(p)
}

// String returns the string representation of the purifier mode.
func (p PurifierMode) String() string {
	switch p {
	case PurifierModeAuto:
		return "AUTO"
	case PurifierModeOn:
		return "ON"
	case PurifierModeOff:
		return "OFF"
	default:
		return fmt.Sprintf("PurifierMode(%d)", byte(p))
	}
}

// List returns all purifier mode values.
func PurifierModeList() []PurifierMode {
	return []PurifierMode{PurifierModeAuto, PurifierModeOn, PurifierModeOff}
}

// PurifierModeFromValue gets a PurifierMode from a value.
func PurifierModeFromValue(value byte) PurifierMode {
	switch value {
	case 0x00:
		return PurifierModeAuto
	case 0x01:
		return PurifierModeOn
	case 0x02:
		return PurifierModeOff
	default:
		logger.Debug("Unknown PurifierMode", "value", value)
		return PurifierModeDefault
	}
}

// AuxHeatMode represents auxiliary heat modes.
// In Python: class AuxHeatMode(MideaIntEnum)
type AuxHeatMode byte

const (
	AuxHeatModeAuto AuxHeatMode = 0x00
	AuxHeatModeOn   AuxHeatMode = 0x01
	AuxHeatModeOff  AuxHeatMode = 0x02

	AuxHeatModeDefault AuxHeatMode = AuxHeatModeOff
)

// Value returns the byte value of the aux heat mode.
func (a AuxHeatMode) Value() byte {
	return byte(a)
}

// String returns the string representation of the aux heat mode.
func (a AuxHeatMode) String() string {
	switch a {
	case AuxHeatModeAuto:
		return "AUTO"
	case AuxHeatModeOn:
		return "ON"
	case AuxHeatModeOff:
		return "OFF"
	default:
		return fmt.Sprintf("AuxHeatMode(%d)", byte(a))
	}
}

// List returns all aux heat mode values.
func AuxHeatModeList() []AuxHeatMode {
	return []AuxHeatMode{AuxHeatModeAuto, AuxHeatModeOn, AuxHeatModeOff}
}

// AuxHeatModeFromValue gets an AuxHeatMode from a value.
func AuxHeatModeFromValue(value byte) AuxHeatMode {
	switch value {
	case 0x00:
		return AuxHeatModeAuto
	case 0x01:
		return AuxHeatModeOn
	case 0x02:
		return AuxHeatModeOff
	default:
		logger.Debug("Unknown AuxHeatMode", "value", value)
		return AuxHeatModeDefault
	}
}

// Capability represents device capability flags.
// In Python: class Capability(Flag)
type Capability int64

const (
	// Presets
	CapabilityECO    Capability = 1 << iota  // 1
	CapabilitySilent                          // 2
	CapabilitySleep                           // 4

	// Swing
	CapabilitySwingHorizontalAngle  // 8
	CapabilitySwingVerticalAngle    // 16

	// Misc
	CapabilityHumidity  // 32
	CapabilityPurifier  // 64
)

// Default capabilities
const CapabilityDefault Capability = CapabilityECO | CapabilitySilent | CapabilitySleep |
	CapabilitySwingHorizontalAngle | CapabilitySwingVerticalAngle | CapabilityHumidity

// Has checks if a capability flag is set.
// In Python: bool(self._flags & flag)
func (c Capability) Has(flag Capability) bool {
	return c&flag != 0
}

// Set enables or disables a capability flag.
// In Python: self._flags |= flag or self._flags &= ~flag
func (c *Capability) Set(flag Capability, enable bool) {
	if enable {
		*c |= flag
	} else {
		*c &= ^flag
	}
}

// CapabilityManager manages device capabilities.
// In Python: self._capabilities = CapabilityManager(CommercialAirConditioner.Capability.DEFAULT)
type CapabilityManager struct {
	flags Capability
}

// NewCapabilityManager creates a new CapabilityManager with default capabilities.
func NewCapabilityManager(default_ Capability) *CapabilityManager {
	return &CapabilityManager{flags: default_}
}

// Has checks if a capability flag is set.
func (cm *CapabilityManager) Has(flag Capability) bool {
	return cm.flags.Has(flag)
}

// Set enables or disables a capability flag.
func (cm *CapabilityManager) Set(flag Capability, enable bool) {
	cm.flags.Set(flag, enable)
}

// Value returns the integer value of the flags.
func (cm *CapabilityManager) Value() int64 {
	return int64(cm.flags)
}

// Flags returns the current flags.
func (cm *CapabilityManager) Flags() Capability {
	return cm.flags
}

// SetFlags sets the flags.
func (cm *CapabilityManager) SetFlags(flags Capability) {
	cm.flags = flags
}

// Note: Device is now embedded from msmart.Device instead of being defined locally.
// This follows the correct inheritance pattern matching the Python implementation.

// CommercialAirConditioner represents a commercial air conditioner device.
// In Python: class CommercialAirConditioner(Device)
type CommercialAirConditioner struct {
	*msmart.Device

	// Basic controls
	powerState         bool
	targetTemperature  float64
	targetHumidity     byte
	operationalMode    OperationalMode
	fanSpeed           FanSpeed
	horizontalSwingAngle SwingAngle
	verticalSwingAngle   SwingAngle

	eco     bool
	silent  bool
	sleep   bool
	purifier PurifierMode
	auxMode  AuxHeatMode

	fahrenheit bool
	// displayOn  bool // TODO

	// Sensors
	indoorTemperature  *float64
	outdoorTemperature *float64
	indoorHumidity     *byte

	updatedControls map[ControlId]bool

	// Capabilities
	minTargetTemperature float64
	maxTargetTemperature float64
	capabilities         *CapabilityManager

	supportedOpModes       []OperationalMode
	supportedSwingModes    []SwingMode
	supportedFanSpeeds     []FanSpeed
	supportedPurifierModes []PurifierMode
	supportedAuxModes      []AuxHeatMode
}

// NewCommercialAirConditioner creates a new CommercialAirConditioner instance.
// In Python: def __init__(self, ip: str, device_id: int, port: int, **kwargs)
func NewCommercialAirConditioner(ip string, deviceID int, port int) *CommercialAirConditioner {
	ac := &CommercialAirConditioner{
		Device: msmart.NewDevice(ip, port, deviceID, msmart.DeviceTypeCommercialAC),

		// Basic controls
		powerState:         false,
		targetTemperature:  17.0,
		targetHumidity:     40,
		operationalMode:    OperationalModeDefault,
		fanSpeed:           FanSpeedDefault,
		horizontalSwingAngle: SwingAngleDefault,
		verticalSwingAngle:   SwingAngleDefault,

		eco:     false,
		silent:  false,
		sleep:   false,
		purifier: PurifierModeDefault,
		auxMode:  AuxHeatModeDefault,

		fahrenheit: false,

		updatedControls: make(map[ControlId]bool),

		minTargetTemperature: 17,
		maxTargetTemperature: 30,
		capabilities:         NewCapabilityManager(CapabilityDefault),

		supportedOpModes:       OperationalModeList(),
		supportedSwingModes:    SwingModeList(),
		supportedFanSpeeds:     FanSpeedList(),
		supportedPurifierModes: PurifierModeList(),
		supportedAuxModes:      AuxHeatModeList(),
	}

	// Set supported capability overrides
	// This is the Go equivalent of Python's _SUPPORTED_CAPABILITY_OVERRIDES
	ac.SetSupportedCapabilityOverrides(map[string]msmart.CapabilityOverrideInfo{
		"min_target_temperature":   {AttrName: "minTargetTemperature", ValueType: reflect.TypeOf(float64(0))},
		"max_target_temperature":   {AttrName: "maxTargetTemperature", ValueType: reflect.TypeOf(float64(0))},
		"supported_modes":          {AttrName: "supportedOpModes", ValueType: reflect.TypeOf(OperationalMode(0))},
		"supported_swing_modes":    {AttrName: "supportedSwingModes", ValueType: reflect.TypeOf(SwingMode(0))},
		"supported_fan_speeds":     {AttrName: "supportedFanSpeeds", ValueType: reflect.TypeOf(FanSpeed(0))},
		"supported_aux_modes":      {AttrName: "supportedAuxModes", ValueType: reflect.TypeOf(AuxHeatMode(0))},
		"supported_purifier_modes": {AttrName: "supportedPurifierModes", ValueType: reflect.TypeOf(PurifierMode(0))},
		"additional_capabilities":  {AttrName: "capabilities", ValueType: reflect.TypeOf(Capability(0))},
	})

	return ac
}

// updateState updates the local state from a device state response.
// In Python: def _update_state(self, res: Response) -> None
func (ac *CommercialAirConditioner) updateState(res interface{}) {
	switch r := res.(type) {
	case *QueryResponse:
		logger.Debug("Query response payload from device", "id", ac.GetID(), "response", r)

		ac.powerState = r.PowerOn
		ac.targetTemperature = r.TargetTemperature
		ac.indoorTemperature = r.IndoorTemperature
		ac.outdoorTemperature = r.OutdoorTemperature
		ac.fahrenheit = r.Fahrenheit
		ac.targetHumidity = r.TargetHumidity
		ac.indoorHumidity = r.IndoorHumidity

		ac.operationalMode = OperationalModeFromValue(r.OperationalMode)
		ac.fanSpeed = FanSpeedFromValue(r.FanSpeed)
		ac.verticalSwingAngle = SwingAngleFromValue(r.VertSwingAngle)
		ac.horizontalSwingAngle = SwingAngleFromValue(r.HorzSwingAngle)

		// TODO wind sense
		// self._soft = res.soft

		ac.eco = r.Eco
		ac.silent = r.Silent
		ac.sleep = r.Sleep

		// self._display_on = res.display  // TODO?

		ac.purifier = PurifierModeFromValue(r.Purifier)
		ac.auxMode = AuxHeatModeFromValue(r.AuxMode)

	default:
		logger.Debug("Ignored unknown response from device", "id", ac.GetID(), "response", res)
	}
}

// updateCapabilities updates device capabilities.
// In Python: def _update_capabilities(self, res: QueryResponse) -> None
func (ac *CommercialAirConditioner) updateCapabilities(res *QueryResponse) {
	ac.minTargetTemperature = res.TargetTemperatureMin
	ac.maxTargetTemperature = res.TargetTemperatureMax

	ac.capabilities.Set(CapabilityHumidity, res.SupportsHumidity)

	// Build list of supported operation modes
	ac.supportedOpModes = make([]OperationalMode, 0)
	validOpModes := map[byte]bool{
		byte(OperationalModeFan):  true,
		byte(OperationalModeCool): true,
		byte(OperationalModeHeat): true,
		byte(OperationalModeAuto): true,
		byte(OperationalModeDry):  true,
	}
	for _, mode := range res.SupportedOpModes {
		if validOpModes[mode] {
			ac.supportedOpModes = append(ac.supportedOpModes, OperationalModeFromValue(mode))
		}
	}

	// Build list of supported fan speeds
	if res.SupportsFanSpeed {
		ac.supportedFanSpeeds = FanSpeedList()
	} else {
		ac.supportedFanSpeeds = []FanSpeed{FanSpeedAuto} // TODO??
	}

	// Build list of supported swing modes
	swingModes := []SwingMode{SwingModeOff}
	if res.SupportsHorzSwingAngle {
		swingModes = append(swingModes, SwingModeHorizontal)
	}
	if res.SupportsVertSwingAngle {
		swingModes = append(swingModes, SwingModeVertical)
	}
	if res.SupportsHorzSwingAngle && res.SupportsVertSwingAngle {
		swingModes = append(swingModes, SwingModeBoth)
	}
	ac.supportedSwingModes = swingModes

	// If device can swing it can control the angle
	ac.capabilities.Set(CapabilitySwingHorizontalAngle,
		msmart.Contains(ac.supportedSwingModes, SwingModeHorizontal))
	ac.capabilities.Set(CapabilitySwingVerticalAngle,
		msmart.Contains(ac.supportedSwingModes, SwingModeVertical))

	ac.capabilities.Set(CapabilityECO, res.SupportsEco)
	ac.capabilities.Set(CapabilitySilent, res.SupportsSilent)
	ac.capabilities.Set(CapabilitySleep, res.SupportsSleep)

	// Build list of supported purifier modes
	purifierModes := []PurifierMode{PurifierModeOff}
	if res.SupportsPurifier {
		purifierModes = append(purifierModes, PurifierModeOn)
	}
	if res.SupportsPurifierAuto {
		purifierModes = append(purifierModes, PurifierModeAuto)
	}
	ac.supportedPurifierModes = purifierModes

	// Build list of supported aux heating modes
	ac.supportedAuxModes = make([]AuxHeatMode, 0)
	validAuxModes := map[byte]bool{
		byte(AuxHeatModeAuto): true,
		byte(AuxHeatModeOn):   true,
		byte(AuxHeatModeOff):  true,
	}
	for _, mode := range res.SupportedAuxModes {
		if validAuxModes[mode] {
			ac.supportedAuxModes = append(ac.supportedAuxModes, AuxHeatModeFromValue(mode))
		}
	}
}

// sendCommandsGetResponses sends a list of commands and return all valid responses.
// In Python: async def _send_commands_get_responses(self, commands: Union[Command, list[Command]]) -> list[Response]
func (ac *CommercialAirConditioner) sendCommandsGetResponses(ctx context.Context, commands []interface{}) ([]interface{}, error) {
	var responses [][]byte

	for _, cmd := range commands {
		var data []byte
		switch c := cmd.(type) {
		case *QueryCommand:
			data = c.ToBytes()
		case *ControlCommand:
			data = c.ToBytes()
		default:
			continue
		}

		// Send command using the device's LAN connection
		resp, err := ac.Device.SendBytes(data)
		if err != nil {
			logger.Error("Failed to send command", "error", err)
			continue
		}
		responses = append(responses, resp...)
	}

	// Device is online if any response received
	ac.Device.SetOnline(len(responses) > 0)

	var validResponses []interface{}
	for _, data := range responses {
		// Construct response from data
		response, err := ConstructResponse(data)
		if err != nil {
			logger.Error("Failed to construct response", "error", err)
			continue
		}

		validResponses = append(validResponses, response)
	}

	// Device is supported if we can process any response
	if ac.GetOnline() && len(validResponses) > 0 {
		ac.Device.SetSupported(true)
	}

	return validResponses, nil
}

// GetCapabilities fetches the device capabilities.
// In Python: async def get_capabilities(self) -> None
func (ac *CommercialAirConditioner) GetCapabilities(ctx context.Context) error {
	// Capabilities are part of query response
	cmd := NewQueryCommand()
	responses, err := ac.sendCommandsGetResponses(ctx, []interface{}{cmd})
	if err != nil {
		return err
	}

	if len(responses) == 0 {
		logger.Error("Failed to query capabilities from device", "id", ac.GetID())
		return fmt.Errorf("failed to query capabilities from device %d", ac.GetID())
	}

	response, ok := responses[0].(*QueryResponse)
	if !ok {
		logger.Error("Unexpected response from device", "id", ac.GetID())
		return fmt.Errorf("unexpected response from device %d", ac.GetID())
	}

	// Get capabilities from query response
	logger.Debug("Parsing capabilities from query response payload from device",
		"id", ac.GetID(), "response", response)
	response.ParseCapabilities()

	// Update device capabilities
	ac.updateCapabilities(response)

	return nil
}

// Refresh refreshes the local copy of the device state by sending a GetState command.
// In Python: async def refresh(self) -> None
func (ac *CommercialAirConditioner) Refresh(ctx context.Context) error {
	// Always request state updates
	commands := []interface{}{NewQueryCommand()}

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

// Apply applies the local state to the device.
// In Python: async def apply(self) -> None
func (ac *CommercialAirConditioner) Apply(ctx context.Context) error {
	// Check if nothing to apply
	if len(ac.updatedControls) == 0 {
		return nil
	}

	// Warn if trying to apply unsupported modes
	if _, ok := ac.updatedControls[ControlIdMode]; ok {
		if !msmart.Contains(ac.supportedOpModes, ac.operationalMode) {
			logger.Warn("Device is not capable of operational mode",
				"id", ac.GetID(), "mode", ac.operationalMode)
		}
	}

	if _, ok := ac.updatedControls[ControlIdFanSpeed]; ok {
		if !msmart.Contains(ac.supportedFanSpeeds, ac.fanSpeed) {
			logger.Warn("Device is not capable of fan speed",
				"id", ac.GetID(), "speed", ac.fanSpeed)
		}
	}

	if _, ok := ac.updatedControls[ControlIdEco]; ok {
		if ac.eco && !ac.capabilities.Has(CapabilityECO) {
			logger.Warn("Device is not capable of eco preset", "id", ac.GetID())
		}
	}

	if _, ok := ac.updatedControls[ControlIdSilent]; ok {
		if ac.silent && !ac.capabilities.Has(CapabilitySilent) {
			logger.Warn("Device is not capable of silent preset", "id", ac.GetID())
		}
	}

	if _, ok := ac.updatedControls[ControlIdSleep]; ok {
		if ac.sleep && !ac.capabilities.Has(CapabilitySleep) {
			logger.Warn("Device is not capable of sleep preset", "id", ac.GetID())
		}
	}

	if _, ok := ac.updatedControls[ControlIdPurifier]; ok {
		if ac.purifier != PurifierModeOff && !msmart.Contains(ac.supportedPurifierModes, ac.purifier) {
			logger.Warn("Device is not capable of purifier mode",
				"id", ac.GetID(), "mode", ac.purifier)
		}
	}

	if _, ok := ac.updatedControls[ControlIdAuxMode]; ok {
		if ac.auxMode != AuxHeatModeOff && !msmart.Contains(ac.supportedAuxModes, ac.auxMode) {
			logger.Warn("Device is not capable of aux mode",
				"id", ac.GetID(), "mode", ac.auxMode)
		}
	}

	// Get current state of updated controls
	controls := make(map[ControlId]interface{})
	for k := range ac.updatedControls {
		switch k {
		case ControlIdPower:
			controls[k] = ac.powerState
		case ControlIdTargetTemperature:
			controls[k] = ac.targetTemperature
		case ControlIdTemperatureUnit:
			controls[k] = ac.fahrenheit
		case ControlIdTargetHumidity:
			controls[k] = ac.targetHumidity
		case ControlIdMode:
			controls[k] = ac.operationalMode
		case ControlIdFanSpeed:
			controls[k] = ac.fanSpeed
		case ControlIdHorzSwingAngle:
			controls[k] = ac.horizontalSwingAngle
		case ControlIdVertSwingAngle:
			controls[k] = ac.verticalSwingAngle
		case ControlIdEco:
			controls[k] = ac.eco
		case ControlIdSilent:
			controls[k] = false // In Python: lambda s: False
		case ControlIdSleep:
			controls[k] = false // In Python: lambda s: False
		case ControlIdPurifier:
			controls[k] = ac.purifier
		case ControlIdAuxMode:
			controls[k] = ac.auxMode
		}
	}

	// If powering off device, only send the power control
	var commands []interface{}
	if power, ok := controls[ControlIdPower]; ok && !power.(bool) {
		if len(controls) > 1 {
			// Log dropped controls
			dropped := make(map[ControlId]interface{})
			for k, v := range controls {
				if k != ControlIdPower {
					dropped[k] = v
				}
			}
			logger.Warn("Device powering off. Dropped additional control updates",
				"id", ac.GetID(), "dropped", dropped)
		}
		commands = []interface{}{NewControlCommand(map[ControlId]interface{}{ControlIdPower: false})}
	} else {
		commands = []interface{}{NewControlCommand(controls)}
	}

	// Process any state responses from the device
	responses, err := ac.sendCommandsGetResponses(ctx, commands)
	if err != nil {
		return err
	}

	for _, response := range responses {
		ac.updateState(response)
	}

	// Clear control
	ac.updatedControls = make(map[ControlId]bool)

	return nil
}

// Property getters and setters

// GetPowerState returns the power state.
// In Python: @property def power_state(self) -> Optional[bool]
func (ac *CommercialAirConditioner) GetPowerState() bool {
	return ac.powerState
}

// SetPowerState sets the power state.
// In Python: @power_state.setter def power_state(self, state: bool)
func (ac *CommercialAirConditioner) SetPowerState(state bool) {
	ac.powerState = state
	ac.updatedControls[ControlIdPower] = true
}

// GetMinTargetTemperature returns the minimum target temperature.
// In Python: @property def min_target_temperature(self) -> float
func (ac *CommercialAirConditioner) GetMinTargetTemperature() float64 {
	return ac.minTargetTemperature
}

// GetMaxTargetTemperature returns the maximum target temperature.
// In Python: @property def max_target_temperature(self) -> float
func (ac *CommercialAirConditioner) GetMaxTargetTemperature() float64 {
	return ac.maxTargetTemperature
}

// GetTargetTemperature returns the target temperature.
// In Python: @property def target_temperature(self) -> Optional[float]
func (ac *CommercialAirConditioner) GetTargetTemperature() float64 {
	return ac.targetTemperature
}

// SetTargetTemperature sets the target temperature.
// In Python: @target_temperature.setter def target_temperature(self, temperature_celsius: float)
func (ac *CommercialAirConditioner) SetTargetTemperature(temperatureCelsius float64) {
	ac.targetTemperature = temperatureCelsius
	ac.updatedControls[ControlIdTargetTemperature] = true
}

// GetIndoorTemperature returns the indoor temperature.
// In Python: @property def indoor_temperature(self) -> Optional[float]
func (ac *CommercialAirConditioner) GetIndoorTemperature() *float64 {
	return ac.indoorTemperature
}

// GetOutdoorTemperature returns the outdoor temperature.
// In Python: @property def outdoor_temperature(self) -> Optional[float]
func (ac *CommercialAirConditioner) GetOutdoorTemperature() *float64 {
	return ac.outdoorTemperature
}

// GetFahrenheit returns whether the temperature unit is Fahrenheit.
// In Python: @property def fahrenheit(self) -> Optional[bool]
func (ac *CommercialAirConditioner) GetFahrenheit() bool {
	return ac.fahrenheit
}

// SetFahrenheit sets whether the temperature unit is Fahrenheit.
// In Python: @fahrenheit.setter def fahrenheit(self, enabled: bool)
func (ac *CommercialAirConditioner) SetFahrenheit(enabled bool) {
	ac.fahrenheit = enabled
	ac.updatedControls[ControlIdTemperatureUnit] = true
}

// GetSupportsHumidity returns whether the device supports humidity control.
// In Python: @property def supports_humidity(self) -> bool
func (ac *CommercialAirConditioner) GetSupportsHumidity() bool {
	return ac.capabilities.Has(CapabilityHumidity)
}

// GetTargetHumidity returns the target humidity.
// In Python: @property def target_humidity(self) -> Optional[int]
func (ac *CommercialAirConditioner) GetTargetHumidity() byte {
	return ac.targetHumidity
}

// SetTargetHumidity sets the target humidity.
// In Python: @target_humidity.setter def target_humidity(self, humidity: int)
func (ac *CommercialAirConditioner) SetTargetHumidity(humidity byte) {
	ac.targetHumidity = humidity
	ac.updatedControls[ControlIdTargetHumidity] = true
}

// GetIndoorHumidity returns the indoor humidity.
// In Python: @property def indoor_humidity(self) -> Optional[int]
func (ac *CommercialAirConditioner) GetIndoorHumidity() *byte {
	return ac.indoorHumidity
}

// GetSupportedOperationModes returns the supported operation modes.
// In Python: @property def supported_operation_modes(self) -> list[OperationalMode]
func (ac *CommercialAirConditioner) GetSupportedOperationModes() []OperationalMode {
	return ac.supportedOpModes
}

// GetOperationalMode returns the operational mode.
// In Python: @property def operational_mode(self) -> OperationalMode
func (ac *CommercialAirConditioner) GetOperationalMode() OperationalMode {
	return ac.operationalMode
}

// SetOperationalMode sets the operational mode.
// In Python: @operational_mode.setter def operational_mode(self, mode: OperationalMode)
func (ac *CommercialAirConditioner) SetOperationalMode(mode OperationalMode) {
	ac.operationalMode = mode
	ac.updatedControls[ControlIdMode] = true
}

// GetSupportedFanSpeeds returns the supported fan speeds.
// In Python: @property def supported_fan_speeds(self) -> list[FanSpeed]
func (ac *CommercialAirConditioner) GetSupportedFanSpeeds() []FanSpeed {
	return ac.supportedFanSpeeds
}

// GetFanSpeed returns the fan speed.
// In Python: @property def fan_speed(self) -> FanSpeed | int
func (ac *CommercialAirConditioner) GetFanSpeed() FanSpeed {
	return ac.fanSpeed
}

// SetFanSpeed sets the fan speed.
// In Python: @fan_speed.setter def fan_speed(self, speed: FanSpeed | int | float)
func (ac *CommercialAirConditioner) SetFanSpeed(speed interface{}) {
	switch s := speed.(type) {
	case FanSpeed:
		ac.fanSpeed = s
	case int:
		ac.fanSpeed = FanSpeed(s)
	case float64:
		ac.fanSpeed = FanSpeed(int(s))
	case byte:
		ac.fanSpeed = FanSpeed(s)
	}
	ac.updatedControls[ControlIdFanSpeed] = true
}

// GetSupportedSwingModes returns the supported swing modes.
// In Python: @property def supported_swing_modes(self) -> list[SwingMode]
func (ac *CommercialAirConditioner) GetSupportedSwingModes() []SwingMode {
	return ac.supportedSwingModes
}

// GetSwingMode returns the swing mode.
// In Python: @property def swing_mode(self) -> SwingMode
func (ac *CommercialAirConditioner) GetSwingMode() SwingMode {
	swingMode := SwingModeOff

	if ac.horizontalSwingAngle == SwingAngleAuto {
		swingMode |= SwingModeHorizontal
	}

	if ac.verticalSwingAngle == SwingAngleAuto {
		swingMode |= SwingModeVertical
	}

	return swingMode
}

// SetSwingMode sets the swing mode.
// In Python: @swing_mode.setter def swing_mode(self, mode: SwingMode)
func (ac *CommercialAirConditioner) SetSwingMode(mode SwingMode) {
	// Helper function to get angle
	getAngle := func(swing, enum SwingMode, state SwingAngle) *SwingAngle {
		if swing&enum != 0 {
			angle := SwingAngleAuto
			return &angle
		} else if state == SwingAngleAuto {
			angle := SwingAngleDefault
			return &angle
		}
		return nil
	}

	// Enable swing on correct axes
	if horzAngle := getAngle(mode, SwingModeHorizontal, ac.horizontalSwingAngle); horzAngle != nil {
		ac.horizontalSwingAngle = *horzAngle
		ac.updatedControls[ControlIdHorzSwingAngle] = true
	}

	if vertAngle := getAngle(mode, SwingModeVertical, ac.verticalSwingAngle); vertAngle != nil {
		ac.verticalSwingAngle = *vertAngle
		ac.updatedControls[ControlIdVertSwingAngle] = true
	}
}

// GetSupportsHorizontalSwingAngle returns whether the device supports horizontal swing angle.
// In Python: @property def supports_horizontal_swing_angle(self) -> bool
func (ac *CommercialAirConditioner) GetSupportsHorizontalSwingAngle() bool {
	return ac.capabilities.Has(CapabilitySwingHorizontalAngle)
}

// GetHorizontalSwingAngle returns the horizontal swing angle.
// In Python: @property def horizontal_swing_angle(self) -> SwingAngle
func (ac *CommercialAirConditioner) GetHorizontalSwingAngle() SwingAngle {
	return ac.horizontalSwingAngle
}

// SetHorizontalSwingAngle sets the horizontal swing angle.
// In Python: @horizontal_swing_angle.setter def horizontal_swing_angle(self, angle: SwingAngle)
func (ac *CommercialAirConditioner) SetHorizontalSwingAngle(angle SwingAngle) {
	ac.horizontalSwingAngle = angle
	ac.updatedControls[ControlIdHorzSwingAngle] = true
}

// GetSupportsVerticalSwingAngle returns whether the device supports vertical swing angle.
// In Python: @property def supports_vertical_swing_angle(self) -> bool
func (ac *CommercialAirConditioner) GetSupportsVerticalSwingAngle() bool {
	return ac.capabilities.Has(CapabilitySwingVerticalAngle)
}

// GetVerticalSwingAngle returns the vertical swing angle.
// In Python: @property def vertical_swing_angle(self) -> SwingAngle
func (ac *CommercialAirConditioner) GetVerticalSwingAngle() SwingAngle {
	return ac.verticalSwingAngle
}

// SetVerticalSwingAngle sets the vertical swing angle.
// In Python: @vertical_swing_angle.setter def vertical_swing_angle(self, angle: SwingAngle)
func (ac *CommercialAirConditioner) SetVerticalSwingAngle(angle SwingAngle) {
	ac.verticalSwingAngle = angle
	ac.updatedControls[ControlIdVertSwingAngle] = true
}

// GetSupportsEco returns whether the device supports eco mode.
// In Python: @property def supports_eco(self) -> bool
func (ac *CommercialAirConditioner) GetSupportsEco() bool {
	return ac.capabilities.Has(CapabilityECO)
}

// GetEco returns the eco mode state.
// In Python: @property def eco(self) -> Optional[bool]
func (ac *CommercialAirConditioner) GetEco() bool {
	return ac.eco
}

// SetEco sets the eco mode state.
// In Python: @eco.setter def eco(self, enabled: bool)
func (ac *CommercialAirConditioner) SetEco(enabled bool) {
	ac.eco = enabled
	ac.updatedControls[ControlIdEco] = true
}

// GetSupportsSilent returns whether the device supports silent mode.
// In Python: @property def supports_silent(self) -> bool
func (ac *CommercialAirConditioner) GetSupportsSilent() bool {
	return ac.capabilities.Has(CapabilitySilent)
}

// GetSilent returns the silent mode state.
// In Python: @property def silent(self) -> Optional[bool]
func (ac *CommercialAirConditioner) GetSilent() bool {
	return ac.silent
}

// SetSilent sets the silent mode state.
// In Python: @silent.setter def silent(self, enabled: bool)
func (ac *CommercialAirConditioner) SetSilent(enabled bool) {
	ac.silent = enabled
	ac.updatedControls[ControlIdSilent] = true
}

// GetSupportsSleep returns whether the device supports sleep mode.
// In Python: @property def supports_sleep(self) -> bool
func (ac *CommercialAirConditioner) GetSupportsSleep() bool {
	return ac.capabilities.Has(CapabilitySleep)
}

// GetSleep returns the sleep mode state.
// In Python: @property def sleep(self) -> Optional[bool]
func (ac *CommercialAirConditioner) GetSleep() bool {
	return ac.sleep
}

// SetSleep sets the sleep mode state.
// In Python: @sleep.setter def sleep(self, enabled: bool)
func (ac *CommercialAirConditioner) SetSleep(enabled bool) {
	ac.sleep = enabled
	ac.updatedControls[ControlIdSleep] = true
}

// GetSupportedPurifierModes returns the supported purifier modes.
// In Python: @property def supported_purifier_modes(self) -> list[PurifierMode]
func (ac *CommercialAirConditioner) GetSupportedPurifierModes() []PurifierMode {
	return ac.supportedPurifierModes
}

// GetPurifier returns the purifier mode.
// In Python: @property def purifier(self) -> PurifierMode
func (ac *CommercialAirConditioner) GetPurifier() PurifierMode {
	return ac.purifier
}

// SetPurifier sets the purifier mode.
// In Python: @purifier.setter def purifier(self, mode: PurifierMode)
func (ac *CommercialAirConditioner) SetPurifier(mode PurifierMode) {
	ac.purifier = mode
	ac.updatedControls[ControlIdPurifier] = true
}

// GetSupportedAuxModes returns the supported aux heat modes.
// In Python: @property def supported_aux_modes(self) -> list[AuxHeatMode]
func (ac *CommercialAirConditioner) GetSupportedAuxModes() []AuxHeatMode {
	return ac.supportedAuxModes
}

// GetAuxMode returns the aux heat mode.
// In Python: @property def aux_mode(self) -> AuxHeatMode
func (ac *CommercialAirConditioner) GetAuxMode() AuxHeatMode {
	return ac.auxMode
}

// SetAuxMode sets the aux heat mode.
// In Python: @aux_mode.setter def aux_mode(self, mode: AuxHeatMode)
func (ac *CommercialAirConditioner) SetAuxMode(mode AuxHeatMode) {
	ac.auxMode = mode
	ac.updatedControls[ControlIdAuxMode] = true
}

// ToDict returns the device as a dictionary.
// In Python: def to_dict(self) -> dict
func (ac *CommercialAirConditioner) ToDict() map[string]interface{} {
	result := ac.Device.ToDict()

	result["power"] = ac.GetPowerState()
	result["target_temperature"] = ac.GetTargetTemperature()
	result["indoor_temperature"] = ac.GetIndoorTemperature()
	result["outdoor_temperature"] = ac.GetOutdoorTemperature()
	result["fahrenheit"] = ac.GetFahrenheit()
	result["target_humidity"] = ac.GetTargetHumidity()
	result["indoor_humidity"] = ac.GetIndoorHumidity()
	result["mode"] = ac.GetOperationalMode()
	result["fan_speed"] = ac.GetFanSpeed()
	result["swing_mode"] = ac.GetSwingMode()
	result["horizontal_swing_angle"] = ac.GetHorizontalSwingAngle()
	result["vertical_swing_angle"] = ac.GetVerticalSwingAngle()
	result["eco"] = ac.GetEco()
	result["silent"] = ac.GetSilent()
	result["sleep"] = ac.GetSleep()
	result["purifier"] = ac.GetPurifier()
	result["aux_mode"] = ac.GetAuxMode()
	// result["display"] = ac.GetDisplay()

	return result
}

// CapabilitiesDict returns the device capabilities as a dictionary.
// In Python: def capabilities_dict(self) -> dict
func (ac *CommercialAirConditioner) CapabilitiesDict() map[string]interface{} {
	// Convert capability flags to list
	var caps []string
	if ac.capabilities.Has(CapabilityECO) {
		caps = append(caps, "ECO")
	}
	if ac.capabilities.Has(CapabilitySilent) {
		caps = append(caps, "SILENT")
	}
	if ac.capabilities.Has(CapabilitySleep) {
		caps = append(caps, "SLEEP")
	}
	if ac.capabilities.Has(CapabilitySwingHorizontalAngle) {
		caps = append(caps, "SWING_HORIZONTAL_ANGLE")
	}
	if ac.capabilities.Has(CapabilitySwingVerticalAngle) {
		caps = append(caps, "SWING_VERTICAL_ANGLE")
	}
	if ac.capabilities.Has(CapabilityHumidity) {
		caps = append(caps, "HUMIDITY")
	}
	if ac.capabilities.Has(CapabilityPurifier) {
		caps = append(caps, "PURIFIER")
	}

	return map[string]interface{}{
		"min_target_temperature":   ac.GetMinTargetTemperature(),
		"max_target_temperature":   ac.GetMaxTargetTemperature(),
		"supported_modes":          ac.GetSupportedOperationModes(),
		"supported_swing_modes":    ac.GetSupportedSwingModes(),
		"supported_fan_speeds":     ac.GetSupportedFanSpeeds(),
		"supported_aux_modes":      ac.GetSupportedAuxModes(),
		"supported_purifier_modes": ac.GetSupportedPurifierModes(),
		"additional_capabilities":  caps,
	}
}

