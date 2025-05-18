// ABOUTME: Implements REPL command handlers for both slash and colon commands
// ABOUTME: Provides functionality for session management, model switching, and chat configuration

package repl

import (
	"fmt"
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
	fmt.Fprintf(r.writer, "Session loaded: %s\n", sessionID)
	return nil
}

// resetConversation clears the conversation history
func (r *REPL) resetConversation() error {
	r.session.Conversation.Messages = []Message{}
	fmt.Fprintln(r.writer, "Conversation reset.")
	return nil
}

// listSessions shows all available sessions
func (r *REPL) listSessions() error {
	sessions, err := r.manager.ListSessions()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Fprintln(r.writer, "No sessions found.")
		return nil
	}

	fmt.Fprintln(r.writer, "Available sessions:")
	for _, s := range sessions {
		status := ""
		if s.ID == r.session.ID {
			status = " (current)"
		}
		fmt.Fprintf(r.writer, "  %s - %s (messages: %d)%s\n", s.ID, s.Name, s.MessageCount, status)
	}
	return nil
}

// showModel shows the current model
func (r *REPL) showModel() error {
	fmt.Fprintf(r.writer, "Current model: %s\n", r.session.Conversation.Model)
	return nil
}

// switchModel switches to a different model
func (r *REPL) switchModel(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("model name required")
	}

	modelName := args[0]
	
	// Try to parse the model name
	parts := strings.Split(modelName, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid model format, expected provider/model (e.g., openai/gpt-4)")
	}

	// Create a new provider with the specified model
	provider, err := llm.NewProvider(parts[0], parts[1])
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Update the REPL provider
	r.provider = provider
	r.session.Conversation.Model = modelName
	r.session.Conversation.Provider = parts[0]

	fmt.Fprintf(r.writer, "Switched to model: %s\n", modelName)
	return nil
}

// toggleStreaming toggles streaming mode
func (r *REPL) toggleStreaming(args []string) error {
	if len(args) == 0 {
		// Toggle current state
		current := r.config.GetBool("stream")
		if err := r.config.SetValue("stream", !current); err != nil {
			logging.LogWarn("Failed to set stream config", "error", err)
		}
		fmt.Fprintf(r.writer, "Streaming mode: %v\n", !current)
		return nil
	}

	switch strings.ToLower(args[0]) {
	case "on", "true", "yes":
		if err := r.config.SetValue("stream", true); err != nil {
			logging.LogWarn("Failed to set stream config", "error", err)
		}
		fmt.Fprintln(r.writer, "Streaming mode: on")
	case "off", "false", "no":
		if err := r.config.SetValue("stream", false); err != nil {
			logging.LogWarn("Failed to set stream config", "error", err)
		}
		fmt.Fprintln(r.writer, "Streaming mode: off")
	default:
		return fmt.Errorf("invalid value: %s (use on/off)", args[0])
	}
	return nil
}

// setVerbosity sets the logging verbosity level
func (r *REPL) setVerbosity(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("verbosity level required (debug, info, warn, error)")
	}

	level := args[0]
	switch strings.ToLower(level) {
	case "debug":
		if err := r.config.SetValue("verbosity", "debug"); err != nil {
			logging.LogWarn("Failed to set verbosity config", "error", err)
		}
		if err := logging.SetLogLevel("debug"); err != nil {
			logging.LogWarn("Failed to set log level", "error", err)
		}
	case "info":
		if err := r.config.SetValue("verbosity", "info"); err != nil {
			logging.LogWarn("Failed to set verbosity config", "error", err)
		}
		if err := logging.SetLogLevel("info"); err != nil {
			logging.LogWarn("Failed to set log level", "error", err)
		}
	case "warn", "warning":
		if err := r.config.SetValue("verbosity", "warn"); err != nil {
			logging.LogWarn("Failed to set verbosity config", "error", err)
		}
		if err := logging.SetLogLevel("warn"); err != nil {
			logging.LogWarn("Failed to set log level", "error", err)
		}
	case "error":
		if err := r.config.SetValue("verbosity", "error"); err != nil {
			logging.LogWarn("Failed to set verbosity config", "error", err)
		}
		if err := logging.SetLogLevel("error"); err != nil {
			logging.LogWarn("Failed to set log level", "error", err)
		}
	default:
		return fmt.Errorf("invalid verbosity level: %s", level)
	}

	fmt.Fprintf(r.writer, "Verbosity level set to: %s\n", level)
	return nil
}

