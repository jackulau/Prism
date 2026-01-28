package builtin

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/sandbox"
)

const (
	maxGrepMatches = 1000 // Maximum number of matches to return
	grepTimeout    = 30 * time.Second
)

// allowedGrepFlags is a whitelist of safe grep flags
var allowedGrepFlags = map[string]bool{
	"-r": true, "-R": true, // recursive
	"-n":            true, // line numbers
	"-i":            true, // case insensitive
	"-l":            true, // files with matches only
	"-c":            true, // count only
	"-w":            true, // word match
	"-E":            true, // extended regex
	"-P":            true, // perl regex
	"-F":            true, // fixed strings
	"-v":            true, // invert match
	"-H":            true, // print filename
	"-h":            true, // no filename
	"-o":            true, // only matching
	"-A":            true, // after context
	"-B":            true, // before context
	"-C":            true, // context
	"--include":     true, // file pattern include
	"--exclude":     true, // file pattern exclude
	"--exclude-dir": true, // directory exclude
}

// dangerousChars are characters that indicate shell metacharacters outside quoted strings
// These are checked on the raw command before parsing to catch obvious injection attempts
var dangerousChars = []*regexp.Regexp{
	regexp.MustCompile("`"),              // command substitution (backticks)
	regexp.MustCompile(`\$\(`),           // command substitution $()
	regexp.MustCompile(`\$\{`),           // variable expansion ${}
	regexp.MustCompile(`\n`),             // newlines
	regexp.MustCompile(`\\x[0-9a-fA-F]`), // hex escapes
}

// shellMetachars are checked per-token to detect command chaining/redirection
// These are safe inside quoted grep patterns but dangerous as shell operators
var shellMetachars = regexp.MustCompile(`^[;&|<>]$|;$|&&|\|\|`)

// SandboxGrepMatch represents a single grep match
type SandboxGrepMatch struct {
	File       string `json:"file"`
	LineNumber int    `json:"lineNumber"`
	Content    string `json:"content"`
}

// SandboxGrepResult represents the result of a sandbox grep operation
type SandboxGrepResult struct {
	Matches    []SandboxGrepMatch `json:"matches"`
	TotalCount int                `json:"totalCount"`
	Truncated  bool               `json:"truncated"`
}

// GrepSandboxTool executes grep commands in the sandbox environment
type GrepSandboxTool struct {
	sandbox *sandbox.Service
}

// NewGrepSandboxTool creates a new sandbox grep tool
func NewGrepSandboxTool(sandbox *sandbox.Service) *GrepSandboxTool {
	return &GrepSandboxTool{sandbox: sandbox}
}

func (t *GrepSandboxTool) Name() string {
	return "sandbox_grep"
}

func (t *GrepSandboxTool) Description() string {
	return "Search for patterns in files within the repository using grep. Executes grep commands in the sandbox environment. Supports common flags like -r (recursive), -n (line numbers), -i (case insensitive), -w (word match), -E (extended regex). Command must start with 'grep'."
}

func (t *GrepSandboxTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"command": {
				Type:        "string",
				Description: "The grep command to execute (e.g., 'grep -rn \"pattern\" .'). Must start with 'grep'.",
			},
		},
		Required: []string{"command"},
	}
}

func (t *GrepSandboxTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract userID from context
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user ID not found in context")
	}

	// Get command from params
	command, ok := params["command"].(string)
	if !ok || command == "" {
		return nil, fmt.Errorf("command parameter is required")
	}

	// Validate and sanitize command
	args, err := sanitizeGrepCommand(command)
	if err != nil {
		return nil, fmt.Errorf("invalid grep command: %w", err)
	}

	// Get user's work directory
	workDir, err := t.sandbox.GetOrCreateWorkDir(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// Execute grep command with timeout
	execCtx, cancel := context.WithTimeout(ctx, grepTimeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "grep", args...)
	cmd.Dir = workDir

	output, err := cmd.Output()

	// Handle grep exit codes:
	// 0 = matches found
	// 1 = no matches found (not an error)
	// 2+ = actual errors
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			if exitCode == 1 {
				// No matches found - return empty result
				return SandboxGrepResult{
					Matches:    []SandboxGrepMatch{},
					TotalCount: 0,
					Truncated:  false,
				}, nil
			}
			// Other error - include stderr in error message
			stderr := string(exitErr.Stderr)
			if stderr != "" {
				return nil, fmt.Errorf("grep failed: %s", strings.TrimSpace(stderr))
			}
			return nil, fmt.Errorf("grep failed with exit code %d", exitCode)
		}
		if execCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("grep command timed out after %v", grepTimeout)
		}
		return nil, fmt.Errorf("grep execution failed: %w", err)
	}

	// Parse output into structured format
	result := parseGrepOutput(string(output))

	return result, nil
}

func (t *GrepSandboxTool) RequiresConfirmation() bool {
	return false // grep is read-only and safe
}

