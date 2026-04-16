// Package msmart provides utility classes and methods for Midea AC.
package msmart

import (
	"log/slog"
)

// Contains checks if a slice contains an element.
// This is a generic helper that works with any comparable type.
func Contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// CapabilityManager is a minimal wrapper class to make mutable capability flags.
// In Go, we use a struct with methods instead of a class.
// Since Go doesn't have Python's Flag enum type, we use int64 for flexibility.
type CapabilityManager struct {
	flags int64
}

// NewCapabilityManager creates a new CapabilityManager with the default value.
func NewCapabilityManager(default_ int64) *CapabilityManager {
	return &CapabilityManager{flags: default_}
}

// Value returns the integer value of the flags.
func (cm *CapabilityManager) Value() int64 {
	return cm.flags
}

// Flags returns the current flags.
func (cm *CapabilityManager) Flags() int64 {
	return cm.flags
}

// SetFlags sets the flags.
func (cm *CapabilityManager) SetFlags(flags int64) {
	cm.flags = flags
}

// Has checks if a flag is set.
func (cm *CapabilityManager) Has(flag int64) bool {
	return cm.flags&flag != 0
}

// Set enables or disables a flag.
func (cm *CapabilityManager) Set(flag int64, enable bool) {
	if enable {
		cm.flags |= flag
	} else {
		cm.flags &= ^flag
	}
}

// MideaIntEnum is a helper interface for IntEnum-like types in Go.
// Since Go doesn't have inheritance for enums like Python's IntEnum,
// we use an interface to define common behavior.
//
// Python's MideaIntEnum provides:
//   - list() -> returns all enum values
//   - get_from_value(value, default) -> gets enum from integer value
//   - get_from_name(name, default) -> gets enum from string name
//
// In Go, the idiomatic approach is to define type-specific functions:
//   - FanSpeedList() -> []FanSpeed
//   - FanSpeedFromValue(value byte) -> FanSpeed
//   - FanSpeedFromName(name string) (FanSpeed, error)
//
// The MideaIntEnum interface defines the common methods that enum types should implement.
// Note: This interface uses int for Value() for generality, but most implementations
// use byte-backed types. Use MideaIntEnumByte for byte-backed enums.
type MideaIntEnum interface {
	Value() int
	String() string
}

// MideaIntEnumByte is an interface for byte-backed IntEnum-like types.
// This matches the common pattern in device.go where enum types are defined as `type X byte`.
type MideaIntEnumByte interface {
	Value() byte
	String() string
}

// MideaIntEnumByteHelper provides helper methods for byte-backed enum types.
// It stores enum values and provides List(), GetFromValue(), and GetFromName() methods.
//
// Usage:
//
//	var fanSpeedHelper = NewMideaIntEnumByteHelper("FanSpeed", FanSpeedDefault).
//		WithValues(FanSpeedL1, FanSpeedL2, FanSpeedAuto)
//
//	// Or register with name mapping for GetFromName support:
//	var fanSpeedHelper = NewMideaIntEnumByteHelper("FanSpeed", FanSpeedDefault).
//		WithNamedValues(map[string]FanSpeed{"L1": FanSpeedL1, "AUTO": FanSpeedAuto})
//
//	// Then use:
//	allSpeeds := fanSpeedHelper.List()
//	speed := fanSpeedHelper.GetFromValue(0x01) // Returns FanSpeedL1
//	speed = fanSpeedHelper.GetFromName("AUTO") // Returns FanSpeedAuto
type MideaIntEnumByteHelper[T MideaIntEnumByte] struct {
	enumName     string
	defaultValue T
	values       []T
	nameMap      map[string]T
	valueMap     map[byte]T
}

// NewMideaIntEnumByteHelper creates a new helper for a byte-backed enum type.
func NewMideaIntEnumByteHelper[T MideaIntEnumByte](enumName string, defaultValue T) *MideaIntEnumByteHelper[T] {
	return &MideaIntEnumByteHelper[T]{
		enumName:     enumName,
		defaultValue: defaultValue,
		values:       nil,
		nameMap:      make(map[string]T),
		valueMap:     make(map[byte]T),
	}
}

// WithValues registers enum values for List() and GetFromValue() support.
func (h *MideaIntEnumByteHelper[T]) WithValues(values ...T) *MideaIntEnumByteHelper[T] {
	h.values = values
	h.valueMap = make(map[byte]T, len(values))
	for _, v := range values {
		h.valueMap[v.Value()] = v
	}
	return h
}

