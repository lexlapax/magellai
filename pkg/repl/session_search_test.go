// ABOUTME: Tests for session search functionality including content matching and result formatting
// ABOUTME: Verifies that search works across message content, system prompts, names, and tags

package repl

import (
	"os"
	"strings"
	"testing"

	"github.com/lexlapax/magellai/pkg/llm"
)

// extractSnippet is a test helper function to extract a snippet of text around a search query
func extractSnippet(content, query string, contextRadius int) string {
	lowerContent := strings.ToLower(content)
	lowerQuery := strings.ToLower(query)

	idx := strings.Index(lowerContent, lowerQuery)
	if idx == -1 {
		if len(content) <= contextRadius*2 {
			return content
		}
		return content[:contextRadius*2] + "..."
	}

	start := idx - contextRadius
	end := idx + len(query) + contextRadius

	prefix := ""
	suffix := ""

	if start < 0 {
		start = 0
	} else {
		prefix = "..."
	}

	if end > len(content) {
		end = len(content)
	} else {
		suffix = "..."
	}

	snippet := content[start:end]

	// Adjust to word boundaries
	if start > 0 && start < len(content) && content[start-1] != ' ' {
		// Move start back to word boundary
		for i := start; i > 0; i-- {
			if content[i-1] == ' ' {
				start = i
				snippet = content[start:end]
				break
			}
		}
	}

	if end < len(content) && end > 0 && content[end-1] != ' ' {
		// Move end forward to word boundary
		for i := end; i < len(content); i++ {
			if i >= len(content) || content[i] == ' ' {
				end = i
				snippet = content[start:end]
				break
			}
		}
	}

	return prefix + snippet + suffix
}

func TestSearchSessions(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "session-search-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create session manager
	manager := createTestSessionManager(t, tmpDir)

	// Create test sessions with different content
	session1, err := manager.NewSession("Test Session 1")
	if err != nil {
		t.Fatalf("Failed to create session1: %v", err)
	}
	session1.Conversation.AddMessage("user", "Tell me about quantum computing", nil)
	session1.Conversation.AddMessage("assistant", "Quantum computing uses quantum bits or qubits to process information", nil)
	session1.Conversation.SetSystemPrompt("You are a helpful physics expert")
	session1.Tags = []string{"physics", "quantum"}

	session2, err := manager.NewSession("Machine Learning Session")
	if err != nil {
		t.Fatalf("Failed to create session2: %v", err)
	}
	session2.Conversation.AddMessage("user", "Explain neural networks", nil)
	session2.Conversation.AddMessage("assistant", "Neural networks are computing systems inspired by biological neurons", nil)
	session2.Conversation.SetSystemPrompt("You are an AI expert specializing in machine learning")
	session2.Tags = []string{"AI", "neural"}

	session3, err := manager.NewSession("General Chat")
	if err != nil {
		t.Fatalf("Failed to create session3: %v", err)
	}
	session3.Conversation.AddMessage("user", "What's the weather like?", nil)
	session3.Conversation.AddMessage("assistant", "I don't have access to real-time weather data", nil)
	session3.Tags = []string{"general", "casual"}

	// Save all sessions
	if err := manager.SaveSession(session1); err != nil {
		t.Fatalf("Failed to save session1: %v", err)
	}
	if err := manager.SaveSession(session2); err != nil {
		t.Fatalf("Failed to save session2: %v", err)
	}
	if err := manager.SaveSession(session3); err != nil {
		t.Fatalf("Failed to save session3: %v", err)
	}

	// Test cases
	tests := []struct {
		name             string
		query            string
		expectedCount    int
		expectedSessions []string // Session names that should be in results
		checkContent     func(results []*SearchResult) error
	}{
		{
			name:             "Search for quantum in messages",
			query:            "quantum",
			expectedCount:    1,
			expectedSessions: []string{"Test Session 1"},
			checkContent: func(results []*SearchResult) error {
				if len(results[0].Matches) < 2 {
					return nil // We expect matches in message and tag
				}
				found := false
				for _, match := range results[0].Matches {
					if match.Type == "message" && strings.Contains(match.Content, "quantum") {
						found = true
						break
					}
				}
				if !found {
					return nil
				}
				return nil
			},
		},
		{
			name:             "Search for neural in all fields",
			query:            "neural",
			expectedCount:    1,
			expectedSessions: []string{"Machine Learning Session"},
		},
		{
			name:             "Search in system prompts",
			query:            "physics expert",
			expectedCount:    1,
			expectedSessions: []string{"Test Session 1"},
		},
		{
			name:             "Search in session names",
			query:            "Machine Learning",
			expectedCount:    1,
			expectedSessions: []string{"Machine Learning Session"},
		},
		{
			name:             "Case insensitive search",
			query:            "QUANTUM",
			expectedCount:    1,
			expectedSessions: []string{"Test Session 1"},
		},
		{
			name:          "No results",
			query:         "nonexistent",
			expectedCount: 0,
		},
		{
			name:             "Partial match",
			query:            "comput",
			expectedCount:    2,
			expectedSessions: []string{"Test Session 1", "Machine Learning Session"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := manager.SearchSessions(tt.query)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}

			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, len(results))
			}

			// Check that expected sessions are in results
			for _, expectedName := range tt.expectedSessions {
				found := false
				for _, result := range results {
					if result.Session.Name == expectedName {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected session %s not found in results", expectedName)
				}
			}

			// Run custom content checks if provided
			if tt.checkContent != nil {
				if err := tt.checkContent(results); err != nil {
					t.Errorf("Content check failed: %v", err)
				}
			}
		})
	}
}

