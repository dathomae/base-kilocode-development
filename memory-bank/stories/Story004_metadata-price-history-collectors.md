# Story004: Market Metadata & Price History Collectors

## Goal

Build the CLI-driven collectors that discover Polymarket markets and download historical price data, storing everything in PostgreSQL. This story delivers immediate analytical value — once the price history collector runs, the maker scenario analysis is fully enabled with historical data and no waiting period.

This story implements [Story Plan: Market Metadata & Price History Collectors](../plans/story-metadata-price-history-collectors.md).

## References

- [Story Plan: Market Metadata & Price History Collectors](../plans/story-metadata-price-history-collectors.md)
- [Polymarket Exploration Plan](../plans/polymarket-exploration-plan.md)

## Dependencies

- [Story001_devcontainer-database-schema.md](Story001_devcontainer-database-schema.md) — provides `db/__init__.py` (connection utilities), `db/queries.py` (`insert_markets`, `insert_token_prices`, `get_markets`), and the `markets` / `token_prices` schema — Not Started
- [Story002_market-interface-module.md](Story002_market-interface-module.md) — provides `market_interfaces/` package (`PolymarketClient`, `RateLimiter`, `MarketData`, `PricePoint`, `MarketClientError`) — Not Started
- [Story003_order-book-snapshot-collector.md](Story003_order-book-snapshot-collector.md) — provides `collectors/` package (`__init__.py`, `__main__.py`, `cli.py` with `order-books` subcommand) — Not Started

## Dependent Stories

None — No stories have been created that depend on this collector. Future analysis stories will query `token_prices` and `markets` for maker scenario modeling, but those stories have not yet been created.

## Tasks

### Task Sizing

- **Task 1** is classified as **[Small]**. Two new files (`base.py` + `test_base.py`), purely additive. Three simple helper functions, each a single-concern delegation following established patterns from Story001 and Story002. Narrow scope, no ambiguity.

- **Task 2** is classified as **medium** (not `[Small]`). Two new files (`market_metadata.py` + `test_market_metadata.py`), but the `run(args)` implementation coordinates multiple concerns: API calls via `PolymarketClient`, rate limiter integration, database batch inserts with `ON CONFLICT`, fee-rate failure handling, market filtering, progress logging, and empty-market-list handling. Eight distinct edge cases tested. Exceeds the "narrow scope" Small threshold.

- **Task 3** is classified as **medium** (not `[Small]`). Two new files (`price_history.py` + `test_price_history.py`), but the `run(args)` implementation requires reasoning about timestamp range computation, skip logic for fully covered markets, partial coverage gap filling, market-limit capping, per-market API failure resilience, batch insert chunking (≤1000 rows), and progress tracking. Multiple interacting concerns exceed the Small threshold.

- **Task 4** is classified as **[Small]**. Two files (`cli.py` + `test_cli.py`), additive change to existing argparse architecture. Adding `metadata` and `prices` subcommands mechanically follows the identical pattern established by the `order-books` subcommand in Story003. Narrow scope, single concern.

- **Task 5** is manual integration verification against live APIs — not a code task. Not annotated `[Small]`.

The architect-for-story-planning evaluated all tasks for decomposition. Tasks 2 and 3 have relatively high subtask counts (10 and 12), but all subtasks are test functions exercising a single `run()` entry point — not indicators of multi-phase structure or broad file span. Splitting would create artificial intermediate modules with no isolation benefit. All tasks are at their natural granularity.

### Parallel Execution

- **Group A** (Tasks 2, 3): Both tasks modify fully disjoint files within the `collectors` package and its test directory. Task 2 creates `market_metadata.py` + `test_market_metadata.py`; Task 3 creates `price_history.py` + `test_price_history.py`. Neither task imports from the other. Both import from `base.py` (Task 1), which is merged before Group A begins. Each can be independently tested.

### Execution Order

```
Task 1 → [Task 2 ‖ Task 3] → Task 4 → Task 5
```

- **Task 1** must complete first — it creates `base.py`, which Tasks 2 and 3 import.
- **Group A** (Tasks 2, 3) run in parallel after Task 1 merge.
- **Task 4** follows sequentially — it adds dispatch branches in `cli.py` that import `market_metadata.run` and `price_history.run`. The dispatch tests require those modules to exist.
- **Task 5** is manual integration verification — runs last after all code is merged.

