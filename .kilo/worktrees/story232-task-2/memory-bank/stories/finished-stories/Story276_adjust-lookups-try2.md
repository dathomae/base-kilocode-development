# Story276: Fix Lookups End-to-End Pipeline and Example Data

## Goal

Fix the blocking issues that prevent the `lookups_basic` example from compiling and executing end-to-end. Two critical gaps from Story275 are addressed: (1) the CIL postfix walker does not handle `CILLookupNode`, causing synthesis to fail for any metric referencing a lookup; and (2) binary tensor operations lack dimension broadcasting, causing runtime evaluation to fail when a lookup tensor (fewer dimensions) is combined with a connector tensor (more dimensions). Additionally, several example data errors introduced during Story275 are corrected.

When this story is complete, `go test ./src/main/go/... -count=1` passes with zero failures, and the `lookups_basic` example produces correct results.

## References

- [Plan: Fix Story275 — Lookups End-to-End Pipeline and Example Data](plans/adjust-lookups-try2-plan.md)
- [Story275: IOSpec Connectors and Lookups](finished-stories/Story275_iospec-connectors-and-lookups.md)
- [Story242: Lookups Increment 1 Analyze](finished-stories/Story242_lookups-increment1-analyze.md)
- [Story243: Lookups Increment 1 Synthesize](finished-stories/Story243_lookups-increment1-synthesize.md)
- [Design Principle: The IOSpec is a Spec, Not Data Storage](specifications/design/design.md)

## Dependencies

- [Story275: IOSpec Connectors and Lookups](finished-stories/Story275_iospec-connectors-and-lookups.md) — Done
  : Must be complete because this story works atop the connectors/lookups structural refactoring from Story275.

## Dependent Stories

None — this is a bug-fix story that unblocks other stories incidentally.

## Tasks

### Task Sizing

Tasks 1, 2a, 3, 4, 6, and 7 are annotated `[Small]` — each touches at most 2 files, addresses a single concern following an existing pattern, and requires no new abstraction creation. Task 2b is not annotated `[Small]` — it introduces a new `TensorDimensionNamesAreBroadcastable` helper function, requires coordinated changes across both the arithmetic and comparison operator dispatch branches in the evaluator, and needs integration tests with controller setup, tensor dimension interactions, and mask-based comparison verification.

Task 5 is a single CLI invocation (no code changes). Tasks 8 and 9 are verification gates.

The architect-for-story-planning confirmed these classifications. Task 2a was a borderline call — it adds a single helper function matching the small-task pattern, but the algorithm involves index-mapping logic with some complexity. It remains `[Small]` per the explicit "single helper function + test" carve-out.

### Parallel Execution

- **Group A** (Tasks 1, 2a, 3, 4, 6, 7): All six tasks touch completely disjoint file sets across different packages (cil, calculation, connector) and example-data directories. No compile-time or semantic dependencies exist among them. Task 2a is a pure tensor helper with no knowledge of Task 1's walker changes. Task 2b (the evaluator integration) is the only task that needs both.

### Execution Order

```
[Task 1 ‖ Task 2a ‖ Task 3 ‖ Task 4 ‖ Task 6 ‖ Task 7] → Task 2b → Task 5 → Task 8 → Task 9
```

- Group A (Tasks 1, 2a, 3, 4, 6, 7) can all execute in parallel.
- Task 2b depends on Task 1 (needs `CILLookupNode` case to write the integration test with a lookup tensor in the controller) and Task 2a (needs `TensorBroadcastTo` function to call).
- Task 5 depends on Tasks 3 and 4 (needs corrected `iospec.json` and `raw_strength.json`).
- Tasks 8 and 9 are sequential final validation gates.

### Reintegration

**Group A (Tasks 1, 2a, 3, 4, 6, 7):**
- Merge order: any order (fully disjoint file sets, no shared infrastructure)
- Integration tests:
  - `go test ./src/main/go/cil/... -count=1 -timeout 60s` — Task 1 adds `CILLookupNode` case to postfix walker; verify walker tests pass
  - `go test ./src/main/go/calculation/... -count=1 -timeout 60s` — Task 2a adds `TensorBroadcastTo` helper; verify helper unit tests pass
  - `go test ./src/main/go/connector/... -count=1 -timeout 60s` — Task 6 fixes stale assertion string; verify connector loader tests pass
  - `go build ./src/main/go/...` — Verify all Go packages compile together after concurrent changes across cil, calculation, and connector packages
