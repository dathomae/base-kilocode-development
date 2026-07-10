# Polymarket Resources - Table of Contents

Reference documentation for the Polymarket prediction market platform, its concepts, APIs, and architecture.

## Concepts

**[concepts/markets-and-events.md](concepts/markets-and-events.md)**
: Markets, events, single vs multi-market events, condition IDs, token IDs, slugs, and identifying markets.

**[concepts/tokens-and-positions.md](concepts/tokens-and-positions.md)**
: Outcome tokens (ERC1155), split/merge/redeem operations, pUSD collateral, position value, holding rewards, and profit/loss.

**[concepts/prices-and-orderbook.md](concepts/prices-and-orderbook.md)**
: Prices as probabilities, CLOB order book, market/limit orders, spread, price discovery, hybrid-decentralized model.

**[concepts/order-lifecycle.md](concepts/order-lifecycle.md)**
: Order types (GTC/GTD/FOK/FAK), EIP-712 signing, order statuses, trade statuses, maker vs taker, cancellation, settlement.

**[concepts/resolution-and-redeeming.md](concepts/resolution-and-redeeming.md)**
: UMA Optimistic Oracle, proposal process, challenge period, dispute flows, DVM vote, redeeming tokens, contract addresses.

**[concepts/authentication.md](concepts/authentication.md)**
: L1/L2 authentication model, EIP-712 signatures, HMAC-SHA256, API credentials, signature types (EOA/POLY_PROXY/GNOSIS_SAFE/POLY_1271).

**[concepts/fees.md](concepts/fees.md)**
: Taker/maker fees, fee formula (curve-based), fee-free markets (geopolitics), maker rebates, builder fee attribution.

**[concepts/apis-overview.md](concepts/apis-overview.md)**
: Three API surfaces (Gamma, Data, CLOB), Bridge API, endpoints summary, public vs authenticated, geographic restrictions.

## Related API Documentation

See **[resources/apis/polymarket/](../apis/polymarket/toc.md)** for detailed documentation on the Python CLOB client (`py-clob-client-v2`).

See **[resources/apis/polymarket-unified-sdk/](../apis/polymarket-unified-sdk/toc.md)** for the recommended unified Python SDK (`polymarket-client`).
