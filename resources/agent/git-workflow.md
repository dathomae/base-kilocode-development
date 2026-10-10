# Git Workflow

This is the branch and integration contract used by the milestone and story coordinators. It allows
agents to execute multi-story milestones while keeping `main` under human control.

Related: [`agent-setup.md`](agent-setup.md), [`change-control.md`](change-control.md),
[`escalation-policy.md`](escalation-policy.md), and [`report-formats.md`](report-formats.md).

## Branch model

| Branch | Owner | Purpose |
|---|---|---|
| `main` | Human | Trunk. Agents never push or merge it. |
| `agent-main` | Milestone implementation coordinator | Append-only agent integration branch. |
| `story/<id>` | Story implementation coordinator | One isolated branch and worktree per story. |
| `chore/<milestone>-tracking` | Milestone implementation coordinator | Tracking and milestone-finalization changes. |

## Invariants

- Agents authenticate as the configured GitHub App for agent-owned commits and pull requests.
- Agents never push or merge `main`.
- `agent-main` is append-only: no direct pushes and no force pushes.
- Story and tracking changes reach `agent-main` through pull requests.
- A story has one PR into `agent-main` and is squash-merged only after its required checks pass.
- Force pushes are allowed only for `story/*` after rebasing.
- The `integrator` performs instructed Git and CI operations but never makes semantic decisions and
  never merges a pull request.
- The milestone implementation coordinator merges agent-owned PRs into `agent-main` in scheduled
  order. The human alone performs the final `agent-main` to `main` merge.
- Never delete an unmerged branch. After a story is merged, remove its remote branch, local worktree,
  and local branch.

## Milestone start and wave start

Before the first wave, and before each later wave, synchronize `main` into `agent-main` through a
merge-based main-sync PR. The integrator opens or updates the PR; the milestone implementation
coordinator merges it. Never rebase or force-push `agent-main`.

After the milestone-start sync, re-read the milestone definition from `agent-main`. If it differs
materially from the version approved by the human, pause before dispatching work. Any sync conflict
stops the wave and is escalated to the human.

## Per-story flow

1. Cut `story/<id>` from the current `agent-main` and create its worktree under the `worktrees/`
   symlink.
2. Execute the story task loop: implement, review for intent, and criticize for standards.
3. Complete the story artifact, tracking updates, and walkthrough.
4. Run the project's recorded pre-PR verification command. The consuming project must define this
   command in its development documentation or milestone; never guess it.
5. Rebase the story branch onto current `agent-main`.
6. Have the integrator push the branch and open or update its PR into `agent-main`.
7. Drive required CI checks to green with bounded repair cycles.
8. The milestone implementation coordinator squash-merges in scheduled order, then instructs the
   integrator to remove the merged branch and worktree.

When one story merge advances `agent-main`, remaining story branches must be rebased again by their
owning story coordinators before merge.

## Tracking changes

Coordinator-owned changes—milestone state, plan state, story moves, tocs, walkthroughs, and tracking
registers—travel through `chore/<milestone>-tracking` into `agent-main`. The integrator opens or updates
the PR; the milestone implementation coordinator squash-merges it.

## Milestone close

1. Run the project's full verification suite and every milestone acceptance flow.
2. Run the milestone reviewer and applicable critics.
3. Resolve all blocking findings or escalate them.
4. Finalize the milestone before requesting the human merge: mark it complete, move completed
   artifacts, update tocs and tracking registers, and land those changes through the tracking PR.
5. Present the close report and ask the human to merge `agent-main` into `main` with a merge commit.
6. Perform no post-merge agent write to `main`. The next milestone begins with another main-sync PR.

## CI contract

The consuming project owns its build and verification workflow. Record one local pre-PR entry point
that runs the same checks as CI, then reference that command from project documentation and milestone
artifacts. Story coordinators run it before pushing and use CI logs to repair failures.

The [authorship workflow](../../.github/workflows/authorship-check.yml) checks that commits introduced
on agent-owned branches use the configured App bot email while exempting commits already reachable
from `main`. It is required when the protected milestone workflow is enabled because the
`agent-main` ruleset requires its `authorship` check.

## Rulesets and local guard

`scripts/apply-github-rulesets.sh` configures the expected protected-branch model. Review its payloads
against the consuming repository before applying them. `scripts/install-git-guards.sh` installs the
tracked pre-push hook as defense in depth; a human can deliberately override the hook with
`ALLOW_MAIN_PUSH=1` or `--no-verify`.

## Escalation and holds

On escalation, stop launching affected waves. Preserve in-flight branches, record the blocking issue,
and wait for the authorized resolution. Canceled work is abandoned through change control; it is not
silently deleted or merged.
