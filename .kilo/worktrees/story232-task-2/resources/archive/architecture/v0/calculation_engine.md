# Calculation Engine

## Purpose

The Calculation Engine is the computational heart of ChronoCone. It executes the transformation logic that converts input metrics into output metrics, coordinating transformers, shifters, and filters to produce planning results. The engine operates in two primary phases: a publish phase that validates and prepares calculation logic, and an update phase that executes calculations when an update is triggered. Additionally, a development mode supports planners during function creation and modification.

This design reflects a fundamental principle of the system: validation and preparation should happen once, at publish time, so that execution time can be as efficient as possible. By doing the expensive work of analyzing function signatures, validating dimension compatibility, and constructing the execution graph during publishing, the engine can execute calculations quickly and deterministically when data arrives. The development mode extends this architecture to support iterative development, allowing planners to validate their work before publishing.

## Responsibilities

The Calculation Engine has five primary responsibilities that together enable the transformation of planning data.

**Validating and preparing calculation logic at publish time** involves analyzing transformer, shifter, and filter functions to ensure they are correctly defined and compatible with the metrics they reference. This validation catches errors early, before calculations are needed, providing clear feedback to planners about what needs to be fixed.

**Constructing the execution DAG** means building a directed acyclic graph that captures the sequence of operations needed to calculate all metrics in a namespace. This DAG represents the compiled form of the planning logic, optimized for efficient execution.

**Executing calculations in a sandboxed environment** ensures that user-defined Python code runs in isolation, protecting the system from malicious or buggy code while providing a consistent execution environment regardless of where the engine itself is running.

**Managing execution context and data flow** involves tracking intermediate results, passing data between operations, and potentially optimizing operations such as batching connector calls for efficiency.

**Supporting iterative development** through development mode allows planners to test their function code against local workspace artifacts, receiving immediate feedback on calculation results and test case outcomes before publishing to production.

---

## Function Definition and Storage

### Python Function Representation

Transformers, shifters, and filters are captured as Python functions stored in the object store. The function names and parameter names correspond directly to metric names, creating a natural mapping between the code and the data it operates on.

**Function Naming Convention**: A transformer function is named after the metric it produces. For example, a transformer that calculates the `labor_hours` metric would be defined as:

```python
def labor_hours(demand_units, productivity_rate):
    return demand_units / productivity_rate
```

The function name `labor_hours` identifies the output metric, while the parameter names `demand_units` and `productivity_rate` identify the input metrics. This convention eliminates the need for separate configuration files to specify input-output relationships—the relationships are embedded directly in the code structure.

### Unpublished Artifacts

When functions are first created or updated, they are pushed to the object store as unpublished artifacts. The `is_dirty` flag in the database tracks this state. Unpublished functions are stored and versioned but are not used in production calculations. This allows planners to iterate on their code, test changes, and validate results before making the changes live.

The unpublished state serves several purposes. It provides a checkpoint where code can be saved and shared without affecting production. It enables review workflows where code can be examined before publication. And it maintains a clear separation between development work and production calculations.

---

## Publish Phase

The publish phase validates and prepares calculation logic for execution. This phase happens when a planner explicitly publishes their changes, transitioning functions from unpublished to published state. The publish phase performs several critical operations that ensure the calculation logic is correct and can be executed efficiently.

### AST Analysis

The Calculation Engine uses Python's `ast` (Abstract Syntax Tree) module to analyze transformer, shifter, and filter functions. This analysis extracts structural information about each function without executing it.

**Function Signature Extraction**: The AST analysis examines the function definition to extract:

- The function name, which identifies the output metric
- The parameter names, which identify the input metrics
- The parameter order, which determines how inputs are passed to the function

This extraction happens without executing the code, making it safe to analyze potentially buggy or incomplete functions. The AST approach also works regardless of the code's internal logic—the engine only needs to understand the interface, not the implementation.

**Example AST Analysis**:

