# Plan: Story266 — Spreaders Increment 1 (Synthesize)

## Purpose

This plan provides the technical writer with sufficient detail to create Story266: **spreaders-increment1-synthesize** — implement the synthesis (code generation) phase for spreader declarations (Increment 1 of Epic008). Covers CIL tree walking for `CILSpreadNode`, spreading matrix emission as Go preamble, `ExpandDimension` integration (from tensorops), manual spreader test code, and end-to-end spreader examples.

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `spreaders-increment1-synthesize`                                                            |
| description  | Implement CIL tree walker for CILSpreadNode, spreading matrix emission, ExpandDimension integration, manual spreader test code, end-to-end spreader examples |

## Source EROs

- [ERO008: Spreaders Incremental Implementation](../specifications/design/cpl/eros/ERO008_spreaders-incremental-implementation.md) — Increment 1 sketch
- [ERO008: Spreaders Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO008_spreaders-synthesis-phase.md) — CIL tree walking, matrix emission, generated code examples

## Dependencies

- Story265 (spreaders-increment1-analyze) — analysis must produce `CILSpreadNode` with pre-built `SpreadingMatrix`
- Story254 (aggregators-increment1-matrix-mult) — `tensorops.ExpandDimension` function is the core synthesis primitive
- Story257 (aggregators-increment2-synthesize) — CIL tree walker infrastructure and preamble generation pattern already exist; spreader synthesis follows the same patterns

## Files to Create or Modify

| File                                                    | Action | Purpose                                                                          |
| ------------------------------------------------------- | ------ | -------------------------------------------------------------------------------- |
| `src/main/go/synthesis/cil_spread_walker.go`            | Create | `WalkCILSpreadNode` handler for CIL tree walker                                   |
| `src/main/go/synthesis/preamble_generator.go`            | Modify | Add `EmitSpreadingMatrix` function to generate spreading matrix Go var declarations |
| `src/main/go/synthesis/cil_spread_walker_test.go`       | Create | Unit tests: generated Go code for manual spread scenarios                         |
| `examples/spreaders/basic/` (in appropriate with_data dir) | Create | End-to-end example: CPL + IO spec + parquet + expected.json                      |

### WalkCILSpreadNode Handler

```go
func (w *CILTreeWalker) WalkCILSpreadNode(node *CILSpreadNode, scope *WalkScope) error {
    // Walk the source CIL subtree first (gets the tensor variable name)
    sourceVar := w.walkSubtree(node.SourceCIL, scope)

    resultVar := scope.freshVar("spread_" + sanitizeVarName(node.NewDimension))

    // Determine source dimension index and new dimension insertion point
    // For from [] (scalar) case: srcDim=0, newDim=0
    if node.FromEmpty {
        scope.emitCodeIndented(scope.indentLevel, fmt.Sprintf(
            `%s, err := tensorops.ExpandDimension(%s, spreadingMatrix_%s, 0, 0)`,
            resultVar, sourceVar, sanitizeVarName(node.NewDimension),
        ))
    } else {
        srcDim := scope.getDimensionIndex(node.FromDimensions, node.FromDimensions[0])
        newDim := scope.getDimensionIndex(node.ToDimensions, node.NewDimension)
        scope.emitCodeIndented(scope.indentLevel, fmt.Sprintf(
            `%s, err := tensorops.ExpandDimension(%s, spreadingMatrix_%s, %d, %d)`,
            resultVar, sourceVar, sanitizeVarName(node.NewDimension), srcDim, newDim,
        ))
    }

    // Error check
    scope.emitCodeIndented(scope.indentLevel, "if err != nil {")
    scope.emitCodeIndented(scope.indentLevel+1,
        `return nil, fmt.Errorf("spread(%s) failed: %%w", err)`)
    scope.emitCodeIndented(scope.indentLevel, "}")

    scope.registerVariable(resultVar, node)
    scope.pushOperand(resultVar)
    return nil
}
```

### EmitSpreadingMatrix — Preamble Generator

