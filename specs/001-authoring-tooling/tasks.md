# Tasks: Compendium Authoring Tooling

**Input**: Design documents from `/specs/001-authoring-tooling/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Tests will be added using Go's testing framework with testdata fixtures.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, etc.)
- Include exact file paths in descriptions

## Path Conventions

Per plan.md, this project uses:
- `scribe/` - Go CLI source at repository root
- `scribe/cmd/scribe/` - Entry point
- `scribe/internal/` - Internal packages
- `scribe/agents/` - Agent prompts (markdown)
- `testdata/` - Test fixtures

---

## Phase 1: Setup

**Purpose**: Project initialization and Go module setup

- [X] T001 Create `scribe/` directory structure per plan.md
- [X] T002 Initialize Go module in `scribe/go.mod` with module path `github.com/matsen/phylogenetic-compendium/scribe`
- [X] T003 [P] Create `scribe/cmd/scribe/main.go` with cobra root command and version flag
- [X] T004 [P] Add `.candidates/` and `.claude/authoring/` to `.gitignore`
- [X] T005 [P] Create `testdata/valid/` and `testdata/invalid/` directories with sample QMD files

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

### Data Types

- [X] T006 [P] Define Candidate struct and types in `scribe/internal/queue/types.go`
- [X] T007 [P] Define CodeLocation struct in `scribe/internal/verify/types.go`
- [X] T008 [P] Define VerificationResult and VerificationReport in `scribe/internal/verify/types.go`
- [X] T009 [P] Define TaskCheckpoint and AgentActionLog in `scribe/internal/status/types.go`

### Common Infrastructure

- [X] T010 Implement JSONL storage read/write utilities in `scribe/internal/queue/store.go`
- [X] T011 [P] Implement LLM client in `scribe/internal/llm/client.go` (prefer Claude Haiku via `claude` CLI; fallback to local Ollama if unavailable)
- [X] T012 [P] Create common CLI output utilities (JSON/human-readable) in `scribe/internal/output/format.go`

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Pre-Commit Verification (Priority: P1)

**Goal**: Authors can verify content before merging - all citations resolve, URLs accessible, code links valid, claims cited, no TODOs

**Independent Test**: Create a QMD file with known good and bad references, run `scribe verify`, confirm correct pass/fail results

### Tests for User Story 1

- [X] T013 [P] [US1] Create test fixtures in `testdata/valid/good-content.qmd` (all valid refs)
- [X] T014 [P] [US1] Create test fixtures in `testdata/invalid/bad-citations.qmd` (unknown paper IDs)
- [X] T015 [P] [US1] Create test fixtures in `testdata/invalid/bad-urls.qmd` (broken URLs)
- [X] T016 [P] [US1] Create test fixtures in `testdata/invalid/bad-codelinks.qmd` (invalid permalinks)
- [X] T017 [P] [US1] Create test fixtures in `testdata/invalid/uncited-claims.qmd` (factual claims without citations)
- [X] T018 [P] [US1] Create test fixtures in `testdata/invalid/todo-markers.qmd` (TODO/FIXME markers)
- [X] T019 [US1] Write verification tests in `scribe/internal/verify/verify_test.go`

### Implementation for User Story 1

- [X] T020 [US1] Implement citation verification (bip lookup) in `scribe/internal/verify/citations.go`
- [X] T021 [US1] Implement URL accessibility check in `scribe/internal/verify/urls.go`
- [X] T022 [US1] Implement code-link validation (GitHub API) in `scribe/internal/verify/codelinks.go`
- [X] T023 [US1] Implement claim detection heuristics (LLM shell-out) in `scribe/internal/verify/claims.go`
- [X] T024 [US1] Implement TODO/FIXME marker detection in `scribe/internal/verify/claims.go`
- [X] T025 [US1] Implement verification report generation in `scribe/internal/verify/report.go`
- [X] T026 [US1] Wire up `scribe verify` command in `scribe/cmd/scribe/verify.go`
- [X] T027 [US1] Add `--human`, `--summary`, and `--json` output flags to verify command
- [X] T028 [US1] Ensure non-zero exit code on verification failure (per FR-007)

**Checkpoint**: User Story 1 complete - `scribe verify` fully functional

---

## Phase 4: User Story 2 - Candidate Queue Management (Priority: P2)

**Goal**: Agents can queue discovered papers/code/concepts for human review; humans can list, approve, reject candidates

**Independent Test**: Add candidates to queue, list them, approve/reject, verify state changes correctly

### Tests for User Story 2

- [X] T029 [P] [US2] Create queue test fixtures in `testdata/queue/empty-queue.jsonl`
- [X] T030 [P] [US2] Create queue test fixtures in `testdata/queue/populated-queue.jsonl`
- [X] T031 [US2] Write queue operation tests in `scribe/internal/queue/candidates_test.go`

### Implementation for User Story 2

- [X] T032 [P] [US2] Implement candidate CRUD operations in `scribe/internal/queue/candidates.go`
- [X] T033 [US2] Implement ID generation (c-YYYYMMDDHHMM format) in `scribe/internal/queue/candidates.go`
- [X] T034 [US2] Implement rejected candidate tracking in `.candidates/rejected.jsonl`
- [X] T035 [US2] Implement duplicate/previously-rejected detection per FR-012
- [X] T036 [US2] Wire up `scribe queue add` command in `scribe/cmd/scribe/queue.go`
- [X] T037 [US2] Wire up `scribe queue list` command with filters in `scribe/cmd/scribe/queue.go`
- [X] T038 [US2] Wire up `scribe queue approve` command (calls bip add) in `scribe/cmd/scribe/queue.go`
- [X] T039 [US2] Wire up `scribe queue reject` command in `scribe/cmd/scribe/queue.go`
- [X] T040 [US2] Wire up `scribe queue get` and `scribe queue stats` commands in `scribe/cmd/scribe/queue.go`
- [X] T041 [US2] Add `--human` and `--json` output flags to all queue subcommands

**Checkpoint**: User Story 2 complete - `scribe queue` fully functional

---

## Phase 5: User Story 7 - Autonomous Long-Running Operation (Priority: P2)

**Goal**: Tasks can run autonomously for hours via Ralph Loop, checkpointing progress, handling context limits

**Independent Test**: Start a task, let it checkpoint, verify state persists across context resets

**Note**: This is listed as P2 because Exploration (P3) and Survey (P3) depend on autonomous operation

### Tests for User Story 7

- [X] T042 [P] [US7] Create checkpoint test fixtures in `testdata/checkpoints/in-progress.json`
- [X] T043 [US7] Write checkpoint tests in `scribe/internal/status/checkpoint_test.go`

### Implementation for User Story 7

- [X] T044 [US7] Implement TaskCheckpoint read/write in `scribe/internal/status/checkpoint.go` (enforce 5-minute minimum checkpoint interval per FR-044a)
- [X] T045 [US7] Implement AgentActionLog append in `scribe/internal/status/logging.go`
- [X] T046 [US7] Implement `scribe status` pretty-print in `scribe/internal/status/display.go`
- [X] T047 [US7] Wire up `scribe status` command in `scribe/cmd/scribe/status.go`
- [X] T048 [US7] Add duration, iteration, progress, cost display per FR-044b
- [X] T049 [US7] Add `--max-iterations` and `--cost-budget` flags for autonomous commands per FR-044 in `scribe/cmd/scribe/status.go`
- [X] T050 [US7] Implement blocking issue detection and queue-for-review logic per FR-042 in `scribe/internal/status/blocking.go`
- [X] T051 [US7] Implement incremental git commit during autonomous operation per FR-041 in `scribe/internal/status/commits.go`

**Checkpoint**: User Story 7 complete - `scribe status` functional, checkpoint infrastructure ready

---

## Phase 6: User Story 3 - Targeted Code Exploration (Priority: P3)

**Goal**: Agent explores a codebase for undocumented techniques, traces provenance, finds papers, queues findings

**Independent Test**: Run agent on known codebase with technique hint, verify it finds relevant code locations and papers

**Depends on**: US2 (queue), US7 (autonomous operation)

### Implementation for User Story 3

- [X] T052 [US3] Create exploration agent prompt in `scribe/agents/exploration/PROMPT.md` covering FR-016 (accept target repo, technique hint, provenance hints)
- [X] T053 [US3] Document codebase search instructions per FR-017 (search for functions/files matching keywords) and FR-018 (extract code context)
- [X] T054 [US3] Document permalink generation per FR-019 and provenance search per FR-020 in exploration prompt
- [X] T055 [US3] Add Asta MCP usage instructions for paper discovery per FR-021 to exploration prompt
- [X] T056 [US3] Add queue integration instructions per FR-022 (queue all findings) and document consulting `rejected.jsonl` per FR-012a
- [X] T057 [US3] Add checkpoint integration instructions to exploration prompt (use scribe status)

**Checkpoint**: User Story 3 complete - exploration agent prompt ready for Ralph Loop

---

## Phase 7: User Story 4 - Comparative Synthesis Survey (Priority: P3)

**Goal**: Agent surveys multiple codebases for how they handle a concept, produces structured comparison

**Independent Test**: Run survey on 2-3 known repos with specific concept, verify comparable findings from each

**Depends on**: US2 (queue), US7 (autonomous operation)

### Implementation for User Story 4

- [X] T058 [US4] Create survey agent prompt in `scribe/agents/survey/PROMPT.md` covering FR-023 (accept concept and repo list)
- [X] T059 [US4] Document systematic search instructions per FR-024 and prioritization per FR-026 (flag repos as "especially interesting")
- [X] T060 [US4] Add structured comparison output format per FR-025 and FR-027 (comparable format with code location links)
- [X] T061 [US4] Add Asta search instructions per FR-028 for finding papers discussing discovered approaches
- [X] T062 [US4] Add queue integration and document consulting `rejected.jsonl` per FR-012a in survey prompt
- [X] T063 [US4] Add checkpoint integration instructions to survey prompt
- [X] T064 [US4] Define SurveyFinding output format in survey prompt

**Checkpoint**: User Story 4 complete - survey agent prompt ready for Ralph Loop

---

## Phase 8: User Story 5 - Agentic Consumer Query (Priority: P3)

**Goal**: Coding agents can query compendium for structured technique information, filter by language, get recommendations

**Independent Test**: Have agent query compendium for known topic, verify it extracts structured actionable information

**Depends on**: Compendium content structure (assumes QMD with YAML frontmatter)

### Implementation for User Story 5

- [X] T065 [US5] Create consumer agent prompt in `scribe/agents/consumer/PROMPT.md` covering FR-029 (queryable by technique/concept)
- [X] T066 [US5] Document query patterns per FR-030 (expose language, data structures, tradeoffs metadata)
- [X] T067 [US5] Add code location retrieval instructions per FR-031 to consumer prompt
- [X] T068 [US5] Add citation extraction instructions per FR-032 (link to papers for citations)
- [X] T069 [US5] Add recommendation synthesis format to consumer prompt

**Checkpoint**: User Story 5 complete - consumer agent prompt ready

---

## Phase 9: User Story 6 - Periodic Verification Sweep (Priority: P4)

**Goal**: Maintainer runs sweep to check claim-source consistency, repo freshness, code link validity, coverage gaps

**Independent Test**: Run on content with known stale repos or moved code, verify they appear in report

**Depends on**: US1 (basic verification infrastructure)

### Tests for User Story 6

- [X] T070 [P] [US6] Create sweep test fixtures in `testdata/sweep/stale-repo.qmd`
- [X] T071 [P] [US6] Create sweep test fixtures in `testdata/sweep/moved-code.qmd`
- [X] T072 [US6] Write sweep tests in `scribe/internal/sweep/sweep_test.go`

### Implementation for User Story 6

- [X] T073 [US6] Implement claim-consistency check (Asta snippet search) per FR-033 in `scribe/internal/sweep/consistency.go`
- [X] T074 [US6] Implement repo freshness check (>2 years stale) per FR-034 in `scribe/internal/sweep/freshness.go`
- [X] T075 [US6] Implement code-link HEAD check (file/lines still exist) per FR-035 in `scribe/internal/sweep/codelinks.go`
- [X] T076 [US6] Implement coverage gap detection per FR-036 in `scribe/internal/sweep/coverage.go`
- [X] T077 [US6] Implement sweep report generation per FR-037 in `scribe/internal/sweep/report.go`
- [X] T078 [US6] Wire up `scribe sweep` command in `scribe/cmd/scribe/sweep.go`
- [X] T079 [US6] Add `--check` flag for selective checks (repo-freshness, claim-consistency, etc.)

**Checkpoint**: User Story 6 complete - `scribe sweep` fully functional

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [X] T080 [P] Add pre-commit hook example in `scribe/examples/pre-commit-hook.sh`
- [X] T081 [P] Update quickstart.md with actual command examples after implementation
- [X] T082 Run all tests and fix any failures
- [X] T083 Run verification on testdata fixtures to validate scribe verify works
- [X] T084 Manual test of queue workflow (add, list, approve, reject)
- [X] T085 Verify JSON output is parseable by jq for all commands

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **US1 Verification (Phase 3)**: Depends on Foundational (T006-T012)
- **US2 Queue (Phase 4)**: Depends on Foundational (T006, T010)
- **US7 Autonomous (Phase 5)**: Depends on Foundational (T009, T010)
- **US3 Exploration (Phase 6)**: Depends on US2 + US7 (T032-T041, T044-T051)
- **US4 Survey (Phase 7)**: Depends on US2 + US7
- **US5 Consumer (Phase 8)**: Depends on content structure (can start after Foundational)
- **US6 Sweep (Phase 9)**: Depends on US1 (verification infrastructure)
- **Polish (Phase 10)**: Depends on all user stories

### User Story Dependencies

```
Phase 1: Setup
    ↓
