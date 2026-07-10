# Story Plan: Deviation Analysis & Annualized Returns

**Parent Plan**: [Polymarket Exploration Plan](polymarket-exploration-plan.md)
**Dependencies**: Story 1 (Database & Schema), Story 3 (Metadata & Price History data), Story 4 (Order Book Snapshot data)
**Parallel With**: None — final story

## Objective

Analyze the collected market data to detect arbitrage opportunities (buying both Yes and No tokens for less than $1 after fees) and compute annualized returns. Produces clear, data-backed answers to the core hypothesis with no estimation or approximation.

## Deliverables

1. `src/prediction_markets/analysis/__init__.py`
2. `src/prediction_markets/analysis/deviation_detector.py` — core opportunity detection
3. `src/prediction_markets/analysis/annualized_return.py` — return calculations
4. `src/prediction_markets/analysis/report.py` — results output (console + CSV)
5. `src/prediction_markets/analysis/cli.py` — CLI entry point: `python -m prediction_markets.analysis`

## Detailed Requirements

### CLI Entry Point (`analysis/cli.py`)

```
python -m prediction_markets.analysis COMMAND [OPTIONS]
```

**Subcommands**:
- `maker` — analyze the maker scenario (uses `token_prices` data, no fees)
- `taker` — analyze the taker scenario (uses `order_book_snapshots` data, with fees)
- `full` — run both and produce a combined report

**Global options**:
- `--db-url TEXT` (optional): override DB connection
- `--output-dir TEXT` (default: `./output`): directory for CSV exports
- `--min-annualized-return FLOAT` (default: `0.0`): filter to opportunities above this threshold

### Deviation Detector (`deviation_detector.py`)

**Taker analysis** — queries `order_book_snapshots` joined with `markets`:

```sql
SELECT
    obs.timestamp,
    obs.market_id,
    m.question,
    m.slug,
    m.fee_rate,
    m.category,
    m.end_date,
    obs.best_ask_yes,
    obs.best_ask_no,
    obs.best_bid_yes,
    obs.best_bid_no
FROM order_book_snapshots obs
JOIN markets m ON obs.market_id = m.id
WHERE obs.best_ask_yes IS NOT NULL
  AND obs.best_ask_no IS NOT NULL
ORDER BY obs.timestamp
```

For each row, compute:

```python
# Taker costs (actual observed ask prices)
ask_yes = row.best_ask_yes
ask_no  = row.best_ask_no

# Fee: formula from Polymarket docs — fee = size × feeRate × price × (1 - price)
# Size = 1 share per side
fee_yes = 1 * row.fee_rate * ask_yes * (1 - ask_yes)
fee_no  = 1 * row.fee_rate * ask_no  * (1 - ask_no)

total_cost = ask_yes + ask_no + fee_yes + fee_no
is_opportunity = total_cost < 1.0
gross_profit = 1.0 - total_cost
gross_return_pct = (gross_profit / total_cost) * 100
```

**Maker analysis** — queries `token_prices` joined with `markets`:

```sql
SELECT
    tp_yes.timestamp,
    tp_yes.market_id,
    m.question,
    m.slug,
    m.category,
    m.end_date,
    tp_yes.price AS mid_yes,
    tp_no.price  AS mid_no
FROM token_prices tp_yes
JOIN token_prices tp_no
    ON  tp_yes.market_id = tp_no.market_id
    AND tp_yes.timestamp = tp_no.timestamp
    AND tp_yes.token_id != tp_no.token_id
JOIN markets m ON tp_yes.market_id = m.id
ORDER BY tp_yes.timestamp
```

For each row, compute:

```python
# Maker costs (midpoints, zero fees)
mid_yes = row.mid_yes
mid_no  = row.mid_no

total_cost = mid_yes + mid_no    # no fees for makers
is_opportunity = total_cost < 1.0
gross_profit = 1.0 - total_cost
gross_return_pct = (gross_profit / total_cost) * 100
```

Both analyses return a list of `Opportunity` objects:

