# Story234: Fix Conditional Synthesis Defects

## Goal

Fix defects in the conditional concept synthesis pipeline identified during the Story227 code review. The analysis already stores `ParentConceptName` on `AnalysisConceptDefinition`, but synthesis discards it and guesses the parent dimension by iterating all other dimensions. This produces incorrect validity masks when dimension values overlap. In addition, the `conditional_grain` example violates project conventions, no end-to-end test verifies conditional synthesis compiles and runs, and several emission paths lack test coverage. Backward compatibility for the conditional mappings format is not a concern — simplify wherever possible.

## References

- [Fix Plan: Conditional Synthesis Defects](../plans/fix-cond-synth-bugs-plan.md)
- [Story227: CPL Conditional Concept Values Synthesize](finished-stories/Story227_cpl-conditional-concept-values-synthesize.md) — _Done_

## Dependencies

- [Story227_cpl-conditional-concept-values-synthesize.md](finished-stories/Story227_cpl-conditional-concept-values-synthesize.md) — _Done_. This story fixes defects discovered during Story227's code review.

## Dependent Stories

None.

## Tasks

### Task Sizing

- **Tasks 1, 2, 3, 6** are medium-sized: each modifies production code across multiple call sites or packages, introduces new structs, or requires reasoning about cross-file coordination. Routed to `code-for-story-implementor`.
- **Task 4** is `[Small]`: single test file, purely additive, follows the existing `TestEndToEndSynthesizeCompileRun` pattern exactly.
- **Task 5** is `[Small]`: two test files across two packages, but each sub-task adds narrowly scoped test functions following established patterns with no production code changes.

### Parallel Execution

- **Group A** (Tasks 1, 3, 6): All three tasks are independent with completely disjoint file sets. Task 1 modifies `synthesis/`; Task 3 modifies `io/` and `cli/parquetgen/`; Task 6 modifies `examples/with_data/conditional_grain/`. No compile-time or semantic dependencies between them.

### Execution Order

```
[Task 1 ‖ Task 3 ‖ Task 6] → Task 2 → Task 4 → Task 5
```

### Reintegration

**Group A (Tasks 1, 3, 6):**
- Merge order: any order (fully disjoint packages, no shared infrastructure)
- Integration tests:
  - `cd src/chronocone_planning_language && go test ./src/main/go/synthesis/... ./src/main/go/io/... ./src/main/go/cli/parquetgen/... -count=1 -timeout 120s` — verify all modified packages compile and existing tests pass
  - `cd src/chronocone_planning_language && shell_scripts/compile_and_run.sh monthly_data examples/with_data/conditional_grain` — verify restructured example compiles and runs
- Watch for conflicts in: `src/chronocone_planning_language/src/main/go/go.sum` (dependency resolution may differ per worktree), `src/chronocone_planning_language/src/main/go/go.mod` (low risk but verify no unexpected module-level changes)

### Task 1: Create ConditionalConceptEntry type and update BuildConditionalConceptValuesMap — Completed

**Goal**: Thread the parent concept name from analysis through synthesis so mask computation knows the correct parent dimension.

**Test-First Development**:

a. Add the exported struct to `src/main/go/synthesis/synthesize.go`:
   ```go
   type ConditionalConceptEntry struct {
       ParentConcept string
       Values        map[string][]string
   }
   ```

b. Update `BuildConditionalConceptValuesMap` to return `map[string]ConditionalConceptEntry` instead of `map[string]map[string][]string`. For each conditional concept, read `concept.ParentConceptName`; if nil, skip the concept and emit a log warning.

c. Update all call sites in `GenerateMainFunction` and `BuildConnectorValidityMasks` to use `.ParentConcept` and `.Values` accessors.

d. Update existing tests in `src/main/go/synthesis/synthesize_test.go` to construct `ConditionalConceptEntry` values instead of nested maps.

