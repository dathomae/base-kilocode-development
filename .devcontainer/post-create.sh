#!/bin/bash
set -e
echo "=== Bootstrapping Python Development Environment ==="
cd "$1"
uv sync
echo "=== Bootstrap Complete! ==="
