//go:build !darwin && !linux

package net

import "net"

// EnableBroadcast is a no-op fallback on platforms where UDP sockets permit broadcast by default.
func EnableBroadcast(*net.UDPConn) error { return nil }
