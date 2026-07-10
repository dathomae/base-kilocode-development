# Plan: Add Lookup Data Sheets to Expected XLSX Generator

## Objective

Fix the `expected_xls_generator` so that `CILLookupNode` expressions produce valid Excel formulas in the "Computed Value" column (column G in the `expected_output` sheet). Currently, `CILLookupNode` translates to an empty string in the formula path, producing broken formulas like `(D2*)` that lack a right-hand operand.

## Background

The CPL metric expression `product_strength = raw_strength * strength_adjusted_for_material` compiles to a CIL binary op tree: `Multiply(CILInputRef("raw_strength"), CILLookupNode("strength_adjusted_for_material"))`.

- **Evaluator path** (`cil_evaluator.go`): Works correctly — evaluates the lookup in-process by matching dimension values against the in-memory lookup data. The "Value" column (F) is correct.
- **Formula path** (`cil_to_xlsx.go`): Broken — `translateExprCellXLSX` and `translateExprXLSX` both return `""` for `CILLookupNode`, producing an invalid formula in the "Computed Value" column (G).

## Strategy

Write each lookup's data as a separate Excel sheet (analogous to how connector data gets its own sheet). Then, for `CILLookupNode` formula generation, emit a `VLOOKUP` (or `INDEX(MATCH(...))`) that references the lookup sheet, matching on the dimension value(s) on the current row.

This mirrors the existing pattern: connector data → its own sheet → cross-sheet cell references. Lookup data → its own sheet → cross-sheet `VLOOKUP` formulas.

## Files Involved (under `src/chronocone_planning_language/src/main/go/`)

| File | Role | Change Type |
|------|------|-------------|
| `cli/expected_xls_generator/generator.go` | Orchestrates XLSX generation; creates sheets, writes rows | Modify |
| `cli/expected_xls_generator/cil_to_xlsx.go` | Translates CIL nodes to Excel formula strings | Modify |
| `cli/expected_xls_generator/cil_evaluator.go` | Evaluates CIL trees (in-process, for Value column) | No change (correct as-is) |
| `cli/expected_xls_generator/xlsx_utils.go` | `ColName`, `CellRef` helpers | No change |
| `cli/expected_xls_generator/main/main.go` | CLI entry point; loads data and calls generator | No change (already loads `lookupData`) |

## Detailed Implementation Steps

### Step 1: Create lookup data sheets in `generator.go`

In `GenerateXLSX`, after the existing Step 1 loop that creates connector sheets, add a new loop that creates one sheet per lookup.

**Details:**
- Iterate over `iospec.Lookups` (the `[]LookupSpecification` from the iospec).
- For each lookup, create a new sheet named after the lookup name (e.g., `strength_adjusted_for_material`).
- Populate it using the existing helper `populateInputSheetXLSX` (already works with `[]map[string]interface{}` — the lookup data format is identical: dimension columns + `Value`).
- The `lookupData` map is already keyed by lookup name and passed to `GenerateXLSX` as the 4th parameter.

### Step 2: Update formula generation for `CILLookupNode` in `cil_to_xlsx.go`

Modify both formula translation paths (`translateExprXLSX` and `translateExprCellXLSX`) to handle `CILLookupNode` by emitting a lookup formula rather than an empty string.

**Approach: Use VLOOKUP with concatenated key**

For a lookup with dimensions `[Material]`, the formula for row N should look like:
```
VLOOKUP(B2,strength_adjusted_for_material!B:C,2,FALSE)
```
Where:
- `B2` is the cell containing the "Material" value on the current row (the dimension column in `expected_output`).
- `strength_adjusted_for_material!B:C` is the lookup sheet's data range (columns B and C — "Material" and "Value").
- `2` is the column index for the "Value" column.
- `FALSE` for exact match.

For multi-dimension lookups (e.g., `dimensions[Country, Product]`), the formula must match on all dimensions. The simplest reliable approach is a helper column: add a concatenated key column to the lookup sheet (column A), then VLOOKUP against a similarly concatenated key on the expected_output row.

**Implementation approach for multi-dimension support:**
- When populating a lookup sheet, prepend a "key" column that concatenates all dimension values (e.g., `Country&"|"&Product`).
- In the formula for `CILLookupNode`, compute the same concatenation from the current row's dimension cells and VLOOKUP against the key column.

**Design consideration:** Both formula translation functions currently don't receive dimension-to-column mapping. They need to know which columns hold the dimension values for the current row in the `expected_output` sheet. The `ExpressionToCellExcel` function signature or the context passed through the recursive calls needs to include a way to resolve dimension concept names to Excel column references.

This could be done by either:
1. Passing a `dimensionColumnIndexes map[string]int` (dimension name → 0-based column index in expected_output) through the translation functions, or
2. Passing a function `dimensionToCell func(dimName string, row int) string` that produces the cell reference for a given dimension on the current row.

