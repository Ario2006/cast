//go:build darwin || linux

package live

import (
	"net"
	"syscall"
)

func enableBroadcast(connection *net.UDPConn) error {
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
