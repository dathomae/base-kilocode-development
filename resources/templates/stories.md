# Story Template

Stories break large plans into manageable chunks. Each story is a single markdown file stored in `memory-bank/stories/`.

## File Naming Convention

Story files use the format `StoryNNN_<short-purpose>.md`, where `NNN` is a unique three-digit number (e.g., `Story001_workspace-setup.md`). If a story is
part of an epic story, the file name should reference the epic story number as well, for example `StoryNNN_epicEEE_<short_purpose>.md`

## Table of Contents

Each story appears in `memory-bank/stories/toc.md` with its current status and a brief goal statement.

Valid statuses are:

- **Not Started**: Work has not begun
- **In Progress**: Work is actively being done
- **Completed**: All tasks are finished

Most stories involve writing code, though some cover documentation or planning tasks.

## Header Section

### 1. Goal

Explain the functionality being implemented in this story. Provide enough context so a reader can understand what the story accomplishes without reading the entire plan.

If the story is part of an epic, include a link to the epic story file in the goal statement.

### 2. References

List links to relevant documents. These links are relative to the `memory-bank` directory:

- [requirements](specifications/requirements/requirements.md)
- [architecture](specifications/architecture/architecture.md)
- [design](specifications/design/design.md)

### 3. Dependencies

List any stories that must complete before work on this story can begin:

- [StoryNNN\_<short-purpose>.md](memory-bank/stories/StoryNNN_<short-purpose>.md) - <status>

If this story has no dependencies, write "None" and briefly explain why (e.g., "This is the foundational story for the project").

### 4. Dependent Stories

List any stories that are blocked by this story:

- [StoryNNN\_<short-purpose>.md](memory-bank/stories/StoryNNN_<short-purpose>.md) - <status>

If no stories depend on this one, write "None".

## Body Section

### 1. Tasks

List the tasks that need to be completed to implement the story. Use simple numbering without grouping into stages or phases. Include sufficient detail so it is clear exactly what needs to be done.

Use this format:

```markdown
1. <task description> - <status>
   a. <sub-task description> - <status>
   b. <sub-task description> - <status>
```

Valid statuses: Not Started, In Progress, Completed.

Tasks can be annotated with `[Small]` on the status line to route them to the lightweight `small-code-for-story-implementor` agent:

```markdown
### Task 3: Extend evaluator for comparison operators — Not Started [Small]
```

When `[Small]` is absent, tasks route to `code-for-story-implementor` (medium) or `technical-writer-for-story-implementor` (writing). See the Task Sizing subsection below for rules on when to use `[Small]`.

### File Path Conventions

All file paths in story tasks are relative to the project's `source_root` — the directory where build and test commands run. The story-implementor passes `source_root` to sub-agents via `story_context`, and sub-agents resolve file paths as `<source_root>/<story_path>`.

**Concrete path rules:**

1. **Include the full path within `source_root`.** File paths in task descriptions must include the complete directory structure from `source_root` down. This ensures sub-agents can locate files without ambiguity regardless of project layout.
   - ✅ `src/packages/utils/helper.py` (full path from source_root)
   - ❌ `utils/helper.py` (ambiguous)

2. **Commands run from `source_root`.** The sub-agent is told to `cd <source_root>` before running build and test commands. All paths in command examples must be relative to `source_root`.
   - ✅ `cd src && go test ./packages/utils/...`
   - ❌ `go test ./utils/...`

3. **Tests that span multiple packages or directories list each path explicitly:**
   - ✅ `go test ./packages/parsing/... ./packages/analysis/...`
   - ❌ `go test ./parsing/... ./analysis/...`

These conventions ensure sub-agents can locate and operate on files without ambiguity, regardless of the project's source directory structure.

### Task Sizing

Tasks should be sized to enable efficient subagent routing by the story-implementor PDD script. A task annotated `[Small]` on its status line is routed to the lightweight `small-code-for-story-implementor` agent; unannotated tasks are routed to `code-for-story-implementor`.

**Design principle — prefer small tasks when natural.** When creating tasks, prefer decomposing work into straightforward, narrow-scope tasks that meet the Small criteria below, provided the decomposition is natural and does not introduce complexity. Do NOT force smallness by:
- Creating artificial intermediate artifacts (temp modules, adapters) just to stay under 2 files
- Splitting a coherent multi-file change (e.g., "add struct AND its method") into separate tasks
- Adding more parallel worktrees and merge steps than the isolation benefit justifies

Rules of thumb — a task is likely **Small** when:
- It modifies a single source file and its corresponding test file (at most 2 files total)
- It adds a test that follows an existing pattern (e.g., a new test function in an existing `_test.go` file for an already-defined function)
- It adds a small code change in one file plus an associated test in a separate test file (the test scaffolds off the code change)
- It adds cases to an existing switch statement or extends an existing function without introducing new structs or interfaces
- It is a "mechanical" change — renaming, moving code between files in the same package, or adding a pattern that's already established elsewhere
- It creates a single example directory following an existing template

A task is likely **not Small** when:
- It introduces coordinated non-test changes across 3+ source files (e.g., new struct in one file, new method in another, new evaluator case in a third)
- It lays down new code infrastructure (multiple structs, new interfaces, new package-level abstractions)
- It requires reasoning across multiple concerns or has complex coordination logic (e.g., a recursive walker that needs to handle a new node type with sub-expressions)
- It adds a new operation kind that requires coordination between a struct producer and a struct consumer
- It spans multiple packages with interdependent changes

