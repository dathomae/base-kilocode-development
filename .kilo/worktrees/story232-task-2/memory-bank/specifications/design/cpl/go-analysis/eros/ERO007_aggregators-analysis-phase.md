# ERO007: Aggregators Analysis Phase

## Title

Aggregators Analysis Phase — Data Structures and Incremental Implementation

## Purpose/Goal

Define the analysis phase data structures for aggregator definitions (including `AggregatorInfo`, the Aggregators lookup map, dimensional unit analysis, and CIL node generation) and document how these structures and analysis functions evolve across each increment of the aggregators incremental implementation as defined in [ERO007 — Aggregators Incremental Implementation](../../eros/ERO007_aggregators-incremental-implementation.md).

## Decision Status

Draft

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic007_cpl-aggregators-implementation.md](../../../../../epics/Epic007_cpl-aggregators-implementation.md)

## Context

### Situation

The analysis phase is responsible for transforming parser-level aggregator definitions into analysis-level structures, validating all cross-references, performing dimensional unit analysis, and generating CIL nodes for each aggregator. Following the [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md), the parser captures raw syntactic information in the symbol table; the analysis phase walks the symbol table tree to assemble semantically validated structures.

The parsing phase ERO ([ERO007_aggregators-parsing-phase.md](../../go-parser-target/eros/ERO007_aggregators-parsing-phase.md)) defines how `ParserAggregatorDefinition` structures and postfix expression sequences are stored under synthetic keys. This ERO defines how those raw structures are transformed into `AggregatorInfo` entries with resolved references, validated dimensional unit effects, and `CILAggregateNode` CIL trees.

**Key Principle:** The analysis phase walks the symbol table tree, extracts parser definitions from synthetic-key entries, validates semantics, performs postfix evaluation of expressions to compute dimensional unit effects, and generates CIL nodes. Aggregators are then applied during metric dependency DAG analysis when dimension mismatches are detected.

### Increment 1 Note

The aggregators ERO's Increment 1 (Tensor Matrix Multiplication Infrastructure) has **no analysis phase component**. It is pure Go tensor math with no CPL analysis work. Analysis begins at Increment 2.

### Constraints

- Must work with parser-level aggregator definitions from the symbol table (synthetic keys)
- Must follow the [Postfix Expression Parsing Pattern](../../../../patterns/postfix-expression-parsing.md) for expression evaluation
- Must follow the [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md) for phase separation
- Must produce structures usable by the CIL generator and synthesizer
- Must validate all concept references in `from`, `to`, and expression
- The Aggregators lookup map must support querying by `(from_concept_set, to_concept)` key
- Dimensional unit analysis must validate `input_units * expression_units = output_units`
- Increment 4 must query the Aggregators map during metric DAG analysis

### Requirements

- Walk the symbol table tree to extract `ParserAggregatorDefinition` from synthetic-key entries
- Perform postfix evaluation of aggregator expressions to compute the dimensional unit effect
- Assemble the Aggregators lookup map keyed by `(from_concept_set, to_concept)` with `AggregatorInfo` values
- Resolve concept references in `from`, `to`, and expression against the Concepts map
- Validate dimensional unit compatibility for each aggregator
- Validate cell unit compatibility for each aggregator
- Generate pre-built `AggregationMatrix` from concept value sets for each aggregator
- In Increment 4: query the Aggregators map during metric DAG analysis, detect dimension mismatches, insert appropriate CIL aggregate nodes into metric CIL trees
- In Increment 5: implement the full aggregator sequence selection algorithm

### Assumptions

- Parser aggregator definitions are complete and syntactically valid
- The analysis phase has access to all parsed files in the compilation unit
- Concepts, types, units, measures, connectors, and metrics are already analyzed
- The `WalkSymbolTableWithHandlers` pattern is available
- Concept value sets are available, enabling pre-computation of aggregation matrices

## Data Structures by Increment

### Increment 1: Tensor Matrix Multiplication Infrastructure

No analysis phase changes. This increment is pure synthesis.

### Increment 2: Basic Aggregator Parsing and Validation

**Overview**

Increment 2 introduces the core aggregator analysis infrastructure. Parser definitions are extracted from synthetic-key symbol table entries, concept references are resolved, function calls are looked up in the `FunctionRegistry` (only `sum` is registered in this increment), and `CILAggregateNode` CIL trees are generated with pre-built aggregation matrices.

**FunctionRegistry (Increment 2 — sum only)**

Aggregator functions are NOT reserved words. The parser stores function calls as `OperandDefinition` with kind `"fncall"` containing a name and parameter list. The analysis phase looks up the function name in a `FunctionRegistry` that maps names to predefined CIL node types.

```go
// AggregatorFunctionInfo defines a registered aggregator function.
type AggregatorFunctionInfo struct {
    Name        string                           // Function name (e.g., "sum")
    CILNodeType string                           // CIL node type identifier (e.g., "CILAggregateSumNode")
    DimUnits    func(params []string) *DimensionalUnitExpression // Dimensional unit effect
    Description string                           // Human-readable description
}

// FunctionRegistry maps function names to their predefined definitions.
// The registry is populated during analysis initialization.
type FunctionRegistry struct {
    Functions map[string]*AggregatorFunctionInfo
}
```

**In Increment 2, only `sum` is registered:**

```go
func NewAggregatorFunctionRegistry() *FunctionRegistry {
    reg := &FunctionRegistry{
        Functions: make(map[string]*AggregatorFunctionInfo),
    }

    // Increment 2: only sum is supported
    reg.Functions["sum"] = &AggregatorFunctionInfo{
        Name:        "sum",
        CILNodeType: "CILAggregateNode", // Generic node in Increment 2; becomes CILAggregateSumNode in Increment 3
        DimUnits: func(params []string) *DimensionalUnitExpression {
            // sum(C) has dim units C (the concept's dim units are the result)
            if len(params) != 1 {
                return nil // Invalid — handled by validation
            }
            return &DimensionalUnitExpression{
                Numerator: []string{params[0]},
            }
        },
        Description: "Sums values across a concept dimension",
    }

    // count, size, 1/sum, 1/count are added in Increment 3

    return reg
}
```

