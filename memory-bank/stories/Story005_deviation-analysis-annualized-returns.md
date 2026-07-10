# Story005: Deviation Analysis & Annualized Returns

## Goal

Analyze the collected market data to detect arbitrage opportunities (buying both Yes and No tokens for less than $1 after fees) and compute annualized returns. This story implements the taker scenario (using order book snapshot data with Polymarket fee computation), the maker scenario (using token price data with no fees), and produces console summaries, CSV exports, and timeseries data for charting.

This story implements [Story Plan: Deviation Analysis & Annualized Returns](../plans/story-deviation-analysis-annualized-returns.md).

## References

- [Story Plan: Deviation Analysis & Annualized Returns](../plans/story-deviation-analysis-annualized-returns.md)
- [Polymarket Exploration Plan](../plans/polymarket-exploration-plan.md)

## Dependencies

- [Story001_devcontainer-database-schema.md](Story001_devcontainer-database-schema.md) — provides `db/__init__.py` (connection utilities) and `db/queries.py` (schema definitions for `order_book_snapshots`, `token_prices`, `markets`) — Not Started
- [Story003_order-book-snapshot-collector.md](Story003_order-book-snapshot-collector.md) — populates `order_book_snapshots` table with bid/ask data for taker analysis — Not Started
- [Story004_metadata-price-history-collectors.md](Story004_metadata-price-history-collectors.md) — populates `markets` and `token_prices` tables for maker analysis — Not Started

Story002 (market interface module) is an indirect dependency — referenced by the collectors that populate the data tables. No direct import from `market_interfaces` is needed in this analysis story.

## Dependent Stories

None — This is the final analysis story for Phase 1. No stories have been created that depend on its output.

## Tasks

### Task Sizing

- **Task 1** is classified as **[Small]**. One new file (`__init__.py`), purely additive, follows Python packaging convention. Trivial — no logic.

- **Task 2** is classified as **[Small]**. Two new files (`models.py` + `test_models.py`), purely additive. One dataclass with typed fields and defaults — no methods, no interfaces, no complex logic. The test exercises construction with various field combinations. Narrow scope, single concern.

- **Task 3** is classified as **medium** (not `[Small]`). Two new files (`deviation_detector.py` + `test_deviation_detector.py`), but the implementation requires reasoning about two independent detection functions (taker with fees, maker without fees), two distinct SQL query patterns with server-side cursors, Polymarket fee formula computation, and edge cases across both scenarios (empty database, price extremes, missing end_date, fee_rate=0). 16 subtask steps exercising two entry-point functions. The architect-for-story-planning evaluated this task: both functions share `deviation_detector.py` as their source file — splitting into separate tasks would create merge conflicts with no isolation benefit. The task is at its natural granularity.

- **Task 4** is classified as **[Small]**. Two new files (`annualized_return.py` + `test_annualized_return.py`), purely additive. One computation function with a straightforward formula: `(gross_profit / total_cost) * (365 / days) * 100` with `days` clamped to a minimum of 1. Narrow scope, single concern, no ambiguity.

- **Task 5** is classified as **medium** (not `[Small]`). Two new files (`report.py` + `test_report.py`), but the implementation coordinates three output formats (console summary with distribution statistics and category breakdown, CSV exports for opportunities and summary statistics, and timeseries export for top 10 markets) through a single `generate_report()` function. 13 subtask steps requiring reasoning about percentile calculations, category grouping, top-N filtering, time-based frequency analysis, and file I/O orchestration. The architect-for-story-planning evaluated this task: the three output formats are orchestrated through a single entry-point function and share `report.py` as their source file — splitting into separate files (console_report.py, csv_report.py, timeseries_report.py) would add files not specified in the plan and create artificial module boundaries. The task is at its natural granularity.

- **Task 6** is classified as **medium** (not `[Small]`). Two new files (`cli.py` + `test_cli.py`), but this is a new CLI entry point (not extending an existing one) with 3 subcommands (`maker`, `taker`, `full`), global option handling (`--db-url`, `--output-dir`, `--min-annualized-return`), `full` subcommand comparative table logic, and import dispatch wiring. Requires reasoning about argparse subparser architecture and conditional imports. While it follows established patterns from Story003 and Story004, creating a new CLI from scratch with 3 subcommands and comparative output exceeds the narrow-scope Small threshold.

### Parallel Execution

- **Group A** (Tasks 1, 2): Fully disjoint files — Task 1 creates `analysis/__init__.py`; Task 2 creates `analysis/models.py` and `tests/analysis/test_models.py`. Neither task depends on the other's output. Both are purely additive.

- **Group B** (Tasks 3, 4): Both tasks depend on Task 2 (import `Opportunity` from `models.py`) but are file-disjoint — Task 3 creates `deviation_detector.py` + test; Task 4 creates `annualized_return.py` + test. Task 4 does not depend on Task 3's deviation detection functions (it can compute annualized returns from any list of `Opportunity` objects, mockable in tests). Task 3 does not depend on Task 4's annualized return computation.

### Execution Order

```
[Task 1 ‖ Task 2] → [Task 3 ‖ Task 4] → Task 5 → Task 6
```

