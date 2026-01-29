package remote

import (
	"testing"
	"time"

	"github.com/jacklau/prism/internal/security"
)

func TestNewHeartbeatHandler(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Stop()

	hh := NewHeartbeatHandler(sm, nil)
	defer hh.Stop()

	if hh == nil {
		t.Fatal("HeartbeatHandler should not be nil")
	}

	config := hh.GetConfig()
	if config == nil {
		t.Fatal("Config should not be nil")
	}

	// Check defaults
	if config.Interval != 30*time.Second {
		t.Errorf("Expected default interval 30s, got %v", config.Interval)
	}
	if config.Timeout != 90*time.Second {
		t.Errorf("Expected default timeout 90s, got %v", config.Timeout)
	}
}

func TestHeartbeatHandler_RegisterUnregister(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Stop()

	hh := NewHeartbeatHandler(sm, MakeTestHeartbeatConfig())
	defer hh.Stop()

	sessionID := "test-session-1"

	// Register
	hh.RegisterSession(sessionID)

	status := hh.GetHeartbeatStatus(sessionID)
	if status == nil {
		t.Fatal("Status should not be nil after registration")
	}
	if status.SessionID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, status.SessionID)
	}

	// Unregister
	hh.UnregisterSession(sessionID)

	status = hh.GetHeartbeatStatus(sessionID)
	if status != nil {
		t.Error("Status should be nil after unregistration")
	}
}

func TestHeartbeatHandler_HandleHeartbeat(t *testing.T) {
	sm := NewSessionManager(MakeTestRemoteConfig())
	defer sm.Stop()

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sm.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	hh := NewHeartbeatHandler(sm, MakeTestHeartbeatConfig())
	defer hh.Stop()

	hh.RegisterSession(session.ID)

	msg := &HeartbeatMessage{
		SessionID: session.ID,
		Timestamp: time.Now(),
		Sequence:  1,
	}

	ack, err := hh.HandleHeartbeat(msg)
	if err != nil {
		t.Fatalf("HandleHeartbeat failed: %v", err)
	}

	if ack == nil {
		t.Fatal("Ack should not be nil")
	}
	if ack.SessionID != session.ID {
		t.Errorf("Expected session ID %s, got %s", session.ID, ack.SessionID)
	}
	if ack.NextHeartbeatIn <= 0 {
		t.Error("NextHeartbeatIn should be positive")
	}
}

func TestHeartbeatHandler_GetHeartbeatStatus_Healthy(t *testing.T) {
	sm := NewSessionManager(MakeTestRemoteConfig())
	defer sm.Stop()

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sm.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	hh := NewHeartbeatHandler(sm, MakeTestHeartbeatConfig())
	defer hh.Stop()

	hh.RegisterSession(session.ID)

	// Send a heartbeat
	msg := &HeartbeatMessage{
		SessionID: session.ID,
		Timestamp: time.Now(),
	}
	hh.HandleHeartbeat(msg)

	status := hh.GetHeartbeatStatus(session.ID)
	if status == nil {
		t.Fatal("Status should not be nil")
	}
	if !status.IsHealthy {
		t.Error("Session should be healthy")
	}
	if status.State != "healthy" {
		t.Errorf("Expected state 'healthy', got '%s'", status.State)
	}
	if status.MissedCount != 0 {
		t.Errorf("Expected missed count 0, got %d", status.MissedCount)
	}
}

func TestHeartbeatHandler_GetHeartbeatStatus_Stale(t *testing.T) {
	config := &HeartbeatConfig{
		Interval:    10 * time.Millisecond,
		Timeout:     100 * time.Millisecond,
		GracePeriod: 5 * time.Millisecond,
	}

	sm := NewSessionManager(MakeTestRemoteConfig())
	defer sm.Stop()

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sm.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	hh := NewHeartbeatHandler(sm, config)
	defer hh.Stop()

	hh.RegisterSession(session.ID)

	// Wait for interval + grace period to pass
	time.Sleep(20 * time.Millisecond)

	status := hh.GetHeartbeatStatus(session.ID)
	if status == nil {
		t.Fatal("Status should not be nil")
	}
	if status.State != "stale" {
		t.Errorf("Expected state 'stale', got '%s'", status.State)
	}
}

func TestHeartbeatHandler_GetHeartbeatStatus_Dead(t *testing.T) {
	config := &HeartbeatConfig{
		Interval:    10 * time.Millisecond,
		Timeout:     30 * time.Millisecond,
		GracePeriod: 5 * time.Millisecond,
	}

	sm := NewSessionManager(MakeTestRemoteConfig())
	defer sm.Stop()

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sm.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	hh := NewHeartbeatHandler(sm, config)
	defer hh.Stop()

	hh.RegisterSession(session.ID)

	// Wait for timeout to pass
	time.Sleep(50 * time.Millisecond)

	status := hh.GetHeartbeatStatus(session.ID)
	if status == nil {
		t.Fatal("Status should not be nil")
	}
	if status.IsHealthy {
		t.Error("Session should not be healthy after timeout")
	}
	if status.State != "dead" {
		t.Errorf("Expected state 'dead', got '%s'", status.State)
	}
}

