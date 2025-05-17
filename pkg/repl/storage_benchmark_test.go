// +build sqlite db

// ABOUTME: Benchmarks comparing database vs filesystem storage performance
// ABOUTME: Measures session operations across different storage backends

package repl

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/lexlapax/magellai/pkg/llm"
)

func init() {
	// Set log level to ERROR for benchmarks to reduce noise
	os.Setenv("MAGELLAI_LOG_LEVEL", "error")
}

// setupBenchmarkSession creates a session with realistic data
func setupBenchmarkSession(storage StorageBackend, name string, messageCount int) *Session {
	session := storage.NewSession(name)
	
	// Add realistic messages
	for i := 0; i < messageCount; i++ {
		if i%2 == 0 {
			session.Conversation.AddMessage("user", generateLongMessage(i), nil)
		} else {
			session.Conversation.AddMessage("assistant", generateLongResponse(i), nil)
		}
		
		// Add attachments to some messages
		if i%3 == 0 {
			attachments := []llm.Attachment{
				{Type: llm.AttachmentTypeFile, FilePath: "test.txt", Content: "test content"},
			}
			session.Conversation.AddMessage("user", "Here's a file", attachments)
		}
	}
	
	return session
}

func generateLongMessage(seed int) string {
	base := "This is a long user message with lots of content to simulate realistic usage. "
	return base + "The message contains various topics and questions about programming, " +
		"machine learning, and software development. Message number: " + 
		string([]rune("0123456789")[seed%10])
}

func generateLongResponse(seed int) string {
	base := "This is a comprehensive assistant response with detailed explanations. "
	return base + "The response covers multiple aspects of the question, provides examples, " +
		"and includes code snippets and technical details. Response number: " + 
		string([]rune("0123456789")[seed%10])
}

// Benchmark session creation
func BenchmarkSessionCreation(b *testing.B) {
	benchmarks := []struct {
		name    string
		factory func() (StorageBackend, func())
	}{
		{
			name: "FileSystem",
			factory: func() (StorageBackend, func()) {
				tmpDir, _ := os.MkdirTemp("", "bench-fs-")
				backend, _ := CreateStorageBackend(FileSystemStorage, map[string]interface{}{
					"base_dir": tmpDir,
				})
				return backend, func() { os.RemoveAll(tmpDir) }
			},
		},
		{
			name: "SQLite",
			factory: func() (StorageBackend, func()) {
				tmpDir, _ := os.MkdirTemp("", "bench-sqlite-")
				dbPath := filepath.Join(tmpDir, "bench.db")
				backend, _ := CreateStorageBackend(SQLiteStorage, map[string]interface{}{
					"path": dbPath,
				})
				return backend, func() { 
					backend.Close()
					os.RemoveAll(tmpDir) 
				}
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			backend, cleanup := bm.factory()
			defer cleanup()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				session := backend.NewSession("bench-session")
				if session == nil {
					b.Fatal("Failed to create session")
				}
			}
		})
	}
}

