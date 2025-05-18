// ABOUTME: Auto-recovery system for REPL sessions after crashes
// ABOUTME: Periodically saves session state and recovers on restart

package repl

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lexlapax/magellai/internal/configdir"
	"github.com/lexlapax/magellai/internal/logging"
	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/storage"
)

// AutoRecoveryConfig defines configuration for auto-recovery
type AutoRecoveryConfig struct {
	Enabled           bool          `json:"enabled"`
	SaveInterval      time.Duration `json:"save_interval"`
	RecoveryFile      string        `json:"recovery_file"`
	MaxRecoveryAge    time.Duration `json:"max_recovery_age"`
	BackupCount       int           `json:"backup_count"`
	RecoveryDirectory string        `json:"recovery_directory"`
}

// DefaultAutoRecoveryConfig returns the default auto-recovery configuration
func DefaultAutoRecoveryConfig() *AutoRecoveryConfig {
	paths, err := configdir.GetPaths()
	if err != nil {
		// Fallback to a reasonable default
		homeDir, _ := os.UserHomeDir()
		configDir := filepath.Join(homeDir, ".config", "magellai")
		return &AutoRecoveryConfig{
			Enabled:           true,
			SaveInterval:      30 * time.Second,
			RecoveryFile:      "recovery.json",
			MaxRecoveryAge:    24 * time.Hour,
			BackupCount:       3,
			RecoveryDirectory: filepath.Join(configDir, "recovery"),
		}
	}

	return &AutoRecoveryConfig{
		Enabled:           true,
		SaveInterval:      30 * time.Second,
		RecoveryFile:      "recovery.json",
		MaxRecoveryAge:    24 * time.Hour,
		BackupCount:       3,
		RecoveryDirectory: filepath.Join(paths.Base, "recovery"),
	}
}

// AutoRecoveryManager handles automatic session recovery
type AutoRecoveryManager struct {
	config         *AutoRecoveryConfig
	storageManager *StorageManager
	stopChan       chan struct{}
	saveTicker     *time.Ticker
	lastSave       time.Time
	done           chan struct{}
}

// RecoveryState represents the state saved for recovery
type RecoveryState struct {
	SessionID        string              `json:"session_id"`
	SessionName      string              `json:"session_name"`
	ConversationData *domain.Session     `json:"conversation_data,omitempty"`
	Timestamp        time.Time           `json:"timestamp"`
	AppVersion       string              `json:"app_version,omitempty"`
	StorageBackend   storage.BackendType `json:"storage_backend"`
}

// NewAutoRecoveryManager creates a new auto-recovery manager
func NewAutoRecoveryManager(config *AutoRecoveryConfig, storageManager *StorageManager) (*AutoRecoveryManager, error) {
	if config == nil {
		config = DefaultAutoRecoveryConfig()
	}

	// Ensure recovery directory exists
	if err := os.MkdirAll(config.RecoveryDirectory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create recovery directory: %w", err)
	}

	return &AutoRecoveryManager{
		config:         config,
		storageManager: storageManager,
		stopChan:       make(chan struct{}),
		done:           make(chan struct{}),
	}, nil
}

// Start begins the auto-recovery background process
func (arm *AutoRecoveryManager) Start() error {
	if !arm.config.Enabled {
		logging.LogDebug("Auto-recovery is disabled")
		return nil
	}

	arm.saveTicker = time.NewTicker(arm.config.SaveInterval)

	go func() {
		defer close(arm.done)
		for {
			select {
			case <-arm.saveTicker.C:
				if err := arm.SaveRecoveryState(); err != nil {
					logging.LogWarn("Failed to save recovery state", "error", err)
				}
			case <-arm.stopChan:
				arm.saveTicker.Stop()
				return
			}
		}
	}()

	logging.LogInfo("Auto-recovery started", "interval", arm.config.SaveInterval)
	return nil
}

// Stop stops the auto-recovery process
func (arm *AutoRecoveryManager) Stop() {
	if arm.saveTicker != nil {
		close(arm.stopChan)
		<-arm.done // Wait for goroutine to finish
		logging.LogDebug("Auto-recovery stopped")
	}
}

