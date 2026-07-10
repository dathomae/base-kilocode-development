# Plan: Story259 — Aggregators Increment 3 (Synthesize)

## Purpose

This plan provides the technical writer with sufficient detail to create Story259: **aggregators-increment3-synthesize** — implement CIL tree walking for all five operation-specific CIL node types, handle `size(C)` constant emission, integrate aggregator calls into expression CIL trees, and create end-to-end examples (Increment 3 of Epic007).

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `aggregators-increment3-synthesize`                                                           |
| description  | Implement CIL tree walker for all function CIL node types (sum, count, size, 1/sum, 1/count), size(C) constant emission, expression-integration examples |

## Source EROs

- [ERO007: Aggregators Incremental Implementation](../specifications/design/cpl/eros/ERO007_aggregators-incremental-implementation.md) — Increment 3 sketch
- [ERO007: Aggregators Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO007_aggregators-synthesis-phase.md) — Walk functions for each CIL node type, size constant, expression code generation

## Dependencies

- Story258 (aggregators-increment3-analyze) — needs typed CIL node types (`CILAggregateSumNode`, etc.) from analysis
- Story257 (aggregators-increment2-synthesize) — needs existing `WalkCILAggregateNode` pattern, matrix emission infrastructure
- Story254 (aggregators-increment1-matrix-mult) — `tensorops.ReduceDimension` for code emission

## Files to Create or Modify

| File                                                    | Action  | Purpose                                                                          |
| ------------------------------------------------------- | ------- | -------------------------------------------------------------------------------- |
| `src/main/go/cil/cil_transformer.go`                    | Modify  | Add walk methods for each typed CIL node type; remove/update generic CILAggregateNode walk |
| `src/main/go/cil/cil_transformer_aggregator_test.go`    | Modify  | Extend tests for all five node types; expression integration tests               |
| Examples:                                               |         |                                                                                  |
| `examples/parse_and_analyze_only/without_errors/aggregators_functions.cpl` | Create | CPL with aggregators using `count`, `size`, `1/sum`, `1/count`                  |
| `examples/parse_and_analyze_only/with_errors/aggregators_unknown_function.cpl` | Create | CPL with unknown function name                                                  |
| `examples/with_data/aggregators_functions/`             | Create  | CPL + IO spec + parquet + expected for multiple function types                  |

### CIL Tree Walker — New Walk Methods

Add five new walk methods to the CIL tree walker. Each follows the same pattern established in Story257 but emits the appropriate operation:

**WalkCILAggregateSumNode** and **WalkCILAggregateCountNode**: Same pattern as `WalkCILAggregateNode` from Increment 2 — walk source subtree, emit `tensorops.ReduceDimension(source, aggregationMatrix_<dim>, collapsedDim, targetDim)`, error check, push operand.

**WalkCILAggregateSizeNode**: Special case — no tensor source. Emits a constant:
```go
func (w *CILTreeWalker) WalkCILAggregateSizeNode(node *CILAggregateSizeNode, scope *WalkScope) error {
    resultVar := scope.freshVar("size_" + node.DimName)
    scope.emitCodeIndented(scope.indentLevel, fmt.Sprintf(
        `%s := float64(%d)  // size(%s)`, resultVar, node.Size, node.DimName,
    ))
    scope.pushOperand(resultVar)
    return nil
}
```

**WalkCILAggregateInverseSumNode** and **WalkCILAggregateInverseCountNode**: Same as sum/count but use inverse matrix (entries = 1/N).

### Walk Dispatch Update

The main walk dispatch (type-switch or visitor method) is extended to handle each new node type:
```go
case *CILAggregateSumNode:
    return w.WalkCILAggregateSumNode(node, scope)
case *CILAggregateCountNode:
    return w.WalkCILAggregateCountNode(node, scope)
case *CILAggregateSizeNode:
    return w.WalkCILAggregateSizeNode(node, scope)
case *CILAggregateInverseSumNode:
    return w.WalkCILAggregateInverseSumNode(node, scope)
case *CILAggregateInverseCountNode:
    return w.WalkCILAggregateInverseCountNode(node, scope)
