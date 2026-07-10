# Story114: CPL Concepts Parser - Basic

## Goal

Implement the basic concepts parsing functionality for the Chronocone Planning Language (CPL) parser. This story adds grammar support for the concepts section and Values generator, enabling the parser to recognize and parse simple concept definitions like `ConceptName=[Values[Value1, Value2]]`. This is part of [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md) - Increment 1.

## References

- [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md)
- [ERO001: Concepts Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO001_concepts-parsing-phase.md) - Increment 1 section
- [Concept Definitions Design](../specifications/design/cpl/concepts-definitions.md)
- [Symbol Table Design](../specifications/design/cpl/go-parser-target/symbol-table-design.md)

## Dependencies

- [Story110_cpl-symbol-table-implementation.md](../stories/finished-stories/Story110_cpl-symbol-table-implementation.md) - Completed - Symbol table infrastructure must be in place

## Dependent Stories

- Story115_cpl-concepts-analyzer-basic.md - Not Started - Analyzer depends on parser structures

## Tasks

1. Create CPLConcepts.g4 grammar file - Completed
   a. Define conceptsSection rule - Completed
   b. Define conceptDefinition rule for Values generator - Completed
   c. Define conceptValueList rule - Completed

2. Update ChronoconePlanningLanguage.g4 to import concepts grammar - Completed
   a. Add conceptsSection to statement rule - Completed
   b. Add import for CPLConcepts - Completed

3. Add lexer tokens for concepts - Completed
   a. Add CONCEPTS_KEYWORD token - Completed
   b. Add VALUES token - Completed
   c. Add LBRACKET, RBRACKET, EQ, COMMA tokens - Completed

4. Define ParserConceptDefinition struct - Completed
   a. Add Name field - Completed
   b. Add Generators field ([]Generator) - Completed
   c. Add SourceLocation field - Completed

5. Define GeneratorType enumeration - Completed
   a. Add ValuesGenerator constant - Completed

6. Add ScopeType for concepts - Completed
   a. Add ConceptsScope ScopeType - Completed
   b. Add ConceptsScopeDefinition DefinitionType - Completed

7. Implement ConceptsListener - Completed
   a. Handle EnterConceptsSection - push ConceptsScope - Completed
   b. Handle concept definition parsing - create ParserConceptDefinition - Completed
   c. Store concept definitions in symbol table - Completed

8. Write unit tests for concepts parsing - Completed
   a. Test simple concept with Values generator - Completed
   b. Test concept with multiple values - Completed
   c. Test duplicate concept detection - Completed
   d. Test error cases - Completed

## Constraints

- Must integrate with existing ANTLR-based parser infrastructure
- Must populate the symbol table for use by subsequent phases
- Must support error reporting with source location information
- Follow the Logical TDD Lifecycle - write tests before implementation
- Each increment must produce a working system with documented limitations

## Intent

The goal of this story is to establish the foundation for concept parsing in CPL. The parser should be able to recognize concept definitions using the Values generator syntax and store them in the symbol table for later analysis. This is the first step in implementing the full concepts functionality.

## Acceptance Criteria

1. CPLConcepts.g4 grammar file exists with proper rules for concepts section
2. Parser can successfully parse concept definitions like `ConceptName=[Values[Value1, Value2]]`
3. ParserConceptDefinition struct is defined with all required fields
4. GeneratorType enumeration includes ValuesGenerator
5. ConceptsListener correctly populates symbol table with concept definitions
6. All unit tests pass following BTDD practices
7. Duplicate concept definitions in the same file are detected and reported as errors

## Notes

**Increment 1 Limitations:**

- No parent-child relationship tracking (added in Increment 2)
- Indentation level is not stored (added in Increment 2)
- No multiple concept declaration sections in the same file (added in Increment 3)
- No Range generator support (added in Increment 5)
- No conditional specifications (added in Increment 4)
- No before/after clauses for value ordering (added in Increment 6)

**Example CPL code that should parse:**

```cpl
concepts
    Country=[Values[USA, Canada, Mexico]]
    Priority=[Values[Low, Medium, High]]
```

**Related Documents:**

- [Concepts Incremental Implementation](../specifications/design/cpl/eros/ERO001_concepts-incremental-implementation.md) - Main incremental implementation ERO
