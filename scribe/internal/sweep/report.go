package sweep

import (
	"time"

	"github.com/google/uuid"
)

// SweepReportBuilder builds a sweep report from results.
type SweepReportBuilder struct {
	files     []string
	checksRun []CheckType
	results   []SweepResult
}

// NewSweepReportBuilder creates a new report builder.
func NewSweepReportBuilder() *SweepReportBuilder {
	return &SweepReportBuilder{
		files:     []string{},
		checksRun: []CheckType{},
		results:   []SweepResult{},
	}
}

// AddFile adds a file to the report.
func (b *SweepReportBuilder) AddFile(file string) {
	b.files = append(b.files, file)
}

// AddCheck adds a check type to the report.
func (b *SweepReportBuilder) AddCheck(check CheckType) {
	for _, c := range b.checksRun {
		if c == check {
			return // Already added
		}
	}
	b.checksRun = append(b.checksRun, check)
}

// AddResult adds a sweep result to the report.
func (b *SweepReportBuilder) AddResult(result SweepResult) {
	b.results = append(b.results, result)
}

// Build creates the final sweep report.
func (b *SweepReportBuilder) Build() SweepReport {
	summary := SweepSummary{
		TotalChecks: len(b.results),
	}

	for _, r := range b.results {
		switch r.Status {
		case SweepStatusOK:
			summary.OK++
		case SweepStatusWarning:
			summary.Warnings++
		case SweepStatusIssue:
			summary.Issues++
		}
	}

	return SweepReport{
		ReportID:     uuid.New().String(),
		GeneratedAt:  time.Now(),
		ContentFiles: b.files,
		ChecksRun:    b.checksRun,
		Summary:      summary,
		Results:      b.results,
	}
}

// FilterByStatus returns results with the given status.
func (r *SweepReport) FilterByStatus(status SweepResultStatus) []SweepResult {
	var filtered []SweepResult
	for _, result := range r.Results {
		if result.Status == status {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

// FilterByCheckType returns results with the given check type.
func (r *SweepReport) FilterByCheckType(checkType CheckType) []SweepResult {
	var filtered []SweepResult
	for _, result := range r.Results {
		if result.CheckType == checkType {
			filtered = append(filtered, result)
		}
	}
	return filtered
}
