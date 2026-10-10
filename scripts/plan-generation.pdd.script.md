# Plan Generation

This script formulates a new plan or revises an existing plan **for one arc within one milestone**. It is executed by the `plan-generation-coordinator` persona, dispatched by `milestone-generation-coordinator` with the milestone and arc as inputs.

The coordinator owns a formulate → review → criticize loop for the plan artifact. It delegates drafting to `plan-author`, intent review to `plan-reviewer`, standards review to `plan-critic`, and design-pattern review to `design-critic`. It does not implement code or author specs.

## Parameters

| Name | Type | Description | Required | Default Value |
|------|------|-------------|----------|---------------|
| milestone | string | The milestone file (`Milestone_M<n>_<short-name>.md`) the plan serves | Yes | N/A |
| arc | string | The arc (`Arc_<ABBR>.md`) the plan covers | Yes | N/A |
| plan_file | string | Target plan filename; defaults to `Plan_M<n>_<ABBR>.md` | No | derived from milestone + arc |
| checkpoint_mode | string | When to pause for user approval: `"every"` or `"none"` | No | `"every"` |

## Operations

The following operations are performed by the **plan-generation-coordinator** unless otherwise specified.

1. **Validate parameters.** Ensure `milestone` and `arc` are provided and non-empty; derive `plan_file` as `Plan_<milestone id>_<ABBR>.md` when omitted. If either is missing, prompt the user and do not proceed.

2. **Determine plan disposition.** Check whether `memory-bank/planning/plans/active-plans/<plan_file>` already exists, and whether a copy exists in `finished-plans/` or `canceled-plans/`. If an active or finished copy exists, this is a **revision**; note the prior version for reference. If none exists, this is a **new plan**.

3. **Read mandatory context.** Read `memory-bank/lessons-learned.md`, `memory-bank/concepts.md`, and `memory-bank/terms.md` per `.kilo/rules/system-rules.md`.

4. **Read authoritative inputs.** Read the milestone and arc the plan serves, and the settled requirements/architecture/design the plan must honor:
   - the **milestone** file — delivered capability, components, flow definitions, gap register, and the evaluation criteria the arc must support;
   - the **arc** file — the component/subsystem scope and its design/history/coherence notes;
   - `memory-bank/requirements.md`
   - `memory-bank/design.md`
   - Relevant architecture specs and subsystem contracts (follow the `toc.md` chain from the project root to the affected subsystem).
   - The existing plan content if revising.

5. **Read templates and standards.** Load `resources/templates/plan.template.md` and the Plan section of `resources/agent/artifact-standards.md` to internalize the required shape and the completion gates.

6. **Confirm parameters with the user once.** Present:
   - Milestone served and arc
   - Objective (derived from the milestone's capability and the arc's scope)
   - Plan filename and disposition (new / revision)
   - Checkpoint mode
   - Key upstream documents identified in Operation 4
   Wait for explicit confirmation or modification. Record the confirmed values.

7. **Dispatch `plan-author` to draft the plan.** Use the Task tool. Pass a `work_context` block per `resources/agent/work-context.md`:
   ```yaml
   work_context:
     work_id: <plan_file>-Draft
     artifact: plan
     level: plan
     assignment: Draft a plan per resources/templates/plan.template.md for the arc, inside the milestone's settled decisions and within settled architecture/design.
     paths:
       primary: [memory-bank/planning/plans/active-plans/<plan_file>]
       read_only: [<milestone and arc files>, <upstream specs from Operation 4>]
     standards: [resources/templates/plan.template.md, resources/agent/artifact-standards.md]
     upstream_decisions: [<the milestone's settled decisions by reference>, <settled decisions by reference>]
     acceptance_criteria:
       - Plan follows the plan template and records the Arc and Milestone served.
       - All settled decisions are explicit.
       - Open Questions is empty.
       - No figure-it-out items in the story breakdown.
     expected_return: The drafted plan content plus any issues that should be escalated.
     blocked_conditions: [the objective requires a requirements/architecture/design decision that is not yet settled]
   ```
   The coordinator MUST NOT draft the plan itself.

8. **Receive the draft.** Validate that the subagent returned the `work_context` unchanged. If `plan-author` returned `Blocked — needs decision`, stop and treat it as an escalation per `resources/agent/escalation-policy.md` (categories 1 or 5). Do not proceed to review until the blocker is resolved.