Phase 2: Foundational
    ↓
    ├─→ US1 (P1): Verification ─────────────→ US6 (P4): Sweep
    ├─→ US2 (P2): Queue ─────┬─→ US3 (P3): Exploration
    ├─→ US7 (P2): Autonomous ┘─→ US4 (P3): Survey
    └─→ US5 (P3): Consumer (independent)
```

### Within Each User Story

1. Tests fixtures first (to understand expected behavior)
2. Tests code (to define interface)
3. Implementation (to make tests pass)
4. CLI wiring (to expose functionality)

### Parallel Opportunities

**Phase 1 (all parallel)**:
- T001, T002, T003, T004, T005

**Phase 2 (types parallel, then infra)**:
- T006, T007, T008, T009 (parallel)
- T010, T011, T012 (parallel after types)

**Phase 3 (fixtures parallel)**:
- T013-T018 (all parallel)
- T019 after fixtures
- T020-T028 (implementation after tests)

**Phase 4 (fixtures parallel)**:
- T029, T030 (parallel)
- T032-T041 (sequential)

**Phase 5 (US7)**:
- T042-T043 (fixtures parallel)
- T044-T051 (sequential)

**Multi-story parallel**:
- Once Phase 2 complete, US1, US2, US7 can proceed in parallel
- Once US2+US7 complete, US3 and US4 can proceed in parallel

---

## Parallel Example: Phase 2 Foundation

```bash
# Launch all type definitions in parallel:
Task: "Define Candidate struct in scribe/internal/queue/types.go"
Task: "Define CodeLocation struct in scribe/internal/verify/types.go"
Task: "Define VerificationResult in scribe/internal/verify/types.go"
Task: "Define TaskCheckpoint in scribe/internal/status/types.go"

