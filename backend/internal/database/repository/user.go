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
	ExpiresAt        time.Time
	LastActivityAt   *time.Time
	IPAddress        string
	UserAgent        string
	DeviceName       string
	CreatedAt        time.Time
}

// SessionRepository handles session database operations
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create creates a new session (basic version for backwards compatibility)
func (r *SessionRepository) Create(userID, refreshTokenHash string, expiresAt time.Time) (*Session, error) {
	return r.CreateWithMetadata(userID, refreshTokenHash, expiresAt, "", "", "")
}

// CreateWithMetadata creates a new session with device metadata
func (r *SessionRepository) CreateWithMetadata(userID, refreshTokenHash string, expiresAt time.Time, ipAddress, userAgent, deviceName string) (*Session, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := r.db.Exec(
		`INSERT INTO sessions (id, user_id, refresh_token_hash, expires_at, last_activity_at, ip_address, user_agent, device_name, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, userID, refreshTokenHash, expiresAt, now, ipAddress, userAgent, deviceName, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &Session{
		ID:               id,
		UserID:           userID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        expiresAt,
		LastActivityAt:   &now,
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
		DeviceName:       deviceName,
		CreatedAt:        now,
	}, nil
}

// GetByRefreshTokenHash retrieves a session by refresh token hash
func (r *SessionRepository) GetByRefreshTokenHash(hash string) (*Session, error) {
	session := &Session{}
	var lastActivityAt sql.NullTime
	var ipAddress, userAgent, deviceName sql.NullString

	err := r.db.QueryRow(
		`SELECT id, user_id, refresh_token_hash, expires_at, last_activity_at, ip_address, user_agent, device_name, created_at FROM sessions WHERE refresh_token_hash = ?`,
		hash,
	).Scan(&session.ID, &session.UserID, &session.RefreshTokenHash, &session.ExpiresAt, &lastActivityAt, &ipAddress, &userAgent, &deviceName, &session.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if lastActivityAt.Valid {
		session.LastActivityAt = &lastActivityAt.Time
	}
	session.IPAddress = ipAddress.String
	session.UserAgent = userAgent.String
	session.DeviceName = deviceName.String

	return session, nil
}

// GetByID retrieves a session by its ID
func (r *SessionRepository) GetByID(id string) (*Session, error) {
	session := &Session{}
	var lastActivityAt sql.NullTime
	var ipAddress, userAgent, deviceName sql.NullString

	err := r.db.QueryRow(
		`SELECT id, user_id, refresh_token_hash, expires_at, last_activity_at, ip_address, user_agent, device_name, created_at FROM sessions WHERE id = ?`,
		id,
	).Scan(&session.ID, &session.UserID, &session.RefreshTokenHash, &session.ExpiresAt, &lastActivityAt, &ipAddress, &userAgent, &deviceName, &session.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if lastActivityAt.Valid {
		session.LastActivityAt = &lastActivityAt.Time
	}
	session.IPAddress = ipAddress.String
	session.UserAgent = userAgent.String
	session.DeviceName = deviceName.String

	return session, nil
}

// GetByUserID retrieves all sessions for a user
func (r *SessionRepository) GetByUserID(userID string) ([]*Session, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, refresh_token_hash, expires_at, last_activity_at, ip_address, user_agent, device_name, created_at FROM sessions WHERE user_id = ? ORDER BY last_activity_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		session := &Session{}
		var lastActivityAt sql.NullTime
		var ipAddress, userAgent, deviceName sql.NullString

		err := rows.Scan(&session.ID, &session.UserID, &session.RefreshTokenHash, &session.ExpiresAt, &lastActivityAt, &ipAddress, &userAgent, &deviceName, &session.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		if lastActivityAt.Valid {
			session.LastActivityAt = &lastActivityAt.Time
		}
		session.IPAddress = ipAddress.String
		session.UserAgent = userAgent.String
		session.DeviceName = deviceName.String

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// UpdateActivity updates the last activity timestamp for a session
func (r *SessionRepository) UpdateActivity(sessionID string) error {
	now := time.Now()
	_, err := r.db.Exec(`UPDATE sessions SET last_activity_at = ? WHERE id = ?`, now, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update activity: %w", err)
	}
	return nil
}

// DeleteByIDAndUserID deletes a specific session for a user
func (r *SessionRepository) DeleteByIDAndUserID(sessionID, userID string) error {
	result, err := r.db.Exec(`DELETE FROM sessions WHERE id = ? AND user_id = ?`, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("session not found")
	}
	return nil
}

// DeleteOthers deletes all sessions for a user except the specified one
func (r *SessionRepository) DeleteOthers(userID, currentSessionID string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE user_id = ? AND id != ?`, userID, currentSessionID)
	if err != nil {
		return fmt.Errorf("failed to delete other sessions: %w", err)
	}
	return nil
}

// DeleteIdle deletes sessions that have been idle longer than the threshold
func (r *SessionRepository) DeleteIdle(idleThreshold time.Duration) (int64, error) {
	threshold := time.Now().Add(-idleThreshold)
	result, err := r.db.Exec(`DELETE FROM sessions WHERE last_activity_at < ? AND last_activity_at IS NOT NULL`, threshold)
	if err != nil {
		return 0, fmt.Errorf("failed to delete idle sessions: %w", err)
	}
	count, _ := result.RowsAffected()
	return count, nil
}

// CountByUserID counts the number of sessions for a user
func (r *SessionRepository) CountByUserID(userID string) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM sessions WHERE user_id = ?`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count sessions: %w", err)
	}
	return count, nil
}

// DeleteOldestByUserID deletes the oldest sessions for a user, keeping only the specified number
func (r *SessionRepository) DeleteOldestByUserID(userID string, keepCount int) error {
	_, err := r.db.Exec(`
		DELETE FROM sessions WHERE user_id = ? AND id NOT IN (
			SELECT id FROM sessions WHERE user_id = ? ORDER BY last_activity_at DESC LIMIT ?
		)
	`, userID, userID, keepCount)
	if err != nil {
		return fmt.Errorf("failed to delete oldest sessions: %w", err)
	}
	return nil
}

// UpdateTokenHash updates the refresh token hash for a session
func (r *SessionRepository) UpdateTokenHash(sessionID, tokenHash string) error {
	_, err := r.db.Exec(`UPDATE sessions SET refresh_token_hash = ? WHERE id = ?`, tokenHash, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update token hash: %w", err)
	}
	return nil
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

// DeleteExpired deletes all expired sessions and returns the count
func (r *SessionRepository) DeleteExpired() (int64, error) {
	result, err := r.db.Exec(`DELETE FROM sessions WHERE expires_at < ?`, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	count, _ := result.RowsAffected()
	return count, nil
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
