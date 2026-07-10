# Story280: Add Lookup Expectations

## Goal

Fix the `expected_xls_generator` so that `CILLookupNode` expressions produce valid Excel formulas in the "Computed Value" column of `expected_output` sheets. Currently, `CILLookupNode` translates to an empty string in the formula path, producing broken formulas like `(D2*)` that lack a right-hand operand. This story adds lookup data as dedicated Excel sheets and generates VLOOKUP formulas that reference those sheets, matching on dimension values from the current output row.

## References

- [Plan: Add Lookup Data Sheets to Expected XLSX Generator](../plans/add-lookup-expectations-plan.md)
- [Project Brief](../../src/chronocone_planning_language/src/main/go/brief.md)

## Dependencies

None. This story is self-contained within the `expected_xls_generator` package and does not require any other story to complete first.

## Dependent Stories

None. No stories are currently blocked by this work.

## Tasks

### Task Sizing

Task 1 is additive (one source file plus its test), follows an existing sheet-creation pattern, and addresses a single concern — it qualifies as `[Small]`. Task 2 touches three files with coordinated signature changes, a new CILNode case, and VLOOKUP generation logic — it is not small. Task 3 adds a cross-sheet CILLookupNode case to `translateExprXLSX` following the pattern established in Task 2; it touches two files (single function change plus test) and is `[Small]`. Task 4 adds an integration test in one file following an existing test pattern — it is `[Small]`.

The architect-for-story-planning's decomposition review found that all tasks are appropriately sized. Task 2 is borderline (7 subtasks, 3 files with coordinated parameter and call-site changes) but has no natural seam for decomposition: the signature change, CILLookupNode case, and call-site update form a tightly coupled unit that cannot compile independently if split.

### Parallel Execution

- **Group A** (Tasks 1, 3): Task 1 modifies `generator.go` and `generator_test.go`; Task 3 modifies `cil_to_xlsx.go` and `cil_to_xlsx_test.go`. The file sets are completely disjoint. Neither task depends on the other's output — Task 1 creates lookup sheets, Task 3 generates formula strings that reference sheet names without needing the sheets to exist. Both tasks are independently testable.

### Execution Order

```
[Task 1 ‖ Task 3] → Task 2 → Task 4
```

Tasks 1 and 3 run in parallel (Group A). Task 2 depends on both: it shares `generator.go` with Task 1 (building on the lookup sheet loop) and shares `cil_to_xlsx.go` with Task 3 (integrating the cross-sheet infrastructure with the same-sheet changes). Task 4's integration test requires both lookup sheets (Task 1) and formula generation (Task 2).

### Reintegration

**Group A (Tasks 1, 3):**
- Merge order: any order (fully disjoint files, no shared infrastructure)
- Integration tests: `go test ./src/main/go/cli/expected_xls_generator/... -count=1 -timeout 60s` — verify both lookup sheet creation and cross-sheet formula tests pass in the merged tree before Task 2 begins
- Watch for conflicts in: none expected (generator.go and cil_to_xlsx.go are disjoint)

### Task 1: Create lookup data sheets in GenerateXLSX — Completed [Small]

**Intent:** Add dedicated Excel sheets for each lookup defined in the IO specification, populated with the lookup's dimension and value data, so that downstream formula generation can reference them via VLOOKUP.

**Files:** `src/main/go/cli/expected_xls_generator/generator.go`, `src/main/go/cli/expected_xls_generator/generator_test.go`

**Acceptance Criteria:**
1. `TestGenerateXLSX_LookupSheets` verifies that for a defined lookup, a sheet with the lookup name exists in the generated XLSX and contains the correct header and data rows.

a. Write `TestGenerateXLSX_LookupSheets`: construct defs/iospec with a lookup specification in `iospec.Lookups`, provide matching `lookupData`, call `GenerateXLSX`, verify the sheet exists via `f.GetSheetList()`, verify header row includes dimension columns and a `Value` column, verify data rows match lookup values — Not Started
b. Add a loop in `GenerateXLSX` after the connector sheet loop (after line 37) that iterates over `iospec.Lookups`, creates a new sheet named after `lookupSpec.LookupName`, and populates it via `populateInputSheetXLSX(f, sheetName, lookupData[lookupSpec.LookupName])` — Not Started
c. For multi-dimension lookups (more than one dimension in the lookup data rows): prepend a `key` column (column A) to each row before calling `populateInputSheetXLSX`. The key value is the pipe-separated concatenation of dimension values (e.g., `Country|Product`). Construct the key by: building a temporary `[]map[string]interface{}` with a `key` field prepended, computed by concatenating the lookup's dimension values with `|`. Then call `populateInputSheetXLSX` with the augmented rows — Not Started
d. Run `go test ./src/main/go/cli/expected_xls_generator/... -run TestGenerateXLSX_LookupSheets -count=1` and verify it passes — Not Started

