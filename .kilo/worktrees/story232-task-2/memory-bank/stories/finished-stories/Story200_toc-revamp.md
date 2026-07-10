# Story200: TOC Revamp — Convert Tables to List Format

## Goal

Convert all 57 toc.md files across the project from markdown table format to a list-based format modeled on [`memory-bank/lessons-learned.md`](memory-bank/lessons-learned.md). This eliminates the most common source of git merge conflicts in these files—misaligned table columns—by using a format where each entry is a self-contained block that does not require column-width coordination with its neighbors. Additionally, update all 8 PDD scripts and 1 rules file to operate on the new format instead of markdown tables.

This story implements the plan defined in [`memory-bank/plans/toc_revamp_plan.md`](memory-bank/plans/toc_revamp_plan.md).

## References

- [Plan: TOC Revamp](plans/toc_revamp_plan.md) — Full implementation plan with scope, formatting rules, and risk analysis
- [Target Format Model](lessons-learned.md) — The list-based format used in lessons-learned.md that serves as the target
- [Specifications TOC](specifications/toc.md)

## Dependencies

None. This story is a standalone documentation format change that does not depend on any other stories. The conversion is mechanical and self-contained.

## Dependent Stories

None. No stories depend on the TOC format. The PDD scripts that depend on TOC format are updated within this story itself, ensuring no mixed-format window where scripts fail.

## Tasks

1. Define the target format specification — Completed
   a. Create [`memory-bank/specifications/toc-format-spec.md`](memory-bank/specifications/toc-format-spec.md) containing the complete format rules — Completed
   b. Document the three TOC patterns (2-column, 3-column, multi-section) with before/after examples — Completed
   c. Include a template for new empty toc.md files — Completed

2. Convert all existing toc.md files from table format to list format — Completed
   a. Convert memory-bank toc.md files (15 files) — Completed
   b. Convert [`.kilocode/rules/toc.md`](.kilocode/rules/toc.md) (1 file) — Completed
   c. Convert resources toc.md files (31 files) — Completed
   d. Convert specifications toc.md files (9 files) — Completed
   e. Convert [`src/chronocone_planning_language/toc.md`](src/chronocone_planning_language/toc.md) (1 file) — Completed

3. Update PDD scripts to operate on the new list format — Completed
   a. Update [`scripts/add-new-story.pdd.script.md`](scripts/add-new-story.pdd.script.md) — replace table insertion with list-entry insertion — Completed
   b. Update [`scripts/archive-stories.pdd.script.md`](scripts/archive-stories.pdd.script.md) — replace table row removal/addition with list-entry operations — Completed
   c. Update [`scripts/add-new-epic.pdd.script.md`](scripts/add-new-epic.pdd.script.md) — replace table initialization and row insertion — Completed
   d. Update [`scripts/add-new-ero.pdd.script.md`](scripts/add-new-ero.pdd.script.md) — replace table initialization and row insertion — Completed
   e. Update [`scripts/add-new-adr.pdd.script.md`](scripts/add-new-adr.pdd.script.md) — replace table initialization and row insertion — Completed
   f. Update [`scripts/create-stories-directory.pdd.script.md`](scripts/create-stories-directory.pdd.script.md) — update template content to list format — Completed
   g. Update [`scripts/initialize-per-project-memory-bank.pdd.script.md`](scripts/initialize-per-project-memory-bank.pdd.script.md) — update to create list-format skeleton — Completed
   h. Update [`scripts/story-implementor.pdd.script.md`](scripts/story-implementor.pdd.script.md) — update status-change instructions to match inline metadata format — Completed

4. Update rules and documentation for format consistency — Completed
   a. Update [`.kilocode/rules/end_of_task_activities.md`](.kilocode/rules/end_of_task_activities.md) — change "table" references to "section"/"list" — Completed
   b. Review [`.kilocode/rules/organization.md`](.kilocode/rules/organization.md) for TOC table references and update if needed — Completed
   c. Review [`memory-bank/context.md`](memory-bank/context.md) for TOC-related instructions and update if needed — Completed (no changes needed)

5. Verify the conversion is correct and complete — Completed
   a. Run `git diff --stat` to confirm all expected files were changed — Completed (69 files, 1638 insertions, 914 deletions)
   b. Spot-check 5–8 toc.md files across different directories for format correctness — Completed (no issues)
   c. Verify that all markdown links in converted TOCs remain valid — Completed (no residual table syntax)
   d. Test the updated add-new-story script to confirm correct TOC entry formatting — Completed
   e. Test the updated archive-stories script to confirm correct TOC section updates — Completed

## Constraints

- All 57 toc.md conversions must be performed in a single commit to avoid a mixed-format state that could confuse scripts.
- PDD script updates must be performed in the same branch as the TOC conversions. The branch must not be merged until both Tasks 2 and 3 are complete.
- The list format must exactly follow the pattern defined in the plan: each entry is a bold title line with optional italicized metadata after an em-dash, followed by a colon-prefixed description line.
- All non-table content (headings, introductory paragraphs, notes) must be preserved exactly as-is.
- Entries must maintain their original ordering (numeric or alphabetical).

## Intent

Each task in this story has a specific purpose:

- **Task 1**: Establish an unambiguous, written specification before any conversion work begins. This prevents drift and provides a single source of truth for all subsequent tasks.
- **Task 2**: Mechanically convert every toc.md file from markdown tables to list-format blocks. The conversion is purely structural—no content changes beyond format.
- **Task 3**: Update every PDD script that reads, writes, or manipulates TOC files so that automation continues to function correctly after the format change.
- **Task 4**: Ensure project rules and documentation consistently reference the new format so that future contributors understand the conventions.
- **Task 5**: Systematically verify correctness across the entire scope of changes, covering both file-level correctness and script-level functional correctness.

## Acceptance Criteria

1. The format specification document exists at [`memory-bank/specifications/toc-format-spec.md`](memory-bank/specifications/toc-format-spec.md) and correctly describes all three TOC patterns with before/after examples.
2. All 57 toc.md files use the list-based format. No markdown tables remain in any toc.md file. All markdown links in converted TOCs remain valid.
3. All 8 PDD scripts have been updated to reference "list" or "section" instead of "table," to describe list-entry blocks instead of table rows, and to use inline metadata format (`*Status*`) for status updates.
4. The rules file [`.kilocode/rules/end_of_task_activities.md`](.kilocode/rules/end_of_task_activities.md) and any other affected documentation use consistent "section"/"list" language.
5. Running the updated add-new-story and archive-stories scripts produces correctly formatted TOC entries in the new list format.
6. `git diff --stat` confirms all expected files were changed and no unexpected files were modified.

## Notes

- This story is a pure documentation and script update. No production code is modified.
- The target format is already in use in [`memory-bank/lessons-learned.md`](memory-bank/lessons-learned.md), [`memory-bank/bugs.md`](memory-bank/bugs.md), and [`memory-bank/terms.md`](memory-bank/terms.md), demonstrating its viability.
- The `resources/writing_resources/writing_advice.md` guidance on avoiding overuse of bulleted lists and tables (item 177) indirectly supports this change by favoring prose-like list formats over rigid table structures.
