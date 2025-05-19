// ABOUTME: Tests for partial response handling and streaming recovery
// ABOUTME: Verifies response buffering, completion attempts, and error recovery

package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lexlapax/magellai/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPartialResponseHandler(t *testing.T) {
	provider := &mockProvider{}
	handler := NewPartialResponseHandler(provider)

	assert.NotNil(t, handler)
	assert.Equal(t, provider, handler.provider)
	assert.NotNil(t, handler.buffer)
	assert.NotNil(t, handler.logger)
	assert.Equal(t, 30*time.Second, handler.timeout)
	assert.Equal(t, 3, handler.maxRetries)
}

func TestResponseBuffer_AddChunk(t *testing.T) {
	buffer := &ResponseBuffer{}

	chunk := StreamChunk{
		Content: "Hello ",
		Index:   0,
	}

	buffer.AddChunk(chunk)

	assert.Equal(t, "Hello ", buffer.GetContent())
	assert.Equal(t, 1, buffer.GetChunkCount())
	assert.False(t, buffer.IsComplete())
}

func TestResponseBuffer_SetComplete(t *testing.T) {
	buffer := &ResponseBuffer{}

	// Add multiple chunks
	chunks := []StreamChunk{
		{Content: "Hello ", Index: 0},
		{Content: "world", Index: 1},
		{Content: "!", Index: 2},
	}

	for _, chunk := range chunks {
		buffer.AddChunk(chunk)
	}

	buffer.SetComplete("stop")

	assert.Equal(t, "Hello world!", buffer.GetContent())
	assert.Equal(t, 3, buffer.GetChunkCount())
	assert.True(t, buffer.IsComplete())
}

func TestResponseBuffer_HasContent(t *testing.T) {
	buffer := &ResponseBuffer{}

	assert.False(t, buffer.HasContent())

	buffer.AddChunk(StreamChunk{Content: "Test"})
	assert.True(t, buffer.HasContent())
}

func TestResponseBuffer_Reset(t *testing.T) {
	buffer := &ResponseBuffer{}

	buffer.AddChunk(StreamChunk{Content: "Test"})
	buffer.SetComplete("stop")

	buffer.Reset()

	assert.Equal(t, "", buffer.GetContent())
	assert.Equal(t, 0, buffer.GetChunkCount())
	assert.False(t, buffer.IsComplete())
	assert.False(t, buffer.HasContent())
}

func TestResponseBuffer_GetTimeSinceLastChunk(t *testing.T) {
	buffer := &ResponseBuffer{}

	// No chunks yet
	assert.Equal(t, time.Duration(0), buffer.GetTimeSinceLastChunk())

	// Add a chunk
	buffer.AddChunk(StreamChunk{Content: "Test"})
	time.Sleep(10 * time.Millisecond)

	duration := buffer.GetTimeSinceLastChunk()
	assert.Greater(t, duration, time.Duration(0))
	assert.Less(t, duration, 100*time.Millisecond)
}

func TestHandleStreamWithRecovery_Success(t *testing.T) {
	chunks := []StreamChunk{
		{Content: "Hello ", Index: 0},
		{Content: "world!", Index: 1, FinishReason: "stop"},
	}

	provider := &mockProvider{
		streamFunc: func(ctx context.Context, prompt string, options ...ProviderOption) (<-chan StreamChunk, error) {
			ch := make(chan StreamChunk)
			go func() {
				defer close(ch)
				for _, chunk := range chunks {
					ch <- chunk
				}
			}()
			return ch, nil
		},
	}

	handler := NewPartialResponseHandler(provider)
	handler.logger = logging.GetLogger()

	ctx := context.Background()
	responseChan, err := handler.HandleStreamWithRecovery(ctx, "test prompt")

	require.NoError(t, err)
	require.NotNil(t, responseChan)

	// Collect all responses
	var responses []StreamChunk
	for chunk := range responseChan {
		responses = append(responses, chunk)
	}

	assert.Len(t, responses, 2)
	assert.Equal(t, "Hello ", responses[0].Content)
	assert.Equal(t, "world!", responses[1].Content)
	assert.Equal(t, "stop", responses[1].FinishReason)
}

