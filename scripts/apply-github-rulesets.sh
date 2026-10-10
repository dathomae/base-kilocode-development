#!/usr/bin/env bash
#
# Apply the branch rulesets for the protected agent workflow:
#   - agent-main-protection: PR required (0 approvals), CI + authorship checks required,
#     force pushes blocked (non_fast_forward), deletions restricted. No linear-history
#     requirement (the main-sync PR uses a merge commit). No App bypass actor, and the
#     payload explicitly empties bypass_actors ("bypass_actors": []) so a re-apply clears
#     any historical App bypass; the App therefore cannot bypass non_fast_forward or the
#     PR/status-check rules — it merges its own PRs through the normal 0-approval, green
#     build/authorship gate.
#   - main-protection: PR required (1 approval), the CI check required (authorship is
#     NOT required on main, because planning commits there are human-authored), force
#     pushes blocked, deletions restricted, with the human account as a pull-requests-only
#     bypass actor so the human can merge their own agent-main -> main PR. No linear-history
#     rule, so the human's merge commit is allowed.
#
# Requires `gh` authenticated with admin on the repository.
#
# Overridable via environment:
#   REPO                default owner/repo
#   BYPASS_USER_ID      default 0 — must be set to the human merge actor's numeric user id
#   CI_CHECK            default "build"        (confirm the exact context, see below)
#   AUTHORS_CHECK       default "authorship"   (the commit-authorship check context)
#
# Confirm the check contexts after one run:
#   gh api repos/$REPO/commits/main/check-runs --jq '.check_runs[].name'
#
set -euo pipefail

REPO="${REPO:-owner/repo}"
BYPASS_USER_ID="${BYPASS_USER_ID:-0}"   # must be the human merge actor's numeric user id
CI_CHECK="${CI_CHECK:-build}"
AUTHORS_CHECK="${AUTHORS_CHECK:-authorship}"

agent_main_ruleset() {
cat <<JSON
{
  "name": "agent-main-protection",
  "target": "branch",
  "enforcement": "active",
  "conditions": { "ref_name": { "include": ["refs/heads/agent-main"], "exclude": [] } },
  "bypass_actors": [],
  "rules": [
    {
      "type": "pull_request",
      "parameters": {
        "required_approving_review_count": 0,
        "dismiss_stale_reviews_on_push": false,
        "require_code_owner_review": false,
        "require_last_push_approval": false,
        "required_review_thread_resolution": false
      }
    },
    {
      "type": "required_status_checks",
      "parameters": {
        "strict_required_status_checks_policy": true,
        "required_status_checks": [ { "context": "$CI_CHECK" }, { "context": "$AUTHORS_CHECK" } ]
      }
    },
    { "type": "non_fast_forward" },
    { "type": "deletion" }
  ]
}
JSON
}

main_ruleset() {
cat <<JSON
{
  "name": "main-protection",
  "target": "branch",
  "enforcement": "active",
  "conditions": { "ref_name": { "include": ["refs/heads/main"], "exclude": [] } },
  "bypass_actors": [
    { "actor_id": $BYPASS_USER_ID, "actor_type": "User", "bypass_mode": "pull_request" }
  ],
  "rules": [
    {
      "type": "pull_request",
      "parameters": {
        "required_approving_review_count": 1,
        "dismiss_stale_reviews_on_push": false,
        "require_code_owner_review": false,
        "require_last_push_approval": false,
        "required_review_thread_resolution": false
      }
    },
    {
      "type": "required_status_checks",
      "parameters": {
        "strict_required_status_checks_policy": true,
        "required_status_checks": [ { "context": "$CI_CHECK" } ]
      }
    },
    { "type": "non_fast_forward" },
    { "type": "deletion" }
  ]
}
JSON
}

apply() {
  local name="$1" payload="$2" id
  id="$(gh api "repos/$REPO/rulesets" --jq ".[] | select(.name==\"$name\") | .id" 2>/dev/null || true)"
  if [ -n "$id" ]; then
    printf '%s' "$payload" | gh api -X PUT "repos/$REPO/rulesets/$id" --input - >/dev/null
    echo "updated ruleset: $name (id $id)"
  else
    printf '%s' "$payload" | gh api -X POST "repos/$REPO/rulesets" --input - >/dev/null
    echo "created ruleset: $name"
  fi
}

gh auth status >/dev/null 2>&1 || {
  echo "gh is not authenticated; run 'gh auth login' with admin rights on $REPO" >&2
  exit 1
}

apply "agent-main-protection" "$(agent_main_ruleset)"
apply "main-protection" "$(main_ruleset)"

echo
echo "Done. Verify with:"
echo "  gh api repos/$REPO/rulesets --jq '.[] | {name, enforcement, target}'"
echo "  gh api repos/$REPO/commits/main/check-runs --jq '.check_runs[].name'   # expect: $CI_CHECK, $AUTHORS_CHECK"
