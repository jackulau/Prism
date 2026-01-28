package remote

import (
	"sync"
	"testing"
	"time"

	"github.com/jacklau/prism/internal/security"
)

func TestSessionManager_CreateSession(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Stop()

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	clientInfo := map[string]string{"device": "test"}

	session, err := sm.CreateSession("user1", claims, clientInfo)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	if session.ID == "" {
		t.Error("Session ID should not be empty")
	}
	if session.UserID != "user1" {
		t.Errorf("Expected UserID 'user1', got '%s'", session.UserID)
	}
	if session.State != StateActive {
		t.Errorf("Expected State Active, got %v", session.State)
	}
	if session.ReconnectToken == "" {
		t.Error("ReconnectToken should not be empty")
	}
}

func TestSessionManager_GetSession(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Stop()

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	created, err := sm.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	retrieved, err := sm.GetSession(created.ID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Session IDs don't match: %s != %s", retrieved.ID, created.ID)
	}

	_, err = sm.GetSession("nonexistent")
	if err != ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got %v", err)
	}
}

func TestSessionManager_GetSessionByReconnectToken(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Stop()

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sm.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	retrieved, err := sm.GetSessionByReconnectToken(session.ReconnectToken)
	if err != nil {
		t.Fatalf("GetSessionByReconnectToken failed: %v", err)
	}

	if retrieved.ID != session.ID {
		t.Errorf("Session IDs don't match")
	}

	_, err = sm.GetSessionByReconnectToken("invalid_token")
	if err != ErrInvalidReconnectToken {
		t.Errorf("Expected ErrInvalidReconnectToken, got %v", err)
	}
}

func TestSessionManager_UpdateActivity(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Stop()

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sm.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	originalActivity := session.LastActivity

	time.Sleep(10 * time.Millisecond)

	err = sm.UpdateActivity(session.ID)
	if err != nil {
		t.Fatalf("UpdateActivity failed: %v", err)
	}

	updated, _ := sm.GetSession(session.ID)
	if !updated.LastActivity.After(originalActivity) {
		t.Error("LastActivity should be updated")
	}
}

func TestSessionManager_MarkDisconnected(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Stop()

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sm.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	err = sm.MarkDisconnected(session.ID)
	if err != nil {
		t.Fatalf("MarkDisconnected failed: %v", err)
	}

	updated, _ := sm.GetSession(session.ID)
	if updated.State != StateDisconnected {
		t.Errorf("Expected StateDisconnected, got %v", updated.State)
	}
	if updated.DisconnectedAt == nil {
		t.Error("DisconnectedAt should be set")
	}
}

func TestSessionManager_Reconnect(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Stop()

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sm.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	originalToken := session.ReconnectToken

	err = sm.MarkDisconnected(session.ID)
	if err != nil {
		t.Fatalf("MarkDisconnected failed: %v", err)
	}

	reconnected, err := sm.Reconnect(originalToken)
	if err != nil {
		t.Fatalf("Reconnect failed: %v", err)
	}

	if reconnected.State != StateActive {
		t.Errorf("Expected StateActive after reconnect, got %v", reconnected.State)
	}
	if reconnected.ReconnectToken == originalToken {
		t.Error("ReconnectToken should be rotated after reconnect")
	}
	if reconnected.DisconnectedAt != nil {
		t.Error("DisconnectedAt should be cleared after reconnect")
	}

	// Old token should not work
	_, err = sm.Reconnect(originalToken)
	if err != ErrInvalidReconnectToken {
		t.Errorf("Expected ErrInvalidReconnectToken with old token, got %v", err)
	}
}

func TestSessionManager_PendingMessages(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Stop()

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sm.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Can't add pending messages to active session
	err = sm.AddPendingMessage(session.ID, "test", "data")
	if err != ErrSessionNotDisconnected {
		t.Errorf("Expected ErrSessionNotDisconnected, got %v", err)
	}

	sm.MarkDisconnected(session.ID)

	// Add pending messages
	err = sm.AddPendingMessage(session.ID, "msg1", "data1")
	if err != nil {
		t.Fatalf("AddPendingMessage failed: %v", err)
	}
	err = sm.AddPendingMessage(session.ID, "msg2", "data2")
	if err != nil {
		t.Fatalf("AddPendingMessage failed: %v", err)
	}

	// Get and clear pending messages
	messages, err := sm.GetAndClearPendingMessages(session.ID)
	if err != nil {
		t.Fatalf("GetAndClearPendingMessages failed: %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("Expected 2 pending messages, got %d", len(messages))
	}

	// Should be cleared
	messages2, _ := sm.GetAndClearPendingMessages(session.ID)
	if len(messages2) != 0 {
		t.Errorf("Expected 0 pending messages after clear, got %d", len(messages2))
	}
}

func TestSessionManager_CloseSession(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Stop()

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sm.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	err = sm.CloseSession(session.ID)
	if err != nil {
		t.Fatalf("CloseSession failed: %v", err)
	}

	_, err = sm.GetSession(session.ID)
	if err != ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound after close, got %v", err)
	}
}

