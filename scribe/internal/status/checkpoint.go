package status

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DefaultCheckpointPath is the default path for task checkpoints.
const DefaultCheckpointPath = ".claude/authoring/checkpoint.json"

// CheckpointStore provides checkpoint read/write operations.
type CheckpointStore struct {
	path string
}

// NewCheckpointStore creates a new CheckpointStore.
func NewCheckpointStore(path string) *CheckpointStore {
	if path == "" {
		path = DefaultCheckpointPath
	}
	return &CheckpointStore{path: path}
}

// Read reads the current checkpoint.
func (s *CheckpointStore) Read() (*TaskCheckpoint, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read checkpoint: %w", err)
	}

	var checkpoint TaskCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, fmt.Errorf("parse checkpoint: %w", err)
	}

	return &checkpoint, nil
}

// Write writes a checkpoint to disk.
// Enforces minimum checkpoint interval of 5 minutes per FR-044a.
func (s *CheckpointStore) Write(checkpoint *TaskCheckpoint) error {
	// Check if we can write (enforce 5-minute minimum interval)
	existing, err := s.Read()
	if err != nil {
		return err
	}

	if existing != nil {
		timeSince := time.Since(existing.LastCheckpoint)
		if timeSince < MinCheckpointInterval {
			return fmt.Errorf("checkpoint interval too short: %v (minimum %v)", timeSince, MinCheckpointInterval)
		}
	}

	// Update last checkpoint time
	checkpoint.LastCheckpoint = time.Now()

	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create checkpoint directory: %w", err)
	}

	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal checkpoint: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return fmt.Errorf("write checkpoint: %w", err)
	}

	return nil
}

// WriteForced writes a checkpoint without checking the interval.
// Use this only for initial checkpoint creation.
func (s *CheckpointStore) WriteForced(checkpoint *TaskCheckpoint) error {
	checkpoint.LastCheckpoint = time.Now()

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create checkpoint directory: %w", err)
	}

	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal checkpoint: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return fmt.Errorf("write checkpoint: %w", err)
	}

	return nil
}

// Delete removes the checkpoint file.
func (s *CheckpointStore) Delete() error {
	err := os.Remove(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// Exists returns true if a checkpoint file exists.
func (s *CheckpointStore) Exists() bool {
	_, err := os.Stat(s.path)
	return err == nil
}

// NewTaskCheckpoint creates a new task checkpoint with default values.
func NewTaskCheckpoint(taskType TaskType, description, promptFile string, maxIterations int, costBudget *float64) *TaskCheckpoint {
	return &TaskCheckpoint{
		TaskID:          generateTaskID(),
		TaskType:        taskType,
		TaskDescription: description,
		StartedAt:       time.Now(),
		LastCheckpoint:  time.Now(),
		IterationCount:  0,
		PromptFile:      promptFile,
		MaxIterations:   maxIterations,
		CostBudgetUSD:   costBudget,
		State: TaskState{
			ItemsCompleted: []string{},
			ItemsPending:   []string{},
			BlockedItems:   []BlockedItem{},
			CurrentFocus:   "",
		},
		Metrics: TaskMetrics{
			CandidatesQueued:   0,
			PapersFound:        0,
			CodeLocationsFound: 0,
			ReposSearched:      0,
			EstimatedCostUSD:   0,
		},
	}
}

// generateTaskID generates a unique task ID.
func generateTaskID() string {
	return fmt.Sprintf("task-%d", time.Now().UnixNano())
}