- **Group A** (Tasks 1, 2) runs first — Task 2 defines the `Opportunity` dataclass that all downstream tasks consume.
- **Group B** (Tasks 3, 4) runs after Group A merge — both import `Opportunity` from Task 2. They run in parallel because they are file-disjoint and semantically independent.
- **Task 5** runs sequentially after Group B merge — it imports `detect_taker_opportunities()` and `detect_maker_opportunities()` from Task 3, and `compute_annualized_return()` from Task 4.
- **Task 6** runs sequentially after Task 5 — it imports all analysis modules for dispatch and runs the comparative `full` subcommand which requires all modules to exist.

### Reintegration

**Group A (Tasks 1, 2):**

- **Merge order**: any order — fully disjoint file sets with no imports between them.
- **Integration tests**: `uv run pytest tests/analysis/test_models.py -v` — verify models tests pass in the merged tree.
- **Watch for conflicts in**: none expected (both tasks create only new files with no shared modification surfaces).

**Group B (Tasks 3, 4):**

- **Merge order**: any order — fully disjoint file sets (`deviation_detector.py` vs `annualized_return.py` and their respective test files). Neither modifies the other's files.
- **Integration tests**:
  1. `uv run pytest tests/analysis/ -v --tb=short` — verify all analysis unit tests (models + detection + annualized return) pass together in the merged tree.
  2. `uv run python -c "from prediction_markets.analysis.models import Opportunity; from prediction_markets.analysis.deviation_detector import detect_taker_opportunities, detect_maker_opportunities; from prediction_markets.analysis.annualized_return import compute_annualized_return; print('All imports OK')"` — verify cross-module imports resolve without conflicts.
- **Watch for conflicts in**:
  - `tests/analysis/__init__.py` — not explicitly created by any task, but both Task 3 and Task 4 create files in `tests/analysis/`. If a sub-agent auto-creates a `__init__.py` for test discovery in a worktree, it could appear in the merge. Verify after merge and remove if present.
  - `src/prediction_markets/analysis/__pycache__/` — bytecode caches may differ per worktree. Git-ignored; no action needed.

**Task 5 → Task 6 sequential (no parallel group):**

After Task 5 is merged, run `uv run pytest tests/analysis/test_report.py -v` to verify report generation passes. Then run `uv run pytest tests/analysis/ -v --tb=short` to confirm the full test suite passes before starting Task 6.

---

### Task 1: Create analysis package file — Not Started [Small]

Create `src/prediction_markets/analysis/__init__.py` — a package marker with a descriptive docstring.

a. Create `src/prediction_markets/analysis/__init__.py` — a package marker containing a docstring that describes the purpose of the analysis package ("Arbitrage opportunity detection and annualized return analysis for prediction market data.") — Not Started

**Acceptance Criteria:**

- `src/prediction_markets/analysis/__init__.py` exists with a descriptive docstring
- The `analysis` directory is a valid Python package (has `__init__.py`)

### Task 2: Create Opportunity dataclass with tests — Not Started [Small]

Create `src/prediction_markets/analysis/models.py` defining the `Opportunity` dataclass used by all analysis modules. Create `tests/analysis/test_models.py` with unit tests for construction and field defaults.

a. Write test `test_opportunity_taker_scenario` in `tests/analysis/test_models.py` — construct an `Opportunity` with `scenario="taker"`, all ask fields populated (`ask_yes=0.49`, `ask_no=0.49`), mid fields `None`, `fee_rate=0.04`, `total_cost=0.999992`, `gross_profit=0.000008`, `gross_return_pct=0.0008`, `end_date` set to a future date, and `days_to_resolution` computed. Assert all fields match expected values. Assert `mid_yes is None` and `mid_no is None`. — Not Started

b. Write test `test_opportunity_maker_scenario` — construct an `Opportunity` with `scenario="maker"`, mid fields populated (`mid_yes=0.49`, `mid_no=0.49`), ask fields `None`, `fee_rate=0.0` (makers pay no fees), `total_cost=0.98`, `gross_profit=0.02`, `gross_return_pct=2.0408`. Assert `ask_yes is None` and `ask_no is None`. — Not Started

c. Write test `test_opportunity_defaults` — construct an `Opportunity` with only required fields. Assert that `category` defaults to `None`, `fee_rate` defaults to `0.0`, `end_date` defaults to `None`, `days_to_resolution` defaults to `None`, `annualized_return_pct` defaults to `None`, ask fields default to `None`, mid fields default to `None`. — Not Started

d. Write test `test_opportunity_without_end_date` — construct an `Opportunity` with `end_date=None`. Assert `days_to_resolution is None` and `annualized_return_pct is None`. — Not Started

e. Implement `Opportunity` dataclass in `src/prediction_markets/analysis/models.py` — a `@dataclass` with fields matching the plan specification:
   - `scenario: str`
   - `timestamp: datetime`
   - `market_id: str`
   - `question: str`
   - `slug: str`
   - `category: str | None = None`
   - `fee_rate: float = 0.0`
   - `ask_yes: float | None = None`
   - `ask_no: float | None = None`
   - `mid_yes: float | None = None`
   - `mid_no: float | None = None`
   - `total_cost: float = 0.0`
   - `gross_profit: float = 0.0`
   - `gross_return_pct: float = 0.0`
   - `end_date: datetime | None = None`
   - `days_to_resolution: float | None = None`
   - `annualized_return_pct: float | None = None`
   Import `datetime` from the standard library. — Not Started

f. Run `uv run pytest tests/analysis/test_models.py -v` and verify all tests pass — Not Started

**Acceptance Criteria:**

