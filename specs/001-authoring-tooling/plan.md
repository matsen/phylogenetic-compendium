# Implementation Plan: Compendium Authoring Tooling

**Branch**: `001-authoring-tooling` | **Date**: 2026-02-04 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-authoring-tooling/spec.md`

## Summary

Build `scribe`, a unified CLI for compendium authoring that implements the constitution's discovery, verification, and human-review workflows. The system enables autonomous multi-hour operation via Ralph Loop for overnight exploration tasks.

## Technical Context

**Language/Version**: Go 1.24+ (consistent with bipartite)
**Primary Dependencies**: bipartite CLI, Asta MCP, GitHub API (via gh), Quarto, Claude Code, Ralph Loop plugin
**NLP Strategy**: Shell out to LLM for language tasks (claim detection, relevance scoring); supports Claude Haiku or local Ollama
**Storage**: JSON/JSONL files for candidates and checkpoints, YAML frontmatter in QMD content
**Testing**: Go testing + testdata fixtures
**Target Platform**: macOS/Linux CLI (single binary distribution)
**Project Type**: Single CLI tooling project
**Performance Goals**: Pre-commit verification <30s for 100 citations; exploration of 100k LOC repo <5min
**Constraints**: Must work offline for cached operations; graceful degradation on rate limits
**Scale/Scope**: Support compendiums up to 100 pages, candidate queues up to 1000 items

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Implementation |
|-----------|--------|----------------|
| I. Every Claim Must Be Traceable | ✅ PASS | FR-001 through FR-004 implement citation and code location verification |
| II. Continuous Verification | ✅ PASS | FR-033 through FR-037 implement periodic verification sweep |
| III. Systematic Discovery | ✅ PASS | FR-016 through FR-028 implement exploration and survey agents |
| IV. Bipartite as Source of Truth | ✅ PASS | FR-001 verifies paper IDs against bipartite; FR-014 adds approved candidates to bipartite |
| V. Human Review Gates | ✅ PASS | FR-008 through FR-015 implement candidate queue with human approval flow |
| VI. Agentic Consumability | ✅ PASS | FR-029 through FR-032 ensure structured, queryable entries |

**Gate Result**: ✅ All principles satisfied. Proceed to Phase 0.

## Project Structure

### Documentation (this feature)

```text
specs/001-authoring-tooling/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (via /speckit.tasks)
```

### Source Code (repository root)

```text
scribe/
├── cmd/scribe/
│   └── main.go                # Entry point
│
├── internal/
│   ├── verify/                # `scribe verify` (P1)
│   │   ├── citations.go       # Paper ID resolution via bip
│   │   ├── urls.go            # Repository URL accessibility
│   │   ├── codelinks.go       # GitHub permalink validation
│   │   ├── claims.go          # Uncited claim detection (shells to LLM)
│   │   └── report.go          # Verification report generation
│   │
│   ├── queue/                 # `scribe queue` (P2)
│   │   ├── store.go           # JSONL storage operations
│   │   ├── candidates.go      # Candidate CRUD operations
│   │   └── types.go           # Candidate structs
│   │
│   ├── sweep/                 # `scribe sweep` (P4)
│   │   ├── consistency.go     # Claim-source consistency (shells to LLM)
│   │   ├── freshness.go       # Repo freshness checks
│   │   └── report.go          # Sweep report generation
│   │
│   ├── status/                # `scribe status` (P2)
│   │   └── display.go         # Pretty-print checkpoint.json
│   │
│   └── llm/                   # LLM integration
│       └── client.go          # Shell to claude/ollama for NLP tasks
│
├── agents/                    # Agent prompts (P3) - not Go code
│   ├── exploration/
│   │   └── PROMPT.md          # Agent prompt for Ralph Loop
│   ├── survey/
│   │   └── PROMPT.md          # Agent prompt for Ralph Loop
│   └── consumer/
│       └── PROMPT.md          # Agent prompt for compendium queries
│
├── go.mod
└── go.sum

testdata/
├── valid/                     # Test content with good refs
├── invalid/                   # Test content with bad refs
├── queue/                     # Test queue fixtures
├── checkpoints/               # Test checkpoint fixtures
└── sweep/                     # Test sweep fixtures
```

**Structure Decision**: Go CLI (`scribe`) following bip's patterns. Shells out to Claude Haiku or Ollama for NLP tasks. Agent prompts remain as markdown for Ralph Loop.

## Complexity Tracking

No violations. Structure follows YAGNI—Go CLI for all commands, shelling to LLM for NLP tasks, prompts where agents do the work.

## Phase 0: Research Tasks

Based on Technical Context analysis, the following require clarification:

1. **Bipartite code location support**: Does bipartite currently support code location metadata on repo nodes, or does this need extension?
2. **Claim detection heuristics**: How to distinguish factual claims requiring citations from definitions, examples, and transitional prose?
3. **Ralph Loop checkpoint format**: What format does the Ralph Loop plugin expect for checkpoints?
4. **Candidate queue file format**: JSONL vs single JSON file vs SQLite for queue storage?

## Phase 1: Design Outputs

✅ **Completed**:
- `research.md` - Decisions on bipartite code locations, claim detection, checkpoint format, queue format
- `data-model.md` - Candidate, CodeLocation, VerificationResult, TaskCheckpoint, AgentActionLog schemas
- `contracts/queue-cli.md` - CLI interface contract for queue operations
- `contracts/verification-report.md` - Structured output format for verification
- `quickstart.md` - How to run verification and queue operations

## Post-Design Constitution Check

*Re-evaluated after Phase 1 design completion.*

| Principle | Status | Verification |
|-----------|--------|--------------|
| I. Every Claim Must Be Traceable | ✅ PASS | Verification report contract defines citation, URL, code-link checks |
| II. Continuous Verification | ✅ PASS | Periodic sweep checks defined in verification-report.md |
| III. Systematic Discovery | ✅ PASS | Exploration and survey agent prompts specified in quickstart |
| IV. Bipartite as Source of Truth | ✅ PASS | Queue approve triggers `bip s2 add` / `bip repo add` |
| V. Human Review Gates | ✅ PASS | Candidate queue with approve/reject workflow defined |
| VI. Agentic Consumability | ✅ PASS | All outputs are JSON by default; structured schemas defined |

**Post-Design Gate Result**: ✅ All principles satisfied. Ready for Phase 2 task generation.

## Next Steps

Run `/speckit.tasks` to generate implementation tasks from this plan.
