# Unified SDK — Overview & Setup

## Installation

```bash
# Using uv
uv add --prerelease allow polymarket-client

# Using pip
pip install --pre polymarket-client
```

The SDK is **beta** — the `--pre` / `--prerelease` flag is needed until a stable release.

## Package Structure

```
polymarket/
├── __init__.py           # Public API: clients, models, types, enums, errors
├── clients/
│   ├── public.py         # PublicClient (sync)
│   ├── async_public.py   # AsyncPublicClient
│   ├── secure.py         # SecureClient (sync)
│   ├── async_secure.py   # AsyncSecureClient
│   └── _transport.py     # SyncTransport, AsyncTransport, header resolvers
├── models/
│   ├── gamma/            # Market, Event, Tag, Series, Comment, Profile
│   ├── clob/             # SignedOrder, OrderBook, OrderResponse, etc.
│   ├── data/             # Position, Trade, Activity, Leaderboard, etc.
│   ├── perps/            # PerpsInstrument, PerpsOrder, PerpsPosition
│   └── rfq.py            # RFQ request/quote/event types
├── streams/              # WebSocket streams
├── _internal/            # Implementation details (actions, dispatch, relayer, EOA)
├── calls.py              # TransactionCall definitions
├── transactions.py       # TransactionHandle classes
├── rfq.py                # RFQ session management
├── auth.py               # ApiKey, BuilderApiKey, RelayerApiKey
├── perps.py              # Perps credentials
├── environments.py       # Production/staging environment configs
├── errors.py             # Error hierarchy
├── pagination.py         # Paginator, AsyncPaginator, Page
├── types.py              # Domain primitives (EvmAddress, HexString, TransactionHash)
└── version.py            # Version info
```

## Design Principles

The SDK follows these principles (from `docs/sdk-direction.md`):

1. **Cohesive interface** — One module surface with stable, idiomatic Python objects
2. **Workflow-oriented** — APIs designed around developer tasks, not API surface area
3. **Sync-first** — Default clients are synchronous; async uses `Async` prefix
4. **Domain types** — `NewType` markers for key identifiers (`MarketId`, `TokenId`, `EvmAddress`, etc.) providing type safety without runtime overhead
5. **Literal for inputs, StrEnum for outputs** — Input parameters use `Literal` types for autocomplete; returned model fields use `StrEnum` for safe comparisons

## Client Classes

| Class | Transport | Auth | Key Methods |
|---|---|---|---|
| `PublicClient` | Sync | None | `get_market()`, `get_event()`, `list_markets()`, `list_positions()`, `get_order_book()`, `get_price()`, `search()`, etc. |
| `AsyncPublicClient` | Async | None | Same as PublicClient but async |
| `SecureClient` | Sync | Private key + API credentials | All PublicClient methods + `place_limit_order()`, `cancel_order()`, `get_balance()`, `execute_transaction()`, `merge_positions()`, etc. |
| `AsyncSecureClient` | Async | Private key + API credentials | Same as SecureClient but async |

## Quickstart

### Public Data (No Auth)

```python
from polymarket import PublicClient

with PublicClient() as client:
    # Get a market by URL
    market = client.get_market(url="https://polymarket.com/event/example")

    # Get the order book
    book = client.get_order_book(token_id=market.clob_token_ids[0])

    # List active events
    paginator = client.list_events(active=True, closed=False, page_size=10)
    for page in paginator:
        for event in page.items:
            print(event.title, event.volume)
```

### Authenticated Trading

```python
from polymarket import SecureClient
import os

client = SecureClient.create(
    private_key=os.environ["POLYMARKET_PRIVATE_KEY"],
)

# Place a limit order
from polymarket import OrderSide
order = client.place_limit_order(
    token_id="1234...",
    side=OrderSide.BUY,
    price="0.50",
    size="10",
)

# Check balance
balance = client.get_balance()
print(balance.usdc_balance, balance.usdc_allowance)
```

### Async Usage

```python
import asyncio
from polymarket import AsyncPublicClient, Market

async def main():
    async with AsyncPublicClient() as client:
        market: Market = await client.get_market(
            url="https://polymarket.com/event/example"
        )
        print(market.question)

asyncio.run(main())
```

## Context Managers

All clients support `with` / `async with` and explicit `close()`:

```python
# Sync
with PublicClient() as client:
    ...
# client is closed here

# Async
async with AsyncPublicClient() as client:
    ...
# client is closed here

# Explicit
client = PublicClient()
data = client.get_market(id="123")
client.close()
```

## Environments

```python
from polymarket import PRODUCTION, Environment

# Default environment
client = PublicClient()  # Uses PRODUCTION

# Custom environment
client = PublicClient(environment=Environment(
    gamma_url="https://gamma-api.polymarket.com",
    data_url="https://data-api.polymarket.com",
    clob_url="https://clob.polymarket.com",
    rfq_url="https://rfq.polymarket.com",
    relayer_url="https://relayer.polymarket.com",
    rpc_url="https://polygon-rpc.com",
))
```

## Domain Types

Key identifiers use `NewType` wrappers for type safety:

```python
from polymarket import (
    EvmAddress,      # Ethereum address (str)
    HexString,        # Hex-encoded string
    TransactionHash,  # Transaction hash (str)
    MarketId,         # Market identifier
    TokenId,          # Token identifier
    ConditionId,      # CTF condition ID
    OrderId,          # Order identifier
)
```

These are plain strings at runtime — no overhead, but IDE autocomplete knows them.

## Errors

```python
from polymarket.errors import (
    PolymarketError,         # Base
    UserInputError,          # Invalid parameters
    RequestRejectedError,    # API rejection (HTTP 4xx/5xx)
    RateLimitError,          # 429
    TimeoutError,            # Request timeout
    TransportError,          # Network/transport failure
    UnexpectedResponseError, # Unparseable response
    SigningError,            # Signature failure
    CancelledSigningError,   # User cancelled signing
    InsufficientAllowanceError,  # Not enough allowance
    InsufficientLiquidityError,  # Can't fill order
    TransactionFailedError,  # Onchain tx failed
)
```

## Pagination

All list methods return `Paginator` instances:

```python
paginator = client.list_markets(closed=False, page_size=10)

# Get first page
page = paginator.first_page()
for market in page.items:
    print(market.question)

# Iterate pages
for page in paginator:
    for market in page.items:
        ...

# Stream all items
for market in paginator.iter_items():
    print(market.question)
```
