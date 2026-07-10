# Stories Table of Contents

This file contains a table of active stories for the project. Stories that are pending or currently in progress are listed below.

**For records of canceled stories, see [`canceled-stories/toc.md`](canceled-stories/toc.md). For records of finished stories, see [`finished-stories/toc.md`](finished-stories/toc.md). Consult these records only when information about canceled or finished stories is required.**

**[Story275_iospec-connectors-and-lookups.md](finished-stories/Story275_iospec-connectors-and-lookups.md)** — _Done_
: Rename iospec `inputs` → `connectors`, add `lookups` section, fix lookups_basic example data, extend expected_xls_generator for CILLookupNode

## Active Stories

Stories that are pending or currently in progress.

**[Story008_docker-demo-container.md](Story008_docker-demo-container.md)** — _Not Started_
: Create Docker demo container with PostgreSQL, autostarted Core Service, web-served frontend, and JSON-based definition initialization

**[Story014_sql-documentation-and-er-diagram.md](Story014_sql-documentation-and-er-diagram.md)** — _Not Started_
: Add comprehensive comments to SQL schema files explaining each attribute's purpose and relationships, and generate an ER diagram for architecture docs

**[Story015_timeseries-parser-go.md](Story015_timeseries-parser-go.md)** — _Not Started_
: Implement Go parser/formatter for reading/writing time series object data in Parquet format with multi-level dictionary in-memory representation

**[Story016_intellij-workspace-setup.md](Story016_intellij-workspace-setup.md)** — _Not Started_
: Set up IntelliJ IDEA plugin workspace with Gradle build, plugin configuration, and development environment for Kotlin-based plugin

**[Story017_timeseries-parser-kotlin.md](Story017_timeseries-parser-kotlin.md)** — _Not Started_
: Implement Kotlin parser/formatter for time series object data using ParquetKt library for IntelliJ plugin

**[Story018_vscode-workspace-setup.md](Story018_vscode-workspace-setup.md)** — _Not Started_
: Set up VSCode extension workspace with npm/webpack build, TypeScript configuration, and development environment for TypeScript-based extension

**[Story019_timeseries-parser-typescript.md](Story019_timeseries-parser-typescript.md)** — _Not Started_
: Implement TypeScript parser/formatter for time series object data for VSCode plugin

**[Story025_identity-authentication-authorization-requirements.md](Story025_identity-authentication-authorization-requirements.md)** — _Not Started_
: Explore and define requirements for user identity management, authentication, and authorization

**[Story026_drivers-requirements.md](Story026_drivers-requirements.md)** — _Not Started_
: Explore and define requirements for storage drivers to abstract storage backends

**[Story031_metadata-requirements.md](Story031_metadata-requirements.md)** — _Not Started_
: Explore and define requirements for metadata management

**[Story034_web-ui-requirements.md](Story034_web-ui-requirements.md)** — _Not Started_
: Explore and define requirements for the browser-based interface for interacting with the system

**[Story035_plugin-ui-requirements.md](Story035_plugin-ui-requirements.md)** — _Not Started_
: Explore and define requirements for IDE-integrated interfaces and capabilities

**[Story036_data-and-plan-definition-lineage-requirements.md](Story036_data-and-plan-definition-lineage-requirements.md)** — _Not Started_
: Explore and define requirements for tracking data lineage and plan definition lineage

**[Story037_agents-and-agentic-interaction-requirements.md](Story037_agents-and-agentic-interaction-requirements.md)** — _Not Started_
: Explore and define requirements for AI agent integration and autonomous planning capabilities

**[Story038_mcp-servers-requirements.md](Story038_mcp-servers-requirements.md)** — _Not Started_
: Explore and define requirements for Model Context Protocol server implementations for LLM integration

**[Story039_external-api-simulator-and-sample-data-requirements.md](Story039_external-api-simulator-and-sample-data-requirements.md)** — _Not Started_
: Explore and define requirements for the external API simulator and sample data for demonstrations

**[Story043L_exemplar_aggregate_entity_gaps.md](Story043L_exemplar_aggregate_entity_gaps.md)** — _Not Started_
: Fill the aggregate-root scaffold gaps for namespace family, namespace, category, and element exemplars

**[Story043M_cpl_schema_alignment_review.md](Story043M_cpl_schema_alignment_review.md)** — _Not Started_
: Review ChronoCone Planning Language drift against the database schema and align schema/model generation where needed

**[Story043N_extend_aggregate_pattern.md](Story043N_extend_aggregate_pattern.md)** — _Not Started_
: Copy the verified aggregate-root pattern to all other similarly shaped entities with clear edge ownership

