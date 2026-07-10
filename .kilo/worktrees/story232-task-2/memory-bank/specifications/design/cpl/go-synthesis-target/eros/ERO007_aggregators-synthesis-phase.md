# ERO007: Aggregators Synthesis Phase

## Title

Aggregators Synthesis Phase — Code Generation, Matrix Multiplication, and Runtime Integration

## Purpose/Goal

Define the synthesis phase implementation for aggregators, including tensor matrix multiplication infrastructure (Increment 1), CIL tree walking for `CILAggregateNode` nodes (Increment 2), aggregator integration into metric computation (Increment 3), and sequence optimization with fused operations (Increment 4), as defined in [ERO007 — Aggregators Incremental Implementation](../../eros/ERO007_aggregators-incremental-implementation.md).

## Decision Status

Draft

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic007_cpl-aggregators-implementation.md](../../../../../epics/Epic007_cpl-aggregators-implementation.md)

## Context

### Situation

The synthesis phase is responsible for walking CIL trees and generating executable Go code. For aggregators, this involves three layers:

1. **Matrix Multiplication Infrastructure (Increment 1):** Standalone tensor math that performs dimension-reducing and dimension-expanding matrix multiplications. This is pure Go code with no CPL dependency.
2. **CILAggregateNode Walking (Increment 2):** The CIL tree walker handles `CILAggregateNode` nodes, emitting Go code that applies the pre-built aggregation matrix to the source tensor.
3. **Metric Integration (Increment 3-4):** Aggregators are applied automatically during metric computation when dimension mismatches are detected by the analysis phase.

This ERO defines the synthesis for all four increments, starting with the foundational matrix multiplication in Increment 1.

### Constraints

- Matrix multiplication must handle both dimension reduction (for aggregators) and dimension expansion (for spreaders)
- Must work with the existing `*ChronoconeTensor[any]` interface
- Must handle arbitrary tensor ranks including rank zero (scalar tensors)
- CILAggregateNode walking must integrate with the existing CIL tree walker
- Synthesizer must not duplicate matrix construction logic — matrices come pre-built from analysis
- Types are initially represented as `float64`

### Requirements

- Implement tensor matrix multiplication: N-D → (N-1)-D for dimension reduction
- Implement tensor matrix multiplication: N-D → (N+1)-D for dimension expansion (reused by spreaders)
- Walk `CILAggregateNode` in the CIL tree walker and emit matrix multiplication calls
- Generate Go code that applies aggregator sequences as part of metric computation
- Support scalar reduction (`to []`) producing rank-zero tensors
- Support fused operations (combining adjacent aggregation matrices)
- Integrate aggregation into the scheduled metric computation order

### Assumptions

- The existing CIL tree walking infrastructure is available
- The `ChronoconeTensor` type and interface are available
- Aggregation matrices are pre-built by the analysis phase
- Concept value sets are known at compile time

## Data Structures and Algorithms by Increment

### Increment 1: Tensor Matrix Multiplication Infrastructure

**Overview**

This is a standalone infrastructure increment with no CPL parsing or analysis. It implements full matrix multiplication operations on tensors capable of changing tensor dimensionality. Both dimension-reducing and dimension-expanding operations are implemented, as both are needed for aggregators and spreaders.

**Tensor Matrix Multiplication — Dimension Reduction**

Reduces an N-dimensional tensor to (N-1) dimensions by contracting along one dimension using a weight matrix.

