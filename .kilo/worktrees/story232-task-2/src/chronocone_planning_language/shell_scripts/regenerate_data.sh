#!/bin/bash
set -e

# Usage: ./regenerate_data.sh <directory_name> [example_base_dir]
#
# Where <directory_name> is a subdirectory under the example base directory
# (default: examples/with_data). The subdirectory must contain one or more
# data_<name>.json files in row-format (JSON array of objects with string
# columns and a "Value" float64 field).
#
# For each data_<name>.json found, this script invokes parquet_gen to produce:
#   data_<name>.parquet   -- the generated parquet file

if [ $# -lt 1 ] || [ $# -gt 2 ]; then
    echo "Usage: $0 <directory_name> [example_base_dir]"
    echo "  <directory_name>    subdirectory (e.g., simple_addition)"
    echo "  [example_base_dir]  base directory (default: examples/with_data)"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "${SCRIPT_DIR}/.."

DIR_NAME="$1"
EXAMPLE_BASE_DIR="${2:-examples/with_data}"
EXAMPLE_DIR="${EXAMPLE_BASE_DIR}/${DIR_NAME}"

if [ ! -d "${EXAMPLE_DIR}" ]; then
    echo "Error: Directory '${EXAMPLE_DIR}' does not exist."
    exit 1
fi

# Collect all data_*.json files in the example subdirectory
DATA_FILES=()
while IFS= read -r -d '' file; do
    DATA_FILES+=("$file")
done < <(find "${EXAMPLE_DIR}" -maxdepth 1 -name 'data_*.json' -print0)

if [ ${#DATA_FILES[@]} -eq 0 ]; then
    echo "Error: No data_*.json files found in '${EXAMPLE_DIR}'."
    exit 1
fi

echo "Regenerating parquet files from '${EXAMPLE_DIR}'..."
echo "  Data files:"
printf '    %s\n' "${DATA_FILES[@]}"
echo ""

for data_file in "${DATA_FILES[@]}"; do
    basename_file="$(basename "${data_file}")"
    # Extract <name> from data_<name>.json
    if [[ "${basename_file}" =~ ^data_(.*)\.json$ ]]; then
        name="${BASH_REMATCH[1]}"
    else
        echo "Warning: Could not parse name from '${basename_file}', skipping."
        continue
    fi

    output_parquet="${EXAMPLE_DIR}/data_${name}.parquet"

    echo "Generating data_${name}.parquet ..."
    echo "  Input:  ${data_file}"
    echo "  Output: ${output_parquet}"

    if ! go run src/main/go/cli/parquetgen/parquet_gen.go \
        -input="${data_file}" \
        -output="${output_parquet}"; then
        echo ""
        echo "Error: Failed to generate parquet for '${basename_file}'."
        exit 1
    fi
    echo "  Done."
    echo ""
done

echo "All parquet files regenerated successfully."
