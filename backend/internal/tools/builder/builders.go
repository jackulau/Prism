package builder

import (
	"database/sql"

	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
	"github.com/jacklau/prism/internal/services/coderunner"
	"github.com/jacklau/prism/internal/tools"
)

// BuilderDependencies contains all dependencies that tool builders may need.
// Not all builders require all dependencies.
type BuilderDependencies struct {
	Sandbox         *sandbox.Service
	CodeRunner      *coderunner.Runner
	DB              *sql.DB
	FileHistoryRepo *repository.FileHistoryRepository
	TodoRepo        *repository.TodoRepository
	LLMProvider     llm.Provider

	// API keys and configuration
	SerpAPIKey     string
	GoogleAPIKey   string
	GoogleSearchCX string
	OpenAIAPIKey   string
}

// ToolBuilderFunc is a function type that builds a tool with dependencies.
type ToolBuilderFunc func(deps *BuilderDependencies) tools.Tool

// FuncBuilder wraps a builder function as a ToolBuilder.
type FuncBuilder struct {
	slug     string
	deps     []string
	buildFn  ToolBuilderFunc
	buildDep *BuilderDependencies
}

// NewFuncBuilder creates a new function-based builder.
func NewFuncBuilder(slug string, deps []string, fn ToolBuilderFunc) *FuncBuilder {
	return &FuncBuilder{
		slug:    slug,
		deps:    deps,
		buildFn: fn,
	}
}

// Slug returns the builder's slug.
func (b *FuncBuilder) Slug() string {
	return b.slug
}

// DependsOn returns the builder's dependencies.
func (b *FuncBuilder) DependsOn() []string {
	return b.deps
}

// Build creates the tool using the stored dependencies.
func (b *FuncBuilder) Build(sandbox *sandbox.Service) tools.Tool {
	if b.buildDep == nil {
		b.buildDep = &BuilderDependencies{Sandbox: sandbox}
	} else {
		b.buildDep.Sandbox = sandbox
	}
	return b.buildFn(b.buildDep)
}

// SetDependencies sets the builder dependencies.
func (b *FuncBuilder) SetDependencies(deps *BuilderDependencies) {
	b.buildDep = deps
}

// SimpleBuilder wraps a simple tool constructor that only needs sandbox.
type SimpleBuilder struct {
	BaseBuilder
	slug    string
	buildFn func(*sandbox.Service) tools.Tool
}

// NewSimpleBuilder creates a builder for tools that only need sandbox.
func NewSimpleBuilder(slug string, fn func(*sandbox.Service) tools.Tool) *SimpleBuilder {
	return &SimpleBuilder{
		slug:    slug,
		buildFn: fn,
	}
}

// Slug returns the builder's slug.
func (b *SimpleBuilder) Slug() string {
	return b.slug
}

// Build creates the tool.
func (b *SimpleBuilder) Build(sandbox *sandbox.Service) tools.Tool {
	return b.buildFn(sandbox)
}

// BuilderWithDeps creates a tool builder that requires additional dependencies.
// This is a more flexible version that stores dependencies for later use.
type BuilderWithDeps struct {
	slug        string
	deps        []string
	buildFn     ToolBuilderFunc
	storedDeps  *BuilderDependencies
}

// NewBuilderWithDeps creates a new builder with dependencies.
func NewBuilderWithDeps(slug string, toolDeps []string, fn ToolBuilderFunc) *BuilderWithDeps {
	return &BuilderWithDeps{
		slug:    slug,
		deps:    toolDeps,
		buildFn: fn,
	}
}

// Slug returns the builder's slug.
func (b *BuilderWithDeps) Slug() string {
	return b.slug
}

// DependsOn returns the builder's tool dependencies.
func (b *BuilderWithDeps) DependsOn() []string {
	return b.deps
}

// Build creates the tool.
func (b *BuilderWithDeps) Build(sandbox *sandbox.Service) tools.Tool {
	deps := b.storedDeps
	if deps == nil {
		deps = &BuilderDependencies{}
	}
	deps.Sandbox = sandbox
	return b.buildFn(deps)
}

// SetDependencies configures the builder's dependencies.
func (b *BuilderWithDeps) SetDependencies(deps *BuilderDependencies) {
	b.storedDeps = deps
}

// DependencyConfigurable is an interface for builders that can have their
// dependencies configured.
type DependencyConfigurable interface {
	SetDependencies(deps *BuilderDependencies)
}

// ConfigureBuilder sets dependencies on a builder if it supports configuration.
func ConfigureBuilder(builder ToolBuilder, deps *BuilderDependencies) {
	if configurable, ok := builder.(DependencyConfigurable); ok {
		configurable.SetDependencies(deps)
	}
}

// ConfigureRegistry configures all builders in a registry with dependencies.
func ConfigureRegistry(registry *BuilderRegistry, deps *BuilderDependencies) {
	for _, builder := range registry.List() {
		ConfigureBuilder(builder, deps)
	}
}