9. **Dispatch reviewers and critics concurrently.** The three subagents are independent; launch them simultaneously in a single message via the Task tool. Each receives the draft plan and the confirmed parameters.

   a. **`plan-reviewer`** — `work_context`:
      ```yaml
      work_context:
        work_id: <plan_file>-Review
        artifact: plan
        level: plan
        assignment: Review the draft plan for conformance to intent: does it correctly and completely implement the requirements and settled architecture/design decisions it must honor?
        paths:
          primary: [memory-bank/planning/plans/active-plans/<plan_file>]
          read_only: [<upstream specs from Operation 4>]
        standards: []
        upstream_decisions: [<settled decisions by reference>]
        acceptance_criteria: [the plan addresses every requirement/decision in scope]
        expected_return: Review report per resources/agent/report-formats.md
        blocked_conditions: []
      ```

   b. **`plan-critic`** — `work_context`:
      ```yaml
      work_context:
        work_id: <plan_file>-Critic
        artifact: plan
        level: plan
        assignment: Critique the draft plan for conformance to the plan template, planning organization, and artifact standards.
        paths:
          primary: [memory-bank/planning/plans/active-plans/<plan_file>]
          read_only: [resources/templates/plan.template.md, resources/agent/artifact-standards.md, .kilo/rules/organization.md]
        standards: [resources/templates/plan.template.md, resources/agent/artifact-standards.md, .kilo/rules/organization.md]
        upstream_decisions: []
        acceptance_criteria: [all required elements present, Open Questions empty, no figure-it-out items]
        expected_return: Critic report per resources/agent/report-formats.md
        blocked_conditions: []
      ```

   c. **`design-critic`** — `work_context`:
      ```yaml
      work_context:
        work_id: <plan_file>-DesignCritic
        artifact: plan
        level: plan
        assignment: Critique the draft plan for pattern/principle conformance and verify it does not require a spec change.
        paths:
          primary: [memory-bank/planning/plans/active-plans/<plan_file>]
          read_only: [<relevant architecture/design specs from Operation 4>]
        standards: [<relevant architecture/design standards per resources/agent/standards-manifests.md>]
        upstream_decisions: [<settled decisions by reference>]
        acceptance_criteria: [plan conforms to architecture patterns; no decision changes requirements/architecture/design]
        expected_return: Critic report per resources/agent/report-formats.md
        blocked_conditions: []
      ```

10. **Collect reports.** Validate returned `work_context` for all three subagents. If any subagent reports `terminated — insufficient context`, stop and re-dispatch with the missing context.

11. **Classify findings.** For every finding across the three reports, apply one of the three dispositions defined in `resources/agent/escalation-policy.md`:
    - **Fix on the spot** — local defect at the plan level that changes no upstream decision.
    - **Record** — divergence that changes no outcome, or already covered by a future story.
    - **Block and escalate** — any of the five escalation categories (categories 1–5 in `resources/agent/escalation-policy.md`).

    The coordinator owns the disposition; reviewers and critics propose one.

12. **Handle escalations.** If any finding is category 1 or 3 (a decision that would change requirements, architecture, or design), produce an escalation report per `resources/agent/report-formats.md`. Include category, artifact, evidence, what is blocked, proposed rung (R2 or R3 per `resources/agent/change-control.md`), options, and recommendation. Stop and present the report to the user for resolution. Do not write the plan file while an escalation is open.

13. **Apply fix-on-the-spot findings.** Dispatch `plan-author` with revision instructions addressing all non-escalated blocking findings. Iterate up to 3 draft–review cycles for this tier. If blocking findings persist after 3 cycles, escalate to the human per `resources/agent/escalation-policy.md` (retry limits).

14. **Enforce completion gates.** Before declaring the plan complete, verify the gates in `resources/agent/artifact-standards.md` (Plan section) and `resources/templates/plan.template.md`:
    - The `Open Questions` section is **empty**.
    - No story in the breakdown is a figure-it-out item.
    - Every settled decision is recorded explicitly.
    - Scope in/out is stated.
   - The story breakdown is present.
   - The **Arc** and **Milestone served** are recorded and match the dispatch inputs; the plan proposes
     no milestones.
   - Deferred items are references to requirements/architecture sections, not open questions.

    If any gate fails, dispatch `plan-author` to close the gap and return to Operation 11.

15. **Write the plan.** Save the finalized plan content to `memory-bank/planning/plans/active-plans/<plan_file>`.

16. **Update the active-plans toc.** Add or update the entry in `memory-bank/planning/plans/active-plans/toc.md` per `memory-bank/specifications/toc-format-spec.md`, grouped by the milestone it serves. For a new plan, set status `Not Started`. For a revision, update the status and append a revision note.

17. **Handle R2 change control for revisions.** If this is a revision of an existing plan:
    - Confirm the rung is **R2** per `resources/agent/change-control.md`.
    - Obtain human sign-off.
    - Append a revision note to the plan file recording what changed, why, and the triggering issue.
    - Dispatch `plan-to-story-coordinator` (via Task tool) to revise the affected story set, passing the revised plan and a `work_context` that identifies the change trigger.
    - Record the remediation in the milestone file if one exists.

18. **Checkpoint.** Present the final plan path and summary to the user.
    - If `checkpoint_mode` is `"every"`, wait for explicit approval before considering the script complete.
    - If `"none"`, present a brief confirmation and complete.

## Output

- The milestone served and the arc covered.
- The path to the written plan file (`memory-bank/planning/plans/active-plans/<plan_file>`).
- Confirmation of new vs. revision disposition.
- The `plan-author` draft summary.
- Review and critic verdicts (pass/fail) and the final disposition of any findings.
- Escalation report, if generated.
- R2 change-control confirmation, if applicable.
- The updated `active-plans/toc.md` entry.

## Error Handling

If required parameters are not specified, the user is prompted. Operations WILL NOT execute until all required parameters are specified.

During execution, if an error is encountered:
- If a subagent fails or returns `Failed`, assess whether to retry (within the 3-cycle limit), record, or escalate.
- If a subagent returns `Blocked — needs decision`, stop and escalate per `resources/agent/escalation-policy.md`.
- If file write operations fail, pause and ask the user whether to retry or terminate.
- If the plan would require changing settled requirements/architecture/design, stop and escalate rather than altering the plan boundary.
