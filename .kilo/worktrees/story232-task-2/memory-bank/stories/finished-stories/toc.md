# Finished Stories Table of Contents

This file contains a table of all finished stories for the project. These stories have been completed or superseded.

**[Story001_workspace-setup.md](Story001_workspace-setup.md)** — _Done_
: Set up CHRONOCONE as a multi-root workspace with Python backend and React/Node frontend environments, including CI/CD configuration

**[Story002_backend-structure.md](Story002_backend-structure.md)** — _Done_
: Establish backend project structure with src/tests directories, configure pytest, migrate main.py, and verify testing infrastructure

**[Story003_frontend-structure.md](Story003_frontend-structure.md)** — _Done_
: Set up React/TypeScript project with Vite, configure testing with Vitest, and verify development and CI environments

**[Story004_core-and-metrics-services.md](Story004_core-and-metrics-services.md)** — _Done_
: Implement Core Service with /endpoint endpoint and Metrics Service with /metadata endpoint (stub implementations with CLIs)

**[Story005_database-initialization-cli.md](Story005_database-initialization-cli.md)** — _Done_
: Create Python CLI to manage PostgreSQL Docker container and initialize database tables from schema SQL files

**[Story006_metrics-service-database-integration.md](Story006_metrics-service-database-integration.md)** — _Done_
: Modify Metrics Service /metadata and /list endpoints to read/write actual database tables instead of stub printouts

**[Story007_ui-navigation-and-pages.md](Story007_ui-navigation-and-pages.md)** — _Done_
: Create UI with navigation bar and three pages: Home (welcome), Definitions (metadata tables with add functionality), and Metrics (placeholder)

**[Story009_frontend-service-integration.md](Story009_frontend-service-integration.md)** — _Done_
: Modify frontend to read from and write to Metrics Service endpoints instead of using in-memory data

**[Story011_validity-field-implementation.md](Story011_validity-field-implementation.md)** — _Done_
: Re-enable validity fields in SQL schema and implement automatic validity range setting in Metrics Service

**[Story020_ontology-services.md](Story020_ontology-services.md)** — _Done_
: Establish a strict bitemporal Ontology Service layer using DTOs and Natural Keys to abstract domain identities and mutations

**[Story020_ontology-services_walkthrough.md](Story020_ontology-services_walkthrough.md)** — _Done_
: Documentation and walkthrough of the ontology services

**[Story021_metrics-service-refactor.md](Story021_metrics-service-refactor.md)** — _Done_
: Refactor the Metrics Service FastAPI application to exclusively use the Ontology Service layer and remove repository dependencies

**[Story022_tdd-robustness-and-magic-values.md](Story022_tdd-robustness-and-magic-values.md)** — _Done_
: Remove magic values and enforce repository-driven TDD testing

**[Story023_fully-containerized-development.md](Story023_fully-containerized-development.md)** — _Done_
: Transition the project to a fully containerized development environment using .devcontainer supporting Python, Node.js, and PostgreSQL

**[Story024_persona-selection-and-safety.md](Story024_persona-selection-and-safety.md)** — _Done_
: Codify persona selection logic and READ-ONLY boundaries for Ask mode to prevent premature implementation and completion bias

**[Story028_demo-environment-core-server-requirements.md](Story028_demo-environment-core-server-requirements.md)** — _Done_
: Explore and define requirements for the demonstration environment's core server and infrastructure

**[Story012A_core-service-patterns-compliance.md](Story012A_core-service-patterns-compliance.md)** — _Done_
: Bring Core Service implementation into compliance with all architectural and design patterns defined in specifications/patterns

**[Story040_core-service-compliance.md](Story040_core-service-compliance.md)** — _Done_
: Bring Core Service implementation into compliance with architecture docs by removing start subcommand and admin endpoint

**[Story041_core-service-layered-compliance.md](Story041_core-service-layered-compliance.md)** — _Done_
: Update Core Service to comply with updated Layered Architecture pattern: enforce separation of concerns between model layers and add Pydantic constraints

**[Story101_remove-docker-dependencies.md](Story101_remove-docker-dependencies.md)** — _Done_
: Remove Docker container management code from database module and initialization scripts

**[Story102_cpl-parser-environment-setup.md](Story102_cpl-parser-environment-setup.md)** — _Done_
: Set up Gradle/Java ANTLR4 environment with Go target and create "hello world" grammar

**[Story103_cpl-parser-units-and-dimensions.md](Story103_cpl-parser-units-and-dimensions.md)** — _Done_
: Extend CPL parser to support units and dimensions declaration sections

**[Story104_cpl-parser-connectors.md](Story104_cpl-parser-connectors.md)** — _Superseded_
: ~~Extend CPL parser to support connector declarations with error handling clauses~~ **SUPERSEDED**: Replaced by Story162