```go
// ReduceDimension applies a dimension-reducing matrix multiplication.
// Given an N-dimensional source tensor and a 2D weight matrix
// (rows = values of dimension being collapsed, columns = values of target dimension),
// produces an (N-1)-dimensional result tensor.
//
// Parameters:
//   src      - N-dimensional source tensor
//   matrix   - 2D aggregation matrix (size: collapsedDimValues × targetDimValues)
//   collDim  - Index of the dimension to collapse (0-based)
//   targetDim - Index to insert the target dimension (0-based, relative to remaining dims)
//
// For sum aggregation: matrix[m][n] = 1.0 for each source value m contributing to target n.
// For size: matrix[m][0] = 1.0 (counting each value once, producing scalar).
func ReduceDimension(
    src *ChronoconeTensor[float64],
    matrix *AggregationMatrix,
    collapsedDim int,
    targetDim int,
) (*ChronoconeTensor[float64], error) {

    srcShape := src.Shape()
    srcRank := len(srcShape)

    if collapsedDim < 0 || collapsedDim >= srcRank {
        return nil, fmt.Errorf("collapsedDim %d out of range for rank %d tensor", collapsedDim, srcRank)
    }

    collapsedSize := srcShape[collapsedDim]
    if collapsedSize != matrix.Rows {
        return nil, fmt.Errorf("collapsed dimension size %d does not match matrix rows %d",
            collapsedSize, matrix.Rows)
    }

    // Build result shape: remove collapsedDim, insert targetDim
    targetSize := matrix.Columns
    if targetSize == 0 {
        targetSize = 1 // Scalar target
    }

    // Result rank = N - 1
    resultShape := make([]int, 0, srcRank-1)
    for d := 0; d < srcRank; d++ {
        if d == collapsedDim {
            continue
        }
        resultShape = append(resultShape, srcShape[d])
    }

    // Result is effectively (N-1)-D
    // The collapsed dimension is contracted via the matrix.
    // Each cell in the result is:
    //   result[..., j, ...] = sum_over_k( src[..., k, ...] * matrix[k][j] )
    //
    // where k indexes over the collapsed dimension and j indexes over the target.

    resultSize := 1
    for _, s := range resultShape {
        resultSize *= s
    }

    resultData := make([]float64, resultSize)

    // For each cell in the result, compute the contracted sum
    // This is a nested loop over all (N-1) dimensions
    for cellIdx := 0; cellIdx < resultSize; cellIdx++ {
        // Decompose cellIdx into (N-1) coordinates
        coords := decomposeIndex(cellIdx, resultShape)

        // Sum over the collapsed dimension
        sum := 0.0
        for k := 0; k < collapsedSize; k++ {
            // Build source coordinates by inserting k at collapsedDim
            srcCoords := make([]int, srcRank)
            resultCoordIdx := 0
            for d := 0; d < srcRank; d++ {
                if d == collapsedDim {
                    srcCoords[d] = k
                } else {
                    srcCoords[d] = coords[resultCoordIdx]
                    resultCoordIdx++
                }
            }
            srcIdx := composeIndex(srcCoords, srcShape)

            // For each target column, accumulate the weight
            // (For sum, target is the preserved concept values;
            //  the result cellIdx corresponds to one target value j)
            // This simplified version assumes a 1:1 mapping from result coords
            // to matrix columns — for general case, further refinement needed.
            j := coords[targetDim]
            sum += src.Data[srcIdx] * matrix.Data[k][j]
        }

        resultData[cellIdx] = sum
    }

    return &ChronoconeTensor[float64]{
        Data:   resultData,
        Dims:   resultShape,
        Labels: nil, // Labels populated by caller based on concept values
    }, nil
}
```

**Tensor Matrix Multiplication — Dimension Expansion**

Expands an N-dimensional tensor to (N+1) dimensions using a weight matrix. Used by spreaders.

