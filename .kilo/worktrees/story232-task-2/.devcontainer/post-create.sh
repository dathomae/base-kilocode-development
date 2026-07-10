#!/bin/bash

# Abort on first error
set -e

echo "=== Bootstrapping Full-Stack Development Environment ==="


echo "--- 1/2 Setting up Go Backend ---"
cd /tmp

echo "   ...Downloading latest version of Go"
GO_VERSION=$(curl -sSL "https://go.dev/VERSION?m=text" | head -n 1)
curl -sSL "https://go.dev/dl/${GO_VERSION}.linux-amd64.tar.gz" | tar -xz -C /usr/local

cd
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

cd $1/src/backend

echo "   ...Downloading Go dependencies"
# Download Go module dependencies
# Detect latest version and download for Linux (amd64)
/usr/local/go/bin/go mod download

# For debugging
echo "   ...Downloading Go Delve for debugging"
/usr/local/go/bin/go install -v github.com/go-delve/delve/cmd/dlv@latest

echo "--- 2/2 Setting up Node Frontend ---"
cd $1/src/frontend

# Use pnpm to rapidly install frontend packages
pnpm install


echo "=== Bootstrapping Complete! ==="
