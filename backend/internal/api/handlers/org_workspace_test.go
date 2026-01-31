package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	_ "github.com/mattn/go-sqlite3"

	"github.com/jacklau/prism/internal/database/repository"
)

func setupTestWorkspaceDB(t *testing.T) *sql.DB {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "test_workspace_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	db, err := sql.Open("sqlite3", tmpFile.Name()+"?_foreign_keys=on")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create the org_workspaces table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS org_workspaces (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			organization_id TEXT NOT NULL,
			github_repository_name TEXT,
			worker_id TEXT,
			current_branch TEXT,
			slack_channel_id TEXT,
			slack_message_ts TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create org_workspaces table: %v", err)
	}

	// Create the organizations table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS organizations (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			workos_organization_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create organizations table: %v", err)
	}

	// Create the organization_members table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS organization_members (
			id TEXT PRIMARY KEY,
			organization_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			role TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (organization_id) REFERENCES organizations(id)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create organization_members table: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
		os.Remove(tmpFile.Name())
	})

	return db
}

func createTestFiberApp(handler *OrgWorkspaceHandler) *fiber.App {
	app := fiber.New()

	// Simulate auth middleware by setting userID in locals
	app.Use(func(c *fiber.Ctx) error {
		userID := c.Get("X-User-ID")
		if userID != "" {
			c.Locals("userID", userID)
		}
		return c.Next()
	})

	// Register routes matching the actual API structure
	workspaces := app.Group("/api/v1/organizations/:orgId/workspaces")
	workspaces.Get("/", handler.List)
	workspaces.Post("/", handler.Create)
	workspaces.Get("/:id", handler.Get)
	workspaces.Put("/:id", handler.Update)
	workspaces.Delete("/:id", handler.Delete)
	workspaces.Patch("/:id/branch", handler.UpdateBranch)

	return app
}

