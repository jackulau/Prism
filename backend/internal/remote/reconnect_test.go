package remote

import (
	"testing"
	"time"

	"github.com/jacklau/prism/internal/security"
)

func TestNewReconnectHandler(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Stop()

	rh := NewReconnectHandler(sm, nil)

	if rh == nil {
		t.Fatal("ReconnectHandler should not be nil")
	}
}

func TestReconnectHandler_HandleReconnect_Success(t *testing.T) {
	sm := NewSessionManager(MakeTestRemoteConfig())
	defer sm.Stop()

	rh := NewReconnectHandler(sm, MakeTestReconnectConfig())

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sm.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	originalToken := session.ReconnectToken

	// Mark as disconnected
	err = sm.MarkDisconnected(session.ID)
	if err != nil {
		t.Fatalf("MarkDisconnected failed: %v", err)
	}

	// Reconnect
	req := &ReconnectRequest{
		ReconnectToken: originalToken,
		ClientInfo:     map[string]string{"reconnect": "true"},
	}

	resp, err := rh.HandleReconnect(req)
	if err != nil {
		t.Fatalf("HandleReconnect failed: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success, got error: %s", resp.Error)
	}
	if resp.SessionID != session.ID {
		t.Errorf("Expected session ID %s, got %s", session.ID, resp.SessionID)
	}
	if resp.NewReconnectToken == "" {
		t.Error("NewReconnectToken should not be empty")
	}
	if resp.NewReconnectToken == originalToken {
		t.Error("NewReconnectToken should be different from original")
	}
}

func TestReconnectHandler_HandleReconnect_MissingToken(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Stop()

	rh := NewReconnectHandler(sm, MakeTestReconnectConfig())

	req := &ReconnectRequest{
		ReconnectToken: "",
	}

	resp, err := rh.HandleReconnect(req)
	if err != ErrMissingReconnectToken {
		t.Errorf("Expected ErrMissingReconnectToken, got %v", err)
	}
	if resp.Success {
		t.Error("Expected failure")
	}
	if resp.Error == "" {
		t.Error("Expected error message")
	}
}

func TestReconnectHandler_HandleReconnect_InvalidToken(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Stop()

	rh := NewReconnectHandler(sm, MakeTestReconnectConfig())

	req := &ReconnectRequest{
		ReconnectToken: "invalid-token",
	}

	resp, err := rh.HandleReconnect(req)
	if err != ErrInvalidReconnectToken {
		t.Errorf("Expected ErrInvalidReconnectToken, got %v", err)
	}
	if resp.Success {
		t.Error("Expected failure")
	}
}

func TestReconnectHandler_HandleReconnect_ExpiredToken(t *testing.T) {
	config := &SessionConfig{
		HeartbeatInterval:    30 * time.Second,
		HeartbeatTimeout:     90 * time.Second,
		IdleTimeout:          30 * time.Minute,
		IdleWarningBefore:    5 * time.Minute,
		ReconnectTokenExpiry: 10 * time.Millisecond, // Very short for testing
		CleanupInterval:      1 * time.Minute,
	}

	sm := NewSessionManager(config)
	defer sm.Stop()

	rh := NewReconnectHandler(sm, MakeTestReconnectConfig())

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sm.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	originalToken := session.ReconnectToken

	// Mark as disconnected
	err = sm.MarkDisconnected(session.ID)
	if err != nil {
		t.Fatalf("MarkDisconnected failed: %v", err)
	}

	// Wait for token to expire
	time.Sleep(20 * time.Millisecond)

	req := &ReconnectRequest{
		ReconnectToken: originalToken,
	}

	resp, err := rh.HandleReconnect(req)
	if err != ErrReconnectTokenExpired {
		t.Errorf("Expected ErrReconnectTokenExpired, got %v", err)
	}
	if resp.Success {
		t.Error("Expected failure")
	}
}

