# Plan: Story261 — Aggregators Increment 4 (Synthesize)

## Purpose

This plan provides the technical writer with sufficient detail to create Story261: **aggregators-increment4-synthesize** — generate metric computation code with aggregator integration, scheduled topological computation order, and end-to-end aggregator integration examples (Increment 4 of Epic007).

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `aggregators-increment4-synthesize`                                                           |
| description  | Generate metric computation code with aggregator integration, scheduled topological computation order, aggregator integration examples |

## Source EROs

- [ERO007: Aggregators Incremental Implementation](../specifications/design/cpl/eros/ERO007_aggregators-incremental-implementation.md) — Increment 4 sketch
- [ERO007: Aggregators Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO007_aggregators-synthesis-phase.md) — Metric computation with aggregators, scheduled computation order, generated code structure

## Dependencies

- Story260 (aggregators-increment4-analyze) — needs CIL trees with inserted aggregate nodes from `resolveMetricDimensionality`
- Story259 (aggregators-increment3-synthesize) — needs walk methods for all five typed CIL node types
- Story257 (aggregators-increment2-synthesize) — needs matrix emission infrastructure
- Story254 (aggregators-increment1-matrix-mult) — `tensorops.ReduceDimension`

## Files to Create or Modify

| File                                                    | Action  | Purpose                                                                          |
| ------------------------------------------------------- | ------- | -------------------------------------------------------------------------------- |
| `src/main/go/synthesis/synthesize.go`                   | Modify  | Ensure topological scheduling accounts for aggregated intermediate tensors       |
| `src/main/go/cil/cil_transformer_metric.go`             | Modify  | Metric computation code generation handles CIL trees that contain aggregate nodes |
| `src/main/go/cil/cil_transformer_aggregator_test.go`    | Modify  | Extend tests: metric computation with aggregate nodes in CIL tree                |
| `src/main/go/synthesis/synthesize_test.go`             | Modify  | Extend end-to-end synthesis tests with aggregator scenarios                      |
| Examples:                                               |         |                                                                                  |
| `examples/with_data/aggregators_integration/`           | Create  | CPL + IO spec + parquet + expected: metric composed from inputs at different dimensional granularity |

### No New Walk Methods

Increment 4 does NOT add new CIL walk methods. The walk methods added in Stories 257 and 259 already handle all aggregate node types. What changes is the **CIL tree structure**: metric inputs now have aggregate node chains inserted by the analysis phase. The existing walk methods naturally walk these modified trees.

### Generated Code Example

For a metric `regional_profit` with inputs `country_revenue` ([Country, Product]) and `country_cost` ([Country, Product]), output [Product], and aggregator `from [Country, Product] to [Product] = sum(Country)`:

```go
func computeRegionalProfit(controller *ChronoconeComputationController, inputs map[string]*ChronoconeTensor[any]) (*ChronoconeTensor[any], error) {
    // Get input metrics
    countryRevenue := inputs["country_revenue"]
    countryCost := inputs["country_cost"]
    
    // Apply aggregator: from [Country, Product] to [Product] = sum(Country)
    aggregatedRevenue, err := tensorops.ReduceDimension(
        countryRevenue,
        aggregationMatrix_sum_Country,
        0,  // collapsedDim: Country is index 0
        0,  // targetDim: Product is index 0 after removing Country
    )
    if err != nil {
        return nil, fmt.Errorf("failed to aggregate country_revenue: %w", err)
    }
    
    aggregatedCost, err := tensorops.ReduceDimension(
        countryCost,
        aggregationMatrix_sum_Country,
        0,
        0,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to aggregate country_cost: %w", err)
    }
    
    // Both inputs now have dimensions [Product] — compute cell-wise
    result, err := tensorops.TensorSubtractFloat64(aggregatedRevenue, aggregatedCost)
    if err != nil {
        return nil, fmt.Errorf("regional_profit computation failed: %w", err)
    }
    return result, nil
}
```

### Scheduled Computation Order

Aggregated inputs must be computed before the metrics that depend on them. The topological sort ensures this naturally:
1. Connector-loaded metrics are computed first (no dependencies)
2. Metrics whose only inputs are connector data are computed next
3. Metrics that depend on other computed metrics (with potential aggregators) are computed last

The existing topological sort (`synthesize.go`) should handle this correctly since aggregate nodes wrap CIL references — the metric still depends on the input metric conceptually. No change to topological sort logic is needed, but the synthesized main function must verify the schedule is correct.

### with_data Example: aggregators_integration

- **CPL**: concepts (Country, Product), connector (produces [Country, Product] data), aggregator (`from [Country, Product] to [Product] = sum(Country)`), metric (`revenue_by_product` inputs connector and applies implicit aggregation)
- **IO spec**: input parquet connector with [Country, Product], output spec for metric with [Product]
- **parquet data**: known values across Country={USA, Canada} × Product={widget, gadget}
- **expected output**: summed values per product

