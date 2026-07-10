# Story Plan: Order Book Snapshot Collector

**Parent Plan**: [Polymarket Exploration Plan](polymarket-exploration-plan.md)
**Dependencies**: Story 1 (Database & Schema), Story 2 (Market Interface Module)
**Parallel With**: Story 3 (Metadata & Price History Collectors)

## Objective

Build the CLI-driven collector that captures order book snapshots from Polymarket at regular intervals. This is the primary data source for the taker analysis — it records actual bid/ask prices for both tokens with no estimation or approximation.

## Deliverables

1. `src/prediction_markets/collectors/order_book_snapshots.py` — snapshot collector
2. New `order-books` subcommand in `src/prediction_markets/collectors/cli.py`

## Detailed Requirements

### CLI Subcommand

Extends the existing CLI in `cli.py`:

```
python -m prediction_markets.collectors order-books [OPTIONS]
```

**Options** (in addition to the global `--rate-limit` and `--db-url`):
- `--interval INT` (default: `0`): seconds between snapshot cycles. `0` means run once and exit.
- `--market-limit INT` (default: `0` = all): max markets to snapshot per cycle

When `--interval 0` (single-shot mode): snapshot all markets once, print summary, exit.
When `--interval 300` (continuous mode): snapshot all markets, sleep 300 seconds, repeat indefinitely. Graceful shutdown on SIGTERM/SIGINT.

### Snapshot Collector (`order_book_snapshots.py`)

**`run(args: argparse.Namespace) -> None`**

Workflow for a single snapshot cycle:

1. Query all markets from the `markets` table (up to `--market-limit`, ordered by volume descending to prioritize liquid markets)
2. For each market:
   a. Call `client.get_order_book(market.token_id_yes)` → get `best_bid`, `best_ask`
   b. Call `client.get_order_book(market.token_id_no)` → get `best_bid`, `best_ask`
   c. Both API calls share the same `timestamp = datetime.now(timezone.utc)`
   d. Build a dict: `{market_id, timestamp, best_bid_yes, best_ask_yes, best_bid_no, best_ask_no}`
   e. Apply rate limiting between each API call (2 calls per market)
3. Batch insert all snapshot rows into `order_book_snapshots` in a single transaction
4. Log per-cycle summary: markets snapshotted, cycle duration, next cycle time

**Continuous mode**:
- Track cycle start time, compute actual cycle duration
- If a cycle takes longer than `--interval`, log a warning and start the next cycle immediately
- Sleep for `--interval - cycle_duration` between cycles (minimum 0)

**Single-shot mode**:
- Run one cycle and exit with summary

### Timestamp Handling

Both `get_order_book` calls for a market happen in rapid succession (separated only by the rate limiter delay). Use the same timestamp for both, recorded at the start of the pair. This ensures the analysis sees the Yes and No prices as belonging to the same logical snapshot, even if there's a small sub-second gap between the two API responses.

```python
timestamp = datetime.now(timezone.utc)
book_yes = client.get_order_book(market.token_id_yes)   # after rate_limiter delay
book_no  = client.get_order_book(market.token_id_no)    # after rate_limiter delay
# Use `timestamp` for both
```

### Performance Considerations

With `--rate-limit 1` and 500 markets:
- 2 API calls per market × 500 markets = 1000 requests
- At 1 req/sec = 1000 seconds (~17 minutes) per cycle
- At `--interval 300` (5 min), this won't complete in time → log a warning

Solutions:
- Reduce `--market-limit` to match expected cycle time: 300 seconds / 2 calls per market = 150 markets max at 1 req/sec
- The CLI output should include recommended `--market-limit` based on the current `--rate-limit` and `--interval`

### Logging

Per cycle:
- Start time
- Progress every 50 markets ("150/500 markets, 5.2 min elapsed, ~12 min remaining")
- Cycle complete: markets snapshotted, duration, rows inserted
- Next cycle start time (continuous mode)
- Warnings if cycle exceeds interval

### Error Handling

- If `get_order_book` fails for a single token, skip that market and continue. Do not abort the cycle.
- If the database insert fails, log the error and continue to the next cycle (or exit if single-shot).
- If all markets fail in a cycle, log an error but continue in continuous mode (don't crash).

## Edge Cases

- **Empty markets table**: Exit with "No markets in database. Run 'metadata' collector first."
- **Order book with zero depth**: A market may have no orders on one side. `best_bid` or `best_ask` may be `None` or missing. Store as `NULL` in the database and handle gracefully in logging.
- **New market appears mid-cycle**: The snapshot queries all markets at cycle start. A market added between cycles will be picked up in the next cycle.
- **Long-running continuous mode**: The process may run for days. Ensure database connections are refreshed periodically (reconnect each cycle or use connection pooling).
- **SIGTERM/SIGINT**: In continuous mode, log a shutdown message, finish the current market (not the full cycle), flush remaining inserts, and exit cleanly.

## Verification

- `python -m prediction_markets.collectors order-books --rate-limit 1` runs one cycle and inserts snapshot rows
- Running it again adds new rows with new timestamps (no duplicate errors)
- `python -m prediction_markets.collectors order-books --rate-limit 1 --interval 10 --market-limit 3` runs continuously, cycling every ~10 seconds, snapshoting 3 markets per cycle
- SIGINT during continuous mode exits cleanly
- Database query `SELECT COUNT(*) FROM order_book_snapshots` shows expected number of rows = markets × cycles
