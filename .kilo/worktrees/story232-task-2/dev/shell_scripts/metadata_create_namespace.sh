#!/bin/bash
#
# metadata_create_namespace.sh - Create a new namespace
#
# This script wraps the CLI 'create namespace' command to create
# a new namespace, which provides isolation and organization for
# metrics and categories within a namespace family.
#
# Usage: ./metadata_create_namespace.sh [OPTIONS]
#   --name TEXT                Name of the namespace (required)
#   --namespace-family TEXT    Name of the parent namespace family (required)
#   --description TEXT         Description of the namespace (required)
#   --owner-subject-id ID      Owner subject ID (default: 1)
#   --ns-type-id ID            Namespace type ID (default: 1)
#   --api-url URL              Base URL of the metrics service API (default: http://localhost:8001)
#
# Examples:
#   ./metadata_create_namespace.sh --name "Production" --namespace-family "AppFamily" --description "Production environment"
#   ./metadata_create_namespace.sh --name "Staging" --namespace-family "AppFamily" --description "Staging environment" --ns-type-id 2
#

set -euo pipefail

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Navigate to project root (one level up from shell_scripts)
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Change to the project root directory where the cli package is located
cd "${PROJECT_ROOT}"

# Run the CLI command, passing all arguments through
python -m cli.main create namespace "$@"

# Exit with the Python script's exit code
exit $?
