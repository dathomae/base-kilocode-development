#!/bin/sh
set -e

# This script runs tests for the chronocone_planning_language project.
# It is invoked from the Agent Manager toolbar (Cmd+E / Ctrl+E).
#
# Injected environment variables:
#   WORKTREE_PATH - Working directory of the selected worktree
#   REPO_PATH     - Repository root

cd "$WORKTREE_PATH"

echo "=== Running chronocone_planning_language tests ==="

# Run all Go tests in the CPL project
if [ -f "src/chronocone_planning_language/go.mod" ]; then
    cd src/chronocone_planning_language
    go test ./... -count=1 -timeout 120s
    cd "$WORKTREE_PATH"
fi

# Run backend Go tests if present
if [ -f "src/backend/go.mod" ]; then
    echo "=== Running backend tests ==="
    cd src/backend
    go test ./... -count=1 -timeout 120s
    cd "$WORKTREE_PATH"
fi

echo "=== All tests complete ==="
