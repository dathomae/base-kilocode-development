# Story183: CPL Metrics Basic Synthesize

## Goal

Implement the Increment 1 synthesis phase for CPL metric declarations as part of [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md). This story extends the existing synthesis infrastructure to generate a complete executable program that:

1. Accepts an IO specification as input, which declares one or more metrics as outputs.
2. Generates metric computation code that produces tensors filled with literal `0`.
3. Writes each output metric specified in the IO spec to a Parquet file containing the computed tensor data, dimensions, and grain.

The connector parquet-reading code generation, Setup/Cleanup/main lifecycle functions, and IO spec infrastructure already exist in [`synthesize.go`](src/chronocone_planning_language/src/main/go/synthesis/synthesize.go) and [`cil_transformer.go`](src/chronocone_planning_language/src/main/go/cil/cil_transformer.go). This story covers:

- Extend [`CILAbstractTransformer`](src/chronocone_planning_language/src/main/go/cil/cil_transformer.go) with `VisitCILLiteral` handler — emits constant value.
- Implement `TransformMetric` (currently a placeholder) to generate `computeX()` functions that allocate output tensors and fill with literal `0`.
- Extend existing `GenerateComputeFunction()` (currently a placeholder) to launch metric computation goroutines.
- Extend existing `GenerateMainFunction()` to wire metric computations into the lifecycle.
- Add Parquet output writing support so the synthesized program writes each IO-spec output metric to a `.parquet` file instead of stdout.

## References

