Directives for agentic or AI writers and editors:

1. You MUST leave this list of directives in place as the first item in the document and unaltered as changes are made to this document.
2. When editing this document you MUST use either the Technical Writer or Technical Editor roles. If you are not in one of these modes you MUST NOT edit this document and inform the user of the requirement.
3. This document contains annotations that are labelled as "Notes:" or "Questions:" and you MUST NOT edit these annotations. You should move them to keep them with the content they relate to.

--- Begin Document Content ---

# Timecone.ai and ChronoCone v0 Requirements

## Part 1: System Overview

### What is Timecone.ai?

Timecone.ai is a visionary planning system designed to act as an "Institutional Brain." It moves beyond "flat time" to manage multiple possible futures simultaneously, providing a flexible framework that supports automated planning, human-in-the-loop planning, and collaborative planning across different organizational levels.

The system overcomes the "Consensus Trap"—the organizational tendency to suppress discordance to force a singular, often inaccurate, future plan. Timecone.ai embraces discordance by allowing Permissionless Assertion Layers, where different domains can model conflicting futures.

At the core of Timecone.ai is the **ChronoCone** calculation engine. ChronoCone provides the deterministic, rule-based calculating substrate needed because pure agentic AI currently fails at complex planning (the "Hallucination Kills Process" problem). By grounding agents in rigid math, ChronoCone allows advanced agentic AI to reason safely at macro levels.

The name "ChronoCone" represents widening "cones" of possible future states. Plans can expand either through refinement, where general plans develop into specific ones (for example, country-level plans break into regional plans), or through temporal extension, where short-term plans extend into longer horizons.

The system serves diverse planning domains including logistics, manufacturing, and healthcare. These domains share common planning challenges such as coordinating across organizational boundaries, managing uncertainty, and balancing resource constraints against operational goals.

### Core Concepts

#### Document Types and Documents

ChronoCone distinguishes between document types and documents. A document type defines the structure and logic of a plan including the rules, calculations, and dependencies that govern how inputs transform into outputs. Planning organizations create and maintain these document types, which specify what data the plan requires, what operations to perform on that data, and what results to produce. For example, a warehouse labor document type might specify how to calculate the number of workers needed based on expected order volume and processing rates.

A document represents a specific document type at a particular point in time, derived from a document type plus input data. When a document type runs with concrete input data it produces a document with specific numerical or categorical outputs. For instance, a warehouse labor document might specify that 15 people should work the picking station from 2 PM to 6 PM on Tuesday, March 15th. These outputs guide actual operational decisions and actions.

ChronoCone versions and timestamps both document types and documents, recording metadata such as who made changes and what data sources contributed to the plan. Once created, a document type or document is immutable. Changes require creating a new version rather than modifying the existing one. This versioning approach maintains a complete audit trail of planning decisions and enables analysis of how plans evolved over time.

**Note:** In the database schema, document types are stored in the [`ts_doc_type`](memory-bank/terms.md:19) table and documents are stored in the [`ts_doc`](memory-bank/terms.md:21) table.

#### Versioning and Immutability

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

#### Lineage Tracking

ChronoCone maintains lineage information at two distinct levels, enabling users to understand both how specific values were computed and how the system's components relate to each other.

**Document-Level Lineage**: For any value in a time series—whether an actual or a forecast—the system can trace the complete flow of inputs and calculations that produced it. This includes:

- The source data that contributed to the value
- The sequence of transformers, shifters, and filters applied
- The specific versions of each function used
- The document type and document that contained the calculation

This lineage enables users to answer questions like "Why does this forecast show 500 units?" by walking backward through the calculation chain to see exactly which inputs and operations produced that result.

**Definition-Level Lineage**: The system tracks relationships between the definitions of metrics, transformers, shifters, and filters as versions change over time. This includes:

- Which transformer versions were associated with which metric versions
- How input-output relationships between metrics evolved
- When shifters or filters were added, modified, or deprecated
- The dependency graph connecting all calculation components

This lineage enables users to understand how the planning system itself has evolved, answering questions like "What changed in the labor calculation between March and April?" by showing the differences in the relevant function versions and their relationships.

#### Diff Operations

An important operation supported by lineage tracking is the ability to compare two versions of an artifact. The system provides diff capabilities that are primarily surfaced as:

- **Compare two document versions**: Shows what changed between two runs of the same document type, including differences in input data, intermediate calculations, and final outputs. This helps users understand why results differ from one planning cycle to the next.

- **Compare two function versions**: Shows the differences between two versions of a transformer, shifter, or filter, including changes to the code logic, input/output specifications, and associated metadata. This helps users understand how calculation logic has evolved.

These diff operations are essential for debugging, auditing, and understanding the impact of changes to the planning system.

#### Information Flow and Assertion Layers

ChronoCone plans exist at two levels from an end-user perspective. High-level plans typically have organizational associations—for example, the Retail plan belongs to one department while different parts of the Logistical plan belong to different departments within the Logistics organization.

Information within a single document flows deterministically as a Directed Acyclic Graph (DAG). A single plan is a strict, point-in-time calculation matrix.

However, Timecone.ai accommodates "Systemic Idea Darwinism" and "Permissionless Assertion Layers" by modeling branching, overriding, and merging of logical assertions. This is achieved by creating multiple, concurrent DAG plan documents representing differing assertions about the future. Generating probability distributions or determining consensus is achieved by creating higher-order, aggregating plan documents that depend on the varied base plans. This ensures that while the overarching system models probabilistic, complex futures, the underlying math remains strictly deterministic, rule-based, and traceable.

