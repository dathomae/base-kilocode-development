# Story274: Fix Conversion Issues & Unify Unit/Measure Implementations

## Goal

Address the five issues identified in the [Story 230](finished-stories/Story230_cpl-conversions.md) review, then unify the duplicated unit-conversion and measure-conversion implementations to eliminate redundant code that handles the same problem in different ways.

## References

- [Plan: Fix Conversion Issues](memory-bank/plans/fix-conversion-issues-plan.md)
- [Story 230: CPL Conversions](finished-stories/Story230_cpl-conversions.md)

## Dependencies

None — this story fixes issues in completed code and refactors independently.

## Dependent Stories

None.

## Tasks

### Test-First Development

This story follows the Logical TDD Lifecycle. For each coding task, tests are written or updated before the implementation is verified together. Within each task, the test-and-implement cycle is integrated — tests are not broken out as separate tasks.

### Task Sizing

**Small tasks** (routed to `small-code-for-story-implementor`):
- **Task 2**: Single-file switch case addition, existing pattern.
- **Task 4**: Single-file godoc update, no behavior change.
- **Task 5a**: Purely additive parameter to existing function, one new switch-case path.
- **Task 6a**: Purely additive new file with generic struct definition.
- **Task 7a**: Purely additive new file with shared generic function.
- **Task 11**: Verification only; no production code changes expected.

**Medium tasks** (routed to `code-for-story-implementor`):
- **Task 1**: Touches 5 files; modifies `invertExpression` AND adds inversion call in a separate file; 8 mechanical subtasks across coordinated logic changes. Architect judged it borderline — no natural seam for decomposition.
- **Task 3**: Introduces new struct in `cil/` AND modifies consumer in `analysis/`; cross-package change with new abstraction.
- **Task 5b**: Touches 2 production files; coordinated signature migration + dead code removal must compile together.
- **Task 6b**: Touches 5+ production files across 2 packages; coordinated type migration cascades field renames through all consumers.
- **Task 7b**: 2 production files; substantial refactoring of full implementations to thin wrappers + dead code removal across both files.
- **Task 9**: 2 production files across 2 packages; CIL type unification + synthesis function replacement.
- **Task 10**: Single-file refactoring within `synthesize.go`, replacing two separate conversion loops with a shared generic helper. Not purely additive — consolidates existing logic into a new parameterized abstraction.

### Parallel Execution

- **Group A** (Tasks 1, 3): Task 1 modifies `measures_conversion_analysis.go` + `units_conversion_analysis.go` + tests; Task 3 creates new files in `cil/` + modifies `expression_to_cil.go`. Fully disjoint files across analysis and cil packages.
- **Group B** (Tasks 2, 3): Task 2 adds default case to `measures_conversion_analysis.go`; Task 3 creates `cil/` files + modifies `expression_to_cil.go`. Fully disjoint files.
- **Group C** (Tasks 3, 4): Task 4 is godoc-only on `measures_conversion_analysis.go`; Task 3 creates `cil/` files + modifies `expression_to_cil.go`. Fully disjoint files.
- **Group D** (Tasks 5a, 6a): Task 5a modifies `build_expression_tree.go` + test; Task 6a creates new `conversion_pair.go` + test. No existing files shared.
- **Group E** (Tasks 7a, 9): Task 7a creates `conversion_analysis_core.go` in `analysis/`; Task 9 touches `cil/cil_conversions.go` + `synthesis/synthesize.go`. Fully disjoint across three packages.

**Phase 1 file contention note**: Tasks 1, 2, and 4 all touch `measures_conversion_analysis.go`. None can parallelize with each other. The maximum Phase 1 parallelism is 2 tasks: Task 3 paired with exactly one of {1, 2, 4}. The other two must run sequentially afterward (or in subsequent phases alongside 5a/6a).

### Execution Order