- `Opportunity` is a `@dataclass` with all 17 fields as specified
- `scenario`, `timestamp`, `market_id`, `question`, and `slug` are required (no defaults)
- All optional fields (`ask_yes`, `ask_no`, `mid_yes`, `mid_no`, `category`, `fee_rate`, `end_date`, `days_to_resolution`, `annualized_return_pct`, `total_cost`, `gross_profit`, `gross_return_pct`) have sensible defaults
- Taker scenario: `ask_*` fields populated, `mid_*` fields are `None`
- Maker scenario: `mid_*` fields populated, `ask_*` fields are `None`
- All unit tests pass

### Task 3: Implement deviation detector with tests — Not Started

Create `src/prediction_markets/analysis/deviation_detector.py` implementing `detect_taker_opportunities(conn) -> list[Opportunity]` and `detect_maker_opportunities(conn) -> list[Opportunity]`. The taker function queries `order_book_snapshots` joined with `markets`, computes the Polymarket fee formula per side, and identifies opportunities where `total_cost < 1.0`. The maker function queries `token_prices` self-joined with `markets` and identifies opportunities where `mid_yes + mid_no < 1.0` (no fees). Both functions use server-side cursors for memory efficiency with large datasets. Create `tests/analysis/test_deviation_detector.py` with unit tests mocking database cursors.

a. Write test `test_detect_taker_detects_opportunity` in `tests/analysis/test_deviation_detector.py` — mock a database cursor to return one row with `ask_yes=0.49`, `ask_no=0.49`, `fee_rate=0.04`, and all required market fields. Call `detect_taker_opportunities(mock_conn)`. Assert the result is a list with one `Opportunity` where `scenario == "taker"`, `is_opportunity = True` (i.e. `total_cost < 1.0`), `fee_yes ≈ 0.009996` (`1 * 0.04 * 0.49 * 0.51`), `fee_no ≈ 0.009996`, `total_cost ≈ 0.999992`, `gross_profit ≈ 0.000008`. Assert `mid_yes is None` and `mid_no is None`. — Not Started

b. Write test `test_detect_taker_no_opportunity_above_1` — mock a cursor returning `ask_yes=0.50`, `ask_no=0.50`, `fee_rate=0.04`. Assert `total_cost ≈ 1.02` and the result is an empty list (filtered out: `total_cost >= 1.0`). — Not Started

c. Write test `test_detect_taker_fee_rate_zero` — mock `fee_rate=0.0`, `ask_yes=0.49`, `ask_no=0.49`. Assert `fee_yes == 0.0`, `fee_no == 0.0`, `total_cost == 0.98`, and an opportunity is detected. — Not Started

d. Write test `test_detect_taker_price_extremes` — test two cases: (1) `ask_yes=0.0`, `ask_no=1.0`, `fee_rate=0.04` → `fee_yes = 0.0` (because `0.0 × 1.0 = 0.0`), `fee_no = 0.0` (because `1.0 × 0.0 = 0.0`), `total_cost = 1.0`, not an opportunity. (2) `ask_yes=0.0`, `ask_no=0.0` → `total_cost = 0.0`, opportunity detected. — Not Started

e. Write test `test_detect_taker_empty_table` — mock a cursor with `fetchone()` returning `None` (no rows). Assert `detect_taker_opportunities(mock_conn)` returns an empty list. Assert no exception is raised. — Not Started

f. Write test `test_detect_taker_server_side_cursor` — mock `conn.cursor()` and verify it was called with `name` parameter (indicating a server-side cursor). Assert `cursor.fetchmany()` or iteration is used rather than `cursor.fetchall()`. — Not Started

g. Write test `test_detect_taker_missing_end_date` — mock a cursor returning a row with `end_date=None`. Assert the returned `Opportunity` has `end_date is None` and `days_to_resolution is None`. Assert no exception is raised. — Not Started

h. Write test `test_detect_maker_detects_opportunity` — mock a cursor returning one row with `mid_yes=0.49`, `mid_no=0.49`, `fee_rate=0.0`, and all required market fields. Call `detect_maker_opportunities(mock_conn)`. Assert the result is a list with one `Opportunity` where `scenario == "maker"`, `total_cost == 0.98`, `gross_profit == 0.02`, `gross_return_pct == (0.02 / 0.98) * 100`. Assert `ask_yes is None` and `ask_no is None`. Assert `fee_rate == 0.0` (makers pay no fees). — Not Started

i. Write test `test_detect_maker_no_opportunity_above_1` — mock `mid_yes=0.50`, `mid_no=0.51`. Assert `total_cost == 1.01` and the result is an empty list. — Not Started

j. Write test `test_detect_maker_price_extremes` — test `mid_yes=0.0`, `mid_no=0.0` → opportunity detected. Also test `mid_yes=0.5`, `mid_no=0.5` → `total_cost == 1.0`, not an opportunity. — Not Started

k. Write test `test_detect_maker_empty_table` — mock a cursor with no rows. Assert empty list returned, no exception. — Not Started

l. Write test `test_detect_maker_server_side_cursor` — mock `conn.cursor()` and verify server-side cursor is used. — Not Started

m. Write test `test_detect_maker_missing_end_date` — mock `end_date=None`. Assert `end_date is None`, no exception. — Not Started

