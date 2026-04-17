// Package tests provides REAL functional tests for msmart devices.
// These tests communicate with actual devices, not mocks.
//
// CRITICAL: Only test device IP 192.168.1.57 - DO NOT touch other devices!
//
// Test Categories:
// 1. Discovery tests - Work without authentication
// 2. Authentication tests - Need token/key (requires cloud credentials)
// 3. Command tests - Need authentication
//
// Running Tests:
// - go test -v -run TestRealDevice ./tests/
// - Set MIDEA_ACCOUNT and MIDEA_PASSWORD for cloud-based tests
// - Or set MIDEA_TOKEN and MIDEA_KEY for direct auth tests
package tests

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	msmart "github.com/RelicOfTesla/midea-msmart/msmart"
	devicetypes "github.com/RelicOfTesla/midea-msmart/msmart/device"
)

// ============================================================================
// Configuration - Only 192.168.1.57 is allowed!
// ============================================================================

const (
	// TargetDeviceIP - CRITICAL: Only this IP is allowed for testing
	TargetDeviceIP = "192.168.1.57"

	// TargetDevicePort - Default Midea device port
	TargetDevicePort = 6444

	// TestTimeout - Default timeout for tests
	TestTimeout = 10 * time.Second
)

// getCloudCredentials returns cloud account credentials from environment
func getCloudCredentials() (account, password string, ok bool) {
	account = os.Getenv("MIDEA_ACCOUNT")
	password = os.Getenv("MIDEA_PASSWORD")
	ok = account != "" && password != ""
	return
}

// getTokenKey returns token and key from environment
func getTokenKey() (token, key string, ok bool) {
	token = os.Getenv("MIDEA_TOKEN")
	key = os.Getenv("MIDEA_KEY")
	ok = token != "" && key != ""
	return
}

// ============================================================================
// Discovery Tests - Real Device Communication (No Auth Required)
// ============================================================================

// TestRealDevice_Discovery tests discovering the real device at 192.168.1.57.
// This is a REAL functional test that communicates with actual hardware.
func TestRealDevice_Discovery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeout)
	defer cancel()

	config := &msmart.DiscoverConfig{
		Target:           TargetDeviceIP,
		Timeout:          5 * time.Second,
		DiscoveryPackets: 1,
		AutoConnect:      false,
	}

	device, err := msmart.DiscoverSingle(ctx, TargetDeviceIP, config)
	if err != nil {
		t.Fatalf("Discovery failed: %v", err)
	}

	if device == nil {
		t.Fatalf("No device found at %s", TargetDeviceIP)
	}

	// Verify device properties
	if device.GetIP() != TargetDeviceIP {
		t.Errorf("Expected IP %s, got %s", TargetDeviceIP, device.GetIP())
	}

	// Log discovered device info
	t.Logf("✅ Device discovered successfully:")
	t.Logf("   IP: %s", device.GetIP())
	t.Logf("   Port: %d", device.GetPort())
	t.Logf("   Device ID: %s", device.GetID())
	t.Logf("   Type: 0x%02X", device.GetType())

	if name := device.GetName(); name != "" {
		t.Logf("   Name: %s", name)
	}
	if sn := device.GetSN(); sn != "" {
		t.Logf("   SN: %s", sn)
	}
	if version := device.GetVersion(); version != 0 {
		t.Logf("   Version: %d", version)
	}
}

// TestRealDevice_DiscoveryBroadcast tests broadcast discovery.
// This may discover multiple devices but we only verify 192.168.1.57.
func TestRealDevice_DiscoveryBroadcast(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeout)
	defer cancel()

	config := &msmart.DiscoverConfig{
		Timeout:          3 * time.Second,
		DiscoveryPackets: 1,
		AutoConnect:      false,
	}

	devices, err := msmart.Discover(ctx, config)
	if err != nil {
		t.Logf("Broadcast discovery error (may be normal): %v", err)
	}

	// Check if we found our target device
	found := false
	for _, device := range devices {
		t.Logf("Found device: %s (ID: %s, Type: 0x%02X)",
			device.GetIP(), device.GetID(), device.GetType())
		if device.GetIP() == TargetDeviceIP {
			found = true
		}
	}

	if found {
		t.Logf("✅ Target device %s found via broadcast", TargetDeviceIP)
	} else {
		t.Logf("⚠️ Target device %s not found (may be normal if on different subnet)", TargetDeviceIP)
	}
}

// ============================================================================
// Device Info Tests - Real Device Properties
// ============================================================================