```
[Task 1 ‖ Task 3] → Task 2 → Task 4 → [Task 5a ‖ Task 6a] → [Task 5b → Task 6b] → [Task 7a ‖ Task 9] → Task 7b → Task 10 → Task 11
```

Alternative scheduling (sequentially processing Phase 1 tasks that share files):

```
[Task 1 ‖ Task 3] → Task 2 → Task 4
        ‖
   Task 5a ‖ Task 6a
        ‖
   Task 5b
        ‖
   Task 6b
        ‖
   [Task 7a ‖ Task 9]
        ‖
   Task 7b → Task 10 → Task 11
```

Tasks 5a/6a can start anytime (file-disjoint from all Phase 1 tasks). Tasks 5b/6b depend on 5a/6a respectively and share files with Phase 1 tasks — their worktrees must branch from a commit that includes all Phase 1 changes.

### Reintegration

All reintegration instructions apply after worktree branches are merged back into the story branch. The story-implementor determines the branch name at runtime.

**Group A (Tasks 1, 3):**
- Merge order: any order (fully disjoint files across analysis and cil packages)
- Integration tests: `go test ./src/main/go/analysis/... ./src/main/go/cil/... -count=1 -timeout 60s` — verify combined conversion analysis and CIL node behavior
- Watch for conflicts in: `src/main/go/analysis/measures_conversion_analysis.go` (Task 1 modifies it; not touched by Task 3 — but future concurrent worktrees touching this file must be aware)

**Group B (Tasks 2, 3):**
- Merge order: any order (fully disjoint files)
- Integration tests: `go test ./src/main/go/analysis/... ./src/main/go/cil/... -count=1 -timeout 60s` — verify error coverage and CIL translation together
- Watch for conflicts in: none expected

**Group C (Tasks 3, 4):**
- Merge order: any order (godoc-only vs new package files)
- Integration tests: `go test ./src/main/go/analysis/... ./src/main/go/cil/... -count=1 -timeout 60s` — minimal risk; verify compilation and test pass
- Watch for conflicts in: none expected

**Group D (Tasks 5a, 6a):**
- Merge order: any order (fully disjoint: `build_expression_tree.go` vs new `conversion_pair.go`)
- Integration tests: `go test ./src/main/go/analysis/... -count=1 -timeout 60s` — both add new code to analysis package; combined compilation and test pass must be verified
- Watch for conflicts in: `src/main/go/analysis/units_conversion_analysis.go` (not modified by either task, but Task 5a changes `BuildExpressionTree` signature — the existing caller must still compile with `nil` passed for `vars` before Task 5b updates it)

**Group E (Tasks 7a, 9):**
- Merge order: any order (fully disjoint across three packages)
- Integration tests: `go test ./... -count=1 -timeout 120s` — both touch the pipeline boundary between analysis→cil→synthesis; full test suite needed
- Watch for conflicts in: `src/main/go/synthesis/synthesize.go` (Task 9 modifies conversion synthesis code; not touched by 7a but build-order matters)

### Task 1: Add algebraic reversal to unit conversions — Completed

**Size**: Medium. Touches 5 files; modifies `invertExpression` AND adds inversion call in a separate file. Architect judged borderline — no natural seam for decomposition into smaller tasks.

**Goal**: Make unit conversions produce both forward and derived inverse pairs, matching the behavior already present in measure conversions.

**Subtasks**:
a. Update test expectations to reflect the desired behavior (test-first):
   - `TestLengthUnitConversions` in `src/main/go/analysis/units_conversion_analysis_test.go`: change `Meter` expectation from 1 conversion to 2, `Centimeter` from 1 to 2
   - `TestTemperatureUnitConversions`: each unit now gets 2 conversions (forward + derived inverse)
   - `TestSynthesis_UnitConversions_EmitsCompilableCode` in `src/main/go/synthesis/synthesize_test.go`: write a single direction, expect the inverse to exist
   - `TestFullPipeline_LengthConversions` and `TestFullPipeline_TemperatureConversions` in `src/main/go/analysis/cpl_conversions_integration_test.go`
   - `TestUnitConversions_NoRegression_ExistingTests` in `units_conversion_analysis_test.go`: expect 2 conversions for `Kilometer` instead of 1 — Completed
