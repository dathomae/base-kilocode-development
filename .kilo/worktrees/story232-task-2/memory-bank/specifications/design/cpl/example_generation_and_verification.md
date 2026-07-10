# Example Generation and Verification

## Overview

This document describes the design of the ODS-based example generation and verification
system for CPL `with_data` examples. It replaces the previous JSON-based expected-output
verification with LibreOffice Calc-compatible `.ods` spreadsheets that support both
human inspection (formula rows viewable in Calc) and automated verification (static
final-value rows compared against program output).

The system consists of two CLIs and a supporting workflow:

1. **ODS Generator** (`expected_ods_generator`): Produces `expected_output.ods` and
   `expected_output_schema.json` from example CPL files, IOSpec, and input data.
2. **ODS Verifier** (`ods_verify`): Reads `expected_output_schema.json` for column
   layout, then compares `expected_output.ods` final rows against `output.parquet`
   produced by the CPL synthesis pipeline.
3. **PDD Workflow** (`create_example.pdd.script.md`): Step-by-step script for creating
   new examples and regenerating existing ones.

ODS was chosen over CSV because: cross-sheet references work natively, a single `.ods`
file contains all sheets, and the format opens natively in LibreOffice Calc.

## Design Principles

1. **Reuse existing infrastructure**: The ODS generator uses the existing CPL parser
   and analysis pipeline (`parser.ParseFiles` + `analysis.Analyze`) rather than
   introducing a new text-based parser. This ensures consistency with the compiler
   and avoids divergence as CPL evolves.

2. **Dimension uniformity at verification level**: All dimensions — whether declared
   as grain (time-related) or concept (categorical) — are structurally identical at
   the verification level. Tools treat them as a single ordered list of named
   columns rather than bifurcating them into separate "dimensions" and "grain"
   categories. See the lesson **Grain dimensions and non-grain dimensions are
   structurally identical at the verification/example level**.

3. **Separate formula rows from static rows**: The fundamental limitation that a
   single ODS cell cannot hold both a working formula and a readable static value
   is solved with a two-row-per-result pattern: branch rows contain ODF formulas
   (human-verifiable in Calc), and final rows contain static numeric values
   (machine-verifiable).

4. **Template-based ODS construction**: The go-ods v1.0.7 `Write` function requires the
   original `*zip.ReadCloser` from a prior `Read` to preserve the ODF zip structure.
   The generator embeds a minimal empty ODS file (via Go `embed.FS`) as a template,
   reads it to obtain the struct and zip reader, then modifies the struct by adding
   sheets/rows/cells before writing the final output. This avoids the risk of
   `ods.Write(path, data, nil)` with a nil zip reader behaving unexpectedly.

5. **CIL tree as single source of truth**: Metrics, expressions, conditionals,
   and input references are all extracted from the CIL tree produced by the
   analysis phase. A CIL-to-text formatter produces human-readable branch labels.

6. **Schema metadata file (configuration by convention)**: The generator produces
   an `expected_output_schema.json` alongside the `.ods` file. This metadata file
   captures column layout (key columns, column indices) so the verifier reads it
   rather than introspecting the ODS header row. The file naming convention —
   `<output_stem>_schema.json` — eliminates the need for explicit CLI flags on the
   verifier.

---

## Example File Conventions

Each `with_data` example directory follows a standard layout. The generator and
verifier tools rely on these conventions rather than explicit configuration,
following a "configuration by convention" approach.

```
examples/with_data/<category>/<example>/
    ├── iospec.json                       ← IO specification
    ├── <example>.cpl                     ← CPL source (may declare connectors + metrics)
    ├── data_*.json                       ← input data (flat arrays)
    ├── data_*.parquet                    ← input parquet files
    ├── output.parquet                    ← synthesis output
    ├── expected_output.ods              ← ODS generator output (human verification)
    └── expected_output_schema.json      ← column layout metadata (machine verification)
```

The parent directory may contain a `core.cpl` file defining shared concepts, types,
units, and measures used across multiple examples in that category.

**Two files always produced together** by the ODS generator:
- `expected_output.ods` — human-verifiable spreadsheet
- `expected_output_schema.json` — column layout consumed by the verifier

