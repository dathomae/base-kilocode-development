# ERO003: Units and Measures Analysis Phase

## Title

Units and Measures Analysis Phase - Data Structures and Incremental Implementation

## Purpose/Goal

Define the analysis phase data structures for units and measures definitions and document how parser-level structures are transformed into analysis-level structures. This ERO explicitly separates parsing-time concerns (raw syntactic capture) from analysis-time concerns (semantic resolution and Expression tree building).

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic003: Units and Measures Implementation](../../../epics/Epic003_units-and-measures-implementation.md) - Epic tracking the implementation of units and measures

**Related Stories:**

- ~~Story141~~ through ~~Story150~~ - Superseded stories (old implementation incompatible with types system)
- Implemented without formal story files: Basic units and measures parser/analyzer (see `cpl_units_listener.go`, `units_analysis.go`, `cpl_measures_listener.go`, `measures_analysis.go`)
- Story228 - CPL Expression Infrastructure (shared grammar, BuildExpressionTree, CIL expression types)
- Story229 - CPL Compound Units (expression-based unit specifications in measures)
- Story230 - CPL Conversions (unit and measure conversion pairs with pattern/calculation clauses)
- Story231 - Unit Type Checking (synthesis-time unit type verification on CIL expression trees)
- Story232 - Conversion Injection (automatic conversion node insertion into CIL)

## Context

### Situation

The analysis phase is responsible for validating units and measure definitions, building the Units and Measures lookup maps, and enabling semantic analysis of unit references in metrics and other CPL constructs. The analysis phase transforms parser-level definitions into analysis-level structures optimized for validation and code generation.

A critical separation of concerns exists between parsing and analysis for unit expressions:

- **Parser-time concern**: The parser captures what was written in source code - raw unit names and operators
- **Analysis-time concern**: Resolving those names to unit definitions and building Expression trees for dimensional analysis

The parsing phase ERO has been updated to use the [Postfix Expression Parsing Pattern](../../../../patterns/postfix-expression-parsing.md), which separates these concerns by capturing unit expression components as raw symbols in postfix order during parsing, leaving semantic tree construction to the analysis phase.

### Constraints

- Must work with parser-level unit and measure definitions from the symbol table
- Must produce structures usable by the CIL generator
- Must validate all unit and measure references in the CPL program
- Must support the incremental feature set of each increment
- Unit expressions must be resolved to their component units during analysis
- Measure unit references must be validated against the Units lookup map

### Requirements

- Build a unified Units lookup map from parser definitions
- Build a unified Measures lookup map from parser definitions
- Transform raw unit expressions into resolved Expression trees
- Validate unit and measure references
- Support unit algebra for dimensional correctness checking
- Enable compound unit simplification and normalization
- Support conversion pattern validation

### Assumptions

- Parser unit and measure definitions are complete and syntactically valid
- The analysis phase has access to all parsed files in the compilation unit
- Analysis errors prevent CIL generation
- Concepts have been analyzed before units and measures (units depend on concepts)

## Parsing vs Analysis Phase Boundary

### The Separation Principle

Following the pattern established in ERO001 concepts analysis, there is a clear boundary between parser-time and analysis-time data. Unit expressions are captured using the [Postfix Expression Parsing Pattern](../../../../patterns/postfix-expression-parsing.md) and [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md), where the parser pushes raw symbols to the scope's Definitions slice in postfix order, leaving semantic tree construction to the analysis phase.

| Aspect              | Parser-Time                             | Analysis-Time                                         |
| ------------------- | --------------------------------------- | ----------------------------------------------------- |
| **Data**            | Raw strings, token sequences            | Resolved references, computed structures              |
| **Unit Expression** | `[]SymbolDefinition` - postfix sequence | `Expression` - semantic tree with resolved references |
| **Unit References** | Unit names as strings                   | Pointers to `AnalysisUnitDefinition` records          |
| **Validation**      | Syntax only (valid identifiers)         | Semantic (unit existence, compatibility)              |
| **Tree Building**   | None - flat capture                     | Expression tree with operator nodes                   |

### Data Flow Across the Boundary

```
ParserMeasureDefinition (from SymbolDefinition.Data)
    ├─ Name: string
    ├─ TypeName: string
    └─ SourceLocation: SourceLocation

MeasureScope.Definitions (postfix expression sequence)
    ├─ SymbolDefinition{Data: &ParserUnitReference{Name: "Length"}}
    ├─ SymbolDefinition{Data: &ParserUnitReference{Name: "Time"}}
    └─ SymbolDefinition{Data: &ParserOperator{Operator: OpDivide}}

        ↓ Transform during analysis

AnalysisMeasureDefinition
    ├─ Name: string
    ├─ TypeName: string
    ├─ UnitExpression: Expression            ← Analysis-time: resolved tree
    └─ SourceLocation: SourceLocation
```

