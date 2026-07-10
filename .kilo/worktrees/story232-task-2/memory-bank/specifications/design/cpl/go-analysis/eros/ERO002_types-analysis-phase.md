# ERO002: Types Analysis Phase

## Title

Types Analysis Phase - Data Structures and Incremental Implementation

## Purpose/Goal

Define the analysis phase data structures for type definitions and document how these structures evolve across each increment of the types incremental implementation.

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic001_cpl-concepts-implementation.md](../../../../../epics/Epic001_cpl-concepts-implementation.md) - Epic for CPL concepts implementation

## Context

### Situation

The analysis phase is responsible for validating type definitions, building the Types lookup map, and enabling semantic analysis of type references in metrics, aggregators, spreaders, and lookups. The analysis phase transforms parser-level type definitions into analysis-level structures optimized for validation and code generation.

### Constraints

- Must work with parser-level type definitions from the symbol table
- Must produce structures usable by the CIL generator
- Must validate all type references in the CPL program
- Must support the incremental feature set of each increment
- Must enforce type constraint inheritance rules

### Requirements

- Build a unified Types lookup map from parser definitions
- Validate type derivation chains (derived must be at least as restrictive as base)
- Compute constraint compatibility for type extensions
- Enable metric type validation

### Assumptions

- Parser type definitions are complete and syntactically valid
- The analysis phase has access to all parsed files in the compilation unit
- Analysis errors prevent CIL generation
- Built-in types are pre-populated before analysis begins

## Data Structures by Increment

### Increment 1: Number Type with Type Capture

**AnalysisDefinitions - Types Lookup Map**

```go
// AnalysisDefinitions contains all analyzed definitions
type AnalysisDefinitions struct {
    Concepts map[string]AnalysisConceptDefinition
    Types    map[string]AnalysisTypeDefinition    // NEW: Types lookup map
    // ... other analysis definitions
}

// AnalysisTypeDefinition contains the analyzed type information
type AnalysisTypeDefinition struct {
    Name           string
    BaseType       BuiltInType
    Constraints    TypeConstraints
    DerivationPath []string  // Chain from this type to built-in root
}

// TypeConstraints is a sum type for different constraint types
type TypeConstraints interface {
    isTypeConstraints()
    Validate() error
}

// NumberConstraints implements TypeConstraints for Number types
type NumberConstraints struct {
    Precision *int
    Minimum   *int
    Maximum   *int
}

func (n NumberConstraints) isTypeConstraints() {}

// Validate checks Number constraint consistency
func (n NumberConstraints) Validate() error {
    if n.Precision != nil && *n.Precision < 0 {
        return fmt.Errorf("precision must be non-negative, got %d", *n.Precision)
    }
    if n.Minimum != nil && n.Maximum != nil && *n.Minimum >= *n.Maximum {
        return fmt.Errorf("minimum (%d) must be less than maximum (%d)", *n.Minimum, *n.Maximum)
    }
    return nil
}

// Note: StringConstraints and BooleanConstraints will be added in Increment 2
```

**CreateConstraints Factory Function**

```go
// CreateConstraints converts parser constraints to analysis constraints (Increment 1).
// Performs semantic validation including duplicate constraint name checking and
// value type validation that was removed from the parsing phase.
func CreateConstraints(baseType BuiltInType, parserConstraints []parser.Constraint) (TypeConstraints, error) {
    // Check for duplicate constraint names
    seenConstraints := make(map[string]bool)
    for _, pc := range parserConstraints {
        if seenConstraints[pc.Name] {
            return nil, fmt.Errorf("duplicate constraint name: %s", pc.Name)
        }
        seenConstraints[pc.Name] = true
    }

    switch baseType {
    case NumberType:
        return createNumberConstraints(parserConstraints)
    default:
        return nil, fmt.Errorf("unsupported base type: %v", baseType)
    }
}

func createNumberConstraints(parserConstraints []parser.Constraint) (NumberConstraints, error) {
    nc := NumberConstraints{}

    for _, pc := range parserConstraints {
        switch pc.Name {
        case "precision":
            // Value type validation: precision must be an integer
            if v, ok := pc.Value.(int); ok {
                nc.Precision = &v
            } else {
                return nc, fmt.Errorf("precision constraint must be an integer, got %T", pc.Value)
            }
        case "minimum":
            // Value type validation: minimum must be an integer
            if v, ok := pc.Value.(int); ok {
                nc.Minimum = &v
            } else {
                return nc, fmt.Errorf("minimum constraint must be an integer, got %T", pc.Value)
            }
        case "maximum":
            // Value type validation: maximum must be an integer
            if v, ok := pc.Value.(int); ok {
                nc.Maximum = &v
            } else {
                return nc, fmt.Errorf("maximum constraint must be an integer, got %T", pc.Value)
            }
        default:
            return nc, fmt.Errorf("unknown constraint for Number: %s", pc.Name)
        }
    }

    // Semantic validation: constraint consistency (e.g., min < max)
    return nc, nc.Validate()
}

// Note: createStringConstraints and createBooleanConstraints will be added in Increment 2
```

