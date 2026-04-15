// Package msmart provides local network control of Midea AC devices.
package msmart

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// Token represents authentication token
type Token []byte

// Key represents encryption key
type Key []byte

var logger = log.Default()

// ProtocolError represents a protocol error
type ProtocolError struct {
	Message string
	Cause   error
}

func (e *ProtocolError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("protocol error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("protocol error: %s", e.Message)
}

func (e *ProtocolError) Unwrap() error {
	return e.Cause
}

// AuthenticationError represents an authentication error
type AuthenticationError struct {
	Message string
	Cause   error
}

func (e *AuthenticationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("authentication error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("authentication error: %s", e.Message)
}

func (e *AuthenticationError) Unwrap() error {
	return e.Cause
}

// Protocol defines the interface for LAN protocol
//
// V2 Packet Overview
//
// Header: 40 bytes
//
//	2 byte start of packet: 0x5A5A
//	2 byte message type: 0x0111
//	2 byte packet length
//	2 byte magic bytes: Usually 0x2000, special responses differ
//	4 byte message ID
//	8 byte timestamp
//	8 byte device ID
//	12 byte ???
//
// Payload: N bytes
//
//	N byte data payload contains encrypted frame
//
// Sign: 16 bytes
//
//	16 byte MD5 of packet + fixed key
type Protocol interface {
	Peer() string
	Alive() bool
	Disconnect() error
	Write(data []byte) error
	Read(timeout time.Duration) ([]byte, error)
	Flush()
	ConnectionMade(conn net.Conn)
	DataReceived(data []byte)
	ConnectionLost(err error)
}

// _LanProtocol represents Midea LAN protocol
type _LanProtocol struct {
	transport net.Conn
	peer      string
	queue     chan interface{} // Can contain []byte or error
	mu        sync.Mutex
}

// NewLanProtocol creates a new LanProtocol instance
func NewLanProtocol() *_LanProtocol {
	return &_LanProtocol{
		queue: make(chan interface{}, 100),
	}
}

// Peer returns the peer address string
func (p *_LanProtocol) Peer() string {
	return p.peer
}

// Alive checks if the connection is alive
func (p *_LanProtocol) Alive() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.transport == nil {
		return false
	}

	// Try to check if connection is still valid
	// In Go, we can't directly check if connection is closing,
	// so we just check if it's nil
	return true
}

func (p *_LanProtocol) formatSocketName(addr net.Addr) string {
	return addr.String()
}

// ConnectionMade handles connection events
func (p *_LanProtocol) ConnectionMade(conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.transport = conn
	p.peer = p.formatSocketName(conn.RemoteAddr())
	logger.Printf("Connected to %s.", p.peer)
}

// DataReceived handles data received events
func (p *_LanProtocol) DataReceived(data []byte) {
	logger.Printf("Received data from %s: %x", p.peer, data)
	p.queue <- data
}

// ConnectionLost logs connection lost
func (p *_LanProtocol) ConnectionLost(err error) {
	if err != nil {
		logger.Printf("Connection to %s lost. Error: %v", p.peer, err)
		p.queue <- err
	}
}

// Disconnect disconnects from the peer
func (p *_LanProtocol) Disconnect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.transport == nil {
		return errors.New("transport is nil")
	}

	logger.Printf("Disconnecting from %s.", p.peer)
	return p.transport.Close()
}

// Write sends data to the peer
func (p *_LanProtocol) Write(data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.transport == nil {
		return errors.New("transport is nil")
	}

	// Check if connection is still valid without calling Alive() to avoid deadlock
	// since we already hold the mutex

	logger.Printf("Sending data to %s: %x", p.peer, data)
	_, err := p.transport.Write(data)
	return err
}

// WriteWithType sends a packet of the specified type to the peer (for V3 protocol)
func (p *_LanProtocol) WriteWithType(data []byte, packetType PacketType) error {
	return p.Write(data)
}

// ReadQueue reads data from the receive queue
func (p *_LanProtocol) ReadQueue(timeout time.Duration) ([]byte, error) {
	var timer <-chan time.Time
	if timeout > 0 {
		timer = time.After(timeout)
	} else {
		// For timeout=0, use a very short timeout
		timer = time.After(10 * time.Millisecond)
	}

	select {
	case item := <-p.queue:
		switch v := item.(type) {
		case []byte:
			return v, nil
		case error:
			return nil, &ProtocolError{Message: "", Cause: v}
		default:
			return nil, &ProtocolError{Message: "unknown item type in queue"}
		}
	case <-timer:
		return nil, &ProtocolError{Message: "timeout waiting for data"}
	}
}

