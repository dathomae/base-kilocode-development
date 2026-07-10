# ERO007: Aggregators Parsing Phase

## Title

Aggregators Parsing Phase — Data Structures and Incremental Implementation

## Purpose/Goal

Define the parsing phase data structures for aggregator definitions (including `ParserAggregatorDefinition`, the `CPLAggregators.g4` grammar, postfix expression storage, and synthetic key assignment) and document how these structures evolve across each increment of the aggregators incremental implementation as defined in [ERO007 — Aggregators Incremental Implementation](../../eros/ERO007_aggregators-incremental-implementation.md).

## Decision Status

Draft

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic007_cpl-aggregators-implementation.md](../../../../../epics/Epic007_cpl-aggregators-implementation.md)

## Parsing vs Analysis Phase Separation

Following the [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md), aggregator processing is divided into two distinct phases:

### Parsing Phase (This ERO)

The parsing phase walks the ANTLR parse tree and builds the symbol table tree. Its responsibilities are:

1. **Push/pop scope tables** as aggregator section contexts are entered and exited
2. **Define symbols** (aggregator definitions) in the current symbol table under synthetic keys
3. **Attach parse-context data** to `SymbolDefinition.Data` for use by the analysis phase
4. **Push expression operands/operators in postfix order** following the [Postfix Expression Parsing Pattern](../../../../patterns/postfix-expression-parsing.md)

The parser does **not** validate cross-references, resolve dimension names, or check semantic rules beyond syntax.

### Unnamed Data Sources — Synthetic Keys

Aggregators (and spreaders) are **unnamed** — they have no user-visible name like metrics, connectors, or lookups. Instead, they are defined by their dimensional pattern (`from [C1, C2] to [C2]`). The parser assigns a **synthetic key** derived from this pattern for storage in the symbol table:

```go
// Synthetic key format for aggregator entries
func aggregatorSyntheticKey(from []string, to string) string {
    if to == "" {
        return fmt.Sprintf("__aggregator__[%s]->[]", strings.Join(from, ","))
    }
    return fmt.Sprintf("__aggregator__[%s]->[%s]", strings.Join(from, ","), to)
}
```

The synthetic key is never visible to the user; it is an internal compiler identifier used to store and retrieve aggregator definitions from the symbol table.

### Expression Handling: Postfix via Symbol Definitions

Aggregator expressions use the same postfix expression parsing pattern as metrics. The parser does **not** build an expression tree. Instead, as ANTLR `Exit*` listener methods visit expression components, the parser pushes `SymbolDefinition` entries onto the current scope's `Definitions` slice in postfix order.

Aggregator functions (`sum`, `count`, `size`) are **not** reserved words in the language. They are parsed using the generic function call syntax `<identifier>(<parameter list>)` — the same pattern used for any function call. The parser stores the function name and parameter list as raw strings; the analysis phase looks up the function name in a `FunctionRegistry` and inserts the corresponding predefined CIL node type.

For the aggregator expression `sum(Country)`:
- `Country` is pushed as an operand (concept reference)
- `sum(Country)` is pushed as a function call operand (`fncall` kind with name `"sum"` and parameter `["Country"]`)

Resulting Definitions: `[OperandDefinition: concept "Country", OperandDefinition: fncall "sum" ["Country"]]` — the analysis phase recognizes that `"sum"` is a registered aggregator function and replaces the fncall with the appropriate CIL node.

### Grammar Organization

Aggregator parsing uses a dedicated grammar file `CPLAggregators.g4` that defines aggregator-specific rules. Expression rules are imported from the root grammar (`ChronoconePlanningLanguage.g4`) because they may be shared across CPL constructs. The root grammar listener handles expression operator and literal push methods.

### Analysis Phase (Separate ERO)

The analysis phase walks the symbol table tree and assembles the Aggregators lookup map. Its responsibilities are:

1. **Walk the symbol table tree** using `WalkSymbolTable`
2. **Extract `ParserAggregatorDefinition`** from synthetic-key symbol table entries
3. **Perform postfix evaluation** of the expression to compute the dimensional unit effect
4. **Validate dimensional unit compatibility**: `input_units * expression_units = output_units`
5. **Assemble the Aggregators lookup map** keyed by `(from_concept_set, to_concept)`

### Increment 1 Note

