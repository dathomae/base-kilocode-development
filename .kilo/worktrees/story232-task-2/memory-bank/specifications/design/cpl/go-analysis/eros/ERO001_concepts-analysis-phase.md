# ERO001: Concepts Analysis Phase

## Title

Concepts Analysis Phase - Data Structures and Incremental Implementation

## Purpose/Goal

Define the analysis phase data structures for concept definitions and document how these structures evolve across each increment of the concepts incremental implementation.

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic001_cpl-concepts-implementation.md](../../../../../epics/Epic001_cpl-concepts-implementation.md) - Epic for CPL concepts implementation

## Context

### Situation

The analysis phase is responsible for validating concept definitions, building the Concepts lookup map, and enabling semantic analysis of concept references in metrics, aggregators, spreaders, and lookups. The analysis phase transforms parser-level concept definitions into analysis-level structures optimized for validation and code generation.

### Constraints

- Must work with parser-level concept definitions from the symbol table
- Must produce structures usable by the CIL generator
- Must validate all concept references in the CPL program
- Must support the incremental feature set of each increment

### Requirements

- Build a unified Concepts lookup map from parser definitions
- Validate concept hierarchies and relationships
- Compute value orderings including before/after insertions
- Support conditional value specifications validation
- Enable metric dimension validation

### Assumptions

- Parser concept definitions are complete and syntactically valid
- The analysis phase has access to all parsed files in the compilation unit
- Analysis errors prevent CIL generation

## Data Structures by Increment

### Increment 1: Basic Concept Definitions

**AnalysisDefinitions - Concepts Lookup Map**

```go
type AnalysisDefinitions struct {
    Concepts map[string]AnalysisConceptDefinition  // Concepts lookup map
    // ... other analysis definitions
}

// AnalysisConceptDefinition contains the analyzed concept information
type AnalysisConceptDefinition struct {
    ConceptName    string
    Generators     []Generator                              // Analysis-phase generators (Increment 1: Values only)
    Values         []AnalysisConceptValueDefinition         // Ordered list of generated values
    index     map[string]int                           // Value name -> position lookup
}

// AnalysisConceptValueDefinition represents a single concept value with its position.
// It's possible that this will eventually just be a string in which case the implementation can be simplified.
// Leaving it in for now until we figure out the end state for how concepts work.
type AnalysisConceptValueDefinition struct {
    Name     string
}
```

**Generator Interface and ValuesGenerator**

The analysis phase defines the `Generator` interface for value generation. The parser only captures AST structures; value generation happens during analysis. For Increment 1, only `ValuesGenerator` is supported.

```go
// Generator is the analysis-phase interface for concept value generators.
// Implementations generate the actual string values for a concept based on
// the parsed generator definition.
type Generator interface {
    // GenerateValues produces the concept values defined by this generator.
    // For ValuesGenerator, values are returned directly.
    // Future increments will support IntRangeGenerator and ConceptRangeGenerator.
    GenerateValues(concepts map[string]AnalysisConceptDefinition) ([]string, error)
}

// ValuesGenerator generates values from an explicit list of value strings.
type ValuesGenerator struct {
    Values []string
}

func (v *ValuesGenerator) GenerateValues(concepts map[string]AnalysisConceptDefinition) ([]string, error) {
    return v.Values, nil
}
```

**CreateGenerator Factory Function**

The analysis phase converts parser generator definitions to analysis generators. For Increment 1, only `ValuesGeneratorDef` is supported:

```go
// CreateGenerator converts a parser generator definition to an analysis generator.
// For Increment 1, only ValuesGeneratorDef is supported.
func CreateGenerator(parserGen interface{}) (Generator, error) {
    switch gen := parserGen.(type) {
    case parser.ValuesGeneratorDef:
        return &ValuesGenerator{Values: gen.Values}, nil
    default:
        return nil, fmt.Errorf("unknown generator type: %T", parserGen)
    }
}
```

**WalkSymbolTable with Handler Map**

The `WalkSymbolTable` function accepts a map from `ScopeType` to analysis functions. This allows a single walk to dispatch to different handlers based on the scope type encountered.

```go
// AnalysisFunc is the signature for scope-specific analysis functions.
// It receives the symbol table for the scope being visited and returns
// an error if analysis fails.
type AnalysisFunc func(table *SymbolTable) error

// WalkSymbolTableWithHandlers recursively walks the symbol table tree,
// calling the appropriate analysis function from the handlers map
// for each table based on its ScopeType.
//
// The walk is performed in pre-order (parent before children).
// If a scope type has no registered handler, that scope is skipped
// but its children are still visited.
// The first error returned by any handler stops the walk and is returned.
func WalkSymbolTableWithHandlers(table *SymbolTable, handlers map[ScopeType]AnalysisFunc) error {
    if table == nil {
        return nil
    }

    // Call the handler for this scope type if one is registered
    if handler, exists := handlers[table.ScopeType]; exists && handler != nil {
        if err := handler(table); err != nil {
            return err
        }
    }

    // Recursively visit all children
    for _, child := range table.Children {
        if err := WalkSymbolTableWithHandlers(child, handlers); err != nil {
            return err
        }
    }

    return nil
}
```

**Concepts Analysis Function**

For Increment 1, the analysis phase only needs to define the analysis function that processes concept definitions. The symbol table walk is handled by `WalkSymbolTable`.

