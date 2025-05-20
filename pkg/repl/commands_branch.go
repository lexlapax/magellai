// ABOUTME: REPL commands for session branching and tree management
// ABOUTME: Implements /branch, /branches, /switch, and branch visualization

package repl

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lexlapax/magellai/internal/logging"
	"github.com/lexlapax/magellai/pkg/domain"
)

// cmdBranch creates a new branch from the current session
func (r *REPL) cmdBranch(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: /branch <name> [at <message_index>]")
	}

	branchName := args[0]
	messageIndex := len(r.session.Conversation.Messages) // Default to end

	// Parse optional message index
	if len(args) >= 3 && args[1] == "at" {
		idx, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid message index: %v", err)
		}
		messageIndex = idx
	}

	// Get current session
	currentSession := r.session
	if currentSession == nil {
		return errors.New("no active session")
	}

	// Generate new branch ID
	branchID := r.manager.GenerateSessionID()

	// Create the branch
	branch, err := currentSession.CreateBranch(branchID, branchName, messageIndex)
	if err != nil {
		return fmt.Errorf("failed to create branch: %v", err)
	}

	// Save both the parent and the new branch
	if err := r.manager.SaveSession(currentSession); err != nil {
		return fmt.Errorf("failed to update parent session: %v", err)
	}

	if err := r.manager.SaveSession(branch); err != nil {
		return fmt.Errorf("failed to save new branch: %v", err)
	}

	logging.LogInfo("Created branch",
		"branch_id", branchID,
		"branch_name", branchName,
		"parent_id", currentSession.ID,
		"parent_name", currentSession.Name,
		"branch_point", messageIndex,
		"message_count", len(branch.Conversation.Messages))

	// Optionally switch to the new branch
	fmt.Fprintf(r.writer, "Created branch '%s' (ID: %s) at message %d\n", branchName, branchID, messageIndex)
	fmt.Fprintf(r.writer, "To switch to this branch, use: /switch %s\n", branchID)

	return nil
}

// cmdBranches lists all branches of the current session
func (r *REPL) cmdBranches(args []string) error {
	currentSession := r.session
	if currentSession == nil {
		return errors.New("no active session")
	}

	var sessionToList *domain.Session

	// If current session is a branch, list siblings too
	if currentSession.IsBranch() {
		parent, err := r.manager.StorageManager.LoadSession(currentSession.ParentID)
		if err != nil {
			return fmt.Errorf("failed to get parent session: %v", err)
		}
		sessionToList = parent
		fmt.Fprintf(r.writer, "Branches of parent session '%s':\n", parent.Name)
	} else {
		sessionToList = currentSession
		fmt.Fprintf(r.writer, "Branches of current session:\n")
	}

	if len(sessionToList.ChildIDs) == 0 {
		fmt.Fprintln(r.writer, "No branches found.")
		return nil
	}

	// Get info for all child branches
	children, err := r.manager.GetChildren(sessionToList.ID)
	if err != nil {
		return fmt.Errorf("failed to get child branches: %v", err)
	}

	// Display branches
	fmt.Fprintln(r.writer, "")
	for _, child := range children {
		indicator := " "
		if child.ID == currentSession.ID {
			indicator = "*" // Mark current branch
		}

		fmt.Fprintf(r.writer, "%s %s - %s (ID: %s) - %d messages, created %s\n",
			indicator,
			child.BranchName,
			child.Name,
			child.ID,
			child.MessageCount,
			child.Created.Format("2006-01-02 15:04"),
		)
	}

	return nil
}

// cmdTree shows the branch tree for the current session
func (r *REPL) cmdTree(args []string) error {
	currentSession := r.session
	if currentSession == nil {
		return errors.New("no active session")
	}

	// Find the root of the tree
	rootID := currentSession.ID
	if currentSession.IsBranch() {
		// Traverse up to find root (this is simplified - in practice would need recursive lookup)
		root, err := r.manager.StorageManager.LoadSession(currentSession.ParentID)
		if err == nil && root.ParentID == "" {
			rootID = root.ID
		}
	}

	// Get the branch tree
	tree, err := r.manager.GetBranchTree(rootID)
	if err != nil {
		return fmt.Errorf("failed to get branch tree: %v", err)
	}

	// Display the tree
	fmt.Fprintln(r.writer, "Session Branch Tree:")
	displayTree(r.writer, tree, "", currentSession.ID)

	return nil
}

// displayTree recursively displays the branch tree
func displayTree(out interface{}, tree *domain.BranchTree, prefix string, currentID string) {
	if tree == nil || tree.Session == nil {
		return
	}

	// Determine if this is the current session
	marker := ""
	if tree.Session.ID == currentID {
		marker = " *"
	}

	// Display this node
	fmt.Fprintf(out.(interface{ Write([]byte) (int, error) }),
		"%s%s (ID: %s) - %d messages%s\n",
		prefix,
		tree.Session.Name,
		tree.Session.ID,
		tree.Session.MessageCount,
		marker,
	)

	// Display children
	for i, child := range tree.Children {
		var childPrefix string
		if i == len(tree.Children)-1 {
			fmt.Fprintf(out.(interface{ Write([]byte) (int, error) }), "%s└─ ", prefix)
			childPrefix = prefix + "   "
		} else {
			fmt.Fprintf(out.(interface{ Write([]byte) (int, error) }), "%s├─ ", prefix)
			childPrefix = prefix + "│  "
		}
		displayTree(out, child, childPrefix, currentID)
	}
}

