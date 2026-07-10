# Story264: Spreaders Increment 1 (Parse)

## Goal

Implement the parsing phase for basic spreader declarations as Increment 1 of Epic008. Spreaders are unnamed — the parser assigns deterministic synthetic keys (`__spreader__[from]->[to]`) and stores expressions in postfix order within the spreader's scope.

This story delivers: `CPLSpreaders.g4` grammar with spreader declaration rules; `SpreadersListener.go` that pushes symbol table entries under synthetic keys; `ParserSpreaderDefinition` struct for parsed data; synthetic key generation helper; lexer token verification for spreader syntax; and root grammar integration to route expression definitions (concept references, function calls, identifiers, operators) to spreader scopes. The key difference from aggregators is that spreader `to` clauses contain multi-concept lists, and spreader expressions may contain generic identifier references (parsed as `Kind: "identifier"`, deferring metric/concept resolution to analysis). All cross-reference resolution is deferred to the analysis phase.

## References

- [ERO008: Spreaders Incremental Implementation](../specifications/design/cpl/eros/ERO008_spreaders-incremental-implementation.md)
- [ERO008: Spreaders Parsing Phase](../specifications/design/cpl/go-parser-target/eros/ERO008_spreaders-parsing-phase.md)
- [Postfix Expression Parsing Pattern](../patterns/postfix-expression-parsing.md)
- [Symbol Table-Driven Parsing Pattern](../patterns/symbol-table-driven-parsing.md)

## Dependencies

- Story254 (aggregators-increment1-matrix-mult) — matrix multiplication infrastructure must exist; spreader parsing references tensorops concepts in test expectations
- Story255 (aggregators-increment2-parse) — spreader parsing mirrors the aggregator parsing pattern; must be complete for shared `FROM`/`TO` lexer tokens, `fnCall` rule, and scope-type-aware root grammar listener routing pattern
- Epic001 (concepts) — parser stores concept names as raw strings; analysis resolves them later

## Dependent Stories

- Story265 (spreaders-increment1-analyze) — analysis phase reads the symbol table structure created here

## Tasks

### Task Sizing

Tasks 1, 2, and 5 are routine, narrow-scope work that follows established grammar, struct-definition, and example-directory patterns. Tasks 3 and 4 involve non-trivial listener implementation logic and coordinated parser state machine integration across multiple files and are not annotated `[Small]`. Task 2 touches two files in different packages (parser and symboltable); the `symbol_table.go` enum additions are mechanically trivial.

### Parallel Execution

- **Group A** (Tasks 2, 3): Task 2 creates `parser_spreader_definition.go` (parser package) and modifies `symbol_table.go` (symboltable package); Task 3 creates `SpreadersListener.go` (parser package). Fully disjoint file sets. Task 3's subagent receives type specifications (struct fields, enum value names, synthetic key signature) to write syntactically correct code; the merged result compiles when both worktree branches are combined.

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
- Watch for conflicts in: `src/main/go/symboltable/symbol_table.go` (Task 2 adds `SpreadersScope`/`SpreaderScope` to `ScopeType` and `SpreadersScopeDefinition`/`SpreaderScopeDefinition` to `DefinitionType`; git merge may need enum value ordering resolved)

### Task 1: Create CPLSpreaders.g4 grammar and ensure lexer tokens exist — Completed [Small]

Create the ANTLR parser grammar for spreader declarations and verify that the required lexer tokens are present in the CPL lexer.

**Intent**: Define the grammar rules that the parser uses to recognize spreader declarations (`from`, `to`, expressions), and ensure the lexer recognizes the `SPREADERS` keyword. The `FROM`, `TO`, `LPAREN`, and `RPAREN` tokens should already exist from the aggregator parsing story.

