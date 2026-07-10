# Story243: Lookups Increment 1 — Synthesize

## Goal

Implement CILAbstractTransformer VisitCILTreeNode for CILLookupNode, synthesis of LookupTensorLoader lookup loader and computeX functions

This story delivers the synthesis phase for basic lookup declarations without grain (Increment 1 of [Epic006: CPL Lookups Implementation](../epics/Epic006_cpl-lookups-implementation.md)). The CIL tree walker handles `CILLookupNode` leaf nodes, emitting Go code that reads lookup data from parquet files (specified in the Lookup IO Specification) and registers each lookup as a `MetricComputation` on the `ChronoconeComputationController`. The synthesis reuses the existing parquet reading and controller infrastructure from connector synthesis. In Increment 1, lookups are loaded but not yet accessed by metric calculations — that integration is added in Increment 3. The generated preamble ensures lookup loading occurs before any metric computation runs.

## References

- [Lookups Incremental Implementation ERO](../specifications/design/cpl/eros/ERO006_lookups-incremental-implementation.md)
- [Lookups Synthesis Phase ERO](../specifications/design/cpl/go-synthesis-target/eros/ERO006_lookups-synthesis-phase.md)
- [Lookups Design](../specifications/design/cpl/lookups.md)
- [Epic006: CPL Lookups Implementation](../epics/Epic006_cpl-lookups-implementation.md)
- [Plan: Story243 Lookups Increment 1 Synthesize](../plans/Story243_lookups-increment1-synthesize.md)

## Dependencies

- Story242 (`lookups-increment1-analyze`) — needs `CILLookupNode` from analysis phase. Not yet created.
- Epic005 (metrics) Increment 1 — CIL tree walking infrastructure must exist (`cil_tree_walker.go`, `preamble_generator.go`, `main_generator.go`).

## Dependent Stories

- `lookups-increment2-synthesize` — extends synthesis with grain handling.
- `lookups-increment3-synthesize` — adds lookup reference resolution in metric calculations.

## Tasks

### Task Sizing

Task 2 is annotated `[Small]` — it adds a single walker case to an existing CIL tree walker switch statement, touching at most 2 files (walker source + test), following the established "add node type case" mechanical pattern. All other tasks are `[Medium]`: Task 1 creates new parsing types and logic (not additive to existing function), Task 3 creates a new code generation module with template emission and multiple code paths, Task 4 touches 3 files with coordinated changes across preamble and main generators, Task 5 creates a multi-file example directory spanning CPL, JSON, parquet, and expected output files, and Task 6 is a regression verification task spanning existing test suites.

### Parallel Execution

- **Group A** (Tasks 1, 2): Task 1 creates `src/main/go/synthesis/lookup_io_spec.go` (new file + test); Task 2 modifies `src/main/go/synthesis/cil_tree_walker.go` (adds walker case + test). Completely disjoint files within the same package. Neither task depends on the other's output. Each can be independently tested and verified.

### Execution Order

```
[Task 1 ‖ Task 2] → Task 3 → Task 4 → Task 5 → Task 6
```

- Tasks 1 and 2 run in parallel (Group A)
- Task 3 (code generation) depends on Task 2 (needs walker metadata to know which lookups exist)
- Task 4 (preamble wiring) depends on Task 3 (needs `generateLookupLoading` function to exist)
- Task 5 (examples) depends on Tasks 3 and 4 (needs full synthesis pipeline)
- Task 6 (regression) depends on all preceding tasks

### Reintegration

**Group A (Tasks 1, 2):**
- Merge order: any order (fully disjoint files, no shared infrastructure)
- Integration tests: `go test ./src/main/go/synthesis/... -count=1 -timeout 60s` — verify IO spec parsing and CIL walker compile and test together in the synthesis package
- Watch for conflicts in: none expected (each task touches distinct files; neither modifies package-level declarations or shared test fixtures)

### Task 1: Implement Lookup IO Specification parsing — Completed

**Intent**: Define the JSON schema for the Lookup IO Specification (mapping lookup names to parquet file paths) and implement parsing logic that produces a `LookupIOSpec` struct. This enables the synthesis pipeline to locate lookup data files at code generation time.

a. Define the Lookup IO Specification JSON schema with `lookups` array containing `name` and `source` (type/path) fields — Not Started

b. Create `src/main/go/synthesis/lookup_io_spec.go` with parsing logic:
   - Define `LookupIOSpec` struct with `Lookups []LookupEntry`
   - Define `LookupEntry` struct with `Name` string and `Source LookupSource`
   - Define `LookupSource` struct with `Type` string and `Path` string
   - Implement `ParseLookupIOSpec(filename string) (*LookupIOSpec, error)` that reads and unmarshals the JSON file
   - Return descriptive errors for missing files, invalid JSON, and empty lookups lists
   — Not Started

c. Write unit tests in `src/main/go/synthesis/lookup_io_spec_test.go`:
   - Test parsing a valid lookup IO spec with one lookup entry
   - Test parsing a spec with multiple lookup entries
   - Test error on missing file
   - Test error on malformed JSON
   - Test error on missing required fields (name, source.path)
   — Not Started

