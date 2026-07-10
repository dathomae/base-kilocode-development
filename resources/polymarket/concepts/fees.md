# Fees

Trading fees on Polymarket.

## Overview

Fees are charged on certain markets, set by the protocol, and applied at match time. Taker fees fund the **Maker Rebates Program**, which redistributes fees daily to market makers for deeper liquidity and tighter spreads. Takers earn a portion back through the tiered **Taker Rebate Program**.

**Geopolitical and world events markets are fee-free.** No fees on USDC deposits or withdrawals (though intermediaries like Coinbase or MoonPay may charge their own fees).

## Fee Formula

Fees use a curve-based formula that peaks at 50% probability:

```
fee = C × feeRate × p × (1 - p)
```

Where **C** = shares traded, **p** = price of shares.

**Makers are never charged fees.** Only takers pay.

## Fee Rates by Category

| Category | Taker Fee Rate | Maker Fee Rate | Maker Rebate |
|---|---|---|---|
| Crypto | 0.07 | 0 | 20% |
| Sports | 0.05 | 0 | 15% |
| Finance | 0.04 | 0 | 25% |
| Politics | 0.04 | 0 | 25% |
| Economics | 0.05 | 0 | 25% |
| Culture | 0.05 | 0 | 25% |
| Weather | 0.05 | 0 | 25% |
| Other / General | 0.05 | 0 | 25% |
| Mentions | 0.04 | 0 | 25% |
| Tech | 0.04 | 0 | 25% |
| Geopolitics | 0 | 0 | — |

The fee in USDC peaks at 50% probability and decreases symmetrically toward both extremes (0% and 100%).

### Example: Finance/Politics/Tech (100 shares at 0.04 rate)

| Price | Trade Value | Taker Fee (USDC) |
|---|---|---|
| $0.01 | $1 | $0.04 |
| $0.10 | $10 | $0.36 |
| $0.25 | $25 | $0.75 |
| $0.50 | $50 | $1.00 |
| $0.75 | $75 | $0.75 |
| $0.90 | $90 | $0.36 |
| $0.99 | $99 | $0.04 |

## Fee Precision

Fees are rounded to 5 decimal places. The smallest fee charged is **0.00001 USDC**. Very small trades near the extremes may incur no fee at all.

## Fee Handling in SDK

Fees are calculated and applied at match time by the protocol. The SDK handles everything automatically. To query fee parameters for a specific market:

```python
info = client.get_clob_market_info(condition_id)
# info["fd"] = { "r": fee_rate, "e": exponent, "to": taker_only }
```

Markets with fees enabled have `feesEnabled` set to `true` on the market object.
