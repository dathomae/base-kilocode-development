# Story154: CPL Measures Tests and Error Examples

## Goal

Create comprehensive parser tests, analyzer tests, and error example files for Story143 (Measures Parser) and Story144 (Measures Analyzer). This story adds test coverage for the basic measure definitions functionality that is already implemented.

## References

- [Epic003: Units and Measures Implementation](../epics/Epic003_cpl-types-implementation.md)
- [Story143: CPL Measures Parser - Basic Measure Definitions](Story143_cpl-measures-parser-basic.md)
- [Story144: CPL Measures Analyzer - Basic Measure Definitions](Story144_cpl-measures-analyzer-basic.md)
- [Design: Units and Measures](../specifications/design/cpl/units-and-measures.md)
- [Parser Test Pattern: cpl_parser_units_test.go](src/chronocone_planning_language/src/main/go/parser/cpl_parser_units_test.go)

## Dependencies

- [Story143_cpl-measures-parser-basic.md](Story143_cpl-measures-parser-basic.md) - Parser implementation must be complete
- [Story144_cpl-measures-analyzer-basic.md](Story144_cpl-measures-analyzer-basic.md) - Analyzer implementation must be complete

## Dependent Stories

None - This is a test coverage story that does not block other work.

## Tasks

1. Create parser tests for measures section - Completed
   a. Create `cpl_parser_measures_test.go` file - Completed
   b. Implement TestParseSingleMeasureDefinition test - Completed
   c. Implement TestParseMultipleMeasureDefinitions test - Completed
   d. Implement TestParseMeasuresWithWhitespaceVariations test - Completed
   e. Implement TestParseMeasuresSourceLocation test - Completed
   f. Implement TestParseMeasuresScopeIntegration test - Completed
   g. Implement TestParseMeasuresSymbolTableIntegration test with helper function - Completed
   h. Implement TestParseMeasuresInvalidSyntax test - Completed
   i. Implement TestParseMeasuresRequiresAtLeastOneDefinition test - Completed
   j. Implement TestParseMeasuresWithTypesAndUnitsTogether test - Completed

2. Create analysis tests for measures - Completed
   a. Create `measures_analysis_test.go` file - Completed
   b. Implement TestAnalyzeMeasureScope_SingleMeasure test - Completed
   c. Implement TestAnalyzeMeasureScope_MultipleMeasures test - Completed
   d. Implement TestAnalyzeMeasureScope_UnitReferenceResolution test - Completed
   e. Implement TestAnalyzeMeasureScope_DuplicateMeasure test - Completed
   f. Implement TestAnalyzeMeasureScope_WrongIdentifierCount test - Completed
   g. Implement TestAnalyzeMeasureScope_EmptyDefinitions test - Completed
   h. Implement TestIntegration_WithWalkSymbolTableWithHandlers test - Completed
   i. Implement TestEndToEnd_ParseAndAnalyze test - Completed

3. Create error example files - Completed
   a. Create `invalid_measure_missing_equals.cpl` - Completed
   b. Create `invalid_measure_missing_unit_keyword.cpl` - Completed
   c. Create `invalid_measure_missing_parens.cpl` - Completed
   d. Create `invalid_measure_empty_section.cpl` - Completed
   e. Create `invalid_measure_duplicate_name.cpl` - Completed
   f. Create `invalid_measure_wrong_syntax.cpl` - Completed

4. Verify all tests pass - Completed
   a. Run parser tests - Completed
   b. Run analysis tests - Completed
   c. Verify integration tests - Completed

## Test-First Development

This story follows the Logical TDD Lifecycle. For each task:

1. **Phase A (Structural Definition)**: Define the test structure and assertions before implementation
2. **Phase B (Logical Tests)**: Write tests that verify the expected behavior
3. **Phase C (Implementation)**: The implementation already exists from Story143 and Story144; tests validate it

Note: Tasks 1 and 2 write tests for already-implemented functionality. Task 3 creates error example files that should fail parsing or analysis as expected.

## Constraints

- All parser tests must use table-driven test patterns following `cpl_parser_units_test.go` conventions
- Tests must use `t.Parallel()` for parallel execution
- Tests must use testify assertions (`assert`, `require`)
- Range variable capture in closures must be handled correctly (use `tc := tc`)
- Error example files must include comment describing expected failure
- Parser test helper function `findMeasureScopeByMeasureName` follows the pattern from `findUnitScopeByUnitName`

## Intent

The intent of this story is to ensure comprehensive test coverage for the measures parsing and analysis functionality, providing:

- Confidence that existing implementation works correctly
- Regression protection for future changes
- Clear documentation of expected behavior through tests
- Error example files for developer reference and debugging

## Acceptance Criteria

1. All parser tests pass (`go test ./parser/... -run Measures`)
2. All analysis tests pass (`go test ./analysis/... -run Measure`)
3. All 6 error example files are created in `examples/with_errors/`
4. Error example files follow the established comment style (first line describes expected failure)
5. Integration tests continue to pass
6. Test coverage for measures functionality is comprehensive (all parsing paths tested)

## Notes

### Implementation Background

The measures implementation from Story143 and Story144 includes:

**Parser Layer** (`cpl_measures_listener.go`):

- `EnterMeasuresSection` - Pushes MeasuresScope onto symbol stack
- `ExitMeasuresSection` - Pops MeasuresScope from symbol stack
- `EnterMeasureDeclaration` - Pushes MeasureScope onto symbol stack
- `ExitMeasureDeclaration` - Pops MeasureScope from symbol stack

**Analysis Layer** (`measures_analysis.go`):

- `AnalyzeMeasureScope` - Extracts measure name, type name, and unit name from symbol table
- `AnalysisMeasureDefinition` struct - Stores analyzed measure data
- Integrated into `CPLAnalyzer` with `measures` map

**Grammar** (`CPLMeasures.g4`):

```
measuresSection: MEASURES measureDeclaration+;
measureDeclaration: indent (identifier EQUALS identifier LPAREN UNIT EQUALS identifier RPAREN)*;
```

**Valid Example** (`measures-example.cpl`):

```cpl
measures
    CurrencyMeasure = CurrencyType(unit=CurrencyUnit)
    Currency = Currency(unit=Currency)
```

### Reference Patterns

- **Parser Tests**: `cpl_parser_units_test.go` - Use as template for structure and patterns
- **Analysis Tests**: `analysis_integration_test.go` - Use for end-to-end test patterns
- **Error Examples**: Existing files in `examples/with_errors/` - Follow comment style and naming conventions
