# Story142: CPL Units Analyzer - Basic Unit Definitions (Increment 1)

## Status

Done

## Goal

Implement the analyzer for basic unit definitions including walking UnitsScope, extracting ParserUnitDefinition, building AnalysisUnitDefinition, and validating concept references. This is Increment 1 of Epic003 units analysis, building on Story141.

## References

- [Epic003_units-and-measures-implementation.md](../epics/Epic003_units-and-measures-implementation.md) - Parent epic for units and measures implementation
- [ERO003_units-measures-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO003_units-measures-analysis-phase.md) - Analysis phase data structures for Increment 1
- [ERO003_units-measures-incremental-implementation.md](../specifications/design/cpl/eros/ERO003_units-measures-incremental-implementation.md) - Main incremental implementation ERO
- [Story141_cpl-units-parser-basic.md](Story141_cpl-units-parser-basic.md) - Parser increment 1 (dependency)
- [Symbol Table-Driven Parsing Pattern](../specifications/patterns/symbol-table-driven-parsing.md) - Architectural pattern for parsing and analysis

### Phase Separation Note

This story implements the analysis phase. It transforms parser-level structures into analysis-level structures with resolved references.

### Migration Reference

This story follows the ERO003 design which replaces the old UnitDecl AST approach with symbol table-driven parsing.

## Dependencies

- [Story141_cpl-units-parser-basic.md](Story141_cpl-units-parser-basic.md) - Must be completed (parser must provide UnitsScope and ParserUnitDefinition)

## Dependent Stories

- [Story143_cpl-measures-parser-basic.md](Story143_cpl-measures-parser-basic.md) - Not Started
- [Story144_cpl-measures-analyzer-basic.md](Story144_cpl-measures-analyzer-basic.md) - Not Started

## Tasks

1. Define AnalysisUnitDefinition and Units lookup map in AnalysisDefinitions - Completed
   a. Added Units map[string]AnalysisUnitDefinition to AnalysisDefinitions struct in cpl_analyzer.go
   b. Defined AnalysisUnitDefinition struct in units_analysis.go with Name, ConceptName, Location fields

   Implementation: AnalysisUnitDefinition struct and Units map are in place.

2. Implement UnitAnalyzer struct with AnalyzeUnitsScope method - Completed
   a. CPLAnalyzer already contains units map (no separate UnitAnalyzer needed)
   b. AnalyzeUnitScope function implemented in units_analysis.go
   c. GetResults provided via analyzer.GetAnalysisDefinitions().Units

   Implementation: AnalyzeUnitScope extracts unit name and concept name from UnitScope.Definitions array.

3. Implement concept reference resolution - Completed
   a. VerifyUnitConceptReferences validates concept existence in concepts map
   b. Called in Analyze() method after symbol table walk

   Implementation: VerifyUnitConceptReferences at units_analysis.go:99-106.

4. Implement duplicate detection in AnalyzeUnitsScope - Completed
   a. Duplicate check implemented at units_analysis.go:65-67
   b. Returns error for duplicate unit names

5. Integrate UnitAnalyzer with CPLCompiler.Analyze method - Completed
   a. UnitScope handler registered with WalkSymbolTableWithHandlers at cpl_analyzer.go:497
   b. Units map included in AnalysisDefinitions return value
   c. VerifyUnitConceptReferences called at cpl_analyzer.go:515-518

6. Write comprehensive unit tests for units analysis - Completed
   a. Parser tests in cpl_parser_units_test.go
   b. Integration tests in analysis_integration_test.go
   c. Symbol table walk tests cover UnitScope handling

## Notes

- This story implements Increment 1 of the units analysis phase as defined in ERO003
- No conversion clauses or compound units in this increment
- Concept existence is validated during analysis (not during parsing)
- Follow Logical TDD Lifecycle: write tests before implementing code for each task
