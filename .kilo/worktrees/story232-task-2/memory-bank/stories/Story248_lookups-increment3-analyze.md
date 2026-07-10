# Story248: Lookups Increment 3 (Analyze)

## Goal

Implement lookup reference resolution during metric analysis, dimensional compatibility validation between lookups and their referencing metrics, and `CILLookupRefNode` generation for CIL expression trees. This story extends the metric analysis phase to handle `InputKind = "lookup"` inputs, resolving lookup names against the `AnalysisDefinitions.Lookups` map, validating that a lookup's key dimensions are available in the metric's dimensional context, and generating `CILLookupRefNode` CIL leaf nodes when a lookup reference appears in a metric's `calculate` expression.

This story is part of [Epic006: CPL Lookups Implementation](../epics/Epic006_cpl-lookups-implementation.md), increment 3 of 5.

## References

- [ERO006: Lookups Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO006_lookups-analysis-phase.md) — Increment 3 lookup input resolution and dimensional validation
- [ERO006: Lookups Incremental Implementation](../specifications/design/cpl/go-analysis/eros/ERO006_lookups-incremental-implementation.md) — Parent ERO defining the 5-increment strategy
- [Metrics Design](../specifications/design/cpl/metrics-design.md) — Metric design (metrics reference lookups as inputs in Increment 3)
- [Symbol Table-Driven Parsing Pattern](../specifications/patterns/symbol-table-driven-parsing.md) — Architectural pattern for phase separation

## Dependencies

- [Story247_lookups-increment3-parse.md](Story247_lookups-increment3-parse.md) — Populates `ParserMetricDefinition.Inputs` with `InputKind = "lookup"` entries from parser
- [Story242_lookups-increment1-analyze.md](Story242_lookups-increment1-analyze.md) — Populates `AnalysisDefinitions.Lookups` map with `LookupInfo` entries and defines `CILLookupNode`

## Dependent Stories

None — Increment 4 (lookups in aggregator expressions) and Increment 5 (lookups in spreader expressions) will depend on this story.

## Tasks

### Task Sizing

Tasks 1 and 2 are annotated `[Small]`: each touches at most 2 files, is purely additive, and follows an existing pattern. Task 1 creates a single struct with an interface marker method (analogous to existing `CILInputRef`). Task 2 adds a single field to an existing struct. Both are single-concern changes that the small-code agent can handle.

Task 3a was decomposed from original Task 3 (which had 9 subtasks spanning two distinct concerns). The decomposition split the self-contained validation preamble (input resolution + dimensional compatibility) from the CIL tree generation changes. Task 3a modifies `validateInput`, `resolveMetricInputs`, and adds `validateLookupMetricCompatibility` — coordinated changes in one source file that require reasoning about the metric analysis pipeline. Task 3b modifies `buildCILTreeFromPostfix` and `collectInputRefsRecursive` — changes to the CIL tree construction pipeline that depend on Task 3a's resolution logic.

Task 4 creates multiple example files across two directories but is mechanically uniform (following existing error/success example patterns). Not annotated `[Small]` because it touches multiple files and requires understanding the example verification infrastructure.

### Parallel Execution

- **Group A** (Tasks 1, 2): Both tasks modify fully disjoint files in the same package. Task 1 creates `cil_lookup_ref_node.go`; Task 2 modifies `analysis_input_definition.go`. Neither depends on the other's output. Both can be independently compiled and tested.

### Execution Order

```
[Task 1 ‖ Task 2] → Task 3a → Task 3b → Task 4
```

### Reintegration

**Group A (Tasks 1, 2):**
- Merge order: any order (fully disjoint files, no shared infrastructure)
- Integration tests: `go test ./src/main/go/analysis/... -count=1 -timeout 60s` — verify both new types compile and coexist in the same package
- Watch for conflicts in: none expected (disjoint files, no shared modifications)

### Task 1: Define CILLookupRefNode type — Not Started [Small]

a. Create `src/main/go/analysis/cil_lookup_ref_node.go` — Not Started
b. Define `CILLookupRefNode` struct with `LookupName` (string), `Dimensions` ([]string), and `Celltype` (string) fields — Not Started
c. Implement `cilNode()` marker method on `CILLookupRefNode` to satisfy the `cil.CILNode` interface — Not Started
d. Write unit tests in `src/main/go/analysis/cil_lookup_ref_node_test.go` verifying struct initialization sets all fields correctly, and that `CILLookupRefNode` satisfies the `cil.CILNode` interface at compile time — Not Started
e. Verify the package compiles with `go build ./src/main/go/analysis/...` — Not Started

#### Intent

Create a CIL leaf node type that represents a lookup value access at runtime. This node is structurally analogous to `CILInputRef` but carries additional metadata (lookup name, dimension context, celltype) needed by the synthesis phase to generate `MetricComputation` controller access code.

#### Acceptance Criteria

- `CILLookupRefNode` exists in `src/main/go/analysis/cil_lookup_ref_node.go`
- `CILLookupRefNode` implements `cil.CILNode` via `cilNode()` marker method
- Unit tests confirm struct field initialization and interface compliance
- `go build ./src/main/go/analysis/...` succeeds

