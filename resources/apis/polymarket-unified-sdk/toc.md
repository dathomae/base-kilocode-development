# Polymarket Unified Python SDK (polymarket-client)

Documentation for the official unified Polymarket Python SDK, combining all REST APIs, trading, wallet operations, and WebSocket streams into one package.

**Package**: `polymarket-client`
**Status**: Beta
**Repository**: [github.com/Polymarket/py-sdk](https://github.com/Polymarket/py-sdk)
**Install**: `pip install --pre polymarket-client`

The unified SDK replaces `py-clob-client-v2` for new projects, providing one coherent, workflow-oriented interface for public data, authenticated account management, trading, builder attribution, onchain transactions, and wallet operations — all from the same module.

## API Documentation

**[overview-and-setup.md](overview-and-setup.md)**
: Installation, package structure, design principles, sync/async clients, domain types, environments, context managers.

**[clients-public.md](clients-public.md)**
: `PublicClient` — Complete reference for public data access: markets, events, prices, orderbooks, positions, trades, activity, search, leaderboards, rewards.

**[clients-secure.md](clients-secure.md)**
: `SecureClient` — Authenticated trading, order creation/posting, wallet operations, onchain transactions, merge/redeem/split, gasless workflows, approvals.

**[clients-async.md](clients-async.md)**
: `AsyncPublicClient` / `AsyncSecureClient` — Async variants with identical APIs, async context managers, asyncio patterns.

**[models.md](models.md)**
: Core data models — Market, Event, Position, Trade, Activity, OrderBook, Tag, Series, Comment, and all domain types (`MarketId`, `TokenId`, `ConditionId`, etc.).

**[clob-types.md](clob-types.md)**
: CLOB-specific types — `SignedOrder`, `OrderResponse`, `OpenOrder`, `OrderBook`, `OrderSide`, `MarketOrderType`, `OrderType`, `CancelOrdersResponse`.

**[perp-types.md](perp-types.md)**
: Perpetuals (perps) types — `PerpsInstrument`, `PerpsOrder`, `PerpsPosition`, `PerpsBook`, `PerpsTrade`, `PerpsTicker`.

**[transactions.md](transactions.md)**
: Transaction system — `TransactionCall`, `TransactionHandle`, `EoaTransactionHandle`, `GaslessTransactionHandle`, merge/redeem/split calls.

**[streams.md](streams.md)**
: WebSocket streams — real-time order book updates, trade feeds, live activity, event streams.

**[rfq-unified.md](rfq-unified.md)**
: RFQ in the unified SDK — `RfqSession`, `RfqEvent` types, quote request/response lifecycle, execution updates.

## Differences from py-clob-client-v2

| Area | py-clob-client-v2 | polymarket-client (unified) |
|---|---|---|
| Scope | CLOB-only | All APIs + WebSockets + transactions + wallet |
| Client model | Single `ClobClient` with L0/L1/L2 modes | `PublicClient` + `SecureClient` (sync + async) |
| Authentication | Manual L1→L2 credential flow | `SecureClient.create()` handles everything automatically |
| Wallet | Manual funder/signature_type | Auto-detected wallet type (EOA, Deposit, Proxy, Safe) |
| Trading | `create_order` + `post_order` | Higher-level order creation with allowance recovery |
| Pagination | Cursor-based loops in client methods | Typed `Paginator` objects with `first_page()`, `iter_items()`, iteration |
| Models | Plain dataclasses | Pydantic models with marked domain types |
| Onchain calls | Manual through order builder | `execute_transaction()`, `merge_positions()`, `redeem_positions()` |
| WebSockets | Not included | `streams` module for live data |

## Related Documentation

- **[resources/apis/polymarket/](../polymarket/toc.md)** — Older `py-clob-client-v2` API docs
- **[resources/polymarket/](../../polymarket/toc.md)** — Polymarket platform concepts
