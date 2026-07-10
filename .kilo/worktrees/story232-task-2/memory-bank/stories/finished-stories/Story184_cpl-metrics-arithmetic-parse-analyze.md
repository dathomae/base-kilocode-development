# Story184: CPL Metrics Arithmetic Parse and Analyze

## Goal

Implement the Increment 2 parsing and analysis phases for CPL metric arithmetic expressions as part of [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md). This story covers:

- Extend root grammar with shared expression rules (`expression`, `additionExpr`, `multiplicationExpr`, `unaryExpr`, `primaryExpr`, `literal`) and arithmetic operator tokens (`+`, `-`, `*`, `/`).
- Root grammar listener pushes operand definitions (identifiers, numeric/string literals) and operator definitions (`PlusOperatorDefinition`, `MultiplyOperatorDefinition`, etc.) in postfix order, scope-type-aware via `shouldPushExpressionDefs`.
- Update [`CPLMetrics.g4`](src/chronocone_planning_language/src/main/antlr/CPLMetrics.g4) `calculateClause` to use shared `expression` rule.
- Extend analysis postfix evaluator to construct `CILBinaryOp`, `CILUnaryOp`, and `CILInputRef` nodes. Type-check arithmetic operands.
- Validate IO specification output metrics against declared entities. IO spec loading already exists in [`io_specification.go`](src/chronocone_planning_language/src/main/go/io/io_specification.go) — this story adds validation that output metric names exist in declared metrics and stores the validated IO spec for synthesis.

## References

