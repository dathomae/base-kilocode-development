# Story279: With-Data Conversion Injection Example

## Goal

Create an end-to-end `with_data/without_errors/units_conversions/conversion_injection` example that exercises Story232's conversion injection pipeline. The example demonstrates that when a metric computes in one unit (EUR), the IO spec requires another unit (USD), and a valid conversion pair exists (EUR↔USD), the injection pass wraps the CIL tree with `CILConversionApplication` nodes, the synthesizer emits conversion function calls, and the generated code produces numerically correct output in the target unit (USD). This story extends the `units_conversions/core.cpl` shared infrastructure established by Story277 with Currency domain declarations and a EUR↔USD conversion pair, then creates the leaf example that validates the full pipeline end-to-end.

## References

- [Epic003: Units and Measures Implementation](../epics/Epic003_units-and-measures-implementation.md)
- [Story230_cpl-conversions.md](finished-stories/Story230_cpl-conversions.md) — provides conversion pair definition and synthesis
- [Story231_cpl-unit-type-checking.md](Story231_cpl-unit-type-checking.md) — provides unit mismatch identification
- [Story232_cpl-conversion-injection.md](Story232_cpl-conversion-injection.md) — provides automatic conversion injection into CIL trees
- [Story277_with_data-restructure-and-unit-checking-examples.md](Story277_with_data-restructure-and-unit-checking-examples.md) — establishes `units_conversions/` directory, `core.cpl`, and test discovery

## Dependencies

- [Story230_cpl-conversions.md](finished-stories/Story230_cpl-conversions.md) — _Done_ — provides `CILConversionPair` and conversion code synthesis
- [Story231_cpl-unit-type-checking.md](Story231_cpl-unit-type-checking.md) — _In Progress_ — provides `CheckMetricOutputUnit` and bridgeable-gap detection; checker must pass the example
- [Story232_cpl-conversion-injection.md](Story232_cpl-conversion-injection.md) — _Not Started_ — provides `InjectConversions` pipeline; must be complete and wired into synthesis before this example can run
- [Story277_with_data-restructure-and-unit-checking-examples.md](Story277_with_data-restructure-and-unit-checking-examples.md) — _Not Started_ — establishes `examples/with_data/without_errors/units_conversions/` directory, `core.cpl` with Length infrastructure, and the `conversion_bridge` example. The `core.cpl` and directory structure must exist before this story can extend them.

## Dependent Stories

None

## Tasks

### Task Sizing

- **Task 1** (Core infrastructure): `[Small]` — extends a single file (`core.cpl`) with Currency domain declarations following the existing Length infrastructure pattern. Purely additive, single concern, single file, no new abstractions beyond what is already templated in the file.
- **Task 2** (Injection example): Medium — creates 4+ coordinated files (CPL, iospec.json, parquet data, expected output XLSX) in a new leaf example directory. Requires CPL authoring, IO spec construction, data generation, expected output computation, and end-to-end verification. Not `[Small]` due to multiple coordinated file types requiring internal consistency. The task is at its natural granularity — the CPL cannot compile without the iospec.json and parquet data, so no natural seam exists for further decomposition.

### Parallel Execution

None — all tasks have sequential dependencies. Task 2 semantically depends on Task 1's Currency infrastructure being present in `core.cpl`; the `conversion_injection.cpl` file will reference concepts, units, types, measures, and conversion pairs that Task 1 adds.

### Execution Order

```
Task 1 → Task 2
```

- Task 1 establishes the shared Currency infrastructure that Task 2's example depends on.

### Reintegration

None — no parallel groups.