b. In `invertExpression` (`src/main/go/analysis/measures_conversion_analysis.go`), add a case for `*UnitReference` that swaps `oldVar`/`newVar` when `Name` matches (same logic as the existing `*VariableRef` case) — Completed
c. In `AnalyzeUnitConversions` (`src/main/go/analysis/units_conversion_analysis.go`), after building the forward `AnalysisConversionPair`, call `invertExpression(calcExpr, sourceVar, targetVar)` to derive the inverse, then append both forward and inverse pairs to `unitDef.Conversions` — Completed
d. Run `go test ./src/main/go/analysis/... ./src/main/go/synthesis/... -count=1 -timeout 120s` to verify all tests pass — Completed

**Acceptance Criteria**:
- `invertExpression` handles `*UnitReference` nodes identically to `*VariableRef` nodes
- `AnalyzeUnitConversions` produces both forward and inverse pairs
- All updated tests pass with new expectations
- No regressions in existing test suites

### Task 2: Add default error case to validateUnitAnnotationsWithContext — Completed [Small]

**Goal**: Prevent silent fall-through for unrecognized expression types during unit annotation validation.

**Subtasks**:
a. Add a test case in `src/main/go/analysis/measures_conversion_analysis_test.go` that triggers the default case (pass an unknown expression type that implements the `Expression` interface but has no case in the switch) — Not Started
b. In `validateUnitAnnotationsWithContext` (`src/main/go/analysis/measures_conversion_analysis.go`), add a `default` branch to the type switch that returns `fmt.Errorf("unknown expression type %T in unit annotation validation", expr)` — Not Started
c. Run `go test ./src/main/go/analysis/... -count=1 -timeout 60s` — Not Started

**Acceptance Criteria**:
- `validateUnitAnnotationsWithContext` returns an error for unrecognized expression types
- Test case verifies the error is produced for an unknown type
- All existing tests continue to pass

### Task 3: Introduce dedicated CIL node for VariableRef — Completed

**Size**: Medium. Introduces new struct in `cil/` AND modifies consumer in `analysis/`; cross-package change with new abstraction.

**Goal**: Replace the incorrect mapping of `VariableRef` to `CILUnitRef` with a dedicated `CILVariableRef` node type, ensuring pattern-bound variables are distinct from unit references in the CIL tree.

**Subtasks**:
a. Create `src/main/go/cil/cil_variable_ref.go` defining `type CILVariableRef struct { Name string }` with `cilNode()` marker method — Completed
b. Create `src/main/go/cil/cil_variable_ref_test.go` with a compile-time assertion `var _ CILNode = (*CILVariableRef)(nil)` and a construction test; run `go test ./src/main/go/cil/... -count=1` to verify the type compiles and tests pass — Completed
c. In `ExpressionToCILNode` (`src/main/go/analysis/expression_to_cil.go`), change the `*VariableRef` case to emit `&cil.CILVariableRef{Name: e.Name}` instead of `&cil.CILUnitRef{UnitGroup: e.Name}` — Completed
d. Update existing tests that expect `CILUnitRef` output from `VariableRef` conversion in `src/main/go/analysis/expression_to_cil_test.go` — Completed
e. Verify `SynthesizeExpression` handles `CILVariableRef` — read the walker in the synthesis package and add a case if needed (it should emit `Name` as a bare identifier, same behavior as `CILUnitRef` today) — Completed
f. Run `go test ./src/main/go/cil/... ./src/main/go/analysis/... ./src/main/go/synthesis/... -count=1 -timeout 120s` — Completed