```go
// AnalyzeConceptsScope processes concept definitions from a ConceptsScope
// symbol table and builds the analysis-level concepts lookup map.
//
// This function is registered with WalkSymbolTable via the handlers map
// and is called automatically when a ConceptsScope is encountered.
func AnalyzeConceptsScope(table *SymbolTable) error {
    // Access the analysis context to store the results
    ctx := GetAnalysisContext(table)
    if ctx.Concepts == nil {
        ctx.Concepts = make(map[string]AnalysisConceptDefinition)
    }

    for _, def := range table.Definitions {
        if def.DefinitionType != ConceptDefinition {
            continue
        }

        conceptDef, ok := def.Data.(*ParserConceptDefinition)
        if !ok {
            return fmt.Errorf("invalid concept definition data for %s", *def.Name)
        }

        if _, exists := ctx.Concepts[*def.Name]; exists {
            return fmt.Errorf("duplicate concept definition: %s", *def.Name)
        }

        info := AnalysisConceptDefinition{
            Name:       *def.Name,
            Values:     make([]AnalysisConceptValueDefinition, 0, len(conceptDef.Values)),
            index: make(map[string]int),
        }

        for i, val := range conceptDef.Values {
            info.Values = append(info.Values, AnalysisConceptValueDefinition{
                Name:     val
            })
            info.index[val] = i
        }

        ctx.Concepts[*def.Name] = info
    }

    return nil
}
```

**Analysis with Handler-Based Walk**

The analysis phase uses a single `WalkSymbolTableWithHandlers` call with a map of scope-specific handlers. Each handler processes its scope type as the tree is traversed:

```go
// CPLAnalyzer orchestrates the analysis phase.
type CPLAnalyzer struct {
    concepts map[string]*AnalysisConceptDefinition
    // Future: units, metrics, aggregators, etc.
}

// NewCPLAnalyzer creates a new analyzer with initialized data structures.
func NewCPLAnalyzer() *CPLAnalyzer {
    return &CPLAnalyzer{
        concepts: make(map[string]*AnalysisConceptDefinition),
    }
}

// AnalyzeConceptsScope processes a ConceptsScope table during the walk.
// This handler is called by WalkSymbolTableWithHandlers for each ConceptsScope encountered.
func (a *CPLAnalyzer) AnalyzeConceptsScope(table *SymbolTable) error {
    for _, def := range table.Definitions {
        if def.DefinitionType != ConceptScopeDefinition {
            continue
        }

        parserDef, ok := def.Data.(*ParserConceptDefinition)
        if !ok {
            return fmt.Errorf("invalid concept definition data for %s", *def.Name)
        }

        if _, exists := a.concepts[*def.Name]; exists {
            return fmt.Errorf("duplicate concept definition: %s", *def.Name)
        }

        analysisDef := &AnalysisConceptDefinition{
            Name:       *def.Name,
            Generators: make([]Generator, 0, len(parserDef.Generators)),
            Values:     make([]AnalysisConceptValueDefinition, 0),
            index: make(map[string]int),
        }

        // Convert parser generators to analysis generators and generate values
        for _, parserGen := range parserDef.Generators {
            gen, err := CreateGenerator(parserGen)
            if err != nil {
                return fmt.Errorf("failed to create generator for concept %s: %w", *def.Name, err)
            }
            analysisDef.Generators = append(analysisDef.Generators, gen)

            // Generate values (for Increment 1, only ValuesGenerator is supported)
            values, err := gen.GenerateValues(a.concepts)
            if err != nil {
                return fmt.Errorf("failed to generate values for concept %s: %w", *def.Name, err)
            }

            // Add values to the concept, tracking position
            for _, val := range values {
                if _, exists := analysisDef.index[val]; exists {
                    return fmt.Errorf("duplicate value '%s' in concept '%s'", val, *def.Name)
                }
                pos := len(analysisDef.Values)
                analysisDef.Values = append(analysisDef.Values, AnalysisConceptValueDefinition{
                    Name:     val,
                    Position: pos,
                })
                analysisDef.index[val] = pos
            }
        }

        a.concepts[*def.Name] = analysisDef
    }
    return nil
}

// GetAnalysisDefinitions returns the completed analysis definitions.
func (a *CPLAnalyzer) GetAnalysisDefinitions() *AnalysisDefinitions {
    return &AnalysisDefinitions{Concepts: a.concepts}
}
```

**Usage in cpl_compiler.go**

```go
func (c *CPLCompiler) Analyze(rootTable *SymbolTable) (*AnalysisDefinitions, error) {
    analyzer := NewCPLAnalyzer()

    // Register handlers for each scope type
    // WalkSymbolTableWithHandlers calls the appropriate handler for each scope
    handlers := map[ScopeType]AnalysisFunc{
        ConceptsScope: analyzer.AnalyzeConceptsScope,
        // Future handlers:
        // UnitsScope: analyzer.AnalyzeUnitsScope,
        // MetricsScope: analyzer.AnalyzeMetricsScope,
    }

    // Single walk with all handlers
    if err := WalkSymbolTableWithHandlers(rootTable, handlers); err != nil {
        return nil, fmt.Errorf("analysis failed: %w", err)
    }

    return analyzer.GetAnalysisDefinitions(), nil
}
```

As more definition types are added in subsequent increments (metrics, aggregators, spreaders, etc.), their scope handlers are registered in the `handlers` map alongside `AnalyzeConceptsScope`.

**New Scope Type for Concepts**

The `ScopeType` enumeration in the symbol table package needs a new entry for concepts:

```go
const (
    // GlobalScope is the root scope of the CPL document.
    GlobalScope ScopeType = iota
    // UnitsScope is the scope containing unit declarations.
    UnitsScope
    // ConceptsScope is the scope containing concept declarations.
    ConceptsScope  // NEW: For Increment 1
)
```

**Validation During Analysis**

- **Duplicate concept definitions are an error** - Each concept name must be unique across the compilation unit
- Value uniqueness within each concept
- No duplicate values in List generators

**Limitations**

