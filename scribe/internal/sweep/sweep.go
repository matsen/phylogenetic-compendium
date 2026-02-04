package sweep

import (
	"os"
)

// Options configures sweep behavior.
type Options struct {
	Checks []CheckType // Which checks to run (empty = all)
}

// DefaultOptions returns the default sweep options.
func DefaultOptions() Options {
	return Options{
		Checks: []CheckType{
			CheckTypeRepoFreshness,
			CheckTypeCodeLinks,
			CheckTypeClaimConsistency,
			CheckTypeCoverage,
		},
	}
}

// SweepFile runs sweep checks on a single file.
func SweepFile(filePath string, opts Options) ([]SweepResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	contentStr := string(content)
	var results []SweepResult

	checks := opts.Checks
	if len(checks) == 0 {
		checks = DefaultOptions().Checks
	}

	for _, check := range checks {
		switch check {
		case CheckTypeRepoFreshness:
			results = append(results, CheckRepoFreshness(contentStr, filePath)...)
		case CheckTypeCodeLinks:
			results = append(results, CheckCodeLinks(contentStr, filePath)...)
		case CheckTypeClaimConsistency:
			results = append(results, CheckClaimConsistency(contentStr, filePath)...)
		case CheckTypeCoverage:
			results = append(results, CheckCoverageGaps(contentStr, filePath)...)
		}
	}

	return results, nil
}

// SweepFiles runs sweep checks on multiple files.
func SweepFiles(files []string, opts Options) (*SweepReport, error) {
	builder := NewSweepReportBuilder()

	checks := opts.Checks
	if len(checks) == 0 {
		checks = DefaultOptions().Checks
	}

	for _, check := range checks {
		builder.AddCheck(check)
	}

	for _, file := range files {
		builder.AddFile(file)

		results, err := SweepFile(file, opts)
		if err != nil {
			// Create an error result for the file
			builder.AddResult(SweepResult{
				CheckType: CheckTypeRepoFreshness,
				Status:    SweepStatusIssue,
				Target:    file,
				File:      file,
				Message:   err.Error(),
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
