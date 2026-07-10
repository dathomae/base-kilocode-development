#!/bin/bash
#
# metadata_create.sh - Generic wrapper for creating metadata entities
#
# This script provides a generic wrapper around the CLI 'create' command,
# allowing creation of any metadata type by specifying it as the first argument.
#
# Usage: ./metadata_create.sh <subcommand> [OPTIONS]
#   subcommand                 The type of metadata to create:
#                              namespace-family, namespace, category, element,
#                              metric, transform, ts-doc-type
#   [OPTIONS]                  Options specific to the subcommand
#   --api-url URL              Base URL of the metrics service API (default: http://localhost:8001)
#
# Examples:
#   ./metadata_create.sh namespace-family --name "MyFamily" --description "My family"
#   ./metadata_create.sh namespace --name "Production" --namespace-family "AppFamily" --description "Prod"
#   ./metadata_create.sh category --name "Sales" --namespace "Production" --description "Sales data"
#   ./metadata_create.sh element --name "Revenue" --category "Sales" --namespace "Production" --description "Revenue"
#   ./metadata_create.sh metric --name "TotalRevenue" --concept-id 5 --description "Total revenue"
#   ./metadata_create.sh transform --name "Sum" --storage-config-id 1 --code-location "/path"
#   ./metadata_create.sh ts-doc-type --name "Daily" --namespace "Production" --description "Daily docs"
#

set -euo pipefail

# Check if at least one argument (subcommand) is provided
if [[ $# -lt 1 ]]; then
    echo "Error: Subcommand is required" >&2
    echo "Usage: $0 <subcommand> [options]" >&2
    echo "Subcommands: namespace-family, namespace, category, element, metric, transform, ts-doc-type" >&2
    exit 1
fi

# Extract the subcommand
SUBCOMMAND="$1"
shift

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Navigate to project root (one level up from shell_scripts)
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Change to the project root directory where the cli package is located
cd "${PROJECT_ROOT}"

# Run the CLI command, passing the subcommand and remaining arguments
python -m cli.main create "${SUBCOMMAND}" "$@"

# Exit with the Python script's exit code
exit $?