**AnalysisDefinitions — Aggregators Lookup Map**

```go
// AnalysisDefinitions contains all analyzed definitions for CIL generation.
type AnalysisDefinitions struct {
    Concepts    map[string]*AnalysisConceptDefinition
    Types       map[string]*AnalysisTypeDefinition
    Units       map[string]*AnalysisUnitDefinition
    Measures    map[string]*AnalysisMeasureDefinition
    Connectors  map[string]*AnalysisConnectorDefinition
    Metrics     map[string]*AnalysisMetricDefinition
    Lookups     map[string]*LookupInfo
    Aggregators map[string]*AggregatorInfo   // NEW: Aggregators lookup map
    CILTrees    map[string]CILNode            // CIL trees keyed by name
}
```

The Aggregators map uses a **compound string key** derived from the `(from_concept_set, to_concept)` pattern:

```go
// AggregatorMapKey constructs the compound key for the Aggregators map.
// The key is deterministic: concept names are sorted to ensure
// [Country, Product]->[Product] and [Product, Country]->[Product] produce the same key.
func AggregatorMapKey(fromConcepts []string, toConcept string) string {
    sorted := make([]string, len(fromConcepts))
    copy(sorted, fromConcepts)
    sort.Strings(sorted)
    if toConcept == "" {
        return fmt.Sprintf("[%s]->[]", strings.Join(sorted, ","))
    }
    return fmt.Sprintf("[%s]->[%s]", strings.Join(sorted, ","), toConcept)
}
```

**AggregatorInfo (Increment 2)**

```go
// AggregatorInfo contains the analyzed aggregator information.
// Concept references are resolved, and the dimensional unit effect is validated.
type AggregatorInfo struct {
    FromConcepts         []*AnalysisConceptDefinition  // Resolved from concept references
    ToConcept            *AnalysisConceptDefinition     // Resolved to concept reference (nil for scalar)
    ToEmpty              bool                           // True if reducing to scalar (to [])
    FromDimensionalUnits DimensionalUnitExpression      // Computed dim units of input
    ToDimensionalUnits   DimensionalUnitExpression      // Computed dim units of output
    ExpressionDimUnits   DimensionalUnitExpression      // Validated dim unit effect of expression
    CellOperation        string                         // The aggregation operation ("sum", "count", "size", etc.)
    SourceLocation       SourceLocation
}
```

**DimensionalUnitExpression (Increment 2)**

```go
// DimensionalUnitExpression represents a symbolic dimensional unit.
// It uses the same algebra as regular unit/measure conversions (*, /, ^ only).
// Examples: "Product/Country", "Country", "Unitless", "1/Country".
type DimensionalUnitExpression struct {
    Numerator   []string  // Concept names in the numerator (sorted for canonical form)
    Denominator []string  // Concept names in the denominator (sorted for canonical form)
}

// IsUnitless returns true if both numerator and denominator are empty.
func (d *DimensionalUnitExpression) IsUnitless() bool {
    return len(d.Numerator) == 0 && len(d.Denominator) == 0
}

// Multiply returns a new DimensionalUnitExpression representing this * other.
func (d *DimensionalUnitExpression) Multiply(other *DimensionalUnitExpression) *DimensionalUnitExpression {
    result := &DimensionalUnitExpression{
        Numerator:   make([]string, len(d.Numerator)),
        Denominator: make([]string, len(d.Denominator)),
    }
    copy(result.Numerator, d.Numerator)
    copy(result.Denominator, d.Denominator)

    // Add other's numerator to this numerator
    result.Numerator = append(result.Numerator, other.Numerator...)

    // Add other's denominator to this denominator
    result.Denominator = append(result.Denominator, other.Denominator...)

    // Cancel common factors
    result.cancel()
    return result
}

// cancel removes elements that appear in both numerator and denominator.
func (d *DimensionalUnitExpression) cancel() {
    // Build a multiset of denominator elements
    denomCounts := make(map[string]int)
    for _, c := range d.Denominator {
        denomCounts[c]++
    }

    // Remove from numerator any elements that appear in denominator
    var newNum []string
    for _, c := range d.Numerator {
        if denomCounts[c] > 0 {
            denomCounts[c]--
        } else {
            newNum = append(newNum, c)
        }
    }
    d.Numerator = newNum

    // Remove from denominator any elements that appear in numerator
    // (rebuilding from the remaining counts)
    var newDen []string
    for c, count := range denomCounts {
        for i := 0; i < count; i++ {
            newDen = append(newDen, c)
        }
    }
    d.Denominator = newDen

    // Sort for canonical form
    sort.Strings(d.Numerator)
    sort.Strings(d.Denominator)
}
```

**Compute Dimensional Units from Concept Set (Increment 2)**

```go
// dimensionalUnitsFromConceptSet computes the dimensional units for a metric
// with the given concept dimensions.
// For dimensions [C1, C2, ..., Cn], the dim units are C1*C2*...*Cn / (none).
// Wait — that's wrong for aggregator semantics. Let's reconsider.
//
// For a metric with dimensions [Country, Product], the dimensional units are
// Product/Country (one time series per product, many country values per product).
// The denominator accumulates all but the last concept, the numerator is the last.
//
// More precisely: for dimensions [D1, D2, ..., Dn], the dimensional units are
// Dn / (D1 * D2 * ... * D(n-1)).
// For a single dimension [Product], the dimensional units are Product.
func dimensionalUnitsFromConceptSet(concepts []*AnalysisConceptDefinition) *DimensionalUnitExpression {
    n := len(concepts)
    if n == 0 {
        // Empty dimensions — scalar metric, unitless
        return &DimensionalUnitExpression{}
    }

    names := make([]string, n)
    for i, c := range concepts {
        names[i] = c.Name
    }

    if n == 1 {
        return &DimensionalUnitExpression{
            Numerator: []string{names[0]},
        }
    }

    // All but the last go in denominator; the last goes in numerator
    return &DimensionalUnitExpression{
        Numerator:   []string{names[n-1]},
        Denominator: names[:n-1],
    }
}
```

