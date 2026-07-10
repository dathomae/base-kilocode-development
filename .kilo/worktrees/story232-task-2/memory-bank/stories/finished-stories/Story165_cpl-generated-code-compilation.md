# Story165: CPL Generated Code Compilation

## Goal

Make CPL compiler-generated Go programs compilable and runnable as standalone executables. The generated code must import the Chronocone runtime libraries, read parquet data via connectors into `ChronoconeTensor` instances, and execute successfully end-to-end. This story builds on the existing synthesizer connector infrastructure (Story156 and Story157) by closing the gap between code generation and execution.

## References

- [CPL Generated Code Compilation Plan](../../cpl_generated_code_compilation_plan.md)
- [CPL Synthesis Design](../../specifications/design/cpl/go-synthesis-target/synthesis.md)
- [CPL Connectors Requirements](../../specifications/requirements/cpl/connectors.md)
- [CPL README](../../../src/chronocone_planning_language/README.md)

## Dependencies

- [Story156_cpl-analyzer-connectors.md](../Story156_cpl-analyzer-connectors.md) - Done
- [Story157_cpl-synthesizer-connectors.md](../Story157_cpl-synthesizer-connectors.md) - Not Started (basic synthesizer connector generation exists but does not yet compile or read data)
- [Story161_parquet-test-utility.md](Story161_parquet-test-utility.md) - Done

## Dependent Stories

None

## Tasks

1. Add runtime parquet-to-tensor helper in the `io` package - Done
   a. Define `ReadParquetToTensor` function signature in `src/main/go/io/parquet_reader.go` (Phase A: Structural Definition) - Done
   b. Write tests for `ReadParquetToTensor` using a test parquet file with known dimensions and values (Phase B: Logical Tests) - Done
   c. Implement `ReadParquetToTensor` using `github.com/apache/arrow/go/v16/parquet/file` to read columns, map dimension values to concept positions, and populate a `ChronoconeTensor[float64]` (Phase C: Implementation) - Done
   d. Verify tests pass and the helper correctly handles dimension column ordering vs connector dimension ordering - Done

2. Update synthesizer to generate proper imports and package-qualified type references - Done
   a. Update `generateMainFunctionSignature()` in `synthesize.go` to emit imports for `calculation`, `connector`, and `cplio` packages (Phase A) - Done
   b. Write synthesis tests verifying generated code contains correct import block and no unqualified type references (Phase B) - Done
   c. Update all synthesizer string builders (`GenerateMainFunction`, `generateInputLoadingCode`, `generateOutputWritingCode`, `GenerateConnectorMetadataMap`, `GenerateParquetLoaderLibrary`, `GenerateConnectorLoaders`, `GenerateSingleConnectorLoader`, `GenerateLoadParquetConnectorFunction`) to use fully-qualified type names (Phase C) - Done
   d. Verify generated code compiles when placed in `generated/cmd/<name>/main.go` - Done

3. Generate driver parameters and concept values in synthesized code - Done
   a. Update `generateInputLoadingCode()` to emit `DriverParameters` slice literals from `ioSpec.Inputs[i].DriverParameters` instead of TODO comments (Phase A) - Done
   b. Add `BuildConceptValuesMap()` to extract `ResolvedValues` from `AnalysisDefinitions.Concepts` and generate a `conceptValues` map literal in synthesized output (Phase A) - Done
   c. Write synthesis tests verifying driver parameters and concept values appear correctly in generated code (Phase B) - Done
   d. Implement the parameter and concept value generation, ensuring the `conceptValues` map is passed into connector loader functions (Phase C) - Done
   e. Verify the `random_data_no_time` example generates code with correct `path: "testdata.parquet"` and concept values for Year and Region - Done

4. Update connector loader generation to use the runtime helper - Done
   a. Modify `GenerateLoadParquetConnectorFunction()` in `connector/connector_loader.go` to generate a thin wrapper that extracts the file path and delegates to `cplio.ReadParquetToTensor` (Phase A) - Done
   b. Write tests verifying the generated loader function calls the runtime helper with correct arguments (Phase B) - Done
   c. Update `GenerateSingleConnectorLoader()` to accept and forward the `conceptValues` map (Phase C) - Done
   d. Verify the generated loader code compiles and the integration test with `random_data_no_time` reads parquet data successfully - Done

5. Create `compile_and_run.sh` shell script - Done
   a. Create `src/chronocone_planning_language/shell_scripts/compile_and_run.sh` that resolves its own directory, cd's to the module root, and runs `go run "$1"` (Phase A/C) - Done
   b. Make the script executable and verify it works with a simple `go run` invocation - Done

6. Update project documentation for new directories - Done
   a. Update `src/chronocone_planning_language/toc.md` to list `generated/cmd/` and `shell_scripts/` directories - Done
   b. Update `src/chronocone_planning_language/README.md` examples to show the `generated/cmd/<name>/main.go` output convention - Done

## Constraints

- Generated code must remain `package main` and be directly compilable via `go run`.
- Parquet reading must use the existing `github.com/apache/arrow/go/v16` dependency already declared in `go.mod`.
- The synthesizer must not generate inline parquet parsing logic; all parquet I/O must delegate to the runtime `io` package helper.
- Concept values must be embedded in generated code as `map[string][]string` literals so runtime tensor shape computation and row-to-index mapping work without reading concept metadata from external files.
- All changes must follow Go naming conventions and the project's co-located test file organization.

## Intent

The overall goal is to close the gap between CPL code generation and executable output. Currently, the synthesizer produces Go files that reference runtime types but do not import them, and contain stub parquet reading logic that never actually parses data. This story makes the generated programs first-class executables that can be compiled, run, and verified against real parquet test data.

## Acceptance Criteria

- Running `go run ./src/main/go/cli/compiler -o generated/cmd/random_data_no_time/main.go -io examples/with_data/random_data_no_time/iospec.json examples/with_data/random_data_no_time/test.cpl` produces a file at that path.
- Running `./shell_scripts/compile_and_run.sh generated/cmd/random_data_no_time/main.go` compiles and executes without errors.
- The executed program loads the `testdata.parquet` file and constructs a `ChronoconeTensor[float64]` with shape `[2, 4]` (Year × Region) and 8 data elements.
- All new code has corresponding tests that pass with `./gradlew test`.
- `toc.md` and `README.md` accurately reflect the new directory structure.

## Open Questions

- Should the `generated/cmd/` directory be added to `.gitignore` since generated code is ephemeral, or should example outputs be committed for documentation purposes?
- Should the shell script support additional arguments (e.g., passing command-line flags to the generated program)?
