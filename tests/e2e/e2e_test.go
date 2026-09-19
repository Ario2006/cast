package e2e

import (
	"archive/zip"
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

var codeRegex = regexp.MustCompile(`Code:\s+([2-9A-HJ-NP-Z]{4,8})|Live code:\s+([2-9A-HJ-NP-Z]{4,8})`)

func buildBinary(t *testing.T) string {
	t.Helper()
	binDir := t.TempDir()
	binPath := filepath.Join(binDir, "cast")

	cmd := exec.Command("go", "build", "-o", binPath, "../../cmd/cast")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build cast binary failed: %v\n%s", err, string(output))
	}
	return binPath
}

func readCodeFromOutput(r io.Reader, timeout time.Duration) (string, error) {
	lines := make(chan string, 100)
	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		close(lines)
	}()

	deadline := time.After(timeout)
	expectCode := false
	for {
		select {
		case line, ok := <-lines:
			if !ok {
				return "", fmt.Errorf("process closed stdout before emitting session code")
			}
			trimmed := strings.TrimSpace(line)
			if trimmed == "Code:" || trimmed == "Live code:" {
				expectCode = true
				continue
			}
			if expectCode && len(trimmed) >= 4 && len(trimmed) <= 8 {
				return trimmed, nil
			}
			if strings.HasPrefix(line, "  cast receive ") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					return parts[2], nil
				}
			}
			if strings.HasPrefix(line, "  cast connect ") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					return parts[2], nil
				}
			}
		case <-deadline:
			return "", fmt.Errorf("timed out waiting for session code")
		}
	}
}

func TestTwoProcessFileShareAndReceive(t *testing.T) {
	binary := buildBinary(t)
	configDir := t.TempDir()

	// Prepare source file
	sourceDir := t.TempDir()
	sourceFile := filepath.Join(sourceDir, "sample.txt")
	originalContent := "Hello from two-process cast end-to-end test!\nDeterministic content."
	if err := os.WriteFile(sourceFile, []byte(originalContent), 0o600); err != nil {
		t.Fatal(err)
	}
	expectedHash := sha256.Sum256([]byte(originalContent))

	// Start Process A: cast share <file>
	shareCmd := exec.Command(binary, "share", sourceFile)
	shareCmd.Env = append(os.Environ(), "CAST_CONFIG_DIR="+configDir)
	stdout, err := shareCmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := shareCmd.Start(); err != nil {
		t.Fatalf("start share process: %v", err)
	}
	defer func() {
		if shareCmd.Process != nil {
			_ = shareCmd.Process.Kill()
		}
	}()

	// Extract short code
	code, err := readCodeFromOutput(stdout, 5*time.Second)
	if err != nil {
		t.Fatalf("read code: %v", err)
	}

	// Start Process B: cast receive <code> in a separate working directory
	receiveDir := t.TempDir()
	receiveCmd := exec.Command(binary, "receive", code)
	receiveCmd.Dir = receiveDir
	receiveCmd.Env = append(os.Environ(), "CAST_CONFIG_DIR="+configDir)
	output, err := receiveCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("receive process failed: %v\noutput: %s", err, string(output))
	}

	// Verify file was received with correct content and SHA-256
	receivedFile := filepath.Join(receiveDir, "sample.txt")
	receivedBytes, err := os.ReadFile(receivedFile)
	if err != nil {
		t.Fatalf("read received file: %v", err)
	}
	actualHash := sha256.Sum256(receivedBytes)
	if actualHash != expectedHash {
		t.Fatalf("hash mismatch: got %x, want %x", actualHash, expectedHash)
	}
	if string(receivedBytes) != originalContent {
		t.Fatalf("content mismatch: got %q, want %q", string(receivedBytes), originalContent)
	}
}

