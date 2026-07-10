# ERO006: Lookups Parsing Phase

## Title

Lookups Parsing Phase — Data Structures and Incremental Implementation

## Purpose/Goal

Define the parsing phase data structures for lookup definitions (including `ParserLookupDefinition`, the `CPLLookups.g4` grammar, and listener integration) and document how these structures evolve across each increment of the lookups incremental implementation as defined in [ERO006 — Lookups Incremental Implementation](../../eros/ERO006_lookups-incremental-implementation.md).

## Decision Status

Draft

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic006_cpl-lookups-implementation.md](../../../../../epics/Epic006_cpl-lookups-implementation.md)

## Parsing vs Analysis Phase Separation

Following the [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md), lookup processing is divided into two distinct phases:

### Parsing Phase (This ERO)

The parsing phase walks the ANTLR parse tree and builds the symbol table tree. Its responsibilities are:

1. **Push/pop scope tables** as lookup declaration contexts are entered and exited
2. **Define symbols** (lookup names and definition types) in the current symbol table
3. **Attach parse-context data** to `SymbolDefinition.Data` for use by the analysis phase
4. **Store raw string values** for dimension concept names, grain measure names, and celltype measure name — all cross-reference resolution and validation is deferred to analysis

The parser does **not** validate cross-references, resolve dimension names, or check semantic rules beyond syntax.

### Grammar Organization

Lookup parsing uses a dedicated grammar file `CPLLookups.g4` that defines lookup-specific rules (`lookupsSection`, `lookupDeclaration`, `dimensionsClause`, `grainClause`, `celltypeClause`). Lookups do **not** require expression parsing — lookup declarations have no expression body (unlike aggregators and spreaders). This makes the lookup grammar simpler than metrics, aggregators, or spreaders.

The lookup grammar imports the standard CPL base grammar for token definitions (indent, NEWLINE, identifiers, commas, etc.) but does not import expression rules.

### Analysis Phase (Separate ERO)

The analysis phase walks the symbol table tree and assembles `LookupInfo` entries. Its responsibilities are:

1. **Walk the symbol table tree** using `WalkSymbolTable`
2. **Assemble `LookupInfo`** entries from parser definitions
3. **Resolve cross-references**: dimension concept names → `AnalysisConceptDefinition`, grain measure names → `AnalysisMeasureDefinition`, celltype measure name → `AnalysisMeasureDefinition`
4. **Validate semantic rules**: concept existence, grain type constraints, celltype existence
5. **Generate `CILLookupNode`** leaf trees for each lookup

### Symbol Table Storage Pattern

Lookup definitions are stored in the symbol table using the following pattern:

```go
// ParserLookupDefinition is stored in SymbolDefinition.Data during parsing.
// The analysis phase retrieves this data to build LookupInfo.
type ParserLookupDefinition struct {
    Name           string              // Lookup identifier
    Dimensions     []string            // Raw concept names for the key dimensions
    Grain          []string            // Raw measure names for optional grain (empty if no grain)
    Celltype       string              // Raw measure name for the cell type
    HasGrain       bool                // True if grain clause was present in source
    SourceLocation SourceLocation      // For error reporting
}

// SymbolDefinition holds the lookup data in its Data field
type SymbolDefinition struct {
    Name           *string
    DefinitionType DefinitionType      // LookupScopeDefinition
    Location       SourceLocation
    Data           interface{}         // Contains *ParserLookupDefinition
}
```

The lookup scope in the symbol table contains a single `SymbolDefinition` entry (the `LookupScopeDefinition` carrying `ParserLookupDefinition`) and zero child scopes (lookups have no nested elements). This is simpler than metrics, which have child scopes for inputs and calculate bodies.

## Context

### Situation

The parsing phase must recognize `lookups` sections and lookup declarations with `dimensions`, optional `grain`, and `celltype` clauses. Lookup parsing is simpler than metrics, connectors, or aggregators because lookups have no expression body — only structural clauses with name references.

The parser stores raw concept and measure names as strings. All cross-reference resolution (validating that concepts, measures, and grains exist) is deferred to the analysis phase, following the established pattern.

### Constraints