- Flat concept namespace (no hierarchy)
- Cannot validate parent-child relationships
- No support for tensor shape computation from hierarchical dimensions
- No multiple concept declaration sections in the same file
- No conditional specifications
- No Range generators
- No before/after clauses for value ordering

**Symbol Table Walking Notes**

The analysis phase uses a single `WalkSymbolTableWithHandlers` call with a map of scope-specific handlers. This approach:

- Centralizes tree-walking logic in the symboltable package
- Allows different scope types to be processed in a single tree walk
- Makes it easy to add new scope type handlers in future increments
- Keeps analysis functions focused on their specific domain (concepts, metrics, etc.)
- Each handler is called when a scope of its registered type is encountered during the walk

### Increment 2: Concept Hierarchy

**Enhanced AnalysisConceptDefinition with Hierarchy**

```go
type AnalysisConceptDefinition struct {
    Name            string
    Generators      []Generator         // Analysis-phase generators
    Values          []AnalysisConceptValueDefinition
    index      map[string]int
    ParentName      string              // NEW: Empty if root concept (computed from parser's IndentationLevel)
    ChildNames      []string            // NEW: Names of child concepts
    Level           int                 // NEW: Hierarchy level (0 = root)
}

type ConceptHierarchy struct {
    RootConcepts []string               // Names of root-level concepts
    AllPaths     map[string][]string    // Concept -> path from root
}
```

**Hierarchy Building During WalkSymbolTable**

The hierarchy is built during `WalkSymbolTable` by tracking the concept stack as we traverse:

```go
// ConceptAnalyzer holds state during symbol table walking for hierarchy tracking.
type ConceptAnalyzer struct {
    concepts    map[string]*AnalysisConceptDefinition
    hierarchy   *ConceptHierarchy
    conceptStack []*AnalysisConceptDefinition  // Tracks current parent chain
}

// NewConceptAnalyzer creates a new analyzer with initialized data structures.
func NewConceptAnalyzer() *ConceptAnalyzer {
    return &ConceptAnalyzer{
        concepts:     make(map[string]*AnalysisConceptDefinition),
        hierarchy: &ConceptHierarchy{
            RootConcepts: make([]string, 0),
            AllPaths:     make(map[string][]string),
        },
        conceptStack: make([]*AnalysisConceptDefinition, 0),
    }
}

// AnalyzeConceptsScope processes concept definitions from a ConceptsScope.
// This function is registered with WalkSymbolTable and called for each
// ConceptsScope encountered during the walk.
func (ca *ConceptAnalyzer) AnalyzeConceptsScope(table *SymbolTable) error {
    // Adjust stack based on indentation level
    targetLevel := table.IndentationLevel

    // Pop stack until we reach the correct level
    for len(ca.conceptStack) > targetLevel {
        ca.conceptStack = ca.conceptStack[:len(ca.conceptStack)-1]
    }

    // Get parent from stack (if any)
    var parentName string
    var parent *AnalysisConceptDefinition
    if len(ca.conceptStack) > 0 {
        parent = ca.conceptStack[len(ca.conceptStack)-1]
        parentName = parent.Name
    }

    // Process each concept definition in this scope
    for _, def := range table.Definitions {
        if def.DefinitionType != ConceptScopeDefinition {
            continue
        }

        parserDef, ok := def.Data.(*ParserConceptDefinition)
        if !ok {
            return fmt.Errorf("invalid concept definition data for %s", *def.Name)
        }

        // Check for duplicates
        if _, exists := ca.concepts[*def.Name]; exists {
            return fmt.Errorf("duplicate concept definition: %s", *def.Name)
        }

        // Create analysis definition with hierarchy info
        analysisDef := &AnalysisConceptDefinition{
            Name:       *def.Name,
            Values:     make([]AnalysisConceptValueDefinition, 0, len(parserDef.Values)),
            index: make(map[string]int),
            ParentName: parentName,
            Level:      targetLevel,
        }

        // Populate values
        for i, val := range parserDef.Values {
            analysisDef.Values = append(analysisDef.Values, AnalysisConceptValueDefinition{
                Name:     val,
                Position: i,
            })
            analysisDef.index[val] = i
        }

        // Register in hierarchy
        ca.concepts[*def.Name] = analysisDef

        if parentName == "" {
            // Root concept
            ca.hierarchy.RootConcepts = append(ca.hierarchy.RootConcepts, *def.Name)
            ca.hierarchy.AllPaths[*def.Name] = []string{*def.Name}
        } else {
            // Child concept
            parent.ChildNames = append(parent.ChildNames, *def.Name)
            parentPath := ca.hierarchy.AllPaths[parentName]
            ca.hierarchy.AllPaths[*def.Name] = append(parentPath, *def.Name)
        }

        // Push onto stack for potential children
        ca.conceptStack = append(ca.conceptStack, analysisDef)
    }

    return nil
}

// GetResults returns the analyzed concepts and hierarchy.
func (ca *ConceptAnalyzer) GetResults() (map[string]*AnalysisConceptDefinition, *ConceptHierarchy) {
    return ca.concepts, ca.hierarchy
}
```

**Integration with WalkSymbolTable**

```go
func (c *CPLCompiler) Analyze(rootTable *SymbolTable) (*AnalysisDefinitions, error) {
    // Create the concept analyzer
    conceptAnalyzer := NewConceptAnalyzer()

    // Register the ConceptsScope handler
    handlers := map[ScopeType]AnalysisFunc{
        ConceptsScope: func(table *SymbolTable) error {
            return conceptAnalyzer.AnalyzeConceptsScope(table)
        },
        // Future handlers for other scope types go here:
        // UnitsScope: AnalyzeUnitsScope,
        // MetricsScope: AnalyzeMetricsScope,
    }

    // Single walk with all handlers
    if err := WalkSymbolTableWithHandlers(rootTable, handlers); err != nil {
        return nil, fmt.Errorf("analysis failed: %w", err)
    }

    // Validate hierarchy (detect cycles)
    concepts, hierarchy := conceptAnalyzer.GetResults()
    if err := validateHierarchy(concepts); err != nil {
        return nil, err
    }

    return &AnalysisDefinitions{
        Concepts:  concepts,
        Hierarchy: hierarchy,
    }, nil
}
```