**[Story044_localization-services.md](Story044_localization-services.md)** — _Not Started_
: Cross-cutting localization services for all entity \_localization tables

**[Story045_policy-services.md](Story045_policy-services.md)** — _Not Started_
: Policy domain services: namespace families, namespaces, subject gateway, cross-namespace promotion

**[Story046_taxonomy-services.md](Story046_taxonomy-services.md)** — _Not Started_
: Taxonomy domain services: categories, elements, hierarchies, batch validation, cascade analysis

**[Story047_taxonomy-extension-services.md](Story047_taxonomy-extension-services.md)** — _Not Started_
: Taxonomy extension services: category_attr and element_attr cross-namespace attributes

**[Story048_plan-type-services.md](Story048_plan-type-services.md)** — _Not Started_
: Plan type services: metrics, metric dimensions, ts_doc_type, ts_doc_row_type

**[Story049_plan-storage-services.md](Story049_plan-storage-services.md)** — _Not Started_
: Plan storage services: ts_doc management and blob storage coordination

**[Story050_connector-services.md](Story050_connector-services.md)** — _Not Started_
: Connector services: language/locale definitions and storage configuration

**[Story051_cross-taxonomy-mapping.md](Story051_cross-taxonomy-mapping.md)** — _Not Started_
: Cross-taxonomy mapping: import and reorganize elements from outside taxonomies (deferred past demo)

**[Story052_morphism-services.md](Story052_morphism-services.md)** — _Not Started_
: Morphism services: transformer, shifter, filter (new), lookup (new), DAG validation

**[Story100_namespace-type-ui.md](Story100_namespace-type-ui.md)** — _Not Started_
: Add namespace type selection to the UI when creating namespaces and display namespace type in the hierarchy view

**[Story186_cpl-metrics-parallelism-synthesize.md](Story186_cpl-metrics-parallelism-synthesize.md)** — _Not Started_
: Increment 3 Synthesis: Configure ChronoconeComputationController from IO spec, worker pools, chunk size optimization

**[Story190_cpl-metrics-aliases-where-synthesize.md](Story190_cpl-metrics-aliases-where-synthesize.md)** — _Not Started_
: Increment 5 Synthesis: Alias-based variable naming, aspect-filtering slice selection on input tensors

**[Story192_cpl-metrics-time-shift-synthesize.md](Story192_cpl-metrics-time-shift-synthesize.md)** — _Not Started_
: Increment 6 Synthesis: VisitCILTimeShift for array index offset calculations, boundary condition handling, self-referencing topological ordering

**[Story215_cpl-metrics-parallel-branch-execution.md](Story215_cpl-metrics-parallel-branch-execution.md)** — _Not Started_
: Convert sequential branch evaluation in conditional postfix evaluator to parallel goroutine dispatch with mask-aware computation skipping

**[Story216_cpl-metrics-boolean-body-expressions.md](Story216_cpl-metrics-boolean-body-expressions.md)** — _Not Started_
: Extend postfix evaluator and type system to support boolean/logical expressions in conditional branch body positions

**[Story218_transition-to-ods-verification.md](Story218_transition-to-ods-verification.md)** — _Not Started_
: With generator and verifier built (Story219), transition all 17 with_data examples to ODS verification: correct JSONs, generate .ods files, create PDD script, validate full test suite.

**[Story226_cpl-refactoring-phase6-test-hardening.md](Story226_cpl-refactoring-phase6-test-hardening.md)** — _Not Started_
: Add unit tests for untested areas, fix hardcoded defaults, implicit contracts, and generated error handling

**[Story229_cpl-compound-units.md](finished-stories/Story229_cpl-compound-units.md)** — _Done_
: Wire expression infrastructure into measure definitions so that measure unit specifications support compound unit expressions (e.g., Length/Time^2). Replace ConstraintScope-based unit extraction with expression-based parsing.

**[Story231_cpl-unit-type-checking.md](Story231_cpl-unit-type-checking.md)** — _In Progress (Task 4g, 3e, 5f incomplete)_
: Implement synthesis-time unit type verification across the full CIL expression tree. Walk CIL expression nodes for metric formulas, track unit types through operations, validate conversion outputs match required units, and verify final calculation results match IO spec output units.
**[Story232_cpl-conversion-injection.md](Story232_cpl-conversion-injection.md)** — _Not Started_
: Implement automatic conversion injection into CIL expression trees. When the unit type checker identifies unit mismatches and a valid conversion pair exists, inject CILConversionApplication nodes so the synthesizer produces code that performs conversions at runtime.

