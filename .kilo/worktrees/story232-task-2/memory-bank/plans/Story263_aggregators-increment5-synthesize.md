# Plan: Story263 — Aggregators Increment 5 (Synthesize)

## Purpose

This plan provides the technical writer with sufficient detail to create Story263: **aggregators-increment5-synthesize** — implement fused matrix operations for multi-step aggregator sequences, handle scalar reduction edge cases, multi-step sequence optimization, and end-to-end sequence optimization examples (Increment 5 of Epic007).

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `aggregators-increment5-synthesize`                                                           |
| description  | Implement fused matrix operations, scalar reduction edge cases, multi-step sequence optimization, sequence optimization examples |

## Source EROs

- [ERO007: Aggregators Incremental Implementation](../specifications/design/cpl/eros/ERO007_aggregators-incremental-implementation.md) — Increment 5 sketch
- [ERO007: Aggregators Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO007_aggregators-synthesis-phase.md) — Matrix fusion, scalar reduction, `EmitFusedAggregatorSequence`, edge cases, `handleScalarAggregation`

## Dependencies

- Story262 (aggregators-increment5-analyze) — needs multi-step sequences from `enumerateAggregatorSequences` + `selectOptimalSequence`
- Story261 (aggregators-increment4-synthesize) — needs metric code generation with aggregate nodes
- Story259 (aggregators-increment3-synthesize) — needs walk methods for all typed CIL nodes
- Story254 (aggregators-increment1-matrix-mult) — `tensorops.ReduceDimension`

## Files to Create or Modify

| File                                                    | Action  | Purpose                                                                          |
| ------------------------------------------------------- | ------- | -------------------------------------------------------------------------------- |
| `src/main/go/tensorops/tensorops.go`                    | Modify  | Add `fuseAggregationMatrices` function for combining adjacent matrices           |
| `src/main/go/tensorops/tensorops_test.go`               | Modify  | Add tests for matrix fusion: 2×3 × 3×2 → 2×2, identity fusion, dimension mismatch error |
| `src/main/go/cil/cil_transformer.go`                    | Modify  | Add `EmitFusedAggregatorSequence` for optimized multi-step aggregation; add scalar reduction handling |
| `src/main/go/cil/cil_transformer_metric.go`             | Modify  | Handle chained aggregate nodes in metric inputs (multi-step sequences)           |
| `src/main/go/cil/cil_transformer_aggregator_test.go`    | Modify  | Extend tests: fused sequence emission, scalar reduction edge case, multi-step chain |
| `src/main/go/synthesis/synthesize_test.go`              | Modify  | Extend end-to-end synthesis tests with multi-step aggregator sequences           |
| Examples:                                               |         |                                                                                  |
| `examples/with_data/aggregators_optimization/`          | Create  | CPL + IO spec + parquet + expected: multi-step aggregation sequences exercising the selection algorithm |

### Matrix Fusion

When multiple aggregators are applied in sequence (e.g., collapse `Region` then `Country`), adjacent aggregation matrices can be combined into a single multiplication:

```go
// fuseAggregationMatrices combines two adjacent aggregation matrices A (M×N) × B (N×P) → C (M×P)
func fuseAggregationMatrices(a, b *AggregationMatrix) *AggregationMatrix {
    m, n, p := a.Rows, a.Columns, b.Columns
    // Validate: a.Columns must == b.Rows
    if n != b.Rows {
        return nil // Cannot fuse — incompatible dimensions
    }
    
    data := make([][]float64, m)
    for i := 0; i < m; i++ {
        data[i] = make([]float64, p)
        for j := 0; j < p; j++ {
            sum := 0.0
            for k := 0; k < n; k++ {
                sum += a.Data[i][k] * b.Data[k][j]
            }
            data[i][j] = sum
        }
    }
    
    return &AggregationMatrix{
        Rows: m, Columns: p,
        Data: data,
        RowLabels: a.RowLabels,
        ColLabels: b.ColLabels,
    }
}
```

### EmitFusedAggregatorSequence

For multi-step aggregator sequences, the CIL tree walker fuses adjacent compatible matrices to reduce the number of tensor operations:

