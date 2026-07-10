# Story197: CPL Types Multiple Sections - Missing Tests and Verification

## Goal

Add the missing parser and analyzer tests for multiple types sections and cross-section duplicate detection. The underlying grammar, listener, and analyzer logic already supports these features (implemented as part of Epic002). This story verifies the existing behavior and fills the test coverage gaps for Story135 and Story136.

## References

- [Epic002_cpl-types-implementation.md](../epics/Epic002_cpl-types-implementation.md) - Parent epic for types implementation
- [ERO002_types-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO002_types-parsing-phase.md) - Parsing phase data structures
- [ERO002_types-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md) - Analysis phase data structures

## Dependencies

- [Story131_cpl-types-parser-number.md](finished-stories/Story131_cpl-types-parser-number.md) - Completed (base types parsing)
- [Story132_cpl-types-analyzer-number.md](finished-stories/Story132_cpl-types-analyzer-number.md) - Completed (base types analysis)
- [Story133_cpl-types-parser-string-boolean.md](Story133_cpl-types-parser-string-boolean.md) - Must be completed (String/Boolean types)
- [Story134_cpl-types-analyzer-string-boolean.md](Story134_cpl-types-analyzer-string-boolean.md) - Must be completed (String/Boolean analysis)

## Dependent Stories

None

## Tasks

1. Add parser tests for multiple types sections - Completed
   a. Test file with two separate types sections produces two TypesScope children - Completed
   b. Test file with three or more types sections produces matching TypesScope children - Completed
   c. Test that each TypesScope contains its own TypeScope children - Completed
   d. Test mixed base types across sections (Number in first, String in second, Boolean in third) - Completed

   **Intent**: Verify the existing parser grammar and listener correctly handle multiple types sections.
   **Acceptance Criteria**: All parser tests pass, demonstrating that multiple types sections create independent TypesScope entries under the root document scope.

2. Add analyzer tests for cross-section duplicate detection - Completed
   a. Test duplicate type name across two sections produces error - Completed
   b. Test unique type names across multiple sections succeed - Completed
   c. Test three sections with one duplicate errors on the duplicate name - Completed
   d. Test symbol table walk visits all TypeScope tables under all TypesScope tables - Completed

   **Intent**: Verify the existing EnterAnalyzeTypeScope duplicate check works across multiple TypesScope tables.
   **Acceptance Criteria**: All analyzer tests pass, demonstrating correct duplicate detection and successful analysis for valid multi-section scenarios.

3. Add integration tests for mixed type definitions across sections - Completed
   a. Test parse+analyze with Number types in first section, String types in second - Completed
   b. Test parse+analyze with Boolean type in separate section - Completed
   c. Test parse+analyze with Instant and Interval types across sections - Completed
   d. Verify all types are collected into the analyzer's single Types lookup map - Completed

   **Intent**: Verify end-to-end parsing and analysis correctly processes mixed type definitions across multiple sections.
   **Acceptance Criteria**: All integration tests pass, confirming all types from all sections are present in the final AnalysisDefinitions.

## Notes

- This story only adds tests; no new parser or analyzer logic is required.
- The parser grammar (document rule with statement\*), listener (EnterTypesSection), and analyzer (EnterAnalyzeTypeScope duplicate check) already implement the required behavior.
- Follow Logical TDD Lifecycle: write tests first, then verify they pass against existing code.
- Test files to update:
  - `src/chronocone_planning_language/src/main/go/parser/cpl_parser_types_test.go`
  - `src/chronocone_planning_language/src/main/go/analysis/types_analysis_test.go`
