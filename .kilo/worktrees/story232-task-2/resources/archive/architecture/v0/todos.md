# ChronoCone v0 Architecture - Outstanding Items

This document tracks items that are either marked as incomplete in the architecture documents or are mentioned in the requirements or concepts but are missing from the architecture documents or are not sufficiently specified.

## 1. Components with Incomplete Specifications

### 1.1 MCP Server - Business Analysis

**Status**: ✅ Documented in architecture and requirements

**What was addressed**:
The MCP Server - Business Analysis interfaces are now fully specified in both requirements and architecture documents.

**Requirements document** ([v0_requirements.md Section 3.3](../../requirements/v0_requirements/v0_requirements.md#33-mcp-server---business-analysis)) now includes:

- Purpose and target users
- Complete responsibilities list
- **Tools Interface** with five tools:
  - `update`: Refresh planning data by invoking calculation engine
  - `read`: Retrieve metric values with dimensions, time grain, length, and relationships
  - `lineage`: Trace data sources for metrics or specific cells
  - `diff`: Compare two metric versions for differences
  - `metric_names`: List published metrics in a namespace
- **Resources Interface**:
  - Metric Form and Purpose: Descriptions of what metrics measure and how they relate
  - Domain Definition and Procedures: Business context stored in object store
- **Prompts Interface**: PDD scripts for analysis workflows

**Architecture document** ([`mcp_server_business_analysis.md`](mcp_server_business_analysis.md)) now includes:

- Detailed tool specifications with parameters, behavior, and return values
- Resource type descriptions
- Available prompts for guided analysis workflows
- Design considerations for data volume management and context preservation

**Requirements Reference**: [v0_requirements.md Section 3.3](../../requirements/v0_requirements/v0_requirements.md#33-mcp-server---business-analysis)

---

### 1.2 MCP Server - Planning

**Status**: ✅ Documented in architecture and requirements

**What was addressed**:
The MCP Server - Planning interfaces are now fully specified in both requirements and architecture documents.

**Requirements document** ([v0_requirements.md Section 3.4](../../requirements/v0_requirements/v0_requirements.md#34-mcp-server---planning)) now includes:

- Purpose and target users
- Complete responsibilities list
- **Tools Interface** with three tools:
  - `metric_names`: List both published and unpublished metrics with status tags
  - `publish`: Transition unpublished metrics to published state
  - `push`: Save transformer, shifter, or filter code as unpublished artifact
- **Resources Interface**:
  - Metric Form and Purpose: Descriptions for understanding existing models
  - Domain Definition and Procedures: Business context for alignment with requirements
- **Prompts Interface**: PDD scripts for development workflows

**Architecture document** ([`mcp_server_planning.md`](mcp_server_planning.md)) now includes:

- Detailed tool specifications with parameters, behavior, and return values
- Resource type descriptions
- Available prompts for guided development workflows (creating transformers, shifters, filters, generating test data, debugging, publishing)
- Design considerations for correctness, transparency, and iterative refinement

**Requirements Reference**: [v0_requirements.md Section 3.4](../../requirements/v0_requirements/v0_requirements.md#34-mcp-server---planning)

---

### 1.3 External API Simulator

**Status**: ✅ Documented in architecture and requirements

**What was addressed**:
The External API Simulator is now fully specified in both requirements and architecture documents.

**Requirements document** ([v0_requirements.md Section 3.6](../../requirements/v0_requirements/v0_requirements.md#36-external-api-simulator)) now includes:

- Purpose and target users
- Complete responsibilities list
- `/metric_data` endpoint with query parameters (metric_name, time_grain, time_range, data_type)
- Response format specification with type-correct JSON values
- Configuration file description for predetermined values
- Random value generation rules (100-1000 for numeric types, 10-character strings for string types)
- Design considerations for type correctness, reproducibility, and simplicity

**Architecture document** ([`external_api_simulator.md`](external_api_simulator.md)) now includes:

- Detailed responsibilities section covering metric data supply, type lookup, configuration management, and random value generation
- Complete `/metric_data` endpoint specification with HTTP method, query parameters table, and example request/response
- YAML configuration file format with examples for both actuals and forecast values
- Random value generation table with rules per data type

**Requirements Reference**: [v0_requirements.md Section 3.6](../../requirements/v0_requirements/v0_requirements.md#36-external-api-simulator)

---

## 2. Components Missing from Architecture

### 2.1 IDE Plugins

**Status**: ✅ Documented in architecture and requirements

**What was addressed**:
The IDE Plugins component is now fully specified in both requirements and architecture documents.

**Requirements document** ([v0_requirements.md Section 3.7](../../requirements/v0_requirements/v0_requirements.md#37-ide-plugins)) includes:

- Purpose and target users (developers and planners creating/modifying transformers, shifters, and filters)
- Supported platforms (IntelliJ IDEA and Visual Studio Code)
- Core workflow: pull-modify-push for calculation assets
- Asset identification scheme (asset type, namespace family, namespace, function type, name)
- Key features: agentic interface, traditional code editing, test data generation, prompt capture, PDD script integration
- Interfaces: Metrics Service API integration, local sandbox management, AI assistant integration, version control integration
- Data captured with assets: Python code, test data, prompts, metadata

**Requirements document** ([v0_requirements.md Section 3.8](../../requirements/v0_requirements/v0_requirements.md#38-plugin-ui)) includes:

- Namespace connection panel with connection status and switching capability
- Explorer panel with three tabs (Metrics, Shifters, Filters) with expand/collapse controls
- Chat window for AI agent interaction
- Metric detail view showing forecast and actuals transformers
- Shifter and filter detail views
- Testing interface with test table, run functionality, and pass/fail display
- AI agent capabilities for test data manipulation
- Test data persistence alongside function code
- Sandboxed execution environment independent of local machine software

**Architecture document** ([`ui.md`](ui.md) - Plugin UI section) includes:

- Purpose and target users
- Supported platforms (IntelliJ Plugin and VS Code Extension)
- Responsibilities (pull-modify-push workflow, code editing, AI-assisted development, test results, sandbox management)
- Docker-based sandbox architecture with isolation, consistent runtime, namespace connection, and dependency management
- Connection flow sequence (container creation, code retrieval, synchronization, UI update)
- Push/pull operations with test data synchronization
- Key features: namespace connection panel, explorer panel, chat window, detail views, testing interface
- Component structure for both IntelliJ and VS Code implementations

**Requirements Reference**: [v0_requirements.md Section 3.7](../../requirements/v0_requirements/v0_requirements.md#37-ide-plugins) and [Section 3.8](../../requirements/v0_requirements/v0_requirements.md#38-plugin-ui)

---

### 2.2 Connectors

**Status**: ✅ Documented in architecture

**What was addressed**:
Connectors are now documented in [v0_architecture.md - Connectors](v0_architecture.md#connectors). The documentation covers:

- Two purposes: external service connectors and object store connectors
- Storage configuration via `storage_config` table
- v0 implementation using "local" driver only
- Time series data storage location resolution

**Remaining work**:

- Connector interface specification for external services (not needed for v0)
- How connectors are registered and discovered (not needed for v0)
- How connectors map external types to ChronoCone types (not needed for v0)
- How actuals transformers call connectors (not needed for v0)
- Lazy evaluation mechanism for batching connector calls (not needed for v0)

**Requirements Reference**: [v0_requirements.md - Working With External Systems](../../requirements/v0_requirements/v0_requirements.md#working-with-external-systems)

**Concepts Reference**: [basic_concepts.md - Working With External Systems](../../../resources/system_concepts/basic_concepts.md#working-with-external-systems)

---

### 2.3 Time Series Data File Format

**Status**: ✅ Documented in architecture

**What was addressed**:
The time series data file format is now documented in [v0_architecture.md - Time Series Object Data Format](v0_architecture.md#time-series-object-data-format). The documentation covers:

- Parquet file format with custom metadata and table structure
- Custom metadata schema (metric definitions, dimensions, time grain, forecast length, data types)
- Table column structure (metric name, dimension values, forecast/actuals flag, timestamp, value)
- In-memory representation as multi-level dictionary
- StorageInterface responsibility for caching objects to avoid duplicate reads

**Requirements Reference**: [v0_requirements.md - Document Types and Documents](../../requirements/v0_requirements/v0_requirements.md#document-types-and-documents)

---

## 3. Features Missing from Architecture

### 3.1 Publishing Workflow

**Status**: ✅ Documented in architecture

**What was addressed**:
The publishing workflow is now documented in both the data storage and metrics service architecture documents.

**Data Storage** ([`data_storage.md`](data_storage.md)) now includes:

- `is_dirty` column added to `transformer`, `shifter`, and `filter` tables
- Documentation of the dirty flag mechanism for controlling publication state
- Updated ER diagram showing the new columns and tables
- Updated "Tables with Validity" section to include shifter and filter tables

**Metrics Service** ([`metrics_service.md`](metrics_service.md)) now includes:

- `/definitions` endpoint for listing, retrieving, and updating transformer, shifter, and filter definitions
- Support for `include_dirty` parameter to filter published vs unpublished definitions
- Support for `include_test_data` parameter to include test data with definitions
- `/admin` endpoint with `publish` action for publishing unpublished definitions
- Complete request/response examples for all operations

**Requirements Reference**: [v0_requirements.md - Function Lifecycle](../../requirements/v0_requirements/v0_requirements.md#function-lifecycle)

---

### 3.2 Calculation Engine

**Status**: ✅ Documented in architecture

**What was addressed**:
The Calculation Engine is now fully documented in [`calculation_engine.md`](calculation_engine.md). The documentation covers:

- **Function Definition and Storage**: Python function representation with naming conventions that map function names and parameters to metric names, and the unpublished artifact workflow
- **Publish Phase**: AST analysis for function signature extraction, name-to-metric mapping, dimension retrieval and matching, shifter discovery (including shifter chains), and DAG construction
- **Update Phase**: DAG traversal in topological order, single-threaded execution (v0), sandboxed Docker container execution, execution context management, and cell-by-cell transformer evaluation
- **Result Capture**: Parquet file generation using the defined time series data format, document creation with lineage tracking
- **Error Handling**: Publish-time and update-time error types with clear messages and resolutions
- **Design Considerations**: Determinism, separation of concerns, and extensibility for future execution modes

**Architecture document** ([`calculation_engine.md`](calculation_engine.md)) includes:

- Purpose and responsibilities
- Function definition and storage with naming conventions
- Publish phase operations (AST analysis, dimension matching, shifter discovery, DAG construction)
- Update phase operations (DAG traversal, sandboxed execution, context management, cell-by-cell evaluation)
- Result capture in Parquet format
- Error handling at publish time and update time
- Design considerations for determinism and extensibility

**Requirements Reference**: [v0_requirements.md - Transformers, Shifters, and Filters](../../requirements/v0_requirements/v0_requirements.md#transformers-shifters-and-filters)

---

### 3.3 Document Types and Documents

**Status**: ✅ Documented in architecture

**What was addressed**:
Document types and documents are now fully documented in the architecture documents.

**Data Storage** ([`data_storage.md`](data_storage.md)) includes:

- **`ts_doc_type` table**: Document type definitions that serve as templates for plans, with validity tracking for versioning
- **`ts_doc_row_type` table**: Metric definitions within a document type, specifying which metrics appear in documents
- **`ts_doc` table**: Actual document instances with `as_of` timestamp, storage configuration, and location
- **`ts_doc_tags` table**: Tags for document categorization and filtering
- **`ts_doc_input_lineage` table**: Tracks document derivation and ancestry for lineage

**Metrics Service** ([`metrics_service.md`](metrics_service.md)) includes:

- `TSDocumentType` Pydantic model for document type validation
- `TSDocumentRowType` Pydantic model for row type validation
- Description of how document types define the structure of planning documents

**Calculation Engine** ([`calculation_engine.md`](calculation_engine.md)) includes:

- Document creation section describing how calculation results are associated with documents
- Document captures: document type, `as_of` timestamp, storage location, and lineage information

**v0 Architecture** ([`v0_architecture.md`](v0_architecture.md)) includes:

- Overview of document type definitions and documents
- Lineage tracking including document type and document references

**Requirements Reference**: [v0_requirements.md - Document Types and Documents](../../requirements/v0_requirements/v0_requirements.md#document-types-and-documents)

---

### 3.4 Built-in Types

**Status**: ✅ Documented in architecture and requirements

**What was addressed**:
The built-in types system is now fully specified in both requirements and architecture documents.

**Requirements document** ([v0_requirements.md - Built-in Types](../../requirements/v0_requirements/v0_requirements.md#built-in-types)) now includes:

- Purpose and scope of the type system
- Complete type definitions with JSON and Python representations:
  - **String**: JSON string ↔ Python `str`
  - **Integer**: JSON integer ↔ Python `int`
  - **Float**: JSON number ↔ Python `float`
  - **Currency**: JSON integer (cents) ↔ Python `int` (rendered as dollars in UI)
  - **DateTime**: JSON ISO 8601 string ↔ Python `datetime` object
- Design considerations for type safety, validation, and serialization

**Architecture document** ([`v0_architecture.md` - Built-in Types](v0_architecture.md#built-in-types)) includes:

- Detailed type specifications with representation mappings
- JSON serialization rules for each type
- Python type annotations and validation requirements
- Currency handling as integer cents with UI rendering as dollars
- DateTime handling with ISO 8601 string representation

**Key Design Decisions**:

1. **Integers and Floats are distinct types**: The system supports both `int` and `float` as separate types, not combined into a single "number" type.

2. **Currency is a ChronoCone type**: Currency values are stored as integers representing cents, avoiding floating-point precision issues. UI displays values as dollars with two decimal places.

3. **DateTime uses ISO 8601**: Dates and times are serialized as ISO 8601 strings in JSON and deserialized to Python `datetime` objects.

4. **No user-defined types**: v0 does not support user-defined types. All types are built-in.

5. **No connector-defined types**: Connector types always match the types of the metrics they serve. Type mismatches are caught at publish time during DAG validation.

6. **Time series and metrics**: Time series and metric representation is documented in the calculation engine and data storage architecture documents.

**Requirements Reference**: [v0_requirements.md - Built-in Types](../../requirements/v0_requirements/v0_requirements.md#built-in-types)

---

## 4. Partially Specified Features

### 4.1 Core Service Admin Interface

**Status**: ✅ Documented in architecture

**What was addressed**:
The Core Service Admin Interface is now fully specified in [`core_service.md`](core_service.md).

**Architecture document** ([`core_service.md`](core_service.md)) now includes:

- **`/admin/health` endpoint**: GET request returns health status of the Core Service and all registered services
  - Overall system status (healthy, degraded, unhealthy)
  - Timestamp of health check
  - Individual service health entries with name, status, URL, details, and error information
  - Health determination via HTTP requests to service endpoints and connectivity checks for database and object store

- **`/admin/config` endpoint**: GET and POST requests for configuration management
  - GET returns current configuration parameters
  - POST updates configuration parameters with partial update support
  - Initial configuration parameters:
    - Database: host and port for PostgreSQL connection
    - Object store: local filesystem path for time series data storage
  - Validation of object store path before applying changes
  - Immediate effect for new connections, existing connections continue with previous config

- **Pydantic models** for admin endpoints:
  - `ServiceHealth`: Health status of individual services
  - `HealthResponse`: Complete health check response structure
  - `DatabaseConfig`: Database connection parameters
  - `ObjectStoreConfig`: Object store configuration
  - `SystemConfig`: Complete system configuration
  - `ConfigUpdateRequest`: Request model for partial configuration updates
  - `ConfigResponse`: Response model for GET requests
  - `ConfigUpdateResponse`: Response model for successful updates
  - `ConfigErrorResponse`: Response model for update failures

**Requirements Reference**: [v0_requirements.md - Core Service Section](../../requirements/v0_requirements/v0_requirements.md#31-core-server)

---

### 4.2 Core Service Queuing

**Status**: ✅ Documented in architecture

**What was addressed**:
The Core Service queuing functionality is now fully specified in [`core_service.md`](core_service.md).

**Architecture document** ([`core_service.md`](core_service.md)) now includes:

- **Motivation**: Asynchronous processing for long-running operations and publish-subscribe pattern for multiple recipients
- **Queue Model**: Named-queue model with in-memory implementation for v0, migration path to SNS/SQS or RabbitMQ for post-v0
- **Queue Operations**:
  - Send message to queue via POST to `/queue` endpoint
  - Register subscriber via POST to `/queue/subscribe` endpoint
  - Unregister subscriber via POST to `/queue/unsubscribe` endpoint
- **Message Delivery**: Synchronous delivery to all registered subscribers with failure logging
- **Queue Utility Object**: Programmatic interface for both senders and receivers with callback dispatch model
- **Pydantic Models**: `QueueSendMessage`, `QueueSubscribe`, `QueueUnsubscribe`, and `QueueMessageDelivery` models

**Requirements Reference**: [v0_requirements.md - Core Service Section](../../requirements/v0_requirements/v0_requirements.md#31-core-server)

---

### 4.3 Metrics Service Object Store

**Status**: ✅ Documented in architecture

**What was addressed**:
The Metrics Service Object Store is now documented in [`metrics_service.md`](metrics_service.md) and [`data_storage.md`](data_storage.md).

**Architecture document** ([`metrics_service.md`](metrics_service.md)) includes:

- **Data Stores section**: Describes the object store purpose for storing metric data files (actuals and forecasts)
- **v0 implementation**: Local filesystem for simplicity
- **Future migration path**: Cloud storage with presigned URLs for scalability
- **/metrics Interface**: Returns references to metric data in object store rather than returning data directly

**Architecture document** ([`data_storage.md`](data_storage.md)) includes:

- **Object Store documentation**: Connectors, Parquet format, time series data format
- **Storage configuration**: How storage locations are configured via `storage_config` table
- **Time Series Object Data Format**: Parquet file structure with custom metadata and table columns

**Requirements Reference**: [v0_requirements.md - Metrics Service Section](../../requirements/v0_requirements/v0_requirements.md#32-metric-service)

---

### 4.4 Metrics Page in UI

**Status**: ✅ Documented in architecture

**What was addressed**:
The Metrics Page in UI is now fully documented in [`ui.md`](ui.md).

**Architecture document** ([`ui.md`](ui.md)) includes:

- **Page Header Controls**: Namespace family dropdown, namespace dropdown, and published/unpublished toggle
- **Time Series Display**: Layout with actuals (left) and forecasts (right), time labels, and cell values
- **Time Label Formats**: Configuration options for weekly and daily time grains
- **Vertical Arrangement**: DAG order with input metrics at top and output metrics at bottom
- **Unpublished Metric Display**: Visual distinction through color differentiation and pattern/icon overlay
- **Metric Detail View**: Graph panel showing historical actuals vs. forecasts and future forecasts
- **Transformer Definition Panels**: For users with plan modification permission, showing forecast and actuals transformer definitions
- **Component Architecture**: React components including MetricsPage, NamespaceSelector, PublishedToggle, MetricsList, MetricRow, MetricDetail, TimeSeriesGraph, and TransformerPanel

**Requirements Reference**: [v0_requirements.md - Web UI Section](../../requirements/v0_requirements/v0_requirements.md#35-web-ui)

---

## 5. Advanced Topics Not Addressed

### 5.1 AI Enablement

**Status**: Discussed in concepts but not addressed in architecture

**Concepts Reference**: [basic_concepts.md - AI Enablement And User Experience](../../../resources/system_concepts/basic_concepts.md#ai-enablement-and-user-experience)

**What's Missing**:

- How AI/agents will be integrated into user experience
- How consistency and reliability will be ensured
- Balance between AI assistance and deterministic planning

---

## 6. Documentation Requests

### 6.1 Data Storage Documentation

**Status**: ✅ Completed

**What was addressed**:
Created comprehensive [`data_storage.md`](data_storage.md) document consolidating all storage-related content:

- Object Store documentation (connectors, Parquet format, time series data format)
- PostgreSQL database overview with schema organization
- Bitemporal validity model explanation with table list
- ER diagram using Mermaid format with all tables and column types
- Detailed table documentation for all five schemas (policy, taxonomy, timeseries, connectors, morphisms)
- Seed data documentation for enumeration tables

**Files Modified**:

- Created: `memory-bank/architecture/v0/data_storage.md`
- Updated: `memory-bank/architecture/v0/v0_architecture.md` (replaced detailed storage sections with summary and link)
- Updated: `memory-bank/architecture/v0/toc.md` (added entry for data_storage.md)

---

## 7. Summary

| Category   | Item                                      | Priority | Blocking |
| ---------- | ----------------------------------------- | -------- | -------- |
| Incomplete | MCP Server - Business Analysis interfaces | ✅ Done  | No       |
| Incomplete | MCP Server - Planning interfaces          | ✅ Done  | No       |
| Incomplete | External API Simulator specification      | ✅ Done  | No       |
| Missing    | IDE Plugins architecture                  | ✅ Done  | No       |
| Missing    | Connectors architecture                   | ✅ Done  | No       |
| Missing    | Time Series Data File Format              | ✅ Done  | No       |
| Missing    | Publishing workflow                       | ✅ Done  | No       |
| Missing    | Calculation engine specification          | ✅ Done  | No       |
| Missing    | Document types/documents storage          | ✅ Done  | No       |
| Missing    | Built-in types system                     | ✅ Done  | No       |
| Partial    | Core Service admin interface              | ✅ Done  | No       |
| Partial    | Core Service queuing                      | ✅ Done  | No       |
| Partial    | Metrics Service object store              | ✅ Done  | No       |
| Partial    | Metrics page in UI                        | ✅ Done  | No       |
| Advanced   | AI enablement approach                    | Medium   | No       |
| Doc        | Data Storage documentation                | ✅ Done  | No       |

**Legend**:

- **Priority**: High = needed for core functionality, Medium = needed for complete v0, Low = nice to have
- **Blocking**: Yes = blocks other work or core use cases, No = can work around
