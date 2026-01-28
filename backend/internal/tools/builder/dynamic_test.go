package builder

import (
	"testing"
)

func TestBuildDynamicTools_Empty(t *testing.T) {
	registry := NewBuilderRegistry()

	result, err := BuildDynamicTools("agent-1", []string{}, nil, registry)
	if err != nil {
		t.Fatalf("BuildDynamicTools() error = %v", err)
	}
	if result == nil {
		t.Fatal("BuildDynamicTools() returned nil registry")
	}

	tools := result.List()
	if len(tools) != 0 {
		t.Errorf("Expected empty registry, got %d tools", len(tools))
	}
}

func TestBuildDynamicTools_SingleTool(t *testing.T) {
	registry := NewBuilderRegistry()
	registry.Register(newMockBuilder("sandbox/file_read", nil))

	result, err := BuildDynamicTools("agent-1", []string{"sandbox/file_read"}, nil, registry)
	if err != nil {
		t.Fatalf("BuildDynamicTools() error = %v", err)
	}

	tools := result.List()
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}
}

func TestBuildDynamicTools_MultipleTools(t *testing.T) {
	registry := NewBuilderRegistry()
	registry.Register(newMockBuilder("sandbox/file_read", nil))
	registry.Register(newMockBuilder("sandbox/file_write", nil))
	registry.Register(newMockBuilder("sandbox/file_list", nil))

	result, err := BuildDynamicTools("agent-1", []string{
		"sandbox/file_read",
		"sandbox/file_write",
	}, nil, registry)
	if err != nil {
		t.Fatalf("BuildDynamicTools() error = %v", err)
	}

	tools := result.List()
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}
}

func TestBuildDynamicTools_InvalidSlug(t *testing.T) {
	registry := NewBuilderRegistry()
	registry.Register(newMockBuilder("sandbox/file_read", nil))

	_, err := BuildDynamicTools("agent-1", []string{"invalid_slug"}, nil, registry)
	if err == nil {
		t.Error("Expected error for invalid slug, got nil")
	}
}

func TestBuildDynamicTools_MissingBuilder(t *testing.T) {
	registry := NewBuilderRegistry()
	registry.Register(newMockBuilder("sandbox/file_read", nil))

	_, err := BuildDynamicTools("agent-1", []string{"sandbox/nonexistent"}, nil, registry)
	if err == nil {
		t.Error("Expected error for missing builder, got nil")
	}
}

func TestBuildDynamicTools_WithDependencies(t *testing.T) {
	registry := NewBuilderRegistry()
	// tool2 depends on tool1
	registry.Register(newMockBuilder("test/tool1", nil))
	registry.Register(newMockBuilder("test/tool2", []string{"test/tool1"}))

	result, err := BuildDynamicTools("agent-1", []string{"test/tool2"}, nil, registry)
	if err != nil {
		t.Fatalf("BuildDynamicTools() error = %v", err)
	}

	tools := result.List()
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools (including dependency), got %d", len(tools))
	}
}

func TestBuildDynamicTools_TransitiveDependencies(t *testing.T) {
	registry := NewBuilderRegistry()
	// tool3 depends on tool2, which depends on tool1
	registry.Register(newMockBuilder("test/tool1", nil))
	registry.Register(newMockBuilder("test/tool2", []string{"test/tool1"}))
	registry.Register(newMockBuilder("test/tool3", []string{"test/tool2"}))

	result, err := BuildDynamicTools("agent-1", []string{"test/tool3"}, nil, registry)
	if err != nil {
		t.Fatalf("BuildDynamicTools() error = %v", err)
	}

	tools := result.List()
	if len(tools) != 3 {
		t.Errorf("Expected 3 tools (including transitive dependencies), got %d", len(tools))
	}
}

func TestBuildDynamicTools_CircularDependency(t *testing.T) {
	registry := NewBuilderRegistry()
	// Circular: tool1 depends on tool2, tool2 depends on tool1
	registry.Register(newMockBuilder("test/tool1", []string{"test/tool2"}))
	registry.Register(newMockBuilder("test/tool2", []string{"test/tool1"}))

	_, err := BuildDynamicTools("agent-1", []string{"test/tool1"}, nil, registry)
	if err == nil {
		t.Error("Expected error for circular dependency, got nil")
	}
}

func TestBuildDynamicTools_DuplicateSlugs(t *testing.T) {
	registry := NewBuilderRegistry()
	registry.Register(newMockBuilder("sandbox/file_read", nil))

	_, err := BuildDynamicTools("agent-1", []string{
		"sandbox/file_read",
		"sandbox/file_read",
	}, nil, registry)
	if err == nil {
		t.Error("Expected error for duplicate slugs, got nil")
	}
}

func TestTopologicalSort_NoDependencies(t *testing.T) {
	registry := NewBuilderRegistry()
	registry.Register(newMockBuilder("test/tool1", nil))
	registry.Register(newMockBuilder("test/tool2", nil))
	registry.Register(newMockBuilder("test/tool3", nil))

	slugs := []string{"test/tool1", "test/tool2", "test/tool3"}
	sorted, err := topologicalSort(slugs, registry)
	if err != nil {
		t.Fatalf("topologicalSort() error = %v", err)
	}

	if len(sorted) != 3 {
		t.Errorf("topologicalSort() returned %d items, want 3", len(sorted))
	}
}

