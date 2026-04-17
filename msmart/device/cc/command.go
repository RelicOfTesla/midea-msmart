// Package cc provides commercial air conditioner device commands and responses.
// This is a translation from msmart-ng Python library.
// Original file: msmart/device/CC/command.py
package cc

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"sync"

	msmart "github.com/RelicOfTesla/midea-msmart/msmart"
)

// Package-level message ID counter (like Python class variable)
// Thread-safe with mutex
var (
	messageID    int
	messageIDMux sync.Mutex
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

// ControlId represents control ID enum
// In Python: class ControlId(IntEnum)
type ControlId uint16

const (
	ControlIdPower             ControlId = 0x0000
	ControlIdTargetTemperature ControlId = 0x0003
	ControlIdTemperatureUnit   ControlId = 0x000C
	ControlIdTargetHumidity    ControlId = 0x000F
	ControlIdMode              ControlId = 0x0012
	ControlIdFanSpeed          ControlId = 0x0015
	ControlIdVertSwingAngle    ControlId = 0x001C
	ControlIdHorzSwingAngle    ControlId = 0x001E
	ControlIdWindSense         ControlId = 0x0020 // Untested
	ControlIdEco               ControlId = 0x0028
	ControlIdSilent            ControlId = 0x002A
	ControlIdSleep             ControlId = 0x002C
	ControlIdSelfClean         ControlId = 0x002E // Untested
	ControlIdPurifier          ControlId = 0x003A
	ControlIdBeep              ControlId = 0x003F
	ControlIdDisplay           ControlId = 0x0040
	ControlIdAuxMode           ControlId = 0x0043 // Untested
)

// IsKnown checks if the ControlId is a known/defined value.
// Returns true if the control ID matches one of the defined constants.
func (c ControlId) IsKnown() bool {
	switch c {
	case ControlIdPower, ControlIdTargetTemperature, ControlIdTemperatureUnit,
		ControlIdTargetHumidity, ControlIdMode, ControlIdFanSpeed,
		ControlIdVertSwingAngle, ControlIdHorzSwingAngle, ControlIdWindSense,
		ControlIdEco, ControlIdSilent, ControlIdSleep, ControlIdSelfClean,
		ControlIdPurifier, ControlIdBeep, ControlIdDisplay, ControlIdAuxMode:
		return true
	default:
		return false
	}
}

// Decode decodes raw control data into a convenient form.
// In Python: def decode(self, data: bytes) -> Any
func (c ControlId) Decode(data []byte) interface{} {
	if c == ControlIdTargetTemperature {
		return float64(data[0])/2.0 - 40
	}
	return data[0]
}

// Encode encodes controls into raw form.
// In Python: def encode(self, *args, **kwargs) -> bytes
func (c ControlId) Encode(args ...interface{}) []byte {
	if c == ControlIdTargetTemperature {
		if len(args) > 0 {
			if f, ok := args[0].(float64); ok {
				return []byte{byte(int(2*f + 80))}
			}
			if i, ok := args[0].(int); ok {
				return []byte{byte(2*i + 80)}
			}
		}
	}
	if len(args) > 0 {
		if b, ok := args[0].(byte); ok {
			return []byte{b}
		}
		if i, ok := args[0].(int); ok {
			return []byte{byte(i)}
		}
		if b, ok := args[0].(bool); ok {
			if b {
				return []byte{1}
			}
			return []byte{0}
		}
	}
	return []byte{}
}

// Command is the base class for CC commands.
// In Python: class Command(Frame)
// Note: messageID is now a package-level variable (like Python class variable)
type Command struct {
	*msmart.Frame
}

// NewCommand creates a new Command instance.
// In Python: def __init__(self, frame_type: FrameType)
func NewCommand(frameType msmart.FrameType) *Command {
	return &Command{
		Frame: msmart.NewFrame(msmart.DeviceTypeCommercialAC, frameType),
	}
}

// nextMessageID generates next message ID (thread-safe).
// In Python: def _next_message_id(self) -> int
// This uses package-level messageID (like Python class variable)
func nextMessageID() int {
	messageIDMux.Lock()
	defer messageIDMux.Unlock()
	messageID++
	return messageID & 0xFF
}

// ToBytes converts command to bytes with payload.
// In Python: def tobytes(self, data: Union[bytes, bytearray] = bytes()) -> bytes
func (c *Command) ToBytes(data []byte) []byte {
	// Append message ID to payload
	// TODO Message ID in reference is just a random value
	msgID := nextMessageID()
	payload := append(data, byte(msgID))

	// Append CRC
	crc := msmart.CalculateCRC8(payload)
	payload = append(payload, crc)

	return c.Frame.ToBytes(payload)
}

// QueryCommand is a command to query state of the device.
// In Python: class QueryCommand(Command)
type QueryCommand struct {
	*Command
}

// NewQueryCommand creates a new QueryCommand instance.
func NewQueryCommand() *QueryCommand {
	return &QueryCommand{
		Command: NewCommand(msmart.FrameTypeQuery),
	}
}

// ToBytes converts QueryCommand to bytes.
// In Python: def tobytes(self) -> bytes
func (c *QueryCommand) ToBytes() []byte {
	// TODO Query format doesn't match plugin but seems to work
	payload := make([]byte, 22)
	payload[0] = 0x01

	return c.Command.ToBytes(payload)
}

// ControlCommand is a command to control state of the device.
// In Python: class ControlCommand(Command)
type ControlCommand struct {
	*Command
	controls map[ControlId]interface{}
}

// NewControlCommand creates a new ControlCommand instance.
// In Python: def __init__(self, controls: Mapping[ControlId, Union[int, float, bool]])
func NewControlCommand(controls map[ControlId]interface{}) *ControlCommand {
	return &ControlCommand{
		Command:  NewCommand(msmart.FrameTypeControl),
		controls: controls,
	}
}

// ToBytes converts ControlCommand to bytes.
// In Python: def tobytes(self) -> bytes
func (c *ControlCommand) ToBytes() []byte {
	payload := []byte{}

	for control, value := range c.controls {
		// Pack 16-bit control ID (big-endian)
		idBytes := make([]byte, 2)
		binary.BigEndian.PutUint16(idBytes, uint16(control))
		payload = append(payload, idBytes...)

		// Encode property value to bytes
		var valueBytes []byte
		switch v := value.(type) {
		case bool:
			if v {
				valueBytes = []byte{1}
			} else {
				valueBytes = []byte{0}
			}
		case int:
			valueBytes = control.Encode(v)
		case float64:
			valueBytes = control.Encode(v)
		case byte:
			valueBytes = []byte{v}
		default:
			valueBytes = control.Encode(v)
		}

		payload = append(payload, byte(len(valueBytes)))
		payload = append(payload, valueBytes...)
		payload = append(payload, 0xFF)
	}

	return c.Command.ToBytes(payload)
}

// Response is the base class for CC responses.
// In Python: class Response()
type Response struct {
	type_   byte
	payload []byte
}

// NewResponse creates a new Response instance.
// In Python: def __init__(self, payload: memoryview)
func NewResponse(payload []byte) *Response {
	return &Response{
		type_:   payload[0],
		payload: payload,
	}
}

// String returns the hex string of the payload.
// In Python: def __str__(self) -> str
func (r *Response) String() string {
	return fmt.Sprintf("%x", r.payload)
}

// Type returns the response type.
// In Python: @property def type(self) -> int
func (r *Response) Type() byte {
	return r.type_
}

// Payload returns the response payload.
// In Python: @property def payload(self) -> bytes
func (r *Response) Payload() []byte {
	return r.payload
}

// Validate validates the response.
// In Python: @classmethod def validate(cls, payload: memoryview) -> None
func ValidateResponse(payload []byte) {
	// TODO
}

type ResponseInterface interface {
	Payload() []byte
	Type() byte
}

// ConstructResponse constructs a response object from raw data.
// In Python: @classmethod def construct(cls, frame: bytes) -> Union[ControlResponse, QueryResponse, Response]
func ConstructResponse(frame []byte) (ResponseInterface, error) {
	// Validate the frame
	err := msmart.Validate(frame, msmart.DeviceTypeCommercialAC)
	if err != nil {
		return nil, err
	}

	// Default to base class
	var responseClass ResponseInterface

	// Fetch the appropriate response class from the frame type
	frameType := frame[9]

	// Validate the payload (frame[10:-1] in Python, which is frame[10:len(frame)-1] in Go)
	ValidateResponse(frame[10 : len(frame)-1])

	// Build the response
	switch msmart.FrameType(frameType) {
	case msmart.FrameTypeQuery, msmart.FrameTypeReport:
		resp, err := NewQueryResponse(frame[10 : len(frame)-1])
		if err != nil {
			return nil, err
		}
		return resp, nil
	case msmart.FrameTypeControl:
		resp, err := NewControlResponse(frame[10 : len(frame)-1])
		if err != nil {
			return nil, err
		}
		return resp, nil
	default:
		responseClass = NewResponse(frame[10 : len(frame)-1])
		return responseClass, nil
	}
}

// QueryResponse is a response to query command.
// In Python: class QueryResponse(Response)
type QueryResponse struct {
	*Response

	// State properties
	PowerOn            bool
	TargetTemperature  float64
	IndoorTemperature  *float64
	OutdoorTemperature *float64
	Fahrenheit         bool
	TargetHumidity     byte
	IndoorHumidity     *byte
	OperationalMode    byte
	FanSpeed           byte
	VertSwingAngle     byte
	HorzSwingAngle     byte
	WindSense          byte
	Eco                bool
	Silent             bool
	Sleep              bool
	Purifier           byte
	Beep               bool
	Display            bool
	AuxMode            byte

	// Capabilities
	TargetTemperatureMin   float64
	TargetTemperatureMax   float64
	SupportsHumidity       bool
	SupportedOpModes       []byte
	SupportsFanSpeed       bool
	SupportsVertSwingAngle bool
	SupportsHorzSwingAngle bool
	SupportsWindSense      bool
	SupportsCO2Level       bool
	SupportsEco            bool
	SupportsSilent         bool
	SupportsSleep          bool
	SupportsSelfClean      bool
	SupportsPurifier       bool
	SupportsPurifierAuto   bool
	SupportsFilterLevel    bool
	SupportedAuxModes      []byte
}

// NewQueryResponse creates a new QueryResponse instance.
// In Python: def __init__(self, payload: memoryview)
func NewQueryResponse(payload []byte) (*QueryResponse, error) {
	r := &QueryResponse{
		Response: NewResponse(payload),

		// Initialize with defaults
		PowerOn:           false,
		TargetTemperature: 24,
		TargetHumidity:    40,
		OperationalMode:   0,
		FanSpeed:          0,
		VertSwingAngle:    0,
		HorzSwingAngle:    0,
		WindSense:         0,
		Purifier:          0,
		AuxMode:           0,

		TargetTemperatureMin: 17,
		TargetTemperatureMax: 30,
	}
	if err := r.parse(payload); err != nil {
		return nil, err
	}
	return r, nil
}

// parseTemperature parses a temperature value.
// In Python: def _parse_temperature(self, data: int) -> float
func (r *QueryResponse) parseTemperature(data byte) float64 {
	return float64(data)/2.0 - 40
}

// parse parses the query response payload.
// In Python: def _parse(self, payload: memoryview) -> None
func (r *QueryResponse) parse(payload []byte) error {
	// Query response starts with an 8 byte header
	// 0x01 - Basic data set
	// 0xFE - Indicates format of data
	// 2 bytes - Start index in protocol's "key_maps"
	// 2 bytes - End index in "key_maps"
	// 2 bytes - Length of section in bytes
	// Our ControlIds are translated indices in "key_maps"

	// Validate header
	if len(payload) < 2 || payload[0] != 0x01 || payload[1] != 0xFE {
		return NewInvalidResponseException(
			fmt.Sprintf("Query response payload '%x' lacks expected header 0x01FE.", payload))
	}

	if len(payload) < 88 {
		return nil
	}

	r.PowerOn = payload[8] != 0
	r.TargetTemperature = r.parseTemperature(payload[11])

	// Indoor temperature
	indoorTemp := float64(uint16(payload[12])<<8|uint16(payload[13])) / 10.0
	r.IndoorTemperature = &indoorTemp

	// Outdoor temperature
	if payload[14] != 0 {
		outdoorTemp := r.parseTemperature(payload[14])
		r.OutdoorTemperature = &outdoorTemp
	} else {
		r.OutdoorTemperature = nil
	}

	r.Fahrenheit = payload[21] != 0

	// Target humidity
	r.TargetHumidity = payload[24]

	// Indoor humidity
	if payload[25] != 0xFF {
		indoorHumidity := payload[25]
		r.IndoorHumidity = &indoorHumidity
	} else {
		r.IndoorHumidity = nil
	}

	r.OperationalMode = payload[31]
	r.FanSpeed = payload[34]

	r.VertSwingAngle = payload[41] // Replicated at payload[36]?
	r.HorzSwingAngle = payload[43] // Not replicated?

	// Wind sense: 0 - "Close", 1 - Follow, 2 - Avoid, 3 - Soft, 4 - Strong
	r.WindSense = payload[45]

	// TODO fault codes at payload[47:50]

	r.Eco = payload[56] != 0
	r.Silent = payload[58] != 0
	r.Sleep = payload[60] != 0

	r.Purifier = payload[75] // 0 - Auto, 1 - On, 2 - Off

	r.Beep = payload[80] != 0
	r.Display = payload[81] != 0

	// Aux mode: 0 - Auto, 1 - On, 2 - Off, 4 - "Separate"
	r.AuxMode = payload[87]

	return nil
}

// ParseCapabilities parses capabilities from the query response payload.
// In Python: def parse_capabilities(self) -> None
func (r *QueryResponse) ParseCapabilities() {
	payload := r.payload

	if len(payload) < 88 {
		return
	}

	// Additional cool/heat min/max temperatures available, but plugin only uses these
	r.TargetTemperatureMin = r.parseTemperature(payload[9])
	r.TargetTemperatureMax = r.parseTemperature(payload[10])

	r.SupportsHumidity = payload[23] != 0 // TODO unverified

	// Supported operational modes (bytes 26-30)
	r.SupportedOpModes = make([]byte, 0)
	for i := 26; i < 31 && i < len(payload); i++ {
		if payload[i] != 0 {
			r.SupportedOpModes = append(r.SupportedOpModes, payload[i])
		}
	}

	r.SupportsFanSpeed = payload[32] != 0

	r.SupportsVertSwingAngle = payload[40] != 0
	r.SupportsHorzSwingAngle = payload[42] != 0

	r.SupportsWindSense = payload[44] != 0

	r.SupportsCO2Level = payload[52] != 0

	r.SupportsEco = payload[55] != 0
	r.SupportsSilent = payload[57] != 0
	r.SupportsSleep = payload[59] != 0

	r.SupportsSelfClean = payload[61] != 0 // TODO unverified

	r.SupportsPurifier = payload[73] != 0
	r.SupportsPurifierAuto = payload[74] != 0 // TODO unverified

	r.SupportsFilterLevel = payload[78] != 0 // TODO unverified

	supportsAuxHeat := payload[82] != 0
	if supportsAuxHeat {
		// Supported aux modes (bytes 83-86)
		r.SupportedAuxModes = make([]byte, 0)
		for i := 83; i < 87 && i < len(payload); i++ {
			if payload[i] != 0 {
				r.SupportedAuxModes = append(r.SupportedAuxModes, payload[i])
			}
		}
	}
}

// ControlResponse is a response to control command.
// In Python: class ControlResponse(Response)
type ControlResponse struct {
	*Response
	states map[ControlId]interface{}
}

// NewControlResponse creates a new ControlResponse instance.
// In Python: def __init__(self, payload: memoryview)
func NewControlResponse(payload []byte) (*ControlResponse, error) {
	r := &ControlResponse{
		Response: NewResponse(payload),
		states:   make(map[ControlId]interface{}),
	}
	if err := r.parse(payload); err != nil {
		return nil, err
	}
	return r, nil
}

// parse parses the control response payload.
// In Python: def _parse(self, payload: memoryview) -> None
func (r *ControlResponse) parse(payload []byte) error {
	// Clear existing states
	r.states = make(map[ControlId]interface{})

	if len(payload) < 6 {
		return NewInvalidResponseException(
			fmt.Sprintf("Control response payload '%x' is too short.", payload))
	}

	// Loop through each entry
	// Each entry is 2 byte ID, 1 byte length, N byte value, 1 byte terminator 0xFF
	for len(payload) >= 5 {
		// Skip empty states
		size := payload[2]
		if size == 0 {
			// Zero length values still are at least 1 byte
			if len(payload) < 5 {
				break
			}
			payload = payload[5:]
			continue
		}

		// Unpack 16 bit ID
		rawId := binary.BigEndian.Uint16(payload[0:2])

		// Convert ID to enumerate type
		control := ControlId(rawId)

		// Check if control ID is known and log warning if not
		if !control.IsKnown() {
			slog.Warn("Unknown control ID", "id", fmt.Sprintf("0x%04X", rawId), "size", size)
		}

		// Parse the property
		if len(payload) >= 4+int(size) {
			value := control.Decode(payload[3 : 4+int(size)])
			if value != nil {
				r.states[control] = value
			}
		}

		// Advance to next entry
		if len(payload) < 4+int(size) {
			break
		}
		payload = payload[4+int(size):]
	}

	return nil
}

// GetControlState gets a control state by ID.
// In Python: def get_control_state(self, id: ControlId) -> Optional[Any]
func (r *ControlResponse) GetControlState(id ControlId) interface{} {
	return r.states[id]
}
