package live

import (
	"context"
	"io"
	"net"
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

func TestLiveProxyForwardsRequestsWithoutToken(t *testing.T) {
	backend := &http.Server{Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api" || request.URL.Query().Get("q") != "cast" || request.URL.Query().Get("token") != "" {
			http.Error(writer, "unexpected forwarded request", http.StatusBadRequest)
			return
		}
		_, _ = writer.Write([]byte("proxied response"))
	})}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go backend.Serve(listener)
	defer backend.Close()

	server, err := StartProxy(listener.Addr().(*net.TCPAddr).Port, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop(context.Background())
	proxyURL := addPath(server.Session().LocalURL, "api") + "&q=cast"
	response, err := http.Get(proxyURL)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)
	response.Body.Close()
	if response.StatusCode != http.StatusOK || string(body) != "proxied response" {
		t.Fatalf("proxy response = %d, %q", response.StatusCode, body)
	}

	withoutToken := strings.Split(server.Session().LocalURL, "?")[0]
	response, err = http.Get(withoutToken)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("missing token status = %d", response.StatusCode)
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

func TestLiveResolveFindsActiveDirectorySession(t *testing.T) {
	root := t.TempDir()
	server, err := StartDirectory(root, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resolved, err := Resolve(ctx, server.Session().Code)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if resolved.Code != server.Session().Code || resolved.URL != server.Session().URL {
		t.Fatalf("Resolve() = %+v, want URL %s", resolved, server.Session().URL)
	}
	if resolved.Project != filepath.Base(root) {
		t.Fatalf("Resolve() Project = %q, want %q", resolved.Project, filepath.Base(root))
	}
	if resolved.Device == "" {
		t.Fatal("Resolve() Device is empty")
	}
}

func TestLiveResolveFindsActiveProxySession(t *testing.T) {
	backend := &http.Server{Handler: http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte("ok"))
	})}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go backend.Serve(listener)
	defer backend.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	server, err := StartProxy(port, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resolved, err := Resolve(ctx, server.Session().Code)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if resolved.Code != server.Session().Code || resolved.URL != server.Session().URL {
		t.Fatalf("Resolve() = %+v", resolved)
	}
}

func TestLiveResolveRejectsInvalidAndExpiredCode(t *testing.T) {
	// Invalid code
	_, err := Resolve(context.Background(), "??")
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("Resolve(??) error = %v, want invalid code error", err)
	}

	// Expired session
	root := t.TempDir()
	server, err := StartDirectory(root, time.Nanosecond)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Stop(context.Background())
	time.Sleep(2 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = Resolve(ctx, server.Session().Code)
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("Resolve(expired) error = %v, want expired error", err)
	}
}
