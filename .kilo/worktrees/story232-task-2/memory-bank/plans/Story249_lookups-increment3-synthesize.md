# Plan: Story249 — Lookups Increment 3 (Synthesize)

## Purpose

This plan provides the technical writer with sufficient detail to create Story249: **lookups-increment3-synthesize** — implement CIL tree walker for `CILLookupRefNode`, Go code emission for `controller.GetMetricComputation(lookupName)`, unit compatibility verification, and metric+lookup examples (Increment 3 of Epic006).

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `lookups-increment3-synthesize`                                                               |
| description  | Implement CIL tree walker for CILLookupRefNode, GetMetricComputation(lookupName) emission, unit compatibility verification, metric+lookup examples |

## Source EROs

- [ERO006: Lookups Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO006_lookups-synthesis-phase.md) — Increment 3 `CILLookupRefNode` walking and code generation

## Dependencies

- Story248 (lookups-increment3-analyze) — needs `CILLookupRefNode` from analysis
- Story243 (lookups-increment1-synthesize) — needs lookup loading infrastructure already in place

## Files to Create or Modify

| File                                                     | Action | Purpose                                                                                    |
| -------------------------------------------------------- | ------ | ------------------------------------------------------------------------------------------ |
| `src/main/go/synthesis/cil_tree_walker.go`               | Modify | Add `WalkCILLookupRefNode` method; add helper `cellKeyCoordinates`                         |
| `src/main/go/synthesis/lookup_synthesis.go`              | Modify | Add `verifyLookupCelltypeCompatibility` helper                                              |
| `src/main/go/synthesis/preamble_generator.go`            | Modify | Ensure lookup loading precedes any metric that references a lookup                         |
| `src/main/go/synthesis/lookup_synthesis_test.go`         | Modify | Add tests for `WalkCILLookupRefNode` code emission and celltype verification               |
| `.../examples/with_data/lookups_in_metrics/`             | Create | Example: metric using a lookup (e.g., applying conversion factor to revenue)              |

### WalkCILLookupRefNode

When a `CILLookupRefNode` is encountered in a metric's CIL tree:
1. `varName := scope.lookupVarName(node.LookupName)` — generate unique variable name
2. Emit code to retrieve lookup value: `varName := controller.GetMetricComputation("lookupName").GetCell(coordinates)`
3. `scope.registerVariable(node.LookupName, varName)` — register for subsequent expression references
4. `scope.pushOperand(varName)` — push onto expression stack for evaluation

### Cell Key Coordinate Computation

`scope.cellKeyCoordinates(lookupDims)` generates Go code that constructs the coordinate tuple for indexing into the lookup tensor:
```go
func (scope *WalkScope) cellKeyCoordinates(lookupDims []string) string {
    // For each lookup dimension, get the loop iteration variable for that dimension
    var parts []string
    for _, dim := range lookupDims {
        loopVar := scope.dimensionLoopVariable(dim)
        parts = append(parts, loopVar)
    }
    return strings.Join(parts, ", ")
}
```

### Unit Compatibility Verification

`verifyLookupCelltypeCompatibility` performs a structural check that the lookup's celltype units are compatible with the metric calculation. Full unit algebra verification is done in analysis; synthesis performs a final structural pass.

### Generated Code Example

For a metric `converted_revenue` that references a `currency_conversion` lookup:
```go
func computeConvertedRevenue(controller *ChronoconeComputationController) *ChronoconeTensor[float64] {
    revenue := controller.GetMetricComputation("revenue").Tensor().(*ChronoconeTensor[float64])
    currencyConversion := controller.GetMetricComputation("currency_conversion").Tensor().(*ChronoconeTensor[float64])

    result := NewChronoconeTensor[float64](revenue.Shape())
    for i, cell := range revenue.Cells() {
        conversionRate := currencyConversion.GetCell(sourceCurrency, targetCurrency)
        result.SetCell(i, cell.Value * conversionRate)
    }
    return result
}
```

### Preamble Generator Changes

Ensure lookup data is loaded and registered on the controller before any metric that references a lookup runs. This is critical for the `controller.GetMetricComputation(...)` call to succeed at runtime.

### with_data Example: lookups_in_metrics

