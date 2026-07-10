# Story259: Aggregators Increment 3 — Synthesize (Epic007)

## Goal

Implement CIL tree walking for all five operation-specific CIL aggregator node types (`CILAggregateSumNode`, `CILAggregateCountNode`, `CILAggregateSizeNode`, `CILAggregateInverseSumNode`, `CILAggregateInverseCountNode`), emit `tensorops.ReduceDimension` calls for matrix-based nodes and `float64(N)` constants for `size(C)`, integrate aggregator nodes into expression CIL trees, and create end-to-end aggregator examples with CPL + parquet + expected output. This is Increment 3 of Epic007 (Aggregators), delivering the synthesis phase for typed aggregator nodes defined in Story258.

## References

- [Plan: Story259](../../memory-bank/plans/Story259_aggregators-increment3-synthesize.md)
- [ERO007: Aggregators Incremental Implementation](../specifications/design/cpl/eros/ERO007_aggregators-incremental-implementation.md)
- [ERO007: Aggregators Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO007_aggregators-synthesis-phase.md) — Walk functions for each CIL node type, size constant, expression code generation
- [Aggregators Design](../specifications/design/cpl/aggregators.md) — Core aggregator design
- [Postfix Expression Parsing Pattern](../specifications/patterns/postfix-expression-parsing.md) — Expression handling pattern
- [Epic007: CPL Aggregators Implementation](../epics/Epic007_cpl-aggregators-implementation.md)

## Dependencies

- **Story258** (`aggregators-increment3-analyze`) — _Not Started_: Must complete first. Provides the five typed CIL node types (`CILAggregateSumNode`, `CILAggregateCountNode`, `CILAggregateSizeNode`, `CILAggregateInverseSumNode`, `CILAggregateInverseCountNode`), the `FunctionRegistry` with all five function entries, `resolveAggregatorFnCall` on `CPLAnalyzer`, and expression CIL tree construction with `CILBinaryOp` wrappers.
- **Story257** (`aggregators-increment2-synthesize`) — _Not Started_: Must complete first. Provides the `CILAbstractTransformer` walker infrastructure, the `WalkCILAggregateNode` pattern for walking aggregator nodes, the `CILCodeGenerator.GenerateAggregationMatrix` method for emission of aggregation matrix Go code, and synthesis preamble integration for aggregation matrix variables.
- **Story254** (`aggregators-increment1-matrix-mult`) — _Not Started_: Provides `tensorops.ReduceDimension` and `tensorops.AggregationMatrix` for code emission.

## Dependent Stories

- **Story260** (`aggregators-increment4-analyze`) — _Not Started_: Depends on typed CIL node synthesis being functional for the dimension-mismatch detection and aggregator insertion analysis.
- **Story261** (`aggregators-increment4-synthesize`) — _Not Started_: Depends on full synthesis pipeline supporting all five aggregator functions.

## Tasks

### Task Sizing

- **Task 1** is not `[Small]` — it adds two new walker methods (`WalkCILAggregateSumNode`, `WalkCILAggregateCountNode`) and wires two dispatch cases. Although the methods follow the existing `WalkCILAggregateNode` pattern from Story257, each requires reasoning about `tensorops.ReduceDimension` code emission, dimension index computation, error handling generation, and variable registration. The walker methods, dispatch wiring, and tests form a single cohesive unit in 2 files.
- **Task 2** is `[Small]` — it adds a single walker method (`WalkCILAggregateSizeNode`) that emits a simple constant `float64(size)`. No `ReduceDimension` call, no tensor source, no dimension index computation. One concern in 2 files (cil_transformer.go + test file). Straightforward and additive.
- **Task 3** is not `[Small]` — it adds two walker methods (`WalkCILAggregateInverseSumNode`, `WalkCILAggregateInverseCountNode`) that follow the same `ReduceDimension` pattern as Task 1 but with inverse matrix naming conventions. Requires coordination with the aggregation matrix naming scheme (`aggregationMatrix_invsum_X`, `aggregationMatrix_invcount_X`). 2 files but multi-method reasoning.
- **Task 4** is `[Small]` — test-only task. Adds test functions to `cil_transformer_aggregator_test.go` that verify the existing walker infrastructure handles aggregate nodes embedded in `CILBinaryOp` expression trees. No new source code. 1 file. One concern: expression integration verification.
- **Task 5a** is `[Small]` — creates two `parse_and_analyze_only` CPL example files in separate error/without-error directories. Purely additive, follows existing example patterns. 2 files. One concern: parser/analyzer-level aggregator function examples.
- **Task 5b** is not `[Small]` — creates a full `with_data/aggregators_functions/` example directory with CPL authoring, IO spec, parquet data generation, ground-truth xlsx computation, expression example (`(a-b)*sum(Country)`), and end-to-end verification. Multiple coordinated file types (CPL, JSON, parquet, xlsx, Go test) that must be internally consistent.

