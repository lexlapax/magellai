// ABOUTME: Partial response handling for streaming and incomplete LLM outputs
// ABOUTME: Provides mechanisms to handle and recover from partial responses

package llm

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/lexlapax/magellai/internal/logging"
)

// PartialResponseHandler manages partial responses and attempts to complete them
type PartialResponseHandler struct {
	provider   Provider
	buffer     *ResponseBuffer
	logger     *logging.Logger
	timeout    time.Duration
	maxRetries int
}

// ResponseBuffer accumulates streaming responses
type ResponseBuffer struct {
	mu           sync.Mutex
	content      strings.Builder
	chunks       []StreamChunk
	lastChunk    time.Time
	complete     bool
	finishReason string
}

// NewPartialResponseHandler creates a handler for partial responses
func NewPartialResponseHandler(provider Provider) *PartialResponseHandler {
	return &PartialResponseHandler{
		provider:   provider,
		buffer:     &ResponseBuffer{},
		logger:     logging.GetLogger(),
		timeout:    30 * time.Second,
		maxRetries: 3,
	}
}

// HandleStreamWithRecovery wraps streaming with recovery for partial responses
func (h *PartialResponseHandler) HandleStreamWithRecovery(
	ctx context.Context,
	prompt string,
	options ...ProviderOption,
) (<-chan StreamChunk, error) {
	// Get the original stream
	stream, err := h.provider.Stream(ctx, prompt, options...)
	if err != nil {
		return nil, err
	}

	// Create wrapped stream with recovery
	wrappedStream := make(chan StreamChunk)

	go func() {
		defer close(wrappedStream)
		h.processStreamWithRecovery(ctx, stream, wrappedStream, prompt, options...)
	}()

	return wrappedStream, nil
}

// processStreamWithRecovery handles the stream processing with recovery logic
func (h *PartialResponseHandler) processStreamWithRecovery(
	ctx context.Context,
	input <-chan StreamChunk,
	output chan<- StreamChunk,
	prompt string,
	options ...ProviderOption,
) {
	defer h.buffer.Reset()

	// Set up timeout for detecting stalled streams
	timeoutTimer := time.NewTimer(h.timeout)
	defer timeoutTimer.Stop()

	// Process chunks
	for {
		select {
		case chunk, ok := <-input:
			if !ok {
				// Stream closed - check if we got a complete response
				if h.buffer.IsComplete() {
					return
				}

				// Attempt recovery for incomplete response
				h.logger.Warn("Stream closed with incomplete response, attempting recovery")
				h.attemptRecovery(ctx, output, prompt, options...)
				return
			}

			// Handle the chunk
			if chunk.Error != nil {
				h.logger.Warn("Error in stream chunk", "error", chunk.Error)
				// Try to recover if we have partial content
				if h.buffer.HasContent() {
					h.attemptRecovery(ctx, output, prompt, options...)
				} else {
					output <- chunk // Forward error
				}
				return
			}

			// Add chunk to buffer
			h.buffer.AddChunk(chunk)

			// Forward chunk
			output <- chunk

			// Reset timeout
			timeoutTimer.Reset(h.timeout)

			// Check if response is complete
			if chunk.FinishReason != "" {
				h.buffer.SetComplete(chunk.FinishReason)
				return
			}

		case <-timeoutTimer.C:
			// Stream timeout - attempt recovery
			h.logger.Warn("Stream timeout, attempting recovery")
			h.attemptRecovery(ctx, output, prompt, options...)
			return

		case <-ctx.Done():
			// Context cancelled
			output <- StreamChunk{Error: ctx.Err()}
			return
		}
	}
}

// attemptRecovery tries to complete a partial response
func (h *PartialResponseHandler) attemptRecovery(
	ctx context.Context,
	output chan<- StreamChunk,
	originalPrompt string,
	options ...ProviderOption,
) {
	partialContent := h.buffer.GetContent()
	if partialContent == "" {
		output <- StreamChunk{Error: fmt.Errorf("empty partial response")}
		return
	}

	h.logger.Info("Attempting to recover partial response",
		"contentLength", len(partialContent))

	// Create continuation prompt
	continuationPrompt := fmt.Sprintf(
		"Continue the following incomplete response:\n\n%s\n\n[Continue from where you left off]",
		partialContent,
	)

	// Try to get completion
	for attempt := 0; attempt < h.maxRetries; attempt++ {
		response, err := h.provider.Generate(ctx, continuationPrompt, options...)
		if err != nil {
			h.logger.Warn("Recovery attempt failed",
				"attempt", attempt+1,
				"error", err)
			continue
		}

		// Check if response continues sensibly
		if isValidContinuation(partialContent, response) {
			// Send continuation as chunks
			for _, line := range strings.Split(response, "\n") {
				output <- StreamChunk{
					Content: line + "\n",
					Index:   h.buffer.GetChunkCount(),
				}
			}

			// Send completion marker
			output <- StreamChunk{
				FinishReason: "recovered",
				Index:        h.buffer.GetChunkCount() + 1,
			}

			h.logger.Info("Successfully recovered partial response")
			return
		}
	}

	// Recovery failed - send error
	output <- StreamChunk{
		Error: fmt.Errorf("failed to recover partial response after %d attempts", h.maxRetries),
	}
}

