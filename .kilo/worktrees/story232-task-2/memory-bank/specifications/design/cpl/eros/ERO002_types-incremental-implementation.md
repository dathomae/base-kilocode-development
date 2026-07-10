# ERO002: Types Incremental Implementation

## Title

Types Incremental Implementation

## Purpose/Goal

Define the incremental approach for implementing types functionality across all CPL phases (parsing, analysis, synthesis, evaluation), detailing which subset of types is implemented in each increment and documenting the limitations of each intermediate state.

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic001_cpl-concepts-implementation.md](../../../../epics/Epic001_cpl-concepts-implementation.md) - Epic for CPL concepts implementation

## Context

### Situation

The types-definitions.md document outlines comprehensive functionality for type definitions in CPL, including:

- Type sections with indentation-based scoping
- Base types: Number, String, Instant, Interval, Boolean
- Type derivation with constraints (e.g., `Currency = Number(precision=2)`)
- Constraint inheritance and validation (derived types must be at least as restrictive as base types)
- Built-in types always available without declaration

The five built-in types are:

1. **Number** - Unrestricted real number with precision, min, max constraints
2. **String** - Unicode sequence with min/max length and regex pattern constraints
3. **Instant** - UTC moment constraints (year, month, day, hour, minute, second, day_of_week, week_number). Omitted fields mean underspecification.
4. **Interval** - Time span with `starts_at` (named `Instant`-derived type or inline `Instant` constraints) and unit reference
5. **Boolean** - Logical true/false value

Note: `Duration` is not a type. It is a concept or value set whose values define units used when measuring an `Interval`.

Implementing all type functionality at once would be complex due to the interdependencies between types, units, and measures—particularly for time-related types (Instant and Interval). An incremental approach allows for validated learning and progressive feature delivery.

### Constraints

- Each increment must produce a working system (though with limitations)
- Later increments must not break functionality from earlier increments
- The incremental evolution must maintain consistent data structures across phases
- Type definitions must remain statically known at compile/analysis/synthesis time
- The approach must align with the existing CPL phase architecture (parse → analyze → synthesize)
- Derived types must be at least as restrictive as their base types

### Requirements

- Support basic type definition and usage in early increments
- Enable progressive enhancement of type capabilities
- Document limitations at each increment clearly
- Ensure each increment provides value for some use cases
- Maintain backward compatibility within the incremental progression
- Keep type constraints synchronized across all phases

### Assumptions

- Type definitions will be primarily static within a compilation unit
- The most common use cases involve Number and String types
- Instant and Interval types require additional unit/measure machinery
- Boolean is a simple type with no constraints
- Type conversion rules (more restrictive → less restrictive only) can be enforced in analysis

## Evaluation Criteria

1. **Risk Reduction** (High importance): Each increment should reduce implementation risk by validating assumptions
2. **Deliverable Value** (High importance): Each increment should provide usable functionality
3. **Complexity Management** (Medium importance): Increments should avoid overwhelming complexity
4. **Consistency** (High importance): Data structures should evolve consistently across all phases

## Options Considered

### Option 1: Big Bang Implementation

**Description:**
Implement all type functionality in a single large effort.

**Pros:**

- Complete functionality available at once
- No intermediate limitations to document or work around
- Single design effort

**Cons:**

- High risk of bugs and design flaws
- Delayed feedback on core functionality
- Difficult to estimate and plan
- All-or-nothing delivery
- Time-related types (Instant, Interval) require complex unit/measure integration

### Option 2: Incremental Implementation by Type Category

**Description:**
Implement types incrementally by category: first basic scalar types (String, Number, Boolean), then time-related types (Instant, Interval) once unit/measure machinery is in place.

**Pros:**

- Early validation of core type system
- Reduced risk through iterative delivery
- Opportunity for feedback between increments
- Ability to course-correct based on learnings
- Time-related types can leverage established unit/measure infrastructure

**Cons:**

- Requires careful planning of increment boundaries
- Intermediate limitations must be clearly documented
- Some architectural decisions must anticipate time-related types

### Option 3: Incremental Implementation by Phase

**Description:**
Implement complete type functionality in one phase at a time (e.g., complete parsing first, then analysis, then synthesis).

**Pros:**

- Completes each phase before moving to next
- Clear separation of concerns

**Cons:**

- No end-to-end functionality until final phase complete
- Risk of discovering phase integration issues late
- Less opportunity for holistic feedback

## Decision

**Chosen Option:** Option 2 - Incremental Implementation by Type Category

**Rationale:**
The incremental-by-category approach provides the best balance of risk reduction and early value delivery. By implementing basic scalar types (String, Number, Boolean) first across all phases, we ensure end-to-end functionality is available early. Time-related types (Instant, Interval) require additional machinery for units and measures that will be developed in parallel, making them natural candidates for a second increment.

The specific increments are defined as follows:

### Increment 1: Number Type with Type Capture

**Type Section Form:**

A types section starts with the keyword `types` followed by a block of type definitions. Each type declaration has the form:

```cpl
types
    <type name> = <base type>(<constraints>)
```

Example:

```cpl
types
    Currency = Number(precision=2)
    BrokerageCurrency = Number(precision=4)
```

**Features Implemented:**

- **Number type** with constraints:
  - `precision`: Number of decimal places (≥ 0)
  - `minimum`: Minimum allowed value
  - `maximum`: Maximum allowed value (must be > minimum)
- Type derivation chain validation (derived must be at least as restrictive as base)
- Single type section per file
- Built-in Number type available without declaration
- **Type capture**: Types not yet defined are captured for later extension

