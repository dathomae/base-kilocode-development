# Plan: Story262 — Aggregators Increment 5 (Analyze)

## Purpose

This plan provides the technical writer with sufficient detail to create Story262: **aggregators-increment5-analyze** — implement full sequence enumeration, the three-tier selection algorithm (least-information-loss → minimal-work → arbitrary deterministic), ambiguous aggregator resolution via specificity comparison, and multi-step aggregator sequence support (Increment 5 of Epic007).

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `aggregators-increment5-analyze`                                                              |
| description  | Implement full sequence enumeration, selection algorithm (least-info-loss → minimal-work → arbitrary), ambiguous aggregator resolution, specificity comparison |

## Source EROs

- [ERO007: Aggregators Incremental Implementation](../specifications/design/cpl/eros/ERO007_aggregators-incremental-implementation.md) — Increment 5 sketch
- [ERO007: Aggregators Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO007_aggregators-analysis-phase.md) — `enumerateAggregatorSequences`, `selectOptimalSequence`, `sequenceInformationLoss`, `resolveAmbiguousAggregators`

## Dependencies

- Story260 (aggregators-increment4-analyze) — extends `findSingleAggregatorSequence` to `enumerateAggregatorSequences` + `selectOptimalSequence`
- Story256 (aggregators-increment2-analyze) — needs `Aggregators` map, `AggregatorInfo`, `AggregationMatrix`
- Epic005 (metrics) — needs concept value sets for information loss computation

## Files to Create or Modify

| File                                                    | Action  | Purpose                                                                          |
| ------------------------------------------------------- | ------- | -------------------------------------------------------------------------------- |
| `src/main/go/analysis/analyze_aggregators.go`           | Modify  | Add `enumerateAggregatorSequences`, `selectOptimalSequence`, `sequenceInformationLoss`; replace `findSingleAggregatorSequence` usage with full enumeration |
| `src/main/go/analysis/analyze_aggregators_test.go`      | Modify  | Extend tests: multi-step sequences, selection algorithm tiebreakers, ambiguous aggregator resolution |

### enumerateAggregatorSequences

Replaces `findSingleAggregatorSequence` from Increment 4. Finds ALL valid sequences of aggregators that reduce `fromDims` to `targetDims`:

```go
func (a *CPLAnalyzer) enumerateAggregatorSequences(fromDims []string, targetDims []string) ([][]string, error) {
    toRemove := setDifference(fromDims, targetDims)
    if len(toRemove) == 0 {
        return nil, nil // No reduction needed
    }
    
    var sequences [][]string
    a.enumerateRecursive(fromDims, targetDims, toRemove, []string{}, &sequences)
    
    if len(sequences) == 0 {
        return nil, fmt.Errorf("no aggregator sequence found to reduce %v to %v", fromDims, targetDims)
    }
    return sequences, nil
}
```

Recursive enumeration:
1. For each concept in `remainingToRemove`: check if an aggregator exists with key `AggregatorMapKey(currentDims, concept)`
2. If found: compute new dimensions (remove concept from currentDims), recurse with reduced remaining set
3. If no remaining concepts to remove AND new dimensions are a subset of target ↔ store the sequence
4. Base case: `len(newRemaining) == 0 && isSubset(targetDims, newDims)` → valid complete sequence

### selectOptimalSequence — Three-Tier Selection

```go
func (a *CPLAnalyzer) selectOptimalSequence(sequences [][]string) []string {
    if len(sequences) <= 1 { return sequences[0] }
    
    sort.Slice(sequences, func(i, j int) bool {
        // Tier 1: Least information loss
        lossI := a.sequenceInformationLoss(sequences[i])
        lossJ := a.sequenceInformationLoss(sequences[j])
        if lossI != lossJ { return lossI < lossJ }
        
        // Tier 2: Minimal work (shorter sequence = fewer operations)
        if len(sequences[i]) != len(sequences[j]) { return len(sequences[i]) < len(sequences[j]) }
        
        // Tier 3: Deterministic tiebreaker (lexicographic by key)
        return sequences[i][0] < sequences[j][0]
    })
    
    return sequences[0]
}
```

**Tier 1 — Least Information Loss**: Each aggregator step loses information proportional to the number of distinct values in the collapsed dimension. Lower total = better. Computed as: `Σ len(collapsedConcept.ValueOrdering)` for each aggregator key in the sequence.

**Tier 2 — Minimal Work**: If multiple sequences have equal information loss, prefer the shorter sequence (fewer `ReduceDimension` calls).

**Tier 3 — Arbitrary Deterministic**: If still tied, sort lexicographically by the first aggregator key. Ensures deterministic results across compilations.

### resolveAmbiguousAggregators

Handles cases where multiple aggregator declarations share the same `(from, to)` key:

```go
func (a *CPLAnalyzer) resolveAmbiguousAggregators() error {
    groups := make(map[string][]*AggregatorInfo)
    for key, info := range a.aggregators {
        groups[key] = append(groups[key], info)
    }
    
    for key, infos := range groups {
        if len(infos) == 1 { continue }
        
        // Specificity = number of concepts in from list (more = more specific)
        maxSpecificity := 0
        var mostSpecific *AggregatorInfo
        tie := false
        
        for _, info := range infos {
            specificity := len(info.FromConcepts)
            if specificity > maxSpecificity {
                maxSpecificity = specificity
                mostSpecific = info
                tie = false
            } else if specificity == maxSpecificity {
                tie = true
            }
        }
        
        if tie {
            return fmt.Errorf("ambiguous aggregators for pattern %s: multiple with specificity %d", key, maxSpecificity)
        }
        a.aggregators[key] = mostSpecific
    }
    return nil
}

```

