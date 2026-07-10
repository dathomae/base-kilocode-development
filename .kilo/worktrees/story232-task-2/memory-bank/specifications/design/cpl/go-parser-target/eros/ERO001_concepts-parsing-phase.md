# ERO001: Concepts Parsing Phase

## Title

Concepts Parsing Phase - Data Structures and Incremental Implementation

## Purpose/Goal

Define the parsing phase data structures for concept definitions and document how these structures evolve across each increment of the concepts incremental implementation.

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic001_cpl-concepts-implementation.md](../../../../../epics/Epic001_cpl-concepts-implementation.md) - Epic for CPL concepts implementation

## Context

### Situation

The parsing phase is responsible for converting CPL source text into a **symbol table tree** containing concept definitions. Following the [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md), the parser's primary output is not an Abstract Syntax Tree (AST) but rather a hierarchical symbol table structure that captures parsed definitions. The parsing phase must support the incremental implementation of concept features as defined in the main Concepts Incremental Implementation ERO.

**Key Principle:** The parser builds the symbol table tree; the analysis phase walks this tree to assemble the final AST structures.

### Constraints

- Must integrate with existing ANTLR-based parser infrastructure
- Must populate the symbol table for use by subsequent phases
- Must support error reporting with source location information
- Parser data structures must evolve to support each increment

### Requirements

- Capture raw concept definitions from the AST
- Store concept names, generators, and relationships
- Support indentation-based parent-child relationships
- Enable validation during parsing where possible

### Assumptions

- The parser uses ANTLR4 with a Go target
- Symbol table entries can be extended incrementally
- Source location information is available for error reporting

## Parsing vs Analysis Phase Separation

Following the [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md), concept processing is divided into two distinct phases:

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
3. **Build analysis-level structures** like `AnalysisConceptDefinition`

### Symbol Table Storage Pattern

Concept definitions are stored in the symbol table using the following pattern:

```go
// ParserConceptDefinition is stored in SymbolDefinition.Data
type ParserConceptDefinition struct {
    Name             string
    Generators       []ValuesGeneratorDef
    SourceLocation   SourceLocation
}

// SymbolDefinition holds the concept data in its Data field
type SymbolDefinition struct {
    Name             *string
    DefinitionType   DefinitionType  // ConceptScopeDefinition
    Location         SourceLocation
    Data             interface{}     // Contains *ParserConceptDefinition
}
```

### Symbol Table Structure Example

After parsing a CPL file with concept definitions, the symbol table tree looks like this:

```
GlobalScope (root)
└── ConceptsScope
    ├── SymbolDefinition: Name="Country"
    │   Data: &ParserConceptDefinition{
    │       Name: "Country",
    │       Generators: []ValuesGeneratorDef{{Values: []string{"USA", "Canada", "Mexico"}}},
    │   }
    └── SymbolDefinition: Name="Quarter"
        Data: &ParserConceptDefinition{
            Name: "Quarter",
            Generators: []ValuesGeneratorDef{{Values: []string{"Q1", "Q2", "Q3", "Q4"}}},
        }
```

The analysis phase walks this tree, extracts the `ParserConceptDefinition` structures from `SymbolDefinition.Data`, and builds the final `AnalysisConceptDefinition` AST nodes.

## Data Structures by Increment

### Increment 1: Basic Concept Definitions

**Grammar Changes**

In [`ChronoconePlanningLanguage.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguage.g4:1), add `conceptsSection` to the statement rule:

```antlr
// Import the concepts grammar
import CPLUnits, CPLConcepts;

// Statements are the top-level elements of a CPL file
statement: unitsSection | conceptsSection;
```

Create new file `CPLConcepts.g4`:

```antlr
// CPLConcepts.g4 - Grammar for CPL Concepts section

grammar CPLConcepts;

import ChronoconePlanningLanguage; // For identifier rule

options {
    tokenVocab = ChronoconePlanningLanguageLexer;
}

// ============================================================================
// CONCEPTS SECTION
// ============================================================================

// A concepts section contains one or more concept definitions
conceptsSection: CONCEPTS_KEYWORD conceptDefinition+;

// A concept definition: ConceptName=[Values[Value1, Value2, ...]]
conceptDefinition: identifier EQ LBRACKET generatorPatternList RBRACKET;

// One or more generator patterns
generatorPatternList: generatorPattern (COMMA generatorPattern)*;

// Generator pattern can be Values, Int, or ConceptRange
generatorPattern
    : valuesPatternGenerator
    ;

// Values generator: Values[Value1, Value2, ...]
valuesPatternGenerator: VALUES LBRACKET conceptValueList RBRACKET;

// Comma-separated list of concept values
conceptValueList: identifier (COMMA identifier)*;
```

**Lexer Tokens** (add to [`ChronoconePlanningLanguageLexer.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguageLexer.g4:1)):

```antlr
// Keywords
CONCEPTS_KEYWORD : 'concepts';
VALUES           : 'Values';

// Punctuation
LBRACKET         : '[';
RBRACKET         : ']';
EQ               : '=';
COMMA            : ',';
```

**GeneratorType (Increment 1)**

```go
// GeneratorType identifies the type of concept value generator
type GeneratorType int

const (
    ValuesGenerator GeneratorType = iota  // Values[...] - explicit value list
)
```

**ValuesGeneratorDef (AST Structure Only)**

```go
// ValuesGeneratorDef represents Values[Value1, Value2, ...] in the AST.
// This is a data structure that captures the parsed syntax only.
// Value generation happens during analysis phase.
type ValuesGeneratorDef struct {
    Values []string
}
```

