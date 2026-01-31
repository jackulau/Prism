package repository

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupBuildHistoryTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Create a temporary database
	tmpFile, err := os.CreateTemp("", "test_build_history_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	db, err := sql.Open("sqlite3", tmpFile.Name()+"?_foreign_keys=on")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create the users table first (for foreign key)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Insert a test user
	_, err = db.Exec(`INSERT INTO users (id, email, password_hash) VALUES ('user-123', 'test@example.com', 'hash')`)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Create the build_history table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS build_history (
			id TEXT PRIMARY KEY,
			workspace_id TEXT,
			org_workspace_id TEXT,
			user_id TEXT NOT NULL,
			command TEXT NOT NULL,
			status TEXT NOT NULL,
			exit_code INTEGER,
			started_at DATETIME NOT NULL,
			completed_at DATETIME,
			duration_ms INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create build_history table: %v", err)
	}

	// Create the build_logs table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS build_logs (
			id TEXT PRIMARY KEY,
			build_id TEXT NOT NULL,
			stream TEXT NOT NULL,
			content TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (build_id) REFERENCES build_history(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create build_logs table: %v", err)
	}

	// Create indexes
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_build_history_user ON build_history(user_id)`)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_build_logs_build ON build_logs(build_id)`)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
		os.Remove(tmpFile.Name())
	})

	return db
}

func TestBuildHistoryRepository_Create(t *testing.T) {
	db := setupBuildHistoryTestDB(t)
	repo := NewBuildHistoryRepository(db)

	build := &BuildHistory{
		UserID:    "user-123",
		Command:   "npm run build",
		Status:    BuildStatusPending,
		StartedAt: time.Now(),
	}

	err := repo.Create(build)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if build.ID == "" {
		t.Error("Expected ID to be set after create")
	}
}

func TestBuildHistoryRepository_GetByID(t *testing.T) {
	db := setupBuildHistoryTestDB(t)
	repo := NewBuildHistoryRepository(db)

	// Create a build
	build := &BuildHistory{
		UserID:    "user-123",
		Command:   "npm run build",
		Status:    BuildStatusRunning,
		StartedAt: time.Now(),
	}
	err := repo.Create(build)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Get by ID
	retrieved, err := repo.GetByID(build.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected build to be found")
	}
	if retrieved.ID != build.ID {
		t.Errorf("Expected ID %s, got %s", build.ID, retrieved.ID)
	}
	if retrieved.Command != "npm run build" {
		t.Errorf("Expected command 'npm run build', got %s", retrieved.Command)
	}
	if retrieved.Status != BuildStatusRunning {
		t.Errorf("Expected status running, got %s", retrieved.Status)
	}
}

func TestBuildHistoryRepository_GetByID_NotFound(t *testing.T) {
	db := setupBuildHistoryTestDB(t)
	repo := NewBuildHistoryRepository(db)

	build, err := repo.GetByID("nonexistent")
	if err != nil {
		t.Fatalf("GetByID should not return error for not found: %v", err)
	}
	if build != nil {
		t.Error("Expected nil for nonexistent build")
	}
}

func TestBuildHistoryRepository_Update(t *testing.T) {
	db := setupBuildHistoryTestDB(t)
	repo := NewBuildHistoryRepository(db)

	// Create a build
	build := &BuildHistory{
		UserID:    "user-123",
		Command:   "npm run build",
		Status:    BuildStatusRunning,
		StartedAt: time.Now(),
	}
	err := repo.Create(build)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update it
	now := time.Now()
	exitCode := 0
	duration := int64(5000)
	build.Status = BuildStatusSuccess
	build.CompletedAt = &now
	build.ExitCode = &exitCode
	build.DurationMs = &duration

	err = repo.Update(build)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify
	retrieved, err := repo.GetByID(build.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Status != BuildStatusSuccess {
		t.Errorf("Expected status success, got %s", retrieved.Status)
	}
	if retrieved.ExitCode == nil || *retrieved.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %v", retrieved.ExitCode)
	}
}

