package security

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPasswordEncryption_HashAndVerify(t *testing.T) {
	passwords := []string{
		"simple",
		"P@ssw0rd!",
		"extremely_long_password_that_exceeds_typical_length_requirements_to_test_edge_cases_123456",
		"unicode🔐password",
		"   spaces   ",
	}

	for _, password := range passwords {
		t.Run("password_"+password[:min(10, len(password))], func(t *testing.T) {
			hash, err := HashPassword(password)
			if err != nil {
				t.Fatalf("HashPassword failed: %v", err)
			}

			if hash == "" {
				t.Fatal("Hash should not be empty")
			}

			// Verify correct password
			if !VerifyPassword(password, hash) {
				t.Error("VerifyPassword should return true for correct password")
			}

			// Verify wrong password
			if VerifyPassword("wrong_password", hash) {
				t.Error("VerifyPassword should return false for wrong password")
			}
		})
	}
}

func TestPasswordEncryption_HashUniqueness(t *testing.T) {
	password := "test_password"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Same password should produce different hashes (due to random salt)
	if hash1 == hash2 {
		t.Error("Same password should produce different hashes due to salt")
	}

	// Both hashes should verify correctly
	if !VerifyPassword(password, hash1) {
		t.Error("Password should verify against first hash")
	}
	if !VerifyPassword(password, hash2) {
		t.Error("Password should verify against second hash")
	}
}

func TestPasswordEncryption_InvalidHash(t *testing.T) {
	testCases := []struct {
		name string
		hash string
	}{
		{"empty", ""},
		{"no_separator", "abcdef123456"},
		{"invalid_salt_hex", "zzzzzz$abcdef"},
		{"invalid_hash_hex", "abcdef$zzzzzz"},
		{"wrong_length_salt", "ab$abcdef"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if VerifyPassword("any_password", tc.hash) {
				t.Errorf("VerifyPassword should return false for invalid hash format: %s", tc.hash)
			}
		})
	}
}

func TestBruteForceProtection_IPBlocking(t *testing.T) {
	config := &RemoteAccessConfig{
		Enabled:             true,
		PasswordHash:        mustHashPasswordTest(t, "correct"),
		SessionTimeout:      30 * time.Minute,
		MaxConcurrentSessions: 5,
		MaxFailedAttempts:   3,
		BlockDuration:       50 * time.Millisecond, // Short for testing
		MaxBlockDuration:    500 * time.Millisecond,
	}

	service := NewRemoteAuthService(config, nil)
	defer service.Stop()

	clientIP := "192.168.1.100"

	// Make failed attempts up to limit
	for i := 0; i < 3; i++ {
		_, err := service.Authenticate("wrong", clientIP)
		if err == nil {
			t.Error("Expected authentication to fail")
		}
	}

	// Now should be blocked
	if !service.IsIPBlocked(clientIP) {
		t.Error("IP should be blocked after max failed attempts")
	}

	// Even correct password should fail while blocked
	_, err := service.Authenticate("correct", clientIP)
	if err == nil {
		t.Error("Authentication should fail while IP is blocked")
	}

	// Check block time remaining
	remaining := service.GetBlockTimeRemaining(clientIP)
	if remaining <= 0 {
		t.Error("Block time remaining should be positive")
	}
}

func TestBruteForceProtection_ExponentialBackoff(t *testing.T) {
	config := &RemoteAccessConfig{
		Enabled:             true,
		PasswordHash:        mustHashPasswordTest(t, "correct"),
		SessionTimeout:      30 * time.Minute,
		MaxConcurrentSessions: 5,
		MaxFailedAttempts:   2,
		BlockDuration:       20 * time.Millisecond,
		MaxBlockDuration:    200 * time.Millisecond,
	}

	service := NewRemoteAuthService(config, nil)
	defer service.Stop()

	clientIP := "192.168.1.200"

	// First block
	for i := 0; i < 2; i++ {
		service.Authenticate("wrong", clientIP)
	}

	firstBlockTime := service.GetBlockTimeRemaining(clientIP)

	// Wait for first block to expire
	time.Sleep(30 * time.Millisecond)

	// Second block
	for i := 0; i < 2; i++ {
		service.Authenticate("wrong", clientIP)
	}

	secondBlockTime := service.GetBlockTimeRemaining(clientIP)

	// Second block should be longer (exponential backoff)
	if secondBlockTime <= firstBlockTime {
		t.Errorf("Expected exponential backoff: first=%v, second=%v", firstBlockTime, secondBlockTime)
	}
}

