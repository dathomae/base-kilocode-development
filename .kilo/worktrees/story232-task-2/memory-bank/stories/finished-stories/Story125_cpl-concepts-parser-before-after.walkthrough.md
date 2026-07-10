# Story125 Walkthrough: CPL Concepts Parser - Before/After

## Overview

Story125 implements before/after clause support in the CPL parser for concept value ordering. This is part of [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md) — Increment 6. It adds syntax for inserting new values at specific positions when extending concepts, using `before` and `after` keywords, rather than only appending to the end.

## Task 1: Update ChronoconePlanningLanguageLexer.g4

### Goal

Add BEFORE and AFTER keyword tokens to the ANTLR lexer grammar.

### Implementation Summary

- Added `BEFORE: 'before';` and `AFTER: 'after';` to the RESERVED KEYWORDS section
- Placed near existing concept-related keywords (`DOTDOT`) for logical grouping
- Regenerated ANTLR Go code to include new token constants (`BEFORE=9`, `AFTER=10`)

### Files Modified

- `src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguageLexer.g4`
- `src/chronocone_planning_language/src/main/go/parser/antlr/chronoconeplanninglanguage_lexer.go` (generated)

### Tests Added

- `TestLexerRecognizesBeforeKeyword`
- `TestLexerRecognizesAfterKeyword`
- `TestLexerBeforeAfterAreNotIdentifiers`
- `TestLexerBeforeAfterInConceptContext`
- `TestLexerBeforeAfterCaseSensitivity`

### Key Decisions

- Keywords are case-sensitive ("before" is a keyword, "Before" is an identifier)
- Token placement ensures keyword precedence over IDENTIFIER/VALUE by appearing earlier in the grammar (ANTLR resolves ambiguity by prioritizing the first matching rule)

## Task 2: Update CPLConcepts.g4 for Before/After

### Goal

Add `conceptExtension` and `insertionPoint` parser rules to CPLConcepts.g4, and modify `conceptDefinition` to accept concept extensions as an alternative.

### Implementation Summary

- Added `conceptExtension` rule: `identifier EQUALS LBRACKET generatorPatternList RBRACKET insertionPoint`
- Added `insertionPoint` rule: `BEFORE identifier | AFTER identifier`
- Modified `conceptDefinition` to include `| conceptExtension` alternative within the `indent(...)*` wrapper
- Regenerated ANTLR Go parser code, producing `ConceptExtensionContext` and `InsertionPointContext` types
- Implemented 6 parser-level tests verifying BEFORE/AFTER parse trees, mixed sections, regression, and fallback behavior
- Fixed error message text in `cpl_concepts_concept_range_test.go` from `"mismatched input"` to `"no viable alternative at input"` due to grammar change

### Files Modified

- `src/chronocone_planning_language/src/main/antlr/CPLConcepts.g4` — 3 grammar rule changes
- `src/chronocone_planning_language/src/main/go/parser/antlr/chronoconeplanninglanguage_parser.go` (generated)
- `src/chronocone_planning_language/src/main/go/parser/antlr/chronoconeplanninglanguage_listener.go` (generated)
- `src/chronocone_planning_language/src/main/go/parser/antlr/chronoconeplanninglanguage_visitor.go` (generated)
- `src/chronocone_planning_language/src/main/go/parser/antlr/chronoconeplanninglanguage_base_listener.go` (generated)
- `src/chronocone_planning_language/src/main/go/parser/antlr/chronoconeplanninglanguage_base_visitor.go` (generated)
- `src/chronocone_planning_language/src/main/go/parser/cpl_concepts_before_after_test.go` — 6 parser tests
- `src/chronocone_planning_language/src/main/go/parser/cpl_concepts_concept_range_test.go` — error message fix

### Tests Added

- `TestParserBeforeClause`
- `TestParserAfterClause`
- `TestParserConceptExtensionFullSection`
- `TestParserInitialDefinitionStillWorks`
- `TestParserBeforeKeywordNotIdentifier`
- `TestParserFallbackWithoutInsertionPoint`

### Key Decisions