// setOutputFormat sets the output format
func (r *REPL) setOutputFormat(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("output format required (text, json, yaml, markdown)")
	}

	format := strings.ToLower(args[0])
	switch format {
	case "text", "json", "yaml", "markdown":
		if err := r.config.SetValue("output", format); err != nil {
			logging.LogWarn("Failed to set output config", "error", err)
		}
		fmt.Fprintf(r.writer, "Output format set to: %s\n", format)
	default:
		return fmt.Errorf("invalid output format: %s", format)
	}
	return nil
}

// setTemperature sets the generation temperature
func (r *REPL) setTemperature(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("temperature value required (0.0-2.0)")
	}

	temp, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("invalid temperature value: %s", args[0])
	}

	if temp < 0 || temp > 2 {
		return fmt.Errorf("temperature must be between 0.0 and 2.0")
	}

	r.session.Conversation.Temperature = temp
	fmt.Fprintf(r.writer, "Temperature set to: %.1f\n", temp)
	return nil
}

// setMaxTokens sets the maximum response tokens
func (r *REPL) setMaxTokens(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("max tokens value required")
	}

	tokens, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid max tokens value: %s", args[0])
	}

	if tokens < 1 {
		return fmt.Errorf("max tokens must be positive")
	}

	r.session.Conversation.MaxTokens = tokens
	fmt.Fprintf(r.writer, "Max tokens set to: %d\n", tokens)
	return nil
}

// switchProfile switches to a different configuration profile
func (r *REPL) switchProfile(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("profile name required")
	}

	profileName := args[0]
	// Get the profile configuration
	profileKey := fmt.Sprintf("profiles.%s", profileName)
	if !r.config.Exists(profileKey) {
		return fmt.Errorf("profile not found: %s", profileName)
	}

	// Since we don't have Sub method, we'll need to handle profile differently
	// This is a simplified version - in a real implementation you'd need to
	// iterate through the profile's configuration
	// For now, just log that we're switching profiles
	fmt.Fprintf(r.writer, "Profile switching not fully implemented yet.\n")

	// For now, we'll skip the model switching from profile
	// In a real implementation, you'd need to access the profile's model configuration

	fmt.Fprintf(r.writer, "Switched to profile: %s\n", profileName)
	return nil
}

// attachFile adds a file attachment to the next message
func (r *REPL) attachFile(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("file path required")
	}

	filePath := strings.Join(args, " ")
	logging.LogDebug("Attaching file", "path", filePath)
	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		logging.LogError(err, "File does not exist", "path", filePath)
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Create attachment
	attachment, err := createFileAttachmentFromPath(filePath)
	if err != nil {
		logging.LogError(err, "Failed to create attachment", "path", filePath)
		return fmt.Errorf("failed to create attachment: %w", err)
	}
	logging.LogDebug("Created attachment", "type", attachment.Type, "mimeType", attachment.MimeType, "filePath", attachment.FilePath)

	// Store pending attachments in the session metadata
	if r.session.Metadata == nil {
		r.session.Metadata = make(map[string]interface{})
	}

	pendingAttachments, ok := r.session.Metadata["pending_attachments"].([]llm.Attachment)
	if !ok {
		pendingAttachments = []llm.Attachment{}
	}

	pendingAttachments = append(pendingAttachments, attachment)
	r.session.Metadata["pending_attachments"] = pendingAttachments

	fmt.Fprintf(r.writer, "File attached: %s\n", filePath)
	logging.LogInfo("File attached", "path", filePath, "pendingCount", len(pendingAttachments))
	return nil
}

// removeAttachment removes a pending attachment
func (r *REPL) removeAttachment(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("attachment file name required")
	}

	fileName := strings.Join(args, " ")

	// Get pending attachments from session metadata
	if r.session.Metadata == nil {
		fmt.Fprintln(r.writer, "No attachments to remove.")
		return nil
	}

	pendingAttachments, ok := r.session.Metadata["pending_attachments"].([]llm.Attachment)
	if !ok || len(pendingAttachments) == 0 {
		fmt.Fprintln(r.writer, "No attachments to remove.")
		return nil
	}

	// Find and remove the attachment
	var newAttachments []llm.Attachment
	found := false
	for _, att := range pendingAttachments {
		name := getAttachmentDisplayName(att)
		if name == fileName || (att.FilePath != "" && filepath.Base(att.FilePath) == fileName) {
			found = true
			continue
		}
		newAttachments = append(newAttachments, att)
	}

	if !found {
		return fmt.Errorf("attachment not found: %s", fileName)
	}

	r.session.Metadata["pending_attachments"] = newAttachments
	fmt.Fprintf(r.writer, "Attachment removed: %s\n", fileName)
	return nil
}

