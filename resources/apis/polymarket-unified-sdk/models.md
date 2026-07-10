# Core Models — Gamma API Domain Types

The unified SDK uses Pydantic models for all response types, organized in `polymarket.models`.

## Importing Models

Most models are re-exported from `polymarket` directly:

```python
from polymarket import (
    Market, Event, Tag, Series, Comment, Team,
    Position, ClosedPosition, ComboPosition,
    Trade, Activity, ComboActivity,
    OrderBook, OrderBookLevel, LastTradePrice,
    MarketId, TokenId, ConditionId, EventId,
    OrderSide, OrderType, MarketOrderType,
    SignedOrder, OpenOrder, OrderResponse,
    PriceHistoryInterval, PriceHistoryPoint,
    SearchResults, PublicProfile, Profile,
    Notification, BalanceAllowance, AssetType,
    # ... many more
)
```

## Domain Type Markers

These are `NewType` wrappers around `str` — runtime is identical but type-checkers/IDEs know the difference:

| Type | Base | Description |
|---|---|---|
| `MarketId` | `str` | Market identifier |
| `TokenId` | `str` | Token identifier |
| `ConditionId` | `str` | CTF condition ID |
| `EventId` | `str` | Event identifier |
| `OrderId` | `str` | Order identifier |
| `PositionId` | `str` | Position identifier |
| `QuestionId` | `str` | Question hash |
| `CommentId` | `str` | Comment ID |
| `TagId` | `str` | Tag ID |
| `SeriesId` | `str` | Series ID |
| `ComboConditionId` | `str` | Combo condition ID |
| `ComboActivityId` | `str` | Combo activity ID |
| `ResolutionRequestId` | `str` | Resolution request ID |
| `CtfConditionId` | `str` | CTF-formatted condition ID |

### Primitive Wrappers

| Type | Base | Description |
|---|---|---|
| `EvmAddress` | `str` | Ethereum address |
| `HexString` | `str` | Hex-encoded data |
| `TransactionHash` | `str` | Transaction hash |

### Conversion Functions

```python
from polymarket import to_ctf_condition_id, to_combo_condition_id

ctf_id = to_ctf_condition_id("0x...")
combo_id = to_combo_condition_id("0x...")
```

## Gamma Models

### Market

Core market model from the Gamma API:

```python
market = client.get_market(id="540816")
market.id              # str
market.question        # str
market.condition_id    # ConditionId
market.clob_token_ids  # list[TokenId] (Yes, No)
market.outcomes        # list[str]
market.outcome_prices  # list[float]
market.volume          # float
market.liquidity       # float
market.start_date      # str | None (ISO format)
market.end_date        # str | None
market.neg_risk        # bool
market.enable_order_book  # bool
market.slug            # str
market.tags            # list[TagReference]
market.comments_count  # int
```

### Event

```python
event = client.get_event(id="123")
event.id             # str
event.title          # str
event.description    # str
event.markets        # list[Market]
event.tags           # list[TagReference]
event.volume         # float
event.liquidity      # float
event.start_date     # str | None
event.end_date       # str | None
```

### Tag

```python
tag = client.get_tag(id="100381")
tag.id             # int
tag.slug           # str
tag.label          # str
tag.description    # str | None
```

### Series

```python
series = client.get_series(id="456")
series.id        # str
series.title     # str
series.events    # list[Event]
```

### SearchResults

```python
results = client.search(q="election").first_page()
for item in results.items:
    pass  # Contains markets, events, and optional profiles/tags
```

### Comment

```python
comment = client.get_comment_thread(id="...")
comment.id        # CommentId
comment.body      # str
comment.created_at  # str
```

### PublicProfile

```python
profile = client.get_public_profile(address="0x...")
if profile:
    print(profile.username)
```

## Data API Models

### Position

