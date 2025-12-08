package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

// EncryptionService handles AES-256-GCM encryption for sensitive data
type EncryptionService struct {
	masterKey []byte
}

// NewEncryptionService creates a new encryption service with the given key
func NewEncryptionService(key string) (*EncryptionService, error) {
	if key == "" {
		// Generate a random key for development (should be set in production)
		randomKey := make([]byte, 32)
		if _, err := rand.Read(randomKey); err != nil {
			return nil, fmt.Errorf("failed to generate random key: %w", err)
		}
		return &EncryptionService{masterKey: randomKey}, nil
	}

	// Decode hex key
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("encryption key must be a valid hex string: %w", err)
	}

	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes (64 hex characters), got %d bytes", len(keyBytes))
	}

	return &EncryptionService{masterKey: keyBytes}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM
func (s *EncryptionService) Encrypt(plaintext []byte) (ciphertext, nonce []byte, err error) {
	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce = make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM
func (s *EncryptionService) Decrypt(ciphertext, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// HashPassword hashes a password using Argon2id
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Argon2id parameters: time=1, memory=64MB, threads=4, keyLen=32
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	// Encode as: salt$hash (both hex encoded)
	return fmt.Sprintf("%s$%s", hex.EncodeToString(salt), hex.EncodeToString(hash)), nil
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(password, hashedPassword string) bool {
	// Parse salt$hash format
	var saltHex, hashHex string
	n, _ := fmt.Sscanf(hashedPassword, "%64s", &saltHex)
	if n != 1 {
		return false
	}

	// Find the $ separator
	for i := 0; i < len(hashedPassword); i++ {
		if hashedPassword[i] == '$' {
			saltHex = hashedPassword[:i]
			hashHex = hashedPassword[i+1:]
			break
		}
	}

	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return false
	}

	expectedHash, err := hex.DecodeString(hashHex)
	if err != nil {
		return false
	}

	// Compute hash with same parameters
	computedHash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	// Constant-time comparison
	if len(computedHash) != len(expectedHash) {
		return false
	}

	var result byte
	for i := 0; i < len(computedHash); i++ {
		result |= computedHash[i] ^ expectedHash[i]
	}

	return result == 0
}

// HashAPIKey creates a SHA-256 hash of an API key for storage
func HashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// GenerateAPIKey generates a random API key with a prefix
func GenerateAPIKey(prefix string) (key string, keyPrefix string, err error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	key = prefix + "_" + hex.EncodeToString(randomBytes)
	keyPrefix = key[:12] // First 12 characters for identification

	return key, keyPrefix, nil
}

// GenerateRandomString generates a random hex string of the specified byte length
func GenerateRandomString(byteLength int) (string, error) {
	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
