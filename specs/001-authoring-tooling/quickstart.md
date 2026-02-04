# Quickstart: Compendium Authoring Tooling

**Date**: 2026-02-04

## Prerequisites

Before using the authoring tools, ensure you have:

1. **bipartite CLI** (`bip`) installed and configured
   ```bash
   bip --version  # Should output version info
   bip config get nexus  # Should show your nexus path
   ```

2. **GitHub CLI** (`gh`) authenticated
   ```bash
   gh auth status  # Should show logged in
   ```

3. **Claude Code** with Ralph Loop plugin
   ```bash
   claude --version
   # Ralph Loop should be listed in plugins
   ```

4. **scribe CLI** installed
   ```bash
   cd scribe && go install ./cmd/scribe
   scribe --help
   ```

5. **LLM for NLP tasks** (one of):
   ```bash
   # Option A: Claude CLI configured
   claude --version

   # Option B: Local Ollama running
   ollama list  # Should show available models
   ```

## Quick Commands

### Verify Content Before Commit

```bash
# Verify all QMD files
scribe verify *.qmd chapters/*.qmd

# Human-readable output
scribe verify --human *.qmd

# Just the summary
scribe verify --summary *.qmd
```

**What it checks**:
- All paper IDs resolve in bipartite
- All repository URLs are accessible
- All code location links are valid
- Every factual claim has a citation
- No TODO/FIXME markers in content

### Manage the Candidate Queue

```bash
# List pending candidates
scribe queue list

# List with human-readable output
scribe queue list --human

# Add a paper candidate
scribe queue add paper --s2-id "S2:649def34f8be52c8b66281af98ae884c09aef38b" \
  --notes "Describes the GTR model implementation"

# Add a code location candidate
scribe queue add code-location \
  --repo "https://github.com/stamatak/standard-RAxML" \
  --file "axml.c" --lines 1423-1498 --sha "abc123def456" \
  --description "GTR rate matrix computation"

# Approve a candidate (adds to bipartite)
scribe queue approve c-2026020401 --notes "Relevant to rate matrices section"

# Reject a candidate
scribe queue reject c-2026020402 --reason "Out of scope - focuses on visualization"

# Check queue statistics
scribe queue stats
```

### Check Autonomous Task Status

```bash
# Pretty-print current task progress
scribe status

# Example output:
# Task: exploration-fasttree
# Running for: 1h 23m (iteration 7/50)
# Progress: 45 items completed, 12 pending, 2 blocked
# Candidates queued: 28 (15 papers, 13 code locations)
# Estimated cost: $4.23
```

### Run Exploration Agent (Autonomous)

```bash
# Start an exploration task (runs via Ralph Loop)
claude --ralph "$(cat scribe/agents/exploration/PROMPT.md)" \
  --var TARGET_REPO="https://github.com/morgannprice/fasttree" \
  --var TECHNIQUE_HINT="rotated likelihood vectors" \
  --var PROVENANCE_HINT="came from RAxML"

# Check progress while running
scribe status

# Review candidates after completion
scribe queue list --human
```

### Run Survey Agent (Autonomous)

```bash
# Start a comparative survey
claude --ralph "$(cat scribe/agents/survey/PROMPT.md)" \
  --var CONCEPT="partial result caching" \
  --var REPOS="RAxML,FastTree,bali-phy,IQ-TREE" \
  --var HIGHLIGHT_REPO="bali-phy"

# Output will be in survey-results.md and candidates queued
```

### Run Periodic Verification Sweep

```bash
# Full sweep (takes longer)
scribe sweep *.qmd chapters/*.qmd

# Just check for stale repos
scribe sweep --check repo-freshness

# Just check claim consistency
scribe sweep --check claim-consistency
```

## Typical Workflows

### Workflow 1: Write and Verify New Content

1. **Write content** in QMD file with citations
   ```markdown
   The GTR model [@paper:tavare1986] generalizes the JC69 model
   by allowing different substitution rates.
   ```

2. **Run verification** before commit
   ```bash
   scribe verify chapters/models.qmd
   ```

3. **Fix any failures** (missing citations, broken links)

4. **Commit** when verification passes

### Workflow 2: Explore Undocumented Code

1. **Start exploration** with what you know
   ```bash
   claude --ralph "$(cat scribe/agents/exploration/PROMPT.md)" \
     --var TARGET_REPO="https://github.com/morgannprice/fasttree" \
     --var TECHNIQUE_HINT="likelihood caching"
   ```

2. **Walk away** - let it run overnight

3. **Check progress** (optional)
   ```bash
   scribe status
   ```

4. **Review candidates** next morning
   ```bash
   scribe queue list --human --type code-location
   scribe queue list --human --type paper
   ```

5. **Approve useful findings**
   ```bash
   scribe queue approve c-001 --notes "Key likelihood caching code"
   scribe queue approve c-002 --notes "Original paper describing technique"
   ```

6. **Write content** using approved references

### Workflow 3: Compare Implementations

1. **Start survey** across repos
   ```bash
   claude --ralph "scribe/agents/survey/PROMPT.md" \
     --var CONCEPT="tree traversal optimization" \
     --var REPOS="RAxML,FastTree,IQ-TREE,PhyML"
   ```

2. **Review structured output** in `survey-results.md`

3. **Approve relevant candidates**

4. **Write comparative section** using the findings

## File Locations

| Path | Contents |
|------|----------|
| `.candidates/queue.jsonl` | Candidate queue |
| `.candidates/rejected.jsonl` | Rejected candidates |
| `.claude/authoring/checkpoint.json` | Current task progress |
| `.claude/authoring/logs/` | Agent action logs |
| `reports/verification/` | Verification reports |

## Troubleshooting

### "Paper ID not found in bipartite"
```bash
# Search for the paper
bip s2 search "paper title keywords"

# Add it manually if found
bip s2 add S2:paper-id
```

### "bipartite unavailable"
```bash
# Check bipartite status
bip check

# Rebuild if needed
bip rebuild
```

### "GitHub rate limit exceeded"
Wait for rate limit reset, or authenticate with `gh auth login` for higher limits.

### "Candidate was previously rejected"
The system prevents re-queuing rejected items. If you want to reconsider:
```bash
# Check why it was rejected
scribe queue get c-2026020402 | jq '.rejection_reason'

# Manually remove from rejected.jsonl if you want to allow re-discovery
```