### Task 1: Extend `units_conversions/core.cpl` with Currency infrastructure — Not Started [Small]
   a. Add Currency concept declaration following the existing `Length` pattern — Not Started
      - Declare `Currency = [Values[EUR, USD, GBP]]` in the concepts block
      - Ensure no conflict with existing Length concept
   b. Add Currency unit, type, and measure declarations — Not Started
      - Declare `units Currency = Currency`
      - Declare `types Amount = Number(precision=2)` (or add to existing types block if one exists)
      - Declare measures for revenue amounts (e.g., `revenue = Amount(unit=Currency)`)
   c. Add EUR↔USD conversion pair using Story230's CPL conversion syntax — Not Started
      - Declare the conversion with a non-trivial factor (e.g., 1 EUR = 1.10 USD) so injection is detectable
      - The conversion must be algebraically reversible (Story230 derives the inverse) so both directions work
      - Reference Story230's integration tests for the exact CPL conversion syntax
   d. Verify `core.cpl` remains backward-compatible with Story277's `conversion_bridge` example — Not Started
      - The existing Length infrastructure must not be modified or broken
      - Compile the `conversion_bridge` example (or the `units_conversions/` category as a whole) to confirm no regressions

**File**: `examples/with_data/without_errors/units_conversions/core.cpl` (relative to `source_root`)

