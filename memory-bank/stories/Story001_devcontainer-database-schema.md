# Story001: Devcontainer Database & Schema

## Goal

Add a PostgreSQL database service to the existing devcontainer and create the database schema needed for the Polymarket exploration. This includes updating Docker Compose and Dockerfile configurations, adding the `psycopg2-binary` dependency, creating the schema DDL, a database connection module, and query helper functions.

This story implements [Story Plan: Devcontainer Database & Schema](../plans/story-devcontainer-database-schema.md).

## References

- [Story Plan: Devcontainer Database & Schema](../plans/story-devcontainer-database-schema.md)
- [Polymarket Exploration Plan](../plans/polymarket-exploration-plan.md)

## Dependencies

None — This is the foundational infrastructure story for the Polymarket exploration.

## Dependent Stories

None — No stories have been created yet.

## Tasks

### Task Sizing

Tasks 1–4 are straightforward single-file changes: each modifies or creates one file, follows an existing pattern, and addresses a single concern. Tasks 1–3 modify existing files with well-defined structure; Task 4 creates a new DDL file. All four qualify as `[Small]`.

Task 5 spans 2 files (source + test) with 9 subtasks following a clean TDD pattern — tests mock `psycopg2` internals so no running database is required. The 2-file boundary and tight test-implementation coupling place it at the top of the Small range. It qualifies as `[Small]` because the implementation is mechanically straightforward (env var reading, logging, context manager, file I/O).

Task 6 spans 2 files (source + test) with 12 subtasks across 5 distinct query patterns plus logging instrumentation. The architect-for-story-planning flagged it as borderline: while it does not trigger multiple "too large" indicators, the 12-subtask count and 5 distinct SQL patterns benefit from the broader-context reasoning of the more capable `code-for-story-implementor` agent. It is **not** annotated `[Small]`. The architect confirmed the task is at its natural granularity — splitting would create merge conflicts on both `queries.py` and `test_queries.py` with no isolation benefit.

### Parallel Execution

- **Group A** (Tasks 1, 2, 3, 4, 5): All five tasks have completely disjoint file-modification sets. No file appears in more than one task's change set. Task 5's `init_db()` references `schema.sql` (from Task 4) by path, but tests mock both `__file__` resolution and the database cursor — no actual `schema.sql` file is needed at test time. All five worktrees merge cleanly because file change sets are fully non-overlapping.

### Execution Order

```
[Task 1 ‖ Task 2 ‖ Task 3 ‖ Task 4 ‖ Task 5] → Task 6
```

Task 6 depends on `db/__init__.py` (created by Task 5) as a module import — it must run sequentially after the Group A merge.

### Reintegration

**Group A (Tasks 1, 2, 3, 4, 5):**

- **Merge order**: any order (fully disjoint files across different directories; no source file is modified by more than one task)
- **Integration tests**:
  1. `uv sync` — Task 3 adds `psycopg2-binary` to `pyproject.toml`; after merge, `uv.lock` must be regenerated to reflect the new dependency. Run this before any `uv run` commands against the merged tree.
  2. `docker compose -f .devcontainer/docker-compose.yml config` — verify the combined docker-compose (Task 1) parses correctly with both `app` and `db` services defined.
  3. `docker build -f .devcontainer/Dockerfile .` — verify the combined Dockerfile (Task 2) builds successfully with `libpq-dev` added.
  4. `uv run pytest tests/db/ -v -x` — run all tests in the `tests/db/` directory to verify the connection module (Task 5) works in the merged state with updated dependencies.
- **Watch for conflicts in**:
  - `uv.lock` — Not directly modified by any task, but Task 3's worktree may regenerate it against the modified `pyproject.toml`. The merged state will have updated `pyproject.toml` but possibly stale `uv.lock`. The `uv sync` command (integration test #1) resolves this. Inspect `uv.lock` after merge and regenerate if needed before running Python tests.
  - `.devcontainer/devcontainer.json` — Not modified by any task, but sits alongside `docker-compose.yml` and `Dockerfile` which are both modified. Verify it still references the correct compose file and service names post-merge (low risk)._

---

### Task 1: Add db service to docker-compose.yml — Not Started [Small]

