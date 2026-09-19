package clipboard

import (
	"testing"
	"time"

	"github.com/aryankumar/cast/internal/storage"
)

type fakeStore struct {
	entries []storage.ClipboardEntry
}

func (s *fakeStore) Add(content string, at time.Time) (bool, error) {
	if len(s.entries) > 0 && s.entries[0].Content == content {
		return false, nil
	}
	s.entries = append([]storage.ClipboardEntry{{Content: content, CreatedAt: at}}, s.entries...)
	return true, nil
}
func (s *fakeStore) List(limit int) ([]storage.ClipboardEntry, error) { return s.entries, nil }
func (s *fakeStore) Clear() error                                     { s.entries = nil; return nil }

type fakeClipboard struct {
	content string
	written string
}

func (c *fakeClipboard) Read() (string, error) { return c.content, nil }
func (c *fakeClipboard) Write(content string) error {
	c.written = content
	return nil
}

func TestCaptureDeduplicatesAndRestores(t *testing.T) {
	store := &fakeStore{}
	clipboard := &fakeClipboard{content: "copied text"}
	service := NewService(store, clipboard)

	added, skipped, err := service.Capture()
	if err != nil || !added || skipped {
		t.Fatalf("Capture() = (%t, %t, %v)", added, skipped, err)
	}
	added, _, err = service.Capture()
	if err != nil || added {
		t.Fatalf("second Capture() = (%t, %v)", added, err)
	}
	if err := service.Restore(store.entries[0]); err != nil {
		t.Fatal(err)
	}
	if clipboard.written != "copied text" {
		t.Fatalf("Write() = %q", clipboard.written)
	}
}

func TestCaptureSkipsObviousSecrets(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, &fakeClipboard{content: "Authorization: Bearer example-token"})

	added, skipped, err := service.Capture()
	if err != nil || added || !skipped {
		t.Fatalf("Capture() = (%t, %t, %v)", added, skipped, err)
	}
}
