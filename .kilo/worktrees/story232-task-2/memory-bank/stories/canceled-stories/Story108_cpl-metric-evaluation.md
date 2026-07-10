# Story 108: CPL Metric Evaluation

## Overview

Implement metric evaluation in the CPL evaluator for metrics that contain only formulas and boolean logic (without requiring aggregators, spreaders, or lookups). Augment the CLI to print evaluated metric data after processing.

## References

- [CPL Language Examples](../specifications/requirements/metrics-definition-examples-2a.md) - Metric calculation examples
- [Story106_cpl-parser-metrics.md](Story106_cpl-parser-metrics.md) - Metric AST types and expressions
- [Story107_cpl-metric-data-structure-design.md](Story107_cpl-metric-data-structure-design.md) - Metric data structure design

## Dependencies

### Prerequisites

- [Story106_cpl-parser-metrics.md](Story106_cpl-parser-metrics.md) - Not Started (requires metric AST types)
- [Story107_cpl-metric-data-structure-design.md](Story107_cpl-metric-data-structure-design.md) - Not Started (requires metric data structure design)

### Dependents

None yet. Future stories will add aggregator, spreader, and lookup support to the evaluator.

## Tasks

1. Implement metric data structures - Not Started
   a. Create `evaluator/metricdata.go` with metric data types per Story107 design
   b. Implement `MetricData` struct for in-memory metric representation
   c. Implement dimension indexing and lookup
   d. Implement time series data storage
   e. Write unit tests for data structure operations

2. Implement expression evaluator - Not Started
   a. Create `evaluator/expressions.go` for expression evaluation
   b. Implement arithmetic operation evaluation (+, -, \*, /)
   c. Implement comparison operator evaluation
   d. Implement logical operator evaluation (and, or, not)
   e. Write unit tests for expression evaluation

3. Implement dimension indexing resolution - Not Started
   a. Implement dimension indexing for metric references (e.g., `metric[dimension]`)
   b. Write unit tests for dimension indexing

4. Implement conditional expression evaluation - Not Started
   a. Implement `if/then/else` expression evaluation
   b. Implement boolean condition evaluation
   c. Handle type coercion in conditional branches
   d. Write unit tests for conditional expressions

5. Implement metric calculation engine - Not Started
   a. Create `evaluator/calculator.go` with calculation logic
   b. Implement input metric resolution from connector data
   c. Implement calculation execution for metric formulas
   e. Write unit tests for calculation engine

6. Implement metric value access - Not Started
   a. Add `GetMetric(name string) (*MetricData, error)` method to Evaluator
   b. Add `GetValue(metric, dimensions, timestamp)` method
   c. Add iteration methods for traversing metric data
   d. Write unit tests for value access methods

7. Augment CLI with metric output - Not Started
   a. Add `--show-metrics` flag to CLI
   b. Implement formatted output for evaluated metrics
   c. Display metric names, dimensions, and sample values
   d. Support JSON output format option
   e. Write tests for CLI output formatting

8. Create test CPL files with calculations - Not Started
   a. Create `testdata/revenue_calculation.cpl` with simple multiplication
   b. Create `testdata/conditional_metric.cpl` using if/then/else
   c. Create corresponding JSON data files
   d. Write integration tests evaluating each metric
   e. Note: Time-series function tests (`lag`, `lead`) will be added in future stories

9. Add calculation validation - Not Started
   a. Validate dimension compatibility in calculations
   b. Detect circular dependencies in metric references
   c. Report meaningful errors for calculation failures
   d. Write tests for validation and error handling

## Decisions

1. **Scope limitation**: This story evaluates only metrics with formulas and boolean logic. Aggregators, spreaders, and lookups are explicitly out of scope and will be handled in future stories.

2. **Metric DAG**: Metrics form a directed acyclic graph based on input dependencies. The calculator resolves these by processing in dependency order. Time-series self-references (using `lag` and `lead`) will be handled in future stories.

3. **Type system**: Numeric values are represented as `float64` for calculation simplicity, with optional preservation of integer types where specified.

4. **CLI output**: The CLI displays evaluated metric data in a human-readable table format by default, with JSON output available for programmatic consumption.

## Notes

- Metrics with `where ACTUAL` clauses filter their inputs to only actual (observed) data points before calculation.
- The calculation engine processes metrics in dependency order, ensuring inputs are calculated before dependent metrics.
- Dimension mismatches in calculations (e.g., trying to add metrics with incompatible dimensions) produce clear error messages.
- The `lag` and `lead` time-series functions are out of scope for this story; they will be added in future stories.