**Decomposition note**: The plan's original Task 5 (7 subtasks) has been decomposed into Task 5a (parse_and_analyze_only examples) and Task 5b (with_data examples). The parse_and_analyze_only examples are a self-contained preamble — they go to a different directory (`examples/parse_and_analyze_only/`), require no synthesis pipeline, and are independently testable. The with_data examples require the full pipeline from Tasks 1–4.

Tasks 1, 3, and 5b route to `code-for-story-implementor`. Tasks 2, 4, and 5a route to `small-code-for-story-implementor`.

### Parallel Execution

- **Group A** (Task 4, Task 5a): Task 4 adds expression integration tests to `src/main/go/cil/cil_transformer_aggregator_test.go`; Task 5a creates `examples/parse_and_analyze_only/` CPL files. Completely file-disjoint (separate directories, no shared files). Task 5a only needs the parse/analyze infrastructure from Story258 — no dependency on Tasks 1–3 walker methods. Each task can be independently tested and verified.

Tasks 1, 2, and 3 all modify the same file (`cil_transformer.go`) and the same test file (`cil_transformer_aggregator_test.go`). Per the architect's parallelism rules, shared file modification prevents parallel worktree execution — these must execute sequentially.

### Execution Order

```
Task 1 → Task 2 → Task 3 → [Task 4 ‖ Task 5a] → Task 5b
```

- Tasks 1, 2, 3 must be sequential: all modify `cil_transformer.go` (add walker methods, add dispatch cases). Task 2 depends on Task 1 establishing the dispatch framework; Task 3 depends on Tasks 1 and 2 establishing the walk pattern.
- Task 4 depends on Tasks 1–3: needs all five walker methods present before writing expression integration tests.
- Task 5a can run in parallel with Task 4: file-disjoint, no semantic dependency on walker methods.
- Task 5b depends on Tasks 1–4: with_data example exercises the full synthesis pipeline with all five walker methods and expression integration.

### Reintegration

**Group A (Tasks 4, 5a):**
- Merge order: any order (fully disjoint directories — `src/main/go/cil/` vs `examples/parse_and_analyze_only/` — no shared infrastructure)
- Integration tests: `go test ./examples/parse_and_analyze_only/... -count=1 -timeout 60s` — verify both CPL example files parse and analyze correctly after merge
- Watch for conflicts in: none expected (disjoint directories, no shared test files, no shared build config)

### Test-First Development

This story follows the Logical TDD Lifecycle. For each coding task, write tests before implementing the code. The test-and-implement cycle is integrated within each task in the correct order:

- **Task 1**: Write tests for `WalkCILAggregateSumNode` and `WalkCILAggregateCountNode` (subtasks c–d) before the walker implementation passes them.
- **Task 2**: Write tests for `WalkCILAggregateSizeNode` constant emission (subtask c) before implementation.
- **Task 3**: Write tests for `WalkCILAggregateInverseSumNode` and `WalkCILAggregateInverseCountNode` (subtask c) before implementation.
- **Task 4**: Write expression integration tests (subtask a) — these fail until Tasks 1–3 are complete and merged.
- **Task 5a**: Create CPL files and verify they parse/analyze correctly.
- **Task 5b**: Create example data and CPL files; verify full pipeline executes correctly.

Tests are NOT broken out as separate tasks. See `memory-bank/concepts.md` for details on the Logical TDD Lifecycle.

### Constraints

