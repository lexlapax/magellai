// ABOUTME: Tests for the auto-recovery system
// ABOUTME: Validates recovery functionality and crash resilience

package repl

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutoRecoveryManager_SaveAndRecover(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()

	// Create test configuration
	config := &AutoRecoveryConfig{
		Enabled:           true,
		SaveInterval:      100 * time.Millisecond,
		RecoveryFile:      "test_recovery.json",
		MaxRecoveryAge:    time.Hour,
		BackupCount:       2,
		RecoveryDirectory: tempDir,
	}

	// Create storage manager
	backend, err := storage.CreateBackend(storage.FileSystemBackend, storage.Config{
		"base_dir": tempDir,
	})
	require.NoError(t, err)

	storageManager, err := NewStorageManager(backend)
	require.NoError(t, err)

	// Create and save a test session
	session := storageManager.NewSession("Test Session")
	// Add a message to the conversation
	message := domain.NewMessage("msg-1", domain.MessageRoleUser, "test message")
	session.Conversation.AddMessage(*message)
	err = storageManager.SaveSession(session)
	require.NoError(t, err)

	// Set as current session
	storageManager.SetCurrentSession(session)

	// Create auto-recovery manager
	arm, err := NewAutoRecoveryManager(config, storageManager)
	require.NoError(t, err)

	// Save recovery state
	err = arm.SaveRecoveryState()
	assert.NoError(t, err)

	// Check recovery
	state, err := arm.CheckRecovery()
	assert.NoError(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, session.ID, state.SessionID)
	assert.Equal(t, session.Name, state.SessionName)

	// Simulate crash by creating new storage manager
	newStorageManager, err := NewStorageManager(backend)
	require.NoError(t, err)

	newARM, err := NewAutoRecoveryManager(config, newStorageManager)
	require.NoError(t, err)

	// Recover session
	recoveredSession, err := newARM.RecoverSession(state)
	assert.NoError(t, err)
	assert.NotNil(t, recoveredSession)
	assert.Equal(t, session.ID, recoveredSession.ID)
	assert.Equal(t, session.Name, recoveredSession.Name)
	assert.Equal(t, len(session.Conversation.Messages), len(recoveredSession.Conversation.Messages))

	// Cleanup
	storageManager.Close()
	newStorageManager.Close()
}

func TestAutoRecoveryManager_AutoSave(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()

	// Create test configuration with short interval
	config := &AutoRecoveryConfig{
		Enabled:           true,
		SaveInterval:      50 * time.Millisecond,
		RecoveryFile:      "test_recovery.json",
		MaxRecoveryAge:    time.Hour,
		BackupCount:       2,
		RecoveryDirectory: tempDir,
	}

	// Create storage manager
	backend, err := storage.CreateBackend(storage.FileSystemBackend, storage.Config{
		"base_dir": tempDir,
	})
	require.NoError(t, err)

	storageManager, err := NewStorageManager(backend)
	require.NoError(t, err)

	// Create test session
	session := storageManager.NewSession("Auto Save Test")
	storageManager.SetCurrentSession(session)

	// Create auto-recovery manager
	arm, err := NewAutoRecoveryManager(config, storageManager)
	require.NoError(t, err)

	// Start auto-recovery
	err = arm.Start()
	assert.NoError(t, err)

	// Wait for auto-save to trigger
	time.Sleep(100 * time.Millisecond)

	// Check that recovery file exists
	recoveryPath := filepath.Join(tempDir, config.RecoveryFile)
	_, err = os.Stat(recoveryPath)
	assert.NoError(t, err)

	// Stop auto-recovery
	arm.Stop()

	// Close the backend to release file handles
	storageManager.Close()
}

func TestAutoRecoveryManager_BackupRotation(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()

	// Create test configuration
	config := &AutoRecoveryConfig{
		Enabled:           true,
		SaveInterval:      time.Hour,
		RecoveryFile:      "test_recovery.json",
		MaxRecoveryAge:    time.Hour,
		BackupCount:       2,
		RecoveryDirectory: tempDir,
	}

	// Create storage manager
	backend, err := storage.CreateBackend(storage.FileSystemBackend, storage.Config{
		"base_dir": tempDir,
	})
	require.NoError(t, err)

	storageManager, err := NewStorageManager(backend)
	require.NoError(t, err)

	// Create test session
	session := storageManager.NewSession("Backup Test")
	storageManager.SetCurrentSession(session)

	// Create auto-recovery manager
	arm, err := NewAutoRecoveryManager(config, storageManager)
	require.NoError(t, err)

	// Save multiple times to test rotation
	for i := 0; i < 4; i++ {
		message := domain.NewMessage(fmt.Sprintf("msg-%d", i+2), domain.MessageRoleUser, fmt.Sprintf("message %d", i))
		session.Conversation.AddMessage(*message)
		err = arm.SaveRecoveryState()
		assert.NoError(t, err)
	}

	// Check backup files exist
	recoveryPath := filepath.Join(tempDir, config.RecoveryFile)

	// Current file should exist
	_, err = os.Stat(recoveryPath)
	assert.NoError(t, err)

	// Backup 1 should exist
	_, err = os.Stat(recoveryPath + ".1")
	assert.NoError(t, err)

	// Backup 2 should exist
	_, err = os.Stat(recoveryPath + ".2")
	assert.NoError(t, err)

	// Backup 3 should not exist (we only keep 2 backups)
	_, err = os.Stat(recoveryPath + ".3")
	assert.True(t, os.IsNotExist(err))

	// Cleanup
	storageManager.Close()
}