#### Working With External Systems

Plans do not exist in isolation, and ChronoCone is not a "do everything" system. It connects to external systems to both pull in data (for example, demand forecasts for products, inventory levels in warehouses, productivity targets or historical data for warehouse associates) and export data (for example, send labor capacity forecasts to a warehouse scheduling system). Within the boundaries of the ChronoCone system only ChronoCone-defined types exist, with the transformation to or from those types happening in connectors.

#### Built-in Types

ChronoCone uses a defined set of built-in types for all metric values. These types ensure consistent data representation across the system, from storage through API interfaces to user interfaces. The type system is intentionally simple for v0, focusing on the most common data types needed for planning applications.

**Type Definitions**:

| Type     | JSON Representation           | Python Type | Description                                                     |
| -------- | ----------------------------- | ----------- | --------------------------------------------------------------- |
| String   | JSON string                   | `str`       | Text values for labels, names, and categorical data             |
| Integer  | JSON integer                  | `int`       | Whole numbers for counts, quantities, and discrete measurements |
| Float    | JSON number                   | `float`     | Decimal numbers for continuous measurements and ratios          |
| Currency | JSON integer (cents)          | `int`       | Monetary values stored as cents, rendered as dollars in UI      |
| DateTime | JSON string (ISO 8601 format) | `datetime`  | Points in time with full date and time precision                |

**Key Design Decisions**:

1. **Integers and Floats are distinct types**: The system supports both `int` and `float` as separate types. This distinction preserves precision for values that should never have fractional components (counts, quantities) versus values that may have fractional components (ratios, calculated values).

2. **Currency is a ChronoCone type**: Currency values are stored as integers representing cents, avoiding floating-point precision issues that can arise with decimal arithmetic. User interfaces render these values as dollars with two decimal places. For example, the integer `1999` represents $19.99.

3. **DateTime uses ISO 8601**: Dates and times are serialized as ISO 8601 strings in JSON for interoperability and deserialized to Python `datetime` objects for internal processing. All DateTime values should include timezone information to avoid ambiguity.

4. **No user-defined types**: v0 does not support user-defined types. All metric values must use one of the built-in types. This simplification reduces complexity while still supporting the most common planning use cases.

5. **No connector-defined types**: Connectors do not define their own types. Instead, connector outputs must match the types of the metrics they serve. Type mismatches between connectors and metrics are caught at publish time during DAG validation, ensuring that the calculation graph is type-consistent before any calculations execute.

**Type Validation Requirements**:

The system must validate type consistency at several points:

- When a transformer is published, the system must verify that the declared output type matches the metric's defined type
- When a shifter is published, the system must verify that input and output types are compatible
- When a filter is published, the system must verify that all output metrics have the same type as the input metric
- When data is retrieved from external systems, the system must validate that the received data matches the expected type

## Part 2: v0 System Architecture

### Design Philosophy

The v0 version of ChronoCone is a client/server system that implements a basic version of the system. This is not the ultimate organization, which will be built using a scalable cloud-based platform such as AWS. The v0 system introduces key abstractions and interfaces that will allow for migration to other platforms once the basic version is working.

The v0 architecture prioritizes:

- **Modularity**: Components are designed to be replaceable as the system evolves
- **Abstraction**: Interfaces hide implementation details to enable future platform migrations
- **Demonstrability**: The system supports end-to-end demonstrations with simulated external data
- **Extensibility**: The foundation supports incremental addition of features

### API Design Standards

All API endpoints in the ChronoCone v0 system must be defined using the OpenAPI Specification (OAS) version 3.2.0. The OpenAPI Specification provides a standard, language-agnostic interface to HTTP APIs that allows both humans and computers to discover and understand the capabilities of the service.

#### OpenAPI Specification Requirements

1. **Endpoint Definition**: All API endpoints must be documented using the OpenAPI Specification. This includes defining paths, operations, parameters, request bodies, and responses according to the specification.

2. **Schema Definitions**: Data structures exchanged between clients and servers must be defined using the OpenAPI Schema Object, which is a superset of JSON Schema. See [Schema Object](../../resources/apis/openapi/schema_object.md) for schema definition details.

3. **Data Exchange Format**: All data exchange between API clients and servers must use JSON (application/json) format. YAML format must not be used for data exchange, although it may be used for authoring OpenAPI documents during development.

4. **Media Type Specification**: All request and response bodies must specify `application/json` as the media type in the OpenAPI Media Type Object. See [Media Type Object](../../resources/apis/openapi/media_type_object.md) for media type definition details.

#### OpenAPI Documentation References

The following OpenAPI documentation provides detailed guidance for implementing API endpoints:

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

For a complete table of contents of OpenAPI documentation, see [OpenAPI TOC](../../resources/apis/openapi/toc.md).

### System Components

The v0 system consists of seven main components:

1. **Core Server**: Handles system startup and provides service discovery for other services in the v0 system. In future versions, this will be replaced with a cloud-based server or calls to an AWS API Gateway that will feed into Lambda functions or an ECS-based server. See [`resources/apis/fastapi/toc.md`](../../resources/apis/fastapi/toc.md) for FastAPI API documentation.

2. **Metrics Service**: Provides an interface to an underlying PostgreSQL database server plus an object store used to store metrics and associated metadata.

