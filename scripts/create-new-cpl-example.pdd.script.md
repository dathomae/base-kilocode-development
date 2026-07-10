# create-new-cpl-example

This script creates a new complete CPL example under `examples/with_data` with all required files for compilation, execution, and verification. It follows the established pattern used by existing example categories (e.g., `metrics_arithmetic`, `metrics_conditionals`, `lookups`).

## Parameters

| Name | Type | Description | Required | Default Value |
|------|------|-------------|----------|---------------|
| category_dir | string | Path to the example category directory (e.g., `examples/with_data/metrics_arithmetic`). This directory must contain or will contain `core.cpl`. | Yes | N/A |
| example_name | string | Name of the new example subdirectory and base name for its `.cpl` file (e.g., `simple_addition`). Must be unique within the category. | Yes | N/A |
| iospec_json | string | JSON string or path to the `iospec.json` file defining inputs, outputs, and architecture. | Yes | N/A |
| cpl_source | string | Path to the example-specific `.cpl` source file, or inline CPL source content. | Yes | N/A |
| data_files | list | List of `data_*.json` file paths to copy into the example directory. At least one is required. | Yes | N/A |

## Operations

1. **Validate parameters and directory structure**:
   - Confirm `category_dir` and `example_name` are provided and non-empty.
   - Compute `example_dir = <category_dir>/<example_name>`.
   - Verify `example_dir` does not already exist. If it does, report an error and terminate.
   - Verify `core.cpl` exists at `<category_dir>/core.cpl`. If not, report a warning that the category may be incomplete.
   - Verify at least one `data_*.json` file is provided in `data_files`.
   - For each data file, confirm the filename matches the pattern `data_<name>.json`.

2. **Create example subdirectory**:
   - Create `example_dir`.

3. **Write `iospec.json`**:
   - If `iospec_json` is a path to an existing file, copy it to `<example_dir>/iospec.json`.
   - Otherwise, write the `iospec_json` string directly to `<example_dir>/iospec.json`.
   - Validate that the JSON contains at least one entry in `inputs` and one in `outputs`.

4. **Write `.cpl` source file**:
   - The target filename is `<example_dir>/<example_name>.cpl`.
   - If `cpl_source` is a path to an existing file, copy it.
   - Otherwise, write the inline string.
   - Verify the file contains at least one `metrics` block.

5. **Copy data files**:
   - Copy each provided `data_*.json` into `example_dir`.
   - Verify each file contains a JSON array of objects with string columns and a `Value` float64 field.

6. **Regenerate parquet files**:
   - Run the following command from the project root (where `src/chronocone_planning_language` resides):
     ```bash
     cd src/chronocone_planning_language && bash shell_scripts/regenerate_data.sh <example_name> <category_dir>
     ```
   - This invokes `go run src/main/go/cli/parquetgen/parquet_gen.go` for each `data_*.json` to produce the corresponding `data_*.parquet`.
   - If the command fails, report the error and terminate the script.

7. **Generate expected output**:
   - Run the XLSX generator CLI:
     ```bash
     cd src/chronocone_planning_language && go run src/main/go/cli/expected_xls_generator/main/main.go -example=<example_dir>
     ```
   - This produces `expected_output.xlsx` and `expected_output_schema.json` in the example directory.
   - If generation fails, report the error and terminate the script.

8. **Compile and verify**:
   - Run the compile-and-run pipeline:
     ```bash
     cd src/chronocone_planning_language && bash shell_scripts/compile_and_run.sh <example_name> <category_dir>
     ```
   - This invokes:
     - `go run src/main/go/cli/compiler/cpl_compiler.go` to compile `core.cpl` + `<example_name>.cpl` into `generated/cmd/<example_name>/main.go`.
     - `go run <generated_main>` to produce `output.parquet`.
     - `go run src/main/go/cli/xls_verify/main.go` to compare `output.parquet` against `expected_output.xlsx`.
   - If compilation fails, report the compiler errors.
   - If verification fails, report the mismatched cells but do NOT attempt to auto-fix data discrepancies.

9. **Report results**:
   - Output a summary containing:
     - Full path to `example_dir`
     - List of all files created (iospec.json, .cpl, data_*.json, data_*.parquet, expected_output.xlsx, expected_output_schema.json)
     - Parquet regeneration status
     - XLSX generation status
     - Compilation and verification status (pass / fail with details)

## Output

The script outputs:
- The absolute path to the newly created example directory.
- A complete file listing within that directory.
- Status of each pipeline stage (parquet regeneration, XLSX generation, compile-and-run verification).
- If verification fails: a list of mismatched cells with expected vs actual values.

## Required File Reference

A complete CPL example requires these files in its subdirectory:

| File | Purpose | Generated or Author-Supplied |
|------|---------|------------------------------|
| `<example_name>.cpl` | Example-specific CPL source defining metrics, connectors, and any local concepts/measures. | Author-supplied |
| `iospec.json` | IO specification mapping connector names to parquet input paths and declaring output metrics. | Author-supplied |
| `core.cpl` (in parent category directory) | Shared infrastructure: concepts, types, units, measures, and sometimes connectors. | Author-supplied (category-level) |
| `data_*.json` | Input data in JSON array-of-objects format with string columns and `Value` float64. | Author-supplied |
| `data_*.parquet` | Parquet files generated from `data_*.json` by `parquet_gen.go`. | **Generated** by `regenerate_data.sh` |
| `expected_output.xlsx` | Expected output spreadsheet for verification. | **Generated** by `expected_xls_generator` |
| `expected_output_schema.json` | Schema describing key columns, input columns, metric/value column indices. | **Generated** by `expected_xls_generator` |
| `output.parquet` | Actual output produced by running the compiled CPL program. | **Generated** by `compile_and_run.sh` |

## Notes

- The `compile_and_run.sh` script automatically discovers `core.cpl` in the parent directory of the example and prepends it to the compiler invocation.
- The `regenerate_data.sh` script expects `data_*.json` files and produces `data_*.parquet` files with matching `<name>`.
- The `expected_xls_generator/main/main.go` CLI expects `data_*.json` files, an `iospec.json`, and `.cpl` files (including `core.cpl` in the parent directory).
- Standard definitions under `standard_definitions/*.cpl` are automatically prepended by the compiler and the XLSX generator.
- If the example category does not yet have `core.cpl`, create it before running this script.