**Type Derivation Path**

```go
// BuildDerivationPath computes the chain from a derived type to its built-in root.
// For Increment 1, types can only derive directly from built-in types.
func BuildDerivationPath(typeName string, types map[string]AnalysisTypeDefinition) ([]string, error) {
    path := []string{typeName}

    current := types[typeName]
    // In Increment 1, types always derive directly from built-in types
    // The path is just [typeName, baseTypeName]
    path = append(path, current.BaseType.String())

    return path, nil
}
```

**Types Analysis Functions**

The types analysis phase uses function-based handlers following the same pattern as the concepts analysis phase. These functions are registered with the symbol table walker and called during traversal.

```go
// EnterAnalyzeTypesScope processes type definitions when entering a TypesScope.
// This function is registered with WalkSymbolTable and called for each
// TypesScope encountered during the walk.
func EnterAnalyzeTypesScope(table *symboltable.SymbolTable, context *AnalysisContext, analyzer *CPLAnalyzer) error {
    for _, def := range table.Definitions {
        if def.DefinitionType != symboltable.TypeDefinition {
            continue
        }

        parserDef, ok := def.Data.(*parser.ParserTypeDefinition)
        if !ok {
            return fmt.Errorf("invalid type definition data for %s", *def.Name)
        }

        // Cross-section duplicate detection: Check for duplicate type names
        if _, exists := analyzer.types[*def.Name]; exists {
            return fmt.Errorf("duplicate type definition: %s", *def.Name)
        }

        // Create constraints
        constraints, err := CreateConstraints(parserDef.BaseType, parserDef.Constraints)
        if err != nil {
            return fmt.Errorf("invalid constraints for type %s: %w", *def.Name, err)
        }

        // Build analysis type definition
        analysisDef := AnalysisTypeDefinition{
            Name:           *def.Name,
            BaseType:       parserDef.BaseType,
            Constraints:    constraints,
            DerivationPath: []string{*def.Name, parserDef.BaseType.String()},
        }

        analyzer.types[*def.Name] = analysisDef
    }

    return nil
}

// ExitAnalyzeTypesScope performs any cleanup or validation when exiting a TypesScope.
// Currently, no cleanup is required for type scopes, but this function is included
// for pattern consistency with other scope types.
func ExitAnalyzeTypesScope(table *symboltable.SymbolTable, context *AnalysisContext, analyzer *CPLAnalyzer) error {
    // No cleanup or validation needed for TypesScope exit
    return nil
}
```

**Integration with WalkSymbolTable**

The types analysis handlers are registered with the symbol table walker using the entryHandlers and exitHandlers maps, following the same pattern as concepts analysis:

