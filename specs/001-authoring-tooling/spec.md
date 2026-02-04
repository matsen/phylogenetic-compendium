# Feature Specification: Compendium Authoring Tooling

**Feature Branch**: `001-authoring-tooling`
**Created**: 2026-02-04
**Status**: Draft
**Input**: Verification scripts, candidate queues, and agent prompts that implement the constitution's workflows

## Scope

**In scope**:
- Verification scripts, candidate queues, agent prompts
- Content generation (agents write compendium prose)
- Proposing bipartite schema extensions (e.g., code-location nodes)
- Git-based concurrent workflows (JSONL storage for mergeability)

**Out of scope**:
- Real-time collaboration (no live multi-user editing; use git branches/merges instead)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Pre-Commit Verification (Priority: P1)

An author (human or agent) has drafted new compendium content with citations and wants to merge it. Before the merge can proceed, the system automatically verifies that all paper IDs resolve in bipartite, all repository URLs are accessible, all code location links are valid (file exists, line range valid), and every factual claim has at least one citation. The author receives a clear pass/fail report.

**Why this priority**: This is the enforcement mechanism for the constitution's traceability principle. Without pre-commit checks, unverified content enters the compendium and degrades quality. This is the minimum viable product for the authoring system.

**Independent Test**: Can be fully tested by creating a content file with known good and bad references, running verification, and confirming correct pass/fail results.

**Acceptance Scenarios**:

1. **Given** content with all valid bipartite paper IDs, **When** pre-commit verification runs, **Then** verification passes with a success report
2. **Given** content with a paper ID not in bipartite, **When** pre-commit verification runs, **Then** verification fails listing the unresolved ID
3. **Given** content with a repository URL that returns 404, **When** pre-commit verification runs, **Then** verification fails listing the broken URL
4. **Given** content with a code location link (e.g., GitHub permalink), **When** pre-commit verification runs, **Then** it verifies the file and line range exist
5. **Given** content with an uncited factual claim, **When** pre-commit verification runs, **Then** verification fails identifying the uncited claim
6. **Given** content with TODO/FIXME markers, **When** pre-commit verification runs, **Then** verification fails listing the marker locations

---

### User Story 2 - Candidate Queue Management (Priority: P2)

An agent running exploration discovers papers, code locations, or concepts that might be relevant to the compendium. Instead of auto-adding them, the agent queues these as candidates for human review. A human author can list pending candidates, review their relevance, and approve or reject them. Approved candidates are added to bipartite; rejected ones are marked to prevent re-discovery.

**Why this priority**: The constitution requires human review gates. Without a candidate queue, agents cannot propose discoveries and humans cannot review them. This enables all discovery workflows.

**Independent Test**: Can be tested by adding candidates to the queue, listing them, and approving/rejecting them—verifying state changes correctly.

**Acceptance Scenarios**:

1. **Given** an empty candidate queue, **When** an agent adds a paper candidate with S2 ID and relevance notes, **Then** the candidate appears in the pending queue
2. **Given** an empty candidate queue, **When** an agent adds a code location candidate with repo, file path, line range, and description, **Then** the candidate appears in the pending queue
3. **Given** pending candidates in the queue, **When** a human lists candidates, **Then** all pending candidates are displayed with their metadata and agent notes
4. **Given** a pending paper candidate, **When** a human approves it, **Then** the paper is added to bipartite and the candidate is marked approved
5. **Given** a pending candidate, **When** a human rejects it with a reason, **Then** the candidate is marked rejected with the reason stored
6. **Given** a previously rejected item, **When** an agent attempts to re-queue it, **Then** the system warns that this candidate was previously rejected

---

### User Story 3 - Targeted Code Exploration (Priority: P3)

An author knows that a specific technique exists in a codebase but is undocumented. For example: "FastTree does something clever with rotated likelihood vectors, a feature it got from RAxML, but it's undocumented." The author invokes an exploration agent with this context. The agent searches the FastTree codebase for relevant code, traces provenance by searching RAxML for similar patterns, finds related papers via Asta, and queues findings (code locations, papers, concepts) for the author to review and synthesize into a compendium section.

