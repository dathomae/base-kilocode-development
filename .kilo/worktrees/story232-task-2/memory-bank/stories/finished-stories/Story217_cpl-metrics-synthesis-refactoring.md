# Story217: CPL Metrics Synthesis Module Refactoring

## Status: Done

## Goal

Refactor the large monolithic Go source files in the CPL synthesis `calculation` and `cil` packages into focused, per-concern files. This reduces per-file line count from 1,000–1,500 lines to 300–600 lines, enabling limited-context LLMs in Code mode to read and edit only the specific file relevant to their task. No API changes, no behavior changes — pure file splits within the same Go packages.

## References

- [Story188: CPL Metrics Conditionals Synthesize](Story188_cpl-metrics-conditionals-synthesize.md)
- [Tensor Representation Design](../specifications/design/cpl/go-synthesis-target/tensor-representation.md)

## Dependencies

None. This is a prerequisite refactoring that blocks Story188's remaining tasks.

## Dependent Stories

- [Story188: CPL Metrics Conditionals Synthesize](Story188_cpl-metrics-conditionals-synthesize.md) — dependent on this refactoring to enable limited-context LLM implementation of remaining tasks (Tasks 3–9)

## Constraints

1. **Zero behavior change** — refactoring must not change any function signatures, logic, or test behavior
2. **Same Go package** — split files remain in `package calculation` and `package cil` respectively; no import changes needed
3. **All existing tests must pass** after each split
4. **No new dependencies** — no new packages, no circular imports
5. Focused test files created for Story188 must remain empty scaffolding that existing test suites ignore (no test functions yet)
6. Line count target: no non-test file exceeds 650 lines after refactoring; no single test file exceeds 1,000 lines

## Tasks

### Task 1: Split chronocone_tensor.go into focused files — Completed

**Goal**: Move arithmetic and comparison functions out of the 1,465-line `chronocone_tensor.go` into dedicated files, reducing `chronocone_tensor.go` to ~530 lines (core struct, controller, allocation, MetricComputation, mask infrastructure).

**Files involved**:
- `src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go` (shrink)
- `src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor_arithmetic.go` (new)
- `src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor_comparison.go` (new)

**Phase A — Plan the boundaries:**

a. Identify the exact line ranges to move:
   - Arithmetic functions: `TensorAddFloat64` through `TensorDivideScalarFloat64` (lines ~526-930). These functions implement `+`, `-`, `*`, `/`, `negate`, and scalar variants with mask propagation.
   - Comparison functions: `TensorGreaterThanFloat64Mask` through `TensorNotEqualScalarFloat64Mask` (lines ~931-1465). These functions implement `>`, `>=`, `<`, `<=`, `==`, `!=`, and scalar variants.

b. What stays in `chronocone_tensor.go`: struct definitions, `ChronoconeComputationController`, `ChronoconeTensorComputationChunk`, `MetricComputation`, `AllocateTensor`, `NewChronoconeTensor`, `NewMetricComputation`, `NewChronoconeComputationController`, mask field + 5 mask methods, `TensorDimensionNamesMatch`.

**Phase B — Verify before refactoring:**

- Run `go test ./calculation/...` to confirm all existing tests pass before any changes
- Record the test count for regression verification after each split

**Phase C — Execute the split:**

a. Create `chronocone_tensor_arithmetic.go`:
   - Copy package declaration and copyright header from `chronocone_tensor.go`
   - Copy all arithmetic functions (Add through DivideScalar, Negate, all scalar variants)
   - Ensure imports are correct (same package, so same import scope)
   - Remove these functions from `chronocone_tensor.go`

b. Create `chronocone_tensor_comparison.go`:
   - Copy package declaration and copyright header
   - Copy all 12 comparison mask functions (GT through NotEqualScalar)
   - Ensure `math` import is present for `math.IsNaN`
   - Remove these functions from `chronocone_tensor.go`

c. Verify: `go test ./calculation/...` must pass with zero failures

### Task 2: Split cil_transformer.go into focused files — Completed

**Goal**: Move `walkToPostfix`/`WalkCILTreeToPostfix` and code generation helpers out of the 1,159-line `cil_transformer.go` into dedicated files. Reduces `cil_transformer.go` to ~600 lines.

**Files involved**:
- `src/chronocone_planning_language/src/main/go/cil/cil_transformer.go` (shrink)
- `src/chronocone_planning_language/src/main/go/cil/postfix_walk.go` (new)
- `src/chronocone_planning_language/src/main/go/cil/postfix_codegen.go` (new)

