// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package calculation

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/chronocone/chronocone_planning_language/src/main/go/io"
)

// ChronoconeTensor represents a multi-dimensional tensor with contiguous data storage.
// The tensor stores data in a one-dimensional array with dimensions arranged
// in row-major order (last dimension varies fastest).
//
// Type parameter T represents the element type of the tensor (e.g., float64, int64).
type ChronoconeTensor[T any] struct {
	// data is the contiguous one-dimensional array storing all tensor elements.
	data []T

	// dimensionNames is the ordered list of dimension names for this tensor.
	// The order matches the shape dimensions.
	dimensionNames []string

	// shape is the size of each dimension in order.
	// The product of all shape elements equals len(data).
	shape []int

	// mask is a bit mask indicating which elements are true/selected.
	// Each uint64 stores 64 bits; a 0 bit means the element is true/selected,
	// a 1 bit means the element is false/unselected. The slice length is
	// (len(data) + 63) / 64.
	mask []uint64
}

// ChronoconeComputationController coordinates all tensor operations during computation.
// It manages concept values, metric computations, and work dispatch for parallel execution.
type ChronoconeComputationController struct {
	// conceptValues is a map from concept name to an array of values for that concept.
	// This provides the dimension sizes and value domains for all tensors.
	conceptValues map[string][]interface{}

	// metricComputations is a map from metric name to its computation state.
	// This enables synchronization between parallel computation groups.
	metricComputations map[string]*MetricComputation

	// workChannel is the buffered channel for dispatching computation chunks.
	// Buffer size equals the number of CPU cores for natural backpressure.
	workChannel chan *ChronoconeTensorComputationChunk[any]

	// workerWaitGroup coordinates the lifecycle of worker goroutines.
	workerWaitGroup sync.WaitGroup

	// numWorkers is the number of worker goroutines to launch.
	numWorkers int

	// cpuCoreCount is the number of CPU cores from the architecture configuration.
	cpuCoreCount int

	// cacheSizeBytes is the cache size from the architecture configuration.
	cacheSizeBytes int

	// mainMemoryBytes is the main memory size from the architecture configuration.
	mainMemoryBytes int
}

// MetricComputation tracks the computation state for a single metric.
// It provides synchronization for cross-metric dependencies using a channel-based pattern.
type MetricComputation struct {
	// done is a channel that closes when the tensor is ready.
	// Goroutines block on this channel until the tensor is set.
	done chan struct{}

	// tensor is an atomic pointer to the computed tensor, initially nil.
	// Using atomic.Pointer ensures safe concurrent access with GetTensor and ReleaseTensor.
	tensor atomic.Pointer[ChronoconeTensor[any]]

	// consumerCount is an atomic counter for memory management.
	// Tracks how many times this metric is referenced by other computations.
	consumerCount int32
}

// ChronoconeTensorComputationChunk represents a unit of work dispatched to workers.
// It contains the operation and operands needed to process a chunk of tensor data.
//
// Type parameter T represents the element type of the tensors being processed.
type ChronoconeTensorComputationChunk[T any] struct {
	// operands is the list of input tensor chunk operands.
	operands []*ChronoconeTensorComputationChunkOperand[T]

	// matrixOp is the matrix operation function (for tensor-tensor operations).
	matrixOp MatrixOperation[T]

	// pairwiseOp is the pairwise operation function (for scalar-tensor operations).
	pairwiseOp PairwiseOperation[any, T]

	// unaryOp is the unary operation function (for single tensor operations).
	unaryOp UnaryOperation[T]

	// output is the output tensor chunk operand where results are written.
	output *ChronoconeTensorComputationChunkOperand[T]

	// waitGroup is used to signal completion of this chunk.
	waitGroup *sync.WaitGroup

	// errorChan is the channel for reporting errors during execution.
	errorChan chan<- error
}

// ChronoconeTensorComputationChunkOperand provides a view into a slice of tensor data.
// It represents a contiguous portion of a tensor for chunk-based processing.
type ChronoconeTensorComputationChunkOperand[T any] struct {
	// tensor is the underlying tensor being referenced.
	tensor *ChronoconeTensor[T]

	// startIndex is the starting index in the tensor's data array.
	startIndex int

	// length is the number of elements in this chunk.
	length int

	// mask is a bit mask indicating which elements are valid.
	// Each bit corresponds to an element; 1 means valid, 0 means invalid.
	mask []uint64
}