```go
func (c *CPLCompiler) Analyze(rootTable *symboltable.SymbolTable) (*AnalysisDefinitions, error) {
    // Create analysis context and analyzer
    context := NewAnalysisContext()
    analyzer := NewCPLAnalyzer()

    // Register entry handlers for each scope type
    entryHandlers := map[symboltable.ScopeType]AnalysisFunc{
        symboltable.ConceptsScope: func(table *symboltable.SymbolTable, ctx *AnalysisContext, a *CPLAnalyzer) error {
            return EnterAnalyzeConceptScope(table, ctx, a)
        },
        symboltable.TypesScope: func(table *symboltable.SymbolTable, ctx *AnalysisContext, a *CPLAnalyzer) error {
            return EnterAnalyzeTypesScope(table, ctx, a)
        },
        // Future handlers:
        // symboltable.UnitsScope: EnterAnalyzeUnitsScope,
        // symboltable.MetricsScope: EnterAnalyzeMetricsScope,
    }

    // Register exit handlers for each scope type
    exitHandlers := map[symboltable.ScopeType]AnalysisFunc{
        symboltable.ConceptsScope: func(table *symboltable.SymbolTable, ctx *AnalysisContext, a *CPLAnalyzer) error {
            return ExitAnalyzeConceptScope(table, ctx, a)
        },
        symboltable.TypesScope: func(table *symboltable.SymbolTable, ctx *AnalysisContext, a *CPLAnalyzer) error {
            return ExitAnalyzeTypesScope(table, ctx, a)
        },
        // Future exit handlers:
        // symboltable.UnitsScope: ExitAnalyzeUnitsScope,
        // symboltable.MetricsScope: ExitAnalyzeMetricsScope,
    }

    // Single walk with all handlers
    if err := WalkSymbolTableWithHandlers(rootTable, entryHandlers, exitHandlers); err != nil {
        return nil, fmt.Errorf("analysis failed: %w", err)
    }

    // Validate results
    if err := validateHierarchy(analyzer.concepts); err != nil {
        return nil, err
    }

    return &AnalysisDefinitions{
        Concepts:  analyzer.concepts,
        Hierarchy: analyzer.hierarchy,
        Types:     analyzer.types,
    }, nil
}
```

**New Scope Type for Types**

```go
const (
    // GlobalScope is the root scope of the CPL document.
    GlobalScope ScopeType = iota
    // UnitsScope is the scope containing unit declarations.
    UnitsScope
    // ConceptsScope is the scope containing concept declarations.
    ConceptsScope
    // TypesScope is the scope containing type declarations.
    TypesScope  // NEW: For Increment 1
)
```

**Validation During Analysis**

- **Duplicate type definitions are an error** - Each type name must be unique across the compilation unit
- Constraint consistency (e.g., min < max for numbers, valid regex for patterns)
- Constraint values must be valid for the base type
- Built-in types cannot be redefined

**Limitations**

- Types can only derive directly from built-in types (no type extension)
- Only Number base type supported (no String, Boolean, Instant, Interval)
- No type compatibility checking with measures/units
- No multiple type sections support

### Increment 2: String and Boolean Types

**Enhanced TypeConstraints for String and Boolean**

```go
// StringConstraints implements TypeConstraints for String types
type StringConstraints struct {
    Minimum *int
    Maximum *int
    Pattern *regexp.Regexp  // Compiled regex
}

func (s StringConstraints) isTypeConstraints() {}

// Validate checks String constraint consistency
func (s StringConstraints) Validate() error {
    if s.Minimum != nil && *s.Minimum < 0 {
        return fmt.Errorf("minimum length must be non-negative, got %d", *s.Minimum)
    }
    if s.Maximum != nil && *s.Maximum < 0 {
        return fmt.Errorf("maximum length must be non-negative, got %d", *s.Maximum)
    }
    if s.Minimum != nil && s.Maximum != nil && *s.Minimum > *s.Maximum {
        return fmt.Errorf("minimum length (%d) must not exceed maximum length (%d)", *s.Minimum, *s.Maximum)
    }
    return nil
}

// BooleanConstraints implements TypeConstraints for Boolean types
type BooleanConstraints struct {
    // Boolean has no constraints
}

func (b BooleanConstraints) isTypeConstraints() {}

// Validate checks Boolean constraints (always valid)
func (b BooleanConstraints) Validate() error {
    return nil
}
```

**CreateConstraints Enhancement**

