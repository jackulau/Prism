package remote

import (
	"testing"
	"time"
)

func TestTunnelConnection_ByteTracking(t *testing.T) {
	conn := NewTunnelConnection("192.168.1.1", nil)

	conn.AddBytesIn(100)
	conn.AddBytesIn(50)
	conn.AddBytesOut(200)

	if got := conn.GetBytesIn(); got != 150 {
		t.Errorf("GetBytesIn() = %d, want %d", got, 150)
	}

	if got := conn.GetBytesOut(); got != 200 {
		t.Errorf("GetBytesOut() = %d, want %d", got, 200)
	}
}

func TestTunnelConnection_Close(t *testing.T) {
	conn := NewTunnelConnection("192.168.1.1", nil)

	if conn.IsClosed() {
		t.Error("IsClosed() = true before Close()")
	}

	conn.Close()

	if !conn.IsClosed() {
		t.Error("IsClosed() = false after Close()")
	}

	// Second close should be safe
	conn.Close()

	if !conn.IsClosed() {
		t.Error("IsClosed() = false after second Close()")
	}
}

func TestTunnelConnection_LastSeen(t *testing.T) {
	conn := NewTunnelConnection("192.168.1.1", nil)
	initialLastSeen := conn.LastSeen

	time.Sleep(10 * time.Millisecond)
	conn.UpdateLastSeen()

	if !conn.LastSeen.After(initialLastSeen) {
		t.Error("UpdateLastSeen() did not update LastSeen")
	}
}

func TestConnectionManager_Add(t *testing.T) {
	manager := NewConnectionManager(10, 5)

	conn := NewTunnelConnection("192.168.1.1", nil)
	err := manager.Add(conn)
	if err != nil {
		t.Errorf("Add() error = %v", err)
	}

	if got := manager.Count(); got != 1 {
		t.Errorf("Count() = %d, want %d", got, 1)
	}

	if got := manager.CountByIP("192.168.1.1"); got != 1 {
		t.Errorf("CountByIP() = %d, want %d", got, 1)
	}
}

func TestConnectionManager_Remove(t *testing.T) {
	manager := NewConnectionManager(10, 5)

	conn := NewTunnelConnection("192.168.1.1", nil)
	_ = manager.Add(conn)

	manager.Remove(conn.ID)

	if got := manager.Count(); got != 0 {
		t.Errorf("Count() after Remove() = %d, want %d", got, 0)
	}

	if got := manager.CountByIP("192.168.1.1"); got != 0 {
		t.Errorf("CountByIP() after Remove() = %d, want %d", got, 0)
	}
}

func TestConnectionManager_GlobalLimit(t *testing.T) {
	manager := NewConnectionManager(2, 5) // Global limit of 2

	conn1 := NewTunnelConnection("192.168.1.1", nil)
	conn2 := NewTunnelConnection("192.168.1.2", nil)
	conn3 := NewTunnelConnection("192.168.1.3", nil)

	if err := manager.Add(conn1); err != nil {
		t.Errorf("Add() first connection error = %v", err)
	}

	if err := manager.Add(conn2); err != nil {
		t.Errorf("Add() second connection error = %v", err)
	}

	if err := manager.Add(conn3); err != ErrConnectionLimitExceeded {
		t.Errorf("Add() third connection error = %v, want %v", err, ErrConnectionLimitExceeded)
	}
}

func TestConnectionManager_PerIPLimit(t *testing.T) {
	manager := NewConnectionManager(10, 2) // Per-IP limit of 2

	conn1 := NewTunnelConnection("192.168.1.1", nil)
	conn2 := NewTunnelConnection("192.168.1.1", nil)
	conn3 := NewTunnelConnection("192.168.1.1", nil)

	if err := manager.Add(conn1); err != nil {
		t.Errorf("Add() first connection error = %v", err)
	}

	if err := manager.Add(conn2); err != nil {
		t.Errorf("Add() second connection error = %v", err)
	}

	if err := manager.Add(conn3); err != ErrConnectionLimitPerIPExceeded {
		t.Errorf("Add() third connection from same IP error = %v, want %v", err, ErrConnectionLimitPerIPExceeded)
	}

	// Different IP should work
	conn4 := NewTunnelConnection("192.168.1.2", nil)
	if err := manager.Add(conn4); err != nil {
		t.Errorf("Add() connection from different IP error = %v", err)
	}
}

