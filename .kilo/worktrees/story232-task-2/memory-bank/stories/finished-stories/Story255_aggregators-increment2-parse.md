# Story255: Aggregators Increment 2 (Parse)

## Goal

Implement the parsing phase for basic aggregator declarations as Increment 2 of Epic007. Aggregators are unnamed — the parser assigns deterministic synthetic keys (`__aggregator__[from]->[to]`) and stores expressions in postfix order within the aggregator's scope.

This story delivers: `CPLAggregators.g4` grammar with aggregator declaration rules; `cpl_aggregators_listener.go` that pushes symbol table entries under synthetic keys; `ParserAggregatorDefinition` struct for parsed data; synthetic key generation helper; lexer tokens for aggregator syntax; and root grammar integration to route expression definitions (concept references, function calls, operators) to aggregator scopes. All cross-reference resolution is deferred to the analysis phase (Story256).

## References

- [ERO007: Aggregators Incremental Implementation](../specifications/design/cpl/eros/ERO007_aggregators-incremental-implementation.md)
- [ERO007: Aggregators Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO007_aggregators-parsing-phase.md)
- [Postfix Expression Parsing Pattern](../patterns/postfix-expression-parsing.md)
- [Symbol Table Design](../specifications/design/cpl/go-target/symbol-table-design.md)

## Dependencies

- Story254 (aggregators-increment1-matrix-mult) — matrix multiplication must exist before synthesis can succeed, but parsing has no hard dependency on it
- Epic007 (aggregators) — concept names are stored as raw strings; resolution is deferred to analysis

## Dependent Stories

- Story256 (aggregators-increment3-analyze) — Not Started (analysis phase reads the symbol table structure created here)

## Tasks

### Task Sizing

Tasks 1 and 5 are routine, narrow-scope work that follows established grammar and example-directory patterns. Tasks 3 and 4 involve coordinated parser state machine integration across multiple files and are not annotated `[Small]`. Task 2 creates a struct/helper in a new file and adds enum entries to an existing file — it touches two source files in different packages; the `symbol_table.go` enum additions are trivially mechanical.

### Parallel Execution

- **Group A** (Tasks 2, 3): Task 2 creates `parser_aggregator_definition.go` (parser package) and modifies `symbol_table.go` (symboltable package); Task 3 creates `cpl_aggregators_listener.go` (parser package). Fully disjoint file sets across both tasks. Task 3's listener agent receives type specifications (struct fields, enum value names) to write syntactically correct code; the merged result compiles when both worktree branches are combined.

### Execution Order

```
Task 1 → [Task 2 ‖ Task 3] → Task 4 → Task 5
```

- Task 1 must complete first — grammar and lexer tokens must exist before listener or definition types can reference them.
- Tasks 2 and 3 run in parallel (Group A) — disjoint file sets in different packages.
- Task 4 depends on Tasks 2 and 3 — integration needs both definition types and listener to exist.
- Task 5 depends on all preceding tasks — examples need the full parser pipeline.

### Reintegration

**Group A (Tasks 2, 3):**
- Merge order: any order (fully disjoint files; Task 2 touches parser+symboltable packages, Task 3 touches parser package only — no shared files modified)
- Integration tests: `go test ./src/main/go/parser/... ./src/main/go/symboltable/... -count=1 -timeout 120s` — verify definition types and listener compile together and all tests pass
- Watch for conflicts in: `src/main/go/symboltable/symbol_table.go` (Task 2 adds `AggregatorsScope`/`AggregatorScope` to `ScopeType` and `AggregatorsScopeDefinition`/`AggregatorScopeDefinition` to `DefinitionType`; git merge may need enum value ordering resolved)

### Task 1: Create CPLAggregators.g4 grammar and add lexer tokens — Not Started [Small]

Create the ANTLR parser grammar for aggregator declarations and add the required lexer tokens to the CPL lexer.

**Intent**: Define the grammar rules that the parser uses to recognize aggregator declarations (`from`, `to`, expressions), and ensure the lexer recognizes the new keywords and symbols.