The verifier auto-discovers `expected_output_schema.json` from the same directory
as the `.ods` file. No additional flags are needed.

---

### Inputs

| Input | Source | Purpose |
|-------|--------|---------|
| `-example=<dir>` | CLI flag | Path to example subdirectory (contains iospec.json, .cpl files, data_*.json) |
| `-output=<path>` | CLI flag | Path for the generated `.ods` file. Defaults to `<example>/expected_output.ods`. |
| `iospec.json` | Example directory | Input connectors, output metrics, dimensions, grain |
| `.cpl` files | Example directory + parent (core.cpl) | Metric definitions, expressions, conditionals |
| `data_*.json` | Example directory | Input values per (dimension, grain) combination |

### Input Data Mapping

Input data files follow the naming convention `data_<connector>.json` (e.g., `data_a.json`
for connector `input_a`). The generator discovers connectors from `iospec.json`, then
reads the corresponding `data_*.json` file for each connector. Each file is a flat JSON
array of objects with dimension keys, grain keys, and a `Value` field.

### Output

Two files produced per example:

1. **`expected_output.ods`**: An ODS file containing:
   - One **input sheet** per connector (named by connector name), with a header row and
     data rows.
   - One **`expected_output` sheet** with per-metric rows (branch formulas and final values).

2. **`expected_output_schema.json`**: A metadata file describing the column layout:
   ```json
   {
       "ods_file": "expected_output.ods",
       "key_columns": ["Region", "Daily"],
       "input_columns": ["input_a", "input_b"],
       "metric_column": 4,
       "flag_column": 7,
       "value_column": 8
   }
   ```
   - `key_columns`: The ordered list of dimension and grain column names (flattened).
   - `input_columns`: The ordered list of input value column names (sorted alphabetically).
   - `metric_column`: Index of the Metric column.
   - `flag_column`: Index of the Flag column (value is `"branch"` or `"final"`).
   - `value_column`: Index of the Value column.
   - The `cpl_file_column` and `branch_column` indices are not included — they are not
     needed for verification.
   - Column indices shift right by the number of input columns (len(input_columns)).

### Pipeline

```
             ┌──────────────────────────────────┐
             │ 0. Read template ODS (embed.FS)   │
             │    ods.Read(template) → data,files│
             ├──────────────────────────────────┤
             │ 1. Parse CPL files                │
             │    parser.ParseFiles(allPaths)    │ → symbol table
             ├──────────────────────────────────┤
             │ 2. Analyze symbol table           │
             │    analysis.Analyze(rootTable)    │ → defs.Metrics, defs.Connectors
             ├──────────────────────────────────┤
             │ 3. Load IO specification          │
             │    io.LoadIOSpecification(path)   │ → dimensions, grain, connectors
             ├──────────────────────────────────┤
             │ 4. Read input data                │
             │    data_*.json (flat arrays)      │ → input values per (dim, grain)
             ├──────────────────────────────────┤
             │ 5. Build ODS sheets               │
             │    a. Input sheet per connector   │
             │    b. expected_output sheet:      │
             │       - CIL→ODF formula translate │
             │       - CIL evaluator (final vals)│
             │       - CIL→text formatter (branch)│
             ├──────────────────────────────────┤
             │ 6. Write outputs                   │
             │    a. ods.Write(path, data, files) │
             │    b. Write schema JSON            │
             └──────────────────────────────────┘
```

### Sheet Structure

#### Input Sheets

One sheet per input connector, named by the connector name (e.g., `data_a`).

| Column 0 | ... | Column N-2 | Column N-1 |
|----------|-----|------------|------------|
| Dim[0] value | ... | Grain[M-1] value | Value (float) |

- Header row: dimension names (from iospec.json `dimensions`) followed by grain
  names (from iospec.json `grain`), then `Value`.
- Data rows: one row per (dimension, grain) combination from the connector's
  `data_*.json` file.
- All dimension and grain columns are treated uniformly as string columns.
- Value column type is `float`.

#### `expected_output` Sheet

