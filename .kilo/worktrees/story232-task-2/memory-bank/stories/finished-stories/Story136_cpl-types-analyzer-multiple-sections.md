# Story136: CPL Types Analyzer - Multiple Type Sections (Increment 3)

## Goal

Implement cross-section duplicate detection during symbol table walk for multiple types sections. The existing AnalyzeTypesScope handler already processes all TypesScope tables during WalkSymbolTableWithHandlers, so this story focuses on verifying and testing the duplicate detection behavior. This is Increment 3 of Epic002 types analysis.

## References

- [Epic002_cpl-types-implementation.md](../epics/Epic002_cpl-types-implementation.md) - Parent epic for types implementation
- [ERO002_types-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md) - Analysis phase data structures for Increment 3
- [ERO002_types-incremental-implementation.md](../specifications/design/cpl/eros/ERO002_types-incremental-implementation.md) - Main incremental implementation ERO
- [Story132_cpl-types-analyzer-number.md](Story132_cpl-types-analyzer-number.md) - Analyzer increment 1 (dependency)
- [Story135_cpl-types-parser-multiple-sections.md](Story135_cpl-types-parser-multiple-sections.md) - Parser increment 3 (dependency)

## Dependencies

- [Story132_cpl-types-analyzer-number.md](Story132_cpl-types-analyzer-number.md) - Must be completed (base analysis infrastructure)
- [Story135_cpl-types-parser-multiple-sections.md](Story135_cpl-types-parser-multiple-sections.md) - Must be completed (parser support for multiple sections)

## Dependent Stories

- [Story137_cpl-types-parser-instant.md](Story137_cpl-types-parser-instant.md) - Not Started
- [Story138_cpl-types-analyzer-instant.md](Story138_cpl-types-analyzer-instant.md) - Not Started

## Tasks

1. Verify existing duplicate detection handles multiple sections - Not Started
   a. Confirm EnterAnalyzeTypesScope is called for each TypesScope during symbol table walk - Not Started
   b. Confirm duplicate check (Types map) catches duplicates across sections - Not Started
   c. Document that no code changes are needed - Not Started

   **Intent**: Verify existing analyzer logic correctly handles multiple types sections.
   **Acceptance Criteria**: Clear understanding that existing duplicate detection works across sections.

   **Pattern**: Handlers use function-based pattern: `EnterAnalyzeTypesScope(table, context, analyzer)`.

2. Write unit tests for cross-section duplicate detection - Not Started
   a. Test duplicate type name in two different types sections (should error) - Not Started
   b. Test unique type names across multiple sections (should succeed) - Not Started
   c. Test three sections with one duplicate (should error on duplicate) - Not Started
   d. Test symbol table walk visits all TypesScope tables - Not Started

   **Intent**: Verify cross-section duplicate detection works correctly.
   **Acceptance Criteria**: All tests pass, demonstrating correct duplicate detection across sections.

3. Test integration with mixed type definitions across sections - Not Started
   a. Test file with Number types in first section, String types in second - Not Started
   b. Test file with Boolean type in separate section - Not Started
   c. Verify all types are collected into single Types lookup map - Not Started

   **Intent**: Verify analyzer correctly processes mixed type definitions across multiple sections.
   **Acceptance Criteria**: All types from all sections are analyzed and stored in the Types map.

## Notes

- This story implements Increment 3 of the types analysis phase as defined in ERO002
- No special merging function is needed; the symbol table walk naturally visits all TypesScope tables
- The existing duplicate check in EnterAnalyzeTypesScope catches cross-section duplicates
- No Instant or Interval type support in this increment
- **Parsing/Analysis Separation**: Cross-section duplicate detection happens in this analysis story, not during parsing (Story135)
- Handlers use function-based pattern: `EnterAnalyzeTypesScope(table, context, analyzer)` not method-based
- Follow Logical TDD Lifecycle: write tests before implementing code for each task
