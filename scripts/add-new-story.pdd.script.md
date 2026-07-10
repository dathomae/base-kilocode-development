# Add New Story

This script automates the process of adding a new story to the project. It generates the next available story number by scanning existing stories in both the active and finished stories directories, creates a new story file from the template, and adds an entry to the Active Stories section in toc.md.

This script is executed by the **technical-writer** agent, assisted by the **architect-for-story-planning** subagent. The technical-writer owns the story structure and prose; the architect-for-story-planning evaluates task decomposition (identifying tasks that are too large for a single sub-agent handoff), judges task sizing, and assesses parallelism opportunities based on knowledge of the codebase file structure and dependency graph. See the [stories template](chronocone/resources/templates/stories.md) for the full story format including Task Sizing, Parallel Execution, and Execution Order sections.

## Parameters

| Name         | Type   | Description                                                                                                                             | Required | Default Value             |
| ------------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------- | -------- | ------------------------- |
| story_name   | string | Short descriptive name for the story (e.g., "database-migration"). Must contain only alphanumeric characters, hyphens, and underscores. | Yes      | N/A                       |
| description  | string | Brief description of the story's goal (1-2 sentences)                                                                                   | Yes      | N/A                       |
| project_root | string | The root directory of the project where memory-bank/stories is located                                                                  | No       | Current working directory |

## Operations

1. Validate parameter constraints:
   - story_name must be provided and non-empty
   - story*name must match the pattern `^[a-zA-Z0-9*-]+$` (alphanumeric, hyphens, underscores only)
   - description must be provided and non-empty
   - description should not exceed 200 characters

2. Construct the stories directory path by combining project_root with `memory-bank/stories`.

3. Construct the finished-stories directory path by combining project_root with `memory-bank/stories/finished-stories`.

4. Verify that the `memory-bank/stories` directory exists. If it does not exist, pause execution and inform the user. Ask if the script should be terminated.

5. Scan for existing story files to determine the highest story number:
   - List all files in `memory-bank/stories/` matching the pattern "StoryNNN\_\*.md" where NNN is a three-digit number
   - List all files in `memory-bank/stories/finished-stories/` matching the same pattern
   - Extract the numeric portion (NNN) from each filename
   - Identify the maximum story number found
   - If no story files exist, use 0 as the maximum

6. Calculate the next story number:
   - next_number = max_story_number + 1
   - Format the number as a three-digit string with leading zeros (e.g., 1 becomes "001", 40 becomes "040")

7. Construct the new story filename:
   - Format: `Story<NNN>_<story_name>.md` (e.g., "Story040_database-migration.md")

8. Read the story template from `resources/templates/stories.md`:
   - Verify the template file exists. If not, pause execution and inform the user. Ask if the script should be terminated.
   - Read the template content