**[Story105_cpl-evaluator-cli.md](Story105_cpl-evaluator-cli.md)** — _Superseded_
: ~~Create CPL evaluator package and CLI program with test data driver support~~ **SUPERSEDED**: Replaced by synthesis architecture

**[Story109_cpl-parser-slog-logging.md](Story109_cpl-parser-slog-logging.md)** — _Done_
: Implement structured logging using Go's slog package for the CPL parser with configurable levels and output destinations

**[Story111_remove-dimensions-from-cpl.md](Story111_remove-dimensions-from-cpl.md)** — _Done_
: Remove the dimensions declaration section from the CPL parser including grammar file, Go code, tests, and examples

**[Story112_refactor-cpl-tests-to-go-best-practices.md](Story112_refactor-cpl-tests-to-go-best-practices.md)** — _Done_
: Refactor CPL parser test code to follow Go testing best practices including table-driven tests, testify assertions, and parallel execution

**[Story113_cpl-listener-error-slice.md](Story113_cpl-listener-error-slice.md)** — _Done_
: Add error[] slice to CPLListener and update listener methods to capture semantic errors instead of silently ignoring them

**[Story110_cpl-symbol-table-implementation.md](Story110_cpl-symbol-table-implementation.md)** — _Done_
: Implement the symbol table data structures and stack management for the CPL parser according to the design document

**[Story130_epic001-story-updates-implementation.md](Story130_epic001-story-updates-implementation.md)** — _Done_
: Update Epic001 stories (114-126) with new generator syntax and create new stories (128-129) for Increment 6 (Concept Range Generators)

**[Story114_cpl-concepts-parser-basic.md](Story114_cpl-concepts-parser-basic.md)** — _Done_
: Implement basic concepts parsing with Values generator support for CPL parser (Epic001 Increment 1)

**[Story115_cpl-concepts-analyzer-basic.md](Story115_cpl-concepts-analyzer-basic.md)** — _Done_
: Implement basic concepts analysis with lookup map and duplicate detection (Epic001 Increment 1)

**[Story116_cpl-concepts-synthesizer-basic.md](Story116_cpl-concepts-synthesizer-basic.md)** — _Done_
: Implement basic synthesizer with placeholder Go CLI generation (Epic001 Increment 1)

**[Story117_cpl-concepts-parser-hierarchy.md](Story117_cpl-concepts-parser-hierarchy.md)** — _Done_
: Implement concept hierarchy parsing with INDENT tokens and parent-child relationships (Epic001 Increment 2)

**[Story118_cpl-concepts-analyzer-hierarchy.md](Story118_cpl-concepts-analyzer-hierarchy.md)** — _Done_
: Implement concept hierarchy analysis with parent-child relationships from indentation (Epic001 Increment 2)

**[Story120_cpl-concepts-analyzer-multi-section.md](Story120_cpl-concepts-analyzer-multi-section.md)** — _Done_
: Implement cross-section duplicate detection in analyzer (Epic001 Increment 3)

**[Story122_cpl-concepts-analyzer-conditional.md](Story122_cpl-concepts-analyzer-conditional.md)** — _Done_
: Implement conditional value resolution and parent-specific mappings (Epic001 Increment 4)

**[Story123_cpl-concepts-parser-range.md](Story123_cpl-concepts-parser-range.md)** — _Done_
: Implement Range generator parsing for numeric concept values (Epic001 Increment 5)

**[Story124_cpl-concepts-analyzer-range.md](Story124_cpl-concepts-analyzer-range.md)** — _Done_
: Implement Range value expansion during analysis (Epic001 Increment 5)

**[Story151_epic002-parsing-analysis-separation.md](Story151_epic002-parsing-analysis-separation.md)** — _Done_
: Correct ERO002 documentation and Epic002 story files to properly separate parsing from analysis actions

**[Story152_epic003-ero003-documentation-update.md](Story152_epic003-ero003-documentation-update.md)** — _Done_
: Update Epic003, ERO003 documents, and stories 141-150 to clarify parsing/analysis separation and document migration from old unit implementation

**[Story132_cpl-types-analyzer-number.md](Story132_cpl-types-analyzer-number.md)** — _Done_
: Implement AnalysisTypeDefinition structure, Types lookup map, NumberConstraints with validation, CreateConstraints factory, duplicate detection

**[Story153_cpl-units-parser-basic-reimplementation.md](Story153_cpl-units-parser-basic-reimplementation.md)** — _Done_
: Reimplement parser for basic unit definitions with symbol table-driven parsing, CPLUnits.g4 grammar, UnitsScope/UnitScope, listener methods, comprehensive tests

