Directives for agentic or AI writers and editors:

1. You MUST leave this list of directives in place as the first item in the document and unaltered as changes are made to this document.
2. When editing this document you MUST use either the Technical Writer or Technical Editor roles. If you are not in one of these modes you MUST NOT edit this document and inform the user of the requirement.
3. This document contains annotations that are labelled as "Notes:" or "Questions:" and you MUST NOT edit these annotations. You should move them to keep them with the content they relate to.

--- Begin Document Content ---

# Timecone.ai and ChronoCone v0 Architecture

## Introduction

**Timecone.ai** is a visionary planning system moving beyond "flat time" to act as an "Institutional Brain." It manages multiple possible futures simultaneously, providing a flexible framework that supports automated planning, human-in-the-loop planning, and collaborative planning across different organizational levels.

At its core, Timecone.ai relies on **ChronoCone**, a deterministic calculation engine that executes strict point-in-time calculation DAGs (Directed Acyclic Graphs). By generating arrays of these deterministic DAGs (representing conflicting "Permissionless Assertion Layers" or alternative futures), Timecone.ai allows agentic AI to reason at a high level. Agents can propose alternative futures, and the ChronoCone engine calculates them deterministically, grounding hallucination-prone AI in verifiable math.

The name "ChronoCone" captures a central metaphor for how planning works in this system. Imagine cones that widen as they extend forward in time, representing the expanding range of possible future states. Plans can expand in two ways: through refinement, where general plans develop into specific ones (for example, country-level plans break into regional plans), or through temporal extension, where short-term plans extend into longer horizons. Longer-term plans typically have less specificity, more uncertainty, or both. This metaphor shapes how the system represents and manages planning data, allowing users to navigate between high-level strategic views and detailed operational plans while understanding the relationships between them.

The system serves diverse planning domains including logistics, manufacturing, and healthcare. These domains share common planning challenges such as coordinating across organizational boundaries, managing uncertainty, and balancing resource constraints against operational goals. Rather than building separate systems for each domain, ChronoCone provides a unified platform that can be configured to address these shared challenges while accommodating domain-specific requirements.

For detailed requirements, see the [v0 Requirements Document](../requirements/v0_requirements/v0_requirements.md).

## System Components

The v0 architecture consists of seven main components, each documented in its own file. These components work together to provide a complete planning system, from data storage and computation through user interaction and external system integration.

### Core Service

The Core Service serves as the central coordinator for the ChronoCone system. It handles system startup and service discovery, allowing other components to find and communicate with each other without hard-coded connection information. This design choice reflects a key architectural principle: components should be loosely coupled so they can be developed, deployed, and scaled independently.

In the v0 release, the Core Service also provides miscellaneous functionality such as queuing in a minimal single-device deployment suitable for testing and demonstration. This arrangement simplifies the initial setup while still demonstrating the full system workflow. In future versions, this centralized service will be replaced with a cloud-based server or calls to an AWS API Gateway that will feed into Lambda functions or an ECS-based server. The service discovery mechanism ensures that this migration can happen without requiring changes to the components that depend on it.

See [Core Service](core_service.md) for detailed documentation.

### Metrics Service

The Metrics Service is the heart of ChronoCone's data management capabilities. It is responsible for storing, maintaining, and updating metrics that are used by the ChronoCone system, as well as the metadata associated with them (e.g., categories, namespaces). The service retrieves metrics including actuals and forecasts, and implements the compute engine that updates metrics when new data arrives.

The compute engine represents one of the most important design decisions in the system. Rather than requiring explicit "update metric" commands, the engine responds automatically when new data arrives. This event-driven approach accommodates a wide variety of usage scenarios, from batch updates to real-time data feeds. The compute engine uses three types of functions to process data: transformers calculate metric values from inputs, shifters align metrics with different dimensional structures, and filters split metrics into constituent parts. This functional approach allows planners to express complex calculations in terms of composable, reusable operations.

See [Metrics Service](metrics_service.md) for detailed documentation.

### MCP Server - Business Analysis

