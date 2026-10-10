# Story Planning

This script formulates **one story** from a settled plan (or a section of one). It reads the plan and its referenced specs, confirms parameters with the user, drafts the story via `plan-author`, reviews it via `story-reviewer`, critiques it via `story-critic` and `design-critic`, resolves findings internal to the story, escalates findings that would change an upstream artifact, and — once the completion gates pass — writes the story file to `memory-bank/planning/stories/active-stories/` and adds it to that directory's `toc.md`.

This script is executed by the **`story-planning-coordinator`** persona. It is invoked per story by `plan-to-story-coordinator`, or directly by a human.

## Parameters

| Name | Type | Description | Required | Default Value |
|---|---|---|---|---|
| `plan_file` | string | Path to the settled plan file, relative to the project root (e.g., `memory-bank/planning/plans/active-plans/notification-delivery-plan.md`) | Yes | N/A |
| `story_file` | string | Target story filename (e.g., `Story040_api-rate-limiting.md`). Must match `^Story\d{3}_[a-zA-Z0-9_-]+\.md$`. | Yes | N/A |
| `checkpoint_mode` | string | When to pause for explicit user confirmation: `every` (after each operation), `milestone` (before writing and at final report only), `none` (errors and escalations only) | No | `every` |

## Canonical Paths

| Constant | Path |
|---|---|
| `ACTIVE_STORIES` | `memory-bank/planning/stories/active-stories` |
| `ACTIVE_TOC` | `memory-bank/planning/stories/active-stories/toc.md` |
| `STORIES_ROOT` | `memory-bank/planning/stories` |
| `FINISHED_STORIES` | `memory-bank/planning/stories/finished-stories` |
| `CANCELED_STORIES` | `memory-bank/planning/stories/canceled-stories` |
| `STORY_TEMPLATE` | `resources/templates/stories.md` |

## Operations

1. **Validate parameter constraints.**
   - `plan_file` must be provided, non-empty, and exist at the given path.
   - `story_file` must be provided, non-empty, and match `^Story\d{3}_[a-zA-Z0-9_-]+\.md$`.
   - `checkpoint_mode` must be one of `every`, `milestone`, `none`.
   - Verify the `ACTIVE_STORIES` directory exists. If it does not, pause and ask the user whether to terminate.
   - Verify no file already exists at `ACTIVE_STORIES/<story_file>`. If a collision exists, pause and ask the user whether to overwrite or terminate.

2. **Read the plan file.**
   - Read `plan_file`.
   - Extract its objective, settled decisions, scope in/out, story breakdown, and references to upstream specs.
   - If the plan's `Open Questions` section is non-empty, stop and escalate (category 2) to the invoker per `resources/agent/escalation-policy.md`; do not author a story from an incomplete plan.

3. **Read referenced specs.**
   - For each spec (requirements, architecture, design) referenced in the plan, read it.
   - If a referenced spec is missing, record a warning but do not halt unless the missing spec is required to settle a decision the story must implement.

4. **Confirm parameters with the user.**
   - Display `plan_file`, `story_file`, and `checkpoint_mode`.
   - Display the plan objective and the story's intended position in the plan breakdown.
   - Ask the user to confirm or adjust before proceeding.
   - If the user declines, terminate without writing any file.
   - If `checkpoint_mode` is `every` or `milestone`, this is a mandatory checkpoint.

5. **Dispatch `plan-author` to draft the story.**
   - Pass a `work_context` block per `resources/agent/work-context.md`:
     - `work_id`: `<story_file>-Draft`
     - `artifact`: `story`
     - `level`: `story`
     - `assignment`: Draft a complete story from the settled plan, conforming to `resources/templates/stories.md` and the Story section of `resources/agent/artifact-standards.md`.
     - `paths.primary`: `[ACTIVE_STORIES/<story_file>]` (intended target)
     - `paths.read_only`: `[plan_file, <referenced specs>, STORY_TEMPLATE]`
     - `standards`: `[STORY_TEMPLATE, resources/agent/artifact-standards.md]`
     - `upstream_decisions`: `[plan §Settled decisions, plan §Scope, referenced spec sections]`
     - `acceptance_criteria`: `[Complete story draft with Goal, References, Dependencies, Dependent Stories, structured footprint/dependency metadata, Tasks with [Small] sizing and registered [Kind: <id>] routing where applicable, complete specialization contracts, Task Sizing, Parallel Execution, Execution Order, Reintegration, Constraints, Intent, Acceptance Criteria, and Logical TDD notes]`
     - `expected_return`: Story draft content, list of files it expects to touch, and any escalation triggers (categories 1–5)
     - `blocked_conditions`: [A missing upstream decision, an open question in the plan, a requirement to make a design or architecture decision]
   - Instruct `plan-author` to produce a draft that includes:
     - **Goal**: what the story accomplishes, with context from the plan.
     - **References**: links to the plan and upstream specs.
     - **Dependencies** (`blocked-by`): hard prerequisites by story id.
     - **Dependent Stories**: stories blocked by this one.
     - **Structured footprint/dependency metadata**: hard `blocked-by`, soft preferred-ordering dependencies, expected `footprint` (files/packages with `shared-infra` flags), and `conflicts-with` (leave empty; computed by `milestone-generation-coordinator`).
     - **Tasks**: numbered tasks with statuses, `[Small]` sizing annotations where appropriate, and
       `[Kind: <id>]` for specialization-defined work. Load `resources/agent/task-kinds.json`; include
       every field required by a selected kind's `contract_schema`.
     - **Task Sizing subsection**: which tasks are `[Small]` and why.
     - **Parallel Execution subsection**: genuinely file-disjoint groups with rationale.
     - **Execution Order subsection**: dependency graph showing sequential chains and parallel branches.
     - **Reintegration subsection**: per parallel group — merge order, integration test commands, conflict-prone files.
     - **Constraints, Intent, Acceptance Criteria**.
     - **Logical TDD notes** for coding tasks (tests integrated within each task, not broken out separately).
   - `plan-author` must NOT write the file; it returns the draft content.

