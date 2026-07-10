# Story123 Walkthrough: CPL Concepts Parser - Range

## Overview

This document describes the implementation of Int range generator support in the CPL parser (Story123). This story adds the `Int[Start, End, Step]` generator as an alternative to `Values` for defining numeric concept values.

## Changes Made

### Grammar Changes

#### `src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguageLexer.g4`

- Added `INT: 'Int';` keyword token to the reserved keywords section.

#### `src/chronocone_planning_language/src/main/antlr/CPLConcepts.g4`

- Modified `generatorPattern` to accept both `valuesPatternGenerator` and `intRangeGenerator`.
- Added `intRangeGenerator` rule: `INT LBRACKET intRangeParams RBRACKET`.
- Added `intRangeParams` rule: `INTEGER COMMA INTEGER (COMMA INTEGER)?`.
  - Supports both 2-parameter (`Int[1, 5]`) and 3-parameter (`Int[1, 12, 1]`) forms.

### AST Changes

#### `src/chronocone_planning_language/src/main/go/ast/cpl_concepts.go`

- Added `IntRangeGenerator` constant to the `GeneratorType` enum.
- Added `IntRangeGeneratorDef` struct with `Start`, `End`, and `Step` fields.

### Symbol Table Changes

#### `src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go`

- Added `IntGeneratorScope` to the `ScopeType` enum.
- Added `IntGeneratorScope` case to `ScopeType.String()`.
- Added `Data interface{}` field to `SymbolTable` to allow attaching scope-specific data (e.g., `IntRangeGeneratorDef`).

### Parser Listener Changes

#### `src/chronocone_planning_language/src/main/go/parser/cpl_concepts_listener.go`

- Added `EnterIntRangeGenerator` stub: pushes `IntGeneratorScope` onto the symbol stack.
- Implemented `ExitIntRangeGenerator`:
  - Extracts integer parameters from `IntRangeParamsContext`.
  - Defaults `Step` to `1` when only 2 parameters are provided.
  - Validates `step != 0`.
  - Validates `start < end` when `step > 0`.
  - Validates `start > end` when `step < 0`.
  - Stores `*ast.IntRangeGeneratorDef` in the scope's `Data` field.
  - Reports semantic errors via `l.AddError()` with source location information.

### ANTLR Regeneration

- Regenerated Go parser code via `./gradlew generateGrammarSource`.
- New generated types: `IntRangeGeneratorContext`, `IntRangeParamsContext`.

### Tests

#### `src/chronocone_planning_language/src/main/go/parser/cpl_concepts_range_test.go`

Created new test file with 4 test functions:

1. **`TestParseIntRangeGenerator_BasicRanges`** — Verifies parsing succeeds for:
   - `Month=[Int[1, 12, 1]]`
   - `FiscalPeriod=[Int[1, 13, 1]]`
   - `Numbers=[Int[1, 5]]` (2 params, step defaults)
   - `EvenNumbers=[Int[2, 10, 2]]`
   - `Mixed=[Values[A, B], Int[1, 3, 1]]`

2. **`TestParseIntRangeGenerator_SymbolTableStructure`** — Verifies the symbol table hierarchy:
   - `ConceptsScope` → `ConceptScope` → `IntGeneratorScope`
   - `IntGeneratorScope.Data` contains correct `*ast.IntRangeGeneratorDef`

3. **`TestParseIntRangeGenerator_TwoParamsDefaultsStep`** — Verifies that `Int[1, 5]` produces `Step=1`.

4. **`TestParseIntRangeGenerator_ValidationErrors`** — Verifies validation catches:
   - Step zero: `Int[1, 5, 0]`
   - Start >= end with positive step: `Int[5, 1, 1]`

## Test Results

All tests pass with `./gradlew test`:

- New Int range tests: 4 functions, 9 subtests — all PASS
- Existing tests: no regressions
- `go vet` passes

## Design Decisions

### Step Defaulting

The grammar allows both 2 and 3 INTEGER parameters. The listener defaults `Step` to `1` when only 2 are present. This keeps the grammar simple while providing a sensible default.

### Scope Data Attachment

A `Data interface{}` field was added to `SymbolTable` to allow attaching structured data (like `IntRangeGeneratorDef`) directly to scopes. This is cleaner than creating anonymous symbol definitions for numeric parameters and aligns with the symbol table-driven parsing pattern.

### Validation in Listener vs. Analyzer

Range bound validation is performed in the parser listener (`ExitIntRangeGenerator`) rather than deferring to the analyzer. This is appropriate because:

- The validation rules are syntactic/structural (step != 0, start < end).
- Errors can be reported with precise source locations during parsing.
- The analyzer can then assume valid range data.

## Post-Mortem

### What Went Well

- The symbol table-driven parsing pattern made the extension straightforward.
- Existing test patterns (table-driven, `t.Parallel()`, `parser_test` package) were easy to follow.
- ANTLR grammar changes were minimal and composable.

### Challenges

- Initial Phase A compilation errors were expected (missing ANTLR-generated types). Regeneration in Phase C resolved this cleanly.
- Determining where to store `IntRangeGeneratorDef` (scope `Data` vs. child definitions) required a brief design decision. Scope `Data` was chosen for simplicity.

### Lessons Learned

- Adding a new generator type to the CPL parser requires coordinated changes across 5+ files (grammar, AST, symbol table, listener, tests). Having a clear checklist (the story tasks) prevented omissions.
- The `Data interface{}` pattern on `SymbolTable` is a useful escape hatch for scope-specific structured data without over-complicating the symbol table schema.
