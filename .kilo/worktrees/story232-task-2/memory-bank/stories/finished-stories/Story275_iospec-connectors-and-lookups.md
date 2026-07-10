# Story275: Rename iospec `inputs` → `connectors`, Add Lookups Section, Fix lookups_basic Example

## Goal

Rename the iospec.json `"inputs"` key to `"connectors"` and the corresponding Go struct `InputSpecification` to `ConnectorSpecification` throughout the codebase. Add a new `"lookups"` section to the iospec (replacing the separate `_lookups.io.json` file) using the same schema as connectors. Fix the lookups_basic example data, CPL, and iospec to produce a coherent demonstration of lookups.

## References

- [Plan: Adjust Lookups](plans/adjust-lookups-plan.md)
- [Brief](../brief.md)

## Dependencies

None — this story's changes are self-contained and do not depend on other stories.

## Dependent Stories

None — other lookups stories are already using lookups, and aggregator/spreader stories do not depend on the iospec field names.

## Tasks

### Task Sizing

Seven tasks are classified as `[Small]` based on architect-for-story-planning review: Tasks 2, 4, 5, 6, 8, 12, 14 (single concerns, at most 2 files, mechanical or following existing patterns). Tasks 3 and 15 were considered borderline but defaulted to **not small** — Task 3 requires reasoning across multiple control-flow paths in a 1015-line function; Task 15 introduces a new data-access path beyond a simple switch-case addition. The architect confirmed no task is too large for single-handoff execution — all 16 tasks are appropriately sized.

### Parallel Execution

- **Group A** (Tasks 1, 2): Task 1 modifies `src/main/go/io/` (adds LookupSpecification struct, renames types); Task 2 deletes files in `src/main/go/synthesis/`. Disjoint files, different packages.
- **Group B** (Tasks 3, 6, 7, 9, 10, 11): All tasks modify disjoint files across 5 different packages/directories. Each depends only on the renamed types from Group A, not on each other.
- **Group C** (Tasks 4, 5, 8, 12, 13, 15): All tasks modify disjoint files. Note: Task 15 depends on Task 7 (Group B) completing first; Task 14 depends on Task 13 (Group C) completing first.

### Execution Order

```
[Task 1 ‖ Task 2] → [Task 3 ‖ Task 6 ‖ Task 7 ‖ Task 9 ‖ Task 10 ‖ Task 11] → [Task 4 ‖ Task 5 ‖ Task 8 ‖ Task 12 ‖ Task 13 ‖ Task 15] → Task 14 → Task 16
```

### Reintegration

**Group A (Tasks 1, 2):**
- Merge order: any order (fully disjoint packages: `io/` vs `synthesis/`)
- Integration tests: `go test ./src/main/go/io/... -count=1 -timeout 60s` (verify renamed types + LookupSpecification); `go build ./src/main/go/synthesis/... -count=1 -timeout 60s` (verify package compiles after deletion)
- Watch for conflicts in: none expected (fully disjoint file sets)

**Group B (Tasks 3, 6, 7, 9, 10, 11):**
- Merge order: any order (disjoint files across `synthesis/`, `connector/`, `expected_xls_generator/`, `compiler/` (test-only), `examples/`)
- Integration tests: `go test ./src/main/go/connector/... -count=1 -timeout 60s`; `go test ./src/main/go/cli/expected_xls_generator/... -count=1 -timeout 60s`; `go vet ./src/main/go/synthesis/... && go build ./src/main/go/synthesis/...` (verify Task 3 compiles — note synthesis/ tests will NOT pass until Group C merges since lookup_synthesis.go is not yet updated)
- Watch for conflicts in: none expected (disjoint file sets)

**Group C (Tasks 4, 5, 8, 12, 13, 15):**
- Merge order: any order (disjoint files across different packages and example directories)
- Integration tests: `go test ./src/main/go/synthesis/... -count=1 -timeout 60s` (synthesis package fully testable for first time after Task 3 signature change); `go test ./src/main/go/cli/compiler/... -count=1 -timeout 120s` (compiler tests with all test fixture updates); `go test ./src/main/go/cli/expected_xls_generator/... -count=1 -timeout 60s` (CILLookupNode evaluator)
- Watch for conflicts in: none expected (disjoint file sets)

