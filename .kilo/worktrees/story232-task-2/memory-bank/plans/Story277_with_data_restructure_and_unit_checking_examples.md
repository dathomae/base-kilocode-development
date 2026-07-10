# Plan: Story277 — Reorganize `with_data` and Add Unit-Checking Examples

## Purpose

Reorganize `examples/with_data/` into `with_data/without_errors/` and `with_data/with_errors/` subdirectories (mirroring `parse_and_analyze_only/`), move all existing 19 leaf examples into `without_errors/`, update the test discovery infrastructure accordingly, and create both end-to-end `with_data/without_errors/` examples and error-case `with_data/with_errors/` examples that exercise Story231's unit type checker through the full compilation pipeline.

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `with_data-restructure-and-unit-checking-examples`                                           |
| story_number | 277                                                                                          |
| description  | Reorganize examples/with_data into without_errors/ and with_errors/ subdirectories, move existing examples, update test discovery, and create end-to-end and error-case examples exercising Story231's synthesis-time unit type checking |

## Scope

This plan covers three distinct work areas, each a separate task:

1. **Directory restructuring** — mechanical move of existing files + test infrastructure update
2. **Error-case examples** — CPL files that must fail compilation with specific unit-checking errors
3. **Correct-case example** — a full `with_data/without_errors/` example with parquet data showing conversion bridging a unit gap

## Source Stories / EROs

- [Story231_cpl-unit-type-checking](../stories/Story231_cpl-unit-type-checking.md) — provides `CheckMetricOutputUnit`, `ComputeUnitExpression`, `UnitExpression` simplification, `ValidateConversionClosure`, and the `Unit` field on `OutputSpecification`
- [Story230_cpl-conversions](../stories/finished-stories/Story230_cpl-conversions.md) — provides `CILConversionPair`, `CILMeasureConversionPair`, and conversion code synthesis
- [Epic003: Units and Measures Implementation](../epics/Epic003_units-and-measures-implementation.md)

## Dependencies

- Story231 (unit type checking) — the checker must be complete and wired into the pipeline before any example can exercise it
- Story230 (conversions) — conversion pairs needed for the correct-case example and for error-case examples that test missing conversions
- Story229 (compound units) — compound unit expressions may appear in example CPL files
- The existing `cpl_with_data_test.go` test infrastructure (must be modified for the new directory layout)

## Files to Create or Modify