**When in doubt, do not annotate `[Small]`.** Defaulting to `code-for-story-implementor` is the safer choice. The story author consults the `architect-for-story-planning` subagent for uncertain cases.

### Parallel Execution

When two or more tasks touch completely disjoint sets of files and neither depends on the other's output, they can run in parallel during story execution. Each task runs in its own isolated git worktree. List each parallel group below with a brief rationale.

Format:
```markdown
- **Group A** (Tasks N, M): Both tasks modify disjoint files in different packages;
  neither depends on the other's output.
- **Group B** (Tasks P–Q): Both tasks modify files in the same package
  (shared compilation risk); listed here because they share a package,
  but each will run in its own isolated worktree.
```

These groups are referenced in the Execution Order subsection as parallel branches. The story-implementor PDD script reads this section to determine which tasks to hand off as concurrent subagent invocations, each in its own git worktree.

If no tasks can be parallelized, write "None — all tasks have sequential dependencies."

### Execution Order

The dependency graph among tasks, showing both sequential chains and parallel branches. Notation:
- `[Task N]` — a single task
- `[Task N ‖ Task M]` — tasks that run in parallel (references groups from Parallel Execution)
- `→` — sequential dependency (left must complete before right begins)

Only list the first occurrence of each task; parallel groups reference the groups defined in the Parallel Execution section above.

Example:
```
[Task 5 ‖ Task 6] → Task 7 → Task 8 → [Task 9.1 ‖ Task 9.2 ‖ Task 9.3] → Task 9.4
```

### Reintegration

Documents how to merge worktree branches back into the story branch and verify the combined result for each parallel group. Generated by the `add-new-story.pdd.script.md` step 9d based on the architect-for-story-planning's reintegration analysis.

Each parallel group must specify:
- **Merge order**: which worktree branch to merge first, or "any order"
- **Integration tests**: exact test command(s) to run after merging all worktree branches
- **Conflict-prone files**: files to watch during merge (lock files, shared test files not modified by any task)

Format:
```markdown
**Group A (Tasks 3, 4):**
- Merge order: Task 3 first (modifies Makefile — shared infrastructure), then Task 4
- Integration tests: `go test ./pkg/evaluator/... -count=1 -timeout 60s` — verify combined evaluator behavior
- Watch for conflicts in: `go.sum` (dependency resolution may differ per worktree)

**Group B (Tasks 7, 8):**
- Merge order: any order (fully disjoint packages, no shared infrastructure)
- Integration tests: `go test ./pkg/parser/... ./pkg/formatter/... -count=1 -timeout 60s`
- Watch for conflicts in: none expected
```

All branch references should use "story branch" generically. The story-implementor determines the actual branch at runtime.

If no parallel groups exist, omit this subsection or write "None — no parallel groups."

### 2. Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

### 3. Constraints

Explicitly call out any constraints to consider when implementing the story, such as:

- Algorithmic constraints (e.g., the solution must run in O(n) time)
- Memory constraints
- Required packages or libraries
- Standards or procedures to follow

### 4. Intent

For each task, explicitly state the overall goal. This ensures the purpose remains clear even if implementation details change.

### 5. Acceptance Criteria

Explicitly define what must be true to consider the task successfully completed. These criteria should be verifiable and specific.

### 6. Requesting Clarification

If at any point during story construction there is confusion or ambiguity about the goal or how to accomplish it, stop and ask the user for clarification.

## Footer Section (Optional)

Include any additional information relevant to the story:

1. **Open Questions**: Questions that need answers before work can begin
2. **Known Issues**: Issues identified that need addressing
3. **Additional Context**: Any other relevant information

## Example Story

Below is a minimal example showing all sections properly filled out:

```markdown
# Story001: Basic Workspace Setup

## Goal

Set up the project workspace with properly configured backend and frontend environments.

## References

- [Architecture](../architecture/v0_architecture.md)
- [Brief](../brief.md)

## Dependencies

None - This is the foundational story for the project.

## Dependent Stories

- [Story002_backend-structure.md](Story002_backend-structure.md) - Not Started

## Tasks

### Task Sizing

All tasks below touch single files or follow established patterns. Tasks 1 and 2 are file-disjoint and can be annotated `[Small]`.

### Parallel Execution

- **Group A** (Tasks 1, 2): Task 1 creates the workspace config file; Task 2 creates the backend environment directory. Disjoint files, no shared dependencies.

### Execution Order

```
[Task 1 ‖ Task 2] → Task 3
```

### Reintegration

**Group A (Tasks 1, 2):**
- Merge order: any order (fully disjoint files, no shared infrastructure)
- Integration tests: `cd <source_root> && <test_command>` — verify combined workspace configuration
- Watch for conflicts in: none expected

### Task 1: Create workspace configuration — Not Started [Small]
   a. Create workspace file defining project roots — Not Started
   b. Configure workspace-specific settings — Not Started

### Task 2: Set up backend environment — Not Started [Small]
   a. Create Python virtual environment — Not Started
   b. Create requirements.txt with dependencies — Not Started

### Task 3: Verify environment configuration — Not Started
   a. Run all setup validation scripts — Not Started
   b. Confirm all dependencies resolve — Not Started

## Notes

- The project uses the chosen language and framework for its backend
- This story focuses on basic setup; detailed structure is in Story002
```