// Read asynchronously reads data from the peer via the queue
func (p *_LanProtocol) Read(timeout time.Duration) ([]byte, error) {
	return p.ReadQueue(timeout)
}

// Flush flushes all data from the receive queue
func (p *_LanProtocol) Flush() {
	for {
		select {
		case <-p.queue:
			// discard
		default:
			return
		}
	}
}

// PacketType represents V3 packet types
type PacketType int

const (
	PacketTypeHandshakeRequest  PacketType = 0x0
	PacketTypeHandshakeResponse PacketType = 0x1
	PacketTypeEncryptedResponse PacketType = 0x3
	PacketTypeEncryptedRequest  PacketType = 0x6
	PacketTypeError             PacketType = 0xF
)

// AuthenticationExpiration is the authentication expiration duration
const AuthenticationExpiration = 12 * time.Hour

// _LanProtocolV3 represents Midea LAN protocol V3
//
// V3 Packet Overview
//
// Header: 6 bytes
//
//	2 byte start of packet: 0x8370
//	2 byte size of data payload, padding and sign
//	1 special byte: 0x20
//	1 padding and type byte: pad << 4 | type
//
// Payload: N + 2 bytes
//
//	2 byte request ID/count
//	N byte data payload
//
// Sign: 32 bytes
//
//	32 byte SHA256 of header + unencrypted payload
//
// Notes
//
//	- For padding purposes the 2 byte request ID is included in size,
//	  but not in the size field
//	- When used for device command/response, the payload contains a V2 packet
type _LanProtocolV3 struct {
	*_LanProtocol
	packetID           uint16
	buffer             []byte
	localKey           []byte
	localKeyExpiration time.Time
}

// Ensure _LanProtocolV3 implements Protocol interface
var _ Protocol = (*_LanProtocolV3)(nil)

// NewLanProtocolV3 creates a new LanProtocolV3 instance
func NewLanProtocolV3() *_LanProtocolV3 {
	return &_LanProtocolV3{
		_LanProtocol: NewLanProtocol(),
		buffer:       make([]byte, 0),
	}
}

// Authenticated checks if the protocol is authenticated
func (p *_LanProtocolV3) Authenticated() bool {
	if p.localKey == nil || p.localKeyExpiration.IsZero() {
		return false
	}

	if time.Now().UTC().After(p.localKeyExpiration) {
		logger.Printf("Authentication with %s has expired.", p.Peer())
		return false
	}

	return true
}

// DataReceived handles data received events
func (p *_LanProtocolV3) DataReceived(data []byte) {
	logger.Printf("Received data from %s: %x", p.Peer(), data)

	// Add incoming data to buffer
	p.buffer = append(p.buffer, data...)

	// Process buffer until empty
	for len(p.buffer) > 0 {
		// Find start of packet
		start := bytes.Index(p.buffer, []byte{0x83, 0x70})
		if start == -1 {
			logger.Printf("Peer %s: No start of packet found. Buffer: %x", p.Peer(), p.buffer)
			return
		}

		// Trim any leading data
		if start != 0 {
			logger.Printf("Peer %s: Ignoring data before packet: %x", p.Peer(), p.buffer[:start])
			p.buffer = p.buffer[start:]
		}

		// Check if the header has been received
		if len(p.buffer) < 6 {
			logger.Printf("Peer %s: Buffer too short. Buffer: %x", p.Peer(), p.buffer)
			return
		}

		// 6 byte header + 2 packet id + padded encrypted payload
		totalSize := int(binary.BigEndian.Uint16(p.buffer[2:4])) + 8

		// Ensure entire packet is received
		if len(p.buffer) < totalSize {
			logger.Printf("Peer %s: Partial packet received. Buffer: %x", p.Peer(), p.buffer)
			return
		}

		// Extract the packet from the buffer
		packet := make([]byte, totalSize)
		copy(packet, p.buffer[:totalSize])
		p.buffer = p.buffer[totalSize:]

		// Queue the received packet
		p.queue <- packet
	}
}