**[Story142_cpl-units-analyzer-basic.md](Story142_cpl-units-analyzer-basic.md)** — _Done_
: Implement analyzer for basic unit definitions with AnalyzeUnitScope, AnalysisUnitDefinition, Units map, concept reference validation, duplicate detection. Reimplemented with symbol table-driven analysis for Epic003 Increment 1

**[Story143_cpl-measures-parser-basic.md](Story143_cpl-measures-parser-basic.md)** — _Done_
: Implement parser for basic measure definitions with CPLMeasures.g4 grammar, MeasuresScope/MeasureScope symbol table, listener methods, comprehensive tests

**[Story144_cpl-measures-analyzer-basic.md](Story144_cpl-measures-analyzer-basic.md)** — _Done_
: Implement analyzer for basic measure definitions with AnalyzeMeasureScope, AnalysisMeasureDefinition, Measures map, unit reference validation, duplicate detection

**[Story162_cpl-parser-connectors-basic.md](Story162_cpl-parser-connectors-basic.md)** — _Done_
: Implement the parsing phase for basic connector declarations (dimensions, celltype)

**[Story156_cpl-analyzer-connectors.md](Story156_cpl-analyzer-connectors.md)** — _Done_
: Validation that referenced dimensions (concepts) and celltype (measure) exist in the symbol table

**[Story158_cpl-parser-connectors-time.md](Story158_cpl-parser-connectors-time.md)** — _Done_
: Support for the grain keyword in connector declarations

**[Story159_cpl-analyzer-connectors-time.md](Story159_cpl-analyzer-connectors-time.md)** — _Done_
: Validation that the grain refers to an existing measure, and that the measure's underlying type is Interval or derived from Interval

**[Story160_cpl-synthesizer-connectors-time.md](Story160_cpl-synthesizer-connectors-time.md)** — _Done_
: Generation of Go code that maps the parquet time column to the connector's grain

**[Story163_cpl-units-measures-multiple-sections.md](Story163_cpl-units-measures-multiple-sections.md)** — _Done_
: Support multiple units and measures sections in parser and analyzer with cross-section duplicate detection

**[Story165_cpl-generated-code-compilation.md](Story165_cpl-generated-code-compilation.md)** — _Done_
: Make CPL compiler-generated Go programs compilable and runnable with real parquet-to-tensor data loading

**[Story166_chronocone-time-concepts-docs.md](Story166_chronocone-time-concepts-docs.md)** — _Done_
: Update active documentation and stories so Instant, Interval.starts_at, Duration, and grain semantics match the corrected Chronocone time model

**[Story167_cpl-parser-instant-constraints-complete.md](Story167_cpl-parser-instant-constraints-complete.md)** — _Done_
: Update the parser to recognize the complete set of Instant constraint names (10 constraints)

**[Story169_cpl-standard-definitions-consistency.md](Story169_cpl-standard-definitions-consistency.md)** — _Done_
: Ensure standard_definitions files and example .cpl files are consistent with the new Instant constraint names

**[Story155_cpl-parser-multiple-files.md](Story155_cpl-parser-multiple-files.md)** — _Done_
: Support for parsing multiple CPL files as a single compilation unit

**[Story131_refactor-cpl-parser-symbol-table-driven.md](Story131_refactor-cpl-parser-symbol-table-driven.md)** — _Done_
: Refactor CPL parser to follow symbol table-driven pattern, eliminating type switches in identifier handler

**[Story127_adr-to-ero-rename.md](Story127_adr-to-ero-rename.md)** — _Done_
: Rename ADR (Architecture Decision Record) to ERO (Epic Restraining Order) across the entire project

**[Story131_cpl-types-parser-number.md](Story131_cpl-types-parser-number.md)** — _Done_
: Implement grammar changes for types section, Number type with precision/min/max constraints, constraint parsing, TypesScope symbol table integration

**[Story154_cpl-measures-tests-and-examples.md](Story154_cpl-measures-tests-and-examples.md)** — _Done_
: Create comprehensive parser tests, analyzer tests, and error examples for Story143 and Story144 measures functionality

**[Story161_parquet-test-utility.md](Story161_parquet-test-utility.md)** — _Done_
: Create a test utility function and CLI to generate parquet files with random cell values based on dimension definitions

**[Story164_cpl-concepts-value-listener.md](Story164_cpl-concepts-value-listener.md)** — _Done_
: Update CPL concept listeners to support both identifier and numeric literal values in concept value lists

**[Story137_cpl-types-parser-instant.md](Story137_cpl-types-parser-instant.md)** — _Done_
: Add Instant base type, field-based constraint support (year through second)

**[Story138_cpl-types-analyzer-instant.md](Story138_cpl-types-analyzer-instant.md)** — _Done_
: Implement InstantConstraints as field-based struct (year, month, day, hour, minute, second, day_of_week, week_number), validation of constraint ranges and combinations