The Business Analysis MCP Server facilitates questions asked about metrics by general ChronoCone users. It answers basic questions about what metrics mean, how they are related to other metrics, and what particular forecasts are for a given set of categories. It also answers questions about why something happened or what changed when something unexpected occurred.

This component addresses a common challenge in planning systems: making the data accessible to users who are not technical experts. By providing a natural language interface through the Model Context Protocol (MCP), the server allows business analysts to explore and understand the planning data without needing to learn query languages or navigate complex interfaces. The server is designed to manage the volume of data carefully, preventing LLMs or agents from being overwhelmed by the sheer amount of information available, which could result in context rot and degraded response quality.

The server implements three MCP interfaces:

**Tools Interface**: Provides callable functions including `update` (refresh planning data), `read` (retrieve metric values with dimensions, time grain, and relationships), `lineage` (trace data sources for metrics or cells), `diff` (compare two metric versions), and `metric_names` (list published metrics in a namespace).

**Resources Interface**: Provides contextual information about metric form and purpose, plus domain definitions and procedures stored in the object store.

**Prompts Interface**: Provides PDD scripts that guide users through common analysis workflows like exploring planning domains, investigating anomalies, and comparing forecasts.

See [MCP Server - Business Analysis](mcp_server_business_analysis.md) for detailed documentation.

### MCP Server - Planning

The Planning MCP Server helps planners create or update models, using a PDD (Prompt Driven Development) style of input that will in turn be translated into a programmatic form. A PDD script is a pseudo-executable script used by LLMs and agentic tools to perform structured tasks. This approach bridges the gap between natural language planning expertise and the formal specifications required by the compute engine.

The server also aids in the generation and updating of test data, executes tests, and provides feedback to the planner to help in fixing errors. This integrated development workflow supports planners through the entire lifecycle of creating and refining calculation logic, from initial concept through tested, production-ready code. By combining AI-assisted development with automated testing, the server reduces the technical barrier to creating sophisticated planning models.

The server implements three MCP interfaces:

**Tools Interface**: Provides callable functions including `metric_names` (list both published and unpublished metrics with status tags), `publish` (transition unpublished metrics to published state), and `push` (save transformer, shifter, or filter code as an unpublished artifact).

**Resources Interface**: Provides contextual information about metric form and purpose, plus domain definitions and procedures stored in the object store. This context helps planners ensure their models align with business requirements.

**Prompts Interface**: Provides PDD scripts that guide planners through development workflows like creating transformers, shifters, and filters, generating test data, debugging errors, and publishing changes.

See [MCP Server - Planning](mcp_server_planning.md) for detailed documentation.

### Web UI

The Web UI is the primary mechanism for providing a visual display of metrics through a browser-based interface. It displays metrics and allows display and update of metadata for end users interacting with the ChronoCone system. The Web UI serves as the primary interface for users who need to view planning results, explore metric relationships, and manage the metadata that defines how the system operates.

The Web UI is designed to be intuitive for business users while providing access to the full power of the underlying system. It presents metrics in context, showing how values relate to organizational structures and time horizons. The metadata management capabilities allow authorized users to define new categories, elements, and metrics without requiring direct database access or technical expertise.

See [Web UI](ui.md) for detailed documentation.

### Plugin UI

The Plugin UI provides the user interface for planners working on transformers, shifters, and filters through IDE plugins. Unlike the Web UI, which is browser-based, the Plugin UI is integrated directly into the developer's IDE (IntelliJ IDEA or Visual Studio Code), providing a seamless development experience for creating and modifying calculation logic.

The Plugin UI supports the pull-modify-push workflow where planners retrieve calculation assets from the Metrics Service, modify them using either traditional code editing or AI-assisted development, test their changes locally, and push modified assets back to the Metrics Service. This integrated development environment bridges the gap between the metrics server and the planner's local workspace.

See [Plugin UI](plugin_ui.md) for detailed documentation.

**Note:** The Plugin UI documentation is currently a placeholder and will be filled in as plugin development progresses.

