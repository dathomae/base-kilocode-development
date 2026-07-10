#!/bin/sh
set -e

# This script runs automatically when a new Agent Manager worktree is created.
# It prepares the worktree environment so agents can build, test, and run the project.
#
# Injected environment variables:
#   WORKTREE_PATH - Absolute path to the new worktree directory
#   REPO_PATH     - Repository root

echo "[kilo:setup] Initializing worktree at $WORKTREE_PATH"

cd "$WORKTREE_PATH"

# --- Go module dependencies (chronocone_planning_language) ---
if [ -f "src/chronocone_planning_language/go.mod" ]; then
    echo "[kilo:setup] Downloading Go module dependencies..."
    cd src/chronocone_planning_language
    go mod download
    cd "$WORKTREE_PATH"
fi

# --- Go module dependencies (backend) ---
if [ -f "src/backend/go.mod" ]; then
    echo "[kilo:setup] Downloading backend Go module dependencies..."
    cd src/backend
    go mod download
    cd "$WORKTREE_PATH"
fi

# --- Python setup (if setup.py or pyproject.toml exists) ---
if [ -f "src/setup.py" ] || [ -f "pyproject.toml" ]; then
    echo "[kilo:setup] Installing Python dependencies..."
    pip install -e src/ 2>/dev/null || pip install -e . 2>/dev/null || true
fi

echo "[kilo:setup] Worktree initialization complete."
