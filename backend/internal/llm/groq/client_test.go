package groq

import (
	"testing"

	"github.com/jacklau/prism/internal/llm"
)

// Compile-time check that Client implements Provider interface
var _ llm.Provider = (*Client)(nil)

func TestClient_Name(t *testing.T) {
	client := NewClient("")
	if client.Name() != "groq" {
		t.Errorf("expected name 'groq', got '%s'", client.Name())
	}
}

func TestClient_Models(t *testing.T) {
	client := NewClient("")
	models := client.Models()
	if len(models) == 0 {
		t.Error("expected at least one model")
	}

	// Check that we have some expected models
	modelIDs := make(map[string]bool)
	for _, m := range models {
		modelIDs[m.ID] = true
	}

	expectedModels := []string{
		"llama-3.3-70b-versatile",
		"llama-3.2-90b-vision-preview",
		"mixtral-8x7b-32768",
	}

	for _, expected := range expectedModels {
		if !modelIDs[expected] {
			t.Errorf("expected model '%s' not found", expected)
		}
	}
}

func TestClient_SupportsTools(t *testing.T) {
	client := NewClient("")
	if !client.SupportsTools() {
		t.Error("expected SupportsTools to return true")
	}
}

func TestClient_SupportsVision(t *testing.T) {
	client := NewClient("")
	if !client.SupportsVision() {
		t.Error("expected SupportsVision to return true")
	}
}

func TestClient_HasConfiguredKey(t *testing.T) {
	client := NewClient("")
	if client.HasConfiguredKey() {
		t.Error("expected HasConfiguredKey to return false for empty key")
	}

	client = NewClient("test-key")
	if !client.HasConfiguredKey() {
		t.Error("expected HasConfiguredKey to return true for non-empty key")
	}
}

func TestClient_SetAPIKey(t *testing.T) {
	client := NewClient("")
	if client.HasConfiguredKey() {
		t.Error("expected HasConfiguredKey to return false initially")
	}

	client.SetAPIKey("new-key")
	if !client.HasConfiguredKey() {
		t.Error("expected HasConfiguredKey to return true after setting key")
	}
}

func TestModels_VisionSupport(t *testing.T) {
	models := GetModels()

	// Check that vision models have SupportsVision = true
	visionModels := map[string]bool{
		"llama-3.2-90b-vision-preview": true,
		"llama-3.2-11b-vision-preview": true,
	}

	for _, m := range models {
		if visionModels[m.ID] && !m.SupportsVision {
			t.Errorf("expected model '%s' to support vision", m.ID)
		}
		if !visionModels[m.ID] && m.SupportsVision && m.ID != "llama-3.2-90b-vision-preview" && m.ID != "llama-3.2-11b-vision-preview" {
			t.Errorf("model '%s' unexpectedly supports vision", m.ID)
		}
	}
}

func TestModels_ContextWindows(t *testing.T) {
	models := GetModels()

	for _, m := range models {
		if m.ContextWindow <= 0 {
			t.Errorf("model '%s' has invalid context window: %d", m.ID, m.ContextWindow)
		}
	}
}
