// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package calculation

import (
	"fmt"
	"math"
)

// TensorGreaterThanFloat64Mask performs element-wise comparison (>) of two *ChronoconeTensor[any]
// tensors, returning a new tensor. Elements must be float64.
// The returned tensor carries the left operand's data and a computed mask
// where bit 0 (true) indicates the element satisfied the comparison.
//
// Mask convention: bit=0 means true/selected (comparison holds), bit=1 means false/unselected.
//
// Parameters:
//   - a: The left tensor operand.
//   - b: The right tensor operand.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with left-operand data and computed mask.
//   - error: An error if tensors are nil, dimensions mismatch, or elements are not float64.
func TensorGreaterThanFloat64Mask(a, b *ChronoconeTensor[any]) (*ChronoconeTensor[any], error) {
	if a == nil || b == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	if !TensorDimensionNamesMatch(a, b) {
		return nil, fmt.Errorf("dimension mismatch")
	}
	dataA, dataB := a.Data(), b.Data()
	copiedData := make([]any, len(dataA))
	copy(copiedData, dataA)
	mask := make([]uint64, (len(dataA)+63)/64)
	for i := range dataA {
		fa, ok := dataA[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataA[i])
		}
		fb, ok := dataB[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataB[i])
		}
		result := fa > fb
		if math.IsNaN(fa) || math.IsNaN(fb) {
			result = false
		}
		if result {
			// Set mask bit to 0 (true/selected)
			mask[i>>6] &^= (1 << (uint(i) & 0x3F))
		} else {
			// Set mask bit to 1 (false/unselected)
			mask[i>>6] |= (1 << (uint(i) & 0x3F))
		}
	}
	result := NewChronoconeTensor(copiedData, a.DimensionNames(), a.Shape())
	result.mask = mask
	return result, nil
}

// TensorGreaterEqualFloat64Mask performs element-wise comparison (>=) of two *ChronoconeTensor[any]
// tensors, returning a new tensor. Elements must be float64.
// The returned tensor carries the left operand's data and a computed mask
// where bit 0 (true) indicates the element satisfied the comparison.
//
// Mask convention: bit=0 means true/selected (comparison holds), bit=1 means false/unselected.
//
// Parameters:
//   - a: The left tensor operand.
//   - b: The right tensor operand.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with left-operand data and computed mask.
//   - error: An error if tensors are nil, dimensions mismatch, or elements are not float64.
func TensorGreaterEqualFloat64Mask(a, b *ChronoconeTensor[any]) (*ChronoconeTensor[any], error) {
	if a == nil || b == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	if !TensorDimensionNamesMatch(a, b) {
		return nil, fmt.Errorf("dimension mismatch")
	}
	dataA, dataB := a.Data(), b.Data()
	copiedData := make([]any, len(dataA))
	copy(copiedData, dataA)
	mask := make([]uint64, (len(dataA)+63)/64)
	for i := range dataA {
		fa, ok := dataA[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataA[i])
		}
		fb, ok := dataB[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataB[i])
		}
		result := fa >= fb
		if math.IsNaN(fa) || math.IsNaN(fb) {
			result = false
		}
		if result {
			// Set mask bit to 0 (true/selected)
			mask[i>>6] &^= (1 << (uint(i) & 0x3F))
		} else {
			// Set mask bit to 1 (false/unselected)
			mask[i>>6] |= (1 << (uint(i) & 0x3F))
		}
	}
	result := NewChronoconeTensor(copiedData, a.DimensionNames(), a.Shape())
	result.mask = mask
	return result, nil
}

// TensorLessThanFloat64Mask performs element-wise comparison (<) of two *ChronoconeTensor[any]
// tensors, returning a new tensor. Elements must be float64.
// The returned tensor carries the left operand's data and a computed mask
// where bit 0 (true) indicates the element satisfied the comparison.
//
// Mask convention: bit=0 means true/selected (comparison holds), bit=1 means false/unselected.
//
// Parameters:
//   - a: The left tensor operand.
//   - b: The right tensor operand.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with left-operand data and computed mask.
//   - error: An error if tensors are nil, dimensions mismatch, or elements are not float64.
func TensorLessThanFloat64Mask(a, b *ChronoconeTensor[any]) (*ChronoconeTensor[any], error) {
	if a == nil || b == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	if !TensorDimensionNamesMatch(a, b) {
		return nil, fmt.Errorf("dimension mismatch")
	}
	dataA, dataB := a.Data(), b.Data()
	copiedData := make([]any, len(dataA))
	copy(copiedData, dataA)
	mask := make([]uint64, (len(dataA)+63)/64)
	for i := range dataA {
		fa, ok := dataA[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataA[i])
		}
		fb, ok := dataB[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataB[i])
		}
		result := fa < fb
		if math.IsNaN(fa) || math.IsNaN(fb) {
			result = false
		}
		if result {
			mask[i>>6] &^= (1 << (uint(i) & 0x3F))
		} else {
			mask[i>>6] |= (1 << (uint(i) & 0x3F))
		}
	}
	result := NewChronoconeTensor(copiedData, a.DimensionNames(), a.Shape())
	result.mask = mask
	return result, nil
}

