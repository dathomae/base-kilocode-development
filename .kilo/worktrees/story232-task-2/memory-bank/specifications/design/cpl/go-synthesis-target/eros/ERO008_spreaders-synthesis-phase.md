# ERO008: Spreaders Synthesis Phase

## Title

Spreaders Synthesis Phase — Code Generation, Matrix Multiplication, and Runtime Integration

## Purpose/Goal

Define the synthesis phase implementation for spreaders, including CIL tree walking for `CILSpreadNode` nodes, Go code generation for dimension-expanding matrix multiplication, metric reference computation, and implicit aggregator execution, as defined in [ERO008 — Spreaders Incremental Implementation](../../eros/ERO008_spreaders-incremental-implementation.md).

## Decision Status

Draft

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic008_cpl-spreaders-implementation.md](../../../../../epics/Epic008_cpl-spreaders-implementation.md)

## Context

### Situation

The synthesis phase is responsible for walking spreader CIL trees and generating executable Go code. Spreaders use the **dimension-expanding** mode of the tensor matrix multiplication infrastructure built in ERO007 Increment 1. No new matrix multiplication infrastructure is needed — spreaders are the first consumer of the expand path.

Spreader synthesis has additional responsibilities beyond aggregators:
- Computing referenced metrics (including implicit aggregator results) before applying the spread
- Constructing spreading matrices from distribution expressions
- Handling `from []` scalar spreading (rank-zero origin)
- Fusing spreader and aggregator sequences in Increment 4

### Constraints

- Must reuse the existing tensor matrix multiplication infrastructure (ERO007 Increment 1)
- The `ExpandDimension` function from ERO007 handles the tensor expansion
- Referenced metrics must be computed and available before spread application
- Implicit aggregator nodes within spreader references must be executed first
- Types are initially represented as `float64`

### Requirements

- Walk `CILSpreadNode` in the CIL tree walker and emit `ExpandDimension` calls
- Generate Go code that applies spreading matrices to source tensors
- Emit code for referenced metrics (including implicit aggregators) before spread execution
- Support scalar spreading (`from []`) producing rank-one tensors
- Support fused spreader/aggregator operations in Increment 4

## Data Structures and Algorithms by Increment

### Increment 1: Basic Spreader Parsing and Validation

**Overview**

Increment 1 implements CIL tree walking for `CILSpreadNode`. The walker emits Go code that applies the spreading matrix using `tensorops.ExpandDimension`. Spreaders are applied manually in test code.

**CIL Tree Walker — CILSpreadNode Handler (Increment 1)**

```go
// WalkCILSpreadNode handles a CILSpreadNode in the CIL tree.
// It generates Go code that applies the spreading matrix to the source tensor
// using the dimension-expanding matrix multiplication from ERO007 Increment 1.
func (w *CILTreeWalker) WalkCILSpreadNode(node *CILSpreadNode, scope *WalkScope) error {
    // Walk the source CIL subtree first
    sourceVar := w.walkSubtree(node.SourceCIL, scope)

    resultVar := scope.freshVar("spread_" + node.NewDimension)

    // Determine source dimension index and new dimension insertion point
    srcDim := scope.getDimensionIndex(node.FromDimensions, node.FromDimensions[0]) // base dimension
    newDim := len(node.FromDimensions) // Insert after existing dimensions

    // Handle from [] (scalar) case
    if node.FromEmpty {
        scope.emitCodeIndented(scope.indentLevel, fmt.Sprintf(
            `%s, err := tensorops.ExpandDimension(%s, spreadingMatrix_%s, 0, 0)`,
            resultVar, sourceVar, sanitizeVarName(node.NewDimension),
        ))
    } else {
        scope.emitCodeIndented(scope.indentLevel, fmt.Sprintf(
            `%s, err := tensorops.ExpandDimension(%s, spreadingMatrix_%s, %d, %d)`,
            resultVar, sourceVar, sanitizeVarName(node.NewDimension), srcDim, newDim,
        ))
    }

    scope.emitCodeIndented(scope.indentLevel, "if err != nil {")
    scope.emitCodeIndented(scope.indentLevel+1,
        `return nil, fmt.Errorf("spread(%s) failed: %%w", err)`)
    scope.emitCodeIndented(scope.indentLevel, "}")

    scope.registerVariable(resultVar, node)
    scope.pushOperand(resultVar)
    return nil
}
```