// MatrixOperation is a function type that performs an element-wise operation
// on input operands and writes results to an output operand.
// Used for tensor-tensor operations like addition and multiplication.
type MatrixOperation[T any] func(
	inputs []*ChronoconeTensorComputationChunkOperand[T],
	output *ChronoconeTensorComputationChunkOperand[T],
) error

// PairwiseOperation is a function type for binary operations with a scalar left operand.
// Used for scalar-tensor operations like adding a constant to each element.
type PairwiseOperation[L any, T any] func(
	left L,
	right *ChronoconeTensorComputationChunkOperand[T],
	output *ChronoconeTensorComputationChunkOperand[T],
) error

// UnaryOperation is a function type for unary operations on a single tensor.
// Used for operations like negation.
type UnaryOperation[T any] func(
	input *ChronoconeTensorComputationChunkOperand[T],
	output *ChronoconeTensorComputationChunkOperand[T],
) error

// NewChronoconeComputationController creates a new controller with empty mappings.
// The work channel and worker pool are initialized when SetArchitectureConfig is called.
//
// Returns:
//   - *ChronoconeComputationController: A new controller instance.
func NewChronoconeComputationController() *ChronoconeComputationController {
	return &ChronoconeComputationController{
		conceptValues:      make(map[string][]interface{}),
		metricComputations: make(map[string]*MetricComputation),
		workChannel:        nil,
		numWorkers:         0,
		cpuCoreCount:       0,
		cacheSizeBytes:     0,
		mainMemoryBytes:    0,
	}
}

// SetArchitectureConfig configures the controller with target hardware specs.
// This initializes the work channel with buffer size equal to core count.
// Must be called before StartWorkers().
//
// Parameters:
//   - config: The architecture specification from the IO spec.
func (c *ChronoconeComputationController) SetArchitectureConfig(config *io.ArchitectureSpecification) {
	// TODO: Implement architecture configuration
	// Initialize work channel with buffer size = config.CPUCoreCount
	// Set numWorkers, cpuCoreCount, cacheSizeBytes, mainMemoryBytes
}

// RegisterMetricComputation registers a MetricComputation for a metric name.
// This must be called during Setup() before any parallel computation begins.
//
// Parameters:
//   - metricName: The unique name of the metric.
//   - mc: The MetricComputation instance to register.
//
// Returns:
//   - error: An error if the controller or metric computation is nil.
func (c *ChronoconeComputationController) RegisterMetricComputation(metricName string, mc *MetricComputation) error {
	if c == nil {
		return fmt.Errorf("computation controller is nil")
	}
	if mc == nil {
		return fmt.Errorf("metric computation is nil")
	}
	c.metricComputations[metricName] = mc
	return nil
}

// GetMetricComputation retrieves the MetricComputation for a metric name.
// Returns nil if the metric is not registered.
//
// Parameters:
//   - metricName: The name of the metric to look up.
//
// Returns:
//   - *MetricComputation: The computation state, or nil if not found.
func (c *ChronoconeComputationController) GetMetricComputation(metricName string) *MetricComputation {
	if c == nil {
		return nil
	}
	return c.metricComputations[metricName]
}

// SetConceptValues populates the controller's concept values map from the
// package-level conceptValues variable. The string values are converted to
// interface{} for use in tensor allocation.
//
// Parameters:
//   - cv: A map from concept name to its ordered string values.
func (c *ChronoconeComputationController) SetConceptValues(cv map[string][]string) {
	if c == nil {
		return
	}
	for name, values := range cv {
		ifaceValues := make([]interface{}, len(values))
		for i, v := range values {
			ifaceValues[i] = v
		}
		c.conceptValues[name] = ifaceValues
	}
}

// GetWorkChannel returns the work channel for submitting computation chunks.
// Generated code submits chunks to this channel for worker execution.
//
// Returns:
//   - chan *ChronoconeTensorComputationChunk[any]: The work dispatch channel.
func (c *ChronoconeComputationController) GetWorkChannel() chan *ChronoconeTensorComputationChunk[any] {
	// TODO: Implement work channel access
	// Return the initialized workChannel
	return nil
}

