# Symbol Table-Driven Parsing Pattern

## Context

The CPL (Chronocone Planning Language) compiler uses ANTLR-generated parsers with listener-based AST construction. As the language grows to support multiple declaration types (units, concepts, connectors, metrics, aggregators, spreaders, etc.), a critical architectural decision emerges: how should identifier handling work across different declaration contexts?

## Problem

A naive approach uses type switches in the identifier listener to determine context based on parent node types:

```go
// ANTI-PATTERN: Do not use this approach
func (l *CPLListener) ExitIdentifier(ctx *antlr.IdentifierContext) {
    text := ctx.IDENTIFIER().GetText()
    parent := ctx.GetParent()

    switch parent.(type) {
    case *antlr.UnitDeclarationContext:
        l.currentUnit.Name = text
    case *antlr.ConceptDefinitionContext:
        l.currentConcept.Name = text
    case *antlr.ConnectorDeclarationContext:
        l.currentConnector.Name = text
    // ... must add new case for every declaration type
    }
}
```

**Problems with this approach:**

1. **Open/Closed Principle Violation**: Adding a new declaration type requires modifying the centralized `ExitIdentifier` method
2. **Tight Coupling**: The identifier handler must know about every declaration type in the language
3. **Mixed Concerns**: Identifier rule logic contains declaration-specific handling
4. **Fragile State Management**: Relies on transient state flags (`currentUnit`, `currentConcept`) that must be carefully synchronized
5. **Poor Separation**: Parsing is mixed with partial AST building

## Solution

Use the **symbol table stack** as the sole mechanism for tracking parsing context. The parser's responsibility is limited to:

1. **Pushing/popping scope tables** as declaration contexts are entered and exited
2. **Defining symbols** (names and definition types) in the current symbol table
3. **Attaching parse-context data** to `SymbolDefinition.Data` if needed by the analysis phase

Declaration assembly happens during the **analysis phase** when walking the symbol table tree.

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        PARSING PHASE                            │
│                                                                 │
│  ANTLR Parse Tree ──► Listener Methods ──► Symbol Table Tree   │
│                                                                 │
│  • EnterDeclaration: Push scope table                          │
│  • ExitIdentifier:   Define symbol in current table            │
│  • ExitDeclaration:  Pop scope table                           │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       ANALYSIS PHASE                            │
│                                                                 │
│  Symbol Table Tree ──► Walker ──► Analysis Definitions         │
│                                                                 │
│  • WalkSymbolTable visits each scope                           │
│  • Assemble declarations from symbol definitions               │
│  • Build analysis-level structures                             │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Implementation Pattern

**Phase 1: Parsing - Populate Symbol Table**

```go
// Each declaration type manages its own scope
func (l *CPLListener) EnterUnitDeclaration(ctx *antlr.UnitDeclarationContext) {
    current, _ := l.symbolStack.Peek()
    unitTable := current.AddChild(symboltable.UnitsScope)
    l.symbolStack.Push(unitTable)
}

// Generic identifier handling - no switches, no context awareness
func (l *CPLListener) ExitIdentifier(ctx *antlr.IdentifierContext) {
    // The identifier is simply registered in the current scope
    // The scope type (from the symbol table) provides the context
    current, _ := l.symbolStack.Peek()
    name := ctx.IDENTIFIER().GetText()
    location := symboltable.SourceLocation{
        Line:   ctx.GetStart().GetLine(),
        Column: ctx.GetStart().GetColumn(),
    }

    // Define symbol with appropriate type based on current scope
    current.Define(&name, inferDefinitionType(current.ScopeType), location)
}

func (l *CPLListener) ExitUnitDeclaration(ctx *antlr.UnitDeclarationContext) {
    l.symbolStack.Pop()
}
```

**Phase 2: Analysis - Walk Symbol Table Tree**

```go
// Analysis phase walks the tree and assembles declarations
func AnalyzeUnitsScope(table *symboltable.SymbolTable) ([]UnitDecl, error) {
    var units []UnitDecl

    for _, def := range table.Definitions {
        if def.DefinitionType == symboltable.UnitDefinition {
            unit := UnitDecl{
                Name: *def.Name,
                // ... assemble from symbol definition
            }
            units = append(units, unit)
        }
    }

    return units, nil
}
```

## Consequences

### Benefits