---

### Task 1: Rename Inputs → Connectors and add Lookups to IOSpecification struct — Completed

a. In `src/main/go/io/io_specification.go`: rename `Inputs []InputSpecification` → `Connectors []ConnectorSpecification`, rename the type `InputSpecification` → `ConnectorSpecification`, change JSON tag from `"inputs"` to `"connectors"` — Not Started
b. In `src/main/go/io/io_specification.go`: add new `Lookups []LookupSpecification` field with JSON tag `"lookups"` — Not Started
c. Define `LookupSpecification` struct in the same file: fields `LookupName` (json:`"lookup"`), `DataFormat`, `Driver`, `DriverParameters` — mirroring ConnectorSpecification's schema — Not Started
d. Update `LoadIOSpecification` docs to mention lookups in the struct comment — Not Started
e. In `src/main/go/io/io_specification_test.go`: update all JSON fixture strings from `"inputs"` → `"connectors"`, rename `InputSpecification` → `ConnectorSpecification` in Go struct literals, rename `.Inputs` → `.Connectors` in struct fields — Not Started
f. In `src/main/go/io/io_specification_test.go`: add tests for Lookups field parsing (at least two test cases: one with lookups populated, one with empty lookups) — Not Started
g. Run `go test ./src/main/go/io/...` to verify all io package tests pass — Not Started

### Task 2: Delete lookup_io_spec.go and its test — Completed

a. Delete `src/main/go/synthesis/lookup_io_spec.go` — Not Started
b. Delete `src/main/go/synthesis/lookup_io_spec_test.go` — Not Started
c. Run `go build ./src/main/go/synthesis/...` to verify no compilation errors from removed files — Not Started

### Task 3: Update SynthesisInput struct and Synthesize signature — Completed

a. In `src/main/go/synthesis/input.go`: remove `LookupIOSpec *LookupIOSpec` field from `SynthesisInput` struct — Not Started
b. In `src/main/go/synthesis/synthesize.go`: remove `lookupIOSpec *LookupIOSpec` parameter from `Synthesize()` function signature — Not Started
c. In `src/main/go/synthesis/synthesize.go`: remove `input.LookupIOSpec = lookupIOSpec` assignment from `BuildTransformerFromAnalysis` — Not Started
d. In `src/main/go/synthesis/synthesize.go`: rename `needsInputs` variable to `needsConnectors`, update `len(ioSpec.Inputs)` → `len(ioSpec.Connectors)` — Not Started
e. In `src/main/go/synthesis/synthesize.go`: also consider `len(ioSpec.Lookups) > 0` when computing `needsConnectors` (so `flag`, `path/filepath` imports are included when lookups are present) — Not Started
f. In `src/main/go/synthesis/synthesize.go`: rename `generateInputLoadingCode` function → `generateConnectorLoadingCode`, update docs and all `ioSpec.Inputs` → `ioSpec.Connectors` references in the function body — Not Started
g. Run `go vet ./src/main/go/synthesis/... && go build ./src/main/go/synthesis/...` to verify compilation (note: tests may not fully pass until Tasks 4 and 8 complete) — Not Started

### Task 4: Update lookup synthesis to use ioSpec.Lookups — Completed [Small]

a. In `src/main/go/synthesis/lookup_synthesis.go`: update `generateLookupIntegration` to check `input.IOSpec.Lookups` instead of `input.LookupIOSpec` — Not Started
b. In `src/main/go/synthesis/lookup_synthesis.go`: update `generateLookupLoading` to accept `[]cio.LookupSpecification` instead of `*LookupIOSpec` — Not Started
c. In `src/main/go/synthesis/lookup_synthesis.go`: iterate `ioSpec.Lookups` and extract path from `DriverParameters[0]["path"]` (same pattern used by connectors) — Not Started
d. Run `go test ./src/main/go/synthesis/...` to verify synthesis package tests pass — Not Started