- The five typed CIL node types (`CILAggregateSumNode`, `CILAggregateCountNode`, `CILAggregateSizeNode`, `CILAggregateInverseSumNode`, `CILAggregateInverseCountNode`) must exist from Story258
- The `CILAbstractTransformer` walker infrastructure and `CILCodeGenerator.GenerateAggregationMatrix` must exist from Story257
- `tensorops.ReduceDimension` and `tensorops.AggregationMatrix` must exist from Story254
- `CILAggregateSizeNode` has no `SourceCIL` field — it produces a compile-time constant from `node.Size`
- All generated code must compile under `source_root` (`src/chronocone_planning_language`)
- Generated code follows existing patterns: `check()` helper or `if err != nil` for error handling, `ChronoconeTensor[float64]` for tensor types
- Aggregator nodes are expression operands — they produce values that participate in arithmetic via `CILBinaryOp` wrappers
- Matrix-based nodes use matrix names of the form `aggregationMatrix_sum_X`, `aggregationMatrix_count_X`, `aggregationMatrix_invsum_X`, `aggregationMatrix_invcount_X`
- `size(C)` emits a float64 constant — no matrix, no tensor operation, no error handling

### Task 1: Add walk methods for CILAggregateSumNode and CILAggregateCountNode — Not Started

Add `WalkCILAggregateSumNode` and `WalkCILAggregateCountNode` methods to `CILAbstractTransformer` in `src/main/go/cil/cil_transformer.go`. Both methods follow the pattern established by `WalkCILAggregateNode` in Story257: walk the source CIL subtree, determine the collapsed dimension index, emit `tensorops.ReduceDimension(sourceVar, aggregationMatrix_<func>_<dim>, collapsedDim, targetDim)` with error handling, and push the result variable as an expression operand. Wire both node types into the main walker dispatch.

   a. Write tests in `src/main/go/cil/cil_transformer_aggregator_test.go` for `CILAggregateSumNode` walking: construct a `CILAggregateSumNode` with a known source CIL subtree, walk it, verify generated code contains `tensorops.ReduceDimension(sourceVar, aggregationMatrix_sum_<dim>, <collapsedDim>, <targetDim>)` — Not Started
   b. Write tests for `CILAggregateCountNode` walking: same pattern but with `aggregationMatrix_count_<dim>` — Not Started
   c. Write tests that verify error handling code emission: `if err != nil { return nil, fmt.Errorf("sum(<dim>) failed: %w", err) }` and `fmt.Errorf("count(<dim>) failed: %w", err)` — Not Started
   d. Run `go test ./src/main/go/cil/... -run Aggregator -count=1` and confirm tests fail (methods do not exist yet) — Not Started
   e. Add `WalkCILAggregateSumNode(node *CILAggregateSumNode, scope *WalkScope) error` to `CILAbstractTransformer`: walk `node.SourceCIL`, determine collapsed dim index, emit `ReduceDimension` with matrix name `aggregationMatrix_sum_<dim>`, emit error check, push result operand — Not Started
   f. Add `WalkCILAggregateCountNode(node *CILAggregateCountNode, scope *WalkScope) error`: same pattern with matrix name `aggregationMatrix_count_<dim>` — Not Started
   g. Wire `case *CILAggregateSumNode: return t.WalkCILAggregateSumNode(node, scope)` and `case *CILAggregateCountNode: return t.WalkCILAggregateCountNode(node, scope)` into the walker dispatch type-switch — Not Started
   h. Run `go test ./src/main/go/cil/... -run Aggregator -count=1` and confirm all tests pass — Not Started

**Files**: `src/main/go/cil/cil_transformer.go` (modify — add 2 methods + 2 dispatch cases), `src/main/go/cil/cil_transformer_aggregator_test.go` (modify — add tests)

**Intent**: The walker must transform typed CIL aggregate nodes into executable Go code. Each `WalkCILAggregate<Func>Node` method follows the existing `WalkCILAggregateNode` pattern from Story257 but uses the correct matrix name and error message for the function type. The dispatch switch maps each concrete CIL node type to its walker method.

**Acceptance criteria**:
- `WalkCILAggregateSumNode` emits `tensorops.ReduceDimension(source, aggregationMatrix_sum_<dim>, collapsedDim, targetDim)` with error handling
- `WalkCILAggregateCountNode` emits `tensorops.ReduceDimension(source, aggregationMatrix_count_<dim>, collapsedDim, targetDim)` with error handling
- Both methods walk the source CIL subtree to obtain the source variable name
- Collapsed dimension index is correctly determined from the source CIL subtree's dimensions
- Error handling code is emitted: `if err != nil { return nil, fmt.Errorf("sum(<dim>) failed: %w", err) }`
- Result variable is pushed as an expression operand via `scope.pushOperand`
- Dispatch switch cases compile correctly

