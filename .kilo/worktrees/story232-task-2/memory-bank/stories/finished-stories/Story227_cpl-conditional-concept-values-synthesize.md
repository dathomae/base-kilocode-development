# Story227: CPL Conditional Concept Values Synthesis

## Goal

Enable the CPL synthesis pipeline to correctly handle conditional concept value declarations when they are used as grain dimensions in connector declarations. Specifically, when a connector has `grain (Parent, Child)` where `Child` has values that vary based on `Parent` values (for example, `DayInMonth` values depend on which `Month` is selected), the synthesized pipeline must produce correctly-shaped tensors, correctly-populated cell data, and bit-packed validity masks that distinguish valid cells from invalid ones.

## References

- [Conditional Synthesis Plan](../plans/conditional_synthesis_plan.md)
- [Story188_cpl-metrics-conditionals-synthesize.md](Story188_cpl-metrics-conditionals-synthesize.md) — provides mask infrastructure (`ChronoconeTensor.mask`, `SetMask()`, bit convention, mask-aware arithmetic)
- [Story121_cpl-concepts-parser-conditional.md](finished-stories/Story121_cpl-concepts-parser-conditional.md) — done
- [Story122_cpl-concepts-analyzer-conditional.md](finished-stories/Story122_cpl-concepts-analyzer-conditional.md) — done
- [Story208_fix-conditional-spec-pipeline-defects.md](Story208_fix-conditional-spec-pipeline-defects.md) — done
- `src/chronocone_planning_language/src/main/go/synthesis/synthesize.go` — `BuildConceptValuesMap`, `GenerateMainFunction`
- `src/chronocone_planning_language/src/main/go/io/parquet_reader.go` — row count validation
- `src/chronocone_planning_language/src/main/go/io/generate_parquet.go` — parquet generation
- `src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go` — mask infrastructure

## Dependencies

- [Story188_cpl-metrics-conditionals-synthesize.md](Story188_cpl-metrics-conditionals-synthesize.md) — _Done_ — provides mask infrastructure: `ChronoconeTensor.mask`, `SetMask()`, bit convention (bit 0 = valid), and mask-aware arithmetic functions
- [Story121_cpl-concepts-parser-conditional.md](finished-stories/Story121_cpl-concepts-parser-conditional.md) — _Done_ — parser support for conditional concept values
- [Story122_cpl-concepts-analyzer-conditional.md](finished-stories/Story122_cpl-concepts-analyzer-conditional.md) — _Done_ — analyzer populates `AnalysisConceptDefinition.ConditionalValues`
- [Story208_fix-conditional-spec-pipeline-defects.md](Story208_fix-conditional-spec-pipeline-defects.md) — _Done_ — fixed analysis pipeline defects for conditionals
- Existing parquet reader, generator, and synthesis infrastructure must be functional

## Dependent Stories

- [Story215_cpl-metrics-parallel-branch-execution.md](Story215_cpl-metrics-parallel-branch-execution.md) — _Not Started_ — benefits automatically from conditional concept masks being pre-set on tensors
- [Story216_cpl-metrics-boolean-body-expressions.md](Story216_cpl-metrics-boolean-body-expressions.md) — _Not Started_ — benefits automatically from conditional concept masks being pre-set on tensors

## Tasks

### Task Sizing

- **Task 1** (`BuildConceptValuesMap` enrichment): Medium — introduces a new nested map data structure, requires coordination between synthesis code and generated code emission, touches `synthesize.go` and its test file. *(Completed — correctly sized as Medium.)*
- **Task 2** (mask computation and embedding): `[Small]` — implementation of `ComputeValidityMask` and connector-loader mask emission is already complete in `synthesize.go`. Remaining work is writing tests for `ComputeValidityMask` in `synthesize_test.go` and verifying behavior. Narrow scope, single concern.
- **Task 3** (reader relaxation): `[Small]` — single condition change in one source file and one test file. Narrow scope, follows an existing pattern.
- **Task 4a** (conditional generator function + tests): Medium — introduces a new `GenerateParquetFileWithConditionals` function with conditional iteration logic and multiple edge cases. Touches `generate_parquet.go` and its test file. Correctly classified as Medium.
- **Task 4b** (conditional generator CLI flag): `[Small]` — adds an optional CLI flag in `cli/parquetgen/parquet_gen.go` plus test additions. Additive change, follows existing CLI patterns, at most 2 files.
- **Task 5** (example creation): Medium — creates multiple new files (CPL, Go script, shell script, README). Cross-cutting but well-scoped.
- **Task 6** (regression): `[Small]` — runs existing test suites, no code changes.

