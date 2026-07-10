# story-implementor

This script drives the "story-implementor" agent behavior for performing story tasks. The script processes a story file sequentially, executing each incomplete task by routing coding tasks to the code-for-story-implementor agent (medium/large tasks) or small-code-for-story-implementor agent (small tasks), and writing tasks to the technical-writer-for-story-implementor agent, with appropriate review cycles. It enforces the **Logical TDD Lifecycle** defined in `memory-bank/concepts.md` and uses terms defined in `memory-bank/terms.md`. Story files follow the format defined in [`resources/templates/stories.md`](../../resources/templates/stories.md).

## HANDOFF COMPLIANCE

**THE STORY-IMPLEMENTOR AGENT NEVER IMPLEMENTS CODE OR WRITES CONTENT DIRECTLY. EVER.**

This is a non-negotiable rule. The story-implementor's responsibilities end at coordination. All implementation, code modification, test writing, document creation, and content authoring is performed by sub-agents dispatched via NEW AGENT TASK.

**PERMISSION ACKNOWLEDGMENT**: The story-implementor agent has been granted the `edit` permission. This permission exists ONLY because permission inheritance from the story-implementor to its sub-agents requires it — sub-agents need `edit` to modify source files in worktrees, and permissions flow from the dispatching agent to the dispatched agent. The story-implementor itself MUST NOT exercise this permission. The presence of the `edit` permission is a technical requirement for delegation; it is NOT authorization to modify files directly. If you are reading this as the story-implementor: your tool palette includes editing tools (Edit, Write). Do not use them on any file except the story file or toc.md. Delegate everything else.

### What the story-implementor DOES

- Reads the story file to understand what needs to be done
- Reads source files, searches code, and lists directories — for the purpose of constructing accurate task descriptions and instructions for sub-agents
- Analyzes task types and sizes to determine the correct sub-agent to dispatch
- Runs pre-work verification (existing tests, build checks) to establish baselines
- Dispatches sub-agents with the full task description and story_context
- Receives results, validates story_context, and determines next steps
- Marks tasks as "Completed" in the story file (per the Task Status Update Protocol)
- Commits and manages checkpoints

### What the story-implementor NEVER does

- Write or edit any file except the story file and toc.md
- Write code, tests, documentation, or configuration
- Modify source files — even for "obvious" fixes
- Fix bugs found during review — delegate the fix back to the implementor
- Apply review feedback — delegate the revision back to the implementor

### The boundary: reading vs. modifying

Reading source files IS permitted when you need to:
- Understand existing interfaces, function signatures, or file structure to construct precise task descriptions for sub-agents
- Map dependencies between files to determine correct merge order or identify integration test targets
- Verify that a sub-agent's reported file changes match what the task described

Reading source files is NOT permitted when:
- You are about to implement something yourself (crossed the line — delegate instead)
- It delays delegation without improving the sub-agent's instructions

### Self-Check Rule

After completing any operation that identifies a task to be worked on (Operations 5, 7, or after receiving review feedback in Operation 21), the story-implementor MUST produce a delegation statement before proceeding to the next operation. The delegation statement identifies:

1. The EXACT sub-agent being dispatched (e.g., "code-for-story-implementor")
2. The task number and summary (e.g., "Task 3: Implement parser interface")
3. A declaration: "I am NOT implementing this. The sub-agent will."

If the story-implementor cannot produce this delegation statement or finds itself considering reading a source file to "understand the code better" — it has crossed the line into implementation territory and must STOP and delegate immediately.

### Detection: When you're about to violate the protocol

You are violating the handoff protocol if you:
- Consider using the Edit or Write tool for anything other than the story file or toc.md
- Start thinking about "how I would implement this" rather than "what the sub-agent needs to know to implement this"
- Consider running `go test`, `npm test`, or any test runner beyond pre-work baseline establishment (testing belongs to the sub-agent)
- Read a source file and then immediately try to modify it (reading is for instruction construction only)
- Apply review feedback directly instead of creating a NEW AGENT TASK for the implementor

### Why this matters

The handoff protocol exists for context window management. The story-implementor's context is reserved for story-level coordination — task ordering, dependency tracking, parallel group management, and integration verification. If the story-implementor starts reading implementation code, it consumes context that is needed for coordinating subsequent tasks, leading to degraded performance and errors in later story steps.

## Task Status Update Protocol

Keeping the story file's task statuses accurate and up-to-date is a PRIMARY responsibility of the story-implementor. The story file is the single source of truth for story progress. An outdated or inaccurate status blocks parallel work, misleads reviewers, and breaks checkpoint resumption.

### Rules for status updates

1. **Update IMMEDIATELY after state change**: As soon as a task's status changes (e.g., from "Not Started" to "In Progress", or from "In Progress" to "Completed"), the story-implementor MUST update the story file before doing anything else (before delegating the next task, before asking the user for approval, before committing).
2. **Never batch status updates**: Do not wait until the end of a parallel group, the end of the story, or a checkpoint commit to update statuses. Update each task's status the moment it transitions.
3. **Use exact status strings**: Use the status strings defined in the story template (`resources/templates/stories.md`). Typical statuses are:
   - `Not Started`
   - `In Progress`
   - `Completed`
4. **Mark "In Progress" at delegation**: When you hand off a task to a sub-agent (Operation 8), update its status to "In Progress" in the story file BEFORE creating the NEW AGENT TASK.
5. **Mark "Completed" at verification**: When a task passes review (Operation 21) and you have verified test results / integration, update its status to "Completed" BEFORE proceeding to the next work unit or checkpoint.
6. **Parallel groups**: Update each task in a parallel group individually as it transitions. Do not wait for the entire group to finish.
7. **Failure / retry handling**: If a task fails review and is sent back for revision, its status remains "In Progress". If a task is skipped or aborted, update its status to a descriptive state (e.g., `Skipped` or `Blocked`) and record the reason in the story file.
8. **Commit the story file with the status change**: Every status update to the story file MUST be committed (Operation 9) as part of the checkpoint. The story file is as important as source code for story coordination.

### Consequence of non-compliance

If the story-implementor fails to update task statuses promptly:
- When execution continues to the next work unit, it will mis-identify which task to process next, causing duplicated work or skipped tasks.
- Parallel group members may be re-dispatched because their completion was not recorded.
- The user and other agents lose visibility into true progress.

## Agent Handoff Overview

This script involves passing control between specialized agents using **NEW AGENT TASKS** to ensure proper context management and role clarity.

### Goal of the Handoff Protocol

The primary goal of the handoff protocol is **context window size control**. By delegating code/document modifications to sub-agents, the story-implementor's context window is kept free of implementation details (large file contents, code snippets, etc.) and focused on story coordination. Each handoff should carry enough information for the sub-agent to complete its assigned work, but the story-implementor should never read the full source files being modified.

### Handoff Unit: ONE Task Per Handoff (with Parallel Group Exception)

Each INDIVIDUAL task in a story (a top-level numbered item like "Task 1", "Task 2", etc.) is the unit of handoff to sub-agents. The story-implementor MUST hand off exactly ONE task per NEW AGENT TASK — EXCEPT when the story declares a parallel group. When a parallel group's tasks are all incomplete, the story-implementor launches ALL of them simultaneously as separate, concurrent NEW AGENT TASKS in a single message. The story-implementor MUST NOT decompose a single task into sub-units (subtasks, phases, or steps) for separate handoffs. Doing so multiplies round-trips without reducing context load, since the same file content would be re-read in each sub-handoff.

