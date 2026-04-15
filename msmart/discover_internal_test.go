// Package msmart provides internal tests for discovery functionality.
// This file tests unexported functions that cannot be tested from external packages.
// Translated from msmart-ng Python test_discover.py TestDiscover class.
package msmart

import (
	"testing"
)

// ============================================================================
// Test Data (from Python test_discover.py)
// ============================================================================

// discoverResponsesV2 is a V2 discovery response for device at 10.100.1.140
// Python: bytes.fromhex("5a5a011178007a8000000000000000000000000060ca0000000e0000000000000000000001000000c08651cb1b88a167bdcf7d37534ef81312d39429bf9b2673f200b635fae369a560fa9655eab8344be22b1e3b024ef5dfd392dc3db64dbffb6a66fb9cd5ec87a78000cd9043833b9f76991e8af29f3496")
var discoverResponseV2 = mustHexDecodeInternal("5a5a011178007a8000000000000000000000000060ca0000000e0000000000000000000001000000c08651cb1b88a167bdcf7d37534ef81312d39429bf9b2673f200b635fae369a560fa9655eab8344be22b1e3b024ef5dfd392dc3db64dbffb6a66fb9cd5ec87a78000cd9043833b9f76991e8af29f3496")

// discoverResponsesV3 is a V3 discovery response for device at 10.100.1.239
// Python: bytes.fromhex("837000c8200f00005a5a0111b8007a800000000061433702060817143daa00000086000000000000000001800000000041c7129527bc03ee009284a90c2fbd2f179764ac35b55e7fb0e4ab0de9298fa1a5ca328046c603fb1ab60079d550d03546b605180127fdb5bb33a105f5206b5f008bffba2bae272aa0c96d56b45c4afa33f826a0a4215d1dd87956a267d2dbd34bdfb3e16e33d88768cc4c3d0658937d0bb19369bf0317b24d3a4de9e6a13106f7ceb5acc6651ce53d684a32ce34dc3a4fbe0d4139de99cc88a0285e14657045")
var discoverResponseV3 = mustHexDecodeInternal("837000c8200f00005a5a0111b8007a800000000061433702060817143daa00000086000000000000000001800000000041c7129527bc03ee009284a90c2fbd2f179764ac35b55e7fb0e4ab0de9298fa1a5ca328046c603fb1ab60079d550d03546b605180127fdb5bb33a105f5206b5f008bffba2bae272aa0c96d56b45c4afa33f826a0a4215d1dd87956a267d2dbd34bdfb3e16e33d88768cc4c3d0658937d0bb19369bf0317b24d3a4de9e6a13106f7ceb5acc6651ce53d684a32ce34dc3a4fbe0d4139de99cc88a0285e14657045")

// mustHexDecodeInternal decodes a hex string, panicking on error
func mustHexDecodeInternal(s string) []byte {
	data := make([]byte, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		var b byte
		for _, c := range s[i : i+2] {
			b <<= 4
			switch {
			case c >= '0' && c <= '9':
				b |= byte(c - '0')
			case c >= 'a' && c <= 'f':
				b |= byte(c - 'a' + 10)
			case c >= 'A' && c <= 'F':
				b |= byte(c - 'A' + 10)
			}
		}
		data[i/2] = b
	}
	return data
}

// ============================================================================
// TestGetDeviceVersion - Device Version Detection Tests
// ============================================================================

// TestGetDeviceVersion_V2 tests that we can detect a V2 device version.
// This is a translation of Python's TestDiscover.test_discover_v2 version check.
func TestGetDeviceVersion_V2(t *testing.T) {
	// Check version
	version, err := getDeviceVersion(discoverResponseV2)
	if err != nil {
		t.Fatalf("getDeviceVersion failed: %v", err)
	}

	// Python: self.assertEqual(version, 2)
	if version != 2 {
		t.Errorf("Expected version 2, got %d", version)
	}
}

// TestGetDeviceVersion_V3 tests that we can detect a V3 device version.
// This is a translation of Python's TestDiscover.test_discover_v3 version check.
func TestGetDeviceVersion_V3(t *testing.T) {
	// Check version
	version, err := getDeviceVersion(discoverResponseV3)
	if err != nil {
		t.Fatalf("getDeviceVersion failed: %v", err)
	}

	// Python: self.assertEqual(version, 3)
	if version != 3 {
		t.Errorf("Expected version 3, got %d", version)
	}
}

