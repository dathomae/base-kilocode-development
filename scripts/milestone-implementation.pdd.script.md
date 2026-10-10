# Milestone Implementation

This script is executed by the `milestone-implementation-coordinator` persona. It executes an approved
milestone **one milestone at a time**: performs the **milestone-start `main`-sync** (a merge-based
main-sync PR that brings `main`'s commits into the append-only `agent-main`), reads the current milestone
definition on `agent-main`, runs the milestone's story waves, integrates their PRs into `agent-main`, runs
the **milestone-close evaluation**, and hands the `agent-main → main` merge decision to the human. After
the human merges, it proceeds to the next milestone (whose start performs the next main-sync); it never
rebases or force-pushes `agent-main`. It does not implement code and does not re-derive the schedule.

All git operations are governed by `resources/agent/git-workflow.md`; escalation by
`resources/agent/escalation-policy.md`; reports by `resources/agent/report-formats.md`; remediation by
`resources/agent/change-control.md`.

> **Revised 2026-10-04 (milestone-start sync; human-directed).** The header and Operation 5 now describe
> the milestone-start `main`-sync — a merge-based **main-sync PR** (the `integrator` opens/updates it; the
> coordinator merges it) — followed by reading the current milestone definition on `agent-main` and pausing
> to discuss a material change with the user before any wave. The former "rebase `agent-main` onto `main`
> and force-push" wording is removed; `agent-main` is append-only and is never rebased or force-pushed.

## Parameters

| Name | Type | Description | Required | Default Value |
|------|------|-------------|----------|---------------|
| `milestone_file` | string | Milestone filename under `memory-bank/planning/milestones/active-milestones/`. If omitted, the first milestone at `Stage: story implementation` in that directory's `toc.md` is used. | No | (auto) |
| `agent_main_branch` | string | The integration branch that story PRs target | No | `"agent-main"` |
| `max_parallel` | number | Maximum `story-implementation-coordinator` subagents dispatched concurrently in one wave | No | `3` |
| `checkpoint_mode` | string | When to pause for approval: `"every"` (after each wave), `"none"`, or a wave number | No | `"every"` |
| `continue_after_merge` | boolean | Proceed to the next milestone after handing the merge to the human; the only follow-up is the next milestone's wave-start main-sync PR | No | `true` |

## Canonical Paths

| Constant | Path |
|----------|------|
| `ACTIVE_MILESTONES` | `memory-bank/planning/milestones/active-milestones` |
| `FINISHED_MILESTONES` | `memory-bank/planning/milestones/finished-milestones` |
| `MILESTONE_TOC` | `memory-bank/planning/milestones/active-milestones/toc.md` |
| `FINISHED_MILESTONE_TOC` | `memory-bank/planning/milestones/finished-milestones/toc.md` |
| `ACTIVE_PLANS` | `memory-bank/planning/plans/active-plans` |
| `TRACKING_REGISTER` | `memory-bank/planning/tracking/system-definition-tracking-status.md` |

## Operations

1. **Validate parameters.** Resolve defaults; clamp `max_parallel` to ≥ 1.

2. **Resolve and read the milestone.** If `milestone_file` is omitted, read `MILESTONE_TOC` and choose the
   first milestone with `Stage: story implementation` (preferring `Status: In Progress`, else
   `Not Started`). Read `ACTIVE_MILESTONES/<milestone_file>`. Extract: `Stage`/`Status`, components, flow
   definitions, arcs and plans, story set, preconditions, wave schedule, evaluation criteria, flow
   verification results, divergences, and any open escalation or remediation entries.

3. **Human gate — start of milestone.** If `Status` is `Not Started`, present the milestone summary
   (name, arcs/plans, story count, wave count, preconditions, resolved parameters) and wait for
   explicit approval. Record it in the milestone file. If not approved, terminate.

4. **Update `Status` to `In Progress`** (keeping `Stage: story implementation`) in the milestone file and
   `MILESTONE_TOC`.

5. **Milestone-start `agent-main` sync, then read the current milestone definition.** Dispatch `integrator`
   to open/update the **main-sync PR** (base `<agent_main_branch>`, head `main`, merge commit); it is a
   no-op when `main` has not advanced. The integrator reports conflicts rather than resolving them; the
   coordinator **merges** the PR (the integrator never merges). On conflict: **stop**, record the conflict,
   mark the milestone blocked, and escalate to the human **before any story work begins**. After a clean
   sync, **read the milestone definition as it now exists on `<agent_main_branch>`**; if it differs
   materially from the definition reviewed at gate C, **pause and discuss** the change with the user before
   dispatching any wave. Never rebase or force-push `agent-main`.

6. **Select the next wave.** Choose the first wave whose status is `Not Started` and whose preconditions
   hold (prior waves `Done`; no story in it `Blocked` by an open escalation). **Do not re-derive
   conflicts** — trust the schedule. If no wave is selectable for an unexpected reason, consult
   `milestone-generation-coordinator` for revision rather than recomputing.

7. **Completion check.** If no selectable wave remains and all waves are `Done`, go to Operation 20
   (milestone close).

8. **Batch the wave** into groups of at most `max_parallel`, preserving the listed order; dispatch one
   batch at a time.

9. **Dispatch `story-implementation-coordinator` subagents** for the batch, concurrently in one message,
   each with a `work_context` per `resources/agent/work-context.md` including at minimum:
   - `work_id`: `<milestone>-Wave<wave>-<story-id>`
   - `artifact`: `story`; `level`: `story`
   - `assignment`: execute the story, produce the walkthrough, push `story/<id>`, open a PR into
     `<agent_main_branch>`, drive CI to green
   - `worktree_path`: `worktrees/<story-id>` (via the `worktrees/` symlink, never `.kilo/worktrees/`)
   - `base_branch`: `<agent_main_branch>`
   - `paths.primary`: `[memory-bank/planning/stories/active-stories/<story-file>]`
   - `expected_return`: per `resources/agent/report-formats.md` — implementation result, walkthrough
     path, PR URL, CI status, or `Blocked — needs decision`
   - `blocked_conditions`: [a requirements/architecture/design decision not already settled]

10. **Monitor results.** For each return: validate the `work_context` is unchanged, then:
    - `Completed` + green PR: record the PR URL and branch.
    - `Blocked — needs decision`: mark the story `Blocked`, stop launching new waves, hold the branch,
      and escalate.
    - `Failed` or CI never green: mark the story `Failed`; consult
      `milestone-generation-coordinator` for revision or `resources/agent/change-control.md` for
      remediation; do not merge it.

11. **Complete the wave.** Repeat 8–10 until every story in the wave is complete or marked
    blocked/failed. If any is blocked/failed, do not merge the wave; pause and follow the
    escalation/consult path.

12. **Merge the wave's PRs in the planned order.** For each story in order: ensure the branch is
    `<agent_main_branch>`, squash-merge `story/<id>` per `resources/agent/git-workflow.md`, abort and
    escalate on conflict, then push.

13. **Re-rebase remaining branches** after each merge: dispatch `integrator` to rebase each unmerged
    `story/<id>` onto the updated `<agent_main_branch>` and force-push; report conflicts. On a conflict,
    mark that story blocked, hold its branch, escalate. On CI failure after re-rebase, dispatch the
    owning `story-implementation-coordinator` to fix it with the CI log.

14. **Update the milestone file:** wave status `Done`, each story `Done`, and record divergences.

15. **Update plan files' story status** for each completed story.

16. **Update `TRACKING_REGISTER`:** per `resources/agent/end_of_task_activities.md` §Tracking status,
    remove completed stories from `Remaining stories`, advance `Implemented`, and advance `Planned`
    when the full set is complete. Update the milestone column/section.

17. **Checkpoint** per `checkpoint_mode` (`every`: summary + wait; `none`: continue; wave number: stop
    at that wave and wait, then ask whether to continue).

18. **Loop** to Operation 6 for the next wave.

19. **Consult on schedule problems.** Dispatch `milestone-generation-coordinator` with the problem and
    the current plan; apply the revised schedule. If the revision requires change control above the
    story level, follow `resources/agent/change-control.md` and obtain human authority where required.

20. **Milestone close.** When all waves are `Done` and no story is blocked or failed, confirm every
    flow's test is present (a flow's test sits where its functionality is expected to exist), then run
    `scripts/milestone-close.pdd.script.md`, which runs the acceptance flows and dispatches
    `flow-verifier` for `agent`/`manual` flows, and **finalizes the milestone on
    `<agent_main_branch>` before requesting the merge** (the pre-merge §13 record,
    `Stage: complete`/`Status: Done`, the finished-tree move, the toc/register updates, and the close
    record, committed on `<agent_main_branch>`). Record the report in the milestone's Evaluation results
    and the per-flow results in its Flow verification results. If the close evaluation finds blocking
    issues — including any flow that is not `Passed` — apply the escalation/change-control path and **do
    not** request the merge; blocking issues still block the merge request.

