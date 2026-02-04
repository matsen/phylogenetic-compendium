package verify

import (
	"testing"
)

func TestExtractCitations(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "single citation",
			content:  "See [@paper:tavare1986] for details.",
			expected: []string{"tavare1986"},
		},
		{
			name:     "multiple citations",
			content:  "Compare @paper:tavare1986 and @paper:felsenstein2004.",
			expected: []string{"tavare1986", "felsenstein2004"},
		},
		{
			name:     "no citations",
			content:  "This has no citations.",
			expected: []string{},
		},
		{
			name:     "duplicate citations",
			content:  "See @paper:abc123 and later @paper:abc123 again.",
			expected: []string{"abc123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractCitations(tt.content)
			if len(result) != len(tt.expected) {
				t.Errorf("got %d citations, want %d", len(result), len(tt.expected))
				return
			}
			for i, c := range result {
				if c != tt.expected[i] {
					t.Errorf("citation %d: got %q, want %q", i, c, tt.expected[i])
				}
			}
		})
	}
}

func TestExtractURLs(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "https url",
			content:  "See https://github.com/example/repo",
			expected: 1,
		},
		{
			name:     "http url",
			content:  "See http://example.com",
			expected: 1,
		},
		{
			name:     "multiple urls",
			content:  "See https://a.com and https://b.com",
			expected: 2,
		},
		{
			name:     "no urls",
			content:  "No URLs here",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractURLs(tt.content)
			if len(result) != tt.expected {
				t.Errorf("got %d URLs, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestExtractCodeLinks(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "single code link",
			content:  "See https://github.com/owner/repo/blob/abc123/file.go#L10-L20",
			expected: 1,
		},
		{
			name:     "single line link",
			content:  "See https://github.com/owner/repo/blob/abc123def456/path/to/file.c#L42",
			expected: 1,
		},
		{
			name:     "no code links",
			content:  "See https://github.com/owner/repo",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractCodeLinks(tt.content)
			if len(result) != tt.expected {
				t.Errorf("got %d code links, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestExtractTodoMarkers(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "TODO marker",
			content:  "TODO: fix this later",
			expected: 1,
		},
		{
			name:     "FIXME marker",
			content:  "FIXME: broken code",
			expected: 1,
		},
		{
			name:     "XXX marker",
			content:  "XXX: needs attention",
			expected: 1,
		},
		{
			name:     "HACK marker",
			content:  "HACK: temporary workaround",
			expected: 1,
		},
		{
			name:     "multiple markers",
			content:  "TODO: fix this\nFIXME: and this",
			expected: 2,
		},
		{
			name:     "no markers",
			content:  "Clean code here",
			expected: 0,
		},
		{
			name:     "case insensitive",
			content:  "todo: lowercase",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTodoMarkers(tt.content)
			if len(result) != tt.expected {
				t.Errorf("got %d markers, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestAnalyzeClaimWithHeuristics(t *testing.T) {
	tests := []struct {
		name          string
		sentence      string
		needsCitation bool
		confidence    string
	}{
		{
			name:          "performance comparison",
			sentence:      "Method A is 40% faster than method B.",
			needsCitation: true,
			confidence:    "high",
		},
		{
			name:          "discovery attribution",
			sentence:      "This algorithm was discovered by Felsenstein.",
			needsCitation: true,
			confidence:    "high",
		},
		{
			name:          "prior work reference",
			sentence:      "As shown by previous studies, this approach works.",
			needsCitation: true,
			confidence:    "high",
		},
		{
			name:          "definition",
			sentence:      "A phylogenetic tree is defined as a branching diagram.",
			needsCitation: false,
			confidence:    "high",
		},
		{
			name:          "example",
			sentence:      "For example, consider the case where n=10.",
			needsCitation: false,
			confidence:    "high",
		},
		{
			name:          "transitional",
			sentence:      "In this section, we discuss the algorithm.",
			needsCitation: false,
			confidence:    "high",
		},
		{
			name:          "causal claim",
			sentence:      "This causes the likelihood to decrease.",
			needsCitation: true,
			confidence:    "medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			needsCitation, confidence, _ := AnalyzeClaimWithHeuristics(tt.sentence)
			if needsCitation != tt.needsCitation {
				t.Errorf("needsCitation: got %v, want %v", needsCitation, tt.needsCitation)
			}
			if confidence != tt.confidence {
				t.Errorf("confidence: got %v, want %v", confidence, tt.confidence)
			}
		})
	}
}

func TestReportBuilder(t *testing.T) {
	builder := NewReportBuilder()
	builder.AddFile("test.qmd")

	// Add some results
	builder.AddResult(VerificationResult{
		CheckID:   "1",
		CheckType: CheckTypeCitation,
		Status:    CheckStatusPass,
	})
	builder.AddResult(VerificationResult{
		CheckID:   "2",
		CheckType: CheckTypeURL,
		Status:    CheckStatusFail,
	})
	builder.AddResult(VerificationResult{
		CheckID:   "3",
		CheckType: CheckTypeCodeLink,
		Status:    CheckStatusWarn,
	})

	report := builder.Build()

	if len(report.ContentFiles) != 1 {
		t.Errorf("got %d files, want 1", len(report.ContentFiles))
	}

	if report.Summary.TotalChecks != 3 {
		t.Errorf("got %d total checks, want 3", report.Summary.TotalChecks)
	}

	if report.Summary.Passed != 1 {
		t.Errorf("got %d passed, want 1", report.Summary.Passed)
	}

	if report.Summary.Failed != 1 {
		t.Errorf("got %d failed, want 1", report.Summary.Failed)
	}

	if report.Summary.Warnings != 1 {
		t.Errorf("got %d warnings, want 1", report.Summary.Warnings)
	}

	if report.ExitCode != 1 {
		t.Errorf("got exit code %d, want 1 (due to failure)", report.ExitCode)
	}
}

func TestReportFilterByStatus(t *testing.T) {
	report := VerificationReport{
		Results: []VerificationResult{
			{CheckID: "1", Status: CheckStatusPass},
			{CheckID: "2", Status: CheckStatusFail},
			{CheckID: "3", Status: CheckStatusPass},
			{CheckID: "4", Status: CheckStatusWarn},
		},
	}

	passed := report.FilterByStatus(CheckStatusPass)
	if len(passed) != 2 {
		t.Errorf("got %d passed, want 2", len(passed))
	}

	failed := report.FilterByStatus(CheckStatusFail)
	if len(failed) != 1 {
		t.Errorf("got %d failed, want 1", len(failed))
	}
}

func TestReportFilterByType(t *testing.T) {
	report := VerificationReport{
		Results: []VerificationResult{
			{CheckID: "1", CheckType: CheckTypeCitation},
			{CheckID: "2", CheckType: CheckTypeURL},
			{CheckID: "3", CheckType: CheckTypeCitation},
		},
	}

	citations := report.FilterByType(CheckTypeCitation)
	if len(citations) != 2 {
		t.Errorf("got %d citations, want 2", len(citations))
	}

	urls := report.FilterByType(CheckTypeURL)
	if len(urls) != 1 {
		t.Errorf("got %d urls, want 1", len(urls))
	}
}