```go
// CreateConstraints (Increment 2) adds support for String and Boolean.
func CreateConstraints(baseType BuiltInType, parserConstraints []parser.Constraint) (TypeConstraints, error) {
    switch baseType {
    case NumberType:
        return createNumberConstraints(parserConstraints)
    case StringType:
        return createStringConstraints(parserConstraints)
    case BooleanType:
        return createBooleanConstraints(parserConstraints)
    default:
        return nil, fmt.Errorf("unsupported base type: %v", baseType)
    }
}

func createStringConstraints(parserConstraints []parser.Constraint) (StringConstraints, error) {
    sc := StringConstraints{}

    for _, pc := range parserConstraints {
        switch pc.Name {
        case "minimum":
            // Value type validation: minimum must be an integer (length)
            if v, ok := pc.Value.(int); ok {
                sc.Minimum = &v
            } else {
                return sc, fmt.Errorf("minimum constraint must be an integer, got %T", pc.Value)
            }
        case "maximum":
            // Value type validation: maximum must be an integer (length)
            if v, ok := pc.Value.(int); ok {
                sc.Maximum = &v
            } else {
                return sc, fmt.Errorf("maximum constraint must be an integer, got %T", pc.Value)
            }
        case "pattern":
            // Value type validation: pattern must be a string
            if v, ok := pc.Value.(string); ok {
                re, err := regexp.Compile(v)
                if err != nil {
                    return sc, fmt.Errorf("invalid pattern regex: %w", err)
                }
                sc.Pattern = re
            } else {
                return sc, fmt.Errorf("pattern constraint must be a string, got %T", pc.Value)
            }
        default:
            return sc, fmt.Errorf("unknown constraint for String: %s", pc.Name)
        }
    }

    // Semantic validation: constraint consistency (e.g., min <= max length)
    return sc, sc.Validate()
}

func createBooleanConstraints(parserConstraints []parser.Constraint) (BooleanConstraints, error) {
    if len(parserConstraints) > 0 {
        return BooleanConstraints{}, fmt.Errorf("Boolean type cannot have constraints")
    }
    return BooleanConstraints{}, nil
}
```

**Validation During Analysis**

- String constraints: minimum/maximum length validation, regex pattern compilation
- Boolean constraints: no constraints allowed
- All basic scalar type constraints validated

**Limitations**

- No Instant type support
- No Interval type support
- No type extension
- No multiple type sections support

### Increment 3: Multiple Type Declaration Sections

**Overview**

Increment 3 allows multiple `types` sections in the same CPL file. Since each section creates its own `TypesScope` symbol table under the GlobalScope, the existing `AnalyzeTypesScope` function automatically processes all type definitions from all sections during the single `WalkSymbolTableWithHandlers` traversal.

**No Special Merging Required**

Unlike what was originally planned, there is no need for a separate merging function. The symbol table walk naturally visits each `TypesScope` table, and the `AnalyzeTypesScope` handler accumulates all type definitions into the analyzer's types map.

**Cross-Section Duplicate Detection**

The duplicate check in `EnterAnalyzeTypesScope` uses the analyzer's types map to track seen type names across all sections:

```go
// Cross-section duplicate detection: Check for duplicate type names
// The analyzer.types map accumulates all type definitions across multiple
// TypesScope tables during the symbol table walk
if _, exists := analyzer.types[*def.Name]; exists {
    return fmt.Errorf("duplicate type definition: %s", *def.Name)
}
```

Since the symbol table walk visits all `TypesScope` tables (regardless of which section they came from), any attempt to define the same type name in multiple sections will be caught as a duplicate. The `analyzer.types` map persists across handler calls, enabling cross-section duplicate detection.

**Validation During Analysis**

- **Duplicate type definitions across sections are an error** - The existing duplicate check in `AnalyzeTypesScope` catches these
- No other cross-section validation is needed for Increment 3

**Limitations**

- No Instant type support
- No Interval type support
- No type extension

### Increment 4: Instant Type

**Enhanced TypeConstraints for Instant**

