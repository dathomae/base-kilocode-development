# src/ Directory Guide

> Progressive-discovery map for agents. Each entry is one line so diffs stay
> merge-friendly. When a directory grows a new major subdirectory, append a
> bullet here in the same section; the flat list keeps conflicts trivial.

---

## Top-level directories under `src/`

| Directory                       | What it is                                 | Keywords / Service names                                                             |
| ------------------------------- | ------------------------------------------ | ------------------------------------------------------------------------------------ |
| `backend/`                      | Go-native deterministic calculation engine | REST API, domain services, bitemporal persistence, EBM, handlers, Go-Jet, PostgreSQL |
| `chronocone_planning_language/` | CPL parser, analyzer, synthesizer          | ANTLR4, grammar, AST, CIL, codegen, tensor, Parquet, connector                       |
| `frontend/`                     | React/TypeScript UI for Timecone.ai        | Vite, React, metadata tables, taxonomy, planning views, mocks                        |

---

## `backend/` — Go backend services

| Subdirectory | What it holds                           | Keywords                                                                                          |
| ------------ | --------------------------------------- | ------------------------------------------------------------------------------------------------- |
| `cmd/`       | CLI entry points                        | `dbtool` (schema management), `generator` (codegen)                                               |
| `compute/`   | Distributed execution fabrics           | Beam (Flink / Spark), Kubernetes                                                                  |
| `constants/` | Cross-layer sentinel values             | `InfinityTime`, zero-dependency shared constants                                                  |
| `domain/`    | Core business logic / "brain"           | Namespace families, namespaces, categories, elements, taxonomy, timeseries, Unified Model, Go-Jet |
| `handlers/`  | HTTP/REST translation layer             | Huma, OpenAPI, thin delegators to domain                                                          |
| `security/`  | AuthN / AuthZ / external connectors     | Authentication, authorization, ReBAR/Zanzibar, blob, database connections                         |
| `serde/`     | Serialization & persistence abstraction | Arrow, JSON Schema, MCP, OpenAPI, RDBMS                                                           |

### `backend/serde/` backends

| Subdirectory         | What it holds                       | Keywords                                                                      |
| -------------------- | ----------------------------------- | ----------------------------------------------------------------------------- |
| `serde/arrow/`       | Apache Arrow columnar serialization | Inter-program communication                                                   |
| `serde/json_schema/` | JSON schema validation              | Foreign namespace assertion validation                                        |
| `serde/mcp/`         | MCP protocol descriptors            | API contracts                                                                 |
| `serde/openapi/`     | OpenAPI spec & response formatting  | Documentation                                                                 |
| `serde/rdbms/`       | PostgreSQL persistence + bitemporal | EBM, Flush/Purge/Assert, Go-Jet generated models, advisory locking, staleness |

---

## `chronocone_planning_language/` — CPL language toolchain

> Owned by the kilocode process; agents should not edit code here directly.
> See `src/chronocone_planning_language/toc.md` for full file-level detail.

| Subdirectory               | What it holds                                 | Keywords                                                                   |
| -------------------------- | --------------------------------------------- | -------------------------------------------------------------------------- |
| `examples/`                | Sample CPL files, test data, expected outputs | `.cpl`, `.json`, `.parquet`, `iospec`, error cases, metrics arithmetic     |
| `src/main/antlr/`          | ANTLR4 grammar files                          | `.g4`, lexer, concepts, types, units, measures, connectors, metrics        |
| `src/main/go/analysis/`    | Semantic analyzer & CIL generation            | Symbol table, type checking, concept analysis, metrics analysis, CIL nodes |
| `src/main/go/ast/`         | Abstract Syntax Tree definitions              | CPL document, type declarations, concept declarations                      |
| `src/main/go/calculation/` | Tensor computation runtime                    | `ChronoconeTensor`, postfix ops, arithmetic                                |
| `src/main/go/cil/`         | Chronocone Intermediate Language              | CIL transformer, code generation intermediate representation               |
| `src/main/go/cli/`         | CLI tools                                     | `compiler`, `parquetgen`, `parquet_verify`                                 |
| `src/main/go/connector/`   | Connector metadata & loader                   | Connector definitions, loader generation                                   |
| `src/main/go/io/`          | Parquet I/O & tensor loading                  | Parquet reader/writer, `iospec` handling                                   |
| `src/main/go/parser/`      | Parser implementation                         | ANTLR listeners, parser wrapper, symbol-table integration                  |
| `src/main/go/symboltable/` | Symbol table stack                            | Scope management, name resolution                                          |
| `src/main/go/synthesis/`   | CPL → Go synthesizer                          | Generates executable Go programs from parsed CPL                           |
| `shell_scripts/`           | Build & run helpers                           | `compile_and_run.sh`, `regenerate_data.sh`                                 |
| `standard_definitions/`    | Built-in CPL definitions                      | Standard concepts, units, measures                                         |

---

## `frontend/` — React/TypeScript UI

| Subdirectory      | What it holds                   | Keywords                                                                                            |
| ----------------- | ------------------------------- | --------------------------------------------------------------------------------------------------- |
| `mocks/`          | **Isolated UI mock prototypes** | Tabulator, modals, views, mock data, tests (separate from production)                               |
| `public/`         | Static assets                   | Favicon                                                                                             |
| `src/components/` | Production React components     | `ErrorMessage`, `LoadingSpinner`, `MetadataTable`, `MetadataTree`, `NavigationBar`, `TestComponent` |
| `src/pages/`      | Route-level pages               | `Home`, `Definitions`, `Metrics`                                                                    |
| `src/services/`   | API service layer               | `metricsApi`, backend HTTP client                                                                   |

### `frontend/mocks/` structure (isolated prototype)

| Subdirectory                              | What it holds     | Keywords                                                                                       |
| ----------------------------------------- | ----------------- | ---------------------------------------------------------------------------------------------- |
| `mocks/src/components/Layout/`            | Layout chrome     | `Header`, `Footer`, `Sidebar`                                                                  |
| `mocks/src/components/Modals/`            | Dialog components | `CategoryModal`, `TaxonomyModal`, `MetricModal`, `MeasureModal`, `ConverterModal`, `PlanModal` |
| `mocks/src/components/Views/Calculation/` | Calculation UI    | `Aggregators`, `Converters`, `Measures`, `Spreaders`, `Units`                                  |
| `mocks/src/components/Views/Definition/`  | Definition UI     | `Categories`, `Taxonomies`, `Metrics`, `PlanTypes`, `FlowView`                                 |
| `mocks/src/components/Views/Integration/` | Integration UI    | `Connectors`, `Mapping`, `Readers`, `Writers`                                                  |
| `mocks/src/components/Views/Planning/`    | Planning UI       | `BusinessOverview`, `PlanExecution`, `Plans`                                                   |
| `mocks/src/data/`                         | Mock data         | `mockData.ts`, `flowData.ts`                                                                   |
| `mocks/src/lib/`                          | Mock utilities    | `tabulatorUtils.ts`, `utils.ts`                                                                |
| `mocks/tests/`                            | Mock tests        | Data tests, lib tests, logic tests                                                             |

---

## How to use this file

1. **Agent receives a user direction** (e.g. "fix the bitemporal mutation logic").
2. **Scan the top-level table** — `backend/serde/rdbms/` is the relevant leaf.
3. **Load only that leaf's brief** (`src/backend/serde/rdbms/brief.md`) and the
   relevant code files; do not load the entire tree.
4. **If the user mentions a service name** (e.g. "metrics API"), grep this file
   for the keyword, then load the matching brief.

---

_Last updated: 2026-05-21_