```python
import ast

# The transformer function
code = """
def labor_hours(demand_units, productivity_rate):
    # Calculate labor hours from demand and productivity
    return demand_units / productivity_rate
"""

# Parse and extract signature
tree = ast.parse(code)
function_def = tree.body[0]

function_name = function_def.name  # "labor_hours"
parameter_names = [arg.arg for arg in function_def.args.args]  # ["demand_units", "productivity_rate"]
```

### Name-to-Metric Mapping

Once the function signature is extracted, the engine maps the function name and parameter names to actual metrics in the database. This mapping validates that:

1. The output metric (from the function name) exists in the namespace
2. Each input metric (from parameter names) exists in the namespace
3. The function is not already defined by another transformer (uniqueness constraint)

If any of these validations fail, the publish operation fails with a clear error message indicating which metric was not found or which constraint was violated.

### Dimension Retrieval and Matching

For each metric referenced by the function (both inputs and output), the engine retrieves the metric's dimensions from the database. Dimensions define how a metric is indexed—for example, a sales metric might be dimensioned by region and product category.

**Dimension Compatibility**: A transformer can only directly process inputs that have the same dimensional structure as its output. When the dimensions match, the engine can call the transformer with corresponding cells from each input metric—cells that share the same dimension values and time period.

**Dimension Mismatch Example**:

Consider a transformer that calculates `regional_labor` from `national_demand`:

- `regional_labor` has dimensions: `[region]`
- `national_demand` has dimensions: `[]` (no dimensions, a single national total)

The dimensions do not match. The engine must find a way to transform `national_demand` into a form compatible with `regional_labor`.

### Shifter Discovery

When input and output dimensions do not match, the engine searches for shifters that can bridge the gap. A shifter transforms a metric with one dimensional structure into a metric with a different dimensional structure.

**Single Shifter Match**: If a shifter exists that directly transforms the input metric's dimensions to match the output metric's dimensions, the engine uses that shifter.

**Shifter Chain**: When no single shifter can bridge the gap, the engine searches for a sequence of shifters that, applied in succession, transform the input dimensions to match the output dimensions. This is essentially a path-finding problem over the space of dimensional structures.

**Shifter Discovery Algorithm**:

1. Start with the input metric's dimensional structure
2. Find all shifters that can be applied to metrics with this structure
3. For each applicable shifter, compute the resulting dimensional structure
4. If any result matches the output structure, use that shifter (or shifter chain)
5. If not, recursively search from each new dimensional structure
6. If no path is found after exhaustive search, the publish fails

**Failure Handling**: If no suitable shifter or shifter chain is found, the publish operation fails with an explanatory error message. The message indicates which input metric has incompatible dimensions and suggests what kind of shifter might be needed to resolve the issue.

### DAG Construction

The output of the publish phase is a Directed Acyclic Graph (DAG) that captures the complete evaluation sequence for all metrics in the namespace. This DAG represents the compiled form of the planning logic.

**DAG Structure**: Each node in the DAG represents an operation—a transformer, shifter, or filter. Each edge represents a data dependency—an output from one operation that serves as input to another. The DAG is acyclic because a metric cannot ultimately depend on itself; such a cycle would make calculation impossible.

**DAG Properties**:

- **Completeness**: The DAG includes all operations needed to calculate every published metric in the namespace
- **Correctness**: The DAG respects all data dependencies—an operation only executes after all its inputs are available
- **Minimality**: The DAG includes only the operations needed for the published metrics, not all possible operations

**DAG Storage**: The DAG is stored in the database, associated with the namespace family and namespace. When an update is triggered, the engine retrieves the DAG and executes it. This separation of construction and execution means the expensive analysis work happens once, at publish time, rather than repeatedly at execution time.

---

## Update Phase

The update phase executes calculations when new data arrives. This phase walks the DAG constructed during publishing, executing each operation in the correct order and passing results between operations.

### DAG Traversal

The engine executes the DAG by walking it in topological order—an order where each node is visited only after all its predecessors (inputs) have been processed. This ensures that when an operation executes, all its input data is available.

