// Package msmart provides Go implementation of msmart-ng base device
package msmart

import (
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"reflect"
	"sync"
	"time"
)

// Global LAN cache to reuse authenticated connections
var (
	lanCache   = make(map[string]*LAN)
	lanCacheMu sync.Mutex
)

// getLANCacheKey generates a cache key from IP and device ID
func getLANCacheKey(ip string, deviceID int) string {
	return fmt.Sprintf("%s:%d", ip, deviceID)
}

// Device represents a base device
// This is a translation of Python's Device class from base_device.py
type Device struct {
	// Private fields (using lowercase naming convention in Go)
	ip         string
	port       int
	id         int
	deviceType DeviceType
	sn         *string
	name       *string
	version    *int
	lan        *LAN
	supported  bool
	online     bool

	// Pre-set token and key for V3 devices (to skip cloud authentication)
	presetToken []byte
	presetKey   []byte

	// Supported capability overrides map
	// In Python: dict[str, tuple[str, type]]
	// In Go: map[string]CapabilityOverrideInfo
	supportedCapabilityOverrides map[string]CapabilityOverrideInfo
}

// CapabilityOverrideInfo stores information about capability overrides
type CapabilityOverrideInfo struct {
	AttrName  string
	ValueType reflect.Type
	IsFlag    bool // true if this is a Flag type (bitwise OR merging)
}

// DeviceOption is a functional option for Device configuration
type DeviceOption func(*Device)