**Acceptance Criteria**:
- `CILVariableRef` is a distinct CIL node type with `Name` field
- `ExpressionToCILNode` maps `*VariableRef` to `*CILVariableRef`, not `*CILUnitRef`
- `SynthesizeExpression` correctly emits `CILVariableRef` as a bare identifier
- All tests pass; no change in generated code behavior

### Task 4: Document invertExpression limitations — Completed [Small]

**Goal**: Make the supported and unsupported expression patterns for algebraic inversion explicit in the godoc.

**Subtasks**:
a. Update the godoc on `invertExpression` in `src/main/go/analysis/measures_conversion_analysis.go` to document:
   - Supported patterns: `source * N → target / N`, `N * source → target / N`, `source / N → target * N`, `source + N → target - N`, `N + source → target - N`, `source - N → target + N`, `N - source → N - target`
   - Unsupported: multiplication/addition with non-constant operands, division with non-constant divisor, power expressions (`^`) — Not Started

**Acceptance Criteria**:
- Godoc clearly describes supported and unsupported patterns
- No code behavior change; no test modifications needed

### Task 5a: Add optional vars parameter to BuildExpressionTree — Completed [Small]

**Goal**: Extend `BuildExpressionTree` with an optional `vars` parameter so it can distinguish pattern-bound variables from unit references, eliminating the need for the duplicate `BuildConversionExpressionTree`.

**Subtasks**:
a. Add tests in `src/main/go/analysis/build_expression_tree_test.go` for: `vars=nil` preserves existing behavior, `vars` lookup creates `VariableRef`, undefined identifier returns error (tests written first to define the desired behavior) — Not Started
b. Add optional `vars map[string]bool` parameter to `BuildExpressionTree` (`src/main/go/analysis/build_expression_tree.go`). When `vars` is nil, behavior is unchanged. When a `UnitReferenceDefinition` is encountered: if the name is in `vars`, push a `VariableRef`; otherwise look up in `units` as before. If the identifier is not in `vars` and not in `units`, return an error — Not Started
c. Run `go test ./src/main/go/analysis/... -run TestBuildExpressionTree -count=1` — Not Started

**Acceptance Criteria**:
- `BuildExpressionTree` accepts optional `vars` parameter
- Nil `vars` preserves all existing behavior
- `VariableRef` created for pattern-bound variables
- Undefined identifiers produce clear errors
- All existing `BuildExpressionTree` callers continue to work (pass `nil`)

### Task 5b: Migrate consumers to unified BuildExpressionTree — Completed

**Size**: Medium. Touches 2 production files; coordinated signature migration + dead code removal must compile together.

**Goal**: Remove `BuildConversionExpressionTree` and `validateCalculationExpression`, migrating all callers to the unified `BuildExpressionTree` with `vars` parameter.

**Subtasks**:
a. Remove `BuildConversionExpressionTree` from `src/main/go/analysis/measures_conversion_analysis.go` (approximately 70 lines) — Not Started
b. In `AnalyzeUnitConversions` (`src/main/go/analysis/units_conversion_analysis.go`), pass `boundVars` to `BuildExpressionTree` instead of calling without vars — Not Started
c. In `AnalyzeMeasureConversions` (`src/main/go/analysis/measures_conversion_analysis.go`), pass `patternVars` to `BuildExpressionTree` instead of calling `BuildConversionExpressionTree` — Not Started
d. In `AnalyzeMeasureScope` (compound unit expressions), pass `nil` (no pattern-bound vars in measure definitions) — Not Started
e. Remove `validateCalculationExpression` from `units_conversion_analysis.go` (approximately 30 lines) — the unified `BuildExpressionTree` already validates identifiers, making this redundant — Not Started
f. Update tests in `units_conversion_analysis_test.go`, `measures_conversion_analysis_test.go`, and `build_expression_tree_test.go` for the new call signatures and removed functions — Not Started
g. Run `go test ./src/main/go/analysis/... -count=1 -timeout 120s` — Not Started