```go
// InstantConstraints implements TypeConstraints for Instant types.
// Every supplied field is a constraint on matching UTC moments.
// Omitted fields are valid and mean underspecification.
type InstantConstraints struct {
    Year       *int
    Month      *string
    Day        *int
    Hour       *int
    Minute     *int
    Second     *int
    DayOfWeek  *string
    WeekNumber *int
}

func (i InstantConstraints) isTypeConstraints() {}

// Validate checks Instant constraint consistency.
// Validates ranges and compatible field combinations.
// Treats omitted fields as underspecification, not an error.
func (i InstantConstraints) Validate() error {
    if i.Year != nil && *i.Year < 0 {
        return fmt.Errorf("year must be non-negative, got %d", *i.Year)
    }
    if i.Month != nil {
        validMonths := map[string]bool{"January":true, "February":true, "March":true, "April":true, "May":true, "June":true, "July":true, "August":true, "September":true, "October":true, "November":true, "December":true}
        if !validMonths[*i.Month] {
            return fmt.Errorf("invalid month: %s", *i.Month)
        }
    }
    if i.Day != nil && (*i.Day < 1 || *i.Day > 31) {
        return fmt.Errorf("day must be between 1 and 31, got %d", *i.Day)
    }
    if i.Hour != nil && (*i.Hour < 0 || *i.Hour > 23) {
        return fmt.Errorf("hour must be between 0 and 23, got %d", *i.Hour)
    }
    if i.Minute != nil && (*i.Minute < 0 || *i.Minute > 59) {
        return fmt.Errorf("minute must be between 0 and 59, got %d", *i.Minute)
    }
    if i.Second != nil && (*i.Second < 0 || *i.Second > 59) {
        return fmt.Errorf("second must be between 0 and 59, got %d", *i.Second)
    }
    if i.DayOfWeek != nil {
        validDays := map[string]bool{"Monday":true, "Tuesday":true, "Wednesday":true, "Thursday":true, "Friday":true, "Saturday":true, "Sunday":true}
        if !validDays[*i.DayOfWeek] {
            return fmt.Errorf("invalid day_of_week: %s", *i.DayOfWeek)
        }
    }
    if i.WeekNumber != nil && (*i.WeekNumber < 1 || *i.WeekNumber > 53) {
        return fmt.Errorf("week_number must be between 1 and 53, got %d", *i.WeekNumber)
    }
    return nil
}
```

**CreateConstraints Enhancement**

```go
// CreateConstraints (Increment 4) adds support for Instant.
func CreateConstraints(baseType BuiltInType, parserConstraints []parser.Constraint) (TypeConstraints, error) {
    switch baseType {
    case NumberType:
        return createNumberConstraints(parserConstraints)
    case StringType:
        return createStringConstraints(parserConstraints)
    case BooleanType:
        return createBooleanConstraints(parserConstraints)
    case InstantType:
        return createInstantConstraints(parserConstraints)
    default:
        return nil, fmt.Errorf("unsupported base type: %v", baseType)
    }
}

func createInstantConstraints(parserConstraints []parser.Constraint) (InstantConstraints, error) {
    ic := InstantConstraints{}

    for _, pc := range parserConstraints {
        switch pc.Name {
        case "year":
            if v, ok := pc.Value.(int); ok {
                ic.Year = &v
            } else {
                return ic, fmt.Errorf("year constraint must be an integer, got %T", pc.Value)
            }
        case "month":
            if v, ok := pc.Value.(string); ok {
                ic.Month = &v
            } else {
                return ic, fmt.Errorf("month constraint must be a string, got %T", pc.Value)
            }
        case "day":
            if v, ok := pc.Value.(int); ok {
                ic.Day = &v
            } else {
                return ic, fmt.Errorf("day constraint must be an integer, got %T", pc.Value)
            }
        case "hour":
            if v, ok := pc.Value.(int); ok {
                ic.Hour = &v
            } else {
                return ic, fmt.Errorf("hour constraint must be an integer, got %T", pc.Value)
            }
        case "minute":
            if v, ok := pc.Value.(int); ok {
                ic.Minute = &v
            } else {
                return ic, fmt.Errorf("minute constraint must be an integer, got %T", pc.Value)
            }
        case "second":
            if v, ok := pc.Value.(int); ok {
                ic.Second = &v
            } else {
                return ic, fmt.Errorf("second constraint must be an integer, got %T", pc.Value)
            }
        case "day_of_week":
            if v, ok := pc.Value.(string); ok {
                ic.DayOfWeek = &v
            } else {
                return ic, fmt.Errorf("day_of_week constraint must be a string, got %T", pc.Value)
            }
        case "week_number":
            if v, ok := pc.Value.(int); ok {
                ic.WeekNumber = &v
            } else {
                return ic, fmt.Errorf("week_number constraint must be an integer, got %T", pc.Value)
            }
        default:
            return ic, fmt.Errorf("unknown constraint for Instant: %s", pc.Name)
        }
    }

    return ic, ic.Validate()
}
```