- CPL: metric `converted_revenue` with inputs `metric revenue` and `lookup currency_conversion`, calculate `revenue * currency_conversion`
- Lookup IO Spec: maps `currency_conversion` to parquet file
- Parquet data: currency conversion rates
- `.expected.json`: expected output after applying conversion factor

## Task Decomposition (for technical writer)

### Task 1: Implement WalkCILLookupRefNode in the CIL tree walker — Not Started
   a. Add `WalkCILLookupRefNode` method to the CIL tree walker — Not Started
   b. Generate unique variable name for the lookup value — Not Started
   c. Emit `controller.GetMetricComputation(lookupName).GetCell(...)` call — Not Started
   d. Implement `cellKeyCoordinates` helper for constructing coordinate tuples from loop variables — Not Started
   e. Register variable for subsequent expression references — Not Started
   f. Push operand onto expression stack — Not Started
   g. Write unit tests: verify emitted Go code for lookup reference access, verify variable registration — Not Started
   h. Verify walker tests pass — Not Started

### Task 2: Implement celltype compatibility verification — Not Started [Small]
   a. Add `verifyLookupCelltypeCompatibility` function — Not Started
   b. Verify structural compatibility between lookup celltype units and metric calculation — Not Started
   c. Write unit tests: compatible units, incompatible units — Not Started
   d. Verify tests pass — Not Started

### Task 3: Update preamble generator for lookup→metric ordering — Not Started [Small]
   a. Ensure generated `main()` places `loadLookupData` before metric computations that reference lookups — Not Started
   b. Write unit test: verify generated code ordering — Not Started
   c. Verify test passes — Not Started

### Task 4: Create with_data/lookups_in_metrics example — Not Started
   a. Create CPL file with metric referencing lookup via `lookup` input kind — Not Started
   b. Create Lookup IO Spec JSON — Not Started
   c. Create parquet data files (both lookup data and metric input data) — Not Started
   d. Create `.expected.json` with expected output — Not Started
   e. Verify end-to-end: parse → analyze → synthesize → run → output matches expected — Not Started

### Task 5: Verification and regression — Not Started
   a. Verify existing timeless and temporal lookup examples (Increments 1-2) still work — Not Started
   b. Run full CPL test suite to confirm no regressions — Not Started

## Execution Order

```
Task 1 → Task 2 → Task 4 → Task 5
                ↗
Task 3 ──────────┘
```

- Task 1 (CIL walker) must complete first
- Task 2 (celltype verification) depends on Task 1 (uses walker scope information)
- Task 3 (preamble ordering) is independent of Task 2 but depends on knowledge of walker output — can run after Task 1
- Task 4 (examples) depends on Tasks 1-3
- Task 5 (regression) depends on all preceding tasks

## Parallel Execution

- **Group A** (Tasks 2, 3): Task 2 creates `verifyLookupCelltypeCompatibility`, Task 3 modifies preamble generator. Disjoint files, independent concerns.

## Reintegration

**Group A (Tasks 2, 3):**
- Merge order: any order (disjoint files)
- Integration tests: `go test ./src/main/go/synthesis/... -count=1 -timeout 60s` — verify celltype verification and preamble ordering compile together
- Watch for conflicts in: `preamble_generator.go` (Task 3 modifies it; Task 2 does not)

## Acceptance Criteria Summary

1. `WalkCILLookupRefNode` emits code calling `controller.GetMetricComputation(lookupName).GetCell(...)`
2. Cell key coordinates are correctly constructed from metric loop iteration variables
3. Lookup variable is registered for subsequent expression references
4. `verifyLookupCelltypeCompatibility` validates structural unit compatibility
5. Generated preamble ensures lookups are loaded before metrics that reference them
6. `with_data/lookups_in_metrics` example passes end-to-end
7. All existing tests pass with zero regressions

## Intent Summary for Technical Writer

This story delivers the synthesis phase for lookup integration with metric calculations. When a metric's CIL tree contains a `CILLookupRefNode`, the walker emits Go code that calls `controller.GetMetricComputation(lookupName).GetCell(...)` to retrieve the lookup value for the current metric cell's dimensional context. The cell key coordinates are constructed from the metric's loop iteration variables, and the lookup value is registered as a variable for use in subsequent expression evaluation. The preamble generator ensures lookup data loading occurs before any metric that references a lookup runs.