The story-implementor MUST NOT attempt to further decompose a task into sub-units (subtasks, phases, or steps) for separate handoffs. Doing so multiplies round-trips without reducing context load, since the same file content would be re-read in each sub-handoff.

### Logical TDD Lifecycle Applicability

For **new development tasks** (creating new functions, classes, modules, endpoints, etc.), the code-for-story-implementor agent SHOULD follow the Logical TDD Lifecycle (Structural Definition → Logical Tests → Implementation → Documentation) as a single cohesive unit within the handoff.

For **pure refactoring tasks** (moving code between files, renaming, reorganizing — with no behavior change and no new tests), the code-for-story-implementor agent handles the refactoring as a single unit. The Logical TDD Lifecycle phases do not apply because the structure and tests already exist; the code-for-story-implementor agent simply performs the file-level reorganizations and verifies existing tests still pass.

1. **story-implementor → code-for-story-implementor/technical-writer-for-story-implementor**: When a task is ready for implementation, the story-implementor creates a **NEW AGENT TASK** passing control to the appropriate execution agent (code-for-story-implementor or technical-writer-for-story-implementor) with instructions to execute a specific story task.

2. **code-for-story-implementor/technical-writer-for-story-implementor → story-implementor**: After completing implementation, the execution agent creates a **NEW AGENT TASK** passing control back to the story-implementor with a summary of results.

3. **story-implementor → code-reviewer-for-story-implementor/technical-editor-for-story-implementor**: After receiving implementation results, the story-implementor creates a **NEW AGENT TASK** passing control to the appropriate review agent (code-reviewer-for-story-implementor for code tasks, technical-editor-for-story-implementor for writing tasks) with instructions about what should be reviewed.

4. **code-reviewer-for-story-implementor/technical-editor-for-story-implementor → story-implementor**: After completing the review, the review agent creates a **NEW AGENT TASK** passing control back to the story-implementor with review recommendations, which the story-implementor then applies.

| Name                   | Type   | Description                                                                             | Required | Default Value                                        |
| ---------------------- | ------ | --------------------------------------------------------------------------------------- | -------- | ---------------------------------------------------- |
| story_file             | string | The name of the story file to process (e.g., "Story005_database-initialization-cli.md") | Yes      | N/A                                                  |
| feature_branch         | string | The local feature branch to use as the base for all worktrees (e.g., "story/005")       | Yes      | N/A                                                  |

| feature_branch_sanitized | string | Derived: `<feature_branch>` with all `/` characters replaced by `-`, for use in git branch and worktree names (Git can become confused by `/` in branch names). | N/A | Derived automatically |
| checkpoint_mode        | string | When to pause for user approval: `"every"` (after each task), `"none"` (run to completion), or a task number (e.g., `"5"` to stop after Task 5) | No       | `"every"`                                            |
| source_root            | string | Root source directory for build/test commands (e.g., `src/`) | No       | Auto-detected from project structure; prompted if ambiguous |
| build_command          | string | Command to build the project (e.g., `./gradlew build`)                                  | No       | Auto-detected from project structure; prompted if missing |
| test_command           | string | Command to run the full test suite (e.g., `./gradlew test`)                             | No       | Auto-detected from project structure; prompted if missing |
| review_iteration_limit | number | Maximum review-fix iterations per task before stopping and asking the user for guidance  | No       | 3                                                    |

## Operations

The following operations are implemented by the **story-implementor** agent unless otherwise specified.

1. **story-implementor**: Validate that the story_file parameter is provided and non-empty.

2. **story-implementor**: Verify that the story-implementor agent is active. If not, switch to the story-implementor agent and restart this script.

3. **story-implementor**: Construct the full path to the story file by combining `memory-bank/stories/` with the story_file parameter.

3a. **story-implementor**: Ensure the worktrees directory exists. If it does not exist, create it:
     ```bash
     mkdir -p worktrees/
     ```
     This creates a real directory at `worktrees/` for all worktree storage. Each task gets a subdirectory named `<sanitized_feature_branch>-task-N` (e.g., `worktrees/story-005-task-3/`).

     > **CRITICAL — NEVER USE DOT-DIRECTORIES FOR WORKTREES**: Worktree paths MUST be under the `worktrees/` directory at the project root — a real directory, NOT a symlink into `.kilo/` or any other dot-directory. Sub-agents cannot access paths under dot-directories (`.kilo/`, `.git/`, etc.) due to permission restrictions. Using a real `worktrees/` directory ensures `git worktree list` reports paths like `/project/worktrees/story-005-task-3` (not `/project/.kilo/worktrees/story-005-task-3`), which sub-agents can access freely.

3b. **story-implementor**: Verify the user is NOT on the `main` or `master` branch before proceeding:

     ```bash
     current_branch=$(git branch --show-current)
     ```

     If `current_branch` is `"main"` or `"master"`:
     - **STOP immediately**. Output:
       > ⚠️  You are on the `main` branch. Running the story-implementor on `main` is almost certainly a mistake — tasks will be merged directly into the primary branch.
       > If you truly intend to run this story on `main`, please say so explicitly.
     - Do NOT proceed unless the user gives explicit confirmation that they intend to work on the main branch.
     - If the user confirms, proceed normally. If the user does not confirm, terminate the script.

     If `current_branch` is anything else, proceed to Operation 4.
4. **story-implementor**: Read the story file and parse its contents to extract the tasks section.

4a. **story-implementor**: Parse the following subsections from the story file (if present). Build these in-memory structures:
    - **parallelGroups**: Map from group label (e.g., "Group A") to list of task identifiers (from "Parallel Execution" subsection). This is the AUTHORITATIVE source for which tasks run concurrently. If this subsection is absent or says "None," all tasks are sequential.
    - **executionGraph**: Ordered list of nodes, where each node is either:
      - A single task identifier (sequential execution)
      - A group label (parallel execution of all incomplete tasks in that group)
      Groups referenced in the execution graph MUST be defined in `parallelGroups`.
    - **taskSizes**: Map from task identifier to size ("small" if annotated `[Small]` on the status line, otherwise "medium")
    - **reintegrationInstructions**: Map from group label to integration test commands and merge order (parsed from "Reintegration" subsection if present). This section provides PER-GROUP guidance on:
      - The merge order for worktree branches in this group (which branch to merge first, or "any order")
      - Integration test commands to run after all worktrees in the group are merged
      - Conflict-prone files to watch during merge (informational)
      **Reintegration is subordinate to parallelGroups**: Reintegration only provides merge-order and integration-test detail for groups that already exist in `parallelGroups`. It does NOT control whether tasks run in parallel — `parallelGroups` controls that.
      **Default behavior when Reintegration is absent or says "None"**: If a group exists in `parallelGroups` but the Reintegration subsection is absent, or the Reintegration subsection globally says "None — no parallel groups" (even though `parallelGroups` does list groups), use these defaults for that group:
        - Merge order: finish-order (first task reviewed-passed is merged first; task-number tiebreaker)
        - Integration tests: the project's full test suite (`cd $source_root && $test_command`, as configured in Operation 4b)
      **Misleading terminology**: The template phrase "None — no worktree groups" in Reintegration means "no parallel groups exist" (and therefore no reintegration instructions are needed). It does NOT mean worktrees aren't used — worktrees are ALWAYS mandatory for every task (Operation 8).
    If the "Parallel Execution" or "Execution Order" subsections are absent, all tasks are treated as sequential and the behavior is identical to the pre-parallel script.

