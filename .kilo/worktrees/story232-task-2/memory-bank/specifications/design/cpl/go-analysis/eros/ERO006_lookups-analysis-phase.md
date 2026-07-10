# ERO006: Lookups Analysis Phase

## Title

Lookups Analysis Phase — Data Structures and Incremental Implementation

## Purpose/Goal

Define the analysis phase data structures for lookup definitions (including `LookupInfo`, integration with `AnalysisDefinitions`, cross-reference resolution, validation, and CIL node generation) and document how these structures and analysis functions evolve across each increment of the lookups incremental implementation as defined in [ERO006 — Lookups Incremental Implementation](../../eros/ERO006_lookups-incremental-implementation.md).

## Decision Status

Draft

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic006_cpl-lookups-implementation.md](../../../../../epics/Epic006_cpl-lookups-implementation.md)

## Context

### Situation

The analysis phase is responsible for transforming parser-level lookup definitions into analysis-level structures, validating all cross-references, resolving dimension/grain/celltype references, and generating CIL leaf nodes for each lookup. Following the [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md), the parser captures raw syntactic information in the symbol table; the analysis phase walks the symbol table tree to assemble semantically validated structures.

The parsing phase ERO ([ERO006_lookups-parsing-phase.md](../../go-parser-target/eros/ERO006_lookups-parsing-phase.md)) defines how `ParserLookupDefinition` structures are stored in the symbol table. This ERO defines how those raw structures are transformed into `LookupInfo` entries with resolved references and `CILLookupNode` CIL trees.

**Key Principle:** The analysis phase walks the symbol table tree, extracts parser definitions, validates semantics, resolves all concept and measure references, and generates CIL leaf nodes. The `CILLookupNode` is a single-node tree (leaf data source), structurally analogous to connector CIL nodes.

### Constraints

- Must work with parser-level lookup definitions from the symbol table
- Must follow the [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md) for phase separation
- Must produce structures usable by the CIL generator and synthesizer
- Must validate all concept, measure, and grain references
- Must support the incremental feature set of each increment (5 total)
- Name uniqueness checking must be performed when assembling the lookups map
- In Increment 3, must validate lookup references in metric inputs and verify dimensional compatibility

### Requirements

- Walk the symbol table tree to extract `ParserLookupDefinition` instances
- Assemble a Lookups lookup map (`map[string]*LookupInfo`) keyed by lookup name
- Resolve dimension concept names against the Concepts map in `AnalysisDefinitions`
- Resolve celltype measure name against the Measures map in `AnalysisDefinitions`
- Resolve grain measure names (when present) against the Measures map and validate `Interval`-derived types
- Validate lookup name uniqueness within the compilation unit
- Generate one `CILLookupNode` per lookup (single-node leaf CIL tree)
- In Increment 3, validate lookup references in metric inputs and verify dimensional compatibility
- In Increments 4 and 5, resolve lookup references in aggregator and spreader expressions

### Assumptions

- Parser lookup definitions are complete and syntactically valid
- The analysis phase has access to all parsed files in the compilation unit
- Concepts, types, units, measures, connectors, and metrics are already analyzed and available in `AnalysisDefinitions`
- Analysis errors prevent CIL generation and synthesis
- The `WalkSymbolTableWithHandlers` pattern from concepts/types/units/connectors analysis is available

## Data Structures by Increment

### Increment 1: Basic Lookup Declarations Without Grain

**Overview**

Increment 1 introduces the minimal lookup analysis infrastructure. The parser produces `ParserLookupDefinition` structures with dimension concept names and a celltype measure name (no grain). The analysis phase walks the symbol table, extracts these definitions, resolves all references, validates semantic rules, and generates one `CILLookupNode` per lookup.

**AnalysisDefinitions — Lookups Lookup Map**

