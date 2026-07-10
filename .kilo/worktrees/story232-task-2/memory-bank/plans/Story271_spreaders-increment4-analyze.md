# Plan: Story271 — Spreaders Increment 4 (Analyze)

## Purpose

This plan provides the technical writer with sufficient detail to create Story271: **spreaders-increment4-analyze** — implement the full selection algorithm for mixed aggregator/spreader sequences, ambiguous spreader resolution via specificity comparison, and multi-step interleaved sequence optimization (Increment 4 of Epic008). No parsing changes required.

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `spreaders-increment4-analyze`                                                               |
| description  | Implement mixed aggregator/spreader sequence enumeration, full selection algorithm (least-info-loss → minimal-work → arbitrary), ambiguous spreader resolution, specificity comparison |

## Source EROs

- [ERO008: Spreaders Incremental Implementation](../specifications/design/cpl/eros/ERO008_spreaders-incremental-implementation.md) — Increment 4 sketch
- [ERO008: Spreaders Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO008_spreaders-analysis-phase.md) — Increment 4 data structures and algorithms

## Dependencies

- Story269 (spreaders-increment3-analyze) — metric reference resolution must be complete; spreader analysis infrastructure fully built
- Story267 (spreaders-increment2-analyze) — single-unambiguous sequence finding must exist; Increment 4 extends this to handle multiple candidates and interleaving
- Story260 (aggregators-increment4-analyze) — aggregator selection algorithm and aggregator integration must be complete; mixed sequences require both aggregator and spreader resolution
- Epic005 (metrics) — metric DAG analysis infrastructure must be fully functional for both aggregator and spreader insertion

## Files to Create or Modify

| File                                                    | Action | Purpose                                                                          |
| ------------------------------------------------------- | ------ | -------------------------------------------------------------------------------- |
| `src/main/go/analysis/mixed_sequence_selector.go`       | Create | `enumerateMixedSequences`, full selection algorithm, specificity comparison       |
| `src/main/go/analysis/spreader_dag_integration.go`      | Modify | Update `findSingleSpreaderSequence` → `findOptimalSpreaderSequence` for Increment 4 |
| `src/main/go/analysis/mixed_sequence_selector_test.go`  | Create | Unit tests: sequence enumeration, selection criteria, specificity, ambiguous resolution |

### Full Selection Algorithm

When multiple spreader/aggregator sequences can bridge the dimensional gap, the algorithm ranks them by three criteria (in priority order):

```go
func (a *CPLAnalyzer) selectOptimalSequence(
    sequences [][]DimensionalOp,
) ([]DimensionalOp, error) {

    if len(sequences) == 0 {
        return nil, fmt.Errorf("no valid spreader/aggregator sequence found")
    }

    // Sort by selection criteria (lower score = better)
    sort.Slice(sequences, func(i, j int) bool {
        scoreI := a.sequenceScore(sequences[i])
        scoreJ := a.sequenceScore(sequences[j])
        return scoreI < scoreJ
    })

    return sequences[0], nil
}

func (a *CPLAnalyzer) sequenceScore(seq []DimensionalOp) int {
    infoLoss := a.computeInformationLoss(seq)     // Priority 1
    workCost := a.computeWorkCost(seq)             // Priority 2
    return infoLoss*1000000 + workCost  // Weight info loss much higher
}
```

### Criterion 1: Least Information Loss

Prefer sequences that preserve more dimension information:
- Spreading as late as possible preserves information (more dimensions in intermediate results)
- Aggregating as early as possible preserves information (removes dimensions only after needed)

```go
func (a *CPLAnalyzer) computeInformationLoss(seq []DimensionalOp) int {
    loss := 0
    currentDims := len(seq[0].FromDimensions)
    for _, op := range seq {
        switch op.OpType {
        case "aggregate":
            loss += currentDims  // Earlier aggregation loses more info
        case "spread":
            currentDims++        // Spreading adds dimensions
        }
        currentDims = len(op.ToDimensions)
    }
    return loss
}
```

