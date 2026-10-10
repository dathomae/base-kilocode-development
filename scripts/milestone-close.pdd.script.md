# Milestone Close

This script is executed by the `milestone-implementation-coordinator` persona at the end of a
milestone, once all of its stories are merged into `agent-main` and no story is blocked or failed. It
runs the **milestone-close evaluation** — including the milestone's **acceptance flows** — and produces
the milestone-close report the human uses to decide whether to merge `agent-main` into `main`.

The evaluation runs the flows suite and dispatches the reviewers and critics, including the dedicated
`milestone-reviewer`, `milestone-critic`, and `flow-verifier`. Escalation is governed by
`resources/agent/escalation-policy.md`; remediation above the story level by
`resources/agent/change-control.md`; report shapes by `resources/agent/report-formats.md`; artifact
completeness by `resources/agent/artifact-standards.md` (Milestone section).

## Parameters

| Name | Type | Description | Required | Default Value |
|---|---|---|---|---|
| `milestone_file` | string | Milestone filename under `memory-bank/planning/milestones/active-milestones/` | Yes | N/A |
| `agent_main_branch` | string | The integration branch holding the milestone's merged stories | No | `agent-main` |
| `test_command` | string | Full-suite command; if omitted, use the project's standard suite | No | (project default) |
| `flows_command` | string | The flows-suite command for this milestone | No | `tests/flows/run.sh --milestone <milestone_file>` |

## Canonical Paths

| Constant | Path |
|---|---|
| `ACTIVE_MILESTONES` | `memory-bank/planning/milestones/active-milestones` |
| `MILESTONE_TOC` | `memory-bank/planning/milestones/active-milestones/toc.md` |
| `TRACKING_REGISTER` | `memory-bank/planning/tracking/system-definition-tracking-status.md` |
| `FLOWS_SUITE` | `tests/flows/` |

## Operations

1. **Validate parameters.** Confirm `milestone_file` is provided and resolves under `ACTIVE_MILESTONES`.
   Resolve defaults.

2. **Confirm the milestone is closeable.** Read the milestone and verify:
   - `Stage: story implementation` with `Status: In Progress` (approved at Gate C, implementation begun);
   - every story in its story set is `Done`;
   - every wave is `Done`;
   - no story is `Blocked` or `Failed`;
   - no open escalation or unresolved remediation remains.
   If any check fails, **stop** and escalate rather than running the evaluation.

3. **Confirm the integration state.** Verify `agent-main` contains every milestone story (all story PRs
   merged) and that its CI is green. If not, stop and report the gap.

4. **Run the full test suite.** Execute the suite (or `test_command`) on the integrated `agent-main`.
   Record the exact command and result. A failure is a blocking finding: do not request the human merge;
   escalate.

5. **Run the acceptance flows.** For every flow in the milestone's §4 Flow definitions:
   - **automated** — run it through the flows suite (`flows_command`, defaulting to
     `tests/flows/run.sh --milestone <milestone_file>`), against a **live database with no mocked
     components**;
   - **agent** / **manual** — dispatch `flow-verifier` with a `work_context` per
     `resources/agent/work-context.md` to execute the flow (agent-driven) or to record the human's
     overall report (manual; exact symptoms on failure).
   Record each result in the milestone's §12 Flow verification results (flow, status, verification mode,
   test, command, evidence, run_by, notes). A flow that is `Failed`, `Blocked`, or `Not Verified` is a
   **blocking finding**: do not request the merge. Once **every** flow is `Passed`, set the milestone
   `Stage: complete` and `Status: Done` — the terminal delivered state and the **gate to finalize
   (op 11)**. This close-evaluation transition (a) enables the stage-scoped `complete`
   `milestone-reviewer`/`milestone-critic` pass at op 7, which selects its scope by reading the
   milestone's `Stage`, and (b) is still **before** the merge request: both the close evaluation and the
   finalization complete before the merge request. The human `agent-main → main` merge is **not** part of
   `Stage`/`Status`.

6. **Evaluate the evaluation criteria.** Check the milestone's §10 Evaluation criteria — which reference
   the §4 flow-group ids plus the other gates. Record each criterion's result. An unmet criterion is a
   blocking finding.

7. **Dispatch the reviewers and critics in parallel**, each with a `work_context` per
   `resources/agent/work-context.md`:
  - **milestone-reviewer** and **milestone-critic** — at stage scope `complete` (they read the milestone's
    `Stage`), does the milestone artifact hold up (intent and standards) now that it is delivered, and do
    the flow results match the §4 flow definitions? The human `agent-main → main` merge is **not** part of
    this scope and must not be raised as a finding.
   - **design-critic** — does the delivered functionality conform to the applicable requirements,
     architecture, design, patterns, and shared vocabulary?
   - **Every distinct critic registered by the milestone's task kinds** — do the changes conform to
     the applicable standards, and do the **walkthroughs** show anomalies (unexplained deviations,
     silent scope changes, skipped verification, decisions without authority)?
   - **story-reviewer** and **plan-reviewer** — does the milestone deliver what it claims, and is it
     consistent with the plans and specs it implements?
   Wait for all results.

