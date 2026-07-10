# Plan: Story268 — Spreaders Increment 2 (Synthesize)

## Purpose

This plan provides the technical writer with sufficient detail to create Story268: **spreaders-increment2-synthesize** — implement synthesis (code generation) for spreader-integrated metric computation (Increment 2 of Epic008). Generates metric computation code with spreader sequences applied in topological order, producing correct results for single unambiguous spreading paths.

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `spreaders-increment2-synthesize`                                                            |
| description  | Generate metric computation code with spreader integration, scheduled topological computation order, spreader integration examples with single unambiguous paths |

## Source EROs

- [ERO008: Spreaders Incremental Implementation](../specifications/design/cpl/eros/ERO008_spreaders-incremental-implementation.md) — Increment 2 sketch
- [ERO008: Spreaders Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO008_spreaders-synthesis-phase.md) — Increment 2 generated code examples

## Dependencies

- Story267 (spreaders-increment2-analyze) — analysis must insert `CILSpreadNode` chains into metric CIL trees
- Story266 (spreaders-increment1-synthesize) — `WalkCILSpreadNode` handler and `EmitSpreadingMatrix` already exist
- Story261 (aggregators-increment4-synthesize) — metric computation code generation infrastructure (topological scheduling, `ChronoconeComputationController`) already exists; spreader synthesis follows the same pattern
- Story254 (aggregators-increment1-matrix-mult) — `tensorops.ExpandDimension` is the synthesis primitive

## Files to Create or Modify

| File                                                    | Action | Purpose                                                                          |
| ------------------------------------------------------- | ------ | -------------------------------------------------------------------------------- |
| `src/main/go/synthesis/spreader_integration_synth.go`   | Create | Generate metric computation code with spreader chains                              |
| `src/main/go/synthesis/spreader_integration_synth_test.go` | Create | Unit tests: verify generated code for spreader-integrated metrics                  |
| `examples/spreaders/integration/` (in appropriate with_data dir) | Create | End-to-end example with spreader-integrated metric computation               |

### Generated Code Pattern — Spreader-Integrated Metric

When a metric's CIL tree contains `CILSpreadNode` chains inserted during Increment 2 analysis, the synthesizer generates code that applies the spreaders in order:

```go
// Generated Go code for metric: regional_targets
//   inputs: metric global_target (scalar [])
//   output: [Country, Product]

func computeRegionalTargets(controller *ChronoconeComputationController) (*ChronoconeTensor[float64], error) {
    globalTarget := controller.GetMetricComputation("global_target").Tensor().(*ChronoconeTensor[float64])

    // Step 1: Spread from [] to [Country] = 1/size(Country) * Country
    spreadToCountry, err := tensorops.ExpandDimension(
        globalTarget,
        spreadingMatrix_Country,
        0, 0,
    )
    if err != nil {
        return nil, fmt.Errorf("spread to country failed: %w", err)
    }

    // Step 2: Spread from [Country] to [Country, Product] = 1/size(Product) * Product
    result, err := tensorops.ExpandDimension(
        spreadToCountry,
        spreadingMatrix_Product,
        0, 1,  // expand dim 0 (Country), insert new dim at position 1 (after Country)
    )
    if err != nil {
        return nil, fmt.Errorf("spread to product failed: %w", err)
    }

    return result, nil
}
```

### Key Synthesis Details

1. **Spreader chains are walked in order**: The CIL tree walker encounters `CILSpreadNode` nodes linked in order. Each `WalkCILSpreadNode` emits an `ExpandDimension` call, and the result variable becomes the source for the next spreader in the chain.

2. **Topological scheduling**: Metrics with spreaders are scheduled after their input metrics. The `ChronoconeComputationController` executes computations in topological order — spreader-dependent metrics naturally run after their inputs.

3. **Spreading matrix emission**: All spreading matrices are emitted in the preamble (reusing `EmitSpreadingMatrix` from Story266).

4. **Error propagation**: Each `ExpandDimension` call is followed by an `if err != nil` check that returns a wrapped error with context.

5. **Variable naming**: Intermediate spreader results use fresh variable names (`spreadTo_Country`, `spreadTo_Product`) to avoid collisions.

### Generated Code Structure (Preamble + Functions)

```
// Preamble
var spreadingMatrix_Country = &tensorops.AggregationMatrix{...}
var spreadingMatrix_Product = &tensorops.AggregationMatrix{...}
var aggregationMatrix_sum_Product = &tensorops.AggregationMatrix{...} // if any aggregators used

// Metric computation functions (ordered by topological dependency)
// 1. Base metric (connector input or simple calculation)
// 2. Spreader-integrated metric (calls base metric, applies spreaders)
// 3. Output metric (uses spreader-integrated metric, applies final ops)

// Main: schedule and execute all computations
```

### Single Unambiguous Path Constraint

At this increment, the analysis phase guarantees exactly one spreader sequence per dimension gap. The synthesizer can assume:
- Each `CILSpreadNode` in the chain has a single, non-nil `SpreadingMatrix`
- The `FromDimensions` of each spreader matches the output dimensions of the previous spreader (or input source)
- The final spreader's `ToDimensions` matches the metric's IO spec output dimensions

### End-to-End Example: with_data/spreaders_integration/

Directory structure:
```
examples/spreaders/integration/
├── spreaders_integration.cpl       # CPL with spreaders + metrics
├── spreaders_integration.io.yaml   # IO spec
├── spreaders_integration_input.xlsx # Input data
├── spreaders_integration.expected.json # Expected output
```