**Validation During Analysis**

- Parent concept must exist
- No cycles in parent-child relationships
- Each concept has at most one parent
- Tree structure consistency

**Limitations**

- No multiple concept declaration sections in the same file
- Child values are global (not conditional on parent)
- No conditional specifications
- No Range generators
- No before/after value ordering
- Aggregation/spreader matrices cannot be computed (need child-per-parent values)

### Increment 3: Multiple Concept Declaration Sections in the Same File

**Overview**

Increment 3 allows multiple `concepts` sections in the same CPL file. Since each section creates its own `ConceptsScope` symbol table under the GlobalScope, the existing `AnalyzeConceptsScope` function automatically processes all concept definitions from all sections during the single `WalkSymbolTableWithHandlers` traversal.

**No Special Merging Required**

Unlike what was originally planned, there is no need for a separate `BuildConceptsFromMultipleSections` or `mergeConceptDefinitions` function. The symbol table walk naturally visits each `ConceptsScope` table, and the `AnalyzeConceptsScope` handler accumulates all concept definitions into the analyzer's concepts map.

**Duplicate Detection**

The existing duplicate check in `AnalyzeConceptsScope` (from Increment 1) is sufficient:

```go
if _, exists := ca.concepts[*def.Name]; exists {
    return fmt.Errorf("duplicate concept definition: %s", *def.Name)
}
```

Since the symbol table walk visits all `ConceptsScope` tables (regardless of which section they came from), any attempt to define the same concept name in multiple sections will be caught as a duplicate.

**Validation During Analysis**

- **Duplicate concept definitions across sections are an error** - The existing duplicate check in `AnalyzeConceptsScope` catches these
- No other cross-section validation is needed for Increment 3

**Limitations**

- No Range generators
- No conditional specifications
- No before/after clauses for value ordering

### Increment 4: Conditional Specifications

**Enhanced AnalysisConceptDefinition with Conditional Values**

```go
type AnalysisConceptDefinition struct {
    Name              string
    Generators        []Generator
    Values            []AnalysisConceptValueDefinition
    index        map[string]int
    ParentName        string
    ChildNames        []string
    Level             int
    ConditionalValues map[string][]string  // NEW: Parent value -> child values
    DefaultValues     []string             // NEW: When no condition matches
}

```

**Conditional Value Resolution**

```go
func ResolveConditionalValues(
    childDef *ParserConceptDefinition,
    parentConcept *AnalysisConceptDefinition) (*AnalysisConceptDefinition, error) {

    info := &AnalysisConceptDefinition{
        Name:              childDef.Name,
        ParentName:        childDef.ParentName,
        ConditionalValues: make(map[string][]string),
    }

    // Process conditional specifications
    for _, spec := range childDef.ConditionalSpecs {
        // Validate parent value exists
        if _, exists := parentConcept.index[spec.ParentValue]; !exists {
            return nil, fmt.Errorf("invalid parent value: %s", spec.ParentValue)
        }

        // Expand values from generator
        values := expandGeneratorValues(spec)
        info.ConditionalValues[spec.ParentValue] = values
    }

    // Process default values (if no conditions match)
    if len(childDef.DefaultValues) > 0 {
        info.DefaultValues = childDef.DefaultValues
    }

    // Build unified value list (union of all conditional values + defaults)
    info.Values = buildUnifiedValueList(info)

    return info, nil
}

func (info *AnalysisConceptDefinition) GetValuesForParent(parentValue string) []string {
    if values, exists := info.ConditionalValues[parentValue]; exists {
        return values
    }
    return info.DefaultValues
}
```

**Validation During Analysis**

- Parent values in conditions must exist
- Child values in conditions must be valid
- At least one value set (conditional or default) must be defined

**Limitations**

- No Range generators
- No before/after clauses for value ordering
- Complex conditional expressions not supported

### Increment 5: Int Range Generators

**Overview**

Increment 5 adds support for Int Range generators. Int Range generators produce numeric values from a start to end bound with an optional step. The values are ordered from start to end.

**Generator Support**

In Increment 5, the analysis phase adds support for `IntRangeGenerator` in addition to `ValuesGenerator`:

```go
// IntRangeGenerator generates integer values from start to end with a step.
type IntRangeGenerator struct {
    Start int
    End   int
    Step  int  // Defaults to 1 if not specified in parser
}

func (i *IntRangeGenerator) GenerateValues(concepts map[string]AnalysisConceptDefinition) ([]string, error) {
    if i.Step <= 0 {
        return nil, fmt.Errorf("int range step must be positive, got %d", i.Step)
    }
    if i.Start > i.End {
        return nil, fmt.Errorf("int range start (%d) must be <= end (%d)", i.Start, i.End)
    }

    var values []string
    for v := i.Start; v <= i.End; v += i.Step {
        values = append(values, strconv.Itoa(v))
    }
    return values, nil
}
```

**CreateGenerator Extension for Increment 5**

The `CreateGenerator` function is extended to support `IntRangeGeneratorDef`:

