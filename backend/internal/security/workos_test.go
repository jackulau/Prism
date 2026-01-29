package security

import (
	"testing"
	"time"
)

func TestNewWorkOSService(t *testing.T) {
	tests := []struct {
		name         string
		clientID     string
		clientSecret string
		redirectURI  string
		encryptKey   string
		wantErr      bool
	}{
		{
			name:         "valid config",
			clientID:     "client_123",
			clientSecret: "secret_456",
			redirectURI:  "http://localhost:3000/callback",
			encryptKey:   "12345678901234567890123456789012",
			wantErr:      false,
		},
		{
			name:         "missing client ID",
			clientID:     "",
			clientSecret: "secret_456",
			redirectURI:  "http://localhost:3000/callback",
			encryptKey:   "12345678901234567890123456789012",
			wantErr:      true,
		},
		{
			name:         "missing client secret",
			clientID:     "client_123",
			clientSecret: "",
			redirectURI:  "http://localhost:3000/callback",
			encryptKey:   "12345678901234567890123456789012",
			wantErr:      true,
		},
		{
			name:         "short encrypt key - should pad",
			clientID:     "client_123",
			clientSecret: "secret_456",
			redirectURI:  "http://localhost:3000/callback",
			encryptKey:   "short",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewWorkOSService(tt.clientID, tt.clientSecret, tt.redirectURI, tt.encryptKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWorkOSService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && svc == nil {
				t.Error("NewWorkOSService() returned nil service without error")
			}
		})
	}
}

func TestWorkOSService_CreateAndDecryptSession(t *testing.T) {
	svc, err := NewWorkOSService("client_123", "secret_456", "http://localhost:3000/callback", "12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("Failed to create WorkOS service: %v", err)
	}

	session := &WorkOSSession{
		ID:             "session_abc123",
		UserID:         "user_xyz789",
		OrganizationID: "org_def456",
		ConnectionID:   "conn_ghi012",
		ExpiresAt:      time.Now().Add(24 * time.Hour),
		CreatedAt:      time.Now(),
	}

	// Create session cookie
	cookieValue, err := svc.CreateSession(session)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if cookieValue == "" {
		t.Error("CreateSession() returned empty cookie value")
	}

	// Decrypt session cookie
	decrypted, err := svc.DecryptSession(cookieValue)
	if err != nil {
		t.Fatalf("DecryptSession() error = %v", err)
	}

	// Verify session fields
	if decrypted.ID != session.ID {
		t.Errorf("Session ID mismatch: got %v, want %v", decrypted.ID, session.ID)
	}
	if decrypted.UserID != session.UserID {
		t.Errorf("Session UserID mismatch: got %v, want %v", decrypted.UserID, session.UserID)
	}
	if decrypted.OrganizationID != session.OrganizationID {
		t.Errorf("Session OrganizationID mismatch: got %v, want %v", decrypted.OrganizationID, session.OrganizationID)
	}
	if decrypted.ConnectionID != session.ConnectionID {
		t.Errorf("Session ConnectionID mismatch: got %v, want %v", decrypted.ConnectionID, session.ConnectionID)
	}
}

func TestWorkOSService_DecryptSession_InvalidCookie(t *testing.T) {
	svc, err := NewWorkOSService("client_123", "secret_456", "http://localhost:3000/callback", "12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("Failed to create WorkOS service: %v", err)
	}

	tests := []struct {
		name        string
		cookieValue string
	}{
		{
			name:        "empty cookie",
			cookieValue: "",
		},
		{
			name:        "invalid base64",
			cookieValue: "not-valid-base64!!!",
		},
		{
			name:        "too short ciphertext",
			cookieValue: "YWJj", // "abc" in base64
		},
		{
			name:        "tampered ciphertext",
			cookieValue: "SGVsbG8gV29ybGQhIFRoaXMgaXMgYSB0ZXN0IG1lc3NhZ2U=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.DecryptSession(tt.cookieValue)
			if err == nil {
				t.Error("DecryptSession() expected error for invalid cookie, got nil")
			}
		})
	}
}

func TestWorkOSService_DifferentKeys(t *testing.T) {
	svc1, _ := NewWorkOSService("client_123", "secret_456", "http://localhost:3000/callback", "12345678901234567890123456789012")
	svc2, _ := NewWorkOSService("client_123", "secret_456", "http://localhost:3000/callback", "different-key-for-encryption!!")

	session := &WorkOSSession{
		ID:             "session_abc123",
		UserID:         "user_xyz789",
		OrganizationID: "org_def456",
		ConnectionID:   "conn_ghi012",
		ExpiresAt:      time.Now().Add(24 * time.Hour),
		CreatedAt:      time.Now(),
	}

	// Create session with first key
	cookieValue, err := svc1.CreateSession(session)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	// Try to decrypt with different key - should fail
	_, err = svc2.DecryptSession(cookieValue)
	if err == nil {
		t.Error("DecryptSession() should fail with different key")
	}
}

func TestWorkOSService_Getters(t *testing.T) {
	clientID := "client_123"
	redirectURI := "http://localhost:3000/callback"

	svc, err := NewWorkOSService(clientID, "secret_456", redirectURI, "12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("Failed to create WorkOS service: %v", err)
	}

	if svc.GetClientID() != clientID {
		t.Errorf("GetClientID() = %v, want %v", svc.GetClientID(), clientID)
	}

	if svc.GetRedirectURI() != redirectURI {
		t.Errorf("GetRedirectURI() = %v, want %v", svc.GetRedirectURI(), redirectURI)
	}
}
