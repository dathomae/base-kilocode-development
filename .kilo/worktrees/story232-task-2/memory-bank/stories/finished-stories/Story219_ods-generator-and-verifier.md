# Story219: ODS Generator and Verifier

## Goal

Implement the ODS generator and verifier CLI tools per the generator/verifier plan, producing human-inspectable LibreOffice Calc spreadsheets with per-branch formula rows and static final-value rows for automated verification against `output.parquet`.

## References

- [Plan: ODS Generator and Verifier](../plans/generator_verifier_plan.md)
- [Design: Example Generation and Verification](../specifications/design/cpl/example_generation_and_verification.md)
- [go-ods API Reference](../../resources/apis/go-ods/api-reference.md)
- [go-ods Version History](../../resources/apis/go-ods/version-history.md)

## Dependencies

None — the go-ods library (`github.com/AlexJarrah/go-ods v1.0.7`) is already present in `go.mod`. The CIL, analysis, parser, and cplio packages required by the generator are existing, completed packages.

## Dependent Stories

- [Story218_transition-to-ods-verification.md](Story218_transition-to-ods-verification.md) — depends on this story's generator and verifier as prerequisites

## Tasks

### Task Sizing

- **Small tasks**: Tasks 1, 2, 5, 8, 9, 10, 11 — each touches at most 2 files, is additive or purely execution-based, and follows an established pattern.
- **Not small tasks**: Tasks 3 (multiple concerns: CLI, ODS reading, parquet reading, matching, comparison), 4 (complex tree-walking evaluator with 10+ node types and 14+ test cases), 6 (sophisticated prior-negation AND chain logic for nested IF formulas), 7 (multi-domain coordination: sheet building, CLI orchestration, schema writing).

### Worktree Suitability

- **Group A (Tasks 1, 2):** — fully disjoint packages (`template/` vs `ods_utils/`), purely additive, safe for subagents
- **Group B (Tasks 4, 5, 6): [Worktree]** — all three tasks create files in `src/main/go/cli/expected_ods_generator/` (same package directory), same-package compilation risk
- **Group C (Tasks 9, 10):** — fully disjoint example directories, safe for subagents

### Parallel Execution

- **Group A** (Tasks 1, 2): Task 1 creates the embedded template binary; Task 2 creates the ods_utils Go package. Fully disjoint packages, no shared infrastructure, no mutual dependencies.
- **Group B** (Tasks 4, 5, 6) — [Worktree]: All in `expected_ods_generator/` package. Disjoint files (`cil_evaluator.go`, `cil_to_text.go`, `cil_to_odf.go`) but same-package compilation risk. Execute in isolated git worktrees.
- **Group C** (Tasks 9, 10): Task 9 validates `simple_addition`; Task 10 validates `conditional_comparison_operators`. Fully disjoint example directories.

### Execution Order

```
[Task 1 ‖ Task 2] → [Task 3 ‖ Group B (Tasks 4, 5, 6)] → Task 7 → Task 8 → [Task 9 ‖ Task 10] → Task 11
```

### Worktree Reintegration

**Group B (Tasks 4, 5, 6):**
- Merge order: any order (fully disjoint files, no cross-references — each task creates its own `.go` file consuming the external `cil` package independently)
- Integration tests: `go test ./src/main/go/cli/expected_ods_generator/... -count=1 -timeout 60s` — verify the combined package compiles and `cil_evaluator_test.go` passes
- Watch for conflicts in: `go.sum` (dependency resolution in separate worktrees may produce different hashes for shared indirect dependencies)

### Task 1: Create minimal ODS template — Completed [Small]

Create `src/main/go/cli/template/template.ods` — a minimal valid ODF spreadsheet with one empty sheet, embedded at compile time via `//go:embed`. LibreOffice Calc is a required development environment dependency.

1. Launch LibreOffice Calc with a new empty spreadsheet:
   ```bash
   libreoffice --calc
   ```
2. Save the new spreadsheet as `template.ods` at `src/main/go/cli/template/template.ods`.
3. Close Calc. Add the `go:embed` directive in the template package:

```go
//go:embed template.ods
var TemplateODS []byte
```

The generator reads this template at runtime to obtain the `ods.ODS` struct and `*zip.ReadCloser` required by `ods.Write`, then modifies and writes the result. Since LibreOffice Calc produces a standards-compliant ODF ZIP file, the `mimetype` entry will already have the correct storage format (stored, not deflated).

