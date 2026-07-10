# Story228: CPL Expression Infrastructure

## Goal

Implement shared expression infrastructure used by compound units, conversions, and (future) metrics as part of [Epic003: Units and Measures Implementation](../epics/Epic003_units-and-measures-implementation.md). This story covers the full parse-to-analyze-to-synthesize pipeline for arithmetic expressions with operator precedence. The expression grammar, parser, analyzer, and synthesizer built here are the foundation used by Story229 (compound units), Story230 (conversions), and future metric formula evaluation.

## References

- [Epic003: Units and Measures Implementation](../epics/Epic003_units-and-measures-implementation.md)
- [ERO003: Units and Measures Parsing Phase — Increment 3](../specifications/design/cpl/go-parser-target/eros/ERO003_units-measures-parsing-phase.md)
- [ERO003: Units and Measures Analysis Phase — Expression Tree Types, BuildExpressionTree](../specifications/design/cpl/go-analysis/eros/ERO003_units-measures-analysis-phase.md)
- [ERO003: Units and Measures Synthesis Phase (Revised)](../specifications/design/cpl/eros/ERO003_units-measures-synthesis-phase.md)
- [ERO003: Units and Measures Incremental Implementation](../specifications/design/cpl/eros/ERO003_units-measures-incremental-implementation.md)
- [Symbol Table-Driven Parsing Pattern](../specifications/patterns/symbol-table-driven-parsing.md)
- [Postfix Expression Parsing Pattern](../specifications/patterns/postfix-expression-parsing.md)
- [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Dependencies

- Existing basic units/measures parser and analyzer implementations (`cpl_units_listener.go`, `units_analysis.go`, `cpl_measures_listener.go`, `measures_analysis.go`) — implemented without formal story files between 2026-04-07 and 2026-06-03. These provide the symbol table scope types and base `AnalysisUnitDefinition` / `AnalysisMeasureDefinition` structures.

## Dependent Stories

- [Story229_cpl-compound-units.md](Story229_cpl-compound-units.md) — _Not Started_ — wires expression infrastructure into measure definitions
- [Story230_cpl-conversions.md](finished-stories/Story230_cpl-conversions.md) — _Done_ — uses expression grammar and BuildExpressionTree for conversion calculations

## Tasks

### Task Sizing

- **Task 1** (Grammar rules and ANTLR regeneration): Medium — touches one grammar file, requires lexer token additions, impacts all generated parser files. Not small because ANTLR regeneration affects the entire parser package.
- **Task 2** (Parser data structures and listener methods): Medium — introduces new parser types (`ParserUnitReference`, `ParserOperator`, `OperatorType` enum) and listener methods in `cpl_document_listener.go`. Coordinated with Task 1.
- **Task 3** (Expression interface and concrete types): `[Small]` — single file (`analysis/expression.go` or similar) defining interface and structs. Straightforward, no external coordination.
- **Task 4** (BuildExpressionTree implementation): Medium — core algorithm with error handling, unit reference resolution, and validation. Requires understanding of postfix evaluation and operator precedence.
- **Task 5** (CIL expression node types with unit tracking): `[Small]` — single file defining CIL expression interface and node structs. Additive, no external coordination.
- **Task 6** (Expression synthesis): Medium — implements `SynthesizeExpression` with Go code emission. Coordinates with Task 5 CIL types.
- **Task 7** (Integration verification): Medium — end-to-end test spanning parse, analyze, and synthesize. Cross-cutting by nature.

### Parallel Execution

- **Group A** (Tasks 3, 5): Task 3 defines analysis expression types; Task 5 defines CIL expression types. These are independent struct-definition tasks in different packages (`analysis/` vs `cil/`). No shared files or semantic dependencies.

### Execution Order

```
Task 1 → Task 2 → Task 4 → [Task 3 ‖ Task 5] → Task 6 → Task 7
```

- Task 1 must complete before Task 2 (grammar must exist before listener methods can reference new rules/tokens).
- Task 2 must complete before Task 4 (listener pushes postfix definitions that BuildExpressionTree consumes).
- Tasks 3 and 5 are struct-definition tasks that can proceed in parallel once the expression grammar is settled.
- Task 6 depends on Task 5 (needs CIL types) and should follow Task 4 (semantic understanding of expression trees).
- Task 7 depends on all preceding tasks.

### Reintegration

**Group A (Tasks 3, 5):**
- Merge order: any order (disjoint packages — `analysis/` vs `cil/`, no shared infrastructure)
- Integration tests: `go test ./src/main/go/analysis/... ./src/main/go/cil/... -count=1 -timeout 60s` — verify both packages compile together
- Watch for conflicts in: none expected

### Task 1: Add expression grammar rules to ChronoconePlanningLanguage.g4 and regenerate ANTLR files — Completed
   a. Add `expression`, `term`, `factor` grammar rules with operator precedence to `ChronoconePlanningLanguage.g4`:
      ```
      expression: expression (PLUS|MINUS) term | term;
      term: term (MULT|DIV) factor | factor;
      factor: factor POW INTEGER | identifier | NUMBER | LPAREN expression RPAREN;
      ```
   b. Add explicit operator callback rules: `multOperator: MULT; divOperator: DIV; powOperator: POW; plusOperator: PLUS; minusOperator: MINUS;`
   c. Add lexer tokens: `MULT`, `DIV`, `POW`, `PLUS`, `MINUS`, `NUMBER`
   d. Regenerate ANTLR Go parser/lexer files from updated grammar — Completed
   e. Verify generated files compile: `go build ./src/main/go/parsing/...` — Completed

### Task 2: Define parser types and implement document-level expression listener methods — Completed
   a. Add `DefinitionType` values to symbol table: `UnitReferenceDefinition`, `MultiplyOperatorDefinition`, `DivideOperatorDefinition`, `PowerOperatorDefinition`, `PlusOperatorDefinition`, `MinusOperatorDefinition`, `NumberDefinition` — Not Started
   b. Define `ParserUnitReference` struct with `Name string` field — Not Started
   c. Define `OperatorType` enum and `ParserOperator` struct with `Operator OperatorType` field — Not Started
   d. Implement `ExitIdentifier` in `cpl_document_listener.go` — pushes `ParserUnitReference` to scope definitions when within an expression context — Not Started
   e. Implement `ExitMultOperator`, `ExitDivOperator`, `ExitPowOperator`, `ExitPlusOperator`, `ExitMinusOperator` — each pushes `ParserOperator` to scope definitions — Not Started
   f. Implement `ExitNumber` — pushes numeric literal to scope definitions — Not Started
   g. Write parser unit tests verifying postfix ordering for expressions: `a*b`, `a/b`, `a^2`, `a+b`, `(a+b)*c` — Not Started
   h. Verify parser tests pass — Not Started

### Task 3: Define Expression interface and concrete analysis node types — Completed
   a. Define `Expression` interface with `isExpression()` marker method — Not Started
   b. Define concrete types: `UnitReference` (Name, UnitGroup *AnalysisUnitDefinition), `MultiplicationNode` (Left, Right Expression), `DivisionNode` (Left, Right Expression), `PowerNode` (Base Expression, Exponent int), `AdditionNode` (Left, Right Expression), `SubtractionNode` (Left, Right Expression), `NumberNode` (Value float64) — Not Started
   c. Write unit tests for expression node construction — Not Started

### Task 4: Implement BuildExpressionTree with postfix evaluation — Completed
   a. Write tests for `BuildExpressionTree` covering: simple unit reference, multiplication, division, power, addition, subtraction, parentheses, nested operators, malformed expressions, undefined units, insufficient operands — Not Started
   b. Implement `BuildExpressionTree(defs []SymbolDefinition, units map[string]AnalysisUnitDefinition) (Expression, error)` — stack-based postfix evaluator that constructs the expression tree — Not Started
   c. Implement unit reference resolution during tree building (lookup in units map) — Not Started
   d. Implement validation: malformed expressions, undefined units, insufficient operands — Not Started
   e. Verify analyzer tests pass — Not Started

### Task 5: Define CIL expression node types with unit tracking — Completed
   a. Define `CILExpression` interface with unit-tracking marker — Not Started
   b. Define `CILUnitRef` (UnitGroup string), `CILBinaryOp` (OperatorType, Left, Right CILExpression, DerivedUnit string), `CILPower` (Base CILExpression, Exponent int, DerivedUnit string), `CILNumber` (Value float64, AnnotatedUnit string) — Not Started
   c. Ensure each node carries unit type information for use in Story231 synthesis-time checking — Not Started
   d. Write unit tests for CIL node construction — Not Started

### Task 6: Implement expression synthesis — Completed
   a. Write tests for `SynthesizeExpression` covering all expression forms and verifying emitted Go code validity — Not Started
   b. Implement `SynthesizeExpression(expr CILExpression) string` — emits Go expression string with proper operator precedence and parentheses — Not Started
   c. Implement visitor/walker pattern or type switch for each CIL expression node type — Not Started
   d. Verify emitted Go expressions compile correctly by generating them in test cases and parsing with `go/parser` — Not Started

### Task 7: Integration verification and regression tests — Completed
   a. Create CPL snippet with a metric or test harness that uses an expression (e.g., `Speed = Number(unit=Length/Time)` for Story229 preview) — Not Started
   b. Run parse → analyze → synthesize pipeline and verify output contains correct Go expression code — Not Started
   c. Verify existing basic units and measures tests still pass (no regression on Increments 1–2) — Not Started
   d. Run full CPL test suite: `go test ./... -count=1` from `src/chronocone_planning_language/src/main/go/` — Not Started

## Test-First Development

This project follows the Logical TDD Lifecycle. For each coding task above, tests are written before implementation and the test-and-implement cycle is integrated within each task rather than broken out as separate tasks. See `memory-bank/concepts.md` for more information on the Logical TDD Lifecycle.

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. Expression grammar rules go in `ChronoconePlanningLanguage.g4` (shared across units, measures, and future metrics)
3. Postfix expression parsing via symbol table stack — no AST during parsing
4. `BuildExpressionTree` must handle all operator precedence correctly
5. Expression synthesis must produce valid Go code (numeric literals, identifiers, operators, parentheses)
6. Go functions that can fail must return `(T, error)`
7. ANTLR listener organization convention: expression listener methods go in `cpl_document_listener.go` because expression rules are in the shared grammar
8. Generated Go tests must follow the [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Intent

- **Task 1 Intent**: Establish the grammar foundation for expressions — define syntactic rules with correct operator precedence that all dependent stories will use
- **Task 2 Intent**: Build the parsing phase that captures expression syntax as postfix-ordered symbol definitions on the symbol table
- **Task 3 Intent**: Define the analysis-side expression tree type system — the shape of expression trees that the analyzer produces
- **Task 4 Intent**: Implement the core postfix-to-tree algorithm that converts parser output into structured expression trees
- **Task 5 Intent**: Define the CIL-side expression node types with unit tracking — the intermediate representation that synthesis consumes
- **Task 6 Intent**: Build the synthesizer's expression code generator that turns CIL expression trees into valid Go code
- **Task 7 Intent**: Verify the complete parse→analyze→synthesize pipeline works end-to-end and produces no regressions

## Acceptance Criteria

1. `ChronoconePlanningLanguage.g4` contains expression grammar rules with correct operator precedence
2. ANTLR-generated Go files compile without errors
3. Listener methods push `ParserUnitReference`, `ParserOperator`, and numeric literals in correct postfix order
4. `BuildExpressionTree` produces correct expression trees for all operator combinations, including nested parentheses
5. `BuildExpressionTree` returns meaningful errors for malformed expressions, undefined units, and insufficient operands
6. CIL expression node types carry unit tracking information for every node kind
7. `SynthesizeExpression` emits valid, compilable Go expression strings for all expression forms
8. Integration test demonstrates parse → analyze → synthesize pipeline for an expression
9. All existing basic units and measures tests pass with zero regressions
10. Full CPL test suite passes

## Footer

### Open Questions

None

### Known Issues

None

### Additional Context

The expression infrastructure built in this story is intentionally generic. While Story229 and Story230 are the immediate consumers, the same grammar rules, listener methods, `BuildExpressionTree`, and CIL expression types will be reused for metric formula evaluation in future Epic005 increments. Design choices should keep this future reuse in mind — for example, `BuildExpressionTree` accepts a generic `units` lookup map rather than being hardcoded to `AnalysisUnitDefinition`.