func TestHeartbeatHandler_HeartbeatResetsStatus(t *testing.T) {
	config := &HeartbeatConfig{
		Interval:    10 * time.Millisecond,
		Timeout:     50 * time.Millisecond,
		GracePeriod: 5 * time.Millisecond,
	}

	sm := NewSessionManager(MakeTestRemoteConfig())
	defer sm.Stop()

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sm.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	hh := NewHeartbeatHandler(sm, config)
	defer hh.Stop()

	hh.RegisterSession(session.ID)

	// Wait for stale
	time.Sleep(20 * time.Millisecond)

	status := hh.GetHeartbeatStatus(session.ID)
	if status.State == "healthy" {
		t.Error("Session should be stale")
	}

	// Send heartbeat
	msg := &HeartbeatMessage{
		SessionID: session.ID,
		Timestamp: time.Now(),
	}
	hh.HandleHeartbeat(msg)

	// Check status after heartbeat
	status = hh.GetHeartbeatStatus(session.ID)
	if status.State != "healthy" {
		t.Errorf("Session should be healthy after heartbeat, got %s", status.State)
	}
	if status.MissedCount != 0 {
		t.Error("MissedCount should be reset after heartbeat")
	}
}

func TestHeartbeatHandler_InvalidSession(t *testing.T) {
	sm := NewSessionManager(nil)
	defer sm.Stop()

	hh := NewHeartbeatHandler(sm, MakeTestHeartbeatConfig())
	defer hh.Stop()

	msg := &HeartbeatMessage{
		SessionID: "nonexistent-session",
		Timestamp: time.Now(),
	}

	_, err := hh.HandleHeartbeat(msg)
	if err == nil {
		t.Error("Expected error for invalid session")
	}
}

func TestHeartbeatHandler_MonitoringDetectsDeadSessions(t *testing.T) {
	config := &HeartbeatConfig{
		Interval:    5 * time.Millisecond,
		Timeout:     15 * time.Millisecond,
		GracePeriod: 2 * time.Millisecond,
	}

	sm := NewSessionManager(&SessionConfig{
		HeartbeatInterval:    5 * time.Millisecond,
		HeartbeatTimeout:     15 * time.Millisecond,
		IdleTimeout:          1 * time.Minute,
		IdleWarningBefore:    10 * time.Second,
		ReconnectTokenExpiry: 30 * time.Second,
		CleanupInterval:      5 * time.Millisecond,
	})
	defer sm.Stop()

	claims := &security.Claims{UserID: "user1", Email: "test@example.com"}
	session, err := sm.CreateSession("user1", claims, nil)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	hh := NewHeartbeatHandler(sm, config)
	defer hh.Stop()

	hh.RegisterSession(session.ID)

	// Verify session is active
	s, _ := sm.GetSession(session.ID)
	if s.State != StateActive {
		t.Fatalf("Session should be active, got %v", s.State)
	}

	// Wait for timeout and monitor loop to run
	time.Sleep(25 * time.Millisecond)

	// The monitoring loop should have marked the session as disconnected
	s, _ = sm.GetSession(session.ID)
	if s.State != StateDisconnected {
		t.Errorf("Session should be disconnected after heartbeat timeout, got %v", s.State)
	}
}

func TestHeartbeatHandler_ConcurrentHeartbeats(t *testing.T) {
	sm := NewSessionManager(MakeTestRemoteConfig())
	defer sm.Stop()

	hh := NewHeartbeatHandler(sm, MakeTestHeartbeatConfig())
	defer hh.Stop()

	// Create multiple sessions
	sessions := make([]*RemoteSession, 5)
	for i := 0; i < 5; i++ {
		claims := &security.Claims{UserID: "user" + string(rune('0'+i)), Email: "test@example.com"}
		session, err := sm.CreateSession(claims.UserID, claims, nil)
		if err != nil {
			t.Fatalf("CreateSession failed: %v", err)
		}
		sessions[i] = session
		hh.RegisterSession(session.ID)
	}

	// Send concurrent heartbeats
	runner := NewConcurrentTestRunner()
	for _, session := range sessions {
		s := session
		runner.Run(func() error {
			msg := &HeartbeatMessage{
				SessionID: s.ID,
				Timestamp: time.Now(),
			}
			_, err := hh.HandleHeartbeat(msg)
			return err
		})
	}

	runner.Wait()

	if runner.HasErrors() {
		t.Errorf("Concurrent heartbeats failed: %v", runner.Errors())
	}

	// Verify all sessions are healthy
	for _, session := range sessions {
		status := hh.GetHeartbeatStatus(session.ID)
		if status == nil {
			t.Errorf("Status for session %s should not be nil", session.ID)
			continue
		}
		if !status.IsHealthy {
			t.Errorf("Session %s should be healthy", session.ID)
		}
	}
}
