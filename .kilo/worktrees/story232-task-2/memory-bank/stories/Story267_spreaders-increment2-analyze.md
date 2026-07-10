# Story267: Spreaders Increment 2 — Analyze

## Goal

Implement spreader input reference resolution in metric calculate clauses. During the second analysis pass (after the metric dependency DAG is assembled), each metric input is checked for dimensional mismatches against the output. When an input has fewer dimensions than the output requires, the analyzer queries the Spreaders map for a single unambiguous spreader sequence, resolves it into a chain of `CILSpreadNode` instances, and inserts them between the input reference and the calculation body in the metric's CIL tree. Multiple candidate spreaders for the same dimension trigger an error (deferred to Increment 4). No parsing changes are required — spreader declarations were already parsed in Increment 1.

This story is part of Epic008 (Spreaders), Increment 2.

## References

- [ERO008: Spreaders Incremental Implementation](../specifications/design/cpl/eros/ERO008_spreaders-incremental-implementation.md)
- [ERO008: Spreaders Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO008_spreaders-analysis-phase.md)
- [Plan: Story267 Spreaders Increment 2](../plans/Story267_spreaders-increment2-analyze.md)

## Dependencies

- Story265 (spreaders-increment1-analyze) — Planned. Spreaders map must exist on `CPLAnalyzer` with `SpreaderInfo` entries (including `FromConcepts`, `ToConcepts`, matrix, and dimension information), `CILSpreadNode` type must be defined with `SourceCIL` field, and spreader declarations must be parsed and registered
- [Story260_aggregators-increment4-analyze.md](Story260_aggregators-increment4-analyze.md) — Not Started. Aggregator integration follows the same analysis-pass pattern; spreader integration uses the same approach. Spreader resolution runs AFTER aggregator resolution in the second pass
- Epic005 (metrics) — Metric DAG analysis infrastructure must exist (`AnalysisMetricDefinition` with resolved inputs, `AnalysisInputDefinition` with `InputKind` and `CILTree` fields); spreaders are applied during the second analysis pass

## Dependent Stories

- [Story269_spreaders-increment3-analyze.md](Story269_spreaders-increment3-analyze.md) — Not Started. Depends on spreader DAG integration being in place for metric reference resolution in spreader expressions

## Tasks

### Test-First Development

This story follows the Logical TDD Lifecycle. For each coding task, write tests before implementing the code. The test-and-implement cycle is integrated within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for details on the Logical TDD Lifecycle.

### Task Sizing

Task 4 is annotated `[Small]` — it creates example subdirectories following existing templates in `examples/parse_and_analyze_only/`, touches only example files (no production code), and addresses a single concern. Tasks 1–3b are not small: Task 1 introduces new algorithmic logic with multiple edge cases (no candidates, multiple candidates, order independence). Task 2 requires understanding CIL tree structure, node linking, and wrapping semantics. Task 3a implements the metric iteration method with dimension comparison logic. Task 3b performs integration wiring into `cpl_analyzer.go`'s `Analyze` function and requires end-to-end pipeline understanding.

### Parallel Execution

None — all tasks have sequential dependencies. Each task builds on the output of the preceding task.

### Execution Order

```
Task 1 → Task 2 → Task 3a → Task 3b → Task 4
```

### Reintegration

None — no parallel groups.

---

### Task 1: Implement spreader map query and single sequence finding — Not Started

Implement the core dimensional analysis functions that query the Spreaders map and find single unambiguous spreader sequences to bridge dimension gaps between metric inputs and outputs.

