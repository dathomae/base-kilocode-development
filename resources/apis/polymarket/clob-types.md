# CLOB Types — All Dataclass Definitions

All dataclasses from `py_clob_client_v2.clob_types`.

## Core Credentials & Request Types

| Type | Fields | Description |
|---|---|---|
| `ApiCreds` | `api_key`, `api_secret`, `api_passphrase` | API credentials for L2 authentication |
| `RequestArgs` | `method`, `request_path`, `body`, `serialized_body` | HTTP request description for header signing |

## Order Input Types

### Limit Orders

| Type | Fields | Description |
|---|---|---|
| `OrderArgsV1` | `token_id`, `price`, `size`, `side`, `expiration=0`, `fee_rate_bps=0`, `nonce=0`, `taker=ZERO_ADDRESS`, `builder_code=BYTES32_ZERO` | V1 (legacy) limit order args |
| `OrderArgsV2` | `token_id`, `price`, `size`, `side`, `expiration=0`, `builder_code=BYTES32_ZERO`, `metadata=BYTES32_ZERO`, `user_usdc_balance=None` | V2 limit order args (default) |
| `OrderArgs` | Alias for `OrderArgsV2` | Shortcut for V2 |

**V1 vs V2 differences**: V2 removes `fee_rate_bps`, `nonce`, and `taker` fields. V2 adds `metadata` and `user_usdc_balance` (for automated fee-adjusted size reduction).

### Market Orders

| Type | Fields | Description |
|---|---|---|
| `MarketOrderArgsV1` | `token_id`, `amount`, `side`, `price=0`, `order_type=FOK`, `fee_rate_bps=0`, `nonce=0`, `taker=ZERO_ADDRESS`, `builder_code=BYTES32_ZERO` | V1 market order args |
| `MarketOrderArgsV2` | `token_id`, `amount`, `side`, `price=0`, `order_type=FOK`, `user_usdc_balance=0`, `builder_code=BYTES32_ZERO`, `metadata=BYTES32_ZERO` | V2 market order args |
| `MarketOrderArgs` | Alias for `MarketOrderArgsV2` | Shortcut for V2 |

## Book & Order Parameters

| Type | Fields | Description |
|---|---|---|
| `BookParams` | `token_id`, `side=""` | Parameters for book/prices queries. Accepts Side int or string. |
| `TradeParams` | `id`, `maker_address`, `market`, `asset_id`, `before`, `after` | Filters for trade history |
| `OpenOrderParams` | `id`, `market`, `asset_id` | Filters for open orders |
| `DropNotificationParams` | `ids: Optional[list]` | Notification IDs to drop |
| `OrderScoringParams` | `orderId: str` | Single order scoring check |
| `OrdersScoringParams` | `orderIds: list` | Multiple orders scoring check |
| `OrderPayload` | `orderID: str` | Order ID for cancellation |
| `OrderMarketCancelParams` | `market`, `asset_id` | Cancel by market or asset |
| `EarningsParams` | `date`, `market` | Earnings query parameters |
| `RewardsMarketsParams` | `condition_id`, `next_cursor` | Rewards markets pagination |

## Balance & Asset Types

| Type | Fields | Description |
|---|---|---|
| `BalanceAllowanceParams` | `asset_type`, `token_id`, `signature_type=-1` | Balance/allowance query params |
| `AssetType` | `COLLATERAL = "COLLATERAL"`, `CONDITIONAL = "CONDITIONAL"` | Asset type enum |

## Order Options

| Type | Fields | Description |
|---|---|---|
| `CreateOrderOptions` | `tick_size: TickSize`, `neg_risk: bool` | Required options for creating an order |
| `PartialCreateOrderOptions` | `tick_size: Optional[TickSize]=None`, `neg_risk: Optional[bool]=None` | Optional overrides (fetched from market if not provided) |

## Tick Size & Rounding

`TickSize = Literal["0.1", "0.01", "0.005", "0.0025", "0.001", "0.0001"]`

| Type | Fields | Description |
|---|---|---|
| `RoundConfig` | `price: float`, `size: float`, `amount: float` | Decimal precision for price/size/amount rounding per tick size |

