## Synthesizing Go Code From Chronocone Intermediate Representation

After synthesis of a complete description of a particular desired set of outputs, based on a set of CPL descriptions or Chronocone element definitions read from (a) data store(s) an executable program will be synthesized from the Chronocone Intermediate Language (CIL). This document specifies the approach used to synthesize Go code from CIL with a particular eye on creating code that will execute efficiently on single CPUs.

Because the kinds of business calculations that Chronocone is designed to support aren't algorithmically intensive, because the data associated with forecasts is potentially large, and because it is likely that most groups of interrelated connectors, lookups and metrics will provide ample opportunities for parallelization, the design here is based on the assumption that the fastest synthesized code will attempt to:

1. Minimize data movement, whether that involves I/O or memory movement. Thus, breaking data into chunks that results in a chain of calculations that reads/writes to cache will substantially outperform a chain of calculations that reads/writes to random parts of main memory. In many operations, particularly those that require data being sent over a network (e.g., to horizontally scale a calculation across a number of machines), breaking data into pieces that are too small may result in the overhead of the data movement overwhelming the gains from parallelization. Thus, all parallelization must take into account those overheads.
2. Subject to the constraints implied by #1, parallelization should be used to minimize the latency of calculating the results of any set of metrics as specified by the user.
3. For the design in the current documents, we assume that all data for a computation will fit in main memory at the same time. Other mechanisms beyond the scope of the current documents will detail how to handle horizontal distribution in cases where data exceeds memory capacity.
4. CPL defines base data types for Number and String as unlimited in size (and for Number, precision). When it is not possible to put limits on the possible size or precision of data, the synthesized program will handle that data in a way that works even when data sizes exceed the natively supported types in the language. However, it is recognized that for most calculations most of the time the overhead of using BigDecimal or other software based types will have high overhead and not be necessary. Therefore, the use of mechanisms that will allow the most restricted type implied by a calculation is encouraged.

Note: A fully general optimization of type handling and I/O vs. latency in horizontal scaling is potentially complex. It is likely that early versions of code synthesized for a target (including Go) will embody some simplifying assumptions, such as the update happening on a single host or numbers fitting in native floats or ints. However, a fully compliant synthesized program will not include those limitations.

---

## Computation Model

### CPU Architecture Overview

The synthesized code execution model is based on modern CPU architecture where:

- A **CPU** contains multiple **cores**, each with an independent computation unit and an independent cache
- All cores in the same CPU share the same **main memory**
- Each core has a cache that provides faster access to frequently used data (the simplified model assumes one cache per core)
- Main memory is significantly slower than cache but provides much larger storage capacity

### Memory Access Pattern

Most low-level operations in the synthesized code follow this pattern:

1. Operations take as input a **pointer** and **indices** that reference data in the common main memory
2. The computation reads from the referenced memory locations
3. Results are written back into the common main memory

From the viewpoint of the Go code, operations read from and write to main memory. However, physically, writes may initially go to the CPU cache, with cache synchronization or flush operations later pushing the data back to main memory. This cache behavior is handled automatically by the hardware and is transparent to the synthesized code.

This design allows multiple goroutines (potentially running on different cores) to operate on the same data structures safely, with each goroutine working on a distinct subset of the data.

### Cache Optimization Strategy

While the synthesized code does not have direct control over what data gets loaded into individual CPU caches, the **chunking strategy** controls the size of inputs and outputs being operated on in low-level operations like `addOp`, `multiplyOp`, and similar primitives. By carefully selecting chunk sizes:

- The working set of a single operation (input data + output data) can fit within CPU cache
- Sequential operations on the same chunk can reuse data already in cache
- Memory bandwidth is used efficiently by accessing contiguous blocks of data

The chunking strategy divides tensors into chunks where each chunk represents a contiguous block of memory. When operations are applied to these chunks:

1. The chunk data is loaded into cache when first accessed
2. Subsequent operations on the same chunk hit the cache rather than main memory
3. Multiple chunks can be processed in parallel on different cores

### Data Locality

The design emphasizes **data locality** through:

- **Contiguous storage**: Tensor data is stored in one-dimensional arrays with values arranged contiguously in memory
- **Chunk-based processing**: Operations process contiguous ranges of data at a time
- **Sequential access patterns**: Within a chunk, operations traverse data sequentially to maximize cache line utilization

This approach ensures that once data is loaded into cache (an automatic hardware operation), subsequent accesses to nearby data elements are cache hits, dramatically improving performance over random memory access patterns.

---

## Conceptual Design

### Overview

The synthesis process transforms CIL into executable Go code by orchestrating tensor operations across multiple dimensions. The conceptual design focuses on how operations are coordinated, how data flows through the computation, and how parallelization is achieved.

### Inputs To The Synthesized Program

The synthesized program will have the following inputs:

1. A set of values for the concepts used (directly or indirectly) in the CIL program, including conditional relationships. It's expected that these will usually be read from a database or obtained from a service.
2. A set of values that represent the machine characteristics that the code is optimized for, including the number of physical cores, the amount of main memory and the size of the cache associated with each core.
3. Optionally, an upper bound on the precision of numbers in metrics that use types that are not more tightly constrained.
4. A set of input files in Parquet format that will be used to define the input metrics.

Discussion: We can do things to handle figure out the magnitudes of numbers in a calcuation and use native types when possible, but we can't as easily figure out the needed precisions required for the output if it's not specified in the type declarations. However, in typical business calcuations there are fairly tight bounds on the precision, so rather than force a general treatment #3 will give the users a way to specify a reasonable catchall. The other way would be to try to build it into the definition of the base Number type itself, but that's trickier because of the 'can add constraints but not take them away rule', so if 5 decimal places is normally plenty but there's a special case where 10 is needed the approach in #3 would handle that, but putting the definition in the base type would not.

### Outputs From The Synthesized Program

The synthesized program will have the following outputs:

1. An output file in Parquet format that will be used to define the output metrics and any metrics that need to be calculated to produce the output metrics.
2. An output that represents the constraints the come from the input specification and the input data dependencies.

### Parallelization

CIL is designed so that any arbitrary subtree of the calculation tree can be calcuated in parallel with any other subtree that is not in the path from the subtree to the root of the tree that it's in AND that it's possible to break the tensor data into sections that can be calculated independently and then combined, provided that certain conditions apply. Those piecewise calcuations can potentially be done on either or both of multiple cores in each machine in a cluster or broadcast over multiple machines. In choosing the size of the tensor sections it's anticipated that the most common constraint will be the desire to break tensors into small enough parts so that a series of calculations can mostly read/write from/to the cache, however other considerations may be important as well such as the desire to horizontally scale calculations or handle data that can't fit into memory. The initial focus of this design will be efficient execution on a single machine, but we should be careful to not build in limitations that would make it difficult to extend to a cluster or to data that can't fit into the memory of any single host.

### Tensor Representation

The underlying data organization for CIL is multi-dimensional tensors. There are two general types of calculations that will be done in the Chronocone system:

1. Element-wise calcuations that come from the calcuation clauses in metrics and in unit conversion operations.
2. Matrix multiply operations that are associated with aggregators, spreaders and time-shifting operations.

For element-wise operations, the key property is the size of the output is always the same as the size of the input, and for any given order of dimensions the same order will be preserved (including treating time as just another dimension). From the viewpoint of the Go code, each tensor will be a one dimensional array of values of appropriate types which will broken into a number of chunks where each chunk is a contiguous block of memory and an order number that indicates the position of the chunk in the tensor. Those chunks, with their associated operations may be arbitrarly distributed to cores for calculation without regard for alignment of dimensions (i.e., the size of the chunks may be chose purely to optimize performance). Once a string of calculations is completed, the resulting chunks can be reassembed into the correct result simply by ordering them by their respective order numbers into a full one dimensional array. When translation into a more structure form is required for dimension shifting operations the translation to a multi-dimensional tensor interpretation can be done using arithmetic that uses the dimesion order and the number of values for each dimension.

To support the matrix multiplication operations associated with aggregators, spreaders and time-shifting operations more structure must be preserved, in particular the set of values that is associated with each row and/or column of a matrix in the calculation must be preserved (whether row or column information needs to be preserved depends on the order of any given matrix in a calculation). The individual row/column operations may be calcuated in parallel and the result reassembled into the correct order, but parallelizing operations within a row/column, while possible, is more difficult. If and when we need Chronocone to do that we can consider the problem at that time.

### Operations and Types

The synthesized Go program will have hard knowledge built into it about the base types, including correspondence between the size of values and their possible representations. The built-in types are defined in the CPL requirements. It will also have hard knowledge built into it about the operations that can be performed on those types. The operations are defined in the Chronocone Intermediate Language (CIL) architectural specification.

---

## CIL to Synthesis Relationship

### Overview

