# Story Plan: Market Interface Module

**Parent Plan**: [Polymarket Exploration Plan](polymarket-exploration-plan.md)
**Dependencies**: None
**Parallel With**: Story 1 (Devcontainer Database & Schema)

## Objective

Create a Python module that provides a clean, abstracted interface to Polymarket's APIs with rate limiting and mild future-proofing for Kalshi integration.

## Deliverables

1. `src/prediction_markets/market_interfaces/__init__.py` — public exports
2. `src/prediction_markets/market_interfaces/interface.py` — abstract base class and data classes
3. `src/prediction_markets/market_interfaces/polymarket_client.py` — Polymarket implementation
4. `src/prediction_markets/market_interfaces/rate_limiter.py` — token bucket rate limiter
5. Updated `pyproject.toml` with `polymarket-client` dependency

## Detailed Requirements

### Data Classes (`interface.py`)

Define simple, framework-agnostic dataclasses used throughout the project:

```python
@dataclass
class MarketData:
    id: str              # Gamma market ID
    condition_id: str    # CTF condition ID
    question: str
    slug: str
    token_id_yes: str    # CLOB token ID for Yes
    token_id_no: str     # CLOB token ID for No
    fee_rate: float
    category: str | None
    end_date: str | None  # ISO format
    volume: float
    liquidity: float
    neg_risk: bool
    enable_order_book: bool

@dataclass
class OrderBookLevel:
    price: float
    size: float

@dataclass
class OrderBookData:
    token_id: str
    bids: list[OrderBookLevel]    # sorted best-to-worst
    asks: list[OrderBookLevel]    # sorted best-to-worst

    @property
    def best_bid(self) -> float: ...
    @property
    def best_ask(self) -> float: ...

@dataclass
class PricePoint:
    token_id: str
    timestamp: int       # unix seconds
    price: float          # midpoint
```

### Abstract Interface (`interface.py`)

```python
class MarketClient(ABC):
    @abstractmethod
    def list_active_markets(self) -> list[MarketData]: ...

    @abstractmethod
    def get_order_book(self, token_id: str) -> OrderBookData: ...

    @abstractmethod
    def get_price_history(
        self, token_id: str, start_ts: int, end_ts: int, interval: str
    ) -> list[PricePoint]: ...

    @abstractmethod
    def get_fee_rate(self, condition_id: str) -> float: ...
```

Design rationale: The interface is intentionally narrow. It exposes only what the collectors and analysis need. Cross-market abstractions (Kalshi) will implement the same protocol later.

### Polymarket Implementation (`polymarket_client.py`)

Wraps `PublicClient` from the `polymarket-client` SDK:

- **`__init__(self, rate_limiter: RateLimiter)`**: Creates `PublicClient(environment=PRODUCTION)`. Accepts a shared rate limiter.
- **`list_active_markets()`**: Calls `client.list_markets(closed=False)`, iterates all pages via the paginator's `iter_items()`, filters to `enable_order_book=True`, maps SDK `Market` objects to `MarketData` dataclasses. Calls `get_clob_market_info(condition_id)` per market for fee rates. Uses rate limiter.
- **`get_order_book(token_id)`**: Calls `client.get_order_book(token_id)`, maps to `OrderBookData`.
- **`get_price_history(token_id, start_ts, end_ts, interval)`**: Calls `client.get_price_history(token_id, start_ts=start_ts, end_ts=end_ts, fidelity=... )` or uses `PriceHistoryInterval`. Maps to `PricePoint` list.
- **`get_fee_rate(condition_id)`**: Calls `client.get_clob_market_info(condition_id)` and extracts `fd.r`.

All SDK calls must pass through the rate limiter. The SDK types are never exposed outside this module — only the dataclasses from `interface.py` are returned.

Error handling: wrap SDK exceptions in a custom `MarketClientError` (also in `interface.py`) so callers don't depend on SDK exception types.

### Rate Limiter (`rate_limiter.py`)

Token bucket algorithm:

```python
class RateLimiter:
    def __init__(self, requests_per_second: float): ...
    def acquire(self) -> None:  # blocks until a token is available
    ...

    @contextmanager
    def __call__(self):  # usage: `with rate_limiter: ...`
        self.acquire()
        yield
```

- Thread-safe implementation using `threading.Lock`
- `requests_per_second` can be a float (e.g., 0.5 for one request every 2 seconds)
- `acquire()` computes the required sleep based on elapsed time since last token consumption
- Context manager form is the primary usage pattern for readability

### Dependencies

Add `polymarket-client` to `pyproject.toml`:
```
[project]
dependencies = ["polymarket-client>=0.1.0", ...]
```
Use the `--pre` flag since the unified SDK is in beta: `pip install --pre polymarket-client`.

## Edge Cases

- **API pagination**: `list_markets` returns a paginator. Must handle pages with zero items, very large result sets, and transient network errors between pages.
- **Fee rate lookup**: `get_clob_market_info` may fail for some markets. Fall back to a default fee rate or log a warning.
- **Token IDs**: The SDK returns `clobTokenIds` as a JSON string like `'["0x...", "0x..."]'`. Must parse this to extract Yes and No token IDs.
- **Rate limiter precision**: Short sleep intervals (< 0.01s) may be imprecise. Handle gracefully.
- **Network errors**: SDK calls may fail with connection errors, timeouts, or 429s. Wrap with retry logic (max 3 attempts with exponential backoff) or surface as `MarketClientError`.

## Verification

- `PolymarketClient(rate_limiter).list_active_markets()` returns real market data
- `get_order_book()` returns populated `OrderBookData` with non-empty bids/asks
- `get_price_history()` returns chronological price points
- Rate limiter enforces the configured rate (verify by timing 10 consecutive calls)
- Calling methods without SDK installed or with bad credentials raises `MarketClientError`