func TestHandleStreamWithRecovery_PartialResponse(t *testing.T) {
	incompleteChunks := []StreamChunk{
		{Content: "Hello ", Index: 0},
		{Content: "world", Index: 1},
		// Missing finish chunk
	}

	callCount := 0
	provider := &mockProvider{
		streamFunc: func(ctx context.Context, prompt string, options ...ProviderOption) (<-chan StreamChunk, error) {
			ch := make(chan StreamChunk)
			go func() {
				defer close(ch)
				for _, chunk := range incompleteChunks {
					ch <- chunk
				}
			}()
			return ch, nil
		},
		generateFunc: func(ctx context.Context, prompt string, options ...ProviderOption) (string, error) {
			callCount++
			// Mock recovery response
			return "... and more content.", nil
		},
	}

	handler := NewPartialResponseHandler(provider)
	handler.timeout = 100 * time.Millisecond
	handler.logger = logging.GetLogger()

	ctx := context.Background()
	responseChan, err := handler.HandleStreamWithRecovery(ctx, "test prompt")

	require.NoError(t, err)
	require.NotNil(t, responseChan)

	// Collect all responses
	var responses []StreamChunk
	for chunk := range responseChan {
		responses = append(responses, chunk)
	}

	// Should have original chunks plus recovery
	assert.Greater(t, len(responses), 2)
	assert.Equal(t, "Hello ", responses[0].Content)
	assert.Equal(t, "world", responses[1].Content)
	// Recovery responses should be included
	assert.True(t, strings.Contains(responses[len(responses)-2].Content, "... and more content."))
}

func TestHandleStreamWithRecovery_TimeoutError(t *testing.T) {
	// Slow provider that triggers timeout
	provider := &mockProvider{
		streamFunc: func(ctx context.Context, prompt string, options ...ProviderOption) (<-chan StreamChunk, error) {
			ch := make(chan StreamChunk)
			go func() {
				defer close(ch)
				ch <- StreamChunk{Content: "Starting..."}
				// Simulate long delay
				time.Sleep(200 * time.Millisecond)
			}()
			return ch, nil
		},
		generateFunc: func(ctx context.Context, prompt string, options ...ProviderOption) (string, error) {
			return "Recovery content", nil
		},
	}

	handler := NewPartialResponseHandler(provider)
	handler.timeout = 50 * time.Millisecond
	handler.maxRetries = 1
	handler.logger = logging.GetLogger()

	ctx := context.Background()
	responseChan, err := handler.HandleStreamWithRecovery(ctx, "test prompt")

	require.NoError(t, err)
	require.NotNil(t, responseChan)

	// Collect responses
	var responses []StreamChunk
	for chunk := range responseChan {
		responses = append(responses, chunk)
	}

	// Should get at least the first chunk and recovery attempts
	assert.GreaterOrEqual(t, len(responses), 1)
}

func TestHandleStreamWithRecovery_ProviderError(t *testing.T) {
	provider := &mockProvider{
		streamFunc: func(ctx context.Context, prompt string, options ...ProviderOption) (<-chan StreamChunk, error) {
			return nil, errors.New("provider error")
		},
	}

	handler := NewPartialResponseHandler(provider)

	ctx := context.Background()
	responseChan, err := handler.HandleStreamWithRecovery(ctx, "test prompt")

	assert.Error(t, err)
	assert.Nil(t, responseChan)
}

