"""Database connection and schema helpers."""

import logging
import os
from contextlib import contextmanager
from pathlib import Path

import psycopg2

logger = logging.getLogger(__name__)


def get_connection():
    """Get a database connection from environment variables."""
    required_vars = [
        "PREDICTION_MARKETS_DB_HOST",
        "PREDICTION_MARKETS_DB_PORT",
        "PREDICTION_MARKETS_DB_USER",
        "PREDICTION_MARKETS_DB_PASSWORD",
        "PREDICTION_MARKETS_DB_NAME",
    ]

    for var in required_vars:
        value = os.environ.get(var)
        if not value:
            logger.error("Missing required environment variable: %s", var)
            raise ValueError(f"Missing required environment variable: {var}")

    host = os.environ["PREDICTION_MARKETS_DB_HOST"]
    port = os.environ["PREDICTION_MARKETS_DB_PORT"]
    user = os.environ["PREDICTION_MARKETS_DB_USER"]
    password = os.environ["PREDICTION_MARKETS_DB_PASSWORD"]
    dbname = os.environ["PREDICTION_MARKETS_DB_NAME"]

    logger.info(
        "Connecting to database %s at %s:%s as user %s",
        dbname,
        host,
        port,
        user,
    )

    return psycopg2.connect(
        host=host,
        port=port,
        user=user,
        password=password,
        dbname=dbname,
    )


@contextmanager
def managed_connection():
    """Context manager for database connections."""
    conn = get_connection()
    try:
        yield conn
        conn.commit()
    except Exception:
        conn.rollback()
        raise
    finally:
        conn.close()


def init_db(conn):
    """Initialize the database schema."""
    schema_path = Path(__file__).with_name("schema.sql").resolve()
    logger.info("Loading schema from %s", schema_path)

    with open(schema_path, "r") as f:
        schema_sql = f.read()

    statements = [s.strip() for s in schema_sql.split(";") if s.strip()]

    with conn.cursor() as cursor:
        for statement in statements:
            cursor.execute(statement)

    logger.info("Executed %d schema statements", len(statements))
