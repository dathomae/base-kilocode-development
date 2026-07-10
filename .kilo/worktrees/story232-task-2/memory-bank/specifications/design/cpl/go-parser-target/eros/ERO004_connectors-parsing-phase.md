# ERO004: Connectors Parsing Phase

## Title

Connectors Parsing Phase

## Purpose/Goal

Define the parsing phase data structures, grammar rules, and listener logic for connector definitions in CPL.

## Implementation Status

Implemented

**Related Epics:**

- [Epic004_cpl-connectors-implementation.md](../../../../../epics/Epic004_cpl-connectors-implementation.md)

## Context

### Situation

The parsing phase must recognize the `connectors` keyword and parse connector declarations. A connector declaration specifies the dimensions, grain, and cell type of the input data. Following the [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md), the parser's primary output is a hierarchical symbol table structure that captures parsed definitions.

### Constraints

- The parser must integrate with the existing ANTLR-based parser infrastructure.
- The parser must populate the symbol table for use by subsequent phases.
- The parser must handle the syntax defined in `memory-bank/specifications/requirements/cpl/connectors.md`.

### Requirements

- Parse `connectors` sections.
- Parse connector declarations within the section.
- Extract the connector name, dimensions list, and celltype (Increment 1).
- Extract one or more grain measure identifiers, preserving their declared order.
- Add the connector to the symbol table.

## Evaluation Criteria

1. **Correctness** (High importance): The parser must correctly identify and extract all parts of a connector declaration.
2. **Symbol Table Integration** (High importance): The connector must be correctly registered in the symbol table for use in the analysis phase.

## Decision

**Chosen Option:** Implement grammar rules, listener methods, and symbol table entries for connectors.

**Rationale:**
The parser will use ANTLR grammar rules to recognize connector declarations and populate the symbol table with a `ConnectorDeclaration` that stores the parsed properties.

### Increment 1: Basic Connector Parsing

**Grammar Changes**

Update `ChronoconePlanningLanguage.g4` to add `connectorsSection` to the statement rule and import the connectors grammar:

```antlr
// Root grammar imports the connectors grammar along with other section grammars
grammar ChronoconePlanningLanguage;

import CPLConcepts, CPLMeasures, CPLTypes, CPLUnits, CPLConnectors;

options {
    tokenVocab = ChronoconePlanningLanguageLexer;
}

// Statements are the top-level elements of a CPL file
statement
    : conceptsSection
    | typesSection
    | unitsSection
    | measuresSection
    | connectorsSection;
```

Create new file `CPLConnectors.g4`:

```antlr
// CPLConnectors.g4 - Grammar for CPL Connectors section

parser grammar CPLConnectors;

options {
    tokenVocab = ChronoconePlanningLanguageLexer;
}

// ============================================================================
// CONNECTORS SECTION
// ============================================================================

// A connectors section contains one or more connector declarations
connectorsSection: CONNECTORS connectorDeclaration+;

// A connector declaration.
// The rule uses a Kleene star so that a single ConnectorDeclarationContext
// may represent zero or more declarations (the loop matches zero times when
// only a newline follows the indent).
connectorDeclaration:
    indent ( identifier EQUALS DIMENSIONS LBRACKET identifierList RBRACKET GRAIN identifierList CELLTYPE identifier )*;

// A comma-separated list of identifiers
identifierList: identifier (COMMA identifier)*;
```

**Lexer Tokens** (defined in `ChronoconePlanningLanguageLexer.g4`):

```antlr
// Keywords
CONNECTORS : 'connectors';
DIMENSIONS : 'dimensions';
GRAIN      : 'grain';
CELLTYPE   : 'celltype';

// Punctuation (EQUALS, LBRACKET, RBRACKET, COMMA already exist)
```

**Scope and Definition Types**

```go
// ScopeType values for connector-related scopes
const (
    // ... existing scopes ...
    ConnectorsScope ScopeType = iota // Scope for connectors section
    ConnectorScope                   // Scope for individual connector definition
)

// DefinitionType values for connector definitions
const (
    // ... existing definition types ...
    ConnectorsScopeDefinition DefinitionType = iota // Definition type for connectors section scope
    ConnectorScopeDefinition                        // Definition type for individual connector scope
    GrainDefinition                                 // Definition type for a grain identifier within a connector
)
```

