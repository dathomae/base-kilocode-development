# Story196: CPL Types - String and Boolean Support

## Goal

Add String and Boolean built-in type support to the CPL analysis phase, including `StringConstraints` (minimum/maximum length, regex pattern with compile-time validation), `BooleanConstraints` (no constraints allowed), and all necessary integration points in `CreateConstraints`, `createTypeReference`, `parseBuiltInTypeName`, constraint validation, and constraint merging. This combines and replaces the remaining work from Story133 and Story134, adapted to the symbol-table-driven parsing pattern where the parser already accepts String/Boolean syntactically via the generic `baseType: identifier;` rule.

This is Increment 2 of [Epic002: CPL Types Implementation](../epics/Epic002_cpl-types-implementation.md).

## References

- [Epic002_cpl-types-implementation.md](../epics/Epic002_cpl-types-implementation.md) - Parent epic for types implementation
- [ERO002_types-incremental-implementation.md](../specifications/design/cpl/eros/ERO002_types-incremental-implementation.md) - Main incremental implementation ERO (Increment 2)
- [ERO002_types-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md) - Analysis phase data structures for Increment 2
- [types-definitions.md](../specifications/design/cpl/types-definitions.md) - Comprehensive type design
- [lessons-learned.md](../lessons-learned.md) - "Do not hard-code built-in type names in parser grammar rules"
- [symbol-table-driven-parsing.md](../specifications/patterns/symbol-table-driven-parsing.md) - Architectural pattern

## Dependencies

- [Story131_cpl-types-parser-number.md](finished-stories/Story131_cpl-types-parser-number.md) - Done (types section parsing, TypesScope/TypeScope)
- [Story132_cpl-types-analyzer-number.md](finished-stories/Story132_cpl-types-analyzer-number.md) - Done (analysis infrastructure, NumberConstraints, CreateConstraints, derivation paths)

## Dependent Stories

- [Story135_cpl-types-parser-multiple-sections.md](Story135_cpl-types-parser-multiple-sections.md) - Not Started
- [Story136_cpl-types-analyzer-multiple-sections.md](Story136_cpl-types-analyzer-multiple-sections.md) - Not Started

## Tasks

1. Add StringType and BooleanType to BuiltInType enum - Completed
   a. Add `StringType` and `BooleanType` constants to [`BuiltInType`](src/chronocone_planning_language/src/main/go/analysis/cpl_types.go:17) enum - Completed
   b. Update `String()` method to return "String" and "Boolean" for new types - Completed

   **Intent**: Extend the built-in type enumeration to recognize String and Boolean as first-class base types.
   **Acceptance Criteria**: `BuiltInType.StringType.String()` returns "String", `BuiltInType.BooleanType.String()` returns "Boolean".

2. Implement StringConstraints struct with full constraint interface - Completed
   a. Define `StringConstraints` with `Minimum` (*int), `Maximum` (*int), and `Pattern` (\*regexp.Regexp) fields - Completed
   b. Implement `isTypeConstraints()` marker method - Completed
   c. Implement `Validate()` checking: min >= 0, max >= 0, min <= max when both specified - Completed
   d. Implement `ConstraintNames()` returning list of set constraint names - Completed
   e. Implement `GetConstraint(name)` returning (value, exists) for a named constraint - Completed
   f. Implement `IsMoreOrEqualRestrictive(constraintName, currentValue, newValue)` for derivation chain validation:
   - For "minimum": higher value is more restrictive (e.g., 10 is more restrictive than 0)
   - For "maximum": lower value is more restrictive (e.g., 50 is more restrictive than 100)
   - For "pattern": any non-nil pattern is more restrictive than nil; same pattern is equal; different pattern is not more restrictive - Completed

   **Intent**: Provide complete constraint storage and validation for String types, following the same interface pattern as `NumberConstraints`, `InstantConstraints`, and `IntervalConstraints`.
   **Acceptance Criteria**: StringConstraints correctly validates length constraints, exposes constraint names, supports constraint lookup, and supports restrictiveness comparison for type derivation.

