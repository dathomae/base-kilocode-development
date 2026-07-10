# Story249: Lookups Increment 3 — Synthesize

## Goal

Implement CIL tree walker for CILLookupRefNode, GetMetricComputation(lookupName) emission, unit compatibility verification, and metric+lookup examples. This story delivers the synthesis phase for lookup integration with metric calculations — when a metric's CIL tree contains a CILLookupRefNode (produced by the Increment 3 analysis phase), the walker emits Go code that calls controller.GetMetricComputation(lookupName).GetCell(...) to retrieve the lookup value for the current metric cell's dimensional context. Cell key coordinates are constructed from the metric's loop iteration variables, the lookup value is registered as a variable for subsequent expression evaluation, and the preamble generator ensures lookup data loading occurs before any metric that references a lookup runs.

This story is part of [Epic006: CPL Lookups Implementation](../epics/Epic006_cpl-lookups-implementation.md), increment 3 of 5.

## References

- [ERO006: Lookups Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO006_lookups-synthesis-phase.md) — Increment 3 CILLookupRefNode walking and code generation
- [ERO006: Lookups Incremental Implementation](../specifications/design/cpl/eros/ERO006_lookups-incremental-implementation.md) — Parent ERO defining the 5-increment strategy
- [Metrics Design](../specifications/design/cpl/metrics-design.md)
- [Lookups Design](../specifications/design/cpl/lookups.md)
- [Plan: Story249 Lookups Increment 3 Synthesize](../plans/Story249_lookups-increment3-synthesize.md)

## Dependencies

- [Story248_lookups-increment3-analyze.md](Story248_lookups-increment3-analyze.md) — Produces CILLookupRefNode in metric CIL trees during analysis
- [Story243_lookups-increment1-synthesize.md](Story243_lookups-increment1-synthesize.md) — Lookup loading infrastructure and controller registration already in place

## Dependent Stories

- [Story250_lookups-increment4-analyze.md](Story250_lookups-increment4-analyze.md) — Lookups in aggregator expressions (Increment 4 analyze)
- [Story252_lookups-increment5-analyze.md](Story252_lookups-increment5-analyze.md) — Lookups in spreader expressions (Increment 5 analyze)

## Tasks

### Task Sizing

Task 1 introduces a coordinated multi-step walker method (`WalkCILLookupRefNode`) that requires reasoning about loop iteration variables, cell key coordinate construction, variable registration, and operand push onto the expression stack. The method, its `cellKeyCoordinates` helper, and all supporting code form a tightly coupled unit in `cil_tree_walker.go`. Although the task touches only 2 files (walker source + test), the scope requires understanding multiple interacting structures (dimension loop variables, coordinate tuples, expression evaluation stacks) and is not a simple additive case to an existing switch. Not annotated `[Small]`.

Tasks 2, 3, and 4 are annotated `[Small]`. Task 2 adds a single helper function (`verifyLookupCelltypeCompatibility`) to an existing file with its test — a straightforward structural check. Task 3 modifies the preamble generator to order lookup loading before metric computation — a single concern in one source file with its test. Task 4 creates one example directory following the established `with_data` example template — mechanically uniform.

Task 5 is a regression verification task involving running the full CPL test suite and confirming existing lookup examples still pass. Not annotated `[Small]` — it requires navigating the story-implementor's integration workflow rather than acting as a single standalone code change.

### Parallel Execution

- **Group A** (Tasks 2, 3): Task 2 modifies `src/main/go/synthesis/lookup_synthesis.go` and `src/main/go/synthesis/lookup_synthesis_test.go`; Task 3 modifies `src/main/go/synthesis/preamble_generator.go` and its test file `src/main/go/synthesis/preamble_generator_test.go`. The file sets are completely disjoint within the same package. Neither task depends on the other's output — celltype verification and preamble ordering are independent concerns. Each can be independently tested and verified.

### Execution Order

```
Task 1 → [Task 2 ‖ Task 3] → Task 4 → Task 5
```

- Task 1 (CIL walker) must complete first — Tasks 2 and 3 depend on the walker scope and code generation infrastructure it provides
- Tasks 2 and 3 run in parallel (Group A) after Task 1
- Task 4 (examples) depends on Tasks 1–3 (needs full synthesis pipeline)
- Task 5 (regression) depends on all preceding tasks

