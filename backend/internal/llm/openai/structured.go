package openai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jacklau/prism/internal/llm"
)

// StructuredOutputRequest contains a chat request with structured output schema
type StructuredOutputRequest struct {
	*llm.ChatRequest
	Schema *JSONSchema
}

// ChatWithSchema sends a chat request with structured output schema
// and returns a streaming response that will be valid JSON matching the schema
func (c *Client) ChatWithSchema(ctx context.Context, req *llm.ChatRequest, schema *JSONSchema) (<-chan llm.StreamChunk, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema is required for structured output")
	}

	// Create options with JSON schema
	opts := DefaultOptions().WithJSONSchema(schema)

	return c.ChatWithOptions(ctx, req, opts)
}

// ParseStructuredOutput parses a structured output response into the target type
func ParseStructuredOutput[T any](content string) (*T, error) {
	var result T
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse structured output: %w", err)
	}
	return &result, nil
}

// NewJSONSchema creates a new JSON schema for structured outputs
func NewJSONSchema(name string, schema map[string]interface{}) *JSONSchema {
	return &JSONSchema{
		Name:   name,
		Schema: schema,
		Strict: true,
	}
}

// NewJSONSchemaWithDescription creates a new JSON schema with a description
func NewJSONSchemaWithDescription(name, description string, schema map[string]interface{}) *JSONSchema {
	return &JSONSchema{
		Name:        name,
		Description: description,
		Schema:      schema,
		Strict:      true,
	}
}

// BuildObjectSchema is a helper to build an object schema
func BuildObjectSchema(properties map[string]interface{}, required []string) map[string]interface{} {
	return map[string]interface{}{
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	}
}

// StringProperty creates a string property definition
func StringProperty(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"description": description,
	}
}

// NumberProperty creates a number property definition
func NumberProperty(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "number",
		"description": description,
	}
}

// IntegerProperty creates an integer property definition
func IntegerProperty(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "integer",
		"description": description,
	}
}

// BooleanProperty creates a boolean property definition
func BooleanProperty(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "boolean",
		"description": description,
	}
}

// ArrayProperty creates an array property definition
func ArrayProperty(description string, items map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type":        "array",
		"description": description,
		"items":       items,
	}
}

// EnumProperty creates a string enum property definition
func EnumProperty(description string, values []string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"description": description,
		"enum":        values,
	}
}
