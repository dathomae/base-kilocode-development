# Story002: Market Interface Module

## Goal

Create a Python module that provides a clean, abstracted interface to Polymarket's APIs with rate limiting and mild future-proofing for Kalshi integration. This story implements [Story Plan: Market Interface Module](../plans/story-market-interface-module.md).

## References

- [Story Plan: Market Interface Module](../plans/story-market-interface-module.md)
- [Polymarket Exploration Plan](../plans/polymarket-exploration-plan.md)

## Dependencies

None — This story has no dependencies on other stories and can run in parallel with Story001.

## Dependent Stories

None — No stories have been created that depend on this module.

## Tasks

### Task Sizing

Tasks 1, 2, and 5 are classified as **medium**. Each introduces new abstractions requiring reasoning about data modeling, thread safety, SDK integration, and cross-cutting concerns (retry logic, error wrapping, rate limiter pass-through).

- **Task 1** (13 subtasks): Introduces 4 dataclasses, a MarketClient ABC, and a MarketClientError exception — all new infrastructure, not additive to existing patterns. Although all work is in 2 files, the number of new types and the design of the abstract interface exceed the Small threshold.
- **Task 2** (8 subtasks): Introduces RateLimiter with token bucket algorithm, threading.Lock, and timing precision edge cases. Requires reasoning about thread safety and concurrency.
- **Task 5** (18 subtasks): Introduces PolymarketClient with 4 methods wrapping SDK calls, retry logic with exponential backoff, error wrapping, rate limiter pass-through, pagination handling, token ID parsing, and initialization logging. High subtask count with multiple coordinated concerns.

Tasks 3 and 4 are classified as **[Small]**. Each modifies a single file with a mechanical, additive change following an established pattern.

- **Task 3**: Single file, purely additive re-exports from sibling modules.
- **Task 4**: Single file, single-line dependency addition with one verification step.

The architect-for-story-planning evaluated all tasks for decomposition. Tasks 1 and 5 were borderline on subtask count but are at natural granularity — splitting would either share files across worktrees (creating merge conflicts with no isolation benefit) or create artificial intermediate modules (retry infrastructure extractable only into `_retry.py` which the story design does not call for).

### Parallel Execution

- **Group A** (Tasks 1, 2, 4): Fully disjoint files across all three tasks. Task 1 creates `interface.py` + `test_interface.py`; Task 2 creates `rate_limiter.py` + `test_rate_limiter.py`; Task 4 modifies `pyproject.toml`. No shared files, no compile-time dependencies between these tasks. Each can be independently tested.

### Execution Order

```
[Task 1 ‖ Task 2 ‖ Task 4] → Task 5 → Task 3 → Task 6
```

- **Group A** (Tasks 1, 2, 4) run in parallel — no inter-dependencies.
- **Task 5** depends on Task 1 (imports `MarketClient` and `MarketData` from `interface.py`) and Task 2 (imports `RateLimiter` from `rate_limiter.py`). Must run after Group A merge.
- **Task 3** exports from all three modules — must run after Task 5 merge so its `PolymarketClient` import resolves.
- **Task 6** is manual verification against live APIs — must run after all merges.

### Reintegration

**Group A (Tasks 1, 2, 4):**

- **Merge order**: Task 4 first (modifies `pyproject.toml` — shared build infrastructure), then Tasks 1 and 2 in any order.
- **Integration tests**:
  1. `uv run pytest tests/market_interfaces/ -v --tb=short` — verify all data class tests (Task 1) and rate limiter tests (Task 2) pass together in the merged tree.
  2. `uv run python -c "from prediction_markets.market_interfaces import MarketClient, MarketData, OrderBookData, OrderBookLevel, PricePoint, MarketClientError, RateLimiter"` — verify both modules coexist and are importable in the merged state. The PolymarketClient import is excluded at this stage (Task 5 not yet merged).
- **Watch for conflicts in**:
  - `uv.lock` — Task 4's `uv sync --pre` updates uv.lock with the `polymarket-client` dependency. Tasks 1 and 2 should not modify uv.lock, but if a sub-agent runs `uv sync` or `uv run pytest` with different dependency resolution timing, minor hash differences could appear. Ensure Task 4's uv.lock takes precedence during merge. Run `uv sync` on the story branch after merging to regenerate a consistent uv.lock.
  - `tests/market_interfaces/__init__.py` — Neither task explicitly creates this file, but both create files in `tests/market_interfaces/`. If the test runner or a sub-agent auto-creates an `__init__.py` for test discovery, it could appear in worktree branches. Not expected but worth verifying after merge.

**Task 5 → Task 3 sequential (no parallel group):**

