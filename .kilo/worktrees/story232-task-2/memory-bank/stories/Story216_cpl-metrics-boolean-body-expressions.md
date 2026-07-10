# Story216: CPL Metrics Boolean Body Expressions

## Goal

Extend the postfix evaluator and type system to support boolean/logical expressions in conditional branch **body** positions, enabling expressions like:

```
if x > 0
    a and b        # boolean expression in body — not yet supported
else
    a or b         # boolean expression in body — not yet supported
end if
```

Currently Story188 only supports producing float64 data tensors from branch bodies. Boolean expressions (`and`, `or`, `not`) in body positions produce boolean results, which need a separate type path in the evaluator.

## References

- [postfix_op.go](src/chronocone_planning_language/src/main/go/calculation/postfix_op.go:1) — current postfix evaluator (float64 only)
- [chronocone_tensor.go](src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go:1) — tensor type definitions
- [ERO005 metrics synthesis phase](specifications/design/cpl/go-synthesis-target/eros/ERO005_metrics-incremental-implementation.md)
- [Story188_cpl-metrics-conditionals-synthesize.md](Story188_cpl-metrics-conditionals-synthesize.md) — conditional synthesis story

## Dependencies

- [Story188_cpl-metrics-conditionals-synthesize.md](Story188_cpl-metrics-conditionals-synthesize.md) — _Done_ — Provides mask infrastructure, comparison ops, Where, and the conditional postfix evaluator
- [Story227_cpl-conditional-concept-values-synthesize.md](finished-stories/Story227_cpl-conditional-concept-values-synthesize.md) — _Done_ — Conditional concept tensors will have pre-computed validity masks set by synthesis, so boolean body expressions operate only on valid cells without additional work

## Dependent Stories

- [Story227_cpl-conditional-concept-values-synthesize.md](Story227_cpl-conditional-concept-values-synthesize.md) — _Not Started_

## Tasks

1. Add boolean tensor type support - Not Started
   a. Design discriminated stack element type (type tag + union of float64/boolean tensor) - Not Started
   b. Implement boolean tensor representation alongside existing float64 tensors - Not Started
   c. Write tests verifying boolean tensor construction and basic properties - Not Started

2. Add comparison operations producing boolean tensors - Not Started
   a. Implement TensorGreaterThanFloat64Bool, TensorLessThanFloat64Bool, etc. returning ChronoconeTensor[bool] - Not Started
   b. Write tests verifying element-wise comparison produces true/false boolean tensors - Not Started

3. Add logical operations on boolean tensors - Not Started
   a. Implement TensorAndBool, TensorOrBool, TensorNotBool consuming and producing boolean tensors - Not Started
   b. Write tests covering truth tables for AND, OR, NOT on per-element boolean tensors - Not Started

4. Extend postfix evaluator with type-aware stack - Not Started
   a. Modify EvaluatePostfixFloat64 (or create new evaluator) to track type of each stack element - Not Started
   b. Implement type-dispatched operation selection based on operand types - Not Started
   c. Add synthesis-time type error detection for invalid operations (e.g., arithmetic on booleans) - Not Started
   d. Write tests verifying correct dispatch and type error reporting - Not Started

5. Extend walkToPostfix for body-position boolean nodes - Not Started
   a. Ensure walkToPostfix emits comparison ops that produce boolean tensors (not masks) in body context - Not Started
   b. Ensure logical ops (and, or, not) are correctly emitted in body context - Not Started
   c. Write tests verifying postfix output for boolean body expressions - Not Started

6. End-to-end synthesis and verification - Not Started
   a. Create CPL example files with boolean body expressions in conditional metrics - Not Started
   b. Synthesize Go programs and verify they compile and run correctly - Not Started
   c. Write integration tests exercising the full parse-analyze-synthesize pipeline - Not Started

## Notes

### Test-First Development

This project follows the Logical TDD Lifecycle. For each coding task above, tests are written before implementation and are integrated within the task. Do not break tests out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

### Constraints

- Backward compatible with existing float64-only evaluator paths
- Type errors must be caught at synthesis time, not runtime
- Follow Go testing best practices including table-driven tests
- Minimum code that solves the problem; nothing speculative

### Intent

The intent of this story is to complete the conditional expression type system so that both condition positions (which produce masks) and body positions (which produce data tensors, now including booleans) are fully supported. This eliminates the current limitation where boolean expressions are only valid in condition context.

### Acceptance Criteria

- Boolean tensors can be created, manipulated, and stored on the evaluator stack
- Comparison operations in body context produce boolean-valued data tensors
- Logical operations (`and`, `or`, `not`) consume and produce boolean tensors
- The postfix evaluator correctly dispatches operations based on operand types
- Type mismatches (e.g., float64 + boolean) are rejected during synthesis with clear error messages
- End-to-end CPL examples with boolean body expressions compile and execute correctly
- All new code is covered by tests written before implementation

## Footer

### Open Questions

None

### Known Issues

None

### Additional Context

This story extends the work begun in Story188. The key distinction is between **condition context** (comparisons produce uint64 masks for branch selection) and **body context** (comparisons produce boolean data tensors that become the result of the branch).
