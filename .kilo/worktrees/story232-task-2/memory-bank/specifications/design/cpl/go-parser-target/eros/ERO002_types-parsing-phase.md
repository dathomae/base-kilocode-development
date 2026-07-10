# ERO002: Types Parsing Phase

## Title

Types Parsing Phase - Data Structures and Incremental Implementation

## Purpose/Goal

Define the parsing phase data structures for type definitions and document how these structures evolve across each increment of the types incremental implementation.

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic001_cpl-concepts-implementation.md](../../../../../epics/Epic001_cpl-concepts-implementation.md) - Epic for CPL concepts implementation

## Context

### Situation

The parsing phase is responsible for converting CPL source text into an Abstract Syntax Tree (AST) and populating the symbol table with type definitions. The parsing phase must support the incremental implementation of type features as defined in the main Types Incremental Implementation ERO.

### Constraints

- Must integrate with existing ANTLR-based parser infrastructure
- Must populate the symbol table for use by subsequent phases
- Must support error reporting with source location information
- Parser data structures must evolve to support each increment
- Type constraints must be parsed and validated for syntax

### Requirements

- Capture raw type definitions from the AST
- Store type names, base types, and constraints
- Support built-in types without declaration
- Perform only syntactic validation during parsing (semantic validation occurs during analysis)

### Assumptions

- The parser uses ANTLR4 with a Go target
- Symbol table entries can be extended incrementally
- Source location information is available for error reporting
- Built-in types (Number, String, Instant, Interval, Boolean) are predefined

## Data Structures by Increment

### Increment 1: Number Type with Type Capture

**Grammar Changes**

In [`ChronoconePlanningLanguage.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguage.g4:1), add `typesSection` to the statement rule:

```antlr
// Import the types grammar
import CPLUnits, CPLConcepts, CPLTypes;

// Statements are the top-level elements of a CPL file
statement: unitsSection | conceptsSection | typesSection;
```

Create new file `CPLTypes.g4`:

```antlr
// CPLTypes.g4 - Grammar for CPL Types section

grammar CPLTypes;

import ChronoconePlanningLanguage; // For identifier rule

options {
    tokenVocab = ChronoconePlanningLanguageLexer;
}

// ============================================================================
// TYPES SECTION
// ============================================================================

// A types section contains one or more type definitions
typesSection: TYPES_KEYWORD typeDefinition+;

// A type definition: TypeName = BaseType(constraints)
typeDefinition: identifier EQ baseType LPAREN constraintList? RPAREN;

// Base type names (built-in types for Increment 1 - Number only)
baseType
    : NUMBER_TYPE
    ;

// Constraint list: comma-separated key=value pairs
constraintList: constraint (COMMA constraint)*;

// Individual constraint: key=value
constraint: identifier EQ constraintValue;

// Constraint values can be integers, strings, or identifiers
constraintValue
    : INTEGER
    | STRING_LITERAL
    | identifier
    ;
```

**Lexer Tokens** (add to [`ChronoconePlanningLanguageLexer.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguageLexer.g4:1)):

```antlr
// Keywords
TYPES_KEYWORD    : 'types';
NUMBER_TYPE      : 'Number';

// Punctuation
LPAREN           : '(';
RPAREN           : ')';
EQ               : '=';
COMMA            : ',';

// Literals
INTEGER          : [0-9]+;
STRING_LITERAL   : '\'' (~['\r\n])* '\'';
```

**Built-in Type Constants**

```go
// BuiltInType represents the Chronocone built-in types
type BuiltInType int

const (
    NumberType BuiltInType = iota
)

func (b BuiltInType) String() string {
    switch b {
    case NumberType:
        return "Number"
    default:
        return "Unknown"
    }
}
```

**Constraint Types**

