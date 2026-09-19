// Package share implements small, local-first HTTP file sharing sessions.
package share

import (
	"archive/zip"
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	netplatform "github.com/aryankumar/cast/internal/platform/net"
	"github.com/aryankumar/cast/internal/transport/discovery"
)

const discoveryPort = discovery.SharePort

// Session describes an active, read-only file share. URL contains a
// high-entropy authorization token; Code is a human-friendly discovery key.
type Session struct {
	Code      string
	Name      string
	Size      int64
	CreatedAt time.Time
	ExpiresAt time.Time
	URL       string
	LocalURL  string
}

// Server owns the HTTP and UDP listeners for one active file share.
type Server struct {
	session  Session
	token    string
	path     string
	tempPath string
	http     *http.Server
	listener net.Listener
	udp      *net.UDPConn
	stopOnce sync.Once
}

// StartFile starts an expiring read-only HTTP share for a regular file.
func StartFile(path string, lifetime time.Duration) (*Server, error) {
	return Start(path, lifetime)
}

// Start starts an expiring read-only HTTP share for a regular file or directory.
// When sharing a directory, a temporary ZIP archive is created and automatically
// removed when the server stops or expires.
func Start(path string, lifetime time.Duration) (*Server, error) {
	if lifetime <= 0 {
		return nil, fmt.Errorf("share lifetime must be positive")
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("inspect shared target: %w", err)
	}
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve shared target path: %w", err)
	}

	var (
		sharePath string
		tempPath  string
		name      string
		size      int64
	)

	if info.IsDir() {
		tempZip, zipSize, err := createTempZip(absolutePath)
		if err != nil {
			return nil, err
		}
		sharePath = tempZip
		tempPath = tempZip
		name = filepath.Base(absolutePath) + ".zip"
		size = zipSize
	} else if info.Mode().IsRegular() {
		sharePath = absolutePath
		name = info.Name()
		size = info.Size()
	} else {
		return nil, fmt.Errorf("only regular files and directories can be shared")
	}

	code, err := randomCode(6)
	if err != nil {
		if tempPath != "" {
			_ = os.Remove(tempPath)
		}
		return nil, err
	}
	token, err := randomToken()
	if err != nil {
		if tempPath != "" {
			_ = os.Remove(tempPath)
		}
		return nil, err
	}
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		if tempPath != "" {
			_ = os.Remove(tempPath)
		}
		return nil, fmt.Errorf("listen for file share: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	createdAt := time.Now()
	server := &Server{
		path:     sharePath,
		tempPath: tempPath,
		token:    token,
		listener: listener,
		session: Session{
			Code:      code,
			Name:      name,
			Size:      size,
			CreatedAt: createdAt,
			ExpiresAt: createdAt.Add(lifetime),
			URL:       sessionURL(localIPv4(), port, code, token),
			LocalURL:  sessionURL(net.IPv4(127, 0, 0, 1), port, code, token),
		},
	}
	server.http = &http.Server{Handler: http.HandlerFunc(server.serveDownload)}
	go func() { _ = server.http.Serve(listener) }()
	if err := server.startDiscovery(); err != nil {
		server.Stop(context.Background())
		return nil, err
	}
	return server, nil
}

// Session returns immutable information for presenting this share.
func (s *Server) Session() Session { return s.session }

// Stop closes listeners and waits briefly for any active HTTP requests.
func (s *Server) Stop(ctx context.Context) error {
	var stopErr error
	s.stopOnce.Do(func() {
		if s.tempPath != "" {
			_ = os.Remove(s.tempPath)
		}
		if s.udp != nil {
			_ = s.udp.Close()
		}
		stopErr = s.http.Shutdown(ctx)
	})
	return stopErr
}

