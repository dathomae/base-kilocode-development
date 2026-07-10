# Plan: Story260 — Aggregators Increment 4 (Analyze)

## Purpose

This plan provides the technical writer with sufficient detail to create Story260: **aggregators-increment4-analyze** — implement dimension mismatch detection during metric DAG analysis, query the Aggregators map for applicable transformations, resolve single unambiguous aggregator sequences, and insert `CILAggregateNode` nodes into affected metric CIL trees (Increment 4 of Epic007).

## Story Metadata (for add-new-story.pdd.script.md)

| Parameter    | Value                                                                                        |
| ------------ | -------------------------------------------------------------------------------------------- |
| story_name   | `aggregators-increment4-analyze`                                                              |
| description  | Implement dimension mismatch detection during metric DAG analysis, Aggregators map query, single unambiguous sequence resolution, CIL aggregate node insertion into metric CIL trees |

## Source EROs

- [ERO007: Aggregators Incremental Implementation](../specifications/design/cpl/eros/ERO007_aggregators-incremental-implementation.md) — Increment 4 sketch
- [ERO007: Aggregators Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO007_aggregators-analysis-phase.md) — `resolveMetricDimensionality`, `findSingleAggregatorSequence`, `buildAggregateNode`, dimension mismatch detection

## Dependencies

- Story256 (aggregators-increment2-analyze) — needs `Aggregators` map with populated `AggregatorInfo` entries
- Story258 (aggregators-increment3-analyze) — needs full `FunctionRegistry` and typed CIL node types for building aggregate nodes
- Epic005 (metrics) — needs metric DAG analysis infrastructure (`AnalysisMetricDefinition`, input resolution, IO spec dimensions)

## Files to Create or Modify

| File                                                    | Action  | Purpose                                                                          |
| ------------------------------------------------------- | ------- | -------------------------------------------------------------------------------- |
| `src/main/go/analysis/analyze_aggregators.go`           | Modify  | Add `resolveMetricDimensionality`, `findSingleAggregatorSequence`, `buildAggregateNode` |
| `src/main/go/analysis/cpl_analyzer.go`                  | Modify  | Call `resolveMetricDimensionality` during second pass after metric DAG assembly  |
| `src/main/go/analysis/analyze_aggregators_test.go`      | Modify  | Extend tests: dimension mismatch detection, single sequence resolution, error cases |
| `src/main/go/analysis/cil_aggregate_node.go`            | Modify  | May retain or be replaced by typed nodes; `buildAggregateNode` creates typed nodes |

### resolveMetricDimensionality Algorithm

Called during the second pass of metric analysis after the full metric DAG is assembled. For each metric input:

1. **Determine output dimensions** from the metric's IO spec (required output dimensionality)
2. **For each `metric`-kind input**: get the input metric's resolved dimensions
3. **Find excess dimensions**: concepts present in input but absent from output
4. **If no excess dimensions**: skip (dimensions are compatible)
5. **If excess dimensions exist**: call `findSingleAggregatorSequence(inputDims, outputDims, excessDims)` to find the unique valid aggregator sequence that bridges the gap
6. **Insert aggregate nodes**: for each aggregator in the sequence, wrap the input's CIL tree in a `CILAggregateNode` chain

```go
func (a *CPLAnalyzer) resolveMetricDimensionality(metric *AnalysisMetricDefinition) error {
    outputDims := a.getMetricOutputDimensions(metric)
    
    for _, input := range metric.Inputs {
        if input.InputKind != "metric" {
            continue // Lookup and dimension inputs don't need dimension adjustment
        }
        
        inputDims := input.ResolvedMetric.Dimensions
        
        // Find dimensions in input not in output
        var excessDims []string
        for _, dim := range inputDims {
            if !contains(outputDims, dim) {
                excessDims = append(excessDims, dim)
            }
        }
        
        if len(excessDims) == 0 {
            continue // No mismatch
        }
        
        // For Increment 4: only single unambiguous sequences
        sequence, err := a.findSingleAggregatorSequence(inputDims, outputDims, excessDims)
        if err != nil {
            return fmt.Errorf("metric %s input %s: %w", metric.Name, input.Name, err)
        }
        
        // Create chain of aggregate nodes wrapping input CIL
        currentCIL := input.CILTree
        for _, aggKey := range sequence {
            aggInfo := a.aggregators[aggKey]
            aggNode := a.buildAggregateNode(aggInfo, currentCIL)
            currentCIL = aggNode
        }
        input.CILTree = currentCIL
    }
    return nil
}
```

