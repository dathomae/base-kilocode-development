# APIs Overview

The three API surfaces that power Polymarket.

## API Endpoints

| API | Base URL | Purpose |
|---|---|---|
| **Gamma API** | `https://gamma-api.polymarket.com` | Markets, events, tags, series, comments, sports, search, public profiles — primary API for discovering and browsing market data |
| **Data API** | `https://data-api.polymarket.com` | User positions, trades, activity, holder data, open interest, leaderboards, builder analytics |
| **CLOB API** | `https://clob.polymarket.com` | Orderbook data, pricing, midpoints, spreads, price history. Also handles order placement, cancellation, and trading operations (authenticated) |

A separate **Bridge API** (`https://bridge.polymarket.com`) handles deposits and withdrawals (proxied from fun.xyz).

## Public vs Authenticated

- **Gamma API and Data API**: Fully public, no authentication required
- **CLOB API**: Public endpoints for orderbook/prices + authenticated endpoints for order management

## Endpoint Summary

### Gamma API — Events, Markets & Discovery

| Endpoint | Description |
|---|---|
| `GET /events` | List events with filtering and pagination |
| `GET /events/{id}` | Get a single event by ID |
| `GET /markets` | List markets with filtering and pagination |
| `GET /markets/{id}` | Get a single market by ID |
| `GET /public-search` | Search across events, markets, and profiles |
| `GET /tags` | Ranked tags/categories |
| `GET /series` | Series (grouped events) |
| `GET /sports` | Sports metadata |
| `GET /teams` | Teams |

### CLOB API — Prices & Orderbooks

| Endpoint | Description |
|---|---|
| `GET /price` | Price for a single token |
| `GET /prices` | Prices for multiple tokens |
| `GET /book` | Order book for a token |
| `POST /books` | Order books for multiple tokens |
| `GET /prices-history` | Historical price data for a token |
| `GET /midpoint` | Midpoint price for a token |
| `GET /spread` | Spread for a token |

### Data API — Positions, Trades & Analytics

| Endpoint | Description |
|---|---|
| `GET /positions?user={address}` | Current positions for a user |
| `GET /closed-positions?user={address}` | Closed positions for a user |
| `GET /activity?user={address}` | Onchain activity for a user |
| `GET /value?user={address}` | Total position value |
| `GET /oi` | Open interest for a market |
| `GET /holders` | Top holders of a market |
| `GET /trades` | Trade history |

### CLOB Trading Endpoints (Authenticated)

| Endpoint | Description |
|---|---|
| `GET /orders` | Get open orders |
| `POST /order` | Place an order |
| `DELETE /order` | Cancel an order |
| `DELETE /cancel-all` | Cancel all open orders |
| `DELETE /cancel-market-orders` | Cancel all orders for a market |
| `GET /balance-allowance` | Check balance and allowance |
| `GET /auth/api-key` | Derive API credentials |
| `POST /auth/api-key` | Create API credentials |

## Server Infrastructure

The CLOB matching engine runs in `eu-west-2` (primary). The closest non-georestricted region is `eu-west-1`. Direct co-location is available after KYC/KYB.

## Geographic Restrictions

Trading operations may be geoblocked from certain regions. Public data endpoints (Gamma, Data) are available worldwide. See [Polymarket Status](https://status.polymarket.com) for service health.