### External API Simulator

The External API Simulator provides simulated external data for demonstrations and testing. It serves canned data for use in demos when the system is not yet connected to real external systems. This component is essential for demonstrating the full system workflow without requiring access to production data sources.

The simulator reflects a practical reality: planning systems must integrate with many external data sources, but those integrations are often complex and time-consuming to establish. By providing a simulated data environment, the v0 system can demonstrate its capabilities while the real integrations are being developed. The simulator also provides a stable, predictable data environment for testing, which is valuable for validating system behavior and reproducing issues.

See [External API Simulator](external_api_simulator.md) for detailed documentation.

## Data Storage

ChronoCone uses a two-tier storage architecture that separates metadata from time series data and function code:

- **PostgreSQL Database**: Stores metadata, configuration, and lineage information. This relational data requires transactional consistency and benefits from complex queries with joins and filters.

- **Object Store**: Stores time series data files in Parquet format and Python function files (transformers, shifters, and filters). Time series data is accessed in bulk and benefits from columnar storage compression and efficient cloud storage access patterns. Function files are stored as Python code, enabling versioning, audit trails, and reproducibility of historical calculations.

For comprehensive documentation of the storage architecture, including:

- Connector abstraction and configuration
- Time series data storage and Parquet format
- PostgreSQL database schema with ER diagram
- Bitemporal validity model
- Detailed table documentation

See [Data Storage](data_storage.md) for detailed documentation.

## Versioning and Immutability

All artifacts in ChronoCone are versioned and immutable. This principle applies comprehensively to:

- **Document type definitions**: The structure and logic that defines how a plan operates
- **Documents**: The concrete outputs produced when a document type executes with specific input data
- **Transformers**: Functions that produce metric values from inputs
- **Shifters**: Functions that align metrics with different dimensional structures
- **Filters**: Functions that split metrics into constituent parts

When any artifact is modified—whether recalculating a plan, updating a transformer, or adjusting a filter—the system creates a new version rather than updating the existing one. This immutability is fundamental to the system's design and enables several critical capabilities:

1. **Historical reproducibility**: Any historical calculation can be reproduced exactly by retrieving the specific versions of all artifacts that were active at that time
2. **Audit trails**: Complete records of what changed, when, and why support compliance and debugging
3. **Safe evolution**: New versions can be tested and validated before being used in production, with the ability to roll back if problems emerge
4. **Lineage tracking**: The system can trace how any value was derived, which is essential for understanding and explaining planning results

## Lineage Tracking

ChronoCone maintains lineage information at two distinct levels, enabling users to understand both how specific values were computed and how the system's components relate to each other.

### Document-Level Lineage

For any value in a time series—whether an actual or a forecast—the system can trace the complete flow of inputs and calculations that produced it. This includes:

- The source data that contributed to the value
- The sequence of transformers, shifters, and filters applied
- The specific versions of each function used
- The document type and document that contained the calculation

This lineage enables users to answer questions like "Why does this forecast show 500 units?" by walking backward through the calculation chain to see exactly which inputs and operations produced that result.

### Definition-Level Lineage

The system tracks relationships between the definitions of metrics, transformers, shifters, and filters as versions change over time. This includes:

- Which transformer versions were associated with which metric versions
- How input-output relationships between metrics evolved
- When shifters or filters were added, modified, or deprecated
- The dependency graph connecting all calculation components

This lineage enables users to understand how the planning system itself has evolved, answering questions like "What changed in the labor calculation between March and April?" by showing the differences in the relevant function versions and their relationships.

### Diff Operations

An important operation supported by lineage tracking is the ability to compare two versions of an artifact. The system provides diff capabilities that are primarily surfaced as:

- **Compare two document versions**: Shows what changed between two runs of the same document type, including differences in input data, intermediate calculations, and final outputs. This helps users understand why results differ from one planning cycle to the next.

- **Compare two function versions**: Shows the differences between two versions of a transformer, shifter, or filter, including changes to the code logic, input/output specifications, and associated metadata. This helps users understand how calculation logic has evolved.