// cmdSwitch switches to a different branch
func (r *REPL) cmdSwitch(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: /switch <branch_id>")
	}

	branchID := args[0]

	// Get the target branch
	branch, err := r.manager.StorageManager.LoadSession(branchID)
	if err != nil {
		return fmt.Errorf("failed to get branch: %v", err)
	}

	// Save current session if it has unsaved changes
	currentSession := r.session
	if currentSession != nil && r.hasUnsavedChanges() {
		if err := r.manager.SaveSession(currentSession); err != nil {
			logging.LogWarn("Failed to save current session before switching", "error", err)
		}
	}

	// Switch to the new branch
	r.session = branch

	logging.LogInfo("Switched to branch",
		"branch_id", branchID,
		"branch_name", branch.Name,
		"is_branch", branch.IsBranch(),
		"parent_id", branch.ParentID,
		"message_count", len(branch.Conversation.Messages),
		"created", branch.Created.Format(time.RFC3339))
	fmt.Fprintf(r.writer, "Switched to branch '%s' (ID: %s)\n", branch.Name, branchID)

	// Show branch info
	if branch.IsBranch() {
		fmt.Fprintf(r.writer, "Branch of: parent session (ID: %s)\n", branch.ParentID)
		fmt.Fprintf(r.writer, "Branched at message: %d\n", branch.BranchPoint)
	}
	fmt.Fprintf(r.writer, "Messages: %d\n", len(branch.Conversation.Messages))

	return nil
}

// hasUnsavedChanges checks if the current session has unsaved changes
func (r *REPL) hasUnsavedChanges() bool {
	// This is a simplified check - in practice might track modifications
	return r.session != nil && r.session.Updated.After(r.lastSaveTime)
}

// cmdMerge merges two sessions
func (r *REPL) cmdMerge(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: /merge <source_session_id> [--type <continuation|rebase>] [--create-branch] [--branch-name <name>]")
	}

	// Parse arguments
	sourceID := args[0]
	targetID := r.session.ID

	// Default options
	mergeType := domain.MergeTypeContinuation
	createBranch := false
	branchName := ""
	mergePoint := len(r.session.Conversation.Messages)

	// Parse flags
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--type":
			if i+1 < len(args) {
				i++
				switch args[i] {
				case "continuation":
					mergeType = domain.MergeTypeContinuation
				case "rebase":
					mergeType = domain.MergeTypeRebase
				default:
					return fmt.Errorf("invalid merge type: %s", args[i])
				}
			}
		case "--create-branch":
			createBranch = true
		case "--branch-name":
			if i+1 < len(args) {
				i++
				branchName = strings.Join(args[i:], " ")
				break
			}
		}
	}

	// Create merge options
	options := domain.MergeOptions{
		Type:         mergeType,
		SourceID:     sourceID,
		TargetID:     targetID,
		MergePoint:   mergePoint,
		CreateBranch: createBranch,
		BranchName:   branchName,
	}

	// Perform the merge
	logging.LogInfo("Starting session merge operation",
		"source_id", sourceID,
		"target_id", targetID,
		"merge_type", fmt.Sprintf("%d", options.Type),
		"create_branch", options.CreateBranch,
		"branch_name", options.BranchName,
		"merge_point", options.MergePoint)

	result, err := r.manager.StorageManager.MergeSessions(targetID, sourceID, options)
	if err != nil {
		logging.LogWarn("Session merge failed",
			"source_id", sourceID,
			"target_id", targetID,
			"error", err)
		return fmt.Errorf("failed to merge sessions: %w", err)
	}

	logging.LogInfo("Session merge completed successfully",
		"source_id", sourceID,
		"target_id", targetID,
		"merged_count", result.MergedCount,
		"new_branch_id", result.NewBranchID,
		"merge_type", fmt.Sprintf("%d", options.Type))

	// Display results
	fmt.Fprintf(r.writer, "Successfully merged %d messages from %s into %s\n", result.MergedCount, sourceID, targetID)

	if result.NewBranchID != "" {
		fmt.Fprintf(r.writer, "Created new branch: %s\n", result.NewBranchID)

		// Ask if user wants to switch to the new branch
		fmt.Fprint(r.writer, "Switch to new branch? (y/n): ")
		var response string
		if _, err := fmt.Fscanln(r.reader, &response); err != nil {
			logging.LogWarn("Failed to read user response", "error", err)
			return nil
		}

		if response == "y" || response == "yes" {
			return r.cmdSwitch([]string{result.NewBranchID})
		}
	}

	return nil
}
