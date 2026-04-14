// Package msmart_ng_go provides Go implementation of msmart-ng base device
package msmart

import (
	"encoding/hex"
	"fmt"
	"log"
	"reflect"
	"time"
)

// Logger for the package
var deviceLogger = log.Default()

// Device represents a base device
// This is a translation of Python's Device class from base_device.py
type Device struct {
	// Private fields (using lowercase naming convention in Go)
	ip        string
	port      int
	id        int
	deviceType DeviceType
	sn        *string
	name      *string
	version   *int
	lan       *LAN
	supported bool
	online    bool

	// Supported capability overrides map
	// In Python: dict[str, tuple[str, type]]
	// In Go: map[string]CapabilityOverrideInfo
	supportedCapabilityOverrides map[string]CapabilityOverrideInfo
}

// CapabilityOverrideInfo stores information about capability overrides
type CapabilityOverrideInfo struct {
	AttrName  string
	ValueType reflect.Type
}

// DeviceOption is a functional option for Device configuration
type DeviceOption func(*Device)

// NewDevice creates a new Device instance
// This is the equivalent of Python's __init__ method
func NewDevice(ip string, port int, deviceID int, deviceType DeviceType, opts ...DeviceOption) *Device {
	d := &Device{
		ip:        ip,
		port:      port,
		id:        deviceID,
		deviceType: deviceType,
		lan:       NewLAN(ip, port, int64(deviceID)),
		supported: false,
		online:    false,
		supportedCapabilityOverrides: make(map[string]CapabilityOverrideInfo),
	}

	// Apply optional parameters
	for _, opt := range opts {
		opt(d)
	}

	return d
}

// WithSN sets the serial number
func WithSN(sn string) DeviceOption {
	return func(d *Device) {
		d.sn = &sn
	}
}

// WithName sets the device name
func WithName(name string) DeviceOption {
	return func(d *Device) {
		d.name = &name
	}
}

// WithVersion sets the device version
func WithVersion(version int) DeviceOption {
	return func(d *Device) {
		d.version = &version
	}
}

// SendCommand sends a command to the device and returns any responses
// This is the equivalent of Python's _send_command method
func (d *Device) SendCommand(command *Frame) ([][]byte, error) {
	data := command.ToBytes(nil)
	deviceLogger.Printf("DEBUG: Sending command to %s:%d: %s", d.ip, d.port, hex.EncodeToString(data))

	start := time.Now()
	responses, err := d.lan.Send(data, Retries)
	if err != nil {
		if _, ok := err.(*ProtocolError); ok {
			deviceLogger.Printf("ERROR: Network error %s:%d: %v", d.ip, d.port, err)
			return nil, err
		}
	}

	responseTime := time.Since(start).Seconds()

	if len(responses) == 0 {
		deviceLogger.Printf("WARNING: No response from %s:%d in %f seconds.", d.ip, d.port, responseTime)
	} else {
		deviceLogger.Printf("DEBUG: Response from %s:%d in %f seconds.", d.ip, d.port, responseTime)
	}

	return responses, nil
}

// Refresh refreshes the device state
// This is the equivalent of Python's async refresh method
// Returns an error to indicate it must be implemented by subclasses
func (d *Device) Refresh() error {
	return fmt.Errorf("not implemented")
}

// Apply applies device configuration
// This is the equivalent of Python's async apply method
// Returns an error to indicate it must be implemented by subclasses
func (d *Device) Apply() error {
	return fmt.Errorf("not implemented")
}

// Authenticate authenticates with a V3 device
// This is the equivalent of Python's async authenticate method
func (d *Device) Authenticate(token Token, key Key) error {
	err := d.lan.Authenticate(token, key, Retries)
	if err != nil {
		if _, ok := err.(*ProtocolError); ok {
			return &AuthenticationError{Cause: err}
		}
		if _, ok := err.(interface{ Timeout() bool }); ok {
			return &AuthenticationError{Cause: err}
		}
		return err
	}
	return nil
}

// SetMaxConnectionLifetime sets the maximum connection lifetime of the LAN protocol
func (d *Device) SetMaxConnectionLifetime(seconds *int) {
	d.lan.SetMaxConnectionLifetime(seconds)
}

// GetIP returns the device IP address
// This is the equivalent of Python's @property def ip(self)
func (d *Device) GetIP() string {
	return d.ip
}

// GetPort returns the device port
// This is the equivalent of Python's @property def port(self)
func (d *Device) GetPort() int {
	return d.port
}

// GetID returns the device ID
// This is the equivalent of Python's @property def id(self)
func (d *Device) GetID() int {
	return d.id
}

// GetToken returns the device token as a hex string
// This is the equivalent of Python's @property def token(self)
func (d *Device) GetToken() *string {
	token := d.lan.Token()
	if token == nil {
		return nil
	}
	hexToken := hex.EncodeToString(token)
	return &hexToken
}

// GetKey returns the device key as a hex string
// This is the equivalent of Python's @property def key(self)
func (d *Device) GetKey() *string {
	key := d.lan.Key()
	if key == nil {
		return nil
	}
	hexKey := hex.EncodeToString(key)
	return &hexKey
}

// GetType returns the device type
// This is the equivalent of Python's @property def type(self)
func (d *Device) GetType() DeviceType {
	return d.deviceType
}