- a. Launch Calc with `libreoffice --calc`, create new spreadsheet, save as `template.ods` — Not Started
- b. Add `go:embed` directive and `TemplateODS` variable — Not Started

### Task 2: Implement shared ods_utils package — Completed [Small]

Create `src/main/go/cli/ods_utils/ods_utils.go` with four shared utilities used by both generator and verifier:

- `ColLetter(col int) string` — converts 0-indexed column index to ODF column letter (A=col0, B=col1, ..., Z=col25, AA=col26, etc.)
- `SchemaMeta` struct — shared type for generator output metadata and verifier input:
  ```go
  type SchemaMeta struct {
      ODSFile      string   `json:"ods_file"`
      KeyColumns   []string `json:"key_columns"`
      MetricColumn int      `json:"metric_column"`
      FlagColumn   int      `json:"flag_column"`
      ValueColumn  int      `json:"value_column"`
  }
  ```
- `OpenODS(path string) (ods.ODS, *zip.ReadCloser, error)` — encapsulates `ods.Read` + `ods.Uncompress`
- `FindSheet(data *ods.ODS, name string) (*ods.Table, error)` — iterates tables to find one by name

Depends on go-ods only (already in go.mod).

- a. Write function signatures and struct definitions — Not Started
- b. Implement ColLetter with full alphabet mapping — Not Started
- c. Implement OpenODS and FindSheet helpers — Not Started

### Task 3: Implement ODS verifier (ods_verify) — Completed

Create `src/main/go/cli/ods_verify/main.go` with a CLI that:
- Accepts `-parquet=<path>` and `-ods=<path>` flags
- Auto-discovers `expected_output_schema.json` from the ODS file path (`<ods_dir>/expected_output_schema.json`)
- Opens the `.ods` file via `ods_utils.OpenODS`, finds the `expected_output` sheet via `ods_utils.FindSheet`
- Iterates rows, filters for `Flag = "final"` by checking `row.TableCell[flagCol].P`
- Extracts key values (dimensions + grain flattened), metric name, and `Value` via `strconv.ParseFloat(cell.Value, 64)`
- Builds a map keyed by `(key_values_string|metric_name)` → expected value
- Reads `output.parquet` using `cplio.ReadParquetFile` and `cplio.ReadParquetColumnData`, builds an actual-values map
- Compares values with tolerance using `compareValues(expected, actual, 1e-9)`
- Reports mismatches: `"missing"` if expected key not in actual, `"value_mismatch"` if values differ beyond tolerance, `"extra"` if actual keys not in expected
- Exits 0 if all match, 1 if differences, 2 on error

Create `src/main/go/cli/ods_verify/main_test.go` with unit tests covering:
- `compareValues` with edge cases (zero, negative, large magnitudes, NaN)
- Composite key string construction
- Schema meta JSON parsing
- Value comparison and matching logic (focused unit tests)

- a. Write test functions in main_test.go first (compareValues, key construction, schema parsing) — Not Started
- b. Implement main.go with CLI, ODS reading, and row filtering — Not Started
- c. Implement parquet reading and composite-key matching — Not Started
- d. Implement value comparison and mismatch reporting — Not Started
- e. Verify tests pass — Not Started

### Task 4: Implement CIL evaluator with TDD — Completed

Following the Logical TDD Lifecycle (Phase A: signatures → Phase B: tests → Phase C: implementation → Phase D: verify).

Create `src/main/go/cli/expected_ods_generator/cil_evaluator.go` with:

**`EvaluateExpression(node cil.CILNode, valueLookup map[string]float64) (float64, error)`** — tree-walking evaluator handling all CIL node types:
- `CILLiteral{Value: v}` → return `v`
- `CILInputRef{Name: name}` → return `valueLookup[name]`
- `CILBinaryOp{Op, Left, Right}` → evaluate left and right, apply `+`, `-`, `*`, `/`
- `CILUnaryOp{Op, Operand}` → return `-operand`
- `CILComparisonOp{Op, Left, Right}` → evaluate left and right, apply `==`, `!=`, `<`, `>`, `<=`, `>=`; return 1.0 if true, 0.0 if false
- `CILLogicalOp{Op, Left, Right}` → evaluate left and right; `and`: 1.0 if both non-zero; `or`: 1.0 if either non-zero
- `CILNotOp{Operand}` → return 1.0 if operand is 0, else 0.0
- `CILBoolLiteral{Value: v}` → return 1.0 if true, 0.0 if false
- `CILNilLiteral{}` → return 0.0
- `CILConditional{Branches}` → evaluate each branch condition in order; first true condition's body value is returned. If none match, return 0.0. A nil condition (else branch) is always true.
- `CILWhenClause{Clauses, Default}` → evaluate each clause condition; first matched body returned. If none match, evaluate and return Default.