9. **technical-writer**: Create the new story file at `memory-bank/stories/<filename>`:
    - Start with the story header: `# Story<NNN>: <Human-Readable Story Name>`
    - Replace the template's "Story Template" title with the actual story title
    - Include the description in the Goal section
    - When generating the Tasks section, you **MUST** read and strictly adhere to the rules defined in the "Test-First Development" section of the template (i.e. applying the Logical TDD Lifecycle where tests are integrated into each task, rather than broken out as separate tasks).
     - Preserve all other template sections (References, Dependencies, Dependent Stories, Constraints, Intent, etc.)
     - **File path convention**: All file paths in task descriptions MUST be relative to the project's `source_root` (where build/test commands run from). For CPL stories, `source_root` is `src/chronocone_planning_language`, and Go packages live under `src/main/go/` within `source_root`. Therefore:
       - Go source file paths MUST include the `src/main/go/` prefix: e.g., `src/main/go/analysis/foo.go`, NOT `analysis/foo.go`
       - Go test commands MUST use fully-qualified package paths from `source_root`: e.g., `go test ./src/main/go/analysis/...`, NOT `go test ./analysis/...`
       - `go build` commands must similarly use `./src/main/go/...` prefix
     - Set all task statuses to "Not Started"
    - **Task Design: Prefer Small Tasks.** When decomposing story requirements into tasks, prefer creating straightforward, narrow-scope tasks that qualify as `[Small]` (at most 2 files, additive change, one concern), PROVIDED that the decomposition is natural and does not introduce additional complexity. Specifically:
      - **Prefer small when**: A piece of work naturally fits in 1-2 files, addresses a single concern, and follows an existing pattern (e.g., adding a switch case, adding a test for an existing function, a single-file helper with its test).
      - **Do NOT force smallness when**: The work would require creating artificial intermediate artifacts (temp modules, adapters, scaffolding files) just to keep each task under the 2-file threshold; OR splitting a coherent multi-file change (e.g., adding a struct to one file and its method to another) into separate tasks would create coordination overhead with no real isolation benefit; OR forcing small tasks would increase the number of parallel worktrees and merge steps with no corresponding isolation gain.
       - **Natural decomposition test**: If you can describe the task in one sentence without using the word "and" to join unrelated concerns, it is a good candidate for `[Small]`. If the description reads "do X AND update related thing Y in another file," those are naturally one task — do not split.
        - **Worktree consideration**: Since every task runs in an isolated git worktree, preferring small tasks can increase the number of worktree branches and merge operations. A single medium task spanning 3 files that are tightly coupled is often simpler to implement, review, and integrate than three small tasks that must be merged across three worktrees with ordering constraints. When the integration cost of multiple small tasks exceeds the benefit of small-code agent simplicity, keep the work together as a single medium task.

