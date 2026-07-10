// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package calculation

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
)

// PostfixOp represents a single entry in a postfix (Reverse Polish Notation)
// expression table. PostfixOps are arranged in post-order to describe a
// computation that can be evaluated by a stack-based evaluator without
// recursion or nested function calls.
//
// The Kind field determines how the evaluator processes the entry:
//   - "push": Push the tensor referenced by Name onto the evaluation stack.
//     When Name is empty and Value is non-zero, the literal float64 value is
//     pushed as a scalar tensor.
//   - "apply": Pop the required number of operands from the stack, apply the
//     operator specified by Op, and push the result back onto the stack.
type PostfixOp struct {
	// Kind indicates the type of postfix operation.
	// Must be either "push" (load a tensor by name or value), "apply" (execute an operator), or "conditional" (branching).
	Kind string

	// Op specifies the operator to apply when Kind is "apply".
	// Supported arithmetic operators: Add, Subtract, Multiply, Divide, Negate.
	// Supported comparison operators: Equal, NotEqual, LessThan, GreaterThan, LessThanOrEqual, GreaterThanOrEqual.
	// Supported logical operators: And, Or, Not.
	// For push operations, Bool and Nil are used as literal markers.
	// Empty (Unknown) when Kind is "push" for normal literals.
	Op operator.Operator

	// Name is the tensor or input name to push onto the stack when Kind is "push".
	// Empty when Kind is "apply", or when pushing a literal value via the Value field.
	Name string

	// Value is a literal float64 value to push as a scalar tensor when Kind is "push".
	// Used for literal numeric operands in arithmetic expressions (e.g., 3.14).
	// Zero when Kind is "apply" or when Name is used.
	Value float64

	// Branches holds the sub-branches for a "conditional" postfix operation.
	// Each branch has a condition (postfix ops evaluating to a mask) and a body
	// (postfix ops evaluating to a result tensor). A nil condition means an
	// always-true mask (else/default branch).
	Branches []PostfixBranch
}

// PostfixBranch represents a single branch within a "conditional" PostfixOp.
// The Condition field is nil for else/default branches (always-true mask).
type PostfixBranch struct {
	Condition []PostfixOp // Postfix ops for the condition expression; nil for always-true
	Body      []PostfixOp // Postfix ops for the body expression
}

// scalarComparisonResult evaluates a comparison operator on two scalar float64
// values and returns a scalar ChronoconeTensor with mask bits set according to
// the comparison result (AllTrueMask for true, AllFalseMask for false).
func scalarComparisonResult(op operator.Operator, a, b float64) *ChronoconeTensor[any] {
	tensor := NewChronoconeTensor([]any{a}, []string{}, []int{})
	var result bool
	switch op {
	case operator.Equal:
		result = a == b
	case operator.NotEqual:
		result = a != b
	case operator.LessThan:
		result = a < b
	case operator.GreaterThan:
		result = a > b
	case operator.LessThanOrEqual:
		result = a <= b
	case operator.GreaterThanOrEqual:
		result = a >= b
	default:
		// Unknown operator falls through to false.
	}
	if result {
		tensor.SetMask(AllTrueMask(1))
	} else {
		tensor.SetMask(AllFalseMask(1))
	}
	return tensor
}

// isBareCondition returns true if the condition expression consists solely of
// push operations (no apply ops that would set a meaningful boolean mask).
// A bare condition is one like "if data_a" where a metric reference is used
// directly as a boolean — the mask on the resulting tensor is an identity
// all-true mask from NewChronoconeTensor, not a computed boolean mask.
func isBareCondition(ops []PostfixOp) bool {
	for _, op := range ops {
		if op.Kind == "apply" {
			return false
		}
	}
	return true
}

// float64TruthinessMaskFromTensor extracts float64 data from a tensor and
// computes a truthiness mask using the convention: zero = false, non-zero = true.
//
// The returned slice is allocated by Float64TruthinessMask and has the same
// length convention as other mask slices ((n+63)/64 uint64 words). NaN values
// are treated as false.
//
// Parameters:
//   - t: The tensor whose raw data should be interpreted for truthiness.
//
// Returns:
//   - []uint64: A truthiness mask, or an AllTrueMask for t.Data() length if
//     the data cannot be converted to float64.
func float64TruthinessMaskFromTensor(t *ChronoconeTensor[any]) []uint64 {
	if t == nil {
		return nil
	}
	data := t.Data()
	n := len(data)
	if n == 0 {
		return AllTrueMask(n)
	}
	f64Data := make([]float64, n)
	for i, v := range data {
		fv, ok := v.(float64)
		if !ok {
			// Non-float64 data: treat as if all elements are truthy
			return AllTrueMask(n)
		}
		f64Data[i] = fv
	}
	return Float64TruthinessMask(f64Data)
}