```go
// ExpandDimension applies a dimension-expanding matrix multiplication.
// Given an N-dimensional source tensor and a 2D weight matrix
// (rows = source dimension values, columns = new dimension values),
// produces an (N+1)-dimensional result tensor.
//
// Parameters:
//   src       - N-dimensional source tensor
//   matrix    - 2D spreading matrix (size: srcDimValues × newDimValues)
//   srcDim    - Index of the source dimension to expand from (0-based)
//   newDim    - Index to insert the new dimension (0-based)
//
// Each result cell: result[..., j, ...] = src[..., i, ...] * matrix[i][j]
// where i is the source dimension coordinate and j is the new dimension coordinate.
func ExpandDimension(
    src *ChronoconeTensor[float64],
    matrix *AggregationMatrix,
    srcDim int,
    newDim int,
) (*ChronoconeTensor[float64], error) {

    srcShape := src.Shape()
    srcRank := len(srcShape)

    if srcDim < 0 || srcDim >= srcRank {
        return nil, fmt.Errorf("srcDim %d out of range for rank %d tensor", srcDim, srcRank)
    }

    srcDimSize := srcShape[srcDim]
    if srcDimSize != matrix.Rows {
        return nil, fmt.Errorf("source dimension size %d does not match matrix rows %d",
            srcDimSize, matrix.Rows)
    }

    newDimSize := matrix.Columns

    // Build result shape: (N+1) dimensions with newDimSize inserted at newDim
    resultShape := make([]int, 0, srcRank+1)
    for d := 0; d < srcRank; d++ {
        if d == newDim {
            resultShape = append(resultShape, newDimSize)
        }
        resultShape = append(resultShape, srcShape[d])
    }
    if newDim >= srcRank {
        resultShape = append(resultShape, newDimSize)
    }

    // Result size
    resultSize := 1
    for _, s := range resultShape {
        resultSize *= s
    }

    resultData := make([]float64, resultSize)

    // For each cell in the (N+1)-D result
    for cellIdx := 0; cellIdx < resultSize; cellIdx++ {
        coords := decomposeIndex(cellIdx, resultShape)

        // Extract source coordinates by removing newDim
        srcCoords := make([]int, srcRank)
        resultCoordIdx := 0
        for d := 0; d < len(coords); d++ {
            if d == newDim {
                newCoord := coords[d] // j — column index into the matrix
                // Map to source: the source value at srcCoord * matrix[srcCoord][newCoord]
                i := srcCoords[srcDim] // source dimension coordinate
                resultData[cellIdx] = src.Data[composeIndex(srcCoords, srcShape)] *
                    matrix.Data[i][newCoord]
                break // Handled one cell already; this needs restructuring
            }
            srcCoords[resultCoordIdx] = coords[d]
            resultCoordIdx++
        }
    }

    // Note: The above loop structure is simplified. A full implementation would
    // iterate over all source cells and distribute values across the new dimension.

    return &ChronoconeTensor[float64]{
        Data: resultData,
        Dims: resultShape,
    }, nil
}
```

**Helper Functions (Increment 1)**

```go
// composeIndex converts N-dimensional coordinates to a flat index (row-major).
func composeIndex(coords []int, shape []int) int {
    index := 0
    stride := 1
    for i := len(shape) - 1; i >= 0; i-- {
        index += coords[i] * stride
        stride *= shape[i]
    }
    return index
}

// decomposeIndex converts a flat index to N-dimensional coordinates (row-major).
func decomposeIndex(index int, shape []int) []int {
    coords := make([]int, len(shape))
    for i := len(shape) - 1; i >= 0; i-- {
        coords[i] = index % shape[i]
        index /= shape[i]
    }
    return coords
}
```

**Testing (Increment 1)**

All tests use hand-constructed tensors with known values and known matrices:

- Dimension reduction: 1D → scalar, 2D → 1D, 3D → 2D with sum matrix
- Dimension expansion: scalar → 1D, 1D → 2D, 2D → 3D with distribution matrix
- Edge cases: rank zero tensors, single-value dimensions, identity matrices
- No CPL files are needed — pure Go tensor math unit tests

### Increment 2: Basic Aggregator Parsing and Validation

**Overview**

Increment 2 implements CIL tree walking for `CILAggregateNode`. The tree walker emits Go code that applies the pre-built aggregation matrix to the source tensor using the matrix multiplication infrastructure from Increment 1. Aggregators are applied manually in test code (not yet integrated into metric computation).

**CIL Tree Walker — CILAggregateNode Handler (Increment 2)**

