// Package msmart provides device discovery functionality for Midea AC devices.
package msmart

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// DiscoverError represents a discovery error
type DiscoverError struct {
	Message string
	Cause   error
}

func (e *DiscoverError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("discover error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("discover error: %s", e.Message)
}

func (e *DiscoverError) Unwrap() error {
	return e.Cause
}

// NewDiscoverError creates a new DiscoverError
func NewDiscoverError(message string, cause error) *DiscoverError {
	return &DiscoverError{Message: message, Cause: cause}
}

// IPv4Broadcast is the IPv4 broadcast address
const IPv4Broadcast = "255.255.255.255"

// DefaultDiscoveryTimeout is the default discovery timeout
const DefaultDiscoveryTimeout = 5 * time.Second

// DefaultDiscoveryPackets is the default number of discovery packets to send
const DefaultDiscoveryPackets = 3

// DiscoveryPorts are the ports used for discovery
var DiscoveryPorts = []int{6445, 20086}

// getBroadcastAddresses returns a list of broadcast addresses for all active network interfaces
// If no interfaces are found, returns the default 255.255.255.255
func getBroadcastAddresses() []string {
	var addresses []string

	interfaces, err := net.Interfaces()
	if err != nil {
		slog.Warn("Failed to get network interfaces", "error", err)
		return []string{IPv4Broadcast}
	}

	for _, iface := range interfaces {
		// Skip loopback and non-up interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			// Only process IPv4 addresses
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			// Convert to 4-byte representation
			ip := ipNet.IP.To4()
			if ip == nil {
				continue // Not IPv4
			}

			mask := ipNet.Mask
			if len(mask) != 4 {
				// Convert 16-byte mask to 4-byte mask
				if len(mask) == 16 {
					mask = mask[12:16]
				} else {
					continue
				}
			}

			// Calculate broadcast address: IP | (^Mask)
			broadcast := make(net.IP, 4)
			for i := 0; i < 4; i++ {
				broadcast[i] = ip[i] | ^mask[i]
			}

			broadcastStr := broadcast.String()
			verboseLog("Found broadcast address %s for interface %s (%s)", broadcastStr, iface.Name, ip)
			addresses = append(addresses, broadcastStr)
		}
	}

	// If no interfaces found, use the default broadcast address
	if len(addresses) == 0 {
		slog.Warn("No suitable network interfaces found, using default broadcast address")
		return []string{IPv4Broadcast}
	}

	return addresses
}

// DeviceInfo represents discovered device information
type DeviceInfo struct {
	IP         string
	Port       int
	DeviceID   int
	Name       string
	SN         string
	DeviceType DeviceType
	Version    int
}

// DiscoverConfig holds configuration for device discovery
type DiscoverConfig struct {
	// Target is the target address for discovery (default: 255.255.255.255)
	Target string
	// Timeout is the discovery timeout (default: 5 seconds)
	Timeout time.Duration
	// DiscoveryPackets is the number of discovery packets to send (default: 3)
	DiscoveryPackets int
	// Interface is the network interface to use for discovery
	Interface string
	// Region is the cloud region
	Region string
	// Account is the cloud account
	Account string
	// Password is the cloud password
	Password string
	// AutoConnect indicates whether to automatically connect and authenticate devices
	AutoConnect bool
	// ExistingToken is pre-existing token for V3 devices (to skip cloud auth)
	ExistingToken []byte
	// ExistingKey is pre-existing key for V3 devices (to skip cloud auth)
	ExistingKey []byte
}

// DiscoverResult represents a discovered device result
type DiscoverResult struct {
	Device *Device
	Error  error
}

