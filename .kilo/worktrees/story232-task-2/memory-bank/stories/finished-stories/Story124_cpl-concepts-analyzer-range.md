# Story124: CPL Concepts Analyzer - Range

## Goal

Implement Int range value expansion in the CPL analyzer. This story expands Int range generators into their full set of values during analysis, storing the expanded values in AnalysisConceptDefinition.Values. The values are ordered from Start to End with the specified step. This is part of [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md) - Increment 5.

## References

- [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md)
- [ERO001: Concepts Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO001_concepts-analysis-phase.md) - Increment 5 section
- [Concept Definitions Design](../specifications/design/cpl/concepts-definitions.md)

## Dependencies

- [Story123_cpl-concepts-parser-range.md](Story123_cpl-concepts-parser-range.md) - Not Started - Parser must capture RangeBounds

## Dependent Stories

- Story125_cpl-concepts-parser-before-after.md - Not Started - Before/After builds on Range

## Tasks

> **story-implementor Handoff Directive**: This is a coding task story. Per [`scripts/story-implementor.pdd.script.md`](../../../scripts/story-implementor.pdd.script.md), the story-implementor MUST:
>
> - Hand off to the **Code** agent via a **NEW AGENT TASK**, passing `story_context` (`story_file`, `current_task`, `activity_type: code`) and the current Logical TDD Lifecycle phase (A–D).
> - Ensure the Code agent returns results via a **NEW AGENT TASK** with the `story_context` section unchanged.
> - After each task completes, hand off to the **code-reviewer** via a **NEW AGENT TASK** with unchanged `story_context`.
> - Receive review results back via a **NEW AGENT TASK** with unchanged `story_context`.
> - **Checkpoint and WAIT for explicit user approval** before proceeding to the next task.

1. Implement ExpandIntRangeValues function - Completed
   a. Create values slice with capacity based on step - Completed
   b. Loop from Start to End using step, appending string values - Completed
   c. Return ordered value slice - Completed

2. Update ConceptAnalyzer for Int range support - Completed
   a. Check child scope type in AnalyzeConceptsScope to detect IntGeneratorScope - Completed
   b. Call ExpandIntRangeValues for IntRangeGenerator - Completed
   c. Populate Values from expanded range - Completed
   d. Build ValueIndex map from expanded values - Completed

3. Implement validation - Completed
   a. Validate range bounds (start < end when step > 0, non-negative) - Completed
   b. Validate step is non-zero - Completed
   c. Validate range doesn't produce too many values - Completed
   d. Report errors with source locations - Completed

4. Write unit tests for Int range analysis - Completed
   a. Test Int range expansion with default step - Completed
   b. Test Int range expansion with custom step - Completed
   c. Test ValueIndex from Int range values - Completed
   d. Test invalid range bounds - Completed

## Constraints

- Range bounds must be validated (start < end when step > 0, non-negative)
- Step must be non-zero
- Values must be ordered from Start to End using the specified step
- Must support Int range in direct concept definitions only (conditional specs not yet implemented)
- Follow the Logical TDD Lifecycle - write tests before implementation
- Must not break existing basic concept functionality (Values generator, multi-section, hierarchy)

## Intent

The goal of this story is to expand Int range generators into their full value sets during analysis. This allows the rest of the analysis phase (and subsequent phases) to work with complete value lists regardless of whether they came from Values or Int range generators. The step parameter enables generating sequences like even numbers or custom increments.

## Acceptance Criteria

1. ExpandIntRangeValues correctly expands Int[start, end, step] to ["start", "start+step", ..., "end"]
2. Int range values are ordered from Start to End using the specified step
3. ConceptAnalyzer handles both Values and IntRange generators
4. ValueIndex map provides O(1) lookup for Int range-expanded values
5. Invalid range bounds are detected and reported as errors
6. Step parameter correctly affects value generation (e.g., step=2 produces even numbers)
7. All unit tests pass following BTDD practices
8. Existing basic concept functionality still works

## Notes

**Increment 5 Capabilities:**

- Int range generator expansion during analysis
- Step parameter support for custom increments
- Consistent value handling (Values and Int range produce same structure)

**Increment 5 Limitations:**

- No conditional specifications (added in Increment 4 / Story121)
- No before/after clauses (added in Increment 7 / Story125)
- No API integration (added in Increment 8)

**Int Range Expansion Examples:**

```cpl
concepts
    Month=[Int[1, 12, 1]]
```

Expands to:

- Month.Values = ["1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"]
- Month.ValueIndex = {"1": 0, "2": 1, ..., "12": 11}

```cpl
concepts
    EvenNumbers=[Int[2, 10, 2]]
```

Expands to:

- EvenNumbers.Values = ["2", "4", "6", "8", "10"]
- EvenNumbers.ValueIndex = {"2": 0, "4": 1, "6": 2, "8": 3, "10": 4}

**ExpandIntRangeValues Implementation:**

```go
func ExpandIntRangeValues(generator IntRangeGeneratorDef) []string {
    if generator.Step == 0 {
        generator.Step = 1
    }

    // Calculate capacity based on step direction
    var count int
    if generator.Step > 0 {
        count = (generator.End-generator.Start)/generator.Step + 1
    } else {
        count = (generator.Start-generator.End)/(-generator.Step) + 1
    }

    values := make([]string, 0, count)
    if generator.Step > 0 {
        for i := generator.Start; i <= generator.End; i += generator.Step {
            values = append(values, strconv.Itoa(i))
        }
    } else {
        for i := generator.Start; i >= generator.End; i += generator.Step {
            values = append(values, strconv.Itoa(i))
        }
    }
    return values
}
```

**Value Ordering:**

Value order is maintained from the declaration:

- For Values generators: order is as declared
- For Int range generators: order is from Start to End using the specified step

**Analysis with Int Range:**

The analyzer walks the symbol table, inspects child scopes to determine generator type, and expands Int range values into the analysis definition:

```go
func (ca *ConceptAnalyzer) AnalyzeConceptsScope(table *SymbolTable) error {
    // ... existing code ...

    analysisDef := &AnalysisConceptDefinition{
        Name:       *def.Name,
        Values:     make([]AnalysisConceptValueDefinition, 0),
        ValueIndex: make(map[string]int),
        ParentName: parentName,
        Level:      targetLevel,
    }

    // Handle different generator types based on child scopes
    for _, child := range def.Children {
        switch child.ScopeType {
        case ValuesGeneratorScope:
            // Existing: process identifier values from ValuesGeneratorScope
            for i, valDef := range child.Definitions {
                valName := *valDef.Name
                analysisDef.Values = append(analysisDef.Values, AnalysisConceptValueDefinition{
                    Name:     valName,
                    Position: i,
                })
                analysisDef.ValueIndex[valName] = i
            }
        case IntGeneratorScope:
            // NEW: expand Int range values
            expanded := ExpandIntRangeValues(child)
            for i, val := range expanded {
                analysisDef.Values = append(analysisDef.Values, AnalysisConceptValueDefinition{
                    Name:     val,
                    Position: i,
                })
                analysisDef.ValueIndex[val] = i
            }
        }
    }

    // ... rest of analysis ...
}
```

**Related Documents:**

- [Concepts Incremental Implementation](../specifications/design/cpl/eros/ERO001_concepts-incremental-implementation.md) - Main incremental implementation ERO
