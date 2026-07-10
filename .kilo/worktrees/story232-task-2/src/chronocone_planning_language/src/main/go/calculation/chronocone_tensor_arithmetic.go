// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package calculation

import (
	"fmt"
)

// TensorAddFloat64 performs element-wise addition of two *ChronoconeTensor[any]
// tensors, returning a new tensor. Elements must be float64.
//
// Mask propagation: the left operand's mask is copied to the output tensor.
//
// Parameters:
//   - a: The first tensor operand.
//   - b: The second tensor operand.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with the sum.
//   - error: An error if tensors are nil, dimensions mismatch, or elements are not float64.
func TensorAddFloat64(a, b *ChronoconeTensor[any]) (*ChronoconeTensor[any], error) {
	if a == nil || b == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	if !TensorDimensionNamesMatch(a, b) {
		return nil, fmt.Errorf("dimension mismatch")
	}
	dataA, dataB := a.Data(), b.Data()
	result := make([]any, len(dataA))
	for i := range dataA {
		fa, ok := dataA[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataA[i])
		}
		fb, ok := dataB[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataB[i])
		}
		result[i] = fa + fb
	}
	resultTensor := NewChronoconeTensor(result, a.DimensionNames(), a.Shape())
	copy(resultTensor.mask, a.mask)
	return resultTensor, nil
}

// TensorSubtractFloat64 performs element-wise subtraction of two *ChronoconeTensor[any]
// tensors (a - b), returning a new tensor. Elements must be float64.
//
// Mask propagation: the left operand's mask is copied to the output tensor.
//
// Parameters:
//   - a: The first tensor operand (minuend).
//   - b: The second tensor operand (subtrahend).
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with the difference.
//   - error: An error if tensors are nil, dimensions mismatch, or elements are not float64.
func TensorSubtractFloat64(a, b *ChronoconeTensor[any]) (*ChronoconeTensor[any], error) {
	if a == nil || b == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	if !TensorDimensionNamesMatch(a, b) {
		return nil, fmt.Errorf("dimension mismatch")
	}
	dataA, dataB := a.Data(), b.Data()
	result := make([]any, len(dataA))
	for i := range dataA {
		fa, ok := dataA[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataA[i])
		}
		fb, ok := dataB[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataB[i])
		}
		result[i] = fa - fb
	}
	resultTensor := NewChronoconeTensor(result, a.DimensionNames(), a.Shape())
	copy(resultTensor.mask, a.mask)
	return resultTensor, nil
}

// TensorMultiplyFloat64 performs element-wise multiplication of two *ChronoconeTensor[any]
// tensors, returning a new tensor. Elements must be float64.
//
// Mask propagation: the left operand's mask is copied to the output tensor.
//
// Parameters:
//   - a: The first tensor operand.
//   - b: The second tensor operand.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with the product.
//   - error: An error if tensors are nil, dimensions mismatch, or elements are not float64.
func TensorMultiplyFloat64(a, b *ChronoconeTensor[any]) (*ChronoconeTensor[any], error) {
	if a == nil || b == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	if !TensorDimensionNamesMatch(a, b) {
		return nil, fmt.Errorf("dimension mismatch")
	}
	dataA, dataB := a.Data(), b.Data()
	result := make([]any, len(dataA))
	for i := range dataA {
		fa, ok := dataA[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataA[i])
		}
		fb, ok := dataB[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataB[i])
		}
		result[i] = fa * fb
	}
	resultTensor := NewChronoconeTensor(result, a.DimensionNames(), a.Shape())
	copy(resultTensor.mask, a.mask)
	return resultTensor, nil
}

// TensorDivideFloat64 performs element-wise division of two *ChronoconeTensor[any]
// tensors (a / b), returning a new tensor. Elements must be float64.
//
// Mask propagation: the left operand's mask is copied to the output tensor.
//
// Parameters:
//   - a: The first tensor operand (dividend).
//   - b: The second tensor operand (divisor).
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with the quotient.
//   - error: An error if tensors are nil, dimensions mismatch, elements are not float64,
//     or division by zero is attempted.
func TensorDivideFloat64(a, b *ChronoconeTensor[any]) (*ChronoconeTensor[any], error) {
	if a == nil || b == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	if !TensorDimensionNamesMatch(a, b) {
		return nil, fmt.Errorf("dimension mismatch")
	}
	dataA, dataB := a.Data(), b.Data()
	result := make([]any, len(dataA))
	for i := range dataA {
		fa, ok := dataA[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataA[i])
		}
		fb, ok := dataB[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataB[i])
		}
		if fb == 0.0 {
			return nil, fmt.Errorf("division by zero at element %d", i)
		}
		result[i] = fa / fb
	}
	resultTensor := NewChronoconeTensor(result, a.DimensionNames(), a.Shape())
	copy(resultTensor.mask, a.mask)
	return resultTensor, nil
}

// TensorNegateFloat64 performs element-wise negation of a *ChronoconeTensor[any]
// tensor, returning a new tensor. Elements must be float64.
//
// Mask propagation: the input tensor's mask is copied to the output tensor.
//
// Parameters:
//   - t: The tensor to negate.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with negated values.
//   - error: An error if the tensor is nil or elements are not float64.
func TensorNegateFloat64(t *ChronoconeTensor[any]) (*ChronoconeTensor[any], error) {
	if t == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	data := t.Data()
	result := make([]any, len(data))
	for i := range data {
		f, ok := data[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, data[i])
		}
		result[i] = -f
	}
	resultTensor := NewChronoconeTensor(result, t.DimensionNames(), t.Shape())
	copy(resultTensor.mask, t.mask)
	return resultTensor, nil
}

