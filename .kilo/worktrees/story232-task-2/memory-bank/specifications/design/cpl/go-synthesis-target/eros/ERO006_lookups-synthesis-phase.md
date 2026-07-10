# ERO006: Lookups Synthesis Phase

## Title

Lookups Synthesis Phase — Code Generation and Runtime Integration

## Purpose/Goal

Define the synthesis phase implementation for lookup definitions, including CIL tree walking for `CILLookupNode` and `CILLookupRefNode` nodes, Go code generation for lookup data loading and `MetricComputation` registration, and runtime lookup access via `controller.GetMetricComputation`. Document how synthesis evolves across each increment of the lookups incremental implementation as defined in [ERO006 — Lookups Incremental Implementation](../../eros/ERO006_lookups-incremental-implementation.md).

## Decision Status

Draft

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic006_cpl-lookups-implementation.md](../../../../../epics/Epic006_cpl-lookups-implementation.md)

## Context

### Situation

The synthesis phase is responsible for walking CIL trees and generating executable Go code. For lookups, this involves:

1. Walking `CILLookupNode` leaf trees to emit code that loads lookup data and registers `MetricComputation` objects
2. Walking `CILLookupRefNode` nodes within metric CIL trees to emit code that accesses lookup data at runtime
3. Integrating lookup loading into the generated executable's initialization sequence

Lookups use the same `MetricComputation` mechanism as connectors: each lookup is registered as a `MetricComputation` on the `ChronoconeComputationController` before the postfix evaluator runs. This enables metric calculations to access lookup data via `controller.GetMetricComputation(lookupName)` during evaluation.

The CIL tree walker infrastructure from existing declaration types (connectors, metrics) is reused. The `CILLookupNode` is structurally analogous to `CILConnectorReferenceNode` — a single-node leaf tree — and is walked using the same traversal pattern.

### Constraints

- Must reuse the existing CIL tree walking infrastructure
- Lookup data loading must occur before metric calculations that reference lookups
- Scheduled metric initiation order must account for lookup → metric dependencies
- The `MetricComputation` interface must support both connector-originated and lookup-originated tensors
- Lookup IO Specification parsing determines the data source (parquet file, CSV, database)
- Types are initially represented as `float64` (detailed type selection deferred to [type-selection.md](../../type-selection.md))

### Requirements

- Walk `CILLookupNode` trees to generate Go code that loads lookup data
- Register each lookup as a `MetricComputation` on the `ChronoconeComputationController`
- Generate code that calls `controller.GetMetricComputation(lookupName)` for lookup references in metric CIL trees
- Support timeless lookups (no grain) and temporal lookups (with grain measures)
- Load lookup data from the source specified in the Lookup IO Specification
- Handle lookups referenced in aggregator and spreader expressions (Increments 4 and 5)

### Assumptions

- CIL tree walking infrastructure from connectors and metrics is available
- The `ChronoconeComputationController` and `MetricComputation` interface from connector synthesis are reusable
- Parquet file reading utilities from connector synthesis can be reused for lookup data loading
- The generated executable follows the same structure (Go CLI with embedded CPL, IO spec, and data)
- The Lookup IO Specification format is defined (though the schema is documented separately)

## Data Structures and Algorithms by Increment

### Increment 1: Basic Lookup Declarations Without Grain

**Overview**

Increment 1 implements synthesis for timeless lookups (no grain clause). The CIL tree walker handles `CILLookupNode` leaf nodes, emitting Go code that reads lookup data from a parquet file (or other source) and registers it as a `MetricComputation`. Lookups are registered before any metric calculations run.

**CIL Tree Walker — CILLookupNode Handler**

File: `src/chronocone_planning_language/src/main/go/synthesis/cil_tree_walker.go` (add to existing walker)

```go
// WalkCILLookupNode handles a CILLookupNode in the CIL tree.
// It generates Go code to load lookup data and register a MetricComputation.
func (w *CILTreeWalker) WalkCILLookupNode(node *CILLookupNode, scope *WalkScope) error {
    // Register the lookup as a data source that needs loading
    scope.addLookupDataSource(node.LookupName, node)

    // Store lookup metadata for the preamble generator
    scope.lookupDefinitions[node.LookupName] = node

    // No CIL subtree to walk — CILLookupNode is a leaf node
    return nil
}
```

**Generated Preamble Code — Lookup Loading (Increment 1)**

The synthesis preamble generator emits Go code that loads lookup data before any metric calculations:

```go
// generateLookupLoading emits code that loads all lookup data
// and registers MetricComputation objects.
func (g *PreambleGenerator) generateLookupLoading(buf io.Writer, lookups map[string]*CILLookupNode) error {
    if len(lookups) == 0 {
        return nil // No lookups to load
    }

    // Generate a function that loads all lookup data
    tmpl := `
// loadLookupData loads all lookup data from parquet files
// and registers each as a MetricComputation on the controller.
func loadLookupData(controller *chronocone_compute.ChronoconeComputationController) error {
{{range .Lookups}}
    // Load {{.Name}} lookup data
    data, err := readParquetFile("{{.FileName}}")
    if err != nil {
        return fmt.Errorf("failed to load lookup '{{.Name}}': %w", err)
    }

    // Build tensor with indexed dimensions
    tensor := chronocone_compute.NewChronoconeTensor[{{.CellType}}](data)
    tensor.SetDimensions([]string{
        {{range .Dimensions}}"{{.}}",{{end}}
    })

    // Register as MetricComputation
    computation := chronocone_compute.NewMetricComputationFromTensor("{{.Name}}", tensor)
    controller.AddMetricComputation(computation)
{{end}}
    return nil
}
`
    return renderTemplate(buf, tmpl, lookups)
}
```

**Parquet Reading for Lookup Data (Increment 1)**

The generated code reuses the existing parquet reading infrastructure from connector synthesis:

```go
// readParquetFile reads a parquet file and returns the data as a [][]interface{}.
// This function is reused from connector synthesis (already generated).
func readParquetFile(filename string) ([][]interface{}, error) {
    // ... existing implementation from connector synthesis ...
}
```

**Metric Calculation with Lookup References (Increment 1)**

When a `CILLookupRefNode` is encountered in a metric's CIL tree, the walker emits a call to `controller.GetMetricComputation`:

```go
// WalkCILLookupRefNode handles a CILLookupRefNode in a metric's CIL tree.
func (w *CILTreeWalker) WalkCILLookupRefNode(node *CILLookupRefNode, scope *WalkScope) error {
    // Emit code that retrieves the lookup value at runtime
    scope.emitCode(fmt.Sprintf(
        "controller.GetMetricComputation(\"%s\").GetCell(rowIndex, colIndex)",
        node.LookupName,
    ))
    // Note: Actual cell indexing depends on the metric's dimension context.
    // For Increment 1, a simple indexed access is emitted.
    return nil
}
```

However, `CILLookupRefNode` is not yet exercised in Increment 1 because lookup integration with metric calculations is introduced in Increment 3. In Increment 1, the only synthesis task is loading and registering lookup data.

**ChronoconeComputationController — MetricComputation Registration (Increment 1)**

The generated code registers lookups on the same controller used for connectors:

```go
// In the generated main function:
func main() {
    controller := chronocone_compute.NewChronoconeComputationController()

    // Load connector data
    // ... (existing connector loading code) ...

    // Load lookup data
    if err := loadLookupData(controller); err != nil {
        log.Fatalf("Failed to load lookup data: %v", err)
    }

    // Run calculations
    // ... (existing metric computation code) ...
}
```

**Tensor Indexing for Lookups (Increment 1)**

Lookup tensors are indexed by their key dimensions. For a timeless lookup keyed by `[Product]`, the tensor is 1-dimensional with one position per product value. For a lookup keyed by `[Currency, Currency]`, the tensor is 2-dimensional (source currency × target currency).

The `ChronoconeTensor` type must support lookup-style indexing — retrieving a value for a specific key combination. If the existing tensor infrastructure uses sequential indices (0, 1, 2, ...), a mapping from concept values to indices is required:

```go
// LookupTensor wraps a ChronoconeTensor with dimension value → index mappings
// to support key-based access (e.g., Get("USA", "EURO")).
type LookupTensor struct {
    Tensor     *ChronoconeTensor[float64]
    ValueMaps  []map[string]int  // One map per dimension: concept value → tensor index
}
```

This lookup tensor wrapper is generated during synthesis if the existing tensor infrastructure does not already support named dimension indexing. Alternatively, the lookup data can be loaded into the existing tensor format and the dimension value → index mapping can be pre-registered on the `MetricComputation`.

**Validation During Synthesis (Increment 1)**

- Each `CILLookupNode` must have a corresponding entry in the Lookup IO Specification (mapping lookup name to file path)
- Lookup data file must exist and be a valid parquet file (runtime check)
- Parquet file columns must match the lookup's declared dimensions and celltype (runtime check)

**Limitations (Increment 1)**

- No temporal lookups (grain not yet supported — added in Increment 2)
- No lookup access from metric calculations (added in Increment 3)
- Lookups loaded from the Lookup IO Specification but the IO spec schema is documented separately
- No optimization for repeated lookup key accesses (caching deferred)

