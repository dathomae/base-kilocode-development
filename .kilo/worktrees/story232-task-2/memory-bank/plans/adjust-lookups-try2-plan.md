# Plan: Fix Story275 — Lookups End-to-End Pipeline and Example Data

## Objective

Fix the blocking issues identified during the Story275 code review so that:
1. Lookups compile and execute end-to-end in the CPL synthesis pipeline
2. The `lookups_basic` example produces correct results
3. `go test ./src/main/go/... -count=1` passes with zero failures

## Root Cause Analysis of Story275 Failure

Story275 was a **structural refactoring** story: rename `inputs` → `connectors`, add `lookups` section to iospec, fix example data. It was NOT a functional story about making lookups work end-to-end. However, its acceptance criterion "`go test ./src/main/go/... -count=1` passes with zero failures" implicitly required the `lookups_basic` end-to-end test to pass.

**Why the end-to-end test failed:**

Two issues in combination:

1. **`CILLookupNode` missing from `postfix_walk.go`**: The CIL-to-postfix walker was written before `CILLookupNode` existed (Story224, 2026-06-08). When `CILLookupNode` was added in Story241/242 (2026-06-10), `postfix_walk.go` was never updated. Story275 didn't include this update because its scope was structural refactoring, not CIL tree walking. Result: synthesis fails with `unsupported CIL node type: *cil.CILLookupNode` for any metric referencing a lookup.

2. **No tensor broadcasting support**: Even after fixing `postfix_walk.go`, the runtime evaluator will fail. Lookup tensors have fewer dimensions than connector tensors (e.g., `strength_adjusted_for_material` has `[Material]` while `raw_strength` has `[Product, Material, Daily]`). The tensor arithmetic functions (`TensorMultiplyFloat64`, etc.) require exact dimension name match via `TensorDimensionNamesMatch`. Without broadcasting, a dimension mismatch error will occur at runtime.

**Additional implementation errors** (not specification gaps):

3. Duplicate `"lookups"` key in `lookups_basic/iospec.json`
4. Stale `data_a.json` with wrong concept values; missing `raw_strength.json`
5. Stale assertion string in connector loader test
6. Old `"inputs"` key in out-of-scope example file

## Design Principle: The IOSpec is a Spec, Not Data Storage

(This principle from the original plan is preserved and unchanged.)

The iospec.json file specifies *where* data comes from and what gets produced — it does not contain the data itself. Connector data lives in `<connector_name>.json`/`.parquet` files. Lookup data lives in `<lookup_name>.json`/`.parquet` files, following the same naming convention.

## Phase 1: Fix the CIL Postfix Walker

### Task 1: Add `CILLookupNode` case to `walkToPostfix`

**File:** `src/chronocone_planning_language/src/main/go/cil/postfix_walk.go`

**Problem:** `walkToPostfix` has no case for `*cil.CILLookupNode`. When a metric references a lookup (e.g., `product_strength = raw_strength * strength_adjusted_for_material`), the tree walker falls through to the default case and returns an error.

**Fix:** Add a case for `*CILLookupNode` that appends a `"push"` PostfixOp with the lookup name:

```go
case *CILLookupNode:
    *ops = append(*ops, calculation.PostfixOp{Kind: "push", Name: n.LookupName})
```

**Why this works at the walker level:** This is exactly analogous to the existing `CILInputRef` case (line 72: `*ops = append(*ops, calculation.PostfixOp{Kind: "push", Name: n.Name})`). At compile time, the lookup name is pushed onto the postfix table, and at runtime the evaluator resolves it from the controller.

**Why this is NOT sufficient alone:** The runtime evaluator resolves the lookup to a `*ChronoconeTensor` with dimensions `[Material]` and shape `[3]`, while the connector tensor has dimensions `[Product, Material, Daily]` and shape `[3, 3, 7]`. When the evaluator attempts to multiply them, `TensorMultiplyFloat64` will reject them with a dimension mismatch error. Task 2 (below) addresses this.

**Test updates:** Add test cases to `postfix_walk` tests that verify a `CILLookupNode` is correctly walked to produce the expected `PostfixOp`. Include at minimum:
- Lookup reference in a simple binary expression (e.g., `connector_ref * lookup_ref`)
- Ensure the generated postfix table contains the lookup name as a push op

