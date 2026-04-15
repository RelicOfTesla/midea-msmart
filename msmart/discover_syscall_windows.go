//go:build windows

package msmart

import (
	"fmt"
	"syscall"
)

// setBroadcastOption sets SO_BROADCAST socket option on Windows
// Returns an error if the operation fails
func setBroadcastOption(fd uintptr) error {
	err := syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	if err != nil {
		return fmt.Errorf("failed to set SO_BROADCAST: %w", err)
	}
	return nil
}

// setReuseAddrOption sets SO_REUSEADDR socket option on Windows
// This is often needed on Windows for broadcast sockets
func setReuseAddrOption(fd uintptr) error {
	err := syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		return fmt.Errorf("failed to set SO_REUSEADDR: %w", err)
	}
	return nil
}
