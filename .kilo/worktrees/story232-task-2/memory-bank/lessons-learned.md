# Lessons Learned

**Vitest must be run by the user, not Kilo [TEMPORARY WORKAROUND]**
: There is a bug in Kilo Code that prevents it from running vitest on its own. Until the bug is resolved in some future version of the software, when it's necessary to run vitest, ask the user to do it and report on or paste in the output.

**Use `gh run list` to check CI status**
: When verifying if CI has succeeded after pushing a branch, use the command `gh run list [--branch <branch_name>]` instead of trying to access GitHub via browser. This is more reliable and provides quick status information.

**Unified Domain Model Pattern**
: Instead of separate models for each layer (DTO/Domain/Record), use a single hand-authored struct that serves as domain type, persistence record, and API contract simultaneously. Use `json` tags for the API contract. The `alias:"<table>[.<column>]"` tag is the only escape hatch for cases where Go-Jet's qrm needs redirection (rare). **Do not use the `sql` struct tag** — Go-Jet's qrm and `INSERT … MODELS` both ignore it. This eliminates the "mapping tax" and ensures the API contract and database schema stay in sync. Only use DTOs for complex aggregations or when the API fundamentally differs from the domain logic.

**Go-Jet Reflection over Generated Structs**
: While Go-Jet generates its own model structs, avoid using them in the service layer. Instead, use Go-Jet's reflective mapping to hydrate your own Unified Domain Models. For this to work the domain type's Go name must normalize (lowercase, strip `_`/`-`/space) to its table name and direct field names must normalize to column names; embedded persistence mixins (`domain.Keyed`, `domain.Bounded`, `domain.Audit`) carry an `alias:"<table_name>"` tag so qrm rebinds the column-prefix lookup to the embedding entity's table. This keeps the domain logic clean of library-specific dependencies.

**Wrappers around generated models are prohibited; domain types implement `LockableEntity` directly**
: Wrapper types of the form `type fooWrapper struct { *model.Foo }` with an `Unwrap() any` method are forbidden. They were a workaround for qrm hydration failing when the domain type name didn't match its table; the correct fix is the naming discipline above. `serde/rdbms.LockableEntity` requires only `EntityID() int64` and `NaturalKey() string`; the domain struct _is_ the persistence record. For entities whose natural-key column is literally `key`, the `Keyed.Key` field collides with a method named `Key()`, so the accessor is named `NaturalKey()` and reads `e.Key`. The single exception is test fixtures inside `serde/rdbms/*_test.go` that embed a generated model to avoid an import cycle with `domain/entities`; those fixtures register themselves in `BitemporalTableRegistry` by their own type.

**Global PDD scripts location**
: PDD (Prompt Driven Development) scripts are stored in `~/.kilocode/scripts/` directory. This is the global location for reusable PDD scripts. When looking for PDD scripts, check this directory first. Per-project scripts may be stored in a local `scripts/` directory if one exists.

**Prefer `.sh` extension for shell scripts to clarify file type**
: Shell scripts should carry a `.sh` file extension even when the platform already knows the file is executable (e.g., via a shebang line or `chmod +x`). The extension makes it immediately obvious to a casual reader browsing the directory that the file is a shell script, without needing to open it. This applies to both project scripts (e.g., `.kilo/setup-script.sh`, `.kilo/run-script.sh`) and general-purpose utility scripts. Kilo's Agent Manager recognizes both `.kilo/run-script` and `.kilo/run-script.sh` as valid run-script filenames, so the `.sh` suffix imposes no compatibility cost.

**AI Collaboration: Overriding Completion Bias**
: To maintain architectural rigor, use a Structural Definition phase before implementation. Agents are biologically (mathematically) inclined to provide complete solutions. By mandating that Phase A consists ONLY of signatures, docstrings, and `pass` statements, we force the design conversation to happen at a logical level without being distracted by implementation details. This facilitates a "Logical TDD" (BDD) workflow where tests are written against the structure before any logic is permitted to exist.

**Enforcing TDD Integrity via Deep Assertions**
: When utilizing AI agents to write TDD iterations, tests that only assert surface-layer error codes can be automatically "gamed" by the AI via hard-coded values and dummy conditionals (e.g. `if entity_name == "expired_ns": return None, ErrorCode.VALIDATION_FAILED`). To preserve TDD integrity, tests must utilize dynamic, randomized input generation. This strictly forbids hard-coding of **ANY** data type—including strings, numeric inputs, dates, times, timestamps, floating point numbers, booleans, lists, matrices, or complex objects. Tests must mandate deep persistence validation and assertions must verify database mutations and/or strict mock contracts (`assert_called_once_with`), expressly forbidding magic values.

