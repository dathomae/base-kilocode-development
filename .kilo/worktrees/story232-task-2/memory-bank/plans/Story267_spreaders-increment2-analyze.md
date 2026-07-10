# Plan: Story267 — Spreaders Increment 2 (Analyze)

## Purpose

This plan provides the technical writer with sufficient detail to create Story267: **spreaders-increment2-analyze** — implement spreader selection and integration into the metric dependency DAG (Increment 2 of Epic008). When a metric's output has more dimensions than its inputs, the analyzer queries the Spreaders map for a single unambiguous spreader sequence and inserts `CILSpreadNode` nodes into metric CIL trees. No parsing changes required.

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `spreaders-increment2-analyze`                                                               |
| description  | Implement dimension mismatch detection during metric DAG analysis, Spreaders map query, single unambiguous sequence resolution, CILSpreadNode insertion into metric CIL trees |

## Source EROs

- [ERO008: Spreaders Incremental Implementation](../specifications/design/cpl/eros/ERO008_spreaders-incremental-implementation.md) — Increment 2 sketch
- [ERO008: Spreaders Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO008_spreaders-analysis-phase.md) — Increment 2 data structures and algorithms

## Dependencies

- Story265 (spreaders-increment1-analyze) — Spreaders map must exist with `SpreaderInfo` entries and `CILSpreadNode` generation
- Story260 (aggregators-increment4-analyze) — same analysis pass pattern (metric DAG assembly → dimension mismatch detection → resolution). The aggregator integration pattern is established; spreader integration follows the same approach.
- Epic005 (metrics) — metric DAG analysis infrastructure must exist; spreaders are applied during the second analysis pass

## Files to Create or Modify

| File                                                    | Action | Purpose                                                                          |
| ------------------------------------------------------- | ------ | -------------------------------------------------------------------------------- |
| `src/main/go/analysis/spreader_dag_integration.go`      | Create | `resolveMetricDimensionalitySpreaders`, `findSingleSpreaderSequence`, `chainSpreadNodes` |
| `src/main/go/analysis/metric_dag_analyzer.go`           | Modify | Add spreader resolution call during metric DAG analysis (second pass)              |
| `src/main/go/analysis/spreader_dag_integration_test.go` | Create | Unit tests: dimension mismatch detection, sequence finding, CIL tree insertion   |

### Dimension Mismatch Detection

During the second analysis pass (after the metric dependency DAG is assembled), the analyzer checks each metric input for dimensional mismatches:

```go
func (a *CPLAnalyzer) resolveMetricDimensionalitySpreaders(
    metric *AnalysisMetricDefinition,
) error {
    outputDims := a.getMetricOutputDimensions(metric)

    for i, input := range metric.Inputs {
        if input.InputKind != "metric" { continue }

        inputDims := input.ResolvedMetric.Dimensions

        // Find dimensions in output that are not in input
        var missingDims []string
        for _, dim := range outputDims {
            if !contains(inputDims, dim.Name) {
                missingDims = append(missingDims, dim.Name)
            }
        }

        if len(missingDims) == 0 { continue }

        // Find single unambiguous spreader sequence
        sequence, err := a.findSingleSpreaderSequence(inputDims, outputDims, missingDims)
        if err != nil {
            return fmt.Errorf("metric %s input %d: %w", metric.Name, i, err)
        }

        // Chain CILSpreadNode instances
        currentCIL := input.CILTree
        for _, spreadKey := range sequence {
            spreadInfo := a.spreaders[spreadKey]
            spreadNode := a.buildSpreadNode(spreadInfo, currentCIL)
            currentCIL = spreadNode
        }
        metric.Inputs[i].CILTree = currentCIL
    }
    return nil
}
```

### Single Unambiguous Sequence Finding

Given a starting dimension set and a target dimension set, find ONE spreader sequence that bridges the gap:

