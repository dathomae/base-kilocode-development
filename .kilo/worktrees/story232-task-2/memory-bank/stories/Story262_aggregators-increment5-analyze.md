# Story262: Aggregators Increment 5 — Analyze

## Goal

Implement full sequence enumeration, selection algorithm (least-info-loss → minimal-work → arbitrary), ambiguous aggregator resolution, specificity comparison. This story (Increment 5 of Epic007) removes Increment 4's limitation of single-dimension-only aggregator sequences. It implements full recursive sequence enumeration that finds all possible multi-step aggregator chains, then selects the optimal one using a three-tier algorithm. The least-information-loss principle ensures the framework preserves as much data granularity as possible by collapsing dimensions with fewer distinct values first. The ambiguous aggregator resolution handles the case where multiple aggregator declarations share the same dimensional pattern — the more specific one (more concepts in `from`) wins. This completes the analysis-phase aggregator logic, giving the framework the intelligence to automatically determine which aggregators to apply and in what order, without user specification.

## References

- [ERO007: Aggregators Incremental Implementation](../specifications/design/cpl/eros/ERO007_aggregators-incremental-implementation.md)
- [ERO007: Aggregators Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO007_aggregators-analysis-phase.md)
- [Plan: Story262 Aggregators Increment 5](../plans/Story262_aggregators-increment5-analyze.md)

## Dependencies

- [Story260_aggregators-increment4-analyze.md](Story260_aggregators-increment4-analyze.md) — provides `findSingleAggregatorSequence`, `resolveMetricDimensionality`, `buildAggregateNode`, and the populated `Aggregators` map with the `aggregators` field on `CPLAnalyzer`
- [Story256_aggregators-increment2-analyze.md](Story256_aggregators-increment2-analyze.md) — provides `AggregatorInfo`, `AggregatorMapKey`, `AggregationMatrix`, and the `Aggregators` map structure
- Epic005 (metrics) — provides concept value sets (`ValueOrdering`) for information loss computation

## Dependent Stories

- Story261 (aggregators-increment4-synthesize) — Not Started — depends on the multi-step CIL aggregate node chains inserted by this story's analysis phase
- Story263 (aggregators-increment5-synthesize) — Not Started — depends on this story's resolved aggregator sequences and ambiguity resolution

## Tasks

### Task Sizing

_To be filled after architect-for-story-planning consultation._

### Parallel Execution

_To be filled after architect-for-story-planning consultation._

### Execution Order

```
Task 1 → [Task 2 ‖ Task 3] → Task 4
```

- Task 1 must complete first (enumerable sequences are the foundation that Tasks 2 and 4 consume)
- Tasks 2 and 3 are planned for parallel execution: Task 2 implements selection (consumes enumerated sequence lists); Task 3 implements ambiguity resolution (operates on the Aggregators map). Both add independent functions to `analyze_aggregators.go`.
- Task 4 depends on all preceding tasks (integrates full enumeration + selection into `resolveMetricDimensionality`)

### Reintegration

_To be filled after architect-for-story-planning consultation._

---

### Task 1: Implement sequence enumeration — Not Started