n. Implement `detect_taker_opportunities(conn) -> list[Opportunity]` in `src/prediction_markets/analysis/deviation_detector.py`:
   - Import `psycopg2.extras` for server-side cursor support, `Opportunity` from `.models`, and `logging`.
   - Module-level logger via `logging.getLogger(__name__)`.
   - Open a server-side named cursor (`conn.cursor(name='taker_detection')`).
   - Execute the taker SQL query from the plan: `SELECT obs.timestamp, obs.market_id, m.question, m.slug, m.fee_rate, m.category, m.end_date, obs.best_ask_yes, obs.best_ask_no FROM order_book_snapshots obs JOIN markets m ON obs.market_id = m.id WHERE obs.best_ask_yes IS NOT NULL AND obs.best_ask_no IS NOT NULL ORDER BY obs.timestamp`.
   - Iterate rows using the cursor as an iterator (server-side cursor streams results).
   - For each row: compute `fee_yes = fee_rate * ask_yes * (1 - ask_yes)`, `fee_no = fee_rate * ask_no * (1 - ask_no)`, `total_cost = ask_yes + ask_no + fee_yes + fee_no`.
   - If `total_cost < 1.0`: compute `gross_profit = 1.0 - total_cost`, `gross_return_pct = (gross_profit / total_cost) * 100`, construct an `Opportunity` with `scenario="taker"`, `ask_yes`/`ask_no` populated, `mid_yes`/`mid_no = None`. Include `days_to_resolution = None` initially (Task 4 computes this).
   - Return `list[Opportunity]` of detected opportunities.
   - Clean up cursor in a `finally` block (call `cursor.close()`).
   - Log summary: `INFO` level with total rows processed and opportunities found. — Not Started

o. Implement `detect_maker_opportunities(conn) -> list[Opportunity]` in `deviation_detector.py`:
   - Open a server-side named cursor (`conn.cursor(name='maker_detection')`).
   - Execute the maker SQL query from the plan: `SELECT tp_yes.timestamp, tp_yes.market_id, m.question, m.slug, m.category, m.end_date, tp_yes.price AS mid_yes, tp_no.price AS mid_no FROM token_prices tp_yes JOIN token_prices tp_no ON tp_yes.market_id = tp_no.market_id AND tp_yes.timestamp = tp_no.timestamp AND tp_yes.token_id != tp_no.token_id JOIN markets m ON tp_yes.market_id = m.id ORDER BY tp_yes.timestamp`.
   - Iterate rows using the cursor as an iterator.
   - For each row: `total_cost = mid_yes + mid_no` (no fees for makers), `fee_rate = 0.0`.
   - If `total_cost < 1.0`: compute `gross_profit = 1.0 - total_cost`, `gross_return_pct = (gross_profit / total_cost) * 100`, construct an `Opportunity` with `scenario="maker"`, `mid_yes`/`mid_no` populated, `ask_yes`/`ask_no = None`.
   - Return `list[Opportunity]` of detected opportunities.
   - Clean up cursor in a `finally` block.
   - Log summary. — Not Started

p. Run `uv run pytest tests/analysis/test_deviation_detector.py -v` and verify all tests pass — Not Started

**Acceptance Criteria:**

- `detect_taker_opportunities(conn)` returns opportunities where `ask_yes + ask_no + fee_yes + fee_no < 1.0`
- `detect_maker_opportunities(conn)` returns opportunities where `mid_yes + mid_no < 1.0`
- Fee formula uses exact Polymarket specification: `fee = fee_rate × price × (1 - price)` per side (size = 1)
- Server-side cursors are used (not `fetchall()`) for memory efficiency with 4M+ rows
- Empty database returns an empty list without error
- `end_date = NULL` rows produce `Opportunity` objects with `end_date = None` (no crash)
- `fee_rate = 0` correctly produces zero fees
- Price extremes (`0.0` or `1.0`) correctly produce `fee = 0.0` (via `p × (1-p) = 0`)
- Taker opportunities have `ask_*` populated and `mid_* = None`
- Maker opportunities have `mid_*` populated, `ask_* = None`, and `fee_rate = 0.0`
- All unit tests pass with mocked database cursors

### Task 4: Implement annualized return calculation with tests — Not Started [Small]

Create `src/prediction_markets/analysis/annualized_return.py` implementing `compute_annualized_return(opportunities: list[Opportunity]) -> list[Opportunity]` that computes `annualized_return_pct` and `days_to_resolution` for each opportunity. Create `tests/analysis/test_annualized_return.py` with unit tests covering the computation formula, edge cases, and in-place mutation verification.

a. Write test `test_computes_annualized_return_normal` in `tests/analysis/test_annualized_return.py` — create an `Opportunity` with `total_cost=0.98`, `gross_profit=0.02`, `timestamp=datetime(2026, 1, 1)`, `end_date=datetime(2026, 2, 1)` (31 days). Call `compute_annualized_return([opp])`. Assert `opp.days_to_resolution == 31`. Assert `opp.annualized_return_pct == pytest.approx((0.02 / 0.98) * (365 / 31) * 100, rel=1e-6)`. — Not Started

b. Write test `test_clamps_days_to_1` — create an `Opportunity` with `timestamp` and `end_date` on the same day (0 days difference). Call `compute_annualized_return([opp])`. Assert `opp.days_to_resolution == 1` (clamped to minimum 1 to avoid division by zero). Assert `opp.annualized_return_pct` is computed using `days=1`. — Not Started

