//go:build !windows

package msmart

import (
	"fmt"
	"syscall"
)

// setBroadcastOption sets SO_BROADCAST socket option on Unix-like systems
// Returns an error if the operation fails
func setBroadcastOption(fd uintptr) error {
	err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	if err != nil {
		return fmt.Errorf("failed to set SO_BROADCAST: %w", err)
	}
	return nil
}
