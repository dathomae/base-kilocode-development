# Story131: CPL Types Parser - Number Type (Increment 1)

## Goal

Implement grammar changes for the types section in the CPL parser, add Number base type with precision/min/max constraint parsing, and integrate TypesScope into the symbol table. This is the first increment of Epic002: CPL Types Implementation, establishing the foundation for all type parsing.

## References

- [Epic002_cpl-types-implementation.md](../epics/Epic002_cpl-types-implementation.md) - Parent epic for types implementation
- [ERO002_types-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO002_types-parsing-phase.md) - Parsing phase data structures for Increment 1
- [ERO002_types-incremental-implementation.md](../specifications/design/cpl/eros/ERO002_types-incremental-implementation.md) - Main incremental implementation ERO
- [Types Definitions Design](../specifications/design/cpl/types-definitions.md) - Comprehensive type design

## Dependencies

- [Story110_cpl-symbol-table-implementation.md](Story110_cpl-symbol-table-implementation.md) - Completed (symbol table infrastructure)
- [Story102_cpl-parser-environment-setup.md](finished-stories/Story102_cpl-parser-environment-setup.md) - Completed (parser environment)

## Dependent Stories

- [Story132_cpl-types-analyzer-number.md](Story132_cpl-types-analyzer-number.md) - Not Started
- [Story133_cpl-types-parser-string-boolean.md](Story133_cpl-types-parser-string-boolean.md) - Not Started

## Tasks

1. Create CPLTypes.g4 grammar file with typesSection, typeDefinition, baseType (Number only), constraintList, constraint, and constraintValue rules - **Completed**
   a. Use `parser grammar CPLTypes` (following pattern from CPLConcepts.g4 and CPLUnits.g4) - **Completed**
   b. Define typesSection rule matching TYPES_KEYWORD followed by one or more typeDefinition rules - **Completed**
   c. Define typeDefinition rule: identifier EQ baseType LPAREN constraintList? RPAREN - **Completed**
   d. Define baseType rule with NUMBER_TYPE only for Increment 1 - **Completed**
   e. Define constraintList as comma-separated constraints - **Completed**
   f. Define constraint as identifier EQ constraintValue - **Completed**
   g. Define constraintValue accepting INTEGER only for Number type constraints - **Completed**

   **Intent**: Establish the grammar foundation for parsing type definitions with Number constraints.
   **Acceptance Criteria**: Grammar file parses valid type definitions like `Currency = Number(precision=2)` and rejects invalid syntax. - **Completed**

2. Add lexer tokens to ChronoconePlanningLanguageLexer.g4 for types support - **Completed**
   a. Add TYPES_KEYWORD token for 'types' keyword - **Completed**
   b. Add NUMBER_TYPE token for 'Number' base type - **Completed**
   c. Add punctuation tokens: LPAREN, RPAREN, EQ, COMMA - **Completed**
   d. Verify INTEGER and STRING_LITERAL tokens exist or add them - **Completed**

   **Intent**: Provide lexical analysis for type-related syntax elements.
   **Acceptance Criteria**: Lexer correctly tokenizes type section keywords, type names, and constraint syntax. - **Completed**

3. Update main grammar to import CPLTypes and add typesSection to statement rule - **Completed**
   a. Add CPLTypes to grammar imports - **Completed**
   b. Add typesSection as an alternative in the statement rule - **Completed**

   **Intent**: Integrate types parsing into the main CPL grammar.
   **Acceptance Criteria**: CPL files with types sections are accepted by the parser alongside existing units and concepts sections. - **Completed**

4. Implement BuiltInType, ConstraintType, and constraint data structures in Go - **Completed**
   a. Define BuiltInType enum with NumberType constant - **Completed**
   b. Define ConstraintType enum with PrecisionConstraint, MinimumConstraint, MaximumConstraint, PatternConstraint - **Completed**
   c. Define Constraint struct with Type, Name, and Value fields - **Completed**
   d. Define NumberConstraints struct with Precision, Minimum, Maximum pointer fields - **Completed**

   **Intent**: Create Go data structures to represent parsed type information.
   **Acceptance Criteria**: Data structures compile and can represent Number type constraints. - **Completed**

