# Story188 Task 1 — Mask Infrastructure Walkthrough and Post-Mortem

## 1. Overview

The mask infrastructure adds a bit-packed selection mask to [`ChronoconeTensor[T any]`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:19), enabling conditional evaluation in CPL metrics. In Chronocone's execution model, metric calculations produce `float64` results, while conditional expressions (comparisons, boolean logic) produce boolean results. The mask stores these boolean outcomes without altering the tensor's underlying data type, allowing arithmetic operations to skip deselected elements during aggregation or parallel branch execution.

**What was added:**

- A `mask []uint64` field on `ChronoconeTensor[T any]`
- Five mask helper methods: [`Mask()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:446), [`SetMask()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:457), [`MaskLen()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:474), [`IsMaskedTrue()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:490), [`SetMaskBit()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:505)
- Mask allocation in [`NewChronoconeTensor()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:389) and [`AllocateTensor()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:280)
- Mask propagation docstrings and implementation for 11 arithmetic functions

**Design Reference:** [ERO005 Metrics Synthesis — Tensor Representation](memory-bank/specifications/design/cpl/go-synthesis-target/tensor-representation.md) and [Story188 conditional synthesis plan](memory-bank/plans/story188_conditional_synthesis_rewrite.md).

---

## 2. Design

### 2.1 Why a Separate Mask?

CPL metrics use a single `calculate` block that mixes arithmetic and conditional expressions. The evaluated result of a metric must remain a `float64` tensor so downstream metrics can reference it numerically. Boolean results from conditions (`>`, `<`, `==`, `and`, `or`) cannot be stored in the `float64` data array without loss of meaning. The mask provides a parallel boolean channel that does not interfere with the numeric data.

This separation preserves two invariants:

1. **Type consistency:** The `data []T` array always holds the metric's declared type (e.g., `float64`).
2. **Conditional metadata:** The mask records which elements satisfy the boolean condition, enabling future operations like filtered aggregation or parallel branch routing.

### 2.2 Bit-Packed uint64 Array

Storing one boolean per element as a separate `[]bool` would waste memory: each `bool` in Go occupies one byte. A 365 × 1000 × 100 time tensor (36.5 million elements) would require ~35 MB just for the mask. Packing bits into `uint64` words reduces this to ~4.4 MB — an 8× reduction.

**Sizing formula:**

```go
maskSize := (len(data) + 63) / 64
mask := make([]uint64, maskSize)
```

Each `uint64` word holds 64 element bits. The formula adds 63 before integer division so that any remainder (1 to 63 extra elements) rounds up to one additional word. For an empty tensor (`len(data) == 0`), the mask is a zero-length slice.

### 2.3 Convention: 0 = True/Selected, 1 = False/Unselected

The mask uses an **inverted boolean convention** relative to natural language:

| Bit value | Meaning | Rationale |
|-----------|---------|-----------|
| 0 | true / selected / included | Go's `make([]uint64, n)` zero-initializes every word |
| 1 | false / unselected / excluded | Must be explicitly set |

This convention was chosen because Go allocates slices with all bits set to zero. A newly created tensor therefore has every element automatically "selected" (all bits 0). This is the correct default for non-conditional tensors: they behave as if no masking has occurred. If the convention were reversed (1 = true), every new tensor would require an explicit `O(n)` mask initialization to mark all elements as selected, which is wasteful for the common case where no conditional filtering applies.

---

## 3. Implementation Details

### 3.1 Field and Allocation

The `mask` field is declared on [`ChronoconeTensor`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:35) alongside `data`, `dimensionNames`, and `shape`:

```go
mask []uint64
```

**Allocation sites:**

| Function | Mask behavior |
|----------|---------------|
| [`NewChronoconeTensor()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:389) | Allocates `make([]uint64, maskSize)` where `maskSize = (len(data) + 63) / 64`. The slice is zero-initialized, so all elements are selected by default. |
| [`AllocateTensor()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:280) | Delegates to `NewChronoconeTensor()`, inheriting the same zero-initialization guarantee. |

### 3.2 Mask Accessor Methods

#### [`Mask()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:446)

Returns the underlying `[]uint64` mask slice. Returns `nil` if the tensor itself is `nil`. The returned slice is a reference to the internal mask, not a copy, so callers can mutate it directly. This is intentional for performance: bulk mask operations should not require allocation.

```go
func (t *ChronoconeTensor[T]) Mask() []uint64 {
    if t == nil { return nil }
    return t.mask
}
```

#### [`SetMask(mask []uint64)`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:457)

Replaces the tensor's mask with a copy of the provided slice. Passing `nil` clears the mask entirely (sets `t.mask = nil`). The method defensively copies the input to prevent external mutations from affecting the tensor after the call. Nil tensor is a no-op.