## Data Structures by Increment

### Increment 1: Basic Unit Definitions

**AnalysisDefinitions - Units Lookup Map**

```go
// AnalysisDefinitions contains all analyzed definitions for CIL generation.
type AnalysisDefinitions struct {
    Concepts map[string]AnalysisConceptDefinition  // From ERO001
    Units    map[string]AnalysisUnitDefinition     // NEW: Units lookup map
    // Measures added in Increment 2
}

// AnalysisUnitDefinition contains the analyzed unit information.
// At Increment 1, units are simple concept references with no conversions.
type AnalysisUnitDefinition struct {
    Name           string                      // Unit group identifier
    ConceptName    string                      // Referenced concept name
    Concept        *AnalysisConceptDefinition  // Pointer to resolved concept
    SourceLocation SourceLocation              // For error reporting
}
```

**Units Analysis Function**

```go
// AnalyzeUnitsScope processes unit definitions from a UnitsScope
// symbol table and builds the analysis-level units lookup map.
func (a *CPLAnalyzer) AnalyzeUnitsScope(table *SymbolTable) error {
    for _, def := range table.Definitions {
        if def.DefinitionType != UnitScopeDefinition {
            continue
        }

        parserDef, ok := def.Data.(*ParserUnitDefinition)
        if !ok {
            return fmt.Errorf("invalid unit definition data for %s", *def.Name)
        }

        // Check for duplicate unit definitions
        if _, exists := a.Units[*def.Name]; exists {
            return fmt.Errorf("duplicate unit definition: %s", *def.Name)
        }

        // Resolve the concept reference
        concept, exists := a.Concepts[parserDef.ConceptName]
        if !exists {
            return fmt.Errorf("unit %s references undefined concept: %s",
                *def.Name, parserDef.ConceptName)
        }

        // Create analysis-level unit definition
        analysisDef := &AnalysisUnitDefinition{
            Name:           *def.Name,
            ConceptName:    parserDef.ConceptName,
            Concept:        concept,
            SourceLocation: parserDef.SourceLocation,
        }

        a.Units[*def.Name] = analysisDef
    }
    return nil
}
```

**Integration with WalkSymbolTableWithHandlers**

```go
func (c *CPLCompiler) Analyze(rootTable *SymbolTable) (*AnalysisDefinitions, error) {
    analyzer := NewCPLAnalyzer()

    // Register handlers for each scope type
    handlers := map[ScopeType]AnalysisFunc{
        ConceptsScope: analyzer.AnalyzeConceptsScope,  // From ERO001
        UnitsScope:    analyzer.AnalyzeUnitsScope,     // NEW: Increment 1
    }

    if err := WalkSymbolTableWithHandlers(rootTable, handlers); err != nil {
        return nil, fmt.Errorf("analysis failed: %w", err)
    }

    return analyzer.GetAnalysisDefinitions(), nil
}
```

**Validation During Analysis**

- Unit group name uniqueness across the compilation unit
- Concept reference must exist in the Concepts lookup map
- No cycles or dependencies to validate at Increment 1

**Limitations**

- No conversion clauses
- No compound units
- No measure definitions yet

### Increment 2: Basic Measure Definitions

**AnalysisMeasureDefinition**

```go
// AnalysisMeasureDefinition contains the analyzed measure information.
// At Increment 2, measures reference simple unit groups (not compound expressions).
type AnalysisMeasureDefinition struct {
    Name            string                     // Measure group identifier
    TypeName        string                     // Type name (e.g., "Number")
    UnitGroupName   string                     // Referenced unit group name
    UnitGroup       *AnalysisUnitDefinition    // Pointer to resolved unit group
    SourceLocation  SourceLocation             // For error reporting
}
```

**Enhanced AnalysisDefinitions**

```go
type AnalysisDefinitions struct {
    Concepts map[string]AnalysisConceptDefinition
    Units    map[string]AnalysisUnitDefinition
    Measures map[string]AnalysisMeasureDefinition  // NEW: Measures lookup map
}
```

**Measures Analysis Function**

