# Polymarket Perpetuals (Perps)

The unified SDK includes support for Polymarket's perpetual futures markets, accessed through `perps.py`.

## Auth

```python
from polymarket.auth import BuilderApiKey, RelayApiKey
```

## Key Types

```python
from polymarket import (
    PerpsInstrument,     # Perpetual instrument definition
    PerpsBook,           # Order book for a perp market
    PerpsOrder,          # Perp order details
    PerpsOrderRequest,   # New order request
    PerpsOrderPlacement,  # Order placement result
    PerpsPosition,        # User's position in a perp market
    PerpsPortfolio,       # Aggregated portfolio
    PerpsBalance,         # Balance snapshot
    PerpsTrade,           # Trade execution
    PerpsFill,            # Order fill
    PerpsTicker,          # Market ticker (price, 24h vol, etc.)
    PerpsDeposit,         # Deposit record
    PerpsWithdrawal,      # Withdrawal record
    PerpsCandle,          # OHLCV candle data
    PerpsFundingRate,     # Funding rate
    PerpsFundingPayment,  # Funding payment record
    PerpsFeeScheduleEntry,  # Fee schedule
    PerpsCredentials,     # Perps API credentials
    PerpsTpSlTrigger,     # Take-profit / stop-loss trigger
    PerpsPlacedTpSlOrder,  # Placed TP/SL order
    PerpsPlacedTpSlOrders,  # Multiple TP/SL orders
)
```

## Usage

Perps functionality is available on `SecureClient` (requires authentication):

```python
from polymarket import SecureClient

client = SecureClient.create(private_key=os.environ["PRIVATE_KEY"])

# Get instrument info
# instrument = client.get_perps_instrument("BTC-USD")

# Get order book
# book = client.get_perps_order_book("BTC-USD")

# Place order
# order = client.place_perps_order(...)
```

**Note**: Perpetuals support is under active development. Check the [py-sdk repository](https://github.com/Polymarket/py-sdk) for the latest APIs and examples.
