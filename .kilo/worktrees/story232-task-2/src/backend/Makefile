.PHONY: test test-verbose clean generate fetch-db

FETCH_DB_PORT ?= 5438

# When switching to the ubuntu user, the default PATH may not include go.
# Prepending /usr/local/go/bin is harmless in environments where go lives
# elsewhere (e.g. local Docker, GitHub CI), but lets Claude Code's container
# (where go is installed at /usr/local/go/bin but missing from ubuntu's PATH)
# find the binary.
UBUNTU_ENV := PATH=/usr/local/go/bin:$$PATH GOPATH=/home/ubuntu/go GOCACHE=/home/ubuntu/.cache/go-build

# Default target
all: generate

# Generates Go-Jet models from schema.sql
generate: serde/rdbms/gen

serde/rdbms/gen: serde/rdbms/schema.sql
	bash ../../scripts/generate-models.sh
	touch serde/rdbms/gen

# Pre-fetches and extracts the embedded postgres binary to avoid race conditions in parallel tests
fetch-db:
	@if [ "$$(id -u)" = "0" ] && id ubuntu >/dev/null 2>&1; then \
		su ubuntu -c "$(UBUNTU_ENV) go run ./cmd/dbtool/ --port $(FETCH_DB_PORT) || true"; \
	else \
		go run ./cmd/dbtool/ --port $(FETCH_DB_PORT) || true; \
	fi

# Runs all tests in the backend package.
# Detects if running as root/with su available (dev container) or as a normal user (CI).
test: generate fetch-db
	@if [ "$$(id -u)" = "0" ] && id ubuntu >/dev/null 2>&1; then \
		su ubuntu -c "$(UBUNTU_ENV) go test ./..."; \
	else \
		go test ./...; \
	fi

# Runs all tests with verbose output
test-verbose: generate fetch-db
	@if [ "$$(id -u)" = "0" ] && id ubuntu >/dev/null 2>&1; then \
		su ubuntu -c "$(UBUNTU_ENV) go test -v ./..."; \
	else \
		go test -v ./...; \
	fi

# Cleans the go build cache and generated models
clean:
	rm -rf serde/rdbms/gen
	@if [ "$$(id -u)" = "0" ] && id ubuntu >/dev/null 2>&1; then \
		chgrp -R ubuntu /Chronocone/scripts /Chronocone/src/backend; \
		chmod -R 775 /Chronocone/scripts /Chronocone/src/backend; \
		su ubuntu -c "$(UBUNTU_ENV) go clean -testcache"; \
	else \
		go clean -testcache; \
	fi
