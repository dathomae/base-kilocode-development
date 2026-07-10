# Story117: CPL Concepts Parser - Hierarchy

## Goal

Implement concept hierarchy support in the CPL parser using indentation-based scoping. This story adds INDENT token handling and parent-child relationship parsing, enabling nested concept definitions like child concepts indented under parent concepts. This is part of [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md) - Increment 2.

## References

- [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md)
- [ERO001: Concepts Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO001_concepts-parsing-phase.md) - Increment 2 section
- [Concept Definitions Design](../specifications/design/cpl/concepts-definitions.md)
- [Symbol Table Design](../specifications/design/cpl/go-parser-target/symbol-table-design.md)

## Dependencies

- [Story114_cpl-concepts-parser-basic.md](../finished-stories/Story114_cpl-concepts-parser-basic.md) - Done - Basic parser is in place
- [Story115_cpl-concepts-analyzer-basic.md](../finished-stories/Story115_cpl-concepts-analyzer-basic.md) - Done - Basic analyzer must be in place

## Dependent Stories

- Story118_cpl-concepts-analyzer-hierarchy.md - Not Started - Analyzer hierarchy depends on parser hierarchy

## Tasks

1. Implement INDENT token handling in grammar - Done
   a. Add INDENT token to lexer grammar - Done
   b. Update CPLConcepts.g4 to use indent in conceptDefinition - Done

2. Update SymbolTable with indentation tracking - Done
   a. Add IndentedSpaceEquivalent field to SymbolTable - Done
   b. Capture indentation in ExitIndent listener - Done

3. Update AnalysisConceptDefinition for hierarchy - Done
   a. Add ParentConceptName field - Done
   b. Add IndentAsSpaces field - Done

4. Implement parent-child relationship detection - Done
   a. Compare indentation levels in AddConcept - Done
   b. Set ParentConceptName when concept is nested - Done

5. Write unit tests for hierarchy parsing - Done
   a. Test parent-child relationship parsing - Done
   b. Test multi-level nesting - Done
   c. Test nested concepts example file - Done

## Constraints

- Must integrate with existing ANTLR-based parser infrastructure
- Must support error reporting with source location information
- Follow the Logical TDD Lifecycle - write tests before implementation
- Indentation must be consistent (no mixing tabs and spaces)
- Must not break Increment 1 functionality

## Implementation Approach

**Note:** The implementation differs from the original plan. Instead of a lexer-based approach with CPLLexerBase and indentStack, the solution uses a **parser/analyzer-based approach**:

1. **Lexer**: Simple `INDENT: [\r\n][ \t]+` token matches newline followed by whitespace
2. **Parser**: `ExitIndent` listener captures indentation and stores it in `SymbolTable.IndentedSpaceEquivalent`
3. **Analyzer**: `AddConcept` compares indentation levels between consecutive concepts. When a concept has greater indentation than the previous concept, it is marked as a child:

```go
// From cpl_analyzer.go
if previousConcept.IndentAsSpaces < concept.IndentAsSpaces {
    // This concept is nested inside the previous concept
    concept.ParentConceptName = &previousConcept.ConceptName
}
```

This approach achieves the same functional result (hierarchical concept relationships) while keeping the lexer simple and handling the logic in the parser and analyzer phases.

## Acceptance Criteria

1. ✅ Parser can successfully parse nested concept definitions
2. ✅ SymbolTable includes IndentedSpaceEquivalent field for tracking indentation
3. ✅ AnalysisConceptDefinition contains ParentConceptName for hierarchy
4. ✅ Analyzer correctly establishes parent-child relationships based on indentation
5. ✅ All unit tests pass following BTDD practices
6. ✅ Increment 1 flat concept parsing still works

## Notes

**Increment 2 Capabilities:**

- Parent-child relationships via indentation
- Multi-level nesting (grandchildren, etc.)
- Indentation tracking in symbol table

**Increment 2 Limitations:**

- No multiple concept declaration sections (added in Increment 3)
- No conditional child values (added in Increment 4)
- No Range generators (added in Increment 5)
- No before/after clauses (added in Increment 6)

**Example CPL code that should parse:**

```cpl
concepts
    Country=[Values[USA, Canada, Mexico]]
      Region=[Values[North, South]]
      City=[Values[NYC, LA]]
        Neighborhood=[Values[Downtown, Uptown]]
```

**Related Documents:**

- [Concepts Incremental Implementation](../specifications/design/cpl/eros/ERO001_concepts-incremental-implementation.md) - Main incremental implementation ERO