These diff operations are essential for debugging, auditing, and understanding the impact of changes to the planning system.

## Function Lifecycle

Transformers, shifters, and filters follow a defined lifecycle that governs how they are created, modified, and made available for production calculations. This lifecycle ensures controlled changes, testing, and auditability while preserving the ability to understand historical calculations. The lifecycle design addresses a fundamental tension in planning systems: the need for agility in updating calculation logic versus the need for stability and reproducibility in production calculations.

### Function Storage

All transformer, shifter, and filter functions are stored as Python code in an object store accessed through a connector abstraction. In v0, the connector provides access to the local filesystem. Future versions will support cloud-based object stores such as AWS S3. This storage approach treats calculation logic as data, which enables versioning, audit trails, and the ability to retrieve the exact code that was used for any historical calculation.

Functions are identified by name within a namespace family. A function name must be unique within its namespace family to prevent ambiguity during plan calculation. However, the same name may appear in different namespace families or in different namespaces within the same family. This scoping allows different teams or organizations to develop their own calculation logic without worrying about name collisions, while ensuring that within any given planning context, function names are unambiguous.

### Three States

Each transformer, shifter, or filter exists in one of three states at any given time. These states control whether the function participates in production calculations and how it can be modified:

| State           | Description                                                    | Used in Calculations |
| --------------- | -------------------------------------------------------------- | -------------------- |
| Published       | Stored in object store and actively used by calculation engine | Yes                  |
| Unpublished     | Stored in object store but not used for plan calculations      | No                   |
| Local Workspace | Exists in local development environment (Docker container)     | No                   |

The Published state represents production-ready code that the calculation engine will use. The Unpublished state represents code that has been saved to the object store but is not yet approved for production use—this might be code that is still being tested or is awaiting review. The Local Workspace state represents code that is actively being developed and exists only in a developer's environment.

### State Transitions

Functions transition between states through specific actions. These transitions are designed to support a safe development workflow while maintaining control over what code runs in production.

**Local Workspace → Unpublished**: When development work is complete, users push functions from their local workspace to the object store. The system creates a new database entry marked as unpublished. The function code is stored but not used by the calculation engine. This transition represents a checkpoint in the development process—the code is saved and can be shared with others, but it is not yet affecting production calculations.

**Unpublished → Published**: When code has been tested and approved, users publish functions to a namespace family and namespace. The system updates the database to mark the function as published, making it active for production calculations. This transition is the gate that ensures only reviewed and approved code affects business decisions.

**Published → Local Workspace**: When a function needs modification, users pull a copy from the object store to their local workspace. The published version remains unchanged and continues to be used for calculations. This transition enables iterative development without disrupting production systems.

### Versioning and Immutability

When a function is modified, the system never overwrites existing code in the object store. Instead, a new object is created with a new database entry. This immutability is a key design principle that enables several critical capabilities.

The approach preserves complete history of function versions, which is essential for understanding how calculations were performed for past documents. When a business user asks why a particular forecast was generated, the system can retrieve the exact version of the transformer that produced it. The approach also supports rollback to previous versions if a new version introduces problems, and maintains an audit trail of all changes for compliance and debugging purposes.

Each function version has an associated validity period recorded in the database. The calculation engine selects the appropriate version based on when a document was created. This temporal scoping ensures that historical documents remain reproducible even as the calculation logic evolves.

### Namespace Scoping

Publication operates within the namespace isolation model, which provides important controls over how functions are deployed and used. Functions are published to a specific namespace family and namespace combination, ensuring that changes affect only the intended calculation context. Function names must be unique within a namespace family, preventing ambiguity when the calculation engine resolves which function to use for a given metric. The same function name may exist in different namespace families, allowing different teams or organizations to develop their own implementations of similar calculations.

This scoping model supports several important use cases. Development and production environments can exist as separate namespaces within the same family, allowing code to be tested in isolation before promotion. Different organizations can maintain their own calculation logic without interference. And the same logical function (for example, "calculate_labor") can have different implementations appropriate to different business contexts.