### Criterion 2: Minimal Work

When information loss is equal, prefer sequences with less computational work:
- Fewer operations are preferred
- Smaller intermediate tensors are preferred
- Matrix multiplications with smaller matrices are preferred

```go
func (a *CPLAnalyzer) computeWorkCost(seq []DimensionalOp) int {
    cost := 0
    for _, op := range seq {
        cost += op.Matrix.Rows * op.Matrix.Columns  // Matrix multiplication cost
        if op.OpType == "spread" {
            cost *= 2  // Expansion is slightly more work than reduction
        }
    }
    return cost
}
```

### Criterion 3: Deterministic Tiebreaker

If still equivalent, select arbitrarily but deterministically (e.g., by sorted operation key).

### Mixed Sequence Enumeration

Generate all valid interleaved aggregator/spreader sequences that transform `fromDims` to `toDims`:

```go
func (a *CPLAnalyzer) enumerateMixedSequences(
    fromDims, toDims []*AnalysisConceptDefinition,
) ([][]DimensionalOp, error) {

    // Determine dimensions to remove (via aggregators) and add (via spreaders)
    toRemove := setDifference(fromDims, toDims)
    toAdd := setDifference(toDims, fromDims)

    var sequences [][]DimensionalOp

    // Enumerate all valid interleavings of aggregator and spreader operations
    // Constraint: can only aggregate dimensions that currently exist
    // Constraint: can only spread to add dimensions that don't currently exist
    a.enumerateMixedRecursive(fromDims, toDims, toRemove, toAdd,
        nil /* current ops */, &sequences)

    return sequences, nil
}
```

### DimensionalOp Type

```go
type DimensionalOp struct {
    OpType          string              // "aggregate" or "spread"
    OpKey           string              // Map key for the operation
    FromDimensions  []string
    ToDimensions    []string
    Matrix          *AggregationMatrix
    CILNode         CILNode             // CILAggregateNode or CILSpreadNode
    NewDimension    string              // For spread: the dimension being added
    CollapsedDim    string              // For aggregate: the dimension being collapsed
}
```

### Ambiguous Spreader Resolution

When multiple spreaders share the same `(from, to)` dimensional pattern (same map key), resolve by specificity:

```go
func (a *CPLAnalyzer) resolveAmbiguousSpreaders() error {
    groups := make(map[string][]*SpreaderInfo)
    for key, info := range a.spreaders {
        groups[key] = append(groups[key], info)
    }

    for key, infos := range groups {
        if len(infos) == 1 { continue }

        maxSpec := -1
        var best *SpreaderInfo
        tie := false
        for _, info := range infos {
            spec := len(info.FromConcepts)  // More from concepts = more specific
            if spec > maxSpec {
                maxSpec = spec
                best = info
                tie = false
            } else if spec == maxSpec {
                tie = true
            }
        }
        if tie {
            return fmt.Errorf(
                "ambiguous spreaders for pattern %s: multiple spreaders with equal specificity (%d)",
                key, maxSpec,
            )
        }
        // Replace with the most specific
        a.spreaders[key] = best
    }
    return nil
}
```

The logic: a spreader `from [Country, Product] to [Country, Product, Region]` is more specific than `from [Country] to [Country, Product, Region]` because it requires more known dimensions. More specific = better dimensional fit.

### Integration with DAG Analysis

The `resolveMetricDimensionalitySpreaders` function from Increment 2 is updated:

1. If a single unambiguous sequence exists → use it (Increment 2 behavior)
2. If multiple valid sequences exist → enumerate all, apply selection algorithm, insert optimal sequence
3. If no sequence exists → error