CPL example:
```
connectors
    connector targets
        grain Product
        data { Product="widget": 100; "gadget": 200 }

concepts
    Country = [USA, Canada]
    Product = [widget, gadget]

spreaders
    from [Product] to [Country, Product] = Product * 1/size(Country) * Country

metrics
    metric regional_targets
        inputs
            source_targets (connector targets)
        output region Country, product Product
        =
            source_targets * 1/size(Country)  // spreader applied automatically
```

Expected behavior: The analyzer detects `source_targets` input has `[Product]` but output requires `[Country, Product]`, finds the spreader, inserts a `CILSpreadNode` chain. The synthesizer generates code that calls `ExpandDimension` before the calculation body.

### Test Cases

1. **Single spreader integration**: Metric input `[Product]`, output `[Country, Product]` → generated code has one `ExpandDimension` call before calculation
2. **Multi-step spreader**: Scalar input `[]`, output `[Country, Region, Product]` → generated code has three chained `ExpandDimension` calls
3. **Multiple metrics with spreaders**: Two output metrics, each getting spreader treatment for different inputs → both generated correctly
4. **Mixed spreader and non-spreader inputs**: Metric with one spreader input and one direct (matching-dim) input → only the mismatched input gets spreader code
5. **Error propagation**: If a dimension mismatch produces an `ExpandDimension` error, the error is wrapped with context and returned
6. **Correct `srcDim` and `newDim`**: Verify the generated indices are correct for various dimension combinations

## Task Decomposition (for technical writer)

### Task 1: Implement spreader chain walking in metric computation synthesis — Not Started
   a. Create `src/main/go/synthesis/spreader_integration_synth.go` — modify the CIL tree walker's metric computation handler to properly walk `CILSpreadNode` chains — Not Started
   b. Each `CILSpreadNode` in the chain emits an `ExpandDimension` call; the result of one becomes the source for the next — Not Started
   c. Intermediate variable naming uses `spreadTo_<dimension>` pattern — Not Started
   d. Write unit tests: verify generated code for single spreader chain, verify variable names are correct, verify chain produces correct sequence of `ExpandDimension` calls — Not Started
   e. Verify walker tests pass — Not Started

### Task 2: Implement full metric function generation with spreader integration — Not Started
   a. Generate complete metric functions that include spreader preprocessing before the calculation body — Not Started
   b. Ensure spreading matrix preamble variables are emitted for all spreaders used by the generated metrics — Not Started
   c. Handle case where metric calculation body references the spreader-preprocessed tensor (not the raw input) — Not Started
   d. Write unit tests: full generated function compiles syntactically, correct topological ordering, errors are properly propagated — Not Started
   e. Verify function generation tests pass — Not Started

### Task 3: Create end-to-end spreader integration examples — Not Started
   a. Create `with_data/spreaders_integration/` directory with CPL file (spreaders + metrics with dimensional mismatches), IO spec, input data, expected output — Not Started
   b. Example 1: Single spreader — input `[Product]`, output `[Country, Product]`, equal-proportion distribution — Not Started
   c. Example 2: Multi-step spreader — scalar input, output `[Country, Product]`, two-step spread (`[]→[Country]→[Country,Product]`) — Not Started
   d. Example 3: Spread then aggregate — spread from `[Product]` to `[Country, Product]` then aggregate back via sum → verify original values preserved — Not Started
   e. Verify all examples produce output matching `.expected.json` — Not Started

### Task 4: Verify full pipeline with tests — Not Started [Small]
   a. Run full test suite: `go test ./src/main/go/... -count=1 -timeout 120s` — Not Started
   b. Verify no regressions in existing metric or aggregator tests — Not Started
   c. Verify with_data examples execute and pass — Not Started

## Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4
```

- Task 1 must complete first (chain walking is the core synthesis primitive)
- Task 2 depends on Task 1 (function generation uses chain walking)
- Task 3 depends on all preceding tasks (examples verify full pipeline)
- Task 4 is verification after all code exists

No parallel groups — each task builds on the previous.

## Acceptance Criteria Summary

1. `CILSpreadNode` chains are correctly walked: each node emits one `ExpandDimension` call
2. Intermediate result variables are correctly linked: output of spreader N becomes source of spreader N+1
3. Full metric functions generated include spreader preprocessing before calculation body
4. Spreading matrix preamble includes all matrices used by the generated metrics
5. Topological computation order ensures metrics with spreader inputs run after their sources
6. Error propagation: `ExpandDimension` errors are wrapped with context
7. End-to-end examples produce correct output matching `.expected.json`
8. Spread-then-aggregate round-trip preserves original values (for equal-proportion distribution)
9. Multi-step spreader sequences produce correct intermediate shapes
10. All existing tests continue to pass

## Intent Summary for Technical Writer

This story delivers the synthesis phase for spreader-integrated metric computation. When the Increment 2 analyzer inserts `CILSpreadNode` chains into metric CIL trees, the synthesizer walks those chains and emits the corresponding `tensorops.ExpandDimension` calls. Each spreader in the chain produces an intermediate tensor that becomes the source for the next spreader. The final spreader's output feeds into the metric's calculation body. Spreading matrices are emitted in the preamble (reusing `EmitSpreadingMatrix` from Story266). The `ChronoconeComputationController` handles topological scheduling — metrics with spreader inputs naturally run after their source metrics. Only single unambiguous spreader sequences are supported at this increment; no interleaved aggregator/spreader optimization (Increment 4). The generated code follows the same pattern as aggregator-integrated metric computation (Story261): input retrieval → preprocessing (spreader chain here) → calculation body → result return.

Note: The walker must ensure that when a metric's input has a `CILSpreadNode` chain, the calculation body references the result of the LAST `CILSpreadNode` in the chain, not the raw input reference. This is handled by the walk scope's variable registration — the last spreader registers its result variable, which is what the calculation body picks up.
