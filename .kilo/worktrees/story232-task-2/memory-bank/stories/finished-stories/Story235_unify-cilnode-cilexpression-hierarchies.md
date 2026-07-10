# Story235: Unify CILNode and CILExpression Hierarchies

## Goal

Merge the `CILExpression` types created by Story228 into the existing `CILNode` hierarchy (created by Story184, unified by Story222). After this story there is exactly one CIL expression tree interface in the `cil/` package. Backward compatibility is NOT required — simplify wherever possible.

Story228 created a new `CILExpression` interface hierarchy in `cil/` alongside the existing `CILNode` hierarchy. Both represent CIL expression trees but are incompatible. Story231 needs to walk metric formula trees (CILNode) AND compute unit expressions (CILExpression); without unification, it cannot function without a fragile bridge layer. This story removes that technical debt before it blocks downstream work.

## References

- [Epic003: Units and Measures Implementation](../epics/Epic003_units-and-measures-implementation.md)
- [Story228_cpl-expression-infrastructure.md](finished-stories/Story228_cpl-expression-infrastructure.md) — provides the CILExpression types to be unified
- [Story222_cpl-refactoring-phase2-normalize-cil.md](Story222_cpl-refactoring-phase2-normalize-cil.md) — provides the existing CILNode hierarchy
- [ERO003: Units and Measures Incremental Implementation](../specifications/design/cpl/eros/ERO003_units-measures-incremental-implementation.md)
- [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Dependencies

- [Story228_cpl-expression-infrastructure.md](finished-stories/Story228_cpl-expression-infrastructure.md) — _Done_ — provides the CILExpression types to be unified
- [Story222_cpl-refactoring-phase2-normalize-cil.md](Story222_cpl-refactoring-phase2-normalize-cil.md) — _Done_ — provides the existing CILNode hierarchy

## Dependent Stories

- [Story229_cpl-compound-units.md](Story229_cpl-compound-units.md) — _Not Started_ — must use unified CILNode types for CIL output
- [Story231_cpl-unit-type-checking.md](Story231_cpl-unit-type-checking.md) — _Not Started_ — must walk a single CIL expression tree interface
- [Story236_cpl-expression-pipeline-integration.md](Story236_cpl-expression-pipeline-integration.md) — _Not Started_ — integration test targets unified CILNode types

## Tasks

### Task Sizing

- **Tasks 1–3** are `[Small]`: each creates a single source file and its corresponding test file (at most 2 files total), follows an existing pattern (CILNode marker method), and is straightforward with no ambiguity.
- **Task 8** is `[Small]`: a single file deletion plus compile verification.
- All other tasks are Medium: they introduce coordinated changes across multiple files, require reasoning about interacting structures, or span packages.

### Parallel Execution

- **Group A** (Tasks 1, 2, 3): Each task creates a new source file and its test file in the `cil/` package. The file sets are completely disjoint and none depends on another's output.

### Execution Order

```
[Task 1 ‖ Task 2 ‖ Task 3] → Task 4 → Task 5 → Task 6 → Task 7 → Task 8 → Task 9
```

- Tasks 1, 2, and 3 can run in parallel because they are file-disjoint struct definitions in the same package.
- Task 4 depends on Tasks 1–3 (modifies `cil_binary_op.go` which now coexists with the new types; all compile together).
- Task 5 depends on Task 4 (tests the migrated `CILBinaryOp` with `OperatorType`).
- Task 6 depends on Task 4 (needs the unified `CILBinaryOp` with `OperatorType`).
- Task 7 depends on Task 6 (`SynthesizeExpression` signature must be updated before its tests can reference it).
- Task 8 depends on Task 4 (moves `OperatorType` out of `cil_expression.go` before it can be deleted).
- Task 9 depends on all preceding tasks.

### Reintegration

**Group A (Tasks 1, 2, 3):**
- Merge order: any order (fully disjoint files, no shared infrastructure)
- Integration tests: `go test ./src/main/go/cil/... -count=1 -timeout 60s` — verify all new types compile and pass tests together
- Watch for conflicts in: `cil/` package-level imports (each task may add the same standard imports)

### Task 1: Create cil_unit_ref.go with CILUnitRef CILNode type — Completed
   a. Write unit test in `src/main/go/cil/cil_unit_ref_test.go`: compile-time assertion `var _ CILNode = (*CILUnitRef)(nil)` and construction/zero-value tests — Not Started
   b. Create `src/main/go/cil/cil_unit_ref.go`: define `CILUnitRef` struct with `UnitGroup string` field and `cilNode()` marker method — Not Started
   c. Verify tests pass: `go test ./src/main/go/cil/... -count=1 -timeout 60s` — Not Started

### Task 2: Create cil_power.go with CILPower CILNode type — Completed
   a. Write unit test in `src/main/go/cil/cil_power_test.go`: compile-time assertion `var _ CILNode = (*CILPower)(nil)` and construction/zero-value tests — Not Started
   b. Create `src/main/go/cil/cil_power.go`: define `CILPower` struct with `Base CILNode`, `Exponent int`, `DerivedUnit string` fields and `cilNode()` marker method — Not Started
   c. Verify tests pass — Not Started

### Task 3: Create cil_number.go with CILNumber CILNode type — Completed
   a. Write unit test in `src/main/go/cil/cil_number_test.go`: compile-time assertion `var _ CILNode = (*CILNumber)(nil)` and construction/zero-value tests — Not Started
   b. Create `src/main/go/cil/cil_number.go`: define `CILNumber` struct with `Value float64`, `AnnotatedUnit string` fields and `cilNode()` marker method — Not Started
   c. Verify tests pass — Not Started

### Task 4: Migrate CILBinaryOp to OperatorType and add unit tracking fields — Completed
   a. Move `OperatorType` type and constants (`CILMultiply`, `CILDivide`, `CILPlus`, `CILMinus`) from `cil_expression.go` to `cil_binary_op.go` — Not Started
   b. Change `CILBinaryOp.Op` from `string` to `OperatorType`; add `DerivedUnit string` and `AnnotatedUnit string` fields after existing fields — Not Started
   c. Add `arithmeticOperatorType(def DefinitionType) (cil.OperatorType, error)` in `src/main/go/analysis/metrics_analysis.go`, mapping `PlusOperatorDefinition` → `cil.CILPlus`, `MinusOperatorDefinition` → `cil.CILMinus`, `MultiplyOperatorDefinition` → `cil.CILMultiply`, `DivideOperatorDefinition` → `cil.CILDivide` — Not Started
   d. Remove the four arithmetic cases from `operatorSymbol` in `metrics_analysis.go` (keep unary, comparison, and logical operators) — Not Started
   e. Update `buildCILTreeFromPostfix` in `metrics_analysis.go` to call `arithmeticOperatorType` instead of `operatorSymbol` for arithmetic binary ops — Not Started
   f. Update test file call sites in 8 test files from `Op: "+"` string literals to `Op: CILPlus` enum values (and similarly for `-`, `*`, `/`) — Not Started
      - `src/main/go/analysis/metrics_analysis_conditional_test.go`
      - `src/main/go/analysis/metrics_analysis_test.go`
      - `src/main/go/cil/cil_transformer_visitor_literal_binary_test.go`
      - `src/main/go/cil/cil_transformer_visitor_unary_input_test.go`
      - `src/main/go/cil/cil_transformer_postfix_test.go`
      - `src/main/go/cil/cil_transformer_metric_test.go`
      - `src/main/go/cli/expected_ods_generator/generator_test.go`
      - `src/main/go/cli/expected_ods_generator/cil_to_odf_test.go`
      - `src/main/go/cli/expected_ods_generator/cil_evaluator_test.go`
   g. Verify all affected packages compile and tests pass — Not Started

### Task 5: Migrate and rewrite CIL expression tests for unified types — Completed
   a. Create `src/main/go/cil/cil_unit_ref_test.go` with migrated `TestCILUnitRefConstruction` and `TestCILUnitRefZeroValue` plus compile assertion — Not Started
   b. Create `src/main/go/cil/cil_power_test.go` with migrated `TestCILPowerConstruction` and `TestCILPowerZeroValue` (using `CILUnitRef` as `Base` child) plus compile assertion — Not Started
   c. Create `src/main/go/cil/cil_number_test.go` with migrated `TestCILNumberConstruction` and `TestCILNumberZeroValue` plus compile assertion — Not Started
   d. Create `src/main/go/cil/cil_binary_op_test.go` with:
      - Rewritten `TestCILBinaryOpConstruction` (OperatorType, DerivedUnit, AnnotatedUnit fields)
      - Rewritten `TestCILBinaryOpNilOperands` (nil Left/Right CILNode children)
      - Migrated `TestOperatorTypeValues`
      - Rewritten `TestNestedCILNodeConstruction` (nested `CILBinaryOp` with `CILUnitRef` leaves)
      - Compile assertion `var _ CILNode = (*CILBinaryOp)(nil)` — Not Started
   e. Delete `src/main/go/cil/cil_expression_test.go` — Not Started
   f. Verify all `cil/` tests pass — Not Started

### Task 6: Update SynthesizeExpression to accept CILNode — Completed
   a. Write tests in `src/main/go/synthesis/expression_synthesis_test.go` for `SynthesizeExpression` with unified `CILNode` types (`CILUnitRef`, `CILBinaryOp`, `CILPower`, `CILNumber`) — Not Started
   b. Change function signature to `func SynthesizeExpression(node cil.CILNode) string` in `src/main/go/synthesis/expression_synthesis.go` — Not Started
   c. Implement type-switch on `*CILUnitRef`, `*CILBinaryOp`, `*CILPower`, `*CILNumber`; for unsupported `CILNode` types return empty string with descriptive error (no panics) — Not Started
   d. Verify synthesis tests pass — Not Started

### Task 7: Update expression synthesis tests for unified API — Completed
   a. Replace all `CILExpression` type references with `CILNode` types in `src/main/go/synthesis/expression_synthesis_test.go` — Not Started
   b. Replace `CILBinaryExpr` with `CILBinaryOp` using `OperatorType` enum values — Not Started
   c. Verify tests pass — Not Started

### Task 8: Delete cil_expression.go — Completed [Small]
   a. Remove `src/main/go/cil/cil_expression.go` containing `CILExpression` interface, `CILBinaryExpr` type, and moved `OperatorType` constants — Not Started
   b. Verify package compiles: `go build ./src/main/go/cil/...` — Not Started

### Task 9: Run full CPL test suite — Completed
   a. Run `go test ./... -count=1 -timeout 300s` from `src/chronocone_planning_language/src/main/go/` — Not Started
   b. Fix any regressions — Not Started

## Test-First Development

This project follows the Logical TDD Lifecycle. For each coding task above, tests are written before implementation and the test-and-implement cycle is integrated within each task rather than broken out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. No backward compatibility required — simplify wherever possible
3. `CILBinaryOp.Op` changes from `string` to `OperatorType`; all call sites must be updated mechanically
4. `CILUnaryOp.Op` remains `string` (unary "negate" is not an `OperatorType`)
5. `CILExpression` interface, `CILBinaryExpr` type, and `cil_expression.go` are fully removed
6. Go functions that can fail must return `(T, error)`
7. Generated Go tests must follow the [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Intent

- **Task 1 Intent**: Introduce `CILUnitRef` as a proper `CILNode` type for unit group references in expression trees
- **Task 2 Intent**: Introduce `CILPower` as a proper `CILNode` type for exponentiation with unit tracking
- **Task 3 Intent**: Introduce `CILNumber` as a proper `CILNode` type for numeric literals with unit annotations
- **Task 4 Intent**: Absorb `CILBinaryExpr` fields into `CILBinaryOp`, upgrading its `Op` field from `string` to the typed `OperatorType` enum
- **Task 5 Intent**: Preserve all existing test coverage from `cil_expression_test.go` by migrating or rewriting tests for the unified types
- **Task 6 Intent**: Make `SynthesizeExpression` operate on the unified `CILNode` interface instead of the removed `CILExpression` interface
- **Task 7 Intent**: Update synthesis tests to reference the unified `CILNode`-based API
- **Task 8 Intent**: Remove the now-redundant `CILExpression` interface and `cil_expression.go` file
- **Task 9 Intent**: Verify no regressions across the entire CPL codebase after unification

## Acceptance Criteria

1. `CILUnitRef`, `CILPower`, `CILNumber` implement `CILNode` (verified by compile-time assertions)
2. `CILBinaryOp` uses `OperatorType` enum for `Op`; has `DerivedUnit` and `AnnotatedUnit` fields
3. `CILExpression` interface, `CILBinaryExpr` type, and `cil/cil_expression.go` are deleted
4. `SynthesizeExpression` accepts `CILNode` and produces identical output for previously-tested expressions
5. All 11 migrated/rewritten tests from `cil_expression_test.go` pass in their new locations
6. All existing metrics CIL transformer tests pass (no regression from `Op` field migration)
7. All existing ODS generator/evaluator tests pass (no regression from `Op` field migration)
8. All existing metrics analysis tests pass (no regression from `operatorSymbol` split)
9. Full CPL test suite passes: `go test ./... -count=1`

## Footer

### Open Questions

None

### Known Issues

None

### Additional Context

**Unification summary:**

| Hierarchy | Interface | Binary Op Type | Op Field Type | Unit Tracking |
|-----------|-----------|---------------|---------------|--------------|
| `CILNode` (Story184) | `cilNode()` marker | `CILBinaryOp` | `string` (`"+"`, `"-"`, `"*"`, `"/"`) | None |
| `CILExpression` (Story228) | `cilExpression()` marker | `CILBinaryExpr` | `OperatorType` enum | `DerivedUnit`, `AnnotatedUnit` |
| **Unified (Story235)** | `cilNode()` marker | `CILBinaryOp` | `OperatorType` enum | `DerivedUnit`, `AnnotatedUnit` |

After this story, there is exactly one CIL expression tree interface: `CILNode`. All expression-related types (`CILUnitRef`, `CILBinaryOp`, `CILPower`, `CILNumber`) implement it. The `OperatorType` enum lives in `cil_binary_op.go` alongside `CILBinaryOp`, which is its primary consumer.

**Files created:**
- `src/main/go/cil/cil_unit_ref.go`
- `src/main/go/cil/cil_power.go`
- `src/main/go/cil/cil_number.go`
- `src/main/go/cil/cil_unit_ref_test.go`
- `src/main/go/cil/cil_power_test.go`
- `src/main/go/cil/cil_number_test.go`
- `src/main/go/cil/cil_binary_op_test.go`

**Files modified:**
- `src/main/go/cil/cil_binary_op.go`
- `src/main/go/analysis/metrics_analysis.go`
- `src/main/go/synthesis/expression_synthesis.go`
- `src/main/go/synthesis/expression_synthesis_test.go`
- 8 additional test files (mechanical `Op` string → enum changes)

**Files removed:**
- `src/main/go/cil/cil_expression.go`
- `src/main/go/cil/cil_expression_test.go`
