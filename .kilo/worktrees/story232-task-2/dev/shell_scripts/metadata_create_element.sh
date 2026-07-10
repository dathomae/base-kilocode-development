#!/bin/bash
#
# metadata_create_element.sh - Create a new element
#
# This script wraps the CLI 'create element' command to create
# a new element, which is an atomic building block within a category.
# Elements represent the leaf nodes in the taxonomy hierarchy.
#
# Usage: ./metadata_create_element.sh [OPTIONS]
#   --name TEXT                Name of the element (required)
#   --category TEXT            Name of the parent category (required)
#   --namespace TEXT           Name of the namespace containing the category (required)
#   --description TEXT         Description of the element (required)
#   --api-url URL              Base URL of the metrics service API (default: http://localhost:8001)
#
# Examples:
#   ./metadata_create_element.sh --name "Revenue" --category "Sales" --namespace "Production" --description "Revenue metric element"
#   ./metadata_create_element.sh --name "UnitsSold" --category "Sales" --namespace "Production" --description "Units sold element"
#

set -euo pipefail

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Navigate to project root (one level up from shell_scripts)
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Change to the project root directory where the cli package is located
cd "${PROJECT_ROOT}"

# Run the CLI command, passing all arguments through
python -m cli.main create element "$@"

# Exit with the Python script's exit code
exit $?
