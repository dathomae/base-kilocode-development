# PublicClient — Public Data Access

The `PublicClient` provides read access to all public Polymarket data — no authentication required.

## Constructor

```python
from polymarket import PublicClient, PRODUCTION

client = PublicClient(
    environment=PRODUCTION,  # default
)
```

## Market & Event Access

### Get Market

Retrieve a market by ID, slug, or URL:

```python
market = client.get_market(id="540816")
market = client.get_market(slug="fed-decision-in-october")
market = client.get_market(url="https://polymarket.com/event/fed-decision-in-october")
```

Returns: `Market` object with `question`, `condition_id`, `clob_token_ids` (Yes, No), `outcomes`, `outcome_prices`, `volume`, `liquidity`, `end_date`, `neg_risk`, `enable_order_book`, etc.

```python
market = client.get_market_tags(id="540816")  # -> tuple[TagReference, ...]
```

### Get Event

```python
event = client.get_event(id="123")
event = client.get_event(slug="presidential-election-2024")
event = client.get_event(url="https://polymarket.com/event/...")
```

Returns: `Event` with `title`, `description`, `markets` (list of `Market`), `volume`, `liquidity`, `start_date`, `end_date`, `tags`.

```python
event = client.get_event_tags(id="123")  # -> tuple[TagReference, ...]
```

### Get Series

```python
series = client.get_series(id="456")
```

### Get Tags

```python
tag = client.get_tag(id="100381")
tag = client.get_tag(slug="crypto")
```

### Related Tags

```python
tags = client.get_related_tags(id="100381")      # -> tuple[RelatedTag, ...]
resources = client.get_related_tag_resources(id="100381")  # -> tuple[Tag, ...]
```

### Sports Metadata

```python
sports = client.get_sports()                  # -> tuple[SportsMetadata, ...]
market_types = client.get_sports_market_types()  # -> SportsMarketTypes
```

## Search

```python
import polymarket
for result in client.search(q="election", sort="volume").iter_items():
    print(result)
```

Parameters: `q` (required), `ascending`, `events_status`, `events_tag`, `sort` (volume, volume_24hr, liquidity, competitive, etc.), `search_profiles`, `search_tags`.

## Price & Order Book

### Midpoint

```python
mid = client.get_midpoint(token_id="1234...")   # -> Decimal
mids = client.get_midpoints(token_ids=["...", "..."])  # -> dict[str, Decimal]
```

### Price

```python
from polymarket import OrderSide
price = client.get_price(token_id="1234...", side=OrderSide.BUY)  # -> Decimal

from polymarket import PriceRequest
prices = client.get_prices(requests=[
    PriceRequest(token_id="...", side=OrderSide.BUY),
    PriceRequest(token_id="...", side=OrderSide.SELL),
])
```

### Order Book

```python
book = client.get_order_book(token_id="1234...")    # -> OrderBook
books = client.get_order_books(token_ids=["...", "..."])  # -> tuple[OrderBook, ...]
```

### Spread

```python
spread = client.get_spread(token_id="1234...")    # -> Decimal
spreads = client.get_spreads(token_ids=["...", "..."])  # -> dict[str, Decimal]
```

### Last Trade Price

```python
ltp = client.get_last_trade_price(token_id="1234...")        # -> LastTradePrice
ltps = client.get_last_trade_prices(token_ids=["...", "..."]) # -> tuple[LastTradePriceForToken, ...]
```

### Price History

```python
from polymarket import PriceHistoryInterval
history = client.get_price_history(
    token_id="1234...",
    start_ts=1700000000,
    end_ts=1700100000,
    fidelity=5,
)
# or by interval
history = client.get_price_history(
    token_id="1234...",
    interval=PriceHistoryInterval.ONE_DAY,
)
```

### Market Price Estimation

```python
# Buy: estimate price for spending a given USD amount
avg_price = client.estimate_market_price(
    token_id="1234...",
    side="BUY",
    amount=100,  # Spend $100
)

# Sell: estimate price for selling a given number of shares
avg_price = client.estimate_market_price(
    token_id="1234...",
    side="SELL",
    shares=200,  # Sell 200 shares
)
```

## List Methods (Paginated)

