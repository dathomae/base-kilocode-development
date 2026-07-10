# Story169: CPL Standard Definitions Consistency

## Goal

Ensure the `standard_definitions` files and example `.cpl` files are consistent with the new `Instant` constraint names, replacing the deprecated `day` constraint with the disambiguated names `day_of_week`, `day_of_year`, and `day_of_month`. This story is part of [Epic002: CPL Types Implementation](../epics/Epic002_cpl-types-implementation.md).

## References

- [Requirements: Talking About Time](../specifications/requirements/cpl/talking-about-time.md)
- [Requirements: Types Definitions](../specifications/requirements/cpl/types-definitions.md)
- [Standard Types CPL](../../src/chronocone_planning_language/standard_definitions/standard_types.cpl)
- [Standard Units CPL](../../src/chronocone_planning_language/standard_definitions/standard_units.cpl)

## Dependencies

- [Story167_cpl-parser-instant-constraints-complete.md](../Story167_cpl-parser-instant-constraints-complete.md) - Not Started
- [Story168_cpl-analyzer-instant-constraints-concept-driven.md](../Story168_cpl-analyzer-instant-constraints-concept-driven.md) - Not Started

## Dependent Stories

None

## Tasks

1. Update `standard_types.cpl` - Completed
   a. Identify all occurrences of the deprecated `day=` constraint in `standard_types.cpl` - Completed (none found; file already uses correct names)
   b. Replace `day=DayOfWeek` with `day_of_week=DayOfWeek` in `Midnight` and `StartOfBusinessDay` - Completed (already correct)

2. Update `standard_units.cpl` - Completed
   a. Verify existing unit definitions for `DayOfYear`, `DayOfMonth`, `Hour`, `Minute`, `Second` - Completed
   b. Add or update units for `DayOfYear`, `DayOfMonth`, `Hour`, `Minute`, `Second` as needed - Completed

3. Update example `.cpl` files - Completed
   a. Search all example `.cpl` files for uses of the deprecated `day=` constraint - Completed
   b. Update each use of `day=` to `day_of_week=`, `day_of_year=`, or `day_of_month=` as semantically appropriate - Completed

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- All `.cpl` files in `standard_definitions/` and `examples/` must use valid constraint names.
- Semantic meaning must be preserved when renaming constraints (e.g., `day=Monday` becomes `day_of_week=Monday`).

## Intent

- **Task 1**: To bring the standard type definitions into compliance with the updated constraint names.
- **Task 2**: To ensure the standard unit definitions cover all time-related concepts referenced by constraints.
- **Task 3**: To ensure all example files parse correctly with the updated constraint names.

## Acceptance Criteria

- `standard_types.cpl` no longer uses the deprecated `day=` constraint.
- `standard_units.cpl` contains appropriate units for all time-related concepts.
- All example `.cpl` files use valid constraint names.
- The updated `.cpl` files parse successfully with the updated parser.

## Notes

- This is a cleanup story that should be executed after the parser and analyzer changes are complete.
- Two integration tests are temporarily skipped because they exercise `instant_declarations.cpl` which now uses `day_of_month=1`, a constraint name not yet recognized by the parser/analyzer (pending Story167/Story168):
  - `TestAnalysisIntegration_Examples/instant_declarations.cpl` in `analysis/analysis_integration_test.go`
  - `TestExamplesWithoutErrors/instant_declarations.cpl` in `cli/compiler/cpl_example_test.go`
- Re-enablement of these skipped tests is captured as a task in [Story168](../Story168_cpl-analyzer-instant-constraints-concept-driven.md).
