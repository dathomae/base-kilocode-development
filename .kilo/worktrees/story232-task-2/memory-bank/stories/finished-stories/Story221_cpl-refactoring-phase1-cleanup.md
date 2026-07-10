# Story221: CPL Refactoring — Phase 1 Cleanup

## Goal

Remove dead code, deprecated fields, and hardcoded magic values from the CPL compiler codebase. This is a low-risk preparatory phase that reduces noise and simplifies the codebase before higher-risk structural changes in subsequent phases.

## References

- [CPL Refactoring Plan](../plans/cpl_refactoring_plan.md)
- [requirements](../specifications/requirements/requirements.md)

## Dependencies

None — This is the foundational cleanup phase. It must complete before Phases 2–6 begin because removing dead code reduces the surface area of later refactors.

## Dependent Stories

- [Story222_cpl-refactoring-phase2-normalize-cil.md](Story222_cpl-refactoring-phase2-normalize-cil.md) — Not Started

## Tasks

### Task Sizing

Tasks 1, 2, and 4 are narrow-scope mechanical removals and qualify as `[Small]`. Task 3 migrates readers of deprecated fields and requires coordination between parser and analysis packages.

### Parallel Execution

- **Group A** (Tasks 1, 2, 4): These tasks modify disjoint files in different packages with no dependencies on each other's output.

### Execution Order

```
[Task 1 ‖ Task 2 ‖ Task 4] → Task 3
```

### Task 1: Remove placeholder/dead methods in CIL transformer — Completed

**Goal**: Eliminate placeholder methods and dead types that confuse maintenance.

**Test-First Development**:

a. Before deleting, run `go test ./cil/...` and record the test count.

b. Verify zero production callers for each target using grep:

   **Confirm dead — zero production callers:**
   - `CILAbstractTransformer.Transform(root interface{})` in `cil_transformer.go`
   - `CILAbstractTransformer.TransformCILTrees` in `cil_transformer.go`
   - `ParallelOperationGroup` type and all methods (`PushComputationGroup`, `PopComputationGroup`, `AddDependency`, `Execute`) in `cil_transformer.go` and `cil_transformer_computation.go`
   - `ComputationGroup` type and all methods (`PushComputation`, `PopComputation`, `GenerateFunction`) in `cil_transformer.go` and `cil_transformer_computation.go`
   - `parallelGroups []*ParallelOperationGroup` field on `CILAbstractTransformer` — never populated, zero production readers
   - `CILTree interface{}` placeholder field in `connector/connector_metadata.go:37` — never populated, dead

   **Call-site cleanup required before removal:**
   - `CILAbstractTransformer.VisitNode(node interface{})` in `cil_transformer.go` — has three production callers in `cil_transformer_visitors.go` (`default:` cases in `VisitCILBinaryOp` at lines ~80 and ~103, and `VisitCILUnaryOp` at line ~215). These fallback calls are no-ops (VisitNode returns nil), so removing the `default:` cases is behavior-preserving. Remove the `default:` cases, then remove `VisitNode`.

   **Actively used — do NOT remove:**
   - `Computation` type and its methods (`PushSnippet`, `PopSnippet`, `AssembleCode`, `GenerateTempVar`, `NewComputation`) are used by `TransformMetric` → `synthesis/synthesize.go` via `cil_transformer_visitors.go`. This is the active codegen path. The `Computation` type must be preserved for Phase 1.

c. Remove:
   - `Transform`, `VisitNode`, and `TransformCILTrees` methods from `CILAbstractTransformer` in `cil_transformer.go`
   - The three `default:` fallback cases that called `VisitNode` in `VisitCILBinaryOp` and `VisitCILUnaryOp` in `cil_transformer_visitors.go`
   - `ParallelOperationGroup` type, `ComputationGroup` type, and all their methods from `cil_transformer.go` and `cil_transformer_computation.go`
   - `parallelGroups` field from `CILAbstractTransformer` in `cil_transformer.go`
   - `CILTree` placeholder field from `connector/connector_metadata.go`
   - Associated test-only callers in `cil_transformer_core_test.go`

d. Run `go test ./cil/... ./connector/...` — must pass with same test count (or fewer if tests of dead code were removed).

