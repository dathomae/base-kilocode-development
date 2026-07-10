# Story174: CPL Analyzer Instant Constraints Completion

## Goal

Finish the remaining work for Story168 by updating tests and examples to use 0-based ranges for `Hour`, `Minute`, and `Second` (matching the corrected `standard_concepts.cpl`), and re-enable the skipped integration tests for `instant_declarations.cpl` and `interval_declarations.cpl`.

## References

- [Story168: CPL Analyzer Instant Constraints Concept-Driven Validation](Story168_cpl-analyzer-instant-constraints-concept-driven.md)
- [Plan: Complete Story168](../plans/complete_story168_plan.md)
- [ERO002: Types Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md)
- [Requirements: Talking About Time](../specifications/requirements/cpl/talking-about-time.md)

## Dependencies

- [Story167_cpl-parser-instant-constraints-complete.md](finished-stories/Story167_cpl-parser-instant-constraints-complete.md) - Done
- [Story168_cpl-analyzer-instant-constraints-concept-driven.md](Story168_cpl-analyzer-instant-constraints-concept-driven.md) - In Progress
- [Story169_cpl-standard-definitions-consistency.md](finished-stories/Story169_cpl-standard-definitions-consistency.md) - Done

## Dependent Stories

None

## Tasks

1. Update standard concepts helper tests for 0-based ranges - Completed
   a. Write tests expecting 0-based bounds (`Hour[0,23]`, `Minute[0,59]`, `Second[0,59]`) - Completed
   b. Update assertions, remove 1-based comments, and verify tests pass - Completed

2. Update CPL types internal tests for 0-based bounds - Completed
   a. Write updated out-of-bounds and boundary test cases for 0-based ranges - Completed
   b. Update test values and expected error messages, verify tests pass - Completed

3. Update inline concept definitions in analysis tests - Completed
   a. Write tests with updated inline 0-based `Second` and `Hour` concepts - Completed
   b. Update inline concept definitions and expected error messages, verify tests pass - Completed

4. Update interval declarations example to 0-based values - Completed
   a. Verify example syntax is valid with 0-based time values - Completed
   b. Update `Midnight` example to use `Instant(hour=0, minute=0, second=0)` - Completed

5. Re-enable instant and interval declarations in analysis integration tests - Completed
   a. Update `getAnalysisForFile` helper to optionally prepend standard definition files - Completed
   b. Remove skip flags for `instant_declarations.cpl` and `interval_declarations.cpl`, load standard definitions, and verify tests pass - Completed

6. Re-enable instant and interval declarations in CLI compiler tests - Completed
   a. Update compiler test setup to configure `standardDefsDir` correctly - Completed
   b. Remove skip flags and verify compilation succeeds for both files - Completed

7. Run full verification and finalize Story168 - Completed
   a. Run all affected test packages (`analysis/...`, `cli/compiler/...`) and fix any remaining failures - Completed
   b. Update Story168 to mark all tasks complete and move it to `finished-stories/` - Completed

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- `Hour`, `Minute`, and `Second` in `standard_concepts.cpl` use 0-based ranges and must remain consistent with test assertions.
- The analyzer must continue to validate `Instant` constraints using concept definitions rather than hard-coded values.
- Integration tests for `instant_declarations.cpl` and `interval_declarations.cpl` must load `standard_definitions/standard_concepts.cpl` without affecting other example files that define their own conflicting concepts.
- All existing tests outside the affected files must continue to pass.

## Intent

- **Task 1**: Align bounds verification tests with the corrected 0-based standard concepts.
- **Task 2**: Ensure out-of-bounds and boundary tests correctly reflect 0-based ranges.
- **Task 3**: Update inline concept definitions in tests to match the corrected standard concepts.
- **Task 4**: Update example files to use semantically correct 0-based time values.
- **Task 5**: Restore integration test coverage for instant declarations by loading standard definitions correctly.
- **Task 6**: Restore CLI compiler test coverage for instant and interval declarations.
- **Task 7**: Verify no regressions exist and properly close out Story168.

## Acceptance Criteria

- All bounds verification tests assert 0-based ranges for `Hour`, `Minute`, and `Second`.
- All out-of-bounds and boundary tests use correct 0-based values and error messages.
- Inline concept definitions in tests use 0-based ranges.
- The `Midnight` interval example uses `Instant(hour=0, minute=0, second=0)`.
- `instant_declarations.cpl` and `interval_declarations.cpl` are no longer skipped in analysis integration tests.
- `instant_declarations.cpl` and `interval_declarations.cpl` are no longer skipped in CLI compiler tests.
- All affected test packages pass without failures.
- Story168 is moved to `finished-stories/` with all tasks marked complete.

## Notes

- The corrected standard concept definitions are:
  - `Hour = [Int[0, 23, 1]]`
  - `Minute = [Int[0, 59, 1]]`
  - `Second = [Int[0, 59, 1]]`
- Files modified will include test files in `analysis/` and `cli/compiler/`, example files, and story tracking files.