Note: `IntRangeGeneratorDef` and `ConceptRangeGeneratorDef` are added in later increments. The parsing phase only captures AST structures; value generation is performed in the analysis phase.

**Scope and Definition Types**

```go
// ScopeType values for concept-related scopes
type ScopeType int

const (
    GlobalScope ScopeType = iota
    UnitsScope
    ConceptsScope           // NEW: Scope for concepts section
    // ConceptScope is not a ScopeType - see DefinitionType below
)

// DefinitionType values for concept definitions
type DefinitionType int

const (
    UnitDefinition DefinitionType = iota
    ConceptsScopeDefinition       // NEW: Definition type for concepts section scope
    ConceptScopeDefinition        // NEW: Definition type for individual concept scope
)
```

**Concept List Scope**

When a concept definition line is started, a new symbol table is pushed onto the scope stack with `ScopeType = ConceptsScope`. From the listener's viewpoint (via `Peek()`), this top symbol table serves as the "concept list scope":

- A `SymbolDefinition` with `DefinitionType = ConceptScopeDefinition` is added to the symbol table
- The `Data` field of this definition holds a `[]string` slice containing the parsed value terms from the `List[]` declaration
- `Name` field points to the concept name being defined
- Upon completing the concept definition, the values are extracted from `Data` and the scope is popped

**ParserConceptDefinition** (Core structure stored in symbol table)

```go
// ParserConceptDefinition is stored in SymbolDefinition.Data during parsing.
// The analysis phase retrieves this data to build AnalysisConceptDefinition.
type ParserConceptDefinition struct {
    Name            string                   // Concept identifier
    Generators      []ValuesGeneratorDef     // One or more generators producing values
    SourceLocation  SourceLocation           // For error reporting
}
```

Note: In Increment 1, only `ValuesGeneratorDef` is supported in the Generators slice. Other generator types are added in later increments as the slice type evolves to an interface{} or sum type to accommodate multiple generator definitions.

**Storage in Symbol Table**

The `ParserConceptDefinition` is stored in the symbol table as follows:

```go
// In the listener, when a concept definition is completed:
func (l *ConceptsListener) ExitConceptDefinition(ctx *parser.ConceptDefinitionContext) {
    // Build the ParserConceptDefinition from parsed data
    conceptDef := &ParserConceptDefinition{
        Name:           conceptName,
        Generators:     generators,
        SourceLocation: sourceLocation,
    }

    // Get the current scope table (ConceptsScope)
    current, _ := l.symbolStack.Peek()

    // Store in SymbolDefinition.Data
    current.Define(&conceptName, ConceptScopeDefinition, sourceLocation)

    // The SymbolDefinition is the last entry in current.Definitions
    lastIdx := len(current.Definitions) - 1
    current.Definitions[lastIdx].Data = conceptDef  // Store ParserConceptDefinition here
}
```

**Note on Increment 1 Scope:**

- Parent-child relationships are not tracked in Increment 1 (added in Increment 2)
- Indentation level is not stored (added in Increment 2)
- Values are parsed and stored for use by the analyzer when building `AnalysisConceptDefinition`

````

**Symbol Table Integration**

- When entering a `concepts` section:
  - A new symbol table with `ScopeType = ConceptsScope` is pushed onto the stack
  - A `SymbolDefinition` with `DefinitionType = ConceptsScopeDefinition` is added to this table
- When a concept definition line is started:
  - A new symbol table with `ScopeType = ConceptsScope` is pushed as the concept list scope
  - A `SymbolDefinition` with `DefinitionType = ConceptScopeDefinition` is added
  - The `Name` field points to the concept name
  - The `Data` field holds a `[]string` slice for accumulating value terms
- As value terms are parsed, they are appended to the `Data` slice
- Upon completing the concept definition:
  - Values are extracted from the `Data` field
  - A `ParserConceptDefinition` with a `ValuesGenerator` is created and stored in the parent scope's symbol table
  - The concept list scope is popped
- Each concept has a unique name within the compilation unit
- Duplicate concept definitions in the same file are errors

**Example Symbol Table Structure (After Parsing)**

Given this CPL source:

```cpl
concepts
    Country=[Values[USA, Canada, Mexico]]
    Quarter=[Values[Q1, Q2, Q3, Q4]]
```

The resulting symbol table tree is:

```
GlobalScope (root)
└── ConceptsScope
    ├── SymbolDefinition
    │   Name: "Country"
    │   DefinitionType: ConceptScopeDefinition
    │   Data: &ParserConceptDefinition{
    │       Name: "Country",
    │       Generators: []ValuesGeneratorDef{
    │           {Values: []string{"USA", "Canada", "Mexico"}},
    │       },
    │   }
    └── SymbolDefinition
        Name: "Quarter"
        DefinitionType: ConceptScopeDefinition
        Data: &ParserConceptDefinition{
            Name: "Quarter",
            Generators: []ValuesGeneratorDef{
                {Values: []string{"Q1", "Q2", "Q3", "Q4"}},
            },
        }
```

The analysis phase walks this tree and converts each `ParserConceptDefinition` into an `AnalysisConceptDefinition` AST node.

**Validation During Parsing**

- Concept name syntax validation
- Values generator value parsing (comma-separated identifiers)
- Duplicate value detection within a single definition

**Limitations**

- No parent-child relationship tracking
- No multiple concept declaration sections in the same file
- No Int or Concept Range generator support
- No conditional specifications
- No before/after clauses for value ordering

### Increment 2: Concept Hierarchy via Indentation Handling

**Overview**

Increment 2 introduces parent-child relationships between concepts using indentation-based scoping. Unlike Increment 1 where all concepts were at the same level, Increment 2 allows concepts to be nested:

