// Package logging configures cast's deliberately quiet diagnostic logger.
package logging

import (
	"io"
	"log/slog"
)

// New returns a logger that emits debug details only when explicitly requested.
// Command output remains separate from this logger so normal usage stays concise.
func New(writer io.Writer, debug bool) *slog.Logger {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{Level: level}))
}
