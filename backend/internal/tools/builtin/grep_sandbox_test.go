package builtin

import (
	"testing"
)

func TestSanitizeGrepCommand(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		wantErr     bool
		errContains string
	}{
		// Valid commands
		{
			name:    "basic grep",
			command: `grep "pattern" .`,
			wantErr: false,
		},
		{
			name:    "grep with recursive flag",
			command: `grep -r "pattern" .`,
			wantErr: false,
		},
		{
			name:    "grep with multiple flags",
			command: `grep -rni "pattern" src/`,
			wantErr: false,
		},
		{
			name:    "grep with extended regex",
			command: `grep -E "foo|bar" .`,
			wantErr: false,
		},
		{
			name:    "grep with word match",
			command: `grep -w "function" .`,
			wantErr: false,
		},
		{
			name:    "grep with context flags",
			command: `grep -A 2 -B 2 "pattern" .`,
			wantErr: false,
		},
		{
			name:    "grep with include pattern",
			command: `grep --include="*.go" "pattern" .`,
			wantErr: false,
		},
		{
			name:    "grep with exclude-dir",
			command: `grep --exclude-dir=node_modules "pattern" .`,
			wantErr: false,
		},
		{
			name:    "grep with single quotes",
			command: `grep 'pattern with spaces' .`,
			wantErr: false,
		},

		// Invalid commands - must start with grep
		{
			name:        "not grep command",
			command:     `ls -la`,
			wantErr:     true,
			errContains: "must start with 'grep'",
		},
		{
			name:        "cat command",
			command:     `cat /etc/passwd`,
			wantErr:     true,
			errContains: "must start with 'grep'",
		},

		// Command injection attempts
		{
			name:        "semicolon injection",
			command:     `grep pattern; rm -rf /`,
			wantErr:     true,
			errContains: "dangerous",
		},
		{
			name:        "pipe injection",
			command:     `grep pattern | xargs rm`,
			wantErr:     true,
			errContains: "dangerous",
		},
		{
			name:        "ampersand injection",
			command:     `grep pattern && rm -rf /`,
			wantErr:     true,
			errContains: "dangerous",
		},
		{
			name:        "backtick injection",
			command:     "grep `whoami` .",
			wantErr:     true,
			errContains: "dangerous",
		},
		{
			name:        "command substitution",
			command:     `grep $(whoami) .`,
			wantErr:     true,
			errContains: "dangerous",
		},
		{
			name:        "variable expansion",
			command:     `grep ${HOME} .`,
			wantErr:     true,
			errContains: "dangerous",
		},
		{
			name:        "output redirection",
			command:     `grep pattern . > /tmp/out`,
			wantErr:     true,
			errContains: "dangerous",
		},
		{
			name:        "input redirection",
			command:     `grep pattern < /etc/passwd`,
			wantErr:     true,
			errContains: "dangerous",
		},
		{
			name:        "newline injection",
			command:     "grep pattern .\nrm -rf /",
			wantErr:     true,
			errContains: "dangerous",
		},

		// Disallowed flags
		{
			name:        "exec flag",
			command:     `grep --exec="rm" pattern .`,
			wantErr:     true,
			errContains: "not allowed",
		},
		{
			name:        "devices flag",
			command:     `grep -D read pattern .`,
			wantErr:     true,
			errContains: "not allowed",
		},

		// Malformed commands
		{
			name:        "unclosed single quote",
			command:     `grep 'unclosed`,
			wantErr:     true,
			errContains: "unclosed quote",
		},
		{
			name:        "unclosed double quote",
			command:     `grep "unclosed`,
			wantErr:     true,
			errContains: "unclosed quote",
		},
		{
			name:        "empty command",
			command:     ``,
			wantErr:     true,
			errContains: "empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sanitizeGrepCommand(tt.command)
			if tt.wantErr {
				if err == nil {
					t.Errorf("sanitizeGrepCommand() expected error containing %q, got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("sanitizeGrepCommand() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("sanitizeGrepCommand() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestParseCommandTokens(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    []string
		wantErr bool
	}{
		{
			name:    "simple command",
			command: "grep pattern .",
			want:    []string{"grep", "pattern", "."},
		},
		{
			name:    "double quoted string",
			command: `grep "hello world" .`,
			want:    []string{"grep", "hello world", "."},
		},
		{
			name:    "single quoted string",
			command: `grep 'hello world' .`,
			want:    []string{"grep", "hello world", "."},
		},
		{
			name:    "mixed quotes",
			command: `grep "foo's bar" .`,
			want:    []string{"grep", "foo's bar", "."},
		},
		{
			name:    "escaped quote",
			command: `grep hello\ world .`,
			want:    []string{"grep", "hello world", "."},
		},
		{
			name:    "multiple spaces",
			command: "grep   pattern    .",
			want:    []string{"grep", "pattern", "."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCommandTokens(tt.command)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseCommandTokens() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("parseCommandTokens() unexpected error = %v", err)
				return
			}
			if !sliceEqual(got, tt.want) {
				t.Errorf("parseCommandTokens() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseGrepLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want *SandboxGrepMatch
	}{
		{
			name: "standard output",
			line: "file.go:42:func main() {",
			want: &SandboxGrepMatch{
				File:       "file.go",
				LineNumber: 42,
				Content:    "func main() {",
			},
		},
		{
			name: "path with directories",
			line: "src/internal/tools/grep.go:15:package builtin",
			want: &SandboxGrepMatch{
				File:       "src/internal/tools/grep.go",
				LineNumber: 15,
				Content:    "package builtin",
			},
		},
		{
			name: "content with colons",
			line: "config.go:10:host := \"localhost:8080\"",
			want: &SandboxGrepMatch{
				File:       "config.go",
				LineNumber: 10,
				Content:    "host := \"localhost:8080\"",
			},
		},
		{
			name: "no line number format",
			line: "file.go:invalid:content",
			want: nil,
		},
		{
			name: "no colon",
			line: "just some text",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseGrepLine(tt.line)
			if tt.want == nil {
				if got != nil {
					t.Errorf("parseGrepLine() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Errorf("parseGrepLine() = nil, want %v", tt.want)
				return
			}
			if got.File != tt.want.File || got.LineNumber != tt.want.LineNumber || got.Content != tt.want.Content {
				t.Errorf("parseGrepLine() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseGrepOutput(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   SandboxGrepResult
	}{
		{
			name:   "empty output",
			output: "",
			want: SandboxGrepResult{
				Matches:    []SandboxGrepMatch{},
				TotalCount: 0,
				Truncated:  false,
			},
		},
		{
			name:   "single match",
			output: "file.go:10:func main() {\n",
			want: SandboxGrepResult{
				Matches: []SandboxGrepMatch{
					{File: "file.go", LineNumber: 10, Content: "func main() {"},
				},
				TotalCount: 1,
				Truncated:  false,
			},
		},
		{
			name:   "multiple matches",
			output: "file1.go:10:match1\nfile2.go:20:match2\nfile3.go:30:match3\n",
			want: SandboxGrepResult{
				Matches: []SandboxGrepMatch{
					{File: "file1.go", LineNumber: 10, Content: "match1"},
					{File: "file2.go", LineNumber: 20, Content: "match2"},
					{File: "file3.go", LineNumber: 30, Content: "match3"},
				},
				TotalCount: 3,
				Truncated:  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseGrepOutput(tt.output)
			if got.TotalCount != tt.want.TotalCount {
				t.Errorf("parseGrepOutput() TotalCount = %d, want %d", got.TotalCount, tt.want.TotalCount)
			}
			if got.Truncated != tt.want.Truncated {
				t.Errorf("parseGrepOutput() Truncated = %v, want %v", got.Truncated, tt.want.Truncated)
			}
			if len(got.Matches) != len(tt.want.Matches) {
				t.Errorf("parseGrepOutput() len(Matches) = %d, want %d", len(got.Matches), len(tt.want.Matches))
				return
			}
			for i, match := range got.Matches {
				wantMatch := tt.want.Matches[i]
				if match.File != wantMatch.File || match.LineNumber != wantMatch.LineNumber || match.Content != wantMatch.Content {
					t.Errorf("parseGrepOutput() Matches[%d] = %+v, want %+v", i, match, wantMatch)
				}
			}
		})
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