3. Implement BooleanConstraints struct with full constraint interface - Completed
   a. Define `BooleanConstraints` as empty struct (no fields) - Completed
   b. Implement `isTypeConstraints()` marker method - Completed
   c. Implement `Validate()` that always returns nil - Completed
   d. Implement `ConstraintNames()` returning empty slice - Completed
   e. Implement `GetConstraint(name)` returning (nil, false) for all names - Completed
   f. Implement `IsMoreOrEqualRestrictive(constraintName, currentValue, newValue)` returning error for unknown names - Completed

   **Intent**: Provide constraint structure for Boolean types following the same interface pattern, even though Boolean has no constraints.
   **Acceptance Criteria**: BooleanConstraints satisfies the TypeConstraints interface and always validates successfully.

4. Add String and Boolean cases to CreateConstraints factory - Completed
   a. Add `StringType` case calling new `createStringConstraints` helper - Completed
   b. Add `BooleanType` case calling new `createBooleanConstraints` helper - Completed
   c. Implement `createStringConstraints` with:
   - Duplicate constraint name detection - Completed
   - Constraint name validation (only "minimum", "maximum", "pattern" are valid) - Completed
   - Value type validation: minimum/maximum must parse as integers via `strconv.Atoi()` - Completed
   - Pattern validation: compile with `regexp.Compile()`, reject empty string patterns, store compiled `*regexp.Regexp` - Completed
   - Call `Validate()` for semantic consistency (min <= max) - Completed
     d. Implement `createBooleanConstraints` that rejects any constraints with an error - Completed

   **Intent**: Enable the analysis phase to convert raw parser constraints into validated StringConstraints and BooleanConstraints.
   **Acceptance Criteria**: String constraints are correctly parsed and validated; Boolean rejects any constraints; duplicate constraint names are detected; invalid regex patterns produce clear error messages.

5. Add "String" and "Boolean" to createTypeReference and parseBuiltInTypeName - Completed
   a. Add `"String"` case to [`createTypeReference()`](src/chronocone_planning_language/src/main/go/analysis/types_analysis.go:176) returning `NewBuiltInTypeRef(StringType)` - Completed
   b. Add `"Boolean"` case to `createTypeReference()` returning `NewBuiltInTypeRef(BooleanType)` - Completed
   c. Add `"String"` case to [`parseBuiltInTypeName()`](src/chronocone_planning_language/src/main/go/analysis/cpl_analyzer.go:639) returning `StringType` - Completed
   d. Add `"Boolean"` case to `parseBuiltInTypeName()` returning `BooleanType` - Completed

   **Intent**: Ensure String and Boolean base type names are recognized as built-in types during analysis, enabling derivation path computation and constraint validation.
   **Acceptance Criteria**: Types declared with `String` or `Boolean` base type are correctly identified as built-in types.

6. Add String and Boolean cases to constraint validation and merging - Completed
   a. Add `*StringConstraints` case to [`validateMoreRestrictive()`](src/chronocone_planning_language/src/main/go/analysis/cpl_analyzer.go:388) - Completed
   b. Add `*BooleanConstraints` case to `validateMoreRestrictive()` - Completed
   c. Add `*StringConstraints` case to [`mergeConstraints()`](src/chronocone_planning_language/src/main/go/analysis/cpl_analyzer.go:483) - Completed
   d. Add `*BooleanConstraints` case to `mergeConstraints()` - Completed
   e. Implement `validateStringMoreRestrictive` helper using `ConstraintNames()`/`GetConstraint()`/`IsMoreOrEqualRestrictive()` - Completed
   f. Implement `validateBooleanMoreRestrictive` helper (Boolean has no constraints, so any new constraints are an error) - Completed
   g. Implement `mergeStringConstraints` helper taking the more restrictive value for each field - Completed
   h. Implement `mergeBooleanConstraints` helper (trivially returns the new constraints since Boolean has no fields) - Completed

   **Intent**: Enable type derivation chains for String and Boolean types, ensuring derived types only narrow constraints.
   **Acceptance Criteria**: String type derivation chains correctly validate and merge constraints; Boolean derivation chains reject any constraint additions.

