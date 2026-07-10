# Story132: CPL Types Analyzer - Number Type (Increment 1)

## Goal

Implement the analysis phase for Number types including AnalysisTypeDefinition structure, Types lookup map, NumberConstraints with validation, CreateConstraints factory function, and duplicate detection. This is the first increment of Epic002 types analysis, building on the parser infrastructure from Story131.

## Pre-Implemented Types

The following type declarations have been pre-implemented and exist in the codebase at `src/chronocone_planning_language/src/main/go/analysis/cpl_types.go`:

- `BuiltInType` enum with `NumberType` constant
- `Constraint` struct for raw constraint name-value pairs
- `NumberConstraints` struct with Precision, Minimum, Maximum pointer fields
- `TypeDefinition` struct with Name, BuiltInType, and Constraints fields

Tasks 1, 2, and 9a-c are marked as completed since these type declarations are already present.

## References

- [Epic002_cpl-types-implementation.md](../epics/Epic002_cpl-types-implementation.md) - Parent epic for types implementation
- [ERO002_types-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md) - Analysis phase data structures for Increment 1
- [ERO002_types-incremental-implementation.md](../specifications/design/cpl/eros/ERO002_types-incremental-implementation.md) - Main incremental implementation ERO
- [Story131_cpl-types-parser-number.md](Story131_cpl-types-parser-number.md) - Parser increment 1 (dependency)

## Dependencies

- [Story131_cpl-types-parser-number.md](Story131_cpl-types-parser-number.md) - Completed (parser provides TypesScope and ParserTypeDefinition)

## Dependent Stories

- [Story133_cpl-types-parser-string-boolean.md](Story133_cpl-types-parser-string-boolean.md) - Not Started
- [Story134_cpl-types-analyzer-string-boolean.md](Story134_cpl-types-analyzer-string-boolean.md) - Not Started

## Tasks

1. Define AnalysisTypeDefinition and Types lookup map in AnalysisDefinitions - **Completed**
   a. Add Types map[string]AnalysisTypeDefinition to AnalysisDefinitions struct - **Completed** (TypeDefinition already in cpl_types.go)
   b. Define AnalysisTypeDefinition struct with Name, BaseType, Constraints, DerivationPath fields - **Completed**
   c. Define TypeConstraints interface with isTypeConstraints() and Validate() methods - Not Started

   **Intent**: Create the core analysis data structures for type definitions.
   **Acceptance Criteria**: Types can be stored in the analysis definitions with proper type information.

2. Implement NumberConstraints struct with Validate method - **Partially Completed**
   a. Define NumberConstraints struct with Precision, Minimum, Maximum pointer fields - **Completed**
   b. Implement isTypeConstraints() marker method - Not Started
   c. Implement Validate() method checking precision >= 0 and minimum < maximum - Not Started

   **Intent**: Provide constraint validation for Number types.
   **Acceptance Criteria**: NumberConstraints validates correctly, rejecting invalid constraint combinations.

3. Implement CreateConstraints factory function with semantic validation - Not Started
   a. Create CreateConstraints(baseType, parserConstraints) function returning TypeConstraints - Not Started
   b. Implement createNumberConstraints helper to convert raw parser constraints to NumberConstraints - Not Started
   c. Validate constraint names are valid for Number (precision, minimum, maximum) - Not Started
   d. Check for duplicate constraint names in parserConstraints - Not Started
   e. Parse constraint values from raw strings to appropriate types (integers for Number constraints) - Not Started
   f. Handle unknown constraint names and invalid values with appropriate errors - Not Started

   **Intent**: Convert parser-level raw constraint pairs to analysis-level constraints with semantic validation.
   **Acceptance Criteria**: Parser constraints are correctly interpreted, duplicate names detected, constraint names validated, and values parsed for Number types.

4. Implement BuildDerivationPath function - Not Started
   a. Create BuildDerivationPath(typeName, types) function - Not Started
   b. For Increment 1, return path [typeName, baseTypeName] since types derive directly from built-ins - Not Started

   **Intent**: Track the derivation chain from derived types to their built-in roots.
   **Acceptance Criteria**: Derivation paths are correctly computed for Number types.

