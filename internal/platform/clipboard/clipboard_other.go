//go:build !darwin

package clipboard

import "fmt"

// System is intentionally explicit until a tested adapter is added for this OS.
type System struct{}

func (System) Read() (string, error) {
	return "", fmt.Errorf("clipboard access is not yet supported on this platform")
}

func (System) Write(string) error {
	return fmt.Errorf("clipboard access is not yet supported on this platform")
}
