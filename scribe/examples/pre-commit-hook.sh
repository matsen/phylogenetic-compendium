#!/bin/bash
#
# Pre-commit hook for the Phylogenetic Compendium
# Install by copying to .git/hooks/pre-commit and making executable:
#   cp scribe/examples/pre-commit-hook.sh .git/hooks/pre-commit
#   chmod +x .git/hooks/pre-commit
#

set -e

# Find all staged QMD files
STAGED_QMD=$(git diff --cached --name-only --diff-filter=ACM | grep '\.qmd$' || true)

if [ -z "$STAGED_QMD" ]; then
    # No QMD files staged, skip verification
    exit 0
fi

echo "Running scribe verify on staged QMD files..."

# Run verification
# shellcheck disable=SC2086
if ! scribe verify --human $STAGED_QMD; then
    echo ""
    echo "Verification failed. Please fix the issues before committing."
    echo "To see details, run: scribe verify --human $STAGED_QMD"
    exit 1
fi

echo "Verification passed!"
exit 0
