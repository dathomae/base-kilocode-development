# Resolution & Redeeming

How markets are resolved and winning positions redeemed.

## Overview

When the outcome of an event becomes known, the market is **resolved**. Resolution determines which outcome won, allowing holders of winning tokens to redeem them for $1 each. Losing tokens become worthless.

Polymarket uses the **UMA Optimistic Oracle** for decentralized, permissionless resolution — anyone can propose an outcome; anyone can dispute it if they believe it's incorrect.

## Resolution Rules

Every market has pre-defined resolution rules specifying:
- **Resolution source** — Where the outcome is determined from (e.g., official announcements, specific websites)
- **End date** — When the market is eligible for resolution
- **Edge cases** — How ambiguous situations should be handled

**Always read the resolution rules before trading.** The market title describes the question, but the rules define how it resolves.

## Resolution Process

### 1. Proposal

Anyone can propose a resolution by:
1. Selecting the winning outcome
2. Posting a bond (typically $750 pUSD)
3. Submitting the proposal to the UMA Oracle

If correct and undisputed, the proposer gets their bond back plus a reward. **Warning**: proposing incorrectly or too early loses the entire bond.

### 2. Challenge Period (2 hours)

After a proposal, a **2-hour** challenge period begins:
- **No dispute**: Proposal accepted, market resolves
- **Disputed**: New proposal round. If second proposal is also disputed, resolution escalates to UMA's DVM (Data Verification Mechanism) for a token holder vote

Three possible flows:
1. **No dispute** — Propose then Resolve (fastest, ~2 hours)
2. **One dispute** — Propose, Challenge, second Propose, Resolve
3. **Two disputes** — Propose, Challenge, second Propose, second Challenge, Resolve via DVM vote

### 3. Dispute Process

To dispute a proposal:
1. Post a counter-bond (same amount as proposer, typically $750)
2. Dispute triggers a new proposal round (or debate period if second round)

During the **24-48 hour debate period**, evidence is submitted in UMA's Discord channels.

### 4. UMA Vote (~48 hours)

After debate, UMA token holders vote on the correct outcome (~48 hours):

| Outcome | Result | Bond Distribution |
|---|---|---|
| **Proposer wins** | Original proposal accepted | Proposer gets bond + half of disputer's bond |
| **Disputer wins** | Proposal rejected | Disputer gets bond + half of proposer's bond |
| **Too Early** | Event hasn't concluded | Disputer gets bond + half of proposer's bond |
| **Unknown/50-50** | Neither outcome applicable (rare) | Market resolves 50/50 — each token redeems for $0.50; disputer gets bond + half of proposer's bond |

## After Resolution

- **Trading stops** — tokens can no longer be bought or sold
- **Winning tokens** → redeemable for $1.00 each
- **Losing tokens** → worthless ($0.00)

### Redeeming Tokens

Redeem through the CTF collateral adapter: burns ERC1155 outcome tokens through the CTF contract, receives released USDC.e collateral, wraps into pUSD, returns pUSD to your wallet.

```
100 winning tokens → $100 pUSD
```

## Clarifications

In rare cases, Polymarket may issue an **"Additional context"** update:
- Cannot change the fundamental intent of the question
- Published onchain via the bulletin board contract
- Considered by UMA voters when resolving disputes

## Resolution Timeline

| Phase | Duration |
|---|---|
| Challenge period | 2 hours |
| Debate period (if disputed) | 24-48 hours |
| UMA voting (if disputed) | ~48 hours |

**Undisputed resolution**: ~2 hours after proposal
**Disputed resolution**: 4-6 days total

## Key Contract Addresses (Polygon)

| Contract | Address |
|---|---|
| UmaCtfAdapter v3.0 | `0x157Ce2d672854c848c9b79C49a8Cc6cc89176a49` |
| UmaCtfAdapter v2.0 | `0x6A9D222616C90FcA5754cd1333cFD9b7fb6a4F74` |
| UmaCtfAdapter v1.0 | `0xCB1822859cEF82Cd2Eb4E6276C7916e692995130` |
