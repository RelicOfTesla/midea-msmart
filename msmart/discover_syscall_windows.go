//go:build windows

package msmart

import (
	"syscall"
)

// setBroadcastOption sets SO_BROADCAST socket option on Windows
func setBroadcastOption(fd uintptr) {
	syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
}
