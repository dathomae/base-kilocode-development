#!/bin/bash
#
# metadata_create_ts_doc_type.sh - Create a new time series document type
#
# This script wraps the CLI 'create ts-doc-type' command to create
# a new time series document type, which defines the structure of
# time series documents including schema and format.
#
# Usage: ./metadata_create_ts_doc_type.sh [OPTIONS]
#   --name TEXT                Name of the time series document type (required)
#   --namespace TEXT           Name of the namespace (required)
#   --description TEXT         Description of the document type (required)
#   --api-url URL              Base URL of the metrics service API (default: http://localhost:8001)
#
# Examples:
#   ./metadata_create_ts_doc_type.sh --name "DailySales" --namespace "Production" --description "Daily sales data"
#   ./metadata_create_ts_doc_type.sh --name "HourlyMetrics" --namespace "Production" --description "Hourly aggregated metrics"
#

set -euo pipefail

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Navigate to project root (one level up from shell_scripts)
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Change to the project root directory where the cli package is located
cd "${PROJECT_ROOT}"

# Run the CLI command, passing all arguments through
python -m cli.main create ts-doc-type "$@"

# Exit with the Python script's exit code
exit $?