| Aspect                     | Symbol Table-Driven                                            | Switch-Based                                        |
| -------------------------- | -------------------------------------------------------------- | --------------------------------------------------- |
| **Extensibility**          | Add new declaration type by adding scope type and walker       | Modify centralized switch for each new type         |
| **Separation of Concerns** | Parsing only builds symbol table; analysis builds declarations | Mixed parsing and partial AST construction          |
| **Testability**            | Each declaration type testable in isolation                    | Complex to test all switch branches                 |
| **Maintainability**        | Distributed, stack-based context                               | Centralized, fragile state                          |
| **Coupling**               | Identifier handler knows nothing about declarations            | Identifier handler coupled to all declaration types |

### Trade-offs

- **Slightly more complex analysis phase**: The analysis phase must walk the symbol table tree rather than receiving pre-built AST nodes
- **Delayed error detection**: Some semantic errors may only be detected during analysis, not parsing

## When to Use

Use this pattern when:

- Building a compiler or interpreter with multiple declaration types
- Using ANTLR or similar parser generators with listener/visitor patterns
- Need to support language extensibility without modifying core parser logic
- Want clean separation between parsing (syntax) and analysis (semantics)

## Example in Codebase

The concept definition parsing in [`cpl_concepts_listener.go`](../../../src/chronocone_planning_language/src/main/go/parser/cpl_concepts_listener.go) follows this pattern:

1. `EnterConceptsSection` pushes a `ConceptsScope` table (line 36)
2. `ExitConceptDefinition` defines the concept in the symbol table (lines 133-152)
3. Analysis phase walks the tree to build full concept definitions

## Common Mistakes and Corrections

This section documents lessons learned from implementing the symbol table-driven parsing pattern, including common mistakes and their corrections.

### Mistake 1: Doing Too Much in Exit Methods

**Initial Approach (Incorrect):**

```go
// Anti-pattern: Exit handler tries to assemble declarations
func (l *CPLListener) ExitTypeDefinition(ctx *antlr.TypeDefinitionContext) {
    typeName := ctx.Identifier().GetText()  // Directly accessing context

    // Trying to extract all data here
    baseType := ctx.BaseType().GetText()
    constraints := l.parseConstraints(ctx.ConstraintList())

    // Creating AST structures in the parser
    typeDef := ast.TypeDefinition{
        Name:        typeName,
        BuiltInType: ast.NumberType,  // Hardcoded, not analyzing
        Constraints: constraints,
    }

    // Pushing to parent scope - too much work for parsing phase
    parent.DefineWithData(&typeName, symboltable.TypeDefinition, ..., typeDef)
}
```

**Problems:**

1. Parser is doing analysis work (extracting constraints, determining built-in types)
2. AST structures (`ast.TypeDefinition`) belong in the analysis phase, not parsing
3. Hardcoded `ast.NumberType` assumes all base types are built-in (wrong for named base types)
4. Mixes parsing (syntax) with semantic analysis

**Correct Approach:**

```go
// Symbol table-driven parsing: exit handler only manages scope
func (l *CPLListener) ExitTypeDefinition(ctx *antlr.TypeDefinitionContext) {
    // Pop the type scope from the symbol stack.
    // The TypeScope and its definitions (added by EnterIdentifier)
    // persist in the symbol table tree as a child of TypesScope.
    // Analysis phase will walk the tree and extract type definitions.
    _, err := l.symbolStack.Pop()
    if err != nil {
        l.AddError(fmt.Errorf("symbol stack error exiting type definition: %w", err))
    }
}
```

**Why This Is Correct:**

1. Parser only manages scope (push/pop symbol tables)
2. Identifiers are added by `EnterIdentifier` to the current scope (TypeScope)
3. The TypeScope and its definitions persist in the symbol table tree
4. Analysis phase (under `src/chronocone_planning_language/src/main/go/analysis`) walks the tree and extracts type definitions from TypeScope children

### Mistake 2: Not Pushing Scopes for Nested Structures

**Initial Approach (Incorrect):**

```go
// Anti-pattern: constraint processing without scope
func (l *CPLListener) EnterConstraint(ctx *antlr.ConstraintContext) {
    l.logEnter("Constraint", &ctx.BaseParserRuleContext)
    // No scope pushed - constraint data gets lost
}

func (l *CPLListener) EnterConstraintValue(ctx *antlr.ConstraintValueContext) {
    l.logEnter("ConstraintValue", &ctx.BaseParserRuleContext)
    // Values have no place to go - symbol stack wasn't updated
}
```