**Acceptance Criteria**:
- `BuildConversionExpressionTree` is removed
- `validateCalculationExpression` is removed
- All consumers pass `vars` (or nil) to the unified `BuildExpressionTree`
- All tests pass; no change in validation behavior

### Task 6a: Create generic ConversionPair[Def any] type — Completed [Small]

**Goal**: Define a generic conversion pair type parameterized by the definition type, replacing the two duplicate concrete types.

**Subtasks**:
a. Create `src/main/go/analysis/conversion_pair.go` defining `type ConversionPair[Def any] struct { SourceDef, TargetDef *Def; TargetVar, SourceVar string; Calculation Expression }` — Not Started
b. Write tests in `src/main/go/analysis/conversion_pair_test.go` verifying construction with `AnalysisUnitDefinition` and `AnalysisMeasureDefinition` type parameters, field access, etc. — Not Started
c. Run `go test ./src/main/go/analysis/... -run ConversionPair -count=1` — Not Started

**Acceptance Criteria**:
- `ConversionPair[Def any]` is defined and compiles
- Tests pass with both type parameters
- No existing code modified (purely additive)

### Task 6b: Migrate all conversion pair types to generic ConversionPair — Completed

**Size**: Medium. Touches 5+ production files across 2 packages; coordinated type migration cascades field renames through all consumers.

**Goal**: Remove the duplicate `AnalysisConversionPair` and `AnalysisMeasureConversionPair` types, migrating all references to the generic `ConversionPair[Def any]`.

**Subtasks**:
a. Remove `AnalysisConversionPair` from `src/main/go/analysis/units_conversion_analysis.go`; update `AnalyzeUnitConversions` to construct `ConversionPair[AnalysisUnitDefinition]` with `SourceDef`/`TargetDef` fields — Not Started
b. Remove `AnalysisMeasureConversionPair` from `src/main/go/analysis/measures_conversion_analysis.go`; update `AnalyzeMeasureConversions` to construct `ConversionPair[AnalysisMeasureDefinition]` with `SourceDef`/`TargetDef` fields — Not Started
c. Update `AnalysisUnitDefinition.Conversions` field in `src/main/go/analysis/units_analysis.go` to `[]ConversionPair[AnalysisUnitDefinition]` — Not Started
d. Update `AnalysisMeasureDefinition.Conversions` field in `src/main/go/analysis/measures_analysis.go` to `[]ConversionPair[AnalysisMeasureDefinition]` — Not Started
e. Update `BuildTransformerFromAnalysis` in `src/main/go/synthesis/synthesize.go`: change field access from `.SourceUnit`/`.TargetUnit` to `.SourceDef.Name`/`.TargetDef.Name`, and from `.SourceMeasure`/`.TargetMeasure` to `.SourceDef.Name`/`.TargetDef.Name` — Not Started
f. Update all test files that construct or reference these types: `units_conversion_analysis_test.go`, `measures_conversion_analysis_test.go`, `cpl_conversions_integration_test.go`, `synthesize_test.go` — Not Started
g. Run `go test ./src/main/go/... -count=1 -timeout 120s && go vet ./src/main/go/...` — Not Started

**Acceptance Criteria**:
- `AnalysisConversionPair` and `AnalysisMeasureConversionPair` are removed
- All code uses `ConversionPair[AnalysisUnitDefinition]` and `ConversionPair[AnalysisMeasureDefinition]`
- All tests pass and `go vet` is clean

### Task 7a: Extract shared analyzeConversions[Def any] core — Completed [Small]

**Goal**: Extract a single generic function containing the common conversion analysis logic shared between unit and measure conversions.