**Topological Sort**: The engine computes a topological ordering of the DAG at the start of each update. This ordering is deterministic given the same DAG, ensuring reproducible calculations.

**Execution Sequence**:

1. Retrieve the DAG for the namespace
2. Compute topological order
3. For each node in order:
   - Load input data from context or storage
   - Execute the operation (transformer, shifter, or filter)
   - Store the output in the context
4. Write final results to storage

### Single-Threaded Execution (v0)

In the v0 version of ChronoCone, the update executes in a single thread. This design prioritizes simplicity and correctness over performance. Single-threaded execution eliminates concerns about race conditions, deadlocks, and the complexity of coordinating parallel operations.

**Future Execution Modes**: Later versions will support additional execution modes:

- **Multi-threaded execution**: For CPU-bound operations on a single machine
- **Distributed execution**: For large-scale calculations across multiple machines
- **GPU acceleration**: For operations that benefit from parallel processing

The DAG structure supports these future modes because it explicitly captures all dependencies, making it possible to identify which operations can execute in parallel.

### Sandboxed Execution Environment

To isolate the code being executed, the Calculation Engine assembles a Docker container that includes:

- The transformer, shifter, and filter code
- All required supporting packages (numpy, pandas, etc.)
- The calculation engine itself
- Any connector implementations needed for external data access

**Container Benefits**:

- **Isolation**: User-defined code cannot access the host system or other containers
- **Consistency**: The execution environment is identical regardless of where the engine runs
- **Reproducibility**: The same container produces the same results given the same inputs
- **Security**: Malicious or buggy code is contained and cannot affect the broader system

**Container Lifecycle**:

1. When an update is triggered, the engine creates a new container
2. The container is populated with the current code and configuration
3. The DAG is executed within the container
4. Results are extracted and the container is terminated

This lifecycle ensures that each update runs in a fresh, clean environment.

### Execution Context

As the engine executes the DAG, it maintains a context that tracks intermediate results and supports data flow between operations.

**Context Structure**: The context is a dictionary-like structure that maps metric names to their current values. As each operation completes, its output is added to the context. Subsequent operations retrieve their inputs from the context rather than reading from storage.

**Context Benefits**:

- **Efficiency**: Intermediate results are kept in memory, avoiding repeated storage reads
- **Simplicity**: Operations have a uniform interface for accessing inputs
- **Traceability**: The context captures the complete state of the calculation, supporting debugging and lineage queries

**Optimization Opportunities**: The context also enables certain optimizations:

- **Batching**: The engine can batch operations that share inputs or outputs
- **Caching**: Results that are used multiple times are computed once and reused
- **Lazy Evaluation**: For large datasets, the engine can load data only when needed

### Cell-by-Cell Transformer Evaluation

The fundamental unit of transformer evaluation is the cell—a single value in a time series, identified by its dimension values and timestamp. The engine evaluates transformers cell by cell, calling the transformer function with corresponding cells from each input metric.

**Correspondence Principle**: Corresponding cells are those that share the same dimension values and time period. For example, if calculating `regional_sales` from `regional_demand` and `regional_price`, the engine calls the transformer with the demand and price values for each region and time period combination.

**Evaluation Process**:

1. Identify all unique dimension value combinations for the output metric
2. For each combination:
   - Retrieve the corresponding cells from each input metric
   - Apply any necessary shifters to align dimensions
   - Call the transformer function with the input values
   - Store the result in the output metric at the corresponding position

**Shifter Application**: When inputs have been shifted to match output dimensions, the shifter logic is applied before the transformer call. The shifter transforms the input data into the correct dimensional structure, and the transformer then operates on the shifted data.

**Example Evaluation**:

Consider a transformer that calculates `store_labor` from `store_demand` and `labor_productivity`:

- `store_labor` has dimensions: `[store, week]`
- `store_demand` has dimensions: `[store, week]`
- `labor_productivity` has dimensions: `[store_type]` (a property of the store)

