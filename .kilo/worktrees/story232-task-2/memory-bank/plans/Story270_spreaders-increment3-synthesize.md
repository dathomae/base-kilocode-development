# Plan: Story270 — Spreaders Increment 3 (Synthesize)

## Purpose

This plan provides the technical writer with sufficient detail to create Story270: **spreaders-increment3-synthesize** — implement synthesis (code generation) for spreaders that reference metrics, including computing referenced metrics (with implicit aggregators) before spread application, dynamic spreading matrix construction from metric data, and weighted distribution examples (Increment 3 of Epic008).

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `spreaders-increment3-synthesize`                                                            |
| description  | Generate code for referenced metric computation with implicit aggregators, dynamic spreading matrix construction from metric data, weighted distribution examples |

## Source EROs

- [ERO008: Spreaders Incremental Implementation](../specifications/design/cpl/eros/ERO008_spreaders-incremental-implementation.md) — Increment 3 sketch
- [ERO008: Spreaders Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO008_spreaders-synthesis-phase.md) — Increment 3 generated code examples

## Dependencies

- Story269 (spreaders-increment3-analyze) — analysis must produce `CILSpreadNode` with populated `ReferencedMetrics` slice (including `CILMetricRefNode` and `CILAggregateNode` for implicit aggregators)
- Story268 (spreaders-increment2-synthesize) — CIL tree walking infrastructure for spreader-integrated metrics must exist
- Story261 (aggregators-increment4-synthesize) — `ChronoconeComputationController`, topological scheduling, and implicit aggregator code generation pattern established
- Story254 (aggregators-increment1-matrix-mult) — `tensorops.ReduceDimension` (for implicit aggregators) and `tensorops.ExpandDimension` (for the spread itself) are the synthesis primitives

## Files to Create or Modify

| File                                                    | Action | Purpose                                                                          |
| ------------------------------------------------------- | ------ | -------------------------------------------------------------------------------- |
| `src/main/go/synthesis/spreader_metric_ref_synth.go`    | Create | Walk `ReferencedMetrics` subtrees, emit code for metric computation + implicit aggregation before spread |
| `src/main/go/synthesis/dynamic_spread_matrix.go`        | Create | Generate dynamic spreading matrix construction code from metric data at runtime   |
| `src/main/go/synthesis/spreader_metric_ref_synth_test.go` | Create | Unit tests: verify generated code for weighted distribution spreaders           |
| `examples/spreaders/metric_references/` (in appropriate with_data dir) | Create | End-to-end example with spreader referencing a metric for weighted distribution |

### Generated Code Pattern — Weighted Distribution

For a spreader `from [Product] to [Country, Product] = Product / customers` where `customers` is a `[Country]` metric that needs implicit aggregation (`from [Country] to [] = 1/sum(Country)`):

```go
func computeWeightedRegionalTargets(controller *ChronoconeComputationController) (*ChronoconeTensor[float64], error) {
    sourceProduct := controller.GetMetricComputation("source_product").Tensor().(*ChronoconeTensor[float64])

    // 1. Compute referenced metric: customers
    customers := controller.GetMetricComputation("customers").Tensor().(*ChronoconeTensor[float64])

    // 2. Apply implicit aggregator: from [Country] to [] = 1/sum(Country)
    customersScalar, err := tensorops.ReduceDimension(
        customers,
        aggregationMatrix_1_sum_Country,
        0, 0,  // reduce dim 0 (Country) → scalar
    )
    if err != nil {
        return nil, fmt.Errorf("implicit aggregation of customers failed: %w", err)
    }

    // 3. Build dynamic spreading matrix from computed weights
    //    For each product: weight per country = product_value / total_customers
    spreadingMatrix := buildDynamicSpreadingMatrix(sourceProduct, customersScalar, countryValues)
    // At runtime, this produces: matrix.Data[0] = [product_widget/customers, product_gadget/customers]

    // 4. Apply spreader
    result, err := tensorops.ExpandDimension(
        sourceProduct,
        spreadingMatrix,
        0,  // source dim: Product
        0,  // new dim insertion point: Country (before Product)
    )
    if err != nil {
        return nil, fmt.Errorf("weighted spread failed: %w", err)
    }

    return result, nil
}
```

### Walking ReferencedMetrics Subtrees

The `CILSpreadNode` walker is updated to first walk the `ReferencedMetrics` subtrees:

```go
func (w *CILTreeWalker) WalkCILSpreadNode(node *CILSpreadNode, scope *WalkScope) error {
    // Walk referenced metrics first (with implicit aggregators)
    for _, refNode := range node.ReferencedMetrics {
        // Walk each referenced metric subtree
        // If it's a CILMetricRefNode → emit controller.GetMetricComputation(name)
        // If it's CILAggregateNode wrapping CILMetricRefNode → emit ReduceDimension call
        refVar := w.walkSubtree(refNode, scope)
        scope.registerVariable(refVar + "_tensor", refNode)
    }

    // Walk the source CIL subtree
    sourceVar := w.walkSubtree(node.SourceCIL, scope)

    // If ReferencedMetrics is non-nil, emit dynamic matrix construction
    if len(node.ReferencedMetrics) > 0 {
        // Emit: spreadingMatrix := buildDynamicSpreadingMatrix(sourceVar, refTensors, colLabels)
        matrixVar := scope.freshVar("dynamicMatrix")
        scope.emitDynamicMatrixConstruction(matrixVar, sourceVar, node)
        resultVar := scope.freshVar("spread_" + sanitizeVarName(node.NewDimension))
        // Emit: result, err := tensorops.ExpandDimension(sourceVar, matrixVar, srcDim, newDim)
        // ...
    } else {
        // Static matrix (from Increment 1): use pre-built spreadingMatrix_X
        // ... existing code from Story266 ...
    }
}
```

### Dynamic Spreading Matrix Construction

```go
// buildDynamicSpreadingMatrix emits runtime Go code that constructs
// a spreading matrix from computed metric data.
func (w *CILTreeWalker) emitDynamicMatrixConstruction(
    matrixVar string,
    sourceVar string,
    node *CILSpreadNode,
    scope *WalkScope,
) {
    // For equal-proportion: weights are uniform (1/N per new dimension value)
    // For weighted: weights are computed from referenced metric data
    //
    // Generated code:
    //   dynamicMatrix := &tensorops.AggregationMatrix{
    //       Rows:       1,  // one row per source cell (expands during iteration)
    //       Columns:    len(countryValues),
    //       ColLabels:  []string{"USA", "Canada"},
    //       Data:       make([][]float64, 1),
    //   }
    //   ...
    //   weight := sourceValue / customersScalarValue
    //   dynamicMatrix.Data[0][j] = weight
}
```

### Topological Scheduling with Referenced Metrics

The `ChronoconeComputationController` must ensure referenced metrics are computed before the spreader that references them. The topological sort naturally handles this because:

1. During analysis, `CILMetricRefNode` references are recorded as dependencies
2. The DAG already includes these dependency edges
3. The topological scheduler ensures correct computation order

The synthesizer does not need to add additional scheduling — the topological order from analysis is correct.

### Generated Code Structure — Full Picture

```
// Preamble
var spreadingMatrix_Country = &tensorops.AggregationMatrix{...}      // Static spread matrices
var aggregationMatrix_1_sum_Country = &tensorops.AggregationMatrix{...} // Aggregator matrices

// Computation functions (topological order)
// 1. customers (base [Country] metric — connector or simple calculation)
// 2. source_product (base [Product] metric)
// 3. weighted_targets (spreader metric — references products and customers)
//    - Gets source_product tensor
//    - Gets customers tensor
//    - Applies implicit aggregator: ReduceDimension(customers, aggMatrix, ...)
//    - Builds dynamic spreading matrix from product/customers
//    - Applies spread: ExpandDimension(source_product, dynamicMatrix, ...)
//    - Returns result
```

### End-to-End Example: with_data/spreaders_metric_references/

Directory structure:
```
examples/spreaders/metric_references/
├── spreaders_metric_refs.cpl         # CPL with spreaders referencing metrics
├── spreaders_metric_refs.io.yaml     # IO spec
├── spreaders_metric_refs_input.xlsx  # Input data (products, customers)
├── spreaders_metric_refs.expected.json # Expected output with weighted distribution
```

CPL example scenario:
- `customers` is a `[Country]` connector metric: `USA=1000, Canada=500`
- `source_product` is a `[Product]` metric: `widget=300, gadget=600`
- Spreader: `from [Product] to [Country, Product] = Product / customers`
  - Implicit aggregation of `customers`: `from [Country] to [] = 1/sum(Country)` → `customersScalar = 1500`
  - For each product, weight per country = `product_value / 1500`
  - Widget: `300/1500 = 0.2` → USA gets 60, Canada gets 30 (of widget)
  - Gadget: `600/1500 = 0.4` → USA gets 80, Canada gets 40 (of gadget)

### Test Cases

1. **Weighted distribution**: Spreader `Product / customers` with `customers` [Country] metric + implicit aggregator → correct weighted spread results
2. **Direct metric reference (no implicit agg)**: Metric dims match spreader needs → no implicit aggregation emitted
3. **Static distribution still works**: Spreader without metric references uses pre-built matrix (regression test from Increment 1)
4. **Multiple referenced metrics**: Spreader expression `Product * metricA / metricB` → both metrics computed before spread
5. **Correct topological order**: Referenced metric computed before spreader metric in generated code
6. **Dynamic matrix shape**: Verified matrix has correct dimensions and column labels
7. **Error propagation**: Metric computation failure, implicit aggregation failure, spread failure — all produce error messages

