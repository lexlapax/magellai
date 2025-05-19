// ABOUTME: Extended tests for auto-recovery system edge cases
// ABOUTME: Tests error conditions, concurrency, and failure scenarios

package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/storage"
	_ "github.com/lexlapax/magellai/pkg/storage/filesystem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutoRecoveryManager_ConcurrentAccess(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()

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
	defer storageManager.Close()

	// Create test session
	session := storageManager.NewSession("Concurrent Test")
	storageManager.SetCurrentSession(session)

	// Create auto-recovery manager
	arm, err := NewAutoRecoveryManager(config, storageManager)
	require.NoError(t, err)

	// Start auto-recovery
	err = arm.Start()
	require.NoError(t, err)

	// Concurrently modify session and force saves
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			message := domain.NewMessage(
				fmt.Sprintf("msg-%d", idx),
				domain.MessageRoleUser,
				fmt.Sprintf("concurrent message %d", idx),
			)
			session.Conversation.AddMessage(*message)
			err := arm.ForceRecoverySave()
			assert.NoError(t, err)
		}(i)
	}

	// Wait for all operations to complete
	wg.Wait()

	// Stop auto-recovery
	arm.Stop()

	// Verify recovery state
	state, err := arm.CheckRecovery()
	assert.NoError(t, err)
	assert.NotNil(t, state)
}

func TestAutoRecoveryManager_CorruptedRecoveryFile(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()

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
	defer storageManager.Close()

	// Create auto-recovery manager
	arm, err := NewAutoRecoveryManager(config, storageManager)
	require.NoError(t, err)

	// Write corrupted recovery file
	recoveryPath := filepath.Join(tempDir, config.RecoveryFile)
	err = os.WriteFile(recoveryPath, []byte("invalid json"), 0600)
	require.NoError(t, err)

	// Check recovery should fail
	state, err := arm.CheckRecovery()
	assert.Error(t, err)
	assert.Nil(t, state)
	assert.Contains(t, err.Error(), "failed to unmarshal recovery state")
}

func TestAutoRecoveryManager_MissingRecoveryDirectory(t *testing.T) {
	// Create a directory that we'll make unwritable
	tempDir := t.TempDir()
	recoveryDir := filepath.Join(tempDir, "recovery")

	config := &AutoRecoveryConfig{
		Enabled:           true,
		SaveInterval:      time.Hour,
		RecoveryFile:      "test_recovery.json",
		MaxRecoveryAge:    time.Hour,
		BackupCount:       0,
		RecoveryDirectory: recoveryDir,
	}

	// Create storage manager
	backend, err := storage.CreateBackend(storage.FileSystemBackend, storage.Config{
		"base_dir": tempDir,
	})
	require.NoError(t, err)

	storageManager, err := NewStorageManager(backend)
	require.NoError(t, err)
	defer storageManager.Close()

	// Create auto-recovery manager - should create directory
	arm, err := NewAutoRecoveryManager(config, storageManager)
	assert.NoError(t, err)
	assert.NotNil(t, arm)

	// Verify directory was created
	_, err = os.Stat(recoveryDir)
	assert.NoError(t, err)
}

func TestAutoRecoveryManager_InvalidSession(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()

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
	defer storageManager.Close()

	// Create auto-recovery manager
	arm, err := NewAutoRecoveryManager(config, storageManager)
	require.NoError(t, err)

	// Ensure no current session is set
	storageManager.SetCurrentSession(nil)

	// Try to save recovery state without active session
	err = arm.SaveRecoveryState()
	assert.NoError(t, err) // Should not error, just log debug

	// When saving recovery state with no active session,
	// the directory might be created but no recovery file should exist
	recoveryPath := filepath.Join(tempDir, config.RecoveryFile)
	_, err = os.Stat(recoveryPath)
	assert.True(t, os.IsNotExist(err))
}

func TestAutoRecoveryManager_RecoverWithNilState(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()

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
	defer storageManager.Close()

	// Create auto-recovery manager
	arm, err := NewAutoRecoveryManager(config, storageManager)
	require.NoError(t, err)

	// Try to recover with nil state
	session, err := arm.RecoverSession(nil)
	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), "invalid recovery state")

	// Try to recover with state but nil conversation data
	state := &RecoveryState{
		SessionID:        "test-id",
		SessionName:      "test-session",
		ConversationData: nil,
		Timestamp:        time.Now(),
	}
	session, err = arm.RecoverSession(state)
	assert.Error(t, err)
	assert.Nil(t, session)
}

func TestAutoRecoveryManager_BackendMismatch(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()

	config := &AutoRecoveryConfig{
		Enabled:           true,
		SaveInterval:      time.Hour,
		RecoveryFile:      "test_recovery.json",
		MaxRecoveryAge:    time.Hour,
		BackupCount:       0,
		RecoveryDirectory: tempDir,
	}

	// Create storage manager with filesystem backend
	backend, err := storage.CreateBackend(storage.FileSystemBackend, storage.Config{
		"base_dir": tempDir,
	})
	require.NoError(t, err)

	storageManager, err := NewStorageManager(backend)
	require.NoError(t, err)
	defer storageManager.Close()

	// Create test session
	session := &domain.Session{
		ID:   "test-session-id",
		Name: "Test Session",
		Conversation: &domain.Conversation{
			Messages: []domain.Message{},
		},
	}

	// Create recovery state with different backend type
	state := &RecoveryState{
		SessionID:        session.ID,
		SessionName:      session.Name,
		ConversationData: session,
		Timestamp:        time.Now(),
		StorageBackend:   storage.SQLiteBackend, // Different backend
	}

	// Create auto-recovery manager
	arm, err := NewAutoRecoveryManager(config, storageManager)
	require.NoError(t, err)

	// Recover session - should warn about mismatch but still recover
	recoveredSession, err := arm.RecoverSession(state)
	assert.NoError(t, err)
	assert.NotNil(t, recoveredSession)
	assert.Equal(t, session.ID, recoveredSession.ID)
}