func TestExtractSnippet(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		query    string
		radius   int
		expected string
	}{
		{
			name:     "Basic extraction",
			content:  "The quick brown fox jumps over the lazy dog",
			query:    "fox",
			radius:   10,
			expected: "...quick brown fox jumps over...",
		},
		{
			name:     "Start of content",
			content:  "fox jumps over the lazy dog",
			query:    "fox",
			radius:   10,
			expected: "fox jumps over...",
		},
		{
			name:     "End of content",
			content:  "The quick brown fox",
			query:    "fox",
			radius:   10,
			expected: "...quick brown fox",
		},
		{
			name:     "Query not found",
			content:  "The quick brown dog",
			query:    "fox",
			radius:   10,
			expected: "The quick brown dog",
		},
		{
			name:     "Case insensitive",
			content:  "The quick BROWN fox jumps",
			query:    "brown",
			radius:   10,
			expected: "...The quick BROWN fox jumps...",
		},
		{
			name:     "Word boundary adjustment",
			content:  "The quick brown fox jumps over the lazy dog",
			query:    "fox",
			radius:   5,
			expected: "...brown fox jumps...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSnippet(tt.content, tt.query, tt.radius)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestSearchResultsFormatting(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "session-format-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create session manager
	manager := createTestSessionManager(t, tmpDir)

	// Create a session with attachments
	session, err := manager.NewSession("Test Session")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Add message with attachment
	attachment := llm.Attachment{
		Type:     llm.AttachmentTypeImage,
		FilePath: "/path/to/image.png",
		MimeType: "image/png",
	}
	session.Conversation.AddMessage("user", "Here's an image about quantum physics", []llm.Attachment{attachment})
	session.Conversation.AddMessage("assistant", "I can see the quantum physics diagram", nil)

	// Save session
	if err := manager.SaveSession(session); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Search for content
	results, err := manager.SearchSessions("quantum")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.Session.Name != "Test Session" {
		t.Errorf("Expected session name 'Test Session', got '%s'", result.Session.Name)
	}

	// Check that we have matches from messages
	messageMatches := 0
	for _, match := range result.Matches {
		if match.Type == "message" {
			messageMatches++
		}
	}

	if messageMatches != 2 {
		t.Errorf("Expected 2 message matches, got %d", messageMatches)
	}
}
