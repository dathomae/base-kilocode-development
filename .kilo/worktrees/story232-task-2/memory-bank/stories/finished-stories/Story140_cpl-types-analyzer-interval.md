# Story140: CPL Types Analyzer - Interval Type (Increment 5)

## Goal

Extend the types analyzer to support Interval types with IntervalConstraints, cross-reference validation (starts_at must be Instant-derived), and unit reference validation. This is the final increment (5) of Epic002 types analysis, completing full type functionality.

## References

- [Epic002_cpl-types-implementation.md](../epics/Epic002_cpl-types-implementation.md) - Parent epic for types implementation
- [ERO002_types-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md) - Analysis phase data structures for Increment 5
- [ERO002_types-incremental-implementation.md](../specifications/design/cpl/eros/ERO002_types-incremental-implementation.md) - Main incremental implementation ERO
- [Story138_cpl-types-analyzer-instant.md](Story138_cpl-types-analyzer-instant.md) - Analyzer increment 4 (dependency)
- [Story139_cpl-types-parser-interval.md](Story139_cpl-types-parser-interval.md) - Parser increment 5 (dependency)

## Dependencies

- [Story138_cpl-types-analyzer-instant.md](Story138_cpl-types-analyzer-instant.md) - Must be completed (Instant analysis)
- [Story139_cpl-types-parser-interval.md](Story139_cpl-types-parser-interval.md) - Must be completed (parser support for Interval)

## Dependent Stories

- None - This completes the types implementation epic

## Tasks

1. Implement IntervalConstraints struct with Validate method - Completed
   a. Define IntervalConstraints with StartsAt (*InstantConstraints, for inline), StartTypeName (*string, for named type reference), and UnitName (\*string) fields - Completed
   b. Implement isTypeConstraints() marker method - Completed
   c. Implement Validate() checking starts_at is specified (either inline or named) and unit is specified - Completed

   **Intent**: Provide constraint validation for Interval types.
   **Acceptance Criteria**: IntervalConstraints validates correctly, requiring both constraints, and accepts either inline or named starts_at.

2. Enhance CreateConstraints factory for Interval type with semantic validation - Completed
   a. Add IntervalType case to CreateConstraints switch - Completed
   b. Implement createIntervalConstraints helper that handles both named type references and inline Instant constraints - Completed
   c. Check for duplicate constraint names in parserConstraints - Completed
   d. Validate starts_at (inline or named) and unit constraints are present - Completed

   **Intent**: Convert parser constraints to analysis constraints for Interval types with full semantic validation.
   **Acceptance Criteria**: Interval parser constraints are correctly transformed, duplicates detected, and presence validated.

3. Implement ValidateIntervalTypes cross-reference function - Completed
   a. Create ValidateIntervalTypes(types) function - Completed
   b. Verify starts_at type exists in Types lookup map - Completed
   c. Verify starts_at type is Instant or derived from Instant - Completed
   d. Return appropriate errors for invalid references - Completed

   **Intent**: Validate Interval type references to ensure starts_at type is a valid Instant or derived from Instant.
   **Acceptance Criteria**: Interval types with invalid starts_at references are detected and reported.

4. Integrate ValidateIntervalTypes into analysis flow - Completed
   a. Call ValidateIntervalTypes after all types are analyzed - Completed
   b. Ensure units are analyzed before unit reference validation (if applicable) - Completed

   **Intent**: Wire Interval cross-reference validation into the analysis pipeline.
   **Acceptance Criteria**: Interval validation runs as part of standard analysis flow.

5. Update analysis handlers to handle Interval types - Completed
   a. Verify EnterAnalyzeTypesScope handles Interval type via CreateConstraints - Completed
   b. Add any type-specific processing if needed - Completed

   **Intent**: Ensure the analyzer processes Interval type definitions correctly.
   **Acceptance Criteria**: Interval types are analyzed and stored in the Types lookup map.

   **Pattern**: Handlers use function-based pattern: `EnterAnalyzeTypesScope(table, context, analyzer)`.

6. Write comprehensive unit tests for Interval analysis - Completed
   a. Test IntervalConstraints.Validate() with both constraints present - Completed
   b. Test IntervalConstraints.Validate() with missing starts_at constraint - Completed
   c. Test IntervalConstraints.Validate() with missing unit constraint - Completed
   d. Test ValidateIntervalTypes with valid Instant-derived starts_at type - Completed
   e. Test ValidateIntervalTypes with non-existent starts_at type - Completed
   f. Test ValidateIntervalTypes with non-Instant-derived starts_at type - Completed
   g. Test CreateConstraints with IntervalType - Completed
   h. Test AnalyzeTypesScope with Interval type definitions - Completed

   **Intent**: Verify Interval analysis functionality works correctly including cross-reference validation.
   **Acceptance Criteria**: All tests pass, covering valid inputs and error conditions.

## Notes

- This story implements Increment 5 (final) of the types analysis phase as defined in ERO002
- Interval constraints require: starts_at (either a named Instant-derived type or an inline Instant constraint) and unit (unit name)
- Cross-reference validation ensures starts_at type exists and is Instant or derived from Instant
- Unit reference validation may require units to be analyzed first
- This completes full type functionality for all five built-in types
- **Parsing/Analysis Separation**: Semantic validation (cross-reference validation, constraint presence checking, duplicate detection) happens in this analysis story, not during parsing (Story139)
- Handlers use function-based pattern: `EnterAnalyzeTypesScope(table, context, analyzer)` not method-based
- Follow Logical TDD Lifecycle: write tests before implementing code for each task