**Why this priority**: This is the core authoring workflow—turning undocumented tribal knowledge into rigorous, cited compendium content. Depends on candidate queue (P2).

**Independent Test**: Can be tested by running the agent on a known codebase with a specific technique hint and verifying it finds relevant code locations and papers.

**Acceptance Scenarios**:

1. **Given** a target repo (FastTree) and technique hint ("rotated likelihood vectors"), **When** the exploration agent runs, **Then** it searches the codebase for relevant functions/files
2. **Given** code locations found in target repo, **When** the agent processes them, **Then** it extracts function names, comments, and surrounding context
3. **Given** a provenance hint ("came from RAxML"), **When** the agent processes it, **Then** it searches the referenced repo for similar patterns
4. **Given** technique keywords, **When** the agent searches Asta, **Then** it finds papers discussing the technique and queues relevant ones
5. **Given** all findings, **When** exploration completes, **Then** the agent queues code location candidates with permalinks, paper candidates, and concept candidates for human review
6. **Given** queued candidates, **When** the author reviews them, **Then** they have enough context to write a compendium section with proper citations and code links

---

### User Story 4 - Comparative Synthesis Survey (Priority: P3)

An author wants to understand how multiple implementations handle a common problem. For example: "Every phylogenetic program does some sort of caching of partial results and needs to track what state is valid. Let's look at all the programs and figure out what they do—I think bali-phy does something especially clever." The author invokes a survey agent with the concept and a list of repos to examine. The agent systematically searches each codebase, extracts how each handles the problem, identifies interesting variations, and queues structured findings for the author to synthesize into a comparative section.

**Why this priority**: This enables the comparative, cross-implementation content that makes a compendium valuable. Depends on candidate queue (P2).

**Independent Test**: Can be tested by running the survey on 2-3 known repos with a specific concept and verifying it extracts comparable findings from each.

**Acceptance Scenarios**:

1. **Given** a concept ("partial result caching/invalidation") and list of repos (RAxML, FastTree, bali-phy, IQ-TREE), **When** the survey agent runs, **Then** it searches each repo for relevant code
2. **Given** findings from multiple repos, **When** the agent processes them, **Then** it structures findings in a comparable format (repo, approach, code location, key insight)
3. **Given** a hint about a specific repo ("bali-phy does something clever"), **When** the agent processes that repo, **Then** it gives extra attention and detail to findings from that repo
4. **Given** technique keywords from findings, **When** the agent searches Asta, **Then** it finds papers discussing the approaches and queues relevant ones
5. **Given** all findings, **When** survey completes, **Then** the agent produces a structured comparison table/summary with code location links for each implementation
6. **Given** the comparison, **When** the author reviews it, **Then** they can write a synthesis section showing how different programs solve the same problem

---

### User Story 5 - Agentic Consumer Query (Priority: P3)

A coding agent is implementing phylogenetic software and needs to make a design decision. For example: "Given that we are using Zig, think about all the options for implementing the phylogenetic tree structure and choose the one that will be efficient, safe, and easy to work with." The agent queries the compendium to find relevant technique entries, extracts structured information about implementations (language, data structures, tradeoffs), filters by language compatibility, and synthesizes a recommendation with citations to papers and code examples.

**Why this priority**: This validates that authored content is actually consumable by its intended agentic audience. Including this in authoring tooling ensures we verify consumability during authoring, not after publication. If agents can't query the compendium effectively, the system fails Principle VI (Agentic Consumability).

**Independent Test**: Can be tested by having an agent query the compendium for a known topic and verifying it can extract structured, actionable information.

**Acceptance Scenarios**:

1. **Given** a compendium with structured technique entries, **When** an agent queries for "tree data structures", **Then** it finds relevant entries programmatically
2. **Given** technique entries with implementation metadata, **When** the agent filters by language (Zig), **Then** it identifies implementations in that language or with compatible patterns
3. **Given** implementation entries with tradeoff information, **When** the agent compares options, **Then** it can rank them by criteria (efficiency, safety, ergonomics)
4. **Given** a selected approach, **When** the agent retrieves details, **Then** it gets code location links showing how it's implemented elsewhere
5. **Given** papers linked to implementations, **When** the agent retrieves them, **Then** it can cite them in its recommendation
6. **Given** all information, **When** the agent synthesizes, **Then** it produces a recommendation with citations and code examples