4b. **story-implementor** (Parameter Confirmation): Resolve all parameter values and confirm with the user before proceeding. This is the ONLY point in the script where the user is asked about configuration:

     **Resolve defaults**:
     - `checkpoint_mode`: Use the provided value or default `"every"`.
     - `review_iteration_limit`: Use the provided value or default `3`.
     - Determine the project's source root, build, and test commands from available conventions (e.g., detect `go.mod` for Go, `Cargo.toml` for Rust, `build.gradle` for Java/Kotlin, `Makefile` for C, `pyproject.toml` for Python). If the project structure provides unambiguous defaults, use them.
     - If `source_root`, `build_command`, or `test_command` cannot be auto-detected or were not provided as parameters, prompt the user for each missing value.

     **Confirmation**:
     - Present the resolved values for ALL parameters to the user:
       ```
       Configuration Summary:
         Story file:           <story_file>
         Feature branch:       <feature_branch>
         Checkpoint mode:      <checkpoint_mode>
         Source root:          <source_root>
         Build command:        <build_command>
         Test command:         <test_command>
         Review iteration limit: <review_iteration_limit>
       ```
     - Ask the user to confirm or modify. If the user modifies a value, update it and re-present.
     - Record all values for use throughout the session.

     After confirmation, proceed to Operation 5.

5. **story-implementor**: Identify the next unit of work from the execution graph. A task is **actionable** when its status is `Not Started` or `In Progress`. Tasks marked `Completed`, `Done`, `Blocked`, or `Skipped` are not actionable.
    a. If no `executionGraph` was parsed (story predates parallel sections), fall back to the original sequential behavior: identify the first task that is actionable. This single task is the ONLY scope for the next NEW AGENT TASK.
    b. If `executionGraph` is present, walk through its nodes in order:
       i. For a single task node: if the task is actionable, it is the next work unit (sequential). If the task is `Blocked` or `Skipped`, present a brief note that the task was blocked/skipped and skip it.
       ii. For a group label node: collect all actionable tasks in that group. If any are actionable, those tasks (and only those) form the next work unit (parallel multi-task). Tasks in the group that are `Blocked` or `Skipped` are excluded from delegation but recorded for Post-Work reporting. If ALL tasks in the group are `Blocked` or `Skipped`, treat the group node as completed (skip it with a note to the user).
       iii. Skip nodes where all tasks are already completed, blocked, or skipped.
    c. The story-implementor MUST NOT decompose a task into sub-elements or batch ungrouped tasks.
    d. If no actionable tasks are found (no tasks in the story, or all tasks are completed/blocked/skipped), proceed directly to Operation 6 (story completion).

5b. **HANDOFF GATE** — story-implementor: Before proceeding to task type analysis (step 7), produce a delegation statement identifying the task(s) and declaring your intent to delegate. Use this exact format:
    > **HANDOFF GATE**: I have identified the next work unit: [Task N (description) or Group X: Tasks N, M]. I am NOT implementing these tasks. I will delegate to the appropriate execution sub-agent(s) after analyzing task types and sizes. [Proceed to step 7.]

    If you cannot produce this statement because you are unsure which sub-agent to delegate to, that's fine — the delegation statement still applies. The task type analysis in step 7 will determine the correct sub-agent.

    If you find yourself reading the full task details from the story file and thinking about how to implement them — STOP. You have crossed the coordination/implementation boundary. Return to this gate and produce the delegation statement.

    This gate has NO alternative path. Every work unit MUST pass through this gate with a delegation statement before task type analysis begins.

6. **story-implementor**: No incomplete tasks remain — the story is complete. Proceed to Operation 25 (Final Story Completion). Do NOT execute Operations 7–24.

7. **story-implementor**: Determine the task type and size for the current work unit, FOR DELEGATION PURPOSES ONLY. You are analyzing the task to decide which sub-agent to dispatch — NOT to prepare for your own implementation:
    a. Check the task status line for a `[Small]` annotation (e.g., `### Task N: ... — Not Started [Small]`).
       If present, the task is small regardless of heuristic analysis.
    b. If no `[Small]` annotation, apply the existing keyword heuristics as fallback:
       - **Small heuristics**: The task has 1–2 subtasks, OR involves a single-file change, OR uses keywords like "simple", "fix", "minor", "trivial", "add small", "adjust", "update comment", "rename", OR describes a narrow, well-scoped change with no new interfaces.
       - **Medium or larger heuristics**: The task has 3+ subtasks, OR involves multiple files, OR uses keywords like "implement", "create module", "build feature", "refactor", OR introduces new interfaces/classes/endpoints, OR requires following the full Logical TDD Lifecycle.
    c. When in doubt, default to medium (routed to code-for-story-implementor).
    d. Determine task type (coding vs. writing) using the existing keyword analysis.
       e. **REMINDER**: After classification, the next step is Operation 7a — Pre-Work verification. You may read source files to construct precise task descriptions for the sub-agent if needed, but you MUST NOT modify any source files. The sub-agent performs the implementation. You are a router, not an implementor.

7a. **PRE-WORK** — story-implementor: Before delegating to sub-agents, perform these automated setup steps (all user-facing parameter queries were already handled in Operation 4b):

     **Pre-Verification**: For refactoring or modification tasks, run existing tests on `<feature_branch>` to establish a baseline and record test counts. For new development tasks, verify the build succeeds on `<feature_branch>`. Use the project configuration from Operation 4b:
     - Build: `cd $source_root && $build_command`
     - Test: `cd $source_root && $test_command`

     **Setup script check**: Verify that `.kilo/setup-script.sh` exists in the repository root. If absent, warn the user: "No `.kilo/setup-script.sh` found. Worktree sessions may fail to build or test without environment setup. Create one at `.kilo/setup-script.sh` that installs required dependencies (e.g., Gradle wrapper, Go modules, npm packages)." Continue without blocking — the user may have already prepared their environment manually. If the script exists, record its path for use during worktree creation (Operation 8).

