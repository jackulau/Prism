package remote

import (
	"sync"
	"time"
)

// HeartbeatHandler manages heartbeat monitoring for remote sessions
type HeartbeatHandler struct {
	sessionManager *SessionManager
	config         *HeartbeatConfig
	monitors       map[string]*heartbeatMonitor
	mu             sync.RWMutex
	stopChan       chan struct{}
}

// HeartbeatConfig holds configuration for heartbeat handling
type HeartbeatConfig struct {
	// Interval at which clients should send heartbeats
	Interval time.Duration
	// Timeout after which a connection is considered dead
	Timeout time.Duration
	// Grace period before marking as disconnected
	GracePeriod time.Duration
}

// DefaultHeartbeatConfig returns the default heartbeat configuration
func DefaultHeartbeatConfig() *HeartbeatConfig {
	return &HeartbeatConfig{
		Interval:    30 * time.Second,
		Timeout:     90 * time.Second,
		GracePeriod: 15 * time.Second,
	}
}

// heartbeatMonitor tracks heartbeat state for a single session
type heartbeatMonitor struct {
	sessionID     string
	lastHeartbeat time.Time
	missedCount   int
	mu            sync.Mutex
}

// HeartbeatMessage represents a heartbeat message
type HeartbeatMessage struct {
	SessionID string    `json:"session_id"`
	Timestamp time.Time `json:"timestamp"`
	Sequence  int64     `json:"sequence,omitempty"`
}

// HeartbeatAck represents a heartbeat acknowledgment
type HeartbeatAck struct {
	SessionID       string    `json:"session_id"`
	Timestamp       time.Time `json:"timestamp"`
	ServerTime      time.Time `json:"server_time"`
	NextHeartbeatIn int64     `json:"next_heartbeat_in_ms"`
}

// NewHeartbeatHandler creates a new heartbeat handler
func NewHeartbeatHandler(sessionManager *SessionManager, config *HeartbeatConfig) *HeartbeatHandler {
	if config == nil {
		config = DefaultHeartbeatConfig()
	}

	hh := &HeartbeatHandler{
		sessionManager: sessionManager,
		config:         config,
		monitors:       make(map[string]*heartbeatMonitor),
		stopChan:       make(chan struct{}),
	}

	// Start monitoring loop
	go hh.monitorLoop()

	return hh
}

// RegisterSession registers a session for heartbeat monitoring
func (hh *HeartbeatHandler) RegisterSession(sessionID string) {
	hh.mu.Lock()
	defer hh.mu.Unlock()

	hh.monitors[sessionID] = &heartbeatMonitor{
		sessionID:     sessionID,
		lastHeartbeat: time.Now(),
		missedCount:   0,
	}
}

// UnregisterSession removes a session from heartbeat monitoring
func (hh *HeartbeatHandler) UnregisterSession(sessionID string) {
	hh.mu.Lock()
	defer hh.mu.Unlock()

	delete(hh.monitors, sessionID)
}

// HandleHeartbeat processes an incoming heartbeat and returns an ack
func (hh *HeartbeatHandler) HandleHeartbeat(msg *HeartbeatMessage) (*HeartbeatAck, error) {
	// Update session manager
	if err := hh.sessionManager.UpdateHeartbeat(msg.SessionID); err != nil {
		return nil, err
	}

	// Update monitor
	hh.mu.RLock()
	monitor, exists := hh.monitors[msg.SessionID]
	hh.mu.RUnlock()

	if exists {
		monitor.mu.Lock()
		monitor.lastHeartbeat = time.Now()
		monitor.missedCount = 0
		monitor.mu.Unlock()
	}

	// Generate acknowledgment
	now := time.Now()
	ack := &HeartbeatAck{
		SessionID:       msg.SessionID,
		Timestamp:       msg.Timestamp,
		ServerTime:      now,
		NextHeartbeatIn: hh.config.Interval.Milliseconds(),
	}

	return ack, nil
}

// GetHeartbeatStatus returns the heartbeat status for a session
func (hh *HeartbeatHandler) GetHeartbeatStatus(sessionID string) *HeartbeatStatus {
	hh.mu.RLock()
	monitor, exists := hh.monitors[sessionID]
	hh.mu.RUnlock()

	if !exists {
		return nil
	}

	monitor.mu.Lock()
	defer monitor.mu.Unlock()

	now := time.Now()
	timeSinceLastHeartbeat := now.Sub(monitor.lastHeartbeat)

	status := &HeartbeatStatus{
		SessionID:     sessionID,
		LastHeartbeat: monitor.lastHeartbeat,
		MissedCount:   monitor.missedCount,
		IsHealthy:     timeSinceLastHeartbeat < hh.config.Timeout,
		LatencyMs:     timeSinceLastHeartbeat.Milliseconds(),
	}

	if timeSinceLastHeartbeat >= hh.config.Timeout {
		status.State = "dead"
	} else if timeSinceLastHeartbeat >= hh.config.Interval+hh.config.GracePeriod {
		status.State = "stale"
	} else {
		status.State = "healthy"
	}

	return status
}

// HeartbeatStatus represents the heartbeat status of a session
type HeartbeatStatus struct {
	SessionID     string    `json:"session_id"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	MissedCount   int       `json:"missed_count"`
	IsHealthy     bool      `json:"is_healthy"`
	State         string    `json:"state"` // healthy, stale, dead
	LatencyMs     int64     `json:"latency_ms"`
}

// monitorLoop continuously monitors heartbeats
func (hh *HeartbeatHandler) monitorLoop() {
	ticker := time.NewTicker(hh.config.Interval / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hh.checkHeartbeats()
		case <-hh.stopChan:
			return
		}
	}
}

// checkHeartbeats checks all monitored sessions for heartbeat health
func (hh *HeartbeatHandler) checkHeartbeats() {
	now := time.Now()
	staleSessionsIDs := make([]string, 0)
	deadSessionIDs := make([]string, 0)

	hh.mu.RLock()
	for sessionID, monitor := range hh.monitors {
		monitor.mu.Lock()
		timeSinceLastHeartbeat := now.Sub(monitor.lastHeartbeat)

		if timeSinceLastHeartbeat >= hh.config.Timeout {
			deadSessionIDs = append(deadSessionIDs, sessionID)
		} else if timeSinceLastHeartbeat >= hh.config.Interval+hh.config.GracePeriod {
			monitor.missedCount++
			staleSessionsIDs = append(staleSessionsIDs, sessionID)
		}
		monitor.mu.Unlock()
	}
	hh.mu.RUnlock()

	// Handle dead sessions
	for _, sessionID := range deadSessionIDs {
		hh.sessionManager.MarkDisconnected(sessionID)
	}
}

// GetConfig returns the heartbeat configuration
func (hh *HeartbeatHandler) GetConfig() *HeartbeatConfig {
	return hh.config
}

// Stop stops the heartbeat handler
func (hh *HeartbeatHandler) Stop() {
	close(hh.stopChan)
}
