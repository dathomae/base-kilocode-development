# ERO008: Spreaders Analysis Phase

## Title

Spreaders Analysis Phase — Data Structures and Incremental Implementation

## Purpose/Goal

Define the analysis phase data structures for spreader definitions (including `SpreaderInfo`, the Spreaders lookup map, the spreader `FunctionRegistry`, metric reference resolution with implicit aggregator determination, and CIL node generation) and document how these structures and analysis functions evolve across each increment of the spreaders incremental implementation as defined in [ERO008 — Spreaders Incremental Implementation](../../eros/ERO008_spreaders-incremental-implementation.md).

## Decision Status

Accepted

## Implementation Status

Increment 1 Implemented — `SpreaderFunctionRegistry`, `AnalyzeSpreaderScope`, `SpreaderInfo`, `CILSpreadNode`, and `SpreadingMatrix` are implemented and tested. Metric reference resolution and DAG integration are deferred to later increments.

**Related Epics:**

- [Epic008_cpl-spreaders-implementation.md](../../../../../epics/Epic008_cpl-spreaders-implementation.md)

## Context

### Situation

The analysis phase transforms parser-level spreader definitions into analysis-level structures. Spreaders share the same postfix evaluation pattern and dimensional unit algebra as aggregators, but add **metric reference resolution** — when a spreader expression references a metric by name (e.g., `Product / customers`), the analyzer must determine the correct dimensional form of that metric and apply implicit aggregators if needed.

The parsing phase ERO ([ERO008_spreaders-parsing-phase.md](../../go-parser-target/eros/ERO008_spreaders-parsing-phase.md)) defines how `ParserSpreaderDefinition` structures and postfix expression sequences are stored under synthetic keys. This ERO defines the transformation into `SpreaderInfo` entries with validated dimensional units, resolved metric references, and `CILSpreadNode` CIL trees.

Spreader functions use the same generic `fnCall` syntax as aggregators and are resolved through a **spreader-specific FunctionRegistry**. The spreader FunctionRegistry maps names like `size`, `1/size`, `sum`, `count`, `1/sum`, `1/count` to their CIL node types and dimensional unit semantics in the **spreading** context (which differ from the aggregating context).

### Constraints

- Must work with parser-level spreader definitions (synthetic keys)
- Must reuse the dimensional unit algebra from aggregators (`DimensionalUnitExpression`, `Multiply`, `cancel`)
- Spreaders add one dimension (inverse of aggregators which remove one)
- Metric references in expressions must be resolved against the Metrics map
- Implicit aggregator resolution required when referenced metrics have incompatible dimensional forms
- Matrix multiplication infrastructure already exists (ERO007 Increment 1)

### Requirements

- Walk the symbol table to extract `ParserSpreaderDefinition` from synthetic-key entries
- Assemble the Spreaders lookup map keyed by `(from_concept_set, to_concept_set)` with `SpreaderInfo` values
- Implement the spreader FunctionRegistry with all supported functions
- Perform postfix evaluation of spreader expressions to compute dimensional unit effects
- Resolve identifier references (metrics → concepts → lookups) in priority order
- Perform implicit aggregator resolution for metric references with incompatible dimensional forms
- Generate `CILSpreadNode` CIL trees with pre-built spreading matrices
- Insert spreader nodes into metric CIL trees when dimension mismatches detected

## Data Structures by Increment

### Increment 1: Basic Spreader Parsing and Validation

**Overview**

Increment 1 introduces the core spreader analysis infrastructure. Spreader functions are resolved through the FunctionRegistry using the same generic fnCall mechanism as aggregators. All supported functions are registered from the start (unlike aggregators which started with only `sum`).

**AnalysisDefinitions — Spreaders Lookup Map**

```go
type AnalysisDefinitions struct {
    // ... existing maps ...
    Aggregators map[string]*AggregatorInfo
    Spreaders   map[string]*SpreaderInfo    // NEW: Spreaders lookup map
    CILTrees    map[string]CILNode
}
```

The Spreaders map uses a compound string key:

```go
// SpreaderMapKey constructs the compound key for the Spreaders map.
// Sorted concept names ensure deterministic keys.
func SpreaderMapKey(fromConcepts, toConcepts []string) string {
    fromSorted := make([]string, len(fromConcepts))
    copy(fromSorted, fromConcepts)
    sort.Strings(fromSorted)
    toSorted := make([]string, len(toConcepts))
    copy(toSorted, toConcepts)
    sort.Strings(toSorted)
    return fmt.Sprintf("[%s]->[%s]", strings.Join(fromSorted, ","), strings.Join(toSorted, ","))
}
```

