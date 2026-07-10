# Story277: Reorganize with_data and Add Unit-Checking Examples

## Goal

Reorganize `examples/with_data/` into `without_errors/` and `with_errors/` subdirectories (mirroring the `parse_and_analyze_only/` structure), move all 19 existing leaf examples under `without_errors/`, update the test discovery infrastructure to match the new layout, add a `TestWithDataExamples_WithErrors` test for error-case examples, and create both end-to-end conversion examples and error-case examples that exercise Story231's synthesis-time unit type checker through the full compilation pipeline.

## References

- [Architecture](../specifications/architecture/architecture.md)
- [Design](../specifications/design/design.md)
- [Plan: Story277 with Data Restructure](../plans/Story277_with_data_restructure_and_unit_checking_examples.md)
- [Story231: CPL Unit Type Checking](Story231_cpl-unit-type-checking.md)

## Dependencies

- [Story231_cpl-unit-type-checking.md](Story231_cpl-unit-type-checking.md) — In Progress (Task 4g, 3e, 5f incomplete). The unit type checker must be complete and wired into the synthesis pipeline before any example can exercise it.
- [Story230_cpl-conversions.md](finished-stories/Story230_cpl-conversions.md) — Done. Provides `CILConversionPair`, `CILMeasureConversionPair`, and conversion code synthesis needed by the correct-case example and error-case examples.
- [Story229_cpl-compound-units.md](finished-stories/Story229_cpl-compound-units.md) — Done. Compound unit expressions may appear in example CPL files.
- The existing `cpl_with_data_test.go` test infrastructure in `src/main/go/cli/compiler/`.

## Dependent Stories

- [Story279_cpl-conversion-injection-examples.md](Story279_cpl-conversion-injection-examples.md) — Not Started. Will extend `units_conversions/` with a `conversion_injection/` example that demonstrates automatic conversion code generation at runtime.

## Tasks

### Task Sizing

None of the three tasks qualifies as `[Small]`:

- **Task 1**: Modifies one test file with coordinated changes across 4 existing functions plus 1 new function, plus mechanical file moves spanning 5 category directories and 19 leaf examples. Not `[Small]` due to the breadth of changes and the need to reason about the full discovery mechanism in context.
- **Task 2**: Creates 4 error-case example directories (16 files total), each requiring internally consistent CPL, iospec.json, parquet data, and expected error text. Not `[Small]` due to multi-file coordination and CPL authoring complexity across unit-checking scenarios.
- **Task 3**: Creates 2 directories with 5 coordinated files (core.cpl, CPL, iospec.json, parquet data, ground-truth XLSX). Not `[Small]` due to the need for internal consistency across file types and numerically correct expected output.

### Parallel Execution

- **Group A** (Tasks 2, 3): Task 2 creates error example directories under `examples/with_data/with_errors/`. Task 3 creates the `units_conversions/` category and `conversion_bridge/` leaf example under `examples/with_data/without_errors/`. Disjoint directory sets, no code dependencies between them. Both depend on Task 1 completing first to establish the `without_errors/` and `with_errors/` directories.

### Execution Order

```
Task 1 → [Task 2 ‖ Task 3]
```

Task 1 establishes the `without_errors/` and `with_errors/` subdirectory structure and updates the discovery logic. Tasks 2 and 3 populate those directories with examples. Tasks 2 and 3 can run in parallel because they operate on disjoint directory trees.

### Reintegration

**Group A (Tasks 2, 3):**
- Merge order: any order (fully disjoint directory sets — `with_errors/` vs `without_errors/units_conversions/`). Neither task modifies files in the other's directory tree.
- Integration tests: `go test ./src/main/go/cli/compiler/... -count=1 -timeout 120s` — verifies that the new `conversion_bridge` example is discovered and passes the full compile → run → verify pipeline, and that all error-case examples are detected and produce their expected compilation failures.
- Watch for conflicts in: none expected. Task 3 creates `examples/with_data/without_errors/units_conversions/core.cpl`; Task 2 does not create any files under `units_conversions/`. The two worktrees do not share any tracked files.

### Task 1: Reorganize `examples/with_data/` directory structure — Completed

Reorganize the `with_data` examples directory to mirror the `parse_and_analyze_only/` layout, update the test discovery mechanism, add a new test for error-case compilation, and verify no existing tests regress.

#### 1a. Create new directory layout and move files

- Create `examples/with_data/without_errors/` directory.
- Create `examples/with_data/with_errors/` directory.
- Move all 5 category directories (`conditional_grain`, `metrics_arithmetic`, `metrics_conditionals`, `metrics_time_shift`, `multi_metric`) into `without_errors/` using `git mv` to preserve git history.
- Verify that each moved leaf example maintains its internal structure (iospec.json, .cpl files, .parquet files, expected_output.xlsx).

#### 1b. Update `cpl_with_data_test.go` discovery logic

- Update `discoverWithDataExamples()` in `src/main/go/cli/compiler/cpl_with_data_test.go` to walk `with_data/without_errors/` as the top-level category directory instead of `with_data/` directly. The second-level walking logic (category entry → leaf sub-entries) remains the same.
- Update `getWithDataDir()` documentation or return path if the path reference changes.

