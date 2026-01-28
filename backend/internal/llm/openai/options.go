package openai

import (
	"fmt"
)

// ResponseFormat specifies the format of the model's output
type ResponseFormat struct {
	Type       string      `json:"type"` // "text", "json_object", "json_schema"
	JSONSchema *JSONSchema `json:"json_schema,omitempty"`
}

// JSONSchema represents a JSON Schema for structured outputs
type JSONSchema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Schema      map[string]interface{} `json:"schema"`
	Strict      bool                   `json:"strict,omitempty"`
}

// ChatOptions contains optional parameters for chat requests
type ChatOptions struct {
	Temperature       *float64        `json:"temperature,omitempty"`
	TopP              *float64        `json:"top_p,omitempty"`
	MaxTokens         *int            `json:"max_tokens,omitempty"`
	PresencePenalty   *float64        `json:"presence_penalty,omitempty"`
	FrequencyPenalty  *float64        `json:"frequency_penalty,omitempty"`
	Stop              []string        `json:"stop,omitempty"`
	ResponseFormat    *ResponseFormat `json:"response_format,omitempty"`
	Seed              *int            `json:"seed,omitempty"`
	ParallelToolCalls *bool           `json:"parallel_tool_calls,omitempty"`
	User              string          `json:"user,omitempty"`
}

// DefaultOptions returns default chat options
func DefaultOptions() *ChatOptions {
	return &ChatOptions{}
}

// WithTemperature sets the temperature
func (o *ChatOptions) WithTemperature(temp float64) *ChatOptions {
	o.Temperature = &temp
	return o
}

// WithTopP sets the top_p parameter
func (o *ChatOptions) WithTopP(topP float64) *ChatOptions {
	o.TopP = &topP
	return o
}

// WithMaxTokens sets the max tokens
func (o *ChatOptions) WithMaxTokens(maxTokens int) *ChatOptions {
	o.MaxTokens = &maxTokens
	return o
}

// WithJSONMode enables JSON mode
func (o *ChatOptions) WithJSONMode() *ChatOptions {
	o.ResponseFormat = &ResponseFormat{Type: "json_object"}
	return o
}

// WithJSONSchema enables structured output with a specific schema
func (o *ChatOptions) WithJSONSchema(schema *JSONSchema) *ChatOptions {
	o.ResponseFormat = &ResponseFormat{
		Type:       "json_schema",
		JSONSchema: schema,
	}
	return o
}

// WithParallelToolCalls enables or disables parallel tool calls
func (o *ChatOptions) WithParallelToolCalls(enabled bool) *ChatOptions {
	o.ParallelToolCalls = &enabled
	return o
}

// WithSeed sets the seed for deterministic outputs
func (o *ChatOptions) WithSeed(seed int) *ChatOptions {
	o.Seed = &seed
	return o
}

// WithStop sets the stop sequences
func (o *ChatOptions) WithStop(stop ...string) *ChatOptions {
	o.Stop = stop
	return o
}

// WithUser sets the user identifier
func (o *ChatOptions) WithUser(user string) *ChatOptions {
	o.User = user
	return o
}

// Validate validates the chat options
func (o *ChatOptions) Validate() error {
	if o.Temperature != nil {
		if *o.Temperature < 0 || *o.Temperature > 2 {
			return fmt.Errorf("temperature must be between 0 and 2, got %f", *o.Temperature)
		}
	}
	if o.TopP != nil {
		if *o.TopP < 0 || *o.TopP > 1 {
			return fmt.Errorf("top_p must be between 0 and 1, got %f", *o.TopP)
		}
	}
	if o.MaxTokens != nil && *o.MaxTokens < 1 {
		return fmt.Errorf("max_tokens must be positive, got %d", *o.MaxTokens)
	}
	if o.PresencePenalty != nil {
		if *o.PresencePenalty < -2 || *o.PresencePenalty > 2 {
			return fmt.Errorf("presence_penalty must be between -2 and 2, got %f", *o.PresencePenalty)
		}
	}
	if o.FrequencyPenalty != nil {
		if *o.FrequencyPenalty < -2 || *o.FrequencyPenalty > 2 {
			return fmt.Errorf("frequency_penalty must be between -2 and 2, got %f", *o.FrequencyPenalty)
		}
	}
	if len(o.Stop) > 4 {
		return fmt.Errorf("maximum 4 stop sequences allowed, got %d", len(o.Stop))
	}
	return nil
}

// ApplyModelDefaults applies model-specific defaults to the options
func (o *ChatOptions) ApplyModelDefaults(modelID string) *ChatOptions {
	config := GetModelConfig(modelID)

	// Set max tokens if not specified, respecting model limits
	if o.MaxTokens == nil {
		maxOutput := config.MaxOutput
		o.MaxTokens = &maxOutput
	} else if *o.MaxTokens > config.MaxOutput {
		o.MaxTokens = &config.MaxOutput
	}

	return o
}

// ToRequestBody converts options to request body fields
func (o *ChatOptions) ToRequestBody() map[string]interface{} {
	body := make(map[string]interface{})

	if o.Temperature != nil {
		body["temperature"] = *o.Temperature
	}
	if o.TopP != nil {
		body["top_p"] = *o.TopP
	}
	if o.MaxTokens != nil {
		body["max_tokens"] = *o.MaxTokens
	}
	if o.PresencePenalty != nil {
		body["presence_penalty"] = *o.PresencePenalty
	}
	if o.FrequencyPenalty != nil {
		body["frequency_penalty"] = *o.FrequencyPenalty
	}
	if len(o.Stop) > 0 {
		body["stop"] = o.Stop
	}
	if o.ResponseFormat != nil {
		body["response_format"] = o.ResponseFormat
	}
	if o.Seed != nil {
		body["seed"] = *o.Seed
	}
	if o.ParallelToolCalls != nil {
		body["parallel_tool_calls"] = *o.ParallelToolCalls
	}
	if o.User != "" {
		body["user"] = o.User
	}

	return body
}