**[Story239_convert-conditional-grain-examples-to-xlsx.md](finished-stories/Story239_convert-conditional-grain-examples-to-xlsx.md)** — _Done_
: Convert all 7 currently-skipped with_data examples (6 conditionals + 1 grain) from JSON/ODS verification to XLSX verification

**[Story242_lookups-increment1-analyze.md](finished-stories/Story242_lookups-increment1-analyze.md)** — _Done_
: Implement AnalyzeLookupScope, LookupInfo, CILLookupNode, cross-reference resolution, name uniqueness validation for timeless lookups

**[Story244_lookups-increment2-parse.md](Story244_lookups-increment2-parse.md)** — _Not Started_
: Extend grammar and parser for grain support in lookup declarations: with_grain keyword and compound key syntax

**[Story245_lookups-increment2-analyze.md](Story245_lookups-increment2-analyze.md)** — _Not Started_
: Extend analysis for grain-bearing lookups: Grain field validation, grain measure reference resolution, multi-key LookupInfo structure

**[Story246_lookups-increment2-synthesize.md](Story246_lookups-increment2-synthesize.md)** — _Not Started_
: Extend synthesis for grain-bearing lookups: CILLookupNode with grain compound key, tensor loader with grain iteration, value collection by key

**[Story247_lookups-increment3-parse.md](Story247_lookups-increment3-parse.md)** — _Not Started_
: Extend grammar and parser for lookup input kind in metric calculate clauses: lookup() reference syntax and lexer token

**[Story248_lookups-increment3-analyze.md](Story248_lookups-increment3-analyze.md)** — _Not Started_
: Extend analysis for lookup input references: validate lookup name in metric input, determine shared vs independent grain model

**[Story249_lookups-increment3-synthesize.md](Story249_lookups-increment3-synthesize.md)** — _Not Started_
: Extend synthesis for lookup input references: CILLookupRefNode, lookup evaluator codegen, lookup value consumption in postfix evaluator

**[Story250_lookups-increment4-analyze.md](Story250_lookups-increment4-analyze.md)** — _Not Started_
: Implement grain validation analysis: verify metric input grain matches lookup grain composition for lookup-scoped inputs

**[Story251_lookups-increment4-synthesize.md](Story251_lookups-increment4-synthesize.md)** — _Not Started_
: Extend synthesis for aggregator-lookup combinations: lookup reference resolution in aggregator CIL nodes, tensor index mapping with lookup dimension awareness, code generation for runtime lookup+aggregation integration

**[Story252_lookups-increment5-analyze.md](Story252_lookups-increment5-analyze.md)** — _Not Started_
: Implement runtime validation analysis for lookups: lookup-to-IO-spec registration, output dimension derivation from input metrics, lookup celltype unit validation against required output units

**[Story253_lookups-increment5-synthesize.md](Story253_lookups-increment5-synthesize.md)** — _Not Started_
: Extend synthesis for lookup IO spec integration: register lookups as MetricComputation objects, generate output parquet writing for lookup cell values, wire lookup compute functions into main execution flow

**[Story258_aggregators-increment3-analyze.md](Story258_aggregators-increment3-analyze.md)** — _Not Started_
: Implement aggregator input reference resolution in metric calculate clauses, CILAggregatorRefNode, dimensional compatibility validation for aggregator inputs

**[Story259_aggregators-increment3-synthesize.md](Story259_aggregators-increment3-synthesize.md)** — _Not Started_
: Implement CIL tree walker for all function CIL node types (sum, count, size, 1/sum, 1/count), size(C) constant emission, expression-integration examples

**[Story260_aggregators-increment4-analyze.md](Story260_aggregators-increment4-analyze.md)** — _Not Started_
: Implement aggregator lookup resolution in aggregator expressions, validate output unit compatibility for aggregator-lookup chains, aggregator output dimension derivation

**[Story261_aggregators-increment4-synthesize.md](Story261_aggregators-increment4-synthesize.md)** — _Not Started_
: Extend synthesis for aggregator+lookup chains: generate runtime aggregation matrix calls with lookup tensor inputs, unit-correct aggregation codegen, output dimension adjustment in preamble

**[Story262_aggregators-increment5-analyze.md](Story262_aggregators-increment5-analyze.md)** — _Not Started_
: Implement aggregator spreader resolution in aggregator expressions, validate output unit compatibility for aggregator-spreader chains, detect spreader pre-aggregation patterns