- Used `EQUALS` token (not `EQ`) for consistency with existing `conceptDefinition` rule
- ANTLR4 ALL(\*) lookahead resolves ambiguity between initial definition and extension alternatives automatically
- Concept extensions appear at same indentation level as initial definitions within the `indent(...)*` wrapper
- Test naming refined during review: `TestParserBeforeAfterKeywordsNotIdentifiers` → `TestParserBeforeKeywordNotIdentifier`, `TestParserErrorBeforeAfterBothMissing` → `TestParserFallbackWithoutInsertionPoint`

## Task 3: Define InsertionPoint struct

### Goal

Define InsertionPoint struct for before/after positioning.

### Implementation Summary

- Added struct to cpl_concepts.go with BeforeValue/AfterValue string fields
- Struct is a simple data holder capturing parsed insertion point syntax

### Files Modified

- `src/chronocone_planning_language/src/main/go/ast/cpl_concepts.go` — InsertionPoint struct definition
- `src/chronocone_planning_language/src/main/go/ast/cpl_concepts_test.go` — 4 unit tests for InsertionPoint

### Tests Added

- `TestInsertionPoint_BeforeOnly`
- `TestInsertionPoint_AfterOnly`
- `TestInsertionPoint_BothFields`
- `TestInsertionPoint_ZeroValue`

### Key Decisions

- Struct is a simple data holder, no methods needed; placed near other generator definition types

## Task 4: Update SymbolDefinition for insertion point data

### Goal

Create ConceptDefinitionData struct with InsertionPoint field for concept extension data storage.

### Implementation Summary

- Added ConceptDefinitionData struct to cpl_concepts.go with Name, InsertionPoint, and SourceLocation fields
- The InsertionPoint field enables concept extension data to flow from parser to analyzer via SymbolDefinition.Data

### Files Modified

- `src/chronocone_planning_language/src/main/go/ast/cpl_concepts.go` — ConceptDefinitionData struct definition
- `src/chronocone_planning_language/src/main/go/ast/cpl_concepts_test.go` — 4 unit tests for ConceptDefinitionData

### Tests Added

- `TestConceptDefinitionData_InitialDefinition`
- `TestConceptDefinitionData_ExtensionWithBefore`
- `TestConceptDefinitionData_ExtensionWithAfter`
- `TestConceptDefinitionData_SourceLocation`

### Key Decisions

- Struct uses ast.SourceLocation (package-local)
- InsertionPoint zero value indicates initial definition (no insertion point)

## Task 5: Update ConceptsListener

### Goal

Implement data extraction in the listener's `ExitInsertionPoint` and `ExitConceptExtension` methods, and write listener integration tests.

### Implementation Summary

- **`ExitInsertionPoint`**: Extracts `BEFORE`/`AFTER` token and anchor identifier from the ANTLR context. Builds an `ast.InsertionPoint` and stores it in the current scope's `Data` field (which is the ConceptScope pushed by `EnterConceptExtension`). Includes error handling for the case where neither BEFORE nor AFTER is present (defensive, should not happen with valid grammar).

- **`ExitConceptExtension`**: Pops the ConceptScope from the symbol stack. Extracts the concept name from the identifier context. Retrieves the `ast.InsertionPoint` stored by `ExitInsertionPoint` from the scope's `Data` field. Builds an `ast.ConceptDefinitionData` containing Name, InsertionPoint, and SourceLocation, then stores it back in the scope's `Data` field for the analysis phase.

- **Scope nesting**: The grammar `conceptDefinition` wraps both initial definitions and `conceptExtension`. For extensions, `EnterConceptDefinition` pushes an outer ConceptScope, then `EnterConceptExtension` pushes an inner ConceptScope. The InsertionPoint data is stored on the inner scope (the one popped by `ExitConceptExtension`), leaving the outer scope empty. This is consistent with the existing pattern where `ExitConceptPatternGenerator` and `ExitIntRangeGenerator` store data on their popped scopes.

### Files Modified

- `src/chronocone_planning_language/src/main/go/parser/cpl_concepts_listener.go` — implemented `ExitInsertionPoint` and `ExitConceptExtension`
- `src/chronocone_planning_language/src/main/go/parser/cpl_concepts_listener_test.go` — 4 listener integration tests

### Tests Added