After Task 5 is merged, merge Task 3 and run:
- `uv run python -c "from prediction_markets.market_interfaces import MarketClient, MarketData, OrderBookData, OrderBookLevel, PricePoint, MarketClientError, RateLimiter, PolymarketClient"` — verify full package imports resolve.

---

### Task 1: Create market interface data classes, abstract base class, and exception with tests — Not Started

Create `src/prediction_markets/market_interfaces/interface.py` with four dataclasses, an abstract base class, and a custom exception. Create `tests/market_interfaces/test_interface.py` with unit tests for all types.

a. Write test for `MarketData` dataclass — verify construction with all 14 fields, `None` defaults for `category` and `end_date`, field types — Not Started
b. Write test for `OrderBookLevel` dataclass — verify construction with `price` and `size` fields — Not Started
c. Write test for `OrderBookData` dataclass — verify construction with `token_id`, `bids`, `asks`; verify `best_bid` property returns highest bid price (or 0.0 when empty); verify `best_ask` property returns lowest ask price (or 1.0 when empty) — Not Started
d. Write test for `PricePoint` dataclass — verify construction with `token_id`, `timestamp` (int), `price` fields — Not Started
e. Write test for `MarketClientError` — verify inherits from `Exception`, can be raised and caught with a message — Not Started
f. Write test for `MarketClient` ABC — verify it cannot be instantiated directly, subclasses missing abstract methods fail to instantiate, subclasses implementing all abstract methods instantiate successfully — Not Started
g. Implement `MarketData` dataclass with 14 fields matching the plan specification — Not Started
h. Implement `OrderBookLevel` dataclass with `price: float` and `size: float` — Not Started
i. Implement `OrderBookData` dataclass with `token_id`, `bids`, `asks`, and `best_bid` / `best_ask` properties — Not Started
j. Implement `PricePoint` dataclass with `token_id`, `timestamp: int`, `price: float` — Not Started
k. Implement `MarketClientError` exception inheriting from `Exception` — Not Started
l. Implement `MarketClient` ABC with four `@abstractmethod` methods: `list_active_markets() -> list[MarketData]`, `get_order_book(token_id: str) -> OrderBookData`, `get_price_history(token_id, start_ts, end_ts, interval) -> list[PricePoint]`, `get_fee_rate(condition_id: str) -> float` — Not Started
m. Run `uv run pytest tests/market_interfaces/test_interface.py -v` and verify all tests pass — Not Started

**Acceptance Criteria:**

- `MarketData`, `OrderBookLevel`, `OrderBookData`, `PricePoint` are `@dataclass` classes with correct fields and defaults
- `OrderBookData.best_bid` returns max bid price (0.0 for empty bids)
- `OrderBookData.best_ask` returns min ask price (1.0 for empty asks)
- `MarketClientError` inherits from `Exception`
- `MarketClient` is an ABC that cannot be instantiated directly
- All four abstract methods have correct signatures and return type annotations
- All unit tests pass

### Task 2: Create rate limiter with tests — Not Started

Create `src/prediction_markets/market_interfaces/rate_limiter.py` implementing a thread-safe token bucket rate limiter. Create `tests/market_interfaces/test_rate_limiter.py` with unit tests.

a. Write test verifying `acquire()` enforces the configured rate — time 10 consecutive calls at `requests_per_second=10` and assert elapsed time is approximately 0.9 seconds (±0.2s tolerance) — Not Started
b. Write test verifying the context manager form — use `with rate_limiter:` and verify the block body executes — Not Started
c. Write test verifying thread safety — launch 5 threads each making 5 `acquire()` calls with `requests_per_second=100`, verify total elapsed time is consistent with rate limiting across threads — Not Started
d. Write test verifying `acquire()` handles very small sleep intervals (e.g., `requests_per_second=1000`) without errors — Not Started
e. Implement `RateLimiter.__init__(self, requests_per_second: float)` — stores rate, initializes `_last_consumed` timestamp and `threading.Lock` — Not Started
f. Implement `acquire(self) -> None` — token bucket algorithm: under lock, compute elapsed time since last consume, sleep if needed to enforce rate, update last consume time — Not Started
g. Implement `__call__(self)` as `@contextmanager` — calls `acquire()` then yields — Not Started
h. Run `uv run pytest tests/market_interfaces/test_rate_limiter.py -v` and verify all tests pass — Not Started

**Acceptance Criteria:**

- `RateLimiter(5.0).acquire()` called 10 times takes approximately 1.8 seconds
- Context manager form works: `with rate_limiter: do_work()`
- Thread-safe under concurrent access from multiple threads
- Handles extreme rates (0.5 requests/sec and 1000 requests/sec) without errors
- All unit tests pass