### Reintegration

**Group A (Tasks 2, 3):**

- **Merge order**: any order (fully disjoint file sets — `market_metadata.py` vs `price_history.py` and their respective test files)
- **Integration tests**:
  1. `uv run pytest tests/collectors/ -v --tb=short` — verify all collector tests (base tests + metadata tests + price history tests) pass together in the merged tree.
  2. `uv run python -c "from prediction_markets.collectors.base import create_client, get_db_connection, log_progress; from prediction_markets.collectors.market_metadata import run as metadata_run; from prediction_markets.collectors.price_history import run as prices_run; print('All imports OK')"` — verify cross-module imports resolve without conflicts.
- **Watch for conflicts in**:
  - `src/prediction_markets/collectors/__init__.py` — not modified by either task, but if a sub-agent auto-creates or appends to it, conflicts may arise. Verify it is unchanged after merge.
  - `src/prediction_markets/collectors/__pycache__/` — bytecode caches may differ per worktree. Git-ignored; no action needed.

**Task 4 → Task 5 sequential (no parallel group):**

After Task 4 is merged, run `uv run pytest tests/collectors/test_cli.py -v` to verify the extended CLI still passes all existing and new tests. Then proceed to Task 5 manual verification.

---

### Task 1: Create collector base utilities module with tests — Not Started [Small]

Create `src/prediction_markets/collectors/base.py` with shared utilities for client setup, database connection, and progress logging. Create `tests/collectors/test_base.py` with unit tests.

a. Write test `test_create_client_returns_polymarket_client` in `tests/collectors/test_base.py` — call `create_client(1.0)` and assert it returns a `PolymarketClient` instance whose `_rate_limiter` has `requests_per_second == 1.0` — Not Started
b. Write test `test_get_db_connection_delegates` — patch `prediction_markets.db.get_connection` with a `Mock(return_value=mock_conn)`, call `get_db_connection()`, and assert the mock was called once and the return value is `mock_conn` — Not Started
c. Write test `test_log_progress_emits_info` — patch `logging.getLogger`, call `log_progress("Processing", 50, 100)`, and assert an `INFO`-level log message was emitted containing `"Processing: 50/100"` — Not Started
d. Implement `create_client(rate_limit: float) -> PolymarketClient` in `src/prediction_markets/collectors/base.py` — creates `RateLimiter(rate_limit)` and `PolymarketClient(rate_limiter)`, returns the client — Not Started
e. Implement `get_db_connection()` in `base.py` — imports and calls `get_connection()` from `prediction_markets.db`, returns the connection — Not Started
f. Implement `log_progress(message: str, current: int, total: int) -> None` in `base.py` — logs an `INFO`-level message via `logging.getLogger(__name__)` with the format `"<message>: <current>/<total>"` — Not Started
g. Run `uv run pytest tests/collectors/test_base.py -v` and verify all tests pass — Not Started

**Acceptance Criteria:**

- `create_client(1.0)` returns a `PolymarketClient` with a `RateLimiter(1.0)`
- `get_db_connection()` delegates to `prediction_markets.db.get_connection()`
- `log_progress("X", 5, 10)` emits an `INFO`-level log message containing `"X: 5/10"`
- All unit tests pass

### Task 2: Create market metadata collector with tests — Not Started

Create `src/prediction_markets/collectors/market_metadata.py` implementing the `run(args: argparse.Namespace) -> None` function that discovers active Polymarket markets and stores them in the database. Create `tests/collectors/test_market_metadata.py` with unit tests.

