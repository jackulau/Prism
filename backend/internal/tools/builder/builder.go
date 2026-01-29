// Package builder provides dynamic tool building capabilities for runtime
// tool collection assembly based on agent configuration and tool slugs.
package builder

import (
	"github.com/jacklau/prism/internal/sandbox"
	"github.com/jacklau/prism/internal/tools"
)

// ToolBuilder defines the interface for building tools dynamically.
// Each builder is responsible for creating a specific tool type.
type ToolBuilder interface {
	// Slug returns the unique identifier for this tool builder.
	// Slugs follow the format "provider/tool" (e.g., "sandbox/listFiles").
	Slug() string

	// Build creates a Tool instance with the given sandbox.
	// The sandbox provides isolated file and command execution capabilities.
	Build(sandbox *sandbox.Service) tools.Tool

	// DependsOn returns slugs of tools this builder depends on.
	// Dependencies are resolved using topological sort before building.
	DependsOn() []string
}

// BaseBuilder provides a default implementation for common builder functionality.
// Embed this struct to avoid implementing DependsOn() for tools without dependencies.
type BaseBuilder struct{}

// DependsOn returns an empty slice indicating no dependencies.
func (b *BaseBuilder) DependsOn() []string {
	return nil
}
