package remote

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// TunnelServerConfig holds configuration for the tunnel server
type TunnelServerConfig struct {
	// ListenAddr is the address to listen on (e.g., ":8443")
	ListenAddr string

	// TLSCertFile is the path to the TLS certificate file
	TLSCertFile string

	// TLSKeyFile is the path to the TLS key file
	TLSKeyFile string

	// TargetURL is the URL of the main Prism server to proxy to
	TargetURL string

	// MaxConnections is the maximum number of concurrent connections
	MaxConnections int

	// MaxConnectionsPerIP is the maximum connections per IP address
	MaxConnectionsPerIP int

	// SessionTimeout is the timeout for idle sessions
	SessionTimeout time.Duration

	// ShutdownTimeout is the timeout for graceful shutdown
	ShutdownTimeout time.Duration

	// RequireClientCert enables client certificate verification
	RequireClientCert bool

	// ClientCAFile is the path to the CA file for client certificate verification
	ClientCAFile string
}

// DefaultTunnelServerConfig returns a configuration with sensible defaults
func DefaultTunnelServerConfig() *TunnelServerConfig {
	return &TunnelServerConfig{
		ListenAddr:          ":8443",
		MaxConnections:      100,
		MaxConnectionsPerIP: 10,
		SessionTimeout:      30 * time.Minute,
		ShutdownTimeout:     30 * time.Second,
		RequireClientCert:   false,
	}
}

// TunnelServer is a TLS-secured proxy server for remote connections
type TunnelServer struct {
	config        *TunnelServerConfig
	authService   *RemoteAuthService
	httpProxy     *httputil.ReverseProxy
	wsProxy       *WebSocketProxy
	connManager   *ConnectionManager
	app           *fiber.App
	listener      net.Listener
	running       atomic.Bool
	mu            sync.RWMutex
	shutdownCh    chan struct{}
}

// NewTunnelServer creates a new tunnel server
func NewTunnelServer(cfg *TunnelServerConfig, authService *RemoteAuthService) (*TunnelServer, error) {
	if cfg == nil {
		cfg = DefaultTunnelServerConfig()
	}

	// Parse target URL
	targetURL, err := url.Parse(cfg.TargetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid target URL: %w", err)
	}

	// Create HTTP reverse proxy
	httpProxy := httputil.NewSingleHostReverseProxy(targetURL)
	httpProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error: %v", err)
		http.Error(w, "Proxy error", http.StatusBadGateway)
	}

	// Create WebSocket proxy
	wsProxy := NewWebSocketProxy(targetURL)

	// Create connection manager
	connManager := NewConnectionManager(cfg.MaxConnections, cfg.MaxConnectionsPerIP)

	server := &TunnelServer{
		config:      cfg,
		authService: authService,
		httpProxy:   httpProxy,
		wsProxy:     wsProxy,
		connManager: connManager,
		shutdownCh:  make(chan struct{}),
	}

	// Setup Fiber app with routes
	server.setupRoutes()

	return server, nil
}

// setupRoutes configures the HTTP routes for the tunnel server
func (t *TunnelServer) setupRoutes() {
	t.app = fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		IdleTimeout:           t.config.SessionTimeout,
	})

	// Health check endpoint (no auth required)
	t.app.Get("/health", t.handleHealth)

	// Authentication endpoints
	t.app.Post("/remote/auth", t.handleAuth)
	t.app.Post("/remote/logout", t.authMiddleware, t.handleLogout)
	t.app.Get("/remote/status", t.authMiddleware, t.handleStatus)

	// Proxy all other requests (requires authentication)
	t.app.Use("/api/*", t.authMiddleware, t.connectionLimitMiddleware, t.proxyMiddleware)

	// WebSocket upgrade handler
	t.app.Get("/ws", t.authMiddleware, t.connectionLimitMiddleware, websocket.New(t.handleWebSocket))
}

// Start starts the tunnel server
func (t *TunnelServer) Start() error {
	if t.running.Load() {
		return fmt.Errorf("server is already running")
	}

	// Create TLS configuration
	tlsConfig, err := t.createTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to create TLS config: %w", err)
	}

	// Create TLS listener
	listener, err := tls.Listen("tcp", t.config.ListenAddr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to create TLS listener: %w", err)
	}
	t.listener = listener
	t.running.Store(true)

	log.Printf("Tunnel server starting on %s (TLS)", t.config.ListenAddr)

	// Start accepting connections
	go func() {
		if err := t.app.Listener(listener); err != nil {
			if t.running.Load() {
				log.Printf("Tunnel server error: %v", err)
			}
		}
	}()

	return nil
}

// Stop gracefully stops the tunnel server
func (t *TunnelServer) Stop() error {
	if !t.running.Load() {
		return nil
	}

	t.running.Store(false)
	close(t.shutdownCh)

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), t.config.ShutdownTimeout)
	defer cancel()

	// Close all active connections
	t.connManager.CloseAll()

	// Shutdown the app
	if err := t.app.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	log.Println("Tunnel server stopped")
	return nil
}

// createTLSConfig creates a secure TLS configuration
func (t *TunnelServer) createTLSConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(t.config.TLSCertFile, t.config.TLSKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
		},
	}

	// Optional client certificate verification
	if t.config.RequireClientCert && t.config.ClientCAFile != "" {
		// Load client CA if provided
		// This would be implemented to load the CA file
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return tlsConfig, nil
}

// GetActiveConnections returns the list of active connections
func (t *TunnelServer) GetActiveConnections() []*TunnelConnection {
	return t.connManager.GetAll()
}

// GetStats returns server statistics
func (t *TunnelServer) GetStats() *TunnelServerStats {
	conns := t.connManager.GetAll()
	var totalBytesIn, totalBytesOut int64
	for _, conn := range conns {
		totalBytesIn += conn.BytesIn.Load()
		totalBytesOut += conn.BytesOut.Load()
	}

	return &TunnelServerStats{
		ActiveConnections: len(conns),
		TotalBytesIn:      totalBytesIn,
		TotalBytesOut:     totalBytesOut,
		Uptime:            time.Since(t.connManager.startTime),
	}
}

// TunnelServerStats holds server statistics
type TunnelServerStats struct {
	ActiveConnections int           `json:"active_connections"`
	TotalBytesIn      int64         `json:"total_bytes_in"`
	TotalBytesOut     int64         `json:"total_bytes_out"`
	Uptime            time.Duration `json:"uptime"`
}