**Expression Evaluation for Aggregators (Increment 2)**

Aggregator expressions use the same postfix evaluation as metric expressions. Function calls (`fncall` operands) are resolved against the FunctionRegistry and push their dimensional unit effect onto the expression stack. Arithmetic operators (`*`, `/`) combine dimensional units using the same algebra.

```go
// evaluateAggregatorExpression evaluates the postfix expression sequence
// and computes the dimensional unit effect.
// The postfix sequence is extracted from the aggregator scope's Definitions slice.
func (a *CPLAnalyzer) evaluateAggregatorExpression(
    definitions []*SymbolDefinition,
    fromConcepts []*AnalysisConceptDefinition,
) (*DimensionalUnitExpression, string, error) {

    var stack []*DimensionalUnitExpression
    var cellOp string

    for _, def := range definitions {
        operand, ok := def.Data.(*OperandData)
        if !ok {
            return nil, "", fmt.Errorf("unexpected definition type in aggregator expression")
        }

        switch operand.Kind {
        case "concept":
            // Concept reference: push its dimensional unit
            stack = append(stack, &DimensionalUnitExpression{
                Numerator: []string{operand.Name},
            })

        case "fncall":
            // Look up the function in the registry
            fnInfo, exists := a.functionRegistry.Functions[operand.Name]
            if !exists {
                return nil, "", fmt.Errorf(
                    "unknown aggregator function %q (available: sum)",
                    operand.Name,
                )
            }

            // Compute the function's dimensional unit effect
            dimUnits := fnInfo.DimUnits(operand.Params)
            if dimUnits == nil {
                return nil, "", fmt.Errorf(
                    "invalid parameters for function %q: %v",
                    operand.Name, operand.Params,
                )
            }

            // The function call pushes its result onto the stack
            stack = append(stack, dimUnits)
            cellOp = operand.Name // "sum" in Increment 2

        case "operator":
            // Arithmetic operators combine dimensional units
            switch operand.Name {
            case "*":
                if len(stack) < 2 {
                    return nil, "", fmt.Errorf("insufficient operands for *")
                }
                right := stack[len(stack)-1]
                left := stack[len(stack)-2]
                stack = stack[:len(stack)-2]
                stack = append(stack, left.Multiply(right))

            case "/":
                if len(stack) < 2 {
                    return nil, "", fmt.Errorf("insufficient operands for /")
                }
                right := stack[len(stack)-1]
                left := stack[len(stack)-2]
                stack = stack[:len(stack)-2]
                inverted := &DimensionalUnitExpression{
                    Numerator:   right.Denominator,
                    Denominator: right.Numerator,
                }
                stack = append(stack, left.Multiply(inverted))

            default:
                return nil, "", fmt.Errorf("unknown operator %q", operand.Name)
            }
        }
    }

    if len(stack) != 1 {
        return nil, "", fmt.Errorf("expression did not reduce to a single dimensional unit")
    }

    return stack[0], cellOp, nil
}
```

Note: In this model, `(a-b)*sum(Country)` works naturally — the concept refs `a` and `b` push their dim units, `-` processes them, `sum(Country)` pushes `Country`, and `*` combines them. The function call is just another operand on the expression stack, not a special operator form.

**Dimensional Unit Validation (Increment 2)**

```go
// validateAggregatorDimensionalUnits verifies that the aggregator's
// dimensional unit effect is compatible with the from → to transformation.
func (a *CPLAnalyzer) validateAggregatorDimensionalUnits(info *AggregatorInfo) error {
    // Compute input dimensional units from the from concept set
    info.FromDimensionalUnits = dimensionalUnitsFromConceptSet(info.FromConcepts)

    // Compute output dimensional units from the to concept set
    if info.ToEmpty {
        info.ToDimensionalUnits = &DimensionalUnitExpression{} // scalar = unitless
    } else {
        info.ToDimensionalUnits = dimensionalUnitsFromConceptSet(
            []*AnalysisConceptDefinition{info.ToConcept},
        )
    }

    // The expression's dimensional unit effect is already computed
    // from postfix evaluation.

    // Validation: input_units * expression_units = output_units
    predicted := info.FromDimensionalUnits.Multiply(info.ExpressionDimUnits)

    if !predicted.Equals(info.ToDimensionalUnits) {
        return fmt.Errorf(
            "aggregator dimensional unit mismatch: input %s * expression %s = %s, but output requires %s",
            info.FromDimensionalUnits.String(),
            info.ExpressionDimUnits.String(),
            predicted.String(),
            info.ToDimensionalUnits.String(),
        )
    }

    return nil
}
```

**CILAggregateNode (Increment 2 — generic)**

In Increment 2, a single generic `CILAggregateNode` is used for all aggregation operations (only `sum` is supported). In Increment 3, this is replaced with operation-specific CIL node types (`CILAggregateSumNode`, etc.) for cleaner synthesis code generation.

```go
// CILAggregateNode represents the application of an aggregator to a metric.
// In Increment 2, it is generic — the CellOperation string determines behavior.
// In Increment 3, it is replaced by operation-specific node types.
type CILAggregateNode struct {
    SourceCIL             CILNode                    // Reference to CIL subtree for source metric
    FromDimensions        []string                   // Concept dimensions in the source
    ToDimension           string                     // Concept being collapsed ("" for scalar)
    AggregationMatrix     *AggregationMatrix         // Pre-computed transformation matrix
    CellOperation         string                     // "sum" in Increment 2; will become node-specific in Increment 3
    DimensionalUnitEffect string                     // String representation of the dim unit effect
    SourceLocation        SourceLocation
}

func (n *CILAggregateNode) cilNode() {}
```