```go
// AnalysisDefinitions contains all analyzed definitions for CIL generation.
type AnalysisDefinitions struct {
    Concepts   map[string]*AnalysisConceptDefinition
    Types      map[string]*AnalysisTypeDefinition
    Units      map[string]*AnalysisUnitDefinition
    Measures   map[string]*AnalysisMeasureDefinition
    Connectors map[string]*AnalysisConnectorDefinition
    Metrics    map[string]*AnalysisMetricDefinition
    Lookups    map[string]*LookupInfo          // NEW: Lookups lookup map
    CILTrees   map[string]CILNode              // CIL trees keyed by name (metrics + connectors + lookups)
}
```

**LookupInfo (Increment 1)**

```go
// LookupInfo contains the analyzed lookup information.
// All concept and measure references are resolved to their analysis-level definitions.
type LookupInfo struct {
    Name              string                           // Lookup identifier
    Dimensions        []*AnalysisConceptDefinition      // Resolved dimension concept references
    CelltypeMeasure   *AnalysisMeasureDefinition       // Resolved celltype measure reference
    Grain             []*AnalysisMeasureDefinition      // Empty in Increment 1
    HasGrain          bool                              // False in Increment 1
    SourceLocation    SourceLocation
}
```

**CILLookupNode (Increment 1)**

```go
// CILLookupNode is a leaf data source node in the CIL tree.
// It represents lookup data loaded at runtime via MetricComputation.
// Structurally analogous to CILConnectorReferenceNode.
type CILLookupNode struct {
    LookupName        string                           // Lookup identifier
    Dimensions        []string                         // Concept names for key dimensions (resolved)
    Celltype          string                           // Measure name for cell type (resolved)
    SourceLocation    SourceLocation
}

// Ensure CILLookupNode implements the CILNode interface.
func (n *CILLookupNode) cilNode() {}
```

The `CILLookupNode` is stored in `AnalysisDefinitions.CILTrees` keyed by the lookup name (prefixed to avoid collision with connector trees of the same name, e.g., `"lookup:product_category"`).

**Lookups Analysis Function (Increment 1)**

The lookups analysis function is registered with `WalkSymbolTableWithHandlers` and called for each `LookupScope` encountered during the walk:

```go
// AnalyzeLookupScope processes a LookupScope symbol table and builds
// a LookupInfo entry with resolved references and CIL node.
func (a *CPLAnalyzer) AnalyzeLookupScope(table *SymbolTable) error {
    // Find the LookupScopeDefinition entry that carries the ParserLookupDefinition
    var parserDef *ParserLookupDefinition
    for _, def := range table.Definitions {
        if def.DefinitionType == LookupScopeDefinition {
            var ok bool
            parserDef, ok = def.Data.(*ParserLookupDefinition)
            if !ok {
                return fmt.Errorf("invalid lookup definition data for %s", *def.Name)
            }
            break
        }
    }
    if parserDef == nil {
        return fmt.Errorf("no LookupScopeDefinition found in lookup scope")
    }

    // Check for duplicate lookup names
    if _, exists := a.lookups[parserDef.Name]; exists {
        return fmt.Errorf("duplicate lookup definition: %s", parserDef.Name)
    }

    // Resolve dimension concept names
    var dimensions []*AnalysisConceptDefinition
    for _, dimName := range parserDef.Dimensions {
        conceptDef, exists := a.concepts[dimName]
        if !exists {
            return fmt.Errorf("lookup %s: dimension concept %q not found", parserDef.Name, dimName)
        }
        dimensions = append(dimensions, conceptDef)
    }

    // Resolve celltype measure name
    celltypeMeasure, exists := a.measures[parserDef.Celltype]
    if !exists {
        return fmt.Errorf("lookup %s: celltype measure %q not found", parserDef.Name, parserDef.Celltype)
    }

    // Build LookupInfo
    info := &LookupInfo{
        Name:             parserDef.Name,
        Dimensions:       dimensions,
        CelltypeMeasure:  celltypeMeasure,
        HasGrain:         false,
        SourceLocation:   parserDef.SourceLocation,
    }

    // Store in lookups map
    a.lookups[parserDef.Name] = info

    // Generate CIL node
    dimensionNames := make([]string, len(dimensions))
    for i, d := range dimensions {
        dimensionNames[i] = d.Name
    }

    cilNode := &CILLookupNode{
        LookupName:     parserDef.Name,
        Dimensions:     dimensionNames,
        Celltype:       parserDef.Celltype,
        SourceLocation: parserDef.SourceLocation,
    }

    // Store CIL tree keyed by "lookup:<name>" to avoid collision with connector trees
    a.cilTrees["lookup:"+parserDef.Name] = cilNode

    return nil
}
```

