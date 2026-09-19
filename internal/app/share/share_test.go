package share

import (
	"archive/zip"
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

func TestFolderShareCreatesTemporaryZipAndCleansUp(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("data1"), 0o600); err != nil {
		t.Fatal(err)
	}
	subDir := filepath.Join(dir, "sub")
	if err := os.MkdirAll(subDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("data2"), 0o600); err != nil {
		t.Fatal(err)
	}

	server, err := Start(dir, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	tempZipPath := server.tempPath
	if tempZipPath == "" {
		t.Fatal("server.tempPath is empty for directory share")
	}
	if _, err := os.Stat(tempZipPath); err != nil {
		t.Fatalf("temp zip file does not exist: %v", err)
	}

	destination := t.TempDir()
	receivedPath, err := Receive(context.Background(), ResolvedShare{
		URL:  server.Session().LocalURL,
		Name: server.Session().Name,
	}, destination, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(receivedPath, ".zip") {
		t.Fatalf("receivedPath = %q, want .zip extension", receivedPath)
	}

	zipReader, err := zip.OpenReader(receivedPath)
	if err != nil {
		t.Fatalf("open received zip: %v", err)
	}
	foundFile1, foundFile2 := false, false
	for _, f := range zipReader.File {
		if f.Name == "file1.txt" {
			foundFile1 = true
		}
		if f.Name == "sub/file2.txt" {
			foundFile2 = true
		}
	}
	_ = zipReader.Close()
	if !foundFile1 || !foundFile2 {
		t.Fatalf("zip contents missing files: file1=%v, file2=%v", foundFile1, foundFile2)
	}

	if err := server.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(tempZipPath); !os.IsNotExist(err) {
		t.Fatalf("temporary zip %q was not removed after Stop()", tempZipPath)
	}
}