| Cols | Content | Type |
|------|---------|------|
| 0..D-1 | Dimension values (from iospec dimensions, in order) | `string` |
| D..D+G-1 | Grain values (from iospec grain, in order) | `string` |
| D+G..D+G+I-1 | Input values (one column per input ref, sorted alphabetically) | `float` |
| D+G+I | Metric name | `string` |
| D+G+I+1 | CPL File (source `.cpl` file) | `string` |
| D+G+I+2 | Branch (conditional text, e.g., `"if data_a == 10"`) | `string` |
| D+G+I+3 | Flag: `"branch"` or `"final"` | `string` |
| D+G+I+4 | Value | `float` |

Where D = number of dimension columns, G = number of grain columns, and I = number of input value columns.

The input value columns are inserted between the key columns (dimensions + grain) and the Metric column.
Each input value column is named after the corresponding input reference (e.g., `input_a`, `data_b`)
and contains the actual numeric value of that input for the current (dimension, grain) combination.
Input names are collected from all metrics' CIL tree input references and sorted alphabetically.

- **Branch rows**: `Flag = "branch"`. The Value cell contains an ODF formula (with
  `of:=` prefix). The input value columns contain the static numeric values that
  the formula references, making manual verification easy — a reviewer can see both
  the input values and the formula in the same row.
- **Final rows**: `Flag = "final"`. The Value cell contains a static numeric value
  computed by the CIL evaluator. `ValueType = "float"`. Input value columns also
  contain the static input values.

#### Branch Label Generation

Branch text is generated from the CIL tree via a CIL→text formatter. Examples:

| CIL Structure | Branch Text |
|---------------|-------------|
| `CILBinaryOp{+, data_a, data_b}` (arithmetic metric) | `"data_a + data_b"` |
| `CILConditional` branch 0: `CILComparisonOp{==, data_a, CILLiteral{10}}` | `"if data_a == 10"` |
| `CILConditional` branch 1: `CILComparisonOp{!=, data_a, CILLiteral{5}}` | `"else if data_a != 5"` |
| `CILConditional` last branch (no condition) | `"else"` |
| `CILWhenClause` case 0: `CILComparisonOp{==, x, CILLiteral{1}}` | `"when x == 1"` |
| `CILWhenClause` default | `"default"` |

#### CPL File Column

Populated from the symbol table's `SourceLocation.File` field, tracing each metric's
definition back to the `.cpl` source file where it was declared. The path is reduced
to its basename for display (e.g., `core.cpl`, `conditional_comparison_operators.cpl`).
This supports future examples where core.cpl provides unit/measure conversions that
contribute calculation rows.

### CIL Evaluator

A tree-walking evaluator that computes final static numeric values from the CIL
tree. For each metric, the evaluator iterates over all (dimension, grain) combinations
from the output's Cartesian product (defined by the iospec dimensions and grain),
performing these steps for each combination:

1. Build a value lookup table for the current (dimension, grain) combination:
   keyed by connector name → `Value` field from `data_<connector>.json`.
2. Walk the CIL tree depth-first.
3. For `CILInputRef` nodes: substitute the actual value from the lookup table.
4. Evaluate arithmetic (`+`, `-`, `*`, `/`), unary negation, and comparisons
   (`==`, `!=`, `<`, `>`, `<=`, `>=`) with float64 semantics.
5. Evaluate logical operators (`and`, `or`, `not`) and `nil`-truthiness.
6. For `CILConditional`: evaluate each branch's condition in order; the first
   true branch's body value is returned. If no branch matches, returns 0.
7. For `CILWhenClause`: evaluate each case's condition; if matched, return body
   value. Otherwise return the default body value.

The evaluator produces one final value per (dimension, grain, metric) combination,
which is written as a final row in the `expected_output` sheet.

### ODF Formula Generation

A separate CIL→ODF translator produces ODF formula strings for branch rows. It
uses the CPL-to-ODF translation rules:

| CPL Construct | ODF Formula Equivalent |
|---------------|------------------------|
| `if <cond> <then> else <rest>` | `IF(<cond>, <then>, <rest>)` |
| `if <cond> <then>` (no else) | `IF(<cond>, <then>, "")` |
| `<expr> and <expr>` | `AND(<expr>, <expr>)` |
| `<expr> or <expr>` | `OR(<expr>, <expr>)` |
| `not <expr>` | `NOT(<expr>)` |
| `True` | `TRUE()` |
| `False` | `FALSE()` |
| `==` | `=` |
| `!=` | `<>` |
| `>=` / `<=` / `>` / `<` | `>=` / `<=` / `>` / `<` |
| `+` / `-` / `*` / `/` | `+` / `-` / `*` / `/` |
| `-<expr>` (unary negation) | `-<expr>` |
| `(<expr>)` | `(<expr>)` |
| `when <c1> => <r1> ... default => <d>` | `IF(<c1>, <r1>, IF(<c2>, <r2>, ..., <d>))` |
| Truthiness (`if data_a`) | `IF(<ref><>0, ...)` |
| Tautological (`if True`) | `IF(TRUE(), ...)` |

#### Per-Branch Activation Conditions

For branch N (1-indexed) in an if/else-if/else chain:
- Prior negation: `AND(NOT(cond_1), NOT(cond_2), ..., NOT(cond_{N-1}))`
- Branch activation: `AND(prior_negation, cond_N)` (for else: just `prior_negation`)
- Full formula: `IF(activation, result_expr, "")`

#### Cross-Sheet References

All cell references in formulas use ODF bracket notation `[$SheetName.ColRow]`:
- Column letters: A=col0, B=col1, ..., Z=col25, AA=col26, etc.
- Row numbers: 1-indexed (ODF convention). Header row is row 1, first data row is row 2.
  Mapping: ODF row = Go struct row index + 1.
- Reference to data_a sheet, column C (Value), row 3: `[$data_a.C3]`

### ODS Library Usage (go-ods v1.0.7)

- **Template**: Embed a minimal empty ODS file via Go `embed.FS`. Read it at startup:
  `data, files, err := ods.Read(templatePath)` then `defer files.Close()`.
  The `files` zip reader is required by `ods.Write` to preserve the ODF zip structure.
- **Uncompress**: `ods.Uncompress(data.Content, 20)` — always call after Read to
  expand column repetition compression.
- **Sheet construction**: Clear the default sheet's rows (`sheet.TableRow = nil`)
  or rename it, then append additional `Table{}` structs as needed.
- **Cell values**: `TableCell.P` for text, `TableCell.Value` + `TableCell.ValueType =
  "float"` for numbers, `TableCell.Formula` (with `of:=` prefix) for formulas.
  For formula cells, leave `Value` empty.
- **Output**: `ods.Write(path, data, files)` — pass the same `*zip.ReadCloser`
  returned by `ods.Read`.
- **Column letter helper**: `colLetter(col int) string` — converts 0-indexed column
  index to ODF column letter (A, B, ..., Z, AA, AB, ...).

---

## Component 2: ODS Verifier (`ods_verify`)

### Inputs

| Input | Source | Purpose |
|-------|--------|---------|
| `-parquet=<path>` | CLI flag | Path to `output.parquet` produced by the synthesis pipeline |
| `-ods=<path>` | CLI flag | Path to `expected_output.ods` produced by the generator |

The verifier auto-discovers the schema file at `<ods_dir>/expected_output_schema.json`.
No `-schema` flag is needed — configuration by convention.

### Verification Workflow

1. Read `expected_output_schema.json` from the same directory as the `.ods` file.
   Obtain `key_columns`, `metric_column`, `flag_column`, and `value_column`.
2. Open the `.ods` file via `ods.Read(path)` and `ods.Uncompress(content, 20)`.
3. Find the `expected_output` sheet by iterating `data.Content.Body.Spreadsheet.Table`
   and matching `Table.Name`.
4. Iterate data rows (row 1 onward), filtering for `Flag = "final"` by checking
   `row.TableCell[flagCol].P`.