e. Run `cd src/chronocone_planning_language && go test ./src/main/go/synthesis/...` — all tests must pass.

**Acceptance Criteria**:
- `ConditionalConceptEntry` struct exists with `ParentConcept` and `Values` fields
- `BuildConditionalConceptValuesMap` returns the new type
- Call sites in synthesis emission code use `.ParentConcept` and `.Values`
- Tests compile and pass

### Task 2: Fix ComputeValidityMask to use explicit parent dimension — Completed

**Goal**: Eliminate the parent-ambiguity bug where `ComputeValidityMask` guesses the parent by iterating all other dimensions.

**Test-First Development**:

a. Update `ComputeValidityMask` signature to accept `map[string]ConditionalConceptEntry`.

b. Replace the "try all other dimensions" loop:
   ```go
   for k, otherDim := range dimNames {
       if k == j { continue }
       parentVal := conceptValues[otherDim][indices[k]]
       if childSet, ok := childCondSet[parentVal]; ok { ... }
   }
   ```
   With a direct parent lookup:
   ```go
   parentName := entry.ParentConcept
   for k, otherDim := range dimNames {
       if otherDim == parentName {
           parentVal := conceptValues[otherDim][indices[k]]
           if childSet, ok := entry.Values[parentVal]; ok { ... }
           break
       }
   }
   ```

c. Update `BuildConnectorValidityMasks` signature and `GenerateMainFunction` emission code accordingly.

d. Add a targeted test with overlapping value names: `Month = ["Jan","Feb"]` and `Season = ["Jan","Feb"]` with `DayInMonth` conditional on `Month`. Verify that cells where `Season = Feb` but `Month = Jan` are correctly resolved to the Jan mask.

e. Run `cd src/chronocone_planning_language && go test ./src/main/go/synthesis/...` — all tests must pass.

**Acceptance Criteria**:
- `ComputeValidityMask` uses the explicit `ParentConcept` field instead of iterating all dimensions
- A test with overlapping dimension values verifies the correct parent is identified
- All existing and new tests pass

### Task 3: Fix GenerateConditionalProduct with explicit parent mapping — Completed

**Goal**: Replace `inferConditionalParentOrder` in parquet generation with an explicit parent mapping, eliminating the same parent-ambiguity bug in the data generation path.

**Test-First Development**:

a. Add `ConditionalDimMapping` struct to `src/main/go/io/generate_parquet.go`:
   ```go
   type ConditionalDimMapping struct {
       Parent string
       Values map[string][]string
   }
   ```

b. Change `GenerateParquetFileWithConditionals` signature to accept `conditionalDims map[string]ConditionalDimMapping` instead of `map[string]map[string][]string`.

c. Update `GenerateConditionalProduct` to use `mapping.Parent` for the parent dimension instead of `inferConditionalParentOrder`.

d. Remove `inferConditionalParentOrder` entirely.

e. Update `src/main/go/cli/parquetgen/parquet_gen.go` to parse the new JSON format:
   ```json
   {
     "DayInMonth": {
       "parent": "Month",
       "values": {
         "Jan": ["1","2","3",...],
         "Feb": ["1","2",...]
       }
     }
   }
   ```
   Update `ParquetGenSpec.ConditionalMappings` and `parseSpec` accordingly.

f. Update all tests in `src/main/go/io/generate_parquet_test.go` and `src/main/go/cli/parquetgen/parquet_gen_test.go` to use the new `ConditionalDimMapping` type and JSON format.

g. Run `cd src/chronocone_planning_language && go test ./src/main/go/io/... ./src/main/go/cli/parquetgen/...` — all tests must pass.

**Acceptance Criteria**:
- `ConditionalDimMapping` struct exists with `Parent` and `Values` fields
- `GenerateParquetFileWithConditionals` uses the new type
- `inferConditionalParentOrder` is removed
- CLI `--conditional_mappings` JSON uses the nested `parent`/`values` format
- All io and parquetgen tests pass