**Distinguish Agent Personas from User Personas**
: There has been confusion between AI agent personas (roles the AI takes on) and user/customer personas (human users of the ChronoCone system). Agent personas have rules in `.kilocode/rules/` (e.g., work-planner, story-implementor), while user personas are defined in specifications (e.g., System Integrator, Analyst, Planner). Always clarify which type of persona is being discussed to avoid misunderstandings.

**SI vs story-implementor Confusion**
: The System Integrator (SI) user persona and the story-implementor agent persona have been confused. The SI is a human consultant who implements ChronoCone for customers, while the story-implementor is an AI agent that executes story files. When referring to "implementor," always clarify whether you mean the human SI role or the AI story-implementor agent.

**work-planner vs Planner Confusion**
: The work-planner agent persona and the Planner user persona have been confused. The work-planner is an AI agent that creates plans and stories, while the Planner is a human user who operates plans day-to-day. The work-planner agent persona was previously named just "Planner," which added to the confusion. Always use the full names "work-planner" (agent) and "Planner" (user) to distinguish them.

**Always look for a PDD script first**
: Before performing any function, always check for an existing PDD (Prompt Driven Development) script in the `scripts/` directory or `~/.kilocode/scripts/`. PDD scripts provide step-by-step instructions for common tasks and ensure consistency. If a script exists for the task at hand, use it rather than performing the task directly. This prevents duplication of effort and ensures best practices are followed.

**Use `go mod tidy` for Dependency Management**
: Use `go mod tidy` to manage Go dependencies. This command ensures that the `go.mod` file matches the source code in the module, adding any missing modules and removing unused ones. This is the standard Go approach for maintaining clean and reproducible builds.

**Read CPL directory README before working in `src/chronocone_planning_language`**
: When performing activities that involve the `src/chronocone_planning_language` directory, always read the [`README.md`](../src/chronocone_planning_language/README.md) file in that directory first to understand its organization, build system, and architectural decisions. This ensures you have the proper context before making changes.

**ANTLR Listener Organization Convention**
: When implementing ANTLR listeners for a grammar split across multiple `.g4` files, listener implementations must be organized to mirror the grammar structure:

1. One listener file per grammar file: Each `.g4` grammar file should have a corresponding listener file in the target language (e.g., `CPLUnits.g4` → `cpl_units_listener.go`).
2. Co-locate Entry and Exit methods: All `Enter*` and `Exit*` methods for rules defined in a specific grammar file must be kept together in the same listener file. Never split Enter and Exit methods for the same rule across different files.
3. Main grammar listener: The root grammar file's listener (e.g., `ChronoconePlanningLanguage.g4` → `chronocone_planning_language_listener.go`) should contain the main listener struct definition and any shared helper methods.
4. Method distribution: When a grammar imports other grammars (e.g., `import CPLUnits;`), the listener methods for imported rules should be in the listener file corresponding to that imported grammar, not the main grammar's listener file.

: This convention ensures:

- Clear mapping between grammar structure and listener implementation
- Easier maintenance when grammar rules change
- Simplified understanding of which listener handles which rules
- Consistent organization across the codebase

**Consult Go Testing Guide when writing Go tests**
: Always consult the [`Go Testing Best Practices Guide`](../resources/go/testing-guide.md) when:

- Writing Go test code
- Creating stories that involve Go testing
- Creating plans that include Go test implementation

: The guide provides essential patterns for code organization (test file naming, package structure), test construction (table-driven tests, parallel tests, subtests), and testing utilities (`t.Cleanup()`, `t.Helper()`, `t.TempDir()`). Following these patterns ensures consistency with Go idioms and best practices across the codebase.

**Use symbol table-driven parsing for CPL compiler extensions**
: When extending the CPL compiler with new declaration types, use the **symbol table-driven parsing pattern** instead of type switches in identifier handlers.

: Why this matters:

- Avoids Open/Closed Principle violations (no need to modify centralized switch statements)
- Eliminates tight coupling between identifier handling and declaration types
- Prevents fragile state management with transient flags like `currentUnit`, `currentConcept`
- Provides clean separation between parsing (populating symbol table) and analysis (assembling declarations)

: Consult the pattern before extending:
: Before adding new declaration types to the CPL compiler, read the [Symbol Table-Driven Parsing Pattern](../specifications/patterns/symbol-table-driven-parsing.md). This pattern documents the correct approach for:

- Pushing/popping scope tables during parsing
- Defining symbols in the current scope without type switches
- Walking the symbol table tree during analysis to assemble declarations

: Anti-pattern to avoid:
: Do NOT use type switches in `ExitIdentifier` to determine context based on parent node types. This approach requires modifying the identifier handler every time a new declaration type is added.

