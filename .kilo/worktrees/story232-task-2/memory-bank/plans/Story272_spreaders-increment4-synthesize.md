# Plan: Story272 — Spreaders Increment 4 (Synthesize)

## Purpose

This plan provides the technical writer with sufficient detail to create Story272: **spreaders-increment4-synthesize** — implement synthesis (code generation) for mixed aggregator/spreader interleaved sequences with fused operations, scalar spreading edge cases, multi-step interleaved sequence optimization, and optimization examples (Increment 4 of Epic008).

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `spreaders-increment4-synthesize`                                                            |
| description  | Implement fused spreader/aggregator operations, scalar spreading edge cases, multi-step interleaved sequence optimization, optimization examples |

## Source EROs

- [ERO008: Spreaders Incremental Implementation](../specifications/design/cpl/eros/ERO008_spreaders-incremental-implementation.md) — Increment 4 sketch
- [ERO008: Spreaders Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO008_spreaders-synthesis-phase.md) — Increment 4 data structures, fused operations, generated code examples

## Dependencies

- Story271 (spreaders-increment4-analyze) — analysis must produce interleaved `CILSpreadNode`/`CILAggregateNode` chains selected by the optimization algorithm
- Story270 (spreaders-increment3-synthesize) — metric reference walking and dynamic matrix construction infrastructure must exist
- Story268 (spreaders-increment2-synthesize) — spreader chain walking must exist
- Story261 (aggregators-increment4-synthesize) — aggregator chain walking and metric function generation infrastructure must exist; fused operations combine both aggregator and spreader patterns
- Story254 (aggregators-increment1-matrix-mult) — `tensorops.ReduceDimension` and `tensorops.ExpandDimension` are the synthesis primitives

## Files to Create or Modify

| File                                                    | Action | Purpose                                                                          |
| ------------------------------------------------------- | ------ | -------------------------------------------------------------------------------- |
| `src/main/go/synthesis/fused_sequence_emitter.go`       | Create | `EmitFusedSpreaderSequence`, matrix fusion, interleaved op emission               |
| `src/main/go/synthesis/cil_aggregate_walker.go`         | Modify | Update aggregator walker to integrate with fused sequence emission                |
| `src/main/go/synthesis/cil_spread_walker.go`            | Modify | Update spreader walker to integrate with fused sequence emission                  |
| `src/main/go/synthesis/fused_sequence_emitter_test.go`  | Create | Unit tests: fused matrix construction, interleaved op emission, optimization verification |
| `examples/spreaders/optimization/` (in appropriate with_data dir) | Create | End-to-end example with mixed aggregator/spreader optimization              |

### Fused Operations Concept

When the analysis produces interleaved spreader/aggregator sequences, the synthesizer can **fuse** adjacent compatible operations:

1. **Adjacent spreading matrices**: If spreader A and spreader B both expand dimensions and operate on compatible dimensions, their matrices can be combined into one larger matrix multiplication
2. **Adjacent aggregation matrices**: Same fusion as spreading — combine two aggregation matrices into one
3. **Aggregate-then-spread on different dimensions**: If aggregation collapses dim X and spreader introduces dim Y (where X ≠ Y), the operations can potentially be fused into a single matrix multiplication

```go
// Without fusion (Increment 2/3 behavior):
result1 := ExpandDimension(source, spreadMatrix1, dim1, pos1)
result2 := ReduceDimension(result1, aggMatrix, dim2, pos2)
result3 := ExpandDimension(result2, spreadMatrix2, dim3, pos3)

// With fusion (Increment 4):
fusedMatrix := fuseAggregationMatrices(spreadMatrix1, aggMatrix, spreadMatrix2)
result := ApplyFusedTransformation(source, fusedMatrix)
```

### Fused Sequence Emission

```go
func (w *CILTreeWalker) EmitFusedSpreaderSequence(
    ops []*DimensionalOp,
    sourceVar string,
    scope *WalkScope,
) (string, error) {

    // Phase 1: Identify fusable groups
    fusedGroups := groupFusableOperations(ops)

    // Phase 2: Build fused matrices for each group
    var fusedMatrices []*tensorops.AggregationMatrix
    var fusedOpTypes []string  // "expand" or "reduce" or "transform"
    currentGroup := fusedGroups[0]

    for i := 1; i < len(fusedGroups); i++ {
        nextGroup := fusedGroups[i]

        // Try to fuse current group with next group
        if canFuse(currentGroup, nextGroup) {
            currentGroup = fuseGroups(currentGroup, nextGroup)
        } else {
            fusedMatrices = append(fusedMatrices, currentGroup.Matrix)
            fusedOpTypes = append(fusedOpTypes, currentGroup.OpKind)
            currentGroup = nextGroup
        }
    }
    fusedMatrices = append(fusedMatrices, currentGroup.Matrix)
    fusedOpTypes = append(fusedOpTypes, currentGroup.OpKind)

    // Phase 3: Emit code for each fused operation
    resultVar := sourceVar
    for i, matrix := range fusedMatrices {
        nextVar := scope.freshVar(fmt.Sprintf("fused_%d", i))
        switch fusedOpTypes[i] {
        case "expand":
            scope.emitExpandDimensionCall(nextVar, resultVar, matrix)
        case "reduce":
            scope.emitReduceDimensionCall(nextVar, resultVar, matrix)
        case "transform":
            scope.emitTransformCall(nextVar, resultVar, matrix)
        }
        resultVar = nextVar
    }

    scope.pushOperand(resultVar)
    return resultVar, nil
}
```