```go
// AnalyzeMeasuresScope processes measure definitions from a MeasuresScope
// symbol table and builds the analysis-level measures lookup map.
func (a *CPLAnalyzer) AnalyzeMeasuresScope(table *SymbolTable) error {
    for _, def := range table.Definitions {
        if def.DefinitionType != MeasureScopeDefinition {
            continue
        }

        parserDef, ok := def.Data.(*ParserMeasureDefinition)
        if !ok {
            return fmt.Errorf("invalid measure definition data for %s", *def.Name)
        }

        // Check for duplicate measure definitions
        if _, exists := a.Measures[*def.Name]; exists {
            return fmt.Errorf("duplicate measure definition: %s", *def.Name)
        }

        // Validate type name (at Increment 2, only "Number" is supported)
        if parserDef.TypeName != "Number" {
            return fmt.Errorf("measure %s: only Number type is supported at this increment, got %s",
                *def.Name, parserDef.TypeName)
        }

        // Resolve the unit group reference
        unitGroup, exists := a.Units[parserDef.UnitGroupName]
        if !exists {
            return fmt.Errorf("measure %s references undefined unit group: %s",
                *def.Name, parserDef.UnitGroupName)
        }

        // Create analysis-level measure definition
        analysisDef := &AnalysisMeasureDefinition{
            Name:           *def.Name,
            TypeName:       parserDef.TypeName,
            UnitGroupName:  parserDef.UnitGroupName,
            UnitGroup:      unitGroup,
            SourceLocation: parserDef.SourceLocation,
        }

        a.Measures[*def.Name] = analysisDef
    }
    return nil
}
```

**Updated Handler Registration**

```go
handlers := map[ScopeType]AnalysisFunc{
    ConceptsScope: analyzer.AnalyzeConceptsScope,
    UnitsScope:    analyzer.AnalyzeUnitsScope,
    MeasuresScope: analyzer.AnalyzeMeasuresScope,  // NEW: Increment 2
}
```

**Validation During Analysis**

- Measure group name uniqueness across the compilation unit
- Type name must be supported at current increment (Increment 2: only "Number")
- Unit group reference must exist in the Units lookup map

**Limitations**

- Only simple unit group references (no compound expressions)
- No measure conversions
- Single unit group per measure

### Increment 3: Compound Units and Unit Algebra Tree

**Expression Tree Structure (Analysis-Time)**

The Expression tree is an analysis-time structure representing resolved unit algebra:

```go
// Expression is the interface for all analysis-time expression nodes.
// This tree represents resolved unit expressions with validated references.
type Expression interface {
    isExpression()
    String() string  // Returns normalized unit expression string
}

// UnitReference represents a resolved unit reference.
// At analysis time, the unit name has been validated against the Units map.
type UnitReference struct {
    Name       string                      // Unit name
    UnitGroup  *AnalysisUnitDefinition     // Resolved pointer to unit definition
}

func (u UnitReference) isExpression() {}
func (u UnitReference) String() string { return u.Name }

// MultiplicationNode represents unit multiplication (e.g., Length * Time).
type MultiplicationNode struct {
    Left  Expression
    Right Expression
}

func (m MultiplicationNode) isExpression() {}
func (m MultiplicationNode) String() string {
    return fmt.Sprintf("(%s * %s)", m.Left.String(), m.Right.String())
}

// DivisionNode represents unit division (e.g., Length / Time).
type DivisionNode struct {
    Left  Expression
    Right Expression
}

func (d DivisionNode) isExpression() {}
func (d DivisionNode) String() string {
    return fmt.Sprintf("(%s / %s)", d.Left.String(), d.Right.String())
}

// PowerNode represents unit exponentiation (e.g., Meter^2).
type PowerNode struct {
    Base     Expression
    Exponent int
}

func (p PowerNode) isExpression() {}
func (p PowerNode) String() string {
    return fmt.Sprintf("(%s^%d)", p.Base.String(), p.Exponent)
}
```

**Postfix Expression Evaluation**

Following the [Postfix Expression Parsing via Symbol Table Stack Pattern](../../../../patterns/postfix-expression-parsing.md), the parser pushes expression components to the scope's Definitions slice in postfix order. The analysis phase evaluates this sequence using a stack-based algorithm:

```go
// BuildExpressionTree evaluates a postfix sequence of symbol definitions
// into a resolved Expression tree.
func (a *CPLAnalyzer) BuildExpressionTree(defs []SymbolDefinition) (Expression, error) {
    var stack []Expression

    for _, def := range defs {
        switch def.DefinitionType {
        case UnitReferenceDefinition:
            ref, ok := def.Data.(*ParserUnitReference)
            if !ok {
                return nil, fmt.Errorf("invalid unit reference data")
            }
            unit, exists := a.Units[ref.Name]
            if !exists {
                return nil, fmt.Errorf("undefined unit: %s", ref.Name)
            }
            stack = append(stack, UnitReference{Name: ref.Name, UnitGroup: unit})

        case MultiplyOperatorDefinition:
            if len(stack) < 2 {
                return nil, fmt.Errorf("multiply operator with insufficient operands")
            }
            right := stack[len(stack)-1]
            left := stack[len(stack)-2]
            stack = stack[:len(stack)-2]
            stack = append(stack, MultiplicationNode{Left: left, Right: right})

        case DivideOperatorDefinition:
            if len(stack) < 2 {
                return nil, fmt.Errorf("divide operator with insufficient operands")
            }
            right := stack[len(stack)-1]
            left := stack[len(stack)-2]
            stack = stack[:len(stack)-2]
            stack = append(stack, DivisionNode{Left: left, Right: right})

        case PowerOperatorDefinition:
            if len(stack) < 2 {
                return nil, fmt.Errorf("power operator with insufficient operands")
            }
            exponentRef, ok := def.Data.(*ParserUnitReference)
            if !ok {
                return nil, fmt.Errorf("invalid power exponent")
            }
            exponent, err := strconv.Atoi(exponentRef.Name)
            if err != nil {
                return nil, fmt.Errorf("power exponent must be integer: %s", exponentRef.Name)
            }
            base := stack[len(stack)-1]
            stack = stack[:len(stack)-1]
            stack = append(stack, PowerNode{Base: base, Exponent: exponent})
        }
    }

    if len(stack) != 1 {
        return nil, fmt.Errorf("malformed expression: %d items remaining on stack", len(stack))
    }
    return stack[0], nil
}
```

**Enhanced AnalysisMeasureDefinition with Expression Tree**

```go
// AnalysisMeasureDefinition for Increment 3 includes the resolved Expression tree.
type AnalysisMeasureDefinition struct {
    Name            string                     // Measure group identifier
    TypeName        string                     // Type name (e.g., "Number")
    UnitExpression  Expression                 // Resolved compound unit expression tree
    SourceLocation  SourceLocation             // For error reporting
}
```

**Updated AnalyzeMeasuresScope for Increment 3**

```go
func (a *CPLAnalyzer) AnalyzeMeasuresScope(table *SymbolTable) error {
    for _, def := range table.Definitions {
        if def.DefinitionType != MeasureScopeDefinition {
            continue
        }

        parserDef, ok := def.Data.(*ParserMeasureDefinition)
        if !ok {
            return fmt.Errorf("invalid measure definition data for %s", *def.Name)
        }

        if _, exists := a.Measures[*def.Name]; exists {
            return fmt.Errorf("duplicate measure definition: %s", *def.Name)
        }

        if parserDef.TypeName != "Number" {
            return fmt.Errorf("measure %s: only Number type is supported, got %s",
                *def.Name, parserDef.TypeName)
        }

        // Build expression tree from postfix definitions in the measure scope
        unitExpr, err := a.BuildExpressionTree(def.Scope.Definitions)
        if err != nil {
            return fmt.Errorf("measure %s: invalid unit expression: %w",
                *def.Name, err)
        }

        analysisDef := &AnalysisMeasureDefinition{
            Name:           *def.Name,
            TypeName:       parserDef.TypeName,
            UnitExpression: unitExpr,
            SourceLocation: parserDef.SourceLocation,
        }

        a.Measures[*def.Name] = analysisDef
    }
    return nil
}
```

**Validation During Analysis**

- All unit references in the expression must exist in the Units map
- Exponent must be an integer (already validated during parsing)
- Expression structure is syntactically valid (already validated during parsing)
- No dimensional analysis at Increment 3 (expressions preserved symbolically)

**Limitations**

- No unit conversions (cannot simplify compound units)
- No derived unit definitions
- Unit trees are preserved symbolically without dimensional simplification
- No tensor support yet (tree designed for future extension)

### Increment 4: Unit Conversions

**Analysis-Time Conversion Structures**

Conversion data comes from the symbol table tree, NOT from parsed structs. The parser stores conversions as child symbol table scopes (ConversionScope containing PatternScope and CalculationScope children). The analysis phase walks these scopes to extract pattern-bound variables and calculation expressions.

