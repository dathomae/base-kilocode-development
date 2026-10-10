# Epic Story Template

> **Optional planning aid.** Epics are not required by the current arcs, milestones, plans, and
> stories framework. Use this template only when the project explicitly enables an epic layer.

Epic stories group related stories into larger initiatives. Each epic is a single markdown file stored in `memory-bank/epics/`.

## File Naming Convention

Epic files use the format `EpicNNN_<short-purpose>.md`, where `NNN` is a unique three-digit number (e.g., `Epic001_database-layer.md`).

## Table of Contents

Each epic appears in `memory-bank/epics/toc.md` with its current status and a brief goal statement.

Valid statuses are:

- **Not Started**: No stories have begun
- **In Progress**: One or more stories are being worked on
- **Completed**: All stories are finished

## Header Section

### 1. Title

Provide a clear, concise title for the epic that captures the overall initiative.

### 2. Ultimate Goal

Explain the high-level objective this epic achieves. Describe the end state when all stories in the epic are complete. Provide enough context so a reader can understand what the epic accomplishes without reading every story.

### 3. References

List links to relevant documents. These links are relative to the `memory-bank` directory:

- [requirements](../../memory-bank/specifications/requirements/README.md)
- [architecture](../../memory-bank/specifications/architecture/README.md)
- [design](../../memory-bank/specifications/design/README.md)

### 4. Dependencies

List any epics or stories that must complete before work on this epic can begin:

- [EpicNNN\_<short-purpose>.md](memory-bank/epics/EpicNNN_<short-purpose>.md) - <status>
- [StoryNNN\_<short-purpose>.md](memory-bank/stories/StoryNNN_<short-purpose>.md) - <status>

If this epic has no dependencies, write "None" and briefly explain why.

## Body Section

### 1. Stories

List the stories that comprise this epic. Stories may be executed in any order unless dependencies are specified. Stories from other epics may be interleaved with these stories.

The Stories table tracks each story's status and any specifically associated EROs. Story files may not exist when the epic is created; in such cases, they are marked as "not yet defined".

Use this format when all stories have associated files:

```markdown
| Story                                                                  | Status      | Description                                       | ERO |
| ---------------------------------------------------------------------- | ----------- | ------------------------------------------------- | --- |
| [StoryNNN\_<short-purpose>.md](../stories/StoryNNN_<short-purpose>.md) | Not Started | Brief description of what this story accomplishes | -   |
```

Valid statuses: Not Started, In Progress, Completed.

When a story file does not yet exist, use this format:

```markdown
| Story                                          | Status      | Description                                          | ERO |
| ---------------------------------------------- | ----------- | ---------------------------------------------------- | --- |
| StoryXXX\_<short-purpose>.md (not yet defined) | Not Started | Brief description of what this story will accomplish | -   |
```

When a story has a specifically associated ERO, include the ERO reference:

```markdown
| Story                                                                  | Status      | Description                                       | ERO                                                                              |
| ---------------------------------------------------------------------- | ----------- | ------------------------------------------------- | -------------------------------------------------------------------------------- |
| [StoryNNN\_<short-purpose>.md](../stories/StoryNNN_<short-purpose>.md) | Not Started | Brief description of what this story accomplishes | [ERO001\_<decision-name>.md](../specifications/design/ERO001_<decision-name>.md) |
| StoryXXX\_<short-purpose>.md (not yet defined)                         | Not Started | Brief description of future story                 | [ERO002\_<decision-name>.md](../specifications/design/ERO002_<decision-name>.md) |
```

Use "-" in the ERO column when no ERO is specifically associated with that story.

### 2. Story Ordering and Dependencies

Describe any ordering constraints or dependencies between stories in this epic. Note which stories can be worked on in parallel, and which must be completed before others can begin. Also note any external stories or epics that may be interleaved with this epic's stories.

### 3. Constraints

Explicitly call out any constraints to consider when implementing stories in this epic, such as:

- Architectural constraints (e.g., all stories must follow the layered architecture pattern)
- Technology constraints (e.g., stories must use specific frameworks or libraries)
- Standards or procedures to follow
- Cross-cutting concerns (e.g., logging, error handling, testing standards)

### 4. Acceptance Criteria

Define what must be true to consider the epic successfully completed. These criteria should be verifiable and specific, covering the collective outcome of all stories.

### 5. Requesting Clarification

If at any point during epic construction there is confusion or ambiguity about the goal or how to accomplish it, stop and ask the user for clarification.

## Footer Section (Optional)

Include any additional information relevant to the epic:

1. **Open Questions**: Questions that need answers before work can begin
2. **Known Issues**: Issues identified that need addressing
3. **Additional Context**: Any other relevant information
4. **Related Epics**: Other epics that are related but not direct dependencies

## Example Epic

Below is a minimal example showing all sections properly filled out:

```markdown
# Epic001: Database Layer Implementation

## Title

Database Layer Implementation

## Ultimate Goal

Create a complete, tested database layer that provides consistent data access across all components. This layer will abstract database operations, enforce transactional boundaries, and support both relational and time-series data stores.

## References

- [Database Design](../specifications/architecture/backend/foundation/database/overview.md)
- [Storage Interface](../specifications/design/storage_interface.md)
- [ERO001: Database Technology Selection](../specifications/design/ERO001_database-technology-selection.md)
- [Story013_storage-interface.md](../stories/Story013_storage-interface.md) - Completed

## Dependencies

None - This is a foundational epic for data persistence.

## Stories

| Story                                                                                                   | Status      | Description                                                | ERO                                                                                                         |
| ------------------------------------------------------------------------------------------------------- | ----------- | ---------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| [Story041_relational-schema-implementation.md](../stories/Story041_relational-schema-implementation.md) | Completed   | Defines and implements the core relational database schema | [ERO001_database-technology-selection.md](../specifications/design/ERO001_database-technology-selection.md) |
| [Story042_sql-preprocessor.md](../stories/Story042_sql-preprocessor.md)                                 | In Progress | Creates the SQL preprocessing system for dynamic queries   | -                                                                                                           |
| StoryXXX_connection-pooling.md (not yet defined)                                                        | Not Started | Implements connection pooling for database efficiency      | -                                                                                                           |
| StoryXXX_migration-system.md (not yet defined)                                                          | Not Started | Creates the database migration and versioning system       | [ERO003_migration-strategy.md](../specifications/design/ERO003_migration-strategy.md)                       |

## Story Ordering and Dependencies

Stories in this epic can generally be worked on in parallel, with the following exceptions:

1. **Story042** depends on **Story041** - The SQL preprocessor requires the schema to be defined first
2. **StoryXXX_migration-system** should follow **Story041** and **Story042** - Migration system needs to understand both schema and query patterns

Stories from **Epic002: Service Infrastructure** may be interleaved with these stories as the domain layer and database layer are developed in parallel.

## Constraints

1. All database operations must use the repository pattern defined in the architecture guide
2. Each story must include both unit and integration tests
3. SQL queries must be validated for SQL injection vulnerabilities
4. Database connections must be managed through the connection pool

## Acceptance Criteria

1. All stories in the epic are marked as Completed
2. Integration tests demonstrate successful CRUD operations across all entity types
3. Performance benchmarks meet the criteria defined in the architecture document
4. Migration system can successfully upgrade from version N to version N+1 without data loss
5. Documentation exists for all public APIs in the database layer

## Notes

- Consider using a third-party migration tool instead of building custom
- PostgreSQL is the primary target, but keep MariaDB compatibility in mind
```
