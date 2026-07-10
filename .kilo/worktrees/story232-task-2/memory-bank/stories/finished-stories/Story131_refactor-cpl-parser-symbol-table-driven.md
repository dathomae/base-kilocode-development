# Story131: Refactor CPL Parser to Symbol Table-Driven Pattern

## Goal

Refactor the CPL parser to follow the symbol table-driven parsing pattern, eliminating type switches in the identifier handler. The parser's responsibility is limited to building the symbol table tree; declaration assembly will be handled in future analysis phase stories. This refactoring establishes the foundation for a cleaner separation between parsing and analysis phases.

## References

- [Symbol Table-Driven Parsing Pattern](../specifications/patterns/symbol-table-driven-parsing.md)
- [Symbol Table Design](../specifications/design/cpl/go-parser-target/symbol-table-design.md)
- [ERO001: Concepts Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO001_concepts-parsing-phase.md)
- [CPL Architecture Introduction](../specifications/architecture/cpl/introduction.md)

## Dependencies

- [Story110_cpl-symbol-table-implementation.md](finished-stories/Story110_cpl-symbol-table-implementation.md) - Completed - Symbol table infrastructure must be in place
- [Story114_cpl-concepts-parser-basic.md](finished-stories/Story114_cpl-concepts-parser-basic.md) - Completed - Basic concepts parsing exists and will be refactored

## Dependent Stories

- Story117_cpl-concepts-parser-hierarchy.md - Not Started - Depends on refactored parser pattern
- Story121_cpl-concepts-parser-conditional.md - Not Started - Depends on refactored parser pattern
- Story123_cpl-concepts-parser-range.md - Not Started - Depends on refactored parser pattern
- Story125_cpl-concepts-parser-before-after.md - Not Started - Depends on refactored parser pattern
- Story128_cpl-concepts-parser-concept-range.md - Not Started - Depends on refactored parser pattern

## Tasks

1. Remove transient state from CPLListener - Completed
   a. Remove `currentUnit *ast.UnitDecl` field from parser_support.go - Completed
   b. Remove `currentConcept *ast.ParserConceptDefinition` field from parser_support.go - Completed
   c. Verify `symbolStack`, `document`, `logger`, and `errors` fields remain - Completed
   d. Update any constructor functions that initialize removed fields - Completed

2. Simplify ExitIdentifier in cpl_document_listener.go - Completed
   a. Remove type-switch-based parent context checking - Completed
   b. Implement simplified identifier extraction - Completed
   c. Store identifier temporarily for current scope if needed - Completed
   d. Verify no parent type switches remain - Completed

3. Update unit declaration entry/exit methods - Completed
   a. Refactor EnterUnitDeclaration to push UnitsScope onto symbol stack - Completed
   b. Refactor ExitUnitDeclaration to define symbol and pop scope - Completed
   c. Remove AST building logic from unit declaration handling - Completed
   d. Update unit declaration tests to verify symbol table structure - Completed

4. Update concept declaration entry/exit methods - Completed
   a. Refactor EnterConceptDefinition to push appropriate scope - Completed
   b. Refactor ExitConceptDefinition to define symbol and pop scope - Completed
   c. Remove ParserConceptDefinition building from concept listener - Completed
   d. Store concept data in symbol definition Data field - Completed

5. Update parser tests for symbol table-driven pattern - Completed
   a. Update cpl_concepts_listener_test.go to verify symbol table structure - Completed
   b. Update cpl_document_listener_test.go to verify symbol definitions - Completed
   c. Add tests for symbol scope hierarchy - Completed
   d. Verify all existing tests pass with new pattern - Completed

6. Update dependent stories to reference symbol table pattern - Completed
   a. Update Story117 - change ParserConceptDefinition references to symbol table - Completed
   b. Update Story121 - change ParserConceptDefinition references to symbol table - Completed
   c. Update Story123 - change ParserConceptDefinition references to symbol table - Completed
   d. Update Story125 - change ParserConceptDefinition references to symbol table - Completed
   e. Update Story128 - change ParserConceptDefinition references to symbol table - Completed

