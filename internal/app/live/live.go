// Package live serves a selected project directory over a small, read-only
// local-network HTTP session.
package live

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"mime"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	pathpkg "path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const discoveryPort = 39422

var codeAlphabet = []byte("23456789ABCDEFGHJKLMNPQRSTUVWXYZ")

// Session contains the user-facing details of an active live directory.
type Session struct {
	Code      string
	Root      string
	CreatedAt time.Time
	ExpiresAt time.Time
	URL       string
	LocalURL  string
}

// Server owns an active read-only project directory session.
type Server struct {
	session  Session
	token    string
	root     string
	realRoot string
	http     *http.Server
	listener net.Listener
	udp      *net.UDPConn
	proxy    *httputil.ReverseProxy
	stopOnce sync.Once
}

// StartDirectory exposes root as a read-only HTTP directory until stopped.
func StartDirectory(root string, lifetime time.Duration) (*Server, error) {
	if lifetime <= 0 {
		return nil, fmt.Errorf("live session lifetime must be positive")
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve project directory: %w", err)
	}
	info, err := os.Stat(absoluteRoot)
	if err != nil {
		return nil, fmt.Errorf("inspect project directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", root)
	}
	realRoot, err := filepath.EvalSymlinks(absoluteRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve project directory links: %w", err)
	}
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return nil, fmt.Errorf("listen for live session: %w", err)
	}
	code, err := randomCode(6)
	if err != nil {
		listener.Close()
		return nil, err
	}
	token, err := randomToken()
	if err != nil {
		listener.Close()
		return nil, err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	createdAt := time.Now()
	server := &Server{
		root:     absoluteRoot,
		realRoot: realRoot,
		token:    token,
		listener: listener,
		session: Session{
			Code:      code,
			Root:      absoluteRoot,
			CreatedAt: createdAt,
			ExpiresAt: createdAt.Add(lifetime),
			URL:       sessionURL(localIPv4(), port, code, token),
			LocalURL:  sessionURL(net.IPv4(127, 0, 0, 1), port, code, token),
		},
	}
	server.http = &http.Server{Handler: http.HandlerFunc(server.serve)}
	go func() { _ = server.http.Serve(listener) }()
	if err := server.startDiscovery(); err != nil {
		server.Stop(context.Background())
		return nil, err
	}
	return server, nil
}

// Session returns the immutable session metadata.
func (s *Server) Session() Session { return s.session }

// Stop cleanly closes the HTTP listener.
func (s *Server) Stop(ctx context.Context) error {
	var stopErr error
	s.stopOnce.Do(func() {
		if s.udp != nil {
			_ = s.udp.Close()
		}
		stopErr = s.http.Shutdown(ctx)
	})
	return stopErr
}

// StartProxy exposes an already-running HTTP service on localPort through a
// token-protected, expiring live session.
func StartProxy(localPort int, lifetime time.Duration) (*Server, error) {
	if localPort < 1 || localPort > 65535 {
		return nil, fmt.Errorf("local service port must be between 1 and 65535")
	}
	if lifetime <= 0 {
		return nil, fmt.Errorf("live session lifetime must be positive")
	}
	targetAddress := net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", localPort))
	connection, err := net.DialTimeout("tcp", targetAddress, time.Second)
	if err != nil {
		return nil, fmt.Errorf("connect to local service on port %d: %w", localPort, err)
	}
	_ = connection.Close()

	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return nil, fmt.Errorf("listen for live session: %w", err)
	}
	code, err := randomCode(6)
	if err != nil {
		listener.Close()
		return nil, err
	}
	token, err := randomToken()
	if err != nil {
		listener.Close()
		return nil, err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	createdAt := time.Now()
	server := &Server{
		token:    token,
		listener: listener,
		session: Session{
			Code:      code,
			Root:      "http://localhost:" + fmt.Sprintf("%d", localPort),
			CreatedAt: createdAt,
			ExpiresAt: createdAt.Add(lifetime),
			URL:       sessionURL(localIPv4(), port, code, token),
			LocalURL:  sessionURL(net.IPv4(127, 0, 0, 1), port, code, token),
		},
	}
	target := &url.URL{Scheme: "http", Host: targetAddress}
	server.proxy = httputil.NewSingleHostReverseProxy(target)
	server.proxy.Director = func(request *http.Request) {
		prefix := "/c/" + server.session.Code
		proxiedPath := strings.TrimPrefix(request.URL.Path, prefix)
		if proxiedPath == "" {
			proxiedPath = "/"
		}
		query := request.URL.Query()
		query.Del("token")
		request.URL.Scheme = target.Scheme
		request.URL.Host = target.Host
		request.URL.Path = proxiedPath
		request.URL.RawPath = ""
		request.URL.RawQuery = query.Encode()
		request.Host = target.Host
	}
	server.proxy.ErrorHandler = func(writer http.ResponseWriter, _ *http.Request, _ error) {
		http.Error(writer, "Could not reach the local service behind this live session.", http.StatusBadGateway)
	}
	server.http = &http.Server{Handler: http.HandlerFunc(server.serveProxy)}
	go func() { _ = server.http.Serve(listener) }()
	if err := server.startDiscovery(); err != nil {
		server.Stop(context.Background())
		return nil, err
	}
	return server, nil
}

