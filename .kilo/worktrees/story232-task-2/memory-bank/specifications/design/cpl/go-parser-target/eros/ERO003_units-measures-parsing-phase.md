# ERO003: Units and Measures Parsing Phase

## Title

Units and Measures Parsing Phase - Data Structures and Incremental Implementation

## Purpose/Goal

Define the parsing phase data structures for units and measures definitions and document how these structures evolve across each increment of the units and measures incremental implementation.

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic003: Units and Measures Implementation](../../../epics/Epic003_units-and-measures-implementation.md) - Epic tracking the implementation of units and measures

**Related Stories:**

- [Story153: CPL Units Parser - Basic Unit Definitions Reimplementation](../../../../stories/Story153_cpl-units-parser-basic-reimplementation.md) - Increment 1 parser implementation (replaces superseded Story141)
- ~~Story141~~ through ~~Story150~~ - Superseded stories (old implementation incompatible with types system)

## Context

### Situation

The parsing phase is responsible for converting CPL source text into a **symbol table tree** containing units and measures definitions. Following the [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md), the parser's primary output is not an Abstract Syntax Tree (AST) but rather a hierarchical symbol table structure that captures parsed definitions. The parsing phase must support the incremental implementation of units and measures features as defined in the main [Units and Measures Incremental Implementation](../../eros/ERO003_units-measures-incremental-implementation.md) ERO.

Units and measures have a dependency relationship with concepts (unit values come from concept values) and types (measures combine types with units). The parser must handle these dependencies while maintaining the symbol table-driven approach.

**Key Principle:** The parser builds the symbol table tree; the analysis phase walks this tree to assemble the final AST structures.

### Constraints

- Must integrate with existing ANTLR-based parser infrastructure
- Must populate the symbol table for use by subsequent phases
- Must support error reporting with source location information
- Parser data structures must evolve to support each increment
- Unit definitions depend on concept names and values being available
- Measure definitions depend on type implementation (particularly the `Number` type)

### Requirements

- Capture raw unit definitions from the AST
- Capture raw measure definitions from the AST
- Store unit group names and their referenced concept names
- Store measure group names, type references, and unit group associations
- Support compound unit expressions using `*`, `/`, and `^` operators
- Support unit conversion clauses with pattern/calculation pairs
- Support measure conversion clauses with context-dependent patterns
- Enable validation during parsing where possible

### Assumptions

- The parser uses ANTLR4 with a Go target
- Symbol table entries can be extended incrementally
- Source location information is available for error reporting
- Concept names and values will be available for unit definitions
- The `Number` type and type derivation will be available for measure definitions

## Migration Note: Old vs New Unit Grammar

### Overview

This section documents the architectural shift in how unit declarations are processed during parsing. The ERO003 implementation follows the [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md), which fundamentally changes how unit definitions flow from the parser to the analysis phase.

### Old Approach (Pre-Epic003)

In the original design, the parser was responsible for building AST nodes directly during parse-time:

| Aspect                    | Old Approach                                                                 |
| ------------------------- | ---------------------------------------------------------------------------- |
| **Parser Responsibility** | Build `UnitDecl` AST nodes directly in listener methods                      |
| **Data Flow**             | ANTLR Parse Tree → `UnitDecl` AST nodes                                      |
| **Identifier Handling**   | Type switches based on parent context to determine declaration type          |
| **State Management**      | Transient state flags (`currentUnit`) synchronized across enter/exit methods |

**Problems with the Old Approach:**

1. **Parsing/Analysis Boundary Violation**: The parser was building semantic structures rather than just capturing raw syntax
2. **Tight Coupling**: Grammar changes required changes to AST node construction
3. **Fragile State Management**: Required careful synchronization of transient state across listener methods
4. **Poor Separation**: Mixing syntax parsing with semantic AST building

**Example Old Pattern (ANTI-PATTERN - Do Not Use):**

```go
// ANTI-PATTERN: Parser building AST nodes directly
func (l *CPLListener) ExitUnitDefinition(ctx *parser.UnitDefinitionContext) {
    l.currentUnit.Name = ctx.Identifier(0).GetText()
    l.currentUnit.ConceptName = ctx.Identifier(1).GetText()
    // Parser is building semantic structure - WRONG!
    l.units = append(l.units, l.currentUnit)
    l.currentUnit = nil
}
```

### New Approach (ERO003)

The ERO003 implementation follows the symbol table-driven parsing pattern with clear phase separation:

| Aspect                    | New Approach                                                                                    |
| ------------------------- | ----------------------------------------------------------------------------------------------- |
| **Parser Responsibility** | Push/pop scope tables; define symbols; attach `ParserUnitDefinition` to `SymbolDefinition.Data` |
| **Data Flow**             | ANTLR Parse Tree → Symbol Table Tree → Analysis Phase → `AnalysisUnitDefinition`                |
| **Identifier Handling**   | Generic identifier registration; scope type provides context                                    |
| **State Management**      | Symbol table stack provides implicit context; no transient state flags                          |

**Benefits of the New Approach:**

1. **Clean Phase Separation**: Parser captures raw syntax; analysis phase builds semantic structures
2. **Loose Coupling**: Grammar changes only require scope type updates
3. **Stack-Based Context**: Symbol table stack provides natural, hierarchical context
4. **Extensibility**: New declaration types add new scope/walker types without modifying existing code

**Example New Pattern (CORRECT - Use This):**

```go
// CORRECT: Parser populates symbol table only
func (l *UnitsListener) ExitUnitDefinition(ctx *parser.UnitDefinitionContext) {
    unitName := ctx.Identifier(0).GetText()
    conceptName := ctx.Identifier(1).GetText()

    // Build ParserUnitDefinition (raw parsed data only)
    unitDef := &ParserUnitDefinition{
        Name:           unitName,
        ConceptName:    conceptName,
        SourceLocation: sourceLocation,
    }

    // Get current scope table
    current, _ := l.symbolStack.Peek()

    // Create SymbolDefinition and attach ParserUnitDefinition via Data field
    symDef := SymbolDefinition{
        Name:           &unitName,
        DefinitionType: UnitScopeDefinition,
        Location:       sourceLocation,
        Data:           unitDef,  // Key: ParserUnitDefinition attached here
    }
    current.Definitions = append(current.Definitions, symDef)
}
```

### Transformation Comparison Table

| Transformation        | Old Approach            | New Approach                                          | Notes                                        |
| --------------------- | ----------------------- | ----------------------------------------------------- | -------------------------------------------- |
| **Parser Output**     | `UnitDecl` AST nodes    | Symbol table tree with `ParserUnitDefinition` entries | Parser output is now a tree of symbol tables |
| **Unit Name Storage** | `UnitDecl.Name` field   | `SymbolDefinition.Name` pointer                       | Name stored at symbol level                  |
| **Concept Reference** | Resolved during parsing | Raw string in `ParserUnitDefinition.ConceptName`      | Resolution happens in analysis phase         |
| **Analysis Input**    | `[]UnitDecl`            | `*SymbolTable` tree                                   | Analysis walks symbol table tree             |
| **Analysis Output**   | (same AST nodes)        | `[]AnalysisUnitDefinition`                            | Analysis builds resolved structures          |
| **Key Attachment**    | N/A (direct building)   | `SymbolDefinition.Data` field                         | `ParserUnitDefinition` attached here         |