### Task 2: Create `conversion_injection` leaf example — Not Started
   a. Create directory and CPL file — Not Started
      - Create directory: `examples/with_data/without_errors/units_conversions/conversion_injection/`
      - Write `conversion_injection.cpl`:
        - Connector: provides EUR data with dimensions (e.g., `[Region]`), grain (e.g., `SingleDay`), and celltype referencing the Currency measure from `core.cpl`
        - Metric: computes revenue as identity passthrough of the connector input
        - All concept, unit, type, measure, and conversion declarations come from `core.cpl` (via the `#include` or shared compilation mechanism used by Story277)
   b. Create IO specification — Not Started
      - Write `iospec.json`:
        - Input: connector_name → parquet file (`data_a.parquet`) via `"file"` driver
        - Output: metric name, dimensions matching the connector, celltype, grain, and `"unit": "USD"` to trigger the conversion injection
        - Architecture block with cpu_cores, cache_size, main_memory_size (follow existing example conventions)
   c. Create parquet input data — Not Started
      - Generate `data_a.parquet` with EUR values for each dimension combination
      - Use simple values to enable manual verification: e.g., North=100.0, South=200.0 (in EUR)
      - Ensure columns match the connector's dimension structure (e.g., Region × SingleDay × Value)
   d. Create expected output — Not Started
      - Compute expected USD values by applying the conversion factor from Task 1c to each EUR input value
      - Create `expected_output.xlsx` with the computed values
      - Use the `xls_verify` tool or manually construct the XLSX following the pattern from existing `with_data` examples
      - If the conversion factor is 1.10, expect: North=110.0, South=220.0 (in USD)
   e. Verify end-to-end — Not Started
      - Run `go test ./src/main/go/cli/compiler/... -run TestWithDataExamples -count=1 -timeout 120s` from `source_root` (`src/chronocone_planning_language`)
      - Confirm the `conversion_injection` example is discovered by the test
      - Confirm compilation succeeds (checker permits the bridgeable gap, injection modifies CIL tree, synthesizer emits code)
      - Confirm generated program runs without error
      - Confirm output values match `expected_output.xlsx`
      - Confirm all existing `with_data` examples continue to pass (including Story277's `conversion_bridge`)
   f. (Optional) Verify generated code contains conversion call — Not Started
      - After compilation, inspect the generated Go source to confirm a conversion function call (e.g., `convert_EUR_to_USD`) appears in the output
      - This is a stronger assertion than runtime verification alone — it proves injection occurred rather than the output just happening to match

**Files**: All under `examples/with_data/without_errors/units_conversions/conversion_injection/` (relative to `source_root`)

## Test-First Development

This story follows the Logical TDD Lifecycle:

- **Task 1**: No test-first cycle applies — the task extends a shared CPL infrastructure file. Backward-compatibility is verified by compiling the existing `conversion_bridge` example.
- **Task 2**: The example itself is the test. The acceptance criteria are: compilation succeeds, generated program runs, output matches expected. The `TestWithDataExamples` test infrastructure provides the automated verification framework. For the optional generated-code inspection (subtask 2f), write the assertion before declaring the task complete — this is an additive, internal test within the example verification.

## Constraints

1. Follow the [Logical TDD Lifecycle](../concepts.md#logical-tdd-lifecycle)
2. All CPL must use valid syntax compatible with the current compiler pipeline (parser, analyzer, synthesizer)
3. The `core.cpl` modification must not break Story277's `conversion_bridge` example (which uses Length infrastructure). The new Currency declarations must be purely additive.
4. The conversion pair must be declared using Story230's actual CPL conversion syntax. Reference Story230's integration tests (e.g., `TestIntegration_ValidPipeline_ConversionBridgesGap`) for the exact syntax.
5. The example must be discoverable by `TestWithDataExamples` — requires `iospec.json`, at least one `.cpl` file, at least one `.parquet` file in the leaf directory
6. Numerical output must be correct — the conversion arithmetic must match the CPL conversion formula. Use known inputs and a clean conversion factor to enable manual verification.
7. This example does NOT need to test error cases — error cases are covered by Story277's `with_errors/` examples.
8. Use a non-trivial conversion factor (e.g., 1 EUR = 1.10 USD) — a factor of 1.0 would not prove injection occurred since input and output would be numerically identical.
9. All file paths in task descriptions are relative to `source_root` (`src/chronocone_planning_language`). Go source files live under `src/main/go/` within `source_root`; example files live under `examples/` at the `source_root` level. Go test commands use fully-qualified package paths from `source_root`.

## Intent

### Task 1
Ensure the `units_conversions/core.cpl` shared infrastructure includes Currency domain declarations (concepts, units, types, measures, and a EUR↔USD conversion pair) so that the `conversion_injection` example (Task 2) can reference them. The existing Length infrastructure from Story277 must remain intact.

### Task 2
Provide a concrete, automated `with_data` example that proves Story232's conversion injection pipeline works end-to-end: CPL authored → compiled (checker permits, injection fires, synthesizer emits conversion code) → generated program runs → output in target unit matches ground truth. This example complements Story277's `conversion_bridge` (which only verifies the checker permits compilation) by also verifying that injection actually produces a runtime conversion.

## Acceptance Criteria

### Task 1
- `core.cpl` contains Currency concept, unit, type, measure, and EUR↔USD conversion pair declarations
- The conversion pair uses Story230's actual CPL conversion syntax (not pseudocode)
- Story277's `conversion_bridge` example compiles and passes without modification — the Length infrastructure is unchanged

### Task 2
- The `conversion_injection` leaf directory exists with all required files: `conversion_injection.cpl`, `iospec.json`, `data_a.parquet`, `expected_output.xlsx`
- `TestWithDataExamples` discovers and runs the example without failure
- Compilation succeeds without errors (checker permits the EUR→USD bridgeable gap; injection wraps CIL tree; synthesizer emits code)
- Generated program runs without runtime errors
- Output values equal the expected values in `expected_output.xlsx` (EUR values correctly converted to USD)
- All existing `with_data` examples continue to pass (no regressions)

## Notes

- This story is intentionally narrow in scope — a single example that proves the conversion injection pipeline works end-to-end. This follows the project pattern where each `with_data` example validates one specific feature or scenario.
- The `units_conversions/` category will have at least two examples after this story: `conversion_bridge` (Story277 — checker permits compilation) and `conversion_injection` (Story279 — injection produces working runtime code).
- This example intentionally tests the simple case: identity metric with a single conversion at the output boundary. More complex scenarios (conversions injected into operand positions, multiple conversions in one expression) are tested by Story232's own Go integration tests.
- If Story277 has already added Currency infrastructure to `core.cpl` (beyond the Length declarations in the plan), Task 1 becomes a minor extension or no-op. The architect-for-story-planning decomposition review confirmed this is a single-file additive change following an established pattern.
- The optional generated-code inspection (subtask 2f) adds confidence that injection actually modified the CIL tree, but the runtime verification (output matches expected) is the primary acceptance criterion.