func (s *Server) serveProxy(writer http.ResponseWriter, request *http.Request) {
	prefix := "/c/" + s.session.Code
	if request.URL.Path != prefix && !strings.HasPrefix(request.URL.Path, prefix+"/") {
		http.NotFound(writer, request)
		return
	}
	if time.Now().After(s.session.ExpiresAt) {
		http.Error(writer, "This live session has expired.", http.StatusGone)
		return
	}
	if subtle.ConstantTimeCompare([]byte(request.URL.Query().Get("token")), []byte(s.token)) != 1 {
		http.Error(writer, "A valid live-session token is required.", http.StatusForbidden)
		return
	}
	s.proxy.ServeHTTP(writer, request)
}

func (s *Server) serve(writer http.ResponseWriter, request *http.Request) {
	prefix := "/c/" + s.session.Code
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		writer.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
		http.Error(writer, "Method not allowed.", http.StatusMethodNotAllowed)
		return
	}
	if request.URL.Path != prefix && !strings.HasPrefix(request.URL.Path, prefix+"/") {
		http.NotFound(writer, request)
		return
	}
	if time.Now().After(s.session.ExpiresAt) {
		http.Error(writer, "This live session has expired.", http.StatusGone)
		return
	}
	if subtle.ConstantTimeCompare([]byte(request.URL.Query().Get("token")), []byte(s.token)) != 1 {
		http.Error(writer, "A valid live-session token is required.", http.StatusForbidden)
		return
	}
	relativePath := strings.TrimPrefix(request.URL.Path, prefix)
	if relativePath == "" {
		relativePath = "/"
	}
	filePath, err := s.safePath(relativePath)
	if err != nil {
		http.Error(writer, "The requested path is outside the live project.", http.StatusForbidden)
		return
	}
	file, err := os.Open(filePath)
	if err != nil {
		http.NotFound(writer, request)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		http.NotFound(writer, request)
		return
	}
	if info.IsDir() {
		if !strings.HasSuffix(request.URL.Path, "/") {
			redirect := request.URL.Path + "/?" + request.URL.RawQuery
			http.Redirect(writer, request, redirect, http.StatusMovedPermanently)
			return
		}
		s.serveDirectory(writer, request, file, relativePath)
		return
	}
	if contentType := mime.TypeByExtension(filepath.Ext(info.Name())); contentType != "" {
		writer.Header().Set("Content-Type", contentType)
	}
	http.ServeContent(writer, request, info.Name(), info.ModTime(), file)
}

func (s *Server) safePath(requestPath string) (string, error) {
	cleaned := pathpkg.Clean("/" + requestPath)
	relative := strings.TrimPrefix(cleaned, "/")
	candidate := filepath.Join(s.root, filepath.FromSlash(relative))
	relativeToRoot, err := filepath.Rel(s.root, candidate)
	if err != nil || relativeToRoot == ".." || strings.HasPrefix(relativeToRoot, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes root")
	}
	realCandidate, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", err
	}
	relativeToRealRoot, err := filepath.Rel(s.realRoot, realCandidate)
	if err != nil || relativeToRealRoot == ".." || strings.HasPrefix(relativeToRealRoot, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("symlink escapes root")
	}
	return realCandidate, nil
}

func (s *Server) serveDirectory(writer http.ResponseWriter, request *http.Request, directory *os.File, relativePath string) {
	entries, err := directory.ReadDir(-1)
	if err != nil {
		http.Error(writer, "Could not list this directory.", http.StatusInternalServerError)
		return
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	if request.Method == http.MethodHead {
		return
	}
	base := "/c/" + s.session.Code + "/" + strings.TrimPrefix(relativePath, "/")
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}
	query := "?token=" + url.QueryEscape(s.token)
	var page strings.Builder
	page.WriteString("<!doctype html><html><head><meta charset=\"utf-8\"><title>cast live</title></head><body><h1>Live project</h1><ul>")
	if strings.Trim(relativePath, "/") != "" {
		parent := pathpkg.Dir(strings.TrimSuffix(base, "/"))
		page.WriteString("<li><a href=\"")
		page.WriteString(html.EscapeString(parent + "/" + query))
		page.WriteString("\">..</a></li>")
	}
	for _, entry := range entries {
		name := entry.Name()
		path := base + url.PathEscape(name)
		label := name
		if entry.IsDir() {
			path += "/"
			label += "/"
		}
		page.WriteString("<li><a href=\"")
		page.WriteString(html.EscapeString(path + query))
		page.WriteString("\">")
		page.WriteString(html.EscapeString(label))
		page.WriteString("</a></li>")
	}
	page.WriteString("</ul></body></html>")
	_, _ = io.WriteString(writer, page.String())
}