### Task 2: Extend AnalysisInputDefinition for lookup resolution — Not Started [Small]

a. Add `ResolvedLookup *LookupInfo` field to `AnalysisInputDefinition` in `src/main/go/analysis/analysis_input_definition.go` — Not Started
b. Write unit tests in `src/main/go/analysis/analysis_input_definition_test.go` verifying that the default zero value is nil and that the field can be populated with a `*LookupInfo` — Not Started
c. Verify the package compiles with `go build ./src/main/go/analysis/...` — Not Started

#### Intent

Extend the `AnalysisInputDefinition` struct to carry a resolved lookup reference for inputs declared with `InputKind = "lookup"`. The new field `ResolvedLookup` is populated during metric analysis when a lookup input reference is successfully resolved against the `Lookups` map. The zero-value default (nil) preserves backward compatibility with existing metric and dimension inputs.

#### Acceptance Criteria

- `AnalysisInputDefinition` has a `ResolvedLookup *LookupInfo` field
- Unit tests confirm nil default value and successful population
- `go build ./src/main/go/analysis/...` succeeds
- Existing tests continue to pass (no regression from new zero-value field)

### Task 3a: Implement lookup input resolution and dimensional compatibility validation — Not Started

a. Add `case "lookup"` branch to the `validateInput` function in `src/main/go/analysis/metrics_analysis.go`, resolving the lookup name against `analyzer.lookups` and returning an error with source location if the lookup is not found — Not Started
b. In `resolveMetricInputs` in `src/main/go/analysis/metrics_analysis.go`, set `ResolvedLookup` on the `AnalysisInputDefinition` when `Kind == "lookup"` and the lookup is found — Not Started
c. Create `validateLookupMetricCompatibility(lookup *LookupInfo, parserDef *parsing.ParserMetricDefinition) error` function in `src/main/go/analysis/metrics_analysis.go` — Not Started
d. Inside `validateLookupMetricCompatibility`, build a set of the metric's dimension concepts from inputs declared with `InputKind == "dimension"` — Not Started
e. Verify each key dimension of the lookup is present in the metric's dimension set; return a clear error if a dimension is missing: `"lookup <name> is keyed by <dim> but metric does not declare <dim> as a dimension"` — Not Started
f. Write unit tests in `src/main/go/analysis/metrics_analysis_test.go`: valid lookup reference in metric inputs, non-existent lookup produces error with source location, compatible dimensions pass validation, incompatible (missing) dimensions produce errors with lookup and dimension names — Not Started
g. Verify `go test ./src/main/go/analysis/... -count=1` passes — Not Started

#### Intent

Wire lookup reference resolution and dimensional compatibility validation into the metric analysis pipeline. When a metric input has `Kind == "lookup"`, the analyzer resolves the lookup name against the `AnalysisDefinitions.Lookups` map and validates that the lookup's key dimensions are a subset of the metric's declared dimensions. The dimensional check ensures that a metric cannot reference a lookup keyed by dimensions the metric doesn't declare — for example, a metric dimensioned by `[Country, Product]` can reference a lookup keyed by `[Country]` but not one keyed by `[Region]`. This is a structural prerequisite check; full dimensional analysis across the metric dependency DAG is deferred to later synthesis-time analysis.

This task was decomposed from original Task 3. The CIL tree generation portion (CILLookupRefNode insertion in calculate expressions) is handled by Task 3b.

#### Acceptance Criteria

- `case "lookup"` in `validateInput` resolves lookup name and reports errors for non-existent lookups with source location
- `resolveMetricInputs` sets `ResolvedLookup` on `AnalysisInputDefinition` for lookup inputs
- `validateLookupMetricCompatibility` checks that all lookup key dimensions are declared as metric dimensions
- Dimensional incompatibility produces clear error messages with lookup name and missing dimension name
- All tests pass with `go test ./src/main/go/analysis/... -count=1`

### Task 3b: Generate CILLookupRefNode in CIL expression trees for lookup references — Not Started

a. In `buildCILTreeFromPostfix` in `src/main/go/analysis/metrics_analysis.go`, when a `PostfixIdentifier` resolves to a lookup input reference, generate a `CILLookupRefNode` instead of a `CILInputRef` — Not Started
b. Update `collectInputRefsRecursive` in `src/main/go/analysis/metrics_analysis.go` to handle `*CILLookupRefNode` by collecting it as an input reference (with kind `"lookup"`) — Not Started
c. Update `isNumericNode` in `src/main/go/analysis/metrics_analysis.go` to return true for `*CILLookupRefNode` — Not Started
d. Update `validateCILTreeWalk` in `src/main/go/analysis/metrics_analysis.go` to handle `*CILLookupRefNode` as a valid leaf node — Not Started
e. Write unit tests in `src/main/go/analysis/metrics_analysis_test.go`: lookup reference in calculate expression generates CILLookupRefNode with correct fields, CILLookupRefNode is collected as an input ref with kind "lookup", CILLookupRefNode is treated as a numeric node in expression validation — Not Started
f. Verify `go test ./src/main/go/analysis/... -count=1` passes — Not Started