#### 1c. Add `TestWithDataExamples_WithErrors` test for error-case examples

Implement a new test function in `src/main/go/cli/compiler/cpl_with_data_test.go` that:
- Walks `with_data/with_errors/` using the same two-level discovery pattern (category entry → leaf sub-entries).
- For each leaf directory, collects `.cpl` files, reads `iospec.json` from the leaf directory, and calls the compiler.
- Asserts that compilation **fails** (the compiler returns a non-nil error).
- Reads `expected_error.txt` from the leaf directory and asserts that the compilation error message contains the substring specified in that file (substring match, not exact match).
- Does **not** attempt to run generated programs — error cases should fail during compilation (synthesis phase).
- Passes trivially if the `with_errors/` directory is empty (no false failures).
- Fails with a clear message if `expected_error.txt` is missing for a discovered error example.
- Fails with a clear message if an error example compiles successfully: "expected compilation error but got success for example X".

#### 1d. Verify existing tests still pass

All 19 existing examples must continue to pass. Run the full test suite to confirm no regressions:
```bash
go test ./src/main/go/cli/compiler/... -count=1 -timeout 120s
```

Confirm each individually:
- `TestWithDataExamples` — all 19 existing examples compile, run, and verify.
- `TestWithDataExamples_Discovery` — known examples still found.
- `TestWithDataExamples_ExpectedOutputExists` — no examples missing expected output.
- `TestWithDataExamples_Structure` — structural validation passes.
- `TestWithDataExamples_CompileOnly` — all examples compile.

### Task 2: Create error-case examples exercising unit type checking — Completed

Create 4 error-case example directories under `examples/with_data/with_errors/`. Each directory contains a self-contained CPL file (including all necessary connector, measure, unit, and conversion declarations), an `iospec.json`, minimal parquet data (1–2 rows; the compilation should fail at synthesis, not at data loading, but the discovery mechanism requires at least one `.parquet` file per leaf directory), and an `expected_error.txt` marker file.

#### 2a. `units_incompatible_add/` — incompatible addition

**Scenario**: A metric formula adds a USD-valued measure to a EUR-valued measure with no conversion defined between them.

CPL must declare:
- Concepts: `Currency = [Values[USD, EUR, GBP]]`
- Units: `Currency = Currency`
- Types: `Amount = Number(precision=2)`
- Measures: `revenue_usd = Amount(unit=Currency)` and `revenue_eur = Amount(unit=Currency)`
- Connector: provides data with both USD and EUR cells
- Metric: computes `input_usd + input_eur` — the unit checker must reject this as incompatible addition

**Expected error**: contains `cannot add`, `USD`, and `EUR`.

#### 2b. `units_io_spec_mismatch/` — IO spec unit mismatch without conversion

**Scenario**: A metric computes in EUR, the IO spec output declares `"unit": "USD"`, and no EUR↔USD conversion is defined.

CPL must declare:
- Same Currency concept, unit, and type infrastructure as 2a
- Connector: provides EUR-valued data
- Metric: identity passthrough of the connector input
- IO spec: `"outputs": [{ "metric": "revenue", ..., "unit": "USD" }]`

**Expected error**: contains `metric` and `unit check` or `no conversion available` (a synthesis-time error from `CheckMetricOutputUnit`).

#### 2c. `units_different_groups/` — incompatible unit groups

**Scenario**: The IO spec requires `Meter` but the metric computes `Celsius` — different unit groups with no inter-group conversion defined.

CPL must declare:
- Concepts for both `Length` and `TemperatureScale` domains
- Two separate unit groups
- Connector: provides Celsius-valued data
- Metric: identity passthrough
- IO spec: `"outputs": [{ "metric": "temp", ..., "unit": "Meter" }]`

**Expected error**: contains `no conversion available` and `different unit groups` (or similar phrasing from the unit checker).

#### 2d. `units_unannotated_constant/` — constant without unit annotation in conversion

**Scenario**: A conversion calculation contains a bare number (e.g., `32`) without an explicit unit annotation.

CPL must declare:
- A Celsius↔Fahrenheit conversion with a pattern or calculation that includes a bare `32` without a `(Celsius)` unit annotation
- A metric that triggers the conversion through an IO spec unit mismatch
- IO spec: requires a unit that would need the conversion

**Expected error**: contains `constant` and `unit annotation`.

### Task 3: Create correct-case end-to-end `units_conversions/conversion_bridge` example — Completed

Create a full `examples/with_data/without_errors/units_conversions/` category with a `conversion_bridge/` leaf example that demonstrates the full compilation pipeline with a unit conversion bridging a gap between computed and required units. The synthesis pipeline must recognize the unit gap is bridgeable (a `Kilometer→Meter` conversion pair exists), produce compilable Go code, and the generated program must produce numerically correct output in Meters.

#### 3a. Create `units_conversions/core.cpl`

Create `examples/with_data/without_errors/units_conversions/core.cpl` with shared infrastructure used by this example and by future examples (Story279):

