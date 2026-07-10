# Plan: Story258 — Aggregators Increment 3 (Analyze)

## Purpose

This plan provides the technical writer with sufficient detail to create Story258: **aggregators-increment3-analyze** — define operation-specific CIL node types for all aggregator functions, register them in the full FunctionRegistry, implement `resolveAggregatorFnCall` that creates the appropriate typed CIL node, and replace the generic `CILAggregateNode` from Increment 2 (Increment 3 of Epic007).

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `aggregators-increment3-analyze`                                                              |
| description  | Define operation-specific CIL node types (sum, count, size, 1/sum, 1/count), extend FunctionRegistry, implement resolveAggregatorFnCall, CIL node insertion in expressions |

## Source EROs

- [ERO007: Aggregators Incremental Implementation](../specifications/design/cpl/eros/ERO007_aggregators-incremental-implementation.md) — Increment 3 sketch
- [ERO007: Aggregators Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO007_aggregators-analysis-phase.md) — FunctionRegistry full registration, CIL node types, `resolveAggregatorFnCall`, expression CIL tree construction

## Dependencies

- Story256 (aggregators-increment2-analyze) — needs `FunctionRegistry`, `DimensionalUnitExpression`, postfix evaluator, concept resolution from Increment 2
- Story257 (aggregators-increment2-synthesize) — needs working synthesis from Increment 2 to validate CIL node types integrate correctly

## Files to Create or Modify

| File                                                    | Action  | Purpose                                                                          |
| ------------------------------------------------------- | ------- | -------------------------------------------------------------------------------- |
| `src/main/go/analysis/cil_aggregate_sum_node.go`        | Create  | `CILAggregateSumNode` struct implementing `CILNode`                             |
| `src/main/go/analysis/cil_aggregate_count_node.go`      | Create  | `CILAggregateCountNode` struct                                                  |
| `src/main/go/analysis/cil_aggregate_size_node.go`       | Create  | `CILAggregateSizeNode` struct (no SourceCIL, just dim name and constant size)   |
| `src/main/go/analysis/cil_aggregate_inverse_sum_node.go` | Create  | `CILAggregateInverseSumNode` struct                                             |
| `src/main/go/analysis/cil_aggregate_inverse_count_node.go` | Create  | `CILAggregateInverseCountNode` struct                                         |
| `src/main/go/analysis/function_registry.go`             | Modify  | Register `count`, `size`, `1/sum`, `1/count` with dim unit effects and CIL node types |
| `src/main/go/analysis/analyze_aggregators.go`           | Modify  | Replace generic `CILAggregateNode` creation with `resolveAggregatorFnCall`; update expression postfix evaluator to create typed CIL nodes |

### New CIL Node Types (Increment 3)

Each function gets its own CIL node type, replacing the generic `CILAggregateNode` from Increment 2:

```go
// CILAggregateSumNode applies sum aggregation.
type CILAggregateSumNode struct {
    SourceCIL   CILNode              // CIL subtree for the source metric/dimension
    DimName     string               // Name of the concept dimension being summed
    Matrix      *AggregationMatrix   // Pre-computed sum aggregation matrix
    SourceLocation SourceLocation
}
func (n *CILAggregateSumNode) cilNode() {}

// CILAggregateCountNode applies count aggregation.
type CILAggregateCountNode struct {
    SourceCIL   CILNode
    DimName     string
    Matrix      *AggregationMatrix   // Uses same sum matrix pattern
    SourceLocation SourceLocation
}

// CILAggregateSizeNode returns size of a concept (compile-time constant).
// No SourceCIL — it produces a scalar from concept metadata.
type CILAggregateSizeNode struct {
    DimName     string
    Size        int                  // Number of distinct values in concept
    SourceLocation SourceLocation
}

// CILAggregateInverseSumNode applies 1/sum aggregation.
type CILAggregateInverseSumNode struct {
    SourceCIL   CILNode
    DimName     string
    Matrix      *AggregationMatrix   // Each entry = 1/N
    SourceLocation SourceLocation
}

// CILAggregateInverseCountNode applies 1/count aggregation.
type CILAggregateInverseCountNode struct {
    SourceCIL   CILNode
    DimName     string
    Matrix      *AggregationMatrix   // Each entry = 1/N
    SourceLocation SourceLocation
}
```

