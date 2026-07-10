# ERO003: Units and Measures Incremental Implementation

## Title

Units and Measures Incremental Implementation

## Purpose/Goal

Define the incremental approach for implementing units and measures functionality across all CPL phases (parsing, analysis, synthesis, evaluation), detailing which subset of features is implemented in each increment and documenting the limitations of each intermediate state.

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic003: Units and Measures Implementation](../../../epics/Epic003_units-and-measures-implementation.md) - Epic tracking the implementation of units and measures

**Related Stories:**

- [Story153: CPL Units Parser - Basic Unit Definitions Reimplementation](../../../stories/Story153_cpl-units-parser-basic-reimplementation.md) - Increment 1 parser implementation (replaces superseded Story141)
- ~~Story141~~ through ~~Story150~~ - Superseded stories (old implementation incompatible with types system)

## Context

### Situation

The units-and-measures.md document outlines comprehensive functionality for units and measures definitions in CPL, including:

- Unit sections with unit group declarations based on concept values
- Unit conversions via pattern/calculation clauses
- Compound units using `*`, `/`, and `^` operators
- Measure definitions combining types with unit groups
- Measure conversions for context-dependent transformations
- Built-in `Unitless` unit for dimensionless quantities
- Conversion precedence rules and path resolution

The implementation of units and measures is somewhat coupled to the implementation of types, particularly the `Number` type which is the foundational type for measures. Additionally, unit conversions and measure conversions require formula evaluation capabilities that will be implemented as part of the metrics feature.

### Constraints

- Each increment must produce a working system (though with limitations)
- Later increments must not break functionality from earlier increments
- The incremental evolution must maintain consistent data structures across phases
- Units and measures definitions must remain statically known at compile/analysis/synthesis time
- The approach must align with the existing CPL phase architecture (parse → analyze → synthesize)
- Unit implementation depends on concept name and value availability (concepts define unit values)
- Measure implementation depends on type implementation (measures combine types with units)
- Unit/measure conversions depend on formula evaluation (to be implemented with metrics)

### Requirements

- Support basic unit group definitions in early increments
- Enable progressive enhancement of unit and measure capabilities
- Document limitations at each increment clearly
- Ensure each increment provides value for some use cases
- Maintain backward compatibility within the incremental progression
- Support unit algebra (\*, /, ^ operations) for compound units
- Handle conversion precedence (specific over general)

### Assumptions

- Concept names and values will be available (units depend on concepts for their values)
- Note: Complex concept features (hierarchy, conditionals, ranges, before/after) are not required for unit declarations
- The `Number` type will be the primary type used with measures initially
- Type derivation from `Number` will be available for creating specialized measure types
- Formula evaluation infrastructure will be implemented as part of metrics
- Unit conversions and measure conversions can be deferred until after formula evaluation

## Migration from Existing Implementation

### Overview

This section documents the architectural shift from the old unit implementation to the new ERO003 design. The change represents a fundamental rethinking of how unit declarations flow through the CPL compiler, moving from a parser-centric AST-building approach to a symbol table-driven parsing approach with clear phase separation.

### What Existed Before (Old Approach)

The original unit implementation followed a pattern where the parser was responsible for building AST nodes directly during parse-time:

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
4. **Poor Separation**: Mixed syntax parsing with semantic AST building

### What the New ERO003 Design Provides

The ERO003 implementation follows the [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md) with clear phase separation:

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

### Why the Change Was Necessary

The migration from the old `UnitDecl` AST approach to the symbol table-driven approach addresses several architectural concerns:

**1. Phase Separation Principle Violation**

The old approach violated the parsing/analysis phase separation principle that underlies the entire CPL architecture. The parser was performing semantic work (building AST nodes) that properly belongs to the analysis phase. This mixing of concerns made the system harder to maintain and extend.

**2. Consistency with Other CPL Features**

