package repository

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/jacklau/prism/internal/security"
)

func setupTestDB(t *testing.T) (*sql.DB, *security.EncryptionService) {
	t.Helper()

	// Create a temporary database
	tmpFile, err := os.CreateTemp("", "test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	db, err := sql.Open("sqlite3", tmpFile.Name()+"?_foreign_keys=on")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create the data_configs table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS data_configs (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			config_type TEXT NOT NULL,
			config_key TEXT NOT NULL,
			encrypted_data BLOB NOT NULL,
			data_nonce BLOB NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, config_type, config_key)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
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

func TestDataConfigRepository_SetAndGet(t *testing.T) {
	db, crypto := setupTestDB(t)
	repo := NewDataConfigRepository(db, crypto)

	userID := "user-123"
	configType := "stripe"
	configKey := "api_credentials"
	data := map[string]interface{}{
		"api_key":     "sk_test_xxx",
		"webhook_key": "whsec_xxx",
		"enabled":     true,
	}

	// Test Set
	err := repo.SetDataConfig(userID, configType, configKey, data)
	if err != nil {
		t.Fatalf("SetDataConfig failed: %v", err)
	}

	// Test Get
	config, err := repo.GetDataConfig(userID, configType, configKey)
	if err != nil {
		t.Fatalf("GetDataConfig failed: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to be non-nil")
	}

	if config.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, config.UserID)
	}
	if config.ConfigType != configType {
		t.Errorf("Expected ConfigType %s, got %s", configType, config.ConfigType)
	}
	if config.ConfigKey != configKey {
		t.Errorf("Expected ConfigKey %s, got %s", configKey, config.ConfigKey)
	}

	// Check the decrypted data
	if config.Data["api_key"] != "sk_test_xxx" {
		t.Errorf("Expected api_key 'sk_test_xxx', got %v", config.Data["api_key"])
	}
	if config.Data["webhook_key"] != "whsec_xxx" {
		t.Errorf("Expected webhook_key 'whsec_xxx', got %v", config.Data["webhook_key"])
	}
	if config.Data["enabled"] != true {
		t.Errorf("Expected enabled true, got %v", config.Data["enabled"])
	}
}

func TestDataConfigRepository_Upsert(t *testing.T) {
	db, crypto := setupTestDB(t)
	repo := NewDataConfigRepository(db, crypto)

	userID := "user-123"
	configType := "stripe"
	configKey := "api_credentials"

	// Initial set
	data1 := map[string]interface{}{"api_key": "old_key"}
	err := repo.SetDataConfig(userID, configType, configKey, data1)
	if err != nil {
		t.Fatalf("First SetDataConfig failed: %v", err)
	}

	// Update with new data
	data2 := map[string]interface{}{"api_key": "new_key", "extra": "value"}
	err = repo.SetDataConfig(userID, configType, configKey, data2)
	if err != nil {
		t.Fatalf("Second SetDataConfig failed: %v", err)
	}

	// Get and verify updated data
	config, err := repo.GetDataConfig(userID, configType, configKey)
	if err != nil {
		t.Fatalf("GetDataConfig failed: %v", err)
	}

	if config.Data["api_key"] != "new_key" {
		t.Errorf("Expected updated api_key 'new_key', got %v", config.Data["api_key"])
	}
	if config.Data["extra"] != "value" {
		t.Errorf("Expected extra 'value', got %v", config.Data["extra"])
	}
}

func TestDataConfigRepository_GetNotFound(t *testing.T) {
	db, crypto := setupTestDB(t)
	repo := NewDataConfigRepository(db, crypto)

	config, err := repo.GetDataConfig("nonexistent", "type", "key")
	if err != nil {
		t.Fatalf("GetDataConfig should not return error for not found: %v", err)
	}
	if config != nil {
		t.Error("Expected nil config for nonexistent entry")
	}
}

func TestDataConfigRepository_Delete(t *testing.T) {
	db, crypto := setupTestDB(t)
	repo := NewDataConfigRepository(db, crypto)

	userID := "user-123"
	configType := "stripe"
	configKey := "api_credentials"

	// Create a config
	data := map[string]interface{}{"api_key": "test"}
	err := repo.SetDataConfig(userID, configType, configKey, data)
	if err != nil {
		t.Fatalf("SetDataConfig failed: %v", err)
	}

	// Delete it
	err = repo.DeleteDataConfig(userID, configType, configKey)
	if err != nil {
		t.Fatalf("DeleteDataConfig failed: %v", err)
	}

	// Verify it's gone
	config, err := repo.GetDataConfig(userID, configType, configKey)
	if err != nil {
		t.Fatalf("GetDataConfig failed: %v", err)
	}
	if config != nil {
		t.Error("Expected nil config after deletion")
	}
}