```go
// WalkCILAggregateNode handles a CILAggregateNode in the CIL tree.
// It generates Go code that applies the aggregation matrix to the source tensor
// using the matrix multiplication infrastructure from Increment 1.
func (w *CILTreeWalker) WalkCILAggregateNode(node *CILAggregateNode, scope *WalkScope) error {
    // Walk the source CIL subtree first (this may be a connector, metric, or another aggregator)
    sourceVar := w.walkSubtree(node.SourceCIL, scope)

    // Generate a variable name for the aggregated result
    resultVar := scope.freshVar("aggregated")

    // Determine the collapsed dimension index
    collapsedDim := scope.getDimensionIndex(node.FromDimensions, node.ToDimension)

    // Emit the aggregation call
    scope.emitCodeIndented(scope.indentLevel, fmt.Sprintf(
        `%s, err := tensorops.ReduceDimension(%s, aggregationMatrix_%s, %d, %d)`,
        resultVar,
        sourceVar,
        sanitizeVarName(node.ToDimension),
        collapsedDim,
        scope.getTargetDimIndex(node),
    ))
    scope.emitCodeIndented(scope.indentLevel, "if err != nil {")
    scope.emitCodeIndented(scope.indentLevel+1, `return nil, fmt.Errorf("aggregation failed: %%w", err)`)
    scope.emitCodeIndented(scope.indentLevel, "}")

    // Register the result variable for use by subsequent operations
    scope.registerVariable(resultVar, node)
    scope.pushOperand(resultVar)

    return nil
}
```

**Aggregation Matrix Emission (Increment 2)**

The synthesis preamble generator emits Go code that materializes the pre-built aggregation matrices:

```go
// EmitAggregationMatrix generates Go code that defines an AggregationMatrix variable.
func (g *PreambleGenerator) EmitAggregationMatrix(
    buf io.Writer,
    name string,
    matrix *AggregationMatrix,
) error {
    tmpl := `
// Aggregation matrix for %s
var aggregationMatrix_%s = &tensorops.AggregationMatrix{
    Rows:    %d,
    Columns: %d,
    RowLabels: []string{%s},
    ColLabels: []string{%s},
    Data: [][]float64{
        %s
    },
}
`
    // Format row labels
    rowLabels := strings.Join(quoteStrings(matrix.RowLabels), ", ")

    // Format column labels
    colLabels := strings.Join(quoteStrings(matrix.ColLabels), ", ")

    // Format matrix data
    var dataLines []string
    for _, row := range matrix.Data {
        rowStrs := make([]string, len(row))
        for j, v := range row {
            rowStrs[j] = strconv.FormatFloat(v, 'f', -1, 64)
        }
        dataLines = append(dataLines, "{"+strings.Join(rowStrs, ", ")+"},")
    }
    dataStr := strings.Join(dataLines, "\n\t\t")

    rendered := fmt.Sprintf(tmpl, name, sanitizeVarName(name), matrix.Rows, matrix.Columns,
        rowLabels, colLabels, dataStr)

    buf.Write([]byte(rendered))
    return nil
}
```

**Generated Code Example — Manual Aggregation (Increment 2)**

```go
// Generated Go code for a manually applied aggregator:
//   from [Country, Product] to [Product] = sum(Country)

// Aggregation matrix for "Country"
var aggregationMatrix_Country = &tensorops.AggregationMatrix{
    Rows:    2,
    Columns: 2,
    RowLabels: []string{"USA", "Canada"},
    ColLabels: []string{"widget", "gadget"},
    Data: [][]float64{
        {1.0, 1.0},
        {1.0, 1.0},
    },
}

// Manual aggregation in test code:
func TestAggregatorSum_CountryProduct_to_Product(t *testing.T) {
    sourceTensor := loadTensor("country_product_metrics.parquet")

    // Apply aggregator
    result, err := tensorops.ReduceDimension(
        sourceTensor,
        aggregationMatrix_Country,
        0,  // collapsedDim: Country is at index 0
        0,  // targetDim: Product is at index 0 after removing Country
    )
    if err != nil {
        t.Fatalf("aggregation failed: %v", err)
    }

    // Verify result shape: 1D tensor with 2 values (one per product)
    expectedShape := []int{2}
    assert.Equal(t, expectedShape, result.Shape())
}
```

