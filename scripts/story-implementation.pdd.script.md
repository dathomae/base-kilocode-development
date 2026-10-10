# story-implementation

This script drives the `story-implementation-coordinator` persona for executing a single story. It processes tasks in execution order, dispatches implementors and reviewers per the retry ladders defined in `resources/agent/escalation-policy.md`, handles parallel groups via isolated git worktrees, and completes with end-of-story verification, a walkthrough, a PR into `agent-main`, and CI.

The executing persona is **`story-implementation-coordinator`**. It never implements code or writes documents directly; it delegates all implementation and review, and edits only the story file, tocs, tracking register, and walkthrough. See `.kilo/personas/story-implementation-coordinator.md`.

## Parameters

| Name | Type | Description | Required | Default Value |
|---|---|---|---|---|
| `story_file` | string | Name of the story file under `memory-bank/planning/stories/` (e.g., `Story005_database-initialization-cli.md`) | Yes | N/A |
| `feature_branch` | string | Local feature branch used as the base for all worktrees (e.g., `story/005`) | Yes | N/A |
| `checkpoint_mode` | string | Pause for approval: `"every"` (after each work unit), `"none"` (run to completion), or a task number (pause after that task) | No | `"every"` |
| `source_root` | string | Root source directory for build/test commands (e.g., `src/service`) | No | Prompted when the story does not define it |
| `build_command` | string | Command to build the affected project (e.g., `make build`) | No | Prompted when the story does not define it |
| `test_command` | string | Command to run the relevant full test suite (e.g., `make test`) | No | Prompted when the story does not define it |
| `review_iteration_limit` | number | Maximum implement → review cycles per implementor tier before escalating up the ladder | No | `3` |

## Derived values

| Name | Derivation |
|---|---|
| `sanitized_feature_branch` | `feature_branch` with all `/` replaced by `-` |
| `story_path` | Resolved by searching `active-stories/`, then `finished-stories/`, then `canceled-stories/` under `memory-bank/planning/stories/` |

## Operations

The following operations are performed by **`story-implementation-coordinator`** unless otherwise specified.

1. **Validate parameters.** Confirm `story_file` and `feature_branch` are provided and non-empty. If not, stop and prompt the user.

2. **Verify persona.** Confirm the active persona is `story-implementation-coordinator`. If not, switch and restart.

3. **Locate story file.** Search `memory-bank/planning/stories/active-stories/`, then `finished-stories/`, then `canceled-stories/` for `story_file`. If not found, stop and report.

4. **Ensure `worktrees/` symlink.** If `.kilo/worktrees/` does not exist, create it. Then ensure `worktrees/` is a symlink pointing to `.kilo/worktrees/`:
   ```bash
   mkdir -p .kilo/worktrees/
   ln -sfn .kilo/worktrees/ worktrees
   ```
   Use the `worktrees/` path in all subsequent references; never use `.kilo/worktrees/` directly.

5. **Not-on-main guard.** Check the current branch:
   ```bash
   current_branch=$(git branch --show-current)
   ```
   If `current_branch` is `main` or `master`, stop immediately. This guard has no agent-side or
   user-confirmation override: switch to `agent-main` or a branch created from it before continuing.

6. **Read and parse the story.** Read the story file and extract:
   - **taskKindRegistry**: load `resources/agent/task-kinds.json`; verify every persona and contract
     schema it names exists in the composed project.
   - **Tasks**: numbered tasks with status lines (`Not Started`, `In Progress`, `Completed`, `Blocked`, `Skipped`), optional `[Small]`, and optional `[Kind: <id>]` annotations.
   - **parallelGroups**: map from group label to list of task identifiers (from "Parallel Execution" subsection). Absent or "None" means all tasks are sequential.
   - **executionGraph**: ordered list of nodes, where each node is either a single task identifier or a group label referencing `parallelGroups`.
   - **taskSizes**: map from task identifier to size (`small` if `[Small]` is present, else `medium`).
   - **taskKinds**: map from task identifier to its registered kind. An explicit annotation selects
     that entry; without one, classify only between the core `code` and `document` kinds.
   - **taskContracts**: the contract block for each task whose registry entry names a
     `contract_schema`.
   - **reintegrationInstructions**: map from group label to merge order, integration test commands, and conflict-prone files (from "Reintegration" subsection). Default merge order is finish-order (first reviewed-passed merged first; task-number tiebreaker). Default integration tests are `cd $source_root && $test_command`.
   - **flowsCovered**: the story's `flows:` coverage from the Footprint and Dependency Metadata, and any `[Flow: <id>]`-annotated tasks with their test paths.

