package handlers

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/api/middleware"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/security"
)

// MFAHandler handles MFA-related endpoints
type MFAHandler struct {
	mfaRepo    *repository.MFARepository
	userRepo   *repository.UserRepository
	mfaService *security.MFAService
	jwtService *security.JWTService

	// Rate limiting state
	rateLimitMu  sync.RWMutex
	rateLimitMap map[string]*rateLimitEntry
	lockoutMap   map[string]time.Time
}

type rateLimitEntry struct {
	attempts    int
	windowStart time.Time
}

// Rate limiting constants
const (
	maxAttemptsPerMinute     = 5
	maxAttemptsBeforeLockout = 10
	lockoutDuration          = 10 * time.Minute
	rateLimitWindow          = time.Minute
)

// NewMFAHandler creates a new MFA handler
func NewMFAHandler(
	mfaRepo *repository.MFARepository,
	userRepo *repository.UserRepository,
	mfaService *security.MFAService,
	jwtService *security.JWTService,
) *MFAHandler {
	return &MFAHandler{
		mfaRepo:      mfaRepo,
		userRepo:     userRepo,
		mfaService:   mfaService,
		jwtService:   jwtService,
		rateLimitMap: make(map[string]*rateLimitEntry),
		lockoutMap:   make(map[string]time.Time),
	}
}

// MFASetupRequest represents a request to start MFA setup
type MFASetupRequest struct {
	// No fields needed - uses authenticated user
}

// MFASetupResponse contains the TOTP setup information
type MFASetupResponse struct {
	Secret    string `json:"secret"`
	QRCodeURL string `json:"qr_code_url"`
}

// MFAVerifyRequest represents a request to verify MFA during setup
type MFAVerifyRequest struct {
	Code string `json:"code"`
}

// MFAVerifyResponse contains the backup codes after successful verification
type MFAVerifyResponse struct {
	Enabled     bool     `json:"enabled"`
	BackupCodes []string `json:"backup_codes"`
}

// MFAValidateRequest represents a request to validate MFA during login
type MFAValidateRequest struct {
	MFAToken string `json:"mfa_token"`
	Code     string `json:"code"`
}

// MFAValidateResponse contains the auth tokens after successful MFA validation
type MFAValidateResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserDTO   `json:"user"`
}

// MFAStatusResponse contains the MFA status for a user
type MFAStatusResponse struct {
	Enabled          bool `json:"enabled"`
	BackupCodesCount int  `json:"backup_codes_count,omitempty"`
	SetupPending     bool `json:"setup_pending,omitempty"`
}

// MFADisableRequest represents a request to disable MFA
type MFADisableRequest struct {
	Password string `json:"password"`
	Code     string `json:"code"`
}

// BackupCodesResponse contains regenerated backup codes
type BackupCodesResponse struct {
	BackupCodes []string `json:"backup_codes"`
}

// BackupCodeVerifyRequest represents a request to verify using a backup code
type BackupCodeVerifyRequest struct {
	MFAToken string `json:"mfa_token"`
	Code     string `json:"code"`
}

// Setup handles POST /api/v1/auth/mfa/setup - Start MFA enrollment
func (h *MFAHandler) Setup(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Get user to obtain email
	user, err := h.userRepo.GetByID(userID)
	if err != nil || user == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get user",
		})
	}

	// Check if MFA is already enabled
	existingMFA, err := h.mfaRepo.GetMFAByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to check MFA status",
		})
	}
	if existingMFA != nil && existingMFA.IsEnabled {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "MFA is already enabled",
		})
	}

	// Generate TOTP secret
	setup, err := h.mfaService.GenerateTOTPSecret(user.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate TOTP secret",
		})
	}

	// Store the encrypted secret (pending verification)
	if err := h.mfaRepo.CreateMFASetup(userID, setup.EncryptedSecret, setup.SecretNonce); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to save MFA setup",
		})
	}

	return c.JSON(MFASetupResponse{
		Secret:    setup.Secret,
		QRCodeURL: setup.QRCodeURL,
	})
}