1. a. Implement `enumerateAggregatorSequences(fromDims []string, targetDims []string) ([][]string, error)` in `src/main/go/analysis/analyze_aggregators.go`. This function replaces `findSingleAggregatorSequence` from Increment 4. It finds ALL valid sequences of aggregators that reduce `fromDims` to `targetDims`. Compute the set of dimensions to remove (`toRemove = setDifference(fromDims, targetDims)`). If `toRemove` is empty (no reduction needed), return nil. Otherwise, call a recursive helper `enumerateRecursive` and return all valid sequences found, or an error if none are found. — Not Started
   b. Implement the recursive helper `enumerateRecursive` in `src/main/go/analysis/analyze_aggregators.go`. For each concept in `remainingToRemove`, check if an aggregator exists with key `AggregatorMapKey(currentDims, concept)`. If found, compute new dimensions (remove concept from `currentDims`), recurse with the reduced remaining set. Base case: when `len(newRemaining) == 0 && isSubset(targetDims, newDims)`, store the accumulated sequence as valid. — Not Started
   c. Implement `setDifference(a, b []string) []string` and `isSubset(subset, superset []string) bool` helpers in `src/main/go/analysis/analyze_aggregators.go`. `setDifference` returns elements in `a` that are not in `b`. `isSubset` returns true if every element of `subset` is present in `superset`. Both functions are unexported helper functions used only by `enumerateAggregatorSequences`. — Not Started
   d. Write tests in `src/main/go/analysis/analyze_aggregators_test.go`:
      - Single-step sequence: dims [Country, Region] → [Region], with an aggregator to collapse Country → sequence `[[Country_to_Region, sum]]` found
      - Two-step sequence: dims [Country, Region, Product] → [Product], with aggregators for collapsing Country and Region → both orderings enumerated (Collapse Country then Region, and Collapse Region then Country)
      - No sequence found: dims [A, B] → [], no aggregators for either A or B → error with diagnostic dimension names
      - No reduction needed: dims [Country] → [Country] (or no excess dimensions) → nil sequence returned — Not Started
   e. Run `go test ./src/main/go/analysis/... -run TestEnumerateAggregatorSequences -count=1` and verify all tests pass. — Not Started

**Intent**: Replace Increment 4's single-sequence lookup (`findSingleAggregatorSequence`) with full recursive enumeration that finds every valid multi-step aggregator chain. This is the foundation for the selection algorithm (Task 2) and enables multi-step dimension reductions in `resolveMetricDimensionality` (Task 4).

**Acceptance Criteria**:
- `enumerateAggregatorSequences` finds ALL valid aggregator sequences that bridge `fromDims` → `targetDims`
- Single-step and multi-step sequences (both orderings) are correctly enumerated
- When no aggregator sequence exists, the error message includes the from and target dimension names
- Zero excess dimensions produce a nil sequence with no error
- All tests pass with `go test ./src/main/go/analysis/... -run TestEnumerateAggregatorSequences`

---

### Task 2: Implement selection algorithm and information loss — Not Started

2. a. Implement `sequenceInformationLoss(sequence []string) int` in `src/main/go/analysis/analyze_aggregators.go`. For each aggregator key in the sequence, look up the `AggregatorInfo` in `a.aggregators`, find the collapsed concept through `a.conceptAnalyzer`, and sum the length of `collapsedConcept.ValueOrdering` (the number of distinct values in the collapsed dimension). Lower total = less information lost. — Not Started
   b. Implement `selectOptimalSequence(sequences [][]string) []string` in `src/main/go/analysis/analyze_aggregators.go`. Sort all sequences using a three-tier comparator:
      - **Tier 1 (Least Information Loss)**: Compare `sequenceInformationLoss` — lower is better
      - **Tier 2 (Minimal Work)**: If Tier 1 is equal, prefer the shorter sequence (fewer aggregator operations)
      - **Tier 3 (Deterministic Tiebreaker)**: If Tiers 1 and 2 are equal, sort lexicographically by the first aggregator key in each sequence to ensure deterministic results across compilations
      Return the first (best) sequence. If there is only one sequence, return it directly. — Not Started
   c. Write tests in `src/main/go/analysis/analyze_aggregators_test.go`:
      - **Tier 1 (info loss)**: Country has 2 values, Region has 10 values → prefer collapsing Country first (less info loss)
      - **Tier 2 (sequence length)**: Two sequences with equal info loss but one is shorter (e.g., 2-step vs 3-step) → shorter wins
      - **Tier 3 (lexicographic tiebreaker)**: Two sequences identical in Tiers 1 and 2 → deterministic lexicographic order by first key — Not Started
   d. Run `go test ./src/main/go/analysis/... -run TestSelectOptimalSequence -count=1` and verify all tests pass. — Not Started

**Intent**: Implement a principled three-tier selection algorithm that chooses the best aggregator sequence from all enumerated possibilities. The least-information-loss principle ensures the framework preserves as much data granularity as possible. When multiple sequences have equal information loss, the minimal-work tiebreaker reduces runtime overhead. When still tied, the lexicographic tiebreaker guarantees deterministic results.