```go
func (a *CPLAnalyzer) resolveMetricDimensionalitySpreaders(
    metric *AnalysisMetricDefinition,
) error {
    // ... dimension mismatch detection ...

    sequences, err := a.enumerateMixedSequences(inputDims, outputDims)
    if err != nil {
        return err
    }

    optimalSeq, err := a.selectOptimalSequence(sequences)
    if err != nil {
        return err
    }

    // Chain CIL nodes for the optimal sequence
    currentCIL := input.CILTree
    for _, op := range optimalSeq {
        switch op.OpType {
        case "spread":
            // Insert CILSpreadNode
        case "aggregate":
            // Insert CILAggregateNode
        }
        currentCIL = op.CILNode
    }
    metric.Inputs[i].CILTree = currentCIL
}
```

### Validation During Analysis (Increment 4 — Additions)

1. **At least one valid sequence**: If no sequence transforms from dims to to dims, error
2. **Selection determinism**: Same inputs always produce the same optimal sequence
3. **Information loss tie-breaking**: When info loss is equal, work cost differentiates
4. **Specificity resolution**: When specificity is equal and maps to same key, error (cannot disambiguate)
5. **Fully interleaved sequences**: Sequences may mix aggregator and spreader operations in any order that respects dimensional constraints

### Limitations (Increment 4)

- Fused operations (combining multiple matrices into one) handled in synthesis (Story272)
- Large state space for many dimensions may need pruning (academic concern for initial implementation)
- No heuristic tuning for work cost weights (uses simple matrix dimension product)

### Test Cases

1. **Single sequence (regression)**: When only one sequence exists, Increment 4 behavior matches Increment 2
2. **Multiple sequences — info loss**: Two sequences, one spreads early (higher info loss), one spreads late (lower) → late spread selected
3. **Multiple sequences — work cost**: Two sequences with equal info loss but different work costs → lower work cost selected
4. **Mixed aggregator/spreader**: Input `[Country, Product]`, output `[Region]` — sequence: aggregate Country then spread to Region; or: spread to Region then aggregate Country → least-info-loss selected
5. **Ambiguous spreaders — specificity**: Two spreaders with same `(from, to)` but different specificity → most specific wins
6. **Ambiguous spreaders — tie**: Two spreaders with same specificity and same pattern → error
7. **No valid sequence**: Input `[Country]`, output `[Region]` with no spreaders or aggregators that can bridge → error
8. **Multi-step interleaved**: Input `[A,B]`, output `[C,D]` — aggregate A, spread to C, aggregate B, spread to D → correct interleaving
9. **Deterministic selection**: Run selection multiple times → same result
10. **Integration with metric DAG**: Multiple metrics in DAG, each with different sequence requirements → all correctly resolved

## Task Decomposition (for technical writer)

### Task 1: Implement mixed sequence enumeration — Not Started
   a. Create `src/main/go/analysis/mixed_sequence_selector.go` — define `DimensionalOp` struct — Not Started
   b. Implement `setDifference` (dims in A not in B, vice versa) — Not Started
   c. Implement `enumerateMixedRecursive` — depth-first enumeration of valid interleaved aggregator/spreader sequences — Not Started
   d. Constraint: only aggregate existing dims, only spread to add missing dims — Not Started
   e. Write unit tests: enumerate for simple 1-step cases, multi-step interleaved, verify all returned sequences are valid — Not Started
   f. Verify enumeration tests pass — Not Started

### Task 2: Implement full selection algorithm — Not Started
   a. Implement `computeInformationLoss` — earlier aggregation = higher loss, later spread = less loss — Not Started
   b. Implement `computeWorkCost` — matrix dimensions product as proxy for computational cost — Not Started
   c. Implement `sequenceScore` — combines info loss and work cost with weighted priority — Not Started
   d. Implement `selectOptimalSequence` — sort by score, return best — Not Started
   e. Write unit tests: verify info loss ranking (late spread beats early spread), verify work cost tie-breaking, verify deterministic selection, verify optimal sequence for mixed aggregator/spreader — Not Started
   f. Verify selection tests pass — Not Started

