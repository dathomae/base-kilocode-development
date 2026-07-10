# Story148: CPL Unit Conversions Analyzer - Conversion Pair Assembly (Increment 4)

## Goal

Implement the analyzer for unit conversions including walking ConversionScope tree, assembling AnalysisConversionPair structures, resolving pattern and calculation expressions, and validating unit references. This is Increment 4 of Epic003 units and measures analysis.

## References

- [Epic003_units-and-measures-implementation.md](../epics/Epic003_units-and-measures-implementation.md) - Parent epic for units and measures implementation
- [ERO003_units-measures-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO003_units-measures-analysis-phase.md) - Analysis phase data structures for Increment 4
- [ERO003_units-measures-incremental-implementation.md](../specifications/design/cpl/eros/ERO003_units-measures-incremental-implementation.md) - Main incremental implementation ERO
- [Postfix Expression Parsing Pattern](../specifications/patterns/postfix-expression-parsing.md) - Pattern for expression parsing
- [Symbol Table-Driven Parsing Pattern](../specifications/patterns/symbol-table-driven-parsing.md) - Architectural pattern for parsing and analysis
- [Story146_cpl-compound-units-analyzer.md](Story146_cpl-compound-units-analyzer.md) - Analyzer increment 3 (dependency)
- [Story147_cpl-unit-conversions-parser.md](Story147_cpl-unit-conversions-parser.md) - Parser increment 4 (dependency)

### Phase Separation Note

This story implements the analysis phase. It transforms parser-level structures into analysis-level structures with resolved references.

### Migration Reference

This story follows the ERO003 design which replaces the old UnitDecl AST approach with symbol table-driven parsing.

## Dependencies

- [Story146_cpl-compound-units-analyzer.md](Story146_cpl-compound-units-analyzer.md) - Must be completed (expression tree infrastructure)
- [Story147_cpl-unit-conversions-parser.md](Story147_cpl-unit-conversions-parser.md) - Must be completed (parser support for conversions)

## Dependent Stories

- [Story149_cpl-measure-conversions-parser.md](Story149_cpl-measure-conversions-parser.md) - Not Started
- [Story150_cpl-measure-conversions-analyzer.md](Story150_cpl-measure-conversions-analyzer.md) - Not Started

## Tasks

1. Define AnalysisConversionPair and related types - Not Started
   a. Define ConversionPattern struct with unit expressions and value aliases - Not Started
   b. Define ConversionCalculation struct with expression tree - Not Started
   c. Define AnalysisConversionPair struct with Pattern and Calculation fields - Not Started

   **Intent**: Create analysis data structures for conversion pairs.
   **Acceptance Criteria**: Conversion pairs can store resolved pattern and calculation expressions.

2. Implement conversion scope walking in UnitAnalyzer - Not Started
   a. Walk ConversionScope children of UnitScope - Not Started
   b. Extract PatternScope and CalculationScope definitions - Not Started
   c. Pair patterns with calculations by order within ConversionScope - Not Started

   **Intent**: Process conversion scopes to build conversion pair structures.
   **Acceptance Criteria**: Conversion pairs are correctly assembled from symbol table tree.

3. Implement pattern expression resolution - Not Started
   a. Parse pattern expression into postfix components - Not Started
   b. Build expression tree for pattern - Not Started
   c. Resolve unit references in pattern - Not Started
   d. Extract and store value aliases - Not Started

   **Intent**: Resolve pattern expressions to expression trees with unit references.
   **Acceptance Criteria**: Pattern expressions are correctly resolved with unit and value aliases.

4. Implement calculation expression resolution - Not Started
   a. Parse calculation expression into postfix components - Not Started
   b. Build expression tree for calculation - Not Started
   c. Resolve value alias references - Not Started

   **Intent**: Resolve calculation expressions to expression trees with value alias references.
   **Acceptance Criteria**: Calculation expressions are correctly resolved with value alias references.

5. Implement unit reference validation in conversions - Not Started
   a. Validate all unit references in patterns resolve to valid units - Not Started
   b. Validate unit references are compatible (same unit group) - Not Started

   **Intent**: Ensure conversion unit references are valid and compatible.
   **Acceptance Criteria**: Invalid unit references in conversions produce appropriate errors.

6. Integrate conversion analysis with unit analysis flow - Not Started
   a. Call conversion analysis after processing unit definition - Not Started
   b. Store conversion pairs in AnalysisUnitDefinition - Not Started

   **Intent**: Wire conversion analysis into the unit analysis pipeline.
   **Acceptance Criteria**: Conversion pairs are attached to unit definitions.

7. Write comprehensive unit tests for unit conversions analysis - Not Started
   a. Test conversion pair assembly from symbol table tree - Not Started
   b. Test pattern expression resolution - Not Started
   c. Test calculation expression resolution - Not Started
   d. Test unit reference validation in conversions - Not Started
   e. Test error handling for unresolvable unit references - Not Started
   f. Test multiple conversion pairs in one conversion section - Not Started

   **Intent**: Verify conversion analysis functionality works correctly.
   **Acceptance Criteria**: All tests pass, covering valid inputs and error conditions.

## Notes

- This story implements Increment 4 of the units and measures analysis phase as defined in ERO003
- Pattern and calculation pairing is done by order within ConversionScope
- Formula evaluation depends on metrics implementation
- No measure conversions in this increment
- Follow Logical TDD Lifecycle: write tests before implementing code for each task