c. Write test `test_none_end_date_produces_none` — create an `Opportunity` with `end_date=None`. Call `compute_annualized_return([opp])`. Assert `opp.days_to_resolution is None`. Assert `opp.annualized_return_pct is None`. — Not Started

d. Write test `test_negative_days_clamped` — create an `Opportunity` where `end_date` is before `timestamp` (e.g., market already resolved before the observation). `days = (end_date - timestamp).days` would be negative. Assert `opp.days_to_resolution` is clamped to 1. — Not Started

e. Write test `test_zero_gross_profit` — create an `Opportunity` with `gross_profit=0.0`, `total_cost=1.0`, 30 days. Assert `annualized_return_pct == 0.0`. — Not Started

f. Write test `test_mutates_in_place` — create two `Opportunity` objects in a list. Call `compute_annualized_return(opportunities)`. Assert the function returns the same list object (mutates in place). Assert both objects have `annualized_return_pct` and `days_to_resolution` set. — Not Started

g. Implement `compute_annualized_return(opportunities: list[Opportunity]) -> list[Opportunity]` in `src/prediction_markets/analysis/annualized_return.py`:
   - Import `Opportunity` from `.models`.
   - For each opportunity: if `opportunity.end_date` is `None`, set `days_to_resolution = None` and `annualized_return_pct = None`.
   - Otherwise: compute `days = max((opportunity.end_date - opportunity.timestamp).days, 1)`. Set `opportunity.days_to_resolution = days`. Compute `opportunity.annualized_return_pct = (opportunity.gross_profit / opportunity.total_cost) * (365 / days) * 100`.
   - Return the list (mutated in place for efficiency). — Not Started

h. Run `uv run pytest tests/analysis/test_annualized_return.py -v` and verify all tests pass — Not Started

**Acceptance Criteria:**

- `compute_annualized_return()` computes `(gross_profit / total_cost) * (365 / days) * 100` for each opportunity with a known `end_date`
- `days` is clamped to a minimum of 1 (handles same-day resolution and negative date differences)
- Opportunities with `end_date = None` have `days_to_resolution = None` and `annualized_return_pct = None`
- Zero gross profit produces `annualized_return_pct = 0.0`
- The function mutates opportunities in place and returns the same list
- All unit tests pass

### Task 5: Implement report generation with tests — Not Started

Create `src/prediction_markets/analysis/report.py` implementing `generate_report(opportunities: list[Opportunity], output_dir: str, scenario_filter: str | None = None) -> None` that produces a console summary, CSV exports, and timeseries data. The `scenario_filter` parameter restricts to `"maker"` or `"taker"` for single-scenario output, or `None` for combined output. Create `tests/analysis/test_report.py` with unit tests mocking file I/O and verifying output correctness.

a. Write test `test_console_summary_basic_stats` in `tests/analysis/test_report.py` — create a list of 10 mock `Opportunity` objects with known `gross_profit`, `gross_return_pct`, `annualized_return_pct` values. Call `generate_report(opportunities, tmp_path)`. Assert the output contains "Total observations analyzed", "Total opportunities found", and the correct count and percentage. — Not Started

b. Write test `test_console_summary_distribution_percentiles` — create 100 `Opportunity` objects with `annualized_return_pct` values from 0 to 99. Assert the output contains min, max, mean, median, 25th, 75th, 90th, 95th, and 99th percentile values for gross profit, gross return, and annualized return. Verify the median of `annualized_return_pct` is approximately 49.5. — Not Started

c. Write test `test_console_summary_category_breakdown` — create `Opportunity` objects across 3 categories ("Politics", "Sports", "Crypto") with different counts and mean returns. Assert the output shows per-category counts and mean annualized returns. — Not Started

d. Write test `test_console_summary_top_20` — create 50 `Opportunity` objects with varying `annualized_return_pct` values. Assert only the top 20 (by annualized return) are listed in the output, sorted descending. — Not Started

e. Write test `test_console_summary_opportunity_frequency` — create `Opportunity` objects spread across 10 timestamps. Assert the output shows a frequency description (clustered vs. distributed). Assert the output includes the count per time bucket. — Not Started

f. Write test `test_csv_export_maker` — create 5 maker `Opportunity` objects. Call `generate_report(opportunities, tmp_path, scenario_filter="maker")`. Assert `tmp_path / "opportunities_maker.csv"` exists. Read the CSV and verify all 5 rows with correct columns: `scenario, timestamp, market_id, question, slug, category, fee_rate, ask_yes, ask_no, mid_yes, mid_no, total_cost, gross_profit, gross_return_pct, end_date, days_to_resolution, annualized_return_pct`. — Not Started

g. Write test `test_csv_export_taker` — create 5 taker `Opportunity` objects. Call `generate_report(opportunities, tmp_path, scenario_filter="taker")`. Assert `tmp_path / "opportunities_taker.csv"` exists with correct data. — Not Started

h. Write test `test_csv_export_summary_statistics` — create mixed opportunities. Assert `tmp_path / "summary_statistics.csv"` exists with rows for both maker and taker scenarios containing aggregate statistics. — Not Started

i. Write test `test_timeseries_top_10_markets` — create opportunities for 15 markets with varying opportunity counts. Assert `tmp_path / "deviation_timeseries.csv"` exists. Assert only the top 10 markets (by opportunity count) appear. Assert columns include `timestamp, market_id, question, sum_of_prices, is_opportunity, gross_profit`. — Not Started