**AggregationMatrix (Increment 2)**

```go
// AggregationMatrix is a 2D weight matrix used for tensor dimension reduction.
// Rows correspond to values of the concept being collapsed.
// Columns correspond to values of the concept being preserved (or 1 column for scalar).
// For sum(Country): matrix[m][n] = 1.0 for each country m → product n mapping.
type AggregationMatrix struct {
    Rows       int           // Number of source concept values being collapsed
    Columns    int           // Number of target concept values (1 for scalar)
    Data       [][]float64   // Matrix data [row][column]
    RowLabels  []string      // Concept values for rows (e.g., ["USA", "Canada"])
    ColLabels  []string      // Concept values for columns (e.g., ["widget", "gadget"])
}

// BuildSumAggregationMatrix creates an aggregation matrix for sum(Concept).
// For each source value, the weight to the target value is 1.0
// (representing the value being included in the sum).
func BuildSumAggregationMatrix(
    collapsedValues []string,     // Values of the concept being collapsed
    targetValues []string,        // Values of the concept being preserved (nil for scalar)
) *AggregationMatrix {
    rows := len(collapsedValues)
    cols := 1
    if len(targetValues) > 0 {
        cols = len(targetValues)
    }

    data := make([][]float64, rows)
    for i := range data {
        data[i] = make([]float64, cols)
        for j := range data[i] {
            data[i][j] = 1.0 // Each source value contributes fully to each target
        }
    }

    return &AggregationMatrix{
        Rows:      rows,
        Columns:   cols,
        Data:      data,
        RowLabels: collapsedValues,
        ColLabels: targetValues,
    }
}

// BuildSizeAggregationMatrix creates an aggregation matrix for size(Concept).
// size() is unitless — each entry is 1.0 (counting each value once).
func BuildSizeAggregationMatrix(collapsedValues []string) *AggregationMatrix {
    rows := len(collapsedValues)
    data := make([][]float64, rows)
    for i := range data {
        data[i] = []float64{1.0}
    }
    return &AggregationMatrix{
        Rows:      rows,
        Columns:   1,
        Data:      data,
        RowLabels: collapsedValues,
        ColLabels: []string{"count"},
    }
}

// BuildInverseSumAggregationMatrix creates a matrix for 1/sum(Concept).
// Each entry is 1.0 / len(collapsedValues), representing the inverse contribution.
func BuildInverseSumAggregationMatrix(collapsedValues []string, targetValues []string) *AggregationMatrix {
    rows := len(collapsedValues)
    cols := 1
    if len(targetValues) > 0 {
        cols = len(targetValues)
    }

    weight := 1.0 / float64(rows)

    data := make([][]float64, rows)
    for i := range data {
        data[i] = make([]float64, cols)
        for j := range data[i] {
            data[i][j] = weight
        }
    }

    return &AggregationMatrix{
        Rows:      rows,
        Columns:   cols,
        Data:      data,
        RowLabels: collapsedValues,
        ColLabels: targetValues,
    }
}
```

**Analysis Function — Aggregator Map Assembly (Increment 2)**

```go
// AnalyzeAggregatorScope processes an aggregator scope from the symbol table
// and builds an AggregatorInfo entry with validated dimensional unit effect.
func (a *CPLAnalyzer) AnalyzeAggregatorScope(table *SymbolTable) error {
    // Find the AggregatorScopeDefinition entry
    var parserDef *ParserAggregatorDefinition
    for _, def := range table.Definitions {
        if def.DefinitionType == AggregatorScopeDefinition {
            var ok bool
            parserDef, ok = def.Data.(*ParserAggregatorDefinition)
            if !ok {
                return fmt.Errorf("invalid aggregator definition data for %s", *def.Name)
            }
            break
        }
    }
    if parserDef == nil {
        return fmt.Errorf("no AggregatorScopeDefinition found in aggregator scope")
    }

    // Resolve from concept names
    var fromConcepts []*AnalysisConceptDefinition
    for _, name := range parserDef.FromConcepts {
        conceptDef, exists := a.concepts[name]
        if !exists {
            return fmt.Errorf("aggregator: from concept %q not found", name)
        }
        fromConcepts = append(fromConcepts, conceptDef)
    }

    // Resolve to concept name
    var toConcept *AnalysisConceptDefinition
    toEmpty := parserDef.ToEmpty
    if !toEmpty {
        var exists bool
        toConcept, exists = a.concepts[parserDef.ToConcept]
        if !exists {
            return fmt.Errorf("aggregator: to concept %q not found", parserDef.ToConcept)
        }

        // Validate: to concept must be in from list
        found := false
        for _, c := range fromConcepts {
            if c.Name == parserDef.ToConcept {
                found = true
                break
            }
        }
        if !found {
            return fmt.Errorf("aggregator: to concept %q is not in from list %v",
                parserDef.ToConcept, parserDef.FromConcepts)
        }
    } else {
        // to []: from must have exactly one concept
        if len(fromConcepts) != 1 {
            return fmt.Errorf("aggregator: to [] requires exactly one concept in from list, got %d: %v",
                len(fromConcepts), parserDef.FromConcepts)
        }
    }

    // Evaluate expression to compute dimensional unit effect
    exprUnits, cellOp, err := a.evaluateAggregatorExpression(table.Definitions, fromConcepts)
    if err != nil {
        return fmt.Errorf("aggregator: %w", err)
    }

    // Build AggregatorInfo
    info := &AggregatorInfo{
        FromConcepts:       fromConcepts,
        ToConcept:          toConcept,
        ToEmpty:            toEmpty,
        ExpressionDimUnits: exprUnits,
        CellOperation:      cellOp,
        SourceLocation:     parserDef.SourceLocation,
    }

    // Validate dimensional unit compatibility
    if err := a.validateAggregatorDimensionalUnits(info); err != nil {
        return fmt.Errorf("aggregator: %w", err)
    }

    // Store in Aggregators map under compound key
    key := AggregatorMapKey(parserDef.FromConcepts, parserDef.ToConcept)
    if _, exists := a.aggregators[key]; exists {
        // Multiple aggregators with the same pattern are allowed if they have
        // different expressions; the specificity check in Increment 5 handles
        // disambiguation. For Increment 2, warn but allow.
        // (Ambigiousity check deferred to Increment 5)
    }
    a.aggregators[key] = info

    // Generate CIL node with pre-built aggregation matrix
    var collapsedValues []string
    for _, c := range fromConcepts {
        if toEmpty || c.Name == parserDef.ToConcept {
            collapsedValues = c.ValueOrdering // Use the concept's ordered values
        }
    }

    targetValues := []string{}
    if !toEmpty {
        targetValues = toConcept.ValueOrdering
    }

    var matrix *AggregationMatrix
    switch cellOp {
    case "sum", "count":
        matrix = BuildSumAggregationMatrix(collapsedValues, targetValues)
    case "size":
        matrix = BuildSizeAggregationMatrix(collapsedValues)
    case "1/sum", "1/count":
        matrix = BuildInverseSumAggregationMatrix(collapsedValues, targetValues)
    default:
        // Complex expression — matrix construction deferred to Increment 5
        matrix = &AggregationMatrix{} // Placeholder
    }

    cilNode := &CILAggregateNode{
        FromDimensions:        parserDef.FromConcepts,
        ToDimension:           parserDef.ToConcept,
        AggregationMatrix:     matrix,
        CellOperation:         cellOp,
        DimensionalUnitEffect: exprUnits.String(),
        SourceLocation:        parserDef.SourceLocation,
    }

    // Store CIL tree for CIL generation (not yet inserted into metric CIL trees)
    a.cilTrees["aggregator:"+key] = cilNode

    return nil
}
```

