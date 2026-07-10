# Story122: CPL Concepts Analyzer - Conditional

## Goal

Implement conditional value resolution in the CPL analyzer. This story validates parent values in conditional specifications and builds the parent-specific child value mappings. The analyzer ensures parent values referenced in conditions exist and builds a unified value list from conditional and default specifications. This is part of [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md) - Increment 4.

## References

- [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md)
- [ERO001: Concepts Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO001_concepts-analysis-phase.md) - Increment 4 section
- [Concept Definitions Design](../specifications/design/cpl/concepts-definitions.md)

## Dependencies

- [Story121_cpl-concepts-parser-conditional.md](Story121_cpl-concepts-parser-conditional.md) - Not Started - Parser must capture conditional specs

## Dependent Stories

- Story123_cpl-concepts-parser-range.md - Not Started - Range builds on conditional

## Tasks

1. Enhance AnalysisConceptDefinition for conditional values - Completed
   a. Add ConditionalValues map field - Completed
   b. Add DefaultValues slice field - Completed

2. Implement conditional value resolution - Completed
   a. Create ResolveConditionalValues function - Completed
   b. Validate parent values exist - Completed
   c. Expand values from generators - Completed
   d. Build unified value list - Completed

3. Update ConceptAnalyzer - Completed
   a. Process conditional specs during analysis - Completed
   b. Call ResolveConditionalValues for child concepts - Completed
   c. Store resolved values in AnalysisConceptDefinition - Completed

4. Implement validation - Completed
   a. Validate parent value exists in parent concept - Completed
   b. Validate at least one value set defined - Completed
   c. Report errors with source locations - Completed

5. Add helper methods - Completed
   a. GetValuesForParent method - Completed
   b. HasConditionalValues method - Completed

6. Write unit tests for conditional analysis - Completed
   a. Test conditional value resolution - Completed
   b. Test invalid parent value detection - Completed
   c. Test default value fallback - Completed
   d. Test unified value list building - Completed

## Constraints

- Must validate parent values exist in parent concept
- Must support both conditional and default values
- Must build unified value list for dimension computations
- Follow the Logical TDD Lifecycle - write tests before implementation
- Must not break Increment 1-3 functionality

## Intent

The goal of this story is to resolve conditional specifications into parent-specific child value mappings. The analyzer validates that parent values referenced in conditions actually exist and builds the data structures needed for metric dimension validation and tensor shape computation.

## Acceptance Criteria

1. AnalysisConceptDefinition includes ConditionalValues and DefaultValues
2. Parent values in conditions are validated to exist
3. Invalid parent values are reported as errors with source locations
4. Conditional and default values are properly resolved
5. Unified value list is built for each child concept
6. GetValuesForParent helper method works correctly
7. All unit tests pass following BTDD practices
8. Increment 1-3 functionality still works

## Notes

**Increment 4 Capabilities:**

- Parent value validation in conditions
- Parent-specific child value resolution
- Default value fallback
- Unified value list building

**Increment 4 Limitations:**

- No Range generators in conditions (added in Increment 5)
- No before/after clauses (added in Increment 6)
- No API integration (added in Increment 7)

**Example Analysis:**

```cpl
concepts
    Country=[Values[USA, Canada]]
      Region=[Values[North, South]]

    Country[USA] => State=[Values[NY, CA, TX]]
    Country[Canada] => Province=[Values[Ontario, Quebec]]
```

Results in:

- State.ConditionalValues["USA"] = ["NY", "CA", "TX"]
- State.ConditionalValues["Canada"] = ["Ontario", "Quebec"]
- State.Values = ["NY", "CA", "TX", "Ontario", "Quebec"] (unified)

**Enhanced AnalysisConceptDefinition:**

```go
type AnalysisConceptDefinition struct {
    Name              string
    Values            []AnalysisConceptValueDefinition
    ValueIndex        map[string]int
    ParentName        string
    ChildNames        []string
    Level             int
    ConditionalValues map[string][]string  // NEW in Increment 4: Parent value -> child values
    DefaultValues     []string             // NEW in Increment 4
}
```

**Conditional Value Resolution:**

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
        if _, exists := parentConcept.ValueIndex[spec.ParentValue]; !exists {
            return nil, fmt.Errorf("invalid parent value: %s", spec.ParentValue)
        }

        // Expand values from generator
        values := expandGeneratorValues(spec)
        info.ConditionalValues[spec.ParentValue] = values
    }

    // Process default values
    if len(childDef.DefaultValues) > 0 {
        info.DefaultValues = childDef.DefaultValues
    }

    // Build unified value list (union of all conditional values + defaults)
    info.Values = buildUnifiedValueList(info)

    return info, nil
}
```

**Helper Method:**

```go
func (info *AnalysisConceptDefinition) GetValuesForParent(parentValue string) []string {
    if values, exists := info.ConditionalValues[parentValue]; exists {
        return values
    }
    return info.DefaultValues
}
```

**Related Documents:**

- [Concepts Incremental Implementation](../specifications/design/cpl/eros/ERO001_concepts-incremental-implementation.md) - Main incremental implementation ERO
