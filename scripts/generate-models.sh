#!/bin/bash
set -e

# This script generates Go-Jet models for the ChronoCone backend.
# It starts a temporary embedded Postgres instance, applies the schema,
# runs the jet generator, and then cleans up.

cd "$(dirname "$0")/../src/backend"

echo "Cleaning up any existing database on port 5438..."
pkill -9 -f dbtool || true
pkill -9 postgres || true
rm -f /tmp/.s.PGSQL.5438.lock || true

# Function to run a command as ubuntu if we are root and the user exists
run_as_db_user() {
    local cmd="$1"
    if [ "$(id -u)" = "0" ] && id ubuntu >/dev/null 2>&1; then
        # Only set these paths when we are explicitly switching to the ubuntu user
        su ubuntu -c "export PATH=/usr/local/go/bin:\$PATH GOPATH=/home/ubuntu/go GOCACHE=/home/ubuntu/.cache/go-build && $cmd"
    else
        # In CI or non-root environments, use the existing environment
        bash -c "export PATH=/usr/local/go/bin:\$PATH && $cmd"
    fi
}

echo "Starting temporary database and applying schema..."
run_as_db_user "cd $(pwd) && go run ./cmd/dbtool/ -port 5438 -apply-schema -keep-alive" &
DB_PID=$!

# Give it a moment to fully start
sleep 20

echo "Running Go-Jet generator..."
mkdir -p serde/rdbms/gen
if [ "$(id -u)" = "0" ] && id ubuntu >/dev/null 2>&1; then
    chown -R root:ubuntu serde/rdbms/gen
    chmod -R 775 serde/rdbms/gen
fi

run_as_db_user "cd $(pwd) && go run ./cmd/generator/main.go -path=./serde/rdbms/gen -port=5438"

echo "Stopping temporary database..."
kill $DB_PID || true

echo "Generation complete."
