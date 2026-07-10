# Plan: Story252 — Lookups Increment 5 (Analyze)

## Purpose

This plan provides the technical writer with sufficient detail to create Story252: **lookups-increment5-analyze** — implement lookup reference resolution in spreader expressions, implicit aggregator resolution for lookups, and dimensional/cell unit validation (Increment 5 of Epic006).

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `lookups-increment5-analyze`                                                                  |
| description  | Implement lookup reference resolution in spreader expressions, implicit aggregator resolution for lookups, dimensional/cell unit validation |

## Source EROs

- [ERO006: Lookups Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO006_lookups-analysis-phase.md) — Increment 5 (documented for completeness; implementation details in ERO008)
- [ERO006: Lookups Incremental Implementation](../specifications/design/cpl/eros/ERO006_lookups-incremental-implementation.md) — Increment 5 sketch

## Prerequisites

- **Epic008 (Spreaders) must be at Increment 3 or above** — spreader analysis infrastructure including `AnalysisDefinitions.Spreaders` map, spreader expression resolution, and implicit aggregator resolution must exist
- Story242 (lookups-increment1-analyze) — `AnalysisDefinitions.Lookups` map must be populated
- This story is **blocked** until ERO008 spreader analysis infrastructure is implemented

## Dependencies

- Epic008 Increment 3 — needs spreader analysis infrastructure (Spreaders map, expression resolution, implicit aggregator resolution)
- Story242 (lookups-increment1-analyze) — needs the Lookups map

## Context — No Parser Changes

Increment 5 requires **no parser changes**. Spreader expression parsing already supports metric references and concept references via the postfix expression pattern. Lookup references appear as ordinary identifiers in spreader expressions; the analysis phase resolves them against the Lookups map.

The same precedence rule from Increment 4 applies: when a name exists in both the Lookups map and another map (Metrics, Concepts), it is an ambiguous reference and analysis reports it as an error.

## Files to Create or Modify

| File                                                     | Action | Purpose                                                                                              |
| -------------------------------------------------------- | ------ | ---------------------------------------------------------------------------------------------------- |
| `src/main/go/analysis/analyze_spreaders.go`              | Modify | Add lookup reference resolution in spreader expression analysis                                      |
| `src/main/go/analysis/spreader_lookup_validation.go`     | Create | Dimensional unit validation for spreader-lookup combinations, implicit aggregator resolution          |
| `src/main/go/analysis/spreader_lookup_validation_test.go`| Create | Tests for spreader-lookup dimensional unit validation and implicit aggregator resolution             |

**Note:** The exact file paths depend on how the spreader analysis phase is structured in ERO008. The paths above are indicative — the technical writer should align with the actual spreader package organization.

## Analysis Logic

### Lookup Reference Resolution in Spreader Expressions

When the spreader analysis function (from ERO008) walks a spreader's expression definitions, it encounters identifiers that could be metric references, lookup references, or concept references.

### Resolution Algorithm

```
for each identifier in spreader expression:
    1. Check AnalysisDefinitions.Lookups[identifier] — lookup reference
    2. Check AnalysisDefinitions.Metrics[identifier] — metric reference
    3. Check AnalysisDefinitions.Concepts[identifier] — concept reference
    4. If NOT FOUND → error: undeclared reference
    5. If found in BOTH Lookups and another map → error: ambiguous reference
```

### Dimensional Unit Analysis for Spreader-Lookup Combinations

When a lookup is referenced in a spreader expression, the analyzer applies the same dimensional unit analysis used for metric references in spreaders:

1. **Determine required dimensional form**: The lookup's dimensions must support the spreader's dimensional transformation (e.g., spreading from `[Country]` to `[Country, Product]`).

2. **Apply implicit aggregators/spreaders to the lookup if needed**: If the lookup's key dimensions do not directly match the required dimensions, the analyzer determines whether an implicit aggregator or spreader should be applied to the lookup to align its dimensions. This is the same mechanism used for metric references in spreaders.