7. **Parameter confirmation.** Resolve defaults and confirm all parameter values with the user **once** before any task begins. This is the only parameter query in the session.
   - `checkpoint_mode`: use provided value or default `"every"`.
   - `review_iteration_limit`: use provided value or default `3`.
   - Read the story and its referenced project documentation for the applicable source root, build
     command, and test command.
   - Prompt for any missing value rather than guessing from a language or directory name.
   - Present a configuration summary and ask the user to confirm or modify. Re-present if modified. Record all values for the session.

8. **Select the next work unit.** Walk the `executionGraph` in order:
   - For a single task node: if the task status is `Not Started` or `In Progress`, it is the next work unit.
   - For a group label node: collect all tasks in the group whose status is `Not Started` or `In Progress`. If any exist, they form a parallel multi-task work unit. Tasks in the group that are `Blocked` or `Skipped` are excluded from delegation but noted for reporting.
   - Skip nodes where all tasks are `Completed`, `Blocked`, or `Skipped`.
   - If no actionable tasks remain, proceed to Operation 25 (story completion).

9. **HANDOFF GATE.** Before analyzing task kinds, produce a delegation statement identifying the work unit and declaring: "I am NOT implementing these tasks. I will delegate to the appropriate execution subagents after analyzing task kinds and sizes."

10. **Task kind and size classification.** For each task in the work unit, determine the dispatch target **for delegation only**:
    - `[Small]` annotation present → `small`.
    - Otherwise, apply keyword heuristics: single-file/narrow/well-scoped → `small`; multi-file/new interfaces/full TDD lifecycle → `medium`. Default to `medium` when in doubt.
    - `[Kind: <id>]` present → load that exact registry entry. If it is absent, stop and return the
      story for correction. Never infer a specialization-defined kind from keywords.
    - No kind annotation → determine only `code` vs `document` by keyword analysis.
    - Select the first implementor for `small`; otherwise select the registry's
      `default_implementor` index. The final implementor is never selected for initial delegation
      unless it is the kind's only implementor.
    - If the entry names `contract_schema`, validate the task contract before delegation. Missing or
      invalid fields are blocking; do not guess domain values.

11. **Design-decision guard.** Before creating worktrees or delegating, inspect the task description(s) for design-decision work. Signals include: investigate/research/evaluate options/compare approaches; determine/decide/choose/design/propose an approach, convention, interface, schema, or subsystem boundary; resolve an open question or Unsettled Issue; judgment-call acceptance criteria; "as appropriate", "based on findings", "if needed". If any task requires a decision, **STOP** and escalate immediately (categories 1 or 5 per `resources/agent/escalation-policy.md`). Do NOT delegate the task and do NOT retry it up the implementor ladder — a larger implementor cannot settle a design decision. Report to the user: the task(s), the specific decision(s) needed, why each is a design decision, and suggested ways forward (user supplies decision; record as ADR/design decision; revise story to implement a settled decision; narrow or defer the task). Wait for user guidance.

12. **Pre-work verification.** Before delegating, run automated setup:
    - For refactoring/modification tasks, run existing tests on `feature_branch` to establish a baseline: `cd $source_root && $test_command`.
    - For new development tasks, verify the build succeeds: `cd $source_root && $build_command`.
    - For a specialization-defined kind, also run any exact environment or baseline command required
      by its task contract. If a generic build/test command is inapplicable, the settled contract must
      say so and provide its replacement; do not infer one.
    - **Setup script check:** If `.kilo/setup-script.sh` exists in the repository root, record its path for use during worktree creation. If absent, warn once: "No `.kilo/setup-script.sh` found. Worktree sessions may fail to build or test without environment setup." Continue without blocking.