```cpl
concepts
    Country=[Values[USA, Canada, Mexico]]
      Region=[Values[North, South]]  # Child of Country (indented)
      City=[Values[NYC, LA]]         # Child of Country (indented)
        Neighborhood=[Values[...]]   # Grandchild of Country (further indented)
```

**CPLLexerBase: Indentation-Aware Lexer**

To handle indentation-based scoping, we use a lexer base struct similar to standard indentation-aware lexers. The lexer emits `INDENT` and `DEDENT` tokens that the parser uses to track nesting levels.

```go
// CPLLexerBase provides indentation handling for CPL source files.
// It is embedded in the generated ChronoconePlanningLanguageLexer.
type CPLLexerBase struct {
    *antlr.BaseLexer

    // Indentation tracking
    indentStack   []int      // Stack of indentation levels
    pendingTokens []antlr.Token // Buffered tokens (for emitting multiple DEDENTs)
    currentIndent int        // Current line's indentation level
    atLineStart   bool       // True at the beginning of a line
}

// NewCPLLexerBase creates a new lexer base with initialized indentation stack.
func NewCPLLexerBase(input antlr.CharStream) *CPLLexerBase {
    return &CPLLexerBase{
        BaseLexer:     antlr.NewBaseLexer(input),
        indentStack:   []int{0}, // Start with 0 indentation
        pendingTokens: make([]antlr.Token, 0),
        currentIndent: 0,
        atLineStart:   true,
    }
}

// ProcessIndentation is called by the lexer when it encounters whitespace at the start of a line.
// It calculates the indentation level and emits INDENT/DEDENT tokens as needed.
func (l *CPLLexerBase) ProcessIndentation(spaces int) {
    if !l.atLineStart {
        return // Not at line start, treat as regular whitespace
    }

    l.atLineStart = false
    l.currentIndent = spaces

    top := l.indentStack[len(l.indentStack)-1]

    if spaces > top {
        // Increased indentation: emit INDENT
        l.indentStack = append(l.indentStack, spaces)
        l.pendingTokens = append(l.pendingTokens, l.createIndentToken())
    } else if spaces < top {
        // Decreased indentation: emit DEDENTs
        for len(l.indentStack) > 0 && l.indentStack[len(l.indentStack)-1] > spaces {
            l.indentStack = l.indentStack[:len(l.indentStack)-1]
            l.pendingTokens = append(l.pendingTokens, l.createDedentToken())
        }

        // Validate the dedent matches a previous level
        if len(l.indentStack) == 0 || l.indentStack[len(l.indentStack)-1] != spaces {
            // Error: dedent doesn't match any previous level
            // This will be handled as a lexer error
        }
    }
}

// NextToken overrides the base lexer to emit pending INDENT/DEDENT tokens.
func (l *CPLLexerBase) NextToken() antlr.Token {
    // Emit any pending tokens first
    if len(l.pendingTokens) > 0 {
        token := l.pendingTokens[0]
        l.pendingTokens = l.pendingTokens[1:]
        return token
    }

    token := l.BaseLexer.NextToken()

    // Track line starts for indentation detection
    if token.GetTokenType() == antlr.TokenEOF || token.GetChannel() == antlr.TokenDefaultChannel {
        if token.GetTokenType() == antlr.TokenEOF {
            // Emit DEDENTs for all remaining indentation levels at EOF
            for len(l.indentStack) > 1 {
                l.indentStack = l.indentStack[:len(l.indentStack)-1]
                l.pendingTokens = append(l.pendingTokens, l.createDedentToken())
            }
            if len(l.pendingTokens) > 0 {
                l.pendingTokens = append(l.pendingTokens, token)
                return l.pendingTokens[0]
            }
        }

        // Check if this token ends the line
        text := token.GetText()
        if strings.Contains(text, "\n") || strings.Contains(text, "\r") {
            l.atLineStart = true
            l.currentIndent = 0
        }
    }

    return token
}

// createIndentToken creates an INDENT token.
func (l *CPLLexerBase) createIndentToken() antlr.Token {
    // Implementation creates a custom token with type INDENT
    return &antlr.CommonToken{
        TokenType: ChronoconePlanningLanguageLexerINDENT,
        Channel:   antlr.TokenDefaultChannel,
    }
}

// createDedentToken creates a DEDENT token.
func (l *CPLLexerBase) createDedentToken() antlr.Token {
    // Implementation creates a custom token with type DEDENT
    return &antlr.CommonToken{
        TokenType: ChronoconePlanningLanguageLexerDEDENT,
        Channel:   antlr.TokenDefaultChannel,
    }
}

// GetIndentationLevel returns the current indentation level (0 = root).
func (l *CPLLexerBase) GetIndentationLevel() int {
    return len(l.indentStack) - 1
}
```

**Enhanced SymbolTable with Indentation Level**

As noted in the symbol table design, the indent count is stored as an attribute on the `SymbolTable` struct:

```go
// SymbolTable captures a set of names and their corresponding definitions.
// It maintains parent-child relationships to form a hierarchical tree structure.
type SymbolTable struct {
    // ScopeType indicates the kind of scope this table defines.
    ScopeType ScopeType
    // Parent is a pointer to the parent symbol table (nil for root/global).
    Parent *SymbolTable
    // Children is a list of pointers to child symbol tables.
    Children []*SymbolTable
    // Definitions is a list of symbol definitions in this scope.
    Definitions []SymbolDefinition
    // IndentationLevel is the nesting level of this scope (0 = root).
    // For ConceptsScope, this indicates the concept hierarchy depth.
    IndentationLevel int  // NEW in Increment 2
}
```

**Grammar Changes for Indentation Tokens**

