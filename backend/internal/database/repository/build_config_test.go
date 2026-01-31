package repository

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/jacklau/prism/internal/security"
)

func setupBuildConfigTestDB(t *testing.T) (*sql.DB, *security.EncryptionService) {
	t.Helper()

	// Create a temporary database
	tmpFile, err := os.CreateTemp("", "test_build_config_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	db, err := sql.Open("sqlite3", tmpFile.Name()+"?_foreign_keys=on")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create the tables
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS build_configs (
			id TEXT PRIMARY KEY,
			workspace_id TEXT,
			org_workspace_id TEXT,
			user_id TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			is_default INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create build_configs table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS build_commands (
			id TEXT PRIMARY KEY,
			config_id TEXT NOT NULL,
			name TEXT NOT NULL,
			command TEXT NOT NULL,
			working_directory TEXT,
			run_order INTEGER DEFAULT 0,
			is_enabled INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (config_id) REFERENCES build_configs(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create build_commands table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS build_env_vars (
			id TEXT PRIMARY KEY,
			config_id TEXT NOT NULL,
			key TEXT NOT NULL,
			value_encrypted BLOB NOT NULL,
			value_nonce BLOB NOT NULL,
			is_secret INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (config_id) REFERENCES build_configs(id) ON DELETE CASCADE,
			UNIQUE(config_id, key)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create build_env_vars table: %v", err)
	}

	// Create encryption service with a test key
	crypto, err := security.NewEncryptionService("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("Failed to create encryption service: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
		os.Remove(tmpFile.Name())
	})

	return db, crypto
}

func TestBuildConfigRepository_Create(t *testing.T) {
	db, crypto := setupBuildConfigTestDB(t)
	repo := NewBuildConfigRepository(db, crypto)

	workspaceID := "ws-123"
	config := &BuildConfig{
		WorkspaceID: &workspaceID,
		UserID:      "user-123",
		Name:        "Test Config",
	}

	err := repo.Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if config.ID == "" {
		t.Error("Expected ID to be set")
	}
}

func TestBuildConfigRepository_GetByID(t *testing.T) {
	db, crypto := setupBuildConfigTestDB(t)
	repo := NewBuildConfigRepository(db, crypto)

	// Create a config
	workspaceID := "ws-123"
	description := "Test description"
	config := &BuildConfig{
		WorkspaceID: &workspaceID,
		UserID:      "user-123",
		Name:        "Test Config",
		Description: &description,
		IsDefault:   true,
	}
	err := repo.Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Retrieve it
	retrieved, err := repo.GetByID(config.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected config to be non-nil")
	}
	if retrieved.Name != "Test Config" {
		t.Errorf("Expected name 'Test Config', got %s", retrieved.Name)
	}
	if *retrieved.Description != "Test description" {
		t.Errorf("Expected description 'Test description', got %s", *retrieved.Description)
	}
	if !retrieved.IsDefault {
		t.Error("Expected IsDefault to be true")
	}
}

func TestBuildConfigRepository_GetByIDNotFound(t *testing.T) {
	db, crypto := setupBuildConfigTestDB(t)
	repo := NewBuildConfigRepository(db, crypto)

	config, err := repo.GetByID("nonexistent")
	if err != nil {
		t.Fatalf("GetByID should not return error for not found: %v", err)
	}
	if config != nil {
		t.Error("Expected nil config for nonexistent entry")
	}
}

func TestBuildConfigRepository_Update(t *testing.T) {
	db, crypto := setupBuildConfigTestDB(t)
	repo := NewBuildConfigRepository(db, crypto)

	// Create a config
	config := &BuildConfig{
		UserID: "user-123",
		Name:   "Original Name",
	}
	err := repo.Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update it
	config.Name = "Updated Name"
	description := "New description"
	config.Description = &description
	err = repo.Update(config)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Retrieve and verify
	retrieved, err := repo.GetByID(config.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got %s", retrieved.Name)
	}
	if *retrieved.Description != "New description" {
		t.Errorf("Expected description 'New description', got %s", *retrieved.Description)
	}
}

func TestBuildConfigRepository_Delete(t *testing.T) {
	db, crypto := setupBuildConfigTestDB(t)
	repo := NewBuildConfigRepository(db, crypto)

	// Create a config
	config := &BuildConfig{
		UserID: "user-123",
		Name:   "To Delete",
	}
	err := repo.Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Delete it
	err = repo.Delete(config.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	retrieved, err := repo.GetByID(config.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if retrieved != nil {
		t.Error("Expected nil config after deletion")
	}
}

func TestBuildConfigRepository_ListByWorkspaceID(t *testing.T) {
	db, crypto := setupBuildConfigTestDB(t)
	repo := NewBuildConfigRepository(db, crypto)

	workspaceID := "ws-123"

	// Create multiple configs
	for i, name := range []string{"Config A", "Config B", "Config C"} {
		config := &BuildConfig{
			WorkspaceID: &workspaceID,
			UserID:      "user-123",
			Name:        name,
			IsDefault:   i == 0, // First one is default
		}
		err := repo.Create(config)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// Create one for different workspace
	otherWorkspaceID := "ws-other"
	config := &BuildConfig{
		WorkspaceID: &otherWorkspaceID,
		UserID:      "user-123",
		Name:        "Other Config",
	}
	err := repo.Create(config)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// List configs for workspace
	configs, err := repo.ListByWorkspaceID(workspaceID)
	if err != nil {
		t.Fatalf("ListByWorkspaceID failed: %v", err)
	}

	if len(configs) != 3 {
		t.Errorf("Expected 3 configs, got %d", len(configs))
	}

	// Default should be first
	if !configs[0].IsDefault {
		t.Error("Expected first config to be default")
	}
}

func TestBuildConfigRepository_SetDefault(t *testing.T) {
	db, crypto := setupBuildConfigTestDB(t)
	repo := NewBuildConfigRepository(db, crypto)

	workspaceID := "ws-123"

	// Create two configs
	config1 := &BuildConfig{
		WorkspaceID: &workspaceID,
		UserID:      "user-123",
		Name:        "Config 1",
		IsDefault:   true,
	}
	err := repo.Create(config1)
	if err != nil {
		t.Fatalf("Create config1 failed: %v", err)
	}

	config2 := &BuildConfig{
		WorkspaceID: &workspaceID,
		UserID:      "user-123",
		Name:        "Config 2",
		IsDefault:   false,
	}
	err = repo.Create(config2)
	if err != nil {
		t.Fatalf("Create config2 failed: %v", err)
	}

	// Set config2 as default
	err = repo.SetDefault(config2.ID, workspaceID)
	if err != nil {
		t.Fatalf("SetDefault failed: %v", err)
	}

	// Verify config1 is no longer default
	retrieved1, _ := repo.GetByID(config1.ID)
	if retrieved1.IsDefault {
		t.Error("Expected config1 to not be default")
	}

	// Verify config2 is default
	retrieved2, _ := repo.GetByID(config2.ID)
	if !retrieved2.IsDefault {
		t.Error("Expected config2 to be default")
	}
}

// ==================== Command Tests ====================

func TestBuildConfigRepository_AddCommand(t *testing.T) {
	db, crypto := setupBuildConfigTestDB(t)
	repo := NewBuildConfigRepository(db, crypto)

	// Create a config
	config := &BuildConfig{
		UserID: "user-123",
		Name:   "Test Config",
	}
	err := repo.Create(config)
	if err != nil {
		t.Fatalf("Create config failed: %v", err)
	}

	// Add a command
	cmd := &BuildCommand{
		ConfigID:  config.ID,
		Name:      "Build",
		Command:   "npm run build",
		RunOrder:  0,
		IsEnabled: true,
	}
	err = repo.AddCommand(cmd)
	if err != nil {
		t.Fatalf("AddCommand failed: %v", err)
	}

	if cmd.ID == "" {
		t.Error("Expected ID to be set")
	}
}

func TestBuildConfigRepository_CommandValidation(t *testing.T) {
	db, crypto := setupBuildConfigTestDB(t)
	repo := NewBuildConfigRepository(db, crypto)

	// Create a config
	config := &BuildConfig{
		UserID: "user-123",
		Name:   "Test Config",
	}
	err := repo.Create(config)
	if err != nil {
		t.Fatalf("Create config failed: %v", err)
	}

	// Try to add command with empty command string
	cmd := &BuildCommand{
		ConfigID:  config.ID,
		Name:      "Build",
		Command:   "",
		IsEnabled: true,
	}
	err = repo.AddCommand(cmd)
	if err == nil {
		t.Error("Expected error for empty command")
	}
}

func TestBuildConfigRepository_GetByIDWithDetails(t *testing.T) {
	db, crypto := setupBuildConfigTestDB(t)
	repo := NewBuildConfigRepository(db, crypto)

	// Create a config
	config := &BuildConfig{
		UserID: "user-123",
		Name:   "Test Config",
	}
	err := repo.Create(config)
	if err != nil {
		t.Fatalf("Create config failed: %v", err)
	}

	// Add commands
	cmd1 := &BuildCommand{
		ConfigID:  config.ID,
		Name:      "Install",
		Command:   "npm install",
		RunOrder:  0,
		IsEnabled: true,
	}
	repo.AddCommand(cmd1)

	cmd2 := &BuildCommand{
		ConfigID:  config.ID,
		Name:      "Build",
		Command:   "npm run build",
		RunOrder:  1,
		IsEnabled: true,
	}
	repo.AddCommand(cmd2)

	// Add env vars
	envVar := &BuildEnvVar{
		ConfigID: config.ID,
		Key:      "NODE_ENV",
		Value:    "production",
		IsSecret: false,
	}
	repo.SetEnvVar(envVar)

	// Get with details
	retrieved, err := repo.GetByIDWithDetails(config.ID)
	if err != nil {
		t.Fatalf("GetByIDWithDetails failed: %v", err)
	}

	if len(retrieved.Commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(retrieved.Commands))
	}
	if len(retrieved.EnvVars) != 1 {
		t.Errorf("Expected 1 env var, got %d", len(retrieved.EnvVars))
	}

	// Verify command order
	if retrieved.Commands[0].Name != "Install" {
		t.Errorf("Expected first command to be 'Install', got %s", retrieved.Commands[0].Name)
	}
}

func TestBuildConfigRepository_ReorderCommands(t *testing.T) {
	db, crypto := setupBuildConfigTestDB(t)
	repo := NewBuildConfigRepository(db, crypto)

	// Create a config
	config := &BuildConfig{
		UserID: "user-123",
		Name:   "Test Config",
	}
	err := repo.Create(config)
	if err != nil {
		t.Fatalf("Create config failed: %v", err)
	}

	// Add commands
	cmd1 := &BuildCommand{ConfigID: config.ID, Name: "First", Command: "echo 1", RunOrder: 0, IsEnabled: true}
	cmd2 := &BuildCommand{ConfigID: config.ID, Name: "Second", Command: "echo 2", RunOrder: 1, IsEnabled: true}
	cmd3 := &BuildCommand{ConfigID: config.ID, Name: "Third", Command: "echo 3", RunOrder: 2, IsEnabled: true}
	repo.AddCommand(cmd1)
	repo.AddCommand(cmd2)
	repo.AddCommand(cmd3)

	// Reorder: Third, First, Second
	newOrder := []string{cmd3.ID, cmd1.ID, cmd2.ID}
	err = repo.ReorderCommands(config.ID, newOrder)
	if err != nil {
		t.Fatalf("ReorderCommands failed: %v", err)
	}

	// Verify order
	retrieved, _ := repo.GetByIDWithDetails(config.ID)
	if retrieved.Commands[0].Name != "Third" {
		t.Errorf("Expected first command to be 'Third', got %s", retrieved.Commands[0].Name)
	}
	if retrieved.Commands[1].Name != "First" {
		t.Errorf("Expected second command to be 'First', got %s", retrieved.Commands[1].Name)
	}
}

// ==================== Environment Variable Tests ====================

func TestBuildConfigRepository_SetEnvVar(t *testing.T) {
	db, crypto := setupBuildConfigTestDB(t)
	repo := NewBuildConfigRepository(db, crypto)

	// Create a config
	config := &BuildConfig{
		UserID: "user-123",
		Name:   "Test Config",
	}
	err := repo.Create(config)
	if err != nil {
		t.Fatalf("Create config failed: %v", err)
	}

	// Set env var
	envVar := &BuildEnvVar{
		ConfigID: config.ID,
		Key:      "API_KEY",
		Value:    "secret123",
		IsSecret: true,
	}
	err = repo.SetEnvVar(envVar)
	if err != nil {
		t.Fatalf("SetEnvVar failed: %v", err)
	}

	// Get env vars
	envVars, err := repo.GetEnvVars(config.ID)
	if err != nil {
		t.Fatalf("GetEnvVars failed: %v", err)
	}

	if len(envVars) != 1 {
		t.Errorf("Expected 1 env var, got %d", len(envVars))
	}
	if envVars[0].Key != "API_KEY" {
		t.Errorf("Expected key 'API_KEY', got %s", envVars[0].Key)
	}
	if envVars[0].Value != "secret123" {
		t.Errorf("Expected value 'secret123', got %s", envVars[0].Value)
	}
	if !envVars[0].IsSecret {
		t.Error("Expected IsSecret to be true")
	}
}

func TestBuildConfigRepository_EnvVarKeyValidation(t *testing.T) {
	db, crypto := setupBuildConfigTestDB(t)
	repo := NewBuildConfigRepository(db, crypto)

	// Create a config
	config := &BuildConfig{
		UserID: "user-123",
		Name:   "Test Config",
	}
	repo.Create(config)

	// Invalid keys
	invalidKeys := []string{
		"",           // empty
		"123ABC",     // starts with number
		"has-dash",   // contains dash
		"has space",  // contains space
		"has.dot",    // contains dot
	}

	for _, key := range invalidKeys {
		envVar := &BuildEnvVar{
			ConfigID: config.ID,
			Key:      key,
			Value:    "test",
		}
		err := repo.SetEnvVar(envVar)
		if err == nil {
			t.Errorf("Expected error for invalid key '%s'", key)
		}
	}

	// Valid keys
	validKeys := []string{
		"API_KEY",
		"_PRIVATE",
		"myVar123",
		"NODE_ENV",
	}

	for _, key := range validKeys {
		envVar := &BuildEnvVar{
			ConfigID: config.ID,
			Key:      key,
			Value:    "test",
		}
		err := repo.SetEnvVar(envVar)
		if err != nil {
			t.Errorf("Expected no error for valid key '%s', got %v", key, err)
		}
	}
}

func TestBuildConfigRepository_EnvVarUpsert(t *testing.T) {
	db, crypto := setupBuildConfigTestDB(t)
	repo := NewBuildConfigRepository(db, crypto)

	// Create a config
	config := &BuildConfig{
		UserID: "user-123",
		Name:   "Test Config",
	}
	repo.Create(config)

	// Set env var
	envVar := &BuildEnvVar{
		ConfigID: config.ID,
		Key:      "API_KEY",
		Value:    "old_value",
		IsSecret: false,
	}
	err := repo.SetEnvVar(envVar)
	if err != nil {
		t.Fatalf("First SetEnvVar failed: %v", err)
	}

	// Update with new value
	envVar.Value = "new_value"
	envVar.IsSecret = true
	err = repo.SetEnvVar(envVar)
	if err != nil {
		t.Fatalf("Second SetEnvVar failed: %v", err)
	}

	// Get and verify
	envVars, _ := repo.GetEnvVars(config.ID)
	if len(envVars) != 1 {
		t.Errorf("Expected 1 env var after upsert, got %d", len(envVars))
	}
	if envVars[0].Value != "new_value" {
		t.Errorf("Expected value 'new_value', got %s", envVars[0].Value)
	}
	if !envVars[0].IsSecret {
		t.Error("Expected IsSecret to be true after update")
	}
}

func TestBuildConfigRepository_EnvVarMasked(t *testing.T) {
	db, crypto := setupBuildConfigTestDB(t)
	repo := NewBuildConfigRepository(db, crypto)

	// Create a config
	config := &BuildConfig{
		UserID: "user-123",
		Name:   "Test Config",
	}
	repo.Create(config)

	// Set secret env var
	secretEnv := &BuildEnvVar{
		ConfigID: config.ID,
		Key:      "SECRET_KEY",
		Value:    "super_secret",
		IsSecret: true,
	}
	repo.SetEnvVar(secretEnv)

	// Set non-secret env var
	normalEnv := &BuildEnvVar{
		ConfigID: config.ID,
		Key:      "NODE_ENV",
		Value:    "production",
		IsSecret: false,
	}
	repo.SetEnvVar(normalEnv)

	// Get masked env vars
	envVars, err := repo.GetEnvVarsMasked(config.ID)
	if err != nil {
		t.Fatalf("GetEnvVarsMasked failed: %v", err)
	}

	for _, ev := range envVars {
		if ev.Key == "SECRET_KEY" && ev.Value != "********" {
			t.Errorf("Expected masked value '********' for secret, got %s", ev.Value)
		}
		if ev.Key == "NODE_ENV" && ev.Value != "production" {
			t.Errorf("Expected non-secret value 'production', got %s", ev.Value)
		}
	}
}

func TestBuildConfigRepository_DeleteEnvVarByKey(t *testing.T) {
	db, crypto := setupBuildConfigTestDB(t)
	repo := NewBuildConfigRepository(db, crypto)

	// Create a config
	config := &BuildConfig{
		UserID: "user-123",
		Name:   "Test Config",
	}
	repo.Create(config)

	// Set env var
	envVar := &BuildEnvVar{
		ConfigID: config.ID,
		Key:      "TO_DELETE",
		Value:    "value",
	}
	repo.SetEnvVar(envVar)

	// Delete by key
	err := repo.DeleteEnvVarByKey(config.ID, "TO_DELETE")
	if err != nil {
		t.Fatalf("DeleteEnvVarByKey failed: %v", err)
	}

	// Verify it's gone
	envVars, _ := repo.GetEnvVars(config.ID)
	if len(envVars) != 0 {
		t.Error("Expected 0 env vars after deletion")
	}
}

func TestBuildConfigRepository_CascadeDelete(t *testing.T) {
	db, crypto := setupBuildConfigTestDB(t)
	repo := NewBuildConfigRepository(db, crypto)

	// Create a config
	config := &BuildConfig{
		UserID: "user-123",
		Name:   "Test Config",
	}
	repo.Create(config)

	// Add commands
	cmd := &BuildCommand{
		ConfigID:  config.ID,
		Name:      "Build",
		Command:   "npm build",
		IsEnabled: true,
	}
	repo.AddCommand(cmd)

	// Add env var
	envVar := &BuildEnvVar{
		ConfigID: config.ID,
		Key:      "ENV_VAR",
		Value:    "value",
	}
	repo.SetEnvVar(envVar)

	// Delete config
	err := repo.Delete(config.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify commands are gone (cascade)
	var cmdCount int
	db.QueryRow("SELECT COUNT(*) FROM build_commands WHERE config_id = ?", config.ID).Scan(&cmdCount)
	if cmdCount != 0 {
		t.Errorf("Expected 0 commands after cascade delete, got %d", cmdCount)
	}

	// Verify env vars are gone (cascade)
	var envCount int
	db.QueryRow("SELECT COUNT(*) FROM build_env_vars WHERE config_id = ?", config.ID).Scan(&envCount)
	if envCount != 0 {
		t.Errorf("Expected 0 env vars after cascade delete, got %d", envCount)
	}
}

func TestBuildConfigRepository_EncryptionDecryption(t *testing.T) {
	db, crypto := setupBuildConfigTestDB(t)
	repo := NewBuildConfigRepository(db, crypto)

	// Create a config
	config := &BuildConfig{
		UserID: "user-123",
		Name:   "Test Config",
	}
	repo.Create(config)

	// Set env var with sensitive value
	sensitiveValue := "sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	envVar := &BuildEnvVar{
		ConfigID: config.ID,
		Key:      "API_KEY",
		Value:    sensitiveValue,
		IsSecret: true,
	}
	err := repo.SetEnvVar(envVar)
	if err != nil {
		t.Fatalf("SetEnvVar failed: %v", err)
	}

	// Verify the value is encrypted in the database
	var encryptedValue []byte
	err = db.QueryRow("SELECT value_encrypted FROM build_env_vars WHERE config_id = ? AND key = ?", config.ID, "API_KEY").Scan(&encryptedValue)
	if err != nil {
		t.Fatalf("Failed to query encrypted value: %v", err)
	}

	// The encrypted value should NOT be equal to the plaintext
	if string(encryptedValue) == sensitiveValue {
		t.Error("Value appears to be stored in plaintext!")
	}

	// Retrieve and verify decryption works
	envVars, err := repo.GetEnvVars(config.ID)
	if err != nil {
		t.Fatalf("GetEnvVars failed: %v", err)
	}

	if envVars[0].Value != sensitiveValue {
		t.Errorf("Expected decrypted value '%s', got '%s'", sensitiveValue, envVars[0].Value)
	}
}
