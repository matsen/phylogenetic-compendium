# Research: Compendium Authoring Tooling

**Date**: 2026-02-04 | **Status**: Complete

## Research Questions

### 1. Bipartite Code Location Support

**Question**: Does bipartite currently support code location metadata (file path, line range, commit SHA) on repo nodes?

**Finding**: No. The `bip repo add` command supports basic GitHub repo metadata (URL, name, description, topics) but not file-level granularity. Repo nodes are coarse-grained.

**Decision**: Store code locations in the **candidate queue** for immediate use. Code location candidates will include full metadata (repo_url, file_path, start_line, end_line, commit_sha, permalink_url).

**Future Work**: Propose bipartite extension for code-location nodes as a new node type, separate from repo nodes. This would enable edges like `paper --implements--> code-location` and `code-location --in-repo--> repo`.

**Rationale**: The candidate queue already needs code location storage for human review. Duplicating this in bipartite now adds complexity without immediate benefit. We can migrate approved code locations to bipartite once the extension exists.

---

### 2. Claim Detection Heuristics

**Question**: How to distinguish factual claims requiring citations from definitions, examples, and transitional prose?

**Finding**: Academic research ([Citation prediction with transformers](https://www.sciencedirect.com/science/article/pii/S0306457323003205)) uses transformer models with NLP heuristics to predict citation need. Key patterns include:
- Causal language: "causes", "leads to", "results in"
- Quantitative claims: numbers, percentages, comparisons
- Attribution language: "X showed that", "according to", "is known to"
- Behavioral claims: "X does Y", "X implements Z"

**Decision**: Use **pattern-based heuristics** initially with manual review fallback:

1. **Must cite** (high confidence):
   - Sentences with performance comparisons ("X is faster than Y")
   - Sentences attributing discoveries ("discovered by", "introduced by")
   - Sentences citing prior work ("as shown in", "as described by")

2. **Should cite** (medium confidence):
   - Sentences with causal claims about algorithms
   - Sentences describing algorithmic complexity
   - Sentences about historical development

3. **Exempt** (no citation needed):
   - Definitions (signaled by "is defined as", "refers to")
   - Examples (signaled by "for example", "e.g.", "such as")
   - Transitional prose (signaled by "in this section", "we now turn to")
   - Code block descriptions

**Implementation**: Bash script using grep patterns for initial pass, with TODO markers on ambiguous sentences for human review.

**Rationale**: Perfect claim detection is an NLP research problem. Pattern-based heuristics catch obvious cases; human review catches the rest. Avoid ML complexity for MVP.

---

### 3. Ralph Loop Checkpoint Format

**Question**: What format does the Ralph Loop plugin expect for checkpoints?

**Finding**: From [ralph-claude-code](https://github.com/frankbria/ralph-claude-code) and the official plugin:
- State files stored as JSON in `.claude/` or `.ralph/` directories
- Session state in `.ralph/.ralph_session` with 24-hour expiration
- Checkpoints can be named: `ralph checkpoint save "checkpoint-name"`
- PreCompact event triggers state save; SessionStart triggers restore

**Decision**: Use JSON files in `.claude/authoring/` directory:

```
.claude/authoring/
├── checkpoint.json       # Current task state
├── session.json          # Session metadata
└── history/              # Completed task summaries
    └── 2026-02-04-exploration-fasttree.json
```

**Checkpoint Schema**:
```json
{
  "task_id": "uuid",
  "task_type": "exploration|survey|verification",
  "started_at": "ISO-8601",
  "last_checkpoint": "ISO-8601",
  "iteration_count": 5,
  "prompt_file": "tools/agents/exploration/PROMPT.md",
  "state": {
    "items_completed": ["item1", "item2"],
    "items_pending": ["item3"],
    "blocked_items": [],
    "current_focus": "searching FastTree for likelihood code"
  },
  "metrics": {
    "candidates_queued": 12,
    "papers_found": 3,
    "code_locations_found": 9
  }
}
```

**Rationale**: JSON is human-readable, git-friendly, and matches Claude Code plugin conventions. Separating history allows reviewing completed tasks.

---

### 4. Candidate Queue File Format

**Question**: JSONL vs single JSON file vs SQLite for queue storage?

**Finding**: Bipartite uses "git-versionable JSONL with ephemeral SQLite for queries" - meaning source data is JSONL, but a SQLite index is built for fast queries.

**Decision**: **JSONL** for candidate storage, matching bipartite's pattern:

```
.candidates/
├── queue.jsonl           # All candidates (append-only log)
├── rejected.jsonl        # Rejected candidates (for re-discovery prevention)
└── .cache/
    └── index.sqlite      # Ephemeral query index (gitignored)
```

**Format**:
```jsonl
{"id":"c-001","type":"paper","external_id":"S2:abc123","status":"pending","discovered_at":"...","discovered_by":"exploration-agent","notes":"Found in FastTree references"}
{"id":"c-002","type":"code-location","repo_url":"https://github.com/...","file_path":"src/likelihood.c","start_line":142,"end_line":180,"commit_sha":"abc123","status":"pending",...}
```

**Rationale**:
- JSONL is git-friendly (line-based diffs, easy merge)
- Append-only log preserves discovery history
- SQLite index is fast to rebuild from JSONL if corrupted
- Matches bipartite's proven pattern

---

## Summary of Decisions

| Question | Decision | Rationale |
|----------|----------|-----------|
| Code locations | Store in candidate queue; propose bipartite extension later | Immediate need is candidate review, not graph storage |
| Claim detection | Pattern-based heuristics with human fallback | ML is overkill for MVP; patterns catch obvious cases |
| Checkpoint format | JSON in `.claude/authoring/` | Matches plugin conventions; human-readable |
| Queue format | JSONL with SQLite index | Git-friendly; matches bipartite pattern |

## References

- [Citation prediction with transformers and NLP heuristics](https://www.sciencedirect.com/science/article/pii/S0306457323003205)
- [ralph-claude-code checkpoint implementation](https://github.com/frankbria/ralph-claude-code)
- [Official Ralph Wiggum plugin](https://github.com/anthropics/claude-code/blob/main/plugins/ralph-wiggum/README.md)
- bipartite CLI: `bip repo --help`, `bip --help`