// ScalarSubtractTensorFloat64 subtracts each element of a *ChronoconeTensor[any]
// tensor from a scalar (scalar - t), returning a new tensor. Elements must be float64.
//
// Mask propagation: the input tensor's mask is copied to the output tensor.
//
// Parameters:
//   - scalar: The scalar minuend.
//   - t: The tensor subtrahend.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with scalar minus each element.
//   - error: An error if the tensor is nil or elements are not float64.
func ScalarSubtractTensorFloat64(scalar float64, t *ChronoconeTensor[any]) (*ChronoconeTensor[any], error) {
	if t == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	data := t.Data()
	result := make([]any, len(data))
	for i := range data {
		f, ok := data[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, data[i])
		}
		result[i] = scalar - f
	}
	resultTensor := NewChronoconeTensor(result, t.DimensionNames(), t.Shape())
	copy(resultTensor.mask, t.mask)
	return resultTensor, nil
}

// ScalarDivideTensorFloat64 divides a scalar by each element of a
// *ChronoconeTensor[any] tensor (scalar / t), returning a new tensor.
// Elements must be float64.
//
// Mask propagation: the input tensor's mask is copied to the output tensor.
//
// Parameters:
//   - scalar: The scalar dividend.
//   - t: The tensor divisor.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with scalar divided by each element.
//   - error: An error if the tensor is nil, elements are not float64,
//     or any tensor element is zero.
func ScalarDivideTensorFloat64(scalar float64, t *ChronoconeTensor[any]) (*ChronoconeTensor[any], error) {
	if t == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	data := t.Data()
	result := make([]any, len(data))
	for i := range data {
		f, ok := data[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, data[i])
		}
		if f == 0.0 {
			return nil, fmt.Errorf("division by zero at element %d", i)
		}
		result[i] = scalar / f
	}
	resultTensor := NewChronoconeTensor(result, t.DimensionNames(), t.Shape())
	copy(resultTensor.mask, t.mask)
	return resultTensor, nil
}

// TensorAddScalarFloat64 adds a scalar to each element of a *ChronoconeTensor[any]
// tensor, returning a new tensor. Elements must be float64.
//
// Mask propagation: the input tensor's mask is copied to the output tensor.
//
// Parameters:
//   - t: The tensor operand.
//   - scalar: The scalar value to add.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with scalar added to each element.
//   - error: An error if the tensor is nil or elements are not float64.
func TensorAddScalarFloat64(t *ChronoconeTensor[any], scalar float64) (*ChronoconeTensor[any], error) {
	if t == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	data := t.Data()
	result := make([]any, len(data))
	for i := range data {
		f, ok := data[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, data[i])
		}
		result[i] = f + scalar
	}
	resultTensor := NewChronoconeTensor(result, t.DimensionNames(), t.Shape())
	copy(resultTensor.mask, t.mask)
	return resultTensor, nil
}

// TensorSubtractScalarFloat64 subtracts a scalar from each element of a
// *ChronoconeTensor[any] tensor (t - scalar), returning a new tensor.
// Elements must be float64.
//
// Mask propagation: the input tensor's mask is copied to the output tensor.
//
// Parameters:
//   - t: The tensor operand (minuend).
//   - scalar: The scalar value to subtract (subtrahend).
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with scalar subtracted from each element.
//   - error: An error if the tensor is nil or elements are not float64.
func TensorSubtractScalarFloat64(t *ChronoconeTensor[any], scalar float64) (*ChronoconeTensor[any], error) {
	if t == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	data := t.Data()
	result := make([]any, len(data))
	for i := range data {
		f, ok := data[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, data[i])
		}
		result[i] = f - scalar
	}
	resultTensor := NewChronoconeTensor(result, t.DimensionNames(), t.Shape())
	copy(resultTensor.mask, t.mask)
	return resultTensor, nil
}

// TensorMultiplyScalarFloat64 multiplies each element of a *ChronoconeTensor[any]
// tensor by a scalar, returning a new tensor. Elements must be float64.
//
// Mask propagation: the input tensor's mask is copied to the output tensor.
//
// Parameters:
//   - t: The tensor operand.
//   - scalar: The scalar value to multiply by.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with each element multiplied by scalar.
//   - error: An error if the tensor is nil or elements are not float64.
func TensorMultiplyScalarFloat64(t *ChronoconeTensor[any], scalar float64) (*ChronoconeTensor[any], error) {
	if t == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	data := t.Data()
	result := make([]any, len(data))
	for i := range data {
		f, ok := data[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, data[i])
		}
		result[i] = f * scalar
	}
	resultTensor := NewChronoconeTensor(result, t.DimensionNames(), t.Shape())
	copy(resultTensor.mask, t.mask)
	return resultTensor, nil
}

// TensorDivideScalarFloat64 divides each element of a *ChronoconeTensor[any]
// tensor by a scalar (t / scalar), returning a new tensor. Elements must be float64.
//
// Mask propagation: the input tensor's mask is copied to the output tensor.
//
// Parameters:
//   - t: The tensor operand (dividend).
//   - scalar: The scalar divisor.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with each element divided by scalar.
//   - error: An error if the tensor is nil, elements are not float64,
//     or scalar is zero.
func TensorDivideScalarFloat64(t *ChronoconeTensor[any], scalar float64) (*ChronoconeTensor[any], error) {
	if t == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	if scalar == 0.0 {
		return nil, fmt.Errorf("division by zero")
	}
	data := t.Data()
	result := make([]any, len(data))
	for i := range data {
		f, ok := data[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, data[i])
		}
		result[i] = f / scalar
	}
	resultTensor := NewChronoconeTensor(result, t.DimensionNames(), t.Shape())
	copy(resultTensor.mask, t.mask)
	return resultTensor, nil
}
