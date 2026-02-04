package verify

import (
	"time"

	"github.com/google/uuid"
)

// ReportBuilder builds a verification report from results.
type ReportBuilder struct {
	files   []string
	results []VerificationResult
}

// NewReportBuilder creates a new report builder.
func NewReportBuilder() *ReportBuilder {
	return &ReportBuilder{
		files:   []string{},
		results: []VerificationResult{},
	}
}

// AddFile adds a file to the report.
func (b *ReportBuilder) AddFile(file string) {
	b.files = append(b.files, file)
}

// AddResult adds a verification result to the report.
func (b *ReportBuilder) AddResult(result VerificationResult) {
	b.results = append(b.results, result)
}

// Build creates the final verification report.
func (b *ReportBuilder) Build() VerificationReport {
	summary := ReportSummary{
		TotalChecks: len(b.results),
	}

	for _, r := range b.results {
		switch r.Status {
		case CheckStatusPass:
			summary.Passed++
		case CheckStatusFail:
			summary.Failed++
		case CheckStatusWarn:
			summary.Warnings++
		}
	}

	exitCode := 0
	if summary.Failed > 0 {
		exitCode = 1
	}

	return VerificationReport{
		ReportID:     uuid.New().String(),
		GeneratedAt:  time.Now(),
		ContentFiles: b.files,
		Summary:      summary,
		Results:      b.results,
		ExitCode:     exitCode,
	}
}

// FilterByStatus returns results with the given status.
func (r *VerificationReport) FilterByStatus(status CheckStatus) []VerificationResult {
	var filtered []VerificationResult
	for _, result := range r.Results {
		if result.Status == status {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

// FilterByType returns results with the given check type.
func (r *VerificationReport) FilterByType(checkType CheckType) []VerificationResult {
	var filtered []VerificationResult
	for _, result := range r.Results {
		if result.CheckType == checkType {
			filtered = append(filtered, result)
		}
	}
	return filtered
}