**Constraint:** The `CILLookupNode` struct has these relevant fields:
- `LookupName string` — the name to push
- `Dimensions []string` — resolved dimension concept names (not needed for the push op itself, but needed for broadcasting in Task 2)
- `Celltype string` — cell type name (not needed for postfix)

---

## Phase 2: Implement Tensor Broadcasting for Binary Operations

### Task 2: Add dimension broadcasting to tensor arithmetic operations

**Files affected:**
- `src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor_arithmetic.go`
- `src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor_comparison.go`
- `src/chronocone_planning_language/src/main/go/calculation/postfix_op.go`
- `src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go`

**Problem:** Binary tensor operations (`TensorMultiplyFloat64`, `TensorAddFloat64`, etc.) require exact dimension name match. When a 3-dim connector tensor (Product × Material × Daily) is multiplied by a 1-dim lookup tensor (Material only), the operation fails with `"dimension mismatch"`.

**Conceptual model:** Lookups are reference data keyed by a *subset* of the connector's dimensions. For example, `strength_adjusted_for_material` has dimension `[Material]` which is a subset of `raw_strength`'s dimensions `[Product, Material, Daily]`. A lookup value for `Material=Steel` (e.g., 1.0) applies to ALL rows where `Material=Steel` regardless of Product or Daily values. This is semantically a **broadcast operation**: repeat the lookup value across the non-matching dimensions.

**Design decision:** Broadcasting belongs in the tensor arithmetic layer, not in the evaluator or code generator. This keeps the postfix walker simple and makes broadcasting available to any future CIL operation that combines tensors with different dimension sets.

**Implementation plan for broadcasting:**

#### 2a. Add `TensorBroadcastTo` helper function

Add to `chronocone_tensor.go` (or a new `chronocone_tensor_broadcast.go`):

```go
// TensorBroadcastTo broadcasts a source tensor to match the dimension
// names and shape of a target tensor. The source tensor's dimension names
// must be a subset of the target tensor's dimension names (same order,
// same names for matching dimensions). For each non-matching dimension
// in the target, the source tensor's data is repeated across all values
// of that dimension.
//
// Example: Broadcasting [Material: Steel→1.0, Al→0.8, Cu→0.9]
// to [Product, Material, Daily] repeats the three values across all
// Products and all Days, producing a 63-element tensor.
//
// Parameters:
//   - src: The source tensor to broadcast.
//   - targetDimNames: The dimension names to broadcast to.
//   - targetShape: The shape of the target tensor.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with the broadcasted data.
//   - error: An error if dimensions are incompatible.
func TensorBroadcastTo(src *ChronoconeTensor[any], targetDimNames []string, targetShape []int) (*ChronoconeTensor[any], error)
```

**Broadcasting algorithm:**
1. Validate that `src`'s dimension names are a subset of `targetDimNames` with the same names in the same order for overlapping dimensions. Non-overlapping dimensions can appear anywhere in the target list (not just at the ends).
2. Build an index mapping: for each dimension name in `targetDimNames`, find its position in `src`'s dimensions (or -1 if not present).
3. Compute the target total size from `targetShape`.
4. For each position in the target tensor, compute the corresponding source index by mapping the target multi-index back to the source's multi-index (skipping dimensions not present in the source).
5. Copy the source value at that index, repeating as needed.

**Example of dimension compatibility check:**
- Source: `[Material]`, Target: `[Product, Material, Daily]` → **compatible** (source is subset)
- Source: `[Material]`, Target: `[Product, Daily]` → **NOT compatible** (Material not in target)
- Source: `[Material, Product]`, Target: `[Product, Material, Daily]` → **compatible** (order preserved for matching dims)
- Source: `[Product, Material]`, Target: `[Material, Product]` → **NOT compatible** (order mismatch)

#### 2b. Add `TensorBinaryOpBroadcast` helper

Add to `chronocone_tensor_arithmetic.go`:

```go
// TensorBinaryOpBroadcast performs a binary operation between two tensors
// that may have different dimensions. If the dimensions don't match, the
// smaller tensor is broadcast to match the larger one before the operation.
//
// The smaller tensor's dimension names must be a subset of the larger
// tensor's dimension names (same order, same names for matching dims).
//
// Parameters:
//   - a, b: The operand tensors.
//   - scalarOp: Function to apply when both are scalar (av, bv float64).
//   - scalarTensorOp: Function to apply when one is scalar, one is tensor.
//   - tensorOp: Function to apply when one is broadcast to match the other.
//
// Returns:
//   - *ChronoconeTensor[any]: The result tensor.
//   - error: An error if broadcasting fails or operation fails.
```