**Acceptance Criteria**:
- `src/main/antlr/CPLAggregators.g4` exists with rules for `aggregatorsSection`, `aggregatorDeclaration`, `toTarget`, `identifierList`, and `fnCall`
- `src/main/antlr/ChronoconePlanningLanguageLexer.g4` defines `AGGREGATORS`, `FROM`, `TO`, `LPAREN`, `RPAREN` tokens
- ANTLR code generation (`./gradlew generateGrammarSource`) succeeds with no errors
- No `SUM`, `COUNT`, or `SIZE` tokens — function names are ordinary identifiers parsed via `fnCall`

   a. Create `src/main/antlr/CPLAggregators.g4` as a parser grammar — Not Started
      - Name: `parser grammar CPLAggregators;`
      - Options: `tokenVocab = ChronoconePlanningLanguageLexer;`
      - Rules: `aggregatorsSection` (`AGGREGATORS aggregatorDeclaration+`), `aggregatorDeclaration` (`indent FROM LSQUARE identifierList RSQUARE TO LSQUARE toTarget RSQUARE EQUALS expression`), `toTarget` (`identifier | // empty`), `identifierList` (`identifier (COMMA identifier)*`), `fnCall` (`identifier LPAREN identifierList RPAREN`)
      - Follow the convention of existing parser grammars (`CPLMeasures.g4`, `CPLMetrics.g4`)
   b. Add lexer tokens to `src/main/antlr/ChronoconePlanningLanguageLexer.g4` — Not Started
      - Add `AGGREGATORS : 'aggregators';`, `FROM : 'from';`, `TO : 'to';`, `LPAREN : '(';`, `RPAREN : ')';`
      - Place near existing section keywords (`MEASURES`, `METRICS`, etc.)
   c. Write a grammar-level verification by running `./gradlew generateGrammarSource` and confirming zero errors — Not Started
   d. Verify the generated Go parser compiles: `go build ./src/main/go/parser/antlr/...` — Not Started

### Task 2: Define ParserAggregatorDefinition and synthetic key helper — Not Started [Small]

Create the Go struct that holds parsed aggregator data, implement the synthetic key generation function, and add the aggregator entries to the ScopeType and DefinitionType enums.

**Intent**: Define the data structures and identifiers the parser listener will populate, and establish the synthetic key convention that ensures aggregator names never collide with user-definable names.

**Acceptance Criteria**:
- `ParserAggregatorDefinition` struct exists with fields: `FromConcepts` (`[]string`), `ToConcept` (`string`), `ToEmpty` (`bool`), `SourceLocation` (`SourceLocation`)
- `aggregatorSyntheticKey(from []string, to string) string` generates keys in the format `__aggregator__[A,B]->[C]` (or `__aggregator__[A,B]->[]` for scalar reduction)
- `ScopeType` enum in `symbol_table.go` includes `AggregatorsScope` and `AggregatorScope`
- `DefinitionType` enum in `symbol_table.go` includes `AggregatorsScopeDefinition` and `AggregatorScopeDefinition`
- `String()` methods for both enums are updated with the new cases
- Unit tests verify struct initialization and synthetic key generation for common patterns

   a. Create `src/main/go/parser/parser_aggregator_definition.go` — Not Started
      - Define `type ParserAggregatorDefinition struct` with `FromConcepts []string`, `ToConcept string`, `ToEmpty bool`, `SourceLocation symboltable.SourceLocation`
      - Import `github.com/chronocone/chronocone_planning_language/src/main/go/symboltable`
   b. Implement `aggregatorSyntheticKey(from []string, to string) string` in the same file — Not Started
      - Use `fmt.Sprintf("__aggregator__[%s]->[%s]", strings.Join(from, ","), to)`; when `to == ""`, emit `[]` for the target
   c. Add `AggregatorsScope` and `AggregatorScope` to the `ScopeType` enum in `src/main/go/symboltable/symbol_table.go` — Not Started
      - Place after the existing connector scope entries, before `MetricsScope`
      - Add corresponding cases to the `ScopeType.String()` method
   d. Add `AggregatorsScopeDefinition` and `AggregatorScopeDefinition` to the `DefinitionType` enum in `src/main/go/symboltable/symbol_table.go` — Not Started
      - Place after `ConnectorScopeDefinition`, before `MetricsScopeDefinition`
      - Add corresponding cases to the `DefinitionType.String()` method
   e. Write unit test in `src/main/go/parser/parser_aggregator_definition_test.go` — Not Started
      - Verify `ParserAggregatorDefinition` struct can be initialized with all fields populated
      - Verify `aggregatorSyntheticKey` produces correct strings for: multi-concept from/single to, single-concept from/single to, scalar reduction (to=[]), empty from list
      - Verify the `__aggregator__` prefix is always present
   f. Verify the package compiles: `go build ./src/main/go/parser/... ./src/main/go/symboltable/...` — Not Started

### Task 3: Implement cpl_aggregators_listener.go — Not Started