// Discover discovers Midea smart devices on the local network
func Discover(ctx context.Context, config *DiscoverConfig) ([]*Device, error) {
	// Set default values
	if config == nil {
		config = &DiscoverConfig{}
	}
	if config.Target == "" {
		config.Target = IPv4Broadcast
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultDiscoveryTimeout
	}
	if config.DiscoveryPackets == 0 {
		config.DiscoveryPackets = DefaultDiscoveryPackets
	}
	if config.Region == "" {
		config.Region = DefaultCloudRegion
	}

	// Create a packet connection for UDP
	conn, err := net.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		return nil, NewDiscoverError("failed to create UDP socket", err)
	}
	defer conn.Close()

	// Set broadcast option
	if config.Target == IPv4Broadcast {
		if syscallConn, ok := conn.(syscall.Conn); ok {
			rawConn, err := syscallConn.SyscallConn()
			if err != nil {
				slog.Warn("Failed to get syscall connection", "error", err)
			} else {
				var setsockoptErr error
				rawConn.Control(func(fd uintptr) {
					setsockoptErr = setBroadcastOption(fd)
				})
				if setsockoptErr != nil {
					slog.Warn("Failed to set SO_BROADCAST option", "error", setsockoptErr)
				} else {
					verboseLog("SO_BROADCAST option set successfully")
				}
			}
		} else {
			slog.Warn("Connection does not implement syscall.Conn")
		}
	}

	// Channel to collect discovered devices
	results := make(chan *DiscoverResult, 100)

	// Map to track discovered IPs (to avoid duplicates)
	discoveredIPs := make(map[string]bool)
	var discoveredIPsMu sync.Mutex

	// WaitGroup to track goroutines
	var wg sync.WaitGroup

	// Create a cancelable context for the response handler
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start a goroutine to handle responses
	wg.Add(1)
	go func() {
		defer wg.Done()
		handleDiscoveryResponses(ctx, conn, discoveredIPs, &discoveredIPsMu, config, results)
	}()

	// Send discovery messages
	// Check if target is a broadcast address or a specific IP
	var targetAddresses []string
	if config.Target == IPv4Broadcast {
		// Get broadcast addresses for all network interfaces
		targetAddresses = getBroadcastAddresses()
		verboseLog("Sending discovery to %d broadcast addresses: %v", len(targetAddresses), targetAddresses)
	} else {
		// Use the specific target address
		targetAddresses = []string{config.Target}
		verboseLog("Sending discovery to specific target: %s", config.Target)
	}

	for _, port := range DiscoveryPorts {
		for _, targetAddr := range targetAddresses {
			addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", targetAddr, port))
			if err != nil {
				slog.Warn("Failed to resolve address", "address", targetAddr, "port", port, "error", err)
				continue
			}

			verboseLog("Discovery sent to %s:%d.", targetAddr, port)

			for i := 0; i < config.DiscoveryPackets; i++ {
				if _, err := conn.WriteTo(DiscoveryMsg, addr); err != nil {
					slog.Warn("Failed to send discovery", "address", targetAddr, "port", port, "error", err)
				}
			}
		}
	}

	// Wait for timeout
	select {
	case <-time.After(config.Timeout):
		verboseLog("Discovery timeout after %s", config.Timeout)
		// Cancel the context to stop the response handler
		cancel()
	case <-ctx.Done():
		verboseLog("Discovery cancelled")
	}

	verboseLog("Closing connection and waiting for response handler...")

	// Close the connection to stop the receive goroutine
	conn.Close()

	// Wait for the response handler to finish
	wg.Wait()

	verboseLog("Response handler finished, collecting results...")

	// Close the results channel
	close(results)

	// Collect results
	var devices []*Device
	var errs []error

	for result := range results {
		if result.Error != nil {
			errs = append(errs, result.Error)
		} else if result.Device != nil {
			devices = append(devices, result.Device)
		}
	}

	verboseLog("Discovered %d devices.", len(devices))

	// Return first error if any
	if len(errs) > 0 {
		return devices, errs[0]
	}

	return devices, nil
}