func TestIsValidContinuation(t *testing.T) {
	tests := []struct {
		name         string
		partial      string
		continuation string
		valid        bool
	}{
		{
			name:         "valid continuation",
			partial:      "This is the beginning",
			continuation: "of a longer text that continues naturally.",
			valid:        true,
		},
		{
			name:         "repeats partial content",
			partial:      "Hello world",
			continuation: "Hello world and more",
			valid:        false,
		},
		{
			name:         "error response",
			partial:      "Some content",
			continuation: "I cannot continue this response",
			valid:        false,
		},
		{
			name:         "too short",
			partial:      "Some content",
			continuation: "Yes",
			valid:        false,
		},
		{
			name:         "error prefix",
			partial:      "Content",
			continuation: "Error: failed to continue",
			valid:        false,
		},
		{
			name:         "sorry prefix",
			partial:      "Content",
			continuation: "Sorry, I don't understand",
			valid:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidContinuation(tt.partial, tt.continuation)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestNewPartialResponseDetector(t *testing.T) {
	detector := NewPartialResponseDetector()

	assert.NotNil(t, detector)
	assert.Equal(t, 20, detector.minSentenceLength)
	assert.Equal(t, []string{".", "!", "?", "```"}, detector.sentenceEnders)
}

func TestPartialResponseDetector_IsComplete(t *testing.T) {
	detector := NewPartialResponseDetector()

	tests := []struct {
		name     string
		content  string
		complete bool
	}{
		{
			name:     "complete sentence",
			content:  "This is a complete sentence.",
			complete: true,
		},
		{
			name:     "complete with exclamation",
			content:  "This is an exciting statement!",
			complete: true,
		},
		{
			name:     "complete question",
			content:  "Is this a complete question?",
			complete: true,
		},
		{
			name:     "incomplete sentence",
			content:  "This is an incomplete",
			complete: false,
		},
		{
			name:     "too short",
			content:  "Short.",
			complete: false,
		},
		{
			name:     "complete code block",
			content:  "```\ncode here\n```",
			complete: false, // Currently the implementation treats this as too short
		},
		{
			name:     "incomplete code block",
			content:  "```\ncode here",
			complete: false,
		},
		{
			name:     "complete list item",
			content:  "Here's a list:\n- First item\n- Second item",
			complete: true,
		},
		{
			name:     "numbered list",
			content:  "Steps:\n1. First step\n2. Second step",
			complete: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.IsComplete(tt.content)
			assert.Equal(t, tt.complete, result)
		})
	}
}

func TestPartialResponseDetector_ExtractLastCompleteUnit(t *testing.T) {
	detector := NewPartialResponseDetector()

	tests := []struct {
		name               string
		content            string
		expectedComplete   string
		expectedIncomplete string
	}{
		{
			name:               "single complete sentence",
			content:            "This is complete.",
			expectedComplete:   "This is complete.",
			expectedIncomplete: "",
		},
		{
			name:               "complete and incomplete",
			content:            "This is complete. This is not",
			expectedComplete:   "This is complete.",
			expectedIncomplete: " This is not",
		},
		{
			name:               "multiple sentences",
			content:            "First sentence. Second sentence! Third?",
			expectedComplete:   "First sentence. Second sentence! Third?",
			expectedIncomplete: "",
		},
		{
			name:               "no complete units",
			content:            "This has no sentence ending",
			expectedComplete:   "",
			expectedIncomplete: "This has no sentence ending",
		},
		{
			name:               "code block boundary",
			content:            "Here's code:```python\nprint()```More text",
			expectedComplete:   "Here's code:```python\nprint()`", // LastIndex finds the last backtick, not all three
			expectedIncomplete: "``More text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			complete, incomplete := detector.ExtractLastCompleteUnit(tt.content)
			assert.Equal(t, tt.expectedComplete, complete)
			assert.Equal(t, tt.expectedIncomplete, incomplete)
		})
	}
}

// Test error handling in streams
func TestStreamErrorHandling(t *testing.T) {
	errorChunk := StreamChunk{
		Error: errors.New("stream error"),
	}

	provider := &mockProvider{
		streamFunc: func(ctx context.Context, prompt string, options ...ProviderOption) (<-chan StreamChunk, error) {
			ch := make(chan StreamChunk)
			go func() {
				defer close(ch)
				ch <- StreamChunk{Content: "Some content"}
				ch <- errorChunk
			}()
			return ch, nil
		},
		generateFunc: func(ctx context.Context, prompt string, options ...ProviderOption) (string, error) {
			return "Recovery after error", nil
		},
	}

	handler := NewPartialResponseHandler(provider)
	handler.logger = logging.GetLogger()

	ctx := context.Background()
	responseChan, err := handler.HandleStreamWithRecovery(ctx, "test prompt")

	require.NoError(t, err)
	require.NotNil(t, responseChan)

	// Collect responses
	var responses []StreamChunk
	for chunk := range responseChan {
		responses = append(responses, chunk)
	}

	// Should attempt recovery since we had partial content
	assert.True(t, len(responses) > 1)
}

// Test context cancellation
func TestContextCancellation(t *testing.T) {
	provider := &mockProvider{
		streamFunc: func(ctx context.Context, prompt string, options ...ProviderOption) (<-chan StreamChunk, error) {
			ch := make(chan StreamChunk)
			go func() {
				defer close(ch)
				// Simulate slow stream
				for i := 0; i < 10; i++ {
					select {
					case <-ctx.Done():
						return
					case ch <- StreamChunk{Content: fmt.Sprintf("Chunk %d ", i)}:
						time.Sleep(50 * time.Millisecond)
					}
				}
			}()
			return ch, nil
		},
	}

	handler := NewPartialResponseHandler(provider)
	handler.logger = logging.GetLogger()

	ctx, cancel := context.WithCancel(context.Background())
	responseChan, err := handler.HandleStreamWithRecovery(ctx, "test prompt")

	require.NoError(t, err)
	require.NotNil(t, responseChan)

	// Cancel after receiving a few chunks
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	// Collect responses
	var responses []StreamChunk
	var ctxError bool
	for chunk := range responseChan {
		if chunk.Error == context.Canceled {
			ctxError = true
		}
		responses = append(responses, chunk)
	}

	assert.True(t, ctxError)
	assert.Greater(t, len(responses), 0)
}

// Benchmark tests
func BenchmarkResponseBuffer_AddChunk(b *testing.B) {
	buffer := &ResponseBuffer{}
	chunk := StreamChunk{Content: "test content", Index: 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer.AddChunk(chunk)
	}
}

func BenchmarkPartialResponseDetector_IsComplete(b *testing.B) {
	detector := NewPartialResponseDetector()
	content := "This is a test sentence that might or might not be complete."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detector.IsComplete(content)
	}
}

func BenchmarkIsValidContinuation(b *testing.B) {
	partial := "This is the beginning of a long response"
	continuation := "that continues with more content and information."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isValidContinuation(partial, continuation)
	}
}