**Acceptance Criteria**:
- `selectOptimalSequence` picks the sequence with least information loss (Tier 1)
- When info loss is equal, shorter sequences are preferred (Tier 2)
- When Tiers 1 and 2 are equal, the lexicographic tiebreaker produces deterministic results (Tier 3)
- `sequenceInformationLoss` correctly sums the collapsed concept value counts across all steps
- All tests pass with `go test ./src/main/go/analysis/... -run TestSelectOptimalSequence`

---

### Task 3: Implement ambiguous aggregator resolution — Not Started

3. a. Implement `resolveAmbiguousAggregators() error` in `src/main/go/analysis/analyze_aggregators.go`. Group all entries in `a.aggregators` by their `AggregatorMapKey`. For each group with more than one entry, resolve by comparing specificity (the number of concepts in `info.FromConcepts`). More concepts = more specific = narrower match criteria. The most specific aggregator replaces the ambiguous group in `a.aggregators`. If two or more aggregators share the same key AND the same specificity, return an error: `"ambiguous aggregators for pattern <key>: multiple with specificity <N>"`. — Not Started
   b. Add a call to `resolveAmbiguousAggregators()` in `src/main/go/analysis/analyze_aggregators.go` at the point after all individual aggregators are analyzed and before dimension mismatch resolution occurs. This ensures the Aggregators map is unambiguous before `resolveMetricDimensionality` queries it. — Not Started
   c. Write tests in `src/main/go/analysis/analyze_aggregators_test.go`:
      - Single aggregator (no ambiguity): one entry for a key → no change, no error
      - Two aggregators (specific wins): same `(from, to)` key, one has 2 concepts in `from`, the other has 3 → more specific (3 concepts) wins
      - Equal specificity → error: two aggregators with same key and same `len(FromConcepts)` → descriptive error message including the key and specificity value — Not Started
   d. Run `go test ./src/main/go/analysis/... -run TestResolveAmbiguousAggregators -count=1` and verify all tests pass. — Not Started

**Intent**: Handle the case where multiple aggregator declarations share the same dimensional pattern `(from, to)`. The resolution rule (prefer more specific match criteria) ensures that when two aggregators could apply, the more constrained one wins. When specificity is equal, the ambiguity is genuinely unresolvable and must be reported as a compile-time error so the user can disambiguate.

**Acceptance Criteria**:
- `resolveAmbiguousAggregators` selects the more specific aggregator (more concepts in `from`) when multiple share the same key
- Equally specific ambiguous aggregators produce a clear, descriptive error
- Single aggregator keys are unchanged
- The function is called at the correct point in the analysis pipeline (after all aggregators are analyzed, before dimension resolution)
- All tests pass with `go test ./src/main/go/analysis/... -run TestResolveAmbiguousAggregators`

---

### Task 4: Integrate full enumeration into resolveMetricDimensionality — Not Started

4. a. In `src/main/go/analysis/analyze_aggregators.go`, modify `resolveMetricDimensionality` to replace the call to `findSingleAggregatorSequence(inputDims, outputDims, excessDims)` with:
      1. `sequences, err := a.enumerateAggregatorSequences(inputDims, outputDims)`
      2. `selectedSequence := a.selectOptimalSequence(sequences)`
      Use the selected sequence to build and insert aggregate nodes (via the existing `buildAggregateNode` calls per aggregator key). — Not Started
   b. Remove the Increment 4 deferred error for multiple excess dimensions (the `return fmt.Errorf("multiple excess dimensions require sequence enumeration (Increment 5)")` in `findSingleAggregatorSequence`). After this task, `resolveMetricDimensionality` handles any number of excess dimensions through full sequence enumeration, and `findSingleAggregatorSequence` itself will be removed (replaced entirely by the new functions from Task 1). — Not Started
   c. Write integration tests in `src/main/go/analysis/analyze_aggregators_test.go`:
      - Multi-step dimension reduction: metric input dims [Country, Region, Product], output dims [Product], aggregators for Country→[] and Region→[] exist. Verify `resolveMetricDimensionality` inserts TWO aggregate nodes (one for each collapsed dimension) with the optimal sequence (least info loss) selected
      - No-sequence-found error: metric input dims [A, B], output [], no aggregators for either dimension → error with descriptive message listing the from and target dimensions — Not Started
   d. Run `go test ./src/main/go/analysis/... -count=1 -timeout 60s` and verify ALL analysis tests pass, including pre-existing tests from Increment 4 and all new tests from Tasks 1–3. — Not Started

