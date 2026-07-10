# Polymarket Python API — Overview & Setup

## Package Structure

```
py_clob_client_v2/
├── __init__.py          # Public API exports (ClobClient, types, enums)
├── client.py            # ClobClient — the main client class
├── clob_types.py        # All dataclass type definitions
├── config.py            # Contract configuration by chain ID
├── constants.py         # Built-in constants (addresses, error messages, etc.)
├── endpoints.py         # API endpoint URL constants
├── exceptions.py        # Exception hierarchy
├── fees.py              # Fee adjustment utilities
├── signer.py            # Signer — ECDSA signing with eth-account
├── utilities.py         # Price/tick validation, orderbook hash generation
├── headers/             # L1 and L2 header generation
├── http_helpers/        # HTTP request helpers (get/post/delete + retry)
├── order_builder/       # Order construction, signing, rounding
│   ├── builder.py       # OrderBuilder class
│   ├── constants.py     # BUY/SELL constants
│   └── helpers.py       # Rounding helpers, decimal conversion
├── order_utils/         # V1/V2 order data models & signing (EIP-712)
│   ├── model/
│   │   ├── order_data_v1.py
│   │   ├── order_data_v2.py
│   │   └── side.py      # Side enum (BUY=0, SELL=1)
│   └── ...              # ExchangeOrderBuilder classes
├── rfq/                 # Request for Quote (RFQ) subsystem
│   ├── __init__.py
│   ├── rfq_client.py    # RfqClient class
│   ├── rfq_types.py     # RFQ-specific dataclasses
│   └── rfq_helpers.py   # RFQ parameter parsing
└── signing/             # EIP-712 and HMAC signing implementations
    ├── eip712.py
    ├── hmac.py
    └── ...
```

## Installation

```bash
pip install py-clob-client-v2
```

## Client Modes

The `ClobClient` operates in three modes depending on what credentials are provided:

| Mode | Has Signer | Has Creds | Capabilities |
|---|---|---|---|
| **L0** | No | No | Public data only (order books, prices, markets) |
| **L1** | Yes | No | Public data + derive API keys + sign orders locally |
| **L2** | Yes | Yes | Full trading: place/cancel orders, check balances, manage keys |

## Basic Usage Pattern

### Step 1: Create L1 client, derive API credentials

```python
from py_clob_client_v2 import ClobClient
import os

host = "https://clob.polymarket.com"
chain_id = 137  # Polygon mainnet (80002 for Amoy testnet)

# L1 client — can create/derive API keys + sign orders
client = ClobClient(host=host, chain_id=chain_id, key=os.environ["PRIVATE_KEY"])
creds = client.create_or_derive_api_key()

# creds is an ApiCreds(api_key="...", api_secret="...", api_passphrase="...")
```

### Step 2: Create fully authenticated L2 client

```python
from py_clob_client_v2 import ClobClient, SignatureTypeV2

client = ClobClient(
    host=host,
    chain_id=chain_id,
    key=os.environ["PRIVATE_KEY"],
    creds=creds,
    signature_type=SignatureTypeV2.POLY_1271,  # For new API users with deposit wallets
    funder=os.environ["DEPOSIT_WALLET_ADDRESS"],
)
```

### Step 3: Fetch market info and place an order

```python
from py_clob_client_v2 import OrderArgs, OrderType, PartialCreateOrderOptions
from py_clob_client_v2.order_builder.constants import BUY

# Fetch token IDs from Gamma API
import requests
markets = requests.get(
    "https://gamma-api.polymarket.com/markets",
    params={"active": "true", "closed": "false", "limit": 1}
).json()

token_id = markets[0]["clobTokenIds"][0]  # Yes token
condition_id = markets[0]["conditionId"]

# Get market info (tick size, neg risk, fees)
market_info = client.get_clob_market_info(condition_id)
tick_size = str(market_info["mts"])

# Place a limit buy order
response = client.create_and_post_order(
    OrderArgs(token_id=token_id, price=0.50, size=10, side=BUY),
    options=PartialCreateOrderOptions(tick_size=tick_size),
    order_type=OrderType.GTC,
)
print(response["orderID"], response["status"])
```

## Quick Patterns

### Market Orders

```python
from py_clob_client_v2 import MarketOrderArgs, OrderType

response = client.create_and_post_market_order(
    MarketOrderArgs(token_id=token_id, amount=100, side=BUY, order_type=OrderType.FOK),
    options=PartialCreateOrderOptions(tick_size=tick_size),
    order_type=OrderType.FOK,
)
```

### Checking Orders and Trades

```python
# Open orders
open_orders = client.get_open_orders()

# Trade history
trades = client.get_trades()

# Cancel an order
client.cancel_order(order_id=response["orderID"])
```

### Cache Behavior

The client caches market info automatically:
- `__tick_sizes`: Maps token_id → tick size
- `__neg_risk`: Maps token_id → neg risk flag
- `__fee_infos`: Maps token_id → FeeInfo (rate + exponent)
- `__token_condition_map`: Maps token_id → condition_id

These caches are populated by `get_clob_market_info()`, `get_tick_size()`, `get_neg_risk()`, etc.