// StartWorkers launches the worker goroutines that listen on the work channel.
// This must be called after SetArchitectureConfig and before any chunks are submitted.
// Workers continue until the work channel is closed.
func (c *ChronoconeComputationController) StartWorkers() {
	// TODO: Implement worker pool startup
	// Launch numWorkers goroutines that read from workChannel
}

// StopWorkers signals all workers to exit by closing the work channel
// and waits for completion. This should be called during Cleanup().
func (c *ChronoconeComputationController) StopWorkers() {
	// TODO: Implement worker pool shutdown
	// Close workChannel and wait for workerWaitGroup
}

// SubmitChunk submits a computation chunk to the work channel for execution.
// This call blocks if the channel buffer is full, providing natural backpressure.
//
// Parameters:
//   - chunk: The computation chunk to execute.
func (c *ChronoconeComputationController) SubmitChunk(chunk *ChronoconeTensorComputationChunk[any]) {
	// TODO: Implement chunk submission
	// Send chunk to workChannel
}

// AllocateTensor creates a new tensor with dimensions sized according to concept values.
//
// Parameters:
//   - c: The computation controller.
//   - dimensionNames: The ordered list of dimension names for the tensor.
//
// Returns:
//   - *ChronoconeTensor[any]: The allocated tensor.
//   - error: An error if allocation fails.
func AllocateTensor(c *ChronoconeComputationController, dimensionNames ...string) (*ChronoconeTensor[any], error) {
	if c == nil {
		return nil, fmt.Errorf("computation controller is nil")
	}

	shape := make([]int, len(dimensionNames))
	for i, dimName := range dimensionNames {
		conceptValues, ok := c.conceptValues[dimName]
		if !ok {
			return nil, fmt.Errorf("unknown dimension: %s", dimName)
		}
		shape[i] = len(conceptValues)
	}

	// Calculate total size
	totalSize := 1
	for _, s := range shape {
		totalSize *= s
	}
	if len(shape) == 0 {
		totalSize = 1 // Scalar case
	}

	data := make([]any, totalSize)
	// Mask is automatically zero-allocated by NewChronoconeTensor
	return NewChronoconeTensor(data, dimensionNames, shape), nil
}

// NewMetricComputation creates a new MetricComputation with an open channel.
// This should be called during Setup() for each metric that may be referenced.
//
// Parameters:
//   - consumerCount: How many times this metric is referenced by other computations.
//
// Returns:
//   - *MetricComputation: A new computation state instance.
func NewMetricComputation(consumerCount int32) *MetricComputation {
	return &MetricComputation{
		done:          make(chan struct{}),
		consumerCount: consumerCount,
	}
}

// GetTensor blocks until the tensor is available, then returns it.
// If the tensor is already set, returns immediately.
//
// Returns:
//   - *ChronoconeTensor[any]: The computed tensor.
func (mc *MetricComputation) GetTensor() *ChronoconeTensor[any] {
	if mc == nil {
		return nil
	}
	<-mc.done
	return mc.tensor.Load()
}

// SetTensor sets the tensor and closes the done channel.
// This unblocks any goroutines waiting in GetTensor().
// Returns an error if SetTensor has already been called.
//
// Parameters:
//   - tensor: The computed tensor to store.
//
// Returns:
//   - error: An error if the tensor was already set.
func (mc *MetricComputation) SetTensor(tensor *ChronoconeTensor[any]) error {
	if mc == nil {
		return fmt.Errorf("metric computation is nil")
	}
	if !mc.tensor.CompareAndSwap(nil, tensor) {
		return fmt.Errorf("tensor already set")
	}
	close(mc.done)
	return nil
}

// ReleaseTensor atomically decrements the consumer count.
// When the count reaches 0, it sets the internal tensor pointer to nil,
// allowing the Go Garbage Collector to reclaim the memory.
func (mc *MetricComputation) ReleaseTensor() {
	if mc == nil {
		return
	}
	newCount := atomic.AddInt32(&mc.consumerCount, -1)
	if newCount <= 0 {
		mc.tensor.Store(nil)
	}
}