| File                                                                    | Action   | Purpose                                                                                                     |
| ----------------------------------------------------------------------- | -------- | ----------------------------------------------------------------------------------------------------------- |
| `examples/with_data/without_errors/`                                    | Create   | New directory — all existing example categories moved here                                                  |
| `examples/with_data/with_errors/`                                       | Create   | New directory — holds error-case example directories                                                        |
| `examples/with_data/without_errors/metrics_arithmetic/...` (all existing) | Move     | All 5 existing category directories (conditional_grain, metrics_arithmetic, metrics_conditionals, metrics_time_shift, multi_metric) moved under `without_errors/` |
| `src/main/go/cli/compiler/cpl_with_data_test.go`                        | Modify   | Update discovery to walk `without_errors/`; add `TestWithDataExamples_WithErrors` test that walks `with_errors/` |
| `examples/with_data/without_errors/units_conversions/`                  | Create   | New category for unit/conversion examples (will also be used by Story279)                                   |
| `examples/with_data/without_errors/units_conversions/core.cpl`          | Create   | Shared concepts, units, types, measures, and conversion pairs for unit-checking examples                    |
| `examples/with_data/without_errors/units_conversions/conversion_bridge/` | Create  | Leaf example: Kilometer→Meter conversion bridging a unit gap                                               |
| `examples/with_data/without_errors/units_conversions/conversion_bridge/iospec.json` | Create | IO spec with `"unit": "Meter"` on output, parquet connectors                                                |
| `examples/with_data/without_errors/units_conversions/conversion_bridge/conversion_bridge.cpl` | Create | CPL: connector input in Kilometers, metric computes Kilometers, IO spec requires Meters, conversion pair exists |
| `examples/with_data/without_errors/units_conversions/conversion_bridge/data_a.parquet` | Create | Parquet input data with Kilometer values                                                                    |
| `examples/with_data/without_errors/units_conversions/conversion_bridge/expected_output.xlsx` | Create | Expected output with Meter values (verifying the conversion produces correct numerical results)               |
| `examples/with_data/with_errors/units_incompatible_add/`                | Create   | Error example: incompatible addition of USD + EUR                                                           |
| `examples/with_data/with_errors/units_incompatible_add/iospec.json`     | Create   | IO spec for the incompatible add example                                                                    |
| `examples/with_data/with_errors/units_incompatible_add/incompatible_add.cpl` | Create | CPL: metric that adds USD + EUR, no conversion defined                                                       |
| `examples/with_data/with_errors/units_incompatible_add/data_a.parquet`  | Create   | Parquet input data                                                                                          |
| `examples/with_data/with_errors/units_io_spec_mismatch/`                | Create   | Error example: metric computes EUR but IO spec requires USD, no conversion                                  |
| `examples/with_data/with_errors/units_io_spec_mismatch/iospec.json`     | Create   | IO spec with `"unit": "USD"` on output                                                                      |
| `examples/with_data/with_errors/units_io_spec_mismatch/io_spec_mismatch.cpl` | Create | CPL: metric computes EUR, IO spec requires USD, no EUR↔USD conversion defined                                |
| `examples/with_data/with_errors/units_io_spec_mismatch/data_a.parquet`  | Create   | Parquet input data                                                                                          |
| `examples/with_data/with_errors/units_different_groups/`                | Create   | Error example: IO spec requires Meter but metric computes Celsius (different unit groups)                   |
| `examples/with_data/with_errors/units_different_groups/iospec.json`     | Create   | IO spec with `"unit": "Meter"` on output                                                                    |
| `examples/with_data/with_errors/units_different_groups/different_groups.cpl` | Create | CPL: metric computes Celsius, IO spec requires Meter, Celsius→Meter conversion undefined (different groups)  |
| `examples/with_data/with_errors/units_different_groups/data_a.parquet`  | Create   | Parquet input data                                                                                          |
| `examples/with_data/with_errors/units_unannotated_constant/`            | Create   | Error example: constant in conversion calculation lacks unit annotation                                     |
| `examples/with_data/with_errors/units_unannotated_constant/iospec.json` | Create   | IO spec                                                                                                     |
| `examples/with_data/with_errors/units_unannotated_constant/unannotated_constant.cpl` | Create | CPL: conversion calculation contains bare number without unit annotation                                    |
| `examples/with_data/with_errors/units_unannotated_constant/data_a.parquet` | Create | Parquet input data                                                                                          |

## Tasks

### Task Sizing

- **Task 1** (Directory restructure): Medium — mechanical file moves across 19 leaf + 5 category directories; updating 4 test functions in `cpl_with_data_test.go`; adding a new test for the `with_errors/` directory. Not `[Small]` due to the number of file moves and the test infrastructure changes.
- **Task 2** (Error-case examples): Medium — creating 4 complete error-case example directories (CPL, iospec.json, parquet data, plus a test that verifies expected compilation errors). Not `[Small]` due to CPL authoring complexity and the need for the new `with_errors` test pattern.
- **Task 3** (Correct-case example): Medium — creating a full `with_data/without_errors/units_conversions/conversion_bridge` example with CPL, parquet data, ground-truth XLSX, and parquet verification. Not `[Small]` due to multiple coordinated file types requiring internal consistency.

### Parallel Execution

- **Group A** (Tasks 2, 3): Task 2 creates directories under `with_errors/`; Task 3 creates directories under `without_errors/units_conversions/`. Disjoint directory sets. Task 3 needs the `core.cpl` for `units_conversions/` (defines shared concepts/units/types/measures/conversions), which can be created first in Task 3 or shared.

### Execution Order

```
Task 1 → [Task 2 ‖ Task 3]
```

- Task 1 must complete first: establishes the `without_errors/` and `with_errors/` directory structure that Tasks 2 and 3 populate.
- Tasks 2 and 3 can run in parallel: disjoint directories, no code dependencies between them.

### Reintegration

**Group A (Tasks 2, 3):**
- Merge order: any order (disjoint directories)
- Rebuild verification: `go test ./src/main/go/cli/compiler/... -count=1 -timeout 120s` plus Parquet test data reconstruction
- Watch for conflicts in: `core.cpl` for `units_conversions/` — Task 3 creates this file; both tasks may reference it conceptually but only Task 3 modifies it

### Task 1: Reorganize `examples/with_data/` directory structure — Not Started

#### 1a. Create new directory layout and move files

- Create `examples/with_data/without_errors/` directory
- Create `examples/with_data/with_errors/` directory
- Move all 5 category directories (conditional_grain, metrics_arithmetic, metrics_conditionals, metrics_time_shift, multi_metric) into `without_errors/`
- Verify that each moved leaf example maintains its internal structure (iospec.json, .cpl, .parquet, expected_output.xlsx)