**Subtasks**:
a. Create `src/main/go/analysis/conversion_analysis_core.go` defining `analyzeConversions[Def any]` with the following logic:
   - Walk children of the scope table for the given `targetScopeType`
   - Collect PatternScopes and CalculationScopes
   - Validate count and ordering
   - Extract pattern pairs via `extractPatternPairs`
   - Resolve source and target from the lookup map
   - Validate variable names (check against the units map)
   - Build expression via unified `BuildExpressionTree` with `vars`
   - Validate no bare numbers via `validateNoBareNumbers`
   - Assemble forward `ConversionPair[Def]`
   - Derive inverse via `invertExpression`
   - Append both to `def.Conversions`
   - Return error on failure — Not Started
b. Write tests in `src/main/go/analysis/conversion_analysis_core_test.go` verifying behavior for both `AnalysisUnitDefinition` and `AnalysisMeasureDefinition` type parameters — Not Started
c. Run `go test ./src/main/go/analysis/... -count=1 -timeout 120s` — Not Started

**Acceptance Criteria**:
- `analyzeConversions[Def any]` is a standalone generic function implementing unified conversion logic
- Tests pass with both type parameters
- No existing code modified (purely additive)

### Task 7b: Convert existing analyzers to thin wrappers, remove dead code — Completed

**Size**: Medium. 2 production files; substantial refactoring of full implementations to thin wrappers + dead code removal across both files.

**Goal**: Replace the duplicated logic in `AnalyzeUnitConversions` and `AnalyzeMeasureConversions` with thin wrappers calling the shared `analyzeConversions` core. Remove all dead code specific to the old duplicate implementations.

**Subtasks**:
a. Convert `AnalyzeUnitConversions` (`src/main/go/analysis/units_conversion_analysis.go`) to a thin wrapper calling `analyzeConversions` with: table, ConversionScope, units map, defName, unitDef pointer, "unit" entity string — Not Started
b. Convert `AnalyzeMeasureConversions` (`src/main/go/analysis/measures_conversion_analysis.go`) to a thin wrapper calling `analyzeConversions` with: measureScope, MeasureConversionScope, measures map (adjusted for the `CPLAnalyzer` struct), defName, measureDef pointer, "measure" entity string — Not Started
c. Remove `extractBoundMeasureRef` from `measures_conversion_analysis.go` (approximately 25 lines) — Not Started
d. Remove `validateUnitAnnotations` and `validateUnitAnnotationsWithContext` from `measures_conversion_analysis.go` (approximately 40 lines) — they are replaced by the stricter `validateNoBareNumbers` + unified `BuildExpressionTree` identifier validation — Not Started
e. Remove any remaining duplicated logic from both wrappers (pattern extraction loops, validation steps, pair assembly that are now in the shared core) — Not Started
f. Run `go test ./src/main/go/analysis/... -count=1 -timeout 120s` — Not Started

**Acceptance Criteria**:
- `AnalyzeUnitConversions` and `AnalyzeMeasureConversions` are thin wrappers calling `analyzeConversions`
- `extractBoundMeasureRef`, `validateUnitAnnotations`, and `validateUnitAnnotationsWithContext` are removed
- All conversion analysis tests pass with identical behavior

### Task 9: Merge CIL conversion types — Completed

**Size**: Medium. 2 production files across 2 packages; CIL type unification + synthesis function replacement.

**Goal**: Merge `CILConversionPair` and `CILMeasureConversionPair` into a single type, and replace the two separate synthesis functions with one parameterized function.

**Subtasks**:
a. In `src/main/go/cil/cil_conversions.go`, merge `CILConversionPair` and `CILMeasureConversionPair` into a single `CILConversionPair` — change `SourceUnit`/`TargetUnit` and `SourceMeasure`/`TargetMeasure` fields to `SourceName`/`TargetName` — Not Started
b. Replace `SynthesizeUnitConversions` and `SynthesizeMeasureConversions` with a single `SynthesizeConversions(pairs []CILConversionPair, mapName, funcName, entityName string) string` — same logic, parameterized by the map name (`"unitConversions"` / `"measureConversions"`), function name (`"convertUnit"` / `"convertMeasure"`), and entity name for panic messages — Not Started
c. In `src/main/go/synthesis/synthesize.go`, update `BuildTransformerFromAnalysis` to use the unified `CILConversionPair` type and call `SynthesizeConversions` instead of `SynthesizeUnitConversions` / `SynthesizeMeasureConversions` — Not Started
d. Update tests in `cil_conversions_test.go` (if exists) and `synthesize_test.go` for the unified types and function — Not Started
e. Run `go test ./src/main/go/cil/... ./src/main/go/synthesis/... -count=1 -timeout 120s` — Not Started

