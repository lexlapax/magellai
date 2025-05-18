// ABOUTME: Test file for session branching functionality
// ABOUTME: Tests branch creation, tree traversal, and branch management

package domain

import (
	"testing"
	"time"
)

func TestSession_CreateBranch(t *testing.T) {
	// Create a parent session with some messages
	parent := NewSession("parent-1")
	parent.Conversation.AddMessage(Message{
		ID:      "msg-1",
		Role:    MessageRoleUser,
		Content: "Hello",
		Timestamp: time.Now(),
	})
	parent.Conversation.AddMessage(Message{
		ID:      "msg-2",
		Role:    MessageRoleAssistant,
		Content: "Hi there!",
		Timestamp: time.Now(),
	})
	parent.Conversation.AddMessage(Message{
		ID:      "msg-3",
		Role:    MessageRoleUser,
		Content: "How are you?",
		Timestamp: time.Now(),
	})
	parent.Tags = []string{"test", "conversation"}
	parent.Config["custom"] = "value"

	tests := []struct {
		name         string
		branchID     string
		branchName   string
		messageIndex int
		expectError  bool
		expectedMsgs int
	}{
		{
			name:         "branch at beginning",
			branchID:     "branch-1",
			branchName:   "Early Branch",
			messageIndex: 0,
			expectError:  false,
			expectedMsgs: 0,
		},
		{
			name:         "branch after first message",
			branchID:     "branch-2",
			branchName:   "After Hello",
			messageIndex: 1,
			expectError:  false,
			expectedMsgs: 1,
		},
		{
			name:         "branch at end",
			branchID:     "branch-3",
			branchName:   "Latest Branch",
			messageIndex: 3,
			expectError:  false,
			expectedMsgs: 3,
		},
		{
			name:         "invalid index negative",
			branchID:     "branch-4",
			branchName:   "Invalid",
			messageIndex: -1,
			expectError:  true,
			expectedMsgs: 0,
		},
		{
			name:         "invalid index too large",
			branchID:     "branch-5",
			branchName:   "Invalid",
			messageIndex: 10,
			expectError:  true,
			expectedMsgs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			branch, err := parent.CreateBranch(tt.branchID, tt.branchName, tt.messageIndex)
			
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			// Verify branch properties
			if branch.ID != tt.branchID {
				t.Errorf("expected branch ID %s, got %s", tt.branchID, branch.ID)
			}
			
			if branch.Name != tt.branchName {
				t.Errorf("expected branch name %s, got %s", tt.branchName, branch.Name)
			}
			
			if branch.ParentID != parent.ID {
				t.Errorf("expected parent ID %s, got %s", parent.ID, branch.ParentID)
			}
			
			if branch.BranchPoint != tt.messageIndex {
				t.Errorf("expected branch point %d, got %d", tt.messageIndex, branch.BranchPoint)
			}
			
			if len(branch.Conversation.Messages) != tt.expectedMsgs {
				t.Errorf("expected %d messages, got %d", tt.expectedMsgs, len(branch.Conversation.Messages))
			}
			
			// Verify conversation properties are copied
			if branch.Conversation.Model != parent.Conversation.Model {
				t.Error("model not copied correctly")
			}
			
			if branch.Conversation.Provider != parent.Conversation.Provider {
				t.Error("provider not copied correctly")
			}
			
			// Verify tags are copied
			if len(branch.Tags) != len(parent.Tags) {
				t.Error("tags not copied correctly")
			}
			
			// Verify config is copied
			if branch.Config["custom"] != parent.Config["custom"] {
				t.Error("config not copied correctly")
			}
			
			// Verify parent has this branch as a child
			found := false
			for _, childID := range parent.ChildIDs {
				if childID == branch.ID {
					found = true
					break
				}
			}
			if !found {
				t.Error("parent does not have branch in child list")
			}
		})
	}
}

func TestSession_BranchManagement(t *testing.T) {
	session := NewSession("test-session")
	
	// Test adding children
	session.AddChild("child-1")
	session.AddChild("child-2")
	session.AddChild("child-1") // duplicate should be ignored
	
	if len(session.ChildIDs) != 2 {
		t.Errorf("expected 2 children, got %d", len(session.ChildIDs))
	}
	
	// Test removing children
	session.RemoveChild("child-1")
	
	if len(session.ChildIDs) != 1 {
		t.Errorf("expected 1 child after removal, got %d", len(session.ChildIDs))
	}
	
	if session.ChildIDs[0] != "child-2" {
		t.Errorf("expected remaining child to be 'child-2', got %s", session.ChildIDs[0])
	}
}

func TestSession_BranchChecks(t *testing.T) {
	// Test root session
	root := NewSession("root")
	
	if root.IsBranch() {
		t.Error("root session should not be a branch")
	}
	
	if root.HasBranches() {
		t.Error("root session should not have branches initially")
	}
	
	// Test branch session
	branch := NewSession("branch")
	branch.ParentID = "root"
	
	if !branch.IsBranch() {
		t.Error("session with parent ID should be a branch")
	}
	
	// Test session with children
	root.AddChild("branch")
	
	if !root.HasBranches() {
		t.Error("session with children should have branches")
	}
}

func TestSessionInfo_BranchInformation(t *testing.T) {
	// Create a branched session
	parent := NewSession("parent")
	branch, _ := parent.CreateBranch("branch-1", "Test Branch", 0)
	
	// Get session info
	parentInfo := parent.ToSessionInfo()
	branchInfo := branch.ToSessionInfo()
	
	// Check parent info
	if parentInfo.ParentID != "" {
		t.Error("parent should not have a parent ID")
	}
	
	if parentInfo.ChildCount != 1 {
		t.Errorf("expected parent to have 1 child, got %d", parentInfo.ChildCount)
	}
	
	if parentInfo.IsBranch {
		t.Error("parent should not be marked as a branch")
	}
	
	// Check branch info
	if branchInfo.ParentID != "parent" {
		t.Errorf("expected branch parent ID to be 'parent', got %s", branchInfo.ParentID)
	}
	
	if branchInfo.BranchName != "Test Branch" {
		t.Errorf("expected branch name to be 'Test Branch', got %s", branchInfo.BranchName)
	}
	
	if !branchInfo.IsBranch {
		t.Error("branch should be marked as a branch")
	}
	
	if branchInfo.ChildCount != 0 {
		t.Errorf("expected branch to have 0 children, got %d", branchInfo.ChildCount)
	}
}