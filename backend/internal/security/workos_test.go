package security

import (
	"testing"
	"time"
)

func TestNewWorkOSService(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		clientID       string
		redirectURI    string
		cookiePassword string
		wantConfigured bool
	}{
		{
			name:           "valid config",
			apiKey:         "sk_test_123",
			clientID:       "client_123",
			redirectURI:    "http://localhost:3000/callback",
			cookiePassword: "12345678901234567890123456789012",
			wantConfigured: true,
		},
		{
			name:           "missing API key",
			apiKey:         "",
			clientID:       "client_123",
			redirectURI:    "http://localhost:3000/callback",
			cookiePassword: "12345678901234567890123456789012",
			wantConfigured: false,
		},
		{
			name:           "missing client ID",
			apiKey:         "sk_test_123",
			clientID:       "",
			redirectURI:    "http://localhost:3000/callback",
			cookiePassword: "12345678901234567890123456789012",
			wantConfigured: false,
		},
		{
			name:           "missing cookie password",
			apiKey:         "sk_test_123",
			clientID:       "client_123",
			redirectURI:    "http://localhost:3000/callback",
			cookiePassword: "",
			wantConfigured: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewWorkOSService(tt.apiKey, tt.clientID, tt.redirectURI, tt.cookiePassword)
			if svc == nil {
				t.Error("NewWorkOSService() returned nil")
				return
			}
			if svc.IsConfigured() != tt.wantConfigured {
				t.Errorf("IsConfigured() = %v, want %v", svc.IsConfigured(), tt.wantConfigured)
			}
		})
	}
}

func TestWorkOSService_CreateAndDecryptSessionCookie(t *testing.T) {
	svc := NewWorkOSService("sk_test_123", "client_123", "http://localhost:3000/callback", "12345678901234567890123456789012")

	session := &SessionData{
		ID:             "session_abc123",
		UserID:         "user_xyz789",
		OrganizationID: "org_def456",
		ConnectionID:   "conn_ghi012",
		Email:          "test@example.com",
		ExpiresAt:      time.Now().Add(24 * time.Hour),
	}

	// Encrypt session cookie
	cookieValue, err := svc.EncryptSessionCookie(session)
	if err != nil {
		t.Fatalf("EncryptSessionCookie() error = %v", err)
	}
	if cookieValue == "" {
		t.Error("EncryptSessionCookie() returned empty cookie value")
	}

	// Decrypt session cookie
	decrypted, err := svc.DecryptSessionCookie(cookieValue)
	if err != nil {
		t.Fatalf("DecryptSessionCookie() error = %v", err)
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
	if decrypted.Email != session.Email {
		t.Errorf("Session Email mismatch: got %v, want %v", decrypted.Email, session.Email)
	}
}

func TestWorkOSService_DecryptSessionCookie_InvalidCookie(t *testing.T) {
	svc := NewWorkOSService("sk_test_123", "client_123", "http://localhost:3000/callback", "12345678901234567890123456789012")

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
			_, err := svc.DecryptSessionCookie(tt.cookieValue)
			if err == nil {
				t.Error("DecryptSessionCookie() expected error for invalid cookie, got nil")
			}
		})
	}
}

func TestWorkOSService_DifferentKeys(t *testing.T) {
	svc1 := NewWorkOSService("sk_test_123", "client_123", "http://localhost:3000/callback", "12345678901234567890123456789012")
	svc2 := NewWorkOSService("sk_test_123", "client_123", "http://localhost:3000/callback", "different-key-for-encryption!!")

	session := &SessionData{
		ID:             "session_abc123",
		UserID:         "user_xyz789",
		OrganizationID: "org_def456",
		ConnectionID:   "conn_ghi012",
		Email:          "test@example.com",
		ExpiresAt:      time.Now().Add(24 * time.Hour),
	}

	// Encrypt session with first key
	cookieValue, err := svc1.EncryptSessionCookie(session)
	if err != nil {
		t.Fatalf("EncryptSessionCookie() error = %v", err)
	}

	// Try to decrypt with different key - should fail
	_, err = svc2.DecryptSessionCookie(cookieValue)
	if err == nil {
		t.Error("DecryptSessionCookie() should fail with different key")
	}
}

func TestWorkOSService_Getters(t *testing.T) {
	clientID := "client_123"
	redirectURI := "http://localhost:3000/callback"
	cookiePassword := "12345678901234567890123456789012"

	svc := NewWorkOSService("sk_test_123", clientID, redirectURI, cookiePassword)

	if svc.GetClientID() != clientID {
		t.Errorf("GetClientID() = %v, want %v", svc.GetClientID(), clientID)
	}

	if svc.GetRedirectURI() != redirectURI {
		t.Errorf("GetRedirectURI() = %v, want %v", svc.GetRedirectURI(), redirectURI)
	}

	if svc.GetCookiePassword() != cookiePassword {
		t.Errorf("GetCookiePassword() = %v, want %v", svc.GetCookiePassword(), cookiePassword)
	}
}