```go
// CreateGenerator converts a parser generator definition to an analysis generator.
// For Increment 5, supports ValuesGeneratorDef and IntRangeGeneratorDef.
func CreateGenerator(parserGen interface{}) (Generator, error) {
    switch gen := parserGen.(type) {
    case parser.ValuesGeneratorDef:
        return &ValuesGenerator{Values: gen.Values}, nil
    case parser.IntRangeGeneratorDef:
        return &IntRangeGenerator{
            Start: gen.Start,
            End:   gen.End,
            Step:  gen.Step,
        }, nil
    default:
        return nil, fmt.Errorf("unknown generator type: %T", parserGen)
    }
}
```

**Value Generation**

During analysis, Int Range generators are expanded into their full set of values using the same `GenerateValues` mechanism as Values generators. The expanded values are stored in `AnalysisConceptDefinition.Values`. No separate tracking of the range origin is needed in the analysis structure.

**Value Ordering**

Value order is maintained from the declaration:

- For Values generators: order is as declared
- For Int Range generators: order is from Start to End by Step
- Multiple generators: values appear in the order generators are declared

**Validation During Analysis**

- Int range bounds are valid integers with start <= end
- Step must be a positive integer
- Range bounds must be non-negative

**Limitations**

- No Concept Range generators (added in Increment 6)
- No before/after clauses for value ordering when extending concepts (added in Increment 7)
- No integration with external Concepts API (added in Increment 8)
- Concept versions are not tracked for runtime validation (added in Increment 8)

### Increment 6: Concept Range Generators

**Overview**

Increment 6 adds support for Concept Range generators. Concept Range generators reference a range of values from another concept, enabling composition and reuse of existing concept definitions.

**Generator Support**

In Increment 6, the analysis phase adds support for `ConceptRangeGenerator` in addition to `ValuesGenerator` and `IntRangeGenerator`:

```go
// ConceptRangeGenerator generates values by referencing a range from another concept.
type ConceptRangeGenerator struct {
    ConceptName string
    StartValue  string
    EndValue    string
}

func (c *ConceptRangeGenerator) GenerateValues(concepts map[string]AnalysisConceptDefinition) ([]string, error) {
    // Look up the referenced concept
    sourceConcept, exists := concepts[c.ConceptName]
    if !exists {
        return nil, fmt.Errorf("referenced concept not found: %s", c.ConceptName)
    }

    // Find start and end value positions
    startIdx, startExists := sourceConcept.index[c.StartValue]
    if !startExists {
        return nil, fmt.Errorf("start value '%s' not found in concept '%s'", c.StartValue, c.ConceptName)
    }

    endIdx, endExists := sourceConcept.index[c.EndValue]
    if !endExists {
        return nil, fmt.Errorf("end value '%s' not found in concept '%s'", c.EndValue, c.ConceptName)
    }

    if startIdx > endIdx {
        return nil, fmt.Errorf("start value '%s' comes after end value '%s' in concept '%s'",
            c.StartValue, c.EndValue, c.ConceptName)
    }

    // Extract values from start to end (inclusive)
    var values []string
    for i := startIdx; i <= endIdx; i++ {
        values = append(values, sourceConcept.Values[i].Name)
    }

    return values, nil
}
```

**CreateGenerator Extension for Increment 6**

The `CreateGenerator` function is extended to support `ConceptRangeGeneratorDef`:

```go
// CreateGenerator converts a parser generator definition to an analysis generator.
// For Increment 6, supports ValuesGeneratorDef, IntRangeGeneratorDef, and ConceptRangeGeneratorDef.
func CreateGenerator(parserGen interface{}) (Generator, error) {
    switch gen := parserGen.(type) {
    case parser.ValuesGeneratorDef:
        return &ValuesGenerator{Values: gen.Values}, nil
    case parser.IntRangeGeneratorDef:
        return &IntRangeGenerator{
            Start: gen.Start,
            End:   gen.End,
            Step:  gen.Step,
        }, nil
    case parser.ConceptRangeGeneratorDef:
        return &ConceptRangeGenerator{
            ConceptName: gen.ConceptName,
            StartValue:  gen.StartValue,
            EndValue:    gen.EndValue,
        }, nil
    default:
        return nil, fmt.Errorf("unknown generator type: %T", parserGen)
    }
}
```

**Concept Range Resolution Algorithm**

Concept Range generators require special handling because they reference other concepts that must be fully resolved before the range can be extracted. The resolution algorithm works as follows:

