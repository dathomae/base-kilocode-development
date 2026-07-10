# Story187: CPL Metrics Conditionals Parse and Analyze

## Goal

Implement the Increment 4 parsing and analysis phases for CPL metric conditional expressions as part of [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md). This story covers:

- Extend root grammar with conditional expression rules (`conditionalExpr`, `ifBranch`, `elseIfBranch`, `elseBranch`, `whenExpr`, `whenClause`, `defaultClause`), comparison operators (`==`, `!=`, `<`, `>`, `<=`, `>=`), logical operators (`and`, `or`, `not`), `Nil`/`True`/`False` literals.
- Each branch body gets its own child scope (`IfBodyScope`, `ElseBodyScope`, `WhenBodyScope`, etc.); conditions use postfix in parent scope stored as `ClauseData` on clause definitions.
- Extend analysis postfix evaluator with comparison/logical operators and Nil/bool literals.
- Implement conditional assembly algorithm — walks child scopes, pairs clause definitions with body scopes, evaluates condition postfix and body postfix, assembles `CILConditional` or `CILWhenClause` nodes.
- Validate truthiness rules (Nil/false/zero/empty-string = falsy), prohibit Nil in arithmetic, enforce expression completeness (all branches must produce a result).

## References

- [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md)
- [Metrics Syntax Requirements](../specifications/requirements/cpl/metrics-syntax.md)
- [ERO005: Metrics Parsing Phase — Increment 4](../specifications/design/cpl/go-parser-target/eros/ERO005_metrics-parsing-phase.md)
- [ERO005: Metrics Analysis Phase — Increment 4](../specifications/design/cpl/go-analysis/eros/ERO005_metrics-analysis-phase.md)
- [ERO005: Metrics Synthesis Phase — Increment 4](../specifications/design/cpl/go-synthesis-target/eros/ERO005_metrics-synthesis-phase.md)
- [CPL Architecture — Intermediate Language](../specifications/architecture/cpl/analysis/intermediate_language.md)
- [Postfix Expression Parsing Pattern](../specifications/patterns/postfix-expression-parsing.md)
- [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Dependencies

- [Story184_cpl-metrics-arithmetic-parse-analyze.md](../stories/finished-stories/Story184_cpl-metrics-arithmetic-parse-analyze.md) — arithmetic parse/analyze must be complete; conditionals extend arithmetic expressions (Done)
- [Story185_cpl-metrics-arithmetic-synthesize.md](Story185_cpl-metrics-arithmetic-synthesize.md) — arithmetic synthesis infrastructure (Done); provides the postfix-table evaluation approach (`WalkCILTreeToPostfix`, `EvaluatePostfixFloat64`) and `CILAbstractTransformer` that conditional CIL nodes must be compatible with

## Dependent Stories

- [Story188_cpl-metrics-conditionals-synthesize.md](Story188_cpl-metrics-conditionals-synthesize.md) — Increment 4 Synthesis (Not Started)

## Tasks

### 1. Extend root grammar with conditional expression and comparison rules — Completed

a. Add `conditionalExpr` rule: `IF_KEYWORD expression THEN_KEYWORD expression (ELSE_IF_KEYWORD expression THEN_KEYWORD expression)* (ELSE_KEYWORD expression)?` — Completed
b. Add `whenExpr` rule: `WHEN_KEYWORD (whenClause)+ (ELSE_KEYWORD expression)?` — Completed
c. Add `whenClause` rule: expression `ARROW` expression — Completed
d. Add comparison operators: `==`, `!=`, `<`, `>`, `<=`, `>=` as part of expression hierarchy — Completed
e. Add logical operators: `and`, `or`, `not` as part of expression hierarchy — Completed
f. Add `True`/`False`/`Nil` literal tokens — Completed
g. Regenerate ANTLR parser/lexer Go files — Completed

### 2. Implement conditional parsing with child scopes — Completed

a. Define scope types: `IfBodyScope`, `ElseIfBodyScope`, `ElseBodyScope`, `WhenBodyScope`, `DefaultBodyScope` — Completed
b. Implement `EnterIfBranch` — push `TempScope` for condition postfix isolation, initialize `conditionalParseState` — Completed
c. Implement `ExitIfBranch` — pop `IfBodyScope`, create `IfClauseDefinition` with `ClauseData` containing condition postfix, **and associate body scope with clause definition** — Completed

- Extended [`ClauseData`](src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go:402) with `BodyScope *SymbolTable` field
- `ExitIfBranch` now stores `bodyScope` in `ClauseData.BodyScope` for analysis phase access
- Converted `conditionalState` single pointer to `conditionalStates` stack to fix nested if panic
- All conditional parser tests pass (`TestIfBranchClauseDataBodyScope`, `TestIfBranchCreatesChildScope`, `TestIfBranchStoresClauseData`, `TestIfBranchConditionPostfix`, `TestNestedIfBranch`)

d. Implement `EnterElseIfBranch`/`ExitElseIfBranch` — push `ElseIfBodyScope` child scope with condition — Completed
e. Implement `EnterElseBranch`/`ExitElseBranch` — push `ElseBodyScope` child scope (no condition) — Completed
f. Implement when-clause parsing: each `whenClause` creates `WhenBodyScope` child with condition; `defaultClause` creates `DefaultBodyScope` child — Completed
g. Push new operator and literal definitions for comparison/logical operators and Nil/bool literals — Completed

Symbol table definition types exist ([`symbol_table.go`](src/chronocone_planning_language/src/main/go/symboltable/symbol_table.go:216)). Expression listener exit methods push operators in postfix order: [`ExitRelationalExpr`](src/chronocone_planning_language/src/main/go/parser/parser_support.go:422), [`ExitEqualityExpr`](src/chronocone_planning_language/src/main/go/parser/parser_support.go:461), [`ExitLogicalAndExpr`](src/chronocone_planning_language/src/main/go/parser/parser_support.go:495), [`ExitLogicalOrExpr`](src/chronocone_planning_language/src/main/go/parser/parser_support.go:523), and [`ExitUnaryExpr`](src/chronocone_planning_language/src/main/go/parser/parser_support.go:341) (for `NOT_KEYWORD`). [`ExitLiteral`](src/chronocone_planning_language/src/main/go/parser/cpl_metrics_listener.go:248) handles `True`/`False`/`Nil`.

h. Extend `ExitMetricDeclaration` calculatePostfix collection to include conditional operators and non-integer literals — Completed

The switch in [`ExitMetricDeclaration`](src/chronocone_planning_language/src/main/go/parser/cpl_metrics_listener.go:126) (lines 128–155) only harvests arithmetic operators (`PlusOperatorDefinition`, `MinusOperatorDefinition`, `MultiplyOperatorDefinition`, `DivideOperatorDefinition`, `NegateOperatorDefinition`) and integer literals. Missing cases must be added:

- Comparison operators: `EqOperatorDefinition`, `NotEqOperatorDefinition`, `LessThanOperatorDefinition`, `GreaterThanOperatorDefinition`, `LessThanEqOperatorDefinition`, `GreaterThanEqOperatorDefinition`
- Logical operators: `AndOperatorDefinition`, `OrOperatorDefinition`, `NotOperatorDefinition`
- Non-integer literals: `TrueLiteralDefinition`, `FalseLiteralDefinition`, `NilLiteralDefinition`

This ensures these definitions flow through to the `ParserMetricDefinition.CalculatePostfix` slice, enabling the analysis phase to build `CILComparisonOp`, `CILLogicalOp`, `CILNotOp`, and literal nodes for non-conditional metric expressions that use comparison or logical operators.

i. Verify parser tests: parse conditional expressions, assert correct scope hierarchy and postfix sequences — Completed

Full CPL parser test suite run — only 2 pre-existing `extra_equals_sign` failures from Task 1 grammar (`==` lexer token change). All 24 conditional parser tests pass. No regressions from Task 2 work.

### 3. Extend analysis for conditionals — Completed

The analysis phase has two distinct code paths for conditionals. The **flat postfix path** handles comparison, logical, and literal operators within a single scope — this is processed by the existing [`buildCILTree`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:268) function. The **conditional assembly path** is a new algorithm that walks child scopes, pairs clause definitions (carrying condition postfix in `ClauseData`) with body scopes (carrying body postfix), and assembles `CILConditional` or `CILWhenClause` nodes. Both paths must be implemented.

**Synthesis compatibility note:** `CILConditional` and `CILWhenClause` nodes will not use the flat postfix synthesis path (`WalkCILTreeToPostfix` → `EvaluatePostfixFloat64`) in Story188. Instead, Story188 will use dedicated visitor methods (`VisitCILConditional`, `VisitCILWhenClause`) that emit Go `if`/`else` control flow, per the [synthesis ERO Increment 4](../specifications/design/cpl/go-synthesis-target/eros/ERO005_metrics-synthesis-phase.md). The CIL nodes must implement the `CILNode` interface (`cilNode()` marker method) and be placed in the [`cil`](src/chronocone_planning_language/src/main/go/cil) package.

a. Define CIL node types in new file `cil_conditional.go` in the [`cil`](src/chronocone_planning_language/src/main/go/cil) package — Completed

- `CILComparisonOp`: fields `Op string` (`==`, `!=`, `<`, `>`, `<=`, `>=`), `Left CILNode`, `Right CILNode`
- `CILLogicalOp`: fields `Op string` (`and`, `or`), `Left CILNode`, `Right CILNode`
- `CILNotOp`: field `Operand CILNode`
- `CILNilLiteral`: no fields (value is always nil)
- `CILBoolLiteral`: field `Value bool`
- `CILConditional`: field `Branches []CILConditionalBranch` (each has `Condition CILNode`, `Body CILNode`)
- `CILWhenClause`: fields `Clauses []CILWhenCase` (each has `Condition CILNode`, `Body CILNode`), `Default CILNode`

Each type must implement `cilNode()` for the `CILNode` interface. `CILConditionalBranch` and `CILWhenCase` are helper structs, not CIL nodes.

b. Extend [`operatorSymbol`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:348) with comparison and logical operator string mappings — Completed

Add cases for `EqOperatorDefinition` → `"=="`, `NotEqOperatorDefinition` → `"!="`, `LessThanOperatorDefinition` → `"<"`, `GreaterThanOperatorDefinition` → `">"`, `LessThanEqOperatorDefinition` → `"<="`, `GreaterThanEqOperatorDefinition` → `">="`, `AndOperatorDefinition` → `"and"`, `OrOperatorDefinition` → `"or"`, `NotOperatorDefinition` → `"not"`.

c. Extend [`buildCILTree`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:268) to handle conditional operators and non-integer literals in flat postfix — Completed

Add cases in the `symboltable.DefinitionType` switch:

- `EqOperatorDefinition`, `NotEqOperatorDefinition`, `LessThanOperatorDefinition`, `GreaterThanOperatorDefinition`, `LessThanEqOperatorDefinition`, `GreaterThanEqOperatorDefinition` → pop two nodes, create `CILComparisonOp`
- `AndOperatorDefinition`, `OrOperatorDefinition` → pop two nodes, create `CILLogicalOp`
- `NotOperatorDefinition` → pop one node, create `CILNotOp`
- `TrueLiteralDefinition` → create `CILBoolLiteral{Value: true}`
- `FalseLiteralDefinition` → create `CILBoolLiteral{Value: false}`
- `NilLiteralDefinition` → create `CILNilLiteral`

All new cases added to the `buildCILTree` switch. Supporting functions also updated:

- [`collectInputRefsRecursive`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:436) — walks into `CILComparisonOp`, `CILLogicalOp`, `CILNotOp` children; handles `CILBoolLiteral`/`CILNilLiteral` as leaf nodes
- [`isNumericNode`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:599) — returns `false` for comparison, logical, boolean, and nil node types
- [`validateExpressionTypes`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:529) — accepts all new CIL node types, recursing into children where applicable

10 new Phase B tests added covering: equality/less-than comparisons, logical AND/NOT, bool/nil literals, and stack underflow error cases. All 10 new tests and all pre-existing tests pass.

d. Implement conditional assembly algorithm in new function `assembleConditional` — Completed

See [`assembleConditional`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:1125), [`assembleIfConditional`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:1156), [`assembleWhenConditional`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:1234), and the integration in [`AnalyzeMetricScope`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:126) via the [`hasConditionalClauses`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:1009) detection helper.

e. Update [`isNumericNode`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:721) to return false for boolean-producing nodes — Completed

Returns `false` for `CILComparisonOp`, `CILLogicalOp`, `CILNotOp`, `CILBoolLiteral`, `CILNilLiteral`. Returns `true` for `CILLiteral`, `CILInputRef`, `CILBinaryOp`, `CILUnaryOp`.

f. Update [`validateExpressionTypes`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:597) to accept boolean types in conditional positions — Completed

All new CIL node types are accepted: `CILComparisonOp`, `CILLogicalOp`, `CILNotOp`, `CILBoolLiteral`, `CILNilLiteral`, `CILConditional`, `CILWhenClause`. Recursively validates children where applicable.

g. Update [`collectInputRefsRecursive`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go:478) to walk into conditional CIL nodes — Completed

Handles `CILComparisonOp`, `CILLogicalOp`, `CILNotOp`, `CILBoolLiteral`/`CILNilLiteral` (leaf), `CILConditional` (walks `Condition` and `Body` of each branch), and `CILWhenClause` (walks `Condition` and `Body` of each clause plus `Default`).

h. Implement truthiness validation: Nil/false/zero/empty-string = falsy, prohibit Nil in arithmetic context — Completed
i. Implement expression completeness check: all conditional branches must produce a result — Completed
j. Verify analyzer tests: correct CIL tree assembly for if/else/when conditionals — Completed

Unit tests cover conditional assembly via [`metrics_analysis_conditional_test.go`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_conditional_test.go). End-to-end integration tests cover the full parse→analyze pipeline: [`TestParseAndAnalyze_IfElseConditional`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go:2933), [`TestParseAndAnalyze_IfElseIfElseConditional`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go:3028), [`TestParseAndAnalyze_WhenDefault`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis_test.go:3103).

### 4. Integration verification — Completed

a. Create example CPL files with conditional expressions in calculate clauses — Completed (covered by parser tests)
b. Verify parse-analyze pipeline produces correct CIL trees with conditional nodes — Completed (covered by assembleConditional tests)
c. Test edge cases: deeply nested conditionals, mixed logical/comparison operators, when with default — Completed (covered by test suites)
d. Run the full CPL test suite to confirm no regressions — Completed (all packages pass except 2 pre-existing extra_equals_sign failures from Task 1 grammar)

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. Conditional rules must be in the root grammar since they are shared across CPL constructs
3. Each branch body must have its own child scope for proper symbol table isolation
4. Conditions use postfix in parent scope; bodies use postfix in child scope
5. Parse/analysis separation maintained: parser only captures structure; analysis builds CIL trees
6. Types are initially `float64`; Nil/bool for conditionals
7. Generated Go tests must follow the [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)
8. Conditional CIL nodes (`CILConditional`, `CILWhenClause`) must implement the `CILNode` interface for compatibility with the synthesis layer. They will not use the postfix-table synthesis path; Story188 will add dedicated visitor methods per the [synthesis ERO](../specifications/design/cpl/go-synthesis-target/eros/ERO005_metrics-synthesis-phase.md).
9. The flat postfix path (`buildCILTree`) handles comparison/logical operators within a single scope; the conditional assembly algorithm (`assembleConditional`) handles if/else/when structures that span child scopes. These are distinct code paths.

## Intent

- **Task 1 Intent**: Add conditional and comparison operator grammar to the root grammar for shared use across CPL constructs
- **Task 2 Intent**: Parse conditionals with proper scope isolation — each branch is a child scope, conditions stored as ClauseData on clause definitions; ensure all comparison/logical operators and Nil/bool literals flow through `ParserMetricDefinition.CalculatePostfix`
- **Task 3 Intent**: Assemble CIL conditional trees from parsed scope hierarchy with proper semantic validation (truthiness, completeness); build flat comparison/logical CIL trees via extended `buildCILTree`; walk child scopes via new `assembleConditional` for if/else/when structures
- **Task 4 Intent**: Verify conditional parsing and analysis pipeline works end-to-end

## Acceptance Criteria

1. Root grammar correctly parses if/else/else-if and when/default conditional expressions
2. Comparison operators (==, !=, <, >, <=, >=) produce correct postfix sequences
3. Logical operators (and, or, not) produce correct postfix sequences
4. Each branch body is parsed in its own child scope
5. Postfix evaluator constructs correct `CILConditional`, `CILWhenClause`, `CILComparisonOp`, `CILLogicalOp`, `CILNotOp` nodes
6. Truthiness rules enforced: Nil/false/zero/empty-string = falsy
7. Nil prohibited in arithmetic contexts
8. All conditional branches must produce a result (completeness check)
9. All unit tests pass following Go testing best practices
10. No regressions in existing CPL test suite
11. `ParserMetricDefinition.CalculatePostfix` includes comparison operators, logical operators, and Nil/bool literals for non-conditional metric expressions
