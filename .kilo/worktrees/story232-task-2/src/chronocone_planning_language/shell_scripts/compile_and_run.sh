#!/bin/bash
set -e

# Usage: ./compile_and_run.sh <directory_name> [example_base_dir]
#
# Where <directory_name> is a subdirectory name searched for under the example
# base directory tree (default: examples/with_data). The subdirectory must
# contain iospec.json, one or more .cpl files, and one or more .parquet files.
#
# If a file named core.cpl exists in the parent directory of the located
# example subdirectory, it is automatically included in the compilation
# alongside the example-specific .cpl files. After the synthesized program
# executes, the script reads iospec.json to discover output metric names and
# runs xls_verify for each <metric>.parquet against expected_output.xlsx when
# the latter exists in the example subdirectory.

if [ $# -lt 1 ] || [ $# -gt 2 ]; then
    echo "Usage: $0 <directory_name> [example_base_dir]"
    echo "  <directory_name>    subdirectory name to search for (e.g., simple_addition)"
    echo "  [example_base_dir]  base directory tree to search within (default: examples/with_data)"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "${SCRIPT_DIR}/.."

DIR_NAME="$1"
EXAMPLE_BASE_DIR="${2:-examples/with_data}"

# Search for the directory name anywhere under the base directory tree
EXAMPLE_DIR=$(find "${EXAMPLE_BASE_DIR}" -mindepth 1 -type d -name "${DIR_NAME}" | head -n 1)

if [ -z "${EXAMPLE_DIR}" ]; then
    echo "Error: Directory '${DIR_NAME}' not found under '${EXAMPLE_BASE_DIR}'."
    exit 1
fi

if [ ! -d "${EXAMPLE_DIR}" ]; then
    echo "Error: Directory '${EXAMPLE_DIR}' does not exist."
    exit 1
fi

IOSPEC_FILE="${EXAMPLE_DIR}/iospec.json"
if [ ! -f "${IOSPEC_FILE}" ]; then
    echo "Error: iospec.json not found in '${EXAMPLE_DIR}'."
    exit 1
fi

# Locate core.cpl in the parent directory of the found example subdirectory
EXAMPLE_PARENT_DIR=$(dirname "${EXAMPLE_DIR}")
CORE_CPL="${EXAMPLE_PARENT_DIR}/core.cpl"
if [ -f "${CORE_CPL}" ]; then
    echo "Found core.cpl: ${CORE_CPL}"
else
    echo "Warning: core.cpl not found in '${EXAMPLE_PARENT_DIR}'. Proceeding without shared infrastructure."
    CORE_CPL=""
fi

# Collect all .cpl files in the example subdirectory
CPL_FILES=()
while IFS= read -r -d '' file; do
    CPL_FILES+=("$file")
done < <(find "${EXAMPLE_DIR}" -maxdepth 1 -name '*.cpl' -print0)

if [ ${#CPL_FILES[@]} -eq 0 ]; then
    echo "Error: No .cpl files found in '${EXAMPLE_DIR}'."
    exit 1
fi

# Prepend core.cpl to the CPL file list if it exists
if [ -n "${CORE_CPL}" ]; then
    CPL_FILES=("${CORE_CPL}" "${CPL_FILES[@]}")
fi

# Check for at least one .parquet file
PARQUET_COUNT=$(find "${EXAMPLE_DIR}" -maxdepth 1 -name '*.parquet' | wc -l)
if [ "${PARQUET_COUNT}" -eq 0 ]; then
    echo "Error: No .parquet files found in '${EXAMPLE_DIR}'."
    exit 1
fi

OUTPUT_DIR="generated/cmd/${DIR_NAME}"
OUTPUT_FILE="${OUTPUT_DIR}/main.go"

mkdir -p "${OUTPUT_DIR}"

echo "Compiling CPL files from '${EXAMPLE_PARENT_DIR}' (example: ${DIR_NAME})..."
echo "  IO spec: ${IOSPEC_FILE}"
echo "  Output:  ${OUTPUT_FILE}"
echo "  CPL files:"
printf '    %s\n' "${CPL_FILES[@]}"

# Run the CPL compiler
if ! go run src/main/go/cli/compiler/cpl_compiler.go \
    -io "${IOSPEC_FILE}" \
    -o "${OUTPUT_FILE}" \
    "${CPL_FILES[@]}"; then
    echo ""
    echo "Compilation failed. See errors above."
    exit 1
fi

echo ""
echo "Compilation successful. Running generated program..."
echo ""

# Run the generated program
go run "${OUTPUT_FILE}" -data-dir="${EXAMPLE_DIR}"

echo ""
echo "Verifying output against expected results..."
echo ""

EXPECTED_XLSX="${EXAMPLE_DIR}/expected_output.xlsx"

if [ -f "${EXPECTED_XLSX}" ]; then
    # Discover output metric names from iospec.json
    if command -v jq >/dev/null 2>&1; then
        METRIC_NAMES=$(jq -r '.outputs[].metric' "${IOSPEC_FILE}")
    else
        echo "Error: jq is required to parse iospec.json but was not found."
        exit 1
    fi

    if [ -z "${METRIC_NAMES}" ]; then
        echo "Error: No output metrics found in iospec.json."
        exit 1
    fi

    for METRIC in ${METRIC_NAMES}; do
        OUTPUT_PARQUET="${METRIC}.parquet"
        if [ ! -f "${OUTPUT_PARQUET}" ]; then
            echo "Error: Output parquet ${OUTPUT_PARQUET} not found."
            exit 1
        fi

        echo "Verifying metric '${METRIC}' (${OUTPUT_PARQUET})..."
        if ! go run src/main/go/cli/xls_verify/main.go \
            -parquet="${OUTPUT_PARQUET}" \
            -xlsx="${EXPECTED_XLSX}" \
            -metric="${METRIC}"; then
            echo ""
            echo "Verification FAILED for metric ${METRIC}. See differences above."
            exit 1
        fi
        echo "Verification PASSED for metric '${METRIC}'."
    done
    echo ""
    echo "All metrics verified successfully."
else
    echo "Warning: expected_output.xlsx not found, skipping verification."
fi