func (s *Server) serveDownload(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet || request.URL.Path != "/s/"+s.session.Code {
		http.NotFound(writer, request)
		return
	}
	if time.Now().After(s.session.ExpiresAt) {
		if s.tempPath != "" {
			_ = os.Remove(s.tempPath)
		}
		http.Error(writer, "This share has expired.", http.StatusGone)
		return
	}
	providedToken := request.URL.Query().Get("token")
	if subtle.ConstantTimeCompare([]byte(providedToken), []byte(s.token)) != 1 {
		http.Error(writer, "A valid share token is required.", http.StatusForbidden)
		return
	}
	file, err := os.Open(s.path)
	if err != nil {
		http.Error(writer, "The shared file is no longer available.", http.StatusNotFound)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		http.Error(writer, "The shared file is no longer available.", http.StatusNotFound)
		return
	}
	contentType := mime.TypeByExtension(filepath.Ext(s.session.Name))
	if contentType == "" && strings.HasSuffix(s.session.Name, ".zip") {
		contentType = "application/zip"
	}
	if contentType != "" {
		writer.Header().Set("Content-Type", contentType)
	}
	writer.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	writer.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": s.session.Name}))
	_, _ = io.Copy(writer, file)
}

func (s *Server) startDiscovery() error {
	connection, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: discoveryPort})
	if err != nil {
		return fmt.Errorf("listen for local share discovery: %w", err)
	}
	s.udp = connection
	go func() {
		buffer := make([]byte, 1024)
		for {
			count, sender, err := connection.ReadFromUDP(buffer)
			if err != nil {
				return
			}
			var request discoveryRequest
			if json.Unmarshal(buffer[:count], &request) != nil || request.Type != "cast-share-discover" || request.Code != s.session.Code {
				continue
			}
			response, _ := json.Marshal(discoveryResponse{Type: "cast-share", Code: s.session.Code, URL: s.session.URL, Name: s.session.Name, ExpiresAt: s.session.ExpiresAt})
			_, _ = connection.WriteToUDP(response, sender)
		}
	}()
	return nil
}

type discoveryRequest struct {
	Type string `json:"type"`
	Code string `json:"code"`
}