7. Update ERO001 Concepts Parsing Phase documentation - Completed
   a. Update Introduction to clarify symbol table as primary output - Completed
   b. Add notes that ParserConceptDefinition is stored in SymbolDefinition.Data - Completed
   c. Update examples to show symbol table structure - Completed
   d. Clarify separation between parsing (symbol table) and analysis (AST assembly) - Completed

## Constraints

- Must not break existing test functionality - all existing tests must continue to pass
- Must maintain backward compatibility with existing Parse() function behavior
- Must eliminate all type switches on parent context types in ExitIdentifier
- Must populate symbol table with all defined symbols and their source locations
- Follow the Logical TDD Lifecycle - write tests before implementation
- Dependent stories (117, 121, 123, 125, 128) must be updated to reference symbol table structures

## Intent

The current implementation in `cpl_document_listener.go` uses type switches in `ExitIdentifier` to determine context based on parent node types. This creates tight coupling and requires modifying `ExitIdentifier` for every new declaration type. The goal is to refactor the parser to follow the symbol table-driven parsing pattern, where the symbol table stack provides context instead of parent node type switches. This creates a cleaner separation between parsing (building the symbol table tree) and analysis (walking the symbol table to assemble declarations), making the parser more maintainable and extensible.

## Acceptance Criteria

1. No type switches in ExitIdentifier - the method does not contain any type switches on parent context types
2. No transient state in CPLListener - `currentUnit`, `currentConcept`, and similar fields are removed
3. Symbol table is populated - after parsing, the symbol table tree contains all defined symbols with proper types and source locations
4. Unit declarations are captured in symbol table with correct scope hierarchy
5. Concept declarations are captured in symbol table with data stored in SymbolDefinition.Data
6. All parser tests pass and verify symbol table structure correctly
7. Dependent stories (117, 121, 123, 125, 128) no longer reference `ParserConceptDefinition` as a standalone structure
8. ERO001 Concepts Parsing Phase document clarifies that `ParserConceptDefinition` is stored in `SymbolDefinition.Data`
9. The `document` field is retained to maintain document structure alongside the symbol table

## Notes

**Anti-Pattern Being Eliminated:**

The current code uses type switches in ExitIdentifier:

```go
// ANTI-PATTERN in current code
if l.currentUnit != nil {
    parent := ctx.GetParent()
    switch parent.(type) {
    case *antlr.UnitDeclarationContext:
        l.currentUnit.Name = text
    case *antlr.TypeReferenceContext:
        l.currentUnit.TypeName = text
    }
}
```

**Target Pattern:**

The refactored code uses the symbol table stack for context:

```go
// TARGET PATTERN
func (l *CPLListener) EnterUnitDeclaration(ctx *antlr.UnitDeclarationContext) {
    current, _ := l.symbolStack.Peek()
    unitTable := current.AddChild(symboltable.UnitsScope)
    l.symbolStack.Push(unitTable)
}

func (l *CPLListener) ExitUnitDeclaration(ctx *antlr.UnitDeclarationContext) {
    // Symbol was defined by scope-specific logic
    l.symbolStack.Pop()
}
```

**Files to Modify:**

| File                                                                                | Changes                                                  |
| ----------------------------------------------------------------------------------- | -------------------------------------------------------- |
| `src/chronocone_planning_language/src/main/go/parser/parser_support.go`             | Remove transient state fields                            |
| `src/chronocone_planning_language/src/main/go/parser/cpl_document_listener.go`      | Simplify ExitIdentifier, update unit declaration methods |
| `src/chronocone_planning_language/src/main/go/parser/cpl_concepts_listener.go`      | Update concept declaration methods                       |
| `src/chronocone_planning_language/src/main/go/parser/cpl_concepts_listener_test.go` | Update tests for symbol table verification               |

**Out of Scope:**

- Analysis phase implementation (walking symbol table to assemble declarations) - will be handled in Story115 and beyond
- New declaration types (connectors, metrics, aggregators, spreaders) - will be added in future stories
- Synthesis phase (code generation) - not part of parsing

**Pattern for Updating Dependent Stories:**

Replace references to `ParserConceptDefinition` struct fields with references to `SymbolDefinition.Data` field containing the relevant parser structures:

```go
// BEFORE (in dependent stories)
ParserConceptDefinition includes IndentationLevel

// AFTER (symbol table pattern)
SymbolTable includes IndentationLevel for concept scopes
```