### Reintegration

**Group A (Tasks 2, 3):**
- Merge order: any order (fully disjoint files, no shared infrastructure)
- Integration tests: `go test ./src/main/go/synthesis/... -count=1 -timeout 60s` — verify celltype verification and preamble ordering compile and test together in the synthesis package
- Watch for conflicts in: `go.sum` (dependency resolution in separate worktrees may produce different hashes)

### Test-First Development

This project follows the Logical TDD Lifecycle. For all coding tasks (Tasks 1–4), write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Tests are integrated into each task, not broken out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

### Task 1: Implement WalkCILLookupRefNode in the CIL tree walker — Not Started

**Intent**: Extend the CIL tree walker to handle `CILLookupRefNode` nodes encountered in metric `calculate` expression trees. When the walker visits a `CILLookupRefNode`, it generates Go code that calls `controller.GetMetricComputation(lookupName).GetCell(...)` to retrieve the lookup value for the current cell's dimensional coordinates. The lookup value is registered as a variable for subsequent expression references and pushed onto the operand stack.

a. Write unit tests in `src/main/go/synthesis/lookup_synthesis_test.go` verifying the emitted Go code for lookup reference access, correct cell key coordinate construction from loop variables, proper variable registration, and operand push onto expression stack — Not Started

b. Add `WalkCILLookupRefNode` method to the CIL tree walker in `src/main/go/synthesis/cil_tree_walker.go` that handles the `CILLookupRefNode` case in the walker's dispatch — Not Started

c. Implement `cellKeyCoordinates` helper method on the walker scope that constructs coordinate tuples from metric loop iteration variables for the lookup's dimension set — Not Started

d. Generate unique variable name for the lookup value via `scope.lookupVarName(node.LookupName)` — Not Started

e. Emit Go code `controller.GetMetricComputation(lookupName).GetCell(coordinates)` to retrieve the lookup cell value — Not Started

f. Register the lookup variable for subsequent expression references via `scope.registerVariable(node.LookupName, varName)` — Not Started

g. Push the lookup value operand onto the expression stack via `scope.pushOperand(varName)` — Not Started

h. Verify all walker tests pass: `go test ./src/main/go/synthesis/... -count=1 -timeout 60s` — Not Started

#### Acceptance Criteria

1. `WalkCILLookupRefNode` emits Go code calling `controller.GetMetricComputation(lookupName).GetCell(...)`
2. Cell key coordinates are correctly constructed from metric loop iteration variables matching the lookup's dimensions
3. The lookup value variable is registered for subsequent expression references in the `calculate` expression
4. The lookup value operand is pushed onto the expression stack for evaluation by the postfix evaluator
5. All walker unit tests pass with zero failures

### Task 2: Implement celltype compatibility verification — Not Started [Small]

**Intent**: Add a `verifyLookupCelltypeCompatibility` function that performs a structural check ensuring the lookup's celltype units are compatible with the metric calculation. Full unit algebra verification is done in analysis; synthesis performs a final structural pass to catch any mismatches missed during analysis that would produce incorrect generated code.

a. Write unit tests in `src/main/go/synthesis/lookup_synthesis_test.go` covering compatible units and incompatible units (expect appropriate error/panic) — Not Started

b. Implement `verifyLookupCelltypeCompatibility` helper function in `src/main/go/synthesis/lookup_synthesis.go` that compares the lookup celltype against the metric's expected celltype at the synthesis code emission point — Not Started

c. Verify tests pass: `go test ./src/main/go/synthesis/... -count=1 -timeout 60s` — Not Started

#### Acceptance Criteria

1. Compatible lookup-metric celltype pairs pass verification silently
2. Incompatible celltype pairs produce a clear error message identifying the mismatch
3. All unit tests pass

### Task 3: Update preamble generator for lookup-to-metric ordering — Not Started [Small]

**Intent**: Ensure the generated `main()` function places `loadLookupData` calls before metric computations that reference lookups. This ordering is critical for the `controller.GetMetricComputation(lookupName)` call to succeed at runtime — the lookup must be loaded and registered on the controller before any metric accesses it.

