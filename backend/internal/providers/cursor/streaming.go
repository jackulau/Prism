package cursor

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/jacklau/prism/internal/providers"
)

// StreamParser handles parsing SSE events from Cursor API
type StreamParser struct {
	reader *bufio.Reader
}

// NewStreamParser creates a new SSE stream parser
func NewStreamParser(r io.Reader) *StreamParser {
	return &StreamParser{
		reader: bufio.NewReader(r),
	}
}

// ParseStream reads SSE events and sends StreamChunks to the channel
func (p *StreamParser) ParseStream(chunks chan<- providers.StreamChunk) {
	defer close(chunks)

	var eventType string
	var dataBuilder strings.Builder

	for {
		line, err := p.reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				chunks <- providers.StreamChunk{
					Error: fmt.Errorf("stream read error: %w", err),
				}
			}
			return
		}

		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")

		// Empty line signals end of event
		if line == "" {
			if dataBuilder.Len() > 0 {
				chunk := p.parseEvent(eventType, dataBuilder.String())
				if chunk != nil {
					chunks <- *chunk
				}
				dataBuilder.Reset()
				eventType = ""
			}
			continue
		}

		// Parse SSE fields
		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimPrefix(data, " ") // Optional space after colon
			dataBuilder.WriteString(data)
		}
		// Ignore comments (lines starting with :) and other fields
	}
}

// parseEvent converts an SSE event to a StreamChunk
func (p *StreamParser) parseEvent(eventType, data string) *providers.StreamChunk {
	// Handle different event types
	switch eventType {
	case "content", "delta", "text":
		return p.parseContentEvent(data)
	case "tool_call", "tool":
		return p.parseToolCallEvent(data)
	case "error":
		return p.parseErrorEvent(data)
	case "done", "finish", "message_stop":
		return p.parseFinishEvent(data)
	default:
		// Try to parse as JSON to determine event type
		return p.parseJSONEvent(data)
	}
}

// parseContentEvent handles text content events
func (p *StreamParser) parseContentEvent(data string) *providers.StreamChunk {
	// Try parsing as JSON first
	var event CursorStreamEvent
	if err := json.Unmarshal([]byte(data), &event); err == nil {
		if event.Delta != "" {
			return &providers.StreamChunk{Delta: event.Delta}
		}
	}

	// If not JSON, use raw data as delta
	if data != "" {
		return &providers.StreamChunk{Delta: data}
	}

	return nil
}

// parseToolCallEvent handles tool call events
func (p *StreamParser) parseToolCallEvent(data string) *providers.StreamChunk {
	var event CursorStreamEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return nil
	}

	if event.ToolCall != nil {
		return &providers.StreamChunk{
			ToolCalls: []providers.ToolCall{
				{
					ID:         event.ToolCall.ID,
					Name:       event.ToolCall.Name,
					Parameters: event.ToolCall.Parameters,
				},
			},
		}
	}

	return nil
}

// parseErrorEvent handles error events
func (p *StreamParser) parseErrorEvent(data string) *providers.StreamChunk {
	var event CursorStreamEvent
	if err := json.Unmarshal([]byte(data), &event); err == nil && event.Error != nil {
		return &providers.StreamChunk{
			Error: &providers.ProviderError{
				Provider: "cursor",
				Code:     event.Error.Code,
				Message:  event.Error.Message,
			},
		}
	}

	// Fallback: use data as error message
	return &providers.StreamChunk{
		Error: fmt.Errorf("stream error: %s", data),
	}
}

// parseFinishEvent handles finish/done events
func (p *StreamParser) parseFinishEvent(data string) *providers.StreamChunk {
	chunk := &providers.StreamChunk{
		FinishReason: "stop",
	}

	// Try to parse JSON for additional info
	var event CursorStreamEvent
	if err := json.Unmarshal([]byte(data), &event); err == nil {
		if event.FinishReason != "" {
			chunk.FinishReason = event.FinishReason
		}
		if event.MessageID != "" {
			chunk.MessageID = event.MessageID
		}
	}

	return chunk
}

// parseJSONEvent tries to parse a generic JSON event
func (p *StreamParser) parseJSONEvent(data string) *providers.StreamChunk {
	if data == "" || data == "[DONE]" {
		return &providers.StreamChunk{FinishReason: "stop"}
	}

	var event CursorStreamEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		// Not JSON, might be raw text
		return nil
	}

	// Check for different fields to determine chunk type
	if event.Delta != "" {
		return &providers.StreamChunk{Delta: event.Delta}
	}

	if event.ToolCall != nil {
		return &providers.StreamChunk{
			ToolCalls: []providers.ToolCall{
				{
					ID:         event.ToolCall.ID,
					Name:       event.ToolCall.Name,
					Parameters: event.ToolCall.Parameters,
				},
			},
		}
	}

	if event.Error != nil {
		return &providers.StreamChunk{
			Error: &providers.ProviderError{
				Provider: "cursor",
				Code:     event.Error.Code,
				Message:  event.Error.Message,
			},
		}
	}

	if event.FinishReason != "" {
		return &providers.StreamChunk{
			FinishReason: event.FinishReason,
			MessageID:    event.MessageID,
		}
	}

	return nil
}