Update `.devcontainer/docker-compose.yml` to add a PostgreSQL `db` service and inject database connection environment variables into the `app` service.

a. Add `postgres-data` named volume to the top-level `volumes` section — Not Started
b. Add a `db` service with `image: postgres:15`, environment variables (`POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` all set to `prediction_markets`), a volume mount for `postgres-data:/var/lib/postgresql/data`, port mapping `"5432"`, `restart: unless-stopped`, and the `dev-network` network — Not Started
c. Add `PREDICTION_MARKETS_DB_HOST=db`, `PREDICTION_MARKETS_DB_PORT=5432`, `PREDICTION_MARKETS_DB_USER=prediction_markets`, `PREDICTION_MARKETS_DB_PASSWORD=prediction_markets`, and `PREDICTION_MARKETS_DB_NAME=prediction_markets` environment variables to the `app` service — Not Started

**Acceptance Criteria:**

- `docker compose -f .devcontainer/docker-compose.yml config` shows both `app` and `db` services
- `db` service uses `postgres:15` image with correct environment variables
- `app` service has all five `PREDICTION_MARKETS_DB_*` environment variables
- `postgres-data` named volume is defined
- Both services share the `dev-network` network

### Task 2: Add libpq-dev to Dockerfile — Not Started [Small]

Update `.devcontainer/Dockerfile` to include the `libpq-dev` package required to build `psycopg2-binary`.

a. Write a manual verification that `libpq-dev` can be installed alongside the existing packages by checking the apt package name — Not Started
b. Add `libpq-dev` to the package list in the `apt-get install` command on line 7, alphabetically between `jq` and `iputils-ping` — Not Started

**Acceptance Criteria:**

- `libpq-dev` appears in the apt package list in the Dockerfile
- Existing packages remain unchanged and in alphabetical order
- `docker build -f .devcontainer/Dockerfile .` succeeds (or runs without `libpq-dev`-related errors)

### Task 3: Add psycopg2-binary dependency — Not Started [Small]

Update `pyproject.toml` to add `psycopg2-binary` to the project dependencies.

a. Add `"psycopg2-binary"` to the `dependencies` list under `[project]` — Not Started

**Acceptance Criteria:**

- `dependencies = ["psycopg2-binary"]` in `pyproject.toml`
- `uv sync` completes successfully and resolves `psycopg2-binary`

### Task 4: Create database schema DDL — Not Started [Small]

Create `src/prediction_markets/db/schema.sql` with the three-table schema for markets, order book snapshots, and token prices.

a. Write a manual check to verify the `src/prediction_markets/db/` directory exists (create it if needed) — Not Started
b. Create the `markets` table with columns: `id TEXT PRIMARY KEY`, `condition_id TEXT NOT NULL UNIQUE`, `question TEXT NOT NULL`, `slug TEXT NOT NULL`, `token_id_yes TEXT NOT NULL UNIQUE`, `token_id_no TEXT NOT NULL UNIQUE`, `fee_rate REAL NOT NULL`, `category TEXT`, `end_date TIMESTAMPTZ`, `volume REAL`, `liquidity REAL`, `neg_risk BOOLEAN NOT NULL DEFAULT FALSE`, `enable_order_book BOOLEAN NOT NULL DEFAULT TRUE`, `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`, `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()` — Not Started
c. Create the `order_book_snapshots` table with columns: `id BIGSERIAL PRIMARY KEY`, `market_id TEXT NOT NULL REFERENCES markets(id)`, `timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()`, `best_bid_yes REAL NOT NULL`, `best_ask_yes REAL NOT NULL`, `best_bid_no REAL NOT NULL`, `best_ask_no REAL NOT NULL`. Add `UNIQUE (market_id, timestamp)` constraint, index on `timestamp`, and index on `(market_id, timestamp)` — Not Started
d. Create the `token_prices` table with columns: `id BIGSERIAL PRIMARY KEY`, `market_id TEXT NOT NULL REFERENCES markets(id)`, `token_id TEXT NOT NULL`, `timestamp TIMESTAMPTZ NOT NULL`, `price REAL NOT NULL`. Add `UNIQUE (market_id, token_id, timestamp)` constraint, index on `timestamp`, and index on `(market_id, timestamp)` — Not Started
e. Wrap each `CREATE TABLE` statement with `CREATE TABLE IF NOT EXISTS` for idempotency — Not Started

