// Package ac provides air conditioner device commands and responses.
// This is a translation from msmart-ng Python library.
package ac

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"

	msmart "github.com/RelicOfTesla/midea-msmart/msmart"
)

// InvalidResponseException represents an error for invalid responses
type InvalidResponseException struct {
	message string
}

// Error implements the error interface
func (e *InvalidResponseException) Error() string {
	return e.message
}

// NewInvalidResponseException creates a new InvalidResponseException
func NewInvalidResponseException(message string) *InvalidResponseException {
	return &InvalidResponseException{message: message}
}

// ResponseId represents response ID enum
type ResponseId byte

const (
	ResponseIdPropertiesAck ResponseId = 0xB0 // In response to property commands
	ResponseIdProperties    ResponseId = 0xB1
	ResponseIdCapabilities  ResponseId = 0xB5
	ResponseIdState         ResponseId = 0xC0
	ResponseIdGroupData     ResponseId = 0xC1
)

// CapabilityId represents capability ID enum
type CapabilityId uint16

const (
	CapabilityIdSwingUdAngle              CapabilityId = 0x0009
	CapabilityIdSwingLrAngle              CapabilityId = 0x000A
	CapabilityIdBreezeless                CapabilityId = 0x0018 // AKA "No Wind Sense"
	CapabilityIdSmartEye                  CapabilityId = 0x0030
	CapabilityIdWindOnMe                  CapabilityId = 0x0032
	CapabilityIdWindOffMe                 CapabilityId = 0x0033
	CapabilityIdSelfClean                 CapabilityId = 0x0039 // AKA Active Clean
	CapabilityId_Unknown                  CapabilityId = 0x0040 // Unknown ID from various logs
	CapabilityIdBreezeAway                CapabilityId = 0x0042 // AKA "Prevent Straight Wind"
	CapabilityIdBreezeControl             CapabilityId = 0x0043 // AKA "FA No Wind Sense"
	CapabilityIdRateSelect                CapabilityId = 0x0048
	CapabilityIdFreshAir                  CapabilityId = 0x004B
	CapabilityIdParentControl             CapabilityId = 0x0051 // ??
	CapabilityIdPreventStraightWindSelect CapabilityId = 0x0058 // ??
	CapabilityIdCascade                   CapabilityId = 0x0059 // AKA "Wind Around"
	CapabilityIdJetCool                   CapabilityId = 0x0067 // ??
	CapabilityIdICheck                    CapabilityId = 0x0091 // ??
	CapabilityIdEmergentHeatWind          CapabilityId = 0x0093 // ??
	CapabilityIdHeatPtcWind               CapabilityId = 0x0094 // ??
	CapabilityIdCVP                       CapabilityId = 0x0098 // ??
	CapabilityIdOutSilent                 CapabilityId = 0x00CD // Portasplit outdoor silent mode
	CapabilityIdPresetIeco                CapabilityId = 0x00E3
	CapabilityIdFanSpeedControl           CapabilityId = 0x0210
	CapabilityIdPresetEco                 CapabilityId = 0x0212
	CapabilityIdPresetFreezeProtection    CapabilityId = 0x0213
	CapabilityIdModes                     CapabilityId = 0x0214
	CapabilityIdSwingModes                CapabilityId = 0x0215
	CapabilityIdEnergy                    CapabilityId = 0x0216 // AKA electricity
	CapabilityIdFilterRemind              CapabilityId = 0x0217
	CapabilityIdAuxElectricHeat           CapabilityId = 0x0219 // AKA PTC
	CapabilityIdPresetTurbo               CapabilityId = 0x021A
	CapabilityIdFilterCheck               CapabilityId = 0x0221
	CapabilityIdAnion                     CapabilityId = 0x021E
	CapabilityIdHumidity                  CapabilityId = 0x021F
	CapabilityIdFahrenheit                CapabilityId = 0x0222
	CapabilityIdDisplayControl            CapabilityId = 0x0224
	CapabilityIdTemperatures              CapabilityId = 0x0225
	CapabilityIdBuzzer                    CapabilityId = 0x022C // TODO Reference refers to this as "sound". Is this different then buzzer?
	CapabilityIdMainHorizontalGuideStrip  CapabilityId = 0x0230 // ??
	CapabilityIdSupHorizontalGuideStrip   CapabilityId = 0x0231 // ??
	CapabilityIdTwinsMachine              CapabilityId = 0x0232 // ??
	CapabilityIdGuideStripType            CapabilityId = 0x0233 // ??
	CapabilityIdBodyCheck                 CapabilityId = 0x0234 // ??
)

// PropertyId represents property ID enum
type PropertyId uint16

const (
	PropertyIdSwingUdAngle   PropertyId = 0x0009
	PropertyIdSwingLrAngle   PropertyId = 0x000A
	PropertyIdIndoorHumidity PropertyId = 0x0015 // TODO Reference refers to a potential bug with this
	PropertyIdBreezeless     PropertyId = 0x0018 // AKA "No Wind Sense"
	PropertyIdBuzzer         PropertyId = 0x001A
	PropertyIdSelfClean      PropertyId = 0x0039
	PropertyIdBreezeAway     PropertyId = 0x0042 // AKA "Prevent Straight Wind"
	PropertyIdBreezeControl  PropertyId = 0x0043 // AKA "FA No Wind Sense"
	PropertyIdRateSelect     PropertyId = 0x0048
	PropertyIdFreshAir       PropertyId = 0x004B
	PropertyIdCascade        PropertyId = 0x0059 // AKA "Wind Around"
	PropertyIdJetCool        PropertyId = 0x0067 // AKA "Flash Cool"
	PropertyIdOutSilent      PropertyId = 0x00CD // Portasplit outdoor silent mode
	PropertyIdIECO           PropertyId = 0x00E3
	PropertyIdAnion          PropertyId = 0x021E
)

// StateId represents state field enum
type StateId uint16

const (
	// WriteOnly fields (only in SetStateCommand, cannot be read)
	StateIdBeepOn StateId = iota + 1
	StateIdForceAuxHeat

	// ReadWrite fields (in both SetStateCommand and StateResponse)
	StateIdTargetTemperature
	StateIdTargetHumidity
	StateIdOperationalMode
	StateIdFanSpeed
	StateIdSwingMode
	StateIdEco
	StateIdTurbo
	StateIdFreezeProtection
	StateIdSleep
	StateIdFahrenheitUnit
	StateIdFollowMe
	StateIdPurifier
	StateIdAuxMode // Derived from AuxHeat/IndependentAuxHeat
	StateIdPowerOn

	// ReadOnly fields (only in StateResponse, cannot be written)
	StateIdIndoorTemperature
	StateIdOutdoorTemperature
	StateIdIndoorHumidity
	StateIdErrorCode
	StateIdFilterAlert
	StateIdOutdoorFanSpeed
	StateIdDefrostActive
	StateIdDisplayOn
)

// IsReadable returns true if the StateId can be read via GetState
// WriteOnly fields return false
func (s StateId) IsReadable() bool {
	switch s {
	case StateIdBeepOn, StateIdForceAuxHeat:
		return false // WriteOnly
	default:
		return true // ReadWrite or ReadOnly
	}
}

// IsWritable returns true if the StateId can be written via SetState
// ReadOnly fields return false
func (s StateId) IsWritable() bool {
	switch s {
	case StateIdIndoorTemperature, StateIdOutdoorTemperature, StateIdIndoorHumidity,
		StateIdErrorCode, StateIdFilterAlert, StateIdOutdoorFanSpeed,
		StateIdDefrostActive, StateIdDisplayOn:
		return false // ReadOnly
	default:
		return true // ReadWrite or WriteOnly
	}
}

// IsSupported checks if a property ID is supported/tested.
func (p PropertyId) IsSupported() bool {
	switch p {
	case PropertyIdBreezeAway,
		PropertyIdBreezeControl,
		PropertyIdBreezeless,
		PropertyIdBuzzer,
		PropertyIdCascade,
		PropertyIdIECO,
		PropertyIdJetCool,
		PropertyIdOutSilent,
		PropertyIdRateSelect,
		PropertyIdSelfClean,
		PropertyIdSwingLrAngle,
		PropertyIdSwingUdAngle:
		return true
	default:
		return false
	}
}