func TestConnectionManager_Get(t *testing.T) {
	manager := NewConnectionManager(10, 5)

	conn := NewTunnelConnection("192.168.1.1", nil)
	_ = manager.Add(conn)

	got := manager.Get(conn.ID)
	if got == nil {
		t.Error("Get() returned nil")
		return
	}
	if got.ID != conn.ID {
		t.Errorf("Get().ID = %v, want %v", got.ID, conn.ID)
	}

	// Non-existent connection
	got = manager.Get("non-existent-id")
	if got != nil {
		t.Error("Get() for non-existent ID should return nil")
	}
}

func TestConnectionManager_GetAll(t *testing.T) {
	manager := NewConnectionManager(10, 5)

	conn1 := NewTunnelConnection("192.168.1.1", nil)
	conn2 := NewTunnelConnection("192.168.1.2", nil)
	_ = manager.Add(conn1)
	_ = manager.Add(conn2)

	all := manager.GetAll()
	if len(all) != 2 {
		t.Errorf("GetAll() returned %d connections, want %d", len(all), 2)
	}
}

func TestConnectionManager_GetByIP(t *testing.T) {
	manager := NewConnectionManager(10, 5)

	conn1 := NewTunnelConnection("192.168.1.1", nil)
	conn2 := NewTunnelConnection("192.168.1.1", nil)
	conn3 := NewTunnelConnection("192.168.1.2", nil)
	_ = manager.Add(conn1)
	_ = manager.Add(conn2)
	_ = manager.Add(conn3)

	ip1Conns := manager.GetByIP("192.168.1.1")
	if len(ip1Conns) != 2 {
		t.Errorf("GetByIP(192.168.1.1) returned %d connections, want %d", len(ip1Conns), 2)
	}

	ip2Conns := manager.GetByIP("192.168.1.2")
	if len(ip2Conns) != 1 {
		t.Errorf("GetByIP(192.168.1.2) returned %d connections, want %d", len(ip2Conns), 1)
	}
}

func TestConnectionManager_CloseAll(t *testing.T) {
	manager := NewConnectionManager(10, 5)

	conn1 := NewTunnelConnection("192.168.1.1", nil)
	conn2 := NewTunnelConnection("192.168.1.2", nil)
	_ = manager.Add(conn1)
	_ = manager.Add(conn2)

	manager.CloseAll()

	if got := manager.Count(); got != 0 {
		t.Errorf("Count() after CloseAll() = %d, want %d", got, 0)
	}

	if !conn1.IsClosed() {
		t.Error("conn1 should be closed after CloseAll()")
	}

	if !conn2.IsClosed() {
		t.Error("conn2 should be closed after CloseAll()")
	}
}

func TestConnectionManager_CloseBySession(t *testing.T) {
	manager := NewConnectionManager(10, 5)

	session1 := &RemoteSession{ID: "session-1"}
	session2 := &RemoteSession{ID: "session-2"}

	conn1 := NewTunnelConnection("192.168.1.1", session1)
	conn2 := NewTunnelConnection("192.168.1.2", session1)
	conn3 := NewTunnelConnection("192.168.1.3", session2)

	_ = manager.Add(conn1)
	_ = manager.Add(conn2)
	_ = manager.Add(conn3)

	manager.CloseBySession("session-1")

	if got := manager.Count(); got != 1 {
		t.Errorf("Count() after CloseBySession() = %d, want %d", got, 1)
	}

	if !conn1.IsClosed() {
		t.Error("conn1 should be closed after CloseBySession(session-1)")
	}

	if !conn2.IsClosed() {
		t.Error("conn2 should be closed after CloseBySession(session-1)")
	}

	if conn3.IsClosed() {
		t.Error("conn3 should not be closed after CloseBySession(session-1)")
	}
}

func TestConnectionManager_CanAccept(t *testing.T) {
	manager := NewConnectionManager(2, 1) // Global: 2, Per-IP: 1

	// Empty manager should accept
	if !manager.CanAccept("192.168.1.1") {
		t.Error("CanAccept() = false for empty manager")
	}

	// Add one connection from IP1
	conn1 := NewTunnelConnection("192.168.1.1", nil)
	_ = manager.Add(conn1)

	// Same IP should be rejected (per-IP limit)
	if manager.CanAccept("192.168.1.1") {
		t.Error("CanAccept() = true when per-IP limit reached")
	}

	// Different IP should be accepted
	if !manager.CanAccept("192.168.1.2") {
		t.Error("CanAccept() = false for different IP")
	}

	// Add connection from IP2
	conn2 := NewTunnelConnection("192.168.1.2", nil)
	_ = manager.Add(conn2)

	// Global limit reached, new IP should be rejected
	if manager.CanAccept("192.168.1.3") {
		t.Error("CanAccept() = true when global limit reached")
	}
}