### Task 3: Create market_interfaces package __init__.py with public exports — Not Started [Small]

Create `src/prediction_markets/market_interfaces/__init__.py` with explicit re-exports defining the public API surface.

a. Import and re-export `MarketClient`, `MarketData`, `OrderBookData`, `OrderBookLevel`, `PricePoint`, `MarketClientError` from `.interface` — Not Started
b. Import and re-export `RateLimiter` from `.rate_limiter` — Not Started
c. Import and re-export `PolymarketClient` from `.polymarket_client` — Not Started
d. Verify imports: `python -c "from prediction_markets.market_interfaces import MarketClient, MarketData, RateLimiter, PolymarketClient"` — Not Started

**Acceptance Criteria:**

- All public types exportable from `prediction_markets.market_interfaces`
- `__all__` is defined (optional but recommended)

### Task 4: Add polymarket-client dependency to pyproject.toml — Not Started [Small]

Add `polymarket-client` to the project dependencies in `pyproject.toml`.

a. Add `"polymarket-client>=0.1.0"` to the `dependencies` list under `[project]` — Not Started
b. Run `uv sync --pre` and verify the dependency resolves successfully — Not Started

**Acceptance Criteria:**

- `dependencies = ["polymarket-client>=0.1.0"]` in `pyproject.toml`
- `uv sync --pre` completes without errors
- `python -c "import polymarket_client"` succeeds

### Task 5: Create Polymarket client implementation with tests — Not Started

Create `src/prediction_markets/market_interfaces/polymarket_client.py` implementing the `MarketClient` abstract interface by wrapping the `polymarket-client` SDK. Create `tests/market_interfaces/test_polymarket_client.py` with unit tests that mock SDK calls.

a. Write test for `list_active_markets()` — mock `PublicClient` and its paginator, verify `MarketData` mapping with correct field extraction from SDK `Market` objects, verify filtering of non-order-book markets, verify `clobTokenIds` JSON parsing for yes/no token IDs, verify fee rate lookup via `get_clob_market_info` — Not Started
b. Write test for `get_order_book()` — mock SDK order book response, verify mapping to `OrderBookData` with bids sorted descending and asks sorted ascending — Not Started
c. Write test for `get_price_history()` — mock SDK price history response, verify mapping to list of `PricePoint` objects in chronological order — Not Started
d. Write test for `get_fee_rate()` — mock `get_clob_market_info` response, verify `fd.r` extraction — Not Started
e. Write test for error handling — mock SDK methods to raise `ConnectionError` or `Timeout`, verify `MarketClientError` is raised wrapping the original exception — Not Started
f. Write test for retry logic — mock `list_active_markets` to fail twice with transient errors then succeed on the third attempt, verify the method returns correctly, verify no more than 3 attempts were made — Not Started
g. Write test for rate limiter integration — pass a mock `RateLimiter`, verify `acquire()` is called before (or inside) each SDK method call — Not Started
h. Write test for token ID parsing — verify `clobTokenIds` JSON string `'["0xabc...", "0xdef..."]'` is correctly parsed into yes/no token IDs — Not Started
i. Write test for initialization logging — mock the module logger, create a `PolymarketClient` instance, assert an `INFO`-level log message was emitted containing the environment name (e.g., "PRODUCTION") — Not Started
j. Implement `PolymarketClient.__init__(self, rate_limiter: RateLimiter)` — imports `logging`, creates a module-level logger via `logging.getLogger(__name__)`, creates `PublicClient(environment=PRODUCTION)` from the SDK, logs an `INFO`-level message with the environment name, stores the rate limiter — Not Started
k. Implement `list_active_markets(self) -> list[MarketData]` — use paginator's `iter_items()`, iterate all pages, filter to `enable_order_book=True`, parse `clobTokenIds` JSON string for yes/no token IDs, call `get_clob_market_info(condition_id)` for fee rates. If `get_clob_market_info` fails for a market, log a `WARNING`-level message with the market ID and continue (do not fall back to a default fee rate — omit the market or set fee_rate to None). Map SDK `Market` to `MarketData` dataclass. Wrap in retry/error handling and pass through rate limiter. — Not Started
l. Implement `get_order_book(self, token_id: str) -> OrderBookData` — call SDK, sort bids descending and asks ascending, map to `OrderBookData`. Wrap in retry/error handling and pass through rate limiter. — Not Started
m. Implement `get_price_history(self, token_id: str, start_ts: int, end_ts: int, interval: str) -> list[PricePoint]` — call SDK `get_price_history`, map to `PricePoint` list. Wrap in retry/error handling and pass through rate limiter. — Not Started
n. Implement `get_fee_rate(self, condition_id: str) -> float` — call `get_clob_market_info(condition_id)`, extract `fd.r` from response. Wrap in retry/error handling and pass through rate limiter. — Not Started
o. Implement error wrapping — all SDK calls wrapped in try/except; non-`MarketClientError` exceptions (ConnectionError, Timeout, HTTP errors) are caught and re-raised as `MarketClientError` with the original exception as cause — Not Started
p. Implement retry logic — max 3 attempts with exponential backoff (1s, 2s, 4s) on transient failures; all SDK calls pass through the same retry decorator or helper — Not Started
q. Implement rate limiter pass-through — before each SDK call, invoke `self._rate_limiter.acquire()` (or use context manager form) — Not Started
r. Run `uv run pytest tests/market_interfaces/test_polymarket_client.py -v` and verify all tests pass — Not Started

