// Package config provides configuration management for the midea CLI.
package config

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Device represents a saved device configuration.
type Device struct {
	ID             string `json:"id"`                         // Device ID (decimal string)
	Name           string `json:"name"`                       // Alias name
	IP             string `json:"ip"`                         // IP address
	Port           int    `json:"port"`                       // Port number
	SN             string `json:"sn"`                         // Serial number
	Type           int    `json:"type"`                       // Device type (0xAC for AC)
	Token          string `json:"token"`                      // Authentication token (hex string)
	Key            string `json:"key"`                        // Authentication key (hex string)
	Version        int    `json:"version"`                    // Protocol version (2 or 3)
	Online         bool   `json:"online"`                     // Online status (from last discovery)
	LocalKey       string `json:"local_key,omitempty"`        // Local key for V3 devices (hex string)
	LocalKeyExpire string `json:"local_key_expire,omitempty"` // Local key expiration time (RFC3339)
}

// DeviceJSON is used for JSON unmarshaling with flexible type field.
type DeviceJSON struct {
	Device
	Type any `json:"type"` // Can be string or number
}

// parseType converts type field to int, handling both string and int formats.
func parseType(t interface{}) int {
	switch v := t.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case string:
		// Convert known type strings to type codes
		switch v {
		case "air_conditioner", "air_conditioner_ac":
			return 0xAC
		case "dishwasher":
			return 0xE1
		default:
			return 0
		}
	default:
		return 0
	}
}

// Config represents the CLI configuration.
type Config struct {
	Devices []Device `json:"devices"`
}

// DefaultConfigPath returns the default configuration file path.
// Priority: current working directory > executable directory > ~/.config/midea/
func DefaultConfigPath() string {
	// First, try current working directory
	cwd, err := os.Getwd()
	_ = err

	findList := []func() string{
		func() string {
			execPath, err := os.Executable()
			if err == nil {
				execDir := filepath.Dir(execPath)
				configPath := filepath.Join(execDir, "midea.json")
				// If config exists in executable directory, use it
				return configPath
			}
			return ""
		},
		func() string {
			if cwd == "" {
				return ""
			}
			return filepath.Join(cwd, "midea.json")
		},
		func() string {
			if cwd == "" {
				return ""
			}
			return filepath.Join(cwd, "midea-config.json")
		},
		func() string {
			configDir := filepath.Join(os.Getenv("HOME"), ".config", "midea")
			return filepath.Join(configDir, "config.json")
		},
	}
	fallbackPath := ""

	for _, find := range findList {
		path := find()
		if path != "" {
			if _, err := os.Stat(path); err == nil {
				return path
			}
			if fallbackPath == "" {
				fallbackPath = path
			}
		}
	}
	return fallbackPath
}

// Load loads the configuration from the specified path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &Config{Devices: []Device{}}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Use intermediate struct for flexible parsing
	var cfgJSON struct {
		Devices []DeviceJSON `json:"devices"`
	}
	if err := json.Unmarshal(data, &cfgJSON); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Convert to Config
	cfg := &Config{Devices: make([]Device, len(cfgJSON.Devices))}
	for i, d := range cfgJSON.Devices {
		cfg.Devices[i] = Device{
			ID:             d.ID,
			Name:           d.Name,
			IP:             d.IP,
			Port:           d.Port,
			SN:             d.SN,
			Type:           parseType(d.Type),
			Token:          d.Token,
			Key:            d.Key,
			Version:        d.Version,
			Online:         d.Online,
			LocalKey:       d.LocalKey,
			LocalKeyExpire: d.LocalKeyExpire,
		}
	}

	return cfg, nil
}

// Save saves the configuration to the specified path.
func (c *Config) Save(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetDevice finds a device by name, ID, SN, or IP.
func (c *Config) GetDevice(identifier string) *Device {
	for i := range c.Devices {
		d := &c.Devices[i]
		if d.Name == identifier || d.ID == identifier || d.SN == identifier || d.IP == identifier {
			return d
		}
	}
	return nil
}

// AddDevice adds a new device to the configuration.
func (c *Config) AddDevice(device Device) {
	// Check if device already exists (by ID)
	for i, d := range c.Devices {
		if d.ID == device.ID {
			// Update existing device
			c.Devices[i] = device
			return
		}
	}
	c.Devices = append(c.Devices, device)
}

// RemoveDevice removes a device by name, ID, SN, or IP.
func (c *Config) RemoveDevice(identifier string) bool {
	for i, d := range c.Devices {
		if d.Name == identifier || d.ID == identifier || d.SN == identifier || d.IP == identifier {
			c.Devices = append(c.Devices[:i], c.Devices[i+1:]...)
			return true
		}
	}
	return false
}

// BindName binds a name to an existing device.
func (c *Config) BindName(identifier, name string) bool {
	for i := range c.Devices {
		d := &c.Devices[i]
		if d.ID == identifier || d.SN == identifier || d.IP == identifier {
			d.Name = name
			return true
		}
	}
	return false
}

// ListDevices returns all saved devices.
func (c *Config) ListDevices() []Device {
	return c.Devices
}

// ////////////////
func (cfg *Device) GetValidKeys() (token []byte, key []byte, localKeyBytes []byte, expiration time.Time, err error) {
	if cfg.Token != "" {
		token, err = hex.DecodeString(cfg.Token)
		if err != nil {
			return nil, nil, nil, time.Time{}, fmt.Errorf("无效的Token: %w", err)
		}
	}
	if cfg.Key != "" {
		key, err = hex.DecodeString(cfg.Key)
		if err != nil {
			return nil, nil, nil, time.Time{}, fmt.Errorf("无效的Key: %w", err)
		}
	}

	if cfg.LocalKey != "" && cfg.LocalKeyExpire != "" {
		localKeyBytes, err = hex.DecodeString(cfg.LocalKey)
		if err == nil {
			expiration, err = time.Parse(time.RFC3339, cfg.LocalKeyExpire)
			if err != nil {
				return nil, nil, nil, time.Time{}, fmt.Errorf("无效的LocalKey过期时间: %w", err)
			}

		}
	}
	return
}