### Parallel Execution

- **Group A** (Tasks 3, 4a): Both tasks modify files in the `io/` package. Even though the specific files are disjoint (`parquet_reader.go` vs `generate_parquet.go`), same-package compilation risk means worktree isolation is required.
- **Cross-package parallelism**: The sequential chain `[Task 1 → Task 2]` (synthesis package) and the parallel group `[Task 3 ‖ Task 4a]` (io package) are independent and can execute concurrently. Task 4b (cli package) is file-disjoint from both but semantically depends on Task 4a, so it joins the flow after Group A completes.

### Execution Order

```
[Task 1 → Task 2] ‖ [Task 3 ‖ Task 4a] → Task 4b → Task 5 → Task 6
```

- Tasks 1 and 2 are sequential because Task 2 depends on the `conditionalConceptValues` data structure created in Task 1.
- Tasks 3 and 4a run in parallel within a worktree group (same `io/` package).
- Task 4b runs after Group A because it semantically depends on the `GenerateParquetFileWithConditionals` function introduced in Task 4a. It can run concurrently with the synthesis chain `[Task 1 → Task 2]` since `cli/parquetgen/` is package-disjoint from `synthesis/`.
- Task 5 depends on all preceding tasks because the end-to-end example needs both the synthesis mask and the parquet reader/generator changes.
- Task 6 (regression) runs last.

### Reintegration

**Group A (Tasks 3, 4a):**
- Merge order: any order (disjoint files — `parquet_reader.go` vs `generate_parquet.go`)
- Integration tests: `go test ./src/main/go/io/... -count=1 -timeout 60s` — verify combined reader + generator behavior
- Watch for conflicts in: none expected

### Task 1: Enrich BuildConceptValuesMap with conditional parent→child mappings — Completed
   a. Write tests for `BuildConditionalConceptValuesMap` covering a concept with conditional values, a concept without conditionals, and multiple parent values each with different child sets. — Completed
   b. Implement `BuildConditionalConceptValuesMap` that extracts `ConditionalValues` from `AnalysisConceptDefinition` for all concepts that have them. — Completed
   c. Compute connector validity masks in `GenerateMainFunction` via `BuildConnectorValidityMasks`, and emit `connectorValidityMasks` into the generated Go code. The `conditionalConceptValues` intermediate map is used only at synthesis time and is not emitted into the generated program (masks are precomputed at synthesis time, making runtime `conditionalConceptValues` unnecessary). — Completed (design refined)
   d. Run synthesis tests to confirm the map contains entries only for concepts with conditional specifications, that concepts without conditionals are absent, and that the existing flat `conceptValues` map continues to work unchanged. — Completed

### Task 2: Compute and embed validity masks for conditional-dimension tensors — Completed [Small]
    a. Write tests for `ComputeValidityMask` covering the Month/DayInMonth example, a single conditional case, no conditionals, and bit-packing correctness across uint64 boundaries. — Completed
    b. Run tests to verify the mask bit for Feb 30 is 1 (invalid), the mask bit for Jan 15 is 0 (valid), and that tensors without conditional dimensions do not receive a mask. — Completed
    *Note: Implementation of `ComputeValidityMask` and connector-loader mask emission (`SetMask` calls) is already complete in `synthesize.go`. Remaining work is test-writing and verification only.*