**Validation During Analysis (Increment 2)**

- **Concept Existence**: All concept names in `from`, `to`, and expression must refer to existing concepts
- **From-To Overlap**: `to` concept must be in `from` list (except when `to` is `[]`)
- **From-To Size**: When `to` is `[]`, `from` must have exactly one concept
- **Dimensional Unit Compatibility**: `input_units * expression_units = output_units`
- **Expression Operands**: All concept references in the expression must be to concepts in the `from` list
- **Duplicate Aggregators**: Multiple aggregators with the same pattern are allowed (disambiguation deferred to Increment 5)

**Limitations (Increment 2)**

- Only the `sum` function is supported in the FunctionRegistry
- Aggregators are validated but not yet applied to metric calculations (Increment 4)
- Aggregation matrices are built but not yet inserted into metric CIL trees
- Operation-specific CIL node types (`CILAggregateSumNode`, etc.) not yet defined (Increment 3)
- Cell unit validation deferred to Increment 4 (requires metric cell type information)

### Increment 3: Additional Aggregator Function CIL Node Types

**Overview**

Increment 3 defines operation-specific CIL node types for the remaining aggregator functions (`count`, `size`, `1/sum`, `1/count`) and replaces the generic `CILAggregateNode` from Increment 2 with typed nodes. The FunctionRegistry is extended to register all functions, each associated with its corresponding CIL node type.

Because aggregator functions fit into expressions (e.g., `(a-b)*sum(Country)`), each CIL node type behaves as an expression operand — it produces a value that participates in arithmetic operations. The synthesis phase generates code that computes the node's value (via matrix multiplication) and stores it in a local variable, which is then used as an operand in the surrounding expression.

**FunctionRegistry — Full Registration (Increment 3)**

```go
func NewAggregatorFunctionRegistry() *FunctionRegistry {
    reg := &FunctionRegistry{
        Functions: make(map[string]*AggregatorFunctionInfo),
    }

    reg.Functions["sum"] = &AggregatorFunctionInfo{
        Name:        "sum",
        CILNodeType: "CILAggregateSumNode",
        DimUnits: func(params []string) *DimensionalUnitExpression {
            if len(params) != 1 { return nil }
            return &DimensionalUnitExpression{Numerator: []string{params[0]}}
        },
        Description: "Sums values across a concept dimension",
    }

    reg.Functions["count"] = &AggregatorFunctionInfo{
        Name:        "count",
        CILNodeType: "CILAggregateCountNode",
        DimUnits: func(params []string) *DimensionalUnitExpression {
            if len(params) != 1 { return nil }
            // count(C) has same dim units as sum(C) — the concept appears in numerator
            return &DimensionalUnitExpression{Numerator: []string{params[0]}}
        },
        Description: "Counts the number of values in a concept dimension",
    }

    reg.Functions["size"] = &AggregatorFunctionInfo{
        Name:        "size",
        CILNodeType: "CILAggregateSizeNode",
        DimUnits: func(params []string) *DimensionalUnitExpression {
            if len(params) != 1 { return nil }
            // size(C) is unitless — it's a count of values, not a dimensional unit
            return &DimensionalUnitExpression{}
        },
        Description: "Returns the number of distinct values in a concept",
    }

    // 1/sum and 1/count are recognized as compound function names
    reg.Functions["1/sum"] = &AggregatorFunctionInfo{
        Name:        "1/sum",
        CILNodeType: "CILAggregateInverseSumNode",
        DimUnits: func(params []string) *DimensionalUnitExpression {
            if len(params) != 1 { return nil }
            // 1/sum(C) inverts: C → 1/C
            return &DimensionalUnitExpression{Denominator: []string{params[0]}}
        },
        Description: "Inverse of sum across a concept dimension",
    }

    reg.Functions["1/count"] = &AggregatorFunctionInfo{
        Name:        "1/count",
        CILNodeType: "CILAggregateInverseCountNode",
        DimUnits: func(params []string) *DimensionalUnitExpression {
            if len(params) != 1 { return nil }
            return &DimensionalUnitExpression{Denominator: []string{params[0]}}
        },
        Description: "Inverse of count across a concept dimension",
    }

    return reg
}
```