6. **Receive the draft from `plan-author`.**
   - If `plan-author` returns `Blocked — needs decision` or reports escalation triggers (categories 1–5), collect them and proceed to step 10 (escalation). Do not write the story file or update `toc.md`.
   - If the draft is missing the structured footprint/dependency metadata, treat it as an internal defect and return to step 5 for revision.

7. **In parallel, dispatch `story-reviewer`, `story-critic`, and `design-critic`.**
   - Pass each a `work_context` block per `resources/agent/work-context.md`:
     - `work_id`: `<story_file>-<role>` (e.g., `Story040-Reviewer`)
     - `artifact`: `story`
     - `level`: `story`
     - `paths.primary`: `[draft story content]`
     - `paths.read_only`: `[plan_file, <referenced specs>]`
     - `upstream_decisions`: same as step 5
     - `expected_return`: Report per `resources/agent/report-formats.md`
   - **`story-reviewer`**: Judge conformance to intent — does the story correctly and completely implement its plan section, with verifiable acceptance criteria and a task breakdown that reaches them? (See `resources/agent/report-formats.md`, Review report.)
   - **`story-critic`**: Judge conformance to standards — story template, task sizing, `[Small]` and
     registered `[Kind: <id>]` routing, specialization contracts, parallel-execution and reintegration
     rules, footprint/dependency schema, empty `Open Questions`, no figure-it-out items. (See
     `resources/agent/report-formats.md`, Critic report, and `resources/agent/standards-manifests.md`.)
   - **`design-critic`**: Judge whether the story avoids requiring a design or architecture decision, and whether it is consistent with applicable specs. (See `resources/agent/report-formats.md`, Critic report.) A task that asks the implementor to "figure out", "design", "choose", or "evaluate" an approach is a category 1 or 5 escalation trigger.

8. **Collect findings and classify dispositions.**
   - For each finding from the three reports, assign one of the three dispositions per `resources/agent/escalation-policy.md`:
     - **Fix on the spot**: a local defect at the story's own level that changes no upstream decision.
     - **Record**: a divergence that changes no outcome and needs no decision, or an issue already covered by a future story.
     - **Block and escalate**: any of the five escalation categories (categories 1–5 per `resources/agent/escalation-policy.md`).
   - Blocking findings from a reviewer or critic are mandatory and never waivable.

9. **Resolve internal findings (review-fix loop).**
   - For findings classified as **fix on the spot**, dispatch `plan-author` to revise the draft.
   - Re-dispatch the affected reviewer/critic(s) after each revision.
   - Limit: up to **3** review-fix iterations per `resources/agent/escalation-policy.md`. If blocking findings persist after the third cycle, classify the remaining findings as escalate and proceed to step 10.
   - If a finding concerns a missing design decision (category 1 or 5) detected by `design-critic`, do not attempt to fix it by revising the story — escalate immediately.

10. **Escalate external findings.**
    - For findings classified as **block and escalate**, build an escalation report per `resources/agent/report-formats.md` (Escalation report).
    - Include: category (1–5), disposition, artifact and revision, evidence, what work is blocked, proposed rung (R0–R4 per `resources/agent/change-control.md`), options, and the coordinator's recommendation.
    - Route to `plan-to-story-coordinator` when this script was invoked as its subagent; otherwise route to the human.
    - Do NOT write the story file (step 12) or update `toc.md` (step 13) while escalations are unresolved.
    - If `checkpoint_mode` is `every` or `milestone`, present the escalation report to the user before proceeding.