---

### User Story 6 - Periodic Verification Sweep (Priority: P4)

A compendium maintainer wants to ensure existing content remains accurate over time. They run a periodic verification sweep that checks claim-source consistency (via Asta snippet matching), repo freshness, code location validity (files/lines still exist at HEAD), and coverage gaps. The sweep generates a report of issues requiring attention.

**Why this priority**: This implements ongoing verification per the constitution. Important for maintenance but not required for initial authoring.

**Independent Test**: Can be tested by running on content with known stale repos or moved code and verifying they appear in the report.

**Acceptance Scenarios**:

1. **Given** content with claims citing papers, **When** the verification sweep runs, **Then** it fetches relevant snippets via Asta and checks claim consistency
2. **Given** a claim that contradicts its cited source, **When** detected, **Then** it appears in the report with the claim text and source snippet
3. **Given** a repo reference with last commit >2 years ago, **When** the sweep checks freshness, **Then** it appears in the report as stale
4. **Given** a code location link, **When** the sweep checks it against current HEAD, **Then** it flags if the file or line range no longer exists
5. **Given** a concept node with no linked implementations, **When** the sweep checks coverage, **Then** it appears in the report as a coverage gap
6. **Given** sweep completion, **When** report is generated, **Then** issues are categorized by type and severity

---

### User Story 7 - Autonomous Long-Running Operation (Priority: P2)

An author wants to kick off a multi-hour exploration or survey task and walk away. They provide a task description and constraints, then the system runs autonomously—persisting progress, handling context limits, and resuming work—until the task is complete or human intervention is required. The author returns to find completed work, a summary of what was done, and any items queued for review.

**Why this priority**: Exploration and survey agents (P3) can take hours for large codebases. Without autonomous operation, an author must babysit the process. This enables "overnight" workflows that multiply authoring throughput.

**Independent Test**: Can be tested by starting an exploration task, letting it run through multiple context resets, and verifying work persists and completes.

**Acceptance Scenarios**:

1. **Given** a task with estimated multi-hour runtime, **When** the author starts it, **Then** the system runs without requiring human presence
2. **Given** an agent approaching context limits, **When** it detects this, **Then** it checkpoints progress and continues with fresh context
3. **Given** a running autonomous task, **When** the author checks status, **Then** they see progress summary and current state
4. **Given** an autonomous task that encounters a blocking issue, **When** detected, **Then** the system queues it for human review and continues with other work
5. **Given** an autonomous task that completes, **When** the author returns, **Then** they see a summary of work done and candidates queued for review
6. **Given** an autonomous task, **When** it runs, **Then** all actions are logged for audit and the work is committed incrementally

---

### Edge Cases

- What happens when bipartite is unavailable? Verification fails with a clear error indicating bipartite connectivity issue.
- What happens when Asta rate limits are hit? Agent pauses, logs the rate limit, and resumes after backoff period.
- What happens when GitHub API rate limits are hit during code location verification? System pauses, logs the rate limit, and resumes after backoff period; verification continues with remaining items.
- What happens when a paper ID format is invalid? Verification fails fast with format error before attempting lookup.
- What happens when the candidate queue file is corrupted? System detects corruption, refuses to proceed, and suggests recovery steps.
- How does the system handle papers with no abstract in Asta? Agent logs the gap and proceeds with available metadata.
- What happens when a code location link uses a commit SHA that no longer exists? Verification flags it; agent suggests updating to current HEAD.
- What happens when a repo has been renamed or moved? Agent detects redirect and suggests updating the reference.
- How does the system handle private repos? Agent notes access failure and flags for human review (may need authentication).
- What happens when an autonomous loop runs out of API budget? System checkpoints state and stops gracefully with clear cost report.
- What happens when an autonomous task hits an unrecoverable error? System logs the error, checkpoints progress, and queues issue for human review.
- How does the system handle multiple autonomous tasks running concurrently? Each task operates in its own branch to prevent conflicts.

