# Verification Report Contract

**Version**: 1.0 | **Date**: 2026-02-04

## Overview

The verification system (`tools/verify/verify-content.sh`) produces structured reports that can be consumed by humans (via formatted output) or agents (via JSON).

## Report Structure

### Full Report (JSON)

```json
{
  "report_id": "ver-20260204-103000",
  "generated_at": "2026-02-04T10:30:00Z",
  "content_files": [
    "intro.qmd",
    "chapters/tree-structures.qmd"
  ],
  "summary": {
    "total_checks": 47,
    "passed": 42,
    "failed": 3,
    "warnings": 2
  },
  "exit_code": 1,
  "results": [
    {
      "check_id": "cit-001",
      "check_type": "citation",
      "target": {
        "file": "intro.qmd",
        "line": 45,
        "text": "...as shown by [@paper:fasttree]..."
      },
      "status": "pass",
      "message": "Paper ID resolves in bipartite",
      "details": {
        "paper_id": "paper:fasttree",
        "resolved": true,
        "bipartite_title": "FastTree: Computing Large Minimum Evolution Trees"
      },
      "checked_at": "2026-02-04T10:30:01Z"
    },
    {
      "check_id": "cit-002",
      "check_type": "citation",
      "target": {
        "file": "chapters/tree-structures.qmd",
        "line": 123,
        "text": "...the approach from [@paper:unknown123]..."
      },
      "status": "fail",
      "message": "Paper ID not found in bipartite",
      "details": {
        "paper_id": "paper:unknown123",
        "resolved": false,
        "suggestion": "Run 'bip s2 search unknown123' to find the paper"
      },
      "checked_at": "2026-02-04T10:30:02Z"
    }
  ]
}
```

## Check Types

### `citation`

Verifies paper IDs resolve in bipartite.

**Target**: Lines containing `[@paper:...]` or `[@concept:...]` references

**Status Outcomes**:
- `pass`: Paper ID found in bipartite
- `fail`: Paper ID not found in bipartite
- `warn`: Paper ID found but metadata incomplete

**Details**:
```json
{
  "paper_id": "paper:fasttree",
  "resolved": true,
  "bipartite_title": "FastTree: ...",
  "bipartite_year": 2010
}
```

---

### `url`

Verifies repository URLs are accessible.

**Target**: Lines containing GitHub URLs or repo references

**Status Outcomes**:
- `pass`: URL returns HTTP 200
- `fail`: URL returns 404 or connection error
- `warn`: URL redirects (repo may have moved)

**Details**:
```json
{
  "url": "https://github.com/org/repo",
  "http_status": 200,
  "response_time_ms": 234
}
```

Or on failure:
```json
{
  "url": "https://github.com/org/deleted-repo",
  "http_status": 404,
  "error": "Repository not found"
}
```

---

### `code-link`

Verifies GitHub permalink URLs point to valid code.

**Target**: Lines containing GitHub blob URLs with line numbers

**Status Outcomes**:
- `pass`: File exists at SHA, line range valid
- `fail`: File doesn't exist, SHA invalid, or line range out of bounds
- `warn`: File exists but line range may have drifted

**Details**:
```json
{
  "permalink": "https://github.com/org/repo/blob/abc123/src/file.c#L10-L20",
  "repo": "org/repo",
  "sha": "abc123",
  "file_path": "src/file.c",
  "start_line": 10,
  "end_line": 20,
  "file_exists": true,
  "file_total_lines": 500,
  "line_range_valid": true
}
```

---

### `claim`

Detects factual claims that may need citations.

**Target**: Sentences matching claim detection heuristics

**Status Outcomes**:
- `pass`: Claim has adjacent citation
- `fail`: High-confidence claim without citation
- `warn`: Medium-confidence claim without citation (needs review)

**Details**:
```json
{
  "claim_text": "FastTree is faster than RAxML for large datasets",
  "confidence": "high",
  "pattern_matched": "comparison",
  "suggested_action": "Add citation for performance comparison"
}
```

---

### `todo-marker`

Detects TODO/FIXME markers in content.

**Target**: Lines containing TODO, FIXME, XXX, HACK

**Status Outcomes**:
- `fail`: Marker found in publishable content
- `pass`: No markers found

**Details**:
```json
{
  "marker": "TODO",
  "full_text": "TODO: Add citation for this claim"
}
```

---

## CLI Output Modes

### JSON (default, for agents)

```bash
verify-content.sh intro.qmd chapters/*.qmd
# Outputs full JSON report to stdout
```

### Human-readable (for terminal)

```bash
verify-content.sh --human intro.qmd chapters/*.qmd
```

Output:
```
Verification Report: 2026-02-04 10:30:00
Files: intro.qmd, chapters/tree-structures.qmd

Summary: 42 passed, 3 failed, 2 warnings

FAILURES:
  ✗ chapters/tree-structures.qmd:123 - citation
    Paper ID not found: paper:unknown123
    Suggestion: Run 'bip s2 search unknown123' to find the paper

  ✗ intro.qmd:89 - code-link
    Permalink invalid: file no longer exists at SHA
    URL: https://github.com/org/repo/blob/old-sha/deleted-file.c

  ✗ intro.qmd:156 - claim
    Uncited claim: "FastTree is faster than RAxML"
    Add citation for this performance comparison

WARNINGS:
  ⚠ chapters/algorithms.qmd:45 - url
    Repository redirect detected: org/old-name → org/new-name

  ⚠ chapters/algorithms.qmd:78 - claim
    Possible uncited claim (medium confidence): "This approach reduces memory usage"
```

### Summary only

```bash
verify-content.sh --summary intro.qmd chapters/*.qmd
```

Output:
```json
{"passed": 42, "failed": 3, "warnings": 2, "exit_code": 1}
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All checks passed |
| 1 | One or more checks failed |
| 2 | Input error (file not found, invalid arguments) |
| 3 | System error (bipartite unavailable, network error) |

---

## Pre-commit Hook Integration

```bash
# .git/hooks/pre-commit
#!/bin/bash
if ! tools/verify/verify-content.sh --summary *.qmd chapters/*.qmd | jq -e '.failed == 0' > /dev/null; then
  echo "Verification failed. Run 'tools/verify/verify-content.sh --human' for details."
  exit 1
fi
```

---

## Periodic Verification (Sweep)

The sweep adds additional checks beyond pre-commit:

```json
{
  "check_type": "claim-consistency",
  "details": {
    "claim_text": "FastTree uses neighbor-joining for initial tree",
    "cited_paper": "paper:fasttree",
    "asta_snippet": "FastTree uses a heuristic similar to neighbor-joining...",
    "consistency": "partial_match"
  }
}
```

```json
{
  "check_type": "repo-freshness",
  "details": {
    "repo_url": "https://github.com/org/old-project",
    "last_commit": "2022-03-15",
    "days_stale": 1420,
    "threshold_days": 730
  }
}
```

```json
{
  "check_type": "coverage-gap",
  "details": {
    "concept": "concept:site-specific-rates",
    "has_papers": true,
    "has_implementations": false,
    "suggestion": "Add implementation references for site-specific rates"
  }
}
```