**Validation During Analysis (Increment 1)**

- **Name Uniqueness**: Lookup name must not duplicate any other lookup name in the compilation unit
- **Dimension Existence**: All concept names in the `dimensions` clause must refer to existing concepts in `AnalysisDefinitions.Concepts`
- **Celltype Existence**: The measure name in the `celltype` clause must refer to an existing measure in `AnalysisDefinitions.Measures`
- **No Grain**: Since Increment 1 has no grain support, `parserDef.HasGrain` must be false (or equivalently, the grain clause must not be present in the parser definition)

**Limitations (Increment 1)**

- No grain validation (added in Increment 2)
- No validation of lookup references in metric inputs (added in Increment 3)
- No dimensional compatibility checks between lookups and metrics (added in Increment 3)
- No validation of lookup references in aggregator or spreader expressions (added in Increments 4 and 5)

### Increment 2: Lookups with Grain

**Overview**

Increment 2 adds grain support to lookup analysis. When `ParserLookupDefinition.HasGrain` is true, the analysis phase resolves the grain measure names against the Measures map and validates that each has a type derived from `Interval`. The `CILLookupNode` gains a `Grain` field.

**LookupInfo (Increment 2 — fields added)**

```go
type LookupInfo struct {
    Name              string
    Dimensions        []*AnalysisConceptDefinition
    CelltypeMeasure   *AnalysisMeasureDefinition
    Grain             []*AnalysisMeasureDefinition      // NEW: Resolved grain measure references (nil if !HasGrain)
    HasGrain          bool                              // True when lookup has temporal dimension
    SourceLocation    SourceLocation
}
```

**CILLookupNode (Increment 2 — fields added)**

```go
type CILLookupNode struct {
    LookupName        string
    Dimensions        []string
    Grain             []string                         // NEW: Grain measure names (empty if no grain)
    HasGrain          bool                              // NEW: Whether grain clause was present
    Celltype          string
    SourceLocation    SourceLocation
}
```

**Grain Validation (Increment 2)**

Added to `AnalyzeLookupScope`:

```go
    // Resolve grain measure names (if present)
    var grainMeasures []*AnalysisMeasureDefinition
    var grainNames []string
    if parserDef.HasGrain {
        if len(parserDef.Grain) == 0 {
            return fmt.Errorf("lookup %s: grain clause present but no grain measures specified", parserDef.Name)
        }
        for _, grainName := range parserDef.Grain {
            measureDef, exists := a.measures[grainName]
            if !exists {
                return fmt.Errorf("lookup %s: grain measure %q not found", parserDef.Name, grainName)
            }
            // Validate that the measure's type is Interval or derived from Interval
            if !isIntervalType(measureDef.TypeRef) {
                return fmt.Errorf(
                    "lookup %s: grain measure %q has type %s, which is not Interval or derived from Interval",
                    parserDef.Name, grainName, measureDef.TypeRef.Name,
                )
            }
            grainMeasures = append(grainMeasures, measureDef)
            grainNames = append(grainNames, grainName)
        }
        // Check for duplicate grain measure names
        if len(grainNames) != len(uniqueStrings(grainNames)) {
            return fmt.Errorf("lookup %s: duplicate grain measure names in grain clause", parserDef.Name)
        }
    }

    info.HasGrain = parserDef.HasGrain
    info.Grain = grainMeasures
```