```go
// AnalysisConversionPair captures a unit conversion in resolved form.
// The analysis phase derives the inverse conversion pair from each forward pair.
// Both forward and inverse are stored so synthesis sees explicit pairs in CIL.
type AnalysisConversionPair struct {
    SourceUnit    *AnalysisUnitDefinition    // Source unit (e.g., Kilometer)
    TargetUnit    *AnalysisUnitDefinition    // Target unit (e.g., Meter)
    SourceVar     string                     // Pattern-bound source variable (e.g., "km")
    TargetVar     string                     // Pattern-bound target variable (e.g., "m")
    Calculation   Expression                 // Resolved calculation expression
}

// Enhanced AnalysisUnitDefinition with conversions.
type AnalysisUnitDefinition struct {
    Name            string
    ConceptName     string
    Concept         *AnalysisConceptDefinition
    Conversions     []AnalysisConversionPair    // NEW: Conversion definitions
    SourceLocation  SourceLocation
}
```

**Conversion Analysis — Symbol Table Walking**

The analysis phase walks the symbol table tree to extract conversion pairs. Conversions are NOT passed as parsed struct arrays. The parser stores them as child scopes under UnitScope.

```go
// AnalyzeUnitConversions walks UnitScope children for ConversionScope entries,
// pairs PatternScope with CalculationScope, extracts pattern-bound variables,
// builds calculation expression trees, and derives inverse conversions.
func (a *CPLAnalyzer) AnalyzeUnitConversions(
    unitScope *SymbolTable,
    unitDef *AnalysisUnitDefinition) error {

    for _, child := range unitScope.Children {
        if child.ScopeType != ConversionScope {
            continue
        }

        // Pair PatternScope with CalculationScope by order within ConversionScope
        var patternScope, calculationScope *SymbolTable
        for _, convChild := range child.Children {
            switch convChild.ScopeType {
            case PatternScope:
                patternScope = convChild
            case CalculationScope:
                calculationScope = convChild
            }
        }

        if patternScope == nil || calculationScope == nil {
            return fmt.Errorf("conversion clause missing pattern or calculation")
        }

        // Extract pattern-bound variables from the PatternScope definitions.
        // Equation form: UnitName(var) = UnitName(var)
        // First unit+variable = target, second = source.
        targetUnitRef, targetVarName, err := extractBoundUnitRef(patternScope, 0)
        if err != nil {
            return fmt.Errorf("conversion pattern: %w", err)
        }
        sourceUnitRef, sourceVarName, err := extractBoundUnitRef(patternScope, 1)
        if err != nil {
            return fmt.Errorf("conversion pattern: %w", err)
        }

        // Resolve unit references against the Units lookup map
        targetUnit, exists := a.Units[targetUnitRef.Name]
        if !exists {
            return fmt.Errorf("conversion pattern references undefined unit: %s",
                targetUnitRef.Name)
        }
        sourceUnit, exists := a.Units[sourceUnitRef.Name]
        if !exists {
            return fmt.Errorf("conversion pattern references undefined unit: %s",
                sourceUnitRef.Name)
        }

        // Validate: source and target must belong to the same concept
        if sourceUnit.ConceptName != targetUnit.ConceptName {
            return fmt.Errorf("conversion between units of different concepts: %s (%s) ↔ %s (%s)",
                sourceUnit.Name, sourceUnit.ConceptName,
                targetUnit.Name, targetUnit.ConceptName)
        }

        // Build calculation expression tree from CalculationScope definitions
        calcExpr, err := a.BuildExpressionTree(calculationScope.Definitions)
        if err != nil {
            return fmt.Errorf("conversion calculation: %w", err)
        }

        // Validate: all numeric constants in calculation must carry explicit unit annotations
        if err := validateUnitAnnotations(calcExpr, a.Units); err != nil {
            return fmt.Errorf("conversion calculation: %w", err)
        }

        // Assemble forward conversion pair
        forwardPair := AnalysisConversionPair{
            SourceUnit:  sourceUnit,
            TargetUnit:  targetUnit,
            SourceVar:   sourceVarName,
            TargetVar:   targetVarName,
            Calculation: calcExpr,
        }

        // Derive inverse conversion by swapping source/target
        // The synthesizer sees both directions as explicit CIL pairs
        inversePair := AnalysisConversionPair{
            SourceUnit:  targetUnit,
            TargetUnit:  sourceUnit,
            SourceVar:   targetVarName,
            TargetVar:   sourceVarName,
            Calculation: invertCalculation(calcExpr),
        }

        unitDef.Conversions = append(unitDef.Conversions, forwardPair, inversePair)
    }

    return nil
}
```

