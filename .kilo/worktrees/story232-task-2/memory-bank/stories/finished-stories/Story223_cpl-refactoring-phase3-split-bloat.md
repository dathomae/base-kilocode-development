# Story223: CPL Refactoring — Phase 3 Split Bloat

## Goal

Decompose the `CPLAnalyzer` god object and the `CILAbstractTransformer` multi-concern struct into focused, single-responsibility sub-components. This improves maintainability and enables limited-context agents to work on individual concerns.

## References

- [CPL Refactoring Plan](../plans/cpl_refactoring_plan.md)
- [Story222: Phase 2 Normalize CIL](Story222_cpl-refactoring-phase2-normalize-cil.md)

## Dependencies

- [Story222_cpl-refactoring-phase2-normalize-cil.md](Story222_cpl-refactoring-phase2-normalize-cil.md) — must be Done. Phase 2 stabilizes the CIL representation and typed bridges, reducing the number of maps and fields that `CPLAnalyzer` must own.

## Dependent Stories

- [Story224_cpl-refactoring-phase4-reduce-duplication.md](Story224_cpl-refactoring-phase4-reduce-duplication.md) — Not Started

## Tasks

### Task Sizing

All tasks in this story are **not small** (medium). Each task introduces a new struct, moves methods across 3+ files within the same package, and requires coordinated changes between struct definition, constructor, scope handlers, and tests. The architect-for-story-planning confirmed that no task can be naturally decomposed further — splitting would create artificial boundaries with no isolation benefit because each extraction is a single cohesive refactor.

### Parallel Execution

- **Group A** (Tasks 1a, 2): Task 1a modifies files in `src/main/go/analysis/`; Task 2 modifies files in `src/main/go/cil/`. The packages are fully disjoint with no compile-time dependency, and the file sets do not overlap.

No other parallel groups are possible. Tasks 1a–1d all modify `src/main/go/analysis/cpl_analyzer.go` (removing fields from the same struct and updating the same constructor), so they must execute sequentially.

### Execution Order

```
[Task 1a ‖ Task 2] → Task 1b → Task 1c → Task 1d → Task 1e
```

### Reintegration

**Group A (Tasks 1a, 2):**
- Merge order: any order (fully disjoint packages, no shared infrastructure)
- Integration tests: `go test ./src/main/go/analysis/... ./src/main/go/cil/... -count=1 -timeout 120s` — verify combined analyzer and transformer behavior
- Watch for conflicts in: none expected

### Task 1a: Extract TypeAnalyzer — Completed

**Goal**: Move type-related maps, derivation path computation, and constraint validation out of `CPLAnalyzer` into a dedicated `TypeAnalyzer`.

**Test-First Development**:

a. In `src/main/go/analysis/type_analyzer.go`, define the `TypeAnalyzer` struct with its `types` map, derivation path computation, and constraint validation methods.

b. In `src/main/go/analysis/type_analyzer_test.go`, write tests for `TypeAnalyzer` covering derivation path computation and constraint validation.

c. Run `go test ./src/main/go/analysis/...` — tests will fail because `CPLAnalyzer` still owns the type-related fields and the new `TypeAnalyzer` is unconnected.

d. Identify all type-related fields and methods in `CPLAnalyzer` (from `src/main/go/analysis/cpl_analyzer.go` and `src/main/go/analysis/cpl_types.go`).

e. Move the identified methods and fields to `TypeAnalyzer`, preserving exact signatures where possible.

f. Update `src/main/go/analysis/cpl_analyzer.go` to embed or delegate to `TypeAnalyzer`.

g. Run `go test ./src/main/go/analysis/...` — must pass.

**Acceptance Criteria**:
- `TypeAnalyzer` exists with all type-related concerns
- `CPLAnalyzer` no longer contains type-specific maps or methods directly
- All tests pass

### Task 1b: Extract ConceptAnalyzer — Completed

**Goal**: Move concept-related maps, range resolution, and extension merging into a dedicated `ConceptAnalyzer`.

**Test-First Development**:

a. In `src/main/go/analysis/concept_analyzer.go`, define the `ConceptAnalyzer` struct with its `concepts` map, range resolution, and extension merging logic.

b. In `src/main/go/analysis/concept_analyzer_test.go`, write tests for `ConceptAnalyzer` covering range resolution and extension merging.

c. Run `go test ./src/main/go/analysis/...` — tests will fail because `CPLAnalyzer` still owns the concept-related fields and the new `ConceptAnalyzer` is unconnected.

d. Identify all concept-related fields and methods in `CPLAnalyzer` (from `src/main/go/analysis/cpl_analyzer.go` and `src/main/go/analysis/concept_analysis.go`).

e. Move the identified methods and fields to `ConceptAnalyzer`.

f. Update `src/main/go/analysis/cpl_analyzer.go` to embed or delegate to `ConceptAnalyzer`.

g. Run `go test ./src/main/go/analysis/...` — must pass.

**Acceptance Criteria**:
- `ConceptAnalyzer` exists with all concept-related concerns
- `CPLAnalyzer` no longer contains concept-specific maps or methods directly
- All tests pass

### Task 1c: Extract MetricAnalyzer — Completed

**Goal**: Move metric-related maps, CIL tree building, and reference validation into a dedicated `MetricAnalyzer`.

**Test-First Development**:

a. In `src/main/go/analysis/metric_analyzer.go`, define the `MetricAnalyzer` struct with its `metrics` map, CIL tree building, and reference validation methods.

b. In `src/main/go/analysis/metric_analyzer_test.go`, write tests for `MetricAnalyzer` covering CIL tree building and reference validation.