type discoveryResponse struct {
	Type      string    `json:"type"`
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	Name      string    `json:"name"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ResolvedShare is the result of finding a share by its short code.
type ResolvedShare struct {
	URL       string
	Name      string
	ExpiresAt time.Time
}

// Resolve broadcasts a short-code lookup over the local network and also probes
// loopback, which supports two cast processes running on the same machine.
func Resolve(ctx context.Context, code string) (ResolvedShare, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if !validCode(code) {
		return ResolvedShare{}, fmt.Errorf("invalid share code")
	}
	connection, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return ResolvedShare{}, fmt.Errorf("start local share discovery: %w", err)
	}
	defer connection.Close()
	if err := netplatform.EnableBroadcast(connection); err != nil {
		return ResolvedShare{}, fmt.Errorf("enable local share discovery: %w", err)
	}
	payload, _ := json.Marshal(discoveryRequest{Type: "cast-share-discover", Code: code})
	for _, target := range []*net.UDPAddr{
		{IP: net.IPv4(127, 0, 0, 1), Port: discoveryPort},
		{IP: net.IPv4bcast, Port: discoveryPort},
	} {
		_, _ = connection.WriteToUDP(payload, target)
	}

	deadline := time.Now().Add(3 * time.Second)
	if requestedDeadline, ok := ctx.Deadline(); ok && requestedDeadline.Before(deadline) {
		deadline = requestedDeadline
	}
	if err := connection.SetReadDeadline(deadline); err != nil {
		return ResolvedShare{}, fmt.Errorf("set share discovery deadline: %w", err)
	}
	buffer := make([]byte, 2048)
	for {
		count, _, err := connection.ReadFromUDP(buffer)
		if err != nil {
			if ctx.Err() != nil {
				return ResolvedShare{}, ctx.Err()
			}
			return ResolvedShare{}, fmt.Errorf("share code was not found on the local network")
		}
		var response discoveryResponse
		if json.Unmarshal(buffer[:count], &response) != nil || response.Type != "cast-share" || response.Code != code || response.URL == "" {
			continue
		}
		if time.Now().After(response.ExpiresAt) {
			return ResolvedShare{}, fmt.Errorf("share has expired")
		}
		return ResolvedShare{URL: response.URL, Name: response.Name, ExpiresAt: response.ExpiresAt}, nil
	}
}

// Receive downloads a resolved share into destinationDir without ever
// overwriting an existing file. Progress receives byte counts as data streams.
func Receive(ctx context.Context, share ResolvedShare, destinationDir string, progress func(received, total int64)) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, share.URL, nil)
	if err != nil {
		return "", fmt.Errorf("create download request: %w", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("connect to sharing device: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("sharing device returned %s", response.Status)
	}
	name := downloadedName(response, share.Name)
	path := filepath.Join(destinationDir, name)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if os.IsExist(err) {
			return "", fmt.Errorf("%s already exists", name)
		}
		return "", fmt.Errorf("create downloaded file: %w", err)
	}
	success := false
	defer func() {
		_ = file.Close()
		if !success {
			_ = os.Remove(path)
		}
	}()
	reader := &progressReader{reader: response.Body, total: response.ContentLength, report: progress}
	if _, err := io.Copy(file, reader); err != nil {
		return "", fmt.Errorf("download shared file: %w", err)
	}
	success = true
	if progress != nil {
		progress(reader.read, response.ContentLength)
	}
	return path, nil
}

type progressReader struct {
	reader io.Reader
	total  int64
	read   int64
	report func(received, total int64)
}

func (r *progressReader) Read(buffer []byte) (int, error) {
	count, err := r.reader.Read(buffer)
	if count > 0 {
		r.read += int64(count)
		if r.report != nil {
			r.report(r.read, r.total)
		}
	}
	return count, err
}

func downloadedName(response *http.Response, fallback string) string {
	if disposition := response.Header.Get("Content-Disposition"); disposition != "" {
		if _, params, err := mime.ParseMediaType(disposition); err == nil && params["filename"] != "" {
			fallback = params["filename"]
		}
	}
	name := filepath.Base(filepath.Clean(fallback))
	if name == "." || name == string(filepath.Separator) || name == "" {
		return "download"
	}
	return name
}

func randomCode(length int) (string, error) {
	return discovery.RandomCode(length)
}

func randomToken() (string, error) {
	return discovery.RandomToken()
}

func validCode(code string) bool {
	return discovery.ValidCode(code)
}

func sessionURL(address net.IP, port int, code, token string) string {
	value := url.URL{Scheme: "http", Host: net.JoinHostPort(address.String(), fmt.Sprintf("%d", port)), Path: "/s/" + code}
	query := value.Query()
	query.Set("token", token)
	value.RawQuery = query.Encode()
	return value.String()
}

func localIPv4() net.IP {
	return discovery.LocalIPv4()
}

func createTempZip(sourceDir string) (string, int64, error) {
	tempFile, err := os.CreateTemp("", "cast-share-*.zip")
	if err != nil {
		return "", 0, fmt.Errorf("create temporary zip file: %w", err)
	}

	zipWriter := zip.NewWriter(tempFile)
	realRoot, err := filepath.EvalSymlinks(sourceDir)
	if err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
		return "", 0, fmt.Errorf("resolve directory symlinks: %w", err)
	}

	walkErr := filepath.Walk(sourceDir, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		realPath, err := filepath.EvalSymlinks(filePath)
		if err != nil {
			return nil
		}
		relToReal, err := filepath.Rel(realRoot, realPath)
		if err != nil || relToReal == ".." || strings.HasPrefix(relToReal, ".."+string(filepath.Separator)) {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, filePath)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)
		if fileInfo.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if fileInfo.IsDir() {
			return nil
		}
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	if walkErr != nil {
		_ = zipWriter.Close()
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
		return "", 0, fmt.Errorf("compress directory: %w", walkErr)
	}
	if err := zipWriter.Close(); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
		return "", 0, fmt.Errorf("close zip writer: %w", err)
	}
	_ = tempFile.Close()

	stat, err := os.Stat(tempFile.Name())
	if err != nil {
		_ = os.Remove(tempFile.Name())
		return "", 0, err
	}
	return tempFile.Name(), stat.Size(), nil
}
