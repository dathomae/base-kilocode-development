# Story257: Aggregators Increment 2 — Synthesize (Epic007)

## Goal

Implement CIL tree walking for `CILAggregateNode`, emission of aggregation matrix Go code in the synthesis preamble, manual sum aggregation test code, and end-to-end aggregator examples with CPL + parquet + expected output. This is Increment 2 of Epic007 (Aggregators), delivering the synthesis phase for basic aggregators (sum only). The CIL tree walker handles `CILAggregateNode` by emitting calls to `tensorops.ReduceDimension`; aggregation matrices are emitted as Go struct literals in the synthesis preamble before metric computation functions.

## References

- [ERO007: Aggregators Incremental Implementation](../specifications/design/cpl/eros/ERO007_aggregators-incremental-implementation.md)
- [ERO007: Aggregators Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO007_aggregators-synthesis-phase.md)
- [Epic007: CPL Aggregators Implementation](../epics/Epic007_cpl-aggregators-implementation.md)
- [Source root: src/chronocone_planning_language](../../src/chronocone_planning_language/)

## Dependencies

- Story254 (aggregators-increment1-matrix-mult) — `tensorops.ReduceDimension` and `tensorops.AggregationMatrix` types must exist
- Story255 (aggregators-increment2-parse) — parser must produce symbol table entries for aggregator declarations
- Story256 (aggregators-increment2-analyze) — analysis must produce `CILAggregateNode` nodes with pre-built aggregation matrices in CIL trees

## Dependent Stories

- [Story258_aggregators-increment3-integrate.md](Story258_aggregators-increment3-integrate.md) — Depends on full synthesis pipeline being functional

## Tasks

### Task Sizing

- **Task 1** is not `[Small]` — it introduces a new walker method with multi-structure reasoning (CILAggregateNode → tensorops.ReduceDimension code emission, dimension index computation, error handling generation). Modifies 2 source files (cil_transformer.go + postfix_walk.go dispatch) plus a new test file. The walker method and dispatch wiring are tightly coupled.
- **Task 2** is `[Small]` — it adds a single method to `CILCodeGenerator` in `src/main/go/cil/cil_code_generator.go` with a corresponding test file. The change is additive, follows the existing `GenerateConnectorMetadataMap` pattern, and addresses one concern (formatting `AggregationMatrix` as Go struct literal). Touches at most 2 files.
- **Task 3** is not `[Small]` — it modifies the synthesis preamble generator (`synthesize.go`, ~1000 lines) and must integrate aggregation matrix variables before metric computation functions. Involves understanding the synthesis pipeline's boundary between preamble and body generation, import block management, and `GeneratedProgram.Variables` construction.
- **Task 4** is not `[Small]` — it creates a complete with_data example directory requiring CPL authoring, parquet data generation, ground-truth xlsx computation, and a manual aggregation test. Multiple coordinated file types (CPL, JSON, parquet, xlsx, Go test) that must be internally consistent.

Tasks 1, 3, and 4 route to `code-for-story-implementor`. Task 2 routes to `small-code-for-story-implementor`.

### Parallel Execution

None — all tasks have sequential dependencies. Task 2 needs Task 1's walker interface; Task 3 needs Task 2's matrix emission; Task 4 needs Task 3's full pipeline.

### Execution Order

```
Task 1 → Task 2 → Task 3 → Task 4
```

### Reintegration

None — no parallel groups.

### Test-First Development

This story follows the Logical TDD Lifecycle. For each coding task, write tests before implementing the code. The test-and-implement cycle is integrated within each task in the correct order:

- **Task 1**: Write tests for CILAggregateNode walking (subtasks c–d) before the walker implementation passes them.
- **Task 2**: Write tests for aggregation matrix emission (subtask c) before the `GenerateAggregationMatrix` method returns valid output.
- **Task 3**: The integration test verifies generated program compilation after preamble integration.
- **Task 4**: Example tests verify full end-to-end pipeline behavior.

Tests are NOT broken out as separate tasks. See `memory-bank/concepts.md` for details on the Logical TDD Lifecycle.

### Constraints

