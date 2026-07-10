# ERO008: Spreaders Parsing Phase

## Title

Spreaders Parsing Phase — Data Structures and Incremental Implementation

## Purpose/Goal

Define the parsing phase data structures for spreader definitions (including `ParserSpreaderDefinition`, the `CPLSpreaders.g4` grammar, postfix expression storage, synthetic key assignment, and function call handling) and document how these structures evolve across each increment of the spreaders incremental implementation as defined in [ERO008 — Spreaders Incremental Implementation](../../eros/ERO008_spreaders-incremental-implementation.md).

## Decision Status

Accepted

## Implementation Status

Increment 1 Implemented — `CPLSpreaders.g4` grammar, `SpreadersListener`, `ParserSpreaderDefinition`, synthetic key assignment, and postfix expression storage are implemented and tested. Metric references in spreader expressions deferred to later increments.

**Related Epics:**

- [Epic008_cpl-spreaders-implementation.md](../../../../../epics/Epic008_cpl-spreaders-implementation.md)

## Parsing vs Analysis Phase Separation

Following the [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md), spreader processing is divided into two distinct phases:

### Parsing Phase (This ERO)

The parsing phase walks the ANTLR parse tree and builds the symbol table tree. Its responsibilities are:

1. **Push/pop scope tables** as spreader section contexts are entered and exited
2. **Define symbols** (spreader definitions) in the current symbol table under synthetic keys
3. **Attach parse-context data** to `SymbolDefinition.Data` for use by the analysis phase
4. **Push expression operands/operators in postfix order** following the [Postfix Expression Parsing Pattern](../../../../patterns/postfix-expression-parsing.md)

The parser does **not** validate cross-references, resolve dimension names, or check semantic rules beyond syntax.

### Unnamed Data Sources — Synthetic Keys

Spreaders (and aggregators) are **unnamed** — they have no user-visible name. Instead, they are defined by their dimensional pattern (`from [C1] to [C1, C2]`). The parser assigns a **synthetic key** derived from this pattern:

```go
// Synthetic key format for spreader entries
func spreaderSyntheticKey(from []string, to []string) string {
    return fmt.Sprintf("__spreader__[%s]->[%s]", strings.Join(from, ","), strings.Join(to, ","))
}

// For from [] to [C]:
//   "__spreader__[]->[Country]"
```

The synthetic key is never visible to the user.

### Function Calls — Not Reserved Words

Like aggregators, spreader functions (`size`, `sum`, `count`, `1/size`, `1/sum`, `1/count`) are **not** reserved words. They use the generic `<identifier>(<paramList>)` syntax via the `fnCall` grammar rule. The parser stores function calls as `OperandDefinition` with kind `"fncall"`. The analysis phase looks up the function name in the spreader `FunctionRegistry` to determine the appropriate CIL node type.

The spreader FunctionRegistry is separate from the aggregator FunctionRegistry — while they share many function names, the CIL node types and dimensional unit semantics differ for spreading vs. aggregating.

### Expression Handling: Postfix via Symbol Definitions

Spreader expressions use the same postfix expression parsing pattern as metrics and aggregators. Function calls, concept references, metric references (as identifiers), and arithmetic operators are all pushed onto the Definitions slice in postfix order.

For the spreader expression `Product / customers`:
- `Product` is pushed as an operand (concept reference)
- `customers` is pushed as an operand (identifier — resolved as a metric reference during analysis)
- `/` is pushed as an operator

Resulting Definitions: `[OperandDefinition: concept "Product", OperandDefinition: identifier "customers", OperatorDefinition: "/"]`

### Grammar Organization

Spreader parsing uses a dedicated grammar file `CPLSpreaders.g4`. Expression rules are imported from the root grammar. The `fnCall` rule is shared with aggregators (defined once in the root grammar or in `CPLAggregators.g4` and imported).

### Analysis Phase (Separate ERO)

The analysis phase walks the symbol table tree and assembles the Spreaders lookup map. Its responsibilities include postfix evaluation, dimensional unit validation, metric reference resolution with implicit aggregator determination, and CIL node generation.

## Context

### Situation

The parsing phase must recognize `spreaders` sections and spreader declarations with `from [<concepts>] to [<concepts>]` (or `from [] to [<C>]`) pattern and expression body. Spreaders are unnamed — the parser assigns synthetic keys. Expressions use the postfix pattern with function calls, concept references, and metric references.

