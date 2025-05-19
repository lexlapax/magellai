// ABOUTME: Example demonstrating session export functionality
// ABOUTME: Shows how to export sessions in JSON and Markdown formats

package repl_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lexlapax/magellai/pkg/domain"
	"github.com/lexlapax/magellai/pkg/repl/session"
	"github.com/lexlapax/magellai/pkg/storage"
	_ "github.com/lexlapax/magellai/pkg/storage/filesystem" // Register filesystem backend
)

func ExampleSessionManager_ExportSession() {
	// Create a temporary directory for the session
	tempDir, err := os.MkdirTemp("", "repl_export_example")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a storage manager
	storageManager, err := session.CreateStorageManager(storage.FileSystemBackend, storage.Config{
		"base_dir": tempDir,
	})
	if err != nil {
		panic(err)
	}

	// Create session manager
	manager := &session.SessionManager{StorageManager: storageManager}

	// Create a new session
	sess, err := manager.NewSession("Example Chat")
	if err != nil {
		panic(err)
	}

	// Add some conversation messages
	sess.Conversation.SetSystemPrompt("You are a helpful assistant.")
	sess.Conversation.AddMessage(*domain.NewMessage("msg1", domain.MessageRoleUser, "Hello, can you help me with a task?"))
	sess.Conversation.AddMessage(*domain.NewMessage("msg2", domain.MessageRoleAssistant, "Of course! I'd be happy to help. What task do you need assistance with?"))

	// Add a message with an attachment
	attachment := domain.Attachment{
		Type:     domain.AttachmentTypeText,
		FilePath: "example.txt",
		MimeType: "text/plain",
		Content:  []byte("This is example content"),
	}
	msg := domain.NewMessage("msg3", domain.MessageRoleUser, "Please analyze this file")
	msg.Attachments = []domain.Attachment{attachment}
	sess.Conversation.AddMessage(*msg)
	sess.Conversation.AddMessage(*domain.NewMessage("msg4", domain.MessageRoleAssistant, "I've analyzed the file. It contains example content."))

	// Save the session
	if err := manager.SaveSession(sess); err != nil {
		panic(err)
	}

	// Export as JSON
	fmt.Println("=== JSON Export ===")
	if err := manager.ExportSession(sess.ID, "json", os.Stdout); err != nil {
		panic(err)
	}
	fmt.Println()

	// Export as Markdown
	fmt.Println("=== Markdown Export ===")
	if err := manager.ExportSession(sess.ID, "markdown", os.Stdout); err != nil {
		panic(err)
	}

	// Export to files
	jsonFile := filepath.Join(tempDir, "session_export.json")
	file, err := os.Create(jsonFile)
	if err == nil {
		if err := manager.ExportSession(sess.ID, "json", file); err != nil {
			panic(err)
		}
		file.Close()
		fmt.Printf("\nExported to file: %s\n", jsonFile)
	}

	mdFile := filepath.Join(tempDir, "session_export.md")
	file, err = os.Create(mdFile)
	if err == nil {
		if err := manager.ExportSession(sess.ID, "markdown", file); err != nil {
			panic(err)
		}
		file.Close()
		fmt.Printf("Exported to file: %s\n", mdFile)
	}
}