### Fusable Conditions

Two operations can be fused when:
1. They operate on **different dimensions** (no data dependency conflict)
2. The intermediate tensor shape is compatible (output shape of op1 matches input expectation of op2)
3. The combined operation produces the same result as sequential execution

```go
func canFuse(op1, op2 *FusedGroup) bool {
    // Aggregate then spread on different dimensions is fusable
    if op1.OpKind == "reduce" && op2.OpKind == "expand" {
        return op1.CollapsedDim != op2.NewDimension
    }
    // Spread then aggregate on different dimensions is fusable
    if op1.OpKind == "expand" && op2.OpKind == "reduce" {
        return op1.NewDimension != op2.CollapsedDim
    }
    // Two spreads on different dimensions is fusable
    if op1.OpKind == "expand" && op2.OpKind == "expand" {
        return op1.NewDimension != op2.NewDimension
    }
    // Two aggregates on different dimensions is NOT directly fusable (matrices have different collapsed dims)
    // Each aggregate's reduction matrix is specific to the collapsed dim
    return false
}
```

### Scalar Spreading Edge Cases (Increment 4)

The `from []` (scalar) case has special synthesis considerations:

```go
func (w *CILTreeWalker) handleScalarSpreadInSequence(
    node *CILSpreadNode,
    sourceVar string,
    scope *WalkScope,
) {
    if node.FromEmpty {
        resultVar := scope.freshVar("scalar_spread")
        // Source is rank-zero (scalar), result is rank-one
        scope.emitCodeIndented(scope.indentLevel, fmt.Sprintf(
            `%s, err := tensorops.ExpandDimension(%s, spreadingMatrix_%s, 0, 0)`,
            resultVar, sourceVar, sanitizeVarName(node.NewDimension),
        ))
        scope.registerVariable(resultVar, node)
        scope.pushOperand(resultVar)
    }
}
```

### Generated Code Example — Interleaved Sequence

For a metric with input `[Country, Product]` and output `[Region]` (hypothetical):
```
Sequence: aggregate Country → spread to Region (fused)

Generated:
// Aggregate Country: from [Country, Product] to [Product]
// Spread to Region: from [Product] to [Region]
// Since Country ≠ Region (different dims), fuse into single transform:
result, err := tensorops.ApplyFusedTransform(sourceTensor, fusedMatrix_CountryProduct_Region)
```

### Optimization Examples

Example: Input `[Country, Product]`, output `[Region]`.
- **Without optimization**: `ReduceDimension(source, aggMatrix_Country, 0, 0)` → `ExpandDimension(intermediate, spreadMatrix_Region, 0, 0)` — two tensor ops, one intermediate allocation
- **With optimization**: Single `ApplyFusedTransform(source, fusedMatrix)` — one tensor op, zero intermediate allocations

The fused matrix is computed at analysis time (in the preamble generator) as the matrix product of the two individual matrices.

### Preamble Generation for Fused Matrices

```go
func (g *PreambleGenerator) EmitFusedMatrix(
    buf io.Writer,
    name string,
    matrices []*AggregationMatrix,
    opTypes []string,
) error {
    // Compute the fused matrix: multiply matrices in sequence
    fused := fuseAggregationMatrices(matrices, opTypes)
    // Emit as Go var declaration
    return g.EmitAggregationMatrix(buf, name, fused)
}
```

### End-to-End Examples: with_data/spreaders_optimization/

Directory structure:
```
examples/spreaders/optimization/
├── spreaders_optimization.cpl          # CPL with multiple spreaders + aggregators
├── spreaders_optimization.io.yaml      # IO spec
├── spreaders_optimization_input.xlsx   # Input data
├── spreaders_optimization.expected.json # Expected output
```

Example scenarios:
1. **Interleaved aggregate-spread**: Input dims partially overlap output dims → mixed sequence generated, optimal interleaving selected
2. **Ambiguous spreader selection**: Two spreaders match same pattern → most specific selected, verified by output
3. **Fused operations**: Sequence where adjacent ops are fusable → verify fused output matches non-fused output
4. **No fusion possible**: Dimensions overlap such that ops cannot be fused → sequential ops emitted correctly

### Test Cases

1. **Simple fusion**: Two adjacent spreads on different dims → fused into single matrix operation
2. **Aggregate-spread fusion**: Aggregate on dim X, spread on dim Y (X ≠ Y) → fused when possible
3. **Non-fusable sequence**: Aggregate on dim X, spread on dim X (same dim) → cannot fuse, sequential ops emitted
4. **Scalar spreading in sequence**: `from []` spreader in interleaved sequence → correct scalar handling
5. **Fused output correctness**: Fused operation produces identical result to sequential operations
6. **Performance verification**: Fused path produces fewer intermediate tensor allocations (verified by shape trace)
7. **Preamble completeness**: All fused matrices emitted in preamble, named correctly
8. **Error propagation**: Fused operation errors are wrapped correctly
9. **Regression**: Non-fused Increment 2/3 sequences still work