For detailed requirements, see the [Function Lifecycle](../requirements/v0_requirements/v0_requirements.md#function-lifecycle) section in the requirements document.

## Design Philosophy

The v0 version of ChronoCone is a client/server system that implements a basic version of the system. This is not the ultimate organization, which will be built using a scalable cloud-based platform such as AWS. The v0 system introduces key abstractions and interfaces that will allow for migration to other platforms once the basic version is working. This approach reflects a pragmatic balance between immediate functionality and long-term scalability.

The v0 architecture prioritizes four key qualities that guide design decisions throughout the system. **Modularity** ensures that components are designed to be replaceable as the system evolves, allowing individual services to be upgraded or replaced without requiring changes to the entire system. **Abstraction** means that interfaces hide implementation details to enable future platform migrations, so that the code that uses a connector does not need to know whether it is accessing a local filesystem or cloud storage. **Demonstrability** ensures that the system supports end-to-end demonstrations with simulated external data, which is essential for validating the design and gathering feedback before investing in production integrations. **Extensibility** means that the foundation supports incremental addition of features, so that new capabilities can be added without requiring fundamental architectural changes.

## API Design Standards

All API endpoints in the ChronoCone v0 system are defined using the OpenAPI Specification (OAS) version 3.2.0. This standard provides a language-agnostic interface definition that enables consistent API development, documentation, and client generation across all system components. The choice of OpenAPI reflects a commitment to interoperability and developer experience: by adopting a widely-supported standard, ChronoCone ensures that its APIs can be easily understood and used by a broad range of tools and developers.

### OpenAPI Specification Compliance

The system adheres to the following OpenAPI standards, which together ensure that APIs are well-documented, consistent, and consumable.

**Endpoint Definition**: All API endpoints are documented using OpenAPI path and operation objects. This documentation serves as both a specification for implementation and a reference for API consumers. See [Paths Object](../../resources/apis/openapi/paths_object.md) and [Operation Object](../../resources/apis/openapi/operation_object.md) for details.

**Schema Definitions**: Data structures are defined using the OpenAPI Schema Object, which extends JSON Schema with additional properties. These schemas define the shape of request and response bodies, enabling automatic validation and documentation generation. See [Schema Object](../../resources/apis/openapi/schema_object.md) for schema definition guidance.

**Data Exchange Format**: All request and response bodies use JSON format (`application/json`). JSON was chosen for its simplicity, wide tooling support, and natural fit with the JavaScript-based frontend. The Media Type Object specifies the content type and schema for each endpoint. See [Media Type Object](../../resources/apis/openapi/media_type_object.md) for details.

**Reusable Components**: Common schemas, parameters, and responses are defined in the Components Object for reuse across multiple endpoints. This approach reduces duplication and ensures consistency across the API surface. See [Components Object](../../resources/apis/openapi/components_object.md) for component organization.

### OpenAPI Documentation References

| Topic                    | Reference                                                                     |
| ------------------------ | ----------------------------------------------------------------------------- |
| OpenAPI Overview         | [overview.md](../../resources/apis/openapi/overview.md)                       |
| Root OpenAPI Object      | [openapi_object.md](../../resources/apis/openapi/openapi_object.md)           |
| Paths and Operations     | [paths_object.md](../../resources/apis/openapi/paths_object.md)               |
| Operation Definitions    | [operation_object.md](../../resources/apis/openapi/operation_object.md)       |
| Parameter Definitions    | [parameter_object.md](../../resources/apis/openapi/parameter_object.md)       |
| Request Body Definitions | [request_body_object.md](../../resources/apis/openapi/request_body_object.md) |
| Response Definitions     | [responses_object.md](../../resources/apis/openapi/responses_object.md)       |
| Schema Definitions       | [schema_object.md](../../resources/apis/openapi/schema_object.md)             |
| Media Type Definitions   | [media_type_object.md](../../resources/apis/openapi/media_type_object.md)     |
| Reusable Components      | [components_object.md](../../resources/apis/openapi/components_object.md)     |
| Data Types               | [data_types.md](../../resources/apis/openapi/data_types.md)                   |

For a complete table of contents, see [OpenAPI TOC](../../resources/apis/openapi/toc.md).

## Built-in Types

ChronoCone uses a defined set of built-in types for all metric values. These types ensure consistent data representation across the system, from storage through API interfaces to user interfaces. The type system is intentionally simple for v0, focusing on the most common data types needed for planning applications.

### Type Definitions

Each built-in type has a defined representation in both JSON (for API serialization) and Python (for internal processing):

| Type     | JSON Representation           | Python Type | Description                                                     |
| -------- | ----------------------------- | ----------- | --------------------------------------------------------------- |
| String   | JSON string                   | `str`       | Text values for labels, names, and categorical data             |
| Integer  | JSON integer                  | `int`       | Whole numbers for counts, quantities, and discrete measurements |
| Float    | JSON number                   | `float`     | Decimal numbers for continuous measurements and ratios          |
| Currency | JSON integer (cents)          | `int`       | Monetary values stored as cents, rendered as dollars in UI      |
| DateTime | JSON string (ISO 8601 format) | `datetime`  | Points in time with full date and time precision                |

### String Type

Strings represent text values including labels, names, identifiers, and categorical data. They are passed through without transformation between JSON and Python.

**JSON Example**: `"Product A"`

**Python Example**: `"Product A"`

### Integer Type

Integers represent whole numbers for counts, quantities, and discrete measurements. The system distinguishes integers from floating-point numbers to preserve precision for values that should never have fractional components.

**JSON Example**: `42`

**Python Example**: `42`

### Float Type

Floats represent decimal numbers for continuous measurements, ratios, and calculated values that may have fractional components. The system distinguishes floats from integers to make the intended use clear and to enable appropriate validation.

**JSON Example**: `3.14159`

**Python Example**: `3.14159`

### Currency Type

Currency values represent monetary amounts. To avoid floating-point precision issues that can arise with decimal arithmetic, currency values are stored as integers representing cents (or the equivalent smallest unit in other currencies). User interfaces render these values as dollars with two decimal places.

**JSON Example**: `1999` (representing $19.99)

**Python Example**: `1999`

**UI Rendering**: `$19.99`

This approach ensures that currency arithmetic remains precise and that rounding behavior is predictable. All currency calculations within the system operate on integer cents, with conversion to display format happening only at the presentation layer.

### DateTime Type

DateTime values represent points in time. They are serialized as ISO 8601 strings in JSON for interoperability and deserialized to Python `datetime` objects for internal processing.

**JSON Example**: `"2024-03-15T14:30:00Z"`

**Python Example**: `datetime(2024, 3, 15, 14, 30, 0, tzinfo=timezone.utc)`

The ISO 8601 format was chosen because it is widely supported, unambiguous, and sorts correctly as a string. All DateTime values should include timezone information to avoid ambiguity.

### Type Constraints

The v0 type system has several important constraints:

**No User-Defined Types**: v0 does not support user-defined types. All metric values must use one of the built-in types. This simplification reduces complexity while still supporting the most common planning use cases.

**No Connector-Defined Types**: Connectors do not define their own types. Instead, connector outputs must match the types of the metrics they serve. Type mismatches between connectors and metrics are caught at publish time during DAG validation, ensuring that the calculation graph is type-consistent before any calculations execute.

**Type Validation**: The system validates type consistency at several points:

- When a transformer is published, the system verifies that the declared output type matches the metric's defined type
- When a shifter is published, the system verifies that input and output types are compatible
- When a filter is published, the system verifies that all output metrics have the same type as the input metric
- When data is retrieved from external systems, the system validates that the received data matches the expected type

### Time Series and Metrics

Time series and metric representation is documented in detail in the [Calculation Engine](calculation_engine.md) and [Data Storage](data_storage.md) architecture documents. These documents describe how typed values are organized into time series, how metrics are defined with their dimensional structure, and how data flows through the calculation pipeline.