## Order Types

| Type | Values | Description |
|---|---|---|
| `OrderType` | `GTC`, `FOK`, `GTD`, `FAK` | Standard order type constants |

## Price History Types

| Type | Fields/Values | Description |
|---|---|---|
| `PriceHistoryInterval` | `MAX="max"`, `ONE_WEEK="1w"`, `ONE_DAY="1d"`, `SIX_HOURS="6h"`, `ONE_HOUR="1h"` | Preset intervals for price history |
| `PricesHistoryParams` | `market`, `start_ts`, `end_ts`, `fidelity`, `interval` | Parameters for `get_prices_history()` |

## Contract Configuration

| Type | Fields | Description |
|---|---|---|
| `ContractConfig` | `exchange`, `neg_risk_adapter`, `neg_risk_exchange`, `collateral`, `conditional_tokens`, `exchange_v2`, `neg_risk_exchange_v2` | Contract addresses for a chain |

### Polygon Mainnet (chain_id=137)

| Contract | Address |
|---|---|
| Exchange V1 | `0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E` |
| Neg Risk Exchange V1 | `0xC5d563A36AE78145C45a50134d48A1215220f80a` |
| Neg Risk Adapter | `0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296` |
| Collateral | `0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB` |
| Conditional Tokens | `0x4D97DCd97eC945f40cF65F87097ACe5EA0476045` |
| Exchange V2 | `0xE111180000d2663C0091e4f400237545B87B996B` |
| Neg Risk Exchange V2 | `0xe2222d279d744050d28e00520010520000310F59` |

## Builder Types

| Type | Fields | Description |
|---|---|---|
| `BuilderConfig` | `builder_address`, `builder_code=BYTES32_ZERO` | Builder identity for fee attribution |
| `BuilderFeeRate` | `maker: float=0`, `taker: float=0` | Builder fee rates (maker/taker) |
| `BuilderApiKey` | `key`, `secret`, `passphrase` | Created builder API key |
| `BuilderApiKeyResponse` | `key`, `created_at`, `revoked_at` | Builder API key from list |
| `BuilderTradeParams` | `builder_code`, `id`, `maker_address`, `market`, `asset_id`, `before`, `after` | Builder trade query params |

## Market Response Types

| Type | Fields | Description |
|---|---|---|
| `OrderBookSummary` | `market`, `asset_id`, `timestamp`, `bids`, `asks`, `min_order_size`, `neg_risk`, `tick_size`, `last_trade_price`, `hash` | Order book response |
| `MarketDetails` | `condition_id`, `tokens`, `min_tick_size`, `neg_risk`, `fee_details`, `maker_base_fee`, `taker_base_fee`, `rewards`, `accepting_orders`, `min_order_size`, `seconds_delay`, `game_start_time`, `clear_book_on_start`, `accepting_orders_timestamp`, `rfq_enabled`, `taker_order_delay_enabled`, `blockaid_check_enabled` | Full market details from CLOB |
| `FeeDetails` | `fee_rate`, `exponent`, `taker_only` | Platform fee details |
| `FeeInfo` | `rate: float=0`, `exponent: float=0` | Cached fee info per token |
| `ClobToken` | `token_id`, `outcome` | YES/NO token in a CLOB market |
| `ClobRewards` | `min_size`, `max_spread`, `enabled`, `skip_min_order_age`, `min_order_age_seconds` | Rewards config |

## Post Order Types

| Type | Fields | Description |
|---|---|---|
| `PostOrdersV1Args` | `order`, `orderType=GTC`, `deferExec=False` | Wrapper for posting a V1 signed order |
| `PostOrdersV2Args` | `order`, `orderType=GTC`, `deferExec=False` | Wrapper for posting a V2 signed order |
| `PostOrdersArgs` | `Union[PostOrdersV1Args, PostOrdersV2Args]` | Union type for `post_orders()` |

## Response/Status Types

| Type | Fields | Description |
|---|---|---|
| `BanStatus` | `closed_only: bool=False` | Closed-only mode status |
| `OrderScoring` | `scoring: bool=False` | Order scoring check result |
| `OrderSummary` | `price`, `size` | Individual order in book |
