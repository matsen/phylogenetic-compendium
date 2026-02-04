<!--
Sync Impact Report
==================
Version change: 2.1.0 → 2.2.0 (minor: added code location requirements to Principle I)
Modified principles: I. Every Claim Must Be Traceable (expanded for code locations)
Added sections: None
Removed sections: None
Templates requiring updates:
  - .specify/templates/plan-template.md: ⚠ pending (needs verification task structure)
  - .specify/templates/spec-template.md: ⚠ pending (needs discovery/verification scenarios)
  - .specify/templates/tasks-template.md: ⚠ pending (needs review agent task types)
Follow-up TODOs:
  - TODO(AGENT_PROMPTS): Define standard prompts for discovery and verification agents
  - TODO(VERIFICATION_SCRIPTS): Create automated verification tooling
  - TODO(QUERY_SCHEMA): Design the structured schema for technique entries (downstream spec)
  - TODO(BIPARTITE_CODE_LOCATIONS): Extend bipartite repo nodes to support code location metadata
-->

# Compendium Authoring System Constitution

This constitution defines the system for writing rigorous, verifiable technical compendiums. It governs how agents and humans collaborate to discover, verify, and integrate knowledge—not the structure of any particular book.

## Core Principles

### I. Every Claim Must Be Traceable

All factual claims MUST trace to verifiable sources with explicit provenance.

- Every technical claim MUST cite at least one paper via bipartite paper ID
- Claims about software MUST link to bipartite repo nodes with verification metadata
- Paraphrased content MUST reference the source paper and relevant section/figure
- No claim may exist without a traceable source—unsourced claims MUST be flagged
- Agents MUST be able to programmatically verify any claim by traversing to its source
- Implementation references SHOULD include specific code locations (file path, function/class name, or line range) where practical, not just repository URLs
- Code references SHOULD include a commit hash or version tag to prevent link rot as code evolves
- Code location links SHOULD use permalink format (e.g., GitHub blob URLs with commit SHA)

**Rationale**: Compendiums rot when claims drift from sources. Traceability enables automated verification and prevents silent degradation of accuracy. Pointing to a repository is insufficient—readers and agents need to find the actual implementation, not search through thousands of files.

### II. Continuous Verification

Agents MUST continuously verify claims against primary sources.

- **Citation verification**: Every paper ID MUST resolve in bipartite; broken references block publication
- **Claim verification**: Agents MUST periodically re-check claims against cited papers via Asta snippet search
- **Implementation verification**: Repo URLs MUST be checked for accessibility; archived/dead projects flagged
- **Cross-reference verification**: Related claims MUST be checked for consistency
- Verification failures MUST be logged, tracked, and surfaced—never silently ignored

**Rationale**: Manual review cannot scale. Automated verification catches errors before readers do.

### III. Systematic Discovery

Agents MUST systematically expand coverage through structured literature exploration.

- **Citation chasing**: For each paper, agents MUST examine references and citations via Asta/S2
- **Concept expansion**: For each technique, agents MUST search for variant names, related methods, implementations
- **Gap detection**: Agents MUST identify missing coverage (techniques mentioned but not explained, papers cited but not integrated)
- **Recency checking**: Agents MUST periodically search for new papers on covered topics
- Discovery findings MUST be logged as candidates for human review, not auto-integrated

**Rationale**: No human can comprehensively survey a field. Agents can systematically explore the literature graph.

### IV. Bipartite as Source of Truth

All references MUST flow through bipartite—no parallel reference systems.

- Papers MUST be added to bipartite before citation (via `bip s2 add` or `bip import`)
- Concepts/techniques MUST be registered as bipartite concept nodes
- Software tools MUST be registered as bipartite repo nodes
- Paper-concept edges MUST be maintained in bipartite's knowledge graph
- BibTeX for publication MUST be generated from bipartite export, never hand-maintained

**Rationale**: Parallel reference systems drift. Bipartite provides the single source of truth that agents can traverse and verify.

### V. Human Review Gates

Agents discover and verify; humans approve integration.

- Agents MUST NOT auto-commit content changes without human review
- Discovery candidates MUST be queued for human evaluation
- Verification failures MUST be escalated to humans with context
- Structural changes (new sections, reorganization) MUST require explicit human approval
- Human reviewers MUST verify agent work samples, not rubber-stamp

**Rationale**: Agents make mistakes. Human gates catch errors and maintain editorial judgment.

### VI. Agentic Consumability

Content MUST be structured for programmatic query by agent consumers, not just human readers.

