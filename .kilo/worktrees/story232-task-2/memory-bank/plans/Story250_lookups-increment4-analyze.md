# Plan: Story250 — Lookups Increment 4 (Analyze)

## Purpose

This plan provides the technical writer with sufficient detail to create Story250: **lookups-increment4-analyze** — implement lookup reference resolution in aggregator expressions, dimensional unit validation for aggregator-lookup combinations (Increment 4 of Epic006).

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `lookups-increment4-analyze`                                                                  |
| description  | Implement lookup reference resolution in aggregator expressions, dimensional unit validation for aggregator-lookup combinations |

## Source EROs

- [ERO006: Lookups Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO006_lookups-analysis-phase.md) — Increment 4 (documented for completeness; implementation details in ERO007)
- [ERO006: Lookups Incremental Implementation](../specifications/design/cpl/eros/ERO006_lookups-incremental-implementation.md) — Increment 4 sketch

## Prerequisites

- **Epic007 (Aggregators) must be at Increment 2 or above** — aggregator analysis infrastructure including `AnalysisDefinitions.Aggregators` map must exist
- Story242 (lookups-increment1-analyze) — `AnalysisDefinitions.Lookups` map must be populated
- This story is **blocked** until ERO007 aggregator analysis infrastructure is implemented

## Dependencies

- Epic007 Increment 2 — needs aggregator analysis infrastructure (`Aggregators` map, aggregator expression resolution)
- Story242 (lookups-increment1-analyze) — needs the Lookups map

## Context — No Parser Changes

Increment 4 requires **no parser changes**. Aggregator expression parsing already supports concept references and operators via the postfix expression pattern. Lookup references appear as ordinary identifiers in aggregator expressions; the analysis phase resolves them against the Lookups map.

The parser cannot distinguish lookup names from concept names in expressions — both are identifiers. The analysis phase resolves ambiguity by searching lookups before concepts (lookups take precedence when a name exists in both maps, which is an error condition that analysis reports).

## Files to Create or Modify

| File                                                     | Action | Purpose                                                                                              |
| -------------------------------------------------------- | ------ | ---------------------------------------------------------------------------------------------------- |
| `src/main/go/analysis/analyze_aggregators.go`            | Modify | Add lookup reference resolution in aggregator expression analysis                                    |
| `src/main/go/analysis/aggregator_lookup_validation.go`   | Create | Dimensional unit validation for aggregator-lookup combinations                                       |
| `src/main/go/analysis/aggregator_lookup_validation_test.go` | Create | Tests for aggregator-lookup dimensional unit validation                                              |

**Note:** The exact file paths depend on how the aggregator analysis phase is structured in ERO007. The paths above are indicative — the technical writer should align with the actual aggregator package organization.

## Analysis Logic — Lookup Reference Resolution in Aggregator Expressions

When the aggregator analysis function (from ERO007) walks an aggregator's expression definitions, it encounters identifiers that could be concept references, metric references, or lookup references.

### Resolution Algorithm

```
for each identifier in aggregator expression:
    1. Check AnalysisDefinitions.Lookups[identifier] — if found, it's a lookup reference
    2. Check AnalysisDefinitions.Metrics[identifier] — if found, it's a metric reference
    3. Check AnalysisDefinitions.Concepts[identifier] — if found, it's a concept reference
    4. If NOT FOUND in any map → error: undeclared reference
    5. If found in BOTH Lookups and another map → error: ambiguous reference
```

### Dimensional Unit Validation

When a lookup is referenced in an aggregator expression:

1. **Dimensional form compatibility**: The lookup's dimensions must be compatible with the aggregator's dimensional transformation. The aggregator transforms data from one dimensional form to another (e.g., `[Product, Currency] → [Currency]`). The lookup's key dimensions must be available in at least one of the aggregator's dimensional contexts (input or output).

2. **Implicit transformations**: If the lookup's dimensions do not match the aggregator's dimensional context, the analyzer may need to determine that the lookup itself needs to be dimensionally transformed (aggregated or spread) before use. This applies implicit aggregators or spreaders to the lookup.

3. **Unit compatibility**: The lookup's celltype units must be compatible with the aggregation expression that references it. Celltype's compound units must compose correctly with the other operands in the expression.

### Recording in CIL Aggregator Node

The resolved lookup reference is recorded in the aggregator's `CILAggregateNode` (or equivalent structure defined in ERO007), including any implicit transformations needed to align the lookup dimensions with the aggregator context.

## Task Decomposition (for technical writer)

### Task 1: Implement lookup reference resolution in aggregator expressions — Not Started
   a. Add lookup name resolution to the aggregator expression identifier resolution logic — Not Started
   b. Search `AnalysisDefinitions.Lookups` before `Concepts` (lookups take precedence) — Not Started
   c. Report error when a name exists in both Lookups and another map (ambiguous reference) — Not Started
   d. Write unit tests: valid lookup reference, non-existent identifier, ambiguous lookup/concept name — Not Started
   e. Verify analyzer tests pass — Not Started

### Task 2: Implement dimensional unit validation for lookup-aggregator combinations — Not Started
   a. Create `validateAggregatorLookupCompatibility` function — Not Started
   b. Check that the lookup's key dimensions are available in the aggregator's dimensional context — Not Started
   c. Determine if the lookup needs implicit dimensional transformation (aggregator/spreader applied to the lookup) — Not Started
   d. Verify lookup celltype units are compatible with the aggregation expression — Not Started
   e. Write unit tests: compatible dimensions, incompatible dimensions, implicit transformation needed, unit incompatibility — Not Started
   f. Verify tests pass — Not Started

### Task 3: Record resolved lookup references in CIL aggregator nodes — Not Started
   a. Store resolved lookup references in the aggregator's CIL node structure (per ERO007's CIL node design) — Not Started
   b. Record any implicit transformation nodes needed for dimensional alignment — Not Started
   c. Write unit tests: verify CIL node structure includes lookup reference metadata — Not Started
   d. Verify tests pass — Not Started

### Task 4: Create parse_and_analyze_only examples — Not Started
   a. In `without_errors/`: add CPL files with aggregators referencing lookups in their expressions — Not Started
   b. In `with_errors/`: add CPL files with incompatible dimensions, ambiguous lookup/concept names, unit mismatches — Not Started
   c. Verify all examples produce expected analysis results — Not Started

## Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4
```

All tasks are sequential — each builds on the previous.

## Parallel Execution

None — all tasks have sequential dependencies.

## Reintegration

None — no parallel groups.

## Acceptance Criteria Summary

1. Aggregator expression identifiers resolve against `AnalysisDefinitions.Lookups` map
2. Lookup names take precedence over concept names when both exist (ambiguous reference is an error)
3. Lookup key dimensions are validated against aggregator dimensional context
4. Implicit dimensional transformations on lookups are detected and recorded
5. Lookup celltype units are validated against aggregation expression units
6. Resolved lookup references are recorded in the aggregator's CIL node
7. Clear error messages for incompatible dimensions and unit mismatches

## Intent Summary for Technical Writer

This story enables aggregator expressions to reference lookups. Since no parser changes are required (lookup names appear as ordinary identifiers in expressions), the work is entirely in the analysis phase. The aggregator analysis function resolves identifiers against the Lookups map, validates dimensional compatibility between the lookup's key dimensions and the aggregator's dimensional context, and records the resolved reference (including any implicit transformations) in the aggregator's CIL node. This story is blocked until Epic007's aggregator analysis infrastructure exists at Increment 2 or above.