3. **Cell unit validation**: The lookup's celltype units must be dimensionally compatible with the spreader expression. For example, if the spreader distributes `Amount * WeightFactor`, the lookup's celltype units must be compatible with the multiplication context.

### Recording in CIL Spreader Node

The resolved lookup reference and any implicit transformation nodes are recorded in the spreader's `CILSpreadNode` (or equivalent structure defined in ERO008).

## Task Decomposition (for technical writer)

### Task 1: Implement lookup reference resolution in spreader expressions — Not Started
   a. Add lookup name resolution to the spreader expression identifier resolution logic — Not Started
   b. Search `AnalysisDefinitions.Lookups` before `Concepts` (lookups take precedence) — Not Started
   c. Report error when a name exists in both Lookups and another map (ambiguous reference) — Not Started
   d. Write unit tests: valid lookup reference, ambiguous lookup/metric name, ambiguous lookup/concept name — Not Started
   e. Verify analyzer tests pass — Not Started

### Task 2: Implement dimensional form analysis for lookup references — Not Started
   a. Determine the required dimensional form for the lookup in the spreader context — Not Started
   b. Check if the lookup's current dimensions match the required form — Not Started
   c. If mismatch, determine what implicit transformations (aggregators/spreaders) can align the lookup's dimensions — Not Started
   d. Record implicit transformations in the spreader's CIL node — Not Started
   e. Write unit tests: lookup dimensions match spreader context, lookup needs aggregation, lookup needs spreading — Not Started
   f. Verify tests pass — Not Started

### Task 3: Implement cell unit validation for lookup-spreader combinations — Not Started
   a. Verify the lookup's celltype units are compatible with the spreader expression's dimensional algebra — Not Started
   b. Apply compound unit composition rules for the specific spreader expression — Not Started
   c. Write unit tests: compatible celltype units, incompatible celltype units, compound unit edge cases — Not Started
   d. Verify tests pass — Not Started

### Task 4: Record resolved references in CIL spreader nodes — Not Started
   a. Store resolved lookup references in the spreader's CIL node — Not Started
   b. Include any implicit transformation nodes needed — Not Started
   c. Write unit tests: verify CIL node structure completeness — Not Started
   d. Verify tests pass — Not Started

### Task 5: Create parse_and_analyze_only examples — Not Started
   a. In `without_errors/`: add CPL files with spreaders referencing lookups in their expressions — Not Started
   b. In `with_errors/`: add CPL files with incompatible dimensions, implicit transformation failures, unit mismatches — Not Started
   c. Verify all examples produce expected analysis results — Not Started

## Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4 → Task 5
```

All tasks are sequential — each builds on the previous.

## Parallel Execution

None — all tasks have sequential dependencies.

## Reintegration

None — no parallel groups.

## Acceptance Criteria Summary

1. Spreader expression identifiers resolve against `AnalysisDefinitions.Lookups` map
2. Ambiguous lookup/metric and lookup/concept references produce clear errors
3. Dimensional form analysis determines required lookup dimensions in spreader context
4. Implicit aggregators/spreaders are generated for lookups when dimensional alignment is needed
5. Cell unit validation ensures lookup celltype units are compatible with spreader expression
6. Resolved references (including implicit transformations) are recorded in CIL spreader nodes
7. Clear error messages for incompatible dimensions, unit mismatches, and unresolvable transformations

## Intent Summary for Technical Writer

This story enables spreader expressions to reference lookups. Since no parser changes are required, the work is entirely in the analysis phase. The spreader analysis function resolves identifiers against the Lookups map, applies the same dimensional unit analysis used for metric references in spreaders (determining required dimensional form, applying implicit aggregators/spreaders when the lookup's dimensions don't directly match), and validates cell unit compatibility. The resolved reference and any implicit transformations are recorded in the spreader's CIL node. This story is blocked until Epic008's spreader analysis infrastructure exists at Increment 3 or above.