### Increment 2: Lookups with Grain

**Overview**

Increment 2 adds temporal lookup support. When `CILLookupNode.HasGrain` is true, the generated code constructs a tensor indexed by both concept dimensions and time. The grain dimensions are part of the lookup key alongside the concept dimensions. Multi-grain lookups (multiple grain measures like `Week, Month`) require resolution at runtime: when a metric requests a lookup value at a specific time point, the framework selects the appropriate row by matching the time coordinate against the grain dimensions.

**CILLookupNode (Increment 2 — fields used)**

The `HasGrain` and `Grain` fields added in Increment 2 of the analysis phase are used by synthesis:

```go
// In the CIL tree walker, grain information determines tensor structure:
func (w *CILTreeWalker) WalkCILLookupNode(node *CILLookupNode, scope *WalkScope) error {
    scope.addLookupDataSource(node.LookupName, node)
    scope.lookupDefinitions[node.LookupName] = node

    if node.HasGrain {
        // Mark this lookup as temporal so the preamble generator
        // includes time in the tensor indexing structure
        scope.temporalLookups[node.LookupName] = node
    }

    return nil
}
```

**Generated Preamble Code — Temporal Lookup Loading (Increment 2)**

```go
// generateLookupLoading extended for temporal lookups:
func (g *PreambleGenerator) generateLookupLoading(buf io.Writer, lookups map[string]*CILLookupNode, temporalLookups map[string]*CILLookupNode) error {
    // ...

    tmpl := `
// loadLookupData loads all lookup data and registers MetricComputation objects.
func loadLookupData(controller *chronocone_compute.ChronoconeComputationController) error {
{{range .Lookups}}
    // Load {{.Name}} lookup data
    data, err := readParquetFile("{{.FileName}}")
    if err != nil {
        return fmt.Errorf("failed to load lookup '{{.Name}}': %w", err)
    }

    // Build tensor with indexed dimensions
    tensor := chronocone_compute.NewChronoconeTensor[{{.CellType}}](data)

    // Set dimension names including grain dimensions
    tensor.SetDimensions([]string{
        {{range .Dimensions}}"{{.}}",{{end}}
        {{if .HasGrain}}{{range .Grain}}"{{.}}",{{end}}{{end}}
    })

    // Register as MetricComputation
    computation := chronocone_compute.NewMetricComputationFromTensor("{{.Name}}", tensor)
    {{if .HasGrain}}
    // Mark as temporal: grain dimensions are used for time-based key matching
    computation.SetGrainDimensions([]string{
        {{range .Grain}}"{{.}}",{{end}}
    })
    {{end}}
    controller.AddMetricComputation(computation)
{{end}}
    return nil
}
`
}
```

**Multi-Grain Resolution at Runtime (Increment 2)**

When a lookup has multiple grain measures (e.g., `grain Week, Month`), the runtime must select the appropriate row based on the metric cell's time coordinate. The resolution algorithm:

```go
// resolveGrainIndex selects the appropriate row index in a temporal lookup
// based on the requesting cell's time coordinate.
func resolveGrainIndex(lookup *MetricComputation, cellTime chronocone_compute.Instant) (int, error) {
    grains := lookup.GrainDimensions()
    for i, grainName := range grains {
        // Check if the cell time matches this grain's interval
        // (e.g., if grain is "Week" and cell time falls within a specific week)
        if matchesGrain(cellTime, grainName) {
            return i, nil
        }
    }
    return 0, fmt.Errorf("time coordinate does not match any grain dimension")
}
```

The specific matching logic depends on the Interval type definitions for each grain measure.

**Validation During Synthesis (Increment 2 — additions)**

- Temporal lookup parquet files must include columns for grain dimension values
- Grain dimension columns must be parseable as time intervals

### Increment 3: Lookup Integration with Metric Calculations

**Overview**

Increment 3 enables metric calculations to access lookup data at runtime. When a metric's CIL tree contains a `CILLookupRefNode`, the walker emits code that calls `controller.GetMetricComputation(lookupName).GetCell(...)` to retrieve the lookup value for the current metric cell.

**CILLookupRefNode Walking (Increment 3)**

The CIL tree walker handles `CILLookupRefNode` nodes within metric CIL trees:

