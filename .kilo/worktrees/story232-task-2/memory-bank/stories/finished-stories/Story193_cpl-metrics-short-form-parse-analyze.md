# Story193: CPL Metrics Short Form Parse and Analyze

## Goal

Implement the Increment 7 parsing and analysis phases for CPL metric short form declarations as part of [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md). **This is a parse-analyze only increment — no synthesis changes needed** since short form produces identical CIL trees to equivalent long form. This story covers:

- Extend [`CPLMetrics.g4`](src/chronocone_planning_language/src/main/antlr/CPLMetrics.g4) with `shortFormDeclaration` rule (`<name> = <calculation>`).
- Parser converts short form to equivalent long form [`ParserMetricDefinition`](src/chronocone_planning_language/src/main/go/parsing/parser_metric_definition.go) internally (sets `ShortForm: true`).
- Analyzer scans CIL tree for `CILInputRef` nodes and resolves each to metric/lookup/dimension using precedence: metrics first, then lookups (future), then dimensions.
- Detect name collisions (a name matching multiple entity types is an error).
- Short form produces identical CIL trees to equivalent long form.

## References

- [Epic005: CPL Metrics Implementation](../epics/Epic005_cpl-metrics-implementation.md)
- [Metrics Syntax Requirements](../specifications/requirements/cpl/metrics-syntax.md)
- [ERO005: Metrics Parsing Phase — Increment 7](../specifications/design/cpl/go-parser-target/eros/ERO005_metrics-parsing-phase.md)
- [ERO005: Metrics Analysis Phase — Increment 7](../specifications/design/cpl/go-analysis/eros/ERO005_metrics-analysis-phase.md)
- [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Dependencies

- [Story184_cpl-metrics-arithmetic-parse-analyze.md](Story184_cpl-metrics-arithmetic-parse-analyze.md) — arithmetic parse/analyze with expression parsing and IO spec must be complete (Done)

## Dependent Stories

- [Story194_cpl-metrics-dag-io-analyze.md](../Story194_cpl-metrics-dag-io-analyze.md) — Increment 8 Analysis (Not Started)

## Tasks

1. Extend [`CPLMetrics.g4`](src/chronocone_planning_language/src/main/antlr/CPLMetrics.g4) with short form declaration rule — Completed
   a. Add `shortFormDeclaration` rule: `indent identifier EQUALS expression` — Completed
   b. Integrate `shortFormDeclaration` into `metricsSection` as an alternative: `METRICS (metricDeclaration | shortFormDeclaration)+` — Completed
   c. Verify `EQUALS` token already exists in lexer — the [`EQUALS: '=';` token](src/chronocone_planning_language/src/main/antlr/ChronoconePlanningLanguageLexer.g4:48) is already defined and used throughout the grammar (concepts, types, units, measures, connectors). No new token needed. — Completed
   d. Regenerate ANTLR parser/lexer Go files — Completed

2. Implement parser conversion to long form — Completed
   a. Implement `ExitShortFormDeclaration` — create `ParserMetricDefinition` with `ShortForm: true` from the short form syntax. The [`ShortForm bool` field](src/chronocone_planning_language/src/main/go/parsing/parser_metric_definition.go:27) already exists in `ParserMetricDefinition` (forward-declared); it just needs to be set to `true`. — Completed
   b. Handle scope edge case: when `ExitShortFormDeclaration` fires, the current scope may be the `MetricsScope` directly (not a `MetricScope` child). In that case, push a new `MetricScope` child onto the symbol table before assembling the definition. **Implementation note:** `EnterShortFormDeclaration` pushes a `MetricScope` before expression parsing, so expressions accumulate in the correct scope. `ExitShortFormDeclaration` verifies the MetricScope, collects the postfix, and pops. — Completed
   c. Extract metric name from the left-hand identifier — Completed
   d. Set the calculate expression postfix from the right-hand expression — Completed
   e. Derive implicit inputs: scan expression postfix for identifiers that are not declared metrics — Completed
   f. Push the converted `ParserMetricDefinition` onto the symbol table (identical structure to long form, differing only in `ShortForm` flag) — Completed
   g. Verify parser tests: short form declarations produce correct `ParserMetricDefinition` with `ShortForm: true` — Completed

3. Implement analysis phase for short form name resolution — Completed
   a. First, build the CIL tree from the postfix definitions in the MetricScope using the same `buildCILTree` mechanism as long form — this produces a complete CIL tree with `CILInputRef` nodes. — Completed
   b. Implement CIL tree scan: traverse the resulting CIL tree for `CILInputRef` nodes (using existing `collectInputRefs` infrastructure) — Completed
   c. Resolve each `CILInputRef` using precedence: metrics first, then lookups (future), then dimensions (concepts) — Completed
   d. Detect name collisions: a name matching multiple entity types is an error — Completed
   e. Short form produces CIL trees identical to equivalent long form — verified via `TestAnalyzeMetrics_ShortForm_EquivalentToLongForm` — Completed
   f. Verify analyzer tests: short form produces same CIL tree as equivalent long form; name collisions detected — Completed

4. Integration verification — Completed
   a. Create example CPL files with short form metric declarations — Completed (via inline test CPL)
   b. Create equivalent long form versions of the same metrics — Completed
   c. Verify parse-analyze pipeline produces identical CIL trees for short form and long form — Completed
   d. Test edge cases: short form referencing other short form metrics, name collisions — Completed
   e. Run the full CPL test suite to confirm no regressions — Completed (all 14 packages pass)

## Constraints

1. Follow the [Logical TDD Lifecycle](../../concepts.md#logical-tdd-lifecycle) (BTDD): Phase A (structural definitions), Phase B (tests), Phase C (implementation) — integrated within each task
2. No synthesis changes needed — short form produces identical CIL trees to long form
3. Implicit input derivation: scan expression for identifiers, classify each as metric/lookup/dimension
4. Name resolution precedence: metrics first, then lookups (future), then dimensions (concepts)
5. Name collisions across entity types must be detected and reported as errors
6. Parse/analysis separation maintained: parser converts to internal long form; analysis resolves names
7. Generated Go tests must follow the [Go Testing Best Practices Guide](../../resources/go/testing-guide.md)

## Intent

- **Task 1 Intent**: Add concise short form syntax to the grammar — `<name> = <calculation>` as an alternative to verbose long form. Uses existing `EQUALS` token and follows the `indent` pattern (without DEDENT) consistent with the rest of the CPL grammar per the [Handling Indentation in CPL Grammar lesson](memory-bank/lessons-learned.md:89-110).
- **Task 2 Intent**: Convert short form to internal long form during parsing — downstream code (analysis, synthesis) doesn't need to know about short form. Handles scope edge case where short form declarations may sit directly in a `MetricsScope`.
- **Task 3 Intent**: Resolve implicit input references with proper precedence and collision detection in two steps: first build the CIL tree from postfix (same as long form), then scan the tree for `CILInputRef` nodes and resolve each name.
- **Task 4 Intent**: Prove short form produces identical results to equivalent long form

## Acceptance Criteria

1. Short form declarations (`<name> = <calculation>`) correctly parsed as alternative to long form
2. Parser converts short form to equivalent `ParserMetricDefinition` with `ShortForm: true`
3. Analysis resolves `CILInputRef` nodes using correct precedence (metrics → lookups → dimensions)
4. Name collisions across entity types detected and reported as errors
5. Short form produces CIL trees identical to equivalent long form
6. All unit tests pass following Go testing best practices
7. No regressions in existing CPL test suite
8. Parse/analysis separation maintained