### Task 5: Update compiler to remove _lookups.io.json discovery — Completed [Small]

a. In `src/main/go/cli/compiler/cpl_compiler.go`: remove the `_lookups.io.json` discovery code in `synthesizePhase` (the block from lines 218-236 that scans for `*_lookups.io.json` files, loads them via `ParseLookupIOSpec`, and stores the result) — Done (already removed by Task 3)
b. In `src/main/go/cli/compiler/cpl_compiler.go`: update the `Synthesize()` call to pass `nil` for the removed `lookupIOSpec` parameter (lookups now come from ioSpec directly) — Done (already updated by Task 3)
c. Run `go build ./src/main/go/cli/compiler/... && go vet ./src/main/go/cli/compiler/...` to verify compilation — Done (builds cleanly, no unused imports)

### Task 6: Update connector_loader.go (Inputs → Connectors rename) — Completed [Small]

a. In `src/main/go/connector/connector_loader.go`: update `ioSpec.Inputs` → `ioSpec.Connectors` in `GenerateConnectorLoaders` — Not Started
b. In `src/main/go/connector/connector_loader.go`: rename `InputSpecification` → `ConnectorSpecification` in function signatures, `LoaderFuncDecl.Render()` generated code body, and `GenerateLoadParquetConnectorFunction()` body — Not Started
c. In `src/main/go/connector/connector_loader_test.go`: update all `io.InputSpecification` → `io.ConnectorSpecification`, `.Inputs` → `.Connectors` field references, and `cplio.InputSpecification` string references → `cplio.ConnectorSpecification` — Not Started
d. Run `go test ./src/main/go/connector/...` to verify connector tests pass — Not Started

### Task 7: Update expected_xls_generator (Inputs → Connectors rename) — Completed

a. In `src/main/go/cli/expected_xls_generator/generator.go`: `iospec.Inputs` → `iospec.Connectors`, `InputSpecification` → `ConnectorSpecification` — Not Started
b. In `src/main/go/cli/expected_xls_generator/main/main.go`: `iospec.Inputs` → `iospec.Connectors`, `InputSpecification` → `ConnectorSpecification` — Not Started
c. In `src/main/go/cli/expected_xls_generator/generator_test.go`: update all `Inputs: []io.InputSpecification{}` → `Connectors: []io.ConnectorSpecification{}` and type references — Not Started
d. Run `go test ./src/main/go/cli/expected_xls_generator/...` to verify tests pass — Not Started

### Task 8: Update synthesize_test.go and lookup_synthesis_test.go — Completed [Small]

a. In `src/main/go/synthesis/synthesize_test.go`: rename all `ioSpec.Inputs` → `ioSpec.Connectors`, `cio.InputSpecification` → `cio.ConnectorSpecification`, and `cplio.InputSpecification` string references → `cplio.ConnectorSpecification` — Not Started
b. In `src/main/go/synthesis/synthesize_test.go`: remove any `lookupIOSpec` parameter from `Synthesize()` calls and update `TestEndToEndSynthesizeLookups_Basic` to use ioSpec with embedded `Lookups` — Not Started
c. In `src/main/go/synthesis/lookup_synthesis_test.go`: update tests to pass `[]ConnectorSpecification` and `[]LookupSpecification` in ioSpec, remove any `LookupIOSpec` construction — Not Started
d. Run `go test ./src/main/go/synthesis/...` to verify synthesis tests pass — Not Started

### Task 9: Update compiler test JSON fixtures — Completed

a. In `src/main/go/cli/compiler/cpl_compiler_test.go`: update all inline JSON fixture strings `"inputs"` → `"connectors"` — Not Started
b. In `src/main/go/cli/compiler/cpl_integration_test.go`: update all inline JSON fixture strings `"inputs"` → `"connectors"` — Not Started
c. In `src/main/go/cli/compiler/cpl_example_test.go`: update all inline JSON fixture strings `"inputs"` → `"connectors"` — Not Started
d. Run `go test ./src/main/go/cli/compiler/...` to verify compiler tests pass — Not Started