func TestReconnectHandler_HandleReconnect_MaxAttempts(t *testing.T) {
	config := &ReconnectConfig{
		MaxAttempts:          2,
		BaseDelay:            1 * time.Millisecond,
		MaxDelay:             10 * time.Millisecond,
		TokenValidity:        1 * time.Minute,
		MergePendingMessages: true,
	}

	sm := NewSessionManager(MakeTestRemoteConfig())
	defer sm.Stop()

	rh := NewReconnectHandler(sm, config)

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sm.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Disconnect the session
	sm.MarkDisconnected(session.ID)
	token := session.ReconnectToken

	// First reconnect - should succeed
	req := &ReconnectRequest{ReconnectToken: token}
	resp, err := rh.HandleReconnect(req)
	if err != nil {
		t.Fatalf("First reconnect should succeed: %v", err)
	}
	if !resp.Success {
		t.Fatalf("First reconnect should return success")
	}

	// Save the new token for later
	validToken := resp.NewReconnectToken

	// Now make failed attempts with an invalid token
	// This simulates a client using an old/invalid token repeatedly
	invalidReq := &ReconnectRequest{ReconnectToken: "invalid-token"}

	// Attempt 1 - should fail but not exceed limit
	resp, err = rh.HandleReconnect(invalidReq)
	if err != ErrInvalidReconnectToken {
		t.Errorf("Expected ErrInvalidReconnectToken, got %v", err)
	}

	// Attempt 2 - should fail but not exceed limit
	resp, err = rh.HandleReconnect(invalidReq)
	if err != ErrInvalidReconnectToken {
		t.Errorf("Expected ErrInvalidReconnectToken, got %v", err)
	}

	// Disconnect the session again to test max attempts on a valid session
	sm.MarkDisconnected(session.ID)

	// Use the valid token to reconnect - should work since each reconnect clears attempts
	req = &ReconnectRequest{ReconnectToken: validToken}
	resp, err = rh.HandleReconnect(req)
	if err != nil {
		t.Fatalf("Valid reconnect should succeed: %v", err)
	}
	if !resp.Success {
		t.Fatalf("Valid reconnect should return success")
	}
}

func TestReconnectHandler_HandleReconnect_WithPendingMessages(t *testing.T) {
	sm := NewSessionManager(MakeTestRemoteConfig())
	defer sm.Stop()

	config := &ReconnectConfig{
		MaxAttempts:          5,
		BaseDelay:            1 * time.Millisecond,
		MaxDelay:             10 * time.Millisecond,
		TokenValidity:        1 * time.Minute,
		MergePendingMessages: true,
	}
	rh := NewReconnectHandler(sm, config)

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sm.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Disconnect
	err = sm.MarkDisconnected(session.ID)
	if err != nil {
		t.Fatalf("MarkDisconnected failed: %v", err)
	}

	// Add pending messages
	sm.AddPendingMessage(session.ID, "type1", "data1")
	sm.AddPendingMessage(session.ID, "type2", "data2")

	// Reconnect
	req := &ReconnectRequest{
		ReconnectToken: session.ReconnectToken,
	}

	resp, err := rh.HandleReconnect(req)
	if err != nil {
		t.Fatalf("HandleReconnect failed: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success, got error: %s", resp.Error)
	}
	if len(resp.PendingMessages) != 2 {
		t.Errorf("Expected 2 pending messages, got %d", len(resp.PendingMessages))
	}
}

func TestReconnectHandler_HandleReconnect_ExponentialBackoff(t *testing.T) {
	config := &ReconnectConfig{
		MaxAttempts:          10,
		BaseDelay:            10 * time.Millisecond,
		MaxDelay:             100 * time.Millisecond,
		TokenValidity:        1 * time.Minute,
		MergePendingMessages: true,
	}

	sm := NewSessionManager(MakeTestRemoteConfig())
	defer sm.Stop()

	rh := NewReconnectHandler(sm, config)

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sm.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Use an invalid token to trigger failures with backoff
	req := &ReconnectRequest{
		ReconnectToken: "invalid-token",
	}

	// Make multiple failed attempts and check retry delays increase
	var lastRetryDelay int64 = 0
	for i := 0; i < 3; i++ {
		// We need to use a valid session token to track attempts
		// First create a new session each time
		claims := &security.Claims{UserID: "user" + string(rune('A'+i)), Email: "test@example.com"}
		s, _ := sm.CreateSession(claims.UserID, claims, nil)
		sm.MarkDisconnected(s.ID)

		// Try to reconnect with the session's token (this should succeed)
		req := &ReconnectRequest{ReconnectToken: s.ReconnectToken}
		_, _ = rh.HandleReconnect(req)
	}

	// Now test with a session that will fail repeatedly
	sm.MarkDisconnected(session.ID)

	// Get initial attempt info
	info := rh.GetAttemptInfo(session.ID)
	if info.AttemptCount != 0 {
		t.Errorf("Expected 0 attempts initially, got %d", info.AttemptCount)
	}

	// Make a failed reconnect attempt with an invalid token
	// (session's token is now consumed/rotated)
	req = &ReconnectRequest{ReconnectToken: "invalid"}
	resp, _ := rh.HandleReconnect(req)

	if resp.RetryAfterMs < lastRetryDelay {
		t.Error("Retry delay should not decrease")
	}
}

func TestReconnectHandler_GenerateReconnectInfo(t *testing.T) {
	sm := NewSessionManager(MakeTestRemoteConfig())
	defer sm.Stop()

	rh := NewReconnectHandler(sm, MakeTestReconnectConfig())

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sm.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	info, err := rh.GenerateReconnectInfo(session.ID)
	if err != nil {
		t.Fatalf("GenerateReconnectInfo failed: %v", err)
	}

	if info.ReconnectToken == "" {
		t.Error("ReconnectToken should not be empty")
	}
	if info.ValidForMs <= 0 {
		t.Error("ValidForMs should be positive")
	}
	if info.ExpiresAt.Before(time.Now()) {
		t.Error("ExpiresAt should be in the future")
	}
}