// Verify handles POST /api/v1/auth/mfa/verify - Verify code during setup (enables MFA)
func (h *MFAHandler) Verify(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req MFAVerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if len(req.Code) != 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "code must be 6 digits",
		})
	}

	// Get pending MFA setup
	mfa, err := h.mfaRepo.GetMFAByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get MFA setup",
		})
	}
	if mfa == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "MFA setup not found. Please start setup first.",
		})
	}
	if mfa.IsEnabled {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "MFA is already enabled",
		})
	}

	// Decrypt the secret
	secret, err := h.mfaService.DecryptSecret(mfa.SecretEncrypted, mfa.SecretNonce)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to verify code",
		})
	}

	// Verify the code
	if !h.mfaService.VerifyTOTPCode(secret, req.Code) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid code",
		})
	}

	// Generate backup codes
	backupCodes, err := h.mfaService.GenerateBackupCodes(10)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate backup codes",
		})
	}

	// Enable MFA with backup codes
	if err := h.mfaRepo.EnableMFA(userID, backupCodes.EncryptedCodes, backupCodes.Nonce); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to enable MFA",
		})
	}

	return c.JSON(MFAVerifyResponse{
		Enabled:     true,
		BackupCodes: backupCodes.Codes,
	})
}

// Validate handles POST /api/v1/auth/mfa/validate - Validate MFA code during login
func (h *MFAHandler) Validate(c *fiber.Ctx) error {
	var req MFAValidateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.MFAToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "mfa_token is required",
		})
	}

	if len(req.Code) != 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "code must be 6 digits",
		})
	}

	// Validate the MFA session token
	claims, err := h.jwtService.ValidateMFAToken(req.MFAToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid or expired MFA token",
		})
	}

	// Check rate limiting
	if h.isRateLimited(claims.UserID) {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error": "too many failed attempts. Please try again later.",
		})
	}

	// Get MFA configuration
	mfa, err := h.mfaRepo.GetMFAByUserID(claims.UserID)
	if err != nil || mfa == nil || !mfa.IsEnabled {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "MFA not configured",
		})
	}

	// Decrypt the secret
	secret, err := h.mfaService.DecryptSecret(mfa.SecretEncrypted, mfa.SecretNonce)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to verify code",
		})
	}

	// Verify the code
	valid := h.mfaService.VerifyTOTPCode(secret, req.Code)

	// Log the attempt
	ipAddress := c.IP()
	h.mfaRepo.LogVerificationAttempt(claims.UserID, valid, ipAddress)

	if !valid {
		h.recordFailedAttempt(claims.UserID)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid code",
		})
	}

	// Reset rate limit on success
	h.resetRateLimit(claims.UserID)

	// Generate full auth tokens
	tokens, err := h.jwtService.GenerateTokenPair(claims.UserID, claims.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate tokens",
		})
	}

	// Get user for response
	user, err := h.userRepo.GetByID(claims.UserID)
	if err != nil || user == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get user",
		})
	}

	return c.JSON(MFAValidateResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
		User: UserDTO{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	})
}

// Status handles GET /api/v1/auth/mfa/status - Check if MFA is enabled
func (h *MFAHandler) Status(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	mfa, err := h.mfaRepo.GetMFAByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get MFA status",
		})
	}

	if mfa == nil {
		return c.JSON(MFAStatusResponse{
			Enabled: false,
		})
	}

	response := MFAStatusResponse{
		Enabled:      mfa.IsEnabled,
		SetupPending: !mfa.IsEnabled,
	}

	if mfa.IsEnabled && mfa.BackupCodesEncrypted != nil {
		count, err := h.mfaService.GetBackupCodeCount(mfa.BackupCodesEncrypted, mfa.BackupCodesNonce)
		if err == nil {
			response.BackupCodesCount = count
		}
	}

	return c.JSON(response)
}

// Disable handles POST /api/v1/auth/mfa/disable - Disable MFA
func (h *MFAHandler) Disable(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req MFADisableRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Get user and verify password
	user, err := h.userRepo.GetByID(userID)
	if err != nil || user == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get user",
		})
	}

	if !security.VerifyPassword(req.Password, user.PasswordHash) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid password",
		})
	}

	// Get MFA configuration
	mfa, err := h.mfaRepo.GetMFAByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get MFA status",
		})
	}
	if mfa == nil || !mfa.IsEnabled {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "MFA is not enabled",
		})
	}

	// Verify the TOTP code
	secret, err := h.mfaService.DecryptSecret(mfa.SecretEncrypted, mfa.SecretNonce)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to verify code",
		})
	}

	if !h.mfaService.VerifyTOTPCode(secret, req.Code) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid code",
		})
	}

	// Disable MFA
	if err := h.mfaRepo.DisableMFA(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to disable MFA",
		})
	}

	return c.JSON(fiber.Map{
		"message": "MFA disabled successfully",
	})
}