```go
// Constraint represents a single constraint on a type.
// During parsing, constraints are captured as raw name/value pairs.
// The analysis phase validates which constraints are valid for each base type.
type Constraint struct {
    Name  string
    Value string // Raw value captured during parsing
}

// NumberConstraints holds constraints for Number types.
// This structure is populated during the analysis phase, not parsing.
type NumberConstraints struct {
    Precision *int   // Number of decimal places
    Minimum   *int   // Minimum value (optional)
    Maximum   *int   // Maximum value (optional)
}

// Note: StringConstraints and BooleanConstraints will be added in Increment 2.
// Constraint extraction methods belong in the analysis phase.
```

**ParserTypeDefinition** (Core structure stored in symbol table)

```go
// ParserTypeDefinition represents a parsed type definition
// This structure holds raw data captured during parsing only.
// Constraint extraction and semantic validation are performed during analysis.
type ParserTypeDefinition struct {
    Name           string
    BaseType       BuiltInType
    Constraints    []Constraint
    SourceLocation SourceLocation
}
```

**Note:** Methods for constraint extraction (e.g., `GetNumberConstraints`, `GetStringConstraints`) belong in the analysis phase and are documented in the Types Analysis Phase ERO.

**Scope and Definition Types**

```go
// ScopeType values for type-related scopes
type ScopeType int

const (
    GlobalScope ScopeType = iota
    UnitsScope
    ConceptsScope
    TypesScope           // NEW: Scope for types section
)

// DefinitionType values for type definitions
type DefinitionType int

const (
    UnitDefinition DefinitionType = iota
    ConceptsScopeDefinition
    ConceptScopeDefinition
    TypeDefinition       // NEW: Definition type for type definitions
)
```

**Symbol Table Integration**

- When entering a `types` section:
  - A new symbol table with `ScopeType = TypesScope` is pushed onto the stack
  - A `SymbolDefinition` with `DefinitionType = TypesScope` is added to this table
- When a type definition is encountered:
  - A `SymbolDefinition` with `DefinitionType = TypeDefinition` is added to the TypesScope table
  - The `Name` field points to the type name being defined
  - The `Data` field holds a `*ParserTypeDefinition`
- Built-in types are pre-populated in the global scope before parsing
- Note: Duplicate type detection (within or across sections) is performed during the analysis phase, not during parsing

**Types Listener Methods**

These methods are defined on the unified `CPLListener` struct in `cpl_types_listener.go`, following the pattern established in `cpl_concepts_listener.go`:

