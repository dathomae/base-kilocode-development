# Plan: Story269 — Spreaders Increment 3 (Analyze)

## Purpose

This plan provides the technical writer with sufficient detail to create Story269: **spreaders-increment3-analyze** — implement metric reference resolution in spreader expressions, implicit aggregator determination via dimensional unit analysis, required aggregator validation, and cell unit compatibility checks (Increment 3 of Epic008). No parsing changes required.

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `spreaders-increment3-analyze`                                                               |
| description  | Implement metric reference resolution in spreader expressions, implicit aggregator determination via dimensional unit analysis, required aggregator validation, cell unit compatibility checks |

## Source EROs

- [ERO008: Spreaders Incremental Implementation](../specifications/design/cpl/eros/ERO008_spreaders-incremental-implementation.md) — Increment 3 sketch
- [ERO008: Spreaders Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO008_spreaders-analysis-phase.md) — Increment 3 data structures and algorithms

## Dependencies

- Story265 (spreaders-increment1-analyze) — `SpreaderInfo`, `SpreaderFunctionRegistry`, postfix expression evaluator and `CILSpreadNode` must exist
- Story256 (aggregators-increment2-analyze) — Aggregators map must exist for implicit aggregator resolution
- Story260 (aggregators-increment4-analyze) — Aggregator selection/integration must be complete; implicit aggregator resolution queries the Aggregators map
- Story267 (spreaders-increment2-analyze) — Spreader integration into metric DAG must exist (spreader expressions with metric references are used in the same DAG context)
- Epic005 (metrics) — Metrics map must exist for resolution; metric CIL trees must be available for `CILMetricRefNode` generation

## Files to Create or Modify

| File                                                    | Action | Purpose                                                                          |
| ------------------------------------------------------- | ------ | -------------------------------------------------------------------------------- |
| `src/main/go/analysis/spreader_metric_resolver.go`      | Create | `resolveSpreaderIdentifier` (metrics→concepts→lookups), `resolveSpreaderMetricReference`, `findAggregatorByDimUnitEffect` |
| `src/main/go/analysis/spreader_analyzer.go`             | Modify | Update `evaluateSpreaderExpression` to call new resolution logic for identifier operands; update `CILSpreadNode` with `ReferencedMetrics` field |
| `src/main/go/analysis/cil_nodes.go`                     | Modify | Update `CILSpreadNode` to include `ReferencedMetrics []CILNode` field              |
| `src/main/go/analysis/spreader_metric_resolver_test.go` | Create | Unit tests: metric reference resolution, implicit aggregator determination, cell unit compatibility |

### Identifier Resolution — Three-Tier Priority

In Increment 1, identifiers in spreader expressions were resolved to concepts only. In Increment 3, the resolution hierarchy is:

```go
func (a *CPLAnalyzer) resolveSpreaderIdentifier(name string) (*DimensionalUnitExpression, CILNode, error) {
    // 1. Check if it's a metric
    if metric, exists := a.metrics[name]; exists {
        dimUnits := dimensionalUnitsFromConceptSet(metric.Dimensions)
        return dimUnits, &CILMetricRefNode{
            MetricName: name,
            CILTree:    metric.CILTree,
        }, nil
    }
    // 2. Check if it's a concept
    if concept, exists := a.concepts[name]; exists {
        return &DimensionalUnitExpression{Numerator: []string{name}}, nil, nil
    }
    // 3. Check if it's a lookup (future integration with Epic006)
    if lookup, exists := a.lookups[name]; exists {
        return dimensionalUnitsFromConceptSet(lookup.KeyDimensions), nil, nil
    }
    return nil, nil, fmt.Errorf("unresolved identifier %q in spreader expression", name)
}
```

