// Package verify implements content verification for the scribe CLI.
package verify

import "time"

// CodeLocation represents a specific location in a codebase.
type CodeLocation struct {
	RepoURL      string  `json:"repo_url"`
	FilePath     string  `json:"file_path"`
	StartLine    int     `json:"start_line"`
	EndLine      int     `json:"end_line"`
	CommitSHA    string  `json:"commit_sha"`
	PermalinkURL string  `json:"permalink_url"`
	FunctionName *string `json:"function_name,omitempty"`
	Description  string  `json:"description"`
}

// GeneratePermalink generates a GitHub permalink URL from the code location fields.
func (c *CodeLocation) GeneratePermalink() string {
	// Format: https://github.com/{org}/{repo}/blob/{commit_sha}/{file_path}#L{start_line}-L{end_line}
	return c.PermalinkURL
}

// CheckType represents the type of verification check.
type CheckType string

const (
	CheckTypeCitation   CheckType = "citation"
	CheckTypeURL        CheckType = "url"
	CheckTypeCodeLink   CheckType = "code-link"
	CheckTypeClaim      CheckType = "claim"
	CheckTypeTodoMarker CheckType = "todo-marker"
)

// CheckStatus represents the outcome of a verification check.
type CheckStatus string

const (
	CheckStatusPass CheckStatus = "pass"
	CheckStatusFail CheckStatus = "fail"
	CheckStatusWarn CheckStatus = "warn"
)

// VerificationTarget identifies the source of a verification issue.
type VerificationTarget struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Text string `json:"text"`
}

// CitationDetails contains details specific to citation checks.
type CitationDetails struct {
	PaperID  string `json:"paper_id"`
	Resolved bool   `json:"resolved"`
}

// URLDetails contains details specific to URL checks.
type URLDetails struct {
	URL        string  `json:"url"`
	HTTPStatus *int    `json:"http_status,omitempty"`
	Error      *string `json:"error,omitempty"`
}

// CodeLinkDetails contains details specific to code link checks.
type CodeLinkDetails struct {
	Permalink      string `json:"permalink"`
	FileExists     bool   `json:"file_exists"`
	LineRangeValid bool   `json:"line_range_valid"`
}

// ClaimDetails contains details specific to claim checks.
type ClaimDetails struct {
	ClaimText       string `json:"claim_text"`
	Confidence      string `json:"confidence"` // high, medium, low
	SuggestedAction string `json:"suggested_action"`
}

// VerificationDetails contains type-specific check details.
type VerificationDetails struct {
	Citation *CitationDetails `json:"citation,omitempty"`
	URL      *URLDetails      `json:"url,omitempty"`
	CodeLink *CodeLinkDetails `json:"code_link,omitempty"`
	Claim    *ClaimDetails    `json:"claim,omitempty"`
}

// VerificationResult represents the outcome of a single verification check.
type VerificationResult struct {
	CheckID   string              `json:"check_id"`
	CheckType CheckType           `json:"check_type"`
	Target    VerificationTarget  `json:"target"`
	Status    CheckStatus         `json:"status"`
	Message   string              `json:"message"`
	Details   VerificationDetails `json:"details"`
	CheckedAt time.Time           `json:"checked_at"`
}

// ReportSummary contains aggregated verification statistics.
type ReportSummary struct {
	TotalChecks int `json:"total_checks"`
	Passed      int `json:"passed"`
	Failed      int `json:"failed"`
	Warnings    int `json:"warnings"`
}

// VerificationReport is the aggregated output of verify-content.
type VerificationReport struct {
	ReportID     string               `json:"report_id"`
	GeneratedAt  time.Time            `json:"generated_at"`
	ContentFiles []string             `json:"content_files"`
	Summary      ReportSummary        `json:"summary"`
	Results      []VerificationResult `json:"results"`
	ExitCode     int                  `json:"exit_code"`
}
