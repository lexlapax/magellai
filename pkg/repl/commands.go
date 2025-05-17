// ABOUTME: Implements REPL command handlers for both slash and colon commands
// ABOUTME: Provides functionality for session management, model switching, and chat configuration

package repl

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lexlapax/magellai/pkg/llm"
)

// saveSession saves the current session
func (r *REPL) saveSession(args []string) error {
	// If a name is provided, update the session name
	if len(args) > 0 {
		r.session.Name = strings.Join(args, " ")
	}

	if err := r.manager.SaveSession(r.session); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	fmt.Fprintf(r.writer, "Session saved: %s\n", r.session.ID)
	return nil
}

// loadSession loads a previous session
func (r *REPL) loadSession(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("session ID required")
	}

	sessionID := args[0]
	session, err := r.manager.LoadSession(sessionID)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	// Save current session before switching
	if err := r.manager.SaveSession(r.session); err != nil {
		fmt.Fprintf(r.writer, "Warning: Failed to save current session: %v\n", err)
	}

	r.session = session
	fmt.Fprintf(r.writer, "Loaded session: %s\n", session.Name)
	fmt.Fprintf(r.writer, "Model: %s\n", session.Conversation.Model)
	fmt.Fprintf(r.writer, "Messages: %d\n", len(session.Conversation.Messages))
	return nil
}

// resetConversation clears the conversation history
func (r *REPL) resetConversation() error {
	r.session.Conversation.Reset()
	fmt.Fprintln(r.writer, "Conversation history cleared.")
	return nil
}

// showModel displays the current model
func (r *REPL) showModel() error {
	fmt.Fprintf(r.writer, "Current model: %s\n", r.session.Conversation.Model)
	fmt.Fprintf(r.writer, "Provider: %s\n", r.session.Conversation.Provider)
	if r.session.Conversation.Temperature > 0 {
		fmt.Fprintf(r.writer, "Temperature: %.2f\n", r.session.Conversation.Temperature)
	}
	if r.session.Conversation.MaxTokens > 0 {
		fmt.Fprintf(r.writer, "Max tokens: %d\n", r.session.Conversation.MaxTokens)
	}
	return nil
}

// setSystemPrompt sets or shows the system prompt
func (r *REPL) setSystemPrompt(args []string) error {
	if len(args) == 0 {
		// Show current system prompt
		if r.session.Conversation.SystemPrompt == "" {
			fmt.Fprintln(r.writer, "No system prompt set.")
		} else {
			fmt.Fprintf(r.writer, "System prompt: %s\n", r.session.Conversation.SystemPrompt)
		}
		return nil
	}

	// Set system prompt
	prompt := strings.Join(args, " ")
	r.session.Conversation.SetSystemPrompt(prompt)
	fmt.Fprintln(r.writer, "System prompt updated.")
	return nil
}

// showHistory displays conversation history
func (r *REPL) showHistory() error {
	messages := r.session.Conversation.Messages
	if len(messages) == 0 {
		fmt.Fprintln(r.writer, "No conversation history.")
		return nil
	}

	fmt.Fprintf(r.writer, "Conversation history (%d messages):\n\n", len(messages))
	for i, msg := range messages {
		fmt.Fprintf(r.writer, "[%d] %s:\n%s\n\n", i+1, title(msg.Role), msg.Content)
		if len(msg.Attachments) > 0 {
			fmt.Fprintf(r.writer, "Attachments: %d\n\n", len(msg.Attachments))
		}
	}
	return nil
}

// listSessions lists all available sessions
func (r *REPL) listSessions() error {
	sessions, err := r.manager.ListSessions()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Fprintln(r.writer, "No sessions found.")
		return nil
	}

	fmt.Fprintf(r.writer, "Available sessions (%d):\n\n", len(sessions))
	for _, session := range sessions {
		current := ""
		if session.ID == r.session.ID {
			current = " (current)"
		}
		fmt.Fprintf(r.writer, "%s: %s%s\n", session.ID, session.Name, current)
		fmt.Fprintf(r.writer, "  Created: %s\n", session.Created.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(r.writer, "  Messages: %d\n", session.MessageCount)
		if len(session.Tags) > 0 {
			fmt.Fprintf(r.writer, "  Tags: %s\n", strings.Join(session.Tags, ", "))
		}
		fmt.Fprintln(r.writer)
	}
	return nil
}