5. For each final row, extract:
   - **Key values**: `TableCell.P` from key columns (indices 0 to K-1, where K is
     `len(key_columns)`).
   - **Metric name**: `row.TableCell[metricCol].P`
   - **Expected value**: `strconv.ParseFloat(row.TableCell[valueCol].Value, 64)`
6. Read `output.parquet` using `cplio.ReadParquetFile` and `cplio.ReadParquetColumnData`.
   Build a parquet row index keyed by (key_values, metric) for O(1) lookup.
   The parquet schema's columns preceding "Metric" provide the key values in the
   same flattened positional order as the ODS.
7. For each expected row, look up the matching parquet row by composite key
   (key_values, metric). Compare values with tolerance.
8. Report mismatches; exit 0 for pass, 1 for mismatch, 2 for error.

### Composite Key Matching

Expected rows are matched to parquet rows by a composite key of:
- The ordered list of key column values (dimension values followed by grain values,
  in the flattened order used by the generator and parquet file)
- The metric name

Both files use the same flattened column ordering. The verifier uses positional
indices (not column names) to read key values from both the ODS file (using
column indices derived from the schema) and the parquet file (columns in schema
order before "Metric").

### Value Comparison

Reuses the `compareValues` logic from `parquet_verify.go`. Absolute tolerance
of 1e-9 with a 1e-7 relative factor to account for IEEE 754 float64 precision
loss across large magnitude ranges:

```go
func compareValues(expected, actual float64, tolerance float64) bool {
    diff := math.Abs(expected - actual)
    return diff <= tolerance*(1+1e-7)
}
```

### Parquet Reading

Uses the existing `cplio` package:

```go
import cplio "github.com/chronocone/chronocone_planning_language/src/main/go/io"

schema, numRows, err := cplio.ReadParquetFile(parquetPath)
columnData, err := cplio.ReadParquetColumnData(parquetPath, colIndex)
// columnData is either []string or []float64
```

The verifier discovers the column layout from the parquet schema: columns
before "Metric" are key columns, and "Value" is the value column.

### Formula Cell Handling

Formula cells (branch rows) have `Value = ""` (empty, per generator design).
The verifier skips these by filtering on `Flag = "final"`. If a final row's
Value cell cannot be parsed as float64, the verifier reports an error (exit code 2).

### Mismatch Reporting

Mismatches are reported with:
- Type: `"missing"` (in ODS but not parquet), `"extra"` (in parquet but not ODS),
  or `"value_mismatch"` (both present but values differ beyond tolerance).
- Row coordinates: ODS row number and parquet row context.
- Expected vs actual values.

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All expected rows match parquet output within tolerance |
| 1 | Differences found between expected and actual values |
| 2 | Error: file not found, parse error, schema mismatch, missing columns |

---

## Design Decisions

**Reuse existing parser/analyzer rather than creating a new CPL text parser**
: The ODS generator uses `parser.ParseFiles` + `analysis.Analyze` to obtain the
  CIL tree for each metric, rather than introducing a new text-based regex or
  indentation parser.

: **Alternatives considered:**
: - Text-based regex/indentation parser: Rejected because it would duplicate
  CPL grammar knowledge, risk divergence from the compiler as CPL evolves,
  and require maintenance of two separate parsers.
: - ANTLR parse tree direct access: Rejected because the analysis phase's CIL
  tree provides higher-level, semantically validated structures (resolved
  conditionals, postfix expressions) that are easier to translate to formulas.

: **Rationale:**
: The existing CIL tree preserves the full expression structure (arithmetic,
  comparisons, conditionals) in a form that maps directly to both formula
  generation and evaluation. Using the same pipeline as the compiler eliminates
  the risk of the generator producing different results from the compiler for
  the same input CPL. The CIL tree is the single source of truth for both code
  generation and expected output generation.

: **References:**
: - [`go-analysis/cil-representation.md`](go-analysis/cil-representation.md)
: - [ERO005: Metrics Analysis Phase](eros/ERO005_metrics-analysis-phase.md)

---

**Minimal template ODS as starting point (required by go-ods v1.0.7 API)**
: The generator embeds a minimal empty ODS file via Go `embed.FS`, reads it with
  `ods.Read` to obtain both the struct and the `*zip.ReadCloser`, then modifies
  the struct and writes the result with `ods.Write`.