**Acceptance Criteria**:
- Single `CILConversionPair` type with `SourceName`/`TargetName` fields
- Single `SynthesizeConversions` function parameterized by map/func/entity names
- Generated Go code is identical in structure for both unit and measure cases
- All tests pass

### Task 10: Unify synthesis pipeline — Completed

**Size**: Medium. Single-file refactoring — consolidates two separate conversion-synthesis loops into a shared generic helper.

**Goal**: Replace the two separate conversion-synthesis loops in `BuildTransformerFromAnalysis` with a single shared helper function.

**Subtasks**:
a. In `src/main/go/synthesis/synthesize.go`, extract a shared helper `synthesizeConversionsFromDefs[Def any]` that iterates over a map of definitions, extracts `Conversions`, converts `Expression` to CIL via `ExpressionToCILNode`, synthesizes Go code via `SynthesizeExpression`, builds `CILConversionPair` entries, and calls `SynthesizeConversions` — Not Started
b. Replace the two separate loops (unit conversions at ~line 155, measure conversions at ~line 180) with calls to the shared helper instantiated with `AnalysisUnitDefinition` → `"unit"` and `AnalysisMeasureDefinition` → `"measure"` — Not Started
c. Keep two fields in `SynthesisInput` (`UnitConversionCode` and `MeasureConversionCode`) but populate both from the shared helper — Not Started
d. Run `go test ./src/main/go/synthesis/... -count=1 -timeout 120s` — Not Started

**Acceptance Criteria**:
- `synthesizeConversionsFromDefs[Def any]` is a single shared helper function
- `BuildTransformerFromAnalysis` uses the helper for both conversion types
- Generated code is identical to pre-refactor output
- All synthesis tests pass

### Task 11: Final verification and cleanup — Completed [Small]

**Goal**: Verify the entire codebase is clean after all refactoring, remove any remaining dead code, and confirm no regressions.

**Subtasks**:
a. Verify dead-code removal is complete: confirm `BuildConversionExpressionTree`, `extractBoundMeasureRef`, `validateUnitAnnotations`, `validateUnitAnnotationsWithContext`, `AnalysisConversionPair`, and `AnalysisMeasureConversionPair` are all removed — Not Started
b. Run `go vet ./src/main/go/...` — confirm clean (except ANTLR-generated warnings) — Not Started
c. Run full test suite: `go test ./src/main/go/... -count=1 -timeout 120s` — Not Started
d. Verify that the generated Go code from `SynthesizeConversions` is identical in structure to the current separate implementations for both unit and measure cases — Not Started

**Acceptance Criteria**:
- No dead code remains from the old duplicate implementations
- `go vet` is clean
- Full test suite passes with zero regressions
- Generated code structure is preserved

## Notes

- **Go generics**: The project targets Go 1.21+, so generics are available (since Go 1.18). The `go.mod` file should already specify a compatible version.
- **Deduplication**: When forward and inverse pairs coexist (e.g., from explicit CPL declarations), duplicates in the emitted Go map are harmless — no deduplication logic is needed.
- **Unified validation strategy**: The stricter approach (reject bare numbers entirely, validate identifiers during tree building) is used everywhere. Measure conversion calculations like `calculation m * 1000` are valid (`m` is a unit reference). Bare-number calculations like `calculation 32` or `calculation 5 + 9` are rejected.
- **No self-conversion check**: Already present in measure analysis; retained in the shared core.
