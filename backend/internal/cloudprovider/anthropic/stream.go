package anthropic

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/jacklau/prism/internal/cloudprovider"
)

// streamReader handles SSE streaming from the Anthropic API
type streamReader struct {
	reader  io.ReadCloser
	scanner *bufio.Scanner
	ctx     context.Context
}

// newStreamReader creates a new stream reader
func newStreamReader(ctx context.Context, reader io.ReadCloser) *streamReader {
	return &streamReader{
		reader:  reader,
		scanner: bufio.NewScanner(reader),
		ctx:     ctx,
	}
}

// close closes the underlying reader
func (s *streamReader) close() error {
	return s.reader.Close()
}

// processStream reads SSE events and sends chunks to the channel
func (s *streamReader) processStream(chunks chan<- cloudprovider.MessageChunk) {
	defer close(chunks)
	defer s.close()

	var currentToolCall *cloudprovider.ToolCall
	var toolInputJSON strings.Builder
	var messageID string

	for s.scanner.Scan() {
		// Check for context cancellation
		select {
		case <-s.ctx.Done():
			chunks <- cloudprovider.MessageChunk{
				Error: s.ctx.Err(),
			}
			return
		default:
		}

		line := s.scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip event type lines (we handle based on data content)
		if strings.HasPrefix(line, "event: ") {
			continue
		}

		// Handle data lines
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		var event streamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		switch event.Type {
		case "message_start":
			if event.Message != nil {
				messageID = event.Message.ID
			}

		case "content_block_start":
			if event.ContentBlock != nil && event.ContentBlock.Type == "tool_use" {
				currentToolCall = &cloudprovider.ToolCall{
					ID:         event.ContentBlock.ID,
					Name:       event.ContentBlock.Name,
					Parameters: make(map[string]interface{}),
				}
				toolInputJSON.Reset()
			}

		case "content_block_delta":
			if event.Delta != nil {
				switch event.Delta.Type {
				case "text_delta":
					chunks <- cloudprovider.MessageChunk{
						Delta:     event.Delta.Text,
						MessageID: messageID,
					}
				case "input_json_delta":
					toolInputJSON.WriteString(event.Delta.PartialJSON)
				}
			}

		case "content_block_stop":
			if currentToolCall != nil {
				// Parse complete tool input
				if toolInputJSON.Len() > 0 {
					json.Unmarshal([]byte(toolInputJSON.String()), &currentToolCall.Parameters)
				}
				chunks <- cloudprovider.MessageChunk{
					ToolCalls: []cloudprovider.ToolCall{*currentToolCall},
					MessageID: messageID,
				}
				currentToolCall = nil
			}

		case "message_delta":
			if event.Delta != nil && event.Delta.StopReason != "" {
				chunks <- cloudprovider.MessageChunk{
					FinishReason: event.Delta.StopReason,
					MessageID:    messageID,
				}
			}

		case "message_stop":
			chunks <- cloudprovider.MessageChunk{
				FinishReason: "stop",
				MessageID:    messageID,
			}

		case "error":
			chunks <- cloudprovider.MessageChunk{
				Error: cloudprovider.NewAPIError(500, "stream_error", "streaming error occurred"),
			}
			return
		}
	}

	if err := s.scanner.Err(); err != nil {
		chunks <- cloudprovider.MessageChunk{
			Error: err,
		}
	}
}
