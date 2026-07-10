# Story141: CPL Units Parser - Basic Unit Definitions (Superseded)

## Status: SUPERSEDED

**This story has been superseded.** The units-related code has been temporarily removed from the implementation because it was incompatible with the new types functionality. The units functionality will be reimplemented properly as part of Epic003: Units and Measures Implementation after the types implementation is complete.

## Original Goal

The original goal was to implement the parser for basic unit definitions including unitsSection grammar, unitDefinition rule, and listener methods that push ParserUnitDefinition to the symbol table. This was Increment 1 of Epic003: Units and Measures Implementation.

## Supersession Reason

The units implementation (CPLUnits.g4 and related code) was incompatible with the new types functionality (CPLTypes.g4). The units grammar file defined:

- `unitsSection: UNITS unitDeclaration+;`
- `unitDeclaration: indent identifier EQUALS typeReference | NEWLINE;`

This conflicted with the types implementation approach. Rather than trying to patch the old implementation, we removed it entirely so it can be reimplemented properly with full awareness of the types system.

## What Was Removed

1. **Grammar**: `src/chronocone_planning_language/src/main/antlr/CPLUnits.g4`
2. **Example files**: `src/chronocone_planning_language/examples/without_errors/units_only.cpl`
3. **AST files**:
   - `src/chronocone_planning_language/src/main/go/ast/cpl_units.go`
   - `src/chronocone_planning_language/src/main/go/ast/cpl_units_test.go`
4. **Parser tests**:
   - `src/chronocone_planning_language/src/main/go/parser/cpl_parser_units_test.go`
   - `src/chronocone_planning_language/src/main/go/parser/exit_unit_declaration_test.go`
5. **Main grammar updated**: `ChronoconePlanningLanguage.g4` no longer imports CPLUnits

## Future Implementation

The units functionality will be reimplemented after the types system is complete. The reimplementation should:

1. Be designed with awareness of the types system (CPLTypes.g4)
2. Follow the symbol table-driven parsing pattern
3. Use the proper listener organization conventions
4. Be compatible with both concepts and types

### Reimplementation Tasks (To Be Defined)

When reimplementing units, the following areas need to be addressed:

- Define units grammar rules that are compatible with the types system
- Design how units reference types and concepts
- Implement proper validation during parsing
- Write comprehensive tests
- Integrate with the existing concepts and types functionality

## References

- [Epic003_units-and-measures-implementation.md](../epics/Epic003_units-and-measures-implementation.md) - Parent epic for units and measures implementation
- [Story131_cpl-types-parser-number.md](Story131_cpl-types-parser-number.md) - Types implementation story
- [Story132_cpl-types-analyzer-number.md](Story132_cpl-types-analyzer-number.md) - Types analyzer story

## Story Metadata

- **Original Created**: 2024
- **Superseded**: 2026-04-07
- **Related Stories**: Story142, Story143, Story144, Story145, Story146, Story147, Story148, Story149, Story150
- **Epic**: Epic003_units-and-measures-implementation
