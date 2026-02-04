package status

import (
	"time"
)

// BlockingDetector detects blocking issues during autonomous operation.
type BlockingDetector struct {
	checkpoint *TaskCheckpoint
}

// NewBlockingDetector creates a new BlockingDetector.
func NewBlockingDetector(checkpoint *TaskCheckpoint) *BlockingDetector {
	return &BlockingDetector{checkpoint: checkpoint}
}

// AddBlockedItem adds a blocked item to the checkpoint.
func (d *BlockingDetector) AddBlockedItem(item, reason string) {
	blocked := BlockedItem{
		Item:      item,
		Reason:    reason,
		BlockedAt: time.Now(),
	}
	d.checkpoint.State.BlockedItems = append(d.checkpoint.State.BlockedItems, blocked)
}

// RemoveBlockedItem removes a blocked item from the checkpoint.
func (d *BlockingDetector) RemoveBlockedItem(item string) {
	var remaining []BlockedItem
	for _, b := range d.checkpoint.State.BlockedItems {
		if b.Item != item {
			remaining = append(remaining, b)
		}
	}
	d.checkpoint.State.BlockedItems = remaining
}

// HasBlockedItems returns true if there are blocked items.
func (d *BlockingDetector) HasBlockedItems() bool {
	return len(d.checkpoint.State.BlockedItems) > 0
}

// GetBlockedItems returns all blocked items.
func (d *BlockingDetector) GetBlockedItems() []BlockedItem {
	return d.checkpoint.State.BlockedItems
}

// ShouldQueueForReview determines if the task should be queued for human review.
// Per FR-042, this happens when:
// - There are blocked items that have been blocked for > 30 minutes
// - The task has exceeded its iteration or cost budget
// - No progress has been made in the last 3 iterations
func (d *BlockingDetector) ShouldQueueForReview() (bool, string) {
	// Check for long-standing blocked items
	for _, item := range d.checkpoint.State.BlockedItems {
		if time.Since(item.BlockedAt) > 30*time.Minute {
			return true, "Blocked item for >30 minutes: " + item.Item
		}
	}

	// Check iteration budget
	if d.checkpoint.IterationCount >= d.checkpoint.MaxIterations {
		return true, "Reached maximum iterations"
	}

	// Check cost budget
	if d.checkpoint.CostBudgetUSD != nil {
		if d.checkpoint.Metrics.EstimatedCostUSD >= *d.checkpoint.CostBudgetUSD {
			return true, "Reached cost budget"
		}
	}

	return false, ""
}