1. a. Create `src/main/go/analysis/spreader_dag_integration.go` — implement `querySpreadersByNewDimension(currentDims []*AnalysisConceptDefinition, newDim string) []string` and `findSingleSpreaderSequence(fromDims []*AnalysisConceptDefinition, toDims []*AnalysisConceptDefinition, missingDims []string) ([]string, error)`. The query function finds all spreader keys that add a specific new dimension while keeping all current dimensions: the spreader's `FromConcepts` must be a subset of `currentDims`, and the new dimension must be in `ToConcepts` but not in `currentDims`. The sequence finder iterates through `missingDims`, adding them one at a time, and rejects ambiguous cases (multiple candidates for the same dimension) with an error. — Not Started

   b. Implement helper functions in the same file:
      - `isSubset(sub, super []*AnalysisConceptDefinition) bool` — returns true if all concepts in `sub` are present in `super`
      - `dimSetsEqual(a, b []*AnalysisConceptDefinition) bool` — returns true if both dimension sets contain the same concept names
      - `dimNames(dims []*AnalysisConceptDefinition) []string` — extracts dimension name strings from concept definitions for error messages
      - `copyDims(dims []*AnalysisConceptDefinition) []*AnalysisConceptDefinition` — deep-copies a dimension slice for independent mutation during multi-step resolution — Not Started

   c. Write unit tests in `src/main/go/analysis/spreader_dag_integration_test.go`:
      - Find single spreader for one missing dimension — spreader `from [Product] to [Country, Product]` bridges `missingDim = "Country"`, verify correct key returned
      - Find multi-step sequence — spreaders `[] → [Country]`, `[Country] → [Country, Region]`, `[Country, Region] → [Country, Region, Product]`; verify all three keys returned in order
      - No candidates returns error — `missingDim = "Region"` but no spreader adds Region from current dims; verify error message includes the missing dimension name
      - Multiple candidates returns error — two spreaders both add dimension X from the same current dims; verify error message includes "ambiguous" and the dimension name
      - Order independence verification — adding missing dimensions in different orders produces the same final set (verify via two different orderings of the same missing dims produce identical sequences)
      - From-subset check — spreader whose `FromConcepts` is not entirely contained in `currentDims` is not returned as a candidate — Not Started

   d. Run `go test ./src/main/go/analysis/... -count=1 -timeout 60s` and verify all tests pass. — Not Started

**Intent**: Provide the fundamental building blocks for spreader integration: the ability to look up what spreaders are available for a given dimension gap and to resolve a complete spreader sequence that bridges from input dimensions to output dimensions. The algorithm must be strict — ambiguity is an error, not a choice.

**Acceptance Criteria**:
1. `querySpreadersByNewDimension` returns correct spreader keys for a given new dimension and current dimension context
2. `findSingleSpreaderSequence` finds correct spreader keys for single-step and multi-step dimensional additions
3. Multiple candidates for the same dimension addition produce an error (not silently choosing one)
4. No spreader existing for a required dimension produces an error
5. Order independence: adding missing dims in any order produces the same final dimensional set; otherwise an error is reported
6. All unit tests pass with `go test ./src/main/go/analysis/... -count=1 -timeout 60s`

---

### Task 2: Implement CILSpreadNode insertion into metric CIL trees — Not Started

Implement the functions that construct and chain `CILSpreadNode` instances, wrapping each spreader's result as the source CIL for the next node in the chain.

2. a. Implement `chainSpreadNodes(sequence []string, initialCIL cil.CILNode) (cil.CILNode, error)` in `src/main/go/analysis/spreader_dag_integration.go`. For each spreader key in the sequence, look up the `SpreaderInfo` from `a.spreaders`, call `buildSpreadNode` to create a `CILSpreadNode` wrapping the current CIL tree, and advance the current CIL to the new node. The first spreader wraps `initialCIL` as `SourceCIL`; each subsequent spreader wraps the previous `CILSpreadNode`. — Not Started

   b. Implement `buildSpreadNode(info *SpreaderInfo, sourceCIL cil.CILNode) *cil.CILSpreadNode` in `src/main/go/analysis/spreader_dag_integration.go`. Constructs a `CILSpreadNode` with the spreader's pre-built matrix, `FromConcepts`, `ToConcepts`, dimension information, and the provided `sourceCIL` as `SourceCIL`. — Not Started

   c. Write unit tests in `src/main/go/analysis/spreader_dag_integration_test.go`:
      - Single-spreader insertion: `CILMetricRef → CILSpreadNode → CILArithmeticBody`. Verify the spread node's `SourceCIL` is the metric ref, and the arithmetic body follows the spread node.
      - Multi-spreader chain: three spreaders producing `CILMetricRef → CILSpreadNode1 → CILSpreadNode2 → CILSpreadNode3 → CILArithmeticBody`. Verify each spread node's `SourceCIL` points to the correct predecessor.
      - Verify `CILSpreadNode` fields are populated correctly: matrix dimensions, `FromConcepts`, `ToConcepts` match the `SpreaderInfo`.
      - Verify the original CIL subtree passed as `initialCIL` is not mutated by `chainSpreadNodes` — the function returns a new chain that wraps (does not modify) the input. — Not Started

   d. Run `go test ./src/main/go/analysis/... -run "TestChainSpreadNodes|TestBuildSpreadNode" -count=1` and verify all tests pass. — Not Started