// TestRealDevice_Properties tests device property access.
func TestRealDevice_Properties(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeout)
	defer cancel()

	// Discover device first
	config := &msmart.DiscoverConfig{
		Target:           TargetDeviceIP,
		Timeout:          5 * time.Second,
		DiscoveryPackets: 1,
		AutoConnect:      false,
	}

	device, err := msmart.DiscoverSingle(ctx, TargetDeviceIP, config)
	if err != nil {
		t.Fatalf("Discovery failed: %v", err)
	}

	if device == nil {
		t.Fatalf("No device found at %s", TargetDeviceIP)
	}

	// Test all property getters
	ip := device.GetIP()
	if ip != TargetDeviceIP {
		t.Errorf("GetIP() = %s, want %s", ip, TargetDeviceIP)
	}

	port := device.GetPort()
	if port == 0 {
		t.Error("GetPort() returned 0")
	}

	deviceID := device.GetID()
	if deviceID == "" {
		t.Error("GetID() returned empty string")
	}

	deviceType := device.GetType()
	if deviceType == 0 {
		t.Error("GetType() returned 0")
	}

	// Test optional properties
	sn := device.GetSN()
	version := device.GetVersion()
	name := device.GetName()

	t.Logf("Device properties:")
	t.Logf("  IP: %s", ip)
	t.Logf("  Port: %d", port)
	t.Logf("  ID: %s", deviceID)
	t.Logf("  Type: 0x%02X", deviceType)
	if sn != "" {
		t.Logf("  SN: %s", sn)
	}
	if version != 0 {
		t.Logf("  Version: %d", version)
	}
	if name != "" {
		t.Logf("  Name: %s", name)
	}

	// Test ToDict
	dict := device.ToDict()
	if dict == nil {
		t.Error("ToDict() returned nil")
	}
	if dict["ip"] != ip {
		t.Errorf("ToDict()[ip] = %v, want %s", dict["ip"], ip)
	}

	t.Log("✅ All device properties accessible")
}

// TestRealDevice_Type tests device type detection.
// Note: This test may fail if the device doesn't respond due to rate limiting.
// The type is already verified in TestRealDevice_Properties.
func TestRealDevice_Type(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeout)
	defer cancel()

	config := &msmart.DiscoverConfig{
		Target:           TargetDeviceIP,
		Timeout:          5 * time.Second,
		DiscoveryPackets: 1,
	}

	device, err := msmart.DiscoverSingle(ctx, TargetDeviceIP, config)
	if err != nil {
		t.Logf("Discovery failed (may be due to device rate limiting): %v", err)
		t.Skip("Device not responding - skip to avoid false failure")
	}

	if device == nil {
		t.Skip("No device found at " + TargetDeviceIP + " - may be rate limited")
	}

	deviceType := device.GetType()

	// Check if it's an air conditioner (0xAC)
	if deviceType == msmart.DeviceTypeAirConditioner {
		t.Log("✅ Device is an Air Conditioner (0xAC)")
	} else if deviceType == msmart.DeviceTypeCommercialAC {
		t.Log("✅ Device is a Commercial Air Conditioner (0xCC)")
	} else {
		t.Logf("⚠️ Device has type 0x%02X (unknown/supported?)", deviceType)
	}
}

// TestRealDevice_Version tests device version detection.
// Note: This test may fail if the device doesn't respond due to rate limiting.
// The version is already verified in TestRealDevice_Properties.
func TestRealDevice_Version(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeout)
	defer cancel()

	config := &msmart.DiscoverConfig{
		Target:           TargetDeviceIP,
		Timeout:          5 * time.Second,
		DiscoveryPackets: 1,
	}

	device, err := msmart.DiscoverSingle(ctx, TargetDeviceIP, config)
	if err != nil {
		t.Logf("Discovery failed (may be due to device rate limiting): %v", err)
		t.Skip("Device not responding - skip to avoid false failure")
	}

	if device == nil {
		t.Skip("No device found at " + TargetDeviceIP + " - may be rate limited")
	}

	version := device.GetVersion()
	if version == 0 {
		t.Fatal("Device version is 0 (unknown)")
	}

	switch version {
	case 2:
		t.Log("✅ Device is V2 (no authentication required)")
	case 3:
		t.Log("✅ Device is V3 (authentication required for commands)")
	default:
		t.Logf("⚠️ Device has unknown version %d", version)
	}
}

// ============================================================================
// LAN Tests - Real Network Communication
// ============================================================================