a. Write test `test_run_discovers_and_inserts_markets` — mock `create_client()` and `get_db_connection()`, mock `client.list_active_markets()` to return 3 `MarketData` objects, mock `client.get_fee_rate()` to return `0.02`, mock `queries.insert_markets()` to do nothing. Call `run(args)` with `rate_limit=1.0`, `db_url=None`. Assert `list_active_markets()` was called once. Assert `get_fee_rate()` was called 3 times. Assert `insert_markets()` was called once with a list of 3 dicts containing all `markets` table columns. — Not Started
b. Write test `test_run_applies_rate_limiting_between_calls` — mock `create_client()` with a mock rate limiter, mock `list_active_markets()` to return 2 markets, mock `get_fee_rate()`. Call `run(args)`. Assert `rate_limiter.acquire()` was called at least twice (once before `list_active_markets`, once per fee rate call). — Not Started
c. Write test `test_run_filters_non_order_book_markets` — mock `list_active_markets()` to return markets with mixed `enable_order_book` and `neg_risk` values. Assert only markets with `enable_order_book=True` and `neg_risk=False` are passed to `insert_markets()`. — Not Started
d. Write test `test_run_handles_fee_rate_failure` — mock `get_fee_rate()` to raise `MarketClientError` for one market. Mock the module logger. Call `run(args)`. Assert a `WARNING`-level log was emitted containing the market's `condition_id`. Assert the failing market uses `fee_rate=0.0` in the insert dict. Assert `insert_markets()` still receives all markets. — Not Started
e. Write test `test_run_batch_inserts` — mock `list_active_markets()` to return 120 markets. Mock `insert_markets()`. Call `run(args)`. Assert `insert_markets()` was called multiple times with batches of approximately 50 markets each. — Not Started
f. Write test `test_run_idempotent_insert` — mock `insert_markets()` to track calls. Call `run(args)` twice with the same mocked market data. Assert `insert_markets()` was called twice (the database's `ON CONFLICT DO UPDATE` handles deduplication; the collector simply calls insert each run). — Not Started
g. Write test `test_run_empty_market_list_logs_message` — mock `list_active_markets()` to return `[]`. Mock the module logger. Call `run(args)`. Assert an `INFO`-level or `WARNING`-level message containing "No markets found" was logged. Assert `insert_markets()` was not called. — Not Started
h. Write test `test_run_progress_logging_every_50_markets` — mock `list_active_markets()` to return 120 markets. Mock the module logger. Call `run(args)`. Assert `INFO`-level progress messages were emitted at market counts 50 and 100. — Not Started
i. Implement `run(args: argparse.Namespace) -> None` in `market_metadata.py` — full implementation:
   - Imports: `argparse`, `logging`, `create_client`, `get_db_connection`, `log_progress` from `.base`; `queries` from `prediction_markets.db`
   - Create client via `create_client(args.rate_limit)`
   - Get DB connection: if `args.db_url` is provided, use `psycopg2.connect(args.db_url)` directly; otherwise use `get_db_connection()`
   - Call `client.list_active_markets()`
   - Filter to `enable_order_book=True` and `neg_risk=False`
   - For each market: call `client.get_fee_rate(market.condition_id)`, build a dict with all `markets` table columns (`id`, `condition_id`, `question`, `slug`, `token_id_yes`, `token_id_no`, `fee_rate`, `category`, `end_date`, `volume`, `liquidity`, `neg_risk`, `enable_order_book`), apply `client.rate_limiter.acquire()` between API calls
   - Batch insert via `queries.insert_markets(conn, batch)` in chunks of 50 markets
   - Log summary: total markets discovered, inserted, errors
   - Module-level logger via `logging.getLogger(__name__)` — Not Started
j. Run `uv run pytest tests/collectors/test_market_metadata.py -v` and verify all tests pass — Not Started

**Acceptance Criteria:**

- `run(args)` discovers markets via `list_active_markets()` and inserts them in batches of ~50
- Markets are filtered to `enable_order_book=True` and `neg_risk=False`
- Fee rate is fetched per market via `get_fee_rate(condition_id)`; failures log a `WARNING` and use `0.0`
- Empty market list logs "No markets found" without erroring
- Progress logged every 50 markets
- Rate limiting applied between API calls
- All unit tests pass

### Task 3: Create price history collector with tests — Not Started

Create `src/prediction_markets/collectors/price_history.py` implementing the `run(args: argparse.Namespace) -> None` function that downloads historical price data for all tracked markets. Create `tests/collectors/test_price_history.py` with unit tests.

a. Write test `test_run_fetches_and_inserts_prices` — mock `create_client()`, `get_db_connection()`, `queries.get_markets()` to return 2 market dicts (each with `id`, `token_id_yes`, `token_id_no`), `client.get_price_history()` to return lists of `PricePoint` objects. Mock `queries.insert_token_prices()` to do nothing. Call `run(args)` with `lookback=7`, `interval='1h'`, `market_limit=0`. Assert `get_price_history()` was called 4 times (2 markets × 2 tokens). Assert `insert_token_prices()` was called at least once with a flat list of all price points. — Not Started
b. Write test `test_run_computes_timestamp_range` — mock `get_markets()` to return 1 market. Mock `get_price_history()`. Call `run(args)` with `lookback=7`. Assert `get_price_history()` was called with `start_ts` approximately `now - 7*86400` and `end_ts` approximately `now`. — Not Started
c. Write test `test_run_skips_fully_covered_markets` — mock `get_markets()` to return 1 market. Mock a cursor executing `SELECT MIN(timestamp), MAX(timestamp) FROM token_prices WHERE market_id = %s` to return `(min_ts, max_ts)` that fully covers the requested range. Assert `get_price_history()` was not called for this market. — Not Started
d. Write test `test_run_adjusts_start_ts_for_partial_coverage` — mock `get_markets()` to return 1 market. Mock coverage query to return partial coverage (e.g., covers last 3 days of a 7-day lookback). Assert `get_price_history()` was called with `start_ts` set to fill only the gap (the uncovered earlier portion). — Not Started
e. Write test `test_run_respects_market_limit` — mock `get_markets()` to return 10 markets. Call `run(args)` with `market_limit=3`. Assert `get_price_history()` was called exactly 6 times (3 markets × 2 tokens). — Not Started
f. Write test `test_run_handles_api_failure_per_market` — mock `get_price_history()` to raise `MarketClientError` for one market's yes token. Mock the module logger. Call `run(args)`. Assert a `WARNING`-level log was emitted with the market ID. Assert `insert_token_prices()` still receives data for other markets. — Not Started
g. Write test `test_run_handles_empty_price_history` — mock `get_price_history()` to return `[]` for a token. Mock the module logger. Call `run(args)`. Assert an `INFO`-level or `WARNING`-level note was logged. Assert no error is raised. — Not Started
h. Write test `test_run_batch_inserts_large_result_sets` — mock `get_price_history()` to return 2500 points per token (5000 per market). Mock `insert_token_prices()` to do nothing. Call `run(args)` with 2 markets. Assert `insert_token_prices()` is called in chunks of 1000 rows or fewer per call. — Not Started
i. Write test `test_run_progress_tracking` — mock `get_markets()` to return 120 markets. Mock `get_price_history()` to return 10 points per token. Mock the module logger. Call `run(args)`. Assert `INFO`-level progress messages show markets processed, total price points inserted, elapsed time, and estimated remaining time. — Not Started
j. Write test `test_run_rate_limiting_between_calls` — mock `create_client()` with a mock rate limiter. Mock `get_markets()` to return 2 markets. Call `run(args)`. Assert `rate_limiter.acquire()` was called at least 4 times (2 markets × 2 tokens). — Not Started
k. Implement `run(args: argparse.Namespace) -> None` in `price_history.py` — full implementation:
   - Imports: `argparse`, `logging`, `time`, `datetime`, `timezone`, `create_client`, `get_db_connection`, `log_progress` from `.base`; `queries` from `prediction_markets.db`
   - Create client via `create_client(args.rate_limit)`
   - Get DB connection (use `args.db_url` override if provided)
   - Query all markets via `queries.get_markets(conn)`; slice to `args.market_limit` if > 0
   - For each market:
     - Query existing coverage: `SELECT MIN(timestamp), MAX(timestamp) FROM token_prices WHERE market_id = %s`
     - Compute `start_ts = int((datetime.now(timezone.utc) - timedelta(days=args.lookback)).timestamp())`
     - Compute `end_ts = int(datetime.now(timezone.utc).timestamp())`
     - If existing coverage fully covers `[start_ts, end_ts]`: log and skip
     - If partial coverage: adjust `start_ts` to `max(start_ts, existing_max_ts + 1)` (or gap-filling logic)
     - Call `client.rate_limiter.acquire()` then `client.get_price_history(market['token_id_yes'], start_ts, end_ts, args.interval)`
     - Call `client.rate_limiter.acquire()` then `client.get_price_history(market['token_id_no'], start_ts, end_ts, args.interval)`
     - On `MarketClientError`: log `WARNING` with market ID, skip this market, continue
     - Map `PricePoint` objects to dicts with `market_id`, `token_id`, `timestamp`, `price`
     - Batch insert via `queries.insert_token_prices(conn, batch)` in chunks of 1000 rows
   - Progress logging every N markets (e.g., every 10) with count, points inserted, elapsed/estimated time
   - Module-level logger via `logging.getLogger(__name__)` — Not Started
l. Run `uv run pytest tests/collectors/test_price_history.py -v` and verify all tests pass — Not Started

**Acceptance Criteria:**

- `run(args)` fetches price history for all markets up to `--market-limit`
- Skip logic avoids redundant API calls for fully covered markets
- Partial coverage adjusts `start_ts` to fill only the gap
- Batch inserts handle large result sets in chunks of ≤1000 rows
- API failures per market are logged and do not abort the run
- Empty price history is noted and skipped without error
- Progress logged with markets processed, points inserted, elapsed/estimated time
- Rate limiting applied between API calls (2 calls per market)
- All unit tests pass

### Task 4: Extend CLI with metadata and prices subcommands — Not Started [Small]

Modify `src/prediction_markets/collectors/cli.py` to add `metadata` and `prices` subcommands. Update `tests/collectors/test_cli.py` with tests for the new subcommands.

a. Write test `test_metadata_subparser_exists` in `tests/collectors/test_cli.py` — call `create_parser()` and verify a subparser named `metadata` exists with `--rate-limit` (type float, default 1.0) and `--db-url` (type str, default None) options. — Not Started
b. Write test `test_prices_subparser_exists` — call `create_parser()` and verify a subparser named `prices` exists with `--rate-limit`, `--db-url`, `--lookback` (type int, default 90), `--market-limit` (type int, default 0), and `--interval` (type str, default '1h'). — Not Started
c. Write test `test_main_dispatches_to_metadata_handler` — patch `prediction_markets.collectors.market_metadata.run` with a `Mock()`, set `sys.argv` to `['collectors', 'metadata', '--rate-limit', '1']`, call `main()`, and assert the mock was called once with an `argparse.Namespace`. — Not Started
d. Write test `test_main_dispatches_to_prices_handler` — patch `prediction_markets.collectors.price_history.run` with a `Mock()`, set `sys.argv` to `['collectors', 'prices', '--lookback', '7']`, call `main()`, and assert the mock was called once with a namespace whose `lookback` attribute equals 7. — Not Started
e. Write test `test_prices_parser_defaults` — call `parse_args(['prices'])` and verify `lookback=90`, `market_limit=0`, `interval='1h'`. — Not Started
f. Write test `test_prices_parser_custom_values` — call `parse_args(['prices', '--lookback', '30', '--market-limit', '50', '--interval', '1d'])` and verify `lookback=30`, `market_limit=50`, `interval='1d'`. — Not Started
g. Implement `metadata` subparser in `cli.py` — add `'metadata'` via `subparsers.add_parser('metadata', help='Discover and store Polymarket market metadata')`. Call `add_global_args(metadata_parser)` to attach `--rate-limit` and `--db-url`. — Not Started
h. Implement `prices` subparser in `cli.py` — add `'prices'` via `subparsers.add_parser('prices', help='Download historical price data')`. Call `add_global_args(prices_parser)`. Add `--lookback` (type int, default 90, help describing days of history to fetch), `--market-limit` (type int, default 0, help describing max markets to process; 0 for all), and `--interval` (type str, default '1h', choices=['1h', '6h', '1d', '1w', 'max'], help describing price history interval). — Not Started
i. Extend `main()` dispatch in `cli.py` — add `elif args.command == 'metadata':` branch with conditional import of `market_metadata.run(args)`, and `elif args.command == 'prices':` branch with conditional import of `price_history.run(args)`. Keep the existing `'order-books'` branch. — Not Started
j. Run `uv run pytest tests/collectors/test_cli.py -v` and verify all tests pass — Not Started

**Acceptance Criteria:**

- `metadata` subcommand exists with `--rate-limit` and `--db-url`
- `prices` subcommand exists with `--lookback`, `--market-limit`, `--interval`, `--rate-limit`, `--db-url`
- `main()` dispatches `metadata` to `market_metadata.run(args)` and `prices` to `price_history.run(args)` via conditional imports
- All CLI tests pass, including existing `order-books` tests

### Task 5: Integration verification — Not Started

Manual verification that the completed collectors function correctly against live Polymarket APIs and a running PostgreSQL database. No code changes — verification only.

a. Verify `python -m prediction_markets.collectors metadata --help` prints usage for the metadata subcommand — Not Started
b. Verify `python -m prediction_markets.collectors metadata --rate-limit 1` populates the `markets` table with live Polymarket data — Not Started
c. Verify running the metadata collector again updates existing rows without duplicating — Not Started
d. Verify `python -m prediction_markets.collectors prices --rate-limit 1 --lookback 7` inserts price points into `token_prices` — Not Started
e. Verify running the prices collector again with the same `--lookback` skips already-fetched data (no duplicate rows, no redundant API calls) — Not Started
f. Verify `python -m prediction_markets.collectors prices --rate-limit 0.5` enforces approximately 1 request per 2 seconds — Not Started
g. Verify interrupting (Ctrl+C) and re-running the prices collector resumes from where it left off — Not Started

**Acceptance Criteria:**

- CLI help output is correct for all subcommands
- Metadata collector populates and updates the `markets` table
- Price collector inserts price points and skips redundant fetches on re-run
- Rate limiting is enforced correctly
- Interrupted runs resume correctly on re-run

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks (Tasks 1–4), tests are written before implementation. Test subtasks precede implementation subtasks within each task. Tests for the CLI mock `market_metadata.run` and `price_history.run` to verify dispatch without requiring the collector modules. Tests for the collectors mock `PolymarketClient`, `RateLimiter`, database query functions, and `logging.getLogger` — no live API or database access is required during unit testing.

## Constraints

- All file paths in task descriptions are relative to `source_root` (`/prediction-markets`). Commands like `uv run pytest` run from `source_root`.
- API calls use `PolymarketClient` from Story002 — do not call the SDK directly
- Database queries use the functions provided by Story001 (`get_markets`, `insert_markets`, `insert_token_prices`) — do not write raw SQL for inserts; raw SQL is acceptable only for the coverage check (`SELECT MIN(timestamp), MAX(timestamp)`) inside the price history collector
- `--rate-limit` and `--db-url` appear on every subcommand (added via `add_global_args` helper), not only on the root parser
- The CLI subparser architecture must support adding future subcommands without restructuring
- Batch insert size: ~50 markets for metadata, ≤1000 price points for price history
- Logging: use Python's standard `logging` module with `logging.getLogger(__name__)` in each module. `INFO` level for cycle start/end and progress. `WARNING` level for skipped markets (API failures, empty lists, cycle-exceeds-interval). `ERROR` level for database insert failures. All log messages must be informative and include relevant context (market ID, counts, timing)
- Mock all external dependencies in unit tests — `unittest.mock` for `PolymarketClient`, `RateLimiter`, database functions, `datetime`, and `time`. Do not require a running database or network access for any test
- Fee rate failures: log `WARNING` and use `0.0` as the fee rate; do not abort the run
- Empty price history: log a note and skip; do not error
- Large result sets: `get_price_history` may return ~2160 points per token for 90-day 1h lookback. Batch insert in chunks of 1000 rows
- Interrupted runs: skip logic based on existing data ensures resuming from where it left off

## Notes

- The `source_root` for this project is the project root (`/prediction-markets`). All file paths in task descriptions are relative to `source_root`. Commands like `uv run pytest` run from `source_root`.
- This story extends the `collectors` package created in Story003 — it modifies `cli.py` and adds three new modules (`base.py`, `market_metadata.py`, `price_history.py`).
- The metadata collector can be run multiple times safely due to `ON CONFLICT (id) DO UPDATE` in `insert_markets()`.
- The price history collector's skip logic ensures idempotency: re-running with the same `--lookback` does not produce duplicate rows or redundant API calls.
- With `--rate-limit 1` and 500 markets, the price history collector makes 1000 API calls (2 per market), taking ~17 minutes for a full fetch. The skip logic makes re-runs nearly instant for already-covered markets.
- Database connections are created fresh per collector run via `get_db_connection()` (or `args.db_url` override) to avoid stale connections.
