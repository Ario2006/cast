// Package qr renders compact terminal QR codes for cast connection payloads.
package qr

import (
	"io"

	"github.com/mdp/qrterminal/v4"
)

// Render writes a terminal-friendly QR code for payload.
func Render(payload string, writer io.Writer) {
	qrterminal.GenerateHalfBlock(payload, qrterminal.M, writer)
}
