# Story147: CPL Unit Conversions Parser - Pattern/Calculation Pairs (Increment 4)

## Goal

Implement the parser for unit conversions including conversionClause/patternClause/calculationClause grammar rules, ConversionScope/PatternScope/CalculationScope symbol table scopes, and extended expression rules with +, -, numeric literals. This is Increment 4 of Epic003 units and measures parsing.

## References

- [Epic003_units-and-measures-implementation.md](../epics/Epic003_units-and-measures-implementation.md) - Parent epic for units and measures implementation
- [ERO003_units-measures-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO003_units-measures-parsing-phase.md) - Parsing phase data structures for Increment 4
- [ERO003_units-measures-incremental-implementation.md](../specifications/design/cpl/eros/ERO003_units-measures-incremental-implementation.md) - Main incremental implementation ERO
- [Postfix Expression Parsing Pattern](../specifications/patterns/postfix-expression-parsing.md) - Pattern for expression parsing
- [Symbol Table-Driven Parsing Pattern](../specifications/patterns/symbol-table-driven-parsing.md) - Architectural pattern for parsing
- [Story145_cpl-compound-units-parser.md](Story145_cpl-compound-units-parser.md) - Parser increment 3 (dependency)

### Phase Separation Note

This story implements the parsing phase. Raw syntactic data is captured into the symbol table; semantic structures are built by the analysis phase.

### Migration Reference

This story follows the ERO003 design which replaces the old UnitDecl AST approach with symbol table-driven parsing.

## Dependencies

- [Story145_cpl-compound-units-parser.md](Story145_cpl-compound-units-parser.md) - Must be completed (expression parsing infrastructure)

## Dependent Stories

- [Story148_cpl-unit-conversions-analyzer.md](Story148_cpl-unit-conversions-analyzer.md) - Not Started
- [Story149_cpl-measure-conversions-parser.md](Story149_cpl-measure-conversions-parser.md) - Not Started

## Tasks

1. Extend expression rules in main grammar for conversion calculations - Not Started
   a. Add PLUS and MINUS alternatives to expression rule - Not Started
   b. Add NUMBER (numeric literal) to factor rule - Not Started
   c. Add identifier LPAREN identifier RPAREN form for unit with value alias (e.g., C(c)) - Not Started

   **Intent**: Extend expression grammar to support arithmetic operations needed in conversion calculations.
   **Acceptance Criteria**: Grammar parses expressions like `((5/9)*(C/F))*(f-32*F)` and `1000`.

2. Add lexer tokens for conversion expressions - Not Started
   a. Add PLUS token for '+' - Not Started
   b. Add MINUS token for '-' - Not Started
   c. Add NUMBER token for numeric literals (integers and decimals) - Not Started

   **Intent**: Provide lexical analysis for conversion expression operators and literals.
   **Acceptance Criteria**: Lexer correctly tokenizes +, -, and numeric literals.

3. Update CPLUnits.g4 to support conversion clauses - Not Started
   a. Extend unitDefinition rule to include optional conversionClause\* - Not Started
   b. Define conversionClause rule: CONVERSION_KEYWORD conversionPair+ - Not Started
   c. Define conversionPair rule: patternClause calculationClause - Not Started
   d. Define patternClause rule: PATTERN_KEYWORD pattern - Not Started
   e. Define calculationClause rule: CALCULATION_KEYWORD expression - Not Started
   f. Define pattern rule supporting ratio and equation forms - Not Started

   **Intent**: Add grammar for unit conversion pattern/calculation pairs.
   **Acceptance Criteria**: Grammar parses conversion clauses with pattern/calculation pairs.

4. Add lexer tokens for conversion keywords - Not Started
   a. Add CONVERSION_KEYWORD token for 'conversion' - Not Started
   b. Add PATTERN_KEYWORD token for 'pattern' - Not Started
   c. Add CALCULATION_KEYWORD token for 'calculation' - Not Started

   **Intent**: Provide lexical analysis for conversion-related keywords.
   **Acceptance Criteria**: Lexer correctly tokenizes conversion, pattern, and calculation keywords.

5. Add ConversionScope, PatternScope, CalculationScope to ScopeType enum - Not Started
   a. Add ConversionScope constant to ScopeType enum - Not Started
   b. Add PatternScope constant to ScopeType enum - Not Started
   c. Add CalculationScope constant to ScopeType enum - Not Started
   d. Add corresponding DefinitionType values - Not Started

   **Intent**: Extend symbol table to recognize conversion-related scopes.
   **Acceptance Criteria**: Symbol table can create and manage conversion, pattern, and calculation scopes.

6. Implement conversion listener methods in UnitsListener - Not Started
   a. Implement EnterConversionClause to push ConversionScope - Not Started
   b. Implement ExitConversionClause to pop ConversionScope - Not Started
   c. Implement EnterPatternClause to push PatternScope - Not Started
   d. Implement ExitPatternClause to pop PatternScope and store pattern expression - Not Started
   e. Implement EnterCalculationClause to push CalculationScope - Not Started
   f. Implement ExitCalculationClause to pop CalculationScope and store calculation expression - Not Started

   **Intent**: Build listener methods that process conversion clauses into symbol table scopes.
   **Acceptance Criteria**: Conversion clauses create proper nested scopes with pattern and calculation expressions.

7. Implement validation during parsing for Increment 4 - Not Started
   a. Validate conversion clause syntax - Not Started
   b. Validate pattern syntax (ratio and equation forms) - Not Started
   c. Validate calculation expression syntax - Not Started
   d. Validate value alias syntax - Not Started

   **Intent**: Ensure conversion clauses are syntactically valid.
   **Acceptance Criteria**: Invalid conversion clauses produce clear error messages.

8. Write comprehensive unit tests for unit conversions parsing - Not Started
   a. Test parsing of simple ratio pattern with numeric calculation - Not Started
   b. Test parsing of equation pattern with complex calculation - Not Started
   b. Test parsing of multiple pattern/calculation pairs in one conversion - Not Started
   c. Test parsing of unit aliases in patterns (e.g., Kilometer(km)) - Not Started
   d. Test parsing of value aliases in calculations (e.g., f, c) - Not Started
   e. Test error handling for missing calculation - Not Started
   f. Test error handling for invalid pattern syntax - Not Started
   g. Test symbol table structure for conversions - Not Started

   **Intent**: Verify unit conversion parsing works correctly and errors are properly reported.
   **Acceptance Criteria**: All tests pass, covering valid inputs and error conditions.

## Notes

- This story implements Increment 4 of the units and measures parsing phase as defined in ERO003
- Conversion clauses are captured as child symbol table scopes, not stored in ParserUnitDefinition
- Pattern and calculation pairing happens during analysis phase
- Formula evaluation depends on metrics implementation
- Follow Logical TDD Lifecycle: write tests before implementing code for each task
