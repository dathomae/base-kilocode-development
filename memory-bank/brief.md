# Prediction Markets Arbitrage

A Python-based exploration of arbitrage opportunities across prediction markets, starting with PolyMarket and expanding to Kalshi. The project aims to identify exploitable arbitrage positions where the gap between equivalent positions on different markets is wide enough and persistent enough to yield significant returns through computerized trading.

## Purpose

The primary goal is to make money by trading both sides of positions/markets where mispricing creates arbitrage. This requires not just identifying theoretical opportunities, but determining which opportunities are practically exploitable — persisting long enough to act on and offering returns that justify the effort.

## Phased Approach

- **Phase 1: PolyMarket** — Upload API docs, understand concepts/data model, explore available data to characterize the market and identify initial signals
- **Phase 2: Kalshi** — Extend the analysis to Kalshi for cross-market arbitrage
- **Phase 3: Arbitrage Engine** — Build automated detection and execution based on findings
