// Package llm provides LLM integration for NLP tasks.
package llm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Client provides access to LLM capabilities.
type Client struct {
	provider string // "claude" or "ollama"
	model    string // model name for ollama
}

// NewClient creates a new LLM client.
// It prefers Claude CLI if available, falling back to local Ollama.
func NewClient() (*Client, error) {
	// Check if claude CLI is available
	if _, err := exec.LookPath("claude"); err == nil {
		return &Client{provider: "claude"}, nil
	}

	// Check if ollama is available
	if _, err := exec.LookPath("ollama"); err == nil {
		return &Client{provider: "ollama", model: "llama3.2"}, nil
	}

	return nil, errors.New("no LLM provider available: install claude CLI or ollama")
}

// NewClientWithProvider creates a client with a specific provider.
func NewClientWithProvider(provider, model string) *Client {
	return &Client{provider: provider, model: model}
}

// IsAvailable returns true if an LLM provider is available.
func IsAvailable() bool {
	_, err := NewClient()
	return err == nil
}

// Complete sends a prompt to the LLM and returns the response.
func (c *Client) Complete(prompt string) (string, error) {
	switch c.provider {
	case "claude":
		return c.completeClaude(prompt)
	case "ollama":
		return c.completeOllama(prompt)
	default:
		return "", fmt.Errorf("unknown provider: %s", c.provider)
	}
}

// completeClaude uses the claude CLI to complete a prompt.
func (c *Client) completeClaude(prompt string) (string, error) {
	cmd := exec.Command("claude", "-p", prompt, "--model", "claude-haiku-4-20250514")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("claude CLI error: %w: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// completeOllama uses local Ollama to complete a prompt.
func (c *Client) completeOllama(prompt string) (string, error) {
	cmd := exec.Command("ollama", "run", c.model, prompt)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ollama error: %w: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

// ClaimAnalysisResult represents the result of analyzing a sentence for citation need.
type ClaimAnalysisResult struct {
	NeedsCitation   bool   `json:"needs_citation"`
	Confidence      string `json:"confidence"` // high, medium, low
	Reason          string `json:"reason"`
	SuggestedAction string `json:"suggested_action"`
}

// AnalyzeClaim determines if a sentence is a factual claim that needs citation.
func (c *Client) AnalyzeClaim(sentence string) (*ClaimAnalysisResult, error) {
	prompt := fmt.Sprintf(`Analyze if this sentence from a scientific document is a factual claim that requires a citation.

Sentence: %q

Respond with JSON only:
{
  "needs_citation": true/false,
  "confidence": "high"/"medium"/"low",
  "reason": "brief explanation",
  "suggested_action": "add citation" or "no action needed" or "review manually"
}

Guidelines:
- Performance comparisons ("X is faster than Y") NEED citations
- Attribution of discoveries ("discovered by", "introduced by") NEED citations
- Quantitative claims (numbers, percentages) NEED citations
- Definitions ("is defined as") do NOT need citations
- Examples ("for example", "e.g.") do NOT need citations
- Transitional prose ("in this section") does NOT need citations`, sentence)

	response, err := c.Complete(prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM analysis failed: %w", err)
	}

	// Extract JSON from response (it might have markdown code fences)
	jsonStr := extractJSON(response)

	var result ClaimAnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("parse LLM response: %w", err)
	}

	return &result, nil
}

// extractJSON extracts JSON from a response that might have markdown fences.
func extractJSON(s string) string {
	s = strings.TrimSpace(s)

	// Remove markdown code fences if present
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
		if idx := strings.Index(s, "```"); idx != -1 {
			s = s[:idx]
		}
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		if idx := strings.Index(s, "```"); idx != -1 {
			s = s[:idx]
		}
	}

	return strings.TrimSpace(s)
}
