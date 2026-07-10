# Story253: Lookups Increment 5 — Synthesize

## Goal

Generate lookup access code within spreader computation, emit implicit aggregator code for lookups requiring dimensional alignment, ensure correct execution ordering in the preamble, and provide a spreader+lookup end-to-end example. This story delivers the synthesis phase for lookup references in spreader expressions (Increment 5 of [Epic006: CPL Lookups Implementation](../epics/Epic006_cpl-lookups-implementation.md)).

When a spreader CIL tree walker encounters a lookup reference node, it emits `controller.GetMetricComputation(lookupName).GetCell(dimCoords)`. If the analysis phase recorded implicit transformation nodes (aggregators) on the lookup to align its dimensions with the spreader context, the synthesizer generates the corresponding aggregation code before the lookup data is used in the spreader expression. The preamble generator ensures lookup data is loaded before any spreader that references it, and that lookup `MetricComputation` objects are not deregistered until all consuming spreader-dependent metrics complete.

No new synthesis infrastructure is required at the lookup level — the spreader CIL tree walker (defined in ERO008) emits the lookup access code as part of the spreading code generation. Lookup data loading and `MetricComputation` registration reuse the infrastructure from Stories 243, 246, and 249.

## References

- [requirements](../specifications/requirements/requirements.md)
- [design](../specifications/design/design.md)
- [ERO006: Lookups Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO006_lookups-synthesis-phase.md)
- [ERO006: Lookups Incremental Implementation](../specifications/design/cpl/eros/ERO006_lookups-incremental-implementation.md)
- [ERO008: Spreaders Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO008_spreaders-synthesis-phase.md)
- [Plan: Story253 Lookups Increment 5 Synthesize](../plans/Story253_lookups-increment5-synthesize.md)

## Dependencies

- **Epic008 (Spreaders) at Increment 3 or above** — spreader CIL tree walker, spreader `CILSpreadNode` with resolved lookup references and implicit transformation nodes, and code generation infrastructure must exist
- [Story252_lookups-increment5-analyze.md](Story252_lookups-increment5-analyze.md) — Not Started (must have resolved lookup references and implicit transformation nodes in spreader CIL nodes)
- [Story243_lookups-increment1-synthesize.md](Story243_lookups-increment1-synthesize.md) — Not Started (lookup loading infrastructure and `MetricComputation` registration must exist)

## Dependent Stories

None — this story completes the lookups synthesis phase for spreader integration. No further stories depend on this story.

## Tasks

### Task Sizing

Three tasks route to `code-for-story-implementor`; two tasks are annotated `[Small]`:

- **Task 1** modifies the spreader CIL tree walker to emit lookup access code — requires understanding the walker's code generation pattern, coordinate argument construction from dimensional context, and Go code emission. Not a mechanical switch-case addition.
- **Task 2** creates implicit aggregator code generation for lookups — multiple test scenarios across aggregation-needed, no-transformation-needed, and nested-spreading cases. Requires reasoning about dimensional alignment and integrated spreader expression handling.
- **Task 3** ensures correct execution ordering in the preamble generator — requires reasoning about the full execution pipeline (lookup loading → implicit aggregator execution → spreader execution → metric computation) and verifying deregistration timing. Despite touching only 1–2 files, the cross-component coordination complexity means this task routes to the medium-code agent.
- **Task 4** is `[Small]` — creates a single example directory following the existing `with_data/` template pattern. One concern, no new infrastructure.
- **Task 5** is `[Small]` — verification-only task: runs existing test suites to confirm no regressions. No code changes, no new test files.

### Parallel Execution

None — all tasks have sequential dependencies. Each task builds on the output of the previous task. Task 4 requires the full synthesis pipeline from Tasks 1–3; Task 5 verifies all changes including Task 4's example.

### Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4 → Task 5
```

- Task 1 must complete first (lookup access code emission is the core synthesis primitive)
- Task 2 depends on Task 1 (implicit aggregator code builds on the lookup access pattern)
- Task 3 depends on Tasks 1 and 2 (execution ordering verification requires both lookup access and aggregator code to exist)
- Task 4 depends on Tasks 1–3 (end-to-end example requires the full synthesis pipeline)
- Task 5 is verification after all code exists

### Reintegration

None — no parallel groups.

### Task 1: Emit lookup access code in spreader synthesis — Not Started

**Intent**: Extend the spreader CIL tree walker to emit `controller.GetMetricComputation(lookupName).GetCell(dimCoords)` when encountering a lookup reference node in the spreader's expression. Construct the coordinate arguments from the spreader's dimensional context so that the generated code correctly indexes into the lookup's tensor at runtime. This enables spreader expressions to access lookup data for weighted distribution calculations.

   a. In the spreader CIL tree walker (`src/main/go/synthesis/spreader_synthesis.go`), when a lookup reference CIL node is encountered during expression walking, emit a call to `controller.GetMetricComputation(lookupName).GetCell(dimCoords)` — Not Started

   b. Construct the coordinate arguments from the spreader's dimensional context: determine which dimensions index into the lookup tensor and generate the correct coordinate variable names in the emitted Go code — Not Started

   c. Write unit tests in `src/main/go/synthesis/spreader_synthesis_test.go`: verify the emitted Go code includes the correct `GetMetricComputation` call, verify coordinate arguments match the spreader's dimensional context, and verify the lookup variable is used in the spreader expression — Not Started

   d. Run `go test ./src/main/go/synthesis/... -count=1 -timeout 60s` and verify all synthesis tests pass — Not Started

**Acceptance Criteria**:
1. The spreader walker emits `controller.GetMetricComputation(lookupName).GetCell(dimCoords)` for each lookup reference
2. Coordinate arguments are constructed from the spreader's dimensional context
3. Unit tests verify the emitted code structure and coordinate correctness

### Task 2: Generate implicit aggregator code for lookups — Not Started

**Intent**: When the analysis phase (Story252) recorded implicit transformation nodes (aggregators) on a lookup to align its key dimensions with the spreader context, generate the corresponding aggregation code. The lookup data must be transformed to the required dimensional form before it is used in the spreader expression. This handles cases where, for example, a lookup keyed by `[Country]` needs to be aggregated to `[]` before being used in a scalar spreader expression.

   a. When the spreader CIL node has implicit transformation nodes for a lookup, emit aggregation code that transforms the lookup's data to the required dimensional form — Not Started

   b. Emit code that reduces the lookup's dimensions using `tensorops.ReduceDimension` (or equivalent) to match the spreader context — Not Started

   c. Ensure the transformed lookup data (not the raw lookup data) is used in the spreader expression — Not Started

   d. Write unit tests covering three scenarios: (1) lookup needs implicit aggregation before spreading, (2) lookup dimensions already match the spreader context (no transformation needed), and (3) lookup needs spreading before spreading (nested spreader on lookup) — Not Started

   e. Run `go test ./src/main/go/synthesis/... -count=1 -timeout 60s` and verify all tests pass — Not Started

**Acceptance Criteria**:
1. Implicit aggregator code is generated when a lookup needs dimensional alignment
2. The transformed lookup data is used in the spreader expression, not the raw lookup
3. Lookups with matching dimensions produce no implicit aggregator code
4. All three test scenarios pass

### Task 3: Ensure correct execution ordering in preamble generation — Not Started

**Intent**: Verify and enforce the correct execution order in the generated preamble: lookup data must be loaded before any spreader that references a lookup runs; implicit aggregators on lookups must execute before the spreader they support; and lookups must not be deregistered until all consuming spreader-dependent metrics have completed. The preamble generator (`src/main/go/synthesis/preamble_generator.go`) is modified to guarantee this ordering.

   a. Verify that `loadLookupData(controller)` is called before any spreader that references a lookup — if the current code does not guarantee this, add ordering enforcement — Not Started

   b. Verify that implicit aggregators on lookups run before the spreader that consumes the transformed lookup data — Not Started

   c. Ensure lookups referenced in spreaders are not deregistered until all consuming spreader-dependent metrics complete — Not Started

   d. Write an integration test that verifies the generated preamble ordering: loadLookupData → implicit aggregators → spreader execution → metric computation — Not Started

   e. Run `go test ./src/main/go/synthesis/... -count=1 -timeout 60s` and verify the integration test passes — Not Started

**Acceptance Criteria**:
1. `loadLookupData` runs before any spreader that references a lookup
2. Implicit aggregators on lookups run before the spreader
3. Lookups are not deregistered until all consuming metrics complete
4. Integration test verifies the preamble ordering is correct

### Task 4: Create with_data/lookups_in_spreaders example — Not Started [Small]

**Intent**: Create an end-to-end example demonstrating a spreader that uses a lookup-based weighting table. The example distributes a global target across countries using country-specific weights from a lookup, verifying that the full pipeline (parse → analyze → synthesize → run) produces correct output.

   a. Create the directory `examples/with_data/lookups_in_spreaders/` with a CPL file defining: a `country_weight` lookup keyed by `[Country]` with celltype `WeightFactor`, a `global_target` connector with scalar data, and a spreader that distributes `global_target` across countries proportionally to the lookup weights — Not Started

   b. Create the IO spec and parquet/XLSX data files for the lookup and connector — Not Started

   c. Create `.expected.json` with the expected output after spreading with lookup-based weighting — Not Started

   d. Run the end-to-end pipeline and verify the output matches `.expected.json` — Not Started

**Acceptance Criteria**:
1. CPL file compiles and parses successfully
2. Analysis resolves lookup references in the spreader expression
3. Synthesis generates correct Go code with lookup access and implicit aggregation
4. Generated executable produces output matching `.expected.json`

### Task 5: Verification and regression — Not Started [Small]

**Intent**: Confirm that the synthesis changes do not break existing functionality. Run the existing lookups examples (Increments 1–4) and the full CPL test suite to verify zero regressions.

   a. Verify existing lookups examples from Increments 1–4 still pass end-to-end (parse → analyze → synthesize → run → output matches expected) — Not Started

   b. Run the full CPL test suite: `go test ./src/main/go/... -count=1 -timeout 120s` — confirm all tests pass with zero regressions — Not Started

**Acceptance Criteria**:
1. Existing Increment 1–4 lookups examples continue to pass
2. Full CPL test suite passes with no failures

## Test-First Development

This project follows the Logical TDD Lifecycle (see `memory-bank/concepts.md`). For each coding task, tests are written before the implementation code. The test-and-implement cycle is integrated within each task rather than broken out as separate tasks. The ordering of subtasks within each task reflects this: test writing precedes the verification step, and the implementation must make the tests pass.

## Constraints

- No new synthesis infrastructure at the lookup level — lookup access in spreaders reuses the same `controller.GetMetricComputation` pattern established in Stories 243 and 249
- The spreader CIL tree walker and code generation infrastructure (from Epic008 Increment 3+) must be available as prerequisites
- The analysis phase (Story252) must provide resolved lookup references and implicit transformation nodes in spreader CIL nodes
- Lookup `MetricComputation` registration and loading reuse the infrastructure from Story243
- All file paths are relative to `source_root` (`src/chronocone_planning_language`); Go source files include the `src/main/go/` prefix
- The spreader synthesis file name (`spreader_synthesis.go`) is indicative — the implementor must align with the actual file names from the Epic008 spreader synthesis implementation

## Intent

This story enables lookup data access within spreader computations. Since no new synthesis infrastructure is needed at the lookup level, the work focuses on the spreader CIL tree walker emitting `controller.GetMetricComputation(lookupName)` calls, generating implicit aggregator code when lookups need dimensional alignment, and the preamble generator ensuring correct execution ordering. Lookup access in spreaders reuses the same `GetMetricComputation` pattern established in Stories 243 and 249. This story is the final synthesis increment for lookups and requires Epic008's spreader synthesis infrastructure at Increment 3 or above.

### Task Intent Summary

| Task | Intent |
|------|--------|
| Task 1 | Emit `controller.GetMetricComputation(lookupName).GetCell(dimCoords)` in spreader code generation |
| Task 2 | Generate implicit aggregator code when lookups need dimensional alignment with spreader context |
| Task 3 | Ensure correct preamble ordering: lookup loading → implicit aggregators → spreader execution |
| Task 4 | End-to-end example: spreader using lookup-based weighting table |
| Task 5 | Regression verification: existing examples still pass, no test failures |

## Notes

- The actual spreader synthesis file names and structure are determined by the Epic008 spreader stories. The `spreader_synthesis.go` name used in this story is a placeholder — the implementor must locate the correct spreader CIL tree walker file where `WalkCILSpreadNode` (or equivalent) is defined. If the spreader synthesis is split across multiple files (e.g., `cil_spread_walker.go` for the walker, `spreader_integration_synth.go` for integration, `spreader_metric_ref_synth.go` for metric references), the implementor should modify the appropriate file(s).
- The `preamble_generator.go` file is the existing preamble generator from the synthesis package, which already contains lookup loading code from Story243.
- The example directory follows the `examples/with_data/` pattern established in earlier lookups stories. Use XLSX format for input data consistent with the project's current convention.
