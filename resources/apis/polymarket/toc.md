# Polymarket Python API Reference (py-clob-client-v2)

Documentation for the official Polymarket Python CLOB client library.

**Package**: `py-clob-client-v2`
**Repository**: [github.com/Polymarket/py-clob-client-v2](https://github.com/Polymarket/py-clob-client-v2)
**Install**: `pip install py-clob-client-v2`

**Note**: A new unified SDK (`polymarket-client`) has been released. See **[resources/apis/polymarket-unified-sdk/](../polymarket-unified-sdk/toc.md)** for the recommended client for new projects.

**Note**: A new unified SDK (`py-sdk`) has been released that combines all REST APIs and WebSockets. For new projects, see [Polymarket/py-sdk](https://github.com/Polymarket/py-sdk).

## API Documentation

**[overview-and-setup.md](overview-and-setup.md)**
: Installation, initialization, client modes (L0/L1/L2), quickstart flows, and the structure of the package.

**[client-module.md](client-module.md)**
: Complete reference for the `ClobClient` class — all methods organized by domain (public data, L1 auth, L2 trading, API key management, builder, earnings).

**[clob-types.md](clob-types.md)**
: All dataclass types used by the client — order args, market order args, book params, trade/order params, options, response types, contract config.

**[order-builder.md](order-builder.md)**
: The `OrderBuilder` class — how orders are constructed, signed, rounded, and converted for V1/V2 exchange contracts.

**[signing-and-headers.md](signing-and-headers.md)**
: The `Signer` class, L1/L2 header generation, EIP-712 signing for auth, HMAC-SHA256 for request authentication.

**[exceptions-and-utilities.md](exceptions-and-utilities.md)**
: Exception hierarchy (`PolyException`, `PolyApiException`), fee adjustment utilities, contract configuration by chain ID.

**[rfq-client.md](rfq-client.md)**
: `RfqClient` — Request for Quote protocol for OTC-style trading, creating requests/quotes, accepting, approving, and cancelling.

## Related Concepts

See **[resources/polymarket/](../../polymarket/toc.md)** for conceptual documentation on markets, tokens, orders, resolution, authentication, fees, and APIs.