```go
func (w *CILTreeWalker) EmitFusedAggregatorSequence(
    aggregateNodes []CILNode,  // Sequence of typed aggregate nodes
    sourceVar string,
    scope *WalkScope,
) (string, error) {
    // Collect matrices from the sequence
    matrices := make([]*AggregationMatrix, 0, len(aggregateNodes))
    for _, node := range aggregateNodes {
        switch n := node.(type) {
        case *CILAggregateSumNode:
            matrices = append(matrices, n.Matrix)
        case *CILAggregateCountNode:
            matrices = append(matrices, n.Matrix)
        case *CILAggregateInverseSumNode:
            matrices = append(matrices, n.Matrix)
        case *CILAggregateInverseCountNode:
            matrices = append(matrices, n.Matrix)
        // CILAggregateSizeNode doesn't have a matrix
        }
    }
    
    // Fuse adjacent compatible matrices
    fused := fuseMatrixChain(matrices)
    
    // Emit code for each fused step
    resultVar := sourceVar
    for i, matrix := range fused {
        nextVar := scope.freshVar(fmt.Sprintf("fused_%d", i))
        scope.emitCodeIndented(scope.indentLevel, fmt.Sprintf(
            `%s, err := tensorops.ReduceDimension(%s, fusedMatrix_%s_%d, collapsedDim, targetDim)`,
            nextVar, resultVar, scope.currentFunction(), i,
        ))
        scope.emitCodeIndented(scope.indentLevel, "if err != nil {")
        scope.emitCodeIndented(scope.indentLevel+1, "return nil, err")
        scope.emitCodeIndented(scope.indentLevel, "}")
        
        // Emit fused matrix definition in preamble
        scope.registerMatrixVar(fmt.Sprintf("fusedMatrix_%s_%d", scope.currentFunction(), i), matrix)
        resultVar = nextVar
    }
    
    scope.pushOperand(resultVar)
    return resultVar, nil
}
```

### Scalar Reduction Edge Case

When `to` is `[]` (scalar reduction), the output tensor is rank-zero (single value). The `ReduceDimension` function must handle this:

- Input: 1D tensor [10, 20] (shape [2]), matrix [1.0, 1.0] (2×1)
- Result: 0D tensor [30] (shape [] — empty, single element)

The `DecomposeIndex`/`ComposeIndex` helpers must handle rank-zero tensors (shape = `[]int{}`, size = 1).

