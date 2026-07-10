# ERO005: Metrics Analysis Phase

## Title

Metrics Analysis Phase — Data Structures and Incremental Implementation

## Purpose/Goal

Define the analysis phase data structures for metric definitions (including `AnalysisMetricDefinition`, postfix-to-CIL expression tree construction, input resolution, and dependency DAG building) and document how these structures and analysis functions evolve across each increment of the metrics incremental implementation as defined in [ERO005 — Metrics Incremental Implementation](../../eros/ERO005_metrics-incremental-implementation.md).

## Implementation Status

**Analysis Phase Data Structures — Implemented**

All analysis phase structures and functions for metric definitions through Increment 8 are implemented and tested. The implementation includes:

- Increments 1–7: `AnalysisMetricDefinition`, `AnalysisInputDefinition`, CIL expression tree construction (literals, binary/unary ops, conditionals, input refs, time shifts, aliases, where clauses, short-form resolution), postfix evaluation, input reference resolution.
- Increment 8: `MetricDAG`, `DAGNode`, `BuildMetricDependencyDAG`, `TopologicalSort` (Kahn's algorithm), `DetectCycles` (DFS with time-shifted self-reference exemption), `MarkReachable`, `EliminateDeadCode`, and full IO specification validation (dimensions, celltype, grain, connector existence, driver parameter path).

**Related Epics:**

- Epic005_cpl-metrics-implementation.md

## Context

### Situation

The analysis phase is responsible for transforming parser-level metric definitions into analysis-level structures, validating all cross-references, resolving input dependencies, and building CIL expression trees from the postfix definition sequences produced by the parser. Following the [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md), the parser captures raw syntactic information in the symbol table; the analysis phase walks the symbol table tree to assemble semantically validated structures.

The parsing phase ERO ([ERO005_metrics-parsing-phase.md](../../go-parser-target/eros/ERO005_metrics-parsing-phase.md)) defines how `ParserMetricDefinition` structures and postfix expression sequences are stored in the symbol table. This ERO defines how those raw structures are transformed into `AnalysisMetricDefinition` entries with resolved references and CIL expression trees.

**Key Principle:** The analysis phase walks the symbol table tree, extracts parser definitions, validates semantics, evaluates postfix sequences into CIL trees, and builds the metric dependency DAG.

### Constraints

- Must work with parser-level metric definitions from the symbol table
- Must follow the [Postfix Expression Parsing Pattern](../../../../patterns/postfix-expression-parsing.md) for expression tree construction
- Must follow the [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md) for phase separation
- Must produce structures usable by the CIL generator and synthesizer
- Must validate all metric, dimension, and lookup references
- Must support the incremental feature set of each increment
- Expression tree construction must handle conditionals (if/else/end if, when/end when) via child scopes in the symbol table tree — each branch body gets its own child scope with its own postfix sequence; conditions use the same postfix mechanism as arithmetic
- Dependency DAG construction is deferred to a second pass after all definitions are assembled

### Requirements

- Build a unified Metrics lookup map from parser definitions
- Perform postfix evaluation to construct CIL expression trees from the Definitions slice in each `MetricScope`
- Resolve input references (metric names, dimension names, lookup names) against the complete definition set
- Validate alias uniqueness within each metric's input block
- Validate where-clause aspect references
- Validate time-shift constraint field references
- Resolve short-form metric names (metric → dimension → lookup resolution)
- Build the metric dependency DAG in a second pass after all definitions across all sources are assembled
- Detect circular dependencies (excluding valid self-references with time shifting)
- Accept and validate full IO specification (output metric existence) starting in Increment 2
- Perform dead-code elimination (only synthesize reachable metrics) in Increment 8
- Support full IO specification integration (dimensional/temporal configuration) in Increment 8

### Assumptions

- Parser metric definitions are complete and syntactically valid
- The analysis phase has access to all parsed files in the compilation unit
- Concepts, types, units, measures, and connectors are already analyzed and available in `AnalysisDefinitions`
- Analysis errors prevent CIL generation and synthesis
- The `WalkSymbolTableWithHandlers` pattern from concepts/types/units/connectors analysis is available

## Data Structures by Increment

### Increment 1: Basic Metric Declarations with Identity Calculation

**Overview**

Increment 1 introduces the minimal metric analysis infrastructure. The parser produces `ParserMetricDefinition` structures with simple `metric`/`dimension` input references and a calculate body containing only the literal `0`. The analysis phase walks the symbol table, extracts these definitions, validates that input names refer to declared metrics and dimensions, performs postfix evaluation of the literal `0` to produce a single-node `CILLiteral` CIL tree, and creates one single-node CIL tree per connector.

**AnalysisDefinitions — Metrics Lookup Map**

```go
// AnalysisDefinitions contains all analyzed definitions for CIL generation.
type AnalysisDefinitions struct {
    Concepts   map[string]*AnalysisConceptDefinition
    Types      map[string]*AnalysisTypeDefinition
    Units      map[string]*AnalysisUnitDefinition
    Measures   map[string]*AnalysisMeasureDefinition
    Connectors map[string]*AnalysisConnectorDefinition
    Metrics    map[string]*AnalysisMetricDefinition  // NEW: Metrics lookup map
    CILTrees   map[string]CILNode                    // CIL trees keyed by name (metrics + connectors)
}
```

**AnalysisMetricDefinition (Increment 1)**

```go
// AnalysisMetricDefinition contains the analyzed metric information.
// Input references are resolved to their corresponding definitions.
// The CIL tree is constructed from postfix evaluation of the calculate expression.
type AnalysisMetricDefinition struct {
    Name           string                          // Metric identifier
    Inputs         []AnalysisInputDefinition        // Resolved input references
    CILTree        CILNode                         // CIL expression tree for calculate body
    SourceLocation SourceLocation                   // For error reporting
}
```

**AnalysisInputDefinition (Increment 1)**

```go
// AnalysisInputDefinition represents a resolved metric input reference.
type AnalysisInputDefinition struct {
    InputKind      string                    // "metric" or "dimension"
    Name           string                    // Referenced name
    ResolvedMetric *AnalysisMetricDefinition // Resolved reference (nil for dimensions)
    SourceLocation SourceLocation
}
```

**CIL Node Types (Increment 1)**

```go
// CILNode is the interface for all CIL tree nodes.
type CILNode interface {
    cilNode() // marker method
}

// CILLiteral represents a literal value in the CIL tree.
type CILLiteral struct {
    Value string  // "0" in Increment 1
}

// CILConnectorReferenceNode represents a leaf data source (connector).
type CILConnectorReferenceNode struct {
    ConnectorName string
}
```

**Metrics Analysis Function**

The metrics analysis function is registered with `WalkSymbolTableWithHandlers` and called for each `MetricScope` encountered during the walk:

```go
// AnalyzeMetricScope processes a MetricScope symbol table and builds
// an AnalysisMetricDefinition with resolved inputs and CIL tree.
func (a *CPLAnalyzer) AnalyzeMetricScope(table *SymbolTable) error {
    // Find the MetricScopeDefinition entry that carries the ParserMetricDefinition
    var parserDef *ParserMetricDefinition
    for _, def := range table.Definitions {
        if def.DefinitionType == MetricScopeDefinition {
            var ok bool
            parserDef, ok = def.Data.(*ParserMetricDefinition)
            if !ok {
                return fmt.Errorf("invalid metric definition data for %s", *def.Name)
            }
            break
        }
    }
    if parserDef == nil {
        return fmt.Errorf("no MetricScopeDefinition found in metric scope")
    }

    // Check for duplicate metric names
    if _, exists := a.metrics[parserDef.Name]; exists {
        return fmt.Errorf("duplicate metric definition: %s", parserDef.Name)
    }

    // Resolve input references
    inputs, err := a.resolveMetricInputs(parserDef.Inputs)
    if err != nil {
        return fmt.Errorf("failed to resolve inputs for metric '%s': %w", parserDef.Name, err)
    }

    // Build CIL tree from postfix definitions in the MetricScope
    cilTree, err := a.buildCILTreeFromScope(table)
    if err != nil {
        return fmt.Errorf("failed to build CIL tree for metric '%s': %w", parserDef.Name, err)
    }

    metricDef := &AnalysisMetricDefinition{
        Name:           parserDef.Name,
        Inputs:         inputs,
        CILTree:        cilTree,
        SourceLocation: parserDef.SourceLocation,
    }

    a.metrics[parserDef.Name] = metricDef
    a.cilTrees[parserDef.Name] = cilTree

    return nil
}
```

**Postfix Evaluation for Increment 1**

In Increment 1, the calculate body contains only the literal `0`. The postfix evaluation is trivial:

```go
// buildCILTreeFromScope evaluates the postfix definition sequence
// in a MetricScope to construct a CIL expression tree.
func (a *CPLAnalyzer) buildCILTreeFromScope(table *SymbolTable) (CILNode, error) {
    // Filter definitions that are part of the calculate expression
    // (exclude InputRefDefinition and MetricScopeDefinition entries)
    var postfix []SymbolDefinition
    for _, def := range table.Definitions {
        switch def.DefinitionType {
        case LiteralDefinition, NumericLiteralDefinition:
            postfix = append(postfix, def)
        case InputRefDefinition, MetricScopeDefinition:
            continue
        }
    }

    // Stack-based postfix evaluation
    stack := make([]CILNode, 0)

    for _, def := range postfix {
        switch def.DefinitionType {
        case LiteralDefinition, NumericLiteralDefinition:
            litData, ok := def.Data.(*LiteralData)
            if !ok {
                return nil, fmt.Errorf("invalid literal data")
            }
            stack = append(stack, &CILLiteral{Value: litData.Value})
        default:
            return nil, fmt.Errorf("unexpected definition type in postfix: %v", def.DefinitionType)
        }
    }

    if len(stack) != 1 {
        return nil, fmt.Errorf("expected exactly one CIL node after postfix evaluation, got %d", len(stack))
    }

    return stack[0], nil
}
```

**Input Resolution for Increment 1**

```go
// resolveMetricInputs validates input references against declared metrics and dimensions.
func (a *CPLAnalyzer) resolveMetricInputs(parserInputs []ParserInputDefinition) ([]AnalysisInputDefinition, error) {
    var resolved []AnalysisInputDefinition

    for _, input := range parserInputs {
        resolvedInput := AnalysisInputDefinition{
            InputKind:      input.InputKind,
            Name:           input.Name,
            SourceLocation: input.NameLocation,
        }

        switch input.InputKind {
        case "metric":
            if _, exists := a.metrics[input.Name]; !exists {
                return nil, fmt.Errorf("metric input references unknown metric '%s'", input.Name)
            }
            // ResolvedMetric populated in second pass (DAG construction)
        case "dimension":
            if _, exists := a.concepts[input.Name]; !exists {
                return nil, fmt.Errorf("metric input references unknown dimension '%s'", input.Name)
            }
        default:
            return nil, fmt.Errorf("unknown input kind: %s", input.InputKind)
        }

        resolved = append(resolved, resolvedInput)
    }

    return resolved, nil
}
```

**Integration with WalkSymbolTableWithHandlers**

```go
func (c *CPLCompiler) Analyze(rootTable *SymbolTable) (*AnalysisDefinitions, error) {
    analyzer := NewCPLAnalyzer()

    handlers := map[ScopeType]AnalysisFunc{
        ConceptsScope:   analyzer.AnalyzeConceptsScope,
        TypesScope:      analyzer.AnalyzeTypesScope,
        UnitsScope:      analyzer.AnalyzeUnitsScope,
        MeasuresScope:   analyzer.AnalyzeMeasuresScope,
        ConnectorScope:  analyzer.AnalyzeConnectorsScope,
        MetricScope:     analyzer.AnalyzeMetricScope,     // NEW
    }

    if err := WalkSymbolTableWithHandlers(rootTable, handlers); err != nil {
        return nil, fmt.Errorf("analysis failed: %w", err)
    }

    // Second pass: build metric dependency DAG
    if err := analyzer.BuildMetricDependencyDAG(); err != nil {
        return nil, fmt.Errorf("failed to build metric dependency DAG: %w", err)
    }

    return analyzer.GetAnalysisDefinitions(), nil
}
```

**Symbol Table Structure Example**

Given this CPL source:

```cpl
metrics
    revenue
        inputs
            metric demand_forecast
            dimension Price
        calculate
            0
```

The parser produces:

```
GlobalScope (root)
└── MetricsScope
    └── MetricScope (Name: "revenue")
        ├── SymbolDefinition: Name="demand_forecast", DefinitionType=InputRefDefinition
        │   Data: &ParserInputDefinition{InputKind: "metric", Name: "demand_forecast"}
        ├── SymbolDefinition: Name="Price", DefinitionType=InputRefDefinition
        │   Data: &ParserInputDefinition{InputKind: "dimension", Name: "Price"}
        ├── SymbolDefinition: Name="0", DefinitionType=LiteralDefinition
        │   Data: &LiteralData{Value: "0"}
        └── SymbolDefinition: Name="revenue", DefinitionType=MetricScopeDefinition
            Data: &ParserMetricDefinition{Name: "revenue", Inputs: [...]}
```

After analysis:

```
AnalysisDefinitions.Metrics["revenue"] = &AnalysisMetricDefinition{
    Name: "revenue",
    Inputs: []AnalysisInputDefinition{
        {InputKind: "metric", Name: "demand_forecast"},
        {InputKind: "dimension", Name: "Price"},
    },
    CILTree: &CILLiteral{Value: "0"},
}
```

**Validation During Analysis**

- Metric name uniqueness across the compilation unit
- Input `metric` references must refer to declared metrics
- Input `dimension` references must refer to existing concepts
- Input kind must be `metric` or `dimension` (no `lookup` in Increment 1)
- Calculate expression must produce exactly one CIL node

**Limitations**

- No arithmetic expressions (added in Increment 2)
- No conditional expressions (added in Increment 3)
- No `as` alias resolution (added in Increment 4)
- No `where` clause analysis (added in Increment 4)
- No `backward`/`forward` time-shift analysis (added in Increment 5)
- No short-form metric name resolution (added in Increment 6)
- No dependency DAG construction beyond trivial (all metrics are leaves with literal `0`)
- No dead-code elimination (added in Increment 7)

### Increment 2: Simple Arithmetic in Calculate

**Overview**

Increment 2 extends the analysis phase to handle arithmetic expressions in the `calculate` clause. The parser produces postfix sequences containing operand definitions (identifiers, literals) and operator definitions (`PlusOperatorDefinition`, `MinusOperatorDefinition`, `MultiplyOperatorDefinition`, `DivideOperatorDefinition`). The analysis phase extends the postfix evaluator to construct `CILBinaryOp` nodes and `CILInputRef` nodes.

**New CIL Node Types**

```go
// CILInputRef represents a reference to an input metric or dimension.
type CILInputRef struct {
    Alias string  // The alias or name used in the expression
}

// CILBinaryOp represents a binary operation in the CIL tree.
type CILBinaryOp struct {
    Operator string   // "+", "-", "*", "/"
    Left     CILNode
    Right    CILNode
}

// CILUnaryOp represents a unary negation operation.
type CILUnaryOp struct {
    Operator string  // "-" (negation)
    Operand  CILNode
}
```

**Extended Postfix Evaluation (Increment 2)**

```go
func (a *CPLAnalyzer) buildCILTreeFromScope(table *SymbolTable) (CILNode, error) {
    var postfix []SymbolDefinition
    for _, def := range table.Definitions {
        switch def.DefinitionType {
        case LiteralDefinition, NumericLiteralDefinition, StringLiteralDefinition:
            postfix = append(postfix, def)
        case InputRefDefinition:
            postfix = append(postfix, def)
        case PlusOperatorDefinition, MinusOperatorDefinition,
             MultiplyOperatorDefinition, DivideOperatorDefinition:
            postfix = append(postfix, def)
        case NegateOperatorDefinition:
            postfix = append(postfix, def)
        case MetricScopeDefinition:
            continue
        }
    }

    stack := make([]CILNode, 0)

    for _, def := range postfix {
        switch def.DefinitionType {
        case LiteralDefinition, NumericLiteralDefinition:
            litData := def.Data.(*LiteralData)
            stack = append(stack, &CILLiteral{Value: litData.Value})

        case StringLiteralDefinition:
            strData := def.Data.(*StringLiteralData)
            stack = append(stack, &CILLiteral{Value: strData.Value})

        case InputRefDefinition:
            refData := def.Data.(*InputRefData)
            stack = append(stack, &CILInputRef{Alias: refData.Name})

        case PlusOperatorDefinition:
            right, left := a.popTwo(stack)
            stack = append(stack, &CILBinaryOp{Operator: "+", Left: left, Right: right})

        case MinusOperatorDefinition:
            right, left := a.popTwo(stack)
            stack = append(stack, &CILBinaryOp{Operator: "-", Left: left, Right: right})

        case MultiplyOperatorDefinition:
            right, left := a.popTwo(stack)
            stack = append(stack, &CILBinaryOp{Operator: "*", Left: left, Right: right})

        case DivideOperatorDefinition:
            right, left := a.popTwo(stack)
            stack = append(stack, &CILBinaryOp{Operator: "/", Left: left, Right: right})

        case NegateOperatorDefinition:
            operand := a.popOne(stack)
            stack = append(stack, &CILUnaryOp{Operator: "-", Operand: operand})
        }
    }

    if len(stack) != 1 {
        return nil, fmt.Errorf("expected exactly one CIL node, got %d", len(stack))
    }
    return stack[0], nil
}
```

**Postfix Evaluation Example**

For the expression `demand_forecast * Price + 10`, the parser produces Definitions: `[demand_forecast, Price, MULT, 10, PLUS]`. Postfix evaluation:

1. Push `CILInputRef{Alias: "demand_forecast"}`
2. Push `CILInputRef{Alias: "Price"}`
3. Pop two → Push `CILBinaryOp{Operator: "*", Left: demand_forecast, Right: Price}`
4. Push `CILLiteral{Value: "10"}`
5. Pop two → Push `CILBinaryOp{Operator: "+", Left: (demand_forecast * Price), Right: 10}`

**Validation During Analysis**

- Arithmetic operators must have two operands (enforced by postfix evaluation)
- Input references in calculate must be valid identifiers that appear in the metric's `inputs` block
- Type-checking of arithmetic operands (numeric types only for `+`, `-`, `*`, `/`)

**IO Specification Acceptance**

Starting in Increment 2, the analysis phase accepts a full IO specification (see [io_specs.md](../../../../requirements/cpl/io_specs.md)) as input. The IO spec specifies output metrics, their dimensions, celltype, grain, connector-to-data-source mappings, and compute architecture. In this increment, the analysis phase:

1. Validates that all output metric names specified in the IO spec exist in the analyzed definitions
2. Stores the IO spec for use by the synthesis phase (dimensional/temporal configuration, compute architecture)

```go
// IOSpecification is accepted by the analyzer starting in Increment 2.
type IOSpecification struct {
    Outputs      []OutputSpec
    Connectors   []ConnectorSpec
    Architecture ArchitectureSpec
}

type OutputSpec struct {
    MetricName string
    Dimensions []string  // Concept names
    CellType   string    // Measure name
    Grain      []string  // Grain measure names (ordered)
}

// ApplyIOSpec validates output metrics and stores the IO spec for synthesis.
func (a *CPLAnalyzer) ApplyIOSpec(ioSpec *IOSpecification) error {
    for _, output := range ioSpec.Outputs {
        if _, exists := a.metrics[output.MetricName]; !exists {
            return fmt.Errorf("IO spec references unknown metric: %s", output.MetricName)
        }
    }
    a.ioSpec = ioSpec
    return nil
}
```

Note: The dimensional/temporal configuration and connector mappings in the IO spec are consumed by the synthesis phase. The analysis phase validates only output metric existence in this increment. Dead-code elimination and dependency DAG construction are deferred to Increment 8. The compute architecture specification (CPU cores, cache size, memory) is consumed by the tensor parallelism increment (Increment 3, synthesis-only).

**Limitations**

- No `Nil` literal (added in Increment 4)
- No boolean literals (added in Increment 4)
- No conditional expressions (added in Increment 4)
- No comparison or logical operators (added in Increment 4)

### Increment 3: Tensor Calculation Parallelism

**Overview**

Increment 3 introduces parallelism in tensor calculations. This increment is **synthesis-only** — there are no changes to the analysis phase.

The synthesis phase uses the compute architecture specification from the IO spec (CPU cores, cache size, main memory size) to partition tensor computation work across multiple goroutines or CPU cores. The analysis phase has already validated and stored the IO spec in Increment 2; the synthesis phase reads the stored IO spec to determine the parallelism configuration.

**Analysis Changes**

None. The analysis phase requires no new data structures or functions for this increment. The IO spec is already accepted and stored in Increment 2. All parallelism work is handled in the synthesis phase.

### Increment 4: Conditional Expressions

**Overview**

Increment 3 extends the analysis phase to handle conditional expressions (`if`/`else if`/`else`/`end if` and `when`/`default`/`end when`), comparison operators (`==`, `!=`, `<`, `>`, `<=`, `>=`), logical operators (`and`, `or`, `not`), `Nil` literal, and boolean literals (`True`, `False`).

**Scope-Based Conditional Structure**

Conditionals do NOT use flat marker definitions in the postfix sequence. Instead, the symbol table tree structure encodes the conditional hierarchy:

- **Condition expressions** use the same infix-to-postfix mechanism as arithmetic. Comparison operators (`==`, `!=`, `<`, `>`, `<=`, `>=`) and logical operators (`and`, `or`, `not`) are pushed as regular operator definitions in the parent `MetricScope`'s `Definitions` slice. The condition's postfix is captured in a temporary scope during parsing and stored as `Data` on a clause-defining `SymbolDefinition` (e.g., `IfClauseDefinition`, `WhenClauseDefinition`) in the parent scope.

- **Branch bodies** each get their own child scope. The parser pushes a new child `SymbolTable` for each branch body (`IfBodyScope`, `ElseIfBodyScope`, `ElseBodyScope`, `WhenBodyScope`, `DefaultBodyScope`). The body's expression postfix lives in that child scope's `Definitions` slice.

For example, given:

```cpl
revenue
    inputs
        metric demand_forecast
        dimension Price
    calculate
        if demand_forecast > 0
            demand_forecast * Price
        else
            0
        end if
```

The parser produces:

```
MetricScope (Name: "revenue")
├── SymbolDefinition: Name="demand_forecast", DefinitionType=InputRefDefinition
├── SymbolDefinition: Name="Price", DefinitionType=InputRefDefinition
├── SymbolDefinition: DefinitionType=IfClauseDefinition
│   Data: &ClauseData{ConditionPostfix: [demand_forecast, 0, >]}
├── IfBodyScope (child)
│   └── Definitions: [demand_forecast, Price, *]
├── SymbolDefinition: DefinitionType=ElseClauseDefinition
├── ElseBodyScope (child)
│   └── Definitions: [0]
└── SymbolDefinition: Name="revenue", DefinitionType=MetricScopeDefinition
    Data: &ParserMetricDefinition{...}
```

**New Scope and Definition Types**

```go
const (
    // Conditional body scope types
    IfBodyScope       ScopeType = iota  // Scope for if branch body
    ElseIfBodyScope                      // Scope for else-if branch body
    ElseBodyScope                        // Scope for else branch body
    WhenBodyScope                        // Scope for when clause body
    DefaultBodyScope                     // Scope for default clause body

    // Clause definition types (stored in parent scope's Definitions)
    IfClauseDefinition     DefinitionType = iota  // Carries condition postfix
    ElseIfClauseDefinition                        // Carries condition postfix
    ElseClauseDefinition                          // No condition (else has no condition)
    WhenClauseDefinition                          // Carries condition postfix
    DefaultClauseDefinition                       // No condition
)

// ClauseData carries the condition's postfix definitions for a conditional clause.
type ClauseData struct {
    ConditionPostfix []SymbolDefinition  // Condition postfix (nil for else/default)
}
```

**New CIL Node Types**

```go
// CILConditional represents an if/else if/else/end if expression.
type CILConditional struct {
    IfBranch   CILConditionalBranch
    ElseIfs    []CILConditionalBranch
    ElseBranch CILNode  // nil if no else clause
}

type CILConditionalBranch struct {
    Condition CILNode
    Body      CILNode
}

// CILWhenClause represents a when/default/end when expression.
type CILWhenClause struct {
    Clauses       []CILWhenBranch
    DefaultBranch CILNode
}

type CILWhenBranch struct {
    Condition CILNode
    Body      CILNode
}

// CILComparisonOp represents a comparison operation.
type CILComparisonOp struct {
    Operator string   // "==", "!=", "<", ">", "<=", ">="
    Left     CILNode
    Right    CILNode
}

// CILLogicalOp represents a logical operation.
type CILLogicalOp struct {
    Operator string   // "and", "or"
    Left     CILNode
    Right    CILNode
}

// CILNotOp represents a logical NOT operation.
type CILNotOp struct {
    Operand CILNode
}

// CILNilLiteral represents the Nil literal.
type CILNilLiteral struct{}

// CILBoolLiteral represents a boolean literal.
type CILBoolLiteral struct {
    Value bool  // true for True, false for False
}
```

**Extended Postfix Evaluation (Increment 3)**

Comparison operators, logical operators, `Nil`, and boolean literals extend the same postfix evaluator used for arithmetic. The `buildCILTreeFromScope` function from Increment 2 is extended with additional definition type cases:

```go
func (a *CPLAnalyzer) buildCILTreeFromScope(table *SymbolTable) (CILNode, error) {
    var postfix []SymbolDefinition
    for _, def := range table.Definitions {
        switch def.DefinitionType {
        case LiteralDefinition, NumericLiteralDefinition, StringLiteralDefinition:
            postfix = append(postfix, def)
        case NilLiteralDefinition, TrueLiteralDefinition, FalseLiteralDefinition:
            postfix = append(postfix, def)
        case InputRefDefinition:
            postfix = append(postfix, def)
        case PlusOperatorDefinition, MinusOperatorDefinition,
             MultiplyOperatorDefinition, DivideOperatorDefinition,
             NegateOperatorDefinition:
            postfix = append(postfix, def)
        case EqOperatorDefinition, NotEqOperatorDefinition,
             LessThanOperatorDefinition, GreaterThanOperatorDefinition,
             LessThanEqOperatorDefinition, GreaterThanEqOperatorDefinition:
            postfix = append(postfix, def)
        case AndOperatorDefinition, OrOperatorDefinition,
             NotOperatorDefinition:
            postfix = append(postfix, def)
        case MetricScopeDefinition, IfClauseDefinition, ElseIfClauseDefinition,
             ElseClauseDefinition, WhenClauseDefinition, DefaultClauseDefinition:
            continue // Structural definitions, not expression components
        }
    }

    stack := make([]CILNode, 0)

    for _, def := range postfix {
        switch def.DefinitionType {
        // ... existing cases from Increment 2 ...

        case NilLiteralDefinition:
            stack = append(stack, &CILNilLiteral{})

        case TrueLiteralDefinition:
            stack = append(stack, &CILBoolLiteral{Value: true})

        case FalseLiteralDefinition:
            stack = append(stack, &CILBoolLiteral{Value: false})

        case EqOperatorDefinition:
            right, left := a.popTwo(stack)
            stack = append(stack, &CILComparisonOp{Operator: "==", Left: left, Right: right})

        case NotEqOperatorDefinition:
            right, left := a.popTwo(stack)
            stack = append(stack, &CILComparisonOp{Operator: "!=", Left: left, Right: right})

        case LessThanOperatorDefinition:
            right, left := a.popTwo(stack)
            stack = append(stack, &CILComparisonOp{Operator: "<", Left: left, Right: right})

        case GreaterThanOperatorDefinition:
            right, left := a.popTwo(stack)
            stack = append(stack, &CILComparisonOp{Operator: ">", Left: left, Right: right})

        case LessThanEqOperatorDefinition:
            right, left := a.popTwo(stack)
            stack = append(stack, &CILComparisonOp{Operator: "<=", Left: left, Right: right})

        case GreaterThanEqOperatorDefinition:
            right, left := a.popTwo(stack)
            stack = append(stack, &CILComparisonOp{Operator: ">=", Left: left, Right: right})

        case AndOperatorDefinition:
            right, left := a.popTwo(stack)
            stack = append(stack, &CILLogicalOp{Operator: "and", Left: left, Right: right})

        case OrOperatorDefinition:
            right, left := a.popTwo(stack)
            stack = append(stack, &CILLogicalOp{Operator: "or", Left: left, Right: right})

        case NotOperatorDefinition:
            operand := a.popOne(stack)
            stack = append(stack, &CILNotOp{Operand: operand})
        }
    }

    if len(stack) != 1 {
        return nil, fmt.Errorf("expected exactly one CIL node, got %d", len(stack))
    }
    return stack[0], nil
}
```

**Conditional Assembly Algorithm**

The `AnalyzeMetricScope` function is extended to walk child scopes and assemble `CILConditional` or `CILWhenClause` nodes. The algorithm:

1. Scans the parent `MetricScope`'s `Definitions` for clause definitions (`IfClauseDefinition`, `ElseIfClauseDefinition`, `ElseClauseDefinition`, `WhenClauseDefinition`, `DefaultClauseDefinition`)
2. Each clause definition is paired with its corresponding child body scope in order
3. The condition postfix (from `ClauseData.ConditionPostfix`) is evaluated using the standard postfix evaluator
4. The body scope's `Definitions` are evaluated using the standard postfix evaluator
5. Results are assembled into `CILConditional` or `CILWhenClause`

```go
// buildConditionalFromScope walks the MetricScope's children to assemble
// a CILConditional or CILWhenClause from clause definitions and body scopes.
func (a *CPLAnalyzer) buildConditionalFromScope(table *SymbolTable) (CILNode, error) {
    // Collect clause definitions in order from the parent scope's Definitions
    type clauseEntry struct {
        defType   DefinitionType
        clauseData *ClauseData
        bodyScope  *SymbolTable
    }
    var clauses []clauseEntry

    childIdx := 0
    for _, def := range table.Definitions {
        switch def.DefinitionType {
        case IfClauseDefinition, ElseIfClauseDefinition, ElseClauseDefinition,
             WhenClauseDefinition, DefaultClauseDefinition:
            entry := clauseEntry{
                defType:    def.DefinitionType,
                clauseData: def.Data.(*ClauseData),
            }
            // The corresponding body scope is the next child of the appropriate type
            if childIdx < len(table.Children) {
                entry.bodyScope = table.Children[childIdx]
                childIdx++
            }
            clauses = append(clauses, entry)
        }
    }

    if len(clauses) == 0 {
        return nil, fmt.Errorf("no conditional clauses found")
    }

    // Determine if this is an if/else or when construct
    switch clauses[0].defType {
    case IfClauseDefinition:
        return a.assembleIfConditional(clauses)
    case WhenClauseDefinition:
        return a.assembleWhenClause(clauses)
    default:
        return nil, fmt.Errorf("unexpected first clause type: %v", clauses[0].defType)
    }
}

func (a *CPLAnalyzer) assembleIfConditional(clauses []clauseEntry) (*CILConditional, error) {
    // First clause must be IfClauseDefinition
    condition, err := a.evaluateSimplePostfix(clauses[0].clauseData.ConditionPostfix)
    if err != nil {
        return nil, fmt.Errorf("failed to evaluate if condition: %w", err)
    }
    body, err := a.buildCILTreeFromScope(clauses[0].bodyScope)
    if err != nil {
        return nil, fmt.Errorf("failed to evaluate if body: %w", err)
    }

    result := &CILConditional{
        IfBranch: CILConditionalBranch{Condition: condition, Body: body},
    }

    // Process else-if clauses
    i := 1
    for i < len(clauses) && clauses[i].defType == ElseIfClauseDefinition {
        condition, err := a.evaluateSimplePostfix(clauses[i].clauseData.ConditionPostfix)
        if err != nil {
            return nil, fmt.Errorf("failed to evaluate else-if condition: %w", err)
        }
        body, err := a.buildCILTreeFromScope(clauses[i].bodyScope)
        if err != nil {
            return nil, fmt.Errorf("failed to evaluate else-if body: %w", err)
        }
        result.ElseIfs = append(result.ElseIfs, CILConditionalBranch{Condition: condition, Body: body})
        i++
    }

    // Process else clause (if present)
    if i < len(clauses) && clauses[i].defType == ElseClauseDefinition {
        body, err := a.buildCILTreeFromScope(clauses[i].bodyScope)
        if err != nil {
            return nil, fmt.Errorf("failed to evaluate else body: %w", err)
        }
        result.ElseBranch = body
        i++
    }

    return result, nil
}

func (a *CPLAnalyzer) assembleWhenClause(clauses []clauseEntry) (*CILWhenClause, error) {
    result := &CILWhenClause{}

    for _, clause := range clauses {
        switch clause.defType {
        case WhenClauseDefinition:
            condition, err := a.evaluateSimplePostfix(clause.clauseData.ConditionPostfix)
            if err != nil {
                return nil, fmt.Errorf("failed to evaluate when condition: %w", err)
            }
            body, err := a.buildCILTreeFromScope(clause.bodyScope)
            if err != nil {
                return nil, fmt.Errorf("failed to evaluate when body: %w", err)
            }
            result.Clauses = append(result.Clauses, CILWhenBranch{Condition: condition, Body: body})

        case DefaultClauseDefinition:
            body, err := a.buildCILTreeFromScope(clause.bodyScope)
            if err != nil {
                return nil, fmt.Errorf("failed to evaluate default body: %w", err)
            }
            result.DefaultBranch = body
        }
    }

    return result, nil
}
```

**Truthiness Rules Validation**

The analysis phase validates that boolean contexts respect CPL truthiness rules:

- `Nil` is falsy; `False` is falsy
- Numeric zero is falsy; non-zero is truthy
- Empty string is falsy; non-empty string is truthy
- All other non-Nil values are truthy

**Nil-in-Arithmetic Validation**

The analysis phase detects and rejects any use of `Nil` as an operand of arithmetic operators (`+`, `-`, `*`, `/`). This is a compile-time error.

**Expression Completeness Validation**

Conditional expressions must be fully specified — every possible path through `if`/`else if`/`else` must produce a result. An `if` without an `else` requires an `else` clause since exhaustive conditions cannot be determined at analysis time.

**Validation During Analysis**

- `if`/`else if`/`else`/`end if` pairing is enforced by grammar (ANTLR); analysis validates clause/body scope correspondence
- `when` must have at least one clause and a `default` clause
- Truthiness rules for Nil, zero, empty strings in boolean contexts
- Nil cannot participate in arithmetic
- Conditional expressions must be fully specified (all branches produce a result)
- Comparison and logical operators must have valid operands

**Limitations**

- Truthiness of complex expressions (involving metric references) is validated structurally but not fully type-checked until the dependency DAG is built
- No `lookup` input kind support (deferred to future design)

### Increment 5: Input Aliases and Where Clauses

**Overview**

Increment 5 extends input resolution to handle `as` aliases and `where` clauses with `aspect=ACTUALS`/`aspect=FORECAST` filtering and boolean combinations (`and`, `or`, `not`).

**Enhanced AnalysisInputDefinition (Increment 5)**

```go
type AnalysisInputDefinition struct {
    InputKind      string                    // "metric" or "dimension"
    Name           string                    // Referenced name
    Alias          string                    // Alias from 'as' clause (equals Name if no alias)
    WhereClause    *AnalysisWhereClause      // Where clause (nil if not present)
    ResolvedMetric *AnalysisMetricDefinition
    SourceLocation SourceLocation
}

type AnalysisWhereClause struct {
    Conditions []AnalysisWhereCondition
    Operators  []string  // "and" or "or" joining conditions
}

type AnalysisWhereCondition struct {
    Negated  bool
    AspectEq *AnalysisAspectEquals
    SubExpr  *AnalysisWhereClause
}

type AnalysisAspectEquals struct {
    Aspect string  // "ACTUALS" or "FORECAST"
}
```

**Input Resolution with Aliases**

```go
func (a *CPLAnalyzer) resolveMetricInputs(parserInputs []ParserInputDefinition, metricName string) ([]AnalysisInputDefinition, error) {
    var resolved []AnalysisInputDefinition
    aliases := make(map[string]bool)

    for _, input := range parserInputs {
        alias := input.Name
        if input.Alias != "" {
            alias = input.Alias
        }

        if aliases[alias] {
            return nil, fmt.Errorf("duplicate alias '%s' in metric '%s'", alias, metricName)
        }
        aliases[alias] = true

        resolvedInput := AnalysisInputDefinition{
            InputKind:      input.InputKind,
            Name:           input.Name,
            Alias:          alias,
            SourceLocation: input.NameLocation,
        }

        if input.WhereClause != nil {
            whereClause, err := a.resolveWhereClause(input.WhereClause)
            if err != nil {
                return nil, fmt.Errorf("invalid where clause for input '%s' in metric '%s': %w",
                    input.Name, metricName, err)
            }
            resolvedInput.WhereClause = whereClause
        }

        resolved = append(resolved, resolvedInput)
    }

    return resolved, nil
}
```

**Validation During Analysis**

- Alias uniqueness within a single metric's input block
- Where clause aspect values must be `ACTUALS` or `FORECAST`
- Boolean combinations in where clauses must be well-formed
- Input references in calculate expressions use aliases when present

### Increment 6: Time Shifting (Backward/Forward)

**Overview**

Increment 6 extends input resolution to handle `backward()` and `forward()` time-shift clauses. The analysis phase validates that constraint fields refer to valid Instant fields (`year`, `month`, `day`, `hour`, `minute`, `second`, `day_of_week`, `week_number`), evaluates the shift amount expression, and supports self-references.

**Enhanced AnalysisInputDefinition (Increment 6)**

```go
type AnalysisInputDefinition struct {
    InputKind      string
    Name           string
    Alias          string
    WhereClause    *AnalysisWhereClause
    TimeShift      *AnalysisTimeShift      // NEW
    ResolvedMetric *AnalysisMetricDefinition
    SourceLocation SourceLocation
}

type AnalysisTimeShift struct {
    Direction string  // "backward" or "forward"
    Field     string  // Instant constraint field (e.g., "week_number", "day")
    Amount    float64 // Shift amount; defaults to 1.0 if not explicitly provided
}
```

**Self-Reference Validation**

```go
func (a *CPLAnalyzer) validateSelfReferences(metricName string, inputs []AnalysisInputDefinition) error {
    for _, input := range inputs {
        if input.Name == metricName && input.Kind == "metric" && input.TimeShift == nil {
            return fmt.Errorf("metric '%s' references itself without backward/forward time shift at line %d, column %d",
                metricName, input.SourceLocation.Line, input.SourceLocation.Column)
        }
    }
    return nil
}
```

**Validation During Analysis**

- Time shift constraint field must be a valid Instant field keyword
- Shift amount is a simple numeric literal (defaults to 1.0); expression evaluation is deferred to a future increment
- Self-references with a backward or forward time shift are permitted (valid time-series dependency)
- Direct self-reference without a time shift is a compile-time error
- `backward` and `forward` are reserved keywords; an input cannot specify both clauses — this mutual exclusivity is enforced by the parser

### Increment 7: Short Form Metrics

**Overview**

Increment 7 introduces short-form metric name resolution. The parser produces `ParserMetricDefinition` entries with `ShortForm: true` and no explicit `Inputs` list. The analysis phase scans the CIL tree for identifier references and resolves each to `metric`, `lookup` (future), or `dimension` using this precedence:

1. Match against declared metrics → `metric` input
2. Match against declared lookups → `lookup` input (deferred)
3. Match against dimensions → `dimension` input
4. Name collision between categories → compile-time error

**Short Form Resolution Algorithm**

```go
func (a *CPLAnalyzer) resolveShortFormInputs(metricName string, cilTree CILNode) ([]AnalysisInputDefinition, error) {
    refs := collectInputRefs(cilTree)
    seen := make(map[string]bool)
    var inputs []AnalysisInputDefinition

    for _, ref := range refs {
        if seen[ref.Alias] {
            continue
        }
        seen[ref.Alias] = true

        isMetric := a.metrics[ref.Alias] != nil
        isDimension := a.concepts[ref.Alias] != nil

        if isMetric && isDimension {
            return nil, fmt.Errorf("short form name collision: '%s' matches both a metric and a dimension in metric '%s'",
                ref.Alias, metricName)
        }

        var inputKind string
        if isMetric {
            inputKind = "metric"
        } else if isDimension {
            inputKind = "dimension"
        } else {
            return nil, fmt.Errorf("unresolved name '%s' in short form metric '%s'", ref.Alias, metricName)
        }

        inputs = append(inputs, AnalysisInputDefinition{
            InputKind: inputKind,
            Name:      ref.Alias,
            Alias:     ref.Alias,
        })
    }

    return inputs, nil
}

// collectInputRefs walks a CIL tree and collects all CILInputRef nodes.
func collectInputRefs(node CILNode) []*CILInputRef {
    var refs []*CILInputRef
    switch n := node.(type) {
    case *CILInputRef:
        refs = append(refs, n)
    case *CILBinaryOp:
        refs = append(refs, collectInputRefs(n.Left)...)
        refs = append(refs, collectInputRefs(n.Right)...)
    case *CILUnaryOp:
        refs = append(refs, collectInputRefs(n.Operand)...)
    case *CILConditional:
        refs = append(refs, collectInputRefs(n.IfBranch.Condition)...)
        refs = append(refs, collectInputRefs(n.IfBranch.Body)...)
        for _, elif := range n.ElseIfs {
            refs = append(refs, collectInputRefs(elif.Condition)...)
            refs = append(refs, collectInputRefs(elif.Body)...)
        }
        if n.ElseBranch != nil {
            refs = append(refs, collectInputRefs(n.ElseBranch)...)
        }
    case *CILWhenClause:
        for _, clause := range n.Clauses {
            refs = append(refs, collectInputRefs(clause.Condition)...)
            refs = append(refs, collectInputRefs(clause.Body)...)
        }
        refs = append(refs, collectInputRefs(n.DefaultBranch)...)
    }
    return refs
}
```

**Validation During Analysis**

- Name collision detection: a name matching multiple entity types is an error
- All names in short form must resolve to a known entity
- Short form metrics are converted to equivalent long form internally

### Increment 8: Full Synthesis with IO Specification Integration

**Overview**

Increment 8 involves no new data structures for the analysis phase. All syntax is already supported by Increments 1-7. Increment 8 adds the second-pass dependency DAG construction, topological ordering, and dead-code elimination.

**Metric Dependency DAG Construction (Second Pass)**

```go
type MetricDAG struct {
    Nodes      map[string]*DAGNode
    RootNodes  []string             // Metrics with no dependencies on other metrics
    Order      []string             // Topological evaluation order
}

type DAGNode struct {
    MetricName   string
    Dependencies []string  // Names of metrics this metric depends on
    Dependents   []string  // Names of metrics that depend on this metric
    Visited      bool
    InProgress   bool      // For cycle detection (DFS gray-set)
}
```

**DAG Construction Algorithm**

```go
func (a *CPLAnalyzer) BuildMetricDependencyDAG() error {
    dag := &MetricDAG{Nodes: make(map[string]*DAGNode)}

    // Create nodes for all metrics
    for name := range a.metrics {
        dag.Nodes[name] = &DAGNode{MetricName: name}
    }

    // Establish edges based on input references
    for name, metric := range a.metrics {
        node := dag.Nodes[name]
        for _, input := range metric.Inputs {
            if input.InputKind == "metric" {
                if _, exists := a.metrics[input.Name]; exists {
                    node.Dependencies = append(node.Dependencies, input.Name)
                    if targetNode, exists := dag.Nodes[input.Name]; exists {
                        targetNode.Dependents = append(targetNode.Dependents, name)
                    }
                }
            }
        }
    }

    // DFS cycle detection (self-references with time shift are allowed)
    for _, node := range dag.Nodes {
        if !node.Visited {
            if err := a.detectCycle(dag, node); err != nil {
                return err
            }
        }
    }

    // Topological sort
    order, err := a.topologicalSort(dag)
    if err != nil {
        return err
    }
    dag.Order = order

    // Find root nodes
    for name, node := range dag.Nodes {
        if len(node.Dependencies) == 0 {
            dag.RootNodes = append(dag.RootNodes, name)
        }
    }

    a.metricDAG = dag
    return nil
}
```

**Dead-Code Elimination**

```go
func (a *CPLAnalyzer) EliminateDeadCode(outputMetrics []string) {
    reachable := make(map[string]bool)

    var markReachable func(name string)
    markReachable = func(name string) {
        if reachable[name] {
            return
        }
        reachable[name] = true
        metric, exists := a.metrics[name]
        if !exists {
            return
        }
        for _, input := range metric.Inputs {
            if input.InputKind == "metric" {
                markReachable(input.Name)
            }
        }
    }

    for _, name := range outputMetrics {
        markReachable(name)
    }

    for name := range a.metrics {
        if !reachable[name] {
            delete(a.metrics, name)
            delete(a.cilTrees, name)
        }
    }
}
```

**IO Specification Integration**

```go
func (a *CPLAnalyzer) ApplyIOSpec(ioSpec *IOSpecification) error {
    for _, output := range ioSpec.Outputs {
        if _, exists := a.metrics[output.MetricName]; !exists {
            return fmt.Errorf("IO spec references unknown metric: %s", output.MetricName)
        }
    }

    outputNames := make([]string, len(ioSpec.Outputs))
    for i, output := range ioSpec.Outputs {
        outputNames[i] = output.MetricName
    }
    a.EliminateDeadCode(outputNames)

    return nil
}
```

**Full AnalysisMetricDefinition (Through Increment 8)**

```go
type AnalysisMetricDefinition struct {
    Name           string
    ShortForm      bool                      // True if derived from short form (Increment 7)
    Inputs         []AnalysisInputDefinition  // Resolved inputs (Increments 1, 5, 6)
    CILTree        CILNode                   // CIL expression tree (Increments 1-4)
    SourceLocation SourceLocation
}

type AnalysisInputDefinition struct {
    InputKind      string                    // "metric" or "dimension"
    Name           string
    Alias          string
    WhereClause    *AnalysisWhereClause      // (Increment 5)
    TimeShift      *AnalysisTimeShift        // (Increment 6)
    SourceLocation SourceLocation            // For error reporting
}

// Note: Dependency resolution is performed by name lookup in the MetricDAG,
// not via a ResolvedMetric pointer field. The DAG's Nodes map provides O(1)
// access to all metric nodes by name.
```

## Approach

### Incremental Structure and Function Evolution

The analysis phase uses an incremental approach where structures and functions evolve by adding fields and capabilities as each increment requires them:

**Structure Evolution:**

| Increment | Structures Added / Modified                                                                                                                                                |
| --------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1         | `AnalysisMetricDefinition` (Name, Inputs, CILTree, SourceLocation), `AnalysisInputDefinition` (InputKind, Name, SourceLocation), `CILLiteral`, `CILConnectorReferenceNode` |
| 2         | `CILBinaryOp`, `CILUnaryOp`, `CILInputRef`, `IOSpecification` acceptance                                                                                                   |
| 3         | None (synthesis-only — tensor parallelism)                                                                                                                                 |
| 4         | `CILConditional`, `CILWhenClause`, `CILComparisonOp`, `CILLogicalOp`, `CILNotOp`, `CILNilLiteral`, `CILBoolLiteral`                                                        |
| 5         | `AnalysisInputDefinition.Alias`, `AnalysisInputDefinition.WhereClause`, `AnalysisWhereClause`, `AnalysisWhereCondition`, `AnalysisAspectEquals`                            |
| 6         | `AnalysisInputDefinition.TimeShift`, `AnalysisTimeShift`                                                                                                                   |
| 7         | `AnalysisMetricDefinition.ShortForm`                                                                                                                                       |
| 8         | `MetricDAG`, `DAGNode` (second-pass structures)                                                                                                                            |

**Function Evolution:**

| Increment | Functions Added / Extended                                                                                                                                                                                                           |
| --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1         | `AnalyzeMetricScope`, `buildCILTreeFromScope` (literal only), `resolveMetricInputs` (basic)                                                                                                                                          |
| 2         | Extended `buildCILTreeFromScope` with arithmetic operators, `CILInputRef` handling; `ApplyIOSpec` (basic validation)                                                                                                                 |
| 3         | None (synthesis-only — tensor parallelism)                                                                                                                                                                                           |
| 4         | Extended `buildCILTreeFromScope` with comparison/logical operators and Nil/bool literals; `buildConditionalFromScope`, `assembleIfConditional`, `assembleWhenClause` for scope-based conditional assembly; truthiness/Nil validation |
| 5         | Extended `resolveMetricInputs` with alias resolution, `resolveWhereClause`                                                                                                                                                           |
| 6         | Extended `resolveMetricInputs` with time-shift resolution, `validateSelfReferences`                                                                                                                                                  |
| 7         | `resolveShortFormInputs`, `collectInputRefs`                                                                                                                                                                                         |
| 8         | `BuildMetricDependencyDAG`, `detectCycle`, `topologicalSort`, `EliminateDeadCode`                                                                                                                                                    |

**Multi-Pass Analysis:**

1. **First pass (WalkSymbolTableWithHandlers):** Extracts `ParserMetricDefinition` from each `MetricScope`, resolves inputs against concepts and declared metrics, evaluates postfix to build CIL trees. Input metric references are validated for existence but `ResolvedMetric` pointers are left nil.

2. **Second pass (BuildMetricDependencyDAG):** After all definitions from all sources are assembled, constructs the dependency DAG, resolves `ResolvedMetric` pointers, detects cycles, computes topological order, performs dead-code elimination, and integrates IO specification.

## Decision

**Chosen Approach:** Incremental evolution of structures and functions with two-pass analysis (symbol table walk → DAG construction).

**Rationale:**
This approach mirrors the natural development process and the established pattern from concepts (ERO001), types (ERO002), units/measures (ERO003), and connectors (ERO004) analysis. Each increment builds upon working code from previous increments by adding only what's needed. The two-pass design is necessitated by the metric dependency DAG: the full set of metric definitions must be assembled before dependencies can be resolved and cycles detected.

## Implications

### Positive Implications

- Each increment builds upon working, tested code from previous increments
- No dead code or placeholder fields in the final implementation
- Two-pass design cleanly separates structural analysis (walk) from dependency analysis (DAG)
- Expression tree construction via postfix evaluation is language-agnostic and reusable
- Consistent with all existing analysis-phase EROs (ERO001-ERO004)

### Negative Implications / Trade-offs

- Two-pass design means metric input references are not fully resolved until the second pass
- Short-form name resolution requires a CIL tree walk to collect `CILInputRef` nodes
- Conditional assembly requires walking child scopes and pairing clause definitions with body scopes; this is a different pattern from simple postfix evaluation
- Need to maintain this ERO to track what exists at each increment

### Validation Strategy

1. **Increment 1:** Metric name uniqueness, input metric/dimension existence, literal `0` postfix evaluation
2. **Increment 2:** Arithmetic operator operand count, input reference validity in expressions, type-checking, IO spec output metric validation
3. **Increment 3:** None (synthesis-only — tensor parallelism)
4. **Increment 4:** Clause/body scope correspondence, truthiness rules, Nil-in-arithmetic prohibition, expression completeness
5. **Increment 5:** Alias uniqueness, where-clause aspect values, boolean combination well-formedness
6. **Increment 6:** Time-shift field validity, self-reference validation, shift expression type-checking
7. **Increment 7:** Short-form name resolution, collision detection, all-names-resolved check
8. **Increment 8:** DAG acyclicity, output metric existence, topological order computability, dead-code elimination correctness

## Notes

**Related Documents:**

- [Metrics Design](../../metrics-design.md) — Full metrics design with core rules, expression handling, dependency DAG, and 8 end-to-end increments
- [ERO005 — Metrics Parsing Phase](../../go-parser-target/eros/ERO005_metrics-parsing-phase.md) — Parsing phase data structures (produces the `ParserMetricDefinition` input to this analysis phase)
- [ERO005 — Metrics Incremental Implementation](../../eros/ERO005_metrics-incremental-implementation.md) — Top-level ERO defining the 8-increment approach
- [Metrics Syntax Requirements](../../../../requirements/cpl/metrics-syntax.md) — Syntax requirements for metric declarations, inputs, and calculations
- [IO Specifications Requirements](../../../../requirements/cpl/io_specs.md) — IO specification requirements for output metric selection
- [Postfix Expression Parsing Pattern](../../../../patterns/postfix-expression-parsing.md) — Pattern for expression handling via postfix notation
- [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md) — Architectural pattern for parse/analysis phase separation
- [CPL Architecture — Intermediate Language](../../../architecture/cpl/analysis/intermediate_language.md) — CIL specification
- [ERO001 — Concepts Analysis Phase](ERO001_concepts-analysis-phase.md) — Reference ERO for analysis phase structure and pattern
- [ERO004 — Connectors Analysis Phase](ERO004_connectors-analysis-phase.md) — Connector analysis (connectors provide leaf data for metrics)