```go
// ResolveConceptRanges processes ConceptRangeGenerators in the correct order.
// It must be called after all concepts have been initially processed but before
// the analysis phase completes.
func (a *CPLAnalyzer) ResolveConceptRanges() error {
    // Track which concepts have ConceptRangeGenerators that need resolution
    type pendingRange struct {
        conceptName string
        generator   *ConceptRangeGenerator
        generatorIdx int
    }
    var pending []pendingRange

    // Collect all ConceptRangeGenerators from all concepts
    for conceptName, concept := range a.concepts {
        for idx, gen := range concept.Generators {
            if crg, ok := gen.(*ConceptRangeGenerator); ok {
                pending = append(pending, pendingRange{
                    conceptName:  conceptName,
                    generator:    crg,
                    generatorIdx: idx,
                })
            }
        }
    }

    // Process ConceptRangeGenerators in rounds until all are resolved
    maxIterations := len(pending) + 1  // Prevent infinite loops
    for iteration := 0; iteration < maxIterations && len(pending) > 0; iteration++ {
        var stillPending []pendingRange

        for _, p := range pending {
            concept := a.concepts[p.conceptName]

            // Check if the referenced concept is resolved
            sourceConcept, exists := a.concepts[p.generator.ConceptName]
            if !exists {
                return fmt.Errorf("concept '%s' references unknown concept '%s'",
                    p.conceptName, p.generator.ConceptName)
            }

            // Check if all values in the source concept are resolved (no pending ranges)
            sourceHasPendingRanges := false
            for _, gen := range sourceConcept.Generators {
                if _, isCRG := gen.(*ConceptRangeGenerator); isCRG {
                    // Check if this CRG has been resolved yet
                    if !isConceptRangeResolved(sourceConcept, gen.(*ConceptRangeGenerator)) {
                        sourceHasPendingRanges = true
                        break
                    }
                }
            }

            if sourceHasPendingRanges {
                // Source concept still has unresolved ranges, defer this one
                stillPending = append(stillPending, p)
                continue
            }

            // Source concept is fully resolved, resolve this range
            values, err := p.generator.GenerateValues(a.concepts)
            if err != nil {
                return fmt.Errorf("failed to resolve concept range for '%s': %w", p.conceptName, err)
            }

            // Add the generated values to the concept
            for _, val := range values {
                if _, exists := concept.index[val]; exists {
                    return fmt.Errorf("duplicate value '%s' in concept '%s' from concept range",
                        val, p.conceptName)
                }
                pos := len(concept.Values)
                concept.Values = append(concept.Values, AnalysisConceptValueDefinition{
                    Name:     val,
                    Position: pos,
                })
                concept.index[val] = pos
            }
        }

        pending = stillPending
    }

    if len(pending) > 0 {
        // There are circular dependencies among concept ranges
        var names []string
        for _, p := range pending {
            names = append(names, fmt.Sprintf("%s->%s", p.conceptName, p.generator.ConceptName))
        }
        return fmt.Errorf("circular dependency detected in concept ranges: %v", names)
    }

    return nil
}

// isConceptRangeResolved checks if a ConceptRangeGenerator has been resolved.
// A range is resolved if all values it would generate are already present in the concept.
func isConceptRangeResolved(concept *AnalysisConceptDefinition, crg *ConceptRangeGenerator) bool {
    // This is a simplified check - in practice, you might track resolved status explicitly
    _, startExists := concept.index[crg.StartValue]
    _, endExists := concept.index[crg.EndValue]
    return startExists && endExists
}
```

**Alternative Resolution Algorithm (Simpler Approach)**

A simpler algorithm that processes generators in order and defers ConceptRangeGenerators until their source concepts are fully resolved:

```go
// GenerateConceptValues generates values for a concept by processing its generators.
// It handles ConceptRangeGenerators by deferring them until the source concept is resolved.
func (a *CPLAnalyzer) GenerateConceptValues(concept *AnalysisConceptDefinition) error {
    // First pass: process non-ConceptRange generators in order
    var deferredCRGs []*ConceptRangeGenerator

    for _, gen := range concept.Generators {
        if crg, ok := gen.(*ConceptRangeGenerator); ok {
            deferredCRGs = append(deferredCRGs, crg)
            continue
        }

        // Process non-CRG generators immediately
        values, err := gen.GenerateValues(a.concepts)
        if err != nil {
            return fmt.Errorf("failed to generate values for concept '%s': %w", concept.Name, err)
        }

        for _, val := range values {
            if _, exists := concept.index[val]; exists {
                return fmt.Errorf("duplicate value '%s' in concept '%s'", val, concept.Name)
            }
            pos := len(concept.Values)
            concept.Values = append(concept.Values, AnalysisConceptValueDefinition{
                Name:     val,
                Position: pos,
            })
            concept.index[val] = pos
        }
    }

    // Store deferred CRGs for later resolution
    concept.pendingCRGs = deferredCRGs
    return nil
}

// ResolveAllConceptRanges processes all deferred ConceptRangeGenerators across all concepts.
// It iterates until all ranges are resolved or a circular dependency is detected.
func (a *CPLAnalyzer) ResolveAllConceptRanges() error {
    maxIterations := len(a.concepts) + 1

    for iteration := 0; iteration < maxIterations; iteration++ {
        allResolved := true

        for _, concept := range a.concepts {
            if len(concept.pendingCRGs) == 0 {
                continue
            }

            var stillPending []*ConceptRangeGenerator

            for _, crg := range concept.pendingCRGs {
                // Check if source concept is fully resolved
                sourceConcept, exists := a.concepts[crg.ConceptName]
                if !exists {
                    return fmt.Errorf("concept '%s' references unknown concept '%s'",
                        concept.Name, crg.ConceptName)
                }

                if len(sourceConcept.pendingCRGs) > 0 {
                    // Source concept still has pending ranges
                    stillPending = append(stillPending, crg)
                    allResolved = false
                    continue
                }

                // Source is resolved, generate values from this range
                values, err := crg.GenerateValues(a.concepts)
                if err != nil {
                    return fmt.Errorf("failed to resolve concept range for '%s': %w", concept.Name, err)
                }

                for _, val := range values {
                    if _, exists := concept.index[val]; exists {
                        return fmt.Errorf("duplicate value '%s' in concept '%s' from concept range",
                            val, concept.Name)
                    }
                    pos := len(concept.Values)
                    concept.Values = append(concept.Values, AnalysisConceptValueDefinition{
                        Name:     val,
                        Position: pos,
                    })
                    concept.index[val] = pos
                }
            }

            concept.pendingCRGs = stillPending
        }

        if allResolved {
            return nil
        }
    }

    return fmt.Errorf("circular dependency detected among concept range generators")
}
```

**Value Ordering**

Value order is maintained from the declaration:

- For Values generators: order is as declared
- For Int Range generators: order is from Start to End by Step
- For Concept Range generators: order follows the referenced concept's ordering from StartValue to EndValue
- Multiple generators: values appear in the order generators are declared

**Validation During Analysis**

- Referenced concept must exist
- Start and end values must exist in the referenced concept
- Start value must come before or at the same position as end value in the referenced concept
- No circular dependencies among concept range generators
- Duplicate detection across all generators

**Limitations**