// IsReady returns true if the tensor has been set.
//
// Returns:
//   - bool: True if the tensor is ready, false otherwise.
func (mc *MetricComputation) IsReady() bool {
	if mc == nil {
		return false
	}
	return mc.tensor.Load() != nil
}

// NewChronoconeTensor creates a new tensor with the specified dimensions and shape.
//
// Parameters:
//   - data: The contiguous data array.
//   - dimensionNames: The ordered list of dimension names.
//   - shape: The size of each dimension.
//
// Returns:
//   - *ChronoconeTensor[T]: A new tensor instance.
func NewChronoconeTensor[T any](data []T, dimensionNames []string, shape []int) *ChronoconeTensor[T] {
	// Copy slices to ensure immutability
	dimsCopy := make([]string, len(dimensionNames))
	copy(dimsCopy, dimensionNames)
	shapeCopy := make([]int, len(shape))
	copy(shapeCopy, shape)

	// Allocate mask zero-initialized with size (len(data) + 63) / 64
	maskSize := (len(data) + 63) / 64
	mask := make([]uint64, maskSize)

	return &ChronoconeTensor[T]{
		data:           data,
		dimensionNames: dimsCopy,
		shape:          shapeCopy,
		mask:           mask,
	}
}

// DimensionNames returns the ordered list of dimension names for this tensor.
//
// Returns:
//   - []string: The dimension names.
func (t *ChronoconeTensor[T]) DimensionNames() []string {
	if t == nil {
		return nil
	}
	return t.dimensionNames
}

// Shape returns the shape (size of each dimension) of this tensor.
//
// Returns:
//   - []int: The tensor shape.
func (t *ChronoconeTensor[T]) Shape() []int {
	if t == nil {
		return nil
	}
	return t.shape
}

// Data returns a reference to the underlying data array.
//
// Returns:
//   - []T: The tensor data.
func (t *ChronoconeTensor[T]) Data() []T {
	if t == nil {
		return nil
	}
	return t.data
}

// IsScalar returns true if the tensor represents a scalar value:
// no dimension names, no shape, and exactly one data element.
//
// Returns:
//   - bool: True if the tensor is a scalar, false otherwise.
func (t *ChronoconeTensor[T]) IsScalar() bool {
	if t == nil {
		return false
	}
	return len(t.dimensionNames) == 0 && len(t.shape) == 0 && len(t.data) == 1
}

// Mask returns the bit mask indicating which elements are true/selected.
// A 0 bit means true/selected, a 1 bit means false/unselected.
//
// Returns:
//   - []uint64: The mask slice, or nil if the tensor is nil.
func (t *ChronoconeTensor[T]) Mask() []uint64 {
	if t == nil {
		return nil
	}
	return t.mask
}

// SetMask sets the bit mask for this tensor.
//
// Parameters:
//   - mask: The new mask slice to assign.
func (t *ChronoconeTensor[T]) SetMask(mask []uint64) {
	if t == nil {
		return
	}
	if mask == nil {
		t.mask = nil
		return
	}
	t.mask = make([]uint64, len(mask))
	copy(t.mask, mask)
}

// MaskLen returns the number of uint64 words needed for the mask.
// This is (len(data) + 63) / 64.
//
// Returns:
//   - int: The mask length in uint64 words.
func (t *ChronoconeTensor[T]) MaskLen() int {
	if t == nil {
		return 0
	}
	return (len(t.data) + 63) / 64
}

// IsMaskedTrue returns true if the element at the given index is true/selected
// (i.e., the corresponding bit is 0). A 0 bit means true/selected,
// a 1 bit means false/unselected.
//
// Parameters:
//   - index: The element index to check.
//
// Returns:
//   - bool: True if the element is true/selected (bit is 0).
func (t *ChronoconeTensor[T]) IsMaskedTrue(index int) bool {
	if t == nil || t.mask == nil || index < 0 || index >= len(t.data) {
		return false
	}
	return (t.mask[index>>6]>>(uint(index)&0x3F))&1 == 0
}