```go
// EnterTypesSection is called when entering the types section.
// It pushes a new TypesScope symbol table onto the stack.
func (l *CPLListener) EnterTypesSection(ctx *antlr.TypesSectionContext) {
    l.logEnter("TypesSection", &ctx.BaseParserRuleContext)

    current, err := l.symbolStack.Peek()
    if err != nil {
        l.AddError(fmt.Errorf("symbol stack error entering types section: %w", err))
        return
    }

    typesTable := current.AddChild(symboltable.TypesScope)
    l.symbolStack.Push(typesTable)
}

// ExitTypesSection is called when exiting the types section.
// It pops the types scope symbol table from the stack.
func (l *CPLListener) ExitTypesSection(ctx *antlr.TypesSectionContext) {
    _, err := l.symbolStack.Pop()
    if err != nil {
        l.AddError(fmt.Errorf("symbol stack error exiting types section: %w", err))
    }
    l.logExit("TypesSection", &ctx.BaseParserRuleContext)
}

// EnterTypeDefinition is called when entering a type definition.
// It pushes a new TypeScope onto the symbol stack.
func (l *CPLListener) EnterTypeDefinition(ctx *antlr.TypeDefinitionContext) {
    l.logEnter("TypeDefinition", &ctx.BaseParserRuleContext)

    current, err := l.symbolStack.Peek()
    if err != nil {
        l.AddError(fmt.Errorf("symbol stack error entering type definition: %w", err))
        return
    }

    l.symbolStack.Push(current.AddChild(symboltable.TypeScope))
}

// ExitTypeDefinition is called when exiting a type definition.
// It captures the raw type definition data and pops the type scope.
func (l *CPLListener) ExitTypeDefinition(ctx *antlr.TypeDefinitionContext) {
    typeName := ctx.Identifier().GetText()
    baseTypeStr := ctx.BaseType().GetText()

    // Parse base type (syntactic only - just capture the keyword)
    var baseType BuiltInType
    switch baseTypeStr {
    case "Number":
        baseType = NumberType
    case "String":
        baseType = StringType
    case "Boolean":
        baseType = BooleanType
    case "Instant":
        baseType = InstantType
    case "Interval":
        baseType = IntervalType
    default:
        // Error: Invalid base type keyword (syntactic validation)
        l.AddError(fmt.Errorf("invalid base type: %s", baseTypeStr))
        l.symbolStack.Pop()
        return
    }

    // Parse constraints (raw capture only - no semantic validation)
    var constraints []Constraint
    if ctx.ConstraintList() != nil {
        for _, c := range ctx.ConstraintList().AllConstraint() {
            constraint := l.parseConstraint(c)
            constraints = append(constraints, constraint)
        }
    }

    // Create type definition (raw data only)
    typeDef := &ParserTypeDefinition{
        Name:        typeName,
        BaseType:    baseType,
        Constraints: constraints,
        SourceLocation: SourceLocation{
            Line:   ctx.GetStart().GetLine(),
            Column: ctx.GetStart().GetColumn(),
        },
    }

    // Pop the type scope
    _, err := l.symbolStack.Pop()
    if err != nil {
        l.AddError(fmt.Errorf("symbol stack error exiting type definition: %w", err))
        l.logExit("TypeDefinition", &ctx.BaseParserRuleContext)
        return
    }

    // Add to parent scope
    parent, _ := l.symbolStack.Peek()
    parent.AddDefinition(&SymbolDefinition{
        DefinitionType: TypeDefinition,
        Name:           &typeName,
        Data:           typeDef,
    })

    l.logExit("TypeDefinition", &ctx.BaseParserRuleContext)
}

// parseConstraint converts a constraint context to a Constraint struct.
// This performs only syntactic parsing - no semantic validation.
func (l *CPLListener) parseConstraint(ctx antlr.IConstraintContext) Constraint {
    name := ctx.Identifier().GetText()
    value := ctx.ConstraintValue()

    // Raw capture of constraint name and value
    // Semantic validation (e.g., checking if constraint is valid for base type)
    // is performed during the analysis phase
    return Constraint{Name: name, Value: value.GetText()}
}
```

**Validation During Parsing (Syntactic Only)**

Parsing performs only syntactic validation. Semantic validation (constraint validity, duplicate detection) occurs during the analysis phase.

- Type name syntax validation (valid identifier)
- Base type is recognized keyword: Number (Increment 1 only)
- Constraint syntax validation (key=value format)
- Constraint values are captured as raw text for later analysis

**Limitations**

- No String type support
- No Boolean type support
- No Instant type support
- No Interval type support
- No multiple type sections in the same file
- No type references from measures/units (requires measure implementation)

### Increment 2: String and Boolean Types

**Grammar Changes**

Update `CPLTypes.g4` to support String and Boolean types:

```antlr
// CPLTypes.g4 - Grammar for CPL Types section (Increment 2)

// Base type names (Increment 2 adds String and Boolean)
baseType
    : NUMBER_TYPE
    | STRING_TYPE       // NEW in Increment 2
    | BOOLEAN_TYPE      // NEW in Increment 2
    ;

// Add new lexer tokens
STRING_TYPE      : 'String';
BOOLEAN_TYPE     : 'Boolean';
```

**Built-in Type Constants**

```go
// BuiltInType represents the Chronocone built-in types (Increment 2)
type BuiltInType int

const (
    NumberType BuiltInType = iota
    StringType              // NEW in Increment 2
    BooleanType             // NEW in Increment 2
)

func (b BuiltInType) String() string {
    switch b {
    case NumberType:
        return "Number"
    case StringType:
        return "String"
    case BooleanType:
        return "Boolean"
    default:
        return "Unknown"
    }
}
```

**Constraint Types for String and Boolean**

```go
// StringConstraints holds constraints for String types
type StringConstraints struct {
    Minimum *int    // Minimum length
    Maximum *int    // Maximum length
    Pattern *string // Regex pattern (optional)
}

// BooleanConstraints holds constraints for Boolean types
type BooleanConstraints struct {
    // Boolean has no constraints
}
```