**[Story139_cpl-types-parser-interval.md](Story139_cpl-types-parser-interval.md)** — _Done_
: Add Interval base type, starts_at reference and unit constraints

**[Story140_cpl-types-analyzer-interval.md](Story140_cpl-types-analyzer-interval.md)** — _Done_
: Implement IntervalConstraints with cross-reference validation, starts_at must be Instant-derived, unit reference validation

**[Story041_relational-schema-implementation.md](Story041_relational-schema-implementation.md)** — _Done_
: Implement relational schema changes from Story027 requirements, starting with consolidation of v0 .sql files

**[Story043_temporal-enforcement.md](Story043_temporal-enforcement.md)** — _Done_
: Cross-cutting temporal enforcement: advisory locks, bitemporal orchestrator, future-dated assertions, temporal query helpers

**[Story043J_remedy_architectural_violations.md](Story043J_remedy_architectural_violations.md)** — _Done_
: Remedy Architectural Violations in Backend (Re-establish Domain Layer)

**[Story043K_domain_package_aggregate_roots.md](Story043K_domain_package_aggregate_roots.md)** — _Done_
: Move entity services into the parent domain package so aggregate-root boundaries are enforced by Go visibility

**[Story168_cpl-analyzer-instant-constraints-concept-driven.md](Story168_cpl-analyzer-instant-constraints-concept-driven.md)** — _Done_
: Refactor the analyzer to validate Instant constraints by looking up corresponding concepts at analysis time

**[Story174_cpl-analyzer-instant-constraints-completion.md](Story174_cpl-analyzer-instant-constraints-completion.md)** — _Done_
: Complete remaining Story168 work: update tests and examples for 0-based time ranges and re-enable skipped integration tests

**[Story175_cpl-mandatory-connector-grain-completion.md](Story175_cpl-mandatory-connector-grain-completion.md)** — _Done_
: Complete remaining Story173 work: update example CPL files to include mandatory grain clauses, run full test suite, and finalize story

**[Story170_cpl-analyzer-remove-hardcoded-time-concepts.md](Story170_cpl-analyzer-remove-hardcoded-time-concepts.md)** — _Done_
: Remove hardcoded standard time concept values from the CPL analyzer and rely on standard_concepts.cpl

**[Story171_cpl-connector-grain-decouple-parser-analyzer.md](Story171_cpl-connector-grain-decouple-parser-analyzer.md)** — _Done_
: Remove parser/analyzer coupling for connector grain by replacing ParserConnectorDefinition with symbol table GrainDefinition

**[Story172_cpl-improve-connector-analysis.md](Story172_cpl-improve-connector-analysis.md)** — _Done_
: Refactor connector analysis with structured per-connector records, remove mixed-identifier reconstruction, and defer grain interval-family validation post-derivation

**[Story173_cpl-mandatory-connector-grain.md](Story173_cpl-mandatory-connector-grain.md)** — _Done_
: Make the grain keyword mandatory in CPL connector declarations across parser, analyzer, tests, and examples

**[Story157_cpl-synthesizer-connectors.md](Story157_cpl-synthesizer-connectors.md)** — _Done_
: Generation of Go code that uses the go-arrow library to read data from a parquet file

**[Story176_cpl-connectors-multiple-grain-dimensions.md](Story176_cpl-connectors-multiple-grain-dimensions.md)** — _Done_
: Extend CPL connector grain declarations to accept multiple Interval-derived measure dimensions

**[Story177_update-parquet-gen-grain.md](Story177_update-parquet-gen-grain.md)** — _Done_
: Update parquetgen CLI and io.GenerateParquetFile to support hierarchical grain measure columns in generated test parquet files

**[Story179_update-parquet-read-for-grain.md](Story179_update-parquet-read-for-grain.md)** — _Done_
: Update Parquet reader to support grain measure columns via column-to-concept mapping, propagate mapping through synthesis/CIL pipeline

**[Story180_fix-tensor-grain-read.md](Story180_fix-tensor-grain-read.md)** — _Done_
: Fix synthesizer bug where GenerateConnectorMetadataMap omits Grain field, causing tensors to lack grain dimensions

**[Story181_reject-insufficient-parquet-data.md](Story181_reject-insufficient-parquet-data.md)** — _Done_
: Modify synthesized parquet reading logic to terminate with a descriptive error when input parquet files do not exactly match CPL connector structure

**[Story181_reject-insufficient-parquet-data.walkthrough.md](Story181_reject-insufficient-parquet-data.walkthrough.md)** — _Done_
: Walkthrough documentation for Story181: rejecting insufficient Parquet data

**[Story135_cpl-types-parser-multiple-sections.md](Story135_cpl-types-parser-multiple-sections.md)** — _Done_
: Support multiple types sections in same file, cross-section duplicate detection in analysis