: **Alternatives considered:**
: - Construct ODS struct from scratch with `ods.Write(path, data, nil)`: Rejected
  because the go-ods `Write` function's `files` parameter is documented to receive
  the `*zip.ReadCloser` from `Read`, and passing `nil` is untested/unproven. The
  library preserves the original zip structure (manifest, mimetype, etc.) from
  the reader, so a nil reader may produce invalid output.
: - Generate template as ZIP via `archive/zip`: Rejected as more complex than
  embedding a pre-built minimal ODS.

: **Rationale:**
: The go-ods API is designed around the Read→Modify→Write pattern. Using a
  minimal embedded template guarantees a valid ODF zip structure (correct mimetype
  entry, manifest, namespace declarations) and is the proven, documented approach
  for the v1.0.7 library. The template is a tiny empty spreadsheet — a few KB —
  embedded at compile time via `//go:embed`.

: **References:**
: - [go-ods API Reference](../../../../resources/apis/go-ods/api-reference.md) —
  "Creating a New ODS File (Template-Based Approach)"

---

**All dimensions treated uniformly at verification level**
: The generator and verifier treat grain dimensions and concept dimensions as
  a single ordered list of named columns. No distinction is made between the
  two categories.

: **Alternatives considered:**
: - Separate "dimensions" and "grain" column groups: Rejected because it
  adds complexity at the verification level for a distinction that is only
  meaningful during CPL parsing/analysis (where the `grain` keyword enforces
  time-related constraints).

: **Rationale:**
: At the verification level, all dimensions derive from concepts and appear
  as named string columns in data files. Bifurcating them into separate
  groups complicates sheet layout, column indexing, and composite key
  construction without providing any benefit to the verification workflow.

: **References:**
: - Lesson: **Grain dimensions and non-grain dimensions are structurally
  identical at the verification/example level**
: - Concept: **Dimension Uniformity at Verification Level**

---

**CIL→text formatter for branch labels rather than raw CPL text**
: Branch labels are generated from the CIL tree using a formatter that
  produces human-readable representations of conditions and expressions
  (e.g., `"data_a == 10"`, `"data_a + data_b"`, `"if data_a == 10"`).

: **Alternatives considered:**
: - Scanning raw CPL text for verbatim clause text: Rejected because the CIL
  tree already contains the canonical expression structure, and scanning raw
  text would require matching parsed constructs back to source positions —
  adding complexity without significant benefit.
: - SourceLocation-based text extraction: Partially considered but rejected
  because SourceLocation gives line/column ranges, requiring source file
  re-reading and text slicing that is fragile against formatting variations.

: **Rationale:**
: The CIL formatter produces labels that are structurally consistent and
  visually matchable to CPL source. The key purpose of the Branch column is
  traceability — a user opening `expected_output.ods` in Calc can identify
  which branch a formula corresponds to. The formatted CIL representation
  achieves this without the complexity of raw source tracking.

: **References:**
: - [`go-analysis/cil-representation.md`](go-analysis/cil-representation.md)

---

**Schema metadata file with configuration-by-convention naming**
: The generator produces an `expected_output_schema.json` alongside the `.ods` file,
  and the verifier discovers it by convention (`<ods_dir>/expected_output_schema.json`)
  rather than requiring a separate CLI flag.

: **Alternatives considered:**
: - `-schema=<path>` CLI flag: Rejected because it's redundant when a naming
  convention makes the relationship implicit.
: - Header row introspection in the verifier: Rejected because it requires the
  verifier to parse the ODS header row and find columns by name, which is more
  code and more error-prone than reading a simple JSON file.

: **Rationale:**
: The schema file decouples the verifier's column discovery from the ODS format.
  It makes the verifier simpler (no header matching, no column-name search) and
  lets the generator — which already knows the column layout — communicate it
  directly. The convention-based naming (`expected_output_schema.json` in the
  same directory as the `.ods` file) is unambiguous within the example directory
  context where these tools operate.

: **References:**
: - Design Principle: **Schema metadata file (configuration by convention)**