- [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md)
- [ERO005: Metrics Synthesis Phase — Increment 1](../specifications/design/cpl/go-synthesis-target/eros/ERO005_metrics-synthesis-phase.md)
- [ERO005: Metrics Incremental Implementation](../specifications/design/cpl/eros/ERO005_metrics-incremental-implementation.md)
- [Synthesis Design](../specifications/design/cpl/go-synthesis-target/synthesis.md)
- [Tensor Representation Design](../specifications/design/cpl/go-synthesis-target/tensor-representation.md)
- [IO Specifications Requirements](../specifications/requirements/cpl/io_specs.md)
- [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Dependencies

- [Story182_cpl-metrics-basic-parse-analyze.md](Story182_cpl-metrics-basic-parse-analyze.md) — parse/analyze phase must produce valid `AnalysisMetricDefinition` entries with `CILLiteral(0)` CIL trees (Not Started)
- [Epic004_cpl-connectors-implementation.md](../epics/Epic004_cpl-connectors-implementation.md) — connectors synthesis must be complete (connector loading codegen already exists)

## Dependent Stories

- [Story184_cpl-metrics-arithmetic-parse-analyze.md](Story184_cpl-metrics-arithmetic-parse-analyze.md) — Increment 2 Parse+Analyze (Not Started)

## Tasks

1. Extend CIL tree walker with literal visitor — Completed
   a. Define `VisitCILLiteral(ctx *CILContext, node *CILLiteral) error` method on `CILAbstractTransformer` — Completed
   b. Write unit tests for `VisitCILLiteral`: verify correct Go code emission for literal values — Completed
   c. Implement `VisitCILLiteral` — emit the literal's `Value` as a Go numeric literal in generated code output — Completed

2. Implement metric compute function generation — Completed
   a. Replace the placeholder `TransformMetric()` in [`cil_transformer.go`](src/chronocone_planning_language/src/main/go/cil/cil_transformer.go:283) with real implementation — Completed
   b. Generate `computeX()` function per metric: allocates output tensor matching dimensionality from input connectors, fills all cells with literal `0` — Completed
   c. Write unit tests for `computeX()` generation: verify output tensor allocation and fill code — Completed
   d. Implement `computeX()` generation: iterate `CILRoot` tree, visit `CILLiteral`, produce Go function body — Completed

3. Add Parquet output writer and integrate IO spec outputs into synthesized program — Completed
   a. Implement `WriteTensorToParquet` in the `io` package (or equivalent helper) that accepts a computed tensor, dimension names, grain, and file path, and writes a Parquet file — Completed
   - **Note:** [`GenerateParquetFile`](src/chronocone_planning_language/src/main/go/io/generate_parquet.go:95) already exists but is designed for generating **test input data** with random values from dimension definitions. It is not suitable for writing computed metric tensors. The new `WriteTensorToParquet` function writes an existing tensor's data to Parquet.
     b. Write unit tests for `WriteTensorToParquet`: verify correct schema, columns, and data values — Completed
     c. Replace the placeholder `generateOutputWritingCode()` in [`synthesize.go`](src/chronocone_planning_language/src/main/go/synthesis/synthesize.go:399) with real implementation that emits code to iterate `ioSpec.Outputs` and call the Parquet writer for each output metric — Completed
     d. Remove all stdout placeholder output from generated code (e.g., `fmt.Printf` statements for output metrics) — Completed
     e. Write integration tests: verify that generated code contains Parquet-writing calls driven by `ioSpec.Outputs` — Completed

4. Extend existing lifecycle functions for metric computations — Completed
   a. Extend `GenerateComputeFunction()` (currently placeholder at [`cil_transformer.go:314`](src/chronocone_planning_language/src/main/go/cil/cil_transformer.go:314)) to launch goroutines for each metric computation — Completed
   b. Extend `GenerateMainFunction()` in [`synthesize.go`](src/chronocone_planning_language/src/main/go/synthesis/synthesize.go:186) to register `MetricComputation` objects in Setup and wire metric compute calls — Completed
   c. Write integration tests: generate complete Go program from a basic metric CPL file, compile it, verify it runs and writes Parquet output files — Completed

5. End-to-end verification — Completed
   a. Create a minimal CPL file with one metric using literal `0` calculation — Completed
   b. Create corresponding IO spec specifying that metric as an output, and create test parquet input data using the existing [`GenerateParquetFile`](src/chronocone_planning_language/src/main/go/io/generate_parquet.go:95) utility (or Story161 parquet test utility if available) — Completed
   c. Run parse → analyze → synthesize → compile → execute pipeline and verify the output Parquet file contains the expected metric data (all zeros) with correct dimensions and grain — Completed
   d. Run the full CPL test suite to confirm no regressions — Completed

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. Generated code must use the [`CILAbstractTransformer`](../specifications/design/cpl/go-synthesis-target/synthesis.md#cilabstracttransformer) infrastructure for CIL tree walking
3. Generated code must follow the Setup/Compute/Cleanup lifecycle defined in the synthesis design
4. Types are initially represented as `float64`
5. Generated Go code must be compilable and runnable
6. Go functions that can fail must return `(T, error)`
7. Existing connector loading, Setup/Cleanup/main scaffolding must not be broken — extend, don't replace
8. The synthesized program must not write metric output to stdout; all output metrics must be written to Parquet files

## Intent

- **Task 1 Intent**: Add the first CIL visitor node — establishes the code generation pattern that all future CIL nodes follow
- **Task 2 Intent**: Generate the metric computation function that transforms input tensors into output tensors (currently just filling with constant `0`)
- **Task 3 Intent**: Ensure the synthesized program can write computed metric tensors to Parquet files, driven by the output metrics declared in the IO spec
- **Task 4 Intent**: Wire metric computations into the existing lifecycle scaffolding (Setup/Compute/Cleanup/main already exist)
- **Task 5 Intent**: Prove the entire pipeline works end-to-end, producing valid Parquet files as output

## Acceptance Criteria

1. `VisitCILLiteral` correctly emits constant value in generated Go code
2. `TransformMetric` generates `computeX()` functions that allocate correctly-dimensioned output tensors and fill with literal `0`
3. `GenerateComputeFunction()` launches metric computation goroutines
4. `GenerateMainFunction()` wires metric computations into the lifecycle
5. The synthesized program accepts an IO spec as input, reads its `outputs` array, and writes one Parquet file per output metric
6. Each output Parquet file contains the computed metric tensor data with correct dimension columns, grain columns (if specified), and a `Value` column of type `float64`
7. No stdout placeholder output remains in generated code for metric writing
8. Full end-to-end test: CPL → generated Go → compiled executable → runs and produces valid Parquet output files containing expected zero-filled metric data
9. All unit tests pass following Go testing best practices
10. No regressions in existing CPL test suite
11. Existing connector loading and lifecycle scaffolding remain functional