// TensorLessEqualFloat64Mask performs element-wise comparison (<=) of two *ChronoconeTensor[any]
// tensors, returning a new tensor. Elements must be float64.
// The returned tensor carries the left operand's data and a computed mask
// where bit 0 (true) indicates the element satisfied the comparison.
//
// Mask convention: bit=0 means true/selected (comparison holds), bit=1 means false/unselected.
//
// Parameters:
//   - a: The left tensor operand.
//   - b: The right tensor operand.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with left-operand data and computed mask.
//   - error: An error if tensors are nil, dimensions mismatch, or elements are not float64.
func TensorLessEqualFloat64Mask(a, b *ChronoconeTensor[any]) (*ChronoconeTensor[any], error) {
	if a == nil || b == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	if !TensorDimensionNamesMatch(a, b) {
		return nil, fmt.Errorf("dimension mismatch")
	}
	dataA, dataB := a.Data(), b.Data()
	copiedData := make([]any, len(dataA))
	copy(copiedData, dataA)
	mask := make([]uint64, (len(dataA)+63)/64)
	for i := range dataA {
		fa, ok := dataA[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataA[i])
		}
		fb, ok := dataB[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataB[i])
		}
		result := fa <= fb
		if math.IsNaN(fa) || math.IsNaN(fb) {
			result = false
		}
		if result {
			mask[i>>6] &^= (1 << (uint(i) & 0x3F))
		} else {
			mask[i>>6] |= (1 << (uint(i) & 0x3F))
		}
	}
	result := NewChronoconeTensor(copiedData, a.DimensionNames(), a.Shape())
	result.mask = mask
	return result, nil
}

// TensorEqualFloat64Mask performs element-wise comparison (==) of two *ChronoconeTensor[any]
// tensors, returning a new tensor. Elements must be float64.
// The returned tensor carries the left operand's data and a computed mask
// where bit 0 (true) indicates the element satisfied the comparison.
//
// Mask convention: bit=0 means true/selected (comparison holds), bit=1 means false/unselected.
//
// Parameters:
//   - a: The left tensor operand.
//   - b: The right tensor operand.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with left-operand data and computed mask.
//   - error: An error if tensors are nil, dimensions mismatch, or elements are not float64.
func TensorEqualFloat64Mask(a, b *ChronoconeTensor[any]) (*ChronoconeTensor[any], error) {
	if a == nil || b == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	if !TensorDimensionNamesMatch(a, b) {
		return nil, fmt.Errorf("dimension mismatch")
	}
	dataA, dataB := a.Data(), b.Data()
	copiedData := make([]any, len(dataA))
	copy(copiedData, dataA)
	mask := make([]uint64, (len(dataA)+63)/64)
	for i := range dataA {
		fa, ok := dataA[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataA[i])
		}
		fb, ok := dataB[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataB[i])
		}
		result := fa == fb
		if math.IsNaN(fa) || math.IsNaN(fb) {
			result = false
		}
		if result {
			mask[i>>6] &^= (1 << (uint(i) & 0x3F))
		} else {
			mask[i>>6] |= (1 << (uint(i) & 0x3F))
		}
	}
	result := NewChronoconeTensor(copiedData, a.DimensionNames(), a.Shape())
	result.mask = mask
	return result, nil
}

// TensorNotEqualFloat64Mask performs element-wise comparison (!=) of two *ChronoconeTensor[any]
// tensors, returning a new tensor. Elements must be float64.
// The returned tensor carries the left operand's data and a computed mask
// where bit 0 (true) indicates the element satisfied the comparison.
//
// Mask convention: bit=0 means true/selected (comparison holds), bit=1 means false/unselected.
//
// Parameters:
//   - a: The left tensor operand.
//   - b: The right tensor operand.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with left-operand data and computed mask.
//   - error: An error if tensors are nil, dimensions mismatch, or elements are not float64.
func TensorNotEqualFloat64Mask(a, b *ChronoconeTensor[any]) (*ChronoconeTensor[any], error) {
	if a == nil || b == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	if !TensorDimensionNamesMatch(a, b) {
		return nil, fmt.Errorf("dimension mismatch")
	}
	dataA, dataB := a.Data(), b.Data()
	copiedData := make([]any, len(dataA))
	copy(copiedData, dataA)
	mask := make([]uint64, (len(dataA)+63)/64)
	for i := range dataA {
		fa, ok := dataA[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataA[i])
		}
		fb, ok := dataB[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataB[i])
		}
		result := fa != fb
		if math.IsNaN(fa) || math.IsNaN(fb) {
			result = false
		}
		if result {
			mask[i>>6] &^= (1 << (uint(i) & 0x3F))
		} else {
			mask[i>>6] |= (1 << (uint(i) & 0x3F))
		}
	}
	result := NewChronoconeTensor(copiedData, a.DimensionNames(), a.Shape())
	result.mask = mask
	return result, nil
}