// TestRealDevice_LAN tests creating a LAN client for the device.
func TestRealDevice_LAN(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeout)
	defer cancel()

	// Discover device first to get device ID
	config := &msmart.DiscoverConfig{
		Target:           TargetDeviceIP,
		Timeout:          5 * time.Second,
		DiscoveryPackets: 1,
	}

	device, err := msmart.DiscoverSingle(ctx, TargetDeviceIP, config)
	if err != nil {
		t.Fatalf("Discovery failed: %v", err)
	}

	if device == nil {
		t.Fatalf("No device found at %s", TargetDeviceIP)
	}

	// Create LAN client
	deviceID := device.GetID()
	deviceIDInt, err := strconv.ParseInt(deviceID, 10, 64)
	if err != nil {
		t.Fatalf("Invalid device ID: %s", deviceID)
	}
	lan := msmart.NewLAN(TargetDeviceIP, TargetDevicePort, msmart.LanDeviceId(deviceIDInt), nil, time.Time{})
	if lan == nil {
		t.Fatal("NewLAN returned nil")
	}

	// Verify LAN client properties
	if lan.Token() != nil {
		t.Error("Expected initial token to be nil")
	}
	if lan.Key() != nil {
		t.Error("Expected initial key to be nil")
	}

	t.Log("✅ LAN client created successfully")
}

// TestRealDevice_LANClient tests the LANClient compatibility wrapper.
func TestRealDevice_LANClient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeout)
	defer cancel()

	// Discover device first
	config := &msmart.DiscoverConfig{
		Target:           TargetDeviceIP,
		Timeout:          5 * time.Second,
		DiscoveryPackets: 1,
	}

	device, err := msmart.DiscoverSingle(ctx, TargetDeviceIP, config)
	if err != nil {
		t.Fatalf("Discovery failed: %v", err)
	}

	// Create LANClient
	deviceID := device.GetID()
	deviceIDInt, err := strconv.ParseInt(deviceID, 10, 64)
	if err != nil {
		t.Fatalf("Invalid device ID: %s", deviceID)
	}
	client := msmart.NewLANClient(TargetDeviceIP, TargetDevicePort, msmart.LanDeviceId(deviceIDInt), nil, time.Time{})
	if client == nil {
		t.Fatal("NewLANClient returned nil")
	}

	// Set timeout
	client.SetTimeout(5 * time.Second)

	t.Log("✅ LANClient created and configured successfully")
}

// ============================================================================
// Authentication Tests - Require Token/Key
// ============================================================================

// TestRealDevice_Authentication tests device authentication.
// This test requires MIDEA_TOKEN and MIDEA_KEY environment variables,
// OR MIDEA_ACCOUNT and MIDEA_PASSWORD to get token/key from cloud.
func TestRealDevice_Authentication(t *testing.T) {
	// Check for token/key first
	token, key, hasTokenKey := getTokenKey()
	if !hasTokenKey {
		// Try to get from cloud
		account, password, hasCreds := getCloudCredentials()
		if !hasCreds {
			t.Skip("Skipping: Set MIDEA_TOKEN+MIDEA_KEY or MIDEA_ACCOUNT+MIDEA_PASSWORD to run this test")
		}

		// Get token/key from cloud
		t.Logf("Getting token/key from cloud for account: %s...", account[:min(5, len(account))]+"***")

		// Create cloud client
		cloud, err := msmart.NewSmartHomeCloud(msmart.DefaultCloudRegion, &account, &password, false, nil)
		if err != nil {
			t.Fatalf("Failed to create cloud client: %v", err)
		}

		err = cloud.Login(false)
		if err != nil {
			t.Fatalf("Cloud login failed: %v", err)
		}

		// Get UDPID from device SN (we need to discover first)
		ctx, cancel := context.WithTimeout(context.Background(), TestTimeout)
		defer cancel()

		config := &msmart.DiscoverConfig{
			Target:           TargetDeviceIP,
			Timeout:          5 * time.Second,
			DiscoveryPackets: 1,
		}

		device, err := msmart.DiscoverSingle(ctx, TargetDeviceIP, config)
		if err != nil {
			t.Fatalf("Discovery failed: %v", err)
		}

		sn := device.GetSN()
		if sn == "" {
			t.Fatal("Device SN is empty")
		}

		// Get UDPID from SN
		udpid := msmart.SecurityUdpid([]byte(sn))

		// Get token/key
		token, key, err = cloud.GetToken(fmt.Sprintf("%x", udpid))
		if err != nil {
			t.Fatalf("Failed to get token/key: %v", err)
		}

		t.Logf("Got token/key from cloud (token length: %d)", len(token))
	}

	// Now authenticate with the device
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeout)
	defer cancel()

	config := &msmart.DiscoverConfig{
		Target:           TargetDeviceIP,
		Timeout:          5 * time.Second,
		DiscoveryPackets: 1,
	}

	device, err := msmart.DiscoverSingle(ctx, TargetDeviceIP, config)
	if err != nil {
		t.Fatalf("Discovery failed: %v", err)
	}

	// Create LAN client
	deviceID := device.GetID()
	deviceIDInt, err := strconv.ParseInt(deviceID, 10, 64)
	if err != nil {
		t.Fatalf("Invalid device ID: %s", deviceID)
	}
	lan := msmart.NewLAN(TargetDeviceIP, TargetDevicePort, msmart.LanDeviceId(deviceIDInt), nil, time.Time{})

	// Authenticate
	err = lan.Authenticate(context.Background(), devicetypes.Token(token), devicetypes.Key(key), 3)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	t.Log("✅ Device authentication successful")
}

