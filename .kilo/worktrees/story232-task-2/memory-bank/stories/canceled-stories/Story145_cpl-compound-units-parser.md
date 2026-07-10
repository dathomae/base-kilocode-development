# Story145: CPL Compound Units Parser - Unit Algebra Expressions (Increment 3)

## Goal

Implement the parser for compound unit expressions using expression/term/factor grammar rules and postfix expression parsing via symbol table stack. This includes ParserUnitReference and ParserOperator types, and extending the measures grammar to use expression rules. This is Increment 3 of Epic003 units and measures parsing.

## References

- [Epic003_units-and-measures-implementation.md](../epics/Epic003_units-and-measures-implementation.md) - Parent epic for units and measures implementation
- [ERO003_units-measures-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO003_units-measures-parsing-phase.md) - Parsing phase data structures for Increment 3
- [ERO003_units-measures-incremental-implementation.md](../specifications/design/cpl/eros/ERO003_units-measures-incremental-implementation.md) - Main incremental implementation ERO
- [Postfix Expression Parsing Pattern](../specifications/patterns/postfix-expression-parsing.md) - Pattern for expression parsing
- [Symbol Table-Driven Parsing Pattern](../specifications/patterns/symbol-table-driven-parsing.md) - Architectural pattern for parsing
- [Story143_cpl-measures-parser-basic.md](Story143_cpl-measures-parser-basic.md) - Parser increment 2 (dependency)

### Phase Separation Note

This story implements the parsing phase. Raw syntactic data is captured into the symbol table; semantic structures are built by the analysis phase.

### Migration Reference

This story follows the ERO003 design which replaces the old UnitDecl AST approach with symbol table-driven parsing.

## Dependencies

- [Story143_cpl-measures-parser-basic.md](Story143_cpl-measures-parser-basic.md) - Must be completed (measures parsing infrastructure)

## Dependent Stories

- [Story146_cpl-compound-units-analyzer.md](Story146_cpl-compound-units-analyzer.md) - Not Started
- [Story147_cpl-unit-conversions-parser.md](Story147_cpl-unit-conversions-parser.md) - Not Started

## Tasks

1. Add expression/term/factor grammar rules to main grammar - Not Started
   a. Define expression rule with term, expression MULT term, expression DIV term alternatives - Not Started
   b. Define term rule with factor, term POW INTEGER alternatives - Not Started
   c. Define factor rule with identifier, LPAREN expression RPAREN alternatives - Not Started
   d. Define explicit operator rules (multOperator, divOperator, powOperator) for listener callbacks - Not Started

   **Intent**: Establish grammar rules for unit algebra expressions with proper operator precedence.
   **Acceptance Criteria**: Grammar parses compound unit expressions like `Length/Time^2` and `Length*Mass/Time`.

2. Add lexer tokens for expression operators - Not Started
   a. Add MULT token for '\*' - Not Started
   b. Add DIV token for '/' - Not Started
   c. Add POW token for '^' - Not Started
   d. Verify INTEGER token exists - Not Started

   **Intent**: Provide lexical analysis for expression operators.
   **Acceptance Criteria**: Lexer correctly tokenizes \*, /, ^, and integer literals.

3. Update CPLMeasures.g4 to use expression rule in measureDefinition - Not Started
   a. Change measureDefinition to use expression instead of simple identifier for unit specification - Not Started
   b. Verify measureDefinition rule: identifier EQ identifier LPAREN UNIT EQUALS expression RPAREN - Not Started

   **Intent**: Allow compound unit expressions in measure declarations.
   **Acceptance Criteria**: Measures can use compound unit expressions like `Number(unit=Length/Time)`.

4. Implement ParserUnitReference and ParserOperator structs - Not Started
   a. Define ParserUnitReference struct with Name field - Not Started
   b. Define ParserOperator struct with Operator (OperatorType) field - Not Started
   c. Define OperatorType enum with OpMultiply, OpDivide, OpPower constants - Not Started

   **Intent**: Create Go data structures for expression components in postfix order.
   **Acceptance Criteria**: Data structures can represent unit references and operators.

5. Add DefinitionType values for expression components - Not Started
   a. Add UnitReferenceDefinition to DefinitionType enum - Not Started
   b. Add MultiplyOperatorDefinition to DefinitionType enum - Not Started
   c. Add DivideOperatorDefinition to DefinitionType enum - Not Started
   d. Add PowerOperatorDefinition to DefinitionType enum - Not Started

   **Intent**: Extend symbol table to recognize expression component definitions.
   **Acceptance Criteria**: Symbol table can store unit references and operators as definitions.

6. Implement expression listener methods in MeasuresListener - Not Started
   a. Implement ExitIdentifier to push ParserUnitReference to definitions slice - Not Started
   b. Implement ExitMultOperator to push ParserOperator{OpMultiply} to definitions slice - Not Started
   c. Implement ExitDivOperator to push ParserOperator{OpDivide} to definitions slice - Not Started
   d. Implement ExitPowOperator to push ParserOperator{OpPower} to definitions slice - Not Started
   e. Implement ExitInteger to push operand value to definitions slice - Not Started

   **Intent**: Build listener methods that produce postfix-ordered expression components.
   **Acceptance Criteria**: For `Length/Time^2`, definitions slice contains [Length, Time, Power, 2, Divide] in postfix order.

7. Implement validation during parsing for Increment 3 - Not Started
   a. Validate unit expression syntax - Not Started
   b. Validate parentheses matching - Not Started
   c. Validate power exponent is an integer - Not Started
   d. Validate unit reference syntax - Not Started

   **Intent**: Ensure compound unit expressions are syntactically valid.
   **Acceptance Criteria**: Invalid expressions produce clear error messages.

8. Write comprehensive unit tests for compound units parsing - Not Started
   a. Test parsing of simple multiplication: Length*Time - Not Started
   b. Test parsing of simple division: Length/Time - Not Started
   c. Test parsing of power: Length^2 - Not Started
   d. Test parsing of compound: Length/Time^2 - Not Started
   e. Test parsing of complex: Length*Mass/Time - Not Started
   f. Test parsing of parenthesized: (Length+Width)/Time - Not Started
   g. Test postfix ordering is correct for various expressions - Not Started
   h. Test error handling for invalid syntax - Not Started

   **Intent**: Verify compound unit expression parsing produces correct postfix order.
   **Acceptance Criteria**: All tests pass, covering valid inputs and error conditions.

## Notes

- This story implements Increment 3 of the units and measures parsing phase as defined in ERO003
- Expression rules are defined in the main grammar for reuse across CPL sections
- Parser uses postfix expression parsing via symbol table stack, not AST building
- No unit conversions in this increment
- Follow Logical TDD Lifecycle: write tests before implementing code for each task
