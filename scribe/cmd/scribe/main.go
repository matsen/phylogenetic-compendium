// Package main is the entry point for the scribe CLI.
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/matsen/phylogenetic-compendium/scribe/internal/output"
	"github.com/matsen/phylogenetic-compendium/scribe/internal/queue"
	"github.com/matsen/phylogenetic-compendium/scribe/internal/status"
	"github.com/matsen/phylogenetic-compendium/scribe/internal/sweep"
	"github.com/matsen/phylogenetic-compendium/scribe/internal/verify"
	"github.com/spf13/cobra"
)

// Version is set at build time.
var Version = "dev"

// getOutputMode determines the output mode from command flags.
// Returns true for JSON output, false for human-readable.
// Defaults to JSON unless --human is specified.
func getOutputMode(cmd *cobra.Command, defaultToJSON bool) bool {
	humanOutput, _ := cmd.Flags().GetBool("human")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	if !humanOutput && !jsonOutput {
		return defaultToJSON
	}
	return jsonOutput
}

// queueContext holds common dependencies for queue commands.
type queueContext struct {
	service   *queue.CandidateService
	formatter *output.Formatter
	jsonMode  bool
}

// newQueueContext creates common dependencies for queue commands.
func newQueueContext(cmd *cobra.Command) *queueContext {
	jsonMode := getOutputMode(cmd, true) // queue commands default to JSON
	store := queue.NewStore("", "")
	return &queueContext{
		service:   queue.NewCandidateService(store),
		formatter: output.NewFormatter(jsonMode),
		jsonMode:  jsonMode,
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "scribe",
		Short: "Compendium authoring toolkit",
		Long: `scribe is a CLI for compendium authoring that implements
discovery, verification, and human-review workflows.`,
	}

	// Add version flag
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("scribe version {{.Version}}\n")

	// Add subcommands
	rootCmd.AddCommand(verifyCmd())
	rootCmd.AddCommand(queueCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(sweepCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func verifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify [files...]",
		Short: "Verify content before commit",
		Long: `Verify QMD files for citations, URLs, code links, claims, and TODO markers.

Checks:
- All paper IDs resolve in bipartite
- All repository URLs are accessible
- All code location links are valid
- Every factual claim has a citation
- No TODO/FIXME markers in content`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			humanOutput, _ := cmd.Flags().GetBool("human")
			summaryOnly, _ := cmd.Flags().GetBool("summary")
			jsonOutput, _ := cmd.Flags().GetBool("json")

			// Default to JSON output unless --human is specified
			if !humanOutput && !jsonOutput {
				jsonOutput = true
			}

			opts := verify.DefaultOptions()
			report, err := verify.VerifyFiles(args, opts)
			if err != nil {
				return fmt.Errorf("verification failed: %w", err)
			}

			formatter := output.NewFormatter(jsonOutput)

			if jsonOutput {
				if err := formatter.JSON(report); err != nil {
					return fmt.Errorf("output error: %w", err)
				}
			} else {
				// Human-readable output
				formatter.Header("Verification Report")
				formatter.Println("Files: %d", len(report.ContentFiles))
				formatter.Println("")

				formatter.Println("Summary:")
				formatter.Println("  Total checks: %d", report.Summary.TotalChecks)
				formatter.Println("  Passed: %s %d", output.FormatStatus(output.StatusOK), report.Summary.Passed)
				formatter.Println("  Failed: %s %d", output.FormatStatus(output.StatusError), report.Summary.Failed)
				formatter.Println("  Warnings: %s %d", output.FormatStatus(output.StatusWarning), report.Summary.Warnings)

				if !summaryOnly && len(report.Results) > 0 {
					// Show failures
					failures := report.FilterByStatus(verify.CheckStatusFail)
					if len(failures) > 0 {
						formatter.Header("Failures")
						for _, r := range failures {
							formatter.Println("%s %s:%d", output.FormatStatus(output.StatusError), r.Target.File, r.Target.Line)
							formatter.Println("   %s", r.Message)
						}
					}

					// Show warnings
					warnings := report.FilterByStatus(verify.CheckStatusWarn)
					if len(warnings) > 0 {
						formatter.Header("Warnings")
						for _, r := range warnings {
							formatter.Println("%s %s:%d", output.FormatStatus(output.StatusWarning), r.Target.File, r.Target.Line)
							formatter.Println("   %s", r.Message)
						}
					}
				}
			}

			// Exit with non-zero code if there are failures (per FR-007)
			if report.ExitCode != 0 {
				os.Exit(report.ExitCode)
			}

			return nil
		},
	}
	cmd.Flags().Bool("human", false, "Human-readable output")
	cmd.Flags().Bool("summary", false, "Show only summary")
	cmd.Flags().Bool("json", false, "JSON output (default)")
	return cmd
}

func queueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "queue",
		Short: "Manage candidate queue",
		Long:  `Add, list, approve, or reject candidates in the review queue.`,
	}

	// Add subcommands
	cmd.AddCommand(queueAddCmd())
	cmd.AddCommand(queueListCmd())
	cmd.AddCommand(queueApproveCmd())
	cmd.AddCommand(queueRejectCmd())
	cmd.AddCommand(queueGetCmd())
	cmd.AddCommand(queueStatsCmd())

	return cmd
}

func queueAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [type]",
		Short: "Add a candidate to the queue",
		Long: `Add a candidate to the review queue.

Types: paper, repo, code-location, concept`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newQueueContext(cmd)
			candidateType := queue.CandidateType(args[0])
			s2ID, _ := cmd.Flags().GetString("s2-id")
			repoURL, _ := cmd.Flags().GetString("repo")
			filePath, _ := cmd.Flags().GetString("file")
			lines, _ := cmd.Flags().GetString("lines")
			sha, _ := cmd.Flags().GetString("sha")
			description, _ := cmd.Flags().GetString("description")
			notes, _ := cmd.Flags().GetString("notes")

			candidate := queue.Candidate{
				Type:             candidateType,
				Status:           queue.CandidateStatusPending,
				DiscoveredAt:     time.Now(),
				DiscoveredBy:     "human",
				DiscoveryContext: "manual addition",
			}

			switch candidateType {
			case queue.CandidateTypePaper:
				if s2ID == "" {
					return fmt.Errorf("--s2-id is required for paper type")
				}
				candidate.PaperData = &queue.PaperData{
					S2ID:           s2ID,
					RelevanceNotes: notes,
				}
			case queue.CandidateTypeRepo:
				if repoURL == "" {
					return fmt.Errorf("--repo is required for repo type")
				}
				candidate.RepoData = &queue.RepoData{
					URL:            repoURL,
					RelevanceNotes: notes,
				}
			case queue.CandidateTypeCodeLocation:
				if repoURL == "" || filePath == "" || sha == "" {
					return fmt.Errorf("--repo, --file, and --sha are required for code-location type")
				}
				startLine, endLine := 1, 1
				if lines != "" {
					parts := strings.Split(lines, "-")
					if len(parts) == 2 {
						startLine, _ = strconv.Atoi(parts[0])
						endLine, _ = strconv.Atoi(parts[1])
					} else if len(parts) == 1 {
						startLine, _ = strconv.Atoi(parts[0])
						endLine = startLine
					}
				}
				candidate.CodeLocationData = &queue.CodeLocationData{
					RepoURL:     repoURL,
					FilePath:    filePath,
					StartLine:   startLine,
					EndLine:     endLine,
					CommitSHA:   sha,
					Description: description,
				}
			default:
				return fmt.Errorf("unknown type: %s", candidateType)
			}

			if err := ctx.service.Add(candidate); err != nil {
				return fmt.Errorf("failed to add candidate: %w", err)
			}

			if ctx.jsonMode {
				return ctx.formatter.JSON(map[string]string{"status": "added", "id": candidate.ID})
			}
			ctx.formatter.Println("Added candidate: %s", candidate.ID)
			return nil
		},
	}
	cmd.Flags().String("s2-id", "", "Semantic Scholar ID (for paper type)")
	cmd.Flags().String("repo", "", "Repository URL")
	cmd.Flags().String("file", "", "File path (for code-location type)")
	cmd.Flags().String("lines", "", "Line range (e.g., 100-150)")
	cmd.Flags().String("sha", "", "Commit SHA")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("notes", "", "Notes about the candidate")
	cmd.Flags().Bool("json", false, "JSON output (default)")
	cmd.Flags().Bool("human", false, "Human-readable output")
	return cmd
}

func queueListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List candidates in the queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newQueueContext(cmd)
			statusStr, _ := cmd.Flags().GetString("status")
			typeStr, _ := cmd.Flags().GetString("type")

			var statusFilter *queue.CandidateStatus
			if statusStr != "" {
				s := queue.CandidateStatus(statusStr)
				statusFilter = &s
			}

			var typeFilter *queue.CandidateType
			if typeStr != "" {
				t := queue.CandidateType(typeStr)
				typeFilter = &t
			}

			candidates, err := ctx.service.List(statusFilter, typeFilter)
			if err != nil {
				return fmt.Errorf("failed to list candidates: %w", err)
			}

			if ctx.jsonMode {
				return ctx.formatter.JSON(candidates)
			}

			if len(candidates) == 0 {
				ctx.formatter.Println("No candidates found")
				return nil
			}

			ctx.formatter.Header("Candidates")
			for _, candidate := range candidates {
				displayStatus := output.StatusPending
				switch candidate.Status {
				case queue.CandidateStatusApproved:
					displayStatus = output.StatusOK
				case queue.CandidateStatusRejected:
					displayStatus = output.StatusError
				}
				ctx.formatter.Println("%s [%s] %s (%s)", output.FormatStatus(displayStatus), candidate.ID, string(candidate.Type), candidate.Status)
			}
			return nil
		},
	}
	cmd.Flags().Bool("human", false, "Human-readable output")
	cmd.Flags().Bool("json", false, "JSON output (default)")
	cmd.Flags().String("status", "", "Filter by status (pending, approved, rejected)")
	cmd.Flags().String("type", "", "Filter by type (paper, repo, code-location, concept)")
	return cmd
}

func queueApproveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve [id]",
		Short: "Approve a candidate",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newQueueContext(cmd)
			id := args[0]
			notes, _ := cmd.Flags().GetString("notes")

			if err := ctx.service.Approve(id, "human", notes); err != nil {
				return fmt.Errorf("failed to approve: %w", err)
			}

			if ctx.jsonMode {
				return ctx.formatter.JSON(map[string]string{"status": "approved", "id": id})
			}
			ctx.formatter.Println("Approved: %s", id)
			return nil
		},
	}
	cmd.Flags().String("notes", "", "Approval notes")
	cmd.Flags().Bool("json", false, "JSON output (default)")
	cmd.Flags().Bool("human", false, "Human-readable output")
	return cmd
}

func queueRejectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reject [id]",
		Short: "Reject a candidate",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newQueueContext(cmd)
			id := args[0]
			reason, _ := cmd.Flags().GetString("reason")

			if err := ctx.service.Reject(id, "human", reason); err != nil {
				return fmt.Errorf("failed to reject: %w", err)
			}

			if ctx.jsonMode {
				return ctx.formatter.JSON(map[string]string{"status": "rejected", "id": id})
			}
			ctx.formatter.Println("Rejected: %s", id)
			return nil
		},
	}
	cmd.Flags().String("reason", "", "Rejection reason")
	cmd.Flags().Bool("json", false, "JSON output (default)")
	cmd.Flags().Bool("human", false, "Human-readable output")
	return cmd
}

func queueGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [id]",
		Short: "Get details of a candidate",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newQueueContext(cmd)
			id := args[0]

			candidate, err := ctx.service.Get(id)
			if err != nil {
				return fmt.Errorf("failed to get candidate: %w", err)
			}
			if candidate == nil {
				return fmt.Errorf("candidate not found: %s", id)
			}

			if ctx.jsonMode {
				return ctx.formatter.JSON(candidate)
			}

			ctx.formatter.Println("ID: %s", candidate.ID)
			ctx.formatter.Println("Type: %s", candidate.Type)
			ctx.formatter.Println("Status: %s", candidate.Status)
			ctx.formatter.Println("Discovered: %s by %s", output.FormatTime(candidate.DiscoveredAt), candidate.DiscoveredBy)
			return nil
		},
	}
	cmd.Flags().Bool("json", false, "JSON output (default)")
	cmd.Flags().Bool("human", false, "Human-readable output")
	return cmd
}

func queueStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show queue statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newQueueContext(cmd)

			stats, err := ctx.service.Stats()
			if err != nil {
				return fmt.Errorf("failed to get stats: %w", err)
			}

			if ctx.jsonMode {
				return ctx.formatter.JSON(stats)
			}

			ctx.formatter.Header("Queue Statistics")
			ctx.formatter.Println("Total: %d", stats.Total)
			ctx.formatter.Println("Pending: %d", stats.Pending)
			ctx.formatter.Println("Approved: %d", stats.Approved)
			ctx.formatter.Println("Rejected: %d", stats.Rejected)
			ctx.formatter.Println("")
			ctx.formatter.Println("By Type:")
			for candidateType, count := range stats.ByType {
				ctx.formatter.Println("  %s: %d", candidateType, count)
			}
			return nil
		},
	}
	cmd.Flags().Bool("human", false, "Human-readable output")
	cmd.Flags().Bool("json", false, "JSON output (default)")
	return cmd
}

func statusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show autonomous task status",
		Long: `Pretty-print the current task checkpoint.

Shows:
- Task name and type
- Running duration and iteration count
- Progress (completed, pending, blocked items)
- Candidates queued
- Estimated cost`,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOutput, _ := cmd.Flags().GetBool("json")
			humanOutput, _ := cmd.Flags().GetBool("human")

			// Default to human output for status command
			if !humanOutput && !jsonOutput {
				humanOutput = true
			}

			store := status.NewCheckpointStore("")
			checkpoint, err := store.Read()
			if err != nil {
				return fmt.Errorf("failed to read checkpoint: %w", err)
			}

			display := status.NewDisplay(!humanOutput)

			if checkpoint == nil {
				display.ShowNoTask()
				return nil
			}

			display.ShowCheckpoint(checkpoint)
			return nil
		},
	}
	cmd.Flags().Bool("human", false, "Human-readable output (default)")
	cmd.Flags().Bool("json", false, "JSON output")
	return cmd
}

func sweepCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sweep [files...]",
		Short: "Run periodic verification sweep",
		Long: `Check claim-source consistency, repo freshness, code link validity.

Checks:
- Claim-consistency: Do cited papers still support the claims?
- Repo-freshness: Are referenced repos still maintained (< 2 years)?
- Code-link validity: Do file paths and line ranges still exist?
- Coverage gaps: Are there undocumented techniques?`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			humanOutput, _ := cmd.Flags().GetBool("human")
			jsonOutput, _ := cmd.Flags().GetBool("json")
			checkStr, _ := cmd.Flags().GetString("check")

			if !humanOutput && !jsonOutput {
				jsonOutput = true
			}

			opts := sweep.DefaultOptions()

			// Filter by specific check if requested
			if checkStr != "" {
				var checkType sweep.CheckType
				switch checkStr {
				case "repo-freshness":
					checkType = sweep.CheckTypeRepoFreshness
				case "claim-consistency":
					checkType = sweep.CheckTypeClaimConsistency
				case "code-links":
					checkType = sweep.CheckTypeCodeLinks
				case "coverage":
					checkType = sweep.CheckTypeCoverage
				default:
					return fmt.Errorf("unknown check type: %s", checkStr)
				}
				opts.Checks = []sweep.CheckType{checkType}
			}

			report, err := sweep.SweepFiles(args, opts)
			if err != nil {
				return fmt.Errorf("sweep failed: %w", err)
			}

			formatter := output.NewFormatter(jsonOutput)

			if jsonOutput {
				if err := formatter.JSON(report); err != nil {
					return fmt.Errorf("output error: %w", err)
				}
			} else {
				formatter.Header("Sweep Report")
				formatter.Println("Files: %d", len(report.ContentFiles))
				formatter.Println("Checks: %v", report.ChecksRun)
				formatter.Println("")

				formatter.Println("Summary:")
				formatter.Println("  Total checks: %d", report.Summary.TotalChecks)
				formatter.Println("  OK: %s %d", output.FormatStatus(output.StatusOK), report.Summary.OK)
				formatter.Println("  Issues: %s %d", output.FormatStatus(output.StatusError), report.Summary.Issues)
				formatter.Println("  Warnings: %s %d", output.FormatStatus(output.StatusWarning), report.Summary.Warnings)

				// Show issues
				issues := report.FilterByStatus(sweep.SweepStatusIssue)
				if len(issues) > 0 {
					formatter.Header("Issues")
					for _, r := range issues {
						formatter.Println("%s [%s] %s", output.FormatStatus(output.StatusError), r.CheckType, r.Target)
						formatter.Println("   %s", r.Message)
						if r.SuggestedFix != "" {
							formatter.Println("   Fix: %s", r.SuggestedFix)
						}
					}
				}

				// Show warnings
				warnings := report.FilterByStatus(sweep.SweepStatusWarning)
				if len(warnings) > 0 {
					formatter.Header("Warnings")
					for _, r := range warnings {
						formatter.Println("%s [%s] %s", output.FormatStatus(output.StatusWarning), r.CheckType, r.Target)
						formatter.Println("   %s", r.Message)
					}
				}
			}

			return nil
		},
	}
	cmd.Flags().String("check", "", "Run specific check (repo-freshness, claim-consistency, code-links, coverage)")
	cmd.Flags().Bool("human", false, "Human-readable output")
	cmd.Flags().Bool("json", false, "JSON output (default)")
	return cmd
}