Alternatively, simplify: modify each tensor operation function to call `TensorBroadcastTo` when dimensions don't match, rather than immediately returning an error.

#### 2c. Update the postfix evaluator

In `postfix_op.go`, update the binary operation handling (lines 270-282) so that when `TensorMultiplyFloat64` (or other tensor op) returns a dimension mismatch error, the evaluator falls back to broadcasting the smaller tensor and retrying:

```go
default:
    // Both tensors with dimensions.
    result, err = TensorMultiplyFloat64(a, b)
    if err != nil && strings.Contains(err.Error(), "dimension mismatch") {
        // Try broadcasting
        bBroadcast, broadcastErr := TensorBroadcastTo(b, a.DimensionNames(), a.Shape())
        if broadcastErr != nil {
            return nil, broadcastErr
        }
        result, err = TensorMultiplyFloat64(a, bBroadcast)
    }
```

**NOTE on architecture:** This fallback approach in the evaluator adds some coupling, but it's the least invasive change. The broadcasting logic itself is cleanly separated in the helper function. An alternative approach (modifying each tensor arithmetic function directly) would require changes to all arithmetic and comparison functions. The evaluator approach is preferred because it's a single change point and the error-based fallback avoids modifying existing test expectations for same-dimension operations.

**WARNING:** The error-based fallback is fragile (depends on error string matching). An alternative is to add a function like `TensorDimensionNamesAreBroadcastable` that checks subset compatibility, and use that as a precondition check before the operation, avoiding the error string dependency.

#### 2d. Update `postfix_walk.go` to store lookup dimensions for broadcasting

The current `CILLookupNode` struct has `Dimensions []string` which contains the lookup's dimension names. These need to be available to the evaluator at runtime for broadcasting. However, the evaluator resolves lookups by name from the controller, and the lookup's tensor already has its dimension names set. The evaluator can get dimension info from the tensor itself. So the walker just needs to push the lookup name, and the evaluator uses the tensor's metadata.

**However**, the evaluator needs to know which tensor to broadcast TO. In a binary expression like `raw_strength * strength_adjusted_for_material`, the postfix order is:
```
push raw_strength
push strength_adjusted_for_material
apply multiply
```
When the evaluator processes `apply multiply`, it pops both tensors from the stack. The lookup tensor will have fewer dimensions. The evaluator can detect this and broadcast the smaller one to match the larger one. This is simpler than passing metadata through the postfix ops.

**Decision:** No changes needed to `postfix_walk.go` beyond Task 1. The evaluator handles broadcasting using the tensor's own dimension metadata.

#### 2e. Tests for broadcasting

Add tests:
1. `TestTensorBroadcastTo` — unit tests for the broadcast helper:
   - Source `[Material:3]` → target `[Product:3, Material:3, Daily:7]`
   - Verify each of the 63 output cells has the correct value
   - Source is NOT a subset → error
   - Single dimension to single dimension (same) → identity
   - Empty source → empty target

2. `TestEvaluatePostfixFloat64_WithLookup` — integration test:
   - Create a controller with a connector tensor (3×3×7, all 1.0) and a lookup tensor (3 values: 1.0, 0.8, 0.9)
   - Evaluate `push "connector", push "lookup", apply multiply`
   - Verify result has correct values: Steel=1.0, Aluminum=0.8, Copper=0.9 for all Products and Days

---

## Phase 3: Fix lookups_basic Example Data

### Task 3: Fix duplicate `"lookups"` key in iospec.json

**File:** `examples/with_data/lookups/lookups_basic/iospec.json`

**Problem:** The JSON has two `"lookups"` arrays. The first (lines 14-25) references `product_category` with path `lookups_basic_product_category.parquet` (old, deleted file). The second (lines 39-50) references `strength_adjusted_for_material` (correct).

**Fix:** Remove the first `"lookups"` block entirely. The final file should have exactly one `"lookups"` array with one entry for `strength_adjusted_for_material`.