**[Story263_aggregators-increment5-synthesize.md](Story263_aggregators-increment5-synthesize.md)** — _Not Started_
: Extend synthesis for aggregator+spreader chains: generate pre-collection aggregation with spreader distribution, runtime spreader key construction from aggregation output, codegen wiring


**[Story265_spreaders-increment1-analyze.md](finished-stories/Story265_spreaders-increment1-analyze.md)** — _Done_
: Implement AnalyzeSpreaderScope, SpreaderInfo, CILSpreaderNode, cross-reference resolution, name uniqueness validation for spreaders

**[Story266_spreaders-increment1-synthesize.md](Story266_spreaders-increment1-synthesize.md)** — _Not Started_
: Implement CILAbstractTransformer VisitCILTreeNode for CILSpreaderNode, synthesis of spreader emission, spreader codegen for preamble functions

**[Story267_spreaders-increment2-analyze.md](Story267_spreaders-increment2-analyze.md)** — _Not Started_
: Implement spreader input reference resolution in metric calculate clauses, CILSpreaderRefNode, dimensional compatibility validation for spreader inputs

**[Story268_spreaders-increment2-synthesize.md](Story268_spreaders-increment2-synthesize.md)** — _Not Started_
: Extend synthesis for spreader input references: CILSpreaderRefNode codegen, tensor dimension tracking through spreader ops, runtime spreader+input integration

**[Story269_spreaders-increment3-analyze.md](Story269_spreaders-increment3-analyze.md)** — _Not Started_
: Implement spreader aggregator resolution in spreader expressions, validate output unit compatibility for spreader-aggregator chains, detect post-spreader aggregation patterns

**[Story270_spreaders-increment3-synthesize.md](Story270_spreaders-increment3-synthesize.md)** — _Not Started_
: Extend synthesis for spreader+aggregator chains: generate post-distribution aggregation with spreader tensor inputs, runtime aggregation integration, codegen wiring

**[Story271_spreaders-increment4-analyze.md](Story271_spreaders-increment4-analyze.md)** — _Not Started_
: Implement spreader output dimension derivation, spreader output unit validation for spreader-lookup chains, detect spreader-to-lookup dimension expansion patterns

**[Story272_spreaders-increment4-synthesize.md](Story272_spreaders-increment4-synthesize.md)** — _Not Started_
: Extend synthesis for spreader+lookup chains: generate spreader tensor code for lookup dimension expansion, runtime lookup integration with spreader outputs, codegen wiring

**[Story279_with_data-conversion-injection-example.md](Story279_with_data-conversion-injection-example.md)** — _Not Started_
: Create a with_data/without_errors example verifying Story232's conversion injection: a metric computing EUR with IO spec requiring USD and an EUR→USD conversion pair, demonstrating that injection produces working runtime code with correct numerical results

## Finished Stories

Stories that have been completed.

**[Story273_fix-architecture-st-drift.md](finished-stories/Story273_fix-architecture-st-drift.md)** — _Done_
: Fix symbol table drift by registering scope-carrying definitions via DefineWithData, iterating Definitions in analysis, and propagating the pattern to aggregators for future spreader consistency

**[Story274_fix-conversion-issues.md](finished-stories/Story274_fix-conversion-issues.md)** — _Done_
: Address the five issues identified in the Story 230 review, then unify the duplicated unit-conversion and measure-conversion implementations to eliminate redundant code

**[Story230_cpl-conversions.md](finished-stories/Story230_cpl-conversions.md)** — _Done_
: Implement unit and measure conversions using pattern/calculation pairs. Covers conversion clause grammar, ConversionScope/PatternScope/CalculationScope symbol table scopes, conversion pair assembly, CIL conversion types, and conversion code synthesis. Simple (non-conditional) conversions only.

**[Story255_aggregators-increment2-parse.md](finished-stories/Story255_aggregators-increment2-parse.md)** — _Done_
: Implement CPLAggregators.g4 grammar, cpl_aggregators_listener.go, ParserAggregatorDefinition, aggregator lexer tokens, symbol table storage for aggregators

**[Story276_adjust-lookups-try2.md](finished-stories/Story276_adjust-lookups-try2.md)** — _Done_
: Fix lookups end-to-end pipeline: add CILLookupNode to postfix walker, implement tensor broadcasting for binary operations, fix lookups_basic example data

**[Story243_lookups-increment1-synthesize.md](finished-stories/Story243_lookups-increment1-synthesize.md)** — _Done_
: Implement CILAbstractTransformer VisitCILTreeNode for CILLookupNode, synthesis of LookupTensorLoader lookup loader and computeX functions