// TestGetDeviceVersion_TooShort tests error handling for short data.
func TestGetDeviceVersion_TooShort(t *testing.T) {
	// Data too short
	shortData := []byte{0x5a}
	_, err := getDeviceVersion(shortData)
	if err == nil {
		t.Error("Expected error for data too short, got nil")
	}
}

// TestGetDeviceVersion_Unknown tests error handling for unknown version.
func TestGetDeviceVersion_Unknown(t *testing.T) {
	// Unknown version (starts with different bytes)
	unknownData := []byte{0x00, 0x00, 0x01, 0x02}
	_, err := getDeviceVersion(unknownData)
	if err == nil {
		t.Error("Expected error for unknown version, got nil")
	}
}

// ============================================================================
// TestGetDeviceInfo - Device Info Extraction Tests
// ============================================================================

// TestGetDeviceInfo_V2 tests that we can extract V2 device info.
// This is a translation of Python's TestDiscover.test_discover_v2 info extraction.
func TestGetDeviceInfo_V2(t *testing.T) {
	host := "10.100.1.140"

	// Check info matches
	info, err := getDeviceInfo(host, 2, discoverResponseV2)
	if err != nil {
		t.Fatalf("getDeviceInfo failed: %v", err)
	}

	if info == nil {
		t.Fatal("Expected info to be non-nil")
	}

	// Python: self.assertEqual(info["ip"], HOST[0])
	if info.IP != host {
		t.Errorf("Expected IP %s, got %s", host, info.IP)
	}

	// Python: self.assertEqual(info["port"], 6444)
	if info.Port != 6444 {
		t.Errorf("Expected port 6444, got %d", info.Port)
	}

	// Python: self.assertEqual(info["device_id"], 15393162840672)
	if info.DeviceID != 15393162840672 {
		t.Errorf("Expected device ID 15393162840672, got %d", info.DeviceID)
	}

	// Python: self.assertEqual(info["device_type"], DeviceType.AIR_CONDITIONER)
	if info.DeviceType != DeviceTypeAirConditioner {
		t.Errorf("Expected device type 0x%02X, got 0x%02X", DeviceTypeAirConditioner, info.DeviceType)
	}

	// Python: self.assertEqual(info["name"], "net_ac_F7B4")
	if info.Name != "net_ac_F7B4" {
		t.Errorf("Expected name 'net_ac_F7B4', got '%s'", info.Name)
	}

	// Python: self.assertEqual(info["sn"], "000000P0000000Q1F0C9D153F7B40000")
	if info.SN != "000000P0000000Q1F0C9D153F7B40000" {
		t.Errorf("Expected SN '000000P0000000Q1F0C9D153F7B40000', got '%s'", info.SN)
	}
}

// TestGetDeviceInfo_V3 tests that we can extract V3 device info.
// This is a translation of Python's TestDiscover.test_discover_v3 info extraction.
func TestGetDeviceInfo_V3(t *testing.T) {
	host := "10.100.1.239"

	// Check info matches
	info, err := getDeviceInfo(host, 3, discoverResponseV3)
	if err != nil {
		t.Fatalf("getDeviceInfo failed: %v", err)
	}

	if info == nil {
		t.Fatal("Expected info to be non-nil")
	}

	// Python: self.assertEqual(info["ip"], HOST[0])
	if info.IP != host {
		t.Errorf("Expected IP %s, got %s", host, info.IP)
	}

	// Python: self.assertEqual(info["port"], 6444)
	if info.Port != 6444 {
		t.Errorf("Expected port 6444, got %d", info.Port)
	}

	// Python: self.assertEqual(info["device_id"], 147334558165565)
	if info.DeviceID != 147334558165565 {
		t.Errorf("Expected device ID 147334558165565, got %d", info.DeviceID)
	}

	// Python: self.assertEqual(info["device_type"], DeviceType.AIR_CONDITIONER)
	if info.DeviceType != DeviceTypeAirConditioner {
		t.Errorf("Expected device type 0x%02X, got 0x%02X", DeviceTypeAirConditioner, info.DeviceType)
	}

	// Python: self.assertEqual(info["name"], "net_ac_63BA")
	if info.Name != "net_ac_63BA" {
		t.Errorf("Expected name 'net_ac_63BA', got '%s'", info.Name)
	}

	// Python: self.assertEqual(info["sn"], "000000P0000000Q1B88C29C963BA0000")
	if info.SN != "000000P0000000Q1B88C29C963BA0000" {
		t.Errorf("Expected SN '000000P0000000Q1B88C29C963BA0000', got '%s'", info.SN)
	}
}