// sanitizeGrepCommand validates and parses a grep command string
// Returns the arguments to pass to grep, or an error if the command is invalid
func sanitizeGrepCommand(command string) ([]string, error) {
	// Trim whitespace
	command = strings.TrimSpace(command)

	if command == "" {
		return nil, fmt.Errorf("empty command")
	}

	// Check for dangerous characters that are never valid in grep commands
	for _, pattern := range dangerousChars {
		if pattern.MatchString(command) {
			return nil, fmt.Errorf("command contains potentially dangerous characters")
		}
	}

	// Parse command into tokens (respecting quotes)
	tokens, err := parseCommandTokens(command)
	if err != nil {
		return nil, fmt.Errorf("failed to parse command: %w", err)
	}

	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	// First token must be grep
	if tokens[0] != "grep" {
		return nil, fmt.Errorf("command must start with 'grep'")
	}

	// Extract arguments (skip the "grep" part)
	args := tokens[1:]

	// Check each token for shell metacharacters that indicate command chaining
	for _, token := range tokens {
		if shellMetachars.MatchString(token) {
			return nil, fmt.Errorf("command contains potentially dangerous characters")
		}
		// Also check for redirection operators
		if strings.Contains(token, ">") || strings.HasPrefix(token, "<") {
			return nil, fmt.Errorf("command contains potentially dangerous characters")
		}
	}

	// Validate flags
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Skip if it's a pattern or path (doesn't start with -)
		if !strings.HasPrefix(arg, "-") {
			continue
		}

		// Handle combined short flags like -rni
		if len(arg) > 2 && !strings.HasPrefix(arg, "--") {
			for _, flag := range arg[1:] {
				flagStr := "-" + string(flag)
				if !allowedGrepFlags[flagStr] {
					return nil, fmt.Errorf("flag '%s' is not allowed", flagStr)
				}
			}
			continue
		}

		// Handle long flags and flags with values
		flagName := arg
		if strings.Contains(arg, "=") {
			flagName = strings.Split(arg, "=")[0]
		}

		if !allowedGrepFlags[flagName] {
			return nil, fmt.Errorf("flag '%s' is not allowed", flagName)
		}

		// For flags that take arguments without =, check if next token is the value
		if flagName == "-A" || flagName == "-B" || flagName == "-C" {
			if !strings.Contains(arg, "=") && i+1 < len(args) {
				// Validate numeric argument
				if _, err := strconv.Atoi(args[i+1]); err != nil {
					return nil, fmt.Errorf("flag %s requires a numeric argument", flagName)
				}
			}
		}
	}

	// Add -n flag if not present to always get line numbers
	hasLineNumbers := false
	for _, arg := range args {
		if arg == "-n" || strings.Contains(arg, "n") && strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") {
			hasLineNumbers = true
			break
		}
	}

	if !hasLineNumbers {
		args = append([]string{"-n"}, args...)
	}

	return args, nil
}

// parseCommandTokens splits a command string into tokens, respecting quotes
func parseCommandTokens(command string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	escaped := false

	for _, r := range command {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}

		if r == '\\' && !inSingleQuote {
			escaped = true
			continue
		}

		if r == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			continue
		}

		if r == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			continue
		}

		if r == ' ' && !inSingleQuote && !inDoubleQuote {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	if inSingleQuote || inDoubleQuote {
		return nil, fmt.Errorf("unclosed quote in command")
	}

	return tokens, nil
}

// parseGrepOutput parses grep output into structured matches
func parseGrepOutput(output string) SandboxGrepResult {
	result := SandboxGrepResult{
		Matches:    []SandboxGrepMatch{},
		TotalCount: 0,
		Truncated:  false,
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	totalCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		totalCount++

		// Only keep first maxGrepMatches matches
		if len(result.Matches) >= maxGrepMatches {
			result.Truncated = true
			continue
		}

		match := parseGrepLine(line)
		if match != nil {
			result.Matches = append(result.Matches, *match)
		}
	}

	result.TotalCount = totalCount
	return result
}

// parseGrepLine parses a single grep output line (format: file:line:content)
func parseGrepLine(line string) *SandboxGrepMatch {
	// Standard grep -n output format: filename:linenum:content
	// Handle case where filename might contain colons

	firstColon := strings.Index(line, ":")
	if firstColon == -1 {
		return nil
	}

	remaining := line[firstColon+1:]
	secondColon := strings.Index(remaining, ":")
	if secondColon == -1 {
		return nil
	}

	file := line[:firstColon]
	lineNumStr := remaining[:secondColon]
	content := remaining[secondColon+1:]

	lineNum, err := strconv.Atoi(lineNumStr)
	if err != nil {
		return nil
	}

	return &SandboxGrepMatch{
		File:       file,
		LineNumber: lineNum,
		Content:    content,
	}
}