The dimensions do not match directly. The engine must shift `labor_productivity` to align with `store_labor` by looking up the store type for each store. For each store and week combination, the engine:

1. Gets the demand value for that store and week
2. Looks up the store's type and gets the productivity for that type
3. Calls the transformer: `labor = demand / productivity`
4. Stores the result for that store and week

---

## Development Mode

The Calculation Engine supports a development mode used by planners when creating or updating transformer, shifter, and filter functions. This mode enables iterative development by allowing planners to test their code before publishing it to the production environment.

### Purpose

Development mode serves the planner's workflow within the Plugin UI. When a planner is working on a new transformer or modifying an existing one, they need to see the results of their changes immediately, without going through the full publish workflow. Development mode provides this capability by combining the analysis and calculation phases into a single operation that runs against local workspace artifacts.

### Artifact Source

In development mode, artifact definitions come from the local workspace rather than the object store. The planner's Docker container (managed by the Plugin UI) contains the function code being developed, along with any associated test data. This local sourcing means:

- **Immediate Feedback**: Changes to function code are reflected immediately in the next development run
- **No Database Updates**: The development run does not modify the production database or object store
- **Isolation**: Development work cannot accidentally affect production calculations

### Combined Analysis and Calculation

Development mode performs both the publish phase operations (AST analysis, dimension matching, shifter discovery) and the update phase operations (DAG traversal, calculation execution) in a single invocation. This combination is possible because:

1. The function code is already present in the local workspace
2. No persistent DAG storage is needed—the DAG is constructed and immediately executed
3. Results are returned directly to the plugin rather than being stored

**Development Mode Flow**:

1. **Artifact Collection**: Gather transformer, shifter, and filter code from the local workspace
2. **AST Analysis**: Analyze function signatures to extract input and output metric names
3. **Dimension Validation**: Retrieve metric dimensions from the database and validate compatibility
4. **Shifter Discovery**: Find shifters needed to align input dimensions with output dimensions
5. **DAG Construction**: Build the execution DAG for the functions being tested
6. **Calculation Execution**: Execute the DAG using test data or sample production data
7. **Result Return**: Return the Parquet file to the plugin for display

### Result Consumption

The Parquet file produced by development mode is read by the Plugin UI and used in several ways:

**Metric Graph Visualization**: The plugin displays the dependency graph showing how metrics relate to each other. This visualization helps planners understand the structure of their calculations and verify that the expected relationships exist.

**Data Display**: The actual calculated values are shown to the planner, allowing them to verify that the calculations produce reasonable results. This display includes:

- Time series values for each metric
- Breakdown by dimension values
- Comparison with expected values (when test data includes expected outputs)

**Test Evaluation**: When the planner has defined test cases with expected outputs, the plugin compares the calculated results against the expected values. The test results show:

- Which test cases passed (calculated values match expected values within tolerance)
- Which test cases failed, with details on the discrepancy
- Summary statistics on pass/fail rates

### Error Handling in Development Mode

Development mode provides detailed error information to help planners debug their code:

| Error Type                | Description                                              | Feedback Provided                                                     |
| ------------------------- | -------------------------------------------------------- | --------------------------------------------------------------------- |
| Syntax Error              | Function code cannot be parsed                           | Line number and description of the syntax problem                     |
| Signature Mismatch        | Function name or parameters don't match expected metrics | List of expected metrics vs. actual function signature                |
| Dimension Incompatibility | No shifter chain can align dimensions                    | Visual diagram showing dimension mismatch and suggested shifter types |
| Runtime Error             | Exception during calculation                             | Stack trace with local variable values at the point of error          |
| Test Failure              | Calculated value differs from expected                   | Side-by-side comparison of expected vs. actual values                 |

### Integration with Plugin UI

The Plugin UI orchestrates development mode execution through the Docker-based sandbox. When a planner requests a development run:

