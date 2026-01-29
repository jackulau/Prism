package builder

import (
	"fmt"

	"github.com/jacklau/prism/internal/sandbox"
	"github.com/jacklau/prism/internal/tools"
)

// BuildDynamicTools assembles a tool collection for an agent based on the
// provided tool slugs. It resolves dependencies using topological sort,
// builds the tools, and registers them in a new registry.
//
// Parameters:
//   - agentID: The ID of the agent requesting the tools (for logging/tracking)
//   - toolSlugs: List of tool slugs to include (e.g., ["sandbox/listFiles", "sandbox/readFile"])
//   - sandbox: The sandbox service for file/command operations
//   - builderRegistry: Registry containing available tool builders
//
// Returns a new tools.Registry with the assembled tools, or an error if
// any slugs are invalid, builders are missing, or circular dependencies exist.
func BuildDynamicTools(
	agentID string,
	toolSlugs []string,
	sandbox *sandbox.Service,
	builderRegistry *BuilderRegistry,
) (*tools.Registry, error) {
	if len(toolSlugs) == 0 {
		return tools.NewRegistry(), nil
	}

	// Parse and validate all slugs
	if err := ValidateSlugs(toolSlugs); err != nil {
		return nil, fmt.Errorf("invalid tool slugs: %w", err)
	}

	// Normalize slugs
	normalizedSlugs, err := NormalizeSlugs(toolSlugs)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize slugs: %w", err)
	}

	// Collect all builders including dependencies
	allSlugs, err := collectWithDependencies(normalizedSlugs, builderRegistry)
	if err != nil {
		return nil, err
	}

	// Perform topological sort to resolve build order
	sortedSlugs, err := topologicalSort(allSlugs, builderRegistry)
	if err != nil {
		return nil, fmt.Errorf("dependency resolution failed: %w", err)
	}

	// Build tools in dependency order
	registry := tools.NewRegistry()
	for _, slug := range sortedSlugs {
		builder, ok := builderRegistry.Get(slug)
		if !ok {
			return nil, fmt.Errorf("builder not found for slug %q", slug)
		}

		tool := builder.Build(sandbox)
		if err := registry.Register(tool); err != nil {
			return nil, fmt.Errorf("failed to register tool %q: %w", slug, err)
		}
	}

	return registry, nil
}

// collectWithDependencies expands the requested slugs to include all dependencies.
func collectWithDependencies(slugs []string, registry *BuilderRegistry) ([]string, error) {
	allSlugs := make(map[string]bool)
	queue := make([]string, len(slugs))
	copy(queue, slugs)

	for len(queue) > 0 {
		// Pop first item
		slug := queue[0]
		queue = queue[1:]

		if allSlugs[slug] {
			continue
		}

		builder, ok := registry.Get(slug)
		if !ok {
			return nil, fmt.Errorf("builder not found for slug %q", slug)
		}

		allSlugs[slug] = true

		// Add dependencies to queue
		for _, dep := range builder.DependsOn() {
			if !allSlugs[dep] {
				queue = append(queue, dep)
			}
		}
	}

	result := make([]string, 0, len(allSlugs))
	for slug := range allSlugs {
		result = append(result, slug)
	}
	return result, nil
}

// topologicalSort performs Kahn's algorithm for topological sorting.
// Returns slugs in dependency order (dependencies first).
func topologicalSort(slugs []string, registry *BuilderRegistry) ([]string, error) {
	// Build adjacency list and in-degree count
	inDegree := make(map[string]int)
	dependents := make(map[string][]string) // slug -> list of slugs that depend on it

	slugSet := make(map[string]bool)
	for _, slug := range slugs {
		slugSet[slug] = true
		inDegree[slug] = 0
	}

	// Calculate in-degrees
	for _, slug := range slugs {
		builder, _ := registry.Get(slug)
		for _, dep := range builder.DependsOn() {
			if !slugSet[dep] {
				// Dependency not in our set, skip
				continue
			}
			inDegree[slug]++
			dependents[dep] = append(dependents[dep], slug)
		}
	}

	// Initialize queue with nodes having no dependencies
	queue := make([]string, 0)
	for slug, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, slug)
		}
	}

	// Process queue
	var sorted []string
	for len(queue) > 0 {
		// Pop first item
		slug := queue[0]
		queue = queue[1:]
		sorted = append(sorted, slug)

		// Reduce in-degree for dependents
		for _, dependent := range dependents[slug] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// Check for cycle
	if len(sorted) != len(slugs) {
		// Find nodes in cycle
		var cycleNodes []string
		for slug, degree := range inDegree {
			if degree > 0 {
				cycleNodes = append(cycleNodes, slug)
			}
		}
		return nil, fmt.Errorf("circular dependency detected involving: %v", cycleNodes)
	}

	return sorted, nil
}

// BuildOptions provides configuration for building dynamic tools.
type BuildOptions struct {
	// IgnoreMissing skips missing builders instead of returning an error.
	IgnoreMissing bool
	// IncludeDependencies automatically includes dependencies of requested tools.
	IncludeDependencies bool
}

// DefaultBuildOptions returns the default build options.
func DefaultBuildOptions() BuildOptions {
	return BuildOptions{
		IgnoreMissing:       false,
		IncludeDependencies: true,
	}
}

// BuildDynamicToolsWithOptions assembles a tool collection with custom options.
func BuildDynamicToolsWithOptions(
	agentID string,
	toolSlugs []string,
	sandbox *sandbox.Service,
	builderRegistry *BuilderRegistry,
	opts BuildOptions,
) (*tools.Registry, error) {
	if len(toolSlugs) == 0 {
		return tools.NewRegistry(), nil
	}

	// Parse and validate all slugs (collect valid ones if IgnoreMissing)
	validSlugs := make([]string, 0, len(toolSlugs))
	for _, slug := range toolSlugs {
		parsed, err := ParseSlug(slug)
		if err != nil {
			if opts.IgnoreMissing {
				continue
			}
			return nil, fmt.Errorf("invalid slug %q: %w", slug, err)
		}

		normalized := parsed.String()
		if !builderRegistry.Has(normalized) {
			if opts.IgnoreMissing {
				continue
			}
			return nil, fmt.Errorf("builder not found for slug %q", normalized)
		}

		validSlugs = append(validSlugs, normalized)
	}

	if len(validSlugs) == 0 {
		return tools.NewRegistry(), nil
	}

	// Collect dependencies if requested
	var allSlugs []string
	if opts.IncludeDependencies {
		var err error
		allSlugs, err = collectWithDependencies(validSlugs, builderRegistry)
		if err != nil {
			return nil, err
		}
	} else {
		allSlugs = validSlugs
	}

	// Perform topological sort
	sortedSlugs, err := topologicalSort(allSlugs, builderRegistry)
	if err != nil {
		return nil, fmt.Errorf("dependency resolution failed: %w", err)
	}

	// Build tools
	registry := tools.NewRegistry()
	for _, slug := range sortedSlugs {
		builder, ok := builderRegistry.Get(slug)
		if !ok {
			continue // Should not happen after validation
		}

		tool := builder.Build(sandbox)
		if err := registry.Register(tool); err != nil {
			return nil, fmt.Errorf("failed to register tool %q: %w", slug, err)
		}
	}

	return registry, nil
}