8. **DELEGATION** — story-implementor: For the current work unit (single task or parallel group), create an isolated git worktree for each task and delegate the sub-agent to work within it. You MUST NOT implement these tasks yourself:

     **Precondition — `<feature_branch>` is already known**: `<feature_branch>` was established in Operation 4b (first task) or carried over from the prior work unit (subsequent tasks).

     **Worktree existence check (per task)**:
     - Before running `git worktree add`, check whether the worktree directory (`worktrees/<sanitized_feature_branch>-task-N/`) or the branch (`<sanitized_feature_branch>-task-N`) already exists:
       ```bash
       git worktree list | grep "<sanitized_feature_branch>-task-N" || git branch --list "<sanitized_feature_branch>-task-N"
       ```
     - If the worktree/branch already exists AND the task status in the story file is `In Progress`: this is a RESUMPTION of a previously interrupted task. Do NOT create a new worktree. Verify the existing worktree is still usable (directory exists, branch checked out). Note in output: "Resuming existing worktree for Task N at worktrees/<sanitized_feature_branch>-task-N/". Dispatch the sub-agent to the existing worktree.
     - If the worktree/branch exists but the task status is NOT `In Progress` (e.g., `Not Started`): delete the stale worktree and branch first (`git worktree remove --force worktrees/<sanitized_feature_branch>-task-N && git branch -D <sanitized_feature_branch>-task-N`), then create fresh.
     - If the worktree/branch does NOT exist: proceed with normal creation below.

     **Worktree Creation (per task)**:
      - For each incomplete task in the work unit, create a git worktree under `worktrees/<sanitized_feature_branch>-task-N/` (where N is the task number):
        ```bash
        git worktree add worktrees/<sanitized_feature_branch>-task-N -b <sanitized_feature_branch>-task-N
        ```
      - The worktree branch is derived from the sanitized feature branch (e.g., if `<feature_branch>` is `story/005`, the sanitized form is `story-005`, and the worktree branch for Task 3 is `story-005-task-3`).
       - **Environment setup in worktree**: If `.kilo/setup-script.sh` was detected in Operation 7a, run it inside each newly-created worktree to install build dependencies (Gradle wrapper, Go modules, etc.):
         ```bash
         bash ../../.kilo/setup-script.sh
         ```
        using `workdir` set to `worktrees/<sanitized_feature_branch>-task-N`. If the setup script fails, report the error to the user, delete the worktree (`git worktree remove --force worktrees/<sanitized_feature_branch>-task-N && git branch -D <sanitized_feature_branch>-task-N`), mark the task as `Blocked` in the story file with the reason "Setup script failed in worktree", and do NOT dispatch the sub-agent. The task cannot proceed without a working build environment.

     **Sub-agent dispatch (per task)**:
     - Based on task type and size, route each task to the appropriate execution agent using a NEW AGENT TASK:
       - For **small coding tasks**, route to small-code-for-story-implementor
       - For **medium or larger coding tasks**, route to code-for-story-implementor
       - For **writing tasks**, route to technical-writer-for-story-implementor
       - **big-code-for-story-implementor** is never used for initial delegation — it is only reached via escalation (see Operation 13a).

      - **Worktree file permissions**: The sub-agent is working in a git worktree located at `worktrees/<sanitized_feature_branch>-task-N`. This directory IS part of the project's directory tree — it is a complete, writable copy of the repository. The sub-agent has FULL permission to read, write, edit, list, search, and otherwise manipulate files and directories within the worktree WITHOUT asking for additional permission, provided the sub-agent has equivalent permission on files in the main project directory. No additional authorization is needed to work within the worktree.

       - Each sub-agent MUST be instructed to use `workdir` parameter set to the worktree directory (`worktrees/<sanitized_feature_branch>-task-N`) for ALL bash commands. The story_context YAML block MUST include `worktree_path` and `source_root`:
       ```yaml
       story_context:
         story_file: "<story_file_name>"
         current_task: "<task_number_and_description>"
         activity_type: "<code|technical_writing>"
         worktree_path: "worktrees/<sanitized_feature_branch>-task-N"
         source_root: "<source_root>"
       ```

     **Delegation patterns**:
     - **Single task (sequential)**: Create one worktree, dispatch one sub-agent, await result.
     - **Parallel group (multiple tasks)**: Create one worktree per task simultaneously, then launch ALL sub-agents simultaneously as concurrent NEW AGENT TASK invocations in a single message. Each task is dispatched to its size-appropriate agent. Mixing agent types within a parallel group is acceptable and expected.

     **ACCOUNTABILITY**: After delegating, explicitly state which agent(s) you dispatched, the task(s) assigned, their worktree paths, and confirm: "I am NOT implementing these tasks. Awaiting sub-agent results."

     **STATUS UPDATE**: Before dispatching the sub-agent(s), update the story file to mark the delegated task(s) as `In Progress`. This is the first status transition and must be recorded immediately so that resumption (Operation 26) and parallel group tracking remain accurate.

9. **story-implementor**: Between EACH task (or parallel group), the agent MUST:
     - **Checkpoint**: Commit the current state with a descriptive message.
     - **Approval**: Determine whether to pause based on `checkpoint_mode`:
       - `"every"`: Present the work summary to the user and WAIT for explicit approval before proceeding.
       - `"none"`: Present a brief status update (task completed, test results) and automatically continue to the next work unit.
       - A task number: Present a brief status update and automatically continue. When the specified task number completes, present the work summary and WAIT for explicit approval. After approval, ask whether to continue with a new checkpoint target or switch to `"none"`.
     - Do not automatically continue to the next task without resolving the checkpoint_mode condition.

10. **story-implementor → code-for-story-implementor / small-code-for-story-implementor** (NEW AGENT TASK): For **coding tasks**, create a new agent task to switch to the appropriate code execution agent based on task size. Pass the ENTIRE task (all phases, subtasks, and verification steps). Pass the following context:
     - **Story Context** (MUST be returned unchanged in the results):
       ```yaml
        story_context:
          story_file: "<story_file_name>"
          current_task: "<task_number_and_description>"
          activity_type: "code"
          worktree_path: "worktrees/<sanitized_feature_branch>-task-N"
          source_root: "<source_root>"
        ```
     - The full task details from the story file (all phases and subtasks within the task)
     - Any relevant file paths or references mentioned in the task
      - **Worktree instruction**: The sub-agent MUST use `workdir` parameter set to `worktree_path` (from story_context) for ALL bash commands. The sub-agent is working in an isolated git worktree. Files inside the worktree directory ARE part of the project's directory tree — do NOT ask for permission to read, write, edit, list, or search files within the worktree.
       - **Source path resolution**: All file paths in story tasks are relative to `source_root` (from story_context). Resolve file paths within the worktree as `<source_root>/<story_path>`. Commands that involve building or testing should run from `source_root`.
       - **Commit requirement**: The sub-agent MUST commit all changes within the worktree before returning results. The merge step (Operation 27) only picks up committed changes; uncommitted changes will cause `git worktree remove` to fail. Use a descriptive commit message referencing the task and story.
      - **Routing logic**:
       - For **small tasks**: Use the small-code-for-story-implementor agent. Instruct it to make the change directly, keeping modifications minimal and focused. Verify existing tests still pass. The full Logical TDD Lifecycle is not required for small tasks.
       - For **medium or larger tasks**: Use the code-for-story-implementor agent. For new development, instruct the agent to follow the Logical TDD Lifecycle internally. For refactoring, instruct the agent that existing tests must continue to pass.
      - **Final build verification** (coding tasks): After completing all implementation, the sub-agent MUST run a final build and test pass within its worktree before returning. Use the project configuration from Operation 4b:
        - Command: `cd $source_root && $test_command`
        - Fix any errors found before returning.
      - Instructions to complete ALL subtasks within the task and return results
     - **CRITICAL**: Instruct the code execution agent to include the story_context section in their results exactly as provided, without modification

11. **code-for-story-implementor**: Execute the coding task according to the provided instructions and project coding standards, following the Logical TDD Lifecycle phases.

12. **small-code-for-story-implementor**: Execute the small coding task directly, making focused modifications. Keep changes minimal and well-scoped. Verify existing tests still pass. The full Logical TDD Lifecycle is not required for small tasks, but follow project coding standards and conventions.

13. **code-for-story-implementor / small-code-for-story-implementor → story-implementor** (NEW AGENT TASK): After completing the current task, create a new agent task to pass control back to the story-implementor with:
     - **Story Context** (returned unchanged from the original request):
       ```yaml
       story_context:
         story_file: "<story_file_name>"
         current_task: "<task_number_and_description>"
         activity_type: "code"
         worktree_path: "worktrees/<sanitized_feature_branch>-task-N"
         source_root: "<source_root>"
       ```
    - Summary of what was implemented
    - List of files created or updated
    - Test results (pass/fail status)
     - Any issues encountered during implementation
     - Confirmation of task completion status
     - **Status reporting**: The execution agent SHOULD report the final status of the task (`Completed`, `Failed`, `Blocked`, etc.) in its results so the story-implementor can update the story file promptly.