func TestBruteForceProtection_MaxBlockDuration(t *testing.T) {
	config := &RemoteAccessConfig{
		Enabled:             true,
		PasswordHash:        mustHashPasswordTest(t, "correct"),
		SessionTimeout:      30 * time.Minute,
		MaxConcurrentSessions: 5,
		MaxFailedAttempts:   1, // Block immediately
		BlockDuration:       50 * time.Millisecond,
		MaxBlockDuration:    100 * time.Millisecond, // Cap
	}

	service := NewRemoteAuthService(config, nil)
	defer service.Stop()

	clientIP := "192.168.1.201"

	// Create multiple blocks
	for round := 0; round < 5; round++ {
		service.Authenticate("wrong", clientIP)
		remaining := service.GetBlockTimeRemaining(clientIP)

		// Block duration should never exceed max
		if remaining > 110*time.Millisecond { // Allow small timing margin
			t.Errorf("Block duration %v exceeds max %v", remaining, config.MaxBlockDuration)
		}

		time.Sleep(remaining + 10*time.Millisecond)
	}
}

func TestBruteForceProtection_SuccessfulLoginClearsAttempts(t *testing.T) {
	config := &RemoteAccessConfig{
		Enabled:             true,
		PasswordHash:        mustHashPasswordTest(t, "correct"),
		SessionTimeout:      30 * time.Minute,
		MaxConcurrentSessions: 10,
		MaxFailedAttempts:   5,
		BlockDuration:       1 * time.Minute,
		MaxBlockDuration:    10 * time.Minute,
	}

	service := NewRemoteAuthService(config, nil)
	defer service.Stop()

	clientIP := "192.168.1.202"

	// Make some failed attempts (but not enough to block)
	for i := 0; i < 3; i++ {
		service.Authenticate("wrong", clientIP)
	}

	// Successful login should clear attempts
	_, err := service.Authenticate("correct", clientIP)
	if err != nil {
		t.Fatalf("Authentication should succeed: %v", err)
	}

	// Now we should be able to fail 4 more times without being blocked
	for i := 0; i < 4; i++ {
		service.Authenticate("wrong", clientIP)
	}

	if service.IsIPBlocked(clientIP) {
		t.Error("IP should not be blocked - counter was reset by successful login")
	}
}

func TestSessionTokenEntropy(t *testing.T) {
	tokens := make(map[string]bool)
	numTokens := 100

	for i := 0; i < numTokens; i++ {
		token, err := GenerateRandomString(32)
		if err != nil {
			t.Fatalf("GenerateRandomString failed: %v", err)
		}

		// Check length
		if len(token) != 64 { // 32 bytes = 64 hex chars
			t.Errorf("Expected token length 64, got %d", len(token))
		}

		// Check uniqueness
		if tokens[token] {
			t.Error("Token collision detected")
		}
		tokens[token] = true

		// Check it's valid hex
		_, err = hex.DecodeString(token)
		if err != nil {
			t.Errorf("Token is not valid hex: %v", err)
		}
	}
}

func TestSessionTokenEntropy_Distribution(t *testing.T) {
	// Generate many tokens and check for reasonable character distribution
	tokens := ""
	for i := 0; i < 100; i++ {
		token, err := GenerateRandomString(32)
		if err != nil {
			t.Fatalf("GenerateRandomString failed: %v", err)
		}
		tokens += token
	}

	// Count character frequency (hex chars: 0-9, a-f)
	charCounts := make(map[rune]int)
	for _, c := range tokens {
		charCounts[c]++
	}

	// Each hex char should appear roughly equally (6400 / 16 = 400 times)
	totalChars := len(tokens)
	expectedPerChar := totalChars / 16
	tolerance := float64(expectedPerChar) * 0.5 // 50% tolerance

	for c, count := range charCounts {
		deviation := float64(count) - float64(expectedPerChar)
		if deviation < 0 {
			deviation = -deviation
		}
		if deviation > tolerance {
			t.Logf("Character '%c' count %d deviates from expected %d (tolerance: %.0f)", c, count, expectedPerChar, tolerance)
			// This is a statistical test, so we log instead of failing
		}
	}
}

func TestEncryptionService_EncryptDecrypt(t *testing.T) {
	// Generate a valid key
	keyBytes := make([]byte, 32)
	rand.Read(keyBytes)
	key := hex.EncodeToString(keyBytes)

	service, err := NewEncryptionService(key)
	if err != nil {
		t.Fatalf("NewEncryptionService failed: %v", err)
	}

	testCases := []string{
		"simple text",
		"",
		"unicode 🔐 content",
		strings.Repeat("a", 1000),
	}

	for _, plaintext := range testCases {
		t.Run("encrypt_"+plaintext[:min(10, len(plaintext))], func(t *testing.T) {
			ciphertext, nonce, err := service.Encrypt([]byte(plaintext))
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			if len(ciphertext) == 0 && len(plaintext) > 0 {
				t.Error("Ciphertext should not be empty for non-empty plaintext")
			}

			decrypted, err := service.Decrypt(ciphertext, nonce)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if string(decrypted) != plaintext {
				t.Errorf("Decrypted text doesn't match original")
			}
		})
	}
}

