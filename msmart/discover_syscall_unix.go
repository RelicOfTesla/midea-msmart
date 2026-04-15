//go:build !windows

package msmart

import (
	"syscall"
)

// setBroadcastOption sets SO_BROADCAST socket option on Unix-like systems
func setBroadcastOption(fd uintptr) {
	syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
}
