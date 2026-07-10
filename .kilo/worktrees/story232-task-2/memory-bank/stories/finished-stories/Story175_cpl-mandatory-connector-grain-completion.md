# Story175: CPL Mandatory Connector Grain Completion

## Goal

Complete the remaining work for [Story173_cpl-mandatory-connector-grain.md](Story173_cpl-mandatory-connector-grain.md) by updating all example CPL files to include mandatory `grain` clauses, running the full test suite to verify correctness, updating the Story173 task statuses, and moving Story173 to `finished-stories/`.

## References

- [Plan: Complete Story173 — CPL Mandatory Connector Grain](../plans/complete_story173_plan.md)
- [Story173: CPL Mandatory Connector Grain](Story173_cpl-mandatory-connector-grain.md)
- [Epic004: CPL Connectors Implementation](../epics/Epic004_cpl-connectors-implementation.md)
- [CPL Requirements for Connectors](../specifications/requirements/cpl/connectors.md)

## Dependencies

- [Story173_cpl-mandatory-connector-grain.md](Story173_cpl-mandatory-connector-grain.md) - In Progress
  The grammar, parser listener, analysis validation, and tests for mandatory connector grain must already be implemented before this completion story can begin.

## Dependent Stories

None

## Tasks

1. Update `examples/with_errors/` connector files to include mandatory grain - Completed
   a. For each `.cpl` file in `examples/with_errors/` that contains a `connectors` section, add an `Interval` measure to the `measures` section (e.g., `Day = Interval(unit=Day)`) - Completed
   b. Add `grain Day` to every connector declaration in those files - Completed
   c. Ensure files testing parse errors still fail at parse time with their original syntax error - Completed
   d. Ensure files testing semantic errors now parse successfully and reach the analysis phase - Completed
   e. Specifically update: `connector_duplicate_name.cpl`, `connector_undefined_celltype.cpl`, `connector_undefined_dimension.cpl`, `connector_invalid_name.cpl`, `connector_empty_dimensions.cpl`, `connector_missing_celltype_keyword.cpl`, `connector_missing_close_bracket.cpl`, `connector_missing_comma.cpl`, `connector_missing_dimensions_keyword.cpl`, `connector_missing_open_bracket.cpl`, `connector_trailing_comma.cpl` - Completed

2. Update `examples/with_data/random_data_no_time/test.cpl` to include mandatory grain - Completed
   a. Add an `Interval` measure to the `measures` section - Completed
   b. Add `grain <Measure>` to the connector declaration - Completed

3. Verify no other `.cpl` files or inline test strings lack grain - Completed
   a. Search the codebase for connector declarations missing `grain` - Completed
   b. Update any found occurrences - Completed (no additional occurrences found; all missing-grain inline strings are intentional negative test cases)

4. Run the full test suite and fix failures - Completed
   a. Execute `go test ./...` from `src/chronocone_planning_language` - Completed
   b. Fix any failures related to the example file changes - Completed (no failures)
   c. Verify all parser, analysis, compiler, and integration tests pass - Completed

5. Finalize Story173 and update project records - Completed
   a. Mark all Story173 tasks as Completed - Completed
   b. Copy `memory-bank/stories/Story173_cpl-mandatory-connector-grain.md` to `memory-bank/stories/finished-stories/` - Completed
   c. Update `memory-bank/stories/toc.md` to move Story173 from "In Progress" to "Done" - Completed
   d. Update `memory-bank/stories/finished-stories/toc.md` to include Story173 - Completed

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

For this completion story, the existing test suite from Story173 serves as the test driver. Running `go test ./...` reveals which example files and inline test strings are still missing mandatory grain. Each failure guides the next fix, ensuring the test-and-verify cycle is integrated within the file update tasks.

## Constraints

- The `grain` keyword must remain syntactically mandatory in the grammar; omitting it must produce a parse error.
- Files in `examples/with_errors/` that test parse errors (missing bracket, missing keyword, trailing comma, etc.) must still fail at parse time with their original syntax error after grain is added.
- Files in `examples/with_errors/` that test semantic errors (duplicate name, undefined celltype, undefined dimension) must now parse successfully and reach the analysis phase, so the grain measure must be valid and resolvable.
- The `standard_definitions/standard_measures.cpl` file already defines Interval measures (`Daily`, `Weekly`) that can be referenced as grain values.

## Intent

- **Task 1**: To ensure all `examples/with_errors/` connector files include mandatory grain so they either reach their intended error or parse correctly.
- **Task 2**: To ensure the `random_data_no_time` example file is valid under the mandatory grain requirement.
- **Task 3**: To perform a comprehensive audit of the codebase for any remaining connector declarations missing grain.
- **Task 4**: To confirm the entire system compiles and passes all tests after the example file updates.
- **Task 5**: To accurately record Story173 as completed and move it to the finished stories archive.

## Acceptance Criteria

- All `.cpl` example files with connectors include `grain` in every connector declaration.
- `go test ./...` passes with no failures.
- Story173 task statuses accurately reflect completed work.
- Story173 file is located in `memory-bank/stories/finished-stories/`.
- Both `memory-bank/stories/toc.md` and `memory-bank/stories/finished-stories/toc.md` are updated.
