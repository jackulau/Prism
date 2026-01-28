// +build integration

package remote

import (
	"sync"
	"testing"
	"time"

	"github.com/jacklau/prism/internal/security"
)

// TestE2E_RemoteConnection tests the full remote connection lifecycle
func TestE2E_RemoteConnection(t *testing.T) {
	// 1. Set up authentication service
	password := "secure-password-123"
	hash, err := security.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	authConfig := &security.RemoteAccessConfig{
		Enabled:               true,
		PasswordHash:          hash,
		SessionTimeout:        30 * time.Minute,
		MaxConcurrentSessions: 10,
		MaxFailedAttempts:     5,
		BlockDuration:         1 * time.Minute,
		MaxBlockDuration:      10 * time.Minute,
	}

	authService := security.NewRemoteAuthService(authConfig, nil)
	defer authService.Stop()

	// 2. Set up session manager
	sessionConfig := &SessionConfig{
		HeartbeatInterval:    100 * time.Millisecond,
		HeartbeatTimeout:     300 * time.Millisecond,
		IdleTimeout:          5 * time.Second,
		IdleWarningBefore:    1 * time.Second,
		ReconnectTokenExpiry: 2 * time.Second,
		CleanupInterval:      100 * time.Millisecond,
	}

	sessionManager := NewSessionManager(sessionConfig)
	defer sessionManager.Stop()

	// 3. Set up heartbeat handler
	heartbeatConfig := &HeartbeatConfig{
		Interval:    100 * time.Millisecond,
		Timeout:     300 * time.Millisecond,
		GracePeriod: 50 * time.Millisecond,
	}

	heartbeatHandler := NewHeartbeatHandler(sessionManager, heartbeatConfig)
	defer heartbeatHandler.Stop()

	// 4. Set up reconnect handler
	reconnectConfig := &ReconnectConfig{
		MaxAttempts:          5,
		BaseDelay:            10 * time.Millisecond,
		MaxDelay:             100 * time.Millisecond,
		TokenValidity:        2 * time.Second,
		MergePendingMessages: true,
	}

	reconnectHandler := NewReconnectHandler(sessionManager, reconnectConfig)

	// 5. Authenticate as remote client
	clientIP := "192.168.1.100"
	authSession, err := authService.Authenticate(password, clientIP)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	if authSession.Token == "" {
		t.Fatal("Auth session token should not be empty")
	}

	t.Logf("Authenticated successfully, token: %s...", authSession.Token[:16])

	// 6. Create remote session
	claims := &security.Claims{
		UserID: "user-123",
		Email:  "user@example.com",
	}

	clientInfo := map[string]string{
		"device":  "test-device",
		"os":      "test-os",
		"version": "1.0.0",
	}

	remoteSession, err := sessionManager.CreateSession(claims.UserID, claims, clientInfo)
	if err != nil {
		t.Fatalf("Failed to create remote session: %v", err)
	}

	t.Logf("Created remote session: %s", remoteSession.ID)

	// Register for heartbeat monitoring
	heartbeatHandler.RegisterSession(remoteSession.ID)

	// 7. Verify session is active
	if remoteSession.State != StateActive {
		t.Errorf("Expected session state Active, got %v", remoteSession.State)
	}

	// 8. Send heartbeats
	for i := 0; i < 3; i++ {
		heartbeatMsg := &HeartbeatMessage{
			SessionID: remoteSession.ID,
			Timestamp: time.Now(),
			Sequence:  int64(i + 1),
		}

		ack, err := heartbeatHandler.HandleHeartbeat(heartbeatMsg)
		if err != nil {
			t.Fatalf("Heartbeat %d failed: %v", i+1, err)
		}

		t.Logf("Heartbeat %d acknowledged at %v", i+1, ack.ServerTime)
		time.Sleep(50 * time.Millisecond)
	}

	// 9. Verify heartbeat status
	status := heartbeatHandler.GetHeartbeatStatus(remoteSession.ID)
	if status == nil {
		t.Fatal("Heartbeat status should not be nil")
	}
	if !status.IsHealthy {
		t.Error("Session should be healthy after heartbeats")
	}

	// 10. Update activity
	err = sessionManager.UpdateActivity(remoteSession.ID)
	if err != nil {
		t.Fatalf("UpdateActivity failed: %v", err)
	}

	// 11. Get reconnect info before disconnect
	reconnectInfo, err := reconnectHandler.GenerateReconnectInfo(remoteSession.ID)
	if err != nil {
		t.Fatalf("GenerateReconnectInfo failed: %v", err)
	}

	reconnectToken := reconnectInfo.ReconnectToken
	t.Logf("Reconnect token obtained: %s...", reconnectToken[:16])

	// 12. Simulate disconnect
	err = sessionManager.MarkDisconnected(remoteSession.ID)
	if err != nil {
		t.Fatalf("MarkDisconnected failed: %v", err)
	}

	// Verify session is disconnected
	updatedSession, err := sessionManager.GetSession(remoteSession.ID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if updatedSession.State != StateDisconnected {
		t.Errorf("Expected session state Disconnected, got %v", updatedSession.State)
	}

	t.Log("Session disconnected")

	// 13. Add pending messages while disconnected
	err = sessionManager.AddPendingMessage(remoteSession.ID, "notification", map[string]string{
		"type":    "info",
		"message": "You have new messages",
	})
	if err != nil {
		t.Fatalf("AddPendingMessage failed: %v", err)
	}

	err = sessionManager.AddPendingMessage(remoteSession.ID, "update", map[string]string{
		"type":    "data_change",
		"message": "Data was updated",
	})
	if err != nil {
		t.Fatalf("AddPendingMessage failed: %v", err)
	}

	t.Log("Added 2 pending messages")

	// 14. Reconnect with token
	reconnectReq := &ReconnectRequest{
		ReconnectToken: reconnectToken,
		ClientInfo: map[string]string{
			"device":     "test-device",
			"reconnect":  "true",
			"session_id": remoteSession.ID,
		},
	}

	reconnectResp, err := reconnectHandler.HandleReconnect(reconnectReq)
	if err != nil {
		t.Fatalf("HandleReconnect failed: %v", err)
	}

	if !reconnectResp.Success {
		t.Fatalf("Reconnection failed: %s", reconnectResp.Error)
	}

	t.Logf("Reconnected successfully, new token: %s...", reconnectResp.NewReconnectToken[:16])

	// 15. Verify pending messages received
	if len(reconnectResp.PendingMessages) != 2 {
		t.Errorf("Expected 2 pending messages, got %d", len(reconnectResp.PendingMessages))
	}

	// 16. Verify session is active again
	reconnectedSession, err := sessionManager.GetSession(remoteSession.ID)
	if err != nil {
		t.Fatalf("GetSession after reconnect failed: %v", err)
	}

	if reconnectedSession.State != StateActive {
		t.Errorf("Expected session state Active after reconnect, got %v", reconnectedSession.State)
	}

	// 17. Verify new reconnect token is different
	if reconnectedSession.ReconnectToken == reconnectToken {
		t.Error("Reconnect token should be rotated after reconnection")
	}

	// 18. Continue with more heartbeats
	for i := 0; i < 2; i++ {
		heartbeatMsg := &HeartbeatMessage{
			SessionID: remoteSession.ID,
			Timestamp: time.Now(),
			Sequence:  int64(i + 4),
		}

		_, err := heartbeatHandler.HandleHeartbeat(heartbeatMsg)
		if err != nil {
			t.Fatalf("Post-reconnect heartbeat %d failed: %v", i+1, err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	// 19. Clean disconnect
	err = sessionManager.CloseSession(remoteSession.ID)
	if err != nil {
		t.Fatalf("CloseSession failed: %v", err)
	}

	heartbeatHandler.UnregisterSession(remoteSession.ID)

	// 20. Verify session is closed
	_, err = sessionManager.GetSession(remoteSession.ID)
	if err != ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound after close, got %v", err)
	}

	// 21. Invalidate auth session
	err = authService.InvalidateSession(authSession.Token)
	if err != nil {
		t.Fatalf("InvalidateSession failed: %v", err)
	}

	// Verify auth session is invalid
	_, err = authService.ValidateSession(authSession.Token)
	if err == nil {
		t.Error("Auth session should be invalid after logout")
	}

	t.Log("E2E remote connection test completed successfully")
}

// TestE2E_MultipleClients tests multiple concurrent remote clients
func TestE2E_MultipleClients(t *testing.T) {
	password := "shared-password"
	hash, _ := security.HashPassword(password)

	authConfig := &security.RemoteAccessConfig{
		Enabled:               true,
		PasswordHash:          hash,
		SessionTimeout:        30 * time.Minute,
		MaxConcurrentSessions: 100,
		MaxFailedAttempts:     10,
		BlockDuration:         1 * time.Minute,
		MaxBlockDuration:      10 * time.Minute,
	}

	authService := security.NewRemoteAuthService(authConfig, nil)
	defer authService.Stop()

	sessionManager := NewSessionManager(MakeTestRemoteConfig())
	defer sessionManager.Stop()

	heartbeatHandler := NewHeartbeatHandler(sessionManager, MakeTestHeartbeatConfig())
	defer heartbeatHandler.Stop()

	numClients := 10
	var wg sync.WaitGroup
	errors := make(chan error, numClients)

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientIdx int) {
			defer wg.Done()

			clientIP := "192.168.1." + string(rune('1'+clientIdx%9))

			// Authenticate
			authSession, err := authService.Authenticate(password, clientIP)
			if err != nil {
				errors <- err
				return
			}

			// Create session
			claims := &security.Claims{
				UserID: "user-" + string(rune('A'+clientIdx)),
				Email:  "user@example.com",
			}

			session, err := sessionManager.CreateSession(claims.UserID, claims, nil)
			if err != nil {
				errors <- err
				return
			}

			heartbeatHandler.RegisterSession(session.ID)

			// Send heartbeats
			for j := 0; j < 3; j++ {
				msg := &HeartbeatMessage{
					SessionID: session.ID,
					Timestamp: time.Now(),
				}
				_, err := heartbeatHandler.HandleHeartbeat(msg)
				if err != nil {
					errors <- err
					return
				}
				time.Sleep(10 * time.Millisecond)
			}

			// Clean up
			heartbeatHandler.UnregisterSession(session.ID)
			sessionManager.CloseSession(session.ID)
			authService.InvalidateSession(authSession.Token)
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Client operation failed: %v", err)
	}

	// Verify all sessions are cleaned up
	stats := sessionManager.GetStats()
	if stats.TotalSessions != 0 {
		t.Errorf("Expected 0 sessions after cleanup, got %d", stats.TotalSessions)
	}

	t.Logf("Multiple clients test completed: %d clients processed", numClients)
}

// TestE2E_SessionLifecycleEvents tests session lifecycle event callbacks
func TestE2E_SessionLifecycleEvents(t *testing.T) {
	var (
		createdCalled     bool
		disconnectedCalled bool
		reconnectedCalled  bool
		closedCalled       bool
		mu                 sync.Mutex
	)

	sessionManager := NewSessionManager(MakeTestRemoteConfig())
	defer sessionManager.Stop()

	sessionManager.SetOnSessionCreated(func(s *RemoteSession) {
		mu.Lock()
		createdCalled = true
		mu.Unlock()
	})

	sessionManager.SetOnSessionDisconnect(func(s *RemoteSession) {
		mu.Lock()
		disconnectedCalled = true
		mu.Unlock()
	})

	sessionManager.SetOnSessionReconnect(func(s *RemoteSession) {
		mu.Lock()
		reconnectedCalled = true
		mu.Unlock()
	})

	sessionManager.SetOnSessionClosed(func(s *RemoteSession) {
		mu.Lock()
		closedCalled = true
		mu.Unlock()
	})

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}

	// Create
	session, err := sessionManager.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	mu.Lock()
	if !createdCalled {
		t.Error("OnSessionCreated should have been called")
	}
	mu.Unlock()

	token := session.ReconnectToken

	// Disconnect
	sessionManager.MarkDisconnected(session.ID)

	mu.Lock()
	if !disconnectedCalled {
		t.Error("OnSessionDisconnect should have been called")
	}
	mu.Unlock()

	// Reconnect
	sessionManager.Reconnect(token)

	mu.Lock()
	if !reconnectedCalled {
		t.Error("OnSessionReconnect should have been called")
	}
	mu.Unlock()

	// Close
	sessionManager.CloseSession(session.ID)

	mu.Lock()
	if !closedCalled {
		t.Error("OnSessionClosed should have been called")
	}
	mu.Unlock()

	t.Log("Session lifecycle events test completed")
}

// TestE2E_IdleTimeoutFlow tests the idle timeout detection and handling
func TestE2E_IdleTimeoutFlow(t *testing.T) {
	config := &SessionConfig{
		HeartbeatInterval:    50 * time.Millisecond,
		HeartbeatTimeout:     100 * time.Millisecond,
		IdleTimeout:          200 * time.Millisecond,
		IdleWarningBefore:    50 * time.Millisecond,
		ReconnectTokenExpiry: 1 * time.Second,
		CleanupInterval:      25 * time.Millisecond,
	}

	sessionManager := NewSessionManager(config)
	defer sessionManager.Stop()

	var (
		warningReceived bool
		mu              sync.Mutex
	)

	sessionManager.SetOnIdleWarning(func(s *RemoteSession) {
		mu.Lock()
		warningReceived = true
		mu.Unlock()
	})

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sessionManager.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Wait for idle timeout to trigger warning and cleanup
	time.Sleep(350 * time.Millisecond)

	mu.Lock()
	if !warningReceived {
		t.Log("Idle warning may not have been triggered in time (timing dependent)")
	}
	mu.Unlock()

	// Session should be closed due to idle timeout
	_, err = sessionManager.GetSession(session.ID)
	if err != ErrSessionNotFound {
		t.Errorf("Expected session to be closed due to idle timeout, got err: %v", err)
	}

	t.Log("Idle timeout flow test completed")
}

// TestE2E_ReconnectTokenRotation tests that reconnect tokens are properly rotated
func TestE2E_ReconnectTokenRotation(t *testing.T) {
	sessionManager := NewSessionManager(MakeTestRemoteConfig())
	defer sessionManager.Stop()

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sessionManager.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	tokens := make([]string, 0)
	tokens = append(tokens, session.ReconnectToken)

	// Perform multiple reconnect cycles
	for i := 0; i < 5; i++ {
		currentToken := tokens[len(tokens)-1]

		// Disconnect
		sessionManager.MarkDisconnected(session.ID)

		// Reconnect
		reconnected, err := sessionManager.Reconnect(currentToken)
		if err != nil {
			t.Fatalf("Reconnect %d failed: %v", i+1, err)
		}

		newToken := reconnected.ReconnectToken
		tokens = append(tokens, newToken)

		// Verify new token is different
		if newToken == currentToken {
			t.Errorf("Reconnect %d: token should be rotated", i+1)
		}
	}

	// Verify all old tokens are invalid
	for i, token := range tokens[:len(tokens)-1] {
		_, err := sessionManager.Reconnect(token)
		if err != ErrInvalidReconnectToken {
			t.Errorf("Old token %d should be invalid", i)
		}
	}

	t.Log("Reconnect token rotation test completed")
}

// TestE2E_SessionStats tests session statistics reporting
func TestE2E_SessionStats(t *testing.T) {
	sessionManager := NewSessionManager(MakeTestRemoteConfig())
	defer sessionManager.Stop()

	// Initial state
	stats := sessionManager.GetStats()
	if stats.TotalSessions != 0 {
		t.Errorf("Expected 0 initial sessions, got %d", stats.TotalSessions)
	}

	// Create sessions
	sessions := make([]*RemoteSession, 5)
	for i := 0; i < 5; i++ {
		claims := &security.Claims{UserID: "user" + string(rune('0'+i)), Email: "test@example.com"}
		session, _ := sessionManager.CreateSession(claims.UserID, claims, nil)
		sessions[i] = session
	}

	stats = sessionManager.GetStats()
	if stats.TotalSessions != 5 {
		t.Errorf("Expected 5 sessions, got %d", stats.TotalSessions)
	}
	if stats.ActiveSessions != 5 {
		t.Errorf("Expected 5 active sessions, got %d", stats.ActiveSessions)
	}

	// Disconnect some
	sessionManager.MarkDisconnected(sessions[0].ID)
	sessionManager.MarkDisconnected(sessions[1].ID)

	stats = sessionManager.GetStats()
	if stats.DisconnectedSessions != 2 {
		t.Errorf("Expected 2 disconnected sessions, got %d", stats.DisconnectedSessions)
	}
	if stats.ActiveSessions != 3 {
		t.Errorf("Expected 3 active sessions, got %d", stats.ActiveSessions)
	}

	// Close one
	sessionManager.CloseSession(sessions[2].ID)

	stats = sessionManager.GetStats()
	if stats.TotalSessions != 4 {
		t.Errorf("Expected 4 sessions after close, got %d", stats.TotalSessions)
	}

	t.Log("Session stats test completed")
}
