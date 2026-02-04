package verify

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// urlPattern matches HTTP(S) URLs
var urlPattern = regexp.MustCompile(`https?://[^\s\)\]>]+`)

// ExtractURLs extracts all URLs from content.
func ExtractURLs(content string) []string {
	return urlPattern.FindAllString(content, -1)
}

// VerifyURL checks if a URL is accessible.
func VerifyURL(url string, file string, line int, text string) VerificationResult {
	result := VerificationResult{
		CheckID:   uuid.New().String(),
		CheckType: CheckTypeURL,
		Target: VerificationTarget{
			File: file,
			Line: line,
			Text: text,
		},
		CheckedAt: time.Now(),
	}

	// Check URL accessibility with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		result.Status = CheckStatusFail
		errStr := err.Error()
		result.Message = fmt.Sprintf("Invalid URL %q: %v", url, err)
		result.Details = VerificationDetails{
			URL: &URLDetails{
				URL:   url,
				Error: &errStr,
			},
		}
		return result
	}

	// Use a custom client with reasonable timeouts
	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow up to 10 redirects
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		// Try GET if HEAD fails (some servers don't support HEAD)
		req, _ = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err = client.Do(req)
		if err != nil {
			result.Status = CheckStatusFail
			errStr := err.Error()
			result.Message = fmt.Sprintf("URL %q is not accessible: %v", url, err)
			result.Details = VerificationDetails{
				URL: &URLDetails{
					URL:   url,
					Error: &errStr,
				},
			}
			return result
		}
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode

	if statusCode >= 200 && statusCode < 400 {
		result.Status = CheckStatusPass
		result.Message = fmt.Sprintf("URL %q is accessible (HTTP %d)", url, statusCode)
	} else if statusCode >= 400 && statusCode < 500 {
		result.Status = CheckStatusFail
		result.Message = fmt.Sprintf("URL %q returned client error (HTTP %d)", url, statusCode)
	} else {
		result.Status = CheckStatusWarn
		result.Message = fmt.Sprintf("URL %q returned unexpected status (HTTP %d)", url, statusCode)
	}

	result.Details = VerificationDetails{
		URL: &URLDetails{
			URL:        url,
			HTTPStatus: &statusCode,
		},
	}

	return result
}