The ERO001 concepts implementation already established the symbol table-driven pattern. The old unit implementation was an outlier that did not follow this established convention. Unifying units with concepts under the same architectural pattern improves consistency across the codebase.

**3. Support for Incremental Implementation**

The symbol table-driven approach better supports the incremental implementation strategy defined in this ERO. Each increment can add new scope types and definition types without requiring changes to existing listener code. This modularity would be difficult to achieve with the old AST-building approach.

**4. Postfix Expression Parsing**

The new design enables the [Postfix Expression Parsing Pattern](../../../../patterns/postfix-expression-parsing.md) for compound units. This pattern requires capturing expression components in postfix order during parsing and building the expression tree during analysis. The old approach would have required building the tree during parsing, which violates phase separation.

### Impact on Existing Code

**Code That May Reference Old Structures:**

Any code that references the old `UnitDecl` AST nodes will need to be updated to work with the new symbol table-driven structures:

1. **Parser Listeners**: Listener methods that previously built `UnitDecl` nodes must be rewritten to populate symbol tables with `ParserUnitDefinition` structures
2. **Analysis Phase**: New analyzer code must be written to walk the symbol table tree and build `AnalysisUnitDefinition` structures
3. **Synthesis Phase**: Code generation will work with `AnalysisUnitDefinition` structures rather than `UnitDecl` AST nodes

**Migration Path:**

1. **Do NOT build AST nodes directly** in listener methods
2. **Create `ParserUnitDefinition`** with raw parsed data (names as strings, not resolved references)
3. **Attach to `SymbolDefinition.Data`** when defining symbols in the symbol table
4. **Let the analysis phase** resolve references and build `AnalysisUnitDefinition`

**Phase Separation Summary:**

| Phase             | Old Approach               | New Approach                                     |
| ----------------- | -------------------------- | ------------------------------------------------ |
| **Parser-Time**   | Build `UnitDecl` AST nodes | Push `ParserUnitDefinition` to symbol table      |
| **Analysis-Time** | (same AST nodes)           | Build `AnalysisUnitDefinition` from symbol table |

The phase-specific EROs document the complete transformation:

- **[Parsing Phase ERO](../go-parser-target/eros/ERO003_units-measures-parsing-phase.md)**: Defines `ParserUnitDefinition`, `ParserMeasureDefinition`, and postfix expression capture
- **[Analysis Phase ERO](../go-analysis/eros/ERO003_units-measures-analysis-phase.md)**: Defines `AnalysisUnitDefinition`, `AnalysisMeasureDefinition`, and Expression tree building

## Evaluation Criteria

1. **Risk Reduction** (High importance): Each increment should reduce implementation risk by validating assumptions
2. **Deliverable Value** (High importance): Each increment should provide usable functionality
3. **Complexity Management** (Medium importance): Increments should avoid overwhelming complexity
4. **Consistency** (High importance): Data structures should evolve consistently across all phases
5. **Dependency Management** (High importance): Increments should respect dependencies on types and formula evaluation

## Options Considered

### Option 1: Big Bang Implementation

**Description:**
Implement all units and measures functionality in a single large effort.

**Pros:**

- Complete functionality available at once
- No intermediate limitations to document or work around
- Single design effort

**Cons:**

- High risk of bugs and design flaws
- Delayed feedback on core functionality
- Complex interactions between units, measures, types, and conversions
- All-or-nothing delivery
- Cannot validate unit algebra independently of conversions

### Option 2: Incremental Implementation by Feature

**Description:**
Implement units and measures features incrementally, adding capabilities across all phases in each increment.

**Pros:**

- Early validation of core unit and measure concepts
- Reduced risk through iterative delivery
- Opportunity for feedback between increments
- Ability to course-correct based on learnings
- Can defer conversion complexity until formula evaluation is ready

**Cons:**

- Requires careful planning of increment boundaries
- Intermediate limitations must be clearly documented
- Some architectural decisions must anticipate future features (conversions)

### Option 3: Type-First Implementation