Called during analysis after all individual aggregators are analyzed and before dimension mismatch resolution. More concepts in `from` set = more specific (narrower match criteria).

### Integration Changes

In `resolveMetricDimensionality`, replace:
- `findSingleAggregatorSequence(inputDims, outputDims, excessDims)` (Increment 4)
- with: `enumerateAggregatorSequences(inputDims, outputDims)` → `selectOptimalSequence(sequences)`

The multi-excess-dimension limitation from Increment 4 is now removed.

### Test Cases

- **Two-step sequence**: dims [Country, Region, Product] → [Product], aggregators for collapsing Country and Region → enumeration finds both orders, selection picks optimal
- **Tier 1 tiebreaker**: Country has 2 values, Region has 10 values → prefer collapsing Country first (less info loss)
- **Tier 2 tiebreaker**: Two sequences with equal info loss but different lengths → shorter sequence wins
- **Tier 3 tiebreaker**: Two sequences identical in tiers 1 and 2 → deterministic lexicographic order
- **Ambiguous aggregators**: Two aggregators with same (from, to) key, one with more specific from set → more specific wins
- **Ambiguous tie**: Two equally specific aggregators for same key → error
- **No sequence found**: No valid path from input dims to output dims → error

## Task Decomposition (for technical writer)

### Task 1: Implement sequence enumeration — Not Started
   a. Implement `enumerateAggregatorSequences` with recursive helper in `analyze_aggregators.go` — Not Started
   b. Implement `setDifference` and `isSubset` helpers — Not Started
   c. Write tests: single-step sequence, two-step sequence (both orderings enumerated), no sequence found, no reduction needed (empty dims) — Not Started
   d. Verify tests pass — Not Started

### Task 2: Implement selection algorithm and information loss — Not Started
   a. Implement `sequenceInformationLoss` — sum of collapsed concept value counts — Not Started
   b. Implement `selectOptimalSequence` with three-tier sort — Not Started
   c. Write tests: tier 1 (info loss comparison), tier 2 (sequence length comparison), tier 3 (lexicographic tiebreaker) — Not Started
   d. Verify tests pass — Not Started

### Task 3: Implement ambiguous aggregator resolution — Not Started
   a. Implement `resolveAmbiguousAggregators` — group by key, compare specificity, resolve ties — Not Started
   b. Call `resolveAmbiguousAggregators` after all aggregators analyzed, before metric dimensionality resolution — Not Started
   c. Write tests: single aggregator (no ambiguity), two aggregators (most specific wins), equal specificity → error — Not Started
   d. Verify tests pass — Not Started

### Task 4: Integrate full enumeration into resolveMetricDimensionality — Not Started
   a. Replace `findSingleAggregatorSequence` call with `enumerateAggregatorSequences` + `selectOptimalSequence` in `resolveMetricDimensionality` — Not Started
   b. Remove Increment 4's "multiple excess dimensions" deferred error — Not Started
   c. Write integration tests: multi-step dimension reduction with optimal sequence selection, no-sequence-found error — Not Started
   d. Verify all analysis tests pass — Not Started

## Execution Order

```
Task 1 → [Task 2 ‖ Task 3] → Task 4
```

- Task 1 must complete first (enumeration is foundational)
- Tasks 2 and 3 are parallel: Task 2 uses enumerated sequences; Task 3 operates on the Aggregators map independently
- Task 4 depends on all preceding tasks (integrates enumeration + selection into resolution)

## Parallel Execution

- **Group A** (Tasks 2, 3): Task 2 implements selection (consumes sequence lists); Task 3 implements ambiguity resolution (consumes Aggregators map). Disjoint functions, both in `analyze_aggregators.go`.

## Reintegration

**Group A (Tasks 2, 3):**
- Merge order: any order (both add independent functions to the same file)
- Integration tests: `go test ./src/main/go/analysis/... -count=1 -timeout 60s`
- Watch for conflicts in: `analyze_aggregators.go` (both tasks add functions)

## Acceptance Criteria Summary

1. `enumerateAggregatorSequences` finds ALL valid aggregator sequences that bridge `fromDims` → `targetDims`
2. Single-step and multi-step sequences (both orderings) are correctly enumerated
3. `selectOptimalSequence` picks the sequence with least information loss (Tier 1)
4. When info loss is equal, shorter sequences are preferred (Tier 2)
5. When tiers 1+2 are equal, lexicographic tiebreaker produces deterministic results (Tier 3)
6. `resolveAmbiguousAggregators` selects the more specific aggregator (more concepts in `from`)
7. Equally specific ambiguous aggregators produce a clear error
8. `resolveMetricDimensionality` now handles multi-step dimension reductions automatically
9. No-sequence-found errors include descriptive messages with dims
10. All tests pass with `go test ./src/main/go/analysis/...`

## Intent Summary for Technical Writer

This story removes Increment 4's limitation of single-dimension-only aggregator sequences. It implements full recursive sequence enumeration that finds all possible multi-step aggregator chains, then selects the optimal one using a three-tier algorithm. The least-information-loss principle ensures the framework preserves as much data granularity as possible by collapsing dimensions with fewer distinct values first. The ambiguous aggregator resolution handles the case where multiple aggregator declarations share the same dimensional pattern — the more specific one (more concepts in `from`) wins. This completes the analysis-phase aggregator logic, giving the framework the intelligence to automatically determine which aggregators to apply and in what order, without user specification.
