# Story120: CPL Concepts Analyzer - Multi-Section

## Goal

Implement cross-section duplicate detection in the CPL analyzer for multiple concepts sections. This story enables the analyzer to detect when the same concept is defined in multiple sections and report it as an error. No special merging is required - the existing WalkSymbolTableWithHandlers naturally visits all ConceptsScope tables. This is part of [Epic001: CPL Concepts Implementation](../../epics/Epic001_cpl-concepts-implementation.md) - Increment 2.

## References

- [Epic001: CPL Concepts Implementation](../../epics/Epic001_cpl-concepts-implementation.md)
- [ERO001: Concepts Analysis Phase](../../specifications/design/cpl/go-analysis/eros/ERO001_concepts-analysis-phase.md) - Increment 3 section
- [Concept Definitions Design](../../specifications/design/cpl/concepts-definitions.md)

## Dependencies

- [Story119_cpl-concepts-parser-multi-section.md](../Story119_cpl-concepts-parser-multi-section.md) - Not Started - Parser must support multiple sections

## Dependent Stories

- Story121_cpl-concepts-parser-conditional.md - Not Started - Conditional builds on multi-section

## Tasks

1. Verify existing duplicate detection works across sections - Completed
   a. Test AnalyzeConceptsScope catches duplicates from different sections - Completed
   b. Verify symbol table walk visits all ConceptsScope tables - Completed

2. Enhance error messages for cross-section duplicates - Completed
   a. Include section information in error messages - Completed
   b. Report source locations from both definitions - Completed

3. Add concept extension detection (optional) - Deferred to later increment
   a. Distinguish between duplicate and extension intent - Deferred
   b. Track extension semantics for later increments - Deferred

4. Write unit tests for multi-section analysis - Completed
   a. Test duplicate detection across sections - Completed
   b. Test valid multi-section scenarios - Completed
   c. Test error message quality - Completed

## Constraints

- Must use existing WalkSymbolTableWithHandlers infrastructure
- Duplicate detection must work across all ConceptsScope tables
- Follow the Logical TDD Lifecycle - write tests before implementation
- Must not break Increment 1 or 2 functionality

## Intent

The goal of this story is to ensure the analyzer correctly handles multiple concepts sections by detecting duplicate concept definitions across sections. The existing AnalyzeConceptsScope duplicate check from Increment 1 is sufficient because the symbol table walk visits all ConceptsScope tables regardless of which section they came from.

## Acceptance Criteria

1. Duplicate concept definitions across multiple sections are detected as errors
2. Error messages include source location information from both definitions
3. Valid multi-section scenarios (different concepts in each section) work correctly
4. Hierarchy is preserved when concepts span multiple sections
5. All unit tests pass following BTDD practices
6. Increment 1 and 2 functionality still works

## Notes

**Increment 3 Analysis Changes:**

Unlike what was originally planned, there is no need for a separate BuildConceptsFromMultipleSections or mergeConceptDefinitions function. The symbol table walk naturally visits each ConceptsScope table, and the AnalyzeConceptsScope handler accumulates all concept definitions into the analyzer's concepts map.

**Duplicate Detection:**

The existing duplicate check in AnalyzeConceptsScope is sufficient:

```go
if _, exists := ca.concepts[*def.Name]; exists {
    return fmt.Errorf("duplicate concept definition: %s", *def.Name)
}
```

Since the symbol table walk visits all ConceptsScope tables (regardless of which section they came from), any attempt to define the same concept name in multiple sections will be caught.

**Increment 3 Capabilities:**

- Cross-section duplicate detection
- Error reporting with source locations
- Support for valid multi-section organization

**Increment 3 Limitations:**

- No conditional specifications (added in Increment 4)
- No Range generators (added in Increment 5)
- No before/after clauses (added in Increment 6)
- No API integration (added in Increment 7)

**Example Error:**

```
Error: duplicate concept definition: Country
  First definition at: line 2, column 5
  Second definition at: line 15, column 5
```

**Related Documents:**

- [Concepts Incremental Implementation](../../specifications/design/cpl/eros/ERO001_concepts-incremental-implementation.md) - Main incremental implementation ERO
