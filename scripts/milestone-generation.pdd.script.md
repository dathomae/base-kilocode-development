# Milestone Generation

This script is executed by the **milestone-generation-coordinator** persona. It defines **milestones**
(delivery increments) and their **milestone acceptance flows**, identifies participating components,
closes the requirements/architecture/design gaps that block planning, defines the **arcs**, drives **plan
generation** and **story generation**, and schedules the story waves.

It runs in **stages**, and each stage ends at a **human gate**. The coordinator MUST stop at every gate
and MUST NOT begin the next stage until the human answers it. The milestone file's `Stage:` field
(`definition` → `arc planning` → `story planning` → `story implementation` → `complete`) records where the
run stopped, so a later invocation resumes there. The same `Stage` is the review scope the
`milestone-reviewer`/`milestone-critic` apply (see the personas).

**Concepts.** A **milestone** delivers one user-level capability, verified by its **flows**, and is
evaluated and merged `agent-main → main`; milestones form a linear sequence (one predecessor, at most one
successor) and are planned in **batches**. A **milestone acceptance flow** is a named, ordered, connected
set of user-visible activities with positive and negative outcomes, organized into **flow groups** (≤5 per
milestone, ≥1 flow). An **arc** is the **plan and set of stories for one component/subsystem within one
milestone**; the arc file (`Arc_<ABBR>.md`) is durable and grows across milestones. Components are
**modified** (need an arc/plan) or **traversed** (flows pass through; no work).

## Parameters

| Name | Type | Description | Required | Default Value |
|---|---|---|---|---|
| `mode` | string | Stage to run: `definition`, `arc-planning`, `story-planning`, or `full` (run to the next gate) | No | `full` |
| `milestone` | string | The milestone to define or continue (`M<n>` or its file). Omit to start a new milestone dialog. | No | prompted |
| `batch` | string | Batch id/name when a group of milestones is planned together | No | `standalone` |
| `agent_main_branch` | string | The integration branch name | No | `agent-main` |

## Canonical Paths

| Constant | Path |
|---|---|
| `PLANNING_ROOT` | `memory-bank/planning` |
| `ACTIVE_PLANS` | `memory-bank/planning/plans/active-plans` |
| `ACTIVE_STORIES` | `memory-bank/planning/stories/active-stories` |
| `ACTIVE_ARCS` | `memory-bank/planning/arcs/active-arcs` |
| `ARC_TOC` | `memory-bank/planning/arcs/active-arcs/toc.md` |
| `ACTIVE_MILESTONES` | `memory-bank/planning/milestones/active-milestones` |
| `MILESTONE_TOC` | `memory-bank/planning/milestones/active-milestones/toc.md` |
| `TRACKING_REGISTER` | `memory-bank/planning/tracking/system-definition-tracking-status.md` |
| `GAP_REGISTER` | `memory-bank/planning/tracking/implementation-gap-tracking.md` |
| `ARC_TEMPLATE` | `resources/templates/arc.template.md` |
| `MILESTONE_TEMPLATE` | `resources/templates/milestone.template.md` |
| `PLAN_TEMPLATE` | `resources/templates/plan.template.md` |
| `TOC_FORMAT_SPEC` | `memory-bank/specifications/toc-format-spec.md` |

## Operations

The operations are performed **sequentially**. Stop at each gate.

### Stage 1 — Definition (`definition`)

1. **Validate parameters.** Confirm `mode` and `batch` are usable. If `milestone` names an existing file,
   read it; otherwise this is a new milestone dialog. On invalid input, pause and ask the user.

2. **Establish the milestone intent (human dialog).** With the human, settle:
   - the **delivered capability** — the user-level value this milestone demonstrates;
   - the **sequence position** — `Order`, `Predecessor`, `Successor` (the last milestone of a batch has an
     empty successor with a note), and `Batch`;
   - the **milestone id** `M<n>` and short name.

3. **Identify participating components.** List every component/subsystem the capability touches, marking
   each **modified** (needs an arc and plan) or **traversed** (no work). Use the register's abbreviations.

4. **Define the flow groups and flows.** Elicit, group by capability, and record:
   - flow groups (≤5), each with its capability, requirement ids, and **actors** (group level);
   - flows, each with preconditions, ordered steps (`actor`, `do`, `via: ui | http | cli`), expected
     **positive and negative** outcomes, `verification` (`automated | agent | manual`), its `test` (when
     automated), `covered_by`, and `checkable_after`.
   - For `agent`/`manual`, require non-empty **prose** `automation.notes` (what makes automation hard, the
     alternatives considered, the decision). For `manual`, note that the human is presented the detailed
     flow and reports the overall result (exact symptoms on failure).
   - If a flow cannot be detailed without a "figure it out" step, **stop** and record it as a
     specification gap (step 6). A flow is never guessed.

