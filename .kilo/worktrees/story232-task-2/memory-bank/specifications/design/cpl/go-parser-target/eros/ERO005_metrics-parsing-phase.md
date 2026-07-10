# ERO005: Metrics Parsing Phase

## Title

Metrics Parsing Phase — Data Structures and Incremental Implementation

## Purpose/Goal

Define the parsing phase data structures for metric definitions (including `ParserMetricDefinition`, expression handling via postfix notation, and `CPLMetrics.g4` grammar) and document how these structures evolve across each increment of the metrics incremental implementation as defined in [ERO005 — Metrics Incremental Implementation](../../eros/ERO005_metrics-incremental-implementation.md).

## Decision Status

Draft

## Implementation Status

Not Yet Implemented

**Related Epics:**

- Epic005_cpl-metrics-implementation.md

## Parsing vs Analysis Phase Separation

Following the [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md), metric processing is divided into two distinct phases:

### Parsing Phase (This ERO)

The parsing phase walks the ANTLR parse tree and builds the symbol table tree. Its responsibilities are:

1. **Push/pop scope tables** as metric declaration contexts are entered and exited
2. **Define symbols** (metric names and definition types) in the current symbol table
3. **Attach parse-context data** to `SymbolDefinition.Data` for use by the analysis phase
4. **Push expression operands/operators in postfix order** following the [Postfix Expression Parsing Pattern](../../../../patterns/postfix-expression-parsing.md)

The parser does **not** build an expression AST tree. Instead, it populates the symbol table with raw parsed data and postfix expression sequences.

### Grammar Organization

Expression rules (arithmetic, comparison, logical, conditional) are defined in [`ChronoconePlanningLanguage.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguage.g4:1) — the **root grammar** — because expressions are needed by multiple CPL constructs:

- Metric `calculate` clauses
- Unit and measure conversion `calculation` clauses
- Aggregator and spreader expressions (future)

[`CPLMetrics.g4`](src/chronocone_planning_language/src/main/antlr/CPLMetrics.g4) contains only metrics-specific rules (`metricsSection`, `metricDeclaration`, `inputsBlock`, `inputDeclaration`, `calculateClause`, `shortFormDeclaration`, `whereExpression`, `timeShiftClause`) and imports the expression rules from the root grammar.

Correspondingly, expression-related listener methods (operators, literals) live in the root grammar listener, not in the metrics listener. These methods check the current scope type to determine whether to push postfix definitions. Conditional structure (pushing clause definitions and child body scopes) is handled by the metrics listener.

### Analysis Phase (Separate ERO)

The analysis phase walks the symbol table tree and assembles the final AST structures. Its responsibilities are:

1. **Walk the symbol table tree** using `WalkSymbolTable`
2. **Assemble `AnalysisMetricDefinition`** entries from parser definitions
3. **Perform postfix evaluation** to build CIL expression trees
4. **Build the metric dependency DAG** in a second pass

### Symbol Table Storage Pattern

Metric definitions are stored in the symbol table using the following pattern:

```go
// ParserMetricDefinition is stored in SymbolDefinition.Data during parsing.
// The analysis phase retrieves this data to build AnalysisMetricDefinition.
type ParserMetricDefinition struct {
    Name           string
    Inputs         []ParserInputDefinition
    SourceLocation SourceLocation
    // CalculateDefs is populated by the metric scope's Definitions slice
    // — the postfix sequence of operands and operators pushed by the listener.
}

// SymbolDefinition holds the metric data in its Data field
type SymbolDefinition struct {
    Name           *string
    DefinitionType DefinitionType  // MetricScopeDefinition
    Location       SourceLocation
    Data           interface{}     // Contains *ParserMetricDefinition
}
```

### Expression Handling: Postfix via Symbol Definitions

The parser does **not** build an expression tree. Instead, it follows the [Postfix Expression Parsing Pattern](../../../../patterns/postfix-expression-parsing.md): as ANTLR `Exit*` listener methods visit expression components in operator precedence order, the parser pushes `SymbolDefinition` entries onto the current scope's `Definitions` slice:

- **Operands**: Identifiers (input references, dimension names), numeric/string/boolean/Nil literals
- **Operators**: Arithmetic (`+`, `-`, `*`, `/`), comparison (`==`, `!=`, `<`, `>`, `<=`, `>=`), logical (`and`, `or`, `not`)

For example, `a + b * c` results in Definitions: `[a, b, c, *, +]` — postfix for `a + (b * c)`.

## Context

### Situation

The parsing phase must recognize `metrics` sections and metric declarations with `inputs` and `calculate` sub-declarations. Metric parsing uses expression handling — arithmetic, comparison, logical, and conditional expressions — within the `calculate` clause. Following the established pattern, expression handling uses postfix notation: the parser pushes operands and operators in precedence order rather than building an AST tree.

Expression rules are defined in the root grammar because they are shared across CPL constructs (metrics, unit/measure conversions, future aggregators and spreaders). The root grammar listener handles expression operator and literal push methods, checking the current scope type to determine context.

### Constraints

- Must integrate with the existing ANTLR-based parser infrastructure
- Must populate the symbol table for use by subsequent phases
- Must follow the [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md)
- Must follow the [Postfix Expression Parsing Pattern](../../../../patterns/postfix-expression-parsing.md)
- Parse/analysis separation must be maintained: parser captures syntax only
- Must use indentation-aware lexing (INDENT/DEDENT tokens) consistent with existing CPL parsing
- Listener methods must be organized one file per grammar file (ANTLR Listener Organization Convention)
- Expression rules live in the root grammar; metrics grammar imports them

### Requirements

- Parse `metrics` sections and metric declarations with `inputs` and `calculate` sub-declarations (see [metrics-syntax.md](../../../../requirements/cpl/metrics-syntax.md))
- Build raw `ParserMetricDefinition` structures stored in `SymbolDefinition.Data` during parsing
- Push operand and operator definitions in postfix order into the metric scope's `Definitions` slice
- Time shift clauses (`backward`/`forward`) must accept the same constraint fields as [`Instant`](../../types-definitions.md) declarations
- Support error reporting with source location information

### Assumptions

- Concepts, types, units, measures, and connectors are already implemented (Epics 001-004 provide the foundation)
- The ANTLR root grammar can be extended with shared expression rules
- A `CPLMetrics.g4` grammar file imports expression rules from the root grammar
- Expression parsing uses the postfix approach as described in the Postfix Expression Parsing Pattern
- The indentation-aware lexer from Increment 2 of concepts is available
- Instant constraint field keywords (`year`, `month`, `day`, `hour`, `minute`, `second`, `day_of_week`, `week_number`) already exist as lexer tokens from the types implementation

## Evaluation Criteria / Strategy Fit

1. **Parse/Analysis Separation** (High importance): The parser must not perform semantic validation; all cross-references and dependency resolution must happen in analysis.
2. **Grammar Organization** (High importance): One listener file per grammar file. Shared expression rules in root grammar; metrics-specific rules in CPLMetrics.g4.
3. **Incremental Evolvability** (High importance): Data structures must support adding fields in later increments without breaking earlier increments.
4. **Consistency** (Medium importance): Data structures and patterns should evolve consistently with the existing concepts, types, units, and connectors implementations.
5. **Source Location Accuracy** (Medium importance): All parsed elements must carry source location information for error reporting in subsequent phases.

## Options Considered

Two approaches were evaluated for parser data structures. **Separate structures per increment** (e.g., `ParserMetricDefinitionV1`, `V2`) was rejected due to duplication and inconsistency with the established pattern. **Single evolving `ParserMetricDefinition` structure** that grows with each increment was selected because it matches the pattern used successfully for concepts, types, units/measures, and connectors — Go zero values provide sensible defaults for unused fields, and documentation and validation rules enforce increment constraints.

## Decision

**Chosen Option:** Option 1 — Single evolving `ParserMetricDefinition` structure with incrementally added fields.

**Rationale:**
The simplicity of a single structure outweighs the benefits of strict versioning. This is the same approach used successfully for concepts (ERO001), types (ERO002), units/measures (ERO003), and connectors (ERO004). Documentation and validation rules clearly define which fields are valid in each increment. Go zero values for unused fields provide a sensible default behavior.

## Implications

### Positive Implications

- Simpler codebase with single structure definition
- Easy to add new fields as increments progress
- Backward compatible (new fields have zero values)
- Consistent with all other CPL parse-phase EROs
- Shared expression grammar enables reuse by unit/measure conversions, aggregators, and spreaders

### Negative Implications / Trade-offs

- Must rely on documentation and validation to enforce increment constraints
- Unit tests should verify correct behavior per increment
- Expression postfix sequences grow more complex with each increment
- Root grammar listener must be scope-type-aware to push definitions in the right context

### Future Considerations

- Lookup support is deferred to a future design/ERO
- Aggregator and spreader integration will reuse the same shared expression grammar
- The `MetricScope` used for postfix definition accumulation, along with the conditional body scopes (`IfBodyScope`, `ElseBodyScope`, etc.), may need analogous scope types for additional expression-using contexts (ConversionScope, AggregatorScope, SpreaderScope)

## Data Structures by Increment

The following sections define the parsing phase data structures for each increment of the metrics incremental implementation. Each increment corresponds to an increment defined in [ERO005 — Metrics Incremental Implementation](../../eros/ERO005_metrics-incremental-implementation.md) and the [Metrics Design](../../metrics-design.md).

### Increment 1: Basic Metric Declarations with Identity Calculation

**Overview**

Increment 1 introduces the minimal metric parsing infrastructure: a `metrics` section containing metric declarations with `inputs` blocks (simple `metric` and `dimension` name references only — no `as` alias, no `where` clause, no `backward`/`forward`) and a `calculate` clause that accepts only the literal `0`. The calculate body results in a single postfix definition: a literal `0` operand.

**Grammar Changes**

In [`ChronoconePlanningLanguage.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguage.g4:1), add `metricsSection` to the statement rule and import the metrics grammar:

```antlr
// Root grammar imports the metrics grammar along with other section grammars
grammar ChronoconePlanningLanguage;

import CPLConcepts, CPLMeasures, CPLTypes, CPLUnits, CPLConnectors, CPLMetrics;

options {
    tokenVocab = ChronoconePlanningLanguageLexer;
}

// Statements are the top-level elements of a CPL file
statement
    : conceptsSection
    | typesSection
    | unitsSection
    | measuresSection
    | connectorsSection
    | metricsSection;
```

Create new file `CPLMetrics.g4` — metrics-specific rules only. Expression rules are defined in the root grammar; `CPLMetrics.g4` imports them:

```antlr
// CPLMetrics.g4 - Grammar for CPL Metrics section (Increment 1)

parser grammar CPLMetrics;

import ChronoconePlanningLanguage; // For identifier, literal, expression rules

options {
    tokenVocab = ChronoconePlanningLanguageLexer;
}

// A metrics section contains one or more metric declarations
metricsSection
    : METRICS_KEYWORD
      INDENT
      metricDeclaration+
      DEDENT
    ;

// A metric declaration: name, inputs block, calculate clause
metricDeclaration
    : identifier
      INDENT
      inputsBlock
      calculateClause
      DEDENT
    ;

// Inputs block: 'inputs' keyword followed by indented input declarations
inputsBlock
    : INPUTS_KEYWORD
      INDENT
      inputDeclaration+
      DEDENT
    ;

// An input declaration: metric|dimension <name>
// (No 'as', 'where', or 'backward'/'forward' in Increment 1)
inputDeclaration
    : inputKind identifier
    ;

// Input kind discriminator
inputKind
    : METRIC_KEYWORD
    | DIMENSION_KEYWORD
    // LOOKUP_KEYWORD is recognized by the lexer but lookups are deferred
    ;

// Calculate clause: 'calculate' keyword followed by indented expression
// In Increment 1, only the literal '0' is accepted
calculateClause
    : CALCULATE_KEYWORD
      INDENT
      literal
      DEDENT
    ;
```

Note: `identifier`, `literal`, and `INDENT`/`DEDENT` tokens are defined in the root grammar and lexer. `CPLMetrics.g4` imports them.