**`EvaluateMetric(metricDef *analysis.AnalysisMetricDefinition, dataRows []InputRow) ([]OutputRow, error)`** — for each unique (dimension, grain) combination in dataRows, builds a valueLookup and calls EvaluateExpression.

Create `src/main/go/cli/expected_ods_generator/cil_evaluator_test.go` with tests covering:
- Literal evaluation (int, float64)
- Binary arithmetic (`+`, `-`, `*`, `/`)
- Unary negation
- Comparison operators (`==`, `!=`, `<`, `>`, `<=`, `>=`)
- Logical operators (`and`, `or`, `not`)
- Boolean/nil literals
- Simple conditional (if/else) with two branches
- Multi-branch conditional (if/else if/else) with three branches
- When/default clause
- Truthiness (nil as false, non-zero as true)
- Nested expressions (`(data_a + data_b) * 2`)
- Unmatched conditional (no branch true → returns 0.0)
- Each tested with hand-constructed CIL nodes using the `cil` package types

- a. Phase A: Write all function signatures and docstrings — Not Started
- b. Phase B: Write comprehensive test cases in cil_evaluator_test.go (must FAIL at this stage) — Not Started
- c. Phase C: Implement EvaluateExpression for all node types — Not Started
- d. Phase C: Implement EvaluateMetric — Not Started
- e. Phase D: Run tests — all must pass — Not Started

### Task 5: Implement CIL→text formatter — Completed [Small]

Create `src/main/go/cli/expected_ods_generator/cil_to_text.go` with:

**`ExpressionToText(node cil.CILNode) string`** — maps CIL node types to human-readable branch labels:

| CIL Node | Text Output |
|----------|-------------|
| `CILLiteral{42}` | `"42"` |
| `CILInputRef{name:"data_a"}` | `"data_a"` |
| `CILBinaryOp{+, L, R}` | `"L + R"` |
| `CILBinaryOp{-, L, R}` | `"L - R"` |
| `CILBinaryOp{*, L, R}` | `"L * R"` |
| `CILBinaryOp{/, L, R}` | `"L / R"` |
| `CILUnaryOp{negate, Op}` | `"-Op"` |
| `CILComparisonOp{==, L, R}` | `"L == R"` |
| `CILComparisonOp{!=, L, R}` | `"L != R"` |
| `CILComparisonOp{<, L, R}` | `"L < R"` (and similarly for `>`, `<=`, `>=`) |
| `CILLogicalOp{and, L, R}` | `"L and R"` |
| `CILLogicalOp{or, L, R}` | `"L or R"` |
| `CILNotOp{Op}` | `"not Op"` |
| `CILBoolLiteral{true}` | `"True"` |
| `CILBoolLiteral{false}` | `"False"` |
| `CILNilLiteral{}` | `"nil"` |
| `CILConditional` branch 0 | `"if <condition_text>"` |
| `CILConditional` branch N | `"else if <condition_text>"` |
| `CILConditional` else | `"else"` |
| `CILWhenClause` case N | `"when <condition_text>"` |
| `CILWhenClause` default | `"default"` |

- a. Write ExpressionToText function signature — Not Started
- b. Implement switch cases for all CIL node types — Not Started
- c. Implement conditional/when branch label formatting — Not Started

### Task 6: Implement CIL→ODF formula translator — Completed

Create `src/main/go/cli/expected_ods_generator/cil_to_odf.go` with:

**`ExpressionToODF(node cil.CILNode, connectorSheet string, rowIdx int, keyColCount int, colLetter func(int) string) string`** — translates a CIL expression tree to an ODF formula string with `of:=` prefix:

| CIL Node | ODF Output |
|----------|------------|
| `CILLiteral{v}` | string representation of `v` |
| `CILInputRef{name}` | `[$name.<colLetter(keyColCount)><rowIdx+1>]` |
| `CILBinaryOp{+/-/*///, L, R}` | `(` + translate(L) + op + translate(R) + `)` |
| `CILUnaryOp{negate, Op}` | `-` + translate(Op) |
| `CILComparisonOp{==, L, R}` | translate(L) + `=` + translate(R) |
| `CILComparisonOp{!=, L, R}` | translate(L) + `<>` + translate(R) |
| `CILComparisonOp{<, L, R}` / `>` / `<=` / `>=` | translate(L) + op + translate(R) |
| `CILLogicalOp{and, L, R}` | `AND(` + translate(L) + `,` + translate(R) + `)` |
| `CILLogicalOp{or, L, R}` | `OR(` + translate(L) + `,` + translate(R) + `)` |
| `CILNotOp{Op}` | `NOT(` + translate(Op) + `)` |
| `CILBoolLiteral{true}` | `TRUE()` |
| `CILBoolLiteral{false}` | `FALSE()` |
| `CILNilLiteral{}` | `0` |
| `CILConditional{Branches}` | Chain of nested `IF(condition, body, rest)` with prior-negation AND logic |
| `CILWhenClause{Clauses, Default}` | Chain of nested `IF(cond1, body1, IF(cond2, body2, ..., default))` |

**`FormulaForBranch(branches []cil.CILConditionalBranch, branchIdx int, ...) string`** — per-branch activation: for branch N (1-indexed), the activation condition is `AND(NOT(cond_1), NOT(cond_2), ..., NOT(cond_{N-1}), cond_N)`. For the else branch (nil condition), it's `AND(NOT(cond_1), ..., NOT(cond_{N-1}))`. Wrap as `IF(activation, body_result, "")`.

**Special case**: Arithmetic-only metrics (no conditional branches) produce one formula row with just the translated expression (no IF wrapper).

- a. Write function signatures — Not Started
- b. Implement ExpressionToODF for literals, input refs, arithmetic ops — Not Started
- c. Implement ExpressionToODF for comparison and logical ops — Not Started
- d. Implement ExpressionToODF for conditionals (nested IF chain) and when clauses — Not Started
- e. Implement FormulaForBranch with prior-negation AND logic — Not Started
- f. Implement arithmetic-only special case — Not Started

### Task 7: Implement generator sheet builder and CLI — Completed

Create `src/main/go/cli/expected_ods_generator/generator.go` with:

**`GenerateODS(defs *analysis.AnalysisDefinitions, iospec *io.IOSpecification, inputData map[string][]map[string]interface{}, data *ods.ODS) error`**:
1. **Input sheets**: For each connector in iospec.inputs, add a sheet with header row (dimension names + grain names + `Value`) and data rows (one per object in the data JSON array). Sheet name = connector name.
2. **`expected_output` sheet**: For each metric in defs.Metrics, for each (dimension, grain) combination from the Cartesian product:
   - Generate branch rows (formula via `cil_to_odf` + label via `cil_to_text`)
   - Generate one final row (static value via `cil_evaluator`)
   - Flag = `"branch"` or `"final"`
3. **Column layout**: Key columns (dimensions + grain flattened), Metric column, CPL File column (basename), Branch column (CIL→text label), Flag column, Value column (formula string for branch, static float64 for final).

**Dimension/grain Cartesian product**: Derive the set of (dimension, grain) combinations from actual input data rows.

**ODF formula construction**: Formulas reference input data sheets using bracket notation `[$SheetName.ColRow]`. The value column in input sheets is at index `len(dimensions) + len(grain)` (0-indexed).

Create `src/main/go/cli/expected_ods_generator/main.go` as the CLI entry point:
- Flags: `-example=<dir>` (required), `-output=<path>` (optional, defaults to `<example>/expected_output.ods`)
- Pipeline: read template → collect `.cpl` files → `parser.ParseFiles` → `analysis.Analyze` → load `iospec.json` → read `data_*.json` → `GenerateODS` → `ods.Write` → write `expected_output_schema.json`

- a. Write function signatures and docstrings — Not Started
- b. Implement input sheet generation in generator.go — Not Started
- c. Implement expected_output sheet generation with Cartesian product and per-metric row generation — Not Started
- d. Implement main.go CLI entry point with full pipeline orchestration — Not Started

### Task 8: Update compile_and_run.sh — Completed [Small]

Modify `shell_scripts/compile_and_run.sh` to use `ods_verify` instead of `parquet_verify`. Ensure the script references `main.go` per the plan's naming convention:

```bash
ODS_FILE="${EXAMPLE_DIR}/expected_output.ods"
OUTPUT_PARQUET="output.parquet"

if [ -f "${ODS_FILE}" ]; then
    if ! go run src/main/go/cli/ods_verify/main.go \
        -parquet="${OUTPUT_PARQUET}" \
        -ods="${ODS_FILE}"; then
        echo ""
        echo "Verification FAILED. See differences above."
        exit 1
    fi
    echo "Verification PASSED."
else
    echo "Warning: expected_output.ods not found, skipping verification."
fi
```

Replace `EXPECTED_JSON` references with `EXPECTED_ODS`. Update comments to reference ODS. All other script logic unchanged.

- a. Update ODS file reference and ods_verify invocation — Not Started
- b. Update comments and variable name references — Not Started
- c. Verify script syntax is valid bash — Not Started

### Task 9: Validate with simple_addition — Completed [Small]

Validate the end-to-end arithmetic-only flow:

1. Run the generator: `go run src/main/go/cli/expected_ods_generator/main.go -example=examples/with_data/metrics_arithmetic/simple_addition`
2. Verify `expected_output.ods` was created
3. Verify `expected_output_schema.json` was created with correct key_columns
4. Run the generated program: `go run generated/cmd/simple_addition/main.go -data-dir=examples/with_data/metrics_arithmetic/simple_addition/`
5. Run the verifier: `go run src/main/go/cli/ods_verify/main.go -parquet=output.parquet -ods=examples/with_data/metrics_arithmetic/simple_addition/expected_output.ods`
6. Confirm verifier exits 0 (pass)

**User inspection stop**: After steps 1-6 pass, pause and ask the user to open `expected_output.ods` in LibreOffice Calc:
```bash
libreoffice examples/with_data/metrics_arithmetic/simple_addition/expected_output.ods
```
The user should confirm:
- Input data sheets have correct data matching `data_*.json`
- The formula row for `simple_addition` computes correctly
- The final row static value matches the formula result

Do not proceed to Task 10 until the user approves.

- a. Run generator against simple_addition example — Not Started
- b. Verify expected_output.ods and expected_output_schema.json created — Not Started
- c. Compare evaluator-computed values against existing `expected_output.json`; overwrite if different (the CIL evaluator is the authoritative source of correct values) — Not Started
- d. Run compile_and_run.sh and confirm ods_verify passes — Not Started
- e. User inspection stop — await user confirmation — Not Started

### Task 10: Validate with conditional_comparison_operators — Completed [Small]

Validate the end-to-end conditional-branch flow:

1. Run the generator against `examples/with_data/metrics_conditionals/conditional_comparison_operators`
2. Verify `expected_output.ods` was created with multiple branch rows per (dimension, grain) combination
3. Verify `expected_output_schema.json` was created
4. Compare evaluator-computed values against existing `expected_output.json`; overwrite if different
5. Run the generated program and verifier (same pattern as Task 9)
6. Confirm verifier exits 0 (pass)

**User inspection stop**: Pause and ask the user to open `expected_output.ods` in LibreOffice Calc:
```bash
libreoffice examples/with_data/metrics_conditionals/conditional_comparison_operators/expected_output.ods
```
The user should confirm:
- Input data sheets have correct data
- Branch formulas compute correctly for each conditional branch
- Each final row static value matches its corresponding active branch result
- The six branches + final per combination are correctly laid out

- a. Run generator against conditional_comparison_operators example — Not Started
- b. Verify expected_output.ods has correct branch structure — Not Started
- c. Compare evaluator-computed values against existing `expected_output.json`; overwrite if different — Not Started
- d. Run compile_and_run.sh and confirm ods_verify passes — Not Started
- e. User inspection stop — await user confirmation — Not Started

### Task 11: Run full test suite — Completed [Small]

Execute all relevant tests to confirm the generator and verifier are working correctly:

```bash
cd src/chronocone_planning_language
go test ./src/main/go/cli/ods_verify/... -v
go test ./src/main/go/cli/expected_ods_generator/... -v
```

All tests must pass.

- a. Run ods_verify unit tests — Not Started
- b. Run evaluator unit tests — Not Started
- c. Summarize results — Not Started

## Test-First Development

