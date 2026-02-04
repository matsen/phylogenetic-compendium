package sweep

import (
	"fmt"
	"time"

	"github.com/matsen/phylogenetic-compendium/scribe/internal/verify"
)

// CheckCodeLinks verifies that code location links are still valid (FR-035).
// Checks that files exist and line ranges are valid at HEAD.
func CheckCodeLinks(content string, file string) []SweepResult {
	var results []SweepResult

	links := verify.ExtractCodeLinks(content)
	for _, link := range links {
		result := checkSingleCodeLink(link, file)
		results = append(results, result)
	}

	return results
}

func checkSingleCodeLink(link verify.CodeLinkMatch, file string) SweepResult {
	result := SweepResult{
		CheckType: CheckTypeCodeLinks,
		Target:    link.FullURL,
		File:      file,
		CheckedAt: time.Now(),
		Details: map[string]any{
			"owner":      link.Owner,
			"repo":       link.Repo,
			"commit_sha": link.CommitSHA,
			"file_path":  link.FilePath,
			"start_line": link.StartLine,
			"end_line":   link.EndLine,
		},
	}

	// Use the verify package's code link checker
	verifyResult := verify.VerifyCodeLink(link, file, 0, "")

	switch verifyResult.Status {
	case verify.CheckStatusPass:
		result.Status = SweepStatusOK
		result.Message = "Code link is valid"
	case verify.CheckStatusFail:
		result.Status = SweepStatusIssue
		result.Message = verifyResult.Message
		result.SuggestedFix = "Update the permalink to point to the current location"
	case verify.CheckStatusWarn:
		result.Status = SweepStatusWarning
		result.Message = verifyResult.Message
	}

	return result
}

// CheckCodeLinksAtHead verifies code links against the HEAD of the repository.
// This catches cases where files have been moved or deleted since the permalink was created.
func CheckCodeLinksAtHead(content string, file string) []SweepResult {
	var results []SweepResult

	links := verify.ExtractCodeLinks(content)
	for _, link := range links {
		result := SweepResult{
			CheckType: CheckTypeCodeLinks,
			Target:    link.FullURL,
			File:      file,
			CheckedAt: time.Now(),
		}

		// This would check if the file exists at HEAD (not just at the permalink commit)
		// For now, we note that this needs manual verification
		result.Status = SweepStatusWarning
		result.Message = fmt.Sprintf("Code link points to commit %s - verify file still exists at HEAD", link.CommitSHA[:8])
		result.SuggestedFix = "Run `gh api repos/{owner}/{repo}/contents/{path}` to check if file exists at HEAD"
		result.Details = map[string]any{
			"commit_sha": link.CommitSHA,
			"file_path":  link.FilePath,
		}

		results = append(results, result)
	}

	return results
}