```cpl
# Core infrastructure for units_conversions examples.
# Defines shared concepts, units, types, measures, and conversion pairs
# for testing unit type checking and conversion injection.

concepts
    Region = [Values[North, South]]
    Length = [Values[Meter, Kilometer, Centimeter]]
    Week = [Values[Monday]]

units
    Region = Region
    Length = Length
    Week = Week

types
    Distance = Number(precision=2)

measures
    SingleDay = Interval(unit=Week)
```

#### 3b. Create `conversion_bridge/` leaf example

Create `examples/with_data/without_errors/units_conversions/conversion_bridge/` with the following files:

**`conversion_bridge.cpl`**: Declares connector, metric, and conversion pair.
- Connector: `dimensions [Region] grain Daily celltype distance_km` where `distance_km = Distance(unit=Length)`
- Conversion: `Kilometer(var) = Meter(var) * 1000(Meter)` (with algebraic inverse derived by analyzer)
- Metric: identity passthrough from connector
- IO spec output: `{ "metric": "distance_meters", "dimensions": ["Region"], "celltype": "distance_km", "grain": ["Daily"], "unit": "Meter" }`

**`data_a.parquet`**: Parquet file with known Kilometer values (e.g., 1.0 and 2.5 for North/South × Monday).

**`expected_output.xlsx`**: Ground-truth XLSX output with Meter values (`Kilometer × 1000`): 1000.0 and 2500.0.

**`iospec.json`**: IO specification with `"unit": "Meter"` on the output metric.

The exact CPL syntax for conversions must match the actual syntax from Story230's implementation. The key requirement is that a `Kilometer→Meter` conversion pair exists in the compiled CPL.

#### 3c. Verify end-to-end

- The example is discovered by `TestWithDataExamples` under `without_errors/`.
- Compilation succeeds (the unit checker accepts the conversion bridge).
- The generated program runs and produces `distance_meters.parquet`.
- Output values match expected (`Meters = Kilometers × 1000`).
- All existing 19 `with_data` examples continue to pass (no regressions).

## Constraints

1. Follow the [Logical TDD Lifecycle](../concepts.md#logical-tdd-lifecycle) — for coding tasks, write tests before implementation. The test-and-implement cycle is integrated within each task; tests are not broken out as separate tasks.
2. All CPL files must use valid CPL syntax compatible with the current parser, analyzer, and synthesizer.
3. Error examples must include an `expected_error.txt` file containing a substring that the expected compilation error must contain.
4. Parquet data files for error examples need only contain minimal valid data (1–2 rows). Compilation should fail at synthesis, not at data loading.
5. The `conversion_bridge` example must produce numerically correct output (`Meters = Kilometers × 1000`).
6. Directory moves must preserve git history — use `git mv`, not `cp`/`rm`.
7. Existing tests must not regress — all 19 existing examples must continue to pass.
8. The `with_errors` test infrastructure must be compatible with the existing compiler CLI, which returns an error exit code on synthesis failure.
9. All error examples must produce errors during compilation (specifically the synthesis phase), not during parse/analyze or runtime.
10. Go files reside under `src/main/go/` within `source_root` (`src/chronocone_planning_language`). Build and test commands run from `source_root` using `./src/main/go/...` package paths. Example directories (`examples/with_data/...`) are at the project root, outside `source_root`.

## Edge Cases

- **Empty `with_errors/` directory**: `TestWithDataExamples_WithErrors` must pass trivially if no error examples exist, to avoid false failures from an empty directory.
- **Error example missing `expected_error.txt`**: The test must report which example is missing the marker, not silently pass.
- **Error example compiles unexpectedly**: The test must fail with a clear message: "expected compilation error but got success for example X".
- **Error message format drift**: Use substring matching, not exact matching, for `expected_error.txt` to tolerate minor message format changes.
- **Parquet data for error examples**: The discovery mechanism requires at least one `.parquet` file per example directory. Provide minimal valid parquet files even though compilation should fail before execution.

## Verification Commands

After all tasks complete, run the following from `source_root` (`src/chronocone_planning_language`):

```bash
# Verify existing examples still pass
go test ./src/main/go/cli/compiler/... -run TestWithDataExamples -count=1 -timeout 120s

# Verify error examples are tested
go test ./src/main/go/cli/compiler/... -run TestWithDataExamples_WithErrors -count=1 -timeout 60s

# Full CPL test suite
go test ./src/main/go/... -count=1 -timeout 180s
```

## Notes

- This story is the first to establish the `with_errors/` convention in `with_data/`. The pattern (directory per error case, `expected_error.txt` marker, test that asserts compilation failure) will be reused by future stories.
- The `conversion_bridge` example intentionally does NOT test conversion injection (Story232). At this point, the correct-case example only verifies that the unit checker permits compilation when a conversion pair exists. Story279 will add a `conversion_injection/` example demonstrating automatic conversion code generation at runtime.
- The `units_conversions/core.cpl` is designed to be extended by Story279, which will add a `conversion_injection/` example.
- Story231's own integration tests (`TestIntegration_ValidPipeline_ConversionBridgesGap` etc.) test the same functionality with synthetic `SynthesisInput` structs. These `with_data` examples test the same behavior through the real CPL file → compile → run → verify pipeline.