**Acceptance Criteria**:
- `Transform`, `VisitNode`, `TransformCILTrees` methods removed from `CILAbstractTransformer`
- `default:` fallback calls to `VisitNode` removed from `VisitCILBinaryOp` and `VisitCILUnaryOp` in `cil_transformer_visitors.go`
- `ParallelOperationGroup` and `ComputationGroup` types removed
- `parallelGroups` field removed from `CILAbstractTransformer`
- `Computation` type preserved (actively used by `TransformMetric` — will be addressed in Phase 5)
- `CILTree` placeholder field removed from `connector/connector_metadata.go`
- All tests in `./cil/...` and `./connector/...` pass

### Task 2: Remove dead CIL node types from analysis package — Completed [Small]

**Goal**: Remove old CIL node types never constructed in production code.

**Test-First Development**:

a. Run `go test ./analysis/... ./cil/... ./synthesis/...` and record test count.

b. Verify zero construction sites for each target using grep:
   - `CILDataConstantNode`
   - `CILUnitNameConstantNode`
   - `CILMeasureNameConstantNode`
   - `CILAbstractOperationNode`
   - `CILMetricReferenceNode`

c. Remove the type definitions from `analysis/cil_nodes.go` and any tests that reference only these types.

d. Run `go test ./analysis/... ./cil/... ./synthesis/...` — must pass.

**Acceptance Criteria**:
- Five dead node types removed from `analysis/cil_nodes.go`
- No test failures in dependent packages

### Task 3: Remove deprecated GrainName/GrainLocation fields — Completed

**Goal**: Eliminate deprecated single-grain fields that coexist with multi-grain arrays.

**Test-First Development**:

a. Audit all readers of `GrainName` and `GrainLocation` in:
   - `symboltable/symbol_table.go`
   - `parser/cpl_connectors_listener.go`
   - `analysis/connectors_analysis.go`
   - `synthesis/synthesize.go`
   - `connector/connector_metadata.go`

b. For each reader, verify that `GrainNames`/`GrainLocations` are already being populated by the parser. If any reader still relies on the deprecated fields, migrate it to use the array fields.

c. Remove `GrainName` and `GrainLocation` fields from `ConnectorDeclaration` in `symboltable/symbol_table.go`.

d. Update any parser code that sets these deprecated fields to set the array fields instead.

e. Run `go test ./parser/... ./analysis/... ./synthesis/... ./connector/...` — must pass.

**Acceptance Criteria**:
- `GrainName` and `GrainLocation` fields removed from `ConnectorDeclaration`
- All readers migrated to `GrainNames`/`GrainLocations`
- No test failures

### Task 4: Address SpacesPerTab hardcoding — Completed [Small]

**Goal**: Make tab spacing configurable or document it as an invariant.

**Test-First Development**:

a. Identify all usages of `SpacesPerTab` in `symboltable/symbol_table.go` and `parser/parser_support.go`.

b. Determine whether configurability is desired (consult existing conventions in the codebase).

c. If configurable: introduce a `ParserConfig` struct with a `SpacesPerTab` field, thread it through the parser listener constructors, and default to 4.

d. If invariant: convert `SpacesPerTab` to a documented package-level constant with a comment explaining it is intentionally fixed.

e. Run `go test ./parser/... ./symboltable/...` — must pass.

**Acceptance Criteria**:
- `SpacesPerTab` either moved to a configuration struct or documented as an invariant
- All existing tests pass
- No behavioral change (default remains 4)

## Constraints

1. Zero behavior change — all removals must be verified dead before deletion
2. All existing tests must pass after each task
3. Changes are limited to the `analysis`, `cil`, `symboltable`, `parser`, `connector`, and `synthesis` packages

## Intent

Reduce maintenance burden and compiler surface area by removing code that is never executed, fields that are superseded, and magic values that are unexplained.

## Acceptance Criteria

1. Dead CIL transformer methods and types removed
2. Dead CIL node types from `analysis/cil_nodes.go` removed
3. Deprecated `GrainName`/`GrainLocation` fields removed and all readers migrated
4. `SpacesPerTab` either configurable or explicitly documented as invariant
5. `go test ./...` from the CPL Go root passes with zero failures

## Notes

- This phase intentionally stays low-risk. If any dead-code removal reveals unexpected callers, pause and reassess before continuing.
- Task 3 is the most complex because it spans parser → analysis → synthesis. The other three tasks are independent and low-risk.