a. Write a unit test in `src/main/go/synthesis/preamble_generator_test.go` verifying that, in the generated preamble, lookup loading calls appear before any metric computation that references a lookup — Not Started

b. Modify `src/main/go/synthesis/preamble_generator.go` to enforce that lookup loading precedes metric computations when a metric's CIL tree contains `CILLookupRefNode` references — Not Started

c. Verify test passes: `go test ./src/main/go/synthesis/... -count=1 -timeout 60s` — Not Started

#### Acceptance Criteria

1. Generated preamble places `loadLookupData` before metric computation functions that reference lookups
2. Preamble generation does not reorder existing non-lookup initialization
3. Unit test verifies the ordering constraint

### Task 4: Create with_data/lookups_in_metrics example — Not Started [Small]

**Intent**: Create an end-to-end example demonstrating a metric that references a lookup. The metric `converted_revenue` takes `metric revenue` and `lookup currency_conversion` as inputs, and calculates `revenue * currency_conversion`. The lookup provides currency conversion factors, and the expected output reflects the applied conversion. This example validates the full pipeline: parse → analyze → synthesize → run → output matches expected.

a. Create the example directory at `src/chronocone_planning_language/examples/with_data/lookups_in_metrics/` — Not Started

b. Create `converted_revenue.cpl` CPL file defining metric `converted_revenue` with inputs `metric revenue` and `lookup currency_conversion`, and calculate expression `revenue * currency_conversion` — Not Started

c. Create `lookup_io_spec.json` mapping `currency_conversion` to a parquet file — Not Started

d. Create parquet data files: `converted_revenue_data.parquet` for the metric input data, and `currency_conversion.parquet` for the lookup data — Not Started

e. Create `.expected.json` with the expected output after applying conversion factors — Not Started

f. Verify end-to-end: parse → analyze → synthesize → run → output matches expected using the CPL test driver — Not Started

#### Acceptance Criteria

1. CPL file parses without errors, with `currency_conversion` recognized as a `lookup` input
2. Analysis resolves the lookup reference and generates a `CILLookupRefNode` in the metric's CIL tree
3. Synthesis emits Go code calling `controller.GetMetricComputation("currency_conversion").GetCell(...)`
4. Generated Go code compiles and runs without errors
5. Actual output matches `.expected.json`

### Task 5: Verification and regression — Not Started

**Intent**: Confirm that existing lookup examples from Increments 1–2 still work correctly and that no regressions are introduced across the full CPL test suite. This task validates that Increment 3 synthesis changes do not break previously delivered functionality.

a. Verify existing timeless and temporal lookup examples from Increments 1–2 still pass: `go test ./src/main/go/... -count=1 -timeout 120s` — Not Started

b. Run the full CPL test suite to confirm zero regressions, paying particular attention to synthesis tests, integration tests, and `with_data` examples: `go test ./src/main/go/... -count=1 -timeout 120s` — Not Started

#### Acceptance Criteria

1. All existing lookup examples from Increments 1–2 pass without modification
2. Full CPL test suite passes with zero failures
3. No unexpected test output differences or newly skipped tests

## Constraints

- The `WalkCILLookupRefNode` method must integrate with the existing CIL tree walker dispatch pattern — it adds a new node type case to an existing switch, not a new walker interface
- Cell key coordinate construction must correctly handle multi-dimensional lookups where the lookup dimensions are a subset of the metric's iteration dimensions
- The generated Go code must compile against the existing `ChronoconeComputationController` interface (which already supports `GetMetricComputation` from Increment 1 synthesis)
- Preamble ordering must not break existing lookup-only (no metric reference) examples from Increments 1–2
- All file paths in task descriptions are relative to `source_root` (`src/chronocone_planning_language`)

## Notes

- The `cellKeyCoordinates` helper generates comma-separated loop variable names that form a coordinate tuple for indexing into the lookup tensor: `GetCell(dim1LoopVar, dim2LoopVar)`
- The generated code pattern for a lookup reference follows the form shown in the plan document (Section: Generated Code Example): `varName := controller.GetMetricComputation("lookupName").GetCell(coordinates)` followed by `scope.pushOperand(varName)`
- The preamble generator change is a straightforward ordering constraint — it does not require new code generation infrastructure, only reordering existing emission calls
