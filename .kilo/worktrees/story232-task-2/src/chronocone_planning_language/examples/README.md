# Chronocone Planning Language Examples

Example CPL files exercising various phases of the Chronocone Planning Language pipeline.

## Top-Level Organization

### `parse_and_analyze_only/`

Examples that exercise the **parser and/or analyzer only** — no synthesis. These files
contain no connector declarations and no data files. Each `.cpl` file is a standalone
test of syntax and semantic analysis.

- **`without_errors/`** — CPL files expected to parse and analyze successfully.
- **`with_errors/`**   — CPL files with intentional errors (parse-time or analysis-time).

When a new language capability is implemented at the parse/analyze level but synthesis
isn't ready yet, examples go here. Once synthesis is complete and data files exist,
they graduate to `with_data/`.

### `with_data/`

End-to-end examples that exercise the **full pipeline** — parse, analyze, synthesize,
and execute. Each example directory structure:

```
<category>/
  core.cpl                    # shared concepts, measures, connectors (optional)
  <example>/
    example.cpl               # metric/feature-specific CPL
    iospec.json               # IO specification (inputs, outputs, architecture)
    <data>.parquet            # input parquet files
    expected_output.xlsx      # expected output for verification
```

- **`without_errors/`** — Examples where the full pipeline should succeed.
- **`with_errors/`**     — Examples with intentional errors in the pipeline.

## `.compile_only` Marker Files

Some examples under `with_data/` contain a `.compile_only` file in their directory.
This is a **temporary marker** indicating the example runs through parse and analysis
successfully but the synthesis phase is not yet fully wired for this feature.
These examples are excluded from end-to-end execution and output verification.

Compile-only examples represent intermediate states of incremental feature
development. When synthesis is completed for that capability, the `.compile_only`
marker is removed and `expected_output.xlsx` is added.

## Subdirectory Inventory

### `parse_and_analyze_only/without_errors/`

- `metrics_time_shift/` — time shift syntax and semantic analysis (forward, backward, self-reference)

### `with_data/without_errors/`

- `aggregators_basic_sum_country/` — sum aggregation across Country dimension (compile-only)
- `conditional_grain/`             — conditional grain examples with monthly data
- `lookups/`                       — lookup table examples
- `metrics_arithmetic/`            — arithmetic operations (addition, multiplication, parens, etc.)
- `metrics_conditionals/`          — conditional metrics (if/else, when/default, comparisons)
- `multi_metric/`                  — multi-metric dependency chains
- `units_conversions/`             — unit conversion bridges

### `with_data/with_errors/`

- `units_errors/` — unit analysis errors (incompatible adds, io spec mismatches, unannotated constants)