// handleDiscoveryResponses handles discovery responses
func handleDiscoveryResponses(ctx context.Context, conn net.PacketConn, discoveredIPs map[string]bool, discoveredIPsMu *sync.Mutex, config *DiscoverConfig, results chan<- *DiscoverResult) {
	buf := make([]byte, 4096)

	for {
		// Set read deadline
		if err := conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
			return
		}

		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Check if context is cancelled before continuing
				select {
				case <-ctx.Done():
					return
				default:
					continue
				}
			}
			// Connection closed or other error
			return
		}

		// Check if context is cancelled after successful read
		select {
		case <-ctx.Done():
			return
		default:
		}

		if n == 0 {
			continue
		}

		// Get the IP address
		ip, _, err := net.SplitHostPort(addr.String())
		if err != nil {
			continue
		}

		// Check if already discovered
		discoveredIPsMu.Lock()
		if discoveredIPs[ip] {
			discoveredIPsMu.Unlock()
			continue
		}
		discoveredIPs[ip] = true
		discoveredIPsMu.Unlock()

		verboseLog("Discovery response from %s: %s", ip, hex.EncodeToString(buf[:n]))

		// Process the response in a goroutine
		go func(ip string, data []byte) {
			device, err := processDiscoveryResponse(ip, data, config)
			results <- &DiscoverResult{Device: device, Error: err}
		}(ip, buf[:n])
	}
}

// processDiscoveryResponse processes a discovery response
func processDiscoveryResponse(ip string, data []byte, config *DiscoverConfig) (*Device, error) {
	// Get device version
	version, err := getDeviceVersion(data)
	if err != nil {
		return nil, err
	}

	verboseLog("Device version for %s: %d", ip, version)

	// Get device info
	info, err := getDeviceInfo(ip, version, data)
	if err != nil {
		return nil, err
	}

	verboseLog("Device info for %s: %+v", ip, info)

	// Construct device with optional pre-set token/key for V3 devices
	opts := []DeviceOption{
		WithSN(info.SN),
		WithName(info.Name),
		WithVersion(info.Version),
	}

	// If existing token/key are provided (for V3 devices to skip cloud auth)
	if config.ExistingToken != nil && config.ExistingKey != nil {
		opts = append(opts, WithTokenKey(config.ExistingToken, config.ExistingKey))
	}

	device := NewDevice(
		info.IP,
		info.Port,
		info.DeviceID,
		info.DeviceType,
		opts...,
	)

	// Auto-connect if requested
	if config.AutoConnect {
		if err := ConnectDevice(device, config); err != nil {
			return nil, err
		}
	}

	return device, nil
}

// getDeviceVersion determines the device version from discovery response
func getDeviceVersion(data []byte) (int, error) {
	// Attempt to parse XML from V1 device
	var v1XML interface{}
	if err := xml.Unmarshal(data, &v1XML); err == nil {
		return 1, nil
	}

	// Use start of packet data to differentiate between V2 and V3
	if len(data) < 2 {
		return 0, NewDiscoverError("data too short", nil)
	}

	startOfPacket := data[:2]
	if bytes.Equal(startOfPacket, []byte{0x5a, 0x5a}) {
		return 2, nil
	} else if bytes.Equal(startOfPacket, []byte{0x83, 0x70}) {
		return 3, nil
	}

	return 0, NewDiscoverError("unknown device version", nil)
}

// getDeviceInfo extracts device information from discovery response
func getDeviceInfo(ip string, version int, data []byte) (*DeviceInfo, error) {
	// Version 1 devices
	if version == 1 {
		return getDeviceInfoV1(ip, data)
	}

	// Version 2 & 3 devices
	return getDeviceInfoV2V3(ip, version, data)
}

// getDeviceInfoV1 extracts device info for V1 devices
func getDeviceInfoV1(ip string, data []byte) (*DeviceInfo, error) {
	// Parse XML
	var root V1DeviceResponse
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, NewDiscoverError("failed to parse V1 XML", err)
	}

	// Find device element
	device := root.Body.Device
	if device == nil {
		return nil, NewDiscoverError("could not find 'body/device' in XML", nil)
	}

	// Parse port (not used since V1 is not supported)
	_, err := strconv.Atoi(device.Port)
	if err != nil {
		return nil, NewDiscoverError("invalid port", err)
	}

	// V1 devices require additional query
	// For now, return not implemented
	return nil, NewDiscoverError("V1 device not supported yet", nil)
}

// V1DeviceResponse represents V1 device XML response
type V1DeviceResponse struct {
	XMLName xml.Name `xml:"root"`
	Body    struct {
		Device *V1Device `xml:"device"`
	} `xml:"body"`
}

