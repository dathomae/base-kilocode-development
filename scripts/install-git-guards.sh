#!/usr/bin/env bash
#
# Install the local git guards (interim, until branch protection is available).
#
# This script installs the local core.hooksPath guard only.
# It is NOT a credential mechanism and configures no credential helper.
#
# Sets core.hooksPath=.githooks so the tracked pre-push hook runs in this clone and
# in every linked worktree (they share the repository config and the tracked .githooks
# directory is present in each checkout).
#
# Note: setting core.hooksPath disables the default .git/hooks directory. If you have
# hooks there you rely on, move them into .githooks/.
#
set -euo pipefail

if ! repo_root="$(git rev-parse --show-toplevel 2>/dev/null)"; then
  echo "error: not inside a git repository" >&2
  exit 1
fi

if [ -z "$repo_root" ] || [ ! -d "$repo_root" ]; then
  echo "error: invalid repository root: ${repo_root:-<empty>}" >&2
  exit 1
fi

cd "$repo_root"

hook_path="$repo_root/.githooks/pre-push"

if [ ! -f "$hook_path" ]; then
  echo "error: .githooks/pre-push not found in $repo_root" >&2
  exit 1
fi

if [ -L "$hook_path" ]; then
  echo "error: .githooks/pre-push must not be a symlink" >&2
  exit 1
fi

repo_real="$(cd "$repo_root" && pwd -P)"
hook_dir_real="$(cd "$(dirname "$hook_path")" && pwd -P)" || { echo "error: cannot resolve $hook_path" >&2; exit 1; }
hook_real="$hook_dir_real/$(basename "$hook_path")"
case "$hook_real" in
  "$repo_real"/*) ;;
  *) echo "error: refusing to chmod outside the repository root: $hook_real" >&2; exit 1 ;;
esac

chmod +x "$hook_path"
git config core.hooksPath .githooks

echo "Installed git guards."
echo "  core.hooksPath = $(git config --get core.hooksPath)"
echo "  pre-push blocks: refs/heads/main, refs/heads/master"
echo "  human override:  ALLOW_MAIN_PUSH=1 git push   (or: git push --no-verify)"
