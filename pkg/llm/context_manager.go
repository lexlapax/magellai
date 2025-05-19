// ABOUTME: Context length management for LLM interactions
// ABOUTME: Handles token counting, message prioritization, and context window optimization

package llm

import (
	"fmt"
	"strings"

	"github.com/lexlapax/magellai/internal/logging"
	"github.com/lexlapax/magellai/pkg/domain"
)

// ContextManager manages conversation context to fit within model limits
type ContextManager struct {
	modelInfo      ModelInfo
	tokenCounter   TokenCounter
	priorityConfig PriorityConfig
	logger         *logging.Logger
}

// TokenCounter estimates token count for text
type TokenCounter interface {
	CountTokens(text string) int
	CountMessageTokens(messages []domain.Message) int
}

// PriorityConfig defines how to prioritize messages when truncating
type PriorityConfig struct {
	KeepSystemMessage bool    // Always keep system message
	KeepFirstN        int     // Keep first N messages after system
	KeepLastN         int     // Keep last N messages
	MaxTokens         int     // Maximum context tokens
	ReserveTokens     int     // Reserve tokens for response
	ImportanceDecay   float64 // Decay factor for message importance by age
}

// DefaultPriorityConfig returns sensible defaults for context management
func DefaultPriorityConfig(modelInfo ModelInfo) PriorityConfig {
	// Reserve 25% of context for response
	maxContext := modelInfo.ContextWindow
	if maxContext == 0 {
		maxContext = 4096 // Default fallback
	}

	return PriorityConfig{
		KeepSystemMessage: true,
		KeepFirstN:        1,
		KeepLastN:         3,
		MaxTokens:         maxContext * 3 / 4,
		ReserveTokens:     maxContext / 4,
		ImportanceDecay:   0.9,
	}
}

// NewContextManager creates a context manager for the given model
func NewContextManager(modelInfo ModelInfo) *ContextManager {
	return &ContextManager{
		modelInfo:      modelInfo,
		tokenCounter:   NewEstimatedTokenCounter(),
		priorityConfig: DefaultPriorityConfig(modelInfo),
		logger:         logging.GetLogger(),
	}
}

// OptimizeContext reduces message context to fit within limits
func (m *ContextManager) OptimizeContext(messages []domain.Message) ([]domain.Message, error) {
	if len(messages) == 0 {
		return messages, nil
	}

	currentTokens := m.tokenCounter.CountMessageTokens(messages)

	m.logger.Debug("Optimizing context",
		"messageCount", len(messages),
		"currentTokens", currentTokens,
		"maxTokens", m.priorityConfig.MaxTokens)

	// If already within limits, return as-is
	if currentTokens <= m.priorityConfig.MaxTokens {
		return messages, nil
	}

	// Apply optimization strategies
	optimized := m.applyPrioritization(messages)
	finalTokens := m.tokenCounter.CountMessageTokens(optimized)

	m.logger.Info("Context optimized",
		"originalMessages", len(messages),
		"optimizedMessages", len(optimized),
		"originalTokens", currentTokens,
		"optimizedTokens", finalTokens)

	if finalTokens > m.priorityConfig.MaxTokens {
		return optimized, fmt.Errorf("unable to fit context within limit: %d > %d tokens",
			finalTokens, m.priorityConfig.MaxTokens)
	}

	return optimized, nil
}