// V1Device represents V1 device element
type V1Device struct {
	Port string `xml:"port,attr"`
}

// getDeviceInfoV2V3 extracts device info for V2/V3 devices
func getDeviceInfoV2V3(ip string, version int, data []byte) (*DeviceInfo, error) {
	// Create a memory view of the data
	dataView := data

	// Strip V3 header and hash
	if version == 3 {
		if len(dataView) < 24 {
			return nil, NewDiscoverError("V3 data too short", nil)
		}
		dataView = dataView[8 : len(dataView)-16]
	}

	// Check minimum length
	if len(dataView) < 56 {
		return nil, NewDiscoverError("data too short for V2/V3", nil)
	}

	// Extract encrypted payload
	encryptedData := dataView[40 : len(dataView)-16]

	// Extract ID (6 bytes, little endian) - matches Python implementation
	// Python: device_id = int.from_bytes(data_mv[20:26], "little")
	if len(dataView) < 26 {
		return nil, NewDiscoverError("data too short for device ID", nil)
	}
	deviceID := int(binary.LittleEndian.Uint16(dataView[20:22]))
	deviceID |= int(binary.LittleEndian.Uint32(dataView[22:26])) << 16

	// Attempt to decrypt the packet
	decryptedData, err := SecurityDecryptAES(encryptedData)
	if err != nil {
		return nil, NewDiscoverError("failed to decrypt discovery response", err)
	}

	verboseLog("Decrypted data from %s: %s", ip, hex.EncodeToString(decryptedData))

	// Check minimum decrypted length
	if len(decryptedData) < 42 {
		return nil, NewDiscoverError("decrypted data too short", nil)
	}

	// Extract IP and port (reverse bytes for IP)
	ipBytes := make([]byte, 4)
	for i := 0; i < 4; i++ {
		ipBytes[i] = decryptedData[3-i]
	}
	ipAddress := net.IP(ipBytes).String()

	port := int(binary.LittleEndian.Uint16(decryptedData[4:6]))

	if ipAddress != ip {
		slog.Debug("Reported device IP does not match received IP, using received IP", "reported", ipAddress, "received", ip)
	}

	// Extract serial number
	if len(decryptedData) < 40 {
		return nil, NewDiscoverError("decrypted data too short for SN", nil)
	}
	sn := string(decryptedData[8:40])

	// Extract name/SSID
	if len(decryptedData) < 42 {
		return nil, NewDiscoverError("decrypted data too short for name", nil)
	}
	nameLength := int(decryptedData[40])
	if len(decryptedData) < 41+nameLength {
		return nil, NewDiscoverError("decrypted data too short for name", nil)
	}
	name := string(decryptedData[41 : 41+nameLength])

	// Extract device type from name
	var deviceType DeviceType
	parts := strings.Split(name, "_")
	if len(parts) >= 2 {
		typeHex := parts[1]
		if typeVal, err := strconv.ParseInt(typeHex, 16, 32); err == nil {
			deviceType = DeviceType(typeVal)
		}
	}

	// Return device info
	return &DeviceInfo{
		IP:         ip,
		Port:       port,
		DeviceID:   deviceID,
		Name:       name,
		SN:         strings.TrimRight(sn, "\x00"), // Trim null characters
		DeviceType: deviceType,
		Version:    version,
	}, nil
}

// DiscoverSingle discovers a single device by hostname or IP
func DiscoverSingle(ctx context.Context, host string, config *DiscoverConfig) (*Device, error) {
	// Set default values
	if config == nil {
		config = &DiscoverConfig{}
	}

	// Set target to the specific host
	config.Target = host
	config.AutoConnect = false // Don't auto-connect for single discovery

	// Discover devices
	devices, err := Discover(ctx, config)
	if err != nil {
		return nil, err
	}

	// Find the device matching the target host
	for _, device := range devices {
		if device.GetIP() == host {
			return device, nil
		}
	}

	return nil, nil
}

