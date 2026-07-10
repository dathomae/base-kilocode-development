# OrderBuilder — Order Construction & Signing

The `OrderBuilder` class handles construction and signing of limit and market orders for V1 and V2 exchange contracts. It is accessible as `client.builder` on any `ClobClient` instance.

## Constructor

```python
OrderBuilder(
    signer: Signer,
    signature_type: Optional[SignatureTypeV2] = None,  # Defaults to EOA (0)
    funder: Optional[str] = None,  # Defaults to signer address
)
```

- `signer` — The `Signer` instance for signing orders
- `signature_type` — 0=EOA, 1=POLY_PROXY, 2=GNOSIS_SAFE, 3=POLY_1271
- `funder` — Address that holds funds; for deposit wallets (type 3), this differs from the signer

### V2 Order Signer Logic

For V2 orders with `POLY_1271` (type 3), the order `signer` field is set to `funder` (the deposit wallet). For all other types, it uses `signer.address()`.

## Rounding Configuration

Each tick size has a `RoundConfig` that specifies decimal precision:

```python
ROUNDING_CONFIG = {
    "0.1":    RoundConfig(price=1, size=2, amount=3),
    "0.01":   RoundConfig(price=2, size=2, amount=4),
    "0.005":  RoundConfig(price=3, size=2, amount=5),
    "0.0025": RoundConfig(price=4, size=2, amount=6),
    "0.001":  RoundConfig(price=3, size=2, amount=5),
    "0.0001": RoundConfig(price=4, size=2, amount=6),
}
```

- `price` — Decimal places for price
- `size` — Decimal places for size
- `amount` — Decimal places for maker/taker amounts

## Core Methods

### `get_order_amounts(side, size, price, round_config) -> tuple`

Computes `(Side, maker_amount, taker_amount)` for a limit order. The amounts are tokenized (scaled to token decimals).

- **BUY**: maker_amount = pUSD paid, taker_amount = tokens received
- **SELL**: maker_amount = tokens given, taker_amount = pUSD received

Rounding uses `round_up`/`round_down` with a safety buffer: if rounding up creates too many decimal places, it rounds back down.

### `get_market_order_amounts(side, amount, price, round_config) -> tuple`

Same as above but for market orders. Key difference: market orders use `round_down` for price (V2 change) instead of `round_normal`.

### `build_order(order_args, options, version=2, fee_rate_bps=None) -> SignedOrder`

Creates and signs a **limit order**.

**V1 orders** (`version=1`):
- Includes: `taker`, `feeRateBps`, `nonce` fields
- Uses V1 exchange or neg_risk_exchange contract
- `POLY_1271` signature type NOT supported

**V2 orders** (`version=2`, default):
- Includes: `timestamp`, `metadata`, `builder` fields
- Uses V2 exchange or neg_risk_exchange_v2 contract
- Uses `maker` as `funder`, `signer` as configured by signature type

### `build_market_order(order_args, options, version=2, fee_rate_bps=None) -> SignedOrder`

Same structure as `build_order` but for market orders. Market orders have expiration=0 and are typically FOK/FAK.

### `calculate_buy_market_price(positions, amount_to_match, order_type) -> float`

Calculate the expected fill price for a buy market order. Walks through asks in reverse order. For FOK, raises if order can't be fully filled.

### `calculate_sell_market_price(positions, amount_to_match, order_type) -> float`

Calculate the expected fill price for a sell market order. Walks through bids in reverse order. For FOK, raises if order can't be fully filled.

## Building Process

1. **Resolve amounts**: `get_order_amounts()` or `get_market_order_amounts()` converts side/price/size to onchain token amounts
2. **Select exchange**: Choose V1 or V2 exchange address based on neg_risk and version
3. **Build order data**: Create `OrderDataV1` or `OrderDataV2` object with all fields
4. **Sign**: `ExchangeOrderBuilderV1.build_signed_order()` or `ExchangeOrderBuilderV2.build_signed_order()` — signs the order using EIP-712 typed data signing

## Usage in ClobClient

The client wraps `OrderBuilder` with additional logic:

- **`create_order()`**: Resolves tick size, validates price, optionally adjusts buy amount for fees, builds and signs
- **`create_market_order()`**: Same but also auto-calculates price from the order book if not provided
- **`create_and_post_order()`**: Calls `create_order()` then `post_order()`, retries on version mismatch
- **`create_and_post_market_order()`**: Same but for market orders

## Version Auto-detection

The client fetches `get_version()` on first use and caches it. If `post_order()` returns a version mismatch error, it re-fetches the version and the `_retry_on_version_update` wrapper retries.
