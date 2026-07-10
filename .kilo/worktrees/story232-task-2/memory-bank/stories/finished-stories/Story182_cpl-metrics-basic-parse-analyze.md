# Story182: CPL Metrics Basic Parse and Analyze

## Goal

Implement the Increment 1 parsing and analysis phases for CPL metric declarations as part of [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md). This story covers:

- Create [`CPLMetrics.g4`](src/chronocone_planning_language/src/main/antlr/CPLMetrics.g4) grammar with `metricsSection`, `metricDeclaration`, `inputs` block (metric/dimension names only, no `as`/`where`/`backward`), and `calculate` clause accepting literal `0`.
- Implement `MetricsListener` that pushes [`ParserMetricDefinition`](src/chronocone_planning_language/src/main/go/parsing/parser_metric_definition.go) onto the symbol table on `ExitMetricDeclaration`, pushes input ref definitions on `ExitInputRef`, and pushes literal `0` operand on `ExitLiteral`.
- Implement [`AnalyzeMetricScope`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go) function registered with `WalkSymbolTable` — assembles [`AnalysisMetricDefinition`](src/chronocone_planning_language/src/main/go/analysis/analysis_metric_definition.go) entries from parser definitions and generates one CIL tree per metric with a single `CILLiteral(0)` node. Generate one single-node CIL tree per connector.
- Validate input metric/dimension references against declared entities.

## References