func TestDataConfigRepository_DeleteNotFound(t *testing.T) {
	db, crypto := setupTestDB(t)
	repo := NewDataConfigRepository(db, crypto)

	err := repo.DeleteDataConfig("nonexistent", "type", "key")
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows for nonexistent delete, got %v", err)
	}
}

func TestDataConfigRepository_List(t *testing.T) {
	db, crypto := setupTestDB(t)
	repo := NewDataConfigRepository(db, crypto)

	userID := "user-123"
	configType := "stripe"

	// Create multiple configs
	for _, key := range []string{"credentials", "settings", "webhooks"} {
		data := map[string]interface{}{"key": key}
		err := repo.SetDataConfig(userID, configType, key, data)
		if err != nil {
			t.Fatalf("SetDataConfig failed for %s: %v", key, err)
		}
	}

	// List all configs for this type
	configs, err := repo.ListDataConfigs(userID, configType)
	if err != nil {
		t.Fatalf("ListDataConfigs failed: %v", err)
	}

	if len(configs) != 3 {
		t.Errorf("Expected 3 configs, got %d", len(configs))
	}

	// Verify ordering (should be alphabetical by key)
	expectedKeys := []string{"credentials", "settings", "webhooks"}
	for i, cfg := range configs {
		if cfg.ConfigKey != expectedKeys[i] {
			t.Errorf("Expected key %s at index %d, got %s", expectedKeys[i], i, cfg.ConfigKey)
		}
		// Data should be nil (not decrypted in list)
		if cfg.Data != nil {
			t.Errorf("Expected nil Data in list, got %v", cfg.Data)
		}
	}
}

func TestDataConfigRepository_ListEmpty(t *testing.T) {
	db, crypto := setupTestDB(t)
	repo := NewDataConfigRepository(db, crypto)

	configs, err := repo.ListDataConfigs("user-123", "nonexistent")
	if err != nil {
		t.Fatalf("ListDataConfigs failed: %v", err)
	}
	if configs != nil && len(configs) != 0 {
		t.Errorf("Expected empty slice, got %v", configs)
	}
}

func TestDataConfigRepository_Has(t *testing.T) {
	db, crypto := setupTestDB(t)
	repo := NewDataConfigRepository(db, crypto)

	userID := "user-123"
	configType := "stripe"
	configKey := "api_credentials"

	// Check before creating
	exists, err := repo.HasDataConfig(userID, configType, configKey)
	if err != nil {
		t.Fatalf("HasDataConfig failed: %v", err)
	}
	if exists {
		t.Error("Expected exists=false before creating")
	}

	// Create a config
	data := map[string]interface{}{"api_key": "test"}
	err = repo.SetDataConfig(userID, configType, configKey, data)
	if err != nil {
		t.Fatalf("SetDataConfig failed: %v", err)
	}

	// Check after creating
	exists, err = repo.HasDataConfig(userID, configType, configKey)
	if err != nil {
		t.Fatalf("HasDataConfig failed: %v", err)
	}
	if !exists {
		t.Error("Expected exists=true after creating")
	}
}

