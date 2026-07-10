# Story135: CPL Types Parser - Multiple Type Sections (Increment 3)

## Goal

Extend the CPL parser to support multiple types sections within the same CPL file. Each types section creates its own TypesScope symbol table, enabling type organization across multiple declarations. This is Increment 3 of Epic002 types parsing.

## References

- [Epic002_cpl-types-implementation.md](../epics/Epic002_cpl-types-implementation.md) - Parent epic for types implementation
- [ERO002_types-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO002_types-parsing-phase.md) - Parsing phase data structures for Increment 3
- [ERO002_types-incremental-implementation.md](../specifications/design/cpl/eros/ERO002_types-incremental-implementation.md) - Main incremental implementation ERO
- [Story131_cpl-types-parser-number.md](Story131_cpl-types-parser-number.md) - Parser increment 1 (dependency)

## Dependencies

- [Story131_cpl-types-parser-number.md](Story131_cpl-types-parser-number.md) - Must be completed (types section parsing)
- [Story196_cpl-types-string-boolean.md](Story196_cpl-types-string-boolean.md) - Should be completed (all base types)

## Dependent Stories

- [Story136_cpl-types-analyzer-multiple-sections.md](Story136_cpl-types-analyzer-multiple-sections.md) - Not Started
- [Story137_cpl-types-parser-instant.md](Story137_cpl-types-parser-instant.md) - Not Started

## Tasks

1. Verify parser already supports multiple types sections - Not Started
   a. Confirm typesSection rule allows multiple occurrences in statement rule - Not Started
   b. Confirm TypesListener creates new TypesScope for each types section - Not Started
   c. Document any changes needed (likely none for parser) - Not Started

   **Intent**: Determine if parser changes are needed for multiple type sections.
   **Acceptance Criteria**: Clear understanding of parser behavior with multiple types sections.

2. Implement any needed parser changes for multiple sections - Not Started
   a. If typesSection is already in statement rule as alternative, no changes needed - Not Started
   b. Verify TypesListener correctly pushes/pops scope for each section - Not Started
   c. Add tests if any changes are made - Not Started

   **Intent**: Ensure parser correctly handles multiple types sections in a single file.
   **Acceptance Criteria**: Parser processes files with multiple types sections without errors.

3. Write unit tests for multiple type sections parsing - Not Started
   a. Test file with two separate types sections - Not Started
   b. Test file with three or more types sections - Not Started
   c. Test that each section creates its own TypesScope in symbol table - Not Started
   d. Test type definitions in different sections are independently parsed - Not Started

   **Intent**: Verify multiple types sections are correctly parsed and scoped.
   **Acceptance Criteria**: All tests pass, demonstrating correct handling of multiple types sections.

## Notes

- This story implements Increment 3 of the types parsing phase as defined in ERO002
- The parser should already support multiple types sections via the statement rule
- Each types section creates a separate TypesScope in the symbol table tree
- **Parsing/Analysis Separation**: Cross-section duplicate detection happens in Story136 (analysis phase), not during parsing
- No Instant or Interval type support in this increment
- Follow Logical TDD Lifecycle: write tests before implementing code for each task
