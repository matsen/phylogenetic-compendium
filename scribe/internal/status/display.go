package status

import (
	"fmt"
	"time"

	"github.com/matsen/phylogenetic-compendium/scribe/internal/output"
)

// Display provides pretty-printing for status information.
type Display struct {
	formatter *output.Formatter
}

// NewDisplay creates a new Display.
func NewDisplay(jsonOutput bool) *Display {
	return &Display{
		formatter: output.NewFormatter(jsonOutput),
	}
}

// ShowCheckpoint displays a task checkpoint in human-readable format.
func (d *Display) ShowCheckpoint(checkpoint *TaskCheckpoint) {
	if d.formatter.IsJSON() {
		d.formatter.JSON(checkpoint)
		return
	}

	duration := time.Since(checkpoint.StartedAt)

	d.formatter.Header(fmt.Sprintf("Task: %s", checkpoint.TaskDescription))
	d.formatter.Println("Type: %s", checkpoint.TaskType)
	d.formatter.Println("ID: %s", checkpoint.TaskID)
	d.formatter.Println("")

	// Timing info
	d.formatter.Println("Running for: %s (iteration %d/%d)",
		output.FormatDuration(duration),
		checkpoint.IterationCount,
		checkpoint.MaxIterations)

	// Progress bar
	progress := 0
	total := len(checkpoint.State.ItemsCompleted) + len(checkpoint.State.ItemsPending)
	if total > 0 {
		progress = (len(checkpoint.State.ItemsCompleted) * 100) / total
	}
	d.formatter.Println("Progress: %s %d%%",
		output.ProgressBar(len(checkpoint.State.ItemsCompleted), total, 20),
		progress)
	d.formatter.Println("")

	// Item counts
	d.formatter.Println("Items: %d completed, %d pending, %d blocked",
		len(checkpoint.State.ItemsCompleted),
		len(checkpoint.State.ItemsPending),
		len(checkpoint.State.BlockedItems))

	if checkpoint.State.CurrentFocus != "" {
		d.formatter.Println("Current focus: %s", checkpoint.State.CurrentFocus)
	}
	d.formatter.Println("")

	// Metrics
	d.formatter.Println("Candidates queued: %d (%d papers, %d code locations)",
		checkpoint.Metrics.CandidatesQueued,
		checkpoint.Metrics.PapersFound,
		checkpoint.Metrics.CodeLocationsFound)
	d.formatter.Println("Repos searched: %d", checkpoint.Metrics.ReposSearched)
	d.formatter.Println("Estimated cost: %s", output.FormatCurrency(checkpoint.Metrics.EstimatedCostUSD))

	if checkpoint.CostBudgetUSD != nil {
		d.formatter.Println("Budget: %s (%.1f%% used)",
			output.FormatCurrency(*checkpoint.CostBudgetUSD),
			(checkpoint.Metrics.EstimatedCostUSD / *checkpoint.CostBudgetUSD) * 100)
	}
	d.formatter.Println("")

	// Blocked items
	if len(checkpoint.State.BlockedItems) > 0 {
		d.formatter.Header("Blocked Items")
		for _, item := range checkpoint.State.BlockedItems {
			d.formatter.Println("%s %s", output.FormatStatus(output.StatusWarning), item.Item)
			d.formatter.Println("   Reason: %s", item.Reason)
			d.formatter.Println("   Since: %s ago", output.FormatTimeSince(item.BlockedAt))
		}
	}

	// Last checkpoint time
	d.formatter.Println("")
	d.formatter.Println("Last checkpoint: %s ago", output.FormatTimeSince(checkpoint.LastCheckpoint))
}

// ShowNoTask displays a message when no task is running.
func (d *Display) ShowNoTask() {
	if d.formatter.IsJSON() {
		d.formatter.JSON(map[string]string{"status": "no_task"})
		return
	}
	d.formatter.Println("No autonomous task is currently running.")
	d.formatter.Println("")
	d.formatter.Println("To start a task, use claude with Ralph Loop:")
	d.formatter.Println("  claude --ralph \"$(cat scribe/agents/exploration/PROMPT.md)\"")
}
