# TOC List Format Specification

This document defines the canonical list-based format for all `toc.md` files in the project. It replaces the previous markdown-table format with a two-line block format modeled on [`memory-bank/lessons-learned.md`](../../lessons-learned.md).

## Rationale

Markdown tables require all rows in the same table to maintain consistent column widths for readability. When two branches both add entries to a TOC table, the merge conflict spans the entire table because the column alignment whitespace changes for every row. A list-based format isolates each entry so that merges produce conflicts only on the lines that actually differ (the added entry and its immediate neighbor), making them trivial to resolve.

## Format Rules

These rules apply to every TOC entry regardless of pattern.

1. Each entry is exactly two lines: a bold title line and a colon-prefixed description line.
2. The title line contains a markdown link (bolded) followed by any metadata in italics after an em-dash.
3. The description line begins with `:` followed by a space and the description text.
4. Entries are separated by blank lines.
5. Entries are ordered alphabetically or by number, as they were in the original table.

## Pattern Definitions

### Two-Column TOCs

A two-column TOC contains a file or directory name and a description. The original table uses columns `| File | Description |` or `| Directory/File | Description |`.

**Conversion rule**: Produce a bold linked title on the first line and a colon-prefixed description on the second line.

#### Before (table format)

```markdown
| File                 | Description                                       |
| -------------------- | ------------------------------------------------- |
| [brief.md](brief.md) | High-level overview of the ChronoCone project.    |
| [terms.md](terms.md) | Definitions of terms used throughout the project. |
```

#### After (list format)

```markdown
**[brief.md](brief.md)**
: High-level overview of the ChronoCone project.

**[terms.md](terms.md)**
: Definitions of terms used throughout the project.
```

### Three-Column TOCs

A three-column TOC contains a file or directory name, a description, and a status or metadata value. The original table uses columns such as `| Story File | Description | Status |` or `| Directory/File | Agent | Description |`.

**Conversion rule**: Move the third column inline after the bold title, separated by an em-dash and italicized. The description remains on the second line, prefixed with a colon.

#### Before (table format)

```markdown
| Story File                                                 | Description                                                                                                                         | Status |
| ---------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- | ------ |
| [Story001_workspace-setup.md](Story001_workspace-setup.md) | Set up CHRONOCONE as a multi-root workspace with Python backend and React/Node frontend environments, including CI/CD configuration | Done   |
```

#### After (list format)

```markdown
**[Story001_workspace-setup.md](Story001_workspace-setup.md)** — _Done_
: Set up CHRONOCONE as a multi-root workspace with Python backend and React/Node frontend environments, including CI/CD configuration
```

When the third column is an agent or role name rather than a status, the same rule applies:

```markdown
**[system-rules/](system-rules/)** — _system_
: Master index referencing all role-specific rules and persona selection guide.
```

### Multi-Section TOCs

A multi-section TOC groups entries under markdown headings, such as `## Active Stories` and `## Finished Stories`. Each section contains a three-column table.

**Conversion rule**: Preserve the section headings. Convert each table in the section using the three-column pattern above. Maintain the original section order and the original entry order within each section.

#### Before (table format)

```markdown
## Active Stories

Stories that are pending or currently in progress.

| Story File                                                             | Description                                                                                                                           | Status      |
| ---------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- | ----------- |
| [Story008_docker-demo-container.md](Story008_docker-demo-container.md) | Create Docker demo container with PostgreSQL, autostarted Core Service, web-served frontend, and JSON-based definition initialization | Not Started |

## Finished Stories

Stories that have been completed or superseded.

| Story File                                                 | Description                                                                                                                         | Status |
| ---------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- | ------ |
| [Story001_workspace-setup.md](Story001_workspace-setup.md) | Set up CHRONOCONE as a multi-root workspace with Python backend and React/Node frontend environments, including CI/CD configuration | Done   |
```

#### After (list format)

```markdown
## Active Stories

Stories that are pending or currently in progress.

**[Story008_docker-demo-container.md](Story008_docker-demo-container.md)** — _Not Started_
: Create Docker demo container with PostgreSQL, autostarted Core Service, web-served frontend, and JSON-based definition initialization

## Finished Stories

Stories that have been completed or superseded.

**[Story001_workspace-setup.md](Story001_workspace-setup.md)** — _Done_
: Set up CHRONOCONE as a multi-root workspace with Python backend and React/Node frontend environments, including CI/CD configuration
```

