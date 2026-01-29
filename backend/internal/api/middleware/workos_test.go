package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/security"
)

func setupTestWorkOSService(t *testing.T) *security.WorkOSService {
	svc, err := security.NewWorkOSService("client_123", "secret_456", "http://localhost:3000/callback", "12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("Failed to create WorkOS service: %v", err)
	}
	return svc
}

func createValidSessionCookie(t *testing.T, svc *security.WorkOSService, expiresAt time.Time) string {
	session := &security.WorkOSSession{
		ID:             "session_abc123",
		UserID:         "user_xyz789",
		OrganizationID: "org_def456",
		ConnectionID:   "conn_ghi012",
		ExpiresAt:      expiresAt,
		CreatedAt:      time.Now(),
	}
	cookie, err := svc.CreateSession(session)
	if err != nil {
		t.Fatalf("Failed to create session cookie: %v", err)
	}
	return cookie
}

func TestWorkOSSessionMiddleware_NoSession(t *testing.T) {
	svc := setupTestWorkOSService(t)

	app := fiber.New()
	app.Use(WorkOSSessionMiddleware(svc))
	app.Get("/test", func(c *fiber.Ctx) error {
		orgID := GetOrganizationID(c)
		return c.JSON(fiber.Map{"org_id": orgID})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if result["org_id"] != "" {
		t.Errorf("Expected empty org_id, got %v", result["org_id"])
	}
}

func TestWorkOSSessionMiddleware_ValidSession(t *testing.T) {
	svc := setupTestWorkOSService(t)

	app := fiber.New()
	app.Use(WorkOSSessionMiddleware(svc))
	app.Get("/test", func(c *fiber.Ctx) error {
		orgID := GetOrganizationID(c)
		sessionID := GetSSOSessionID(c)
		userID := GetSSOUserID(c)
		connID := GetConnectionID(c)
		hasOrg := HasOrganizationContext(c)
		return c.JSON(fiber.Map{
			"org_id":     orgID,
			"session_id": sessionID,
			"user_id":    userID,
			"conn_id":    connID,
			"has_org":    hasOrg,
		})
	})

	cookie := createValidSessionCookie(t, svc, time.Now().Add(24*time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "wos-session", Value: cookie})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if result["org_id"] != "org_def456" {
		t.Errorf("Expected org_id 'org_def456', got %v", result["org_id"])
	}
	if result["session_id"] != "session_abc123" {
		t.Errorf("Expected session_id 'session_abc123', got %v", result["session_id"])
	}
	if result["user_id"] != "user_xyz789" {
		t.Errorf("Expected user_id 'user_xyz789', got %v", result["user_id"])
	}
	if result["conn_id"] != "conn_ghi012" {
		t.Errorf("Expected conn_id 'conn_ghi012', got %v", result["conn_id"])
	}
	if result["has_org"] != true {
		t.Errorf("Expected has_org true, got %v", result["has_org"])
	}
}

func TestWorkOSSessionMiddleware_ExpiredSession(t *testing.T) {
	svc := setupTestWorkOSService(t)

	app := fiber.New()
	app.Use(WorkOSSessionMiddleware(svc))
	app.Get("/test", func(c *fiber.Ctx) error {
		orgID := GetOrganizationID(c)
		return c.JSON(fiber.Map{"org_id": orgID})
	})

	// Create expired session
	cookie := createValidSessionCookie(t, svc, time.Now().Add(-1*time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "wos-session", Value: cookie})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	// Expired session should not set org context
	if result["org_id"] != "" {
		t.Errorf("Expected empty org_id for expired session, got %v", result["org_id"])
	}
}

func TestWorkOSSessionMiddleware_InvalidSession(t *testing.T) {
	svc := setupTestWorkOSService(t)

	app := fiber.New()
	app.Use(WorkOSSessionMiddleware(svc))
	app.Get("/test", func(c *fiber.Ctx) error {
		orgID := GetOrganizationID(c)
		return c.JSON(fiber.Map{"org_id": orgID})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "wos-session", Value: "invalid-cookie-value"})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	// Invalid session should not set org context
	if result["org_id"] != "" {
		t.Errorf("Expected empty org_id for invalid session, got %v", result["org_id"])
	}
}

func TestRequireOrganization_WithOrg(t *testing.T) {
	svc := setupTestWorkOSService(t)

	app := fiber.New()
	app.Use(WorkOSSessionMiddleware(svc))
	app.Get("/test", RequireOrganization(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	cookie := createValidSessionCookie(t, svc, time.Now().Add(24*time.Hour))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "wos-session", Value: cookie})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestRequireOrganization_WithoutOrg(t *testing.T) {
	svc := setupTestWorkOSService(t)

	app := fiber.New()
	app.Use(WorkOSSessionMiddleware(svc))
	app.Get("/test", RequireOrganization(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if result["error"] != "Organization context required" {
		t.Errorf("Expected error message 'Organization context required', got %v", result["error"])
	}
}

func TestContextHelpers_NoContext(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		orgID := GetOrganizationID(c)
		sessionID := GetSSOSessionID(c)
		userID := GetSSOUserID(c)
		connID := GetConnectionID(c)
		hasOrg := HasOrganizationContext(c)

		if orgID != "" || sessionID != "" || userID != "" || connID != "" || hasOrg {
			return c.Status(500).JSON(fiber.Map{"error": "expected empty values"})
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