**CIL Node Types (Increment 3)**

Each function gets its own CIL node type. These replace the generic `CILAggregateNode` from Increment 2. Each node type is a `CILNode` that wraps a source metric and carries an aggregation matrix and dimensional metadata:

```go
// CILAggregateSumNode applies sum aggregation to a metric.
// In the expression (a-b)*sum(Country), this node computes sum(Country)
// and the value participates in arithmetic operations.
type CILAggregateSumNode struct {
    SourceCIL   CILNode              // CIL subtree for the source metric/dimension
    DimName     string               // Name of the concept dimension being summed
    Matrix      *AggregationMatrix   // Pre-computed aggregation matrix
    SourceLocation SourceLocation
}
func (n *CILAggregateSumNode) cilNode() {}

// CILAggregateCountNode applies count aggregation to a metric.
type CILAggregateCountNode struct {
    SourceCIL   CILNode
    DimName     string
    Matrix      *AggregationMatrix
    SourceLocation SourceLocation
}
func (n *CILAggregateCountNode) cilNode() {}

// CILAggregateSizeNode returns the size (number of values) of a concept dimension.
// size(C) is unitless — it produces a scalar value from the concept's value count.
type CILAggregateSizeNode struct {
    DimName     string               // Name of the concept dimension
    Size        int                  // Number of distinct values in the concept
    SourceLocation SourceLocation
}
func (n *CILAggregateSizeNode) cilNode() {}

// CILAggregateInverseSumNode applies 1/sum aggregation.
type CILAggregateInverseSumNode struct {
    SourceCIL   CILNode
    DimName     string
    Matrix      *AggregationMatrix   // Pre-computed inverse sum matrix
    SourceLocation SourceLocation
}
func (n *CILAggregateInverseSumNode) cilNode() {}

// CILAggregateInverseCountNode applies 1/count aggregation.
type CILAggregateInverseCountNode struct {
    SourceCIL   CILNode
    DimName     string
    Matrix      *AggregationMatrix   // Pre-computed inverse count matrix
    SourceLocation SourceLocation
}
func (n *CILAggregateInverseCountNode) cilNode() {}
```

**CIL Node Insertion During Expression Analysis (Increment 3)**

When the analysis phase encounters an `fncall` operand during expression evaluation, it looks up the function in the registry and creates the corresponding CIL node type:

```go
// resolveAggregatorFnCall creates the appropriate CIL node for a function call.
func (a *CPLAnalyzer) resolveAggregatorFnCall(
    fnCall *OperandData,           // The fncall from parser
    sourceCIL CILNode,             // The CIL subtree for the source (concept or metric)
) (CILNode, error) {

    fnInfo, exists := a.functionRegistry.Functions[fnCall.Name]
    if !exists {
        return nil, fmt.Errorf("unknown aggregator function %q", fnCall.Name)
    }

    dimName := ""
    if len(fnCall.Params) > 0 {
        dimName = fnCall.Params[0]
    }

    switch fnInfo.CILNodeType {
    case "CILAggregateSumNode":
        matrix := BuildSumAggregationMatrix(
            a.conceptValues(dimName),
            a.targetConceptValues(dimName),
        )
        return &CILAggregateSumNode{
            SourceCIL:      sourceCIL,
            DimName:        dimName,
            Matrix:         matrix,
        }, nil

    case "CILAggregateCountNode":
        matrix := BuildSumAggregationMatrix(
            a.conceptValues(dimName),
            a.targetConceptValues(dimName),
        )
        return &CILAggregateCountNode{
            SourceCIL:      sourceCIL,
            DimName:        dimName,
            Matrix:         matrix,
        }, nil

    case "CILAggregateSizeNode":
        values := a.conceptValues(dimName)
        return &CILAggregateSizeNode{
            DimName:        dimName,
            Size:           len(values),
        }, nil

    case "CILAggregateInverseSumNode":
        matrix := BuildInverseSumAggregationMatrix(
            a.conceptValues(dimName),
            a.targetConceptValues(dimName),
        )
        return &CILAggregateInverseSumNode{
            SourceCIL:      sourceCIL,
            DimName:        dimName,
            Matrix:         matrix,
        }, nil

    case "CILAggregateInverseCountNode":
        matrix := BuildInverseSumAggregationMatrix(
            a.conceptValues(dimName),
            a.targetConceptValues(dimName),
        )
        return &CILAggregateInverseCountNode{
            SourceCIL:      sourceCIL,
            DimName:        dimName,
            Matrix:         matrix,
        }, nil

    default:
        return nil, fmt.Errorf("unsupported CIL node type %q", fnInfo.CILNodeType)
    }
}
```

**Example: `(a-b)*sum(Country)` CIL Tree Construction (Increment 3)**