Update [`ChronoconePlanningLanguageLexer.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguageLexer.g4:1) to add INDENT/DEDENT tokens and modify whitespace handling:

```antlr
lexer grammar ChronoconePlanningLanguageLexer;

options {
    superClass = CPLLexerBase;  // Use our custom lexer base
}

// ... existing tokens ...

// NEW: Indentation tokens (generated by CPLLexerBase, not matched directly)
tokens {
    INDENT,    // Increased indentation level
    DEDENT     // Decreased indentation level
}

// Modified WS rule to track leading whitespace for indentation
// The lexer base's ProcessIndentation method handles INDENT/DEDENT generation
LEADING_WS: [ \t]+ {l.atLineStart}? -> channel(HIDDEN);

// Regular whitespace (not at line start)
WS: [ \t\r\n]+ -> skip;
```

Update `CPLConcepts.g4` to use INDENT/DEDENT tokens for hierarchy:

```antlr
// CPLConcepts.g4 - Grammar for CPL Concepts section (Increment 2)
// Following the pattern for indent/dedent handling in CPL

parser grammar CPLConcepts;

options {
    tokenVocab = ChronoconePlanningLanguageLexer;
    superClass = CPLParserBase;  // Optional: custom parser base if needed
}

// Concepts section: 'concepts' keyword followed by indented concept definitions
// The INDENT token is emitted by the lexer when it detects increased indentation
// The DEDENT token is emitted when indentation decreases
conceptsSection
    : CONCEPTS_KEYWORD
      INDENT
      conceptDefinition+
      DEDENT
    ;

// A concept definition consists of its declaration line followed by
// optional child concepts at a deeper indentation level
conceptDefinition
    : conceptLine
      (INDENT conceptDefinition+ DEDENT)?  // Recursive: children are also concept definitions
    ;

// Base concept line: ConceptName=[Values[Value1, Value2, ...]]
conceptLine
    : identifier EQ LBRACKET generatorPatternList RBRACKET
    ;

// One or more generator patterns
generatorPatternList
    : generatorPattern (COMMA generatorPattern)*
    ;

// Generator pattern (Values only in Increment 2)
generatorPattern
    : valuesPatternGenerator
    ;

// Values generator: Values[Value1, Value2, ...]
valuesPatternGenerator
    : VALUES LBRACKET conceptValueList RBRACKET
    ;

// Comma-separated list of concept values
conceptValueList
    : identifier (COMMA identifier)*
    ;
```

**Key Points from Indentation-Aware Lexer Pattern:**

1. **INDENT/DEDENT are parser tokens**: The lexer emits these as regular tokens that the parser consumes just like any other token (e.g., `CONCEPTS_KEYWORD`)

2. **INDENT starts a block**: After the `concepts` keyword, an `INDENT` token signals the start of the concept definition block

3. **DEDENT ends a block**: When the lexer detects decreased indentation, it emits `DEDENT` to close the block

4. **Recursive structure**: Child concepts use the same `conceptDefinition` rule recursively, with their own INDENT/DEDENT pairs

5. **Multiple DEDENTs**: If indentation decreases by multiple levels (e.g., exiting a grandchild back to root), the lexer emits multiple DEDENT tokens

**ParserConceptDefinition**

```go
type ParserConceptDefinition struct {
    Name             string
    Generators       []Generator  // One or more generators (Values only in Increment 2)
    SourceLocation   SourceLocation
    // Note: IndentationLevel is stored in the SymbolTable, not here
}
```

**Concept Listener with Indentation Tracking**

The listener uses the indentation level from the symbol table to track concept hierarchy:

```go
// ConceptsListener handles concept section parsing with indentation awareness.
type ConceptsListener struct {
    *parser.BaseChronoconePlanningLanguageListener
    symbolTable    *symboltable.SymbolTableStack
    currentTable   *symboltable.SymbolTable  // Current concept scope table
}

// EnterConceptsSection is called when entering the concepts section.
func (l *ConceptsListener) EnterConceptsSection(ctx *parser.ConceptsSectionContext) {
    // Push a new ConceptsScope symbol table
    // IndentationLevel is 0 at the section level
    conceptsTable := l.symbolTable.Push(symboltable.ConceptsScope)
    conceptsTable.IndentationLevel = 0
    l.currentTable = conceptsTable
}

// EnterConceptLine is called when entering a concept definition line.
func (l *ConceptsListener) EnterConceptLine(ctx *parser.ConceptLineContext) {
    conceptName := ctx.Identifier().GetText()

    // Get current indentation level from the symbol table stack
    currentLevel := l.currentTable.IndentationLevel

    // If this concept has children (next line has INDENT),
    // we push a new scope for the child concepts
    // The actual child handling happens in EnterChildConceptDefinition
}

// EnterChildConceptDefinition handles child concepts (after INDENT token).
func (l *ConceptsListener) EnterChildConceptDefinition(ctx *parser.ChildConceptDefinitionContext) {
    // Push a new scope for child concepts
    childTable := l.symbolTable.Push(symboltable.ConceptsScope)
    childTable.IndentationLevel = l.currentTable.IndentationLevel + 1
    l.currentTable = childTable
}

// ExitChildConceptDefinition is called when exiting a child concept block (after DEDENT).
func (l *ConceptsListener) ExitChildConceptDefinition(ctx *parser.ChildConceptDefinitionContext) {
    // Pop back to parent scope
    l.symbolTable.Pop()
    l.currentTable = l.symbolTable.Peek()
}

// Note: Building the concept hierarchy from indentation levels is done during analysis,
// not during parsing. The parser's job is to capture the indentation level in the
// SymbolTable. The analyzer's AnalyzeConceptsScope function (registered with WalkSymbolTable)
// builds the parent-child relationships by examining the IndentationLevel of each
// ConceptsScope table encountered during the walk.
```

**Validation During Parsing**

- Indentation consistency (must be consistent multiple of base indent, typically 2 or 4 spaces)
- Dedent must match a previous indentation level (no arbitrary dedents)
- No mixing of tabs and spaces in indentation

**Limitations**

- No multiple concept declaration sections in the same file
- No conditional specifications
- Child values are global (not per-parent)
- No Int or Concept Range generators
- No before/after clauses for value ordering

### Increment 3: Multiple Concept Declaration Sections in the Same File

**Overview**

Increment 3 allows multiple `concepts` sections to appear in the same CPL file. When a new `concepts` section is encountered, it is processed just like in Increments 1 and 2: a new symbol table with `ScopeType = ConceptsScope` is pushed onto the stack, and concept definitions are added to it. There is no separate merging step during parsing.

**Symbol Table Structure with Multiple Sections**

```
GlobalScope (root)
├── ConceptsScope (first concepts section)
│   ├── SymbolDefinition: ConceptA
│   └── SymbolDefinition: ConceptB
├── ConceptsScope (second concepts section)
│   ├── SymbolDefinition: ConceptC
│   └── SymbolDefinition: ConceptA (duplicate - error detected during analysis)
└── Other scopes...
```

**ParserConceptDefinition**

```go
type ParserConceptDefinition struct {
    Name                string
    Generators          []Generator  // One or more generators (Values only in Increment 3)
    IndentationLevel    int
    SourceLocation      SourceLocation
}
```

Note: The `IsExtension` and `SectionIndex` fields are not needed in Increment 3. Each concepts section is simply added to the symbol table tree. The analyzer handles duplicate detection when walking the tree.

**Symbol Table Integration**

- Each `concepts` section in the file creates a new `ConceptsScope` symbol table
- These tables are siblings under the GlobalScope (or appropriate parent scope)
- Concept definitions are added to their respective ConceptsScope tables
- No special merging logic during parsing - the symbol table naturally accumulates all definitions
- Duplicate concept detection happens during the analysis phase when walking the symbol table tree

**Validation During Parsing**

- Same as previous increments per section
- Cross-section validation (duplicate concept detection) happens during analysis phase

**Limitations**

- No Int or Concept Range generators
- No conditional specifications
- No before/after clauses for value ordering

### Increment 4: Conditional Specifications

**Grammar Changes**

Add new tokens to [`ChronoconePlanningLanguageLexer.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguageLexer.g4:1):

