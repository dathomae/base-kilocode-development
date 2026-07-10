# Story220: Tweak Expected Value Output Format

## Goal

Restructure the `expected_output` ODS sheet produced by the expected ODS generator to eliminate the branch/final row distinction and replace it with a single-row-per-metric format. The new format adds Computed Value and Match formula columns, makes input columns reference input sheets via ODF formulas, and removes CPL File, Branch, and Flag columns entirely. This enables manual spreadsheet inspection of metric correctness while simplifying the verifier to read all rows unconditionally.

## References

- [Plan: Tweak Expected Output ODS Format](../plans/tweak_expected_value_output_format_plan.md)
- [Story219: ODS Generator and Verifier](finished-stories/Story219_ods-generator-and-verifier.md) — foundational ODS generation and verification infrastructure

## Dependencies

- [Story219_ods-generator-and-verifier.md](finished-stories/Story219_ods-generator-and-verifier.md) — Completed. This story builds on the ODS generator and verifier that Story219 created.

## Dependent Stories

- [Story218_transition-to-ods-verification.md](Story218_transition-to-ods-verification.md) — Not Started. This story blocks Story218 because Story218 depends on the final single-row output format to generate and verify all example ODS files.

## Tasks

### Task Sizing

- **Tasks 1, 2, and 4** are `[Small]`: each touches at most 2 files, follows established patterns, and involves straightforward mechanical changes (field removal, mirrored function addition, or conditional guard removal).
- **Task 3** is not small: it rewrites the central generation loop in `generator.go`, restructures header layout, changes input column semantics from static values to formulas, adds two new formula columns, and updates `makeOutputRow` signatures. This requires simultaneous reasoning about CIL tree structure, formula generation, and cell reference mapping.
- **Task 5** is not small: it is a cross-cutting integration validation gate that exercises the full pipeline (ODS regeneration, test suites, verifier invocations) and may require diagnosing failures across multiple components.

### Worktree Suitability

- **Group A (Tasks 1, 2):** — fully disjoint packages (`ods_utils`/`main` vs `expected_ods_generator`), no shared infrastructure, no subprocess spawning, safe for subagents
- **Group B (Tasks 3, 4):** — fully disjoint packages (`expected_ods_generator` vs `main` in `ods_verify/`), no shared infrastructure, no subprocess spawning, safe for subagents

### Parallel Execution

- **Group A** (Tasks 1, 2): Task 1 modifies `ods_utils/ods_utils.go` and `expected_ods_generator/main/main.go`; Task 2 modifies `expected_ods_generator/cil_to_odf.go` and `cil_to_odf_test.go`. No file overlap, no compile-time dependencies, no shared infrastructure.
- **Group B** (Tasks 3, 4): Task 3 modifies `expected_ods_generator/generator.go`; Task 4 modifies `ods_verify/main.go` and `ods_verify/main_test.go`. No file overlap, no compile-time dependencies. Both depend on Group A (Task 3 needs `ExpressionToCellODF` from Task 2; Task 4 needs `SchemaMeta` without `FlagColumn` from Task 1), but they have no mutual dependency.

### Execution Order

```
[Task 1 ‖ Task 2] → [Task 3 ‖ Task 4] → Task 5
```

### Worktree Reintegration

None — no worktree groups.

### Task 1: Remove FlagColumn from SchemaMeta and update schema writer — Completed [Small]

**Intent:** Eliminate the `FlagColumn` field from the shared `SchemaMeta` struct and update the schema writer to produce the new column index layout without it.

1. Remove `FlagColumn int` from the `SchemaMeta` struct in `src/chronocone_planning_language/src/main/go/cli/ods_utils/ods_utils.go`.
2. Update `SchemaMeta` instantiation in `src/chronocone_planning_language/src/main/go/cli/expected_ods_generator/main/main.go`: remove `FlagColumn` assignment, update `MetricColumn` and `ValueColumn` indices to reflect the new layout (Metric immediately followed by Value, Computed Value, Match).
3. Write or update tests that verify `SchemaMeta` serializes/deserializes correctly without `flag_column`.