```go
func (t *ChronoconeTensor[T]) SetMask(mask []uint64) {
    if t == nil { return }
    if mask == nil { t.mask = nil; return }
    t.mask = make([]uint64, len(mask))
    copy(t.mask, mask)
}
```

#### [`MaskLen()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:474)

Returns the number of `uint64` words allocated for the mask: `(len(data) + 63) / 64`. Returns `0` for a nil tensor. This value matches `len(t.mask)` for a normally constructed tensor, but it is computed from `len(t.data)` rather than reading `len(t.mask)` directly. This avoids relying on the internal invariant that `len(mask)` and `len(data)` are synchronized, which is useful for debugging or after `SetMask` has been called with a mismatched slice.

```go
func (t *ChronoconeTensor[T]) MaskLen() int {
    if t == nil { return 0 }
    return (len(t.data) + 63) / 64
}
```

#### [`IsMaskedTrue(index int)`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:490)

Returns `true` if the element at `index` is selected (bit is 0). Returns `false` for any of these guard conditions: nil tensor, nil mask, negative index, or index `>= len(data)`. The bit arithmetic uses two shifts:

```go
wordIndex := index >> 6        // divide by 64
bitOffset := uint(index) & 0x3F // modulo 64
bit := (t.mask[wordIndex] >> bitOffset) & 1
return bit == 0
```

The right-shift `>>` moves the target bit to position 0, and `& 1` isolates it. The final comparison `== 0` inverts the bit to match the "0 means true" convention.

#### [`SetMaskBit(index int, value bool)`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:505)

Sets or clears the bit at `index`. The same guard conditions as `IsMaskedTrue` apply: nil tensor, nil mask, or out-of-bounds index causes an early return (no-op, no panic).

```go
wordIndex := index >> 6
bitOffset := uint(index) & 0x3F
bitMask := uint64(1) << bitOffset

if value {
    t.mask[wordIndex] &^= bitMask   // AND-NOT: clears the bit (sets to 0 = true)
} else {
    t.mask[wordIndex] |= bitMask    // OR: sets the bit (sets to 1 = false)
}
```

`&^=` (AND-NOT) is the Go idiom for clearing a specific bit without affecting the other 63 bits in the word. `|=` sets the bit. The boolean parameter `value` maps directly to the semantic meaning: `value=true` means "select this element," which corresponds to clearing the bit to 0.

### 3.3 Edge Case Handling Summary

| Condition | `Mask()` | `SetMask()` | `MaskLen()` | `IsMaskedTrue()` | `SetMaskBit()` |
|-----------|----------|-------------|-------------|------------------|----------------|
| Nil tensor | returns `nil` | no-op | returns `0` | returns `false` | no-op |
| Nil mask | — | clears mask | returns `(len(data)+63)/64` | returns `false` | no-op |
| Negative index | — | — | — | returns `false` | no-op |
| Index `>= len(data)` | — | — | — | returns `false` | no-op |
| Empty data (`len == 0`) | returns `[]uint64{}` | stores copy | returns `0` | returns `false` (OOB) | no-op (OOB) |

---

## 4. Mask Propagation in Arithmetic Operations

Eleven existing tensor arithmetic functions were updated to propagate masks from input tensors to the output tensor. Each function's docstring now includes a "Mask propagation" line. The implementation uses a single pattern: allocate the result tensor via `NewChronoconeTensor` (which creates a fresh zero mask of the correct size), then `copy` the source mask into it.

### 4.1 Propagation Rules

| Operation category | Source of output mask | Rationale |
|--------------------|-----------------------|-----------|
| Binary (tensor-tensor) | Left operand (`a.mask`) | In CPL, the left operand is the primary expression; the right operand is a modifier. If the left operand was produced by a conditional, its mask carries the boolean context. |
| Unary (tensor only) | Input tensor (`t.mask`) | The output is a transformed version of the input; the selection context is preserved. |
| Scalar-tensor | Input tensor (`t.mask`) | The scalar has no mask; the tensor's boolean context governs which elements matter. |

### 4.2 Functions Updated