**Description:**
Implement all types first, then units, then measures, then conversions.

**Pros:**

- Clear separation of concerns
- Complete types available before measures need them

**Cons:**

- No end-to-end units/measures functionality until late
- Risk of discovering integration issues late
- Less opportunity for holistic feedback

## Decision

**Chosen Option:** Option 2 - Incremental Implementation by Feature

**Rationale:**
The incremental-by-feature approach provides the best balance of risk reduction and early value delivery. By implementing a subset of features across all phases in each increment, we ensure end-to-end functionality is available early, enabling validation and feedback. This approach respects the coupling between units and measures while allowing basic functionality to be used before advanced features like conversions are implemented.

The specific increments are defined as follows:

### Increment 1: Basic Unit Definitions

**Unit Section Form:**

A unit section starts with the keyword `units` followed by a block of unit definitions. Each unit declaration has the form:

```cpl
units
    <unit group>=<concept name>
```

Example:

```cpl
concepts
    Currency = [Values[USD, EURO, YEN]]
    Length = [Values[Meter, Kilometer, Centimeter]]

units
    Currency = Currency
    Length = Length
```

**Features Implemented:**

- Unit group declarations referencing concept names
- Single units section per file
- Built-in `Unitless` unit support
- Basic unit reference validation

**Limitations:**

- No conversion clauses (pattern/calculation)
- No compound units using operators
- Single units section per file
- No measure definitions yet
- Requires only basic concept name and value availability (complex concept relationships not required)

**Value Provided:**

Basic unit groups can be defined and validated. Unit groups serve as namespaces for units derived from concept values.

**Dependencies:**

- Requires concept name and value availability - concepts must be resolvable to their values
- Note: Complex concept features (hierarchy, conditionals, ranges, before/after) are not required for basic unit declarations

### Increment 2: Basic Measure Definitions

**Measure Section Form:**

A measure section starts with the keyword `measures` followed by a block of measure definitions. Each measure declaration has the form:

```cpl
measures
    <measure group> = <type name>(unit=<unit group name>)
```

Example:

```cpl
concepts
    TemperatureScale = [Values[C, F]]

units
    TemperatureScale = TemperatureScale

measures
    Temperature = Number(unit=TemperatureScale)
```

**Features Implemented:**

- Measure group declarations combining types with unit groups
- Type reference with unit specification
- Single measures section per file
- Basic measure reference validation

**Limitations:**

- Only `Number` type and types derived from `Number` supported
- No measure conversions
- No compound unit expressions in measure declarations
- Single measures section per file
- No validation of unit group existence (assumes units already validated)

**Value Provided:**

Measures can be defined combining the `Number` type (or `Number`-derived types) with unit groups. Each unit in the unit group creates a distinct measure.

**Dependencies:**

- Requires Increment 1 (basic units) to be complete
- Requires `Number` type implementation
- Requires ability to derive types from `Number`

### Increment 3: Compound Units and Unit Algebra Tree

**Compound Unit Form:**

Units can be combined using `*`, `/`, and `^` operators to form compound units:

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
```

**Features Implemented:**

- Tree representation for unit combinations
- Support for `*` (multiply), `/` (divide), and `^` (power) operators
- Compound unit expressions in measure declarations
- General tree structure where leaf elements can be units or tensors
- Unit algebra validation

**Tree Representation Design:**

The unit algebra tree is designed to be general enough to support:

1. **Current Use Case:** Units as leaf nodes
   - Simple units (Meter, Second)
   - Compound units (Meter/Second, Kilogram\*Meter/Second^2)

2. **Future Extension:** Tensors as leaf nodes
   - Metric tensors with unit components
   - Mixed unit-tensor expressions

The tree structure supports:

- Binary operations: MultiplicationNode, DivisionNode
- Unary operations: PowerNode (with integer exponent)
- Leaf nodes: UnitNode, TensorNode (placeholder for future)

Example tree for `Length/Time^2`:

```
DivisionNode
├── left: UnitNode(Length)
└── right: PowerNode
    ├── base: UnitNode(Time)
    └── exponent: 2
