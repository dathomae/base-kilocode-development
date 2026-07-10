# Story224: CPL Refactoring — Phase 4 Reduce Duplication

## Goal

Unify duplicated operator mappings, scalar detection logic, and constraint validation patterns across the CPL compiler. Introduce a shared operator registry, tensor helpers, and generic constraint validation.

## References

- [CPL Refactoring Plan](../plans/cpl_refactoring_plan.md)
- [Story223: Phase 3 Split Bloat](Story223_cpl-refactoring-phase3-split-bloat.md)

## Dependencies

- [Story223_cpl-refactoring-phase3-split-bloat.md](Story223_cpl-refactoring-phase3-split-bloat.md) — must be Done. Phase 3 splits monolithic structs, making it easier to introduce shared registries and helpers in isolated components.

## Dependent Stories

- [Story225_cpl-refactoring-phase5-address-fragility.md](Story225_cpl-refactoring-phase5-address-fragility.md) — Not Started

## Tasks

### Task Sizing

Tasks 1 and 2 are additive helper introductions (new enum, new method) and qualify as `[Small]` — each touches at most 2 files in a single package with narrow scope. Task 3 introduces a new interface and implements it for 5 constraint types with non-trivial validation/merge logic; it spans 3+ files and multiple constraint domains, so it is **not** `[Small]`. Task 4 migrates all operator dispatch sites to use the new registry and touches 5+ files across 4 packages, so it is also not `[Small]`.

### Parallel Execution

- **Group A** (Tasks 1, 2, 3): These tasks introduce independent helpers in different packages. They are purely additive with no shared files.

### Execution Order

```
[Task 1 ‖ Task 2 ‖ Task 3] → Task 4
```

### Reintegration

**Group A (Tasks 1, 2, 3):**
- Merge order: any order (fully disjoint packages — `cil`, `calculation`, `analysis` — with no shared files or infrastructure modifications)
- Integration tests: `go test ./src/main/go/cil/... ./src/main/go/calculation/... ./src/main/go/analysis/... -count=1 -timeout 120s` — verify that each package's tests pass without cross-contamination after all worktrees are merged
- Watch for conflicts in: `go.sum` (dependency resolution in separate worktrees may produce different hashes), `go.mod` (unlikely but possible if any task adds imports)

### Task 1: Define shared operator registry — Not Started [Small]

**Goal**: Replace string-based operator dispatch with a typed `Operator` enum.

**Test-First Development**:

a. Define an `Operator` type (e.g., `type Operator int`) in `src/main/go/cil/operator.go` with constants for all supported operators (`Add`, `Subtract`, `Multiply`, `Divide`, `Negate`, `GreaterThan`, etc.).

b. Implement `String() string` method for `Operator` for debugging and code generation.

c. Write unit tests in `src/main/go/cil/operator_test.go` for the enum (construction, String output, exhaustiveness).

d. Run `go test ./src/main/go/cil/...` — must pass.

**Acceptance Criteria**:
- `Operator` enum defined with all operators used in codegen and evaluation
- `String()` method implemented and tested
- No changes to existing dispatch sites yet

### Task 2: Add IsScalar method to ChronoconeTensor — Not Started [Small]

**Goal**: Replace the repeated scalar-detection pattern with a single method.

**Test-First Development**:

a. Add `IsScalar() bool` to `ChronoconeTensor` in `src/main/go/calculation/chronocone_tensor.go`.

b. Implement as `return len(t.DimensionNames()) == 0 && len(t.Shape()) == 0 && len(t.Data()) == 1`.

c. Write unit tests in `src/main/go/calculation/chronocone_tensor_test.go` covering scalar tensors, multi-dimensional tensors, and empty tensors.

d. Run `go test ./src/main/go/calculation/...` — must pass.

**Acceptance Criteria**:
- `IsScalar()` method exists and is tested
- Existing behavior unchanged

### Task 3: Define ConstraintValidator interface — Not Started

**Goal**: Replace five near-identical constraint merge functions with an interface-driven approach.

**Test-First Development**:

a. Define a `ConstraintValidator` interface in `src/main/go/analysis/constraint_validator.go` with methods like `ValidateMoreRestrictive(other ConstraintValidator) error` and `Merge(other ConstraintValidator) ConstraintValidator`.

b. Implement the interface for each constraint type in `src/main/go/analysis/cpl_types.go` (or new files per type if the package is large):
   - `NumberConstraints`
   - `StringConstraints`
   - `InstantConstraints`
   - `IntervalConstraints`
   - `BooleanConstraints`

c. Write unit tests in `src/main/go/analysis/constraint_validator_test.go` for each implementation.

d. Run `go test ./src/main/go/analysis/...` — must pass.

**Acceptance Criteria**:
- `ConstraintValidator` interface defined
- All five constraint types implement the interface
- Unit tests for each implementation pass

### Task 4: Migrate all operator dispatch to typed registry — Not Started

**Goal**: Replace all string-based operator dispatch with the `Operator` enum.

**Test-First Development**:

a. Change `PostfixOp.Op` from `string` to `Operator` in `src/main/go/calculation/postfix_op.go`.

b. Update `src/main/go/cil/cil_transformer_visitors.go` operator-to-function mapping to use `Operator`.

c. Update `src/main/go/cil/postfix_walk.go` dispatch to use `Operator`.

d. Update `src/main/go/calculation/postfix_op.go` evaluation dispatch to use `Operator`.

e. Update `src/main/go/parser/parser_support.go` parser operator definition types to produce `Operator`. — Superseded: the parser continues to produce `symboltable.DefinitionType` typed enums; the `operatorSymbol()` helper in `metrics_analysis.go` maps them to `operator.Operator` constants at the analysis layer. This avoids coupling the parser to the `operator` package and makes no change to `parser_support.go` necessary.

f. Update `src/main/go/analysis/metrics_analysis.go` CIL building to use `Operator`.

g. Run `go test ./src/main/go/calculation/... ./src/main/go/cil/... ./src/main/go/parser/... ./src/main/go/analysis/...` — must pass.

**Acceptance Criteria**:
- No string-based operator dispatch remains in production code
- `PostfixOp.Op` is typed as `Operator`
- All tests pass

## Constraints

1. The operator registry must be exhaustive — every operator used in codegen, evaluation, and parsing must have a constant
2. `IsScalar()` must be behaviorally identical to the existing inline pattern
3. `ConstraintValidator` must preserve the exact narrowing semantics of the existing functions
4. All existing tests must pass after each task

## Intent

Replace scattered duplicated logic with single sources of truth so that adding a new operator or constraint type requires changes in only one place.

## Acceptance Criteria

1. `Operator` enum is the single source of truth for all operators
2. `IsScalar()` replaces all inline scalar-detection patterns
3. `ConstraintValidator` replaces the five duplicated merge functions
4. `go test ./...` passes with zero failures

## Notes

- Task 4 is the largest because it touches files across `calculation`, `cil`, `parser`, and `analysis` packages. It should only begin after Task 1 completes.
- The constraint validation refactor (Task 3) is independent of the operator registry and can proceed in parallel.
