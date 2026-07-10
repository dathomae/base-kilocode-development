# Story265: Spreaders Increment 1 — Analysis

## Goal

Implement the analysis phase for spreader declarations (Increment 1 of Epic008). This story delivers `SpreaderFunctionRegistry` with all six spreader functions registered from the start, `AnalyzeSpreaderScope` with postfix expression evaluation supporting concept-only identifier resolution, dimensional unit validation (`input_units * expression_units = output_units`), `CILSpreadNode` generation with pre-built `SpreadingMatrix`, and the `Spreaders` lookup map stored in `AnalysisDefinitions`. Cross-reference validation ensures `from` concepts are a proper subset of `to` concepts with exactly one dimension added, and duplicate `(from, to)` patterns are rejected.

## References

- [ERO008: Spreaders Incremental Implementation](../specifications/design/cpl/eros/ERO008_spreaders-incremental-implementation.md)
- [ERO008: Spreaders Analysis Phase](../specifications/design/cpl/go-analysis/eros/ERO008_spreaders-analysis-phase.md)
- [Plan: Story265_spreaders-increment1-analyze](../plans/Story265_spreaders-increment1-analyze.md)
- [Story Template](../../resources/templates/stories.md)

## Dependencies

- [Story264_spreaders-increment1-parse.md](Story264_spreaders-increment1-parse.md) — Parser must produce `ParserSpreaderDefinition` under synthetic keys in the symbol table
- [Story256_aggregators-increment2-analyze.md](Story256_aggregators-increment2-analyze.md) — Spreader analysis mirrors the aggregator analysis pattern (postfix evaluation, dimensional unit algebra, FunctionRegistry pattern); `DimensionalUnitExpression` type and its `Multiply`/`Equals` methods must exist
- [Story254_aggregators-increment1-matrix-mult.md](Story254_aggregators-increment1-matrix-mult.md) — `AggregationMatrix` struct reused for `SpreadingMatrix` construction
- Epic001 (concepts) — Concept definitions must be available for `from`/`to` resolution and matrix construction
- Epic005 (metrics) — Metrics map must exist in `AnalysisDefinitions`

## Dependent Stories

- Spreader Increment 2 (synthesis) — Depends on spreaders being available in `AnalysisDefinitions.Spreaders`
- Spreader Increment 3 (metric references) — Depends on spreader analysis infrastructure

## Tasks

### Task Sizing

Tasks 1–3 each introduce new struct types across multiple files with coordinated changes, or implement multiple helpers serving a coherent analysis function. None of these qualify as `[Small]` under the rules of thumb (they introduce new abstractions, touch 2+ source files with coordinated changes, or require reasoning across multiple concerns). Task 4 is a narrow-scope wiring task (1 source file modified + template-following example files) and is annotated `[Small]`.

The original Task 2 (10 subtasks a–j) was decomposed into Tasks 2a and 2b during architect review. The natural seam is the expression evaluator: 2a builds the core analyzer with the expression evaluation machinery, and 2b validates the output and builds `SpreaderInfo`. Task 2b has a compile-time dependency on 2a (it calls the expression evaluator produced by 2a), creating a correct sequential dependency.

### Parallel Execution

None — all tasks share `spreader_analyzer.go` (Tasks 1, 2a, 2b, 3) and each task semantically depends on the output of the preceding task. Task 4 depends on all prior tasks for end-to-end verification.

### Execution Order

```
Task 1 → Task 2a → Task 2b → Task 3 → Task 4
```

- Task 1 must complete first: `FunctionRegistry` is required by the expression evaluator in Task 2a
- Task 2a depends on Task 1: expression evaluation uses the `FunctionRegistry`
- Task 2b depends on Task 2a: dimensional validation requires expression evaluation output
- Task 3 depends on Task 2b: CIL node generation needs `SpreaderInfo`
- Task 4 depends on all preceding tasks: end-to-end verification requires the full analysis pipeline

### Reintegration