## Task Decomposition (for technical writer)

### Task 1: Implement fusion grouping and fusable condition analysis — Not Started
   a. Create `src/main/go/synthesis/fused_sequence_emitter.go` — define `FusedGroup` struct — Not Started
   b. Implement `groupFusableOperations` — scan sequence, identify adjacent fusable groups — Not Started
   c. Implement `canFuse` — check dimension independence and shape compatibility — Not Started
   d. Write unit tests: verify adjacent spreads on different dims are fusable, aggregate+spread on different dims fusable, aggregate+spread on same dim not fusable, two aggregates not fusable — Not Started
   e. Verify grouping tests pass — Not Started

### Task 2: Implement matrix fusion and fused code emission — Not Started
   a. Implement `fuseGroups` — combine matrices from fusable operations into single AggregationMatrix — Not Started
   b. Implement `EmitFusedSpreaderSequence` in CIL tree walker — walks ops, fuses groups, emits fused operations — Not Started
   c. Emit `ExpandDimension` for fused expand groups, `ReduceDimension` for fused reduce groups — Not Started
   d. Write unit tests: verify fused matrix dimensions are correct, verify emitted code for fused sequence, verify fused output matches sequential output — Not Started
   e. Verify fusion emission tests pass — Not Started

### Task 3: Implement scalar spreading and edge case handling — Not Started
   a. Update `handleScalarSpreadInSequence` — handle `from []` spreaders within interleaved sequences — Not Started
   b. Handle edge case: scalar input to entire fused sequence (rank-zero to multi-dim output) — Not Started
   c. Handle edge case: fused sequence ending in scalar (multi-dim to rank-zero) — Not Started
   d. Write unit tests: scalar→1D→2D fused sequence, 2D→scalar fused sequence, scalar→scalar no-op — Not Started
   e. Verify edge case tests pass — Not Started

### Task 4: Integrate into metric computation pipeline and create examples — Not Started
   a. Wire fused sequence emission into the metric computation code generator — Not Started
   b. Ensure fused matrix preamble variables are emitted — Not Started
   c. Write integration tests: full generated program with fused sequence produces correct output — Not Started
   d. Create `with_data/spreaders_optimization/` examples: interleaved aggregate-spread, ambiguous spreader selection, fused vs non-fused comparison — Not Started
   e. Verify all end-to-end examples produce output matching `.expected.json` — Not Started

### Task 5: Regression verification — Not Started [Small]
   a. Verify Increment 2 single-spreader sequences still work (non-fused path) — Not Started
   b. Verify Increment 3 metric references still work (with fused sequences) — Not Started
   c. Run full test suite: `go test ./src/main/go/... -count=1 -timeout 120s` — Not Started

## Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4 → Task 5
```

- Task 1 must complete first (grouping is the foundation for fusion)
- Task 2 depends on Task 1 (fusion uses groups)
- Task 3 depends on Tasks 1 and 2 (edge cases build on fused emission)
- Task 4 depends on all preceding tasks (integration uses fusion)
- Task 5 is verification

No parallel groups — each task builds on the previous.

## Acceptance Criteria Summary

1. Adjacent operations on different dimensions are fused into single matrix operations
2. Fused operations produce identical results to sequential operations
3. Non-fusable sequences (same-dimension aggregate+spread) are emitted as sequential ops
4. Scalar spreading (`from []`) is correctly handled in interleaved sequences
5. Scalar-to-multi-dim and multi-dim-to-scalar edge cases work correctly
6. All fused matrices are emitted in the preamble
7. Error propagation works for fused operations
8. End-to-end examples with optimization produce correct output
9. Regressions: Increment 2 and 3 behavior preserved for non-fused sequences
10. Performance: fused sequences use fewer intermediate tensor allocations

## Intent Summary for Technical Writer

This story delivers the synthesis-phase optimization for mixed aggregator/spreader sequences. When the Increment 4 analyzer produces interleaved `CILSpreadNode`/`CILAggregateNode` chains, the synthesizer analyzes adjacent operations for fusability: operations on different dimensions can be combined into single matrix multiplications, reducing intermediate tensor allocations and improving runtime performance. Two spreads on different dimensions fuse into one expand; an aggregate and spread on different dimensions fuse into a single transform. Operations on the same dimension cannot fuse (they have a data dependency). Scalar spreading (`from []`) is handled as a special case within interleaved sequences. The fused matrices are pre-computed at synthesis time and emitted in the preamble. The `EmitFusedSpreaderSequence` function in the walker replaces the simple chain walking from Increment 2 — it groups, fuses, and emits optimized operations. For non-fusable sequences, the Increment 2/3 sequential emission path is preserved as a fallback.

Note: The fusion optimization is purely a synthesis-phase optimization — it does not change the analysis results. The analysis selects the optimal sequence; the synthesis optimizes how that sequence is executed. The correctness invariant is: fused output == sequential output for all valid inputs.