### Increment 3: Additional Aggregator Function CIL Node Types

**Overview**

Increment 3 handles the new CIL node types defined in the analysis phase. Each function has its own node type, and the CIL tree walker generates the corresponding Go code. Because aggregator functions fit into expressions, the walker emits a local variable computation followed by the variable reference as an expression operand.

**CILAggregateSumNode Walk (Increment 3)**

```go
func (w *CILTreeWalker) WalkCILAggregateSumNode(node *CILAggregateSumNode, scope *WalkScope) error {
    // Walk the source CIL subtree first
    sourceVar := w.walkSubtree(node.SourceCIL, scope)

    // Generate a local variable holding the aggregated result
    resultVar := scope.freshVar("sum_" + node.DimName)

    // Determine the collapsed dimension index in the source tensor
    collapsedDim := scope.getDimensionIndex(node.SourceCIL, node.DimName)

    scope.emitCodeIndented(scope.indentLevel, fmt.Sprintf(
        `%s, err := tensorops.ReduceDimension(%s, aggregationMatrix_sum_%s, %d, 0)`,
        resultVar, sourceVar, sanitizeVarName(node.DimName), collapsedDim,
    ))
    scope.emitCodeIndented(scope.indentLevel, "if err != nil {")
    scope.emitCodeIndented(scope.indentLevel+1, `return nil, fmt.Errorf("sum(%s) failed: %%w", err)`)
    scope.emitCodeIndented(scope.indentLevel, "}")

    // Push as expression operand
    scope.pushOperand(resultVar)
    return nil
}
```

**CILAggregateCountNode Walk (Increment 3)**

Same structure as `CILAggregateSumNode` — uses the same `ReduceDimension` with a sum matrix (count is semantically a sum with counting semantics).

**CILAggregateSizeNode Walk (Increment 3)**

```go
func (w *CILTreeWalker) WalkCILAggregateSizeNode(node *CILAggregateSizeNode, scope *WalkScope) error {
    // size(C) is a compile-time constant — emit the integer directly
    resultVar := scope.freshVar("size_" + node.DimName)
    scope.emitCodeIndented(scope.indentLevel, fmt.Sprintf(
        `%s := float64(%d)  // size(%s) = %d values`,
        resultVar, node.Size, node.DimName, node.Size,
    ))
    scope.pushOperand(resultVar)
    return nil
}
```

**CILAggregateInverseSumNode / CILAggregateInverseCountNode (Increment 3)**

Same as `ReduceDimension` but with the inverse matrix (each entry = 1/N).

**Synthesis Summary for Increment 3**

| CIL Node Type | Generated Code |
|---------------|---------------|
| `CILAggregateSumNode` | `tensorops.ReduceDimension(source, sumMatrix, dim, target)` |
| `CILAggregateCountNode` | `tensorops.ReduceDimension(source, countMatrix, dim, target)` |
| `CILAggregateSizeNode` | `float64(len(conceptValues))` (compile-time constant) |
| `CILAggregateInverseSumNode` | `tensorops.ReduceDimension(source, invSumMatrix, dim, target)` |
| `CILAggregateInverseCountNode` | `tensorops.ReduceDimension(source, invCountMatrix, dim, target)` |

### Increment 4: Aggregator Selection and Integration

**Overview**

Increment 3 integrates aggregators into metric computation. The CIL tree walker generates code that applies aggregator sequences as part of the metric's calculation pipeline. Aggregators are automatically inserted by the analysis phase when dimension mismatches are detected.

**Metric Computation with Aggregators (Increment 3)**

```go
// Generated Go code for a metric that uses an aggregated input:
// metric: regional_profit
//   inputs: metric country_revenue, metric country_cost
//   (country_revenue has dimensions [Country, Product],
//    output metric requires [Product])