**Spreading Matrix Emission (Increment 1)**

```go
func (g *PreambleGenerator) EmitSpreadingMatrix(
    buf io.Writer,
    name string,
    matrix *AggregationMatrix,
) error {
    tmpl := `
var spreadingMatrix_%s = &tensorops.AggregationMatrix{
    Rows:    %d,
    Columns: %d,
    ColLabels: []string{%s},
    Data: [][]float64{
        %s
    },
}
`
    colLabels := strings.Join(quoteStrings(matrix.ColLabels), ", ")
    var dataLines []string
    for _, row := range matrix.Data {
        rowStrs := make([]string, len(row))
        for j, v := range row {
            rowStrs[j] = strconv.FormatFloat(v, 'f', -1, 64)
        }
        dataLines = append(dataLines, "{"+strings.Join(rowStrs, ", ")+"},")
    }
    dataStr := strings.Join(dataLines, "\n\t\t")
    rendered := fmt.Sprintf(tmpl, sanitizeVarName(name), matrix.Rows, matrix.Columns,
        colLabels, dataStr)
    buf.Write([]byte(rendered))
    return nil
}
```

**Generated Code Example — Manual Spreading (Increment 1)**

```go
// Spreading matrix for equal distribution across countries
// from [Product] to [Country, Product] = Product * 1/size(Country) * Country
var spreadingMatrix_Country = &tensorops.AggregationMatrix{
    Rows:    1,
    Columns: 2,
    ColLabels: []string{"USA", "Canada"},
    Data: [][]float64{
        {0.5, 0.5},  // Equal weight to each country
    },
}

func TestSpreader_EqualDistribution(t *testing.T) {
    // Source tensor: 1D, values per product [100.0, 200.0]
    sourceTensor := chronocone_compute.NewChronoconeTensor[float64](
        []float64{100.0, 200.0},
        []int{2},
    )

    // Apply spreader
    result, err := tensorops.ExpandDimension(
        sourceTensor,
        spreadingMatrix_Country,
        0,  // srcDim
        0,  // newDim inserted before Product
    )
    if err != nil {
        t.Fatalf("spread failed: %v", err)
    }

    // Result tensor: 2D [Country=2 × Product=2]
    // result[0] = 100.0 * 0.5 = 50.0  (USA, widget)
    // result[1] = 100.0 * 0.5 = 50.0  (Canada, widget)
    // result[2] = 200.0 * 0.5 = 100.0 (USA, gadget)
    // result[3] = 200.0 * 0.5 = 100.0 (Canada, gadget)
    assert.Equal(t, []int{2, 2}, result.Shape())
}
```

**Distributing Expression Weights to Matrix (Increment 1)**

For expressions like `Product * 1/size(Country) * Country`:
- `1/size(Country)` computes to `1.0 / len(country_values)` = `0.5` for 2 countries
- `Product` is a per-product multiplier (handled during iteration)
- `Country` indicates the distribution is across country values
- The matrix entry for each new dimension value is the weight from the expression

```go
func computeSpreaderWeights(
    spreaderExpr CILNode,
    newDimValues []string,
    conceptValues map[string][]string,
) (map[string]float64, error) {

    // For 1/size(Country): each country gets 1/N
    // For expression with metric references: weights come from metric data (Increment 3)

    weights := make(map[string]float64)
    switch expr := spreaderExpr.(type) {
    case *CILSpreadInverseSizeNode:
        n := float64(len(newDimValues))
        for _, v := range newDimValues {
            weights[v] = 1.0 / n
        }
    case *CILSpreadSizeNode:
        for _, v := range newDimValues {
            weights[v] = 1.0  // size(C) distributes the count
        }
    }
    return weights, nil
}
```

### Increment 2: Spreader Selection and Integration

**Overview**

Increment 2 integrates spreaders into metric computation. The CIL tree walker generates code that applies spreader sequences as part of the metric's calculation pipeline.

**Metric Computation with Spreaders (Increment 2)**