Spreaders differ from aggregators in one key parsing aspect: expressions may reference **metrics** by name (for weighted distributions like `Product / customers`). Metric references are parsed as ordinary identifiers — the parser does not distinguish them from concept references. The analysis phase resolves the identifier against metrics, concepts, and lookups in priority order.

### Constraints

- Must integrate with the existing ANTLR-based parser infrastructure
- Must follow the [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md)
- Must follow the [Postfix Expression Parsing Pattern](../../../../patterns/postfix-expression-parsing.md) for expression handling
- Parse/analysis separation must be maintained: parser captures syntax only
- Spreader definitions are stored under synthetic keys
- Function calls use `<identifier>(<paramList>)` syntax — no reserved function name tokens
- Metric references in expressions are parsed as identifiers — no parser-level distinction from concept references
- The `indent` token enforces readable indentation but does not define scope boundaries
- Expression rules live in the root grammar; spreader grammar imports them
- Matrix multiplication infrastructure is already available from ERO007 Increment 1 (no parser work needed for this)

### Requirements

- Parse `spreaders` sections with `from [<concepts>] to [<concepts>]` pattern (or `from [] to [<C>]` for scalar spreading) and expression body (see [aggregators-and-spreaders-syntax.md](../../../../requirements/cpl/aggregators-and-spreaders-syntax.md))
- Build raw `ParserSpreaderDefinition` structures stored in `SymbolDefinition.Data`
- Assign synthetic keys to spreader definitions (`__spreader__` prefix convention)
- Push expression operands and operators in postfix order into the spreader scope's Definitions slice
- Support function calls (`fnCall` rule), concept references, identifier references (metrics), and arithmetic operators (`*`, `/`) in expressions
- Support error reporting with source location information

### Assumptions

- Concepts, types, units, measures, connectors, metrics, and aggregators (ERO007 Increment 2) are already implemented
- The ANTLR root grammar can be extended with shared expression rules including the `fnCall` rule
- A `CPLSpreaders.g4` grammar file imports expression rules from the root grammar
- Expression parsing uses the postfix approach
- The root grammar listener is scope-type-aware and can route fnCall definitions to the spreader scope
- Tensor matrix multiplication infrastructure (both reduction and expansion) is available from ERO007

## Evaluation Criteria

1. **Parse/Analysis Separation** (High importance): The parser must not perform semantic validation; all dimensional unit analysis, metric resolution, and spreader selection must happen in analysis.
2. **Grammar Organization** (High importance): One listener file per grammar file. Expression rules and fnCall in root grammar; spreader-specific rules in CPLSpreaders.g4.
3. **Synthetic Key Convention** (High importance): Synthetic keys must be unique, deterministic, and not collide with any user-definable name.
4. **Consistency with Aggregators** (High importance): The parsing pattern should mirror the aggregator parsing pattern, differing only where spreaders require additional features (metric references in expressions).
5. **Incremental Evolvability** (Medium importance): Data structures must support adding fields in later increments.

## Options Considered

Two approaches were evaluated for handling metric references in spreader expressions. **Separate identifier kinds** (e.g., the parser tagging identifiers as "metric" vs. "concept") was rejected due to parse/analysis separation — the parser should not resolve what a name refers to. **Generic identifiers** (all references parsed as identifiers, resolved during analysis against metrics → concepts → lookups in priority order) was selected to maintain the separation.

For function calls, **reserved word tokens** (`SUM`, `COUNT`, `SIZE`) were rejected in favor of **generic fnCall syntax** to keep the language extensible — new functions can be added to the FunctionRegistry without grammar changes.

## Decision

**Chosen Option:** Generic identifiers for metric/concept references, generic fnCall syntax for function calls, pattern-derived synthetic keys (`__spreader__[from]->[to]`).

**Rationale:**
Maintaining parse/analysis separation requires the parser to be agnostic about whether an identifier is a metric, concept, or lookup reference. The generic fnCall syntax enables new functions to be added without grammar changes — only the FunctionRegistry in analysis needs updating.

## Implications

### Positive Implications

- Extensible function support — new spreader functions require no grammar changes
- Consistent with the aggregator parsing pattern (same fnCall rule, same synthetic key convention)
- Parse/analysis separation is preserved for identifier resolution
- Expression infrastructure (postfix, fnCall) is shared with aggregators

### Negative Implications / Trade-offs

- The analysis phase must resolve identifier ambiguity (metric vs. concept vs. lookup)
- Synthetic keys for spreaders differ from aggregators in that the `to` side is a list (not a single concept)
- The root grammar listener must be scope-type-aware and route definitions to both aggregator and spreader scopes

