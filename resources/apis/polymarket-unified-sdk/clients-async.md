# Async Clients — AsyncPublicClient & AsyncSecureClient

The unified SDK provides full async variants of both client classes. APIs are identical to their sync counterparts but use `async`/`await`.

## AsyncPublicClient

```python
import asyncio
from polymarket import AsyncPublicClient, Market

async def main():
    async with AsyncPublicClient() as client:
        market: Market = await client.get_market(
            url="https://polymarket.com/event/example"
        )

        async for page in client.list_events(active=True, closed=False):
            for event in page.items:
                print(event.title)

asyncio.run(main())
```

## AsyncSecureClient

```python
from polymarket import AsyncSecureClient, OrderSide

async def main():
    client = await AsyncSecureClient.create(
        private_key=os.environ["POLYMARKET_PRIVATE_KEY"],
    )
    async with client:
        order = await client.place_limit_order(
            token_id="1234...",
            side=OrderSide.BUY,
            price="0.50",
            size="10",
        )
        print(order.order_id)

asyncio.run(main())
```

## Async Pagination

Async paginators support `async for` over both pages and items:

```python
# Iterate over pages
async for page in client.list_markets(closed=False):
    for market in page.items:
        print(market.question)

# Stream all items
async for market in client.list_markets(closed=False).iter_items():
    print(market.question)

# Get first page
page = await client.list_events(active=True).first_page()
```

## Async Transport Differences

- `AsyncPublicClient` uses `httpx.AsyncClient` under the hood
- `AsyncSecureClient` uses async relayer, async RPC, async secure CLOB
- All HTTP calls are non-blocking
- Session lifecycle is managed by the context manager or explicit `await client.close()`

## Context Manager

```python
# Preferred
async with AsyncPublicClient() as client:
    ...

# Equivalent explicit
client = AsyncPublicClient()
try:
    ...
finally:
    await client.close()
```