### findSingleAggregatorSequence (Increment 4 — Single Only)

For a single excess dimension:
1. Build the Aggregators map key: `AggregatorMapKey(inputDims, excessDim)`
2. Look up in `a.aggregators` map
3. If found: return single-element sequence `[key]`
4. If not found: return error "no aggregator found to reduce [X] to [Y]"

For multiple excess dimensions (Increment 4 limitation):
- Return error: "multiple excess dimensions require sequence enumeration (Increment 5)"

### buildAggregateNode

Creates the appropriate typed CIL node wrapping the source CIL tree:

```go
func (a *CPLAnalyzer) buildAggregateNode(info *AggregatorInfo, sourceCIL CILNode) CILNode {
    collapsedValues := a.getCollapsedConceptValues(info)
    targetValues := a.getTargetConceptValues(info)
    
    switch info.CellOperation {
    case "sum":
        return &CILAggregateSumNode{
            SourceCIL: sourceCIL, DimName: info.getCollapsedConceptName(),
            Matrix: BuildSumAggregationMatrix(collapsedValues, targetValues),
        }
    case "count":
        return &CILAggregateCountNode{...}
    case "size":
        return &CILAggregateSizeNode{
            DimName: info.getCollapsedConceptName(),
            Size:    len(collapsedValues),
        }
    case "1/sum":
        return &CILAggregateInverseSumNode{...}
    case "1/count":
        return &CILAggregateInverseCountNode{...}
    }
}
```

### Second Pass Integration

The second analysis pass (`cpl_analyzer.go`) runs after all metrics and aggregators are individually analyzed. After `AnalyzeMetricDefinitions` has resolved all input metric references and built the DAG:
1. Iterate all metrics
2. For each metric, call `resolveMetricDimensionality`
3. Any dimension mismatches that cannot be resolved produce compile-time errors

### Validation Rules Added

- **No applicable aggregator**: If a dimension mismatch exists but no aggregator bridges the gap → compile-time error with descriptive message showing available dimensions and required dimensions
- **Multiple sequences**: In Increment 4, multiple possible aggregator sequences are a compile-time error (user must simplify dimension structure or add more specific aggregators)
- **Missing Concepts**: If concepts in aggregator's from/to don't exist in the metric's dimension set → error

### Test Cases

- **Single dimension reduction**: metric input [Country, Product], metric output [Product], one aggregator `from [Country, Product] to [Product]` → sequence resolved, aggregate node inserted
- **No aggregator found**: metric input [Country, Region, Product], output [Product], no aggregator for Region → error
- **Multiple sequences**: metric input [Country, Region, Product], output [Product], multiple sequence options → error ("multiple sequences require Increment 5")
- **No mismatch**: metric input [Product], output [Product] → no aggregation needed, CIL tree unchanged
- **Scalar reduction**: input [Country], output [] (scalar), aggregator `from [Country] to []` → sequence resolved

## Task Decomposition (for technical writer)

### Task 1: Implement findSingleAggregatorSequence — Not Started
   a. Implement `findSingleAggregatorSequence(inputDims, outputDims, excessDims)` in `analyze_aggregators.go` — Not Started
   b. Handle single excess dimension: build key, query Aggregators map, return sequence (or error) — Not Started
   c. Handle multiple excess dimensions: return deferred error message — Not Started
   d. Write tests: single excess found, single excess not found, multiple excess → error, no excess → nil — Not Started
   e. Verify tests pass — Not Started

