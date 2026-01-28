package openai

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/jacklau/prism/internal/llm"
	"github.com/pkoukk/tiktoken-go"
)

// TokenCounter counts tokens for OpenAI models
type TokenCounter struct {
	encoding *tiktoken.Tiktoken
	model    string
	mu       sync.RWMutex
}

// encodingCache caches tiktoken encodings by model
var (
	encodingCache   = make(map[string]*tiktoken.Tiktoken)
	encodingCacheMu sync.RWMutex
)

// NewTokenCounter creates a new token counter for the specified model
func NewTokenCounter(model string) (*TokenCounter, error) {
	encoding, err := getEncodingForModel(model)
	if err != nil {
		return nil, err
	}

	return &TokenCounter{
		encoding: encoding,
		model:    model,
	}, nil
}

// getEncodingForModel returns the tiktoken encoding for a model
func getEncodingForModel(model string) (*tiktoken.Tiktoken, error) {
	encodingCacheMu.RLock()
	if enc, ok := encodingCache[model]; ok {
		encodingCacheMu.RUnlock()
		return enc, nil
	}
	encodingCacheMu.RUnlock()

	encodingCacheMu.Lock()
	defer encodingCacheMu.Unlock()

	// Double-check after acquiring write lock
	if enc, ok := encodingCache[model]; ok {
		return enc, nil
	}

	// Get encoding for model
	enc, err := tiktoken.EncodingForModel(model)
	if err != nil {
		// Fallback to cl100k_base for newer models
		enc, err = tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			return nil, fmt.Errorf("failed to get encoding: %w", err)
		}
	}

	encodingCache[model] = enc
	return enc, nil
}

// CountTokens counts tokens in a text string
func (tc *TokenCounter) CountTokens(text string) int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	return len(tc.encoding.Encode(text, nil, nil))
}

// CountMessage counts tokens in a single message
func (tc *TokenCounter) CountMessage(msg llm.Message) int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	// Base tokens per message (role, content separators)
	tokens := 4 // <|start|>role<|sep|>content<|end|>

	// Count role
	tokens += len(tc.encoding.Encode(msg.Role, nil, nil))

	// Count content
	tokens += len(tc.encoding.Encode(msg.Content, nil, nil))

	// Count tool calls
	for range msg.ToolCalls {
		tokens += tc.countToolCallTokens()
	}

	// Tool call ID
	if msg.ToolCallID != "" {
		tokens += len(tc.encoding.Encode(msg.ToolCallID, nil, nil))
	}

	return tokens
}

// countToolCallTokens counts tokens in a tool call
func (tc *TokenCounter) countToolCallTokens() int {
	// This is a rough estimate; actual tokenization varies
	return 10 // Base overhead for tool call structure
}

// CountMessages counts tokens in a slice of messages
func (tc *TokenCounter) CountMessages(messages []llm.Message) int {
	total := 3 // Base tokens for the conversation structure

	for _, msg := range messages {
		total += tc.CountMessage(msg)
	}

	return total
}

// CountToolDefinition counts tokens in a tool definition
func (tc *TokenCounter) CountToolDefinition(tool llm.ToolDefinition) int {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	tokens := 0

	// Count name and description
	tokens += len(tc.encoding.Encode(tool.Name, nil, nil))
	tokens += len(tc.encoding.Encode(tool.Description, nil, nil))

	// Count parameters schema (serialize to JSON and count)
	paramsJSON, _ := json.Marshal(tool.Parameters)
	tokens += len(tc.encoding.Encode(string(paramsJSON), nil, nil))

	// Add overhead for structure
	tokens += 10

	return tokens
}

// CountTools counts tokens for all tool definitions
func (tc *TokenCounter) CountTools(tools []llm.ToolDefinition) int {
	if len(tools) == 0 {
		return 0
	}

	total := 3 // Base tokens for tools array

	for _, tool := range tools {
		total += tc.CountToolDefinition(tool)
	}

	return total
}

// CountRequest counts total tokens for a chat request
func (tc *TokenCounter) CountRequest(req *llm.ChatRequest) int {
	total := tc.CountMessages(req.Messages)
	total += tc.CountTools(req.Tools)
	return total
}

// EstimateTokens provides a quick estimate without loading the tokenizer
func EstimateTokens(text string) int {
	// Rough estimate: ~4 characters per token for English text
	return len(text) / 4
}

// EstimateMessagesTokens estimates tokens for messages without a tokenizer
func EstimateMessagesTokens(messages []llm.Message) int {
	total := 3

	for _, msg := range messages {
		total += 4 // Message overhead
		total += EstimateTokens(msg.Content)
		total += EstimateTokens(msg.Role)
	}

	return total
}

// GetContextUsage returns the percentage of context window used
func (tc *TokenCounter) GetContextUsage(tokens int) float64 {
	config := GetModelConfig(tc.model)
	return float64(tokens) / float64(config.ContextWindow) * 100
}

// RemainingTokens returns the remaining tokens available for output
func (tc *TokenCounter) RemainingTokens(inputTokens int) int {
	return EstimateMaxOutputTokens(tc.model, inputTokens)
}
