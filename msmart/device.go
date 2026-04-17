// Package msmart provides Go implementation of msmart-ng base device
package msmart

import (
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/RelicOfTesla/midea-msmart/msmart/device"
)

// Global LAN cache to reuse authenticated connections
var (
	lanCache   = make(map[string]*LAN)
	lanCacheMu sync.Mutex
)

// getLANCacheKey generates a cache key from IP and device ID
func getLANCacheKey(ip string, deviceID LanDeviceId) string {
	return fmt.Sprintf("%s:%d", ip, deviceID)
}

// DeviceBase represents a base device implementation
// This is a translation of Python's Device class from base_device.py
type DeviceBase struct {
	// Private fields (using lowercase naming convention in Go)
	ip         string
	port       int
	id         string
	deviceType device.DeviceType
	sn         string
	name       string
	version    int
	lan        *LAN
	supported  bool
	online     bool

	// Pre-set token and key for V3 devices (to skip cloud authentication)
	presetToken           device.Token
	presetKey             device.Key
	presetLocalKey        device.LocalKey
	presetLocalKeyExpired *time.Time

	// Supported capability overrides map
	// In Python: dict[str, tuple[str, type]]
	// In Go: map[string]CapabilityOverrideInfo
	supportedCapabilityOverrides map[string]CapabilityOverrideInfo
}

var _ device.Device = (*DeviceBase)(nil)
var _ device.DeviceAuthV3 = (*DeviceBase)(nil)

// CapabilityOverrideInfo stores information about capability overrides
type CapabilityOverrideInfo struct {
	AttrName  string
	ValueType reflect.Type
	IsFlag    bool // true if this is a Flag type (bitwise OR merging)
}

// NewBaseDevice creates a new DeviceBase instance
// This is the equivalent of Python's __init__ method
func NewBaseDevice(ip string, port int, _id string, deviceType device.DeviceType, opts ...device.DeviceOption) *DeviceBase {
	d := &DeviceBase{
		ip:                           ip,
		port:                         port,
		id:                           _id,
		deviceType:                   deviceType,
		supported:                    false,
		online:                       false,
		supportedCapabilityOverrides: make(map[string]CapabilityOverrideInfo),
	}

	// Apply optional parameters
	cfg := device.ApplyOptions(opts...)

	deviceId64, err := strconv.ParseInt(_id, 10, 64)
	if err != nil {
		slog.Warn("invalid device ID, using 0", "id", _id, "error", err)
		deviceId64 = 0
	}
	deviceID := LanDeviceId(deviceId64)
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
			d.lan = NewLAN(ip, port, deviceID, cfg.LocalKey, cfg.GetLocalKeyExpired())
			lanCache[cacheKey] = d.lan
			lanCacheMu.Unlock()
		}
	} else {
		// No cached LAN, create new one
		d.lan = NewLAN(ip, port, deviceID, cfg.LocalKey, cfg.GetLocalKeyExpired())
		lanCache[cacheKey] = d.lan
		lanCacheMu.Unlock()
	}

	if cfg.SN != nil {
		d.sn = *cfg.SN
	}
	if cfg.Name != nil {
		d.name = *cfg.Name
	}
	if cfg.Version != nil {
		d.version = *cfg.Version
	}
	if len(cfg.PresetToken) > 0 {
		d.presetToken = cfg.PresetToken
	}
	if len(cfg.PresetKey) > 0 {
		d.presetKey = cfg.PresetKey
	}
	if len(cfg.LocalKey) > 0 {
		d.presetLocalKey = cfg.LocalKey
	}
	if cfg.LocalKeyExpired != nil {
		d.presetLocalKeyExpired = cfg.LocalKeyExpired
	}

	return d
}