```

**Limitations:**

- No unit conversions (cannot simplify compound units)
- No derived unit definitions (e.g., `Force = Mass * Length / Time^2`)
- Unit trees are preserved symbolically without simplification
- No validation that compound units make business sense

**Value Provided:**

Measures can use compound units for dimensional correctness. Unit combinations are tracked symbolically through the tree structure.

**Dependencies:**

- Requires Increment 2 (basic measures) to be complete
- Tree structure must support future tensor extension for metrics

### Increment 4: Unit Conversions

**Unit Conversion Form:**

Unit declarations may include optional conversion clauses:

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

**Features Implemented:**

- Conversion clauses with pattern/calculation pairs
- Unit aliases in patterns (e.g., `Kilometer(km)`)
- Unit expressions in calculations (e.g., `C/F`)
- Value aliases in calculations (e.g., `km`, `m`)
- Conversion precedence (specific patterns over general patterns)
- Multi-step conversion path resolution

**Limitations:**

- Only unit-level conversions (measure conversions in Increment 5)
- Formula evaluation must support arithmetic operations used in calculations
- No unit group patterns yet (specific units only)

**Value Provided:**

Units can define conversion relationships. CPL can automatically convert between compatible units using defined conversion paths.

**Dependencies:**

- Requires Increment 3 (compound units) to be complete
- **Requires formula evaluation from metrics implementation** (arithmetic, boolean operations)
- Conversion calculations use the same expression evaluation as metrics

### Increment 5: Measure Conversions

**Measure Conversion Form:**

Measure declarations may include optional conversion clauses:

```cpl
concepts
    Duration = [Values[Second, Minute, Hour, Day]]

units
    Duration = Duration

types
    Midnight = Instant(hour=0, minute=0, second=0)
    DayInterval = Interval(starts_at = Midnight, unit = Duration)

measures
    Day = DayInterval(unit = Duration)
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

**Features Implemented:**

- Measure-level conversion clauses
- Context-dependent conversions (e.g., daylight savings time)
- Conditional logic in conversion calculations (when clauses)
- Measure reference patterns (e.g., `Hour=Day(other_day)`)
- Conversion precedence hierarchy (measure conversions over unit conversions)

**Limitations:**

- Complex time-related conversions may require additional time-handling features
- Integration with external conversion data sources not included

**Value Provided:**

Measures can define context-dependent conversions that go beyond simple unit conversions. This enables business-specific transformations.

**Dependencies:**

- Requires Increment 4 (unit conversions) to be complete
- **Requires formula evaluation from metrics implementation** (conditional expressions, when clauses)
- Measure conversions use the same expression evaluation as metrics

## Summary of Increments

| Increment | Feature             | Dependencies                            | Key Capability                   |
| --------- | ------------------- | --------------------------------------- | -------------------------------- |
| 1         | Basic Units         | Concepts                                | Define unit groups from concepts |
| 2         | Basic Measures      | Increment 1, Number type                | Define measures (Number + unit)  |
| 3         | Compound Units      | Increment 2                             | Unit algebra tree (\*, /, ^)     |
| 4         | Unit Conversions    | Increment 3, Metrics formula evaluation | Convert between units            |
| 5         | Measure Conversions | Increment 4, Metrics formula evaluation | Context-dependent conversions    |

## Implications

### Positive Implications

- Reduced implementation risk through iterative delivery
- Early availability of core units and measures functionality
- Opportunity for feedback and course correction
- Clear milestones and progress tracking
- Each increment provides usable value
- Conversion complexity isolated until formula evaluation is ready

### Negative Implications / Trade-offs

