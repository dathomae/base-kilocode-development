# Story161: Parquet Test Utility

## Goal

Create a test utility function and CLI to generate parquet files with random cell values based on dimension definitions. This utility will be used in the testing framework to generate mock data for connectors. This story is part of [Epic004: CPL Connectors Implementation](../epics/Epic004_cpl-connectors-implementation.md).

## References

- [ERO004: Connectors Synthesis Phase](../specifications/design/cpl/go-synthesis-target/eros/ERO004_connectors-synthesis-phase.md)
- [Epic004: CPL Connectors Implementation](../epics/Epic004_cpl-connectors-implementation.md)

## Dependencies

None

## Dependent Stories

None

## Tasks

1. Create the parquet generation utility function - Completed
   a. Write tests for the utility function that verify it creates a valid parquet file with the correct schema and row count - Completed
   b. Define the function signature: `GenerateParquetFile(filePath string, dimensions map[string][]string) error` - Completed
   c. Implement the function using `go-arrow` to generate all combinations of dimension values and assign a random float64 to the cell value - Completed
   d. Ensure the function writes the generated Arrow table to the specified parquet file path - Completed

2. Create the CLI tool for generating parquet files - Completed
   a. Write tests for the CLI command parsing and execution logic - Completed
   b. Define the CLI command structure and arguments (e.g., input JSON file path, output parquet file path) - Completed
   c. Implement the CLI command to read the JSON file containing dimension definitions, parse it into the `map[string][]string` format, and call `GenerateParquetFile` - Completed
   d. Add error handling for invalid JSON input or file writing issues - Completed

## Test-First Development

This project follows the Logical TDD Lifecycle. For coding tasks, write tests before implementing the code. Include the test-and-implement cycle within each task in the correct order. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

- The utility must use the `github.com/apache/arrow/go/v16/parquet` library.
- The generated parquet file must have a schema where each dimension is a string column and the cell value is a float64 column named "Value".
- The generated data must include every combination of the provided dimension values (Cartesian product).

## Intent

- **Task 1**: To provide a core programmatic utility for generating mock parquet data based on specified dimensions.
- **Task 2**: To provide a command-line interface for the utility, making it easy to generate mock data files for testing purposes.

## Acceptance Criteria

- The `GenerateParquetFile` function successfully creates a parquet file with the expected schema and data (Cartesian product of dimensions + random float64 value).
- The CLI tool can read a JSON file defining dimensions and output a valid parquet file.
- All tests for the utility function and CLI tool pass.