### FunctionRegistry — Full Registration

Extend `NewAggregatorFunctionRegistry()` to register all five functions:

| Function   | CILNodeType                     | DimUnits Formula              |
|-----------|--------------------------------|-------------------------------|
| `sum`     | `CILAggregateSumNode`          | `{Numerator: [param]}`        |
| `count`   | `CILAggregateCountNode`        | `{Numerator: [param]}`        |
| `size`    | `CILAggregateSizeNode`         | `{}` (unitless)               |
| `1/sum`   | `CILAggregateInverseSumNode`   | `{Denominator: [param]}`      |
| `1/count` | `CILAggregateInverseCountNode` | `{Denominator: [param]}`      |

The `1/sum` and `1/count` names use `/` as part of a compound identifier. The parser handles this as a single identifier token.

### resolveAggregatorFnCall

```go
func (a *CPLAnalyzer) resolveAggregatorFnCall(
    fnCall *OperandData,    // {"fncall", "sum", ["Country"]}
    sourceCIL CILNode,      // CIL subtree for the source
) (CILNode, error) {
    fnInfo, exists := a.functionRegistry.Functions[fnCall.Name]
    if !exists {
        return nil, fmt.Errorf("unknown aggregator function %q", fnCall.Name)
    }
    
    dimName := fnCall.Params[0]  // The concept being aggregated
    
    switch fnInfo.CILNodeType {
    case "CILAggregateSumNode":
        matrix := BuildSumAggregationMatrix(...)
        return &CILAggregateSumNode{...}, nil
    case "CILAggregateCountNode":
        matrix := BuildSumAggregationMatrix(...)
        return &CILAggregateCountNode{...}, nil
    case "CILAggregateSizeNode":
        return &CILAggregateSizeNode{DimName: dimName, Size: len(values)}, nil
    case "CILAggregateInverseSumNode":
        matrix := BuildInverseSumAggregationMatrix(...)
        return &CILAggregateInverseSumNode{...}, nil
    case "CILAggregateInverseCountNode":
        matrix := BuildInverseSumAggregationMatrix(...)
        return &CILAggregateInverseCountNode{...}, nil
    }
}
```

### Expression CIL Tree Construction

When processing `(a-b)*sum(Country)`, the CIL tree becomes:
```
CILBinaryOp("*",
    CILBinaryOp("-", conceptRef_a, conceptRef_b),
    CILAggregateSumNode{source: countryRef, dim: "Country", matrix: ...}
)
```

The postfix evaluator in `AnalyzeAggregatorScope` is updated to:
1. Recognize `fncall` operands and call `resolveAggregatorFnCall` which creates the typed CIL node
2. Handle arithmetic operators: `+, -, *, /` create `CILBinaryOp` nodes wrapping the operand CIL nodes
3. The resulting CIL tree is stored in `AnalysisDefinitions.CILTrees`

### replaceGenericAggregateNode

The generic `CILAggregateNode` from Increment 2 must be retained (or cleaned up) — Increment 3 replaces its usage in `AnalyzeAggregatorScope`. The struct can remain defined but is no longer created during analysis.

### Test Cases

- **All functions create correct CIL node type**: `sum` → `CILAggregateSumNode`, `count` → `CILAggregateCountNode`, `size` → `CILAggregateSizeNode` (with correct Size value), `1/sum` → `CILAggregateInverseSumNode`, `1/count` → `CILAggregateInverseCountNode`
- **Expression CIL tree**: `(a-b)*sum(Country)` produces correct tree structure
- **Registry dim units**: each function produces correct `DimensionalUnitExpression`
- **Error cases**: unknown function name, invalid parameter count
- **Inverse matrices**: `BuildInverseSumAggregationMatrix` produces entries = 1/N

## Task Decomposition (for technical writer)