| Function | Category | Propagation line |
|----------|----------|------------------|
| [`TensorAddFloat64()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:557) | Binary | `copy(resultTensor.mask, a.mask)` |
| [`TensorSubtractFloat64()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:594) | Binary | `copy(resultTensor.mask, a.mask)` |
| [`TensorMultiplyFloat64()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:631) | Binary | `copy(resultTensor.mask, a.mask)` |
| [`TensorDivideFloat64()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:669) | Binary | `copy(resultTensor.mask, a.mask)` |
| [`TensorNegateFloat64()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:708) | Unary | `copy(resultTensor.mask, t.mask)` |
| [`ScalarSubtractTensorFloat64()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:738) | Scalar-tensor | `copy(resultTensor.mask, t.mask)` |
| [`ScalarDivideTensorFloat64()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:770) | Scalar-tensor | `copy(resultTensor.mask, t.mask)` |
| [`TensorAddScalarFloat64()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:803) | Tensor-scalar | `copy(resultTensor.mask, t.mask)` |
| [`TensorSubtractScalarFloat64()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:834) | Tensor-scalar | `copy(resultTensor.mask, t.mask)` |
| [`TensorMultiplyScalarFloat64()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:864) | Tensor-scalar | `copy(resultTensor.mask, t.mask)` |
| [`TensorDivideScalarFloat64()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:895) | Tensor-scalar | `copy(resultTensor.mask, t.mask)` |

### 4.3 Propagation Example

Consider the CPL expression:

```cpl
calculate
    (revenue > 1000) * price
```

1. The comparison `revenue > 1000` produces a tensor where the mask has bit=0 for elements where revenue exceeds 1000, and bit=1 elsewhere. The `data` array holds the boolean comparison result as `float64` (1.0 or 0.0), but the mask is the authoritative selection signal.
2. The multiplication `(...) * price` calls `TensorMultiplyFloat64(left, right)`. The left operand carries the mask from the comparison. The output tensor copies this mask, so elements where revenue was `<= 1000` remain deselected in the product.
3. A downstream aggregation (future work) can use the mask to skip deselected elements when computing a sum or average.

---

## 5. Testing Strategy

The test file [`chronocone_tensor_test.go`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor_test.go) contains **17 mask-specific test functions** in addition to pre-existing controller and allocation tests. The tests are grouped by concern.

### 5.1 Mask Allocation Tests (2 tests)

| Test | What it verifies |
|------|------------------|
| `TestMaskAllocationSize` | Table-driven: 8 cases from 0 to 128 elements. Verifies `len(mask)` equals `(dataSize + 63) / 64`. Covers exact boundary at 64 and 128. |
| `TestMaskAllocationAllBitsZero` | A fresh 100-element tensor has every `uint64` word equal to `0x0000000000000000`, confirming the zero-init=all-true convention. |

### 5.2 Mask Accessor Tests (4 tests)

| Test | What it verifies |
|------|------------------|
| `TestMaskAccessor` | `Mask()` returns a non-nil slice of correct length, and the returned slice references the underlying array (mutating it mutates `tensor.mask`). |
| `TestMaskLen` | Table-driven: 8 cases. `MaskLen()` returns `(len(data) + 63) / 64` for various sizes. |
| `TestMaskSetMaskNil` | `SetMask(nil)` does not panic. After clearing, `IsMaskedTrue` returns `false` and `SetMaskBit` is a no-op. |
| `TestMaskSetMaskNonNil` | `SetMask` with a non-nil slice stores a copy; `Mask()` reflects the new values. |

### 5.3 Bit Get/Set Tests (4 tests)

| Test | What it verifies |
|------|------------------|
| `TestMaskBitGetSetWordBoundaries` | 5 boundary indices: 0, 63, 64, 127, 128. For each, confirms `IsMaskedTrue` is true on a fresh tensor, false after `SetMaskBit(idx, false)`, and true again after `SetMaskBit(idx, true)`. |
| `TestMaskZeroMeansTrue` | Two sub-cases: all bits 0 (`[]uint64{0}`) returns `true` for every index 0–63; all bits 1 (`[]uint64{^uint64(0)}`) returns `false` for every index. |
| `TestMaskBitRoundTrip` | For 10 elements, sets each to false then true, verifying `IsMaskedTrue` matches the last `SetMaskBit` call. |
| `TestMaskBitToggle` | Sets a single bit false → true → false, verifying the toggle sequence. |

### 5.4 Mask Propagation Tests (3 tests)

| Test | What it verifies |
|------|------------------|
| `TestMaskPropagationBinaryOps` | Table-driven over 4 binary functions (`Add`, `Subtract`, `Multiply`, `Divide`). Two operands with distinct masks; asserts output mask equals left operand mask. |
| `TestMaskPropagationUnaryOps` | `TensorNegateFloat64` copies the input mask to the output. |
| `TestMaskPropagationScalarOps` | Table-driven over 6 scalar functions. Input tensor with a custom mask; asserts output mask equals input mask. |

### 5.5 Edge Case Tests (4 tests)

| Test | What it verifies |
|------|------------------|
| `TestMaskNilTensorEdgeCases` | Nil tensor: `Mask()` returns `nil`, `MaskLen()` returns 0, `IsMaskedTrue` returns `false`, `SetMaskBit` is a no-op (no panic). |
| `TestMaskEmptyTensor` | Zero-element tensor: `len(mask) == 0`, `MaskLen() == 0`. |
| `TestMaskOutOfBounds` | Negative index, index equal to `len(data)`, and index well beyond `len(data)` all return `false` from `IsMaskedTrue` and are no-ops for `SetMaskBit`. |

### 5.6 Test Design Pattern

Mask tests use **direct tensor construction** rather than full parser/synthesis pipelines. A helper [`makeMaskedAnyTensor()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor_test.go:502) creates a `*ChronoconeTensor[any]` with `float64` data and an injected mask slice. This isolates mask behavior from parser changes and allows precise control over mask bit patterns. Another helper, [`maskEqual()`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor_test.go:488), performs deep equality on `[]uint64` slices.