9a. **technical-writer** (Task Decomposition Review): Consult the **architect-for-story-planning** subagent to evaluate whether any tasks in the draft task list are too large for a single sub-agent handoff. Pass the architect:
      - The full task list with all subtask breakdowns (lettered steps a, b, c, ...)
      - The files each task is expected to touch (source + test)
      - Any intermediate verification steps within each task (compile checks or test passes that occur before the final acceptance criteria)

      The architect evaluates each task against the decomposition rules defined in [architect-for-story-planning's Task Decomposition Review section](chronocone/.kilo/agent/architect-for-story-planning.md) and returns a structured analysis:
      - For each task flagged as `too_large` or `borderline`: the architect **MUST propose the best possible alternate task breakdown**. This is not optional — the architect always attempts to find a decomposition that creates smaller, more focused tasks. Only when every potential split would violate the "do NOT decompose when" rules (creating artificial complexity, intermediate artifacts, or merge conflicts with no isolation benefit) does the architect state that the task is at its natural granularity and cannot be further decomposed.
      - If all tasks are appropriately sized: a confirmation that no changes are needed

      **Apply the decomposition**:
      - For each task where the architect proposes a decomposition: split the original task into the proposed sub-tasks. Assign each a task number with a letter suffix (e.g., Task 5 becomes Task 5a and Task 5b). Update all task references in the Execution Order graph to reflect the new task structure.
      - For each task where the architect states decomposition is not feasible (the task is at its natural granularity): keep as one task. Record the architect's analysis and justification in the task description so the story-implementor understands the sizing decision.

      **Important**: Always apply the architect's proposed decomposition when one is provided. The architect's "do NOT decompose when" rules exist to prevent artificial splits that increase worktree overhead and merge complexity with no isolation benefit.

9b. **technical-writer**: Classify each task by size using the rules of thumb in the stories template's Task Sizing section:
    - For each task, determine whether it meets the "Small" criteria (at most 2 files, additive change, narrow scope, STRAIGHTFORWARD — no ambiguity or complex multi-structure reasoning).
    - **Design intent check**: Tasks that were intentionally designed to be small (per step 9's "Prefer Small Tasks" principle) should naturally meet the straightforwardness test. If a task was designed small but upon classification review you realize it requires reasoning about multiple interacting structures, has hidden complexity, or would benefit from broader context — it is not small. Remove the `[Small]` annotation.
    - Annotate small tasks with `[Small]` on the status line (e.g., `### Task 3: ... — Not Started [Small]`)
    - For tasks where sizing is ambiguous, consult the **architect-for-story-planning** subagent. Pass it:
      - The task description and scope
      - The files the task would touch (source + test)
      - Whether the change introduces new structs, interfaces, or operation kinds
      - Whether the change touches multiple files with coordinated changes
    - Apply the **when in doubt** rule: default to not annotating `[Small]` (routes to `code-for-story-implementor`). The small-code-for-story-implementor agent has less capability to resolve ambiguity than code-for-story-implementor — a task that is questionable in complexity should go to the more capable agent. The step 9 design bias toward small tasks should never override this safety rule.

9c. **technical-writer**: Identify parallel execution opportunities and reintegration requirements:
     - For each pair (or larger group) of tasks, consult the **architect-for-story-planning** subagent with:
       - The list of files each task touches
       - Whether any task in the group depends on another's output
       - Whether any task modifies shared infrastructure files (build scripts, base structs, shared evaluators, CI config, `.gitignore`, Makefile, etc.)
       - For every parallel group: also ask for reintegration instructions (merge order, integration test targets, conflict-prone files)
     - The architect-for-story-planning evaluates:
       - File disjointness (are the file sets completely non-overlapping?)
       - Package boundaries (are they in different packages, reducing compile-time coupling?)
       - Semantic independence (can each task be tested and verified without the other's changes?)
       - **Reintegration** (for every parallel group: merge order, integration test targets, conflict-prone files — see architect-for-story-planning's Reintegration Analysis section). All tasks now run in mandatory git worktrees, so reintegration instructions apply to every parallel group.
     - Record parallel groups in the "Parallel Execution" subsection using the format from the template.
     - Build the execution order graph in the "Execution Order" subsection, showing both parallel branches (referencing groups) and sequential chains.
     - **When in doubt, default to serial.** False parallelism (tasks that appear independent but produce merge conflicts or compile errors) is significantly worse than missed parallelism.

9d. **technical-writer**: Add the Task Sizing subsection before the first task, using the format from the template. Document which tasks are small and (briefly) why, referencing the architect-for-story-planning's analysis.

9e. **technical-writer**: Add a Reintegration subsection for each parallel group, documenting how the story-implementor should merge worktree branches back and verify the result. Use the reintegration instructions from the architect-for-story-planning subagent:
     - For each parallel group:
       - **Merge order**: which worktree branch to merge first and why (e.g., "Task 3 first — modifies Makefile, then Task 4")
       - **Integration test targets**: the exact test command to run after all worktrees in the group are merged, with the reason
       - **Conflict-prone files**: files NOT modified by any task but at risk during merge (lock files, shared test files)
     - Format example:
       ```
       ### Reintegration
       
       **Group A (Tasks 3, 4):**
       - Merge order: Task 3 first (modifies Makefile — shared infrastructure), then Task 4
       - Integration tests: `go test ./pkg/evaluator/... -count=1 -timeout 60s` — verify combined evaluator behavior
       - Watch for conflicts in: `go.sum` (dependency resolution may differ per worktree)
       
       **Group B (Tasks 7, 8):**
       - Merge order: any order (fully disjoint packages, no shared infrastructure)
       - Integration tests: `go test ./pkg/parser/... ./pkg/formatter/... -count=1 -timeout 60s`
       - Watch for conflicts in: none expected
       ```
     - If no parallel groups exist, omit this subsection or write "None — no parallel groups."
      - **Branch reference**: The reintegration instructions should refer to the "story branch" generically (the branch the story-implementor selected at runtime). Do not hardcode a specific branch name — the story-implementor determines the branch at execution time via Operation 4b (User Configuration).

10. Read the current `memory-bank/stories/toc.md` file:
    - Verify the file exists and contains an "Active Stories" section. If not, pause execution and inform the user. Ask if the script should be terminated.

11. Add the new story entry to the Active Stories section in toc.md:
    - Insert a blank line before the new entry block
    - Insert the entry as a two-line block:
      - Line 1: `**[Story<NNN>_<story_name>.md](Story<NNN>_<story_name>.md)** — _Not Started_`
      - Line 2: `: <description>` (the provided description parameter)
    - Maintain proper list formatting
    - Insert the new entry in the appropriate position (maintain numeric order if stories are numbered sequentially, or append at the end)

12. Verify the new story file was created successfully:
    - Confirm the file exists at the expected path
    - Confirm the file is readable and contains the expected content

13. Verify the toc.md file was updated successfully:
    - Confirm the new story entry appears in the Active Stories section
    - Confirm the entry has the correct filename, description, and status

14. Verify the story's structural quality:
       - Confirm that the story references the "Test-First Development" or "Logical TDD Lifecycle" principles
       - Confirm that coding tasks include tests written before implementation
       - Confirm that the test-and-implement cycle is integrated within each task rather than broken out as separate tasks
       - **Task decomposition quality**: Confirm that no task has intermediate compile/test verification steps that signal a self-contained preamble was not split (i.e., no task should say "run go test to verify the new types pass" followed by "now update the consumers"). If such a pattern exists, re-consult the architect-for-story-planning — a decomposition opportunity was likely missed in step 9a.
       - **Task size signals**: Confirm that no task has 10+ lettered subtasks spanning multiple "modify → compile → test" cycles. If such a task exists and step 9a did not flag it, re-consult the architect-for-story-planning.
       - Confirm `[Small]` annotations are on tasks that meet the sizing rules of thumb (consult architect-for-story-planning if uncertain)
       - Confirm parallel groups list only tasks that are truly file-disjoint (consult architect-for-story-planning if uncertain)
       - Confirm the Reintegration subsection (9e) is present for every parallel group and includes merge order, integration test targets, and conflict-prone files
       - Confirm reintegration instructions reference the "story branch" generically, not a hardcoded branch name
       - Confirm the execution order graph is consistent with declared parallel groups and sequential dependencies
       - If any of these checks fail, revise the story before proceeding

15. Pass the story to Review mode for assessment:
    - Switch to Review mode
    - Provide the following list of questions to be answered:
      a) Are the tests appropriate and adequate for the functionality described?
      b) Are there any security concerns that need to be addressed?
      c) Do the activities in the story follow the conventions in the code base for overall organization and type of code? If not, are those changes listed with a statement that the story is changing one or more conventions?
      d) Are there any places in which logging should exist but has been ignored?
      e) If new interfaces are being created are they defined so that they're consistent with the current conventions in the code base?
    - Await the review results from Review mode

