//go:build darwin || linux

package net

import (
	"net"
	"syscall"
)

// EnableBroadcast configures a UDP socket to permit broadcasting.
func EnableBroadcast(connection *net.UDPConn) error {
	raw, err := connection.SyscallConn()
	if err != nil {
		return err
	}
	var socketErr error
	if err := raw.Control(func(fileDescriptor uintptr) {
		socketErr = syscall.SetsockoptInt(int(fileDescriptor), syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	}); err != nil {
		return err
	}
	return socketErr
}