**Handling Indentation in CPL Grammar and Analysis**
: In the Chronocone Planning Language (CPL), indentation is significant for defining hierarchical structures (like concepts) and multi-line declarations (like connectors). The grammar handles this using an `INDENT` token defined in the lexer as `[\r\n][ \t]*`.
:
: **Grammar Pattern:**
: Declarations that span multiple lines or require indentation should start with the `indent` rule (which matches the `INDENT` token). For example:
: `antlr
: connectorDeclaration
:     : indent identifier
:       indent DIMENSIONS_KEYWORD LBRACKET identifierList RBRACKET
:       indent GRAIN_KEYWORD identifier
:       indent CELLTYPE_KEYWORD identifier
:     ;
: `
:
: **Analysis Pattern:**
: The `INDENT` token captures both the newline and the leading whitespace. During the analysis phase, the length of the whitespace portion of the token's text is used to determine the indentation level. This level is crucial for establishing parent-child relationships in hierarchical structures.
:
: **No DEDENT Token Required:**
: Unlike Python-style indentation-aware parsing, CPL does NOT use explicit DEDENT tokens to mark the end of indented blocks. The grammar uses only INDENT tokens to indicate the start of new indented lines. Block structure is determined by observing when indentation levels change between consecutive lines, not by explicit delimiters. This significantly simplifies both the lexer and the grammar rules.
:
: Anti-pattern to avoid:
: Do NOT add DEDENT tokens to the lexer or grammar. Rules like `foo: INDENT bar+ DEDENT;` will not work with the current CPL lexer design. Instead, use patterns like `foo: indent bar+;` where each nested element starts with its own `indent`.

**Do not hard-code built-in type names in parser grammar rules**
: Parser grammar rules should only recognize syntactic structure, not interpret semantic meaning. Built-in type names (like `Number`, `Instant`, `Interval`) should be parsed as generic identifiers, with the determination of whether an identifier refers to a built-in type or a user-defined type deferred to the analysis phase.

: Why this matters:

- Mixing parser-time and analysis-time responsibilities creates tight coupling between syntax and semantics
- The parser should not need to change when new built-in types are added or renamed
- The `identifier` rule in the root grammar already accepts built-in type keywords as valid identifiers (e.g., `identifier: IDENTIFIER | NUMBER_TYPE | INSTANT_TYPE | INTERVAL_TYPE | UNIT | TRUE | FALSE | NIL;`)
- Duplicating this knowledge in sub-grammar rules (like `baseType: NUMBER_TYPE | INSTANT_TYPE | INTERVAL_TYPE | identifier;`) is redundant and error-prone

: Correct pattern:
: ```antlr
baseType: identifier;

````
: The analysis phase then resolves the identifier string (e.g., `"Number"`, `"True"`) to its semantic meaning using [`createTypeReference()`](src/chronocone_planning_language/src/main/go/analysis/types_analysis.go:106) or literal handling.

: **Exception — Dedicated tokens for unambiguous literal contexts:** Dedicated lexer tokens (e.g., `TRUE: 'True';`) are still defined for unambiguous literal contexts like the `literal` rule (`literal: INTEGER | TRUE | FALSE | NIL;`), where the token can only mean the literal value. The dedicated token ensures the literal is recognized without ambiguity. However, the token keyword MUST also appear in the `identifier` rule so that existing CPL sources using these words as names (metric/concept/connector) do not break — the analysis phase determines semantic meaning based on context.

: Anti-pattern to avoid:
: ```antlr
baseType: NUMBER_TYPE | INSTANT_TYPE | INTERVAL_TYPE | identifier;
````

: This requires the listener to handle built-in tokens specially (since `EnterIdentifier` is not called for them) and duplicates the list of built-in types in both the lexer and the parser. |

**Bitemporal Determinism (Valid Time vs System Time)**
: Prohibition: Never use time.Now() or ambient clocks in service/domain layers.
: Constraint: All time-dependent logic must receive both ValidTime and SystemTime as input parameters.
: Logic: Business rules compute only against Valid Time (the model axis—when facts are true in the world). System Time (the DB transaction axis) filters query results but never drives calculations or user-facing logic.
: Implementation: Domain services call query helpers like `ActiveAt(validTime, systemTime)` to fetch records visible at both points in time. System Time defaults to CURRENT_TIMESTAMP if not provided (nil). This ensures: (1) determinism for a given ValidTime regardless of current wall-clock, (2) correct temporal filtering across both axes, (3) forward compatibility with future UI features like "what was the situation on this past date using data known at that time?"
: Requirement: Logic must be deterministic; re-running a scenario for a specific ValidTime and SystemTime must yield identical results regardless of when the code executes.