// listAttachments shows all pending attachments
func (r *REPL) listAttachments() error {
	if r.session.Metadata == nil {
		fmt.Fprintln(r.writer, "No attachments.")
		return nil
	}

	pendingAttachments, ok := r.session.Metadata["pending_attachments"].([]llm.Attachment)
	if !ok || len(pendingAttachments) == 0 {
		fmt.Fprintln(r.writer, "No attachments.")
		return nil
	}

	fmt.Fprintln(r.writer, "Pending attachments:")
	for i, att := range pendingAttachments {
		name := getAttachmentDisplayName(att)
		if att.MimeType != "" {
			fmt.Fprintf(r.writer, "  %d. %s (%s)\n", i+1, name, att.MimeType)
		} else {
			fmt.Fprintf(r.writer, "  %d. %s (%s)\n", i+1, name, att.Type)
		}
	}
	return nil
}

// toggleMultiline toggles multi-line input mode
func (r *REPL) toggleMultiline() error {
	r.multiline = !r.multiline
	if r.multiline {
		fmt.Fprintln(r.writer, "Multi-line mode: on (empty line to submit)")
	} else {
		fmt.Fprintln(r.writer, "Multi-line mode: off")
	}
	return nil
}

// setSystemPrompt sets the system prompt for the conversation
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
	r.session.Conversation.SystemPrompt = prompt
	fmt.Fprintln(r.writer, "System prompt updated.")
	return nil
}

// showHistory displays the conversation history
func (r *REPL) showHistory() error {
	if len(r.session.Conversation.Messages) == 0 {
		fmt.Fprintln(r.writer, "No conversation history.")
		return nil
	}

	fmt.Fprintln(r.writer, "Conversation history:")
	for i, msg := range r.session.Conversation.Messages {
		role := title(string(msg.Role))
		fmt.Fprintf(r.writer, "\n%d. %s:\n%s\n", i+1, role, msg.Content)
		
		if len(msg.Attachments) > 0 {
			fmt.Fprintln(r.writer, "Attachments:")
			for _, att := range msg.Attachments {
				name := getDomainAttachmentDisplayName(att)
				if att.MimeType != "" {
					fmt.Fprintf(r.writer, "  - %s (%s)\n", name, att.MimeType)
				} else {
					fmt.Fprintf(r.writer, "  - %s (%s)\n", name, att.Type)
				}
			}
		}
	}
	return nil
}

// showConfig displays the current configuration
func (r *REPL) showConfig() error {
	fmt.Fprintln(r.writer, "Current configuration:")
	
	// Show relevant config values
	fmt.Fprintf(r.writer, "  Model: %s\n", r.session.Conversation.Model)
	fmt.Fprintf(r.writer, "  Stream: %v\n", r.config.GetBool("stream"))
	fmt.Fprintf(r.writer, "  Temperature: %.1f\n", r.session.Conversation.Temperature)
	fmt.Fprintf(r.writer, "  Max tokens: %d\n", r.session.Conversation.MaxTokens)
	fmt.Fprintf(r.writer, "  Verbosity: %s\n", r.config.GetString("verbosity"))
	fmt.Fprintf(r.writer, "  Auto-save: %v\n", r.autoSave)
	
	return nil
}

// setConfig sets a configuration value
func (r *REPL) setConfig(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: /config set <key> <value>")
	}

	key := args[0]
	value := strings.Join(args[1:], " ")

	// Handle special cases
	switch key {
	case "model":
		return r.switchModel([]string{value})
	case "stream":
		return r.toggleStreaming([]string{value})
	case "temperature":
		return r.setTemperature([]string{value})
	case "max_tokens":
		return r.setMaxTokens([]string{value})
	case "verbosity":
		return r.setVerbosity([]string{value})
	case "auto_save":
		switch strings.ToLower(value) {
		case "on", "true", "yes":
			r.autoSave = true
			fmt.Fprintln(r.writer, "Auto-save enabled")
		case "off", "false", "no":
			r.autoSave = false
			fmt.Fprintln(r.writer, "Auto-save disabled")
		default:
			return fmt.Errorf("invalid value for auto_save: %s", value)
		}
		return nil
	default:
		// Set generic config value
		if err := r.config.SetValue(key, value); err != nil {
			logging.LogWarn("Failed to set config value", "key", key, "error", err)
		}
		fmt.Fprintf(r.writer, "Config %s set to: %s\n", key, value)
	}

	return nil
}

// exportSession exports the current session
func (r *REPL) exportSession(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("export format required: json or markdown")
	}

	format := strings.ToLower(args[0])
	var filename string
	
	if len(args) > 1 {
		filename = args[1]
	} else {
		// Generate default filename
		timestamp := r.session.Created.Format("20060102-150405")
		ext := format
		if format == "markdown" {
			ext = "md"
		}
		filename = fmt.Sprintf("session_%s.%s", timestamp, ext)
	}

	// Export to file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := r.manager.ExportSession(r.session.ID, format, file); err != nil {
		return fmt.Errorf("failed to export session: %w", err)
	}

	fmt.Fprintf(r.writer, "Session exported to: %s\n", filename)
	return nil
}