// Decode decodes raw property data into a convenient form.
// Returns the decoded value as interface{} and error if not supported.
func (p PropertyId) Decode(data []byte) (interface{}, error) {
	if !p.IsSupported() {
		return nil, fmt.Errorf("%v decode is not supported", p)
	}

	switch p {
	case PropertyIdBreezeless, PropertyIdJetCool, PropertyIdSelfClean:
		return data[0] != 0, nil
	case PropertyIdBreezeAway:
		return data[0] == 2, nil
	case PropertyIdBuzzer:
		return nil, nil // Don't decode buzzer
	case PropertyIdCascade:
		// data[0] - wind_around, data[1] - wind_around_ud
		if data[0] != 0 {
			return data[1], nil
		}
		return byte(0), nil
	case PropertyIdIECO:
		// data[0] - ieco_number, data[1] - ieco_switch
		return data[1] != 0, nil
	case PropertyIdOutSilent:
		return data[0] == 3, nil
	default:
		return data[0], nil
	}
}

// Encode encodes property into raw form.
func (p PropertyId) Encode(args ...interface{}) ([]byte, error) {
	if !p.IsSupported() {
		return nil, fmt.Errorf("%v encode is not supported", p)
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("%v encode requires at least one argument", p)
	}

	switch p {
	case PropertyIdBreezeAway:
		// Accept bool or int (like Python)
		if b, ok := args[0].(bool); ok {
			if b {
				return []byte{2}, nil
			}
			return []byte{1}, nil
		}
		if i, ok := args[0].(int); ok {
			if i != 0 {
				return []byte{2}, nil
			}
			return []byte{1}, nil
		}
		return nil, fmt.Errorf("%v encode requires bool or int argument, got %T", p, args[0])

	case PropertyIdCascade:
		// data[0] - wind_around, data[1] - wind_around_ud
		// Python: bytes([1 if args[0] else 0, args[0]])
		if b, ok := args[0].(bool); ok {
			if b {
				return []byte{1, 1}, nil
			}
			return []byte{0, 0}, nil
		}
		if i, ok := args[0].(int); ok {
			if i != 0 {
				return []byte{1, byte(i)}, nil
			}
			return []byte{0, 0}, nil
		}
		return nil, fmt.Errorf("%v encode requires bool or int argument, got %T", p, args[0])

	case PropertyIdIECO:
		// ieco_frame, ieco_number, ieco_switch, ...
		// Python: bytes([0, 1, args[0]]) + bytes(10)
		if b, ok := args[0].(bool); ok {
			if b {
				return append([]byte{0, 1, 1}, make([]byte, 10)...), nil
			}
			return append([]byte{0, 1, 0}, make([]byte, 10)...), nil
		}
		if i, ok := args[0].(int); ok {
			return append([]byte{0, 1, byte(i)}, make([]byte, 10)...), nil
		}
		return nil, fmt.Errorf("%v encode requires bool or int argument, got %T", p, args[0])

	case PropertyIdOutSilent:
		if b, ok := args[0].(bool); ok {
			if b {
				return []byte{3}, nil
			}
			return []byte{0}, nil
		}
		if i, ok := args[0].(int); ok {
			if i != 0 {
				return []byte{3}, nil
			}
			return []byte{0}, nil
		}
		return nil, fmt.Errorf("%v encode requires bool or int argument, got %T", p, args[0])

	default:
		// Python: bytes(args[0:1]) - converts first arg to a single byte
		if b, ok := args[0].(byte); ok {
			return []byte{b}, nil
		}
		if b, ok := args[0].(int); ok {
			return []byte{byte(b)}, nil
		}
		if b, ok := args[0].(bool); ok {
			if b {
				return []byte{1}, nil
			}
			return []byte{0}, nil
		}
		return nil, fmt.Errorf("%v encode requires byte, int, or bool argument, got %T", p, args[0])
	}
}

// TemperatureType represents temperature type enum
type TemperatureType byte

const (
	TemperatureTypeUnknown TemperatureType = 0
	TemperatureTypeIndoor  TemperatureType = 0x2
	TemperatureTypeOutdoor TemperatureType = 0x3
)

// Command is the base class for AC commands.
type Command struct {
	*msmart.Frame
	controlSource byte
	messageID     int
}

// NewCommand creates a new Command instance.
func NewCommand(frameType msmart.FrameType) *Command {
	return &Command{
		Frame:         msmart.NewFrame(msmart.DeviceTypeAirConditioner, frameType),
		controlSource: 0x2, // App control
		messageID:     0,
	}
}

// ToBytes converts command to bytes with payload.
// This is the Go equivalent of Python tobytes method.
func (c *Command) ToBytes(data []byte) []byte {
	// Append message ID to payload
	payload := append(data, byte(c.nextMessageID()))

	// Append CRC
	crc := msmart.CalculateCRC8(payload)
	payload = append(payload, crc)

	return c.Frame.ToBytes(payload)
}

// nextMessageID generates next message ID.
func (c *Command) nextMessageID() int {
	c.messageID++
	return c.messageID & 0xFF
}

// GetCapabilitiesCommand is a command to query capabilities of the device.
type GetCapabilitiesCommand struct {
	*Command
	additional bool
}

// NewGetCapabilitiesCommand creates a new GetCapabilitiesCommand instance.
func NewGetCapabilitiesCommand(additional bool) *GetCapabilitiesCommand {
	return &GetCapabilitiesCommand{
		Command:    NewCommand(msmart.FrameTypeQuery),
		additional: additional,
	}
}

// ToBytes converts GetCapabilitiesCommand to bytes.
func (c *GetCapabilitiesCommand) ToBytes() []byte {
	var payload []byte
	if !c.additional {
		// Get capabilities
		payload = []byte{0xB5, 0x01, 0x00}
	} else {
		// Get more capabilities
		payload = []byte{0xB5, 0x01, 0x01, 0x1}
	}
	return c.Command.ToBytes(payload)
}

// GetStateCommand is a command to query basic state of the device.
type GetStateCommand struct {
	*Command
	TemperatureType TemperatureType
}

// NewGetStateCommand creates a new GetStateCommand instance.
func NewGetStateCommand() *GetStateCommand {
	return &GetStateCommand{
		Command:         NewCommand(msmart.FrameTypeQuery),
		TemperatureType: TemperatureTypeIndoor,
	}
}

// ToBytes converts GetStateCommand to bytes.
func (c *GetStateCommand) ToBytes() []byte {
	payload := []byte{
		// Get state
		0x41,
		// Unknown
		0x81, 0x00, 0xFF, 0x03, 0xFF, 0x00,
		// Temperature request
		byte(c.TemperatureType),
		// Unknown
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		// Unknown
		0x03,
	}
	return c.Command.ToBytes(payload)
}

// GetEnergyUsageCommand is a command to query energy usage from device.
type GetEnergyUsageCommand struct {
	*Command
}

// NewGetEnergyUsageCommand creates a new GetEnergyUsageCommand instance.
func NewGetEnergyUsageCommand() *GetEnergyUsageCommand {
	return &GetEnergyUsageCommand{
		Command: NewCommand(msmart.FrameTypeQuery),
	}
}

// ToBytes converts GetEnergyUsageCommand to bytes.
func (c *GetEnergyUsageCommand) ToBytes() []byte {
	payload := make([]byte, 20)

	payload[0] = 0x41
	payload[1] = 0x21
	payload[2] = 0x01
	payload[3] = 0x44

	return c.Command.ToBytes(payload)
}

// GetGroup5Command is a command to query group 5 data from device.
type GetGroup5Command struct {
	*Command
}

// NewGetGroup5Command creates a new GetGroup5Command instance.
func NewGetGroup5Command() *GetGroup5Command {
	return &GetGroup5Command{
		Command: NewCommand(msmart.FrameTypeQuery),
	}
}

// ToBytes converts GetGroup5Command to bytes.
func (c *GetGroup5Command) ToBytes() []byte {
	payload := make([]byte, 20)

	payload[0] = 0x41
	payload[1] = 0x21
	payload[2] = 0x01
	payload[3] = 0x45

	return c.Command.ToBytes(payload)
}

