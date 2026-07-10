# Story 106: CPL Parser - Metrics

## Overview

Extend the CPL parser to support `metric` declarations. Metrics define calculated time series with inputs and formulas. This story is limited to metrics whose input dimensions match the output dimensions; aggregators and spreaders will be added in future stories.

## References

- [CPL Language Examples](../specifications/requirements/metrics-definition-examples-2a.md) - See "Metrics" section with calculation examples
- [Story105_cpl-evaluator-cli.md](finished-stories/Story105_cpl-evaluator-cli.md) - Evaluator foundation

## Dependencies

### Prerequisites

- [Story105_cpl-evaluator-cli.md](finished-stories/Story105_cpl-evaluator-cli.md) - Not Started (requires evaluator infrastructure)

### Dependents

- [Story107_cpl-metric-data-structure-design.md](Story107_cpl-metric-data-structure-design.md) - Not Started

## Tasks

1. Extend grammar for metric declarations - Not Started
   a. Update grammar with `metric` keyword and assignment syntax
   b. Define parser rules for metric headers (name, dimensions, grain)
   c. Define parser rules for `inputs` block referencing other metrics
   d. Define parser rules for `calculate` block with expressions
   e. Define parser rules for `where` clauses (ACTUAL vs FORECAST filtering)
   f. Generate updated Go parser code
   g. Write unit tests for metric declaration parsing

2. Extend grammar for expressions - Not Started
   a. Define parser rules for arithmetic expressions (+, -, \*, /)
   b. Define parser rules for comparison operators (==, !=, <, >, <=, >=)
   c. Define parser rules for logical operators (and, or, not)
   d. Support parenthesized expressions
   e. Generate updated Go parser code
   f. Write unit tests for expression parsing

3. Extend grammar for dimension indexing - Not Started
   a. Define parser rules for dimension indexing (e.g., `metric[dimension]`)
   b. Generate updated Go parser code
   c. Write unit tests for dimension indexing parsing
   d. Note: `lag` and `lead` time-series functions will be added in future stories

4. Extend grammar for conditional expressions - Not Started
   a. Define parser rules for `if/then/else` expressions
   b. Support boolean literals and conditions
   c. Generate updated Go parser code
   d. Write unit tests for conditional parsing
   e. Note: Aggregator declarations are out of scope; will be added in future stories

5. Create AST types for metrics - Not Started
   a. Define `MetricDecl` struct with inputs and calculation
   b. Define `InputRef` struct for metric input references
   c. Define `Expression` interface and concrete types
   d. Create `ast/metrics.go` file
   e. Write unit tests for AST type construction
   f. Note: Aggregator AST types will be added in future stories

6. Implement listener methods for metrics - Not Started
   a. Implement methods for metric declaration entry/exit
   b. Implement methods for inputs block parsing
   c. Implement methods for calculate block and expressions
   d. Write unit tests for listener functionality
   e. Note: Aggregator listener methods will be added in future stories

7. Create test CPL files - Not Started
   a. Create `examples/simple_metric.cpl` with basic calculation
   b. Create `examples/metric_with_conditional.cpl` using if/then/else
   c. Create `examples/metric_matching_dimensions.cpl` with matching input/output dimensions
   d. Write integration tests parsing each example file
   e. Note: Time-series functions (`lag`, `lead`) and aggregators will be added in future stories

## Decisions

1. **Expression tree**: The AST uses an `Expression` interface with concrete types for each expression kind (BinaryExpr, UnaryExpr, FunctionCall, ConditionalExpr, etc.) to support evaluation in Story108.

2. **Metric self-reference**: Metrics may reference themselves in calculations (e.g., `inventory = lag(inventory,1) + arrivals - demand`), which forms a time-series dependency rather than a circular reference.

3. **Dimension matching**: This story is limited to metrics where input dimensions match output dimensions. Aggregators and spreaders for dimensional transformation will be added in future stories.

## Notes

- The `where ACTUAL` clause filters input metrics to only include actual (non-forecast) data points.
- Dimension indexing (e.g., `metric[dimension]`) retrieves a specific dimension value from a metric with multiple dimensions.
- The grammar supports complex expressions combining arithmetic, time-series functions, and conditionals.
- Aggregator and spreader declarations for dimensional transformation are out of scope for this story; they will be added in future stories.
