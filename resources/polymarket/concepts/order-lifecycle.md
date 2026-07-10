# Order Lifecycle

How orders flow from creation to settlement on the Polymarket CLOB.

## Overview

Every trade follows a specific lifecycle: orders are created offchain (EIP-712 signed messages), matched by the CLOB operator, and settled onchain through smart contracts on Polygon.

All orders on Polymarket are **limit orders** that are **EIP-712 signed messages**. The signature authorizes the Exchange contract to execute on the user's behalf — without ever taking custody of funds.

## Order Types

| Type | Behavior | Use Case |
|---|---|---|
| **GTC** | Good Till Cancelled — rests on book until filled or cancelled | Standard limit orders |
| **GTD** | Good Till Date — auto-expires at specified time | Time-limited orders |
| **FOK** | Fill Or Kill — fill entirely or cancel immediately | All-or-nothing execution |
| **FAK** | Fill And Kill — fill what's available, cancel the rest | Partial fills acceptable |

### Post-Only Orders

Post-only orders only rest on the book. If they would cross the spread and match immediately, they are rejected. This guarantees the user is always the **maker**.

## Lifecycle Stages

### 1. Create and Sign

The client creates an order object containing:
- Token ID (which outcome)
- Side (buy or sell)
- Price and size
- Expiration time
- Timestamp (in milliseconds, for order uniqueness)

The user signs this with their private key (EIP-712).

### 2. Submit to CLOB

The signed order is submitted to the CLOB operator, which validates:
- Valid signature
- Sufficient balance
- Required allowances set (approve Exchange contract)
- Price meets minimum tick size requirements

### 3. Match or Rest

**If marketable** (buy price ≥ lowest ask, or sell price ≤ highest bid):
- Some markets have a **taker delay** (~250ms) on selected crypto/finance up/down markets
- Sports markets may have a game-delay during live conditions
- During delay, order is pending and cannot be cancelled

**If not marketable**: order rests on the book until:
- Another order matches against it
- User cancels it
- It expires (GTD only)

### 4. Settlement

When orders match, the operator submits the trade onchain. The Exchange contract:
- Verifies both signatures
- Transfers tokens from seller to buyer
- Transfers pUSD from buyer to seller
- Settlement is **atomic** — all or nothing

### 5. Confirmation

Trade achieves finality on Polygon. Token balances update and the trade appears in history.

## Order Statuses

| Status | Description |
|---|---|
| `live` | Order is resting on the book |
| `matched` | Order matched immediately |
| `delayed` | Marketable order accepted into an async delay window |
| `unmatched` | Marketable order placed on the book after delay expired without a match |

## Trade Statuses

| Status | Terminal | Description |
|---|---|---|
| `MATCHED` | No | Trade matched, sent to executor for onchain submission |
| `MINED` | No | Transaction mined into the blockchain |
| `CONFIRMED` | Yes | Trade achieved finality, successful |
| `RETRYING` | No | Transaction failed, being retried |
| `FAILED` | Yes | Trade failed permanently |

## Maker vs Taker

| Role | Description | When |
|---|---|---|
| **Maker** | Adds liquidity | Order rests and is later matched |
| **Taker** | Removes liquidity | Order matches immediately against resting orders |

Price improvement benefits the taker: if you place a buy at $0.55 and it matches a resting sell at $0.52, you pay $0.52.

## Cancellation

Orders can be cancelled at any time before matching via the CLOB API, except during a pending delay window. Partial fills cannot be cancelled — only the unfilled portion of an order can be cancelled.

## Requirements Before Placing Orders

| Requirement | Description |
|---|---|
| **Balance** | Sufficient pUSD (for buys) or tokens (for sells) |
| **Allowance** | Approve the Exchange contract to spend your assets |
| **API Credentials** | Valid API key for authenticated endpoints |

Max order size is limited by available balance minus amounts reserved by existing open orders:

```
maxOrderSize = balance - Σ(openOrderSize - filledAmount)
```
