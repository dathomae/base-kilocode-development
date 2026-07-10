# Prices & Order Book

How prices work and how the Central Limit Order Book enables peer-to-peer trading.

## Overview

Polymarket uses a **Central Limit Order Book (CLOB)** for trading in a hybrid-decentralized model:
1. **Offchain matching** — An operator matches compatible orders with low latency
2. **Onchain settlement** — Matched trades settle atomically via smart contracts on Polygon

This gives the speed of centralized matching with the security of onchain settlement. Users always maintain custody of their funds.

## Prices as Probabilities

Every share is priced between $0.00 and $1.00. The price directly represents the market's belief in the probability of that outcome:

| Price | Implied Probability |
|---|---|
| $0.25 | 25% chance |
| $0.50 | 50% chance |
| $0.75 | 75% chance |

The displayed price is the **midpoint** of the bid-ask spread. If the spread is wider than $0.10, the last traded price is shown instead.

### Midpoint Example

If best bid is $0.34 and best ask is $0.40:
```
Displayed price = ($0.34 + $0.40) / 2 = $0.37 (37% probability)
```
The actual trade price depends on which side you take: pay $0.40 when buying, receive $0.34 when selling.

## The Order Book

The order book has two sides:

| Side | Description |
|---|---|
| **Bids** | Buy orders — highest prices traders are willing to pay |
| **Asks** | Sell orders — lowest prices traders are willing to accept |

The **spread** is the gap between the highest bid and lowest ask. Tighter spreads indicate more liquid markets.

## Order Types

### Market Orders

Execute immediately at the best available price. For instant execution where you accept the spread:
- **Buying**: Pay the lowest ask price
- **Selling**: Receive the highest bid price

### Limit Orders

Execute only at your specified price or better. For price control with willingness to wait:
- Orders sit in the book until matched
- Can **partially fill** as different traders match portions
- Unfilled orders can be cancelled at any time

**Note**: All orders on Polymarket are technically limit orders. A "market order" is a limit order priced to execute immediately against resting orders.

## Price Discovery

When a new market launches, there's no initial price. The first price emerges when:
1. Someone places a limit buy at a price (e.g., $0.60)
2. Someone places a limit buy on the opposite side at the complementary price (e.g., $0.40)
3. Since $0.60 + $0.40 = $1.00, the orders match, creating 1 Yes + 1 No token

No trading size limits exist — the CLOB matches any amount, but large orders may move the price significantly. Always check order book depth before trading in size.