// searchSessions searches for sessions containing the query
func (r *REPL) searchSessions(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("search query required")
	}

	query := strings.Join(args, " ")
	results, err := r.manager.SearchSessions(query)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Fprintln(r.writer, "No sessions found matching query.")
		return nil
	}

	fmt.Fprintf(r.writer, "Found %d sessions:\n", len(results))
	for _, result := range results {
		fmt.Fprintf(r.writer, "\n%s - %s\n", result.Session.ID, result.Session.Name)
		fmt.Fprintf(r.writer, "  Matches: %d\n", result.GetMatchCount())
		
		// Show a sample of matches
		for _, match := range result.Matches {
			if match.Type == "message" {
				fmt.Fprintf(r.writer, "  - %s message: %s\n", match.Role, match.Context)
			} else {
				fmt.Fprintf(r.writer, "  - %s: %s\n", match.Type, match.Context)
			}
			if len(result.Matches) > 3 {
				fmt.Fprintln(r.writer, "  ...")
				break
			}
		}
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

// listTags lists all tags for the current session
func (r *REPL) listTags() error {
	if len(r.session.Tags) == 0 {
		fmt.Fprintln(r.writer, "No tags assigned to this session.")
		return nil
	}

	fmt.Fprintln(r.writer, "Tags:")
	for _, tag := range r.session.Tags {
		fmt.Fprintf(r.writer, "  - %s\n", tag)
	}
	return nil
}

// addTag adds a tag to the current session
func (r *REPL) addTag(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: /tag <tag>")
	}

	tag := strings.Join(args, " ")
	r.session.AddTag(tag)
	
	fmt.Fprintf(r.writer, "Tag '%s' added to session.\n", tag)
	
	// Auto-save if enabled
	if r.autoSave {
		if err := r.performAutoSave(); err != nil {
			fmt.Fprintf(r.writer, "Warning: Failed to auto-save after adding tag: %v\n", err)
		}
	}
	
	return nil
}

// removeTag removes a tag from the current session
func (r *REPL) removeTag(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: /untag <tag>")
	}

	tag := strings.Join(args, " ")
	r.session.RemoveTag(tag)
	
	fmt.Fprintf(r.writer, "Tag '%s' removed from session.\n", tag)
	
	// Auto-save if enabled
	if r.autoSave {
		if err := r.performAutoSave(); err != nil {
			fmt.Fprintf(r.writer, "Warning: Failed to auto-save after removing tag: %v\n", err)
		}
	}
	
	return nil
}

// showMetadata displays the session's metadata
func (r *REPL) showMetadata() error {
	if len(r.session.Metadata) == 0 {
		fmt.Fprintln(r.writer, "No metadata set for this session.")
		return nil
	}

	fmt.Fprintln(r.writer, "Metadata:")
	for key, value := range r.session.Metadata {
		// Skip internal metadata
		if key == "pending_attachments" {
			continue
		}
		fmt.Fprintf(r.writer, "  %s: %v\n", key, value)
	}
	return nil
}

// setMetadata sets a metadata value for the session
func (r *REPL) setMetadata(key, value string) error {
	if r.session.Metadata == nil {
		r.session.Metadata = make(map[string]interface{})
	}

	r.session.Metadata[key] = value
	r.session.UpdateTimestamp()
	
	fmt.Fprintf(r.writer, "Metadata '%s' set to '%s'.\n", key, value)
	
	// Auto-save if enabled
	if r.autoSave {
		if err := r.performAutoSave(); err != nil {
			fmt.Fprintf(r.writer, "Warning: Failed to auto-save after setting metadata: %v\n", err)
		}
	}
	
	return nil
}

// deleteMetadata removes a metadata key from the session
func (r *REPL) deleteMetadata(key string) error {
	if r.session.Metadata == nil {
		fmt.Fprintln(r.writer, "No metadata to delete.")
		return nil
	}

	if key == "pending_attachments" {
		return fmt.Errorf("cannot delete internal metadata key: %s", key)
	}

	delete(r.session.Metadata, key)
	r.session.UpdateTimestamp()
	
	fmt.Fprintf(r.writer, "Metadata key '%s' deleted.\n", key)
	
	// Auto-save if enabled
	if r.autoSave {
		if err := r.performAutoSave(); err != nil {
			fmt.Fprintf(r.writer, "Warning: Failed to auto-save after deleting metadata: %v\n", err)
		}
	}
	
	return nil
}