// attachFile attaches a file to the next message
func (r *REPL) attachFile(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("file path required")
	}

	filePath := args[0]

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Determine file type
	mimeType := "application/octet-stream"
	ext := strings.ToLower(filepath.Ext(filePath))

	var attachmentType llm.AttachmentType
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		attachmentType = llm.AttachmentTypeImage
		mimeType = "image/" + strings.TrimPrefix(ext, ".")
	case ".mp3", ".wav", ".ogg", ".m4a":
		attachmentType = llm.AttachmentTypeAudio
		mimeType = "audio/" + strings.TrimPrefix(ext, ".")
	case ".mp4", ".avi", ".mov", ".webm":
		attachmentType = llm.AttachmentTypeVideo
		mimeType = "video/" + strings.TrimPrefix(ext, ".")
	case ".txt", ".md", ".log", ".csv":
		attachmentType = llm.AttachmentTypeText
		mimeType = "text/plain"
	default:
		attachmentType = llm.AttachmentTypeFile
	}

	// Create attachment
	attachment := llm.Attachment{
		Type:     attachmentType,
		FilePath: filePath,
		MimeType: mimeType,
		Content:  string(content), // For now, store content directly
	}

	// Store in session metadata for next message
	if r.session.Metadata == nil {
		r.session.Metadata = make(map[string]interface{})
	}

	attachments, ok := r.session.Metadata["pending_attachments"].([]llm.Attachment)
	if !ok {
		attachments = []llm.Attachment{}
	}
	attachments = append(attachments, attachment)
	r.session.Metadata["pending_attachments"] = attachments

	fmt.Fprintf(r.writer, "Attached: %s (%s)\n", filepath.Base(filePath), mimeType)
	return nil
}

// listAttachments lists current attachments
func (r *REPL) listAttachments() error {
	attachments, ok := r.session.Metadata["pending_attachments"].([]llm.Attachment)
	if !ok || len(attachments) == 0 {
		fmt.Fprintln(r.writer, "No pending attachments.")
		return nil
	}

	fmt.Fprintf(r.writer, "Pending attachments (%d):\n", len(attachments))
	for i, att := range attachments {
		fmt.Fprintf(r.writer, "%d. %s (%s)\n", i+1, filepath.Base(att.FilePath), att.MimeType)
	}
	return nil
}

// switchModel switches to a different model
func (r *REPL) switchModel(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("model name required")
	}

	modelStr := args[0]

	// Parse model string
	parts := strings.Split(modelStr, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid model format, expected provider/model")
	}
	providerType := parts[0]
	modelName := parts[1]

	// Create new provider
	provider, err := llm.NewProvider(providerType, modelName)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Update session
	r.provider = provider
	r.session.Conversation.Model = modelStr
	r.session.Conversation.Provider = providerType

	fmt.Fprintf(r.writer, "Switched to model: %s\n", modelStr)
	return nil
}

// toggleStreaming toggles streaming mode
func (r *REPL) toggleStreaming(args []string) error {
	if len(args) > 0 {
		switch strings.ToLower(args[0]) {
		case "on", "true", "1":
			if err := r.config.SetValue("stream", true); err != nil {
				return fmt.Errorf("error enabling streaming: %w", err)
			}
			fmt.Fprintln(r.writer, "Streaming enabled.")
		case "off", "false", "0":
			if err := r.config.SetValue("stream", false); err != nil {
				return fmt.Errorf("error disabling streaming: %w", err)
			}
			fmt.Fprintln(r.writer, "Streaming disabled.")
		default:
			return fmt.Errorf("invalid value: use 'on' or 'off'")
		}
	} else {
		// Toggle current value
		current := r.config.GetBool("stream")
		if err := r.config.SetValue("stream", !current); err != nil {
			return fmt.Errorf("error toggling streaming: %w", err)
		}
		if !current {
			fmt.Fprintln(r.writer, "Streaming enabled.")
		} else {
			fmt.Fprintln(r.writer, "Streaming disabled.")
		}
	}
	return nil
}

// setTemperature sets the generation temperature
func (r *REPL) setTemperature(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("temperature value required")
	}

	temp, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid temperature: %w", err)
	}

	if temp < 0.0 || temp > 2.0 {
		return fmt.Errorf("temperature must be between 0.0 and 2.0")
	}

	r.session.Conversation.Temperature = temp
	fmt.Fprintf(r.writer, "Temperature set to: %.2f\n", temp)
	return nil
}

// setMaxTokens sets the maximum response tokens
func (r *REPL) setMaxTokens(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("max tokens value required")
	}

	maxTokens, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid max tokens: %w", err)
	}

	if maxTokens < 1 {
		return fmt.Errorf("max tokens must be positive")
	}

	r.session.Conversation.MaxTokens = maxTokens
	fmt.Fprintf(r.writer, "Max tokens set to: %d\n", maxTokens)
	return nil
}

// toggleMultiline toggles multi-line input mode
func (r *REPL) toggleMultiline() error {
	r.multiline = !r.multiline
	if r.multiline {
		fmt.Fprintln(r.writer, "Multi-line mode enabled. Press Enter twice to send.")
	} else {
		fmt.Fprintln(r.writer, "Multi-line mode disabled.")
	}
	return nil
}

// Helper function to capitalize first letter
func title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