// SendCommand sends a command to the device and returns any responses
// This is the equivalent of Python's _send_command method
func (d *DeviceBase) SendCommand(ctx context.Context, command *Frame) ([][]byte, error) {
	data := command.ToBytes(nil)
	verboseLog("Sending command to %s:%d: %s", d.ip, d.port, hex.EncodeToString(data))

	start := time.Now()
	responses, err := d.lan.Send(ctx, data, Retries)
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
func (d *DeviceBase) SendBytes(ctx context.Context, data []byte) ([][]byte, error) {
	verboseLog("Sending bytes to %s:%d: %s", d.ip, d.port, hex.EncodeToString(data))

	start := time.Now()
	responses, err := d.lan.Send(ctx, data, Retries)
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
func (d *DeviceBase) SetOnline(online bool) {
	d.online = online
}

// SetSupported sets the supported status
func (d *DeviceBase) SetSupported(supported bool) {
	d.supported = supported
}

// Refresh refreshes the device state
// This is the equivalent of Python's async refresh method
// Returns an error to indicate it must be implemented by subclasses
func (d *DeviceBase) Refresh(context.Context) error {
	return fmt.Errorf("not implemented")
}

// Apply applies device configuration
// This is the equivalent of Python's async apply method
// Returns an error to indicate it must be implemented by subclasses
func (d *DeviceBase) Apply(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

// GetCapabilities implements [device.Device].
func (d *DeviceBase) GetCapabilities(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

// Authenticate authenticates with a V3 device
// This is the equivalent of Python's async authenticate method
func (d *DeviceBase) Authenticate(ctx context.Context, token device.Token, key device.Key) error {
	err := d.lan.Authenticate(ctx, token, key, Retries)
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

// IsAuthenticated checks if the device is authenticated (for V3 devices)
func (d *DeviceBase) IsAuthenticated() bool {
	return d.lan != nil && d.lan.IsAuthenticated()
}

// SetMaxConnectionLifetime sets the maximum connection lifetime of the LAN protocol
func (d *DeviceBase) SetMaxConnectionLifetime(seconds *int) {
	d.lan.SetMaxConnectionLifetime(seconds)
}

// GetIP returns the device IP address
// This is the equivalent of Python's @property def ip(self)
func (d *DeviceBase) GetIP() string {
	return d.ip
}

// GetPort returns the device port
// This is the equivalent of Python's @property def port(self)
func (d *DeviceBase) GetPort() int {
	return d.port
}

// GetID returns the device ID
// This is the equivalent of Python's @property def id(self)
func (d *DeviceBase) GetID() string {
	return d.id
}

// GetKeyInfo implements [device.DeviceAuthV3].
func (d *DeviceBase) GetKeyInfo() (token device.Token, key device.Key, localKey device.LocalKey, expired time.Time) {
	token = d.lan.Token()
	key = d.lan.Key()
	localKey, expired = d.lan.GetLocalKey()
	return
}

// GetType returns the device type
// This is the equivalent of Python's @property def type(self)
func (d *DeviceBase) GetType() device.DeviceType {
	return d.deviceType
}

// GetName returns the device name
// This is the equivalent of Python's @property def name(self)
func (d *DeviceBase) GetName() string {
	return d.name
}

// GetSN returns the device serial number
// This is the equivalent of Python's @property def sn(self)
func (d *DeviceBase) GetSN() string {
	return d.sn
}

// GetVersion returns the device version
// This is the equivalent of Python's @property def version(self)
func (d *DeviceBase) GetVersion() int {
	return d.version
}

// GetOnline returns whether the device is online
// This is the equivalent of Python's @property def online(self)
func (d *DeviceBase) GetOnline() bool {
	return d.online
}

// GetSupported returns whether the device is supported
// This is the equivalent of Python's @property def supported(self)
func (d *DeviceBase) GetSupported() bool {
	return d.supported
}

// ToDict returns the device as a dictionary
// This is the equivalent of Python's to_dict method
func (d *DeviceBase) ToDict() map[string]interface{} {
	result := map[string]interface{}{
		"ip":        d.ip,
		"port":      d.port,
		"id":        d.id,
		"online":    d.online,
		"supported": d.supported,
		"type":      d.deviceType,
	}

	// Handle optional fields
	if d.name != "" {
		result["name"] = d.name
	}
	if d.sn != "" {
		result["sn"] = d.sn
	}

	token, key, localKey, expired := d.GetKeyInfo()
	if len(key) > 0 {
		result["key"] = hex.EncodeToString(key)
	}
	if len(token) > 0 {
		result["token"] = hex.EncodeToString(token)
	}
	if len(localKey) > 0 {
		result["localKey"] = hex.EncodeToString(localKey)
	}
	if len(localKey) > 0 {
		result["localKeyExpired"] = expired.Format(time.RFC3339)
	}

	return result
}

// CapabilitiesDict returns the device capabilities as a dictionary
// This is the equivalent of Python's async capabilities_dict method
// Returns an error to indicate it must be implemented by subclasses
func (d *DeviceBase) CapabilitiesDict() map[string]interface{} {
	return nil
}

// String returns the string representation of the device
// This is the equivalent of Python's __str__ method
func (d *DeviceBase) String() string {
	return fmt.Sprintf("%v", d.ToDict())
}

// SerializeCapabilities dumps device capabilities as an easily serializable dict
// This is the equivalent of Python's serialize_capabilities method
func (d *DeviceBase) SerializeCapabilities() (map[string]interface{}, error) {
	capabilities := d.CapabilitiesDict()
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
func (d *DeviceBase) OverrideCapabilities(overrides map[string]interface{}, merge bool) error {
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
func (d *DeviceBase) applyFlagOverride(attrName string, valueType reflect.Type, names []interface{}, merge bool, field reflect.Value) error {
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
func (d *DeviceBase) applyEnumOverride(attrName string, valueType reflect.Type, names []interface{}, merge bool, field reflect.Value) error {
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

// GetSupportedCapabilityOverrides returns the supported capability overrides map
func (d *DeviceBase) GetSupportedCapabilityOverrides() map[string]CapabilityOverrideInfo {
	return d.supportedCapabilityOverrides
}

// SetSupportedCapabilityOverrides sets the supported capability overrides map
func (d *DeviceBase) SetSupportedCapabilityOverrides(overrides map[string]CapabilityOverrideInfo) {
	d.supportedCapabilityOverrides = overrides
}

func (d *DeviceBase) AuthenticateFromPreset(ctx context.Context) (bool, error) {
	if d.presetToken != nil && d.presetKey != nil {
		verboseLog("Using pre-set token/key for V3 device authentication (skipping cloud)")
		if err := d.Authenticate(ctx, d.presetToken, d.presetKey); err != nil {
			verboseLog("Pre-set token/key authentication failed: %v", err)
			// Fall through to cloud authentication
		} else {
			return true, nil
		}
	}
	return false, nil
}
