// ABOUTME: Implements REPL command handlers for both slash and colon commands
// ABOUTME: Provides functionality for session management, model switching, and chat configuration

package repl

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lexlapax/magellai/internal/logging"
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
	logging.LogInfo("Switching model", "from", r.session.Conversation.Model, "to", modelStr)

	// Parse model string
	parts := strings.Split(modelStr, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid model format, expected provider/model")
	}
	providerType := parts[0]
	modelName := parts[1]

	// Create new provider
	logging.LogDebug("Creating new provider", "provider", providerType, "model", modelName)
	provider, err := llm.NewProvider(providerType, modelName)
	if err != nil {
		logging.LogError(err, "Failed to create provider for model switch", "provider", providerType, "model", modelName)
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Update session
	r.provider = provider
	r.session.Conversation.Model = modelStr
	r.session.Conversation.Provider = providerType

	logging.LogInfo("Model switched successfully", "model", modelStr)
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

// verbosity sets the verbosity level
func (r *REPL) setVerbosity(args []string) error {
	if len(args) == 0 {
		// Show current verbosity
		level := r.config.GetString("verbosity")
		fmt.Fprintf(r.writer, "Current verbosity: %s\n", level)
		return nil
	}

	level := strings.ToLower(args[0])
	validLevels := []string{"debug", "info", "warn", "error"}
	isValid := false
	for _, valid := range validLevels {
		if level == valid {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("invalid verbosity level: %s (valid: debug, info, warn, error)", level)
	}

	if err := r.config.SetValue("verbosity", level); err != nil {
		return fmt.Errorf("failed to set verbosity: %w", err)
	}

	// Update the logger with new verbosity level
	if err := logging.SetLogLevel(level); err != nil {
		return fmt.Errorf("failed to update logger: %w", err)
	}

	fmt.Fprintf(r.writer, "Verbosity set to: %s\n", level)
	return nil
}

// setOutput sets the output format
func (r *REPL) setOutput(args []string) error {
	if len(args) == 0 {
		// Show current output format
		format := r.config.GetString("output_format")
		if format == "" {
			format = "text"
		}
		fmt.Fprintf(r.writer, "Current output format: %s\n", format)
		return nil
	}

	format := strings.ToLower(args[0])
	validFormats := []string{"text", "json", "yaml", "markdown"}
	isValid := false
	for _, valid := range validFormats {
		if format == valid {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("invalid output format: %s (valid: text, json, yaml, markdown)", format)
	}

	if err := r.config.SetValue("output_format", format); err != nil {
		return fmt.Errorf("failed to set output format: %w", err)
	}

	fmt.Fprintf(r.writer, "Output format set to: %s\n", format)
	return nil
}

// switchProfile switches to a different profile
func (r *REPL) switchProfile(args []string) error {
	if len(args) == 0 {
		// Show current profile
		profile := r.config.GetString("profile")
		if profile == "" {
			profile = "default"
		}
		fmt.Fprintf(r.writer, "Current profile: %s\n", profile)
		return nil
	}

	profile := args[0]

	// Check if profile exists
	profiles := r.config.GetString("available_profiles")
	if profiles != "" {
		availableProfiles := strings.Split(profiles, ",")
		found := false
		for _, p := range availableProfiles {
			if strings.TrimSpace(p) == profile {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("profile '%s' not found", profile)
		}
	}

	if err := r.config.SetValue("profile", profile); err != nil {
		return fmt.Errorf("failed to switch profile: %w", err)
	}

	fmt.Fprintf(r.writer, "Switched to profile: %s\n", profile)
	fmt.Fprintln(r.writer, "Note: Some settings may require a restart to take effect.")
	return nil
}

// removeAttachment removes a pending attachment
func (r *REPL) removeAttachment(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("attachment filename required")
	}

	filename := args[0]

	attachments, ok := r.session.Metadata["pending_attachments"].([]llm.Attachment)
	if !ok || len(attachments) == 0 {
		fmt.Fprintln(r.writer, "No pending attachments.")
		return nil
	}

	// Find and remove the attachment
	var newAttachments []llm.Attachment
	found := false
	for _, att := range attachments {
		if filepath.Base(att.FilePath) == filename {
			found = true
			continue
		}
		newAttachments = append(newAttachments, att)
	}

	if !found {
		return fmt.Errorf("attachment '%s' not found", filename)
	}

	r.session.Metadata["pending_attachments"] = newAttachments
	fmt.Fprintf(r.writer, "Removed attachment: %s\n", filename)
	return nil
}

// showConfig displays the current configuration
func (r *REPL) showConfig() error {
	// Show relevant configuration values
	fmt.Fprintln(r.writer, "Current configuration:")
	fmt.Fprintf(r.writer, "  model: %s\n", r.session.Conversation.Model)
	fmt.Fprintf(r.writer, "  stream: %v\n", r.config.GetBool("stream"))
	fmt.Fprintf(r.writer, "  temperature: %.2f\n", r.session.Conversation.Temperature)
	fmt.Fprintf(r.writer, "  max_tokens: %d\n", r.session.Conversation.MaxTokens)
	fmt.Fprintf(r.writer, "  verbosity: %s\n", r.config.GetString("verbosity"))
	fmt.Fprintf(r.writer, "  output_format: %s\n", r.config.GetString("output_format"))
	fmt.Fprintf(r.writer, "  profile: %s\n", r.config.GetString("profile"))

	return nil
}

// setConfig sets a configuration value
func (r *REPL) setConfig(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: /config set <key> <value>")
	}

	key := args[0]
	value := strings.Join(args[1:], " ")

	// Handle different value types
	switch key {
	case "stream":
		boolVal := strings.ToLower(value) == "true" || value == "1" || strings.ToLower(value) == "on"
		if err := r.config.SetValue(key, boolVal); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	case "temperature":
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value for %s: %w", key, err)
		}
		if floatVal < 0.0 || floatVal > 2.0 {
			return fmt.Errorf("temperature must be between 0.0 and 2.0")
		}
		r.session.Conversation.Temperature = floatVal
	case "max_tokens":
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value for %s: %w", key, err)
		}
		if intVal < 1 {
			return fmt.Errorf("max_tokens must be positive")
		}
		r.session.Conversation.MaxTokens = intVal
	default:
		// For all other keys, store as string
		if err := r.config.SetValue(key, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}

	fmt.Fprintf(r.writer, "Set %s = %s\n", key, value)
	return nil
}

// exportSession exports the current session to a file or stdout
func (r *REPL) exportSession(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: /export <format> [filename]")
	}

	format := strings.ToLower(args[0])
	if format != "json" && format != "markdown" {
		return fmt.Errorf("unsupported format: %s (valid: json, markdown)", format)
	}

	logging.LogInfo("Exporting session", "sessionID", r.session.ID, "format", format)

	// Determine output destination
	var writer io.Writer
	var filename string
	if len(args) > 1 {
		filename = args[1]
		file, err := os.Create(filename)
		if err != nil {
			logging.LogError(err, "Failed to create export file", "filename", filename)
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer file.Close()
		writer = file
		logging.LogDebug("Exporting to file", "filename", filename)
	} else {
		writer = r.writer
		logging.LogDebug("Exporting to stdout")
	}

	// Export the session
	if err := r.manager.ExportSession(r.session.ID, format, writer); err != nil {
		logging.LogError(err, "Failed to export session", "sessionID", r.session.ID, "format", format)
		return fmt.Errorf("failed to export session: %w", err)
	}

	if filename != "" {
		fmt.Fprintf(r.writer, "Session exported to: %s\n", filename)
		logging.LogInfo("Session exported to file", "sessionID", r.session.ID, "filename", filename, "format", format)
	} else {
		// Don't add extra text when exporting to stdout to avoid breaking the format
		logging.LogInfo("Session exported to stdout", "sessionID", r.session.ID, "format", format)
	}

	return nil
}

// searchSessions searches for sessions by content
func (r *REPL) searchSessions(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: /search <query>")
	}

	query := strings.Join(args, " ")
	logging.LogInfo("Searching sessions", "query", query)

	results, err := r.manager.SearchSessions(query)
	if err != nil {
		logging.LogError(err, "Failed to search sessions", "query", query)
		return fmt.Errorf("failed to search sessions: %w", err)
	}

	if len(results) == 0 {
		fmt.Fprintf(r.writer, "No sessions found matching: %s\n", query)
		return nil
	}

	fmt.Fprintf(r.writer, "Found %d sessions matching '%s':\n\n", len(results), query)
	
	for _, result := range results {
		// Session info
		fmt.Fprintf(r.writer, "Session: %s (%s)\n", result.Session.Name, result.Session.ID)
		fmt.Fprintf(r.writer, "Created: %s\n", result.Session.Created.Format("2006-01-02 15:04:05"))
		
		// Show matches
		for _, match := range result.Matches {
			fmt.Fprintf(r.writer, "  %s: %s\n", match.Context, match.Content)
		}
		fmt.Fprintln(r.writer)
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
