# Story144: CPL Measures Analyzer - Basic Measure Definitions (Increment 2)

## Goal

Implement the analyzer for basic measure definitions including walking MeasuresScope, extracting ParserMeasureDefinition, resolving UnitGroupName to AnalysisUnitDefinition, and building AnalysisMeasureDefinition. This is Increment 2 of Epic003 measures analysis, building on Story142.

## References

- [Epic003_units-and-measures-implementation.md](../epics/Epic003_units-and-measures-implementation.md) - Parent epic for units and measures implementation
- [ERO003_units-measures-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO003_units-measures-analysis-phase.md) - Analysis phase data structures for Increment 2
- [ERO003_units-measures-incremental-implementation.md](../specifications/design/cpl/eros/ERO003_units-measures-incremental-implementation.md) - Main incremental implementation ERO
- [Story142_cpl-units-analyzer-basic.md](Story142_cpl-units-analyzer-basic.md) - Analyzer increment 1 (dependency)
- [Story143_cpl-measures-parser-basic.md](Story143_cpl-measures-parser-basic.md) - Parser increment 2 (dependency)
- [Symbol Table-Driven Parsing Pattern](../specifications/patterns/symbol-table-driven-parsing.md) - Architectural pattern for parsing and analysis

### Phase Separation Note

This story implements the analysis phase. It transforms parser-level structures into analysis-level structures with resolved references.

### Migration Reference

This story follows the ERO003 design which replaces the old UnitDecl AST approach with symbol table-driven parsing.

## Dependencies

- [Story142_cpl-units-analyzer-basic.md](Story142_cpl-units-analyzer-basic.md) - Must be completed (units analysis infrastructure)
- [Story143_cpl-measures-parser-basic.md](Story143_cpl-measures-parser-basic.md) - Must be completed (parser must provide MeasuresScope and ParserMeasureDefinition)

## Dependent Stories

- [Story145_cpl-compound-units-parser.md](Story145_cpl-compound-units-parser.md) - Not Started
- [Story146_cpl-compound-units-analyzer.md](Story146_cpl-compound-units-analyzer.md) - Not Started

## Tasks

1. Define AnalysisMeasureDefinition and Measures lookup map in AnalysisDefinitions - Not Started
   a. Add Measures map[string]AnalysisMeasureDefinition to AnalysisDefinitions struct - Not Started
   b. Define AnalysisMeasureDefinition struct with Name, TypeName, UnitGroup (\*AnalysisUnitDefinition) fields - Not Started

   **Intent**: Create the core analysis data structures for measure definitions.
   **Acceptance Criteria**: Measures can be stored in the analysis definitions with resolved unit group references.

2. Implement MeasureAnalyzer struct with AnalyzeMeasuresScope method - Not Started
   a. Define MeasureAnalyzer struct with measures map - Not Started
   b. Implement NewMeasureAnalyzer() constructor - Not Started
   c. Implement AnalyzeMeasuresScope(table) to process measure definitions from symbol table - Not Started
   d. Implement GetResults() to return analyzed measures map - Not Started

   **Intent**: Build the analyzer that processes parsed measure definitions into validated analysis structures.
   **Acceptance Criteria**: Analyzer processes MeasuresScope symbol tables and produces validated AnalysisMeasureDefinition entries.

3. Implement unit group reference resolution - Not Started
   a. Look up unit group name in analysis units map - Not Started
   b. Store pointer to AnalysisUnitDefinition in AnalysisMeasureDefinition - Not Started
   c. Return error if unit group reference cannot be resolved - Not Started

   **Intent**: Resolve measure unit group references to actual unit definitions.
   **Acceptance Criteria**: Measures are populated with pointers to their referenced unit groups.

4. Implement type name validation - Not Started
   a. Validate that type name is Number or Number-derived - Not Started
   b. Return error for unsupported type names - Not Started

   **Intent**: Ensure measures only use supported types in this increment.
   **Acceptance Criteria**: Measures with non-Number types produce appropriate errors.

5. Implement duplicate detection in AnalyzeMeasuresScope - Not Started
   a. Check if measure group name already exists in measures map before adding - Not Started
   b. Return error for duplicate measure group definitions - Not Started

   **Intent**: Prevent duplicate measure group definitions within the same compilation unit.
   **Acceptance Criteria**: Duplicate measure definitions are detected and reported as errors.

6. Integrate MeasureAnalyzer with CPLCompiler.Analyze method - Not Started
   a. Create MeasureAnalyzer instance in Analyze method - Not Started
   b. Register MeasuresScope handler with WalkSymbolTableWithHandlers - Not Started
   c. Add Measures to AnalysisDefinitions return value - Not Started

   **Intent**: Wire the measure analyzer into the main compilation pipeline.
   **Acceptance Criteria**: Measures are analyzed as part of the standard compilation flow.

7. Write comprehensive unit tests for measures analysis - Not Started
   a. Test AnalyzeMeasuresScope with single measure definition - Not Started
   b. Test unit group reference resolution - Not Started
   c. Test error handling for unresolvable unit group references - Not Started
   d. Test type name validation (Number vs unsupported types) - Not Started
   e. Test duplicate measure group detection - Not Started
   f. Test integration with WalkSymbolTableWithHandlers - Not Started

   **Intent**: Verify all analysis functionality works correctly and errors are properly reported.
   **Acceptance Criteria**: All tests pass, covering valid inputs and error conditions.

## Notes

- This story implements Increment 2 of the measures analysis phase as defined in ERO003
- Only Number type and Number-derived types supported
- No compound unit expressions in measure declarations
- No measure conversions in this increment
- Follow Logical TDD Lifecycle: write tests before implementing code for each task