7. Write comprehensive unit tests for String and Boolean types - Completed
   a. Test `StringConstraints.Validate()` with valid and invalid length combinations - Completed
   b. Test `StringConstraints.ConstraintNames()` and `GetConstraint()` - Completed
   c. Test `StringConstraints.IsMoreOrEqualRestrictive()` for minimum, maximum, and pattern - Completed
   d. Test `BooleanConstraints` satisfies TypeConstraints interface and Validate always succeeds - Completed
   e. Test `CreateConstraints(StringType, ...)` with valid constraints (minimum, maximum, pattern individually and combined) - Completed
   f. Test `CreateConstraints(StringType, ...)` with invalid constraints (unknown names, non-integer values, invalid regex, empty pattern, duplicates) - Completed
   g. Test `CreateConstraints(BooleanType, ...)` rejecting any constraints - Completed
   h. Test `createTypeReference("String")` and `createTypeReference("Boolean")` return correct BuiltInTypeRef - Completed
   i. Test `parseBuiltInTypeName("String")` and `parseBuiltInTypeName("Boolean")` return correct BuiltInType - Completed
   j. Test full parse+analyze pipeline with String type definitions (e.g., `ShortString = String(maximum=50)`) - Completed
   k. Test full parse+analyze pipeline with Boolean type definitions (e.g., `ActiveFlag = Boolean()`) - Completed
   l. Test String type derivation chain with constraint narrowing - Completed
   m. Test String type derivation chain with constraint expansion (error) - Completed
   n. Test Boolean type derivation chain (no constraints to narrow) - Completed

   **Intent**: Verify all String and Boolean type functionality works correctly end-to-end.
   **Acceptance Criteria**: All tests pass, covering valid inputs, error conditions, and integration with the full parse+analyze pipeline.

## Notes

- **No grammar changes**: The parser already accepts String and Boolean via the generic `baseType: identifier;` rule. This follows the lesson-learned "Do not hard-code built-in type names in parser grammar rules."
- **No parser listener changes**: The symbol-table-driven parser only manages scopes. All type interpretation happens in analysis.
- **String constraints**: minimum (length), maximum (length), pattern (regex). Pattern is compiled to `*regexp.Regexp` during analysis for syntax validation.
- **Regex validation at analysis time**:
  1. Syntax validation via `regexp.Compile()` — catches malformed patterns (unmatched brackets, invalid escapes, etc.)
  2. Empty pattern rejection — an empty string `""` matches everything and provides no constraint value
  3. Store compiled `*regexp.Regexp` for later runtime use
  4. Go's RE2 engine guarantees linear-time matching — no catastrophic backtracking concerns
- **Boolean constraints**: No constraints allowed. `Boolean()` is the required syntax (parentheses required by grammar).
- **No Instant or Interval changes**: Those are already implemented (Increments 4-5 were implemented ahead of Increment 2).
- **No multiple type sections**: Already supported by existing infrastructure.
- **Follow Logical TDD Lifecycle**: Write tests before implementing code for each task.
- **Pattern**: Follow existing constraint interface pattern (`isTypeConstraints()`, `Validate()`, `ConstraintNames()`, `GetConstraint()`, `IsMoreOrEqualRestrictive()`) established by `NumberConstraints`, `InstantConstraints`, and `IntervalConstraints`.
- **Supersedes**: This story combines and replaces [Story133](Story133_cpl-types-parser-string-boolean.md) and [Story134](Story134_cpl-types-analyzer-string-boolean.md), which were written against an earlier parser-centric design that was superseded by the symbol-table-driven parsing pattern.
