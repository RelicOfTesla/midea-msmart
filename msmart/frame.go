package msmart

import (
	"fmt"
)

// InvalidFrameException represents an error for invalid frames
type InvalidFrameException struct {
	message string
}

// Error implements the error interface
func (e *InvalidFrameException) Error() string {
	return e.message
}

// NewInvalidFrameException creates a new InvalidFrameException
func NewInvalidFrameException(message string) *InvalidFrameException {
	return &InvalidFrameException{message: message}
}

// Frame represents a communication frame
type Frame struct {
	headerLength    int
	deviceType      DeviceType
	frameType       FrameType
	protocolVersion int
}

// NewFrame creates a new Frame instance
func NewFrame(deviceType DeviceType, frameType FrameType) *Frame {
	return &Frame{
		headerLength:    10,
		deviceType:      deviceType,
		frameType:       frameType,
		protocolVersion: 0,
	}
}

// ToBytes converts data to frame bytes
// This is the Go equivalent of the Python tobytes method
func (f *Frame) ToBytes(data []byte) []byte {
	// Build frame header
	header := make([]byte, f.headerLength)

	// Start byte
	header[0] = 0xAA

	// Length of header and data
	header[1] = byte(len(data) + f.headerLength)

	// Device/appliance type
	header[2] = byte(f.deviceType)

	// Device protocol version
	header[8] = byte(f.protocolVersion)

	// Frame type
	header[9] = byte(f.frameType)

	// Build frame from header and data
	frame := append(header, data...)

	// Calculate total frame checksum
	checksum := Checksum(frame[1:])
	frame = append(frame, checksum)

	return frame
}

// Checksum calculates the frame checksum
// This is the Go equivalent of the Python checksum classmethod
func Checksum(frame []byte) byte {
	// Calculate sum of all bytes in frame
	var sum int
	for _, b := range frame {
		sum += int(b)
	}
	
	// Return checksum: (~sum + 1) & 0xFF
	return byte((^sum + 1) & 0xFF)
}

// Validate validates a frame
// This is the Go equivalent of the Python validate classmethod
func Validate(frame []byte, expectedDeviceType DeviceType) error {
	// Ensure length is sane
	if len(frame) < 10 { // _HEADER_LENGTH
		return NewInvalidFrameException(fmt.Sprintf("Frame is too short: %x", frame))
	}

	// Validate frame checksum
	checksum := Checksum(frame[1 : len(frame)-1])
	if checksum != frame[len(frame)-1] {
		return NewInvalidFrameException(
			fmt.Sprintf("Frame '%x' failed checksum. Received: 0x%X, Expected: 0x%X.",
				frame, frame[len(frame)-1], checksum))
	}

	// Check device type matches
	deviceType := frame[2]
	if DeviceType(deviceType) != expectedDeviceType {
		return NewInvalidFrameException(
			fmt.Sprintf("Received device type 0x%X does not match expected device type 0x%X.",
				deviceType, expectedDeviceType))
	}

	return nil
}
