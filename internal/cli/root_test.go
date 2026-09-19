package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	liveapp "github.com/aryankumar/cast/internal/app/live"
	"github.com/aryankumar/cast/internal/apperror"
)

func TestVersionCommand(t *testing.T) {
	var out, errOut bytes.Buffer
	cmd := NewRootCommand(Options{Version: "v0.1.0", Out: &out, ErrOut: &errOut})
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := out.String(); got != "v0.1.0\n" {
		t.Fatalf("version output = %q", got)
	}
}

func TestVersionFlag(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(Options{Version: "v0.1.0", Out: &out})
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := out.String(); got != "cast version v0.1.0\n" {
		t.Fatalf("version flag output = %q", got)
	}
}

func TestPrintErrorIncludesHint(t *testing.T) {
	var output bytes.Buffer
	PrintError(&output, apperror.New("Could not find a share.", "Check the code and try again."))

	if !strings.Contains(output.String(), "Check the code") {
		t.Fatalf("PrintError() = %q", output.String())
	}
}

func TestCalcCommand(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(Options{Out: &out})
	cmd.SetArgs([]string{"calc", "(12 + 8) * 3"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := out.String(); got != "60\n" {
		t.Fatalf("calc output = %q", got)
	}
}

func TestConvCommandWithDirectTarget(t *testing.T) {
	var out bytes.Buffer
	cmd := NewRootCommand(Options{Out: &out})
	cmd.SetArgs([]string{"conv", "10", "km", "mi"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := out.String(); got != "10 km = 6.2137119223733395 mi\n" {
		t.Fatalf("conv output = %q", got)
	}
}

func TestClipClearCommand(t *testing.T) {
	t.Setenv("CAST_CONFIG_DIR", t.TempDir())
	var out bytes.Buffer
	cmd := NewRootCommand(Options{Out: &out})
	cmd.SetArgs([]string{"clip", "clear"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got := out.String(); got != "Clipboard history cleared.\n" {
		t.Fatalf("clip clear output = %q", got)
	}
}

func TestConnectCommand(t *testing.T) {
	root := t.TempDir()
	server, err := liveapp.StartDirectory(root, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop(context.Background())

	var out bytes.Buffer
	cmd := NewRootCommand(Options{Out: &out})
	cmd.SetArgs([]string{"connect", server.Session().Code})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "Connected to:") || !strings.Contains(output, "URL:") || !strings.Contains(output, server.Session().URL) {
		t.Fatalf("connect command output = %q", output)
	}
}