// TensorGreaterThanScalarFloat64Mask compares each element of a *ChronoconeTensor[any] tensor
// against a scalar, returning a new tensor. Elements must be float64.
// The returned tensor carries the input tensor's data and a computed mask
// where bit 0 (true) indicates the element satisfied the comparison.
//
// Mask convention: bit=0 means true/selected (comparison holds), bit=1 means false/unselected.
//
// Parameters:
//   - a: The tensor operand.
//   - scalar: The scalar value to compare against.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with input data and computed mask.
//   - error: An error if the tensor is nil or elements are not float64.
func TensorGreaterThanScalarFloat64Mask(a *ChronoconeTensor[any], scalar float64) (*ChronoconeTensor[any], error) {
	if a == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	dataA := a.Data()
	copiedData := make([]any, len(dataA))
	copy(copiedData, dataA)
	mask := make([]uint64, (len(dataA)+63)/64)
	for i := range dataA {
		fa, ok := dataA[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataA[i])
		}
		result := fa > scalar
		if math.IsNaN(fa) || math.IsNaN(scalar) {
			result = false
		}
		if result {
			mask[i>>6] &^= (1 << (uint(i) & 0x3F))
		} else {
			mask[i>>6] |= (1 << (uint(i) & 0x3F))
		}
	}
	result := NewChronoconeTensor(copiedData, a.DimensionNames(), a.Shape())
	result.mask = mask
	return result, nil
}

// TensorGreaterEqualScalarFloat64Mask compares each element of a *ChronoconeTensor[any] tensor
// against a scalar, returning a new tensor. Elements must be float64.
// The returned tensor carries the input tensor's data and a computed mask
// where bit 0 (true) indicates the element satisfied the comparison.
//
// Mask convention: bit=0 means true/selected (comparison holds), bit=1 means false/unselected.
//
// Parameters:
//   - a: The tensor operand.
//   - scalar: The scalar value to compare against.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with input data and computed mask.
//   - error: An error if the tensor is nil or elements are not float64.
func TensorGreaterEqualScalarFloat64Mask(a *ChronoconeTensor[any], scalar float64) (*ChronoconeTensor[any], error) {
	if a == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	dataA := a.Data()
	copiedData := make([]any, len(dataA))
	copy(copiedData, dataA)
	mask := make([]uint64, (len(dataA)+63)/64)
	for i := range dataA {
		fa, ok := dataA[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataA[i])
		}
		result := fa >= scalar
		if math.IsNaN(fa) || math.IsNaN(scalar) {
			result = false
		}
		if result {
			mask[i>>6] &^= (1 << (uint(i) & 0x3F))
		} else {
			mask[i>>6] |= (1 << (uint(i) & 0x3F))
		}
	}
	result := NewChronoconeTensor(copiedData, a.DimensionNames(), a.Shape())
	result.mask = mask
	return result, nil
}

// TensorLessThanScalarFloat64Mask compares each element of a *ChronoconeTensor[any] tensor
// against a scalar, returning a new tensor. Elements must be float64.
// The returned tensor carries the input tensor's data and a computed mask
// where bit 0 (true) indicates the element satisfied the comparison.
//
// Mask convention: bit=0 means true/selected (comparison holds), bit=1 means false/unselected.
//
// Parameters:
//   - a: The tensor operand.
//   - scalar: The scalar value to compare against.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with input data and computed mask.
//   - error: An error if the tensor is nil or elements are not float64.
func TensorLessThanScalarFloat64Mask(a *ChronoconeTensor[any], scalar float64) (*ChronoconeTensor[any], error) {
	if a == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	dataA := a.Data()
	copiedData := make([]any, len(dataA))
	copy(copiedData, dataA)
	mask := make([]uint64, (len(dataA)+63)/64)
	for i := range dataA {
		fa, ok := dataA[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataA[i])
		}
		result := fa < scalar
		if math.IsNaN(fa) || math.IsNaN(scalar) {
			result = false
		}
		if result {
			mask[i>>6] &^= (1 << (uint(i) & 0x3F))
		} else {
			mask[i>>6] |= (1 << (uint(i) & 0x3F))
		}
	}
	result := NewChronoconeTensor(copiedData, a.DimensionNames(), a.Shape())
	result.mask = mask
	return result, nil
}