## Complete Examples from the Project

### Example: memory-bank/toc.md (Two-Column with Subsections)

```markdown
# Memory Bank - Table of Contents

This directory contains the project documentation and specifications for ChronoCone.

## Project Reference

**[brief.md](brief.md)**
: High-level overview of the ChronoCone project.

**[terms.md](terms.md)**
: Definitions of terms used throughout the project.

## Specifications

The [specifications/](specifications/) directory contains the organized specification of the ChronoCone system:

**[specifications/toc.md](specifications/toc.md)**
: Table of contents for the specifications directory.

**[specifications/chronocone-concepts.md](specifications/chronocone-concepts.md)**
: Key concepts implemented by the ChronoCone system, written for a user audience.
```

### Example: .kilocode/rules/toc.md (Three-Column with Agent)

```markdown
# Rules Directory - Table of Contents

This directory contains all rule definitions for the project agents.

## Rule Files

**[system-rules/](system-rules/)** — _system_
: Master index referencing all role-specific rules and persona selection guide.

**[organization.md](organization.md)** — _system_
: Defines directory structure and organization of memory-bank, scripts, and resources directories.

**[rules-architect/](rules-architect/)** — _Architect_
: System architect rules for designing and planning software systems.

**[rules-ask/](rules-ask/)** — _Ask_
: Rules for explanations and answering questions; strictly read-only inquiry mode.
```

### Example: Multi-Section TOC (Illustrative)

This example shows how a TOC with two sections—Active and Finished—appears in list format. The actual project file at [`memory-bank/stories/toc.md`](../../stories/toc.md) contains only the `## Active Stories` section; finished stories live in [`memory-bank/stories/finished-stories/toc.md`](../../stories/finished-stories/toc.md).

```markdown
# Stories Table of Contents

This file contains a table of active stories for the project.

**For records of canceled stories, see [`canceled-stories/toc.md`](canceled-stories/toc.md). For records of finished stories, see [`finished-stories/toc.md`](finished-stories/toc.md). Consult these records only when information about canceled or finished stories is required.**

## Active Stories

Stories that are pending or currently in progress.

**[Story008_docker-demo-container.md](Story008_docker-demo-container.md)** — _Not Started_
: Create Docker demo container with PostgreSQL, autostarted Core Service, web-served frontend, and JSON-based definition initialization

**[Story014_sql-documentation-and-er-diagram.md](Story014_sql-documentation-and-er-diagram.md)** — _Not Started_
: Add comprehensive comments to SQL schema files explaining each attribute's purpose and relationships, and generate an ER diagram for architecture docs

## Finished Stories

Stories that have been completed or superseded.

**[Story198_cpl-concepts-concept-range-examples.md](Story198_cpl-concepts-concept-range-examples.md)** — _Done_
: Create valid and invalid CPL example files demonstrating Concept Range parser and analyzer functionality

**[Story199_cpl-types-string-boolean-examples.md](Story199_cpl-types-string-boolean-examples.md)** — _Done_
: Create CPL example files demonstrating Story196 String and Boolean built-in types with valid declarations, derivations, and intentional errors
```

## Empty TOC Template

Use this skeleton when creating a new `toc.md` file:

```markdown
# <Directory Name> - Table of Contents

Brief description of what this directory contains.

## <Section Name>

**[file-or-directory-name](path)**
: Description of the file or directory.

**[another-file.md](another-file.md)**
: Description of the file.
```

For a three-column TOC, include the metadata on the title line:

```markdown
# <Directory Name> - Table of Contents

Brief description of what this directory contains.

## <Section Name>

**[file-or-directory-name](path)** — _Status_
: Description of the file or directory.

**[another-file.md](another-file.md)** — _Status_
: Description of the file.
```

## Summary of Conversion Rules

| Original Pattern                              | New Pattern                                  |
| --------------------------------------------- | -------------------------------------------- |
| `\| [file](path) \| Description \|`           | `**[file](path)**\n: Description`            |
| `\| [file](path) \| Description \| Status \|` | `**[file](path)** — _Status_\n: Description` |
| Section headings (`## Active Stories`)        | Preserved unchanged                          |
| Table header and alignment rows               | Removed entirely                             |
| Blank line between table rows                 | One blank line between entries               |

Following these rules ensures every TOC file in the project uses a consistent, merge-friendly format.