// SetStateCommand is a command to set basic state of the device.
type SetStateCommand struct {
	*Command
	BeepOn             bool
	PowerOn            bool
	TargetTemperature  float64
	OperationalMode    byte
	FanSpeed           byte
	Eco                bool
	SwingMode          byte
	Turbo              bool
	Fahrenheit         bool
	Sleep              bool
	FreezeProtection   bool
	FollowMe           bool
	Purifier           bool
	TargetHumidity     byte
	AuxHeat            bool
	ForceAuxHeat       bool
	IndependentAuxHeat bool
}

// NewSetStateCommand creates a new SetStateCommand instance.
func NewSetStateCommand() *SetStateCommand {
	return &SetStateCommand{
		Command:            NewCommand(msmart.FrameTypeControl),
		BeepOn:             true,
		PowerOn:            false,
		TargetTemperature:  25.0,
		OperationalMode:    0,
		FanSpeed:           0,
		Eco:                true,
		SwingMode:          0,
		Turbo:              false,
		Fahrenheit:         true,
		Sleep:              false,
		FreezeProtection:   false,
		FollowMe:           false,
		Purifier:           false,
		TargetHumidity:     40,
		AuxHeat:            false,
		ForceAuxHeat:       false,
		IndependentAuxHeat: false,
	}
}

// FillSetStateCommandFromMap fills a SetStateCommand from a state map.
// This is useful for applying state changes to an existing command.
// Only WriteOnly and ReadWrite StateIds are processed (ReadOnly are ignored).
// Returns the number of fields that were set.
func FillSetStateCommandFromMap(cmd *SetStateCommand, stateMap map[StateId]any) (int, error) {
	count := 0

	for key, value := range stateMap {
		// Skip ReadOnly fields (they cannot be written)
		if !key.IsWritable() {
			continue
		}

		var err error
		var set bool
		switch key {
		case StateIdBeepOn:
			cmd.BeepOn, err = convertToBool(value, "BeepOn")
			set = true
		case StateIdPowerOn:
			cmd.PowerOn, err = convertToBool(value, "PowerOn")
			set = true
		case StateIdTargetTemperature:
			cmd.TargetTemperature, err = convertToFloat64(value, "TargetTemperature")
			set = true
		case StateIdTargetHumidity:
			var val float64
			val, err = convertToFloat64(value, "TargetHumidity")
			if err == nil {
				cmd.TargetHumidity = byte(val)
				set = true
			}
		case StateIdOperationalMode:
			var val float64
			val, err = convertToFloat64(value, "OperationalMode")
			if err == nil {
				cmd.OperationalMode = byte(val)
				set = true
			}
		case StateIdFanSpeed:
			var val float64
			val, err = convertToFloat64(value, "FanSpeed")
			if err == nil {
				cmd.FanSpeed = byte(val)
				set = true
			}
		case StateIdSwingMode:
			var val float64
			val, err = convertToFloat64(value, "SwingMode")
			if err == nil {
				cmd.SwingMode = byte(val)
				set = true
			}
		case StateIdEco:
			cmd.Eco, err = convertToBool(value, "Eco")
			set = true
		case StateIdTurbo:
			cmd.Turbo, err = convertToBool(value, "Turbo")
			set = true
		case StateIdFreezeProtection:
			cmd.FreezeProtection, err = convertToBool(value, "FreezeProtection")
			set = true
		case StateIdSleep:
			cmd.Sleep, err = convertToBool(value, "Sleep")
			set = true
		case StateIdFahrenheitUnit:
			cmd.Fahrenheit, err = convertToBool(value, "FahrenheitUnit")
			set = true
		case StateIdFollowMe:
			cmd.FollowMe, err = convertToBool(value, "FollowMe")
			set = true
		case StateIdPurifier:
			cmd.Purifier, err = convertToBool(value, "Purifier")
			set = true
		case StateIdAuxMode:
			// AuxMode is derived from AuxHeat/IndependentAuxHeat
			// AuxHeatModeOff = 0, AuxHeatModeAuxHeat = 1, AuxHeatModeAuxOnly = 2
			var val float64
			val, err = convertToFloat64(value, "AuxMode")
			if err == nil {
				switch byte(val) {
				case 1: // AuxHeatModeAuxHeat
					cmd.AuxHeat = true
					cmd.IndependentAuxHeat = false
				case 2: // AuxHeatModeAuxOnly
					cmd.AuxHeat = false
					cmd.IndependentAuxHeat = true
				default: // AuxHeatModeOff
					cmd.AuxHeat = false
					cmd.IndependentAuxHeat = false
				}
				set = true
			}
		case StateIdForceAuxHeat:
			cmd.ForceAuxHeat, err = convertToBool(value, "ForceAuxHeat")
			set = true
		}

		if err != nil {
			return count, fmt.Errorf("state %d: %w", key, err)
		}
		if set {
			count++
		}
	}

	return count, nil
}

// toBool converts a value to bool with error context
func convertToBool(value any, fieldName string) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		// Non-zero values are true
		return reflect.ValueOf(value).Int() != 0, nil
	case nil:
		return false, nil
	default:
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return rv.Int() != 0, nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return rv.Uint() != 0, nil
		case reflect.Float32, reflect.Float64:
			return rv.Float() != 0, nil
		default:
			return false, fmt.Errorf("%s: cannot convert %T to bool", fieldName, value)
		}
	}
}

// toFloat64 converts a value to float64 with error context
func convertToFloat64(value any, fieldName string) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case nil:
		return 0, nil
	default:
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return float64(rv.Int()), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return float64(rv.Uint()), nil
		case reflect.Float32, reflect.Float64:
			return rv.Float(), nil
		default:
			return 0, fmt.Errorf("%s: cannot convert %T to float64", fieldName, value)
		}
	}
}

// ToBytes converts SetStateCommand to bytes.
func (c *SetStateCommand) ToBytes() []byte {
	// Build beep and power status bytes
	var beep byte
	if c.BeepOn {
		beep = 0x40
	}
	var power byte
	if c.PowerOn {
		power = 0x1
	}

	// Get integer and fraction components of target temp
	integralTemp, fractionalTemp := math.Modf(c.TargetTemperature)
	integralTempInt := int(integralTemp)

	var temperature byte
	var temperatureAlt byte

	if integralTempInt >= 17 && integralTempInt <= 30 {
		// Use primary method
		temperature = byte((integralTempInt - 16) & 0xF)
		temperatureAlt = 0
	} else {
		// Out of range, use alternate method
		// TODO additional range possible according to Lua code
		temperature = 0
		temperatureAlt = byte((integralTempInt - 12) & 0x1F)
	}

	// Set half degree bit
	if fractionalTemp > 0 {
		temperature |= 0x10
	}

	mode := byte((c.OperationalMode & 0x7) << 5)

	// Build swing mode byte
	swingMode := byte(0x30 | (c.SwingMode & 0x3F))

	// Build eco mode, purifier, and aux heat byte
	var eco byte
	if c.Eco {
		eco = 0x80
	}
	var purifier byte
	if c.Purifier {
		purifier = 0x20
	}
	var auxHeat byte
	if c.AuxHeat {
		auxHeat = 0x08
	}
	var forceAuxHeat byte
	if c.ForceAuxHeat {
		forceAuxHeat = 0x10
	}

	// Build sleep, turbo and fahrenheit byte
	var sleep byte
	if c.Sleep {
		sleep = 0x01
	}
	var turbo byte
	if c.Turbo {
		turbo = 0x02
	}
	var fahrenheit byte
	if c.Fahrenheit {
		fahrenheit = 0x04
	}

	// Build alternate turbo byte
	var turboAlt byte
	if c.Turbo {
		turboAlt = 0x20
	}
	var followMe byte
	if c.FollowMe {
		followMe = 0x80
	}

	// Build target humidity byte
	humidity := c.TargetHumidity & 0x7F

	// Build freeze protection byte
	var freezeProtect byte
	if c.FreezeProtection {
		freezeProtect = 0x80
	}

	// Build independent aux heat
	var independentAuxHeat byte
	if c.IndependentAuxHeat {
		independentAuxHeat = 0x08
	}

	payload := []byte{
		// Set state
		0x40,
		// Beep and power state
		c.controlSource | beep | power,
	// Temperature and operational mode
		temperature | mode,
		// Fan speed
		c.FanSpeed,
		// Timer
		0x7F, 0x7F, 0x00,
		// Swing mode
		swingMode,
		// Follow me and alternate turbo mode
		followMe | turboAlt,
		// ECO mode, purifier/anion, and aux heat
		eco | purifier | forceAuxHeat | auxHeat,
		// Sleep mode, turbo mode and fahrenheit
		sleep | turbo | fahrenheit,
		// Unknown
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00,
		// Alternate temperature
		temperatureAlt,
		// Target humidity
		humidity,
		// Unknown
		0x00,
		// Frost/freeze protection
		freezeProtect,
		// Independent aux heat
		independentAuxHeat,
		// Unknown
		0x00,
	}

	return c.Command.ToBytes(payload)
}