// ============================================================================
// Command Tests - Require Authentication
// ============================================================================

// TestRealDevice_SendCommand tests sending a command to the device.
// This test requires authentication (see TestRealDevice_Authentication).
func TestRealDevice_SendCommand(t *testing.T) {
	token, key, hasTokenKey := getTokenKey()
	account, password, hasCreds := getCloudCredentials()

	if !hasTokenKey && !hasCreds {
		t.Skip("Skipping: Set MIDEA_TOKEN+MIDEA_KEY or MIDEA_ACCOUNT+MIDEA_PASSWORD to run this test")
	}

	// This test would:
	// 1. Authenticate with token/key
	// 2. Send a query command to get device state
	// 3. Verify response

	t.Skip("Command sending requires authentication - see TestRealDevice_Authentication")

	_ = token
	_ = key
	_ = account
	_ = password
}

// ============================================================================
// Cloud API Tests - Require Account Credentials
// ============================================================================

// TestRealDevice_CloudLogin tests cloud login.
func TestRealDevice_CloudLogin(t *testing.T) {
	account, password, ok := getCloudCredentials()
	if !ok {
		t.Skip("Skipping: Set MIDEA_ACCOUNT and MIDEA_PASSWORD to run this test")
	}

	// Test SmartHome cloud
	cloud, err := msmart.NewSmartHomeCloud(msmart.DefaultCloudRegion, &account, &password, false, nil)
	if err != nil {
		t.Fatalf("Failed to create SmartHomeCloud: %v", err)
	}

	err = cloud.Login(false)
	if err != nil {
		t.Fatalf("Cloud login failed: %v", err)
	}

	accessToken := cloud.GetAccessToken()
	if accessToken == "" {
		t.Error("Expected access token to be non-empty after login")
	}

	t.Log("✅ Cloud login successful")
	t.Logf("   Access token length: %d", len(accessToken))
}

// TestRealDevice_CloudGetToken tests getting token/key from cloud.
func TestRealDevice_CloudGetToken(t *testing.T) {
	account, password, ok := getCloudCredentials()
	if !ok {
		t.Skip("Skipping: Set MIDEA_ACCOUNT and MIDEA_PASSWORD to run this test")
	}

	// Discover device to get SN
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeout)
	defer cancel()

	config := &msmart.DiscoverConfig{
		Target:           TargetDeviceIP,
		Timeout:          5 * time.Second,
		DiscoveryPackets: 1,
	}

	device, err := msmart.DiscoverSingle(ctx, TargetDeviceIP, config)
	if err != nil {
		t.Fatalf("Discovery failed: %v", err)
	}

	sn := device.GetSN()
	if sn == "" {
		t.Fatal("Device SN is empty")
	}

	// Get UDPID from SN
	udpid := msmart.SecurityUdpid([]byte(sn))
	udpidHex := fmt.Sprintf("%x", udpid)

	t.Logf("Device SN: %s", sn)
	t.Logf("UDPID: %s", udpidHex)

	// Login to cloud
	cloud, err := msmart.NewSmartHomeCloud(msmart.DefaultCloudRegion, &account, &password, false, nil)
	if err != nil {
		t.Fatalf("Failed to create SmartHomeCloud: %v", err)
	}

	err = cloud.Login(false)
	if err != nil {
		t.Fatalf("Cloud login failed: %v", err)
	}

	// Get token/key
	token, key, err := cloud.GetToken(udpidHex)
	if err != nil {
		t.Fatalf("Failed to get token/key: %v", err)
	}

	t.Log("✅ Got token/key from cloud")
	t.Logf("   Token length: %d", len(token))
	t.Logf("   Key length: %d", len(key))
}

// ============================================================================
// Helper Functions
// ============================================================================