func TestDataConfigRepository_UserIsolation(t *testing.T) {
	db, crypto := setupTestDB(t)
	repo := NewDataConfigRepository(db, crypto)

	configType := "stripe"
	configKey := "api_credentials"

	// Create config for user1
	data1 := map[string]interface{}{"api_key": "user1_key"}
	err := repo.SetDataConfig("user1", configType, configKey, data1)
	if err != nil {
		t.Fatalf("SetDataConfig for user1 failed: %v", err)
	}

	// Create config for user2
	data2 := map[string]interface{}{"api_key": "user2_key"}
	err = repo.SetDataConfig("user2", configType, configKey, data2)
	if err != nil {
		t.Fatalf("SetDataConfig for user2 failed: %v", err)
	}

	// Verify user1 can only see their own data
	config1, err := repo.GetDataConfig("user1", configType, configKey)
	if err != nil {
		t.Fatalf("GetDataConfig for user1 failed: %v", err)
	}
	if config1.Data["api_key"] != "user1_key" {
		t.Errorf("User1 should see their own key, got %v", config1.Data["api_key"])
	}

	// Verify user2 can only see their own data
	config2, err := repo.GetDataConfig("user2", configType, configKey)
	if err != nil {
		t.Fatalf("GetDataConfig for user2 failed: %v", err)
	}
	if config2.Data["api_key"] != "user2_key" {
		t.Errorf("User2 should see their own key, got %v", config2.Data["api_key"])
	}

	// Verify user1 can't access user2's configs in list
	configs, err := repo.ListDataConfigs("user1", configType)
	if err != nil {
		t.Fatalf("ListDataConfigs for user1 failed: %v", err)
	}
	if len(configs) != 1 {
		t.Errorf("User1 should only see 1 config, got %d", len(configs))
	}
}

func TestDataConfigRepository_ListConfigTypes(t *testing.T) {
	db, crypto := setupTestDB(t)
	repo := NewDataConfigRepository(db, crypto)

	userID := "user-123"

	// Create configs with different types
	types := []string{"stripe", "github", "aws"}
	for _, configType := range types {
		data := map[string]interface{}{"type": configType}
		err := repo.SetDataConfig(userID, configType, "default", data)
		if err != nil {
			t.Fatalf("SetDataConfig failed for type %s: %v", configType, err)
		}
	}

	// Create multiple keys under one type to verify distinctness
	data := map[string]interface{}{"key": "extra"}
	err := repo.SetDataConfig(userID, "stripe", "extra_key", data)
	if err != nil {
		t.Fatalf("SetDataConfig for extra key failed: %v", err)
	}

	// List all types
	resultTypes, err := repo.ListConfigTypes(userID)
	if err != nil {
		t.Fatalf("ListConfigTypes failed: %v", err)
	}

	if len(resultTypes) != 3 {
		t.Errorf("Expected 3 distinct types, got %d", len(resultTypes))
	}

	// Verify ordering (should be alphabetical)
	expectedTypes := []string{"aws", "github", "stripe"}
	for i, typ := range resultTypes {
		if typ != expectedTypes[i] {
			t.Errorf("Expected type %s at index %d, got %s", expectedTypes[i], i, typ)
		}
	}
}

func TestDataConfigRepository_ListConfigTypesEmpty(t *testing.T) {
	db, crypto := setupTestDB(t)
	repo := NewDataConfigRepository(db, crypto)

	types, err := repo.ListConfigTypes("nonexistent-user")
	if err != nil {
		t.Fatalf("ListConfigTypes failed: %v", err)
	}
	if types != nil && len(types) != 0 {
		t.Errorf("Expected empty slice, got %v", types)
	}
}

func TestDataConfigRepository_ListConfigTypesUserIsolation(t *testing.T) {
	db, crypto := setupTestDB(t)
	repo := NewDataConfigRepository(db, crypto)

	// Create configs for user1
	data := map[string]interface{}{"key": "value"}
	err := repo.SetDataConfig("user1", "typeA", "key1", data)
	if err != nil {
		t.Fatalf("SetDataConfig failed: %v", err)
	}
	err = repo.SetDataConfig("user1", "typeB", "key1", data)
	if err != nil {
		t.Fatalf("SetDataConfig failed: %v", err)
	}

	// Create config for user2
	err = repo.SetDataConfig("user2", "typeC", "key1", data)
	if err != nil {
		t.Fatalf("SetDataConfig failed: %v", err)
	}

	// Verify user1 only sees their types
	types1, err := repo.ListConfigTypes("user1")
	if err != nil {
		t.Fatalf("ListConfigTypes for user1 failed: %v", err)
	}
	if len(types1) != 2 {
		t.Errorf("User1 should see 2 types, got %d", len(types1))
	}

	// Verify user2 only sees their type
	types2, err := repo.ListConfigTypes("user2")
	if err != nil {
		t.Fatalf("ListConfigTypes for user2 failed: %v", err)
	}
	if len(types2) != 1 {
		t.Errorf("User2 should see 1 type, got %d", len(types2))
	}
	if types2[0] != "typeC" {
		t.Errorf("User2's type should be 'typeC', got %s", types2[0])
	}
}
