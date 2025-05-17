// ABOUTME: Example demonstrating session export functionality
// ABOUTME: Shows how to export sessions in JSON and Markdown formats

package repl_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lexlapax/magellai/pkg/llm"
	"github.com/lexlapax/magellai/pkg/repl"
)

func ExampleSessionManager_ExportSession() {
	// Create a temporary directory for the session
	tempDir, err := os.MkdirTemp("", "repl_export_example")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a session manager
	storage, err := repl.CreateStorageBackend(repl.FileSystemStorage, map[string]interface{}{
		"base_dir": tempDir,
	})
	if err != nil {
		panic(err)
	}

	manager, err := repl.NewSessionManager(storage)
	if err != nil {
		panic(err)
	}

	// Create a new session
	session, err := manager.NewSession("Example Chat")
	if err != nil {
		panic(err)
	}

	// Add some conversation messages
	session.Conversation.SetSystemPrompt("You are a helpful assistant.")
	session.Conversation.AddMessage("user", "Hello, can you help me with a task?", nil)
	session.Conversation.AddMessage("assistant", "Of course! I'd be happy to help. What task do you need assistance with?", nil)

	// Add a message with an attachment
	attachment := llm.Attachment{
		Type:     llm.AttachmentTypeText,
		FilePath: "example.txt",
		MimeType: "text/plain",
		Content:  "This is example content",
	}
	session.Conversation.AddMessage("user", "Please analyze this file", []llm.Attachment{attachment})
	session.Conversation.AddMessage("assistant", "I've analyzed the file. It contains example content.", nil)

	// Save the session
	if err := manager.SaveSession(session); err != nil {
		panic(err)
	}

	// Export as JSON
	fmt.Println("=== JSON Export ===")
	if err := manager.ExportSession(session.ID, "json", os.Stdout); err != nil {
		panic(err)
	}
	fmt.Println()

	// Export as Markdown
	fmt.Println("=== Markdown Export ===")
	if err := manager.ExportSession(session.ID, "markdown", os.Stdout); err != nil {
		panic(err)
	}

	// Export to files
	jsonFile := filepath.Join(tempDir, "session_export.json")
	file, err := os.Create(jsonFile)
	if err == nil {
		if err := manager.ExportSession(session.ID, "json", file); err != nil {
			panic(err)
		}
		file.Close()
		fmt.Printf("\nExported to file: %s\n", jsonFile)
	}

	mdFile := filepath.Join(tempDir, "session_export.md")
	file, err = os.Create(mdFile)
	if err == nil {
		if err := manager.ExportSession(session.ID, "markdown", file); err != nil {
			panic(err)
		}
		file.Close()
		fmt.Printf("Exported to file: %s\n", mdFile)
	}
}