**Validation During Analysis**

- Instant constraints: field values must be in valid ranges, compatible field combinations are accepted, omitted fields are valid underspecification

**Limitations**

- No Interval type support
- No type extension

### Increment 5: Interval Type

**Enhanced TypeConstraints for Interval**

```go
// IntervalConstraints implements TypeConstraints for Interval types.
// starts_at is either a named Instant-derived type reference or an inline Instant constraint expression.
// unit references a normal unit group (no special built-in treatment).
type IntervalConstraints struct {
    StartsAt      *InstantConstraints // Inline Instant constraints, or nil if using a named type
    StartTypeName *string             // Named Instant-derived type reference, or nil if inline
    UnitName      *string             // Name of the unit group
}

func (i IntervalConstraints) isTypeConstraints() {}

// Validate checks Interval constraint consistency.
func (i IntervalConstraints) Validate() error {
    if i.StartsAt == nil && i.StartTypeName == nil {
        return fmt.Errorf("Interval type must specify starts_at as either an inline Instant constraint or a named Instant-derived type")
    }
    if i.StartsAt != nil && i.StartTypeName != nil {
        return fmt.Errorf("Interval starts_at cannot be both an inline Instant constraint and a named type reference")
    }
    if i.UnitName == nil {
        return fmt.Errorf("Interval type must specify a unit")
    }
    return nil
}
```

**CreateConstraints Enhancement**

```go
// CreateConstraints (Increment 5) adds support for Interval.
func CreateConstraints(baseType BuiltInType, parserConstraints []parser.Constraint) (TypeConstraints, error) {
    switch baseType {
    case NumberType:
        return createNumberConstraints(parserConstraints)
    case StringType:
        return createStringConstraints(parserConstraints)
    case BooleanType:
        return createBooleanConstraints(parserConstraints)
    case InstantType:
        return createInstantConstraints(parserConstraints)
    case IntervalType:
        return createIntervalConstraints(parserConstraints)
    default:
        return nil, fmt.Errorf("unsupported base type: %v", baseType)
    }
}

func createIntervalConstraints(parserConstraints []parser.Constraint) (IntervalConstraints, error) {
    ic := IntervalConstraints{}

    for _, pc := range parserConstraints {
        switch pc.Name {
        case "starts_at":
            // starts_at can be either a named type reference (string) or inline Instant constraints
            if v, ok := pc.Value.(string); ok {
                ic.StartTypeName = &v
            } else if v, ok := pc.Value.(parser.InlineInstantConstraints); ok {
                // Inline constraints are parsed as a nested constraint set
                inlineConstraints, err := createInstantConstraints(v.Constraints)
                if err != nil {
                    return ic, fmt.Errorf("invalid inline Instant constraints in starts_at: %w", err)
                }
                ic.StartsAt = &inlineConstraints
            } else {
                return ic, fmt.Errorf("starts_at must be a named Instant type or inline Instant constraints, got %T", pc.Value)
            }
        case "unit":
            if v, ok := pc.Value.(string); ok {
                ic.UnitName = &v
            } else {
                return ic, fmt.Errorf("unit constraint must be a string, got %T", pc.Value)
            }
        default:
            return ic, fmt.Errorf("unknown constraint for Interval: %s", pc.Name)
        }
    }

    return ic, ic.Validate()
}
```

**Cross-Reference Validation for Interval Types**

