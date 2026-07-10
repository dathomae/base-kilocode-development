# Postfix Expression Parsing via Symbol Table Stack Pattern

## Context

The CPL (Chronocone Planning Language) grammar includes expression rules with operators that have precedence (e.g., `*`, `/`, `^` for unit algebra). The parser uses ANTLR with listener-based traversal. A key question arises: how should the parser capture expression structure so that operator precedence is preserved for the analysis phase?

## Problem

A naive approach builds expression tree nodes during parsing, with left and right operand fields:

```go
// ANTI-PATTERN: Do not use this approach
type ParserOperatorNode struct {
    Operator UnitOperatorType
    Left     ParserUnitExpression
    Right    ParserUnitExpression
}
```

**Problems with this approach:**

1. **Conflates parser-time and analysis-time concerns**: The parser should capture raw syntactic information, not build semantic trees
2. **Duplicated effort**: The ANTLR parser already builds an AST that encodes precedence; building a parallel tree structure is redundant
3. **Complexity**: Managing left/right operand assembly during listener callbacks is error-prone
4. **Phase boundary violation**: The parser should not be responsible for expression tree construction - that's analysis work

## Solution

Leverage the **grammar structure** and **symbol table stack** to naturally convert infix notation to postfix (reverse Polish) notation through the order of `Exit*` listener method calls.

### How It Works

ANTLR's listener methods are called in a depth-first traversal of the parse tree. For an expression like `a+b/c`, the grammar ensures that `/` binds tighter than `+`. The listener exit calls occur in this order:

```
Source: a + b / c

Grammar rules (simplified):
  expression: expression PLUS term | term
  term:       term DIV factor | factor
  factor:     identifier

Listener Exit call order:
  ExitIdentifier(a)     → push "a" onto definitions
  ExitIdentifier(b)     → push "b" onto definitions
  ExitDivOperator       → push DIV operator onto definitions
  ExitIdentifier(c)     → push "c" onto definitions
  ExitExpression        → complete (definitions are in postfix order)
```

The definitions slice ends up containing: `["a", "b", DIV, "c"]` — which is postfix notation for `a + (b / c)`.

The analysis phase can then evaluate this postfix sequence using a simple stack-based algorithm, naturally respecting operator precedence.

### Implementation Pattern