13. **Worktree creation and resumption (per task).** For each task in the work unit:
    - Check whether the worktree directory `worktrees/<sanitized_feature_branch>-task-N/` or the branch `<sanitized_feature_branch>-task-N` already exists:
      ```bash
      git worktree list | grep "<sanitized_feature_branch>-task-N" || git branch --list "<sanitized_feature_branch>-task-N"
      ```
    - **Resumption**: if the worktree/branch exists AND the task status is `In Progress`, do NOT create a new worktree. Verify the directory is usable and note: "Resuming existing worktree for Task N at worktrees/<sanitized_feature_branch>-task-N/".
    - **Stale worktree**: if the worktree/branch exists but the task status is NOT `In Progress` (e.g., `Not Started`), delete the stale worktree and branch first:
      ```bash
      git worktree remove --force worktrees/<sanitized_feature_branch>-task-N && git branch -D <sanitized_feature_branch>-task-N
      ```
      Then proceed with normal creation.
    - **Creation**: if neither exists, create:
      ```bash
      git worktree add worktrees/<sanitized_feature_branch>-task-N -b <sanitized_feature_branch>-task-N
      ```
    - **Environment setup**: If `.kilo/setup-script.sh` exists, run it inside the newly-created worktree using `workdir` set to `worktrees/<sanitized_feature_branch>-task-N`. If it fails, report the error, delete the worktree and branch, mark the task as `Blocked` with reason "Setup script failed in worktree", and do NOT dispatch.

14. **Dispatch implementors.** Update the story file to mark each delegated task as `In Progress` **before** dispatch. Then dispatch each task to its size-appropriate implementor via the Task tool, passing a `work_context` block per `resources/agent/work-context.md`:
    ```yaml
    work_context:
      work_id: "<story_id>-TaskN"
      artifact: "<artifact from the task-kind registry>"
      task_kind: "<registered task-kind id>"
      level: task
      assignment: "<one-sentence statement of what the subagent must produce>"
      source_root: "<source_root>"
      worktree_path: "worktrees/<sanitized_feature_branch>-task-N"
      base_branch: "<feature_branch>"
      paths:
        primary: [<files/packages expected to be touched>]
        read_only: [<files to read for context>]
      standards: [<extends the default manifest>]
      upstream_decisions: [<settled decisions the work must implement, by reference>]
      acceptance_criteria: [<verifiable criteria>]
      constraints: []
      task_contract: {<required when the task kind names a contract schema>}
      expected_return: "implementation summary, files, test results, Blocked status if applicable"
      blocked_conditions: [<conditions requiring a decision the subagent is not authorized to make>]
    ```
    - **Single task**: dispatch one implementor, await result.
    - **Parallel group**: dispatch all implementors simultaneously in a single message as concurrent Task tool invocations. Mixing task kinds within a group is acceptable when their files and dependencies are disjoint.
    - Instruct each implementor to use `workdir` set to `worktree_path` for all bash commands.
    - Instruct each implementor to commit all changes in the worktree before returning.
    - Instruct each implementor to return the `work_context` unchanged with its result.
    - After dispatching, explicitly state which agents were dispatched, the tasks assigned, their worktree paths, and confirm: "I am NOT implementing these tasks. Awaiting sub-agent results."

15. **Receive implementor results.** Upon receiving a result from an implementor:
    - **Validate `work_context`**: confirm `work_id`, `worktree_path`, and task identity match.
    - **`Blocked — needs decision`**: If the status is `Blocked — needs decision`, do NOT consume a retry ladder rung. Escalate immediately per `resources/agent/escalation-policy.md` (categories 1 or 5). Hold the branch — do not merge. Stop and report to the user or `milestone-implementation-coordinator` per the escalation routing rules in `resources/agent/escalation-policy.md`.
    - **`Completed`**: Proceed to review (Operation 16).
    - **`Failed` or incomplete**: Apply the selected task kind's registered implementor ladder:
      1. Re-dispatch the current implementor from a fresh worktree up to
         `failure_retries_per_implementor` times, carrying the full failure evidence.
      2. When those retries are exhausted and another registered implementor exists, advance to it,
         recreate the worktree from `feature_branch`, remove `[Small]` if present, and re-dispatch.
      3. When the final registered implementor exhausts its retries, stop, delete the worktree and
         branch, and report the full failure chain and diagnosis to the invoker.
    - **Operation-limit retry:** An implementor that hit only its operation limit may be re-dispatched
      in the same worktree up to 3 times. If still incomplete, treat it as a failure under the
      registered ladder.

