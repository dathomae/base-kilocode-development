# Transaction System

The unified SDK provides a typed transaction system for executing onchain operations through either EOA (self-funded) or gasless (relay) flows.

## TransactionCall

A `TransactionCall` represents a single contract interaction:

```python
from polymarket.calls import TransactionCall

call = TransactionCall(
    to="0x...",         # Contract address
    data="0x...",       # Encoded function call
    value=0,            # ETH value (always 0 for Polymarket)
)
```

## Pre-built Calls

### Position Management

```python
from polymarket.calls import (
    merge_v2_call,         # Merge a single combo position set
    redeem_v2_call,        # Redeem winning positions via V2 exchange
    split_v2_call,         # Split collateral into Yes/No tokens via V2
    split_position_call,   # Split via CTF directly
    merge_positions_call,  # Merge positions via CTF
    ctf_redeem_positions_call,  # Redeem via CTF
    combinatorial_prepare_condition_call,  # Prepare condition for combinatorial
)

# Example: merge a combo position
call = merge_v2_call(
    router=EvmAddress("0xE111..."),
    condition_id="0x03...",
    amount=1_000_000,
)
```

### ERC20 / ERC1155 Operations

```python
from polymarket.calls import (
    erc20_approval_call,
    erc20_transfer_call,
    erc1155_set_approval_for_all_call,
    erc1155_balance_of_call,
    erc1155_balance_of_batch_call,
)
```

### Utilities

```python
from polymarket.calls import (
    decode_erc1155_balance_of_result,
    decode_erc1155_balance_of_batch_result,
)
```

## Executing Transactions

```python
handle = client.execute_transaction(
    calls=[merge_v2_call(...), merge_v2_call(...)],
    metadata="Batch merge 2 positions",
)
outcome = handle.wait()
```

### TransactionHandle Types

The return type depends on wallet type:

| Wallet Type | Handle Type | Gas Payment |
|---|---|---|
| EOA | `SyncEoaTransactionHandle` | POL from wallet |
| Deposit | `SyncGaslessTransactionHandle` | Relay (gasless) |

Union type: `SyncTransactionHandle = SyncEoaTransactionHandle | SyncGaslessTransactionHandle`

### handle.wait()

Blocks until the transaction reaches finality or fails:
- Returns `TransactionOutcome` with `success`, `transaction_hash`, `block_number`
- Raises `TransactionFailedError` on failure

## Deposit Wallet Transactions (Gasless)

For deposit wallet users, transactions are:
1. Encoded as relayer calls
2. Submitted to the Polymarket relayer
3. Executed gaslessly on behalf of the deposit wallet

The relayer needs a valid `ApiKey` (not to be confused with CLOB API credentials):

```python
from polymarket import ApiKey

client = SecureClient.create(
    private_key=os.environ["PRIVATE_KEY"],
    api_key=ApiKey(os.environ["POLYMARKET_API_KEY"]),
)
```

## EOA Transactions

For EOA wallets, transactions are:
1. Encoded as Ethereum transactions
2. Broadcast via RPC (Polygon)
3. The wallet pays gas in POL

## Position Context

The SDK can resolve position details from onchain for merge/redeem operations:

```python
# Merge positions by condition ID
handle = client.merge_positions(
    positions=[
        {"condition_id": "0x..."},
        {"condition_id": "0x...", "amount": "max"},  # Max available
        {"market_id": "123", "amount": 500_000},     # By market ID
    ],
)
```

For combo positions, use the `merge_combo_positions` variant:

```python
handle = client.merge_combo_positions(
    positions=[
        {"position_id": "0x..."},
        {"position_id": "0x...", "amount": "max"},
    ],
)
```

The `"max"` amount resolves the full available balance onchain before executing.

### Position Request Types

```python
from polymarket.transactions import (
    MergeMarketConditionRequest,
    MergeMarketIdRequest,
    MergeMarketPositionRequest,
    MergeComboPositionRequest,
    MergePositionRequest,
    # Union type
    MergePositionRequest,  # MergeMarketConditionRequest | MergeMarketIdRequest | MergeComboPositionRequest
)
```

## Example: Complete Merge Flow

```python
from polymarket import SecureClient
import os

client = SecureClient.create(
    private_key=os.environ["POLYMARKET_PRIVATE_KEY"],
)

# List mergeable positions
positions = client.list_positions(mergeable=True).first_page()

# Merge all at once
handle = client.merge_positions(
    positions=[
        {"condition_id": pos.condition_id}
        for pos in positions.items
    ],
)
outcome = handle.wait()
print(f"Merged: {outcome.transaction_hash}")
```