- Must integrate with the existing ANTLR-based parser infrastructure
- Must follow the [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md)
- Must populate the symbol table for use by subsequent phases
- Parse/analysis separation must be maintained: parser captures syntax only; no cross-reference validation
- Must use indentation-aware lexing (the `indent` token from existing CPL lexing enforces readable indentation but does not define scope)
- Listener methods must be organized one file per grammar file (ANTLR Listener Organization Convention)
- The lookup grammar does not import expression rules (no expression body in lookup declarations)

### Requirements

- Parse `lookups` sections and lookup declarations with `dimensions`, optional `grain`, and `celltype` clauses (see [lookups-syntax.md](../../../../requirements/cpl/lookups-syntax.md))
- Build raw `ParserLookupDefinition` structures stored in `SymbolDefinition.Data` during parsing
- Store dimension concept names, grain measure names, and celltype measure name as raw strings
- Support source location tracking for error reporting
- In Increment 3, extend the metrics grammar to recognize `lookup` as a valid input kind in metric `inputs` blocks

### Assumptions

- Concepts, types, units, measures, connectors, and metrics are already implemented (Epics 001-005 provide the foundation)
- The ANTLR grammar infrastructure supports a new `CPLLookups.g4` grammar file
- The indentation-aware lexer from Increment 2 of concepts is available
- A `LookupsListener` class can be generated from `CPLLookups.g4`
- The existing root grammar listener infrastructure can register lookup scope handlers
- Increment 3 changes to the metrics grammar (adding `lookup` input kind) are handled in the existing `CPLMetrics.g4` file and `MetricsListener`

## Evaluation Criteria / Strategy Fit

1. **Parse/Analysis Separation** (High importance): The parser must not perform semantic validation; all cross-references and dependency resolution must happen in analysis.
2. **Grammar Simplicity** (High importance): The lookup grammar should be minimal — no expression rules, no nested child scopes, just structural clauses with name lists.
3. **Incremental Evolvability** (High importance): Data structures must support adding fields in later increments without breaking earlier increments.
4. **Consistency** (Medium importance): Data structures and patterns should evolve consistently with the existing concepts, types, units, connectors, and metrics implementations.
5. **Source Location Accuracy** (Medium importance): All parsed elements must carry source location information for error reporting in subsequent phases.

## Options Considered

Two approaches were evaluated for parser data structures. **Separate structures per increment** (e.g., `ParserLookupDefinitionV1`, `V2`) was rejected due to duplication and inconsistency with the established pattern. **Single evolving `ParserLookupDefinition` structure** that grows with each increment was selected because it matches the pattern used successfully for concepts, types, units/measures, connectors, and metrics — Go zero values provide sensible defaults for unused fields, and documentation and validation rules enforce increment constraints.

A second design decision was whether to handle `grain` as an entirely separate `ParserGrainClause` struct or inline the grain fields in `ParserLookupDefinition`. **Inline fields** (`Grain []string`, `HasGrain bool`) were selected because the grain clause is simple (a comma-separated list of measure names) and does not warrant a separate structure. This keeps the definition compact and follows the same approach used for `ParserConnectorDefinition` grain fields.

## Decision

**Option 1:** Single evolving `ParserLookupDefinition` structure with incrementally added fields, with inline grain fields (no separate `ParserGrainClause` struct).

**Rationale:**
The simplicity of a single structure with inline fields outweighs the benefits of strict versioning or separate sub-structures. The `Grain` field naturally extends to a `[]string` slice when grain is introduced in Increment 2, and `HasGrain` provides a boolean flag to distinguish "grain not present" from "grain present but empty" (which is an error). This is the same approach used successfully for all other CPL parsing phase EROs.

## Implications

### Positive Implications

- Simple, minimal grammar with no expression rules
- Inline fields keep the definition compact
- Consistent with the established pattern across all CPL parse-phase EROs
- Backward compatible (new fields have zero values)
- Lookups have no nested scopes, simplifying symbol table structure

### Negative Implications / Trade-offs

- Must rely on documentation and validation to enforce increment constraints (e.g., `HasGrain` must be false in Increment 1)
- Increment 3 requires modifying the existing `CPLMetrics.g4` grammar, which has cross-cutting implications
- The `lookup` input kind in metrics adds a new `InputKind` value that must be handled consistently across parser and analysis

