# Story164: CPL Concepts Value Listener

## Goal

Update the CPL parser and analysis pipeline to support both identifier-style concept values (e.g., `USA`, `Canada`) and numeric literal values (e.g., `2024`, `2025`) within concept value lists. This requires a grammar change to the `value` rule, updates to the custom listener to handle both token types, and comprehensive tests across parser, listener, and analysis layers. This is part of [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md).

## References

- [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md)
- [ERO001: Concepts Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO001_concepts-parsing-phase.md)
- [ERO001: Concepts Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO001_concepts-analysis-phase.md)
- [Concept Definitions Design](../specifications/design/cpl/concepts-definitions.md)

## Dependencies

- [Story114_cpl-concepts-parser-basic.md](finished-stories/Story114_cpl-concepts-parser-basic.md) - Done - Basic concepts parser with Values generator support
- [Story115_cpl-concepts-analyzer-basic.md](finished-stories/Story115_cpl-concepts-analyzer-basic.md) - Done - Basic concepts analyzer with duplicate detection

## Dependent Stories

None

## Tasks

1. Update grammar rule and regenerate ANTLR artifacts - Completed
   a. Change `value` rule in `ChronoconePlanningLanguage.g4` from `value: VALUE;` to `value: IDENTIFIER | VALUE | INTEGER;` - Completed
   b. Regenerate Go parser, listener, base listener, and visitor from updated `.g4` files - Completed
   c. Verify generated `ValueContext` exposes `IDENTIFIER()`, `VALUE()`, and `INTEGER()` terminal node accessors - Completed

   **Intent**: Allow the parser to accept both `IDENTIFIER` tokens (for traditional names) and `VALUE` tokens (for numeric literals) in concept value lists without changing lexer priorities.
   **Acceptance Criteria**: Generated code compiles and `ValueContext` has accessors for both token types.

2. Implement EnterValue/ExitValue in custom listener with tests - Completed
   a. Write listener tests for identifier values (e.g., `Values[USA, Canada]`) - Completed
   b. Write listener tests for numeric values (e.g., `Values[2024, 2025]`) - Completed
   c. Add `EnterValue` that peeks the current symbol table and extracts text from the matched token (`IDENTIFIER`, `VALUE`, or `INTEGER`) - Completed
   d. Store extracted text as a definition with type `symboltable.Identifier` - Completed
   e. Add `ExitValue` for logging symmetry - Completed
   f. Run listener tests and verify they pass - Completed

   **Intent**: Ensure the custom listener correctly captures both identifier and numeric concept values and stores them in the symbol table with consistent typing.
   **Acceptance Criteria**: Listener tests pass for both identifier and numeric values, and symbol table entries use `symboltable.Identifier` type.

3. Add parser-level tests for numeric concept values - Completed
   a. Write parser test for numeric values (`Year=[Values[2024, 2025]]`) - Completed
   b. Write parser test for mixed identifier and numeric values if supported by grammar - Completed
   c. Verify existing identifier-based concept tests still pass - Completed

   **Intent**: Verify the grammar change enables correct parsing of concept declarations containing numeric values.
   **Acceptance Criteria**: All parser tests pass, including new numeric value tests and existing identifier-based tests without modification.

4. Add analysis-level tests for numeric concept values - Completed
   a. Write analysis test where value generator scopes contain numeric strings - Completed
   b. Verify `AnalyzeValueGeneratorScope` builds correct `ValuesGenerator` with numeric entries - Completed
   c. Verify `AnalyzeValueGeneratorScope` builds correct `ValuesGenerator` with alphabetic entries - Completed
   d. Confirm no regressions in existing concept analysis tests - Completed

   **Intent**: Verify the analysis phase correctly processes numeric concept values and produces valid `ValuesGenerator` definitions.
   **Acceptance Criteria**: Analysis tests pass for both alphabetic and numeric values, and all existing analysis tests continue to pass.

5. Run full test suite and verify completion criteria - Completed
   a. Execute all parser tests - Completed
   b. Execute all analysis tests - Completed
   c. Confirm no regressions in existing identifier-based concept tests - Completed
   d. Verify completion criteria from plan are met - Completed

   **Intent**: Ensure the entire change set is correct and does not introduce regressions.
   **Acceptance Criteria**: All tests pass, `Values[USA, Canada]` and `Values[2024, 2025]` both parse correctly, and the symbol table contains value definitions with `symboltable.Identifier` type.

## Constraints

- Must not change lexer priorities; `IDENTIFIER` and `VALUE` tokens remain in their current lexer order
- Must integrate with existing ANTLR-based parser infrastructure
- Must use existing symbol table stack infrastructure in the listener
- Must not break existing concept parsing or analysis functionality
- Follow the Logical TDD Lifecycle - write tests before implementation within each task
- Each task must include its own test-and-implement cycle

## Intent

The goal of this story is to make the CPL parser resilient to the grammar change that allows concept value lists to contain both traditional identifier-style values and numeric literals. Because the `VALUE` lexer token is defined after `IDENTIFIER`, alphabetic strings like `USA` are lexed as `IDENTIFIER` tokens. By updating the parser rule to accept both token types, we enable numeric concept values without breaking existing identifier-based values. The listener and analyzer must continue to store and process these values correctly.

## Acceptance Criteria

1. `Values[USA, Canada]` continues to parse correctly
2. `Values[2024, 2025]` parses correctly
3. The symbol table contains value definitions with `symboltable.Identifier` type
4. `AnalyzeValueGeneratorScope` produces correct `ValuesGenerator` entries for both alphabetic and numeric values
5. All existing tests pass without modification
6. New tests covering numeric values pass

## Notes

**Grammar Change Summary:**

```antlr
// Before
value: VALUE;

// After
value: IDENTIFIER | VALUE;
```

This change allows the `conceptValueList` rule (already changed to `value (COMMA value)*`) to accept both `IDENTIFIER` and `VALUE` tokens.

**Files Affected:**

- `src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguage.g4`
- Generated ANTLR files under `src/chronocone_planning_language/src/main/go/parser/antlr/`
- `src/chronocone_planning_language/src/main/go/parser/cpl_document_listener.go`
- `src/chronocone_planning_language/src/main/go/parser/cpl_concepts_parser_test.go`
- `src/chronocone_planning_language/src/main/go/parser/cpl_concepts_listener_test.go`
- `src/chronocone_planning_language/src/main/go/analysis/concept_analysis_test.go`

**Example CPL that should parse:**

```cpl
concepts
    Country=[Values[USA, Canada, Mexico]]
    Year=[Values[2024, 2025, 2026]]
```
