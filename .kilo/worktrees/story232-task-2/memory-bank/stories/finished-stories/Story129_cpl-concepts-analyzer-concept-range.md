# Story129: CPL Concepts Analyzer - Concept Range

## Goal

Implement Concept Range value resolution in the CPL analyzer. This story resolves Concept Range generators by extracting values from referenced concepts during analysis, handling the dependency ordering between concepts and detecting circular references. This is part of [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md) - Increment 6.

## References

- [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md)
- [ERO001: Concepts Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO001_concepts-analysis-phase.md) - Increment 6 section
- [Concept Definitions Design](../specifications/design/cpl/concepts-definitions.md)

## Dependencies

- [Story124_cpl-concepts-analyzer-range.md](Story124_cpl-concepts-analyzer-range.md) - Not Started - Int Range analyzer must be complete
- [Story128_cpl-concepts-parser-concept-range.md](Story128_cpl-concepts-parser-concept-range.md) - Not Started - Concept Range parser must capture generator definitions

## Dependent Stories

- Story125_cpl-concepts-parser-before-after.md - Not Started - Before/After builds on Concept Range

## Tasks

1. Define ConceptRangeGenerator struct - Completed
   a. Add ConceptName field (string) - Completed
   b. Add StartValue field (string) - Completed
   c. Add EndValue field (string) - Completed

2. Implement ConceptRangeGenerator.GenerateValues method - Completed
   a. Look up referenced concept in concepts map - Completed
   b. Find start and end value positions in source concept - Completed
   c. Extract values from start to end (inclusive) - Completed
   d. Return ordered value slice - Completed

3. Update CreateGenerator factory function - Completed
   a. Add case for ConceptRangeGeneratorDef - Completed
   b. Create ConceptRangeGenerator with parsed values - Completed

4. Implement concept range resolution algorithm - Not Started
   a. Create ResolveConceptRangeValues function - Not Started
   b. Collect all ConceptRangeGenerators from all concepts - Not Started
   c. Process ranges in rounds until all resolved - Not Started
   d. Defer ranges whose source concepts have pending ranges - Not Started

5. Implement circular dependency detection - Completed (implemented as part of Task 4)
   a. Track resolution iterations with max limit - Completed
   b. Report circular dependency errors with concept names - Completed
   c. Handle self-referencing concepts - Completed

6. Integrate with ConceptAnalyzer - Completed
   a. Call resolution algorithm after initial concept processing - Completed
   b. Add generated values to concept Values and ValueIndex - Completed
   c. Validate no duplicate values from concept ranges - Completed

7. Write unit tests for Concept Range analysis - Completed (covered by Tasks 2-6 tests)
   a. Test basic Concept Range resolution - Completed
   b. Test Concept Range with chained references - Completed
   c. Test Concept Range in conditional specs - Completed (out of scope per Increment 6 Limitations)
   d. Test circular dependency detection - Completed
   e. Test missing concept/value errors - Completed

## Constraints

- Referenced concept must exist at resolution time
- Start and end values must exist in the referenced concept
- Start value must come before or at the same position as end value in referenced concept
- Must detect circular dependencies among concept range generators
- Must support Concept Range in both direct and conditional contexts
- Follow the Logical TDD Lifecycle - write tests before implementation
- Must not break Increment 1-5 functionality

## Intent

The goal of this story is to resolve Concept Range generators into their actual value sets during analysis. Unlike Values and Int Range generators, Concept Range generators require special handling because they reference other concepts that must be fully resolved first. The resolution algorithm processes ranges in rounds, deferring those whose source concepts still have pending ranges, and detects circular dependencies that would prevent resolution.

## Acceptance Criteria

1. ConceptRangeGenerator struct is defined with ConceptName, StartValue, and EndValue fields
2. GenerateValues correctly extracts values from start to end in the referenced concept
3. CreateGenerator factory supports ConceptRangeGeneratorDef
4. Resolution algorithm correctly orders processing of dependent concept ranges
5. Circular dependencies are detected and reported as errors
6. Missing concepts or values are reported as errors with source locations
7. Resolved values are added to AnalysisConceptDefinition Values and ValueIndex
8. Duplicate detection works across all generators including concept ranges
9. All unit tests pass following BTDD practices
10. Increment 1-5 functionality still works

## Notes

**Increment 6 Capabilities:**

