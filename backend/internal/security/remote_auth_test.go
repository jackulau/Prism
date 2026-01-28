package security

import (
	"testing"
	"time"
)

func TestNewRemoteAuthService(t *testing.T) {
	config := &RemoteAccessConfig{
		Enabled:             true,
		PasswordHash:        mustHashPassword(t, "testpassword"),
		SessionTimeout:      30 * time.Minute,
		MaxConcurrentSessions: 5,
		MaxFailedAttempts:   3,
		BlockDuration:       1 * time.Minute,
		MaxBlockDuration:    10 * time.Minute,
	}

	service := NewRemoteAuthService(config, nil)
	defer service.Stop()

	if service == nil {
		t.Fatal("expected service to be created")
	}

	if !service.IsEnabled() {
		t.Error("expected service to be enabled")
	}
}

func TestNewRemoteAuthService_NilConfig(t *testing.T) {
	service := NewRemoteAuthService(nil, nil)
	defer service.Stop()

	if service == nil {
		t.Fatal("expected service to be created")
	}

	if service.IsEnabled() {
		t.Error("expected service to be disabled with nil config")
	}
}

func TestAuthenticate_Success(t *testing.T) {
	password := "testpassword123"
	hash := mustHashPassword(t, password)

	config := &RemoteAccessConfig{
		Enabled:             true,
		PasswordHash:        hash,
		SessionTimeout:      30 * time.Minute,
		MaxConcurrentSessions: 5,
		MaxFailedAttempts:   3,
		BlockDuration:       1 * time.Minute,
		MaxBlockDuration:    10 * time.Minute,
	}

	service := NewRemoteAuthService(config, nil)
	defer service.Stop()

	session, err := service.Authenticate(password, "192.168.1.1")
	if err != nil {
		t.Fatalf("expected authentication to succeed: %v", err)
	}

	if session == nil {
		t.Fatal("expected session to be returned")
	}

	if session.Token == "" {
		t.Error("expected session token to be non-empty")
	}

	if session.ClientIP != "192.168.1.1" {
		t.Errorf("expected client IP 192.168.1.1, got %s", session.ClientIP)
	}

	if !session.IsActive {
		t.Error("expected session to be active")
	}

	if session.ExpiresAt.Before(time.Now()) {
		t.Error("expected session expiry to be in the future")
	}
}

func TestAuthenticate_InvalidPassword(t *testing.T) {
	hash := mustHashPassword(t, "correctpassword")

	config := &RemoteAccessConfig{
		Enabled:             true,
		PasswordHash:        hash,
		SessionTimeout:      30 * time.Minute,
		MaxConcurrentSessions: 5,
		MaxFailedAttempts:   3,
		BlockDuration:       1 * time.Minute,
		MaxBlockDuration:    10 * time.Minute,
	}

	service := NewRemoteAuthService(config, nil)
	defer service.Stop()

	session, err := service.Authenticate("wrongpassword", "192.168.1.1")
	if err == nil {
		t.Error("expected authentication to fail")
	}

	if session != nil {
		t.Error("expected no session to be returned")
	}
}

func TestAuthenticate_Disabled(t *testing.T) {
	config := &RemoteAccessConfig{
		Enabled:      false,
		PasswordHash: "somehash",
	}

	service := NewRemoteAuthService(config, nil)
	defer service.Stop()

	session, err := service.Authenticate("password", "192.168.1.1")
	if err == nil {
		t.Error("expected authentication to fail when disabled")
	}

	if session != nil {
		t.Error("expected no session")
	}
}

func TestValidateSession(t *testing.T) {
	password := "testpassword"
	hash := mustHashPassword(t, password)

	config := &RemoteAccessConfig{
		Enabled:             true,
		PasswordHash:        hash,
		SessionTimeout:      30 * time.Minute,
		MaxConcurrentSessions: 5,
		MaxFailedAttempts:   3,
		BlockDuration:       1 * time.Minute,
		MaxBlockDuration:    10 * time.Minute,
	}

	service := NewRemoteAuthService(config, nil)
	defer service.Stop()

	session, err := service.Authenticate(password, "192.168.1.1")
	if err != nil {
		t.Fatalf("authentication failed: %v", err)
	}

	validated, err := service.ValidateSession(session.Token)
	if err != nil {
		t.Fatalf("session validation failed: %v", err)
	}

	if validated.Token != session.Token {
		t.Error("validated session token doesn't match")
	}
}