func TestTwoProcessFolderShareAndReceive(t *testing.T) {
	binary := buildBinary(t)
	configDir := t.TempDir()

	// Prepare source folder
	sourceDir := filepath.Join(t.TempDir(), "project-docs")
	if err := os.MkdirAll(filepath.Join(sourceDir, "nested"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "doc.txt"), []byte("root doc"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "nested", "sub.txt"), []byte("sub doc"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Start Process A: cast share <folder>
	shareCmd := exec.Command(binary, "share", sourceDir)
	shareCmd.Env = append(os.Environ(), "CAST_CONFIG_DIR="+configDir)
	stdout, err := shareCmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := shareCmd.Start(); err != nil {
		t.Fatalf("start share process: %v", err)
	}
	defer func() {
		if shareCmd.Process != nil {
			_ = shareCmd.Process.Kill()
		}
	}()

	code, err := readCodeFromOutput(stdout, 5*time.Second)
	if err != nil {
		t.Fatalf("read code: %v", err)
	}

	// Start Process B: cast receive <code>
	receiveDir := t.TempDir()
	receiveCmd := exec.Command(binary, "receive", code)
	receiveCmd.Dir = receiveDir
	receiveCmd.Env = append(os.Environ(), "CAST_CONFIG_DIR="+configDir)
	output, err := receiveCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("receive process failed: %v\noutput: %s", err, string(output))
	}

	receivedZip := filepath.Join(receiveDir, "project-docs.zip")
	zipReader, err := zip.OpenReader(receivedZip)
	if err != nil {
		t.Fatalf("open received zip: %v", err)
	}
	defer zipReader.Close()

	foundRootDoc, foundSubDoc := false, false
	for _, f := range zipReader.File {
		if f.Name == "doc.txt" {
			foundRootDoc = true
		}
		if f.Name == "nested/sub.txt" {
			foundSubDoc = true
		}
	}
	if !foundRootDoc || !foundSubDoc {
		t.Fatalf("zip missing expected files: rootDoc=%v, subDoc=%v", foundRootDoc, foundSubDoc)
	}
}

func TestTwoProcessLiveAndConnect(t *testing.T) {
	binary := buildBinary(t)
	configDir := t.TempDir()

	// Prepare project
	projectDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, "index.html"), []byte("<h1>Cast Live Web</h1>"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Start Process A: cast live <projectDir>
	liveCmd := exec.Command(binary, "live", projectDir)
	liveCmd.Env = append(os.Environ(), "CAST_CONFIG_DIR="+configDir)
	stdout, err := liveCmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := liveCmd.Start(); err != nil {
		t.Fatalf("start live process: %v", err)
	}
	defer func() {
		if liveCmd.Process != nil {
			_ = liveCmd.Process.Kill()
		}
	}()

	code, err := readCodeFromOutput(stdout, 5*time.Second)
	if err != nil {
		t.Fatalf("read live code: %v", err)
	}

	// Start Process B: cast connect <code>
	connectCmd := exec.Command(binary, "connect", code)
	connectCmd.Env = append(os.Environ(), "CAST_CONFIG_DIR="+configDir)
	connectOutput, err := connectCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("connect process failed: %v\noutput: %s", err, string(connectOutput))
	}

	outStr := string(connectOutput)
	if !strings.Contains(outStr, "Connected to:") || !strings.Contains(outStr, "Live project:") || !strings.Contains(outStr, "URL:") {
		t.Fatalf("connect output missing expected fields:\n%s", outStr)
	}

	// Extract URL from output and verify HTTP access
	urlRegex := regexp.MustCompile(`URL:\s+(http://[^\s]+)`)
	matches := urlRegex.FindStringSubmatch(outStr)
	if len(matches) < 2 {
		t.Fatalf("could not parse URL from connect output: %s", outStr)
	}
	remoteURL := matches[1]

	parsedURL, err := url.Parse(remoteURL)
	if err != nil {
		t.Fatalf("parse live URL: %v", err)
	}
	parsedURL.Path = strings.TrimSuffix(parsedURL.Path, "/") + "/index.html"

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(parsedURL.String())
	if err != nil {
		t.Fatalf("GET remote live URL: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || string(body) != "<h1>Cast Live Web</h1>" {
		t.Fatalf("unexpected live response: status=%d, body=%q", resp.StatusCode, string(body))
	}
}
