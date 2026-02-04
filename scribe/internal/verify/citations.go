package verify

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// citationPattern matches citations like [@paper:id] or @paper:id
var citationPattern = regexp.MustCompile(`@paper:([a-zA-Z0-9_-]+)`)

// ExtractCitations extracts all paper citations from content.
func ExtractCitations(content string) []string {
	matches := citationPattern.FindAllStringSubmatch(content, -1)
	seen := make(map[string]bool)
	var citations []string
	for _, m := range matches {
		if len(m) > 1 && !seen[m[1]] {
			seen[m[1]] = true
			citations = append(citations, m[1])
		}
	}
	return citations
}

// VerifyCitation checks if a paper ID exists in bipartite.
func VerifyCitation(paperID string, file string, line int, text string) VerificationResult {
	result := VerificationResult{
		CheckID:   uuid.New().String(),
		CheckType: CheckTypeCitation,
		Target: VerificationTarget{
			File: file,
			Line: line,
			Text: text,
		},
		CheckedAt: time.Now(),
	}

	// Try to look up the paper in bipartite
	resolved, err := lookupPaperInBipartite(paperID)
	if err != nil {
		result.Status = CheckStatusFail
		result.Message = fmt.Sprintf("Failed to verify paper ID %q: %v", paperID, err)
		result.Details = VerificationDetails{
			Citation: &CitationDetails{
				PaperID:  paperID,
				Resolved: false,
			},
		}
		return result
	}

	if resolved {
		result.Status = CheckStatusPass
		result.Message = fmt.Sprintf("Paper ID %q resolved successfully", paperID)
	} else {
		result.Status = CheckStatusFail
		result.Message = fmt.Sprintf("Paper ID %q not found in bipartite", paperID)
	}

	result.Details = VerificationDetails{
		Citation: &CitationDetails{
			PaperID:  paperID,
			Resolved: resolved,
		},
	}

	return result
}

// lookupPaperInBipartite checks if a paper exists in the bipartite library.
func lookupPaperInBipartite(paperID string) (bool, error) {
	// Check if bip CLI is available
	if _, err := exec.LookPath("bip"); err != nil {
		return false, fmt.Errorf("bip CLI not found: %w", err)
	}

	// Try to get the paper from bipartite
	cmd := exec.Command("bip", "s2", "get", paperID)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Check if it's a "not found" error vs other error
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "not found") || strings.Contains(stderrStr, "no such") {
			return false, nil
		}
		return false, fmt.Errorf("bip lookup failed: %s", stderrStr)
	}

	return true, nil
}