### Validation Strategy

| Increment | Validation |
|-----------|-----------|
| 1 | `spreaders` section recognized, `from/to` pattern parsed, fnCall + concept + identifier references in postfix, synthetic keys assigned |
| 2 | No parser changes — spreader selection is analysis-phase |
| 3 | No parser changes — metric references already parsed as identifiers in Increment 1 |
| 4 | No parser changes — interleaving and optimization are analysis-phase |

## Data Structures by Increment

### Increment 1: Basic Spreader Parsing and Validation

**Overview**

Increment 1 introduces spreader parsing. CPL source files may contain `spreaders` sections with unnamed spreader declarations using the `from/to` pattern and expression body.

CPL source example:
```
spreaders
    from [Product] to [Country, Product] = Product * 1/size(Country) * Country
```

Scalar spreading example:
```
spreaders
    from [] to [Country] = 1/size(Country) * Country
```

**Grammar Changes**

File: [`CPLSpreaders.g4`](src/chronocone_planning_language/src/main/antlr/CPLSpreaders.g4) (new file, Increment 1)

```antlr
grammar CPLSpreaders;

// Import CPL base grammar for token definitions and expression rules
// (indent, NEWLINE, identifiers, LSQUARE, RSQUARE, COMMA, EQUALS, etc.)
// Expression rules and fnCall are imported from the root grammar.

// --- Spreaders Section ---
spreadersSection
    : SPREADERS spreaderDeclaration+
    ;

// --- Spreader Declaration ---
spreaderDeclaration
    : indent FROM LSQUARE identifierList? RSQUARE TO LSQUARE identifierList? RSQUARE EQUALS expression
    ;

// --- Identifier List (imported/shared) ---
// identifierList: identifier (COMMA identifier)*
// Defined in the root grammar or CPLAggregators.g4

// --- Lexer tokens ---
SPREADERS : 'spreaders';
FROM : 'from';     // Already registered in ERO007 lexer tokens
TO : 'to';          // Already registered
```

Note: `FROM` and `TO` lexer tokens are shared with `CPLAggregators.g4` and are registered once in `ChronoconePlanningLanguageLexer.g4`. The `fnCall` rule is defined in the root grammar (or imported from `CPLAggregators.g4`). No `SUM`, `COUNT`, or `SIZE` tokens — these are ordinary identifiers.

**ParserSpreaderDefinition (Increment 1)**

```go
// ParserSpreaderDefinition represents a parsed spreader declaration.
// Stored in SymbolDefinition.Data under a synthetic key.
type ParserSpreaderDefinition struct {
    FromConcepts    []string            // Raw concept names for input dimensions (nil if from [])
    ToConcepts      []string            // Raw concept names for output dimensions
    FromEmpty       bool                // True if the from clause was empty ([])
    SourceLocation  SourceLocation      // For error reporting
    // Expression operands/operators are stored in the scope's Definitions slice
    // in postfix order. Includes fncall operands for functions like size(C),
    // identifier operands for metric references, and arithmetic operators.
}
```

**SpreadersListener (Increment 1)**

File: `src/chronocone_planning_language/src/main/go/parser/SpreadersListener.go` (new file, Increment 1)

```go
package parser

type SpreadersListener struct {
    *BaseCPLSpreadersListener
    parser *CPLParser
}

func (l *SpreadersListener) EnterSpreadersSection(ctx *SpreadersSectionContext) {
    l.parser.pushSectionContext("spreaders")
}

func (l *SpreadersListener) ExitSpreadersSection(ctx *SpreadersSectionContext) {
    l.parser.popSectionContext()
}

func (l *SpreadersListener) EnterSpreaderDeclaration(ctx *SpreaderDeclarationContext) {
    l.parser.pushScope(SpreaderScopeType)
}

func (l *SpreadersListener) ExitSpreaderDeclaration(ctx *SpreaderDeclarationContext) {
    // Extract from concepts
    var fromConcepts []string
    fromEmpty := true
    if idList := ctx.IdentifierList(0); idList != nil {
        fromEmpty = false
        for _, id := range idList.AllIdentifier() {
            fromConcepts = append(fromConcepts, id.GetText())
        }
    }

    // Extract to concepts
    var toConcepts []string
    if idList := ctx.IdentifierList(1); idList != nil {
        for _, id := range idList.AllIdentifier() {
            toConcepts = append(toConcepts, id.GetText())
        }
    }

    def := &ParserSpreaderDefinition{
        FromConcepts:   fromConcepts,
        ToConcepts:     toConcepts,
        FromEmpty:      fromEmpty,
        SourceLocation: l.parser.getSourceLocation(ctx),
    }

    // Assign synthetic key
    key := spreaderSyntheticKey(fromConcepts, toConcepts)
    l.parser.pushDefinition(key, SpreaderScopeDefinition, def, l.parser.getSourceLocation(ctx))
    l.parser.popScope()
}
```