```antlr
// NEW: Conditional specification tokens
ARROW            : '=>';
LBRACKET         : '[';  // Already defined, used for parent value
RBRACKET         : ']';  // Already defined
```

Update `CPLConcepts.g4` to support conditional value specifications:

```antlr
// CPLConcepts.g4 - Grammar for CPL Concepts section (Increment 4)

// A concept definition can be a simple list or a conditional specification
conceptDefinition
    : conceptLine                          // Simple values generator
    | conditionalSpec                      // Conditional specification
    ;

// Base concept line: ConceptName=[Values[Value1, Value2, ...]]
conceptLine
    : identifier EQ LBRACKET generatorPatternList RBRACKET
    ;

// Generator pattern list (Values only in Increment 4)
generatorPatternList
    : generatorPattern (COMMA generatorPattern)*
    ;

// Generator pattern (Values only in Increment 4)
generatorPattern
    : valuesPatternGenerator
    ;

// Values generator: Values[Value1, Value2, ...]
valuesPatternGenerator
    : VALUES LBRACKET conceptValueList RBRACKET
    ;

// Conditional specification: Parent[Value] => Child=[Values[...]]
conditionalSpec
    : identifier LBRACKET identifier RBRACKET ARROW conceptLine
    ;
```

**Conditional Value Specification**

```go
type ParserConceptDefinition struct {
    Name                string
    Generators          []Generator        // Base generators (when no conditions match)
    IndentationLevel    int
    ConditionalSpecs    []ConditionalSpec  // NEW
    SourceLocation      SourceLocation
}

type ConditionalSpec struct {
    ParentValue      string      // Value of parent concept
    Generators       []Generator // Generators for this condition
}
```

Note: `IsExtension` is not tracked by the parser. Whether a concept extends an existing definition is determined during analysis when duplicate concept names are detected across sections.

**Validation During Parsing**

- Parent value in condition must be a valid identifier
- At most one default specification (no parent value qualifier)
- Conditional specs only valid for child concepts

**Limitations**

- No Int or Concept Range generators
- No before/after clauses for value ordering
- Parent value references not validated against actual parent values during parsing
- Full validation deferred to analysis phase

### Increment 5: Int Range Generators

**Overview**

Increment 5 adds Int Range generators as an alternative to Values generators for defining numeric concept values. This is useful for concepts like months (1-12), fiscal periods, years, etc.

- **IntRangeGenerator**: Defines numeric ranges like `Int[1, 31, 1]` for days of month

**Grammar Changes**