### Task 2: Add walk method for CILAggregateSizeNode (constant emission) — Not Started [Small]

Add `WalkCILAggregateSizeNode` to `CILAbstractTransformer` in `src/main/go/cil/cil_transformer.go`. Unlike the other aggregator walk methods, `size(C)` emits a compile-time constant — no tensor source, no `ReduceDimension` call, no error handling. The method emits `size_<dim> := float64(<Size>)` and pushes the result as an expression operand.

   a. Write tests in `src/main/go/cil/cil_transformer_aggregator_test.go`: construct a `CILAggregateSizeNode{DimName: "Country", Size: 3}` and verify emitted code contains `size_Country := float64(3)` — Not Started
   b. Write tests that verify no `tensorops.ReduceDimension` call is emitted (size is a constant, not a tensor operation) and no error handling `if err != nil` block is generated — Not Started
   c. Run `go test ./src/main/go/cil/... -run Aggregator -count=1` and confirm size tests fail — Not Started
   d. Add `WalkCILAggregateSizeNode(node *CILAggregateSizeNode, scope *WalkScope) error` to `CILAbstractTransformer`: emit `resultVar := float64(node.Size)  // size(<dim>)` and push as operand — Not Started
   e. Wire `case *CILAggregateSizeNode: return t.WalkCILAggregateSizeNode(node, scope)` into the walker dispatch type-switch — Not Started
   f. Run `go test ./src/main/go/cil/... -run Aggregator -count=1` and confirm tests pass — Not Started

**Files**: `src/main/go/cil/cil_transformer.go` (modify — add 1 method + 1 dispatch case), `src/main/go/cil/cil_transformer_aggregator_test.go` (modify — add tests)

**Intent**: `size(C)` is a compile-time constant that counts the number of distinct values in concept `C`. No tensor source exists — `CILAggregateSizeNode` carries the size directly. The walker emits a simple `float64(N)` assignment. No `ReduceDimension` call, no matrix lookup, no error handling.

**Acceptance criteria**:
- `WalkCILAggregateSizeNode` emits `size_<dim> := float64(<Size>)  // size(<dim>)`
- No `tensorops.ReduceDimension` call is emitted
- No `if err != nil` error handling block is emitted
- Result variable is pushed as an expression operand
- Dispatch switch case compiles correctly

### Task 3: Add walk methods for CILAggregateInverseSumNode and CILAggregateInverseCountNode — Not Started

Add `WalkCILAggregateInverseSumNode` and `WalkCILAggregateInverseCountNode` methods to `CILAbstractTransformer` in `src/main/go/cil/cil_transformer.go`. Both methods follow the same `ReduceDimension` pattern as Task 1 but reference inverse aggregation matrices (`aggregationMatrix_invsum_<dim>` and `aggregationMatrix_invcount_<dim>`). Wire both node types into the main walker dispatch.

   a. Write tests in `src/main/go/cil/cil_transformer_aggregator_test.go` for `CILAggregateInverseSumNode`: construct with source CIL, walk, verify `ReduceDimension` call uses matrix name `aggregationMatrix_invsum_<dim>` — Not Started
   b. Write tests for `CILAggregateInverseCountNode`: verify `ReduceDimension` call uses matrix name `aggregationMatrix_invcount_<dim>` — Not Started
   c. Run `go test ./src/main/go/cil/... -run Aggregator -count=1` and confirm inverse tests fail — Not Started
   d. Add `WalkCILAggregateInverseSumNode(node *CILAggregateInverseSumNode, scope *WalkScope) error`: same pattern as sum node with matrix name `aggregationMatrix_invsum_<dim>` and error message `fmt.Errorf("1/sum(<dim>) failed: %w", err)` — Not Started
   e. Add `WalkCILAggregateInverseCountNode(node *CILAggregateInverseCountNode, scope *WalkScope) error`: same pattern with matrix name `aggregationMatrix_invcount_<dim>` and error message `fmt.Errorf("1/count(<dim>) failed: %w", err)` — Not Started
   f. Wire dispatch cases for both node types into the walker type-switch — Not Started
   g. Run `go test ./src/main/go/cil/... -run Aggregator -count=1` and confirm all tests pass — Not Started