// ToggleDisplayCommand is a command to toggle the LED display of the device.
type ToggleDisplayCommand struct {
	*Command
	BeepOn bool
}

// NewToggleDisplayCommand creates a new ToggleDisplayCommand instance.
func NewToggleDisplayCommand() *ToggleDisplayCommand {
	return &ToggleDisplayCommand{
		Command: NewCommand(msmart.FrameTypeQuery), // For whatever reason, toggle display uses a request type...
		BeepOn:  true,
	}
}

// ToBytes converts ToggleDisplayCommand to bytes.
func (c *ToggleDisplayCommand) ToBytes() []byte {
	// Set beep bit
	var beep byte
	if c.BeepOn {
		beep = 0x40
	}

	payload := []byte{
		// Get state
		0x41,
		// Beep and other flags
		c.controlSource | beep,
		// Unknown
		0x00, 0xFF, 0x02,
		0x00, 0x02, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
	}

	return c.Command.ToBytes(payload)
}

// GetPropertiesCommand is a command to query specific properties from the device.
type GetPropertiesCommand struct {
	*Command
	properties []PropertyId
}

// NewGetPropertiesCommand creates a new GetPropertiesCommand instance.
func NewGetPropertiesCommand(props []PropertyId) *GetPropertiesCommand {
	return &GetPropertiesCommand{
		Command:    NewCommand(msmart.FrameTypeQuery),
		properties: props,
	}
}

// ToBytes converts GetPropertiesCommand to bytes.
func (c *GetPropertiesCommand) ToBytes() []byte {
	var payload []byte
	payload = append(payload, 0xB1) // Property request
	payload = append(payload, byte(len(c.properties)))

	for _, prop := range c.properties {
		propBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(propBytes, uint16(prop))
		payload = append(payload, propBytes...)
	}

	return c.Command.ToBytes(payload)
}

// SetPropertiesCommand is a command to set specific properties of the device.
type SetPropertiesCommand struct {
	*Command
	properties map[PropertyId]interface{}
}

// NewSetPropertiesCommand creates a new SetPropertiesCommand instance.
func NewSetPropertiesCommand(props map[PropertyId]interface{}) *SetPropertiesCommand {
	return &SetPropertiesCommand{
		Command:    NewCommand(msmart.FrameTypeControl),
		properties: props,
	}
}

// ToBytes converts SetPropertiesCommand to bytes.
func (c *SetPropertiesCommand) ToBytes() []byte {
	var payload []byte
	payload = append(payload, 0xB0) // Property request
	payload = append(payload, byte(len(c.properties)))

	for prop, value := range c.properties {
		propBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(propBytes, uint16(prop))
		payload = append(payload, propBytes...)

		// Encode property value to bytes
		valueBytes, _ := prop.Encode(value)

		payload = append(payload, byte(len(valueBytes)))
		payload = append(payload, valueBytes...)
	}

	return c.Command.ToBytes(payload)
}

// ResponseInterface is the interface for all response types
type ResponseInterface interface {
	ID() byte
	Payload() []byte
	String() string
}

// Response is the base class for AC responses.
type Response struct {
	id      byte
	payload []byte
}

// NewResponse creates a new Response instance.
func NewResponse(payload []byte) *Response {
	return &Response{
		id:      payload[0],
		payload: payload,
	}
}

// String returns the hex string of the payload.
func (r *Response) String() string {
	return fmt.Sprintf("%x", r.payload)
}

// ID returns the response ID.
func (r *Response) ID() byte {
	return r.id
}

// Payload returns the response payload.
func (r *Response) Payload() []byte {
	return r.payload
}

// Validate validates a response by checking the frame checksum and payload CRC.
func ValidateResponse(payload []byte) error {
	// Some devices use a CRC others seem to use a 2nd checksum
	payloadCrc := msmart.CalculateCRC8(payload[0 : len(payload)-1])
	payloadChecksum := msmart.Checksum(payload[0 : len(payload)-1])

	if payloadCrc != payload[len(payload)-1] && payloadChecksum != payload[len(payload)-1] {
		return NewInvalidResponseException(
			fmt.Sprintf("Payload '%x' failed CRC and checksum. Received: 0x%X, Expected: 0x%X or 0x%X.",
				payload, payload[len(payload)-1], payloadCrc, payloadChecksum))
	}

	return nil
}

// ConstructResponse constructs a response object from raw data.
func ConstructResponse(frame []byte) (ResponseInterface, error) {
	// Validate the frame
	err := msmart.Validate(frame, msmart.DeviceTypeAirConditioner)
	if err != nil {
		return nil, err
	}

	// Default to base class
	var responseClass ResponseInterface

	// Fetch the appropriate response class from the ID
	frameType := frame[9]
	responseId := frame[10]

	switch ResponseId(responseId) {
	case ResponseIdState:
		responseClass = NewStateResponse(frame[10 : len(frame)-2])
	case ResponseIdCapabilities:
		if msmart.FrameType(frameType) == msmart.FrameTypeQuery {
			// Some devices have unsolicited "capabilities" responses with a frame type of 0x5
			responseClass = NewCapabilitiesResponse(frame[10 : len(frame)-2])
		} else {
			responseClass = NewResponse(frame[10 : len(frame)-2])
		}
	case ResponseIdProperties, ResponseIdPropertiesAck:
		responseClass = NewPropertiesResponse(frame[10 : len(frame)-2])
	case ResponseIdGroupData:
		// Response type depends on an additional "group" byte
		group := frame[13] & 0xF
		if group == 4 {
			responseClass = NewEnergyUsageResponse(frame[10 : len(frame)-2])
		} else if group == 5 {
			responseClass = NewGroup5Response(frame[10 : len(frame)-2])
		} else {
			responseClass = NewResponse(frame[10 : len(frame)-2])
		}
	default:
		responseClass = NewResponse(frame[10 : len(frame)-2])
	}

	// Validate the payload CRC
	// ...except for properties which certain devices send invalid CRCs
	if _, ok := responseClass.(*PropertiesResponse); !ok {
		err = ValidateResponse(frame[10 : len(frame)-1])
		if err != nil {
			return nil, err
		}
	}

	return responseClass, nil
}

// CapabilitiesResponse is a response to capabilities query.
type CapabilitiesResponse struct {
	*Response
	capabilities           map[string]interface{}
	additionalCapabilities bool
}

// NewCapabilitiesResponse creates a new CapabilitiesResponse instance.
func NewCapabilitiesResponse(payload []byte) *CapabilitiesResponse {
	r := &CapabilitiesResponse{
		Response:     NewResponse(payload),
		capabilities: make(map[string]interface{}),
	}
	r.parseCapabilities(payload)
	return r
}

// RawCapabilities returns the raw capabilities map.
func (r *CapabilitiesResponse) RawCapabilities() map[string]interface{} {
	return r.capabilities
}

// capabilityDecoder represents a named decoder for capability values
type capabilityDecoder struct {
	name string
	read func(int) bool
}

