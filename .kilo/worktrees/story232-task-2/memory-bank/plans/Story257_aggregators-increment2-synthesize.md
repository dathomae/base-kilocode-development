# Plan: Story257 — Aggregators Increment 2 (Synthesize)

## Purpose

This plan provides the technical writer with sufficient detail to create Story257: **aggregators-increment2-synthesize** — implement CIL tree walking for `CILAggregateNode`, emission of aggregation matrix definitions in generated preamble, manual aggregation test code, and end-to-end examples (Increment 2 of Epic007).

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `aggregators-increment2-synthesize`                                                           |
| description  | Implement CIL tree walker for CILAggregateNode, aggregation matrix Go code emission, manual sum aggregation test code, end-to-end aggregator examples with CPL+parquet+expected output |

## Source EROs

- [ERO007: Aggregators Incremental Implementation](../specifications/design/cpl/eros/ERO007_aggregators-incremental-implementation.md) — Increment 2 sketch
- [ERO007: Aggregators Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO007_aggregators-synthesis-phase.md) — CIL tree walker, matrix emission, manual aggregation, generated code structure

## Dependencies

- Story254 (aggregators-increment1-matrix-mult) — `tensorops.ReduceDimension` must exist for code emission
- Story256 (aggregators-increment2-analyze) — needs `CILAggregateNode` with pre-built matrices in CIL trees
- Story255 (aggregators-increment2-parse) — needs parser to produce symbol table entries

## Files to Create or Modify

| File                                                    | Action  | Purpose                                                                          |
| ------------------------------------------------------- | ------- | -------------------------------------------------------------------------------- |
| `src/main/go/cil/cil_transformer.go`                    | Modify  | Add `WalkCILAggregateNode` handler in CIL tree walker                            |
| `src/main/go/cil/cil_code_generator.go`                 | Modify  | Add `GenerateAggregationMatrix` method for preamble matrix emission              |
| `src/main/go/synthesis/synthesize.go`                   | Modify  | Integrate aggregation matrix emission into synthesis preamble generation         |
| `src/main/go/cil/cil_transformer_aggregator_test.go`    | Create  | Tests for CILAggregateNode walking: generated code emits ReduceDimension call    |
| Examples:                                               |         |                                                                                  |
| `examples/parse_and_analyze_only/without_errors/aggregators_basic.cpl` | Create  | CPL with aggregator declarations (already created in Story255)                   |
| `examples/with_data/aggregators_basic/`                 | Create  | CPL + IO spec + parquet data + `.expected.xlsx` for manual sum aggregation       |

### CIL Tree Walker — CILAggregateNode Handler

The existing CIL tree walker (`cil_transformer.go`) must be extended to handle `CILAggregateNode`:

```go
func (w *CILTreeWalker) WalkCILAggregateNode(node *CILAggregateNode, scope *WalkScope) error {
    // Walk the source CIL subtree first (may be a connector reference)
    sourceVar := w.walkSubtree(node.SourceCIL, scope)

    // Generate a variable name for the aggregated result
    resultVar := scope.freshVar("aggregated")

    // Determine the collapsed dimension index in source tensor
    collapsedDim := scope.getDimensionIndex(node.FromDimensions, node.ToDimension)

    // Emit the aggregation call using tensorops.ReduceDimension
    scope.emitCodeIndented(scope.indentLevel, fmt.Sprintf(
        `%s, err := tensorops.ReduceDimension(%s, aggregationMatrix_%s, %d, %d)`,
        resultVar, sourceVar, sanitizeVarName(node.ToDimension),
        collapsedDim, scope.getTargetDimIndex(node),
    ))
    scope.emitCodeIndented(scope.indentLevel, "if err != nil {")
    scope.emitCodeIndented(scope.indentLevel+1, `return nil, fmt.Errorf("aggregation failed: %%w", err)`)
    scope.emitCodeIndented(scope.indentLevel, "}")

    // Register result for use by subsequent operations
    scope.registerVariable(resultVar, node)
    scope.pushOperand(resultVar)
    return nil
}
```

### Aggregation Matrix Emission in Preamble

`CILCodeGenerator` must emit Go code that materializes pre-built aggregation matrices as package-level variables:

```go
func (g *CILCodeGenerator) GenerateAggregationMatrix(name string, matrix *AggregationMatrix) string {
    // Produces:
    // var aggregationMatrix_Country = &tensorops.AggregationMatrix{
    //     Rows: 2, Columns: 2,
    //     RowLabels: []string{"USA", "Canada"},
    //     ColLabels: []string{"widget", "gadget"},
    //     Data: [][]float64{{1.0, 1.0}, {1.0, 1.0}},
    // }
}
```

The synthesis preamble generator (`synthesize.go`) iterates over aggregator CIL trees and calls `GenerateAggregationMatrix` for each, placing declarations before metric computation functions.

### Generated Code Structure (Increment 2 additions)