### Task 4: Add conditional end-to-end synthesis test — Completed

**Goal**: Add an automated test that verifies synthesized code with conditional grain concepts compiles and runs correctly end-to-end.

**Test-First Development**:

a. In `src/main/go/synthesis/synthesize_test.go`, add `TestEndToEndSynthesizeCompileRun_ConditionalGrain` modeled on the existing `TestEndToEndSynthesizeCompileRun`.

b. Construct `AnalysisDefinitions` with:
   - Concepts: `Month` (values: Jan/Feb/Mar), `DayInMonth` (conditional on Month; Jan=31, Feb=28, Mar=31)
   - Measures: `Monthly` (Interval, unit=Month), `data_count` (Number, unit=Month)
   - Connector: `monthly_data` with dimensions `[Month, DayInMonth]`, grain `Monthly`, celltype `data_count`
   - Metric: references `data_count` with grain dimension `Monthly` using a CIL node that reads connector data (not a literal zero)

c. Generate a conditional parquet file with `GenerateParquetFileWithConditionals` (90 rows: 31+28+31), setting fixed cell values (sequential 1.0, 2.0, ... 90.0).

d. Build an `IOSpecification` pointing at the generated parquet file, with output specifying the metric.

e. Call `Synthesize(analysisDefs, ioSpec, outputPath)`.

f. Run the synthesized program with `go run outputPath`.

g. Verify:
   - Output parquet is created
   - Column `Value` contains the expected metric results
   - The synthesized program runs without error

h. Run `cd src/chronocone_planning_language && go test ./src/main/go/synthesis/... -run TestEndToEndSynthesizeCompileRun_ConditionalGrain` — must pass.

**Acceptance Criteria**:
- `TestEndToEndSynthesizeCompileRun_ConditionalGrain` exists and passes
- The test generates conditional data, synthesizes Go code, compiles it, runs it, and verifies output parquet values

### Task 5: Add missing test coverage for conditional synthesis — Completed

**Goal**: Close test coverage gaps for `connectorValidityMasks` emission, conditional loader signatures, and the `hasConditionalMasks=true` connector loader path.

**Test-First Development**:

a. **5a — connectorValidityMasks emission** (`src/main/go/synthesis/synthesize_test.go`): Add a test that calls `GenerateMainFunction` with conditional concepts and verifies the emitted Go code contains:
   - `var connectorValidityMasks = map[string][]uint64{`
   - `{0x...}` hex-encoded mask values
   - `tensor.SetMask(mask)` inside the loader function

b. **5b — Conditional loader signature** (`src/main/go/synthesis/synthesize_test.go`): Add a test for `generateInputLoadingCode(ioSpec, "connectorMetadata", analysisDefs, true)` that verifies the emitted code contains:
   - `conditionalConceptValues` as the third argument to loader calls
   - `loadMonthlyData(monthly_dataInputSpec, conceptValues, conditionalConceptValues)`

c. **5c — hasConditionalMasks=true in connector loader** (`src/main/go/connector/connector_loader_test.go`): Add tests for `GenerateConnectorLoaders(mock, ioSpec, true)` and `GenerateSingleConnectorLoader(name, metadata, inputSpec, true)` that verify:
   - The generated function accepts `conditionalConceptValues map[string]map[string][]string` as a third parameter
   - The generated body contains `connectorValidityMasks[...]` and `tensor.SetMask(mask)`
   - When `hasConditionalMasks=false`, these are absent (existing tests already confirm)

d. Run `cd src/chronocone_planning_language && go test ./src/main/go/synthesis/... ./src/main/go/connector/...` — all tests must pass.

**Acceptance Criteria**:
- Tests exist for connector validity mask emission in generated main function
- Tests exist for conditional loader signature in input loading code
- Tests exist for `hasConditionalMasks=true` in connector loader generation
- All tests pass