// broadcastIfNeeded attempts to broadcast the smaller tensor to match the larger
// tensor's dimensions. Returns the (possibly broadcasted) tensors and an error
// if broadcasting is not possible.
func broadcastIfNeeded(a, b *ChronoconeTensor[any]) (*ChronoconeTensor[any], *ChronoconeTensor[any], error) {
	if TensorDimensionNamesMatch(a, b) {
		return a, b, nil
	}

	var smaller, larger *ChronoconeTensor[any]
	if len(a.DimensionNames()) < len(b.DimensionNames()) {
		smaller = a
		larger = b
	} else if len(b.DimensionNames()) < len(a.DimensionNames()) {
		smaller = b
		larger = a
	} else {
		return nil, nil, fmt.Errorf("dimension mismatch")
	}

	if TensorDimensionNamesAreBroadcastable(smaller.DimensionNames(), larger.DimensionNames()) {
		broadcasted, err := TensorBroadcastTo(smaller, larger.DimensionNames(), larger.Shape())
		if err != nil {
			return nil, nil, err
		}
		if smaller == a {
			return broadcasted, b, nil
		}
		return a, broadcasted, nil
	}

	return nil, nil, fmt.Errorf("dimension mismatch")
}

// EvaluatePostfixFloat64 evaluates a postfix expression table using a stack-based
// evaluator. It processes each PostfixOp in order: "push" operations load a tensor
// from the controller by name onto the stack; "apply" operations pop the required
// number of operands from the stack, apply the specified operator by dispatching to
// the corresponding tensor operation function (TensorAddFloat64, TensorSubtractFloat64,
// TensorMultiplyFloat64, TensorDivideFloat64, or TensorNegateFloat64), and push the
// result back onto the stack. After all ops are processed, the single remaining
// tensor on the stack is returned as the expression result.
//
// Parameters:
//   - controller: The computation controller used to resolve "push" names and pass
//     to tensor operation functions.
//   - ops: The postfix expression table to evaluate.
//
// Returns:
//   - *ChronoconeTensor[any]: The result tensor, or nil on error.
//   - error: An error if the ops table is empty, the stack underflows, an unknown
//     operator is encountered, or a tensor operation fails.
func EvaluatePostfixFloat64(controller *ChronoconeComputationController, ops []PostfixOp) (*ChronoconeTensor[any], error) {
	if controller == nil {
		return nil, fmt.Errorf("controller is nil")
	}
	if len(ops) == 0 {
		return nil, fmt.Errorf("ops table is empty")
	}

	stack := make([]*ChronoconeTensor[any], 0, len(ops))

	for _, op := range ops {
		switch op.Kind {
		case "push":
			if op.Op == operator.Bool && op.Value == 1.0 {
				// Bool literal true: push as a scalar tensor with mask bit 0 (true/selected).
				tensor := NewChronoconeTensor([]any{1.0}, []string{}, []int{})
				tensor.SetMask(AllTrueMask(1))
				stack = append(stack, tensor)
			} else if op.Op == operator.Bool && op.Value == 0.0 {
				// Bool literal false: push as a scalar tensor with mask bit 1 (false/unselected).
				tensor := NewChronoconeTensor([]any{0.0}, []string{}, []int{})
				tensor.SetMask(AllFalseMask(1))
				stack = append(stack, tensor)
			} else if op.Op == operator.Nil {
				// Nil literal: push as a scalar tensor with mask bit 1 (false/unselected).
				tensor := NewChronoconeTensor([]any{0.0}, []string{}, []int{})
				tensor.SetMask(AllFalseMask(1))
				stack = append(stack, tensor)
			} else if op.Name != "" {
				mc := controller.GetMetricComputation(op.Name)
				if mc == nil {
					return nil, fmt.Errorf("metric computation not found: %q", op.Name)
				}
				tensor := mc.GetTensor()
				if tensor == nil {
					return nil, fmt.Errorf("tensor is nil for metric: %q", op.Name)
				}
				stack = append(stack, tensor)
			} else {
				// Literal value: push as a scalar tensor with empty dimensions.
				tensor := NewChronoconeTensor([]any{op.Value}, []string{}, []int{})
				stack = append(stack, tensor)
			}
		case "apply":
			var result *ChronoconeTensor[any]
			var err error
			switch op.Op {
			case operator.Add, operator.Subtract, operator.Multiply, operator.Divide:
				if len(stack) < 2 {
					return nil, fmt.Errorf("stack underflow for binary operator %s", op.Op.Symbol())
				}
				b := stack[len(stack)-1]
				a := stack[len(stack)-2]
				stack = stack[:len(stack)-2]

				aIsScalar := a.IsScalar()
				bIsScalar := b.IsScalar()

				switch {
				case aIsScalar && bIsScalar:
					// Both scalars: extract float64 values and perform Go arithmetic.
					av, aok := a.Data()[0].(float64)
					bv, bok := b.Data()[0].(float64)
					if !aok || !bok {
						return nil, fmt.Errorf("scalar operand is not float64")
					}
					var rv float64
					switch op.Op {
					case operator.Add:
						rv = av + bv
					case operator.Subtract:
						rv = av - bv
					case operator.Multiply:
						rv = av * bv
					case operator.Divide:
						if bv == 0.0 {
							return nil, fmt.Errorf("division by zero")
						}
						rv = av / bv
					}
					tensor := NewChronoconeTensor([]any{rv}, []string{}, []int{})
					stack = append(stack, tensor)
					continue
				case aIsScalar:
					// Left operand is a scalar literal.
					av, aok := a.Data()[0].(float64)
					if !aok {
						return nil, fmt.Errorf("scalar operand is not float64")
					}
					switch op.Op {
					case operator.Add:
						result, err = TensorAddScalarFloat64(b, av)
					case operator.Subtract:
						result, err = ScalarSubtractTensorFloat64(av, b)
					case operator.Multiply:
						result, err = TensorMultiplyScalarFloat64(b, av)
					case operator.Divide:
						result, err = ScalarDivideTensorFloat64(av, b)
					}
				case bIsScalar:
					// Right operand is a scalar literal.
					bv, bok := b.Data()[0].(float64)
					if !bok {
						return nil, fmt.Errorf("scalar operand is not float64")
					}
					switch op.Op {
					case operator.Add:
						result, err = TensorAddScalarFloat64(a, bv)
					case operator.Subtract:
						result, err = TensorSubtractScalarFloat64(a, bv)
					case operator.Multiply:
						result, err = TensorMultiplyScalarFloat64(a, bv)
					case operator.Divide:
						result, err = TensorDivideScalarFloat64(a, bv)
					}
				default:
					// Both tensors with dimensions.
					a, b, err = broadcastIfNeeded(a, b)
					if err != nil {
						return nil, err
					}
					switch op.Op {
					case operator.Add:
						result, err = TensorAddFloat64(a, b)
					case operator.Subtract:
						result, err = TensorSubtractFloat64(a, b)
					case operator.Multiply:
						result, err = TensorMultiplyFloat64(a, b)
					case operator.Divide:
						result, err = TensorDivideFloat64(a, b)
					}
				}
			case operator.Equal, operator.NotEqual, operator.LessThan, operator.GreaterThan, operator.LessThanOrEqual, operator.GreaterThanOrEqual:
				if len(stack) < 2 {
					return nil, fmt.Errorf("stack underflow for binary operator %s", op.Op.Symbol())
				}
				b := stack[len(stack)-1]
				a := stack[len(stack)-2]
				stack = stack[:len(stack)-2]

				aIsScalar := a.IsScalar()
				bIsScalar := b.IsScalar()

				switch {
				case aIsScalar && bIsScalar:
					// Both scalars: extract float64 values and perform Go comparison.
					av, aok := a.Data()[0].(float64)
					bv, bok := b.Data()[0].(float64)
					if !aok || !bok {
						return nil, fmt.Errorf("scalar operand is not float64")
					}
					stack = append(stack, scalarComparisonResult(op.Op, av, bv))
					continue
				case aIsScalar:
					// Left operand is scalar, right is tensor. Use opposite comparison functions.
					av, aok := a.Data()[0].(float64)
					if !aok {
						return nil, fmt.Errorf("scalar operand is not float64")
					}
					switch op.Op {
					case operator.Equal:
						result, err = TensorEqualScalarFloat64Mask(b, av)
					case operator.NotEqual:
						result, err = TensorNotEqualScalarFloat64Mask(b, av)
					case operator.LessThan:
						// scalar < b is equivalent to b > scalar
						result, err = TensorGreaterThanScalarFloat64Mask(b, av)
					case operator.GreaterThan:
						// scalar > b is equivalent to b < scalar
						result, err = TensorLessThanScalarFloat64Mask(b, av)
					case operator.LessThanOrEqual:
						// scalar <= b is equivalent to b >= scalar
						result, err = TensorGreaterEqualScalarFloat64Mask(b, av)
					case operator.GreaterThanOrEqual:
						// scalar >= b is equivalent to b <= scalar
						result, err = TensorLessEqualScalarFloat64Mask(b, av)
					}
				case bIsScalar:
					// Right operand is scalar. Use scalar comparison functions directly.
					bv, bok := b.Data()[0].(float64)
					if !bok {
						return nil, fmt.Errorf("scalar operand is not float64")
					}
					switch op.Op {
					case operator.Equal:
						result, err = TensorEqualScalarFloat64Mask(a, bv)
					case operator.NotEqual:
						result, err = TensorNotEqualScalarFloat64Mask(a, bv)
					case operator.LessThan:
						result, err = TensorLessThanScalarFloat64Mask(a, bv)
					case operator.GreaterThan:
						result, err = TensorGreaterThanScalarFloat64Mask(a, bv)
					case operator.LessThanOrEqual:
						result, err = TensorLessEqualScalarFloat64Mask(a, bv)
					case operator.GreaterThanOrEqual:
						result, err = TensorGreaterEqualScalarFloat64Mask(a, bv)
					}
				default:
					// Both tensors with dimensions.
					a, b, err = broadcastIfNeeded(a, b)
					if err != nil {
						return nil, err
					}
					switch op.Op {
					case operator.Equal:
						result, err = TensorEqualFloat64Mask(a, b)
					case operator.NotEqual:
						result, err = TensorNotEqualFloat64Mask(a, b)
					case operator.LessThan:
						result, err = TensorLessThanFloat64Mask(a, b)
					case operator.GreaterThan:
						result, err = TensorGreaterThanFloat64Mask(a, b)
					case operator.LessThanOrEqual:
						result, err = TensorLessEqualFloat64Mask(a, b)
					case operator.GreaterThanOrEqual:
						result, err = TensorGreaterEqualFloat64Mask(a, b)
					}
				}
			case operator.And, operator.Or:
				if len(stack) < 2 {
					return nil, fmt.Errorf("stack underflow for binary operator %s", op.Op.Symbol())
				}
				bTensor := stack[len(stack)-1]
				aTensor := stack[len(stack)-2]
				stack = stack[:len(stack)-2]

				aMask := aTensor.Mask()
				bMask := bTensor.Mask()
				if len(aMask) != len(bMask) {
					return nil, fmt.Errorf("mask length mismatch for %s: %d != %d", op.Op.Symbol(), len(aMask), len(bMask))
				}
				switch op.Op {
				case operator.And:
					// Logical AND: combine masks using bitwise OR (inverted convention)
					combinedMask, err := TensorAndMask(aMask, bMask)
					if err != nil {
						return nil, err
					}
					result = NewChronoconeTensor(append([]any{}, aTensor.Data()...), aTensor.DimensionNames(), aTensor.Shape())
					result.SetMask(combinedMask)
				case operator.Or:
					// Logical OR: combine masks using bitwise AND (inverted convention)
					combinedMask, err := TensorOrMask(aMask, bMask)
					if err != nil {
						return nil, err
					}
					result = NewChronoconeTensor(append([]any{}, aTensor.Data()...), aTensor.DimensionNames(), aTensor.Shape())
					result.SetMask(combinedMask)
				}
				stack = append(stack, result)
				continue
			case operator.Not:
				if len(stack) < 1 {
					return nil, fmt.Errorf("stack underflow for unary operator %s", op.Op.Symbol())
				}
				aTensor := stack[len(stack)-1]
				stack = stack[:len(stack)-1]

				// Logical NOT: invert mask using bitwise NOT
				invertedMask := TensorNotMask(aTensor.Mask())
				result = NewChronoconeTensor(append([]any{}, aTensor.Data()...), aTensor.DimensionNames(), aTensor.Shape())
				if invertedMask != nil {
					result.SetMask(invertedMask)
				}
				stack = append(stack, result)
				continue
			case operator.Negate:
				if len(stack) < 1 {
					return nil, fmt.Errorf("stack underflow for unary operator %s", op.Op.Symbol())
				}
				t := stack[len(stack)-1]
				stack = stack[:len(stack)-1]

				if t.IsScalar() {
					fv, ok := t.Data()[0].(float64)
					if !ok {
						return nil, fmt.Errorf("scalar operand is not float64")
					}
					tensor := NewChronoconeTensor([]any{-fv}, []string{}, []int{})
					stack = append(stack, tensor)
					continue
				}
				result, err = TensorNegateFloat64(t)
			default:
				return nil, fmt.Errorf("unknown operator: %s", op.Op.Symbol())
			}
			if err != nil {
				return nil, err
			}
			stack = append(stack, result)
		case "conditional":
			branchResults := make([]*ChronoconeTensor[any], 0, len(op.Branches))
			for i, branch := range op.Branches {
				if len(branch.Body) == 0 {
					return nil, fmt.Errorf("conditional branch %d body is empty", i)
				}
				bodyTensor, err := EvaluatePostfixFloat64(controller, branch.Body)
				if err != nil {
					return nil, fmt.Errorf("conditional branch %d body evaluation failed: %w", i, err)
				}
				var condMask []uint64
				if branch.Condition != nil {
					condTensor, err := EvaluatePostfixFloat64(controller, branch.Condition)
					if err != nil {
						return nil, fmt.Errorf("conditional branch %d condition evaluation failed: %w", i, err)
					}
					if isBareCondition(branch.Condition) {
						// Condition is a bare value with no boolean operator (e.g.,
						// "if data_a" uses a metric reference for truthiness testing).
						// The tensor's mask from NewChronoconeTensor is an identity
						// all-true mask, not a boolean result. Compute the truthiness
						// mask from the raw float64 data: zero = false, non-zero = true.
						condMask = float64TruthinessMaskFromTensor(condTensor)
					} else {
						condMask = condTensor.Mask()
					}
				} else {
					condMask = AllTrueMask(len(bodyTensor.Data()))
				}
				branchTensor := NewChronoconeTensor(append([]any{}, bodyTensor.Data()...), bodyTensor.DimensionNames(), bodyTensor.Shape())
				if condMask != nil {
					branchTensor.SetMask(append([]uint64{}, condMask...))
				}
				branchResults = append(branchResults, branchTensor)
			}
			if len(branchResults) == 0 {
				return nil, fmt.Errorf("conditional has no branches")
			}

			// Broadcast scalar branch results to match the shape of a non-scalar
			// reference branch. A branch body may produce a scalar (e.g., a bare
			// literal constant like 0 in an else branch), which needs to be
			// expanded to fill the same shape as tensor-producing branches before
			// TensorWhereFloat64 can combine them.
			var ref *ChronoconeTensor[any]
			for _, br := range branchResults {
				if !br.IsScalar() {
					ref = br
					break
				}
			}
			if ref != nil {
				for i, br := range branchResults {
					if br.IsScalar() {
						data := make([]any, len(ref.Data()))
						scalarVal := br.Data()[0]
						for j := range data {
							data[j] = scalarVal
						}
						broadcasted := NewChronoconeTensor(data, ref.DimensionNames(), ref.Shape())
						if branchResults[i].mask != nil {
							// The scalar branch has a mask (e.g., an else branch
							// has AllTrueMask(1)). Broadcast the mask intent:
							// an all-true (0) mask means "use these values".
							// Rebuild an all-true mask for the full size.
							broadcasted.SetMask(AllTrueMask(len(data)))
						}
						branchResults[i] = broadcasted
					}
				}
			}

			combined, err := TensorWhereFloat64(branchResults)
			if err != nil {
				return nil, fmt.Errorf("TensorWhereFloat64 failed: %w", err)
			}
			stack = append(stack, combined)
		default:
			return nil, fmt.Errorf("unknown postfix operation kind: %q", op.Kind)
		}
	}

	if len(stack) != 1 {
		return nil, fmt.Errorf("expected exactly 1 result on stack, got %d", len(stack))
	}

	return stack[0], nil
}
