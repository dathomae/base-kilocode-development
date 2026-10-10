# Change Control Remediation

This is the shared remediation procedure executed by the coordinator that owns the affected artifact set:
`milestone-generation-coordinator` during an active milestone; otherwise `plan-generation-coordinator` (for R2/R3)
or `plan-to-story-coordinator` (for R1). It resolves escalated issues that require changes above the
story level per `resources/agent/change-control.md`.

## Parameters

The coordinator executing this script supplies the following parameters.

All required parameters MUST be present before execution begins.

| Name | Type | Description | Required | Default Value |
|---|---|---|---|---|
| `issue` | string | The escalated issue and its evidence, including the originating escalation report. | Yes | — |
| `category` | integer | The escalation category, 1–5 per `resources/agent/escalation-policy.md`. | Yes | — |
| `artifacts` | list of strings | The artifact(s) and revision(s) affected by the issue. | Yes | — |
| `proposed_rung` | string | The proposed remediation rung (R0–R4). If omitted, the coordinator determines the rung. | No | — |

## Operations

Execute the following operations in strict sequence. Do not skip or reorder.

1. **Validate parameters.**
   - Ensure `issue`, `category`, and `artifacts` are present and non-empty.
   - Ensure `category` is in the range 1–5 per `resources/agent/escalation-policy.md`.
   - If `proposed_rung` is provided, ensure it is one of `R0`, `R1`, `R2`, `R3`, or `R4`.
   - If validation fails, report the specific deficiency and stop. Do not proceed.

2. **Load context.**
   - Read the originating escalation report, the affected artifact(s), and the active milestone file if one exists (`memory-bank/planning/milestones/active-milestones/`).
   - Read `resources/agent/change-control.md` and `resources/agent/escalation-policy.md`.

3. **Confirm the rung and identify the affected artifact set.**
   - Using the remediation ladder in `resources/agent/change-control.md`, determine the lowest sufficient rung given the `issue`, `category`, and `artifacts`. If `proposed_rung` was provided, evaluate whether it is sufficient or a lower rung applies.
   - If the confirmed rung is **R0**, route the issue to `story-implementation-coordinator` per `resources/agent/change-control.md`, record the routing, and terminate this script.
   - Identify the exact artifact set touched by the confirmed rung (spec, plan, stories, schedule). Record the confirmed rung and the artifact set.

4. **Determine ownership.**
   - Check whether an active milestone exists in `memory-bank/planning/arcs/active-arcs/`.
   - If an active milestone exists, the executing coordinator is the change-control owner (normally `milestone-generation-coordinator`).
   - If no active milestone exists, the owner is the invoking generation coordinator: `plan-generation-coordinator` for R2/R3, or `plan-to-story-coordinator` for R1, per `resources/agent/change-control.md`.

5. **Obtain authority.**
   - Consult `resources/agent/change-control.md` for the authority rules associated with the confirmed rung.
   - For **R0/R1**: autonomous execution is permitted unless a category triggered the escalation; if the escalation category requires it, or if cancellation is involved, obtain human sign-off.
   - For **R2–R4**: obtain human sign-off. Present the confirmed rung, issue, artifacts, and proposed cascade to the human and await approval.
   - For **R3**: additionally engage `architect`. The human and `architect` are the decision authority for spec revisions.
   - For **all cancellations** (R4 or any rung requiring cancellation): human authorization is mandatory per `resources/agent/change-control.md`.
   - Record the authority obtained (autonomous, human sign-off, human + `architect`).

6. **Pause affected work.**
   - If an active milestone exists, mark affected stories as blocked in the milestone file.
   - In-flight story branches are held, not discarded.
   - Work independent of the issue may continue unless the coordinator judges the issue contaminates it.
   - Record the pause state.