- Watch for conflicts in: `go.sum` (dependency resolution in separate worktrees may produce different hashes); `go.mod` (same reason)

---

### Task 1: Add CILLookupNode case to postfix walker — Completed [Small]

The CIL-to-postfix walker `walkToPostfix` in `src/main/go/cil/postfix_walk.go` dispatches on CIL node types via a type switch. When `CILLookupNode` was added in Story241, `postfix_walk.go` was never updated to handle it. Any metric expression referencing a lookup (e.g., `product_strength = raw_strength * strength_adjusted_for_material`) falls through to the `default` case and returns `"unsupported CIL node type: *cil.CILLookupNode"`.

This is a one-line fix, exactly analogous to the existing `*CILInputRef` case.

#### Intent

Make the postfix walker recognize `CILLookupNode` so lookup references are correctly translated to postfix `push` operations.

#### Acceptance Criteria

- `src/main/go/cil/postfix_walk.go` contains a `case *CILLookupNode` that appends `PostfixOp{Kind: "push", Name: n.LookupName}`.
- A new test function `TestWalkCILTreeToPostfix_LookupRef` in `src/main/go/cil/cil_transformer_postfix_test.go` verifies:
  - A `CILLookupNode` with `LookupName = "strength_factor"` produces a single `PostfixOp` with `Kind: "push"` and `Name: "strength_factor"`.
  - A lookup reference in a binary expression (e.g., `connector_ref * lookup_ref`) produces the correct postfix sequence with both operands pushed.
- `go test ./src/main/go/cil/... -count=1` passes.

#### Tasks

a. Write test function `TestWalkCILTreeToPostfix_LookupRef` (and, if needed for the binary-expr case, `TestWalkCILTreeToPostfix_InputTimesLookup`) following existing test patterns in `src/main/go/cil/cil_transformer_postfix_test.go`. Run tests to confirm they fail (missing CILLookupNode case) — Not Started
b. Add case for `*CILLookupNode` in `walkToPostfix` in `src/main/go/cil/postfix_walk.go` at line ~134 — Not Started
c. Run `go test ./src/main/go/cil/... -count=1` and verify all pass — Not Started

---

### Task 2a: Implement TensorBroadcastTo helper function — Completed [Small]

When the evaluator multiplies a connector tensor (dimensions `[Product, Material, Daily]`, 63 elements) by a lookup tensor (dimensions `[Material]`, 3 elements), `TensorMultiplyFloat64` rejects them with a dimension mismatch error. A `TensorBroadcastTo` helper function is needed that can repeat a small tensor's values across missing dimensions to match a target shape.

#### Intent

Create a self-contained broadcasting helper that can expand a tensor with fewer dimensions to match a target dimension set, following the existing `TensorWhereFloat64` broadcasting precedent.

#### Acceptance Criteria

- `src/main/go/calculation/chronocone_tensor.go` contains a `TensorBroadcastTo(src *ChronoconeTensor[any], targetDimNames []string, targetShape []int) (*ChronoconeTensor[any], error)` function, following the existing `*ChronoconeTensor[any]` convention used by all other tensor helpers in the package.
- The function validates that `src`'s dimension names are a subset of `targetDimNames`, with matching dimensions in the same order.
- For each position in the target tensor, the function maps back to the corresponding source position, repeating source values across non-matching dimensions.
- Unit tests in `src/main/go/calculation/chronocone_tensor_test.go` cover:
  - Source `[Material:3]` → target `[Product:3, Material:3, Daily:7]` produces 63 output cells with correct per-Material values repeated across all Products and Days.
  - Source is NOT a subset of target → error.
  - Source equals target → identity (no broadcasting).
  - Source dimensions in wrong order → error.
  - Source `[Material:3]` → target `[Product:3, Material:3]` (subset, adjacent) → correct.
- `go test ./src/main/go/calculation/... -count=1` passes.

#### Tasks

