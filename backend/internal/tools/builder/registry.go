package builder

import (
	"fmt"
	"sync"
)

// BuilderRegistry manages the collection of available tool builders.
// It provides thread-safe registration and lookup of builders by slug.
type BuilderRegistry struct {
	builders map[string]ToolBuilder
	mu       sync.RWMutex
}

// NewBuilderRegistry creates a new builder registry.
func NewBuilderRegistry() *BuilderRegistry {
	return &BuilderRegistry{
		builders: make(map[string]ToolBuilder),
	}
}

// Register adds a tool builder to the registry.
// Returns an error if a builder with the same slug is already registered.
func (r *BuilderRegistry) Register(builder ToolBuilder) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	slug := builder.Slug()
	if _, exists := r.builders[slug]; exists {
		return fmt.Errorf("builder with slug %q already registered", slug)
	}

	r.builders[slug] = builder
	return nil
}

// MustRegister adds a tool builder to the registry and panics on error.
// This is useful for registering builders during package initialization.
func (r *BuilderRegistry) MustRegister(builder ToolBuilder) {
	if err := r.Register(builder); err != nil {
		panic(err)
	}
}

// Get returns a builder by its slug.
// Returns the builder and true if found, nil and false otherwise.
func (r *BuilderRegistry) Get(slug string) (ToolBuilder, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	builder, ok := r.builders[slug]
	return builder, ok
}

// List returns all registered builders.
func (r *BuilderRegistry) List() []ToolBuilder {
	r.mu.RLock()
	defer r.mu.RUnlock()

	builders := make([]ToolBuilder, 0, len(r.builders))
	for _, builder := range r.builders {
		builders = append(builders, builder)
	}
	return builders
}

// Slugs returns a list of all registered builder slugs.
func (r *BuilderRegistry) Slugs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	slugs := make([]string, 0, len(r.builders))
	for slug := range r.builders {
		slugs = append(slugs, slug)
	}
	return slugs
}

// Has checks if a builder with the given slug is registered.
func (r *BuilderRegistry) Has(slug string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.builders[slug]
	return ok
}

// Count returns the number of registered builders.
func (r *BuilderRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.builders)
}

// Unregister removes a builder from the registry.
// Returns true if the builder was found and removed, false otherwise.
func (r *BuilderRegistry) Unregister(slug string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.builders[slug]; !exists {
		return false
	}

	delete(r.builders, slug)
	return true
}