// WithNamedValues registers named enum values for GetFromName() support.
func (h *MideaIntEnumByteHelper[T]) WithNamedValues(nameMap map[string]T) *MideaIntEnumByteHelper[T] {
	h.nameMap = nameMap
	return h
}

// List returns all registered enum values.
// In Python: @classmethod def list(cls) -> list[MideaIntEnum]
func (h *MideaIntEnumByteHelper[T]) List() []T {
	if h.values == nil {
		return []T{}
	}
	result := make([]T, len(h.values))
	copy(result, h.values)
	return result
}

// GetFromValue gets an enum value from its byte value.
// Returns the default value if the value is not found.
// In Python: @classmethod def get_from_value(cls, value: Optional[int], default: Optional[MideaIntEnum] = None)
func (h *MideaIntEnumByteHelper[T]) GetFromValue(value byte) T {
	if v, ok := h.valueMap[value]; ok {
		return v
	}
	slog.Debug("Unknown enum value", "enum", h.enumName, "value", value)
	return h.defaultValue
}

// GetFromValueWithDefault gets an enum value from its byte value with a custom default.
func (h *MideaIntEnumByteHelper[T]) GetFromValueWithDefault(value byte, default_ T) T {
	if v, ok := h.valueMap[value]; ok {
		return v
	}
	slog.Debug("Unknown enum value", "enum", h.enumName, "value", value)
	return default_
}

// GetFromName gets an enum value from its string name.
// Returns the default value if the name is not found.
// In Python: @classmethod def get_from_name(cls, name: Optional[str], default: Optional[MideaIntEnum] = None)
func (h *MideaIntEnumByteHelper[T]) GetFromName(name string) T {
	if v, ok := h.nameMap[name]; ok {
		return v
	}
	slog.Debug("Unknown enum name", "enum", h.enumName, "name", name)
	return h.defaultValue
}

// GetFromNameWithDefault gets an enum value from its string name with a custom default.
func (h *MideaIntEnumByteHelper[T]) GetFromNameWithDefault(name string, default_ T) T {
	if v, ok := h.nameMap[name]; ok {
		return v
	}
	slog.Debug("Unknown enum name", "enum", h.enumName, "name", name)
	return default_
}

// GetNameMap returns the name-to-value mapping for this enum.
func (h *MideaIntEnumByteHelper[T]) GetNameMap() map[string]T {
	result := make(map[string]T, len(h.nameMap))
	for k, v := range h.nameMap {
		result[k] = v
	}
	return result
}

// IsValidValue checks if a value is a valid enum value.
func (h *MideaIntEnumByteHelper[T]) IsValidValue(value byte) bool {
	_, ok := h.valueMap[value]
	return ok
}

// IsValidName checks if a name is a valid enum name.
func (h *MideaIntEnumByteHelper[T]) IsValidName(name string) bool {
	_, ok := h.nameMap[name]
	return ok
}

// MideaIntEnumHelper provides helper methods for enum types.
// It stores enum values and provides List(), GetFromValue(), and GetFromName() methods.
//
// Usage:
//
//	type FanSpeed byte
//	var fanSpeedHelper = NewMideaIntEnumHelper("FanSpeed", FanSpeedDefault).
//		WithValues(FanSpeedL1, FanSpeedL2, FanSpeedAuto)
//
//	// Or register with name mapping for GetFromName support:
//	var fanSpeedHelper = NewMideaIntEnumHelper("FanSpeed", FanSpeedDefault).
//		WithNamedValues(map[string]FanSpeed{"L1": FanSpeedL1, "AUTO": FanSpeedAuto})
type MideaIntEnumHelper[T MideaIntEnum] struct {
	enumName  string
	defaultValue T
	values    []T
	nameMap   map[string]T
	valueMap  map[int]T
}

// NewMideaIntEnumHelper creates a new helper for an enum type.
func NewMideaIntEnumHelper[T MideaIntEnum](enumName string, defaultValue T) *MideaIntEnumHelper[T] {
	return &MideaIntEnumHelper[T]{
		enumName:     enumName,
		defaultValue: defaultValue,
		values:       nil,
		nameMap:      make(map[string]T),
		valueMap:     make(map[int]T),
	}
}