// NewDevice creates a new Device instance
// This is the equivalent of Python's __init__ method
func NewDevice(ip string, port int, deviceID int, deviceType DeviceType, opts ...DeviceOption) *Device {
	d := &Device{
		ip:                           ip,
		port:                         port,
		id:                           deviceID,
		deviceType:                   deviceType,
		supported:                    false,
		online:                       false,
		supportedCapabilityOverrides: make(map[string]CapabilityOverrideInfo),
	}

	// Try to reuse cached LAN object
	lanCacheMu.Lock()
	cacheKey := getLANCacheKey(ip, deviceID)
	if cachedLAN, exists := lanCache[cacheKey]; exists {
		// Check if the cached LAN is still usable (has valid authentication)
		if cachedLAN.IsAuthenticated() {
			d.lan = cachedLAN
			lanCacheMu.Unlock()
		} else {
			// Cached LAN is not usable, create new one
			d.lan = NewLAN(ip, port, int64(deviceID))
			lanCache[cacheKey] = d.lan
			lanCacheMu.Unlock()
		}
	} else {
		// No cached LAN, create new one
		d.lan = NewLAN(ip, port, int64(deviceID))
		lanCache[cacheKey] = d.lan
		lanCacheMu.Unlock()
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

// WithTokenKey sets the pre-set token and key for V3 devices
func WithTokenKey(token, key []byte) DeviceOption {
	return func(d *Device) {
		d.presetToken = token
		d.presetKey = key
	}
}

// SendCommand sends a command to the device and returns any responses
// This is the equivalent of Python's _send_command method
func (d *Device) SendCommand(command *Frame) ([][]byte, error) {
	data := command.ToBytes(nil)
	verboseLog("Sending command to %s:%d: %s", d.ip, d.port, hex.EncodeToString(data))

	start := time.Now()
	responses, err := d.lan.Send(data, Retries)
	if err != nil {
		if _, ok := err.(*ProtocolError); ok {
			slog.Error("Network error", "ip", d.ip, "port", d.port, "error", err)
			return nil, err
		}
	}

	responseTime := time.Since(start).Seconds()

	if len(responses) == 0 {
		slog.Warn("No response from device", "ip", d.ip, "port", d.port, "response_time", responseTime)
	} else {
		verboseLog("Response from %s:%d in %f seconds.", d.ip, d.port, responseTime)
	}

	return responses, nil
}

// SendBytes sends raw bytes to the device and returns any responses
// This is used by sub-packages that have their own command serialization
func (d *Device) SendBytes(data []byte) ([][]byte, error) {
	verboseLog("Sending bytes to %s:%d: %s", d.ip, d.port, hex.EncodeToString(data))

	start := time.Now()
	responses, err := d.lan.Send(data, Retries)
	if err != nil {
		if _, ok := err.(*ProtocolError); ok {
			slog.Error("Network error", "ip", d.ip, "port", d.port, "error", err)
			return nil, err
		}
	}

	responseTime := time.Since(start).Seconds()

	if len(responses) == 0 {
		slog.Warn("No response from device", "ip", d.ip, "port", d.port, "response_time", responseTime)
	} else {
		verboseLog("Response from %s:%d in %f seconds.", d.ip, d.port, responseTime)
	}

	return responses, nil
}

// SetOnline sets the online status
func (d *Device) SetOnline(online bool) {
	d.online = online
}

// SetSupported sets the supported status
func (d *Device) SetSupported(supported bool) {
	d.supported = supported
}

// Refresh refreshes the device state
// This is the equivalent of Python's async refresh method
// Returns an error to indicate it must be implemented by subclasses
func (d *Device) Refresh(context.Context) error {
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

// GetLocalKey returns the current local key and its expiration time
// Returns nil if not authenticated or no local key is set
func (d *Device) GetLocalKey() ([]byte, time.Time) {
	if d.lan == nil {
		return nil, time.Time{}
	}
	return d.lan.GetLocalKey()
}

// SetLocalKey sets the local key and expiration time directly
// This allows reusing a cached local key without re-authenticating
// Returns true if the key was set successfully, false if expired
func (d *Device) SetLocalKey(localKey []byte, expiration time.Time) bool {
	if d.lan == nil {
		return false
	}
	return d.lan.SetLocalKey(localKey, expiration)
}

// IsAuthenticated checks if the device is authenticated (for V3 devices)
func (d *Device) IsAuthenticated() bool {
	return d.lan != nil && d.lan.IsAuthenticated()
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

// EnumNameResolver is an interface for enum types that support name-to-value resolution
// This allows OverrideCapabilities to convert string names to enum values
type EnumNameResolver interface {
	// GetFromNameWithDefault returns the enum value for a given name, or the default if not found
	GetFromNameWithDefault(name string, default_ interface{}) interface{}
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

		// Handle enum/flag overrides
		// Value should be a list of enum names
		sliceVal, ok := value.([]interface{})
		if !ok {
			// Also accept []string directly
			if strSlice, ok := value.([]string); ok {
				sliceVal = make([]interface{}, len(strSlice))
				for i, s := range strSlice {
					sliceVal[i] = s
				}
			} else {
				return fmt.Errorf("'%s' must be a list", key)
			}
		}

		// Get the field for potential merge
		rv := reflect.ValueOf(d).Elem()
		field := rv.FieldByName(attrName)

		// Handle Flag types (bitwise OR merging)
		if overrideInfo.IsFlag {
			if err := d.applyFlagOverride(attrName, valueType, sliceVal, merge, field); err != nil {
				return err
			}
			continue
		}

		// Handle regular enum types
		if err := d.applyEnumOverride(attrName, valueType, sliceVal, merge, field); err != nil {
			return err
		}
	}

	return nil
}

// applyFlagOverride handles Flag enum types with bitwise OR merging
func (d *Device) applyFlagOverride(attrName string, valueType reflect.Type, names []interface{}, merge bool, field reflect.Value) error {
	// Convert names to flag values using reflection
	// We need to find the GetFromName method on the type
	flags := reflect.New(valueType).Elem()

	// Get the zero value and FindByName method
	zeroValue := reflect.Zero(valueType)

	// Look for GetFromName method (case insensitive match)
	getFromNameMethod := zeroValue.MethodByName("GetFromName")
	if !getFromNameMethod.IsValid() {
		// Try to find Values method and then match by String()
		valuesMethod := zeroValue.MethodByName("Values")
		if valuesMethod.IsValid() {
			// Use Values() to find matching names
			allValues := valuesMethod.Call(nil)[0]

			for _, name := range names {
				nameStr, ok := name.(string)
				if !ok {
					return fmt.Errorf("flag name must be a string")
				}

				found := false
				for i := 0; i < allValues.Len(); i++ {
					v := allValues.Index(i)
					stringMethod := v.MethodByName("String")
					if stringMethod.IsValid() {
						strVal := stringMethod.Call(nil)[0].String()
						if strVal == nameStr {
							// Found matching value, OR it into flags
							flags = reflect.ValueOf(flags.Int() | v.Int())
							found = true
							break
						}
					}
				}

				if !found {
					return fmt.Errorf("invalid value '%s' for flag type", nameStr)
				}
			}
		} else {
			return fmt.Errorf("unsupported flag type for '%s'", attrName)
		}
	} else {
		// Use GetFromName method, but also validate using Values()
		// We need Values() to validate that the name is actually valid
		valuesMethod := zeroValue.MethodByName("Values")
		if !valuesMethod.IsValid() {
			// No Values() method, just use GetFromName without validation
			for _, name := range names {
				nameStr, ok := name.(string)
				if !ok {
					return fmt.Errorf("flag name must be a string")
				}

				// Call GetFromName(name)
				result := getFromNameMethod.Call([]reflect.Value{reflect.ValueOf(nameStr)})
				if len(result) > 0 {
					flags = reflect.ValueOf(flags.Int() | result[0].Int())
				}
			}
		} else {
			// Use Values() to validate names, then GetFromName to get values
			allValues := valuesMethod.Call(nil)[0]

			for _, name := range names {
				nameStr, ok := name.(string)
				if !ok {
					return fmt.Errorf("flag name must be a string")
				}

				// First validate that the name exists by checking String() values
				found := false
				for i := 0; i < allValues.Len(); i++ {
					v := allValues.Index(i)
					stringMethod := v.MethodByName("String")
					if stringMethod.IsValid() {
						strVal := stringMethod.Call(nil)[0].String()
						if strVal == nameStr {
							found = true
							break
						}
					}
				}

				if !found {
					return fmt.Errorf("invalid value '%s' for flag type", nameStr)
				}

				// Now call GetFromName to get the value
				result := getFromNameMethod.Call([]reflect.Value{reflect.ValueOf(nameStr)})
				if len(result) > 0 {
					flags = reflect.ValueOf(flags.Int() | result[0].Int())
				}
			}
		}
	}

	// Handle merge with existing flags
	if merge {
		// Check if field is a CapabilityManager
		if field.IsValid() && field.Type() == reflect.TypeOf(&CapabilityManager{}) {
			if !field.IsNil() {
				cm := field.Interface().(*CapabilityManager)
				existingFlags := cm.Flags()
				flags = reflect.ValueOf(flags.Int() | existingFlags)
			}
		} else if field.IsValid() && field.CanSet() {
			flags = reflect.ValueOf(flags.Int() | field.Int())
		}
	}

	// Set the flags
	if field.IsValid() {
		// Check if field is a CapabilityManager
		if field.Type() == reflect.TypeOf(&CapabilityManager{}) {
			if field.IsNil() {
				field.Set(reflect.ValueOf(NewCapabilityManager(flags.Int())))
			} else {
				cm := field.Interface().(*CapabilityManager)
				cm.SetFlags(flags.Int())
			}
		} else if field.CanSet() {
			field.SetInt(flags.Int())
		}
	}

	return nil
}

// applyEnumOverride handles regular enum types
func (d *Device) applyEnumOverride(attrName string, valueType reflect.Type, names []interface{}, merge bool, field reflect.Value) error {
	// Convert names to enum values using reflection
	var members []interface{}

	// Get the zero value and try to find a lookup method
	zeroValue := reflect.Zero(valueType)

	// Look for GetFromName method
	getFromNameMethod := zeroValue.MethodByName("GetFromName")
	if !getFromNameMethod.IsValid() {
		// Try to find Values method and then match by String()
		valuesMethod := zeroValue.MethodByName("Values")
		if !valuesMethod.IsValid() {
			return fmt.Errorf("unsupported enum type for '%s'", attrName)
		}

		// Use Values() to find matching names
		allValues := valuesMethod.Call(nil)[0]

		for _, name := range names {
			nameStr, ok := name.(string)
			if !ok {
				return fmt.Errorf("enum name must be a string")
			}

			found := false
			for i := 0; i < allValues.Len(); i++ {
				v := allValues.Index(i)
				stringMethod := v.MethodByName("String")
				if stringMethod.IsValid() {
					strVal := stringMethod.Call(nil)[0].String()
					if strVal == nameStr {
						members = append(members, v.Interface())
						found = true
						break
					}
				}
			}

			if !found {
				return fmt.Errorf("invalid value '%s' for enum type '%s'", nameStr, attrName)
			}
		}
	} else {
		// Use GetFromName method, but also validate using Values()
		// We need Values() to validate that the name is actually valid
		valuesMethod := zeroValue.MethodByName("Values")
		if !valuesMethod.IsValid() {
			// No Values() method, just use GetFromName without validation
			for _, name := range names {
				nameStr, ok := name.(string)
				if !ok {
					return fmt.Errorf("enum name must be a string")
				}

				// Call GetFromName(name)
				result := getFromNameMethod.Call([]reflect.Value{reflect.ValueOf(nameStr)})
				if len(result) > 0 {
					members = append(members, result[0].Interface())
				}
			}
		} else {
			// Use Values() to validate names, then GetFromName to get values
			allValues := valuesMethod.Call(nil)[0]

			for _, name := range names {
				nameStr, ok := name.(string)
				if !ok {
					return fmt.Errorf("enum name must be a string")
				}

				// First validate that the name exists by checking String() values
				found := false
				for i := 0; i < allValues.Len(); i++ {
					v := allValues.Index(i)
					stringMethod := v.MethodByName("String")
					if stringMethod.IsValid() {
						strVal := stringMethod.Call(nil)[0].String()
						if strVal == nameStr {
							found = true
							break
						}
					}
				}

				if !found {
					return fmt.Errorf("invalid value '%s' for enum type '%s'", nameStr, attrName)
				}

				// Now call GetFromName to get the value
				result := getFromNameMethod.Call([]reflect.Value{reflect.ValueOf(nameStr)})
				if len(result) > 0 {
					members = append(members, result[0].Interface())
				}
			}
		}
	}

	// Handle merge with existing values
	if merge && field.IsValid() {
		// Get existing slice values
		if field.Kind() == reflect.Slice {
			for i := 0; i < field.Len(); i++ {
				members = append(members, field.Index(i).Interface())
			}
			// Remove duplicates (simple approach)
			members = uniqueInterfaceSlice(members)
		}
	}

	// Create a slice of the correct type and set it
	if field.IsValid() && field.CanSet() {
		sliceType := reflect.SliceOf(valueType)
		slice := reflect.MakeSlice(sliceType, len(members), len(members))
		for i, m := range members {
			slice.Index(i).Set(reflect.ValueOf(m))
		}
		field.Set(slice)
	}

	return nil
}

// uniqueInterfaceSlice removes duplicate values from a slice
func uniqueInterfaceSlice(slice []interface{}) []interface{} {
	keys := make(map[interface{}]bool)
	result := []interface{}{}
	for _, v := range slice {
		if !keys[v] {
			keys[v] = true
			result = append(result, v)
		}
	}
	return result
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

// GetSupportedCapabilityOverrides returns the supported capability overrides map
func (d *Device) GetSupportedCapabilityOverrides() map[string]CapabilityOverrideInfo {
	return d.supportedCapabilityOverrides
}

// SetSupportedCapabilityOverrides sets the supported capability overrides map
func (d *Device) SetSupportedCapabilityOverrides(overrides map[string]CapabilityOverrideInfo) {
	d.supportedCapabilityOverrides = overrides
}