Option 2 is more natural as a callback. Since the existing `colName func(int) string` parameter already provides column naming, a complementary `dimCell func(string) string` that wraps dimension name → cell reference would fit the existing pattern.

**For single-dimension lookups:** `VLOOKUP(<dim_cell>,<lookup_sheet>!<dim_col>:<val_col>,2,FALSE)`

**For multi-dimension lookups:** `VLOOKUP(<concat_cells>,<lookup_sheet>!A:<val_col>,<val_col_idx>,FALSE)`

**Initial scope:** Handle the single-dimension case first (matches the existing `lookups_basic` example). Multi-dimension support can be added as a follow-up if needed, or can be supported from the start using the concatenated-key approach.

### Step 3: Update `GenerateXLSX` to pass dimension context to formula generation

When calling `ExpressionToCellExcel` (generator.go line 164), the generator currently passes `inputColumnIndexes`, `ColName`, `currentRow`, and `keyColCount`. It needs to also provide dimension-to-column mapping.

**Details:**
- The header row already writes key columns (dimensions + grain) starting at column 0. So dimension "Product" is column 0, "Material" is column 1, etc.
- Build a `dimensionColumnIndexes map[string]int` from `keyColumns` during the header-writing loop.
- Thread this through to `ExpressionToCellExcel` (either as a new parameter or as part of the translation context).

Since both `ExpressionToExcel` and `ExpressionToCellExcel` are public API functions (and may be called by external code), adding a new parameter would be a breaking change. Instead, consider one of:
- Add a new function `ExpressionToCellExcelWithLookups` that accepts the additional context, and have the generator call that.
- Pass the lookup sheet name and dimension column info as part of a struct context.
- Add the dimension context as an additional parameter (accepting the minor API change since these are internal CLI tools).

### Step 4: Handle the `translateExprXLSX` cross-sheet lookup case

For `translateExprXLSX` (used for cross-sheet references in input column formulas and branch formulas), `CILLookupNode` also returns `""`. This code path is used by `FormulaForBranchXLSX` (line 253). If a conditional branch body contains a lookup reference, this path would also be broken.

For this path, the formula needs to reference the lookup sheet from another sheet using the full cross-sheet VLOOKUP syntax:
```
VLOOKUP(expected_output!B2,strength_adjusted_for_material!B:C,2,FALSE)
```

### Step 5: Write/update tests

**New test cases needed:**
- `TestExpressionToCellExcel_LookupNode_SingleDim` — Verifies that a single-dimension `CILLookupNode` produces a correct `VLOOKUP` formula.
- `TestExpressionToCellExcel_LookupNode_MultiDim` — Verifies that a multi-dimension `CILLookupNode` produces a correct multi-key `VLOOKUP` formula.
- `TestExpressionToCellExcel_LookupNode_BinaryOp` — Verifies a binary op containing a lookup node (e.g., `CILInputRef * CILLookupNode`) produces a correct formula like `(C2*VLOOKUP(B2,...))`.
- `TestExpressionToExcel_LookupNode_SingleDim` — Verifies the cross-sheet case.
- `TestGenerateXLSX_LookupSheets` — Verifies that lookup data sheets are created in the generated XLSX.
- `TestGenerateXLSX_LookupFormulaInOutput` — Verifies that the `expected_output` sheet's "Computed Value" column contains valid formulas with `VLOOKUP`.
- Update `TestGenerateXLSX_Header` if the layout changes (should not).

**Files:** Add tests to `cil_to_xlsx_test.go` and `generator_test.go`.

### Step 6: Regenerate the `lookups_basic` expected output

After the fix, re-run the expected XLS generator on the `lookups_basic` example to regenerate `expected_output.xlsx` and verify the formula in column G is now valid.

### Step 7: Verify end-to-end

Run the `compile_and_run.sh` script on `lookups_basic` to verify that `xls_verify` passes (i.e., the Excel `Computed Value` column now matches the `Value` column after formula evaluation).

## Potential Complications

1. **Excel formula limitations with VLOOKUP on concatenated keys:** `VLOOKUP` expects the lookup value in the first column of the range. The concatenated-key approach satisfies this by putting the key in column A.

2. **Lookup data file naming:** The `main.go` CLI already loads lookup JSON files as `<lookupName>.json`. The sheet naming should match this: use `lookupSpec.LookupName` as the sheet name.

3. **Lookup sheets need to be distinct from connector sheets:** If a lookup shares a name with a connector, sheet creation will fail. This is unlikely in practice (CPL namespaces enforce uniqueness) but worth noting.

4. **Formula escaping:** Excel sheet names with special characters need single-quote escaping (e.g., `'Sheet Name'!A1`). The `excelize` library may handle this. If lookup names contain special characters, we should sanitize or quote the sheet reference.

5. **Performance impact:** Each lookup row in the expected_output sheet gets a VLOOKUP formula. For hundreds of rows this is fine; VLOOKUP is a standard Excel operation.