func computeRegionalProfit(controller *ChronoconeComputationController) (*ChronoconeTensor[float64], error) {
    // Get input metrics
    countryRevenue := controller.GetMetricComputation("country_revenue").Tensor().(*ChronoconeTensor[float64])
    countryCost := controller.GetMetricComputation("country_cost").Tensor().(*ChronoconeTensor[float64])

    // Apply aggregator: from [Country, Product] to [Product] = sum(Country)
    revenueAggregated, err := tensorops.ReduceDimension(
        countryRevenue,
        aggregationMatrix_Country,  // Pre-built matrix for collapsing Country
        0,  // collapsedDim: Country
        0,  // targetDim: Product
    )
    if err != nil {
        return nil, fmt.Errorf("failed to aggregate country_revenue: %w", err)
    }

    costAggregated, err := tensorops.ReduceDimension(
        countryCost,
        aggregationMatrix_Country,
        0,  // collapsedDim: Country
        0,  // targetDim: Product
    )
    if err != nil {
        return nil, fmt.Errorf("failed to aggregate country_cost: %w", err)
    }

    // Both inputs now have dimensions [Product] — compute cell-wise
    result := chronocone_compute.NewChronoconeTensor[float64](revenueAggregated.Shape())
    for i := 0; i < len(revenueAggregated.Data); i++ {
        result.Data[i] = revenueAggregated.Data[i] - costAggregated.Data[i]
    }

    return result, nil
}
```

**Scheduled Computation Order (Increment 3)**

Aggregated inputs must be computed before the metrics that depend on them. The generated preamble schedules metric computation in topological order:

```go
// In the generated main() function:
func main() {
    // ... load data ...

    // Schedule computation order (topological sort of metric DAG)
    // Aggregated versions of metrics are computed as intermediate tensors.
    metrics := []string{
        "country_revenue",    // Connector-loaded
        "country_cost",       // Connector-loaded
        // "aggregated:country_revenue" — computed implicitly by regional_profit
        // "aggregated:country_cost"    — computed implicitly by regional_profit
        "regional_profit",    // References aggregated inputs
    }

    // ... compute in order ...
}
```

### Increment 5: Aggregator Sequence Optimization

**Overview**

Increment 4 implements optimizations: fusing adjacent aggregation matrices into single multiplications, and handling multi-step aggregator sequences.

**Matrix Fusion (Increment 4)**

When two aggregators are applied in sequence (e.g., collapse Region then Country), their matrices can be fused:

```go
// EmitFusedAggregatorSequence generates Go code that applies multiple
// aggregators as a single fused matrix multiplication where possible.
func (w *CILTreeWalker) EmitFusedAggregatorSequence(
    nodes []*CILAggregateNode,
    sourceVar string,
    scope *WalkScope,
) (string, error) {

    // Fuse adjacent matrices
    fusedMatrices := []*AggregationMatrix{}
    currentMatrix := nodes[0].AggregationMatrix

    for i := 1; i < len(nodes); i++ {
        nextMatrix := nodes[i].AggregationMatrix
        if currentMatrix.Columns == nextMatrix.Rows {
            // Compatible — fuse
            currentMatrix = fuseAggregationMatrices(currentMatrix, nextMatrix)
        } else {
            // Not compatible — emit current fused matrix, start new
            fusedMatrices = append(fusedMatrices, currentMatrix)
            currentMatrix = nextMatrix
        }
    }
    fusedMatrices = append(fusedMatrices, currentMatrix)

    // Emit code for each fused matrix multiplication
    resultVar := sourceVar
    for i, matrix := range fusedMatrices {
        nextVar := scope.freshVar(fmt.Sprintf("fused_%d", i))
        scope.emitCodeIndented(scope.indentLevel, fmt.Sprintf(
            `%s, err := tensorops.ReduceDimension(%s, fusedMatrix_%d_%d, /* collapsedDim */, /* targetDim */)`,
            nextVar, resultVar, i, len(fusedMatrices),
        ))
        scope.emitCodeIndented(scope.indentLevel, "if err != nil {")
        scope.emitCodeIndented(scope.indentLevel+1, "return nil, err")
        scope.emitCodeIndented(scope.indentLevel, "}")
        resultVar = nextVar
    }

    scope.pushOperand(resultVar)
    return resultVar, nil
}
```

**Edge Cases (Increment 4)**

```go
// handleScalarAggregation generates code for to [] aggregation (scalar result).
func (w *CILTreeWalker) handleScalarAggregation(node *CILAggregateNode, sourceVar string, scope *WalkScope) {
    resultVar := scope.freshVar("scalar")

    // Scalar reduction produces a rank-zero tensor (single value)
    scope.emitCodeIndented(scope.indentLevel, fmt.Sprintf(
        `%s, err := tensorops.ReduceToScalar(%s, aggregationMatrix_%s, %d)`,
        resultVar, sourceVar, sanitizeVarName(node.ToDimension), collapsedDim,
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
├── Define aggregation matrices (Increment 2+)               ← Aggregator matrices
├── Schedule metric computation order (topological)
├── Compute metrics
│   ├── GetMetricComputation("connector_A")
│   ├── ReduceDimension(input, aggregationMatrix_X, ...)     ← Aggregator applied
│   ├── ReduceDimension(intermediate, aggregationMatrix_Y, ...) ← Sequence
│   └── Calculate cell-wise result
├── Apply spreaders (if any, ERO008)
└── Output results
```

## Code Generation Strategy

| Aspect | Strategy |
|--------|----------|
| Matrix multiplication | `tensorops.ReduceDimension(src, matrix, collDim, targetDim)` |
| Scalar reduction | `tensorops.ReduceToScalar(src, matrix, collDim)` |
| Matrix definition | Preamble: `var aggregationMatrix_X = &AggregationMatrix{...}` |
| Fused operations | Combine adjacent matrices before emitting code |
| Scheduled order | Topological sort ensures inputs aggregated before consumers |
| Error handling | Matrix dimension mismatch is a runtime error |
| Type representation | `float64` for all numeric values |

## Decision

**Approach:** Standalone tensor operations package (`tensorops`) providing dimension-changing matrix multiplication, with CIL tree walker emitting calls to these operations.

**Rationale:**
Separating tensor math from CPL code generation provides clean isolation and reusability. The `tensorops` package can be independently tested with hand-constructed tensors (Increment 1) and can be used by both aggregators and spreaders. The CIL tree walker's role is to emit the correct matrix multiplication calls with the pre-built matrices from analysis — no matrix construction logic in the walker.

## Implications

### Positive Implications

- Reusable tensor operations package for both aggregators and spreaders
- Independently testable matrix multiplication infrastructure
- Clean separation between tensor math and code generation
- Matrix construction is centralized in analysis; synthesis only applies matrices

### Negative Implications / Trade-offs

- Matrix multiplication is computationally expensive for large tensors
- The `ReduceDimension` algorithm iterates over all result cells and all collapsed-dimension values (O(resultSize × collapsedSize))
- Fused matrix multiplication produces intermediate matrices that must be materialized
- Scalar reduction (to `[]`) is a special case that requires separate handling

### Future Considerations

- Parallel tensor operations using goroutines for large tensors
- Sparse matrix representations for large concept value sets
- Caching repeated aggregation results when the same aggregator is applied to the same input multiple times
- GPU acceleration for matrix multiplication on very large tensors

### Notes

**Related Documents:**

- [ERO007 — Aggregators Incremental Implementation](../../eros/ERO007_aggregators-incremental-implementation.md) — Parent ERO
- [ERO007_aggregators-parsing-phase.md](../../go-parser-target/eros/ERO007_aggregators-parsing-phase.md) — Parsing phase
- [ERO007_aggregators-analysis-phase.md](../../go-analysis/eros/ERO007_aggregators-analysis-phase.md) — Analysis phase
- [Aggregators Design](../../aggregators.md) — Full aggregators design
- [Spreaders Design](../../spreaders.md) — Spreader design (reuses matrix multiplication infrastructure)
- [Metrics Design](../../metrics-design.md) — Metric design (aggregators applied in metric computation)
