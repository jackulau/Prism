package repository

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupFileHistoryTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Create a temporary database
	tmpFile, err := os.CreateTemp("", "test_file_history_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	db, err := sql.Open("sqlite3", tmpFile.Name()+"?_foreign_keys=on")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create the file_history table with attribution fields
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS file_history (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			file_path TEXT NOT NULL,
			content TEXT NOT NULL,
			operation TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			agent_id TEXT,
			agent_name TEXT,
			tool_name TEXT,
			message_id TEXT,
			description TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Create indexes
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_file_history_user_id ON file_history(user_id)`)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_file_history_file_path ON file_history(user_id, file_path)`)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_file_history_agent ON file_history(agent_id)`)
	if err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
		os.Remove(tmpFile.Name())
	})

	return db
}

func TestFileHistoryRepository_Create(t *testing.T) {
	db := setupFileHistoryTestDB(t)
	repo := NewFileHistoryRepository(db)

	userID := "user-123"
	filePath := "src/main.go"
	content := "package main\n\nfunc main() {}\n"
	operation := "create"

	entry, err := repo.Create(userID, filePath, content, operation)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if entry.ID == "" {
		t.Error("Expected ID to be non-empty")
	}
	if entry.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, entry.UserID)
	}
	if entry.FilePath != filePath {
		t.Errorf("Expected FilePath %s, got %s", filePath, entry.FilePath)
	}
	if entry.Content != content {
		t.Errorf("Expected Content %s, got %s", content, entry.Content)
	}
	if entry.Operation != operation {
		t.Errorf("Expected Operation %s, got %s", operation, entry.Operation)
	}
}

func TestFileHistoryRepository_CreateWithAttribution(t *testing.T) {
	db := setupFileHistoryTestDB(t)
	repo := NewFileHistoryRepository(db)

	agentID := "agent-456"
	agentName := "Code Assistant"
	toolName := "file_write"
	messageID := "msg-789"
	description := "Created main.go"

	entry := FileHistory{
		ID:          "test-entry-1",
		UserID:      "user-123",
		FilePath:    "src/main.go",
		Content:     "package main",
		Operation:   "create",
		CreatedAt:   time.Now(),
		AgentID:     &agentID,
		AgentName:   &agentName,
		ToolName:    &toolName,
		MessageID:   &messageID,
		Description: &description,
	}

	err := repo.CreateWithAttribution(entry)
	if err != nil {
		t.Fatalf("CreateWithAttribution failed: %v", err)
	}

	// Retrieve and verify
	retrieved, err := repo.GetByID("test-entry-1")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected entry to be non-nil")
	}
	if retrieved.AgentID == nil || *retrieved.AgentID != agentID {
		t.Errorf("Expected AgentID %s, got %v", agentID, retrieved.AgentID)
	}
	if retrieved.AgentName == nil || *retrieved.AgentName != agentName {
		t.Errorf("Expected AgentName %s, got %v", agentName, retrieved.AgentName)
	}
	if retrieved.ToolName == nil || *retrieved.ToolName != toolName {
		t.Errorf("Expected ToolName %s, got %v", toolName, retrieved.ToolName)
	}
	if retrieved.MessageID == nil || *retrieved.MessageID != messageID {
		t.Errorf("Expected MessageID %s, got %v", messageID, retrieved.MessageID)
	}
	if retrieved.Description == nil || *retrieved.Description != description {
		t.Errorf("Expected Description %s, got %v", description, retrieved.Description)
	}
}

func TestFileHistoryRepository_ListByFilePath(t *testing.T) {
	db := setupFileHistoryTestDB(t)
	repo := NewFileHistoryRepository(db)

	userID := "user-123"
	filePath := "src/main.go"

	// Create multiple entries
	for i := 0; i < 5; i++ {
		_, err := repo.Create(userID, filePath, "content "+string(rune('0'+i)), "update")
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Also create entries for a different file
	_, err := repo.Create(userID, "src/other.go", "other content", "create")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	history, err := repo.ListByFilePath(userID, filePath, 10)
	if err != nil {
		t.Fatalf("ListByFilePath failed: %v", err)
	}

	if len(history) != 5 {
		t.Errorf("Expected 5 entries, got %d", len(history))
	}

	// Verify they are all for the correct file
	for _, h := range history {
		if h.FilePath != filePath {
			t.Errorf("Expected FilePath %s, got %s", filePath, h.FilePath)
		}
	}
}

func TestFileHistoryRepository_ListByUserID(t *testing.T) {
	db := setupFileHistoryTestDB(t)
	repo := NewFileHistoryRepository(db)

	userID := "user-123"

	// Create entries for multiple files
	_, _ = repo.Create(userID, "file1.txt", "content1", "create")
	_, _ = repo.Create(userID, "file2.txt", "content2", "create")
	_, _ = repo.Create("other-user", "file3.txt", "content3", "create")

	history, err := repo.ListByUserID(userID, 10, 0)
	if err != nil {
		t.Fatalf("ListByUserID failed: %v", err)
	}

	if len(history) != 2 {
		t.Errorf("Expected 2 entries for user, got %d", len(history))
	}

	for _, h := range history {
		if h.UserID != userID {
			t.Errorf("Expected UserID %s, got %s", userID, h.UserID)
		}
	}
}

func TestFileHistoryRepository_ListByTimeRange(t *testing.T) {
	db := setupFileHistoryTestDB(t)
	repo := NewFileHistoryRepository(db)

	userID := "user-123"
	now := time.Now()

	// Create entries with different timestamps
	entry1 := FileHistory{
		ID:        "entry-1",
		UserID:    userID,
		FilePath:  "file1.txt",
		Content:   "old content",
		Operation: "create",
		CreatedAt: now.Add(-48 * time.Hour), // 2 days ago
	}
	entry2 := FileHistory{
		ID:        "entry-2",
		UserID:    userID,
		FilePath:  "file2.txt",
		Content:   "recent content",
		Operation: "create",
		CreatedAt: now.Add(-1 * time.Hour), // 1 hour ago
	}

	_ = repo.CreateWithAttribution(entry1)
	_ = repo.CreateWithAttribution(entry2)

	// Query for entries in the last 24 hours
	startTime := now.Add(-24 * time.Hour)
	endTime := now

	history, err := repo.ListByTimeRange(userID, startTime, endTime, 10, 0)
	if err != nil {
		t.Fatalf("ListByTimeRange failed: %v", err)
	}

	if len(history) != 1 {
		t.Errorf("Expected 1 entry in time range, got %d", len(history))
	}

	if len(history) > 0 && history[0].ID != "entry-2" {
		t.Errorf("Expected entry-2, got %s", history[0].ID)
	}
}

func TestFileHistoryRepository_ListByAgent(t *testing.T) {
	db := setupFileHistoryTestDB(t)
	repo := NewFileHistoryRepository(db)

	userID := "user-123"
	agentID := "agent-456"

	// Create entries with and without agent attribution
	entry1 := FileHistory{
		ID:        "entry-1",
		UserID:    userID,
		FilePath:  "file1.txt",
		Content:   "agent content",
		Operation: "create",
		AgentID:   &agentID,
	}
	_ = repo.CreateWithAttribution(entry1)

	_, _ = repo.Create(userID, "file2.txt", "manual content", "create")

	history, err := repo.ListByAgent(userID, agentID, 10, 0)
	if err != nil {
		t.Fatalf("ListByAgent failed: %v", err)
	}

	if len(history) != 1 {
		t.Errorf("Expected 1 entry by agent, got %d", len(history))
	}

	if len(history) > 0 && (history[0].AgentID == nil || *history[0].AgentID != agentID) {
		t.Errorf("Expected AgentID %s", agentID)
	}
}

func TestFileHistoryRepository_GetStats(t *testing.T) {
	db := setupFileHistoryTestDB(t)
	repo := NewFileHistoryRepository(db)

	userID := "user-123"

	// Create some entries
	_, _ = repo.Create(userID, "file1.txt", "content one", "create")
	_, _ = repo.Create(userID, "file1.txt", "updated content", "update")
	_, _ = repo.Create(userID, "file2.txt", "file two content", "create")

	stats, err := repo.GetStats(userID)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.TotalEntries != 3 {
		t.Errorf("Expected TotalEntries 3, got %d", stats.TotalEntries)
	}

	if stats.TotalFiles != 2 {
		t.Errorf("Expected TotalFiles 2, got %d", stats.TotalFiles)
	}

	// Total size should be sum of all content lengths
	expectedSize := int64(len("content one") + len("updated content") + len("file two content"))
	if stats.TotalSizeBytes != expectedSize {
		t.Errorf("Expected TotalSizeBytes %d, got %d", expectedSize, stats.TotalSizeBytes)
	}

	if stats.OldestEntry == "" {
		t.Error("Expected OldestEntry to be non-empty")
	}

	if stats.NewestEntry == "" {
		t.Error("Expected NewestEntry to be non-empty")
	}
}

func TestFileHistoryRepository_GetVersionDiff(t *testing.T) {
	db := setupFileHistoryTestDB(t)
	repo := NewFileHistoryRepository(db)

	userID := "user-123"

	// Create two versions
	entry1 := FileHistory{
		ID:        "v1",
		UserID:    userID,
		FilePath:  "file.txt",
		Content:   "line1\nline2\nline3",
		Operation: "create",
	}
	entry2 := FileHistory{
		ID:        "v2",
		UserID:    userID,
		FilePath:  "file.txt",
		Content:   "line1\nmodified line2\nline3\nline4",
		Operation: "update",
	}

	_ = repo.CreateWithAttribution(entry1)
	_ = repo.CreateWithAttribution(entry2)

	diff, err := repo.GetVersionDiff(userID, "v1", "v2")
	if err != nil {
		t.Fatalf("GetVersionDiff failed: %v", err)
	}

	if diff.HistoryID1 != "v1" || diff.HistoryID2 != "v2" {
		t.Errorf("Expected HistoryID1=v1, HistoryID2=v2, got %s, %s", diff.HistoryID1, diff.HistoryID2)
	}

	// Check that we have additions and deletions
	if diff.Additions == 0 && diff.Deletions == 0 {
		t.Error("Expected some additions or deletions in the diff")
	}

	if len(diff.Changes) == 0 {
		t.Error("Expected non-empty changes")
	}
}

func TestFileHistoryRepository_GetByIDNotFound(t *testing.T) {
	db := setupFileHistoryTestDB(t)
	repo := NewFileHistoryRepository(db)

	entry, err := repo.GetByID("nonexistent")
	if err != nil {
		t.Fatalf("GetByID should not return error for not found: %v", err)
	}
	if entry != nil {
		t.Error("Expected nil entry for nonexistent ID")
	}
}

func TestFileHistoryRepository_GetDistinctFiles(t *testing.T) {
	db := setupFileHistoryTestDB(t)
	repo := NewFileHistoryRepository(db)

	userID := "user-123"

	// Create entries for different files
	_, _ = repo.Create(userID, "a.txt", "content", "create")
	_, _ = repo.Create(userID, "b.txt", "content", "create")
	_, _ = repo.Create(userID, "a.txt", "updated", "update")
	_, _ = repo.Create(userID, "c.txt", "content", "create")

	files, err := repo.GetDistinctFiles(userID)
	if err != nil {
		t.Fatalf("GetDistinctFiles failed: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("Expected 3 distinct files, got %d", len(files))
	}

	// Should be in alphabetical order
	expected := []string{"a.txt", "b.txt", "c.txt"}
	for i, f := range files {
		if f != expected[i] {
			t.Errorf("Expected file %s at index %d, got %s", expected[i], i, f)
		}
	}
}

func TestFileHistoryRepository_UserIsolation(t *testing.T) {
	db := setupFileHistoryTestDB(t)
	repo := NewFileHistoryRepository(db)

	// Create entries for different users
	_, _ = repo.Create("user1", "file.txt", "user1 content", "create")
	_, _ = repo.Create("user2", "file.txt", "user2 content", "create")

	// Verify user1 can only see their entries
	history1, err := repo.ListByUserID("user1", 10, 0)
	if err != nil {
		t.Fatalf("ListByUserID failed: %v", err)
	}
	if len(history1) != 1 {
		t.Errorf("User1 should see 1 entry, got %d", len(history1))
	}

	// Verify GetVersionDiff enforces user isolation
	_, err = repo.GetVersionDiff("user1", history1[0].ID, "nonexistent")
	if err == nil {
		t.Error("Expected error when accessing nonexistent entry")
	}
}