16. Present review results to the user:
    - Display all review findings and responses to the questions from step 15
    - Ask the user how they would like to proceed based on the review results
    - Wait for user direction before continuing

## Output

The script outputs:

- Confirmation of the new story creation including:
  - Story number assigned (e.g., "Story040")
  - Full filename (e.g., "Story040_database-migration.md")
  - Absolute path to the created file
  - The description that was used
- Confirmation that the story was added to the Active Stories section in toc.md
- Summary statistics:
  - Total number of active stories (after addition)
  - Total number of finished stories
  - Number of tasks total (after any decomposition from step 9a)
  - Number of tasks annotated as `[Small]`
  - Number of tasks decomposed from larger tasks (e.g., "Task 5 split into Tasks 5a and 5b")
  - Number of tasks flagged as "borderline — no natural seam" and kept as single units
  - Number of parallel groups declared
  - Execution order graph
- Any warnings encountered (e.g., if finished-stories directory does not exist, note that only active stories were scanned)

## Error Handling

If required constraints aren't met, the user will be prompted. The operations in the script WILL NOT be executed until/unless all parameters are valid. During execution, if an error is encountered, the execution of the script will be paused and the user will be asked if script execution should continue or if the script should be terminated.

Specific error conditions handled:

- Missing required parameter (story_name or description)
- Invalid story_name format (contains characters other than alphanumeric, hyphens, underscores)
- Description exceeds maximum length (200 characters)
- Stories directory does not exist
- Story template file (resources/templates/stories.md) is missing
- toc.md file is missing or malformed (missing Active Stories section)
- Story file already exists (collision with existing filename)
- Permission denied when creating the story file
- Failure to read or write toc.md

## Example Usage

**Creating a new story:**

- story_name: "api-rate-limiting"
- description: "Implement rate limiting for API endpoints to prevent abuse"

**Expected output:**

- Story number 040 assigned
- File created: memory-bank/stories/Story040_api-rate-limiting.md
- Entry added to toc.md Active Stories section