```python
position.id          # str
position.market_id   # MarketId
position.condition_id  # ConditionId
position.outcome     # str ("Yes" / "No")
position.size        # float (token balance)
position.avg_price   # float
position.current_price  # float
position.pnl         # float
position.redeemable  # bool
position.mergeable   # bool
```

### Trade

```python
trade.id              # str
trade.market_id       # MarketId
trade.side            # str
trade.price           # float
trade.size            # float
trade.timestamp       # int
```

### Activity

Union type of trade, split, merge, redeem activities. Use `isinstance()` to discriminate:

```python
from polymarket import TradeActivity, SplitActivity, MergeActivity, RedeemActivity

for activity in client.list_activity(user="0x...").iter_items():
    if isinstance(activity, TradeActivity):
        print("Trade:", activity.price, activity.size)
    elif isinstance(activity, SplitActivity):
        print("Split:", activity.amount)
```

### ComboPosition

```python
combo.id          # str
combo.status      # ComboPositionStatus
combo.legs        # list[ComboPositionLeg]
combo.value       # float
```

### LeaderboardEntry

```python
entry = paginator.first_page().items[0]
entry.rank         # int
entry.address      # str
entry.volume       # float | None
entry.profit       # float | None
entry.markets_traded  # int | None
```

## CLOB Models

### OrderBook

```python
book = client.get_order_book(token_id="...")
book.bids  # list[OrderBookLevel] — (price, size)
book.asks  # list[OrderBookLevel]
book.min_order_size  # str
book.tick_size       # TickSize | None
book.neg_risk        # bool | None
```

### OrderBookLevel

```python
level.price  # str
level.size   # str
```

### PriceRequest

```python
from polymarket import PriceRequest
pr = PriceRequest(token_id="...", side=OrderSide.BUY)
```

### SignedOrder

The full EIP-712 signed order representation. Used internally by `place_limit_order` / `place_market_order`. Contains: `token_id`, `price`, `size`, `side`, `signature`, `salt`, `maker`, `signer`, `expiration`, `maker_amount`, `taker_amount`, `signature_type`, `order_type`.

### OrderResponse

```python
response = client.place_limit_order(...)
response.order_id      # OrderId
response.status        # OrderPostStatus
response.making_amount # str
response.taking_amount # str
response.success       # bool
response.error_msg     # str
```

### OpenOrder

```python
order.id              # OrderId
order.status          # str
order.side            # str
order.price           # str
order.original_size   # str
order.size_matched    # str
order.token_id        # TokenId
order.created_at      # str
```

## Enums (StrEnum for Outputs)

| Enum | Values |
|---|---|
| `OrderSide` | `BUY`, `SELL` |
| `OrderType` | `GTC`, `GTD`, `FOK`, `FAK` |
| `MarketOrderType` | `FOK`, `FAK` |
| `OrderPostStatus` | `live`, `matched`, `delayed`, `unmatched` |
| `AssetType` | `COLLATERAL`, `CONDITIONAL` |
| `TickSize` | `"0.1"`, `"0.01"`, `"0.001"`, `"0.0001"` |
| `PriceHistoryInterval` | `max`, `1w`, `1d`, `6h`, `1h` |
| `ComboPositionStatus` | `OPEN`, `CLOSED` |
| `ActivityType` | `TRADE`, `SPLIT`, `MERGE`, `REDEEM`, `CONVERSION`, `REWARD`, `YIELD`, etc. |
| `BuilderVolumeTimePeriod` | `DAY`, `WEEK`, `MONTH`, `ALL` |
| `LeaderboardTimePeriod` | `DAY`, `WEEK`, `MONTH`, `ALL` |
| `LeaderboardCategory` | `PROFIT`, `VOLUME`, `ROI`, `MARKETS` |
| `LeaderboardOrderBy` | `profit`, `volume`, `roi`, `markets_traded` |
| `UmaResolutionStatus` | `PENDING`, `RESOLVED`, `DISPUTED`, etc. |
| `SportsMarketTypes` | Various market type identifiers |
