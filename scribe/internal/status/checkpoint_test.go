package status

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCheckpointStore_ReadWrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	checkpointPath := filepath.Join(tmpDir, "checkpoint.json")
	store := NewCheckpointStore(checkpointPath)

	// Test read from non-existent file
	checkpoint, err := store.Read()
	if err != nil {
		t.Fatalf("Read non-existent: %v", err)
	}
	if checkpoint != nil {
		t.Error("expected nil checkpoint for non-existent file")
	}

	// Test write (forced, to bypass interval check)
	cp := NewTaskCheckpoint(TaskTypeExploration, "Test task", "test/PROMPT.md", 50, nil)
	if err := store.WriteForced(cp); err != nil {
		t.Fatalf("WriteForced: %v", err)
	}

	// Test read existing
	read, err := store.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if read == nil {
		t.Fatal("expected checkpoint to exist")
	}
	if read.TaskDescription != "Test task" {
		t.Errorf("unexpected description: %s", read.TaskDescription)
	}
	if read.TaskType != TaskTypeExploration {
		t.Errorf("unexpected type: %s", read.TaskType)
	}
}

func TestCheckpointStore_IntervalEnforcement(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	checkpointPath := filepath.Join(tmpDir, "checkpoint.json")
	store := NewCheckpointStore(checkpointPath)

	// Create initial checkpoint
	cp := NewTaskCheckpoint(TaskTypeExploration, "Test task", "test/PROMPT.md", 50, nil)
	if err := store.WriteForced(cp); err != nil {
		t.Fatalf("WriteForced: %v", err)
	}

	// Try to write again immediately (should fail due to interval)
	cp.IterationCount = 1
	err = store.Write(cp)
	if err == nil {
		t.Error("expected error due to checkpoint interval")
	}
}

func TestCheckpointStore_Delete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	checkpointPath := filepath.Join(tmpDir, "checkpoint.json")
	store := NewCheckpointStore(checkpointPath)

	// Create checkpoint
	cp := NewTaskCheckpoint(TaskTypeExploration, "Test task", "test/PROMPT.md", 50, nil)
	store.WriteForced(cp)

	if !store.Exists() {
		t.Error("expected checkpoint to exist")
	}

	// Delete
	if err := store.Delete(); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if store.Exists() {
		t.Error("expected checkpoint to not exist after delete")
	}

	// Delete non-existent should not error
	if err := store.Delete(); err != nil {
		t.Errorf("Delete non-existent: %v", err)
	}
}

func TestActionLogger_Log(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scribe-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logPath := filepath.Join(tmpDir, "logs", "actions.jsonl")
	logger := NewActionLogger(logPath)

	// Log some actions
	msg := "Found interesting code"
	if err := logger.Log("task-1", AgentTypeExploration, "search_codebase", "fasttree", AgentActionResultSuccess, &msg); err != nil {
		t.Fatalf("Log: %v", err)
	}

	if err := logger.Log("task-1", AgentTypeExploration, "queue_candidate", "S2:abc123", AgentActionResultSuccess, nil); err != nil {
		t.Fatalf("Log: %v", err)
	}

	// Read all
	logs, err := logger.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(logs) != 2 {
		t.Errorf("expected 2 logs, got %d", len(logs))
	}

	// Read for task
	taskLogs, err := logger.ReadForTask("task-1")
	if err != nil {
		t.Fatalf("ReadForTask: %v", err)
	}
	if len(taskLogs) != 2 {
		t.Errorf("expected 2 task logs, got %d", len(taskLogs))
	}

	// Clear
	if err := logger.Clear(); err != nil {
		t.Fatalf("Clear: %v", err)
	}

	logs, _ = logger.ReadAll()
	if len(logs) != 0 {
		t.Errorf("expected 0 logs after clear, got %d", len(logs))
	}
}

func TestBlockingDetector_ShouldQueueForReview(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*TaskCheckpoint)
		expectedQueue  bool
		expectedReason string
	}{
		{
			name: "no blocked items",
			setup: func(cp *TaskCheckpoint) {
				// Default - no blocked items
			},
			expectedQueue:  false,
			expectedReason: "",
		},
		{
			name: "recent blocked item",
			setup: func(cp *TaskCheckpoint) {
				cp.State.BlockedItems = []BlockedItem{
					{Item: "test", Reason: "rate limited", BlockedAt: time.Now()},
				}
			},
			expectedQueue:  false,
			expectedReason: "",
		},
		{
			name: "old blocked item",
			setup: func(cp *TaskCheckpoint) {
				cp.State.BlockedItems = []BlockedItem{
					{Item: "test", Reason: "rate limited", BlockedAt: time.Now().Add(-35 * time.Minute)},
				}
			},
			expectedQueue:  true,
			expectedReason: "Blocked item for >30 minutes: test",
		},
		{
			name: "max iterations reached",
			setup: func(cp *TaskCheckpoint) {
				cp.IterationCount = 50
				cp.MaxIterations = 50
			},
			expectedQueue:  true,
			expectedReason: "Reached maximum iterations",
		},
		{
			name: "cost budget exceeded",
			setup: func(cp *TaskCheckpoint) {
				budget := 10.0
				cp.CostBudgetUSD = &budget
				cp.Metrics.EstimatedCostUSD = 11.0
			},
			expectedQueue:  true,
			expectedReason: "Reached cost budget",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := NewTaskCheckpoint(TaskTypeExploration, "Test", "test.md", 50, nil)
			tt.setup(cp)

			detector := NewBlockingDetector(cp)
			shouldQueue, reason := detector.ShouldQueueForReview()

			if shouldQueue != tt.expectedQueue {
				t.Errorf("shouldQueue: got %v, want %v", shouldQueue, tt.expectedQueue)
			}
			if reason != tt.expectedReason {
				t.Errorf("reason: got %q, want %q", reason, tt.expectedReason)
			}
		})
	}
}

func TestNewTaskCheckpoint(t *testing.T) {
	budget := 10.0
	cp := NewTaskCheckpoint(TaskTypeSurvey, "Test survey", "survey/PROMPT.md", 100, &budget)

	if cp.TaskID == "" {
		t.Error("expected task ID to be set")
	}
	if cp.TaskType != TaskTypeSurvey {
		t.Errorf("unexpected type: %s", cp.TaskType)
	}
	if cp.TaskDescription != "Test survey" {
		t.Errorf("unexpected description: %s", cp.TaskDescription)
	}
	if cp.MaxIterations != 100 {
		t.Errorf("unexpected max iterations: %d", cp.MaxIterations)
	}
	if cp.CostBudgetUSD == nil || *cp.CostBudgetUSD != 10.0 {
		t.Error("unexpected cost budget")
	}
	if cp.IterationCount != 0 {
		t.Errorf("unexpected iteration count: %d", cp.IterationCount)
	}
}