**[Story136_cpl-types-analyzer-multiple-sections.md](Story136_cpl-types-analyzer-multiple-sections.md)** — _Done_
: Cross-section duplicate detection during symbol table walk

**[Story128_cpl-concepts-parser-concept-range.md](Story128_cpl-concepts-parser-concept-range.md)** — _Done_
: Implement Concept Range generator parsing for referencing value ranges from other concepts (Epic001 Increment 6)

**[Story129_cpl-concepts-analyzer-concept-range.md](Story129_cpl-concepts-analyzer-concept-range.md)** — _Done_
: Implement Concept Range value resolution with dependency ordering and circular reference detection (Epic001 Increment 6)

**[Story182_cpl-metrics-basic-parse-analyze.md](Story182_cpl-metrics-basic-parse-analyze.md)** — _Done_
: Increment 1 Parse+Analyze: Create CPLMetrics.g4 grammar, MetricsListener, and AnalyzeMetricScope for basic metric declarations with literal 0 calculation

**[Story183_cpl-metrics-basic-synthesize.md](Story183_cpl-metrics-basic-synthesize.md)** — _Done_
: Increment 1 Synthesis: Extend CILAbstractTransformer with VisitCILLiteral, generate computeX/Setup/Compute/Cleanup/main for basic metrics

**[Story184_cpl-metrics-arithmetic-parse-analyze.md](Story184_cpl-metrics-arithmetic-parse-analyze.md)** — _Done_
: Increment 2 Parse+Analyze: Extend root grammar with shared expression rules, arithmetic operators, postfix evaluation to CILBinaryOp/CILUnaryOp/CILInputRef

**[Story185_cpl-metrics-arithmetic-synthesize.md](Story185_cpl-metrics-arithmetic-synthesize.md)** — _Done_
: Increment 2 Synthesis: VisitCILBinaryOp, VisitCILUnaryOp, VisitCILInputRef for element-wise arithmetic on tensors

**[Story196_cpl-types-string-boolean.md](Story196_cpl-types-string-boolean.md)** — _Done_
: Add String and Boolean built-in types to analysis phase: StringConstraints (min/max length, regex pattern), BooleanConstraints, CreateConstraints, derivation

**[Story125_cpl-concepts-parser-before-after.md](Story125_cpl-concepts-parser-before-after.md)** — _Done_
: Implement before/after clause parsing for value ordering (Epic001 Increment 7)

**[Story125_cpl-concepts-parser-before-after.walkthrough.md](Story125_cpl-concepts-parser-before-after.walkthrough.md)** — _Done_
: Walkthrough documentation for Story125: CPL Concepts Parser - Before/After

**[Story126_cpl-concepts-analyzer-before-after.md](Story126_cpl-concepts-analyzer-before-after.md)** — _Done_
: Implement value ordering computation with insertion points (Epic001 Increment 7)

**[Story126_cpl-concepts-analyzer-before-after.walkthrough.md](Story126_cpl-concepts-analyzer-before-after.walkthrough.md)** — _Done_
: Walkthrough documentation for Story126: CPL Concepts Analyzer - Before/After

**[Story197_cpl-types-multiple-sections-tests.md](Story197_cpl-types-multiple-sections-tests.md)** — _Done_
: Add missing parser and analyzer tests for multiple types sections and cross-section duplicate detection

**[Story198_cpl-concepts-concept-range-examples.md](Story198_cpl-concepts-concept-range-examples.md)** — _Done_
: Create valid and invalid CPL example files demonstrating Concept Range parser and analyzer functionality

**[Story199_cpl-types-string-boolean-examples.md](Story199_cpl-types-string-boolean-examples.md)** — _Done_
: Create CPL example files demonstrating Story196 String and Boolean built-in types with valid declarations, derivations, and intentional errors

**[Story200_toc-revamp.md](Story200_toc-revamp.md)** — _Done_
: Convert all 57 toc.md files across the project from markdown table format to list-based format to eliminate git merge conflicts, and update 8 PDD scripts and 1 rules file

**[Story043B_active_historical_partitioning.md](Story043B_active_historical_partitioning.md)** — _Done_
: Temporal Enforcement Refactoring (Active/Historical Partitions)

**[Story043C_purge_store_abstraction.md](Story043C_purge_store_abstraction.md)** — _Done_
: Purged the legacy store/repository-pattern abstraction; migrated bitemporal infrastructure to serde/rdbms/

**[Story043D_reimplement_handlers_services_pattern.md](Story043D_reimplement_handlers_services_pattern.md)** — _Superseded_
: Reimplement API Handlers with Services Pattern (Superseded by Story 043J)