**ConnectorDeclaration** (core structure stored in symbol table)

```go
// ConnectorDeclaration is stored in SymbolTable.Data during parsing.
// The analysis phase retrieves this data to build the CIL representation.
type ConnectorDeclaration struct {
    // Name is the connector identifier.
    Name string
    // NameLocation is the source location where the connector name appears.
    NameLocation SourceLocation
    // DimensionNames is the ordered list of dimension concept names declared by the connector.
    DimensionNames []string
    // DimensionLocations contains the source location for each dimension in DimensionNames.
    DimensionLocations []SourceLocation
    // GrainName is the first grain measure name.
    // Deprecated: kept for backward compatibility; use GrainNames and GrainLocations.
    GrainName string
    // GrainLocation is the source location where the first grain measure name appears.
    // Deprecated: kept for backward compatibility; use GrainNames and GrainLocations.
    GrainLocation SourceLocation
    // GrainNames is the ordered list of grain measure names declared by the connector.
    // The parser preserves the declared order of identifiers.
    GrainNames []string
    // GrainLocations contains the source location for each grain in GrainNames.
    GrainLocations []SourceLocation
    // CellTypeName is the celltype measure name declared by the connector.
    CellTypeName string
    // CellTypeLocation is the source location where the celltype measure name appears.
    CellTypeLocation SourceLocation
}
```

**Listener Implementation**

The unified `CPLListener` handles the ANTLR callbacks and populates the symbol table.

```go
// EnterConnectorsSection pushes a new ConnectorsScope onto the symbol stack.
func (l *CPLListener) EnterConnectorsSection(ctx *antlr.ConnectorsSectionContext) {
    // Get the current table and create a child for the Connectors scope
    current, err := l.symbolStack.Peek()
    if err != nil {
        l.AddError(fmt.Errorf("symbol stack error entering Connectors section: %w", err))
        return
    }

    connectorsTable := current.AddChild(symboltable.ConnectorsScope)
    l.symbolStack.Push(connectorsTable)
}

// ExitConnectorsSection pops the ConnectorsScope from the symbol stack.
func (l *CPLListener) ExitConnectorsSection(ctx *antlr.ConnectorsSectionContext) {
    _, err := l.symbolStack.Pop()
    if err != nil {
        l.AddError(fmt.Errorf("symbol stack error exiting Connectors section: %w", err))
    }
}

// EnterConnectorDeclaration pushes a new ConnectorScope onto the symbol stack.
func (l *CPLListener) EnterConnectorDeclaration(ctx *antlr.ConnectorDeclarationContext) {
    // Get the current table and create a child for the Connector scope
    current, err := l.symbolStack.Peek()
    if err != nil {
        l.AddError(fmt.Errorf("symbol stack error entering Connector declaration: %w", err))
        return
    }

    l.symbolStack.Push(current.AddChild(symboltable.ConnectorScope))
}

// ExitConnectorDeclaration extracts the parsed data and adds it to the symbol table.
func (l *CPLListener) ExitConnectorDeclaration(ctx *antlr.ConnectorDeclarationContext) {
    // Get the current ConnectorScope before popping
    connectorScope, err := l.symbolStack.Peek()
    if err != nil {
        l.AddError(fmt.Errorf("symbol stack error exiting Connector declaration: %w", err))
        return
    }

    // Direct identifier order: [0] connector name, [last] celltype.
    allIdentifiers := ctx.AllIdentifier()

    // The grammar allows zero or more connector declarations per
    // ConnectorDeclarationContext. When there are no identifiers, remove
    // the parser-created empty ConnectorScope before popping.
    if len(allIdentifiers) == 0 {
        removeChildSymbolTable(connectorScope.Parent, connectorScope)
        l.symbolStack.Pop()
        return
    }

    // A valid connector declaration requires at least 2 direct identifiers.
    if len(allIdentifiers) < 2 {
        l.AddError(fmt.Errorf("connector declaration has insufficient identifiers"))
        removeChildSymbolTable(connectorScope.Parent, connectorScope)
        l.symbolStack.Pop()
        return
    }

    connectorName := allIdentifiers[0].GetText()
    connectorLocation := l.makeSourceLocation(allIdentifiers[0].GetStart())

    // Extract grain identifiers from IdentifierList(1).
    grainList := ctx.IdentifierList(1)
    if grainList == nil || len(grainList.AllIdentifier()) == 0 {
        l.AddError(fmt.Errorf("connector declaration missing grain"))
        removeChildSymbolTable(connectorScope.Parent, connectorScope)
        l.symbolStack.Pop()
        return
    }

    grainIdentifiers := grainList.AllIdentifier()
    grain := grainIdentifiers[0].GetText()
    grainLocation := l.makeSourceLocation(grainIdentifiers[0].GetStart())

    // Build GrainNames and GrainLocations for multi-grain support.
    grainNames := make([]string, 0, len(grainIdentifiers))
    grainLocations := make([]symboltable.SourceLocation, 0, len(grainIdentifiers))
    for _, gi := range grainIdentifiers {
        grainNames = append(grainNames, gi.GetText())
        grainLocations = append(grainLocations, l.makeSourceLocation(gi.GetStart()))
    }

    cellTypeIdentifier := allIdentifiers[len(allIdentifiers)-1]
    cellTypeName := cellTypeIdentifier.GetText()
    cellTypeLocation := l.makeSourceLocation(cellTypeIdentifier.GetStart())

    var dimNames []string
    var dimLocations []symboltable.SourceLocation
    if identifierList := ctx.IdentifierList(0); identifierList != nil {
        dimensionIdentifiers := identifierList.AllIdentifier()
        dimNames = make([]string, 0, len(dimensionIdentifiers))
        dimLocations = make([]symboltable.SourceLocation, 0, len(dimensionIdentifiers))
        for _, dimensionIdentifier := range dimensionIdentifiers {
            dimNames = append(dimNames, dimensionIdentifier.GetText())
            dimLocations = append(dimLocations, l.makeSourceLocation(dimensionIdentifier.GetStart()))
        }
    }

    connectorScope.Data = &symboltable.ConnectorDeclaration{
        Name:               connectorName,
        NameLocation:       connectorLocation,
        DimensionNames:     dimNames,
        DimensionLocations: dimLocations,
        // Deprecated: kept for backward compatibility; use GrainNames/GrainLocations.
        GrainName:        grain,
        GrainLocation:    grainLocation,
        GrainNames:       grainNames,
        GrainLocations:   grainLocations,
        CellTypeName:     cellTypeName,
        CellTypeLocation: cellTypeLocation,
    }

    // Register each grain identifier as a GrainDefinition in the connector scope.
    for i, gi := range grainIdentifiers {
        grainName := gi.GetText()
        _ = connectorScope.Define(&grainName, symboltable.GrainDefinition, grainLocations[i])
    }

    // Pop the Connector scope from the symbol stack.
    _, err = l.symbolStack.Pop()
    if err != nil {
        l.AddError(fmt.Errorf("symbol stack error exiting Connector declaration: %w", err))
        return
    }
}
```

