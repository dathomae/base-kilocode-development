# Story258: Aggregators Increment 3 (Analyze)

## Goal

Define operation-specific CIL node types for all aggregator functions (`sum`, `count`, `size`, `1/sum`, `1/count`), extend the `FunctionRegistry` to register these functions with their dimensional unit effects and CIL node type mappings, implement `resolveAggregatorFnCall` that creates the appropriate typed CIL node based on registry lookup, and integrate typed aggregate nodes into the expression postfix evaluator so that aggregator function calls produce correctly typed nodes within CIL expression trees.

This story replaces the generic `CILAggregateNode` from [Increment 2](#dependencies) with five operation-specific CIL node types, each carrying its own aggregation semantics (sum, count, size, inverse-sum, inverse-count).

This story is Increment 3 of Epic007 (Aggregators Incremental Implementation).

## References

- [ERO007: Aggregators Incremental Implementation](specifications/design/cpl/eros/ERO007_aggregators-incremental-implementation.md) — Increment 3 sketch
- [ERO007: Aggregators Analysis Phase](specifications/design/cpl/go-analysis/eros/ERO007_aggregators-analysis-phase.md) — FunctionRegistry full registration, CIL node types, `resolveAggregatorFnCall`, expression CIL tree construction
- [Aggregators Design](specifications/design/cpl/aggregators.md) — Core aggregator design
- [Postfix Expression Parsing Pattern](specifications/patterns/postfix-expression-parsing.md) — Expression handling pattern
- [Symbol Table-Driven Parsing Pattern](specifications/patterns/symbol-table-driven-parsing.md) — Phase separation pattern

## Dependencies

- **Story256** (`aggregators-increment2-analyze`) — _Not Started_: Must complete first. Provides `FunctionRegistry` struct, `AggregatorFunctionInfo`, `DimensionalUnitExpression` with `Multiply`/`cancel`/`Equals` methods, `AggregationMatrix` struct, `BuildSumAggregationMatrix`, `BuildInverseSumAggregationMatrix`, the postfix expression evaluator in `analyze_aggregators.go`, concept resolution in `CPLAnalyzer`, and the generic `CILAggregateNode` struct.
- **Story257** (`aggregators-increment2-synthesize`) — _Not Started_: Must complete first. Provides working synthesis from Increment 2 to validate that CIL node types integrate correctly into the synthesis pipeline.

## Dependent Stories

- **Story259** (`aggregators-increment3-synthesize`) — Will depend on this story: the synthesis CIL tree walker must handle the five new CIL node types defined here.

## Tasks

### Task Sizing

All four tasks in this story are medium size — each modifies multiple coordinated files, introduces new structs and methods, and requires reasoning across multiple concerns. No task qualifies as `[Small]`:

- **Task 1** touches six files (five new struct files plus a test file). Each struct is mechanically simple, but the task introduces five new CIL-node types that must each implement the `cil.CILNode` interface from the external `cil` package, requiring coordinated knowledge of the interface contract and the aggregation matrix types from the analysis package.
- **Task 2** modifies `function_registry.go` and a test file. It extends an existing function (`NewAggregatorFunctionRegistry`) with four new entries. Although the change is additive and follows the existing `sum` registration pattern, each registration requires constructing the correct `DimUnits` closure with distinct dimensional unit semantics — this is not purely mechanical and requires understanding of the dimensional unit algebra.
- **Task 3** (decomposed from original Task 3) implements `resolveAggregatorFnCall` on `CPLAnalyzer` — a single new method with its own tests. It touches `analyze_aggregators.go` and its test file. Although it is the more focused half of the original coordinated change, it requires interacting with `FunctionRegistry`, concept value maps, and aggregation matrix builders, making it a cross-concern task that is **not small**.
- **Task 4** (decomposed from original Task 3) updates the postfix evaluator and CIL tree construction to use `resolveAggregatorFnCall` and wraps aggregate nodes in `CILBinaryOp` for arithmetic expressions. It modifies the same files as Task 3 with coordinated evaluator changes and requires integration testing that all existing Increment 2 aggregator tests continue to pass. **Not small** — tightly coupled evaluator changes across the postfix loop and CIL tree construction logic.

### Parallel Execution

None — all tasks have sequential dependencies.

Task 1 must complete first: the five CIL node struct types must exist before the `FunctionRegistry` can reference their `CILNodeType` strings. Task 2 must complete second: the registry must have all five function entries before `resolveAggregatorFnCall` (Task 3) can perform lookups. Task 3 must complete third: `resolveAggregatorFnCall` must be implemented and tested before the postfix evaluator (Task 4) can call it. Task 4 must complete last: it integrates the resolver into the evaluator and CIL tree construction pipeline.

### Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4
```

### Test-First Development

This project follows the Logical TDD Lifecycle. For every coding subtask, write tests before implementing the code. Each task below includes a "run tests and confirm they fail" verification step after test writing and before implementation — this confirms the tests are correct and fail for the expected reason (types or functions not yet defined). The implementation then makes the tests pass. Tests are not broken out as separate tasks; the test-and-implement cycle is integrated within each task.

### Task 1: Define typed CIL node structs — Not Started

Define five operation-specific CIL node struct types in the `analysis` package, each implementing the `cil.CILNode` interface. These replace the generic `CILAggregateNode` from Increment 2. The `CILAggregateSizeNode` is distinct in that it has no `SourceCIL` field — it produces a compile-time constant (the number of distinct concept values) rather than wrapping a CIL subtree.

**Files**: Five new files (`src/main/go/analysis/cil_aggregate_sum_node.go`, `cil_aggregate_count_node.go`, `cil_aggregate_size_node.go`, `cil_aggregate_inverse_sum_node.go`, `cil_aggregate_inverse_count_node.go`) plus one test file (`src/main/go/analysis/cil_aggregate_node_test.go`).

**Note on SourceCIL**: The `SourceCIL` field on each node holds `cil.CILNode` (the interface from the `cil` package). The `cilNode()` marker method is defined in the `cil` package — each node struct must implement it (typically as `func (n *TypeName) cilNode() {}`).

   a. Write tests in `src/main/go/analysis/cil_aggregate_node_test.go` that verify all five struct types implement the `cil.CILNode` interface — Not Started
   b. Write tests that verify each struct can be initialized with its expected fields and that `CILAggregateSizeNode` correctly stores a size value without a `SourceCIL` field — Not Started
   c. Run `go test ./src/main/go/analysis/ -run TestCILAggregate -count=1` and confirm tests fail (types do not exist yet) — Not Started
   d. Create `src/main/go/analysis/cil_aggregate_sum_node.go` with `CILAggregateSumNode` struct containing `SourceCIL cil.CILNode`, `DimName string`, `Matrix *AggregationMatrix`, and `SourceLocation SourceLocation`; implement `cilNode()` marker method — Not Started
   e. Create `src/main/go/analysis/cil_aggregate_count_node.go` with `CILAggregateCountNode` struct (same field pattern as `CILAggregateSumNode`) and `cilNode()` marker method — Not Started
   f. Create `src/main/go/analysis/cil_aggregate_size_node.go` with `CILAggregateSizeNode` struct containing `DimName string`, `Size int`, and `SourceLocation SourceLocation` — **no `SourceCIL` field**; implement `cilNode()` marker method — Not Started
   g. Create `src/main/go/analysis/cil_aggregate_inverse_sum_node.go` with `CILAggregateInverseSumNode` struct (same field pattern as `CILAggregateSumNode`) and `cilNode()` marker method — Not Started
   h. Create `src/main/go/analysis/cil_aggregate_inverse_count_node.go` with `CILAggregateInverseCountNode` struct (same field pattern) and `cilNode()` marker method — Not Started
   i. Run `go test ./src/main/go/analysis/ -run TestCILAggregate -count=1` and confirm all interface and initialization tests pass — Not Started
   j. Run `go build ./src/main/go/analysis/` and confirm the package compiles without errors — Not Started

#### Acceptance Criteria

1. Five new Go source files exist in `src/main/go/analysis/`, each defining one CIL node struct
2. All five struct types implement `cil.CILNode` (verified by compile-time check and test)
3. `CILAggregateSumNode` has `SourceCIL`, `DimName`, `Matrix`, `SourceLocation` fields
4. `CILAggregateCountNode` has `SourceCIL`, `DimName`, `Matrix`, `SourceLocation` fields
5. `CILAggregateSizeNode` has `DimName`, `Size`, `SourceLocation` fields and **no `SourceCIL` field**
6. `CILAggregateInverseSumNode` has `SourceCIL`, `DimName`, `Matrix`, `SourceLocation` fields
7. `CILAggregateInverseCountNode` has `SourceCIL`, `DimName`, `Matrix`, `SourceLocation` fields
8. All tests in `cil_aggregate_node_test.go` pass
9. `go build ./src/main/go/analysis/` succeeds

### Task 2: Extend FunctionRegistry with all functions — Not Started

Extend `NewAggregatorFunctionRegistry()` in `function_registry.go` to register all five aggregator functions (`count`, `size`, `1/sum`, `1/count`) in addition to the `sum` entry that already exists from Story256. Each registration maps the function name to the corresponding `CILNodeType` string (from Task 1) and provides a `DimUnits` closure that computes the function's dimensional unit effect.

**Files**: `src/main/go/analysis/function_registry.go` (modify), `src/main/go/analysis/function_registry_test.go` (modify or create).

**Note on file existence**: `function_registry.go` is created by Story256. If the file does not yet exist at the time this story is implemented, the task must first create the file with the `FunctionRegistry` struct, `AggregatorFunctionInfo` struct, and `NewAggregatorFunctionRegistry` function (as defined in the analysis phase ERO), then extend with the additional four functions. The existing `sum` registration must remain intact.

   a. Write tests in `src/main/go/analysis/function_registry_test.go` that look up each of the five function names in the registry — Not Started
   b. Write tests that verify each function's `CILNodeType` string matches the expected type name (e.g., `"sum"` → `"CILAggregateSumNode"`, `"count"` → `"CILAggregateCountNode"`, `"size"` → `"CILAggregateSizeNode"`, `"1/sum"` → `"CILAggregateInverseSumNode"`, `"1/count"` → `"CILAggregateInverseCountNode"`) — Not Started
   c. Write tests that verify each function's `DimUnits` closure produces the correct dimensional unit expression: `sum(C)`→`{Num:[C]}`, `count(C)`→`{Num:[C]}`, `size(C)`→`{}` (unitless), `1/sum(C)`→`{Den:[C]}`, `1/count(C)`→`{Den:[C]}` — Not Started
   d. Run `go test ./src/main/go/analysis/ -run TestFunctionRegistry -count=1` and confirm tests fail (entries not yet registered) — Not Started
   e. Modify `NewAggregatorFunctionRegistry()` to register `count` with `CILNodeType: "CILAggregateCountNode"` and `DimUnits` producing `{Numerator: [param]}` — Not Started
   f. Register `size` with `CILNodeType: "CILAggregateSizeNode"` and `DimUnits` producing `{}` (unitless — empty numerator and empty denominator) — Not Started
   g. Register `1/sum` with `CILNodeType: "CILAggregateInverseSumNode"` and `DimUnits` producing `{Denominator: [param]}` — Not Started
   h. Register `1/count` with `CILNodeType: "CILAggregateInverseCountNode"` and `DimUnits` producing `{Denominator: [param]}` — Not Started
   i. Run `go test ./src/main/go/analysis/ -run TestFunctionRegistry -count=1` and confirm all lookup, CILNodeType, and DimUnits tests pass — Not Started

#### Acceptance Criteria

1. `FunctionRegistry` maps all five function names (`sum`, `count`, `size`, `1/sum`, `1/count`) to `AggregatorFunctionInfo` entries
2. Each entry's `CILNodeType` string matches the corresponding struct type name from Task 1
3. `sum(C)` and `count(C)` `DimUnits` closures produce `{Numerator: [C]}`
4. `size(C)` `DimUnits` closure produces `{}` (unitless)
5. `1/sum(C)` and `1/count(C)` `DimUnits` closures produce `{Denominator: [C]}`
6. The existing `sum` entry from Story256 is preserved with `CILNodeType` updated to `"CILAggregateSumNode"` (from Increment 2's `"CILAggregateNode"`)
7. All function registry tests pass

### Task 3: Implement resolveAggregatorFnCall — Not Started

Implement `resolveAggregatorFnCall` on `CPLAnalyzer`, which takes a parser-level function call operand (`*OperandData`) and a source CIL subtree (`cil.CILNode`) and returns the appropriate typed CIL aggregate node based on registry lookup. This function is the core dispatch mechanism that replaces generic `CILAggregateNode` creation from Increment 2. It is independently testable — call it directly with the desired function name and verify the returned node's concrete type and fields.

**Files**: `src/main/go/analysis/analyze_aggregators.go` (modify — add `resolveAggregatorFnCall` method), `src/main/go/analysis/analyze_aggregators_test.go` (modify or create).

**Note on file existence**: `analyze_aggregators.go` is created by Story256. If the file does not yet exist, this task must first create the file with the `CPLAnalyzer` method placeholder and imports, then add `resolveAggregatorFnCall`. The test file may need to set up a `CPLAnalyzer` with a populated `FunctionRegistry` and concept maps, using test helper patterns from existing analysis tests.

**Note on `1/sum` and `1/count`**: These compound function names use `/` as part of the identifier. The parser already stores them as single identifier token strings in `OperandData.Name` — no special parsing is needed.

   a. Write tests that verify `resolveAggregatorFnCall` with `fnCall.Name = "sum"` and a concept source CIL returns a `*CILAggregateSumNode` — Not Started
   b. Write tests that verify the returned `CILAggregateSumNode` has the correct `DimName`, `Matrix` with correct dimensions, and `SourceCIL` pointing to the supplied concept reference — Not Started
   c. Write tests that verify `resolveAggregatorFnCall` produces the correct node type for each remaining function: `*CILAggregateCountNode` for `"count"`, `*CILAggregateSizeNode` for `"size"` (with correct `Size`), `*CILAggregateInverseSumNode` for `"1/sum"`, `*CILAggregateInverseCountNode` for `"1/count"` — Not Started
   d. Write tests that verify `CILAggregateSizeNode` has `Size` equal to the number of distinct values in the concept (from the concept's `ValueOrdering`) and has no `SourceCIL` field — Not Started
   e. Write error tests: `resolveAggregatorFnCall` with unknown function name (e.g., `"foo"`) returns a clear error message — Not Started
   f. Write error tests: `resolveAggregatorFnCall` with a function name not registered in the `FunctionRegistry` returns an error — Not Started
   g. Run `go test ./src/main/go/analysis/ -run TestResolveAggregatorFnCall -count=1` and confirm tests fail (function not yet implemented) — Not Started
   h. Implement `resolveAggregatorFnCall(fnCall *OperandData, sourceCIL cil.CILNode) (cil.CILNode, error)` on `CPLAnalyzer` — Not Started
   i. Look up `fnCall.Name` in `a.functionRegistry.Functions`; return `fmt.Errorf("unknown aggregator function %q", fnCall.Name)` if not found — Not Started
   j. Extract `dimName` from `fnCall.Params[0]` (the concept name passed to the function call) — Not Started
   k. Switch on `fnInfo.CILNodeType`: `"CILAggregateSumNode"` → construct matrix via `BuildSumAggregationMatrix` and return `&CILAggregateSumNode{SourceCIL, DimName, Matrix, SourceLocation}`; `"CILAggregateCountNode"` → same matrix pattern, return `&CILAggregateCountNode{...}`; `"CILAggregateSizeNode"` → set `Size` from `len(conceptValues)`, return `&CILAggregateSizeNode{...}` (no SourceCIL); `"CILAggregateInverseSumNode"` → construct matrix via `BuildInverseSumAggregationMatrix`, return `&CILAggregateInverseSumNode{...}`; `"CILAggregateInverseCountNode"` → same matrix pattern, return `&CILAggregateInverseCountNode{...}` — Not Started
   l. Run `go test ./src/main/go/analysis/ -run TestResolveAggregatorFnCall -count=1` and confirm all positive and error tests pass — Not Started

#### Acceptance Criteria

1. `resolveAggregatorFnCall("sum", sourceCIL)` returns `*CILAggregateSumNode` with correct `DimName`, `SourceCIL`, and `Matrix`
2. `resolveAggregatorFnCall("count", sourceCIL)` returns `*CILAggregateCountNode` with correct fields
3. `resolveAggregatorFnCall("size", sourceCIL)` returns `*CILAggregateSizeNode` with `Size` equal to the number of concept values and no `SourceCIL` field
4. `resolveAggregatorFnCall("1/sum", sourceCIL)` returns `*CILAggregateInverseSumNode` with `Matrix` entries equal to `1/N`
5. `resolveAggregatorFnCall("1/count", sourceCIL)` returns `*CILAggregateInverseCountNode` with `Matrix` entries equal to `1/N`
6. `resolveAggregatorFnCall("foo", sourceCIL)` returns a clear error: `"unknown aggregator function \"foo\""`
7. All tests in `analyze_aggregators_test.go` tagged with `TestResolveAggregatorFnCall` pass
8. `go build ./src/main/go/analysis/` succeeds

### Task 4: Update postfix evaluator and CIL tree construction — Not Started

Update the expression postfix evaluator in the aggregator analysis flow (`AnalyzeAggregatorScope` or its equivalent) to call `resolveAggregatorFnCall` (Task 3) when it encounters `fncall` operands, replacing the previous generic `CILAggregateNode` creation from Increment 2. Update the CIL tree construction so that arithmetic operators (`+`, `-`, `*`, `/`) wrap their operand CIL nodes (including typed aggregate nodes) in `CILBinaryOp` nodes, producing correct expression CIL trees.

**Files**: `src/main/go/analysis/analyze_aggregators.go` (modify — update postfix evaluator loop and CIL tree construction logic), `src/main/go/analysis/analyze_aggregators_test.go` (modify — add expression tree tests and regression tests).

**Note**: This task modifies the same files as Task 3 but changes different code paths: the evaluator call site (where `fncall` operands are handled) and the CIL tree construction logic (where `CILBinaryOp` wrappers are created). Since Task 4 depends on Task 3's `resolveAggregatorFnCall` implementation, these tasks must execute sequentially — Task 3's worktree changes must be merged before Task 4 begins.

   a. Write tests that verify expression CIL tree construction: the postfix evaluator processes `(a-b)*sum(Country)` and produces `CILBinaryOp("*", CILBinaryOp("-", ref_a, ref_b), CILAggregateSumNode{source, "Country", matrix})` — Not Started
   b. Write tests that verify arithmetic operators in aggregator expressions correctly wrap aggregate nodes: `sum(Country) + count(Product)` produces `CILBinaryOp("+", CILAggregateSumNode{...}, CILAggregateCountNode{...})` — Not Started
   c. Write tests that verify the postfix evaluator correctly handles nested expressions containing aggregate function calls (e.g., `(sum(A) * count(B)) / size(C)`) — Not Started
   d. Run `go test ./src/main/go/analysis/ -run TestAggregatorExpressionCILTree -count=1` and confirm tests fail (evaluator not yet updated) — Not Started
   e. Update the postfix evaluator loop in the aggregator scope handler so that when it encounters `OperandData` with `Kind == "fncall"`, it calls `a.resolveAggregatorFnCall(operand, sourceCIL)` instead of creating a generic `CILAggregateNode` — Not Started
   f. Update the CIL tree construction so that when the evaluator processes arithmetic operators (`+`, `-`, `*`, `/`), it pops operand CIL nodes from the CIL stack and wraps them in `CILBinaryOp` nodes — Not Started
   g. Ensure that the CIL stack holds `cil.CILNode` values (not dimensional units), and that the expression evaluation correctly builds the tree bottom-up — Not Started
   h. Run `go test ./src/main/go/analysis/ -run TestAggregatorExpressionCILTree -count=1` and confirm all expression tree tests pass — Not Started
   i. Run `go test ./src/main/go/analysis/ -run TestAggregator -count=1` and confirm existing Increment 2 aggregator tests continue to pass with no regressions — Not Started

#### Acceptance Criteria

1. The postfix evaluator calls `resolveAggregatorFnCall` for `fncall` operands instead of creating generic `CILAggregateNode`
2. `(a-b)*sum(Country)` produces `CILBinaryOp("*", CILBinaryOp("-", ref_a, ref_b), CILAggregateSumNode{...})`
3. `sum(Country) + count(Product)` produces `CILBinaryOp("+", CILAggregateSumNode{...}, CILAggregateCountNode{...})`
4. Nested expressions containing aggregate function calls are correctly handled
5. The generic `CILAggregateNode` is no longer created during analysis (the struct may remain defined)
6. All new expression tree tests pass
7. All existing Increment 2 aggregator tests continue to pass with no regressions
8. `go test ./src/main/go/analysis/...` passes with no new failures

## Constraints

1. **Increment 2 prerequisite**: This story assumes that the following types and functions from Story256 exist: `FunctionRegistry`, `AggregatorFunctionInfo`, `DimensionalUnitExpression`, `AggregationMatrix`, `BuildSumAggregationMatrix`, `BuildInverseSumAggregationMatrix`, `OperandData`, `SourceLocation`, and the `CPLAnalyzer` struct with its `functionRegistry` field. If any of these are missing, they must be created as defined in [ERO007: Aggregators Analysis Phase](specifications/design/cpl/go-analysis/eros/ERO007_aggregators-analysis-phase.md).
2. **CILNode interface**: All CIL node types must implement `cil.CILNode` from the `cil` package (`src/main/go/cil/cil_nodes.go`), which requires the `cilNode()` marker method.
3. **Function names with `/`**: `1/sum` and `1/count` use `/` as part of a compound identifier. The parser already handles these as single identifier tokens — no parser changes are needed.
4. **size(C) semantics**: `CILAggregateSizeNode` has no `SourceCIL` field because it produces a compile-time constant (the number of distinct values in the concept), not a value derived from tensor data.
5. **Expression tree correctness**: Aggregate nodes are expression operands — they produce values that participate in arithmetic operations via `CILBinaryOp` wrappers. The CIL tree structure must be correct so that the synthesis phase (Story259) can generate code for each node type individually.
6. **Test isolation**: Tests for resolveAggregatorFnCall may need to set up a `CPLAnalyzer` with a populated `FunctionRegistry` and concept maps. Use the test helper patterns from existing analysis tests.

## Intent

The purpose of this story is to move from a single generic aggregator node (`CILAggregateNode` from Increment 2) to operation-specific typed nodes. Each aggregator function has different semantics — `sum` and `count` use a standard aggregation matrix, `size` produces a scalar constant, and `1/sum`/`1/count` use inverse-weight matrices. Typed nodes allow the synthesis phase to generate function-specific code and make the CIL tree self-documenting. The `FunctionRegistry` acts as the central dispatch mechanism: given a function name from the parser, it returns the CIL node type to instantiate and the dimensional unit effect to validate. The `resolveAggregatorFnCall` function bridges the registry to the postfix evaluator, creating the correct typed node during expression analysis.

## Notes

- The generic `CILAggregateNode` struct from Incement 2 may remain defined in the codebase but must not be created by the updated aggregator analysis code.
- Increment 2's `FunctionRegistry` initially maps `sum` to `"CILAggregateNode"`. Task 2 of this story updates that entry to `"CILAggregateSumNode"` — this is the only change to the existing `sum` registration.
- The `CILBinaryOp` struct already exists in the `cil` package and is used for metric expressions. This story extends its usage to aggregator expressions following the same pattern.
