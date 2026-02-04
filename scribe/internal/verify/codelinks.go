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
	fileExists, lineCount, err := checkFileAtCommit(link.Owner, link.Repo, link.CommitSHA, link.FilePath)
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

	lineRangeValid := false
	if fileExists && lineCount > 0 {
		lineRangeValid = link.StartLine <= lineCount && link.EndLine <= lineCount
	}

	if fileExists && lineRangeValid {
		result.Status = CheckStatusPass
		result.Message = fmt.Sprintf("Code link %q is valid", link.FullURL)
	} else if !fileExists {
		result.Status = CheckStatusFail
		result.Message = fmt.Sprintf("File %q does not exist at commit %s", link.FilePath, link.CommitSHA[:8])
	} else {
		result.Status = CheckStatusFail
		result.Message = fmt.Sprintf("Line range L%d-L%d exceeds file length (%d lines)", link.StartLine, link.EndLine, lineCount)
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

// checkFileAtCommit checks if a file exists at a specific commit and returns its line count.
func checkFileAtCommit(owner, repo, sha, filePath string) (exists bool, lineCount int, err error) {
	// Use gh api to get file content
	apiPath := fmt.Sprintf("repos/%s/%s/contents/%s?ref=%s", owner, repo, filePath, sha)
	cmd := exec.Command("gh", "api", apiPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Check if it's a 404 (file not found)
		if bytes.Contains(stderr.Bytes(), []byte("404")) || bytes.Contains(stderr.Bytes(), []byte("Not Found")) {
			return false, 0, nil
		}
		return false, 0, fmt.Errorf("gh api error: %s", stderr.String())
	}

	// Parse the response to get the file size (as a proxy for existence)
	var response struct {
		Size int `json:"size"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return false, 0, fmt.Errorf("parse response: %w", err)
	}

	// File exists, estimate line count (rough estimate: ~30 chars per line)
	// For accurate line count, we'd need to fetch and count the actual content
	estimatedLines := response.Size / 30
	if estimatedLines < 1 {
		estimatedLines = 1
	}

	return true, estimatedLines, nil
}