**Lexer Tokens** (add to [`ChronoconePlanningLanguageLexer.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguageLexer.g4:1)):

```antlr
// Metrics keywords
METRICS_KEYWORD   : 'metrics';
INPUTS_KEYWORD    : 'inputs';
CALCULATE_KEYWORD : 'calculate';
METRIC_KEYWORD    : 'metric';
LOOKUP_KEYWORD    : 'lookup';
DIMENSION_KEYWORD : 'dimension';
```

**Scope and Definition Types**

```go
// ScopeType values for metric-related scopes
const (
    // ... existing scopes ...
    MetricsScope ScopeType = iota // Scope for metrics section
    MetricScope                   // Scope for individual metric declaration
)

// DefinitionType values for metric definitions
const (
    // ... existing definition types ...
    MetricsScopeDefinition DefinitionType = iota // Definition type for metrics section scope
    MetricScopeDefinition                        // Definition type for individual metric scope
    InputRefDefinition                           // Definition type for input references
    // LiteralDefinition — already defined in root grammar for shared expression types
)
```

**ParserMetricDefinition (Increment 1)**

```go
// ParserMetricDefinition is stored in SymbolDefinition.Data during parsing.
// The analysis phase retrieves this data to build AnalysisMetricDefinition.
type ParserMetricDefinition struct {
    Name           string                  // Metric identifier
    Inputs         []ParserInputDefinition  // Input declarations
    SourceLocation SourceLocation           // For error reporting
    // Note: CalculateDefs is NOT a field on ParserMetricDefinition.
    // The postfix sequence is accumulated in the MetricScope's Definitions slice.
}

// ParserInputDefinition captures a single input declaration during parsing.
type ParserInputDefinition struct {
    InputKind      string         // "metric" or "dimension"
    Name           string         // Referenced metric/dimension name
    NameLocation   SourceLocation // Location of the referenced name
    // Future fields: Alias, WhereClause, TimeShift — all added in later increments
}
```

**Note on CalculateDefs:** The postfix expression sequence is NOT stored as a field on `ParserMetricDefinition`. Instead, it is accumulated directly in the `MetricScope` symbol table's `Definitions` slice. When the listener enters a metric declaration, it pushes a `MetricScope` table. As expression components are visited by the root grammar listener (`ExitLiteral`, later `ExitIdentifier`, operator `Exit*` methods), operand and operator `SymbolDefinition` entries are appended to that scope's `Definitions` slice in postfix order. The root grammar listener checks `scope.ScopeType == MetricScope` to determine whether to push definitions. The analysis phase reads these definitions from the symbol table to reconstruct the expression.

**Listener Implementation** (file: `cpl_metrics_listener.go` — metrics-specific listener methods only)

```go
// CPLMetricsListener handles metric section parsing.
// Following the ANTLR Listener Organization Convention, this file contains
// all Enter* and Exit* methods for rules defined in CPLMetrics.g4.
// Expression-related listener methods live in the root grammar listener.
type CPLMetricsListener struct {
    *parser.BaseChronoconePlanningLanguageListener
    symbolStack *symboltable.SymbolTableStack
}

// EnterMetricsSection pushes a new MetricsScope onto the symbol stack.
func (l *CPLMetricsListener) EnterMetricsSection(ctx *parser.MetricsSectionContext) {
    current, _ := l.symbolStack.Peek()
    metricsTable := current.AddChild(symboltable.MetricsScope)
    l.symbolStack.Push(metricsTable)
}

// ExitMetricsSection pops the MetricsScope from the symbol stack.
func (l *CPLMetricsListener) ExitMetricsSection(ctx *parser.MetricsSectionContext) {
    l.symbolStack.Pop()
}

// EnterMetricDeclaration pushes a new MetricScope onto the symbol stack.
func (l *CPLMetricsListener) EnterMetricDeclaration(ctx *parser.MetricDeclarationContext) {
    current, _ := l.symbolStack.Peek()
    metricTable := current.AddChild(symboltable.MetricScope)
    l.symbolStack.Push(metricTable)
}

// ExitMetricDeclaration assembles the ParserMetricDefinition.
func (l *CPLMetricsListener) ExitMetricDeclaration(ctx *parser.MetricDeclarationContext) {
    metricScope, _ := l.symbolStack.Peek()

    metricName := ctx.Identifier().GetText()
    metricLocation := makeSourceLocation(ctx.Identifier().GetStart())

    // Collect input definitions from child scope
    inputs := l.collectInputDefinitions(metricScope)

    metricDef := &ParserMetricDefinition{
        Name:           metricName,
        Inputs:         inputs,
        SourceLocation: metricLocation,
    }

    // Store in SymbolDefinition.Data
    metricScope.Define(&metricName, symboltable.MetricScopeDefinition, metricLocation)
    lastIdx := len(metricScope.Definitions) - 1
    metricScope.Definitions[lastIdx].Data = metricDef

    l.symbolStack.Pop()
}

// ExitInputDeclaration captures an input reference definition.
func (l *CPLMetricsListener) ExitInputDeclaration(ctx *parser.InputDeclarationContext) {
    metricScope, _ := l.symbolStack.Peek()

    inputKind := ctx.InputKind().GetText()
    inputName := ctx.Identifier().GetText()
    inputLoc := makeSourceLocation(ctx.Identifier().GetStart())

    inputDef := &ParserInputDefinition{
        InputKind:    inputKind,
        Name:         inputName,
        NameLocation: inputLoc,
    }

    // Define in the metric scope so it's part of the symbol table tree
    metricScope.Define(&inputName, symboltable.InputRefDefinition, inputLoc)
    lastIdx := len(metricScope.Definitions) - 1
    metricScope.Definitions[lastIdx].Data = inputDef
}
```

Note: `ExitLiteral` and `ExitIdentifier` (for expression operands) are handled by the root grammar listener. The root listener checks `scope.ScopeType == MetricScope` before pushing definitions. This pattern is used by all expression-related listener methods.

**Symbol Table Structure Example (After Parsing)**

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

The resulting symbol table tree is:

```
GlobalScope (root)
└── MetricsScope
    └── MetricScope (Name: "revenue")
        ├── SymbolDefinition
        │   Name: "demand_forecast"
        │   DefinitionType: InputRefDefinition
        │   Data: &ParserInputDefinition{InputKind: "metric", Name: "demand_forecast"}
        ├── SymbolDefinition
        │   Name: "Price"
        │   DefinitionType: InputRefDefinition
        │   Data: &ParserInputDefinition{InputKind: "dimension", Name: "Price"}
        ├── SymbolDefinition
        │   Name: "0"
        │   DefinitionType: LiteralDefinition
        │   Data: &LiteralData{Value: "0"}
        └── SymbolDefinition
            Name: "revenue"
            DefinitionType: MetricScopeDefinition
            Data: &ParserMetricDefinition{Name: "revenue", Inputs: [...]}
```

The analysis phase walks this tree. It identifies `MetricScopeDefinition` entries to assemble `AnalysisMetricDefinition`, reads `InputRefDefinition` entries to build the inputs list, and iterates the `LiteralDefinition` entries (in order) as the postfix expression sequence.

**Validation During Parsing**

- Metric name syntax validation (valid identifier)
- `inputs` block must be present and indented correctly
- `calculate` clause must be present and indented correctly
- Literal `0` must be a valid integer token
- Input kind must be `metric` or `dimension` (not an unrecognized keyword)
- Indentation consistency (tabs for CPL indentation)

**Limitations**

- No `as` alias clause on inputs (added in Increment 4)
- No `where` clause on inputs (added in Increment 4)
- No `backward`/`forward` time shifting (added in Increment 5)
- No `lookup` input kind support (deferred to future design)
- Calculate clause accepts only the literal `0` — no arithmetic, conditionals, or input references
- No short form metric syntax (added in Increment 6)
- No multiple metrics sections in the same file (added alongside Increment 3 by expanding the grammar)

### Increment 2: Simple Arithmetic in Calculate

**Overview**

Increment 2 extends the `calculate` clause to support arithmetic expressions with `+`, `-`, `*`, `/`, parenthesized expressions, numeric literals, string literals, and input references as operands. Expression rules are added to the **root grammar** since they will be reused by unit/measure conversions, aggregators, and spreaders. The parser pushes operands (identifiers, literals) and operators in postfix order following operator precedence.

**Grammar Changes**

Add expression rules to the root grammar `ChronoconePlanningLanguage.g4`:

```antlr
// ChronoconePlanningLanguage.g4 - Root grammar (Increment 2 additions)

// ============================================================================
// SHARED EXPRESSION RULES (in operator precedence order, lowest to highest)
// Used by: metrics calculate, unit/measure conversions, aggregators, spreaders
// ============================================================================

// Top-level expression
expression
    : additionExpr
    ;

// Addition/subtraction (lowest precedence)
additionExpr
    : multiplicationExpr ((PLUS | MINUS) multiplicationExpr)*
    ;

// Multiplication/division
multiplicationExpr
    : unaryExpr ((MULTIPLY | DIVIDE) unaryExpr)*
    ;

// Unary operators (negation)
unaryExpr
    : (MINUS)? primaryExpr
    ;

// Primary expressions: literals, identifiers, parenthesized sub-expressions
primaryExpr
    : literal
    | identifier
    | LPAREN expression RPAREN
    ;

// Literal values (extended in later increments)
literal
    : INTEGER
    | FLOAT
    | STRING
    ;
```

Update `CPLMetrics.g4` to use the shared `expression` rule:

```antlr
// CPLMetrics.g4 (Increment 2) — only the calculateClause changes

// Calculate clause: 'calculate' keyword followed by indented expression
// Uses the shared 'expression' rule from the root grammar
calculateClause
    : CALCULATE_KEYWORD
      INDENT
      expression
      DEDENT
    ;
```

**Lexer Tokens** (add to `ChronoconePlanningLanguageLexer.g4`):

```antlr
// Arithmetic operators (shared across all expression contexts)
PLUS      : '+';
MINUS     : '-';
MULTIPLY  : '*';
DIVIDE    : '/';

// Parentheses
LPAREN    : '(';
RPAREN    : ')';

// Literals
FLOAT     : [0-9]+ '.' [0-9]+;
STRING    : '"' (~["\r\n])* '"';
```

**New Definition Types** (in root grammar's scope, shared across expression contexts)

```go
const (
    // ... existing definition types ...
    PlusOperatorDefinition     DefinitionType = iota // '+' operator
    MinusOperatorDefinition                           // '-' operator (binary)
    MultiplyOperatorDefinition                        // '*' operator
    DivideOperatorDefinition                          // '/' operator
    NegateOperatorDefinition                          // Unary '-' (negation)
    NumericLiteralDefinition                          // Numeric literal (int or float)
    StringLiteralDefinition                           // String literal
    // InputRefDefinition already exists from Increment 1
)
```

**Listener Implementation** (in root grammar listener, not metrics listener)

The key pattern: each `Exit*` method for a binary operator rule pushes an operator definition onto the current scope's `Definitions` slice in postfix order — but only when the current scope is a `MetricScope` (or other expression-using scope in the future). Because ANTLR calls `Exit*` methods for sub-expressions before the parent expression, operands are naturally pushed before operators.

```go
// In the root grammar listener (chronocone_planning_language_listener.go):

// ExitAdditionExpr pushes a PLUS or MINUS operator definition for each operator.
func (l *CPLListener) ExitAdditionExpr(ctx *parser.AdditionExprContext) {
    current, _ := l.symbolStack.Peek()
    if !l.shouldPushExpressionDefs(current) {
        return
    }

    for _, op := range ctx.AllPLUS() {
        l.pushOperator(current, symboltable.PlusOperatorDefinition, op.GetSymbol())
    }
    for _, op := range ctx.AllMINUS() {
        l.pushOperator(current, symboltable.MinusOperatorDefinition, op.GetSymbol())
    }
}

// ExitMultiplicationExpr pushes MULTIPLY or DIVIDE operator definitions.
func (l *CPLListener) ExitMultiplicationExpr(ctx *parser.MultiplicationExprContext) {
    current, _ := l.symbolStack.Peek()
    if !l.shouldPushExpressionDefs(current) {
        return
    }

    for _, op := range ctx.AllMULTIPLY() {
        l.pushOperator(current, symboltable.MultiplyOperatorDefinition, op.GetSymbol())
    }
    for _, op := range ctx.AllDIVIDE() {
        l.pushOperator(current, symboltable.DivideOperatorDefinition, op.GetSymbol())
    }
}

// ExitIdentifier pushes an input reference operand when inside an expression scope.
func (l *CPLListener) ExitIdentifier(ctx *parser.IdentifierContext) {
    current, _ := l.symbolStack.Peek()
    if !l.shouldPushExpressionDefs(current) {
        return // Identifier used outside expression context — handled elsewhere
    }

    name := ctx.GetText()
    loc := makeSourceLocation(ctx.GetStart())

    current.Define(&name, symboltable.InputRefDefinition, loc)
    lastIdx := len(current.Definitions) - 1
    current.Definitions[lastIdx].Data = &InputRefData{
        Name: name,
        Loc:  loc,
    }
}

// shouldPushExpressionDefs returns true when the current scope is one that
// accumulates expression postfix definitions.
func (l *CPLListener) shouldPushExpressionDefs(scope *symboltable.SymbolTable) bool {
    switch scope.ScopeType {
    case symboltable.MetricScope:
        return true
    // Future: ConversionScope, AggregatorScope, SpreaderScope
    default:
        return false
    }
}

// pushOperator is a helper that pushes an operator definition onto a scope.
func (l *CPLListener) pushOperator(scope *symboltable.SymbolTable, defType symboltable.DefinitionType, token antlr.Token) {
    opText := token.GetText()
    loc := makeSourceLocation(token)
    scope.Define(&opText, defType, loc)
}
```

`ExitLiteral` is also in the root grammar listener, following the same scope-check pattern as `ExitIdentifier`. The `LiteralData` and `InputRefData` types are shared across all expression contexts.

**How Postfix Order Is Achieved**

Consider the expression `demand_forecast * Price + 10`. ANTLR's parse tree for `additionExpr → multiplicationExpr PLUS multiplicationExpr` results in the following listener call order:

1. `ExitIdentifier("demand_forecast")` — pushes operand
2. `ExitIdentifier("Price")` — pushes operand
3. `ExitMultiplicationExpr` — detects `*`, pushes `MultiplyOperatorDefinition`
4. `ExitLiteral("10")` — pushes operand
5. `ExitAdditionExpr` — detects `+`, pushes `PlusOperatorDefinition`

Resulting Definitions order: `[demand_forecast, Price, MULT, 10, PLUS]` — the postfix representation of `(demand_forecast * Price) + 10`.

**Validation During Parsing**

- Arithmetic operators must have two operands (validated during analysis, not parsing)
- Literals must be valid integer, float, or string tokens
- Parenthesized expressions must be properly balanced (enforced by ANTLR grammar)
- Input references in calculate must be valid identifiers

**Limitations**

- No `Nil` literal (added in Increment 3)
- No boolean literals (added in Increment 3)
- No conditional expressions (added in Increment 3)
- No comparison or logical operators (added in Increment 3)
- No dimension indexing on input refs (e.g., `order_lead_time[product]` — deferred to future)

> **Note: No Increment 3 in the parsing phase.** Increment 3 (Tensor Calculation Parallelism) is synthesis-only. It introduces no parser changes. The parser proceeds directly from Increment 2 to Increment 4.

### Increment 4: Conditional Expressions

**Overview**

Increment 4 extends the shared expression rules in the root grammar to support conditional expressions (`if`/`else if`/`else`/`end if`), the `when`/`default`/`end when` alternative syntax, comparison operators (`==`, `!=`, `<`, `>`, `<=`, `>=`), logical operators (`and`, `or`, `not`), `Nil` literal, and boolean literals (`True`, `False`).

**Scope-Based Conditional Structure**

Conditionals do NOT use flat marker definitions in the postfix sequence. Instead, the symbol table tree structure encodes the conditional hierarchy:

- **Condition expressions** use the same infix-to-postfix mechanism as arithmetic and boolean formulas. Comparison operators (`==`, `!=`, `<`, `>`, `<=`, `>=`) and logical operators (`and`, `or`, `not`) are pushed as regular operator definitions in the parent `MetricScope`'s `Definitions` slice. The condition's postfix is captured in a temporary scope during parsing and stored as `Data` on a clause-defining `SymbolDefinition` (e.g., `IfClauseDefinition`, `WhenClauseDefinition`) in the parent scope.

- **Branch bodies** each get their own child scope. The parser pushes a new child `SymbolTable` for each branch body (`IfBodyScope`, `ElseIfBodyScope`, `ElseBodyScope`, `WhenBodyScope`, `DefaultBodyScope`). The body's expression postfix lives in that child scope's `Definitions` slice, accumulated by the root grammar listener as it visits expression rules.

**Grammar Changes**

Extend the shared expression rules in `ChronoconePlanningLanguage.g4`:

```antlr
// ChronoconePlanningLanguage.g4 - Root grammar (Increment 3 additions)

// Top-level expression: now includes conditional expressions
expression
    : conditionalExpr
    ;

// Conditional expression: if/else if/else/end if
// Note: each branch body is a child scope in the symbol table;
// the grammar does not produce flat marker tokens for branch boundaries.
conditionalExpr
    : IF_KEYWORD
      INDENT
      ifBranch
      elseIfBranch*
      elseBranch?
      DEDENT
      END_KEYWORD IF_KEYWORD
    | whenExpr
    | logicalOrExpr  // Fall-through to logical/arithmetic when no conditional
    ;

// If branch: the condition is evaluated in the parent scope;
// the body expression defines a new child scope.
ifBranch
    : expression       // Condition — postfix pushed to parent scope
      INDENT
      expression       // Body — postfix pushed to child IfBodyScope
      DEDENT
    ;

// Else if branch: condition in parent scope, body in child ElseIfBodyScope
elseIfBranch
    : ELSE_KEYWORD IF_KEYWORD
      expression       // Condition — postfix pushed to parent scope
      INDENT
      expression       // Body — postfix pushed to child ElseIfBodyScope
      DEDENT
    ;

// Else branch: no condition, body in child ElseBodyScope
elseBranch
    : ELSE_KEYWORD
      INDENT
      expression       // Body — postfix pushed to child ElseBodyScope
      DEDENT
    ;

// When expression: when / default / end when
whenExpr
    : WHEN_KEYWORD
      INDENT
      whenClause+
      defaultClause
      DEDENT
      END_KEYWORD WHEN_KEYWORD
    ;

// When clause: <condition> '=>' <expression>
// Condition in parent scope, body in child WhenBodyScope
whenClause
    : expression ARROW
      INDENT
      expression
      DEDENT
    ;

// Default clause: 'default' '=>' <expression>
// No condition, body in child DefaultBodyScope
defaultClause
    : DEFAULT_KEYWORD ARROW
      INDENT
      expression
      DEDENT
    ;

// Logical OR (lowest precedence among logical operators)
logicalOrExpr
    : logicalAndExpr (OR_KEYWORD logicalAndExpr)*
    ;

// Logical AND
logicalAndExpr
    : equalityExpr (AND_KEYWORD equalityExpr)*
    ;

// Equality comparisons
equalityExpr
    : relationalExpr ((EQEQ | NOTEQ) relationalExpr)*
    ;

// Relational comparisons
relationalExpr
    : additionExpr ((LT | GT | LTEQ | GTEQ) additionExpr)*
    ;

// Addition/subtraction (unchanged from Increment 2)
additionExpr
    : multiplicationExpr ((PLUS | MINUS) multiplicationExpr)*
    ;

// Multiplication/division (unchanged from Increment 2)
multiplicationExpr
    : unaryExpr ((MULTIPLY | DIVIDE) unaryExpr)*
    ;

// Unary operators (now includes 'not')
unaryExpr
    : (MINUS | NOT_KEYWORD)? primaryExpr
    ;

// Primary expressions: expanded with Nil, True, False
primaryExpr
    : literal
    | identifier
    | LPAREN expression RPAREN
    ;

// Literal values: expanded
literal
    : INTEGER
    | FLOAT
    | STRING
    | NIL_KEYWORD
    | TRUE_KEYWORD
    | FALSE_KEYWORD
    ;
```

**Key Grammar Change:** The `ifBranch`, `elseIfBranch`, `elseBranch`, `whenClause`, and `defaultClause` rules now explicitly separate the condition (which stays in the parent scope) from the body expression (which gets its own indented block → new child scope). This replaces the previous flat approach where both condition and body were in the same postfix sequence separated by markers.

No changes to `CPLMetrics.g4` — the `calculateClause` already uses the shared `expression` rule.

**Lexer Tokens** (add to `ChronoconePlanningLanguageLexer.g4`):

```antlr
// Conditional keywords
IF_KEYWORD       : 'if';
ELSE_KEYWORD     : 'else';
END_KEYWORD      : 'end';
WHEN_KEYWORD     : 'when';
DEFAULT_KEYWORD  : 'default';

// Logical operators
AND_KEYWORD      : 'and';
OR_KEYWORD       : 'or';
NOT_KEYWORD      : 'not';

// Comparison operators
EQEQ             : '==';
NOTEQ            : '!=';
LT               : '<';
GT               : '>';
LTEQ             : '<=';
GTEQ             : '>=';

// Arrow token (for when clauses)
ARROW            : '=>';

// Literal keywords
NIL_KEYWORD      : 'Nil';
TRUE_KEYWORD     : 'True';
FALSE_KEYWORD    : 'False';
```

**New Scope and Definition Types**

```go
const (
    // Conditional body scope types — each branch body gets its own scope
    IfBodyScope       ScopeType = iota  // Scope for if branch body
    ElseIfBodyScope                      // Scope for else-if branch body
    ElseBodyScope                        // Scope for else branch body
    WhenBodyScope                        // Scope for when clause body
    DefaultBodyScope                     // Scope for default clause body

    // Clause definition types — stored in parent scope's Definitions,
    // carrying the condition's postfix as Data
    IfClauseDefinition     DefinitionType = iota  // Carries condition postfix
    ElseIfClauseDefinition                        // Carries condition postfix
    ElseClauseDefinition                          // No condition (else has no condition)
    WhenClauseDefinition                          // Carries condition postfix
    DefaultClauseDefinition                       // No condition

    // Comparison operators (shared, pushed to parent scope's Definitions in postfix order)
    EqOperatorDefinition
    NotEqOperatorDefinition
    LessThanOperatorDefinition
    GreaterThanOperatorDefinition
    LessThanEqOperatorDefinition
    GreaterThanEqOperatorDefinition

    // Logical operators (shared, pushed to parent scope's Definitions in postfix order)
    AndOperatorDefinition
    OrOperatorDefinition
    NotOperatorDefinition

    // Literal types
    NilLiteralDefinition
    TrueLiteralDefinition
    FalseLiteralDefinition
)

// ClauseData carries the condition's postfix definitions for a conditional clause.
// Stored as Data on IfClauseDefinition, ElseIfClauseDefinition, or WhenClauseDefinition.
type ClauseData struct {
    ConditionPostfix []SymbolDefinition  // Postfix for the condition expression
}
```

**Listener Implementation** (in root grammar listener and metrics listener)

The root grammar listener handles comparison/logical operators and literals — same pattern as Increment 2 operators. The `shouldPushExpressionDefs` function is extended to also return `true` for the new body scope types:

```go
func (l *CPLListener) shouldPushExpressionDefs(scope *symboltable.SymbolTable) bool {
    switch scope.ScopeType {
    case symboltable.MetricScope, symboltable.IfBodyScope,
         symboltable.ElseIfBodyScope, symboltable.ElseBodyScope,
         symboltable.WhenBodyScope, symboltable.DefaultBodyScope:
        return true
    // Future: ConversionScope, AggregatorScope, SpreaderScope
    default:
        return false
    }
}
```

The metrics listener (`cpl_metrics_listener.go`) handles the conditional structure by pushing clause definitions and child body scopes:

```go
// EnterIfBranch pushes a clause definition for the if condition,
// then pushes a new IfBodyScope for the body.
func (l *CPLMetricsListener) EnterIfBranch(ctx *parser.IfBranchContext) {
    parentScope, _ := l.symbolStack.Peek()

    // Push a temporary scope to capture the condition's postfix
    tempScope := parentScope.AddChild(symboltable.TempScope)
    l.symbolStack.Push(tempScope)
    l.conditionCapture = tempScope
}

func (l *CPLMetricsListener) ExitIfBranch(ctx *parser.IfBranchContext) {
    // Pop the temporary condition scope
    l.symbolStack.Pop()

    // Create an IfClauseDefinition in the parent scope with the condition postfix
    clauseData := &ClauseData{
        ConditionPostfix: l.conditionCapture.Definitions,
    }
    parentScope, _ := l.symbolStack.Peek()
    clauseName := "if"
    parentScope.Define(&clauseName, symboltable.IfClauseDefinition, makeSourceLocation(ctx.GetStart()))
    lastIdx := len(parentScope.Definitions) - 1
    parentScope.Definitions[lastIdx].Data = clauseData

    // The body scope (IfBodyScope) was pushed by the indentation handling
    // when the INDENT token after the condition was encountered
}

// EnterIfBodyScope is called when the INDENT after the condition triggers
// a new scope push (via the existing indentation-aware scope mechanism).
// The root listener will push expression defs into this scope.
func (l *CPLMetricsListener) EnterIfBodyScope() {
    parentScope, _ := l.symbolStack.Peek()
    bodyScope := parentScope.AddChild(symboltable.IfBodyScope)
    l.symbolStack.Push(bodyScope)
}

func (l *CPLMetricsListener) ExitIfBodyScope() {
    l.symbolStack.Pop()
}

// EnterElseIfBranch pushes a clause definition for the else-if condition,
// then pushes a new ElseIfBodyScope for the body.
func (l *CPLMetricsListener) EnterElseIfBranch(ctx *parser.ElseIfBranchContext) {
    parentScope, _ := l.symbolStack.Peek()

    tempScope := parentScope.AddChild(symboltable.TempScope)
    l.symbolStack.Push(tempScope)
    l.conditionCapture = tempScope
}

func (l *CPLMetricsListener) ExitElseIfBranch(ctx *parser.ElseIfBranchContext) {
    l.symbolStack.Pop()

    clauseData := &ClauseData{
        ConditionPostfix: l.conditionCapture.Definitions,
    }
    parentScope, _ := l.symbolStack.Peek()
    clauseName := "elseif"
    parentScope.Define(&clauseName, symboltable.ElseIfClauseDefinition, makeSourceLocation(ctx.GetStart()))
    lastIdx := len(parentScope.Definitions) - 1
    parentScope.Definitions[lastIdx].Data = clauseData
}

// EnterElseBranch pushes an ElseClauseDefinition (no condition) and an ElseBodyScope.
func (l *CPLMetricsListener) EnterElseBranch(ctx *parser.ElseBranchContext) {
    parentScope, _ := l.symbolStack.Peek()
    clauseName := "else"
    parentScope.Define(&clauseName, symboltable.ElseClauseDefinition, makeSourceLocation(ctx.GetStart()))
}

// EnterWhenExpr is handled similarly — pushes a WhenClauseDefinition with the
// condition postfix captured in a temporary scope, then a WhenBodyScope for the body.
// EnterDefaultClause pushes a DefaultClauseDefinition and a DefaultBodyScope.
```

**Symbol Table Structure Example (After Parsing)**

Given this CPL source:

```cpl
metrics
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

The resulting symbol table tree is:

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
    Data: &ParserMetricDefinition{Name: "revenue", Inputs: [...]}
```

The analysis phase walks this tree: it encounters the `IfClauseDefinition`, evaluates the condition postfix, walks the `IfBodyScope` child and evaluates its postfix for the true branch body, encounters the `ElseClauseDefinition`, walks the `ElseBodyScope` child for the else branch body, and assembles a `CILConditional` node.

**Validation During Parsing**

- `if` must be paired with `end if` (enforced by ANTLR grammar)
- `when` must be paired with `end when` (enforced by ANTLR grammar)
- `when` must have at least one clause and a `default` (enforced by grammar)
- Nil, True, False are recognized as literal keywords
- Comparison and logical operators have the expected tokens
- Condition expressions and body expressions must be syntactically valid

**Limitations**

- Truthiness evaluation is deferred to analysis phase
- Nil-in-arithmetic validation is deferred to analysis phase
- Expression completeness (all branches produce a result) is deferred to analysis phase
- No `lookup` input kind support (deferred to future design)

### Increment 5: Input Aliases and Where Clauses

**Overview**

Increment 4 extends input declarations to support the `as` alias clause and the `where` clause with `aspect=ACTUALS`/`aspect=FORECAST` filtering and boolean combinations (`and`, `or`, `not`). The `ParserInputDefinition` structure grows to accommodate these new fields. Where-clause rules are metrics-specific and live in `CPLMetrics.g4`, reusing the boolean operator tokens from the shared expression lexer.

**Grammar Changes**

Update the `inputDeclaration` rule in `CPLMetrics.g4`:

```antlr
// An input declaration: metric|dimension <name> [as <alias>] [where <boolean expression>]
inputDeclaration
    : inputKind identifier (AS_KEYWORD identifier)? (WHERE_KEYWORD whereExpression)?
    ;

// Where expression: boolean expression for filtering
// Supports aspect=ACTUALS, aspect=FORECAST, and boolean combinations
// Uses boolean operator tokens (AND_KEYWORD, OR_KEYWORD, NOT_KEYWORD) from shared lexer
whereExpression
    : whereAndExpr (OR_KEYWORD whereAndExpr)*
    ;

whereAndExpr
    : wherePrimary (AND_KEYWORD wherePrimary)*
    ;

wherePrimary
    : aspectEquals
    | NOT_KEYWORD wherePrimary
    | LPAREN whereExpression RPAREN
    ;

// aspect=ACTUALS or aspect=FORECAST
aspectEquals
    : ASPECT_KEYWORD EQ aspectValue
    ;

// Aspect values
aspectValue
    : ACTUALS_KEYWORD
    | FORECAST_KEYWORD
    ;
```

**Lexer Tokens** (add to `ChronoconePlanningLanguageLexer.g4`):

```antlr
AS_KEYWORD       : 'as';
WHERE_KEYWORD    : 'where';
ASPECT_KEYWORD   : 'aspect';
ACTUALS_KEYWORD  : 'ACTUALS';
FORECAST_KEYWORD : 'FORECAST';
```

**ParserInputDefinition (Increment 4)**

```go
type ParserInputDefinition struct {
    InputKind      string              // "metric" or "dimension"
    Name           string              // Referenced metric/dimension name
    NameLocation   SourceLocation
    Alias          string              // NEW: Alias from 'as' clause (empty if no alias)
    AliasLocation  SourceLocation      // NEW: Location of alias identifier
    WhereClause    *ParserWhereClause  // NEW: Where clause (nil if not present)
}

type ParserWhereClause struct {
    Conditions []ParserWhereCondition
    Operators  []DefinitionType  // AndOperatorDefinition or OrOperatorDefinition
}

type ParserWhereCondition struct {
    Negated  bool
    AspectEq *AspectEquals
}

type AspectEquals struct {
    Aspect string  // "ACTUALS" or "FORECAST"
    Loc    SourceLocation
}
```

**Listener Implementation** (additions to `cpl_metrics_listener.go`)

```go
func (l *CPLMetricsListener) ExitInputDeclaration(ctx *parser.InputDeclarationContext) {
    metricScope, _ := l.symbolStack.Peek()

    inputKind := ctx.InputKind().GetText()
    inputName := ctx.Identifier(0).GetText()
    inputLoc := makeSourceLocation(ctx.Identifier(0).GetStart())

    inputDef := &ParserInputDefinition{
        InputKind:    inputKind,
        Name:         inputName,
        NameLocation: inputLoc,
    }

    if ctx.AS_KEYWORD() != nil && len(ctx.AllIdentifier()) > 1 {
        alias := ctx.Identifier(1).GetText()
        inputDef.Alias = alias
        inputDef.AliasLocation = makeSourceLocation(ctx.Identifier(1).GetStart())
    }

    if ctx.WHERE_KEYWORD() != nil {
        inputDef.WhereClause = l.parseWhereClause(ctx.WhereExpression())
    }

    metricScope.Define(&inputName, symboltable.InputRefDefinition, inputLoc)
    lastIdx := len(metricScope.Definitions) - 1
    metricScope.Definitions[lastIdx].Data = inputDef
}
```

**Validation During Parsing**

- `as` clause identifier must be a valid identifier
- `where` clause must contain valid aspect references and boolean operators
- Aspect values must be `ACTUALS` or `FORECAST` (enforced by lexer tokens)
- No two aliases may be the same within a single metric's input block (deferred to analysis)

**Limitations**

- Alias uniqueness validation deferred to analysis phase
- Aspect value cross-referencing deferred to analysis
- Boolean expression semantics in where clause deferred to analysis
- No `backward`/`forward` time shifting (added in Increment 5)

### Increment 6: Time Shifting (Backward/Forward)

**Overview**

Increment 6 extends input declarations to support `backward(<instant field>[, <amount>])` and `forward(<instant field>[, <amount>])` clauses for time shifting. The constraint field uses the same field names as [`Instant`](../../types-definitions.md) declarations: `year`, `month`, `day`, `hour`, `minute`, `second`, `day_of_week`, or `week_number`. The optional amount is a numeric literal that defaults to `1.0` if not specified.

**Grammar Changes**

Update the `inputDeclaration` rule and add `backwardClause`, `forwardClause`, and `shiftExpression` in `CPLMetrics.g4`:

```antlr
// An input declaration: with optional backward/forward time shift
inputDeclaration
    : indent inputKind identifier inputAsClause? (WHERE_KEYWORD whereExpression)? backwardClause? forwardClause?
    ;

// Backward time shift clause for input declarations
backwardClause
    : BACKWARD_KEYWORD LPAREN shiftExpression RPAREN
    ;

// Forward time shift clause for input declarations
forwardClause
    : FORWARD_KEYWORD LPAREN shiftExpression RPAREN
    ;

// Shift expression: constraint field name with optional numeric offset
shiftExpression
    : identifier (COMMA NUMBER)?
    ;
```

**Lexer Tokens** (add to `ChronoconePlanningLanguageLexer.g4` — these may already exist from the types implementation):

```antlr
BACKWARD_KEYWORD   : 'backward';
FORWARD_KEYWORD    : 'forward';
```

Note: The Instant constraint field keywords (`year`, `month`, `day`, `hour`, `minute`, `second`, `day_of_week`, `week_number`) are parsed as identifiers in the shift expression. Validity against the known set of Instant constraint field keywords is deferred to analysis.

**Note on Reserved Keywords:** `backward` and `forward` are lexer-level keywords. As with all CPL keywords (e.g., `metrics`, `inputs`, `calculate`, `if`, `when`), they cannot be used as identifiers in any context. This is consistent with the existing CPL keyword reservation pattern.

**ParserInputDefinition and ParserTimeShift (Increment 6)**

```go
type ParserInputDefinition struct {
    InputKind      string
    Name           string
    NameLocation   SourceLocation
    Alias          string
    AliasLocation  SourceLocation
    WhereClause    *ParserWhereClause
    TimeShift      *ParserTimeShift  // NEW: Time shift clause (nil if not present)
}

// ParserTimeShift captures a backward() or forward() clause.
// Uses Instant constraint field keywords for Field.
type ParserTimeShift struct {
    Direction string  // "backward" or "forward"
    Field     string  // Instant constraint field (e.g., "week_number", "day")
    Amount    float64 // Shift amount; defaults to 1.0 if not explicitly provided
}
```

**Note on TimeShift.Field:** The constraint field uses the same keywords as `Instant` declarations (e.g., `backward(week_number)`, `forward(day, 7)`). This ensures consistency across CPL — time shifting refers to the same temporal fields that define `Instant` constraints. The analysis phase validates that the field is recognized.

**Note on TimeShift.Amount:** The shift amount is a simple numeric literal parsed from the optional second argument in the shift expression. If omitted, the amount defaults to `1.0`. Future increments may support arbitrary expression evaluation for the shift amount.

**Listener Implementation** (additions to `cpl_metrics_listener.go`)

The listener uses per-declaration state (`currentTimeShift` and `hasTimeShift`) to capture time shift data and detect mutually exclusive clauses:

```go
// EnterInputDeclaration resets time shift state for each input declaration.
func (l *CPLListener) EnterInputDeclaration(ctx *antlr.InputDeclarationContext) {
    l.currentTimeShift = nil
    l.hasTimeShift = false
}

// ExitBackwardClause extracts the field name and optional numeric amount
// from the shift expression and stores them in currentTimeShift.
func (l *CPLListener) ExitBackwardClause(ctx *antlr.BackwardClauseContext) {
    if l.hasTimeShift {
        l.AddError(fmt.Errorf("input declaration at line %d, column %d cannot have both backward() and forward() time shifts",
            ctx.GetStart().GetLine(), ctx.GetStart().GetColumn()))
        return
    }
    l.hasTimeShift = true

    shiftExpr := ctx.ShiftExpression()
    field := shiftExpr.Identifier().GetText()

    amount := 1.0
    if numberToken := shiftExpr.NUMBER(); numberToken != nil {
        parsedAmount, _ := strconv.ParseFloat(numberToken.GetText(), 64)
        amount = parsedAmount
    }

    l.currentTimeShift = &parsing.ParserTimeShift{
        Direction: "backward",
        Field:     field,
        Amount:    amount,
    }
}

// ExitForwardClause mirrors ExitBackwardClause with direction "forward".
func (l *CPLListener) ExitForwardClause(ctx *antlr.ForwardClauseContext) {
    // Same pattern as ExitBackwardClause with direction "forward"
}

// ExitInputDeclaration attaches the captured time shift (if any) to the
// ParserInputDefinition before defining it in the metric scope.
func (l *CPLListener) ExitInputDeclaration(ctx *antlr.InputDeclarationContext) {
    // ... build inputDef ...
    if l.currentTimeShift != nil {
        inputDef.TimeShift = l.currentTimeShift
        l.currentTimeShift = nil
    }
    // ... define in metric scope ...
}
```

**Validation During Parsing**

- `backward`/`forward` keyword must be followed by parenthesized shift expression
- Constraint field is parsed as an identifier; validity against Instant constraint field keywords is deferred to analysis
- Shift amount, if present, must be a valid numeric literal
- `backward` and `forward` are reserved keywords and cannot be used as identifiers in any CPL context
- An input declaration cannot specify both `backward()` and `forward()` — this is enforced by the parser listener as a semantic error
- Self-references (metric referencing itself via backward/forward alias) are syntactically valid; circularity detection is deferred to analysis

**Limitations**

- Constraint field validity against known Instant fields is deferred to analysis
- Shift amount is limited to simple numeric literals; expression evaluation is deferred to a future increment
- Self-reference cycle detection deferred to analysis (dependency DAG pass)

### Increment 7: Short Form Metrics

**Overview**

Increment 6 introduces the short form syntax for metric definitions: `<metric name> = <calculation>`. The parser converts short form declarations into the equivalent long form `ParserMetricDefinition` internally. This means the symbol table output is identical whether the input uses short form or long form.

**Grammar Changes**

Update `CPLMetrics.g4` to add short form as an alternative within the metrics section:

```antlr
// A metrics section contains metric declarations and/or short form definitions
metricsSection
    : METRICS_KEYWORD
      INDENT
      (metricDeclaration | shortFormDeclaration)+
      DEDENT
    ;

// Short form: <name> = <calculation>
// Uses the shared 'expression' rule from the root grammar
shortFormDeclaration
    : identifier EQ expression
    ;
```

**Listener Implementation** (additions to `cpl_metrics_listener.go`)

```go
func (l *CPLMetricsListener) ExitShortFormDeclaration(ctx *parser.ShortFormDeclarationContext) {
    metricScope := l.symbolStack.CurrentScope()
    if metricScope.ScopeType != symboltable.MetricScope {
        current, _ := l.symbolStack.Peek()
        metricScope = current.AddChild(symboltable.MetricScope)
        l.symbolStack.Push(metricScope)
        defer l.symbolStack.Pop()
    }

    metricName := ctx.Identifier().GetText()
    metricLoc := makeSourceLocation(ctx.Identifier().GetStart())

    // The expression's postfix definitions are already in the scope's Definitions
    // slice (pushed by root listener). We scan them for identifier references
    // to build inputs during analysis.

    // During parsing, the input kind (metric/lookup/dimension) is NOT determined.
    // The parser stores all identifier references with a generic marker.
    // The analysis phase resolves kind using name resolution rules:
    //   1. If name matches a metric → metric input
    //   2. If name matches a lookup → lookup input
    //   3. If name matches a dimension → dimension input
    //   4. Collision → error

    metricDef := &ParserMetricDefinition{
        Name:           metricName,
        ShortForm:      true,  // NEW field
        SourceLocation: metricLoc,
    }

    metricScope.Define(&metricName, symboltable.MetricScopeDefinition, metricLoc)
    lastIdx := len(metricScope.Definitions) - 1
    metricScope.Definitions[lastIdx].Data = metricDef
}
```

**ParserMetricDefinition (Increment 6)**

```go
type ParserMetricDefinition struct {
    Name           string
    ShortForm      bool                    // NEW: True if short form declaration
    Inputs         []ParserInputDefinition
    SourceLocation SourceLocation
}
```

**Validation During Parsing**

- Short form name must be a valid identifier
- Calculation expression must be syntactically valid
- No `as`, `where`, or `backward`/`forward` clauses permitted in short form (enforced by grammar)
- Name collision detection (metric vs dimension vs lookup) is deferred to analysis

**Limitations**

- Name resolution (metric/lookup/dimension) is fully deferred to analysis
- Lookup support is deferred to future design
- No dimension indexing in short form (e.g., `order_lead_time[product]`)

### Increment 8: Full Synthesis with IO Specification Integration

**Overview**

Increment 7 involves no parser changes. All syntax is already supported by Increments 1-6. The parser's responsibility for metrics ends here; all remaining work (dependency DAG construction, topological ordering, dead-code elimination, IO spec integration) is handled in the analysis and synthesis phases.

**Parser Changes**

None. The parser is complete.

**Full ParserMetricDefinition (Through Increment 7)**

```go
// ParserMetricDefinition — final form after all increments.
type ParserMetricDefinition struct {
    Name           string                  // Metric identifier
    ShortForm      bool                    // True if short form declaration (Increment 6)
    Inputs         []ParserInputDefinition  // Input declarations (Increment 1, extended 4-5)
    SourceLocation SourceLocation
}

// ParserInputDefinition — final form after all increments.
type ParserInputDefinition struct {
    InputKind      string              // "metric" or "dimension" (Increment 1)
    Name           string              // Referenced metric/dimension name (Increment 1)
    NameLocation   SourceLocation      // (Increment 1)
    Alias          string              // Alias from 'as' clause (Increment 4)
    AliasLocation  SourceLocation      // (Increment 4)
    WhereClause    *ParserWhereClause  // Where clause (Increment 4)
    TimeShift      *ParserTimeShift    // Time shift clause (Increment 5)
}

// ParserTimeShift — final form.
// Uses Instant constraint field keywords for Field.
type ParserTimeShift struct {
    Direction string  // "backward" or "forward"
    Field     string  // Instant constraint field (e.g., "week_number", "day")
    Amount    float64 // Shift amount; defaults to 1.0 if not explicitly provided
}
```

**Grammar Ownership Summary:**

| Grammar File                                   | Owns                                                                                                                                                                                                                                                                                 |
| ---------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Root grammar (`ChronoconePlanningLanguage.g4`) | `expression`, `conditionalExpr`, `ifBranch`, `elseIfBranch`, `elseBranch`, `whenExpr`, `whenClause`, `defaultClause`, `logicalOrExpr`, `logicalAndExpr`, `equalityExpr`, `relationalExpr`, `additionExpr`, `multiplicationExpr`, `unaryExpr`, `primaryExpr`, `literal`, `identifier` |
| `CPLMetrics.g4`                                | `metricsSection`, `metricDeclaration`, `inputsBlock`, `inputDeclaration`, `inputKind`, `calculateClause`, `shortFormDeclaration`, `whereExpression`, `whereAndExpr`, `wherePrimary`, `aspectEquals`, `aspectValue`, `backwardClause`, `forwardClause`, `shiftExpression`                                |
| Root listener                                  | Expression operator/literal push methods (scope-type-aware via `shouldPushExpressionDefs`); conditional body scopes also included                                                                                                                                                    |
| Metrics listener                               | Section/declaration/input/where/time-shift structural methods; conditional clause definitions and body scope push/pop                                                                                                                                                                |

**Flow:**

```
ANTLR Parse Tree
    │
    ├──► Root Listener (expression defs pushed when scope is MetricScope)
    │
    └──► Metrics Listener (metric structs, input defs, where/time-shift)
              │
              ▼
        Symbol Table (ParserMetricDefinition + postfix defs in MetricScope)
              │
              ▼
   Analysis Phase (WalkSymbolTable → postfix eval → CIL trees)
              │
              ▼
          CIL Trees (one per metric + one per connector)
```

**Validation During Parsing — Full Summary by Increment**

1. **Increment 1:** Metric name syntax, `inputs`/`calculate` block presence, indentation, literal `0` format, input kind keywords
2. **Increment 2:** Arithmetic operator tokens, literal formats (int, float, string), balanced parentheses (ANTLR-enforced), identifier syntax for input refs
3. **Increment 3:** (synthesis-only — no parser changes)
4. **Increment 4:** `if`/`end if` pairing (ANTLR-enforced), `when`/`end when` pairing (ANTLR-enforced), `when` must have default (ANTLR-enforced), comparison/logical operator tokens, Nil/True/False keywords
5. **Increment 5:** `as` alias syntax, `where` clause syntax, aspect value keywords (ACTUALS/FORECAST), boolean combinations in where clause
6. **Increment 6:** `backward`/`forward` clause syntax, Instant constraint field keyword (enforced by grammar — `year`, `month`, `day`, `hour`, `minute`, `second`, `day_of_week`, `week_number`), shift expression syntax, mutual exclusivity of backward/forward on the same input (enforced by parser listener), `backward`/`forward` reserved as keywords
7. **Increment 7:** Short form name and expression syntax
8. **Increment 8:** No changes

## Notes

**Related Documents:**

- [Metrics Design](../../metrics-design.md) — Full metrics design with core rules, expression handling, dependency DAG, and 7 end-to-end increments
- [ERO005 — Metrics Incremental Implementation](../../eros/ERO005_metrics-incremental-implementation.md) — Top-level ERO defining the 7-increment approach
- [Metrics Syntax Requirements](../../../../requirements/cpl/metrics-syntax.md) — Syntax requirements for metric declarations, inputs, and calculations
- [IO Specifications Requirements](../../../../requirements/cpl/io_specs.md) — IO specification requirements for output metric selection
- [Types Definition Design](../../types-definitions.md) — Instant constraint fields reference for time shift clauses
- [Postfix Expression Parsing Pattern](../../../../patterns/postfix-expression-parsing.md) — Pattern for expression handling via postfix notation
- [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md) — Architectural pattern for parsing phase design
- [CPL Architecture — Parsing](../../../architecture/cpl/parsing/toc.md) — Parsing phase architecture
- [ERO001 — Concepts Parsing Phase](ERO001_concepts-parsing-phase.md) — Reference ERO for parsing phase ERO structure and pattern
- [ERO004 — Connectors Parsing Phase](ERO004_connectors-parsing-phase.md) — Reference ERO for connector parsing integration (connectors provide leaf data for metrics)