- The `tensorops.ReduceDimension` function and `tensorops.AggregationMatrix` type must exist (built in Story254)
- The `CILAggregateNode` type with pre-built aggregation matrices must exist in the CIL node hierarchy (built in Story256)
- Aggregation matrices are pre-built from concept value sets at compile time and emitted as Go struct literals
- Aggregators reduce tensor dimensionality by exactly one per aggregator
- All generated code must compile under `source_root` (`src/chronocone_planning_language`)
- Generated code follows the existing pattern: `check()` helper for error handling, `ChronoconeTensor[any]` for tensor types
- Types are represented as `float64` for this increment
- Aggregators are applied manually in test code — not yet integrated into metric computation (Increment 3/4)

### Task 1: Add CILAggregateNode walking to CIL tree walker — Completed

Add a `WalkCILAggregateNode` handler in the CIL tree walker (`src/main/go/cil/cil_transformer.go`) that walks a `CILAggregateNode` and emits Go code calling `tensorops.ReduceDimension`. The handler must: walk the source CIL subtree (which may reference connector data), determine the collapsed dimension index in the source tensor, emit the `ReduceDimension` call with the correct variable names and aggregation matrix parameter, emit error handling after the call (`if err != nil { return ... }`), and register the result variable for use by subsequent operations. Wire the `CILAggregateNode` case into the main walker dispatch.

   a. Add CILAggregateNode handling to `TransformMetric` in `cil_transformer_metric.go` — Completed
   b. Wire `CILAggregateNode` case into `walkToPostfix` dispatch in `cil/postfix_walk.go` — Completed
   c. Write unit tests in `cil_transformer_aggregator_test.go` — Completed
   d. Write tests: verify error handling code emission — Completed
   e. Verify walker tests pass: `go test ./src/main/go/cil/... -run Aggregator -count=1` — Completed

**Files**: `src/main/go/cil/cil_transformer.go` (modify), `src/main/go/cil/postfix_walk.go` (modify, add case), `src/main/go/cil/cil_transformer_aggregator_test.go` (create)

**Intent**: The walker must transform a `CILAggregateNode` into executable Go code that calls the matrix multiplication infrastructure (`tensorops.ReduceDimension`) built in Story254. The generated code follows existing patterns: variable naming via `freshVar`, error propagation via `check()` or `if err != nil`, and tensor variable registration for downstream use.

**Acceptance criteria**:
- `CILAggregateNode` with known source tensor, dimension mappings, and aggregation matrix produces correct `tensorops.ReduceDimension` call in generated code
- Generated code includes error handling: `if err != nil { return nil, fmt.Errorf("aggregation failed: %w", err) }`
- Collapsed dimension index is correctly computed from `FromDimensions` and `ToDimension`
- Result variable is registered for use by subsequent operations

### Task 2: Implement aggregation matrix emission in code generator — Completed [Small]

Add a `GenerateAggregationMatrix` method to `CILCodeGenerator` in `src/main/go/cil/cil_code_generator.go`. The method formats a pre-built `AggregationMatrix` as a Go package-level variable declaration using `&tensorops.AggregationMatrix{...}` struct literal syntax. The generated code must include: the matrix variable name (derived from the collapsed dimension name), `Rows` and `Columns` counts, `RowLabels` and `ColLabels` slices, and `Data` as a `[][]float64` literal. Follow the existing code generator pattern (e.g., `GenerateConnectorMetadataMap`).

   a. Add `GenerateAggregationMatrix(name string, matrix *tensorops.AggregationMatrix) string` method to `CILCodeGenerator` — Completed
   b. Implement matrix data formatting: row labels, column labels, data rows — Completed
   c. Write tests in `cil_code_generator_aggregator_test.go` — Completed
   d. Verify code generator tests pass: `go test ./src/main/go/cil/... -run AggregationMatrix -count=1` — Completed

**Files**: `src/main/go/cil/cil_code_generator.go` (modify), `src/main/go/cil/cil_code_generator_aggregator_test.go` (create)

**Intent**: The code generator must produce valid Go struct literals for `tensorops.AggregationMatrix` so the synthesis preamble can declare pre-built aggregation matrices as package-level variables. Matrix values come from concept value sets and are materialized at compile time.