**Current incorrect structure (simplified):**
```json
{
    "connectors": [{"connector": "raw_strength", ...}],
    "lookups": [
        {"lookup": "product_category", ...}    ← DELETE THIS BLOCK
    ],
    "outputs": [{"metric": "product_strength", ...}],
    "lookups": [                               ← KEEP THIS BLOCK
        {"lookup": "strength_adjusted_for_material", ...}
    ],
    "architecture": {...}
}
```

### Task 4: Fix data files — delete `data_a.json`, create `raw_strength.json`

**Files affected:**
- **DELETE:** `examples/with_data/lookups/lookups_basic/data_a.json`
- **CREATE:** `examples/with_data/lookups/lookups_basic/raw_strength.json`

**Problem:** `data_a.json` contains incorrect concept values (`Product_A` instead of `WidgetA`; date strings instead of `Sunday`-`Saturday` for Daily). The `expected_xls_generator` loads this file via glob pattern (`data_*.json`) and uses it as the connector data, producing incorrect expected output. The story requires `raw_strength.json` (named after the connector per convention), not `data_a.json`.

**`raw_strength.json` specification:**
- 63 rows (3 Products × 3 Materials × 7 Days)
- Columns: `Product`, `Material`, `Daily`, `Value`
- Products: `WidgetA`, `WidgetB`, `WidgetC`
- Materials: `Steel`, `Aluminum`, `Copper`
- Daily: `Sunday`, `Monday`, `Tuesday`, `Wednesday`, `Thursday`, `Friday`, `Saturday`
- All `Value` = 1.0

**Note:** The `raw_strength.parquet` file already exists and is correct. We only need to create `raw_strength.json` for the expected_xls_generator tool (which reads JSON, not parquet) and for human inspectability.

**Also verify these files exist and are correct (they should already be OK):**
- `strength_adjusted_for_material.json` — 3 rows: Steel=1.0, Aluminum=0.8, Copper=0.9
- `strength_adjusted_for_material.parquet` — same data in parquet format

### Task 5: Regenerate expected_output.xlsx

**Command:** Run the expected_xls_generator:
```
cd src/chronocone_planning_language
go run ./src/main/go/cli/expected_xls_generator/main -example examples/with_data/lookups/lookups_basic
```

**Verify:** After generation, verify that `expected_output.xlsx` contains correct expected values:
- WidgetA + Steel on any day = 1.0 × 1.0 = 1.0
- WidgetA + Aluminum on any day = 1.0 × 0.8 = 0.8
- WidgetA + Copper on any day = 1.0 × 0.9 = 0.9
- (Same for WidgetB, WidgetC)

**NOTE:** This regeneration requires `raw_strength.json` to exist (Task 4) and the iospec.json to be correct (Task 3). Run after Tasks 3 and 4 are complete. The generated XLSX will have empty formula cells for lookup references (since `cil_to_xlsx.go` returns `""` for `CILLookupNode`), but the computed values will be correct because the evaluator handles lookups.

---

## Phase 4: Fix Minor Issues

### Task 6: Update stale assertion in connector_loader_test.go

**File:** `src/chronocone_planning_language/src/main/go/connector/connector_loader_test.go`

**Line 303:** Update `"(inputSpec InputSpecification)"` → `"(inputSpec ConnectorSpecification)"` in the `unqualified` slice. This assertion checks that generated code doesn't contain unqualified type names. Since the type was renamed, the check string should be updated to match the new type name.

### Task 7: Update `example/example_arithmetic_io_spec.json`

**File:** `example/example_arithmetic_io_spec.json`

**Fix:** Rename `"inputs"` → `"connectors"` in this file. This file was missed by Tasks 10-11 of Story275 because it's outside the `examples/with_data/` and `examples/parse_and_analyze_only/` directories.

---

## Phase 5: Validation

### Task 8: Run full test suite

```
cd src/chronocone_planning_language
go test ./src/main/go/... -count=1 -timeout 300s
```

**Expected:** All tests pass with zero failures, including:
- `TestWithDataExamples/lookups_basic` — compiles, runs, produces correct output
- All `io/` package tests
- All `connector/` package tests
- All `synthesis/` package tests
- All `calculation/` package tests (new broadcasting tests)
- All `compiler/` package tests
- All `expected_xls_generator/` package tests

### Task 9: Run the lookups_basic example end-to-end manually