The aggregators ERO's Increment 1 (Tensor Matrix Multiplication Infrastructure) has **no parsing phase component**. It is pure Go tensor math with no CPL grammar, listener, or symbol table work. Parsing begins at Increment 2.

## Context

### Situation

The parsing phase must recognize `aggregators` sections and aggregator declarations with `from [<concepts>] to [<concept>]` (or `to []`) pattern and expression body. Aggregators are unnamed — the parser assigns synthetic keys. Expressions use the postfix pattern: concept references and operators are pushed onto the Definitions slice in precedence order.

### Constraints

- Must integrate with the existing ANTLR-based parser infrastructure
- Must follow the [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md)
- Must follow the [Postfix Expression Parsing Pattern](../../../../patterns/postfix-expression-parsing.md) for expression handling
- Parse/analysis separation must be maintained: parser captures syntax only
- Aggregator definitions are stored under synthetic keys — no user-visible name
- The `indent` token enforces readable indentation but does not define scope boundaries
- Expression rules live in the root grammar; aggregator grammar imports them
- Listener methods must be organized one file per grammar file

### Requirements

- Parse `aggregators` sections with `from [<concepts>] to [<concept>]` pattern (or `to []` for scalar reduction) and expression body (see [aggregators-and-spreaders-syntax.md](../../../../requirements/cpl/aggregators-and-spreaders-syntax.md))
- Build raw `ParserAggregatorDefinition` structures stored in `SymbolDefinition.Data`
- Assign synthetic keys to aggregator definitions (the `__aggregator__` prefix convention)
- Push expression operands and operators in postfix order into the aggregator scope's `Definitions` slice
- Support concept references, aggregation operations (`sum`, `count`, `size`), and arithmetic operators (`+`, `-`, `*`, `/`) in expressions
- Support error reporting with source location information

### Assumptions

- Concepts, types, units, measures, connectors, and metrics are already implemented (Epics 001-005 provide the foundation)
- The ANTLR root grammar can be extended with shared expression rules
- A `CPLAggregators.g4` grammar file imports expression rules from the root grammar
- Expression parsing uses the postfix approach as described in the Postfix Expression Parsing Pattern
- The `indent` token from existing CPL lexing is available
- The root grammar listener (which handles expression operator pushing) is scope-type-aware

## Evaluation Criteria / Strategy Fit

1. **Parse/Analysis Separation** (High importance): The parser must not perform semantic validation; all dimensional unit analysis and aggregator selection must happen in analysis.
2. **Grammar Organization** (High importance): One listener file per grammar file. Expression rules in root grammar; aggregator-specific rules in CPLAggregators.g4.
3. **Synthetic Key Convention** (High importance): Synthetic keys must be unique, deterministic, and not collide with any user-definable name.
4. **Incremental Evolvability** (High importance): Data structures must support adding fields in later increments without breaking earlier increments.
5. **Consistency** (Medium importance): Data structures and patterns should evolve consistently with the existing implementation.

## Options Considered

Two approaches were evaluated for synthetic key assignment. **Hash-based keys** (e.g., SHA256 of the from/to pattern) would guarantee uniqueness but are not human-readable for debugging. **Pattern-derived keys** (`__aggregator__[C1,C2]->[C3]`) are human-readable, deterministic, and cannot collide with user-defined names (which cannot start with `__`). Pattern-derived keys were selected.

For expression parsing, **inline expression trees** (building an AST during parsing) was rejected due to inconsistency with the established postfix pattern used for metrics. **Postfix storage** in the Definitions slice was selected to maintain consistency with the metrics implementation.

## Decision

**Chosen Option:** Pattern-derived synthetic keys (`__aggregator__[from]->[to]`) with postfix expression storage in the Definitions slice.

**Rationale:**
Pattern-derived keys are deterministic, human-readable for debugging, and cannot collide with user-defined names. Postfix expression storage is consistent with the metrics implementation and enables reuse of the postfix evaluation machinery in the analysis phase.

## Implications

### Positive Implications

- Deterministic, human-readable synthetic keys
- Reuses the same postfix expression infrastructure as metrics
- Expression rules are shared via the root grammar — no duplication
- The root grammar listener can push definitions in any scope context (metrics, aggregators, spreaders)

### Negative Implications / Trade-offs