```go
func (a *CPLAnalyzer) findSingleSpreaderSequence(
    fromDims []*AnalysisConceptDefinition,
    toDims []*AnalysisConceptDefinition,
    missingDims []string,
) ([]string, error) {

    // Strategy: add missing dimensions one by one
    // For each missing dimension, find a spreader that adds it while keeping all existing dims
    currentDims := copyDims(fromDims)
    var sequence []string

    for _, missingDim := range missingDims {
        candidates := a.querySpreadersByNewDimension(currentDims, missingDim)

        if len(candidates) == 0 {
            return nil, fmt.Errorf(
                "no spreader found to add dimension %q (current dims: %v)",
                missingDim, dimNames(currentDims),
            )
        }
        if len(candidates) > 1 {
            return nil, fmt.Errorf(
                "ambiguous: multiple spreaders add dimension %q from dims %v — defer to Increment 4",
                missingDim, dimNames(currentDims),
            )
        }

        spreadKey := candidates[0]
        sequence = append(sequence, spreadKey)

        // Update current dims: add the new dimension
        spreadInfo := a.spreaders[spreadKey]
        currentDims = spreadInfo.ToConcepts
    }

    // Verify final dims match target
    if !dimSetsEqual(currentDims, toDims) {
        return nil, fmt.Errorf("spreader sequence produced dims %v, expected %v",
            dimNames(currentDims), dimNames(toDims))
    }

    return sequence, nil
}
```

### Spreader Map Query

```go
// querySpreadersByNewDimension finds all spreaders that add a specific dimension
// while keeping all current dimensions.
func (a *CPLAnalyzer) querySpreadersByNewDimension(
    currentDims []*AnalysisConceptDefinition,
    newDim string,
) []string {
    var candidates []string
    for key, info := range a.spreaders {
        // Check that the spreader's from is a subset of current dims
        if isSubset(info.FromConcepts, currentDims) {
            // Check that the new dim is in the spreader's to but not in current dims
            if containsConcept(info.ToConcepts, newDim) && !containsConcept(currentDims, newDim) {
                candidates = append(candidates, key)
            }
        }
    }
    return candidates
}
```

### CILSpreadNode Insertion

`CILSpreadNode` nodes are inserted between the input metric reference and the calculation body in the metric's CIL tree:

```
Before:  CILMetricRef → CILArithmeticBody
After:   CILMetricRef → CILSpreadNode → CILSpreadNode → CILArithmeticBody
         (one CILSpreadNode per spreader in the sequence)
```

Each `CILSpreadNode` wraps the result of the previous node as its `SourceCIL`.

### Validation During Analysis (Increment 2)

1. **Dimension Mismatch Detection**: Input metric has fewer dimensions than output requires
2. **Spreader Existence**: At least one spreader must exist for each missing dimension
3. **Single Unambiguous Sequence**: Exactly one spreader per missing dimension; multiple candidates → error (deferred to Increment 4)
4. **Spreader Order Independence**: The order of adding missing dimensions must not affect the final dimensional set (if order matters, an error is reported)
5. **From-Subset Check**: Spreader's `from` concepts must be subset of current dimensions at the point of application

### Limitations (Increment 2)

- Only single unambiguous sequences (one spreader per missing dimension); multiple candidates → error
- No interleaved aggregator/spreader sequences (Increment 4)
- No metric references in spreader expressions (Increment 3)
- No sequence optimization (Increment 4)

### Test Cases

- **Simple single spreader**: Input `[Product]`, output `[Country, Product]`, one spreader `from [Product] to [Country, Product]` → sequence inserted
- **Multi-step spreader sequence**: Input `[]` (scalar), output `[Country, Region, Product]` with spreaders `[] → [Country]`, `[Country] → [Country, Region]`, `[Country, Region] → [Country, Region, Product]` → three CILSpreadNode chain
- **No spreader needed**: Input `[Country, Product]`, output `[Country, Product]` → no insertion, passes through
- **Missing spreader**: Input `[Product]`, output `[Country, Region, Product]`, no spreader for `Region` from current dims → error
- **Ambiguous spreaders**: Two spreaders both add dimension `X` from same current dims → error
- **Non-metric inputs**: Connector inputs are skipped (only metric inputs are checked for dimension mismatches)
- **CIL tree integrity**: After insertion, verify CIL tree structure is valid (no broken parent-child links)
- **Multiple inputs with different mismatches**: Each metric input gets its own spreader sequence

## Task Decomposition (for technical writer)

### Task 1: Implement spreader map query and single sequence finding — Not Started
   a. Create `src/main/go/analysis/spreader_dag_integration.go` — implement `querySpreadersByNewDimension`, `findSingleSpreaderSequence` — Not Started
   b. Implement helper functions: `isSubset` (check from concepts subset of current dims), `dimSetsEqual`, `dimNames` — Not Started
   c. Write unit tests: find single spreader for one missing dim, find multi-step sequence, no candidates returns error, multiple candidates returns error, order independence verification — Not Started
   d. Verify sequence tests pass — Not Started