5. Implement function-based handlers for types analysis - Not Started
   a. Create EnterAnalyzeTypesScope function taking (*symboltable.SymbolTable, *AnalysisContext, *CPLAnalyzer) parameters - Not Started
   b. Create ExitAnalyzeTypesScope function taking (*symboltable.SymbolTable, *AnalysisContext, *CPLAnalyzer) parameters - Not Started
   c. Implement type definition processing in EnterAnalyzeTypesScope to build AnalysisTypeDefinition entries - Not Started
   d. Register handlers with WalkSymbolTableWithHandlers - Not Started

   **Intent**: Build the analyzer handlers that process parsed type definitions into validated analysis structures.
   **Acceptance Criteria**: Handlers process TypesScope symbol tables and produce validated AnalysisTypeDefinition entries.

   **Pattern**: Use function-based handlers (not method-based) matching existing analysis pattern: `EnterAnalyzeTypesScope(table, context, analyzer)`.

6. Implement duplicate detection during symbol table walk - Not Started
   a. Check if type name already exists in Types map before adding - Not Started
   b. Return error for duplicate type definitions (including across multiple types sections) - Not Started
   c. Document that duplicate detection happens across all TypesScope tables during walk - Not Started

   **Intent**: Prevent duplicate type definitions within the same compilation unit, including across multiple types sections.
   **Acceptance Criteria**: Duplicate type definitions are detected and reported as errors, including cross-section duplicates.

7. Integrate types analysis handlers with CPLCompiler.Analyze method - Not Started
   a. Register EnterAnalyzeTypesScope and ExitAnalyzeTypesScope with WalkSymbolTableWithHandlers - Not Started
   b. Add Types to AnalysisDefinitions return value - Not Started

   **Intent**: Wire the type analysis handlers into the main compilation pipeline.
   **Acceptance Criteria**: Types are analyzed as part of the standard compilation flow.

8. Add TypesScope to ScopeType enum (if not already added by parser story) - Not Started
   a. Verify TypesScope constant exists in ScopeType enum - Not Started
   b. Add if missing - Not Started

   **Intent**: Ensure the scope type system recognizes types sections.
   **Acceptance Criteria**: TypesScope is available for symbol table operations.

9. Write comprehensive unit tests for types analysis - **Partially Completed**
   a. Test NumberConstraints.Validate() with valid and invalid constraints - Not Started
   b. Test CreateConstraints with various Number constraint combinations - Not Started
   c. Test AnalyzeTypesScope with single type definition - Not Started
   d. Test duplicate type detection - Not Started
   e. Test BuildDerivationPath for Number types - Not Started
   f. Test integration with WalkSymbolTableWithHandlers - Not Started

   **Already Implemented**: Basic type declaration tests exist in `src/chronocone_planning_language/src/main/go/analysis/cpl_types_test.go`:
   - TestBuiltInType - **Completed**
   - TestConstraint - **Completed**
   - TestNumberConstraints (struct creation) - **Completed**
   - TestTypeDefinition - **Completed**

   **Intent**: Verify all analysis functionality works correctly and errors are properly reported.
   **Acceptance Criteria**: All tests pass, covering valid inputs and error conditions.

## Notes

- This story implements Increment 1 of the types analysis phase as defined in ERO002
- Types can only derive directly from built-in types in this increment (no type extension)
- Only Number base type is supported
- Built-in types are pre-populated before analysis begins
- **Parsing/Analysis Separation**: Story131 (parsing) captures raw constraint pairs (identifier=integer). This story (analysis) interprets those constraints: validates constraint names, parses values, checks for duplicates
- Handlers use function-based pattern: `EnterAnalyzeTypesScope(table, context, analyzer)` not method-based
- Duplicate detection includes cross-section duplicates during symbol table walk
- Follow Logical TDD Lifecycle: write tests before implementing code for each task
- **SourceLocation**: Use `symboltable.SourceLocation` consistently (eliminate duplicate in ast package)
- Type declarations (BuiltInType, Constraint, NumberConstraints, TypeDefinition) are pre-implemented and exist at `go/analysis/cpl_types.go`