### Task 6: Restructure conditional_grain example — Completed

**Goal**: Align the `conditional_grain` example with project conventions: example-specific definitions never appear in `core.cpl`, and each example subdirectory contains a `.cpl` file named after the directory, `iospec.json`, and data files.

**Test-First Development**:

a. Create the directory structure:
   ```
   examples/with_data/conditional_grain/
   ├── core.cpl                          # shared infrastructure ONLY
   └── monthly_data/
       ├── monthly_data.cpl              # example-specific declarations
       ├── iospec.json                   # IO specification
       ├── monthly_data.parquet          # generated test data
       └── generate_data.go              # script to generate the parquet
   ```

b. Update `core.cpl` to keep **only** shared infrastructure: concept declarations (`Month`, conditional `DayInMonth`), unit declarations (`Month = Month`), and type declarations (`Count = Number(precision=0)`). Remove measures, connectors, and metrics.

c. Create `monthly_data/monthly_data.cpl` with example-specific declarations: measures (`data_count`, `Monthly`), connector (`monthly_data`), and metrics.

d. Create `monthly_data/iospec.json` mapping the input connector to the parquet file and specifying output metrics. Must include `connector_name`, `data_format`, `driver`, and `driver_parameters` as required by `compile_and_run.sh`.

e. Move `generate_data.go` to `monthly_data/` and update file paths to be relative to the example subdirectory.

f. Remove `run_test.sh` and `verify_test.go`.

g. Regenerate `monthly_data.parquet` if needed.

h. Verify with `cd src/chronocone_planning_language && shell_scripts/compile_and_run.sh monthly_data examples/with_data/conditional_grain`:
   - The compiler runs without errors
   - The synthesized program compiles and executes
   - Output parquet is generated

**Acceptance Criteria**:
- Example directory structure matches project conventions
- `core.cpl` contains only shared infrastructure
- `monthly_data.cpl` contains only example-specific declarations
- `iospec.json` is present and valid
- `compile_and_run.sh` succeeds for the example

## Constraints

1. Backward compatibility for the conditional mappings JSON format is **not** required — simplify the format to use explicit parent fields.
2. The emitted Go code format (`conditionalConceptValues` map, `connectorValidityMasks` map) must **not** change — only the internal computation pipeline changes.
3. All file paths in tasks are relative to `src/chronocone_planning_language/` (the `source_root`). Go commands must use `./src/main/go/...` package paths.
4. `go test` commands must include `-count=1` to disable test caching.

## Intent

Eliminate the parent-dimension ambiguity bug in both synthesis and parquet generation by threading the explicit parent concept name through the pipeline. Ensure the conditional synthesis path has end-to-end test coverage and that the example directory follows project conventions so future examples have a correct reference.

## Acceptance Criteria

1. `ComputeValidityMask` resolves parent dimensions unambiguously using `ConditionalConceptEntry.ParentConcept`
2. A test with overlapping dimension value names confirms correct mask behavior
3. `GenerateParquetFileWithConditionals` uses explicit `ConditionalDimMapping.Parent`
4. An end-to-end synthesis test compiles and runs conditional grain code, verifying output parquet values
5. Tests cover connector validity mask emission, conditional loader signatures, and `hasConditionalMasks=true` loader generation
6. The `conditional_grain` example follows project conventions and passes `compile_and_run.sh`
7. All existing tests continue to pass

## Notes

- The analysis phase already sets `ParentConceptName` on `AnalysisConceptDefinition` (via `ResolveConditionalValues` in `concept_analysis.go`). The fix is purely in synthesis and parquet generation to read and use this field.
- `inferConditionalParentOrder` exists in both `generate_parquet.go` and may have a related function in the CLI spec parser. Remove it entirely — no fallback heuristic is needed once the parent is explicit.
- Task 4 requires `go` in PATH (same requirement as the existing `TestEndToEndSynthesizeCompileRun`).
