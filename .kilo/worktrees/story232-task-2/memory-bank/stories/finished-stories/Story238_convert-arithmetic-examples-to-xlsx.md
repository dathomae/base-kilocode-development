# Story238: Convert Arithmetic with_data Examples to XLSX Verification

## Goal

Convert all 11 currently-passing with_data examples (7 metrics_arithmetic + 4 simple metrics_conditionals) from JSON/ODS verification to XLSX verification using the PDD script created in Story 237.

## References

- [Test Conversion Plan](../plans/test-conversion-plan.md)
- [Story237_create-xlsx-conversion-script.md](Story237_create-xlsx-conversion-script.md) — Not Started

## Dependencies

- [Story237_create-xlsx-conversion-script.md](Story237_create-xlsx-conversion-script.md) — Not Started
  Must complete before this story can begin; the PDD script and test driver changes are required.
- [Story219_ods-generator-and-verifier.md](../finished-stories/Story219_ods-generator-and-verifier.md) — Done

## Dependent Stories

- [Story239_convert-conditional-grain-examples-to-xlsx.md](Story239_convert-conditional-grain-examples-to-xlsx.md) — Not Started

## Tasks

### Task Sizing

Tasks 1 and 2 are medium: each spans multiple example directories but follows the same PDD script pattern with no cross-directory file modifications. Task 3 is a test verification.

### Parallel Execution

- **Group A** (Tasks 1, 2): Task 1 converts metrics_arithmetic examples; Task 2 converts metrics_conditionals simple examples. They touch disjoint example directories and share no files.

### Execution Order

```
[Task 1 ‖ Task 2] → Task 3
```

### Reintegration

**Group A (Tasks 1, 2):**
- Merge order: any order (disjoint example directories, no shared infrastructure)
- Integration tests: `cd src/chronocone_planning_language && go test ./src/main/go/cli/compiler/... -count=1 -timeout 300s`
- Watch for conflicts in: none expected

### Task 1: Convert metrics_arithmetic examples — Completed

**Intent**: Run the PDD script for all 7 arithmetic examples to regenerate their expected output as XLSX and clean up redundant files.

a. Execute the PDD script for each of the 7 metrics_arithmetic examples:
   - `simple_addition` — Completed (pre-converted before story start)
   - `short_form_addition` — Completed
   - `unary_negation` — Completed
   - `subtraction_division` — Completed
   - `multiplication_precedence` — Completed
   - `parenthesized` — Completed
   - `complex_expression` — Completed

b. Verify each example directory now contains `expected_output.xlsx` and `expected_output_schema.json` — Confirmed for all 7

c. Confirm redundant files (`expected_output.json`, `expected_output.ods`, `parquet_gen_spec_*.json`) are deleted from each directory — Confirmed for all 7

**Acceptance Criteria**:
- All 7 metrics_arithmetic directories contain `expected_output.xlsx` and `expected_output_schema.json`.
- No metrics_arithmetic directory contains `expected_output.json`, `expected_output.ods`, or `parquet_gen_spec_*.json`.
- `simple_addition`'s `.xlsx` was regenerated (not relying on the pre-existing file).

### Task 2: Convert metrics_conditionals simple examples — Completed

**Intent**: Run the PDD script for the 4 simple conditional examples that are already passing tests.

a. Execute the PDD script for each of the 4 metrics_conditionals examples:
   - `conditional_if_elseif_else` — Completed
   - `conditional_short_form` — Completed
   - `conditional_tautological_if_only` — Completed
   - `conditional_unary_negation` — Completed

b. Verify each example directory now contains `expected_output.xlsx` and `expected_output_schema.json` — Confirmed for all 4

c. Confirm redundant files are deleted from each directory — Confirmed for all 4

**Acceptance Criteria**:
- All 4 metrics_conditionals simple example directories contain `expected_output.xlsx` and `expected_output_schema.json`.
- No example directory contains `expected_output.json`, `expected_output.ods`, or `parquet_gen_spec_*.json`.

### Task 3: Run full test suite and verify — Completed

**Intent**: Confirm that all 11 converted examples pass the updated test suite.

a. Run `cd src/chronocone_planning_language && go test ./src/main/go/cli/compiler/... -count=1 -timeout 300s` — Completed

b. Confirm all 11 converted examples pass — Confirmed. All 11 examples pass `TestWithDataExamples_CompileOnly` and `TestWithDataExamples`.

c. Document any unexpected failures (example name, failure reason) but do not attempt fixes — 6 pre-existing failures documented below:
   - `conditional_when_default`: Parse error — missing INDENT (pre-existing CPL syntax issue)
   - `conditional_if_else_arithmetic`: Analysis error — duplicate measure definition "Daily" (pre-existing)
   - `conditional_logical_operators`: Missing `expected_output.xlsx` (Story239 territory)
   - `conditional_nested_arithmetic`: Missing `expected_output.xlsx` (Story239 territory)
   - `conditional_truthiness`: Missing `expected_output.xlsx` (Story239 territory)
   - `monthly_data`: Missing `expected_output.xlsx` (conditional_grain, Story239 territory)

None of the 6 failing tests are part of the 11 examples converted in this story.

**Acceptance Criteria**:
- `go test ./src/main/go/cli/compiler/...` passes for all 11 converted examples.
- Any failures are documented with example name and reason.

## Constraints

- `simple_addition` already has a `.xlsx` file — it must be regenerated to ensure it reflects current `data_*.json`.
- Pre-existing data discrepancies are NOT to be fixed in this story; they are tracked separately.
- The PDD script must be invoked from the project root with correct parameters.

## Notes

- These 11 examples correspond to Phase 3a of the test-conversion-plan.
- All examples in this phase are currently passing tests, so failures after conversion indicate script or data issues rather than pre-existing defects.