#### 1b. Update `cpl_with_data_test.go` discovery logic

- Update `discoverWithDataExamples()` to walk `with_data/without_errors/` instead of `with_data/` directly
- The second-level walking logic (category → leaf directories) remains the same
- Update `getWithDataDir()` documentation if needed (path reference changes)

#### 1c. Add `TestWithDataExamples_WithErrors` test for error-case examples

- Implement a new test function that:
  - Walks `with_data/with_errors/` directory (same discovery pattern)
  - For each leaf directory, attempts compilation
  - Asserts that compilation **fails** with a specific expected error
  - Uses a marker file convention: each error example directory contains an `expected_error.txt` that specifies the expected error substring
  - If an error example compiles successfully, the test fails
- The test should NOT run generated programs (error cases should not reach code generation)

#### 1d. Verify existing tests still pass

- `TestWithDataExamples` — all 19 existing examples compile, run, and verify
- `TestWithDataExamples_Discovery` — known examples still found
- `TestWithDataExamples_ExpectedOutputExists` — no examples missing expected output
- `TestWithDataExamples_Structure` — structural validation passes
- `TestWithDataExamples_CompileOnly` — all examples compile

### Task 2: Create error-case examples exercising unit type checking — Not Started

Create 4 error-case example directories under `with_data/with_errors/`. Each directory contains a self-contained CPL file (including all necessary connector, measure, unit, and conversion declarations) an iospec.json, minimal parquet data (the compilation should fail before execution, but parquet files are needed for the compiler to proceed to synthesis where unit checking occurs), and an `expected_error.txt` marker.

#### 2a. `units_incompatible_add/` — incompatible addition

**Scenario**: Metric formula adds USD-valued measure to EUR-valued measure, no conversion defined.

CPL should declare:
- Concepts: `Currency = [Values[USD, EUR, GBP]]`
- Units: `Currency = Currency`
- Types: `Amount = Number(precision=2)`
- Measures: `revenue_usd = Amount(unit=Currency)` and `revenue_eur = Amount(unit=Currency)` — these share the same unit group but the formula adds them directly
- Connector: provides data in USD and EUR
- Metric: computes `input_usd + input_eur` — the unit checker should reject this as incompatible addition

**Expected error**: contains "cannot add" and "USD" and "EUR" (the specific error message from Story231 Task 4e)

#### 2b. `units_io_spec_mismatch/` — IO spec unit mismatch without conversion

**Scenario**: Metric computes in EUR, IO spec output declares `"unit": "USD"`, no EUR↔USD conversion defined.

CPL should declare:
- Same Currency concept/unit/type infrastructure
- Connector: provides EUR data
- Metric: identity passthrough of input
- IO spec: `"outputs": [{ "metric": "revenue", ..., "unit": "USD" }]`

**Expected error**: contains "metric" and "unit check failed" and "no conversion available" (or similar synthesis-time error from `CheckMetricOutputUnit`)

#### 2c. `units_different_groups/` — incompatible unit groups

**Scenario**: IO spec requires `Meter` but metric computes `Celsius` — different unit groups with no inter-group conversion.

CPL should declare:
- Concepts for both `Length` and `TemperatureScale` domains
- Two separate unit groups
- Connector: provides Celsius data
- Metric: identity passthrough
- IO spec: `"outputs": [{ "metric": "temp", ..., "unit": "Meter" }]`

**Expected error**: contains "no conversion available" and "different unit groups" (or similar)

#### 2d. `units_unannotated_constant/` — constant without unit annotation in conversion

**Scenario**: A conversion calculation contains a bare number (e.g., `32`) without an explicit unit annotation.

CPL should declare:
- Celsius↔Fahrenheit conversion with a pattern/calculation that includes `32` without `(Celsius)` annotation
- Metric that triggers the conversion through IO spec mismatch
- IO spec: requires unit to be something that would need the conversion

**Expected error**: contains "constant" and "unit annotation" (from Story231 Task 4e)

### Task 3: Create correct-case end-to-end `units_conversions/conversion_bridge` example — Not Started

Create a full `with_data/without_errors/units_conversions/` category with a `conversion_bridge/` leaf example that demonstrates the full pipeline with a unit conversion bridging a gap between computed and required units.

#### 3a. Create `units_conversions/core.cpl`

Shared infrastructure used by this example (and future examples added by Story279):

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

