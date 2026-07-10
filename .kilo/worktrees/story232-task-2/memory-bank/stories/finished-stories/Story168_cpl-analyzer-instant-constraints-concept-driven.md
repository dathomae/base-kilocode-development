# Story168: CPL Analyzer Instant Constraints Concept-Driven Validation

## Goal

Refactor the CPL analyzer to validate `Instant` constraints by looking up corresponding predefined concepts at analysis time, rather than using hard-coded maps and constants. This story is part of [Epic002: CPL Types Implementation](../epics/Epic002_cpl-types-implementation.md).

## References

- [ERO002: Types Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md)
- [Requirements: Talking About Time](../specifications/requirements/cpl/talking-about-time.md)
- [Requirements: Concepts, Types, Units, Measures Syntax](../specifications/requirements/cpl/concepts-types-units-measures-syntax.md)

## Dependencies

- [Story138_cpl-types-analyzer-instant.md](finished-stories/Story138_cpl-types-analyzer-instant.md) - Done
- [Story167_cpl-parser-instant-constraints-complete.md](finished-stories/Story167_cpl-parser-instant-constraints-complete.md) - Done

## Dependent Stories

- [Story169_cpl-standard-definitions-consistency.md](../finished-stories/Story169_cpl-standard-definitions-consistency.md) - Done

## Tasks

1. Update `InstantConstraints` struct - Completed
   a. Write tests for updated `AnalysisInstantConstraints` struct with new fields - Completed
   b. Update `AnalysisInstantConstraints` struct to add `DayOfYear`, `FiscalQuarter`, `DayOfMonth`; rename `Day` to `DayOfMonth` - Completed

2. Remove hard-coded validation maps - Completed
   a. Write tests verifying hard-coded maps are removed and concepts are looked up instead - Completed
   b. Remove `validMonths`, `validDaysOfWeek` maps and hard-coded range constants from the analyzer - Completed

3. Implement concept lookup for string constraints - Completed
   a. Write tests for concept-driven membership validation (`month`, `day_of_week`) - Completed
   b. Implement lookup of `Values` concepts and membership validation in `createInstantConstraints` - Completed

4. Implement concept lookup for integer constraints - Completed
   a. Write tests for concept-driven bounds validation (`year`, `day_of_year`, `day_of_month`, `week_number`, `fiscal_quarter`, `hour`, `minute`, `second`) - Completed
   b. Implement lookup of `Int` concepts and bounds validation in `createInstantConstraints` - Completed

5. Add error handling for missing concepts - Completed
   a. Write tests for missing concept errors - Completed
   b. Implement error reporting when a required concept is missing from the Concepts map - Completed

6. Re-enable skipped integration tests - Completed
   a. Remove the `t.Skip` call for `instant_declarations.cpl` in `analysis/analysis_integration_test.go` - Completed
   b. Remove the `t.Skip` call for `instant_declarations.cpl` in `cli/compiler/cpl_example_test.go` - Completed

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- The analyzer must look up concepts by standard name (`Year`, `Month`, `DayOfMonth`, `DayOfYear`, `WeekNumber`, `FiscalQuarter`, `DayOfWeek`, `Hour`, `Minute`, `Second`) in the Concepts map.
- For `Values` concepts, the analyzer must verify constraint value membership in the concept's ordered value set.
- For `Int` concepts, the analyzer must verify constraint value falls within the concept's generator bounds.
- If a required concept is missing from the Concepts map, the analyzer must report an error.

## Intent

- **Task 1**: To align the analyzer's internal representation with the updated parser constraint fields.
- **Task 2**: To eliminate hard-coded validation data in favor of concept-driven validation.
- **Task 3**: To validate string-valued constraints by checking membership in the corresponding concept's value set.
- **Task 4**: To validate integer-valued constraints by checking bounds against the corresponding concept's generator.
- **Task 5**: To ensure robust error handling when required concepts are not defined.

## Acceptance Criteria

- The analyzer validates `Instant` constraints using concept definitions rather than hard-coded values.
- The analyzer reports errors for out-of-bounds integer constraints.
- The analyzer reports errors for invalid string concept values.
- The analyzer reports errors when required concepts are missing from the Concepts map.
- All unit tests pass, including new tests for concept-driven validation.
- Skipped integration tests for `instant_declarations.cpl` are re-enabled.

## Notes

- The constraint-to-concept mapping is: `year`→`Year`, `month`→`Month`, `day_of_month`→`DayOfMonth`, `day_of_year`→`DayOfYear`, `week_number`→`WeekNumber`, `fiscal_quarter`→`FiscalQuarter`, `day_of_week`→`DayOfWeek`, `hour`→`Hour`, `minute`→`Minute`, `second`→`Second`.