// WithValues registers enum values for List() and GetFromValue() support.
func (h *MideaIntEnumHelper[T]) WithValues(values ...T) *MideaIntEnumHelper[T] {
	h.values = values
	h.valueMap = make(map[int]T, len(values))
	for _, v := range values {
		h.valueMap[v.Value()] = v
	}
	return h
}

// WithNamedValues registers named enum values for GetFromName() support.
func (h *MideaIntEnumHelper[T]) WithNamedValues(nameMap map[string]T) *MideaIntEnumHelper[T] {
	h.nameMap = nameMap
	return h
}

// List returns all registered enum values.
// In Python: @classmethod def list(cls) -> list[MideaIntEnum]
func (h *MideaIntEnumHelper[T]) List() []T {
	if h.values == nil {
		return []T{}
	}
	result := make([]T, len(h.values))
	copy(result, h.values)
	return result
}

// GetFromValue gets an enum value from its integer value.
// Returns the default value if the value is not found.
// In Python: @classmethod def get_from_value(cls, value: Optional[int], default: Optional[MideaIntEnum] = None)
func (h *MideaIntEnumHelper[T]) GetFromValue(value int) T {
	if v, ok := h.valueMap[value]; ok {
		return v
	}
	slog.Debug("Unknown enum value", "enum", h.enumName, "value", value)
	return h.defaultValue
}

// GetFromValueWithDefault gets an enum value from its integer value with a custom default.
func (h *MideaIntEnumHelper[T]) GetFromValueWithDefault(value int, default_ T) T {
	if v, ok := h.valueMap[value]; ok {
		return v
	}
	slog.Debug("Unknown enum value", "enum", h.enumName, "value", value)
	return default_
}

// GetFromName gets an enum value from its string name.
// Returns the default value if the name is not found.
// In Python: @classmethod def get_from_name(cls, name: Optional[str], default: Optional[MideaIntEnum] = None)
func (h *MideaIntEnumHelper[T]) GetFromName(name string) T {
	if v, ok := h.nameMap[name]; ok {
		return v
	}
	slog.Debug("Unknown enum name", "enum", h.enumName, "name", name)
	return h.defaultValue
}

// GetFromNameWithDefault gets an enum value from its string name with a custom default.
func (h *MideaIntEnumHelper[T]) GetFromNameWithDefault(name string, default_ T) T {
	if v, ok := h.nameMap[name]; ok {
		return v
	}
	slog.Debug("Unknown enum name", "enum", h.enumName, "name", name)
	return default_
}

// GetNameMap returns the name-to-value mapping for this enum.
func (h *MideaIntEnumHelper[T]) GetNameMap() map[string]T {
	result := make(map[string]T, len(h.nameMap))
	for k, v := range h.nameMap {
		result[k] = v
	}
	return result
}

// IsValidValue checks if a value is a valid enum value.
func (h *MideaIntEnumHelper[T]) IsValidValue(value int) bool {
	_, ok := h.valueMap[value]
	return ok
}

// IsValidName checks if a name is a valid enum name.
func (h *MideaIntEnumHelper[T]) IsValidName(name string) bool {
	_, ok := h.nameMap[name]
	return ok
}

// Deprecated marks a function as deprecated and recommends a replacement.
// In Go, we don't have decorators like Python, but we can use build tags or comments.
// The standard way is to add a comment:
//
//	Deprecated: Use 'replacement' instead. msg
//
// This function logs a deprecation warning and is provided for runtime checks.
func Deprecated(funcName, replacement, msg string) {
	if msg == "" {
		msg = "Please use '" + replacement + "' instead."
	}
	slog.Warn("function is deprecated", "function", funcName, "message", msg)
}

// DeprecatedWithWarning logs a deprecation warning only once per function.
// This mimics Python's @deprecated decorator behavior.
type DeprecatedWithWarning struct {
	warned map[string]bool
}

// NewDeprecatedWithWarning creates a new DeprecatedWithWarning instance.
func NewDeprecatedWithWarning() *DeprecatedWithWarning {
	return &DeprecatedWithWarning{
		warned: make(map[string]bool),
	}
}

// Check logs a deprecation warning if not already warned for this function.
func (d *DeprecatedWithWarning) Check(funcName, replacement, msg string) {
	if d.warned[funcName] {
		return
	}

	d.warned[funcName] = true

	if msg == "" {
		msg = "Please use '" + replacement + "' instead."
	}
	slog.Warn("function is deprecated", "function", funcName, "message", msg)
}