**[Story043E_ebm_splinter_correctness.md](Story043E_ebm_splinter_correctness.md)** — _Done_
: EBM Splinter Correctness: data-loss bug in future-record purge fixed; bitemporal-honesty invariants enforced

**[Story043F_read_path_correctness.md](Story043F_read_path_correctness.md)** — _Done_
: Read Path Correctness

**[Story043G_test_integrity.md](Story043G_test_integrity.md)** — _Done_
: Test Integrity

**[Story043H_registry_and_mutation_hardening.md](Story043H_registry_and_mutation_hardening.md)** — _Done_
: Registry and Mutation Hardening

**[Story043I_docs_and_scaffold_cleanup.md](Story043I_docs_and_scaffold_cleanup.md)** — _Done_
: Docs and Scaffold Cleanup

**[Story205_fix-connector-metrics.md](Story205_fix-connector-metrics.md)** — _Done_
: Fix the bug where EvaluatePostfixFloat64 returns nil when calling GetMetricComputation() for connector-based inputs by registering connectors as MetricComputation objects.

**[Story203_fix-synthesis-bugs-arithmetic-metrics.md](Story203_fix-synthesis-bugs-arithmetic-metrics.md)** — _Done_
: Fix synthesis-layer type mismatch and measure reference bugs preventing arithmetic metric generated Go code from compiling.

**[Story204_fix-nested-binary-op-codegen.md](Story204_fix-nested-binary-op-codegen.md)** — _Done_
: Fix nested arithmetic expression code generation by replacing per-operation Go statement blocks with a table-driven postfix evaluator that walks the CIL tree.

**[Story207_correct-metrics-arithmetic-examples.md](Story207_correct-metrics-arithmetic-examples.md)** — _Done_
: Correct all metrics_arithmetic examples to follow the reference pattern with proper connectors, Daily grain, and full 7-day data

**[Story206_fix-measure-conceptvalue-confusion.md](Story206_fix-measure-conceptvalue-confusion.md)** — _Done_
: Eliminate naming confusion between grain measure names and underlying concept names in code generation and runtime pipeline

**[Story202_cpl-metrics-arithmetic-examples-revamp.md](Story202_cpl-metrics-arithmetic-examples-revamp.md)** — _Done_
: Transform metrics_arithmetic examples into fully synthesizable, runnable examples with shared core.cpl, input data generation, and automated verification

**[Story193_cpl-metrics-short-form-parse-analyze.md](Story193_cpl-metrics-short-form-parse-analyze.md)** — _Done_
: Increment 7 Parse+Analyze: Short form declarations (name = calculation), implicit input derivation, name resolution with precedence

**[Story210_cpl-concepts-conditional-examples.md](Story210_cpl-concepts-conditional-examples.md)** — _Done_
: Create valid and invalid CPL example files demonstrating conditional concept parser and analyzer functionality

**[Story212_cpl-short-form-metrics-examples.md](Story212_cpl-short-form-metrics-examples.md)** — _Done_
: Add CPL example files exercising short form metric syntax across parse_and_analyze_only and with_data hierarchies

**[Story187_cpl-metrics-conditionals-parse-analyze.md](Story187_cpl-metrics-conditionals-parse-analyze.md)** — _Done_
: Increment 4 Parse+Analyze: Conditional expressions, comparison/logical operators, Nil/bool literals, branch child scopes, CILConditional/CILWhenClause assembly

**[Story188_cpl-metrics-conditionals-synthesize.md](Story188_cpl-metrics-conditionals-synthesize.md)** — _Done_
: Increment 4 Synthesis: Extend postfix evaluator with comparison/logical operators, conditional postfix op kind, bit-packed uint64 masks, and TensorWhereFloat64 for mask-based branch selection

**[Story189_cpl-metrics-aliases-where-parse-analyze.md](Story189_cpl-metrics-aliases-where-parse-analyze.md)** — _Done_
: Increment 5 Parse+Analyze: Input aliases (as clause) and where clauses (aspect=ACTUALS/FORECAST filtering)

**[Story213_cpl-consolidate-cil-tree-validation-walkers.md](Story213_cpl-consolidate-cil-tree-validation-walkers.md)** — _Done_
: Replace the three separate recursive tree-walking validation functions in metrics_analysis.go with a single multi-purpose CIL tree walk that performs all validation checks in one traversal

**[Story214_cpl-metrics-conditional-examples.md](Story214_cpl-metrics-conditional-examples.md)** — _Done_
: Create valid and invalid CPL example files demonstrating conditional expressions in metric calculate clauses

**[Story219_ods-generator-and-verifier.md](Story219_ods-generator-and-verifier.md)** — _Done_
: Implement ODS generator and verifier CLI tools for producing human-inspectable spreadsheets with formula rows and automated parquet verification.

