# Concepts

**Cross-Market Arbitrage**
: The core concept of this project. When the same event (or economically equivalent events) trades on both Polymarket and Kalshi, price differences create arbitrage opportunities. For example, if "Trump wins 2024" trades at $0.55 on Polymarket and $0.58 on Kalshi, buying on Polymarket and selling on Kalshi yields a $0.03 spread per contract.
: Related: Arbitrage, Spread, Prediction Market

**Arbitrage Exploitability**
: Not all theoretical arbitrage opportunities are practically exploitable. Key factors include: (1) Spread magnitude — must exceed transaction costs (gas fees, platform fees, withdrawal fees); (2) Persistence — the opportunity must last long enough to execute both legs; (3) Liquidity depth — there must be sufficient volume on both sides to fill the desired position; (4) Execution speed — automated systems may be required for short-lived opportunities.
: Related: Arbitrage, Slippage, Liquidity

**Equivalent Position Mapping**
: Markets on different platforms may not be identically structured. Mapping equivalent positions requires understanding event semantics, resolution criteria, and payout structures. Two "Trump wins" markets on different platforms may have subtly different resolution rules that affect equivalence.
: Related: Binary Outcome Market, Prediction Market

**Phase-Based Exploration**
: The project proceeds in phases: Phase 1 explores PolyMarket in isolation to understand the data model, API behavior, and market characteristics. Phase 2 adds Kalshi for cross-market analysis. Phase 3 builds automated arbitrage detection and execution. This phased approach allows requirements and design to adapt based on discoveries.
: Related: PolyMarket, Kalshi