- Synthetic keys are implementation details that must not leak into user-visible error messages
- The root grammar listener must be scope-type-aware to correctly route expression definitions
- Increment 1 of the ERO (matrix multiplication infrastructure) has no parsing component — parsing starts at Increment 2

### Validation Strategy

| Increment | Validation |
|-----------|-----------|
| 1 | No parser changes — this is a synthesis-only infrastructure increment |
| 2 | `aggregators` section recognized, `from/to` pattern parsed, fnCall expressions stored as operand definitions, synthetic keys assigned |
| 3 | No parser changes — function CIL node types are defined in analysis |
| 4 | No parser changes — aggregator selection is an analysis-phase operation |
| 5 | No parser changes — sequence optimization is an analysis-phase operation |

## Data Structures by Increment

### Increment 1: Tensor Matrix Multiplication Infrastructure

No parsing phase changes. This increment is pure synthesis — implementing tensor matrix multiplication operations.

### Increment 2: Basic Aggregator Parsing and Validation

**Overview**

Increment 2 introduces aggregator parsing. CPL source files may contain `aggregators` sections with unnamed aggregator declarations. Each declaration specifies a `from` concept list, a `to` concept (or `[]`), and an expression body using the postfix pattern. The parser assigns synthetic keys.

CPL source example:
```
aggregators
    from [Country, Product] to [Product] = sum(Country)
```

Scalar reduction example:
```
aggregators
    from [Country] to [] = sum(Country)
```

**Grammar Changes**

File: [`CPLAggregators.g4`](src/chronocone_planning_language/src/main/antlr/CPLAggregators.g4) (new file, Increment 2)

```antlr
grammar CPLAggregators;

// Import CPL base grammar for token definitions and expression rules
// (indent, NEWLINE, identifiers, LSQUARE, RSQUARE, COMMA, EQUALS, LPAREN, RPAREN, etc.)
// Import expression rules from root grammar for aggregator expression body

// --- Aggregators Section ---
aggregatorsSection
    : AGGREGATORS aggregatorDeclaration+
    ;

// --- Aggregator Declaration ---
// An aggregator has a from/to pattern and an expression body.
aggregatorDeclaration
    : indent FROM LSQUARE identifierList RSQUARE TO LSQUARE toTarget RSQUARE EQUALS expression
    ;

// --- To Target ---
toTarget
    : identifier     // Reduce to a single remaining concept dimension
    |                // Empty alternative: reduce to scalar (no dimensions)
    ;

// --- Identifier List ---
identifierList
    : identifier (COMMA identifier)*
    ;

// --- Function Call (in expressions) ---
// Functions are NOT reserved words. They use the generic <identifier>(<param list>) syntax.
// The parser stores the function name and parameters as raw strings.
// The analysis phase looks up the function name in the FunctionRegistry
// to determine which predefined CIL node type to insert.
fnCall
    : identifier LPAREN identifierList RPAREN
    ;

// --- Lexer tokens ---
AGGREGATORS : 'aggregators';
FROM : 'from';
TO : 'to';
LPAREN : '(';
RPAREN : ')';
```

Note: `LPAREN` and `RPAREN` are registered once in `ChronoconePlanningLanguageLexer.g4` and shared across all grammars. There are no `SUM`, `COUNT`, or `SIZE` lexer tokens — these are ordinary identifiers in the function call position.

**Lexer Token Registration**

