// Package identity manages the local device identity and cryptographic keypair.
package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Device represents the local device identity.
type Device struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

// PublicDevice contains non-sensitive device details safe to share across the network.
type PublicDevice struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

// Public returns public metadata without the private key.
func (d *Device) Public() PublicDevice {
	return PublicDevice{
		ID:        d.ID,
		Name:      d.Name,
		PublicKey: d.PublicKey,
	}
}

// LoadOrCreate retrieves the persistent local device identity, generating a new
// Ed25519 keypair on first run.
func LoadOrCreate(configDir string) (*Device, error) {
	if configDir == "" {
		return nil, fmt.Errorf("configuration directory cannot be empty")
	}
	identityDir := filepath.Join(configDir, "identity")
	devicePath := filepath.Join(identityDir, "device.json")

	data, err := os.ReadFile(devicePath)
	if err == nil {
		var dev Device
		if err := json.Unmarshal(data, &dev); err == nil && dev.ID != "" && dev.PublicKey != "" {
			return &dev, nil
		}
	}

	if err := os.MkdirAll(identityDir, 0o700); err != nil {
		return nil, fmt.Errorf("create identity directory: %w", err)
	}

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate device keypair: %w", err)
	}

	hash := sha256.Sum256(pubKey)
	id := fmt.Sprintf("%x", hash[:8])

	name := "Local Device"
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		name = hostname
	}

	dev := &Device{
		ID:         id,
		Name:       name,
		PublicKey:  base64.StdEncoding.EncodeToString(pubKey),
		PrivateKey: base64.StdEncoding.EncodeToString(privKey),
	}

	content, err := json.MarshalIndent(dev, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal device identity: %w", err)
	}

	if err := os.WriteFile(devicePath, content, 0o600); err != nil {
		return nil, fmt.Errorf("write device identity file: %w", err)
	}

	return dev, nil
}