// decodeEncryptedResponse decodes an encrypted response packet
func (p *_LanProtocolV3) decodeEncryptedResponse(packet []byte) ([]byte, error) {
	// We should always have a key by the time we're received data
	if p.localKey == nil {
		return nil, &ProtocolError{Message: "local key is nil"}
	}

	// Extract header, encrypted payload and hash
	header := packet[:6]
	payload := packet[6 : len(packet)-32]
	rxHash := packet[len(packet)-32:]

	// Decrypt payload
	decryptedPayload, err := SecurityDecryptAESCBC(p.localKey, payload)
	if err != nil {
		return nil, err
	}

	// Verify hash
	hash := sha256.Sum256(append(header, decryptedPayload...))
	if !bytes.Equal(hash[:], rxHash) {
		return nil, &ProtocolError{Message: "Calculated and received SHA256 digest do not match."}
	}

	// Decrypted payload consists of 2 byte packet ID + actual payload + padding
	// Get pad count from header
	pad := header[5] >> 4

	// Get the frame from payload (skip 2 byte packet ID, remove padding)
	if len(decryptedPayload) < 2+int(pad) {
		return nil, &ProtocolError{Message: "decrypted payload too short"}
	}

	return decryptedPayload[2 : len(decryptedPayload)-int(pad)], nil
}

// decodeHandshakeResponse decodes a handshake response packet
func (p *_LanProtocolV3) decodeHandshakeResponse(packet []byte) ([]byte, error) {
	// Get payload from packet
	// Return remaining raw payload (skip 2 byte packet ID)
	if len(packet) < 8 {
		return nil, &ProtocolError{Message: "packet too short for handshake response"}
	}
	return packet[8:], nil
}

// processPacket processes a received packet based on its type
func (p *_LanProtocolV3) processPacket(packet []byte) ([]byte, error) {
	if len(packet) < 6 {
		return nil, &ProtocolError{Message: fmt.Sprintf("packet too short: %x", packet)}
	}

	if !bytes.Equal(packet[:2], []byte{0x83, 0x70}) {
		return nil, &ProtocolError{Message: fmt.Sprintf("Invalid start of packet: %x", packet[:2])}
	}

	if packet[4] != 0x20 {
		return nil, &ProtocolError{Message: fmt.Sprintf("Invalid magic byte: 0x%X", packet[4])}
	}

	// Handle packet based on type
	packetType := PacketType(packet[5] & 0xF)
	switch packetType {
	case PacketTypeEncryptedResponse:
		return p.decodeEncryptedResponse(packet)
	case PacketTypeHandshakeResponse:
		return p.decodeHandshakeResponse(packet)
	case PacketTypeError:
		return nil, &ProtocolError{Message: "Error packet received."}
	default:
		return nil, &ProtocolError{Message: fmt.Sprintf("Unexpected type: %d", packetType)}
	}
}

// Read asynchronously reads data from the peer via the queue
func (p *_LanProtocolV3) Read(timeout time.Duration) ([]byte, error) {
	// Fetch a packet from the queue
	packet, err := p.ReadQueue(timeout)
	if err != nil {
		return nil, err
	}

	return p.processPacket(packet)
}

// BuildHeader builds the packet header
func (p *_LanProtocolV3) BuildHeader(length int, extra byte) []byte {
	// Build header
	header := []byte{0x83, 0x70}
	header = binary.BigEndian.AppendUint16(header, uint16(length))
	header = append(header, 0x20)
	header = append(header, extra)

	return header
}

// encodeEncryptedRequest encodes an encrypted request packet
func (p *_LanProtocolV3) encodeEncryptedRequest(packetID uint16, data []byte) ([]byte, error) {
	// Raise an error if attempting to send an encrypted request without authenticating
	if p.localKey == nil {
		return nil, &ProtocolError{Message: "Protocol has not been authenticated."}
	}

	// Compute required padding for 16 byte alignment
	// Include 2 bytes for packet ID in total length
	remainder := (len(data) + 2) % 16
	pad := 0
	if remainder != 0 {
		pad = 16 - remainder
	}

	// Compute total length of payload, pad and hash
	length := len(data) + pad + 32

	// Build header
	header := p.BuildHeader(length, byte(pad<<4)|byte(PacketTypeEncryptedRequest))

	// Build payload to encrypt
	pidBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(pidBytes, packetID)
	payload := append(pidBytes, data...)

	// Add random padding
	randomPad := make([]byte, pad)
	if pad > 0 {
		rand.Read(randomPad)
		payload = append(payload, randomPad...)
	}

	// Encrypt payload
	encryptedPayload, err := SecurityEncryptAESCBC(p.localKey, payload)
	if err != nil {
		return nil, err
	}

	// Calculate hash
	hash := sha256.Sum256(append(header, payload...))

	return append(append(header, encryptedPayload...), hash[:]...), nil
}