### Task 3: Implement ambiguous spreader resolution — Not Started
   a. Implement `resolveAmbiguousSpreaders` — group by map key, select by specificity — Not Started
   b. Specificity = number of concepts in `from` (more = more specific) — Not Started
   c. Equal specificity produces error (cannot disambiguate) — Not Started
   d. Write unit tests: two spreaders with same pattern and different specificity → most specific wins, equal specificity → error, single spreader → no change — Not Started
   e. Verify ambiguity resolution tests pass — Not Started

### Task 4: Integrate Increment 4 into DAG analysis — Not Started
   a. Update `resolveMetricDimensionalitySpreaders` in `spreader_dag_integration.go` to use Increment 4 selection when multiple sequences exist — Not Started
   b. Fall back to Increment 2 single-sequence path when only one valid sequence exists — Not Started
   c. Wire ambiguous spreader resolution into analysis pipeline (runs once before DAG analysis) — Not Started
   d. Write integration tests: metric with multiple valid spreader sequences, metric with mixed aggregator/spreader requirement, metric with interleaved operations — Not Started
   e. Create `parse_and_analyze_only/` examples: mixed sequences, ambiguous spreaders, optimal selection scenarios — Not Started
   f. Verify full analyze pipeline with Increment 4 features

### Task 5: Regression tests — Not Started [Small]
   a. Verify Increment 2 behavior unchanged (single unambiguous sequences still work) — Not Started
   b. Verify Increment 3 behavior unchanged (metric references in expressions still resolve) — Not Started
   c. Run full test suite — Not Started

## Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4 → Task 5
```

- Task 1 must complete first (enumeration is the foundation)
- Task 2 depends on Task 1 (selection ranks enumerated sequences)
- Task 3 depends on Tasks 1 and 2 (ambiguous resolution may change spreader map before enumeration)
- Task 4 depends on all preceding tasks (integration uses all)
- Task 5 is verification

No parallel groups — each task builds on the previous.

## Acceptance Criteria Summary

1. `enumerateMixedSequences` generates all valid interleaved aggregator/spreader sequences
2. Selection algorithm picks least-information-loss sequence when multiple exist
3. When info loss is equal, minimal-work sequence is selected
4. Selection is deterministic (same inputs → same output)
5. Single unambiguous sequences (Increment 2 behavior) still work as regression
6. Ambiguous spreaders resolved by specificity (more `from` concepts = more specific)
7. Equal-specificity ambiguous spreaders produce error
8. Mixed aggregator/spreader sequences are correctly chained in metric CIL trees
9. Interleaved operations maintain dimensional correctness at each step
10. All parse_and_analyze_only examples pass verification
11. No regressions in Increment 1, 2, or 3 behavior

## Intent Summary for Technical Writer

This story implements the final analysis-phase optimization: intelligent sequence selection when multiple spreader/aggregator sequences can satisfy dimensional requirements. The algorithm enumerates all valid interleaved sequences, then ranks them by information loss (primary) and computational work (secondary). Information loss scoring prefers spreading as late as possible and aggregating as early as possible — preserving more dimension information in intermediate results. Work cost uses matrix dimension products as a proxy for computational expense. When multiple spreaders match the same dimensional pattern, specificity comparison (number of `from` concepts) selects the best fit. The resolved sequences are inserted into metric CIL trees as interleaved `CILSpreadNode` and `CILAggregateNode` chains. Increment 2's single-sequence behavior is preserved as a regression path. The synthesis phase (Story272) will implement fused operations to optimize the selected sequences.

Note: The `DimensionalOp` type bridges analysis and synthesis — it carries both the dimensional transformation metadata and the CIL nodes that the synthesizer will walk. The sequence enumeration algorithm must respect the constraint that aggregators can only collapse dimensions that currently exist, and spreaders can only add dimensions that don't currently exist. The state space grows with the number of dimensions, but for typical CPL metrics (3-5 dimensions), enumeration is fast.
