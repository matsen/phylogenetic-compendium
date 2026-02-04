// Package status implements task checkpoint and status display for the scribe CLI.
package status

import "time"

// TaskType represents the type of autonomous task.
type TaskType string

const (
	TaskTypeExploration       TaskType = "exploration"
	TaskTypeSurvey            TaskType = "survey"
	TaskTypeVerificationSweep TaskType = "verification-sweep"
)

// BlockedItem represents an item that is blocked from progress.
type BlockedItem struct {
	Item      string    `json:"item"`
	Reason    string    `json:"reason"`
	BlockedAt time.Time `json:"blocked_at"`
}

// TaskState represents the current progress state of a task.
type TaskState struct {
	ItemsCompleted []string      `json:"items_completed"`
	ItemsPending   []string      `json:"items_pending"`
	BlockedItems   []BlockedItem `json:"blocked_items"`
	CurrentFocus   string        `json:"current_focus"`
}

// TaskMetrics contains metrics about task progress.
type TaskMetrics struct {
	CandidatesQueued   int     `json:"candidates_queued"`
	PapersFound        int     `json:"papers_found"`
	CodeLocationsFound int     `json:"code_locations_found"`
	ReposSearched      int     `json:"repos_searched"`
	EstimatedCostUSD   float64 `json:"estimated_cost_usd"`
}

// TaskCheckpoint represents the progress state for autonomous operation.
type TaskCheckpoint struct {
	TaskID          string    `json:"task_id"`
	TaskType        TaskType  `json:"task_type"`
	TaskDescription string    `json:"task_description"`
	StartedAt       time.Time `json:"started_at"`
	LastCheckpoint  time.Time `json:"last_checkpoint"`
	IterationCount  int       `json:"iteration_count"`
	PromptFile      string    `json:"prompt_file"`
	MaxIterations   int       `json:"max_iterations"`
	CostBudgetUSD   *float64  `json:"cost_budget_usd,omitempty"`
	State           TaskState `json:"state"`
	Metrics         TaskMetrics `json:"metrics"`
}

// AgentActionResult represents the outcome of an agent action.
type AgentActionResult string

const (
	AgentActionResultSuccess AgentActionResult = "success"
	AgentActionResultFailure AgentActionResult = "failure"
	AgentActionResultSkipped AgentActionResult = "skipped"
)

// AgentType represents the type of agent.
type AgentType string

const (
	AgentTypeExploration  AgentType = "exploration"
	AgentTypeSurvey       AgentType = "survey"
	AgentTypeConsumer     AgentType = "consumer"
	AgentTypeVerification AgentType = "verification"
)

// AgentActionLog records an agent's action.
type AgentActionLog struct {
	LogID     string            `json:"log_id"`
	TaskID    string            `json:"task_id"`
	AgentType AgentType         `json:"agent_type"`
	Action    string            `json:"action"`
	Target    string            `json:"target"`
	Result    AgentActionResult `json:"result"`
	Message   *string           `json:"message,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// MinCheckpointInterval is the minimum time between checkpoints (5 minutes per FR-044a).
const MinCheckpointInterval = 5 * time.Minute