**Validation During Analysis**

- Pattern units must exist in the Units map
- Source and target units must belong to the same concept (same unit group)
- Pattern-bound variable names must not collide with unit names (error)
- All numeric constants in calculation expressions must carry explicit unit annotations (bare `32` is an error; `32*F` is required)
- Unparenthesized identifiers in calculation expressions are unit references — must exist in Units map
- Calculation expression must be structurally valid (BuildExpressionTree succeeds)
- Conversion patterns must not create cycles (no self-conversion `Meter(m) = Meter(m)`)

**Unit Annotation Rule**

Every numeric constant in a conversion calculation must carry explicit unit annotation. For example:
- `32*F` — correct: "the number 32 with units Fahrenheit"
- `(5/9)*(C/F)` — correct: "the fraction 5/9 with units Celsius per Fahrenheit"
- `32` alone — error: missing unit annotation
- Unparenthesized identifier → unit reference. Parenthesized `(var)` → pattern-bound variable. Name collision between a unit and a pattern variable is an error.

**Equation-Form Patterns Only**

Patterns use equation form: `UnitName(var1) = UnitName(var2)`. Ratio form (e.g., `Kilometer(km) / Meter(m)`) is not supported. Equation form makes pattern variable binding unambiguous and calculation expression construction straightforward.

### Increment 5: Measure Conversions

**Analysis-Time Measure Conversion Structures**

Measure conversions follow the same pattern as unit conversions but reference measures instead of units. A measure conversion is equivalent to a unit conversion on the measure's unit.

```go
// AnalysisMeasureConversionPair captures a measure-level conversion in resolved form.
// Same structure as AnalysisConversionPair but references measures.
type AnalysisMeasureConversionPair struct {
    SourceMeasure *AnalysisMeasureDefinition  // Source measure (e.g., Hour)
    TargetMeasure *AnalysisMeasureDefinition  // Target measure (e.g., Second)
    SourceVar     string                      // Pattern-bound source variable (e.g., "h")
    TargetVar     string                      // Pattern-bound target variable (e.g., "s")
    Calculation   Expression                  // Resolved calculation expression
}

// Enhanced AnalysisMeasureDefinition with conversions.
type AnalysisMeasureDefinition struct {
    Name            string
    TypeName        string
    UnitExpression  Expression
    Conversions     []AnalysisMeasureConversionPair  // NEW: Measure conversions
    SourceLocation  SourceLocation
}
```

**Measure Conversion Analysis — Symbol Table Walking**

The analysis phase walks MeasureScope children for MeasureConversionScope entries. The pattern is identical to unit conversions but operates on measure references.

```go
// AnalyzeMeasureConversions walks MeasureScope children for
// MeasureConversionScope entries, pairs PatternScope with
// CalculationScope, extracts pattern-bound variables, builds
// calculation expression trees, and derives inverse conversions.
func (a *CPLAnalyzer) AnalyzeMeasureConversions(
    measureScope *SymbolTable,
    measureDef *AnalysisMeasureDefinition) error {

    for _, child := range measureScope.Children {
        if child.ScopeType != MeasureConversionScope {
            continue
        }

        // Pair PatternScope with CalculationScope by order
        var patternScope, calculationScope *SymbolTable
        for _, convChild := range child.Children {
            switch convChild.ScopeType {
            case PatternScope:
                patternScope = convChild
            case CalculationScope:
                calculationScope = convChild
            }
        }

        if patternScope == nil || calculationScope == nil {
            return fmt.Errorf("measure conversion clause missing pattern or calculation")
        }

        // Extract pattern-bound variables: first measure+variable = target, second = source
        targetRef, targetVarName, err := extractBoundMeasureRef(patternScope, 0)
        if err != nil {
            return fmt.Errorf("measure conversion pattern: %w", err)
        }
        sourceRef, sourceVarName, err := extractBoundMeasureRef(patternScope, 1)
        if err != nil {
            return fmt.Errorf("measure conversion pattern: %w", err)
        }

        // Resolve measure references
        targetMeasure, exists := a.Measures[targetRef.Name]
        if !exists {
            return fmt.Errorf("measure conversion references undefined target measure: %s",
                targetRef.Name)
        }
        sourceMeasure, exists := a.Measures[sourceRef.Name]
        if !exists {
            return fmt.Errorf("measure conversion references undefined source measure: %s",
                sourceRef.Name)
        }

        // Build calculation expression tree from CalculationScope definitions
        calcExpr, err := a.BuildExpressionTree(calculationScope.Definitions)
        if err != nil {
            return fmt.Errorf("measure conversion calculation: %w", err)
        }

        // Validate unit annotations
        if err := validateUnitAnnotations(calcExpr, a.Units); err != nil {
            return fmt.Errorf("measure conversion calculation: %w", err)
        }

        // Assemble forward pair
        forwardPair := AnalysisMeasureConversionPair{
            SourceMeasure: sourceMeasure,
            TargetMeasure: targetMeasure,
            SourceVar:     sourceVarName,
            TargetVar:     targetVarName,
            Calculation:   calcExpr,
        }

        // Derive inverse conversion (swap source/target)
        inversePair := AnalysisMeasureConversionPair{
            SourceMeasure: targetMeasure,
            TargetMeasure: sourceMeasure,
            SourceVar:     targetVarName,
            TargetVar:     sourceVarName,
            Calculation:   invertCalculation(calcExpr),
        }

        measureDef.Conversions = append(measureDef.Conversions, forwardPair, inversePair)
    }

    return nil
}
```