func TestOrgWorkspaceHandler_Create(t *testing.T) {
	db := setupTestWorkspaceDB(t)
	wsRepo := repository.NewOrgWorkspaceRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)

	// Create test organization and add admin member
	org, _ := orgRepo.Create("Test Org", "")
	orgRepo.AddMember(org.ID, "admin-user", "admin")
	orgRepo.AddMember(org.ID, "member-user", "member")

	handler := NewOrgWorkspaceHandler(wsRepo, orgRepo)
	app := createTestFiberApp(handler)

	t.Run("success", func(t *testing.T) {
		body := `{"name": "Test Workspace", "github_repository_name": "test/repo"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/organizations/"+org.ID+"/workspaces", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "admin-user")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 201, got %d: %s", resp.StatusCode, body)
		}

		var result WorkspaceResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result.Name != "Test Workspace" {
			t.Errorf("Expected name 'Test Workspace', got '%s'", result.Name)
		}
		if result.OrganizationID != org.ID {
			t.Errorf("Expected org ID '%s', got '%s'", org.ID, result.OrganizationID)
		}
	})

	t.Run("unauthorized without auth", func(t *testing.T) {
		body := `{"name": "Test Workspace"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/organizations/"+org.ID+"/workspaces", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", resp.StatusCode)
		}
	})

	t.Run("forbidden for non-admin", func(t *testing.T) {
		body := `{"name": "Test Workspace"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/organizations/"+org.ID+"/workspaces", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "member-user")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", resp.StatusCode)
		}
	})

	t.Run("name required", func(t *testing.T) {
		body := `{"github_repository_name": "test/repo"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/organizations/"+org.ID+"/workspaces", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "admin-user")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("duplicate name conflict", func(t *testing.T) {
		// First create should succeed
		body := `{"name": "Duplicate Workspace"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/organizations/"+org.ID+"/workspaces", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "admin-user")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()

		// Second create with same name should fail
		req = httptest.NewRequest(http.MethodPost, "/api/v1/organizations/"+org.ID+"/workspaces", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "admin-user")

		resp, err = app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Expected status 409 for duplicate name, got %d", resp.StatusCode)
		}
	})
}

func TestOrgWorkspaceHandler_Get(t *testing.T) {
	db := setupTestWorkspaceDB(t)
	wsRepo := repository.NewOrgWorkspaceRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)

	// Create test organization and members
	org, _ := orgRepo.Create("Test Org", "")
	orgRepo.AddMember(org.ID, "member-user", "member")

	// Create test workspace
	ws, _ := wsRepo.Create(&repository.OrgWorkspace{
		Name:           "Test Workspace",
		OrganizationID: org.ID,
	})

	handler := NewOrgWorkspaceHandler(wsRepo, orgRepo)
	app := createTestFiberApp(handler)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/"+org.ID+"/workspaces/"+ws.ID, nil)
		req.Header.Set("X-User-ID", "member-user")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, body)
		}

		var result WorkspaceResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result.ID != ws.ID {
			t.Errorf("Expected ID '%s', got '%s'", ws.ID, result.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/"+org.ID+"/workspaces/nonexistent", nil)
		req.Header.Set("X-User-ID", "member-user")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("wrong organization", func(t *testing.T) {
		// Create another organization
		otherOrg, _ := orgRepo.Create("Other Org", "")
		orgRepo.AddMember(otherOrg.ID, "member-user", "member")

		// Try to access workspace from org1 via org2's route
		req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/"+otherOrg.ID+"/workspaces/"+ws.ID, nil)
		req.Header.Set("X-User-ID", "member-user")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("forbidden for non-member", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/"+org.ID+"/workspaces/"+ws.ID, nil)
		req.Header.Set("X-User-ID", "other-user")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", resp.StatusCode)
		}
	})
}

func TestOrgWorkspaceHandler_List(t *testing.T) {
	db := setupTestWorkspaceDB(t)
	wsRepo := repository.NewOrgWorkspaceRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)

	// Create test organization and members
	org, _ := orgRepo.Create("Test Org", "")
	orgRepo.AddMember(org.ID, "member-user", "member")

	// Create test workspaces
	for i := 0; i < 5; i++ {
		wsRepo.Create(&repository.OrgWorkspace{
			Name:           "Workspace " + string(rune('A'+i)),
			OrganizationID: org.ID,
		})
	}

	handler := NewOrgWorkspaceHandler(wsRepo, orgRepo)
	app := createTestFiberApp(handler)

	t.Run("list all", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/"+org.ID+"/workspaces", nil)
		req.Header.Set("X-User-ID", "member-user")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		var result struct {
			Workspaces []*WorkspaceResponse `json:"workspaces"`
			Total      int                  `json:"total"`
			Limit      int                  `json:"limit"`
			Offset     int                  `json:"offset"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result.Total != 5 {
			t.Errorf("Expected total 5, got %d", result.Total)
		}
		if len(result.Workspaces) != 5 {
			t.Errorf("Expected 5 workspaces, got %d", len(result.Workspaces))
		}
	})

	t.Run("pagination", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/"+org.ID+"/workspaces?limit=2&offset=1", nil)
		req.Header.Set("X-User-ID", "member-user")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		var result struct {
			Workspaces []*WorkspaceResponse `json:"workspaces"`
			Total      int                  `json:"total"`
			Limit      int                  `json:"limit"`
			Offset     int                  `json:"offset"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result.Total != 5 {
			t.Errorf("Expected total 5, got %d", result.Total)
		}
		if len(result.Workspaces) != 2 {
			t.Errorf("Expected 2 workspaces with limit, got %d", len(result.Workspaces))
		}
		if result.Limit != 2 {
			t.Errorf("Expected limit 2, got %d", result.Limit)
		}
		if result.Offset != 1 {
			t.Errorf("Expected offset 1, got %d", result.Offset)
		}
	})
}

func TestOrgWorkspaceHandler_Update(t *testing.T) {
	db := setupTestWorkspaceDB(t)
	wsRepo := repository.NewOrgWorkspaceRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)

	// Create test organization and members
	org, _ := orgRepo.Create("Test Org", "")
	orgRepo.AddMember(org.ID, "admin-user", "admin")
	orgRepo.AddMember(org.ID, "member-user", "member")

	// Create test workspace
	ws, _ := wsRepo.Create(&repository.OrgWorkspace{
		Name:           "Original Name",
		OrganizationID: org.ID,
	})

	handler := NewOrgWorkspaceHandler(wsRepo, orgRepo)
	app := createTestFiberApp(handler)

	t.Run("success", func(t *testing.T) {
		body := `{"name": "Updated Name", "github_repository_name": "updated/repo"}`
		req := httptest.NewRequest(http.MethodPut, "/api/v1/organizations/"+org.ID+"/workspaces/"+ws.ID, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "admin-user")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, body)
		}

		var result WorkspaceResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result.Name != "Updated Name" {
			t.Errorf("Expected name 'Updated Name', got '%s'", result.Name)
		}
	})

	t.Run("forbidden for non-admin", func(t *testing.T) {
		body := `{"name": "Hacked Name"}`
		req := httptest.NewRequest(http.MethodPut, "/api/v1/organizations/"+org.ID+"/workspaces/"+ws.ID, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "member-user")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", resp.StatusCode)
		}
	})

	t.Run("name uniqueness on update", func(t *testing.T) {
		// Create another workspace
		ws2, _ := wsRepo.Create(&repository.OrgWorkspace{
			Name:           "Existing Name",
			OrganizationID: org.ID,
		})

		// Try to rename first workspace to existing name
		body := `{"name": "Existing Name"}`
		req := httptest.NewRequest(http.MethodPut, "/api/v1/organizations/"+org.ID+"/workspaces/"+ws.ID, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "admin-user")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusConflict {
			t.Errorf("Expected status 409 for duplicate name, got %d", resp.StatusCode)
		}

		// Clean up
		wsRepo.Delete(ws2.ID)
	})
}

func TestOrgWorkspaceHandler_Delete(t *testing.T) {
	db := setupTestWorkspaceDB(t)
	wsRepo := repository.NewOrgWorkspaceRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)

	// Create test organization and members
	org, _ := orgRepo.Create("Test Org", "")
	orgRepo.AddMember(org.ID, "admin-user", "admin")
	orgRepo.AddMember(org.ID, "member-user", "member")

	handler := NewOrgWorkspaceHandler(wsRepo, orgRepo)
	app := createTestFiberApp(handler)

	t.Run("success", func(t *testing.T) {
		// Create workspace to delete
		ws, _ := wsRepo.Create(&repository.OrgWorkspace{
			Name:           "To Delete",
			OrganizationID: org.ID,
		})

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/organizations/"+org.ID+"/workspaces/"+ws.ID, nil)
		req.Header.Set("X-User-ID", "admin-user")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 204, got %d: %s", resp.StatusCode, body)
		}

		// Verify deleted
		deleted, _ := wsRepo.GetByID(ws.ID)
		if deleted != nil {
			t.Error("Workspace should have been deleted")
		}
	})

	t.Run("forbidden for non-admin", func(t *testing.T) {
		ws, _ := wsRepo.Create(&repository.OrgWorkspace{
			Name:           "Protected",
			OrganizationID: org.ID,
		})

		req := httptest.NewRequest(http.MethodDelete, "/api/v1/organizations/"+org.ID+"/workspaces/"+ws.ID, nil)
		req.Header.Set("X-User-ID", "member-user")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/organizations/"+org.ID+"/workspaces/nonexistent", nil)
		req.Header.Set("X-User-ID", "admin-user")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})
}

func TestOrgWorkspaceHandler_UpdateBranch(t *testing.T) {
	db := setupTestWorkspaceDB(t)
	wsRepo := repository.NewOrgWorkspaceRepository(db)
	orgRepo := repository.NewOrganizationRepository(db)

	// Create test organization and members
	org, _ := orgRepo.Create("Test Org", "")
	orgRepo.AddMember(org.ID, "admin-user", "admin")

	// Create test workspace
	ws, _ := wsRepo.Create(&repository.OrgWorkspace{
		Name:           "Test Workspace",
		OrganizationID: org.ID,
		CurrentBranch:  "main",
	})

	handler := NewOrgWorkspaceHandler(wsRepo, orgRepo)
	app := createTestFiberApp(handler)

	t.Run("success", func(t *testing.T) {
		body := `{"branch": "feature/new-branch"}`
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/organizations/"+org.ID+"/workspaces/"+ws.ID+"/branch", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "admin-user")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Expected status 200, got %d: %s", resp.StatusCode, body)
		}

		var result WorkspaceResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if result.CurrentBranch != "feature/new-branch" {
			t.Errorf("Expected branch 'feature/new-branch', got '%s'", result.CurrentBranch)
		}
	})
}

func TestOrgWorkspaceRepository_Pagination(t *testing.T) {
	db := setupTestWorkspaceDB(t)
	repo := repository.NewOrgWorkspaceRepository(db)

	orgID := "test-org"

	// Create 10 workspaces
	for i := 0; i < 10; i++ {
		repo.Create(&repository.OrgWorkspace{
			Name:           "Workspace " + string(rune('A'+i)),
			OrganizationID: orgID,
		})
	}

	t.Run("count", func(t *testing.T) {
		count, err := repo.CountByOrganizationID(orgID)
		if err != nil {
			t.Fatalf("CountByOrganizationID failed: %v", err)
		}
		if count != 10 {
			t.Errorf("Expected count 10, got %d", count)
		}
	})

	t.Run("pagination", func(t *testing.T) {
		workspaces, err := repo.ListByOrganizationIDWithPagination(orgID, 3, 2)
		if err != nil {
			t.Fatalf("ListByOrganizationIDWithPagination failed: %v", err)
		}
		if len(workspaces) != 3 {
			t.Errorf("Expected 3 workspaces, got %d", len(workspaces))
		}
	})
}

func TestOrgWorkspaceRepository_GetByName(t *testing.T) {
	db := setupTestWorkspaceDB(t)
	repo := repository.NewOrgWorkspaceRepository(db)

	orgID := "test-org"
	ws, _ := repo.Create(&repository.OrgWorkspace{
		Name:           "Unique Name",
		OrganizationID: orgID,
	})

	t.Run("found", func(t *testing.T) {
		found, err := repo.GetByName(orgID, "Unique Name")
		if err != nil {
			t.Fatalf("GetByName failed: %v", err)
		}
		if found == nil {
			t.Fatal("Expected to find workspace")
		}
		if found.ID != ws.ID {
			t.Errorf("Expected ID '%s', got '%s'", ws.ID, found.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		found, err := repo.GetByName(orgID, "Nonexistent")
		if err != nil {
			t.Fatalf("GetByName failed: %v", err)
		}
		if found != nil {
			t.Error("Expected nil for nonexistent name")
		}
	})

	t.Run("wrong org", func(t *testing.T) {
		found, err := repo.GetByName("other-org", "Unique Name")
		if err != nil {
			t.Fatalf("GetByName failed: %v", err)
		}
		if found != nil {
			t.Error("Expected nil for different org")
		}
	})
}

func TestOrgWorkspaceRepository_BulkDelete(t *testing.T) {
	db := setupTestWorkspaceDB(t)
	repo := repository.NewOrgWorkspaceRepository(db)

	orgID := "test-org"

	// Create workspaces
	var ids []string
	for i := 0; i < 5; i++ {
		ws, _ := repo.Create(&repository.OrgWorkspace{
			Name:           "Workspace " + string(rune('A'+i)),
			OrganizationID: orgID,
		})
		ids = append(ids, ws.ID)
	}

	t.Run("bulk delete subset", func(t *testing.T) {
		// Delete first 3
		err := repo.BulkDelete(ids[:3])
		if err != nil {
			t.Fatalf("BulkDelete failed: %v", err)
		}

		// Verify count
		count, _ := repo.CountByOrganizationID(orgID)
		if count != 2 {
			t.Errorf("Expected 2 remaining, got %d", count)
		}
	})

	t.Run("bulk delete empty", func(t *testing.T) {
		err := repo.BulkDelete([]string{})
		if err != nil {
			t.Fatalf("BulkDelete with empty slice should not fail: %v", err)
		}
	})
}