**Intent**: Transform a resolved spreader key sequence into actual CIL tree nodes. Each spreader in the sequence becomes a `CILSpreadNode` that wraps the output of the previous step. The resulting chain is inserted between the input metric reference and the calculation body so that at synthesis time the spread operation is applied before the metric body executes.

**Acceptance Criteria**:
1. `chainSpreadNodes` produces a linked chain of `CILSpreadNode` instances for a valid sequence of spreader keys
2. `buildSpreadNode` correctly constructs a `CILSpreadNode` with the spreader's matrix, from/to concepts, and `SourceCIL`
3. Single-spreader insertion: CIL tree structure is `CILMetricRef → CILSpreadNode → CILArithmeticBody`
4. Multi-spreader insertion: chain preserves order, each node's `SourceCIL` points to the correct predecessor
5. The original input CIL subtree is not mutated; the chain is a new structure that wraps the input
6. All unit tests pass with `go test ./src/main/go/analysis/... -count=1 -timeout 60s`

---

### Task 3a: Implement resolveMetricDimensionalitySpreaders with unit tests — Not Started

Implement the dimension-mismatch detection and spreader-chain-wrapping method on `CPLAnalyzer`, isolated and unit-tested with directly constructed test data before integration into the analysis pipeline.

3a. a. Implement `resolveMetricDimensionalitySpreaders(metric *AnalysisMetricDefinition) error` in `src/main/go/analysis/spreader_dag_integration.go`. For each input in the metric's input list:
       - Skip inputs where `InputKind` is not `"metric"` (connector and dimension inputs are not dimension-checked).
       - Extract the input metric's resolved dimensions from `input.ResolvedMetric.Dimensions` and the output dimensions from the metric's IO spec (using `getMetricOutputDimensions`).
       - Compare: find dimensions present in the output but absent from the input (`missingDims`).
       - If `len(missingDims) == 0`, skip this input (no mismatch).
       - Call `findSingleSpreaderSequence(inputDims, outputDims, missingDims)` to resolve the gap.
       - Call `chainSpreadNodes(sequence, input.CILTree)` to build the spreader node chain.
       - Assign the returned chain to `metric.Inputs[i].CILTree` (replace the original CIL tree with the wrapped version). — Not Started

   b. Implement `getMetricOutputDimensions(metric *AnalysisMetricDefinition) []*AnalysisConceptDefinition` in the same file. Extract the required output dimension concepts from the metric's IO spec (the concepts defined in the metric's output structure). — Not Started

   c. Write unit tests in `src/main/go/analysis/spreader_dag_integration_test.go` using directly constructed `AnalysisMetricDefinition` and `AnalysisInputDefinition` objects (no dependency on the full analyzer pipeline):
      - Output dimensions correctly extracted from a metric with multi-dimensional IO spec
      - Input with matching dimensions has unchanged CIL tree (no spreader inserted)
      - Input with one missing dimension and a single spreader: CIL tree is wrapped with one `CILSpreadNode` and the new tree is assigned to `input.CILTree`
      - Input with multiple missing dimensions and a multi-step spreader chain: CIL tree is wrapped with the full chain
      - Connector inputs (`InputKind != "metric"`) are skipped — their CIL trees are unmodified
      - Mixed inputs: one metric input (gets spreader chain) and one connector input (untouched)
      - Unresolvable mismatch (no spreader bridges the gap) returns an error with dimension names
      - Ambiguous spreader candidates return an error forwarded from `findSingleSpreaderSequence` — Not Started

   d. Run `go test ./src/main/go/analysis/... -run "TestResolveMetricDimensionalitySpreaders|TestGetMetricOutputDimensions" -count=1` and verify all tests pass. — Not Started