The Chronocone Intermediate Language (CIL) provides an abstract, language-independent representation of computations. The synthesis phase transforms this abstract representation into executable Go code. This section describes the relationship between the CIL representation and the synthesized code.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        CIL Representation Layer                              │
│  ┌──────────────────┐     ┌──────────────────┐     ┌──────────────────┐     │
│  │ CILAbstractNode  │────▶│ CILAbstractNode  │────▶│ CILAbstractNode  │     │
│  │ (Data Constant)  │     │ (Operation)      │     │ (Metric Ref)     │     │
│  └──────────────────┘     └──────────────────┘     └──────────────────┘     │
│                                  │                                          │
│                                  ▼                                          │
│                         ┌──────────────────┐                                │
│                         │ operationType    │                                │
│                         │ children[]       │                                │
│                         └──────────────────┘                                │
└─────────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼ Transform via CILAbstractTransformer
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Synthesized Go Code Layer                             │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  Generated Computation Function (one per metric)                    │    │
│  │                                                                     │    │
│  │  func computeX(controller, inputs) {                                │    │
│  │      // 1. Create input operands (ChronoconeTensorComputationChunkOperand)│
│  │      operand1 := NewChronoconeTensorComputationChunkOperand(...)    │    │
│  │      operand2 := NewChronoconeTensorComputationChunkOperand(...)    │    │
│  │                                                                     │    │
│  │      // 2. Create output operand                                    │    │
│  │      output := NewChronoconeTensorComputationChunkOperand(...)      │    │
│  │                                                                     │    │
│  │      // 3. Create chunk with operation function                     │    │
│  │      chunk := &ChronoconeTensorComputationChunk{                   │    │
│  │          operands: []*ChronoconeTensorComputationChunkOperand{operand1, operand2},│
│  │          matrixOp: addTensorsOp,  // from CIL operationType         │
│  │          output: output,                                            │    │
│  │      }                                                              │    │
│  │                                                                     │    │
│  │      // 4. Submit to work channel                                   │    │
│  │      controller.SubmitChunk(chunk)                                  │    │
│  │  }                                                                  │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                       │                                      │
│                                       ▼                                      │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  Work Channel Dispatch (controller.workChannel)                     │    │
│  │  Buffer size = CPUCoreCount (from IO specification)                 │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                       │                                      │
│                                       ▼                                      │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  Worker Goroutines (numWorkers = CPUCoreCount)                      │    │
│  │                                                                     │    │
│  │  for chunk := range workChannel {                                   │    │
│  │      chunk.matrixOp(chunk.operands, chunk.output)                   │    │
│  │  }                                                                  │    │
│  │  // Exits when channel closed                                       │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Mapping CIL Nodes to Synthesized Code