**Validation During Analysis**

- Target and source measures must exist in the Measures map
- Pattern-bound variable names must not collide with measure names
- All numeric constants in calculation expressions must carry explicit unit annotations
- Unparenthesized identifiers in calculation expressions are unit references
- Calculation expressions must be structurally valid
- Conversion patterns must not create cycles

**Simple Conversions Only — No Conditional Branches**

This increment supports only simple (non-conditional) conversions with a single calculation expression. Context-dependent conversions (e.g., date-specific hour counts for daylight savings time) using `when` clauses are a future extension and explicitly out of scope.

## Cross-Phase Transformation Summary

### ParserMeasureDefinition → AnalysisMeasureDefinition

| Increment | Parser Structure               | Analysis Structure                  | Transformation                 |
| --------- | ------------------------------ | ----------------------------------- | ------------------------------ |
| 1         | N/A                            | N/A                                 | No measures in Increment 1     |
| 2         | `UnitGroupName string`         | `UnitGroup *AnalysisUnitDefinition` | Resolve name to unit           |
| 3         | `[]SymbolDefinition` (postfix) | `UnitExpression Expression`         | Evaluate postfix to build tree |
| 4         | `[]SymbolDefinition` (postfix) | `UnitExpression Expression`         | Same as Increment 3            |
| 5         | `[]SymbolDefinition` (postfix) | `UnitExpression Expression`         | Same as Increment 3            |

### ParserUnitDefinition → AnalysisUnitDefinition

| Increment | Parser Structure              | Analysis Structure                   | Transformation                    |
| --------- | ----------------------------- | ------------------------------------ | --------------------------------- |
| 1         | `ConceptName string`          | `Concept *AnalysisConceptDefinition` | Resolve name to concept           |
| 2         | Same as 1                     | Same as 1                            | No change                         |
| 3         | Same as 1                     | Same as 1                            | No change                         |
| 4         | ConversionScope children      | `[]AnalysisConversionPair`           | Walk ConversionScope → PatternScope + CalculationScope; derive inverse |
| 5         | MeasureConversionScope children | `[]AnalysisMeasureConversionPair`  | Walk MeasureConversionScope → PatternScope + CalculationScope; derive inverse |

## Approach

### Incremental Structure and Function Evolution

The analysis phase uses an incremental approach where structures and functions evolve by adding fields and capabilities as each increment requires them:

**Structure Evolution:**

- **Increment 1:** `AnalysisUnitDefinition` with `Name`, `ConceptName`, `Concept`
- **Increment 2:** `AnalysisMeasureDefinition` with `Name`, `TypeName`, `UnitGroupName`, `UnitGroup`
- **Increment 3:** Enhanced `AnalysisMeasureDefinition` with `UnitExpression Expression` (replaces `UnitGroup`), Expression tree types defined
- **Increment 4:** Enhanced `AnalysisUnitDefinition` with `Conversions []AnalysisConversionPair` (SourceUnit, TargetUnit, SourceVar, TargetVar, Calculation Expression). Inverse conversions derived at analysis time.
- **Increment 5:** Enhanced `AnalysisMeasureDefinition` with `Conversions []AnalysisMeasureConversionPair` (SourceMeasure, TargetMeasure, SourceVar, TargetVar, Calculation Expression). Inverse conversions derived at analysis time.

**Function Evolution:**

