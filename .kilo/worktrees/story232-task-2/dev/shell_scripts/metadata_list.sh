#!/bin/bash
#
# metadata_list.sh - List metadata in hierarchical tree format
#
# This script wraps the CLI 'list' command to display metadata
# definitions in a tree structure showing the hierarchy:
# namespace_family -> namespace -> category -> element
#
# Usage: ./metadata_list.sh [OPTIONS]
#   --namespace-family TEXT    Filter by namespace family name
#   --namespace TEXT           Filter by namespace name
#   --category TEXT            Filter by category name
#   --type [namespace_family|namespace|category|element|metric|transform|ts_doc_type]
#                              Filter by metadata type
#   --api-url URL              Base URL of the metrics service API (default: http://localhost:8001)
#
# Examples:
#   ./metadata_list.sh
#   ./metadata_list.sh --type category
#   ./metadata_list.sh --namespace-family AppFamily --namespace Production
#

set -euo pipefail

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Navigate to project root (one level up from shell_scripts)
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Change to the project root directory where the cli package is located
cd "${PROJECT_ROOT}"

# Run the CLI command, passing all arguments through
python -m cli.main list "$@"

# Exit with the Python script's exit code
exit $?
