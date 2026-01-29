package remote

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// TunnelConnection represents an active tunnel connection
type TunnelConnection struct {
	ID        string          `json:"id"`
	ClientIP  string          `json:"client_ip"`
	Session   *RemoteSession  `json:"session"`
	CreatedAt time.Time       `json:"created_at"`
	LastSeen  time.Time       `json:"last_seen"`
	BytesIn   atomic.Int64    `json:"-"`
	BytesOut  atomic.Int64    `json:"-"`
	closeCh   chan struct{}
	closed    atomic.Bool
}

// NewTunnelConnection creates a new tunnel connection
func NewTunnelConnection(clientIP string, session *RemoteSession) *TunnelConnection {
	return &TunnelConnection{
		ID:        uuid.New().String(),
		ClientIP:  clientIP,
		Session:   session,
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
		closeCh:   make(chan struct{}),
	}
}

// UpdateLastSeen updates the last seen timestamp
func (c *TunnelConnection) UpdateLastSeen() {
	c.LastSeen = time.Now()
}

// AddBytesIn adds to the bytes received counter
func (c *TunnelConnection) AddBytesIn(n int64) {
	c.BytesIn.Add(n)
}

// AddBytesOut adds to the bytes sent counter
func (c *TunnelConnection) AddBytesOut(n int64) {
	c.BytesOut.Add(n)
}

// Close marks the connection as closed
func (c *TunnelConnection) Close() {
	if c.closed.CompareAndSwap(false, true) {
		close(c.closeCh)
	}
}

// IsClosed returns whether the connection is closed
func (c *TunnelConnection) IsClosed() bool {
	return c.closed.Load()
}

// Done returns a channel that's closed when the connection is closed
func (c *TunnelConnection) Done() <-chan struct{} {
	return c.closeCh
}

// GetBytesIn returns the total bytes received
func (c *TunnelConnection) GetBytesIn() int64 {
	return c.BytesIn.Load()
}

// GetBytesOut returns the total bytes sent
func (c *TunnelConnection) GetBytesOut() int64 {
	return c.BytesOut.Load()
}

// ConnectionManager manages active tunnel connections
type ConnectionManager struct {
	connections    map[string]*TunnelConnection
	byIP           map[string]map[string]*TunnelConnection
	maxConnections int
	maxPerIP       int
	mu             sync.RWMutex
	startTime      time.Time
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(maxConnections, maxPerIP int) *ConnectionManager {
	return &ConnectionManager{
		connections:    make(map[string]*TunnelConnection),
		byIP:           make(map[string]map[string]*TunnelConnection),
		maxConnections: maxConnections,
		maxPerIP:       maxPerIP,
		startTime:      time.Now(),
	}
}

// Add adds a new connection if limits allow
func (m *ConnectionManager) Add(conn *TunnelConnection) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check global limit
	if len(m.connections) >= m.maxConnections {
		return ErrConnectionLimitExceeded
	}

	// Check per-IP limit
	ipConns := m.byIP[conn.ClientIP]
	if len(ipConns) >= m.maxPerIP {
		return ErrConnectionLimitPerIPExceeded
	}

	// Add to maps
	m.connections[conn.ID] = conn
	if m.byIP[conn.ClientIP] == nil {
		m.byIP[conn.ClientIP] = make(map[string]*TunnelConnection)
	}
	m.byIP[conn.ClientIP][conn.ID] = conn

	return nil
}

// Remove removes a connection
func (m *ConnectionManager) Remove(connID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, ok := m.connections[connID]
	if !ok {
		return
	}

	delete(m.connections, connID)
	if ipConns, ok := m.byIP[conn.ClientIP]; ok {
		delete(ipConns, connID)
		if len(ipConns) == 0 {
			delete(m.byIP, conn.ClientIP)
		}
	}
}

// Get returns a connection by ID
func (m *ConnectionManager) Get(connID string) *TunnelConnection {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connections[connID]
}

// GetAll returns all active connections
func (m *ConnectionManager) GetAll() []*TunnelConnection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conns := make([]*TunnelConnection, 0, len(m.connections))
	for _, conn := range m.connections {
		conns = append(conns, conn)
	}
	return conns
}

// GetByIP returns all connections for a given IP
func (m *ConnectionManager) GetByIP(ip string) []*TunnelConnection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ipConns := m.byIP[ip]
	conns := make([]*TunnelConnection, 0, len(ipConns))
	for _, conn := range ipConns {
		conns = append(conns, conn)
	}
	return conns
}

// Count returns the total number of active connections
func (m *ConnectionManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.connections)
}

// CountByIP returns the number of connections for a given IP
func (m *ConnectionManager) CountByIP(ip string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.byIP[ip])
}

// CloseAll closes all active connections
func (m *ConnectionManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, conn := range m.connections {
		conn.Close()
	}
	m.connections = make(map[string]*TunnelConnection)
	m.byIP = make(map[string]map[string]*TunnelConnection)
}

// CloseBySession closes all connections for a given session
func (m *ConnectionManager) CloseBySession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	toRemove := make([]string, 0)
	for id, conn := range m.connections {
		if conn.Session != nil && conn.Session.ID == sessionID {
			conn.Close()
			toRemove = append(toRemove, id)
		}
	}

	for _, id := range toRemove {
		if conn, ok := m.connections[id]; ok {
			delete(m.connections, id)
			if ipConns, ok := m.byIP[conn.ClientIP]; ok {
				delete(ipConns, id)
				if len(ipConns) == 0 {
					delete(m.byIP, conn.ClientIP)
				}
			}
		}
	}
}

// CanAccept checks if a new connection from the given IP can be accepted
func (m *ConnectionManager) CanAccept(ip string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.connections) >= m.maxConnections {
		return false
	}
	if len(m.byIP[ip]) >= m.maxPerIP {
		return false
	}
	return true
}