// TestGetDeviceInfo_V3Version tests that version is set correctly for V3.
func TestGetDeviceInfo_V3Version(t *testing.T) {
	host := "10.100.1.239"

	info, err := getDeviceInfo(host, 3, discoverResponseV3)
	if err != nil {
		t.Fatalf("getDeviceInfo failed: %v", err)
	}

	if info.Version != 3 {
		t.Errorf("Expected version 3, got %d", info.Version)
	}
}

// TestGetDeviceInfo_V2Version tests that version is set correctly for V2.
func TestGetDeviceInfo_V2Version(t *testing.T) {
	host := "10.100.1.140"

	info, err := getDeviceInfo(host, 2, discoverResponseV2)
	if err != nil {
		t.Fatalf("getDeviceInfo failed: %v", err)
	}

	if info.Version != 2 {
		t.Errorf("Expected version 2, got %d", info.Version)
	}
}

// ============================================================================
// TestDeviceConstruction - Device Construction from Discovery Tests
// ============================================================================

// TestDeviceConstruction_V2 tests that we can construct a device from V2 discovery.
// This is a translation of Python's TestDiscover.test_discover_v2 device construction.
func TestDeviceConstruction_V2(t *testing.T) {
	host := "10.100.1.140"

	// Get device info
	info, err := getDeviceInfo(host, 2, discoverResponseV2)
	if err != nil {
		t.Fatalf("getDeviceInfo failed: %v", err)
	}

	// Python: device = Device.construct(type=info["device_type"], **info)
	device := NewDevice(
		info.IP,
		info.Port,
		info.DeviceID,
		info.DeviceType,
		WithSN(info.SN),
		WithName(info.Name),
		WithVersion(info.Version),
	)

	if device == nil {
		t.Fatal("Expected device to be non-nil")
	}

	// Python: self.assertIsNotNone(device)
	// Python: self.assertIsInstance(device, AC) - In Go, we check device type
	if device.GetType() != DeviceTypeAirConditioner {
		t.Errorf("Expected device type 0x%02X, got 0x%02X", DeviceTypeAirConditioner, device.GetType())
	}

	// Python: self.assertEqual(device.version, 2)
	deviceVersion := device.GetVersion()
	if deviceVersion == nil || *deviceVersion != 2 {
		t.Errorf("Expected version 2, got %v", deviceVersion)
	}
}

// TestDeviceConstruction_V3 tests that we can construct a device from V3 discovery.
// This is a translation of Python's TestDiscover.test_discover_v3 device construction.
func TestDeviceConstruction_V3(t *testing.T) {
	host := "10.100.1.239"

	// Get device info
	info, err := getDeviceInfo(host, 3, discoverResponseV3)
	if err != nil {
		t.Fatalf("getDeviceInfo failed: %v", err)
	}

	// Python: device = Device.construct(type=info["device_type"], **info)
	device := NewDevice(
		info.IP,
		info.Port,
		info.DeviceID,
		info.DeviceType,
		WithSN(info.SN),
		WithName(info.Name),
		WithVersion(info.Version),
	)

	if device == nil {
		t.Fatal("Expected device to be non-nil")
	}

	// Python: self.assertIsNotNone(device)
	// Python: self.assertIsInstance(device, AC) - In Go, we check device type
	if device.GetType() != DeviceTypeAirConditioner {
		t.Errorf("Expected device type 0x%02X, got 0x%02X", DeviceTypeAirConditioner, device.GetType())
	}

	// Python: self.assertEqual(device.version, 3)
	deviceVersion := device.GetVersion()
	if deviceVersion == nil || *deviceVersion != 3 {
		t.Errorf("Expected version 3, got %v", deviceVersion)
	}
}
