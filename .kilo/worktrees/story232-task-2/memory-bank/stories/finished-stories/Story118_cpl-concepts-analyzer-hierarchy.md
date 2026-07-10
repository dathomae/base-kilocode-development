# Story118: CPL Concepts Analyzer - Hierarchy

## Goal

Implement concept hierarchy analysis in the CPL analyzer. This story builds parent-child relationships from the indentation levels captured during parsing. This is part of [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md) - Increment 2.

## References

- [Epic001: CPL Concepts Implementation](../epics/Epic001_cpl-concepts-implementation.md)
- [ERO001: Concepts Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO001_concepts-analysis-phase.md) - Increment 2 section
- [Concept Definitions Design](../specifications/design/cpl/concepts-definitions.md)

## Dependencies

- [Story117_cpl-concepts-parser-hierarchy.md](../finished-stories/Story117_cpl-concepts-parser-hierarchy.md) - Done - Parser captures indentation levels

## Dependent Stories

- Story119_cpl-concepts-parser-multi-section.md - Not Started - Multi-section builds on hierarchy

## Tasks

1. Enhance AnalysisConceptDefinition with hierarchy fields - Done
   a. Add ParentConceptName field - Done (renamed from ParentName)
   b. Add IndentAsSpaces field - Done (captures indentation level)

2. Implement parent-child relationship detection in analyzer - Done
   a. Compare indentation levels in AddConcept - Done
   b. Set ParentConceptName when concept is nested - Done

3. Write unit tests for hierarchy analysis - Done
   a. Test parent-child relationship building - Done
   b. Test multi-level hierarchy - Done

## Constraints

- Must work with parser-level indentation information
- Follow the Logical TDD Lifecycle - write tests before implementation
- Must not break Increment 1 functionality

## Implementation Approach

**Note:** The implementation is simpler than originally specified. Instead of a separate ConceptHierarchy struct with RootConcepts and AllPaths, the hierarchy is tracked directly on each AnalysisConceptDefinition:

1. **Parent Relationship**: The `AddConcept` method in `cpl_analyzer.go` compares indentation levels between consecutive concepts. When a concept has greater indentation, it is marked as a child:

```go
if previousConcept.IndentAsSpaces < concept.IndentAsSpaces {
    concept.ParentConceptName = &previousConcept.ConceptName
}
```

2. **Indentation Tracking**: The `IndentAsSpaces` field captures the indentation level, which implicitly represents the hierarchy depth.

**Features NOT implemented** (may be added in future increments):

- ChildNames slice (reverse parent-child lookup)
- ConceptHierarchy struct with RootConcepts
- AllPaths map for root-to-concept paths
- Cycle detection validation
- Tree structure validation

These features were deemed unnecessary for the current increment but can be added when tensor shape computation from hierarchical dimensions is required.

## Acceptance Criteria

1. ✅ AnalysisConceptDefinition includes ParentConceptName field
2. ✅ AnalysisConceptDefinition includes IndentAsSpaces field for hierarchy depth
3. ✅ ConceptAnalyzer correctly builds parent-child relationships from indentation levels
4. ✅ All unit tests pass following BTDD practices
5. ✅ Increment 1 flat concept analysis still works

## Notes

**Increment 2 Capabilities:**

- Parent-child relationship building from indentation
- Hierarchy depth tracking via IndentAsSpaces

**Increment 2 Limitations:**

- No ChildNames slice (reverse lookup from parent to children)
- No ConceptHierarchy struct with RootConcepts and AllPaths
- No cycle detection (assumed valid input)
- No tree structure validation
- Child values are global, not conditional on parent (added in Increment 4)
- No multiple concept declaration sections (added in Increment 3)
- No Range generators (added in Increment 5)
- No before/after clauses (added in Increment 6)

**Related Documents:**

- [Concepts Incremental Implementation](../specifications/design/cpl/eros/ERO001_concepts-incremental-implementation.md) - Main incremental implementation ERO
- [Intermediate Language](../specifications/architecture/cpl/analysis/intermediate_language.md) - CIL specification
