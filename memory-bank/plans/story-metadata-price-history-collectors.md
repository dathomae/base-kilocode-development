# Story Plan: Market Metadata & Price History Collectors

**Parent Plan**: [Polymarket Exploration Plan](polymarket-exploration-plan.md)
**Dependencies**: Story 1 (Database & Schema), Story 2 (Market Interface Module)
**Parallel With**: Story 4 (Order Book Snapshot Collector)

## Objective

Build the CLI-driven collectors that discover Polymarket markets and download historical price data, storing everything in PostgreSQL. This story delivers immediate analytical value — once the price history collector runs, the maker scenario analysis is fully enabled with historical data and no waiting period.

## Deliverables

1. `src/prediction_markets/collectors/__init__.py`
2. `src/prediction_markets/collectors/cli.py` — unified CLI entry point
3. `src/prediction_markets/collectors/market_metadata.py` — market discovery
4. `src/prediction_markets/collectors/price_history.py` — historical midpoint download
5. `src/prediction_markets/collectors/base.py` — shared collector utilities (rate-limited client setup)

## Detailed Requirements

### CLI Entry Point (`cli.py`)

Uses `argparse` with subcommands. Not a heavyweight framework — simple and direct.

```
python -m prediction_markets.collectors COMMAND [OPTIONS]
```

**Global options** (accepted by all subcommands):
- `--rate-limit FLOAT` (default: `1.0`): maximum API requests per second
- `--db-url TEXT` (optional): override the default database URL from environment variables

**Subcommands** (for this story):
- `metadata` — discover and store markets
- `prices` — download historical price data

Each subcommand calls the corresponding module's `run(args)` function with the parsed arguments.

### Shared Base (`base.py`)

Provides:
- `create_client(rate_limit: float) -> tuple[MarketClient, ...]`: Creates the Polymarket client with the configured rate limiter
- `get_db_connection()`: Returns a database connection using the `db` module
- `log_progress(message: str, current: int, total: int)`: Standardized progress logging

All collectors use this module for consistent setup.

### Market Metadata Collector (`market_metadata.py`)

**`run(args: argparse.Namespace) -> None`**

Workflow:
1. Create client and DB connection
2. Call `client.list_active_markets()`
3. For each market:
   a. Call `client.get_fee_rate(condition_id)` to get the taker fee rate
   b. Build a dict with all `markets` table columns from the `MarketData` object
   c. Batch insert (e.g., 50 markets at a time) via `queries.insert_markets()`
   d. Call `client.rate_limiter.acquire()` between API calls
4. Log summary: total markets discovered, inserted, updated, errors

**Filtering**: Only markets with `enable_order_book=True` and `neg_risk=False` (binary, non-multi-market). Apply this filter in `list_active_markets()` or immediately after.

**Fee rate**: Call `get_clob_market_info(condition_id)` for each market. The SDK returns `fd.r` (fee rate) and `fd.to` (taker_only flag). Store `fd.r` as `fee_rate`. If the call fails, log a warning and use `0.0` with a note.

**Idempotency**: The collector can be run multiple times. Use `INSERT ... ON CONFLICT (id) DO UPDATE` to refresh market metadata (volume, liquidity, end_date may change).

**Logging**: Progress every 50 markets. Log warnings for markets that can't be processed. Final summary with counts.

### Price History Collector (`price_history.py`)

**`run(args: argparse.Namespace) -> None`**

Additional options:
- `--lookback INT` (default: `90`): days of history to fetch
- `--market-limit INT` (default: `0` = all): max markets to process
- `--interval TEXT` (default: `"1h"`): price history interval (`"1h"`, `"6h"`, `"1d"`, `"1w"`, `"max"`)

Workflow:
1. Create client and DB connection
2. Query all markets from the `markets` table
3. For each market (up to `--market-limit`):
   a. Compute `start_ts` from `--lookback` (unix timestamp: now - lookback_days * 86400)
   b. Compute `end_ts` as current unix timestamp
   c. Call `client.get_price_history(token_id=market.token_id_yes, start_ts, end_ts, interval)`
   d. Call `client.get_price_history(token_id=market.token_id_no, start_ts, end_ts, interval)`
   e. Map `PricePoint` objects to dicts with `market_id`, `token_id`, `timestamp`, `price`
   f. Batch insert into `token_prices` via `queries.insert_token_prices()`
   g. Apply rate limiting between API calls (2 calls per market: Yes + No)
4. Log per-market progress and final summary

**Skip logic**: Before fetching, check if the market already has price data covering the requested range. Query `SELECT MIN(timestamp), MAX(timestamp) FROM token_prices WHERE market_id = ?`. If the range is already fully covered, skip the market. If partially covered, adjust `start_ts` to fill only the gap.

**Error handling**: If a market's price history call fails, log the error and continue to the next market. Do not abort the entire run.

**Progress tracking**: Print a line every N markets showing: markets processed, total price points inserted, elapsed time, estimated time remaining.

## Edge Cases

- **Empty market list**: The `metadata` collector should log "No markets found" rather than erroring.
- **API returns no price history**: A market may have no historical data (newly created). Log a note and skip, don't error.
- **Large result sets**: `get_price_history` with a 90-day lookback at 1-hour intervals may return ~2160 points per token. For 500 markets, that's ~2.16M rows. The batch insert must handle this (insert in chunks of 1000 rows).
- **Interrupted runs**: If the collector is killed mid-run, running it again should pick up where it left off (skip logic based on existing data).
- **Rate limit violations**: If the API returns 429, the collector should back off and retry. Implement a simple exponential backoff in the rate limiter or collector.

## Verification

- `python -m prediction_markets.collectors metadata --rate-limit 1` populates the `markets` table
- Running it again updates existing rows without duplicating
- `python -m prediction_markets.collectors prices --rate-limit 1 --lookback 7` inserts price points
- Running it again with the same `--lookback` skips already-fetched data (no duplicate rows, no redundant API calls)
- The `--rate-limit 0.5` option enforces 1 request per 2 seconds (verify with timing logs)
- Interrupting and re-running the prices collector resumes from where it left off