**Phase A — Plan the boundaries:**

a. Identify exact functions to move:
   - `postfix_walk.go`: `WalkCILTreeToPostfix` (line 573), `walkToPostfix` (line 602) — the CIL tree → postfix table walker methods
   - `postfix_codegen.go`: `GeneratePostfixTableLiteral` (line 648), `generatePostfixComputeBody` (line 719) — the postfix table → Go code generation methods

b. What stays in `cil_transformer.go`: `NewCILAbstractTransformer`, `Transform`, `TransformCILTrees`, `TransformMetric`, `VisitNode`, `VisitCILLiteral`, `VisitCILBinaryOp`, `VisitCILUnaryOp`, `VisitCILInputRef`, connector metadata, `ParallelOperationGroup`, `ComputationGroup`, `Computation`, `GenerateSetupFunction`, `GenerateComputeFunction`, `GenerateCleanupFunction`.

**Phase B — Verify before refactoring:**

- Run `go test ./cil/...` to confirm all existing tests pass
- Record test count

**Phase C — Execute the split:**

a. Create `postfix_walk.go` with `WalkCILTreeToPostfix` and `walkToPostfix`
b. Create `postfix_codegen.go` with `GeneratePostfixTableLiteral` and `generatePostfixComputeBody`
c. Remove moved functions from `cil_transformer.go`
d. Verify: `go test ./cil/...` must pass with zero failures

### Task 3: Create empty focused test files for Story188 conditional work — Completed

**Goal**: Create empty test files (package declarations only, no test functions) that Story188 Tasks 3–9 will populate. This prevents any task from needing to grow the 3,119-line `cil_transformer_test.go` or the 1,085-line `chronocone_tensor_test.go`.

**Phase A — Create files:**

a. `src/chronocone_planning_language/src/main/go/calculation/mask_ops_test.go` — for Tasks 3 and 4a tests (logical mask ops, AllTrueMask, AllFalseMask)
b. `src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor_where_test.go` — for Task 4b tests (TensorWhereFloat64)
c. `src/chronocone_planning_language/src/main/go/cil/cil_transformer_conditional_test.go` — for Tasks 5, 7b, 8 tests (walkToPostfix conditional nodes, TransformMetric conditional roots)
d. `src/chronocone_planning_language/src/main/go/calculation/postfix_op_conditional_test.go` — for Tasks 6 and 7c tests (EvaluatePostfixFloat64 conditional ops)

Each file must contain: copyright header, `package` declaration, and required imports (e.g., `import "testing"`). No test functions.

**Phase B — Verify:**

- `go test ./...` from `src/chronocone_planning_language/` must pass — empty test files compile but contribute zero test cases
- `go build ./...` must succeed for all packages

### Task 4: Split cil_transformer_test.go into focused test files — Completed

**Goal**: Split the 3,119-line `cil_transformer_test.go` into 5 focused test files, each under 850 lines. This eliminates the single largest context-bomb in the synthesis codebase and ensures future conditional-synthesis tests (Story188 Tasks 5-8) target small, per-concern files.

**Files created** (all in `src/chronocone_planning_language/src/main/go/cil/`):

| New file | Line source | Lines | Test count | Scope |
|----------|------------|-------|------------|-------|
| `cil_transformer_core_test.go` | 1-872 | ~872 | 25 | constructor, registration, metadata, compute/cleanup, parallel ops |
| `cil_transformer_visitor_literal_binary_test.go` | 874-1622 | ~748 | 6 | VisitCILLiteral*, VisitCILBinaryOp* |
| `cil_transformer_visitor_unary_input_test.go` | 1623-1799, 2077-2259 | ~360 | 4 | VisitCILUnaryOp*, VisitCILInputRef* (two contiguous ranges, skipping interleaved TransformMetric tests at 1805-1994) |
| `cil_transformer_metric_test.go` | 1805-1994, 2264-2562 | ~497 | 5 | TestTransformMetric* (extracted individually from two interleaved ranges) |
| `cil_transformer_postfix_test.go` | 2648-3119 | ~472 | 8 + mock | WalkCILTreeToPostfix*, GeneratePostfixTableLiteral*, code gen, mockCILNode |

**Phase A — Plan the boundaries:**

The original file has test functions interleaved: TransformMetric tests (1805-1994) sit between VisitCILUnaryOp (1623-1799) and VisitCILInputRef (2077-2259). More TransformMetric tests follow at 2264-2562. Four of the five files are contiguous copies; `cil_transformer_metric_test.go` requires extracting 5 individual functions from the original.

**Phase B — Verify before refactoring:**

