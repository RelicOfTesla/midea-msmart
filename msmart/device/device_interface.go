// Package device provides the Device interface for all Midea smart devices
package device

import (
	"context"
	"time"
)

// Device defines the base interface for all Midea smart devices
// This interface represents the common functionality shared by all device types
type Device interface {
	// Device identification
	GetIP() string
	GetPort() int
	GetID() string
	GetType() DeviceType
	GetName() string
	GetSN() string
	GetVersion() int

	// Status
	GetOnline() bool
	GetSupported() bool

	// Core operations
	Refresh(ctx context.Context) error
	Apply(ctx context.Context) error
	GetCapabilities(ctx context.Context) error

	// Capabilities
	CapabilitiesDict() map[string]interface{}
	ToDict() map[string]interface{}

	// Configuration
	SetMaxConnectionLifetime(seconds *int)
}

type DeviceAuthV3 interface {
	Device

	// V3
	IsAuthenticated() bool
	AuthenticateFromPreset(ctx context.Context) (bool, error)
	Authenticate(ctx context.Context, token Token, key Key) error

	GetKeyInfo() (Token, Key, LocalKey, time.Time)
}

type SimpleDevice interface {
	Device
	SetPowerState(bool)
	PowerState() bool
}