**Files**: `src/main/go/cil/cil_transformer.go` (modify — add 2 methods + 2 dispatch cases), `src/main/go/cil/cil_transformer_aggregator_test.go` (modify — add tests)

**Intent**: Inverse aggregator functions (`1/sum`, `1/count`) use inverse-weight matrices where each entry is `1/N` (for N source concept values). The walker pattern is identical to sum/count but references different matrix variable names and uses different error messages. The analysis phase pre-builds inverse matrices via `BuildInverseSumAggregationMatrix` — the walker just references them.

**Acceptance criteria**:
- `WalkCILAggregateInverseSumNode` emits `ReduceDimension` with `aggregationMatrix_invsum_<dim>`
- `WalkCILAggregateInverseCountNode` emits `ReduceDimension` with `aggregationMatrix_invcount_<dim>`
- Error messages reference `1/sum` and `1/count` function names
- Both methods follow the same source-walking, dimension-index, error-handling pattern as sum/count
- Dispatch switch cases compile correctly

### Task 4: Expression integration test — Not Started [Small]

Write expression integration tests in `src/main/go/cil/cil_transformer_aggregator_test.go` that verify the CIL tree walker correctly handles aggregate nodes embedded in arithmetic expression trees. The tests construct CIL trees where aggregate nodes appear as operands in `CILBinaryOp` expressions, walk the tree, and verify the generated code computes the aggregate result as a local variable and then uses it in the arithmetic operation.

   a. Write test: construct `CILBinaryOp("*", CILBinaryOp("-", ref_a, ref_b), CILAggregateSumNode{SourceCIL: source_cil, DimName: "Country", Matrix: sumMatrix})`, walk the tree, and verify generated code: (1) computes `a - b` as a local variable, (2) emits `ReduceDimension` for sum with result variable `sum_Country`, (3) multiplies the two results — Not Started
   b. Write test: verify variable flow — source variable → aggregated variable → expression operand. Confirm the aggregate result variable appears as an operand in the subsequent `*` operation — Not Started
   c. Run `go test ./src/main/go/cil/... -run Aggregator -count=1` and confirm all tests pass (this is an integration verification — all walker methods from Tasks 1–3 must be present) — Not Started

**Files**: `src/main/go/cil/cil_transformer_aggregator_test.go` (modify — add test functions)

**Intent**: Expression integration is the culmination of the walker infrastructure. Aggregate nodes are expression operands — they produce values consumed by arithmetic operations. The walker naturally handles this: `CILBinaryOp("*")` walks its left and right subtrees, the right subtree (an aggregate node) emits `ReduceDimension` and pushes the result variable, and the `*` operator combines the operands. This task validates the entire composition without adding new walker code.

**Acceptance criteria**:
- CIL tree `(a-b)*sum(Country)` produces generated code that:
  1. Computes `a - b` as a local variable (`sub_<n>`)
  2. Emits `ReduceDimension` for `sum(Country)` with result variable `sum_Country`
  3. Multiplies `sub_<n>` and `sum_Country`
- Aggregate result variable flows correctly from `ReduceDimension` output to expression operand
- No new walker methods are added — this task validates the existing infrastructure

### Task 5a: Create parse_and_analyze_only aggregator CPL examples — Not Started [Small]

Create CPL example files in the `examples/parse_and_analyze_only/` directories that test the parser and analyzer's handling of aggregator function names. These examples validate that `count`, `size`, `1/sum`, and `1/count` are recognized as valid aggregator functions, and that unknown function names produce appropriate errors.

   a. Create `examples/parse_and_analyze_only/without_errors/aggregators_functions.cpl` with CPL definitions containing aggregator declarations using `count`, `size`, `1/sum`, and `1/count` — Not Started
   b. Create `examples/parse_and_analyze_only/with_errors/aggregators_unknown_function.cpl` with an aggregator declaration using an unknown function name (e.g., `foo(Country)`) — Not Started
   c. Verify the without_errors CPL parses and analyzes without errors: `go test ./examples/parse_and_analyze_only/... -count=1 -timeout 30s` — Not Started
   d. Verify the with_errors CPL produces the expected error for unknown function name — Not Started

