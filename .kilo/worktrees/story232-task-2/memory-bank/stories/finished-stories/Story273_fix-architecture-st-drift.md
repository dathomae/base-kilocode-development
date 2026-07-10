# Story273: Fix Architecture Drift — Symbol Table Definition Registration

## Goal

Fix symbol table drift by registering scope-carrying definitions (metrics, aggregators) via `DefineWithData`, iterating `Definitions` in analysis, and propagating the consistent pattern to aggregators for future spreader consistency. Remediate the drift that started with Metrics and propagated to Aggregators, establishing a single pattern for all scope-carrying definitions.

## References

- [Plan: Fix Architecture Drift](../plans/fix-architecture-st-drift-plan.md)
- [Symbol Table-Driven Parsing Pattern](../../src/chronocone_planning_language/specifications/patterns/symbol-table-driven-parsing.md)

## Dependencies

- [Story255_aggregators-increment2-parse.md](finished-stories/Story255_aggregators-increment2-parse.md) — Done (the aggregator code that needs fixing)
- [Story182_cpl-metrics-basic-parse-analyze.md](finished-stories/Story182_cpl-metrics-basic-parse-analyze.md) — Done (the metrics code that needs retrofitting)
- [Story222_cpl-refactoring-phase2-normalize-cil.md](finished-stories/Story222_cpl-refactoring-phase2-normalize-cil.md) — Done (changed `DefineWithData`'s `data` parameter from `interface{}` to `DefinitionData`, which is why metrics/aggregators stopped being passable)

## Dependent Stories

- [Story256_aggregators-increment2-analyze.md](Story256_aggregators-increment2-analyze.md) — Not Started (benefits from the consistent `Definitions`-based retrieval pattern established here)

## Tasks

### Task Sizing

- **Task 1** `[Small]` — Adds two empty `IsDefinitionData()` methods in 2 files (different packages), each a single-line additive pattern following existing precedent. No new abstractions.
- **Task 3** `[Small]` — Adds one `DefineWithData` call and corresponding test assertions in 2 files (single package), following an established pattern. Purely additive.

Task 2 is sized for `code-for-story-implementor` — it spans 5 files across 2 packages with coordinated production logic changes.

### Parallel Execution

- **Group A** (Tasks 2, 3): Fully disjoint file sets (Task 2 touches metrics listener/analysis/test files; Task 3 touches aggregators listener/test files). Both depend only on Task 1's `IsDefinitionData()` methods, not on each other.

### Execution Order

```
Task 1 → [Task 2 ‖ Task 3] → Task 4
```

### Reintegration

**Group A (Tasks 2, 3):**
- Merge order: any order (fully disjoint files, no shared infrastructure)
- Integration tests: `go test ./src/main/go/parser/... -count=1 -timeout 120s` — both tasks modify parser package files; must verify combined DefineWithData calls compile and pass together. Then `go test ./src/main/go/analysis/... -count=1 -timeout 120s` — Task 2 modifies analysis code.
- Watch for conflicts in: `src/main/go/go.sum` (dependency resolution may differ per worktree); `src/main/go/parser/symbol_table_integration_test.go` (not modified by either task, but existing integration tests may surface merged behavior)

### Task 1: Add `DefinitionData` interface satisfaction to parser structs — Completed [Small]

**Intent**: `ParserMetricDefinition` and `ParserAggregatorDefinition` must implement the `symboltable.DefinitionData` interface so they can be passed to `DefineWithData`. This is a structural prerequisite for Tasks 2 and 3.

**Completion criteria**:
- `ParserMetricDefinition` has `func (ParserMetricDefinition) IsDefinitionData() {}`
- `ParserAggregatorDefinition` has `func (ParserAggregatorDefinition) IsDefinitionData() {}`
- `go build ./src/main/go/parsing/... ./src/main/go/parser/...` succeeds

    a. Add `func (ParserMetricDefinition) IsDefinitionData() {}` to `src/main/go/parsing/parser_metric_definition.go` — Completed
       - The `symboltable` package is already imported
       - Place method after the struct definition, before any other methods

    b. Add `func (ParserAggregatorDefinition) IsDefinitionData() {}` to `src/main/go/parser/parser_aggregator_definition.go` — Completed
       - The `symboltable` package is already imported
       - Place method after the struct definition

    c. Verify: `go build ./src/main/go/parsing/... ./src/main/go/parser/...` — Completed

### Task 2: Add `DefineWithData` call for `MetricScopeDefinition` and update analysis — Completed

**Intent**: `ExitMetricDeclaration` and `ExitShortFormDeclaration` must call `DefineWithData` to push a `MetricScopeDefinition` entry into the scope's `Definitions` slice, enabling the analysis phase to retrieve the definition by walking `Definitions` rather than casting `scope.Data`. The `scope.Data` assignment is preserved for backward compatibility and convenience.

**Completion criteria**:
- `ExitMetricDeclaration` (long form) pushes `MetricScopeDefinition` via `DefineWithData` before popping
- `ExitShortFormDeclaration` pushes `MetricScopeDefinition` via `DefineWithData` before popping
- `AnalyzeMetricScope` iterates `Definitions` instead of casting `table.Data`
- Parser tests and analysis tests pass

    a. Modify `ExitMetricDeclaration` in `src/main/go/parser/cpl_metrics_listener.go` — Completed
       - After `metricScope.Data = metricDef` (line 143), add:
         ```go
         nameCopy := metricName
         _ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, location, metricDef)
         ```
       - This mirrors the Lookups pattern (`ExitLookupDeclaration` line 137)
       - Use `_` to discard the error (duplicate check is deferred to analysis, matching the existing `Data` assignment pattern)

    b. Modify `ExitShortFormDeclaration` in `src/main/go/parser/cpl_metrics_listener.go` — Completed
       - After `metricScope.Data = metricDef` (line 450), add the same `DefineWithData` call as in subtask (a)

     c. Update `AnalyzeMetricScope` in `src/main/go/analysis/metrics_analysis.go` — Completed
        - Replace the `table.Data.(*parsing.ParserMetricDefinition)` cast (line 135) with iteration over `table.Definitions`:
          ```go
          var metricDef *parsing.ParserMetricDefinition
          for _, def := range table.Definitions {
              if def.DefinitionType == symboltable.MetricScopeDefinition {
                  if md, ok := def.Data.(*parsing.ParserMetricDefinition); ok {
                      metricDef = md
                      break
                  }
              }
          }
          if metricDef == nil {
              return fmt.Errorf("metric scope has no MetricScopeDefinition in its Definitions")
          }
          ```
        - No `table.Data` fallback is kept: if the `Definitions` walk yields no `MetricScopeDefinition`, analysis returns an error immediately. This enforces the canonical `Definitions`-iteration path and eliminates implicit reliance on the legacy `Data` field.

    d. Update `src/main/go/parser/cpl_metrics_parser_test.go` — Completed
       - After verifying `metricScope.Data`, also verify that `metricScope.Definitions` contains at least one entry with `DefinitionType == symboltable.MetricScopeDefinition`
       - Verify the `SymbolDefinition.Name` matches the metric name
       - Verify the `SymbolDefinition.Data` is the same `*parsing.ParserMetricDefinition` pointer

    e. Update `src/main/go/analysis/metrics_analysis_test.go` — Completed
       - Replace direct assignments to `metricScope.Data` (approximately 37 occurrences in manual test tables) with `metricScope.DefineWithData` calls that push `MetricScopeDefinition` entries, so the primary `Definitions`-iteration path is exercised in unit tests
       - The `requireParsedMetricDefinition` helper in `src/main/go/analysis/metrics_time_shift_integration_test.go` reads from `table.Data` and does not need modification
       - Add a dedicated analysis unit test where `metricScope.Data = nil` and `Definitions` contains only a `MetricScopeDefinition` entry, verifying `AnalyzeMetricScope` succeeds without falling back to the `Data` field

    f. Verify: `go test ./src/main/go/analysis/... ./src/main/go/parser/... -count=1 -timeout 120s` — Completed

### Task 3: Add `DefineWithData` call for `AggregatorScopeDefinition` — Completed [Small]

**Intent**: `ExitAggregatorDeclaration` must call `DefineWithData` to push an `AggregatorScopeDefinition` entry into the scope's `Definitions` slice, establishing the same pattern used by Lookups and Metrics. The `scope.Data` assignment is preserved.

**Completion criteria**:
- `ExitAggregatorDeclaration` pushes `AggregatorScopeDefinition` via `DefineWithData`
- Aggregator listener tests verify the `Definitions` entry
- `scope.Data` assignment is preserved alongside the new `DefineWithData` call

    a. Modify `ExitAggregatorDeclaration` in `src/main/go/parser/cpl_aggregators_listener.go` — Completed
       - After `aggregatorScope.Data = def` (line 175), add:
         ```go
         keyCopy := def.Key
         _ = aggregatorScope.DefineWithData(&keyCopy, symboltable.AggregatorScopeDefinition, location, def)
         ```
       - The synthetic key string is used as the definition name
       - Use `_` to discard the error (duplicate detection deferred to analysis)

    b. Update `src/main/go/parser/cpl_aggregators_listener_test.go` — Completed
       - After verifying `aggregatorScope.Data`, also verify that `aggregatorScope.Definitions` contains at least one entry with `DefinitionType == symboltable.AggregatorScopeDefinition`
       - Verify the `SymbolDefinition.Name` matches the synthetic key
       - Verify the `SymbolDefinition.Data` is the same `*ParserAggregatorDefinition` pointer
       - Add this assertion to every test that creates a valid aggregator declaration

    c. Verify: `go test ./src/main/go/parser/... -count=1 -timeout 120s` — Completed

### Task 4: Full regression verification — Completed

**Intent**: Confirm that no downstream consumers are broken by the changes and all existing tests continue to pass.

**Completion criteria**:
- All tests pass: `go test ./src/main/go/... -count=1 -timeout 120s`
- No new `go vet` warnings in non-generated code

    a. Run full test suite — Completed
       - `go test ./src/main/go/... -count=1 -timeout 120s`

    b. Run `go vet` — Completed
       - `go vet ./src/main/go/parser/... ./src/main/go/analysis/...`
       - Ignore pre-existing ANTLR-generated warnings

    c. Build check — Completed
       - `go build ./src/main/go/...`

### Test-First Development

This story follows the Logical TDD Lifecycle. Each behavioral task (Tasks 2, 3) includes test assertion additions alongside the implementation code within the same task. The analysis retrieval change in Task 2c is verified by the existing test suite against the new `Definitions` entries produced by Tasks 2a–2b. Task 1 is purely structural (adding empty interface satisfaction methods) and requires only a build check.

### Constraints

- `ParserMetricDefinition` is in package `parsing` (not `parser`) — its path is `src/main/go/parsing/parser_metric_definition.go`
- `ParserAggregatorDefinition` is in package `parser` with path `src/main/go/parser/parser_aggregator_definition.go`
- The `DefineWithData` error must be discarded with `_` (duplicate detection is deferred to analysis, matching the existing `scope.Data` assignment pattern)
- **Convention deviation — error handling**: This story discards `DefineWithData` errors in metric/aggregator listeners with `_`, which differs from the existing Lookups and Inputs patterns (where errors are checked and reported via `l.AddError`). The deviation is intentional: `DefineWithData` can only fail when a name collides within the same scope, which cannot happen inside a single metric or aggregator declaration. The pattern will be aligned when duplicate detection is moved to analysis in a future story.
- The `scope.Data` assignment is preserved alongside the new `DefineWithData` call for backward compatibility and convenience, but `AnalyzeMetricScope` no longer reads from it — the canonical retrieval path is through `Definitions` iteration

### Intent

**Task 1 Intent**: Make `ParserMetricDefinition` and `ParserAggregatorDefinition` eligible to be passed as the `data` parameter to `DefineWithData`, which changed from `interface{}` to `DefinitionData` in Story222. Without this structural change, Tasks 2 and 3 cannot compile.

**Task 2 Intent**: Establish the `DefineWithData` → `Definitions` iteration pipeline for metrics. The parser pushes a `MetricScopeDefinition` entry into the metric scope's `Definitions` slice, and the analysis phase retrieves it by walking `Definitions` rather than casting `scope.Data`. This is the same pattern already used by Lookups and is the target pattern for all scope-carrying definitions.

**Task 3 Intent**: Propagate the `DefineWithData` pattern to aggregators, ensuring they follow the same convention as Lookups and Metrics. This is a forward-looking change — Story256 (`aggregators-increment2-analyze`) will retrieve aggregator definitions from `Definitions` rather than from `scope.Data`.

**Task 4 Intent**: Comprehensive regression verification across the entire test suite to confirm that the structural and behavioral changes do not break any existing functionality. This is the final safety check before the story is complete.

### Acceptance Criteria

- **Task 1**: `ParserMetricDefinition` and `ParserAggregatorDefinition` each have an `IsDefinitionData()` method; `go build ./src/main/go/parsing/... ./src/main/go/parser/...` succeeds
- **Task 2**: `ExitMetricDeclaration` and `ExitShortFormDeclaration` each call `DefineWithData` with `MetricScopeDefinition`; `AnalyzeMetricScope` retrieves the definition from `Definitions`; all parser and analysis tests pass
- **Task 3**: `ExitAggregatorDeclaration` calls `DefineWithData` with `AggregatorScopeDefinition`; aggregator listener tests verify the `Definitions` entry; all parser tests pass
- **Task 4**: `go test ./src/main/go/... -count=1 -timeout 120s` passes with no new failures; `go vet` produces no new warnings in non-generated code; `go build ./src/main/go/...` succeeds

## Notes

- The plan references `src/main/go/parsing/` for `parser_metric_definition.go` and `src/main/go/parser/` for all other files — this is correct and intentional; the two packages are distinct
- Story256 (`aggregators-increment2-analyze`) will benefit from this story's `AggregatorScopeDefinition` entry, as it can retrieve aggregator definitions from `Definitions` rather than from `scope.Data`
- Future spreader stories will also follow this pattern, making this story a foundational consistency fix
