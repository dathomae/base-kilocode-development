# Story268: Spreaders Increment 2 — Synthesize (Epic008)

## Goal

Generate metric computation code with spreader integration: walk `CILSpreadNode` chains inserted during Increment 2 analysis, emit `tensorops.ExpandDimension` calls in topological computation order, and produce correct results for single unambiguous spreading paths. When the Increment 2 analyzer inserts `CILSpreadNode` chains into metric CIL trees (bridging dimension gaps between input references and output specifications), the synthesizer walks those chains and emits the corresponding `ExpandDimension` calls. Each spreader in the chain produces an intermediate tensor that becomes the source for the next spreader. The final spreader's output feeds into the metric's calculation body. Spreading matrices are emitted in the preamble (reusing `EmitSpreadingMatrix` from Story266). The `ChronoconeComputationController` handles topological scheduling — metrics with spreader inputs naturally run after their source metrics. Only single unambiguous spreader sequences are supported at this increment; no interleaved aggregator/spreader optimization (deferred to Increment 4).

## References

- [Plan: Story268 Spreaders Increment 2 — Synthesize](../plans/Story268_spreaders-increment2-synthesize.md)
- [ERO008: Spreaders Incremental Implementation](../specifications/design/cpl/eros/ERO008_spreaders-incremental-implementation.md)
- [ERO008: Spreaders Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO008_spreaders-synthesis-phase.md)
- [Epic008: CPL Spreaders Implementation](../epics/Epic008_cpl-spreaders-implementation.md)

## Dependencies

- [Story267_spreaders-increment2-analyze.md](Story267_spreaders-increment2-analyze.md) — Not Started. Analysis must insert `CILSpreadNode` chains into metric CIL trees via the second analysis pass
- [Story266_spreaders-increment1-synthesize.md](Story266_spreaders-increment1-synthesize.md) — Not Started. `WalkCILSpreadNode` handler and `EmitSpreadingMatrix` preamble primitive must exist
- [Story261_aggregators-increment4-synthesize.md](Story261_aggregators-increment4-synthesize.md) — Not Started. Metric computation code generation infrastructure (`ChronoconeComputationController`, topological scheduling) must exist; spreader synthesis follows the same pattern
- [Story254_aggregators-increment1-matrix-mult.md](Story254_aggregators-increment1-matrix-mult.md) — Not Started. `tensorops.ExpandDimension` is the core synthesis primitive

## Dependent Stories

- Story for Spreaders Increment 3 (synthesize) — extends synthesis for additional spreader function types and interleaved aggregator/spreader optimization

## Tasks

### Task Sizing

- **Task 1** is not `[Small]` — it creates new code infrastructure (`spreader_integration_synth.go` + test), introducing spreader chain walking logic that requires understanding of the CIL tree walker, `CILSpreadNode` type hierarchy, `tensorops.ExpandDimension` invocation patterns, and metric computation synthesis infrastructure. 5 subtasks spanning walker integration, variable naming, and test coverage.
- **Task 2** is not `[Small]` — it coordinates multiple synthesis concerns (preamble emission integrity, calculation body wiring to spreader-preprocessed tensors, topological order verification) across the same file area. Builds on Task 1's infrastructure; requires understanding of how generated metric functions, preamble variables, and the walker output integrate. 5 subtasks.
- **Task 3** is not `[Small]` — it creates multiple end-to-end example scenarios (single spreader, multi-step spreader, spread-then-aggregate) requiring understanding of spreader semantics, correctness of generated output, and the full CPL pipeline (parse → analyze → synthesize → execute). 5 subtasks spanning CPL authoring, IO spec design, input data creation, and expected output verification.
- **Task 4** is `[Small]` — it is pure verification: run the full test suite and confirm no regressions. Touches no source files; single concern.

### Parallel Execution

None — all tasks have sequential dependencies. Task 2 depends on Task 1's spreader chain walking infrastructure. Task 3 depends on the full synthesis pipeline (Tasks 1 and 2). Task 4 is verification after all code exists.

### Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4
```

- Task 1 must complete first (spreader chain walking is the core synthesis primitive that Task 2's function generation uses)
- Task 2 depends on Task 1 (metric function generation calls the chain walker; preamble emission must be verified against actual walker output)
- Task 3 depends on all preceding tasks (end-to-end examples verify the full pipeline: parse → analyze → synthesize → execute)
- Task 4 is verification after all code and examples exist

### Reintegration

None — no parallel groups.

---

### Task 1: Implement spreader chain walking in metric computation synthesis — Not Started

Create the synthesis logic that walks `CILSpreadNode` chains within metric CIL trees and emits the corresponding `tensorops.ExpandDimension` calls.

1. a. Create `src/main/go/synthesis/spreader_integration_synth.go` — implement the spreader chain walker that modifies the CIL tree walker's metric computation handler to properly walk `CILSpreadNode` chains. When a metric's input reference has a `CILSpreadNode` chain (inserted during Increment 2 analysis), the walker traverses the chain in order and emits `ExpandDimension` calls for each node. — Not Started

   b. Each `CILSpreadNode` in the chain emits one `tensorops.ExpandDimension` call. The result of one `ExpandDimension` becomes the source tensor for the next spreader in the chain. The last spreader's result variable is registered in the walk scope so the calculation body references the spreader-preprocessed tensor (not the raw input reference). — Not Started

   c. Intermediate variable naming uses the `spreadTo_<dimension>` pattern (e.g., `spreadTo_Country`, `spreadTo_Product`) to avoid collisions with other variable names in the generated function body. Each intermediate result receives a fresh, unique variable name. — Not Started

   d. Write unit tests in `src/main/go/synthesis/spreader_integration_synth_test.go`:
      - Single spreader chain: verify one `ExpandDimension` call emitted with correct `srcDim` and `newDim` indices
      - Multi-step spreader chain (scalar `[]` → `[Country]` → `[Country, Product]`): verify chain of `ExpandDimension` calls with correct source-to-result variable linking
      - Variable naming: verify intermediate variables use `spreadTo_<dimension>` pattern and are unique within the generated function
      - Scalar source (`from []`): verify `srcDim=0, newDim=0` edge case is handled correctly — Not Started

   e. Run `go test ./src/main/go/synthesis/... -count=1 -timeout 60s` and verify all walker tests pass. — Not Started

**Intent**: Provide the core synthesis building block for spreader integration: when the CIL tree walker encounters a metric input reference with attached `CILSpreadNode` chains, the walker emits the proper sequence of `ExpandDimension` calls, linking each spreader's output to the next spreader's input, and making the final result available to the calculation body.

**Acceptance Criteria**:
1. `CILSpreadNode` chains are correctly walked: each node emits exactly one `ExpandDimension` call
2. Intermediate result variables are correctly linked: output of spreader N becomes source of spreader N+1
3. The last spreader's result variable is what the calculation body references (not the raw input)
4. Variable naming follows `spreadTo_<dimension>` pattern with no collisions
5. All walker unit tests pass

---

### Task 2: Implement full metric function generation with spreader integration — Not Started

Generate complete metric computation functions that include spreader preprocessing before the calculation body. Ensure spreading matrix preamble variables are emitted for all spreaders used by the generated metrics.

1. a. Generate complete metric functions in the synthesis output that include spreader preprocessing (via the chain walker from Task 1) before the calculation body. The generated function structure follows: input retrieval → spreader chain preprocessing → calculation body → result return. — Not Started

   b. Ensure spreading matrix preamble variables are emitted for ALL spreaders used by the generated metrics. All spreading matrices referenced by the `ExpandDimension` calls in the generated metric functions must have corresponding `var spreadingMatrix_<Dimension> = &tensorops.AggregationMatrix{...}` declarations in the synthesis preamble. Reuse `EmitSpreadingMatrix` from Story266. — Not Started

   c. Handle the case where the metric's calculation body references the spreader-preprocessed tensor (the result of the last `CILSpreadNode` in the chain), NOT the raw input reference. The walk scope's variable registration (established in Task 1) ensures the calculation body picks up the correct variable. — Not Started

   d. Write unit tests in `src/main/go/synthesis/spreader_integration_synth_test.go` (extending Task 1's test file):
      - Full generated function: verify the function body contains both spreader preprocessing calls AND the calculation body, in correct order
      - Correct topological ordering: verify metrics with spreader inputs are scheduled after their source metrics in the generated `main()` controller setup
      - Error propagation: verify each `ExpandDimension` call is followed by an `if err != nil` check that returns a wrapped error with context (e.g., `"spread to country failed: %w"`)
      - Preamble integrity: verify the preamble contains spreading matrix declarations for all spreaders used by the generated metrics — Not Started

   e. Run `go test ./src/main/go/synthesis/... -count=1 -timeout 60s` and verify all function generation tests pass. — Not Started

**Intent**: Deliver complete, compilable metric functions with integrated spreader preprocessing. The generated code must be correct end-to-end: preamble matrices are declared, spreader chains are walked and emitted in order, the calculation body uses the correct preprocessed tensor, errors are propagated, and metrics appear in topological dependency order.

**Acceptance Criteria**:
1. Full metric functions generated include spreader preprocessing before calculation body
2. Spreading matrix preamble includes all matrices used by the generated metrics
3. Topological computation order ensures metrics with spreader inputs run after their sources
4. Error propagation: `ExpandDimension` errors are wrapped with context (e.g., `"spread to country failed: %w"`)
5. All function generation unit tests pass

---

### Task 3: Create end-to-end spreader integration examples — Not Started

Create end-to-end examples in `examples/with_data/spreaders_integration/` that exercise the full CPL pipeline (parse → analyze → synthesize → execute) with spreader-integrated metric computation. The example directory follows the existing `with_data/` template structure.

1. a. Create the `examples/with_data/spreaders_integration/` directory with the complete example infrastructure: CPL file defining spreaders and metrics with dimensional mismatches, IO spec file, input data (XLSX), and expected output (JSON). The directory structure follows the standard with_data example layout (e.g., `metrics_arithmetic/`, `metrics_conditionals/`). — Not Started

   b. **Example 1 — Single spreader**: Metric input has `[Product]` dimensions, output requires `[Country, Product]`. The spreader `from [Product] to [Country, Product]` bridges the gap via equal-proportion distribution (`1/size(Country) * Country`). Verify that generated code has one `ExpandDimension` call before the calculation body, and the output tensor has correct `[Country, Product]` shape with values evenly distributed across countries. — Not Started

   c. **Example 2 — Multi-step spreader**: Scalar input `[]`, output requires `[Country, Product]`. Two spreaders form a chain: `[] → [Country]` then `[Country] → [Country, Product]`. Verify generated code has two chained `ExpandDimension` calls with intermediate variable `spreadTo_Country` feeding into the second call, and the final output tensor has correct shape and values. — Not Started

   d. **Example 3 — Spread then aggregate**: Metric input has `[Product]` dimensions, output requires `[Country, Product]`, but the calculation includes an aggregator (e.g., `sum` over Country). The spreader expands from `[Product]` to `[Country, Product]`, then aggregation collapses back. Verify original values are preserved (spread-then-aggregate round-trip for equal-proportion distribution). — Not Started

   e. Run the verification: `go test ./src/main/go/... -count=1 -timeout 120s` and confirm all example tests pass with output matching `.expected.json`. — Not Started

**Intent**: Provide end-to-end validation that the full CPL pipeline correctly handles spreader-integrated metrics. Each example targets a distinct scenario (single spreader, multi-step chain, spread-then-aggregate) to build confidence in the synthesis output.

**Acceptance Criteria**:
1. Single spreader example produces correct `[Country, Product]` output with evenly distributed values
2. Multi-step spreader example produces correct intermediate shapes and final output
3. Spread-then-aggregate round-trip preserves original values (for equal-proportion distribution)
4. All example tests pass with output matching `.expected.json`

---

### Task 4: Verify full pipeline with tests — Not Started [Small]

Run the complete test suite and confirm the pipeline is intact with no regressions.

1. a. Run full test suite: `go test ./src/main/go/... -count=1 -timeout 120s` — verify all tests pass including metrics, aggregators, and the new spreader integration tests. — Not Started

   b. Verify no regressions in existing metric or aggregator tests — confirm that pre-existing test suites continue to pass unchanged. — Not Started

   c. Verify `with_data` examples execute and pass: the spreader integration examples (created in Task 3) produce output matching `.expected.json`, and all previously-passing `with_data` examples continue to pass. — Not Started

**Intent**: Confirm that the spreader integration synthesis does not break any existing functionality and that all new tests and examples pass cleanly.

**Acceptance Criteria**:
1. Full test suite passes (`go test ./src/main/go/...`)
2. No regressions in existing metric, aggregator, or any other test suites
3. All `with_data` examples (including new spreader integration examples) pass verification

---

## Test-First Development

This story follows the Logical TDD Lifecycle. For each coding task, write tests before implementing the code. The test-and-implement cycle is integrated within each task in the correct order:

- **Task 1**: Write walker tests (subtask d) before the chain walker implementation passes them (subtasks a–c).
- **Task 2**: Write function generation tests (subtask d) before the full metric function generation passes them (subtasks a–c).
- **Task 3**: Examples serve as integration tests — expected output (`.expected.json`) defines success before verification (subtask e).
- **Task 4**: Verification-only task — confirms all tests pass with no regressions.

Tests are NOT broken out as separate tasks. See `memory-bank/concepts.md` for details on the Logical TDD Lifecycle.

## Constraints

- `CILSpreadNode` chains must be present in metric CIL trees (inserted by Story267 analysis phase)
- `WalkCILSpreadNode` handler from Story266 must exist for individual `CILSpreadNode` walking
- `EmitSpreadingMatrix` from Story266 must exist for preamble variable emission
- `tensorops.ExpandDimension` and `tensorops.AggregationMatrix` types must exist (built in Story254)
- `ChronoconeComputationController` infrastructure must exist (built in Story261)
- Spreading matrices are pre-built from concept value sets at compile time and emitted as Go struct literals
- Only single unambiguous spreader sequences are supported — the analysis phase guarantees exactly one spreader per dimension gap
- Each `CILSpreadNode` in the chain has a single, non-nil `SpreadingMatrix`
- The `FromDimensions` of each spreader matches the output dimensions of the previous spreader (or input source)
- The final spreader's `ToDimensions` matches the metric's IO spec output dimensions
- All generated code must compile under `source_root` (`src/chronocone_planning_language`)
- Generated code follows the existing pattern: `check()` helper for error handling, `ChronoconeTensor[float64]` for tensor types

## Intent

This story delivers the synthesis phase for spreader-integrated metric computation. When the Increment 2 analyzer inserts `CILSpreadNode` chains into metric CIL trees, the synthesizer walks those chains and emits the corresponding `tensorops.ExpandDimension` calls. Each spreader in the chain produces an intermediate tensor that becomes the source for the next spreader. The final spreader's output feeds into the metric's calculation body. Spreading matrices are emitted in the preamble (reusing `EmitSpreadingMatrix` from Story266). The `ChronoconeComputationController` handles topological scheduling — metrics with spreader inputs naturally run after their source metrics. The generated code follows the same pattern as aggregator-integrated metric computation (Story261): input retrieval → preprocessing (spreader chain here) → calculation body → result return.

The walker must ensure that when a metric's input has a `CILSpreadNode` chain, the calculation body references the result of the LAST `CILSpreadNode` in the chain, not the raw input reference. This is handled by the walk scope's variable registration — the last spreader registers its result variable, which is what the calculation body picks up.

## Acceptance Criteria

1. `CILSpreadNode` chains are correctly walked: each node emits one `ExpandDimension` call
2. Intermediate result variables are correctly linked: output of spreader N becomes source of spreader N+1
3. Full metric functions generated include spreader preprocessing before calculation body
4. Spreading matrix preamble includes all matrices used by the generated metrics
5. Topological computation order ensures metrics with spreader inputs run after their sources
6. Error propagation: `ExpandDimension` errors are wrapped with context
7. End-to-end examples produce correct output matching `.expected.json`
8. Spread-then-aggregate round-trip preserves original values (for equal-proportion distribution)
9. Multi-step spreader sequences produce correct intermediate shapes
10. All existing tests continue to pass