```

### Expression Integration

When aggregator functions appear inside arithmetic expressions (e.g., `(a-b)*sum(Country)`), the CIL tree is: `CILBinaryOp("*", CILBinaryOp("-", a_ref, b_ref), CILAggregateSumNode{...})`. The walker naturally handles this:
1. Walking `CILBinaryOp("*")` walks left and right subtrees
2. The right subtree is `CILAggregateSumNode` — walker emits `ReduceDimension` and pushes the result variable
3. The left subtree computes `a - b`
4. The `*` operator combines them

The existing binary op walker already pushes operands to a stack; the aggregate node walker simply pushes its result as an operand.

### Matrix Emission for New Types

Each aggregator function needs its aggregation matrix emitted in the preamble. The `CILCodeGenerator` already supports `GenerateAggregationMatrix` from Story257. For `CILAggregateSizeNode`, no matrix is emitted (it's a constant). For inverse nodes, `BuildInverseSumAggregationMatrix` is used (entries = 1/N where N is the number of source concept values).

### with_data Examples

**aggregators_functions/**: CPL + IO spec + parquet + expected:
- `sum` aggregation (validates Increment 2 behavior still works after refactor)
- `count` aggregation (same as sum but with count semantics)
- `size` aggregation (produces constant value)
- `1/sum` aggregation (inverse weights)
- Expression: `(a-b)*sum(Country)` (validates aggregator in arithmetic expression)

## Task Decomposition (for technical writer)

### Task 1: Add walk methods for sum and count nodes — Not Started [Small]
   a. Add `WalkCILAggregateSumNode` to `CILTreeWalker` in `cil_transformer.go` — Not Started
   b. Add `WalkCILAggregateCountNode` to `CILTreeWalker` — Not Started
   c. Wire both cases into the main walk dispatch — Not Started
   d. Write tests: walking `CILAggregateSumNode` → emits `ReduceDimension` with `aggregationMatrix_sum_X`; walking `CILAggregateCountNode` → emits `ReduceDimension` with `aggregationMatrix_count_X` — Not Started
   e. Verify tests pass — Not Started

### Task 2: Add walk method for size node (constant emission) — Not Started [Small]
   a. Add `WalkCILAggregateSizeNode` to `CILTreeWalker` — emit constant `float64(size)` — Not Started
   b. Wire case into main dispatch — Not Started
   c. Write tests: `CILAggregateSizeNode{Size: 3}` → emitted code contains `size_<dim> := float64(3)` — Not Started
   d. Verify tests pass — Not Started

### Task 3: Add walk methods for inverse sum and inverse count nodes — Not Started [Small]
   a. Add `WalkCILAggregateInverseSumNode` and `WalkCILAggregateInverseCountNode` — same ReduceDimension pattern with inverse matrix — Not Started
   b. Wire cases into main dispatch — Not Started
   c. Write tests: verify `ReduceDimension` call uses `aggregationMatrix_invsum_X` / `aggregationMatrix_invcount_X` — Not Started
   d. Verify tests pass — Not Started

### Task 4: Expression integration test — Not Started
   a. Write test: CIL tree `CILBinaryOp("*", CILBinaryOp("-", ref_a, ref_b), CILAggregateSumNode{...})` → walk → verify generated code computes `a-b`, computes sum, multiplies results — Not Started
   b. Verify generated code variable flow: source → aggregated → expression operand — Not Started
   c. Verify tests pass — Not Started

### Task 5: Create with_data aggregator functions example — Not Started
   a. Create `examples/with_data/aggregators_functions/` directory — Not Started
   b. Write CPL file with aggregators using `count`, `size`, `1/sum`, `1/count` — Not Started
   c. Write IO spec and generate parquet data — Not Started
   d. Create `.expected.xlsx` with ground-truth values — Not Started
   e. Create expression example: `(a-b)*sum(Country)` with parquet data — Not Started
   f. Create error examples: unknown function name — Not Started
   g. Verify all examples pass validation — Not Started

## Execution Order

```
[Task 1 ‖ Task 2 ‖ Task 3] → Task 4 → Task 5
```

- Tasks 1, 2, 3 are parallel (disjoint walk methods in the same file but separate functions)
- Task 4 depends on Tasks 1-3 (expression integration uses sum node from Task 1)
- Task 5 depends on all preceding tasks (examples need full pipeline)

## Parallel Execution

- **Group A** (Tasks 1, 2, 3): All three tasks add walk methods to `cil_transformer.go` (separate functions) and add dispatch cases. Same file, different regions. Worktree isolation required; merge may need resolution in the dispatch switch.

## Reintegration

**Group A (Tasks 1, 2, 3):**
- Merge order: Task 1 first (sum/count = main pattern), then Task 3 (inverse = variant of Task 1), then Task 2 (size = separate pattern)
- Integration tests: `go test ./src/main/go/cil/... -count=1 -timeout 60s` — verify all five walk methods compile and test together
- Watch for conflicts in: `cil_transformer.go` (all tasks add methods and dispatch cases)

## Acceptance Criteria Summary

1. Five typed walk methods exist on `CILTreeWalker`: `WalkCILAggregateSumNode`, `WalkCILAggregateCountNode`, `WalkCILAggregateSizeNode`, `WalkCILAggregateInverseSumNode`, `WalkCILAggregateInverseCountNode`
2. `WalkCILAggregateSumNode` and `WalkCILAggregateCountNode` emit `tensorops.ReduceDimension(source, aggregationMatrix_<name>, collDim, targetDim)` with error checking
3. `WalkCILAggregateSizeNode` emits `float64(N)` constant (no tensor source, no ReduceDimension)
4. `WalkCILAggregateInverseSumNode` and `WalkCILAggregateInverseCountNode` emit `ReduceDimension` with inverse matrix names
5. Expression CIL trees containing aggregate nodes produce correct generated code where aggregate results are computed as local variables and used in arithmetic
6. Synthesis preamble emits matrices for sum, count, inverse_sum, inverse_count (no matrix for size)
7. `with_data/aggregators_functions/` examples pass: all five functions produce correct aggregated results
8. Expression example `(a-b)*sum(Country)` produces correct results
9. All tests pass with `go test ./src/main/go/cil/... ./src/main/go/synthesis/...`

## Intent Summary for Technical Writer

This story delivers synthesis for all five aggregator function CIL node types. The CIL tree walker is extended from handling one generic `CILAggregateNode` to five typed methods. Three patterns exist: (1) `sum`/`count` — standard `ReduceDimension` call; (2) `1/sum`/`1/count` — same pattern with inverse matrix; (3) `size` — compile-time constant emission (no tensor operation). Each aggregator function fits into the CIL expression tree as an operand, so the walker emits a local variable computation followed by the variable reference as an expression operand. The existing binary op walker naturally composes these operands into arithmetic operations. The `size(C)` function is unique: it has no `SourceCIL` (no tensor data access), just a constant integer from the concept's value count — this paths through synthesis differently from the matrix-based nodes.