The walker handles scalar reduction by computing the collapsed dimension index (always 0 for scalar reduction since there's exactly one from-concept) and target dimension index (0, since the remaining tensor after removing the collapsed dim has no dimensions left).

### Preamble Matrix Emission for Fused Matrices

The synthesis preamble must emit both:
1. Original per-aggregator matrices (for single-step applications)
2. Fused matrices for multi-step sequences (named `fusedMatrix_<metric>_<N>`)

The `CILCodeGenerator` already supports matrix emission — fusion adds new fused matrices that must also be emitted.

### with_data Example: aggregators_optimization

- **Scenario**: input metric with dimensions [Country, Region, Product], output metric with [Product]
- **Aggregators**: `from [Country, Region, Product] to [Region, Product] = sum(Country)` and `from [Country, Region, Product] to [Country, Product] = sum(Region)` — both are needed
- **Expected behavior**: framework enumerates both orderings, selects optimal (less info loss), applies fused or sequential matrix multiplications
- **Test data**: known values for Country={USA, Canada} (2 values), Region={North, South, East, West} (4 values), Product={widget, gadget} (2 values)
- **Expected output**: aggregated to [Product] = 2 values

The framework should prefer collapsing Country (2 values) before Region (4 values) for least information loss.

## Task Decomposition (for technical writer)

### Task 1: Implement matrix fusion — Not Started
   a. Implement `fuseAggregationMatrices(a, b *AggregationMatrix) *AggregationMatrix` in `tensorops.go` — Not Started
   b. Implement `fuseMatrixChain(matrices []*AggregationMatrix) []*AggregationMatrix` — greedily fuse adjacent compatible matrices — Not Started
   c. Write tests: 2×3 × 3×2 → 2×2 with correct values, chain of 3 matrices → 2 fused (if second pair incompatible) or 1 fused (if all compatible), identity matrix fusion (no-op), dimension mismatch returns nil — Not Started
   d. Verify tests pass — Not Started

### Task 2: Implement EmitFusedAggregatorSequence in CIL walker — Not Started
   a. Implement `EmitFusedAggregatorSequence` in `cil_transformer.go` — extract matrices from typed nodes, fuse, emit ReduceDimension calls for each fused step — Not Started
   b. Integrate fused emission into the walker's handling of chained aggregate nodes in metric inputs — Not Started
   c. Write tests: 2-step sequence emits fused `ReduceDimension` call with fused matrix, 2-step incompatible sequence emits two separate `ReduceDimension` calls — Not Started
   d. Write tests: verify fused matrix variables are emitted in preamble — Not Started
   e. Verify tests pass — Not Started

### Task 3: Handle scalar reduction edge case — Not Started [Small]
   a. Verify `tensorops.ReduceDimension` handles rank-zero result tensors (shape [] after scalar reduction) — Not Started
   b. Verify `composeIndex`/`decomposeIndex` handle empty shape — Not Started
   c. Write test: 1D tensor [10, 20] → scalar reduction → 0D tensor [30] — Not Started
   d. Write walker test: scalar reduction aggregate node emits correct `ReduceDimension` call targeting rank-zero — Not Started
   e. Verify tests pass — Not Started

### Task 4: Create with_data sequence optimization example — Not Started
   a. Create `examples/with_data/aggregators_optimization/` directory — Not Started
   b. Write CPL file: concepts (Country=2, Region=4, Product=2), connector ([Country, Region, Product]), two aggregators, metric with [Product] output — Not Started
   c. Write IO spec: input parquet connector data, output metric spec — Not Started
   d. Generate parquet test data — Not Started
   e. Create `.expected.xlsx` with ground-truth aggregated values — Not Started
   f. Verify framework selects optimal order (Country before Region) — Not Started
   g. Verify example test passes — Not Started

### Task 5: End-to-end multi-step synthesis test — Not Started
   a. Write synthesis integration test: `AnalysisDefinitions` with metric requiring 2-step aggregation → synthesize → verify generated code uses fused matrix or sequential ReduceDimension as appropriate → verify correct numerical results — Not Started
   b. Verify test passes — Not Started

## Execution Order

```
Task 1 → Task 2 → [Task 3 ‖ Task 4 ‖ Task 5]
```

- Task 1 must complete first (matrix fusion is used by Task 2)
- Task 2 depends on Task 1 (fused emission uses matrix fusion)
- Tasks 3, 4, 5 are parallel after Task 2: Task 3 is tensorops edge case, Task 4 is examples, Task 5 is integration tests

## Parallel Execution

- **Group A** (Tasks 3, 4, 5): Task 3 modifies `tensorops_test.go`; Task 4 creates files under `examples/`; Task 5 modifies `synthesize_test.go`. Fully disjoint files and directories.

## Reintegration

**Group A (Tasks 3, 4, 5):**
- Merge order: any order (fully disjoint file paths)
- Integration tests: `go test ./src/main/go/tensorops/... ./src/main/go/cil/... ./src/main/go/synthesis/... -count=1 -timeout 60s` — verify all pass together
- Watch for conflicts in: none expected

## Acceptance Criteria Summary

1. `fuseAggregationMatrices` correctly multiplies two compatible matrices and produces correct fused matrix
2. `fuseMatrixChain` greedily fuses adjacent compatible matrices in a chain
3. Incompatible matrices (dimensions don't match) are left unfused — emitted as separate ReduceDimension calls
4. `EmitFusedAggregatorSequence` generates optimized code for multi-step sequences
5. Fused matrices are emitted in the synthesis preamble with unique variable names
6. Scalar reduction (`to []`) produces correct rank-zero tensor results
7. `composeIndex`/`decomposeIndex` handle empty shapes (rank-zero tensors)
8. `with_data/aggregators_optimization/` example: multi-step sequence, optimal order selected, correct results
9. Multi-step aggregator sequences produce identical numerical results whether fused or sequential
10. All tests pass across `tensorops`, `cil`, and `synthesis` packages

## Intent Summary for Technical Writer

This story delivers the final optimization layer for aggregator synthesis. When multi-step aggregator sequences are selected (e.g., collapse Region then Country to reach Product), adjacent aggregation matrices are fused into single matrix multiplications where possible, reducing the number of tensor operations and intermediate tensors. The fusion is mathematically equivalent to sequential application — it produces identical numerical results while being more efficient. Scalar reduction (`to []`) is handled as a special edge case producing rank-zero tensors. The combined effect of Increments 4 and 5 is a fully automatic aggregator system: users declare aggregator transformations, and the framework detects dimension mismatches in the metric DAG, enumerates valid aggregator sequences, selects the optimal one using the three-tier algorithm, inserts aggregate nodes into CIL trees, and generates optimized code with fused matrix multiplications.