```go
// WalkCILLookupRefNode handles a CILLookupRefNode in a metric's CIL tree.
// It emits Go code that accesses lookup data via the MetricComputation controller.
func (w *CILTreeWalker) WalkCILLookupRefNode(node *CILLookupRefNode, scope *WalkScope) error {
    // A lookup reference in a metric's calculate expression becomes
    // a variable access in the generated Go code.
    //
    // The variable is pre-computed before the main calculation loop
    // and holds the lookup value for the current cell's dimensional context.

    varName := scope.lookupVarName(node.LookupName)

    // Emit code to access the lookup tensor at the current cell's key coordinates
    scope.emitCodeIndented(scope.indentLevel, fmt.Sprintf(
        `%s := controller.GetMetricComputation("%s").GetCell(%s)`,
        varName,
        node.LookupName,
        scope.cellKeyCoordinates(node.Dimensions),
    ))

    // Register the variable so subsequent references use the pre-computed value
    scope.registerVariable(node.LookupName, varName)

    // The lookup variable name is pushed onto the expression stack
    // so that expression evaluation uses it as an operand
    scope.pushOperand(varName)

    return nil
}
```

**Cell Key Coordinate Computation (Increment 3)**

The key coordinate computation generates Go code that extracts the current cell's dimension values for the lookup's key dimensions:

```go
// cellKeyCoordinates generates Go code that constructs the coordinates
// for indexing into a lookup tensor based on the current metric cell.
func (scope *WalkScope) cellKeyCoordinates(lookupDims []string) string {
    // For each lookup dimension, generate code that gets the current cell's
    // value for that dimension from the loop iteration variables.
    var parts []string
    for _, dim := range lookupDims {
        loopVar := scope.dimensionLoopVariable(dim)
        parts = append(parts, loopVar)
    }
    return strings.Join(parts, ", ")
}
```

**Unit Compatibility Verification (Increment 3)**

The synthesis phase verifies that the lookup's celltype units are compatible with the metric calculation that uses it. This is a compile-time check at synthesis time (the analysis phase already confirmed dimensional compatibility):

```go
// verifyLookupCelltypeCompatibility checks that a lookup's celltype
// can be used in the metric calculation that references it.
func (w *CILTreeWalker) verifyLookupCelltypeCompatibility(
    lookupNode *CILLookupNode,
    metricNode *CILMetricNode,
) error {
    // The lookup's celltype measure has a unit (e.g., Currency/Currency).
    // The metric calculation must be dimensionally valid when this unit
    // is substituted for the lookup reference.
    // Full unit algebra verification is performed during analysis;
    // synthesis performs a final structural check.
    return nil // Placeholder — full implementation in analysis phase
}
```

**Generated Metric Calculation with Lookup (Increment 3)**

Example generated Go code for a metric that references a lookup:

```go
// Generated code for metric "converted_revenue":
func computeConvertedRevenue(
    controller *chronocone_compute.ChronoconeComputationController,
) *chronocone_compute.ChronoconeTensor[float64] {

    revenue := controller.GetMetricComputation("revenue").Tensor().(*chronocone_compute.ChronoconeTensor[float64])
    currencyConversion := controller.GetMetricComputation("currency_conversion").Tensor().(*chronocone_compute.ChronoconeTensor[float64])

    result := chronocone_compute.NewChronoconeTensor[float64](revenue.Shape())

    // Iterate over all cells in the revenue tensor
    for i, cell := range revenue.Cells() {
        // Get the conversion rate for this cell's dimensional context
        sourceCurrency := revenue.DimensionValue(i, "source_currency")
        targetCurrency := revenue.DimensionValue(i, "target_currency")
        conversionRate := currencyConversion.GetCell(sourceCurrency, targetCurrency)

        // Apply conversion
        result.SetCell(i, cell.Value * conversionRate)
    }

    return result
}
```

**Validation During Synthesis (Increment 3 — additions)**

- Lookup tensor must be registered on the controller before the metric calculation that references it
- Lookup celltype units must be compatible with the metric calculation (structural check)
- Scheduled metric initiation order must respect lookup → metric dependencies

### Increment 4: Lookups in Aggregator Expressions

**Overview**