### Task 10: Update all 18 with_data example iospec.json files — Completed

a. Rename `"inputs"` → `"connectors"` in every `examples/with_data/**/iospec.json` file (18 files total). This is a mechanical JSON key rename — use grep/sed across all matching files. — Not Started
b. Run `go test ./src/main/go/cli/compiler/...` to verify example tests pass with the updated JSON keys — Not Started

### Task 11: Update all 6 parse_and_analyze_only JSON files — Completed

a. Rename `"inputs"` → `"connectors"` in every `examples/parse_and_analyze_only/without_errors/*.json` file (6 files total). — Not Started
b. Run `go test ./src/main/go/cli/compiler/...` to verify parse tests pass — Not Started

### Task 12: Fix lookups_basic CPL comment and iospec — Completed [Small]

a. In `examples/with_data/lookups/lookups_basic/lookups_basic.cpl`: update stale comment on line 2 from "Product to CategoryLabel" to reflect "Material-based strength adjustment" — Not Started
b. In `examples/with_data/lookups/lookups_basic/iospec.json`: rename `"inputs"` → `"connectors"` — Not Started
c. In `examples/with_data/lookups/lookups_basic/iospec.json`: fix connector name `"dummy"` → `"raw_strength"`, fix output metric `"result"` → `"product_strength"`, fix dimensions `["Category"]` → `["Product", "Material"]`, fix celltype `"CategoryLabel"` → `"Strength"`, keep grain `["Daily"]` — Not Started
d. In `examples/with_data/lookups/lookups_basic/iospec.json`: add `"lookups"` section with one entry for `strength_adjusted_for_material`, using `data_format: "parquet"`, `driver: "file"`, `driver_parameters: [{"path": "strength_adjusted_for_material.parquet"}]` — Not Started

### Task 13: Create corrected lookups_basic data files — Completed

a. Create `examples/with_data/lookups/lookups_basic/raw_strength.json`: 63 rows (3 Products × 3 Materials × 7 Days), columns: Product, Material, Daily, Value, all values = 1.0 — Not Started
b. Create `examples/with_data/lookups/lookups_basic/strength_adjusted_for_material.json`: 3 rows, columns: Material, Value — values: Steel=1.0, Aluminum=0.8, Copper=0.9 — Not Started
c. Regenerate `raw_strength.parquet` from `raw_strength.json` and `strength_adjusted_for_material.parquet` from `strength_adjusted_for_material.json` using the parquet generation Go tool in `src/main/go/io/generate_parquet.go` or the `parquetgen` script — Not Started
d. Delete old data files: `data_dummy.json`, `data_dummy.parquet`, `lookups_basic_product_category.json`, `lookups_basic_product_category.parquet` — Not Started
e. Delete `examples/with_data/lookups/lookups_basic/lookups_basic_lookups.io.json` — no longer needed since lookups are in iospec.json — Not Started

### Task 14: Regenerate expected_output.xlsx for lookups_basic — Completed [Small]

a. Run the expected_xls_generator tool on the updated lookups_basic example: `go run ./src/main/go/cli/expected_xls_generator/main -example examples/with_data/lookups/lookups_basic` — Not Started
b. Verify the generated `expected_output.xlsx` contains correct expected values (1.0 for Steel products, 0.8 for Aluminum products, 0.9 for Copper products across all days) — Not Started
c. Run `go test ./src/main/go/cli/compiler/...` to verify the full lookups_basic example test passes — Not Started

### Task 15: Extend expected_xls_generator evaluator for CILLookupNode — Completed