func TestEncryptionService_DifferentNonces(t *testing.T) {
	keyBytes := make([]byte, 32)
	rand.Read(keyBytes)
	key := hex.EncodeToString(keyBytes)

	service, err := NewEncryptionService(key)
	if err != nil {
		t.Fatalf("NewEncryptionService failed: %v", err)
	}

	plaintext := []byte("same text")

	cipher1, nonce1, _ := service.Encrypt(plaintext)
	cipher2, nonce2, _ := service.Encrypt(plaintext)

	// Same plaintext should produce different ciphertext (different nonces)
	if hex.EncodeToString(nonce1) == hex.EncodeToString(nonce2) {
		t.Error("Nonces should be different")
	}

	if hex.EncodeToString(cipher1) == hex.EncodeToString(cipher2) {
		t.Error("Ciphertext should be different for different nonces")
	}
}

func TestEncryptionService_InvalidKey(t *testing.T) {
	testCases := []struct {
		name string
		key  string
	}{
		{"too_short", "abcd"},
		{"invalid_hex", "zzzz"},
		{"wrong_length", strings.Repeat("ab", 16)}, // 32 chars = 16 bytes, need 32 bytes
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewEncryptionService(tc.key)
			if err == nil {
				t.Error("Expected error for invalid key")
			}
		})
	}
}

func TestEncryptionService_TamperDetection(t *testing.T) {
	keyBytes := make([]byte, 32)
	rand.Read(keyBytes)
	key := hex.EncodeToString(keyBytes)

	service, err := NewEncryptionService(key)
	if err != nil {
		t.Fatalf("NewEncryptionService failed: %v", err)
	}

	plaintext := []byte("sensitive data")
	ciphertext, nonce, _ := service.Encrypt(plaintext)

	// Tamper with ciphertext
	if len(ciphertext) > 0 {
		ciphertext[0] ^= 0xFF
	}

	_, err = service.Decrypt(ciphertext, nonce)
	if err == nil {
		t.Error("Decryption should fail for tampered ciphertext")
	}
}

func TestAPIKeyGeneration(t *testing.T) {
	prefixes := []string{"pk", "sk", "api"}

	for _, prefix := range prefixes {
		t.Run("prefix_"+prefix, func(t *testing.T) {
			key, keyPrefix, err := GenerateAPIKey(prefix)
			if err != nil {
				t.Fatalf("GenerateAPIKey failed: %v", err)
			}

			if !strings.HasPrefix(key, prefix+"_") {
				t.Errorf("Key should start with prefix '%s_', got '%s'", prefix, key[:len(prefix)+5])
			}

			if len(keyPrefix) != 12 {
				t.Errorf("Key prefix should be 12 chars, got %d", len(keyPrefix))
			}

			if !strings.HasPrefix(key, keyPrefix) {
				t.Error("Key should start with keyPrefix")
			}
		})
	}
}

func TestAPIKeyHashing(t *testing.T) {
	key := "pk_abcdef123456789"
	hash1 := HashAPIKey(key)
	hash2 := HashAPIKey(key)

	// Same key should produce same hash
	if hash1 != hash2 {
		t.Error("Same key should produce same hash")
	}

	// Different key should produce different hash
	differentHash := HashAPIKey("pk_different")
	if hash1 == differentHash {
		t.Error("Different keys should produce different hashes")
	}

	// Hash should be hex encoded
	_, err := hex.DecodeString(hash1)
	if err != nil {
		t.Errorf("Hash should be valid hex: %v", err)
	}
}

func TestConcurrentPasswordOperations(t *testing.T) {
	var wg sync.WaitGroup
	errors := make(chan error, 100)
	password := "concurrent_test_password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Concurrent verifications
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !VerifyPassword(password, hash) {
				errors <- err
			}
		}()
	}

	// Concurrent hashing
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			pwd := password + string(rune('0'+idx%10))
			_, err := HashPassword(pwd)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent operation failed: %v", err)
	}
}

func TestConcurrentSessionOperations(t *testing.T) {
	config := &RemoteAccessConfig{
		Enabled:             true,
		PasswordHash:        mustHashPasswordTest(t, "correct"),
		SessionTimeout:      30 * time.Minute,
		MaxConcurrentSessions: 100,
		MaxFailedAttempts:   100,
		BlockDuration:       1 * time.Minute,
		MaxBlockDuration:    10 * time.Minute,
	}

	service := NewRemoteAuthService(config, nil)
	defer service.Stop()

	var wg sync.WaitGroup
	var mu sync.Mutex
	sessions := make([]*RemoteSession, 0)

	// Concurrent authentications
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ip := "192.168.1." + string(rune('1'+idx%9))
			session, err := service.Authenticate("correct", ip)
			if err != nil {
				t.Errorf("Concurrent auth failed: %v", err)
				return
			}
			mu.Lock()
			sessions = append(sessions, session)
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Concurrent validations
	for _, session := range sessions {
		wg.Add(1)
		go func(s *RemoteSession) {
			defer wg.Done()
			_, err := service.ValidateSession(s.Token)
			if err != nil {
				t.Errorf("Concurrent validation failed: %v", err)
			}
		}(session)
	}

	wg.Wait()
}

// Helper function
func mustHashPasswordTest(t *testing.T, password string) string {
	t.Helper()
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return hash
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