**SpreaderInfo (Increment 1)**

```go
type SpreaderInfo struct {
    FromConcepts         []*AnalysisConceptDefinition
    ToConcepts           []*AnalysisConceptDefinition
    FromEmpty            bool
    NewDimension         *AnalysisConceptDefinition     // The concept being introduced
    FromDimensionalUnits *DimensionalUnitExpression
    ToDimensionalUnits   *DimensionalUnitExpression
    ExpressionDimUnits   *DimensionalUnitExpression
    SourceLocation       SourceLocation
}
```

**Spreader FunctionRegistry (Increment 1)**

```go
type SpreaderFunctionInfo struct {
    Name        string
    CILNodeType string
    // DimUnits computes the dimensional unit effect of the function.
    // For size(C): returns Unitless (empty expression)
    // For 1/size(C): returns Unitless
    // For sum(C): returns C (used when implicitly aggregating a referenced metric)
    DimUnits    func(params []string) *DimensionalUnitExpression
    Description string
}

type SpreaderFunctionRegistry struct {
    Functions map[string]*SpreaderFunctionInfo
}

func NewSpreaderFunctionRegistry() *SpreaderFunctionRegistry {
    reg := &SpreaderFunctionRegistry{
        Functions: make(map[string]*SpreaderFunctionInfo),
    }

    reg.Functions["size"] = &SpreaderFunctionInfo{
        Name:        "size",
        CILNodeType: "CILSpreadSizeNode",
        DimUnits: func(params []string) *DimensionalUnitExpression {
            return &DimensionalUnitExpression{} // Unitless
        },
        Description: "Returns the number of distinct values in a concept",
    }

    reg.Functions["1/size"] = &SpreaderFunctionInfo{
        Name:        "1/size",
        CILNodeType: "CILSpreadInverseSizeNode",
        DimUnits: func(params []string) *DimensionalUnitExpression {
            return &DimensionalUnitExpression{} // Unitless
        },
        Description: "Returns 1/N where N is the number of values in a concept",
    }

    reg.Functions["sum"] = &SpreaderFunctionInfo{
        Name:        "sum",
        CILNodeType: "CILSpreadSumNode",
        DimUnits: func(params []string) *DimensionalUnitExpression {
            if len(params) != 1 { return nil }
            return &DimensionalUnitExpression{Numerator: []string{params[0]}}
        },
        Description: "Sums values across a dimension (used for implicit aggregations)",
    }

    reg.Functions["count"] = &SpreaderFunctionInfo{
        Name:        "count",
        CILNodeType: "CILSpreadCountNode",
        DimUnits: func(params []string) *DimensionalUnitExpression {
            if len(params) != 1 { return nil }
            return &DimensionalUnitExpression{Numerator: []string{params[0]}}
        },
        Description: "Counts values in a dimension",
    }

    reg.Functions["1/sum"] = &SpreaderFunctionInfo{
        Name:        "1/sum",
        CILNodeType: "CILSpreadInverseSumNode",
        DimUnits: func(params []string) *DimensionalUnitExpression {
            if len(params) != 1 { return nil }
            return &DimensionalUnitExpression{Denominator: []string{params[0]}}
        },
        Description: "Inverse of sum across a dimension",
    }

    reg.Functions["1/count"] = &SpreaderFunctionInfo{
        Name:        "1/count",
        CILNodeType: "CILSpreadInverseCountNode",
        DimUnits: func(params []string) *DimensionalUnitExpression {
            if len(params) != 1 { return nil }
            return &DimensionalUnitExpression{Denominator: []string{params[0]}}
        },
        Description: "Inverse of count across a dimension",
    }

    return reg
}
```

**Expression Evaluation for Spreaders (Increment 1)**

Spreader expression evaluation is similar to aggregator evaluation but additionally handles metric references:

```go
func (a *CPLAnalyzer) evaluateSpreaderExpression(
    definitions []*SymbolDefinition,
    fromConcepts, toConcepts []*AnalysisConceptDefinition,
) (*DimensionalUnitExpression, error) {

    var stack []*DimensionalUnitExpression

    for _, def := range definitions {
        operand, ok := def.Data.(*OperandData)
        if !ok {
            continue // Skip non-operand definitions
        }

        switch operand.Kind {
        case "concept":
            // Concept reference -> push its dimensional unit
            stack = append(stack, &DimensionalUnitExpression{
                Numerator: []string{operand.Name},
            })

        case "identifier":
            // Identifier: resolve as metric -> concept -> lookup
            resolved, err := a.resolveSpreaderIdentifier(operand.Name)
            if err != nil {
                return nil, err
            }
            stack = append(stack, resolved)

        case "fncall":
            // Look up function in spreader registry
            fnInfo, exists := a.spreaderFunctionRegistry.Functions[operand.Name]
            if !exists {
                return nil, fmt.Errorf("unknown spreader function %q", operand.Name)
            }
            dimUnits := fnInfo.DimUnits(operand.Params)
            if dimUnits == nil {
                return nil, fmt.Errorf("invalid params for %q: %v", operand.Name, operand.Params)
            }
            stack = append(stack, dimUnits)

        case "operator":
            // Same arithmetic handling as aggregators (*, /)
            switch operand.Name {
            case "*":
                if len(stack) < 2 { return nil, errInsufficientOps }
                right := stack[len(stack)-1]
                left := stack[len(stack)-2]
                stack = stack[:len(stack)-2]
                stack = append(stack, left.Multiply(right))
            case "/":
                if len(stack) < 2 { return nil, errInsufficientOps }
                right := stack[len(stack)-1]
                left := stack[len(stack)-2]
                stack = stack[:len(stack)-2]
                inverted := &DimensionalUnitExpression{
                    Numerator:   right.Denominator,
                    Denominator: right.Numerator,
                }
                stack = append(stack, left.Multiply(inverted))
            }
        }
    }

    if len(stack) != 1 {
        return nil, fmt.Errorf("expression did not reduce to a single dimensional unit")
    }
    return stack[0], nil
}
```

**Identifier Resolution (Increment 1 — concepts only; Increment 3 — metric/lookup)**

```go
// resolveSpreaderIdentifier resolves an identifier in a spreader expression.
// Priority: metric reference > concept reference > lookup reference.
func (a *CPLAnalyzer) resolveSpreaderIdentifier(name string) (*DimensionalUnitExpression, error) {
    // Increment 1: concepts only. Increment 3 adds metric and lookup resolution.
    // For Increment 1, identifier resolution is handled inline in
    // evaluateSpreaderExpression: identifiers that match known functions are
    // treated as fncall operands; identifiers that match concepts in from/to
    // are treated as concept references; all other identifiers produce an
    // "unknown identifier" error.
    //
    // The full priority-ordered resolution (metric > concept > lookup) shown
    // below is the Increment 3 target:
    /*
    if metric, exists := a.metrics[name]; exists {
        return dimensionalUnitsFromConceptSet(metric.Dimensions), nil
    }
    // 2. Check if it's a concept
    if concept, exists := a.concepts[name]; exists {
        return &DimensionalUnitExpression{Numerator: []string{name}}, nil
    }
    // 3. Check if it's a lookup
    if _, exists := a.lookups[name]; exists {
        // Lookups have their own dimensional units based on key dimensions
        return dimensionalUnitsFromConceptSet(lookup.KeyDimensions), nil
    }
    return nil, fmt.Errorf("unresolved identifier %q in spreader expression", name)
    */
}
```

**CILSpreadNode (Increment 1)**

```go
type CILSpreadNode struct {
    SourceCIL             CILNode
    FromDimensions        []string
    ToDimensions          []string
    NewDimension          string
    SpreadingMatrix       *AggregationMatrix
    DistributionExpression CILNode         // CIL tree for the distribution expression
    ReferencedMetrics     []CILNode        // CIL trees for referenced metrics (nil in Increment 1)
    DimensionalUnitEffect string
    SourceLocation        SourceLocation
}
func (n *CILSpreadNode) cilNode() {}
```

**Dimensional Unit Validation for Spreaders (Increment 1)**

The validation formula for spreaders is the inverse of aggregators:
- Input dim units: computed from `from` concept set
- Output dim units: computed from `to` concept set  
- Spreader expression dim units: computed from postfix evaluation
- Validation: `input_units * expression_units = output_units`

