//go:build darwin

package clipboard

import (
	"fmt"
	"os/exec"
	"strings"
)

// System reads and writes the macOS text clipboard using the built-in tools.
type System struct{}

func (System) Read() (string, error) {
	output, err := exec.Command("pbpaste").Output()
	if err != nil {
		return "", fmt.Errorf("read macOS clipboard: %w", err)
	}
	return string(output), nil
}

func (System) Write(text string) error {
	command := exec.Command("pbcopy")
	command.Stdin = strings.NewReader(text)
	if err := command.Run(); err != nil {
		return fmt.Errorf("write macOS clipboard: %w", err)
	}
	return nil
}