## Requirements *(mandatory)*

### Functional Requirements

**Pre-Commit Verification**:
- **FR-001**: System MUST verify all paper IDs in content resolve to entries in bipartite
- **FR-002**: System MUST verify all repository URLs return successful HTTP status
- **FR-003**: System MUST verify all code location links are valid (commit SHA exists, file path exists at that SHA, line range is within file bounds)
- **FR-004**: System MUST verify every factual claim has at least one citation
- **FR-005**: System MUST detect and report TODO/FIXME markers in publishable content
- **FR-006**: System MUST produce a structured verification report (pass/fail with details)
- **FR-007**: System MUST exit with non-zero status on any verification failure

**Candidate Queue**:
- **FR-008**: System MUST support candidate types: paper, concept, repo, code-location
- **FR-009**: System MUST store candidates with metadata: type, external ID, agent notes, timestamp, status
- **FR-010**: Code location candidates MUST include: repo URL, file path, line range, commit SHA, description
- **FR-011**: System MUST support candidate statuses: pending, approved, rejected
- **FR-012**: System MUST persist rejected candidates with rejection reasons to prevent re-discovery
- **FR-012a**: Agents SHOULD consult rejected.jsonl to learn relevance patterns and avoid queuing similar candidates
- **FR-013**: System MUST allow listing candidates filtered by type and status
- **FR-014**: System MUST allow approving candidates (triggering addition to bipartite)
- **FR-015**: System MUST allow rejecting candidates with a reason

**Targeted Exploration Agent**:
- **FR-016**: Agent MUST accept a target repo, technique hint, and optional provenance hints
- **FR-017**: Agent MUST search codebase for functions/files matching technique keywords
- **FR-018**: Agent MUST extract code context (function signature, comments, surrounding code)
- **FR-019**: Agent MUST generate permalink URLs for found code locations
- **FR-020**: Agent MUST search provenance repos for similar patterns when hints provided
- **FR-021**: Agent MUST search Asta for papers related to discovered techniques
- **FR-022**: Agent MUST queue all findings (code locations, papers, concepts) for human review

**Comparative Survey Agent**:
- **FR-023**: Agent MUST accept a concept/technique and list of repos to survey
- **FR-024**: Agent MUST systematically search each repo for relevant implementations
- **FR-025**: Agent MUST structure findings in comparable format across repos
- **FR-026**: Agent MUST prioritize repos flagged as "especially interesting"
- **FR-027**: Agent MUST produce a comparison summary with code location links
- **FR-028**: Agent MUST search Asta for papers discussing discovered approaches

**Agentic Consumability**:
- **FR-029**: Compendium entries MUST be queryable by technique/concept name
- **FR-030**: Implementation entries MUST expose structured metadata (language, data structures, tradeoffs)
- **FR-031**: Entries MUST include code location links that agents can follow
- **FR-032**: Entries MUST link to papers that agents can retrieve for citations

**Periodic Verification**:
- **FR-033**: Verification sweep MUST check claim-source consistency via Asta snippet search
- **FR-034**: Verification sweep MUST check repository freshness (flag if >2 years since last commit)
- **FR-035**: Verification sweep MUST check code location links against current HEAD
- **FR-036**: Verification sweep MUST identify coverage gaps (concepts without implementations)
- **FR-037**: Verification sweep MUST generate a categorized report of issues

**Autonomous Long-Running Operation**:
- **FR-038**: System MUST support autonomous multi-hour operation via Ralph Loop plugin
- **FR-039**: System MUST checkpoint progress to disk before context resets
- **FR-040**: System MUST resume from checkpoint after context reset
- **FR-041**: System MUST commit work incrementally (not batch at end)
- **FR-042**: System MUST detect blocking issues and queue them for human review without stopping
- **FR-043**: System MUST produce a summary of completed work when task finishes
- **FR-044**: System MUST respect configurable iteration limits and cost budgets
- **FR-044a**: System MUST update checkpoint.json at least every 5 minutes during autonomous operation
- **FR-044b**: System MUST provide `scribe status` command that pretty-prints checkpoint.json (task name, duration, progress, candidates queued, cost)