- Run `go test ./cil/...` to record the current test count
- Record the exact count for regression verification

**Phase C — Execute the split:**

a. Create `cil_transformer_core_test.go` — copy lines 1-872 (copyright header, package, imports, 25 test functions from `TestNewCILAbstractTransformer` through `TestGenerateConnectorMetadataMap_NoGrain`)

b. Create `cil_transformer_visitor_literal_binary_test.go` — copy lines 874-1622 (6 test functions: `TestVisitCILLiteral`, `TestVisitCILLiteral_NilSafety`, `TestVisitCILBinaryOp_NilSafety`, `TestVisitCILBinaryOp`, `TestVisitCILBinaryOp_Nested`, `TestVisitCILBinaryOp_ShapeCompatibility`)

c. Create `cil_transformer_visitor_unary_input_test.go` — copy two contiguous ranges, skipping the interleaved TransformMetric tests:
   - Lines 1623-1799: `TestVisitCILUnaryOp_NilSafety` and `TestVisitCILUnaryOp`
   - Lines 2077-2259: `TestVisitCILInputRef_NilSafety` and `TestVisitCILInputRef`

d. Create `cil_transformer_metric_test.go` — extract these 5 individual functions from the original file, each with its preceding comment block and trailing blank line:
   1. `TestTransformMetric` (starts line 1803, with comment at 1803-1804)
   2. `TestTransformMetric_ValueTable` (starts line 1994)
   3. `TestTransformMetric_ArithmeticBinaryOp` (starts line 2261, with comment at 2261-2263)
   4. `TestTransformMetric_PostfixTableOutput` (starts line 2437)
   5. `TestTransformMetric_LiteralPathUnchanged` (starts line 2559, with comment at 2559-2561)
   
   Add copyright header and `package cil` + imports to the file.

e. Create `cil_transformer_postfix_test.go` — copy lines 2648-3119 (8 test functions from `TestWalkCILTreeToPostfix_SimpleBinaryOps` through `TestGeneratePostfixComputeBody_DoesNotContainConnectorRegistration`, plus `mockCILNode` struct and method at lines 3117-3119)

f. Delete the original `cil_transformer_test.go`

g. Verify: `go test ./cil/...` passes with the same test count as Phase B

h. The mockCILNode (defined in file e) is used by `TestTransformMetric_ValueTable` (in file d, line 1935) and `TestWalkCILTreeToPostfix_ErrorCases` (in file e, line 2872). Since all test files share `package cil`, the mock type is accessible across all files — no special handling required.

## Acceptance Criteria

1. `chronocone_tensor.go` is ≤ 650 lines (target ~530)
2. `cil_transformer.go` is ≤ 650 lines (target ~600)
3. `chronocone_tensor_arithmetic.go` exists with all 11 arithmetic functions and passes tests
4. `chronocone_tensor_comparison.go` exists with all 12 comparison functions and passes tests
5. `postfix_walk.go` exists with `WalkCILTreeToPostfix` and `walkToPostfix`
6. `postfix_codegen.go` exists with `GeneratePostfixTableLiteral` and `generatePostfixComputeBody`
7. Four empty focused test files exist with correct package declarations (for Story188 Tasks 3-9 to populate)
8. Original `cil_transformer_test.go` is deleted and replaced by 5 focused test files, all ≤ 850 lines:
   - `cil_transformer_core_test.go` (~872 lines, 25 tests)
   - `cil_transformer_visitor_literal_binary_test.go` (~748 lines, 6 tests)
   - `cil_transformer_visitor_unary_input_test.go` (~360 lines, 4 tests)
   - `cil_transformer_metric_test.go` (~497 lines, 5 tests)
   - `cil_transformer_postfix_test.go` (~472 lines, 8 tests)
9. `go test ./calculation/...` passes with zero failures (same test count as before)
10. `go test ./cil/...` passes with zero failures (same test count as before)
11. No changes to function signatures, public APIs, or test logic
12. Story188's Dependencies section is updated to reference this story

## Notes

- This refactoring is a strict prerequisite for completing Story188 Tasks 3-9. Without it, a limited-context LLM in Code mode must read 1,000+ line files to make any edit, consuming its working memory before it can produce output.
- The refactoring earns back its cost immediately: every subsequent Story188 task reads files that are 3-10x smaller than before.
- `cil_transformer_test.go` at 3,119 lines is the single largest file in the synthesis codebase. Splitting it alongside the source files ensures future test additions (in Story188 Tasks 5-8) target focused files with no temptation to append to a monolithic file.
