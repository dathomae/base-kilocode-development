# SecureClient — Authenticated Trading & Wallet Operations

The `SecureClient` extends `PublicClient` with trading, wallet operations, and onchain transaction capabilities.

**Must be created with `SecureClient.create()`** — never construct directly.

## Constructor

```python
from polymarket import SecureClient

client = SecureClient.create(
    private_key=os.environ["POLYMARKET_PRIVATE_KEY"],
    wallet=None,                 # Defaults to deposit wallet
    environment=PRODUCTION,
    credentials=None,            # Auto-derived if omitted
    api_key=None,                # For gasless/relayer workflows
    nonce=0,                     # Credential derivation nonce
)
```

### What `create()` does:

1. Derives/validates API credentials via L1→L2 authentication
2. Detects wallet type (EOA, Deposit Wallet, Proxy, Safe)
3. Initializes transports for all API surfaces (Gamma, Data, CLOB, RFQ, Relayer, RPC)
4. Validates the wallet is ready for trading

### Wallet Properties

```python
client.wallet       # -> EvmAddress (funding wallet)
client.signer       # -> EvmAddress (signing key)
client.wallet_type  # -> WalletType (EOA, DEPOSIT, PROXY, SAFE)
client.credentials  # -> ApiKeyCreds
client.environment  # -> Environment
```

## Trading

### Place a Limit Order

```python
from polymarket import OrderSide, OrderType

order = client.place_limit_order(
    token_id="1234...",
    side=OrderSide.BUY,
    price="0.50",
    size="10",
    order_type=OrderType.GTC,
    post_only=False,
)
# Returns OrderResponse with order_id, status, making_amount, taking_amount
```

The method:
1. Validates tick size, neg risk, and price
2. Fetches the order book midpoint for validation
3. Signs the order (EIP-712)
4. Posts with automatic allowance recovery

### Place a Market Order

```python
from polymarket import MarketOrderType

# Buy $100 worth
order = client.place_market_order(
    token_id="1234...",
    side=OrderSide.BUY,
    amount="100",         # USD amount
    worst_price="0.55",   # Slippage protection
    order_type=MarketOrderType.FOK,
)

# Sell 200 shares
order = client.place_market_order(
    token_id="1234...",
    side=OrderSide.SELL,
    shares="200",          # Number of shares
    worst_price="0.45",    # Slippage protection
    order_type=MarketOrderType.FOK,
)
```

### Cancel Orders

```python
# Cancel single
client.cancel_order(order_id="0x...")

# Cancel all open orders
client.cancel_all_orders()

# Cancel market-specific orders
client.cancel_market_orders(market="0x...")
```

### List Open Orders

```python
paginator = client.list_open_orders(
    market="0x...",           # optional filter
    asset_id="1234...",       # optional filter
    page_size=20,
)
```

### CLOB Trades (Account Trades)

```python
paginator = client.list_clob_trades(
    market="0x...",
    page_size=20,
)
```

## Balance & Allowance

```python
balance = client.get_balance()
# balance.usdc_balance       # -> Decimal
# balance.usdc_allowance     # -> Decimal
# balance.token_balances     # -> dict (by asset_type)
# balance.token_allowances   # -> dict (by asset_type)
```

### Check Specific Allowance

```python
from polymarket import AssetType
allowance = client.get_balance(
    asset_type=AssetType.COLLATERAL,   # or CONDITIONAL
    token_id="1234...",
)
```

## Onchain Transactions

### Execute Custom Transaction Calls

```python
from polymarket.calls import (
    merge_v2_call, redeem_v2_call, split_v2_call,
    erc20_approval_call, erc20_transfer_call,
    erc1155_set_approval_for_all_call,
)
from polymarket.types import EvmAddress

# Execute a single call
handle = client.execute_transaction(
    calls=[merge_v2_call(router="0x...", condition_id="0x...", amount=1_000_000)],
    metadata="Merge position",
)
outcome = handle.wait()  # -> TransactionOutcome
```

### EOA vs Gasless

- **EOA wallet** (type 0): Transaction is broadcast via RPC; wallet pays gas. Returns `EoaTransactionHandle`.
- **Deposit wallet** (type 3): Transaction is relayed gaslessly. Returns `GaslessTransactionHandle`.
- Return type is a union: `SyncTransactionHandle`.

### Merge Positions

```python
# Merge by condition ID
handle = client.merge_positions(
    positions=[
        {"condition_id": "0x..."},
        {"condition_id": "0x...", "amount": "max"},
        {"market_id": "123", "amount": 500_000},
    ],
)
outcome = handle.wait()

# Merge combo positions by position ID
handle = client.merge_combo_positions(
    positions=[
        {"position_id": "0x..."},
        {"position_id": "0x...", "amount": "max"},
    ],
)
outcome = handle.wait()
```

### Redeem Positions

```python
# Redeem by condition ID
handle = client.redeem_positions(
    conditions=["0x...", "0x..."],
)
outcome = handle.wait()
```

### Split Positions

```python
handle = client.split_position(
    condition_id="0x...",
    amount=1_000_000,
)
outcome = handle.wait()
```

## Approvals

The SDK automatically handles approvals when placing orders. You can also manage them directly:

```python
# Set USDC approval for the exchange
client.approve_usdc()

# Set ERC1155 approval
client.approve_tokens()
```

## API Key Management

```python
# List keys
keys = client.fetch_api_keys()  # -> tuple[str, ...]

# Delete current key
client.delete_api_key()
```

## Builder API Keys

```python
from polymarket import BuilderApiKey

key: BuilderApiKey = client.create_builder_api_key()
# key.key, key.secret, key.passphrase

keys = client.get_builder_api_keys()  # -> tuple[BuilderApiKeyInfo, ...]
client.revoke_builder_api_key()
```

## Account & Profile

```python
# Public profile
profile = client.get_public_profile(address="0x...")  # -> PublicProfile | None
```

## Notifications

```python
notifications = client.get_notifications()  # -> tuple[Notification, ...]
client.drop_notifications(ids=["..."])
```

## Earnings & Rewards

```python
from polymarket import (
    TotalUserEarning, UserEarning, UserRewardsEarning, RewardsPercentages
)

# Total earnings for a day
total: TotalUserEarning = client.get_total_earnings(date="2024-01-15")

# Per-market earnings breakdown (paginated)
paginator = client.list_earnings(date="2024-01-15")

# Rewards earnings and configs
paginator = client.list_rewards_earnings(date="2024-01-15")

# Reward percentages
pcts: RewardsPercentages = client.get_reward_percentages()
```

## EOA-Specific

For EOA wallets (type 0), the client provides direct RPC access:

```python
client.rpc_send_transaction(...)
client.rpc_get_balance(...)
```
