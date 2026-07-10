# Story167: CPL Parser Instant Constraints Complete Set

## Goal

Update the CPL parser to recognize the complete set of `Instant` constraint names (10 constraints). This replaces the deprecated `day` constraint with the disambiguated names `day_of_week`, `day_of_year`, and `day_of_month`, and adds `fiscal_quarter`, `hour`, `minute`, and `second`. This story is part of [Epic002: CPL Types Implementation](../epics/Epic002_cpl-types-implementation.md).

## References

- [ERO002: Types Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO002_types-parsing-phase.md)
- [Requirements: Talking About Time](../specifications/requirements/cpl/talking-about-time.md)
- [Requirements: Types Definitions](../specifications/requirements/cpl/types-definitions.md)

## Dependencies

- [Story137_cpl-types-parser-instant.md](finished-stories/Story137_cpl-types-parser-instant.md) - Done

## Dependent Stories

- [Story168_cpl-analyzer-instant-constraints-concept-driven.md](Story168_cpl-analyzer-instant-constraints-concept-driven.md) - Not Started

## Tasks

1. Update parser constraint name recognition - Completed
   a. Write tests for `validInstantConstraintNames` to include new names and exclude `day` - Completed
   b. Update `validInstantConstraintNames` to include `day_of_year`, `fiscal_quarter`, `day_of_month`, `hour`, `minute`, `second` and remove `day` - Completed

2. Update parser constraint extraction - Completed
   a. Write tests for `InstantConstraints` struct with new fields - Completed
   b. Update `InstantConstraints` struct to include `DayOfYear`, `FiscalQuarter`, `DayOfMonth`; rename `Day` to `DayOfMonth` - Completed

3. Verify backward compatibility - Completed
   a. Write tests ensuring existing constraints (`year`, `month`, `week_number`, `day_of_week`) continue to parse correctly - Completed
   b. Run existing parser tests and verify no regressions - Completed

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- The parser must remain syntactic-only; no semantic validation changes at parse time.
- The deprecated `day` constraint name must no longer be recognized.
- All 10 constraint names (`year`, `month`, `day_of_month`, `day_of_year`, `week_number`, `fiscal_quarter`, `day_of_week`, `hour`, `minute`, `second`) must be recognized.

## Intent

- **Task 1**: To update the parser's list of valid `Instant` constraint names to match the complete set defined in the requirements.
- **Task 2**: To update the internal parser representation to store the newly added constraint fields.
- **Task 3**: To ensure existing functionality is not broken by the constraint name changes.

## Acceptance Criteria

- The parser successfully parses all 10 `Instant` constraint names.
- The parser rejects the deprecated `day` constraint name.
- The `ParserInstantConstraints` struct correctly contains all new fields.
- All unit tests pass, including new tests for the complete constraint set.

## Notes

- This story is a parser-only change; semantic validation of constraint values occurs in [Story168_cpl-analyzer-instant-constraints-concept-driven.md](Story168_cpl-analyzer-instant-constraints-concept-driven.md).