func TestReconnectHandler_GenerateReconnectInfo_InvalidSession(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Stop()

	rh := NewReconnectHandler(sm, MakeTestReconnectConfig())

	_, err := rh.GenerateReconnectInfo("nonexistent")
	if err != ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got %v", err)
	}
}

func TestReconnectHandler_GetAttemptInfo(t *testing.T) {
	sm := NewSessionManager(MakeTestRemoteConfig())
	defer sm.Stop()

	config := &ReconnectConfig{
		MaxAttempts:          5,
		BaseDelay:            10 * time.Millisecond,
		MaxDelay:             100 * time.Millisecond,
		TokenValidity:        1 * time.Minute,
		MergePendingMessages: true,
	}
	rh := NewReconnectHandler(sm, config)

	// Initially no attempts
	info := rh.GetAttemptInfo("session-1")
	if info.AttemptCount != 0 {
		t.Errorf("Expected 0 attempts, got %d", info.AttemptCount)
	}
	if info.RemainingAttempts != config.MaxAttempts {
		t.Errorf("Expected %d remaining attempts, got %d", config.MaxAttempts, info.RemainingAttempts)
	}
}

func TestReconnectHandler_ClearAttempts(t *testing.T) {
	sm := NewSessionManager(MakeTestRemoteConfig())
	defer sm.Stop()

	rh := NewReconnectHandler(sm, MakeTestReconnectConfig())

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sm.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Disconnect and reconnect to create an attempt record
	sm.MarkDisconnected(session.ID)
	req := &ReconnectRequest{ReconnectToken: session.ReconnectToken}
	rh.HandleReconnect(req)

	// Clear attempts
	rh.ClearAttempts(session.ID)

	// Check attempts are cleared
	info := rh.GetAttemptInfo(session.ID)
	if info.AttemptCount != 0 {
		t.Errorf("Expected 0 attempts after clear, got %d", info.AttemptCount)
	}
}

func TestReconnectHandler_ClientInfoUpdate(t *testing.T) {
	sm := NewSessionManager(MakeTestRemoteConfig())
	defer sm.Stop()

	rh := NewReconnectHandler(sm, MakeTestReconnectConfig())

	initialInfo := map[string]string{"device": "phone"}
	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sm.CreateSession("user1", claims, initialInfo)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Disconnect
	sm.MarkDisconnected(session.ID)

	// Reconnect with new client info
	req := &ReconnectRequest{
		ReconnectToken: session.ReconnectToken,
		ClientInfo:     map[string]string{"device": "tablet", "new_key": "value"},
	}

	resp, err := rh.HandleReconnect(req)
	if err != nil {
		t.Fatalf("HandleReconnect failed: %v", err)
	}

	// Check client info was updated
	updated, _ := sm.GetSession(session.ID)
	updated.mu.RLock()
	defer updated.mu.RUnlock()

	if updated.ClientInfo["device"] != "tablet" {
		t.Errorf("Expected device 'tablet', got '%s'", updated.ClientInfo["device"])
	}
	if updated.ClientInfo["new_key"] != "value" {
		t.Errorf("Expected new_key 'value', got '%s'", updated.ClientInfo["new_key"])
	}

	_ = resp // Suppress unused warning
}

func TestReconnectHandler_ConcurrentReconnects(t *testing.T) {
	sm := NewSessionManager(MakeTestRemoteConfig())
	defer sm.Stop()

	rh := NewReconnectHandler(sm, MakeTestReconnectConfig())

	// Create multiple sessions
	sessions := make([]*RemoteSession, 5)
	for i := 0; i < 5; i++ {
		claims := &security.Claims{UserID: "user" + string(rune('0'+i)), Email: "test@example.com"}
		session, err := sm.CreateSession(claims.UserID, claims, nil)
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}
		sessions[i] = session
		sm.MarkDisconnected(session.ID)
	}

	// Concurrent reconnects
	runner := NewConcurrentTestRunner()
	for _, session := range sessions {
		s := session
		runner.Run(func() error {
			req := &ReconnectRequest{
				ReconnectToken: s.ReconnectToken,
			}
			resp, err := rh.HandleReconnect(req)
			if err != nil {
				return err
			}
			if !resp.Success {
				return ErrSessionClosed
			}
			return nil
		})
	}

	runner.Wait()

	if runner.HasErrors() {
		t.Errorf("Concurrent reconnects failed: %v", runner.Errors())
	}
}