- `TestListener_BeforeClauseCreatesInsertionPoint` — verifies parsing "before Low" produces `BeforeValue="Low"`
- `TestListener_AfterClauseCreatesInsertionPoint` — verifies parsing "after Medium" produces `AfterValue="Medium"`
- `TestListener_ConceptExtensionStoresData` — verifies `ConceptDefinitionData` is stored with correct Name, InsertionPoint, and SourceLocation
- `TestListener_InitialDefinitionNoInsertionPoint` — verifies regular concept definitions do not produce ConceptScope with data

### Key Decisions

- Data is stored on the inner ConceptScope (pushed by `EnterConceptExtension`) because that scope is popped by `ExitConceptExtension`, making it the natural place to attach concept-extension-specific data.
- The outer ConceptScope (pushed by `EnterConceptDefinition`) remains empty for extensions; for initial definitions it is also empty because `ExitConceptDefinition` does not set any data. This is acceptable because initial definition data extraction is not yet implemented (future increment).
- Tests use `findConceptScopeWithData` helper to locate the inner scope with non-nil Data, rather than assuming a fixed number of ConceptScopes, making tests robust against the double-scope nesting.

## Task 6: Implement Validation

### Goal

Add semantic validation guards to `ExitInsertionPoint` and `ExitConceptExtension` to ensure the extracted data is valid before storing it in the symbol table. These are belt-and-suspenders checks since the grammar largely enforces correctness at the syntax level.

### Implementation Summary

- **`ExitInsertionPoint` validation**: Three defensive checks were added:
  1. **Anchor identifier non-empty**: Validates that the identifier following `BEFORE`/`AFTER` is present and non-empty. The grammar (`insertionPoint: BEFORE identifier | AFTER identifier`) guarantees an identifier token is present, but this check guards against nil context or empty token text.
  2. **At least one direction specified**: Validates that at least one of `BeforeValue`/`AfterValue` is non-empty after extraction. The grammar ensures one `BEFORE`/`AFTER` token is present, but this check verifies the extracted value is meaningful.
  3. **Invalid data not stored**: A `hasValidationError` flag prevents storing incomplete `InsertionPoint` data when any validation fails. This follows the same pattern used in `ExitIntRangeGenerator`.

- **`ExitConceptExtension` validation**: One defensive check was added:
  1. **Concept name non-empty**: After extracting the concept name from the identifier context, validates that the name string is non-empty. The grammar (`conceptExtension: identifier EQUALS ...`) guarantees an identifier is present, but this guards against empty token text.

- **`hasValidationError` pattern**: Both methods use a local `hasValidationError` flag to track whether any semantic error was recorded. If set, the method returns early without storing invalid data. This is consistent with the pattern established in `ExitIntRangeGenerator` and `ExitConceptPatternGenerator`.

### Files Modified

- [`cpl_concepts_listener.go`](src/chronocone_planning_language/src/main/go/parser/cpl_concepts_listener.go) — added validation logic to `ExitInsertionPoint` (3 checks) and `ExitConceptExtension` (1 check)
- [`cpl_concepts_listener_test.go`](src/chronocone_planning_language/src/main/go/parser/cpl_concepts_listener_test.go) — 2 validation tests

### Tests Added

- `TestListener_BeforeAfterValidation_ValidInsertionPoint` — verifies that valid before/after with identifier produces NO semantic errors and correctly populates `ConceptDefinitionData`
- `TestListener_BeforeAfterValidation_InsertionPointMissingAnchor` — verifies that valid before/after input (with proper anchor identifiers) does not trigger false-positive validation errors; uses table-driven subtests for both before and after variants

### Key Decisions

- Validation checks are "belt-and-suspenders": the grammar enforces these constraints at the syntax level, so truly malformed input cannot reach the listener. The semantic checks serve as a second line of defense and document the expected invariants.
- The `hasValidationError` flag pattern prevents corrupt data from entering the symbol table. If any validation error is recorded, `scope.Data` is left unset, so downstream phases (analysis) will see a nil Data field and can report a missing definition error.
- `TestListener_BeforeAfterValidation_InsertionPointMissingAnchor` tests valid input because the grammar prevents empty identifiers from reaching `ExitInsertionPoint`. The test verifies the validation guards do not produce false positives and documents the defensive nature of the checks.
- All existing tests continue to pass, confirming that the validation changes are backward-compatible with Increment 1-5 functionality.

## Post-Mortem

(To be completed at end of story)
