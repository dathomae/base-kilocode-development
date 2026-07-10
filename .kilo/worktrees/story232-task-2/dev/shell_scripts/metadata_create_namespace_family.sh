#!/bin/bash
#
# metadata_create_namespace_family.sh - Create a new namespace family
#
# This script wraps the CLI 'create namespace-family' command to create
# a new namespace family, which is the top-level organizational unit
# that groups related namespaces together.
#
# Usage: ./metadata_create_namespace_family.sh [OPTIONS]
#   --name TEXT                Name of the namespace family (required)
#   --description TEXT         Description of the namespace family (required)
#   --tooltip TEXT             Tooltip for the namespace family (required)
#   --customer-subject-id ID   Customer subject ID (default: 1)
#   --owner-subject-id ID      Owner subject ID (default: 1)
#   --ns-fam-type-id ID        Namespace family type ID (default: 1)
#   --api-url URL              Base URL of the metrics service API (default: http://localhost:8001)
#
# Examples:
#   ./metadata_create_namespace_family.sh --name "MyFamily" --description "My family description" --tooltip "Family tooltip"
#   ./metadata_create_namespace_family.sh --name "AppFamily" --description "Application metrics family" --tooltip "App tooltip" --owner-subject-id 2
#

set -euo pipefail

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Navigate to project root (one level up from shell_scripts)
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Change to the project root directory where the cli package is located
cd "${PROJECT_ROOT}"

# Run the CLI command, passing all arguments through
python -m cli.main create namespace-family "$@"

# Exit with the Python script's exit code
exit $?
