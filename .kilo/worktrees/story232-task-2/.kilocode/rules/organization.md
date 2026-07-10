There are three key top level directories:

1. memory-bank which is used to capture information will be referenced for each task
2. scripts which will define scripts which include step by step instructions to accomplish some specific function.
3. resources which capture a variety of miscellaneous information.

# The rules directory

The `.kilocode/rules/` directory contains all rule definitions for the project:

- `system-rules/system-rules.md` - Master index referencing all role-specific rules
- `organization.md` - This file; defines directory structure and organization
- `end_of_task_activities.md` - Standard activities to perform at task completion
- `rules-<role>/<role>.md` - Role-specific instruction files (e.g., rules-work-planner, rules-code, rules-architect). When asked to take a particular role, you should read the rules-<role>/<role>.md file and follow the instructions in that file.

# The memory-bank directory

The following files are standard in memory banks:

1. terms.md is a markdown file that contains a table that captures the definition of terms. There are two columns, the first is the term being defined, the second is a definition of the terms, sometimes including examples.
2. concepts.md is a markdown file that contains a table that captures the description of a concept. Concepts tend to be more general or abstract than terms. The concepts.md file includes a table with three columns, the first the name or a phrase that references the concept, the second text that describes the concept, possibly with examples and the third that lists concepts and/or terms that are CLOSELY related. Many concepts can be quite general, so the related concepts or terms should be limited to those that are CLOSELY related.
3. lessons-learned.md is a markdown file with two columns. The first is a short phrase the describes the lesson, the second is a longer description, possibly containing examples. The purpose of lessons-learned.md is to capture longer term learnings that will result in mistakes not being repeated.
4. bugs.md is a markdown file that is like lessons-learned, but that concentrates on coding specific issues. bugs.md contains a table with two columns, the first is a short description of the bug the second is a longer description of the bug, possibly with examples. bugs.md is intended to capture learnings about coding bugs that have been encountered in the past so that they won't be repeated.
5. brief.md contains a high level overview of the project.
6. context.md describes the task or task that is in progress, along with relevant details and pointers to other files (e.g., file that capture a plan or story) that are relevant to a task. Note that context.md is NOT a historical log of work. When a task or story is completed information for it should be removed from context.md, with the possible exception of a BRIEF statement that some task or story has already been completed. context.md should be updated at both the start and end of every task. (Note: context.md tracks active tasks and current context, while concepts.md captures broader, general concepts and should be updated as needed for conceptual changes, not necessarily tied to task start/end.)
7. requirements.md defines the requirements for the project.
8. design.md captures the design of the project.
9. Tasks will always be broken down into fine-grained segments that will come in one of two forms, either story directories or plan files. Both are defined below.

Plan files shorter or more targeted objectives will be captured in "plan" files.
Plan files contain a description of the objective and a step by step plan to
achieve it. Plans will usually be executed across multiple tasks, where the user
will check the result at the end of each step. Steps should be numbered simply, without
breaking them into phases or stages.

More elaborate/longer term plans will be captured in a stories directory
that will contain a group of markdown files with one story defined in each
file. In addition there will be a toc.md file that contains a list of
contents with entries that name the story file,
provide a brief description of the story, and
capture the state of the story as one of "Not Started", "In Progress" or "Done".

# The scripts directory

The scripts directory contains scripts that will be used to follow some specific step-by-step procedure. Script files will
always end in .script.md. A specialized type of script called a "prompt driven development" (PDD) script will always
have a name that ends in .pdd.script.md.

# The resources directory

The resources directory will capture information of many types two of which are:

1. A templates directory, which defines template files that will always end in .template.md
2. An apis directory that will contain subdirectories for information about using various apis or libraries.

Both the templates and apis directories will include a toc.md file that contains a list with
entries. Each entry names the file or directory that contains
a particular type of template or api definition and provides a brief description of
what the file or directory contains. The purpose of the toc.md file is to make it easy to
find the information required for some task without having to search through a lot of files

# Mode Discipline and Transitions

Movement between agent personas and project lifecycle stages must follow a rigorous, non-linear path that prevents premature action and implementation:

1. **The Investigation-to-Action Boundary**: An inquiry conducted in **Ask** mode must never lead directly to a change in the filesystem. Findings from an investigation must be presented to the user, who then decides whether to proceed to a planning stage.
2. **Strategy Before Execution**: No implementation work (**Code**, **Debug**, **Tech Writer**) may begin without an approved **Plan** (via `*_plan.md`) or **Story** (via `StoryNNN_*.md`).
3. **Transition Approval**: When finishing a plan in **work-planner** mode, you must explicitly ask the user for approval of the plan and for permission to switch to a routing persona (**story-implementor** or **Orchestrator**) for execution.
4. **Router Agency**: Major execution tasks should always be driven by a routing persona following a PDD script (e.g., `story-implementor.pdd.script.md`) to ensure consistency and proper delegation to specialized agents.
