#!/bin/bash
#
# Start ChronoCone Backend Services
#
# This script starts all ChronoCone backend services:
# - Core Service (port 8000): Service registry and routing
# - Metrics Service (port 8001): Metadata management (auto-started by Core Service)
#
# Usage: ./start_services.sh [options]
#   --host HOST       Host to bind the Core Service to (default: 0.0.0.0)
#   --port PORT       Port to bind the Core Service to (default: 8000)
#   --no-autostart    Disable automatic start of dependent services
#

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "${SCRIPT_DIR}/../src/backend" && pwd)"
HOST="0.0.0.0"
PORT="8000"
AUTOSTART="--autostart"

# Parse command-line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --host)
            HOST="$2"
            shift 2
            ;;
        --port)
            PORT="$2"
            shift 2
            ;;
        --no-autostart)
            AUTOSTART="--no-autostart"
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --host HOST       Host to bind the Core Service to (default: 0.0.0.0)"
            echo "  --port PORT       Port to bind the Core Service to (default: 8000)"
            echo "  --no-autostart    Disable automatic start of dependent services"
            echo "  --help, -h        Display this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Change to the backend directory
cd "${BACKEND_DIR}"

echo "=========================================="
echo "ChronoCone Backend Services Startup"
echo "=========================================="
echo ""
echo "Configuration:"
echo "  Host:      ${HOST}"
echo "  Port:      ${PORT}"
echo "  Autostart: ${AUTOSTART}"
echo ""

# Verify Python is available
if ! command -v python &> /dev/null; then
    echo "Error: Python is not installed or not in PATH"
    exit 1
fi

echo "Python version: $(python --version)"
echo ""

# Start the Core Service with autostart enabled (starts Metrics Service automatically)
echo "Starting Core Service..."
echo "  - Core Service will be available at: http://${HOST}:${PORT}"
if [[ "${AUTOSTART}" == "--autostart" ]]; then
    echo "  - Metrics Service will be available at: http://${HOST}:8001"
fi
echo ""
echo "Press Ctrl+C to stop all services"
echo ""

# Execute the Core Service start command
# The Core Service will automatically start the Metrics Service when --autostart is enabled
exec python -m src.core.core_service start \
    --host "${HOST}" \
    --port "${PORT}" \
    "${AUTOSTART}"
