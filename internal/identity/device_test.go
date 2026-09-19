package identity

import (
	"crypto/ed25519"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOrCreateInitializesAndPersists(t *testing.T) {
	dir := t.TempDir()

	device, err := LoadOrCreate(dir)
	if err != nil {
		t.Fatalf("LoadOrCreate() error = %v", err)
	}

	if device.ID == "" || device.Name == "" || device.PublicKey == "" || device.PrivateKey == "" {
		t.Fatalf("LoadOrCreate() incomplete device = %+v", device)
	}

	pubBytes, err := base64.StdEncoding.DecodeString(device.PublicKey)
	if err != nil || len(pubBytes) != ed25519.PublicKeySize {
		t.Fatalf("invalid public key = %q", device.PublicKey)
	}

	privBytes, err := base64.StdEncoding.DecodeString(device.PrivateKey)
	if err != nil || len(privBytes) != ed25519.PrivateKeySize {
		t.Fatalf("invalid private key")
	}

	// Verify file permissions
	info, err := os.Stat(filepath.Join(dir, "identity", "device.json"))
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("device.json perm = %o, want %o", perm, 0o600)
	}

	// Loading again should return identical device
	device2, err := LoadOrCreate(dir)
	if err != nil {
		t.Fatalf("second LoadOrCreate() error = %v", err)
	}
	if device2.ID != device.ID || device2.PublicKey != device.PublicKey {
		t.Fatalf("device mismatch: %+v vs %+v", device2, device)
	}

	public := device.Public()
	if public.ID != device.ID || public.Name != device.Name || public.PublicKey != device.PublicKey {
		t.Fatalf("Public() = %+v", public)
	}
}