// encodeHandshakeRequest encodes a handshake request packet
func (p *_LanProtocolV3) encodeHandshakeRequest(packetID uint16, data []byte) []byte {
	// Build header
	header := p.BuildHeader(len(data), byte(PacketTypeHandshakeRequest))

	// Build payload to encrypt
	pidBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(pidBytes, packetID)
	payload := append(pidBytes, data...)

	return append(header, payload...)
}

// Write sends data to the peer (uses EncryptedRequest by default for V3)
func (p *_LanProtocolV3) Write(data []byte) error {
	return p.WriteWithType(data, PacketTypeEncryptedRequest)
}

// WriteWithType sends a packet of the specified type to the peer
func (p *_LanProtocolV3) WriteWithType(data []byte, packetType PacketType) error {
	var packet []byte
	var err error

	// Encode the data according to the supplied type
	switch packetType {
	case PacketTypeEncryptedRequest:
		packet, err = p.encodeEncryptedRequest(p.packetID, data)
		if err != nil {
			return err
		}
	case PacketTypeHandshakeRequest:
		packet = p.encodeHandshakeRequest(p.packetID, data)
	default:
		return fmt.Errorf("unknown type: %d", packetType)
	}

	// Write to the peer
	if err := p._LanProtocol.Write(packet); err != nil {
		return err
	}

	// Increment packet ID and handle rollover
	p.packetID++
	p.packetID &= 0xFFF // Mask to 12 bits

	return nil
}

// GetLocalKey generates the local key from the cloud key and response data
func (p *_LanProtocolV3) GetLocalKey(key []byte, data []byte) ([]byte, error) {
	if len(data) != 64 {
		return nil, &AuthenticationError{Message: "Invalid data length for key handshake."}
	}

	// Extract payload and hash
	payload := data[:32]
	rxHash := data[32:]

	// Decrypt the payload with the provided key
	decryptedPayload, err := SecurityDecryptAESCBC(key, payload)
	if err != nil {
		return nil, err
	}

	// Verify hash
	hash := sha256.Sum256(decryptedPayload)
	if !bytes.Equal(hash[:], rxHash) {
		return nil, &AuthenticationError{Message: "Calculated and received SHA256 digest do not match."}
	}

	// Construct the local key using XOR
	// Python's strxor requires equal-length inputs, so we must match that behavior
	if len(key) != len(decryptedPayload) {
		return nil, &AuthenticationError{Message: "Key length must match decrypted payload length."}
	}

	localKey := make([]byte, len(decryptedPayload))
	for i := range decryptedPayload {
		localKey[i] = decryptedPayload[i] ^ key[i]
	}

	return localKey, nil
}

// Authenticate authenticates the connection with token and key
func (p *_LanProtocolV3) Authenticate(token []byte, key []byte) error {
	// Raise an exception if trying to auth without any token or key
	if len(token) == 0 || len(key) == 0 {
		return &AuthenticationError{Message: "Token and key must be supplied."}
	}

	// Flush any existing data from the queue
	p.Flush()

	// Send handshake request
	if err := p.WriteWithType(token, PacketTypeHandshakeRequest); err != nil {
		return err
	}

	// Read response
	response, err := p.Read(2 * time.Second)
	if err != nil {
		// Promote any protocol error to auth error
		if pe, ok := err.(*ProtocolError); ok {
			return &AuthenticationError{Message: "", Cause: pe}
		}
		return &AuthenticationError{Message: "", Cause: err}
	}

	// Generate local key from cloud key
	p.localKey, err = p.GetLocalKey(key, response)
	if err != nil {
		return err
	}

	// Set expiration time
	p.localKeyExpiration = time.Now().UTC().Add(AuthenticationExpiration)

	logger.Printf("Authentication with %s successful. Expiration: %s, Local key: %x",
		p.Peer(), p.localKeyExpiration.Format(time.RFC3339), p.localKey)

	return nil
}