```go
func (g *PreambleGenerator) EmitSpreadingMatrix(
    buf io.Writer,
    name string,
    matrix *AggregationMatrix,
) error {
    // Emits: var spreadingMatrix_<name> = &tensorops.AggregationMatrix{...}
    // ColLabels as []string literal
    // Data as [][]float64 literal
}
```

### Generated Code Example — Equal-Proportion Distribution

For spreader: `from [Product] to [Country, Product] = Product * 1/size(Country) * Country`

```go
// Preamble:
var spreadingMatrix_Country = &tensorops.AggregationMatrix{
    Rows:      1,
    Columns:   2,
    ColLabels: []string{"USA", "Canada"},
    Data: [][]float64{
        {0.5, 0.5},
    },
}

// Test function:
func TestSpreader_EqualDistribution(t *testing.T) {
    // Source tensor: 1D [Product=2], values [100.0, 200.0]
    sourceTensor := chronocone_compute.NewChronoconeTensor[float64](
        []float64{100.0, 200.0},
        []int{2},
    )

    result, err := tensorops.ExpandDimension(
        sourceTensor,
        spreadingMatrix_Country,
        0,  // srcDim: expand from dimension 0 (Product)
        0,  // newDim: insert at position 0 (before Product in shape)
    )
    if err != nil {
        t.Fatalf("spread failed: %v", err)
    }

    // Result: 2D [Country=2 × Product=2]
    // USA,widget=50, USA,gadget=100, Canada,widget=50, Canada,gadget=100
    assert.Equal(t, []int{2, 2}, result.Shape())
}
```

### Scalar Spreading (`from [] to [Country]`)

For `from [] to [Country] = 1/size(Country) * Country`:

```go
// Source is rank-0 (scalar value 42)
sourceScalar := ... // scalar tensor

result, err := tensorops.ExpandDimension(
    sourceScalar,
    spreadingMatrix_Country,
    0,  // srcDim: only dimension (rank-0 treated as dim 0)
    0,  // newDim: insert at position 0
)
// Result: 1D [Country=2], values [21.0, 21.0] (42 * 0.5 each)
```

### end-to-end Example: with_data/spreaders_basic/

Directory structure:
```
examples/spreaders/basic/
├── spreaders_basic.cpl          # CPL with spreaders section and manual test
├── spreaders_basic.io.yaml      # IO spec
├── spreaders_basic_input.xlsx   # Input data (or parquet)
├── spreaders_basic.expected.json # Expected output
```

CPL file example:
```
spreaders
    from [Product] to [Country, Product] = Product * 1/size(Country) * Country

concepts
    Country = [USA, Canada]
    Product = [widget, gadget]

connectors
    connector sales
        grain Country, Product
        data sales_input.xlsx
```