No synthesis changes are required in Increment 4 at the lookup synthesis level. When an aggregator expression references a lookup, the aggregator CIL tree walker (defined in ERO007's synthesis ERO) emits calls to `controller.GetMetricComputation(lookupName)` as part of the aggregation code.

The lookup synthesis phase ensures that lookup data is loaded and registered before any aggregator that references it runs.

**Lookup Deregistration Safety (Increment 4)**

Lookups referenced in aggregator expressions must not be deregistered (garbage collected) until all metrics that use the aggregator have completed. The scheduled execution order ensures this.

### Increment 5: Lookups in Spreader Expressions

**Overview**

No synthesis changes are required in Increment 5 at the lookup synthesis level. When a spreader expression references a lookup, the spreader CIL tree walker (defined in ERO008's synthesis ERO) emits calls to `controller.GetMetricComputation(lookupName)` as part of the spreading code.

The lookup synthesis phase ensures that lookup data is loaded and registered before any spreader that references it runs.

## Generated Code Structure (All Increments)

The final generated Go executable follows this structure:

```
main()
├── Parse CPL source
├── Parse IO specifications (including Lookup IO Spec)
├── Create ChronoconeComputationController
├── Load connector data → register MetricComputations
├── Load lookup data → register MetricComputations   ← Lookup synthesis
├── Schedule metric computation order (topological)
├── Compute metrics
│   ├── GetMetricComputation("connector_A")
│   ├── GetMetricComputation("lookup_X")             ← Lookup access
│   └── Calculate result
├── Apply aggregators (if any, ERO007)
│   └── GetMetricComputation("lookup_X")             ← Lookup access in aggregators
├── Apply spreaders (if any, ERO008)
│   └── GetMetricComputation("lookup_X")             ← Lookup access in spreaders
└── Output results
```

## Code Generation Strategy

| Aspect | Strategy |
|--------|----------|
| Lookup data loading | Reuses parquet reading infrastructure from connector synthesis |
| MetricComputation registration | Same controller as connectors: `controller.AddMetricComputation(...)` |
| Lookup access in metrics | `controller.GetMetricComputation(lookupName).GetCell(...)` |
| Temporal lookup indexing | Grain dimensions included in tensor dimension list; time matching at runtime |
| Multi-grain resolution | Runtime selection of appropriate grain row based on cell time coordinate |
| Lookup IO Specification | Parsed alongside other IO specs; maps lookup names to file paths |
| Error handling | Missing lookup files, malformed data, and key misses are runtime errors |
| Type representation | `float64` for all numeric cell values (detailed type selection deferred) |

## Decision

**Approach:** Single synthesis pipeline that walks CIL lookup nodes and emits Go code for data loading, registration, and access. Uses the same `MetricComputation` interface and controller as connectors.

**Rationale:**
Reusing the connector `MetricComputation` pattern reduces implementation complexity and maintains consistency across the codebase. Lookups are conceptually similar to connectors (both are data sources) and sharing the same controller interface validates that design choice. The CIL tree walker can handle `CILLookupNode` with the same traversal pattern as `CILConnectorReferenceNode`, minimizing new infrastructure.

## Implications

### Positive Implications

- Reuses existing `MetricComputation` and `ChronoconeComputationController` infrastructure
- Consistent pattern with connector synthesis reduces cognitive load
- Lookup data loading uses the same parquet reading utilities as connectors
- Single CIL node type (`CILLookupNode`) maps cleanly to a single synthesis code path

### Negative Implications / Trade-offs

- Lookup-specific tensor indexing (key-based access) may require wrapper types if the existing tensor infrastructure uses only sequential indices
- Temporal lookup resolution (multi-grain matching) is a new runtime mechanism
- Lookup loading and metric computation ordering must be handled correctly in the generated preamble
- Lookups referenced in aggregators/spreaders (Increments 4 and 5) require coordination with EROs 007 and 008

### Future Considerations

- Dynamic lookup data reloading for temporal lookups
- Lookup IO Specification schema unification with connector IO spec
- Key-miss caching to avoid repeated runtime errors for known-missing key combinations
- Batch lookup access optimization (prefetch lookup values for all cells at once rather than per-cell access)

### Notes

**Related Documents:**

- [ERO006 — Lookups Incremental Implementation](../../eros/ERO006_lookups-incremental-implementation.md) — Parent ERO defining the 5-increment strategy
- [ERO006_lookups-parsing-phase.md](../../go-parser-target/eros/ERO006_lookups-parsing-phase.md) — Parsing phase structures
- [ERO006_lookups-analysis-phase.md](../../go-analysis/eros/ERO006_lookups-analysis-phase.md) — Analysis phase structures
- [Lookups Design](../../lookups.md) — Full lookups design with data flow and synthesized code description
- [Connectors Design](../../connectors.md) — Connector design (shared MetricComputation pattern)
- [Metrics Design](../../metrics-design.md) — Metric design (metrics reference lookups via GetMetricComputation)
- [CPL Requirements — Lookups](../../../../requirements/cpl/lookups-syntax.md) — Syntax requirements for lookups
- [CPL Requirements — IO Specifications](../../../../requirements/cpl/io_specs.md) — IO specification requirements (separate lookup IO spec)