func TestValidateSession_NotFound(t *testing.T) {
	service := NewRemoteAuthService(DefaultRemoteAccessConfig(), nil)
	defer service.Stop()

	_, err := service.ValidateSession("nonexistenttoken")
	if err == nil {
		t.Error("expected error for non-existent session")
	}
}

func TestValidateSessionWithIP_Success(t *testing.T) {
	password := "testpassword"
	hash := mustHashPassword(t, password)
	clientIP := "192.168.1.1"

	config := &RemoteAccessConfig{
		Enabled:             true,
		PasswordHash:        hash,
		SessionTimeout:      30 * time.Minute,
		MaxConcurrentSessions: 5,
		MaxFailedAttempts:   3,
		BlockDuration:       1 * time.Minute,
		MaxBlockDuration:    10 * time.Minute,
	}

	service := NewRemoteAuthService(config, nil)
	defer service.Stop()

	session, err := service.Authenticate(password, clientIP)
	if err != nil {
		t.Fatalf("authentication failed: %v", err)
	}

	validated, err := service.ValidateSessionWithIP(session.Token, clientIP)
	if err != nil {
		t.Fatalf("session validation with IP failed: %v", err)
	}

	if validated.ClientIP != clientIP {
		t.Error("IP mismatch in validated session")
	}
}

func TestValidateSessionWithIP_Mismatch(t *testing.T) {
	password := "testpassword"
	hash := mustHashPassword(t, password)

	config := &RemoteAccessConfig{
		Enabled:             true,
		PasswordHash:        hash,
		SessionTimeout:      30 * time.Minute,
		MaxConcurrentSessions: 5,
		MaxFailedAttempts:   3,
		BlockDuration:       1 * time.Minute,
		MaxBlockDuration:    10 * time.Minute,
	}

	service := NewRemoteAuthService(config, nil)
	defer service.Stop()

	session, err := service.Authenticate(password, "192.168.1.1")
	if err != nil {
		t.Fatalf("authentication failed: %v", err)
	}

	_, err = service.ValidateSessionWithIP(session.Token, "192.168.1.2")
	if err == nil {
		t.Error("expected error for IP mismatch")
	}
}

func TestInvalidateSession(t *testing.T) {
	password := "testpassword"
	hash := mustHashPassword(t, password)

	config := &RemoteAccessConfig{
		Enabled:             true,
		PasswordHash:        hash,
		SessionTimeout:      30 * time.Minute,
		MaxConcurrentSessions: 5,
		MaxFailedAttempts:   3,
		BlockDuration:       1 * time.Minute,
		MaxBlockDuration:    10 * time.Minute,
	}

	service := NewRemoteAuthService(config, nil)
	defer service.Stop()

	session, err := service.Authenticate(password, "192.168.1.1")
	if err != nil {
		t.Fatalf("authentication failed: %v", err)
	}

	err = service.InvalidateSession(session.Token)
	if err != nil {
		t.Fatalf("failed to invalidate session: %v", err)
	}

	_, err = service.ValidateSession(session.Token)
	if err == nil {
		t.Error("expected validation to fail after invalidation")
	}
}

