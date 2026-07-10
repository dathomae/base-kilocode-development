-- Database schema for prediction markets data

-- Markets table: stores metadata for each prediction market
CREATE TABLE IF NOT EXISTS markets (
    id                  TEXT PRIMARY KEY,
    condition_id        TEXT NOT NULL UNIQUE,
    question            TEXT NOT NULL,
    slug                TEXT NOT NULL,
    token_id_yes        TEXT NOT NULL UNIQUE,
    token_id_no         TEXT NOT NULL UNIQUE,
    fee_rate            REAL NOT NULL,
    category            TEXT,
    end_date            TIMESTAMPTZ,
    volume              REAL,
    liquidity           REAL,
    neg_risk            BOOLEAN NOT NULL DEFAULT FALSE,
    enable_order_book   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Order book snapshots table: stores best bid/ask prices at a given timestamp
CREATE TABLE IF NOT EXISTS order_book_snapshots (
    id              BIGSERIAL PRIMARY KEY,
    market_id       TEXT NOT NULL REFERENCES markets(id),
    timestamp       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    best_bid_yes    REAL NOT NULL,
    best_ask_yes    REAL NOT NULL,
    best_bid_no     REAL NOT NULL,
    best_ask_no     REAL NOT NULL,
    UNIQUE (market_id, timestamp)
);

CREATE INDEX IF NOT EXISTS idx_order_book_snapshots_timestamp
    ON order_book_snapshots(timestamp);

CREATE INDEX IF NOT EXISTS idx_order_book_snapshots_market_timestamp
    ON order_book_snapshots(market_id, timestamp);

-- Token prices table: stores token price history
CREATE TABLE IF NOT EXISTS token_prices (
    id          BIGSERIAL PRIMARY KEY,
    market_id   TEXT NOT NULL REFERENCES markets(id),
    token_id    TEXT NOT NULL,
    timestamp   TIMESTAMPTZ NOT NULL,
    price       REAL NOT NULL,
    UNIQUE (market_id, token_id, timestamp)
);

CREATE INDEX IF NOT EXISTS idx_token_prices_timestamp
    ON token_prices(timestamp);

CREATE INDEX IF NOT EXISTS idx_token_prices_market_timestamp
    ON token_prices(market_id, timestamp);