// GetName returns the device name
// This is the equivalent of Python's @property def name(self)
func (d *Device) GetName() *string {
	return d.name
}

// GetSN returns the device serial number
// This is the equivalent of Python's @property def sn(self)
func (d *Device) GetSN() *string {
	return d.sn
}

// GetVersion returns the device version
// This is the equivalent of Python's @property def version(self)
func (d *Device) GetVersion() *int {
	return d.version
}

// GetOnline returns whether the device is online
// This is the equivalent of Python's @property def online(self)
func (d *Device) GetOnline() bool {
	return d.online
}

// GetSupported returns whether the device is supported
// This is the equivalent of Python's @property def supported(self)
func (d *Device) GetSupported() bool {
	return d.supported
}

// ToDict returns the device as a dictionary
// This is the equivalent of Python's to_dict method
func (d *Device) ToDict() map[string]interface{} {
	result := map[string]interface{}{
		"ip":        d.ip,
		"port":      d.port,
		"id":        d.id,
		"online":    d.online,
		"supported": d.supported,
		"type":      d.deviceType,
	}

	// Handle optional fields
	if d.name != nil {
		result["name"] = *d.name
	} else {
		result["name"] = nil
	}

	if d.sn != nil {
		result["sn"] = *d.sn
	} else {
		result["sn"] = nil
	}

	if d.GetKey() != nil {
		result["key"] = *d.GetKey()
	} else {
		result["key"] = nil
	}

	if d.GetToken() != nil {
		result["token"] = *d.GetToken()
	} else {
		result["token"] = nil
	}

	return result
}

// CapabilitiesDict returns the device capabilities as a dictionary
// This is the equivalent of Python's async capabilities_dict method
// Returns an error to indicate it must be implemented by subclasses
func (d *Device) CapabilitiesDict() (map[string]interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

// String returns the string representation of the device
// This is the equivalent of Python's __str__ method
func (d *Device) String() string {
	return fmt.Sprintf("%v", d.ToDict())
}

// SerializeCapabilities dumps device capabilities as an easily serializable dict
// This is the equivalent of Python's serialize_capabilities method
func (d *Device) SerializeCapabilities() (map[string]interface{}, error) {
	capabilities, err := d.CapabilitiesDict()
	if err != nil {
		return nil, err
	}
	result := serialize(capabilities)
	if m, ok := result.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, fmt.Errorf("failed to serialize capabilities")
}

// serialize recursively converts values into serializable primitives
// This is the equivalent of Python's _serialize function inside serialize_capabilities
func serialize(value interface{}) interface{} {
	// Handle nil
	if value == nil {
		return nil
	}

	// Handle Enum types (in Go, we'd use a custom interface or type assertion)
	// For now, we handle it generically
	if enum, ok := value.(interface{ Name() string }); ok {
		return enum.Name()
	}

	// Handle map
	if m, ok := value.(map[string]interface{}); ok {
		result := make(map[string]interface{})
		for k, v := range m {
			result[k] = serialize(v)
		}
		return result
	}

	// Handle slice/array
	if slice, ok := value.([]interface{}); ok {
		result := make([]interface{}, len(slice))
		for i, v := range slice {
			result[i] = serialize(v)
		}
		return result
	}

	// Handle other slice types via reflection
	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		result := make([]interface{}, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			result[i] = serialize(rv.Index(i).Interface())
		}
		return result
	}

	// Return as-is for primitive types
	return value
}

// OverrideCapabilities overrides device capabilities via serialized dict
// This is the equivalent of Python's override_capabilities method
func (d *Device) OverrideCapabilities(overrides map[string]interface{}, merge bool) error {
	// Get supported overrides
	supportedOverrides := d.supportedCapabilityOverrides

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

		// Handle numeric overrides
		if valueType == reflect.TypeOf(float64(0)) {
			// Check if value is numeric
			floatVal, ok := toFloat(value)
			if !ok {
				return fmt.Errorf("'%s' must be a number", key)
			}

			// Apply using reflection
			rv := reflect.ValueOf(d).Elem()
			field := rv.FieldByName(attrName)
			if field.IsValid() && field.CanSet() {
				field.SetFloat(floatVal)
			}
			continue
		}

		// Handle enum/flag overrides (simplified for Go)
		// In Python, this uses issubclass(value_type, Enum)
		// In Go, we'd need a different approach for enums
		// This is a simplified version
		sliceVal, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("'%s' must be a list", key)
		}

		// For simplicity, we just set the slice value
		// The original Python code has more complex enum handling
		rv := reflect.ValueOf(d).Elem()
		field := rv.FieldByName(attrName)
		if field.IsValid() && field.CanSet() {
			field.Set(reflect.ValueOf(sliceVal))
		}
	}

	return nil
}

// toFloat attempts to convert a value to float64
func toFloat(value interface{}) (float64, bool) {
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

// Construct constructs a device object based on the provided device type
// This is the equivalent of Python's @classmethod construct
// In Go, we use a function instead of a class method
func Construct(deviceType DeviceType, opts ...DeviceOption) *Device {
	// In the original Python code, this creates different device types
	// (AirConditioner, CommercialAirConditioner) based on deviceType
	// For this translation, we just return a generic Device
	return NewDevice("", 0, 0, deviceType, opts...)
}