j. Write test `test_empty_opportunities_list` — call `generate_report([], tmp_path)`. Assert the console output contains a message like "No opportunities found" or "No data available". Assert no CSV files are created (or empty files with headers only). Assert no exception is raised. — Not Started

k. Write test `test_missing_end_date_opportunities_handled` — create 5 `Opportunity` objects where 2 have `end_date=None`. Assert annualized return stats exclude the 2 without end dates and note the count. Assert absolute return stats include all 5. — Not Started

l. Write test `test_output_directory_created` — call `generate_report(opportunities, tmp_path / "nested" / "output")`. Assert the nested output directory was created. Assert CSV files are written inside it. — Not Started

m. Implement `generate_report(opportunities: list[Opportunity], output_dir: str, scenario_filter: str | None = None) -> None` in `src/prediction_markets/analysis/report.py`:
   - Import `csv`, `os`, `statistics`, `logging`, `datetime`. Import `Opportunity` from `.models`.
   - Module-level logger via `logging.getLogger(__name__)`.
   - If `scenario_filter` is provided, filter opportunities to that scenario.
   - Create `output_dir` via `os.makedirs(output_dir, exist_ok=True)`.
   - **Console summary** (printed via `print()` and logger):
     - Total observations (count of opportunities), total opportunities found, and percentage (of total opportunities relative to observations — use `len(opportunities)` as the count since only opportunities are passed in).
     - Distribution statistics: extract lists of `gross_profit`, `gross_return_pct`, and `annualized_return_pct` (excluding `None` values for annualized). For each: min, max, mean, median, 25th, 75th, 90th, 95th, 99th percentiles. Use `statistics.quantiles()` or manual sorted-index calculation.
     - Category breakdown: group by `category` (use `"Uncategorized"` for `None`). For each category: count and mean `annualized_return_pct` (excluding `None`).
     - Top 20 opportunities ranked by `annualized_return_pct` descending. For each: `market_id`, `question`, `scenario`, `gross_return_pct`, `annualized_return_pct`.
     - Opportunity frequency over time: group by day (truncate `timestamp` to date), count opportunities per day. Note whether distribution appears clustered or evenly distributed.
   - **CSV exports**:
     - `opportunities_maker.csv`: all maker opportunities, same columns as test (f). Use `csv.DictWriter`.
     - `opportunities_taker.csv`: all taker opportunities.
     - `summary_statistics.csv`: rows for "maker" and "taker" with columns: `scenario, count, mean_gross_profit, mean_gross_return_pct, mean_annualized_return_pct, median_annualized_return_pct, max_annualized_return_pct, opportunities_above_10pct_annualized`. Compute each aggregate from the filtered opportunity lists.
   - **Timeseries export**:
     - `deviation_timeseries.csv`: identify top 10 markets by opportunity count. For each of those markets, export every observation (not just opportunities) with columns `timestamp, market_id, question, sum_of_prices, is_opportunity, gross_profit`. For taker scenario, `sum_of_prices = ask_yes + ask_no`; for maker scenario, `sum_of_prices = mid_yes + mid_no`.
   - **Combined scenario** (`scenario_filter=None`): after printing per-scenario summaries, print a comparative table:
     ```
     | Metric | Maker | Taker |
     |--------|-------|-------|
     | Observations | N | N |
     | Opportunities | N (P%) | N (P%) |
     | Mean gross return | X% | X% |
     | Mean annualized return | X% | X% |
     | Opportunities with return > 10% annualized | N | N |
     ```
   - **Edge cases**: empty list → print "No opportunities found for analysis." and exit without creating files. Opportunities with `annualized_return_pct=None` → exclude from annualized stats, include count in notes. — Not Started

n. Run `uv run pytest tests/analysis/test_report.py -v` and verify all tests pass — Not Started

**Acceptance Criteria:**

- `generate_report(opportunities, output_dir)` prints console summary with distribution statistics (min/max/mean/median/percentiles), category breakdown, top 20, and frequency over time
- CSV exports are written to `output_dir`: `opportunities_maker.csv`, `opportunities_taker.csv`, `summary_statistics.csv`
- Timeseries CSV (`deviation_timeseries.csv`) exports top 10 markets by opportunity count
- Opportunities with `end_date=None` are excluded from annualized return stats with a note about the count
- Empty opportunities list prints a descriptive message and creates no CSV files (no crash)
- Nested output directories are created automatically
- Per-scenario filtering (`scenario_filter`) correctly restricts output
- Combined output (`scenario_filter=None`) produces a comparative maker-vs-taker table
- Console output uses `print()` for formatted summaries and `logging` for progress/debug messages
- All unit tests pass with mocked file I/O

### Task 6: Implement CLI entry point with tests — Not Started

Create `src/prediction_markets/analysis/cli.py` providing `python -m prediction_markets.analysis` with `maker`, `taker`, and `full` subcommands. The CLI orchestrates the full analysis pipeline: detect opportunities, compute annualized returns, filter by threshold, and generate reports. Create `tests/analysis/test_cli.py` with unit tests verifying the parser structure and handler dispatch.

a. Write test `test_maker_subparser_exists` in `tests/analysis/test_cli.py` — call `create_parser()` and verify a subparser named `maker` exists. Verify it accepts the global options: `--db-url` (type str, default None), `--output-dir` (type str, default `./output`), `--min-annualized-return` (type float, default 0.0). — Not Started