### Task 2: Implement CILSpreadNode insertion into metric CIL trees — Not Started
   a. Implement `chainSpreadNodes` — creates and links `CILSpreadNode` instances, wrapping each spreader's matrix and CIL tree — Not Started
   b. Implement `buildSpreadNode` — constructs a `CILSpreadNode` from `SpreaderInfo` and source CIL, using pre-built matrix and dimension info — Not Started
   c. Write unit tests: verify CIL tree structure after single insert, after multi-step insert, verify SourceCIL links are correct, verify no mutation to original CIL subtree — Not Started
   d. Verify insertion tests pass — Not Started

### Task 3: Integrate spreader resolution into metric DAG analysis — Not Started
   a. Modify `src/main/go/analysis/metric_dag_analyzer.go` — add call to `resolveMetricDimensionalitySpreaders` during the second analysis pass (after metric dependency DAG is assembled) — Not Started
   b. Ensure spreader resolution runs AFTER aggregator resolution (aggregators handle dimension reduction, spreaders handle dimension expansion) but both during the same pass — Not Started
   c. Write integration tests: full metric with one spreader-needed input, full metric with multi-step spreader sequence, metric with no spreader needed, metric with mixed connector and metric inputs — Not Started
   d. Write error case tests: metric with unresolvable dimension mismatch, metric with ambiguous spreader candidates — Not Started
   e. Verify integration tests pass — Not Started

### Task 4: Create analyze-phase examples — Not Started [Small]
   a. Create `parse_and_analyze_only/without_errors/` examples: CPL with spreaders + metrics where single unambiguous sequences exist — Not Started
   b. Create `parse_and_analyze_only/with_errors/` examples: CPL where dimension mismatch exists but no spreader bridges the gap, ambiguous spreader candidates — Not Started
   c. Verify parse→analyze pipeline produces correct errors or successful analysis — Not Started

## Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4
```

- Task 1 must complete first (sequence finding is the core algorithm)
- Task 2 depends on Task 1 (needs sequences to insert)
- Task 3 depends on Tasks 1 and 2 (integration wires sequence finding + insertion into DAG analysis)
- Task 4 depends on all preceding tasks (examples verify full pipeline)

No parallel groups — each task builds on the previous.

## Acceptance Criteria Summary

1. `findSingleSpreaderSequence` finds correct spreader keys for single and multi-step dimensional additions
2. Multiple candidates for the same dimension addition produce an error
3. No spreader existing for a required dimension produces an error
4. `CILSpreadNode` chain correctly wraps intermediate results: each spreader's `SourceCIL` is the previous `CILSpreadNode`
5. Spreader resolution integrated into metric DAG analysis — runs during second pass after DAG assembly
6. Metrics with no dimension mismatches pass through unchanged
7. Metrics with connector inputs are not checked for spreader resolution (only metric inputs)
8. Order independence: adding missing dims in any order produces the same final dimensional set (error if not)
9. All parse_and_analyze_only examples pass verification
10. All tests pass with `go test ./src/main/go/analysis/... -count=1 -timeout 60s`

## Intent Summary for Technical Writer

This story integrates spreaders into the metric dependency DAG analysis. After the metric DAG is assembled (during the second analysis pass), each metric input is checked for dimensional mismatches: if the output has more dimensions than an input, the analyzer queries the Spreaders map for single unambiguous sequences to bridge the gap. The algorithm adds missing dimensions one at a time, each step finding exactly one spreader whose `from` concepts are a subset of current dimensions and whose `to` adds exactly one missing dimension. Multiple candidates trigger an error (deferred to Increment 4). The `CILSpreadNode` instances are linked in a chain and inserted between the input reference and the calculation body in the metric's CIL tree. No parser changes are needed — spreader declarations are already parsed in Increment 1. Only single unambiguous sequences are supported; interleaved aggregator/spreader sequences are deferred to Increment 4.

Note: The spreader resolution must run AFTER aggregator resolution in the same analysis pass. Aggregators handle the case where input has MORE dimensions than output; spreaders handle the case where input has FEWER dimensions than output. Both are dimensional adjustments applied to metric inputs before the calculation body executes.