// parseCapabilities parses the capabilities from the payload.
func (r *CapabilitiesResponse) parseCapabilities(payload []byte) {
	// Clear existing capabilities
	r.capabilities = make(map[string]interface{})

	// Define a local function to parse capability values
	getValue := func(w int) func(int) bool {
		return func(v int) bool { return v == w }
	}

	// Create a map of capability ID to decoders
	capabilityReaders := map[CapabilityId]interface{}{
		CapabilityIdAnion:           capabilityDecoder{"anion", getValue(1)},
		CapabilityIdAuxElectricHeat: capabilityDecoder{"aux_electric_heat", getValue(1)},
		CapabilityIdBreezeAway:      capabilityDecoder{"breeze_away", getValue(1)},
		CapabilityIdBreezeControl:   capabilityDecoder{"breeze_control", getValue(1)},
		CapabilityIdBreezeless:      capabilityDecoder{"breezeless", getValue(1)},
		CapabilityIdBuzzer:          capabilityDecoder{"buzzer", getValue(1)},
		CapabilityIdCascade:         capabilityDecoder{"cascade", getValue(1)},
		CapabilityIdDisplayControl:  capabilityDecoder{"display_control", func(v int) bool { return v == 1 || v == 2 || v == 100 }},
		CapabilityIdEnergy: []capabilityDecoder{
			{"energy_stats", func(v int) bool { return v == 2 || v == 3 || v == 4 || v == 5 }},
			{"energy_setting", func(v int) bool { return v == 3 || v == 5 }},
			{"energy_bcd", func(v int) bool { return v == 2 || v == 3 }},
		},
		CapabilityIdFahrenheit: capabilityDecoder{"fahrenheit", getValue(0)},
		CapabilityIdFanSpeedControl: []capabilityDecoder{
			{"fan_silent", getValue(6)},
			{"fan_low", func(v int) bool { return v == 3 || v == 4 || v == 5 || v == 6 || v == 7 }},
			{"fan_medium", func(v int) bool { return v == 5 || v == 6 || v == 7 }},
			{"fan_high", func(v int) bool { return v == 3 || v == 4 || v == 5 || v == 6 || v == 7 }},
			{"fan_auto", func(v int) bool { return v == 4 || v == 5 || v == 6 }},
			{"fan_custom", getValue(1)},
		},
		CapabilityIdFilterRemind: []capabilityDecoder{
			{"filter_notice", func(v int) bool { return v == 1 || v == 2 || v == 4 }},
			{"filter_clean", func(v int) bool { return v == 3 || v == 4 }},
		},
		CapabilityIdHumidity: []capabilityDecoder{
			{"humidity_auto_set", func(v int) bool { return v == 1 || v == 2 }},
			{"humidity_manual_set", func(v int) bool { return v == 2 || v == 3 }},
		},
		CapabilityIdJetCool: capabilityDecoder{"jet_cool", getValue(1)},
		CapabilityIdModes: []capabilityDecoder{
			{"heat_mode", func(v int) bool {
				return v == 1 || v == 2 || v == 4 || v == 6 || v == 7 || v == 9 || v == 10 || v == 11 || v == 12 || v == 13
			}},
			{"cool_mode", func(v int) bool { return v != 2 && v != 10 && v != 12 }},
			{"dry_mode", func(v int) bool { return v == 0 || v == 1 || v == 5 || v == 6 || v == 9 || v == 11 || v == 13 }},
			{"auto_mode", func(v int) bool { return v == 0 || v == 1 || v == 2 || v == 7 || v == 8 || v == 9 || v == 13 }},
			{"aux_heat_mode", func(v int) bool { return v == 9 }},                             // Heat & Aux
			{"aux_mode", func(v int) bool { return v == 9 || v == 10 || v == 11 || v == 13 }}, // Aux only
		},
		CapabilityIdOutSilent:              capabilityDecoder{"out_silent", func(v int) bool { return v == 1 || v == 3 }},
		CapabilityIdPresetEco:              capabilityDecoder{"eco", func(v int) bool { return v == 1 || v == 2 }},
		CapabilityIdPresetFreezeProtection: capabilityDecoder{"freeze_protection", getValue(1)},
		CapabilityIdPresetIeco:             capabilityDecoder{"ieco", getValue(1)},
		CapabilityIdPresetTurbo: []capabilityDecoder{
			{"turbo_heat", func(v int) bool { return v == 1 || v == 3 }},
			{"turbo_cool", func(v int) bool { return v < 2 }},
		},
		CapabilityIdRateSelect: []capabilityDecoder{
			{"rate_select_2_level", getValue(1)},                                  // Gear
			{"rate_select_5_level", func(v int) bool { return v == 2 || v == 3 }}, // Genmode and Gear5
		},
		CapabilityIdSelfClean:    capabilityDecoder{"self_clean", getValue(1)},
		CapabilityIdSmartEye:     capabilityDecoder{"smart_eye", getValue(1)},
		CapabilityIdSwingLrAngle: capabilityDecoder{"swing_horizontal_angle", getValue(1)},
		CapabilityIdSwingUdAngle: capabilityDecoder{"swing_vertical_angle", getValue(1)},
		CapabilityIdSwingModes: []capabilityDecoder{
			{"swing_horizontal", func(v int) bool { return v == 1 || v == 3 }},
			{"swing_vertical", func(v int) bool { return v < 2 }},
		},
		// CapabilityIdTemperatures too complex to be handled here
		CapabilityIdWindOffMe: capabilityDecoder{"wind_off_me", getValue(1)},
		CapabilityIdWindOnMe:  capabilityDecoder{"wind_on_me", getValue(1)},
		// CapabilityId_Unknown is a special case
	}

	count := int(payload[1])
	caps := payload[2:]

	// Loop through each capability
	for i := 0; i < count; i++ {
		// Stop if out of data
		if len(caps) < 3 {
			break
		}

		// Skip empty capabilities
		size := int(caps[2])
		if size == 0 {
			caps = caps[3:]
			continue
		}

		// Unpack 16 bit ID
		rawId := binary.LittleEndian.Uint16(caps[0:2])

		// Convert ID to enumerate type
		capabilityId := CapabilityId(rawId)

		// Fetch first cap value
		value := int(caps[3])

		// Apply predefined capability reader if it exists
		if reader, ok := capabilityReaders[capabilityId]; ok {
			// Local function to apply a reader
			apply := func(d capabilityDecoder, v int) {
				r.capabilities[d.name] = d.read(v)
			}

			switch r := reader.(type) {
			case []capabilityDecoder:
				// Apply each reader in the list
				for _, dec := range r {
					apply(dec, value)
				}
			case capabilityDecoder:
				// Apply the single reader
				apply(r, value)
			}
		} else if capabilityId == CapabilityIdTemperatures {
			// Skip if capability size is too small
			if size < 6 {
				caps = caps[3+size:]
				continue
			}

			r.capabilities["cool_min_temperature"] = float64(caps[3]) * 0.5
			r.capabilities["cool_max_temperature"] = float64(caps[4]) * 0.5
			r.capabilities["auto_min_temperature"] = float64(caps[5]) * 0.5
			r.capabilities["auto_max_temperature"] = float64(caps[6]) * 0.5
			r.capabilities["heat_min_temperature"] = float64(caps[7]) * 0.5
			r.capabilities["heat_max_temperature"] = float64(caps[8]) * 0.5

			// TODO The else of this condition is commented out in reference code
			if size > 6 {
				r.capabilities["decimals"] = caps[9] != 0
			} else {
				r.capabilities["decimals"] = caps[2] != 0
			}
		} else if capabilityId == CapabilityId_Unknown {
			// Suppress warnings from unknown capability
			// Ignored
		} else {
			// Unsupported capability
		}

		// Advanced to next capability
		caps = caps[3+size:]
	}

	// Check if there are additional capabilities
	if len(caps) > 1 {
		r.additionalCapabilities = caps[len(caps)-2] != 0
	}
}

// getFanSpeed gets fan speed capability.
func (r *CapabilitiesResponse) getFanSpeed(speed string) bool {
	// If any fan_ capability was received, check against them
	for k := range r.capabilities {
		if len(k) >= 4 && k[0:4] == "fan_" {
			// Assume that a fan capable of custom speeds is capable of any speed
			if v, ok := r.capabilities[fmt.Sprintf("fan_%s", speed)].(bool); ok {
				return v
			}
			if v, ok := r.capabilities["fan_custom"].(bool); ok {
				return v
			}
			return false
		}
	}

	// Otherwise return a default set for devices that don't send the capability
	return speed == "low" || speed == "medium" || speed == "high" || speed == "auto"
}

