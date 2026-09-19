package config

import (
	"path/filepath"
	"testing"
)

func TestDirUsesOverride(t *testing.T) {
	want := filepath.Join(t.TempDir(), "cast-config")
	t.Setenv("CAST_CONFIG_DIR", want)

	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir() error = %v", err)
	}
	if dir != want {
		t.Fatalf("Dir() = %q, want %q", dir, want)
	}
}

func TestEnsureDirCreatesConfiguredDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "cast")
	t.Setenv("CAST_CONFIG_DIR", dir)

	got, err := EnsureDir()
	if err != nil {
		t.Fatalf("EnsureDir() error = %v", err)
	}
	if got != dir {
		t.Fatalf("EnsureDir() = %q, want %q", got, dir)
	}
}