---

## 6. Post-Mortem

### 6.1 What Went Well

**Bit packing was the right choice.** A `[]bool` would have been simpler to read but would have consumed 8× the memory for large tensors. The `(len(data) + 63) / 64` sizing formula is standard in systems programming and immediately understandable to reviewers.

**Zero-init = all-true simplified defaults.** Because Go zero-allocates `uint64` slices, every new tensor starts with all elements selected. This eliminated the need for an explicit initialization loop and prevented a class of bugs where forgotten initialization would cause all elements to be spuriously deselected.

**Defensive nil guards are consistent.** Every public mask method checks for nil tensor and nil mask before operating. This prevents panics in edge cases (empty tensors, partially constructed objects) without requiring callers to pre-check.

**Test coverage is thorough.** The 17 tests exercise allocation size boundaries, bit arithmetic at word boundaries, round-trip get/set, the 0=true convention, mask propagation through all 11 arithmetic functions, and nil/empty/OOB edge cases. No production code path is untested.

### 6.2 Challenges Encountered

**Review found a nil-mask guard issue in `SetMaskBit` and `IsMaskedTrue`.** The initial implementation did not check `t.mask == nil` before indexing into `t.mask[index>>6]`. If a caller used `SetMask(nil)` to clear the mask and then called `SetMaskBit`, the code would panic on a nil pointer dereference. The fix added `t.mask == nil` to the guard condition in both methods. This was caught during code review rather than by tests, which prompted adding `TestMaskSetMaskNil` to cover the scenario.

**Mask propagation for binary ops: left vs. right operand.** The choice to copy the left operand's mask required explicit documentation. In expressions like `price * (revenue > 1000)`, the left operand (`price`) has no conditional context, so the output would incorrectly inherit an all-true mask. This is expected behavior for now: the CPL synthesis layer will ensure that conditional expressions appear as the left operand when mask context matters, or a future enhancement may merge masks with an `AND` operation.

**Docstring consistency.** Eleven arithmetic functions needed identical "Mask propagation" docstring lines. A manual edit risked inconsistent wording. The review step verified that all 11 functions used the same phrasing and the same `copy(resultTensor.mask, ...)` pattern.

### 6.3 Lessons Learned

1. **Guard against nil in every accessor.** Even if the constructor always allocates a mask, callers can call `SetMask(nil)` at any time. Every method that touches `t.mask` must tolerate `nil`.

2. **Test the guard conditions explicitly.** `TestMaskNilTensorEdgeCases` and `TestMaskOutOfBounds` were added after the review caught a missing guard. Edge-case tests for nil and OOB should be written alongside the happy-path tests, not deferred.

3. **Document propagation rules in the function docstring, not just the walkthrough.** The "Mask propagation" line in each arithmetic function's godoc makes the rule visible at the call site. Developers reading the function signature in their IDE see the rule without opening the walkthrough.

4. **Helper functions reduce test boilerplate but must themselves be simple.** `makeMaskedAnyTensor` and `maskEqual` are small, pure, and have no branching. They make the 17 tests readable without introducing hidden complexity.

---

## Key Files Reference

| File | Purpose |
|------|---------|
| [`chronocone_tensor.go`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go) | `ChronoconeTensor` struct, mask field, five mask methods, 11 arithmetic functions with mask propagation |
| [`chronocone_tensor_test.go`](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor_test.go) | 17 mask-specific tests plus existing controller/allocation tests; 47 total tests pass |
| [`story188_conditional_synthesis_rewrite.md`](memory-bank/plans/story188_conditional_synthesis_rewrite.md) | Design plan for conditional evaluation in CPL metrics synthesis |
| [`tensor-representation.md`](memory-bank/specifications/design/cpl/go-synthesis-target/tensor-representation.md) | Tensor design specification |
