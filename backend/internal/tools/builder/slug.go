package builder

import (
	"fmt"
	"strings"
)

// ParsedSlug represents a parsed tool slug with provider and tool components.
type ParsedSlug struct {
	// Provider is the namespace/category of the tool (e.g., "sandbox", "posthog").
	Provider string
	// Tool is the specific tool name within the provider (e.g., "listFiles", "errors").
	Tool string
}

// String returns the string representation of the parsed slug.
func (p *ParsedSlug) String() string {
	return p.Provider + "/" + p.Tool
}

// ParseSlug parses a tool slug into its provider and tool components.
// Slugs must be in the format "provider/tool" (e.g., "sandbox/listFiles").
// Returns an error if the slug format is invalid.
func ParseSlug(slug string) (*ParsedSlug, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, fmt.Errorf("empty slug")
	}

	parts := strings.SplitN(slug, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid slug format %q: expected provider/tool", slug)
	}

	provider := strings.TrimSpace(parts[0])
	tool := strings.TrimSpace(parts[1])

	if provider == "" {
		return nil, fmt.Errorf("invalid slug %q: provider cannot be empty", slug)
	}
	if tool == "" {
		return nil, fmt.Errorf("invalid slug %q: tool cannot be empty", slug)
	}

	// Validate characters (alphanumeric, dash, underscore)
	if !isValidIdentifier(provider) {
		return nil, fmt.Errorf("invalid slug %q: provider contains invalid characters", slug)
	}
	if !isValidIdentifier(tool) {
		return nil, fmt.Errorf("invalid slug %q: tool contains invalid characters", slug)
	}

	return &ParsedSlug{
		Provider: provider,
		Tool:     tool,
	}, nil
}

// isValidIdentifier checks if a string contains only valid identifier characters.
// Valid characters are alphanumeric, dash, and underscore.
func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// MustParseSlug parses a slug and panics if invalid.
// This is useful for compile-time slug validation.
func MustParseSlug(slug string) *ParsedSlug {
	parsed, err := ParseSlug(slug)
	if err != nil {
		panic(err)
	}
	return parsed
}

// ValidateSlugs validates a list of slugs and returns any errors.
// Returns nil if all slugs are valid.
func ValidateSlugs(slugs []string) error {
	seen := make(map[string]bool)
	for _, slug := range slugs {
		parsed, err := ParseSlug(slug)
		if err != nil {
			return err
		}
		normalized := parsed.String()
		if seen[normalized] {
			return fmt.Errorf("duplicate slug: %s", normalized)
		}
		seen[normalized] = true
	}
	return nil
}

// NormalizeSlugs parses and normalizes a list of slugs.
// Returns the normalized slugs or an error if any slug is invalid.
func NormalizeSlugs(slugs []string) ([]string, error) {
	result := make([]string, 0, len(slugs))
	for _, slug := range slugs {
		parsed, err := ParseSlug(slug)
		if err != nil {
			return nil, err
		}
		result = append(result, parsed.String())
	}
	return result, nil
}