c. Run `go test ./src/main/go/analysis/...` — tests will fail because `CPLAnalyzer` still owns the metric-related fields and the new `MetricAnalyzer` is unconnected.

d. Identify all metric-related fields and methods in `CPLAnalyzer` (from `src/main/go/analysis/cpl_analyzer.go` and `src/main/go/analysis/metrics_analysis.go`).

e. Move the identified methods and fields to `MetricAnalyzer`.

f. Update `src/main/go/analysis/cpl_analyzer.go` to embed or delegate to `MetricAnalyzer`.

g. Run `go test ./src/main/go/analysis/...` — must pass.

**Acceptance Criteria**:
- `MetricAnalyzer` exists with all metric-related concerns
- `CPLAnalyzer` no longer contains metric-specific maps or methods directly
- All tests pass

### Task 1d: Extract ConnectorAnalyzer — Completed

**Goal**: Move connector-related maps and grain validation into a dedicated `ConnectorAnalyzer`.

**Test-First Development**:

a. In `src/main/go/analysis/connector_analyzer.go`, define the `ConnectorAnalyzer` struct with its `connectors` map and grain validation logic.

b. In `src/main/go/analysis/connector_analyzer_test.go`, write tests for `ConnectorAnalyzer` covering grain validation.

c. Run `go test ./src/main/go/analysis/...` — tests will fail because `CPLAnalyzer` still owns the connector-related fields and the new `ConnectorAnalyzer` is unconnected.

d. Identify all connector-related fields and methods in `CPLAnalyzer` (from `src/main/go/analysis/cpl_analyzer.go` and `src/main/go/analysis/connectors_analysis.go`).

e. Move the identified methods and fields to `ConnectorAnalyzer`.

f. Update `src/main/go/analysis/cpl_analyzer.go` to embed or delegate to `ConnectorAnalyzer`.

g. Run `go test ./src/main/go/analysis/...` — must pass.

**Acceptance Criteria**:
- `ConnectorAnalyzer` exists with all connector-related concerns
- `CPLAnalyzer` no longer contains connector-specific maps or methods directly
- All tests pass

### Task 1e: CPLAnalyzer becomes thin orchestrator — Completed

**Goal**: Refactor `CPLAnalyzer` to delegate all work to sub-analyzers and only aggregate `GetAnalysisDefinitions()`.

**Test-First Development**:

a. Remove all direct map fields from `CPLAnalyzer` (they now live in sub-analyzers).

b. Ensure `CPLAnalyzer` initializes all sub-analyzers and delegates method calls.

c. Verify `GetAnalysisDefinitions()` aggregates results from all sub-analyzers correctly.

d. Run `go test ./src/main/go/analysis/...` — must pass.

**Acceptance Criteria**:
- `CPLAnalyzer` contains only sub-analyzer references and orchestration logic
- `GetAnalysisDefinitions()` returns the same result as before
- All tests pass

### Task 2: Split CILAbstractTransformer into registry and code generator — Completed

**Goal**: Separate registration, code generation, and tree walking concerns in the CIL transformer.

**Test-First Development**:

a. In `src/main/go/cil/cil_registry.go`, define the `CILRegistry` struct with `RegisterMeasureType`, `RegisterMetricType`, `RegisterConnector`, `GetConnectorMetadata`.

b. In `src/main/go/cil/cil_registry_test.go`, write tests for `CILRegistry` covering registration and metadata retrieval.

c. In `src/main/go/cil/cil_code_generator.go`, define the `CILCodeGenerator` struct with `GenerateComputeFunction`, `GenerateSetupFunction`, `GenerateCleanupFunction`, `GenerateConnectorMetadataMap`.

d. In `src/main/go/cil/cil_code_generator_test.go`, write tests for `CILCodeGenerator` covering code generation methods.

e. Convert `walkToPostfix` and `WalkCILTreeToPostfix` in `src/main/go/cil/postfix_walk.go` to free functions (no receiver).

f. Update `src/main/go/cil/cil_transformer.go` so that `CILAbstractTransformer` composes `CILRegistry` and `CILCodeGenerator`.

g. Run `go test ./src/main/go/cil/...` — must pass.

**Acceptance Criteria**:
- `CILRegistry` and `CILCodeGenerator` exist as separate structs
- `walkToPostfix` is a free function
- `CILAbstractTransformer` composes the new structs
- All tests pass

## Constraints

1. Zero behavior change — all public method signatures preserved where external callers exist
2. Same Go package for all analyzer splits (`package analysis`)
3. Same Go package for transformer split (`package cil`)
4. All existing tests must pass after each task

## Intent

Replace monolithic structs with cohesive sub-components so that each concern can be understood, tested, and modified in isolation.

## Acceptance Criteria

1. `TypeAnalyzer`, `ConceptAnalyzer`, `MetricAnalyzer`, and `ConnectorAnalyzer` exist and own their respective concerns
2. `CPLAnalyzer` is ≤ 300 lines and only orchestrates
3. `CILRegistry` and `CILCodeGenerator` exist as separate structs
4. `walkToPostfix` is a free function
5. `go test ./src/main/go/analysis/... ./src/main/go/cil/...` passes with zero failures

## Notes

- Task 1e must wait for all four sub-analyzer extractions to complete because it removes the direct map fields.
- Task 2 (transformer split) runs in parallel with Task 1a only, since `cil/` and `analysis/` are fully disjoint packages. Tasks 1a–1d cannot run in parallel with each other because they all modify `src/main/go/analysis/cpl_analyzer.go`.