The `isIntervalType` helper checks whether a type reference (directly or through its parent chain) is `Interval`:

```go
// isIntervalType returns true if the given AnalysisTypeDefinition is Interval
// or is a type derived from Interval.
func isIntervalType(typeDef *AnalysisTypeDefinition) bool {
    if typeDef == nil {
        return false
    }
    if typeDef.Name == "Interval" {
        return true
    }
    return isIntervalType(typeDef.ParentType)
}
```

**CILLookupNode Construction (Increment 2 — modified)**

```go
    cilNode := &CILLookupNode{
        LookupName:     parserDef.Name,
        Dimensions:     dimensionNames,
        Grain:          grainNames,
        HasGrain:       parserDef.HasGrain,
        Celltype:       parserDef.Celltype,
        SourceLocation: parserDef.SourceLocation,
    }
```

**Validation During Analysis (Increment 2 — additions)**

- **Grain Existence**: All measure names in the `grain` clause must refer to existing measures
- **Grain Type**: All grain measures must have types that are `Interval` or derived from `Interval`
- **Empty Grain**: If `HasGrain` is true, `len(Grain)` must be > 0 (a grain clause with zero measures is an error)
- **Duplicate Grain**: Grain measure names must be unique within a single lookup's grain clause

### Increment 3: Lookup Integration with Metric Calculations

**Overview**

Increment 3 adds lookup reference resolution during metric analysis. When a metric input has `InputKind == "lookup"`, the analyzer resolves the lookup name against the Lookups map and validates dimensional compatibility between the lookup's dimensions and the metric's dimensions.

**AnalysisInputDefinition (Increment 3 — fields added)**

```go
// AnalysisInputDefinition represents a resolved metric input reference.
// Existing struct from Increment 1 of metrics analysis.
type AnalysisInputDefinition struct {
    InputKind       string                              // "metric", "dimension", or "lookup" (NEW value)
    Name            string                              // Referenced name
    ResolvedMetric  *AnalysisMetricDefinition           // Resolved reference (for metric inputs)
    ResolvedLookup  *LookupInfo                         // NEW: Resolved reference (for lookup inputs)
    SourceLocation  SourceLocation
}
```

**Lookup Input Resolution (Increment 3)**

Added to the metric analysis function when processing input definitions:

```go
    // Resolve input references
    var inputs []AnalysisInputDefinition
    for _, parserInput := range parserDef.Inputs {
        input := AnalysisInputDefinition{
            InputKind:      parserInput.InputKind,
            Name:           parserInput.Name,
            SourceLocation: parserInput.SourceLocation,
        }

        switch parserInput.InputKind {
        case "metric":
            metricDef, exists := a.metrics[parserInput.Name]
            if !exists {
                return fmt.Errorf("metric %s: input metric %q not found", parserDef.Name, parserInput.Name)
            }
            input.ResolvedMetric = metricDef

        case "dimension":
            // Dimension references are resolved later during dimensional analysis
            // (no immediate resolution needed)

        case "lookup":      // NEW
            lookupDef, exists := a.lookups[parserInput.Name]
            if !exists {
                return fmt.Errorf("metric %s: input lookup %q not found", parserDef.Name, parserInput.Name)
            }
            input.ResolvedLookup = lookupDef

            // Validate dimensional compatibility between lookup and metric
            if err := a.validateLookupMetricCompatibility(lookupDef, parserDef); err != nil {
                return fmt.Errorf("metric %s: %w", parserDef.Name, err)
            }

        default:
            return fmt.Errorf("metric %s: unknown input kind %q", parserDef.Name, parserInput.InputKind)
        }

        inputs = append(inputs, input)
    }
```

**Dimensional Compatibility Validation (Increment 3)**

