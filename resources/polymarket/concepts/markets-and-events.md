# Markets & Events

The fundamental building blocks of Polymarket.

## Overview

Every prediction on Polymarket is structured around two core concepts: **events** and **markets**.

- **Market**: The fundamental tradable unit. Each market represents a single binary question with Yes/No outcomes. Markets map directly to two ERC1155 token IDs (Yes and No) on the CLOB.
- **Event**: A container that groups one or more related markets together. Events provide organizational structure and enable multi-outcome predictions.

## Markets

A market is identified by:

| Identifier | Description |
|---|---|
| **Condition ID** | Unique identifier for the market's condition in the Conditional Token Framework (CTF) contracts |
| **Question ID** | Hash of the market question, used for resolution |
| **Token IDs** | ERC1155 token IDs — one for Yes, one for No |
| **Slug** | URL-friendly identifier (e.g., `fed-decision-in-october`) |

**Important**: Markets can only be traded via the CLOB if `enableOrderBook` is `true`. Some markets may exist on-chain but not be available for order book trading.

### Example

A market like _"Will Bitcoin reach $150,000 by December 2026?"_ creates two outcome tokens:
- **Yes token** — Redeemable for $1 if Bitcoin reaches $150k
- **No token** — Redeemable for $1 if Bitcoin doesn't reach $150k

## Events

### Single-Market Events

When an event contains just one market, the event and market are essentially equivalent:

```
Event: Will Bitcoin reach $100,000 by December 2024?
└── Market: Will Bitcoin reach $100,000 by December 2024? (Yes/No)
```

### Multi-Market Events

When an event contains two or more markets, it enables mutually exclusive multi-outcome predictions:

```
Event: Who will win the 2024 Presidential Election?
├── Market: Donald Trump? (Yes/No)
├── Market: Joe Biden? (Yes/No)
├── Market: Kamala Harris? (Yes/No)
└── Market: Other? (Yes/No)
```

Multi-market events use **negative risk (neg risk)** to enable exclusive outcomes (see [Resolution & Redeeming](resolution-and-redeeming.md)).

## Identifying Markets

Every market and event has a unique **slug** in the URL:

```
https://polymarket.com/event/fed-decision-in-october
                              └── slug: fed-decision-in-october
```

Slugs can be used to fetch specific markets or events:

```bash
curl "https://gamma-api.polymarket.com/events?slug=fed-decision-in-october"
```

## Sports Markets

For sports markets, outstanding limit orders are **automatically cancelled** once the game begins, clearing the order book at the official start time. Game start times can shift — if a game starts earlier than scheduled, orders may not be cleared in time.

## Data Model (API Perspective)

The Gamma API returns data structured as:

1. **Event** — Top-level object containing metadata and nested markets
2. **Market** — Binary outcome within an event; contains `clobTokenIds` (array of [Yes, No]), condition ID, question ID

Each market has `outcomes` and `outcomePrices` arrays that map 1:1:
```json
{
  "outcomes": "[\"Yes\", \"No\"]",
  "outcomePrices": "[\"0.20\", \"0.80\"]"
}
```
- Index 0: "Yes" → 0.20 (20% probability)
- Index 1: "No" → 0.80 (80% probability)