```go
// Generated Go code for metric: regional_targets
//   inputs: metric global_target (scalar)
//   output: [Country, Product]

func computeRegionalTargets(controller *ChronoconeComputationController) (*ChronoconeTensor[float64], error) {
    globalTarget := controller.GetMetricComputation("global_target").Tensor().(*ChronoconeTensor[float64])

    // Spread from [] to [Country] = 1/size(Country) * Country
    spreadToCountry, err := tensorops.ExpandDimension(
        globalTarget,
        spreadingMatrix_Country,
        0, 0,
    )
    if err != nil {
        return nil, fmt.Errorf("spread to country failed: %w", err)
    }

    // Spread from [Country] to [Country, Product] = 1/size(Product) * Product
    result, err := tensorops.ExpandDimension(
        spreadToCountry,
        spreadingMatrix_Product,
        0, 1,
    )
    if err != nil {
        return nil, fmt.Errorf("spread to product failed: %w", err)
    }

    return result, nil
}
```

### Increment 3: Metric References in Spreader Expressions

**Overview**

Increment 3 enables spreaders to reference metric data for weighted distributions. Referenced metrics (including those with implicit aggregators) are computed before the spread operation.

**Generated Code — Weighted Distribution via Metric (Increment 3)**

```go
// Spreader: from [Product] to [Country, Product] = Product / customers
// where customers is a [Country] metric that needs implicit aggregation

func computeWeightedSpread(controller *ChronoconeComputationController) (*ChronoconeTensor[float64], error) {
    sourceProduct := controller.GetMetricComputation("source_product").Tensor().(*ChronoconeTensor[float64])
    customers := controller.GetMetricComputation("customers").Tensor().(*ChronoconeTensor[float64])

    // Implicit aggregator: from [Country] to [] = 1/sum(Country)
    customersScalar, err := tensorops.ReduceDimension(
        customers,
        aggregationMatrix_1_sum_Country,
        0, 0,
    )
    if err != nil {
        return nil, fmt.Errorf("implicit aggregation of customers failed: %w", err)
    }

    // Compute product / customers proportion
    // (simplified — full implementation iterates over cells)
    proportion := sourceProduct.Data[0] / customersScalar.Data[0]

    // Build dynamic spreading matrix from computed weights
    spreadingMatrix := buildDynamicSpreadingMatrix(proportion, countryValues)

    // Apply spread
    result, err := tensorops.ExpandDimension(
        sourceProduct,
        spreadingMatrix,
        0, 0,
    )
    if err != nil {
        return nil, fmt.Errorf("weighted spread failed: %w", err)
    }

    return result, nil
}
```

### Increment 4: Spreader/Aggregator Interleaving and Optimization

**Overview**

Increment 4 optimizes mixed spreader/aggregator sequences by fusing adjacent operations where possible.

**Fused Operations (Increment 4)**

```go
// EmitFusedSpreaderSequence combines adjacent spreading matrices
// and interleaved aggregator+spreader pairs into fused operations.
func (w *CILTreeWalker) EmitFusedSpreaderSequence(
    ops []DimensionalOp,
    sourceVar string,
    scope *WalkScope,
) (string, error) {

    fusedMatrices := []*AggregationMatrix{}
    currentMatrix := ops[0].Matrix

    for i := 1; i < len(ops); i++ {
        nextMatrix := ops[i].Matrix
        // Aggregate then spread on different dimensions can be fused
        if ops[i-1].OpType == "aggregate" && ops[i].OpType == "spread" &&
            ops[i-1].Dimension != ops[i].NewDimension {
            // Fuse: apply spread to intermediate aggregated tensor
            fused := fuseAndSpread(currentMatrix, nextMatrix)
            currentMatrix = fused
        } else if currentMatrix.Columns == nextMatrix.Rows {
            currentMatrix = fuseAggregationMatrices(currentMatrix, nextMatrix)
        } else {
            fusedMatrices = append(fusedMatrices, currentMatrix)
            currentMatrix = nextMatrix
        }
    }
    fusedMatrices = append(fusedMatrices, currentMatrix)

    // Emit code for each fused operation
    resultVar := sourceVar
    for i, matrix := range fusedMatrices {
        nextVar := scope.freshVar(fmt.Sprintf("fused_%d", i))
        // ... emit ExpandDimension or ReduceDimension based on op type ...
        resultVar = nextVar
    }

    scope.pushOperand(resultVar)
    return resultVar, nil
}
```