13a. **story-implementor**: Evaluate the execution agent's results against the task's success criteria. Each agent in the three-level hierarchy gets **one attempt** per task:
      - **Did the agent report `Completed` status?**
      - **Do the tests that the agent was instructed to run actually pass** when verified by the story-implementor?
      - **Is all required functionality implemented** (no missing functions, partial implementations, or build errors)?
      - **Was the work returned without obvious errors?**

      If ANY criteria are not met, the agent has failed. The story-implementor MUST escalate to the next level.

      **Special case — `small-code-for-story-implementor` operation-limit retry**: If `small-code-for-story-implementor` reports that it was otherwise successful but hit its operation limit before completing all work, the task may be re-delegated to `small-code-for-story-implementor` (same agent, same worktree, without escalation). Pass the current state back with instructions to complete the remaining work. This retry is permitted up to **3 times**. If the task is still incomplete after 3 such retries, treat it as a failure and escalate to `code-for-story-implementor` via the ladder below. This rule does NOT apply to `code-for-story-implementor` or `big-code-for-story-implementor` — those agents have different limits and do not hit operation-limit truncation in the same way.

      **Worktree cleanup on failure**: The failed sub-agent's changes are isolated in the worktree and have not been merged into `<feature_branch>`. Before escalating, delete the failed task's worktree:
      ```bash
      git worktree remove --force worktrees/<sanitized_feature_branch>-task-N
      git branch -D <sanitized_feature_branch>-task-N
      ```

      **Escalation ladder**:
      1. **`small-code-for-story-implementor` failed → escalate to `code-for-story-implementor`**:
         - Delete the worktree (see above).
         - Reclassify the task as medium: remove the `[Small]` annotation from the task in the story file. The task is larger or more difficult than originally estimated.
         - Re-create the worktree from `<feature_branch>`: `git worktree add worktrees/<sanitized_feature_branch>-task-N -b <sanitized_feature_branch>-task-N` (the branch was deleted, so a fresh branch with the same name is created).
         - Redelegate to `code-for-story-implementor` (via NEW AGENT TASK with `workdir` set to the re-created worktree path) with a retry description that notes the task was previously attempted by `small-code-for-story-implementor` and failed. The agent should take extra care, ask questions about anything unclear, and surface issues about the specified approach.

      2. **`code-for-story-implementor` failed → escalate to `big-code-for-story-implementor`**:
         - Delete the worktree (see above).
         - Re-create the worktree from `<feature_branch>`.
         - Redelegate to `big-code-for-story-implementor` with a task description that explicitly includes:
           - The full task requirements.
           - Which agent(s) previously attempted the task and why they failed (build errors, test failures, incomplete work, etc.).
           - A directive: **"If you are unable to complete this task, return a detailed diagnosis explaining why the task is inherently difficult or poorly specified rather than producing broken code. The diagnosis should identify the root causes, not just symptoms."**

      3. **`big-code-for-story-implementor` failed**:
         - Delete the worktree (see above). Do NOT re-create.
         - Stop work immediately. Report to the user:
           - The task's goal and requirements.
           - The full failure chain (small-code → code-for → big-code, with reasons for each failure).
           - The diagnosis from `big-code-for-story-implementor` if one was provided.
           - Recommendations for how to proceed (e.g., break the task into smaller subtasks, revisit requirements, seek architectural guidance).
         - Wait for explicit user direction before continuing.

      After escalation, proceed to review (Operation 18) once the escalated agent completes.

14. **story-implementor → technical-writer-for-story-implementor** (NEW AGENT TASK): For **writing tasks**, create a new agent task to switch to the technical-writer-for-story-implementor agent. Pass the ENTIRE task. Pass the following context:
    - **Story Context** (MUST be returned unchanged in the results):
       ```yaml
       story_context:
         story_file: "<story_file_name>"
         current_task: "<task_number_and_description>"
         activity_type: "technical_writing"
         worktree_path: "worktrees/<sanitized_feature_branch>-task-N"
         source_root: "<source_root>"
       ```
    - The full task details from the story file
    - Instructions to execute the task following the writing advice in `resources/writing_resources/writing_advice.md`
    - Any relevant file paths or references mentioned in the task
    - The story file name for reference
     - **Worktree instruction**: The sub-agent MUST use `workdir` parameter set to `worktree_path` (from story_context) for ALL bash commands and file operations. Files inside the worktree directory ARE part of the project's directory tree — do NOT ask for permission to read, write, edit, list, or search files within the worktree.
      - **Source path resolution**: All file paths in story tasks are relative to `source_root` (from story_context). Resolve file paths within the worktree as `<source_root>/<story_path>`.
     - **Commit requirement**: The sub-agent MUST commit all changes within the worktree before returning results. Use a descriptive commit message referencing the task and story.
    - **CRITICAL**: Instruct the technical-writer-for-story-implementor to include the story_context section in their results exactly as provided, without modification

15. **technical-writer-for-story-implementor**: Execute the writing task according to the provided instructions and project writing standards.

16. **technical-writer-for-story-implementor → story-implementor** (NEW AGENT TASK): After completing the current single element, create a new agent task to pass control back to the story-implementor with:
     - **Story Context** (returned unchanged from the original request):
       ```yaml
       story_context:
         story_file: "<story_file_name>"
         current_task: "<task_number_and_description>"
         activity_type: "technical_writing"
         worktree_path: "worktrees/<sanitized_feature_branch>-task-N"
         source_root: "<source_root>"
       ```
     - Summary of what was written/modified
     - List of files created or updated
     - Any issues encountered during writing
     - Confirmation of task completion status

17. **story-implementor**: Wait for the execution agent to complete the task and return results via the NEW AGENT TASK handoff.

18. **story-implementor → code-reviewer-for-story-implementor/technical-editor-for-story-implementor** (NEW AGENT TASK): Once the current work unit is complete, invoke the appropriate review agent(s):
     - **Single task**: Create one review handoff as currently defined.
     - **Parallel group**: Create one review handoff PER task in the group, launching all reviews simultaneously in a single message. Each review is for the specific task's changes.
     - **Story Context** (MUST be returned unchanged in the results):
       ```yaml
       story_context:
         story_file: "<story_file_name>"
         current_task: "<task_number_and_description>"
         activity_type: "<code_review|technical_editing>"
         worktree_path: "worktrees/<sanitized_feature_branch>-task-N"
         source_root: "<source_root>"
       ```
     - For **coding tasks**:
       - Pass control to the code-reviewer-for-story-implementor agent
       - Include the task description and scope
       - Provide instructions to review for code quality and adherence to project ontology
       - Include file paths of modified/created files from the worktree
     - For **writing tasks**:
       - Pass control to the technical-editor-for-story-implementor agent
       - Include the written content or document paths
       - Provide instructions to review for clarity and adherence to guidelines
     - **CRITICAL**: Instruct the code-reviewer-for-story-implementor or technical-editor-for-story-implementor to include the story_context section in their results exactly as provided, without modification

19. **code-reviewer-for-story-implementor/technical-editor-for-story-implementor**: Execute the review according to the provided instructions:
     - For **code-reviewer-for-story-implementor**: Review code quality, patterns adherence, and test coverage
     - For **technical-editor-for-story-implementor**: Review document clarity, grammar, and style guidelines

20. **code-reviewer-for-story-implementor/technical-editor-for-story-implementor → story-implementor** (NEW AGENT TASK): After completing the review, create a new agent task to pass control back to the story-implementor with:
     - **Story Context** (returned unchanged from the original request):
       ```yaml
       story_context:
         story_file: "<story_file_name>"
         current_task: "<task_number_and_description>"
         activity_type: "<code_review|technical_editing>"
         worktree_path: "worktrees/<sanitized_feature_branch>-task-N"
         source_root: "<source_root>"
       ```
     - Review results (pass/fail with issues)
     - List of issues found (if any)
     - Specific recommendations for fixes
     - Severity of each issue (blocking vs. minor)