- No before/after clauses for value ordering when extending concepts (added in Increment 7)
- No integration with external Concepts API (added in Increment 8)
- Concept versions are not tracked for runtime validation (added in Increment 8)

### Increment 7: Before/After Clauses for Value Ordering

**Overview**

Increment 7 adds support for before/after clauses when extending concepts. This allows new values to be inserted at specific positions rather than only appending.

**Important Notes:**

- Value order is determined at initial concept declaration and cannot be changed after
- Before/after clauses only apply when adding new values to an existing concept
- The initial declaration establishes the canonical order
- New values inserted between original values change the absolute positions but preserve the relative order of the original values

**Value Ordering Computation**

```go
// ComputeValueOrdering applies before/after insertions to a base value list.
// It is called when processing concept extensions in secondary concept sections.
func ComputeValueOrdering(
    baseValues []AnalysisConceptValueDefinition,
    newValues []string,
    insertionPoint InsertionPoint) ([]AnalysisConceptValueDefinition, error) {

    // Create a working copy of base values
    ordered := make([]AnalysisConceptValueDefinition, len(baseValues))
    copy(ordered, baseValues)

    var insertPos int

    if insertionPoint.BeforeValue != "" {
        // Find position of "before" value
        pos, exists := findValuePosition(ordered, insertionPoint.BeforeValue)
        if !exists {
            return nil, fmt.Errorf("'before' value not found: %s", insertionPoint.BeforeValue)
        }
        insertPos = pos
    } else if insertionPoint.AfterValue != "" {
        // Find position after "after" value
        pos, exists := findValuePosition(ordered, insertionPoint.AfterValue)
        if !exists {
            return nil, fmt.Errorf("'after' value not found: %s", insertionPoint.AfterValue)
        }
        insertPos = pos + 1
    } else {
        // No insertion point specified - append to end
        insertPos = len(ordered)
    }

    // Insert the new values
    for i, val := range newValues {
        // Check for duplicates
        if _, exists := findValuePosition(ordered, val); exists {
            return nil, fmt.Errorf("value '%s' already exists in concept", val)
        }

        // Insert at the computed position
        newValDef := AnalysisConceptValueDefinition{
            Name:     val,
            Position: insertPos + i,
        }
        ordered = insertAt(ordered, insertPos+i, newValDef)
    }

    // Recompute positions for all values
    for i := range ordered {
        ordered[i].Position = i
    }

    return ordered, nil
}

func findValuePosition(values []AnalysisConceptValueDefinition, name string) (int, bool) {
    for i, v := range values {
        if v.Name == name {
            return i, true
        }
    }
    return -1, false
}

func insertAt(slice []AnalysisConceptValueDefinition, index int, value AnalysisConceptValueDefinition) []AnalysisConceptValueDefinition {
    if index >= len(slice) {
        return append(slice, value)
    }
    slice = append(slice[:index+1], slice[index:]...)
    slice[index] = value
    return slice
}
```

**Integration with Generator Processing**

When processing concept extensions with before/after clauses:

1. First, generate values from the extension's generators (using the standard generator processing)
2. Then, apply the before/after insertion to integrate the new values into the existing concept

```go
func (a *CPLAnalyzer) ProcessConceptExtension(
    existingConcept *AnalysisConceptDefinition,
    extensionDef *ParserConceptDefinition) error {

    // Generate new values from extension generators
    var newValues []string
    for _, parserGen := range extensionDef.Generators {
        gen, err := CreateGenerator(parserGen)
        if err != nil {
            return fmt.Errorf("failed to create generator: %w", err)
        }

        values, err := gen.GenerateValues(a.concepts)
        if err != nil {
            return fmt.Errorf("failed to generate values: %w", err)
        }
        newValues = append(newValues, values...)
    }

    // Apply before/after insertion
    updatedValues, err := ComputeValueOrdering(
        existingConcept.Values,
        newValues,
        extensionDef.InsertionPoint,
    )
    if err != nil {
        return fmt.Errorf("failed to compute value ordering: %w", err)
    }

    // Update the concept with new ordering
    existingConcept.Values = updatedValues

    // Rebuild the index map
    existingConcept.index = make(map[string]int)
    for i, val := range updatedValues {
        existingConcept.index[val.Name] = i
    }

    return nil
}
```

**Validation During Analysis**

- Before/after values must exist in the target concept
- Cannot reorder existing values - before/after only applies to new values being added
- No duplicate insertions at same position
- Extension before/after values must exist in merged concept
- At least one of `before` or `after` must be specified

**Limitations**

- Cannot reorder existing values - before/after only applies to new values being added
- No integration with external Concepts API (added in Increment 8)
- Concept versions are not tracked for runtime validation (added in Increment 8)

### Increment 8: Concepts API Integration

**Loading API Concepts into the Analyzer**

API concepts are loaded into the `ConceptAnalyzer` before processing file concept sections. They are treated exactly like concepts defined in earlier concept sections within the same compilation unit:

```go
// LoadAPIConcepts fetches concepts from the API and populates
// the analyzer's concepts map before file processing begins.
func (ca *ConceptAnalyzer) LoadAPIConcepts(apiConcepts []*APIConceptDefinition) error {
    for _, apiDef := range apiConcepts {
        // Check for duplicates - API concepts follow same rules as file concepts
        if _, exists := ca.concepts[apiDef.Name]; exists {
            return fmt.Errorf("duplicate concept definition from API: %s", apiDef.Name)
        }

        info, err := ca.buildConceptFromAPIDefinition(apiDef)
        if err != nil {
            return fmt.Errorf("failed to build concept %s from API: %w", apiDef.Name, err)
        }

        ca.concepts[apiDef.Name] = info

        // Track in hierarchy if root concept
        if info.ParentName == "" {
            ca.hierarchy.RootConcepts = append(ca.hierarchy.RootConcepts, apiDef.Name)
            ca.hierarchy.AllPaths[apiDef.Name] = []string{apiDef.Name}
        }
    }
    return nil
}
```

