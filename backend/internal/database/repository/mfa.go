package repository

import (
	"database/sql"
	"fmt"
	"time"
)

// MFA represents a user's MFA configuration in the database
type MFA struct {
	ID                   int64
	UserID               string
	SecretEncrypted      []byte
	SecretNonce          []byte
	IsEnabled            bool
	BackupCodesEncrypted []byte
	BackupCodesNonce     []byte
	CreatedAt            time.Time
	VerifiedAt           *time.Time
}

// MFAVerificationAttempt represents a verification attempt log entry
type MFAVerificationAttempt struct {
	ID          int64
	UserID      string
	Success     bool
	IPAddress   string
	AttemptedAt time.Time
}

// MFARepository handles MFA database operations
type MFARepository struct {
	db *sql.DB
}

// NewMFARepository creates a new MFA repository
func NewMFARepository(db *sql.DB) *MFARepository {
	return &MFARepository{db: db}
}

// CreateMFASetup creates a new MFA setup (pending verification)
func (r *MFARepository) CreateMFASetup(userID string, encryptedSecret, nonce []byte) error {
	// Delete any existing MFA setup for this user first
	_, err := r.db.Exec(`DELETE FROM user_mfa WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("failed to delete existing MFA setup: %w", err)
	}

	_, err = r.db.Exec(
		`INSERT INTO user_mfa (user_id, secret_encrypted, secret_nonce, is_enabled, created_at)
		 VALUES (?, ?, ?, 0, ?)`,
		userID, encryptedSecret, nonce, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create MFA setup: %w", err)
	}

	return nil
}

// GetMFAByUserID retrieves MFA configuration for a user
func (r *MFARepository) GetMFAByUserID(userID string) (*MFA, error) {
	mfa := &MFA{}
	var verifiedAt sql.NullTime
	var backupCodesEncrypted, backupCodesNonce []byte

	err := r.db.QueryRow(
		`SELECT id, user_id, secret_encrypted, secret_nonce, is_enabled,
		 backup_codes_encrypted, backup_codes_nonce, created_at, verified_at
		 FROM user_mfa WHERE user_id = ?`,
		userID,
	).Scan(
		&mfa.ID, &mfa.UserID, &mfa.SecretEncrypted, &mfa.SecretNonce,
		&mfa.IsEnabled, &backupCodesEncrypted, &backupCodesNonce,
		&mfa.CreatedAt, &verifiedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get MFA: %w", err)
	}

	mfa.BackupCodesEncrypted = backupCodesEncrypted
	mfa.BackupCodesNonce = backupCodesNonce
	if verifiedAt.Valid {
		mfa.VerifiedAt = &verifiedAt.Time
	}

	return mfa, nil
}

// EnableMFA enables MFA for a user (after successful verification)
func (r *MFARepository) EnableMFA(userID string, backupCodesEncrypted, backupCodesNonce []byte) error {
	now := time.Now()
	result, err := r.db.Exec(
		`UPDATE user_mfa SET is_enabled = 1, backup_codes_encrypted = ?, backup_codes_nonce = ?, verified_at = ?
		 WHERE user_id = ?`,
		backupCodesEncrypted, backupCodesNonce, now, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to enable MFA: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("MFA setup not found for user")
	}

	return nil
}

// DisableMFA disables MFA for a user
func (r *MFARepository) DisableMFA(userID string) error {
	_, err := r.db.Exec(`DELETE FROM user_mfa WHERE user_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("failed to disable MFA: %w", err)
	}
	return nil
}

// UpdateBackupCodes updates the backup codes for a user
func (r *MFARepository) UpdateBackupCodes(userID string, backupCodesEncrypted, backupCodesNonce []byte) error {
	result, err := r.db.Exec(
		`UPDATE user_mfa SET backup_codes_encrypted = ?, backup_codes_nonce = ? WHERE user_id = ? AND is_enabled = 1`,
		backupCodesEncrypted, backupCodesNonce, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update backup codes: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("MFA not enabled for user")
	}

	return nil
}

// LogVerificationAttempt logs an MFA verification attempt
func (r *MFARepository) LogVerificationAttempt(userID string, success bool, ipAddress string) error {
	_, err := r.db.Exec(
		`INSERT INTO mfa_verification_attempts (user_id, success, ip_address, attempted_at) VALUES (?, ?, ?, ?)`,
		userID, success, ipAddress, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to log verification attempt: %w", err)
	}
	return nil
}

// GetRecentFailedAttempts returns the count of failed attempts in the given duration
func (r *MFARepository) GetRecentFailedAttempts(userID string, duration time.Duration) (int, error) {
	since := time.Now().Add(-duration)
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM mfa_verification_attempts
		 WHERE user_id = ? AND success = 0 AND attempted_at > ?`,
		userID, since,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get failed attempts: %w", err)
	}
	return count, nil
}

// GetLastSuccessfulAttempt returns the most recent successful verification time
func (r *MFARepository) GetLastSuccessfulAttempt(userID string) (*time.Time, error) {
	var attemptedAt sql.NullTime
	err := r.db.QueryRow(
		`SELECT attempted_at FROM mfa_verification_attempts
		 WHERE user_id = ? AND success = 1
		 ORDER BY attempted_at DESC LIMIT 1`,
		userID,
	).Scan(&attemptedAt)

	if err == sql.ErrNoRows || !attemptedAt.Valid {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get last successful attempt: %w", err)
	}

	return &attemptedAt.Time, nil
}

// CleanupOldAttempts removes verification attempts older than the given duration
func (r *MFARepository) CleanupOldAttempts(olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	_, err := r.db.Exec(
		`DELETE FROM mfa_verification_attempts WHERE attempted_at < ?`,
		cutoff,
	)
	if err != nil {
		return fmt.Errorf("failed to cleanup old attempts: %w", err)
	}
	return nil
}

// IsMFAEnabled checks if MFA is enabled for a user
func (r *MFARepository) IsMFAEnabled(userID string) (bool, error) {
	var isEnabled bool
	err := r.db.QueryRow(
		`SELECT is_enabled FROM user_mfa WHERE user_id = ?`,
		userID,
	).Scan(&isEnabled)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check MFA status: %w", err)
	}

	return isEnabled, nil
}