func TestIPBlocking(t *testing.T) {
	hash := mustHashPassword(t, "correctpassword")

	config := &RemoteAccessConfig{
		Enabled:             true,
		PasswordHash:        hash,
		SessionTimeout:      30 * time.Minute,
		MaxConcurrentSessions: 5,
		MaxFailedAttempts:   3,
		BlockDuration:       1 * time.Second, // Short for testing
		MaxBlockDuration:    10 * time.Second,
	}

	service := NewRemoteAuthService(config, nil)
	defer service.Stop()

	clientIP := "192.168.1.100"

	// Make failed attempts
	for i := 0; i < 3; i++ {
		_, err := service.Authenticate("wrongpassword", clientIP)
		if err == nil {
			t.Error("expected authentication to fail")
		}
	}

	// Check if IP is blocked
	if !service.IsIPBlocked(clientIP) {
		t.Error("expected IP to be blocked after max failed attempts")
	}

	// Attempt should be rejected due to block
	_, err := service.Authenticate("correctpassword", clientIP)
	if err == nil {
		t.Error("expected authentication to be rejected while blocked")
	}

	// Wait for block to expire
	time.Sleep(1100 * time.Millisecond)

	if service.IsIPBlocked(clientIP) {
		t.Error("expected IP to be unblocked after block duration")
	}

	// Should be able to authenticate now
	session, err := service.Authenticate("correctpassword", clientIP)
	if err != nil {
		t.Fatalf("expected authentication to succeed after block expired: %v", err)
	}

	if session == nil {
		t.Error("expected session to be returned")
	}
}

func TestExponentialBackoff(t *testing.T) {
	hash := mustHashPassword(t, "correctpassword")

	config := &RemoteAccessConfig{
		Enabled:             true,
		PasswordHash:        hash,
		SessionTimeout:      30 * time.Minute,
		MaxConcurrentSessions: 5,
		MaxFailedAttempts:   2,
		BlockDuration:       100 * time.Millisecond,
		MaxBlockDuration:    1 * time.Second,
	}

	service := NewRemoteAuthService(config, nil)
	defer service.Stop()

	clientIP := "192.168.1.200"

	// First block - 100ms
	for i := 0; i < 2; i++ {
		service.Authenticate("wrong", clientIP)
	}

	firstBlockTime := service.GetBlockTimeRemaining(clientIP)

	// Wait for block to expire
	time.Sleep(150 * time.Millisecond)

	// Second block - should be 200ms (2x)
	for i := 0; i < 2; i++ {
		service.Authenticate("wrong", clientIP)
	}

	secondBlockTime := service.GetBlockTimeRemaining(clientIP)

	// Second block should be roughly 2x the first
	if secondBlockTime <= firstBlockTime {
		t.Errorf("expected exponential backoff: first block ~%v, second block ~%v", firstBlockTime, secondBlockTime)
	}
}

func TestMaxConcurrentSessions(t *testing.T) {
	password := "testpassword"
	hash := mustHashPassword(t, password)

	config := &RemoteAccessConfig{
		Enabled:             true,
		PasswordHash:        hash,
		SessionTimeout:      30 * time.Minute,
		MaxConcurrentSessions: 2, // Very low for testing
		MaxFailedAttempts:   3,
		BlockDuration:       1 * time.Minute,
		MaxBlockDuration:    10 * time.Minute,
	}

	service := NewRemoteAuthService(config, nil)
	defer service.Stop()

	// Create max sessions
	_, err := service.Authenticate(password, "192.168.1.1")
	if err != nil {
		t.Fatalf("first auth failed: %v", err)
	}

	_, err = service.Authenticate(password, "192.168.1.2")
	if err != nil {
		t.Fatalf("second auth failed: %v", err)
	}

	// Third should fail
	_, err = service.Authenticate(password, "192.168.1.3")
	if err == nil {
		t.Error("expected third authentication to fail due to max sessions")
	}
}

func TestCleanupExpiredSessions(t *testing.T) {
	password := "testpassword"
	hash := mustHashPassword(t, password)

	config := &RemoteAccessConfig{
		Enabled:             true,
		PasswordHash:        hash,
		SessionTimeout:      100 * time.Millisecond, // Very short for testing
		MaxConcurrentSessions: 10,
		MaxFailedAttempts:   3,
		BlockDuration:       1 * time.Minute,
		MaxBlockDuration:    10 * time.Minute,
	}

	service := NewRemoteAuthService(config, nil)
	defer service.Stop()

	session, err := service.Authenticate(password, "192.168.1.1")
	if err != nil {
		t.Fatalf("authentication failed: %v", err)
	}

	// Session should be valid initially
	_, err = service.ValidateSession(session.Token)
	if err != nil {
		t.Error("expected session to be valid initially")
	}

	// Wait for session to expire
	time.Sleep(150 * time.Millisecond)

	// Session should be expired now
	_, err = service.ValidateSession(session.Token)
	if err == nil {
		t.Error("expected session validation to fail after expiry")
	}

	// Cleanup should remove it
	service.CleanupExpiredSessions()

	if service.GetActiveSessions() != 0 {
		t.Error("expected no active sessions after cleanup")
	}
}