**Scenario**: Connector provides data in Kilometers. Metric computes in Kilometers. IO spec declares output unit as Meter. A `Kilometer→Meter` conversion pair exists. The synthesis pipeline should recognize the unit gap is bridgeable and produce compilable code. The generated code performs the conversion at runtime, producing output in Meters.

Files:
- `conversion_bridge.cpl`: Declares connector, metric, and conversion pair
  - Connector: `dimensions [Region] grain Daily celltype distance_km` (where `distance_km = Distance(unit=Length)`)
  - Conversion: `Kilometer(var) = Meter(var) * 1000(Meter)` (with algebraic inverse derived by analyzer)
  - Metric: identity passthrough from connector
  - IO spec output: `{ "metric": "distance_meters", "dimensions": ["Region"], "celltype": "distance_km", "grain": ["Daily"], "unit": "Meter" }`
- `data_a.parquet`: Parquet file with known Kilometer values (e.g., 1.0, 2.5 for North/South × Monday)
- `expected_output.xlsx`: Ground-truth output with Meter values (1000.0, 2500.0)
- `iospec.json`: IO specification with `"unit": "Meter"` on output

Note on conversion CPL syntax: If the exact CPL syntax for conversions differs from the pseudocode above, use the actual syntax from Story230's implementation. The key requirement is that a `Kilometer→Meter` conversion exists in the compiled CPL.

#### 3c. Verify end-to-end

- The example is discovered by `TestWithDataExamples` (under `without_errors/`)
- Compilation succeeds (unit checker accepts the conversion bridge)
- Generated program runs and produces `distance_meters.parquet`
- Output values match expected (Meters = Kilometers × 1000)
- All existing `with_data` examples continue to pass (no regressions)

## Constraints

1. Follow the [Logical TDD Lifecycle](../concepts.md#logical-tdd-lifecycle) — for coding tasks, tests before implementation
2. All CPL files must use valid CPL syntax compatible with the current parser, analyzer, and synthesizer
3. Error examples must include an `expected_error.txt` file with a substring that the expected compilation error must contain
4. Parquet data files for error examples need only contain minimal valid data (1-2 rows) — compilation should fail at synthesis, not at data loading
5. The `conversion_bridge` example must produce numerically correct output (Meters = Kilometers × 1000)
6. Directory moves must preserve git history — use `git mv` not `cp`/`rm`
7. Existing tests must not regress — all 19 examples must continue to pass
8. The `with_errors` test infrastructure must be compatible with the existing compiler CLI (which returns an error exit code on synthesis failure)
9. All error examples must produce errors during compilation (synthesis phase), not during parse/analyze or runtime

## Edge Cases

- **Empty `with_errors/` directory**: The `TestWithDataExamples_WithErrors` test should pass trivially if no error examples exist (avoid false failures from an empty directory)
- **Error example missing `expected_error.txt`**: The test should report which example is missing the marker, not silently pass
- **Error example compiles unexpectedly**: The test should fail with a clear message: "expected compilation error but got success for example X"
- **Error message format drift**: Use substring matching, not exact matching, for `expected_error.txt` to allow for minor message format changes
- **Parquet data for error examples**: Even though compilation should fail before execution, the discovery mechanism requires at least one `.parquet` file per example directory. Provide minimal valid parquet files.

## Verification Commands

After all tasks complete:

```bash
# Verify existing examples still pass
go test ./src/main/go/cli/compiler/... -run TestWithDataExamples -count=1 -timeout 120s

# Verify error examples are tested
go test ./src/main/go/cli/compiler/... -run TestWithDataExamples_WithErrors -count=1 -timeout 60s

# Full CPL test suite
go test ./src/chronocone_planning_language/src/main/go/... -count=1 -timeout 180s
```

## Notes

- This story is the first to establish the `with_errors` convention in `with_data/`. The pattern (directory per error case, `expected_error.txt` marker, test that asserts compilation failure) will be reused by future stories.
- The `conversion_bridge` example intentionally does NOT test conversion injection (Story232). At this point, the correct-case example only verifies that the unit checker permits the compilation when a conversion pair exists. A separate story (Story279) will add an example where conversion injection actually produces working runtime code.
- The `units_conversions/core.cpl` is designed to be extended by Story279, which will add a `conversion_injection/` example that demonstrates automatic conversion application.
- Story231's own integration tests (`TestIntegration_ValidPipeline_ConversionBridgesGap` etc.) test the same functionality but with synthetic `SynthesisInput` structs. These `with_data` examples test the same behavior through the real CPL file → compile → run → verify pipeline.
