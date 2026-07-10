# ClobClient — Complete Method Reference

## Constructor

```python
ClobClient(
    host: str,                    # Base URL (e.g., "https://clob.polymarket.com")
    chain_id: int,                # 137 (Polygon) or 80002 (Amoy testnet)
    key: Optional[str] = None,    # Private key for signing
    creds: Optional[ApiCreds] = None,  # API credentials (apiKey, secret, passphrase)
    signature_type: Optional[int] = None,  # 0=EOA, 1=POLY_PROXY, 2=GNOSIS_SAFE, 3=POLY_1271
    funder: Optional[str] = None,          # Wallet address holding funds
    builder_config: Optional[BuilderConfig] = None,  # Builder fee attribution
    use_server_time: bool = False,  # Whether to fetch server time for timestamps
    retry_on_error: bool = False,   # Retry POST requests on failure
    fee_slippage: float = 0,        # Fee slippage percentage (0 or 1-100)
)
```

**Note**: The `key` is always the signing key (owner key or session signer). The `funder` is where funds are held — for EOA (type 0) they're the same; for deposit wallets (type 3) the funder is the deposit wallet address.

### Key Attributes

- `.rfq` — Access to `RfqClient` for Request for Quote operations
- `.builder` — Access to `OrderBuilder` for order construction/signing

---

## Public Data Methods (L0) — No Auth Required

### Health & Meta

| Method | Description |
|---|---|
| `get_ok()` | Health check — GET `/ok` |
| `get_server_time()` | Server time — GET `/time` |
| `get_version() -> int` | API version (1 or 2). Cached after first call. |

### Markets & Orderbooks

| Method | Signature | Description |
|---|---|---|
| `get_markets(next_cursor)` | `-> dict` | Paged list of all markets |
| `get_market(condition_id)` | `-> dict` | Single market by condition ID |
| `get_clob_market_info(condition_id)` | `-> dict` | Market info (tokens, tick size, neg risk, fees). Also populates caches. |
| `get_simplified_markets(next_cursor)` | `-> dict` | Simplified market list |
| `get_sampling_markets(next_cursor)` | `-> dict` | Sampling-eligible markets |
| `get_sampling_simplified_markets(next_cursor)` | `-> dict` | Simplified sampling-eligible markets |
| `get_order_book(token_id)` | `-> dict` | Order book for a token |
| `get_order_books(params: list[BookParams])` | `-> list` | Order books for multiple tokens |
| `get_order_book_hash(orderbook)` | `-> str` | Generate orderbook summary hash |

### Prices & Spreads

| Method | Signature | Description |
|---|---|---|
| `get_price(token_id, side)` | `-> dict` | Best price for token (BUY/SELL) |
| `get_prices(params: list[BookParams])` | `-> list` | Prices for multiple tokens |
| `get_midpoint(token_id)` | `-> dict` | Midpoint price |
| `get_midpoints(params: list[BookParams])` | `-> list` | Midpoints for multiple tokens |
| `get_spread(token_id)` | `-> dict` | Bid-ask spread |
| `get_spreads(params: list[BookParams])` | `-> list` | Spreads for multiple tokens |
| `get_last_trade_price(token_id)` | `-> dict` | Last trade price |
| `get_last_trades_prices(params: list[BookParams])` | `-> list` | Last trade prices for multiple tokens |
| `get_prices_history(params: PricesHistoryParams)` | `-> dict` | Historical price data (requires interval or start_ts/end_ts) |

### Market Property Resolvers (with caching)

| Method | Signature | Description |
|---|---|---|
| `get_tick_size(token_id) -> TickSize` | `-> str` | Minimum tick size. Cached after first call. |
| `get_neg_risk(token_id) -> bool` | `-> bool` | Whether market uses neg risk. Cached. |
| `get_fee_rate_bps(token_id) -> int` | `-> int` | Fee rate in basis points |
| `get_fee_exponent(token_id) -> float` | `-> float` | Fee exponent |

### Rewards

| Method | Signature | Description |
|---|---|---|
| `get_current_rewards()` | `-> list` | Current rewards markets (paged) |
| `get_raw_rewards_for_market(condition_id)` | `-> list` | Raw rewards data per market (paged) |

### Market Price Calculation

| Method | Signature | Description |
|---|---|---|
| `calculate_market_price(token_id, side, amount, order_type)` | `-> float` | Calculate the expected fill price for a market order |

### Market Trades Events

| Method | Signature | Description |
|---|---|---|
| `get_market_trades_events(condition_id)` | `-> dict` | Trade events for a market — GET `/live-activity/{condition_id}` |

---

## L1 Authentication Methods

### API Key Management

| Method | Signature | Description |
|---|---|---|
| `create_api_key(nonce=None) -> ApiCreds` | POST `/auth/api-key` | Create new API credentials |
| `derive_api_key(nonce=None) -> ApiCreds` | GET `/auth/derive-api-key` | Derive existing API credentials |
| `create_or_derive_api_key(nonce=None) -> ApiCreds` | Create → fallback to derive | Get/create credentials in one call |

### Address

| Method | Description |
|---|---|
| `get_address() -> str` | Get the signer's Ethereum address |

---

## L2 Authenticated Methods

### Order Creation (Signing)