**Expression Postfix Storage (Increment 1)**

Function calls are pushed by the root grammar listener as `OperandDefinition` with kind `"fncall"`:

```go
func (l *RootGrammarListener) ExitFnCall(ctx *FnCallContext) {
    scopeType := l.parser.getCurrentScopeType()
    if scopeType == AggregatorScopeType || scopeType == SpreaderScopeType {
        fnName := ctx.Identifier(0).GetText()
        var params []string
        if idList := ctx.IdentifierList(); idList != nil {
            for _, id := range idList.AllIdentifier() {
                params = append(params, id.GetText())
            }
        }
        def := &SymbolDefinition{
            DefinitionType: OperandDefinition,
            Data: &OperandData{
                Kind:   "fncall",
                Name:   fnName,
                Params: params,
            },
            Location: l.parser.getSourceLocation(ctx),
        }
        l.parser.currentScope.Definitions = append(l.parser.currentScope.Definitions, def)
    }
}
```

Identifier references (which may be metrics, concepts, or lookups) are pushed as `OperandDefinition` with kind `"identifier"`:

```go
func (l *RootGrammarListener) ExitIdentifier(ctx *IdentifierContext) {
    scopeType := l.parser.getCurrentScopeType()
    if scopeType == SpreaderScopeType {
        def := &SymbolDefinition{
            DefinitionType: OperandDefinition,
            Data: &OperandData{
                Kind: "identifier",
                Name: ctx.GetText(),
            },
            Location: l.parser.getSourceLocation(ctx),
        }
        l.parser.currentScope.Definitions = append(l.parser.currentScope.Definitions, def)
    }
}
```

**Symbol Table Structure (After Parsing, Increment 1)**

```
SymbolTable (root)
├── Definitions:
│   └── [SpreaderScopeDefinition] "__spreader__[Product]->[Country,Product]" →
│       ParserSpreaderDefinition{
│         FromConcepts: ["Product"],
│         ToConcepts: ["Country", "Product"],
│         FromEmpty: false
│       }
│       // Scope's Definitions slice contains postfix expression:
│       //   [OperandDefinition: concept "Product",
│       //    OperandDefinition: fncall "size" ["Country"],
│       //    OperatorDefinition: "/",
│       //    OperandDefinition: fncall "1/size" ["Country"],
│       //    OperatorDefinition: "*"]
│       // (postfix for: Product * 1/size(Country) * Country)
├── ChildScopes: []
```

**Validation During Parsing (Increment 1)**

- `from` clause may be empty `[]` or contain at least one identifier
- `to` clause must contain at least one identifier; its size must be exactly one more than the `from` clause (validated in analysis)
- All concepts in `from` must be in `to` (validated in analysis)
- Expression body is syntactically valid
- `indent` tokens enforce readable indentation

**Limitations (Increment 1)**

- No cross-reference validation — concept/metric names resolved in analysis
- Metric references in expressions are parsed as generic identifiers; resolution is deferred to analysis
- No dimensional unit validation (analysis)
- Spreaders not yet applied to metric calculations (Increments 2-4)

### Increment 2: Spreader Selection and Integration

No parsing phase changes. Spreader declarations already parsed in Increment 1. Selection happens during analysis.

### Increment 3: Metric References in Spreader Expressions

No parsing phase changes. Metric references are already parsed as generic identifiers in Increment 1. Resolution to metrics, concepts, and lookups happens during analysis.

### Increment 4: Spreader/Aggregator Interleaving and Optimization

No parsing phase changes. Interleaving and optimization are analysis-phase operations.

### Notes

**Related Documents:**

- [ERO008 — Spreaders Incremental Implementation](../../eros/ERO008_spreaders-incremental-implementation.md) — Parent ERO
- [Spreaders Design](../../spreaders.md) — Full spreaders design
- [Aggregators Design](../../aggregators.md) — Aggregator design (shared patterns)
- [CPL Requirements — Aggregators and Spreaders](../../../../requirements/cpl/aggregators-and-spreaders-syntax.md)
- [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md)
- [Postfix Expression Parsing Pattern](../../../../patterns/postfix-expression-parsing.md)