```go
// validateLookupMetricCompatibility verifies that a lookup's dimensions are
// compatible with the metric that references it.
func (a *CPLAnalyzer) validateLookupMetricCompatibility(lookup *LookupInfo, parserDef *ParserMetricDefinition) error {
    // The metric's declared dimensions must include all of the lookup's key dimensions.
    // If the metric is dimensioned by [Country, Product, Currency] and the lookup
    // is keyed by [Currency], that's valid — the metric can provide the currency
    // value for each cell.
    // If the lookup is keyed by [Region] but the metric is only dimensioned by
    // [Country, Product], that's an error — Region is not available in the metric's context.
    //
    // Full dimensional analysis is deferred to the metric dependency DAG pass.
    // This is a preliminary structural check.

    // Build a set of the metric's dimension concepts
    metricDims := make(map[string]bool)
    for _, input := range parserDef.Inputs {
        if input.InputKind == "dimension" {
            metricDims[input.Name] = true
        }
    }

    // Check each lookup dimension
    for _, dim := range lookup.Dimensions {
        if !metricDims[dim.Name] {
            return fmt.Errorf(
                "lookup %s is keyed by %q but metric does not declare %q as a dimension",
                lookup.Name, dim.Name, dim.Name,
            )
        }
    }

    return nil
}
```

**CILLookupRefNode (Increment 3 — new CIL node)**

When a metric's `calculate` expression references a lookup input, the CIL tree must include a node that represents the lookup value access:

```go
// CILLookupRefNode represents a reference to a lookup value at runtime.
// The synthesizer emits controller.GetMetricComputation(lookupName) for this node.
type CILLookupRefNode struct {
    LookupName    string
    Dimensions    []string                         // Metric's dimension context for indexing
    Celltype      string
}

func (n *CILLookupRefNode) cilNode() {}
```

This node is generated during expression analysis (postfix evaluation) when an identifier resolves to a lookup input name.

**Validation During Analysis (Increment 3 — additions)**

- **Lookup Reference Existence**: The lookup name in a metric's `lookup` input must refer to an existing lookup in `AnalysisDefinitions.Lookups`
- **Dimensional Compatibility**: The lookup's key dimensions must be a subset of the metric's declared dimensions
- No grain compatibility check at this level (the full metric dependency DAG pass handles grain alignment)

### Increment 4: Lookups in Aggregator Expressions

**Overview**