Create the listener that responds to ANTLR parse-tree events for the aggregators grammar and populates the symbol table with `AggregatorScopeDefinition` entries containing `ParserAggregatorDefinition` data and postfix expression sequences.

**Intent**: Build the parse-tree-to-symbol-table bridge that extracts aggregator declarations, assigns synthetic keys, and captures expression operands/operators in postfix order within the aggregator scope.

**Acceptance Criteria**:
- `cpl_aggregators_listener.go` exists in the parser package and compiles
- `EnterAggregatorsSection` pushes an `AggregatorsScope` child scope onto the symbol stack
- `ExitAggregatorsSection` pops the scope from the symbol stack
- `EnterAggregatorDeclaration` pushes an `AggregatorScope` child scope (expression definitions accumulate here)
- `ExitAggregatorDeclaration` extracts `from` concept names, `to` target (identifier or empty), creates `ParserAggregatorDefinition`, computes synthetic key, calls `pushDefinition` with `AggregatorScopeDefinition`, stores the definition pointer in `scope.Data`, and pops the aggregator scope
- Unit tests verify correct symbol table tree structure, synthetic key correctness, definition data integrity, and postfix expression content
- No cross-reference validation (deferred to analysis)

   a. Create `src/main/go/parser/cpl_aggregators_listener.go` — Not Started
      - Package `parser`; embed `*antlr.BaseCPLAggregatorsListener` (the ANTLR-generated base listener for the CPLAggregators grammar)
      - Hold a `*CPLListener` reference (the main parser listener) for access to `symbolStack`, `AddError`, `pushDefinition`, `pushOperatorDef`, and logging utilities
   b. Implement `EnterAggregatorsSection` / `ExitAggregatorsSection` — Not Started
      - `EnterAggregatorsSection`: peek current scope from `symbolStack`, create child with `AddChild(symboltable.AggregatorsScope)`, push child onto stack
      - `ExitAggregatorsSection`: pop scope from `symbolStack` (error handling via AddError)
      - Follow pattern from `cpl_measures_listener.go` (`EnterMeasuresSection`/`ExitMeasuresSection`)
   c. Implement `EnterAggregatorDeclaration` — Not Started
      - Peek current scope, create child with `AddChild(symboltable.AggregatorScope)`, push child
      - This child scope will accumulate expression definitions (concept references, fnCalls, operators) in postfix order via the root grammar listener's `shouldPushExpressionDefs()` routing
   d. Implement `ExitAggregatorDeclaration` — Not Started
      - Extract from concepts: call `ctx.IdentifierList().AllIdentifier()` to collect raw identifier text into `[]string`
      - Extract to target: if `ctx.ToTarget().Identifier() != nil`, store its text in `ToConcept`; if the to-target is empty (the empty alternative), set `ToEmpty = true` and `ToConcept = ""`
      - Create `ParserAggregatorDefinition{FromConcepts, ToConcept, ToEmpty, sourceLocation}`
      - Compute synthetic key via `aggregatorSyntheticKey(fromConcepts, toConcept)`
      - Call `pushDefinition(syntheticKey, symboltable.AggregatorScopeDefinition, &def, sourceLocation)` — the definition pointer is stored in the aggregator scope's `Data` field via the symbol table's `DefineWithData` convention
      - Pop the aggregator scope from the symbol stack
      - The aggregator scope's `Definitions` slice already contains postfix expression entries accumulated during parsing (concept references, fnCalls, operators pushed by the root listener)
   e. Write listener unit tests in `src/main/go/parser/cpl_aggregators_listener_test.go` — Not Started
      - Test symbol table tree structure: parse a CPL snippet with `aggregators` section → verify root scope contains `AggregatorsScope` child → verify `AggregatorScope` child with `AggregatorScopeDefinition` in Data
      - Test synthetic key: verify key format `__aggregator__[Country,Product]->[Product]` for multi-concept from, `__aggregator__[Country]->[Country]` for single from/to, `__aggregator__[Country]->[]` for scalar reduction
      - Test definition data: verify `ParserAggregatorDefinition.FromConcepts`, `ToConcept`, `ToEmpty` match input
      - Test postfix definitions: verify the aggregator scope's `Definitions` slice contains `OperandDefinition` entries (concept identifiers, fnCall names) in declaration order
      - Test zero declarations: an `aggregators` section with no declarations produces an `AggregatorsScope` with zero children and no errors
      - Use the test helper patterns from `cpl_metrics_listener_test.go` as a model
   f. Verify pass: `go test ./src/main/go/parser/... -count=1 -timeout 60s` — Not Started

