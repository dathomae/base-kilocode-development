# Story209: Fix Post-Rebase Issues for Conditional Concept Pipeline

## Goal

Fix all issues discovered during review of branch `story121_and_122` after rebasing from `main`. The rebase restructured the examples directory from `examples/without_errors/` → `examples/parse_and_analyze_only/without_errors/` (and similarly for `with_errors/`), which broke the test file referencing example `.cpl` files. Additional documentation and process hygiene issues were also identified.

## References

- [Fix Post-Rebase Issues Plan](../plans/fix_post_rebase_issues_plan.md)
- [Code Review of branch story121_and_122](../context.md)
- [Story210: CPL Concepts Conditional Examples](Story210_cpl-concepts-conditional-examples.md)
- [Story208: Fix Conditional Spec Pipeline Defects](Story208_fix-conditional-spec-pipeline-defects.md)
- [Augment Examples Plan](../plans/augment_examples_plan.md)
- [Concept Analysis Test](../src/chronocone_planning_language/src/main/go/analysis/concept_analysis_test.go)
- [Example README](../src/chronocone_planning_language/examples/README.md)

## Dependencies

- [Story121_cpl-concepts-parser-conditional.md](Story121_cpl-concepts-parser-conditional.md) - Done
- [Story122_cpl-concepts-analyzer-conditional.md](Story122_cpl-concepts-analyzer-conditional.md) - Done
- [Story208_fix-conditional-spec-pipeline-defects.md](Story208_fix-conditional-spec-pipeline-defects.md) - Done
- Branch `story121_and_122` must be rebased on `main`

## Dependent Stories

None

## Tasks

1. Fix broken test file paths - Completed
   a. Update all 6 `filePath` values - Completed
   b. Run `go test ./analysis/... -run TestConditionalSpec_ExampleFiles` - Completed
   c. Run `go test ./...` to verify full suite passes with zero failures - Completed (12 packages pass)

2. Fix Story208 file extension - Completed
   a. Rename `memory-bank/stories/Story208_fix-conditional-spec-pipeline-defects` → `memory-bank/stories/Story208_fix-conditional-spec-pipeline-defects.md` - Completed
   b. Verify link in `toc.md` resolves correctly - Completed

3. Fix Story204 number collision - Completed
   a. Identify next available story number (210) - Completed
   b. Rename `memory-bank/stories/Story204_cpl-concepts-conditional-examples.md` to `memory-bank/stories/Story210_cpl-concepts-conditional-examples.md` - Completed
   c. Update all references to Story204 in other documents to use Story210 - Completed
   d. Update `toc.md` with the new filename and maintain correct status - Completed

4. Remove deprecated `ProcessConceptConditionals` - Completed
   a. Verify no production code calls `ProcessConceptConditionals` - Completed (only call was in ExitAnalyzeConceptScope, documented as no-op)
   b. Evaluate whether `TestProcessConceptConditionals` tests add value or are redundant with `processConditionalSpecChild` tests - Completed (redundant with TestResolveConditionalValues)
   c. Decision: remove both function and tests if redundant - Completed (~275 lines removed)
   d. Run `go test ./analysis/...` after any changes to verify no compilation errors or regressions - Completed (all pass)

5. Create missing invalid example files - Completed
   a. Create 5 parser-error examples - Completed (all trigger parse errors, verified via TestInvalidCondFiles_TriggerParseErrors)
   b. Create 4 analyzer-error examples - Completed (all parse OK but fail analysis, verified via TestInvalidCondFiles_TriggerAnalyzerErrors)
   c. Each file follows header-comment convention - Completed
   d. Verify parser-error files trigger parse errors - Completed
   e. Verify analyzer-error files parse successfully but fail at analysis time - Completed

6. Update README with conditional concept entries - Completed
   a. Add a section for each of the 6 valid example files - Completed
   b. Add a section for each of the 9 invalid example files with expected error descriptions - Completed
   c. Update the directory structure section - Completed

7. Run full test suite regression - Completed
   a. Run `go test ./... -count=1` - Completed
   b. Verify all 12 test packages pass (2 have no test files) - Completed
   c. Verify `TestConditionalSpec_ExampleFiles` specifically passes - Completed

## Test-First Development

This story follows the Logical TDD Lifecycle as defined in `memory-bank/concepts.md`.

- **Task 1** is driven by the existing failing test `TestConditionalSpec_ExampleFiles`. The test already exists and fails because the rebase changed example directory paths. The fix is to update the paths, then confirm the test passes.
- **Task 4** is a code-removal task. Before removing `ProcessConceptConditionals`, run the existing test `go test ./analysis/...` to establish the current passing baseline. After removal, rerun the same test suite to confirm zero regressions. If the removal breaks compilation or tests, the change must be reverted or adjusted.
- **Task 5** creates test fixtures. Each invalid example file is itself a test input. After creating each file, verify it produces the expected error category (parse vs. analyze) using the parser and analyzer. The expected error behavior is the "test" that validates the fixture is correct.

## Constraints

- Must not break existing functionality (Increments 1-7: basic concepts, hierarchy, range, before/after, conditional)
- Changes should be minimal and focused
- All example files must follow existing conventions (header comments, tab indentation, naming)
- The branch must be ready to merge to `main` after all tasks are complete

## Intent

The intent of this story is to restore the conditional concept example/test pipeline to a clean, mergeable state after the rebase introduced path changes and documentation drift. The work is purely cleanup and hygiene — no new features are introduced.

## Acceptance Criteria

1. `TestConditionalSpec_ExampleFiles` passes — all 6 sub-tests find and verify their example files
2. Full `go test ./...` passes with zero regressions
3. `Story208_fix-conditional-spec-pipeline-defects.md` has correct `.md` extension
4. Story204 is renamed to avoid number collision and correctly referenced in `toc.md`
5. `ProcessConceptConditionals` is either removed cleanly with no compilation errors, or an explicit decision to keep it is documented in a comment
6. 9 invalid example `.cpl` files exist in `examples/parse_and_analyze_only/with_errors/` and each triggers the correct error category
7. README updated with entries for all 15 new example files (6 valid + 9 invalid)
8. Branch is ready to merge to `main`

## Open Questions

None

## Known Issues

- `Story208_fix-conditional-spec-pipeline-defects` is currently missing its `.md` extension (Step 2)
- `Story210_cpl-concepts-conditional-examples.md` uses a story number that collides with a finished story (Step 3)

## Additional Context

These issues were discovered during code review of the `story121_and_122` branch after rebasing onto `main`. The rebase itself was clean, but the directory restructure on `main` broke relative paths in tests. Other issues (missing extension, number collision, missing examples) were pre-existing but only became visible after the rebase forced a full review.