| Method | Signature | Description |
|---|---|---|
| `create_order(order_args, options)` | `-> SignedOrderV1/V2` | Build and sign a limit order (does NOT post) |
| `create_market_order(order_args, options)` | `-> SignedOrderV1/V2` | Build and sign a market order (does NOT post) |
| `create_and_post_order(order_args, options, order_type, post_only, defer_exec)` | `-> dict` | Create + post a limit order (convenience) |
| `create_and_post_market_order(order_args, options, order_type, defer_exec)` | `-> dict` | Create + post a market order (convenience) |

### Posting Orders

| Method | Signature | Description |
|---|---|---|
| `post_order(order, order_type, post_only, defer_exec)` | `-> dict` | Post a signed order to the CLOB |
| `post_orders(args: list, post_only, defer_exec)` | `-> list` | Post multiple signed orders |

### Order Management

| Method | Signature | Description |
|---|---|---|
| `get_order(order_id)` | `-> dict` | Get a single order by ID |
| `get_open_orders(params, only_first_page, next_cursor)` | `-> list` | Get open orders (paged by default) |
| `get_pre_migration_orders(only_first_page, next_cursor)` | `-> list` | Get orders from pre-V2 migration |
| `cancel_order(payload: OrderPayload)` | `-> dict` | Cancel a single order by orderID |
| `cancel_orders(order_hashes: list)` | `-> dict` | Cancel multiple orders by hash |
| `cancel_all()` | `-> dict` | Cancel all open orders |
| `cancel_market_orders(payload: OrderMarketCancelParams)` | `-> dict` | Cancel all orders for a market/asset |

### Trade History

| Method | Signature | Description |
|---|---|---|
| `get_trades(params, only_first_page, next_cursor)` | `-> list` | Get trade history (paged) |
| `get_trades_paginated(params, next_cursor)` | `-> dict` | Get trades with pagination metadata |

### Balance & Allowance

| Method | Signature | Description |
|---|---|---|
| `get_balance_allowance(params)` | `-> dict` | Check balance and allowance |
| `update_balance_allowance(params)` | `-> dict` | Sync balance/allowance with onchain state |

### Notifications

| Method | Signature | Description |
|---|---|---|
| `get_notifications()` | `-> dict` | Get order notifications |
| `drop_notifications(params)` | `-> dict` | Drop/acknowledge notifications |

### API Key Management (L2)

| Method | Signature | Description |
|---|---|---|
| `get_api_keys()` | `-> dict` | List API keys |
| `get_closed_only_mode()` | `-> dict` | Check if closed-only mode is active |
| `delete_api_key()` | `-> dict` | Delete the current API key |

### Builder API Keys

| Method | Signature | Description |
|---|---|---|
| `create_builder_api_key()` | `-> dict` | Create a builder API key |
| `get_builder_api_keys()` | `-> dict` | List builder API keys |
| `revoke_builder_api_key()` | `-> dict` | Revoke a builder API key |

### Readonly API Keys

| Method | Signature | Description |
|---|---|---|
| `create_readonly_api_key()` | `-> dict` | Create a readonly API key |
| `get_readonly_api_keys()` | `-> dict` | List readonly API keys |
| `delete_readonly_api_key(key: str)` | `-> dict` | Delete a readonly API key |

### Builder Trades

| Method | Signature | Description |
|---|---|---|
| `get_builder_trades(params, next_cursor)` | `-> dict` | Get trades attributed to a builder code (pagination metadata) |

### Order Scoring & Earnings

| Method | Signature | Description |
|---|---|---|
| `is_order_scoring(params)` | `-> dict` | Check if a specific order is scoring |
| `are_orders_scoring(params)` | `-> dict` | Check if multiple orders are scoring |
| `get_earnings_for_user_for_day(date)` | `-> list` | Earnings for a date (paged) |
| `get_total_earnings_for_user_for_day(date)` | `-> dict` | Total earnings for a date |
| `get_user_earnings_and_markets_config(date, order_by, position, no_competition)` | `-> list` | Earnings and market configs (paged) |
| `get_reward_percentages()` | `-> dict` | Liquidity reward percentages |

### Heartbeat

| Method | Signature | Description |
|---|---|---|
| `post_heartbeat(heartbeat_id)` | `-> dict` | Post a heartbeat signal |

---

## Internal/Protected Methods

| Method | Description |
|---|---|
| `_get_client_mode() -> int` | Determine L0/L1/L2 from signer + creds presence |
| `assert_level_1_auth()` | Raise if no signer |
| `assert_level_2_auth()` | Raise if no signer or creds |
| `_get_timestamp() -> Optional[int]` | Get server time if `use_server_time` is True |
| `_l1_headers(nonce) -> dict` | Generate L1 authentication headers |
| `_l2_headers(method, endpoint, body) -> dict` | Generate L2 HMAC authentication headers |
| `__resolve_tick_size(token_id, tick_size) -> TickSize` | Resolve tick size (validate against market minimum) |
| `__resolve_version(force_update) -> int` | Get/cache API version |
| `__ensure_market_info_cached(token_id)` | Ensure fee info is cached for a token |
| `__ensure_builder_fee_rate_cached(builder_code)` | Ensure builder fee rate is cached |
| `_adjust_buy_amount_for_balance(...)` | Adjust buy amount for fee deduction from balance |
| `_is_order_version_mismatch(resp) -> bool` | Check if response indicates version mismatch |
| `_retry_on_version_update(func)` | Retry order posting if version changed |