func randomCode(length int) (string, error) {
	bytes := make([]byte, length)
	for index := range bytes {
		for {
			var byteValue [1]byte
			if _, err := rand.Read(byteValue[:]); err != nil {
				return "", fmt.Errorf("generate live code: %w", err)
			}
			limit := 256 - (256 % len(codeAlphabet))
			if int(byteValue[0]) < limit {
				bytes[index] = codeAlphabet[int(byteValue[0])%len(codeAlphabet)]
				break
			}
		}
	}
	return string(bytes), nil
}

func randomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate live token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func sessionURL(address net.IP, port int, code, token string) string {
	value := url.URL{Scheme: "http", Host: net.JoinHostPort(address.String(), fmt.Sprintf("%d", port)), Path: "/c/" + code}
	query := value.Query()
	query.Set("token", token)
	value.RawQuery = query.Encode()
	return value.String()
}

func localIPv4() net.IP {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, networkInterface := range interfaces {
			if networkInterface.Flags&net.FlagUp == 0 || networkInterface.Flags&net.FlagLoopback != 0 {
				continue
			}
			addresses, err := networkInterface.Addrs()
			if err != nil {
				continue
			}
			for _, address := range addresses {
				if ip, ok := address.(*net.IPNet); ok && ip.IP.To4() != nil && !ip.IP.IsLoopback() {
					return ip.IP.To4()
				}
			}
		}
	}
	return net.IPv4(127, 0, 0, 1)
}

func (s *Server) startDiscovery() error {
	connection, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: discoveryPort})
	if err != nil {
		return fmt.Errorf("listen for local live discovery: %w", err)
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
			if json.Unmarshal(buffer[:count], &request) != nil || request.Type != "cast-live-discover" || request.Code != s.session.Code {
				continue
			}
			project := s.session.Root
			if s.root != "" {
				project = filepath.Base(s.root)
			}
			response, _ := json.Marshal(discoveryResponse{
				Type:      "cast-live",
				Code:      s.session.Code,
				URL:       s.session.URL,
				Project:   project,
				Device:    deviceName(),
				ExpiresAt: s.session.ExpiresAt,
			})
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
	Project   string    `json:"project"`
	Device    string    `json:"device"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ResolvedSession is the result of discovering an active live session by short code.
type ResolvedSession struct {
	Code      string
	URL       string
	Project   string
	Device    string
	ExpiresAt time.Time
}

// Resolve broadcasts a short-code lookup for an active live session over the local network
// and loopback interface.
func Resolve(ctx context.Context, code string) (ResolvedSession, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if !validCode(code) {
		return ResolvedSession{}, fmt.Errorf("invalid live session code")
	}
	connection, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return ResolvedSession{}, fmt.Errorf("start local live discovery: %w", err)
	}
	defer connection.Close()
	if err := enableBroadcast(connection); err != nil {
		return ResolvedSession{}, fmt.Errorf("enable local live discovery: %w", err)
	}
	payload, _ := json.Marshal(discoveryRequest{Type: "cast-live-discover", Code: code})
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
		return ResolvedSession{}, fmt.Errorf("set live discovery deadline: %w", err)
	}
	buffer := make([]byte, 2048)
	for {
		count, _, err := connection.ReadFromUDP(buffer)
		if err != nil {
			if ctx.Err() != nil {
				return ResolvedSession{}, ctx.Err()
			}
			return ResolvedSession{}, fmt.Errorf("live session code was not found on the local network")
		}
		var response discoveryResponse
		if json.Unmarshal(buffer[:count], &response) != nil || response.Type != "cast-live" || response.Code != code || response.URL == "" {
			continue
		}
		if time.Now().After(response.ExpiresAt) {
			return ResolvedSession{}, fmt.Errorf("live session has expired")
		}
		return ResolvedSession{
			Code:      response.Code,
			URL:       response.URL,
			Project:   response.Project,
			Device:    response.Device,
			ExpiresAt: response.ExpiresAt,
		}, nil
	}
}

func validCode(code string) bool {
	if len(code) < 4 || len(code) > 8 {
		return false
	}
	for _, character := range code {
		if !strings.ContainsRune(string(codeAlphabet), character) {
			return false
		}
	}
	return true
}

func deviceName() string {
	if host, err := os.Hostname(); err == nil && host != "" {
		return host
	}
	return "Local Device"
}
