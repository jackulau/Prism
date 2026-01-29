package remote

import (
	"testing"
	"time"

	"github.com/jacklau/prism/internal/security"
)

func TestRemoteAuthService_Authenticate(t *testing.T) {
	// Create password hash for testing
	testPassword := "test-password-123"
	passwordHash, err := security.HashPassword(testPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	cfg := &RemoteAuthConfig{
		Enabled:         true,
		PasswordHash:    passwordHash,
		SessionDuration: time.Hour,
		MaxSessions:     10,
	}

	authService := NewRemoteAuthService(cfg)

	tests := []struct {
		name        string
		password    string
		wantErr     error
		wantSession bool
	}{
		{
			name:        "valid credentials",
			password:    testPassword,
			wantErr:     nil,
			wantSession: true,
		},
		{
			name:        "invalid password",
			password:    "wrong-password",
			wantErr:     ErrInvalidCredentials,
			wantSession: false,
		},
		{
			name:        "empty password",
			password:    "",
			wantErr:     ErrInvalidCredentials,
			wantSession: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, token, err := authService.Authenticate(tt.password, "192.168.1.1", "TestClient/1.0")

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Authenticate() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Authenticate() unexpected error: %v", err)
				return
			}

			if tt.wantSession {
				if session == nil {
					t.Error("Authenticate() session is nil, want non-nil")
					return
				}
				if token == "" {
					t.Error("Authenticate() token is empty")
				}
				if session.ID == "" {
					t.Error("Authenticate() session.ID is empty")
				}
				if session.ClientIP != "192.168.1.1" {
					t.Errorf("Authenticate() session.ClientIP = %v, want %v", session.ClientIP, "192.168.1.1")
				}
			}
		})
	}
}

func TestRemoteAuthService_ValidateToken(t *testing.T) {
	testPassword := "test-password-123"
	passwordHash, err := security.HashPassword(testPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	cfg := &RemoteAuthConfig{
		Enabled:         true,
		PasswordHash:    passwordHash,
		SessionDuration: time.Hour,
		MaxSessions:     10,
	}

	authService := NewRemoteAuthService(cfg)

	// Create a session first
	session, token, err := authService.Authenticate(testPassword, "192.168.1.1", "TestClient/1.0")
	if err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{
			name:    "valid token",
			token:   token,
			wantErr: nil,
		},
		{
			name:    "invalid token",
			token:   "invalid-token",
			wantErr: ErrSessionNotFound,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validated, err := authService.ValidateToken(tt.token)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateToken() unexpected error: %v", err)
				return
			}

			if validated.ID != session.ID {
				t.Errorf("ValidateToken() session.ID = %v, want %v", validated.ID, session.ID)
			}
		})
	}
}

func TestRemoteAuthService_InvalidateSession(t *testing.T) {
	testPassword := "test-password-123"
	passwordHash, err := security.HashPassword(testPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	cfg := &RemoteAuthConfig{
		Enabled:         true,
		PasswordHash:    passwordHash,
		SessionDuration: time.Hour,
		MaxSessions:     10,
	}

	authService := NewRemoteAuthService(cfg)

	// Create a session
	session, token, err := authService.Authenticate(testPassword, "192.168.1.1", "TestClient/1.0")
	if err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}

	// Invalidate the session
	err = authService.InvalidateSession(session.ID)
	if err != nil {
		t.Errorf("InvalidateSession() error = %v", err)
	}

	// Try to validate the token - should fail
	_, err = authService.ValidateToken(token)
	if err != ErrSessionNotFound {
		t.Errorf("ValidateToken() after invalidation error = %v, want %v", err, ErrSessionNotFound)
	}
}

func TestRemoteAuthService_SessionLimit(t *testing.T) {
	testPassword := "test-password-123"
	passwordHash, err := security.HashPassword(testPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	cfg := &RemoteAuthConfig{
		Enabled:         true,
		PasswordHash:    passwordHash,
		SessionDuration: time.Hour,
		MaxSessions:     2, // Limit to 2 sessions
	}

	authService := NewRemoteAuthService(cfg)

	// Create first session
	_, _, err = authService.Authenticate(testPassword, "192.168.1.1", "TestClient/1.0")
	if err != nil {
		t.Fatalf("Failed to create first session: %v", err)
	}

	// Create second session
	_, _, err = authService.Authenticate(testPassword, "192.168.1.2", "TestClient/1.0")
	if err != nil {
		t.Fatalf("Failed to create second session: %v", err)
	}

	// Try to create third session - should fail
	_, _, err = authService.Authenticate(testPassword, "192.168.1.3", "TestClient/1.0")
	if err != ErrConnectionLimitExceeded {
		t.Errorf("Authenticate() with full sessions error = %v, want %v", err, ErrConnectionLimitExceeded)
	}
}

func TestRemoteAuthService_Disabled(t *testing.T) {
	cfg := &RemoteAuthConfig{
		Enabled: false, // Disabled
	}

	authService := NewRemoteAuthService(cfg)

	_, _, err := authService.Authenticate("any-password", "192.168.1.1", "TestClient/1.0")
	if err != ErrRemoteAccessDisabled {
		t.Errorf("Authenticate() with disabled service error = %v, want %v", err, ErrRemoteAccessDisabled)
	}
}

func TestRemoteSession_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "not expired",
			expiresAt: time.Now().Add(time.Hour),
			want:      false,
		},
		{
			name:      "expired",
			expiresAt: time.Now().Add(-time.Hour),
			want:      true,
		},
		{
			name:      "just expired",
			expiresAt: time.Now().Add(-time.Millisecond),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &RemoteSession{
				ExpiresAt: tt.expiresAt,
			}
			if got := session.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}
