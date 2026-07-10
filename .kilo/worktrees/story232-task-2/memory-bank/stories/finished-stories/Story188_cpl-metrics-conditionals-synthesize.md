# Story188: CPL Metrics Conditionals Synthesize

## Status: Done

## Goal

Implement synthesis of CPL metrics conditionals, including `when` clauses and `else` branches.

## References

- [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md)
- [ERO005: Metrics Synthesis Phase — Increment 4](../specifications/design/cpl/go-synthesis-target/eros/ERO005_metrics-synthesis-phase.md)
- [Synthesis Design](../specifications/design/cpl/go-synthesis-target/synthesis.md)
- [Tensor Representation Design](../specifications/design/cpl/go-synthesis-target/tensor-representation.md)
- [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Dependencies

- [Story187_cpl-metrics-conditionals-parse-analyze.md](../stories/finished-stories/Story187_cpl-metrics-conditionals-parse-analyze.md) — parse/analyze produces valid `CILConditional`, `CILWhenClause`, comparison/logical CIL nodes (Done)
- [Story185_cpl-metrics-arithmetic-synthesize.md](../stories/finished-stories/Story185_cpl-metrics-arithmetic-synthesize.md) — arithmetic synthesis via postfix evaluator must be complete (Done)
- **[Story217_cpl-metrics-synthesis-refactoring.md](Story217_cpl-metrics-synthesis-refactoring.md) — refactoring of monolithic source files into focused per-concern files; required before Tasks 3-9 can be executed by limited-context Code mode (Done)**

## Dependent Stories

- [Story194_cpl-metrics-dag-io-analyze.md](Story194_cpl-metrics-dag-io-analyze.md) — Increment 8 Analysis (Not Started)
- [Story195_cpl-metrics-dag-io-synthesize.md](Story195_cpl-metrics-dag-io-synthesize.md) — Increment 8 Synthesis (Not Started)

## Design Decisions

- **Postfix evaluator extension** — all code gen uses one unified postfix path
- **Bit-packed uint64 masks** at whole-tensor level on `ChronoconeTensor`
- **Mask convention**: bit=`0` = true/selected, bit=`1` = false/unselected; zero-init = all-positions-true (default for non-conditional tensors)
- **Condition positions produce masks** — comparisons and logical ops in condition positions write to the tensor's mask, NOT the data
- **Body positions produce float64 data** — branch bodies use standard arithmetic ops (already supported)
- **Sequential "break into parts, then combine"** — each branch evaluated independently, computing all positions, then combined via `Where`
- **Where semantics**: element-wise, first-matching-branch-wins per position
- **Boolean expressions in body positions deferred** to a future type-system story

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. Generated code must use the postfix-table evaluator (`WalkCILTreeToPostfix` → `EvaluatePostfixFloat64`)
3. Conditionals use the break-into-parts-and-combine model (sequential, all positions computed)
4. Mask convention: bit=0 means true/selected; newly allocated tensors have all-zero mask (all positions selected)
5. Comparison ops in conditions produce masks on tensors; body ops produce float64 data
6. Go functions that can fail must return `(T, error)`
7. Types are float64 for data, uint64 bit-packed for masks
8. Follow Go testing best practices from [`resources/go/testing-guide.md`](../../resources/go/testing-guide.md)

## Tasks

### Task Sizing

- Tasks 5–4 are completed; sizing annotations are for remaining work only
- **Task 5**: Medium — extends `walkToPostfix` in `postfix_walk.go` with multiple new node type cases; introduces handling for comparison, logical, and literal CIL nodes
- **Task 6**: Small — single file change to `postfix_op.go` adding operator dispatch cases to an existing evaluator switch; narrow, well-scoped
- **Task 7**: Medium — touches `postfix_op.go` (new struct + evaluator) and `postfix_walk.go` (walker extension); introduces new operation kind and coordinated multi-file changes
- **Task 8**: Small — single file change to `cil_transformer_visitors.go` adding two case clauses that follow an existing pattern; no new structs or interfaces
- **Tasks 9.1–9.10**: Each is Small — creates a single example subdirectory following the `metrics_arithmetic` template
- **Task 9.11**: Medium — multiple files for shared infrastructure and integration test harness

### Parallel Execution

- **Group A** (Tasks 5, 6): Task 5 modifies `cil/postfix_walk.go`; Task 6 modifies `calculation/postfix_op.go`. Disjoint files in different packages. Share only the stable `PostfixOp` struct (neither task modifies it).
- **Group B** (Tasks 9.1–9.10): Each task creates an independent example subdirectory. No file conflicts; each follows the same `metrics_arithmetic` template structure.

### Execution Order

```
[Task 5 ‖ Task 6] → Task 7 → Task 8 → [Task 9.1 ‖ Task 9.2 ‖ Task 9.3 ‖ Task 9.4 ‖ Task 9.5 ‖ Task 9.6 ‖ Task 9.7 ‖ Task 9.8 ‖ Task 9.9 ‖ Task 9.10] → Task 9.11
```

### Task 1: Add mask infrastructure to ChronoconeTensor — Completed

**Goal**: Add a bit-packed uint64 mask array to `ChronoconeTensor` so conditional evaluations can store boolean results separately from float64 data.

**Phase A — Structural definitions:**

a. Add `mask []uint64` field to `ChronoconeTensor[T any]`. The mask is parallel to the `data` slice — one bit per element.

b. Add mask helper methods to `ChronoconeTensor`:

```go
func (t *ChronoconeTensor[T]) Mask() []uint64
func (t *ChronoconeTensor[T]) SetMask(mask []uint64)
func (t *ChronoconeTensor[T]) MaskLen() int  // (len(data) + 63) / 64
func (t *ChronoconeTensor[T]) IsMaskedTrue(index int) bool  // (mask[index>>6]>>(index&0x3F))&1 == 0
func (t *ChronoconeTensor[T]) SetMaskBit(index int, value bool)  // true = bit=0, false = bit=1
```

c. Update `NewChronoconeTensor` to allocate zero-initialized mask of size `(len(data) + 63) / 64`. Zero mask = all positions true/selected — correct default for non-conditional tensors.

d. Update `AllocateTensor` to also allocate the mask alongside data.

e. Update all existing tensor arithmetic functions to propagate masks: `TensorAddFloat64`, `TensorSubtractFloat64`, `TensorMultiplyFloat64`, `TensorDivideFloat64`, `TensorNegateFloat64`, and all scalar variants. For binary ops, copy the left operand's mask to the output when dimensions match.

**Phase B — Tests:**

- Mask allocation for various tensor sizes
- Bit get/set correctness at word boundaries
- Mask propagation through arithmetic ops
- Dimension mismatch handling with masks

**Files**: `chronocone_tensor.go`, `chronocone_tensor_test.go`

---

### Task 2: Implement comparison operations producing mask tensors — Completed

**Goal**: Add tensor-level comparison functions that compare two float64 tensors element-wise and produce a mask indicating which positions satisfy the comparison. The returned tensor carries the same data as the left operand with the comparison result in its mask.

**Phase A — Structural definitions:**

Add to `chronocone_tensor.go`:

```go
func TensorGreaterThanFloat64Mask(a, b *ChronoconeTensor[any]) (*ChronoconeTensor[any], error)
func TensorGreaterEqualFloat64Mask(a, b *ChronoconeTensor[any]) (*ChronoconeTensor[any], error)
func TensorLessThanFloat64Mask(a, b *ChronoconeTensor[any]) (*ChronoconeTensor[any], error)
func TensorLessEqualFloat64Mask(a, b *ChronoconeTensor[any]) (*ChronoconeTensor[any], error)
func TensorEqualFloat64Mask(a, b *ChronoconeTensor[any]) (*ChronoconeTensor[any], error)
func TensorNotEqualFloat64Mask(a, b *ChronoconeTensor[any]) (*ChronoconeTensor[any], error)
```

Plus scalar variants for each (scalar on right side):

```go
func TensorGreaterThanScalarFloat64Mask(a *ChronoconeTensor[any], scalar float64) (*ChronoconeTensor[any], error)
// ... (all 6 operators)
```

Each function validates dimensions, extracts float64 elements, allocates a mask, performs element-wise comparison setting mask bits, and returns a tensor with the left operand's data and the computed mask.

**Phase B — Tests:**

- Tensor-tensor comparison for each operator
- Scalar-tensor comparison for each operator
- Tensor-scalar comparison for each operator
- Edge cases: equal values for strict inequalities, NaN handling
- Dimension mismatch errors
- Nil input errors

**Files**: `chronocone_tensor.go`

---

### Task 3: Implement logical mask operations — Completed

**Goal**: Add functions that combine boolean masks using logical AND, OR, and NOT. These operate on raw `[]uint64` slices for composition flexibility.

**Phase A — Structural definitions:**

Create new file `mask_ops.go`:

```go
func TensorAndMask(a, b []uint64) ([]uint64, error)
func TensorOrMask(a, b []uint64) ([]uint64, error)
func TensorNotMask(mask []uint64) []uint64
```

Mask truth tables (convention: 0=true/selected, 1=false/unselected):

| A     | B     | AND   | OR    |
| ----- | ----- | ----- | ----- |
| 0 (T) | 0 (T) | 0 (T) | 0 (T) |
| 0 (T) | 1 (F) | 1 (F) | 0 (T) |
| 1 (F) | 0 (T) | 1 (F) | 0 (T) |
| 1 (F) | 1 (F) | 1 (F) | 1 (F) |

NOT flips all bits: 0→1, 1→0.

**Phase B — Tests:**

- All bit combinations for AND, OR
- NOT inversion
- Different mask lengths
- Nil input handling
- Word-boundary behavior

**Files**: `mask_ops.go` (new), `mask_ops_test.go` (new)

---

### Task 4: Implement literal-to-mask conversion and Where operation — Completed

**Goal**: Convert True/False/Nil literals to appropriate masks, and implement element-wise mask-based selection from multiple branch tensors.

**Phase A — Structural definitions:**

a. Add to `mask_ops.go`:

```go
func AllTrueMask(n int) []uint64   // every bit = 0
func AllFalseMask(n int) []uint64  // every bit = 1
```

b. Add to `chronocone_tensor.go`:

```go
// TensorWhereFloat64 combines branch tensors element-wise by mask.
// Branches are processed in order; first branch with mask bit=0 contributes its value.
// All branches must have matching dimensions and shapes.
// Returns a new tensor with all-zero mask (all positions resolved).
func TensorWhereFloat64(branches []*ChronoconeTensor[any]) (*ChronoconeTensor[any], error)
```

The Where function allocates a new data array, and for each element position iterates branches in order taking the value from the first branch whose mask bit is `0` at that position. If no branch has mask bit `0` (shouldn't happen if the last branch is else/default with all-true mask), the value from the last branch is used.

**Phase B — Tests:**

- Single branch (always selected)
- Two branches with complementary masks (if/else)
- Three branches with disjoint masks (if/elseif/else)
- Overlapping masks (first branch wins)
- Branch dimension mismatch errors
- All branches with nil mask edge case

**Files**: `mask_ops.go`, `chronocone_tensor.go`

---

### Task 5: Extend walkToPostfix for flat conditional CIL nodes — Not Started

**Goal**: Extend the `walkToPostfix` method to handle comparison, logical, and literal CIL nodes, producing appropriate PostfixOp entries.

**Phase A — Structural definitions:**

Add cases to the switch in `walkToPostfix`:

- `*CILComparisonOp`: recurse Left, recurse Right, append `PostfixOp{Kind: "apply", Op: n.Op}` (one of `"==", "!=", "<", ">", "<=", ">="`)
- `*CILLogicalOp`: recurse Left, recurse Right, append `PostfixOp{Kind: "apply", Op: n.Op}` (`"and"` or `"or"`)
- `*CILNotOp`: recurse Operand, append `PostfixOp{Kind: "apply", Op: "not"}`
- `*CILBoolLiteral`: append `PostfixOp{Kind: "push", Op: "bool", Value: 1.0}` for True or `PostfixOp{Kind: "push", Op: "bool", Value: 0.0}` for False
- `*CILNilLiteral`: append `PostfixOp{Kind: "push", Op: "nil"}`

The `Op` field on `PostfixOp` (normally used for operators on `"apply"` kind) is repurposed for `"push"` kind as a type discriminator. The existing `PostfixOp` struct supports this without changes since `Op` is a string field.

**Phase B — Tests:**

- `a > 5` → `[push "a", push{scalar 5.0}, apply ">"]`
- `a > 0 and b > 0` → proper compound postfix
- `not a > 0` → proper unary postfix
- True/False/Nil literal pushes
- Negative: unsupported node types still return errors

**Files**: `cil_transformer.go`, `cil_transformer_test.go`

---

### Task 6: Extend EvaluatePostfixFloat64 for comparison, logical, and literal operators — Completed [Small]

**Goal**: Extend the postfix evaluator to handle new apply operators and literal types.

**Phase A — Structural definitions:**

Add cases to the `switch op.Op` in `EvaluatePostfixFloat64`:

- `"==", "!=", "<", ">", "<=", ">="`: pop two float64 tensors, call corresponding `TensorXxxFloat64Mask`, push result (tensor with mask)
- `"and"`: pop two masked tensors, call `TensorAndMask` on their masks, attach to copy of first tensor's data, push
- `"or"`: pop two masked tensors, call `TensorOrMask` on their masks, attach to copy, push
- `"not"`: pop one masked tensor, call `TensorNotMask` on its mask, attach to copy, push

Handle new push literal types in the `"push"` case:

- `op.Op == "bool" && op.Value == 1.0` → scalar tensor with `AllTrueMask(1)` and data `[]any{1.0}`
- `op.Op == "bool" && op.Value == 0.0` → scalar tensor with `AllFalseMask(1)` and data `[]any{0.0}`
- `op.Op == "nil"` → scalar tensor with `AllFalseMask(1)` and data `[]any{0.0}`

**Phase B — Tests:**

- Postfix table evaluation producing correct mask for `a > b`
- Multi-step: `a > 0 and b > 0` producing correct compound mask
- NOT inversion verification
- True/False/Nil literal pushes
- Stack underflow for missing operands
- Unknown operator error

**Files**: `postfix_op.go`, `postfix_op_test.go`

---

### Task 7: Add conditional postfix operator and evaluator support — Completed

**Goal**: Add a new `"conditional"` PostfixOp kind for if/else and when/default structures, and extend the evaluator to process it by breaking each expression into parts, evaluating independently, and combining with Where.

**Phase A — Structural definitions:**

a. Extend `PostfixOp` with a new field:

```go
type PostfixOp struct {
    Kind     string           // "push", "apply", or "conditional"
    Op       string           // operator for "apply"; type discriminator for "push"
    Name     string           // tensor name for "push"
    Value    float64          // scalar value for "push"
    Branches []PostfixBranch  // sub-tables for "conditional" kind
}
```

b. Define `PostfixBranch` in the same file:

```go
type PostfixBranch struct {
    Condition []PostfixOp  // nil for else/default branch (always-true mask)
    Body      []PostfixOp  // body expression postfix table
}
```

c. Extend `walkToPostfix`:

- `*CILConditional`: for each `CILConditionalBranch`:
  - Walk condition (if non-nil) → condition ops; walk body → body ops
  - Else branch has nil condition
  - Append `PostfixOp{Kind: "conditional", Branches: [...]}`
- `*CILWhenClause`: for each `CILWhenCase`:
  - Walk condition → condition ops; walk body → body ops
  - Walk `node.Default` → body ops, nil condition
  - Append `PostfixOp{Kind: "conditional", Branches: [...]}`

d. Extend `EvaluatePostfixFloat64` for `Kind: "conditional"`:

For each branch in order:

1. If `branch.Condition != nil`: recursively evaluate → get a masked tensor (condition mask)
2. If `branch.Condition == nil`: create an all-true mask tensor (else/default)
3. Evaluate `branch.Body` → get body result tensor
4. Apply condition mask to body result (attach condition's mask)
5. Collect branch result
6. After all branches: call `TensorWhereFloat64(branchResults)` → combined result
7. Push combined result

**Phase B — Tests:**

- Simple if/else: two branches, complementary masks
- If/elseif/else: three branches, proper fallthrough
- When/default: multiple when clauses + default
- Nested conditionals (conditional inside a branch body)
- Negative: missing branches, empty conditional

**Files**: `postfix_op.go`, `cil_transformer.go`

---

### Task 8: Extend TransformMetric for conditional CIL root nodes — Completed [Small]

**Goal**: Extend the type switch in `TransformMetric` to handle `*CILConditional` and `*CILWhenClause` as root nodes, routing them to the postfix evaluator path.

**Phase A — Structural definitions:**

Add cases to the switch in `TransformMetric`:

```go
case *CILConditional:
    body, err := t.generatePostfixComputeBody(metricName, metricNode)
    if err != nil {
        return "", fmt.Errorf("failed to generate postfix compute body: %w", err)
    }
    postfixBody = body
case *CILWhenClause:
    // same pattern as CILConditional
```

These follow the exact same pattern as `*CILBinaryOp` and `*CILUnaryOp` — they delegate to `generatePostfixComputeBody` which calls `WalkCILTreeToPostfix` then generates the `EvaluatePostfixFloat64` call.

**Phase B — Tests:**

- `TransformMetric` with `CILConditional` root produces valid Go code
- `TransformMetric` with `CILWhenClause` root produces valid Go code
- Short form metric with conditional (e.g., `m = if a > b then a+b else a-b end if`)
- Negative: nil root, unsupported node types

**Files**: `cil_transformer.go`, `cil_transformer_test.go`

---

### Task 9: End-to-end verification with example CPL files — Not Started

**Goal**: Verify the full pipeline (parse → analyze → synthesize → compile → execute) works for conditional metrics.

**Implementation details:**

Create a new directory `examples/with_data/metrics_conditionals` following the same structure as `metrics_arithmetic` (shared `core.cpl`, per-example subdirectories with `.cpl`, `iospec.json`, `expected_output.json`, parquet data files, and `parquet_gen_spec_*.json` files). Each example uses concepts `Region=[North,South]`, `Package=[Each]` and connectors `data_a`, `data_b`.

Write Go integration tests that: parse CPL → analyze to CIL → synthesize Go code → compile Go → run with test data → verify output tensors.

Run full CPL test suite (`go test ./...` from `src/chronocone_planning_language`) to confirm no regressions.

**Tooling and workflow:**

*Generating parquet files:*
The project provides a CLI tool `parquetgen` at `src/main/go/cli/parquetgen/parquet_gen.go` for generating deterministic test parquet files from JSON specifications. Each example subdirectory contains one `parquet_gen_spec_<connector>.json` per input connector. The spec format is:

```json
{
  "dimensions": { "Region": ["North", "South"] },
  "grain": [{ "measure": "Daily", "values": ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"] }],
  "values": [10.0, 10.0, ...]
}
```

To generate a parquet file manually, run from the repository root:

```bash
go run src/chronocone_planning_language/src/main/go/cli/parquetgen/parquet_gen.go \
  -input=<path/to/parquet_gen_spec_a.json> \
  -output=<path/to/data_a.parquet> \
  -json_output=<path/to/data_a.json>
```

The `-json_output` flag is optional and produces a human-readable JSON representation of the same data for inspection and for authoring `expected_output.json`.

*Running a single example manually:*
Use the `compile_and_run.sh` script located at `src/chronocone_planning_language/shell_scripts/compile_and_run.sh`. The script compiles the CPL files (including `core.cpl` when present), runs the synthesized Go program against the example's parquet data, and verifies the output against `expected_output.json`.

```bash
cd src/chronocone_planning_language/shell_scripts
./compile_and_run.sh <example_name> ../examples/with_data/metrics_conditionals
```

Example:

```bash
./compile_and_run.sh conditional_if_else_arithmetic ../examples/with_data/metrics_conditionals
```

*Running all examples via Go test:*
The integration test harness (`metrics_conditionals_test.go`) discovers all leaf example directories under `metrics_conditionals`, parses each `.cpl` file with `core.cpl`, and runs `compile_and_run.sh` for each. To execute:

```bash
cd src/chronocone_planning_language
go test ./examples/with_data/metrics_conditionals/... -v
```

To run the full CPL test suite and check for regressions:

```bash
cd src/chronocone_planning_language
go test ./... -v
```

---

#### Task 9.1: Example — conditional_if_else_arithmetic — Completed [Small]

**Goal**: End-to-end test for basic if/else with arithmetic in both branches.

- Create `metrics_conditionals/conditional_if_else_arithmetic/` subdirectory.
- Adapt the CPL from `parse_and_analyze_only/without_errors/conditional_if_else_arithmetic.cpl` to use the shared `core.cpl` pattern (declare connectors in the example file, metric references them).
- Create `iospec.json`, `expected_output.json`, `parquet_gen_spec_a.json`, `parquet_gen_spec_b.json`, and parquet data files.
- Data design: `data_a` values > `data_b` values for North (true branch: addition), `data_a` values < `data_b` values for South (false branch: subtraction).
- Verify output tensor values match branch selection per Region.

**Files**: `conditional_if_else_arithmetic.cpl`, `iospec.json`, `expected_output.json`, `data_a.parquet`, `data_b.parquet`, `parquet_gen_spec_a.json`, `parquet_gen_spec_b.json`

---

#### Task 9.2: Example — conditional_if_elseif_else — Completed [Small]

**Goal**: End-to-end test for three-branch if/elseif/else.

- Create `metrics_conditionals/conditional_if_elseif_else/` subdirectory.
- Adapt the CPL from `parse_and_analyze_only/without_errors/conditional_if_elseif_else.cpl`.
- Data design: North values > 10 (first branch), South values between 5 and 10 (second branch), no remaining positions (else branch not triggered in this data set, or design data so else is hit).
- Verify correct branch selection per position.

**Files**: `conditional_if_elseif_else.cpl`, `iospec.json`, `expected_output.json`, data files, gen specs

---

#### Task 9.3: Example — conditional_when_default — Completed [Small]

**Goal**: End-to-end test for when/default syntax.

- Create `metrics_conditionals/conditional_when_default/` subdirectory.
- Adapt the CPL from `parse_and_analyze_only/without_errors/conditional_when_default.cpl`.
- Data design: values that trigger each when clause and the default clause.
- Verify output matches first-matching-branch-wins semantics.

**Files**: `conditional_when_default.cpl`, `iospec.json`, `expected_output.json`, data files, gen specs

---

#### Task 9.4: Example — conditional_comparison_operators — Completed [Small]

**Goal**: End-to-end test exercising all six comparison operators across multiple branches.

- Create `metrics_conditionals/conditional_comparison_operators/` subdirectory.
- Adapt the CPL from `parse_and_analyze_only/without_errors/conditional_comparison_operators.cpl`.
- Data design: values that exercise `==`, `!=`, `>=`, `<=`, `>`, `<` in sequence.
- Verify each operator selects the correct branch.

**Files**: `conditional_comparison_operators.cpl`, `iospec.json`, `expected_output.json`, data files, gen specs

---

#### Task 9.5: Example — conditional_logical_operators — Completed [Small]

**Goal**: End-to-end test for logical `and`, `or`, and `not` in conditions.

- Create `metrics_conditionals/conditional_logical_operators/` subdirectory.
- Adapt the CPL from `parse_and_analyze_only/without_errors/conditional_logical_operators.cpl`.
- Data design: combinations of positive and negative values to exercise all logical combinations.
- Verify compound masks select correct branches.

**Files**: `conditional_logical_operators.cpl`, `iospec.json`, `expected_output.json`, data files, gen specs

---

#### Task 9.6: Example — conditional_unary_negation — Completed [Small]

**Goal**: End-to-end test for unary negation in conditions and branch bodies.

- Create `metrics_conditionals/conditional_unary_negation/` subdirectory.
- Adapt the CPL from `parse_and_analyze_only/without_errors/conditional_unary_negation.cpl`.
- Data design: values where `not (a > b)` is true for some positions and false for others.
- Verify negation flips branch selection correctly and body negation produces correct values.

**Files**: `conditional_unary_negation.cpl`, `iospec.json`, `expected_output.json`, data files, gen specs

---

#### Task 9.7: Example — conditional_short_form — Completed [Small]

**Goal**: End-to-end test for short-form metric with conditional expression.

- Create `metrics_conditionals/conditional_short_form/` subdirectory.
- Adapt the CPL from `parse_and_analyze_only/without_errors/conditional_short_form.cpl`.
- Data design: same complementary pattern as if/else arithmetic.
- Verify short-form syntax synthesizes and executes identically to long-form.

**Files**: `conditional_short_form.cpl`, `iospec.json`, `expected_output.json`, data files, gen specs

---

#### Task 9.8: Example — conditional_truthiness — Completed [Small]

**Goal**: End-to-end test for numeric truthiness and True/False/Nil literals in conditions.

- Create `metrics_conditionals/conditional_truthiness/` subdirectory.
- Adapt the CPL from `parse_and_analyze_only/without_errors/conditional_truthiness.cpl`.
- Data design: zero values (falsy) and non-zero values (truthy) for `input_a`; `False` and `True` literal branches.
- Verify truthiness evaluation and literal branch selection.

**Files**: `conditional_truthiness.cpl`, `iospec.json`, `expected_output.json`, data files, gen specs

---

#### Task 9.9: Example — conditional_tautological_if_only — Completed [Small]

**Goal**: End-to-end test for tautological if without else.

- Create `metrics_conditionals/conditional_tautological_if_only/` subdirectory.
- Adapt the CPL from `parse_and_analyze_only/without_errors/conditional_tautological_if_only.cpl`.
- Data design: any values (condition is always true).
- Verify all positions select the single branch and produce correct values.

**Files**: `conditional_tautological_if_only.cpl`, `iospec.json`, `expected_output.json`, data files, gen specs

---

#### Task 9.10: Example — conditional_nested_arithmetic — Completed [Small]

**Goal**: End-to-end test for complex parenthesized arithmetic inside conditional branches.

- Create `metrics_conditionals/conditional_nested_arithmetic/` subdirectory.
- Adapt the CPL from `parse_and_analyze_only/without_errors/conditional_nested_arithmetic.cpl`.
- Data design: positive values for first branch, negative values for second branch, zero for else.
- Verify nested arithmetic expressions evaluate correctly within each branch.

**Files**: `conditional_nested_arithmetic.cpl`, `iospec.json`, `expected_output.json`, data files, gen specs

---

#### Task 9.11: Shared infrastructure and integration test harness — Completed

**Goal**: Create shared `core.cpl`, Go package files, and integration test harness for all conditional examples.

**Phase A — Structural definitions:**

a. Create `metrics_conditionals/core.cpl` with shared concepts, types, units, and measures (same pattern as `metrics_arithmetic/core.cpl`).

b. Create `metrics_conditionals/example_definitions.go` with constants for each example CPL path.

c. Create `metrics_conditionals/iospec_definitions.go` with `IOSpec` factory functions for each example.

d. Create `metrics_conditionals/iospec_examples_test.go` verifying each IO spec loads correctly.

e. Create `metrics_conditionals/metrics_conditionals_test.go` with:
- `TestExampleFilesParse` — parse each example with `core.cpl`
- `TestMetricsConditionalsEndToEnd` — run `compile_and_run.sh` for each example and verify output

**Phase B — Tests:**

- All 10 examples parse without errors
- All 10 examples run end-to-end and produce expected output
- No regressions in existing CPL test suite

**Files**: `core.cpl`, `example_definitions.go`, `iospec_definitions.go`, `iospec_examples_test.go`, `metrics_conditionals_test.go`

## Acceptance Criteria

1. Mask infrastructure exists on `ChronoconeTensor` with correct allocation and helper methods
2. Comparison operations produce correct mask tensors for all 6 operators
3. Logical AND, OR, NOT operate correctly on bit-packed masks
4. `TensorWhereFloat64` correctly selects first-matching-branch per position
5. `walkToPostfix` handles all flat conditional CIL nodes (comparison, logical, literal)
6. `EvaluatePostfixFloat64` correctly evaluates comparison, logical, and literal operators
7. `PostfixOp` supports `"conditional"` kind with `PostfixBranch` sub-tables
8. `EvaluatePostfixFloat64` correctly processes conditional branches and combines with Where
9. `TransformMetric` routes `CILConditional` and `CILWhenClause` roots through postfix path
10. End-to-end tests verify correct results for all 10 example CPL files
11. All unit tests pass following Go testing best practices
12. No regressions in existing CPL test suite

## Notes

- The original Story188 assumed snippet-stack visitors (`VisitCILComparisonOp`, etc.). This rewrite aligns with the postfix-table architecture established in Story185 and Story204.
- Boolean expressions in body positions are intentionally deferred to a future type-system story.
- All arithmetic tensor functions must propagate masks to maintain correctness when conditionals are nested inside arithmetic.