// ConnectDevice connects, authenticates as needed and refreshes a device
func ConnectDevice(device *Device, config *DiscoverConfig) error {
	version := 2
	if v := device.GetVersion(); v != nil {
		version = *v
	}

	// Authenticate V3 devices
	if version == 3 {
		success, err := authenticateDevice(device, config)
		if err != nil {
			return err
		}
		if !success {
			return NewDiscoverError("failed to authenticate V3 device", nil)
		}
	}

	// Attempt to refresh the device state
	if err := device.Refresh(context.Background()); err != nil {
		slog.Warn("Device refresh failed", "error", err)
	}

	return nil
}

// authenticateDevice attempts to authenticate a V3 device
func authenticateDevice(device *Device, config *DiscoverConfig) (bool, error) {
	// First, try to use pre-set token/key if available (skip cloud authentication)
	if device.presetToken != nil && device.presetKey != nil {
		verboseLog("Using pre-set token/key for V3 device authentication (skipping cloud)")
		if err := device.Authenticate(device.presetToken, device.presetKey); err != nil {
			verboseLog("Pre-set token/key authentication failed: %v", err)
			// Fall through to cloud authentication
		} else {
			return true, nil
		}
	}

	// Note: Cloud credentials can be provided explicitly or will use defaults in getCloud()
	// So we don't check for empty Account/Password here

	// Get cloud connection
	cloud, err := getCloud(config)
	if err != nil {
		return false, err
	}

	if cloud == nil {
		return false, NewDiscoverError("cloud connection is nil", nil)
	}

	// Try authenticating with udpids generated from both endians
	for _, endian := range []string{"little", "big"} {
		// Generate udpid
		deviceID := device.GetID()
		var udpid []byte
		if endian == "little" {
			udpid = SecurityUdpid(intToBytesLittleEndian(deviceID, 6))
		} else {
			udpid = SecurityUdpid(intToBytesBigEndian(deviceID, 6))
		}

		udpidHex := hex.EncodeToString(udpid)
		verboseLog("Fetching token and key for udpid '%s' (%s).", udpidHex, endian)

		// Get token and key from cloud
		token, key, err := cloud.GetToken(udpidHex)
		if err != nil {
			slog.Warn("Failed to get token from cloud", "error", err)
			continue
		}

		// Convert hex strings to bytes
		tokenBytes, err := hex.DecodeString(token)
		if err != nil {
			continue
		}
		keyBytes, err := hex.DecodeString(key)
		if err != nil {
			continue
		}

		// Authenticate
		if err := device.Authenticate(tokenBytes, keyBytes); err != nil {
			if _, ok := err.(*AuthenticationError); ok {
				continue
			}
			return false, err
		}

		return true, nil
	}

	return false, nil
}

// getCloud returns a cloud connection
var cloudInstance *NetHomePlusCloud
var cloudMutex sync.Mutex

func getCloud(config *DiscoverConfig) (*NetHomePlusCloud, error) {
	cloudMutex.Lock()
	defer cloudMutex.Unlock()

	// Create cloud connection if nonexistent
	if cloudInstance == nil {
		var account *string
		var password *string

		if config.Account != "" {
			account = &config.Account
		}
		if config.Password != "" {
			password = &config.Password
		}

		cloud, err := NewNetHomePlusCloud(config.Region, account, password, nil)
		if err != nil {
			return nil, NewDiscoverError("failed to create cloud connection", err)
		}

		// Login
		if err := cloud.Login(false); err != nil {
			return nil, NewDiscoverError("failed to login to cloud", err)
		}

		cloudInstance = cloud
	}

	return cloudInstance, nil
}

// intToBytesLittleEndian converts an int to little endian bytes
func intToBytesLittleEndian(n int, size int) []byte {
	b := make([]byte, size)
	for i := 0; i < size; i++ {
		b[i] = byte(n >> (i * 8))
	}
	return b
}

// intToBytesBigEndian converts an int to big endian bytes
func intToBytesBigEndian(n int, size int) []byte {
	b := make([]byte, size)
	for i := 0; i < size; i++ {
		b[size-1-i] = byte(n >> (i * 8))
	}
	return b
}
