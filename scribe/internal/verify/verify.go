package verify

import (
	"bufio"
	"os"
	"strings"
)

// Options configures verification behavior.
type Options struct {
	UseLLM bool // Whether to use LLM for claim detection
}

// DefaultOptions returns the default verification options.
func DefaultOptions() Options {
	return Options{
		UseLLM: true,
	}
}

// VerifyFile runs all verification checks on a single file.
func VerifyFile(filePath string, opts Options) ([]VerificationResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	contentStr := string(content)
	var results []VerificationResult

	// Read file line by line for line-specific checks
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	var lines []string

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		lines = append(lines, line)

		// Check for TODO markers on each line
		if todoMarkerPattern.MatchString(line) {
			markers := todoMarkerPattern.FindAllString(line, -1)
			for _, marker := range markers {
				results = append(results, VerifyTodoMarker(marker, filePath, lineNum, line))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Extract and verify citations from full content
	citations := ExtractCitations(contentStr)
	for _, citation := range citations {
		// Find line number for this citation
		lineNum := findLineNumber(lines, "@paper:"+citation)
		lineText := ""
		if lineNum > 0 && lineNum <= len(lines) {
			lineText = lines[lineNum-1]
		}
		results = append(results, VerifyCitation(citation, filePath, lineNum, lineText))
	}

	// Extract and verify URLs
	urls := ExtractURLs(contentStr)
	for _, url := range urls {
		// Skip GitHub blob URLs (handled by code link checker)
		if strings.Contains(url, "github.com") && strings.Contains(url, "/blob/") {
			continue
		}
		lineNum := findLineNumber(lines, url)
		lineText := ""
		if lineNum > 0 && lineNum <= len(lines) {
			lineText = lines[lineNum-1]
		}
		results = append(results, VerifyURL(url, filePath, lineNum, lineText))
	}

	// Extract and verify code links
	codeLinks := ExtractCodeLinks(contentStr)
	for _, link := range codeLinks {
		lineNum := findLineNumber(lines, link.FullURL)
		lineText := ""
		if lineNum > 0 && lineNum <= len(lines) {
			lineText = lines[lineNum-1]
		}
		results = append(results, VerifyCodeLink(link, filePath, lineNum, lineText))
	}

	// Check for uncited claims (sentence by sentence)
	for i, line := range lines {
		lineNum := i + 1
		// Skip YAML frontmatter, code blocks, headers, and empty lines
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "```") ||
			strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		// Split line into sentences
		sentences := splitSentences(line)
		for _, sentence := range sentences {
			sentence = strings.TrimSpace(sentence)
			if len(sentence) < 20 { // Skip very short fragments
				continue
			}

			// Check if sentence has a citation nearby
			hasCitation := citationPattern.MatchString(sentence) ||
				(lineNum > 0 && lineNum <= len(lines) && citationPattern.MatchString(lines[lineNum-1]))

			result := VerifyClaim(sentence, filePath, lineNum, hasCitation, opts.UseLLM)
			// Only add failed claim checks to avoid noise
			if result.Status == CheckStatusFail {
				results = append(results, result)
			}
		}
	}

	return results, nil
}

// VerifyFiles runs verification on multiple files.
func VerifyFiles(files []string, opts Options) (*VerificationReport, error) {
	builder := NewReportBuilder()

	for _, file := range files {
		builder.AddFile(file)

		results, err := VerifyFile(file, opts)
		if err != nil {
			// Create a failed result for the file read error
			builder.AddResult(VerificationResult{
				CheckType: CheckTypeCitation,
				Target: VerificationTarget{
					File: file,
					Line: 0,
					Text: "",
				},
				Status:  CheckStatusFail,
				Message: err.Error(),
			})
			continue
		}

		for _, result := range results {
			builder.AddResult(result)
		}
	}

	report := builder.Build()
	return &report, nil
}

// findLineNumber finds the line number containing a substring.
func findLineNumber(lines []string, substr string) int {
	for i, line := range lines {
		if strings.Contains(line, substr) {
			return i + 1
		}
	}
	return 0
}

// splitSentences splits text into sentences (simple heuristic).
func splitSentences(text string) []string {
	// Simple sentence splitting on . ? !
	var sentences []string
	current := ""
	for _, r := range text {
		current += string(r)
		if r == '.' || r == '?' || r == '!' {
			s := strings.TrimSpace(current)
			if len(s) > 0 {
				sentences = append(sentences, s)
			}
			current = ""
		}
	}
	if s := strings.TrimSpace(current); len(s) > 0 {
		sentences = append(sentences, s)
	}
	return sentences
}