5. **Write the milestone file.** Create `Milestone_M<n>_<short-name>.md` from `MILESTONE_TEMPLATE` with
   `Stage: definition` and `Status: In Progress`; fill §1 Delivered functionality, §2 Milestone order,
   §3 Components, §4 Flow definitions (the YAML block), §8 Preconditions, and §10 Evaluation criteria. Add
   it to `MILESTONE_TOC` and record its batch group.

6. **Investigate gaps and identify the arcs.**
   - Read `TRACKING_REGISTER` and `GAP_REGISTER` for the participating components' open items,
     requirements/architecture/design gaps, and unscheduled gaps.
   - Dispatch `architect` (and the critics as needed) to identify/verify the required arcs and the
     requirements/architecture/design gaps that must close before arc planning and story generation. Spec
     changes are authored by `architect` under human authority.
   - Record each issue in the milestone's **Gap register** with its status (`open` until it is resolved,
     then `closed`), its resolution, and its authority. Promote resolved spec gaps to
     `TRACKING_REGISTER` §4.
   - Identify one **arc per modified component** (proposed: `Arc_<ABBR>.md`). Do **not** create the arc
     files yet — they are created in Stage 2, after Gate A approves the component set.

7. **Review.** Dispatch `milestone-reviewer` (intent) and `milestone-critic` (standards) in parallel with
   `work_context` blocks; the reviewers read the milestone's `Stage` (`definition`) to select their
   definition-scope checklist. Process findings with the three dispositions (fix / record / block and
   escalate). A missing plan, story set, wave schedule, verification result, or merge record is a
   later-stage element and is **not** a finding at this stage.

8. **GATE A — milestone gate.** Present the milestone (capability, sequence, components, flows, proposed
   arcs, gap resolutions) and state: "No plan generation will begin until you approve this milestone."
   Pause and wait. On approval, set `Stage: arc planning` and `Status: Not Started`; on change, revise and
   re-present.

### Stage 2 — Arc planning (`arc-planning`)

9. **Create the arcs and dispatch plan generation.** For each modified component, create the durable
   `Arc_<ABBR>.md` from `ARC_TEMPLATE`, adding this milestone to its milestones-served table. Then dispatch
   `plan-generation-coordinator` with a `work_context` whose inputs are the **milestone** and the **arc**;
   it produces `Plan_M<n>_<ABBR>.md` under `ACTIVE_PLANS`. There is one plan per arc.

10. **Review each plan.** Dispatch `plan-reviewer` and `plan-critic` per plan (they may run in parallel
    across arcs); process findings with the three dispositions. A plan is executable only when its
    `Open Questions` section is empty and it has no figure-it-out items.

11. **Cross-plan consistency.** Check the plans across arcs for consistent interfaces, ownership,
    sequencing, and shared assumptions; resolve or escalate per `resources/agent/escalation-policy.md`.

12. **Update the arc files.** For each arc, record the milestone, the plan link, and the design/history/
    coherence notes in `Arc_<ABBR>.md`; add the plan to the plan toc grouped by milestone.

13. **Review the arc-file updates.** Include the arc-file changes in `milestone-reviewer`/
    `milestone-critic` scope; at this stage (`arc planning`) they apply the plan-scope checklist, so the
    absence of a story set or wave schedule is **not** a finding.

14. **GATE B — plan gate.** Present **all** arc plans together with the cross-plan consistency result and
    state: "No story generation will begin until you approve these plans." Pause and wait. On approval,
    set `Stage: story planning` and `Status: Not Started`; on change, revise and re-present.

### Stage 3 — Story planning (`story-planning`)

15. **Dispatch story generation per plan.** For each plan, dispatch `plan-to-story-coordinator`; it
    produces the story files under `ACTIVE_STORIES` and populates each story's `flows:` coverage from the
    milestone's flow set.

16. **Validate flow coverage and timing.** Confirm every flow is covered by ≥1 story (flow → story
    coverage), and that a story carries a flow's test only where the flow's functionality is expected to
    exist (`checkable_after`). → Fix a local gap on the spot; a flow that cannot be covered is a
    specification gap and escalates.

17. **Compute the wave schedule.** Build a DAG from hard `blocked-by` edges; add soft preferred-ordering
    edges as weights; add footprint-overlap conflict edges. Compute waves by topological level,
    maximizing non-conflicting parallelism; document the rationale per wave.

18. **Review.** Dispatch `story-reviewer` and `story-critic` for the story set, and re-run
    `milestone-reviewer`/`milestone-critic` on the completed **planning** artifact at stage scope
    `story planning` (story set, flow coverage, and wave schedule are now in scope; verification results
    and the merge record are not); process findings.

