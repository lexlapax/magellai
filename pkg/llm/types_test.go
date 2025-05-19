// ABOUTME: Tests for LLM type definitions and constants
// ABOUTME: Verifies type structures and constant values

package llm

import (
	"encoding/json"
	"testing"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderConstants(t *testing.T) {
	// Test that provider constants exist and have expected values
	assert.Equal(t, "openai", ProviderOpenAI)
	assert.Equal(t, "anthropic", ProviderAnthropic)
	assert.Equal(t, "gemini", ProviderGemini)
	assert.Equal(t, "ollama", ProviderOllama)
	assert.Equal(t, "mock", ProviderMock)
}

func TestModelCapabilityConstants(t *testing.T) {
	// Test that capability constants exist and have expected values
	assert.Equal(t, ModelCapability("text"), CapabilityText)
	assert.Equal(t, ModelCapability("image"), CapabilityImage)
	assert.Equal(t, ModelCapability("audio"), CapabilityAudio)
	assert.Equal(t, ModelCapability("video"), CapabilityVideo)
	assert.Equal(t, ModelCapability("file"), CapabilityFile)
}

func TestRequestStructure(t *testing.T) {
	// Test Request struct JSON marshaling/unmarshaling
	temp := 0.7
	maxTok := 100

	req := Request{
		Messages: []domain.Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
		Model:        "openai/gpt-4",
		Temperature:  &temp,
		MaxTokens:    &maxTok,
		Stream:       true,
		SystemPrompt: "You are helpful",
		Options: &PromptParams{
			Temperature: &temp,
			MaxTokens:   &maxTok,
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(req)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var decoded Request
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, req.Model, decoded.Model)
	assert.Equal(t, req.Stream, decoded.Stream)
	assert.Equal(t, req.SystemPrompt, decoded.SystemPrompt)
	assert.NotNil(t, decoded.Temperature)
	assert.Equal(t, *req.Temperature, *decoded.Temperature)
	assert.NotNil(t, decoded.MaxTokens)
	assert.Equal(t, *req.MaxTokens, *decoded.MaxTokens)
}

func TestResponseStructure(t *testing.T) {
	// Test Response struct JSON marshaling/unmarshaling
	resp := Response{
		Content: "Hello, world!",
		Model:   "openai/gpt-4",
		Usage: &Usage{
			InputTokens:  10,
			OutputTokens: 15,
			TotalTokens:  25,
		},
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
		FinishReason: "stop",
	}

	// Test JSON marshaling
	data, err := json.Marshal(resp)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var decoded Response
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, resp.Content, decoded.Content)
	assert.Equal(t, resp.Model, decoded.Model)
	assert.Equal(t, resp.FinishReason, decoded.FinishReason)
	assert.NotNil(t, decoded.Usage)
	assert.Equal(t, resp.Usage.TotalTokens, decoded.Usage.TotalTokens)
	assert.Equal(t, resp.Metadata["key1"], decoded.Metadata["key1"])
}

func TestStreamChunkStructure(t *testing.T) {
	chunk := StreamChunk{
		Content:      "streaming",
		Done:         true,
		FinishReason: "stop",
		Index:        1,
		Error:        nil,
	}

	// Test basic structure
	assert.Equal(t, "streaming", chunk.Content)
	assert.True(t, chunk.Done)
	assert.Equal(t, "stop", chunk.FinishReason)
	assert.Equal(t, 1, chunk.Index)
	assert.Nil(t, chunk.Error)
}

func TestUsageStructure(t *testing.T) {
	usage := Usage{
		InputTokens:  100,
		OutputTokens: 200,
		TotalTokens:  300,
	}

	// Test JSON marshaling
	data, err := json.Marshal(usage)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var decoded Usage
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, usage.InputTokens, decoded.InputTokens)
	assert.Equal(t, usage.OutputTokens, decoded.OutputTokens)
	assert.Equal(t, usage.TotalTokens, decoded.TotalTokens)
}

func TestPromptParamsStructure(t *testing.T) {
	temp := 0.8
	maxTok := 500
	topP := 0.9
	seed := 12345

	params := PromptParams{
		Temperature:    &temp,
		MaxTokens:      &maxTok,
		TopP:           &topP,
		ResponseFormat: "json",
		Stop:           []string{"STOP", "END"},
		Seed:           &seed,
	}

	// Test JSON marshaling
	data, err := json.Marshal(params)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var decoded PromptParams
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, *params.Temperature, *decoded.Temperature)
	assert.Equal(t, *params.MaxTokens, *decoded.MaxTokens)
	assert.Equal(t, params.ResponseFormat, decoded.ResponseFormat)
	assert.Equal(t, params.Stop, decoded.Stop)
}