```
main()
├── Parse CPL source
├── Parse IO specifications
├── Create ChronoconeComputationController
├── Load connector data → register MetricComputations
├── Define aggregation matrices                         ← NEW: aggregator matrices
├── Schedule metric computation order (topological)
├── Compute metrics
│   ├── ReduceDimension(input, aggregationMatrix_X, ...)  ← NEW: aggregator applied
│   └── Calculate cell-wise result
└── Output results
```

### Test Code Pattern (Manual Aggregation)

In Increment 2, aggregators are applied manually in test code — not yet integrated into metric computation. Test verifies:
1. Load source tensor from parquet data
2. Compute expected output (hand-computed ground truth)
3. Apply `ReduceDimension` with the pre-built matrix
4. Verify result shape (one fewer dimension) and values match expected

### with_data Example: aggregators_basic

- **CPL**: `aggregators` section with `from [Country, Product] to [Product] = sum(Country)`, concepts, connectors for source data
- **IO spec**: defines input parquet connector with `[Country, Product]` dimensions, output spec for `[Product]` result
- **parquet data**: small dataset with known values (e.g., Country={USA,Canada} × Product={widget,gadget})
- **expected output**: hand-computed sum-aggregated values

## Task Decomposition (for technical writer)

### Task 1: Add CILAggregateNode walking to CIL tree walker — Not Started
   a. Add `WalkCILAggregateNode` method to `CILTreeWalker` in `src/main/go/cil/cil_transformer.go` — Not Started
   b. Wire `CILAggregateNode` case into the main `Walk` dispatch (type-switch or visitor method) — Not Started
   c. Write unit tests: walk a `CILAggregateNode` → verify generated code contains `tensorops.ReduceDimension` call with correct variable names, dimension indices, and matrix parameter — Not Started
   d. Write tests: verify error handling code emission (err check after ReduceDimension) — Not Started
   e. Verify walker tests pass — Not Started

### Task 2: Implement aggregation matrix emission in code generator — Not Started
   a. Add `GenerateAggregationMatrix` method to `CILCodeGenerator` in `cil_code_generator.go` — Not Started
   b. Implement matrix data formatting: row labels, column labels, data rows as `[][]float64` literals — Not Started
   c. Write tests: verify generated Go code for a 2×2 sum matrix, a 3×1 size matrix, an inverse matrix — Not Started
   d. Verify code generator tests pass — Not Started

### Task 3: Integrate matrix emission into synthesis preamble — Not Started
   a. In `synthesize.go`, add preamble generation that iterates aggregator CIL trees and calls `GenerateAggregationMatrix` for each — Not Started
   b. Add `"tensorops"` import to generated code's import block — Not Started
   c. Verify synthesized program compiles (with synthetic test input that triggers aggregator code path) — Not Started
   d. Verify integration tests pass — Not Started

### Task 4: Create with_data aggregator example and manual test — Not Started
   a. Create `examples/with_data/aggregators_basic/` directory — Not Started
   b. Write CPL file with concepts, connector, aggregator declaration — Not Started
   c. Write IO specification YAML — Not Started
   d. Generate parquet test data — Not Started
   e. Create `.expected.xlsx` with ground-truth aggregated values — Not Started
   f. Write manual test function that loads parquet, applies `ReduceDimension`, verifies against expected — Not Started
   g. Verify example test passes — Not Started

## Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4
```

All tasks have sequential dependencies:
- Task 2 needs Task 1's walker (to know the interface between walker and code generator)
- Task 3 needs Task 2 (matrix emission must exist before integration)
- Task 4 needs Task 3 (full pipeline must work for examples)

## Parallel Execution

None — all tasks have sequential dependencies.

## Acceptance Criteria Summary

1. `CILTreeWalker` correctly handles `CILAggregateNode`: emits `tensorops.ReduceDimension` call with correct variable name, dimension index, and matrix parameter name
2. Generated code includes error handling (`if err != nil`) after ReduceDimension call
3. `CILCodeGenerator.GenerateAggregationMatrix` emits valid Go `tensorops.AggregationMatrix` literals with correct row/column labels and data
4. Synthesis preamble includes matrix definitions for all aggregators in compilation unit, before metric computation functions
5. Generated program imports `tensorops` package
6. Manual test: loading parquet data, applying `ReduceDimension` with sum matrix, produces correct aggregated results matching ground truth
7. `with_data/aggregators_basic/` example passes validation (parse → analyze → CIL → synthesize → execute → verify)
8. All tests pass with `go test ./src/main/go/cil/... ./src/main/go/synthesis/...`

## Intent Summary for Technical Writer

This story delivers the synthesis phase for basic aggregators (sum only). The CIL tree walker handles `CILAggregateNode` by emitting calls to `tensorops.ReduceDimension` from the matrix multiplication infrastructure built in Story254. Aggregation matrices are emitted as Go literals in the synthesis preamble (matrix values from concept value sets are materialized at compile time). The generated code follows the existing patterns: `inputs` map access, error propagation, and tensor variable naming. Aggregators are applied manually in test code — not yet integrated into metric computation (that happens in Increment 4). The `with_data/aggregators_basic/` example provides the first end-to-end aggregator validation: CPL file → parse → analyze → CIL gen → synthesize → execute → verify aggregated output.
