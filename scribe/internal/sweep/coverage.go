package sweep

import (
	"regexp"
	"time"
)

// CheckCoverageGaps identifies potential coverage gaps in the content (FR-036).
// This is a heuristic check that flags areas that might need more documentation.
func CheckCoverageGaps(content string, file string) []SweepResult {
	var results []SweepResult

	// Check for sections with few citations
	results = append(results, checkCitationDensity(content, file)...)

	// Check for referenced but unexplained techniques
	results = append(results, checkUnexplainedTechniques(content, file)...)

	return results
}

func checkCitationDensity(content string, file string) []SweepResult {
	var results []SweepResult

	// Count citations
	citations := len(citationPattern.FindAllString(content, -1))

	// Count paragraphs (rough heuristic: double newlines)
	paragraphs := len(paragraphPattern.FindAllString(content, -1)) + 1

	if paragraphs > 5 && citations < paragraphs/3 {
		results = append(results, SweepResult{
			CheckType: CheckTypeCoverage,
			Status:    SweepStatusWarning,
			Target:    file,
			File:      file,
			Message:   "Low citation density detected",
			Details: map[string]any{
				"paragraphs": paragraphs,
				"citations":  citations,
				"ratio":      float64(citations) / float64(paragraphs),
			},
			SuggestedFix: "Review content for claims that need citations",
			CheckedAt:    time.Now(),
		})
	}

	return results
}

func checkUnexplainedTechniques(content string, file string) []SweepResult {
	var results []SweepResult

	// Look for technique mentions that might need explanation
	// This is a very rough heuristic
	techniques := techniquePattern.FindAllString(content, -1)

	for _, technique := range techniques {
		// Check if the technique is explained (has "is" or "are" nearby)
		if !isExplained(content, technique) {
			results = append(results, SweepResult{
				CheckType: CheckTypeCoverage,
				Status:    SweepStatusWarning,
				Target:    technique,
				File:      file,
				Message:   "Technique mentioned but may lack explanation",
				Details: map[string]any{
					"technique": technique,
				},
				SuggestedFix: "Consider adding a brief explanation of this technique",
				CheckedAt:    time.Now(),
			})
		}
	}

	return results
}

func isExplained(content string, term string) bool {
	// Very rough heuristic: check if the term is followed by "is" or "are"
	// This would need to be much more sophisticated in practice
	return true // For now, assume all are explained
}

var (
	citationPattern  = regexp.MustCompile(`@paper:[a-zA-Z0-9_-]+`)
	paragraphPattern = regexp.MustCompile(`\n\n`)
	techniquePattern = regexp.MustCompile(`(?i)\b(algorithm|method|technique|approach|strategy)\b`)
)