3. **MCP Server (Business Analysis)**: Intended for use by business analysts when using an LLM or agentic system to ask questions about forecasts.

4. **MCP Server (Planning)**: Intended for use by planners who are creating or adjusting planning models.

5. **Web UI**: A browser-based user interface for interacting with the ChronoCone system.

6. **Plugin UI**: IDE-integrated user interfaces (IntelliJ and VS Code plugins) for planners working on transformers, shifters, and filters.

7. **External API Simulator**: Serves canned data for use in demos.

## Part 3: Component Requirements

### 3.1 Core Server

**Purpose**: The "core" server is a centralized service that will be responsible for system startup and provide a mechanism for discovering endpoints for services, queues or other functionality.

**Interfaces**:

These are the major top level interfaces that the core server will provide:

1. /endpoint - When supplied with an endpoint type the core service will reply with a datatype that is appropriate for the endpoint type requested.
2. /admin - An interface that is primarily intended to be used by administrators of the system for set up, configuration or other administrative tasks.

**Autostart Capability**:

The Core Server supports an autostart option that controls whether dependent services (such as the Metrics Service) are automatically started when the Core Server starts. When autostart is enabled, the Core Server launches all required dependent services. When disabled, services must be started independently, which facilitates debugging and development scenarios.

**Future Migration Path**:

It is likely that the core server will eventually be replaced by configuration or an API Gateway instance.

### 3.2 Metric Service

**Purpose**: The Metrics Service is responsible for storing, maintaining and updating metrics that are used by the ChronoCone system, as well as the metadata associated with them (e.g., categories, name spaces).

**Responsibilities**:

1. Retrieve metrics, including actuals and forecasts
2. Store and maintain metrics and associated metadata
3. Implement the compute engine that will update metrics when new data arrives. In order to accommodate a wide variety of usage scenarios, the actual triggering of the compute engine will be done when data arrives, so there is no 'update a metric' interface.

**Interfaces**:

- `/metrics` - Retrieves metrics, including actuals and forecasts. Since metrics can be large it is likely that the response will provide a reference to the metric data in an object store. While this will be the local file system in the v0 prototype, it is likely that a more mature version will use mechanisms like presigned S3 URLs.
- `/metadata` - Retrieves, creates, updates and deletes metadata associated with metrics.
- `/list` - Returns a list of metadata. Accepts an optional type filter to return metadata of a specific type, or returns all metadata when no type is specified.
- `/types` - Returns a list of all supported metadata types.

**Metadata Types**:

The Metrics Service manages the following metadata types, where the name field must be unique within each type:

- `namespace` - Isolation boundaries for taxonomies, metrics, and time series data
- `namespace_family` - Groups related namespaces under a common owner and type
- `category` - Classification dimensions or groupings that form hierarchies; used as the basis for metric dimensions
- `element` - Specific members within a category; leaf nodes used to tag data
- `metric` - Measurable quantities with defined dimensions (categories that index the metric)
- `transform` - Executable code that transforms input metrics into output metrics
- `shifter` - Functions that align metrics with different dimensional structures
- `filter` - Functions that split a single metric into multiple metrics whose values sum to the input

**Data Stores**:

- PostgreSQL database
- Object store

### 3.3 MCP Server - Business Analysis

**Purpose**: The purpose of the business analysis mcp server is to facilitate questions asked about metrics by general ChronoCone users.  
Many of these questions will be about metric values, and it's important that LLMs or agents not be overwhelmed by the volume of data that is available since that could result in context rot.

**Target Users**: Business analysts using LLM or agentic systems

**Responsibilities**:

1. Answer questions about:
   1. What metrics mean,
   2. How they're related to other metrics,
   3. What particular forecasts are,
   4. Why something happened (e.g., an unexpected increase or drop), or
   5. What changed when something unexpected occurred.
2. Control and monitor execution of plans.
3. Monitor business health.

**Interfaces**:

The Business Analysis MCP Server implements three MCP interfaces: Tools, Resources, and Prompts.

#### Tools Interface

The Tools interface provides callable functions that AI assistants can invoke to retrieve and analyze planning data.

**update**: Takes a namespace family and namespace as inputs. Updates a plan by invoking the calculation engine to pull new data for inputs and creating a new version of a document for the document types in the namespace family/namespace. This tool enables business analysts to refresh their planning data to reflect the latest available information.

**read**: Takes a namespace family and namespace as inputs, with an optional list of metric names. Returns the forecast and actuals values for all metrics if no metrics are specified, or for the given metrics if they are specified. Each metric response includes:

