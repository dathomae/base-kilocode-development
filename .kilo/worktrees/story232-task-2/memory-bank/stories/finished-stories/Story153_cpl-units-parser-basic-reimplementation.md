# Story153: CPL Units Parser - Basic Unit Definitions Reimplementation

## Title

CPL Units Parser - Basic Unit Definitions Reimplementation (Replaces Superseded Story141)

## Goal

Reimplement the parser for basic unit definitions (Increment 1 of Epic003) with full awareness of the types system. This story replaces the superseded Story141, which was removed due to incompatibility with the types implementation. The reimplementation must follow the symbol table-driven parsing pattern and be compatible with both concepts and types.

## Status

Done

## Context

### Background

The original units implementation (Story103, superseded Story141-150) was removed from the codebase because it was incompatible with the new types functionality. The old implementation:

1. Used a separate `CPLUnits.g4` grammar file that conflicted with `CPLTypes.g4`
2. Defined `unitDeclaration: indent identifier EQUALS typeReference | NEWLINE;` which was semantically incorrect
3. Used an AST-building approach rather than symbol table-driven parsing
4. Did not account for the relationship between units and the types system

### What Was Removed

1. **Grammar**: `src/chronocone_planning_language/src/main/antlr/CPLUnits.g4`
2. **AST files**: `cpl_units.go`, `cpl_units_test.go`
3. **Parser tests**: `cpl_parser_units_test.go`, `exit_unit_declaration_test.go`
4. **Example files**: Units-only CPL examples

### Reimplementation Requirements

The new units implementation must:

1. **Be compatible with types**: Units grammar must coexist with `CPLTypes.g4` without conflicts
2. **Follow symbol table-driven parsing**: No AST building in the parser; use `ParserUnitDefinition` stored in `SymbolDefinition.Data`
3. **Use proper listener organization**: Follow the pattern established by concepts and types
4. **Reference concepts correctly**: Unit values come from concept values (`UnitGroup = ConceptName`)
5. **Integrate with the main grammar**: Use the `statement` rule in `ChronoconePlanningLanguage.g4`

## References

- [Epic003_units-and-measures-implementation.md](../epics/Epic003_units-and-measures-implementation.md) - Parent epic
- [ERO003_units-measures-incremental-implementation.md](../specifications/design/cpl/eros/ERO003_units-measures-incremental-implementation.md) - Incremental implementation ERO
- [ERO003_units-measures-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO003_units-measures-parsing-phase.md) - Parsing phase ERO
- [Story141_cpl-units-parser-basic.md](Story141_cpl-units-parser-basic.md) - Superseded story (for historical context)
- [Story131_cpl-types-parser-number.md](finished-stories/Story131_cpl-types-parser-number.md) - Types parser (completed)
- [Symbol Table-Driven Parsing Pattern](../specifications/patterns/symbol-table-driven-parsing.md) - Architectural pattern

## Dependencies

### Must Be Completed First

- [Story131_cpl-types-parser-number.md](finished-stories/Story131_cpl-types-parser-number.md) - Done - Types infrastructure must be in place
- [Story110_cpl-symbol-table-implementation.md](finished-stories/Story110_cpl-symbol-table-implementation.md) - Done - Symbol table infrastructure

### Coordinates With

- [Story132_cpl-types-analyzer-number.md](Story132_cpl-types-analyzer-number.md) - Types analysis (can proceed in parallel)

### Enables

- Story154_cpl-units-analyzer-basic.md - Analyzer increment 1 (to be created)
- Story155_cpl-measures-parser-basic.md - Parser increment 2 (to be created)

## Acceptance Criteria

1. **Grammar Integration**: `unitsSection` added to `ChronoconePlanningLanguage.g4` without conflicts with types grammar
2. **Lexer Tokens**: `UNITS_KEYWORD` added to `ChronoconePlanningLanguageLexer.g4`
3. **Symbol Table Types**: `UnitsScope` and `UnitScope` scope types defined
4. **Parser Types**: `ParserUnitDefinition` struct defined with `Name`, `ConceptName`, and `SourceLocation` fields
5. **Listener Implementation**: `EnterUnitsSection`, `ExitUnitsSection`, `EnterUnitDeclaration`, `ExitUnitDeclaration` methods implemented
6. **Symbol Table Population**: Parser pushes `ParserUnitDefinition` to symbol table via `SymbolDefinition.Data`
7. **Test Coverage**: Unit tests verify symbol table structure after parsing units sections
8. **Example Files**: At least one example CPL file with units section parses correctly

## Tasks

1. **Add Lexer Token** - Completed
   - Added `UNITS: 'units';` to `ChronoconePlanningLanguageLexer.g4`

2. **Update Main Grammar** - Completed
   - Added `unitsSection` to the `statement` rule in `ChronoconePlanningLanguage.g4`

3. **Define Grammar Rules** - Completed
   - Added `unitsSection: UNITS unitDeclaration+;` and `unitDeclaration: indent (identifier EQUALS identifier)*;` to CPLUnits.g4

4. **Add Scope Types** - Completed
   - Added `UnitsScope` and `UnitScope` to `ScopeType` enum in symbol_table.go

5. **Define Parser Types** - Completed
   - Unit definitions are stored directly as identifier Definitions in UnitScope (simpler pattern than ParserUnitDefinition)

6. **Implement Listener Methods** - Completed
   - All four listener methods implemented in cpl_units_listener.go

7. **Write Tests** - Completed
   - Comprehensive tests in cpl_parser_units_test.go covering single/multiple definitions, whitespace, error handling, and integration

8. **Create Example Files** - Completed
   - Created units-example.cpl with concepts and units sections

## Implementation Notes

### Grammar Structure

```antlr
// In ChronoconePlanningLanguage.g4
statement
    : conceptsSection
    | typesSection
    | unitsSection  // NEW
    ;

// Units section rules (can be inline or in separate file)
unitsSection: UNITS_KEYWORD unitDeclaration+;
unitDeclaration: identifier EQ identifier;
```

### Symbol Table Integration

```go
// ParserUnitDefinition - stored in SymbolDefinition.Data
type ParserUnitDefinition struct {
    Name           string
    ConceptName    string
    SourceLocation SourceLocation
}

// In listener ExitUnitDeclaration:
unitDef := &ParserUnitDefinition{
    Name:           unitName,
    ConceptName:    conceptName,
    SourceLocation: sourceLocation,
}

symDef := SymbolDefinition{
    Name:           &unitName,
    DefinitionType: UnitScopeDefinition,
    Location:       sourceLocation,
    Data:           unitDef,
}
```

### Key Differences from Superseded Story141

| Aspect           | Old (Superseded)                         | New (This Story)                   |
| ---------------- | ---------------------------------------- | ---------------------------------- |
| Grammar file     | Separate `CPLUnits.g4`                   | Integrated into main grammar       |
| Unit declaration | `indent identifier EQUALS typeReference` | `identifier EQ identifier`         |
| AST building     | Direct `UnitDecl` nodes                  | Symbol table only                  |
| Types awareness  | None                                     | Designed with types system in mind |
| Pattern          | Old listener pattern                     | Symbol table-driven parsing        |

## Story Metadata

- **Created**: 2026-04-07
- **Replaces**: Story141_cpl-units-parser-basic.md (Superseded)
- **Epic**: Epic003_units-and-measures-implementation
- **Increment**: 1 (Basic Unit Definitions)
