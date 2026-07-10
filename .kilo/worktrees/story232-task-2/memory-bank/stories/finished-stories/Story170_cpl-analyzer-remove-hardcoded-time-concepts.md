# Story170: CPL Analyzer Remove Hardcoded Time Concepts

## Goal

Remove the hardcoded standard time concept values from the CPL analyzer and rely solely on the `standard_concepts.cpl` file for Instant constraint validation. This story is part of [Epic002: CPL Types Implementation](../epics/Epic002_cpl-types-implementation.md).

The analyzer currently falls back to English month and day names baked into the compiler binary via `buildStandardTimeConcepts()`. This contradicts the architectural principle that concept values should be determined at CPL compile time, not fixed in the compiler binary. The compiler already loads `standard_concepts.cpl` automatically during the parse phase, so the hardcoded Go fallback is redundant and prevents localization or customization of time concept values.

## References

- [ERO002: Types Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md)
- [Requirements: Talking About Time](../specifications/requirements/cpl/talking-about-time.md)
- [Requirements: Concepts, Types, Units, Measures Syntax](../specifications/requirements/cpl/concepts-types-units-measures-syntax.md)
- [Standard Concepts CPL](../../src/chronocone_planning_language/standard_definitions/standard_concepts.cpl)

## Dependencies

- [Story168_cpl-analyzer-instant-constraints-concept-driven.md](../Story168_cpl-analyzer-instant-constraints-concept-driven.md) - Done
- [Story169_cpl-standard-definitions-consistency.md](Story169_cpl-standard-definitions-consistency.md) - Done

## Dependent Stories

None

## Tasks

1. Create test helper for loading standard concepts from CPL file - Completed
   a. Write tests for the new test helper that loads `standard_concepts.cpl` and returns a concepts map - Completed
   b. Implement a test helper function that parses `standard_concepts.cpl` and produces a `map[string]AnalysisConceptDefinition` - Completed
   c. Verify the helper produces identical concept definitions to the current `buildStandardTimeConcepts()` output - Completed

2. Remove `buildStandardTimeConcepts()` from `cpl_types.go` - Completed
   a. Write tests confirming that `CreateConstraints` returns an error when required concepts are missing from the concepts map - Completed
   b. Delete the `buildStandardTimeConcepts()` function and `mustBuildIntRangeConcept()` helper from `cpl_types.go` - Completed
   c. Update any tests in `cpl_types_internal_test.go` that directly call `buildStandardTimeConcepts()` to use the new test helper - Completed

3. Remove fallback concept injection from `Analyze()` - Completed
   a. Write tests confirming that analysis fails with a clear error when a CPL file uses Instant constraints but standard concepts are not available - Completed
   b. Remove the fallback injection loop in `Analyze()` that calls `buildStandardTimeConcepts()` - Completed
   c. Verify that `Analyze()` still succeeds when concepts are provided via the parsed symbol table - Completed

4. Remove fallback concept injection from `ComputeDerivationPaths()` - Completed
   a. Write tests confirming that `ComputeDerivationPaths()` fails with a clear error when required time concepts are missing - Completed
   b. Remove the fallback injection loop in `ComputeDerivationPaths()` that calls `buildStandardTimeConcepts()` - Completed
   c. Verify that type derivation with Instant constraints still works when concepts are present in the analyzer's concepts map - Completed

5. Update `types_analysis_test.go` to use CPL-based test concepts - Completed
   a. Identify all tests that call `buildStandardTimeConcepts()` directly - Completed
   b. Replace direct calls with the new test helper or an equivalent programmatic construction - Completed
   c. Ensure all existing test assertions continue to pass - Completed

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- The constraint-to-concept name mapping (`instantConstraintToConceptName`) must remain unchanged; only the fallback value injection is being removed.
- The `standard_concepts.cpl` file must continue to define all required time concepts (`Year`, `Month`, `DayOfMonth`, `DayOfYear`, `WeekNumber`, `FiscalQuarter`, `DayOfWeek`, `Hour`, `Minute`, `Second`).
- The compiler CLI already prepends standard definition files to user inputs during `parsePhase()`, so normal compilation flows should be unaffected.
- Missing required concepts must produce a clear, actionable error message indicating which concept is missing.

## Intent

- **Task 1**: To provide tests with a clean, maintainable way to obtain standard time concepts without relying on hardcoded Go values.
- **Task 2**: To eliminate the source of hardcoded English month and day names from the compiler binary.
- **Task 3**: To ensure the analysis phase no longer silently injects binary-baked concept values when a CPL file omits time concept definitions.
- **Task 4**: To ensure type derivation does not silently inject binary-baked concept values.
- **Task 5**: To bring all analyzer tests into compliance with the new no-hardcoded-defaults policy.

## Acceptance Criteria

- `buildStandardTimeConcepts()` no longer exists in the codebase.
- No fallback injection of hardcoded time concepts occurs in `Analyze()` or `ComputeDerivationPaths()`.
- When required time concepts are missing, the analyzer reports an error naming the missing concept.
- All existing tests pass without calling `buildStandardTimeConcepts()`.
- The compiler CLI continues to load `standard_concepts.cpl` automatically and uses those values for Instant constraint validation.
- A user can customize time concept values (e.g., localization) by providing a different `standard_concepts.cpl` or by defining concepts in their CPL source, without recompiling the compiler.

## Notes

- The standard definitions directory (`standard_definitions/`) is discovered by the compiler CLI and prepended to input paths. This mechanism is the correct source of built-in concept values and is already implemented.
- The override behavior is preserved: if a CPL file defines its own `Month` or `DayOfWeek` concepts, those definitions take precedence over the standard definitions.
- This change moves the system closer to ERO001 Increment 8 (API Integration), where concepts may ultimately be loaded from an external Concepts API rather than local files.