**Correct Approach:**

```go
// Push a scope for constraints to hold identifier and value
func (l *CPLListener) EnterConstraint(ctx *antlr.ConstraintContext) {
    l.logEnter("Constraint", &ctx.BaseParserRuleContext)

    current, err := l.symbolStack.Peek()
    if err != nil {
        l.AddError(...)
        return
    }

    // Push a new scope for this constraint
    l.symbolStack.Push(current.AddChild(symboltable.TypeScope))
}

func (l *CPLListener) ExitConstraint(ctx *antlr.ConstraintContext) {
    // Pop the constraint scope from the symbol stack
    _, err := l.symbolStack.Pop()
    if err != nil {
        l.AddError(...)
    }
    l.logExit("Constraint", &ctx.BaseParserRuleContext)
}

// Constraint value pushed onto symbol table for later analysis
func (l *CPLListener) EnterConstraintValue(ctx *antlr.ConstraintValueContext) {
    current, err := l.symbolStack.Peek()
    if err != nil {
        l.AddError(...)
        return
    }

    value := ctx.GetText()
    err = current.DefineWithData(
        nil, // anonymous definition
        symboltable.Identifier,
        symboltable.SourceLocation{...},
        value, // Store value in Data field
    )
    if err != nil {
        l.AddError(...)
    }
}
```

## Analysis Phase Implementation

The modern analysis approach is implemented under `src/chronocone_planning_language/src/main/go/analysis/`:

### Core Components

1. **Symbol Table Walker** ([`symbol_table_walk.go`](src/chronocone_planning_language/src/main/go/analysis/symbol_table_walk.go)):

```go
// WalkSymbolTableWithHandlers recursively walks the symbol table tree,
// calling the appropriate analysis function for each scope type.
func WalkSymbolTableWithHandlers(
    table *symboltable.SymbolTable,
    enterHandlers map[symboltable.ScopeType]AnalysisFunc,
    exitHandlers map[symboltable.ScopeType]AnalysisFunc,
    analyzer *CPLAnalyzer) error
```

2. **Analysis Context** - A stack-based context for building up analysis structures during the walk:

```go
type AnalysisContext struct {
    // Stack of partially-built analysis definitions
}

func (ac *AnalysisContext) Push(definition *AnalysisDefinition) error
func (ac *AnalysisContext) Pop() (*AnalysisDefinition, error)
```

3. **Analysis Handlers** ([`concept_analysis.go`](src/chronocone_planning_language/src/main/go/analysis/concept_analysis.go)):

```go
// Enter handler for concept scopes
func EnterAnalyzeConceptScope(
    table *symboltable.SymbolTable,
    context *AnalysisContext,
    analyzer *CPLAnalyzer) error {

    // Extract concept name from symbol table
    concept := NewAnalysisConceptDefinition(*table.Definitions[0].Name, ...)
    return context.Push(concept)
}

// Exit handler for concept scopes
func ExitAnalyzeConceptScope(
    table *symboltable.SymbolTable,
    context *AnalysisContext,
    analyzer *CPLAnalyzer) error {

    concept, _ := context.Pop()
    return analyzer.AddConcept(concept.ConceptDefinition)
}
```

### Analysis Definition Structure

The analysis phase uses a unified `AnalysisDefinition` structure:

```go
type AnalysisDefinition struct {
    DefinitionType DefinitionType
    ConceptDefinition *AnalysisConceptDefinition  // For concepts
    TypeDefinition *AnalysisTypeDefinition        // For types (future)
    // Other definition types as needed
}
```

## Related Patterns

- [Layered Architecture](layered-architecture.md) - The parsing/analysis separation mirrors the layered architecture pattern
- [ANTLR Listener Organization](layered-architecture.md) - How to organize listener files when using this pattern with ANTLR

## References

- [Symbol Table Design](../design/cpl/go-parser-target/symbol-table-design.md) - Detailed symbol table implementation
- [CPL Architecture Introduction](../architecture/cpl/introduction.md) - Overview of parsing and analysis phases
- [Analysis Phase ERO001](../design/cpl/go-analysis/eros/ERO001_concepts-analysis-phase.md) - Incremental analysis implementation for concepts
- [Analysis Phase ERO002](../design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md) - Incremental analysis implementation for types
