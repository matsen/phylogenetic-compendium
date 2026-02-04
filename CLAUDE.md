# phylogenetic-compendium Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-02-04

## Active Technologies

- Go 1.24 + spf13/cobra (scribe CLI), bipartite CLI, Asta MCP, GitHub API (via gh), Quarto, Claude Code, Ralph Loop plugin (001-authoring-tooling)

## Project Structure

```text
scribe/           # Go CLI (scribe verify, queue, status, sweep)
  cmd/scribe/     # Main entry point
  internal/       # Internal packages (verify, queue, status, sweep, llm, output)
specs/            # Feature specifications
testdata/         # Test fixtures
```

## Commands

```bash
cd scribe && go build -o scribe ./cmd/scribe && ./scribe --help
go test ./...
go vet ./...
go fmt ./...
```

## Code Style

Go: Follow standard conventions (`go fmt`, `go vet`)

## Recent Changes

- 001-authoring-tooling: Added Go 1.24 + spf13/cobra (scribe CLI), bipartite CLI, Asta MCP, GitHub API (via gh), Quarto, Claude Code, Ralph Loop plugin

<!-- MANUAL ADDITIONS START -->

## Pre-PR Quality Checklist

Before any pull request, ensure the following workflow is completed:

### Requirement Verification (Do This First!)
1. **Spec Compliance**: Review the feature's `spec.md` and `tasks.md` to verify 100% completion of all specified requirements. If any requirement cannot be met, engage with the user to resolve blockers before proceeding

### Code Quality Foundation
2. **Format Code**: Run `go fmt ./...` to apply consistent formatting
3. **Documentation**: Ensure all exported functions and types have doc comments

### Architecture and Implementation Review
4. **Clean Code Review**: Run `@clean-code-reviewer` agent on all new/modified code for architectural review

### Test Quality Validation
5. **Test Implementation Audit**: Scan all test files for partially implemented tests or placeholder implementations. All tests must provide real validation
6. **Run Tests**: Ensure all tests pass: `go test ./...`

### Final Static Analysis
7. **Vet and Lint**: Run static analysis to verify code quality: `go vet ./...`

### Documentation Sync
8. **Documentation Update**: If the feature adds new commands or changes user-facing behavior, update relevant documentation

<!-- MANUAL ADDITIONS END -->
