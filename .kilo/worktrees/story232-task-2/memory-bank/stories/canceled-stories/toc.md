# Canceled Stories Table of Contents

This file contains a table of all canceled stories for the project. These stories have been canceled or superseded.

**[Story010_metadata-update-delete.md](Story010_metadata-update-delete.md)** — _Canceled_
: Add update and delete functionality for metadata entities to the Metrics Service API and Definitions page UI

**[Story012B_metrics-service-compliance.md](Story012B_metrics-service-compliance.md)** — _Canceled_
: Refactor Metrics Service for architecture compliance and upgrade endpoints to be fully compliant with OpenAPI Specification 3.2.0

**[Story013_storage-interface.md](Story013_storage-interface.md)** — _Canceled_
: Implement singleton StorageInterface class with driver registry loaded from storage_config table, dispatching storage operations to appropriate drivers

**[Story027_relational-store-requirements.md](Story027_relational-store-requirements.md)** — _Canceled_
: Explore and define requirements for relational database usage for metadata storage

**[Story029_object-store-requirements.md](Story029_object-store-requirements.md)** — _Canceled_
: Explore and define requirements for object storage usage for timeseries data

**[Story030_versioning-requirements.md](Story030_versioning-requirements.md)** — _Canceled_
: Explore and define requirements for versioning of plans, metrics, and system artifacts

**[Story032_metrics-and-calculation-requirements.md](Story032_metrics-and-calculation-requirements.md)** — _Canceled_
: Explore and define requirements for how metrics are defined, calculated, and related to one another

**[Story033_connectors-requirements.md](Story033_connectors-requirements.md)** — _Canceled_
: Explore and define requirements for connectors for external system integration

**[Story042_sql-preprocessor.md](Story042_sql-preprocessor.md)** — _Canceled_
: ~~Implement lightweight SQL preprocessor for file-based SQL templates with conditional blocks and parameterized queries~~ **SUPERSEDED** by Go-Jet

**[Story106_cpl-parser-metrics.md](Story106_cpl-parser-metrics.md)** — _Canceled_
: Extend CPL parser to support metric and aggregator declarations with expressions

**[Story107_cpl-metric-data-structure-design.md](Story107_cpl-metric-data-structure-design.md)** — _Canceled_
: Design in-memory metric data structure with database and Parquet translation

**[Story108_cpl-metric-evaluation.md](Story108_cpl-metric-evaluation.md)** — _Canceled_
: Implement metric evaluation for formulas and boolean logic with CLI output

**[Story133_cpl-types-parser-string-boolean.md](Story133_cpl-types-parser-string-boolean.md)** — _Superseded by Story196_
: Add String and Boolean base types, String constraints (min/max length, pattern), Boolean constraints (none allowed)

**[Story134_cpl-types-analyzer-string-boolean.md](Story134_cpl-types-analyzer-string-boolean.md)** — _Superseded by Story196_
: Implement StringConstraints with regex compilation, BooleanConstraints, enhanced CreateConstraints for String and Boolean

**[Story145_cpl-compound-units-parser.md](Story145_cpl-compound-units-parser.md)** — _Superseded_
: CPL Compound Units Parser — Unit algebra expression parsing with postfix sequence. Replaced by Story228 and Story229.

**[Story146_cpl-compound-units-analyzer.md](Story146_cpl-compound-units-analyzer.md)** — _Superseded_
: CPL Compound Units Analyzer — BuildExpressionTree and unit algebra resolution. Replaced by Story228 and Story229.

**[Story147_cpl-unit-conversions-parser.md](Story147_cpl-unit-conversions-parser.md)** — _Superseded_
: CPL Unit Conversions Parser — Pattern/calculation pairs, ConversionScope grammar. Replaced by Story228 and Story230.

**[Story148_cpl-unit-conversions-analyzer.md](Story148_cpl-unit-conversions-analyzer.md)** — _Superseded_
: CPL Unit Conversions Analyzer — Conversion pair assembly and unit validation. Replaced by Story230.

**[Story149_cpl-measure-conversions-parser.md](Story149_cpl-measure-conversions-parser.md)** — _Superseded_
: CPL Measure Conversions Parser — Measure conversion declarations. Replaced by Story230.

**[Story150_cpl-measure-conversions-analyzer.md](Story150_cpl-measure-conversions-analyzer.md)** — _Superseded_
: CPL Measure Conversions Analyzer — Conversion pair resolution for measures. Replaced by Story230.

**[Story218_transition-to-ods-verification.md](Story218_transition-to-ods-verification.md)** — _Superseded_
: With generator and verifier built (Story219), transition all 17 with_data examples to ODS verification: correct JSONs, generate .ods files, create PDD script, validate full test suite. Superseded by Story237–Story240 (XLSX-based verification).