a. Located the expression evaluator code in `src/main/go/cli/expected_xls_generator/cil_evaluator.go` — Done
b. Added a case for `*cil.CILLookupNode` in the expression evaluator's dispatch switch. The case resolves the lookup value by matching the current row's dimension values against the lookup data rows loaded from `<lookup_name>.json`. All dimensions declared on the lookup must match for a row to be selected, and the `"Value"` field is returned — Done
c. In `src/main/go/cli/expected_xls_generator/main/main.go`: after loading connector input data, also loads lookup data files. For each lookup in `iospec.Lookups`, loads `<lookup_name>.json` from the example directory and stores the data in a map keyed by lookup name — Done
d. Passed the lookup data map through to the evaluator via `GenerateXLSX`, `buildValueLookup`, and `EvaluateExpressionWithContext` using the new `EvalContext` struct — Done
e. Added test cases for `CILLookupNode` evaluation (direct lookup, missing data, binary op with lookup, error cases). Ran `go test ./src/main/go/cli/expected_xls_generator/...` and all tests pass — Done

### Task 16: Run full test suite validation — Completed [Small]

a. Run `go test ./src/main/go/... -count=1 -timeout 300s` to confirm all tests pass across all packages — Not Started
b. If any test failure is found, identify the failing package, fix the issue, and re-run — Not Started

## Test-First Development

This project follows the Logical TDD Lifecycle. For each coding task, write or update tests before implementing the code changes. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks.

Tasks 1, 3, 4, 6, 7, 15 involve Go code changes and must start with test updates (subtask (e) or equivalent appearing before the implementation subtasks verify tests fail/pass appropriately). For mechanical renames (Tasks 8, 9, 10, 11), update test fixtures simultaneously with the production code since the rename affects both at once.

## Constraints

- The `"connectors"` and `"lookups"` JSON keys must use the same structure (`connector`/`lookup` name, `data_format`, `driver`, `driver_parameters`) to maintain consistency
- The `LookupSpecification` Go struct must mirror `ConnectorSpecification` exactly (same field types, same JSON tag pattern)
- Lookup data files are named `<lookup_name>.json`/`.parquet` — matching the lookup name in the iospec, the same convention as connectors
- All JSON fixture strings in Go files (test data inline strings) must use correct JSON field names
- The CPL file format is unchanged; only the iospec.json format changes

## Intent

Task 1 defines the new type system — this is the foundation all other tasks build on. Tasks 2-9 are mechanical renames and structural updates to eliminate the old `Inputs`/`InputSpecification` naming and the separate `LookupIOSpec`. Tasks 10-11 are JSON file mechanical renames. Tasks 12-14 fix the specific lookups_basic example to demonstrate correctly functioning lookups. Task 15 enables the test infrastructure to validate lookup examples. Task 16 is the final verification gate.

## Acceptance Criteria

- All occurrences of `Inputs` and `InputSpecification` in Go source code are renamed to `Connectors` and `ConnectorSpecification`
- All JSON fixture strings in tests and example files use `"connectors"` instead of `"inputs"`
- `io_specification.go` contains a `Lookups []LookupSpecification` field with a matching struct definition
- `lookup_io_spec.go` and `lookup_io_spec_test.go` are deleted
- `SynthesisInput.LookupIOSpec` is removed; lookup synthesis uses `ioSpec.Lookups`
- `Synthesize()` no longer takes a separate `lookupIOSpec` parameter
- `cpl_compiler.go` no longer discovers `_lookups.io.json` files
- `lookups_basic` example iospec uses `"connectors"` key, connector name `"raw_strength"`, metric `"product_strength"`, correct dimensions and celltype
- `lookups_basic` data files map 63 connector rows and 3 lookup rows with correct values
- `expected_xls_generator` evaluator handles `CILLookupNode` in CIL expressions
- `go test ./src/main/go/... -count=1` passes with zero failures

## Requesting Clarification

If at any point during story implementation there is confusion or ambiguity about the goal or how to accomplish it, stop and consult the plan at `memory-bank/plans/adjust-lookups-plan.md`. If the plan is insufficient, escalate to the user.
