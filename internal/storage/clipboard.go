// Package storage contains cast's small local SQLite persistence layer.
package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// ClipboardEntry is a locally stored text clipboard item.
type ClipboardEntry struct {
	ID        int64
	Content   string
	CreatedAt time.Time
}

// ClipboardStore persists clipboard history and applies the configured limit.
type ClipboardStore struct {
	db  *sql.DB
	max int
}

// OpenClipboardStore opens a SQLite history database and applies its schema.
func OpenClipboardStore(path string, maxEntries int) (*ClipboardStore, error) {
	if maxEntries < 1 {
		return nil, fmt.Errorf("clipboard history limit must be at least one")
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open clipboard database: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS clipboard_entries (
			id INTEGER PRIMARY KEY,
			content TEXT NOT NULL,
			created_at INTEGER NOT NULL
		);`); err != nil {
		db.Close()
		return nil, fmt.Errorf("create clipboard schema: %w", err)
	}
	return &ClipboardStore{db: db, max: maxEntries}, nil
}

// Close closes the local database.
func (s *ClipboardStore) Close() error { return s.db.Close() }

// Add records content unless it exactly matches the newest entry. The return
// value reports whether a new history item was added.
func (s *ClipboardStore) Add(content string, at time.Time) (bool, error) {
	var latest string
	err := s.db.QueryRow(`SELECT content FROM clipboard_entries ORDER BY id DESC LIMIT 1`).Scan(&latest)
	if err == nil && latest == content {
		return false, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return false, fmt.Errorf("read latest clipboard entry: %w", err)
	}
	if _, err := s.db.Exec(`INSERT INTO clipboard_entries (content, created_at) VALUES (?, ?)`, content, at.UnixMilli()); err != nil {
		return false, fmt.Errorf("save clipboard entry: %w", err)
	}
	if _, err := s.db.Exec(`
		DELETE FROM clipboard_entries
		WHERE id NOT IN (
			SELECT id FROM clipboard_entries ORDER BY id DESC LIMIT ?
		)`, s.max); err != nil {
		return false, fmt.Errorf("trim clipboard history: %w", err)
	}
	return true, nil
}

// List returns the most recent entries first.
func (s *ClipboardStore) List(limit int) ([]ClipboardEntry, error) {
	if limit < 1 {
		return []ClipboardEntry{}, nil
	}
	rows, err := s.db.Query(`SELECT id, content, created_at FROM clipboard_entries ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("list clipboard entries: %w", err)
	}
	defer rows.Close()

	entries := make([]ClipboardEntry, 0, limit)
	for rows.Next() {
		var entry ClipboardEntry
		var createdAt int64
		if err := rows.Scan(&entry.ID, &entry.Content, &createdAt); err != nil {
			return nil, fmt.Errorf("read clipboard entry: %w", err)
		}
		entry.CreatedAt = time.UnixMilli(createdAt)
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate clipboard entries: %w", err)
	}
	return entries, nil
}

// Clear deletes every locally stored clipboard entry.
func (s *ClipboardStore) Clear() error {
	if _, err := s.db.Exec(`DELETE FROM clipboard_entries`); err != nil {
		return fmt.Errorf("clear clipboard entries: %w", err)
	}
	return nil
}
