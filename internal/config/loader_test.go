package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	if cfg.Clipboard.MaxHistory != 200 {
		t.Errorf("default MaxHistory = %d, want 200", cfg.Clipboard.MaxHistory)
	}
	if cfg.Share.ExpiresMinutes != 10 {
		t.Errorf("default ExpiresMinutes = %d, want 10", cfg.Share.ExpiresMinutes)
	}
	if cfg.Network.PreferredPort != 0 {
		t.Errorf("default PreferredPort = %d, want 0", cfg.Network.PreferredPort)
	}
}

func TestLoadWithCustomToml(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CAST_CONFIG_DIR", dir)

	tomlContent := `
# Sample configuration
[clipboard]
max_history = 350

[share]
expires_minutes = 25 # custom duration

[network]
preferred_port = 8888
`
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(tomlContent), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := Load()
	if cfg.Clipboard.MaxHistory != 350 {
		t.Errorf("loaded MaxHistory = %d, want 350", cfg.Clipboard.MaxHistory)
	}
	if cfg.Share.ExpiresMinutes != 25 {
		t.Errorf("loaded ExpiresMinutes = %d, want 25", cfg.Share.ExpiresMinutes)
	}
	if cfg.Network.PreferredPort != 8888 {
		t.Errorf("loaded PreferredPort = %d, want 8888", cfg.Network.PreferredPort)
	}
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CAST_CONFIG_DIR", dir)

	cfg := Load()
	def := Default()
	if cfg != def {
		t.Errorf("Load() on missing file = %+v, want %+v", cfg, def)
	}
}