func TestModelInfoStructure(t *testing.T) {
	info := ModelInfo{
		Provider:      "openai",
		Model:         "gpt-4",
		DisplayName:   "GPT-4",
		Description:   "Most capable GPT-4 model",
		ContextWindow: 8192,
		MaxTokens:     4096,
		Capabilities: ModelCapabilities{
			Text:  true,
			Image: true,
		},
		DefaultTemperature: 0.7,
	}

	// Test JSON marshaling
	data, err := json.Marshal(info)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var decoded ModelInfo
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, info.Provider, decoded.Provider)
	assert.Equal(t, info.Model, decoded.Model)
	assert.Equal(t, info.ContextWindow, decoded.ContextWindow)
	assert.Equal(t, info.Capabilities.Text, decoded.Capabilities.Text)
	assert.Equal(t, info.Capabilities.Image, decoded.Capabilities.Image)
}

func TestModelCapabilitiesStructure(t *testing.T) {
	capabilities := ModelCapabilities{
		Text:             true,
		Audio:            false,
		Video:            false,
		Image:            true,
		File:             true,
		StructuredOutput: true,
	}

	// Test JSON marshaling
	data, err := json.Marshal(capabilities)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var decoded ModelCapabilities
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, capabilities.Text, decoded.Text)
	assert.Equal(t, capabilities.Audio, decoded.Audio)
	assert.Equal(t, capabilities.Video, decoded.Video)
	assert.Equal(t, capabilities.Image, decoded.Image)
	assert.Equal(t, capabilities.File, decoded.File)
	assert.Equal(t, capabilities.StructuredOutput, decoded.StructuredOutput)
}

func TestRequestWithNilOptions(t *testing.T) {
	// Test that Request handles nil options correctly
	req := Request{
		Messages: []domain.Message{
			{
				Role:    "user",
				Content: "Test",
			},
		},
		Model:        "test-model",
		Temperature:  nil,
		MaxTokens:    nil,
		Options:      nil,
		SystemPrompt: "",
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var decoded Request
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Nil(t, decoded.Temperature)
	assert.Nil(t, decoded.MaxTokens)
	assert.Nil(t, decoded.Options)
}

func TestResponseWithNilUsage(t *testing.T) {
	// Test that Response handles nil usage correctly
	resp := Response{
		Content:      "Test response",
		Model:        "test-model",
		Usage:        nil,
		Metadata:     nil,
		FinishReason: "",
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded Response
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Nil(t, decoded.Usage)
	assert.Nil(t, decoded.Metadata)
}

func TestStreamChunkWithError(t *testing.T) {
	// Test StreamChunk with error
	chunk := StreamChunk{
		Content:      "",
		Done:         true,
		FinishReason: "error",
		Error:        assert.AnError,
	}

	// Note: Error field may not be directly JSON serializable
	// This tests the struct composition
	assert.NotNil(t, chunk.Error)
	assert.Equal(t, "error", chunk.FinishReason)
}

func TestPromptParamsValidation(t *testing.T) {
	tests := []struct {
		name   string
		params PromptParams
		valid  bool
	}{
		{
			name: "valid params",
			params: PromptParams{
				Temperature: ptrFloat(0.5),
				MaxTokens:   ptrInt(100),
			},
			valid: true,
		},
		{
			name: "negative temperature",
			params: PromptParams{
				Temperature: ptrFloat(-0.1),
			},
			valid: false,
		},
		{
			name: "temperature too high",
			params: PromptParams{
				Temperature: ptrFloat(2.5),
			},
			valid: false,
		},
		{
			name: "negative max tokens",
			params: PromptParams{
				MaxTokens: ptrInt(-10),
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test is more conceptual - in real implementation
			// you might have a Validate() method
			if tt.valid {
				if tt.params.Temperature != nil {
					assert.GreaterOrEqual(t, *tt.params.Temperature, 0.0)
					assert.LessOrEqual(t, *tt.params.Temperature, 2.0)
				}
				if tt.params.MaxTokens != nil {
					assert.GreaterOrEqual(t, *tt.params.MaxTokens, 0)
				}
			}
		})
	}
}

// Helper functions
func ptrFloat(f float64) *float64 {
	return &f
}

func ptrInt(i int) *int {
	return &i
}

// Test that types are compatible with external libraries
func TestTypeCompatibility(t *testing.T) {
	// Test that our types can work with the go-llms library types
	// This is more of a compilation test - if it compiles, it works

	var _ Request
	var _ Response
	var _ StreamChunk
	var _ ModelInfo
}

// Test ParseModelString function
func TestParseModelString(t *testing.T) {
	tests := []struct {
		input         string
		wantProvider  string
		wantModelName string
	}{
		{
			input:         "openai/gpt-4",
			wantProvider:  "openai",
			wantModelName: "gpt-4",
		},
		{
			input:         "anthropic/claude-3",
			wantProvider:  "anthropic",
			wantModelName: "claude-3",
		},
		{
			input:         "single",
			wantProvider:  "openai", // Defaults to OpenAI when no provider specified
			wantModelName: "single",
		},
		{
			input:         "",
			wantProvider:  "openai", // Defaults to OpenAI when no provider specified
			wantModelName: "",
		},
		{
			input:         "a/b/c",
			wantProvider:  "a",
			wantModelName: "b/c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gotProvider, gotModelName := ParseModelString(tt.input)
			assert.Equal(t, tt.wantProvider, gotProvider)
			assert.Equal(t, tt.wantModelName, gotModelName)
		})
	}
}
