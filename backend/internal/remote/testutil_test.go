package remote

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"sync"
	"time"

	"github.com/jacklau/prism/internal/security"
)

// MakeTestTLSConfig generates a test TLS configuration with self-signed certificates
func MakeTestTLSConfig() (*tls.Config, error) {
	cert, key, err := GenerateTestCertificate()
	if err != nil {
		return nil, err
	}

	tlsCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		InsecureSkipVerify: true, // For testing only
	}, nil
}

// GenerateTestCertificate generates a self-signed certificate for testing
func GenerateTestCertificate() (certPEM, keyPEM []byte, err error) {
	// Generate ECDSA private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Prism Test"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:              []string{"localhost"},
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	// Encode certificate to PEM
	certPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Encode private key to PEM
	keyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, nil, err
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	})

	return certPEM, keyPEM, nil
}

// MockRemoteClient simulates a remote client for testing
type MockRemoteClient struct {
	SessionID      string
	ReconnectToken string
	AuthToken      string
	ClientIP       string
	ClientInfo     map[string]string

	Connected bool
	mu        sync.Mutex
}

// NewMockRemoteClient creates a new mock remote client
func NewMockRemoteClient(clientIP string) *MockRemoteClient {
	return &MockRemoteClient{
		ClientIP: clientIP,
		ClientInfo: map[string]string{
			"device":   "test-device",
			"os":       "test-os",
			"version":  "1.0.0",
		},
	}
}

// Authenticate simulates authentication against the auth service
func (c *MockRemoteClient) Authenticate(authService *security.RemoteAuthService, password string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	session, err := authService.Authenticate(password, c.ClientIP)
	if err != nil {
		return err
	}

	c.AuthToken = session.Token
	c.Connected = true
	return nil
}

// CreateSession simulates creating a session with the session manager
func (c *MockRemoteClient) CreateSession(sm *SessionManager, claims *security.Claims) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	session, err := sm.CreateSession(claims.UserID, claims, c.ClientInfo)
	if err != nil {
		return err
	}

	c.SessionID = session.ID
	c.ReconnectToken = session.ReconnectToken
	c.Connected = true
	return nil
}

// Disconnect simulates a disconnect
func (c *MockRemoteClient) Disconnect(sm *SessionManager) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.SessionID == "" {
		return ErrSessionNotFound
	}

	err := sm.MarkDisconnected(c.SessionID)
	if err != nil {
		return err
	}

	c.Connected = false
	return nil
}

// Reconnect simulates a reconnection
func (c *MockRemoteClient) Reconnect(sm *SessionManager) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ReconnectToken == "" {
		return ErrInvalidReconnectToken
	}

	session, err := sm.Reconnect(c.ReconnectToken)
	if err != nil {
		return err
	}

	c.ReconnectToken = session.ReconnectToken
	c.Connected = true
	return nil
}

// SendHeartbeat simulates sending a heartbeat
func (c *MockRemoteClient) SendHeartbeat(hh *HeartbeatHandler) (*HeartbeatAck, error) {
	c.mu.Lock()
	sessionID := c.SessionID
	c.mu.Unlock()

	msg := &HeartbeatMessage{
		SessionID: sessionID,
		Timestamp: time.Now(),
	}

	return hh.HandleHeartbeat(msg)
}

// TestSessionFixture creates a standard test session for use in tests
type TestSessionFixture struct {
	SessionManager *SessionManager
	AuthService    *security.RemoteAuthService
	Session        *RemoteSession
	Claims         *security.Claims
	Password       string
	PasswordHash   string
}

// NewTestSessionFixture creates a new test session fixture
func NewTestSessionFixture() (*TestSessionFixture, error) {
	password := "test-password-123"
	hash, err := security.HashPassword(password)
	if err != nil {
		return nil, err
	}

	authConfig := &security.RemoteAccessConfig{
		Enabled:              true,
		PasswordHash:         hash,
		SessionTimeout:       30 * time.Minute,
		MaxConcurrentSessions: 10,
		MaxFailedAttempts:    5,
		BlockDuration:        1 * time.Minute,
		MaxBlockDuration:     10 * time.Minute,
	}

	authService := security.NewRemoteAuthService(authConfig, nil)
	sessionManager := NewSessionManager(nil)

	claims := &security.Claims{
		UserID: "test-user-1",
		Email:  "test@example.com",
	}

	session, err := sessionManager.CreateSession(claims.UserID, claims, map[string]string{
		"device": "test",
	})
	if err != nil {
		authService.Stop()
		sessionManager.Stop()
		return nil, err
	}

	return &TestSessionFixture{
		SessionManager: sessionManager,
		AuthService:    authService,
		Session:        session,
		Claims:         claims,
		Password:       password,
		PasswordHash:   hash,
	}, nil
}

// Cleanup cleans up the test fixture resources
func (f *TestSessionFixture) Cleanup() {
	if f.AuthService != nil {
		f.AuthService.Stop()
	}
	if f.SessionManager != nil {
		f.SessionManager.Stop()
	}
}

// ConcurrentTestRunner helps run concurrent tests
type ConcurrentTestRunner struct {
	wg      sync.WaitGroup
	errors  []error
	mu      sync.Mutex
}

// NewConcurrentTestRunner creates a new concurrent test runner
func NewConcurrentTestRunner() *ConcurrentTestRunner {
	return &ConcurrentTestRunner{
		errors: make([]error, 0),
	}
}

// Run executes a function concurrently
func (r *ConcurrentTestRunner) Run(fn func() error) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		if err := fn(); err != nil {
			r.mu.Lock()
			r.errors = append(r.errors, err)
			r.mu.Unlock()
		}
	}()
}

// Wait waits for all goroutines to complete
func (r *ConcurrentTestRunner) Wait() {
	r.wg.Wait()
}

// Errors returns any errors that occurred
func (r *ConcurrentTestRunner) Errors() []error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.errors
}

// HasErrors returns true if any errors occurred
func (r *ConcurrentTestRunner) HasErrors() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.errors) > 0
}

// MakeTestRemoteConfig creates a test configuration for remote access
func MakeTestRemoteConfig() *SessionConfig {
	return &SessionConfig{
		HeartbeatInterval:    1 * time.Hour,
		HeartbeatTimeout:     1 * time.Hour,
		IdleTimeout:          1 * time.Hour,
		IdleWarningBefore:    1 * time.Hour,
		ReconnectTokenExpiry: 1 * time.Hour,
		CleanupInterval:      1 * time.Hour, // Very long to avoid interference
	}
}

// MakeTestHeartbeatConfig creates a test configuration for heartbeat handling
func MakeTestHeartbeatConfig() *HeartbeatConfig {
	return &HeartbeatConfig{
		Interval:    50 * time.Millisecond,
		Timeout:     150 * time.Millisecond,
		GracePeriod: 25 * time.Millisecond,
	}
}

// MakeTestReconnectConfig creates a test configuration for reconnection handling
func MakeTestReconnectConfig() *ReconnectConfig {
	return &ReconnectConfig{
		MaxAttempts:          3,
		BaseDelay:            10 * time.Millisecond,
		MaxDelay:             100 * time.Millisecond,
		TokenValidity:        500 * time.Millisecond,
		MergePendingMessages: true,
	}
}