### Future Considerations

- If lookup declarations later add an expression body (e.g., for default value computation), the grammar would need to import expression rules
- If the Lookup IO Specification is unified with the Connector IO Specification, the parser might require changes to IO spec parsing

### Validation Strategy (Per Increment)

| Increment | Validation |
|-----------|-----------|
| 1 | `lookups` section recognized, `dimensions` and `celltype` parsed, names stored as raw strings, `HasGrain` is false |
| 2 | Optional `grain` clause parsed, `HasGrain` is true when grain present, `Grain` populated with measure names |
| 3 | Metrics grammar recognizes `lookup` as valid input kind, existing `ParserInputDefinition` extended with `InputKind = "lookup"` |
| 4 | No parser changes — lookup references in aggregator expressions are resolved during analysis |
| 5 | No parser changes — lookup references in spreader expressions are resolved during analysis |

## Data Structures by Increment

### Increment 1: Basic Lookup Declarations Without Grain

**Overview**

Increment 1 introduces minimal lookup parsing. CPL source files may contain `lookups` sections with lookup declarations that specify `dimensions` (a list of concept names) and `celltype` (a measure name). No grain clause is supported. The parser stores raw string values; all resolution is deferred to analysis.

CPL source example:
```
lookups
    product_category
        dimensions [Product]
        celltype CategoryLabel
```

**Grammar Changes**

File: [`CPLLookups.g4`](src/chronocone_planning_language/src/main/antlr/CPLLookups.g4) (new file, Increment 1)

```antlr
grammar CPLLookups;

// Import CPL base grammar for token definitions
// (indent, NEWLINE, identifiers, LSQUARE, RSQUARE, COMMA, etc.)

// --- Lookups Section ---
// A lookups section contains one or more lookup declarations.
lookupsSection
    : LOOKUPS lookupDeclaration+
    ;

// --- Lookup Declaration ---
// A lookup declaration consists of a name followed by dimensions and celltype clauses.
// Each clause is on an indented line. Clauses are separated by NEWLINE.
lookupDeclaration
    : indent identifier dimensionsClause celltypeClause
    ;

// --- Dimensions Clause ---
dimensionsClause
    : indent DIMENSIONS LSQUARE identifier (COMMA identifier)* RSQUARE
    ;

// --- Celltype Clause ---
celltypeClause
    : indent CELLTYPE identifier
    ;

// --- Shared parsing rules ---
// (indent, NEWLINE, LSQUARE, RSQUARE, COMMA are defined in base grammar)
// (LOOKUPS, DIMENSIONS, CELLTYPE are lexer tokens defined below)

// --- Lexer tokens ---
LOOKUPS : 'lookups';
DIMENSIONS : 'dimensions';
CELLTYPE : 'celltype';
```

Note: The `indent` token enforces readable indentation but does **not** define scope boundaries. Scoping is determined by the section keyword (`LOOKUPS`) followed by repeated `lookupDeclaration` rules. This is consistent with how `connectorsSection` and `metricsSection` work in the existing grammar — `indent` is a readability constraint, not a scoping mechanism.

**Lexer Token Registration**