d. Run `go test ./src/main/go/synthesis/... -count=1 -timeout 60s` and verify all IO spec tests pass — Not Started

**Acceptance Criteria**:
- `ParseLookupIOSpec` successfully parses a valid JSON file and returns a populated `LookupIOSpec`
- All error cases (missing file, malformed JSON, missing fields) are handled with descriptive error messages
- All unit tests pass

### Task 2: Implement lookup registration in the CIL tree walker — Completed [Small]

**Intent**: Extend the existing CIL tree walker (`CILTreeWalker`) to discover and register lookup metadata from `CILLookupNode` entries stored in `analysisDefs.CILTrees`. The walker already handles connectors and measure types; lookups are registered by a new `registerLookups` helper that type-dispatches on `CILLookupNode` instances keyed with the `"lookup:"` prefix. This is a mechanical addition — lookup nodes are leaves, so no recursive subtree walking is required.

a. Add `registerLookups` helper method to `src/main/go/synthesis/cil_tree_walker.go`:
   - Iterate `analysisDefs.CILTrees` and filter for `*cil.CILLookupNode` entries
   - Verify the map key has the `"lookup:"` prefix; skip entries without it
   - Extract `lookupName` (trimming the prefix), `Dimensions`, and `Celltype`
   - Call `transformer.RegisterLookup(...)` to store metadata in the walker's `CILRegistry`
   — Not Started

b. Wire `registerLookups` into the existing `WalkCILTrees` method, invoked after connector registration and before measure type registration
   — Not Started

c. Write unit test in `src/main/go/synthesis/cil_tree_walker_test.go`:
   - Construct a minimal `CILLookupNode` with a lookup name and dimension concepts
   - Walk the node using the CIL tree walker
   - Verify the walker scope (transformer registry) is populated with the expected lookup metadata (name and dimension concept names present)
   - Verify entries without `"lookup:"` prefix are silently skipped
   — Not Started

d. Run `go test ./src/main/go/synthesis/... -count=1 -timeout 60s` and verify walker test passes — Not Started

**Acceptance Criteria**:
- `registerLookups` registers the lookup name and its dimension concepts in the walker's transformer registry
- The walker scope is queryable for lookup metadata after walking
- Walker test passes
- No regressions in existing walker tests
- The walker scope is queryable for lookup metadata after walking
- Walker test passes
- No regressions in existing walker tests

### Task 3: Implement lookup code generation (generateLookupLoading) — Completed

**Intent**: Create the `generateLookupLoading` function that emits Go source code for reading lookup data from parquet files and registering each lookup as a `MetricComputation` on the `ChronoconeComputationController`. The emitted code reuses the existing `readParquetFile` helper and controller infrastructure from connector synthesis.

a. Create `src/main/go/synthesis/lookup_synthesis.go` with the `generateLookupLoading` function — Not Started

b. Emit Go code for each lookup that:
   - Reads the parquet file from the path specified in the Lookup IO Spec
   - Builds a tensor with dimension names matching the lookup's key concept names
   - Registers the tensor as a `MetricComputation` via `controller.AddMetricComputation(...)`
   - Cell type defaults to `float64` (type selection is deferred per the ERO)
   — Not Started

c. Emit dimension name registration on the tensor using the lookup's key concept names (grain not yet applicable in Increment 1) — Not Started

d. Handle the empty lookups case: when no lookups are declared, emit a no-op `loadLookupData` function that returns `nil` — Not Started

e. Write unit tests in `src/main/go/synthesis/lookup_synthesis_test.go`:
   - Test that `generateLookupLoading` emits valid Go code for a single lookup with one dimension
   - Test code emission for a lookup with multiple dimension concepts
   - Test that the emitted code includes `readParquetFile`, tensor construction, and `AddMetricComputation` calls
   - Test the empty lookups case (emits a no-op function)
   — Not Started

f. Run `go test ./src/main/go/synthesis/... -count=1 -timeout 60s` and verify all lookup synthesis tests pass — Not Started

**Acceptance Criteria**:
- `generateLookupLoading` emits valid, compilable Go code
- Emitted code calls `readParquetFile` with the correct parquet file path
- Emitted code registers a `MetricComputation` on the controller for each lookup
- Tensor dimensions are the lookup's key concept names (no grain)
- Empty lookups produce a no-op function, not a compile error
- All unit tests pass

### Task 4: Wire lookup loading into the preamble and main generator — Completed

**Intent**: Integrate the lookup loading code into the generated Go program's preamble (where data sources are loaded and registered) and the `main()` function (where the loading is invoked before any metric calculations). This ensures lookup data is available before computations run.

a. Modify `src/main/go/synthesis/preamble_generator.go` to call `generateLookupLoading`:
   - After connector loading code is emitted in the preamble, emit the `loadLookupData` function
   - Ensure the emitted function signature is `func loadLookupData(controller *ChronoconeComputationController) error`
   — Not Started