The [`CILAbstractTransformer`](#cilabstracttransformer) maps each CIL node type to specific synthesis patterns:

| CIL Node Type                                                                                   | Synthesis Target  | Generated Code Pattern                                        |
| ----------------------------------------------------------------------------------------------- | ----------------- | ------------------------------------------------------------- |
| [`CILDataConstantNode`](../go-analysis/cil-representation.md#cildataconstantnode)               | Constant value    | Scalar literal or constant tensor creation                    |
| [`CILUnitNameConstantNode`](../go-analysis/cil-representation.md#cilunitnameconstantnode)       | Unit reference    | String reference for unit conversion lookup                   |
| [`CILMeasureNameConstantNode`](../go-analysis/cil-representation.md#cilmeasurenameconstantnode) | Measure reference | Type information lookup                                       |
| [`CILMetricReferenceNode`](../go-analysis/cil-representation.md#cilmetricreferencenode)         | Tensor input      | `controller.GetMetricComputation(name).GetTensor()`           |
| [`CILAbstractOperationNode`](../go-analysis/cil-representation.md#cilabstractoperationnode)     | Chunk submission  | `ChronoconeTensorComputationChunk` creation + `SubmitChunk()` |

### Operation Transformation

Each [`CILOperationType`](../go-analysis/cil-representation.md#operation-type-constants) maps to a specific operation function that is embedded in the `ChronoconeTensorComputationChunk`:

| CIL Operation Type                    | Chunk Field                | Operation Function Type    |
| ------------------------------------- | -------------------------- | -------------------------- |
| `CILOperationTypeNegate`              | `unaryOp`                  | `UnaryOperation[T]`        |
| `CILOperationTypeAdd` (tensor-tensor) | `matrixOp`                 | `MatrixOperation[T]`       |
| `CILOperationTypeAdd` (scalar-tensor) | `pairwiseOp`               | `PairwiseOperation[L,T]`   |
| `CILOperationTypeMultiply`            | `matrixOp` or `pairwiseOp` | Depends on operand types   |
| `CILOperationTypeAnd`                 | `matrixOp` or `pairwiseOp` | Boolean operation function |
| `CILOperationTypeMaskFromData`        | `unaryOp`                  | `UnaryOperation[T]`        |

The transformer examines operand types during tree walking to determine whether to use matrix (tensor-tensor) or pairwise (scalar-tensor) operations.

---

## CILAbstractTransformer and Code Generation

### Overview

The [`CILAbstractTransformer`](#cilabstracttransformer) is the core component responsible for transforming CIL abstract syntax trees into executable Go code. Unlike simple template-based code generation, the transformer employs a sophisticated tree-walking strategy that organizes code generation around computational dependencies and parallelization opportunities.

### Tree Walking and Operation Grouping

A CIL abstract tree consists of a tree for each metric or connector. As the [`CILAbstractTransformer`](#cilabstracttransformer) walks each tree, it encounters various situations:

- Operations that act on measures or units
- Operations with different unit expressions requiring unit conversion
- Operations with different dimension structures
- References to other metrics or connectors

The transformer organizes code generation into a hierarchical structure that captures both data dependencies and parallelization boundaries.

#### Hierarchy of Computation Organization

The transformer maintains the following hierarchy during tree walking:

```
ParallelOperationGroup
    └── Stack of ComputationGroup
            └── Stack of Computation
                    └── Stack of code snippets (strings)
```

**ParallelOperationGroup**: Represents a collection of computations that can execute in parallel as goroutines. Each metric or connector typically has its own ParallelOperationGroup.

**ComputationGroup**: Represents a sequence of computations that share the same general calculation type, units, measures, and dimensions. When the transformer detects a change in any of these properties during tree walking, it creates a new ComputationGroup.

**Computation**: Represents a single cohesive computation within a group. It maintains a stack of code snippets that will be assembled into a function.

**Code Snippets**: Individual lines of Go code (strings) that perform specific operations.

### Code Snippet Accumulation Strategy

As the [`CILAbstractTransformer`](#cilabstracttransformer) walks the CIL abstract tree, it generates code snippets and pushes them onto the stack in the current [`Computation`](#computation) object. The stack ordering ensures that leaf-level computations (operations at the top of the stack) are performed first when the code is executed.

For example, given the CPL expression `1.2 * (inventory - 8)`, the transformer would generate:

```go
temp1, err := inventory.SubtractScalar(controller, 8.0)
if err != nil {
    // Handle error
}
temp2, err := temp1.MultiplyScalar(controller, 1.2)
if err != nil {
    // Handle error
}
```

These snippets are pushed onto the stack in the order encountered during the tree walk, but when assembled into a function, they execute in stack order (last-in, first-out), ensuring proper dependency ordering.

**Note on Generated Code Style**: The generated code uses the high-level tensor operation methods (`SubtractScalar`, `MultiplyScalar`, etc.) which internally handle chunking and parallelization. This provides a clean abstraction that matches the CIL operations while delegating the complexity of chunk management to the tensor library. See [Three-Level Function Hierarchy](tensor-representation.md#three-level-function-hierarchy) for details on how these high-level operations map to low-level chunk execution.

### Dependency Order and Function Assembly

When a change in calculation is detected (change in unit, measure, or dimensions), a new [`ComputationGroup`](#computationgroup) is created. Each ComputationGroup on the stack represents a distinct phase of computation with consistent properties.

The functions that represent operations for a [`ParallelOperationGroup`](#paralleloperationgroup) are called as goroutines. The snippets from each [`Computation`](#computation) object are assembled into functions, with ComputationGroups being called in bottom-to-top stack order. This mechanism captures the dependency order of operations:

1. **Bottom of stack**: Earlier computation phases with resolved dependencies
2. **Top of stack**: Later computation phases that depend on earlier results

### Two Levels of Parallelism

The synthesis design employs two distinct levels of parallelism:

1. **Goroutine-Level Parallelism**: Functions representing different [`ParallelOperationGroup`](#paralleloperationgroup) instances execute as goroutines. This level is managed by the synthesized code structure.

2. **Tensor-Level Parallelism**: Within tensor operations, data may be sliced and processed in parallel across multiple goroutines. This level is implemented within the synthesis support library and is not a synthesis concept per se.

### Metric and Connector References

One important type of [`CILAbstractNode`](../go-analysis/cil-representation.md) is a reference to a metric or connector name. These references indicate that a ChronoconeTensor computed in another [`ParallelOperationGroup`](#paralleloperationgroup) should be used as input. However, the computation may not yet be complete when the reference is encountered.

To handle this synchronization challenge, the system uses a channel-based coordination mechanism:

#### MetricComputation and Synchronization

The [`ChronoconeComputationController`](#chronoconecomputationcontroller) maintains a map of metric names to [`MetricComputation`](#metriccomputation) objects. Each [`MetricComputation`](#metriccomputation) contains:

1. A channel initialized when the object is allocated
2. A pointer to a ChronoconeTensor, initially nil
3. A consumer count used for memory management

The channel is never used to send data. Instead, a function associated with [`MetricComputation`](#metriccomputation) attempts to read from the channel:

- **If the channel is open**: The function blocks until the channel is closed
- **If the channel is closed**: The function returns the pointer to the tensor

When a [`ComputationGroup`](#computationgroup) finishes its calculation, it calls the appropriate [`MetricComputation.SetTensor()`](#metriccomputation) function, passing the computed tensor. The `SetTensor` function closes the channel, unblocking any goroutines waiting for that metric's result.

**Critical Requirement: DAG Validation**
Because this synchronization mechanism relies on blocking channel reads, circular dependencies between metrics (e.g., Metric A depends on Metric B, which depends on Metric A) will result in a deadlock where goroutines wait indefinitely. It is a strict requirement that the CIL Analysis phase validates the metric dependency graph to ensure it is a Directed Acyclic Graph (DAG) before synthesis begins. Any circular dependencies must be caught and reported as errors during analysis.

#### Memory Management via Static Reference Counting

To prevent excessive memory bloat in large DAGs, the system employs static reference counting for memory management. Since the `CILAbstractTransformer` has complete knowledge of the computation DAG before synthesis, it can statically determine exactly how many times each metric is consumed as an input by other metrics.

1. **Initialization**: During `Setup()`, the transformer generates `NewMetricComputation(N)` where `N` is the exact number of times that metric is referenced by other computations. (If a metric is an output of the entire program, its count is incremented by 1 to ensure it survives until `writeMetricOutput`).
2. **Release**: In the generated `computeX()` functions, after a consumer has finished using an input tensor, it calls `controller.GetMetricComputation("input_metric").ReleaseTensor()`.
3. **Reclamation**: `ReleaseTensor()` atomically decrements the consumer count. When the count reaches 0, it sets the internal tensor pointer to `nil`. This removes the last strong reference to the large data arrays, allowing the Go Garbage Collector to reclaim the memory immediately, rather than waiting for the end of the program.

_Note: While this static reference counting mechanism is part of the computation model design, its implementation is deferred and out of scope for the initial release. The initial implementation will hold all tensors in memory until `Cleanup()`. We will revisit this memory management strategy if and when we need to optimize memory usage._

### Setup, Compute, and Cleanup Lifecycle

For the metric map and similar initialization needs to be properly set up before computation begins, the synthesized program follows a three-phase lifecycle:

1. **Setup()**: Initializes all data structures, including the metric map in [`ChronoconeComputationController`](#chronoconecomputationcontroller)
2. **Compute()**: Executes the parallel computation groups and waits for completion
3. **Cleanup()**: Releases resources and performs any necessary finalization

The main function of the generated code calls these three functions in sequence. It is the responsibility of the [`CILAbstractTransformer`](#cilabstracttransformer) to generate appropriate code for each phase.

---

## Go Implementation

### ChronoconeComputationController

The `ChronoconeComputationController` struct serves as the central coordinator for tensor operations during computation. It holds the context for all tensor allocations, provides access to concept values and dimension information, manages the metric computation map for cross-metric dependencies, and provides a work dispatch channel for parallel chunk execution.

```go
// ChronoconeComputationController coordinates all tensor operations during computation.
type ChronoconeComputationController struct {
    conceptValues       map[string][]interface{}                       // Map from concept name to array of values
    metricComputations  map[string]*MetricComputation                  // Map from metric name to computation state
    workChannel         chan *ChronoconeTensorComputationChunk[any]    // Channel for dispatching computation chunks
    workerWaitGroup     sync.WaitGroup                                 // Coordinates worker goroutine lifecycle
    numWorkers          int                                            // Number of worker goroutines (matches core count)
}

// NewChronoconeComputationController creates a new controller with empty mappings.
// The work channel and worker pool are initialized when SetArchitectureConfig is called.
func NewChronoconeComputationController() *ChronoconeComputationController

// RegisterMetricComputation registers a MetricComputation for a metric name.
// This must be called during Setup() before any parallel computation begins.
func (c *ChronoconeComputationController) RegisterMetricComputation(
    metricName string,
    mc *MetricComputation,
)

// GetMetricComputation retrieves the MetricComputation for a metric name.
// Returns nil if the metric is not registered.
func (c *ChronoconeComputationController) GetMetricComputation(
    metricName string,
) *MetricComputation

// GetWorkChannel returns the work channel for submitting computation chunks.
// Generated code submits chunks to this channel for worker execution.
func (c *ChronoconeComputationController) GetWorkChannel() chan *ChronoconeTensorComputationChunk[any]

// StartWorkers launches the worker goroutines that listen on the work channel.
// This must be called after SetArchitectureConfig and before any chunks are submitted.
// Workers continue until the work channel is closed.
func (c *ChronoconeComputationController) StartWorkers()

// StopWorkers signals all workers to exit by closing the work channel and waits for completion.
// This should be called during Cleanup() to ensure graceful shutdown.
func (c *ChronoconeComputationController) StopWorkers()

// SubmitChunk submits a computation chunk to the work channel for execution.
// This call blocks if the channel buffer is full, providing natural backpressure.
// Parameters:
//   - chunk: The computation chunk to execute
func (c *ChronoconeComputationController) SubmitChunk(chunk *ChronoconeTensorComputationChunk[any])
```

**Field Descriptions:**

- **conceptValues**: A map indexed by concept name (string) where each value is an array of the values for that concept. This provides the dimension sizes and value domains for all tensors created during computation.

- **metricComputations**: A map from metric name to [`MetricComputation`](#metriccomputation) objects. This map enables synchronization between parallel computation groups that depend on each other's results.

- **workChannel**: A buffered channel of [`ChronoconeTensorComputationChunk`](#chronoconetensorcomputationchunk) pointers. The buffer size equals the number of cores from the IO specification. Generated code submits chunks to this channel; worker goroutines receive and execute them.

- **workerWaitGroup**: A [`sync.WaitGroup`](https://pkg.go.dev/sync#WaitGroup) that tracks active worker goroutines. Used to coordinate graceful shutdown when the channel is closed.

- **numWorkers**: The number of worker goroutines to launch, which equals the number of CPU cores specified in the IO specification's architecture configuration.

The controller is passed to tensor creation functions to allow tensors to look up their dimension names and sizes from the central concept registry rather than having these passed as separate parameters.

#### Worker Goroutine Execution Model

The `ChronoconeComputationController` manages a pool of worker goroutines that execute computation chunks dispatched through the work channel:

```go
// worker is the goroutine function that listens on the work channel and executes chunks.
// It runs until the work channel is closed, then exits.
// Parameters:
//   - c: The computation controller (provides access to execution context)
func (c *ChronoconeComputationController) worker() {
    defer c.workerWaitGroup.Done()

    for chunk := range c.workChannel {
        // Execute the operation defined in the chunk
        // Only one of matrixOp, pairwiseOp, or unaryOp will be non-nil
        var err error
        if chunk.matrixOp != nil {
            // Matrix operation: operates on all input operands
            err = chunk.matrixOp(chunk.operands, chunk.output)
        } else if chunk.pairwiseOp != nil {
            // Pairwise operation: binary operation with scalar left operand
            // Note: pairwiseOp requires special handling as it combines scalar and tensor
            err = c.executePairwiseOp(chunk)
        } else if chunk.unaryOp != nil {
            // Unary operation: single input operand
            // Extract the single input operand
            if len(chunk.operands) != 1 {
                err = fmt.Errorf("unary operation requires exactly 1 operand, got %d", len(chunk.operands))
            } else {
                result := chunk.unaryOp(chunk.operands[0])
                // Write result to output operand
                c.writeUnaryResult(chunk.output, result)
            }
        }

        // Report error if any
        if err != nil && chunk.errorChan != nil {
            // Non-blocking send to error channel
            select {
            case chunk.errorChan <- err:
            default:
            }
        }

        // Signal completion
        if chunk.waitGroup != nil {
            chunk.waitGroup.Done()
        }
    }

    // Channel closed, worker exits
}
```

**Key Characteristics of the Worker Model:**

1. **Blocking Receive**: Workers block on `range c.workChannel` until a chunk is available or the channel is closed.

2. **Operation Dispatch**: Each chunk contains exactly one operation type (matrix, pairwise, or unary). The worker dispatches to the appropriate execution logic based on which function pointer is non-nil.

3. **Operand Access**: The `operands` field contains pointers to [`ChronoconeTensorComputationChunkOperand`](tensor-representation.md#chronoconetensorcomputationchunkoperand) objects that provide access to tensor data slices and their associated masks.

4. **Output Writing**: The `output` field specifies where results are written. Workers write directly to the output operand's tensor data slice.

5. **Error Handling**: Errors during chunk execution are reported through the controller's error handling mechanism. Individual chunk failures do not stop the worker.

6. **Never Exits Until Channel Closed**: The `for range` loop ensures the worker continues processing chunks indefinitely until `StopWorkers()` closes the channel.

7. **Parallelism Control**: The number of workers equals the core count, ensuring natural parallelization without oversubscription. The buffered channel provides backpressure when all workers are busy.

### MetricComputation

The `MetricComputation` struct provides synchronization for cross-metric dependencies using a channel-based coordination pattern.

```go
// MetricComputation tracks the computation state for a single metric.
// It enables goroutines to wait for a metric's tensor to be computed.
type MetricComputation struct {
    done          chan struct{}           // Channel that closes when tensor is ready
    tensor        *ChronoconeTensor[any]  // Pointer to the computed tensor
    consumerCount int32                   // Atomic counter for memory management
}

// NewMetricComputation creates a new MetricComputation with an open channel.
// This should be called during Setup() for each metric that may be referenced.
// The consumerCount specifies how many times this metric is referenced by other computations.
func NewMetricComputation(consumerCount int32) *MetricComputation

// GetTensor blocks until the tensor is available, then returns it.
// If the tensor is already set, returns immediately.
func (mc *MetricComputation) GetTensor() *ChronoconeTensor[any]

// SetTensor sets the tensor and closes the done channel.
// This unblocks any goroutines waiting in GetTensor().
// Returns an error if SetTensor has already been called.
func (mc *MetricComputation) SetTensor(tensor *ChronoconeTensor[any]) error

// ReleaseTensor atomically decrements the consumer count.
// When the count reaches 0, it sets the internal tensor pointer to nil,
// allowing the Go Garbage Collector to reclaim the memory.
func (mc *MetricComputation) ReleaseTensor()

// IsReady returns true if the tensor has been set.
func (mc *MetricComputation) IsReady() bool
```

**Synchronization Pattern:**

The channel-based synchronization works as follows:

1. During Setup(), a [`MetricComputation`](#metriccomputation) is created for each metric with an open (unbuffered) channel
2. When a computation needs a tensor from another metric, it calls [`GetTensor()`](#metriccomputation), which attempts to receive from the channel
3. If the channel is still open, [`GetTensor()`](#metriccomputation) blocks
4. When the computation producing the tensor completes, it calls [`SetTensor()`](#metriccomputation), which stores the tensor pointer and closes the channel
5. All blocked goroutines wake up and return the tensor pointer

This pattern is safe because:

- Closing a channel is a one-time, permanent operation
- Multiple goroutines can wait on the same closed channel and all will proceed
- The zero-value of the tensor pointer is nil, which is a valid state before computation

### ParallelOperationGroup

The `ParallelOperationGroup` struct represents a collection of computations that can execute in parallel.

```go
// ParallelOperationGroup represents a group of computations that can run in parallel.
// Typically corresponds to a single metric or connector.
type ParallelOperationGroup struct {
    name               string                      // Name of the metric/connector
    computationGroups  []*ComputationGroup         // Stack of computation groups
    dependencies       []string                    // Names of metrics this group depends on
}

// NewParallelOperationGroup creates a new group for the given metric name.
func NewParallelOperationGroup(name string) *ParallelOperationGroup

// PushComputationGroup adds a new computation group to the top of the stack.
func (pog *ParallelOperationGroup) PushComputationGroup(cg *ComputationGroup)

// PopComputationGroup removes and returns the top computation group.
func (pog *ParallelOperationGroup) PopComputationGroup() *ComputationGroup

// AddDependency registers a dependency on another metric.
func (pog *ParallelOperationGroup) AddDependency(metricName string)

// Execute runs all computation groups in dependency order and sets the result.
// This function is executed as a goroutine.
func (pog *ParallelOperationGroup) Execute(controller *ChronoconeComputationController)
```

### ComputationGroup

The `ComputationGroup` struct represents a sequence of computations that share the same calculation properties (units, measures, dimensions).

```go
// ComputationGroup represents a sequence of computations with consistent
// calculation properties (same units, measures, dimensions).
type ComputationGroup struct {
    unitExpression     string           // Unit expression for this group
    measureName        string           // Measure name for this group
    dimensionNames     []string         // Dimensions for this group
    computations       []*Computation   // Stack of computations
}

// NewComputationGroup creates a new group with the specified properties.
func NewComputationGroup(
    unitExpression string,
    measureName string,
    dimensionNames []string,
) *ComputationGroup

// PushComputation adds a new computation to the top of the stack.
func (cg *ComputationGroup) PushComputation(c *Computation)

// PopComputation removes and returns the top computation.
func (cg *ComputationGroup) PopComputation() *Computation

// GenerateFunction creates a Go function string from all computations in this group.
// Computations are processed in stack order (bottom to top of stack).
func (cg *ComputationGroup) GenerateFunction() string
```

### Computation

The `Computation` struct represents a single cohesive computation within a group, maintaining a stack of code snippets.

```go
// Computation represents a single cohesive computation with a stack of code snippets.
type Computation struct {
    snippets  []string  // Stack of code snippets (leaf operations at top)
    tempVarCounter int  // Counter for generating unique temporary variable names
}

// NewComputation creates a new empty computation.
func NewComputation() *Computation

// PushSnippet adds a code snippet to the top of the stack.
func (c *Computation) PushSnippet(snippet string)

// PopSnippet removes and returns the top code snippet.
func (c *Computation) PopSnippet() string

// GenerateTempVar generates a unique temporary variable name.
func (c *Computation) GenerateTempVar() string

// AssembleCode combines all snippets into a single function body.
// Snippets are assembled in reverse stack order so leaf operations execute first.
func (c *Computation) AssembleCode() string
```

### CILAbstractTransformer

The `CILAbstractTransformer` struct orchestrates the transformation of CIL abstract nodes into executable Go code using the hierarchical computation organization.

```go
// CILAbstractTransformer transforms CIL abstract nodes into executable Go code.
type CILAbstractTransformer struct {
    measureTypes         map[string]string               // Map from measure name to Go type
    metricTypes          map[string]string               // Map from metric name to element type
    parallelGroups       []*ParallelOperationGroup       // All parallel operation groups
}

// NewCILAbstractTransformer creates a new transformer with empty type mappings.
func NewCILAbstractTransformer() *CILAbstractTransformer

// RegisterMeasureType registers the Go type for a measure name.
func (t *CILAbstractTransformer) RegisterMeasureType(measureName string, goType string)

// RegisterMetricType registers the element type for a metric name.
func (t *CILAbstractTransformer) RegisterMetricType(metricName string, elementType string)

// Transform walks the CIL abstract tree and builds the parallel operation group hierarchy.
// Returns the generated Go code as a string and any error encountered.
func (t *CILAbstractTransformer) Transform(root cil.CILAbstractNode) (string, error)

// VisitNode processes a single CIL abstract node during tree walking.
// It dispatches to appropriate handlers based on node type.
func (t *CILAbstractTransformer) VisitNode(node cil.CILAbstractNode) error

// TransformMetric generates a complete Go function for computing a metric.
// The generated function takes a controller and inputs map, and returns the result tensor.
func (t *CILAbstractTransformer) TransformMetric(
    metricName string,
    root cil.CILAbstractNode,
) (string, error)

// GenerateSetupFunction generates the Setup() function for initialization.
func (t *CILAbstractTransformer) GenerateSetupFunction() string

// GenerateComputeFunction generates the Compute() function that orchestrates execution.
func (t *CILAbstractTransformer) GenerateComputeFunction() string

// GenerateCleanupFunction generates the Cleanup() function for resource release.
func (t *CILAbstractTransformer) GenerateCleanupFunction() string
```

### Generated Code Structure

The synthesized Go program has the following structure:

```go
// Setup initializes all data structures before computation.
// Generated by CILAbstractTransformer based on the CIL tree analysis.
func Setup(controller *ChronoconeComputationController) error {
    // Register MetricComputation objects for all metrics
    // The integer argument is the static consumer count determined during analysis
    controller.RegisterMetricComputation("inventory", NewMetricComputation(2))
    controller.RegisterMetricComputation("revenue", NewMetricComputation(1))
    // ... more registrations
    return nil
}

// Compute orchestrates the parallel execution of all computation groups.
// Generated by CILAbstractTransformer based on dependency analysis.
func Compute(controller *ChronoconeComputationController, inputs map[string]*ChronoconeTensor[float64]) error {
    // Launch parallel operation groups as goroutines
    go computeInventory(controller, inputs)
    go computeRevenue(controller, inputs)
    // ... more goroutines

    // Wait for all computations to complete
    // The MetricComputation synchronization handles dependencies
    return nil
}

// Cleanup releases resources after computation.
func Cleanup(controller *ChronoconeComputationController) error {
    // Release resources, close channels, etc.
    return nil
}

// Example computation function for a specific metric (generated).
func computeInventory(
    controller *ChronoconeComputationController,
    inputs map[string]*ChronoconeTensor[float64],
) {
    // Allocate output tensor
    result := controller.AllocateTensor[float64](dimensions...)

    // Get input tensors (may block if not yet computed)
    rawInventory := controller.GetMetricComputation("raw_inventory").GetTensor()

    // Perform operations using synthesizer library
    temp1 := tensor.Add(rawInventory, constant)
    temp2 := tensor.Multiply(temp1, factor)

    // Set the result to unblock dependent computations
    controller.GetMetricComputation("inventory").SetTensor(temp2)

    // Release input tensors to allow memory reclamation
    controller.GetMetricComputation("raw_inventory").ReleaseTensor()
}

// Main function that orchestrates the computation lifecycle.
func main() {
    controller := NewChronoconeComputationController()

    if err := Setup(controller); err != nil {
        log.Fatal(err)
    }

    if err := Compute(controller, inputs); err != nil {
        log.Fatal(err)
    }

    if err := Cleanup(controller); err != nil {
        log.Fatal(err)
    }
}
```

### Tensor Operation Primitives

Tensor operations are built on top of two fundamental primitives: chunking strategies and the `apply_to_tensor` function. These primitives enable flexible parallelization and support different execution strategies based on tensor characteristics and hardware constraints.

#### Chunking Strategy

A chunking strategy is a function that divides a tensor into chunks for parallel processing. The strategy determines how tensor data is partitioned for cache-efficient and parallel execution.

```go
// ChunkingStrategy is a function type that determines how a tensor is divided
// into chunks for parallel processing. It takes a tensor and returns a slice
// of ChronoconeTensorComputationChunk objects representing the chunks.
type ChunkingStrategy[T any] func(tensor *ChronoconeTensor[T]) []*ChronoconeTensorComputationChunk[T]

// Standard chunking strategies:

// CacheOptimizedChunking divides the tensor into chunks sized to fit in CPU cache.
// This is the default strategy for most element-wise operations.
func CacheOptimizedChunking[T any](tensor *ChronoconeTensor[T]) []*ChronoconeTensorComputationChunk[T]

// SingleChunk returns the entire tensor as a single chunk.
// Used when parallelization overhead would exceed benefits.
func SingleChunk[T any](tensor *ChronoconeTensor[T]) []*ChronoconeTensorComputationChunk[T]

// FixedSizeChunking creates chunks of a specified fixed size.
// Useful for predictable workload distribution.
func FixedSizeChunking[T any](chunkSize int) ChunkingStrategy[T]

// RowAlignedChunking creates chunks aligned to row boundaries.
// Used for operations that require row-level coherence.
func RowAlignedChunking[T any](tensor *ChronoconeTensor[T]) []*ChronoconeTensorComputationChunk[T]
```

**Chunking Strategy Selection:**

The CILAbstractTransformer selects chunking strategies based on:

- Tensor size (small tensors may not benefit from chunking)
- Operation type (element-wise vs. aggregation)
- Hardware characteristics (cache size, core count)
- Data dependencies (whether chunks can be computed independently)

#### apply_to_tensor Primitive

The `apply_to_tensor` function is the fundamental primitive for applying operations to tensors. It takes a matrix operation function, a chunking strategy, and input tensors, then orchestrates the parallel execution of the operation across chunks.

```go
// MatrixOperation is a function type that performs an element-wise operation
// on a single chunk. It reads from input operands and writes to an output operand.
// All operands extend from ChronoconeTensorComputationChunkOperand.
type MatrixOperation[T any] func(
    inputs []*ChronoconeTensorComputationChunkOperand[T],
    output *ChronoconeTensorComputationChunkOperand[T],
) error

// PairwiseOperation is a function type for binary operations on two input elements.
// It computes output element = op(left, right).
// The right parameter extends from ChronoconeTensorComputationChunkOperand,
// the left parameter extends from any.
// Used when left is a scalar and right is a tensor.
type PairwiseOperation[L any, T any] func(left L, right *ChronoconeTensorComputationChunkOperand[T]) T

// TensorPairwiseOperation is a function type for binary operations on two tensor inputs.
// It computes output element = op(left, right) where both inputs are tensors.
// Both parameters extend from ChronoconeTensorComputationChunkOperand.
// Used for element-wise operations between two tensors of the same shape.
type TensorPairwiseOperation[T any] func(left, right *ChronoconeTensorComputationChunkOperand[T]) T

// UnaryOperation is a function type for unary operations on a single input element.
// It computes output element = op(input).
// The operand extends from ChronoconeTensorComputationChunkOperand.
type UnaryOperation[T any] func(input *ChronoconeTensorComputationChunkOperand[T]) T

// apply_to_tensor applies a matrix operation to input tensors using the specified
// chunking strategy. It creates chunks, executes the operation in parallel across
// chunks, and writes results to the output tensor.
//
// Parameters:
//   - operation: The MatrixOperation to apply to each chunk
//   - chunkingStrategy: The ChunkingStrategy to use for partitioning
//   - inputs: Input tensors for the operation
//   - output: The output tensor where results will be written
//   - controller: The computation controller for tensor allocation
//
// Returns:
//   - An error if the operation fails
func apply_to_tensor[T any](
    operation MatrixOperation[T],
    chunkingStrategy ChunkingStrategy[T],
    inputs []*ChronoconeTensor[T],
    output *ChronoconeTensor[T],
    controller *ChronoconeComputationController,
) error

// apply_unary_to_tensor is a convenience wrapper for unary operations.
// It applies a UnaryOperation to each element of the input tensor and writes results to the output tensor.
//
// Parameters:
//   - op: The UnaryOperation to apply
//   - chunkingStrategy: The ChunkingStrategy to use for partitioning
//   - input: The input tensor
//   - output: The output tensor where results will be written
//   - controller: The computation controller for tensor allocation
//
// Returns:
//   - An error if the operation fails
func apply_unary_to_tensor[T any](
    op UnaryOperation[T],
    chunkingStrategy ChunkingStrategy[T],
    input *ChronoconeTensor[T],
    output *ChronoconeTensor[T],
    controller *ChronoconeComputationController,
) error

// apply_binary_to_tensor is a convenience wrapper for binary operations.
// It applies a PairwiseOperation to combine the left scalar with each element of the right tensor.
//
// Parameters:
//   - op: The PairwiseOperation to apply
//   - chunkingStrategy: The ChunkingStrategy to use for partitioning
//   - left: The left scalar value (primitive type)
//   - right: The right input tensor
//   - output: The output tensor where results will be written
//   - controller: The computation controller for tensor allocation
//
// Returns:
//   - An error if the operation fails
func apply_binary_to_tensor[T any](
    op PairwiseOperation[T],
    chunkingStrategy ChunkingStrategy[T],
    left T,
    right *ChronoconeTensor[T],
    output *ChronoconeTensor[T],
    controller *ChronoconeComputationController,
) error

// apply_binary_tensors_to_tensor is a convenience wrapper for binary tensor-tensor operations.
// It applies a TensorPairwiseOperation to combine two tensors element-wise.
// Both input tensors must have the same dimensions and shape.
//
// Parameters:
//   - op: The TensorPairwiseOperation to apply
//   - chunkingStrategy: The ChunkingStrategy to use for partitioning
//   - left: The left input tensor
//   - right: The right input tensor
//   - output: The output tensor where results will be written
//   - controller: The computation controller for tensor allocation
//
// Returns:
//   - An error if the operation fails or if tensor dimensions do not match
func apply_binary_tensors_to_tensor[T any](
    op TensorPairwiseOperation[T],
    chunkingStrategy ChunkingStrategy[T],
    left *ChronoconeTensor[T],
    right *ChronoconeTensor[T],
    output *ChronoconeTensor[T],
    controller *ChronoconeComputationController,
) error
```

**Execution Model:**

When `apply_to_tensor` is called:

1. **Chunking**: The chunking strategy divides input tensors into chunks
2. **Parallel Execution**: For each chunk, the matrix operation is executed in a goroutine, writing results to the corresponding location in the output tensor
3. **Synchronization**: The function waits for all chunk operations to complete
4. **Error Return**: Returns any error encountered during execution

### Operation Code Generation

Each CIL operation maps to a Go code pattern. The synthesis generates code that creates [`ChronoconeTensorComputationChunk`](tensor-representation.md#chronoconetensorcomputationchunk) objects and submits them to the controller's work channel for execution by worker goroutines.

#### CIL to Chunk Mapping

The [`CILAbstractOperationNode`](../go-analysis/cil-representation.md#cilabstractoperationnode) from the CIL representation is transformed into chunk submissions:

```
CILAbstractOperationNode
    ├── operationType: CILOperationType (Add, Multiply, etc.)
    └── children: []CILAbstractNode (operands)
         ↓ Synthesis Transform
ChronoconeTensorComputationChunk
    ├── operands: []*ChronoconeTensorComputationChunkOperand (input tensor slices)
    ├── matrixOp OR pairwiseOp OR unaryOp (operation function)
    └── output: *ChronoconeTensorComputationChunkOperand (output tensor slice)
         ↓ Submit to Work Channel
controller.SubmitChunk(chunk)
         ↓ Worker Goroutine
Execution on CPU core
```

The [`CILAbstractTransformer`](#cilabstracttransformer) walks the CIL tree and generates code that:

1. **Creates Output Operands**: Allocates output tensors using `controller.AllocateTensor()`
2. **Creates Input Operands**: Wraps input tensor references as `ChronoconeTensorComputationChunkOperand` objects
3. **Applies Chunking Strategy**: Divides tensors into chunks using the configured `ChunkingStrategy`
4. **Creates Chunks**: Constructs `ChronoconeTensorComputationChunk` objects with operation functions and operands
5. **Submits to Channel**: Calls `controller.SubmitChunk()` to dispatch chunks to worker goroutines

| CIL Operation      | Left Operand | Right Operand | Generated Chunk Submission Pattern                                             |
| ------------------ | ------------ | ------------- | ------------------------------------------------------------------------------ |
| Negate             | N/A          | Tensor        | Submit chunks with `unaryOp: negateOp` for each input chunk                    |
| Add                | Scalar       | Tensor        | Submit chunks with `pairwiseOp: addOp` for each tensor chunk                   |
| Add                | Tensor       | Tensor        | Submit chunks with `matrixOp: addTensorsOp` for aligned chunks of both tensors |
| Subtract           | Scalar       | Tensor        | Submit chunks with `pairwiseOp: subtractOp` for each tensor chunk              |
| Subtract           | Tensor       | Tensor        | Submit chunks with `matrixOp: subtractTensorsOp` for aligned chunks            |
| Multiply           | Scalar       | Tensor        | Submit chunks with `pairwiseOp: multiplyOp` for each tensor chunk              |
| Multiply           | Tensor       | Tensor        | Submit chunks with `matrixOp: multiplyTensorsOp` for aligned chunks            |
| Divide             | Scalar       | Tensor        | Submit chunks with `pairwiseOp: divideOp` for each tensor chunk                |
| Divide             | Tensor       | Tensor        | Submit chunks with `matrixOp: divideTensorsOp` for aligned chunks              |
| Not                | N/A          | Tensor        | Submit chunks with `unaryOp: notOp` for each input chunk                       |
| And                | Scalar       | Tensor        | Submit chunks with `pairwiseOp: andOp` for each tensor chunk                   |
| Or                 | Scalar       | Tensor        | Submit chunks with `pairwiseOp: orOp` for each tensor chunk                    |
| EqualTo            | Scalar       | Tensor        | Submit chunks with `pairwiseOp: equalOp` for each tensor chunk                 |
| GreaterThan        | Scalar       | Tensor        | Submit chunks with `pairwiseOp: gtOp` for each tensor chunk                    |
| LessThan           | Scalar       | Tensor        | Submit chunks with `pairwiseOp: ltOp` for each tensor chunk                    |
| GreaterThanOrEqual | Scalar       | Tensor        | Submit chunks with `pairwiseOp: gteOp` for each tensor chunk                   |
| LessThanOrEqual    | Scalar       | Tensor        | Submit chunks with `pairwiseOp: lteOp` for each tensor chunk                   |
| MaskFromData       | N/A          | Tensor        | Submit single chunk with `unaryOp: maskOp` (entire tensor as one chunk)        |

**Key Generation Pattern:**

For each operation, the synthesized code:

1. Applies the `ChunkingStrategy` to divide input tensors into chunks
2. For each chunk pair (or single chunk for unary operations):
   - Creates operand objects referencing the chunk's slice of the tensor
   - Creates a `ChronoconeTensorComputationChunk` with the operation function
   - Calls `controller.SubmitChunk(chunk)` to enqueue for execution
3. The `SubmitChunk` call blocks if all workers are busy (channel full), providing backpressure
4. Workers receive chunks and execute the operation function

**Example: Generated Code for Tensor-Tensor Addition**

For the CPL expression `revenue + costs` where both are metrics (tensors):

```go
func computeProfit(
    controller *ChronoconeComputationController,
    inputs map[string]*ChronoconeTensor[float64],
) {
    revenue := inputs["revenue"]
    costs := inputs["costs"]

    // Allocate output tensor
    profit, err := controller.AllocateTensor[float64](revenue.dimensionNames...)
    if err != nil {
        // Handle error
        return
    }

    // Apply chunking strategy to divide tensors into chunks
    revenueChunks := CacheOptimizedChunking(revenue)
    costsChunks := CacheOptimizedChunking(costs)
    profitChunks := CacheOptimizedChunking(profit)

    // Set up synchronization and error handling
    var wg sync.WaitGroup
    errChan := make(chan error, len(revenueChunks))

    // Submit chunks to work channel for parallel execution
    for i := 0; i < len(revenueChunks); i++ {
        wg.Add(1)

        // Create input operands for this chunk
        revenueOperand := NewChronoconeTensorComputationChunkOperand(revenue,
            revenueChunks[i].startIndex, revenueChunks[i].length)
        costsOperand := NewChronoconeTensorComputationChunkOperand(costs,
            costsChunks[i].startIndex, costsChunks[i].length)

        // Create output operand for this chunk
        profitOperand := NewChronoconeTensorComputationChunkOperand(profit,
            profitChunks[i].startIndex, profitChunks[i].length)

        // Create computation chunk with matrix operation
        chunk := &ChronoconeTensorComputationChunk[float64]{
            operands:   []*ChronoconeTensorComputationChunkOperand{revenueOperand, costsOperand},
            matrixOp:   addTensorsOp,  // MatrixOperation function
            pairwiseOp: nil,
            unaryOp:    nil,
            output:     profitOperand,
            waitGroup:  &wg,
            errorChan:  errChan,
        }

        // Submit to work channel (blocks if channel buffer is full)
        controller.SubmitChunk(chunk)
    }

    // Wait for all chunks to complete
    wg.Wait()
    close(errChan)

    // Check for errors
    for err := range errChan {
        if err != nil {
            // Handle error (e.g., return or log)
            return
        }
    }

    // Store result tensor reference for dependent computations
    controller.GetMetricComputation("profit").SetTensor(profit)

    // Release input tensors
    controller.GetMetricComputation("revenue").ReleaseTensor()
    controller.GetMetricComputation("costs").ReleaseTensor()
}
```

**Worker Execution Flow:**

When a worker goroutine receives the chunk:

1. It checks which operation field is non-nil (`matrixOp`, `pairwiseOp`, or `unaryOp`)
2. For matrix operations (like `addTensorsOp`):
   - Calls `chunk.matrixOp(chunk.operands, chunk.output)`
   - The operation function iterates over the chunk's data slice
   - Writes results to the output operand's tensor slice
3. The worker then processes the next chunk from the channel

This design decouples the computation definition (CIL) from the execution mechanism (work channel + workers), allowing the same generated code to run on different hardware configurations by simply adjusting the core count in the IO specification.

#### Operation Function Definitions

The operation functions define the element-wise behavior for each CIL operation. These low-level functions (Level 3 in the hierarchy) take `ChronoconeTensorComputationChunkOperand` objects as inputs, perform the operation element-wise on the chunk data, and write results to the output operand. All functions return only an error result.

See [Three-Level Function Hierarchy](tensor-representation.md#three-level-function-hierarchy) for the complete design.

**Unary Operations:**

Unary operations take a single `ChronoconeTensorComputationChunkOperand` as input and write results to an output `ChronoconeTensorComputationChunkOperand`.

```go
// negateOp negates each element of the input chunk and stores the result in the output chunk.
// Parameters:
//   - input: The input chunk operand containing values to negate
//   - output: The output chunk operand where results will be written
// Returns:
//   - An error if the operation fails
func negateOp[T Numeric](
    input *ChronoconeTensorComputationChunkOperand[T],
    output *ChronoconeTensorComputationChunkOperand[T],
) error {
    for i := 0; i < input.length; i++ {
        output.tensor.data[output.startIndex+i] = -input.tensor.data[input.startIndex+i]
    }
    return nil
}
```

**Binary Operations:**

Binary operations have two variants:

1. **Scalar-Tensor Operations**: The left operand is a primitive scalar value and the right operand is a `ChronoconeTensorComputationChunkOperand`. The scalar value is combined with each element of the right-hand chunk, and results are written to the output chunk.

2. **Tensor-Tensor Operations**: Both operands are `ChronoconeTensorComputationChunkOperand` instances referencing chunks of the same shape. The operation is performed element-wise, combining corresponding elements from both chunks.

**Scalar-Tensor Operations:**

```go
// addOp adds the left scalar value to each element of the right chunk.
// Parameters:
//   - left: The scalar value to add (any primitive numeric type)
//   - right: The input chunk operand
//   - output: The output chunk operand where results will be written
// Returns:
//   - An error if the operation fails
func addOp[T Numeric](
    left T,
    right *ChronoconeTensorComputationChunkOperand[T],
    output *ChronoconeTensorComputationChunkOperand[T],
) error {
    for i := 0; i < right.length; i++ {
        output.tensor.data[output.startIndex+i] = left + right.tensor.data[right.startIndex+i]
    }
    return nil
}

// subtractOp subtracts each element of the right chunk from the left scalar value.
// Parameters:
//   - left: The scalar value to subtract from (any primitive numeric type)
//   - right: The input chunk operand to subtract
//   - output: The output chunk operand where results will be written
// Returns:
//   - An error if the operation fails
func subtractOp[T Numeric](
    left T,
    right *ChronoconeTensorComputationChunkOperand[T],
    output *ChronoconeTensorComputationChunkOperand[T],
) error {
    for i := 0; i < right.length; i++ {
        output.tensor.data[output.startIndex+i] = left - right.tensor.data[right.startIndex+i]
    }
    return nil
}

// multiplyOp multiplies the left scalar value with each element of the right chunk.
// Parameters:
//   - left: The scalar value to multiply (any primitive numeric type)
//   - right: The input chunk operand
//   - output: The output chunk operand where results will be written
// Returns:
//   - An error if the operation fails
func multiplyOp[T Numeric](
    left T,
    right *ChronoconeTensorComputationChunkOperand[T],
    output *ChronoconeTensorComputationChunkOperand[T],
) error {
    for i := 0; i < right.length; i++ {
        output.tensor.data[output.startIndex+i] = left * right.tensor.data[right.startIndex+i]
    }
    return nil
}

// divideOp divides the left scalar value by each element of the right chunk.
// Parameters:
//   - left: The scalar dividend (any primitive numeric type)
//   - right: The input chunk operand containing divisors
//   - output: The output chunk operand where results will be written
// Returns:
//   - An error if division by zero occurs or the operation fails
func divideOp[T Numeric](
    left T,
    right *ChronoconeTensorComputationChunkOperand[T],
    output *ChronoconeTensorComputationChunkOperand[T],
) error {
    for i := 0; i < right.length; i++ {
        if right.tensor.data[right.startIndex+i] == 0 {
            return fmt.Errorf("division by zero at index %d", right.startIndex+i)
        }
        output.tensor.data[output.startIndex+i] = left / right.tensor.data[right.startIndex+i]
    }
    return nil
}

// Tensor-Tensor Binary Operations
//
// The following operations perform element-wise binary operations where both
// operands are ChronoconeTensorComputationChunkOperand instances. These are
// used when both inputs to a CIL binary operation are metrics or expressions
// that produce tensors.

// addTensorsOp adds two chunks element-wise: output[i] = left[i] + right[i]
// Parameters:
//   - left: The left input chunk operand
//   - right: The right input chunk operand
//   - output: The output chunk operand where results will be written
// Returns:
//   - An error if chunk dimensions do not match or the operation fails
func addTensorsOp[T Numeric](
    left *ChronoconeTensorComputationChunkOperand[T],
    right *ChronoconeTensorComputationChunkOperand[T],
    output *ChronoconeTensorComputationChunkOperand[T],
) error {
    for i := 0; i < left.length; i++ {
        output.tensor.data[output.startIndex+i] =
            left.tensor.data[left.startIndex+i] + right.tensor.data[right.startIndex+i]
    }
    return nil
}

// subtractTensorsOp subtracts right chunk from left chunk element-wise: output[i] = left[i] - right[i]
// Parameters:
//   - left: The left input chunk operand (minuend)
//   - right: The right input chunk operand (subtrahend)
//   - output: The output chunk operand where results will be written
// Returns:
//   - An error if chunk dimensions do not match or the operation fails
func subtractTensorsOp[T Numeric](
    left *ChronoconeTensorComputationChunkOperand[T],
    right *ChronoconeTensorComputationChunkOperand[T],
    output *ChronoconeTensorComputationChunkOperand[T],
) error {
    for i := 0; i < left.length; i++ {
        output.tensor.data[output.startIndex+i] =
            left.tensor.data[left.startIndex+i] - right.tensor.data[right.startIndex+i]
    }
    return nil
}

// multiplyTensorsOp multiplies two chunks element-wise: output[i] = left[i] * right[i]
// Parameters:
//   - left: The left input chunk operand
//   - right: The right input chunk operand
//   - output: The output chunk operand where results will be written
// Returns:
//   - An error if chunk dimensions do not match or the operation fails
func multiplyTensorsOp[T Numeric](
    left *ChronoconeTensorComputationChunkOperand[T],
    right *ChronoconeTensorComputationChunkOperand[T],
    output *ChronoconeTensorComputationChunkOperand[T],
) error {
    for i := 0; i < left.length; i++ {
        output.tensor.data[output.startIndex+i] =
            left.tensor.data[left.startIndex+i] * right.tensor.data[right.startIndex+i]
    }
    return nil
}

// divideTensorsOp divides left chunk by right chunk element-wise: output[i] = left[i] / right[i]
// Parameters:
//   - left: The left input chunk operand (dividend)
//   - right: The right input chunk operand (divisor)
//   - output: The output chunk operand where results will be written
// Returns:
//   - An error if division by zero occurs, chunk dimensions do not match, or the operation fails
func divideTensorsOp[T Numeric](
    left *ChronoconeTensorComputationChunkOperand[T],
    right *ChronoconeTensorComputationChunkOperand[T],
    output *ChronoconeTensorComputationChunkOperand[T],
) error {
    for i := 0; i < left.length; i++ {
        if right.tensor.data[right.startIndex+i] == 0 {
            return fmt.Errorf("division by zero at index %d", right.startIndex+i)
        }
        output.tensor.data[output.startIndex+i] =
            left.tensor.data[left.startIndex+i] / right.tensor.data[right.startIndex+i]
    }
    return nil
}
```

**Boolean Operations:**

```go
func notOp(input bool) bool {
    return !input
}

func andOp(left, right bool) bool {
    return left && right
}

func orOp(left, right bool) bool {
    return left || right
}
```

**Comparison Operations:**

```go
func equalOp[T comparable](left, right T) bool {
    return left == right
}

func gtOp[T Ordered](left, right T) bool {
    return left > right
}

func ltOp[T Ordered](left, right T) bool {
    return left < right
}

func gteOp[T Ordered](left, right T) bool {
    return left >= right
}

func lteOp[T Ordered](left, right T) bool {
    return left <= right
}
```

**Mask Formation:**

```go
func maskOp[T any](input T) bool {
    // Converts data values to mask values based on type-specific rules
    // Zero values are typically invalid (false), non-zero are valid (true)
    return !isZeroValue(input)
}
```

#### Generated Code Examples

**Example 1: Simple Addition**

For the CPL expression `revenue + bonus`, the generated Go code is:

```go
// Generated computation code
func computeTotalCompensation(
    controller *ChronoconeComputationController,
    inputs map[string]*ChronoconeTensor[float64],
) {
    revenue := inputs["revenue"]
    bonus := inputs["bonus"]

    // Perform addition using high-level tensor operation
    // Under the covers, this uses apply_binary_tensors_to_tensor with addTensorsOp
    totalComp, err := revenue.Add(controller, bonus)
    if err != nil {
        // Handle error
    }

    controller.GetMetricComputation("total_compensation").SetTensor(totalComp)

    // Release input tensors
    controller.GetMetricComputation("revenue").ReleaseTensor()
    controller.GetMetricComputation("bonus").ReleaseTensor()
}
```

**Example 2: Complex Expression**

For the CPL expression `1.2 * (inventory - 8)`, the generated Go code is:

```go
// Generated computation code
func computeAdjustedInventory(
    controller *ChronoconeComputationController,
    inputs map[string]*ChronoconeTensor[float64],
) {
    inventory := inputs["inventory"]

    // First operation: inventory - 8
    // SubtractScalar internally uses apply_binary_to_tensor with subtractOp
    temp1, err := inventory.SubtractScalar(controller, 8.0)
    if err != nil {
        // Handle error
    }

    // Second operation: 1.2 * temp1
    // MultiplyScalar internally uses apply_binary_to_tensor with multiplyOp
    result, err := temp1.MultiplyScalar(controller, 1.2)
    if err != nil {
        // Handle error
    }

    controller.GetMetricComputation("adjusted_inventory").SetTensor(result)

    // Release input tensors
    controller.GetMetricComputation("inventory").ReleaseTensor()
}
```

**Example 3: Tensor-Tensor Addition**

For the CPL expression `revenue + costs` where both operands are metrics (tensors), the generated Go code uses the tensor-tensor variant:

```go
// Generated computation code
func computeProfit(
    controller *ChronoconeComputationController,
    inputs map[string]*ChronoconeTensor[float64],
) {
    revenue := inputs["revenue"]
    costs := inputs["costs"]

    // Perform element-wise addition: revenue + costs
    // Add internally uses apply_binary_tensors_to_tensor with addTensorsOp
    profit, err := revenue.Add(controller, costs)
    if err != nil {
        // Handle error
    }

    controller.GetMetricComputation("profit").SetTensor(profit)
}
```

**Example 4: Custom Chunking Strategy**

For operations requiring specific chunking behavior, the tensor's chunk size can be set before operations:

```go
// Generated computation with custom chunking
func computeLargeTensorOperation(
    controller *ChronoconeComputationController,
    inputs map[string]*ChronoconeTensor[float64],
) {
    largeTensor := inputs["large_data"]

    // Set custom chunk size for this tensor
    // All subsequent operations on this tensor will use this chunk size
    largeTensor.SetChunkSize(1024 * 1024) // 1M elements per chunk

    // Perform complex transformation
    // The operation will use the tensor's configured chunk size
    result, err := largeTensor.Transform(controller, complexTransformOp)
    if err != nil {
        // Handle error
    }

    controller.GetMetricComputation("transformed_data").SetTensor(result)

    // Release input tensors
    controller.GetMetricComputation("large_data").ReleaseTensor()
}
```

### Generation Process

The code generation process follows these steps:

1. **Tree Traversal**: Walking the [`CILAbstractNode`](../go-analysis/cil-representation.md) tree in post-order (children before parents).

2. **Parallel Operation Group Creation**: For each metric or connector, creating a [`ParallelOperationGroup`](#paralleloperationgroup).

3. **Computation Group Management**: Creating new [`ComputationGroup`](#computationgroup) instances when calculation properties (units, measures, dimensions) change.

4. **Code Snippet Generation**: For each node, generating Go code snippets that:
   - Load input tensors or create constants
   - Apply operations using the synthesizer support library
   - Push snippets onto the current [`Computation`](#computation) stack

5. **Dependency Tracking**: Recording metric references to establish cross-metric dependencies.

6. **Function Assembly**: Assembling code snippets from the [`Computation`](#computation) stack into complete functions, processing in bottom-to-top order to respect dependencies.

7. **Setup/Compute/Cleanup Generation**: Generating the three lifecycle functions that orchestrate execution.

8. **Parallelization**: Generating goroutine launches for each [`ParallelOperationGroup`](#paralleloperationgroup) and channel-based synchronization using [`MetricComputation`](#metriccomputation).

9. **Memory Management**: Generating code that reuses tensor memory where possible to minimize allocations.

---

## Type Relationships

The following diagram shows the relationships between the types defined in this document:

### Synthesis Type Hierarchy

```
ChronoconeComputationController
    ├── conceptValues: map[string][]interface{}
    ├── metricComputations: map[string]*MetricComputation
    │       └── MetricComputation
    │               ├── done: chan struct{}
    │               └── tensor: *ChronoconeTensor[any]
    ├── workChannel: chan *ChronoconeTensorComputationChunk[any]
    ├── workerWaitGroup: sync.WaitGroup
    └── numWorkers: int

ChronoconeTensorComputationChunk (submitted to workChannel)
    ├── operands: []*ChronoconeTensorComputationChunkOperand
    │       └── ChronoconeTensorComputationChunkOperand
    │               ├── startIndex: int
    │               ├── length: int
    │               ├── mask: []uint64
    │               └── tensor: *ChronoconeTensor[T]
    ├── matrixOp: MatrixOperation[T] (optional)
    ├── pairwiseOp: PairwiseOperation[L,T] (optional)
    ├── unaryOp: UnaryOperation[T] (optional)
    ├── output: *ChronoconeTensorComputationChunkOperand
    ├── waitGroup: *sync.WaitGroup
    └── errorChan: chan<- error

CILAbstractTransformer
    ├── parallelGroups: []*ParallelOperationGroup
    │       └── ParallelOperationGroup
    │               ├── computationGroups: []*ComputationGroup
    │               │       └── ComputationGroup
    │               │               ├── computations: []*Computation
    │               │               │       └── Computation
    │               │               │               └── snippets: []string
    │               │               ├── unitExpression: string
    │               │               ├── measureName: string
    │               │               └── dimensionNames: []string
    │               ├── name: string
    │               └── dependencies: []string
    ├── measureTypes: map[string]string
    └── metricTypes: map[string]string
```

The hierarchical organization of [`ParallelOperationGroup`](#paralleloperationgroup), [`ComputationGroup`](#computationgroup), and [`Computation`](#computation) enables the transformer to capture both data dependencies and parallelization opportunities during code generation.

The work dispatch mechanism connects the CIL transformation to actual execution:

1. **CILAbstractTransformer** generates code that creates `ChronoconeTensorComputationChunk` objects
2. **Generated Code** calls `controller.SubmitChunk()` to send chunks to the work channel
3. **Worker Goroutines** (equal to core count) receive chunks from the channel and execute operations
4. **ChronoconeTensorComputationChunk** encapsulates the operation (matrix, pairwise, or unary) and its operands
5. **ChronoconeTensorComputationChunkOperand** provides a view into a contiguous slice of tensor data with an associated validity mask

This architecture separates computation definition (CIL transformation) from computation execution (worker pool), enabling hardware-independent code generation with runtime-configurable parallelism.

---

## IO Specification Handling

### Overview

The synthesized Go program accepts an IO specification via a command-line parameter. This JSON file defines the inputs (connectors with data sources), outputs (metrics with dimensions and types), and the target architecture configuration.

### IO Specification Structure

The IO specification is defined by the following Go structs that map to the JSON format:

```go
// IOSpecification represents the complete IO configuration for a computation.
type IOSpecification struct {
    Inputs       []InputSpecification       `json:"inputs"`
    Outputs      []OutputSpecification      `json:"outputs"`
    Architecture ArchitectureSpecification  `json:"architecture"`
}

// InputSpecification defines a connector input with its data source.
type InputSpecification struct {
    ConnectorName      string                 `json:"connector"`
    DataFormat         string                 `json:"data_format"`
    Driver             string                 `json:"driver"`
    DriverParameters   []map[string]interface{} `json:"driver_parameters"`
}

// OutputSpecification defines a metric output with its dimensions and type.
type OutputSpecification struct {
    MetricName   string   `json:"metric"`
    Dimensions   []string `json:"dimensions"`
    CellType     string   `json:"celltype"`
    Grain        string   `json:"grain"`
}

// ArchitectureSpecification defines the target compute resources.
type ArchitectureSpecification struct {
    CPUCoreCount     int `json:"cpu_cores"`
    CacheSizeBytes   int `json:"cache_size"`
    MainMemoryBytes  int `json:"main_memory_size"`
}
```

### Reading the IO Specification

The synthesized program reads the IO specification from a JSON file path provided as a command-line argument:

```go
// LoadIOSpecification reads and parses the IO specification from a JSON file.
// Parameters:
//   - filePath: Path to the JSON IO specification file
// Returns:
//   - A pointer to the parsed IOSpecification
//   - An error if the file cannot be read or parsed
func LoadIOSpecification(filePath string) (*IOSpecification, error) {
    data, err := os.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read IO specification file: %w", err)
    }

    var spec IOSpecification
    if err := json.Unmarshal(data, &spec); err != nil {
        return nil, fmt.Errorf("failed to parse IO specification JSON: %w", err)
    }

    return &spec, nil
}
```

### Generated Main Function with IO Specification

The generated main function accepts the IO specification file path as a command-line parameter and uses it to configure the computation. The controller is configured via `SetArchitectureConfig`, which initializes the work channel with a buffer size equal to the core count, and worker goroutines are launched to process computation chunks:

```go
func main() {
    // Check command-line arguments
    if len(os.Args) < 2 {
        log.Fatal("Usage: <program> <io_specification.json>")
    }

    ioSpecPath := os.Args[1]

    // Load the IO specification
    ioSpec, err := LoadIOSpecification(ioSpecPath)
    if err != nil {
        log.Fatalf("Failed to load IO specification: %v", err)
    }

    // Initialize the computation controller
    controller := NewChronoconeComputationController()

    // Configure with architecture settings from IO spec
    // This initializes the work channel with buffer size = core count
    controller.SetArchitectureConfig(&ioSpec.Architecture)

    // Setup phase: register all metrics and prepare data structures
    if err := Setup(controller); err != nil {
        log.Fatalf("Setup failed: %v", err)
    }

    // Load input tensors from specified connectors
    inputs := make(map[string]*ChronoconeTensor[float64])
    for _, inputSpec := range ioSpec.Inputs {
        tensor, err := loadConnectorData(controller, &inputSpec)
        if err != nil {
            log.Fatalf("Failed to load input %s: %v", inputSpec.ConnectorName, err)
        }
        inputs[inputSpec.ConnectorName] = tensor
    }

    // Start worker goroutines that listen on the work channel
    // Workers will process ChronoconeTensorComputationChunk objects until the channel is closed
    controller.StartWorkers()

    // Compute phase: execute all parallel operation groups
    // During this phase, chunks are submitted to the work channel via SubmitChunk()
    if err := Compute(controller, inputs); err != nil {
        log.Fatalf("Compute failed: %v", err)
    }

    // Stop workers: close the work channel and wait for all workers to exit
    // This ensures all pending chunks are processed before cleanup
    controller.StopWorkers()

    // Write output tensors according to output specifications
    for _, outputSpec := range ioSpec.Outputs {
        if err := writeMetricOutput(controller, &outputSpec); err != nil {
            log.Fatalf("Failed to write output %s: %v", outputSpec.MetricName, err)
        }
    }

    // Cleanup phase: release resources
    if err := Cleanup(controller); err != nil {
        log.Fatalf("Cleanup failed: %v", err)
    }
}

// loadConnectorData loads tensor data for a connector based on the input specification.
// The driver and driver_parameters from the specification determine how data is loaded.
func loadConnectorData(
    controller *ChronoconeComputationController,
    spec *InputSpecification,
) (*ChronoconeTensor[float64], error) {
    // Dispatch to appropriate driver based on spec.Driver
    switch spec.Driver {
    case "file":
        return loadFileDriver(controller, spec)
    case "chronocone_database":
        return loadDatabaseDriver(controller, spec)
    default:
        return nil, fmt.Errorf("unknown driver: %s", spec.Driver)
    }
}

// writeMetricOutput writes the computed metric tensor according to the output specification.
func writeMetricOutput(
    controller *ChronoconeComputationController,
    spec *OutputSpecification,
) error {
    // Retrieve the computed tensor for the specified metric
    metricComp := controller.GetMetricComputation(spec.MetricName)
    if metricComp == nil {
        return fmt.Errorf("metric not found: %s", spec.MetricName)
    }

    tensor := metricComp.GetTensor()

    // Validate dimensions match the specification
    if !dimensionsMatch(tensor.DimensionNames(), spec.Dimensions) {
        return fmt.Errorf("dimension mismatch for metric %s", spec.MetricName)
    }

    // Write output in the appropriate format
    return writeTensorToOutput(tensor, spec)
}
```

### IO Specification JSON Example

```json
{
  "inputs": [
    {
      "connector": "sales_data",
      "data_format": "parquet",
      "driver": "file",
      "driver_parameters": [{ "path": "/data/sales.parquet" }]
    }
  ],
  "outputs": [
    {
      "metric": "revenue",
      "dimensions": ["Product", "Region", "Time"],
      "celltype": "Currency",
      "grain": "Month"
    }
  ],
  "architecture": {
    "cpu_cores": 8,
    "cache_size": 16777216,
    "main_memory_size": 34359738368
  }
}
```

### Architecture Configuration Integration

The architecture settings from the IO specification influence computation behavior. The `SetArchitectureConfig` method initializes the work channel with a buffer size equal to the core count:

```go
// SetArchitectureConfig configures the controller with target hardware specs.
// This initializes the work channel and worker pool configuration.
// Must be called before StartWorkers().
func (c *ChronoconeComputationController) SetArchitectureConfig(
    config *ArchitectureSpecification,
) {
    c.cpuCoreCount = config.CPUCoreCount
    c.cacheSizeBytes = config.CacheSizeBytes
    c.mainMemoryBytes = config.MainMemoryBytes
    c.numWorkers = config.CPUCoreCount

    // Initialize the work channel with buffer size equal to core count
    // This provides natural backpressure when all workers are busy
    c.workChannel = make(chan *ChronoconeTensorComputationChunk[any], config.CPUCoreCount)
}
```

#### Core Count and Work Channel Sizing

The number of CPU cores from the IO specification directly determines:

1. **Worker Count**: The number of worker goroutines launched equals `CPUCoreCount`
2. **Channel Buffer Size**: The work channel buffer size equals `CPUCoreCount`

This 1:1 ratio provides natural backpressure: when all workers are busy, the channel buffer fills, and subsequent `SubmitChunk()` calls block until a worker becomes available. This prevents memory bloat from unbounded work queues while maximizing CPU utilization.

```go
// Example: Architecture configuration and controller initialization
func main() {
    // ... load IO specification ...
    ioSpec, err := LoadIOSpecification(ioSpecPath)
    if err != nil {
        log.Fatalf("Failed to load IO specification: %v", err)
    }

    // Create controller and configure with architecture settings
    controller := NewChronoconeComputationController()
    controller.SetArchitectureConfig(&ioSpec.Architecture)

    // Setup phase: register metrics, allocate tensors
    if err := Setup(controller); err != nil {
        log.Fatalf("Setup failed: %v", err)
    }

    // Start worker goroutines
    controller.StartWorkers()

    // ... computation phase submits chunks to work channel ...

    // Cleanup phase: close channel and wait for workers
    controller.StopWorkers()
}
```
