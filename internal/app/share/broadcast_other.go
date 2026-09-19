//go:build !darwin && !linux

package share

import "net"

// Most platforms enable broadcast automatically for UDP sockets. Dedicated
// socket configuration can be added behind a platform file when required.
func enableBroadcast(*net.UDPConn) error { return nil }
