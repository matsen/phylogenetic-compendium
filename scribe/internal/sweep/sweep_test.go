package sweep

import (
	"testing"
)

func TestExtractRepoURLs(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "single repo",
			content:  "See https://github.com/owner/repo for details",
			expected: 1,
		},
		{
			name:     "multiple repos",
			content:  "Compare https://github.com/a/b and https://github.com/c/d",
			expected: 2,
		},
		{
			name:     "no repos",
			content:  "No GitHub URLs here",
			expected: 0,
		},
		{
			name:     "duplicate repos",
			content:  "See https://github.com/owner/repo and again https://github.com/owner/repo",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractRepoURLs(tt.content)
			if len(result) != tt.expected {
				t.Errorf("got %d URLs, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestSweepReportBuilder(t *testing.T) {
	builder := NewSweepReportBuilder()
	builder.AddFile("test.qmd")
	builder.AddCheck(CheckTypeRepoFreshness)

	builder.AddResult(SweepResult{
		CheckType: CheckTypeRepoFreshness,
		Status:    SweepStatusOK,
		Target:    "https://github.com/a/b",
	})
	builder.AddResult(SweepResult{
		CheckType: CheckTypeRepoFreshness,
		Status:    SweepStatusIssue,
		Target:    "https://github.com/c/d",
	})
	builder.AddResult(SweepResult{
		CheckType: CheckTypeCodeLinks,
		Status:    SweepStatusWarning,
		Target:    "https://github.com/e/f/blob/...",
	})

	report := builder.Build()

	if len(report.ContentFiles) != 1 {
		t.Errorf("got %d files, want 1", len(report.ContentFiles))
	}

	if report.Summary.TotalChecks != 3 {
		t.Errorf("got %d total checks, want 3", report.Summary.TotalChecks)
	}

	if report.Summary.OK != 1 {
		t.Errorf("got %d OK, want 1", report.Summary.OK)
	}

	if report.Summary.Issues != 1 {
		t.Errorf("got %d issues, want 1", report.Summary.Issues)
	}

	if report.Summary.Warnings != 1 {
		t.Errorf("got %d warnings, want 1", report.Summary.Warnings)
	}
}

func TestSweepReportFilter(t *testing.T) {
	report := SweepReport{
		Results: []SweepResult{
			{CheckType: CheckTypeRepoFreshness, Status: SweepStatusOK},
			{CheckType: CheckTypeRepoFreshness, Status: SweepStatusIssue},
			{CheckType: CheckTypeCodeLinks, Status: SweepStatusWarning},
		},
	}

	// Filter by status
	issues := report.FilterByStatus(SweepStatusIssue)
	if len(issues) != 1 {
		t.Errorf("got %d issues, want 1", len(issues))
	}

	// Filter by check type
	freshness := report.FilterByCheckType(CheckTypeRepoFreshness)
	if len(freshness) != 2 {
		t.Errorf("got %d freshness results, want 2", len(freshness))
	}
}

func TestCheckCoverageGaps(t *testing.T) {
	// Test with content that has low citation density
	content := `This is paragraph one. No citations here.

This is paragraph two. Still no citations.

This is paragraph three. Continuing without citations.

This is paragraph four. The content keeps going.

This is paragraph five. Almost done.

This is paragraph six. Final paragraph.`

	results := CheckCoverageGaps(content, "test.qmd")

	// Should detect low citation density
	hasWarning := false
	for _, r := range results {
		if r.Status == SweepStatusWarning && r.Message == "Low citation density detected" {
			hasWarning = true
			break
		}
	}

	if !hasWarning {
		t.Error("expected low citation density warning")
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if len(opts.Checks) == 0 {
		t.Error("expected default checks to be non-empty")
	}

	// Should include all check types
	expectedChecks := map[CheckType]bool{
		CheckTypeRepoFreshness:    false,
		CheckTypeCodeLinks:        false,
		CheckTypeClaimConsistency: false,
		CheckTypeCoverage:         false,
	}

	for _, check := range opts.Checks {
		expectedChecks[check] = true
	}

	for check, found := range expectedChecks {
		if !found {
			t.Errorf("expected check %s to be in default options", check)
		}
	}
}