Add new tokens to [`ChronoconePlanningLanguageLexer.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguageLexer.g4:1):

```antlr
// NEW: Int Range generator tokens
INT              : 'Int';
INTEGER          : [0-9]+;
```

Update `CPLConcepts.g4` to support Int Range generators:

```antlr
// CPLConcepts.g4 - Grammar for CPL Concepts section (Increment 5)

// A concept definition can use Values or Int generators
conceptDefinition
    : conceptLine                          // Values or Int generator(s)
    | conditionalSpec                      // Conditional specification
    ;

// Base concept line: ConceptName=[generator patterns]
conceptLine
    : identifier EQ LBRACKET generatorPatternList RBRACKET
    ;

// Generator patterns: Values or Int
generatorPatternList
    : generatorPattern (COMMA generatorPattern)*
    ;

generatorPattern
    : valuesPatternGenerator
    | intPatternGenerator
    ;

// Values generator: Values[Value1, Value2, ...]
valuesPatternGenerator
    : VALUES LBRACKET conceptValueList RBRACKET
    ;

// Int range generator: Int[start, end, step]
intPatternGenerator
    : INT LBRACKET intRange RBRACKET
    ;

// Comma-separated list of concept values (for Values generator)
conceptValueList
    : identifier (COMMA identifier)*
    ;

// Integer range: start, end, step(optional)
intRange
    : INTEGER COMMA INTEGER (COMMA INTEGER)?
    ;
```

**Generator Type Discriminator (Increment 5)**

```go
// GeneratorType identifies the type of concept value generator
type GeneratorType int

const (
    ValuesGenerator GeneratorType = iota
    IntRangeGenerator  // NEW in Increment 5
)
```

**IntRangeGeneratorDef (AST Structure Only)**

```go
// IntRangeGeneratorDef represents Int[start, end, step] in the AST.
// This is a data structure that captures the parsed syntax only.
// Value generation happens during analysis phase.
type IntRangeGeneratorDef struct {
    Start int
    End   int
    Step  int  // Defaults to 1 if not specified
}
```

**Parser Generator Types (Increment 5)**

In Increment 5, the parser needs to handle both Values and Int Range generators. The Generators field becomes a slice of `interface{}` (or a sum type) to accommodate both generator definition types:

```go
type ParserConceptDefinition struct {
    Name                string
    Generators          []interface{}    // Can contain ValuesGeneratorDef or IntRangeGeneratorDef
    IndentationLevel    int
    ConditionalSpecs    []ConditionalSpec
    SourceLocation      SourceLocation
}
```

Note: In Go, this could also be implemented as a discriminated union using a struct with a Type field and pointer fields for each generator type. The key point is that the parser only captures the AST structure; no value generation occurs at parse time.

**ParserConceptDefinition**

```go
type ParserConceptDefinition struct {
    Name                string
    Generators          []Generator  // Can include Values and IntRange generators
    IndentationLevel    int
    ConditionalSpecs    []ConditionalSpec
    SourceLocation      SourceLocation
}
```

**Example CPL Code**

```cpl
concepts
    // Values generator
    Country=[Values[USA, Canada, Mexico]]

    // Int range generator for day numbers
    DayNumbers=[Int[1, 31, 1]]

    // Multiple generators combined
    MixedValues=[Values[A, B], Int[1, 3, 1]]  // Produces: A, B, 1, 2, 3
```

**Validation During Parsing**

- Int range generator: start <= end, step must be positive integer

**Value Ordering Note**

Value order is maintained from the declaration:
- Values generators: values appear in declaration order
- Int range generators: values ordered from start to end by step
- Multiple generators: values appear in the order generators are declared

**Limitations**

- No Concept Range generators (added in Increment 6)
- No before/after clauses for value ordering when extending concepts (added in Increment 7)
- No integration with external Concepts API
- Concepts are defined only within CPL files

### Increment 6: Concept Range Generators

**Overview**

Increment 6 adds Concept Range generators as an alternative to Values and Int Range generators. This allows referencing a range of values from another concept.

- **ConceptRangeGenerator**: References values from another concept like `Quarter[Q1..Q4]`

**Grammar Changes**

Add new tokens to [`ChronoconePlanningLanguageLexer.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguageLexer.g4:1):

```antlr
// NEW: Concept Range generator tokens
DOTDOT           : '..';
```

Update `CPLConcepts.g4` to support Concept Range generators:

```antlr
// CPLConcepts.g4 - Grammar for CPL Concepts section (Increment 6)

// A concept definition can use Values, Int, or Concept Range generators
conceptDefinition
    : conceptLine                          // Values, Int, or Concept Range generator(s)
    | conditionalSpec                      // Conditional specification
    ;

// Base concept line: ConceptName=[generator patterns]
conceptLine
    : identifier EQ LBRACKET generatorPatternList RBRACKET
    ;

// Generator patterns: Values, Int, or Concept Range
generatorPatternList
    : generatorPattern (COMMA generatorPattern)*
    ;

generatorPattern
    : valuesPatternGenerator
    | intPatternGenerator
    | conceptPatternGenerator
    ;

// Values generator: Values[Value1, Value2, ...]
valuesPatternGenerator
    : VALUES LBRACKET conceptValueList RBRACKET
    ;

// Int range generator: Int[start, end, step]
intPatternGenerator
    : INT LBRACKET intRange RBRACKET
    ;

// Concept range generator: Concept[StartValue..EndValue]
conceptPatternGenerator
    : identifier LBRACKET conceptRange RBRACKET
    ;

// Comma-separated list of concept values (for Values generator)
conceptValueList
    : identifier (COMMA identifier)*
    ;

// Integer range: start, end, step(optional)
intRange
    : INTEGER COMMA INTEGER (COMMA INTEGER)?
    ;

// Concept range: StartValue..EndValue
conceptRange
    : identifier DOTDOT identifier
    ;
```

**Generator Type Discriminator (Increment 6)**

```go
// GeneratorType identifies the type of concept value generator
type GeneratorType int

const (
    ValuesGenerator GeneratorType = iota
    IntRangeGenerator
    ConceptRangeGenerator  // NEW in Increment 6
)
```

**ConceptRangeGeneratorDef (AST Structure Only)**