### Critical Implementation Detail: SymbolDefinition.Data Field

The `SymbolDefinition.Data` field is the bridge between parsing and analysis phases:

```go
type SymbolDefinition struct {
    Name           *string
    DefinitionType DefinitionType
    Location       SourceLocation
    Data           interface{}  // Contains *ParserUnitDefinition or *ParserMeasureDefinition
}
```

**How It Works:**

1. **Parse Time**: The parser creates a `ParserUnitDefinition` with raw parsed data (names as strings, not resolved references)
2. **Storage**: The `ParserUnitDefinition` is attached to `SymbolDefinition.Data` as an `interface{}`
3. **Analysis Time**: The analysis phase retrieves the `ParserUnitDefinition` from `Data`, resolves references, and builds `AnalysisUnitDefinition`

**Type Assertion in Analysis Phase:**

```go
// Analysis phase retrieves and uses ParserUnitDefinition
for _, def := range unitsScope.Definitions {
    if def.DefinitionType == UnitScopeDefinition {
        // Type assertion to retrieve ParserUnitDefinition
        parserDef := def.Data.(*ParserUnitDefinition)

        // Build AnalysisUnitDefinition with resolved references
        analysisDef := AnalysisUnitDefinition{
            Name:        *def.Name,
            Concept:     resolveConcept(parserDef.ConceptName),  // Resolution happens here
            SourceLocation: parserDef.SourceLocation,
        }
        units = append(units, analysisDef)
    }
}
```

### Migration Path for Implementers

When implementing ERO003 parser stories:

1. **Do NOT build AST nodes directly** in listener methods
2. **Create `ParserUnitDefinition`** with raw parsed data
3. **Attach to `SymbolDefinition.Data`** when defining symbols
4. **Let the analysis phase** resolve references and build `AnalysisUnitDefinition`

See the [Analysis Phase ERO](../../go-analysis/eros/ERO003_units-measures-analysis-phase.md) for details on how parser structures are transformed into analysis structures.

## Parsing vs Analysis Phase Separation

Following the [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md), units and measures processing is divided into two distinct phases:

### Parsing Phase (This ERO)

The parsing phase walks the ANTLR parse tree and builds the symbol table tree. Its responsibilities are:

1. **Push/pop scope tables** as declaration contexts are entered and exited
2. **Define symbols** (names and definition types) in the current symbol table
3. **Attach parse-context data** to `SymbolDefinition.Data` for use by the analysis phase

The parser does **not** build AST nodes directly. Instead, it populates the symbol table with raw parsed data.

### Analysis Phase (Separate ERO)

The analysis phase walks the symbol table tree and assembles the final AST structures. Its responsibilities are:

1. **Walk the symbol table tree** using [`WalkSymbolTable`](../../symbol-table-design.md)
2. **Assemble declarations** from symbol definitions
3. **Build analysis-level structures** like `AnalysisUnitDefinition` and `AnalysisMeasureDefinition`

### Parsing vs Analysis Phase Boundary

**Critical Design Principle**: The parser captures **raw syntactic information** (what was written), not **semantic structures** (what it means). The analysis phase transforms raw parser structures into resolved analysis structures.

| Aspect              | Parser-Time (This ERO)                       | Analysis-Time (ERO003 Analysis Phase)                 |
| ------------------- | -------------------------------------------- | ----------------------------------------------------- |
| **Data**            | Raw strings, token sequences                 | Resolved references, computed structures              |
| **Unit Expression** | `RawUnitExpression` - captures source syntax | `Expression` - semantic tree with resolved references |
| **Unit References** | Unit names as strings                        | Pointers to `AnalysisUnitDefinition` records          |
| **Validation**      | Syntax only (valid identifiers)              | Semantic (unit existence, compatibility)              |
| **Tree Building**   | None - flat capture                          | Expression tree with operator nodes                   |

The analysis phase ERO (`ERO003_units-measures-analysis-phase.md`) defines the transformation from parser structures to analysis structures.

### Analysis Phase Assembly of Conversion Pairs

The analysis phase walks the symbol table tree to assemble `ConversionPair` and `MeasureConversionPair` structures. See the [Analysis Phase ERO](../../go-analysis/eros/ERO003_units-measures-analysis-phase.md) for details on conversion pair assembly.

**Key Points:**

1. **Parser Responsibility**: Push scope tables for conversions, patterns, calculations; store raw expressions as definitions
2. **Analysis Responsibility**: Walk the tree, pair patterns with calculations, build semantic structures
3. **Pairing Semantics**: Patterns and calculations are paired by their order within the ConversionScope (first pattern with first calculation, etc.)

### Symbol Table Storage Pattern

Unit and measure definitions are stored in the symbol table using the following pattern:

```go
// ParserUnitDefinition is stored in SymbolDefinition.Data
// Contains raw parsed information - names as strings, not resolved references.
// Note: Conversions are NOT captured here; they are child symbol table scopes.
type ParserUnitDefinition struct {
    Name             string
    ConceptName      string              // Raw concept name (resolved in analysis)
    SourceLocation   SourceLocation
}

// ParserMeasureDefinition is stored in SymbolDefinition.Data
// Contains raw parsed information - the parser does NOT build Expression trees.
// Note: Conversions are NOT captured here; they are child symbol table scopes.
type ParserMeasureDefinition struct {
    Name             string
    TypeName         string
    UnitGroupName    string              // Increment 2: simple unit group name
    // For Increment 3+, use RawUnitExpression defined below
    SourceLocation   SourceLocation
}

// SymbolDefinition holds the unit/measure data in its Data field
type SymbolDefinition struct {
    Name             *string
    DefinitionType   DefinitionType  // UnitScopeDefinition or MeasureScopeDefinition
    Location         SourceLocation
    Data             interface{}     // Contains *ParserUnitDefinition or *ParserMeasureDefinition
}
```

### Symbol Table Hierarchy for Conversions

Following the symbol table-driven parsing pattern, conversion clauses create nested scopes rather than being stored as structures within the unit definition. This properly separates parser-time (scope creation) from analysis-time (pair assembly).

**Parser-Time Behavior:**

1. When entering a unit definition: Push a `UnitScope` symbol table
2. When entering a conversion section: Push a `ConversionScope` symbol table (child of UnitScope)
3. When entering a pattern clause: Push a `PatternScope` symbol table (child of ConversionScope)
4. When pattern expression is parsed: Create `SymbolDefinition` with expression data, append to PatternScope definitions
5. When entering a calculation clause: Push a `CalculationScope` symbol table (child of ConversionScope)
6. When calculation expression is parsed: Create `SymbolDefinition` with expression data, append to CalculationScope definitions

