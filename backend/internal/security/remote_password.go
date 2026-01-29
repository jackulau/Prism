package security

import (
	"errors"
	"unicode"
)

// RemotePasswordConfig contains password policy configuration
type RemotePasswordConfig struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireDigit     bool
	RequireSpecial   bool
}

// DefaultRemotePasswordConfig returns the default password policy for remote access
func DefaultRemotePasswordConfig() *RemotePasswordConfig {
	return &RemotePasswordConfig{
		MinLength:        12,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
		RequireSpecial:   true,
	}
}

// Password validation errors
var (
	ErrPasswordTooShort    = errors.New("password must be at least 12 characters")
	ErrPasswordNoUppercase = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLowercase = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoDigit     = errors.New("password must contain at least one digit")
	ErrPasswordNoSpecial   = errors.New("password must contain at least one special character")
)

// ValidateRemotePassword validates a password against the remote access password policy
func ValidateRemotePassword(password string) error {
	return ValidateRemotePasswordWithConfig(password, DefaultRemotePasswordConfig())
}

// ValidateRemotePasswordWithConfig validates a password against the given policy configuration
func ValidateRemotePasswordWithConfig(password string, config *RemotePasswordConfig) error {
	if len(password) < config.MinLength {
		return ErrPasswordTooShort
	}

	var hasUppercase, hasLowercase, hasDigit, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUppercase = true
		case unicode.IsLower(char):
			hasLowercase = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if config.RequireUppercase && !hasUppercase {
		return ErrPasswordNoUppercase
	}
	if config.RequireLowercase && !hasLowercase {
		return ErrPasswordNoLowercase
	}
	if config.RequireDigit && !hasDigit {
		return ErrPasswordNoDigit
	}
	if config.RequireSpecial && !hasSpecial {
		return ErrPasswordNoSpecial
	}

	return nil
}

// HashRemotePassword hashes a password for remote access storage using Argon2id
// This wraps the existing HashPassword function for clarity and potential future customization
func HashRemotePassword(password string) (string, error) {
	if err := ValidateRemotePassword(password); err != nil {
		return "", err
	}
	return HashPassword(password)
}

// VerifyRemotePassword verifies a password against a stored hash
// This wraps the existing VerifyPassword function for clarity
func VerifyRemotePassword(password, hashedPassword string) bool {
	return VerifyPassword(password, hashedPassword)
}
