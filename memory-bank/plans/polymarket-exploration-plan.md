# Polymarket Exploration Plan

## Objective

Evaluate whether buying both sides (Yes + No tokens) of Polymarket trades at a combined cost below $1 minus fees can yield non-trivial annualized returns. No estimation or approximation — every data point used in analysis must come from an actual API observation.

## Background

Every Polymarket binary market has two outcome tokens (Yes and No) that are backed by $1 of pUSD locked in the CTF contract. The split/merge mechanism lets anyone convert $1 into 1 Yes + 1 No (split) and 1 Yes + 1 No back into $1 (merge). If Yes + No tokens can be acquired on the order book for less than $1 minus fees, merging locks in a risk-free profit. This plan describes the empirical exploration to determine whether such arbitrage opportunities exist in practice, how large they are, and whether the annualized returns justify further investment.

This is a Phase 1 exploration — PolyMarket only.

## Two Scenarios, Two Data Sources

The analysis covers two distinct strategies, each with its own data requirements:

| Scenario | What You Pay | Fee | Data Source | Historical? |
|---|---|---|---|---|
| **Taker** | Best ask for Yes + best ask for No | Taker fee (0-0.07%) | `get_order_book` snapshots | No — must accumulate in real time |
| **Maker** | Midpoint of Yes + midpoint of No | $0 (makers pay no fees) | `get_price_history` | Yes — API returns historical midpoints |

