package builder

import (
	"testing"
)

func TestParseSlug(t *testing.T) {
	tests := []struct {
		name    string
		slug    string
		want    *ParsedSlug
		wantErr bool
	}{
		{
			name: "valid simple slug",
			slug: "sandbox/listFiles",
			want: &ParsedSlug{Provider: "sandbox", Tool: "listFiles"},
		},
		{
			name: "valid slug with dashes",
			slug: "my-provider/my-tool",
			want: &ParsedSlug{Provider: "my-provider", Tool: "my-tool"},
		},
		{
			name: "valid slug with underscores",
			slug: "my_provider/my_tool",
			want: &ParsedSlug{Provider: "my_provider", Tool: "my_tool"},
		},
		{
			name: "valid slug with numbers",
			slug: "provider1/tool2",
			want: &ParsedSlug{Provider: "provider1", Tool: "tool2"},
		},
		{
			name: "valid slug with whitespace (trimmed)",
			slug: "  sandbox/file_read  ",
			want: &ParsedSlug{Provider: "sandbox", Tool: "file_read"},
		},
		{
			name:    "empty slug",
			slug:    "",
			wantErr: true,
		},
		{
			name:    "no separator",
			slug:    "sandboxlistFiles",
			wantErr: true,
		},
		{
			name:    "empty provider",
			slug:    "/tool",
			wantErr: true,
		},
		{
			name:    "empty tool",
			slug:    "provider/",
			wantErr: true,
		},
		{
			name:    "invalid characters in provider",
			slug:    "pro@vider/tool",
			wantErr: true,
		},
		{
			name:    "invalid characters in tool",
			slug:    "provider/to.ol",
			wantErr: true,
		},
		{
			name:    "multiple separators",
			slug:    "sandbox/sub/tool",
			want:    &ParsedSlug{Provider: "sandbox", Tool: "sub/tool"},
			wantErr: true, // sub/tool has invalid character '/'
		},
		{
			name:    "whitespace only",
			slug:    "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSlug(tt.slug)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSlug() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Provider != tt.want.Provider || got.Tool != tt.want.Tool {
					t.Errorf("ParseSlug() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestParsedSlug_String(t *testing.T) {
	slug := &ParsedSlug{Provider: "sandbox", Tool: "file_read"}
	if slug.String() != "sandbox/file_read" {
		t.Errorf("ParsedSlug.String() = %s, want sandbox/file_read", slug.String())
	}
}

func TestMustParseSlug(t *testing.T) {
	// Valid slug should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustParseSlug() panicked for valid slug: %v", r)
		}
	}()
	parsed := MustParseSlug("sandbox/file_read")
	if parsed.Provider != "sandbox" || parsed.Tool != "file_read" {
		t.Errorf("MustParseSlug() returned wrong values: %v", parsed)
	}
}

func TestMustParseSlug_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParseSlug() did not panic for invalid slug")
		}
	}()
	MustParseSlug("invalid")
}

func TestValidateSlugs(t *testing.T) {
	tests := []struct {
		name    string
		slugs   []string
		wantErr bool
	}{
		{
			name:    "valid slugs",
			slugs:   []string{"sandbox/file_read", "sandbox/file_write"},
			wantErr: false,
		},
		{
			name:    "empty list",
			slugs:   []string{},
			wantErr: false,
		},
		{
			name:    "invalid slug in list",
			slugs:   []string{"sandbox/file_read", "invalid"},
			wantErr: true,
		},
		{
			name:    "duplicate slugs",
			slugs:   []string{"sandbox/file_read", "sandbox/file_read"},
			wantErr: true,
		},
		{
			name:    "normalized duplicates",
			slugs:   []string{"sandbox/file_read", "  sandbox/file_read  "},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSlugs(tt.slugs)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSlugs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeSlugs(t *testing.T) {
	tests := []struct {
		name    string
		slugs   []string
		want    []string
		wantErr bool
	}{
		{
			name:  "already normalized",
			slugs: []string{"sandbox/file_read", "sandbox/file_write"},
			want:  []string{"sandbox/file_read", "sandbox/file_write"},
		},
		{
			name:  "with whitespace",
			slugs: []string{"  sandbox/file_read  ", "sandbox/file_write"},
			want:  []string{"sandbox/file_read", "sandbox/file_write"},
		},
		{
			name:    "invalid slug",
			slugs:   []string{"invalid"},
			wantErr: true,
		},
		{
			name:  "empty list",
			slugs: []string{},
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeSlugs(tt.slugs)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeSlugs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("NormalizeSlugs() returned %d slugs, want %d", len(got), len(tt.want))
					return
				}
				for i, slug := range got {
					if slug != tt.want[i] {
						t.Errorf("NormalizeSlugs()[%d] = %s, want %s", i, slug, tt.want[i])
					}
				}
			}
		})
	}
}

func TestIsValidIdentifier(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"lowercase letters", "abc", true},
		{"uppercase letters", "ABC", true},
		{"mixed case", "AbC", true},
		{"numbers", "123", true},
		{"alphanumeric", "abc123", true},
		{"with dash", "abc-123", true},
		{"with underscore", "abc_123", true},
		{"empty", "", false},
		{"with dot", "abc.123", false},
		{"with space", "abc 123", false},
		{"with slash", "abc/123", false},
		{"with at", "abc@123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidIdentifier(tt.input); got != tt.want {
				t.Errorf("isValidIdentifier(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
