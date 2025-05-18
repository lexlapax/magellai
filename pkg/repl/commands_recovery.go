// ABOUTME: Recovery-related commands for REPL
// ABOUTME: Implements manual recovery and crash recovery features

package repl

import (
	"fmt"
	"strings"

	"github.com/lexlapax/magellai/internal/logging"
)

// cmdRecover manually triggers recovery operations
func (r *REPL) cmdRecover(args []string) error {
	if r.autoRecovery == nil {
		return fmt.Errorf("auto-recovery is not enabled")
	}

	if len(args) == 0 {
		// Show recovery status
		return r.showRecoveryStatus()
	}

	subcommand := args[0]
	switch subcommand {
	case "check":
		return r.checkRecoveryState()
	case "save":
		return r.saveRecoveryState()
	case "clear":
		return r.clearRecoveryState()
	case "restore":
		return r.restoreFromRecovery()
	default:
		return fmt.Errorf("unknown recover subcommand: %s (use check, save, clear, or restore)", subcommand)
	}
}

// showRecoveryStatus displays current recovery status
func (r *REPL) showRecoveryStatus() error {
	if r.autoRecovery == nil {
		fmt.Fprintln(r.writer, "Auto-recovery is not enabled")
		return nil
	}

	// Check for existing recovery state
	state, err := r.autoRecovery.CheckRecovery()
	if err != nil {
		return fmt.Errorf("failed to check recovery state: %w", err)
	}

	if state == nil {
		fmt.Fprintln(r.writer, "No recovery state found")
		return nil
	}

	fmt.Fprintf(r.writer, "Recovery State Found:\n")
	fmt.Fprintf(r.writer, "  Session ID: %s\n", state.SessionID)
	fmt.Fprintf(r.writer, "  Session Name: %s\n", state.SessionName)
	fmt.Fprintf(r.writer, "  Last Saved: %s\n", state.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(r.writer, "  Storage Backend: %s\n", state.StorageBackend)
	
	// Show last save time
	lastSave := r.autoRecovery.GetLastSaveTime()
	if !lastSave.IsZero() {
		fmt.Fprintf(r.writer, "  Current Session Last Saved: %s\n", lastSave.Format("2006-01-02 15:04:05"))
	}

	return nil
}

// checkRecoveryState checks if recovery state exists
func (r *REPL) checkRecoveryState() error {
	state, err := r.autoRecovery.CheckRecovery()
	if err != nil {
		return fmt.Errorf("failed to check recovery state: %w", err)
	}

	if state == nil {
		fmt.Fprintln(r.writer, "No recovery state found")
		return nil
	}

	fmt.Fprintf(r.writer, "Found recovery state for session: %s\n", state.SessionID)
	return nil
}

// saveRecoveryState manually saves the current recovery state
func (r *REPL) saveRecoveryState() error {
	if err := r.autoRecovery.ForceRecoverySave(); err != nil {
		return fmt.Errorf("failed to save recovery state: %w", err)
	}

	fmt.Fprintln(r.writer, "Recovery state saved successfully")
	logging.LogInfo("Manual recovery save completed")
	return nil
}

// clearRecoveryState clears any existing recovery state
func (r *REPL) clearRecoveryState() error {
	// Ask for confirmation
	fmt.Fprint(r.writer, "Are you sure you want to clear the recovery state? (y/n): ")
	response, err := r.reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Fprintln(r.writer, "Cancelled")
		return nil
	}

	if err := r.autoRecovery.ClearRecoveryState(); err != nil {
		return fmt.Errorf("failed to clear recovery state: %w", err)
	}

	fmt.Fprintln(r.writer, "Recovery state cleared")
	logging.LogInfo("Recovery state manually cleared")
	return nil
}

// restoreFromRecovery attempts to restore from recovery state
func (r *REPL) restoreFromRecovery() error {
	state, err := r.autoRecovery.CheckRecovery()
	if err != nil {
		return fmt.Errorf("failed to check recovery state: %w", err)
	}

	if state == nil {
		fmt.Fprintln(r.writer, "No recovery state found")
		return nil
	}

	// Show recovery info
	fmt.Fprintf(r.writer, "Found recovery state:\n")
	fmt.Fprintf(r.writer, "  Session ID: %s\n", state.SessionID)
	fmt.Fprintf(r.writer, "  Session Name: %s\n", state.SessionName)
	fmt.Fprintf(r.writer, "  Last Saved: %s\n", state.Timestamp.Format("2006-01-02 15:04:05"))
	
	// Ask for confirmation
	fmt.Fprint(r.writer, "Restore this session? (y/n): ")
	response, err := r.reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Fprintln(r.writer, "Cancelled")
		return nil
	}

	// Save current session if needed
	if r.session != nil && r.hasUnsavedChanges() {
		fmt.Fprint(r.writer, "Save current session before restoring? (y/n): ")
		saveResponse, err := r.reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		saveResponse = strings.TrimSpace(strings.ToLower(saveResponse))
		if saveResponse == "y" || saveResponse == "yes" {
			if err := r.manager.SaveSession(r.session); err != nil {
				logging.LogWarn("Failed to save current session", "error", err)
			}
		}
	}

	// Restore the session
	session, err := r.autoRecovery.RecoverSession(state)
	if err != nil {
		return fmt.Errorf("failed to recover session: %w", err)
	}

	// Switch to recovered session
	r.session = session
	
	// Clear recovery state after successful recovery
	if err := r.autoRecovery.ClearRecoveryState(); err != nil {
		logging.LogWarn("Failed to clear recovery state after successful recovery", "error", err)
	}

	fmt.Fprintf(r.writer, "Session recovered successfully\n")
	fmt.Fprintf(r.writer, "Session ID: %s\n", session.ID)
	fmt.Fprintf(r.writer, "Messages: %d\n", len(session.Conversation.Messages))
	
	logging.LogInfo("Session manually recovered", "id", session.ID)
	return nil
}