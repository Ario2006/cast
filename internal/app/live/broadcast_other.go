//go:build !darwin && !linux

package live

import "net"

func enableBroadcast(*net.UDPConn) error { return nil }