7. **Execute the cascade top-down.**
   Proceed only after authority is obtained and affected work is paused. Execute each phase in order; do not begin a downstream phase until the upstream phase is settled.
   - **a. Spec (R3 only):** The human and `architect` revise the requirements/architecture/design document. Record the revision in the spec. Dispatch `architect` via the Task tool with a `work_context` per `resources/agent/work-context.md` if the coordinator is not performing the revision directly. The coordinator receives the result as a normal tool result.
   - **b. Plan (R2/R3):** Dispatch `plan-generation-coordinator` via the Task tool with a `work_context` per `resources/agent/work-context.md` to revise the plan in place. Mark the plan revised. Wait for the plan revision to be settled before proceeding.
   - **c. Plan (R4):** Dispatch `plan-generation-coordinator` via the Task tool with a `work_context` per `resources/agent/work-context.md` to author a new plan. Wait for the new plan to be settled before proceeding.
   - **d. Stories (R1–R4):** Dispatch `plan-to-story-coordinator` via the Task tool with a `work_context` per `resources/agent/work-context.md` to revise, cancel, or regenerate the affected stories through `story-planning-coordinator`. For R1, this is an autonomous story revision. For R2/R3, this follows the revised plan. For R4, this follows the new plan. Wait for the story set to be settled before proceeding.

8. **Apply revision-vs-replacement rules.**
   - For **R1–R3**: append a revision note to each revised plan and story file. The note records what changed, why, and the triggering issue per `resources/agent/change-control.md`.
   - For **R4**: move the cancelled plan to `memory-bank/planning/plans/canceled-plans/` and cancelled stories to `memory-bank/planning/stories/canceled-stories/`. Update the relevant `toc.md` files (`active-plans/toc.md`, `active-stories/toc.md`, `canceled-plans/toc.md`, `canceled-stories/toc.md`) per `.kilo/rules/organization.md`. Abandon in-flight branches for cancelled stories; never merge them.
   - For **revised stories**: decide whether to continue the existing branch or restart it from the revised story. Record this decision in the story's revision note or the Remediation log.

9. **Record the remediation.**
   - Add a Remediation log entry to the active milestone file (or to the plan's tracking section if no active milestone exists) per the Remediation log entry shape in `resources/agent/report-formats.md`.
   - Add a §4 Resolution log entry to `memory-bank/planning/tracking/system-definition-tracking-status.md` per `resources/agent/report-formats.md`.
   - Update the originating artifact (plan, story, or spec) with a reference to the resolution.

10. **Reschedule and resume.**
    - If an active milestone exists, the change-control owner recomputes the waves and updates the milestone file. Coordinate with `milestone-implementation-coordinator` to resume execution of the revised schedule.
    - If no active milestone exists, update any relevant plan or story tracking and declare the remediation complete.

11. **Report outcome.**
    Produce a final report per `resources/agent/report-formats.md` summarizing:
    - the confirmed rung,
    - the authority obtained,
    - the artifacts touched (revised / cancelled / created),
    - the resulting schedule change,
    - references to the Remediation log and §4 Resolution log entries,
    - any open items or deferred decisions.

## Output

- A change-control outcome report (per `resources/agent/report-formats.md`) containing the confirmed rung, authority, artifacts touched, schedule change, and log references.
- Updated artifact files with revision notes, moved cancellation artifacts, and updated `toc.md` files.
- Updated milestone file Remediation log and `system-definition-tracking-status.md` §4 Resolution log.

## Error Handling

- **Missing required parameters:** If `issue`, `category`, or `artifacts` are missing or invalid, stop execution, report the specific deficiency, and prompt for the missing values. Operations WILL NOT execute until all required parameters are specified.
- **Rung mismatch:** If the confirmed rung differs from `proposed_rung`, record the rationale for selecting the lower or different rung and continue.
- **Authority denied:** If human sign-off is denied or `architect` is unavailable, stop execution, record the blockage, and escalate to the human per `resources/agent/escalation-policy.md`.
- **Cascade failure:** If a dispatched coordinator (`plan-generation-coordinator`, `plan-to-story-coordinator`, or `story-planning-coordinator`) returns `Blocked — needs decision` or `Failed`, pause execution. Ask the human whether to retry, revise the rung, or terminate. Do not silently skip a failed cascade step.
- **General execution error:** If an unexpected error occurs during any operation, pause execution and ask the human whether to continue or terminate.