- Technique entries MUST have machine-readable metadata, not just prose
- Implementations MUST expose queryable properties (language, complexity, constraints, tradeoffs)
- Comparisons and tradeoffs MUST be structured so agents can filter and rank options
- An agent asked "what data structure should I use for X?" MUST be able to:
  - Discover relevant technique entries programmatically
  - Extract and compare implementation properties
  - Find papers that analyze tradeoffs
  - Synthesize a recommendation with citations
- The specific schema for structured entries is deferred to compendium-level specification

**Rationale**: A compendium that only humans can navigate fails half its audience. Agents need structured data to answer decision questions, not just prose to summarize.

## Agentic Discovery Workflow

### Literature Expansion Loop

Agents execute this loop to systematically expand coverage:

```
1. SELECT a frontier paper (cited but not yet processed)
2. FETCH paper details via bipartite/Asta (abstract, references, citations)
3. FOR each reference:
   a. CHECK if already in bipartite
   b. IF relevant AND missing: ADD to candidate queue
4. FOR each citing paper (via Asta get_citations):
   a. CHECK if already in bipartite
   b. IF relevant AND missing: ADD to candidate queue
5. EXTRACT concepts/techniques mentioned in abstract
6. FOR each concept:
   a. CHECK if bipartite concept node exists
   b. IF missing: ADD to concept candidate queue
   c. CREATE edge from paper to concept
7. LOG expansion results
8. REPEAT with next frontier paper
```

### Implementation Survey Loop

Agents execute this loop to find software implementations:

```
1. SELECT a technique/concept node
2. SEARCH GitHub/repos for implementations (via keyword, paper citations)
3. FOR each candidate repo:
   a. CHECK if already in bipartite as repo node
   b. VERIFY repo is accessible and active
   c. EXTRACT metadata (language, last commit, stars)
   d. IF relevant AND missing: ADD to repo candidate queue
4. LOG survey results with verification status
5. REPEAT with next concept
```

### Verification Sweep

Agents execute periodic verification:

```
1. FOR each paper reference in content:
   a. VERIFY paper ID resolves in bipartite
   b. VERIFY bipartite entry has required fields (title, authors, year)
   c. FLAG any failures
2. FOR each claim with citation:
   a. FETCH relevant snippet from paper via Asta
   b. CHECK claim consistency with source
   c. FLAG discrepancies for human review
3. FOR each repo reference:
   a. CHECK URL accessibility
   b. CHECK last commit date (flag if >2 years stale)
   c. FLAG archived/deleted repos
4. GENERATE verification report
```

## Verification Pipeline

### Pre-Commit Checks (Blocking)

These MUST pass before any content merge:

- [ ] All paper IDs resolve in bipartite
- [ ] All repo URLs return 200 status
- [ ] No orphaned citations (cited but not in bipartite)
- [ ] All claims have at least one citation
- [ ] No TODO/FIXME markers in publishable content

### Periodic Checks (Non-Blocking, Logged)

These run on schedule and generate reports:

- [ ] Claim-source consistency via Asta snippet matching
- [ ] Repo freshness (last commit within 2 years)
- [ ] Coverage gaps (concepts without implementations, papers without concept edges)
- [ ] New papers on covered topics (recency check)

### Verification Metadata

Each verifiable element tracks:

```yaml
verified_at: ISO-8601 timestamp
verified_by: agent-id or human-id
verification_type: [citation|claim|repo|link]
status: [passed|failed|pending]
notes: optional context
```

## Technology Requirements

### Bipartite Integration
- bipartite CLI MUST be available in authoring environment
- Asta MCP MUST be configured for snippet search and citations
- Semantic Scholar API MUST be configured for paper lookup

### Agent Infrastructure
- Agents MUST log all discovery and verification actions
- Agent outputs MUST be structured (JSON) for downstream processing
- Agent candidate queues MUST be persistent and reviewable

### Content Format
- Quarto/Markdown for human+agent readable source
- YAML frontmatter for machine-parseable metadata
- Stable identifiers for cross-referencing

## Governance

This constitution governs the authoring system. Individual compendiums may add domain-specific requirements but MUST NOT weaken verification standards.

### Amendment Process

1. Propose amendment with rationale
2. Assess impact on verification rigor
3. Update constitution with version bump
4. Update agent workflows if affected

### Versioning Policy

- **MAJOR**: Changes to verification requirements or agent workflows
- **MINOR**: New verification checks or discovery strategies added
- **PATCH**: Clarifications, wording improvements

### Compliance

- All content changes MUST pass pre-commit verification
- Verification failures MUST be resolved or explicitly documented
- Agent workflows MUST log actions for audit

**Version**: 2.2.0 | **Ratified**: 2026-02-04 | **Last Amended**: 2026-02-04