**Analysis-Time Behavior:**

The analyzer walks the tree and assembles `ConversionPair` structures by pairing PatternScope definitions with CalculationScope definitions within each ConversionScope.

### Symbol Table Structure Example

After parsing a CPL file with units and measures definitions, the symbol table tree looks like this:

```
GlobalScope (root)
├── ConceptsScope
│   ├── SymbolDefinition: Name="Currency"
│   └── SymbolDefinition: Name="Length"
├── UnitsScope
│   ├── SymbolDefinition: Name="Currency"
│   │   Data: &ParserUnitDefinition{
│   │       Name: "Currency",
│   │       ConceptName: "Currency",
│   │   }
│   └── SymbolDefinition: Name="Length"
│       Data: &ParserUnitDefinition{
│           Name: "Length",
│           ConceptName: "Length",
│       }
└── MeasuresScope
    └── SymbolDefinition: Name="Temperature"
        Data: &ParserMeasureDefinition{
            Name: "Temperature",
            TypeName: "Number",
            UnitGroupName: "TemperatureScale",  // Increment 2: simple name
        }
```

After Increment 3 (compound units), the measure definition's scope contains expression component definitions:

```
└── MeasuresScope
    └── SymbolDefinition: Name="Speed"
        Data: &ParserMeasureDefinition{
            Name: "Speed",
            TypeName: "Number",
        }
        Definitions slice (postfix order for "Length / Time"):
          [0] ParserUnitReference{Name: "Length"}
          [1] ParserUnitReference{Name: "Time"}
          [2] ParserOperator{Operator: OpDivide}
```

The analysis phase walks this tree, extracts the `ParserUnitDefinition` and `ParserMeasureDefinition` structures from `SymbolDefinition.Data`, and builds the final analysis AST nodes.

## Data Structures by Increment

### Increment 1: Basic Unit Definitions

**Grammar Changes**

Update [`ChronoconePlanningLanguage.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguage.g4:1) to add `unitsSection` to the statement rule:

```antlr
// Import the units grammar
import CPLUnits, CPLConcepts;

// Statements are the top-level elements of a CPL file
statement: unitsSection | conceptsSection;
```

Create new file `CPLUnits.g4`:

```antlr
// CPLUnits.g4 - Grammar for CPL Units section (Increment 1)

grammar CPLUnits;

import ChronoconePlanningLanguage; // For identifier rule

options {
    tokenVocab = ChronoconePlanningLanguageLexer;
}

// ============================================================================
// UNITS SECTION
// ============================================================================

// A units section contains one or more unit group definitions
unitsSection: UNITS_KEYWORD unitDefinition+;

// A unit definition: UnitGroup = ConceptName
unitDefinition: identifier EQ identifier;
```

**Lexer Tokens** (add to [`ChronoconePlanningLanguageLexer.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguageLexer.g4:1)):

```antlr
// Keywords
UNITS_KEYWORD    : 'units';

// Punctuation (already defined in concepts)
// LBRACKET, RBRACKET, EQ, COMMA reused from concepts
```

**Scope and Definition Types**

```go
// ScopeType values for unit-related scopes
type ScopeType int

const (
    GlobalScope ScopeType = iota
    ConceptsScope           // Scope for concepts section
    UnitsScope              // NEW: Scope for units section
    UnitScope               // NEW: Scope for individual unit definition
    ConversionScope         // NEW: Scope for conversion section (Increment 4)
    PatternScope            // NEW: Scope for pattern clause (Increment 4)
    CalculationScope        // NEW: Scope for calculation clause (Increment 4)
    MeasuresScope           // Added in Increment 2
    MeasureScope            // NEW: Scope for individual measure definition (Increment 5)
)

// DefinitionType values for unit definitions
type DefinitionType int

const (
    UnitDefinition DefinitionType = iota
    UnitsScopeDefinition        // NEW: Definition type for units section scope
    UnitScopeDefinition         // NEW: Definition type for individual unit group scope
    ConversionScopeDefinition   // NEW: Definition type for conversion scope (Increment 4)
    PatternScopeDefinition      // NEW: Definition type for pattern scope (Increment 4)
    PatternExpressionDefinition // NEW: Definition type for pattern expression (Increment 4)
    CalculationScopeDefinition  // NEW: Definition type for calculation scope (Increment 4)
    CalculationExpressionDefinition // NEW: Definition type for calculation expression (Increment 4)
    MeasureScopeDefinition      // Added in Increment 2
    MeasureConversionScopeDefinition // NEW: Definition type for measure conversion scope (Increment 5)
)
```

**ParserUnitDefinition** (Core structure stored in symbol table)

```go
// ParserUnitDefinition is stored in SymbolDefinition.Data during parsing.
// The analysis phase retrieves this data to build AnalysisUnitDefinition.
type ParserUnitDefinition struct {
    Name            string         // Unit group identifier
    ConceptName     string         // Referenced concept name (values define units)
    SourceLocation  SourceLocation // For error reporting
}
```

**Storage in Symbol Table**

The `ParserUnitDefinition` is attached to the symbol table via the `SymbolDefinition.Data` field. This is the key mechanism that enables the symbol table-driven parsing pattern:

1. The parser creates a `ParserUnitDefinition` containing raw parsed data (names as strings)
2. The `ParserUnitDefinition` is attached to `SymbolDefinition.Data` as an `interface{}`
3. The analysis phase retrieves it via type assertion and builds `AnalysisUnitDefinition`

```go
// In the listener, when a unit definition is completed:
func (l *UnitsListener) ExitUnitDefinition(ctx *parser.UnitDefinitionContext) {
    unitName := ctx.Identifier(0).GetText()
    conceptName := ctx.Identifier(1).GetText()

    // Build the ParserUnitDefinition from parsed data
    // This contains ONLY raw strings - no resolved references
    unitDef := &ParserUnitDefinition{
        Name:           unitName,
        ConceptName:    conceptName,
        SourceLocation: sourceLocation,
    }

    // Get the current scope table (UnitsScope)
    current, _ := l.symbolStack.Peek()

    // Create SymbolDefinition and attach ParserUnitDefinition via Data field
    // The Data field is an interface{} that holds *ParserUnitDefinition
    symDef := SymbolDefinition{
        Name:           &unitName,
        DefinitionType: UnitScopeDefinition,
        Location:       sourceLocation,
        Data:           unitDef,  // Key: ParserUnitDefinition attached here
    }
    current.Definitions = append(current.Definitions, symDef)
}
```

**How the Data Field Works:**

