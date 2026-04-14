// Package msmart provides utility classes and methods for Midea AC.
package msmart

import (
	"log/slog"
)

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
type MideaIntEnum interface {
	Value() int
	String() string
}

// MideaIntEnumHelper provides helper methods for enum types.
// In Go, we can't add methods to all enum types, so this is a utility struct.
type MideaIntEnumHelper struct {
	enumName string
}

// NewMideaIntEnumHelper creates a new helper for an enum type.
func NewMideaIntEnumHelper(enumName string) *MideaIntEnumHelper {
	return &MideaIntEnumHelper{enumName: enumName}
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
