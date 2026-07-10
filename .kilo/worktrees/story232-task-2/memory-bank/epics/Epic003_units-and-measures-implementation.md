# Epic003: Units and Measures Implementation

## Title

Units and Measures Implementation

## Ultimate Goal

Implement complete units and measures support in the CPL compiler, including basic unit definitions (Increment 1), basic measure definitions (Increment 2), compound units and unit algebra (Increment 3), unit and measure conversions with pattern/calculation pairs (Increments 4–5 combined), synthesis-time unit type checking, and automatic conversion injection. The parser captures definitions into symbol tables using postfix expression parsing; the analysis phase resolves references and builds expression trees; the synthesis phase generates conversion code and verifies unit type consistency across the full CIL expression tree.

## References

- [ERO003 Units and Measures Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO003_units-measures-parsing-phase.md)
- [ERO003 Units and Measures Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO003_units-measures-analysis-phase.md)
- [ERO003 Units and Measures Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO003_units-measures-synthesis-phase.md)
- [ERO003 Units and Measures Incremental Implementation](../specifications/design/cpl/eros/ERO003_units-measures-incremental-implementation.md)
- [Postfix Expression Parsing Pattern](../specifications/patterns/postfix-expression-parsing.md)
- [Symbol Table-Driven Parsing Pattern](../specifications/patterns/symbol-table-driven-parsing.md)

## Dependencies

- [Epic001_cpl-concepts-implementation.md](../epics/Epic001_cpl-concepts-implementation.md) - Units depend on concepts being defined (unit values come from concept values)

## Stories

| Story File                                                                                                            | Status      | Description                                                                                                                                                                | ERO                                                                                                                                 |
| --------------------------------------------------------------------------------------------------------------------- | ----------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| ~~Story141_cpl-units-parser-basic.md~~                                                                                | Canceled    | ~~Implement parser for basic unit definitions~~ - **CANCELED**: Old implementation incompatible with types system; reimplemented without formal story file                 | N/A                                                                                                                                 |
| ~~Story142_cpl-units-analyzer-basic.md~~                                                                              | Canceled    | ~~Implement analyzer for basic unit definitions~~ - **CANCELED**: Old implementation incompatible with types system; reimplemented without formal story file               | N/A                                                                                                                                 |
| ~~Story143_cpl-measures-parser-basic.md~~                                                                             | Canceled    | ~~Implement parser for basic measure definitions~~ - **CANCELED**: Old implementation incompatible with types system; reimplemented without formal story file              | N/A                                                                                                                                 |
| ~~Story144_cpl-measures-analyzer-basic.md~~                                                                           | Canceled    | ~~Implement analyzer for basic measure definitions~~ - **CANCELED**: Old implementation incompatible with types system; reimplemented without formal story file            | N/A                                                                                                                                 |
| ~~Story145_cpl-compound-units-parser.md~~                                                                             | Canceled    | ~~Implement parser for compound unit expressions~~ - **CANCELED**: Moved to canceled-stories/; replaced by Story228 and Story229                                           | N/A                                                                                                                                 |
| ~~Story146_cpl-compound-units-analyzer.md~~                                                                           | Canceled    | ~~Implement analyzer for compound unit expressions~~ - **CANCELED**: Moved to canceled-stories/; replaced by Story228 and Story229                                         | N/A                                                                                                                                 |
| ~~Story147_cpl-unit-conversions-parser.md~~                                                                           | Canceled    | ~~Implement parser for unit conversions~~ - **CANCELED**: Moved to canceled-stories/; replaced by Story228 and Story230                                                    | N/A                                                                                                                                 |
| ~~Story148_cpl-unit-conversions-analyzer.md~~                                                                         | Canceled    | ~~Implement analyzer for unit conversions~~ - **CANCELED**: Moved to canceled-stories/; replaced by Story230                                                               | N/A                                                                                                                                 |
| ~~Story149_cpl-measure-conversions-parser.md~~                                                                        | Canceled    | ~~Implement parser for measure conversions~~ - **CANCELED**: Moved to canceled-stories/; replaced by Story230                                                              | N/A                                                                                                                                 |
| ~~Story150_cpl-measure-conversions-analyzer.md~~                                                                      | Canceled    | ~~Implement analyzer for measure conversions~~ - **CANCELED**: Moved to canceled-stories/; replaced by Story230                                                            | N/A                                                                                                                                 |
| [Story228_cpl-expression-infrastructure.md](../stories/Story228_cpl-expression-infrastructure.md)                     | Not Started | Expression grammar rules, postfix parsing, BuildExpressionTree, CIL expression types, expression synthesis — shared foundation for compound units, conversions, and metrics (to be created) | ERO003 parsing/analysis/synthesis                                                                                                   |
| [Story229_cpl-compound-units.md](../stories/Story229_cpl-compound-units.md)                                           | Not Started | Wire expression infrastructure into measure definitions; replace ConstraintScope-based unit extraction with expression-based UnitExpression (to be created)                 | ERO003 Increment 3                                                                                                                  |
| [Story230_cpl-conversions.md](../stories/Story230_cpl-conversions.md)                                                 | Not Started | Unit and measure conversions with pattern/calculation pairs; simple conversions only; ConversionScope/PatternScope/CalculationScope grammar; conversion code synthesis (to be created) | ERO003 Increments 4–5                                                                                                               |
| [Story231_cpl-unit-type-checking.md](../stories/Story231_cpl-unit-type-checking.md)                                   | Not Started | Synthesis-time unit type verification: walk CIL expression trees with unit tracking, validate conversion outputs match IO spec (to be created)                             | ERO003 synthesis                                                                                                                    |
| [Story232_cpl-conversion-injection.md](../stories/Story232_cpl-conversion-injection.md)                               | Not Started | Automatic CILConversionApplication node injection when unit mismatches have valid conversion pairs (to be created)                                                         | ERO003 synthesis                                                                                                                    |