16. **Dispatch reviewers and verifiers.** Once a task's implementor reports `Completed`, dispatch the
    task kind's registered reviewer via the Task tool with the unchanged `work_context`:
    - For parallel groups, dispatch one reviewer per task simultaneously in a single message.
    - Include the task description, scope, file paths of modified/created files, and instruct the reviewer to return the `work_context` unchanged.
    - Reviewers judge conformance to intent per `resources/agent/report-formats.md`. They do not modify files.
    - After the reviewer passes, if the registry entry names a `verifier`, dispatch it independently
      with the same revision and unchanged context. It returns the Artifact-verification result from
      `resources/agent/report-formats.md`. A task cannot merge until the verifier returns `Passed`.

17. **Evaluate review results.** Upon receiving a review report:
    - Validate `work_context`.
    - **Pass** (no issues or only preexisting issues): proceed to the registered verifier, or to merge
      (Operation 19) when the kind has no verifier.
    - **Fail** (task-introduced issues found):
      - All task-introduced issues — blocking, minor, and DRY recommendations — MUST be addressed. Do not filter or skip any issue.
      - First, determine whether the issues indicate poor implementation quality (blocking issues, multiple issues suggesting the implementation is not solid, structural problems). If so, apply the implementor escalation ladder (Operation 15) before re-review.
      - If the implementation is fundamentally sound but needs adjustments, delegate the fix back to the **original** implementor agent by creating a new Task with `workdir` set to the same worktree path. Pass **all** review feedback as part of the task description. Instruct the agent to address every issue and commit.
      - Bound the review-fix loop: up to `review_iteration_limit` cycles per implementor tier. If issues persist after the limit, stop and ask the user for guidance. Do NOT silently escalate.
      - Self-check before each fix delegation: "I am not applying these fixes myself. I am delegating the revision to [agent name]."
    - **Verifier `Failed`:** re-dispatch the current implementor with every failed check and its exact
      evidence, then repeat reviewer and verifier. Apply the same review-fix and registered failure
      bounds. **Verifier `Blocked`** escalates immediately and does not consume a retry.

18. **Update task status immediately.** As soon as a task passes review and any registered independent
    verification, update the story file to mark it `Completed` before proceeding. Do not batch status updates. Commit the story file as part of the checkpoint.

19. **Merge worktree into feature branch (Worktree Merge Protocol).** Merge successful worktrees one at a time:
    - **Merge order**: Follow the reintegration instructions for the group if specified. If "any order" or absent, use finish-order (first reviewed-passed first; task-number tiebreaker). Sequential tasks merge trivially.
    - **Pre-merge commit check**: Verify the worktree branch has at least one commit beyond `feature_branch`:
      ```bash
      git log <feature_branch>..<sanitized_feature_branch>-task-N --oneline
      ```
      If no commits, treat as subagent failure: report, delete the worktree, do NOT mark the task `Completed`.
    - Merge:
      ```bash
      git checkout <feature_branch>
      git merge <sanitized_feature_branch>-task-N
      ```
    - On merge conflict: pause immediately. Present conflict details to the user. Do NOT auto-resolve. Options: user resolves manually, or abort with `git merge --abort`.
    - After successful merge, delete the worktree and branch:
      ```bash
      git worktree remove worktrees/<sanitized_feature_branch>-task-N
      git branch -d <sanitized_feature_branch>-task-N
      ```
      If `git branch -d` fails because already merged, use `git branch -D <sanitized_feature_branch>-task-N`.
    - **Integration verification**:
      - For sequential tasks: run a lightweight build check on `feature_branch`: `cd $source_root && $build_command`.
      - For parallel groups: run integration tests only after **all** worktrees in the group have been merged. Use the reintegration instructions' integration test command, or default to `cd $source_root && $test_command`.

20. **Checkpoint and approval.** After each work unit (single task or parallel group):
    - Commit the current state with a descriptive message.
    - If `checkpoint_mode` is `"every"`: present a work summary and **wait** for explicit user approval before proceeding.
    - If `checkpoint_mode` is `"none"`: present a brief status update and continue automatically.
    - If `checkpoint_mode` is a task number: present a brief status and continue automatically until that task completes, then pause for approval. After approval, ask the user for the next checkpoint target (`"none"`, `"every"`, or another task number).

21. **Loop to next work unit.** After checkpoint resolution, return to Operation 8 to select the next work unit. Continue within the same session.