func TestAutoRecoveryManager_ForceRecoverySave(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()

	config := &AutoRecoveryConfig{
		Enabled:           true,
		SaveInterval:      time.Hour, // Long interval
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
	defer storageManager.Close()

	// Create test session
	session := storageManager.NewSession("Force Save Test")
	storageManager.SetCurrentSession(session)

	// Create auto-recovery manager
	arm, err := NewAutoRecoveryManager(config, storageManager)
	require.NoError(t, err)

	// Force save
	err = arm.ForceRecoverySave()
	assert.NoError(t, err)

	// Verify save time
	lastSave := arm.GetLastSaveTime()
	assert.WithinDuration(t, time.Now(), lastSave, 1*time.Second)

	// Verify recovery file exists
	recoveryPath := filepath.Join(tempDir, config.RecoveryFile)
	_, err = os.Stat(recoveryPath)
	assert.NoError(t, err)
}

func TestAutoRecoveryManager_DefaultConfig(t *testing.T) {
	config := DefaultAutoRecoveryConfig()
	assert.NotNil(t, config)
	assert.True(t, config.Enabled)
	assert.Equal(t, 30*time.Second, config.SaveInterval)
	assert.Equal(t, "recovery.json", config.RecoveryFile)
	assert.Equal(t, 24*time.Hour, config.MaxRecoveryAge)
	assert.Equal(t, 3, config.BackupCount)
	assert.NotEmpty(t, config.RecoveryDirectory)
}

func TestAutoRecoveryManager_NilConfig(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()

	// Create storage manager
	backend, err := storage.CreateBackend(storage.FileSystemBackend, storage.Config{
		"base_dir": tempDir,
	})
	require.NoError(t, err)

	storageManager, err := NewStorageManager(backend)
	require.NoError(t, err)
	defer storageManager.Close()

	// Create auto-recovery manager with nil config - should use defaults
	arm, err := NewAutoRecoveryManager(nil, storageManager)
	assert.NoError(t, err)
	assert.NotNil(t, arm)
	assert.NotNil(t, arm.config)
	assert.True(t, arm.config.Enabled)
}

func TestAutoRecoveryManager_StopWithoutStart(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()

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
	defer storageManager.Close()

	// Create auto-recovery manager
	arm, err := NewAutoRecoveryManager(config, storageManager)
	require.NoError(t, err)

	// Stop without starting - should not panic
	assert.NotPanics(t, func() {
		arm.Stop()
	})
}

func TestAutoRecoveryManager_RotateBackupsError(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()

	config := &AutoRecoveryConfig{
		Enabled:           true,
		SaveInterval:      time.Hour,
		RecoveryFile:      "test_recovery.json",
		MaxRecoveryAge:    time.Hour,
		BackupCount:       -1, // Invalid backup count
		RecoveryDirectory: tempDir,
	}

	// Create storage manager
	backend, err := storage.CreateBackend(storage.FileSystemBackend, storage.Config{
		"base_dir": tempDir,
	})
	require.NoError(t, err)

	storageManager, err := NewStorageManager(backend)
	require.NoError(t, err)
	defer storageManager.Close()

	// Create test session
	session := storageManager.NewSession("Rotation Test")
	storageManager.SetCurrentSession(session)

	// Create auto-recovery manager
	arm, err := NewAutoRecoveryManager(config, storageManager)
	require.NoError(t, err)

	// Save recovery state - rotation should be skipped
	err = arm.SaveRecoveryState()
	assert.NoError(t, err)

	// Verify recovery file exists
	recoveryPath := filepath.Join(tempDir, config.RecoveryFile)
	_, err = os.Stat(recoveryPath)
	assert.NoError(t, err)

	// No backups should exist
	_, err = os.Stat(recoveryPath + ".1")
	assert.True(t, os.IsNotExist(err))
}

func TestAutoRecoveryManager_RecoveryDirectoryCreationError(t *testing.T) {
	// Create a file where the recovery directory should be
	tempDir := t.TempDir()
	blockerFile := filepath.Join(tempDir, "recovery")
	err := os.WriteFile(blockerFile, []byte("blocker"), 0600)
	require.NoError(t, err)

	config := &AutoRecoveryConfig{
		Enabled:           true,
		SaveInterval:      time.Hour,
		RecoveryFile:      "test_recovery.json",
		MaxRecoveryAge:    time.Hour,
		BackupCount:       0,
		RecoveryDirectory: blockerFile, // This is a file, not a directory
	}

	// Create storage manager
	backend, err := storage.CreateBackend(storage.FileSystemBackend, storage.Config{
		"base_dir": tempDir,
	})
	require.NoError(t, err)

	storageManager, err := NewStorageManager(backend)
	require.NoError(t, err)
	defer storageManager.Close()

	// Creating auto-recovery manager should fail
	arm, err := NewAutoRecoveryManager(config, storageManager)
	assert.Error(t, err)
	assert.Nil(t, arm)
	assert.Contains(t, err.Error(), "failed to create recovery directory")
}

func TestRecoveryState_EmptySessionData(t *testing.T) {
	// Create recovery state with empty session
	state := &RecoveryState{
		SessionID:        "",
		SessionName:      "",
		ConversationData: nil,
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

	// Verify nil conversation data
	assert.Nil(t, recovered.ConversationData)
	assert.Empty(t, recovered.SessionID)
	assert.Empty(t, recovered.SessionName)
}
