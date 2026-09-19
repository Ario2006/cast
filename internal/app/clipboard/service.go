// Package clipboard coordinates local clipboard access and persistent history.
package clipboard

import (
	"context"
	"strings"
	"time"

	platform "github.com/aryankumar/cast/internal/platform/clipboard"
	"github.com/aryankumar/cast/internal/storage"
)

// HistoryStore is the persistence boundary used by Service.
type HistoryStore interface {
	Add(content string, at time.Time) (bool, error)
	List(limit int) ([]storage.ClipboardEntry, error)
	Clear() error
}

// Service keeps text clipboard history local to the current device.
type Service struct {
	store     HistoryStore
	clipboard platform.Clipboard
	now       func() time.Time
}

// NewService creates a local-only clipboard history service.
func NewService(store HistoryStore, clipboard platform.Clipboard) *Service {
	return &Service{store: store, clipboard: clipboard, now: time.Now}
}

// Capture reads the current clipboard and stores it if it is text worth keeping.
// Clipboard contents are deliberately never logged.
func (s *Service) Capture() (added bool, skippedSensitive bool, err error) {
	content, err := s.clipboard.Read()
	if err != nil {
		return false, false, err
	}
	if content == "" || looksSensitive(content) {
		return false, content != "", nil
	}
	added, err = s.store.Add(content, s.now())
	return added, false, err
}

// Watch captures clipboard changes on a best-effort interval until the context
// is cancelled. Errors are returned only when the watcher cannot start; polling
// errors are intentionally ignored to keep a transient clipboard issue quiet.
func (s *Service) Watch(ctx context.Context, interval time.Duration) error {
	if interval <= 0 {
		return nil
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			_, _, _ = s.Capture()
		}
	}
}

// List returns recent entries in reverse chronological order.
func (s *Service) List(limit int) ([]storage.ClipboardEntry, error) { return s.store.List(limit) }

// Restore copies a selected entry back to the system clipboard.
func (s *Service) Restore(entry storage.ClipboardEntry) error {
	return s.clipboard.Write(entry.Content)
}

// Clear deletes local history. It does not alter the system clipboard.
func (s *Service) Clear() error { return s.store.Clear() }

func looksSensitive(content string) bool {
	normalized := strings.ToLower(content)
	for _, marker := range []string{"begin private key", "password=", "api_key=", "authorization: bearer", "bearer "} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}
