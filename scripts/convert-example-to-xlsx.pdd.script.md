# convert-example-to-xlsx

This script converts a single with_data example directory from JSON/ODS verification to XLSX verification. It regenerates parquet files from `data_*.json`, generates `expected_output.xlsx` and `expected_output_schema.json` using the XLSX generator CLI, deletes redundant files (`expected_output.json`, `expected_output.ods`, `parquet_gen_spec_*.json`), and runs the full compile-and-run verification pipeline.

## Parameters

| Name | Type | Description | Required | Default Value |
|------|------|-------------|----------|---------------|
| example_dir | string | Path to the example subdirectory (e.g., `examples/with_data/metrics_arithmetic/simple_addition`) | Yes | N/A |
| example_base_dir | string | Parent directory containing `core.cpl` (e.g., `examples/with_data/metrics_arithmetic`) | Yes | N/A |

## Operations

1. **Validate parameters**:
   - Confirm the `example_dir` parameter was provided and is non-empty.
   - Verify the path `<example_dir>` exists as a directory.
   - Verify that `<example_base_dir>/core.cpl` exists as a regular file.
   - Extract the example name: `<name>` is the last path component of `<example_dir>`.
   - If any validation fails, report the error and terminate the script.

2. **Regenerate parquet from data_*.json**:
   - Run the following command:
     ```bash
     cd src/chronocone_planning_language && bash shell_scripts/regenerate_data.sh <name> <example_base_dir>
     ```
   - If the command fails, report the error and terminate the script.

3. **Generate expected_output.xlsx**:
   - Run the following command:
     ```bash
     cd src/chronocone_planning_language && go run src/main/go/cli/expected_xls_generator/main/main.go -example=<example_dir>
     ```
   - This produces `expected_output.xlsx` and `expected_output_schema.json` in the example directory.
   - If the command fails, report the error and terminate the script.

4. **Delete redundant files**:
   - Delete `<example_dir>/expected_output.json` if it exists.
   - Delete `<example_dir>/expected_output.ods` if it exists.
   - Delete any `<example_dir>/parquet_gen_spec_*.json` files if they exist.
   - Skip any file that does not exist (do not treat missing files as errors).

5. **Verify the conversion**:
   - Run the full compile-and-run pipeline:
     ```bash
     cd src/chronocone_planning_language && bash shell_scripts/compile_and_run.sh <name>
     ```
   - If verification fails with value mismatches, report the mismatched cells clearly. **DO NOT attempt to fix pre-existing data discrepancies.** These represent data issues that exist independently of the conversion and will be investigated separately.
   - If verification passes, note the success.

6. **Report results**:
   - Output a summary including:
     - Example name and directory path
     - Files that were deleted (list each file)
     - Whether XLSX generation succeeded or failed
     - Whether compile_and_run.sh verification passed or failed
     - If verification failed: list the mismatched cells with their expected vs actual values

## Output

The script outputs a results summary containing:
- Example name and full directory path
- List of files deleted (expected_output.json, expected_output.ods, parquet_gen_spec_*.json)
- XLSX generation status (success or failure with error message)
- Verification status (pass or failure with mismatched cell details)
- Any warnings or non-fatal issues encountered

## Error Handling

If required parameters aren't specified, the script terminates with an error message describing which parameter is missing. If validation of parameters fails (directory does not exist, core.cpl missing), the script terminates with a descriptive error. If parquet regeneration or XLSX generation fails, the script terminates with the error output. File deletion errors for non-existent files are not considered failures — the script silently skips missing files. Verification failures are reported but do NOT terminate the script; they are pre-existing data discrepancies that will be investigated separately.