**Limitations:**

- No String type support
- No Boolean type support
- No Instant type support
- No Interval type support
- No multiple type sections in the same file
- No type references from measures/units (requires measure implementation)

**Value Provided:**

Numeric types with precision constraints support financial calculations. Type capture allows forward references and progressive type definition.

### Increment 2: String and Boolean Types

**Features Added:**

- **String type** with constraints:
  - `minimum`: Minimum length in Unicode code points
  - `maximum`: Maximum length in Unicode code points (must be > minimum)
  - `pattern`: Regular expression pattern (optional)
- **Boolean type**: No constraints, simple true/false values

Example:

```cpl
types
    Currency = Number(precision=2)
    BrokerageCurrency = Number(precision=4)
    ShortString = String(maximum=50)
    CurrencyCode = String(pattern="[A-Z]{3}")
    ActiveFlag = Boolean
```

**Limitations:**

- No Instant type support
- No Interval type support
- No multiple type sections in the same file

**Value Provided:**

Basic scalar types can be defined and used in metrics. String types with length and pattern constraints support validation of identifiers and codes. Boolean enables flag and state representations.

### Increment 3: Multiple Type Declaration Sections

**Features Added:**

- Multiple type declaration sections within the same file
- Type organization across multiple declarations

Example:

```cpl
types
    Currency = Number(precision=2)

types
    AccountNumber = String(pattern="[0-9]{10}")
```

**Limitations:**

- No Instant type support
- No Interval type support

**Value Provided:**

Types can be organized in multiple sections for better code organization.

### Increment 4: Instant Type

**Features Added:**

- **Instant type** with field-based constraints:
  - `year`, `month`, `day`, `hour`, `minute`, `second`, `day_of_week`, `week_number`
  - Omitted fields are valid and mean underspecification
  - `Instant` has maximum one-second resolution and always uses UTC
- No `precision` parameter; no `restrictions` parameter
- Integration with unit and measure declarations for time-related conversions through normal CPL rules

Example:

```cpl
types
    FiscalYearStart = Instant(month=January, day=1, hour=0, minute=0, second=0)
    TradingDay = Instant(day_of_week=Monday)
    Midnight = Instant(hour=0, minute=0, second=0)
```

**Prerequisites:**

- Unit declaration machinery for duration units
- Measure declaration machinery for time calculations
- Standard definitions for `Duration` concept and unit group

**Limitations:**

- No Interval type support

**Value Provided:**

Instant type enables temporal calculations, timestamp tracking, and scheduling. Constraint-based matching allows underspecified instants (e.g., every Monday, every midnight). Integration with units and measures follows normal CPL rules.

### Increment 5: Interval Type

**Features Added:**

- **Interval type** with canonical `starts_at` parameter:
  - `starts_at` accepts a named `Instant`-derived type reference or an inline `Instant` constraint expression
  - `unit` references a normal unit group (no special built-in treatment)
- `Duration` is not a type; it only provides concept values for units through normal CPL rules

Example:

```cpl
concepts
    Duration = [Values[Second, Minute, Hour, Day]]

units
    Duration = Duration

types
    FiscalYearStart = Instant(month=January, day=1, hour=0, minute=0, second=0)
    TradingDay = Instant(day_of_week=Monday)
    QuarterInterval = Interval(starts_at=FiscalYearStart, unit=Duration)
```

**Prerequisites:**

- Instant type implementation
- Unit declaration machinery for duration units
- Measure declaration machinery for time calculations

**Limitations:**

None - Full type functionality for all five built-in types.

**Value Provided:**

Interval type enables period-based calculations, fiscal period tracking, and duration-based scheduling. Integration with units and measures follows normal CPL rules without special built-in `Duration` treatment.

## Implications

### Positive Implications

- Reduced implementation risk through iterative delivery
- Early availability of core type functionality
- Opportunity for feedback and course correction
- Clear milestones and progress tracking
- Each increment provides usable value
- Time-related types benefit from mature unit/measure infrastructure

### Negative Implications / Trade-offs

- Requires discipline to maintain backward compatibility between increments
- Some architectural decisions must anticipate future features (e.g., time-related types)
- Intermediate limitations may constrain early use cases
- Documentation must clearly track which types are available in each increment
- Testing must cover both current increment features and ensure no regression of prior increments

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
- Type conversion rules may need refinement based on real-world usage

## Notes

**Related Documents:**

- [Types Definitions Design](../types-definitions.md) - Comprehensive design for type functionality
- [CPL Architecture - Introduction](../../architecture/cpl/introduction.md) - Overall CPL processing lifecycle
- [CPL Requirements - Concepts, Types, Units, Measures](../../requirements/cpl/concepts-types-units-measures-syntax.md) - Syntax requirements
- Parsing Phase ERO: [ERO002_types-parsing-phase.md](../../go-parser-target/eros/ERO002_types-parsing-phase.md) - Phase-specific parsing structures
- Analysis Phase ERO: [ERO002_types-analysis-phase.md](../../go-analysis/eros/ERO002_types-analysis-phase.md) - Phase-specific analysis structures
- Synthesis Phase ERO: [ERO002_types-synthesis-phase.md](../../go-synthesis-target/eros/ERO002_types-synthesis-phase.md) - Phase-specific synthesis structures

**Phase-Specific EROs:**

Each phase (parsing, analysis, synthesis) has its own ERO documenting the data structures and functionality specific to that phase across all increments. These EROs should be consulted for detailed phase-specific designs.
