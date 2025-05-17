package repl

import (
	"fmt"
	"sync"
)

// StorageType represents the type of storage backend
type StorageType string

const (
	// FileSystemStorage represents file system storage
	FileSystemStorage StorageType = "filesystem"
	// SQLiteStorage represents SQLite database storage (future)
	SQLiteStorage StorageType = "sqlite"
	// PostgreSQLStorage represents PostgreSQL database storage (future)
	PostgreSQLStorage StorageType = "postgresql"
)

// StorageFactory creates storage backends based on configuration
type StorageFactory struct {
	mu       sync.RWMutex
	backends map[StorageType]func(config map[string]interface{}) (StorageBackend, error)
	config   map[string]interface{}
}

// NewStorageFactory creates a new storage factory
func NewStorageFactory() *StorageFactory {
	factory := &StorageFactory{
		backends: make(map[StorageType]func(config map[string]interface{}) (StorageBackend, error)),
		config:   make(map[string]interface{}),
	}

	// Register default filesystem backend
	factory.RegisterBackend(FileSystemStorage, func(config map[string]interface{}) (StorageBackend, error) {
		baseDir, ok := config["base_dir"].(string)
		if !ok {
			baseDir = ""
		}
		return NewFileSystemBackend(baseDir)
	})

	return factory
}

// RegisterBackend registers a new storage backend factory function
func (f *StorageFactory) RegisterBackend(storageType StorageType, factory func(config map[string]interface{}) (StorageBackend, error)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.backends[storageType] = factory
}

// SetConfig sets the configuration for the factory
func (f *StorageFactory) SetConfig(config map[string]interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.config = config
}

// CreateBackend creates a storage backend of the specified type
func (f *StorageFactory) CreateBackend(storageType StorageType) (StorageBackend, error) {
	f.mu.RLock()
	factory, ok := f.backends[storageType]
	config := f.config
	f.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown storage type: %s", storageType)
	}

	return factory(config)
}

// DefaultFactory is the global storage factory instance
var DefaultFactory = NewStorageFactory()

// CreateStorageBackend is a convenience function that uses the default factory
func CreateStorageBackend(storageType StorageType, config map[string]interface{}) (StorageBackend, error) {
	DefaultFactory.SetConfig(config)
	return DefaultFactory.CreateBackend(storageType)
}
