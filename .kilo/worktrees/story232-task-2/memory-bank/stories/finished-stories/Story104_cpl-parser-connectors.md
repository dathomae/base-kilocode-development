# Story 104: CPL Parser - Connectors

> **OBSOLETE / SUPERSEDED**: This story has been superseded by [Story162_cpl-parser-connectors-basic.md](../Story162_cpl-parser-connectors-basic.md) which aligns with the ERO004 design for connectors.

## Overview

Extend the CPL parser to support `connector` declarations. Connectors define how data is imported from external sources, including metric connectors (for time series data) and lookup connectors (for reference data).

## References

- [CPL Language Examples](../specifications/requirements/metrics-definition-examples-2a.md) - See "Importers" and connector examples
- [Story103_cpl-parser-units-and-dimensions.md](Story103_cpl-parser-units-and-dimensions.md) - Units and dimensions parsing

## Dependencies

### Prerequisites

- [Story103_cpl-parser-units-and-dimensions.md](Story103_cpl-parser-units-and-dimensions.md) - Not Started (requires units and dimensions AST types)

### Dependents

- [Story105_cpl-evaluator-cli.md](Story105_cpl-evaluator-cli.md) - Not Started

## Tasks

1. Extend grammar for connector declarations - Not Started
   a. Update grammar with `connector` keyword and block structure
   b. Define parser rules for metric connectors (`metric` sub-block)
   c. Define parser rules for connector properties (driver, secret_path, server, path, params)
   d. Define parser rules for trigger and grain specifications
   e. Define parser rules for horizon specifications
   f. Generate updated Go parser code
   g. Write unit tests for connector parsing
   h. Note: Lookup connectors will be added in a future lookup story

2. Extend grammar for error handling clauses - Not Started
   a. Define parser rules for `on missing <dimension>` clauses
   b. Support error message syntax: `error "message ${variable}"`
   c. Generate updated Go parser code
   d. Write unit tests for error handling parsing
   e. Note: `fetch from` fallback syntax is out of scope; will be added in lookup stories

3. Extend grammar for parameter mappings - Not Started
   a. Define parser rules for `params` block
   b. Support static parameter values and variable references (e.g., `$fetch_time`)
   c. Support dimension-to-parameter mappings (e.g., `country=country`)
   d. Generate updated Go parser code
   e. Write unit tests for parameter mapping parsing

4. Create AST types for connectors - Not Started
   a. Define `ConnectorDecl` struct with metric connector fields
   b. Define `OnMissingClause` struct for error handling
   c. Define `ParamMapping` and related types
   d. Create `ast/connectors.go` file
   e. Write unit tests for AST type construction
   f. Note: Lookup connector types will be added in future lookup story

5. Implement listener methods for connectors - Not Started
   a. Implement `EnterConnectorDecl` and `ExitConnectorDecl` methods
   b. Implement methods for metric sub-blocks
   c. Implement methods for property parsing
   d. Implement methods for error handling clauses
   e. Implement methods for parameter mappings
   f. Write unit tests for listener functionality

6. Create test CPL files - Not Started
   a. Create `examples/metric_connector.cpl` with metric connector example
   b. Create `examples/connector_with_errors.cpl` showing error handling
   c. Create `examples/multiple_connectors.cpl` with several connectors
   d. Write integration tests parsing each example file
   e. Note: Lookup connector examples will be added in future lookup story

## Decisions

1. **Connector scope**: This story implements metric connectors only. Lookup connectors will be added in a future story focused on lookup functionality.

2. **Variable interpolation**: Error messages support `${variable}` syntax for interpolating dimension values into error strings.

## Notes

- Connectors specify how to fetch data from external systems via drivers like `https_api`.
- The `on missing` clauses provide error handling when referenced dimensions are not found.
- Connector declarations may include `trigger` (fetch schedule), `grain` (data frequency), and `horizon` (actuals vs forecast).
- The `fetch from` fallback syntax is out of scope for this story; it will be added when lookup support is implemented.
