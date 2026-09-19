// Package discovery provides local UDP broadcast discovery for cast sessions.
package discovery

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	netplatform "github.com/aryankumar/cast/internal/platform/net"
)

const (
	// SharePort is the designated UDP port for file share discovery.
	SharePort = 39421
	// LivePort is the designated UDP port for live session discovery.
	LivePort = 39422
)

var codeAlphabet = []byte("23456789ABCDEFGHJKLMNPQRSTUVWXYZ")

// RandomCode generates an uppercase, human-friendly code avoiding ambiguous runes.
func RandomCode(length int) (string, error) {
	bytes := make([]byte, length)
	for index := range bytes {
		for {
			var randomByte [1]byte
			if _, err := rand.Read(randomByte[:]); err != nil {
				return "", fmt.Errorf("generate code: %w", err)
			}
			limit := 256 - (256 % len(codeAlphabet))
			if int(randomByte[0]) < limit {
				bytes[index] = codeAlphabet[int(randomByte[0])%len(codeAlphabet)]
				break
			}
		}
	}
	return string(bytes), nil
}

// ValidCode checks if a short code has valid length and characters.
func ValidCode(code string) bool {
	if len(code) < 4 || len(code) > 8 {
		return false
	}
	for _, character := range code {
		if !strings.ContainsRune(string(codeAlphabet), character) {
			return false
		}
	}
	return true
}

// RandomToken generates a high-entropy base64 URL-safe token.
func RandomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// LocalIPv4 discovers the primary non-loopback IPv4 address for local network access.
func LocalIPv4() net.IP {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, networkInterface := range interfaces {
			if networkInterface.Flags&net.FlagUp == 0 || networkInterface.Flags&net.FlagLoopback != 0 {
				continue
			}
			addresses, err := networkInterface.Addrs()
			if err != nil {
				continue
			}
			for _, address := range addresses {
				if ip, ok := address.(*net.IPNet); ok && ip.IP.To4() != nil && !ip.IP.IsLoopback() {
					return ip.IP.To4()
				}
			}
		}
	}
	return net.IPv4(127, 0, 0, 1)
}

// Responder serves UDP discovery queries for a specific port.
type Responder struct {
	conn     *net.UDPConn
	stopOnce sync.Once
}

// QueryHandler processes an incoming discovery query and optionally returns a response to send back.
type QueryHandler func(query []byte, sender *net.UDPAddr) ([]byte, bool)

// StartResponder binds to the discovery port and runs the query handler in a background goroutine.
func StartResponder(port int, handler QueryHandler) (*Responder, error) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: port})
	if err != nil {
		return nil, fmt.Errorf("listen for local discovery on port %d: %w", port, err)
	}
	r := &Responder{conn: conn}
	go func() {
		buffer := make([]byte, 2048)
		for {
			count, sender, err := conn.ReadFromUDP(buffer)
			if err != nil {
				return
			}
			if response, shouldRespond := handler(buffer[:count], sender); shouldRespond && len(response) > 0 {
				_, _ = conn.WriteToUDP(response, sender)
			}
		}
	}()
	return r, nil
}

// Close closes the UDP responder socket.
func (r *Responder) Close() error {
	var err error
	r.stopOnce.Do(func() {
		if r.conn != nil {
			err = r.conn.Close()
		}
	})
	return err
}

// ResponseHandler handles an incoming UDP response. Returning true stops discovery waiting.
type ResponseHandler func(data []byte) (done bool)

// BroadcastQuery broadcasts a query to both the loopback and broadcast addresses,
// invoking onResponse until onResponse returns true or the context/deadline expires.
func BroadcastQuery(ctx context.Context, port int, query []byte, onResponse ResponseHandler) error {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return fmt.Errorf("start discovery query: %w", err)
	}
	defer conn.Close()

	if err := netplatform.EnableBroadcast(conn); err != nil {
		return fmt.Errorf("enable discovery broadcast: %w", err)
	}

	for _, target := range []*net.UDPAddr{
		{IP: net.IPv4(127, 0, 0, 1), Port: port},
		{IP: net.IPv4bcast, Port: port},
	} {
		_, _ = conn.WriteToUDP(query, target)
	}

	deadline := time.Now().Add(3 * time.Second)
	if requestedDeadline, ok := ctx.Deadline(); ok && requestedDeadline.Before(deadline) {
		deadline = requestedDeadline
	}
	if err := conn.SetReadDeadline(deadline); err != nil {
		return fmt.Errorf("set discovery deadline: %w", err)
	}

	buffer := make([]byte, 2048)
	for {
		count, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("discovery response timed out")
		}
		if onResponse(buffer[:count]) {
			return nil
		}
	}
}