### Task 3: Relax parquet reader row count validation — Completed [Small]
    a. Write tests for the parquet reader covering a file with fewer rows than the cartesian product, a file with excess rows, a full-product file, and a zero-row file. — Completed
    b. Change the row count check in `parquet_reader.go` from `len(rows) == product(shape)` to `len(rows) <= product(shape)`. — Completed
    c. Run `go test ./io/...` to confirm full-product files still read correctly, that unpopulated cells contain zero values, and that files with more rows than the product still produce errors. — Completed

### Task 4a: Implement `GenerateParquetFileWithConditionals` function and tests — Completed
    a. Write tests for `GenerateParquetFileWithConditionals` covering valid combinations only, flat dimensions without conditionals, a missing parent value, and an empty child set. — Completed
    b. Implement `GenerateParquetFileWithConditionals` that accepts `conditionalDims map[string]map[string][]string`, iterates parent values, looks up child values per parent, and produces rows only for valid combinations. — Completed
    c. Run tests to confirm conditional generation produces only valid rows and that the existing full-cartesian generation path continues to work unchanged. — Completed

### Task 4b: Add conditional mappings CLI flag to parquetgen — Completed [Small]
   a. Add an optional `--conditional-mappings` flag (or equivalent) to `cli/parquetgen/parquet_gen.go` that accepts a JSON file specifying conditional dimension mappings. — Completed
   b. Parse the conditional mapping JSON and pass it to `GenerateParquetFileWithConditionals`. — Completed
   c. Add CLI-level tests for the new flag. — Completed
   d. Run tests to verify the CLI produces correct parquet files with conditional dimensions. — Completed

### Task 5: Create end-to-end CPL example with conditional grain concepts — Completed
    a. Create `core.cpl` declaring `Month` and conditional `DayInMonth` concepts, plus a connector with `grain (Month, DayInMonth)`. — Completed
    b. Create `generate_data.go` script that uses `GenerateParquetFileWithConditionals` to produce a parquet file with only valid month/day combinations. — Completed
    c. Create `run_test.sh` that compiles the synthesized Go program and verifies output shape, mask presence, and cell correctness. — Completed
    d. Run the full example pipeline (parse → analyze → synthesize → compile → run) and verify the tensor shape is dense, the validity mask marks invalid cells, and valid cells have expected values. — Completed

### Task 6: Run full test suite regression — Completed [Small]
    a. Run `go test ./... -count=1` from `src/chronocone_planning_language/src/main/go/` and verify all packages pass. — Completed (all 19 packages pass)
    b. Run existing example verification scripts and confirm non-conditional grain examples still produce correct output. — Completed (pre-existing conditional_when_default parse issue unrelated)
    c. Verify Story188 mask infrastructure tests still pass. — Completed (calculation package: PASS)

## Test-First Development

This project follows the Logical TDD Lifecycle. For each coding task above, tests are written before implementation and the test-and-implement cycle is integrated within each task rather than broken out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- Must follow the Logical TDD Lifecycle — write tests before implementation changes.
- Must not break existing connector, metrics, or concept functionality.
- The dense tensor strategy must be consistent: every dimension pair with conditional relationships produces valid-max-size tensors with correct masks.
- Parquet files with full cartesian product (non-conditional grain) must continue to work.
- Must use the existing `ChronoconeTensor.mask` field and bit convention from Story188 (bit 0 = valid, bit 1 = invalid).
- Mask computation must handle multi-dimensional tensors where multiple child concepts are conditional on different parents.
- The `conditionalConceptValues` structure must handle nesting: if `DayInMonth` is conditional on `Month`, and `HourOfDay` is conditional on `DayInMonth`, both conditionals must be reflected in the mask.

## Intent