21. **story-implementor**: Process the review results. You process review feedback and route it — you do NOT apply fixes yourself:
      - Upon receiving results from any lower-level agent, **first extract and validate the story_context** to confirm:
        - The story file being processed
        - The current task and its position in the story workflow
        - The type of activity that was performed (coding, writing, review, or editing)
        - The worktree path for the task
      - Use the story_context to determine the next appropriate step in the workflow

      **Evaluate reviewer findings**:
      - If the review agent reports no issues (review passes), proceed to merge the worktree (see Operation 27).
      - If the review agent identifies issues, categorize them for handling:
        - **Preexisting issues**: Problems in the codebase that predate this task. These are informational only — do NOT hold them against the task implementation and do NOT delegate them for fixing within this task.
        - **Task-introduced issues**: All issues introduced by or related to this task's implementation. This includes blocking issues (bugs, missing functionality, incorrect logic, test failures, broken build), minor issues (style, naming, documentation gaps, code clarity, adherence to conventions), and DRY refactoring recommendations (code duplication, structural improvements).

      **Handle task-introduced issues**:
      - ALL task-introduced issues are actionable and MUST be addressed. Do not filter, skip, or ignore issues based on severity. Every issue the reviewer reports for this task is passed to the fix delegation.
      - **First, evaluate whether escalation is needed**: Determine whether the issues indicate poor implementation quality. Consider:
        - Are there blocking issues (broken build, test failures, incorrect logic, missing functionality)?
        - Are there multiple issues that collectively suggest the implementation is not solid?
        - Are there structural problems (missing tests, incorrect logic, broken build)?
      - **If issues DO indicate poor quality**: The sub-agent's changes are isolated in the worktree and have NOT been merged into `<feature_branch>`. Trigger the three-level escalation ladder:
        1. **Determine current level**: Identify which agent completed the implementation.
        2. **Escalate one level up** using the same rules as Operation 13a:
           - `small-code-for-story-implementor` → escalate to `code-for-story-implementor`
           - `code-for-story-implementor` → escalate to `big-code-for-story-implementor`
           - `big-code-for-story-implementor` → stop, report to user with full failure chain including the reviewer's feedback and the big-code diagnosis
        3. Delete the failed task's worktree and re-create fresh from `<feature_branch>` BEFORE escalating:
           ```bash
           git worktree remove --force worktrees/<sanitized_feature_branch>-task-N && git branch -D <sanitized_feature_branch>-task-N
           git worktree add worktrees/<sanitized_feature_branch>-task-N -b <sanitized_feature_branch>-task-N
           ```
        4. After the escalated agent completes, return to review (Operation 18).
      - **If issues DO NOT indicate poor quality** (the implementation is fundamentally sound but needs adjustments): ALL task-introduced issues — blocking, minor, and DRY recommendations alike — MUST be addressed. Do NOT filter or skip any reviewer comment.
        1. DELEGATE the fix back to the ORIGINAL EXECUTION AGENT by creating a NEW AGENT TASK with `workdir` set to the same worktree path. The sub-agent's changes are still in the worktree (unmerged), so it can fix in place. Pass ALL review feedback (every issue the reviewer reported for this task) as part of the task description. Instruct the agent to address every issue. Do NOT edit files yourself to fix issues — even small, obvious fixes.
        2. Limit to `review_iteration_limit` total iterations (implementation + review + fix) per task, as configured in Operation 4b.
        3. If issues persist after `review_iteration_limit` iterations: STOP the fix loop. Inform the user with a summary of the unresolved issues, what was attempted, and ask for guidance on how to proceed. Do NOT silently escalate.
       **SELF-CHECK**: Before proceeding to the next task or revision, confirm: "I am not applying these fixes myself. I am delegating the revision to [agent name]."

22. **story-implementor**: Post-Work — after sub-agents return results and pass review (Operation 21), perform these coordination steps:

      - Verify the story_context returned by each sub-agent includes the correct `worktree_path`.
      - For each completed task that passes review, initiate merge via Operation 27 (Worktree Merge Protocol) — one at a time per the merge serialization rule. Merge order is determined by the reintegration instructions for the group; see Operation 27 for the merge order resolution rule.
      - For each task that does NOT pass review, do NOT merge. Delete its worktree before escalating (see Operations 13a and 28).
      - For parallel groups, present a combined summary of all completed tasks and their individual results.
       - Run integration verification per Operation 27 (build verification for sequential tasks; integration test suite for parallel groups).

     **Blocked/Skipped tasks in work unit**: After merging all successful tasks and running integration verification, if any tasks in the work unit were `Blocked` or `Skipped`:
      - Present the user with a summary of which tasks were blocked or skipped and the reason for each (setup script failure, user skip directive, etc.).
      - Ask the user for guidance on how to proceed (e.g., retry the blocked tasks after fixing the underlying issue, skip them permanently and continue, abort the story).
      - Do NOT proceed to the next work unit until the user provides direction.

     **Note**: Marking tasks as "Completed" in the story file is handled by Operation 23, not here.

23. **story-implementor**: Once the current task is complete and reviewed, update the story file to mark the task as "Completed" per the Task Status Update Protocol. Do this immediately upon verification; do not defer the update to a later checkpoint.

24. **story-implementor**: After completing a work unit (single task or parallel group) and updating the story file, determine the appropriate next action:
     - If `checkpoint_mode` is `"every"`:
       - **Pause execution** and present a summary to the user that includes:
         - The task(s) that were just completed (for parallel groups, list all tasks and their individual results)
         - A list of all changes made (files created, modified, or deleted)
         - The review result (pass/fail) and any notable findings from the review
         - Any decisions made or clarifications obtained during task execution
         - Integration test results (for parallel groups)
       - **Wait for explicit user direction** before proceeding to the next work unit.
       - Do not automatically continue to the next work unit without user approval.
     - If `checkpoint_mode` is `"none"`:
       - Present a concise status update: tasks completed, test results, brief summary.
       - Automatically proceed to the next work unit.
     - If `checkpoint_mode` is a task number:
       - Present a concise status update and automatically proceed.
       - When the specified task number completes, pause with a full summary and wait for approval.
       - After approval, ask the user for the next checkpoint target (new task number, `"none"`, or `"every"`).

25. **story-implementor**: Final Story Completion - If no more tasks remain:
      - Ensure the walkthrough is saved adjacent to the story file.
      - Append the post-mortem to the walkthrough.
      - **Clean up remaining worktrees**: List any remaining worktrees under `worktrees/`:
        ```bash
        git worktree list | grep "<sanitized_feature_branch>-"
        ```
        For each remaining worktree, delete it and its branch:
        ```bash
        git worktree remove --force worktrees/<sanitized_feature_branch>-task-N
        git branch -D <sanitized_feature_branch>-task-N
        ```
        Verify no worktree directories matching `worktrees/<sanitized_feature_branch>-*` remain (or note orphaned files for manual cleanup). If worktrees remain that cannot be deleted, note them and inform the user.
      - Commit the final product and push the branch to origin.
      - Update the story entry in `toc.md` by changing the inline status metadata on the title line to `_Done_`.
      - Ask the user what, if anything, from the post-mortem should be added to the lessons learned section of the project.

26. **story-implementor**: Continue to the next work unit. After receiving checkpoint approval from the user (or after auto-continuing under `"none"` checkpoint mode), return to Operation 5 to identify and process the next incomplete work unit. The story-implementor continues within the same session — there is no context reset, no subagent launch, and no handoff between story-implementor instances. This is simply the natural loop back to the top of the work-unit processing cycle.