// Benchmark saving sessions
func BenchmarkSessionSave(b *testing.B) {
	messageCounts := []int{10, 50, 100}
	
	for _, msgCount := range messageCounts {
		benchmarks := []struct {
			name    string
			factory func() (StorageBackend, func())
		}{
			{
				name: "FileSystem",
				factory: func() (StorageBackend, func()) {
					tmpDir, _ := os.MkdirTemp("", "bench-fs-")
					backend, _ := CreateStorageBackend(FileSystemStorage, map[string]interface{}{
						"base_dir": tmpDir,
					})
					return backend, func() { os.RemoveAll(tmpDir) }
				},
			},
			{
				name: "SQLite",
				factory: func() (StorageBackend, func()) {
					tmpDir, _ := os.MkdirTemp("", "bench-sqlite-")
					dbPath := filepath.Join(tmpDir, "bench.db")
					backend, _ := CreateStorageBackend(SQLiteStorage, map[string]interface{}{
						"path": dbPath,
					})
					return backend, func() { 
						backend.Close()
						os.RemoveAll(tmpDir) 
					}
				},
			},
		}

		for _, bm := range benchmarks {
			b.Run(bm.name+"-"+strconv.Itoa(msgCount)+"msgs", func(b *testing.B) {
				backend, cleanup := bm.factory()
				defer cleanup()

				// Create sessions outside of benchmark loop
				sessions := make([]*Session, b.N)
				for i := 0; i < b.N; i++ {
					sessions[i] = setupBenchmarkSession(backend, "bench-session", msgCount)
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if err := backend.SaveSession(sessions[i]); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

// Benchmark loading sessions
func BenchmarkSessionLoad(b *testing.B) {
	benchmarks := []struct {
		name    string
		factory func() (StorageBackend, func())
	}{
		{
			name: "FileSystem",
			factory: func() (StorageBackend, func()) {
				tmpDir, _ := os.MkdirTemp("", "bench-fs-")
				backend, _ := CreateStorageBackend(FileSystemStorage, map[string]interface{}{
					"base_dir": tmpDir,
				})
				return backend, func() { os.RemoveAll(tmpDir) }
			},
		},
		{
			name: "SQLite",
			factory: func() (StorageBackend, func()) {
				tmpDir, _ := os.MkdirTemp("", "bench-sqlite-")
				dbPath := filepath.Join(tmpDir, "bench.db")
				backend, _ := CreateStorageBackend(SQLiteStorage, map[string]interface{}{
					"path": dbPath,
				})
				return backend, func() { 
					backend.Close()
					os.RemoveAll(tmpDir) 
				}
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			backend, cleanup := bm.factory()
			defer cleanup()

			// Create and save a session
			session := setupBenchmarkSession(backend, "bench-session", 50)
			backend.SaveSession(session)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				loaded, err := backend.LoadSession(session.ID)
				if err != nil {
					b.Fatal(err)
				}
				if loaded == nil {
					b.Fatal("Failed to load session")
				}
			}
		})
	}
}

// Benchmark listing sessions
func BenchmarkSessionList(b *testing.B) {
	sessionCounts := []int{10, 50, 100}
	
	for _, count := range sessionCounts {
		benchmarks := []struct {
			name    string
			factory func() (StorageBackend, func())
		}{
			{
				name: "FileSystem",
				factory: func() (StorageBackend, func()) {
					tmpDir, _ := os.MkdirTemp("", "bench-fs-")
					backend, _ := CreateStorageBackend(FileSystemStorage, map[string]interface{}{
						"base_dir": tmpDir,
					})
					return backend, func() { os.RemoveAll(tmpDir) }
				},
			},
			{
				name: "SQLite",
				factory: func() (StorageBackend, func()) {
					tmpDir, _ := os.MkdirTemp("", "bench-sqlite-")
					dbPath := filepath.Join(tmpDir, "bench.db")
					backend, _ := CreateStorageBackend(SQLiteStorage, map[string]interface{}{
						"path": dbPath,
					})
					return backend, func() { 
						backend.Close()
						os.RemoveAll(tmpDir) 
					}
				},
			},
		}

		for _, bm := range benchmarks {
			b.Run(bm.name+"-"+strconv.Itoa(count)+"sessions", func(b *testing.B) {
				backend, cleanup := bm.factory()
				defer cleanup()

				// Create sessions
				for i := 0; i < count; i++ {
					session := setupBenchmarkSession(backend, "bench-session-"+strconv.Itoa(i), 10)
					backend.SaveSession(session)
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					sessions, err := backend.ListSessions()
					if err != nil {
						b.Fatal(err)
					}
					if len(sessions) != count {
						b.Fatalf("Expected %d sessions, got %d", count, len(sessions))
					}
				}
			})
		}
	}
}

// Benchmark searching sessions
func BenchmarkSessionSearch(b *testing.B) {
	benchmarks := []struct {
		name    string
		factory func() (StorageBackend, func())
	}{
		{
			name: "FileSystem",
			factory: func() (StorageBackend, func()) {
				tmpDir, _ := os.MkdirTemp("", "bench-fs-")
				backend, _ := CreateStorageBackend(FileSystemStorage, map[string]interface{}{
					"base_dir": tmpDir,
				})
				return backend, func() { os.RemoveAll(tmpDir) }
			},
		},
		{
			name: "SQLite",
			factory: func() (StorageBackend, func()) {
				tmpDir, _ := os.MkdirTemp("", "bench-sqlite-")
				dbPath := filepath.Join(tmpDir, "bench.db")
				backend, _ := CreateStorageBackend(SQLiteStorage, map[string]interface{}{
					"path": dbPath,
				})
				return backend, func() { 
					backend.Close()
					os.RemoveAll(tmpDir) 
				}
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			backend, cleanup := bm.factory()
			defer cleanup()

			// Create sessions with searchable content
			for i := 0; i < 20; i++ {
				session := backend.NewSession("search-session-" + strconv.Itoa(i))
				session.Conversation.AddMessage("user", "Tell me about programming languages", nil)
				session.Conversation.AddMessage("assistant", "Programming languages are tools for software development", nil)
				if i%2 == 0 {
					session.Conversation.AddMessage("user", "What about golang?", nil)
					session.Conversation.AddMessage("assistant", "Golang is a statically typed language", nil)
				}
				backend.SaveSession(session)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				results, err := backend.SearchSessions("golang")
				if err != nil {
					b.Fatal(err)
				}
				if len(results) == 0 {
					b.Fatal("No search results found")
				}
			}
		})
	}
}

// Benchmark complete workflow
func BenchmarkCompleteWorkflow(b *testing.B) {
	benchmarks := []struct {
		name    string
		factory func() (StorageBackend, func())
	}{
		{
			name: "FileSystem",
			factory: func() (StorageBackend, func()) {
				tmpDir, _ := os.MkdirTemp("", "bench-fs-")
				backend, _ := CreateStorageBackend(FileSystemStorage, map[string]interface{}{
					"base_dir": tmpDir,
				})
				return backend, func() { os.RemoveAll(tmpDir) }
			},
		},
		{
			name: "SQLite",
			factory: func() (StorageBackend, func()) {
				tmpDir, _ := os.MkdirTemp("", "bench-sqlite-")
				dbPath := filepath.Join(tmpDir, "bench.db")
				backend, _ := CreateStorageBackend(SQLiteStorage, map[string]interface{}{
					"path": dbPath,
				})
				return backend, func() { 
					backend.Close()
					os.RemoveAll(tmpDir) 
				}
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			backend, cleanup := bm.factory()
			defer cleanup()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Create session
				session := backend.NewSession("workflow-session")
				
				// Add messages
				session.Conversation.AddMessage("user", "Hello", nil)
				session.Conversation.AddMessage("assistant", "Hi there!", nil)
				
				// Save
				if err := backend.SaveSession(session); err != nil {
					b.Fatal(err)
				}
				
				// Load
				loaded, err := backend.LoadSession(session.ID)
				if err != nil {
					b.Fatal(err)
				}
				
				// Update
				loaded.Conversation.AddMessage("user", "How are you?", nil)
				if err := backend.SaveSession(loaded); err != nil {
					b.Fatal(err)
				}
				
				// List
				sessions, err := backend.ListSessions()
				if err != nil {
					b.Fatal(err)
				}
				if len(sessions) == 0 {
					b.Fatal("No sessions found")
				}
				
				// Search
				results, err := backend.SearchSessions("Hello")
				if err != nil {
					b.Fatal(err)
				}
				if len(results) == 0 {
					b.Fatal("No search results")
				}
				
				// Delete
				if err := backend.DeleteSession(session.ID); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}