19. **GATE C — story gate.** Present the story set, waves, and flow coverage, and state: "No
    implementation will begin until you approve this story set and the human directs work to start."
    Pause and wait. On approval, set `Stage: story implementation` and `Status: Not Started`, and report
    that `milestone-implementation-coordinator` may begin.

### Stage 4 — Escalation and change control

20. **Escalation handling.** Produce an escalation report per `resources/agent/report-formats.md`
    (category, disposition, artifacts and revisions, evidence, what is blocked, proposed rung R0–R4,
    options, recommendation). **Pause and wait for the human.**

21. **Act as change-control owner during remediation.** On a human resolution (or an autonomous R0/R1
    path): confirm the lowest sufficient rung; identify affected artifacts; drive the cascade top-down
    (R0 → `story-implementation-coordinator`; R1 → `plan-to-story-coordinator`; R2 →
    `plan-generation-coordinator` then `plan-to-story-coordinator`; R3 → human + `architect`, then R2; R4
    → human-authorized cancellation, then new plan and stories); record the remediation in the milestone's
    Remediation log and human resolutions in `TRACKING_REGISTER` §4; re-run the affected stage(s) and
    **re-gate**; then tell `milestone-implementation-coordinator` execution may continue. Never skip the
    cascade: spec before plan before story before schedule.

### Stage 5 — Revision consults during execution

22. **Handle revision consults from `milestone-implementation-coordinator`.** Read the milestone and the
    problem, then classify:
    - **In scope** (recorded as a divergence, no re-ask): reordering stories within a milestone, moving a
      story to a later milestone, adjusting a parallel group, splitting a wave, or adding/changing flows
      that keep the delivered functionality.
    - **Out of scope** (escalate, Stage 4): changing plan scope, adding/removing stories, changing a
      story's footprint or acceptance criteria, changing the delivered capability or the sequence, or any
      finding implying a plan or spec defect.
   A **substantive** change (flows, components, arcs, gaps, scope, sequencing) re-opens the downstream
   gates: set `Stage` back to the earliest affected stage (e.g. `story planning`) with `Status:
   In Progress`, re-run the affected stage(s), re-invoke `milestone-reviewer`/`milestone-critic` (which
   then apply that stage's scope), and re-gate back to `story implementation`. Back-filling a
   predecessor's successor is substantive.

## Output

- The milestone file path, its `Stage`/`Status`, and the gate status (`awaiting gate A/B/C`, `approved`,
  `rejected`).
- The participating components (modified/traversed) and the flow groups and flows defined.
- The Gap register entries and their resolutions, and the arcs defined.
- The per-arc plan paths and the cross-plan consistency result.
- The story set, flow coverage, and computed wave schedule.
- The `milestone-reviewer` / `milestone-critic` / plan / story review summaries, or the escalation report.
- Any recorded divergences and remediation-log entries.

## Error Handling

- Missing or invalid parameters pause execution and prompt the user.
- A flow that cannot be detailed without a "figure it out" step stops Stage 1 and is recorded as a
  specification gap.
- More than five flow groups stops Stage 1; the resolution is to split the milestone (human decision).
- A non-automated flow without `automation.notes` is a blocking milestone-critic finding.
- A referenced artifact that belongs to a later stage is expected and does not pause execution; a
  referenced artifact **required by the current stage** that is missing pauses execution and the user
  chooses to correct or terminate.
- A plan or story with a non-empty `Open Questions` section, or a figure-it-out item, stops the stage.
- A cycle in hard `blocked-by` dependencies stops Stage 3 with an escalation report (category 3).
- A flow not covered by any story is a specification gap and escalates (category 3).
- A failed subagent dispatch retries once; if persistent, pause and ask the user.
- A `terminated — insufficient context` review result is given the missing context and re-dispatched.
- A human rejection at a gate is revised and re-presented; if it implies a plan or spec defect, enter
  Stage 4.
- All pauses require explicit user direction to continue or terminate.

## Notes

- **Coordinator conventions** (dispatch, work context, parallelism, dispositions, remediation, completion
  gates, reports) follow the "Coordinator scripts" section of
  `resources/templates/prompt-driven-development-script.template.md`.
- **Gate discipline.** Do not begin a stage until the human answers the preceding gate, including on
  re-entry after a substantive change.
- **Subagents do not "create a new task back."** The coordinator receives each result as a normal tool
  result and decides the next step.
- **Permission boundary.** The coordinator writes only the arc files, the milestone files, their tocs, and
  the milestone column of the tracking register.
- **Context management.** The coordinator's context is reserved for milestone-level coordination; it does
  not read full plans or stories for detail.