**Truth in Documentation: Synchronize Story State with Verified Reality**
: **Symptom**: Marking a story as `COMPLETED ✅` or checking off Acceptance Criteria `[x]` before they are verified by automated integration tests.
: **Root Cause**: Premature status updates create "Completion Bias" in future conversation turns. Because agents prioritize the Memory Bank as a source of truth, a false "Complete" status prevents the agent from noticing missing logic or tests, as it assumes the work is already done.
: **Rule**: Never mark a story status as `COMPLETED` or check off an Acceptance Criterion until the corresponding integration tests have passed in the current workspace. Documentation must trail verification, never lead it.
: **Incremental Updates**: Each time you implement a task within a story, mark that task and only that task complete without changing anything else about the story file. This prevents "drive-by rewrites" where unrelated parts of the story file are accidentally reworded or deleted during status updates.

**Using `ctx` parameter for tree-walking state avoids coupling**
: By passing the context as an explicit parameter rather than relying solely on the transformer's internal `ctx` field, visitor methods remain testable with minimal fixture setup. Tests can construct an ad-hoc `CILContext` and `Computation` without needing a fully configured transformer. This decouples the visitor method signature from the transformer's lifecycle and makes unit testing individual CIL node handlers straightforward.

**Establish patterns early in CIL visitor implementation**
: The first CIL visitor method (`VisitCILLiteral`) sets the template for all subsequent CIL visitors (arithmetic, conditionals, metric references). Getting the nil safety and context passing conventions right in the first visitor prevents refactoring churn later in the story. Future visitors should follow the same `VisitX(ctx *CILContext, node *X) error` signature and three-tier nil checking pattern established by `VisitCILLiteral`.

**Tests must fail loudly when dependencies are missing**
: Unit tests must never be skipped or fail silently when required files, tools, or external dependencies are unavailable. A skipped test hides the fact that the test environment is incomplete or broken. If a test requires a file on disk, a command in PATH, or any other external resource, it must fail with a descriptive error (`t.Fatalf`) rather than calling `t.Skip`. This ensures that CI and local test runs surface infrastructure or setup problems immediately instead of masking them as passing builds.

**Do not assign documentation tasks to the technical writer before the underlying code exists**
: Tasks routed to `technical-writer-for-story-implementor` (e.g., drafting PDD scripts, README updates, or usage guides) must not be started until the code, utilities, or procedures they document are fully implemented and validated. Attempting to write documentation for non-existent code confuses the agent, produces incorrect or speculative content, and requires rework once the actual implementation is in place. Schedule technical-writer tasks as serial steps that run after the code tasks they depend on have completed.

**Grain dimensions and non-grain dimensions are structurally identical at the verification/example level**
: In CPL, the distinction between "grain" dimensions and other ("concept") dimensions is conceptual: grain dimensions deal with time and are associated with every metric via the compulsory `grain <measure>` declaration, while other dimensions are arbitrary. However, all dimensions — whether grain or not — are derived from concepts and have the same structure (a named column with concept values). By the time data reaches the verification/examples level (parquet files, expected output files, ODS spreadsheets), there is no structural difference between grain and non-grain dimensions. Tools that operate at this level (generators, verifiers) should treat all dimensions uniformly as an ordered list of named columns rather than bifurcating them into separate "grain" and "dimensions" categories. The grain-vs-non-grain distinction is only meaningful during CPL parsing/analysis where the `grain` keyword enforces that a time-related measure exists.

**Always remove unused parameters during code review and refactoring**
: When reviewing or refactoring code, unused function parameters should always be removed. They add noise, suggest functionality that does not exist, and increase the maintenance burden of every call site. A parameter that is never read within the function body has no reason to exist. If a future story needs the parameter, it can be reintroduced at that time — restoring a deleted parameter is trivial, while keeping spurious parameters around indefinitely creates ongoing confusion. This applies equally to new code and pre-existing code encountered during refactoring.

**core.cpl in example directories is for shared infrastructure only**
: Example directories follow a two-level structure: a group directory (e.g., `metrics_arithmetic/`) contains a `core.cpl` with shared infrastructure, and subdirectories (e.g., `simple_addition/`) contain example-specific `.cpl` files, `iospec.json`, and data files. Example-specific definitions (connectors, measures, metrics) must NEVER appear in `core.cpl`. `core.cpl` is reserved exclusively for shared infrastructure that would be used by multiple examples or would ship pre-configured in a real-world system — things like common time period concepts (`Month`, `Year`), base types (`Count = Number(precision=0)`), and unit declarations. Placing example-specific declarations in `core.cpl` causes them to be loaded into every example under that group directory (since `compile_and_run.sh` always includes the parent `core.cpl`), which can lead to duplicate declaration errors, unexpected behavior, and maintenance confusion. When in doubt, put a declaration in the example-specific `.cpl` file, not in `core.cpl`.