This generates a manual spreader test (not yet integrated into metric DAG — that's Increment 2).

### Test Cases for the Walker

1. **Single-dimension expansion**: `from [Product] to [Country, Product]` — walker emits correct `ExpandDimension(src, matrix, 0, 1)` call
2. **Scalar spreading**: `from [] to [Country]` — walker emits `ExpandDimension(src, matrix, 0, 0)`
3. **Multi-concept from**: `from [Country, Product] to [Country, Product, Region]` — walker emits correct dimensions
4. **Matrix variable naming**: Spreading matrix variable name uses sanitized concept name (`spreadingMatrix_Country`)
5. **Error handling**: Generated code includes `if err != nil` check
6. **Round-trip verification**: Spread then aggregate back yields original values (for equal-proportion distribution with sum aggregator)

## Task Decomposition (for technical writer)

### Task 1: Implement WalkCILSpreadNode in CIL tree walker — Not Started
   a. Create `src/main/go/synthesis/cil_spread_walker.go` — implement `WalkCILSpreadNode` method — Not Started
   b. Handle `FromEmpty` (scalar) case: `srcDim=0, newDim=0` — Not Started
   c. Handle non-empty `from` case: compute correct `srcDim` and `newDim` indices from `FromDimensions`/`ToDimensions` — Not Started
   d. Emit error-checking block after `ExpandDimension` call — Not Started
   e. Write unit tests: verify generated Go code contains correct `ExpandDimension` call with correct parameters for single-dimension expansion, scalar case, multi-concept from — Not Started
   f. Verify walker tests pass — Not Started

### Task 2: Implement EmitSpreadingMatrix preamble generation — Not Started
   a. Add `EmitSpreadingMatrix` function to the preamble generator — writes `var spreadingMatrix_X = &tensorops.AggregationMatrix{...}` Go code — Not Started
   b. Emit `ColLabels` as `[]string` literal, `Data` as `[][]float64` literal — Not Started
   c. Handle empty matrix edge case (0 columns) gracefully — Not Started
   d. Write unit tests: verify generated preamble contains correct matrix data, column labels, dimensions — Not Started
   e. Verify preamble tests pass — Not Started

### Task 3: Create manual spreader test code generation — Not Started
   a. Implement the code generation path that produces a standalone test function for manual spreader application — Not Started
   b. Generated test creates source tensor, calls `tensorops.ExpandDimension`, verifies output shape and values — Not Started
   c. Write unit tests verifying the full generated test function compiles and executes correctly — Not Started
   d. Verify test generation tests pass — Not Started

### Task 4: Create end-to-end spreader examples — Not Started [Small]
   a. Create `with_data/spreaders_basic/` directory with CPL file defining spreaders and concepts — Not Started
   b. Create IO spec and input parquet/XLSX data — Not Started
   c. Create `.expected.json` with ground-truth expected output values — Not Started
   d. Verify full end-to-end pipeline: parse → analyze → synthesize → execute → output matches expected — Not Started
   e. Create scalar spreading example (`from []`) with its own expected output — Not Started

## Execution Order

```
Task 1 → [Task 2 ‖ Task 3] → Task 4
```

- Task 1 must complete first (walker defines the contract for emitted calls)
- Tasks 2 and 3 are parallel (disjoint: preamble generator vs test code generation)
- Task 4 depends on all preceding tasks (examples need full pipeline)

## Parallel Execution

- **Group A** (Tasks 2, 3): Task 2 modifies `preamble_generator.go`; Task 3 creates separate test generation code. Same package but different files and concerns.

## Reintegration

**Group A (Tasks 2, 3):**
- Merge order: Task 2 first (preamble emission is infrastructure), then Task 3 (test generation uses preamble)
- Integration tests: `go test ./src/main/go/synthesis/... -count=1 -timeout 60s` — verify both preamble and test generation compile together
- Watch for conflicts in: `preamble_generator.go` (Task 2 modifies it)

## Acceptance Criteria Summary

1. `WalkCILSpreadNode` emits correct `tensorops.ExpandDimension` calls with correct `srcDim` and `newDim` parameters
2. Scalar spreading (`from []`) handled as special case with `srcDim=0, newDim=0`
3. `EmitSpreadingMatrix` produces valid Go var declarations with correct matrix data
4. Generated test code produces correct results: spread a tensor using equal-proportion distribution
5. Spread result has correct shape: output has one more dimension than input
6. Spread then aggregate (with matching aggregator) yields original values
7. Error case: matrix dimension mismatch produces runtime error with clear message
8. End-to-end examples pass: `with_data/spreaders_basic/` produces output matching `.expected.json`
9. All tests pass with `go test ./src/main/go/synthesis/... -count=1 -timeout 60s`

## Intent Summary for Technical Writer

This story delivers the synthesis (code generation) phase for spreader declarations. The CIL tree walker handles `CILSpreadNode` by emitting calls to `tensorops.ExpandDimension` — the dimension-expanding matrix multiplication function built in Story254. No new matrix multiplication infrastructure is needed; spreaders are the first consumer of the expansion path. The preamble generator emits `spreadingMatrix_X` Go var declarations for each spreader's pre-built matrix. Spreader synthesis generates manually-applied test code (not yet integrated into metric computation — that's Increment 2). The walker must handle the `from []` (scalar) edge case where the source has rank zero. For equal-proportion distribution, spread then aggregate yields the original values, providing a natural verification. The generated code follows the same patterns as aggregator synthesis (Story257): emitted function calls + error checking + result registration in walk scope.
