// Package session defines the canonical domain model for sharing and live sessions.
package session

import (
	"time"
)

// Type distinguishes between different session modes.
type Type string

const (
	TypeShare         Type = "share"
	TypeLiveDirectory Type = "live_directory"
	TypeLiveProxy     Type = "live_proxy"
)

// Permission represents an authorization grant for a session.
type Permission string

const (
	PermissionRead  Permission = "read"
	PermissionWrite Permission = "write"
)

// Session represents an active sharing or live session across cast.
type Session struct {
	ID           string       `json:"id"`
	Code         string       `json:"code"`
	Type         Type         `json:"type"`
	DeviceID     string       `json:"device_id"`
	DeviceName   string       `json:"device_name"`
	Address      string       `json:"address"`
	LocalAddress string       `json:"local_address"`
	CreatedAt    time.Time    `json:"created_at"`
	ExpiresAt    time.Time    `json:"expires_at"`
	Permissions  []Permission `json:"permissions"`
}

// IsExpired reports whether the session's lifetime has elapsed.
func (s Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