**Acceptance Criteria**:
- `src/main/antlr/CPLSpreaders.g4` exists with rules for `spreadersSection` and `spreaderDeclaration`
- `src/main/antlr/ChronoconePlanningLanguageLexer.g4` defines the `SPREADERS` token (`SPREADERS : 'spreaders';`)
- `FROM`, `TO`, `LPAREN`, `RPAREN` tokens are verified as present in the lexer
- ANTLR code generation (`./gradlew generateGrammarSource`) succeeds with no errors
- No `SUM`, `COUNT`, or `SIZE` tokens — function names are ordinary identifiers parsed via `fnCall`

   a. Create `src/main/antlr/CPLSpreaders.g4` as a parser grammar — Not Started
      - Name: `parser grammar CPLSpreaders;`
      - Options: `tokenVocab = ChronoconePlanningLanguageLexer;`
      - Rules: `spreadersSection` (`SPREADERS spreaderDeclaration+`), `spreaderDeclaration` (`indent FROM LSQUARE identifierList? RSQUARE TO LSQUARE identifierList? RSQUARE EQUALS expression`)
      - Both `from` and `to` clauses use `identifierList?` (both may be empty or contain identifiers)
      - Follow the convention of existing parser grammars (`CPLMeasures.g4`, `CPLAggregators.g4`)
   b. Verify lexer tokens in `src/main/antlr/ChronoconePlanningLanguageLexer.g4` — Not Started
      - Add `SPREADERS : 'spreaders';` if not already present (place near existing section keywords)
      - Verify `FROM : 'from';`, `TO : 'to';`, `LPAREN : '(';`, `RPAREN : ')';` exist (should be present from aggregator story)
      - If any are missing, add them
   c. Write a grammar-level verification by running `./gradlew generateGrammarSource` and confirming zero errors — Not Started
   d. Verify the generated Go parser compiles: `go build ./src/main/go/parser/antlr/...` — Not Started

### Task 2: Define ParserSpreaderDefinition and synthetic key helper — Completed [Small]

Create the Go struct that holds parsed spreader data, implement the synthetic key generation function, and add the spreader entries to the ScopeType and DefinitionType enums.

**Intent**: Define the data structures and identifiers the parser listener will populate, and establish the synthetic key convention that ensures spreader names never collide with user-definable names. The key difference from aggregators: `ToConcepts` is a slice (spreaders expand to multiple concepts), and the synthetic key format uses list notation on both sides.

**Acceptance Criteria**:
- `ParserSpreaderDefinition` struct exists with fields: `FromConcepts` (`[]string`), `ToConcepts` (`[]string`), `FromEmpty` (`bool`), `SourceLocation` (`SourceLocation`)
- `spreaderSyntheticKey(from []string, to []string) string` generates keys in the format `__spreader__[A,B]->[C,D]` (or `__spreader__[]->[Country]` for scalar spreading)
- `ScopeType` enum in `symbol_table.go` includes `SpreadersScope` and `SpreaderScope`
- `DefinitionType` enum in `symbol_table.go` includes `SpreadersScopeDefinition` and `SpreaderScopeDefinition`
- `String()` methods for both enums are updated with the new cases
- Unit tests verify struct initialization and synthetic key generation for common patterns

   a. Create `src/main/go/parser/parser_spreader_definition.go` — Not Started
      - Define `type ParserSpreaderDefinition struct` with `FromConcepts []string`, `ToConcepts []string`, `FromEmpty bool`, `SourceLocation symboltable.SourceLocation`
      - Import `github.com/chronocone/chronocone_planning_language/src/main/go/symboltable`
   b. Implement `spreaderSyntheticKey(from []string, to []string) string` in the same file — Not Started
      - Use `fmt.Sprintf("__spreader__[%s]->[%s]", strings.Join(from, ","), strings.Join(to, ","))`
      - Examples: `__spreader__[Product]->[Country,Product]`, `__spreader__[]->[Country]`
   c. Add `SpreadersScope` and `SpreaderScope` to the `ScopeType` enum in `src/main/go/symboltable/symbol_table.go` — Not Started
      - Place after the existing aggregator scope entries
      - Add corresponding cases to the `ScopeType.String()` method
   d. Add `SpreadersScopeDefinition` and `SpreaderScopeDefinition` to the `DefinitionType` enum in `src/main/go/symboltable/symbol_table.go` — Not Started
      - Place after `AggregatorScopeDefinition`
      - Add corresponding cases to the `DefinitionType.String()` method
   e. Write unit test in `src/main/go/parser/parser_spreader_definition_test.go` — Not Started
      - Verify `ParserSpreaderDefinition` struct can be initialized with all fields populated
      - Verify `spreaderSyntheticKey` produces correct strings for: multi-concept from/multi-concept to, single from/single to, scalar spreading (`from [] to [Country]`), both empty
      - Verify the `__spreader__` prefix is always present
   f. Verify the package compiles: `go build ./src/main/go/parser/... ./src/main/go/symboltable/...` — Not Started