**Grammar Structure:**

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
// ExitIdentifier pushes the identifier onto the current scope's definitions.
func (l *UnitsListener) ExitIdentifier(ctx *parser.IdentifierContext) {
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

// ExitMultOperator pushes the multiplication operator onto definitions.
func (l *UnitsListener) ExitMultOperator(ctx *parser.MultOperatorContext) {
    current, _ := l.symbolStack.Peek()
    location := sourceLocationFromContext(ctx)

    symDef := SymbolDefinition{
        DefinitionType: MultiplyOperatorDefinition,
        Location:       location,
        Data:           &ParserOperator{Operator: OpMultiply},
    }
    current.Definitions = append(current.Definitions, symDef)
}

// ExitDivOperator pushes the division operator onto definitions.
func (l *UnitsListener) ExitDivOperator(ctx *parser.DivOperatorContext) {
    current, _ := l.symbolStack.Peek()
    location := sourceLocationFromContext(ctx)

    symDef := SymbolDefinition{
        DefinitionType: DivideOperatorDefinition,
        Location:       location,
        Data:           &ParserOperator{Operator: OpDivide},
    }
    current.Definitions = append(current.Definitions, symDef)
}

// ExitPowOperator pushes the power operator onto definitions.
func (l *UnitsListener) ExitPowOperator(ctx *parser.PowOperatorContext) {
    current, _ := l.symbolStack.Peek()
    location := sourceLocationFromContext(ctx)

    symDef := SymbolDefinition{
        DefinitionType: PowerOperatorDefinition,
        Location:       location,
        Data:           &ParserOperator{Operator: OpPower},
    }
    current.Definitions = append(current.Definitions, symDef)
}
```

**Parser-Time Data Structures:**

```go
// ParserUnitReference captures a simple unit name reference.
type ParserUnitReference struct {
    Name string
}

// ParserOperator captures an operator token.
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

**Analysis Phase Evaluation:**

The analysis phase walks the definitions slice (which is in postfix order) and builds the expression tree:

```go
func (a *CPLAnalyzer) BuildExpressionTree(defs []SymbolDefinition) (Expression, error) {
    var stack []Expression

    for _, def := range defs {
        switch def.DefinitionType {
        case UnitReferenceDefinition:
            if ref, ok := def.Data.(*ParserUnitReference); ok {
                unit, exists := a.Units[ref.Name]
                if !exists {
                    return nil, fmt.Errorf("undefined unit: %s", ref.Name)
                }
                stack = append(stack, &UnitReference{Name: ref.Name, UnitGroup: unit})
            }
        case MultiplyOperatorDefinition, DivideOperatorDefinition, PowerOperatorDefinition:
            if len(stack) < 2 {
                return nil, fmt.Errorf("operator %v with insufficient operands", def.DefinitionType)
            }
            right := stack[len(stack)-1]
            left := stack[len(stack)-2]
            stack = stack[:len(stack)-2]

            var node Expression
            switch def.DefinitionType {
            case MultiplyOperatorDefinition:
                node = &MultiplicationNode{Left: left, Right: right}
            case DivideOperatorDefinition:
                node = &DivisionNode{Left: left, Right: right}
            case PowerOperatorDefinition:
                node = &PowerNode{Base: left, Exponent: right}
            }
            stack = append(stack, node)
        }
    }

    if len(stack) != 1 {
        return nil, fmt.Errorf("malformed expression: %d items remaining on stack", len(stack))
    }
    return stack[0], nil
}
```

### Scope Management for Expressions

Expressions appear within various contexts (measure definitions, pattern clauses, calculation clauses). The scope management handles this naturally:

1. When entering a context that contains an expression: push the appropriate scope
2. Expression components (identifiers, operators) are pushed to that scope's definitions
3. When exiting the context: pop the scope
4. The analysis phase processes each scope's definitions in order

## Consequences

### Benefits

| Aspect                  | Postfix via Symbol Table Stack                    | Expression Tree During Parsing             |
| ----------------------- | ------------------------------------------------- | ------------------------------------------ |
| **Parser simplicity**   | Listener methods just push to definitions slice   | Must manage left/right operand assembly    |
| **Precedence handling** | Handled by grammar + listener call order          | Must be manually encoded in tree building  |
| **Phase separation**    | Parser captures tokens; analysis builds trees     | Parser builds partial semantic structures  |
| **Extensibility**       | Add new operators with new listener + definition  | Must modify tree node types and builders   |
| **Parentheses**         | Handled naturally by grammar recursion            | Requires special case handling             |

### Trade-offs

- **Analysis phase complexity**: Must implement postfix evaluation to build expression trees
- **Definition slice ordering**: Relies on correct listener call order (guaranteed by ANTLR grammar)

## When to Use

Use this pattern when:

- Parsing expressions with operator precedence using ANTLR listeners
- Maintaining strict separation between parsing (syntax) and analysis (semantics)
- Want to avoid building intermediate tree structures during parsing
- The grammar already encodes operator precedence through rule structure

## Related Patterns

- [Symbol Table-Driven Parsing](symbol-table-driven-parsing.md) - The overarching pattern for using symbol tables to track parsing context
- [Layered Architecture](layered-architecture.md) - The parsing/analysis phase separation

## References

- [ERO003 Units and Measures Parsing Phase](../design/cpl/go-parser-target/eros/ERO003_units-measures-parsing-phase.md) - Incremental implementation using this pattern
- [ERO003 Units and Measures Analysis Phase](../design/cpl/go-analysis/eros/ERO003_units-measures-analysis-phase.md) - Analysis phase expression tree construction