**Scalar Spreading Edge Case (Increment 4)**

```go
func (w *CILTreeWalker) handleScalarSpread(node *CILSpreadNode, sourceVar string, scope *WalkScope) {
    resultVar := scope.freshVar("scalar_spread")
    // from [] to [C] — source is rank-zero, result is rank-one
    scope.emitCodeIndented(scope.indentLevel, fmt.Sprintf(
        `%s, err := tensorops.ExpandDimension(%s, spreadingMatrix_%s, 0, 0)`,
        resultVar, sourceVar, sanitizeVarName(node.NewDimension),
    ))
    scope.registerVariable(resultVar, node)
    scope.pushOperand(resultVar)
}
```

## Generated Code Structure (All Increments)

```
main()
├── Parse CPL source
├── Parse IO specifications
├── Create ChronoconeComputationController
├── Load connector data → register MetricComputations
├── Load lookup data → register MetricComputations
├── Define aggregation matrices (ERO007)
├── Define spreading matrices (ERO008)                         ← Spreader matrices
├── Schedule metric computation order (topological)
├── Compute metrics
│   ├── GetMetricComputation("connector_A")
│   ├── Compute referenced metric "customers"                  ← Referenced metric
│   ├── ReduceDimension(customers, aggMatrix, ...)             ← Implicit aggregator
│   ├── ExpandDimension(source, spreadMatrix, ...)             ← Spreader applied
│   └── Calculate cell-wise result
├── Apply aggregators (if any)
└── Output results
```

## Code Generation Strategy

| Aspect | Strategy |
|--------|----------|
| Tensor expansion | `tensorops.ExpandDimension(src, matrix, srcDim, newDim)` |
| Scalar spreading | `tensorops.ExpandDimension(scalar, matrix, 0, 0)` |
| Matrix definition | Preamble: `var spreadingMatrix_X = &AggregationMatrix{...}` |
| Referenced metrics | Computed and registered as intermediate MetricComputations |
| Implicit aggregators | `ReduceDimension` applied to referenced metrics before spread |
| Fused operations | Combine adjacent matrices; fuse aggregate+spread pairs |
| Dynamic weights | Weights computed from metric data at runtime (Increment 3+) |
| Scheduled order | Topological sort ensures referenced metrics computed before spreadsheet |

## Decision

**Approach:** Reuse the `tensorops.ExpandDimension` from ERO007 Increment 1 for all spread operations. CIL tree walker emits dimension-expanding matrix multiplication calls with pre-built spreading matrices from analysis. Referenced metrics are computed first, including implicit aggregators.

**Rationale:**
The dimension-expanding matrix multiplication was designed and built in ERO007 specifically to serve both aggregators and spreaders. No new matrix infrastructure is needed. The main synthesis complexity for spreaders is handling referenced metrics — the walker must generate code that computes the referenced metric (and any implicit aggregator results) before the spread application.

## Implications

### Positive Implications

- Reuses existing matrix multiplication infrastructure with zero new tensor math
- Consistent code generation pattern with aggregators
- Referenced metric computation follows the same topological scheduling as metric dependencies
- Scalar spreading (`from []`) is a clean special case of dimension expansion

### Negative Implications / Trade-offs

- Dynamic spreading matrices (computed from metric data at runtime) add indirection
- Metric references in spreader expressions require the walker to know about metric scheduling order
- Fused spreader/aggregator operations in Increment 4 are complex to emit correctly

### Notes

**Related Documents:**

- [ERO008 — Spreaders Incremental Implementation](../../eros/ERO008_spreaders-incremental-implementation.md)
- [ERO008_spreaders-parsing-phase.md](../../go-parser-target/eros/ERO008_spreaders-parsing-phase.md)
- [ERO008_spreaders-analysis-phase.md](../../go-analysis/eros/ERO008_spreaders-analysis-phase.md)
- [Spreaders Design](../../spreaders.md)
- [Aggregators Design](../../aggregators.md)
- [ERO007_aggregators-synthesis-phase.md](../ERO007_aggregators-synthesis-phase.md) — Matrix multiplication infrastructure
