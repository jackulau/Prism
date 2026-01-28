package security

import (
	"testing"
)

func TestValidateRemotePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "valid password",
			password: "MyP@ssword123!",
			wantErr:  nil,
		},
		{
			name:     "too short",
			password: "Short1!",
			wantErr:  ErrPasswordTooShort,
		},
		{
			name:     "no uppercase",
			password: "myp@ssword123!",
			wantErr:  ErrPasswordNoUppercase,
		},
		{
			name:     "no lowercase",
			password: "MYP@SSWORD123!",
			wantErr:  ErrPasswordNoLowercase,
		},
		{
			name:     "no digit",
			password: "MyP@sswordTest!",
			wantErr:  ErrPasswordNoDigit,
		},
		{
			name:     "no special character",
			password: "MyPassword12345",
			wantErr:  ErrPasswordNoSpecial,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRemotePassword(tt.password)
			if err != tt.wantErr {
				t.Errorf("ValidateRemotePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHashRemotePassword(t *testing.T) {
	password := "SecureP@ssword123!"

	hash, err := HashRemotePassword(password)
	if err != nil {
		t.Fatalf("HashRemotePassword() error = %v", err)
	}

	if hash == "" {
		t.Error("HashRemotePassword() returned empty hash")
	}

	// Verify the hash works
	if !VerifyRemotePassword(password, hash) {
		t.Error("VerifyRemotePassword() failed to verify correct password")
	}

	// Verify wrong password fails
	if VerifyRemotePassword("WrongPassword123!", hash) {
		t.Error("VerifyRemotePassword() verified incorrect password")
	}
}

func TestHashRemotePasswordValidation(t *testing.T) {
	// Should fail validation before hashing
	_, err := HashRemotePassword("weak")
	if err == nil {
		t.Error("HashRemotePassword() should fail for weak password")
	}
}