// Merge merges other capabilities into this one.
func (r *CapabilitiesResponse) Merge(other *CapabilitiesResponse) {
	for k, v := range other.capabilities {
		r.capabilities[k] = v
	}
	r.additionalCapabilities = other.additionalCapabilities
}

// AdditionalCapabilities returns whether there are additional capabilities.
func (r *CapabilitiesResponse) AdditionalCapabilities() bool {
	return r.additionalCapabilities
}

// Anion returns the anion capability.
func (r *CapabilitiesResponse) Anion() bool {
	if v, ok := r.capabilities["anion"].(bool); ok {
		return v
	}
	return false
}

// TODO rethink these properties for fan speed, operation mode and swing mode
// Surely there's a better way than define props for each possible cap

// FanSilent returns the fan silent capability.
func (r *CapabilitiesResponse) FanSilent() bool {
	return r.getFanSpeed("silent")
}

// FanLow returns the fan low capability.
func (r *CapabilitiesResponse) FanLow() bool {
	return r.getFanSpeed("low")
}

// FanMedium returns the fan medium capability.
func (r *CapabilitiesResponse) FanMedium() bool {
	return r.getFanSpeed("medium")
}

// FanHigh returns the fan high capability.
func (r *CapabilitiesResponse) FanHigh() bool {
	return r.getFanSpeed("high")
}

// FanAuto returns the fan auto capability.
func (r *CapabilitiesResponse) FanAuto() bool {
	return r.getFanSpeed("auto")
}

// FanCustom returns the fan custom capability.
func (r *CapabilitiesResponse) FanCustom() bool {
	if v, ok := r.capabilities["fan_custom"].(bool); ok {
		return v
	}
	return false
}

// BreezeAway returns the breeze away capability.
func (r *CapabilitiesResponse) BreezeAway() bool {
	if v, ok := r.capabilities["breeze_away"].(bool); ok {
		return v
	}
	return false
}

// BreezeControl returns the breeze control capability.
func (r *CapabilitiesResponse) BreezeControl() bool {
	if v, ok := r.capabilities["breeze_control"].(bool); ok {
		return v
	}
	return false
}

// Breezeless returns the breezeless capability.
func (r *CapabilitiesResponse) Breezeless() bool {
	if v, ok := r.capabilities["breezeless"].(bool); ok {
		return v
	}
	return false
}

// Cascade returns the cascade capability.
func (r *CapabilitiesResponse) Cascade() bool {
	if v, ok := r.capabilities["cascade"].(bool); ok {
		return v
	}
	return false
}

// SwingHorizontalAngle returns the swing horizontal angle capability.
func (r *CapabilitiesResponse) SwingHorizontalAngle() bool {
	if v, ok := r.capabilities["swing_horizontal_angle"].(bool); ok {
		return v
	}
	return false
}

// SwingVerticalAngle returns the swing vertical angle capability.
func (r *CapabilitiesResponse) SwingVerticalAngle() bool {
	if v, ok := r.capabilities["swing_vertical_angle"].(bool); ok {
		return v
	}
	return false
}

// SwingHorizontal returns the swing horizontal capability.
func (r *CapabilitiesResponse) SwingHorizontal() bool {
	if v, ok := r.capabilities["swing_horizontal"].(bool); ok {
		return v
	}
	return false
}

// SwingVertical returns the swing vertical capability.
func (r *CapabilitiesResponse) SwingVertical() bool {
	if v, ok := r.capabilities["swing_vertical"].(bool); ok {
		return v
	}
	return false
}

// SwingBoth returns whether both swing modes are available.
func (r *CapabilitiesResponse) SwingBoth() bool {
	return r.SwingVertical() && r.SwingHorizontal()
}

// DryMode returns the dry mode capability.
func (r *CapabilitiesResponse) DryMode() bool {
	if v, ok := r.capabilities["dry_mode"].(bool); ok {
		return v
	}
	return false
}

// CoolMode returns the cool mode capability.
func (r *CapabilitiesResponse) CoolMode() bool {
	if v, ok := r.capabilities["cool_mode"].(bool); ok {
		return v
	}
	return false
}

// HeatMode returns the heat mode capability.
func (r *CapabilitiesResponse) HeatMode() bool {
	if v, ok := r.capabilities["heat_mode"].(bool); ok {
		return v
	}
	return false
}

// AutoMode returns the auto mode capability.
func (r *CapabilitiesResponse) AutoMode() bool {
	if v, ok := r.capabilities["auto_mode"].(bool); ok {
		return v
	}
	return false
}

// AuxHeatMode returns the aux heat mode capability.
func (r *CapabilitiesResponse) AuxHeatMode() bool {
	if v, ok := r.capabilities["aux_heat_mode"].(bool); ok {
		return v
	}
	return false
}

// AuxMode returns the aux mode capability.
func (r *CapabilitiesResponse) AuxMode() bool {
	if v, ok := r.capabilities["aux_mode"].(bool); ok {
		return v
	}
	return false
}

// AuxElectricHeat returns the aux electric heat capability.
// TODO How does electric aux heat differ from aux mode?
func (r *CapabilitiesResponse) AuxElectricHeat() bool {
	if v, ok := r.capabilities["aux_electric_heat"].(bool); ok {
		return v
	}
	return false
}

// Eco returns the eco capability.
func (r *CapabilitiesResponse) Eco() bool {
	if v, ok := r.capabilities["eco"].(bool); ok {
		return v
	}
	return false
}

// Ieco returns the ieco capability.
func (r *CapabilitiesResponse) Ieco() bool {
	if v, ok := r.capabilities["ieco"].(bool); ok {
		return v
	}
	return false
}

// JetCool returns the jet cool capability.
func (r *CapabilitiesResponse) JetCool() bool {
	if v, ok := r.capabilities["jet_cool"].(bool); ok {
		return v
	}
	return false
}

// Turbo returns the turbo capability.
func (r *CapabilitiesResponse) Turbo() bool {
	heat := false
	cool := false
	if v, ok := r.capabilities["turbo_heat"].(bool); ok {
		heat = v
	}
	if v, ok := r.capabilities["turbo_cool"].(bool); ok {
		cool = v
	}
	return heat || cool
}

// FreezeProtection returns the freeze protection capability.
func (r *CapabilitiesResponse) FreezeProtection() bool {
	if v, ok := r.capabilities["freeze_protection"].(bool); ok {
		return v
	}
	return false
}

// DisplayControl returns the display control capability.
func (r *CapabilitiesResponse) DisplayControl() bool {
	if v, ok := r.capabilities["display_control"].(bool); ok {
		return v
	}
	return false
}

// FilterReminder returns the filter reminder capability.
// TODO unsure of difference between filter_notice and filter_clean
func (r *CapabilitiesResponse) FilterReminder() bool {
	if v, ok := r.capabilities["filter_notice"].(bool); ok {
		return v
	}
	return false
}

// MinTemperature returns the minimum temperature.
func (r *CapabilitiesResponse) MinTemperature() int {
	modes := []string{"cool", "auto", "heat"}
	minTemp := 16
	for _, m := range modes {
		if v, ok := r.capabilities[fmt.Sprintf("%s_min_temperature", m)].(float64); ok {
			temp := int(v)
			if temp < minTemp {
				minTemp = temp
			}
		}
	}
	return minTemp
}

// MaxTemperature returns the maximum temperature.
func (r *CapabilitiesResponse) MaxTemperature() int {
	modes := []string{"cool", "auto", "heat"}
	maxTemp := 30
	for _, m := range modes {
		if v, ok := r.capabilities[fmt.Sprintf("%s_max_temperature", m)].(float64); ok {
			temp := int(v)
			if temp > maxTemp {
				maxTemp = temp
			}
		}
	}
	return maxTemp
}

// EnergyStats returns the energy stats capability.
func (r *CapabilitiesResponse) EnergyStats() bool {
	if v, ok := r.capabilities["energy_stats"].(bool); ok {
		return v
	}
	return false
}

