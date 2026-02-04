package verify

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/matsen/phylogenetic-compendium/scribe/internal/llm"
)

// todoMarkerPattern matches TODO, FIXME, XXX, HACK markers
var todoMarkerPattern = regexp.MustCompile(`(?i)\b(TODO|FIXME|XXX|HACK)\b:?\s*`)

// ExtractTodoMarkers extracts TODO/FIXME markers from content.
func ExtractTodoMarkers(content string) []string {
	return todoMarkerPattern.FindAllString(content, -1)
}

// VerifyTodoMarker creates a verification result for a TODO marker.
func VerifyTodoMarker(marker string, file string, line int, text string) VerificationResult {
	return VerificationResult{
		CheckID:   uuid.New().String(),
		CheckType: CheckTypeTodoMarker,
		Target: VerificationTarget{
			File: file,
			Line: line,
			Text: text,
		},
		Status:    CheckStatusFail,
		Message:   "TODO/FIXME marker found - remove before publishing",
		CheckedAt: time.Now(),
	}
}

// Pattern-based claim detection heuristics
var (
	// mustCitePatterns - high confidence claims that need citations
	mustCitePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(\d+%|\d+x|faster|slower|better|worse)\s+(than|compared to)`),
		regexp.MustCompile(`(?i)\b(discovered|introduced|invented|developed)\s+by`),
		regexp.MustCompile(`(?i)\b(as shown|as described|according to|as demonstrated)\s+(in|by)`),
		regexp.MustCompile(`(?i)\bstudies\s+(have\s+)?(shown|demonstrated|found|revealed)`),
	}

	// shouldCitePatterns - medium confidence claims
	shouldCitePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(causes?|leads?\s+to|results?\s+in)\b`),
		regexp.MustCompile(`(?i)\b(O\(n|O\(log|complexity\s+of)`),
		regexp.MustCompile(`(?i)\b(historically|traditionally|originally)\b`),
	}

	// exemptPatterns - sentences that don't need citations
	exemptPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(is\s+defined\s+as|refers?\s+to|means?)\b`),
		regexp.MustCompile(`(?i)\b(for\s+example|e\.g\.|such\s+as|i\.e\.)\b`),
		regexp.MustCompile(`(?i)\b(in\s+this\s+(section|chapter)|we\s+now|let\s+us)\b`),
		regexp.MustCompile("^```"), // Code block
	}
)

// AnalyzeClaimWithHeuristics uses pattern matching to detect claims.
func AnalyzeClaimWithHeuristics(sentence string) (needsCitation bool, confidence string, reason string) {
	sentence = strings.TrimSpace(sentence)

	// Check exempt patterns first
	for _, p := range exemptPatterns {
		if p.MatchString(sentence) {
			return false, "high", "Sentence is a definition, example, or transitional prose"
		}
	}

	// Check must-cite patterns
	for _, p := range mustCitePatterns {
		if p.MatchString(sentence) {
			return true, "high", "Sentence contains comparison, attribution, or cites prior work"
		}
	}

	// Check should-cite patterns
	for _, p := range shouldCitePatterns {
		if p.MatchString(sentence) {
			return true, "medium", "Sentence contains causal claim or complexity statement"
		}
	}

	return false, "low", "No citation-requiring patterns detected"
}

// VerifyClaim checks if a sentence is an uncited factual claim.
func VerifyClaim(sentence string, file string, line int, hasCitation bool, useLLM bool) VerificationResult {
	result := VerificationResult{
		CheckID:   uuid.New().String(),
		CheckType: CheckTypeClaim,
		Target: VerificationTarget{
			File: file,
			Line: line,
			Text: sentence,
		},
		CheckedAt: time.Now(),
	}

	// If sentence already has a citation, it passes
	if hasCitation {
		result.Status = CheckStatusPass
		result.Message = "Claim has citation"
		result.Details = VerificationDetails{
			Claim: &ClaimDetails{
				ClaimText:       sentence,
				Confidence:      "high",
				SuggestedAction: "no action needed",
			},
		}
		return result
	}

	// First try pattern-based heuristics
	needsCitation, confidence, reason := AnalyzeClaimWithHeuristics(sentence)

	// If heuristics are inconclusive and LLM is available, use it
	if confidence == "low" && useLLM && llm.IsAvailable() {
		client, err := llm.NewClient()
		if err == nil {
			analysis, err := client.AnalyzeClaim(sentence)
			if err == nil {
				needsCitation = analysis.NeedsCitation
				confidence = analysis.Confidence
				reason = analysis.Reason
			}
		}
	}

	if needsCitation {
		result.Status = CheckStatusFail
		result.Message = "Uncited factual claim detected"
		result.Details = VerificationDetails{
			Claim: &ClaimDetails{
				ClaimText:       sentence,
				Confidence:      confidence,
				SuggestedAction: "add citation",
			},
		}
	} else {
		result.Status = CheckStatusPass
		result.Message = reason
		result.Details = VerificationDetails{
			Claim: &ClaimDetails{
				ClaimText:       sentence,
				Confidence:      confidence,
				SuggestedAction: "no action needed",
			},
		}
	}

	return result
}