```go
func (a *CPLAnalyzer) validateSpreaderDimensionalUnits(info *SpreaderInfo) error {
    info.FromDimensionalUnits = dimensionalUnitsFromConceptSet(info.FromConcepts)
    info.ToDimensionalUnits = dimensionalUnitsFromConceptSet(info.ToConcepts)

    predicted := info.FromDimensionalUnits.Multiply(info.ExpressionDimUnits)
    if !predicted.Equals(info.ToDimensionalUnits) {
        return fmt.Errorf(
            "spreader dimensional unit mismatch: input %s * expression %s = %s, but output requires %s",
            info.FromDimensionalUnits.String(),
            info.ExpressionDimUnits.String(),
            predicted.String(),
            info.ToDimensionalUnits.String(),
        )
    }
    return nil
}
```

**SpreadingMatrix Construction (Increment 1)**

```go
// BuildSpreadingMatrix creates a matrix for dimension expansion.
// Each row corresponds to a source concept value, each column to a new dimension value.
// For equal distribution: matrix[i][j] = weight (same for all j given i).
// The weights are determined by the distribution expression.
func BuildSpreadingMatrix(
    newDimValues []string,    // Values of the new dimension (e.g., ["USA", "Canada"])
    expressionWeights map[string]float64, // Weights per new dimension value from expression
) *AggregationMatrix {
    rows := 1  // One source value per cell in the source tensor
    cols := len(newDimValues)

    data := make([][]float64, rows)
    for i := range data {
        data[i] = make([]float64, cols)
        for j, val := range newDimValues {
            data[i][j] = expressionWeights[val]
        }
    }

    return &AggregationMatrix{
        Rows:      rows,
        Columns:   cols,
        Data:      data,
        ColLabels: newDimValues,
    }
}
```

**Validation During Analysis (Increment 1)**

- **From-To Containment**: All concepts in `from` must appear in `to`
- **Exactly One Dimension Added**: `len(to) == len(from) + 1`
- **Concept Existence**: All concept names in `from`, `to`, and expression must exist
- **Dimensional Unit Compatibility**: `input_units * expression_units = output_units`
- **No Metric References**: In Increment 1, identifiers in expressions must resolve to concepts (not metrics or lookups — those are Increment 3)

**Limitations (Increment 1)**

- Metrics cannot be referenced in spreader expressions (Increment 3)
- Spreaders not yet integrated into metric DAG (Increment 2)
- Only concept-based distribution expressions (no weighted distributions via metric data)

### Increment 2: Spreader Selection and Integration

**Overview**

Increment 2 integrates spreaders into the metric dependency DAG analysis. When a metric's output has more dimensions than its inputs, the analyzer queries the Spreaders map for single unambiguous sequences.

**Dimension Mismatch Detection (Increment 2)**

```go
func (a *CPLAnalyzer) resolveMetricDimensionalitySpreaders(
    metric *AnalysisMetricDefinition,
) error {
    outputDims := a.getMetricOutputDimensions(metric)

    for _, input := range metric.Inputs {
        if input.InputKind != "metric" { continue }

        inputDims := input.ResolvedMetric.Dimensions

        // Find dimensions in output that are not in input
        var missingDims []string
        for _, dim := range outputDims {
            if !contains(inputDims, dim.Name) {
                missingDims = append(missingDims, dim.Name)
            }
        }

        if len(missingDims) == 0 { continue }

        // Find single unambiguous spreader sequence
        sequence, err := a.findSingleSpreaderSequence(inputDims, outputDims, missingDims)
        if err != nil {
            return fmt.Errorf("metric %s: %w", metric.Name, err)
        }

        // Chain CILSpreadNode instances
        currentCIL := input.CILTree
        for _, spreadKey := range sequence {
            spreadInfo := a.spreaders[spreadKey]
            spreadNode := a.buildSpreadNode(spreadInfo, currentCIL)
            currentCIL = spreadNode
        }
        input.CILTree = currentCIL
    }
    return nil
}
```

**Limitations (Increment 2)**

- Only single unambiguous sequences supported; multiple sequences deferred to Increment 4
- No metric references in spreader expressions (Increment 3)
- No interleaved aggregator/spreader sequences (Increment 4)

### Increment 3: Metric References in Spreader Expressions

**Overview**

Increment 3 enables spreader expressions to reference metrics by name (e.g., `Product / customers`). When a metric reference is detected, the analyzer resolves its dimensional form and determines whether implicit aggregators are needed.

**Metric Reference Resolution with Implicit Aggregators (Increment 3)**

