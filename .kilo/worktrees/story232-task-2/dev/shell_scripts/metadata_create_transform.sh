#!/bin/bash
#
# metadata_create_transform.sh - Create a new transform
#
# This script wraps the CLI 'create transform' command to create
# a new transform, which defines data transformations for metrics.
# Transforms specify where the transformation code is located.
#
# Usage: ./metadata_create_transform.sh [OPTIONS]
#   --name TEXT                Name of the transform (required)
#   --storage-config-id ID     Storage configuration ID (required)
#   --code-location PATH       Path or reference to transformation code (required)
#   --description TEXT         Description of the transform (default: "")
#   --api-url URL              Base URL of the metrics service API (default: http://localhost:8001)
#
# Examples:
#   ./metadata_create_transform.sh --name "SumTransform" --storage-config-id 1 --code-location "/path/to/code"
#   ./metadata_create_transform.sh --name "AvgTransform" --storage-config-id 1 --code-location "/transforms/avg.py" --description "Average calculation"
#

set -euo pipefail

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Navigate to project root (one level up from shell_scripts)
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Change to the project root directory where the cli package is located
cd "${PROJECT_ROOT}"

# Run the CLI command, passing all arguments through
python -m cli.main create transform "$@"

# Exit with the Python script's exit code
exit $?