```go
// ValidateIntervalTypes checks that Interval types reference valid types/units.
// This function operates on the analyzer's types map following analysis completion.
func ValidateIntervalTypes(analyzer *CPLAnalyzer) error {
    for name, t := range analyzer.types {
        switch tc := t.Constraints.(type) {
        case IntervalConstraints:
            // If starts_at is a named type, verify it exists and is an Instant-derived type
            if tc.StartTypeName != nil {
                startType, exists := analyzer.types[*tc.StartTypeName]
                if !exists {
                    return fmt.Errorf("Interval type %s references unknown starts_at type: %s", name, *tc.StartTypeName)
                }
                if startType.BaseType != InstantType {
                    return fmt.Errorf("Interval type %s starts_at type %s must be Instant or derived from Instant, got %v",
                        name, *tc.StartTypeName, startType.BaseType)
                }
            }

            // If starts_at is inline, the inline constraints were already validated by createIntervalConstraints

            // Verify unit exists through normal unit definitions (requires units to be analyzed)
            // This validation happens after units analysis; no special built-in Duration treatment
        }
    }
    return nil
}
```

**Validation During Analysis**

- Interval constraints: starts_at type must exist and be Instant or derived from Instant, unit must exist
- Cross-references between Interval and Instant types validated

**Limitations**

None - Full type analysis functionality for all five built-in types.

## Approach

### Incremental Structure and Function Evolution

The analysis phase uses an incremental approach where structures and functions evolve by adding fields and capabilities as each increment requires them:

**Structure Evolution:**

- **Increment 1:** `AnalysisTypeDefinition` starts with basic fields: `Name`, `BaseType`, `Constraints`, `DerivationPath`
  - `NumberConstraints` defined
- **Increment 2:** Adds `StringConstraints`, `BooleanConstraints` for basic scalar types
- **Increment 3:** No new fields - uses existing duplicate detection across multiple sections
- **Increment 4:** Adds `InstantConstraints` with field-based constraints for time types
- **Increment 5:** Adds `IntervalConstraints` with `starts_at` and unit references for interval types

**Function Evolution:**

- **Increment 1:** Basic `AnalyzeTypesScope` processes type definitions; `CreateConstraints` factory function; constraint validation
- **Increment 2:** Enhanced with String and Boolean type constraints
- **Increment 3:** No changes needed - existing duplicate detection handles multiple sections
- **Increment 4:** Enhanced with time type constraints (`InstantConstraints`)
- **Increment 5:** Enhanced with `IntervalConstraints`, `ValidateIntervalTypes` for cross-reference validation

**No Feature Flags:**

There are no feature flags or runtime checks for increment capabilities. Each increment's code uses only the fields and functions available at that point in development.

**Minimal Nil Values:**

Fields are only added when needed, and nil values are used only when the final system design requires optional fields (e.g., `ParentType` is nil for types deriving directly from built-ins). There are no "placeholder" nil fields waiting for future increments.

## Decision

**Chosen Approach:** Incremental evolution of structures and functions

**Rationale:**
This approach mirrors the natural development process. Each increment builds upon working code from previous increments by adding only what's needed. No feature flags or complex abstractions are required because each increment's implementation only uses the capabilities it defines. The final result is clean, focused code where every field and function serves a purpose in the completed system.

## Implications

### Positive Implications

- Each increment builds upon working, tested code from previous increments
- No dead code or placeholder fields in the final implementation
- Structures contain only fields that are actively used
- Functions evolve naturally as requirements grow
- No cognitive overhead from feature flags or conditional logic

### Negative Implications

- Need to maintain this ERO to track what exists at each increment
- Documentation must clearly indicate which increment introduced each field/function
- Cannot easily "backport" features to earlier increments

### Validation Strategy

Each increment validates progressively, with each new capability adding its own validation:

1. **Increment 1:** Basic type validation, Number constraint consistency, unique type names, type capture
2. **Increment 2:** String and Boolean constraint validation
3. **Increment 3:** Duplicate detection across multiple type sections
4. **Increment 4:** Instant constraint validation, range validation, compatible field-combination validation
5. **Increment 5:** Interval constraint validation, cross-reference validation for Interval types

## Notes

**Related Documents:**

- [Types Incremental Implementation](../../eros/ERO002_types-incremental-implementation.md) - Main incremental implementation ERO
- [Types Definitions Design](../../types-definitions.md) - Full type design
- [Intermediate Language](../../architecture/cpl/analysis/intermediate_language.md) - CIL specification
