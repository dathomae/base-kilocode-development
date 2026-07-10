"""Tests for database query helper functions."""

import logging
from unittest.mock import MagicMock, patch

import pytest

from prediction_markets.db import queries


class TestInsertMarkets:
    """Tests for insert_markets function."""

    def test_insert_markets_upsert(self):
        """Test that insert_markets generates correct UPSERT SQL and passes rows as parameters."""
        mock_conn = MagicMock()
        mock_cursor = MagicMock()
        mock_conn.cursor.return_value.__enter__ = MagicMock(return_value=mock_cursor)
        mock_conn.cursor.return_value.__exit__ = MagicMock(return_value=False)

        markets = [
            {
                "id": "m1",
                "condition_id": "c1",
                "question": "Will it rain?",
                "slug": "will-it-rain",
                "token_id_yes": "ty1",
                "token_id_no": "tn1",
                "fee_rate": 0.02,
                "category": "weather",
                "end_date": "2025-01-01T00:00:00+00:00",
                "volume": 1000.0,
                "liquidity": 500.0,
                "neg_risk": False,
                "enable_order_book": True,
                "updated_at": "2025-01-01T00:00:00+00:00",
            },
            {
                "id": "m2",
                "condition_id": "c2",
                "question": "Will the stock go up?",
                "slug": "will-stock-go-up",
                "token_id_yes": "ty2",
                "token_id_no": "tn2",
                "fee_rate": 0.03,
                "category": "finance",
                "end_date": "2025-06-01T00:00:00+00:00",
                "volume": 2000.0,
                "liquidity": 1000.0,
                "neg_risk": True,
                "enable_order_book": True,
                "updated_at": "2025-01-01T00:00:00+00:00",
            },
        ]

        queries.insert_markets(mock_conn, markets)

        mock_cursor.execute.assert_called_once()
        call_args = mock_cursor.execute.call_args
        sql = call_args[0][0]
        params = call_args[0][1]

        assert "INSERT INTO markets" in sql
        assert "ON CONFLICT (id) DO UPDATE SET" in sql

        # Verify all columns except created_at are present in the INSERT
        expected_columns = [
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
        for col in expected_columns:
            assert col in sql, f"Column {col} not found in SQL"

        # Verify params contain the right number of rows (2 markets * 14 columns = 28 params)
        assert len(params) == 28
        assert params[0] == "m1"
        assert params[14] == "m2"


class TestInsertOrderBookSnapshot:
    """Tests for insert_order_book_snapshot function."""

    def test_insert_order_book_snapshot_conflict_do_nothing(self):
        """Test that insert_order_book_snapshot uses ON CONFLICT DO NOTHING and parameterized values."""
        mock_conn = MagicMock()
        mock_cursor = MagicMock()
        mock_conn.cursor.return_value.__enter__ = MagicMock(return_value=mock_cursor)
        mock_conn.cursor.return_value.__exit__ = MagicMock(return_value=False)

        queries.insert_order_book_snapshot(
            mock_conn,
            market_id="m1",
            best_bid_yes=0.45,
            best_ask_yes=0.50,
            best_bid_no=0.48,
            best_ask_no=0.52,
            timestamp="2025-01-01T12:00:00+00:00",
        )

        mock_cursor.execute.assert_called_once()
        call_args = mock_cursor.execute.call_args
        sql = call_args[0][0]
        params = call_args[0][1]

        assert "INSERT INTO order_book_snapshots" in sql
        assert "ON CONFLICT (market_id, timestamp) DO NOTHING" in sql
        assert params == ("m1", "2025-01-01T12:00:00+00:00", 0.45, 0.50, 0.48, 0.52)


class TestInsertTokenPrices:
    """Tests for insert_token_prices function."""

    def test_insert_token_prices_conflict_do_nothing(self):
        """Test that insert_token_prices uses ON CONFLICT DO NOTHING and parameterized rows."""
        mock_conn = MagicMock()
        mock_cursor = MagicMock()
        mock_conn.cursor.return_value.__enter__ = MagicMock(return_value=mock_cursor)
        mock_conn.cursor.return_value.__exit__ = MagicMock(return_value=False)

        rows = [
            {"market_id": "m1", "token_id": "ty1", "timestamp": "2025-01-01T12:00:00+00:00", "price": 0.55},
            {"market_id": "m1", "token_id": "tn1", "timestamp": "2025-01-01T12:00:00+00:00", "price": 0.45},
        ]

        queries.insert_token_prices(mock_conn, rows)

        mock_cursor.execute.assert_called_once()
        call_args = mock_cursor.execute.call_args
        sql = call_args[0][0]
        params = call_args[0][1]

        assert "INSERT INTO token_prices" in sql
        assert "ON CONFLICT (market_id, token_id, timestamp) DO NOTHING" in sql

        # Verify all rows are in params (2 rows * 4 columns = 8 params)
        assert len(params) == 8
        assert params[0] == "m1"
        assert params[1] == "ty1"
        assert params[2] == "2025-01-01T12:00:00+00:00"
        assert params[3] == 0.55
        assert params[4] == "m1"
        assert params[5] == "tn1"
        assert params[6] == "2025-01-01T12:00:00+00:00"
        assert params[7] == 0.45


class TestGetMarkets:
    """Tests for get_markets function."""

    def test_get_markets_returns_list_of_dicts(self):
        """Test that get_markets returns a list of dicts keyed by column name."""
        mock_conn = MagicMock()
        mock_cursor = MagicMock()
        mock_conn.cursor.return_value.__enter__ = MagicMock(return_value=mock_cursor)
        mock_conn.cursor.return_value.__exit__ = MagicMock(return_value=False)

        mock_cursor.description = [
            ("id",),
            ("question",),
            ("slug",),
        ]
        mock_cursor.fetchall.return_value = [
            ("m1", "Will it rain?", "will-it-rain"),
            ("m2", "Will the stock go up?", "will-stock-go-up"),
        ]

        result = queries.get_markets(mock_conn)

        mock_cursor.execute.assert_called_once()
        call_args = mock_cursor.execute.call_args
        sql = call_args[0][0]
        assert "SELECT * FROM markets" in sql

        assert result == [
            {"id": "m1", "question": "Will it rain?", "slug": "will-it-rain"},
            {"id": "m2", "question": "Will the stock go up?", "slug": "will-stock-go-up"},
        ]


class TestGetMarketById:
    """Tests for get_market_by_id function."""

    def test_get_market_by_id_returns_single_dict(self):
        """Test that get_market_by_id returns a single dict matching the mocked row."""
        mock_conn = MagicMock()
        mock_cursor = MagicMock()
        mock_conn.cursor.return_value.__enter__ = MagicMock(return_value=mock_cursor)
        mock_conn.cursor.return_value.__exit__ = MagicMock(return_value=False)

        mock_cursor.description = [
            ("id",),
            ("question",),
            ("slug",),
        ]
        mock_cursor.fetchone.return_value = ("m1", "Will it rain?", "will-it-rain")

        result = queries.get_market_by_id(mock_conn, "m1")

        mock_cursor.execute.assert_called_once()
        call_args = mock_cursor.execute.call_args
        sql = call_args[0][0]
        params = call_args[0][1]
        assert "SELECT * FROM markets WHERE id = %s" in sql
        assert params == ("m1",)

        assert result == {"id": "m1", "question": "Will it rain?", "slug": "will-it-rain"}


class TestQueryLogging:
    """Tests for DEBUG-level logging in query functions."""

    @patch("prediction_markets.db.queries.logger")
    def test_query_functions_log_at_debug_level(self, mock_logger):
        """Test that each query function emits a DEBUG-level log containing its function name."""
        mock_conn = MagicMock()
        mock_cursor = MagicMock()
        mock_conn.cursor.return_value.__enter__ = MagicMock(return_value=mock_cursor)
        mock_conn.cursor.return_value.__exit__ = MagicMock(return_value=False)

        mock_cursor.description = [("id",)]
        mock_cursor.fetchall.return_value = []
        mock_cursor.fetchone.return_value = None

        queries.insert_markets(mock_conn, [])
        assert any(
            "insert_markets" in str(call)
            for call in mock_logger.debug.call_args_list
        ), "insert_markets did not log DEBUG with function name"

        queries.insert_order_book_snapshot(mock_conn, "m1", 0.1, 0.2, 0.3, 0.4, "ts")
        assert any(
            "insert_order_book_snapshot" in str(call)
            for call in mock_logger.debug.call_args_list
        ), "insert_order_book_snapshot did not log DEBUG with function name"

        queries.insert_token_prices(mock_conn, [])
        assert any(
            "insert_token_prices" in str(call)
            for call in mock_logger.debug.call_args_list
        ), "insert_token_prices did not log DEBUG with function name"

        queries.get_markets(mock_conn)
        assert any(
            "get_markets" in str(call)
            for call in mock_logger.debug.call_args_list
        ), "get_markets did not log DEBUG with function name"

        queries.get_market_by_id(mock_conn, "m1")
        assert any(
            "get_market_by_id" in str(call)
            for call in mock_logger.debug.call_args_list
        ), "get_market_by_id did not log DEBUG with function name"