a. Write unit tests in `src/main/go/calculation/chronocone_tensor_test.go` for subset, identity, and error cases. Run tests to confirm they fail (function not yet defined) — Not Started
b. Add `TensorBroadcastTo` function to `src/main/go/calculation/chronocone_tensor.go` — Not Started
c. Run `go test ./src/main/go/calculation/... -count=1` and verify all pass — Not Started

---

### Task 2b: Update postfix evaluator to broadcast on dimension mismatch — Completed

After Task 1 (walker handles `CILLookupNode`) and Task 2a (`TensorBroadcastTo` exists), the postfix evaluator must be updated to detect dimension mismatches in binary operations and broadcast the smaller tensor before applying the operation.

#### Intent

Wire broadcasting into the evaluator so that binary tensor operations succeed when operands have different dimensions but one is a compatible subset of the other. This applies to both arithmetic and comparison operators.

#### Acceptance Criteria

- A `TensorDimensionNamesAreBroadcastable` helper function is added to `src/main/go/calculation/chronocone_tensor.go` that checks whether one tensor's dimension names are a compatible subset of another's (same names in same order for matching dimensions). This check is used as a **precondition** before applying the `Tensor*Float64` operation — do NOT rely on catching a dimension-mismatch error string, which is fragile.
- In `src/main/go/calculation/postfix_op.go`, for both the arithmetic operator branch (the `default` case at line 270 for `Add`/`Subtract`/`Multiply`/`Divide`) and the comparison operator branch (the `default` case at line 348 for `Equal`/`NotEqual`/`LessThan`/`GreaterThan`/`LessThanOrEqual`/`GreaterThanOrEqual`): when `TensorDimensionNamesMatch` returns false, check `TensorDimensionNamesAreBroadcastable`. If compatible, call `TensorBroadcastTo` to expand the smaller tensor before calling the corresponding `Tensor*Float64` or `Tensor*Float64Mask` function.
- Integration tests in `src/main/go/calculation/postfix_op_test.go` verify:
  - **Arithmetic**: `TestEvaluatePostfixFloat64_MultiplyWithLookup` — controller with connector tensor `[Product:3, Material:3, Daily:7]` (all 1.0) and lookup tensor `[Material:3]` (values 1.0, 0.8, 0.9). Postfix `push "connector", push "lookup", apply multiply`. Result: Steel = 1.0, Aluminum = 0.8, Copper = 0.9 for all Products and Days.
  - **Comparison**: `TestEvaluatePostfixFloat64_ComparisonWithLookup` — controller with connector tensor `[Product:3, Material:3, Daily:7]` (values 5.0 for Steel, 3.0 for Aluminum, 1.0 for Copper) and lookup tensor `[Material:3]` (values 2.0, 2.0, 2.0 — all Materials have threshold 2.0). Postfix `push "connector", push "lookup", apply greater_than`. Result mask shows: Steel (5.0 > 2.0) = valid, Aluminum (3.0 > 2.0) = valid, Copper (1.0 > 2.0) = invalid for all Products and Days.
- `go test ./src/main/go/calculation/... -count=1` passes.

#### Tasks

a. Write integration tests `TestEvaluatePostfixFloat64_MultiplyWithLookup` and `TestEvaluatePostfixFloat64_ComparisonWithLookup` in `src/main/go/calculation/postfix_op_test.go`. Run tests to confirm they fail (broadcasting not yet wired in) — Not Started
b. Add `TensorDimensionNamesAreBroadcastable` helper function to `src/main/go/calculation/chronocone_tensor.go`. Write a unit test in `chronocone_tensor_test.go` for subset-compatible and incompatible cases. Run `go test ./src/main/go/calculation/... -count=1` to confirm the helper compiles and tests pass — Not Started
c. Update both the arithmetic operator branch (line ~270) and the comparison operator branch (line ~348) in `src/main/go/calculation/postfix_op.go` to use `TensorDimensionNamesAreBroadcastable` as a precondition before calling the operation — Not Started
d. Run `go test ./src/main/go/calculation/... -count=1` and verify all pass — Not Started

---

### Task 3: Fix duplicate "lookups" key in iospec.json — Completed [Small]