5. Implement ParserTypeDefinition struct with constraint capture - **Completed**
   a. Define ParserTypeDefinition with Name, BaseType, Constraints, SourceLocation fields - **Completed**
   b. Define Constraint struct with Name and Value fields (capturing raw identifier=integer pairs) - **Completed**
   c. Use symboltable.SourceLocation (eliminate duplicate in ast package) - **Completed**

   **Intent**: Provide a unified structure for parsed type definitions with raw constraint capture.
   **Acceptance Criteria**: ParserTypeDefinition can store type definitions with raw constraint pairs (identifier=integer). - **Completed**
   **Note**: Constraint interpretation and validation happens in Story132 (analysis phase), not during parsing. - **Completed**

6. Add TypesScope to ScopeType enum and TypeDefinition to DefinitionType enum - **Completed**
   a. Add TypesScope constant to ScopeType enum (if not already present) - **Completed**
   b. Add TypeDefinition constant to DefinitionType enum - **Completed**
   c. Consider adding TypeScope for individual type definitions (following ConceptScope pattern) - **Completed**

   **Intent**: Extend symbol table type system to recognize types sections and definitions.
   **Acceptance Criteria**: Symbol table can create and manage types scopes and type definitions. - **Completed**

7. Add types listener methods to CPLListener in cpl_types_listener.go - **Completed**
   a. Create cpl_types_listener.go file with methods on \*CPLListener - **Completed**
   b. Implement EnterTypesSection to push new TypesScope symbol table - **Completed**
   c. Implement ExitTypesSection to pop the types scope - **Completed**
   d. Implement EnterTypeDefinition to parse type name, base type, constraints, and add to symbol table - **Completed**
   e. Implement parseConstraint helper function to convert constraint context to Constraint struct - **Completed**

   **Intent**: Build the ANTLR listener methods that process type sections and populate the symbol table.
   **Acceptance Criteria**: Listener correctly processes type definitions and stores them in the symbol table with proper constraint parsing. - **Completed**

   **Pattern**: Follow cpl_concepts_listener.go structure - methods on unified CPLListener struct, not separate TypesListener struct. - **Completed**

8. Verify ANTLR grammar enforces syntactic constraints - **Completed**
   a. Verify type names must be valid identifiers (enforced by grammar) - **Completed**
   b. Verify base type must be Number keyword (enforced by grammar) - **Completed**
   c. Verify constraint syntax is identifier=integer format (enforced by grammar) - **Completed**

   **Intent**: Confirm the ANTLR grammar correctly rejects invalid syntax.
   **Acceptance Criteria**: Grammar tests verify that invalid syntax produces parse errors. - **Completed**

   **Note**: No explicit validation code needed - the ANTLR generated parser enforces syntax. Semantic validation (constraint name validity, constraint value interpretation, duplicate detection) happens in Story132 (analysis phase). - **Completed**

9. Write comprehensive unit tests for types parsing - **Completed**
   a. Test parsing of Number type with precision constraint only - **Completed**
   b. Test parsing of Number type with precision, minimum, and maximum constraints - **Completed**
   c. Test parsing of Number type with no constraints - **Completed**
   d. Test error handling for invalid base types - **Completed**
   e. Test error handling for invalid constraint names - **Completed**
   f. Test error handling for duplicate type definitions - **Completed**
   g. Test symbol table integration (TypesScope pushed/popped correctly) - **Completed**

   **Intent**: Verify all parsing functionality works correctly and errors are properly reported.
   **Acceptance Criteria**: All tests pass, covering valid inputs and error conditions. - **Completed**

## Notes

- This story implements Increment 1 of the types parsing phase as defined in ERO002
- Only Number base type is supported in this increment
- Type capture (forward references) should be supported
- Built-in types are pre-populated in the global scope before parsing begins
- **Parsing/Analysis Separation**: Constraint interpretation, semantic validation, and duplicate type detection happen in Story132 (analysis phase), not during parsing
- Listener methods follow unified `CPLListener` pattern (see `cpl_concepts_listener.go`)
- Follow Logical TDD Lifecycle: write tests before implementing code for each task
- **SourceLocation**: Use `symboltable.SourceLocation` throughout; eliminate duplicate `SourceLocation` in `ast/cpl_concepts.go` and update existing code to use the symboltable version
- **Grammar Pattern**: Follow `parser grammar` pattern (not `grammar`) used by CPLConcepts.g4 and CPLUnits.g4
- **Constraint Capture**: Parser captures raw identifier=integer pairs; interpretation happens in analysis phase (Story132)