| Field                             | Type             | Purpose                                               |
| --------------------------------- | ---------------- | ----------------------------------------------------- |
| `SymbolDefinition.Name`           | `*string`        | Points to the unit group name                         |
| `SymbolDefinition.DefinitionType` | `DefinitionType` | Identifies this as `UnitScopeDefinition`              |
| `SymbolDefinition.Location`       | `SourceLocation` | Source location for error reporting                   |
| `SymbolDefinition.Data`           | `interface{}`    | Contains `*ParserUnitDefinition` with raw parsed data |

**Analysis Phase Retrieval:**

The analysis phase retrieves the `ParserUnitDefinition` from `Data` using type assertion:

```go
// Analysis phase walks the symbol table tree
for _, def := range unitsScope.Definitions {
    if def.DefinitionType == UnitScopeDefinition {
        // Type assertion to retrieve ParserUnitDefinition from Data field
        parserDef := def.Data.(*ParserUnitDefinition)

        // Build AnalysisUnitDefinition with resolved references
        analysisDef := AnalysisUnitDefinition{
            Name:        *def.Name,
            Concept:     resolveConcept(parserDef.ConceptName),  // Resolved here
            SourceLocation: parserDef.SourceLocation,
        }
        units = append(units, analysisDef)
    }
}
```

See the [Migration Note section above](#migration-note-old-vs-new-unit-grammar) for a complete comparison of old vs new approaches.

**Symbol Table Integration**

- When entering a `units` section:
  - A new symbol table with `ScopeType = UnitsScope` is pushed onto the stack
  - A `SymbolDefinition` with `DefinitionType = UnitsScopeDefinition` is added to this table
- When a unit definition is encountered:
  - A `SymbolDefinition` with `DefinitionType = UnitScopeDefinition` is added
  - The `Name` field points to the unit group name
  - The `Data` field holds a `*ParserUnitDefinition`
- Each unit group has a unique name within the compilation unit
- Duplicate unit group definitions in the same file are errors

**Example Symbol Table Structure (After Parsing)**

Given this CPL source:

```cpl
concepts
    Currency = [Values[USD, EURO, YEN]]
    Length = [Values[Meter, Kilometer, Centimeter]]

units
    Currency = Currency
    Length = Length
```

The resulting symbol table tree is:

```
GlobalScope (root)
├── ConceptsScope
│   ├── SymbolDefinition: Name="Currency"
│   └── SymbolDefinition: Name="Length"
└── UnitsScope
    ├── SymbolDefinition
    │   Name: "Currency"
    │   DefinitionType: UnitScopeDefinition
    │   Data: &ParserUnitDefinition{
    │       Name: "Currency",
    │       ConceptName: "Currency",
    │   }
    └── SymbolDefinition
        Name: "Length"
        DefinitionType: UnitScopeDefinition
        Data: &ParserUnitDefinition{
            Name: "Length",
            ConceptName: "Length",
        }
```

**Validation During Parsing**

- Unit group name syntax validation (valid identifier)
- Concept name syntax validation (valid identifier)
- Note: Concept existence validation deferred to analysis phase

**Limitations**

- No conversion clauses (pattern/calculation)
- No compound units using operators
- Single units section per file
- No measure definitions yet
- Concept existence not validated during parsing

### Increment 2: Basic Measure Definitions

**Grammar Changes**

Update [`ChronoconePlanningLanguage.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguage.g4:1) to add `measuresSection`:

```antlr
// Import the measures grammar
import CPLUnits, CPLConcepts, CPLMeasures;

// Statements are the top-level elements of a CPL file
statement: unitsSection | conceptsSection | measuresSection;
```

Create new file `CPLMeasures.g4`:

```antlr
// CPLMeasures.g4 - Grammar for CPL Measures section (Increment 2)

grammar CPLMeasures;

import ChronoconePlanningLanguage; // For identifier rule

options {
    tokenVocab = ChronoconePlanningLanguageLexer;
}

// ============================================================================
// MEASURES SECTION
// ============================================================================

// A measures section contains one or more measure definitions
measuresSection: MEASURES_KEYWORD measureDefinition+;

// A measure definition: MeasureGroup = TypeName(unit=UnitGroupName)
measureDefinition: identifier EQ identifier LPAREN UNIT EQUALS identifier RPAREN;
```

**Lexer Tokens** (add to [`ChronoconePlanningLanguageLexer.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguageLexer.g4:1)):

```antlr
// Keywords
MEASURES_KEYWORD : 'measures';
UNIT             : 'unit';

// Punctuation
LPAREN           : '(';
RPAREN           : ')';
```

**ParserMeasureDefinition** (Core structure stored in symbol table)

```go
// ParserMeasureDefinition is stored in SymbolDefinition.Data during parsing.
// The analysis phase retrieves this data to build AnalysisMeasureDefinition.
// At Increment 2, measures reference simple unit groups by name.
type ParserMeasureDefinition struct {
    Name            string         // Measure group identifier
    TypeName        string         // Type name (e.g., "Number")
    UnitGroupName   string         // Referenced unit group name (raw string, not resolved)
    SourceLocation  SourceLocation // For error reporting
}
```

**Note on Phase Separation:** The parser captures `UnitGroupName` as a raw string. The analysis phase resolves this to an `*AnalysisUnitDefinition` pointer. This maintains the separation between parsing-time (raw capture) and analysis-time (semantic resolution).

**Symbol Table Integration**

- When entering a `measures` section:
  - A new symbol table with `ScopeType = MeasuresScope` is pushed onto the stack
  - A `SymbolDefinition` with `DefinitionType = MeasuresScopeDefinition` is added
- When a measure definition is encountered:
  - A `SymbolDefinition` with `DefinitionType = MeasureScopeDefinition` is added
  - The `Name` field points to the measure group name
  - The `Data` field holds a `*ParserMeasureDefinition`
- Each measure group has a unique name within the compilation unit

**Example Symbol Table Structure (After Parsing)**

Given this CPL source:

```cpl
concepts
    TemperatureScale = [Values[C, F]]

units
    TemperatureScale = TemperatureScale

measures
    Temperature = Number(unit=TemperatureScale)
```

The resulting symbol table tree is:

```
GlobalScope (root)
├── ConceptsScope
│   └── SymbolDefinition: Name="TemperatureScale"
├── UnitsScope
│   └── SymbolDefinition: Name="TemperatureScale"
│       Data: &ParserUnitDefinition{...}
└── MeasuresScope
    └── SymbolDefinition: Name="Temperature"
        DefinitionType: MeasureScopeDefinition
        Data: &ParserMeasureDefinition{
            Name: "Temperature",
            TypeName: "Number",
            UnitGroupName: "TemperatureScale",
        }
```

**Validation During Parsing**

- Measure group name syntax validation
- Type name syntax validation
- Unit group name syntax validation
- Note: Type and unit group existence validation deferred to analysis phase

**Limitations**

- Only `Number` type and `Number`-derived types supported (validated in analysis)
- No measure conversions
- No compound unit expressions in measure declarations
- Single measures section per file
- No validation of unit group existence during parsing

