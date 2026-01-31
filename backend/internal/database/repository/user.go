package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the database
type User struct {
	ID                string
	Email             string
	PasswordHash      string
	GitHubToken       string
	GitHubUsername    string
	GitHubConnectedAt *time.Time
	// WorkOS SSO fields
	WorkOSID        string
	OrganizationID  string
	SSOConnectionID string
	SSOProvider     string // "saml", "oidc", etc.
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// UserRepository handles user database operations
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(email, passwordHash string) (*User, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := r.db.Exec(
		`INSERT INTO users (id, email, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		id, email, passwordHash, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &User{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id string) (*User, error) {
	user := &User{}
	var githubToken, githubUsername sql.NullString
	var githubConnectedAt sql.NullTime
	var workosID, organizationID, ssoConnectionID, ssoProvider sql.NullString

	err := r.db.QueryRow(
		`SELECT id, email, password_hash, github_token, github_username, github_connected_at,
		 workos_id, organization_id, sso_connection_id, sso_provider, created_at, updated_at
		 FROM users WHERE id = ?`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &githubToken, &githubUsername, &githubConnectedAt,
		&workosID, &organizationID, &ssoConnectionID, &ssoProvider, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.GitHubToken = githubToken.String
	user.GitHubUsername = githubUsername.String
	if githubConnectedAt.Valid {
		user.GitHubConnectedAt = &githubConnectedAt.Time
	}
	user.WorkOSID = workosID.String
	user.OrganizationID = organizationID.String
	user.SSOConnectionID = ssoConnectionID.String
	user.SSOProvider = ssoProvider.String

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(email string) (*User, error) {
	user := &User{}
	var githubToken, githubUsername sql.NullString
	var githubConnectedAt sql.NullTime
	var workosID, organizationID, ssoConnectionID, ssoProvider sql.NullString

	err := r.db.QueryRow(
		`SELECT id, email, password_hash, github_token, github_username, github_connected_at,
		 workos_id, organization_id, sso_connection_id, sso_provider, created_at, updated_at
		 FROM users WHERE email = ?`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &githubToken, &githubUsername, &githubConnectedAt,
		&workosID, &organizationID, &ssoConnectionID, &ssoProvider, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.GitHubToken = githubToken.String
	user.GitHubUsername = githubUsername.String
	if githubConnectedAt.Valid {
		user.GitHubConnectedAt = &githubConnectedAt.Time
	}
	user.WorkOSID = workosID.String
	user.OrganizationID = organizationID.String
	user.SSOConnectionID = ssoConnectionID.String
	user.SSOProvider = ssoProvider.String

	return user, nil
}

// EmailExists checks if an email already exists
func (r *UserRepository) EmailExists(email string) (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM users WHERE email = ?`, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check email: %w", err)
	}
	return count > 0, nil
}

// Session represents a user session
type Session struct {
	ID               string
	UserID           string
	RefreshTokenHash string
	DeviceInfo       string
	IPAddress        string
	ExpiresAt        time.Time
	CreatedAt        time.Time
	LastUsedAt       time.Time
	IsRevoked        bool
}

// SessionRepository handles session database operations
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// CreateSession creates a new session with device info and IP address
func (r *SessionRepository) CreateSession(userID, refreshTokenHash, deviceInfo, ipAddress string, expiresAt time.Time) (*Session, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := r.db.Exec(
		`INSERT INTO sessions (id, user_id, refresh_token_hash, device_info, ip_address, expires_at, created_at, last_used_at, is_revoked)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0)`,
		id, userID, refreshTokenHash, deviceInfo, ipAddress, expiresAt, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &Session{
		ID:               id,
		UserID:           userID,
		RefreshTokenHash: refreshTokenHash,
		DeviceInfo:       deviceInfo,
		IPAddress:        ipAddress,
		ExpiresAt:        expiresAt,
		CreatedAt:        now,
		LastUsedAt:       now,
		IsRevoked:        false,
	}, nil
}

// Create creates a new session (backward compatible)
func (r *SessionRepository) Create(userID, refreshTokenHash string, expiresAt time.Time) (*Session, error) {
	return r.CreateSession(userID, refreshTokenHash, "", "", expiresAt)
}

// GetByRefreshTokenHash retrieves a session by refresh token hash
func (r *SessionRepository) GetByRefreshTokenHash(hash string) (*Session, error) {
	session := &Session{}
	var deviceInfo, ipAddress sql.NullString
	var lastUsedAt sql.NullTime
	var isRevoked sql.NullBool

	err := r.db.QueryRow(
		`SELECT id, user_id, refresh_token_hash, device_info, ip_address, expires_at, created_at, last_used_at, is_revoked
		 FROM sessions WHERE refresh_token_hash = ? AND (is_revoked = 0 OR is_revoked IS NULL)`,
		hash,
	).Scan(&session.ID, &session.UserID, &session.RefreshTokenHash, &deviceInfo, &ipAddress,
		&session.ExpiresAt, &session.CreatedAt, &lastUsedAt, &isRevoked)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	session.DeviceInfo = deviceInfo.String
	session.IPAddress = ipAddress.String
	if lastUsedAt.Valid {
		session.LastUsedAt = lastUsedAt.Time
	}
	session.IsRevoked = isRevoked.Valid && isRevoked.Bool

	return session, nil
}

// GetSessionByID retrieves a session by its ID
func (r *SessionRepository) GetSessionByID(id string) (*Session, error) {
	session := &Session{}
	var deviceInfo, ipAddress sql.NullString
	var lastUsedAt sql.NullTime
	var isRevoked sql.NullBool

	err := r.db.QueryRow(
		`SELECT id, user_id, refresh_token_hash, device_info, ip_address, expires_at, created_at, last_used_at, is_revoked
		 FROM sessions WHERE id = ?`,
		id,
	).Scan(&session.ID, &session.UserID, &session.RefreshTokenHash, &deviceInfo, &ipAddress,
		&session.ExpiresAt, &session.CreatedAt, &lastUsedAt, &isRevoked)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	session.DeviceInfo = deviceInfo.String
	session.IPAddress = ipAddress.String
	if lastUsedAt.Valid {
		session.LastUsedAt = lastUsedAt.Time
	}
	session.IsRevoked = isRevoked.Valid && isRevoked.Bool

	return session, nil
}

// GetUserSessions retrieves all active sessions for a user
func (r *SessionRepository) GetUserSessions(userID string) ([]*Session, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, refresh_token_hash, device_info, ip_address, expires_at, created_at, last_used_at, is_revoked
		 FROM sessions WHERE user_id = ? AND (is_revoked = 0 OR is_revoked IS NULL) AND expires_at > ?
		 ORDER BY last_used_at DESC`,
		userID, time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		session := &Session{}
		var deviceInfo, ipAddress sql.NullString
		var lastUsedAt sql.NullTime
		var isRevoked sql.NullBool

		err := rows.Scan(&session.ID, &session.UserID, &session.RefreshTokenHash, &deviceInfo, &ipAddress,
			&session.ExpiresAt, &session.CreatedAt, &lastUsedAt, &isRevoked)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		session.DeviceInfo = deviceInfo.String
		session.IPAddress = ipAddress.String
		if lastUsedAt.Valid {
			session.LastUsedAt = lastUsedAt.Time
		}
		session.IsRevoked = isRevoked.Valid && isRevoked.Bool

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// UpdateLastUsed updates the last_used_at timestamp for a session
func (r *SessionRepository) UpdateLastUsed(sessionID string) error {
	_, err := r.db.Exec(`UPDATE sessions SET last_used_at = ? WHERE id = ?`, time.Now(), sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session last used: %w", err)
	}
	return nil
}

// UpdateRefreshTokenHash updates the refresh token hash for a session (token rotation)
func (r *SessionRepository) UpdateRefreshTokenHash(sessionID, newHash string, newExpiresAt time.Time) error {
	_, err := r.db.Exec(
		`UPDATE sessions SET refresh_token_hash = ?, expires_at = ?, last_used_at = ? WHERE id = ?`,
		newHash, newExpiresAt, time.Now(), sessionID,
	)
	if err != nil {
		return fmt.Errorf("failed to update refresh token: %w", err)
	}
	return nil
}

// RevokeSession revokes a specific session
func (r *SessionRepository) RevokeSession(sessionID string) error {
	_, err := r.db.Exec(`UPDATE sessions SET is_revoked = 1 WHERE id = ?`, sessionID)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}
	return nil
}

// RevokeAllUserSessions revokes all sessions for a user
func (r *SessionRepository) RevokeAllUserSessions(userID string) error {
	_, err := r.db.Exec(`UPDATE sessions SET is_revoked = 1 WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke all sessions: %w", err)
	}
	return nil
}

// RevokeOtherSessions revokes all sessions for a user except the current one
func (r *SessionRepository) RevokeOtherSessions(userID, currentSessionID string) error {
	_, err := r.db.Exec(`UPDATE sessions SET is_revoked = 1 WHERE user_id = ? AND id != ?`, userID, currentSessionID)
	if err != nil {
		return fmt.Errorf("failed to revoke other sessions: %w", err)
	}
	return nil
}

// CleanupExpiredSessions deletes all expired or revoked sessions
func (r *SessionRepository) CleanupExpiredSessions() error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE expires_at < ? OR is_revoked = 1`, time.Now())
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	return nil
}

// CountUserSessions counts active sessions for a user
func (r *SessionRepository) CountUserSessions(userID string) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM sessions WHERE user_id = ? AND (is_revoked = 0 OR is_revoked IS NULL) AND expires_at > ?`,
		userID, time.Now(),
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count sessions: %w", err)
	}
	return count, nil
}

