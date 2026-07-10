# Plan: Story245 — Lookups Increment 2 (Analyze)

## Purpose

This plan provides the technical writer with sufficient detail to create Story245: **lookups-increment2-analyze** — implement grain validation in the analysis phase for lookup declarations with grain (Increment 2 of Epic006).

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `lookups-increment2-analyze`                                                                  |
| description  | Implement grain validation (Interval type constraint, duplicate check, empty check), extend LookupInfo and CILLookupNode with grain fields |

## Source EROs

- [ERO006: Lookups Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO006_lookups-analysis-phase.md) — Increment 2 grain validation and structure extensions

## Dependencies

- Story244 (lookups-increment2-parse) — needs `ParserLookupDefinition.HasGrain` and `Grain` fields
- Story242 (lookups-increment1-analyze) — extends existing `AnalyzeLookupScope` function

## Files to Create or Modify

| File                                              | Action | Purpose                                                                                       |
| ------------------------------------------------- | ------ | --------------------------------------------------------------------------------------------- |
| `src/main/go/analysis/lookup_info.go`             | Modify | Add `Grain []*AnalysisMeasureDefinition` and `HasGrain bool` fields                           |
| `src/main/go/analysis/cil_lookup_node.go`         | Modify | Add `Grain []string` and `HasGrain bool` fields                                               |
| `src/main/go/analysis/analyze_lookups.go`         | Modify | Extend `AnalyzeLookupScope` with grain resolution, type validation, duplicate checks          |
| `src/main/go/analysis/analyze_lookups_test.go`    | Modify | Add tests for grain validation: Interval types, non-Interval types, empty grain, duplicates  |

### LookupInfo Changes (Increment 2 — fields added)

```go
type LookupInfo struct {
    Name              string
    Dimensions        []*AnalysisConceptDefinition
    CelltypeMeasure   *AnalysisMeasureDefinition
    Grain             []*AnalysisMeasureDefinition      // NEW: resolved grain references (nil if !HasGrain)
    HasGrain          bool                              // NEW: true when lookup has temporal dimension
    SourceLocation    SourceLocation
}
```

### CILLookupNode Changes (Increment 2 — fields added)

```go
type CILLookupNode struct {
    LookupName        string
    Dimensions        []string
    Grain             []string     // NEW: grain measure names (empty if no grain)
    HasGrain          bool         // NEW: whether grain clause was present
    Celltype          string
    SourceLocation    SourceLocation
}
```

### AnalyzeLookupScope — Grain Validation Extension

When `parserDef.HasGrain` is true:
1. **Empty check**: If `len(parserDef.Grain) == 0` → error "grain clause present but no grain measures specified"
2. **Duplicate check**: Grain measure names must be unique within a lookup's grain clause → use `uniqueStrings()` helper
3. **Existence**: Each grain measure name must resolve to an existing `AnalysisMeasureDefinition`
4. **Type constraint**: Each grain measure's type must be `Interval` or derived from `Interval` → use `isIntervalType()` helper

The `isIntervalType` helper walks the type parent chain:
```go
func isIntervalType(typeDef *AnalysisTypeDefinition) bool {
    if typeDef == nil { return false }
    if typeDef.Name == "Interval" { return true }
    return isIntervalType(typeDef.ParentType)
}
```

### CILLookupNode Construction (modified)

When `HasGrain` is true, populate `Grain` and `HasGrain` fields:
```go
cilNode := &CILLookupNode{
    LookupName:     parserDef.Name,
    Dimensions:     dimensionNames,
    Grain:          grainNames,       // resolved grain measure names
    HasGrain:       parserDef.HasGrain,
    Celltype:       parserDef.Celltype,
    SourceLocation: parserDef.SourceLocation,
}
```

## Task Decomposition (for technical writer)

### Task 1: Extend LookupInfo and CILLookupNode with grain fields — Not Started [Small]
   a. Add `Grain []*AnalysisMeasureDefinition` and `HasGrain bool` to `LookupInfo` — Not Started
   b. Add `Grain []string` and `HasGrain bool` to `CILLookupNode` — Not Started
   c. Update existing tests to verify zero values for grain fields when no grain — Not Started
   d. Verify package compiles — Not Started

### Task 2: Implement grain validation in AnalyzeLookupScope — Not Started
   a. Implement empty grain check: `HasGrain && len(Grain) == 0` → error — Not Started
   b. Implement duplicate grain check: collect grain names, verify uniqueness — Not Started
   c. Implement grain measure resolution: foreach grain name, find in `a.measures` — Not Started
   d. Implement `isIntervalType` helper: walk type parent chain to find `Interval` — Not Started
   e. Implement Interval type validation: each grain measure's `TypeRef` must be `Interval` or derived — Not Started
   f. Populate `info.Grain` and `info.HasGrain` — Not Started
   g. Populate `cilNode.Grain` and `cilNode.HasGrain` — Not Started
   h. Write unit tests: valid single-grain, valid multi-grain, missing measure, non-Interval type, duplicates, empty grain clause — Not Started
   i. Verify analyzer tests pass — Not Started

### Task 3: Verify backward compatibility and edge cases — Not Started [Small]
   a. Verify timeless lookups (no grain) still analyze successfully — Not Started
   b. Verify lookups with both grain and no-grain coexist in the same compilation unit — Not Started
   c. Verify error messages include lookup name and source location — Not Started
   d. Verify full test suite passes — Not Started

### Task 4: Create parse_and_analyze_only examples with grain — Not Started
   a. In `without_errors/`: add CPL files with temporal lookups (single grain, multiple grain measures) that analyze successfully — Not Started
   b. In `with_errors/`: add CPL files that produce analysis errors — non-Interval grain types, non-existent grain measures, empty grain clause, duplicate grain measure names — Not Started
   c. Verify all examples produce expected analysis results — Not Started

## Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4
```

All tasks are sequential — Task 2 needs Task 1 types, Task 3 validates the combined behavior, Task 4 needs full pipeline.

## Parallel Execution

None — all tasks have sequential dependencies.

## Reintegration

None — no parallel groups.

## Acceptance Criteria Summary

1. `LookupInfo.Grain` contains resolved `AnalysisMeasureDefinition` references when grain is present
2. `LookupInfo.HasGrain` correctly reflects presence of grain clause
3. `CILLookupNode.Grain` contains resolved grain measure names
4. Grain measure names must exist in the `Measures` map
5. Grain measure types must be `Interval` or derived from `Interval`
6. Empty grain clause (`HasGrain=true, len(Grain)=0`) is rejected with a clear error
7. Duplicate grain measure names in a single lookup's grain clause are rejected
8. Timeless lookups (no grain) continue to work unchanged

## Intent Summary for Technical Writer

This story extends the analysis phase to validate grain clauses in lookup declarations. When a parser lookup definition has `HasGrain=true`, the analyzer resolves each grain measure name against the Measures map, validates that each has an `Interval`-derived type, checks for empty and duplicate grain measures, and populates the extended `LookupInfo` and `CILLookupNode` structures. The `isIntervalType` helper walks the type parent chain to handle derived Interval types. Timeless lookups (no grain) are unaffected — no structural changes to their analysis path.