**Acceptance Criteria:**
- `ods_utils.SchemaMeta` no longer contains a `FlagColumn` field.
- `expected_ods_generator/main/main.go` writes a schema JSON with `MetricColumn` and `ValueColumn` pointing to the correct columns in the new single-row layout.
- Unit tests for schema JSON parsing pass without referencing `flag_column`.

### Task 2: Add ExpressionToCellODF for same-sheet cell references — Completed [Small]

**Intent:** Provide a function that translates CIL expression trees into ODF formulas using same-sheet cell references (e.g., `E5+F5`), so the Computed Value column can evaluate the metric expression directly in the spreadsheet.

1. Add `ExpressionToCellODF(node cil.CILNode, inputColumnIndexes map[string]int, colLetter func(int) string, row int, keyColCount int) string` to `src/chronocone_planning_language/src/main/go/cli/expected_ods_generator/cil_to_odf.go`.
2. Implement by reusing or mirroring the existing `ExpressionToODF` internal walker, but translating `CILInputRef` nodes into cell references like `${colLetter(keyColCount + inputColumnIndexes[name])}${row}` instead of sheet bracket references like `[$data_a.C3]`.
3. All other node types (binary ops, comparisons, logical ops, conditionals, when clauses, literals) translate identically to `ExpressionToODF`.
4. Add comprehensive tests in `cil_to_odf_test.go` covering:
   - `CILLiteral` → literal string (unchanged)
   - `CILInputRef` → cell reference (e.g., `E5` given inputColumnIndexes["data_a"]=0, row=5, keyColCount=2)
   - `CILBinaryOp` with cell refs (e.g., `(E5+F5)`)
   - `CILConditional` with cell refs (e.g., `IF(E5=10,E5+F5,IF(E5<>5,E5*2,E5-F5))`)
   - `CILWhenClause` with cell refs
   - Nested expressions with cell refs

**Acceptance Criteria:**
- `ExpressionToCellODF` is implemented and exported.
- All test cases for `ExpressionToCellODF` pass.
- Existing `ExpressionToODF` and `FormulaForBranch` tests remain passing.

### Task 3: Restructure generator to single-row format — Completed

**Intent:** Rewrite the ODS generation logic to produce one row per (metric, dimension grain combo) instead of multiple branch rows plus a final row, with input columns as formulas, and add Computed Value and Match formula columns.

1. Update the header row in `src/chronocone_planning_language/src/main/go/cli/expected_ods_generator/generator.go`: remove CPL File, Branch, and Flag columns; add Computed Value and Match columns.
2. Build an `inputColumnIndexes` map that maps each input variable name to its 0-based index within the input columns block.
3. For each (metric, combo), produce exactly one row containing:
   - Key columns (unchanged)
   - Input columns: ODF formulas referencing input sheets using existing `[$SheetName.ColRow]` bracket notation
   - Metric column (unchanged)
   - Value column: static float computed by the Go evaluator (`strconv.FormatFloat`)
   - Computed Value column: ODF formula from `ExpressionToCellODF`
   - Match column: ODF formula `IF(ABS(<valueCell>-<computedCell>)<1e-9,TRUE(),FALSE())`
4. Update `makeOutputRow` signature: remove `cplFile`, `branch`, `flag` parameters; add `computedValueFormula` and `matchFormula` parameters.
5. Remove the per-branch row generation loop entirely (the `for bIdx, branch := range conditional.Branches` loop).
6. Ensure input column formulas correctly reference the input sheet row using `findRowIndex` + header offset.

**Acceptance Criteria:**
- Generated ODS has exactly one data row per (metric, combo).
- Input columns contain ODF formulas referencing input sheets, not static floats.
- Value column contains the static float result from the Go evaluator.
- Computed Value column contains an ODF formula evaluating the metric expression with cell references.
- Match column contains an ODF formula that returns TRUE when Value ≈ Computed Value.
- No CPL File, Branch, or Flag columns remain in the output.
- Existing expected_ods_generator tests pass.

### Task 4: Update verifier to read all rows and update verifier tests — Completed [Small]

