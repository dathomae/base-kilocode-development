# Story003: Order Book Snapshot Collector

## Goal

Build the CLI-driven collector that captures order book snapshots from Polymarket at regular intervals. This is the primary data source for the taker analysis — it records actual bid/ask prices for both Yes and No tokens at shared timestamps, with no estimation or approximation. Supports both single-shot execution (`--interval 0`) and continuous cycling at a configurable interval.

This story implements [Story Plan: Order Book Snapshot Collector](../plans/story-order-book-snapshot-collector.md).

## References

- [Story Plan: Order Book Snapshot Collector](../plans/story-order-book-snapshot-collector.md)
- [Polymarket Exploration Plan](../plans/polymarket-exploration-plan.md)

## Dependencies

- [Story001_devcontainer-database-schema.md](Story001_devcontainer-database-schema.md) — provides `db/__init__.py` (connection utilities) and `db/queries.py` (`get_markets`, `insert_order_book_snapshot`) — Not Started
- [Story002_market-interface-module.md](Story002_market-interface-module.md) — provides `market_interfaces/` package (`PolymarketClient`, `RateLimiter`, `MarketData`, `OrderBookData`, `MarketClientError`) — Not Started

## Dependent Stories

None — No stories have been created that depend on this collector. The analysis stories (Phase 1, Step 5) will depend on accumulated snapshot data, but those stories have not yet been created.

## Tasks

### Task Sizing

Tasks 2 and 3 are classified as **medium** (not `[Small]`). Each introduces new infrastructure and requires reasoning about coordinated concerns across test and source files:

- **Task 2** (11 subtasks): Introduces the argparse subparser architecture for the `collectors` CLI — a new package-level abstraction that must be designed for future extension (metadata, prices subcommands). Requires reasoning about argument groups, handler dispatch, and testability of `main()` with mocked handler functions. The architect-for-story-planning evaluated this task: the subparser setup and handler dispatch are one cohesive unit — splitting would create artificial intermediate modules with no isolation benefit.

- **Task 3** (15 subtasks): Implements the core `run(args)` function spanning API calls, rate limiting, database interaction, control flow (single-shot vs. continuous), error handling, graceful shutdown, logging, and market-limit calculation. The architect-for-story-planning evaluated this task: the 13 test subtasks all exercise the same `run()` function from different angles; there is no self-contained preamble; all work is in one source file and one test file. Splitting would require creating artificial helper modules with no isolation benefit. The task is at its natural granularity.

Task 1 is classified as **[Small]**. It creates two trivial file stubs with no logic — purely additive, following Python packaging conventions.

### Parallel Execution

- **Group A** (Tasks 1, 2, 3): All three tasks modify fully disjoint files within the `collectors` package and its test directory. Task 1 creates `__init__.py` and `__main__.py`; Task 2 creates `cli.py` and `tests/collectors/test_cli.py`; Task 3 creates `order_book_snapshots.py` and `tests/collectors/test_order_book_snapshots.py`. Tasks 2 and 3 each mock the other's module in their tests (Task 2 mocks `order_book_snapshots.run`; Task 3 receives a plain `argparse.Namespace`, not the CLI's dispatcher) — neither requires the other's code to be present at test time. All files are additive (no existing files modified) and no file appears in more than one task's change set.

### Execution Order

```
[Task 1 ‖ Task 2 ‖ Task 3] → Task 4
```

Task 4 (integration verification) requires the full `collectors` package to exist — it must run sequentially after the Group A merge. Within Group A, tasks may be merged in any order (fully disjoint, purely additive files).

### Reintegration

**Group A (Tasks 1, 2, 3):**