// SaveRecoveryState saves the current session state for recovery
func (arm *AutoRecoveryManager) SaveRecoveryState() error {
	currentSession := arm.storageManager.CurrentSession()
	if currentSession == nil {
		logging.LogDebug("No active session to save for recovery")
		return nil
	}

	state := &RecoveryState{
		SessionID:        currentSession.ID,
		SessionName:      currentSession.Name,
		ConversationData: currentSession,
		Timestamp:        time.Now(),
		StorageBackend:   arm.storageManager.backendType,
	}

	// Rotate backups
	if err := arm.rotateBackups(); err != nil {
		logging.LogWarn("Failed to rotate recovery backups", "error", err)
	}

	// Ensure recovery directory exists
	if err := os.MkdirAll(arm.config.RecoveryDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create recovery directory: %w", err)
	}

	// Save state to file
	recoveryPath := filepath.Join(arm.config.RecoveryDirectory, arm.config.RecoveryFile)
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal recovery state: %w", err)
	}

	if err := os.WriteFile(recoveryPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write recovery file: %w", err)
	}

	arm.lastSave = time.Now()
	logging.LogDebug("Recovery state saved", "session", currentSession.ID)
	return nil
}

// CheckRecovery checks if a recoverable session exists
func (arm *AutoRecoveryManager) CheckRecovery() (*RecoveryState, error) {
	recoveryPath := filepath.Join(arm.config.RecoveryDirectory, arm.config.RecoveryFile)

	// Check if recovery file exists
	if _, err := os.Stat(recoveryPath); os.IsNotExist(err) {
		return nil, nil
	}

	// Read recovery file
	data, err := os.ReadFile(recoveryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read recovery file: %w", err)
	}

	var state RecoveryState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recovery state: %w", err)
	}

	// Check if recovery state is too old
	if time.Since(state.Timestamp) > arm.config.MaxRecoveryAge {
		logging.LogDebug("Recovery state too old", "age", time.Since(state.Timestamp))
		return nil, nil
	}

	return &state, nil
}

// RecoverSession attempts to recover a session from saved state
func (arm *AutoRecoveryManager) RecoverSession(state *RecoveryState) (*domain.Session, error) {
	if state == nil || state.ConversationData == nil {
		return nil, fmt.Errorf("invalid recovery state")
	}

	// Check if backend types match
	if state.StorageBackend != arm.storageManager.backendType {
		logging.LogWarn("Storage backend mismatch in recovery",
			"saved", state.StorageBackend,
			"current", arm.storageManager.backendType)
	}

	// Try to load the session from storage first
	session, err := arm.storageManager.LoadSession(state.SessionID)
	if err == nil {
		logging.LogInfo("Session found in storage", "id", state.SessionID)
		return session, nil
	}

	// If not found in storage, use the recovery data
	logging.LogInfo("Recovering session from crash", "id", state.SessionID)
	recoveredSession := state.ConversationData

	// Save the recovered session to storage
	if err := arm.storageManager.SaveSession(recoveredSession); err != nil {
		logging.LogWarn("Failed to save recovered session", "error", err)
	}

	return recoveredSession, nil
}

// ClearRecoveryState removes the recovery state file
func (arm *AutoRecoveryManager) ClearRecoveryState() error {
	recoveryPath := filepath.Join(arm.config.RecoveryDirectory, arm.config.RecoveryFile)

	if err := os.Remove(recoveryPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove recovery file: %w", err)
	}

	logging.LogDebug("Recovery state cleared")
	return nil
}

// rotateBackups manages recovery file backups
func (arm *AutoRecoveryManager) rotateBackups() error {
	if arm.config.BackupCount <= 0 {
		return nil
	}

	recoveryPath := filepath.Join(arm.config.RecoveryDirectory, arm.config.RecoveryFile)

	// Rotate existing backups
	for i := arm.config.BackupCount - 1; i > 0; i-- {
		oldPath := fmt.Sprintf("%s.%d", recoveryPath, i)
		newPath := fmt.Sprintf("%s.%d", recoveryPath, i+1)

		if _, err := os.Stat(oldPath); err == nil {
			if err := os.Rename(oldPath, newPath); err != nil {
				logging.LogWarn("Failed to rotate backup", "from", oldPath, "to", newPath, "error", err)
			}
		}
	}

	// Move current recovery file to first backup
	firstBackup := fmt.Sprintf("%s.1", recoveryPath)
	if _, err := os.Stat(recoveryPath); err == nil {
		if err := os.Rename(recoveryPath, firstBackup); err != nil {
			logging.LogWarn("Failed to create backup", "error", err)
		}
	}

	return nil
}

// GetLastSaveTime returns the time of the last recovery save
func (arm *AutoRecoveryManager) GetLastSaveTime() time.Time {
	return arm.lastSave
}

// ForceRecoverySave forces an immediate save of the recovery state
func (arm *AutoRecoveryManager) ForceRecoverySave() error {
	return arm.SaveRecoveryState()
}
