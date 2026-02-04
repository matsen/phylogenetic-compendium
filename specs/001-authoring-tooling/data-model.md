# Data Model: Compendium Authoring Tooling

**Date**: 2026-02-04 | **Status**: Draft

## Overview

This document defines the data structures used by the compendium authoring system. All structures are stored as JSONL files for git-friendliness and compatibility with bipartite's storage pattern.

## Entities

### Candidate

A discovered item awaiting human review. Stored in `.candidates/queue.jsonl`.

```json
{
  "id": "string",           // Unique ID (e.g., "c-2026020401")
  "type": "paper|concept|repo|code-location",
  "status": "pending|approved|rejected",

  // Discovery metadata
  "discovered_at": "ISO-8601",
  "discovered_by": "string", // Agent ID or "human"
  "discovery_context": "string", // Task that found this

  // Type-specific data (one of these based on type)
  "paper_data": {
    "s2_id": "string",       // Semantic Scholar ID
    "title": "string",
    "authors": ["string"],
    "year": "number",
    "relevance_notes": "string"
  },
  "concept_data": {
    "name": "string",
    "description": "string",
    "related_papers": ["string"], // S2 IDs
    "related_repos": ["string"]   // Repo URLs
  },
  "repo_data": {
    "url": "string",
    "name": "string",
    "description": "string",
    "relevance_notes": "string"
  },
  "code_location_data": {
    "repo_url": "string",
    "file_path": "string",
    "start_line": "number",
    "end_line": "number",
    "commit_sha": "string",
    "permalink_url": "string",
    "function_name": "string|null",
    "description": "string",
    "surrounding_context": "string" // ~10 lines around the location
  },

  // Review metadata (populated on approve/reject)
  "reviewed_at": "ISO-8601|null",
  "reviewed_by": "string|null",
  "review_notes": "string|null",
  "rejection_reason": "string|null" // Only if rejected
}
```

**Validation Rules**:
- `id` must be unique across all candidates
- `type` must be one of the enum values
- `status` transitions: `pending` → `approved` OR `pending` → `rejected`
- `discovered_at` must be valid ISO-8601
- Type-specific data field must match `type` (e.g., `paper_data` when `type: "paper"`)

---

### CodeLocation

A specific location in a codebase. Embedded in Candidate but also used standalone in verification.

```json
{
  "repo_url": "string",      // Full GitHub URL
  "file_path": "string",     // Relative path from repo root
  "start_line": "number",    // 1-indexed
  "end_line": "number",      // 1-indexed, inclusive
  "commit_sha": "string",    // 40-char SHA
  "permalink_url": "string", // Full GitHub blob URL
  "function_name": "string|null",
  "description": "string"
}
```

**Permalink Format**:
```
https://github.com/{org}/{repo}/blob/{commit_sha}/{file_path}#L{start_line}-L{end_line}
```

**Validation Rules**:
- `start_line` ≤ `end_line`
- `commit_sha` must be 40 hex characters
- `permalink_url` must match the generated format from other fields

---

### VerificationResult

Outcome of a verification check. Stored in verification reports.

```json
{
  "check_id": "string",       // Unique check ID
  "check_type": "citation|url|code-link|claim|todo-marker",
  "target": {
    "file": "string",         // Content file path
    "line": "number",         // Line number in file
    "text": "string"          // Relevant text snippet
  },
  "status": "pass|fail|warn",
  "message": "string",        // Human-readable result
  "details": {                // Type-specific details
    // For citation checks:
    "paper_id": "string",
    "resolved": "boolean",

    // For URL checks:
    "url": "string",
    "http_status": "number|null",
    "error": "string|null",

    // For code-link checks:
    "permalink": "string",
    "file_exists": "boolean",
    "line_range_valid": "boolean",

    // For claim checks:
    "claim_text": "string",
    "confidence": "high|medium|low",
    "suggested_action": "string"
  },
  "checked_at": "ISO-8601"
}
```

---

### VerificationReport

Aggregated verification results. Output of `verify-content.sh`.

```json
{
  "report_id": "string",
  "generated_at": "ISO-8601",
  "content_files": ["string"], // Files that were verified
  "summary": {
    "total_checks": "number",
    "passed": "number",
    "failed": "number",
    "warnings": "number"
  },
  "results": [VerificationResult],
  "exit_code": "number"        // 0 = all pass, 1 = failures
}
```

---

### TaskCheckpoint

Progress state for autonomous operation. Stored in `.claude/authoring/checkpoint.json`.

```json
{
  "task_id": "string",        // UUID
  "task_type": "exploration|survey|verification-sweep",
  "task_description": "string",

  // Timing
  "started_at": "ISO-8601",
  "last_checkpoint": "ISO-8601",
  "iteration_count": "number",

  // Configuration
  "prompt_file": "string",    // Path to PROMPT.md
  "max_iterations": "number",
  "cost_budget_usd": "number|null",

  // Progress
  "state": {
    "items_completed": ["string"],
    "items_pending": ["string"],
    "blocked_items": [{
      "item": "string",
      "reason": "string",
      "blocked_at": "ISO-8601"
    }],
    "current_focus": "string"
  },

  // Metrics
  "metrics": {
    "candidates_queued": "number",
    "papers_found": "number",
    "code_locations_found": "number",
    "repos_searched": "number",
    "estimated_cost_usd": "number"
  }
}
```

---

### AgentActionLog

Record of an agent's action. Stored in `.claude/authoring/logs/actions.jsonl`.

```json
{
  "log_id": "string",
  "task_id": "string",        // Links to TaskCheckpoint
  "agent_type": "exploration|survey|consumer|verification",
  "action": "string",         // e.g., "search_codebase", "queue_candidate"
  "target": "string",         // What was acted on
  "result": "success|failure|skipped",
  "message": "string|null",
  "timestamp": "ISO-8601"
}
```

---

### SurveyFinding

A structured finding from comparative survey. Embedded in survey output.

```json
{
  "finding_id": "string",
  "repo": "string",           // Repo URL
  "repo_name": "string",      // Display name
  "concept": "string",        // What was being surveyed
  "approach_summary": "string", // How this repo handles it
  "code_location": CodeLocation,
  "key_insight": "string",    // What's notable/unique
  "related_papers": ["string"], // S2 IDs
  "priority": "normal|highlighted" // "highlighted" for "especially clever"
}
```

---

## Storage Layout

```
.candidates/
├── queue.jsonl           # All candidates (Candidate records)
├── rejected.jsonl        # Rejected candidates (for re-discovery prevention)
└── .cache/
    └── index.sqlite      # Ephemeral query index (gitignored)

.claude/authoring/
├── checkpoint.json       # Current TaskCheckpoint
├── session.json          # Ralph Loop session state
├── logs/
│   └── actions.jsonl     # AgentActionLog records
└── history/
    └── {date}-{type}-{target}.json  # Completed task summaries

reports/
└── verification/
    └── {date}-{time}.json  # VerificationReport records
```

## Relationships

```
TaskCheckpoint ──creates──► Candidate (via discovery)
Candidate ──contains──► CodeLocation (when type = code-location)
TaskCheckpoint ──logs-to──► AgentActionLog
VerificationReport ──contains──► VerificationResult
SurveyFinding ──references──► CodeLocation
```

## State Transitions

### Candidate Status
```
[created] → pending → approved → [added to bipartite]
                   ↘
                    rejected → [stored in rejected.jsonl]
```

### Task Lifecycle
```
[started] → running → [checkpoint saved] → running → ... → completed
                                                        ↘
                                                         blocked → [human review]
```