**[Story220_tweak-expected-value-output-format.md](Story220_tweak-expected-value-output-format.md)** — _Done_
: Restructure the expected_output ODS sheet to a single-row-per-metric format with Computed Value and Match formula columns, removing CPL File, Branch, and Flag columns.

**[Story221_cpl-refactoring-phase1-cleanup.md](Story221_cpl-refactoring-phase1-cleanup.md)** — _Done_
: Remove dead code, deprecated fields, and hardcoded magic values from the CPL compiler codebase

**[Story222_cpl-refactoring-phase2-normalize-cil.md](Story222_cpl-refactoring-phase2-normalize-cil.md)** — _Done_
: Unify competing CIL representations and replace interface{} bridges with typed structures

**[Story208_fix-conditional-spec-pipeline-defects.md](Story208_fix-conditional-spec-pipeline-defects.md)** — _Done_
: Fix the four critical defects in the conditional spec end-to-end pipeline discovered during review of branch story121_and_122

**[Story209_fix-post-rebase-issues.md](Story209_fix-post-rebase-issues.md)** — _Done_
: Fix post-rebase issues: broken test paths, file extensions, story number collisions, deprecated code, missing examples, and README updates for conditional concept pipeline

**[Story217_cpl-metrics-synthesis-refactoring.md](Story217_cpl-metrics-synthesis-refactoring.md)** — _Done_
: Refactor large monolithic synthesis source files into focused, per-concern files to reduce context-window load for limited-context LLMs implementing conditional metrics synthesis

**[Story227_cpl-conditional-concept-values-synthesize.md](Story227_cpl-conditional-concept-values-synthesize.md)** — _Done_
: Enable CPL synthesis pipeline to handle conditional concept values in connector grain dimensions with dense tensors, validity masks, and conditional-aware parquet generation

**[Story228_cpl-expression-infrastructure.md](finished-stories/Story228_cpl-expression-infrastructure.md)** — _Done_
: Implement shared expression infrastructure used by compound units, conversions, and (future) metrics. Covers ANTLR grammar rules for expressions with operator precedence, postfix expression parsing, BuildExpressionTree, CIL expression types, and expression code synthesis.

**[Story224_cpl-refactoring-phase4-reduce-duplication.md](Story224_cpl-refactoring-phase4-reduce-duplication.md)** — _Done_
: Unify duplicated operator mappings, scalar detection, and constraint validation patterns

**[Story191_cpl-metrics-time-shift-parse-analyze.md](Story191_cpl-metrics-time-shift-parse-analyze.md)** — _Done_
: Increment 6 Parse+Analyze: backward()/forward() clauses with Instant constraint field keywords, self-reference semantics

**[Story273_fix-architecture-st-drift.md](Story273_fix-architecture-st-drift.md)** — _Done_
: Fix symbol table drift by registering scope-carrying definitions via DefineWithData, iterating Definitions in analysis, and propagating the pattern to aggregators for future spreader consistency

**[Story274_fix-conversion-issues.md](Story274_fix-conversion-issues.md)** — _Done_
: Address the five issues identified in the Story 230 review, then unify the duplicated unit-conversion and measure-conversion implementations to eliminate redundant code

**[Story194_cpl-metrics-dag-io-analyze.md](Story194_cpl-metrics-dag-io-analyze.md)** — _Done_
: Increment 8 Analysis: Dependency DAG construction, topological sort, cycle detection, dead-code elimination, full IO spec validation

**[Story195_cpl-metrics-dag-io-synthesize.md](Story195_cpl-metrics-dag-io-synthesize.md)** — _Done_
: Increment 8 Synthesis: Multi-metric executables with topological evaluation order, consumer counts, dead-code elimination, output parquet writing

**[Story223_cpl-refactoring-phase3-split-bloat.md](Story223_cpl-refactoring-phase3-split-bloat.md)** — _Done_
: Decompose CPLAnalyzer and CILAbstractTransformer into focused single-responsibility sub-components

**[Story225_cpl-refactoring-phase5-address-fragility.md](Story225_cpl-refactoring-phase5-address-fragility.md)** — _Done_
: Restructure fragile patterns: template-based codegen, conditional parsing state machine, visitor removal, stable analysis→synthesis contract

**[Story226_cpl-refactoring-phase6-test-hardening.md](Story226_cpl-refactoring-phase6-test-hardening.md)** — _Done_
: Add unit tests for untested areas, fix hardcoded defaults, implicit contracts, and generated error handling

**[Story229_cpl-compound-units.md](Story229_cpl-compound-units.md)** — _Done_
: Wire expression infrastructure into measure definitions so that measure unit specifications support compound unit expressions (e.g., Length/Time^2). Replace ConstraintScope-based unit extraction with expression-based parsing.