func TestSessionManager_GetUserSessions(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Stop()

	claims1 := &security.Claims{UserID: "user1", Email: "test1@example.com"}
	claims2 := &security.Claims{UserID: "user2", Email: "test2@example.com"}

	sm.CreateSession("user1", claims1, nil)
	sm.CreateSession("user1", claims1, nil)
	sm.CreateSession("user2", claims2, nil)

	user1Sessions := sm.GetUserSessions("user1")
	if len(user1Sessions) != 2 {
		t.Errorf("Expected 2 sessions for user1, got %d", len(user1Sessions))
	}

	user2Sessions := sm.GetUserSessions("user2")
	if len(user2Sessions) != 1 {
		t.Errorf("Expected 1 session for user2, got %d", len(user2Sessions))
	}
}

func TestSessionManager_ListSessions(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Stop()

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}

	sm.CreateSession("user1", claims, nil)
	sm.CreateSession("user1", claims, nil)
	sm.CreateSession("user2", claims, nil)

	sessions := sm.ListSessions()
	if len(sessions) != 3 {
		t.Errorf("Expected 3 sessions, got %d", len(sessions))
	}
}

func TestSessionManager_GetStats(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Stop()

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}

	s1, _ := sm.CreateSession("user1", claims, nil)
	sm.CreateSession("user1", claims, nil)
	sm.CreateSession("user2", claims, nil)

	sm.MarkDisconnected(s1.ID)

	stats := sm.GetStats()
	if stats.TotalSessions != 3 {
		t.Errorf("Expected TotalSessions 3, got %d", stats.TotalSessions)
	}
	if stats.ActiveSessions != 2 {
		t.Errorf("Expected ActiveSessions 2, got %d", stats.ActiveSessions)
	}
	if stats.DisconnectedSessions != 1 {
		t.Errorf("Expected DisconnectedSessions 1, got %d", stats.DisconnectedSessions)
	}
	if stats.TotalUsers != 2 {
		t.Errorf("Expected TotalUsers 2, got %d", stats.TotalUsers)
	}
}

func TestSessionManager_Callbacks(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Stop()

	var createdCalled, closedCalled, disconnectCalled, reconnectCalled bool
	var mu sync.Mutex

	sm.SetOnSessionCreated(func(*RemoteSession) {
		mu.Lock()
		createdCalled = true
		mu.Unlock()
	})
	sm.SetOnSessionClosed(func(*RemoteSession) {
		mu.Lock()
		closedCalled = true
		mu.Unlock()
	})
	sm.SetOnSessionDisconnect(func(*RemoteSession) {
		mu.Lock()
		disconnectCalled = true
		mu.Unlock()
	})
	sm.SetOnSessionReconnect(func(*RemoteSession) {
		mu.Lock()
		reconnectCalled = true
		mu.Unlock()
	})

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, _ := sm.CreateSession("user1", claims, nil)

	mu.Lock()
	if !createdCalled {
		t.Error("OnSessionCreated callback not called")
	}
	mu.Unlock()

	token := session.ReconnectToken
	sm.MarkDisconnected(session.ID)

	mu.Lock()
	if !disconnectCalled {
		t.Error("OnSessionDisconnect callback not called")
	}
	mu.Unlock()

	sm.Reconnect(token)

	mu.Lock()
	if !reconnectCalled {
		t.Error("OnSessionReconnect callback not called")
	}
	mu.Unlock()

	sm.CloseSession(session.ID)

	mu.Lock()
	if !closedCalled {
		t.Error("OnSessionClosed callback not called")
	}
	mu.Unlock()
}

func TestSessionManager_GetInfo(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Stop()

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	clientInfo := map[string]string{"device": "test", "os": "linux"}
	session, err := sm.CreateSession("user1", claims, clientInfo)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	info := session.GetInfo()

	if info.ID != session.ID {
		t.Errorf("Info ID mismatch")
	}
	if info.UserID != "user1" {
		t.Errorf("Info UserID mismatch")
	}
	if info.State != "active" {
		t.Errorf("Expected State 'active', got '%s'", info.State)
	}
	if info.ClientInfo["device"] != "test" {
		t.Error("ClientInfo not copied correctly")
	}
}

func TestSessionState_String(t *testing.T) {
	tests := []struct {
		state    SessionState
		expected string
	}{
		{StateActive, "active"},
		{StateDisconnected, "disconnected"},
		{StateReconnecting, "reconnecting"},
		{StateClosed, "closed"},
		{SessionState(99), "unknown"},
	}

	for _, tt := range tests {
		result := tt.state.String()
		if result != tt.expected {
			t.Errorf("SessionState(%d).String() = %s, expected %s", tt.state, result, tt.expected)
		}
	}
}