The file `examples/with_data/lookups/lookups_basic/iospec.json` contains two `"lookups"` arrays. The first references `product_category` with a stale `.parquet` path (deleted in Story275). The second (correctly) references `strength_adjusted_for_material`. Having two `"lookups"` keys produces parsing errors.

#### Intent

Ensure the iospec has exactly one `"lookups"` array containing the `strength_adjusted_for_material` entry.

#### Acceptance Criteria

- `examples/with_data/lookups/lookups_basic/iospec.json` contains exactly one `"lookups"` array.
- The `"lookups"` array contains one entry with `"lookup": "strength_adjusted_for_material"`.
- The JSON is valid (parseable).

#### Tasks

a. Remove the first `"lookups"` block (the `product_category` entry) from `examples/with_data/lookups/lookups_basic/iospec.json` — Not Started
b. Verify the resulting JSON is valid and has exactly one `"lookups"` array — Not Started

---

### Task 4: Delete data_a.json, create raw_strength.json — Completed [Small]

The file `examples/with_data/lookups/lookups_basic/data_a.json` contains incorrect concept values: `Product_A` instead of `WidgetA`, and date strings instead of `Sunday`–`Saturday` for the Daily dimension. The `expected_xls_generator` loads this file via glob pattern and produces incorrect expected output. Per convention, connector data should be in `<connector_name>.json` (i.e., `raw_strength.json`).

#### Intent

Replace the stale `data_a.json` with a correctly named and correctly valued `raw_strength.json`.

#### Acceptance Criteria

- `examples/with_data/lookups/lookups_basic/data_a.json` is deleted.
- `examples/with_data/lookups/lookups_basic/raw_strength.json` is created with:
  - 63 rows (3 Products × 3 Materials × 7 Days).
  - Columns: `Product`, `Material`, `Daily`, `Value`.
  - Products: `WidgetA`, `WidgetB`, `WidgetC`.
  - Materials: `Steel`, `Aluminum`, `Copper`.
  - Daily: `Sunday`, `Monday`, `Tuesday`, `Wednesday`, `Thursday`, `Friday`, `Saturday`.
  - All `Value` = 1.0.
- The existing `strength_adjusted_for_material.json` (3 rows: Steel=1.0, Aluminum=0.8, Copper=0.9) remains unchanged.

#### Tasks

a. Delete `examples/with_data/lookups/lookups_basic/data_a.json` — Not Started
b. Create `examples/with_data/lookups/lookups_basic/raw_strength.json` with the 63-row specification above — Not Started

---

### Task 5: Regenerate expected_output.xlsx — Completed

With the iospec and connector data corrected (Tasks 3 and 4), regenerate the expected output spreadsheet for verification.

#### Intent

Produce a correct `expected_output.xlsx` that reflects the fixed input data (all connector Values = 1.0, lookup values Steel=1.0, Aluminum=0.8, Copper=0.9).

#### Acceptance Criteria

- `examples/with_data/lookups/lookups_basic/expected_output.xlsx` exists and contains:
  - WidgetA × Steel × any day = 1.0
  - WidgetA × Aluminum × any day = 0.8
  - WidgetA × Copper × any day = 0.9
  - Same pattern for WidgetB and WidgetC.

#### Tasks

a. Run from `src/chronocone_planning_language`: `go run ./src/main/go/cli/expected_xls_generator/main -example examples/with_data/lookups/lookups_basic` — Not Started
b. Verify the generated XLSX contains correct expected values as specified above — Not Started

---

### Task 6: Update stale assertion in connector_loader_test.go — Completed [Small]

In `src/main/go/connector/connector_loader_test.go`, line 303, the unqualified-type check slice still references the old type name `InputSpecification` instead of the renamed `ConnectorSpecification`. This causes the test to fail when the code generator produces code with the new type name (which the test incorrectly considers "unqualified").

#### Intent

Update the assertion string to match the current type name.

#### Acceptance Criteria

- Line 303 of `src/main/go/connector/connector_loader_test.go` reads `"(inputSpec ConnectorSpecification)"` instead of `"(inputSpec InputSpecification)"`.
- `go test ./src/main/go/connector/... -count=1` passes.

#### Tasks

