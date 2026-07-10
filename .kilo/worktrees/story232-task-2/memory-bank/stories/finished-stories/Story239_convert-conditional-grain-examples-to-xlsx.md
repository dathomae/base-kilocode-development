# Story239: Convert Conditional and Grain with_data Examples to XLSX Verification

## Goal

Convert all 7 currently-skipped with_data examples (6 metrics_conditionals + 1 conditional_grain) from JSON/ODS verification to XLSX verification using the PDD script created in Story 237.

## References

- [Test Conversion Plan](../plans/test-conversion-plan.md)
- [Story237_create-xlsx-conversion-script.md](Story237_create-xlsx-conversion-script.md) — Not Started
- [Story238_convert-arithmetic-examples-to-xlsx.md](Story238_convert-arithmetic-examples-to-xlsx.md) — Not Started

## Dependencies

- [Story237_create-xlsx-conversion-script.md](Story237_create-xlsx-conversion-script.md) — Not Started
- [Story238_convert-arithmetic-examples-to-xlsx.md](Story238_convert-arithmetic-examples-to-xlsx.md) — Not Started
  Should complete before this story to ensure the conversion pipeline is validated on passing examples first.

## Dependent Stories

- [Story240_delete-ods-code-and-resources.md](Story240_delete-ods-code-and-resources.md) — Not Started

## Tasks

### Task Sizing

Tasks 1 and 2 are medium: each spans multiple example directories but follows the same PDD script pattern. Task 3 is a verification and documentation task.

### Parallel Execution

- **Group B** (Tasks 1, 2): Task 1 converts 6 skipped conditional examples; Task 2 converts the grain example. Disjoint directories, no shared files.

### Execution Order

```
[Task 1 ‖ Task 2] → Task 3
```

### Reintegration

**Group B (Tasks 1, 2):**
- Merge order: any order (disjoint example directories)
- Integration tests: `cd src/chronocone_planning_language && go test ./src/main/go/cli/compiler/... -count=1 -timeout 300s`
- Watch for conflicts in: none expected

### Task 1: Convert skipped conditional examples — Completed

**Intent**: Run the PDD script for all 6 currently-skipped metrics_conditionals examples to regenerate their expected output as XLSX and clean up redundant files.

a. Execute the PDD script for each of the 6 skipped metrics_conditionals examples:
   - `conditional_comparison_operators`
   - `conditional_if_else_arithmetic`
   - `conditional_logical_operators`
   - `conditional_when_default`
   - `conditional_nested_arithmetic`
   - `conditional_truthiness`
   — Not Started

b. Verify each example directory now contains `expected_output.xlsx` and `expected_output_schema.json` — Not Started

c. Confirm redundant files are deleted from each directory — Not Started

**Acceptance Criteria**:
- All 6 skipped metrics_conditionals directories contain `expected_output.xlsx` and `expected_output_schema.json`.
- No example directory contains `expected_output.json`, `expected_output.ods`, or `parquet_gen_spec_*.json`.
- `conditional_comparison_operators`'s `.xlsx` was regenerated (not relying on the pre-existing file).

### Task 2: Convert monthly_data grain example — Completed

**Intent**: Run the PDD script for the single grain example to complete the with_data conversion.

a. Execute the PDD script for `conditional_grain/monthly_data` — Not Started

b. Verify the example directory now contains `expected_output.xlsx` and `expected_output_schema.json` — Not Started

c. Confirm redundant files are deleted from the directory — Not Started

**Acceptance Criteria**:
- `conditional_grain/monthly_data` contains `expected_output.xlsx` and `expected_output_schema.json`.
- No `expected_output.json`, `expected_output.ods`, or `parquet_gen_spec_*.json` remain.

### Task 3: Run full test suite and document results — Completed

**Intent**: Verify the full test suite with all examples converted, documenting which previously-skipped examples now pass and which still fail.

a. Run `cd src/chronocone_planning_language && go test ./src/main/go/cli/compiler/... -count=1 -timeout 300s` — Not Started

b. For each example that now passes, confirm it is no longer skipped — Not Started

c. For any examples that still fail, document the failure (example name, mismatched cells, failure reason) but do NOT attempt fixes — Not Started

**Acceptance Criteria**:
- Test results are documented: which of the 7 examples pass, which fail (with reasons).
- All previously-skipped examples are no longer skipped in the test code (this was done in Story237).
- No fixes are attempted for pre-existing data discrepancies.

## Constraints

- These examples were skipped because `expected_output.json` had incorrect expectations. The XLSX expected output is computed from current `data_*.json` via the CIL evaluator, so it reflects the actual expected values.
- Pre-existing data discrepancies (if any remain after conversion) are NOT to be fixed in this story.
- `conditional_comparison_operators` already has `.xlsx` and `.ods` files — it must be regenerated to ensure correctness.

## Notes

- These 7 examples correspond to Phase 3b and 3c of the test-conversion-plan.
- Unlike Story238, some of these examples may still fail after conversion due to pre-existing data issues. The goal is to convert the verification format, not to fix all data mismatches.