**ParserTypeDefinition (Increment 2)**

```go
// ParserTypeDefinition represents a parsed type definition (Increment 2)
// This structure holds raw data captured during parsing only.
type ParserTypeDefinition struct {
    Name           string
    BaseType       BuiltInType
    Constraints    []Constraint
    SourceLocation SourceLocation
}
```

**Note:** Constraint extraction methods belong in the analysis phase and are documented in the Types Analysis Phase ERO.

**Validation During Parsing (Syntactic Only)**

- Base type is recognized keyword: Number, String, Boolean
- Constraint syntax validation (key=value format)
- Constraint values are captured as raw text for later analysis

**Limitations**

- No Instant type support
- No Interval type support
- No multiple type sections in the same file

### Increment 3: Multiple Type Declaration Sections

**Overview**

Increment 3 allows multiple `types` sections to appear in the same CPL file. When a new `types` section is encountered, it is processed just like in previous increments: a new symbol table with `ScopeType = TypesScope` is pushed onto the stack, and type definitions are added to it.

**Symbol Table Structure with Multiple Sections**

```
GlobalScope (root)
├── TypesScope (first types section)
│   ├── SymbolDefinition: Currency
│   └── SymbolDefinition: ShortString
├── TypesScope (second types section)
│   └── SymbolDefinition: AccountNumber
└── Other scopes...
```

**ParserTypeDefinition**

No changes to the structure. Each types section simply adds to the symbol table tree. The analyzer handles duplicate detection when walking the tree.

**Validation During Parsing (Syntactic Only)**

- Same syntactic validation as previous increments per section
- Note: Cross-section duplicate type detection requires analysis phase visibility and is performed during analysis

**Limitations**

- No Instant type support
- No Interval type support

### Increment 4: Instant Type

**Grammar Changes**

Update `CPLTypes.g4` to support Instant type:

```antlr
// CPLTypes.g4 - Grammar for CPL Types section (Increment 4)

// Base type names (Increment 4 adds Instant)
baseType
    : NUMBER_TYPE
    | STRING_TYPE
    | BOOLEAN_TYPE
    | INSTANT_TYPE      // NEW in Increment 4
    ;

// Add new lexer token
INSTANT_TYPE     : 'Instant';
```

**Built-in Type Constants**

```go
// BuiltInType represents the Chronocone built-in types (Increment 4)
type BuiltInType int

const (
    NumberType BuiltInType = iota
    StringType
    BooleanType
    InstantType             // NEW in Increment 4
)

func (b BuiltInType) String() string {
    switch b {
    case NumberType:
        return "Number"
    case StringType:
        return "String"
    case BooleanType:
        return "Boolean"
    case InstantType:
        return "Instant"
    default:
        return "Unknown"
    }
}
```

**Constraint Types for Instant**

```go
// InstantConstraints holds field-based constraints for Instant types.
// Every supplied field is a constraint on matching UTC moments.
// Omitted fields are valid and mean underspecification.
type InstantConstraints struct {
    Year       *int    // e.g., 2025
    Month      *string // e.g., "January"
    Day        *int    // e.g., 14
    Hour       *int    // e.g., 9
    Minute     *int    // e.g., 30
    Second     *int    // e.g., 0
    DayOfWeek  *string // e.g., "Monday"
    WeekNumber *int    // e.g., 1
}
```

**Validation During Parsing (Syntactic Only)**

- Base type is recognized keyword: Number, String, Boolean, Instant
- Constraint syntax validation (key=value format)
- Constraint values are captured as raw text for later analysis

**Limitations**

- No Interval type support

### Increment 5: Interval Type

**Grammar Changes**

Update `CPLTypes.g4` to support Interval type:

```antlr
// CPLTypes.g4 - Grammar for CPL Types section (Increment 5)

// Base type names (Increment 5 adds Interval)
baseType
    : NUMBER_TYPE
    | STRING_TYPE
    | BOOLEAN_TYPE
    | INSTANT_TYPE
    | INTERVAL_TYPE     // NEW in Increment 5
    ;

// Add new lexer token
INTERVAL_TYPE    : 'Interval';
```