27. **story-implementor** (Worktree Merge Protocol): After a task passes review (Operation 21 confirms no issues), merge the worktree branch into `<feature_branch>`:

     **Merge serialization rule**: Merges happen ONE AT A TIME. A task is eligible for merge as soon as it **completes + passes review** — it is NOT necessary to wait for all parallel tasks to finish before merging begins. Merging can begin while other tasks are still in progress or under review. Only one `git merge` may be in progress at any time.

      **Merge order resolution**: When the story's "Reintegration" subsection specifies a merge order for the current group (e.g., "Task 3 first, then Task 4"), the story-implementor MUST follow the specified order — even if a later-ordered task completes and passes review before an earlier-ordered one. Delay the merge of later-ordered tasks until earlier-ordered tasks have been merged. When the reintegration instructions explicitly say "any order" for a group, use **finish-order**: the first task to reach "review passed → ready to merge" is merged first. If multiple tasks reach this state simultaneously (e.g., completing in the same batch), use task number order as a stable tiebreaker. When no "Reintegration" subsection exists for a group, or the Reintegration subsection says "None" for the group (including a global "None — no parallel groups"), also use finish-order. The execution graph ensures predecessors complete before successors start, so any order within a parallel group is structurally valid.

     **Pre-merge commit check**: Before merging, verify the worktree branch has at least one commit beyond `<feature_branch>`:
     ```bash
     git log <feature_branch>..<sanitized_feature_branch>-task-N --oneline
     ```
     If the branch has no new commits (the sub-agent produced no changes or forgot to commit), this is an error. Treat as a sub-agent failure: report to user, delete the worktree (see Operation 28), and do NOT mark the task as Completed.

     **Merge procedure**:
     1. Ensure you are on `<feature_branch>`:
        ```bash
        git checkout <feature_branch>
        ```
     2. Merge the worktree branch:
        ```bash
        git merge <sanitized_feature_branch>-task-N
        ```
        For sequential tasks (only one active worktree), this is always a fast-forward merge. For parallel groups where a prior worktree was already merged, this may be a non-fast-forward merge if the tasks touched different files — git handles this automatically.
     3. If the merge produces conflicts:
        - Pause immediately. Present the conflict details (conflicting files, conflict markers) to the user.
        - Do NOT auto-resolve conflicts. Options: user resolves manually, or abort merge with `git merge --abort` and handle the worktree separately.
     4. After successful merge, delete the worktree directory and its branch:
        ```bash
        git worktree remove worktrees/<sanitized_feature_branch>-task-N
        git branch -d <sanitized_feature_branch>-task-N
        ```
        If `git branch -d` fails because the branch was already merged, use `git branch -D <sanitized_feature_branch>-task-N`.
     5. Verify the `worktrees/<sanitized_feature_branch>-task-N/` directory is removed. If not (e.g., uncommitted files), use `git worktree remove --force worktrees/<sanitized_feature_branch>-task-N` and note orphaned files for manual cleanup.

      **Integration verification**: 
      - **For sequential tasks**: Run a build verification on `<feature_branch>` after merge to catch stale-base compilation errors. Use the project configuration from Operation 4b: `cd $source_root && $build_command`. (The sub-agent already ran the full test suite in the worktree, so a lightweight build check is sufficient.)
      - **For parallel groups**: Run integration tests after ALL worktrees in the group have been merged. Use the parsed `reintegrationInstructions` (from Operation 4a) for the current group — it provides integration test commands. If no reintegration instructions exist for this group (or the story predates this feature), run the project's full test suite: `cd $source_root && $test_command`.
      - For parallel groups, do NOT run the full integration suite after each individual worktree merge — only after the final worktree in the group is merged.

28. **story-implementor** (Worktree Failure Cleanup): If a task fails for any reason OTHER than the three-level escalation ladder (i.e., the sub-agent errors, times out, or reports a non-recoverable failure before review), or if the user chooses to skip/abort:

     **Cleanup procedure**:
     1. Delete the worktree for the failed task. The worktree's changes have NOT been merged into `<feature_branch>`, so deletion is safe:
        ```bash
        git worktree remove --force worktrees/<sanitized_feature_branch>-task-N
        git branch -D <sanitized_feature_branch>-task-N
        ```
     2. Verify `worktrees/<sanitized_feature_branch>-task-N/` is removed.

     **For timeout or sub-agent error** (not an escalation scenario):
     - Present the failure to the user with:
       - The failing task number and description
       - The worktree branch name that was deleted
       - The error output or last known state
       - Options: retry (re-create the worktree from `<feature_branch>` and re-dispatch to the same agent type per the worktree creation protocol in Operation 8), skip (mark task as `Skipped` in the story file with reason), abort story
     - Wait for explicit user direction before proceeding.

     NOTE: The three-level escalation ladder (small-code→code-for→big-code) is handled by Operations 13a and 21, which include their own worktree deletion and re-creation steps. Operation 28 is for non-escalation failure scenarios only.

## Output

The script outputs:

- Confirmation of which story file is being processed
- The checkpoint mode in effect
- The current work unit being executed (single task or parallel group with task list)
- Worktree paths and branch names created for each task (e.g., `worktrees/<sanitized_feature_branch>-task-3/` on branch `story-005-task-3`)
- Setup script execution status per worktree (if `.kilo/setup-script.sh` is present; otherwise a warning if absent)
- Task type and size classification for each task
- Progress updates as each work unit is completed and reviewed
- For each completed task: merge confirmation, worktree deletion confirmation
- For parallel groups: combined summary showing all tasks, their individual results, integration test results, and merge order followed
- Reintegration integration test results (using the project test command from Operation 4b: `cd $source_root && $test_command` as the default when no reintegration command is specified)
- Summary of changes made for each task
- Review results (pass or issues found)
- The number of review iterations for each task
- Git merge confirmations with commit hashes
- Escalation notifications: which agent failed, which agent was escalated to, worktree deletion and re-creation steps
- big-code-for-story-implementor diagnosis (if it fails)
- Final completion status when all tasks are done
- Error messages if maximum review iterations are exceeded (per `review_iteration_limit` from Operation 4b)
- Error messages with full failure chain if all three agent levels are exhausted

## Error Handling

If required parameters aren't specified, the user should be prompted. The operations in the script WILL NOT be executed until/unless all required parameters have been specified.

During execution, if an error is encountered, the execution of the script will be paused and the user will be asked if script execution should continue or if the script should be terminated.

Specific error conditions handled:

- Story file not found or unreadable
- Story file format is invalid (missing Tasks section)
- Unable to determine task type (ambiguous content)
- Execution agent fails to complete the task
- Review agent encounters an error
- Maximum review iterations exceeded (configured via `review_iteration_limit` in Operation 4b, default 3)
- Git merge failures (merge conflicts during worktree reintegration)
- Git commit failures within worktrees (sub-agent responsibility)
- File write operations fail
- Story toc.md update failures
- Small-code-for-story-implementor task failure (escalates to code-for-story-implementor; see Operation 13a)
- code-for-story-implementor task failure (escalates to big-code-for-story-implementor; see Operation 13a)
- big-code-for-story-implementor task failure (stops and consults user with diagnosis; see Operation 13a)
- Reviewer finds poor quality implementation (escalates to next level of agent hierarchy; see Operation 21)
- Worktree creation fails (git worktree add error, branch already exists, disk full)
- Worktree merge produces conflicts (presented to user, not auto-resolved)
- Worktree deletion fails (orphaned directories, branch delete error, locked worktree)
- `worktrees/` directory creation fails (disk full, permissions)
- Worktree branch naming collision with existing local branches
- `.kilo/setup-script.sh` execution fails within a worktree (build dependencies not installable — task cannot proceed)