### Task 4: Integrate AggregatorsListener into CPLParser, root grammar, and expression routing — Not Started

Wire the new aggregator grammar and listener into the main CPL parser pipeline so that aggregator declarations are recognized during parsing and expression definitions are correctly routed to aggregator scopes.

**Intent**: Connect all the pieces created in Tasks 1–3 so the full parser pipeline processes aggregator sections end-to-end.

**Acceptance Criteria**:
- Root grammar (`ChronoconePlanningLanguage.g4`) imports `CPLAggregators` and includes `aggregatorsSection` in the `statement` rule
- ANTLR code generation (`./gradlew generateGrammarSource`) produces `CPLAggregators` listener base classes
- Main `CPLListener` (in `cpl_document_listener.go`) dispatches `aggregatorsSection` to the `AggregatorsListener` via the existing section-routing pattern
- `shouldPushExpressionDefs()` in `parser_support.go` returns true for `AggregatorScope` so concept references, fnCalls, and operators accumulate in postfix order
- Full pipeline: parse CPL with `aggregators` section → symbol table contains `AggregatorScopeDefinition` with correct `ParserAggregatorDefinition` and postfix `SymbolDefinition` entries
- Parser rejects malformed aggregator syntax (missing `from`, missing `to`, missing expression body) with parse errors

   a. Modify `src/main/antlr/ChronoconePlanningLanguage.g4` to import and include CPLAggregators — Not Started
      - Add `CPLAggregators` to the `import` statement: `import CPLConcepts, CPLMeasures, CPLTypes, CPLUnits, CPLConnectors, CPLMetrics, CPLAggregators;`
      - Add `aggregatorsSection` to the `statement` rule: `| aggregatorsSection`
      - Run `./gradlew generateGrammarSource` to regenerate parser code
   b. Register `AggregatorsListener` in `src/main/go/parser/cpl_parser.go` — Not Started
      - Create an instance of the aggregators listener (holding a reference to the main `CPLListener`)
      - Add it to the ANTLR parse-tree walker in the appropriate initialization method (follow existing pattern for `CPLMetricsListener`, `CPLMeasuresListener`, etc.)
   c. Add `aggregatorsSection` dispatch in `src/main/go/parser/cpl_document_listener.go` — Not Started
      - Add an `EnterAggregatorsSection` / `ExitAggregatorsSection` pair in the root listener (`CPLListener`) that delegates to the `AggregatorsListener` — follow the same delegation pattern used for `EnterMeasuresSection` / `ExitMeasuresSection`
      - Alternatively, add a `statement`-level dispatch that routes `aggregatorsSection` context to the `AggregatorsListener` — match the pattern used for other imported section grammars
   d. Update `shouldPushExpressionDefs()` in `src/main/go/parser/parser_support.go` — Not Started
      - Add `symboltable.AggregatorScope` to the switch statement's case list so that concept references, function calls (`fnCall` rule via `ExitIdentifier`), and arithmetic operators (`+`, `-`, `*`, `/`) accumulate in the aggregator scope's `Definitions` slice
   e. Verify the full pipeline end-to-end — Not Started
      - Write an integration test (or extend existing parser test) that parses a CPL file containing an `aggregators` section
      - Assert the symbol table root contains an `AggregatorsScope` child
      - Assert the child's `AggregatorScope` has `Data.(*ParserAggregatorDefinition)` with correct from/to
      - Assert the scope's `Definitions` slice contains postfix entries (concept identifiers, operators) in the correct order
      - Run `go test ./src/main/go/parser/... -count=1 -timeout 120s`
   f. Verify malformed syntax is rejected — Not Started
      - Parse CPL snippets with missing `from` clause, missing `to` clause, empty `from` list (`from []`), missing expression after `=`
      - Assert parse errors are produced (the parser should report clear error messages)
      - Assert no partial symbol table entries are left for malformed input

### Task 5: Create parse_and_analyze_only examples — Not Started [Small]

Create CPL example files for the aggregators feature that exercise the parser pipeline, covering both valid and invalid inputs.

**Intent**: Provide test fixtures that validate the full parsing pipeline end-to-end and serve as regression tests for future changes.