b. Write test `test_taker_subparser_exists` — verify `taker` subparser exists with the same global options as `maker`. — Not Started

c. Write test `test_full_subparser_exists` — verify `full` subparser exists with the same global options. — Not Started

d. Write test `test_parse_args_defaults` — call `parse_args(['maker'])` and verify `command='maker'`, `db_url=None`, `output_dir='./output'`, `min_annualized_return=0.0`. — Not Started

e. Write test `test_parse_args_custom_values` — call `parse_args(['taker', '--db-url', 'postgresql://localhost/test', '--output-dir', '/tmp/reports', '--min-annualized-return', '25.0'])` and verify all custom values. — Not Started

f. Write test `test_maker_dispatches_to_detector_and_report` — patch `detect_maker_opportunities` with a `Mock(return_value=[mock_opp])`, patch `compute_annualized_return` with a `Mock(side_effect=lambda x: x)`, patch `generate_report` with a `Mock()`. Set `sys.argv` to `['analysis', 'maker']`. Call `main()`. Assert `detect_maker_opportunities` was called once. Assert `compute_annualized_return` was called once with the list of opportunities. Assert `generate_report` was called once with the opportunities, `'./output'`, and `scenario_filter='maker'`. — Not Started

g. Write test `test_taker_dispatches_to_detector_and_report` — same pattern using `detect_taker_opportunities` and `scenario_filter='taker'`. — Not Started

h. Write test `test_full_runs_both_scenarios` — patch `detect_maker_opportunities`, `detect_taker_opportunities`, `compute_annualized_return`, and `generate_report`. Set `sys.argv` to `['analysis', 'full']`. Call `main()`. Assert `detect_maker_opportunities` and `detect_taker_opportunities` were both called. Assert `generate_report` was called with `scenario_filter=None` (combined report). — Not Started

i. Write test `test_min_annualized_return_filter` — patch `detect_taker_opportunities` to return 3 opportunities with `annualized_return_pct` values of 5.0, 15.0, and 30.0 (set after `compute_annualized_return` mock). Call `main()` with `--min-annualized-return 10.0`. Assert `generate_report` was called with only the 2 opportunities ≥ 10.0%. Assert the filtered-out count is logged. — Not Started

j. Write test `test_empty_database_message_maker` — patch `detect_maker_opportunities` to return `[]`. Patch `generate_report`. Call `main()` with `['analysis', 'maker']`. Assert a message like "No data available for maker analysis" is printed/logged. Assert `generate_report` is not called (or called with empty list). Assert `main()` exits without error. — Not Started

k. Write test `test_empty_database_message_taker` — same pattern using `detect_taker_opportunities` returning `[]`. — Not Started

l. Implement `add_global_args(parser: argparse.ArgumentParser) -> None` in `src/prediction_markets/analysis/cli.py` — adds `--db-url` (type str, default None, help describing optional database connection string override), `--output-dir` (type str, default './output', help describing directory for CSV exports), and `--min-annualized-return` (type float, default 0.0, help describing minimum annualized return percentage threshold to include in reports) to the provided subparser. — Not Started

m. Implement `create_parser() -> argparse.ArgumentParser` in `cli.py` — creates the root `ArgumentParser` with a project-level description ("Prediction market arbitrage analysis — detect deviations and compute annualized returns."). Adds subparsers with `dest='command'` and `required=True`. Adds three subparsers:
    - `maker` — `subparsers.add_parser('maker', help='Analyze maker scenario using token_prices data (no fees)')`
    - `taker` — `subparsers.add_parser('taker', help='Analyze taker scenario using order_book_snapshots data (with fees)')`
    - `full` — `subparsers.add_parser('full', help='Run both maker and taker analyses with combined report')`
    - Calls `add_global_args()` on each subparser. — Not Started

n. Implement `main() -> None` in `cli.py`:
    - Parse args via `create_parser().parse_args()`.
    - Get database connection: if `args.db_url`, use `psycopg2.connect(args.db_url)`; otherwise use `get_connection()` from `prediction_markets.db`.
    - Create output directory: `os.makedirs(args.output_dir, exist_ok=True)`.
    - Module-level logger via `logging.getLogger(__name__)`.
    - Configure logging: `INFO` level by default; add a `StreamHandler` if none exists (avoid double-handler issues when pytest captures logs). Use `logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')` inside a `if __name__ == '__main__'` guard or in `main()` with a guard against re-initialization.
    - **For `maker` subcommand**:
      - Log `INFO`: "Starting maker scenario analysis..."
      - Call `detect_maker_opportunities(conn)`.
      - If empty: print "No data available for maker analysis. Run the price history collector first." and exit.
      - Call `compute_annualized_return(opportunities)`.
      - Filter by `args.min_annualized_return`: keep only opportunities where `annualized_return_pct is None or annualized_return_pct >= args.min_annualized_return`. Log `INFO` count of filtered-out opportunities.
      - Call `generate_report(opportunities, args.output_dir, scenario_filter='maker')`.
      - Log `INFO`: "Maker scenario analysis complete."
    - **For `taker` subcommand**:
      - Same pattern using `detect_taker_opportunities(conn)` and `scenario_filter='taker'`.
      - Empty message: "No data available for taker analysis. Run the order book snapshot collector first."
    - **For `full` subcommand**:
      - Log `INFO`: "Starting full analysis (maker + taker)..."
      - Run maker detection, compute annualized returns, filter.
      - Run taker detection (separate connection or same connection), compute annualized returns, filter.
      - Combine both lists.
      - If both are empty: print "No data available for analysis. Run the collectors first." and exit.
      - If only one has data: run single-scenario report for that scenario with a note about the other being empty.
      - If both have data: call `generate_report(combined_opportunities, args.output_dir, scenario_filter=None)`.
      - Log `INFO`: "Full analysis complete."
    - Close database connection after analysis. — Not Started