## Story Ordering and Dependencies

Stories must be executed in dependency order. The dependency chain is strictly linear: 228 → 229 → 230 → 231 → 232.

### Implemented (No Formal Story Files)

#### Phases 1–2: Basic Units and Measures

_Implemented in code between 2026-04-07 and 2026-06-03 without formal story files. See:_
- `cpl_units_listener.go`, `cpl_parser_units_test.go` — units parser
- `units_analysis.go`, `cpl_analyzer_units_multi_section_test.go` — units analyzer
- `cpl_measures_listener.go`, `cpl_parser_measures_test.go` — measures parser
- `measures_analysis.go`, `cpl_analyzer_measures_multi_section_test.go` — measures analyzer

### To Be Implemented

#### Phase 3: Expression Infrastructure (Shared Foundation)

_Prerequisites: Basic units and measures implementations (above)_

5. **[Story228_cpl-expression-infrastructure.md](../stories/Story228_cpl-expression-infrastructure.md)** (Increments 3–5 shared): Expression grammar rules, postfix parsing listener methods, BuildExpressionTree, CIL expression types with unit tracking, expression synthesis — the foundation used by compound units, conversions, and metrics _(to be created)_

#### Phase 4: Compound Units (Increment 3)

_Prerequisites: Story228_

6. **[Story229_cpl-compound-units.md](../stories/Story229_cpl-compound-units.md)** (Increment 3): Wire expression infrastructure into measure definitions — replace ConstraintScope-based unit extraction with expression-based UnitExpression; test with compound unit specifications like Length/Time^2 _(to be created)_

#### Phase 5: Conversions (Increments 4–5 Combined)

_Prerequisites: Story228, Story229_

7. **[Story230_cpl-conversions.md](../stories/Story230_cpl-conversions.md)** (Increments 4–5 combined): Unit and measure conversions with pattern/calculation pairs — ConversionScope/PatternScope/CalculationScope grammar, conversion pair assembly via scope walking, CIL conversion types, conversion code synthesis. Simple (non-conditional) conversions only. _(to be created)_

#### Phase 6: Unit Type Checking

_Prerequisites: Story228, Story229, Story230_

8. **[Story231_cpl-unit-type-checking.md](../stories/Story231_cpl-unit-type-checking.md)** (Synthesis-phase): Walk CIL expression trees with unit tracking, validate that conversion outputs match required IO spec units, catch bogus formulas due to unit confusion. _(to be created)_

#### Phase 7: Conversion Injection

_Prerequisites: Story231_

9. **[Story232_cpl-conversion-injection.md](../stories/Story232_cpl-conversion-injection.md)** (Synthesis-phase): When the unit type checker identifies mismatches that have valid conversion pairs, inject CILConversionApplication nodes into the CIL tree so the synthesizer produces code that performs conversions at runtime. _(to be created)_

### Superseded Stories

The following stories were superseded due to incompatibility with the types system. They have been moved to `memory-bank/stories/canceled-stories/`:

- ~~Story141~~ — Reimplemented without formal story file (basic units parser)
- ~~Story142~~ — Reimplemented without formal story file (basic units analyzer)
- ~~Story143~~ — Reimplemented without formal story file (basic measures parser)
- ~~Story144~~ — Reimplemented without formal story file (basic measures analyzer)
- ~~Story145~~ → Story228 (expression infrastructure) + Story229 (compound units)
- ~~Story146~~ → Story228 (expression infrastructure) + Story229 (compound units)
- ~~Story147~~ → Story228 (expression infrastructure) + Story230 (conversions)
- ~~Story148~~ → Story230 (conversions)
- ~~Story149~~ → Story230 (conversions)
- ~~Story150~~ → Story230 (conversions)

## Constraints

1. All parser work must follow the [Symbol Table-Driven Parsing Pattern](../specifications/patterns/symbol-table-driven-parsing.md) — no AST building in the parser
2. All expression parsing must follow the [Postfix Expression Parsing Pattern](../specifications/patterns/postfix-expression-parsing.md) — no left/right operator nodes during parsing
3. Each story must include tests for the parsing, analysis, or synthesis functionality following Logical TDD Lifecycle
4. Grammar changes must be in the appropriate .g4 files (CPLUnits.g4, CPLMeasures.g4, ChronoconePlanningLanguage.g4)
5. Synthesis IS required for expressions, conversions, unit type checking, and conversion injection — the synthesis ERO has been revised to document this
6. Unit type model: units are declared on measures only (`MeasureName = TypeName(unit=UnitGroup)`). Types do not carry unit information.
7. Simple conversions only (non-conditional). Context-dependent conversions (when clauses) are a future extension.

## Acceptance Criteria

1. Stories 228, 229, 230, 231, and 232 are marked as Completed
2. All superseded stories (Story141-Story150) are moved to canceled-stories/ and marked as Canceled
3. Parser correctly captures unit and measure definitions into symbol tables for all increments
4. Analyzer correctly resolves references and builds expression trees for all increments
5. Compound unit expressions (e.g., `Length/Time^2`) are parsed into postfix order and resolved correctly
6. Unit and measure conversion pattern/calculation pairs are assembled correctly via scope walking
7. Synthesis generates correct conversion code from CIL conversion pairs
8. Unit type checking verifies dimensional consistency and IO spec matching at synthesis time
9. Conversion injection automatically inserts conversion nodes when valid pairs exist
10. All tests pass for parsing, analysis, and synthesis functionality

## Notes

- Stories 228-232 are the new implementation stories for the reimplementation. They replace the superseded Stories 145-150 (canceled) and the never-created Stories 153-162.
- Stories 141-144 (basic units/measures parser and analyzer) were reimplemented in code between 2026-04-07 and 2026-06-03 without formal story files. See `cpl_units_listener.go`, `units_analysis.go`, `cpl_measures_listener.go`, `measures_analysis.go`.
- The parser uses postfix expression parsing via the symbol table stack, not expression tree building.
- The analysis phase evaluates postfix sequences to build resolved expression trees and walks symbol table scopes for conversions.
- The synthesis phase generates conversion code, injects conversion nodes into CIL trees, and verifies unit type consistency.
- Pattern/calculation design: equation form only (`UnitName(var) = UnitName(var)`), pattern-bound variables via parenthesized syntax, all numeric constants must carry explicit unit annotations.
- Algebraic reversibility: inverse conversions are derived at analysis time. The synthesizer sees both forward and inverse pairs as explicit CIL entries.
- Unit type model: units are declared on measures only. Types do not carry unit information.

## Migration from Old Unit Implementation

### Current State (Pre-Epic003)

Before Epic003, the unit implementation used a traditional AST-based approach:

- **`UnitDecl` AST Node**: The parser directly built `UnitDecl` AST nodes during parsing
- **Tight Coupling**: The parser was tightly coupled with AST construction logic
- **Direct AST Building**: Unit definitions were immediately constructed as AST nodes without an intermediate representation

This approach violated the separation of concerns between parsing (syntax) and analysis (semantics) that has been established for concepts and types.

### Terminology Clarification

The following table clarifies the differences between old and new terminology:

| Old Term (Pre-Epic003) | New Term (Epic003/ERO003)   | Description                                         |
| ---------------------- | --------------------------- | --------------------------------------------------- |
| `UnitDecl`             | `ParserUnitDefinition`      | Parser-time structure stored in symbol table        |
| Direct AST building    | Symbol table-driven parsing | Parser populates symbol table; analysis builds AST  |
| N/A                    | `AnalysisUnitDefinition`    | Analysis-time structure with resolved references    |
| N/A                    | `ParserMeasureDefinition`   | Parser-time measure definition in symbol table      |
| N/A                    | `AnalysisMeasureDefinition` | Analysis-time measure with resolved unit expression |
| N/A                    | `UnitScope`, `MeasureScope` | Symbol table scope types for units and measures     |

