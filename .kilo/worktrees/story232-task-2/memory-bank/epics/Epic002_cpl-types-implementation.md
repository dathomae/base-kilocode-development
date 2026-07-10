# Epic002: CPL Types Implementation

## Title

CPL Types Implementation

## Ultimate Goal

Implement the complete types functionality for the compiler for the Chronocone Planning Language (CPL) across all processing phases. Types define constraints and validation rules for data values used in metrics, providing precision control for numeric values, length and pattern constraints for strings, and temporal constraints for time-related calculations. This epic delivers incremental feature implementation from basic Number types through full Interval type support. The compiler operation is broken into three phases: parsing, analysis, and synthesis.

## References

- [Types Definitions Design](../specifications/design/cpl/types-definitions.md) - Comprehensive design for type functionality
- [CPL Architecture - Introduction](../specifications/architecture/cpl/introduction.md) - Overall CPL processing lifecycle
- [ERO002: Types Incremental Implementation](../specifications/design/cpl/eros/ERO002_types-incremental-implementation.md) - Main incremental implementation ERO
- [ERO002: Types Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO002_types-parsing-phase.md) - Phase-specific parsing structures
- [ERO002: Types Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md) - Phase-specific analysis structures
- [ERO002: Types Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO002_types-synthesis-phase.md) - Phase-specific synthesis structures

## Dependencies

- [Epic001_cpl-concepts-implementation.md](Epic001_cpl-concepts-implementation.md) - Concepts implementation must be complete as types build on the same parser infrastructure

## Stories

| Story File                                                                                                      | Status      | Description                                                                                                                                                                                                                                                                                                         | ERO                                                                                                               |
| --------------------------------------------------------------------------------------------------------------- | ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| [Story131_cpl-types-parser-number.md](../stories/finished-stories/Story131_cpl-types-parser-number.md)          | Done        | **Parser (Syntactic)**: Grammar changes for types section, Number type recognition, raw constraint capture (key=value format), TypesScope symbol table integration. Performs syntactic validation only (valid identifiers, recognized keywords).                                                                    | [ERO002_types-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO002_types-parsing-phase.md) |
| [Story132_cpl-types-analyzer-number.md](../stories/finished-stories/Story132_cpl-types-analyzer-number.md)      | Done        | **Analyzer (Semantic)**: AnalysisTypeDefinition structure, Types lookup map, NumberConstraints validation (constraint values, duplicate constraint names, min<max), CreateConstraints factory function, cross-section duplicate detection during symbol table walk.                                                 | [ERO002_types-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md)    |
| [Story196_cpl-types-string-boolean.md](../stories/Story196_cpl-types-string-boolean.md)                         | Not Started | **Analysis**: Add StringType/BooleanType to BuiltInType, StringConstraints (min/max length, regex pattern with compile-time validation), BooleanConstraints, CreateConstraints cases, createTypeReference/parseBuiltInTypeName entries, derivation validation/merging. Combines and supersedes Story133 + Story134. | [ERO002_types-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md)    |
| [Story135_cpl-types-parser-multiple-sections.md](../stories/Story135_cpl-types-parser-multiple-sections.md)     | Not Started | **Parser (Syntactic)**: Support multiple types sections in same file. Each section creates a separate TypesScope. No duplicate detection at parse time.                                                                                                                                                             | [ERO002_types-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO002_types-parsing-phase.md) |
| [Story136_cpl-types-analyzer-multiple-sections.md](../stories/Story136_cpl-types-analyzer-multiple-sections.md) | Not Started | **Analyzer (Semantic)**: Cross-section duplicate detection during symbol table walk. The analyzer's types map accumulates definitions across all TypesScope tables.                                                                                                                                                 | [ERO002_types-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md)    |
| [Story137_cpl-types-parser-instant.md](../stories/Story137_cpl-types-parser-instant.md)                         | Not Started | **Parser (Syntactic)**: Add Instant base type keyword, raw capture of field constraints (year, month, day, hour, minute, second, day_of_week, week_number). Omitted fields are valid underspecification. Syntactic validation only.                                                                                 | [ERO002_types-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO002_types-parsing-phase.md) |
| [Story138_cpl-types-analyzer-instant.md](../stories/Story138_cpl-types-analyzer-instant.md)                     | Not Started | **Analyzer (Semantic)**: InstantConstraints with field-based constraint validation (ranges, compatible combinations), underspecification acceptance. No precision or restriction validation.                                                                                                                        | [ERO002_types-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md)    |
| [Story139_cpl-types-parser-interval.md](../stories/Story139_cpl-types-parser-interval.md)                       | Not Started | **Parser (Syntactic)**: Add Interval base type keyword, raw capture of `starts_at` (named Instant type or inline constraints) and unit reference. Syntactic validation only.                                                                                                                                        | [ERO002_types-parsing-phase.md](../specifications/design/cpl/go-parser-target/eros/ERO002_types-parsing-phase.md) |
| [Story140_cpl-types-analyzer-interval.md](../stories/Story140_cpl-types-analyzer-interval.md)                   | Not Started | **Analyzer (Semantic)**: IntervalConstraints validation, `starts_at` validation (named Instant type or inline constraints), unit reference validation through normal unit definitions. No special built-in Duration treatment.                                                                                      | [ERO002_types-analysis-phase.md](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md)    |

