package verify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// estimatedCharsPerLine is a rough estimate for calculating line count from file size.
// This is unreliable (minified code may have 1000+ chars/line, sparse code 10-20).
// Used only as a heuristic - line range validation is skipped when unreliable.
const estimatedCharsPerLine = 30

// githubPermalinkPattern matches GitHub permalink URLs with line numbers
var githubPermalinkPattern = regexp.MustCompile(`https://github\.com/([^/]+)/([^/]+)/blob/([a-f0-9]+)/([^#]+)#L(\d+)(?:-L(\d+))?`)

// CodeLinkMatch represents a parsed GitHub permalink.
type CodeLinkMatch struct {
	FullURL   string
	Owner     string
	Repo      string
	CommitSHA string
	FilePath  string
	StartLine int
	EndLine   int
}

// ExtractCodeLinks extracts GitHub permalinks from content.
func ExtractCodeLinks(content string) []CodeLinkMatch {
	matches := githubPermalinkPattern.FindAllStringSubmatch(content, -1)
	var links []CodeLinkMatch
	for _, m := range matches {
		if len(m) < 6 {
			continue
		}
		startLine, _ := strconv.Atoi(m[5])
		endLine := startLine
		if len(m) > 6 && m[6] != "" {
			endLine, _ = strconv.Atoi(m[6])
		}
		links = append(links, CodeLinkMatch{
			FullURL:   m[0],
			Owner:     m[1],
			Repo:      m[2],
			CommitSHA: m[3],
			FilePath:  m[4],
			StartLine: startLine,
			EndLine:   endLine,
		})
	}
	return links
}

// VerifyCodeLink checks if a GitHub permalink points to valid code.
func VerifyCodeLink(link CodeLinkMatch, file string, line int, text string) VerificationResult {
	result := VerificationResult{
		CheckID:   uuid.New().String(),
		CheckType: CheckTypeCodeLink,
		Target: VerificationTarget{
			File: file,
			Line: line,
			Text: text,
		},
		CheckedAt: time.Now(),
	}

	// Check if gh CLI is available
	if _, err := exec.LookPath("gh"); err != nil {
		result.Status = CheckStatusWarn
		result.Message = fmt.Sprintf("Cannot verify code link %q: gh CLI not found", link.FullURL)
		result.Details = VerificationDetails{
			CodeLink: &CodeLinkDetails{
				Permalink:      link.FullURL,
				FileExists:     false,
				LineRangeValid: false,
			},
		}
		return result
	}

	// Check if the file exists at the given commit
	fileExists, estimatedLineCount, isEstimated, err := checkFileExistsAtCommit(link.Owner, link.Repo, link.CommitSHA, link.FilePath)
	if err != nil {
		result.Status = CheckStatusFail
		result.Message = fmt.Sprintf("Failed to verify code link %q: %v", link.FullURL, err)
		result.Details = VerificationDetails{
			CodeLink: &CodeLinkDetails{
				Permalink:      link.FullURL,
				FileExists:     false,
				LineRangeValid: false,
			},
		}
		return result
	}

	// Determine line range validity
	// If line count is estimated (unreliable), skip line range validation and warn
	lineRangeValid := false
	lineRangeSkipped := false
	if fileExists {
		if isEstimated {
			// Can't reliably validate line ranges with estimated counts
			lineRangeSkipped = true
		} else if estimatedLineCount > 0 {
			lineRangeValid = link.StartLine <= estimatedLineCount && link.EndLine <= estimatedLineCount
		}
	}

	if !fileExists {
		result.Status = CheckStatusFail
		result.Message = fmt.Sprintf("File %q does not exist at commit %s", link.FilePath, link.CommitSHA[:8])
	} else if lineRangeSkipped {
		// File exists but we can't verify line range - warn instead of false positive
		result.Status = CheckStatusWarn
		result.Message = fmt.Sprintf("Code link %q: file exists but line range L%d-L%d could not be verified (line count estimated)", link.FullURL, link.StartLine, link.EndLine)
		lineRangeValid = true // Mark as valid for details since we couldn't check
	} else if !lineRangeValid {
		result.Status = CheckStatusFail
		result.Message = fmt.Sprintf("Line range L%d-L%d exceeds file length (%d lines)", link.StartLine, link.EndLine, estimatedLineCount)
	} else {
		result.Status = CheckStatusPass
		result.Message = fmt.Sprintf("Code link %q is valid", link.FullURL)
	}

	result.Details = VerificationDetails{
		CodeLink: &CodeLinkDetails{
			Permalink:      link.FullURL,
			FileExists:     fileExists,
			LineRangeValid: lineRangeValid,
		},
	}

	return result
}

// checkFileExistsAtCommit checks if a file exists at a specific commit.
// Returns: exists, estimatedLineCount, isEstimated (true if line count is unreliable), error.
// Line count is estimated from file size and is unreliable - callers should treat it as a hint only.
func checkFileExistsAtCommit(owner, repo, sha, filePath string) (exists bool, lineCount int, isEstimated bool, err error) {
	// Use gh api to get file content
	apiPath := fmt.Sprintf("repos/%s/%s/contents/%s?ref=%s", owner, repo, filePath, sha)
	cmd := exec.Command("gh", "api", apiPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Check if it's a 404 (file not found)
		if bytes.Contains(stderr.Bytes(), []byte("404")) || bytes.Contains(stderr.Bytes(), []byte("Not Found")) {
			return false, 0, false, nil
		}
		return false, 0, false, fmt.Errorf("gh api error: %s", stderr.String())
	}

	// Parse the response to get the file size (as a proxy for existence)
	var response struct {
		Size int `json:"size"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return false, 0, false, fmt.Errorf("parse response: %w", err)
	}

	// File exists, estimate line count from size
	// This is unreliable - minified files may have very few lines, sparse code many
	estimatedLines := response.Size / estimatedCharsPerLine
	if estimatedLines < 1 {
		estimatedLines = 1
	}

	// Always mark as estimated since we're using a heuristic
	return true, estimatedLines, true, nil
}
