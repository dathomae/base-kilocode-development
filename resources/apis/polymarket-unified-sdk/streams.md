# WebSocket Streams

The unified SDK includes a `streams` module for real-time WebSocket data feeds.

## Stream Types

| Stream | Description |
|---|---|
| Market trades | Real-time trade execution feed |
| Order book updates | Order book snapshots and diffs |
| Live activity | Market-tagged live activity |
| Event streams | Event-level updates |

## Usage Patterns

Streams use the async client by default since WebSockets are inherently async.

```python
import asyncio
from polymarket import AsyncPublicClient
from polymarket.streams import trade_stream, order_book_stream

async def main():
    async with AsyncPublicClient() as client:
        # Stream trades for a market
        async for trade in trade_stream(client, asset_id="..."):
            print(trade.price, trade.size)

        # Stream order book updates
        async for update in order_book_stream(client, asset_id="..."):
            print(update.bids, update.asks)

asyncio.run(main())
```

## Available Stream Functions

```python
from polymarket.streams import (
    trade_stream,          # Individual trades
    order_book_stream,     # Order book snapshots/deltas
    live_activity_stream,  # Real-time activity
    event_stream,          # Event status changes
)
```

**Note**: The streams module is under active development. Check the [py-sdk repository](https://github.com/Polymarket/py-sdk) for the latest stream implementations and examples.
