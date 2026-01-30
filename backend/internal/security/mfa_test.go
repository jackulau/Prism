package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMFAService_GenerateTOTPSecret(t *testing.T) {
	encService, err := NewEncryptionService("")
	require.NoError(t, err)

	mfaService := NewMFAService("TestApp", encService)

	setup, err := mfaService.GenerateTOTPSecret("test@example.com")
	require.NoError(t, err)

	// Verify secret is generated
	assert.NotEmpty(t, setup.Secret)
	assert.NotEmpty(t, setup.QRCodeURL)
	assert.NotEmpty(t, setup.EncryptedSecret)
	assert.NotEmpty(t, setup.SecretNonce)

	// Verify QR code URL contains expected parts
	assert.Contains(t, setup.QRCodeURL, "otpauth://totp/")
	assert.Contains(t, setup.QRCodeURL, "TestApp")
	assert.Contains(t, setup.QRCodeURL, "test@example.com")

	// Verify secret can be decrypted
	decryptedSecret, err := mfaService.DecryptSecret(setup.EncryptedSecret, setup.SecretNonce)
	require.NoError(t, err)
	assert.Equal(t, setup.Secret, decryptedSecret)
}

func TestMFAService_VerifyTOTPCode(t *testing.T) {
	encService, err := NewEncryptionService("")
	require.NoError(t, err)

	mfaService := NewMFAService("TestApp", encService)

	setup, err := mfaService.GenerateTOTPSecret("test@example.com")
	require.NoError(t, err)

	// Invalid code should fail
	assert.False(t, mfaService.VerifyTOTPCode(setup.Secret, "000000"))
	assert.False(t, mfaService.VerifyTOTPCode(setup.Secret, "12345"))   // Wrong length
	assert.False(t, mfaService.VerifyTOTPCode(setup.Secret, "1234567")) // Wrong length
}

func TestMFAService_GenerateBackupCodes(t *testing.T) {
	encService, err := NewEncryptionService("")
	require.NoError(t, err)

	mfaService := NewMFAService("TestApp", encService)

	backup, err := mfaService.GenerateBackupCodes(10)
	require.NoError(t, err)

	// Verify 10 codes generated
	assert.Len(t, backup.Codes, 10)
	assert.NotEmpty(t, backup.EncryptedCodes)
	assert.NotEmpty(t, backup.Nonce)

	// Verify all codes are unique and properly formatted
	codeSet := make(map[string]bool)
	for _, code := range backup.Codes {
		assert.Len(t, code, 8, "backup code should be 8 characters")
		assert.False(t, codeSet[code], "backup codes should be unique")
		codeSet[code] = true
	}
}

func TestMFAService_VerifyBackupCode(t *testing.T) {
	encService, err := NewEncryptionService("")
	require.NoError(t, err)

	mfaService := NewMFAService("TestApp", encService)

	backup, err := mfaService.GenerateBackupCodes(3)
	require.NoError(t, err)

	// Valid code should work
	remainingEncrypted, remainingNonce, valid, err := mfaService.VerifyBackupCode(
		backup.EncryptedCodes, backup.Nonce, backup.Codes[0],
	)
	require.NoError(t, err)
	assert.True(t, valid)

	// Count remaining codes
	count, err := mfaService.GetBackupCodeCount(remainingEncrypted, remainingNonce)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	// Same code should not work again
	_, _, valid, err = mfaService.VerifyBackupCode(
		remainingEncrypted, remainingNonce, backup.Codes[0],
	)
	require.NoError(t, err)
	assert.False(t, valid)

	// Invalid code should fail
	_, _, valid, err = mfaService.VerifyBackupCode(
		remainingEncrypted, remainingNonce, "INVALIDX",
	)
	require.NoError(t, err)
	assert.False(t, valid)

	// Second valid code should work
	remainingEncrypted, remainingNonce, valid, err = mfaService.VerifyBackupCode(
		remainingEncrypted, remainingNonce, backup.Codes[1],
	)
	require.NoError(t, err)
	assert.True(t, valid)

	count, err = mfaService.GetBackupCodeCount(remainingEncrypted, remainingNonce)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestMFAService_BackupCodeCaseInsensitive(t *testing.T) {
	encService, err := NewEncryptionService("")
	require.NoError(t, err)

	mfaService := NewMFAService("TestApp", encService)

	backup, err := mfaService.GenerateBackupCodes(1)
	require.NoError(t, err)

	// Lowercase version should work
	lowerCode := ""
	for _, c := range backup.Codes[0] {
		if c >= 'A' && c <= 'Z' {
			lowerCode += string(c + 32)
		} else {
			lowerCode += string(c)
		}
	}

	_, _, valid, err := mfaService.VerifyBackupCode(
		backup.EncryptedCodes, backup.Nonce, lowerCode,
	)
	require.NoError(t, err)
	assert.True(t, valid, "backup code verification should be case-insensitive")
}

func TestMFAService_GetBackupCodeCount(t *testing.T) {
	encService, err := NewEncryptionService("")
	require.NoError(t, err)

	mfaService := NewMFAService("TestApp", encService)

	// Nil should return 0
	count, err := mfaService.GetBackupCodeCount(nil, nil)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Generate codes and count
	backup, err := mfaService.GenerateBackupCodes(5)
	require.NoError(t, err)

	count, err = mfaService.GetBackupCodeCount(backup.EncryptedCodes, backup.Nonce)
	require.NoError(t, err)
	assert.Equal(t, 5, count)
}