// SetMaskBit sets the mask bit for the element at the given index.
// value=true sets the bit to 0 (true/selected), value=false sets the
// bit to 1 (false/unselected). This follows the convention where 0 means
// true/selected and 1 means false/unselected.
//
// Parameters:
//   - index: The element index to set.
//   - value: If true, selects the element (sets bit to 0). If false, deselects it (sets bit to 1).
func (t *ChronoconeTensor[T]) SetMaskBit(index int, value bool) {
	if t == nil || t.mask == nil || index < 0 || index >= len(t.data) {
		return
	}
	if value {
		t.mask[index>>6] &^= (1 << (uint(index) & 0x3F))
	} else {
		t.mask[index>>6] |= (1 << (uint(index) & 0x3F))
	}
}

// TensorDimensionNamesMatch checks that two tensors have the same dimension names
// in the same order, returning true if they match or if both are nil.
//
// Parameters:
//   - a: The first tensor.
//   - b: The second tensor.
//
// Returns:
//   - bool: True if dimension names match, false otherwise.
func TensorDimensionNamesMatch(a, b *ChronoconeTensor[any]) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	dimsA := a.DimensionNames()
	dimsB := b.DimensionNames()
	if len(dimsA) != len(dimsB) {
		return false
	}
	for i := range dimsA {
		if dimsA[i] != dimsB[i] {
			return false
		}
	}
	return true
}

// TensorDimensionNamesAreBroadcastable checks whether one tensor's dimension names
// are a compatible subset of another's. Returns true if all dimension names in
// `smaller` appear in `larger` in the same order (contiguous or non-contiguous).
// Equal dimension lists are also considered broadcastable.
//
// Parameters:
//   - smaller: The dimension names of the potentially smaller tensor.
//   - larger:  The dimension names of the potentially larger tensor.
//
// Returns:
//   - bool: True if smaller's dimensions are a compatible subset of larger's.
func TensorDimensionNamesAreBroadcastable(smaller, larger []string) bool {
	if len(smaller) > len(larger) {
		return false
	}
	si := 0
	for _, ld := range larger {
		if si < len(smaller) && smaller[si] == ld {
			si++
		}
	}
	return si == len(smaller)
}

// TensorWhereFloat64 combines branch tensors element-wise by mask.
// Branches are processed in order; first branch with mask bit=0 contributes its value.
// All branches must have matching dimensions and shapes.
// Returns a new tensor with all-zero mask (all positions resolved).
//
// Parameters:
//   - branches: A slice of branch tensors to combine. Must be non-empty.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with resolved values and all-zero mask.
//   - error: An error if branches is empty, any branch is nil, or dimensions/shapes mismatch.
func TensorWhereFloat64(branches []*ChronoconeTensor[any]) (*ChronoconeTensor[any], error) {
	if len(branches) == 0 {
		return nil, fmt.Errorf("branches is empty")
	}

	var ref *ChronoconeTensor[any]
	for i, branch := range branches {
		if branch == nil {
			return nil, fmt.Errorf("branch %d is nil", i)
		}
		if ref == nil {
			ref = branch
			continue
		}
		if len(branch.Data()) != len(ref.Data()) {
			return nil, fmt.Errorf("branch %d data length %d does not match reference length %d", i, len(branch.Data()), len(ref.Data()))
		}
		if !TensorDimensionNamesMatch(branch, ref) {
			return nil, fmt.Errorf("branch %d dimension mismatch: got %v, want %v", i, branch.DimensionNames(), ref.DimensionNames())
		}
		shapeA := branch.Shape()
		shapeB := ref.Shape()
		if len(shapeA) != len(shapeB) {
			return nil, fmt.Errorf("branch %d shape mismatch: got %v, want %v", i, shapeA, shapeB)
		}
		for j := range shapeA {
			if shapeA[j] != shapeB[j] {
				return nil, fmt.Errorf("branch %d shape mismatch at dimension %d: got %d, want %d", i, j, shapeA[j], shapeB[j])
			}
		}
	}

	n := len(ref.Data())
	result := make([]any, n)
	resultTensor := NewChronoconeTensor(result, ref.DimensionNames(), ref.Shape())

	for i, branch := range branches {
		if branch.mask != nil && len(branch.mask) < (n+63)/64 {
			return nil, fmt.Errorf("branch %d: mask length %d insufficient for %d elements", i, len(branch.mask), n)
		}
	}

	for i := 0; i < n; i++ {
		selected := false
		for _, branch := range branches {
			if branch.mask == nil || (branch.mask[i>>6]>>(uint(i)&0x3F))&1 == 0 {
				result[i] = branch.Data()[i]
				selected = true
				break
			}
		}
		if !selected {
			result[i] = branches[len(branches)-1].Data()[i]
		}
	}

	return resultTensor, nil
}