The function now returns both a `DimensionalUnitExpression` (for dim unit validation) and an optional `CILNode` (for code generation — the metric reference's CIL subtree).

### Implicit Aggregator Determination

When a spreader expression references a metric whose dimensional form doesn't match what the expression needs, the analyzer determines whether an implicit aggregator can bridge the gap:

```go
func (a *CPLAnalyzer) resolveSpreaderMetricReference(
    metricName string,
    requiredDimUnits *DimensionalUnitExpression,
    spreaderCtx *SpreaderInfo,
) (CILNode, error) {

    metricDef, exists := a.metrics[metricName]
    if !exists {
        return nil, fmt.Errorf("metric %q not found", metricName)
    }

    rawDimUnits := dimensionalUnitsFromConceptSet(metricDef.Dimensions)

    // If raw dim units satisfy the requirement, use metric directly
    if rawDimUnits.Equals(requiredDimUnits) {
        return &CILMetricRefNode{
            MetricName: metricName,
            CILTree:    metricDef.CILTree,
        }, nil
    }

    // Raw dim units don't match — find an aggregator to bridge the gap
    // Formula: raw * factor = required
    // Therefore: factor = required / raw
    requiredFactor := requiredDimUnits.Divide(rawDimUnits)

    // Search aggregator map for a transformation that produces this factor
    aggKey := findAggregatorByDimUnitEffect(a.aggregators, requiredFactor)
    if aggKey == "" {
        // Also try inverse: check if raw / aggregator = required
        // (some aggregators reduce dimensional units)
        aggKey = findAggregatorByInverseEffect(a.aggregators, rawDimUnits, requiredDimUnits)
        if aggKey == "" {
            return nil, fmt.Errorf(
                "spreader references metric %q but no aggregator exists that transforms %s to %s",
                metricName, rawDimUnits.String(), requiredDimUnits.String(),
            )
        }
    }

    aggInfo := a.aggregators[aggKey]

    // Build implicit aggregator node wrapping the metric reference
    return &CILAggregateNode{
        SourceCIL: &CILMetricRefNode{
            MetricName: metricName,
            CILTree:    metricDef.CILTree,
        },
        FromDimensions:    getConceptNames(aggInfo.FromConcepts),
        ToDimension:       aggInfo.ToConcept,
        AggregationMatrix: aggInfo.Matrix,
        CellOperation:     aggInfo.CellOperation,
    }, nil
}
```

### The Dim Unit Effect Search

```go
// findAggregatorByDimUnitEffect searches for an aggregator whose dim unit effect
// (input_units / aggregator_collapsed_dim) equals the required factor.
func findAggregatorByDimUnitEffect(
    aggregators map[string]*AggregatorInfo,
    factor *DimensionalUnitExpression,
) string {
    for key, info := range aggregators {
        // The aggregator's effect: cancels its from dimension
        // If the factor matches canceling that dim, use this aggregator
        aggEffect := computeAggregatorDimEffect(info)
        if aggEffect.Equals(factor) {
            return key
        }
    }
    return ""
}
```

### Expression Evaluation Updates

The `evaluateSpreaderExpression` function in `spreader_analyzer.go` is updated to handle the `"identifier"` case differently:

```go
case "identifier":
    dimUnits, refNode, err := a.resolveSpreaderIdentifier(operand.Name)
    if err != nil {
        return nil, err
    }
    // If it resolved to a metric, record the reference for CIL node population
    if refNode != nil {
        a.currentSpreaderMetricRefs = append(a.currentSpreaderMetricRefs, refNode)
    }
    // If the metric's raw dim units don't match what's needed, attempt implicit aggregation
    // ... (handled in a second pass after initial dim unit analysis)
    stack = append(stack, dimUnits)
```

### CILSpreadNode Updates (Increment 3)

The `CILSpreadNode` gains a `ReferencedMetrics` field:

```go
type CILSpreadNode struct {
    SourceCIL               CILNode
    FromDimensions          []string
    ToDimensions            []string
    NewDimension            string
    SpreadingMatrix         *AggregationMatrix
    DistributionExpression  CILNode
    ReferencedMetrics       []CILNode       // NEW: CIL subtrees for referenced metrics (with implicit aggregators)
    DimensionalUnitEffect   string
    SourceLocation          SourceLocation
}
```

Each referenced metric (possibly wrapped in `CILAggregateNode` for implicit resolution) is stored in this slice. The synthesizer (Increment 3) walks these subtrees to generate code that computes the reference before the spread.

### Cell Unit Compatibility

After metric reference resolution, the analyzer verifies that the cell units of the referenced metric (after any implicit aggregation) are compatible with the spreader's cell unit context:

```go
func validateCellUnitCompatibility(spreaderCellUnits *Unit, refCellUnits *Unit) error {
    if !spreaderCellUnits.Equals(refCellUnits) {
        return fmt.Errorf(
            "cell unit mismatch: spreader expects %s but metric %s provides %s after aggregation",
            spreaderCellUnits, metricName, refCellUnits,
        )
    }
    return nil
}
```

### Validation During Analysis (Increment 3 — Additions)

1. **Metric Reference Existence**: Referenced metric must exist in the Metrics map
2. **Required Aggregator Existence**: If a metric needs an implicit aggregator, it must exist in the Aggregators map
3. **Cell Unit Compatibility**: The spreader expression must produce result cell units compatible with the output metric
4. **Self-Reference Detection**: A spreader cannot reference the metric that the spreader is being applied to (prevents infinite recursion)
5. **Circular Reference Detection**: Spreader expression metric references must not create circular dependencies in the metric DAG

### SpreadingMatrix Construction Changes (Increment 3)

For spreaders that reference metrics (Increment 3), the spreading matrix cannot be fully pre-built at analysis time — weights come from metric data at runtime. The `SpreadingMatrix` is a **template** that indicates which metric provides the weights:

```go
// For a spreader expression "Product / customers" where customers is a [Country] metric:
// - The spreading matrix is marked as "dynamic" with the reference to the customers metric
// - At runtime, the customers metric is computed first, then its values are used as weights
```

### Limitations (Increment 3)

- Only single metric references (not expressions combining multiple metrics)
- Implicit aggregators must be found in the Aggregators map; no cross-epic aggregation discovery
- Cell unit checking relies on `float64` type (detailed type selection deferred)
- Dynamic spreading matrix from metric data must be re-computed if concept values change

### Test Cases

1. **Direct metric reference**: Spreader `from [Product] to [Country, Product] = Product / customers` where `customers` is a `[Country]` metric — resolves `customers` as metric, records in `ReferencedMetrics`
2. **Implicit aggregator**: Spreader references `[Country]` metric but needs scalar — finds `from [Country] to [] = 1/sum(Country)` aggregator, wraps metric ref in `CILAggregateNode`
3. **Metric not found**: Spreader references non-existent name → error
4. **Aggregator not found**: Spreader references metric needing implicit aggregation but no matching aggregator exists → error
5. **Self-reference prevention**: Metric's own spreader expression references the metric itself → error
6. **Concept resolution still works**: Identifier that matches a concept name still resolves to concept (not metric)
7. **Cell unit compatibility**: Metric with `float64` cell type is compatible with spreader's `float64` cell type
8. **CILSpreadNode ReferencedMetrics**: After resolution, `ReferencedMetrics` slice contains correct CIL nodes

## Task Decomposition (for technical writer)

### Task 1: Implement three-tier identifier resolution — Not Started
   a. Create `src/main/go/analysis/spreader_metric_resolver.go` — implement `resolveSpreaderIdentifier` with priority order: metrics → concepts → lookups — Not Started
   b. Each resolution tier returns both `DimensionalUnitExpression` (for dim unit validation) and optional `CILNode` (for code generation) — Not Started
   c. Write unit tests: metric name resolves to metric, concept name resolves to concept even if a metric with same name exists (concept takes priority for concepts), unknown identifier returns error — Not Started
   d. Verify resolution tests pass — Not Started

### Task 2: Implement implicit aggregator determination — Not Started
   a. Implement `resolveSpreaderMetricReference` — checks if metric's raw dim units match required, if not searches for aggregator — Not Started
   b. Implement `findAggregatorByDimUnitEffect` — searches Aggregators map for transformation matching the required dim unit factor — Not Started
   c. Implement `findAggregatorByInverseEffect` — fallback search for aggregators whose inverse transformation matches — Not Started
   d. Build `CILAggregateNode` wrapping `CILMetricRefNode` when implicit aggregation needed — Not Started
   e. Write unit tests: direct match (no aggregator needed), implicit aggregator found, implicit aggregator not found, multiple candidate aggregators (first match wins), verify CIL node structure after implicit aggregation — Not Started
   f. Verify implicit aggregation tests pass — Not Started

### Task 3: Update expression evaluation and CILSpreadNode — Not Started
   a. Update `evaluateSpreaderExpression` in `spreader_analyzer.go` to call `resolveSpreaderIdentifier` instead of concept-only resolution for identifiers — Not Started
   b. Record resolved metric references (with implicit aggregators) in the spreader's context for `CILSpreadNode.ReferencedMetrics` population — Not Started
   c. Update `CILSpreadNode` struct in `cil_nodes.go` to include `ReferencedMetrics []CILNode` — Not Started
   d. Implement self-reference detection: spreader expression cannot reference the output metric — Not Started
   e. Implement circular reference detection: spreader's metric references must not create cycles in the dependency graph — Not Started
   f. Write unit tests: expression evaluation with metric reference produces correct dim units, `ReferencedMetrics` populated correctly, self-reference and circular reference produce errors — Not Started
   g. Verify expression evaluation tests pass — Not Started

### Task 4: Cell unit compatibility and end-to-end analysis — Not Started
   a. Implement cell unit compatibility validation between referenced metric cell type and spreader cell type — Not Started
   b. Wire complete Increment 3 analysis into the main analysis pipeline — Not Started
   c. Create `parse_and_analyze_only/` examples for Increment 3: valid metric references, implicit aggregation cases, self-reference errors, circular reference errors — Not Started
   d. Verify full parse→analyze pipeline produces correct results and errors — Not Started

## Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4
```

- Task 1 must complete first (resolution hierarchy is foundational)
- Task 2 depends on Task 1 (implicit aggregation uses resolution)
- Task 3 depends on Tasks 1 and 2 (expression evaluation uses both resolution and implicit aggregation)
- Task 4 depends on all preceding tasks (validation + end-to-end verification)

No parallel groups — each task builds on the previous.

## Acceptance Criteria Summary

1. Identifier resolution follows priority: metrics → concepts → lookups
2. Metric references in spreader expressions resolve to `CILMetricRefNode` with the metric's CIL tree
3. When a metric's dimensional form doesn't match requirements, implicit aggregator resolution finds the right aggregator
4. `CILAggregateNode` correctly wraps `CILMetricRefNode` when implicit aggregation is applied
5. Missing aggregator for required transformation produces error
6. Self-referencing spreader expressions produce error
7. Circular metric dependencies via spreader expressions produce error
8. `CILSpreadNode.ReferencedMetrics` populated correctly with resolved metric subtrees
9. Cell unit compatibility validated between referenced metric and spreader context
10. Concept resolution in spreader expressions still works (non-metric identifiers resolve to concepts)
11. All parse_and_analyze_only examples pass
12. All tests pass with `go test ./src/main/go/analysis/... -count=1 -timeout 60s`

## Intent Summary for Technical Writer

This story introduces the most complex analysis-phase feature in the spreaders epic: metric reference resolution with implicit aggregator determination. When a spreader expression references a metric by name (e.g., `Product / customers` for weighted distribution), the analyzer must resolve the identifier, determine the metric's dimensional form, and if it doesn't match what the expression needs, find an implicit aggregator to bridge the gap. The resolution priority is metrics → concepts → lookups (consistent with the language semantics). The dimensional algebra is: `raw_dim_units * aggregator_factor = required_dim_units` — the analyzer solves for the aggregator factor and searches the Aggregators map. When found, a `CILAggregateNode` wrapping a `CILMetricRefNode` is stored in `CILSpreadNode.ReferencedMetrics`. The synthesizer (Story270) walks these subtrees to emit code that computes referenced metrics (with implicit aggregators) before the spread operation. Self-referencing and circular dependencies are detected and rejected.

Note: This story requires the Aggregators map (from Story256) to be populated and the aggregator integration (Story260) to be complete, as implicit resolution queries the map for matching aggregators. This is the most coordination-dependent story in the spreaders epic.