- **Merge order**: any order — all three tasks create only new files with no shared modification surfaces. No task modifies an existing file, and no file appears in more than one task's change set.
- **Integration tests**: `uv run pytest tests/collectors/ -v --tb=short` — verify all collector unit tests (CLI tests and snapshot collector tests) pass together in the merged tree with correct inter-module imports.
- **Watch for conflicts in**:
  - `tests/collectors/__init__.py` — not explicitly created by any task, but both Task 2 and Task 3 create files in `tests/collectors/`. If a test runner or sub-agent auto-creates a `__init__.py` for test discovery in a worktree, it could appear in the merge. Verify after merge and remove if present (the project's `testpaths` in `pyproject.toml` points to `tests/` which already has an `__init__.py`; nested init files are unnecessary).
  - `src/prediction_markets/collectors/__pycache__/` — bytecode caches may differ per worktree. Not expected to cause conflicts (`__pycache__` is git-ignored). No action needed.

---

### Task 1: Create collectors package files — Not Started [Small]

Create the `collectors` Python package with a package marker and a `__main__.py` entry point so `python -m prediction_markets.collectors` invokes the CLI.

a. Create `src/prediction_markets/collectors/__init__.py` — a package marker containing a docstring that describes the purpose of the collectors package ("Data collectors for Polymarket market metadata, order book snapshots, and price history.") — Not Started

b. Create `src/prediction_markets/collectors/__main__.py` — imports `main` from `prediction_markets.collectors.cli` and invokes it. Contents:
   ```
   """Entry point for python -m prediction_markets.collectors."""
   from prediction_markets.collectors.cli import main
   main()
   ```
   — Not Started

**Acceptance Criteria:**

- `src/prediction_markets/collectors/__init__.py` exists with a descriptive docstring
- `src/prediction_markets/collectors/__main__.py` exists and imports from `.cli`
- The `collectors` directory is a valid Python package (has `__init__.py`)

### Task 2: Create CLI module with argparse subparser architecture — Not Started

Create `src/prediction_markets/collectors/cli.py` with an argparse-based CLI supporting global options and the `order-books` subcommand. The architecture must be designed for future extension (metadata and prices subcommands from the parent plan). Create `tests/collectors/test_cli.py` with unit tests that verify the parser structure, argument parsing, and handler dispatch.

a. Write test `test_create_parser_structure` in `tests/collectors/test_cli.py` — call `create_parser()` and verify it returns an `ArgumentParser`. Verify a subparser named `order-books` exists. Verify the `order-books` subparser accepts `--interval` (type int, default 0), `--market-limit` (type int, default 0), `--rate-limit` (type float, default 1.0), and `--db-url` (type str, default None). — Not Started

b. Write test `test_parse_args_defaults` — call `parse_args(['order-books'])` and verify the returned namespace has `command='order-books'`, `interval=0`, `market_limit=0`, `rate_limit=1.0`, `db_url=None`. — Not Started

c. Write test `test_parse_args_custom_values` — call `parse_args(['order-books', '--interval', '300', '--market-limit', '100', '--rate-limit', '2.5', '--db-url', 'postgresql://localhost/test'])` and verify `interval=300`, `market_limit=100`, `rate_limit=2.5`, `db_url='postgresql://localhost/test'`. — Not Started

d. Write test `test_main_dispatches_to_order_books_handler` — patch `prediction_markets.collectors.order_book_snapshots.run` with a `Mock()`, set `sys.argv` to `['collectors', 'order-books', '--interval', '60']`, call `main()`, and assert the mock was called once with an `argparse.Namespace` whose `interval` attribute equals 60. — Not Started

e. Write test `test_help_exits_cleanly` — call `parse_args(['--help'])` or `parse_args(['order-books', '--help'])` and verify `SystemExit` is raised (or the parser prints help text without error). (Note: argparse's `--help` calls `sys.exit(0)` by default, so the test should catch `SystemExit`.) — Not Started

f. Write test `test_unknown_subcommand_exits_with_error` — call `parse_args(['nonexistent'])` and verify `SystemExit` is raised with a non-zero code. — Not Started

g. Write test `test_rate_limit_accepts_float` — call `parse_args(['order-books', '--rate-limit', '0.5'])` and verify `rate_limit=0.5`. — Not Started

h. Write test `test_missing_subcommand_shows_help` — call `parse_args([])` (no arguments) and verify either `SystemExit` is raised or the returned namespace has no `command` attribute (argparse behavior for missing required subcommand). — Not Started

i. Implement `add_global_args(parser: argparse.ArgumentParser) -> None` in `src/prediction_markets/collectors/cli.py` — adds `--rate-limit` (type float, default 1.0, help describing maximum API requests per second) and `--db-url` (type str, default None, help describing an optional database connection string override) to the provided parser. This helper is called on each subparser to ensure consistent global options across future subcommands. — Not Started

j. Implement `create_parser() -> argparse.ArgumentParser` — creates the root `ArgumentParser` with a project-level description ("Polymarket data collectors — capture market metadata, order book snapshots, and price history."). Adds subparsers with `dest='command'` and `required=True`. Adds an `order-books` subparser via `subparsers.add_parser('order-books', help='Capture order book snapshots at regular intervals')`. Calls `add_global_args(order_books_parser)`. Adds `--interval` (type int, default 0, help describing seconds between cycles; 0 for single-shot) and `--market-limit` (type int, default 0, help describing max markets per cycle; 0 for all). — Not Started

k. Implement `main() -> None` in `cli.py` — calls `create_parser().parse_args()`, checks `args.command`, and if `'order-books'`, imports and calls `order_book_snapshots.run(args)`. Uses a conditional import inside the dispatch branch to avoid importing the snapshot module at CLI startup time. Includes a module-level logger via `logging.getLogger(__name__)`. — Not Started

**Acceptance Criteria:**

- `create_parser()` returns a correctly configured `ArgumentParser` with an `order-books` subcommand
- `--rate-limit` and `--db-url` appear on the `order-books` subparser (added via `add_global_args()`)
- `--interval` and `--market-limit` are specific to the `order-books` subparser with correct types and defaults
- `main()` dispatches `order-books` to `order_book_snapshots.run(args)` with a conditional import
- All pytest tests for CLI pass

### Task 3: Create order book snapshot collector with tests — Not Started

Create `src/prediction_markets/collectors/order_book_snapshots.py` implementing the `run(args: argparse.Namespace) -> None` function that captures order book snapshots for all tracked markets. Create `tests/collectors/test_order_book_snapshots.py` with unit tests mocking the PolymarketClient, database, and rate limiter.

a. Write test `test_single_shot_mode_snapshots_and_inserts` — mock `get_markets()` to return a list of 3 market dicts (each with `id`, `token_id_yes`, `token_id_no` fields). Mock `PolymarketClient.get_order_book()` to return `OrderBookData` objects with known bid/ask values. Mock `get_connection()` with a context manager returning a mock connection. Mock `insert_order_book_snapshot()` to do nothing. Patch `datetime.now()` to return a fixed timestamp. Call `run(args)` with `interval=0`, `market_limit=0`, `rate_limit=10.0`. Assert `get_order_book()` was called exactly 6 times (2 per market). Assert `insert_order_book_snapshot()` was called 3 times (once per market). Assert each call to `insert_order_book_snapshot()` passes the fixed timestamp and the expected bid/ask values for both tokens. — Not Started

b. Write test `test_continuous_mode_runs_multiple_cycles` — mock `get_markets()` to return 2 markets. Mock `time.sleep()` with a no-op. Mock `get_order_book()` to return valid `OrderBookData` objects. Mock `insert_order_book_snapshot()` to do nothing. Call `run(args)` with `interval=300`, `market_limit=0`, `rate_limit=100`. Set up the test to stop after 3 cycles (e.g., by patching `time.sleep` to raise `StopIteration` on the 3rd call, or using a side effect that counts). Assert `get_order_book()` was called 12 times (2 markets × 2 tokens × 3 cycles). Assert `insert_order_book_snapshot()` was called 6 times (2 markets × 3 cycles). — Not Started

c. Write test `test_rate_limiter_called_per_api_call` — mock `get_markets()` to return 2 markets. Mock `RateLimiter` with a mock instance that records `acquire()` calls. Mock `PolymarketClient` to use the mock rate limiter. Mock `get_order_book()` to return valid data. Call `run(args)` in single-shot mode with `rate_limit=5.0`. Assert `acquire()` was called exactly 4 times (2 markets × 2 tokens per market). — Not Started

d. Write test `test_shared_timestamp_per_market` — mock `get_markets()` to return 1 market. Mock `get_order_book()` to return valid data. Patch `datetime.now()` to return incrementing timestamps (e.g., side_effect = [T1, T1, T2, T2] for two markets). Call `run(args)` in single-shot mode. Assert that `insert_order_book_snapshot()` received the same timestamp argument for both tokens of the same market. — Not Started

e. Write test `test_batch_insert_in_transaction` — mock `get_markets()` to return 3 markets. Mock `get_connection()` context manager. Mock `insert_order_book_snapshot()` to do nothing. Call `run(args)` in single-shot mode. Assert `get_connection()` was used as a context manager (commit called once at the end of the cycle). Assert `insert_order_book_snapshot()` was called within the context manager block. — Not Started

f. Write test `test_api_failure_skips_market` — mock `get_markets()` to return 3 markets. Mock `get_order_book()` to return valid data for markets 1 and 3, but raise `MarketClientError` for market 2's first call (token_id_yes). Mock `insert_order_book_snapshot()` to do nothing. Mock the module logger. Call `run(args)` in single-shot mode. Assert `insert_order_book_snapshot()` was called for markets 1 and 3 only (not market 2). Assert a `WARNING`-level log message was emitted containing the failed market's ID. — Not Started

g. Write test `test_db_insert_failure_logs_error_single_shot` — mock `get_markets()` to return 2 markets. Mock `get_order_book()` to return valid data. Mock `insert_order_book_snapshot()` to raise `Exception("DB error")` on the second call. Mock the module logger. Call `run(args)` in single-shot mode with `interval=0`. Assert an `ERROR`-level log message was emitted. Verify the function exits (does not hang or crash). — Not Started

h. Write test `test_empty_markets_table_exits_with_message` — mock `get_markets()` to return an empty list `[]`. Mock the module logger. Call `run(args)` in single-shot mode. Assert an `INFO`-level or `WARNING`-level log message was emitted indicating no markets are in the database and suggesting the metadata collector be run first. Assert the function exits without calling `get_order_book()` or `insert_order_book_snapshot()`. — Not Started

i. Write test `test_market_limit_caps_markets_per_cycle` — mock `get_markets()` to return 10 markets (sorted by volume descending). Call `run(args)` in single-shot mode with `market_limit=3`. Assert `get_order_book()` was called exactly 6 times (3 markets × 2 tokens). Assert the first 3 markets from the `get_markets()` result were processed. — Not Started

j. Write test `test_progress_logging_every_50_markets` — mock `get_markets()` to return 120 markets. Mock `get_order_book()` to return valid data. Mock `insert_order_book_snapshot()` to do nothing. Mock the module logger. Call `run(args)` in single-shot mode with `rate_limit=1000`. Assert `INFO`-level progress log messages were emitted at market counts 50 and 100. Assert the log messages include the current count, total markets, elapsed time, and estimated remaining time. — Not Started

k. Write test `test_recommended_market_limit_logged` — mock `get_markets()` to return 2 markets. Mock the module logger. Call `run(args)` in single-shot mode with `rate_limit=1.0`, `interval=300`. Assert an `INFO`-level log message was emitted containing a recommended `--market-limit` value. The recommended value should be approximately `int(interval * rate_limit / 2)` = `int(300 * 1.0 / 2)` = 150. — Not Started

l. Write test `test_cycle_exceeds_interval_warning` — mock `get_markets()` to return 2 markets. Patch `time.time()` to simulate a cycle that takes longer than the interval (e.g., first call returns 0.0, second returns `interval + 10`). Mock the module logger. Call `run(args)` in continuous mode (simulate one cycle then stop). Assert a `WARNING`-level log message was emitted indicating the cycle duration exceeded the interval. — Not Started

m. Write test `test_graceful_shutdown_on_sigint` — mock `get_markets()` to return 2 markets. Mock `time.sleep()` to raise `KeyboardInterrupt` on the first sleep call (simulating SIGINT). Mock `insert_order_book_snapshot()` to do nothing. Mock the module logger. Call `run(args)` in continuous mode with `interval=300`. Assert an `INFO`-level log message was emitted indicating graceful shutdown. Assert the function exits cleanly (no unhandled exceptions). — Not Started

n. Implement `run(args: argparse.Namespace) -> None` in `src/prediction_markets/collectors/order_book_snapshots.py` — the complete snapshot collector function. Implementation details:

   * **Imports and logger**: Import `logging`, `time`, `signal`, `datetime`, `timezone` from standard library. Import `PolymarketClient`, `RateLimiter`, `MarketClientError` from `prediction_markets.market_interfaces`. Import `get_connection` from `prediction_markets.db`. Import `get_markets`, `insert_order_book_snapshot` from `prediction_markets.db.queries`. Create a module-level logger via `logging.getLogger(__name__)`.

   * **Initialization**: Create `RateLimiter(args.rate_limit)` and `PolymarketClient(rate_limiter)`. If `args.db_url` is provided, construct a `psycopg2.connect(args.db_url)` connection directly; otherwise, use `get_connection()`. Set up a SIGTERM/SIGINT handler that sets a `shutdown_requested` flag (use `threading.Event` or a simple boolean). In single-shot mode (`interval == 0`), the shutdown handler is still registered for clean exit on Ctrl+C.

   * **Cycle loop** (continuous mode: `while not shutdown_requested`; single-shot: run once):
     1. Record cycle start time via `time.monotonic()`.
     2. Get DB connection via `get_connection()` (or args.db_url).
     3. Query markets: `markets = get_markets(conn)`. Sort by `volume` descending (handle `None` volume as 0). Slice to `args.market_limit` if > 0. If empty, log and exit (single-shot) or continue to next cycle (continuous).
     4. Log cycle start: number of markets to process.
     5. Compute and log recommended `--market-limit`: `int(args.interval * args.rate_limit / 2)` when `interval > 0`.
     6. For each market, with a counter:
        - If `shutdown_requested`: break out of the market loop, flush collected snapshots, exit.
        - Record `timestamp = datetime.now(timezone.utc)` (same timestamp for both token calls).
        - Call `rate_limiter.acquire()` then `client.get_order_book(market['token_id_yes'])`. On `MarketClientError`: log `WARNING` with market ID, skip this market, continue.
        - Call `rate_limiter.acquire()` then `client.get_order_book(market['token_id_no'])`. On `MarketClientError`: log `WARNING` with market ID, skip this market, continue.
        - Extract `best_bid` and `best_ask` from both `OrderBookData` objects (use `getattr` with `None` default for missing attributes to handle empty order books).
        - Append a snapshot dict: `{'market_id': ..., 'timestamp': ..., 'best_bid_yes': ..., 'best_ask_yes': ..., 'best_bid_no': ..., 'best_ask_no': ...}`.
        - Every 50 markets: log `INFO` progress message with markets processed, total, elapsed time, estimated remaining time.
     7. After processing all markets: batch insert all snapshots within a single transaction (using the connection context manager from `db/__init__.py`). Call `insert_order_book_snapshot(conn, **snapshot)` for each snapshot. On DB error: log `ERROR`, rollback, and either exit (single-shot) or continue to next cycle (continuous).
     8. Log cycle complete: markets snapshotted, cycle duration (seconds), rows inserted.
     9. In continuous mode: compute `elapsed = time.monotonic() - cycle_start`. If `elapsed > args.interval`: log `WARNING` that cycle exceeded interval. Sleep for `max(0, args.interval - elapsed)` seconds.
     10. In single-shot mode: exit the loop (break).

   * **Exit**: Log final summary — total cycles, total snapshots captured. — Not Started

o. Run `uv run pytest tests/collectors/test_order_book_snapshots.py -v` and verify all 13 tests pass — Not Started

**Acceptance Criteria:**

- `run(args)` in single-shot mode snapshots all markets, batch-inserts rows, and exits with summary
- `run(args)` in continuous mode cycles at the configured interval until interrupted
- Rate limiter `acquire()` is called twice per market per cycle (once for each token)
- Both API calls for a market share the same `datetime.now(timezone.utc)` timestamp
- API failures for a single market skip that market; the cycle continues with remaining markets
- DB insert failures log an error and either exit (single-shot) or continue (continuous)
- Empty markets table exits with a descriptive message
- Progress logging occurs every 50 markets with count, elapsed time, and estimated remaining time
- Recommended `--market-limit` value is computed and logged at cycle start (continuous mode)
- A warning is logged when a cycle exceeds the configured interval
- SIGINT/SIGTERM triggers graceful shutdown: finish current market, flush pending inserts, log shutdown message, exit
- All 13 unit tests pass with all external dependencies (client, database, rate limiter, time, signal) mocked

### Task 4: Integration verification — Not Started

Manual verification that the completed `collectors` package functions correctly when its dependencies are also implemented. No code changes — verification only.

a. Verify `python -m prediction_markets.collectors order-books --help` prints usage information showing the `order-books` subcommand with `--interval`, `--market-limit`, `--rate-limit`, and `--db-url` options — Not Started

b. Verify `python -m prediction_markets.collectors order-books --rate-limit 1 --market-limit 3` runs a single snapshot cycle against real Polymarket APIs and a running PostgreSQL database, inserting rows into `order_book_snapshots` — Not Started

c. Verify `python -m prediction_markets.collectors order-books --rate-limit 1 --interval 10 --market-limit 3` runs continuously, cycling every ~10 seconds — Not Started

d. Verify `SELECT COUNT(*) FROM order_book_snapshots` shows the expected number of rows (markets × cycles) after the tests in (b) and (c) — Not Started

e. Verify SIGINT (Ctrl+C) during continuous mode exits cleanly with a shutdown message — Not Started

**Acceptance Criteria:**

- CLI help output is correct and complete
- Single-shot mode inserts snapshot rows into the database
- Continuous mode cycles at the configured interval
- Database row counts match expectations
- Continuous mode shuts down cleanly on SIGINT

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks (Tasks 2 and 3), tests are written before implementation. Test subtasks precede implementation subtasks within each task. Tests for the CLI mock `order_book_snapshots.run` to verify dispatch without requiring the collector module. Tests for the snapshot collector mock `PolymarketClient`, `RateLimiter`, `get_markets`, `insert_order_book_snapshot`, and `get_connection` — no live API or database access is required during unit testing.

## Constraints

- All file paths in task descriptions are relative to `source_root` (`/prediction-markets`). Commands like `uv run pytest` run from `source_root`.
- No estimation or data transformation beyond extracting fields from API responses — store raw observations
- All timestamps are actual observation times (`datetime.now(timezone.utc)`), shared between both token calls for a market
- Rate limiting applies per API call (2 calls per market per cycle)
- Database queries use the functions provided by Story001 (`get_markets`, `insert_order_book_snapshot`) — do not write raw SQL in the collector
- API calls use the `PolymarketClient` from Story002 — do not call the SDK directly
- `--rate-limit` and `--db-url` appear on every subcommand (added via `add_global_args` helper), not only on the root parser
- The CLI subparser architecture must support adding new subcommands (`metadata`, `prices`) in future stories without restructuring
- Logging: use Python's standard `logging` module with `logging.getLogger(__name__)` in each module. `INFO` level for cycle start/end, progress (every 50 markets), and recommended market-limit. `WARNING` level for skipped markets (API failures) and cycle-exceeds-interval warnings. `ERROR` level for database insert failures. All log messages must be informative and include relevant context (market ID, counts, timing).
- Mock all external dependencies in unit tests — `unittest.mock` for `PolymarketClient`, `RateLimiter`, database functions, `time.sleep`, `datetime.now`, and `signal.signal`. Do not require a running database or network access for any test.

## Notes

- The `source_root` for this project is the project root (`/prediction-markets`). All file paths in task descriptions are relative to `source_root`. Commands like `uv run pytest` run from `source_root`.
- This story creates the `collectors` package from scratch — the directory `src/prediction_markets/collectors/` does not exist before this story runs.
- The collector does not perform analysis — it only stores raw observations. Analysis stories will query `order_book_snapshots` for arbitrage detection.
- With `--rate-limit 1` and 500 markets, a single cycle takes ~17 minutes (1000 API calls at 1 req/sec). The recommended `--market-limit` formula (`interval × rate_limit / 2`) helps users select appropriate limits for their polling interval.
- Database connections are created fresh each cycle (via `get_connection()`) to handle long-running continuous mode without stale connections.
- The `insert_order_book_snapshot` function from Story001 uses `ON CONFLICT (market_id, timestamp) DO NOTHING` — re-running the collector will not produce duplicate rows for the same market at the same timestamp.