// RegenerateBackupCodes handles POST /api/v1/auth/mfa/backup-codes - Regenerate backup codes
func (h *MFAHandler) RegenerateBackupCodes(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Get MFA configuration
	mfa, err := h.mfaRepo.GetMFAByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get MFA status",
		})
	}
	if mfa == nil || !mfa.IsEnabled {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "MFA is not enabled",
		})
	}

	// Verify the TOTP code
	secret, err := h.mfaService.DecryptSecret(mfa.SecretEncrypted, mfa.SecretNonce)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to verify code",
		})
	}

	if !h.mfaService.VerifyTOTPCode(secret, req.Code) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid code",
		})
	}

	// Generate new backup codes
	backupCodes, err := h.mfaService.GenerateBackupCodes(10)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate backup codes",
		})
	}

	// Update backup codes in database
	if err := h.mfaRepo.UpdateBackupCodes(userID, backupCodes.EncryptedCodes, backupCodes.Nonce); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to save backup codes",
		})
	}

	return c.JSON(BackupCodesResponse{
		BackupCodes: backupCodes.Codes,
	})
}

// VerifyBackupCode handles POST /api/v1/auth/mfa/backup-codes/verify - Use a backup code
func (h *MFAHandler) VerifyBackupCode(c *fiber.Ctx) error {
	var req BackupCodeVerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.MFAToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "mfa_token is required",
		})
	}

	if req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "backup code is required",
		})
	}

	// Validate the MFA session token
	claims, err := h.jwtService.ValidateMFAToken(req.MFAToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid or expired MFA token",
		})
	}

	// Check rate limiting
	if h.isRateLimited(claims.UserID) {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error": "too many failed attempts. Please try again later.",
		})
	}

	// Get MFA configuration
	mfa, err := h.mfaRepo.GetMFAByUserID(claims.UserID)
	if err != nil || mfa == nil || !mfa.IsEnabled {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "MFA not configured",
		})
	}

	if mfa.BackupCodesEncrypted == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "no backup codes available",
		})
	}

	// Verify and consume the backup code
	remainingEncrypted, remainingNonce, valid, err := h.mfaService.VerifyBackupCode(
		mfa.BackupCodesEncrypted, mfa.BackupCodesNonce, req.Code,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to verify backup code",
		})
	}

	// Log the attempt
	ipAddress := c.IP()
	h.mfaRepo.LogVerificationAttempt(claims.UserID, valid, ipAddress)

	if !valid {
		h.recordFailedAttempt(claims.UserID)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid backup code",
		})
	}

	// Update remaining backup codes
	if err := h.mfaRepo.UpdateBackupCodes(claims.UserID, remainingEncrypted, remainingNonce); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update backup codes",
		})
	}

	// Reset rate limit on success
	h.resetRateLimit(claims.UserID)

	// Generate full auth tokens
	tokens, err := h.jwtService.GenerateTokenPair(claims.UserID, claims.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate tokens",
		})
	}

	// Get user for response
	user, err := h.userRepo.GetByID(claims.UserID)
	if err != nil || user == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get user",
		})
	}

	return c.JSON(MFAValidateResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
		User: UserDTO{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	})
}

// Rate limiting helper methods

func (h *MFAHandler) isRateLimited(userID string) bool {
	h.rateLimitMu.RLock()
	defer h.rateLimitMu.RUnlock()

	// Check lockout
	if lockoutUntil, exists := h.lockoutMap[userID]; exists {
		if time.Now().Before(lockoutUntil) {
			return true
		}
	}

	// Check rate limit
	entry, exists := h.rateLimitMap[userID]
	if !exists {
		return false
	}

	// Reset if window has passed
	if time.Since(entry.windowStart) > rateLimitWindow {
		return false
	}

	return entry.attempts >= maxAttemptsPerMinute
}

func (h *MFAHandler) recordFailedAttempt(userID string) {
	h.rateLimitMu.Lock()
	defer h.rateLimitMu.Unlock()

	entry, exists := h.rateLimitMap[userID]
	if !exists || time.Since(entry.windowStart) > rateLimitWindow {
		h.rateLimitMap[userID] = &rateLimitEntry{
			attempts:    1,
			windowStart: time.Now(),
		}
		return
	}

	entry.attempts++

	// Check for lockout threshold
	if entry.attempts >= maxAttemptsBeforeLockout {
		h.lockoutMap[userID] = time.Now().Add(lockoutDuration)
	}
}

func (h *MFAHandler) resetRateLimit(userID string) {
	h.rateLimitMu.Lock()
	defer h.rateLimitMu.Unlock()

	delete(h.rateLimitMap, userID)
	delete(h.lockoutMap, userID)
}