```
cd examples/with_data/lookups/lookups_basic
go run ../../../../src/chronocone_planning_language/generated/cmd/lookups_basic/main.go -data-dir=.
```

**Verify:** The output `output.parquet` contains the expected values (1.0 for Steel, 0.8 for Aluminum, 0.9 for Copper across all Products and Days).

---

## Execution Order and Dependencies

```
[Task 1 ‖ Task 3 ‖ Task 4 ‖ Task 6 ‖ Task 7]  ← Independent, can run in parallel
         ↓
[Task 2]  ← Depends on Task 1 (needs CILLookupNode case to test broadcasting integration)
         ↓
[Task 5]  ← Depends on Tasks 3 and 4 (needs correct iospec.json and raw_strength.json)
         ↓
[Task 8]  ← Depends on all preceding tasks
         ↓
[Task 9]  ← Depends on Task 8
```

Tasks 1, 3, 4, 6, and 7 are independent — they modify disjoint files and can run in parallel. Task 2 depends on Task 1 because we should write broadcasting integration tests that reference the `CILLookupNode` case in the walker. Task 5 (regenerate XLSX) depends on Tasks 3 and 4 for correct source data. Tasks 8 and 9 are sequential validation gates.

## Risks and Considerations

1. **Broadcasting could have edge cases with comparison operators**: Broadcasting a lookup for `connector > lookup` is semantically different from `connector * lookup`. The postfix evaluator handles comparisons by converting to masks, and the broadcasting must preserve the intent. Test thoroughly with comparison operators involving lookups.

2. **Dimension order sensitivity**: The broadcast helper requires matching dimensions to appear in the same order. If `core.cpl` defines dimensions as `[Product, Material, Daily]` but the lookup uses `[Material]`, the order is preserved for the matching dimension. This should work correctly.

3. **Performance**: Broadcasting creates a new tensor by repeating data. For large tensors this could be expensive. In practice, lookup tensors are small (a few values), so broadcasting is cheap.

4. **Mask propagation**: When broadcasting, the source tensor's mask must also be broadcast. Currently lookup tensors have default masks (all valid). The helper should handle this correctly.

5. **The `lookups_basic_lookups.io.json` file**: Task 13e of Story275 required deleting this file. Verify it doesn't exist (it should have been deleted). If it still exists, delete it.

## Comparison with Original Story275

| Aspect | Story275 (Original) | This Plan (Try2) |
|--------|---------------------|-------------------|
| `postfix_walk.go` | NOT addressed | Task 1: Add `CILLookupNode` case |
| Tensor broadcasting | NOT addressed | Task 2: Full broadcasting support |
| iospec.json duplicate key | Task 12d (implemented incorrectly) | Task 3: Explicit fix instructions |
| `data_a.json` / missing `raw_strength.json` | Task 13a (not followed correctly) | Task 4: Explicit delete + create instructions |
| Connector loader test assertion | Task 6c (missed this string) | Task 6: Explicit fix |
| Old example file | Not in scope | Task 7: Fix |
| Full test suite | Task 16 (didn't pass) | Task 8: Must pass including lookups_basic |

## Clarification Points for the Technical Writer

When writing the story from this plan, please ensure:

1. **Tasks 1 and 2 are the critical path.** Without them, lookups don't work end-to-end. These tasks were the gap in Story275.
2. **Task 2 (broadcasting) is the most complex.** Be very specific about the broadcasting algorithm, the test cases, and where the changes go. The plan provides a detailed algorithm above.
3. **Tasks 3 and 4 are data fixes.** The previous implementor made errors here (duplicate key, wrong file). Be explicit about exact file contents.
4. **All tasks should follow TDD:** Write/update tests first, then implement, then verify tests pass.
5. **The `lookups_basic` example is the integration test.** If `TestWithDataExamples/lookups_basic` passes, the core functionality is working.
6. **Consider splitting Tasks 1 and 2 into separate subtasks for clarity:**
   - Task 1a: Add `CILLookupNode` case to walker (1 line)
   - Task 1b: Add postfix walker test for `CILLookupNode`
   - Task 2a: Implement `TensorBroadcastTo` helper
   - Task 2b: Update postfix evaluator to use broadcasting
   - Task 2c: Add broadcasting tests
7. **Document the `LookupSpecification` struct and `iospec.json` format** as reference for future stories.