## Task Decomposition (for technical writer)

### Task 1: Implement ReferencedMetrics subtree walking — Not Started
   a. Update `WalkCILSpreadNode` in `src/main/go/synthesis/cil_spread_walker.go` to first walk `node.ReferencedMetrics` before walking `SourceCIL` — Not Started
   b. Implement walking for `CILMetricRefNode`: emit `controller.GetMetricComputation(name).Tensor()` call — Not Started
   c. Implement walking for `CILAggregateNode` wrapping `CILMetricRefNode`: emit `ReduceDimension` call with correct aggregation matrix — Not Started
   d. Write unit tests: verify walker emits `GetMetricComputation` for direct metric ref, emits `ReduceDimension` for implicitly aggregated metric ref — Not Started
   e. Verify walker tests pass — Not Started

### Task 2: Implement dynamic spreading matrix code generation — Not Started
   a. Create `src/main/go/synthesis/dynamic_spread_matrix.go` — implement `emitDynamicMatrixConstruction` — Not Started
   b. Generate code that constructs `AggregationMatrix` at runtime from metric data values — Not Started
   c. Handle equal-proportion case (uniform `1/N` weights when no metric reference) vs weighted case (metric-derived weights) — Not Started
   d. Write unit tests: verify generated code creates matrix with correct dimensions, verify weight computation for weighted distribution, verify uniform weights for equal proportion — Not Started
   e. Verify dynamic matrix tests pass — Not Started

### Task 3: Integrate metric reference synthesis into full pipeline — Not Started
   a. Wire Increment 3 synthesis into the metric computation code generator — Not Started
   b. Ensure topological ordering respects metric references: referenced metrics scheduled before spreader — Not Started
   c. Ensure all spreading + aggregation matrices needed are emitted in the preamble — Not Started
   d. Write integration tests: full generated program for weighted distribution scenario, verify generated code compiles, verify output tensor shapes and values — Not Started
   e. Verify integration tests pass — Not Started

### Task 4: Create end-to-end weighted distribution examples — Not Started [Small]
   a. Create `with_data/spreaders_metric_references/` directory with CPL + IO spec + input data + expected output — Not Started
   b. Example 1: Weighted distribution via `Product / customers` with implicit aggregator — Not Started
   c. Example 2: Weighted distribution where metric dims already match (no implicit agg) — Not Started
   d. Example 3: Metric reference with equal-proportion fallback (metric dims match, used as weight factor) — Not Started
   e. Verify all examples produce output matching `.expected.json` — Not Started

## Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4
```

- Task 1 must complete first (subtree walking is the entry point)
- Task 2 depends on Task 1 (dynamic matrix construction uses walked result variables)
- Task 3 depends on Tasks 1 and 2 (integration wires everything into the pipeline)
- Task 4 depends on all preceding tasks (examples verify full pipeline)

No parallel groups — each task builds on the previous.

## Acceptance Criteria Summary

1. `CILSpreadNode.ReferencedMetrics` subtrees are walked before the source tree
2. `CILMetricRefNode` emits `controller.GetMetricComputation(name).Tensor()` call
3. `CILAggregateNode` wrapping `CILMetricRefNode` emits `tensorops.ReduceDimension` call
4. Dynamic spreading matrix constructed at runtime from computed metric values
5. Static spreading matrix still used when no metric references (regression from Increment 1)
6. Topological computation order ensures referenced metrics are computed before spreader
7. Weighted distribution produces correct results matching expected output
8. Equal-proportion distribution still works (uniform weights, no metric reference)
9. Error propagation: metric failures, aggregation failures, spread failures all produce meaningful errors
10. End-to-end examples with metric references pass verification

## Intent Summary for Technical Writer

This story delivers the synthesis phase for spreaders that reference metrics. When the Increment 3 analyzer populates `CILSpreadNode.ReferencedMetrics`, the walker first walks these subtrees to emit code that computes the referenced metrics (with implicit aggregator calls if needed) before the spread operation. Dynamic spreading matrices are constructed at runtime from the computed metric data (e.g., `product_value / total_customers` as the per-country weight). This is different from Increment 1's static pre-built matrices — the matrix is constructed in generated code using runtime data. The `ChronoconeComputationController` handles topological scheduling naturally because the analysis-phase dependency graph already includes metric references as edges. The walker must correctly handle the three cases: (1) direct metric reference with matching dimensions → `GetMetricComputation`, (2) metric reference needing implicit aggregation → `ReduceDimension(GetMetricComputation(...), ...)`, (3) no metric reference → static pre-built matrix (Increment 1 behavior, regression-preserved).

Note: The `ReduceDimension` calls for implicit aggregators reuse the same aggregator matrix preamble variables emitted by the aggregator synthesis story (Story257/Story261). The spreader synthesizer should reference these variables by name rather than re-emitting them.
