// Package sweep implements periodic verification sweep for the scribe CLI.
package sweep

import "time"

// CheckType represents the type of sweep check.
type CheckType string

const (
	CheckTypeClaimConsistency CheckType = "claim-consistency"
	CheckTypeRepoFreshness    CheckType = "repo-freshness"
	CheckTypeCodeLinks        CheckType = "code-links"
	CheckTypeCoverage         CheckType = "coverage"
)

// SweepResultStatus represents the status of a sweep check.
type SweepResultStatus string

const (
	SweepStatusOK      SweepResultStatus = "ok"
	SweepStatusWarning SweepResultStatus = "warning"
	SweepStatusIssue   SweepResultStatus = "issue"
)

// SweepResult represents the result of a single sweep check.
type SweepResult struct {
	CheckType   CheckType         `json:"check_type"`
	Status      SweepResultStatus `json:"status"`
	Target      string            `json:"target"`
	File        string            `json:"file"`
	Line        int               `json:"line,omitempty"`
	Message     string            `json:"message"`
	Details     map[string]any    `json:"details,omitempty"`
	SuggestedFix string           `json:"suggested_fix,omitempty"`
	CheckedAt   time.Time         `json:"checked_at"`
}

// SweepSummary contains aggregated sweep statistics.
type SweepSummary struct {
	TotalChecks  int `json:"total_checks"`
	OK           int `json:"ok"`
	Warnings     int `json:"warnings"`
	Issues       int `json:"issues"`
}

// SweepReport is the aggregated output of a sweep.
type SweepReport struct {
	ReportID     string        `json:"report_id"`
	GeneratedAt  time.Time     `json:"generated_at"`
	ContentFiles []string      `json:"content_files"`
	ChecksRun    []CheckType   `json:"checks_run"`
	Summary      SweepSummary  `json:"summary"`
	Results      []SweepResult `json:"results"`
}

// StaleThreshold is the age after which a repository is considered stale.
const StaleThreshold = 2 * 365 * 24 * time.Hour // 2 years
