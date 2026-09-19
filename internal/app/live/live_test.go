package live

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLiveDirectoryServesFilesAndListsDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "hello.txt"), []byte("hello from cast"), 0o600); err != nil {
		t.Fatal(err)
	}
	server, err := StartDirectory(root, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop(context.Background())

	response, err := http.Get(server.Session().LocalURL)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)
	response.Body.Close()
	if response.StatusCode != http.StatusOK || !strings.Contains(string(body), "hello.txt") {
		t.Fatalf("directory response = %d, %q", response.StatusCode, body)
	}

	fileURL := addPath(server.Session().LocalURL, "hello.txt")
	response, err = http.Get(fileURL)
	if err != nil {
		t.Fatal(err)
	}
	body, _ = io.ReadAll(response.Body)
	response.Body.Close()
	if response.StatusCode != http.StatusOK || string(body) != "hello from cast" {
		t.Fatalf("file response = %d, %q", response.StatusCode, body)
	}
}

func TestLiveDirectoryRejectsMissingTokenAndEscapingSymlink(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(outside, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "outside.txt")); err != nil {
		t.Fatal(err)
	}
	server, err := StartDirectory(root, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop(context.Background())

	withoutToken := strings.Split(server.Session().LocalURL, "?")[0]
	response, err := http.Get(withoutToken)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("missing token status = %d", response.StatusCode)
	}

	response, err = http.Get(addPath(server.Session().LocalURL, "outside.txt"))
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("symlink escape status = %d", response.StatusCode)
	}
}

func TestLiveDirectoryExpires(t *testing.T) {
	root := t.TempDir()
	server, err := StartDirectory(root, time.Nanosecond)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop(context.Background())
	time.Sleep(time.Millisecond)
	response, err := http.Get(server.Session().LocalURL)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusGone {
		t.Fatalf("expired status = %d", response.StatusCode)
	}
}

func addPath(rawURL, name string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	parsed.Path += "/" + name
	return parsed.String()
}