// isValidContinuation checks if a response is a valid continuation
func isValidContinuation(partial, continuation string) bool {
	// Basic heuristics for valid continuation
	// 1. Should not repeat the entire partial content
	if strings.Contains(continuation, partial) {
		return false
	}

	// 2. Should not start with common error patterns
	errorPrefixes := []string{
		"I cannot continue",
		"I don't have enough context",
		"Error:",
		"Sorry,",
	}

	lowerCont := strings.ToLower(continuation)
	for _, prefix := range errorPrefixes {
		if strings.HasPrefix(lowerCont, strings.ToLower(prefix)) {
			return false
		}
	}

	// 3. Should have reasonable length
	if len(continuation) < 10 {
		return false
	}

	return true
}

// ResponseBuffer methods

// AddChunk adds a chunk to the buffer
func (b *ResponseBuffer) AddChunk(chunk StreamChunk) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.content.WriteString(chunk.Content)
	b.chunks = append(b.chunks, chunk)
	b.lastChunk = time.Now()
}

// GetContent returns the accumulated content
func (b *ResponseBuffer) GetContent() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.content.String()
}

// GetChunkCount returns the number of chunks received
func (b *ResponseBuffer) GetChunkCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return len(b.chunks)
}

// HasContent checks if buffer has any content
func (b *ResponseBuffer) HasContent() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.content.Len() > 0
}

// IsComplete checks if response is complete
func (b *ResponseBuffer) IsComplete() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.complete
}

// SetComplete marks the response as complete
func (b *ResponseBuffer) SetComplete(reason string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.complete = true
	b.finishReason = reason
}

// GetTimeSinceLastChunk returns time since last chunk
func (b *ResponseBuffer) GetTimeSinceLastChunk() time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.lastChunk.IsZero() {
		return 0
	}
	return time.Since(b.lastChunk)
}

// Reset clears the buffer
func (b *ResponseBuffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.content.Reset()
	b.chunks = nil
	b.lastChunk = time.Time{}
	b.complete = false
	b.finishReason = ""
}

// PartialResponseDetector analyzes responses for completeness
type PartialResponseDetector struct {
	minSentenceLength int
	sentenceEnders    []string
}

// NewPartialResponseDetector creates a detector for partial responses
func NewPartialResponseDetector() *PartialResponseDetector {
	return &PartialResponseDetector{
		minSentenceLength: 20,
		sentenceEnders:    []string{".", "!", "?", "```"},
	}
}

// IsComplete analyzes if a response appears complete
func (d *PartialResponseDetector) IsComplete(content string) bool {
	trimmed := strings.TrimSpace(content)
	if len(trimmed) < d.minSentenceLength {
		return false
	}

	// Check for sentence endings
	for _, ender := range d.sentenceEnders {
		if strings.HasSuffix(trimmed, ender) {
			return true
		}
	}

	// Check for code block completion
	codeBlockCount := strings.Count(content, "```")
	if codeBlockCount > 0 && codeBlockCount%2 == 0 {
		return true
	}

	// Check for list completion (ends with list item)
	lines := strings.Split(trimmed, "\n")
	if len(lines) > 0 {
		lastLine := strings.TrimSpace(lines[len(lines)-1])
		if strings.HasPrefix(lastLine, "- ") || strings.HasPrefix(lastLine, "* ") {
			return true
		}
		// Numbered list
		if len(lastLine) > 2 && lastLine[1] == '.' && lastLine[0] >= '0' && lastLine[0] <= '9' {
			return true
		}
	}

	return false
}

// ExtractLastCompleteUnit extracts the last complete unit (sentence, paragraph, etc.)
func (d *PartialResponseDetector) ExtractLastCompleteUnit(content string) (complete, incomplete string) {
	// Find the last sentence boundary
	lastBoundary := -1

	for _, ender := range d.sentenceEnders {
		idx := strings.LastIndex(content, ender)
		if idx > lastBoundary {
			lastBoundary = idx
		}
	}

	if lastBoundary == -1 {
		return "", content
	}

	// Include the sentence ender in the complete part
	complete = content[:lastBoundary+1]
	incomplete = content[lastBoundary+1:]

	return complete, incomplete
}