- Requires discipline to maintain backward compatibility between increments
- Some architectural decisions must anticipate future features (tree structure for tensors)
- Intermediate limitations may constrain early use cases (no conversions until Increments 4-5)
- Documentation must clearly track which features are available in each increment
- Testing must cover both current increment features and ensure no regression of prior increments
- Dependency on metrics formula evaluation for conversions creates coupling

### Impact on System Qualities

| Quality Attribute | Impact | Notes                                               |
| ----------------- | ------ | --------------------------------------------------- |
| Performance       | =      | No significant impact expected                      |
| Scalability       | =      | Incremental approach doesn't affect scalability     |
| Maintainability   | +      | Smaller increments easier to understand and modify  |
| Security          | =      | No security implications                            |
| Usability         | +      | Early availability of features improves usability   |
| Testability       | +      | Each increment can be thoroughly tested before next |

### Future Considerations

- Each increment should be considered complete before starting the next
- EROs for each phase should document the phase-specific data structures for each increment
- This ERO should be updated as increments are completed
- If significant learnings emerge during implementation, the increment plan may be adjusted
- The tree structure for unit algebra should be reviewed when metric tensor implementation begins
- Consider creating epic(s) for units and measures implementation to track related stories

## Notes

**Related Documents:**

- [Units and Measures Design](../units-and-measures.md) - Comprehensive design for units and measures functionality
- [ERO001: Concepts Incremental Implementation](ERO001_concepts-incremental-implementation.md) - Concepts implementation pattern (units depend on concepts)
- [CPL Architecture - Introduction](../../architecture/cpl/introduction.md) - Overall CPL processing lifecycle
- [CPL Requirements - Concepts, Types, Units, Measures](../../requirements/cpl/concepts-types-units-measures-syntax.md) - Syntax requirements

**Pattern References:**

- [Symbol Table-Driven Parsing Pattern](../../../../patterns/symbol-table-driven-parsing.md) - Architectural pattern for parsing phase design
- [Postfix Expression Parsing Pattern](../../../../patterns/postfix-expression-parsing.md) - Pattern for compound unit expression parsing

**Phase-Specific EROs:**

Phase-specific EROs document the data structures for each CPL phase:

1. **[Parsing Phase ERO](../go-parser-target/eros/ERO003_units-measures-parsing-phase.md)**: Parser data structures for units and measures - Defines `ParserUnitDefinition`, `ParserMeasureDefinition`, and `RawUnitExpression` structures
2. **[Analysis Phase ERO](../go-analysis/eros/ERO003_units-measures-analysis-phase.md)**: Analysis structures for units and measures - Defines `AnalysisUnitDefinition`, `AnalysisMeasureDefinition`, and `Expression` tree structures (transforms parser → analysis)
3. **Synthesis Phase ERO**: Synthesis structures for units and measures (ERO003 synthesis phase) - To be created

**Phase Separation Note:**

The parsing phase uses `ParserUnitReference` and `ParserOperator` structures pushed to the symbol table's Definitions slice in postfix order (following the [Postfix Expression Parsing Pattern](../../../../patterns/postfix-expression-parsing.md)). The analysis phase evaluates this postfix sequence to build `Expression` trees with resolved references to `AnalysisUnitDefinition` records. This separation keeps the parser focused on syntax while the analysis phase handles semantic resolution.

**Dependencies on Other Features:**

- **Concepts**: Unit values come from concept values. Concepts must be implemented before units.
- **Types (Number)**: Measures combine types with units. The Number type must be implemented before measures.
- **Metrics/Formula Evaluation**: Conversions require expression evaluation. Formula evaluation from metrics must be implemented before conversions.

**Coupling Note:**

The implementation of units and measures is somewhat coupled to the implementation of types, particularly for measures which combine types with units. The incremental breakdown respects this coupling by:

1. Implementing basic units first (depends only on concepts)
2. Implementing basic measures second (depends on units and Number type)
3. Implementing compound units third (extends measures with unit algebra)
4. Deferring conversions until formula evaluation is available (from metrics)