a. Update the assertion string on line 303 from `"(inputSpec InputSpecification)"` to `"(inputSpec ConnectorSpecification)"` in `src/main/go/connector/connector_loader_test.go` — Not Started
b. Run `go test ./src/main/go/connector/... -count=1` and verify all pass — Not Started

---

### Task 7: Rename "inputs" to "connectors" in example_arithmetic_io_spec.json — Completed [Small]

The file `example/example_arithmetic_io_spec.json` still uses the old `"inputs"` key (renamed to `"connectors"` in Story275). This file was missed by Story275 because it resides outside the `examples/with_data/` and `examples/parse_and_analyze_only/` directories.

#### Intent

Complete the `inputs` → `connectors` rename across all example files.

#### Acceptance Criteria

- `example/example_arithmetic_io_spec.json` uses `"connectors"` instead of `"inputs"`.
- The JSON is valid (parseable).

#### Tasks

a. Rename the `"inputs"` key to `"connectors"` in `example/example_arithmetic_io_spec.json` — Not Started
b. Verify JSON remains valid — Not Started

---

### Task 8: Run full test suite — Completed

After all preceding tasks are complete and reintegrated, run the full Go test suite to confirm all changes work together.

#### Intent

Verify that all packages compile and all tests pass with the combined changes from this story.

#### Acceptance Criteria

- Running `go test ./src/main/go/... -count=1 -timeout 300s` from `src/chronocone_planning_language` produces zero failures.
- The `TestWithDataExamples/lookups_basic` end-to-end test specifically passes.

#### Tasks

a. Run `go test ./src/main/go/... -count=1 -timeout 300s` from `src/chronocone_planning_language` — Not Started
b. Verify zero failures — Not Started

---

### Task 9: Run lookups_basic example end-to-end manually — Completed

As a final manual sanity check, build and run the `lookups_basic` example directly and inspect the output.

#### Intent

Confirm the generated executable produces the correct `output.parquet` file with the expected broadcasted values.

#### Acceptance Criteria

- Running `go run ../../../../src/chronocone_planning_language/generated/cmd/lookups_basic/main.go -data-dir=.` from `examples/with_data/lookups/lookups_basic` succeeds.
- `output.parquet` contains rows with `product_strength` values: 1.0 for Steel, 0.8 for Aluminum, 0.9 for Copper across all Products and Days.

#### Tasks

a. Build and run the example — Not Started
b. Verify `output.parquet` contents — Not Started

---

## Constraints

- **Test-First Development**: All coding tasks (Tasks 1, 2a, 2b) must follow TDD — write tests before implementation code.
- **File path convention**: All file paths are relative to `source_root` = `src/chronocone_planning_language`. Go test commands include the `./src/main/go/...` prefix.
- **Go version**: The project uses Go 1.24+.
- **Broadcasting approach**: Task 2b must use a `TensorDimensionNamesAreBroadcastable` precondition check (not error-string matching) to determine when broadcasting is needed. This avoids the fragility of depending on dimension-mismatch error strings.
- **Comparison operator coverage**: Task 2b explicitly requires broadcasting support for both arithmetic and comparison operators. The integration tests include a mask-based comparison case (connector > lookup via `greater_than`).
- **Tensor type convention**: `TensorBroadcastTo` and `TensorDimensionNamesAreBroadcastable` use `*ChronoconeTensor[any]` (not generic `T`), consistent with all existing tensor helper functions in the `calculation` package.
- **Dimension order sensitivity**: The broadcast helper requires matching dimensions in the same order. This is consistent with existing tensor dimension conventions.
- **Mask propagation**: When broadcasting, ensure the source tensor's mask is also broadcast correctly (lookup tensors use default all-valid masks).
- **The `lookups_basic_lookups.io.json` file**: If this file still exists (from Story275 Task 13e), delete it as part of Task 4.

## Notes

- Task 2a (TensorBroadcastTo) and Task 1 (CILLookupNode walker) are independent of each other — they were intentionally separated rather than being parts of one large task. Task 2b depends on both.
- The `example/example_arithmetic_io_spec.json` file is referenced by parsing-only tests, not by the end-to-end example tests. Fixing it ensures consistency but does not affect the `lookups_basic` integration test.