### Migration Steps for Story Implementation

When implementing the stories in this epic, developers should follow these migration steps:

1. **Implement Symbol Table-Driven Units Parsing** (already done — see `cpl_units_listener.go`):
   - Follow the pattern documented in [ERO003 Units and Measures Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO003_units-measures-parsing-phase.md)
   - Use `EnterUnitsSection` to push `UnitsScope` onto the symbol stack
   - Use `ExitUnitDeclaration` to push/pop `UnitScope`

2. **Build Analysis-Level Structures** (already done — see `units_analysis.go`):
   - Walk the symbol table tree via `WalkSymbolTableWithHandlers`
   - Transform parse-time identifier definitions into `AnalysisUnitDefinition`
   - Resolve concept references during analysis, not parsing

3. **Apply Same Pattern for Measures** (already done — see `cpl_measures_listener.go`, `measures_analysis.go`):
   - Use `MeasureScope` and `ConstraintScope` for basic measure definitions
   - Build `AnalysisMeasureDefinition` during analysis phase
   - Resolve unit references at analysis time

4. **Expression Infrastructure** (Story228 — to be created):
   - Add expression/term/factor grammar rules to `ChronoconePlanningLanguage.g4`
   - Implement postfix parsing listener methods for `*`, `/`, `^`, `+`, `-`, `NUMBER`
   - Implement `BuildExpressionTree` for stack-based postfix evaluation
   - Define CIL expression types with unit tracking

5. **Compound Units** (Story229 — to be created):
   - Wire expression infrastructure into measure definitions
   - Replace ConstraintScope-based unit extraction with expression-based `UnitExpression`

6. **Conversions** (Story230 — to be created):
   - Push `ConversionScope`/`PatternScope`/`CalculationScope` to symbol table
   - Walk conversion scopes during analysis to assemble `AnalysisConversionPair`
   - Derive inverse conversions at analysis time

7. **Unit Type Checking and Conversion Injection** (Stories 231-232 — to be created):
   - Walk CIL expression trees with unit tracking for dimensional verification
   - Inject `CILConversionApplication` nodes when mismatches have valid conversion pairs

### What Was Removed (Pre-Epic003 Units Implementation)

The following units-related code was removed from the codebase on 2026-04-07 because it was incompatible with the types system:

1. **Grammar**: `src/chronocone_planning_language/src/main/antlr/CPLUnits.g4`
2. **AST files**:
   - `src/chronocone_planning_language/src/main/go/ast/cpl_units.go`
   - `src/chronocone_planning_language/src/main/go/ast/cpl_units_test.go`
3. **Parser tests**:
   - `src/chronocone_planning_language/src/main/go/parser/cpl_parser_units_test.go`
   - `src/chronocone_planning_language/src/main/go/parser/exit_unit_declaration_test.go`
4. **Example files**: Units-only CPL examples

The old implementation used a different grammar structure (`unitDeclaration: indent identifier EQUALS typeReference | NEWLINE;`) that conflicted with the types system. The reimplementation (completed 2026-04-07 to 2026-06-03; see Stories 228-232 for remaining work) follows the symbol table-driven parsing pattern and is designed with full awareness of the types system.

### Backward Compatibility Notes

This is a **breaking change** with no backward compatibility:

- **Old `UnitDecl` AST nodes will no longer be produced by the parser** — code that expects `UnitDecl` nodes will need to be updated
- **Parser output changes from AST to symbol table tree** — downstream consumers must walk the symbol table instead of traversing an AST
- **Analysis phase becomes mandatory for unit/measure definitions** — unlike the old approach where AST nodes were ready after parsing, the new approach requires the analysis phase to build usable definitions

**Impact Areas to Check:**

- Any code that imports or references the old `UnitDecl` AST type
- Test cases that verify unit parsing by checking AST structure
- Any intermediate processing that expected parser-output AST nodes
- Code generators that consumed `UnitDecl` nodes directly from parser output

**Recommended Approach:**

- Remove old `UnitDecl` type definitions once Epic003 implementation begins
- Update all tests to verify symbol table structure instead of AST structure
- Ensure analysis phase runs before any code generation that depends on unit/measure definitions