**Acceptance Criteria:**

- `PolymarketClient` implements `MarketClient` with all four abstract methods
- `list_active_markets()` returns `MarketData` objects with correctly parsed fields
- `get_order_book()` returns `OrderBookData` with sorted bids/asks
- `get_price_history()` returns `PricePoint` objects in chronological order
- `get_fee_rate()` returns the fee rate as a float
- All SDK calls pass through the rate limiter
- SDK exceptions are wrapped in `MarketClientError`
- Transient failures are retried up to 3 attempts with exponential backoff
- Tokens IDs are correctly parsed from `clobTokenIds` JSON strings
- `PolymarketClient.__init__` logs an `INFO`-level message containing the environment name via `logging.getLogger(__name__)`
- `list_active_markets()` logs a `WARNING`-level message if `get_clob_market_info` fails for a market (no default fee rate fallback)
- All unit tests pass with mocked SDK responses

### Task 6: Integration verification against real Polymarket APIs — Not Started

Run the completed module against real Polymarket APIs to verify end-to-end functionality. No code changes — manual verification only.

a. Create a temporary script or use `python -c` to instantiate `PolymarketClient` with `RateLimiter(5.0)` and call `list_active_markets()` — verify returns a non-empty list of `MarketData` objects — Not Started
b. Pick a `token_id` from a market returned in step a, call `get_order_book(token_id)` — verify returns `OrderBookData` with non-empty bids/asks — Not Started
c. Call `get_price_history(token_id, start_ts, end_ts, interval="1h")` with a recent 24-hour window — verify returns a list of `PricePoint` objects in chronological order — Not Started
d. Time 10 consecutive `list_active_markets()` calls with `RateLimiter(5.0)` and verify actual rate stays at or below the configured 5 requests/second — Not Started

**Acceptance Criteria:**

- All four SDK methods return valid data from the live Polymarket API
- Rate limiter enforces the configured rate under real usage
- No unhandled exceptions during real API calls

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks (Tasks 1, 2, and 5), tests are written before implementation. Test subtasks precede implementation subtasks within each task. Tests use `unittest.mock` to patch SDK internals so no running Polymarket API access is required during unit testing.

## Constraints

- SDK types (`polymarket_client` library types) must never be exposed outside the `market_interfaces` package — only dataclasses from `interface.py` are returned
- All SDK calls must pass through the rate limiter (either `acquire()` or context manager form)
- SDK exceptions must be wrapped in `MarketClientError` — callers must not depend on SDK exception types
- Retry logic: max 3 attempts with exponential backoff on transient failures
- `requests_per_second` can be a float (e.g., 0.5 for one request every 2 seconds)
- Rate limiter must be thread-safe using `threading.Lock`
- `pyproject.toml` dependency uses the `--pre` flag since the unified SDK is in beta
- Logging: use Python's standard `logging` module with `logging.getLogger(__name__)` in each module; `INFO` level for initialization/lifecycle events, `WARNING` level for non-fatal operational issues (e.g., fee-rate lookup failures). Do not add DEBUG-level or per-API-call tracing at this time — it will be added later if needed.
- File paths in task descriptions are relative to `source_root` (`/prediction-markets`). Commands like `uv run pytest` run from `source_root`.

## Notes

- The `source_root` for this project is the project root (`/prediction-markets`). All file paths in task descriptions are relative to `source_root`. Commands like `uv run pytest` run from `source_root`.
- This story runs in parallel with Story001 (Devcontainer Database & Schema) — their file sets are fully disjoint.
- The interface is intentionally narrow — only four methods. Cross-market abstractions (Kalshi) will implement the same protocol later.
- Fee rate lookup via `get_clob_market_info` may fail for some markets. The implementation logs a `WARNING`-level message and continues without the market — no default fee rate is substituted.