**Intent**: Wire the full enumeration and selection machinery into the existing dimension-resolution pipeline. After this task, `resolveMetricDimensionality` can handle arbitrarily complex multi-step dimension reductions without any limitation on the number of excess dimensions. The Increment 4 deferred error is removed since this story provides the necessary multi-step capability.

**Acceptance Criteria**:
- `resolveMetricDimensionality` uses `enumerateAggregatorSequences` + `selectOptimalSequence` instead of `findSingleAggregatorSequence`
- Multi-step dimension reductions produce aggregate node chains with the optimal sequence selected
- No-sequence-found errors include descriptive messages with from/to dimension names
- All analysis tests pass with `go test ./src/main/go/analysis/... -count=1 -timeout 60s`

---

## Test-First Development

This story follows the Logical TDD Lifecycle. For each coding task, tests are written before (or alongside) the implementation code. The test-and-implement cycle is integrated within each task, not broken out as separate tasks. See `memory-bank/concepts.md` for details on the Logical TDD Lifecycle.

## Constraints

- **Existing aggregator infrastructure**: This story depends on the `aggregators` map on `CPLAnalyzer` (populated with `AggregatorInfo` entries), the existing `buildAggregateNode` factory function, and the `resolveMetricDimensionality` function — all provided by Story260. It does not create new aggregator types, CIL node definitions, or aggregation infrastructure.
- **File scope**: All changes are confined to `src/main/go/analysis/analyze_aggregators.go` and its test file. No changes to `cpl_analyzer.go`, `cil/`, or other packages.
- **Lexicographic tiebreaker determinism**: The Tier 3 lexicographic sort must use Go's standard string comparison on aggregator keys to guarantee deterministic results across all platforms and Go versions.
- **Information loss via concept value counts**: `sequenceInformationLoss` depends on `conceptAnalyzer` having resolved all concept `ValueOrdering` slices. It must not be called before concept ranges are resolved (i.e., after `AnalyzeConceptsScope` and `ResolveConceptRanges` complete).
- **Error message clarity**: All dimension-mismatch errors must include the names of the input dimensions, output dimensions, and the excess dimensions, so users can diagnose misconfigurations without reading source code.

## Intent

This story removes Increment 4's limitation of single-dimension-only aggregator sequences. It implements full recursive sequence enumeration that finds all possible multi-step aggregator chains, then selects the optimal one using a three-tier algorithm. The least-information-loss principle ensures the framework preserves as much data granularity as possible by collapsing dimensions with fewer distinct values first. The ambiguous aggregator resolution handles the case where multiple aggregator declarations share the same dimensional pattern — the more specific one (more concepts in `from`) wins. After this story, the CPL analyzer can automatically determine which aggregators to apply and in what order for any dimension mismatch, without relying on manual specification by the metric author.

## Acceptance Criteria

1. `enumerateAggregatorSequences` finds ALL valid aggregator sequences that bridge `fromDims` → `targetDims`
2. Single-step and multi-step sequences (both orderings) are correctly enumerated
3. `selectOptimalSequence` picks the sequence with least information loss (Tier 1)
4. When info loss is equal, shorter sequences are preferred (Tier 2)
5. When Tiers 1+2 are equal, the lexicographic tiebreaker produces deterministic results (Tier 3)
6. `resolveAmbiguousAggregators` selects the more specific aggregator (more concepts in `from`)
7. Equally specific ambiguous aggregators produce a clear error
8. `resolveMetricDimensionality` now handles multi-step dimension reductions automatically
9. No-sequence-found errors include descriptive messages with from/to dimension names
10. All tests pass with `go test ./src/main/go/analysis/... -count=1 -timeout 60s`