**Note on Synthesis Phase:** The synthesis phase requires no work for types implementation. Types are used by the analyzer for validation only; they do not generate CIL elements. The synthesizer generates Go code using appropriate types based on analysis decisions. See [ERO002_types-synthesis-phase.md](../specifications/design/cpl/go-synthesis-target/eros/ERO002_types-synthesis-phase.md) for details.

## Story Ordering and Dependencies

Stories in this epic follow a strict incremental sequence based on the ERO002 types incremental implementation plan:

### Phase 1: Number Type (Must be completed first)

_Prerequisites: Epic001 (Concepts Implementation)_

1. **[Story131_cpl-types-parser-number.md](../stories/Story131_cpl-types-parser-number.md)**: Grammar for types section, Number base type, constraint parsing
2. **[Story132_cpl-types-analyzer-number.md](../stories/Story132_cpl-types-analyzer-number.md)**: Type definition analysis, NumberConstraints validation

### Phase 2: String and Boolean Types

_Prerequisites: Story131, Story132 (Phase 1: Number Type)_

3. **[Story196_cpl-types-string-boolean.md](../stories/Story196_cpl-types-string-boolean.md)**: Add String and Boolean built-in type support to analysis phase (combines and supersedes Story133 + Story134)

### Phase 3: Multiple Type Sections

_Prerequisites: Story196 (Phase 2: String and Boolean Types)_

5. **[Story135_cpl-types-parser-multiple-sections.md](../stories/Story135_cpl-types-parser-multiple-sections.md)**: Support for multiple types sections
6. **[Story136_cpl-types-analyzer-multiple-sections.md](../stories/Story136_cpl-types-analyzer-multiple-sections.md)**: Cross-section duplicate detection

### Phase 4: Instant Type

_Prerequisites: Story135, Story136 (Phase 3: Multiple Type Sections)_

7. **[Story137_cpl-types-parser-instant.md](../stories/Story137_cpl-types-parser-instant.md)**: Add Instant base type with field-based constraints
8. **[Story138_cpl-types-analyzer-instant.md](../stories/Story138_cpl-types-analyzer-instant.md)**: Instant constraint validation with range and combination checks

### Phase 5: Interval Type (Final phase)

_Prerequisites: Story137, Story138 (Phase 4: Instant Type)_

9. **[Story139_cpl-types-parser-interval.md](../stories/Story139_cpl-types-parser-interval.md)**: Add Interval base type with `starts_at` and unit constraints
10. **[Story140_cpl-types-analyzer-interval.md](../stories/Story140_cpl-types-analyzer-interval.md)**: Interval cross-reference validation for `starts_at` and unit references

Each phase builds upon the previous phases. Within each increment, parser stories precede analyzer stories.

**Pattern Note:** Story131 and Story132 follow the original parser/analyzer pair pattern. Story196 consolidates the String/Boolean work into a single analysis-focused story since the parser already handles these types via the generic `baseType: identifier;` rule. Remaining stories (135-140) follow the parser/analyzer pair pattern.

## Built-in Types

The following five built-in types are always available without declaration:

### Number

Unrestricted real number with precision, minimum, and maximum constraints:

```cpl
types
    Currency = Number(precision=2)
    Temperature = Number(precision=1, minimum=-100, maximum=200)
```

### String

Unicode sequence with length and pattern constraints:

```cpl
types
    ShortString = String(maximum=50)
    CurrencyCode = String(pattern="[A-Z]{3}")
    Description = String(minimum=10, maximum=500)
```

### Boolean