8. **Process findings** with one of the three dispositions per `resources/agent/escalation-policy.md`:
   - **Fix on the spot** — a small issue within the milestone's own scope that changes no upstream
     decision: fix it (via the owning story coordinator) and re-verify.
   - **Record** — a divergence that changes no outcome: record it in the milestone's Divergences.
   - **Block and escalate** — a category 1–5 issue, a spec/plan defect, an unresolved blocking finding,
     or a non-passing flow: escalate and do **not** request the merge.
   Blocking findings from either a reviewer or a critic are never waivable.

9. **Produce the milestone-close report** per `resources/agent/report-formats.md`, including: the
   milestone and revision; the test-suite result; the **per-flow verification results**; the evaluation
   criteria results; each reviewer/critic verdict and findings with severity and disposition; recorded
   divergences; and a clear **recommendation** (`ready to merge` or `blocked`).

10. **Record the report** in the milestone's §11 Evaluation results and §12 Flow verification results.

11. **Finalize on `agent-main` before requesting the merge.** With every flow `Passed` (the gate to
    finalize) and no unresolved blocking finding:
    1. **Confirm** `Stage: complete` and `Status: Done` — set in op 5; set them here if not already set.
    2. **Pre-fill the milestone's §13 pre-merge Merge record** with the `main` **baseline** SHA and the
       pre-merge `agent-main` **head** SHA (the exact head the human will merge). The `main` **baseline**
       SHA is the **merge-base of `main` and `agent-main`** (`git merge-base origin/main origin/agent-main`)
       — the base of the `main...agent-main` diff the human reviews (`human-merge-runbook.md` step 5
       reviews `git diff main...agent-main`). The merged `main` SHA is **GitHub-authoritative** and
       informational only — do **not** read it or write it back.
    3. **Move the finished trees.** Move the milestone file to `finished-milestones/`; move its plans to
       `finished-plans/` and its arc to `finished-arcs/` **only when the component has no remaining
       planned work**.
    4. **Update the tocs** and `memory-bank/planning/tracking/system-definition-tracking-status.md`.
    5. **Write the close record.**
    6. **Commit the finalized state on `agent-main`** (through the dedicated
       `chore/<milestone>-tracking` PR per `resources/agent/git-workflow.md`). State that the finalized
       state is carried by the merge request, so **no agent-side step requires a second `main` merge**.
    If the recommendation is `blocked`: present the escalation and stop. Resolve through
    `resources/agent/change-control.md` before re-running this script.

12. **Request the merge; no post-merge action / no monitoring.** Present the finalized milestone and the
    close report and request the human's **single** `agent-main → main` merge. **Never merge or push
    `main`.** The merge is the **boundary after** the flow: perform **no post-merge step on `main`**, open
    **no second `main` merge**, and do not monitor the merge PR or the post-merge CI — the human
    reports any merge or post-merge difficulty. The only follow-up is the next milestone's wave-start
    **main-sync** PR; report completion to the caller (`milestone-implementation.pdd.script.md`).

## Output

- The milestone-close report (test result, per-flow verification results, evaluation-criteria results,
  reviewer/critic verdicts, divergences, recommendation).
- Confirmation that the report was recorded in the milestone file.
- The finalized state on `agent-main` (the pre-merge §13 Merge record, the finished-tree move, the
  toc/register updates, and the close record) and the merge request presented to the human; **no merged
  SHA is read or written back**.
- Any escalation report, if the milestone is blocked.

## Error Handling

- Milestone not closeable (stories/waves incomplete, a story blocked/failed, an open escalation): stop
  and escalate; do not evaluate.
- Integration state incomplete (`agent-main` missing a story, CI red): stop and report the gap.
- Full suite fails: blocking finding; escalate; do not request the merge.
- A flow is `Failed`, `Blocked`, or `Not Verified`: blocking finding; escalate; do not request the merge.
- The flows suite cannot run (live database unavailable, harness missing): `flow-verifier` returns
  `Blocked` with what is needed; blocking finding.
- A `manual` flow has no human report: it is `Not Verified` and blocking; present the flow to the human.
- An evaluation criterion is unmet: blocking finding; escalate; do not request the merge.
- A reviewer or critic returns `terminated — insufficient context`: supply the missing context and
  re-dispatch.
- A failed subagent dispatch: retry once; if persistent, pause and ask the user.
- The human merge is the boundary after the flow: do not monitor the merge PR or the post-merge CI,
  do **not** perform a post-merge step on `main`, and do **not** wait for or confirm the merge; the human
  reports any merge or post-merge difficulty.
- All pauses require explicit user direction to continue or terminate.