- [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md)
- [Metrics Syntax Requirements](../specifications/requirements/cpl/metrics-syntax.md)
- [Metrics Design](../specifications/design/cpl/metrics-design.md)
- [ERO005: Metrics Parsing Phase — Increment 1](../specifications/design/cpl/go-parser-target/eros/ERO005_metrics-parsing-phase.md)
- [ERO005: Metrics Analysis Phase — Increment 1](../specifications/design/cpl/go-analysis/eros/ERO005_metrics-analysis-phase.md)
- [ERO005: Metrics Incremental Implementation](../specifications/design/cpl/eros/ERO005_metrics-incremental-implementation.md)
- [CPL Architecture — Intermediate Language](../specifications/architecture/cpl/analysis/intermediate_language.md)
- [Symbol Table-Driven Parsing Pattern](../specifications/patterns/symbol-table-driven-parsing.md)
- [Postfix Expression Parsing Pattern](../specifications/patterns/postfix-expression-parsing.md)
- [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Dependencies

- [Epic001_cpl-concepts-implementation.md](../epics/Epic001_cpl-concepts-implementation.md) — concepts implementation must be complete; metrics reference concepts as dimensions
- [Epic002_cpl-types-implementation.md](../epics/Epic002_cpl-types-implementation.md) — types implementation must be complete
- [Epic003_units-and-measures-implementation.md](../epics/Epic003_units-and-measures-implementation.md) — units and measures implementation must be complete
- [Epic004_cpl-connectors-implementation.md](../epics/Epic004_cpl-connectors-implementation.md) — connectors implementation must be complete; connectors provide leaf data that metrics consume

## Dependent Stories

- [Story183_cpl-metrics-basic-synthesize.md](Story183_cpl-metrics-basic-synthesize.md) — Increment 1 Synthesis (Not Started)

## Tasks

1. Create `CPLMetrics.g4` grammar with basic metric rules — Completed
   a. Define `metricsSection` rule referencing an `INDENT`-delimited block containing `metricDeclaration` entries — Completed
   b. Define `metricDeclaration` rule: `indent METRICS_KEYWORD identifier inputsBlock calculateClause` — Completed
   c. Define `inputsBlock` rule: `indent INPUT_KEYWORD INDENT inputDeclaration+ DEDENT` — Completed
   d. Define `inputDeclaration` rule accepting `metric` or `dimension` keyword followed by an identifier (no `as`/`where`/`backward` yet) — Completed
   e. Define `calculateClause` rule: `indent CALCULATE_KEYWORD literal` where `literal` is `0` for this increment — Completed
   f. Update root grammar (`ChronoconePlanningLanguage.g4`) to include `metricsSection` and needed lexer tokens (`METRICS_KEYWORD`, `INPUT_KEYWORD`, `CALCULATE_KEYWORD`, `metric`, `dimension`) — Completed
   g. Regenerate ANTLR parser/lexer Go files — Completed

2. Implement parser data structures and MetricsListener — Completed
   a. Define [`ParserMetricDefinition`](src/chronocone_planning_language/src/main/go/parsing/parser_metric_definition.go) with fields: `Name`, `Inputs []ParserInputDefinition`, `CalculatePostfix []interface{}`, `ShortForm bool` — Completed
   b. Define [`ParserInputDefinition`](src/chronocone_planning_language/src/main/go/parsing/parser_input_definition.go) with fields: `Kind` (metric/dimension), `Name string` — Completed
   c. Create `cpl_metrics_listener.go` with listener methods following the ANTLR Listener Organization Convention (one listener file per grammar file) — Completed
   d. Implement `EnterMetricsSection`/`ExitMetricsSection` — push/pop a `MetricsScope` on the symbol table — Completed
   e. Implement `EnterMetricDeclaration`/`ExitMetricDeclaration` — push/pop a `MetricScope`; on exit, assemble `ParserMetricDefinition` and store in symbol table via `DefineSymbol` — Completed
   f. Implement `ExitInputDeclaration` — create `ParserInputDefinition` and attach to the current `ParserMetricDefinition.Inputs` — Completed
   g. Implement `ExitLiteral` — push literal operand (value `0`) onto the current scope's postfix slice — Completed
   h. Verify parser tests pass: parse basic metric declarations and assert correct `ParserMetricDefinition` values on the symbol table — Completed

3. Implement analysis phase for basic metrics — Completed
   a. Define [`AnalysisMetricDefinition`](src/chronocone_planning_language/src/main/go/analysis/analysis_metric_definition.go) with fields: `Name`, `Inputs []AnalysisInputDefinition`, `CILRoot *CILMetricNode` — Completed
   b. Define [`AnalysisInputDefinition`](src/chronocone_planning_language/src/main/go/analysis/analysis_input_definition.go) with fields: `Kind`, `Name string` — Completed
   c. Define [`CILMetricNode`](src/chronocone_planning_language/src/main/go/cil/cil_metric_node.go) with fields: `CILNode` (a single CIL root), `InputRefs []CILInputRef` — Completed
   d. Define [`CILLiteral`](src/chronocone_planning_language/src/main/go/cil/cil_literal.go) node with `Value float64` field — Completed
   e. Create [`AnalyzeMetricScope`](src/chronocone_planning_language/src/main/go/analysis/metrics_analysis.go) function: reads `MetricsScope` from symbol table, iterates over `MetricScope` children, assembles `AnalysisMetricDefinition` from each `ParserMetricDefinition` — Completed
   f. Build CIL tree: for each metric's `CalculatePostfix`, create a single `CILLiteral(0)` node — Completed
   g. Register `AnalyzeMetricScope` with the `WalkSymbolTable` callback — Completed
   h. Validate input references: verify each input metric/dimension name exists in declared entities (connectors for metrics, concepts for dimensions) — Completed
   i. Generate one single-node CIL tree per connector declaration (for synthesis to use) — Completed
   j. Verify analyzer tests pass: assemble `AnalysisMetricDefinition` with correct structure and validations — Completed

4. Create test examples and integration verification — Completed
   a. Create example CPL file with basic metric declarations (one or more metrics, each with inputs block and calculate clause returning literal `0`) — Completed
   b. Create example IO specification file if needed for the increment — Completed
   c. Verify end-to-end parse-analyze pipeline produces expected `AnalysisMetricDefinition` entries with correct CIL trees — Completed
   d. Run the full CPL test suite to confirm no regressions — Completed

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. Parse/analysis separation must be maintained: the parser captures syntax only; the analyzer performs semantic validation
3. Follow the [Symbol Table-Driven Parsing Pattern](../specifications/patterns/symbol-table-driven-parsing.md) — push/pop scope tables during parsing, define symbols in current scope, attach parse-context data to `SymbolDefinition.Data`
4. Follow the [Postfix Expression Parsing Pattern](../specifications/patterns/postfix-expression-parsing.md) — parser pushes operands/operators in postfix order
5. Follow the [ANTLR Listener Organization Convention](../../lessons-learned.md#antlr-listener-organization-convention) — one listener file per grammar file
6. Data structures evolve incrementally — fields are added only in this increment, not pre-declared. Go zero values provide sensible defaults for unused fields
7. Types are initially represented as `float64`; detailed type selection is deferred
8. The `lookup` input kind keyword may be recognized by the lexer but lookup resolution is not implemented in this increment
9. Generated Go tests must follow the [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)
10. Go functions that can fail must return `(T, error)`

## Intent

- **Task 1 Intent**: Establish the grammar foundation for metrics — define the syntactic structure that all subsequent increments will extend
- **Task 2 Intent**: Build the parsing phase that captures raw syntax into `ParserMetricDefinition` data structures on the symbol table, with no semantic validation
- **Task 3 Intent**: Build the analysis phase that reads parser data, performs semantic validation (name resolution, duplicate detection), and constructs the CIL intermediate representation with `CILLiteral(0)` nodes
- **Task 4 Intent**: Verify the complete parse-analyze pipeline works end-to-end with test examples and produces correct `AnalysisMetricDefinition` entries

## Acceptance Criteria

1. `CPLMetrics.g4` grammar correctly parses basic metric declarations (metric name, inputs block with metric/dimension names, calculate clause with literal `0`)
2. `MetricsListener` pushes correct `ParserMetricDefinition` onto the symbol table with all inputs captured
3. `AnalyzeMetricScope` assembles `AnalysisMetricDefinition` with correct structure, including a single-node `CILLiteral(0)` CIL tree per metric
4. Input references are validated against declared entities (connectors for metrics, concepts for dimensions)
5. One single-node CIL tree is generated per connector declaration
6. All unit tests pass following Go testing best practices
7. No regressions in existing CPL test suite
8. Parse/analysis separation is maintained — no semantic validation occurs in the parser