// LAN represents a LAN connection to a Midea device
type LAN struct {
	ip                    string
	port                  int
	deviceID              int64
	token                 []byte
	key                   []byte
	protocolVersion       int
	protocol              Protocol
	protocolV3            *_LanProtocolV3
	connectionExpiration  time.Time
	maxConnectionLifetime *time.Duration
	mu                    sync.Mutex
	conn                  net.Conn // Store connection for deadline management
	readTimeout           time.Duration
}

// Retries is the default number of retries
const Retries = 3

// NewLAN creates a new LAN instance
func NewLAN(ip string, port int, deviceID int64) *LAN {
	return &LAN{
		ip:              ip,
		port:            port,
		deviceID:        deviceID,
		protocolVersion: 2,
		readTimeout:     5 * time.Second, // Default read timeout
	}
}

// Token returns the current token
func (l *LAN) Token() []byte {
	return l.token
}

// Key returns the current key
func (l *LAN) Key() []byte {
	return l.key
}

// MaxConnectionLifetime returns the maximum connection lifetime in seconds
func (l *LAN) MaxConnectionLifetime() *int {
	if l.maxConnectionLifetime == nil {
		return nil
	}
	seconds := int(l.maxConnectionLifetime.Seconds())
	return &seconds
}

// SetMaxConnectionLifetime sets the maximum connection lifetime in seconds
func (l *LAN) SetMaxConnectionLifetime(seconds *int) {
	if seconds == nil {
		l.maxConnectionLifetime = nil
	} else {
		d := time.Duration(*seconds) * time.Second
		l.maxConnectionLifetime = &d
	}
}

// Alive checks if the connection is alive
func (l *LAN) alive() bool {
	// Check if protocol exists, and if it's alive
	if l.protocol == nil || !l.protocol.Alive() {
		return false
	}

	// Use connection expiration if set
	if !l.connectionExpiration.IsZero() && time.Now().UTC().After(l.connectionExpiration) {
		logger.Printf("Connection to %s has expired.", l.protocol.Peer())
		return false
	}

	return true
}

// Connect establishes a connection to the device
// Note: This function does NOT acquire the mutex lock. Callers must hold l.mu.Lock().
func (l *LAN) connect() error {
	logger.Printf("Creating new connection to %s:%d.", l.ip, l.port)

	// Connect with timeout
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", l.ip, l.port), 5*time.Second)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return &ProtocolError{Message: "Connect timeout.", Cause: err}
		}
		return &ProtocolError{Message: "Connect failed.", Cause: err}
	}

	// Store connection for deadline management
	l.conn = conn

	if l.protocolVersion == 3 {
		l.protocolV3 = NewLanProtocolV3()
		l.protocol = l.protocolV3
	} else {
		l.protocol = NewLanProtocol()
	}

	l.protocol.ConnectionMade(conn)

	if l.maxConnectionLifetime != nil {
		l.connectionExpiration = time.Now().UTC().Add(*l.maxConnectionLifetime)
	}

	// Start a goroutine to read data with deadline management
	// Store local reference to protocol to avoid nil pointer issues if disconnect is called
	localProtocol := l.protocol
	go func() {
		buf := make([]byte, 4096)
		for {
			// Set a read deadline to allow periodic checks
			// The deadline is set to a reasonable value to avoid blocking forever
			if l.readTimeout > 0 {
				conn.SetReadDeadline(time.Now().Add(l.readTimeout + 1*time.Second))
			}

			n, err := conn.Read(buf)
			if err != nil {
				// Check if it's a deadline error (timeout)
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// This is expected - continue waiting for data
					continue
				}
				if err != io.EOF {
					localProtocol.ConnectionLost(err)
				}
				return
			}
			data := make([]byte, n)
			copy(data, buf[:n])
			localProtocol.DataReceived(data)
		}
	}()

	return nil
}

// Disconnect closes the connection
func (l *LAN) disconnect() {
	if l.protocol != nil {
		l.protocol.Disconnect()
		l.protocol = nil
		l.protocolV3 = nil
	}
	l.conn = nil
}

