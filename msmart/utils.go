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

type NumTypeT interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}

func EnumFromInt[E NumTypeT](val int, enumList []E, defaultEnum E) E {
	for _, e := range enumList {
		if int(e) == val {
			return e
		}
	}
	return defaultEnum
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
type MideaIntEnum[VT NumTypeT] interface {
	Value() VT
	String() string
}

// MideaNumEnumByteHelper provides helper methods for byte-backed enum types.
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
type MideaNumEnumByteHelper[T MideaIntEnum[VT], VT NumTypeT] struct {
	enumName     string
	defaultValue T
	values       []T
	nameMap      map[string]T
	valueMap     map[VT]T
}

// NewMideaNumEnumByteHelper creates a new helper for a byte-backed enum type.
func NewMideaNumEnumByteHelper[T MideaIntEnum[VT], VT NumTypeT](enumName string, defaultValue T) *MideaNumEnumByteHelper[T, VT] {
	return &MideaNumEnumByteHelper[T, VT]{
		enumName:     enumName,
		defaultValue: defaultValue,
		values:       nil,
		nameMap:      make(map[string]T),
		valueMap:     make(map[VT]T),
	}
}

// WithValues registers enum values for List() and GetFromValue() support.
func (h *MideaNumEnumByteHelper[T, VT]) WithValues(values ...T) *MideaNumEnumByteHelper[T, VT] {
	h.values = values
	h.valueMap = make(map[VT]T, len(values))
	for _, v := range values {
		h.valueMap[v.Value()] = v
	}
	return h
}

// WithNamedValues registers named enum values for GetFromName() support.
func (h *MideaNumEnumByteHelper[T, VT]) WithNamedValues(nameMap map[string]T) *MideaNumEnumByteHelper[T, VT] {
	h.nameMap = nameMap
	return h
}

// List returns all registered enum values.
func (h *MideaNumEnumByteHelper[T, VT]) List() []T {
	if h.values == nil {
		return []T{}
	}
	result := make([]T, len(h.values))
	copy(result, h.values)
	return result
}

// GetFromValue gets an enum value from its byte value.
// Returns the default value if the value is not found.
func (h *MideaNumEnumByteHelper[T, VT]) GetFromValue(value VT) T {
	if v, ok := h.valueMap[value]; ok {
		return v
	}
	return h.defaultValue
}

// GetFromName gets an enum value from its string name.
// Returns the default value if the name is not found.
func (h *MideaNumEnumByteHelper[T, VT]) GetFromName(name string) T {
	if v, ok := h.nameMap[name]; ok {
		return v
	}
	return h.defaultValue
}

// IsValidValue checks if a value is a valid enum value.
func (h *MideaNumEnumByteHelper[T, VT]) IsValidValue(value VT) bool {
	_, ok := h.valueMap[value]
	return ok
}

// IsValidName checks if a name is a valid enum name.
func (h *MideaNumEnumByteHelper[T, VT]) IsValidName(name string) bool {
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