// TensorBroadcastTo broadcasts a source tensor to a target shape by repeating
// source values across missing dimensions. The source dimension names must be
// a subset of the target dimension names, appearing in the same order.
//
// Parameters:
//   - src: The source tensor to broadcast.
//   - targetDimNames: The ordered list of target dimension names.
//   - targetShape: The size of each target dimension.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with the broadcasted shape and data.
//   - error: An error if dimensions are incompatible or parameters are invalid.
func TensorBroadcastTo(src *ChronoconeTensor[any], targetDimNames []string, targetShape []int) (*ChronoconeTensor[any], error) {
	if src == nil {
		return nil, fmt.Errorf("source tensor is nil")
	}
	if len(targetDimNames) != len(targetShape) {
		return nil, fmt.Errorf("targetDimNames length %d does not match targetShape length %d", len(targetDimNames), len(targetShape))
	}

	srcDims := src.DimensionNames()
	srcShape := src.Shape()

	// Validate: src dims are a subset of target dims in the same order.
	srcIdx := 0
	mapping := make([]int, len(targetDimNames))
	for i := range targetDimNames {
		mapping[i] = -1
	}

	for ti, tDim := range targetDimNames {
		if srcIdx < len(srcDims) && srcDims[srcIdx] == tDim {
			// Verify matching dimension has the same size.
			if srcShape[srcIdx] != targetShape[ti] {
				return nil, fmt.Errorf("dimension %s shape mismatch: source %d, target %d", tDim, srcShape[srcIdx], targetShape[ti])
			}
			mapping[ti] = srcIdx
			srcIdx++
		}
	}

	if srcIdx != len(srcDims) {
		return nil, fmt.Errorf("source dimensions %v are not a compatible subset of target dimensions %v", srcDims, targetDimNames)
	}

	// Compute target total size.
	targetSize := 1
	for _, s := range targetShape {
		targetSize *= s
	}

	// Compute target strides (row-major: last dimension varies fastest).
	targetStrides := make([]int, len(targetShape))
	for i := len(targetShape) - 1; i >= 0; i-- {
		if i == len(targetShape)-1 {
			targetStrides[i] = 1
		} else {
			targetStrides[i] = targetStrides[i+1] * targetShape[i+1]
		}
	}

	// Compute source strides.
	srcStrides := make([]int, len(srcShape))
	for i := len(srcShape) - 1; i >= 0; i-- {
		if i == len(srcShape)-1 {
			srcStrides[i] = 1
		} else {
			srcStrides[i] = srcStrides[i+1] * srcShape[i+1]
		}
	}

	// Allocate target tensor.
	targetData := make([]any, targetSize)
	targetTensor := NewChronoconeTensor(targetData, targetDimNames, targetShape)

	srcData := src.Data()

	// Broadcast data and mask.
	for flatIdx := 0; flatIdx < targetSize; flatIdx++ {
		// Decompose flat index to target coordinates and compute source flat index.
		srcFlatIdx := 0
		remaining := flatIdx
		for ti := 0; ti < len(targetShape); ti++ {
			coord := remaining / targetStrides[ti]
			remaining = remaining % targetStrides[ti]
			if mapping[ti] != -1 {
				si := mapping[ti]
				srcFlatIdx += coord * srcStrides[si]
			}
		}

		targetData[flatIdx] = srcData[srcFlatIdx]

		// Broadcast mask: if source position is invalid, mark target position invalid.
		if src.mask != nil && !src.IsMaskedTrue(srcFlatIdx) {
			targetTensor.SetMaskBit(flatIdx, false)
		}
	}

	return targetTensor, nil
}
