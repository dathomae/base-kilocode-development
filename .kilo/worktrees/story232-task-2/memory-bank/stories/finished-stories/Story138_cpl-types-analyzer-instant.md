# Story138: CPL Types Analyzer - Instant Type (Increment 4)

## Goal

Extend the types analyzer to support Instant types with field-based InstantConstraints and field validation. This is Increment 4 of Epic002 types analysis, building on previous analyzer stories.

## References

- [Epic002_cpl-types-implementation.md](../epics/Epic002_cpl-types-implementation.md) - Parent epic for types implementation
- [ERO002_types-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md) - Analysis phase data structures for Increment 4
- [ERO002_types-incremental-implementation.md](../specifications/design/cpl/eros/ERO002_types-incremental-implementation.md) - Main incremental implementation ERO
- [Story134_cpl-types-analyzer-string-boolean.md](../Story134_cpl-types-analyzer-string-boolean.md) - Analyzer increment 2 (dependency)
- [Story137_cpl-types-parser-instant.md](Story137_cpl-types-parser-instant.md) - Parser increment 4 (dependency)

## Dependencies

- [Story134_cpl-types-analyzer-string-boolean.md](../Story134_cpl-types-analyzer-string-boolean.md) - Must be completed (String/Boolean analysis)
- [Story137_cpl-types-parser-instant.md](Story137_cpl-types-parser-instant.md) - Must be completed (parser support for Instant)

## Dependent Stories

- [Story139_cpl-types-parser-interval.md](../Story139_cpl-types-parser-interval.md) - Not Started
- [Story140_cpl-types-analyzer-interval.md](../Story140_cpl-types-analyzer-interval.md) - Not Started

## Tasks

1. Define field-based InstantConstraints struct - Completed
   a. Define InstantConstraints with Year (*int), Month (*string), Day (*int), Hour (*int), Minute (*int), Second (*int), DayOfWeek (*string), WeekNumber (*int) fields - Completed
   b. Implement isTypeConstraints() marker method - Completed
   c. Implement Validate() checking field values are within valid ranges if specified - Completed

   **Intent**: Provide constraint validation for Instant types.
   **Acceptance Criteria**: InstantConstraints validates correctly with proper field-based constraints.

2. Enhance CreateConstraints factory for Instant type with semantic validation - Completed
   a. Add InstantType case to CreateConstraints switch - Completed
   b. Implement createInstantConstraints helper with field constraint parsing - Completed
   c. Implement field constraint parsing to convert constraint names to struct fields - Completed
   d. Check for duplicate constraint names in parserConstraints - Completed
   e. Validate field constraint values are within valid ranges - Completed

   **Intent**: Convert parser constraints to analysis constraints for Instant types with full semantic validation.
   **Acceptance Criteria**: Instant parser constraints are correctly transformed, duplicates detected, and values validated.

3. Update analysis handlers to handle Instant types - Completed
   a. Verify EnterAnalyzeTypesScope handles Instant type via CreateConstraints - Completed
   b. Add any type-specific processing if needed - Completed

   **Intent**: Ensure the analyzer processes Instant type definitions correctly.
   **Acceptance Criteria**: Instant types are analyzed and stored in the Types lookup map.

   **Pattern**: Handlers use function-based pattern: `EnterAnalyzeTypesScope(table, context, analyzer)`.

4. Write comprehensive unit tests for Instant analysis - Completed
   a. Test InstantConstraints.Validate() with valid field constraints - Completed
   b. Test InstantConstraints.Validate() with invalid field constraint values - Completed
   c. Test createInstantConstraints with field constraints - Completed
   d. Test createInstantConstraints with day_of_week constraint - Completed
   e. Test createInstantConstraints with multiple field constraints - Completed
   f. Test CreateConstraints with InstantType - Completed
   g. Test AnalyzeTypesScope with Instant type definitions - Completed

   **Intent**: Verify Instant analysis functionality works correctly.
   **Acceptance Criteria**: All tests pass, covering valid inputs and error conditions.

## Notes

- This story implements Increment 4 of the types analysis phase as defined in ERO002
- Instant constraints include: field-based constraints (year, month, day, hour, minute, second, day_of_week, week_number)
- No Interval type support in this increment
- **Parsing/Analysis Separation**: Semantic validation (field constraint range validation, duplicate detection) happens in this analysis story, not during parsing (Story137)
- Handlers use function-based pattern: `EnterAnalyzeTypesScope(table, context, analyzer)` not method-based
- Follow Logical TDD Lifecycle: write tests before implementing code for each task
