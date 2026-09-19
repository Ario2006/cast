// Package config provides the small, local configuration footprint used by cast.
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const appName = "cast"

// Dir returns cast's platform-appropriate configuration directory. CAST_CONFIG_DIR
// is intentionally supported to make local development and tests self-contained.
func Dir() (string, error) {
	if dir := os.Getenv("CAST_CONFIG_DIR"); dir != "" {
		return dir, nil
	}

	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("find the user configuration directory: %w", err)
	}
	return filepath.Join(base, appName), nil
}

// EnsureDir creates the configuration directory with owner-only permissions where
// the operating system honours them.
func EnsureDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create configuration directory %q: %w", dir, err)
	}
	return dir, nil
}

// FilePath is the default location for future user configuration.
func FilePath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}
