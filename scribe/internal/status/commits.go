package status

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// GitCommitter provides incremental git commit functionality during autonomous operation.
type GitCommitter struct {
	taskID      string
	commitCount int
}

// NewGitCommitter creates a new GitCommitter.
func NewGitCommitter(taskID string) *GitCommitter {
	return &GitCommitter{
		taskID:      taskID,
		commitCount: 0,
	}
}

// HasUncommittedChanges checks if there are uncommitted changes in the repository.
func (c *GitCommitter) HasUncommittedChanges() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("git status: %w", err)
	}

	return strings.TrimSpace(stdout.String()) != "", nil
}

// Commit creates an incremental commit per FR-041.
func (c *GitCommitter) Commit(message string) error {
	// Check if there are changes to commit
	hasChanges, err := c.HasUncommittedChanges()
	if err != nil {
		return err
	}
	if !hasChanges {
		return nil // Nothing to commit
	}

	c.commitCount++

	// Stage all changes
	stageCmd := exec.Command("git", "add", "-A")
	if err := stageCmd.Run(); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	// Create commit with task ID and timestamp
	timestamp := time.Now().Format("2006-01-02 15:04")
	fullMessage := fmt.Sprintf("[scribe/%s] %s\n\nCommit %d at %s",
		c.taskID[:8], message, c.commitCount, timestamp)

	commitCmd := exec.Command("git", "commit", "-m", fullMessage)
	var stderr bytes.Buffer
	commitCmd.Stderr = &stderr

	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("git commit: %s", stderr.String())
	}

	return nil
}

// GetCommitCount returns the number of commits made by this committer.
func (c *GitCommitter) GetCommitCount() int {
	return c.commitCount
}

// IsGitRepo checks if the current directory is a git repository.
func IsGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}