func TestAutoRecoveryManager_AgeCheck(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()

	// Create test configuration with short max age
	config := &AutoRecoveryConfig{
		Enabled:           true,
		SaveInterval:      time.Hour,
		RecoveryFile:      "test_recovery.json",
		MaxRecoveryAge:    100 * time.Millisecond,
		BackupCount:       0,
		RecoveryDirectory: tempDir,
	}

	// Create storage manager
	backend, err := storage.CreateBackend(storage.FileSystemBackend, storage.Config{
		"base_dir": tempDir,
	})
	require.NoError(t, err)

	storageManager, err := NewStorageManager(backend)
	require.NoError(t, err)

	// Create test session
	session := storageManager.NewSession("Age Test")
	storageManager.SetCurrentSession(session)

	// Create auto-recovery manager
	arm, err := NewAutoRecoveryManager(config, storageManager)
	require.NoError(t, err)

	// Save recovery state
	err = arm.SaveRecoveryState()
	assert.NoError(t, err)

	// Check immediately - should find recovery
	state, err := arm.CheckRecovery()
	assert.NoError(t, err)
	assert.NotNil(t, state)

	// Wait for max age to pass
	time.Sleep(150 * time.Millisecond)

	// Check again - should not find recovery (too old)
	state, err = arm.CheckRecovery()
	assert.NoError(t, err)
	assert.Nil(t, state)

	// Cleanup
	storageManager.Close()
}

func TestAutoRecoveryManager_ClearRecovery(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()

	// Create test configuration
	config := &AutoRecoveryConfig{
		Enabled:           true,
		SaveInterval:      time.Hour,
		RecoveryFile:      "test_recovery.json",
		MaxRecoveryAge:    time.Hour,
		BackupCount:       0,
		RecoveryDirectory: tempDir,
	}

	// Create storage manager
	backend, err := storage.CreateBackend(storage.FileSystemBackend, storage.Config{
		"base_dir": tempDir,
	})
	require.NoError(t, err)

	storageManager, err := NewStorageManager(backend)
	require.NoError(t, err)

	// Create test session
	session := storageManager.NewSession("Clear Test")
	storageManager.SetCurrentSession(session)

	// Create auto-recovery manager
	arm, err := NewAutoRecoveryManager(config, storageManager)
	require.NoError(t, err)

	// Save recovery state
	err = arm.SaveRecoveryState()
	assert.NoError(t, err)

	// Verify file exists
	recoveryPath := filepath.Join(tempDir, config.RecoveryFile)
	_, err = os.Stat(recoveryPath)
	assert.NoError(t, err)

	// Clear recovery state
	err = arm.ClearRecoveryState()
	assert.NoError(t, err)

	// Verify file is gone
	_, err = os.Stat(recoveryPath)
	assert.True(t, os.IsNotExist(err))

	// Cleanup
	storageManager.Close()
}

func TestAutoRecoveryManager_DisabledRecovery(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()

	// Create test configuration with recovery disabled
	config := &AutoRecoveryConfig{
		Enabled:           false,
		SaveInterval:      time.Hour,
		RecoveryFile:      "test_recovery.json",
		MaxRecoveryAge:    time.Hour,
		BackupCount:       0,
		RecoveryDirectory: tempDir,
	}

	// Create storage manager
	backend, err := storage.CreateBackend(storage.FileSystemBackend, storage.Config{
		"base_dir": tempDir,
	})
	require.NoError(t, err)

	storageManager, err := NewStorageManager(backend)
	require.NoError(t, err)

	// Create auto-recovery manager
	arm, err := NewAutoRecoveryManager(config, storageManager)
	require.NoError(t, err)

	// Start should not error but should not start ticker
	err = arm.Start()
	assert.NoError(t, err)
	assert.Nil(t, arm.saveTicker)

	// Cleanup
	storageManager.Close()
}

func TestRecoveryState_Serialization(t *testing.T) {
	// Create test recovery state
	session := &domain.Session{
		ID:   "test-session-id",
		Name: "Test Session",
		Conversation: &domain.Conversation{
			Messages: []domain.Message{
				{
					ID:      "msg-1",
					Role:    domain.MessageRoleUser,
					Content: "Hello",
				},
			},
		},
	}

	state := &RecoveryState{
		SessionID:        session.ID,
		SessionName:      session.Name,
		ConversationData: session,
		Timestamp:        time.Now(),
		StorageBackend:   storage.FileSystemBackend,
	}

	// Serialize to JSON
	data, err := json.Marshal(state)
	assert.NoError(t, err)

	// Deserialize back
	var recovered RecoveryState
	err = json.Unmarshal(data, &recovered)
	assert.NoError(t, err)

	// Verify fields
	assert.Equal(t, state.SessionID, recovered.SessionID)
	assert.Equal(t, state.SessionName, recovered.SessionName)
	assert.Equal(t, state.StorageBackend, recovered.StorageBackend)
	assert.Equal(t, len(state.ConversationData.Conversation.Messages),
		len(recovered.ConversationData.Conversation.Messages))
}