func TestTopologicalSort_LinearDependencies(t *testing.T) {
	registry := NewBuilderRegistry()
	// tool3 -> tool2 -> tool1
	registry.Register(newMockBuilder("test/tool1", nil))
	registry.Register(newMockBuilder("test/tool2", []string{"test/tool1"}))
	registry.Register(newMockBuilder("test/tool3", []string{"test/tool2"}))

	slugs := []string{"test/tool1", "test/tool2", "test/tool3"}
	sorted, err := topologicalSort(slugs, registry)
	if err != nil {
		t.Fatalf("topologicalSort() error = %v", err)
	}

	// tool1 must come before tool2, tool2 must come before tool3
	indexOf := make(map[string]int)
	for i, s := range sorted {
		indexOf[s] = i
	}

	if indexOf["test/tool1"] > indexOf["test/tool2"] {
		t.Error("tool1 should come before tool2")
	}
	if indexOf["test/tool2"] > indexOf["test/tool3"] {
		t.Error("tool2 should come before tool3")
	}
}

func TestTopologicalSort_DiamondDependencies(t *testing.T) {
	registry := NewBuilderRegistry()
	// Diamond: tool4 depends on tool2 and tool3, both depend on tool1
	//       tool4
	//      /     \
	//   tool2   tool3
	//      \     /
	//       tool1
	registry.Register(newMockBuilder("test/tool1", nil))
	registry.Register(newMockBuilder("test/tool2", []string{"test/tool1"}))
	registry.Register(newMockBuilder("test/tool3", []string{"test/tool1"}))
	registry.Register(newMockBuilder("test/tool4", []string{"test/tool2", "test/tool3"}))

	slugs := []string{"test/tool1", "test/tool2", "test/tool3", "test/tool4"}
	sorted, err := topologicalSort(slugs, registry)
	if err != nil {
		t.Fatalf("topologicalSort() error = %v", err)
	}

	indexOf := make(map[string]int)
	for i, s := range sorted {
		indexOf[s] = i
	}

	// tool1 must come before tool2 and tool3
	if indexOf["test/tool1"] > indexOf["test/tool2"] || indexOf["test/tool1"] > indexOf["test/tool3"] {
		t.Error("tool1 should come before tool2 and tool3")
	}
	// tool2 and tool3 must come before tool4
	if indexOf["test/tool2"] > indexOf["test/tool4"] || indexOf["test/tool3"] > indexOf["test/tool4"] {
		t.Error("tool2 and tool3 should come before tool4")
	}
}

func TestBuildDynamicToolsWithOptions_IgnoreMissing(t *testing.T) {
	registry := NewBuilderRegistry()
	registry.Register(newMockBuilder("sandbox/file_read", nil))

	opts := BuildOptions{
		IgnoreMissing:       true,
		IncludeDependencies: true,
	}

	result, err := BuildDynamicToolsWithOptions("agent-1", []string{
		"sandbox/file_read",
		"sandbox/nonexistent",
	}, nil, registry, opts)
	if err != nil {
		t.Fatalf("BuildDynamicToolsWithOptions() error = %v", err)
	}

	tools := result.List()
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool (ignoring missing), got %d", len(tools))
	}
}

func TestBuildDynamicToolsWithOptions_NoDependencies(t *testing.T) {
	registry := NewBuilderRegistry()
	// tool2 depends on tool1
	registry.Register(newMockBuilder("test/tool1", nil))
	registry.Register(newMockBuilder("test/tool2", []string{"test/tool1"}))

	opts := BuildOptions{
		IgnoreMissing:       false,
		IncludeDependencies: false,
	}

	result, err := BuildDynamicToolsWithOptions("agent-1", []string{"test/tool2"}, nil, registry, opts)
	if err != nil {
		t.Fatalf("BuildDynamicToolsWithOptions() error = %v", err)
	}

	tools := result.List()
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool (no dependencies), got %d", len(tools))
	}
}

func TestCollectWithDependencies(t *testing.T) {
	registry := NewBuilderRegistry()
	registry.Register(newMockBuilder("test/tool1", nil))
	registry.Register(newMockBuilder("test/tool2", []string{"test/tool1"}))
	registry.Register(newMockBuilder("test/tool3", []string{"test/tool2"}))
	registry.Register(newMockBuilder("test/tool4", nil)) // Independent tool

	slugs := []string{"test/tool3"}
	collected, err := collectWithDependencies(slugs, registry)
	if err != nil {
		t.Fatalf("collectWithDependencies() error = %v", err)
	}

	// Should have tool1, tool2, tool3 (but not tool4)
	if len(collected) != 3 {
		t.Errorf("collectWithDependencies() returned %d slugs, want 3", len(collected))
	}

	collectedSet := make(map[string]bool)
	for _, s := range collected {
		collectedSet[s] = true
	}

	if !collectedSet["test/tool1"] || !collectedSet["test/tool2"] || !collectedSet["test/tool3"] {
		t.Errorf("collectWithDependencies() missing expected slugs, got %v", collected)
	}
	if collectedSet["test/tool4"] {
		t.Error("collectWithDependencies() should not include unrelated tool")
	}
}

func TestCollectWithDependencies_MissingDependency(t *testing.T) {
	registry := NewBuilderRegistry()
	registry.Register(newMockBuilder("test/tool2", []string{"test/tool1"})) // tool1 not registered

	slugs := []string{"test/tool2"}
	_, err := collectWithDependencies(slugs, registry)
	if err == nil {
		t.Error("Expected error for missing dependency, got nil")
	}
}