## Notes

- The script processes work units sequentially within a single session, looping from Operation 5 through Operation 26 for each work unit until all tasks are complete.
- **Context Size Management**: The primary goal of the handoff protocol is to keep implementation details (large file contents, code snippets) out of the story-implementor's context window. The story-implementor coordinates at the task level while sub-agents handle the full file content needed for implementation.
- **One Task Per Handoff Rule**: The story-implementor MUST NOT further decompose tasks into sub-elements (phases, subtasks, steps) for separate handoffs. Doing so would multiply round-trips without reducing context load, since the same files would be re-read in each sub-handoff. Each task is handed off as a single unit. The EXCEPTION is parallel groups: multiple complete tasks can be launched simultaneously as separate handoffs.
- **Parallel Execution**: When a story declares parallel groups in its "Parallel Execution" section, the story-implementor creates one git worktree per incomplete task in the group (under `worktrees/<sanitized_feature_branch>-task-N/`) and launches all sub-agents concurrently. Each task is still a single-unit handoff. After individual tasks are reviewed and their worktrees are merged (one at a time), the story-implementor runs integration tests before proceeding.
- **Mandatory Worktree Isolation**: EVERY task — sequential or parallel — is executed in an isolated git worktree under `worktrees/<sanitized_feature_branch>-task-N/`. `worktrees/` is a real directory at the project root (never a symlink into a dot-directory). Worktrees are created from `<feature_branch>` using `git worktree add`, sub-agents work in the worktree via the `workdir` parameter, and only successful, reviewed work is merged back into `<feature_branch>`. Failed work is deleted with the worktree and never contaminates `<feature_branch>`. This eliminates the need for stash, reset, or revert operations entirely.
- **No Dot-Directories**: All worktree paths must resolve to `worktrees/` (no `.kilo/` prefix). This requirement exists because sub-agents are blocked from accessing dot-directories (`.kilo/`, `.git/`, etc.). The `worktrees/` directory at the project root is a real directory — never a symlink into `.kilo/`. This ensures `git worktree list` output and all git commands report clean paths like `/project/worktrees/...`.
- **Worktree Environment Setup**: If `.kilo/setup-script.sh` exists in the repository root, the story-implementor runs it inside each newly-created worktree to install build dependencies (Gradle wrapper, Go modules, etc.). If absent, the story-implementor warns the user once at the start of the story and proceeds — the user may have already prepared their environment manually or the worktree may not need additional setup.
- **Worktree Merge Serialization**: Merges happen one at a time. For parallel groups, a task is eligible for merge as soon as it completes + passes review — it is NOT necessary to wait for all parallel tasks to complete. Merge order follows the reintegration instructions from the story's "Reintegration" subsection when specified; when reintegration says "any order" or "None", or is absent, use finish-order (first task reviewed-ok first, task-number tiebreaker). Sequential tasks merge trivially (fast-forward) since no other worktree is active concurrently.
- **Task Size Determination**: Tasks annotated with `[Small]` on their status line are routed to small-code-for-story-implementor. Unannotated tasks fall back to heuristic analysis. When in doubt, default to medium (code-for-story-implementor).
- **Checkpoint Mode**: The story-implementor confirms all parameters with the user at the start of the story session in Operation 4b, before any task analysis or execution begins. No further user queries for parameters occur during the session.
- **Refactoring vs New Development**: For medium or larger new development tasks, the code-for-story-implementor agent internally follows the Logical TDD Lifecycle. For medium or larger refactoring tasks (no behavior change, no new tests), the code-for-story-implementor agent performs the file-level reorganizations directly and verifies existing tests still pass. For small tasks, the small-code-for-story-implementor agent makes focused modifications directly without requiring the full TDD Lifecycle, while still verifying existing tests pass.
- **Worktree Cleanup**: After successful merge, the story-implementor removes the worktree directory (`git worktree remove`) and deletes the worktree branch (`git branch -d`). On failure, the worktree is force-removed (`git worktree remove --force`) and the branch is force-deleted (`git branch -D`). Orphaned directories under `worktrees/` should be noted for manual cleanup.
- The distinction between coding and writing tasks is heuristic-based and may require manual override in ambiguous cases.
- Review iterations are capped at `review_iteration_limit` (configured in Operation 4b, default 3) to prevent infinite loops between implementor and reviewer agents. If issues persist after the limit, the story-implementor informs the user and asks for guidance.
- The script assumes the presence of small-code-for-story-implementor, code-for-story-implementor, big-code-for-story-implementor, code-reviewer-for-story-implementor, technical-writer-for-story-implementor, and technical-editor-for-story-implementor agents.
- **CRITICAL**: All handoffs between agents MUST be done via **NEW AGENT TASKS** to ensure proper context isolation and role clarity.
- **Story Context**: Every handoff to a lower-level agent includes a `story_context` section that contains the story file name, current task, activity type, and worktree path. Lower-level agents MUST return this context unchanged in their results. The story-implementor uses this context to determine the next steps in the workflow without requiring the user to manually specify what to do next.
- **Sub-agent Worktree Directive**: All sub-agents receive a `worktree_path` in their `story_context` and MUST use the `workdir` parameter set to that path for all bash commands. The sub-agent is working in an isolated git worktree and commits changes there. The story-implementor handles merging the worktree branch back to `<feature_branch>` after review passes.
- **Three-Level Agent Hierarchy**: Coding tasks follow a three-level escalation ladder based on agent capability. Each agent gets exactly one attempt per task. On failure, the worktree is deleted and re-created fresh from `<feature_branch>` before the escalated agent begins:
  1. **small-code-for-story-implementor** — for tasks annotated `[Small]` (at most 2 files, narrow scope, straightforward). Never used for tasks that require the full Logical TDD Lifecycle or have ambiguity.
  2. **code-for-story-implementor** — for medium tasks. Follows the Logical TDD Lifecycle for new development. Handles refactoring tasks that span multiple files.
  3. **big-code-for-story-implementor** — emergency escalation only. Never used for initial delegation in planning. Reached when both smaller agents have failed. If big-code fails, it must return a diagnosis explaining why the task is inherently difficult rather than producing broken code.
- **Escalation Triggers**: Any failure that does not meet the task's success criteria triggers escalation — build failures, test failures, incomplete work, reviewer finding poor quality, or any outcome not matching the expected success criteria.
- **Reviewer Quality Evaluation**: Reviewers evaluate code for quality, not just correctness. Preexisting codebase issues are not held against the task. All task-introduced issues (blocking, minor, and DRY recommendations) are actionable and MUST be delegated back to the implementor for fixes. Poor quality findings additionally trigger escalation to the next level in the hierarchy.
- **Branch and Worktree Name Sanitization**: To avoid Git confusion with `/` characters in branch names, the `<feature_branch>` parameter is sanitized for use in worktree branch names and worktree directory names: all `/` characters are replaced with `-`. For example, if `<feature_branch>` is `story/005`, the sanitized form is `story-005`, and worktree branches are named `story-005-task-N` (not `story/005/task-N`), and worktree directories are named `worktrees/story-005-task-N/` (not `worktrees/story/005/task-N/`). The original unsanitized `<feature_branch>` is still used for git operations on the feature branch itself (e.g., `git checkout <feature_branch>`, `git merge` into `<feature_branch>`).