// Authenticate authenticates against a V3 device
func (l *LAN) Authenticate(token Token, key Key, retries int) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Use existing token and key if none provided
	if token == nil || key == nil {
		token = l.token
		key = l.key
	} else {
		// Ensure token and key are in byte form
		// (In Go, we assume they're already bytes, unlike Python where they could be hex strings)
	}

	// Create a connection if not alive or protocol isn't V3
	if !l.alive() || l.protocolV3 == nil {
		l.disconnect()
		l.protocolVersion = 3
		if err := l.connect(); err != nil {
			return err
		}
	}

	// A V3 protocol should exist at this point
	if l.protocolV3 == nil {
		return &ProtocolError{Message: "protocol V3 is nil"}
	}

	logger.Printf("Authenticating with %s.", l.protocol.Peer())

	// Attempt to authenticate
	for retries > 0 {
		err := l.protocolV3.Authenticate(token, key)
		if err == nil {
			break
		}

		var authErr *AuthenticationError
		if errors.As(err, &authErr) {
			l.disconnect()
			return err
		}

		// Handle timeout
		if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
			if retries > 1 {
				logger.Printf("Authentication timeout. Resending to %s.", l.protocol.Peer())
				retries--
			} else {
				l.disconnect()
				return &ProtocolError{Message: "No response from host.", Cause: err}
			}
		} else {
			l.disconnect()
			return err
		}
	}

	// Protocol should be authenticated by now
	if !l.protocolV3.Authenticated() {
		return &AuthenticationError{Message: "protocol not authenticated"}
	}

	// Update stored token and key if successful
	l.token = token
	l.key = key

	// Sleep briefly before requesting more data
	time.Sleep(1 * time.Second)

	return nil
}

// Read reads and decodes a frame from the protocol
func (l *LAN) read(timeout time.Duration) ([]byte, error) {
	// A protocol should exist at this point
	if l.protocol == nil {
		return nil, &ProtocolError{Message: "protocol is nil"}
	}

	// Await a response
	packet, err := l.protocol.Read(timeout)
	if err != nil {
		return nil, err
	}
	logger.Printf("Received packet from %s: %x", l.protocol.Peer(), packet)

	// Decode packet to frame
	response, err := PacketDecode(packet)
	if err != nil {
		return nil, err
	}
	logger.Printf("Received response from %s: %x", l.protocol.Peer(), response)

	return response, nil
}

// ReadAvailable reads responses from the queue without blocking
func (l *LAN) readAvailable() ([][]byte, error) {
	var responses [][]byte
	for {
		resp, err := l.read(0)
		if err != nil {
			break
		}
		responses = append(responses, resp)
	}
	return responses, nil
}

// Send sends data via the LAN protocol
func (l *LAN) Send(data []byte, retries int) ([][]byte, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Connect if protocol doesn't exist or is dead
	if !l.alive() {
		l.disconnect()
		if err := l.connect(); err != nil {
			return nil, err
		}
	}

	// A protocol should exist at this point
	if l.protocol == nil {
		return nil, &ProtocolError{Message: "protocol is nil"}
	}

	// Authenticate as needed
	if l.protocolV3 != nil && !l.protocolV3.Authenticated() {
		// Unlock during authentication to allow other operations
		l.mu.Unlock()
		err := l.Authenticate(nil, nil, Retries)
		l.mu.Lock()

		if err != nil {
			return nil, err
		}

		// Protocol should be authenticated now
		if !l.protocolV3.Authenticated() {
			return nil, &AuthenticationError{Message: "protocol not authenticated after authentication"}
		}
	}

	// Encode frame to packet
	packet, err := PacketEncode(l.deviceID, data)
	if err != nil {
		return nil, err
	}

	var responses [][]byte

	// Read any responses that may have been received sporadically
	available, _ := l.readAvailable()
	responses = append(responses, available...)

	// Send the request and wait for a response
	for retries > 0 {
		// Send the request
		logger.Printf("Sending packet to %s: %x", l.protocol.Peer(), packet)

		if l.protocolV3 != nil {
			err = l.protocolV3.WriteWithType(packet, PacketTypeEncryptedRequest)
		} else {
			err = l.protocol.Write(packet)
		}
		if err != nil {
			return nil, err
		}

		// Await a response with timeout
		// Use the configured read timeout
		readTimeout := l.readTimeout
		if readTimeout == 0 {
			readTimeout = 5 * time.Second
		}
		resp, err := l.read(readTimeout)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				if retries > 1 {
					logger.Printf("Read timeout. Resending to %s.", l.protocol.Peer())
					retries--
					continue
				} else {
					l.disconnect()
					return nil, &ProtocolError{Message: "No response from host.", Cause: err}
				}
			}

			// Disconnect on protocol errors and reraise
			logger.Printf("Send failed to %s. Error: %v", l.protocol.Peer(), err)
			l.disconnect()
			return nil, err
		}

		responses = append(responses, resp)
		break
	}

	// Read any additional responses without blocking
	available, _ = l.readAvailable()
	responses = append(responses, available...)

	return responses, nil
}