- [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md)
- [Metrics Syntax Requirements](../specifications/requirements/cpl/metrics-syntax.md)
- [IO Specifications Requirements](../specifications/requirements/cpl/io_specs.md)
- [Metrics Design](../specifications/design/cpl/metrics-design.md)
- [ERO005: Metrics Parsing Phase — Increment 2](../specifications/design/cpl/go-parser-target/eros/ERO005_metrics-parsing-phase.md)
- [ERO005: Metrics Analysis Phase — Increment 2](../specifications/design/cpl/go-analysis/eros/ERO005_metrics-analysis-phase.md)
- [CPL Architecture — Intermediate Language](../specifications/architecture/cpl/analysis/intermediate_language.md)
- [Postfix Expression Parsing Pattern](../specifications/patterns/postfix-expression-parsing.md)
- [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Dependencies

- [Story182_cpl-metrics-basic-parse-analyze.md](Story182_cpl-metrics-basic-parse-analyze.md) — Increment 1 Parse+Analyze must be complete (Not Started)
- [Story183_cpl-metrics-basic-synthesize.md](Story183_cpl-metrics-basic-synthesize.md) — Increment 1 Synthesis must be complete (Not Started)

## Dependent Stories

- [Story185_cpl-metrics-arithmetic-synthesize.md](Story185_cpl-metrics-arithmetic-synthesize.md) — Increment 2 Synthesis (Not Started)
- [Story186_cpl-metrics-parallelism-synthesize.md](Story186_cpl-metrics-parallelism-synthesize.md) — Increment 3 Synthesis (parallel development possible after this story) (Not Started)
- [Story187_cpl-metrics-conditionals-parse-analyze.md](Story187_cpl-metrics-conditionals-parse-analyze.md) — Increment 4 Parse+Analyze (parallel development possible after this story) (Not Started)
- [Story189_cpl-metrics-aliases-where-parse-analyze.md](Story189_cpl-metrics-aliases-where-parse-analyze.md) — Increment 5 Parse+Analyze (parallel development possible after this story) (Not Started)
- [Story191_cpl-metrics-time-shift-parse-analyze.md](Story191_cpl-metrics-time-shift-parse-analyze.md) — Increment 6 Parse+Analyze (parallel development possible after this story) (Not Started)
- [Story193_cpl-metrics-short-form-parse-analyze.md](Story193_cpl-metrics-short-form-parse-analyze.md) — Increment 7 Parse+Analyze (parallel development possible after this story) (Not Started)

## Tasks

1. Extend root grammar with shared arithmetic expression rules — Completed
   a. Add `expression` rule: `additionExpr` (left-associative `+`, `-`) — Completed
   b. Add `additionExpr` rule: `multiplicationExpr` with `+` and `-` operators — Completed
   c. Add `multiplicationExpr` rule: `unaryExpr` with `*` and `/` operators — Completed
   d. Add `unaryExpr` rule: optional `-` prefix + `primaryExpr` — Completed
   e. Add `primaryExpr` rule: `literal` | `identifier` | `LPAREN expression RPAREN` — Completed
   f. Add arithmetic operator lexer tokens (`PLUS: '+'`, `MINUS: '-'`, `MULTIPLY: '*'`, `DIVIDE: '/'`) — Completed
   g. Regenerate ANTLR parser/lexer Go files — Completed

2. Implement root grammar listener for expression rules — Completed
   a. Define operator definitions: `PlusOperatorDefinition`, `MinusOperatorDefinition`, `MultiplyOperatorDefinition`, `DivideOperatorDefinition`, `NegateOperatorDefinition` — Completed
   b. Implement `ExitPrimaryExpr` — for identifier operands, push identifier name as `OperandDefinition`; for literals, push `LiteralOperandDefinition` — Completed
   c. Implement `ExitAdditionExpr` / `ExitMultiplicationExpr` — push operator definitions in postfix order — Completed
   d. Implement `ExitUnaryExpr` — push `NegateOperatorDefinition` if unary minus present — Completed
   e. Implement `shouldPushExpressionDefs` to control when expression definitions are pushed (based on current scope type) — Completed
   f. Verify parser tests pass: parse arithmetic expressions and assert correct postfix operator sequences — Completed

3. Update CPLMetrics.g4 calculate clause — Completed
   a. Change `calculateClause` to use shared `expression` rule instead of just `literal` — Completed
   b. Update `MetricsListener` to handle expression postfix in calculate clause — Completed
   c. Verify parser tests: parse metrics with arithmetic expressions (e.g., `calculate a + b * c`), assert correct postfix on `ParserMetricDefinition` — Completed

4. Extend analysis postfix evaluator for arithmetic — Completed
   a. Define `CILBinaryOp` node with fields: `Op string` (+, -, \*, /), `Left CILNode`, `Right CILNode` — Completed
   b. Define `CILUnaryOp` node with field: `Op string` (negate), `Operand CILNode` — Completed
   c. Define `CILInputRef` node with field: `Name string` (references an input connector or metric) — Completed (pre-existing)
   d. Implement postfix evaluator: process the postfix slice from `ParserMetricDefinition`, build CIL tree for arithmetic expressions — Completed
   e. Implement type checking: verify operands in arithmetic expressions are numeric types — Completed
   f. Verify analyzer tests: CIL tree correctly represents parsed arithmetic expressions — Completed

5. Implement IO specification validation against metrics — Completed
   a. Validate output metric names: verify each output metric in IO spec exists in declared metrics — Completed
   b. Store validated IO spec on analysis context for synthesis phase use — Completed
   c. Verify IO spec validation tests: reject unknown metrics, accept valid output selections — Completed

6. Integration verification — Completed
   a. Create example CPL files with arithmetic expressions in calculate clauses — Completed
   b. Create corresponding IO spec examples — Completed
   c. Verify parse-analyze pipeline produces correct CIL trees with binary/unary ops and input refs — Completed
   d. Run the full CPL test suite to confirm no regressions — Completed

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. Expression rules live in the root grammar (`ChronoconePlanningLanguage.g4`) since they are shared across CPL constructs
3. Parse/analysis separation must be maintained: parser captures postfix operations only; analysis builds CIL trees from postfix
4. Follow the [Postfix Expression Parsing Pattern](../specifications/patterns/postfix-expression-parsing.md)
5. All CIL nodes must implement the `CILNode` interface
6. Types are initially `float64`; detailed type selection is deferred
7. Generated Go tests must follow the [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)
8. IO spec loading already exists — this story adds validation, not loading

## Intent

- **Task 1 Intent**: Define the shared expression grammar in the root grammar so all CPL constructs can use arithmetic
- **Task 2 Intent**: Implement expression parsing via postfix — the parser pushes operands and operators in postfix order into the current scope
- **Task 3 Intent**: Connect metrics calculate clause to the shared expression grammar, replacing the literal-only approach from Increment 1
- **Task 4 Intent**: Build the CIL tree from postfix sequences — convert operator/operand sequences into proper `CILBinaryOp`/`CILUnaryOp`/`CILInputRef` tree structures with type checking
- **Task 5 Intent**: Validate IO spec output metrics against declared entities — ensure synthesis receives consistent configuration
- **Task 6 Intent**: Verify the complete arithmetic pipeline works end-to-end

## Acceptance Criteria

1. Root grammar correctly parses arithmetic expressions with `+`, `-`, `*`, `/`, unary minus, parentheses, and operator precedence
2. Root grammar listener pushes correct postfix sequences for all expression types
3. `CPLMetrics.g4` calculate clause accepts arbitrary arithmetic expressions
4. Postfix evaluator constructs correct `CILBinaryOp`, `CILUnaryOp`, and `CILInputRef` CIL tree nodes
5. Type checking flags non-numeric operands in arithmetic expressions
6. IO specification output metrics are validated against declared metrics
7. All unit tests pass following Go testing best practices
8. No regressions in existing CPL test suite
9. Parse/analysis separation maintained — no semantic validation in parser
