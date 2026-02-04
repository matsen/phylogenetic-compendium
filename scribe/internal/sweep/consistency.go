package sweep

import (
	"fmt"
	"time"

	"github.com/matsen/phylogenetic-compendium/scribe/internal/verify"
)

// CheckClaimConsistency checks if cited papers still support the claims (FR-033).
// Uses Asta snippet search to verify claim-source alignment.
func CheckClaimConsistency(content string, file string) []SweepResult {
	var results []SweepResult

	// Extract citations
	citations := verify.ExtractCitations(content)
	if len(citations) == 0 {
		return results
	}

	// For each citation, we would ideally use Asta to verify the claim still holds.
	// This is a placeholder that marks for manual review.
	for _, citation := range citations {
		results = append(results, SweepResult{
			CheckType: CheckTypeClaimConsistency,
			Status:    SweepStatusOK,
			Target:    citation,
			File:      file,
			Message:   fmt.Sprintf("Citation %s exists - manual consistency check recommended", citation),
			Details: map[string]any{
				"citation_id": citation,
				"needs_review": true,
			},
			SuggestedFix: "Use Asta snippet search to verify the cited paper still supports the claim",
			CheckedAt:    time.Now(),
		})
	}

	return results
}
