# Positions & Tokens

Outcome tokens, positions, and the lifecycle of collateral on Polymarket.

## Outcome Tokens

Every market has exactly two outcome tokens:

| Token | Redeems for | If... |
|---|---|---|
| **Yes** | $1.00 | The event occurs |
| **No** | $1.00 | The event does not occur |

Tokens are **ERC1155** assets on Polygon, using the [Gnosis Conditional Token Framework (CTF)](https://github.com/gnosis/conditional-tokens-contracts/). Every Yes/No pair is backed by exactly $1 of **pUSD** collateral locked in the CTF contract.

## Token Operations

### Split

Convert pUSD into outcome tokens. Splitting $1 creates 1 Yes token and 1 No token:

```
$100 pUSD → 100 Yes tokens + 100 No tokens
```

Use for: creating market-making inventory, obtaining both sides of a market.

### Trade

Buy or sell tokens on the order book. This is how most users acquire positions:

- **Buy Yes** at $0.60 → Pay $0.60, receive 1 Yes token
- **Sell Yes** at $0.60 → Give up 1 Yes token, receive $0.60

Positions can be sold at any time before resolution.

### Merge

Convert a complete set of tokens back into pUSD. Requires equal amounts of Yes and No tokens:

```
100 Yes tokens + 100 No tokens → $100 pUSD
```

Use for: exiting a position without trading, converting accumulated tokens back to collateral.

### Redeem

After a market resolves, exchange winning tokens for pUSD:

```
100 winning tokens → $100 pUSD
```

## Position Value

The value of a position depends on the current market price:

```
Position value = Token balance × Current price
```

Example: holding 100 Yes tokens when Yes is trading at $0.75 → position value = $75.

## Profit and Loss

Profit depends on how the market resolves compared to entry price.

### Example — Buying Yes at $0.40

| Scenario | Outcome | Return | Profit |
|---|---|---|---|
| Event occurs | Yes wins | $1.00 | +$0.60/token (150%)|
| Event doesn't occur | No wins | $0.00 | -$0.40/token (-100%)|

Positions can be closed before resolution to lock in profits or cut losses:
- Bought Yes at $0.40
- Price rises to $0.70
- Sell at $0.70 → Profit of $0.30/token (75%)

## Holding Rewards

Polymarket pays a **4.00% annualized** Holding Reward based on total position value in eligible markets. The position value is randomly sampled once each hour, and the reward is distributed daily. The rate is variable and subject to change at Polymarket's discretion.
