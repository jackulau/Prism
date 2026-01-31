package repository

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupSessionTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	// Create a temporary database file
	tmpFile, err := os.CreateTemp("", "test-session-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()

	db, err := sql.Open("sqlite3", tmpFile.Name()+"?_foreign_keys=on")
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("failed to open database: %v", err)
	}

	// Create the required tables
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		db.Close()
		os.Remove(tmpFile.Name())
		t.Fatalf("failed to create users table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			refresh_token_hash TEXT NOT NULL,
			device_info TEXT DEFAULT '',
			ip_address TEXT DEFAULT '',
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_used_at DATETIME,
			is_revoked INTEGER DEFAULT 0
		)
	`)
	if err != nil {
		db.Close()
		os.Remove(tmpFile.Name())
		t.Fatalf("failed to create sessions table: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(tmpFile.Name())
	}

	return db, cleanup
}

func createTestUserForSession(t *testing.T, db *sql.DB, id, email string) {
	t.Helper()
	_, err := db.Exec(
		"INSERT INTO users (id, email, password_hash, created_at, updated_at) VALUES (?, ?, 'hash', datetime('now'), datetime('now'))",
		id, email,
	)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
}

func TestSessionRepository_CreateSession(t *testing.T) {
	db, cleanup := setupSessionTestDB(t)
	defer cleanup()

	repo := NewSessionRepository(db)
	createTestUserForSession(t, db, "user1", "test@example.com")

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	session, err := repo.CreateSession("user1", "token_hash_123", "Mozilla/5.0", "192.168.1.1", expiresAt)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	if session.ID == "" {
		t.Error("expected session ID to be set")
	}
	if session.UserID != "user1" {
		t.Errorf("expected user ID to be 'user1', got '%s'", session.UserID)
	}
	if session.RefreshTokenHash != "token_hash_123" {
		t.Errorf("expected refresh token hash to be 'token_hash_123', got '%s'", session.RefreshTokenHash)
	}
	if session.DeviceInfo != "Mozilla/5.0" {
		t.Errorf("expected device info to be 'Mozilla/5.0', got '%s'", session.DeviceInfo)
	}
	if session.IPAddress != "192.168.1.1" {
		t.Errorf("expected IP address to be '192.168.1.1', got '%s'", session.IPAddress)
	}
	if session.IsRevoked {
		t.Error("expected session to not be revoked")
	}
}

func TestSessionRepository_GetByRefreshTokenHash(t *testing.T) {
	db, cleanup := setupSessionTestDB(t)
	defer cleanup()

	repo := NewSessionRepository(db)
	createTestUserForSession(t, db, "user1", "test@example.com")

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	created, err := repo.CreateSession("user1", "token_hash_456", "Chrome", "10.0.0.1", expiresAt)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Get by hash
	session, err := repo.GetByRefreshTokenHash("token_hash_456")
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}
	if session == nil {
		t.Fatal("expected session to be found")
	}
	if session.ID != created.ID {
		t.Errorf("expected session ID to be '%s', got '%s'", created.ID, session.ID)
	}

	// Get non-existent hash
	notFound, err := repo.GetByRefreshTokenHash("nonexistent_hash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if notFound != nil {
		t.Error("expected no session to be found")
	}
}

func TestSessionRepository_GetByRefreshTokenHash_RevokedSession(t *testing.T) {
	db, cleanup := setupSessionTestDB(t)
	defer cleanup()

	repo := NewSessionRepository(db)
	createTestUserForSession(t, db, "user1", "test@example.com")

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	session, err := repo.CreateSession("user1", "token_hash_789", "Firefox", "172.16.0.1", expiresAt)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Revoke the session
	err = repo.RevokeSession(session.ID)
	if err != nil {
		t.Fatalf("failed to revoke session: %v", err)
	}

	// Should not find revoked session
	found, err := repo.GetByRefreshTokenHash("token_hash_789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Error("expected revoked session to not be found")
	}
}

func TestSessionRepository_GetUserSessions(t *testing.T) {
	db, cleanup := setupSessionTestDB(t)
	defer cleanup()

	repo := NewSessionRepository(db)
	createTestUserForSession(t, db, "user1", "test@example.com")
	createTestUserForSession(t, db, "user2", "test2@example.com")

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Create sessions for user1
	_, err := repo.CreateSession("user1", "hash1", "Device1", "1.1.1.1", expiresAt)
	if err != nil {
		t.Fatalf("failed to create session 1: %v", err)
	}
	_, err = repo.CreateSession("user1", "hash2", "Device2", "2.2.2.2", expiresAt)
	if err != nil {
		t.Fatalf("failed to create session 2: %v", err)
	}

	// Create session for user2
	_, err = repo.CreateSession("user2", "hash3", "Device3", "3.3.3.3", expiresAt)
	if err != nil {
		t.Fatalf("failed to create session 3: %v", err)
	}

	// Get user1's sessions
	sessions, err := repo.GetUserSessions("user1")
	if err != nil {
		t.Fatalf("failed to get user sessions: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(sessions))
	}

	// Get user2's sessions
	sessions2, err := repo.GetUserSessions("user2")
	if err != nil {
		t.Fatalf("failed to get user2 sessions: %v", err)
	}
	if len(sessions2) != 1 {
		t.Errorf("expected 1 session, got %d", len(sessions2))
	}
}

func TestSessionRepository_UpdateRefreshTokenHash(t *testing.T) {
	db, cleanup := setupSessionTestDB(t)
	defer cleanup()

	repo := NewSessionRepository(db)
	createTestUserForSession(t, db, "user1", "test@example.com")

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	session, err := repo.CreateSession("user1", "old_hash", "Device", "1.1.1.1", expiresAt)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Update the refresh token hash
	newExpiresAt := time.Now().Add(14 * 24 * time.Hour)
	err = repo.UpdateRefreshTokenHash(session.ID, "new_hash", newExpiresAt)
	if err != nil {
		t.Fatalf("failed to update refresh token hash: %v", err)
	}

	// Old hash should not find session
	oldSession, err := repo.GetByRefreshTokenHash("old_hash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if oldSession != nil {
		t.Error("expected old hash to not find session")
	}

	// New hash should find session
	newSession, err := repo.GetByRefreshTokenHash("new_hash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newSession == nil {
		t.Fatal("expected new hash to find session")
	}
	if newSession.ID != session.ID {
		t.Error("expected same session ID after update")
	}
}

func TestSessionRepository_RevokeSession(t *testing.T) {
	db, cleanup := setupSessionTestDB(t)
	defer cleanup()

	repo := NewSessionRepository(db)
	createTestUserForSession(t, db, "user1", "test@example.com")

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	session, err := repo.CreateSession("user1", "hash", "Device", "1.1.1.1", expiresAt)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Revoke the session
	err = repo.RevokeSession(session.ID)
	if err != nil {
		t.Fatalf("failed to revoke session: %v", err)
	}

	// Get by ID should still return it (but marked as revoked)
	revokedSession, err := repo.GetSessionByID(session.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if revokedSession == nil {
		t.Fatal("expected session to exist")
	}
	if !revokedSession.IsRevoked {
		t.Error("expected session to be marked as revoked")
	}
}

func TestSessionRepository_RevokeAllUserSessions(t *testing.T) {
	db, cleanup := setupSessionTestDB(t)
	defer cleanup()

	repo := NewSessionRepository(db)
	createTestUserForSession(t, db, "user1", "test@example.com")

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Create multiple sessions
	_, err := repo.CreateSession("user1", "hash1", "Device1", "1.1.1.1", expiresAt)
	if err != nil {
		t.Fatalf("failed to create session 1: %v", err)
	}
	_, err = repo.CreateSession("user1", "hash2", "Device2", "2.2.2.2", expiresAt)
	if err != nil {
		t.Fatalf("failed to create session 2: %v", err)
	}

	// Revoke all
	err = repo.RevokeAllUserSessions("user1")
	if err != nil {
		t.Fatalf("failed to revoke all sessions: %v", err)
	}

	// Should have no active sessions
	sessions, err := repo.GetUserSessions("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 active sessions, got %d", len(sessions))
	}
}

func TestSessionRepository_RevokeOtherSessions(t *testing.T) {
	db, cleanup := setupSessionTestDB(t)
	defer cleanup()

	repo := NewSessionRepository(db)
	createTestUserForSession(t, db, "user1", "test@example.com")

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Create multiple sessions
	currentSession, err := repo.CreateSession("user1", "hash1", "Device1", "1.1.1.1", expiresAt)
	if err != nil {
		t.Fatalf("failed to create session 1: %v", err)
	}
	_, err = repo.CreateSession("user1", "hash2", "Device2", "2.2.2.2", expiresAt)
	if err != nil {
		t.Fatalf("failed to create session 2: %v", err)
	}
	_, err = repo.CreateSession("user1", "hash3", "Device3", "3.3.3.3", expiresAt)
	if err != nil {
		t.Fatalf("failed to create session 3: %v", err)
	}

	// Revoke other sessions (keep current)
	err = repo.RevokeOtherSessions("user1", currentSession.ID)
	if err != nil {
		t.Fatalf("failed to revoke other sessions: %v", err)
	}

	// Should have only current session active
	sessions, err := repo.GetUserSessions("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 1 {
		t.Errorf("expected 1 active session, got %d", len(sessions))
	}
	if sessions[0].ID != currentSession.ID {
		t.Error("expected current session to remain active")
	}
}

func TestSessionRepository_CountUserSessions(t *testing.T) {
	db, cleanup := setupSessionTestDB(t)
	defer cleanup()

	repo := NewSessionRepository(db)
	createTestUserForSession(t, db, "user1", "test@example.com")

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Initially no sessions
	count, err := repo.CountUserSessions("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 sessions, got %d", count)
	}

	// Create sessions
	_, err = repo.CreateSession("user1", "hash1", "Device1", "1.1.1.1", expiresAt)
	if err != nil {
		t.Fatalf("failed to create session 1: %v", err)
	}
	_, err = repo.CreateSession("user1", "hash2", "Device2", "2.2.2.2", expiresAt)
	if err != nil {
		t.Fatalf("failed to create session 2: %v", err)
	}

	count, err = repo.CountUserSessions("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 sessions, got %d", count)
	}
}

func TestSessionRepository_UpdateLastUsed(t *testing.T) {
	db, cleanup := setupSessionTestDB(t)
	defer cleanup()

	repo := NewSessionRepository(db)
	createTestUserForSession(t, db, "user1", "test@example.com")

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	session, err := repo.CreateSession("user1", "hash", "Device", "1.1.1.1", expiresAt)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	originalLastUsed := session.LastUsedAt

	// Wait a bit and update
	time.Sleep(10 * time.Millisecond)

	err = repo.UpdateLastUsed(session.ID)
	if err != nil {
		t.Fatalf("failed to update last used: %v", err)
	}

	// Get updated session
	updated, err := repo.GetSessionByID(session.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated == nil {
		t.Fatal("expected session to exist")
	}
	if !updated.LastUsedAt.After(originalLastUsed) {
		t.Error("expected last_used_at to be updated")
	}
}

func TestSessionRepository_CleanupExpiredSessions(t *testing.T) {
	db, cleanup := setupSessionTestDB(t)
	defer cleanup()

	repo := NewSessionRepository(db)
	createTestUserForSession(t, db, "user1", "test@example.com")

	// Create expired session
	expiredAt := time.Now().Add(-1 * time.Hour)
	_, err := repo.CreateSession("user1", "expired_hash", "Device1", "1.1.1.1", expiredAt)
	if err != nil {
		t.Fatalf("failed to create expired session: %v", err)
	}

	// Create valid session
	validAt := time.Now().Add(7 * 24 * time.Hour)
	validSession, err := repo.CreateSession("user1", "valid_hash", "Device2", "2.2.2.2", validAt)
	if err != nil {
		t.Fatalf("failed to create valid session: %v", err)
	}

	// Create revoked session
	revokedSession, err := repo.CreateSession("user1", "revoked_hash", "Device3", "3.3.3.3", validAt)
	if err != nil {
		t.Fatalf("failed to create revoked session: %v", err)
	}
	_ = repo.RevokeSession(revokedSession.ID)

	// Cleanup
	err = repo.CleanupExpiredSessions()
	if err != nil {
		t.Fatalf("failed to cleanup sessions: %v", err)
	}

	// Valid session should remain
	found, err := repo.GetSessionByID(validSession.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil {
		t.Error("expected valid session to remain after cleanup")
	}

	// Expired session should be deleted
	expiredFound, err := repo.GetByRefreshTokenHash("expired_hash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if expiredFound != nil {
		t.Error("expected expired session to be deleted")
	}

	// Revoked session should be deleted
	revokedFound, err := repo.GetSessionByID(revokedSession.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if revokedFound != nil {
		t.Error("expected revoked session to be deleted after cleanup")
	}
}

func TestSessionRepository_BackwardCompatibility(t *testing.T) {
	db, cleanup := setupSessionTestDB(t)
	defer cleanup()

	repo := NewSessionRepository(db)
	createTestUserForSession(t, db, "user1", "test@example.com")

	// Test backward compatible Create method
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	session, err := repo.Create("user1", "hash", expiresAt)
	if err != nil {
		t.Fatalf("failed to create session with backward compatible method: %v", err)
	}

	if session.DeviceInfo != "" {
		t.Errorf("expected empty device info, got '%s'", session.DeviceInfo)
	}
	if session.IPAddress != "" {
		t.Errorf("expected empty IP address, got '%s'", session.IPAddress)
	}
}