**Acceptance Criteria**:
- `parse_and_analyze_only/without_errors/` contains CPL examples that parse successfully
- `parse_and_analyze_only/with_errors/` contains CPL examples that produce expected parse errors
- All without_errors examples parse without errors when run through the full parser pipeline
- All with_errors examples produce parse errors (not panics or silent success)

   a. Create `parse_and_analyze_only/without_errors/` CPL examples — Not Started
      - `aggregators-single-concept.cpl`: `from [Country] to [Country]` with simple expression reference
      - `aggregators-multi-concept.cpl`: `from [Country, Product] to [Product]` with `sum(Country)` function call
      - `aggregators-scalar-reduction.cpl`: `from [Country] to []` — scalar (dimensionless) reduction
      - `aggregators-expression-arithmetic.cpl`: expression with arithmetic operators in postfix sequence
      - `aggregators-zero-declarations.cpl`: `aggregators` section with no declarations (valid edge case)
   b. Create `parse_and_analyze_only/with_errors/` CPL examples — Not Started
      - `aggregators-missing-from.cpl`: declaration with `to` but no `from` clause
      - `aggregators-missing-to.cpl`: declaration with `from` but no `to` clause
      - `aggregators-empty-from-list.cpl`: `from [] to [Country]` — empty from concept list
      - `aggregators-malformed-syntax.cpl`: mismatched brackets, missing equals, extra tokens
   c. Verify all examples — Not Started
      - Run the CPL parser on each without_errors example and confirm zero parse errors
      - Run the CPL parser on each with_errors example and confirm parse errors are produced
      - Run tests: `go test ./src/main/go/parser/... -count=1 -timeout 60s`

## Test-First Development

This story follows the Logical TDD Lifecycle. For each coding task (Tasks 1–4), tests are written as part of the implementation steps:

- **Task 1c–1d**: Grammar generation and compilation checks serve as the tests for grammar correctness
- **Task 2e**: Unit tests for struct initialization and synthetic key generation are written before final compilation verification (2f)
- **Task 3e**: Listener unit tests verify symbol table structure, synthetic keys, and postfix ordering before final test pass (3f)
- **Task 4e–4f**: Integration tests verify the full pipeline and error handling after wiring is complete

Tests are integrated within each task rather than separated into standalone test tasks.

## Constraints

- No `SUM`, `COUNT`, or `SIZE` reserved tokens — function names are ordinary identifiers in `fnCall` position; the analysis phase resolves them via the FunctionRegistry
- Synthetic keys (`__aggregator__[...]->[...]`) must never appear in user-visible output
- The `indent` token enforces readable indentation but does NOT define scope boundaries for aggregators (aggregators are flat — no nested child scopes beyond the aggregator scope itself)
- Parser validates syntax only — no cross-reference or semantic validation (that is Story256)
- Aggregators are unnamed — the grammar has no `identifier` before the `from` clause; the parser assigns keys
- The `toTarget` rule supports an empty alternative (`to []`) for scalar reduction

## Intent

This story delivers the parsing phase for aggregator declarations. Aggregators are unnamed — the parser assigns deterministic synthetic keys that never collide with user-definable names. Expression handling reuses the postfix pattern established by metrics: concept references, function calls (`fncall`), and arithmetic operators are pushed onto the aggregator scope's `Definitions` slice in postfix order. Function names are not reserved words — they are parsed as generic identifiers in the function call position, with the analysis phase resolving them via the FunctionRegistry. The root grammar listener must be scope-type-aware (via `shouldPushExpressionDefs()`) to correctly route expression definitions to aggregator scopes (vs. metric scopes). The parser stores only raw strings — all cross-reference resolution is deferred to analysis (Story256).

## Acceptance Criteria

1. `CPLAggregators.g4` grammar compiles via ANTLR without errors
2. `AGGREGATORS`, `FROM`, `TO`, `LPAREN`, `RPAREN` lexer tokens are recognized by the lexer
3. `ParserAggregatorDefinition` struct stores `from` concepts (raw strings), `to` concept (or empty with `ToEmpty=true`), and source location
4. `cpl_aggregators_listener.go` pushes `AggregatorScopeDefinition` with `ParserAggregatorDefinition` in Data under synthetic key `__aggregator__[...]->[...]`
5. Aggregator scope's `Definitions` slice contains postfix expression sequence (concept identifiers, fnCall names, operator definitions)
6. Root grammar listener routes concept references, fnCalls, and operators to aggregator scopes via `shouldPushExpressionDefs()`
7. Symbol table for aggregators is flat (no nested child scopes within `AggregatorScope`)
8. Parser validates syntax but performs NO cross-reference or semantic validation
9. `aggregators` sections with zero declarations produce no parse errors
10. `parse_and_analyze_only/` examples parse successfully (without_errors) or produce parse errors (with_errors)