### Task 2: Implement VLOOKUP formula generation for CILLookupNode (same-sheet path) — Completed

**Intent:** Add a `CILLookupNode` case to `translateExprCellXLSX` that emits a valid VLOOKUP formula referencing a lookup data sheet, matching the current row's dimension cells. Thread dimension-to-column mapping through the formula translation functions so the VLOOKUP can reference the correct dimension cells on the expected_output sheet.

**Files:** `src/main/go/cli/expected_xls_generator/cil_to_xlsx.go`, `src/main/go/cli/expected_xls_generator/generator.go`, `src/main/go/cli/expected_xls_generator/cil_to_xlsx_test.go`

**Design decisions captured from the plan:**
- Pass a `dimensionColumnIndexes map[string]int` (dimension name → 0-based column index in `expected_output`) through `ExpressionToCellExcel` → `translateExprCellXLSX`.
- For single-dimension lookups: `VLOOKUP(<dim_cell>,<lookup_sheet>!<dim_col>:<val_col>,2,FALSE)` where `<dim_cell>` is the dimension value cell on the current output row (e.g., `B2`) and `<lookup_sheet>` is the lookup name.
- For multi-dimension lookups: `VLOOKUP(<concat_cells>,<lookup_sheet>!A:<val_col>,<val_col_idx>,FALSE)` where `<concat_cells>` concatenates all dimension cells with `&"|"&` (same separator as the key column in Task 1c), and column A holds the concatenated key.

**Acceptance Criteria:**
1. `TestExpressionToCellExcel_LookupNode_SingleDim` passes — single-dimension lookup produces correct VLOOKUP formula.
2. `TestExpressionToCellExcel_LookupNode_MultiDim` passes — multi-dimension lookup produces concatenated-key VLOOKUP formula.
3. `TestExpressionToCellExcel_LookupNode_BinaryOp` passes — binary operation containing a LookupNode produces formula with VLOOKUP as an operand.
4. All existing `cil_to_xlsx_test.go` tests and `generator_test.go` tests continue to pass (no regressions).

a. Write `TestExpressionToCellExcel_LookupNode_SingleDim` in `cil_to_xlsx_test.go`: create a `CILLookupNode` with `LookupName: "strength_adjusted_for_material"` and `Dimensions: []string{"Material"}`, construct `dimensionColumnIndexes` mapping `"Material"` to column index 1, set `inputColumnIndexes` and `keyColCount` to match a layout where Material is a key column. Call `ExpressionToCellExcel`. Assert the result matches the expected VLOOKUP formula (e.g., `VLOOKUP(B2,strength_adjusted_for_material!B:C,2,FALSE)` where `B` corresponds to the Material column). Verify the test fails (CILLookupNode currently returns `""`) — Not Started
b. Write `TestExpressionToCellExcel_LookupNode_MultiDim` in `cil_to_xlsx_test.go`: create a `CILLookupNode` with `Dimensions: []string{"Country", "Product"}` and `LookupName: "country_product_index"`. Set `dimensionColumnIndexes` mapping `"Country"` to column 0 and `"Product"` to column 1. Assert the result matches `VLOOKUP(A2&"|"&B2,country_product_index!A:C,3,FALSE)` (key in column A, Value in column C since there are two dimension columns before it). Verify the test fails — Not Started
c. Write `TestExpressionToCellExcel_LookupNode_BinaryOp` in `cil_to_xlsx_test.go`: create `CILBinaryOp{Op: operator.Multiply, Left: CILInputRef{Name: "raw_strength"}, Right: CILLookupNode{...}}`. Set up `inputColumnIndexes` and `dimensionColumnIndexes` so that the input ref resolves to column C and the lookup dimension resolves to column B (both relative to key column start). Assert result is `(C2*VLOOKUP(B2,strength_adjusted_for_material!B:C,2,FALSE))`. Verify the test fails — Not Started
d. Add a `dimensionColumnIndexes map[string]int` parameter to `ExpressionToCellExcel` and thread it through to `translateExprCellXLSX`. Update the `translateExprCellXLSX` signature and all recursive call sites (both conditional and when-clause helper functions) to accept and forward this parameter. All recursive calls: `translateExprCellXLSX`, `translateConditionalCellXLSX`, `buildIfChainCellXLSX`, `translateWhenClauseCellXLSX` — Not Started
e. Add the `*cil.CILLookupNode` case in `translateExprCellXLSX` (after the `*cil.CILNilLiteral` case at line 169):
   - Retrieve `lookupSheet := n.LookupName`
   - For single-dimension lookups (`len(n.Dimensions) == 1`): get the dimension column index via `dimensionColumnIndexes[n.Dimensions[0]]`, build `dimCell := colName(dimIndex) + strconv.Itoa(row)`, compute `valCol := dimIndex + 2` (the Value column is two columns right of the dimension column in the lookup sheet: key column A → dim in B → Value in C for single-dim; for multi-dim: key A → dim1 B → dim2 C → Value D). Build the formula: `fmt.Sprintf("VLOOKUP(%s,%s!%s:%s,2,FALSE)", dimCell, lookupSheet, colName(dimIndex), colName(valCol))`. Adjust `valCol` computation based on whether a key column exists (multi-dim).
   - For multi-dimension lookups: build a concatenation expression by iterating over `n.Dimensions`, building each dimension cell reference via `colName(dimensionColumnIndexes[dim]) + strconv.Itoa(row)`, joining with `&\"|\"&`. Determine `valColIdx := len(n.Dimensions) + 1` (key column + N dimension columns → Value is at 1-indexed position `len(n.Dimensions) + 1`; for VLOOKUP 3rd argument, use that as the column index). The data range is `A:<valCol>` where `valCol` uses `colName(len(n.Dimensions))` (0-based: key is 0, dims 1..N, Value is N). Formula: `VLOOKUP(<concat>,<lookupSheet>!A:<valCol>,<valColIdx>,FALSE)` — Not Started
