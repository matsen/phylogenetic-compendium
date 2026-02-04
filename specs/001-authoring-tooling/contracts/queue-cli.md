# Scribe Queue CLI Contract

**Version**: 1.0 | **Date**: 2026-02-04

## Overview

The `scribe queue` subcommand provides human and agent access to the candidate queue. All commands output JSON by default for agent integration.

## Commands

### `scribe queue add`

Add a candidate to the queue.

```bash
scribe queue add paper --s2-id "S2:abc123" --notes "Found in FastTree references"
scribe queue add code-location --repo "https://github.com/org/repo" \
  --file "src/likelihood.c" --lines 142-180 --sha "abc123def" \
  --description "Likelihood computation for GTR model"
scribe queue add concept --name "Site-specific rates" --description "..."
scribe queue add repo --url "https://github.com/org/repo" --notes "Implements X"
```

**Arguments**:
- `TYPE`: paper | code-location | concept | repo

**Flags**:
| Flag | Required | Description |
|------|----------|-------------|
| `--s2-id` | paper | Semantic Scholar paper ID |
| `--repo` | code-location, repo | GitHub repository URL |
| `--file` | code-location | File path relative to repo root |
| `--lines` | code-location | Line range (e.g., "142-180") |
| `--sha` | code-location | Commit SHA |
| `--name` | concept | Concept name |
| `--description` | concept, code-location | Human-readable description |
| `--notes` | optional | Agent notes about relevance |
| `--context` | optional | Discovery context (task ID) |

**Output** (JSON):
```json
{
  "status": "added",
  "candidate_id": "c-2026020401",
  "type": "paper",
  "external_id": "S2:abc123"
}
```

**Exit Codes**:
- `0`: Success
- `1`: Validation error (missing required field, invalid format)
- `2`: Duplicate candidate (same external_id already exists)
- `3`: Previously rejected (candidate was rejected before)

---

### `scribe queue list`

List candidates with optional filtering.

```bash
scribe queue list                           # All pending
scribe queue list --status approved         # All approved
scribe queue list --type paper              # All pending papers
scribe queue list --type paper --status all # All papers regardless of status
scribe queue list --json                    # Force JSON output (default)
scribe queue list --human                   # Human-readable table
```

**Flags**:
| Flag | Default | Description |
|------|---------|-------------|
| `--status` | pending | pending \| approved \| rejected \| all |
| `--type` | all | paper \| code-location \| concept \| repo \| all |
| `--limit` | 50 | Maximum results |
| `--json` | true | JSON output |
| `--human` | false | Human-readable table |

**Output** (JSON):
```json
{
  "count": 12,
  "candidates": [
    {
      "id": "c-2026020401",
      "type": "paper",
      "status": "pending",
      "external_id": "S2:abc123",
      "discovered_at": "2026-02-04T10:30:00Z",
      "notes": "Found in FastTree references"
    }
  ]
}
```

---

### `scribe queue approve`

Approve a pending candidate.

```bash
scribe queue approve c-2026020401 --notes "Relevant to tree traversal section"
```

**Arguments**:
- `CANDIDATE_ID`: ID of the candidate to approve

**Flags**:
| Flag | Required | Description |
|------|----------|-------------|
| `--notes` | optional | Review notes |

**Side Effects**:
- For paper candidates: Calls `bip s2 add {s2_id}` to add to bipartite
- For repo candidates: Calls `bip repo add {url}` to add to bipartite
- For concept candidates: Calls `bip concept add {name}` to add to bipartite
- For code-location candidates: Stores in approved list (pending bipartite extension)

**Output** (JSON):
```json
{
  "status": "approved",
  "candidate_id": "c-2026020401",
  "bipartite_added": true,
  "bipartite_id": "paper:abc123"
}
```

**Exit Codes**:
- `0`: Success
- `1`: Candidate not found
- `2`: Candidate not pending
- `3`: Bipartite add failed

---

### `scribe queue reject`

Reject a pending candidate.

```bash
scribe queue reject c-2026020401 --reason "Not relevant to compendium scope"
```

**Arguments**:
- `CANDIDATE_ID`: ID of the candidate to reject

**Flags**:
| Flag | Required | Description |
|------|----------|-------------|
| `--reason` | required | Reason for rejection |

**Side Effects**:
- Moves candidate to `.candidates/rejected.jsonl`
- Prevents re-discovery by future agents

**Output** (JSON):
```json
{
  "status": "rejected",
  "candidate_id": "c-2026020401",
  "reason": "Not relevant to compendium scope"
}
```

---

### `scribe queue get`

Get details of a specific candidate.

```bash
scribe queue get c-2026020401
```

**Output**: Full Candidate JSON object (see data-model.md)

---

### `scribe queue stats`

Get queue statistics.

```bash
scribe queue stats
```

**Output** (JSON):
```json
{
  "total": 45,
  "by_status": {
    "pending": 12,
    "approved": 30,
    "rejected": 3
  },
  "by_type": {
    "paper": 20,
    "code-location": 15,
    "concept": 8,
    "repo": 2
  },
  "oldest_pending": "2026-02-01T10:30:00Z"
}
```

---

## Agent Integration

Agents should use JSON output for parsing:

```bash
# Add a candidate and capture the ID
result=$(scribe queue add paper --s2-id "S2:abc123" --notes "...")
candidate_id=$(echo "$result" | jq -r '.candidate_id')

# Check if a paper was previously rejected
if scribe queue add paper --s2-id "S2:xyz789" 2>&1 | jq -e '.status == "previously_rejected"' > /dev/null; then
  echo "Skipping previously rejected paper"
fi
```

## Error Responses

All errors return JSON with `error` field:

```json
{
  "error": "validation_failed",
  "message": "Missing required field: --s2-id",
  "field": "s2_id"
}
```

```json
{
  "error": "previously_rejected",
  "message": "This candidate was rejected on 2026-02-01",
  "original_id": "c-2026020101",
  "rejection_reason": "Not relevant to compendium scope"
}
```