o. Create `src/prediction_markets/analysis/__main__.py` — imports `main` from `prediction_markets.analysis.cli` and invokes it:
    ```
    """Entry point for python -m prediction_markets.analysis."""
    from prediction_markets.analysis.cli import main
    main()
    ```
    — Not Started

p. Run `uv run pytest tests/analysis/test_cli.py -v` and verify all tests pass — Not Started

**Acceptance Criteria:**

- `maker` subcommand exists with `--db-url`, `--output-dir` (default `./output`), `--min-annualized-return` (default `0.0`)
- `taker` subcommand exists with the same global options
- `full` subcommand exists with the same global options
- `main()` dispatches `maker` to `detect_maker_opportunities → compute_annualized_return → generate_report`
- `main()` dispatches `taker` to `detect_taker_opportunities → compute_annualized_return → generate_report`
- `main()` dispatches `full` to both maker and taker analysis with combined report
- `--min-annualized-return` filters opportunities below the threshold
- Empty database produces a descriptive "No data available" message without crash
- Database connection is handled via `get_connection()` (with `--db-url` override)
- All CLI tests pass with mocked detector and report functions

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks (Tasks 2–6), tests are written before implementation. Test subtasks precede implementation subtasks within each task. All tests use `unittest.mock` to patch external dependencies:

- **Task 2 (models)**: No external dependencies — tests construct `Opportunity` objects directly and assert field values.
- **Task 3 (deviation_detector)**: Tests mock `psycopg2` cursors to return controlled row data. No live database or network access is required.
- **Task 4 (annualized_return)**: Tests construct `Opportunity` objects with known `timestamp`/`end_date`/`gross_profit` values and assert the computed output.
- **Task 5 (report)**: Tests use `tmp_path` (pytest fixture) for CSV output verification. Console output is captured via `capsys` (pytest fixture) or by patching `print()`. No real file I/O side effects.
- **Task 6 (cli)**: Tests mock `detect_*_opportunities`, `compute_annualized_return`, and `generate_report` to verify dispatch without requiring the analysis modules.

## Constraints

- All file paths in task descriptions are relative to `source_root` (`/prediction-markets`). Commands like `uv run pytest` run from `source_root`.
- Database queries use raw SQL with `psycopg2` cursors (following the same pattern as the `db` module from Story001). The `detect_*` functions accept a `psycopg2` connection object — callers pass either `get_connection()` or a `--db-url` connection.
- Server-side cursors (`conn.cursor(name=...)`) must be used instead of `fetchall()` for memory efficiency with 4M+ row datasets. Iterate the cursor directly (it streams rows).
- Fee computation follows the exact Polymarket formula: `fee = 1 × fee_rate × price × (1 - price)` per side. Do not approximate or simplify.
- Annualized return formula: `(gross_profit / total_cost) * (365 / days) * 100`. Days clamped to minimum 1.
- CLI global options (`--db-url`, `--output-dir`, `--min-annualized-return`) must appear on every subcommand via the `add_global_args()` helper.
- Logging: use Python's standard `logging` module with `logging.getLogger(__name__)` in each module. `INFO` level for progress and summary messages. `WARNING` level for empty data. `ERROR` level for database connection failures. All log messages must be informative and include relevant context (counts, percentages).
- Mock all external dependencies in unit tests — `unittest.mock` for database cursors, detector functions, report generation, and file I/O. Do not require a running database or network access for any test.
- CSV exports use Python's standard `csv.DictWriter`.
- Distribution statistics use Python's `statistics` module where available; manual sorted-index calculation for percentiles not covered by `statistics.quantiles()`.

## Notes

- The `source_root` for this project is the project root (`/prediction-markets`). All file paths in task descriptions are relative to `source_root`. Commands like `uv run pytest` run from `source_root`.
- This story creates the `analysis` package from scratch — the directory `src/prediction_markets/analysis/` does not exist before this story runs.
- The Opportunity dataclass is in `models.py` rather than `deviation_detector.py` to enable parallel execution of Tasks 3 and 4 (both import `Opportunity` from `models` without depending on each other).
- Server-side cursors require transactions — the `detect_*` functions assume the caller manages the connection lifecycle. The cursor is opened inside each function and closed in a `finally` block.
- The `full` subcommand's combined report uses `scenario_filter=None` to trigger the comparative maker-vs-taker table. When both scenarios have data, all opportunities are passed as a single combined list.
- With 500 markets and 30 days of 1-minute snapshots (~21.6M rows in `order_book_snapshots`), the taker analysis is the most computationally intensive step. Server-side cursors ensure constant memory usage regardless of dataset size.
- The `deviation_timeseries.csv` for top 10 markets re-queries the database to get ALL observations (not just opportunities) for those markets. This may require additional SQL queries in the report module — for efficiency, consider passing the raw observation data alongside opportunities, or implement a separate query function in `deviation_detector.py` that returns all observations (not just opportunities) for specified market IDs.