func TestBuildHistoryRepository_Delete(t *testing.T) {
	db := setupBuildHistoryTestDB(t)
	repo := NewBuildHistoryRepository(db)

	// Create a build
	build := &BuildHistory{
		UserID:    "user-123",
		Command:   "npm run build",
		Status:    BuildStatusSuccess,
		StartedAt: time.Now(),
	}
	err := repo.Create(build)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Delete it
	err = repo.Delete(build.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	retrieved, err := repo.GetByID(build.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved != nil {
		t.Error("Expected nil after deletion")
	}
}

func TestBuildHistoryRepository_Delete_NotFound(t *testing.T) {
	db := setupBuildHistoryTestDB(t)
	repo := NewBuildHistoryRepository(db)

	err := repo.Delete("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent delete")
	}
}

func TestBuildHistoryRepository_ListByUserID(t *testing.T) {
	db := setupBuildHistoryTestDB(t)
	repo := NewBuildHistoryRepository(db)

	// Create multiple builds
	for i := 0; i < 5; i++ {
		build := &BuildHistory{
			UserID:    "user-123",
			Command:   "npm run build",
			Status:    BuildStatusSuccess,
			StartedAt: time.Now(),
		}
		err := repo.Create(build)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// List with pagination
	builds, err := repo.ListByUserID("user-123", 3, 0)
	if err != nil {
		t.Fatalf("ListByUserID failed: %v", err)
	}
	if len(builds) != 3 {
		t.Errorf("Expected 3 builds, got %d", len(builds))
	}

	// List with offset
	builds, err = repo.ListByUserID("user-123", 3, 3)
	if err != nil {
		t.Fatalf("ListByUserID failed: %v", err)
	}
	if len(builds) != 2 {
		t.Errorf("Expected 2 builds, got %d", len(builds))
	}
}

func TestBuildHistoryRepository_ListByWorkspaceID(t *testing.T) {
	db := setupBuildHistoryTestDB(t)
	repo := NewBuildHistoryRepository(db)

	workspaceID := "workspace-456"

	// Create builds with workspace
	for i := 0; i < 3; i++ {
		build := &BuildHistory{
			UserID:      "user-123",
			WorkspaceID: &workspaceID,
			Command:     "npm run build",
			Status:      BuildStatusSuccess,
			StartedAt:   time.Now(),
		}
		err := repo.Create(build)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// Create a build without workspace
	build := &BuildHistory{
		UserID:    "user-123",
		Command:   "npm run build",
		Status:    BuildStatusSuccess,
		StartedAt: time.Now(),
	}
	err := repo.Create(build)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// List by workspace
	builds, err := repo.ListByWorkspaceID(workspaceID, 10, 0)
	if err != nil {
		t.Fatalf("ListByWorkspaceID failed: %v", err)
	}
	if len(builds) != 3 {
		t.Errorf("Expected 3 builds for workspace, got %d", len(builds))
	}
}

func TestBuildHistoryRepository_CountByUserID(t *testing.T) {
	db := setupBuildHistoryTestDB(t)
	repo := NewBuildHistoryRepository(db)

	// Create builds
	for i := 0; i < 7; i++ {
		build := &BuildHistory{
			UserID:    "user-123",
			Command:   "npm run build",
			Status:    BuildStatusSuccess,
			StartedAt: time.Now(),
		}
		err := repo.Create(build)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	count, err := repo.CountByUserID("user-123")
	if err != nil {
		t.Fatalf("CountByUserID failed: %v", err)
	}
	if count != 7 {
		t.Errorf("Expected count 7, got %d", count)
	}
}

func TestBuildHistoryRepository_AppendLog(t *testing.T) {
	db := setupBuildHistoryTestDB(t)
	repo := NewBuildHistoryRepository(db)

	// Create a build
	build := &BuildHistory{
		UserID:    "user-123",
		Command:   "npm run build",
		Status:    BuildStatusRunning,
		StartedAt: time.Now(),
	}
	err := repo.Create(build)
	if err != nil {
		t.Fatalf("Create build failed: %v", err)
	}

	// Append logs
	log1 := &BuildLog{
		BuildID: build.ID,
		Stream:  LogStreamStdout,
		Content: "Building...",
	}
	err = repo.AppendLog(log1)
	if err != nil {
		t.Fatalf("AppendLog failed: %v", err)
	}

	log2 := &BuildLog{
		BuildID: build.ID,
		Stream:  LogStreamStderr,
		Content: "Warning: deprecated",
	}
	err = repo.AppendLog(log2)
	if err != nil {
		t.Fatalf("AppendLog failed: %v", err)
	}

	if log1.ID == "" {
		t.Error("Expected log ID to be set")
	}
}

func TestBuildHistoryRepository_GetLogs(t *testing.T) {
	db := setupBuildHistoryTestDB(t)
	repo := NewBuildHistoryRepository(db)

	// Create a build
	build := &BuildHistory{
		UserID:    "user-123",
		Command:   "npm run build",
		Status:    BuildStatusRunning,
		StartedAt: time.Now(),
	}
	err := repo.Create(build)
	if err != nil {
		t.Fatalf("Create build failed: %v", err)
	}

	// Append logs
	for i := 0; i < 3; i++ {
		log := &BuildLog{
			BuildID: build.ID,
			Stream:  LogStreamStdout,
			Content: "Log line",
		}
		err = repo.AppendLog(log)
		if err != nil {
			t.Fatalf("AppendLog failed: %v", err)
		}
	}

	// Get logs
	logs, err := repo.GetLogs(build.ID)
	if err != nil {
		t.Fatalf("GetLogs failed: %v", err)
	}
	if len(logs) != 3 {
		t.Errorf("Expected 3 logs, got %d", len(logs))
	}
}

func TestBuildHistoryRepository_GetLogsSince(t *testing.T) {
	db := setupBuildHistoryTestDB(t)
	repo := NewBuildHistoryRepository(db)

	// Create a build
	build := &BuildHistory{
		UserID:    "user-123",
		Command:   "npm run build",
		Status:    BuildStatusRunning,
		StartedAt: time.Now(),
	}
	err := repo.Create(build)
	if err != nil {
		t.Fatalf("Create build failed: %v", err)
	}

	// Record timestamp before adding logs
	before := time.Now()
	time.Sleep(10 * time.Millisecond)

	// Append logs
	for i := 0; i < 3; i++ {
		log := &BuildLog{
			BuildID: build.ID,
			Stream:  LogStreamStdout,
			Content: "Log line",
		}
		err = repo.AppendLog(log)
		if err != nil {
			t.Fatalf("AppendLog failed: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Get logs since before
	logs, err := repo.GetLogsSince(build.ID, before)
	if err != nil {
		t.Fatalf("GetLogsSince failed: %v", err)
	}
	if len(logs) != 3 {
		t.Errorf("Expected 3 logs since before, got %d", len(logs))
	}

	// Get logs since now (should be empty)
	logs, err = repo.GetLogsSince(build.ID, time.Now())
	if err != nil {
		t.Fatalf("GetLogsSince failed: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("Expected 0 logs since now, got %d", len(logs))
	}
}

func TestBuildHistoryRepository_UpdateStatus(t *testing.T) {
	db := setupBuildHistoryTestDB(t)
	repo := NewBuildHistoryRepository(db)

	// Create a build
	build := &BuildHistory{
		UserID:    "user-123",
		Command:   "npm run build",
		Status:    BuildStatusRunning,
		StartedAt: time.Now(),
	}
	err := repo.Create(build)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update status
	now := time.Now()
	exitCode := 1
	duration := int64(3000)
	err = repo.UpdateStatus(build.ID, BuildStatusFailed, &exitCode, &now, &duration)
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	// Verify
	retrieved, err := repo.GetByID(build.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Status != BuildStatusFailed {
		t.Errorf("Expected status failed, got %s", retrieved.Status)
	}
	if retrieved.ExitCode == nil || *retrieved.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got %v", retrieved.ExitCode)
	}
	if retrieved.DurationMs == nil || *retrieved.DurationMs != 3000 {
		t.Errorf("Expected duration 3000, got %v", retrieved.DurationMs)
	}
}

func TestBuildHistoryRepository_GetRunningBuilds(t *testing.T) {
	db := setupBuildHistoryTestDB(t)
	repo := NewBuildHistoryRepository(db)

	// Create builds with different statuses
	statuses := []BuildStatus{BuildStatusRunning, BuildStatusRunning, BuildStatusSuccess, BuildStatusFailed}
	for _, status := range statuses {
		build := &BuildHistory{
			UserID:    "user-123",
			Command:   "npm run build",
			Status:    status,
			StartedAt: time.Now(),
		}
		err := repo.Create(build)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// Get running builds
	builds, err := repo.GetRunningBuilds("user-123")
	if err != nil {
		t.Fatalf("GetRunningBuilds failed: %v", err)
	}
	if len(builds) != 2 {
		t.Errorf("Expected 2 running builds, got %d", len(builds))
	}
	for _, b := range builds {
		if b.Status != BuildStatusRunning {
			t.Errorf("Expected all builds to be running, got %s", b.Status)
		}
	}
}

func TestBuildHistoryRepository_DeleteLogsByBuildID(t *testing.T) {
	db := setupBuildHistoryTestDB(t)
	repo := NewBuildHistoryRepository(db)

	// Create a build
	build := &BuildHistory{
		UserID:    "user-123",
		Command:   "npm run build",
		Status:    BuildStatusRunning,
		StartedAt: time.Now(),
	}
	err := repo.Create(build)
	if err != nil {
		t.Fatalf("Create build failed: %v", err)
	}

	// Append logs
	for i := 0; i < 3; i++ {
		log := &BuildLog{
			BuildID: build.ID,
			Stream:  LogStreamStdout,
			Content: "Log line",
		}
		err = repo.AppendLog(log)
		if err != nil {
			t.Fatalf("AppendLog failed: %v", err)
		}
	}

	// Delete logs
	err = repo.DeleteLogsByBuildID(build.ID)
	if err != nil {
		t.Fatalf("DeleteLogsByBuildID failed: %v", err)
	}

	// Verify logs are gone
	logs, err := repo.GetLogs(build.ID)
	if err != nil {
		t.Fatalf("GetLogs failed: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("Expected 0 logs after deletion, got %d", len(logs))
	}
}

func TestBuildHistoryRepository_CascadeDelete(t *testing.T) {
	db := setupBuildHistoryTestDB(t)
	repo := NewBuildHistoryRepository(db)

	// Create a build
	build := &BuildHistory{
		UserID:    "user-123",
		Command:   "npm run build",
		Status:    BuildStatusRunning,
		StartedAt: time.Now(),
	}
	err := repo.Create(build)
	if err != nil {
		t.Fatalf("Create build failed: %v", err)
	}

	// Append logs
	for i := 0; i < 3; i++ {
		log := &BuildLog{
			BuildID: build.ID,
			Stream:  LogStreamStdout,
			Content: "Log line",
		}
		err = repo.AppendLog(log)
		if err != nil {
			t.Fatalf("AppendLog failed: %v", err)
		}
	}

	// Delete build (should cascade delete logs)
	err = repo.Delete(build.ID)
	if err != nil {
		t.Fatalf("Delete build failed: %v", err)
	}

	// Verify logs are gone due to cascade
	logs, err := repo.GetLogs(build.ID)
	if err != nil {
		t.Fatalf("GetLogs failed: %v", err)
	}
	if len(logs) != 0 {
		t.Errorf("Expected 0 logs after cascade delete, got %d", len(logs))
	}
}