Both data sources are fully empirical. The taker scenario uses actual order book asks with no spread estimation. The maker scenario uses actual historical midpoints (the API returns these directly, they aren't derived or estimated).

---

## Plan Outline

### Step 1: Add PostgreSQL Database to Devcontainer

**Objective**: Add a `db` service to the existing `.devcontainer/docker-compose.yml` following the chronocone pattern.

**Deliverables**:
- Updated `.devcontainer/docker-compose.yml` with a `db` service running `postgres:15`
- Named volume (`postgres-data`) for persistence
- Environment variables for DB connection injected to the `app` service (`PREDICTION_MARKETS_DB_HOST`, `_PORT`, `_USER`, `_PASSWORD`, `_NAME`)
- Updated `pyproject.toml` with `psycopg2-binary` dependency

**Reference pattern** (from chronocone):
```yaml
db:
  image: postgres:15
  restart: unless-stopped
  environment:
    POSTGRES_USER: prediction_markets
    POSTGRES_PASSWORD: prediction_markets
    POSTGRES_DB: prediction_markets
  volumes:
    - postgres-data:/var/lib/postgresql/data
  ports:
    - "5432"
  networks:
    - dev-network
```

**Success criteria**: `docker compose up` starts both `app` and `db` containers; `psql` can connect from the app container.

---

### Step 2: Create Database Schema

**Objective**: Design and create tables to store market metadata, order book snapshots (primary, for taker analysis), and historical midpoints (supplementary, for maker analysis).

**Deliverables**:
- `src/prediction_markets/db/schema.sql` — DDL file
- `src/prediction_markets/db/__init__.py` — database connection utilities (connect using env vars, context manager)
- `src/prediction_markets/db/queries.py` — query functions (insert markets, insert snapshots, query for analysis)

**Schema design**:

```sql
-- Core market metadata (populated once, updated periodically)
CREATE TABLE markets (
    id              TEXT PRIMARY KEY,              -- Gamma market ID
    condition_id    TEXT NOT NULL UNIQUE,           -- CTF condition ID
    question        TEXT NOT NULL,
    slug            TEXT NOT NULL,
    token_id_yes    TEXT NOT NULL UNIQUE,           -- CLOB token ID for Yes
    token_id_no     TEXT NOT NULL UNIQUE,           -- CLOB token ID for No
    fee_rate        REAL NOT NULL,                  -- taker fee rate (e.g., 0.04)
    category        TEXT,                           -- category label
    end_date        TIMESTAMPTZ,                    -- resolution date (for annualization)
    volume          REAL,                           -- total volume
    liquidity       REAL,                           -- current liquidity
    neg_risk        BOOLEAN NOT NULL DEFAULT FALSE,
    enable_order_book BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Primary data source for taker analysis: actual order book snapshots
-- One row per market per snapshot, capturing both sides simultaneously
CREATE TABLE order_book_snapshots (
    id              BIGSERIAL PRIMARY KEY,
    market_id       TEXT NOT NULL REFERENCES markets(id),
    timestamp       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    best_bid_yes    REAL NOT NULL,                  -- best bid price for Yes token
    best_ask_yes    REAL NOT NULL,                  -- best ask price for Yes token
    best_bid_no     REAL NOT NULL,                  -- best bid price for No token
    best_ask_no     REAL NOT NULL,                  -- best ask price for No token
    UNIQUE (market_id, timestamp)
);
CREATE INDEX idx_obs_ts ON order_book_snapshots (timestamp);
CREATE INDEX idx_obs_market_ts ON order_book_snapshots (market_id, timestamp);

-- Supplementary data source for maker analysis: historical midpoints
-- Populated via get_price_history (returns actual historical data, not estimates)
CREATE TABLE token_prices (
    id              BIGSERIAL PRIMARY KEY,
    market_id       TEXT NOT NULL REFERENCES markets(id),
    token_id        TEXT NOT NULL,                  -- Yes or No token ID
    timestamp       TIMESTAMPTZ NOT NULL,           -- observation time from API
    price           REAL NOT NULL,                  -- midpoint at that time
    UNIQUE (market_id, token_id, timestamp)
);
CREATE INDEX idx_tp_ts ON token_prices (timestamp);
CREATE INDEX idx_tp_market_ts ON token_prices (market_id, timestamp);
```

**Key design decisions**:
- `order_book_snapshots` stores both Yes and No book data in one row per timestamp, eliminating join complexity during analysis and avoiding timing skew between separate `get_order_book` calls.
- `token_prices` is supplementary — it enables the maker scenario analysis immediately using historical data, but taker analysis depends entirely on accumulated `order_book_snapshots`.
- No spread estimation column — the spread is directly computable as `best_ask - best_bid` from real data.

**Success criteria**: Schema creates cleanly via `psql -f schema.sql`; Python can connect, insert, and query.

---

### Step 3: Polymarket Market Interface Module

**Objective**: Create a Python module for interacting with Polymarket APIs with a mild abstraction layer that anticipates future Kalshi integration.

**Deliverables**:
- `src/prediction_markets/market_interfaces/__init__.py` — exports
- `src/prediction_markets/market_interfaces/interface.py` — abstract base class defining the market data protocol
- `src/prediction_markets/market_interfaces/polymarket_client.py` — Polymarket implementation using `polymarket-client` SDK
- `src/prediction_markets/market_interfaces/rate_limiter.py` — configurable rate limiter (token bucket)

**Interface protocol**:
```python
class MarketClient(ABC):
    """Protocol for prediction market data access."""

    @abstractmethod
    def list_active_markets(self) -> list[MarketData]:
        """Return all active, order-book-enabled binary markets."""

    @abstractmethod
    def get_order_book(self, token_id: str) -> OrderBookData:
        """Return current order book for a token (best bid, best ask, full depth)."""

    @abstractmethod
    def get_price_history(
        self, token_id: str, start_ts: int, end_ts: int, interval: str
    ) -> list[PricePoint]:
        """Return historical price (midpoint) points for a token."""

    @abstractmethod
    def get_fee_rate(self, condition_id: str) -> float:
        """Return the taker fee rate for a market."""
```

**Polymarket implementation**: Wraps `PublicClient` from `polymarket-client` SDK. Maps SDK types (`Market`, `OrderBook`, `PriceHistoryPoint`) to simple dataclasses (`MarketData`, `OrderBookData`, `PricePoint`) defined in `interface.py` to keep the interface decoupled from the SDK.

**Rate limiter**: Token bucket implementation. Configurable requests/second. Used by all API calls in the collector.

**Success criteria**: Can instantiate `PolymarketClient()`, call `list_active_markets()` and see real market data, with rate limiting enforced.

---

### Step 4: Market Data Collection

**Objective**: Collect Polymarket market metadata, order book snapshots, and historical price data, storing everything in PostgreSQL. All collection tools run as CLI commands with configurable rate limiting.

**Deliverables**:
- `src/prediction_markets/collectors/__init__.py`
- `src/prediction_markets/collectors/cli.py` — unified CLI entry point for all collectors
- `src/prediction_markets/collectors/market_metadata.py` — discovers and stores all active markets
- `src/prediction_markets/collectors/order_book_snapshots.py` — periodic order book capture (primary taker data)
- `src/prediction_markets/collectors/price_history.py` — historical midpoint download (supplementary maker data)

**CLI design**:

All collectors are invoked via a unified CLI using `python -m prediction_markets.collectors`:

```
# Discover and store market metadata
python -m prediction_markets.collectors metadata --rate-limit 1

# Take a single order book snapshot of all tracked markets
python -m prediction_markets.collectors order-books --rate-limit 1

# Run continuous order book snapshots at a specified interval
python -m prediction_markets.collectors order-books --rate-limit 1 --interval 300

# Download historical price data for maker analysis
python -m prediction_markets.collectors prices --rate-limit 1 --lookback 90
```

All collectors accept:
- `--rate-limit` (float, default `1.0`): maximum API requests per second
- `--db-url` (string, optional): override the default connection string from env vars

**Workflow**:

1. **Market discovery** (`metadata` subcommand):
   - Call `list_markets(closed=False)` and iterate all pages
   - Filter: `enable_order_book=True`, binary (not neg_risk multi-market)
   - For each market, call `get_clob_market_info(condition_id)` to get fee parameters (`fd.r`)
   - INSERT INTO `markets` (ON CONFLICT UPDATE for refresh)
   - Apply rate limiting between API calls

2. **Order book snapshots** (`order-books` subcommand):
   - Query all markets from the `markets` table
   - For each market, call `get_order_book(token_id_yes)` and `get_order_book(token_id_no)` in rapid succession
   - Extract `best_bid` and `best_ask` from each
   - INSERT one row into `order_book_snapshots` with all four prices
   - When `--interval` is provided, loop continuously: snapshot all markets, sleep for interval seconds, repeat
   - Track per-market timing so the snapshot cycle time is visible in logs
   - Rate limit applies across both token_id calls (2 API calls per market per snapshot)

3. **Price history download** (`prices` subcommand):
   - Query all markets from the `markets` table
   - For each market, call `get_price_history` for both token_id_yes and token_id_no
   - Use `start_ts` derived from `--lookback` (default 90 days)
   - INSERT price points into `token_prices` (ON CONFLICT DO NOTHING for idempotency)
   - Rate limit applies

**Design constraints**:
- No estimation or data transformation beyond extracting fields from API responses
- All timestamps are the actual observation times reported by the API
- The collector does not perform analysis — it only stores raw observations

**Success criteria**:
- `metadata` stores all active markets with fee rates and resolution dates
- `order-books --interval 300` runs continuously, populating `order_book_snapshots` at 5-minute intervals
- `prices --lookback 90` backfills 90 days of hourly midpoint data for all active markets

---

### Step 5: Price Deviation Analysis & Annualized Return

**Objective**: Analyze the collected data to detect arbitrage opportunities and compute annualized returns after fees. This step is purely analytical — it queries the database and produces reports.

**Deliverables**:
- `src/prediction_markets/analysis/__init__.py`
- `src/prediction_markets/analysis/deviation_detector.py` — core detection logic
- `src/prediction_markets/analysis/annualized_return.py` — return calculations
- `src/prediction_markets/analysis/report.py` — results output (console + CSV)

**Taker scenario** (uses `order_book_snapshots`):

For each snapshot row, compute directly from observed data:

```
ask_yes = best_ask_yes
ask_no  = best_ask_no

fee_yes = 1 × fee_rate × ask_yes × (1 - ask_yes)
fee_no  = 1 × fee_rate × ask_no  × (1 - ask_no)

total_cost = ask_yes + ask_no + fee_yes + fee_no

is_opportunity = (total_cost < 1.0)
gross_profit   = 1.0 - total_cost
```

Annualized return:
```
days_to_resolution = max((markets.end_date - snapshot.timestamp).days, 1)
annualized_return  = (gross_profit / total_cost) × (365 / days_to_resolution)
```

An "immediate merge" scenario (selling immediately rather than waiting for resolution) can also be computed by omitting the time factor and reporting the absolute return.

**Maker scenario** (uses `token_prices`):

No fees, no spread — the maker places limit orders at the midpoint and pays zero fees if filled:

```
mid_yes = price_yes   -- directly from token_prices
mid_no  = price_no    -- directly from token_prices

maker_cost = mid_yes + mid_no

is_opportunity = (maker_cost < 1.0)
gross_profit   = 1.0 - maker_cost

annualized_return = (gross_profit / maker_cost) × (365 / days_to_resolution)
```

**Output**:
- Summary statistics for both scenarios:
  - Total snapshot observations
  - Number and percentage of observations where an opportunity existed
  - Distribution of gross profits (min, max, mean, median, percentiles)
  - Distribution of annualized returns
  - Breakdown by market category and fee rate
- Top N opportunities ranked by annualized return
- CSV export of all detected opportunities with market details
- Timeseries chart data: `ask_yes + ask_no` for a sample of markets over the observation period, to visualize how the sum deviates from $1 over time

**Success criteria**: Analysis produces a clear, data-backed answer to the hypothesis:
- How often does `ask_yes + ask_no + fees < 1.0` actually occur in observed order books?
- How often does `mid_yes + mid_no < 1.0` occur in historical midpoints?
- What is the distribution of annualized returns?
- Which market categories show the most promise?
- Is the magnitude sufficient to justify automated trading?

---

## Migration: Raw SQL Files

Following the chronocone pattern, schema migrations are managed as raw SQL files (`schema.sql`) rather than using Alembic or another migration tool. This keeps things simple for an exploration project. If the project grows, a migration tool can be introduced later.

## Implementation Order

| Step | Description | Dependencies |
|------|-------------|-------------|
| 1 | PostgreSQL in Devcontainer | None |
| 2 | Database Schema | Step 1 |
| 3 | Polymarket Market Interface | None (can parallel Step 1-2) |
| 4 | Market Data Collection | Steps 1, 2, 3 |
| 5 | Price Deviation Analysis | Steps 1, 2, 4 |

Steps 1-3 can proceed in parallel. Steps 4-5 are serial and depend on 1-3.

## Potential Risks and Questions

1. **Taker data requires accumulation time**: Order book snapshots must be collected over days/weeks before the taker analysis is meaningful. The maker analysis can proceed immediately using historical midpoint data. _Mitigation_: Start the order book collector immediately after Step 4 is implemented and let it accumulate data while proceeding to Step 5 (which can run incrementally).

2. **API rate limits**: The Gamma API rate limits are undocumented. We may need to tune the rate limiter based on empirical testing. _Mitigation_: Default to 1 req/sec (conservative). The `--rate-limit` flag allows experimentation to find the highest sustainable rate.

3. **Data volume — order_book_snapshots**: 500 markets × 12 snapshots/hour × 24 hours × 30 days = ~4.3M rows. This is within PostgreSQL capabilities but queries on the full table will benefit from the timestamp index. _Mitigation_: The collector can be limited to a subset of markets (e.g., top N by volume) via a `--market-limit` option if needed.

4. **Data volume — token_prices**: 500 markets × 2 tokens × 720 hourly points (30 days) = 720K rows. Well within PostgreSQL capabilities without partitioning.

5. **Merge gas costs**: Merging tokens on-chain costs Polygon gas (~$0.01-0.05 per merge). This is negligible for positions above ~$10 but should be noted in the analysis output. _Mitigation_: Report gas costs as a separate line item.

6. **Maker fill probability**: The maker scenario assumes limit orders at the midpoint will eventually fill. In practice, fills depend on market activity and competing orders. _Mitigation_: Report maker and taker scenarios separately with clear assumptions about fill probability. The maker scenario represents a ceiling on what's theoretically achievable.