#### Intent

Integrate `CILLookupRefNode` into the CIL expression tree construction pipeline. When the postfix evaluator encounters an identifier that resolves to a lookup input (as determined by Task 3a's resolution logic), it generates a `CILLookupRefNode` carrying the lookup's name, dimension context, and celltype. The CIL tree walker and validator must recognize this new node type to avoid falling through to the "unknown node type" error case.

This task was decomposed from original Task 3. It depends on Task 3a for the lookup resolution logic that determines which identifiers correspond to lookup inputs.

#### Acceptance Criteria

- `buildCILTreeFromPostfix` generates `CILLookupRefNode` for lookup identifiers in calculate expressions
- `collectInputRefsRecursive` handles `CILLookupRefNode` and collects it as an input ref with kind "lookup"
- `isNumericNode` returns true for `CILLookupRefNode`
- `validateCILTreeWalk` does not produce "unknown node type" for `CILLookupRefNode`
- All tests pass with `go test ./src/main/go/analysis/... -count=1`

### Task 4: Create parse_and_analyze_only examples with lookup inputs — Not Started

a. In `examples/parse_and_analyze_only/with_errors/`: add CPL files with metrics referencing non-existent lookups, and metrics with lookups whose key dimensions are not declared as metric dimensions — Not Started
b. In `examples/parse_and_analyze_only/without_errors/`: add CPL files with metrics correctly referencing lookups with compatible dimensions, and a metric that uses a lookup reference in its `calculate` expression — Not Started
c. Verify all error examples produce the expected analysis errors, and all success examples complete analysis without errors — Not Started

#### Intent

Extend the `parse_and_analyze_only` example suite to cover lookup input scenarios. Error examples test the analyzer's error reporting for non-existent lookup references and dimensional incompatibility. Success examples verify the happy path for lookup reference resolution and CIL tree construction with lookup inputs.

#### Acceptance Criteria

- At least 2 error example CPL files added: one for non-existent lookup reference, one for dimensional incompatibility
- At least 2 success example CPL files added: one for compatible lookup dimensions, one for lookup reference in calculate expression
- All error examples produce the expected analysis errors
- All success examples complete analysis without errors

## Test-First Development

This project follows the Logical TDD Lifecycle. For each coding task, write tests before implementing the code. The test-and-implement cycle is integrated within each task — do not separate tests into standalone tasks.

## Constraints

- Must work with the `AnalysisDefinitions.Lookups` map populated by Story242
- Must follow the Symbol Table-Driven Parsing Pattern for phase separation
- The `CILLookupRefNode` must implement the `cil.CILNode` interface
- Error messages must include the metric name, lookup name, and source location where possible
- `validateLookupMetricCompatibility` is a preliminary structural check only; full unit algebra verification is deferred to synthesis
- The `resolveInputRefKind` function in `metrics_analysis.go` must also recognize lookup names to determine the correct input kind for short-form metrics
- Existing tests must continue to pass — new fields must have zero-value defaults that do not break existing behavior
- All CIL tree walker/validator functions (`isNumericNode`, `collectInputRefsRecursive`, `validateCILTreeWalk`) must handle `CILLookupRefNode` to avoid "unknown node type" errors

## Intent

This story extends the metric analysis phase to handle `lookup` input references. When a metric input has `InputKind == "lookup"`, the analyzer resolves the lookup name against the `AnalysisDefinitions.Lookups` map, validates that the lookup's key dimensions are available in the metric's dimensional context, and generates a `CILLookupRefNode` for use in the metric's CIL tree. The dimensional compatibility check ensures that a metric cannot reference a lookup keyed by dimensions the metric doesn't declare — for example, a metric dimensioned by `[Country, Product]` can reference a lookup keyed by `[Country]` but not one keyed by `[Region]`. This is a structural prerequisite check; full dimensional analysis across the metric dependency DAG is deferred to later synthesis-time analysis.

## Acceptance Criteria

1. `AnalysisInputDefinition.ResolvedLookup` populated when `InputKind == "lookup"`
2. `CILLookupRefNode` implements `CILNode` and stores lookup reference metadata (lookup name, dimension context, celltype)
3. Metric analysis resolves lookup input references against `AnalysisDefinitions.Lookups`
4. Non-existent lookup references produce clear error messages with source locations
5. `validateLookupMetricCompatibility` ensures lookup key dimensions are a subset of metric dimensions
6. Dimensional incompatibility produces clear errors with lookup name and missing dimension name
7. Lookups with grain are correctly handled in dimensional analysis
8. `buildCILTreeFromPostfix` generates `CILLookupRefNode` for lookup identifiers in calculate expressions
9. `collectInputRefsRecursive` collects lookup references as input refs with kind "lookup"
10. `validateCILTreeWalk` accepts `CILLookupRefNode` as a valid leaf node