**Acceptance Criteria:**

- `schema.sql` exists at `src/prediction_markets/db/schema.sql`
- Contains `CREATE TABLE IF NOT EXISTS` for `markets`, `order_book_snapshots`, and `token_prices`
- All columns, constraints, foreign keys, and indexes match the plan specification

### Task 5: Create database connection module with tests — Not Started [Small]

Create `src/prediction_markets/db/__init__.py` with connection helper functions, and `tests/db/test_connection.py` with unit tests that mock the database connection.

a. Write test `test_get_connection_reads_env_vars` in `tests/db/test_connection.py` that patches environment variables with `PREDICTION_MARKETS_DB_*` values, calls `get_connection()`, and asserts that `psycopg2.connect` was called with the correct parameters — Not Started
b. Write test `test_context_manager_commits_and_closes` that uses the context manager, verifies `commit()` and `close()` are called on the connection — Not Started
c. Write test `test_init_db_idempotent` that patches `__file__` to resolve `schema.sql`, mocks a connection cursor, calls `init_db()` twice, and asserts that schema SQL was executed exactly once (cursor.execute called twice — each `CREATE TABLE IF NOT EXISTS` is a separate execute call — but the function ran successfully both times) — Not Started
d. Write test `test_get_connection_raises_on_missing_env_vars` that clears one `PREDICTION_MARKETS_DB_*` environment variable at a time and asserts `get_connection()` raises a clear error whose message includes the name of the missing variable — Not Started
e. Write test `test_init_db_logs_schema_path` that patches `__file__`, mocks a connection cursor and the module logger, calls `init_db()`, and asserts an `INFO`-level log message was emitted containing the resolved schema file path — Not Started
f. Implement `get_connection()` in `src/prediction_markets/db/__init__.py` — imports `logging`, creates a module-level logger via `logging.getLogger(__name__)`, reads `PREDICTION_MARKETS_DB_HOST`, `PREDICTION_MARKETS_DB_PORT`, `PREDICTION_MARKETS_DB_USER`, `PREDICTION_MARKETS_DB_PASSWORD`, `PREDICTION_MARKETS_DB_NAME` from environment, raises a clear error (with `logger.error(...)`) naming the missing variable if any are absent, logs connection parameters at `INFO` level (host, port, dbname — redact password), and returns a `psycopg2.connect()` connection — Not Started
g. Implement a `@contextmanager` in `__init__.py` that yields a connection from `get_connection()`, calls `conn.commit()` on successful exit, and calls `conn.close()` on both success and failure — Not Started
h. Implement `init_db()` in `__init__.py` that locates `schema.sql` relative to `__file__` (the `__init__.py` file's directory), logs the resolved path at `INFO` level, reads it, splits on semicolons, executes each non-empty statement against a connection, and logs the number of executed statements at `INFO` level — Not Started
i. Run `uv run pytest tests/db/test_connection.py -v` and verify all tests pass — Not Started

**Acceptance Criteria:**

- `get_connection()` reads from `PREDICTION_MARKETS_DB_*` env vars and raises clear error on missing vars
- Error messages name the specific missing environment variable
- `get_connection()` logs connection parameters at `INFO` level (password redacted) and logs errors via `logger.error(...)`
- Context manager commits on success, closes connection on both success and failure
- `init_db()` logs the resolved schema file path and statement count at `INFO` level
- `init_db()` locates `schema.sql` relative to module path and executes DDL statements
- `init_db()` is safe to call multiple times (idempotent via `IF NOT EXISTS`)
- All unit tests pass with mocked database connections

### Task 6: Create query helper functions with tests — Not Started

Create `src/prediction_markets/db/queries.py` with insert/upsert and select functions, and `tests/db/test_queries.py` with unit tests that mock database cursors.

a. Write test `test_insert_markets_upsert` in `tests/db/test_queries.py` that passes a list of market dicts, verifies `INSERT ... ON CONFLICT ... UPDATE` SQL is generated with correct columns, and verifies rows are passed as parameters — Not Started
b. Write test `test_insert_order_book_snapshot_conflict_do_nothing` that passes snapshot data, verifies `INSERT ... ON CONFLICT DO NOTHING` SQL, and asserts rows are passed as parameters — Not Started
c. Write test `test_insert_token_prices_conflict_do_nothing` that passes a list of price dicts, verifies `INSERT ... ON CONFLICT DO NOTHING` SQL, and asserts rows are passed as parameters — Not Started
d. Write test `test_get_markets_returns_list_of_dicts` that mocks a cursor returning `description` and rows, calls `get_markets()`, and verifies the returned list of dicts matches the mocked rows — Not Started
e. Write test `test_get_market_by_id_returns_single_dict` that mocks a cursor returning one row, calls `get_market_by_id()`, and verifies the returned dict matches — Not Started
f. Write test `test_query_functions_log_at_debug_level` that patches the module logger and verifies each query function emits a `DEBUG`-level log message containing the function name on call — Not Started
g. Implement `insert_markets(conn, markets: list[dict])` in `src/prediction_markets/db/queries.py` — imports `logging`, creates a module-level logger, logs a `DEBUG`-level message with the function name and market count, uses parameterized `INSERT ... ON CONFLICT (id) DO UPDATE SET ...` with all market columns — Not Started
h. Implement `insert_order_book_snapshot(conn, market_id, best_bid_yes, best_ask_yes, best_bid_no, best_ask_no, timestamp)` — logs `DEBUG`-level message, uses parameterized `INSERT ... ON CONFLICT (market_id, timestamp) DO NOTHING` — Not Started
i. Implement `insert_token_prices(conn, rows: list[dict])` — logs `DEBUG`-level message with row count, uses parameterized `INSERT ... ON CONFLICT (market_id, token_id, timestamp) DO NOTHING`, iterates over rows to build the multi-row INSERT — Not Started
j. Implement `get_markets(conn)` — logs `DEBUG`-level message, executes `SELECT * FROM markets` and returns a list of dicts keyed by column name — Not Started
k. Implement `get_market_by_id(conn, market_id)` — logs `DEBUG`-level message, executes `SELECT * FROM markets WHERE id = %s` with parameter, returns a single dict — Not Started
l. Run `uv run pytest tests/db/test_queries.py -v` and verify all tests pass — Not Started

**Acceptance Criteria:**

- All query functions use parameterized queries (no string interpolation of values)
- `insert_markets()` performs UPSERT (insert or update on conflict)
- `insert_order_book_snapshot()` and `insert_token_prices()` silently ignore duplicate conflicts
- `get_markets()` returns a list of dicts; `get_market_by_id()` returns a single dict
- Each query function emits a `DEBUG`-level log message on call via `logging.getLogger(__name__)`
- All unit tests pass with mocked database cursors

## Test-First Development

This project follows the Logical TDD Lifecycle. For Tasks 5 and 6, tests are written before implementation. Test subtasks (a-e for Task 5, a-f for Task 6) precede implementation subtasks. Tests for `get_connection()` and query functions use `unittest.mock` to patch `psycopg2` internals so no running database is required during unit testing. Logging behavior is tested by patching the module logger and asserting log messages were emitted at the expected levels.

## Constraints

- No ORM — raw SQL via `psycopg2` only
- All queries must use parameterized statements (never string interpolation for values)
- `init_db()` must be idempotent — safe to call multiple times
- `schema.sql` must be locatable via `__file__`-relative path regardless of working directory
- Docker Compose configuration must preserve the existing `app` service structure
- Database credentials are for development only and may be hardcoded in the devcontainer configuration
- Logging: use Python's standard `logging` module with `logging.getLogger(__name__)` in each module; `INFO` level for lifecycle events (connection, schema loading), `DEBUG` level for per-operation tracing (query function calls). Passwords must never appear in log messages.

## Notes

- The `source_root` for this project is the project root (`/prediction-markets`). All file paths in task descriptions are relative to `source_root`. Commands like `uv run pytest` run from `source_root`.
- The plan calls this "Story 1" in the Polymarket Exploration Plan; it has no dependencies and can run in parallel with Story 2 (Market Interface Module, not yet created).
- Environment variables follow the `PREDICTION_MARKETS_` prefix convention for the app service, while the `db` service uses standard PostgreSQL env vars (`POSTGRES_*`).
