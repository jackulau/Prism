package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// WorkOSSession represents a WorkOS SSO session
type WorkOSSession struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	OrganizationID string    `json:"organization_id"`
	ConnectionID   string    `json:"connection_id"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// WorkOSService handles WorkOS SSO operations
type WorkOSService struct {
	clientID     string
	clientSecret string
	redirectURI  string
	encryptKey   []byte
}

// NewWorkOSService creates a new WorkOS service
func NewWorkOSService(clientID, clientSecret, redirectURI, encryptKey string) (*WorkOSService, error) {
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("WorkOS client ID and secret are required")
	}

	// Decode encryption key (must be 32 bytes for AES-256)
	keyBytes, err := base64.StdEncoding.DecodeString(encryptKey)
	if err != nil {
		// If not base64, try using the string directly (padded/truncated to 32 bytes)
		keyBytes = []byte(encryptKey)
	}

	// Ensure key is exactly 32 bytes
	if len(keyBytes) < 32 {
		// Pad with zeros if too short
		padded := make([]byte, 32)
		copy(padded, keyBytes)
		keyBytes = padded
	} else if len(keyBytes) > 32 {
		// Truncate if too long
		keyBytes = keyBytes[:32]
	}

	return &WorkOSService{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		encryptKey:   keyBytes,
	}, nil
}

// CreateSession creates an encrypted session cookie value
func (s *WorkOSService) CreateSession(session *WorkOSSession) (string, error) {
	// Serialize session to JSON
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return "", fmt.Errorf("failed to marshal session: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(s.encryptKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the session data
	ciphertext := gcm.Seal(nonce, nonce, sessionJSON, nil)

	// Encode as base64 for cookie storage
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// DecryptSession decrypts and parses a session cookie
func (s *WorkOSService) DecryptSession(cookieValue string) (*WorkOSSession, error) {
	// Decode base64
	ciphertext, err := base64.URLEncoding.DecodeString(cookieValue)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cookie: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(s.encryptKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Check minimum length
	if len(ciphertext) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt session: %w", err)
	}

	// Parse JSON
	var session WorkOSSession
	if err := json.Unmarshal(plaintext, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// GetClientID returns the WorkOS client ID
func (s *WorkOSService) GetClientID() string {
	return s.clientID
}

// GetRedirectURI returns the configured redirect URI
func (s *WorkOSService) GetRedirectURI() string {
	return s.redirectURI
}
