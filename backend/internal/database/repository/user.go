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
	CreatedAt         time.Time
	UpdatedAt         time.Time
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

	err := r.db.QueryRow(
		`SELECT id, email, password_hash, github_token, github_username, github_connected_at, created_at, updated_at FROM users WHERE id = ?`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &githubToken, &githubUsername, &githubConnectedAt, &user.CreatedAt, &user.UpdatedAt)

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

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(email string) (*User, error) {
	user := &User{}
	var githubToken, githubUsername sql.NullString
	var githubConnectedAt sql.NullTime

	err := r.db.QueryRow(
		`SELECT id, email, password_hash, github_token, github_username, github_connected_at, created_at, updated_at FROM users WHERE email = ?`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &githubToken, &githubUsername, &githubConnectedAt, &user.CreatedAt, &user.UpdatedAt)

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
