// Package output provides common CLI output utilities.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

// Formatter handles output formatting for CLI commands.
type Formatter struct {
	writer     io.Writer
	jsonOutput bool
}

// NewFormatter creates a new Formatter.
func NewFormatter(jsonOutput bool) *Formatter {
	return &Formatter{
		writer:     os.Stdout,
		jsonOutput: jsonOutput,
	}
}

// NewFormatterWithWriter creates a Formatter with a custom writer.
func NewFormatterWithWriter(w io.Writer, jsonOutput bool) *Formatter {
	return &Formatter{
		writer:     w,
		jsonOutput: jsonOutput,
	}
}

// IsJSON returns true if the formatter is in JSON mode.
func (f *Formatter) IsJSON() bool {
	return f.jsonOutput
}

// JSON outputs data as JSON.
func (f *Formatter) JSON(v any) error {
	enc := json.NewEncoder(f.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// Print outputs a formatted string.
func (f *Formatter) Print(format string, args ...any) {
	fmt.Fprintf(f.writer, format, args...)
}

// Println outputs a formatted string with a newline.
func (f *Formatter) Println(format string, args ...any) {
	fmt.Fprintf(f.writer, format+"\n", args...)
}

// Header outputs a section header.
func (f *Formatter) Header(title string) {
	f.Println("\n%s", title)
	f.Println(strings.Repeat("─", len(title)))
}

// Table creates a new table writer for aligned columns.
func (f *Formatter) Table() *tabwriter.Writer {
	return tabwriter.NewWriter(f.writer, 0, 0, 2, ' ', 0)
}

// FormatDuration formats a duration in human-readable form.
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		secs := int(d.Seconds()) % 60
		if secs > 0 {
			return fmt.Sprintf("%dm %ds", mins, secs)
		}
		return fmt.Sprintf("%dm", mins)
	}
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	if mins > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dh", hours)
}

// FormatTime formats a time in a human-readable form.
func FormatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02 15:04:05")
}

// FormatTimeSince formats the duration since a time.
func FormatTimeSince(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return FormatDuration(time.Since(t))
}

// Status represents a status indicator.
type Status int

const (
	StatusOK Status = iota
	StatusWarning
	StatusError
	StatusPending
)

// FormatStatus returns a status indicator string.
func FormatStatus(s Status) string {
	switch s {
	case StatusOK:
		return "✓"
	case StatusWarning:
		return "⚠"
	case StatusError:
		return "✗"
	case StatusPending:
		return "○"
	default:
		return "?"
	}
}

// FormatCount formats a count with a label.
func FormatCount(count int, singular, plural string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	return fmt.Sprintf("%d %s", count, plural)
}

// FormatCurrency formats a USD amount.
func FormatCurrency(amount float64) string {
	return fmt.Sprintf("$%.2f", amount)
}

// ProgressBar generates a simple text progress bar.
func ProgressBar(current, total int, width int) string {
	if total == 0 {
		return strings.Repeat("░", width)
	}
	filled := (current * width) / total
	if filled > width {
		filled = width
	}
	empty := width - filled
	return strings.Repeat("█", filled) + strings.Repeat("░", empty)
}

// Color constants for terminal output.
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorGray   = "\033[90m"
)

// Colorize wraps text in ANSI color codes.
// Returns plain text if output is not a terminal.
func Colorize(text, color string) string {
	// For now, always colorize. Could add terminal detection later.
	return color + text + ColorReset
}
