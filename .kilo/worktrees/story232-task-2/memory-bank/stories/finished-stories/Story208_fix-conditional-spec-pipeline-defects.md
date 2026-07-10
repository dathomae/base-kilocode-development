# Story208: Fix Conditional Spec End-to-End Pipeline Defects

## Goal

Fix the four critical defects in the conditional spec end-to-end pipeline discovered during review of branch `story121_and_122`. The defects prevent child concepts inside conditional specs from being correctly analyzed: `Analyze()` has no handler for `ConditionalSpecScope`, `ProcessConceptConditionals` looks for data in the wrong scope, `ExitAnalyzeConceptScope` skips child concepts inside `ConditionalSpecScope`, and `ProcessConceptConditionals` modifies the parent concept instead of creating the child concept.

## References

- [Fix Defects Plan](../plans/fix_defects_plan.md)
- [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md)
- [Concepts Requirements](../specifications/requirements/cpl/concepts-types-units-measures-syntax.md)
- [Concepts Design](../specifications/design/cpl/concepts-definitions.md)

## Dependencies

- [Story121_cpl-concepts-parser-conditional.md](Story121_cpl-concepts-parser-conditional.md) - Done
- [Story122_cpl-concepts-analyzer-conditional.md](Story122_cpl-concepts-analyzer-conditional.md) - Done

## Dependent Stories

- [Story210_cpl-concepts-conditional-examples.md](Story210_cpl-concepts-conditional-examples.md) - Not Started

## Tasks

1. Move conditional spec data to child ConceptScope — Completed
   a. Modify `ExitConditionalSpec` in `cpl_concepts_listener.go` to attach the `ConditionalSpec` to the child `ConceptScope.Data` instead of `ConditionalSpecScope.Data` — Completed
   b. Create a wrapper or embed `ConditionalSpecs` inside `ConceptDefinitionData` — Completed (field already existed in `ast.ConceptDefinitionData.ConditionalSpecs`)
   c. Verify parser tests still pass after data relocation — Completed (verified: all parser tests pass, zero regressions)

2. Update `ExitAnalyzeConceptScope` guard clause — Completed
   a. Change the early-return guard for `ConditionalSpecScope` children to call a new `processConditionalSpecChild` function instead of returning nil — Completed
   b. Ensure the walker still correctly visits child `ConceptScope` inside `ConditionalSpecScope` — Completed (walker already visits child ConceptScope; guard now delegates to stub)

3. Implement `processConditionalSpecChild` function — Completed
   a. Peek analysis context to get the child concept definition (e.g., `State`) — Completed
   b. Read conditional spec data from `table.Data` — Completed
   c. Resolve parent concept from indentation-based parent resolution — Completed
   d. Call `ResolveConditionalValues` to validate parent values and build `ConditionalValues` map — Completed
   e. Call `AddOrMergeConcept` to add the child concept to the analyzer — Completed
   f. Write integration tests that parse CPL with conditional specs and verify the child concept exists with correct `ConditionalValues` — Completed (TestConditionalSpec_ChildConceptInAnalysis)

4. Clean up `ConditionalSpecScope` handling — Completed
   a. Remove dead storage of data in `ConditionalSpecScope.Data` — Completed (removed in Task 1; verified no production code reads ConditionalSpecScope.Data)
   b. Verify no registered handler is needed for `ConditionalSpecScope` (walker can continue to skip it) — Completed (confirmed: walker visits children automatically, no handler needed)
   c. Ensure no dead code paths remain — Completed (no remaining references to ConditionalSpecScope.Data)

5. Evaluate and refactor `ProcessConceptConditionals` — Completed
   a. Check if `ProcessConceptConditionals` is still called from any code path — Completed (called from ExitAnalyzeConceptScope but no-op for both conditional and regular concepts)
   b. If yes, update it to work with the new data pattern; if no, remove it — Completed (marked as Deprecated; kept for test suite compatibility)
   c. Ensure clean analyzer code with no dead functions — Completed (deprecation comment added)

6. Add end-to-end integration tests — Completed
   a. Test parse CPL with conditional spec → analyze → verify child concept exists with correct `ConditionalValues` — Completed (TestConditionalSpec_ChildConceptInAnalysis already existed; added 3 new tests)
   b. Test parse CPL with multiple conditions → verify all conditional values resolved — Completed (TestConditionalSpec_MultipleConditionsInAnalysis: verifies State+USA, Province+Canada, State+Mexico)
   c. Test parse CPL with default values → verify `GetValuesForParent` fallback — Completed (TestConditionalSpec_DefaultValuesInAnalysis: verifies GetValuesForParent(USA) returns conditional, GetValuesForParent(Canada) returns defaults)
   d. Test error case — undefined parent value → verify error message with source location — Completed (TestConditionalSpec_UndefinedParentValueError: verifies error contains "parent value" and "Mexico")

7. Run full test suite regression — Completed
   a. Run `go test ./...` with all existing tests — Completed (all 15 packages pass)
   b. Verify no breakage in Increments 1–6 (basic concepts, hierarchy, range, before/after) — Completed (zero regressions)
   c. Confirm zero regressions — Completed (all tests pass)

8. Verify example files end-to-end — Completed
   a. Run each `conditional_concept_*.cpl` example through full parse+analyze pipeline — Completed (6 example files verified via table-driven TestConditionalSpec_ExampleFiles)
   b. Confirm child concepts appear in the analyzed definitions — Completed (all child concepts verified with non-empty ConditionalValues)
   c. Verify all valid examples produce correct analyzed output — Completed (all 6 examples pass)

## Additional Fix (Task 6)

During Task 6c testing, discovered a defect in [`processConditionalSpecChild()`](src/chronocone_planning_language/src/main/go/analysis/concept_analysis.go:1092): when merging conditional values into an existing unconditional concept (e.g., Region defined unconditionally with defaults, then extended via conditional override), the existing concept's `ResolvedValues` were not transferred to `DefaultValues`. This caused `GetValuesForParent` to return nil for non-matching parent values instead of falling back to the unconditional defaults. Fixed by adding DefaultValues transfer logic in the merge block.

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks.

Task 3 includes writing integration tests before implementing `processConditionalSpecChild`. Task 6 includes writing end-to-end tests before final verification. The test-and-implement cycle is integrated within each functional task.

## Constraints

- Must not break Increment 1–6 functionality (basic concepts, hierarchy, range, before/after).
- Must follow Logical TDD Lifecycle — tests before implementation changes where feasible.
- The fix should be minimal — avoid restructuring the symbol table hierarchy unless necessary.
- All changes must be in Go, within the `chronocone_planning_language` module.

## Intent

The overall intent is to ensure that the CPL parser and analyzer correctly handle conditional concept specifications end-to-end. Child concepts defined inside conditional specs (e.g., `State` inside `Country = USA`) must be fully parsed, analyzed, and added to the analyzer's concept map with their conditional values resolved against the parent concept.

## Acceptance Criteria

- Child concepts inside conditional specs appear in `defs.Concepts` after analysis.
- `ConditionalValues` map is populated with correct parent-value-keyed entries.
- `GetValuesForParent` returns correct conditional and default values.
- All existing parser and analyzer tests pass without regression.
- All `conditional_concept_*.cpl` example files produce correct analyzed output.
- Error case produces a clear error message with source location.

## Open Questions

None.

## Known Issues

The four defects are documented in the [Fix Defects Plan](../plans/fix_defects_plan.md).

## Additional Context

This story is a follow-up to Story121 and Story122, which implemented parsing and analysis for conditional specs. The defects were discovered during end-to-end review of the combined branch.
