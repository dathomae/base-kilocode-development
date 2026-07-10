# Story146: CPL Compound Units Analyzer - Unit Algebra Tree (Increment 3)

## Goal

Implement the analyzer for compound unit expressions including BuildExpressionTree function that evaluates postfix sequence, Expression tree types (UnitReference, MultiplicationNode, DivisionNode, PowerNode), and unit resolution. This is Increment 3 of Epic003 units and measures analysis.

## References

- [Epic003_units-and-measures-implementation.md](../epics/Epic003_units-and-measures-implementation.md) - Parent epic for units and measures implementation
- [ERO003_units-measures-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO003_units-measures-analysis-phase.md) - Analysis phase data structures for Increment 3
- [ERO003_units-measures-incremental-implementation.md](../specifications/design/cpl/eros/ERO003_units-measures-incremental-implementation.md) - Main incremental implementation ERO
- [Postfix Expression Parsing Pattern](../specifications/patterns/postfix-expression-parsing.md) - Pattern for expression parsing
- [Symbol Table-Driven Parsing Pattern](../specifications/patterns/symbol-table-driven-parsing.md) - Architectural pattern for parsing and analysis
- [Story144_cpl-measures-analyzer-basic.md](Story144_cpl-measures-analyzer-basic.md) - Analyzer increment 2 (dependency)
- [Story145_cpl-compound-units-parser.md](Story145_cpl-compound-units-parser.md) - Parser increment 3 (dependency)

### Phase Separation Note

This story implements the analysis phase. It transforms parser-level structures into analysis-level structures with resolved references.

### Migration Reference

This story follows the ERO003 design which replaces the old UnitDecl AST approach with symbol table-driven parsing.

## Dependencies

- [Story144_cpl-measures-analyzer-basic.md](Story144_cpl-measures-analyzer-basic.md) - Must be completed (measures analysis infrastructure)
- [Story145_cpl-compound-units-parser.md](Story145_cpl-compound-units-parser.md) - Must be completed (parser support for compound units)

## Dependent Stories

- [Story147_cpl-unit-conversions-parser.md](Story147_cpl-unit-conversions-parser.md) - Not Started
- [Story148_cpl-unit-conversions-analyzer.md](Story148_cpl-unit-conversions-analyzer.md) - Not Started

## Tasks

1. Define Expression interface and implementing types - Not Started
   a. Define Expression interface with isExpression() marker method - Not Started
   b. Define UnitReference struct with Name and UnitGroup (\*AnalysisUnitDefinition) fields - Not Started
   c. Define MultiplicationNode struct with Left and Right (Expression) fields - Not Started
   d. Define DivisionNode struct with Left and Right (Expression) fields - Not Started
   e. Define PowerNode struct with Base (Expression) and Exponent (int) fields - Not Started

   **Intent**: Create expression tree types for resolved unit algebra.
   **Acceptance Criteria**: Expression tree types can represent compound unit expressions.

2. Implement BuildExpressionTree function - Not Started
   a. Create BuildExpressionTree(postfix []SymbolDefinition, units map[string]AnalysisUnitDefinition) function - Not Started
   b. Implement stack-based postfix evaluation algorithm - Not Started
   c. Push UnitReference nodes for unit references - Not Started
   d. Push MultiplicationNode, DivisionNode, PowerNode for operators - Not Started
   e. Return single Expression tree root - Not Started

   **Intent**: Convert postfix expression components into resolved expression tree.
   **Acceptance Criteria**: Postfix sequence [Length, Time, Power, 2, Divide] produces DivisionNode with UnitReference(Length) and PowerNode(Time, 2).

3. Implement unit resolution during expression tree building - Not Started
   a. Look up unit names in units map during postfix evaluation - Not Started
   b. Store pointer to AnalysisUnitDefinition in UnitReference nodes - Not Started
   c. Return error for unresolvable unit references - Not Started

   **Intent**: Resolve unit references to actual unit definitions.
   **Acceptance Criteria**: All unit references in expression trees point to valid AnalysisUnitDefinition entries.

4. Update AnalyzeMeasuresScope to handle compound unit expressions - Not Started
   a. Extract expression component definitions from MeasureScope's Definitions slice - Not Started
   b. Call BuildExpressionTree to create Expression tree - Not Started
   c. Store Expression tree in AnalysisMeasureDefinition - Not Started

   **Intent**: Integrate expression tree building into measure analysis.
   **Acceptance Criteria**: Measures with compound units have resolved Expression trees.

5. Implement unit algebra validation - Not Started
   a. Validate expression tree structure is well-formed - Not Started
   b. Validate all unit references are resolved - Not Started

   **Intent**: Ensure compound unit expressions are semantically valid.
   **Acceptance Criteria**: Invalid unit algebra produces appropriate errors.

6. Write comprehensive unit tests for compound units analysis - Not Started
   a. Test BuildExpressionTree with simple multiplication - Not Started
   b. Test BuildExpressionTree with simple division - Not Started
   c. Test BuildExpressionTree with power - Not Started
   d. Test BuildExpressionTree with compound expression (Length/Time^2) - Not Started
   e. Test unit resolution in expression trees - Not Started
   f. Test error handling for unresolvable unit references - Not Started
   g. Test error handling for invalid postfix sequences - Not Started

   **Intent**: Verify expression tree building and unit resolution work correctly.
   **Acceptance Criteria**: All tests pass, covering valid inputs and error conditions.

## Notes

- This story implements Increment 3 of the units and measures analysis phase as defined in ERO003
- Expression tree is designed for future tensor extension (TensorNode placeholder)
- Unit trees are preserved symbolically without simplification in this increment
- No unit conversions in this increment
- Follow Logical TDD Lifecycle: write tests before implementing code for each task