**Intent**: Implement the core dimension-mismatch detection and CIL-tree-wrapping logic as a self-contained method on `CPLAnalyzer`. This method iterates a metric's inputs, identifies dimensional gaps by comparing input dimensions against output dimensions, and automatically bridges those gaps with spreader chains. The unit tests verify this logic independently of the analyzer pipeline, using directly constructed test data.

**Acceptance Criteria**:
1. `getMetricOutputDimensions` correctly extracts dimension concepts from the metric's IO spec
2. `resolveMetricDimensionalitySpreaders` detects dimension mismatches between input and output and inserts spreader node chains into the input CIL trees
3. Inputs with matching dimensions have unmodified CIL trees
4. Connector and dimension inputs (`InputKind != "metric"`) are skipped
5. Missing spreaders produce errors with descriptive dimension information
6. All unit tests pass with `go test ./src/main/go/analysis/... -count=1 -timeout 60s`

---

### Task 3b: Integrate spreader resolution into the CPL analyzer second pass — Not Started

Wire `resolveMetricDimensionalitySpreaders` into the `Analyze` function so that dimension mismatches on metric inputs are automatically detected and resolved during analysis.

3b. a. Modify `src/main/go/analysis/cpl_analyzer.go` — in the `Analyze` function, after the symbol table walk completes, concept ranges are resolved, derivation paths are computed, unit and connector validations complete, AND after aggregator resolution (Story260) has run, add a second-pass loop that iterates all metrics from `analyzer.metricAnalyzer.GetMetrics()` and calls `analyzer.resolveMetricDimensionalitySpreaders(&metric)` for each. Ensure spreader resolution runs AFTER aggregator resolution: aggregators handle dimension reduction (input has more dims than output), spreaders handle dimension expansion (input has fewer dims than output). Accumulate any errors and return them collectively or on first failure. — Not Started

   b. Write integration tests in `src/main/go/analysis/spreader_dag_integration_test.go`:
      - Full pipeline: define a CPL input with spreader declarations and a metric where an input has fewer dimensions than the output, run the analyzer end-to-end, and verify the metric's input CIL tree contains a `CILSpreadNode` wrapping the original input reference
      - Multi-step pipeline: define spreaders forming a multi-step chain and a metric requiring all steps, run the analyzer, verify the full chain is present in the CIL tree
      - No mismatch: metric where all inputs already have the output dimensions — verify CIL trees are unchanged
      - Error case: metric with a dimension mismatch but no applicable spreader — verify analyzer produces an error with descriptive dimension information
      - Error case: metric with ambiguous spreader candidates — verify analyzer produces an error
      - Run all pre-existing tests: ensure no existing test behavior is broken by the new second-pass loop — Not Started

   c. Run `go test ./src/main/go/analysis/... -count=1 -timeout 60s` and verify all tests pass, including both the new integration tests and all pre-existing tests. — Not Started

**Intent**: Wire `resolveMetricDimensionalitySpreaders` into the CPL analyzer's second pass, so that after all metric input references are resolved and the DAG is assembled, the analyzer automatically detects dimension mismatches and inserts spreader node chains. This integration is the step that makes spreaders actionable — without it, the resolution logic exists but is never called during compilation.

**Acceptance Criteria**:
1. The CPL analyzer's second pass iterates all metrics and calls `resolveMetricDimensionalitySpreaders` after aggregator resolution
2. The full pipeline produces CIL trees with inserted spreader node chains when dimension mismatches have matching spreaders
3. Missing spreaders produce compile-time errors with descriptive dimension information
4. Metrics with no dimension mismatches pass through unchanged
5. Connector inputs are not checked for spreader resolution
6. All pre-existing tests continue to pass (no regressions)
7. All tests pass with `go test ./src/main/go/analysis/... -count=1 -timeout 60s`

---

### Task 4: Create analyze-phase examples — Not Started [Small]

Create parse-and-analyze example files that exercise the spreader resolution pipeline end-to-end, covering both successful resolution and error cases.