**Files**: `examples/parse_and_analyze_only/without_errors/aggregators_functions.cpl` (create), `examples/parse_and_analyze_only/with_errors/aggregators_unknown_function.cpl` (create)

**Intent**: These CPL examples serve as fast, lightweight validation that the parser and analyzer correctly handle all five aggregator function names. They exercise the `FunctionRegistry` lookup path and the `resolveAggregatorFnCall` dispatch from Story258 without requiring the full synthesis pipeline. The parse_and_analyze_only framework runs parse → analyze and checks for expected errors/passes.

**Acceptance criteria**:
- `aggregators_functions.cpl` parses and analyzes without errors using all four non-sum function names
- `aggregators_unknown_function.cpl` produces a clear error for the unknown function name
- Both examples follow existing `parse_and_analyze_only` conventions (file structure, naming)

### Task 5b: Create with_data aggregator functions example — Not Started

Create a full `examples/with_data/aggregators_functions/` example directory that validates the complete aggregator pipeline for all five function types: CPL → parse → analyze → CIL → synthesize → execute → verify. Include expression integration: an example where `(a-b)*sum(Country)` combines arithmetic with aggregation. Create parquet test data with known values, hand-computed expected output xlsx, and end-to-end tests.

   a. Create `examples/with_data/aggregators_functions/` directory with standard category structure — Not Started
   b. Write CPL file `core.cpl` with concepts (e.g., Country, Product), a connector referencing source data, and aggregator declarations: `sum` aggregation across Country, `count` aggregation across Product, `size` extraction, `1/sum` and `1/count` inverse aggregations — Not Started
   c. Write IO specification JSON defining input parquet connectors and output specs for each aggregated result — Not Started
   d. Generate parquet test data with known numeric values (e.g., Country={USA,Canada} × Product={widget,gadget} with 4 cell values) — Not Started
   e. Create `.expected.xlsx` with hand-computed ground-truth values for each aggregator function: sum across Country, count across Country, size(Country)=2, 1/sum weights, 1/count weights — Not Started
   f. Create expression sub-example: `(a-b)*sum(Country)` — define two connectors (a with Country×Product and b with Country×Product), write aggregator that sums across Country, write CPL that computes `(a-b)*sum(Country)`, generate parquet data, create .expected.xlsx — Not Started
   g. Write end-to-end test that loads parquet data, runs the full pipeline for each aggregator, and verifies output against expected xlsx: `go test ./examples/with_data/aggregators_functions/... -count=1 -timeout 120s` — Not Started

**Files**: `examples/with_data/aggregators_functions/core.cpl` (create), `examples/with_data/aggregators_functions/aggregators_functions/` subdirectory with .cpl, io_spec.json, .parquet data files, `.expected.xlsx` (create), `examples/with_data/aggregators_functions/aggregators_functions_test.go` (create)

**Intent**: This example provides the first end-to-end validation of all five aggregator function types in the full CPL pipeline. The expression sub-example validates that aggregator results compose with arithmetic operations. The manual test pattern follows existing `with_data/` example conventions. This example supersedes the Increment 2 `with_data/aggregators_basic/` example by covering all function types.

**Acceptance criteria**:
- CPL parses and analyzes without errors for all five aggregator function types
- Synthesized program compiles and executes
- Sum aggregation produces correct results matching ground-truth xlsx
- Count aggregation produces correct results
- Size produces correct compile-time constant (not a tensor operation)
- Inverse sum and inverse count produce correct weighted results
- Expression example `(a-b)*sum(Country)` produces correct combined result
- All end-to-end tests pass

## Notes

- The five walker methods replace the generic `WalkCILAggregateNode` from Story257 — the generic method may remain defined but should not be the primary dispatch for the new typed nodes
- `CILAggregateSizeNode` is the only node type with no `SourceCIL` field — the walker accesses `node.Size` directly
- Matrix names follow the convention `aggregationMatrix_<func>_<dim>`: `sum`, `count`, `invsum`, `invcount` — the `_` suffix separates function from dimension
- `size(C)` uses `size` prefix but with dimension name, not a matrix name: `size_<dim> := float64(N)`
- The dispatch type-switch (or dispatch mechanism established in Story257) must handle all five `case` branches
- File paths in all task descriptions are relative to `source_root` (`src/chronocone_planning_language`)
- Go test commands use fully-qualified package paths from `source_root`
