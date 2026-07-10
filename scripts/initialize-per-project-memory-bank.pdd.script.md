# Initialize Per-Project Memory Bank

This script initializes a complete per-project memory bank structure in a specified project directory, including all directories and files referenced in the global organization rules. Normally the project root directory will be the root directory of the project that is open when the command to initialize the memory bank is given.

## Parameters

| Name         | Type   | Description                                                                  | Required | Default Value |
| ------------ | ------ | ---------------------------------------------------------------------------- | -------- | ------------- |
| project_path | string | The path to the project root directory where the memory bank will be created | Yes      | N/A           |
| project_name | string | A short name for the project to use in placeholder content                   | No       | "this project" |
| project_description | string | A brief description of the project for brief.md                         | No       | "TBD - please describe the project" |

## Operations

1. Create the `memory-bank/` directory in the project root.

2. Create the core memory bank files:
   a. `terms.md` with an empty definitions section and a header: `# Terms`
   b. `concepts.md` with an empty section and a header: `# Concepts`
   c. `lessons-learned.md` with an empty section and a header: `# Lessons Learned`
   d. `bugs.md` with an empty table and a header: `# Bugs`
   e. `brief.md` with the content: `# <project_name>` followed by the project_description.
   f. `context.md` with placeholder content describing the context-tracking purpose of the file.
   g. `toc.md` as the master index (see content specification below).

3. Create the `memory-bank/specifications/` directory with subdirectories:
   a. `requirements/` — Create `README.md` describing the purpose of requirements specs and conventions, and `toc.md` with an empty list section.
   b. `architecture/` — Create `README.md` describing the purpose of architecture specs and conventions, and `toc.md` with an empty list section.
   c. `design/` — Create `README.md` describing the purpose of design specs (version-specific decisions) and conventions, and `toc.md` with an empty list section.

4. Create the `memory-bank/stories/` directory with:
   a. `toc.md` with a header `# Stories Table of Contents`, an introductory paragraph, and two sections:
      - `## Active Stories` — currently empty
      - `## Finished Stories` — placeholder note pointing to `finished-stories/toc.md`

5. Create `memory-bank/stories/finished-stories/` subdirectory.

6. Create `memory-bank/stories/canceled-stories/` subdirectory.

7. Create the `memory-bank/epics/` directory with:
   a. `toc.md` with a header `# Epics Table of Contents` and an empty list section.

8. Create the `memory-bank/plans/` directory.

9. Verify all directories and files were created successfully.

## master toc.md content

The `memory-bank/toc.md` file should have this structure:

```markdown
# Memory Bank - Table of Contents

This directory contains the project documentation and specifications.

## Project Reference

**[brief.md](brief.md)**
: High-level overview of the project.

**[terms.md](terms.md)**
: Definitions of terms used throughout the project.

**[concepts.md](concepts.md)**
: Quick reference guide to concepts used throughout the project.

**[context.md](context.md)**
: Current task context and active work tracking.

**[lessons-learned.md](lessons-learned.md)**
: Captured learnings to avoid repeating mistakes.

**[bugs.md](bugs.md)**
: Documentation of past bugs and their resolutions.

## Specifications

The [specifications/](specifications/) directory contains the organized specification of the system:

**[specifications/toc.md](specifications/toc.md)**
: Table of contents for the specifications directory.

**[specifications/requirements/](specifications/requirements/)**
: Requirements specifications extending concepts with lower-level details.

**[specifications/architecture/](specifications/architecture/)**
: System organization, schema definitions, service division, and endpoint definitions.

**[specifications/design/](specifications/design/)**
: Version-specific design decisions not general enough for architecture.

## Stories

The [stories/](stories/) directory contains implementation stories:

**[stories/toc.md](stories/toc.md)**
: Table of contents for implementation stories.

**[stories/Story###\_\*.md](stories/)**
: Individual story files capturing implementation tasks.

## Epics

The [epics/](epics/) directory contains epic stories grouping related stories:

**[epics/toc.md](epics/toc.md)**
: Table of contents for epic stories.

## Plans

The [plans/](plans/) directory contains shorter plan files for targeted objectives.

## Scripts

The [scripts/](../scripts/) directory contains PDD (Prompt Driven Development) scripts for project operations.

## Resources

**[resources/toc.md](../resources/toc.md)**
: Table of contents for resources (templates, APIs, writing guides).
```

## Output

The script outputs:
- The path to the initialized memory bank directory
- A confirmation message indicating all files and directories were created successfully
- A count of directories and files created

## Error Handling

If required parameters aren't specified, the user should be prompted. The operations in the script WILL NOT be executed until/unless all required parameters have been specified. During execution, if an error is encountered (e.g., permission denied), the execution of the script will be paused and the user will be asked if script execution should continue or if the script should be terminated.