21. **Human merge gate.** Present the close report and the finalized milestone, and request the human's
    single `<agent_main_branch>` → `main` merge. You **do not merge or push `main`** yourself — the
    finalization is already committed on `<agent_main_branch>`, so the merge request carries it. You
    **do not wait for** or confirm the merge, and you **do not write the merged SHA back** into the
    milestone file.

22. **No post-merge action.** The human merge is the boundary after the flow.
    There is **no post-merge step on `main`** and **no second `main` merge**.
    You **do not rebase** `<agent_main_branch>` onto the new `main`; the append-only
    `<agent_main_branch>` is synced at the next milestone's wave-start **main-sync PR** (scope G). Do not
    monitor the merge PR or the post-merge CI ("no news is good news"); the human reports any merge or
    post-merge difficulty. The finished-tree move and the toc/register updates are done by the close
    script's finalization, not repeated here. The only follow-up is the next milestone's wave-start
    main-sync.

23. **Next milestone.** If `continue_after_merge` and another milestone is ready (preconditions hold),
    return to Operation 2. Otherwise present a final summary and stop.

## Output

- The milestone file being executed and the resolved parameters.
- Per wave: wave number, stories dispatched, `work_id`s, per-story status, PR URLs and CI status, merge
  SHAs, re-rebase results, and any conflicts or escalation items.
