# Plan: Story248 — Lookups Increment 3 (Analyze)

## Purpose

This plan provides the technical writer with sufficient detail to create Story248: **lookups-increment3-analyze** — implement lookup reference resolution in metric inputs, dimensional compatibility validation, and `CILLookupRefNode` generation (Increment 3 of Epic006).

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `lookups-increment3-analyze`                                                                  |
| description  | Implement lookup reference resolution in metric inputs, CILLookupRefNode, dimensional compatibility validation, lookup input example project |

## Source EROs

- [ERO006: Lookups Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO006_lookups-analysis-phase.md) — Increment 3 lookup input resolution and dimensional validation

## Dependencies

- Story247 (lookups-increment3-parse) — needs `InputKind = "lookup"` in parser input definitions
- Story242 (lookups-increment1-analyze) — needs `AnalysisDefinitions.Lookups` map populated

## Files to Create or Modify

| File                                              | Action | Purpose                                                                                        |
| ------------------------------------------------- | ------ | ---------------------------------------------------------------------------------------------- |
| `src/main/go/analysis/analysis_definitions.go`    | Modify | Add `ResolvedLookup *LookupInfo` field to `AnalysisInputDefinition`                            |
| `src/main/go/analysis/cil_lookup_ref_node.go`     | Create | `CILLookupRefNode` struct implementing `CILNode`                                               |
| `src/main/go/analysis/analyze_metrics.go`         | Modify | Extend metric input resolution to handle `"lookup"` InputKind; add `validateLookupMetricCompatibility` |
| `src/main/go/analysis/analyze_metrics_test.go`    | Modify | Add tests for lookup input resolution and dimensional compatibility validation                 |

### AnalysisInputDefinition Extension

```go
type AnalysisInputDefinition struct {
    InputKind       string
    Name            string
    ResolvedMetric  *AnalysisMetricDefinition
    ResolvedLookup  *LookupInfo                     // NEW: populated for "lookup" InputKind
    SourceLocation  SourceLocation
}
```

### CILLookupRefNode

```go
type CILLookupRefNode struct {
    LookupName    string
    Dimensions    []string    // Metric's dimension context for indexing
    Celltype      string
}

func (n *CILLookupRefNode) cilNode() {}
```

### Metric Input Resolution — Lookup Case

In the metric analysis function, when processing inputs, add the `"lookup"` case:

1. `ResolveReference`: Find `lookupInfo` in `a.lookups[parserInput.Name]`; error if not found
2. `SetResolvedLookup`: Set `input.ResolvedLookup = lookupInfo`
3. `ValidateDimensionalCompatibility`: Call `validateLookupMetricCompatibility`

### validateLookupMetricCompatibility

This function checks that the lookup's key dimensions are available in the metric's dimensional context:

```go
func (a *CPLAnalyzer) validateLookupMetricCompatibility(
    lookup *LookupInfo,
    parserDef *ParserMetricDefinition,
) error {
    // Build set of metric's dimension concepts from its inputs
    metricDims := map[string]bool
    for _, input := range parserDef.Inputs {
        if input.InputKind == "dimension" {
            metricDims[input.Name] = true
        }
    }
    // Each lookup dimension must be in the metric's dimension set
    for _, dim := range lookup.Dimensions {
        if !metricDims[dim.Name] {
            return fmt.Errorf("lookup %s is keyed by %q but metric does not declare %q as a dimension", ...)
        }
    }
    return nil
}
```

## Task Decomposition (for technical writer)

### Task 1: Define CILLookupRefNode type — Not Started [Small]
   a. Create `src/main/go/analysis/cil_lookup_ref_node.go` — Not Started
   b. Define `CILLookupRefNode` struct with `LookupName`, `Dimensions`, `Celltype` fields — Not Started
   c. Implement `cilNode()` interface method — Not Started
   d. Write unit test for struct initialization and interface compliance — Not Started
   e. Verify package compiles — Not Started