- Dimensions
- Cell value type(s)
- Time grain
- Length
- Predecessor metrics (metrics that provide inputs)
- Successor metrics (metrics that consume this metric's outputs)

**lineage**: Takes either a metric name or a date/time for a particular cell in a metric. Returns a trace of the source(s) of the data for that metric and/or cell. This tool enables users to understand how a particular value was derived, which is essential for debugging and explaining planning results.

**diff**: Takes a metric and two version numbers. Returns the differences between the two versions, including:

- Differences in actuals values
- Differences in forecasts values
- Differences in transformers used
- Differences in shifters used
- Differences in filters used

This tool helps users understand what changed between two planning cycles or document versions.

**metric_names**: Takes a namespace family and namespace as inputs. Returns the names of the _published_ metrics in that family/namespace. This tool helps users discover what metrics are available for analysis.

#### Resources Interface

The Resources interface provides contextual information that helps users understand the planning domain.

**Metric Form and Purpose**: For each metric, the server provides information about:

- What the metric measures
- How it is calculated (at a conceptual level)
- What its dimensions represent
- How it relates to other metrics in the planning hierarchy

This information helps business analysts understand the semantics of the data they are exploring.

**Domain Definition and Procedures**: When defined through the administration endpoint of the Metrics Service, the server provides descriptions and terms used in a namespace family. These definitions are stored in the object store and presented as a resource that explains the business context and procedures specific to that planning domain.

#### Prompts Interface

The Prompts interface provides PDD (Prompt Driven Development) scripts that guide users through common analysis workflows. These scripts offer structured approaches to:

- Exploring a new planning domain
- Investigating anomalies or unexpected changes
- Comparing forecasts across time periods
- Understanding metric relationships and dependencies

### 3.4 MCP Server - Planning

**Purpose**: The purpose of the planning MCP server is to help human planners leverage AI to construct and adjust plans. The ChronoCone engine provides a deterministic, rule-based calculation layer, and the MCP server acts as the interface where four primary personas (Planners, Power-planners, Analysts, and System Integrators) use agent-chat interfaces to govern these models safely.

**Target Users**: Planners, Power-planners, Analysts, and System Integrators utilizing human-initiated AI agent interactions.

**Responsibilities**:

1. Act as the primary interface for governed autonomy, whereby humans initiate agent interactions, and agents use the MCP to read/write the ontology which governs the planning models.
2. Generate programmatic code from PDD (Prompt Driven Development) style input.
3. Generate test data for planning models.
4. Execute tests and provide feedback to the planner and help in fixing errors.

**Interfaces**:

The Planning MCP Server implements three MCP interfaces: Tools, Resources, and Prompts.

#### Tools Interface

The Tools interface provides callable functions that AI assistants can invoke to create and manage planning models.

**metric_names**: Takes a namespace family and namespace as inputs. Returns the names of both the published and unpublished metrics in that family/namespace, with a tag to indicate their current status. This tool helps planners see what metrics exist and which ones have pending changes.

**publish**: Takes a namespace family and namespace as inputs. Publishes any unpublished metrics in that family/namespace. This action transitions metrics from the unpublished state to the published state, making them active for production calculations. Publication is a critical action that affects business decisions, so it should be used with appropriate care and permissions.

**push**: Takes a namespace family, namespace, and text content as inputs. The text content is a transformer, shifter, or filter definition that is pushed to that namespace as an unpublished artifact. This tool enables planners to save their work to the object store without immediately affecting production calculations. The pushed content can later be published using the publish tool.

#### Resources Interface

The Resources interface provides contextual information that helps planners understand the planning domain and existing models.

**Metric Form and Purpose**: For each metric, the server provides information about:

- What the metric measures
- How it is calculated (at a conceptual level)
- What its dimensions represent
- How it relates to other metrics in the planning hierarchy

This information helps planners understand the existing planning models and design new metrics that integrate properly with the existing structure.

**Domain Definition and Procedures**: When defined through the administration endpoint of the Metrics Service, the server provides descriptions and terms used in a namespace family. These definitions are stored in the object store and presented as a resource that explains the business context and procedures specific to that planning domain. This context is essential for planners who need to ensure their models align with business requirements.

#### Prompts Interface

The Prompts interface provides PDD (Prompt Driven Development) scripts that guide planners through common development workflows. These scripts offer structured approaches to:

- Creating a new transformer for a metric
- Creating a shifter to align dimensional structures
- Creating a filter to split a metric into components
- Generating test data for validation
- Debugging and fixing errors in existing functions
- Publishing changes to a namespace

### 3.5 Web UI

**Purpose**: The Web UI is the primary mechanism for providing a visual display of metrics through a browser-based interface.

**Target Users**: End users interacting with the ChronoCone system

**Responsibilities**:

1. Display metrics.
2. Display and update metadata.

**Key Features**:

- **Navigation Structure**: A top-level navigation bar provides access to three main sections: Home (landing page with welcome message), Definitions (metadata management interface), and Metrics (metrics display).
- **Metadata Management**: The Definitions page allows users to view metadata organized by type in tabular format, with the ability to add new metadata definitions through an inline add row interface.
- **Responsive Layout**: The navigation bar remains visible across all pages for consistent navigation.

#### Metrics Display

The Metrics page provides a comprehensive interface for viewing and exploring time series data. The display follows specific visual and interaction patterns designed to make planning data accessible and understandable.

**Time Series Layout**:

Each metric is displayed as a horizontal time series arranged in chronological order. The display divides into two distinct regions:

- **Actuals Region**: The left portion of the time series displays actual values, which represent historical measurements. Since actuals are always in the past, they appear on the left side of the display.
- **Forecast Region**: The right portion of the time series displays forecast values, which represent predicted future values. Forecasts appear to the right of the actuals.

A vertical dividing line separates the actuals region from the forecast region, providing a clear visual distinction between historical measurements and predictions.

**Time Labels**:

Each cell in the metric time series displays an associated time label positioned just above the cell. These labels appear in small, greyed out, italicized characters to provide temporal context without dominating the visual display.

The format of time labels depends on the metric's time grain and user preferences:

- **Weekly Grain**: The default format displays the week number (e.g., "Week 12"). The alternative format displays the date in YYYY-MM-DD format.
- **Daily Grain**: The default format displays the day number within the year (e.g., "Day 85"). The alternative format displays the date in YYYY-MM-DD format.

Users can configure the default time label format for each metric through an administrative endpoint in the Metrics Service.

**Vertical Arrangement**:

Multiple metrics are displayed vertically on the page in directed acyclic graph (DAG) order. This arrangement places input metrics at the top of the display and output metrics at the bottom, reflecting the flow of data through the calculation pipeline. This ordering helps users understand dependencies between metrics and trace how values propagate through the system.

**Namespace Selection**:

At the top of the Metrics page, dropdown controls allow users to select the namespace family and namespace combination they wish to view. This selection determines which metrics are displayed and provides access to different planning contexts (e.g., development, production, personal namespaces).

**Published vs. Unpublished Metrics**:

A dropdown control allows users to choose whether to view published metrics only or to include unpublished metrics as well. When unpublished metrics are included, they appear directly below their corresponding published metric(s), enabling comparison between the current production state and pending changes.

Unpublished metrics are visually distinguished from published metrics using:

- **Color differentiation**: A distinct color scheme indicates unpublished status.
- **Accessibility consideration**: An additional visual indicator (such as pattern, icon, or border style) ensures that color-blind users can distinguish between published and unpublished metrics.

**Metric Detail View**:

When a user clicks on a metric, the system displays a detail view with additional information:

- **All Users**: A graph showing actual values versus forecast values in the past (demonstrating forecast accuracy) and forecast values in the future. This visualization helps users understand both historical performance and predicted outcomes.

- **Users with Plan Modification Permission**: In addition to the graph, these users see panels displaying the definitions of forecast transformers and actuals transformers (if they exist). This information helps planners understand how metric values are calculated and identify opportunities for improvement.

**Administrative Configuration**:

The Metrics Service provides an administrative endpoint that allows authorized users to specify the default time label format for each metric. This configuration persists across sessions and applies to all users viewing that metric.

### 3.6 External API Simulator

**Purpose**: The External API Simulator provides simulated external data for demonstrations and testing. It serves canned data for use in demos when the system is not yet connected to real external systems, enabling the ChronoCone system to be demonstrated and tested with realistic data that mimics what real integrations would provide.

**Target Users**: Developers, testers, and demonstrators working with the ChronoCone system

**Responsibilities**:

1. Supply data for individual metrics through a standardized API endpoint
2. Look up metric type information from the database to ensure type-correct responses
3. Return configured data values when specified in a configuration file
4. Generate random data values when no configuration is provided
5. Support both actuals and forecast data retrieval

**Interfaces**:

- `/metric_data` - Retrieves simulated metric data. Accepts a GET request with parameters specifying the metric, time grain, time range, and data type (actuals or forecast). Returns a JSON list of values matching the types defined for the metric in the database.

**Endpoint Parameters**:

The `/metric_data` endpoint accepts the following query parameters:

| Parameter     | Description                                                                         |
| ------------- | ----------------------------------------------------------------------------------- |
| `metric_name` | The name of the metric for which data is being requested                            |
| `time_grain`  | The temporal interval for the data (e.g., hour, day, week, month)                   |
| `time_range`  | The time range for which data should be returned, specified as a start and end time |
| `data_type`   | Whether to return actuals or forecast values                                        |

**Response Format**:

The endpoint returns a JSON list of values. Each value in the list corresponds to a time point within the requested time range, ordered chronologically. The data type of each value matches the type defined for the metric in the database:

- **Integer metrics**: Values are JSON integers
- **Float metrics**: Values are JSON numbers with decimal precision
- **String metrics**: Values are JSON strings

**Configuration File**:

An optional configuration file may be used to specify data values for specific metrics. The configuration file supports:

- Specification of actuals values for defined time ranges
- Specification of forecast values for defined time ranges
- Combination of both actuals and forecast values for the same metric

When a configuration file specifies values for a requested metric and time range, those configured values are returned instead of randomly generated values. This allows demonstrators and testers to create predictable, reproducible scenarios.

**Random Value Generation**:

When no configuration file specifies values for a requested metric and time range, the simulator generates random values:

- **Integer and Float types**: Random numeric values in the range 100 to 1000 (inclusive)
- **String types**: Random alphanumeric strings of length 10

The random value generation ensures that demonstrations have realistic-looking data without requiring manual configuration for every metric.

**Design Considerations**:

The External API Simulator must balance several concerns:

- **Type Correctness**: Values returned must match the data types defined for the metric in the database. The simulator queries the Metrics Service to determine the correct type before generating or returning values.

- **Reproducibility**: When using a configuration file, the same request should produce the same response consistently. This supports testing scenarios where predictable results are essential.

- **Simplicity**: The simulator provides a straightforward interface that mimics what a real external API might provide, without implementing complex business logic or data transformations.

### 3.7 IDE Plugins

**Purpose**: The IDE plugins provide an integrated development environment for creating, editing, and testing transformers, shifters, and filters. These plugins bridge the gap between the metrics server and the developer's local workspace, enabling a workflow that combines AI-assisted development with traditional code editing.

**Target Users**: Developers and planners who create or modify transformers, shifters, and filters

**Supported Platforms**:

- IntelliJ IDEA (and compatible JetBrains IDEs)
- Visual Studio Code

**Core Workflow**:

The IDE plugins support a pull-modify-push workflow for calculation assets:

1. **Pull**: Users retrieve transformers, shifters, or filters from the metrics server into a local sandbox environment. The sandbox isolates development work from production calculations.

2. **Edit**: Users modify the assets using either traditional code editing or AI-assisted development through an integrated agentic interface.

3. **Test**: Users validate their changes using generated test data within the sandbox.

4. **Push**: Users push modified assets back to the metrics server as dirty assets. These dirty assets exist in storage but are excluded from production calculations until published.

**Asset Identification**:

When working with transformers, shifters, or filters, users must specify:

- **Asset Type**: Whether the asset is a transformer, shifter, or filter
- **Namespace Family**: The namespace family that owns the asset
- **Namespace**: The specific namespace within the family
- **Function Type** (transformers only): Whether the transformer produces actuals or forecasts
- **Name**: A unique identifier for the asset within the namespace family

The name must be unique within the namespace family to prevent ambiguity in calculation logic.

**Key Features**:

- **Agentic Interface**: A text-based conversational interface allows users to interact with an AI assistant for code generation, modification, and explanation. The assistant understands the ChronoCone domain model and can help users write correct transformers, shifters, and filters.

- **Traditional Code Editing**: Full support for manual code editing with syntax highlighting, code completion, and error detection appropriate for Python development.

- **Test Data Generation**: The plugins assist users in creating test data for validating their assets. Test data is captured alongside the asset to ensure reproducibility and to document expected behavior.

- **Prompt Capture**: Prompts used to generate or modify assets through the agentic interface are captured and stored with the asset. This creates an auditable record of how the asset was developed and enables future refinement.

- **PDD Script Integration**: The plugins support Prompt Driven Development (PDD) scripts that guide users through structured workflows for creating and modifying assets. PDD scripts provide step-by-step instructions that ensure consistency and completeness in asset development.

**Interfaces**:

- Integration with the Metrics Service API for pulling and pushing assets
- Local sandbox management for isolated testing
- AI assistant integration for agentic interactions
- Version control integration for tracking changes

**Data Captured with Assets**:

- The Python code implementing the transformer, shifter, or filter
- Test data used for validation
- Prompts used during development
- Metadata linking the asset to its namespace family, namespace, and related metrics

### 3.8 Plugin UI

**Purpose**: The Plugin UI provides the user interface for planners working on transformers, shifters, and filters through IDE plugins. Unlike the Web UI, which is browser-based, the Plugin UI is integrated directly into the developer's IDE (IntelliJ IDEA or Visual Studio Code), providing a seamless development experience for creating and modifying calculation logic.

**Target Users**: Planners and developers who create or modify transformers, shifters, and filters

**Supported Platforms**:

- IntelliJ IDEA (and compatible JetBrains IDEs) via the IntelliJ plugin
- Visual Studio Code via the VS Code extension

**Responsibilities**:

1. Provide an integrated interface for the pull-modify-push workflow
2. Display and edit transformer, shifter, and filter code
3. Provide access to the agentic interface for AI-assisted development
4. Display test results and validation feedback
5. Manage local sandbox state and synchronization with the Metrics Service

**Key Features**:

#### Namespace Connection

When a user opens the plugin in a supported IDE, the first interaction concerns namespace selection. The plugin displays a connection panel that prompts the user to connect to a namespace family and namespace combination. Examples include Retail/Dev, Retail/Prod, or Retail/My Personal Namespace. This connection establishes the context for all subsequent operations within the plugin.

If the user is already connected to a namespace, the connection panel displays the current connection status. A menu item allows the user to switch to a different namespace family or namespace at any time. This flexibility supports workflows where planners work across multiple environments, such as developing in a personal namespace before publishing to a shared development namespace.

#### Explorer Panel

Once connected, the plugin displays an explorer panel that provides navigation through the calculation assets in the connected namespace. The explorer panel contains three vertically arranged tabs, each with an expand/collapse control (shown as a `>` indicator) on the left side:

- **Metrics Tab**: Displays the metrics defined in the connected namespace, arranged according to the directed acyclic graph (DAG) that connects them. Metrics flow from top to bottom, with upstream metrics (those that provide inputs to other metrics) appearing above downstream metrics (those that receive inputs). This visual arrangement helps planners understand dependencies and the flow of calculations.

- **Shifters Tab**: Displays the shifters defined in the connected namespace. Each shifter entry shows its name and a brief description of its dimensional transformation.

- **Filters Tab**: Displays the filters defined in the connected namespace. Each filter entry shows its name and a brief description of how it splits its input metric.

The vertical tab arrangement with expand/collapse controls allows users to focus on one category of assets while minimizing visual clutter from the others. Users can expand multiple tabs simultaneously to see relationships across asset types.

#### Chat Window

At the bottom of the plugin interface, a chat window provides interaction with AI agents. This chat interface supports the agentic workflow described in section 3.7, allowing users to request code generation, modification, and explanation through natural language prompts. The chat window remains accessible regardless of which tab is active in the explorer panel, providing consistent access to AI assistance.

#### Metric Detail View

When a user selects a metric from the Metrics tab, the plugin opens a detail view to the right of the explorer panel. This detail view displays the transformers associated with the selected metric:

- **Forecast Transformer**: Every metric has at least one forecast transformer, which produces predicted values for future time periods. The forecast transformer appears first in the detail view.

- **Actuals Transformer**: If the metric receives data from external systems, an actuals transformer may also exist. This transformer appears below the forecast transformer when present.

The transformer display shows the code with greyed-out annotations indicating the dimensions and value types in each cell. These annotations help users understand the structure of the data flowing through the transformer without cluttering the actual code. Users can edit the transformer code directly in this view and run tests to validate their changes.

#### Shifter and Filter Detail View

When a user selects a shifter or filter from its respective tab, the plugin opens a detail view to the right of the explorer panel. This view displays the Python code implementing the selected function. Users can edit the code and run tests to validate that the shifter or filter produces the expected outputs.

#### Testing Interface

Each transformer, shifter, and filter function displays a test icon (an Erlenmeyer flask shape) that opens a test panel when clicked. The test panel appears to the right of the function code and contains a table of input/output test data.

**Test Table Structure**: The test table displays rows of test cases, with columns for input values and expected output values. Each row represents a single test case that can be executed against the function.

**Running Tests**: A "run" icon (depicted as a running figure) appears next to the test table. When the user clicks this icon, the plugin executes all test cases by passing the input values from each row into the function and comparing the actual output to the expected output.

**Test Results Display**: After execution, each row displays visual feedback:

- **Passed Tests**: Rows where the actual output matches the expected output are highlighted in green. The run icon displays in green with an annotation showing "Passed".

- **Failed Tests**: Rows where the actual output does not match the expected output are highlighted in red. The actual output value is displayed in the output column alongside the expected value. The run icon displays in red with an annotation showing "Failed - X", where X is the number of failed test cases.

**Managing Test Data**: Users can modify the test table through the following actions:

- **Add Rows**: A plus icon at the bottom of the table allows users to add new test case rows.

- **Delete Rows**: A trash icon at the end of each row allows users to delete that test case.

- **Edit Values**: Users can click on any input or expected output cell to edit the value directly.

#### AI Agent Capabilities

AI agents interacting through the chat window can manipulate and generate test data. Agents can create new test cases, modify existing test data, and suggest expected outputs based on their understanding of the function's intended behavior. Agents can also create and modify the code for transformers, shifters, and filters, working iteratively with the user to refine both the code and its test coverage.

#### Test Data Persistence

Test data is persisted in the object store alongside the code for the transformer, shifter, or filter it tests. This co-location ensures that test data remains synchronized with the specific version of the function code it was designed to test. When a function is pulled from the object store, its associated test data is also retrieved. When a modified function is pushed back to the object store, the updated test data accompanies it.

#### Sandboxed Execution Environment

Execution of transformer, shifter, and filter functions occurs in a controlled, sandboxed environment that is independent of the software installed on the user's machine. This isolation ensures consistent behavior across different development environments and prevents dependencies on local machine configurations from affecting test results. The sandboxed environment provides a standardized runtime with known versions of Python and required libraries, ensuring that functions behave the same way during testing as they will in production calculations.

**Interfaces**:

- Integration with the Metrics Service API for pulling and pushing assets
- Local sandbox management for isolated testing
- AI assistant integration for agentic interactions
- Version control integration for tracking changes

## Transformers, Shifters, and Filters

ChronoCone uses three types of Python functions to define and manipulate metrics: transformers, shifters, and filters. These functions are tracked in the database rather than defined through decorators, enabling dynamic discovery and management of calculation logic.

### Transformers

A **transformer** is a Python function that produces metric values. Each metric has at least one transformer associated with it:

- **Forecast transformer**: Produces predicted values for future time periods. Every metric must have a forecast transformer.
- **Actuals transformer**: Produces historical values representing ground truth measurements. Only metrics that receive data from external systems typically have actuals transformers.

The database tracks the location of transformer functions and their correspondence to namespace families, namespaces, and metric names. This approach replaces decorator-based definitions, allowing the system to discover and invoke transformers dynamically.

#### Transformer Execution Model

Transformers are called by the calculation engine with inputs and outputs that correspond to elements of a time series. For a transformer to operate correctly:

1. **Dimension Matching**: Input and output metrics must have exactly the same dimensions. If dimensions differ, shifters (described below) are used to align them.

2. **Time Grain and Forecast Length Matching**: Input metrics must share the same time grain and forecast length as the output metric.

3. **Position Correspondence**: The calculation engine iterates over time series elements where input values have matching dimension values and positions in their respective time series.

#### Transformer Registration

Transformers are registered in the database through the [`transform`](memory-bank/terms.md:16) table, which stores:

- The storage location of the code
- The specific code location (module and function name)
- Input and output metric relationships
- Validity period for version control

### Shifters

A **shifter** is a Python function that transforms a metric with one set of dimensions into a metric with a different set of dimensions. Shifters enable transformers to work with inputs that have incompatible dimensional structures.

#### When Shifters Are Needed

Consider a metric for ordered units defined by Product and Vendor dimensions. If this metric serves as input to a transformer whose output metric is defined only by Product, the dimensions don't match. A shifter can sum the ordered units across the Vendor dimension, producing a temporary in-memory metric defined only by Product.

#### Shifter Operations

Common shifter operations include:

- **Aggregation**: Summing or averaging across one or more dimensions (e.g., summing Vendor to get Product-only totals)
- **Selection**: Selecting a specific element from a dimension (e.g., selecting a specific Vendor)
- **Expansion**: Replicating values across a new dimension when needed for calculation

Shifters are tracked in the database similar to transformers, with their input and output dimension specifications recorded for automatic invocation by the calculation engine.

### Filters

A **filter** is a Python function that splits a single input metric into multiple output metrics. The key property of filters is that the values in the input metric equal the sum of the corresponding values in the output metrics.

#### Filter Use Cases

Filters are useful for:

- **Segmentation**: Splitting a total into segments (e.g., splitting total demand into demand by region)
- **Allocation**: Distributing an aggregate value across categories based on proportions
- **Scenario Analysis**: Creating multiple scenario variants from a base forecast

#### Filter Registration

Like transformers and shifters, filters are tracked in the database with their input and output metric relationships, enabling the calculation engine to invoke them automatically when needed.

### Function Definition Structure

1. **Function Registration**: Functions are registered in the database with their module path and function name, along with metadata about their inputs, outputs, and dimensional requirements.

2. **Input Parameters**: Input parameters represent metrics used for input. The calculation engine provides values based on the registered input metric specifications.

3. **Dimension Handling**: When input metrics have different dimensions than the output, the calculation engine automatically invokes appropriate shifters to align the data before calling the transformer.

4. **Forecast vs Actuals**: Every metric has two components:
   - **Forecasts**: Associated with date/times in the future
   - **Actuals**: Data collected, associated with date/times in the past

   The transformer type (forecast or actuals) is specified during registration.

5. **Dimension and Time Parameters**: Input metrics may be followed by parameters matching the metric's dimension names/types, time grain, and the date/time of the time series entry being calculated.

### Actuals from External Systems

Actuals often come from external systems via connectors. Actuals transformers can be defined to fetch data from connectors:

```python
def inventory_actuals(product: str, when: datetime) -> int:
    return InventoryConnector.fetch(product, when)
```

**Efficiency Considerations**: The calculation engine should lazily evaluate connector calls, collecting all product/when combinations across metric iterations before batching calls to the connector for efficient data retrieval.

## Function Lifecycle

Transformers, shifters, and filters follow a defined lifecycle that governs how they are created, modified, and made available for production calculations. This lifecycle ensures that changes are controlled, tested, and auditable while preserving the ability to understand how calculations were performed in the past.

### Function Storage

All transformer, shifter, and filter functions are stored as Python code in an object store. The object store is accessed through a connector abstraction. In v0, this connector provides access to the local filesystem. In future versions, the connector will support cloud-based object stores such as AWS S3.

Functions are identified by name within a namespace family. A function name must be unique within its namespace family to prevent ambiguity during plan calculation. However, the same name may appear in different namespace families. For example, a function named "calculate_labor" might exist in both the Retail namespace family and the Warehouse Labor namespace family, and could also exist in both Prod and Dev namespaces within the same family.

### Three States

Each transformer, shifter, or filter exists in one of three states at any given time:

1. **Published**: The function is stored in the object store and is actively used by the calculation engine to compute plans. Published functions are the authoritative version for their namespace family and namespace.

2. **Unpublished**: The function exists in the object store but is not used for plan calculations. Unpublished functions represent new or modified code that has been pushed from a local workspace but has not yet been approved for production use.

3. **Local Workspace**: The function exists in a local development environment, typically a Docker container running on the user's local machine. Functions in this state are being created, modified, or tested through IDE plugins.

### State Transitions

Functions transition between states through specific actions:

**Local Workspace to Unpublished**: When a user completes development work in their local workspace, they push the function to the object store. This creates a new entry in the database marked as unpublished. The function code is stored in the object store, but the calculation engine does not use it.

**Unpublished to Published**: When a user publishes changes to a namespace family and namespace, the system updates the database to mark the function as published. Once published, the function becomes active and participates in production calculations.

**Published to Local Workspace**: When a user wants to modify an existing function, they pull a copy from the object store to their local workspace. The published version remains unchanged and continues to be used for calculations.

### Versioning and Immutability

When a function is modified, the system never overwrites the existing code in the object store. Instead, a new object is created with a new database entry. This approach preserves the complete history of function versions, enabling the system to:

- Answer questions about how calculations were performed for documents generated in the past
- Maintain an audit trail of all changes to calculation logic
- Support rollback to previous versions if needed

Each version of a function has an associated validity period recorded in the database. The calculation engine uses the appropriate version based on when a document was created or when a calculation was performed.

### Namespace Scoping

Publication operates within the namespace isolation model:

- **Namespace Family/Namespace**: Functions are published to a specific namespace family and namespace combination. This scoping ensures that changes affect only the intended calculation context.

- **Uniqueness Within Family**: Function names must be unique within a namespace family. This prevents ambiguity when the calculation engine resolves which function to use for a given metric.

- **Cross-Family Duplication**: The same function name may exist in different namespace families. For example, a "calculate_demand" function might exist in both Retail and Manufacturing namespace families with different implementations appropriate to each domain.

### Development Workflow

The development workflow supports creating and modifying functions safely:

1. **Pull**: Users retrieve a function from the object store into their local workspace. For new functions, this step creates a skeleton in the local environment.

2. **Edit**: Users modify the function code using IDE plugins that provide syntax highlighting, code completion, and AI-assisted development.

3. **Test**: Users validate their changes using test data within the local workspace. The workspace provides an isolated environment that does not affect production calculations.

4. **Push**: Users push the modified function to the object store as an unpublished artifact. The function exists in storage but is excluded from production calculations.

5. **Publish**: Users with appropriate permissions publish the function, making it active for production calculations within the specified namespace family and namespace.

### Audit Trail

The system maintains a record of all lifecycle events, including:

- When functions were created, modified, or deprecated
- Who made each change
- What prompts or instructions were used during development (for AI-assisted creation)
- When functions were published and to which namespace family and namespace
- Which version of a function was used for any given calculation