22. **Blocked/Skipped tasks in work unit.** After merging all successful tasks in a work unit, if any tasks were `Blocked` or `Skipped`:
    - Present a summary of which tasks and why.
    - Ask the user for direction: retry after fixing the underlying issue, skip permanently, or abort the story.
    - Do NOT proceed to the next work unit without user direction.

23. **Story completion — flow tests and result review.** When all tasks in the story are `Completed`, `Blocked`, or `Skipped` and no actionable work remains:
    - **Run declared flow tests.** For each flow in the story's `flowsCovered` whose functionality now exists on the branch (its `checkable_after` is satisfied), run the flow's automated test through the flows-suite interface, or dispatch `flow-verifier` when the flow's `verification` is `agent`. Record each result per `resources/agent/report-formats.md`. A `Failed` flow blocks story completion: re-dispatch the relevant implementor(s) with the failure, bounded by the review-fix limit. A flow whose functionality is not yet present is left to milestone close — do not fake a result.
    - Dispatch `story-reviewer` via the Task tool with a `work_context` block. Instruct it to verify the union of tasks fulfills the story goal, completion criteria, and acceptance criteria per `resources/agent/report-formats.md`.
    - Evaluate the `story-reviewer` result. If it reports blocking findings, address them by re-dispatching the relevant implementor(s) with the review feedback, applying the review-fix bound per `resources/agent/escalation-policy.md`.

24. **Story completion — critic pass.** After `story-reviewer` passes:
    - Collect the distinct critics registered by all task kinds in the story and dispatch them
      simultaneously. Core-only stories therefore retain the code/write behavior; composed
      specializations add their critics without changing this script.
    - Include the walkthrough (once written) in the critic's scope so it can review for anomalies (unexplained deviations, silent scope changes, skipped verification, decisions made without authority) per `resources/agent/report-formats.md`.
    - Evaluate the critic report. Blocking findings from a critic are mandatory and never waivable. Address them by re-dispatching implementors, applying the review-fix bound. If findings cannot be resolved within the bound, escalate per `resources/agent/escalation-policy.md`.

25. **Write walkthrough.** Write the completion walkthrough per `resources/templates/walkthrough.template.md`. Name it with the story file stem plus `.walkthrough.md` and save it to `memory-bank/planning/stories/finished-stories/walkthroughs/`. List it in that directory's `toc.md`.

26. **Push and open PR via `integrator`.**
    - Rebase `story/<id>` onto current `agent-main` (or the base branch provided by `milestone-implementation-coordinator`).
    - Dispatch `integrator` via the Task tool with a `work_context` block instructing it to:
      - Force-push the story branch (`story/<id>`).
      - Open a PR from `story/<id>` into `agent-main`.
      - Report the PR URL and CI status.
    - The integrator never merges and never pushes `main`. See `.kilo/personas/integrator.md`.

27. **Drive CI green.**
    - If CI passes, proceed to Operation 28.
    - If CI fails, read the failure log. Re-dispatch the relevant implementor with the failure log and instructions to fix in the existing worktree (or a fresh one if the worktree was already deleted). Bound retries per `resources/agent/escalation-policy.md`.
    - On repeated CI failure after exhausting retries, escalate to `milestone-implementation-coordinator` or the human.

28. **End-of-story activities.** Perform the story-completion steps in `.kilo/rules/end_of_task_activities.md` **before** presenting the result to the user:
    1. **Run tests** (`cd $source_root && $test_command`) where applicable, plus every verification
       command required by completed task contracts. If any fail for any reason, stop and ask the user what to do next.
    2. **Update the artifact.** Update task statuses in the story file and the story's status in `memory-bank/planning/stories/active-stories/toc.md`.
    3. **Move to finished.** Move the story file from `active-stories/` to `finished-stories/`. Remove the entry from `active-stories/toc.md` and add it to `finished-stories/toc.md` with status `_Done_`; update the link to be relative to the new `toc.md` (just the filename).
    4. **Update tracking register.** Update `memory-bank/planning/tracking/system-definition-tracking-status.md`: remove the story from `Remaining stories`, advance `Implemented`, and advance `Planned` if the full story set now exists.
    5. **Walkthrough already written** in Operation 25; ensure it is listed in `finished-stories/walkthroughs/toc.md`.
    6. Commits and pushes follow `resources/agent/git-workflow.md`.