For the aggregator expression `(a-b)*sum(Country)`:
1. Parser produces postfix: `[concept "a", concept "b", operator "-", concept "Country", fncall "sum" ["Country"], operator "*"]`
2. Analysis evaluates postfix:
   - Push `a` dim unit, push `b` dim unit
   - `-` operator pops `a` and `b`, validates same units, pushes result
   - Push `Country` concept (discovers it's consumed by the fncall)
   - `fncall "sum" ["Country"]` — lookup in registry returns `CILAggregateSumNode`; push `Country` dim unit to stack
   - `*` operator combines dim units
3. CIL tree: `CILBinaryOp("*", CILBinaryOp("-", a_ref, b_ref), CILAggregateSumNode{source, "Country", matrix})`

**Limitations (Increment 3)**

- The `size(C)` function produces a scalar value from concept metadata (no tensor source) — synthesis must handle this case where there is no SourceCIL

### Increment 4: Aggregator Selection and Integration

**Overview**

**Overview**

Increment 4 integrates aggregators into the metric dependency DAG analysis. After all metric definitions are assembled, the analyzer detects dimension mismatches between metric inputs and outputs, queries the Aggregators map, and inserts the appropriate CIL aggregate node into metric CIL trees for single unambiguous sequences.

**Dimension Mismatch Detection (Increment 4)**

```go
// resolveMetricDimensionality checks a metric's inputs against its output dimensions,
// and applies aggregators/spreaders as needed to bridge dimensional gaps.
// Called during the second pass of metric analysis after the full DAG is assembled.
func (a *CPLAnalyzer) resolveMetricDimensionality(metric *AnalysisMetricDefinition) error {
    // Determine the metric's required output dimensions from IO spec
    outputDims := a.getMetricOutputDimensions(metric)

    for _, input := range metric.Inputs {
        if input.InputKind != "metric" {
            continue // Dimension and lookup inputs don't have dimension adjustments here
        }

        // Get the input metric's dimensions
        inputDims := input.ResolvedMetric.Dimensions

        // Find dimensions in input that are not in output
        var excessDims []string
        for _, dim := range inputDims {
            if !contains(outputDims, dim) {
                excessDims = append(excessDims, dim)
            }
        }

        if len(excessDims) == 0 {
            continue // No dimension mismatch
        }

        // For Increment 4: only handle single unambiguous sequences
        sequence, err := a.findSingleAggregatorSequence(
            inputDims,
            outputDims,
            excessDims,
        )
        if err != nil {
            return fmt.Errorf("metric %s: %w", metric.Name, err)
        }

        // Create a chain of CILAggregateNode for the input
        currentCIL := input.CILTree
        for _, aggKey := range sequence {
            aggInfo := a.aggregators[aggKey]
            aggNode := a.buildAggregateNode(aggInfo, currentCIL)
            currentCIL = aggNode
        }

        // Replace the input's CIL reference with the aggregated chain
        input.CILTree = currentCIL
    }

    return nil
}

// findSingleAggregatorSequence finds the unique sequence of aggregators
// that reduces inputDims to outputDims.
// In Increment 4, only cases with exactly one valid sequence are supported.
// If zero or multiple sequences exist, an error is returned.
func (a *CPLAnalyzer) findSingleAggregatorSequence(
    inputDims []string,
    outputDims []string,
    excessDims []string,
) ([]string, error) {

    // For a single excess dimension, check if an aggregator exists
    if len(excessDims) == 1 {
        key := AggregatorMapKey(inputDims, excessDims[0])
        if _, exists := a.aggregators[key]; exists {
            return []string{key}, nil
        }
        return nil, fmt.Errorf(
            "no aggregator found to reduce %v to %v (excess: %s)",
            inputDims, outputDims, excessDims[0],
        )
    }

    // For multiple excess dimensions, this is deferred to Increment 5
    return nil, fmt.Errorf(
        "multiple excess dimensions %v require sequence enumeration (deferred to Increment 5)",
        excessDims,
    )
}
```

**CILAggregateNode Insertion (Increment 4)**

```go
// buildAggregateNode creates a CILAggregateNode that wraps an input CIL tree.
func (a *CPLAnalyzer) buildAggregateNode(info *AggregatorInfo, sourceCIL CILNode) *CILAggregateNode {
    // Build the aggregation matrix for this specific usage
    // (matrices are pre-built during Increment 2 but may need context-specific adjustments)
    collapsedValues := a.getCollapsedConceptValues(info)
    targetValues := a.getTargetConceptValues(info)

    var matrix *AggregationMatrix
    switch info.CellOperation {
    case "sum", "count":
        matrix = BuildSumAggregationMatrix(collapsedValues, targetValues)
    case "size":
        matrix = BuildSizeAggregationMatrix(collapsedValues)
    case "1/sum", "1/count":
        matrix = BuildInverseSumAggregationMatrix(collapsedValues, targetValues)
    }

    return &CILAggregateNode{
        SourceCIL:             sourceCIL,
        FromDimensions:        a.getConceptNames(info.FromConcepts),
        ToDimension:           a.getToConceptName(info),
        AggregationMatrix:     matrix,
        CellOperation:         info.CellOperation,
        DimensionalUnitEffect: info.ExpressionDimUnits.String(),
    }
}
```

**Validation During Analysis (Increment 4 — additions)**

- **No Applicable Aggregator**: If a dimension mismatch exists but no aggregator bridges the gap, compile-time error
- **Multiple Sequences**: In Increment 4, multiple possible sequences are a compile-time error (user must simplify)

### Increment 5: Aggregator Sequence Optimization

**Overview**

Increment 5 implements the full aggregator sequence selection algorithm and optimizations. When multiple aggregator sequences could accomplish the required dimensional transformation, the framework selects the optimal one.

**Sequence Enumeration (Increment 5)**

```go
// enumerateAggregatorSequences finds all valid sequences of aggregators
// that reduce from fromDims to toDims, ordered by specificity.
func (a *CPLAnalyzer) enumerateAggregatorSequences(
    fromDims []string,
    toDims []string,
) ([][]string, error) {
    // Find dimensions to remove
    toRemove := setDifference(fromDims, toDims)

    if len(toRemove) == 0 {
        return nil, nil // No reduction needed
    }

    var sequences [][]string
    a.enumerateSequencesRecursive(fromDims, toDims, toRemove, []string{}, &sequences)

    if len(sequences) == 0 {
        return nil, fmt.Errorf(
            "no aggregator sequence found to reduce %v to %v", fromDims, toDims,
        )
    }

    return sequences, nil
}

func (a *CPLAnalyzer) enumerateSequencesRecursive(
    currentDims []string,
    targetDims []string,
    remainingToRemove []string,
    currentSequence []string,
    results *[][]string,
) {
    for i, concept := range remainingToRemove {
        // Check if there's an aggregator that collapses this concept from currentDims
        key := AggregatorMapKey(currentDims, concept)
        agg, exists := a.aggregators[key]
        if !exists {
            continue
        }

        // Compute new dimensions after applying this aggregator
        newDims := removeElement(currentDims, concept)

        // New sequence
        newSequence := make([]string, len(currentSequence)+1)
        copy(newSequence, currentSequence)
        newSequence[len(currentSequence)] = key

        // New remaining concepts to remove
        newRemaining := removeElement(remainingToRemove, concept)

        // Recurse or store result
        if len(newRemaining) == 0 && isSubset(targetDims, newDims) {
            *results = append(*results, newSequence)
        } else if len(newRemaining) > 0 {
            a.enumerateSequencesRecursive(newDims, targetDims, newRemaining, newSequence, results)
        }

        _ = agg // used in specificity computation below
    }
}
```

**Selection Algorithm (Increment 5)**

```go
// selectOptimalSequence chooses the best sequence using the three-tier criteria:
// 1. Least-information-loss: prefer aggregating dimensions with fewer distinct values first
// 2. Minimal-work: prefer sequences with fewer total aggregations
// 3. Deterministic tiebreaker: lexicographic by concept name
func (a *CPLAnalyzer) selectOptimalSequence(sequences [][]string) []string {
    if len(sequences) == 0 {
        return nil
    }
    if len(sequences) == 1 {
        return sequences[0]
    }

    // Sort by tier-1: least information loss
    // (fewer distinct values in collapsed dimension = less info loss)
    sort.Slice(sequences, func(i, j int) bool {
        // Tier 1: Information loss
        lossI := a.sequenceInformationLoss(sequences[i])
        lossJ := a.sequenceInformationLoss(sequences[j])
        if lossI != lossJ {
            return lossI < lossJ
        }

        // Tier 2: Minimal work (shorter sequence = fewer operations)
        if len(sequences[i]) != len(sequences[j]) {
            return len(sequences[i]) < len(sequences[j])
        }

        // Tier 3: Deterministic tiebreaker
        for k := 0; k < len(sequences[i]); k++ {
            if sequences[i][k] != sequences[j][k] {
                return sequences[i][k] < sequences[j][k]
            }
        }
        return false
    })

    return sequences[0]
}

// sequenceInformationLoss computes the total information loss for a sequence.
// Each step's loss is the number of distinct values in the dimension being collapsed.
// Lower total = less information lost.
func (a *CPLAnalyzer) sequenceInformationLoss(sequence []string) int {
    total := 0
    for _, key := range sequence {
        agg := a.aggregators[key]
        collapsedConcept := a.getCollapsedConcept(agg)
        total += len(collapsedConcept.ValueOrdering)
    }
    return total
}
```

**Ambiguous Aggregator Resolution (Increment 5)**

```go
// resolveAmbiguousAggregators checks for and resolves patterns with multiple aggregators.
// If two aggregators share the same (from, to) key, specificity determines which to use.
// If specificity is equal, this is an error.
func (a *CPLAnalyzer) resolveAmbiguousAggregators() error {
    // Group aggregators by key
    groups := make(map[string][]*AggregatorInfo)
    for key, info := range a.aggregators {
        groups[key] = append(groups[key], info)
    }

    for key, infos := range groups {
        if len(infos) == 1 {
            continue
        }

        // Multiple aggregators with same pattern — check specificity
        // Specificity = number of concepts in from list (more = more specific)
        maxSpecificity := 0
        var mostSpecific *AggregatorInfo
        tie := false

        for _, info := range infos {
            specificity := len(info.FromConcepts)
            if specificity > maxSpecificity {
                maxSpecificity = specificity
                mostSpecific = info
                tie = false
            } else if specificity == maxSpecificity {
                tie = true
            }
        }

        if tie {
            return fmt.Errorf(
                "ambiguous aggregators for pattern %s: multiple aggregators with specificity %d",
                key, maxSpecificity,
            )
        }

        // Replace map entry with the most specific one
        a.aggregators[key] = mostSpecific
    }

    return nil
}
```

**Fused Operations (Increment 5 — Synthesis)**

When multiple aggregators are applied in sequence, adjacent aggregation matrices can be combined into a single matrix multiplication:

```go
// fuseAggregationMatrices combines two adjacent aggregation matrices
// into a single matrix, eliminating an intermediate tensor.
// Matrix A (M×N) × Matrix B (N×P) → Matrix C (M×P)
func fuseAggregationMatrices(a, b *AggregationMatrix) *AggregationMatrix {
    m := a.Rows
    n := a.Columns // must equal b.Rows
    p := b.Columns

    data := make([][]float64, m)
    for i := 0; i < m; i++ {
        data[i] = make([]float64, p)
        for j := 0; j < p; j++ {
            sum := 0.0
            for k := 0; k < n; k++ {
                sum += a.Data[i][k] * b.Data[k][j]
            }
            data[i][j] = sum
        }
    }

    return &AggregationMatrix{
        Rows:    m,
        Columns: p,
        Data:    data,
        // Row labels from a, column labels from b
    }
}
```

### Notes

**Related Documents:**

- [ERO007 — Aggregators Incremental Implementation](../../eros/ERO007_aggregators-incremental-implementation.md) — Parent ERO
- [ERO007_aggregators-parsing-phase.md](../../go-parser-target/eros/ERO007_aggregators-parsing-phase.md) — Parsing phase structures
- [ERO007_aggregators-synthesis-phase.md](../../go-synthesis-target/eros/ERO007_aggregators-synthesis-phase.md) — Synthesis phase
- [Aggregators Design](../../aggregators.md) — Full aggregators design
- [Spreaders Design](../../spreaders.md) — Spreader design (symmetric inverse)
- [Metrics Design](../../metrics-design.md) — Metric design (aggregators applied during DAG analysis)
- [Concept Definitions Design](../../concepts-definitions.md) — Concept design (aggregator matrices from concept values)
- [Units and Measures Design](../../units-and-measures.md) — Unit/measure design (dimensional unit algebra)
- [Postfix Expression Parsing Pattern](../../../../patterns/postfix-expression-parsing.md) — Expression handling
- [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md) — Phase separation