### Task 3: Implement SpreadersListener — Completed

Create the listener that responds to ANTLR parse-tree events for the spreader grammar and populates the symbol table with `SpreaderScopeDefinition` entries containing `ParserSpreaderDefinition` data and postfix expression sequences.

**Intent**: Build the parse-tree-to-symbol-table bridge that extracts spreader declarations, assigns synthetic keys, and captures expression operands/operators in postfix order within the spreader scope.

**Acceptance Criteria**:
- `SpreadersListener.go` exists in the parser package and compiles
- `EnterSpreadersSection` pushes a `SpreadersScope` child scope onto the symbol stack
- `ExitSpreadersSection` pops the scope from the symbol stack
- `EnterSpreaderDeclaration` pushes a `SpreaderScope` child scope (expression definitions accumulate here)
- `ExitSpreaderDeclaration` extracts `from` concept names, `to` concept names, creates `ParserSpreaderDefinition`, computes synthetic key via `spreaderSyntheticKey`, calls `pushDefinition` with `SpreaderScopeDefinition`, and pops the spreader scope
- Unit tests verify correct symbol table tree structure, synthetic key correctness for multi-concept `to`, definition data integrity, and postfix expression content
- No cross-reference validation (deferred to analysis)

   a. Create `src/main/go/parser/SpreadersListener.go` — Not Started
      - Package `parser`; embed `*antlr.BaseCPLSpreadersListener` (the ANTLR-generated base listener for the `CPLSpreaders` grammar)
      - Hold a `*CPLListener` reference (the main parser listener) for access to `symbolStack`, `AddError`, `pushDefinition`, `pushOperatorDef`, and logging utilities
   b. Implement `EnterSpreadersSection` / `ExitSpreadersSection` — Not Started
      - `EnterSpreadersSection`: peek current scope from `symbolStack`, create child with `AddChild(symboltable.SpreadersScope)`, push child onto stack
      - `ExitSpreadersSection`: pop scope from `symbolStack` (error handling via `AddError`)
      - Follow pattern from `cpl_aggregators_listener.go` (`EnterAggregatorsSection`/`ExitAggregatorsSection`)
   c. Implement `EnterSpreaderDeclaration` — Not Started
      - Peek current scope, create child with `AddChild(symboltable.SpreaderScope)`, push child
      - This child scope will accumulate expression definitions (concept references, fnCalls, identifiers, operators) in postfix order via the root grammar listener's routing
   d. Implement `ExitSpreaderDeclaration` — Not Started
      - Extract from concepts: if `ctx.IdentifierList(0)` is not nil, collect identifier text into `fromConcepts []string`; if nil, `fromConcepts = nil` and set `FromEmpty = true`
      - Extract to concepts: `ctx.IdentifierList(1)` must be non-nil; collect identifier text into `toConcepts []string` (slice, unlike aggregator's single `toTarget`)
      - Create `ParserSpreaderDefinition{FromConcepts, ToConcepts, FromEmpty, sourceLocation}`
      - Compute synthetic key via `spreaderSyntheticKey(fromConcepts, toConcepts)`
      - Call `pushDefinition(syntheticKey, symboltable.SpreaderScopeDefinition, &def, sourceLocation)` — the definition pointer is stored in the spreader scope's `Data` field
      - Pop the spreader scope from the symbol stack
      - The spreader scope's `Definitions` slice already contains postfix expression entries accumulated during parsing (concept references, fnCalls, identifiers, operators pushed by the root listener)
   e. Write listener unit tests in `src/main/go/parser/spreaders_listener_test.go` — Not Started
      - Test symbol table tree structure: parse a CPL snippet with `spreaders` section → verify root scope contains `SpreadersScope` child → verify `SpreaderScope` child with `SpreaderScopeDefinition` in Data
      - Test synthetic key: verify key format `__spreader__[Product]->[Country,Product]` for single from/multi to, `__spreader__[Country,Product]->[Country,Product]` for multi from/multi to, `__spreader__[]->[Country]` for scalar spreading
      - Test definition data: verify `ParserSpreaderDefinition.FromConcepts`, `ToConcepts` (slice), `FromEmpty` match input
      - Test postfix definitions: verify the spreader scope's `Definitions` slice contains `OperandDefinition` entries (concept identifiers, fnCall names, identifier references) in declaration order
      - Test zero declarations: a `spreaders` section with no declarations produces a `SpreadersScope` with zero children and no errors
      - Use the test helper patterns from `cpl_aggregators_listener_test.go` as a model
   f. Verify pass: `go test ./src/main/go/parser/... -count=1 -timeout 60s` — Not Started

### Task 4: Integrate SpreadersListener into CPLParser, root grammar, and expression routing — Completed

Wire the new spreader grammar and listener into the main CPL parser pipeline so that spreader declarations are recognized during parsing and expression definitions are correctly routed to spreader scopes. This task also adds generic identifier routing for spreader expressions — a new behavior not present in aggregators.

**Intent**: Connect all the pieces created in Tasks 1–3 so the full parser pipeline processes spreader sections end-to-end. The root grammar listener must be scope-type-aware and route definitions to both aggregator and spreader scopes. Identifier references in spreader expressions are pushed as `Identifier` definitions with `StringData` (generic, not resolved to concept) — resolution is deferred to analysis.

**Acceptance Criteria**:
- Root grammar (`ChronoconePlanningLanguage.g4`) imports `CPLSpreaders` and includes `spreadersSection` in the `statement` rule
- ANTLR code generation (`./gradlew generateGrammarSource`) produces `CPLSpreaders` listener base classes
- Main `CPLListener` (in `cpl_document_listener.go`) dispatches `spreadersSection` to the `SpreadersListener` via the existing section-routing pattern
- `shouldPushExpressionDefs()` in `parser_support.go` returns true for `SpreaderScope` so concept references, fnCalls, and operators accumulate in postfix order
- Root grammar listener routes `identifier` references to spreader scopes as `Identifier` definitions with `StringData` when current scope is `SpreaderScope`
- Full pipeline: parse CPL with `spreaders` section → symbol table contains `SpreaderScopeDefinition` with correct `ParserSpreaderDefinition` and postfix `OperandDefinition` entries
- Parser rejects malformed spreader syntax (missing `from`, missing `to`, empty `to` clause, missing expression body) with parse errors

   a. Modify `src/main/antlr/ChronoconePlanningLanguage.g4` to import and include CPLSpreaders — Not Started
      - Add `CPLSpreaders` to the `import` statement (alongside existing `CPLAggregators`)
      - Add `spreadersSection` to the `statement` rule: `| spreadersSection`
      - Run `./gradlew generateGrammarSource` to regenerate parser code
   b. Register `SpreadersListener` in `src/main/go/parser/cpl_parser.go` — Not Started
      - Create an instance of the `SpreadersListener` (holding a reference to the main `CPLListener`)
      - Add it to the ANTLR parse-tree walker in the appropriate initialization method (follow existing pattern for `CPLAggregatorsListener`, `CPLMetricsListener`, etc.)
   c. Add `spreadersSection` dispatch in `src/main/go/parser/cpl_document_listener.go` — Not Started
      - Add an `EnterSpreadersSection` / `ExitSpreadersSection` pair in the root listener (`CPLListener`) that delegates to the `SpreadersListener`
      - Follow the same delegation pattern used for `EnterAggregatorsSection` / `ExitAggregatorsSection`
   d. Update `shouldPushExpressionDefs()` in `src/main/go/parser/parser_support.go` — Not Started
      - Add `symboltable.SpreaderScope` to the switch statement's case list so that concept references, function calls (`fnCall` rule), and arithmetic operators (`+`, `-`, `*`, `/`) accumulate in the spreader scope's `Definitions` slice
   e. Update root grammar listener to route `identifier` references to spreader scopes — Not Started
       - In the root grammar listener's `ExitPrimaryExpr` handler, when the current scope is an expression-accumulating scope (including `SpreaderScope`), push an `Identifier` definition with `StringData{Value: <identifier text>}`.
       - NOTE: This is the key new behavior over aggregators — identifiers in spreader expressions are parsed as generic identifiers, NOT concept references. The analysis phase resolves them against metrics → concepts → lookups in priority order. The parser stores them as `Identifier` definitions with `StringData`, consistent with the existing postfix expression pattern.
   f. Verify the full pipeline end-to-end — Not Started
      - Write an integration test (or extend existing parser test) that parses a CPL file containing a `spreaders` section
      - Assert the symbol table root contains a `SpreadersScope` child
      - Assert the child's `SpreaderScope` has `Data.(*ParserSpreaderDefinition)` with correct from/to slices
      - Assert the scope's `Definitions` slice contains postfix entries (concept identifiers, fnCall names, identifier operands, operators) in the correct order
      - Run `go test ./src/main/go/parser/... -count=1 -timeout 120s`
   g. Verify malformed syntax is rejected — Not Started
      - Parse CPL snippets with missing `from` clause, missing `to` clause, empty spreader declaration, missing expression after `=`
      - Assert parse errors are produced (the parser should report clear error messages)
      - Assert no partial symbol table entries are left for malformed input

### Task 5: Create parse_and_analyze_only examples — Completed [Small]

Create CPL example files for the spreaders feature that exercise the parser pipeline, covering both valid and invalid inputs.

**Intent**: Provide test fixtures that validate the full parsing pipeline end-to-end and serve as regression tests for future changes.

**Acceptance Criteria**:
- `parse_and_analyze_only/without_errors/` contains CPL examples that parse successfully
- `parse_and_analyze_only/with_errors/` contains CPL examples that produce expected parse errors
- All without_errors examples parse without errors when run through the full parser pipeline
- All with_errors examples produce parse errors (not panics or silent success)

   a. Create `parse_and_analyze_only/without_errors/` CPL examples — Not Started
      - `spreaders-single-dimension.cpl`: `from [Product] to [Country, Product]` with expression referencing concepts
      - `spreaders-scalar-spreading.cpl`: `from [] to [Country]` — scalar (dimensionless) spreading
      - `spreaders-expression-fncall.cpl`: expression with function calls (`size`, `sum`) and arithmetic
      - `spreaders-identifier-expressions.cpl`: expression containing generic identifiers (representing future metric references)
      - `spreaders-zero-declarations.cpl`: `spreaders` section with no declarations (valid edge case)
   b. Create `parse_and_analyze_only/with_errors/` CPL examples — Not Started
      - `spreaders-missing-from.cpl`: declaration with `to` but no `from` clause
      - `spreaders-missing-to.cpl`: declaration with `from` but no `to` clause
      - `spreaders-malformed-syntax.cpl`: mismatched brackets, missing equals, extra tokens
      - `spreaders-invalid-expression.cpl`: expression with invalid syntax after `=`
   c. Verify all examples — Not Started
      - Run the CPL parser on each without_errors example and confirm zero parse errors
      - Run the CPL parser on each with_errors example and confirm parse errors are produced
      - Run tests: `go test ./src/main/go/parser/... -count=1 -timeout 60s`

## Test-First Development

This story follows the Logical TDD Lifecycle. For each coding task (Tasks 1–4), tests are written as part of the implementation steps:

- **Task 1c–1d**: Grammar generation and compilation checks serve as the tests for grammar correctness
- **Task 2e**: Unit tests for struct initialization and synthetic key generation are written before final compilation verification (2f)
- **Task 3e**: Listener unit tests verify symbol table structure, synthetic keys, and postfix ordering before final test pass (3f)
- **Task 4f–4g**: Integration tests verify the full pipeline and error handling after wiring is complete

Tests are integrated within each task rather than separated into standalone test tasks.

## Constraints

- No `SUM`, `COUNT`, or `SIZE` reserved tokens — function names are ordinary identifiers in `fnCall` position; the analysis phase resolves them via the FunctionRegistry
- Synthetic keys (`__spreader__[...]->[...]`) must never appear in user-visible output
- The `indent` token enforces readable indentation but does NOT define scope boundaries for spreaders (spreaders are flat — no nested child scopes beyond the spreader scope itself)
- Parser validates syntax only — no cross-reference or semantic validation (that is the analysis story)
- Spreaders are unnamed — the grammar has no `identifier` before the `from` clause; the parser assigns keys
- `FromConcepts` and `ToConcepts` are both slices (`[]string`); `FromEmpty` is `true` when the `from` clause is `[]`
- `to` clause must contain at least one identifier (validated by the `ExitSpreaderDeclaration` listener — the grammar allows empty lists for symmetry with `from`; size and containment checks are deferred to analysis)
- `FROM` and `TO` lexer tokens are shared with `CPLAggregators.g4` and registered once in `ChronoconePlanningLanguageLexer.g4`

## Intent

This story delivers the parsing phase for spreader declarations. Spreaders are unnamed — the parser assigns deterministic synthetic keys that never collide with user-definable names. The key differences from aggregators are: (1) `to` is a **slice** of concepts (not a single concept), because spreaders expand to a multi-concept dimensional set, and (2) spreader expressions may contain **identifier references** (for metric references in later increments). These identifiers are parsed as generic `Identifier` definitions with `StringData` — the parser makes no distinction between metric, concept, or lookup references. Resolution is fully deferred to the analysis phase. Expression handling reuses the postfix pattern from metrics and aggregators: concept references, function calls (`fnCall`), identifiers, and arithmetic operators (`*`, `/`) are pushed onto the spreader scope's `Definitions` slice in postfix order. Function names are not reserved words — they are parsed as generic identifiers in the function call position. The root grammar listener must be scope-type-aware (via `shouldPushExpressionDefs()`) to correctly route expression definitions to both aggregator and spreader scopes. The parser stores only raw strings — all cross-reference resolution is deferred to analysis.

## Acceptance Criteria

1. `CPLSpreaders.g4` grammar compiles via ANTLR without errors
2. `SPREADERS` lexer token is recognized by the lexer; `FROM`, `TO`, `LPAREN`, `RPAREN` are shared from the aggregator story
3. `ParserSpreaderDefinition` struct stores `from` concepts (raw strings, nil if empty), `to` concepts (slice), `FromEmpty` bool, and source location
4. `SpreadersListener` pushes `SpreaderScopeDefinition` with `ParserSpreaderDefinition` in Data under synthetic key `__spreader__[from]->[to]`
5. Spreader scope's `Definitions` slice contains postfix expression sequence (concept, fnCall, identifier, operator operand definitions)
6. Root grammar listener routes concept references, fnCalls, identifiers, and operators to spreader scopes via `shouldPushExpressionDefs()`
7. Identifier references in spreader expressions are pushed as `Identifier` definitions with `StringData` (generic, not resolved to concept)
8. Symbol table for spreaders is flat (no nested child scopes within `SpreaderScope`)
9. Parser validates syntax but performs NO cross-reference or semantic validation
10. `spreaders` sections with zero declarations produce no parse errors
11. `parse_and_analyze_only/` examples parse successfully (without_errors) or produce parse errors (with_errors)
12. Synthetic keys never leak into user-visible output