func TestWorkOSService_GenerateAndValidateState(t *testing.T) {
	svc := NewWorkOSService("sk_test_123", "client_123", "http://localhost:3000/callback", "12345678901234567890123456789012")

	// Generate state with associated data
	data := "user_123"
	state, err := svc.GenerateState(data)
	if err != nil {
		t.Fatalf("GenerateState() error = %v", err)
	}
	if state == "" {
		t.Error("GenerateState() returned empty state")
	}

	// Validate state
	retrievedData, ok := svc.ValidateState(state)
	if !ok {
		t.Error("ValidateState() returned false for valid state")
	}
	if retrievedData != data {
		t.Errorf("ValidateState() data = %v, want %v", retrievedData, data)
	}

	// State should be consumed (one-time use)
	_, ok = svc.ValidateState(state)
	if ok {
		t.Error("ValidateState() should return false for already-consumed state")
	}
}

func TestWorkOSService_ValidateState_Invalid(t *testing.T) {
	svc := NewWorkOSService("sk_test_123", "client_123", "http://localhost:3000/callback", "12345678901234567890123456789012")

	// Validate non-existent state
	_, ok := svc.ValidateState("nonexistent")
	if ok {
		t.Error("ValidateState() should return false for non-existent state")
	}
}

func TestWorkOSService_HealthCheck(t *testing.T) {
	// Configured service
	svc := NewWorkOSService("sk_test_123", "client_123", "http://localhost:3000/callback", "12345678901234567890123456789012")
	if err := svc.HealthCheck(); err != nil {
		t.Errorf("HealthCheck() error = %v for configured service", err)
	}

	// Unconfigured service
	svc2 := NewWorkOSService("", "", "", "")
	if err := svc2.HealthCheck(); err == nil {
		t.Error("HealthCheck() should return error for unconfigured service")
	}
}

func TestWorkOSService_CreateSessionData(t *testing.T) {
	svc := NewWorkOSService("sk_test_123", "client_123", "http://localhost:3000/callback", "12345678901234567890123456789012")

	profile := &SSOProfile{
		ID:             "profile_123",
		Email:          "test@example.com",
		FirstName:      "Test",
		LastName:       "User",
		OrganizationID: "org_456",
		ConnectionID:   "conn_789",
		ConnectionType: "GoogleOAuth",
	}

	duration := 24 * time.Hour
	session := svc.CreateSessionData(profile, duration)

	if session.UserID != profile.ID {
		t.Errorf("CreateSessionData() UserID = %v, want %v", session.UserID, profile.ID)
	}
	if session.OrganizationID != profile.OrganizationID {
		t.Errorf("CreateSessionData() OrganizationID = %v, want %v", session.OrganizationID, profile.OrganizationID)
	}
	if session.Email != profile.Email {
		t.Errorf("CreateSessionData() Email = %v, want %v", session.Email, profile.Email)
	}
	// Check expiration is approximately correct (within 1 second)
	expectedExpiry := time.Now().Add(duration)
	if session.ExpiresAt.Before(expectedExpiry.Add(-time.Second)) || session.ExpiresAt.After(expectedExpiry.Add(time.Second)) {
		t.Errorf("CreateSessionData() ExpiresAt = %v, expected around %v", session.ExpiresAt, expectedExpiry)
	}
}

func TestDefaultAttributeMapping(t *testing.T) {
	mapping := DefaultAttributeMapping()

	if len(mapping.EmailAttributes) == 0 {
		t.Error("DefaultAttributeMapping() should have email attributes")
	}
	if len(mapping.FirstNameAttributes) == 0 {
		t.Error("DefaultAttributeMapping() should have first name attributes")
	}
	if len(mapping.LastNameAttributes) == 0 {
		t.Error("DefaultAttributeMapping() should have last name attributes")
	}
	if len(mapping.GroupsAttributes) == 0 {
		t.Error("DefaultAttributeMapping() should have groups attributes")
	}
}

func TestWorkOSService_ExtractUserInfo(t *testing.T) {
	svc := NewWorkOSService("sk_test_123", "client_123", "http://localhost:3000/callback", "12345678901234567890123456789012")

	rawAttrs := map[string]interface{}{
		"email":      "test@example.com",
		"firstName":  "Test",
		"lastName":   "User",
		"groups":     []interface{}{"admins", "developers"},
	}

	info := svc.ExtractUserInfo(rawAttrs, nil)

	if info["email"] != "test@example.com" {
		t.Errorf("ExtractUserInfo() email = %v, want %v", info["email"], "test@example.com")
	}
	if info["first_name"] != "Test" {
		t.Errorf("ExtractUserInfo() first_name = %v, want %v", info["first_name"], "Test")
	}
	if info["last_name"] != "User" {
		t.Errorf("ExtractUserInfo() last_name = %v, want %v", info["last_name"], "User")
	}
}