// Security provides encryption and signing utilities
type Security struct{}

// Security constants
var (
	SecuritySignKey = []byte("xhdiwjnchekd4d512chdjx5d8e4c394D2D7S")
	SecurityEncKey  = md5.Sum(SecuritySignKey)
)

// SecurityDecryptAESCBC decrypts data using AES-CBC
func SecurityDecryptAESCBC(key []byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(data)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("data is not a multiple of block size")
	}

	iv := make([]byte, aes.BlockSize) // Zero IV
	mode := cipher.NewCBCDecrypter(block, iv)

	decrypted := make([]byte, len(data))
	mode.CryptBlocks(decrypted, data)

	return decrypted, nil
}

// SecurityEncryptAESCBC encrypts data using AES-CBC
// Note: Data must already be padded to block size by the caller
func SecurityEncryptAESCBC(key []byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(data)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("data is not a multiple of block size")
	}

	iv := make([]byte, aes.BlockSize) // Zero IV
	mode := cipher.NewCBCEncrypter(block, iv)

	encrypted := make([]byte, len(data))
	mode.CryptBlocks(encrypted, data)

	return encrypted, nil
}

// SecurityDecryptAES decrypts data using AES-ECB
func SecurityDecryptAES(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(SecurityEncKey[:])
	if err != nil {
		return nil, err
	}

	if len(data)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("data is not a multiple of block size")
	}

	decrypted := make([]byte, len(data))
	for i := 0; i < len(data); i += aes.BlockSize {
		block.Decrypt(decrypted[i:i+aes.BlockSize], data[i:i+aes.BlockSize])
	}

	// Remove padding
	return PKCS7Unpad(decrypted)
}

// SecurityEncryptAES encrypts data using AES-ECB
func SecurityEncryptAES(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(SecurityEncKey[:])
	if err != nil {
		return nil, err
	}

	// Pad data to block size
	padded := PKCS7Pad(data, aes.BlockSize)

	encrypted := make([]byte, len(padded))
	for i := 0; i < len(padded); i += aes.BlockSize {
		block.Encrypt(encrypted[i:i+aes.BlockSize], padded[i:i+aes.BlockSize])
	}

	return encrypted, nil
}

// SecuritySign signs data with MD5
func SecuritySign(data []byte) []byte {
	// Create a new slice to avoid modifying the original data
	// when append is used on a slice with extra capacity
	combined := make([]byte, 0, len(data)+len(SecuritySignKey))
	combined = append(combined, data...)
	combined = append(combined, SecuritySignKey...)
	hash := md5.Sum(combined)
	return hash[:]
}

// SecurityUdpid generates UDP ID from device ID
func SecurityUdpid(deviceID []byte) []byte {
	hash := sha256.Sum256(deviceID)
	result := make([]byte, 16)
	for i := 0; i < 16; i++ {
		result[i] = hash[i] ^ hash[i+16]
	}
	return result
}

// PKCS7Pad pads data using PKCS7
func PKCS7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// PKCS7Unpad removes PKCS7 padding
func PKCS7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	padding := int(data[len(data)-1])
	if padding > len(data) || padding > aes.BlockSize {
		return nil, fmt.Errorf("invalid padding")
	}

	// Verify padding
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding")
		}
	}

	return data[:len(data)-padding], nil
}

// _Packet handles encoding/decoding command frames to packets
type _Packet struct{}