**Same Extension Rules Apply**

Once API concepts are loaded into the analyzer, file concept sections follow the same extension rules as multiple concept sections in the same file:

- File concepts can extend API concepts by adding new values
- File concepts can declare conditional values for child concepts
- Cannot redefine existing values or change parent-child relationships
- Duplicate value detection applies across API and file sources

```go
func (c *CPLCompiler) Analyze(rootTable *SymbolTable, apiConcepts []*APIConceptDefinition) (*AnalysisDefinitions, error) {
    conceptAnalyzer := NewConceptAnalyzer()

    // Step 1: Load API concepts first (like earlier concept sections)
    if err := conceptAnalyzer.LoadAPIConcepts(apiConcepts); err != nil {
        return nil, fmt.Errorf("failed to load API concepts: %w", err)
    }

    // Step 2: Process file concept sections (can extend API concepts)
    handlers := map[ScopeType]AnalysisFunc{
        ConceptsScope: func(table *SymbolTable) error {
            return conceptAnalyzer.AnalyzeConceptsScope(table)
        },
    }

    if err := WalkSymbolTableWithHandlers(rootTable, handlers); err != nil {
        return nil, fmt.Errorf("analysis failed: %w", err)
    }

    // Validate hierarchy and extract results
    concepts, hierarchy := conceptAnalyzer.GetResults()
    if err := validateHierarchy(concepts); err != nil {
        return nil, err
    }

    return &AnalysisDefinitions{
        Concepts:  concepts,
        Hierarchy: hierarchy,
    }, nil
}
```

**Version Extraction for Synthesis**

The version tracked for synthesis-time consistency checking is a hash of the concept's definition (values and structure):

```go
func ExtractConceptVersions(
    definitions *AnalysisDefinitions) map[string]ConceptVersion {
    versions := make(map[string]ConceptVersion)

    for name, info := range definitions.Concepts {
        versions[name] = ConceptVersion{
            Name:         name,
            VersionHash:  computeValueHash(info.Values),
        }
    }

    return versions
}

type ConceptVersion struct {
    Name        string
    VersionHash string  // Hash of concept values for synthesis-time consistency checking
}
```

**Validation During Analysis**

- API concepts must be successfully fetched before analysis
- File concepts can extend but not fundamentally alter API concept structure
- Same extension rules as multiple concept sections in the same file
- Concept version (hash) is computed for synthesis-time consistency checking
- Duplicate value detection across API and file sources

**Limitations**

None - Full concept analysis functionality with API integration.

## Approach

### Incremental Structure and Function Evolution

The analysis phase uses an incremental approach where structures and functions evolve by adding fields and capabilities as each increment requires them:

**Structure Evolution:**

- **Increment 1:** `AnalysisConceptDefinition` starts with basic fields: `Name`, `Generators`, `Values`, `index`
- **Increment 2:** Adds `ParentName`, `ChildNames`, `Level` for hierarchy support
- **Increment 3:** No new fields - uses existing duplicate detection across multiple sections
- **Increment 4:** Adds `ConditionalValues`, `DefaultValues` for conditional specifications
- **Increment 5:** No new fields - adds `IntRangeGenerator` implementation; Int range values are expanded into `Values` during analysis
- **Increment 6:** No new fields - adds `ConceptRangeGenerator` implementation with resolution algorithm
- **Increment 7:** No new fields - adds `ComputeValueOrdering` for before/after clauses
- **Increment 8:** No new fields - API concepts loaded into existing structures

**Function Evolution:**

- **Increment 1:** Basic `AnalyzeConceptsScope` processes flat concept definitions; `Generator` interface and `ValuesGenerator` defined
- **Increment 2:** Enhanced with hierarchy tracking using a concept stack
- **Increment 3:** No changes needed - existing duplicate detection handles multiple sections
- **Increment 4:** Enhanced to process conditional specifications
- **Increment 5:** Enhanced with `IntRangeGenerator` and `CreateGenerator` factory function
- **Increment 6:** Enhanced with `ConceptRangeGenerator` and `ResolveConceptRanges` resolution algorithm
- **Increment 7:** Enhanced with `ComputeValueOrdering` for before/after clauses
- **Increment 8:** Enhanced with `LoadAPIConcepts` to pre-populate concepts from external sources

**No Feature Flags:**

There are no feature flags or runtime checks for increment capabilities. Each increment's code uses only the fields and functions available at that point in development.

**Minimal Nil Values:**

Fields are only added when needed, and nil values are used only when the final system design requires optional fields (e.g., a concept may legitimately have no parent). There are no "placeholder" nil fields waiting for future increments.

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

1. **Increment 1:** Basic concept validation, value uniqueness, Generator interface definition
2. **Increment 2:** Hierarchy validation, tree structure, cycle detection
3. **Increment 3:** Duplicate detection across multiple concept sections
4. **Increment 4:** Conditional spec validation, parent value existence
5. **Increment 5:** Int range bounds validation (start <= end, positive step), range value expansion
6. **Increment 6:** Concept range resolution, circular dependency detection, referenced concept existence
7. **Increment 7:** Before/after resolution, value ordering validation, insertion point existence
8. **Increment 8:** API concept loading, version hash computation, API/file concept merging

## Notes

**Related Documents:**

- [Concepts Incremental Implementation](../../eros/ERO001_concepts-incremental-implementation.md) - Main incremental implementation ERO
- [Concept Definitions Design](../../concepts-definitions.md) - Full concept design
- [Intermediate Language](../../architecture/cpl/analysis/intermediate_language.md) - CIL specification