This example validates the full end-to-end flow: parse → analyze (with dimension mismatch detection and aggregator injection) → CIL transform → synthesize → execute → verify.

## Task Decomposition (for technical writer)

### Task 1: Verify and test metric computation with aggregate nodes — Not Started
   a. Review `cil_transformer_metric.go` to confirm existing metric computation code generation correctly handles CIL trees containing aggregate nodes inserted by the analysis phase — Not Started
   b. Write test: construct a CIL tree with metric input wrapped in `CILAggregateSumNode` → `cil_transformer_metric` generates code with `ReduceDimension` before cell-wise computation — Not Started
   c. Write test: verify generated code structure: metric input loading → aggregate application → cell-wise computation → return — Not Started
   d. If adjustments needed to `cil_transformer_metric.go`, make them (expect minor changes for imported tensorops package reference) — Not Started
   e. Verify tests pass — Not Started

### Task 2: Verify topological scheduling with aggregated dependencies — Not Started [Small]
   a. Review `synthesize.go` scheduling logic — confirm metrics with aggregate-wrapped inputs schedule after their input metrics — Not Started
   b. Write test: two-metric DAG where metric B aggregates metric A's output → verify B is scheduled after A in main function — Not Started
   c. If adjustments needed, make them — Not Started
   d. Verify tests pass — Not Started

### Task 3: Create with_data aggregator integration example — Not Started
   a. Create `examples/with_data/aggregators_integration/` directory — Not Started
   b. Write CPL file: concepts, connector, aggregator declaration, metric with implicit aggregation — Not Started
   c. Write IO spec: input parquet connector with [Country, Product], output metric with [Product] — Not Started
   d. Generate parquet test data with known values — Not Started
   e. Create `.expected.xlsx` with aggregated ground truth — Not Started
   f. Run full pipeline (parse → analyze → CIL → synthesize → execute) and verify output matches expected — Not Started

### Task 4: End-to-end synthesis test with aggregators — Not Started
   a. Write synthesis integration test: construct full `AnalysisDefinitions` with metric that has aggregate-wrapped input → synthesize → verify generated Go code compiles and produces correct results — Not Started
   b. Test error propagation: verify that `ReduceDimension` errors in generated code are properly handled with descriptive messages — Not Started
   c. Verify all end-to-end tests pass — Not Started

## Execution Order

```
Task 1 → Task 2 → [Task 3 ‖ Task 4]
```

- Task 1 must complete first (metric code gen must be verified before scheduling can be tested)
- Task 2 depends on Task 1 (scheduling uses metric code generation patterns)
- Tasks 3 and 4 are parallel: Task 3 is a concrete example; Task 4 is a programmatic integration test

## Parallel Execution

- **Group A** (Tasks 3, 4): Task 3 creates example files in `examples/with_data/`; Task 4 writes Go test code in `synthesize_test.go`. Fully disjoint files and directories.

## Reintegration

**Group A (Tasks 3, 4):**
- Merge order: any order (fully disjoint: examples directory vs test files)
- Integration tests: verify example can be run end-to-end via `go test ./src/main/go/synthesis/...`
- Watch for conflicts in: none expected (disjoint file paths)

## Acceptance Criteria Summary

1. Metric computation code generation handles CIL trees containing aggregate node chains: emits `ReduceDimension` calls before cell-wise operations
2. Generated metric functions properly load inputs, apply aggregators, compute cell-wise, and return results
3. Error propagation from `ReduceDimension` is properly wired with descriptive messages
4. Topological scheduling ensures aggregated metrics are computed before their consumers
5. `with_data/aggregators_integration/` example: CPL + IO spec + parquet → synthesized executable produces correct aggregated results matching expected output
6. End-to-end synthesis test with aggregate nodes in CIL tree produces compilable code with correct results
7. All tests pass with `go test ./src/main/go/cil/... ./src/main/go/synthesis/...`

## Intent Summary for Technical Writer

This story completes Increment 4 by connecting the analysis-phase aggregator injection (Story260) with the synthesis-phase code generation (Stories 257, 259). No new CIL walk methods are needed — the existing walk methods for aggregate node types handle the modified CIL trees automatically. The key change is that metric CIL trees now contain aggregate node chains inserted by the analysis phase between input references and calculation bodies. The metric computation code generator (`cil_transformer_metric.go`) naturally walks these trees and emits `ReduceDimension` calls as part of metric computation. The topological schedule ensures aggregated intermediate tensors are computed before their consumers. This increment delivers the first fully automatic aggregator integration: the user declares aggregators, and the framework applies them when dimension mismatches are detected.
