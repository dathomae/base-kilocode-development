## Tensor Representation In The Synthesizer Support Library

Data in Chronocone metrics is handled using tensors within the library that supports the generated Go code.

---

## Conceptual Model

### Overview

From a conceptual standpoint, a tensor is a multi-dimensional array as described in the [Calculation/Operations Framework In Chronocone](../../architecture/cpl/analysis/intermediate_language.md). The tensor abstraction provides the foundation for representing metric data across multiple conceptual dimensions.

### Tensor Structure

A ChronoconeTensor consists of two conceptual components:

1. **Dimension Names**: An ordered list of concept names that define the tensor's dimensions
2. **Data Tensor**: The structure containing the actual metric data values

Conceptually, every tensor has a validity mask. However, due to CPL semantics where every boolean must be fully specified, boolean formulas always start with masks that are entirely valid and end with masks that are entirely valid. The partially-valid state only exists as a temporary state during intermediate calculations.

### Array Arithmetic and Indexing

To locate a specific element within a tensor, array arithmetic is used based on the dimension sizes and the indices of interest.

For example, given a tensor with dimensions Country, Region, and Product, with 3 Country values, 2 Region values, and 10 Product values, the offset for the element at Country index 1, Region index 0, and Product index 2 is calculated as:

    offset = 1*2*10 + 0*10 + 2 = 20 + 0 + 2 = 22

#### General Stride Formula

For a tensor with dimensions `[D₀, D₁, ..., Dₙ₋₁]` where each dimension Dᵢ has size `size(Dᵢ)`, the stride for each dimension is calculated as the product of the sizes of all dimensions below it:

    stride(Dᵢ) = ∏(size(Dⱼ)) for j from i+1 to n-1

The stride for the innermost dimension (last in the order) is always 1.

Given a set of indices `[i₀, i₁, ..., iₙ₋₁]` corresponding to each dimension, the offset into the flat array is:

    offset = Σ(iₖ × stride(Dₖ)) for k from 0 to n-1

For example, with dimension order `[Country, Region, Product]` and sizes `[3, 2, 10]`:

- stride(Country) = 2 × 10 = 20
- stride(Region) = 10
- stride(Product) = 1

The offset for indices `[1, 0, 2]` is:

    offset = (1 × 20) + (0 × 10) + (2 × 1) = 22

### Dimension Ordering and Memory Layout

The order of dimensions in a tensor has significant implications for memory layout and cache efficiency. Since tensor data is stored in a single contiguous array, the dimension order determines which values are adjacent in memory.

#### Cache Efficiency Considerations

Values that are accessed together should ideally be stored contiguously to maximize cache utilization. Consider a tensor with dimensions `[Product, Time, Country]` where:

- Product has 1000 values
- Time has 26 values (13 actuals + 13 forecasts)
- Country has 3 values

**Row-Major (Product-Time-Country order):**
In this layout, all Country values for a given Product-Time combination are contiguous:

    [P0,T0,C0], [P0,T0,C1], [P0,T0,C2], [P0,T1,C0], [P0,T1,C1], ...

This layout is efficient when aggregating across Country (the innermost dimension), as all values to be summed are contiguous in memory.

**Alternative Order (Time-Product-Country):**
If we pivot to `[Time, Product, Country]` order:

    [T0,P0,C0], [T0,P0,C1], [T0,P0,C2], [T0,P1,C0], [T0,P1,C1], ...

This layout is efficient when performing time-series operations across all products for a given time period.

#### Dimension Order and Aggregators/Spreaders

The dimension ordering is particularly important for aggregators and spreaders because these operations perform matrix operations on the bottom two levels of the tensor:

1. **Aggregators** collapse the lowest dimension by multiplying with a transformation matrix on the left. For efficient aggregation, the dimension being aggregated should be the innermost (lowest) dimension.

2. **Spreaders** expand by multiplying with a transformation matrix on the right. The dimension being spread should also be positioned at the lowest level.

When a dimension needs to be aggregated or spread but is not in the correct position, a `pivot` operation is performed first to reorder dimensions. A pivot that swaps the bottom two dimensions is equivalent to a transpose operation.

For example, to aggregate Country when dimensions are `[Product, Time, Country]`:

```
matrix_operation([Product, Time], sum, [1 1 1], tensor)
```

If dimensions were `[Country, Product, Time]`, a pivot would be required first:

```
pivot([Product, Time, Country], tensor)
matrix_operation([Product, Time], sum, [1 1 1], pivoted_tensor)
```

#### Demand-Driven Dimension Order Propagation

Because pivoting a large tensor requires an out-of-place full data copy, it is an expensive operation that should be minimized. The `CILAbstractTransformer` employs a demand-driven strategy to optimize dimension ordering and minimize physical pivots:

1. **Demand Generation**: Aggregators and spreaders "demand" that their target dimension is the innermost dimension to maximize cache efficiency during matrix operations. Output metrics "demand" the dimension ordering specified in the IO Specification.
2. **Demand Propagation**: Element-wise operations (Add, Multiply, etc.) do not inherently care about dimension ordering, but they require all their operands to have the _same_ ordering. Therefore, an element-wise operation adopts the dimension ordering demanded by its consumer, and propagates this demand down to its children.
3. **Conflict Resolution & Pivot Injection**: If a tensor is consumed by multiple operations that demand _different_ dimension orderings, a physical `pivot` operation is unavoidable for at least one path. The transformer resolves this by choosing one ordering as the "primary" layout for the tensor (e.g., based on the most computationally expensive consumer), and explicitly injects a `pivot` node into the AST for the conflicting paths. If an input connector provides data in a fixed order that differs from the demanded order, a `pivot` is injected immediately after loading.

_Future Optimization Note: While physical pivoting is the default strategy for cache efficiency, future optimizations could introduce "strided" aggregators. A strided aggregator would perform the math without a physical copy, which might be faster for simple, one-off aggregations where the cost of the full copy outweighs the cache miss penalty. For now, the design relies on physical pivots to keep the low-level worker operations simple and cache-optimized._

### Parallelization and Chunking

Tensor operations can be parallelized by dividing data tensors into chunks. In general, chunks do not need to align with tensor dimensions. However, dimension alignment is required in two specific cases:

1. When performing aggregation or spread operations that depend on the row/column structure
2. When dimension/concept values are passed to a metric and must align with data values

---

## Go Implementation Details

### Library Structure

The tensor operations library is implemented as an ordinary Go package that the synthesized Go program calls. It is not generated at synthesize time.

### Physical Storage

Each tensor is stored as a single-dimensional Go slice. This design choice supports Chronocone's performance strategy:

- **Data Locality**: Go slices are backed by physically contiguous arrays, providing the necessary control to ensure cache locality during operations
- **Memory Layout**: The underlying arrays guarantee contiguous memory allocation, which is critical for high-performance numerical operations

### ChronoconeTensor[T] Struct

The Go representation of a tensor is defined by the generic `ChronoconeTensor[T]` struct:

```go
type ChronoconeTensor[T any] struct {
    data            []T           // Flat slice of data values backed by a contiguous array
    dimensionNames  []string      // Ordered list of concept names
    effectiveLength int           // Product of dimension sizes
    physicalLength  int           // Actual allocated array length
    chunkSize       int           // Parallelization granularity
}
```

**Field Descriptions:**

- **data**: A flat array storing all tensor values contiguously. The generic type parameter `T` allows the same structure to be used for any data type (integers, floats, etc.).

- **dimensionNames**: An ordered list of concept names defining the tensor's dimensions. The order determines the memory layout via the stride calculation.

- **effectiveLength**: The product of the number of values for each dimension. This represents the logical size of the tensor and may be smaller than the physical array length.

- **physicalLength**: The actual allocated length of the data array. Separating effective length from physical length allows operations to expand and contract the logical tensor size without requiring array reallocation and data copying.

