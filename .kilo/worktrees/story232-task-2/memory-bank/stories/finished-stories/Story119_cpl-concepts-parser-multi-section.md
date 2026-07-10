# Story119: CPL Concepts Parser - Multi-Section

## Goal

Implement support for multiple concepts sections in the same CPL file. This story allows concept definitions to be spread across multiple `concepts` blocks, enabling better organization and modularity of concept declarations. Each section creates its own ConceptsScope symbol table. This is part of [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md) - Increment 2.

## References

- [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md)
- [ERO001: Concepts Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO001_concepts-parsing-phase.md) - Increment 3 section
- [Concept Definitions Design](../specifications/design/cpl/concepts-definitions.md)

## Dependencies

- [Story118_cpl-concepts-analyzer-hierarchy.md](Story118_cpl-concepts-analyzer-hierarchy.md) - Not Started - Hierarchy must be in place

## Dependent Stories

- Story120_cpl-concepts-analyzer-multi-section.md - Not Started - Analyzer multi-section depends on parser

## Tasks

1. Update CPLConcepts.g4 for multiple sections - Completed
   a. Modify grammar to allow conceptsSection+ - Completed (grammar already supports multiple sections via `document: (indent | statement)* EOF`)
   b. Ensure each section can have its own INDENT/DEDENT block - Completed

2. Update ChronoconePlanningLanguage.g4 - Completed
   a. Modify statement rule to allow multiple conceptsSection - Completed (statement rule already allows multiple sections via document repetition)

3. Enhance ConceptsListener - Completed
   a. Handle multiple EnterConceptsSection calls - Completed (listener correctly handles multiple sections)
   b. Create separate ConceptsScope for each section - Completed
   c. Maintain proper scope hierarchy - Completed

4. Update symbol table handling - Completed
   a. Ensure multiple ConceptsScope tables are siblings - Completed (each section creates a sibling ConceptsScope under GlobalScope)
   b. Each section pushes/pops its own scope - Completed

5. Write unit tests for multi-section parsing - Completed
   a. Test two concepts sections in same file - Completed
   b. Test concepts in different sections with hierarchy - Completed
   c. Test empty concepts section - Completed
   d. Test error cases - Completed

## Constraints

- Must integrate with existing ANTLR-based parser infrastructure
- Must not break Increment 1 or 2 functionality
- Cross-section duplicate detection happens in analysis phase, not parsing
- Follow the Logical TDD Lifecycle - write tests before implementation

## Intent

The goal of this story is to enable multiple `concepts` sections in a single CPL file. This allows developers to organize their concept definitions more flexibly, potentially grouping related concepts together or separating base concepts from extensions. The parser creates separate ConceptsScope tables for each section; the analyzer handles merging and duplicate detection.

## Acceptance Criteria

1. Parser can successfully parse multiple concepts sections in the same file
2. Each concepts section creates its own ConceptsScope symbol table
3. Multiple ConceptsScope tables are siblings under GlobalScope
4. Hierarchy within each section is preserved
5. ConceptsListener correctly handles multiple section entries
6. All unit tests pass following BTDD practices
7. Increment 1 and 2 functionality still works

## Notes

**Increment 3 Capabilities:**

- Multiple concepts sections in the same file
- Each section has its own scope
- Hierarchical concepts can span sections (with limitations)

**Increment 3 Limitations:**

- No conditional specifications (added in Increment 4)
- No Range generators (added in Increment 5)
- No before/after clauses (added in Increment 6)
- Cross-section duplicate detection in analysis phase

**Example CPL code that should parse:**

```cpl
concepts
    Country=[Values[USA, Canada, Mexico]]
      Region=[Values[North, South]]

# Other CPL content...

concepts
    Priority=[Values[Low, Medium, High]]
    Status=[Values[Active, Inactive]]
```

**Symbol Table Structure:**

```
GlobalScope (root)
├── ConceptsScope (first section)
│   ├── SymbolDefinition: Country
│   └── SymbolDefinition: Region
├── ConceptsScope (second section)
│   ├── SymbolDefinition: Priority
│   └── SymbolDefinition: Status
```

**Design Decision:**

No special merging logic during parsing. Each `concepts` section creates a new ConceptsScope table, and these tables are siblings under the GlobalScope. The symbol table naturally accumulates all definitions. Duplicate concept detection happens during the analysis phase when walking the symbol table tree.

**Related Documents:**

- [Concepts Incremental Implementation](../specifications/design/cpl/eros/ERO001_concepts-incremental-implementation.md) - Main incremental implementation ERO