- The milestone-close report (evaluation results, including the per-flow verification results).
- The finalized pre-merge state on `<agent_main_branch>` (the §13 pre-merge record, the finished-tree
  move, the toc/register updates, and the close record) and the merge request presented to the human; no
  merged SHA is read or written back.
- Updated milestone status, plan-file story statuses, and tracking-register rows.
- Final summary, including the moved milestone file location.

## Error Handling

- Missing/invalid required parameters: prompt before execution.
- Milestone file not found or unreadable: stop and prompt for the correct path.
- `main-sync` PR conflict: stop, never auto-resolve, record it, and escalate; `<agent_main_branch>` is
  append-only and is never rebased or force-pushed.
- Story `Blocked — needs decision`: stop new waves, mark `Blocked`, hold the branch, escalate.
- Story failure or persistent CI failure: mark `Failed`, consult `milestone-generation-coordinator`;
  never silently skip.
- Merge conflict during integration: abort, mark the story blocked, hold the wave's in-flight branches,
  escalate.
- Re-rebase conflict: mark the story blocked, hold its branch, escalate.
- Unexpected schedule inconsistency: consult `milestone-generation-coordinator`; never re-derive.
- Push `main` or merge `agent-main` into `main`: prohibited; these remain human gates.
- A flow is not `Passed` at close, or its test is missing: blocking; escalate (a flow change that keeps
  the delivered functionality is in scope for `milestone-generation-coordinator`); do not request the
  merge.
- Close evaluation finds blocking issues: escalate; do not request the merge.
- Tracking update failure: report and ask whether to retry or continue.
- All pauses require explicit user direction to continue or terminate.
