package sweep

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"time"
)

// repoURLPattern matches GitHub repository URLs
var repoURLPattern = regexp.MustCompile(`https://github\.com/([^/]+)/([^/\s#)]+)`)

// ExtractRepoURLs extracts GitHub repository URLs from content.
func ExtractRepoURLs(content string) []string {
	matches := repoURLPattern.FindAllString(content, -1)
	seen := make(map[string]bool)
	var urls []string
	for _, url := range matches {
		if !seen[url] {
			seen[url] = true
			urls = append(urls, url)
		}
	}
	return urls
}

// CheckRepoFreshness checks if referenced repos are still maintained (FR-034).
// Repos not updated in > 2 years are flagged as stale.
func CheckRepoFreshness(content string, file string) []SweepResult {
	var results []SweepResult

	urls := ExtractRepoURLs(content)
	for _, url := range urls {
		result := checkSingleRepoFreshness(url, file)
		results = append(results, result)
	}

	return results
}

func checkSingleRepoFreshness(url string, file string) SweepResult {
	result := SweepResult{
		CheckType: CheckTypeRepoFreshness,
		Target:    url,
		File:      file,
		CheckedAt: time.Now(),
	}

	// Check if gh CLI is available
	if _, err := exec.LookPath("gh"); err != nil {
		result.Status = SweepStatusWarning
		result.Message = "Cannot check repo freshness: gh CLI not found"
		return result
	}

	// Parse owner/repo from URL
	matches := repoURLPattern.FindStringSubmatch(url)
	if len(matches) < 3 {
		result.Status = SweepStatusWarning
		result.Message = "Could not parse repository URL"
		return result
	}
	owner, repo := matches[1], matches[2]

	// Get repo info via gh api
	cmd := exec.Command("gh", "api", fmt.Sprintf("repos/%s/%s", owner, repo))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		result.Status = SweepStatusIssue
		result.Message = fmt.Sprintf("Failed to fetch repo info: %s", stderr.String())
		return result
	}

	var repoInfo struct {
		PushedAt string `json:"pushed_at"`
		Archived bool   `json:"archived"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &repoInfo); err != nil {
		result.Status = SweepStatusWarning
		result.Message = fmt.Sprintf("Failed to parse repo info: %v", err)
		return result
	}

	// Check if archived
	if repoInfo.Archived {
		result.Status = SweepStatusIssue
		result.Message = "Repository is archived"
		result.SuggestedFix = "Consider finding an active fork or alternative implementation"
		result.Details = map[string]any{
			"archived": true,
		}
		return result
	}

	// Check last push date
	pushedAt, err := time.Parse(time.RFC3339, repoInfo.PushedAt)
	if err != nil {
		result.Status = SweepStatusWarning
		result.Message = fmt.Sprintf("Could not parse push date: %v", err)
		return result
	}

	age := time.Since(pushedAt)
	if age > StaleThreshold {
		result.Status = SweepStatusIssue
		result.Message = fmt.Sprintf("Repository has not been updated in %.1f years", age.Hours()/(24*365))
		result.SuggestedFix = "Verify the code is still relevant or find an active alternative"
		result.Details = map[string]any{
			"last_push": pushedAt,
			"age_days":  int(age.Hours() / 24),
		}
		return result
	}

	result.Status = SweepStatusOK
	result.Message = fmt.Sprintf("Repository was updated %.0f days ago", age.Hours()/24)
	result.Details = map[string]any{
		"last_push": pushedAt,
	}
	return result
}
