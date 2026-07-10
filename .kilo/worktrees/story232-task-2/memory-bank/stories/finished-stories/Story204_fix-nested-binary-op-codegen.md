# Story204: Fix Nested Binary Op Code Generation

## Goal

Fix the generated Go code for nested arithmetic expressions (e.g., `input_a - input_b / c`) that currently produces invalid Go syntax by embedding multi-line if/error blocks inside function call expressions. Replace per-operation statement block generation with a table-driven postfix evaluator that walks the CIL tree in post-order, keeping the CIL tree as the canonical intermediate representation.

This story is part of [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md).

## References

- [Fix Synthesis Plan](../plans/fix_synthesis_plan.md)
- [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md)
- [calculation/chronocone_tensor.go](../../../src/chronocone_planning_language/src/main/go/calculation/chronocone_tensor.go)
- [cil/cil_transformer.go](../../../src/chronocone_planning_language/src/main/go/cil/cil_transformer.go)
- [cil/cil_transformer_test.go](../../../src/chronocone_planning_language/src/main/go/cil/cil_transformer_test.go)

## Dependencies

- [Story185_cpl-metrics-arithmetic-synthesize.md](Story185_cpl-metrics-arithmetic-synthesize.md) - Done
- [Story203_fix-synthesis-bugs-arithmetic-metrics.md](Story203_fix-synthesis-bugs-arithmetic-metrics.md) - Done

Both dependencies must be complete because they establish the existing CIL tree, tensor operations, and generated-code compilation path that this story modifies.

## Dependent Stories

- [Story202_cpl-metrics-arithmetic-examples-revamp.md](Story202_cpl-metrics-arithmetic-examples-revamp.md) - Not Started

## Tasks

1. Add `PostfixOp` type and `EvaluatePostfixFloat64` to `calculation` package - Completed
   a. Define `PostfixOp` struct with `Kind`, `Op`, and `Name` fields and doc comments - Completed
   b. Define `EvaluatePostfixFloat64` function signature with doc comments - Completed
   c. Write tests for `EvaluatePostfixFloat64` covering single ops, nested ops, dimension mismatches, empty table, stack underflow, unknown operator, and nil inputs - Completed
   d. Implement `EvaluatePostfixFloat64` using a `[]*ChronoconeTensor[any]` stack, dispatching to existing `TensorAddFloat64`, `TensorSubtractFloat64`, `TensorMultiplyFloat64`, and `TensorDivideFloat64` functions - Completed

2. Add CIL tree-walk to postfix table generation in `cil` package - Completed
   a. Define `WalkCILTreeToPostfix` and `walkToPostfix` method signatures with doc comments - Completed
   b. Define `GeneratePostfixTableLiteral` method signature with doc comments - Completed
   c. Write tests for `WalkCILTreeToPostfix` covering simple binary ops, nested expressions, unary negation, scalar literal operands, and input references - Completed
   d. Write tests for `GeneratePostfixTableLiteral` verifying valid Go struct literal output - Completed
   e. Implement `walkToPostfix` recursive helper: `CILLiteral` -> pushLiteral, `CILInputRef` -> push, `CILBinaryOp` -> recurse left/right then apply, `CILUnaryOp` -> recurse operand then apply - Completed
   f. Implement `GeneratePostfixTableLiteral` formatting `[]calculation.PostfixOp{{...}}` with proper quoting - Completed

3. Update `TransformMetric` to emit postfix tables for arithmetic expressions - Completed
   a. Update `TransformMetric` switch on `metricNode.CILNode.(type)` to route `*CILBinaryOp` and `*CILUnaryOp` through `WalkCILTreeToPostfix` -> `GeneratePostfixTableLiteral` -> emit table + `EvaluatePostfixFloat64` call - Completed
   b. Handle `*CILInputRef` as a single-entry postfix table with one push - Completed
   c. Keep existing `*CILLiteral` `isLiteral` path unchanged (Increment 1) - Completed
   d. Write tests for updated `TransformMetric` verifying table-driven output, literal path still works, and `SetTensor` call remains present - Completed
   e. Update existing `cil_transformer_test.go` expectations: mark old snippet-generation visitor tests as legacy, add `TestWalkCILTreeToPostfix` and `TestWalkCILTreeToPostfix_Nested` - Completed

4. End-to-end verification - Completed
   a. Run `bash shell_scripts/compile_and_run.sh subtraction_division` and confirm compilation and execution without errors - Completed
   b. Run `bash shell_scripts/compile_and_run.sh simple_addition` and confirm it still compiles and runs correctly - Completed
   c. Run full test suite: `cd src/chronocone_planning_language && go test ./...` and confirm all tests pass - Completed

## Constraints

- The CIL tree remains the sole intermediate representation; no data may flow directly from `CalculatePostfix` to code generation.
- All existing tensor operation functions (`TensorAddFloat64`, `TensorSubtractFloat64`, `TensorMultiplyFloat64`, `TensorDivideFloat64`) must be reused; no duplication of dimension-check or arithmetic logic.
- Generated code must remain valid Go that compiles without syntax errors for arbitrarily nested expressions.
- Old snippet-generation visitor methods (`VisitCILBinaryOp`, `VisitCILUnaryOp`, etc.) may be retained for backward compatibility but must not be invoked during arithmetic synthesis after this change.

## Intent

The generated code for arithmetic metrics should be declarative, concise, and free of nested statement blocks inside function arguments. Centralizing the postfix evaluation in a runtime function eliminates repeated per-operation boilerplate and makes the CIL transformer simpler.

## Acceptance Criteria

1. `EvaluatePostfixFloat64` exists in the `calculation` package with tests that exercise all four arithmetic operators, nested expressions, and error paths.
2. `WalkCILTreeToPostfix` and `GeneratePostfixTableLiteral` exist in the `cil` package with tests covering literals, input refs, binary ops, unary ops, and nested trees.
3. `TransformMetric` emits a `[]calculation.PostfixOp` literal and a single `calculation.EvaluatePostfixFloat64` call for arithmetic expressions.
4. The `subtraction_division` example compiles and runs without errors.
5. The `simple_addition` example continues to compile and run correctly.
6. All existing tests in `cil_transformer_test.go` and `calculation` pass.

## Open Questions

None.

## Known Issues

- Current generated code for `subtraction_division` produces syntactically invalid Go where an entire `if` block is embedded as a function argument.
- This bug only manifests for nested binary operations; single-operation expressions are unaffected.

## Additional Context

This plan complements the fixes in Story203 (type mismatch and measure reference `Kind` bugs). It uses the same existing tensor operation functions that Story203/Step 2 implemented, and the postfix table naturally uses the correct input names resolved by Story203/Step 3-4.