**Intent:** Remove the verifier's dependency on the `Flag="final"` filter so it reads all data rows unconditionally, aligning with the single-row format.

1. In `src/chronocone_planning_language/src/main/go/cli/ods_verify/main.go`, remove the `flagCol` variable and the `if row.TableCell[flagCol].P != "final" { continue }` filtering logic.
2. The verifier should read every row after the header (row 0) and construct keys from `meta.KeyColumns`, `meta.MetricColumn`, and `meta.ValueColumn`.
3. In `src/chronocone_planning_language/src/main/go/cli/ods_verify/main_test.go`, remove all `flag_column` fields from schema JSON test fixtures and update any assertions that depend on them.
4. Verify that `SchemaMeta` unmarshals correctly without the `flag_column` field.

**Acceptance Criteria:**
- `ods_verify/main.go` no longer references `FlagColumn` or filters rows by flag.
- The verifier reads all data rows in the ODS sheet.
- `ods_verify/main_test.go` passes with updated fixtures.
- Existing verifier comparison and parquet logic remains unchanged.

### Task 5: End-to-end validation and script review — Completed

**Intent:** Validate the complete pipeline by regenerating example ODS files, running all tests, and executing the verifier against generated outputs.

1. Regenerate example ODS files:
   ```bash
   go run src/main/go/cli/expected_ods_generator/main.go -example=examples/with_data/metrics_arithmetic/simple_addition
   go run src/main/go/cli/expected_ods_generator/main.go -example=examples/with_data/metrics_conditionals/conditional_comparison_operators
   ```
2. Run full test suites:
   ```bash
   go test ./src/main/go/cli/ods_verify/... -v
   go test ./src/main/go/cli/expected_ods_generator/... -v
   ```
3. Run verifier against generated ODS files:
   ```bash
   # simple_addition
   go run src/main/go/cli/ods_verify/main.go -parquet=output.parquet -ods=examples/with_data/metrics_arithmetic/simple_addition/expected_output.ods
   # conditional_comparison_operators
   go run generated/cmd/conditional_comparison_operators/main.go -data-dir=examples/with_data/metrics_conditionals/conditional_comparison_operators/
   go run src/main/go/cli/ods_verify/main.go -parquet=output.parquet -ods=examples/with_data/metrics_conditionals/conditional_comparison_operators/expected_output.ods
   ```
4. Review `src/chronocone_planning_language/shell_scripts/compile_and_run.sh` — confirm verifier invocation logic is unaffected.
5. Ensure both verifier invocations exit 0.

**Acceptance Criteria:**
- Both example ODS files regenerate successfully.
- All `ods_verify` and `expected_ods_generator` tests pass.
- Both verifier invocations against generated ODS files exit 0.
- No regressions in `compile_and_run.sh`.

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- ODF formula separator: existing code uses `,` as the function argument separator (e.g., `AND(a,b)`). Continue using `,` for Calc en-US locale compatibility.
- The Match formula uses `IF(ABS(<valueCell>-<computedCell>)<1e-9,TRUE(),FALSE())`.
- Do not modify `compile_and_run.sh` unless the verifier invocation logic is actually broken.
- Keep `InputColumns` in `SchemaMeta` (deferred cleanup beyond this story).
- `FormulaForBranch` in `cil_to_odf.go` must remain; do not delete existing exported functions.

## Acceptance Criteria

- The `expected_output` ODS file contains exactly one row per (metric, dimension grain combo).
- The `expected_output` ODS file contains no CPL File, Branch, or Flag columns.
- Input columns in `expected_output` reference input sheet cells via ODF formulas.
- The Computed Value column contains an ODF formula that evaluates the metric expression.
- The Match column contains an ODF formula that returns TRUE when Value ≈ Computed Value.
- The verifier reads all rows without flag filtering.
- All unit tests in both `expected_ods_generator` and `ods_verify` packages pass.
- End-to-end verifier runs against regenerated example ODS files exit 0.

## Requesting Clarification

If at any point during story construction there is confusion or ambiguity about the goal or how to accomplish it, stop and ask the user for clarification.