```go
// ConceptRangeGeneratorDef represents Concept[StartValue..EndValue] in the AST.
// This is a data structure that captures the parsed syntax only.
// Value resolution happens during analysis phase after all concepts are defined.
type ConceptRangeGeneratorDef struct {
    ConceptName string
    StartValue  string
    EndValue    string
}
```

**ParserConceptDefinition (Increment 6)**

```go
type ParserConceptDefinition struct {
    Name                string
    Generators          []interface{}    // Can contain ValuesGeneratorDef, IntRangeGeneratorDef, or ConceptRangeGeneratorDef
    IndentationLevel    int
    ConditionalSpecs    []ConditionalSpec
    SourceLocation      SourceLocation
}
```

**Note on Concept Range Resolution:**

Concept range generators reference values from other concepts. During parsing, only the AST structure (concept name and start/end value strings) is captured. The actual value resolution - looking up the referenced concept, finding the start and end value positions, and extracting the range of values - happens during the analysis phase. This is necessary because:

1. The referenced concept must be fully defined before its values can be referenced
2. The order of concept definitions in the source file may not match dependency order
3. Circular references must be detected and reported as errors during analysis

**Example CPL Code**

```cpl
concepts
    // Values generator
    Quarter=[Values[Q1, Q2, Q3, Q4]]

    // Int range generator
    DayNumbers=[Int[1, 31, 1]]

    // Concept range generator
    QuarterRange=[Quarter[Q1..Q4]]

    // Multiple generators combined
    MixedValues=[Values[A, B], Int[1, 3, 1], Quarter[Q1..Q2]]
```

**Validation During Parsing**

- Concept range generator: start and end must be valid identifiers
- Full validation of concept range references deferred to analysis phase

**Value Ordering Note**

Value order is maintained from the declaration:
- Values generators: values appear in declaration order
- Int range generators: values ordered from start to end by step
- Concept range generators: values ordered from start to end based on referenced concept's ordering
- Multiple generators: values appear in the order generators are declared

**Limitations**

- No before/after clauses for value ordering when extending concepts (added in Increment 7)
- No integration with external Concepts API
- Concepts are defined only within CPL files

### Increment 7: Before/After Clauses for Value Ordering

**Overview**

Increment 7 adds before/after clauses for inserting values when extending concepts. This allows new values to be inserted at specific positions rather than only appending.

**Grammar Changes**

Add new tokens to [`ChronoconePlanningLanguageLexer.g4`](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguageLexer.g4:1):

```antlr
// NEW: Before/after clause tokens
BEFORE           : 'before';
AFTER            : 'after';
```

Update `CPLConcepts.g4` to support before/after clauses in concept extensions:

```antlr
// CPLConcepts.g4 - Grammar for CPL Concepts section (Increment 7)

// A concept definition in an extension section can include before/after
conceptDefinition
    : conceptLine                          // Simple values generator(s)
    | conditionalSpec                      // Conditional specification
    | conceptExtension                     // Extension with before/after
    ;

// Concept extension with before/after: Concept=[Values[NewValues]] before|after ExistingValue
conceptExtension
    : identifier EQ LBRACKET generatorPatternList RBRACKET insertionPoint
    ;

// Generator patterns (can include Values, Int, or Concept Range)
generatorPatternList
    : generatorPattern (COMMA generatorPattern)*
    ;

generatorPattern
    : valuesPatternGenerator
    | intPatternGenerator
    | conceptPatternGenerator
    ;

// Values generator: Values[Value1, Value2, ...]
valuesPatternGenerator
    : VALUES LBRACKET conceptValueList RBRACKET
    ;

// Int generator: Int[start, end, step]
intPatternGenerator
    : INT LBRACKET intRange RBRACKET
    ;

// Concept range generator: Concept[StartValue..EndValue]
conceptPatternGenerator
    : identifier LBRACKET conceptRange RBRACKET
    ;

// Insertion point specifies where to insert new values
insertionPoint
    : BEFORE identifier                    // Insert before this existing value
    | AFTER identifier                     // Insert after this existing value
    | BEFORE identifier AFTER identifier   // Both before and after (for positioning)
    ;
```

**Important Notes:**

- Before/after clauses only apply when adding new values to an existing concept
- At least one of `before` or `after` must be specified (both may be specified, but not neither)
- The initial declaration establishes the relative order of all values in that declaration
- Before/after clauses cannot alter the relative order of original values from the initial declaration
- New values inserted between original values change the absolute positions but preserve the relative order of the original values

**Enhanced ParserConceptDefinition (Increment 7)**

```go
type ParserConceptDefinition struct {
    Name                string
    Generators          []interface{}    // ValuesGeneratorDef, IntRangeGeneratorDef, or ConceptRangeGeneratorDef
    IndentationLevel    int
    InsertionPoint      InsertionPoint   // NEW in Increment 7: For before/after
    ConditionalSpecs    []ConditionalSpec
    SourceLocation      SourceLocation
}

type InsertionPoint struct {
    BeforeValue string  // Insert before this value (if set)
    AfterValue  string  // Insert after this value (if set)
}
```

**Example:**

```cpl
concepts
    Priority=[Values[Low, Medium, High]]

concepts
    Priority=[Values[Critical]] before Low
    Priority=[Values[Urgent]] after Medium
```

**Validation During Parsing**

- Before/after values must refer to existing values (if specified)
- Cannot specify both before and after in same definition
- Cross-section validation for before/after references

**Limitations**

- Cannot reorder existing values - before/after only applies to new values being added
- No integration with external Concepts API
- Concepts are defined only within CPL files

**Full ParserConceptDefinition (Through Increment 7)**

