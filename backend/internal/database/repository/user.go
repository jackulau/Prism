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
	Role              string // "user" or "admin"
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
	var workosID, organizationID, ssoConnectionID, ssoProvider, role sql.NullString

	err := r.db.QueryRow(
		`SELECT id, email, password_hash, role, github_token, github_username, github_connected_at,
		 workos_id, organization_id, sso_connection_id, sso_provider, created_at, updated_at
		 FROM users WHERE id = ?`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &role, &githubToken, &githubUsername, &githubConnectedAt,
		&workosID, &organizationID, &ssoConnectionID, &ssoProvider, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.Role = role.String
	if user.Role == "" {
		user.Role = "user" // Default role
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
	var workosID, organizationID, ssoConnectionID, ssoProvider, role sql.NullString

	err := r.db.QueryRow(
		`SELECT id, email, password_hash, role, github_token, github_username, github_connected_at,
		 workos_id, organization_id, sso_connection_id, sso_provider, created_at, updated_at
		 FROM users WHERE email = ?`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &role, &githubToken, &githubUsername, &githubConnectedAt,
		&workosID, &organizationID, &ssoConnectionID, &ssoProvider, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.Role = role.String
	if user.Role == "" {
		user.Role = "user" // Default role
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

// Create creates a new session
func (r *SessionRepository) Create(userID, refreshTokenHash string, expiresAt time.Time) (*Session, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := r.db.Exec(
		`INSERT INTO sessions (id, user_id, refresh_token_hash, expires_at, created_at) VALUES (?, ?, ?, ?, ?)`,
		id, userID, refreshTokenHash, expiresAt, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &Session{
		ID:               id,
		UserID:           userID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        expiresAt,
		CreatedAt:        now,
	}, nil
}

// GetByRefreshTokenHash retrieves a session by refresh token hash
func (r *SessionRepository) GetByRefreshTokenHash(hash string) (*Session, error) {
	session := &Session{}
	err := r.db.QueryRow(
		`SELECT id, user_id, refresh_token_hash, expires_at, created_at FROM sessions WHERE refresh_token_hash = ?`,
		hash,
	).Scan(&session.ID, &session.UserID, &session.RefreshTokenHash, &session.ExpiresAt, &session.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
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
	var workosIDVal, organizationID, ssoConnectionID, ssoProvider, role sql.NullString

	err := r.db.QueryRow(
		`SELECT id, email, password_hash, role, github_token, github_username, github_connected_at,
		 workos_id, organization_id, sso_connection_id, sso_provider, created_at, updated_at
		 FROM users WHERE workos_id = ?`,
		workosID,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &role, &githubToken, &githubUsername, &githubConnectedAt,
		&workosIDVal, &organizationID, &ssoConnectionID, &ssoProvider, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by WorkOS ID: %w", err)
	}

	user.Role = role.String
	if user.Role == "" {
		user.Role = "user" // Default role
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
		`SELECT id, email, password_hash, role, github_token, github_username, github_connected_at,
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
		var workosID, organizationID, ssoConnectionID, ssoProvider, role sql.NullString

		err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &role, &githubToken, &githubUsername, &githubConnectedAt,
			&workosID, &organizationID, &ssoConnectionID, &ssoProvider, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		user.Role = role.String
		if user.Role == "" {
			user.Role = "user" // Default role
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

// GetUserRole retrieves just the role for a user by ID
func (r *UserRepository) GetUserRole(userID string) (string, error) {
	var role sql.NullString
	err := r.db.QueryRow(`SELECT role FROM users WHERE id = ?`, userID).Scan(&role)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("user not found")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get user role: %w", err)
	}
	if role.String == "" {
		return "user", nil // Default role
	}
	return role.String, nil
}

// SetUserRole updates the role for a user
func (r *UserRepository) SetUserRole(userID string, role string) error {
	now := time.Now()
	result, err := r.db.Exec(
		`UPDATE users SET role = ?, updated_at = ? WHERE id = ?`,
		role, now, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to set user role: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// GetUsersByRole retrieves all users with a specific role
func (r *UserRepository) GetUsersByRole(role string) ([]*User, error) {
	rows, err := r.db.Query(
		`SELECT id, email, password_hash, role, github_token, github_username, github_connected_at,
		 workos_id, organization_id, sso_connection_id, sso_provider, created_at, updated_at
		 FROM users WHERE role = ?`,
		role,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by role: %w", err)
	}
	defer rows.Close()

	return r.scanUsers(rows)
}

// GetAllUsers retrieves all users with pagination
func (r *UserRepository) GetAllUsers(limit, offset int) ([]*User, error) {
	rows, err := r.db.Query(
		`SELECT id, email, password_hash, role, github_token, github_username, github_connected_at,
		 workos_id, organization_id, sso_connection_id, sso_provider, created_at, updated_at
		 FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	defer rows.Close()

	return r.scanUsers(rows)
}

// CountUsersByRole counts users with a specific role
func (r *UserRepository) CountUsersByRole(role string) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = ?`, role).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users by role: %w", err)
	}
	return count, nil
}

// CountAllUsers counts all users
func (r *UserRepository) CountAllUsers() (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count all users: %w", err)
	}
	return count, nil
}

// scanUsers is a helper to scan user rows
func (r *UserRepository) scanUsers(rows *sql.Rows) ([]*User, error) {
	var users []*User
	for rows.Next() {
		user := &User{}
		var githubToken, githubUsername sql.NullString
		var githubConnectedAt sql.NullTime
		var workosID, organizationID, ssoConnectionID, ssoProvider, role sql.NullString

		err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &role, &githubToken, &githubUsername, &githubConnectedAt,
			&workosID, &organizationID, &ssoConnectionID, &ssoProvider, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		user.Role = role.String
		if user.Role == "" {
			user.Role = "user" // Default role
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