**Logging and Audit**:
- **FR-045**: All agents MUST log actions in structured format for audit
- **FR-046**: Logs MUST include: agent type, action, target, result, timestamp

### Key Entities

- **Candidate**: A discovered item awaiting human review. Attributes: type (paper|concept|repo|code-location), external_id, agent_notes, discovered_at, discovered_by, status, review_notes, reviewed_at, reviewed_by
- **Code Location**: A specific location in a codebase. Attributes: repo_url, file_path, start_line, end_line, commit_sha, permalink_url, description, function_name
- **Verification Result**: Outcome of a verification check. Attributes: check_type, target_id, status (pass/fail), message, checked_at
- **Agent Action Log**: Record of an agent's action. Attributes: agent_type, action, target, result, timestamp
- **Survey Finding**: A structured finding from comparative survey. Attributes: repo, concept, approach_summary, code_location, key_insight, related_papers
- **Task Checkpoint**: Progress state for autonomous operation. Attributes: task_id, task_description, iteration_count, items_completed, items_pending, blocked_items, last_checkpoint_at, total_cost

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Pre-commit verification completes in under 30 seconds for content with up to 100 citations
- **SC-002**: Authors can review and process 20 candidates in under 10 minutes using the queue interface
- **SC-003**: Targeted exploration agent finds relevant code locations in a 100k LOC repo in under 5 minutes
- **SC-004**: Comparative survey agent completes survey of 5 repos for one concept in under 15 minutes
- **SC-005**: An agentic consumer can query the compendium and synthesize a recommendation in under 2 minutes
- **SC-006**: Periodic verification sweep completes for a 100-page compendium in under 15 minutes
- **SC-007**: 100% of verification failures include actionable information (what failed, where, why)
- **SC-008**: Zero false positives in pre-commit verification (valid content never incorrectly blocked)
- **SC-009**: Autonomous tasks can run for 3+ hours across multiple context resets without data loss
- **SC-010**: Autonomous tasks recover gracefully from transient failures (network, rate limits) without human intervention

## Clarifications

### Session 2026-02-04

- Q: How should authentication for external APIs (GitHub, Asta, S2) be handled? → A: Delegate to existing CLI auth (`gh auth`, `~/.config/bip/config.yml`)
- Q: What is explicitly out of scope? → A: Real-time collaboration only; content generation, bipartite schema proposals, and git-based concurrent workflows ARE in scope
- Q: How do users monitor autonomous task progress? → A: CLI status command that pretty-prints checkpoint.json (e.g., `authoring status`)
- Q: What makes a candidate "relevant"? → A: Liberal queuing with relevance notes; humans filter during review; rejection reasons feed back to inform future agent runs
- Q: How should CLI tools be organized? → A: Single CLI named `scribe` with subcommands (`scribe verify`, `scribe queue list`, `scribe status`)
- Q: What language for scribe CLI? → A: Go (consistent with bip); shell out to Claude Haiku or local Ollama for NLP tasks (claim detection, relevance scoring)

## Assumptions

- bipartite CLI is installed and configured with access to the nexus
- Authentication: System delegates to pre-configured CLI auth (`gh auth` for GitHub CLI, `~/.config/bip/config.yml` for S2/Asta API keys) rather than managing credentials directly
- bipartite repo nodes support code location metadata (file path, line range, commit SHA)—will be extended if needed
- Asta MCP is available for snippet search and citation lookup
- Semantic Scholar API is available for paper metadata
- GitHub API is available for code search and file retrieval
- Content follows Quarto/Markdown format with YAML frontmatter
- Citations use bipartite paper IDs (not raw DOIs or ad-hoc formats)
- Code location links use GitHub permalink format: `https://github.com/{org}/{repo}/blob/{sha}/{path}#L{start}-L{end}`
- Agents run in an environment with network access to GitHub and academic APIs
- Target repositories are public (private repos require additional authentication setup)
- Claude Code Ralph Loop plugin is available for autonomous long-running operation