// PacketEncode encodes a command frame to a LAN packet
func PacketEncode(deviceID int64, command []byte) ([]byte, error) {
	// Encrypt command
	encryptedPayload, err := SecurityEncryptAES(command)
	if err != nil {
		return nil, err
	}

	// Compute total length
	length := 40 + len(encryptedPayload) + 16

	header := []byte{0x5A, 0x5A} // Start of packet
	header = append(header, 0x01, 0x11) // Message type
	header = binary.LittleEndian.AppendUint16(header, uint16(length)) // Packet size
	header = append(header, 0x20, 0x00) // Magic bytes
	header = append(header, 0, 0, 0, 0) // Message ID
	header = append(header, packetTimestamp()...) // Timestamp

	// Device ID (8 bytes, little endian)
	deviceIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(deviceIDBytes, uint64(deviceID))
	header = append(header, deviceIDBytes...)

	header = append(header, make([]byte, 12)...) // ???

	packet := append(header, encryptedPayload...)

	// Append hash
	sign := SecuritySign(packet)
	return append(packet, sign...), nil
}

// PacketDecode decodes a LAN packet to a command frame
func PacketDecode(data []byte) ([]byte, error) {
	if len(data) < 6 {
		return nil, &ProtocolError{Message: fmt.Sprintf("Packet is too short: %x", data)}
	}

	if !bytes.Equal(data[:2], []byte{0x5a, 0x5a}) {
		// TODO old code handled raw frames? e.g start = 0xAA
		return nil, &ProtocolError{Message: fmt.Sprintf("Unsupported packet: %x", data)}
	}

	length := int(binary.LittleEndian.Uint16(data[4:6]))

	if len(data) < length {
		return nil, &ProtocolError{Message: fmt.Sprintf("Packet is truncated. Expected %d bytes, only have %d bytes: %x", length, len(data), data)}
	}

	packet := data[:length]
	encryptedFrame := packet[40 : len(packet)-16]
	rxHash := packet[len(packet)-16:]

	// Check that received hash matches
	sign := SecuritySign(packet[:len(packet)-16])
	if !bytes.Equal(sign, rxHash) {
		return nil, &ProtocolError{Message: "Calculated and received MD5 digest do not match."}
	}

	// Decrypt frame
	return SecurityDecryptAES(encryptedFrame)
}

// packetTimestamp generates a timestamp for the packet
func packetTimestamp() []byte {
	now := time.Now().UTC()

	// Each byte is a 2 digit component of the timestamp
	// YYYYMMDDHHMMSSmm
	microsecond := now.Nanosecond() / 10000

	return []byte{
		byte(microsecond),
		byte(now.Second()),
		byte(now.Minute()),
		byte(now.Hour()),
		byte(now.Day()),
		byte(now.Month()),
		byte(now.Year() % 100),
		byte(now.Year() / 100),
	}
}

// ============================================================================
// Compatibility Layer - LANClient for backward compatibility with device.go
// ============================================================================

// LANClient is a compatibility wrapper around LAN for backward compatibility
// This matches the interface expected by device.go
type LANClient struct {
	lan *LAN
}

// NewLANClient creates a new LANClient (for backward compatibility)
func NewLANClient(ip string, port int, deviceID uint64) *LANClient {
	return &LANClient{
		lan: NewLAN(ip, port, int64(deviceID)),
	}
}

// Connect establishes a connection to the device
func (c *LANClient) Connect(ctx context.Context) error {
	return c.lan.connect()
}

// Close closes the connection
func (c *LANClient) Close() error {
	c.lan.disconnect()
	return nil
}

// Authenticate authenticates with a V3 device
func (c *LANClient) Authenticate(token, key []byte) error {
	return c.lan.Authenticate(token, key, Retries)
}

// Send sends data to the device
func (c *LANClient) Send(ctx context.Context, data []byte) ([][]byte, error) {
	return c.lan.Send(data, Retries)
}

// Token returns the current token
func (c *LANClient) Token() []byte {
	return c.lan.Token()
}

// Key returns the current key
func (c *LANClient) Key() []byte {
	return c.lan.Key()
}

// SetMaxConnectionLifetime sets the maximum connection lifetime
func (c *LANClient) SetMaxConnectionLifetime(seconds int) {
	c.lan.SetMaxConnectionLifetime(&seconds)
}

// SetTimeout sets the connection timeout (for compatibility)
func (c *LANClient) SetTimeout(timeout time.Duration) {
	// Store timeout for future use
	// The LAN type uses a default timeout of 5 seconds
}