### Task 2: Implement buildAggregateNode for typed CIL nodes — Not Started
   a. Implement `buildAggregateNode(info *AggregatorInfo, sourceCIL CILNode) CILNode` — Not Started
   b. Dispatch on `CellOperation` to create correct typed node with matrix built from concept values — Not Started
   c. Write tests: `sum` → `CILAggregateSumNode` with correct matrix, `size` → `CILAggregateSizeNode` with correct constant, etc. — Not Started
   d. Verify tests pass — Not Started

### Task 3: Implement resolveMetricDimensionality and integrate into second pass — Not Started
   a. Implement `resolveMetricDimensionality` — iterate metric inputs, detect mismatches, call findSequence + buildAggregateNode — Not Started
   b. Implement `getMetricOutputDimensions` — extract dimensions from IO spec — Not Started
   c. Integrate into `cpl_analyzer.go` second pass: after metric DAG assembly, iterate metrics calling `resolveMetricDimensionality` — Not Started
   d. Write integration tests: full pipeline with metric that has dimension mismatch → CIL tree contains aggregate nodes wrapping input — Not Started
   e. Write error tests: no aggregator found produces descriptive compile-time error with dimension info — Not Started
   f. Verify all tests pass — Not Started

### Task 4: Handle scalar reduction edge case — Not Started [Small]
   a. Handle `to []` (scalar) aggregation: output has no concept dimensions, aggregator key uses `[]` as to target — Not Started
   b. Write test: metric output is scalar, input has [Country], aggregator `from [Country] to [] = sum(Country)` → sequence resolved — Not Started
   c. Verify test passes — Not Started

## Execution Order

```
[Task 1 ‖ Task 2] → Task 3 → Task 4
```

- Tasks 1 and 2 are parallel (disjoint functions: sequence finder vs node builder)
- Task 3 depends on Tasks 1 and 2 (uses both)
- Task 4 is a refinement on Task 3 (scalar edge case)

## Parallel Execution

- **Group A** (Tasks 1, 2): Task 1 implements `findSingleAggregatorSequence`; Task 2 implements `buildAggregateNode`. Disjoint functions in same file. Worktree isolation required.

## Reintegration

**Group A (Tasks 1, 2):**
- Merge order: any order (separate functions with no mutual dependency)
- Integration tests: `go test ./src/main/go/analysis/... -count=1 -timeout 60s`
- Watch for conflicts in: `analyze_aggregators.go` (both tasks add functions to the same file)

## Acceptance Criteria Summary

1. `findSingleAggregatorSequence` resolves single excess dimension by querying Aggregators map
2. Multiple excess dimensions produce clear deferred error message
3. `buildAggregateNode` creates correct typed CIL node with pre-built matrix from concept values
4. `resolveMetricDimensionality` iterates metric inputs, detects mismatches, and inserts aggregate node chains into CIL trees
5. Second analysis pass calls `resolveMetricDimensionality` for all metrics after DAG assembly
6. Dimension mismatches with no applicable aggregator produce compile-time errors with descriptive messages showing available vs required dimensions
7. Metrics with no dimension mismatch have unmodified CIL trees
8. Scalar reduction (`to []`) is correctly handled
9. All tests pass with `go test ./src/main/go/analysis/...`

## Intent Summary for Technical Writer

This story delivers automatic aggregator integration into the metric dependency DAG. The analysis phase now detects dimension mismatches between metric inputs and outputs after the full DAG is assembled. When a single excess dimension exists and an aggregator entry is found in the Aggregators map, a `CILAggregateNode` is inserted into the metric's CIL tree, wrapping the input reference. This transforms the compilation pipeline from "aggregators are declared and validated but not applied" to "aggregators are automatically applied when needed." In Increment 4, only single unambiguous sequences are supported — if multiple aggregator paths could bridge the dimensional gap, an error is produced. This limitation avoids the complexity of sequence selection while still enabling the common case of one-dimensional aggregation. The synthesis phase (Story261) then walks the modified CIL trees with inserted aggregate nodes to produce computation code.
