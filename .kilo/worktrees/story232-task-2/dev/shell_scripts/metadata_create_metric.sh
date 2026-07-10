#!/bin/bash
#
# metadata_create_metric.sh - Create a new metric
#
# This script wraps the CLI 'create metric' command to create
# a new metric, which defines measurable quantities associated
# with concepts (elements).
#
# Usage: ./metadata_create_metric.sh [OPTIONS]
#   --name TEXT                Name of the metric (required)
#   --element TEXT             Name of the element (concept) this metric references (required)
#   --description TEXT         Description of the metric (required)
#   --concept-id ID            Direct element/concept ID (alternative to element name)
#   --api-url URL              Base URL of the metrics service API (default: http://localhost:8001)
#
# Examples:
#   ./metadata_create_metric.sh --name "TotalRevenue" --element "Revenue" --description "Total revenue metric"
#   ./metadata_create_metric.sh --name "DailyRevenue" --concept-id 5 --description "Daily revenue metric"
#

set -euo pipefail

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Navigate to project root (one level up from shell_scripts)
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Change to the project root directory where the cli package is located
cd "${PROJECT_ROOT}"

# Run the CLI command, passing all arguments through
python -m cli.main create metric "$@"

# Exit with the Python script's exit code
exit $?
