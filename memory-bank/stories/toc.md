# Stories Table of Contents

This file contains a table of active stories for the project. Stories that are pending or currently in progress are listed below.

**For records of canceled stories, see [`canceled-stories/toc.md`](canceled-stories/toc.md). For records of finished stories, see [`finished-stories/toc.md`](finished-stories/toc.md). Consult these records only when information about canceled or finished stories is required.**

## Active Stories

Stories that are pending or currently in progress.

**[Story001_devcontainer-database-schema.md](Story001_devcontainer-database-schema.md)** — _Done_
: Add a PostgreSQL database service to the devcontainer and create the prediction markets database schema

**[Story002_market-interface-module.md](Story002_market-interface-module.md)** — _Not Started_
: Create a Python module that provides a clean, abstracted interface to Polymarket's APIs with rate limiting and mild future-proofing for Kalshi integration.

**[Story003_order-book-snapshot-collector.md](Story003_order-book-snapshot-collector.md)** — _Not Started_
: Build the CLI-driven collector that captures order book snapshots from Polymarket at regular intervals for taker analysis

**[Story004_metadata-price-history-collectors.md](Story004_metadata-price-history-collectors.md)** — _Not Started_
: Build the CLI-driven collectors that discover Polymarket markets and download historical price data, storing everything in PostgreSQL

**[Story005_deviation-analysis-annualized-returns.md](Story005_deviation-analysis-annualized-returns.md)** — _Not Started_
: Analyze collected market data to detect arbitrage opportunities and compute annualized returns for maker and taker scenarios