**Built-in Type Constants**

```go
// BuiltInType represents the Chronocone built-in types (Increment 5)
type BuiltInType int

const (
    NumberType BuiltInType = iota
    StringType
    BooleanType
    InstantType
    IntervalType            // NEW in Increment 5
)

func (b BuiltInType) String() string {
    switch b {
    case NumberType:
        return "Number"
    case StringType:
        return "String"
    case BooleanType:
        return "Boolean"
    case InstantType:
        return "Instant"
    case IntervalType:
        return "Interval"
    default:
        return "Unknown"
    }
}
```

**Constraint Types for Interval**

```go
// IntervalConstraints holds constraints for Interval types.
// starts_at is either a named Instant-derived type reference or an inline Instant constraint expression.
// unit references a normal unit group (no special built-in treatment).
type IntervalConstraints struct {
    StartsAt *InstantConstraints // Inline Instant constraints, or nil if using a named type
    StartTypeName *string        // Named Instant-derived type reference, or nil if inline
    UnitName *string             // Name of the unit group
}
```

**Validation During Parsing (Syntactic Only)**

- Base type is recognized keyword: Number, String, Boolean, Instant, Interval
- Constraint syntax validation (key=value format)
- Constraint values are captured as raw text for later analysis

**Note:** Constraint extraction methods (e.g., `GetIntervalConstraints`) belong in the analysis phase and are documented in the Types Analysis Phase ERO.

**Limitations**

None - Full type parsing functionality for all five built-in types.

## Options Considered

### Option 1: Single Evolving Structure

**Description:**
Use a single ParserTypeDefinition structure that grows with each increment.

**Pros:**

- Simple, one structure to maintain
- No versioning complexity

**Cons:**

- Early increments have unused fields
- Harder to reason about what's valid in each increment

**Decision:** Selected. The unused fields in early increments are acceptable for simplicity.

### Option 2: Versioned Structures

**Description:**
Create separate structures for each increment (e.g., ParserTypeDefinitionV1, V2, etc.).

**Pros:**

- Clear what's supported in each increment
- Type safety for increment-specific features

**Cons:**

- Duplication and maintenance burden
- Conversion code between versions

**Decision:** Rejected. Unnecessary complexity for this use case.

## Decision

**Chosen Option:** Single evolving ParserTypeDefinition structure

**Rationale:**
The simplicity of a single structure outweighs the benefits of strict versioning. Documentation and validation rules clearly define which fields are valid in each increment. The Go zero values for unused fields provide a sensible default behavior.

## Implications

### Positive Implications

- Simpler codebase with single structure definition
- Easy to add new fields as increments progress
- Backward compatible (new fields have zero values)

### Negative Implications

- Must rely on documentation and validation to enforce increment constraints
- Unit tests should verify correct behavior per increment

### Validation Strategy

The parsing phase performs only syntactic validation. Semantic validation occurs during the analysis phase.

**Parsing Phase (Syntactic Only):**

1. **Increment 1:** Type name syntax validation, base type keyword validation (Number), constraint syntax (key=value format)
2. **Increment 2:** Add String and Boolean as recognized base type keywords
3. **Increment 3:** No additional parsing validation (multiple sections are syntactically valid)
4. **Increment 4:** Add Instant as recognized base type keyword
5. **Increment 5:** Add Interval as recognized base type keyword

**Analysis Phase (Semantic):**

- Constraint name validity per base type
- Constraint value type compatibility
- Duplicate constraint names within a definition
- Duplicate type definitions (within and across sections)
- Cross-reference validation (e.g., Instant type references for Interval constraints)

## Notes

**Related Documents:**

- [Types Incremental Implementation](../../eros/ERO002_types-incremental-implementation.md) - Main incremental implementation ERO
- [Types Definitions Design](../../types-definitions.md) - Full type design
- [Symbol Table Design](../symbol-table-design.md) - Symbol table integration details