Logical true/false value with no constraints:

```cpl
types
    ActiveFlag = Boolean
```

### Instant

UTC moment constraints. Every parameter is a constraint on matching moments. Omitted fields mean underspecification. Maximum one-second resolution, always UTC:

```cpl
types
    FiscalYearStart = Instant(month=January, day=1, hour=0, minute=0, second=0)
    TradingDay = Instant(day_of_week=Monday)
    Midnight = Instant(hour=0, minute=0, second=0)
```

### Interval

Time span with `starts_at` (named `Instant`-derived type or inline constraints) and unit reference. `Duration` is not a type:

```cpl
concepts
    Duration = [Values[Second, Minute, Hour, Day]]

units
    Duration = Duration

types
    FiscalYearStart = Instant(month=January, day=1, hour=0, minute=0, second=0)
    QuarterInterval = Interval(starts_at=FiscalYearStart, unit=Duration)
```

## Constraints

1. Each increment must produce a working system with documented limitations
2. Later increments must not break functionality from earlier increments
3. Type definitions must remain statically known at compile/analysis/synthesis time
4. All phases must align with the existing CPL phase architecture (parse → analyze → synthesize)
5. Follow the Logical TDD Lifecycle for all coding stories
6. Parser changes require corresponding ANTLR grammar updates
7. Analysis phase changes must maintain symbol table walking patterns
8. Derived types must be at least as restrictive as their base types (constraint inheritance)

## Acceptance Criteria

1. All parser stories implement the grammar features defined in ERO002 parsing phase
2. All analyzer stories implement the validation and structure building defined in ERO002 analysis phase
3. Each increment provides usable functionality with clearly documented limitations
4. Type definitions are validated at analysis time, not deferred to runtime
5. Constraint inheritance rules are enforced (derived types at least as restrictive as base)
6. All stories include comprehensive unit tests following BTDD practices
7. Documentation is updated to reflect which features are available at each increment

## Notes

**Implementation Details:**

- Type constraints define validation rules that are enforced during analysis
- The analyzer uses type definitions to validate metric type compatibility
- Type definitions do not generate CIL elements; they are metadata for the analyzer
- The synthesizer uses Go types (`float64`, `string`, `bool`, `time.Time`, etc.) based on analysis decisions
- Built-in types are pre-populated in the global scope before parsing begins
- Type capture allows forward references and progressive type definition

**Pre-Implementation Requirement:**

- The `TypesScope` constant must be added to the `ScopeType` enum in the symbol table before any types implementation stories can be completed. This is Task 1 of [Story151](../stories/Story151_epic002-parsing-analysis-separation.md), which establishes the correct parsing/analysis separation patterns.

**Parsing vs Analysis Phase Separation:**

The types implementation follows a strict separation of concerns between parsing and analysis phases:

- **Parsing Phase (Syntactic)**: Responsible for grammar recognition, raw data capture, and syntactic validation only
  - Validates type name syntax (valid identifiers)
  - Validates base type keywords are recognized (Number, String, Boolean, Instant, Interval)
  - Captures constraints as raw key=value pairs (no value validation)
  - Populates symbol table with `ParserTypeDefinition` structures
  - Uses unified `CPLListener` with methods in separate `cpl_types_listener.go` file

- **Analysis Phase (Semantic)**: Responsible for semantic validation and cross-reference checking
  - Validates constraint values (e.g., precision must be non-negative integer)
  - Validates constraint consistency (e.g., min < max for numbers, min <= max for strings)
  - Performs duplicate type name detection (within and across sections)
  - Compiles regex patterns for String constraints
  - Validates cross-references (e.g., Interval start type must reference a valid Instant type)
  - Uses function-based handlers (`EnterAnalyzeTypesScope`, `ExitAnalyzeTypesScope`)

This separation aligns with the patterns established in Epic001 (concepts implementation). See [Story151](../stories/Story151_epic002-parsing-analysis-separation.md) for details on the corrections made to establish these patterns.

**Corrected Documentation:**

The following ERO002 documents were corrected in Story151 to align with established patterns:

- [ERO002: Types Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO002_types-parsing-phase.md) - Updated to show syntactic validation only, unified listener pattern
- [ERO002: Types Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO002_types-analysis-phase.md) - Updated to show function-based handlers, semantic validation

**Related Epics:**

- Future epics for metrics, aggregators, and spreaders will use the type system for validation
- Instant and Interval types require unit/measure machinery that will be developed in parallel
