package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// MFAService handles TOTP-based multi-factor authentication
type MFAService struct {
	issuer            string
	encryptionService *EncryptionService
}

// NewMFAService creates a new MFA service
func NewMFAService(issuer string, encryptionService *EncryptionService) *MFAService {
	if issuer == "" {
		issuer = "Prism"
	}
	return &MFAService{
		issuer:            issuer,
		encryptionService: encryptionService,
	}
}

// TOTPSetup contains the information needed for MFA setup
type TOTPSetup struct {
	Secret          string
	QRCodeURL       string
	EncryptedSecret []byte
	SecretNonce     []byte
}

// GenerateTOTPSecret generates a new TOTP secret for the given email
func (s *MFAService) GenerateTOTPSecret(email string) (*TOTPSetup, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: email,
		Period:      30,
		SecretSize:  20, // 160-bit secret
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	// Encrypt the secret for storage
	encryptedSecret, nonce, err := s.encryptionService.Encrypt([]byte(key.Secret()))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret: %w", err)
	}

	return &TOTPSetup{
		Secret:          key.Secret(),
		QRCodeURL:       key.URL(),
		EncryptedSecret: encryptedSecret,
		SecretNonce:     nonce,
	}, nil
}

// VerifyTOTPCode verifies a 6-digit TOTP code with ±1 time step tolerance
func (s *MFAService) VerifyTOTPCode(secret, code string) bool {
	// Validate code format (6 digits)
	if len(code) != 6 {
		return false
	}

	return totp.Validate(code, secret)
}

// VerifyTOTPCodeWithSkew verifies a TOTP code with custom time skew tolerance
func (s *MFAService) VerifyTOTPCodeWithSkew(secret, code string, skew uint) bool {
	if len(code) != 6 {
		return false
	}

	valid, err := totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      skew, // Allow ±skew time steps
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})

	return err == nil && valid
}

// DecryptSecret decrypts an encrypted TOTP secret
func (s *MFAService) DecryptSecret(encryptedSecret, nonce []byte) (string, error) {
	decrypted, err := s.encryptionService.Decrypt(encryptedSecret, nonce)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secret: %w", err)
	}
	return string(decrypted), nil
}

// BackupCodes contains backup codes and their encrypted form
type BackupCodes struct {
	Codes          []string
	EncryptedCodes []byte
	Nonce          []byte
}

// GenerateBackupCodes generates backup codes for account recovery
func (s *MFAService) GenerateBackupCodes(count int) (*BackupCodes, error) {
	if count <= 0 {
		count = 10
	}

	codes := make([]string, count)
	for i := 0; i < count; i++ {
		code, err := generateBackupCode(8)
		if err != nil {
			return nil, fmt.Errorf("failed to generate backup code: %w", err)
		}
		codes[i] = code
	}

	// Hash all codes for storage (separated by newlines)
	hashedCodes := s.hashBackupCodes(codes)

	// Encrypt the hashed codes
	encrypted, nonce, err := s.encryptionService.Encrypt([]byte(hashedCodes))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt backup codes: %w", err)
	}

	return &BackupCodes{
		Codes:          codes,
		EncryptedCodes: encrypted,
		Nonce:          nonce,
	}, nil
}

// hashBackupCodes hashes backup codes for storage
func (s *MFAService) hashBackupCodes(codes []string) string {
	hashed := make([]string, len(codes))
	for i, code := range codes {
		hashed[i] = HashAPIKey(strings.ToUpper(code))
	}
	return strings.Join(hashed, "\n")
}

// VerifyBackupCode verifies a backup code and returns the remaining hashed codes
func (s *MFAService) VerifyBackupCode(encryptedCodes, nonce []byte, code string) (remainingEncrypted []byte, remainingNonce []byte, valid bool, err error) {
	// Decrypt the hashed codes
	decrypted, err := s.encryptionService.Decrypt(encryptedCodes, nonce)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to decrypt backup codes: %w", err)
	}

	hashedCodes := strings.Split(string(decrypted), "\n")
	codeHash := HashAPIKey(strings.ToUpper(code))

	// Find and remove the matching code
	foundIndex := -1
	for i, hashed := range hashedCodes {
		if subtle.ConstantTimeCompare([]byte(hashed), []byte(codeHash)) == 1 {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		return encryptedCodes, nonce, false, nil
	}

	// Remove the used code
	remainingCodes := append(hashedCodes[:foundIndex], hashedCodes[foundIndex+1:]...)
	newHashedCodes := strings.Join(remainingCodes, "\n")

	// Re-encrypt the remaining codes
	newEncrypted, newNonce, err := s.encryptionService.Encrypt([]byte(newHashedCodes))
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to re-encrypt backup codes: %w", err)
	}

	return newEncrypted, newNonce, true, nil
}

// GetBackupCodeCount returns the number of remaining backup codes
func (s *MFAService) GetBackupCodeCount(encryptedCodes, nonce []byte) (int, error) {
	if encryptedCodes == nil {
		return 0, nil
	}

	decrypted, err := s.encryptionService.Decrypt(encryptedCodes, nonce)
	if err != nil {
		return 0, fmt.Errorf("failed to decrypt backup codes: %w", err)
	}

	if len(decrypted) == 0 {
		return 0, nil
	}

	codes := strings.Split(string(decrypted), "\n")
	count := 0
	for _, c := range codes {
		if c != "" {
			count++
		}
	}
	return count, nil
}

// generateBackupCode generates a random alphanumeric backup code
func generateBackupCode(length int) (string, error) {
	// Use base32 encoding (A-Z, 2-7) for unambiguous characters
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Convert to hex and take first 'length' characters
	encoded := hex.EncodeToString(bytes)
	if len(encoded) > length {
		encoded = encoded[:length]
	}

	return strings.ToUpper(encoded), nil
}

// GenerateBase32Secret generates a random base32-encoded secret
func GenerateBase32Secret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes), nil
}