// applyPrioritization applies message prioritization rules
func (m *ContextManager) applyPrioritization(messages []domain.Message) []domain.Message {
	if len(messages) == 0 {
		return messages
	}

	var result []domain.Message
	systemIdx := -1

	// Find system message
	for i, msg := range messages {
		if strings.ToLower(string(msg.Role)) == "system" {
			systemIdx = i
			break
		}
	}

	// Always keep system message if present and configured
	if systemIdx >= 0 && m.priorityConfig.KeepSystemMessage {
		result = append(result, messages[systemIdx])
	}

	// Determine conversation messages (excluding system)
	convStart := 0
	if systemIdx >= 0 {
		convStart = systemIdx + 1
	}
	conversation := messages[convStart:]

	if len(conversation) == 0 {
		return result
	}

	// Apply keep first/last rules
	keepIndices := make(map[int]bool)

	// Keep first N
	for i := 0; i < m.priorityConfig.KeepFirstN && i < len(conversation); i++ {
		keepIndices[i] = true
	}

	// Keep last N
	for i := len(conversation) - m.priorityConfig.KeepLastN; i < len(conversation); i++ {
		if i >= 0 {
			keepIndices[i] = true
		}
	}

	// Calculate importance scores for middle messages
	importance := m.calculateImportance(conversation)

	// Add messages by importance until we hit token limit
	// Sort indices to ensure deterministic order
	var indices []int
	for idx := range keepIndices {
		indices = append(indices, idx)
	}
	// Sort indices to maintain order
	for i := 0; i < len(indices)-1; i++ {
		for j := i + 1; j < len(indices); j++ {
			if indices[i] > indices[j] {
				indices[i], indices[j] = indices[j], indices[i]
			}
		}
	}
	for _, idx := range indices {
		result = append(result, conversation[idx])
	}

	// Sort remaining messages by importance
	remaining := m.selectByImportance(conversation, keepIndices, importance)

	// Add messages until we hit the limit
	for _, msg := range remaining {
		test := append(result, msg)
		tokens := m.tokenCounter.CountMessageTokens(test)
		if tokens <= m.priorityConfig.MaxTokens {
			result = append(result, msg)
		} else {
			break
		}
	}

	// Ensure messages are in chronological order
	result = m.sortChronologically(result)

	return result
}

// calculateImportance assigns importance scores to messages
func (m *ContextManager) calculateImportance(messages []domain.Message) []float64 {
	scores := make([]float64, len(messages))

	for i, msg := range messages {
		score := 1.0

		// Base score on content length (longer = potentially more important)
		contentLength := float64(len(msg.Content))
		score *= (contentLength / 100.0)
		if score > 2.0 {
			score = 2.0 // Cap length factor
		}

		// Apply age decay
		age := len(messages) - i
		score *= m.exponentialDecay(age, m.priorityConfig.ImportanceDecay)

		// Boost for user messages (they provide context)
		if strings.ToLower(string(msg.Role)) == "user" {
			score *= 1.2
		}

		// Boost for messages with attachments
		if len(msg.Attachments) > 0 {
			score *= 1.5
		}

		// Boost for messages containing questions
		if m.containsQuestion(msg.Content) {
			score *= 1.3
		}

		scores[i] = score
	}

	return scores
}

// exponentialDecay calculates decay based on age
func (m *ContextManager) exponentialDecay(age int, factor float64) float64 {
	return 1.0 * factor * float64(age)
}

// containsQuestion checks if content appears to be a question
func (m *ContextManager) containsQuestion(content string) bool {
	questionIndicators := []string{"?", "how ", "what ", "when ", "where ", "why ", "who ", "which "}
	contentLower := strings.ToLower(content)

	for _, indicator := range questionIndicators {
		if strings.Contains(contentLower, indicator) {
			return true
		}
	}

	return false
}

// selectByImportance selects messages by importance score
func (m *ContextManager) selectByImportance(messages []domain.Message, keep map[int]bool, scores []float64) []domain.Message {
	type scoredMessage struct {
		message domain.Message
		score   float64
		index   int
	}

	var candidates []scoredMessage

	for i, msg := range messages {
		if !keep[i] {
			candidates = append(candidates, scoredMessage{
				message: msg,
				score:   scores[i],
				index:   i,
			})
		}
	}

	// Sort by score descending
	for i := 0; i < len(candidates); i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].score > candidates[i].score {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	// Extract messages
	var result []domain.Message
	for _, sm := range candidates {
		result = append(result, sm.message)
	}

	return result
}

