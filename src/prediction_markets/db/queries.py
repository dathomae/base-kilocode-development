"""Database query helper functions."""

import logging

logger = logging.getLogger(__name__)

# Columns for the markets table (excluding created_at which should not be updated)
_MARKET_COLUMNS = [
    "id",
    "condition_id",
    "question",
    "slug",
    "token_id_yes",
    "token_id_no",
    "fee_rate",
    "category",
    "end_date",
    "volume",
    "liquidity",
    "neg_risk",
    "enable_order_book",
    "updated_at",
]


def insert_markets(conn, markets: list[dict]):
    """Insert or update markets using UPSERT.

    On conflict (id), update all columns except id and created_at.
    """
    logger.debug("insert_markets called with %d markets", len(markets))
    if not markets:
        return

    columns = _MARKET_COLUMNS
    col_str = ", ".join(columns)
    placeholders = ", ".join(["%s"] * len(columns))
    updates = ", ".join([f"{col} = EXCLUDED.{col}" for col in columns if col != "id"])

    sql = f"""
        INSERT INTO markets ({col_str})
        VALUES ({placeholders})
        ON CONFLICT (id) DO UPDATE SET {updates}
    """

    rows = []
    for market in markets:
        rows.extend([market[col] for col in columns])

    with conn.cursor() as cursor:
        # Build multi-row VALUES clause
        num_cols = len(columns)
        num_rows = len(markets)
        row_placeholders = ", ".join(
            [f"({placeholders})"] * num_rows
        )
        sql = f"""
            INSERT INTO markets ({col_str})
            VALUES {row_placeholders}
            ON CONFLICT (id) DO UPDATE SET {updates}
        """
        cursor.execute(sql, tuple(rows))


def insert_order_book_snapshot(
    conn,
    market_id,
    best_bid_yes,
    best_ask_yes,
    best_bid_no,
    best_ask_no,
    timestamp,
):
    """Insert an order book snapshot, ignoring duplicates."""
    logger.debug("insert_order_book_snapshot called for market %s", market_id)
    sql = """
        INSERT INTO order_book_snapshots (market_id, timestamp, best_bid_yes, best_ask_yes, best_bid_no, best_ask_no)
        VALUES (%s, %s, %s, %s, %s, %s)
        ON CONFLICT (market_id, timestamp) DO NOTHING
    """
    with conn.cursor() as cursor:
        cursor.execute(sql, (market_id, timestamp, best_bid_yes, best_ask_yes, best_bid_no, best_ask_no))


def insert_token_prices(conn, rows: list[dict]):
    """Insert token prices, ignoring duplicates."""
    logger.debug("insert_token_prices called with %d rows", len(rows))
    if not rows:
        return

    placeholders = "(%s, %s, %s, %s)"
    values_clause = ", ".join([placeholders] * len(rows))
    sql = f"""
        INSERT INTO token_prices (market_id, token_id, timestamp, price)
        VALUES {values_clause}
        ON CONFLICT (market_id, token_id, timestamp) DO NOTHING
    """

    flat_rows = []
    for row in rows:
        flat_rows.extend([row["market_id"], row["token_id"], row["timestamp"], row["price"]])

    with conn.cursor() as cursor:
        cursor.execute(sql, tuple(flat_rows))


def get_markets(conn):
    """Return all markets as a list of dicts keyed by column name."""
    logger.debug("get_markets called")
    with conn.cursor() as cursor:
        cursor.execute("SELECT * FROM markets")
        columns = [desc[0] for desc in cursor.description]
        return [dict(zip(columns, row)) for row in cursor.fetchall()]


def get_market_by_id(conn, market_id):
    """Return a single market dict by id."""
    logger.debug("get_market_by_id called for %s", market_id)
    with conn.cursor() as cursor:
        cursor.execute("SELECT * FROM markets WHERE id = %s", (market_id,))
        columns = [desc[0] for desc in cursor.description]
        row = cursor.fetchone()
        if row is None:
            return None
        return dict(zip(columns, row))