None — no parallel groups. All tasks execute sequentially.

### Task 1: Define SpreaderFunctionRegistry with all six functions — Completed

   a. Create `src/main/go/analysis/spreader_analyzer.go` — define `SpreaderFunctionInfo` and `SpreaderFunctionRegistry` struct types with Godoc comments following the conventions in `analysis_definitions.go` — Not Started
   b. Implement `NewSpreaderFunctionRegistry()` registering all six functions (`size`, `1/size`, `sum`, `count`, `1/sum`, `1/count`), each returning the correct `DimUnits` callback as specified in the plan — Not Started
   c. Write unit tests in `src/main/go/analysis/spreader_analyzer_test.go`: verify each function name is registered, verify `DimUnits` output for each function with various params, verify unknown function lookup returns an error — Not Started
   d. Verify all registry tests pass: `go test ./src/main/go/analysis/... -count=1 -run SpreaderFunctionRegistry` — Not Started

### Task 2a: Implement AnalyzeSpreaderScope core — concept resolution, dimensional rules, and expression evaluation — Completed

   a. Implement `AnalyzeSpreaderScope` method on `*CPLAnalyzer` — iterate scope definitions for `SpreaderScopeDefinition` entries, follow the plan's pseudocode structure — Not Started
   b. Implement `resolveConceptList` helper — resolve raw concept name strings to `*AnalysisConceptDefinition` pointers, return error for any concept name not found in `conceptAnalyzer.Concepts` — Not Started
   c. Implement `validateSpreaderDimensionalRules` — verify all `from` concepts appear in `to` (containment) AND `len(to) == len(from) + 1` (exactly one dimension added) — Not Started
   d. Implement `findNewDimension` — identify which concept in `to` is not in `from` and return it — Not Started
   e. Implement `evaluateSpreaderExpression` — postfix stack evaluator handling four operand kinds: `concept` (pushes concept's dimensional unit), `identifier` (Increment 1: resolves to concept only, errors if not found; metric resolution deferred to Increment 3), `fncall` (looks up in `spreaderFunctionRegistry`, applies `DimUnits` callback), `operator` (pops two from stack, applies `Multiply` for `*`, inverts and multiplies for `/`) — Not Started
   f. Write unit tests in `spreader_analyzer_test.go`: verify expression evaluation with `1/size(Country)` produces unitless result, verify `Product * 1/size(Country)` produces `Product` numerator, verify multi-concept expression, verify identifier resolution for concepts, verify unknown identifier error — Not Started
   g. Verify expression evaluation tests pass: `go test ./src/main/go/analysis/... -count=1 -run Expression` — Not Started

### Task 2b: Dimensional unit validation, SpreaderInfo building, and storage — Completed

   a. Implement `validateSpreaderDimUnits` — enforce `input_units * expression_units = output_units` using the `DimensionalUnitExpression.Multiply` and `Equals` methods (inverse of aggregator formula `input_units = expression_units * output_units`) — Not Started
   b. Implement `dimUnitsFromConcepts` helper — construct `DimensionalUnitExpression` from a slice of `*AnalysisConceptDefinition` by placing each concept name in the denominator — Not Started
   c. Implement `SpreaderMapKey` function — generate compound string key from sorted, unique `from` and `to` concept name sets using the format `[fromSorted]->[toSorted]` — Not Started
   d. Build `SpreaderInfo` struct (populate all fields including `FromEmpty` from the parser definition, `NewDimension` from `findNewDimension`, dimensional unit expressions, and `SourceLocation`), store in `CPLAnalyzer.spreaders` map under the compound key — Not Started
   e. Add `Spreaders map[string]*SpreaderInfo` field to `AnalysisDefinitions` struct, populate it in `GetAnalysisDefinitions()` (store as `Spreaders` rather than `spreaders` for consistency with other map fields) — Not Started
   f. Detect duplicate spreader patterns: if `SpreaderMapKey` for a `(from, to)` pair already exists in the spreaders map, return an error with source locations of both declarations — Not Started
   g. Write unit tests in `spreader_analyzer_test.go`: verify valid spreader produces correct `SpreaderInfo` with all fields populated, verify dimensional unit validation catches mismatches (e.g., `from [Product] * expression {Country} = to [Country,Product]` is invalid — missing `Product` in expression), verify from/to validation rejects `from [A] to [A,B,C]` (adds 2 dims) and `from [A] to [B,C]` (from not subset of to) — Not Started
   h. Write error case tests: missing concept in from/to, unknown identifier in expression, unknown function in expression, duplicate `(from, to)` pattern — Not Started
   i. Verify all analysis tests pass: `go test ./src/main/go/analysis/... -count=1 -timeout 60s` — Not Started

### Task 3: Add CILSpreadNode and SpreadingMatrix construction — Completed

   a. Create `src/main/go/cil/cil_spread_node.go` — define `CILSpreadNode` struct with all Increment 1 fields (`SourceCIL`, `FromDimensions`, `ToDimensions`, `NewDimension`, `SpreadingMatrix`, `DistributionExpression`, `ReferencedMetrics` set to nil, `DimensionalUnitEffect`, `SourceLocation`), implement `cilNode()` marker method — Not Started
   b. Implement `BuildSpreadingMatrix` function in `spreader_analyzer.go` — construct `AggregationMatrix` with 1 source row, columns equal to concept value count, weights from expression dimension values — Not Started
   c. Implement CIL node generation from `SpreaderInfo`: produce a `CILSpreadNode` with `FromDimensions` and `ToDimensions` as sorted concept name slices, `NewDimension` set, `SpreadingMatrix` pre-built with equal-proportion weights from the distribution expression — Not Started
   d. Store `CILSpreadNode` in `AnalysisDefinitions.CILTrees` map under the spreader's compound map key (same `SpreaderMapKey` format) — Not Started
   e. Write unit tests in `spreader_analyzer_test.go`: verify `CILSpreadNode` structure has correct `FromDimensions`, `ToDimensions`, and `NewDimension`, verify `SpreadingMatrix` dimensions match concept value counts, verify equal-proportion weights (e.g., `1/size(Country) * Country` produces weight `1/N` for each country value) — Not Started
   f. Verify CIL tests pass: `go test ./src/main/go/analysis/... -count=1 -run CILSpread` and `go test ./src/main/go/cil/... -count=1` — Not Started

### Task 4: End-to-end analysis verification with parse_and_analyze_only examples — Completed [Small]

   a. Create `examples/parse_and_analyze_only/without_errors/spreaders/` directory with CPL example files demonstrating valid spreader declarations — expect successful analysis (no errors) — Not Started
   b. Create `examples/parse_and_analyze_only/with_errors/spreaders/` directory with CPL example files demonstrating intentional analysis errors: dimensional mismatches, missing concepts, unknown functions, duplicate `(from, to)` patterns — Not Started
   c. Wire `AnalyzeSpreaderScope` into the main analysis pipeline: add `SpreaderScope` handler to the `enterHandlers` map in `WalkSymbolTableWithHandlers` (or the appropriate handler setup location in `cpl_analyzer.go`) — Not Started
   d. Verify full parse→analyze pipeline: valid CPL spreader examples produce no errors, invalid CPL spreader examples produce meaningful error messages with source locations — run `go test ./src/main/go/analysis/... -count=1 -timeout 60s` to confirm no regressions — Not Started

## Test-First Development

This project follows the Logical TDD Lifecycle. For each coding task, write tests before implementing the code. Each task integrates its test-and-implement cycle — tests are not broken out as separate tasks. Follow the pattern described in `memory-bank/concepts.md` (Logical TDD Lifecycle):

1. Define the structural shape (function signatures, struct types)
2. Write the logical contract (failing tests that verify behavior)
3. Implement the code to satisfy the tests

## Constraints

- **Increment 1 scope**: Expression identifiers resolve to concepts only. Metric references are deferred to Increment 3. Do not add metric resolution to the expression evaluator.
- **All six functions**: The `SpreaderFunctionRegistry` registers all six functions from the start (unlike aggregators, which started with only `sum`). The `DimUnits` callback for each function must match the plan's specification.
- **Separate registry**: The spreader `FunctionRegistry` is separate from the aggregator `FunctionRegistry`. Create a new struct; do not reuse or extend the aggregator registry.
- **Dimensional unit formula**: Spreaders use `input_units * expression_units = output_units` — the inverse of the aggregator formula (`input_units = expression_units * output_units`).
- **Reuse existing types**: `DimensionalUnitExpression` (with `Multiply`/`Equals` methods) and `AggregationMatrix` must already exist from dependent stories. Do not redefine them.
- **Compound keys**: The `Spreaders` map uses compound string keys derived from sorted concept sets in `[fromSorted]->[toSorted]` format. Use the `SpreaderMapKey` function consistently.
- **No metric DAG integration**: `CILSpreadNode` is generated and stored in `CILTrees`, but not yet wired into metric calculation DAGs (deferred to Increment 2).
- **Error messages**: All analysis errors must include source locations for actionable diagnostics.

## Intent

**Task 1**: Establish the spreader function registry as a standalone, testable component. The registry provides dimensional unit semantics for spreader expressions and serves as the foundation for expression evaluation in Task 2a.

**Task 2a**: Build the core analysis machinery — scope iteration, concept cross-reference resolution, dimensional rule validation, and postfix expression evaluation. The expression evaluator handles all four operand kinds following the same stack-based pattern used by aggregators. This task produces a working expression evaluation pipeline for spreaders.

**Task 2b**: Complete the analysis pipeline by validating dimensional unit compatibility, constructing `SpreaderInfo` records, detecting duplicate spreader patterns, and integrating the `Spreaders` map into `AnalysisDefinitions`. This task ensures spreaders are fully validated and available for downstream phases.

**Task 3**: Generate `CILSpreadNode` from validated `SpreaderInfo` with pre-built `SpreadingMatrix`. The matrix uses equal-proportion weights derived from the distribution expression, reusing the `AggregationMatrix` struct. This task creates the CIL representation needed for synthesis.

**Task 4**: Verify the full parse→analyze pipeline end-to-end with example CPL files. Wire `AnalyzeSpreaderScope` into the handler map so it participates in the standard analysis walk, and confirm that valid spreaders pass analysis while invalid spreaders produce clear, source-located error messages.

## Acceptance Criteria

1. `SpreaderFunctionRegistry` registers all six functions with correct `DimUnits` callbacks
2. `AnalyzeSpreaderScope` walks symbol table and produces `SpreaderInfo` entries in the `Spreaders` map
3. From/to containment validated: all `from` concepts must appear in `to`; `len(to) == len(from) + 1`
4. Expression postfix evaluation produces correct `DimensionalUnitExpression` for all operand kinds
5. Dimensional unit validation enforces `input_units * expression_units = output_units`
6. Duplicate `(from, to)` patterns produce errors with source locations
7. Unknown concepts, identifiers, and functions in expressions produce errors with source locations
8. `CILSpreadNode` generated with pre-built `SpreadingMatrix` from concept value sets
9. Equal-proportion distribution (`1/size(C) * C`) produces uniform weights per concept value
10. `Spreaders` map integrated into `AnalysisDefinitions` and available via `GetAnalysisDefinitions()`
11. `AnalyzeSpreaderScope` wired into the handler map for `WalkSymbolTableWithHandlers`
12. `parse_and_analyze_only/without_errors/spreaders/` examples pass analysis
13. `parse_and_analyze_only/with_errors/spreaders/` examples produce meaningful error messages
14. All tests pass with `go test ./src/main/go/analysis/... -count=1 -timeout 60s`
