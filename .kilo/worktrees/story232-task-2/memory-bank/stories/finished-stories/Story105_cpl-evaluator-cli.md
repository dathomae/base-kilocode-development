# Story 105: CPL Evaluator and CLI

> **OBSOLETE / SUPERSEDED**: This story has been superseded by the new synthesis architecture. The CPL compiler now synthesizes Go code instead of evaluating it directly.

## Overview

Create a CPL evaluator package that processes parsed CPL documents and executes connectors against data sources. Include a companion CLI program that accepts a `.cpl` file and a data source directory. The evaluator reads JSON test data files when it encounters connectors with `driver=test_data` and
a `path` property that specifies the relative path to the JSON file within the data source directory.

## References

- [CPL Language Examples](../specifications/requirements/metrics-definition-examples-2a.md) - Connector and data examples
- [Story104_cpl-parser-connectors.md](finished-stories/Story104_cpl-parser-connectors.md) - Connector parsing

## Dependencies

### Prerequisites

- [Story104_cpl-parser-connectors.md](finished-stories/Story104_cpl-parser-connectors.md) - Not Started (requires connector AST types)

### Dependents

- [Story106_cpl-parser-metrics.md](Story106_cpl-parser-metrics.md) - Not Started

## Tasks

1. Create evaluator package structure - Not Started
   a. Create `evaluator/` directory for evaluator package
   b. Create `evaluator/evaluator.go` with main `Evaluator` struct
   c. Create `evaluator/datasource.go` for data source abstractions
   d. Create `evaluator/testdriver.go` for test data driver implementation
   e. Initialize Go module for evaluator package
   f. Write unit tests for package initialization

2. Implement test data driver - Not Started
   a. Define `TestDataDriver` struct implementing data source interface
   b. Implement JSON file reading from data source directory
   c. Support relative path resolution from connector `path` property
   d. Implement dimension-based data lookup
   e. Write unit tests for test driver functionality

3. Implement connector execution - Not Started
   a. Create `ExecuteConnector` method on `Evaluator`
   b. Dispatch to appropriate driver based on connector `driver` property
   c. Handle `test_data` driver using `TestDataDriver`
   d. Implement error handling for missing dimensions
   e. Write unit tests for connector execution

4. Implement data source directory handling - Not Started
   a. Create `DataSource` type representing the data source directory
   b. Implement file discovery and loading
   c. Support JSON array and object formats
   d. Map JSON data to dimension keys
   e. Write unit tests for data source handling

5. Create CLI program - Not Started
   a. Create `cmd/cpl-eval/` directory for CLI
   b. Create `cmd/cpl-eval/main.go` with command-line argument parsing
   c. Accept `.cpl` file path as first argument
   d. Accept data source directory path as second argument
   e. Integrate parser and evaluator
   f. Write tests for CLI argument handling

6. Implement CLI output formatting - Not Started
   a. Add formatted output for parsed CPL document summary
   b. Display loaded connectors and their data sources
   c. Add verbose mode for debugging
   d. Write tests for output formatting

7. Create test data files - Not Started
   a. Create `testdata/` directory with sample JSON files
   b. Create `testdata/products.json` with product dimension data
   c. Create `testdata/countries.json` with country dimension data
   d. Create example `.cpl` file using `driver=test_data`
   e. Write integration test using test data

8. Add error handling and validation - Not Started
   a. Validate data source directory exists and is readable
   b. Validate JSON files match expected schema
   c. Report meaningful errors for missing or malformed data
   d. Write tests for error conditions

## Decisions

1. **Package separation**: The evaluator is in its own `evaluator/` package separate from the CLI program in `cmd/cpl-eval/`. This allows the evaluator to be imported by other tools without pulling in CLI dependencies.

2. **Driver interface**: A `DataSourceDriver` interface abstracts different connector drivers, with `TestDataDriver` as the first implementation.

3. **Test data format**: JSON files in the data source directory contain arrays of objects where object properties map to dimensions.

4. **CLI design**: The CLI is minimal, accepting only the CPL file path and data source directory as arguments. Additional options (like verbose mode) use flags.

## Notes

- When `driver=test_data`, the connector's `path` property is treated as a relative path from the data source directory.
- The evaluator does not yet compute metrics (that comes in Story108) but does load connector data.
- JSON files should contain dimension keys that match the connector's declared dimensions for proper lookup.
- Error handling follows the `on missing` clauses defined in connectors, falling back to error messages or fetch operations.
- **Test Data JSON Structure**: The supported JSON structure is a multilevel hash table (nested object structure) with the following constraints:
  1. **Top-level metadata fields**: The JSON must include these fields at the top level:
     - `time_grain`: Either `"day"` or `"week"`
     - `forecast_start_date`: The date when forecast data begins
     - `actuals_length`: Number of periods in the actuals data
     - `forecast_length`: Number of periods in the forecast data
  2. **Leaf-level structure**: The leaf nodes must be objects containing `actuals` and/or `forecast` arrays (either may be missing). The length of these arrays must match the `actuals_length` and `forecast_length` specifications respectively.

  For example:

  ```json
  {
    "time_grain": "week",
    "forecast_start_date": "2024-03-01",
    "actuals_length": 4,
    "forecast_length": 8,
    "data": {
      "product_A": {
        "region_1": {
          "actuals": [100, 200, 150, 175],
          "forecast": [180, 190, 200, 210, 220, 230, 240, 250]
        },
        "region_2": {
          "actuals": [80, 120, 110, 130],
          "forecast": [140, 150, 160, 170, 180, 190, 200, 210]
        }
      },
      "product_B": {
        "region_1": {
          "actuals": [300, 310, 320, 330],
          "forecast": [340, 350, 360, 370, 380, 390, 400, 410]
        },
        "region_2": {
          "forecast": [90, 110, 130, 140, 150, 160, 170, 180]
        }
      }
    }
  }
  ```

  In this structure, the dimension data is nested under the `data` key, `product_A` and `product_B` are top-level dimension keys within `data`, `region_1` and `region_2` are nested dimension keys, and the leaf nodes contain `actuals` and `forecast` arrays. Note that `product_B.region_2` has only `forecast` data (no `actuals`), which is valid.
