# Plan-to-Story Coordinator Script

Executed by the `plan-to-story-coordinator` persona.

Decomposes a single approved plan into its story set, ensures cross-story coherence, resolves conflicts, and prepares the set for milestone scheduling. All generation-stage review and criticism is delegated to subagents; the coordinator owns dispositions, updates tracking artifacts, and applies change-control rungs R1–R3.

---

## Parameters

| Name | Type | Description | Required | Default Value |
|---|---|---|---|---|
| `plan_file` | string | Absolute path to the approved plan markdown file under `memory-bank/planning/plans/active-plans/` | Yes | — |
| `milestone` | string | The milestone the plan serves (its `Milestone served`); defaults to the plan's front matter | No | from plan |
| `checkpoint_mode` | enum (`every` \| `end`) | When to persist the plan’s story-breakdown/status and the active-stories toc: after each story (`every`) or only after the full set is settled (`end`) | No | `every` |

---

## Operations

1. **Validate parameters.** If `plan_file` is not provided, prompt the user for it and do not proceed until it is supplied. Default `checkpoint_mode` to `every` if omitted. Verify the file exists; if not, report the missing path and terminate.

2. **Read the plan file.** Read `plan_file` in full. Capture the objective, settled decisions, scope, sequence overview, the approximate story breakdown, and its **Arc** and **Milestone served**. Read the milestone's flow definitions (its §4 definition block) so story flow coverage can be populated.

3. **Read required references.** Read every document referenced in the plan’s `Context` or `Settled decisions` sections that is needed to hand off accurate upstream decisions to story subagents (e.g., architecture contract files, design documents, requirements). Read `memory-bank/planning/stories/active-stories/toc.md`.

4. **Confirm parameters with the user once.** Present: the plan title, the number of stories in the breakdown, the `checkpoint_mode`, and the target `active-stories/` directory. Ask the user to confirm or correct before proceeding. If the user declines, terminate.

5. **For each story in the breakdown, dispatch `story-planning-coordinator`.**
   - Pass a `work_context` per `resources/agent/work-context.md` with:
     - `work_id`: `<PlanStem>-Story<N>` (e.g., `notification-plan-Story3`)
     - `artifact`: `story`
     - `level`: `story`
     - `assignment`: Draft the story file for story N per the plan breakdown and settled decisions; record the plan's **arc** and **milestone**, and — where this story is where a milestone flow's functionality is expected to exist — the flow(s) it covers (`flows:`), for `milestone-generation-coordinator` to validate.
     - `paths.primary`: `[memory-bank/planning/stories/active-stories/]`
     - `paths.read_only`: `[plan_file, <referenced specs>]`
     - `upstream_decisions`: `[plan §Settled decisions]`
     - `acceptance_criteria`: [the story-level criteria implied by the plan]
     - `expected_return`: story file path and review/critic summary
   - Dispatch in the order given by the plan’s sequence overview. Where the plan marks stories as independent (no sequencing edge between them), dispatch them in parallel in a single message; otherwise dispatch sequentially and wait for the preceding result before continuing.
   - For each returned story, record its file path and status.

6. **Checkpoint (if `checkpoint_mode` is `every`).** After each `story-planning-coordinator` returns a completed story, append the story file name and status `Not Started` to the plan’s story-breakdown/status section, and append the story entry to `memory-bank/planning/stories/active-stories/toc.md`. Do not modify the plan’s settled-decision text.

7. **Whole-set review.** Once every story is drafted, dispatch the following three subagents in parallel in a single message:
   - `story-reviewer` — judging whether the story set collectively covers the plan objective, settled decisions, and scope, and whether dependencies and acceptance criteria are consistent across stories.
   - `story-critic` — judging conformance of the full story set to the story template, sizing/parallelism rules, footprint/dependency schema, and toc format.
   - `design-critic` — judging whether any story requires an unstated design or architecture decision, and whether the set is consistent with the applicable specs.
   - For each, pass a `work_context` whose `paths.read_only` includes the plan file and every generated story file.

8. **Collect and classify whole-set findings.** For each finding, choose one of the three dispositions per `resources/agent/escalation-policy.md`:
   - **Fix on the spot** — a story-internal defect (e.g., missing `shared-infra` flag, typo in acceptance criteria) that changes no upstream decision.
   - **Record** — a divergence already covered by a later story or a non-blocking misalignment.
   - **Block and escalate** — any gap that requires changing a plan decision, spec, or milestone schedule.

9. **Resolve cross-story conflicts.**
   - Overlapping footprints: identify `shared-infra` flags, revise footprint metadata, or reorder soft dependencies.
   - Contradictory dependencies: correct `blocked-by` or soft entries so the graph is acyclic and consistent with the plan sequence.
   - Inconsistent acceptance criteria: align criteria that reference the same subsystem or contract.
   - If the fix is story-internal and changes no design or plan decision, execute **R1 autonomously**: revise the affected story in place, add a revision note, and record the change in the milestone tracking comments (or in the plan file if no milestone file exists yet).
   - If the fix requires changing the plan, escalate per `resources/agent/change-control.md` R2/R3 by invoking `scripts/change-control.pdd.script.md`.
   - After any story is revised, re-run the relevant reviewer or critic on that story alone and collect new findings.

10. **Verify footprint/dependency metadata completeness across the set.** Ensure every story has:
    - `blocked-by` (may be empty)
    - `soft` (may be empty)
    - `footprint` with specific file/package paths and correct `shared-infra` flags
     - `conflicts-with` left empty (to be computed by `milestone-generation-coordinator`)
     - its `arc` (from the plan) and its `flows:` coverage (the milestone flows this story covers; may
       be empty for infrastructure stories)
   Ensure the dependency graph is acyclic and every plan obligation is traceable to at least one story task or acceptance criterion.

11. **Apply completion gates.** Verify the story set satisfies the Story completeness criteria in `resources/agent/artifact-standards.md`: no open questions, no figure-it-out items, no missing footprint metadata. Verify **flow → story coverage**: every flow in the milestone's §4 flow definitions is covered by at least one story in the set (across all plans for the milestone; `milestone-generation-coordinator`'s coverage check is the backstop). If any gate fails, return to Operation 9.

12. **Final checkpoint (if `checkpoint_mode` is `end`).** Update the plan file’s story-breakdown/status section to list every generated story with status `Not Started`. Update `memory-bank/planning/stories/active-stories/toc.md` with the full set.

13. **Update the active-plans toc if necessary.** If the plan’s status or location has changed, update `memory-bank/planning/plans/active-plans/toc.md`.

14. **Report completion.** Return:
    - The list of generated story files
    - The plan file revision note (if any)
    - Any recorded divergences
    - Any escalations raised (with category, rung, and disposition)
    - Confirmation that the set is ready for `milestone-generation-coordinator` scheduling

---

## Output

- Generated story files under `memory-bank/planning/stories/active-stories/`
- Updated plan file with story-breakdown/status
- Updated `memory-bank/planning/stories/active-stories/toc.md`
- Coordinator completion report per `resources/agent/report-formats.md`

---

## Error Handling

- If a required parameter is missing, prompt the user and do not execute further operations until it is provided.
- If the plan file cannot be read, report the error and terminate.
- If a `story-planning-coordinator` subagent returns `Blocked — needs decision` or `Failed`, pause execution, present the result to the user, and ask whether to retry, skip, or terminate.
- If a whole-set reviewer or critic returns `terminated — insufficient context`, supply the missing context and re-dispatch.
- If an unresolvable cross-story conflict is found after the R1 attempts, escalate per `resources/agent/escalation-policy.md` and `resources/agent/change-control.md` rather than continuing silently.