All list methods return `Paginator` objects. Use `.first_page()`, `for page in`, or `.iter_items()`.

### Events

```python
paginator = client.list_events(
    active=True,
    closed=False,
    tag_ids=[100381],
    order="volume_24hr",
    ascending=False,
    liquidity_min=10000,
    page_size=20,
)
```

### Markets

```python
paginator = client.list_markets(
    closed=False,
    tag_id=100381,
    condition_ids=["0x..."],
    slug="example-market",
    order="volume",
    page_size=20,
)
```

### Series

```python
paginator = client.list_series(closed=False, page_size=20)
```

### Tags

```python
paginator = client.list_tags(is_carousel=True, order="markets_count", page_size=20)
```

### Teams

```python
paginator = client.list_teams(league="NBA", page_size=20)
```

### Comments

```python
from polymarket._internal.actions.gamma import CommentParentEntityType

paginator = client.list_comments(
    parent_entity_id="123",
    parent_entity_type=CommentParentEntityType.MARKET,
    order="created_at",
    page_size=20,
)

# Comments by a user
paginator = client.list_comments_by_user_address(address="0x...", page_size=20)
```

### Positions

```python
paginator = client.list_positions(
    user="0x...",
    market=["0x..."],
    redeemable=True,
    mergeable=True,
    sort_by="size",
    sort_direction="desc",
    page_size=20,
)
```

### Closed Positions

```python
paginator = client.list_closed_positions(
    user="0x...",
    sort_by="close_time",
    sort_direction="desc",
    page_size=20,
)
```

### Combo Positions

```python
paginator = client.list_combo_positions(
    user="0x...",
    status="OPEN",
    page_size=20,
)
```

### Market Positions

```python
paginator = client.list_market_positions(
    market="0x...",
    user="0x...",
    sort_by="size",
    page_size=20,
)
```

### Trades

```python
paginator = client.list_trades(
    user="0x...",
    market=["0x..."],
    taker_only=True,
    start=1700000000,
    end=1700100000,
    page_size=20,
)
```

### Activity

```python
paginator = client.list_activity(
    user="0x...",
    activity_types=["TRADE", "SPLIT", "MERGE", "REDEEM"],
    start=1700000000,
    page_size=20,
)
```

### Combo Activity

```python
paginator = client.list_combo_activity(
    user="0x...",
    condition_id="0x...",
    page_size=50,
)
```

## Data API Methods

### Live Volumes

```python
volumes = client.get_event_live_volumes(id="123")  # -> tuple[LiveVolume, ...]
```

### Open Interest

```python
oi = client.get_open_interests(market=["0x...", "0x..."])  # -> tuple[OpenInterest, ...]
```

### Market Holders

```python
holders = client.get_market_holders(
    market=["0x..."],
    limit=100,
    min_balance=1000000,
)
```

### Portfolio Values

```python
values = client.get_portfolio_values(
    user="0x...",
    market=["0x...", "0x..."],
)
```

### Traded Market Count

```python
count = client.get_traded_market_count(user="0x...")  # -> TradedMarketCount
```

### Accounting Snapshot

```python
archive = client.download_accounting_snapshot(user="0x...")  # -> bytes
```

## Leaderboard Methods

### Builder Leaderboard

```python
from polymarket import LeaderboardTimePeriod

paginator = client.list_builder_leaderboard(
    time_period=LeaderboardTimePeriod.WEEK,
    page_size=20,
)
```

### Trader Leaderboard

```python
paginator = client.list_trader_leaderboard(
    category="PROFIT",
    time_period=LeaderboardTimePeriod.MONTH,
    order_by="profit",
    page_size=20,
)
```

## Builder Data

### Builder Volumes

```python
from polymarket import BuilderVolumeTimePeriod

volumes = client.get_builder_volumes(
    time_period=BuilderVolumeTimePeriod.DAY
)
```

### Builder Trades

```python
paginator = client.list_builder_trades(
    builder_code="0x...",
    market="0x...",
    after="1700000000",
)
```

## Rewards

```python
# Current rewards
paginator = client.list_current_rewards(sponsored=True)

# Market rewards
paginator = client.list_market_rewards(
    condition_id="0x...",
    sponsored=False,
)
```