- **chunkSize**: Controls the granularity of parallelization. See [Computation Chunking](#computation-chunking) for details.

Note: The validity mask has been moved to `ChronoconeTensorComputationChunkOperand` because partially-valid masks only occur as temporary states during intermediate calculations. Complete tensors always have fully valid data.

#### ChronoconeTensor Methods

```go
// SetChunkSize sets the chunk size for parallelization of operations on this tensor.
// The chunk size determines how the tensor is divided for parallel processing.
// Parameters:
//   - size: The number of elements per chunk
func (t *ChronoconeTensor[T]) SetChunkSize(size int)

// GetChunkSize returns the current chunk size for this tensor.
func (t *ChronoconeTensor[T]) GetChunkSize() int
```

**High-Level Tensor Operations:**

The following methods operate on whole tensors and internally handle chunking using the tensor's `chunkSize` attribute. These are the functions that synthesized code calls directly.

```go
// Add performs element-wise addition of two tensors: result = left + right
// Both tensors must have the same dimensions. The operation is parallelized
// based on the left tensor's chunkSize.
func (t *ChronoconeTensor[T]) Add(
    controller *ChronoconeComputationController,
    right *ChronoconeTensor[T],
) (*ChronoconeTensor[T], error)

// Subtract performs element-wise subtraction: result = left - right
// Both tensors must have the same dimensions.
func (t *ChronoconeTensor[T]) Subtract(
    controller *ChronoconeComputationController,
    right *ChronoconeTensor[T],
) (*ChronoconeTensor[T], error)

// Multiply performs element-wise multiplication: result = left * right
// Both tensors must have the same dimensions.
func (t *ChronoconeTensor[T]) Multiply(
    controller *ChronoconeComputationController,
    right *ChronoconeTensor[T],
) (*ChronoconeTensor[T], error)

// Divide performs element-wise division: result = left / right
// Both tensors must have the same dimensions.
func (t *ChronoconeTensor[T]) Divide(
    controller *ChronoconeComputationController,
    right *ChronoconeTensor[T],
) (*ChronoconeTensor[T], error)

// AddScalar performs scalar-tensor addition: result = scalar + tensor
func (t *ChronoconeTensor[T]) AddScalar(
    controller *ChronoconeComputationController,
    scalar T,
) (*ChronoconeTensor[T], error)

// SubtractScalar performs scalar-tensor subtraction: result = scalar - tensor
func (t *ChronoconeTensor[T]) SubtractScalar(
    controller *ChronoconeComputationController,
    scalar T,
) (*ChronoconeTensor[T], error)

// MultiplyScalar performs scalar-tensor multiplication: result = scalar * tensor
func (t *ChronoconeTensor[T]) MultiplyScalar(
    controller *ChronoconeComputationController,
    scalar T,
) (*ChronoconeTensor[T], error)

// DivideScalar performs scalar-tensor division: result = scalar / tensor
func (t *ChronoconeTensor[T]) DivideScalar(
    controller *ChronoconeComputationController,
    scalar T,
) (*ChronoconeTensor[T], error)

// Negate performs element-wise negation: result = -tensor
func (t *ChronoconeTensor[T]) Negate(
    controller *ChronoconeComputationController,
) (*ChronoconeTensor[T], error)
```

These high-level methods mirror the CIL operations (`addOp`, `subtractOp`, `multiplyOp`, etc.) but operate at the tensor level. Under the covers, they:

1. Use the tensor's `chunkSize` to divide the operation into chunks
2. Create `ChronoconeTensorComputationChunkOperand` objects for each chunk
3. Submit work to the controller's work channel for parallel execution
4. Return the resulting tensor once all chunks complete

See [Three-Level Function Hierarchy](#three-level-function-hierarchy) for the complete design.

### Array Capacity Management

While the number of values for each dimension determines the minimum array size, the actual allocated size is stored separately. This allows operations to expand and contract the logical tensor size without requiring array reallocation.

When an operation needs to grow a tensor:

1. If `physicalLength >= newEffectiveLength`, simply update `effectiveLength`
2. If `physicalLength < newEffectiveLength`, allocate new arrays and copy data

This optimization is particularly valuable for iterative algorithms where tensors grow gradually.

### Generic Data Types

The data tensor uses a generic type parameter. The synthesized code determines the concrete data type based on:

- The metric's data type
- The constraints defined for that metric

### Mask Tensor Implementation

Although the mask tensor is conceptually boolean-valued, the implementation uses an array of `uint64` for efficiency:

- **Word Location**: Found by integer division of the offset by 64: `wordIndex = offset / 64`
- **Bit Location**: Found by taking the offset modulo 64: `bitIndex = offset % 64`
- **Bit Values**: A bit value of 0 indicates valid data, 1 indicates masked/invalid data

#### Mask Bit Convention

**Important**: The mask uses an inverted bit convention where **0 = valid** and **1 = invalid**.

This convention may seem counterintuitive (one might expect 1 to mean "valid"), but it was chosen deliberately for performance reasons:

1. **Zero-initialization means all valid**: In Go, newly allocated arrays are zero-initialized. With this convention, a freshly created tensor automatically has all values marked as valid without any additional initialization code. This eliminates the need to explicitly set all mask bits to valid during tensor creation, which would require O(n) operations on the mask array.

2. **Common case optimization**: In typical computations, most data values are valid. The common operations are checking validity (reading) and marking specific values as invalid (writing). Reading a bit to check validity is equally fast regardless of the convention, but avoiding the initial O(n) write to set all values as valid provides a significant performance benefit when creating tensors.

3. **Cache efficiency**: Eliminating the initialization pass means the mask array starts cold in cache and is only touched when specific values need to be marked invalid, reducing memory bandwidth usage.

#### Mask Operations

The following operations are provided for mask manipulation on `ChronoconeTensorComputationChunkOperand`:

**Check if Mask is True (Valid):**

```go
func (c *ChronoconeTensorComputationChunkOperand[T]) IsValid(offset int) bool {
    wordIndex := offset >> 6        // Divide by 64
    bitIndex := offset & 0x3F       // Modulo 64
    return (c.mask[wordIndex] & (1 << bitIndex)) == 0
}
```

**Set Mask to False (Invalid):**

```go
func (c *ChronoconeTensorComputationChunkOperand[T]) SetInvalid(offset int) {
    wordIndex := offset >> 6        // Divide by 64
    bitIndex := offset & 0x3F       // Modulo 64
    c.mask[wordIndex] |= (1 << bitIndex)
}
```

**Set Mask to True (Valid):**

```go
func (c *ChronoconeTensorComputationChunkOperand[T]) SetValid(offset int) {
    wordIndex := offset >> 6        // Divide by 64
    bitIndex := offset & 0x3F       // Modulo 64
    c.mask[wordIndex] &^= (1 << bitIndex)
}
```

The intent, and belief, is that these mask operations will be inlined by the Go compiler, making them as efficient as direct bit manipulation operations.

**Creating an Operand Chunk with All Valid Mask:**
When a `ChronoconeTensorComputationChunkOperand` is created, the mask should be initialized to all true (all bits set to 0):

```go
func NewChronoconeTensorComputationChunkOperand[T any](tensor *ChronoconeTensor[T], startIndex, length int) *ChronoconeTensorComputationChunkOperand[T] {
    // Round up to nearest 64 for mask array
    maskLength := (length + 63) / 64

    chunk := &ChronoconeTensorComputationChunkOperand[T]{
        startIndex: startIndex,
        length:     length,
        mask:       make([uint64], maskLength),
        tensor:     tensor,
    }

    // Mask is zero-initialized (all valid)
    return chunk
}
```

```go
func NewChronoconeTensor[T any](controller *ChronoconeComputationController, dimensionNames []string, physicalLength *int) (*ChronoconeTensor[T], error) {
    // Calculate dimension sizes from the controller's concept values
    dimensionSizes := make([]int, len(dimensionNames))
    effectiveLength := 1
    for i, name := range dimensionNames {
        values := controller.conceptValues[name]
        size := len(values)
        dimensionSizes[i] = size
        effectiveLength *= size
    }

    // Determine physical length
    physLen := effectiveLength
    if physicalLength != nil {
        if *physicalLength < effectiveLength {
            return nil, fmt.Errorf("physical length (%d) must be at least the effective length (%d)", *physicalLength, effectiveLength)
        }
        physLen = *physicalLength
    }

    tensor := &ChronoconeTensor[T]{
        data:            make([T], physLen),
        dimensionNames:  dimensionNames,
        effectiveLength: effectiveLength,
        physicalLength:  physLen,
    }

    return tensor, nil
}
```

**Parameter Descriptions:**

- **controller**: Pointer to the `ChronoconeComputationController` that provides access to concept values. The controller is defined in the [synthesis.md](synthesis.md) document.
- **dimensionNames**: Ordered list of concept names defining the tensor's dimensions.
- **physicalLength**: Optional pointer to an int specifying the physical array length. If nil, the physical length equals the effective length. If provided, must be at least the effective length.

**Note on ChronoconeComputationController:** The controller provides centralized access to concept values, allowing tensors to look up dimension sizes from the shared concept registry rather than having sizes passed as separate parameters. This ensures consistency across all tensors in a computation and simplifies the API for tensor creation.

Note: Complete `ChronoconeTensor` instances do not have masks because boolean formulas in CPL always start and end with fully specified (all valid) states. The mask only exists on `ChronoconeTensorComputationChunkOperand` to represent the temporary partially-valid states that occur during intermediate calculations.

### Computation Chunking

Each tensor carries a computation chunk size attribute that controls parallelization granularity. See [Synthesizing Go Code From Chronocone Intermediate Representation](synthesis.md) for a detailed discussion of chunking strategies, parallelization, and the synthesis context for tensor operations.

#### ChronoconeTensorComputationChunkOperand

A computation chunk operand represents a contiguous range of values within a tensor that serves as an input or output to a computation:

```go
type ChronoconeTensorComputationChunkOperand[T any] struct {
    startIndex  int        // Starting offset in the tensor arrays
    length      int        // Number of elements in this chunk
    mask        []uint64   // Bit-packed mask slice for this chunk
    tensor      *ChronoconeTensor[T]  // Reference to parent tensor
}
```

**Field Descriptions:**

- **startIndex**: The starting offset into the data array.
- **length**: The number of contiguous elements this chunk covers.
- **mask**: A bit-packed array of `uint64` values representing the boolean mask for this chunk. The mask indicates which data values in this chunk are valid (0 = valid, 1 = invalid). The mask is sized to cover `length` elements.
- **tensor**: A reference to the parent ChronoconeTensor. The operand does not own the data; it provides a view into the parent tensor's arrays.

#### ChronoconeTensorComputationChunk

A computation chunk represents a unit of computation that operates on one or more input operands and produces output to a destination operand. It encapsulates the operation to be performed and its inputs/outputs:

```go
// MatrixOperation is a function type that performs an element-wise operation
// on a single chunk. It reads from input operands and writes to an output operand.
// All operands extend from ChronoconeTensorComputationChunkOperand.
type MatrixOperation[T any] func(
    inputs []*ChronoconeTensorComputationChunkOperand[T],
    output *ChronoconeTensorComputationChunkOperand[T],
) error

// PairwiseOperation is a function type for binary operations.
// The right parameter extends from ChronoconeTensorComputationChunkOperand,
// the left parameter extends from any.
type PairwiseOperation[L any, T any] func(left L, right *ChronoconeTensorComputationChunkOperand[T]) T

// UnaryOperation is a function type for unary operations on a single input operand.
// The operand extends from ChronoconeTensorComputationChunkOperand.
type UnaryOperation[T any] func(input *ChronoconeTensorComputationChunkOperand[T]) T

type ChronoconeTensorComputationChunk[T any] struct {
    operands   []*ChronoconeTensorComputationChunkOperand[T]  // Input operands for the computation
    matrixOp   MatrixOperation[T]                             // Matrix operation function (nil if not used)
    pairwiseOp PairwiseOperation[T]                           // Pairwise operation function (nil if not used)
    unaryOp    UnaryOperation[T]                              // Unary operation function (nil if not used)
    output     *ChronoconeTensorComputationChunkOperand[T]    // Output location
    waitGroup  *sync.WaitGroup                                // WaitGroup to signal completion
    errorChan  chan<- error                                   // Channel to report errors back to the caller
}
```

**Field Descriptions:**

- **operands**: A list of input operands for the computation. Each operand is a `ChronoconeTensorComputationChunkOperand` that references a contiguous range of values in a tensor.
- **matrixOp**: A `MatrixOperation` function that operates on all input operands to produce output. Function types in Go are nullable (can be nil), so this field does not need to be a pointer. Only one of `matrixOp`, `pairwiseOp`, or `unaryOp` will be non-nil for any given computation chunk.
- **pairwiseOp**: A `PairwiseOperation` function for binary element-wise operations. Function types in Go are nullable (can be nil), so this field does not need to be a pointer. Only one of `matrixOp`, `pairwiseOp`, or `unaryOp` will be non-nil for any given computation chunk.
- **unaryOp**: A `UnaryOperation` function for unary element-wise operations. Function types in Go are nullable (can be nil), so this field does not need to be a pointer. Only one of `matrixOp`, `pairwiseOp`, or `unaryOp` will be non-nil for any given computation chunk.
- **output**: The output operand where computation results are written.
- **waitGroup**: A pointer to a `sync.WaitGroup` that the worker will decrement when the chunk is processed. This allows the caller to wait for all chunks of an operation to complete.
- **errorChan**: A channel used by the worker to report any errors (e.g., division by zero) back to the caller.

#### Chunk Properties

A computation chunk operand always applies to a **physically contiguous** block of values in the ChronoconeTensor. This is a critical design constraint that enables efficient parallel processing:

1. **Data Contiguity**: All data values in the operand are contiguous in memory
2. **Mask Ownership**: Each operand owns its own mask array, which represents the temporary validity state during intermediate calculations
3. **No Data Array Copying**: Operands are views into existing data arrays, never copies

#### Mask-Based Parallel Computation

Boolean operations in CPL expressions result in mask values being set to indicate which branch of the computation applies to which data elements. This enables a powerful parallelization strategy:

1. Each branch of a boolean operation creates a tensor where the mask is set to `true` for entries where that branch's condition evaluates to `true`
2. The tensors associated with each branch can be computed in parallel
3. Computation within each branch only sets values where the corresponding mask value is `true`

For example, consider the expression:

```
if inventory % 1 = 0
    inventory
else
    inventory + 1
```

This generates two computation chunks:

- Chunk 1: Mask = true where `inventory % 1 = 0`, computes `inventory`
- Chunk 2: Mask = true where `inventory % 1 ≠ 0`, computes `inventory + 1`

These chunks can be computed in parallel on different CPU cores. After both complete, the results are combined using the `combine` operation.

**Critical Performance Constraint**: We NEVER want to copy the tensor data arrays as an intermediate operation. Only operate on indexes and pointers. Mask arrays are owned by operands and can be modified freely. This principle ensures that:

- Memory bandwidth is not wasted on copying data
- Cache coherency is maintained
- Parallel workers can operate on shared read-only data with thread-local write targets and thread-local masks

### Computation Chunk Size

Each tensor carries a computation chunk size attribute that controls parallelization granularity. See [Synthesizing Go Code From Chronocone Intermediate Representation](synthesis.md) for a detailed discussion of chunking strategies, parallelization, and the synthesis context for tensor operations.

---

## Three-Level Function Hierarchy

The tensor operation design uses a three-level hierarchy that separates concerns between code generation, parallelization orchestration, and actual computation:

### Level 1: High-Level Tensor Operations (Generated Code Interface)

**Purpose**: Provide a clean API for synthesized code to perform tensor operations.

**Functions**: `tensor.Add()`, `tensor.Subtract()`, `tensor.Multiply()`, `tensor.Divide()`, `tensor.Negate()`, and scalar variants like `tensor.AddScalar()`, `tensor.MultiplyScalar()`, etc.

**Parameters**: Whole `ChronoconeTensor` objects

**Behavior**:

- Accept complete tensors as operands
- Internally apply chunking based on the tensor's `chunkSize` attribute
- Create `ChronoconeTensorComputationChunkOperand` objects for each chunk
- Submit work to the controller's work channel
- Block until all chunks complete
- Return the resulting tensor

**Who Calls**: Synthesized Go code (e.g., `computeInventory()` function)

### Level 2: Application Primitives (Parallelization Orchestration)

**Purpose**: Coordinate parallel execution of operations across tensor chunks.

**Functions**: `apply_to_tensor()`, `apply_unary_to_tensor()`, `apply_binary_to_tensor()`, `apply_binary_tensors_to_tensor()`

**Parameters**: Operation functions + chunking strategy + input/output tensors

**Behavior**:

- Apply the chunking strategy to divide tensors into chunks
- For each chunk, create `ChronoconeTensorComputationChunk` objects
- Submit chunks to the controller's work channel
- Manage synchronization and error handling
- Wait for all chunks to complete

**Who Calls**: Level 1 tensor operations (under the covers)

### Level 3: Low-Level Operations (Worker Execution)

**Purpose**: Perform the actual element-wise computation on a single chunk.

**Functions**: `addOp()`, `subtractOp()`, `multiplyOp()`, `divideOp()`, `addTensorsOp()`, `subtractTensorsOp()`, `multiplyTensorsOp()`, `divideTensorsOp()`, `negateOp()`, boolean operations, comparison operations

**Parameters**: `ChronoconeTensorComputationChunkOperand` objects (not whole tensors)

**Behavior**:

- Read from input operand data slices
- Perform element-wise computation
- Write results to output operand's data slice
- Handle mask-based validity checking
- Return errors for exceptional conditions (e.g., division by zero)

**Who Calls**: Worker goroutines from the controller's worker pool

### Hierarchy Summary

```
Level 1 (Generated Code)
    ↓ Calls
    tensor.Subtract(inventory, constant_8)
        ↓ Internally uses
Level 2 (Orchestration)
    apply_binary_to_tensor(subtractOp, chunkingStrategy, left, right, output, controller)
        ↓ Submits chunks to work channel
Level 3 (Worker Execution)
    subtractOp(leftChunkOperand, rightChunkOperand, outputChunkOperand)
        ↓ Executes element-wise
    for i := 0; i < chunkLength; i++ {
        output.data[i] = left.data[i] - right.data[i]
    }
```

### Why Three Levels?

1. **Clean Abstraction for Generated Code**: Level 1 provides a simple, intuitive API that matches the CIL operations. Generated code doesn't need to know about chunking.

2. **Flexible Parallelization**: Level 2 can be optimized independently (different chunking strategies, work-stealing, etc.) without changing generated code or low-level operations.

3. **Efficient Execution**: Level 3 operates on small, cache-friendly chunks with minimal overhead. Workers execute simple loops without synchronization concerns.

4. **Separation of Concerns**: Each level has a single responsibility:
   - Level 1: User-facing API
   - Level 2: Parallelization orchestration
   - Level 3: Computational kernel

### Operation Function Signatures (Level 3)

The low-level operation functions take `ChronoconeTensorComputationChunkOperand` objects:

```go
// UnaryOperation is a function type for unary operations on a single chunk operand.
// It computes output[i] = op(input[i]) for all elements in the chunk.
type UnaryOperation[T any] func(
    input *ChronoconeTensorComputationChunkOperand[T],
    output *ChronoconeTensorComputationChunkOperand[T],
) error

// PairwiseOperation is a function type for binary operations where the left operand
// is a scalar and the right operand is a tensor chunk.
// It computes output[i] = op(left, right[i]) for all elements in the chunk.
type PairwiseOperation[L any, T any] func(
    left L,
    right *ChronoconeTensorComputationChunkOperand[T],
    output *ChronoconeTensorComputationChunkOperand[T],
) error

// TensorPairwiseOperation is a function type for binary operations on two tensor chunks.
// It computes output[i] = op(left[i], right[i]) for all elements in the chunk.
type TensorPairwiseOperation[T any] func(
    left *ChronoconeTensorComputationChunkOperand[T],
    right *ChronoconeTensorComputationChunkOperand[T],
    output *ChronoconeTensorComputationChunkOperand[T],
) error

// MatrixOperation is a function type for element-wise operations on multiple chunk operands.
// It reads from all input operands and writes to the output operand.
type MatrixOperation[T any] func(
    inputs []*ChronoconeTensorComputationChunkOperand[T],
    output *ChronoconeTensorComputationChunkOperand[T],
) error
```

### Example Operation Functions (Level 3)

```go
// negateOp negates each element of the input chunk and stores in the output chunk.
func negateOp[T Numeric](
    input *ChronoconeTensorComputationChunkOperand[T],
    output *ChronoconeTensorComputationChunkOperand[T],
) error {
    for i := 0; i < input.length; i++ {
        output.tensor.data[output.startIndex+i] = -input.tensor.data[input.startIndex+i]
    }
    return nil
}

// subtractOp performs scalar-tensor subtraction: output = left - right
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

// addTensorsOp performs tensor-tensor addition: output = left + right
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
```

Note: In actual implementation, these operations should also check and propagate masks. The mask operations are omitted here for clarity but would be included in the production implementation.
