# Story150: CPL Measure Conversions Analyzer - Measure Conversion Pair Assembly (Increment 5)

## Goal

Implement the analyzer for measure conversions including walking MeasureConversionScope tree, assembling AnalysisMeasureConversionPair structures, resolving measure references in patterns and calculations, and validating measure references. This is the final increment (5) of Epic003 units and measures analysis, completing full units and measures functionality.

## References

- [Epic003_units-and-measures-implementation.md](../epics/Epic003_units-and-measures-implementation.md) - Parent epic for units and measures implementation
- [ERO003_units-measures-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO003_units-measures-analysis-phase.md) - Analysis phase data structures for Increment 5
- [ERO003_units-measures-incremental-implementation.md](../specifications/design/cpl/eros/ERO003_units-measures-incremental-implementation.md) - Main incremental implementation ERO
- [Postfix Expression Parsing Pattern](../specifications/patterns/postfix-expression-parsing.md) - Pattern for expression parsing
- [Symbol Table-Driven Parsing Pattern](../specifications/patterns/symbol-table-driven-parsing.md) - Architectural pattern for parsing and analysis
- [Story148_cpl-unit-conversions-analyzer.md](Story148_cpl-unit-conversions-analyzer.md) - Analyzer increment 4 (dependency)
- [Story149_cpl-measure-conversions-parser.md](Story149_cpl-measure-conversions-parser.md) - Parser increment 5 (dependency)

### Phase Separation Note

This story implements the analysis phase. It transforms parser-level structures into analysis-level structures with resolved references.

### Migration Reference

This story follows the ERO003 design which replaces the old UnitDecl AST approach with symbol table-driven parsing.

## Dependencies

- [Story148_cpl-unit-conversions-analyzer.md](Story148_cpl-unit-conversions-analyzer.md) - Must be completed (conversion analysis infrastructure)
- [Story149_cpl-measure-conversions-parser.md](Story149_cpl-measure-conversions-parser.md) - Must be completed (parser support for measure conversions)

## Dependent Stories

- None - This completes the units and measures implementation epic

## Tasks

1. Define AnalysisMeasureConversionPair and related types - Not Started
   a. Define MeasureConversionPattern struct with measure expressions and aliases - Not Started
   b. Define MeasureConversionCalculation struct with expression tree - Not Started
   c. Define AnalysisMeasureConversionPair struct with Pattern and Calculation fields - Not Started

   **Intent**: Create analysis data structures for measure conversion pairs.
   **Acceptance Criteria**: Measure conversion pairs can store resolved pattern and calculation expressions.

2. Implement measure conversion scope walking in MeasureAnalyzer - Not Started
   a. Walk MeasureConversionScope children of MeasureScope - Not Started
   b. Extract PatternScope and CalculationScope definitions - Not Started
   c. Pair patterns with calculations by order within MeasureConversionScope - Not Started

   **Intent**: Process measure conversion scopes to build measure conversion pair structures.
   **Acceptance Criteria**: Measure conversion pairs are correctly assembled from symbol table tree.

3. Implement measure reference resolution in patterns - Not Started
   a. Resolve measure references in pattern expressions - Not Started
   b. Validate measure references point to valid AnalysisMeasureDefinition entries - Not Started

   **Intent**: Resolve measure references to actual measure definitions.
   **Acceptance Criteria**: Measure references in patterns are resolved and validated.

4. Implement context-dependent calculation resolution - Not Started
   a. Parse context-dependent calculation expressions - Not Started
   b. Resolve measure and value references in calculations - Not Started

   **Intent**: Resolve context-dependent calculation expressions.
   **Acceptance Criteria**: Context-dependent calculations are correctly resolved.

5. Implement measure reference validation in conversions - Not Started
   a. Validate all measure references in patterns resolve to valid measures - Not Started
   b. Validate measure references are compatible - Not Started

   **Intent**: Ensure conversion measure references are valid and compatible.
   **Acceptance Criteria**: Invalid measure references in conversions produce appropriate errors.

6. Integrate measure conversion analysis with measure analysis flow - Not Started
   a. Call measure conversion analysis after processing measure definition - Not Started
   b. Store measure conversion pairs in AnalysisMeasureDefinition - Not Started

   **Intent**: Wire measure conversion analysis into the measure analysis pipeline.
   **Acceptance Criteria**: Measure conversion pairs are attached to measure definitions.

7. Write comprehensive unit tests for measure conversions analysis - Not Started
   a. Test measure conversion pair assembly from symbol table tree - Not Started
   b. Test measure reference resolution in patterns - Not Started
   c. Test context-dependent calculation resolution - Not Started
   d. Test measure reference validation in conversions - Not Started
   e. Test error handling for unresolvable measure references - Not Started
   f. Test multiple measure conversion pairs in one conversion section - Not Started

   **Intent**: Verify measure conversion analysis functionality works correctly.
   **Acceptance Criteria**: All tests pass, covering valid inputs and error conditions.

## Notes

- This story implements Increment 5 (final) of the units and measures analysis phase as defined in ERO003
- Measure conversions support context-dependent transformations
- Complex time-related conversions may require additional time-handling features
- This completes full units and measures functionality for all 5 increments
- No synthesis-phase work is required for units and measures
- Follow Logical TDD Lifecycle: write tests before implementing code for each task