File: [`ChronoconePlanningLanguageLexer.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguageLexer.g4) (add to existing file, Increment 2)

```antlr
// Add alongside existing keywords:
AGGREGATORS : 'aggregators';
FROM : 'from';
TO : 'to';
LPAREN : '(';
RPAREN : ')';
```

Note: No `SUM`, `COUNT`, or `SIZE` tokens — these are ordinary identifiers parsed via the `fnCall` rule.

**ParserAggregatorDefinition (Increment 2)**

```go
// ParserAggregatorDefinition represents a parsed aggregator declaration.
// Stored in SymbolDefinition.Data under a synthetic key.
type ParserAggregatorDefinition struct {
    FromConcepts    []string            // Raw concept names for input dimensions
    ToConcept       string              // Raw concept name for target ("" if to [] for scalar reduction)
    ToEmpty         bool                // True if the to clause was empty ([])
    SourceLocation  SourceLocation      // For error reporting
    // Note: Expression operands/operators are stored in the scope's Definitions slice
    // in postfix order, following the Postfix Expression Parsing Pattern.
    // The analysis phase evaluates the postfix sequence to compute the dimensional unit effect.
}
```

**AggregatorsListener (Increment 2)**

File: `src/chronocone_planning_language/src/main/go/parser/AggregatorsListener.go` (new file, Increment 2)

```go
package parser

// AggregatorsListener handles events from CPLAggregators.g4.
type AggregatorsListener struct {
    *BaseCPLAggregatorsListener
    parser *CPLParser  // Reference to main parser for symbol table access
}

// EnterAggregatorsSection pushes a section context.
func (l *AggregatorsListener) EnterAggregatorsSection(ctx *AggregatorsSectionContext) {
    l.parser.pushSectionContext("aggregators")
}

// ExitAggregatorsSection pops the section context.
func (l *AggregatorsListener) ExitAggregatorsSection(ctx *AggregatorsSectionContext) {
    l.parser.popSectionContext()
}

// EnterAggregatorDeclaration pushes a scope for the aggregator.
// This scope will accumulate postfix expression definitions.
func (l *AggregatorsListener) EnterAggregatorDeclaration(ctx *AggregatorDeclarationContext) {
    l.parser.pushScope(AggregatorScopeType)
}

// ExitAggregatorDeclaration creates a ParserAggregatorDefinition and stores it
// under a synthetic key derived from the from/to pattern.
func (l *AggregatorsListener) ExitAggregatorDeclaration(ctx *AggregatorDeclarationContext) {
    // Extract from concept names
    idListCtx := ctx.IdentifierList(0)
    var fromConcepts []string
    for _, id := range idListCtx.AllIdentifier() {
        fromConcepts = append(fromConcepts, id.GetText())
    }

    // Extract to target
    targetCtx := ctx.ToTarget().(*ToTargetContext)
    var toConcept string
    toEmpty := false
    if targetCtx.Identifier() != nil {
        toConcept = targetCtx.Identifier().GetText()
    } else {
        // Empty target (scalar reduction)
        toEmpty = true
    }

    def := &ParserAggregatorDefinition{
        FromConcepts:   fromConcepts,
        ToConcept:      toConcept,
        ToEmpty:        toEmpty,
        SourceLocation: l.parser.getSourceLocation(ctx),
    }

    // Assign synthetic key
    key := aggregatorSyntheticKey(fromConcepts, toConcept)
    if toEmpty {
        key = aggregatorSyntheticKey(fromConcepts, "")
    }

    // Store under synthetic key in symbol table
    l.parser.pushDefinition(
        key,
        AggregatorScopeDefinition,
        def,
        l.parser.getSourceLocation(ctx),
    )

    // Pop the aggregator scope (postfix definitions captured)
    l.parser.popScope()
}
```

**Storage in Symbol Table (Increment 2)**

Aggregator definitions are stored in the symbol table under synthetic keys. Each aggregator declaration creates exactly one symbol table entry — an `AggregatorScopeDefinition` with `ParserAggregatorDefinition` in its `Data` field. The aggregator scope also contains the `Definitions` slice with the postfix expression sequence.

```go
// pushDefinition called by AggregatorsListener.ExitAggregatorDeclaration
func (p *CPLParser) pushDefinition(name string, defType DefinitionType, data interface{}, loc SourceLocation) {
    def := &SymbolDefinition{
        Name:           &name,
        DefinitionType: defType,
        Location:       loc,
        Data:           data,
    }
    p.currentScope.Definitions = append(p.currentScope.Definitions, def)
}
```

For aggregators, `name` is the synthetic key (e.g., `__aggregator__[Country,Product]->[Product]`).

**Expression Postfix Storage (Increment 2)**

The root grammar listener pushes expression components onto the current scope's `Definitions` slice. When the current scope is an `AggregatorScope`, these definitions are pushed into the aggregator's scope:

```go
// In the root grammar listener (showing aggregator context):
func (l *RootGrammarListener) ExitConceptReference(ctx *ConceptReferenceContext) {
    scopeType := l.parser.getCurrentScopeType()
    if scopeType == AggregatorScopeType || scopeType == SpreaderScopeType {
        // Push concept reference as an operand in postfix order
        def := &SymbolDefinition{
            DefinitionType: OperandDefinition,
            Data: &OperandData{
                Kind:  "concept",
                Name:  ctx.Identifier().GetText(),
            },
            Location: l.parser.getSourceLocation(ctx),
        }
        l.parser.currentScope.Definitions = append(l.parser.currentScope.Definitions, def)
    }
    // (For non-aggregator scopes like metrics, metrics-listener handles concept refs)
}

func (l *RootGrammarListener) ExitFnCall(ctx *FnCallContext) {
    scopeType := l.parser.getCurrentScopeType()
    if scopeType == AggregatorScopeType || scopeType == SpreaderScopeType {
        fnName := ctx.Identifier(0).GetText()

        // Extract parameter list from the fnCall's identifierList
        var params []string
        for _, id := range ctx.IdentifierList().AllIdentifier() {
            params = append(params, id.GetText())
        }

        // Push function call as a special operand in postfix order
        def := &SymbolDefinition{
            DefinitionType: OperandDefinition,
            Data: &OperandData{
                Kind:   "fncall",
                Name:   fnName,        // e.g., "sum", "count", "size"
                Params: params,        // e.g., ["Country"]
            },
            Location: l.parser.getSourceLocation(ctx),
        }
        l.parser.currentScope.Definitions = append(l.parser.currentScope.Definitions, def)
    }
}
```

Note: The root grammar listener must be scope-type-aware to correctly route expression definitions to the appropriate context (metrics vs. aggregators vs. spreaders). This is consistent with how the metrics listener handles expression definitions.

**SymSymbol Table Structure (After Parsing, Increment 2)**

```
SymbolTable (root)
├── Definitions:
│   └── [AggregatorScopeDefinition] "__aggregator__[Country,Product]->[Product]" →
│       ParserAggregatorDefinition{
│         FromConcepts: ["Country", "Product"],
│         ToConcept: "Product",
│         ToEmpty: false
│       }
│       // Scope's Definitions slice contains postfix expression:
│       //   [OperandDefinition: concept "Country", OperandDefinition: fncall "sum" ["Country"]]
├── ChildScopes: []
```

**Validation During Parsing (Increment 2)**

- `from` clause must contain at least one identifier
- `to` clause must be either an identifier or the empty list `[]`
- Expression body is syntactically valid (grammar enforces expression rules)
- When `to` is `[]`, the `from` list must contain exactly one concept (validated in analysis, not parsing — the parser accepts any `from` list size and lets analysis report errors)
- `indent` tokens enforce readable indentation

**Limitations (Increment 2)**

- No cross-reference validation: concept names stored as raw strings, resolved in analysis
- No dimensional unit validation (analysis phase)
- No validation of `to []` with multi-concept `from` (analysis phase)
- Aggregators are not yet applied to metric calculations (Increments 4-5)

### Increment 3: Additional Aggregator Function CIL Node Types

No parsing phase changes. The CIL node types for `count`, `size`, `1/sum`, and `1/count` are defined in analysis and generated during synthesis, but the parser already supports them via the generic `fnCall` rule. Any new function name in a function call position is accepted by the parser; the analysis phase validates whether it exists in the FunctionRegistry.

### Increment 4: Aggregator Selection and Integration

No parsing phase changes. Aggregator declarations are already parsed in Increment 2. Aggregator selection happens during analysis.

### Increment 5: Aggregator Sequence Optimization

No parsing phase changes. Sequence optimization is an analysis-phase operation.

### Notes

**Related Documents:**

- [ERO007 — Aggregators Incremental Implementation](../../eros/ERO007_aggregators-incremental-implementation.md) — Parent ERO defining the 5-increment strategy
- [Aggregators Design](../../aggregators.md) — Full aggregators design with definition types and data flow
- [Spreaders Design](../../spreaders.md) — Spreader design (symmetric inverse, shared expression patterns)
- [CPL Requirements — Aggregators and Spreaders](../../../../requirements/cpl/aggregators-and-spreaders-syntax.md) — Syntax requirements
- [Concept Definitions Design](../../concepts-definitions.md) — Concept design (aggregator dimensions reference concepts)
- [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md) — Architectural pattern
- [Postfix Expression Parsing Pattern](../../../../patterns/postfix-expression-parsing.md) — Pattern for expression handling