// Humidity returns the humidity capability.
// TODO Unsure the difference between these two
func (r *CapabilitiesResponse) Humidity() bool {
	autoSet := false
	manualSet := false
	if v, ok := r.capabilities["humidity_auto_set"].(bool); ok {
		autoSet = v
	}
	if v, ok := r.capabilities["humidity_manual_set"].(bool); ok {
		manualSet = v
	}
	return autoSet || manualSet
}

// TargetHumidity returns the target humidity capability.
func (r *CapabilitiesResponse) TargetHumidity() bool {
	if v, ok := r.capabilities["humidity_manual_set"].(bool); ok {
		return v
	}
	return false
}

// SelfClean returns the self clean capability.
func (r *CapabilitiesResponse) SelfClean() bool {
	if v, ok := r.capabilities["self_clean"].(bool); ok {
		return v
	}
	return false
}

// RateSelectLevels returns the rate select levels.
func (r *CapabilitiesResponse) RateSelectLevels() *int {
	if v, ok := r.capabilities["rate_select_5_level"].(bool); ok && v {
		levels := 5
		return &levels
	}
	if v, ok := r.capabilities["rate_select_2_level"].(bool); ok && v {
		levels := 2
		return &levels
	}
	return nil
}

// OutSilent returns the out silent capability.
func (r *CapabilitiesResponse) OutSilent() bool {
	if v, ok := r.capabilities["out_silent"].(bool); ok {
		return v
	}
	return false
}

// StateResponse is a response to state query.
type StateResponse struct {
	*Response
	PowerOn            *bool
	TargetTemperature  *float64
	OperationalMode    *byte
	FanSpeed           *byte
	SwingMode          *byte
	Turbo              *bool
	Eco                *bool
	Sleep              *bool
	Fahrenheit         *bool
	IndoorTemperature  *float64
	OutdoorTemperature *float64
	FilterAlert        *bool
	DisplayOn          *bool
	FreezeProtection   *bool
	FollowMe           *bool
	Purifier           *bool
	TargetHumidity     *byte
	AuxHeat            *bool
	IndependentAuxHeat *bool
	ErrorCode          *byte
}

// StateKvResponse is a key-value response for state data
// Used to simplify updateState() logic
type StateKvResponse struct {
	Values map[StateId]any
}

// ToKv converts StateResponse to StateKvResponse
// This allows updateState() to handle state data in a unified way
// convertFunc is an optional function to convert values (e.g., for type conversions)
func (r *StateResponse) ToKv(convertFunc func(StateId, any) any) *StateKvResponse {
	kv := &StateKvResponse{
		Values: make(map[StateId]any),
	}

	// Helper to add value with optional conversion
	addValue := func(key StateId, value any) {
		if convertFunc != nil {
			value = convertFunc(key, value)
		}
		kv.Values[key] = value
	}

	// Add non-nil values to the map
	if r.PowerOn != nil {
		addValue(StateIdPowerOn, *r.PowerOn)
	}
	if r.TargetTemperature != nil {
		addValue(StateIdTargetTemperature, *r.TargetTemperature)
	}
	if r.OperationalMode != nil {
		addValue(StateIdOperationalMode, *r.OperationalMode)
	}
	if r.FanSpeed != nil {
		addValue(StateIdFanSpeed, *r.FanSpeed)
	}
	if r.SwingMode != nil {
		addValue(StateIdSwingMode, *r.SwingMode)
	}
	if r.Turbo != nil {
		addValue(StateIdTurbo, *r.Turbo)
	}
	if r.Eco != nil {
		addValue(StateIdEco, *r.Eco)
	}
	if r.Sleep != nil {
		addValue(StateIdSleep, *r.Sleep)
	}
	if r.Fahrenheit != nil {
		addValue(StateIdFahrenheitUnit, *r.Fahrenheit)
	}
	if r.IndoorTemperature != nil {
		addValue(StateIdIndoorTemperature, *r.IndoorTemperature)
	}
	if r.OutdoorTemperature != nil {
		addValue(StateIdOutdoorTemperature, *r.OutdoorTemperature)
	}
	if r.FilterAlert != nil {
		addValue(StateIdFilterAlert, *r.FilterAlert)
	}
	if r.DisplayOn != nil {
		addValue(StateIdDisplayOn, *r.DisplayOn)
	}
	if r.FreezeProtection != nil {
		addValue(StateIdFreezeProtection, *r.FreezeProtection)
	}
	if r.FollowMe != nil {
		addValue(StateIdFollowMe, *r.FollowMe)
	}
	if r.Purifier != nil {
		addValue(StateIdPurifier, *r.Purifier)
	}
	if r.TargetHumidity != nil {
		addValue(StateIdTargetHumidity, int(*r.TargetHumidity))
	}
	// Aux mode logic
	if r.IndependentAuxHeat != nil && *r.IndependentAuxHeat {
		addValue(StateIdAuxMode, AuxHeatModeAuxOnly)
	} else if r.AuxHeat != nil && *r.AuxHeat {
		addValue(StateIdAuxMode, AuxHeatModeAuxHeat)
	} else {
		addValue(StateIdAuxMode, AuxHeatModeOff)
	}
	if r.ErrorCode != nil {
		addValue(StateIdErrorCode, int(*r.ErrorCode))
	}

	return kv
}

// NewStateResponse creates a new StateResponse instance.
func NewStateResponse(payload []byte) *StateResponse {
	r := &StateResponse{
		Response: NewResponse(payload),
	}
	r.parse(payload)
	return r
}

// parseTemperature parses a temperature value from the payload using additional precision bits as needed.
func (r *StateResponse) parseTemperature(data byte, decimals float64, fahrenheit bool) *float64 {
	if data == 0xFF {
		return nil
	}

	// Temperature parsing lifted from https://github.com/dudanov/MideaUART
	temperature := (float64(data) - 50.0) / 2.0

	// In Celsius, use additional precision from decimals if present
	if !fahrenheit && decimals != 0 {
		temp := int(temperature)
		if temperature >= 0 {
			return floatPtr(float64(temp) + decimals)
		}
		return floatPtr(float64(temp) - decimals)
	}

	if decimals >= 0.5 {
		temp := int(temperature)
		if temperature >= 0 {
			return floatPtr(float64(temp) + 0.5)
		}
		return floatPtr(float64(temp) - 0.5)
	}

	return floatPtr(temperature)
}