4. a. Create `examples/parse_and_analyze_only/without_errors/` examples: CPL files with spreader declarations and metric definitions where single unambiguous spreader sequences exist. At minimum, include:
      - A single-spreader case: spreader `[Product] → [Country, Product]`, metric input has `[Product]`, output requires `[Country, Product]` — analysis succeeds with one `CILSpreadNode` inserted
      - A multi-step case: spreaders `[] → [Country]`, `[Country] → [Country, Region]`, and a metric requiring `[Country, Region]` output from a scalar input — analysis succeeds with two `CILSpreadNode` instances in a chain — Not Started

   b. Create `examples/parse_and_analyze_only/with_errors/` examples:
      - Unresolvable mismatch: spreader declarations exist but none bridges the required dimension gap (e.g., input has `[Product]`, output requires `[Country, Region, Product]`, no spreader adds `Region` from current dims) — analysis produces an error
      - Ambiguous spreaders: two different spreaders both add the same dimension from the same starting dims — analysis produces an ambiguous-candidates error — Not Started

   c. Verify the parse→analyze pipeline produces correct results: `without_errors` examples complete analysis successfully, `with_errors` examples produce the expected error messages (containing dimension names and "ambiguous" where applicable). Use the existing example verification framework. — Not Started

**Intent**: Provide end-to-end verification of the spreader resolution pipeline through the existing parse-and-analyze example framework. These examples serve as both regression tests and documentation of expected behavior for single unambiguous spreader sequences and error conditions.

**Acceptance Criteria**:
1. At least one `without_errors` example with single-spreader resolution completes analysis successfully
2. At least one `without_errors` example with multi-step spreader resolution completes analysis successfully
3. At least one `with_errors` example with an unresolvable dimension mismatch produces an error
4. At least one `with_errors` example with ambiguous spreader candidates produces an error
5. All `parse_and_analyze_only` examples pass the existing example verification framework

---

## Constraints

- Only single unambiguous sequences are supported in Increment 2; multiple candidates for the same dimension gap must produce an error (resolution is deferred to Increment 4)
- No interleaved aggregator/spreader sequences — spreader resolution runs after aggregator resolution and does not interact with aggregators during this increment
- No metric references in spreader expressions — spreader inputs are resolved in Increment 3
- No sequence optimization or shortest-path search — deferred to Increment 4
- Spreader resolution must be order-independent: if adding missing dimensions in a different order produces a different dimensional set, the analyzer must report an error rather than silently choosing one ordering
- All Go source and test file paths use the `src/main/go/` prefix; test commands use `./src/main/go/...`
- All commands run from `source_root` (`src/chronocone_planning_language`)

## Intent

This story integrates spreaders into the metric dependency DAG analysis. After the metric DAG is assembled during the second analysis pass, each metric input is checked for dimensional mismatches. When the output has more dimensions than an input, the analyzer queries the Spreaders map for single unambiguous sequences to bridge the gap. The algorithm adds missing dimensions one at a time, each step finding exactly one spreader whose from-concepts are a subset of current dimensions and whose to-concepts add exactly one missing dimension. Multiple candidates trigger an error (deferred to Increment 4). The `CILSpreadNode` instances are linked in a chain and inserted between the input reference and the calculation body in the metric's CIL tree. No parser changes are needed — spreader declarations are already parsed in Increment 1.

## Acceptance Criteria

1. `findSingleSpreaderSequence` finds correct spreader keys for single and multi-step dimensional additions
2. Multiple candidates for the same dimension addition produce an error
3. No spreader existing for a required dimension produces an error
4. `CILSpreadNode` chain correctly wraps intermediate results: each spreader's `SourceCIL` is the previous `CILSpreadNode`
5. Spreader resolution integrated into the CPL analyzer's second pass — runs after aggregator resolution and before returning `AnalysisDefinitions`
6. Metrics with no dimension mismatches pass through unchanged
7. Metrics with connector inputs are not checked for spreader resolution (only metric inputs)
8. Order independence: adding missing dims in any order produces the same final dimensional set (error if not)
9. All `parse_and_analyze_only` examples pass verification
10. All tests pass with `go test ./src/main/go/analysis/... -count=1 -timeout 60s`
