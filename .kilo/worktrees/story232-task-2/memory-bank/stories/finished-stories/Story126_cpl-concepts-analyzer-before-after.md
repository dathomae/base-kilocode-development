# Story126: CPL Concepts Analyzer - Before/After

## Goal

Implement value ordering computation with before/after clauses in the CPL analyzer. This story computes the final value order when concepts are extended with insertion points. The analyzer validates that before/after values exist and computes new value positions while preserving the relative order of original values. This is part of [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md) - Increment 6.

## References

- [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md)
- [ERO001: Concepts Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO001_concepts-analysis-phase.md) - Increment 6 section
- [Concept Definitions Design](../specifications/design/cpl/concepts-definitions.md)

## Dependencies

- [Story125_cpl-concepts-parser-before-after.md](Story125_cpl-concepts-parser-before-after.md) - Not Started - Parser must capture insertion points

## Dependent Stories

None - This is the final story for Increments 1-6.

## Tasks

1. Implement ComputeValueOrdering function - Completed
   a. Copy base values to ordered slice - Completed
   b. Process insertions in order - Completed
   c. Find insertion positions based on before/after - Completed
   d. Insert new values at computed positions - Completed
   e. Return ordered AnalysisConceptValueDefinition slice - Completed

2. Update ConceptAnalyzer for before/after support - Completed
   a. Detect concept extensions (same name in multiple sections) - Completed
   b. Merge initial definition with extensions - Completed
   c. Call ComputeValueOrdering for extended concepts - Completed

3. Implement validation - Completed
   a. Validate before/after values exist in target concept - Completed
   b. Validate no duplicate insertions at same position - Completed
   c. Report errors with source locations - Completed

4. Update value indexing - Completed (implemented in Task 2's mergeConceptValues)
   a. Rebuild ValueIndex map after ordering computation - Completed
   b. Update Position fields in AnalysisConceptValueDefinition - Completed

5. Write unit tests for before/after analysis - Completed (covered by compute_value_ordering_test.go and concept_extension_merge_test.go)
   a. Test before clause computation - Completed
   b. Test after clause computation - Completed
   c. Test multiple extensions - Completed
   d. Test invalid before/after value detection - Completed
   e. Test original value order preservation - Completed

## Constraints

- Cannot reorder existing values - only insert new ones
- Before/after values must exist in the target concept
- Original value relative order must be preserved
- Follow the Logical TDD Lifecycle - write tests before implementation
- Must not break Increment 1-5 functionality

## Intent

The goal of this story is to compute the final value order when concepts are extended using before/after clauses. The analyzer processes extensions in order, validates insertion points, and builds the final ordered value list while preserving the relative order of values from the initial declaration.

## Acceptance Criteria

1. ComputeValueOrdering correctly inserts values at specified positions
2. Before clause inserts values before the specified existing value
3. After clause inserts values after the specified existing value
4. Invalid before/after values are detected and reported as errors
5. Original value relative order is preserved
6. ValueIndex map reflects final computed order
7. Multiple extensions are processed in section order
8. All unit tests pass following BTDD practices
9. Increment 1-5 functionality still works

## Notes

**Increment 6 Capabilities:**

- Value ordering computation with before/after
- Concept extension merging across sections
- Original value order preservation

**Value Ordering Rules:**

1. Initial declaration establishes canonical order
2. Extensions can only add new values, not reorder existing
3. Multiple extensions processed in section order
4. Before/after values must exist in merged concept

**Increment 6 Limitations:**

- No API integration (added in Increment 7)

**Value Ordering Computation:**

```go
func ComputeValueOrdering(
    baseValues []string,
    insertions []InsertionPoint) ([]AnalysisConceptValueDefinition, error) {

    ordered := make([]string, len(baseValues))
    copy(ordered, baseValues)

    for _, ins := range insertions {
        var insertPos int

        if ins.BeforeValue != "" {
            // Find position of "before" value
            pos, exists := findValue(ordered, ins.BeforeValue)
            if !exists {
                return nil, fmt.Errorf("'before' value not found: %s", ins.BeforeValue)
            }
            insertPos = pos
        } else if ins.AfterValue != "" {
            // Find position after "after" value
            pos, exists := findValue(ordered, ins.AfterValue)
            if !exists {
                return nil, fmt.Errorf("'after' value not found: %s", ins.AfterValue)
            }
            insertPos = pos + 1
        } else {
            return nil, fmt.Errorf("insertion must specify before or after")
        }

        // Insert the new values
        ordered = insertValues(ordered, insertPos, ins.NewValues)
    }

    // Convert to AnalysisConceptValueDefinition slice
    result := make([]AnalysisConceptValueDefinition, len(ordered))
    for i, val := range ordered {
        result[i] = AnalysisConceptValueDefinition{Name: val, Position: i}
    }

    return result, nil
}
```

**Generator-Based Structure:**

The analyzer works with the new generator wrapper syntax where concept declarations use `=[Values[...]]` format. The `Generators` field in `ParserConceptDefinition` contains the generator patterns, and extensions reference these via the `InsertionPoint` structure.

**Example Analysis:**

```cpl
concepts
    Priority=[Values[Low, Medium, High]]

concepts
    Priority=[Values[Critical]] before Low
    Priority=[Values[Urgent]] after Medium
```

Computation:

1. Base: [Low, Medium, High]
2. Insert [Critical] before Low: [Critical, Low, Medium, High]
3. Insert [Urgent] after Medium: [Critical, Low, Medium, Urgent, High]

Final Values: [Critical, Low, Medium, Urgent, High]

**Validation During Analysis:**

- Before/after values must exist in the target concept
- Cannot reorder existing values - before/after only applies to new values being added
- No duplicate insertions at same position
- Extension before/after values must exist in merged concept

**Related Documents:**

- [Concepts Incremental Implementation](../specifications/design/cpl/eros/ERO001_concepts-incremental-implementation.md) - Main incremental implementation ERO
