# Exceptions & Utilities

## Exception Hierarchy

### `PolyException(msg)`

Base exception for the client. Stores a message string.

### `PolyApiException`

Subclass for API response errors. Created from an HTTP response or error message:

```python
class PolyApiException(PolyException):
    def __init__(self, resp: Optional[httpx.Response] = None, error_msg=None):
        # If resp provided: extracts status_code and error_msg from response
        # If error_msg provided: status_code = None, error_msg = error_msg
```

Properties:
- `status_code` — HTTP status code (or None)
- `error_msg` — Response body parsed as JSON (or raw text if not valid JSON)

String representation: `PolyApiException[status_code=X, error_message=Y]`

## Fee Utilities

From `py_clob_client_v2.fees`:

### `validate_fee_slippage(fee_slippage: float)`

Validates the `fee_slippage` constructor parameter. Rules:
- Must be 0 or between 1 and 100 (inclusive)
- Must be a finite number
- Raises `ValueError` if invalid

### `adjust_buy_amount_for_fees(amount, price, user_usdc_balance, fee_rate, fee_exponent, builder_taker_fee_rate, fee_slippage=0) -> float`

Adjusts a buy order amount to account for fees. Used internally when `user_usdc_balance` is provided on order args.

The formula:
```
platform_fee_rate = fee_rate * (price * (1 - price)) ** fee_exponent
effective_platform_fee_rate = platform_fee_rate * (1 + fee_slippage / 100)

fee_base_amount = min(amount, user_usdc_balance)
platform_fee = (fee_base_amount / price) * effective_platform_fee_rate
builder_fee = fee_base_amount * builder_taker_fee_rate
total_cost = amount + platform_fee + builder_fee

if user_usdc_balance <= total_cost:
    return max(user_usdc_balance - platform_fee - builder_fee, 0)
return amount
```

This is called automatically by `ClobClient.create_order()` and `create_market_order()` for BUY orders when `user_usdc_balance` is set on V2 order args. It also accounts for builder taker fees when a builder code is configured.

## Contract Configuration

From `py_clob_client_v2.config`:

### `get_contract_config(chain_id: int) -> ContractConfig`

Returns contract addresses for the given chain. Supports:
- **137** — Polygon Mainnet
- **80002** — Amoy Testnet

Raises if chain_id is unrecognized.

Example:
```python
config = get_contract_config(137)
# config.exchange -> "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E"
# config.exchange_v2 -> "0xE111180000d2663C0091e4f400237545B87B996B"
# config.collateral -> "0xC011a7E12a19f7B1f670d46F03B03f3342E82DFB"
```

### Decimals

```python
COLLATERAL_TOKEN_DECIMALS = 6   # Token amount scaling (USDC)
CONDITIONAL_TOKEN_DECIMALS = 6  # Token amount scaling (outcome tokens)
```

## Utilities

From `py_clob_client_v2.utilities`:

| Function | Description |
|---|---|
| `price_valid(price, tick_size) -> bool` | Checks if price is within valid range: `tick_size <= price <= 1 - tick_size` |
| `is_tick_size_smaller(tick_size1, tick_size2) -> bool` | Returns True if tick_size1 is a smaller (more precise) increment than tick_size2 |
| `generate_orderbook_summary_hash(orderbook) -> str` | Generates a hash from an OrderBookSummary for verification |
| `parse_raw_orderbook_summary(book) -> OrderBookSummary` | Parses a raw dict into an OrderBookSummary dataclass |