**Acceptance criteria**:
- `GenerateAggregationMatrix` emits valid Go `&tensorops.AggregationMatrix{...}` literals with correct row/column labels and data
- Generated code compiles (verified via test that checks output against expected format)
- RowLabels and ColLabels reflect the concept value names
- Data matrix dimensions match Rows × Columns

### Task 3: Integrate matrix emission into synthesis preamble — Completed

Modify the synthesis preamble generator in `src/main/go/synthesis/synthesize.go` to iterate over aggregator CIL trees and call `GenerateAggregationMatrix` for each, placing declarations before metric computation functions. Add the `"tensorops"` import to the generated code's import block. The integration must follow the existing synthesis pattern: build declarations in the `GeneratedProgram.Variables` slice during `GenerateMainFunction`.

   a. Iterate aggregator CIL trees and collect aggregation matrices in `GenerateMainFunction` — Completed
   b. Call `transformer.GenerateAggregationMatrix` for each aggregator, adding to preamble — Completed
   c. Add tensorops import when aggregators are present in `buildMainImports` — Completed
   d. Verify synthesized program compiles: integration test added — Completed
   e. Verify integration tests pass: `go test ./src/main/go/synthesis/... -count=1` — Completed

**Files**: `src/main/go/synthesis/synthesize.go` (modify), synthesis integrator test file (create or modify existing)

**Intent**: The synthesis pipeline must emit all aggregation matrix declarations in the preamble so they are available when metric computation functions reference them. The preamble structure follows the existing pattern: concept values map → aggregation matrices → metric computation functions.

**Acceptance criteria**:
- Generated program includes matrix variable declarations for all aggregators in the compilation unit
- Matrix declarations appear before metric computation functions in the preamble
- Generated program imports `tensorops` package when aggregators are present
- Generated program compiles successfully with `go build`

### Task 4: Create with_data aggregator example and manual test — Completed

Create an end-to-end `with_data/aggregators_basic/` example that validates the full aggregator pipeline: CPL → parse → analyze → CIL → synthesize → execute → verify. The example uses a small sum aggregation (e.g., Country={USA,Canada} × Product={widget,gadget} → sum across Country → Product={widget,gadget}). Create the CPL file with concepts, connector, and aggregator declaration. Create IO specification, parquet test data, and expected output xlsx. Write a manual test that loads parquet data, applies `tensorops.ReduceDimension` with the sum matrix, and verifies the result matches ground truth.

   a. Create `examples/with_data/aggregators_basic/` directory with standard category structure — Completed
   b. Write core.cpl with concepts (Country, Product), connector, and measure — Completed
   c. Write sum_country.cpl with metric and aggregator declaration — Completed
   d. Write IO specification JSON defining input connector and output spec — Completed
   e. Generate parquet test data with known values — Completed
   f. Manual tensorops ReduceDimension test verifies correct aggregated values — Completed
   g. Verify example tests pass: parse test, E2E compile test, ReduceDimension test — Completed

**Files**: `examples/with_data/aggregators_basic/core.cpl` (create), `examples/with_data/aggregators_basic/aggregators_basic/` subdirectory with .cpl, iospec.json, data files, .expected.xlsx (create), `examples/with_data/aggregators_basic/aggregators_basic_test.go` (create), `examples/with_data/aggregators_basic/example_definitions.go` (create)

**Intent**: This example provides the first end-to-end aggregator validation: CPL file → parse → analyze → CIL gen → synthesize → execute → verify aggregated output. The manual test pattern follows metrics_arithmetic and other existing with_data examples.

**Acceptance criteria**:
- CPL parses and analyzes without errors
- Synthesized program compiles and executes
- Manual aggregation with `tensorops.ReduceDimension` produces correct aggregated values matching ground truth
- Example test passes: parse → analyze → synthesize → execute → verify

## Notes

- Aggregators are applied manually in test code in this increment — not yet integrated into automatic metric computation (that happens in Increment 3/4)
- Matrix values from concept value sets are materialized at compile time (not loaded at runtime)
- The `with_data/aggregators_basic/` example follows the standard category structure: `core.cpl` with shared definitions, subdirectories for each example variant
- File paths in all task descriptions are relative to `source_root` (`src/chronocone_planning_language`)