**[Story241_lookups-increment1-parse.md](finished-stories/Story241_lookups-increment1-parse.md)** — _Done_
: Implement CPLLookups.g4 grammar, LookupsListener, ParserLookupDefinition, lexer tokens, symbol table storage for timeless lookups

**[Story254_aggregators-increment1-matrix-mult.md](finished-stories/Story254_aggregators-increment1-matrix-mult.md)** — _Done_
: Implement tensor matrix multiplication infrastructure: ReduceDimension, ExpandDimension, composeIndex/decomposeIndex helpers, unit tests with hand-constructed tensors

**[Story191_cpl-metrics-time-shift-parse-analyze.md](finished-stories/Story191_cpl-metrics-time-shift-parse-analyze.md)** — _Done_
: Increment 6 Parse+Analyze: backward()/forward() clauses with Instant constraint field keywords, self-reference semantics

**[Story194_cpl-metrics-dag-io-analyze.md](finished-stories/Story194_cpl-metrics-dag-io-analyze.md)** — _Done_
: Increment 8 Analysis: Dependency DAG construction, topological sort, cycle detection, dead-code elimination, full IO spec validation

**[Story226_cpl-refactoring-phase6-test-hardening.md](finished-stories/Story226_cpl-refactoring-phase6-test-hardening.md)** — _Done_
: Add unit tests for untested areas, fix hardcoded defaults, implicit contracts, and generated error handling

**[Story240_delete-ods-code-and-resources.md](finished-stories/Story240_delete-ods-code-and-resources.md)** — _Done_
: Delete all ODS-related CLI code, tests, utilities, and API documentation once all examples are converted to XLSX

**[Story238_convert-arithmetic-examples-to-xlsx.md](finished-stories/Story238_convert-arithmetic-examples-to-xlsx.md)** — _Done_
: Convert all 11 currently-passing with_data examples from JSON/ODS verification to XLSX verification

**[Story237_create-xlsx-conversion-script.md](finished-stories/Story237_create-xlsx-conversion-script.md)** — _Done_
: Create the reusable PDD script for per-example XLSX conversion and convert the test driver to use xls_verify instead of parquet_verify

**[Story234_fix-conditional-synthesis-defects.md](finished-stories/Story234_fix-conditional-synthesis-defects.md)** — _Done_
: Fix defects in the conditional concept synthesis pipeline: parent ambiguity in mask computation, missing e2e test, example convention violations, and test coverage gaps

**[Story227_cpl-conditional-concept-values-synthesize.md](finished-stories/Story227_cpl-conditional-concept-values-synthesize.md)** — _Done_
: Enable CPL synthesis pipeline to handle conditional concept values in connector grain dimensions with dense tensors, validity masks, and conditional-aware parquet generation

**[Story195_cpl-metrics-dag-io-synthesize.md](finished-stories/Story195_cpl-metrics-dag-io-synthesize.md)** — _Done_
: Increment 8 Synthesis: Multi-metric executables with topological evaluation order, consumer counts, dead-code elimination, output parquet writing

**[Story225_cpl-refactoring-phase5-address-fragility.md](finished-stories/Story225_cpl-refactoring-phase5-address-fragility.md)** — _Done_
: Restructure fragile patterns: template-based codegen, conditional parsing state machine, visitor removal, stable analysis→synthesis contract

**[Story223_cpl-refactoring-phase3-split-bloat.md](finished-stories/Story223_cpl-refactoring-phase3-split-bloat.md)** — _Done_
: Decompose CPLAnalyzer and CILAbstractTransformer into focused single-responsibility sub-components

**[Story229_cpl-compound-units.md](finished-stories/Story229_cpl-compound-units.md)** — _Done_
: Wire expression infrastructure into measure definitions so that measure unit specifications support compound unit expressions (e.g., Length/Time^2). Replace ConstraintScope-based unit extraction with expression-based parsing.

**[Story239_convert-conditional-grain-examples-to-xlsx.md](finished-stories/Story239_convert-conditional-grain-examples-to-xlsx.md)** — _Done_
: Convert all 7 currently-skipped with_data examples (6 conditionals + 1 grain) from JSON/ODS verification to XLSX verification

**[Story242_lookups-increment1-analyze.md](finished-stories/Story242_lookups-increment1-analyze.md)** — _Done_
: Implement AnalyzeLookupScope, LookupInfo, CILLookupNode, cross-reference resolution, name uniqueness validation for timeless lookups