### Task 1: Define typed CIL node structs — Not Started
   a. Create `src/main/go/analysis/cil_aggregate_sum_node.go` with `CILAggregateSumNode` — Not Started
   b. Create `src/main/go/analysis/cil_aggregate_count_node.go` with `CILAggregateCountNode` — Not Started
   c. Create `src/main/go/analysis/cil_aggregate_size_node.go` with `CILAggregateSizeNode` — Not Started
   d. Create `src/main/go/analysis/cil_aggregate_inverse_sum_node.go` with `CILAggregateInverseSumNode` — Not Started
   e. Create `src/main/go/analysis/cil_aggregate_inverse_count_node.go` with `CILAggregateInverseCountNode` — Not Started
   f. Write tests verifying each struct initializes correctly and implements `CILNode` interface — Not Started
   g. Verify package compiles — Not Started

### Task 2: Extend FunctionRegistry with all functions — Not Started
   a. Modify `NewAggregatorFunctionRegistry()` in `function_registry.go` to register `count`, `size`, `1/sum`, `1/count` — Not Started
   b. Verify dim unit functions for each: `count(C)`→`{Num:[C]}`, `size(C)`→`{}`, `1/sum(C)`→`{Den:[C]}`, `1/count(C)`→`{Den:[C]}` — Not Started
   c. Write tests: lookup each function, verify CILNodeType string, verify DimUnits output — Not Started
   d. Verify registry tests pass — Not Started

### Task 3: Implement resolveAggregatorFnCall and integrate into analysis — Not Started
   a. Implement `resolveAggregatorFnCall` in `analyze_aggregators.go` — dispatch on CILNodeType to create correct typed node — Not Started
   b. Update postfix evaluator to call `resolveAggregatorFnCall` for fncall operands, replacing generic `CILAggregateNode` creation — Not Started
   c. Update expression CIL tree construction: arithmetic operators create `CILBinaryOp` wrapping operand CIL nodes — Not Started
   d. Write tests: `sum(Country)` creates `CILAggregateSumNode`; `size(Product)` creates `CILAggregateSizeNode` with correct Size; `(a-b)*sum(Country)` produces correct CIL tree — Not Started
   e. Write error tests: unknown function `foo(Country)` returns error — Not Started
   f. Verify all analysis tests pass — Not Started

## Execution Order

```
Task 1 → Task 2 → Task 3
```

- Task 1 must complete first (CIL node types must exist before registry can reference them)
- Task 2 depends on Task 1 (registry maps to CIL node types)
- Task 3 depends on Task 2 (resolveAggregatorFnCall uses registry)

All sequential — no parallelization.

## Parallel Execution

None — all tasks have sequential dependencies.

## Acceptance Criteria Summary

1. Five typed CIL node structs defined, all implementing `CILNode` interface
2. `CILAggregateSizeNode` has no `SourceCIL` field (size is a compile-time constant)
3. `FunctionRegistry` registers all five functions with correct `CILNodeType` strings and `DimUnits` formulas
4. `resolveAggregatorFnCall` dispatches to correct node type for each function name
5. Expression postfix evaluator creates correct CIL tree: `(a-b)*sum(Country)` → `CILBinaryOp("*", CILBinaryOp("-", ref_a, ref_b), CILAggregateSumNode{...})`
6. Unknown function names produce clear error messages
7. Inverse matrices (`1/sum`, `1/count`) use `BuildInverseSumAggregationMatrix` with entries = 1/N
8. Generic `CILAggregateNode` from Increment 2 is no longer created (struct may remain for backward compat)
9. All tests pass with `go test ./src/main/go/analysis/...`

## Intent Summary for Technical Writer

This story replaces the generic `CILAggregateNode` from Increment 2 with five operation-specific CIL node types, each carrying its own aggregation semantics. The FunctionRegistry is extended to map all function names (`sum`, `count`, `size`, `1/sum`, `1/count`) to their CIL node types and dimensional unit effects. The `resolveAggregatorFnCall` function creates the appropriate typed node based on registry lookup. Because aggregator functions fit into expressions (e.g., `(a-b)*sum(Country)`), each CIL node type behaves as an expression operand — its value participates in arithmetic operations via `CILBinaryOp` wrappers. The `size(C)` function is special: it has no `SourceCIL` field since it produces a compile-time constant (the number of concept values). No parser changes are needed — the parser already stores function calls as generic `fncall` operands, and the analysis phase resolves them against the registry. The CIL tree now contains typed aggregate nodes that the synthesis phase (Story259) will walk individually.