11. **Enforce completion gates.**
    - Apply the Story section of `resources/agent/artifact-standards.md`:
      - `Open Questions` section is **empty**.
      - No task is a figure-it-out item (no "investigate/evaluate/choose/design/determine" where the subject is a requirements, architecture, or design matter).
      - Footprint/dependency metadata is present, specific, and sufficient for `milestone-generation-coordinator` to schedule without reading prose.
      - Every task implements an already-settled decision.
      - Every specialized task kind is registered, explicitly annotated when required, and carries a
        complete contract conforming to the registry entry's `contract_schema`.
      - Acceptance criteria are verifiable.
      - Dependencies are consistent with other stories in the plan.
    - If any gate fails, return to step 9 for revision or escalate per step 10 as appropriate.
    - If `checkpoint_mode` is `every` or `milestone`, this is a mandatory checkpoint before writing.

12. **Write the story file.**
    - Write the finalized draft to `ACTIVE_STORIES/<story_file>`.
    - Verify the file exists, is readable, and contains the expected sections (Goal, Tasks, Footprint and Dependency Metadata, etc.).
    - The story file MUST NOT be written before the design-decision guard (step 9/10) passes and the completion gates (step 11) are satisfied.

13. **Update `ACTIVE_TOC`.**
    - Read `ACTIVE_TOC`.
    - Verify it exists and contains an "Active Stories" section. If not, pause and ask the user whether to terminate.
    - Insert the new entry as a two-line block in the appropriate position (maintain numeric order by story number):
      - Line 1: `**[<story_file>](<story_file>)** — _Not Started_`
      - Line 2: `: <goal summary from the story Goal section>`
    - Insert a blank line before the new entry block and maintain proper list formatting.

14. **Verify the writes.**
    - Confirm the story file exists at `ACTIVE_STORIES/<story_file>` and contains the expected content.
    - Confirm the new entry appears in the Active Stories section of `ACTIVE_TOC` with the correct filename, goal summary, and status.
    - Confirm the execution order graph is consistent with declared parallel groups.
    - Confirm reintegration instructions reference the "story branch" generically, not a hardcoded branch name.
    - If any check fails, revise the story or toc before proceeding.

15. **Present results to the user / invoker.**
    - Story file path and name.
    - Confirmation that the entry was added to `ACTIVE_TOC`.
    - Summary statistics:
      - Total tasks
      - Tasks annotated as `[Small]`
      - Tasks grouped by registered task kind
      - Parallel groups declared
      - Execution order graph
      - Number of review-fix iterations performed
    - Any warnings (missing referenced specs, fewer directories scanned, etc.).
    - Any escalations reported (categories 1–5) with their current status.
    - If `checkpoint_mode` is `every` or `milestone`, this is a mandatory checkpoint.

## Output

The script outputs:

- Confirmation of the new story creation including:
  - Full filename (`story_file`)
  - Absolute path to the created file (`ACTIVE_STORIES/<story_file>`)
  - Plan from which the story was derived
- Completion gate result: confirmation that `Open Questions` is empty, no figure-it-out items exist, and footprint/dependency metadata is complete.
- Confirmation that the story was added to the Active Stories section of `ACTIVE_TOC`.
- Summary statistics:
  - Total number of tasks
  - Number of tasks annotated as `[Small]`
  - Number of tasks by registered task kind
  - Number of parallel groups declared
  - Execution order graph
  - Number of review-fix iterations performed
- Escalation report if any:
  - Category (1–5)
  - Evidence and blocked work
  - Proposed rung (R0–R4) and options
  - Coordinator recommendation
- Any warnings encountered.

## Error Handling

If required parameters are missing or invalid, the user is prompted. The operations WILL NOT execute until all required parameters are valid. During execution, if an error is encountered, execution pauses and the user is asked whether to continue or terminate.

Specific error conditions handled:

- Missing required parameter (`plan_file` or `story_file`)
- Invalid `story_file` format (does not match `^Story\d{3}_[a-zA-Z0-9_-]+\.md$`)
- `plan_file` does not exist
- `ACTIVE_STORIES` directory does not exist
- Story file already exists at `ACTIVE_STORIES/<story_file>` (collision)
- Plan has a non-empty `Open Questions` section (escalation category 2)
- A story task requires a design or architecture decision (escalation categories 1 or 5)
- Missing footprint/dependency metadata (internal defect, returned to `plan-author`)
- Blocking findings persist after 3 review-fix iterations (escalate to invoker)
- `ACTIVE_TOC` file is missing or malformed (missing Active Stories section)
- Permission denied when creating the story file or updating `ACTIVE_TOC`
