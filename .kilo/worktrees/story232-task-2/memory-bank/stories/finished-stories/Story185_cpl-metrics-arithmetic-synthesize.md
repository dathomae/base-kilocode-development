# Story185: CPL Metrics Arithmetic Synthesize

## Goal

Implement the Increment 2 synthesis phase for CPL metric arithmetic expressions as part of [Epic005: CPL Metrics Implementation](../../epics/Epic005_cpl-metrics-implementation.md). This story covers:

- Extend CIL tree walker with `VisitCILBinaryOp` (handles scalar/tensor operand combinations, emits high-level tensor op methods: `Add`, `SubtractScalar`, `Multiply`, etc.), `VisitCILUnaryOp` (negation), and `VisitCILInputRef` (accesses input tensors via `MetricComputation.GetTensor()` or `inputs` map).
- Use IO spec output metrics (already loaded and validated by Story184) to determine which metrics to compute.
- Generated code performs element-wise arithmetic on tensors.

## References

- [Epic005: CPL Metrics Implementation](../../epics/Epic005_cpl-metrics-implementation.md)
- [ERO005: Metrics Synthesis Phase — Increment 2](../../specifications/design/cpl/go-synthesis-target/eros/ERO005_metrics-synthesis-phase.md)
- [Synthesis Design](../../specifications/design/cpl/go-synthesis-target/synthesis.md)
- [Tensor Representation Design](../../specifications/design/cpl/go-synthesis-target/tensor-representation.md)
- [IO Specifications Requirements](../../specifications/requirements/cpl/io_specs.md)
- [Go Testing Best Practices Guide](../../../resources/go/testing-guide.md)

## Dependencies

- [Story184_cpl-metrics-arithmetic-parse-analyze.md](Story184_cpl-metrics-arithmetic-parse-analyze.md) — parse/analyze phase must produce valid `CILBinaryOp`, `CILUnaryOp`, and `CILInputRef` nodes with validated IO spec (Done)
- [Story183_cpl-metrics-basic-synthesize.md](Story183_cpl-metrics-basic-synthesize.md) — basic synthesizer infrastructure must be in place (Done)

## Dependent Stories

- [Story186_cpl-metrics-parallelism-synthesize.md](../../stories/Story186_cpl-metrics-parallelism-synthesize.md) — Increment 3 Synthesis (Not Started)

## Tasks

1. Implement CIL binary operation visitor — Completed
   a. Define `VisitCILBinaryOp(ctx *CILContext, node *CILBinaryOp) error` method on `CILAbstractTransformer` — Completed
   b. Write tests for binary op code generation: scalar-scalar, scalar-tensor, tensor-scalar, tensor-tensor combinations for each operator (+, -, \*, /) — Completed
   c. Implement `VisitCILBinaryOp`: determine operand types (scalar/tensor), emit appropriate high-level tensor operation method calls (`Add`, `SubtractScalar`, `Multiply`, `Divide`, etc.) — Completed
   d. Handle tensor shape compatibility: generated code must verify tensor dimensions match before operations — Completed

2. Implement CIL unary operation visitor — Completed
   a. Define `VisitCILUnaryOp(ctx *CILContext, node *CILUnaryOp) error` method — Completed
   b. Write tests for unary negation code generation — Completed
   c. Implement `VisitCILUnaryOp`: emit negation code (`Negate` tensor method call) — Completed

3. Implement CIL input reference visitor — Completed
   a. Define `VisitCILInputRef(ctx *CILContext, node *CILInputRef) error` method — Completed
   b. Write tests for input ref code generation: verify correct tensor access via `MetricComputation.GetTensor()` or `inputs` map — Completed
   c. Implement `VisitCILInputRef`: emit code to access input tensor by name — Completed

4. Wire IO spec output metrics into compute function generation — Completed
   a. Use IO spec output metrics (already loaded and validated) to determine which metrics to generate compute functions for — Completed
   b. Generate code that only computes metrics listed in IO spec output section — Completed
   c. Write integration tests: verify IO spec correctly filters generated compute functions — Completed

5. End-to-end verification — Completed
   a. Create example CPL files with arithmetic metric calculations (e.g., `calculate revenue - cost`) — Completed
   b. Create corresponding IO specs and test parquet data — Completed
   c. Run parse → analyze → synthesize → compile → execute pipeline and verify arithmetic correctness — Completed
   d. Run the full CPL test suite to confirm no regressions — Completed

## Constraints

1. Follow the [Logical TDD Lifecycle](../../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. Generated code must use the [`CILAbstractTransformer`](../../specifications/design/cpl/go-synthesis-target/synthesis.md#cilabstracttransformer) infrastructure
3. Element-wise tensor operations must handle scalar broadcast correctly
4. Generated Go code must be compilable and runnable
5. Go functions that can fail must return `(T, error)`
6. IO spec loading and basic integration already exist — this story adds metric-specific output filtering

## Intent

- **Task 1 Intent**: Generate Go code for binary arithmetic operations on tensors — the core computation engine
- **Task 2 Intent**: Support unary negation as a building block for more complex expressions
- **Task 3 Intent**: Connect generated code to input tensors from connector parquet data
- **Task 4 Intent**: Control which metrics the generated program computes based on IO specification output section
- **Task 5 Intent**: Prove arithmetic pipeline works end-to-end with correct numeric output

## Acceptance Criteria

1. `VisitCILBinaryOp` generates correct Go code for all four arithmetic operators with scalar/tensor operand combinations
2. `VisitCILUnaryOp` generates correct negation code
3. `VisitCILInputRef` generates correct tensor access code
4. IO spec output metrics correctly determine which metrics are computed and output
5. Full end-to-end test: arithmetic expressions produce correct numeric results
6. All unit tests pass following Go testing best practices
7. No regressions in existing CPL test suite
