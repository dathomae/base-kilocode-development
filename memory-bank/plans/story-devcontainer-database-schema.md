# Story Plan: Devcontainer Database & Schema

**Parent Plan**: [Polymarket Exploration Plan](polymarket-exploration-plan.md)
**Dependencies**: None
**Parallel With**: Story 2 (Market Interface Module)

## Objective

Add a PostgreSQL database service to the existing devcontainer and create the database schema needed for the polymarket exploration.

## Deliverables

1. Updated `.devcontainer/docker-compose.yml` with a `db` service
2. Updated `.devcontainer/.env` if needed (environment variable additions only)
3. Database schema DDL at `src/prediction_markets/db/schema.sql`
4. Database connection module at `src/prediction_markets/db/__init__.py`
5. Query helper functions at `src/prediction_markets/db/queries.py`
6. Updated `pyproject.toml` with `psycopg2-binary` dependency

## Detailed Requirements

### Docker Compose

Add a `db` service to `.devcontainer/docker-compose.yml` following the chronocone pattern:

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

Add the `postgres-data` named volume to the volumes section.

Inject DB connection environment variables into the `app` service using the naming convention `PREDICTION_MARKETS_DB_HOST`, `PREDICTION_MARKETS_DB_PORT`, `PREDICTION_MARKETS_DB_USER`, `PREDICTION_MARKETS_DB_PASSWORD`, `PREDICTION_MARKETS_DB_NAME`. Values should match the `db` service's credentials and service name (`db` as host, `5432` as port).

Add `libpq-dev` to the Dockerfile's apt packages so `psycopg2-binary` can build.

### Schema

Create `src/prediction_markets/db/schema.sql` with three tables:

**`markets`** — Static/reference data about each market:
- `id TEXT PRIMARY KEY` — Gamma market ID
- `condition_id TEXT NOT NULL UNIQUE` — CTF condition ID
- `question TEXT NOT NULL` — market question text
- `slug TEXT NOT NULL` — URL-friendly identifier
- `token_id_yes TEXT NOT NULL UNIQUE` — CLOB token ID for Yes outcome
- `token_id_no TEXT NOT NULL UNIQUE` — CLOB token ID for No outcome
- `fee_rate REAL NOT NULL` — taker fee rate (e.g., 0.04)
- `category TEXT` — category label
- `end_date TIMESTAMPTZ` — resolution/end date (for annualization)
- `volume REAL` — total volume
- `liquidity REAL` — current liquidity
- `neg_risk BOOLEAN NOT NULL DEFAULT FALSE`
- `enable_order_book BOOLEAN NOT NULL DEFAULT TRUE`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`

**`order_book_snapshots`** — Primary data for taker analysis, one row per market per snapshot:
- `id BIGSERIAL PRIMARY KEY`
- `market_id TEXT NOT NULL REFERENCES markets(id)`
- `timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `best_bid_yes REAL NOT NULL`
- `best_ask_yes REAL NOT NULL`
- `best_bid_no REAL NOT NULL`
- `best_ask_no REAL NOT NULL`
- `UNIQUE (market_id, timestamp)`
- Index on `timestamp`, index on `(market_id, timestamp)`

**`token_prices`** — Supplementary data for maker analysis, historical midpoints:
- `id BIGSERIAL PRIMARY KEY`
- `market_id TEXT NOT NULL REFERENCES markets(id)`
- `token_id TEXT NOT NULL` — which token (Yes or No)
- `timestamp TIMESTAMPTZ NOT NULL` — observation time from API
- `price REAL NOT NULL` — midpoint at that time
- `UNIQUE (market_id, token_id, timestamp)`
- Index on `timestamp`, index on `(market_id, timestamp)`

### DB Connection Module

`src/prediction_markets/db/__init__.py`:

- Read connection parameters from environment variables (`PREDICTION_MARKETS_DB_HOST`, etc.)
- Provide a `get_connection()` function returning a `psycopg2` connection
- Provide a context manager (`@contextmanager`) that yields a connection and auto-commits/closes
- Provide an `init_db()` function that reads `schema.sql` and executes it (idempotent — uses `IF NOT EXISTS`)

### Query Functions

`src/prediction_markets/db/queries.py`:

- `insert_markets(conn, markets: list[dict])` — UPSERT (INSERT ON CONFLICT UPDATE) for market metadata
- `insert_order_book_snapshot(conn, market_id, best_bid_yes, best_ask_yes, best_bid_no, best_ask_no, timestamp)` — INSERT with ON CONFLICT DO NOTHING
- `insert_token_prices(conn, rows: list[dict])` — INSERT with ON CONFLICT DO NOTHING
- `get_markets(conn) -> list[dict]` — SELECT all markets for collector usage
- `get_market_by_id(conn, market_id) -> dict` — SELECT single market

Use parameterized queries throughout. No ORM.

### Dependencies

Add `psycopg2-binary` to `pyproject.toml` under `[project].dependencies`.

## Edge Cases

- `init_db()` must be safe to call multiple times (tables already exist)
- Connection failures should raise clear errors with the failed host/port/dbname
- `schema.sql` must be locatable regardless of working directory (use `__file__`-relative path)

## Verification

- `docker compose up` starts both `app` and `db` without errors
- `psql -h db -U prediction_markets -d prediction_markets` connects from the app container
- `init_db()` runs successfully, then runs successfully again (idempotent)
- `insert_markets()` with a test row succeeds, then succeeds again with updated data (UPSERT)
- `insert_order_book_snapshot()` with a duplicate timestamp is silently ignored