```python
@dataclass
class Opportunity:
    scenario: str           # "maker" or "taker"
    timestamp: datetime
    market_id: str
    question: str
    slug: str
    category: str | None
    fee_rate: float
    ask_yes: float | None   # None for maker
    ask_no: float | None    # None for maker
    mid_yes: float | None   # None for taker
    mid_no: float | None    # None for taker
    total_cost: float
    gross_profit: float
    gross_return_pct: float
    end_date: datetime | None
    days_to_resolution: float | None
    annualized_return_pct: float | None
```

### Annualized Return (`annualized_return.py`)

For each `Opportunity` with a known `end_date`:

```python
days = max((end_date - timestamp).days, 1)
# Annualized return = (profit / cost) * (365 / days)
annualized = (gross_profit / total_cost) * (365 / days) * 100  # percentage
```

If `end_date` is `NULL` (market has no resolution date configured), `annualized_return_pct` is `None`. The report handles this by showing "N/A" and noting the count of opportunities without resolution dates.

Edge cases:
- `days = 0`: clamp to 1 to avoid division by zero (the market resolves today)
- `days > 365`: annualized return will be lower than absolute return (money tied up longer)

### Report Generation (`report.py`)

**`generate_report(opportunities: list[Opportunity], output_dir: str) -> None`**

Produces:

1. **Console summary** for each scenario:
   - Total observations analyzed
   - Total opportunities found (count and percentage)
   - Distribution: min, max, mean, median, 25th, 75th, 90th, 95th, 99th percentile for:
     - Gross profit ($)
     - Gross return (%)
     - Annualized return (%) (for those with end_dates)
   - Breakdown by category (count, mean annualized return)
   - Top 20 opportunities (ranked by annualized return)
   - Opportunity frequency over time (are they clustered or evenly distributed?)

2. **CSV export** to `{output_dir}/`:
   - `opportunities_maker.csv` — all maker opportunities
   - `opportunities_taker.csv` — all taker opportunities
   - `summary_statistics.csv` — aggregate statistics for both scenarios

3. **Timeseries data** to `{output_dir}/`:
   - `deviation_timeseries.csv` — for the top 10 markets by opportunity count, export the full timeseries of `ask_yes + ask_no` (taker) or `mid_yes + mid_no` (maker) for charting
   - Columns: `timestamp, market_id, question, sum_of_prices, is_opportunity, gross_profit`

### Combined Analysis (`full` subcommand)

Runs both maker and taker analyses, then prints a comparative summary:

| Metric | Maker | Taker |
|--------|-------|-------|
| Observations | N | N |
| Opportunities | N (P%) | N (P%) |
| Mean gross return | X% | X% |
| Mean annualized return | X% | X% |
| Opportunities with return > 10% annualized | N | N |

## Edge Cases

- **No data in database**: Either `token_prices` or `order_book_snapshots` may be empty. Print "No data available for [maker|taker] analysis. Run the corresponding collector first." and exit gracefully.
- **Missing end_date**: Markets without `end_date` can't compute annualized return. Include them in absolute return stats but exclude from annualized stats. Note the count.
- **Single observation**: A market with only one snapshot can't show trends but still contributes to the opportunity count.
- **Very large datasets**: 4M+ snapshot rows. Use server-side cursors or `fetchmany()` to stream results without loading everything into memory.
- **Fee rate edge cases**: Fee_rate of 0 (geopolitics markets) means fees are $0. The formula `fee_rate × p × (1-p)` correctly produces 0.
- **Price = 0 or 1**: At exact extremes, `p × (1-p) = 0`, so fees are 0. This is correct per the fee formula.

## Verification

- Run `maker` analysis with populated `token_prices` → produces report with opportunities (if any exist)
- Run `taker` analysis with populated `order_book_snapshots` → produces report with opportunities (if any exist)
- `full` command produces combined report with comparative table
- CSV files are written to the specified `--output-dir`
- Empty database produces clean "No data" message, not a crash
- A market with `fee_rate = 0.04`, `ask_yes = 0.49`, `ask_no = 0.49` should show `total_cost = 0.49 + 0.49 + 0.009996 + 0.009996 ≈ 0.999992` → `is_opportunity = True`
- A market with `fee_rate = 0.04`, `ask_yes = 0.50`, `ask_no = 0.50` should show `total_cost = 0.50 + 0.50 + 0.01 + 0.01 = 1.02` → `is_opportunity = False`