// parse parses the state response payload.
func (r *StateResponse) parse(payload []byte) {
	// Power on
	powerOn := (payload[1] & 0x1) != 0
	r.PowerOn = &powerOn
	// self.imode_resume = payload[1] & 0x4
	// self.timer_mode = (payload[1] & 0x10) > 0
	// self.appliance_error = (payload[1] & 0x80) > 0

	// Unpack target temp and mode byte
	targetTemp := float64(payload[2]&0xF) + 16.0
	if payload[2]&0x10 != 0 {
		targetTemp += 0.5
	}
	r.TargetTemperature = &targetTemp
	opMode := byte((payload[2] >> 5) & 0x7)
	r.OperationalMode = &opMode

	// Fan speed
	// Fan speed can be auto = 102, or value from 0 - 100
	// On my unit, Low == 40 (LED < 40), Med == 60 (LED < 60), High == 100 (LED < 100)
	fanSpeed := payload[3] & 0x7F
	r.FanSpeed = &fanSpeed

	// on_timer_value = payload[4]
	// on_timer_minutes = payload[6]
	// self.on_timer = {
	//     'status': ((on_timer_value & 0x80) >> 7) > 0,
	//     'hour': (on_timer_value & 0x7c) >> 2,
	//     'minutes': (on_timer_value & 0x3) | ((on_timer_minutes & 0xf0) >> 4)
	// }

	// off_timer_value = payload[5]
	// off_timer_minutes = payload[6]
	// self.off_timer = {
	//     'status': ((off_timer_value & 0x80) >> 7) > 0,
	//     'hour': (off_timer_value & 0x7c) >> 2,
	//     'minutes': (off_timer_value & 0x3) | (off_timer_minutes & 0xf)
	// }

	// Swing mode
	swingMode := payload[7] & 0xF
	r.SwingMode = &swingMode

	// self.cozy_sleep = payload[8] & 0x03
	// self.save = (payload[8] & 0x08) > 0
	// self.low_frequency_fan = (payload[8] & 0x10) > 0
	turbo := (payload[8] & 0x20) != 0
	r.Turbo = &turbo
	indAuxHeat := (payload[8] & 0x40) != 0
	r.IndependentAuxHeat = &indAuxHeat
	followMe := (payload[8] & 0x80) != 0
	r.FollowMe = &followMe

	eco := (payload[9] & 0x10) != 0
	r.Eco = &eco
	purifier := (payload[9] & 0x20) != 0
	r.Purifier = &purifier
	// self.child_sleep = (payload[9] & 0x01) > 0
	// self.exchange_air = (payload[9] & 0x02) > 0
	// self.dry_clean = (payload[9] & 0x04) > 0
	auxHeat := (payload[9] & 0x08) != 0
	r.AuxHeat = &auxHeat
	// self.temp_unit = (payload[9] & 0x80) > 0

	sleep := (payload[10] & 0x1) != 0
	r.Sleep = &sleep
	if (payload[10] & 0x2) != 0 {
		turbo = true
		r.Turbo = &turbo
	}
	fahrenheit := (payload[10] & 0x4) != 0
	r.Fahrenheit = &fahrenheit
	// self.catch_cold = (payload[10] & 0x08) > 0
	// self.night_light = (payload[10] & 0x10) > 0
	// self.peak_elec = (payload[10] & 0x20) > 0
	// self.natural_fan = (payload[10] & 0x40) > 0

	// Decode temperatures using additional precision bits
	fahrenheit = r.Fahrenheit != nil && *r.Fahrenheit
	r.IndoorTemperature = r.parseTemperature(payload[11], float64(payload[15]&0xF)/10.0, fahrenheit)
	r.OutdoorTemperature = r.parseTemperature(payload[12], float64(payload[15]>>4)/10.0, fahrenheit)

	// Decode alternate target temperature
	targetTempAlt := payload[13] & 0x1F
	if targetTempAlt != 0 {
		// TODO additional range possible according to Lua code
		newTargetTemp := float64(targetTempAlt) + 12
		if payload[2]&0x10 != 0 {
			newTargetTemp += 0.5
		}
		r.TargetTemperature = &newTargetTemp
	}

	filterAlert := (payload[13] & 0x20) != 0
	r.FilterAlert = &filterAlert

	displayOn := payload[14] != 0x70
	r.DisplayOn = &displayOn

	errorCode := payload[16]
	r.ErrorCode = &errorCode

	if len(payload) < 20 {
		return
	}

	targetHumidity := payload[19] & 0x7F
	r.TargetHumidity = &targetHumidity

	if len(payload) < 22 {
		return
	}

	freezeProt := (payload[21] & 0x80) != 0
	r.FreezeProtection = &freezeProt
}

// PropertiesResponse is a response to properties query.
type PropertiesResponse struct {
	*Response
	properties map[PropertyId]interface{}
}

// NewPropertiesResponse creates a new PropertiesResponse instance.
func NewPropertiesResponse(payload []byte) *PropertiesResponse {
	r := &PropertiesResponse{
		Response:   NewResponse(payload),
		properties: make(map[PropertyId]interface{}),
	}
	r.parse(payload)
	return r
}

// parse parses the properties response payload.
func (r *PropertiesResponse) parse(payload []byte) {
	// Clear existing properties
	r.properties = make(map[PropertyId]interface{})

	count := int(payload[1])
	props := payload[2:]

	// Loop through each property
	for i := 0; i < count; i++ {
		// Stop if out of data
		if len(props) < 4 {
			break
		}

		// Skip empty properties
		size := int(props[3])
		if size == 0 {
			props = props[4:]
			continue
		}

		// Unpack 16 bit ID
		rawId := binary.LittleEndian.Uint16(props[0:2])

		// Convert ID to enumerate type
		property := PropertyId(rawId)

		// Check execution result and log any errors
		error := props[2] & 0x10
		if error != 0 {
			// Property failed
		}

		// Parse the property
		if !property.IsSupported() {
			// Unsupported property
		} else {
			value, err := property.Decode(props[4:])
			if err == nil && value != nil {
				r.properties[property] = value
			}
		}

		// Advanced to next property
		props = props[4+size:]
	}
}

// GetProperty gets a property value by ID.
func (r *PropertiesResponse) GetProperty(id PropertyId) interface{} {
	return r.properties[id]
}

// EnergyUsageResponse is a response to a GetEnergyUsageCommand.
type EnergyUsageResponse struct {
	*Response
	TotalEnergy         *float64
	CurrentEnergy       *float64
	RealTimePower       *float64
	TotalEnergyBinary   *float64
	CurrentEnergyBinary *float64
	RealTimePowerBinary *float64
}

// NewEnergyUsageResponse creates a new EnergyUsageResponse instance.
func NewEnergyUsageResponse(payload []byte) *EnergyUsageResponse {
	r := &EnergyUsageResponse{
		Response: NewResponse(payload),
	}
	r.parse(payload)
	return r
}

// parse parses the energy usage response payload.
func (r *EnergyUsageResponse) parse(payload []byte) {
	// Response is technically a "group data 4" response
	// and may contain other interesting data

	decodeBcd := func(d byte) int {
		return 10*int(d>>4) + int(d&0xF)
	}

	parseEnergy := func(d []byte) (float64, float64) {
		bcd := float64(10000*decodeBcd(d[0]) +
			100*decodeBcd(d[1]) +
			1*decodeBcd(d[2]) +
			int(0.01*float64(decodeBcd(d[3]))))
		binary := float64((int(d[0])<<24)+(int(d[1])<<16)+(int(d[2])<<8)+int(d[3])) / 10.0
		return bcd, binary
	}

	parsePower := func(d []byte) (float64, float64) {
		bcd := float64(1000*decodeBcd(d[0])+
			10*decodeBcd(d[1])) +
			0.1*float64(decodeBcd(d[2]))
		binary := float64((int(d[0])<<16)+(int(d[1])<<8)+int(d[2])) / 10.0
		return bcd, binary
	}

	// Lua reference decodes real time power field in BCD and binary form
	// JS reference decodes multiple energy/power fields in BCD only.

	// Total energy in bytes 4 - 7
	totalEnergyBcd, totalEnergyBinary := parseEnergy(payload[4:8])

	// JS references decodes bytes 8 - 11 as "total running energy"
	// Older JS does not decode these bytes, and sample payloads contain bogus data

	// Current run energy consumption bytes 12 - 15
	currentEnergyBcd, currentEnergyBinary := parseEnergy(payload[12:16])

	// Real time power usage bytes 16 - 18
	realTimePowerBcd, realTimePowerBinary := parsePower(payload[16:19])

	// Assume energy monitory is valid if at least one stat is non zero
	valid := totalEnergyBcd != 0 || currentEnergyBcd != 0 || realTimePowerBcd != 0

	if valid {
		r.TotalEnergy = &totalEnergyBcd
		r.CurrentEnergy = &currentEnergyBcd
		r.RealTimePower = &realTimePowerBcd
		r.TotalEnergyBinary = &totalEnergyBinary
		r.CurrentEnergyBinary = &currentEnergyBinary
		r.RealTimePowerBinary = &realTimePowerBinary
	}
}

// Group5Response is a group 5 response with humidity, defrost and more.
type Group5Response struct {
	*Response
	Humidity        *byte
	Defrost         *bool
	OutdoorFanSpeed *byte
}

// NewGroup5Response creates a new Group5Response instance.
func NewGroup5Response(payload []byte) *Group5Response {
	r := &Group5Response{
		Response: NewResponse(payload),
	}
	r.parse(payload)
	return r
}

// parse parses the group 5 response payload.
func (r *Group5Response) parse(payload []byte) {
	if payload[4] != 0 {
		humidity := payload[4]
		r.Humidity = &humidity
	}

	outdoorFanSpeed := byte(8 * payload[8])
	r.OutdoorFanSpeed = &outdoorFanSpeed

	defrost := payload[10] != 0
	r.Defrost = &defrost
}

// Helper function to create float pointer
func floatPtr(f float64) *float64 {
	return &f
}