```go
type ParserConceptDefinition struct {
    Name                string
    Generators          []interface{}      // ValuesGeneratorDef, IntRangeGeneratorDef, or ConceptRangeGeneratorDef
    IndentationLevel    int
    InsertionPoint      InsertionPoint     // For before/after clauses
    ConditionalSpecs    []ConditionalSpec  // Conditional value specifications
    SourceLocation      SourceLocation
}
```

**Note on Parser vs Analyzer Responsibilities:**

The parsing and analysis phases have clearly separated responsibilities:

| Aspect | Parsing Phase | Analysis Phase |
|--------|---------------|----------------|
| **Input** | ANTLR parse tree | Symbol table tree built by parser |
| **Output** | Symbol table with `ParserConceptDefinition` entries in `SymbolDefinition.Data` | AST with `AnalysisConceptDefinition` nodes |
| **Data Captured** | Name, Generators (AST structures only), IndentationLevel, InsertionPoint, ConditionalSpecs, SourceLocation | ParentName, ChildNames, Level (hierarchy), ValueIndex, resolved concept values |
| **Processing** | Populates `SymbolDefinition.Data` with `*ParserConceptDefinition` | Reads `SymbolDefinition.Data`, builds `AnalysisConceptDefinition` and `AnalysisConceptValueDefinition` |

**Key Principle:** Value generation happens entirely in the analysis phase; the parser only captures the raw AST structures in the symbol table needed to perform value generation later.

**Flow:**

```
ANTLR Parse Tree ──► Parser ──► Symbol Table (with ParserConceptDefinition in Data)
                                                  │
                                                  ▼
                                      Analysis Phase (WalkSymbolTable)
                                                  │
                                                  ▼
                                              AST Nodes (AnalysisConceptDefinition)
```

### Increment 8: Concepts API Integration

**API Concept Loading**

```go
type ConceptsAPIClient interface {
    // Fetch concepts from the API before parsing begins
    FetchConcepts(conceptNames []string) ([]*APIConceptDefinition, error)
    GetVersion() string
}

type APIConceptDefinition struct {
    Name             string
    Generators       []interface{}  // ValuesGeneratorDef, IntRangeGeneratorDef, or ConceptRangeGeneratorDef
    ParentName       string
    ChildNames       []string
    Version          string    // API concept version for consistency checking
    Source           string    // "api" to distinguish from file-based
}

type ParserWithAPI struct {
    APIClient       ConceptsAPIClient
    FileConcepts    []*ParserConceptDefinition
    APIConcepts     []*APIConceptDefinition
    SymbolTable     *SymbolTable
}
```

**Pre-Parsing API Integration**

```go
func (p *ParserWithAPI) LoadConceptsFromAPI(conceptNames []string) error {
    // Query the Concepts API for relevant concepts before parsing
    apiConcepts, err := p.APIClient.FetchConcepts(conceptNames)
    if err != nil {
        return fmt.Errorf("failed to fetch concepts from API: %w", err)
    }

    p.APIConcepts = apiConcepts

    // Populate symbol table with API concepts
    for _, apiConcept := range apiConcepts {
        def := convertAPIToParserDefinition(apiConcept)
        p.SymbolTable.AddConcept(def)
    }

    return nil
}

func convertAPIToParserDefinition(api *APIConceptDefinition) *ParserConceptDefinition {
    return &ParserConceptDefinition{
        Name:        api.Name,
        Generators:  api.Generators,
        ParentName:  api.ParentName,
        ChildNames:  api.ChildNames,
        // Mark as API-sourced for later version checking
        SourceLocation: SourceLocation{
            Source: "concepts-api",
            Version: api.Version,
        },
    }
}
```

**Validation During Parsing**

- API concepts are pre-populated in the symbol table before file parsing
- File-based definitions can extend API concepts (with IsExtension flag)
- Duplicate detection between API and file concepts (file takes precedence with warning)
- API concept versions are tracked for synthesis-time validation

**Limitations**

None - Full parser concept definition functionality with API integration.

## Options Considered

### Option 1: Single Evolving Structure

**Description:**
Use a single ParserConceptDefinition structure that grows with each increment.

**Pros:**

- Simple, one structure to maintain
- No versioning complexity

**Cons:**

- Early increments have unused fields
- Harder to reason about what's valid in each increment

**Decision:** Selected. The unused fields in early increments are acceptable for simplicity.

### Option 2: Versioned Structures

**Description:**
Create separate structures for each increment (e.g., ParserConceptDefinitionV1, V2, etc.).

**Pros:**

- Clear what's supported in each increment
- Type safety for increment-specific features

**Cons:**

- Duplication and maintenance burden
- Conversion code between versions

**Decision:** Rejected. Unnecessary complexity for this use case.

## Decision

**Chosen Option:** Single evolving ParserConceptDefinition structure

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

Each increment adds validation rules:

1. **Increment 1:** Validate concept names, Values generator values, no duplicates
2. **Increment 2:** Validate indentation levels, parent-child consistency
3. **Increment 3:** Cross-section duplicate detection, extension validation
4. **Increment 4:** Validate conditional spec syntax, single default
5. **Increment 5:** Validate IntRange generator bounds (start <= end, positive step)
6. **Increment 6:** Validate ConceptRange generator syntax, referenced concept existence
7. **Increment 7:** Validate before/after value references, cross-section before/after consistency
8. **Increment 8:** API concept loading, version tracking, API/file concept merging

## Notes

**Related Documents:**

- [Concepts Incremental Implementation](../../eros/ERO001_concepts-incremental-implementation.md) - Main incremental implementation ERO
- [Concept Definitions Design](../../concepts-definitions.md) - Full concept design
- [Symbol Table Design](../symbol-table-design.md) - Symbol table integration details
- [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md) - Architectural pattern for parsing phase design
````