func TestRefreshSession(t *testing.T) {
	password := "testpassword"
	hash := mustHashPassword(t, password)

	config := &RemoteAccessConfig{
		Enabled:             true,
		PasswordHash:        hash,
		SessionTimeout:      1 * time.Second,
		MaxConcurrentSessions: 5,
		MaxFailedAttempts:   3,
		BlockDuration:       1 * time.Minute,
		MaxBlockDuration:    10 * time.Minute,
	}

	service := NewRemoteAuthService(config, nil)
	defer service.Stop()

	session, err := service.Authenticate(password, "192.168.1.1")
	if err != nil {
		t.Fatalf("authentication failed: %v", err)
	}

	originalExpiry := session.ExpiresAt

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Refresh
	err = service.RefreshSession(session.Token)
	if err != nil {
		t.Fatalf("failed to refresh session: %v", err)
	}

	// Get updated session
	refreshed, _ := service.ValidateSession(session.Token)
	if !refreshed.ExpiresAt.After(originalExpiry) {
		t.Error("expected expiry to be extended after refresh")
	}
}

func TestClearFailedAttemptsOnSuccess(t *testing.T) {
	password := "testpassword"
	hash := mustHashPassword(t, password)

	config := &RemoteAccessConfig{
		Enabled:             true,
		PasswordHash:        hash,
		SessionTimeout:      30 * time.Minute,
		MaxConcurrentSessions: 10,
		MaxFailedAttempts:   5,
		BlockDuration:       1 * time.Minute,
		MaxBlockDuration:    10 * time.Minute,
	}

	service := NewRemoteAuthService(config, nil)
	defer service.Stop()

	clientIP := "192.168.1.1"

	// Make some failed attempts (but not enough to block)
	for i := 0; i < 3; i++ {
		service.Authenticate("wrong", clientIP)
	}

	// Successful auth should clear failed attempts
	session, err := service.Authenticate(password, clientIP)
	if err != nil {
		t.Fatalf("authentication should succeed: %v", err)
	}

	if session == nil {
		t.Fatal("expected session")
	}

	// Now make 4 more failed attempts - should NOT result in block
	// because the previous count was cleared
	for i := 0; i < 4; i++ {
		service.Authenticate("wrong", clientIP)
	}

	if service.IsIPBlocked(clientIP) {
		t.Error("expected IP not to be blocked after successful auth cleared counter")
	}
}

func TestGetActiveSessions(t *testing.T) {
	password := "testpassword"
	hash := mustHashPassword(t, password)

	config := &RemoteAccessConfig{
		Enabled:             true,
		PasswordHash:        hash,
		SessionTimeout:      30 * time.Minute,
		MaxConcurrentSessions: 10,
		MaxFailedAttempts:   5,
		BlockDuration:       1 * time.Minute,
		MaxBlockDuration:    10 * time.Minute,
	}

	service := NewRemoteAuthService(config, nil)
	defer service.Stop()

	if service.GetActiveSessions() != 0 {
		t.Error("expected 0 active sessions initially")
	}

	session1, _ := service.Authenticate(password, "192.168.1.1")
	if service.GetActiveSessions() != 1 {
		t.Error("expected 1 active session")
	}

	_, _ = service.Authenticate(password, "192.168.1.2")
	if service.GetActiveSessions() != 2 {
		t.Error("expected 2 active sessions")
	}

	service.InvalidateSession(session1.Token)
	if service.GetActiveSessions() != 1 {
		t.Error("expected 1 active session after invalidation")
	}
}

// mustHashPassword is a helper that creates a password hash or fails the test
func mustHashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return hash
}
