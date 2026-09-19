package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// Config holds user-configurable options with safe defaults.
type Config struct {
	Clipboard ClipboardConfig
	Share     ShareConfig
	Network   NetworkConfig
}

// ClipboardConfig defines options for clipboard history.
type ClipboardConfig struct {
	MaxHistory int
}

// ShareConfig defines options for file and project sharing.
type ShareConfig struct {
	ExpiresMinutes int
}

// NetworkConfig defines networking preferences.
type NetworkConfig struct {
	PreferredPort int
}

// Default returns standard out-of-the-box settings.
func Default() Config {
	return Config{
		Clipboard: ClipboardConfig{MaxHistory: 200},
		Share:     ShareConfig{ExpiresMinutes: 10},
		Network:   NetworkConfig{PreferredPort: 0},
	}
}

// Load reads config.toml from the configuration directory if present,
// falling back to defaults for any missing or unconfigured values.
func Load() Config {
	cfg := Default()
	filePath, err := FilePath()
	if err != nil {
		return cfg
	}
	file, err := os.Open(filePath)
	if err != nil {
		return cfg
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentSection := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.ToLower(strings.Trim(line[1:len(line)-1], " \t"))
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		valStr := strings.TrimSpace(parts[1])
		if idx := strings.IndexAny(valStr, "#;"); idx != -1 {
			valStr = strings.TrimSpace(valStr[:idx])
		}
		valStr = strings.Trim(valStr, `"'`)

		valInt, err := strconv.Atoi(valStr)
		if err != nil {
			continue
		}

		switch currentSection {
		case "clipboard":
			if key == "max_history" && valInt > 0 {
				cfg.Clipboard.MaxHistory = valInt
			}
		case "share":
			if key == "expires_minutes" && valInt > 0 {
				cfg.Share.ExpiresMinutes = valInt
			}
		case "network":
			if key == "preferred_port" && valInt >= 0 && valInt <= 65535 {
				cfg.Network.PreferredPort = valInt
			}
		}
	}
	return cfg
}