**Limitations:**

- The parser only extracts the names (strings) of the dimensions, grain, and celltype. It does not validate that these names refer to existing concepts or measures. That validation is deferred to the analysis phase.
- Connector grain parsing accepts one or more comma-separated grain measure identifiers. The grammar enforces at least one identifier via `identifierList`.
- Parser and symbol table state preserve the declared order of grain identifiers in `GrainNames` and `GrainLocations`.
- Legacy single-grain fields (`GrainName` and `GrainLocation`) mirror the first grain identifier while the ordered slice fields (`GrainNames` and `GrainLocations`) are authoritative.
- The parser does not validate that every grain measure is based on `Interval`. Semantic validation of grain measure types belongs to the analysis phase.
- Synthesis multi-grain output generation is not described in this document; it is deferred to [`Story160_cpl-synthesizer-connectors-time.md`](../../../../../../stories/Story160_cpl-synthesizer-connectors-time.md).

**Value Provided:**

The parser can successfully read connector declarations and make their structure available to the analysis phase via the symbol table.

## Implications

### Positive Implications

- Clean separation of concerns: parsing extracts structure, analysis validates semantics.
- The symbol table is populated with the necessary information for the analysis phase to perform its checks.

### Negative Implications / Trade-offs

- None identified.

## Notes

**Related Documents:**

- [CPL Requirements - Connectors](../../../../requirements/cpl/connectors.md)
- [Symbol Table Design](../symbol-table-design.md)