f. Update the call site in `GenerateXLSX` at line 164 (the `ExpressionToCellExcel` call): build `dimensionColumnIndexes` from `keyColumns` during the header-writing loop (lines 102-106). Add a `dimensionColumnIndexes := make(map[string]int)` before the header loop, and inside the loop, after line 105, add `dimensionColumnIndexes[kc] = col`. Pass this map as the new parameter to `ExpressionToCellExcel` — Not Started
g. Run `go test ./src/main/go/cli/expected_xls_generator/... -count=1` and verify all tests pass (both new lookup tests and all existing tests) — Not Started

### Task 3: Implement VLOOKUP formula generation for CILLookupNode (cross-sheet path) — Completed [Small]

**Intent:** Handle the `CILLookupNode` case in `translateExprXLSX` so that cross-sheet formula contexts (e.g., conditional branch activation formulas referencing lookup values) produce valid cross-sheet VLOOKUP formulas.

**Files:** `src/main/go/cli/expected_xls_generator/cil_to_xlsx.go`, `src/main/go/cli/expected_xls_generator/cil_to_xlsx_test.go`

**Acceptance Criteria:**
1. `TestExpressionToExcel_LookupNode_SingleDim` passes — cross-sheet VLOOKUP formula references the expected_output sheet for dimension values.
2. All existing tests in `cil_to_xlsx_test.go` continue to pass.

a. Write `TestExpressionToExcel_LookupNode_SingleDim` in `cil_to_xlsx_test.go`: create a `CILLookupNode` with a single dimension, call `ExpressionToExcel` (which goes through `translateExprXLSX`), verify the result is a cross-sheet VLOOKUP like `VLOOKUP(B2,strength_adjusted_for_material!B:C,2,FALSE)`. Because this path does not receive dimension context yet, the test will initially fail. For the test, update the `ExpressionToExcel` call signature when the parameter is added in subtask c — Not Started
b. Add the `*cil.CILLookupNode` case in `translateExprXLSX` (after the `*cil.CILNilLiteral` case at line 56): follow the same VLOOKUP pattern as Task 2e but reference dimension cells on the `expected_output` sheet. For single-dimension lookups: `VLOOKUP(expected_output!<dim_col><row>,<lookupSheet>!<dim_col>:<val_col>,2,FALSE)`. Note that `translateExprXLSX` currently receives `rowIdx` (0-indexed input sheet row), so the expected_output row is `rowIdx + 2` (since the caller in `ExpressionToExcel` is called from generator.go at line 160, which receives `colRowIdx` — the same row offset). Use `expectedOutputRow := rowIdx + 2` for the dimension cell row reference. For the dimension column, use `dimensionColumnIndexes[n.Dimensions[0]]` which is the column index (0-based) on expected_output — Not Started
c. Add the `dimensionColumnIndexes map[string]int` parameter to `ExpressionToExcel` and `translateExprXLSX` signatures. Update all recursive call sites: `translateExprXLSX`, `translateConditionalXLSX`, `buildIfChainXLSX`, `translateWhenClauseXLSX`, and `FormulaForBranchXLSX` — Not Started
d. Run `go test ./src/main/go/cli/expected_xls_generator/... -run "TestExpressionToExcel" -count=1` and verify the cross-sheet lookup test passes, and all existing cross-sheet tests still pass — Not Started