File: [`ChronoconePlanningLanguageLexer.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguageLexer.g4) (add to existing file)

```antlr
// Add alongside existing keywords:
LOOKUPS : 'lookups';
DIMENSIONS : 'dimensions';
CELLTYPE : 'celltype';
```

Note: `DIMENSIONS` (plural) is distinct from the existing `DIMENSION` (singular) token used for metric input dimension references.

**ParserLookupDefinition (Increment 1)**

```go
// ParserLookupDefinition represents a parsed lookup declaration.
// Stored in SymbolDefinition.Data during parsing.
type ParserLookupDefinition struct {
    Name           string              // Lookup identifier (e.g., "product_category")
    Dimensions     []string            // Raw concept names for key dimensions
    Celltype       string              // Raw measure name for cell type
    SourceLocation SourceLocation      // For error reporting
}
```

Note: No `Grain` or `HasGrain` fields in Increment 1 — these are added in Increment 2.

**LookupsListener (Increment 1)**

File: `src/chronocone_planning_language/src/main/go/parser/LookupsListener.go` (new file)

```go
package parser

// LookupsListener handles events from CPLLookups.g4.
type LookupsListener struct {
    *BaseCPLLookupsListener
    parser *CPLParser  // Reference to main parser for symbol table access
}

// EnterLookupsSection pushes a context onto the parser's section stack.
func (l *LookupsListener) EnterLookupsSection(ctx *LookupsSectionContext) {
    l.parser.pushSectionContext("lookups")
}

// ExitLookupsSection pops the context from the parser's section stack.
func (l *LookupsListener) ExitLookupsSection(ctx *LookupsSectionContext) {
    l.parser.popSectionContext()
}

// ExitLookupDeclaration is called when a complete lookup declaration has been parsed.
// It creates a ParserLookupDefinition and stores it in the symbol table.
func (l *LookupsListener) ExitLookupDeclaration(ctx *LookupDeclarationContext) {
    lookupName := ctx.Identifier(0).GetText()

    // Extract dimension concept names from dimensionsClause
    dimsCtx := ctx.DimensionsClause().(*DimensionsClauseContext)
    var dimensions []string
    for _, id := range dimsCtx.AllIdentifier() {
        dimensions = append(dimensions, id.GetText())
    }

    // Extract celltype measure name
    celltypeCtx := ctx.CelltypeClause().(*CelltypeClauseContext)
    celltype := celltypeCtx.Identifier().GetText()

    def := &ParserLookupDefinition{
        Name:           lookupName,
        Dimensions:     dimensions,
        Celltype:       celltype,
        SourceLocation: l.parser.getSourceLocation(ctx),
    }

    // Store in symbol table
    l.parser.pushDefinition(
        lookupName,
        LookupScopeDefinition,
        def,
        l.parser.getSourceLocation(ctx),
    )
}
```

**Storage in Symbol Table (Increment 1)**

Lookup definitions are stored in the symbol table under the current scope. Each lookup declaration creates exactly one symbol table entry — a `LookupScopeDefinition` with `ParserLookupDefinition` in its `Data` field. The lookup scope has no child scopes (no nested declarations).

```go
// pushDefinition called by LookupsListener.EnterLookupDeclaration
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

**Example Symbol Table Structure (After Parsing)**

```
SymbolTable (root)
├── Definitions:
│   └── [LookupScopeDefinition] "product_category" → ParserLookupDefinition{
│         Name: "product_category",
│         Dimensions: ["Product"],
│         Celltype: "CategoryLabel"
│       }
├── ChildScopes: []
```

Contrast with a metric scope, which has child scopes for its inputs and calculate body. Lookup scopes are flat — no child scopes exist.

**Validation During Parsing (Increment 1)**

- `dimensions` clause must contain at least one identifier (grammar enforces `identifier (COMMA identifier)*`)
- `celltype` clause must contain exactly one identifier (grammar enforces this)
- `indent` tokens enforce readable indentation (clauses must be indented relative to the lookup name and section keyword) but do not define scope boundaries
- All required clauses present: `dimensions` and `celltype` are mandatory

**Limitations (Increment 1)**

- No `grain` clause support (added in Increment 2)
- No cross-reference validation: dimension concept names and celltype measure name are stored as raw strings; existence is checked in analysis
- No validation of duplicate lookup names (checked in analysis)
- Lookups cannot be referenced in metric inputs (added in Increment 3)
- Lookups are not referenced in aggregator or spreader expressions (added in Increments 4 and 5)

### Increment 2: Lookups with Grain

**Overview**

Increment 2 adds the optional `grain` clause to lookup declarations. A grain clause specifies one or more measure names (comma-separated) whose types must be `Interval` or derived from `Interval`. When present, the lookup becomes temporal — its data varies by time as well as by concept dimensions. When absent, the lookup is timeless.

CPL source example:
```
lookups
    exchange_rate
        dimensions [Currency, Currency]
        grain Week, Month
        celltype ConversionRate
```

**Grammar Changes**

File: [`CPLLookups.g4`](src/chronocone_planning_language/src/main/antlr/CPLLookups.g4) (modify existing, Increment 2)

```antlr
// --- Lookup Declaration (modified) ---
lookupDeclaration
    : indent identifier dimensionsClause grainClause? celltypeClause
    ;  // Note: grainClause is optional (?)

// --- Grain Clause (new rule) ---
grainClause
    : indent GRAIN identifier (COMMA identifier)*
    ;
```

**Lexer Token (new)**

File: [`ChronoconePlanningLanguageLexer.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguageLexer.g4) (add to existing file)

```antlr
GRAIN : 'grain';
```

**ParserLookupDefinition (Increment 2 — fields added)**

```go
type ParserLookupDefinition struct {
    Name           string              // Lookup identifier
    Dimensions     []string            // Raw concept names for key dimensions
    Grain          []string            // NEW: Raw measure names for optional grain (nil if not present)
    HasGrain       bool                // NEW: True if grain clause was present in source
    Celltype       string              // Raw measure name for cell type
    SourceLocation SourceLocation
}
```

`HasGrain` is distinct from `len(Grain) > 0`: `HasGrain` is true when the user explicitly wrote a `grain` clause, even if it contains zero measures (though a zero-measure grain is a validation error in analysis). This allows the analysis phase to distinguish between "grain not specified" (timeless lookup) and "grain specified" (temporal lookup).

**LookupsListener (Increment 2 — modified)**

Modify `ExitLookupDeclaration` to handle the optional grain clause:

```go
func (l *LookupsListener) ExitLookupDeclaration(ctx *LookupDeclarationContext) {
    lookupName := ctx.Identifier(0).GetText()

    // ... (dimensions and celltype extraction unchanged) ...

    // Extract optional grain measure names
    var grain []string
    hasGrain := false
    if grainCtx := ctx.GrainClause(); grainCtx != nil {
        hasGrain = true
        for _, id := range grainCtx.(*GrainClauseContext).AllIdentifier() {
            grain = append(grain, id.GetText())
        }
    }

    def := &ParserLookupDefinition{
        Name:           lookupName,
        Dimensions:     dimensions,
        Grain:          grain,
        HasGrain:       hasGrain,
        Celltype:       celltype,
        SourceLocation: l.parser.getSourceLocation(ctx),
    }

    l.parser.pushDefinition(lookupName, LookupScopeDefinition, def, l.parser.getSourceLocation(ctx))
}
```

**Validation During Parsing (Increment 2)**

- `grain` clause, when present, must contain at least one identifier (grammar enforces `identifier (COMMA identifier)*`)
- `HasGrain` is set to `true` when the grain clause is present, `false` otherwise

**Limitations (Increment 2)**

- No validation that grain measure names exist or have `Interval`-derived types (analysis-phase validation)
- No validation of duplicate grain measure names (if present, analysis phase catches this)
- A zero-length grain when `HasGrain` is true is syntactically valid at the parser level; the analysis phase reports it as an error

### Increment 3: Lookup Integration with Metric Calculations

**Overview**

Increment 3 modifies the existing metrics grammar (`CPLMetrics.g4`) to recognize `lookup` as a new input kind in metric `inputs` blocks, alongside the existing `metric` and `dimension` input kinds. No changes to the lookup grammar are needed.

CPL source example:
```
metrics
    converted_revenue
        inputs
            metric revenue
            lookup currency_conversion
            dimension Currency
        calculate
            revenue * currency_conversion
```

**Grammar Changes (Metrics Grammar)**

File: [`CPLMetrics.g4`](src/chronocone_planning_language/src/main/antlr/CPLMetrics.g4) (modify existing, Increment 3)

```antlr
// --- Input Declaration (modified) ---
// Existing: inputDeclaration : METRIC identifier | DIMENSION identifier ;
// Modified to add a third alternative:
inputDeclaration
    : METRIC identifier
    | DIMENSION identifier
    | LOOKUP identifier      // NEW: lookup input kind
    ;
```

**Lexer Token Registration**

The `LOOKUP` token is added alongside the existing `LOOKUPS` token (added in Increment 1):

File: [`ChronoconePlanningLanguageLexer.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguageLexer.g4)

```antlr
LOOKUP : 'lookup';
```

**ParserInputDefinition (Increment 3 — fields extended)**

The existing `ParserInputDefinition` struct (from metrics parsing) gains a new `InputKind` value:

```go
// ParserInputDefinition represents a parsed metric input reference.
// Existing struct from Increment 1 of metrics parsing.
type ParserInputDefinition struct {
    InputKind      string              // "metric", "dimension", or "lookup" (NEW value)
    Name           string              // Referenced name
    SourceLocation SourceLocation
}
```

No new fields are needed — the existing `InputKind` string field accommodates the new `"lookup"` value naturally.

**MetricsListener (Increment 3 — modified)**

Modify the existing `ExitInputRef` method in `MetricsListener.go` (or equivalent input declaration handler) to handle the `LOOKUP` alternative:

```go
func (l *MetricsListener) ExitInputDeclaration(ctx *InputDeclarationContext) {
    inputDef := &ParserInputDefinition{
        SourceLocation: l.parser.getSourceLocation(ctx),
    }

    if ctx.METRIC() != nil {
        inputDef.InputKind = "metric"
        inputDef.Name = ctx.Identifier(0).GetText()
    } else if ctx.DIMENSION() != nil {
        inputDef.InputKind = "dimension"
        inputDef.Name = ctx.Identifier(0).GetText()
    } else if ctx.LOOKUP() != nil {
        inputDef.InputKind = "lookup"      // NEW
        inputDef.Name = ctx.Identifier(0).GetText()
    }

    l.parser.pushInputDefinition(inputDef)
}
```

**Validation During Parsing (Increment 3)**

- `lookup <name>` is syntactically valid in any metric `inputs` block
- The lookup name is stored as a raw string; existence validation is deferred to analysis
- No parser-level distinction between lookup names and metric names when used as identifiers in the `calculate` body (the analysis phase resolves the identifier against metrics, dimensions, and lookups)

**Limitations (Increment 3)**

- Lookup names used in `calculate` expressions are parsed as ordinary identifiers; the parser does not distinguish them from metric or dimension references
- No validation that the lookup input kind is semantically appropriate for the context (analysis-phase validation)

### Increment 4: Lookups in Aggregator Expressions

**Overview**

No parser changes are required in Increment 4. Aggregator expression parsing already supports concept references and operators via the postfix expression pattern. Lookup references appear as ordinary identifiers in aggregator expressions; the analysis phase resolves them against the Lookups map.

This increment is documented for completeness but contains no new grammar, data structures, or listener methods at the parser level.

**Limitations**

- Lookup names and concept names share the same identifier namespace in expressions; the parser cannot distinguish them. The analysis phase resolves ambiguity by searching lookups before concepts (lookups take precedence when a name exists in both maps, which is an error condition that analysis reports).

### Increment 5: Lookups in Spreader Expressions

**Overview**

No parser changes are required in Increment 5. Spreader expression parsing already supports metric references and concept references via the postfix expression pattern. Lookup references appear as ordinary identifiers in spreader expressions; the analysis phase resolves them against the Lookups map, applying the same dimensional unit analysis used for metric references in spreaders.

This increment is documented for completeness but contains no new grammar, data structures, or listener methods at the parser level.

### Notes

**Related Documents:**

- [ERO006 — Lookups Incremental Implementation](../../eros/ERO006_lookups-incremental-implementation.md) — Parent ERO defining the 5-increment strategy
- [Lookups Design](../../lookups.md) — Full lookups design with definition types and data flow
- [CPL Requirements — Lookups](../../../../requirements/cpl/lookups-syntax.md) — Syntax requirements for lookups
- [Metrics Design](../../metrics-design.md) — Metric design (metrics reference lookups as inputs in Increment 3)
- [Aggregators Design](../../aggregators.md) — Aggregator design (lookups referenced in aggregator expressions in Increment 4)
- [Spreaders Design](../../spreaders.md) — Spreader design (lookups referenced in spreader expressions in Increment 5)
- [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md) — Architectural pattern for parsing phase design
- [Postfix Expression Parsing Pattern](../../../../patterns/postfix-expression-parsing.md) — Pattern for expression handling (relevant for Increment 3 where lookups appear in metric expressions)