### Task 2: Extend AnalysisInputDefinition for lookup resolution — Not Started [Small]
   a. Add `ResolvedLookup *LookupInfo` field to `AnalysisInputDefinition` — Not Started
   b. Write unit test verifying default nil value and population — Not Started
   c. Verify package compiles — Not Started

### Task 3: Implement lookup input resolution in metric analysis — Not Started
   a. Add `case "lookup"` branch to metric input resolution switch — Not Started
   b. Resolve lookup name against `a.lookups` map; error if not found — Not Started
   c. Set `input.ResolvedLookup = lookupInfo` — Not Started
   d. Generate `CILLookupRefNode` and weave into the metric's CIL tree where the lookup is used in the calculate expression — Not Started
   e. Write unit tests: valid lookup reference, non-existent lookup, lookup used in calculate expression — Not Started
   f. Verify analyzer tests pass — Not Started

### Task 4: Implement dimensional compatibility validation — Not Started
   a. Create `validateLookupMetricCompatibility` function — Not Started
   b. Build set of metric's dimension concepts from `dimension` input kind declarations — Not Started
   c. Verify each lookup key dimension is present in the metric's dimension set — Not Started
   d. Report clear errors when dimensions are missing: "lookup X is keyed by Y but metric does not declare Y as a dimension" — Not Started
   e. Write unit tests: compatible dimensions, incompatible (missing) dimensions, edge case with grain dimensions — Not Started
   f. Verify tests pass — Not Started

### Task 5: Create parse_and_analyze_only examples with lookup inputs — Not Started
   a. In `with_errors/`: add CPL files with metrics referencing non-existent lookups, metrics missing lookup key dimensions — Not Started
   b. In `without_errors/`: add CPL files with metrics correctly referencing lookups with compatible dimensions — Not Started
   c. Verify all examples produce expected analysis results — Not Started

## Execution Order

```
[Task 1 ‖ Task 2] → Task 3 → Task 4 → Task 5
```

- Tasks 1 and 2 create independent types and can run in parallel
- Task 3 depends on Tasks 1 and 2 (uses the types to build resolution logic)
- Task 4 depends on Task 3 (dimensional validation uses resolved lookup references)
- Task 5 depends on all preceding tasks

## Parallel Execution

- **Group A** (Tasks 1, 2): Disjoint files — `cil_lookup_ref_node.go` vs `analysis_definitions.go`. No shared dependency.

## Reintegration

**Group A (Tasks 1, 2):**
- Merge order: any order (disjoint files)
- Integration tests: `go test ./src/main/go/analysis/... -count=1 -timeout 60s` — verify both types compile together
- Watch for conflicts in: none expected

## Acceptance Criteria Summary

1. `AnalysisInputDefinition.ResolvedLookup` populated when `InputKind == "lookup"`
2. `CILLookupRefNode` implements `CILNode` and stores lookup reference metadata
3. Metric analysis resolves lookup input references against `AnalysisDefinitions.Lookups`
4. Non-existent lookup references produce clear error messages with source locations
5. `validateLookupMetricCompatibility` ensures lookup key dimensions are a subset of metric dimensions
6. Dimensional incompatibility produces clear errors with lookup name and missing dimension name
7. Lookups with grain are correctly handled in dimensional analysis

## Intent Summary for Technical Writer

This story extends the metric analysis phase to handle `lookup` input references. When a metric input has `InputKind == "lookup"`, the analyzer resolves the lookup name against the `AnalysisDefinitions.Lookups` map, validates that the lookup's key dimensions are available in the metric's dimensional context, and generates a `CILLookupRefNode` for use in the metric's CIL tree. The dimensional compatibility check is a preliminary structural check — full unit algebra verification is deferred to later synthesis-time analysis. Grain dimensions on lookups are carried through for the synthesis phase to handle time-indexed lookup access.