### Task 4: End-to-end lookup integration verification — Completed [Small]

**Intent:** Verify that the end-to-end pipeline — lookup sheet creation + VLOOKUP formula generation — produces a valid expected XLSX with working formulas in the Computed Value column.

**Files:** `src/main/go/cli/expected_xls_generator/generator_test.go`

**Acceptance Criteria:**
1. `TestGenerateXLSX_LookupFormulaInOutput` passes — generated XLSX's expected_output sheet has a valid VLOOKUP formula in the Computed Value column.
2. All existing tests in the package continue to pass.

a. Write `TestGenerateXLSX_LookupFormulaInOutput` in `generator_test.go`: construct defs with a metric whose CIL expression is `CILBinaryOp{Op: operator.Multiply, Left: CILInputRef{Name: "raw_strength"}, Right: CILLookupNode{LookupName: "strength_adjusted_for_material", Dimensions: []string{"Material"}}}`. Set up iospec with one connector, one output dimension (`"Material"`), one lookup named `"strength_adjusted_for_material"` with `"Material"` dimension. Provide input data (Material = "Steel", Value = raw_strength value) and lookup data (Material → strength_adjusted_for_material value). Call `GenerateXLSX`. Retrieve the Computed Value column formula from the data row. Verify the formula is non-empty, contains `VLOOKUP`, and does not contain unbalanced parentheses or invalid operands (e.g., no trailing `*` with nothing after it). A reasonable check: formula starts with `(` and contains `VLOOKUP` — Not Started
b. Run `go test ./src/main/go/cli/expected_xls_generator/... -run TestGenerateXLSX_LookupFormulaInOutput -count=1` and verify it passes — Not Started
c. Run `go test ./src/main/go/... -count=1 -timeout 120s` from `source_root` (`src/chronocone_planning_language`) and verify the full test suite passes with no new failures — Not Started

## Test-First Development

This story follows the Logical TDD Lifecycle. Each task's tests are written before the implementation code and verified to fail before the implementation is added. The test-and-implement cycle is integrated within each task; tests are not broken out as separate tasks.

## Constraints

- **Excel formula compatibility**: VLOOKUP formulas must use the standard Excel syntax with `FALSE` for exact match.
- **Multi-dimension key column**: The pipe `"|"` character is used as the dimension separator in concatenated keys. Both the lookup sheet key column (Task 1) and the VLOOKUP concatenation expression (Task 2) must use the same separator.
- **Backward compatibility**: Existing tests must continue to pass after all changes. The `TestGenerateXLSX_Header` test should not need modification (no layout change).
- **Column indexing**: All column indices are 0-based in Go code. `VLOOKUP`'s third argument (column index) is 1-based. The `colName` function converts 0-based to Excel letter.
- **File path convention**: Go source paths include the `src/main/go/` prefix. Test commands use fully-qualified package paths from `source_root` (`src/chronocone_planning_language`).

## Intent

This story fixes a defect where `CILLookupNode` expressions in CIL metric trees produce empty strings in the Excel formula generation path, resulting in broken Computed Value formulas. The fix follows the existing pattern: connector data gets its own sheet; lookup data should get its own sheet. VLOOKUP formulas then reference these lookup sheets with dimension-value matching, analogous to how input column cross-sheet references work. The evaluator path (`cil_evaluator.go`) already handles lookups correctly — this story brings the formula path to parity.

## Notes

- The `main/main.go` CLI already loads `lookupData` and passes it to `GenerateXLSX`. No changes needed there.
- Sheet names are derived from `lookupSpec.LookupName`. If a lookup name conflicts with a connector name, Excel sheet creation would fail. The CPL namespace system enforces uniqueness, making this collision unlikely in practice.
- The cross-sheet CILLookupNode case (Task 3) is forward-looking: current usage patterns do not place lookups in conditions or body expressions of conditionals that are handled by `FormulaForBranchXLSX`. The case is included for correctness and future-proofing.
- After this story is complete, the `lookups_basic` example's `expected_output.xlsx` should be regenerated and verified via `compile_and_run.sh` to confirm end-to-end correctness.