# Then launch infrastructure in parallel:
Task: "Implement JSONL storage in scribe/internal/queue/store.go"
Task: "Implement LLM client in scribe/internal/llm/client.go"
Task: "Create CLI output utilities in scribe/internal/output/format.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: US1 (Pre-Commit Verification)
4. **STOP and VALIDATE**: Run `scribe verify` on test fixtures
5. Deploy as pre-commit hook

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 (Verification) → Test → pre-commit hook works (MVP!)
3. Add US2 (Queue) + US7 (Autonomous) → Test → queue management works
4. Add US3 (Exploration) + US4 (Survey) → Test → agent prompts ready
5. Add US5 (Consumer) → Test → agentic queries work
6. Add US6 (Sweep) → Test → periodic verification works

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: US1 (Verification)
   - Developer B: US2 (Queue) + US7 (Autonomous)
3. After dependencies satisfied:
   - Developer A: US6 (Sweep) - builds on US1
   - Developer B: US3 + US4 (Exploration/Survey prompts)
   - Developer C: US5 (Consumer prompt)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Agent prompts (US3, US4, US5) are markdown files, not Go code
- All CLI commands output JSON by default for agent integration
- `--human` flag available on all commands for terminal use
- **Total tasks**: 85 (increased from 80 to cover all FRs)
