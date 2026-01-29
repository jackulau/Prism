package builder

import (
	"context"
	"testing"

	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
	"github.com/jacklau/prism/internal/tools"
)

// mockTool is a simple mock implementation of tools.Tool for testing
type mockTool struct {
	name string
}

func (t *mockTool) Name() string                     { return t.name }
func (t *mockTool) Description() string              { return "Mock tool: " + t.name }
func (t *mockTool) Parameters() llm.JSONSchema       { return llm.JSONSchema{Type: "object"} }
func (t *mockTool) RequiresConfirmation() bool       { return false }
func (t *mockTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return map[string]string{"tool": t.name}, nil
}

// mockBuilder is a simple builder for testing
type mockBuilder struct {
	slug    string
	deps    []string
	toolName string
}

func (b *mockBuilder) Slug() string                           { return b.slug }
func (b *mockBuilder) DependsOn() []string                    { return b.deps }
func (b *mockBuilder) Build(sandbox *sandbox.Service) tools.Tool { return &mockTool{name: b.toolName} }

func newMockBuilder(slug string, deps []string) *mockBuilder {
	return &mockBuilder{
		slug:     slug,
		deps:     deps,
		toolName: slug,
	}
}

func TestBaseBuilder_DependsOn(t *testing.T) {
	b := &BaseBuilder{}
	deps := b.DependsOn()
	if deps != nil {
		t.Errorf("BaseBuilder.DependsOn() = %v, want nil", deps)
	}
}

func TestBuilderRegistry_Register(t *testing.T) {
	registry := NewBuilderRegistry()

	builder := newMockBuilder("test/tool", nil)
	err := registry.Register(builder)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Try to register the same slug again
	err = registry.Register(builder)
	if err == nil {
		t.Error("Register() expected error for duplicate slug, got nil")
	}
}

func TestBuilderRegistry_Get(t *testing.T) {
	registry := NewBuilderRegistry()
	builder := newMockBuilder("test/tool", nil)
	registry.Register(builder)

	// Test found case
	got, ok := registry.Get("test/tool")
	if !ok {
		t.Error("Get() returned false for existing builder")
	}
	if got.Slug() != "test/tool" {
		t.Errorf("Get() returned wrong builder, got slug %s", got.Slug())
	}

	// Test not found case
	_, ok = registry.Get("nonexistent/tool")
	if ok {
		t.Error("Get() returned true for non-existent builder")
	}
}

func TestBuilderRegistry_List(t *testing.T) {
	registry := NewBuilderRegistry()
	registry.Register(newMockBuilder("test/tool1", nil))
	registry.Register(newMockBuilder("test/tool2", nil))
	registry.Register(newMockBuilder("test/tool3", nil))

	builders := registry.List()
	if len(builders) != 3 {
		t.Errorf("List() returned %d builders, want 3", len(builders))
	}
}

func TestBuilderRegistry_Slugs(t *testing.T) {
	registry := NewBuilderRegistry()
	registry.Register(newMockBuilder("test/tool1", nil))
	registry.Register(newMockBuilder("test/tool2", nil))

	slugs := registry.Slugs()
	if len(slugs) != 2 {
		t.Errorf("Slugs() returned %d slugs, want 2", len(slugs))
	}

	// Check that expected slugs are present
	slugMap := make(map[string]bool)
	for _, s := range slugs {
		slugMap[s] = true
	}
	if !slugMap["test/tool1"] || !slugMap["test/tool2"] {
		t.Errorf("Slugs() missing expected slugs, got %v", slugs)
	}
}

func TestBuilderRegistry_Has(t *testing.T) {
	registry := NewBuilderRegistry()
	registry.Register(newMockBuilder("test/tool", nil))

	if !registry.Has("test/tool") {
		t.Error("Has() returned false for existing slug")
	}
	if registry.Has("nonexistent/tool") {
		t.Error("Has() returned true for non-existent slug")
	}
}

func TestBuilderRegistry_Count(t *testing.T) {
	registry := NewBuilderRegistry()
	if registry.Count() != 0 {
		t.Errorf("Count() = %d, want 0 for empty registry", registry.Count())
	}

	registry.Register(newMockBuilder("test/tool1", nil))
	registry.Register(newMockBuilder("test/tool2", nil))

	if registry.Count() != 2 {
		t.Errorf("Count() = %d, want 2", registry.Count())
	}
}

func TestBuilderRegistry_Unregister(t *testing.T) {
	registry := NewBuilderRegistry()
	registry.Register(newMockBuilder("test/tool", nil))

	// Unregister existing builder
	if !registry.Unregister("test/tool") {
		t.Error("Unregister() returned false for existing builder")
	}
	if registry.Has("test/tool") {
		t.Error("Builder still exists after Unregister()")
	}

	// Unregister non-existent builder
	if registry.Unregister("nonexistent/tool") {
		t.Error("Unregister() returned true for non-existent builder")
	}
}

func TestBuilderRegistry_MustRegister(t *testing.T) {
	registry := NewBuilderRegistry()

	// Should not panic for valid registration
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustRegister() panicked unexpectedly: %v", r)
		}
	}()
	registry.MustRegister(newMockBuilder("test/tool", nil))

	// Should panic for duplicate
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustRegister() did not panic for duplicate registration")
		}
	}()
	registry.MustRegister(newMockBuilder("test/tool", nil))
}