29. **Clean up remaining worktrees.** List any remaining worktrees:
    ```bash
    git worktree list | grep "<sanitized_feature_branch>-"
    ```
    For each, delete it and its branch:
    ```bash
    git worktree remove --force worktrees/<sanitized_feature_branch>-task-N
    git branch -D <sanitized_feature_branch>-task-N
    ```
    Note any orphaned directories for manual cleanup.

30. **Report to milestone tracking.** Report the story completion to `milestone-implementation-coordinator` (if invoked as its subagent) or note it for the milestone file update. `milestone-implementation-coordinator` owns updating the milestone file, wave progress, and remediation log per `.kilo/rules/end_of_task_activities.md` step 5.

## Output

The script outputs:

- Confirmation of the story file being processed and the resolved configuration summary.
- The current work unit (single task or parallel group with task list) being executed.
- Worktree paths and branch names created or resumed for each task.
- Setup script execution status per worktree (or the one-time warning if absent).
- Task-kind and size classification for each task, and the implementor dispatched.
- Implementor result status (`Completed`, `Blocked — needs decision`, `Failed`).
- Review verdict (`pass` / `fail`) and a summary of findings.
- Flow-test results for the flows covered where the functionality exists (`Passed`/`Failed`), and any flow deferred to milestone close.
- The number of review iterations consumed per task.
- Merge confirmations with commit hashes, and integration test results.
- Escalation notifications: which agent failed, which tier it escalated to, and worktree deletion/re-creation steps.
- Big-tier diagnosis if the top implementor tier failed.
- Final completion status, PR URL, and CI status when all tasks are done.

## Error Handling

If required parameters are not specified, the user is prompted. Operations will not execute until all required parameters are provided.

During execution, if an error is encountered, execution pauses and the user is asked whether to continue or terminate. Specific error conditions:

- Story file not found or unreadable.
- Story file format invalid (missing Tasks section).
- Unable to determine task type (ambiguous content).
- Task requires a design or architecture decision (Operation 11 stops for user guidance).
- Implementor returns `Blocked — needs decision` (immediate escalation, not ladder consumption).
- Implementor fails at any tier (escalation ladder per `resources/agent/escalation-policy.md`).
- Reviewer finds blocking or quality-indicating issues (review-fix loop or escalation).
- A flow test fails at story completion (blocking; re-dispatch or defer to milestone close if the flow's functionality is not yet present).
- Maximum review iterations exceeded (`review_iteration_limit`).
- Git merge conflict during worktree reintegration (presented to user, not auto-resolved).
- Git commit failure within a worktree (subagent responsibility).
- Pre-merge commit check finds no commits on the worktree branch (subagent failure).
- Worktree creation fails (`git worktree add` error, branch collision, disk full).
- Worktree deletion fails (orphaned directories, locked worktree).
- `.kilo/setup-script.sh` execution fails in a worktree (task marked `Blocked`).
- `worktrees/` symlink creation fails.
- Not-on-main guard triggered; switch to `agent-main` or a branch created from it.
- CI failure (re-dispatch implementor with failure log; bounded retries).
- End-of-story test failure (Operation 28 step 1).
- Story toc.md or tracking register update failures.

## Notes

- **Coordinator conventions** (dispatch, work context, parallelism, disposition, remediation, completion gate, reports) follow the "Coordinator scripts" section of `resources/templates/prompt-driven-development-script.template.md`.
- **Escalation policy, retry definitions, report formats, and git rules** are not restated here; cite `resources/agent/escalation-policy.md`, `resources/agent/report-formats.md`, and `resources/agent/git-workflow.md`.
- **Subagents do not "create a new task back."** The coordinator receives each subagent's result as a normal tool result. The coordinator then decides the next step.
- **Worktrees/** is the only path used; never `.kilo/worktrees/`.
- **Status discipline:** Update task statuses in the story file immediately upon state change. Never batch updates. Commit the story file with each status change.
- **Context management:** The coordinator's context is reserved for story-level coordination. It does not read full source files for implementation detail; it delegates that to implementors.
- **Permission boundary:** The coordinator holds `edit` only for subagent permission inheritance. It writes only the story file, walkthrough, tocs, and tracking register. All source code and document changes are delegated to implementors.