// TensorLessEqualScalarFloat64Mask compares each element of a *ChronoconeTensor[any] tensor
// against a scalar, returning a new tensor. Elements must be float64.
// The returned tensor carries the input tensor's data and a computed mask
// where bit 0 (true) indicates the element satisfied the comparison.
//
// Mask convention: bit=0 means true/selected (comparison holds), bit=1 means false/unselected.
//
// Parameters:
//   - a: The tensor operand.
//   - scalar: The scalar value to compare against.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with input data and computed mask.
//   - error: An error if the tensor is nil or elements are not float64.
func TensorLessEqualScalarFloat64Mask(a *ChronoconeTensor[any], scalar float64) (*ChronoconeTensor[any], error) {
	if a == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	dataA := a.Data()
	copiedData := make([]any, len(dataA))
	copy(copiedData, dataA)
	mask := make([]uint64, (len(dataA)+63)/64)
	for i := range dataA {
		fa, ok := dataA[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataA[i])
		}
		result := fa <= scalar
		if math.IsNaN(fa) || math.IsNaN(scalar) {
			result = false
		}
		if result {
			mask[i>>6] &^= (1 << (uint(i) & 0x3F))
		} else {
			mask[i>>6] |= (1 << (uint(i) & 0x3F))
		}
	}
	result := NewChronoconeTensor(copiedData, a.DimensionNames(), a.Shape())
	result.mask = mask
	return result, nil
}

// TensorEqualScalarFloat64Mask compares each element of a *ChronoconeTensor[any] tensor
// against a scalar, returning a new tensor. Elements must be float64.
// The returned tensor carries the input tensor's data and a computed mask
// where bit 0 (true) indicates the element satisfied the comparison.
//
// Mask convention: bit=0 means true/selected (comparison holds), bit=1 means false/unselected.
//
// Parameters:
//   - a: The tensor operand.
//   - scalar: The scalar value to compare against.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with input data and computed mask.
//   - error: An error if the tensor is nil or elements are not float64.
func TensorEqualScalarFloat64Mask(a *ChronoconeTensor[any], scalar float64) (*ChronoconeTensor[any], error) {
	if a == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	dataA := a.Data()
	copiedData := make([]any, len(dataA))
	copy(copiedData, dataA)
	mask := make([]uint64, (len(dataA)+63)/64)
	for i := range dataA {
		fa, ok := dataA[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataA[i])
		}
		result := fa == scalar
		if math.IsNaN(fa) || math.IsNaN(scalar) {
			result = false
		}
		if result {
			mask[i>>6] &^= (1 << (uint(i) & 0x3F))
		} else {
			mask[i>>6] |= (1 << (uint(i) & 0x3F))
		}
	}
	result := NewChronoconeTensor(copiedData, a.DimensionNames(), a.Shape())
	result.mask = mask
	return result, nil
}

// TensorNotEqualScalarFloat64Mask compares each element of a *ChronoconeTensor[any] tensor
// against a scalar, returning a new tensor. Elements must be float64.
// The returned tensor carries the input tensor's data and a computed mask
// where bit 0 (true) indicates the element satisfied the comparison.
//
// Mask convention: bit=0 means true/selected (comparison holds), bit=1 means false/unselected.
//
// Parameters:
//   - a: The tensor operand.
//   - scalar: The scalar value to compare against.
//
// Returns:
//   - *ChronoconeTensor[any]: A new tensor with input data and computed mask.
//   - error: An error if the tensor is nil or elements are not float64.
func TensorNotEqualScalarFloat64Mask(a *ChronoconeTensor[any], scalar float64) (*ChronoconeTensor[any], error) {
	if a == nil {
		return nil, fmt.Errorf("tensor is nil")
	}
	dataA := a.Data()
	copiedData := make([]any, len(dataA))
	copy(copiedData, dataA)
	mask := make([]uint64, (len(dataA)+63)/64)
	for i := range dataA {
		fa, ok := dataA[i].(float64)
		if !ok {
			return nil, fmt.Errorf("element %d not float64: %T", i, dataA[i])
		}
		result := fa != scalar
		if math.IsNaN(fa) || math.IsNaN(scalar) {
			result = false
		}
		if result {
			mask[i>>6] &^= (1 << (uint(i) & 0x3F))
		} else {
			mask[i>>6] |= (1 << (uint(i) & 0x3F))
		}
	}
	result := NewChronoconeTensor(copiedData, a.DimensionNames(), a.Shape())
	result.mask = mask
	return result, nil
}