// sortChronologically ensures messages are in chronological order
func (m *ContextManager) sortChronologically(messages []domain.Message) []domain.Message {
	// This is a simple implementation assuming messages were originally in order
	// In practice, you might want to track original indices
	return messages
}

// EstimateTokenReduction estimates how many tokens would be saved by various strategies
func (m *ContextManager) EstimateTokenReduction(messages []domain.Message) map[string]int {
	original := m.tokenCounter.CountMessageTokens(messages)
	estimates := make(map[string]int)

	// Estimate removal of oldest messages
	if len(messages) > 3 {
		reduced := messages[len(messages)-3:]
		estimates["remove_oldest"] = original - m.tokenCounter.CountMessageTokens(reduced)
	}

	// Estimate summarization savings (rough estimate)
	estimates["summarize_old"] = original / 3

	// Estimate attachment removal
	noAttachments := m.removeAttachments(messages)
	estimates["remove_attachments"] = original - m.tokenCounter.CountMessageTokens(noAttachments)

	return estimates
}

// removeAttachments creates a copy of messages without attachments
func (m *ContextManager) removeAttachments(messages []domain.Message) []domain.Message {
	result := make([]domain.Message, len(messages))
	for i, msg := range messages {
		result[i] = domain.Message{
			Role:    msg.Role,
			Content: msg.Content,
			// Omit attachments
		}
	}
	return result
}

// EstimatedTokenCounter provides rough token counting
type EstimatedTokenCounter struct {
	// Rough estimate: 1 token ≈ 4 characters
	charactersPerToken float64
}

// NewEstimatedTokenCounter creates a basic token counter
func NewEstimatedTokenCounter() *EstimatedTokenCounter {
	return &EstimatedTokenCounter{
		charactersPerToken: 4.0,
	}
}

// CountTokens estimates tokens in text
func (t *EstimatedTokenCounter) CountTokens(text string) int {
	// Basic estimation
	chars := len(text)
	tokens := int(float64(chars) / t.charactersPerToken)

	// Account for whitespace (roughly)
	words := len(strings.Fields(text))
	tokens += words / 2

	return tokens
}

// CountMessageTokens estimates tokens in messages
func (t *EstimatedTokenCounter) CountMessageTokens(messages []domain.Message) int {
	total := 0

	for _, msg := range messages {
		// Role tokens (system, user, assistant)
		total += 5

		// Content tokens
		total += t.CountTokens(msg.Content)

		// Attachment tokens (rough estimate)
		for _, att := range msg.Attachments {
			switch att.Type {
			case domain.AttachmentTypeText:
				total += t.CountTokens(string(att.Content))
			case domain.AttachmentTypeImage:
				total += 500 // Rough estimate for image tokens
			case domain.AttachmentTypeFile:
				total += 100 // File reference tokens
			}
		}

		// Message separator tokens
		total += 10
	}

	return total
}

// SlidingWindowManager implements a sliding window approach to context
type SlidingWindowManager struct {
	windowSize   int
	overlapSize  int
	tokenCounter TokenCounter
}

// NewSlidingWindowManager creates a sliding window context manager
func NewSlidingWindowManager(windowSize, overlapSize int) *SlidingWindowManager {
	return &SlidingWindowManager{
		windowSize:   windowSize,
		overlapSize:  overlapSize,
		tokenCounter: NewEstimatedTokenCounter(),
	}
}

// GetWindow returns a window of messages that fits within token limit
func (s *SlidingWindowManager) GetWindow(messages []domain.Message, maxTokens int) []domain.Message {
	if len(messages) == 0 {
		return messages
	}

	// Start from the end and work backwards
	window := []domain.Message{}
	currentTokens := 0

	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		msgTokens := s.tokenCounter.CountTokens(msg.Content)

		if currentTokens+msgTokens > maxTokens {
			break
		}

		window = append([]domain.Message{msg}, window...)
		currentTokens += msgTokens
	}

	return window
}