// Delete deletes a session by ID
func (r *SessionRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// DeleteByUserID deletes all sessions for a user
func (r *SessionRepository) DeleteByUserID(userID string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete sessions: %w", err)
	}
	return nil
}

// DeleteExpired deletes all expired sessions
func (r *SessionRepository) DeleteExpired() error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE expires_at < ?`, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	return nil
}

// SaveGitHubConnection saves a GitHub OAuth connection for a user
func (r *UserRepository) SaveGitHubConnection(userID, encryptedToken, username string) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE users SET github_token = ?, github_username = ?, github_connected_at = ?, updated_at = ? WHERE id = ?`,
		encryptedToken, username, now, now, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to save GitHub connection: %w", err)
	}
	return nil
}

// RemoveGitHubConnection removes a GitHub OAuth connection for a user
func (r *UserRepository) RemoveGitHubConnection(userID string) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE users SET github_token = NULL, github_username = NULL, github_connected_at = NULL, updated_at = ? WHERE id = ?`,
		now, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove GitHub connection: %w", err)
	}
	return nil
}

// GetByWorkOSID finds a user by their WorkOS SSO ID
func (r *UserRepository) GetByWorkOSID(workosID string) (*User, error) {
	user := &User{}
	var githubToken, githubUsername sql.NullString
	var githubConnectedAt sql.NullTime
	var workosIDVal, organizationID, ssoConnectionID, ssoProvider sql.NullString

	err := r.db.QueryRow(
		`SELECT id, email, password_hash, github_token, github_username, github_connected_at,
		 workos_id, organization_id, sso_connection_id, sso_provider, created_at, updated_at
		 FROM users WHERE workos_id = ?`,
		workosID,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &githubToken, &githubUsername, &githubConnectedAt,
		&workosIDVal, &organizationID, &ssoConnectionID, &ssoProvider, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by WorkOS ID: %w", err)
	}

	user.GitHubToken = githubToken.String
	user.GitHubUsername = githubUsername.String
	if githubConnectedAt.Valid {
		user.GitHubConnectedAt = &githubConnectedAt.Time
	}
	user.WorkOSID = workosIDVal.String
	user.OrganizationID = organizationID.String
	user.SSOConnectionID = ssoConnectionID.String
	user.SSOProvider = ssoProvider.String

	return user, nil
}

// CreateFromSSO creates a new user from SSO profile
func (r *UserRepository) CreateFromSSO(email, workosID, orgID, connectionID, provider string) (*User, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := r.db.Exec(
		`INSERT INTO users (id, email, password_hash, workos_id, organization_id, sso_connection_id, sso_provider, created_at, updated_at)
		 VALUES (?, ?, '', ?, ?, ?, ?, ?, ?)`,
		id, email, workosID, orgID, connectionID, provider, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user from SSO: %w", err)
	}

	return &User{
		ID:              id,
		Email:           email,
		WorkOSID:        workosID,
		OrganizationID:  orgID,
		SSOConnectionID: connectionID,
		SSOProvider:     provider,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// LinkWorkOSAccount links an existing user to WorkOS
func (r *UserRepository) LinkWorkOSAccount(userID, workosID, orgID, connectionID, provider string) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE users SET workos_id = ?, organization_id = ?, sso_connection_id = ?, sso_provider = ?, updated_at = ? WHERE id = ?`,
		workosID, orgID, connectionID, provider, now, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to link WorkOS account: %w", err)
	}
	return nil
}

// GetByOrganization returns all users in an organization
func (r *UserRepository) GetByOrganization(orgID string) ([]*User, error) {
	rows, err := r.db.Query(
		`SELECT id, email, password_hash, github_token, github_username, github_connected_at,
		 workos_id, organization_id, sso_connection_id, sso_provider, created_at, updated_at
		 FROM users WHERE organization_id = ?`,
		orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by organization: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		var githubToken, githubUsername sql.NullString
		var githubConnectedAt sql.NullTime
		var workosID, organizationID, ssoConnectionID, ssoProvider sql.NullString

		err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &githubToken, &githubUsername, &githubConnectedAt,
			&workosID, &organizationID, &ssoConnectionID, &ssoProvider, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		user.GitHubToken = githubToken.String
		user.GitHubUsername = githubUsername.String
		if githubConnectedAt.Valid {
			user.GitHubConnectedAt = &githubConnectedAt.Time
		}
		user.WorkOSID = workosID.String
		user.OrganizationID = organizationID.String
		user.SSOConnectionID = ssoConnectionID.String
		user.SSOProvider = ssoProvider.String

		users = append(users, user)
	}

	return users, nil
}