**[Story230_cpl-conversions.md](Story230_cpl-conversions.md)** — _Done_
: Implement unit and measure conversions using pattern/calculation pairs. Covers conversion clause grammar, ConversionScope/PatternScope/CalculationScope symbol table scopes, conversion pair assembly, CIL conversion types, and conversion code synthesis. Simple (non-conditional) conversions only.

**[Story234_fix-conditional-synthesis-defects.md](Story234_fix-conditional-synthesis-defects.md)** — _Done_
: Fix defects in the conditional concept synthesis pipeline: parent ambiguity in mask computation, missing e2e test, example convention violations, and test coverage gaps

**[Story237_create-xlsx-conversion-script.md](Story237_create-xlsx-conversion-script.md)** — _Done_
: Create the reusable PDD script for per-example XLSX conversion and convert the test driver to use xls_verify instead of parquet_verify

**[Story238_convert-arithmetic-examples-to-xlsx.md](Story238_convert-arithmetic-examples-to-xlsx.md)** — _Done_
: Convert all 11 currently-passing with_data examples from JSON/ODS verification to XLSX verification

**[Story239_convert-conditional-grain-examples-to-xlsx.md](Story239_convert-conditional-grain-examples-to-xlsx.md)** — _Done_
: Convert all 7 currently-skipped with_data examples (6 conditionals + 1 grain) from JSON/ODS verification to XLSX verification

**[Story240_delete-ods-code-and-resources.md](Story240_delete-ods-code-and-resources.md)** — _Done_
: Delete all ODS-related CLI code, tests, utilities, and API documentation once all examples are converted to XLSX

**[Story280_add-lookup-expectations.md](Story280_add-lookup-expectations.md)** — _Done_
: Fix expected_xls_generator to produce valid VLOOKUP formulas for CILLookupNode expressions in the Computed Value column of expected_output sheets

**[Story241_lookups-increment1-parse.md](Story241_lookups-increment1-parse.md)** — _Done_
: Implement CPLLookups.g4 grammar, LookupsListener, ParserLookupDefinition, lexer tokens, symbol table storage for timeless lookups

**[Story242_lookups-increment1-analyze.md](Story242_lookups-increment1-analyze.md)** — _Done_
: Implement AnalyzeLookupScope, LookupInfo, CILLookupNode, cross-reference resolution, name uniqueness validation for timeless lookups

**[Story254_aggregators-increment1-matrix-mult.md](Story254_aggregators-increment1-matrix-mult.md)** — _Done_
: Implement tensor matrix multiplication infrastructure: ReduceDimension, ExpandDimension, composeIndex/decomposeIndex helpers, unit tests with hand-constructed tensors

**[Story255_aggregators-increment2-parse.md](Story255_aggregators-increment2-parse.md)** — _Done_
: Implement CPLAggregators.g4 grammar, cpl_aggregators_listener.go, ParserAggregatorDefinition, aggregator lexer tokens, symbol table storage for aggregators

**[Story277_with_data-restructure-and-unit-checking-examples.md](Story277_with_data-restructure-and-unit-checking-examples.md)** — _Done_
: Reorganize examples/with_data into without_errors/ and with_errors/ subdirectories, move existing examples, update test discovery, and create end-to-end and error-case examples exercising Story231's synthesis-time unit type checking

**[Story278_aggregators-increment2-analyze.md](Story278_aggregators-increment2-analyze.md)** — _Done_
: Implement AnalyzeAggregatorScope, AggregatorInfo, CILAggregateNode, cross-reference resolution, name uniqueness validation for aggregators

**[Story256_aggregators-increment2-analyze.md](Story256_aggregators-increment2-analyze.md)** — _Done_
: Implement AnalyzeAggregatorScope, AggregatorInfo, CILAggregateNode, cross-reference resolution, name uniqueness validation for aggregators

**[Story264_spreaders-increment1-parse.md](Story264_spreaders-increment1-parse.md)** — _Done_
: Implement CPLSpreaders.g4 grammar, SpreadersListener, ParserSpreaderDefinition, synthetic key assignment, postfix expression storage, lexer tokens

**[Story257_aggregators-increment2-synthesize.md](Story257_aggregators-increment2-synthesize.md)** — _Done_
: Implement CIL tree walking for CILAggregateNode, aggregation matrix Go code emission in synthesis preamble, manual sum aggregation test, and end-to-end aggregator examples

**[Story265_spreaders-increment1-analyze.md](Story265_spreaders-increment1-analyze.md)** — _Done_
: Implement AnalyzeSpreaderScope, SpreaderInfo, SpreaderFunctionRegistry, postfix expression evaluation, dimensional unit validation, CILSpreadNode, SpreadingMatrix construction, parse_and_analyze_only examples

(End of file - total 407 lines)