1. The plugin sends the local function code to the sandbox container
2. The container executes the calculation engine in development mode
3. Results are returned to the plugin as a Parquet file
4. The plugin parses the file and updates the UI displays

This integration ensures that development runs happen in the same sandboxed environment used for production calculations, providing consistent behavior between development and production.

---

## Result Capture

At the end of the calculation process, results are captured as Parquet files using the format defined in the [Data Storage](data_storage.md) documentation. This format stores the complete output of the calculation—the values of all metrics in the namespace for the calculation's time range.

### Parquet File Generation

The engine constructs the Parquet file from the final state of the execution context. Each metric in the context is written to the file with its dimension values, timestamps, and values.

**File Structure**: The Parquet file follows the schema defined in [Time Series Object Data Format](data_storage.md#time-series-object-data-format):

- Custom metadata captures metric definitions, dimensions, time grain, and data types
- Table columns store the actual data: metric name, dimension values, data type (forecast/actuals), timestamp, and value

**Storage Location**: The Parquet file is written to the object store at a location determined by the storage configuration and the document metadata. The location is recorded in the database for later retrieval.

### Document Creation

The calculation results are associated with a document record in the database. This document captures:

- The document type (the template that defines the structure)
- The `as_of` timestamp (when this calculation was performed)
- The storage location (where the Parquet file is stored)
- Lineage information (which input documents were used)

This document-based organization enables the versioning and lineage tracking that are central to ChronoCone's design. Each calculation produces a new document, preserving the history of how values have changed over time.

---

## Error Handling

The Calculation Engine provides clear error messages at both publish time and update time.

### Publish-Time Errors

Publish-time errors indicate problems with the calculation logic definition:

| Error Type            | Description                                                                    | Resolution                                              |
| --------------------- | ------------------------------------------------------------------------------ | ------------------------------------------------------- |
| Metric Not Found      | A function name or parameter does not match any metric in the namespace        | Create the missing metric or fix the function signature |
| Duplicate Transformer | Multiple transformers claim to produce the same metric                         | Remove or rename one of the transformers                |
| Dimension Mismatch    | Input and output dimensions are incompatible and no shifter can bridge the gap | Create an appropriate shifter or modify the metrics     |
| Circular Dependency   | The DAG contains a cycle, making calculation impossible                        | Restructure the calculations to eliminate the cycle     |

### Update-Time Errors

Update-time errors indicate problems during execution:

| Error Type         | Description                                             | Resolution                                     |
| ------------------ | ------------------------------------------------------- | ---------------------------------------------- |
| Missing Input Data | An input metric has no data for the required time range | Load the missing data or adjust the time range |
| Execution Error    | A transformer, shifter, or filter raised an exception   | Debug the function code                        |
| Resource Limit     | The calculation exceeded memory or time limits          | Optimize the calculation or increase resources |

---

## Design Considerations

### Determinism

The Calculation Engine is designed to be deterministic: given the same inputs and the same code, it produces the same outputs. This determinism is essential for reproducibility and debugging. It enables users to understand exactly how a particular result was produced and to reproduce historical calculations.

### Separation of Concerns

The engine separates the concerns of validation, planning, and execution:

- **Validation** happens at publish time, catching errors early
- **Planning** (DAG construction) happens at publish time, optimizing for execution efficiency
- **Execution** happens at update time, simply following the pre-computed plan

This separation makes each phase simpler and more focused, and it makes the overall system more predictable.

### Extensibility

The engine is designed to support future extensions:

- New operation types can be added alongside transformers, shifters, and filters
- New execution modes (parallel, distributed, GPU) can be implemented without changing the DAG structure
- New optimization strategies can be incorporated into the context management

---

## Related Documentation

- [v0 Architecture Overview](v0_architecture.md)
- [Metrics Service](metrics_service.md)
- [Data Storage](data_storage.md)
- [MCP Server - Planning](mcp_server_planning.md)
- [Plugin UI](ui.md#plugin-ui)