### Increment 3: Compound Units and Unit Algebra Tree

**Grammar Changes**

Update [`ChronoconePlanningLanguage.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguage.g4:1) to add expression rules (shared across CPL):

```antlr
// ChronoconePlanningLanguage.g4 - Root grammar (Increment 3 additions)

grammar ChronoconePlanningLanguage;

import CPLUnits, CPLConcepts;

options {
    tokenVocab = ChronoconePlanningLanguageLexer;
}

// ... existing rules ...

// ============================================================================
// EXPRESSION RULES (shared across CPL for units, measures, and metrics)
// ============================================================================

// Expression supports *, /, and ^ operators for unit algebra
expression
    : term                              // Single unit or parenthesized expression
    | expression MULT term              // Multiplication: Length * Time
    | expression DIV term               // Division: Length / Time
    ;

term
    : factor                            // Simple unit reference or power
    | term POW INTEGER                  // Power: Meter^2, Second^-1
    ;

factor
    : identifier                        // Simple unit reference
    | LPAREN expression RPAREN          // Parenthesized expression
    ;
```

Update `CPLUnits.g4` (expression rules are now imported from main grammar):

```antlr
// CPLUnits.g4 - Grammar for CPL Units section (Increment 3)

parser grammar CPLUnits;

options {
    tokenVocab = ChronoconePlanningLanguageLexer;
}

// ============================================================================
// UNITS SECTION (unchanged from Increment 1)
// ============================================================================

unitsSection: UNITS_KEYWORD unitDefinition+;
unitDefinition: identifier EQ identifier;
```

Update `CPLMeasures.g4` to use expressions:

```antlr
// CPLMeasures.g4 - Grammar for CPL Measures section (Increment 3)

parser grammar CPLMeasures;

options {
    tokenVocab = ChronoconePlanningLanguageLexer;
}

// ============================================================================
// MEASURES SECTION
// ============================================================================

measuresSection: MEASURES_KEYWORD measureDefinition+;

// A measure definition with expression support for compound units
measureDefinition: identifier EQ identifier LPAREN UNIT EQUALS expression RPAREN;
```

**Lexer Tokens** (add to [`ChronoconePlanningLanguageLexer.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguageLexer.g4:1)):

```antlr
// Operators for expressions (shared across CPL)
MULT             : '*';
DIV              : '/';
POW              : '^';
INTEGER          : [0-9]+;
```

**Note on Grammar Organization:**
The `expression`, `term`, and `factor` rules are defined in the top-level [`ChronoconePlanningLanguage.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguage.g4:1) rather than in `CPLUnits.g4` or `CPLMeasures.g4` because these arithmetic expression rules will be reused across multiple CPL sections, including:

- Compound unit expressions in measures (Increment 3)
- Unit conversion calculations (Increment 4)
- Measure conversion calculations (Increment 5)
- Metrics formulas (future implementation)

This follows the pattern of defining shared grammar rules in the root grammar file that can be imported by specialized section grammars.

**Parser Types for Expression Components (Increment 3)**

Following the [Postfix Expression Parsing via Symbol Table Stack Pattern](../../../../patterns/postfix-expression-parsing.md), the parser does NOT build expression trees. Instead, listener methods push definitions onto the symbol table's definitions slice in postfix order, naturally respecting operator precedence through the grammar structure and ANTLR's depth-first traversal.

```go
// ParserUnitReference captures a simple unit name reference (e.g., "Meter", "Length").
// Pushed to definitions slice when ExitIdentifier is called.
type ParserUnitReference struct {
    Name string
}

// ParserOperator captures an operator token.
// Pushed to definitions slice when Exit<Operator> is called.
type ParserOperator struct {
    Operator OperatorType
}

type OperatorType int

const (
    OpMultiply OperatorType = iota
    OpDivide
    OpPower
)
```

**How Expression Parsing Works:**

For an expression like `Length/Time^2`, the grammar ensures `^` binds tighter than `/`. The listener exit calls occur in this order:

```
Source: Length / Time ^ 2

Listener Exit call order:
  ExitIdentifier("Length")    → push ParserUnitReference{Name: "Length"} to definitions
  ExitIdentifier("Time")      → push ParserUnitReference{Name: "Time"} to definitions
  ExitPowOperator             → push ParserOperator{Operator: OpPower} to definitions
  ExitInteger(2)              → push ParserUnitReference{Name: "2"} to definitions
  ExitDivOperator             → push ParserOperator{Operator: OpDivide} to definitions

Resulting definitions slice (postfix order):
  [UnitReference("Length"), UnitReference("Time"), Power, UnitReference("2"), Divide]
```

The analysis phase evaluates this postfix sequence using a stack-based algorithm to build the resolved expression tree.

**Grammar Structure for Postfix Parsing:**

```antlr
expression
    : term                              // Single term
    | expression MULT term              // Multiplication
    | expression DIV term               // Division
    ;

term
    : factor                            // Simple factor
    | term POW INTEGER                  // Power
    ;

factor
    : identifier                        // Unit reference
    | LPAREN expression RPAREN          // Parenthesized expression
    ;

// Explicit operator rules for listener callbacks
multOperator: MULT;
divOperator:  DIV;
powOperator:  POW;
```

**Listener Implementation:**

```go
// ExitIdentifier pushes the unit reference onto definitions.
func (l *MeasuresListener) ExitIdentifier(ctx *parser.IdentifierContext) {
    current, _ := l.symbolStack.Peek()
    name := ctx.GetText()
    location := sourceLocationFromContext(ctx)

    symDef := SymbolDefinition{
        Name:           &name,
        DefinitionType: UnitReferenceDefinition,
        Location:       location,
        Data:           &ParserUnitReference{Name: name},
    }
    current.Definitions = append(current.Definitions, symDef)
}

// ExitDivOperator pushes the division operator onto definitions.
func (l *MeasuresListener) ExitDivOperator(ctx *parser.DivOperatorContext) {
    current, _ := l.symbolStack.Peek()
    location := sourceLocationFromContext(ctx)

    symDef := SymbolDefinition{
        DefinitionType: DivideOperatorDefinition,
        Location:       location,
        Data:           &ParserOperator{Operator: OpDivide},
    }
    current.Definitions = append(current.Definitions, symDef)
}
```

**DefinitionType Values for Expression Components:**

```go
const (
    // ... existing definition types ...
    UnitReferenceDefinition    // NEW: A unit name reference in an expression
    MultiplyOperatorDefinition // NEW: Multiplication operator in an expression
    DivideOperatorDefinition   // NEW: Division operator in an expression
    PowerOperatorDefinition    // NEW: Power operator in an expression
)
```

**Key Distinction:**

- Parser-time: Pushes `ParserUnitReference` and `ParserOperator` to definitions slice in postfix order
- Analysis-time: Evaluates postfix sequence to build expression tree with resolved references

See the [Analysis Phase ERO](../../go-analysis/eros/ERO003_units-measures-analysis-phase.md) for details on expression tree construction and unit resolution.

**Enhanced ParserMeasureDefinition (Increment 3)**

```go
// ParserMeasureDefinition stores expression component definitions for compound units.
// The parser pushes unit references and operators to definitions in postfix order.
// The analysis phase evaluates the postfix sequence to build the Expression tree.
type ParserMeasureDefinition struct {
    Name              string             // Measure group identifier
    TypeName          string             // Type name (e.g., "Number")
    SourceLocation    SourceLocation     // For error reporting
    // Expression components are stored in the MeasureScope's Definitions slice,
    // not in this struct. The analysis phase retrieves them by walking the scope.
}

// Note on Increment 2 Compatibility:
// Increment 2 used UnitGroupName (string) for simple unit references.
// Increment 3+ stores expression components as definitions in the scope:
// - Simple unit: Single ParserUnitReference{Name: "TemperatureScale"} in definitions
// - Compound: Multiple ParserUnitReference and ParserOperator entries in postfix order
```

**Transformation Boundary (Parser → Analysis)**

The analysis phase evaluates the postfix definitions sequence to build resolved expression trees. See the [Analysis Phase ERO](../../go-analysis/eros/ERO003_units-measures-analysis-phase.md) for details on the `BuildExpressionTree` function and unit resolution.

**Example CPL Code with Compound Units**

```cpl
concepts
    Length = [Meter, Kilometer, Centimeter]
    Time = [Second, Minute, Hour]

units
    Length = Length
    Time = Time

measures
    Speed = Number(unit=Length/Time)
    Acceleration = Number(unit=Length/Time^2)
    Area = Number(unit=Length^2)
    Momentum = Number(unit=Length*Mass/Time)  // Assuming Mass is defined
```

**Example Definitions Sequence for `Length/Time^2`**

Parser-time (definitions pushed in postfix order):

```
Definitions slice after parsing "Length / Time ^ 2":
  [0] SymbolDefinition{DefinitionType: UnitReferenceDefinition,    Data: &ParserUnitReference{Name: "Length"}}
  [1] SymbolDefinition{DefinitionType: UnitReferenceDefinition,    Data: &ParserUnitReference{Name: "Time"}}
  [2] SymbolDefinition{DefinitionType: PowerOperatorDefinition,    Data: &ParserOperator{Operator: OpPower}}
  [3] SymbolDefinition{DefinitionType: UnitReferenceDefinition,    Data: &ParserUnitReference{Name: "2"}}
  [4] SymbolDefinition{DefinitionType: DivideOperatorDefinition,   Data: &ParserOperator{Operator: OpDivide}}
```

Analysis-time (after postfix evaluation):

```
Expression tree:
DivisionNode
├── left: UnitReference{Name: "Length", UnitGroup: *AnalysisUnitDefinition}
└── right: PowerNode
    ├── base: UnitReference{Name: "Time", UnitGroup: *AnalysisUnitDefinition}
    └── exponent: 2
```

The parser produces definitions in postfix order; the analysis phase evaluates them into a resolved tree with unit definition pointers.

**Validation During Parsing**

- Unit expression syntax validation
- Parentheses matching
- Power exponent must be an integer
- Unit reference syntax validation

**Limitations**

- No unit conversions (cannot simplify compound units)
- No derived unit definitions (e.g., `Force = Mass * Length / Time^2`)
- Unit trees are preserved symbolically without simplification
- No validation that compound units make business sense
- Tensor extension not yet implemented (tree designed for future extension)

### Increment 4: Unit Conversions

**Grammar Changes**

Update [`ChronoconePlanningLanguage.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguage.g4:1) to extend expression rules for conversion calculations:

```antlr
// ChronoconePlanningLanguage.g4 - Root grammar (Increment 4 additions)

grammar ChronoconePlanningLanguage;

import CPLUnits, CPLConcepts;

options {
    tokenVocab = ChronoconePlanningLanguageLexer;
}

// ... existing rules from Increment 3 ...

// ============================================================================
// EXTENDED EXPRESSION RULES (support for +, -, and numeric literals)
// ============================================================================

// Extended expression supports + and - for conversion calculations
expression
    : term
    | expression PLUS term
    | expression MINUS term
    ;

term
    : factor
    | term MULT factor
    | term DIV factor
    ;

factor
    : NUMBER
    | identifier                         // Unit reference in expression
    | identifier LPAREN identifier RPAREN // Unit with value alias: C(c)
    | LPAREN expression RPAREN
    ;
```

Update `CPLUnits.g4` to support conversion clauses (expression rules are imported from main grammar):

```antlr
// CPLUnits.g4 - Grammar for CPL Units section (Increment 4)

parser grammar CPLUnits;

options {
    tokenVocab = ChronoconePlanningLanguageLexer;
}

// ============================================================================
// UNITS SECTION WITH CONVERSIONS
// ============================================================================

unitsSection: UNITS_KEYWORD unitDefinition+;

// Unit definition with optional conversion clauses
unitDefinition
    : identifier EQ identifier conversionClause*
    ;

// Conversion clause contains one or more pattern/calculation pairs
conversionClause: CONVERSION_KEYWORD conversionPair+;

// A conversion pair consists of pattern and calculation
conversionPair: patternClause calculationClause;

// Pattern clause defines the unit relationship
patternClause: PATTERN_KEYWORD pattern;

// Pattern can be simple ratio or equation form
pattern
    : expression LPAREN identifier RPAREN DIV expression LPAREN identifier RPAREN  // Ratio: Unit1(u1) / Unit2(u2)
    | expression LPAREN identifier RPAREN EQ expression LPAREN identifier RPAREN   // Equation: Unit1(u1) = Unit2(u2)
    ;

// Calculation clause defines the conversion formula
calculationClause: CALCULATION_KEYWORD expression;
```