b. Modify `src/main/go/synthesis/main_generator.go` to include `loadLookupData` call:
   - In the generated `main()` function, add a call to `loadLookupData(controller)` after connector loading and before scheduling metric computation order
   - Handle the returned error with appropriate logging/error propagation
   — Not Started

c. Write an integration test in `src/main/go/synthesis/lookup_pipeline_test.go`:
   - Full pipeline: parse a CPL file with lookup declarations → analyze → synthesize
   - Verify the generated Go code compiles (`go build` on the generated output)
   - Verify the generated `main()` calls `loadLookupData` before any `ComputeMetrics` calls
   — Not Started

d. Run `go test ./src/main/go/synthesis/... -count=1 -timeout 120s` and verify integration test passes — Not Started

**Acceptance Criteria**:
- Generated preamble includes a `loadLookupData` function after connector loading functions
- Generated `main()` calls `loadLookupData` before `ScheduleMetricComputationOrder` or any `ComputeMetrics` calls
- Full pipeline integration test passes (parse → analyze → synthesize → compile generated Go)
- No regressions in existing synthesis tests

### Task 5: Create with_data/lookups_basic example — Completed

**Intent**: Provide an end-to-end example demonstrating a timeless lookup (product-to-category mapping). This validates that the full pipeline — CPL parsing, analysis, synthesis, and generated code execution — works correctly for basic lookup declarations.

a. Create the example directory at `.../with_data/examples/lookups_basic/` with the following files — Not Started

b. Create a CPL file `lookups_basic.cpl` with a timeless lookup declaration:
   - Lookup name: `product_category`
   - Dimensions: `[Product]` (referencing a product concept)
   - Celltype: `CategoryLabel`
   — Not Started

c. Create a Lookup IO Spec JSON file `lookups_basic.io.json` mapping the lookup name `product_category` to a parquet file path `product_category.parquet` — Not Started

d. Create a parquet data file `product_category.parquet` with test values:
   - Columns: `Product` (string), `CategoryLabel` (string)
   - Include at least 3 rows of product-to-category mappings
   — Not Started

e. Create `.expected.json` with the expected output after running the generated program:
   - The output should include the loaded lookup data (or a verification that the lookup loads without error in Increment 1)
   — Not Started

f. Verify end-to-end: run the full compiler pipeline on the example → verify the generated Go program output matches `.expected.json` — Not Started

**Acceptance Criteria**:
- `lookups_basic.cpl` parses and analyzes without errors
- Synthesis generates compilable Go code from the example
- The generated Go program executes and produces output matching `.expected.json`
- The example demonstrates that lookup data is successfully loaded and registered as a `MetricComputation`

### Task 6: Verify parse_and_analyze_only synthesis-level examples and regression suite — Completed

**Intent**: Ensure that synthesis does not break existing functionality. Verify that parse-and-analyze-only CPL files from the analysis phase (Stories 241/242) synthesize correctly, and run the full existing CPL test suite to confirm zero regressions.

a. Verify that `parse_and_analyze_only` CPL files from the analysis phase stories (241, 242) synthesize without errors:
   - Run synthesis on each parse-and-analyze-only CPL file
   - Confirm generated Go code compiles
   — Not Started

b. Run the full existing CPL test suite:
   - `go test ./src/main/go/... -count=1 -timeout 300s`
   - Confirm zero test regressions — all previously passing tests still pass
   — Not Started

**Acceptance Criteria**:
- All parse-and-analyze-only CPL files synthesize and produce compilable Go code
- Full CPL test suite passes with zero regressions
- Any test failures are investigated and resolved or documented as pre-existing

## Test-First Development

This project follows the Logical TDD Lifecycle. For each coding task above, tests are written before implementation and the test-and-implement cycle is integrated within each task rather than broken out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

When implementing each task:
1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. Tests must exercise the full logical contract: valid inputs produce correct output; invalid inputs produce descriptive errors
3. All tests must pass before marking a task complete

## Constraints

- Reuses existing `readParquetFile` from connector synthesis — do not create a separate parquet reading mechanism
- Tensor cell type defaults to `float64`; type selection is deferred to a later increment per the ERO
- Lookup loading must occur BEFORE any metric calculation in the generated `main()` function
- Lookup IO Specification is a separate file from the Connector IO Specification (distinct JSON schema)
- No grain handling in Increment 1 — lookups use only dimension concept names for tensor indexing
- The `registerLookups` helper treats `CILLookupNode` as a leaf — no recursive CIL subtree walking

## Acceptance Criteria

1. Lookup IO Specification parsing produces a mapping of lookup names to data source paths
2. `registerLookups` registers lookup metadata in the walker's CILRegistry
3. `generateLookupLoading` emits compilable Go code that reads parquet and registers `MetricComputation`
4. Generated `main()` calls `loadLookupData` before any metric calculations
5. Tensor indexing is by concept dimension names (no grain handling yet)
6. `with_data/lookups_basic` example passes end-to-end
7. Existing test suite passes with zero regressions