This project follows the Logical TDD Lifecycle:
- **Task 3**: Write verifier unit tests first (3a), then implement the verifier (3b–3d).
- **Task 4**: Phase A (signatures) → Phase B (tests, must FAIL) → Phase C (implementation) → Phase D (verify tests pass).
- **Tasks 9, 10**: Each validation task uses `compile_and_run.sh` and `ods_verify` as acceptance tests. User inspection stops provide manual validation of spreadsheet correctness.
- **Task 11**: The full test suite serves as the final acceptance test.

## Constraints

- **Library**: Must use `github.com/AlexJarrah/go-ods` v1.0.7 (already in go.mod). See [go-ods API Reference](../../resources/apis/go-ods/api-reference.md) for the v1.0.7 XML struct API.
- **LibreOffice Calc**: Required development environment dependency for creating the ODS template (Task 1) and for user inspection stops (Tasks 9, 10).
- **Template ODS via go:embed**: The go-ods `Write` function requires the `*zip.ReadCloser` from a prior `Read`. Create the template in Calc, embed it, read it at runtime, then modify and write.
- **CIL tree as single source of truth**: Use `parser.ParseFiles` + `analysis.Analyze` for validated CIL trees. No new CPL parsing.
- **Dimension uniformity**: All columns (dimension + grain) are treated as a single flattened ordered list.
- **Schema metadata by convention**: Generator produces `expected_output_schema.json`; verifier discovers it at `<ods_dir>/expected_output_schema.json`.
- **Branch labels from CIL formatter**: CIL→text produces labels; no raw CPL text scanning.
- **Formula cell value left empty**: Branch rows have `Value = ""` (not `"0"`).
- **Verifier reads schema for column layout**: No ODS header parsing; column indices come from `SchemaMeta`.

## Intent

- **Task 1**: Create a minimal ODS template in LibreOffice Calc and embed it for use by the go-ods library.
- **Task 2**: Provide shared utilities (column letters, schema metadata struct, ODS helpers) consumed by both generator and verifier.
- **Task 3**: Provide an automated verifier that reads human-verifiable spreadsheets and compares final-row values against actual program output.
- **Task 4**: Provide a tree-walking CIL evaluator that computes correct final values from the same CIL tree the compiler uses.
- **Task 5**: Provide human-readable branch labels from CIL expressions.
- **Task 6**: Provide ODF formula translation from CIL expressions, enabling LibreOffice Calc to verify branch logic.
- **Task 7**: Provide the generator CLI that orchestrates the full pipeline: parse, analyze, load data, build sheets, write ODS.
- **Task 8**: Wire the new verifier into the existing build-and-run workflow.
- **Task 9**: Validate the arithmetic-only end-to-end flow with a user inspection stop.
- **Task 10**: Validate the conditional-branch end-to-end flow with a user inspection stop.
- **Task 11**: Confirm all unit tests pass.

## Acceptance Criteria

1. The ODS template (`template.ods`) is embeddable and produces a valid `ods.ODS` struct when read.
2. `ods_utils` package compiles and its helpers are usable by both generator and verifier.
3. `ods_verify` CLI compares `expected_output.ods` final rows against `output.parquet` and reports mismatches accurately.
4. `ods_verify` unit tests cover comparison, key construction, and schema parsing.
5. `cil_evaluator` correctly evaluates all CIL node types and tests pass for all 14+ scenarios.
6. `cil_to_text` produces correct human-readable labels for all CIL node types.
7. `cil_to_odf` produces syntactically valid ODF formulas with correct bracket references and IF-chain logic.
8. `expected_ods_generator` CLI produces a valid `.ods` file with input sheets and an `expected_output` sheet.
9. `compile_and_run.sh` uses `ods_verify` instead of `parquet_verify`.
10. `simple_addition` generates, runs, and verifies successfully; user approves the ODS output in LibreOffice Calc.
11. `conditional_comparison_operators` generates, runs, and verifies successfully; user approves the ODS output.
12. All Go tests (`ods_verify`, `expected_ods_generator`) pass.

## Open Questions

None — all design decisions resolved in the plan.

## Additional Context

- The go-ods v1.0.7 library provides the stable `ods.Read`/`ods.Write`/`ods.Uncompress` API with raw `TableCell` struct access. v2.0.0 was evaluated but is an experimental rewrite not production-ready. See [go-ods Version History](../../resources/apis/go-ods/version-history.md).
- The plan reference is `../plans/generator_verifier_plan.md`. The design document is at `../specifications/design/cpl/example_generation_and_verification.md`.
