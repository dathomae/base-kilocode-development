#!/bin/bash
#
# metadata_create_category.sh - Create a new category
#
# This script wraps the CLI 'create category' command to create
# a new category, which provides hierarchical classifications
# within a namespace. Categories can have parent-child relationships.
#
# Usage: ./metadata_create_category.sh [OPTIONS]
#   --name TEXT                Name of the category (required)
#   --namespace TEXT           Name of the parent namespace (required)
#   --description TEXT         Description of the category (required)
#   --parent-category-id ID    ID of the parent category (for sub-categories)
#   --api-url URL              Base URL of the metrics service API (default: http://localhost:8001)
#
# Examples:
#   ./metadata_create_category.sh --name "Sales" --namespace "Production" --description "Sales data"
#   ./metadata_create_category.sh --name "SubCategory" --namespace "Production" --description "Sub-category" --parent-category-id 5
#

set -euo pipefail

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Navigate to project root (one level up from shell_scripts)
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Change to the project root directory where the cli package is located
cd "${PROJECT_ROOT}"

# Run the CLI command, passing all arguments through
python -m cli.main create category "$@"

# Exit with the Python script's exit code
exit $?
