"""Tests for database connection module."""

import logging
import os
from contextlib import contextmanager
from pathlib import Path
from unittest.mock import MagicMock, mock_open, patch

import pytest

from prediction_markets.db import get_connection, init_db, managed_connection


class TestGetConnection:
    """Tests for get_connection function."""

    @patch.dict(
        os.environ,
        {
            "PREDICTION_MARKETS_DB_HOST": "testhost",
            "PREDICTION_MARKETS_DB_PORT": "5433",
            "PREDICTION_MARKETS_DB_USER": "testuser",
            "PREDICTION_MARKETS_DB_PASSWORD": "testpass",
            "PREDICTION_MARKETS_DB_NAME": "testdb",
        },
        clear=True,
    )
    @patch("prediction_markets.db.psycopg2.connect")
    def test_get_connection_reads_env_vars(self, mock_connect):
        """Test that get_connection reads environment variables and calls psycopg2.connect correctly."""
        mock_connect.return_value = MagicMock()

        conn = get_connection()

        mock_connect.assert_called_once_with(
            host="testhost",
            port="5433",
            user="testuser",
            password="testpass",
            dbname="testdb",
        )
        assert conn is mock_connect.return_value

    @patch.dict(os.environ, {}, clear=True)
    def test_get_connection_raises_on_missing_env_vars(self):
        """Test that get_connection raises a clear error naming the missing variable."""
        required_vars = [
            "PREDICTION_MARKETS_DB_HOST",
            "PREDICTION_MARKETS_DB_PORT",
            "PREDICTION_MARKETS_DB_USER",
            "PREDICTION_MARKETS_DB_PASSWORD",
            "PREDICTION_MARKETS_DB_NAME",
        ]

        for missing_var in required_vars:
            env_vars = {
                "PREDICTION_MARKETS_DB_HOST": "testhost",
                "PREDICTION_MARKETS_DB_PORT": "5432",
                "PREDICTION_MARKETS_DB_USER": "testuser",
                "PREDICTION_MARKETS_DB_PASSWORD": "testpass",
                "PREDICTION_MARKETS_DB_NAME": "testdb",
            }
            del env_vars[missing_var]

            with patch.dict(os.environ, env_vars, clear=True):
                with pytest.raises(ValueError) as exc_info:
                    get_connection()

                assert missing_var in str(exc_info.value)


class TestManagedConnection:
    """Tests for the managed_connection context manager."""

    @patch("prediction_markets.db.get_connection")
    def test_context_manager_commits_and_closes(self, mock_get_connection):
        """Test that the context manager commits on success and closes on both success and failure."""
        mock_conn = MagicMock()
        mock_get_connection.return_value = mock_conn

        with managed_connection() as conn:
            assert conn is mock_conn

        mock_conn.commit.assert_called_once()
        mock_conn.close.assert_called_once()

    @patch("prediction_markets.db.get_connection")
    def test_context_manager_closes_on_failure(self, mock_get_connection):
        """Test that the context manager closes the connection even when an exception occurs."""
        mock_conn = MagicMock()
        mock_get_connection.return_value = mock_conn

        with pytest.raises(RuntimeError):
            with managed_connection() as conn:
                assert conn is mock_conn
                raise RuntimeError("Test error")

        mock_conn.commit.assert_not_called()
        mock_conn.close.assert_called_once()


class TestInitDb:
    """Tests for init_db function."""

    @patch("prediction_markets.db.__file__", "/fake/path/db/__init__.py")
    def test_init_db_idempotent(self):
        """Test that init_db is safe to call multiple times."""
        mock_conn = MagicMock()
        mock_cursor = MagicMock()
        mock_conn.cursor.return_value.__enter__ = MagicMock(return_value=mock_cursor)
        mock_conn.cursor.return_value.__exit__ = MagicMock(return_value=False)

        schema_sql = (
            "CREATE TABLE IF NOT EXISTS t1 (id INT);"
            "CREATE TABLE IF NOT EXISTS t2 (id INT);"
        )

        with patch("builtins.open", mock_open(read_data=schema_sql)):
            init_db(mock_conn)
            init_db(mock_conn)

        # Each call should execute both statements, so 4 total execute calls
        assert mock_cursor.execute.call_count == 4

    @patch("prediction_markets.db.__file__", "/fake/path/db/__init__.py")
    def test_init_db_logs_schema_path(self):
        """Test that init_db logs the resolved schema file path at INFO level."""
        mock_conn = MagicMock()
        mock_cursor = MagicMock()
        mock_conn.cursor.return_value.__enter__ = MagicMock(return_value=mock_cursor)
        mock_conn.cursor.return_value.__exit__ = MagicMock(return_value=False)

        schema_sql = "CREATE TABLE IF NOT EXISTS t1 (id INT);"
        expected_path = str(Path("/fake/path/db/schema.sql").resolve())

        mock_logger = MagicMock()
        with patch("builtins.open", mock_open(read_data=schema_sql)):
            with patch("prediction_markets.db.logger", mock_logger):
                init_db(mock_conn)

        info_calls = [call for call in mock_logger.info.call_args_list]
        logged_paths = [
            str(call)
            for call in info_calls
            if expected_path in str(call)
        ]
        assert len(logged_paths) >= 1, f"Expected INFO log with schema path {expected_path}, got: {info_calls}"

        # Verify statement count was logged at INFO level
        logged_counts = [
            str(call)
            for call in info_calls
            if "Executed" in str(call) and "schema statements" in str(call)
        ]
        assert len(logged_counts) >= 1, f"Expected INFO log with statement count, got: {info_calls}"