1. **Task 1 (Data structure enrichment)**: Preserve conditional parent→child value mappings from analysis through synthesis so that mask computation has access to which combinations are valid.
2. **Task 2 (Mask computation)**: Compute and embed bit-packed validity masks at synthesis time so downstream metric computation automatically skips invalid cells without runtime conditional resolution.
3. **Task 3 (Reader relaxation)**: Allow parquet files to contain only semantically valid rows rather than requiring the full cartesian product.
4. **Task 4a (Conditional generator function)**: Enable generation of parquet files that contain only valid (parent, child) combinations via the `GenerateParquetFileWithConditionals` library function.
5. **Task 4b (Conditional generator CLI)**: Expose conditional parquet generation via an optional CLI flag in the parquetgen tool.
6. **Task 5 (End-to-end example)**: Demonstrate and validate the full pipeline from CPL declaration through synthesis, parquet generation, and execution.
7. **Task 6 (Regression)**: Guarantee that no existing functionality is broken by the changes.

## Acceptance Criteria

1. A CPL file with conditional concept grain dimensions (for example, Month → DayInMonth) parses, analyzes, and synthesizes without errors.
2. The synthesized Go code includes `connectorValidityMasks` map for connectors with conditional grain dimensions. (The `conditionalConceptValues` structure is consumed at synthesis time to precompute masks; it is not emitted into the generated program.)
3. The synthesized Go code compiles and allocates a dense tensor with shape equal to the cartesian product of dimension value sets.
4. The tensor's validity mask correctly marks valid cells (bit 0) and invalid cells (bit 1) — for example, Feb 30 is invalid and Jan 15 is valid.
5. A parquet file containing only valid (parent, child) combinations is read successfully by the generated Go code.
6. The parquet generator can produce files with only valid conditional combinations.
7. All existing tests pass with zero regressions across all packages.
8. All existing examples continue to produce correct output.

## Footer

### Open Questions

None

### Known Issues

None

### Additional Context

#### Why `BuildConceptValuesMap` now needs enrichment

Previously the plan deferred mask-related work and asserted `BuildConceptValuesMap` did not need changes. With the validity mask now in scope, the conditional parent→child value mappings (`ConditionalValues`) must flow from analysis into synthesis. The flat `conceptValues` map (used for shape computation) stays unchanged, but a parallel `conditionalConceptValues` map (used for mask computation) is added.

#### Nested conditionals

When a concept hierarchy has multiple levels of conditional relationships (for example, `Year → Month → DayInMonth` where `DayInMonth` is conditional on `Month`, and `Month` is conditional on `Year` for leap years), the mask computation must handle the full chain. The flat index calculation for the tensor is the cartesian product of all dimensions, and each conditional relationship independently contributes to the mask by invalidating specific bit positions. The computation iterates over all conditional relationships and clears valid-cell bits; any bit not cleared remains invalid.

#### Deferred concerns

- **Grandchild conditionals** (where child values depend on parent, and grandchild values depend on child): initially supported by the mask computation using the full dimensional space. Per-parent-child value resolution may need enhancement in a future story for deeply nested conditional hierarchies.
- **Dynamic concept values** (loaded from external API at runtime): the current design assumes all concept values are known at synthesis time. If concepts gain runtime-loaded values in a future increment, mask computation would need to move to runtime or accept a mask-update mechanism.

### Mask Architecture

The validity mask for conditional concept dimensions uses the same bit convention and data structure as the conditional metric masks introduced in Story188:
- `ChronoconeTensor.mask` field: `[]uint64`
- Bit 0 = valid/selected (the cell should be computed)
- Bit 1 = invalid/skipped (the cell should be excluded)
- Zero-initialized mask = all-valid (backward compatible with non-conditional tensors)

For a tensor with dimensions `(Month, DayInMonth)` where `Month` values are `[Jan, Feb, Mar]` (indices 0, 1, 2) and `DayInMonth` values are `[1..31]` (indices 0..30), the flat index for cell `(monthIndex, dayIndex)` is `monthIndex * 31 + dayIndex`. The mask is a `[]uint64` where bits are packed left-to-right within each uint64. Bits 0..92 are set per valid/invalid status; bits 93+ are 1 (invalid, beyond tensor bounds).

The existing `TensorAddFloat64`, `TensorMultiplyFloat64`, and so on already respect masks from Story188. Invalid cells are skipped during computation, so no changes to `chronocone_tensor.go` arithmetic functions are required.
