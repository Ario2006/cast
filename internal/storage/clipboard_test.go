package storage

import (
	"path/filepath"
	"testing"
	"time"
)

func TestClipboardStoreDeduplicatesAndTrims(t *testing.T) {
	store, err := OpenClipboardStore(filepath.Join(t.TempDir(), "history.db"), 2)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	now := time.Now()
	for _, content := range []string{"first", "first", "second", "third"} {
		if _, err := store.Add(content, now); err != nil {
			t.Fatalf("Add(%q) error = %v", content, err)
		}
	}
	entries, err := store.List(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].Content != "third" || entries[1].Content != "second" {
		t.Fatalf("List() = %#v", entries)
	}
}

func TestClipboardStoreClear(t *testing.T) {
	store, err := OpenClipboardStore(filepath.Join(t.TempDir(), "history.db"), 2)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if _, err := store.Add("item", time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := store.Clear(); err != nil {
		t.Fatal(err)
	}
	entries, err := store.List(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("List() returned %d entries, want 0", len(entries))
	}
}