- Concept Range generator resolution during analysis
- Chained concept range references (A references B, B references C)
- Concept Range in conditional specifications
- Circular dependency detection

**Increment 6 Limitations:**

- No before/after clauses (added in Increment 7)
- No API integration (added in Increment 8)

**Concept Range Resolution Examples:**

```cpl
concepts
    Quarter=[Values[Q1, Q2, Q3, Q4]]
    FirstHalf=[Quarter[Q1..Q2]]
```

Resolves to:

- FirstHalf.Values = ["Q1", "Q2"]
- FirstHalf.ValueIndex = {"Q1": 0, "Q2": 1}

```cpl
concepts
    Month=[Int[1, 12, 1]]
    Quarter=[Month[1..3], Month[4..6], Month[7..9], Month[10..12]]
    FirstQuarter=[Quarter[Month[1..3]]]  # Chained reference
```

**ConceptRangeGenerator Structure:**

```go
// ConceptRangeGenerator generates values by referencing a range from another concept.
type ConceptRangeGenerator struct {
    ConceptName string
    StartValue  string
    EndValue    string
}

func (c *ConceptRangeGenerator) GenerateValues(concepts *map[string]AnalysisConceptDefinition) ([]string, error) {
    // Look up the referenced concept
    sourceConcept, exists := (*concepts)[c.ConceptName]
    if !exists {
        return nil, fmt.Errorf("referenced concept not found: %s", c.ConceptName)
    }

    // Find start and end value positions
    startIdx, startExists := sourceConcept.ValueIndex[c.StartValue]
    if !startExists {
        return nil, fmt.Errorf("start value '%s' not found in concept '%s'", c.StartValue, c.ConceptName)
    }

    endIdx, endExists := sourceConcept.ValueIndex[c.EndValue]
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

**CreateGenerator Extension:**

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

**Resolution Algorithm:**

```go
// ResolveConceptRanges processes ConceptRangeGenerators in the correct order.
// It must be called after all concepts have been initially processed but before
// the analysis phase completes.
func (a *CPLAnalyzer) ResolveConceptRanges() error {
    // Track which concepts have ConceptRangeGenerators that need resolution
    type pendingRange struct {
        conceptName  string
        generator    *ConceptRangeGenerator
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

            // Check if all values in the source concept are resolved
            sourceHasPendingRanges := false
            for _, gen := range sourceConcept.Generators {
                if _, isCRG := gen.(*ConceptRangeGenerator); isCRG {
                    sourceHasPendingRanges = true
                    break
                }
            }

            if sourceHasPendingRanges {
                // Source concept still has pending ranges, defer this one
                stillPending = append(stillPending, p)
                continue
            }

            // Source concept is fully resolved, resolve this range
            values, err := p.generator.GenerateValues(&a.concepts)
            if err != nil {
                return fmt.Errorf("failed to resolve concept range for '%s': %w", p.conceptName, err)
            }

            // Add the generated values to the concept
            for _, val := range values {
                if _, exists := concept.ValueIndex[val]; exists {
                    return fmt.Errorf("duplicate value '%s' in concept '%s' from concept range",
                        val, p.conceptName)
                }
                pos := len(concept.Values)
                concept.Values = append(concept.Values, AnalysisConceptValueDefinition{
                    Name:     val,
                    Position: pos,
                })
                concept.ValueIndex[val] = pos
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
```

**Value Ordering:**

Value order is maintained from the referenced concept:

- For Values generators: order is as declared
- For Int Range generators: order is from Start to End by Step
- For Concept Range generators: order follows the referenced concept's ordering from StartValue to EndValue
- Multiple generators: values appear in the order generators are declared

**Validation During Analysis:**

- Referenced concept must exist
- Start and end values must exist in the referenced concept
- Start value must come before or at the same position as end value in the referenced concept
- No circular dependencies among concept range generators
- Duplicate detection across all generators
- Error messages include source location

**Circular Dependency Example:**

```cpl
concepts
    A=[B[Value1..Value2]]  # A references B
    B=[A[Value1..Value2]]  # B references A - CIRCULAR!
```

This should be detected and reported as an error during analysis.

**Related Documents:**

- [Concepts Incremental Implementation](../specifications/design/cpl/eros/ERO001_concepts-incremental-implementation.md) - Main incremental implementation ERO
