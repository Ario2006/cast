package share

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFileShareStreamsDownloadAndRequiresToken(t *testing.T) {
	sourcePath := filepath.Join(t.TempDir(), "report.txt")
	contents := "share this text safely"
	if err := os.WriteFile(sourcePath, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	server, err := StartFile(sourcePath, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop(context.Background())

	response, err := http.Get(strings.Split(server.Session().LocalURL, "?")[0])
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("unauthorized status = %d, want %d", response.StatusCode, http.StatusForbidden)
	}

	destination := t.TempDir()
	downloadedPath, err := Receive(context.Background(), ResolvedShare{URL: server.Session().LocalURL, Name: "report.txt"}, destination, nil)
	if err != nil {
		t.Fatal(err)
	}
	downloaded, err := os.ReadFile(downloadedPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(downloaded) != contents {
		t.Fatalf("downloaded file = %q, want %q", downloaded, contents)
	}
}

func TestReceiveNeverOverwrites(t *testing.T) {
	sourcePath := filepath.Join(t.TempDir(), "report.txt")
	if err := os.WriteFile(sourcePath, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	server, err := StartFile(sourcePath, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop(context.Background())

	destination := t.TempDir()
	if err := os.WriteFile(filepath.Join(destination, "report.txt"), []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = Receive(context.Background(), ResolvedShare{URL: server.Session().LocalURL, Name: "report.txt"}, destination, nil)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("Receive() error = %v", err)
	}
}

func TestExpiredShareIsGone(t *testing.T) {
	sourcePath := filepath.Join(t.TempDir(), "report.txt")
	if err := os.WriteFile(sourcePath, []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}
	server, err := StartFile(sourcePath, time.Nanosecond)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop(context.Background())
	time.Sleep(time.Millisecond)
	response, err := http.Get(server.Session().LocalURL)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	_, _ = io.ReadAll(response.Body)
	if response.StatusCode != http.StatusGone {
		t.Fatalf("expired status = %d, want %d", response.StatusCode, http.StatusGone)
	}
}

func TestResolveFindsActiveLocalShare(t *testing.T) {
	sourcePath := filepath.Join(t.TempDir(), "report.txt")
	if err := os.WriteFile(sourcePath, []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}
	server, err := StartFile(sourcePath, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop(context.Background())

	context, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resolved, err := Resolve(context, server.Session().Code)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.URL != server.Session().URL || resolved.Name != "report.txt" {
		t.Fatalf("Resolve() = %#v", resolved)
	}
}
