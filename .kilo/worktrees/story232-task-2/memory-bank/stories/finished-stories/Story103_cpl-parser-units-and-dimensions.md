# Story 103: CPL Parser - Units and Dimensions

## Overview

Extend the CPL parser to support `units` and `dimensions` declaration sections. These sections define the fundamental types and dimensions used throughout CPL documents.

## References

- [CPL Units and Dimensions Requirements](../specifications/requirements/cpl/units-and-dimensions.md) - Complete specification for unit and dimension declarations
- [CPL Architecture - Introduction](../specifications/architecture/cpl/introduction.md) - Overview of four-phase CPL lifecycle
- [CPL Architecture - Parsing Units and Dimensions](../specifications/architecture/cpl/parsing/units-and-dimensions.md) - Parsing phase details and tokenization decisions
- [Story102_cpl-parser-environment-setup.md](Story102_cpl-parser-environment-setup.md) - Foundation parser environment

## Dependencies

### Prerequisites

- [Story102_cpl-parser-environment-setup.md](Story102_cpl-parser-environment-setup.md) - Not Started (requires parser environment and build system)

### Dependents

- [Story104_cpl-parser-connectors.md](Story104_cpl-parser-connectors.md) - Not Started

## Tasks

1. Extend grammar for unit declarations - **Completed**
   a. Update grammar to include `units` section keyword
   b. Define parser rule for unit declarations (e.g., `unit = Count`)
   c. Support built-in unit types: `String`, `Number`, `Count`, `Currency`, and time units (`Second`, `Minute`, `Hour`, `Day`, `Week`, `Month`, `Year`)—treated as identifiers (not keywords) per architecture
   d. Support optional type parameters in parentheses (e.g., `Number(decimal_separator: '.')`, `Currency(symbol: '$')`)
   e. Generate updated Go parser code
   f. Write unit tests for unit declaration parsing including parameterized types

2. Extend grammar for dimension declarations - **Completed**
   a. Update grammar to include `dimensions` section keyword
   b. Define parser rule for dimension declarations (e.g., `price = price[currency]`)
   c. Support dimension references with optional bracket notation for type parameters
   d. Generate updated Go parser code
   e. Write unit tests for dimension declaration parsing

3. Create AST types for units and dimensions - **Completed**
   a. Define `UnitDecl` struct with `name`, `type_name`, and optional `parameters` fields
   b. Define `DimensionDecl` struct with `name` and `unit_name` fields
   c. Create `ast/units.go` and `ast/dimensions.go` files
   d. Write unit tests for AST type construction

4. Implement listener methods for units and dimensions - **Completed**
   a. Create `CPLListener` struct implementing ANTLR listener interface
   b. Implement `EnterUnitsSection` and `ExitUnitsSection` methods
   c. Implement `EnterUnitDecl` and `ExitUnitDecl` methods
   d. Implement `EnterDimensionsSection` and `ExitDimensionsSection` methods
   e. Implement `EnterDimensionDecl` and `ExitDimensionDecl` methods
   f. Write unit tests for listener functionality

5. Update parser wrapper - **Completed**
   a. Extend `Parse(input string)` to return AST root containing units and dimensions
   b. Create `CPLDocument` struct to hold parsed declarations
   c. Note: Type name validation is deferred to analysis phase per architecture—parser stores type names as raw strings only
   d. Write unit tests for complete document parsing

6. Create test CPL files - **Completed**
   a. Create `examples/units_only.cpl` with various unit declarations including parameterized types
   b. Create `examples/dimensions_only.cpl` with various dimension declarations using both built-in types and user-defined units
   c. Create `examples/units_and_dimensions.cpl` combining both sections, demonstrating that sections need not be defined before use
   d. Create `examples/multiple_sections.cpl` demonstrating multiple units/dimensions sections
   e. Write integration tests parsing each example file

7. Clean up hello world artifacts - **Completed**
   a. Remove `examples/hello.cpl` or rename to `examples/hello_world.cpl` with deprecation note
   b. Remove or rename `src/test/go/hello_grammar_test.go` to reflect new grammar name
   c. Remove or update `src/test/go/parser/hello_parser_test.go` to test new parser functionality
   d. Update any references to `HelloCPL` in documentation

## Decisions

1. **Grammar evolution**: The grammar file is renamed from `HelloCPL.g4` to `ChronoconePlanningLanguage.g4` in this story, replacing the minimal "hello world" version.

2. **Section ordering and multiplicity**: Units and dimensions sections may appear in any order, and each may appear multiple times within a document. When included multiple times, the declared units/dimensions are cumulative. Units and dimensions need not be defined before they are used.

3. **Type syntax**: Unit types follow a simple `name = type` pattern, while dimensions may include bracket notation for parameterized types like `price[currency]`.

4. **Parsing versus Analysis responsibilities**: The parser captures type names and unit references as raw strings without validation. Semantic validation (checking that type names are valid built-in types or user-defined units) is deferred to the analysis phase per the CPL architecture.

## Notes

- Unit declarations establish the base types for time series data (e.g., `unit = Count`, `lead_time = Day`).
- Built-in unit types include: `String`, `Number`, `Count`, `Currency`, and time units (`Second`, `Minute`, `Hour`, `Day`, `Week`, `Month`, `Year`).
- Unit types support optional parameters for presentation attributes (e.g., `Currency(symbol: '$')`).
- Dimension declarations define the axes along which metrics are organized (e.g., `country = country[String]`, `price = price[Currency]`).
- The unit name in dimension declarations can be either a built-in type or a user-defined unit from the `units` section.
- Per the architecture, type name validation occurs during analysis phase, not parsing—the parser captures raw strings only.