**Lexer Tokens** (add to [`ChronoconePlanningLanguageLexer.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguageLexer.g4:1)):

```antlr
// Keywords for conversions
CONVERSION_KEYWORD : 'conversion';
PATTERN_KEYWORD    : 'pattern';
CALCULATION_KEYWORD: 'calculation';

// Additional operators for conversion calculations
PLUS               : '+';
MINUS              : '-';

// Numeric literals for conversion calculations
NUMBER             : [0-9]+ ('.' [0-9]+)?;
```

**Note on Grammar Organization:**
The `expression`, `term`, and `factor` rules are defined in the top-level [`ChronoconePlanningLanguage.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguage.g4:1) rather than in `CPLUnits.g4` because these arithmetic expression rules will be reused across multiple CPL sections, including:

- Compound unit expressions in measures (Increment 3)
- Unit conversion calculations (Increment 4)
- Measure conversion calculations (Increment 5)
- Metrics formulas (future implementation)

This follows the pattern of defining shared grammar rules in the root grammar file that can be imported by specialized section grammars.

**Parser Types for Expression Components (Increment 4)**

Following the [Postfix Expression Parsing via Symbol Table Stack Pattern](../../../../patterns/postfix-expression-parsing.md), pattern and calculation expressions are parsed into postfix definitions, not captured as raw text.

```go
// Note: ConversionPair, ConversionPattern, and ConversionCalculation are
// analysis-time structures built by walking the symbol table tree.
// They do NOT exist at parse time.
//
// Pattern and calculation expressions use the same postfix approach as compound units:
// - Unit references and value aliases are pushed as ParserUnitReference definitions
// - Operators (*, /, ^, +, -) are pushed as ParserOperator definitions
// - The analysis phase evaluates the postfix sequence to build resolved expression trees
```

**Enhanced ParserUnitDefinition (Increment 4)**

```go
// ParserUnitDefinition does NOT include conversion pairs at parse time.
// Conversions are captured as child symbol table scopes (ConversionScope).
// The analysis phase walks these scopes to build ConversionPair structures.
type ParserUnitDefinition struct {
    Name            string              // Unit group identifier
    ConceptName     string              // Referenced concept name
    SourceLocation  SourceLocation      // For error reporting
}
```

**Symbol Table Tree Structure with Conversions (After Parsing)**

Given this CPL source:

```cpl
units
    Length = Length
        conversion
            pattern
                Kilometer(km) / Meter(m)
            calculation
                1000
            pattern
                Centimeter(cm) / Meter(m)
            calculation
                0.01
```

The resulting symbol table tree is:

```
GlobalScope (root)
└── UnitsScope
    └── UnitScope (for "Length" unit)
        ├── SymbolDefinition: Name="Length"
        │   DefinitionType: UnitScopeDefinition
        │   Data: &ParserUnitDefinition{
        │       Name: "Length",
        │       ConceptName: "Length",
        │   }
        ├── SymbolDefinition: Name="Length"  // Scope marker
        │   DefinitionType: UnitScopeDefinition
        └── ConversionScope (first conversion section)
            ├── SymbolDefinition: DefinitionType=ConversionScopeDefinition
            ├── PatternScope (first pattern)
            │   ├── SymbolDefinition: DefinitionType=PatternScopeDefinition
            │   ├── SymbolDefinition: DefinitionType=UnitReferenceDefinition
            │   │   Data: &ParserUnitReference{Name: "Kilometer"}
            │   ├── SymbolDefinition: DefinitionType=UnitReferenceDefinition
            │   │   Data: &ParserUnitReference{Name: "km"}
            │   ├── SymbolDefinition: DefinitionType=DivideOperatorDefinition
            │   │   Data: &ParserOperator{Operator: OpDivide}
            │   ├── SymbolDefinition: DefinitionType=UnitReferenceDefinition
            │   │   Data: &ParserUnitReference{Name: "Meter"}
            │   └── SymbolDefinition: DefinitionType=UnitReferenceDefinition
            │       Data: &ParserUnitReference{Name: "m"}
            ├── CalculationScope (first calculation)
            │   ├── SymbolDefinition: DefinitionType=CalculationScopeDefinition
            │   └── SymbolDefinition: DefinitionType=UnitReferenceDefinition
            │       Data: &ParserUnitReference{Name: "1000"}
            ├── PatternScope (second pattern)
            │   ├── SymbolDefinition: DefinitionType=UnitReferenceDefinition
            │   │   Data: &ParserUnitReference{Name: "Centimeter"}
            │   ├── SymbolDefinition: DefinitionType=UnitReferenceDefinition
            │   │   Data: &ParserUnitReference{Name: "cm"}
            │   ├── SymbolDefinition: DefinitionType=DivideOperatorDefinition
            │   │   Data: &ParserOperator{Operator: OpDivide}
            │   ├── SymbolDefinition: DefinitionType=UnitReferenceDefinition
            │   │   Data: &ParserUnitReference{Name: "Meter"}
            │   └── SymbolDefinition: DefinitionType=UnitReferenceDefinition
            │       Data: &ParserUnitReference{Name: "m"}
            └── CalculationScope (second calculation)
                └── SymbolDefinition: DefinitionType=UnitReferenceDefinition
                    Data: &ParserUnitReference{Name: "0.01"}
```

**Example CPL Code with Unit Conversions**

```cpl
concepts
    Length = [Meter, Kilometer, Centimeter]
    TemperatureScale = [Values[C, F]]

units
    Length = Length
        conversion
            pattern
                Kilometer(km) / Meter(m)
            calculation
                1000
            pattern
                Centimeter(cm) / Meter(m)
            calculation
                0.01

    TemperatureScale = TemperatureScale
        conversion
            pattern
                C(c) = F(f)
            calculation
                ((5 / 9)*(C/F))*(f-32*F)
```

**Validation During Parsing**

- Conversion clause syntax validation
- Pattern syntax validation
- Calculation expression syntax validation
- Value alias syntax validation
- Note: Pattern/calculation semantic validation deferred to analysis

**Limitations**

- Only unit-level conversions (measure conversions in Increment 5)
- Formula evaluation depends on metrics implementation
- No unit group patterns yet (specific units only)
- Conversion path resolution happens during analysis/synthesis

### Increment 5: Measure Conversions

**Grammar Changes**

Update `CPLMeasures.g4` to support conversion clauses (expression rules are imported from main grammar):

```antlr
// CPLMeasures.g4 - Grammar for CPL Measures section (Increment 5)

parser grammar CPLMeasures;

options {
    tokenVocab = ChronoconePlanningLanguageLexer;
}

// ============================================================================
// MEASURES SECTION WITH CONVERSIONS
// ============================================================================

measuresSection: MEASURES_KEYWORD measureDefinition+;

// Measure definition with optional conversion clauses
measureDefinition
    : identifier EQ identifier LPAREN UNIT EQUALS expression RPAREN measureConversionClause*
    ;

// Measure conversion clause
measureConversionClause: CONVERSION_KEYWORD measureConversionPair+;

// A measure conversion pair consists of pattern and calculation
measureConversionPair: patternClause measureCalculationClause;

// Measure calculation clause
measureCalculationClause: CALCULATION_KEYWORD expression;
```

**Lexer Tokens** (add to [`ChronoconePlanningLanguageLexer.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguageLexer.g4:1)):

```antlr
// Keywords for measure conversions
```

**Note:** The `expression` rule referenced in `measureCalculationClause` is defined in the top-level [`ChronoconePlanningLanguage.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguage.g4:1) and shared across all CPL sections.

**Parser Types for Measure Conversion Components (Increment 5)**

Following the [Postfix Expression Parsing via Symbol Table Stack Pattern](../../../../patterns/postfix-expression-parsing.md), measure conversion pattern and calculation expressions are parsed into postfix definitions, not captured as raw text.

```go
// Note: MeasureConversionPair, MeasureConversionPattern, MeasureConversionCalculation
// are analysis-time structures built by walking the symbol table tree.
// They do NOT exist at parse time.
//
// Pattern and calculation expressions use the same postfix approach as compound units:
// - Measure references are pushed as ParserUnitReference definitions
// - Operators are pushed as ParserOperator definitions
// - The analysis phase evaluates the postfix sequence to build resolved expression trees
```

**ParserMeasureDefinition (Increment 5)**

```go
// ParserMeasureDefinition does NOT include conversion pairs at parse time.
// Conversions are captured as child symbol table scopes (MeasureConversionScope).
// The analysis phase walks these scopes to build MeasureConversionPair structures.
// Expression components (unit references, operators) are stored in the scope's
// Definitions slice in postfix order, following the postfix expression parsing pattern.
type ParserMeasureDefinition struct {
    Name              string                     // Measure group identifier
    TypeName          string                     // Type name
    SourceLocation    SourceLocation             // For error reporting
}

// Note: The Expression tree (with resolved references) is built by the analysis phase.
// The parser pushes ParserUnitReference and ParserOperator definitions in postfix order.
```

**Symbol Table Tree Structure with Measure Conversions (After Parsing)**

Given this CPL source:

```cpl
measures
    Day = Interval(starts_at = Midnight, unit = Duration)
        conversion
      pattern
        Hour=Day(other_day)
      calculation
        when
          other_day = 03-09-2025 => 23
          other_day = 11-02-2025 => 25
          default => 24
        end when
```

The resulting symbol table tree includes:

```
GlobalScope (root)
└── MeasuresScope
    └── MeasureScope (for "Day" measure)
        ├── SymbolDefinition: Name="Day"
        │   DefinitionType: MeasureScopeDefinition
        │   Data: &ParserMeasureDefinition{
        │       Name: "Day",
        │       TypeName: "Interval",
        │   }
        │   Definitions slice (postfix order for "Duration"):
        │     [0] ParserUnitReference{Name: "Duration"}
        └── MeasureConversionScope
            ├── SymbolDefinition: DefinitionType=MeasureConversionScopeDefinition
            ├── PatternScope
            │   ├── SymbolDefinition: DefinitionType=PatternScopeDefinition
            │   ├── SymbolDefinition: DefinitionType=UnitReferenceDefinition
            │   │   Data: &ParserUnitReference{Name: "Hour"}
            │   ├── SymbolDefinition: DefinitionType=UnitReferenceDefinition
            │   │   Data: &ParserUnitReference{Name: "Day"}
            │   └── SymbolDefinition: DefinitionType=UnitReferenceDefinition
            │       Data: &ParserUnitReference{Name: "other_day"}
            └── CalculationScope
                ├── SymbolDefinition: DefinitionType=CalculationScopeDefinition
                └── SymbolDefinition: DefinitionType=UnitReferenceDefinition
                    Data: &ParserUnitReference{Name: "23"}
```

**Example CPL Code with Measure Conversions**

```cpl
concepts
    Duration = [Year, Quarter, Month, Week, Day, Hour, Minute, Second]

units
    Duration = Duration

measures
    Day = Interval(starts_at = Midnight, unit = Duration)
        conversion
      pattern
        Hour=Day(other_day)
      calculation
        when
          other_day = 03-09-2025 => 23
          other_day = 11-02-2025 => 25
          default => 24
        end when
```

**Validation During Parsing**

- Measure conversion clause syntax validation
- When clause syntax validation
- Default clause required for when expressions
- Pattern syntax validation
- Note: Semantic validation deferred to analysis

**Limitations**

- Complex time-related conversions may require additional time-handling features
- Integration with external conversion data sources not included
- Formula evaluation depends on metrics implementation

## Options Considered

### Option 1: Separate Structures per Increment

**Description:**
Create separate ParserUnitDefinition and ParserMeasureDefinition structures for each increment.

**Pros:**

- Clear what's supported in each increment
- Type safety for increment-specific features

**Cons:**

- Duplication and maintenance burden
- Conversion code between versions
- Complex symbol table handling

**Decision:** Rejected. Unnecessary complexity for this use case.

### Option 2: Single Evolving Structure

**Description:**
Use single ParserUnitDefinition and ParserMeasureDefinition structures that grow with each increment.

**Pros:**

- Simpler codebase with fewer structure definitions
- Easy to add new fields as increments progress
- Backward compatible (new fields have zero values)

**Cons:**

- Early increments have unused fields
- Must rely on documentation and validation to enforce increment constraints

**Decision:** Selected. The simplicity of single structures outweighs the benefits of strict versioning.

## Decision

**Chosen Option:** Single evolving structures for both units and measures

**Rationale:**
The simplicity of single structures outweighs the benefits of strict versioning. Documentation and validation rules clearly define which fields are valid in each increment. The Go zero values for unused fields provide a sensible default behavior.

## Implications

### Positive Implications

- Simpler codebase with single structure definitions
- Easy to add new fields as increments progress
- Backward compatible (new fields have zero values)
- Consistent with the ERO001 concepts parsing phase approach

### Negative Implications

- Must rely on documentation and validation to enforce increment constraints
- Unit tests should verify correct behavior per increment
- Unused fields in early increments may confuse developers

### Validation Strategy

Each increment adds validation rules:

1. **Increment 1:** Validate unit group names, concept name references, no duplicate unit groups
2. **Increment 2:** Validate measure group names, type names, unit group references, no duplicate measures
3. **Increment 3:** Validate unit expression syntax, parentheses matching, power exponents
4. **Increment 4:** Validate conversion clause syntax, pattern/calculation pairing, expression syntax
5. **Increment 5:** Validate measure conversion syntax, when clause structure, default clause presence

## Notes

**Related Documents:**

- [Units and Measures Incremental Implementation](../../eros/ERO003_units-measures-incremental-implementation.md) - Main incremental implementation ERO
- [Units and Measures Design](../../units-and-measures.md) - Full units and measures design
- [Units and Measures Analysis Phase](../../go-analysis/eros/ERO003_units-measures-analysis-phase.md) - Analysis phase data structures (transforms Parser → Analysis structures)
- [Symbol Table Design](../../symbol-table-design.md) - Symbol table integration details
- [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md) - Architectural pattern for parsing phase design

**Dependencies on Other Features:**

- **Concepts**: Unit values come from concept values. Concepts must be parsed before units.
- **Types (Number)**: Measures combine types with units. Type parsing must support measures.
- **Metrics/Formula Evaluation**: Conversions require expression evaluation. Formula evaluation from metrics must be implemented before conversions.

**Tree Structure for Future Extension:**

The expression tree (`Expression` interface and implementing types) is designed to be extensible for future tensor support:

```go
// Future extension for metric tensors
type TensorNode struct {
    TensorName string
    Components []Expression
}

func (t TensorNode) isExpression() {}
```

This allows mixed unit-tensor expressions like `MetricTensor * Meter/Second^2` in future increments.