No analysis changes are required in Increment 4 at the lookup analysis level. When an aggregator expression references a lookup by name, the aggregator analysis phase (defined in ERO007's analysis ERO) resolves the reference against the Lookups map and validates dimensional unit compatibility.

The lookup analysis phase provides the Lookups map to the aggregator analysis — no new lookup-level data structures are needed.

**Dependency on Lookups Map**

The aggregator analysis function accesses `AnalysisDefinitions.Lookups` to resolve lookup references:

```go
// In aggregator analysis (ERO007 analysis-phase ERO):
func (a *CPLAnalyzer) resolveAggregatorExpressionReferences(expr []SymbolDefinition) error {
    for _, def := range expr {
        // Check if def references a lookup
        if lookup, exists := a.lookups[def.Name]; exists {
            // Validate dimensional unit compatibility with aggregator's transformation
            // ...
        }
    }
}
```

This is documented here for completeness but the implementation details belong in the aggregator analysis phase ERO (ERO007).

### Increment 5: Lookups in Spreader Expressions

**Overview**

No analysis changes are required in Increment 5 at the lookup analysis level. When a spreader expression references a lookup by name, the spreader analysis phase (defined in ERO008's analysis ERO) resolves the reference against the Lookups map and applies the same dimensional unit analysis used for metric references in spreaders.

The lookup analysis phase provides the Lookups map to the spreader analysis — no new lookup-level data structures are needed.

## Function Evolution Across Increments

| Function | Inc 1 | Inc 2 | Inc 3 | Inc 4 | Inc 5 |
|----------|-------|-------|-------|-------|-------|
| `AnalyzeLookupScope` | Dimension/celltype resolution | Grain resolution + Interval validation | No changes | No changes | No changes |
| `validateLookupMetricCompatibility` | — | — | Added | No changes | No changes |
| Lookups Map | Populated | Populated with grain | Populated (used by metric analysis) | Used by aggregator analysis | Used by spreader analysis |
| CIL Node Generation | `CILLookupNode` | `CILLookupNode` with grain | `CILLookupRefNode` in metrics | No changes | No changes |

## Structure Evolution Across Increments

| Structure | Inc 1 | Inc 2 | Inc 3 | Inc 4 | Inc 5 |
|-----------|-------|-------|-------|-------|-------|
| `LookupInfo.Dimensions` | Present | Present | Present | Present | Present |
| `LookupInfo.CelltypeMeasure` | Present | Present | Present | Present | Present |
| `LookupInfo.Grain` | nil | `[]*AnalysisMeasureDefinition` | Present | Present | Present |
| `LookupInfo.HasGrain` | — | Added | Present | Present | Present |
| `CILLookupNode.Grain` | — | Added | Present | Present | Present |
| `CILLookupNode.HasGrain` | — | Added | Present | Present | Present |
| `AnalysisInputDefinition.ResolvedLookup` | — | — | Added | Present | Present |
| `CILLookupRefNode` | — | — | Added | Present | Present |

## Decision

**Approach:** Single evolving structures (`LookupInfo`, `CILLookupNode`) with incrementally added fields, following the same evolution pattern as `AnalysisConceptDefinition` and `AnalysisMetricDefinition`.

**Rationale:**
The simplicity of evolving a single structure outweighs the benefits of separate structs per increment. Go zero values provide sensible defaults for unused fields — `Grain` is nil when not present (Increment 1), `HasGrain` defaults to false. This is the same approach used for all other CPL analysis-phase EROs.

## Implications

### Positive Implications

- Single structure definition per concept, reducing code duplication
- Easy to add new fields as increments progress
- Backward compatible (new fields have zero values)
- Consistent with all other CPL analysis-phase EROs
- `CILLookupNode` mirrors `CILConnectorReferenceNode`, reducing cognitive load for the CIL tree walker

### Negative Implications / Trade-offs

- Must rely on documentation and validation to enforce increment constraints
- Increment 3 requires coordination with the metrics analysis phase (modifying `AnalysisInputDefinition`)
- Increments 4 and 5 require the Lookups map to be accessible by aggregator and spreader analyzers

### Validation Strategy (Per Increment)

| Increment | Validation |
|-----------|-----------|
| 1 | Name uniqueness, dimension existence, celltype existence |
| 2 | Grain existence, grain type (`Interval`-derived), empty grain check, duplicate grain check |
| 3 | Lookup reference existence in metrics, dimensional compatibility (subset check) |
| 4 | No new lookup-level validation (aggregator analysis handles lookup expression references) |
| 5 | No new lookup-level validation (spreader analysis handles lookup expression references) |

### Notes

**Related Documents:**

- [ERO006 — Lookups Incremental Implementation](../../eros/ERO006_lookups-incremental-implementation.md) — Parent ERO defining the 5-increment strategy
- [ERO006_lookups-parsing-phase.md](../../go-parser-target/eros/ERO006_lookups-parsing-phase.md) — Parsing phase structures and grammar
- [ERO006_lookups-synthesis-phase.md](../../go-synthesis-target/eros/ERO006_lookups-synthesis-phase.md) — Synthesis phase code generation
- [Lookups Design](../../lookups.md) — Full lookups design with definition types and data flow
- [CPL Requirements — Lookups](../../../../requirements/cpl/lookups-syntax.md) — Syntax requirements for lookups
- [Metrics Design](../../metrics-design.md) — Metric design (metrics reference lookups as inputs in Increment 3)
- [Aggregators Design](../../aggregators.md) — Aggregator design (lookups referenced in aggregator expressions in Increment 4)
- [Spreaders Design](../../spreaders.md) — Spreader design (lookups referenced in spreader expressions in Increment 5)
- [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md) — Architectural pattern for phase separation