- **Increment 1:** `AnalyzeUnitsScope` processes unit definitions
- **Increment 2:** `AnalyzeMeasuresScope` processes simple measure definitions
- **Increment 3:** Enhanced `AnalyzeMeasuresScope` with `BuildExpressionTree` for compound units
- **Increment 4:** `AnalyzeUnitConversions` walks ConversionScope children of UnitScope, pairs PatternScope with CalculationScope, extracts pattern-bound variables, builds calculation trees, derives inverses
- **Increment 5:** `AnalyzeMeasureConversions` walks MeasureConversionScope children of MeasureScope, same pattern applied to measure-level conversions

**No Feature Flags:**

There are no feature flags or runtime checks for increment capabilities. Each increment's code uses only the fields and functions available at that point in development.

## Decision

**Chosen Approach:** Strict separation between parser-time raw capture and analysis-time semantic resolution

**Rationale:**
This separation maintains clean architectural boundaries:

1. **Parser Responsibility**: Capture exactly what was written in source code (names, operators, structure)
2. **Analysis Responsibility**: Resolve names to definitions, build semantic trees, validate references

The Expression tree represents a semantic understanding of unit algebra that requires:

- Knowledge of the Units lookup map (built during analysis)
- Resolution of unit names to their definitions
- Building a tree structure for dimensional analysis

These are fundamentally analysis-time activities, not parsing-time activities. The parser produces `RawUnitExpression` which the analysis phase transforms into `Expression`.

## Implications

### Positive Implications

- Clear separation of concerns between parsing and analysis
- Parser remains focused on syntax, analysis handles semantics
- Analysis structures can reference other analysis structures via pointers
- Unit expressions can be validated against the complete Units map
- Compound unit expressions can reference any unit in the compilation unit
- Expression trees can be used for dimensional analysis

### Negative Implications / Trade-offs

- Slightly more complex data model (two Expression types)
- Transformation step required during analysis
- Raw unit expressions cannot be validated during parsing
- Parser cannot detect undefined unit references

### Validation Strategy

Each increment adds validation during the analysis phase:

1. **Increment 1:** Validate unit names are unique, concept references exist
2. **Increment 2:** Validate measure names are unique, unit group references exist, type is supported
3. **Increment 3:** Validate all unit names in compound expressions exist, build resolved Expression tree
4. **Increment 4:** Validate conversion pattern units exist and belong to the same concept, pattern-bound variable names don't collide with unit names, all numeric constants carry explicit unit annotations, calculation expressions valid, derive inverse conversions
5. **Increment 5:** Validate measure conversion references exist, same validation rules as Increment 4 applied to measure-level conversions. Simple (non-conditional) conversions only — context-dependent conversions with `when` clauses are a future extension.

## Notes

**Related Documents:**

- [Units and Measures Incremental Implementation](../../eros/ERO003_units-measures-incremental-implementation.md) - Main incremental implementation ERO
- [Units and Measures Parsing Phase](../go-parser-target/eros/ERO003_units-measures-parsing-phase.md) - Parser data structures (must be updated to use RawUnitExpression)
- [Units and Measures Design](../../units-and-measures.md) - Full units and measures design
- [ERO001: Concepts Analysis Phase](ERO001_concepts-analysis-phase.md) - Pattern for analysis phase EROs
- [Symbol Table Design](../../go-parser-target/symbol-table-design.md) - Symbol table integration

**Key Structural Changes to Parsing Phase:**

The parsing phase ERO must use postfix expression components stored in the scope's Definitions slice rather than building expression trees:

```go
// Parser pushes components to scope definitions in postfix order
type ParserUnitReference struct {
    Name string
}

type ParserOperator struct {
    Operator OperatorType  // OpMultiply, OpDivide, OpPower
}

// ParserMeasureDefinition does NOT contain expression data
type ParserMeasureDefinition struct {
    Name            string
    TypeName        string
    SourceLocation  SourceLocation
    // Expression components are in the MeasureScope's Definitions slice
}
```

**Dependencies:**

- Analysis phase requires Concepts analysis to be complete (ERO001)
- Unit conversions (Increment 4) depend on metrics formula evaluation
- Measure conversions (Increment 5) depend on metrics formula evaluation

**Future Tensor Extension:**

The Expression tree is designed to be extensible for future tensor support:

```go
// Future extension for metric tensors
type TensorNode struct {
    TensorName string
    Components []Expression
}

func (t TensorNode) isExpression() {}
func (t TensorNode) String() string { ... }
```

This allows mixed unit-tensor expressions like `MetricTensor * Meter/Second^2` in future increments.
