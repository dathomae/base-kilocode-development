# Story111: Remove Dimensions from CPL

## Goal

Remove the dimensions declaration section from the Chronocone Planning Language (CPL) parser. This includes removing the CPLDimensions.g4 grammar file, all Go code related to dimensions (AST types, listener callbacks, tests), and updating the main grammar and lexer to remove dimensions support. This prepares the CPL definition for upcoming changes where dimensions will be handled differently.

## References

- [CPL Requirements - Units and Dimensions](../../specifications/requirements/cpl/concepts-types-units-measures-syntax.md)
- [CPL Architecture - Units and Dimensions Analysis](../../specifications/architecture/cpl/analysis/units-and-dimensions.md)
- [CPL Architecture - Units and Dimensions Parsing](../../specifications/architecture/cpl/parsing/units-and-dimensions.md)
- [Story103_cpl-parser-units-and-dimensions.md](Story103_cpl-parser-units-and-dimensions.md) - Original story that added dimensions support (Done)

## Dependencies

None - This is a cleanup story that removes existing functionality in preparation for future changes.

## Dependent Stories

None

## Tasks

1. Verify baseline - units functionality has test coverage - Not Started
   a. Run existing tests to confirm they pass - Not Started
   b. Review existing units tests in [`src/test/go/parser/units_test.go`](src/chronocone_planning_language/src/test/go/parser/units_test.go) - Not Started
   c. Review existing units tests in [`src/test/go/ast/units_test.go`](src/chronocone_planning_language/src/test/go/ast/units_test.go) - Not Started

2. Remove dimensions-related test files - Not Started
   a. Delete [`src/test/go/ast/dimensions_test.go`](src/chronocone_planning_language/src/test/go/ast/dimensions_test.go) - Not Started
   b. Delete [`src/test/go/parser/dimensions_test.go`](src/chronocone_planning_language/src/test/go/parser/dimensions_test.go) - Not Started

3. Remove dimensions-related AST types - Not Started
   a. Delete [`src/main/go/ast/cpl_dimensions.go`](src/chronocone_planning_language/src/main/go/ast/cpl_dimensions.go) - Not Started
   b. Update [`src/main/go/ast/cpl_ast.go`](src/chronocone_planning_language/src/main/go/ast/cpl_ast.go) to remove Dimensions field from CPLDocument - Not Started
   c. Update [`src/main/go/ast/cpl_ast.go`](src/chronocone_planning_language/src/main/go/ast/cpl_ast.go) to remove AddDimension method - Not Started

4. Remove dimensions-related listener callbacks - Not Started
   a. Delete [`src/main/go/parser/cpl_dimensions_listener.go`](src/chronocone_planning_language/src/main/go/parser/cpl_dimensions_listener.go) - Not Started
   b. Update [`src/main/go/parser/chronocone_planning_language_listener.go`](src/chronocone_planning_language/src/main/go/parser/chronocone_planning_language_listener.go) to remove EnterDimensionsSection and ExitDimensionsSection methods - Not Started

5. Remove the CPLDimensions.g4 grammar file and update main grammar - Not Started
   a. Delete [`src/main/antlr/CPLDimensions.g4`](src/chronocone_planning_language/src/main/antlr/CPLDimensions.g4) - Not Started
   b. Update [`src/main/antlr/ChronoconePlanningLanguage.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguage.g4) to remove the import of CPLDimensions - Not Started
   c. Update [`src/main/antlr/ChronoconePlanningLanguage.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguage.g4) to remove dimensionsSection from statement rule - Not Started

6. Remove DIMENSIONS keyword from lexer - Not Started
   a. Update [`src/main/antlr/ChronoconePlanningLanguageLexer.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguageLexer.g4) to remove DIMENSIONS token - Not Started

7. Update or remove dimensions example files - Not Started
   a. Delete [`examples/without_errors/dimensions_only.cpl`](src/chronocone_planning_language/examples/without_errors/dimensions_only.cpl) - Not Started
   b. Update [`examples/without_errors/multiple_sections.cpl`](src/chronocone_planning_language/examples/without_errors/multiple_sections.cpl) to remove dimensions sections - Not Started
   c. Update [`examples/without_errors/units_and_dimensions.cpl`](src/chronocone_planning_language/examples/without_errors/units_and_dimensions.cpl) to remove dimensions sections - Not Started
   d. Delete [`examples/with_errors/empty_dimensions.cpl`](src/chronocone_planning_language/examples/with_errors/empty_dimensions.cpl) - Not Started

8. Regenerate ANTLR parser code - Not Started
   a. Run Gradle generateGrammarSource task to regenerate parser - Not Started
   b. Verify generated code no longer contains dimensions-related types - Not Started

9. Verify units functionality preserved - Not Started
   a. Run go test to verify all remaining tests pass - Not Started
   b. Verify units tests specifically pass (confirming removal didn't break units) - Not Started
   c. Verify examples without dimensions parse correctly - Not Started

10. Update documentation and references - Not Started
    a. Update [`src/chronocone_planning_language/toc.md`](src/chronocone_planning_language/toc.md) if it references dimensions - Not Started
    b. Update [`src/chronocone_planning_language/README.md`](src/chronocone_planning_language/README.md) if it documents dimensions - Not Started
    c. Update [`src/chronocone_planning_language/examples/README.md`](src/chronocone_planning_language/examples/README.md) if it describes dimensions examples - Not Started

## Test-First Development

This story follows the Logical TDD Lifecycle. For a removal story, the approach is:

1. **Phase A (Structural Definition)**: Task 1 establishes the baseline - verifying that units tests exist and pass before any removals
2. **Phase B (Logical Tests)**: The existing units tests serve as our "tests" - they must continue to pass after dimensions removal
3. **Phase C (Implementation)**: Tasks 2-8 perform the actual removal
4. **Phase D (Verification)**: Task 9 verifies the preserved functionality (units) still works correctly

The key principle is that we verify what should remain working BEFORE we remove code, then confirm it still works AFTER removal. The existing units tests are our safety net.

## Constraints

- Units functionality must be preserved - only dimensions should be removed
- The build must remain functional after all removals
- All remaining tests must pass
- ANTLR parser code must be regenerated after grammar changes

## Intent

The intent is to cleanly remove dimensions support from the CPL parser to prepare for a new approach to handling dimensional data. This is not a breaking change for the overall system as dimensions were only used in the parser, not in any production code.

## Acceptance Criteria

1. CPLDimensions.g4 file is deleted
2. ChronoconePlanningLanguage.g4 no longer imports CPLDimensions and has no dimensions-related rules
3. All Go code related to dimensions (AST types, listener callbacks) is removed
4. All test files specifically testing dimensions are removed
5. DIMENSIONS keyword is removed from the lexer
6. Example files containing dimensions are either deleted or updated to remove dimensions sections
7. ANTLR parser code is regenerated and contains no dimensions-related generated code
8. The project builds successfully
9. All remaining tests pass, especially units tests
10. Units functionality continues to work correctly (verified by passing tests)

## Notes

- This story is a prerequisite for upcoming changes to how dimensions will be defined in CPL
- The removal is intentional and prepares the codebase for a new dimensions model
- Care should be taken to ensure no dimensions references remain in the codebase after completion