```go
// resolveSpreaderMetricReference resolves a metric reference in a spreader expression.
// If the metric's dimensional form is incompatible with the required form,
// an implicit aggregator is applied.
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
    // Determine the dim unit factor needed: raw * factor = required
    // factor = required / raw
    requiredFactor := requiredDimUnits.Divide(rawDimUnits)

    // Search aggregator map for a transformation that produces this factor
    aggKey := findAggregatorByDimUnitEffect(a.aggregators, requiredFactor)
    if aggKey == "" {
        return nil, fmt.Errorf(
            "spreader references metric %q but no aggregator exists that transforms %s to %s",
            metricName, rawDimUnits.String(), requiredDimUnits.String(),
        )
    }

    aggInfo := a.aggregators[aggKey]

    // Build implicit aggregator node wrapping the metric reference
    return &CILAggregateNode{
        SourceCIL: &CILMetricRefNode{
            MetricName: metricName,
            CILTree:    metricDef.CILTree,
        },
        FromDimensions:    a.getConceptNames(aggInfo.FromConcepts),
        ToDimension:       "",
        AggregationMatrix: aggInfo.Matrix,
        CellOperation:     aggInfo.CellOperation,
    }, nil
}
```

**CILSpreadNode with ReferencedMetrics (Increment 3 — modified)**

The `CILSpreadNode` gains a `ReferencedMetrics` field:

```go
type CILSpreadNode struct {
    // ... existing fields ...
    ReferencedMetrics     []CILNode  // NEW: CIL subtrees for referenced metrics (with implicit aggregators)
}
```

**Validation During Analysis (Increment 3 — additions)**

- **Metric Reference Existence**: Referenced metric must exist in the Metrics map
- **Required Aggregator Existence**: If a metric needs an implicit aggregator, it must exist
- **Cell Unit Compatibility**: The spreader expression must produce result cell units compatible with the output metric

### Increment 4: Spreader/Aggregator Interleaving and Optimization

**Overview**

Increment 4 implements the full selection algorithm for mixed aggregator/spreader sequences, ambiguous spreader resolution, and fused operations.

**Mixed Sequence Enumeration (Increment 4)**

```go
func (a *CPLAnalyzer) enumerateMixedSequences(
    fromDims, toDims []string,
) ([][]DimensionalOp, error) {

    toRemove := setDifference(fromDims, toDims)
    toAdd := setDifference(toDims, fromDims)

    // Enumerate all combinations of aggregator sequences (remove excess)
    // and spreader sequences (add missing), interleaved optimally
    var sequences [][]DimensionalOp
    a.enumerateMixedRecursive(fromDims, toDims, toRemove, toAdd, nil, &sequences)

    // Selection criteria:
    // 1. Least information loss (prefer spreading late, aggregating early)
    // 2. Minimal work
    // 3. Deterministic tiebreaker
    sort.Slice(sequences, func(i, j int) bool {
        return a.compareMixedSequences(sequences[i], sequences[j]) < 0
    })

    return sequences, nil
}
```

**Specificity Comparison (Increment 4)**

When multiple spreaders share the same `(from, to)` key, specificity (number of concepts in `from`) determines which to use:

```go
func (a *CPLAnalyzer) resolveAmbiguousSpreaders() error {
    groups := make(map[string][]*SpreaderInfo)
    for key, info := range a.spreaders {
        groups[key] = append(groups[key], info)
    }

    for key, infos := range groups {
        if len(infos) == 1 { continue }

        maxSpec := 0
        var best *SpreaderInfo
        tie := false
        for _, info := range infos {
            spec := len(info.FromConcepts)
            if spec > maxSpec {
                maxSpec = spec
                best = info
                tie = false
            } else if spec == maxSpec {
                tie = true
            }
        }
        if tie {
            return fmt.Errorf("ambiguous spreaders for pattern %s", key)
        }
        a.spreaders[key] = best
    }
    return nil
}
```

### Notes

**Related Documents:**

- [ERO008 — Spreaders Incremental Implementation](../../eros/ERO008_spreaders-incremental-implementation.md)
- [ERO008_spreaders-parsing-phase.md](../../go-parser-target/eros/ERO008_spreaders-parsing-phase.md)
- [ERO008_spreaders-synthesis-phase.md](../../go-synthesis-target/eros/ERO008_spreaders-synthesis-phase.md)
- [Spreaders Design](../../spreaders.md)
- [Aggregators Design](../../aggregators.md)
- [Metrics Design](../../metrics-design.md)
- [Units and Measures Design](../../units-and-measures.md)
- [Postfix Expression Parsing Pattern](../../../../patterns/postfix-expression-parsing.md)
- [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md)
