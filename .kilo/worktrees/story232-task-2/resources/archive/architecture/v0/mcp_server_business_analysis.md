Directives for agentic or AI writers and editors:

1. You MUST leave this list of directives in place as the first item in the document and unaltered as changes are made to this document.
2. When editing this document you MUST use either the Technical Writer or Technical Editor roles. If you are not in one of these modes you MUST NOT edit this document and inform the user of the requirement.
3. This document contains annotations that are labelled as "Notes:" or "Questions:" and you MUST NOT edit these annotations. You should move them to keep them with the content they relate to.

--- Begin Document Content ---

# MCP Server - Business Analysis

## Purpose

The Business Analysis MCP Server facilitates questions asked about metrics by general ChronoCone users. This server acts as an intelligent intermediary between business users and the underlying data, translating natural language questions into the appropriate queries and presenting results in a form that users can understand and act upon.

The server is built using the Model Context Protocol (MCP), which provides a standardized way for AI assistants and agentic tools to interact with external systems. This protocol choice reflects a fundamental design principle: business users should be able to interact with the planning system using natural language rather than learning specialized query languages or navigating complex interfaces.

Many of the questions handled by this server will be about metric values, and this presents a significant challenge. The volume of data available in a planning system can be enormous—potentially millions of data points across multiple metrics, dimensions, and time periods. It is critically important that LLMs or agents not be overwhelmed by this volume of data, as that could result in context rot: the degradation of response quality that occurs when an AI system's context window is filled with irrelevant or excessive information. The Business Analysis MCP Server is designed to manage this data volume carefully, retrieving and presenting only the information that is relevant to the user's question.

## Target Users

Business analysts using LLM or agentic systems represent the primary users of this server. These users are typically domain experts who understand the business context of the planning data but may not have technical expertise in query languages or database operations. They need to explore data, understand trends, and investigate anomalies as part of their daily work.

The server supports these users by providing a conversational interface to the planning data. Rather than writing SQL queries or navigating through a traditional business intelligence tool, users can ask questions in natural language and receive answers that are tailored to their needs.

## Responsibilities

1. Answer questions about:
   1. What metrics mean,
   2. How they're related to other metrics,
   3. What particular forecasts are,
   4. Why something happened (e.g., an unexpected increase or drop), or
   5. What changed when something unexpected occurred.
2. Control and monitor execution of plans.
3. Monitor business health.
The Business Analysis MCP Server has two primary responsibilities that together provide comprehensive support for business analysis workflows.

**Answering basic questions about metrics** involves explaining what metrics mean, how they are related to other metrics, and what particular forecasts are for a given set of categories. These questions might include: "What does the 'perfect units' metric measure?", "How is revenue related to units sold?", or "What is the forecast for electronics in the West region next month?" The server retrieves the relevant metadata and data to provide accurate, contextual answers.

**Answering questions about changes and anomalies** involves explaining why something happened or what changed when something unexpected occurred. These questions might include: "Why did forecast accuracy drop last week?", "What changed between this month's forecast and last month's?", or "Which regions contributed most to the variance?" The server analyzes the data to identify the factors that explain the observed changes.

## Interfaces

The Business Analysis MCP Server implements three MCP interfaces: Tools, Resources, and Prompts. These interfaces follow the Model Context Protocol specification and are designed to support the question-answering workflows described above.

### Tools Interface

The Tools interface provides callable functions that AI assistants can invoke to retrieve and analyze planning data. Each tool is designed to return focused, relevant information rather than large volumes of raw data, helping to prevent context rot.

#### update

**Purpose**: Refreshes planning data to reflect the latest available information.

**Parameters**:

- `namespace_family` (string, required): The namespace family containing the target namespace
- `namespace` (string, required): The namespace to update

**Behavior**: Invokes the calculation engine to pull new data for inputs and creates a new version of a document for the document types in the specified namespace family/namespace. This tool enables business analysts to ensure they are working with the most current planning data.

**Returns**: Confirmation of the update operation, including the new document version identifier.

#### read

**Purpose**: Retrieves metric values with their associated metadata and relationships.

**Parameters**:

- `namespace_family` (string, required): The namespace family containing the target namespace
- `namespace` (string, required): The namespace to read from
- `metric_names` (array of strings, optional): Specific metrics to retrieve; if omitted, returns all published metrics

**Behavior**: Returns the forecast and actuals values for the specified metrics (or all metrics if none specified). Each metric response includes:

- **Dimensions**: The categories that index the metric
- **Cell value type(s)**: The data types of values in the metric
- **Time grain**: The temporal resolution (e.g., daily, weekly)
- **Length**: The number of time periods in the time series
- **Predecessor metrics**: Metrics that provide inputs to this metric
- **Successor metrics**: Metrics that consume this metric's outputs

**Returns**: A structured representation of the requested metrics with all associated metadata.

#### lineage

**Purpose**: Traces the source(s) of data for a metric or specific cell.

**Parameters**:

- `namespace_family` (string, required): The namespace family containing the target namespace
- `namespace` (string, required): The namespace containing the metric
- `metric_name` (string, required): The metric to trace
- `date_time` (string, optional): A specific date/time for a cell; if omitted, traces the entire metric

**Behavior**: Returns a trace of the source(s) of the data for the specified metric or cell. This includes the chain of transformers, shifters, and filters that were applied, along with the specific versions of each function used.

**Returns**: A lineage trace showing the complete derivation path from source data to the specified value(s).

#### diff

**Purpose**: Compares two versions of a metric to identify differences.

**Parameters**:

- `namespace_family` (string, required): The namespace family containing the target namespace
- `namespace` (string, required): The namespace containing the metric
- `metric_name` (string, required): The metric to compare
- `version1` (integer, required): The first version number
- `version2` (integer, required): The second version number

**Behavior**: Returns the differences between the two specified versions, including:

- Differences in actuals values
- Differences in forecasts values
- Differences in transformers used
- Differences in shifters used
- Differences in filters used

**Returns**: A structured diff report highlighting all changes between the two versions.

#### metric_names

**Purpose**: Lists the published metrics available in a namespace.

**Parameters**:

- `namespace_family` (string, required): The namespace family to query
- `namespace` (string, required): The namespace to query

**Behavior**: Returns only the names of _published_ metrics. Unpublished metrics are not included, as business analysts work with production data.

**Returns**: A list of metric names available for analysis in the specified namespace.

### Resources Interface

The Resources interface provides contextual information that helps users understand the planning domain. Resources are read-only and provide background context rather than actionable functions.

#### Metric Form and Purpose

For each metric in the system, the server provides a resource that describes:

- **What the metric measures**: A plain-language description of the business quantity
- **How it is calculated**: A conceptual explanation of the calculation logic
- **What its dimensions represent**: The meaning of each dimensional category
- **Relationships to other metrics**: How this metric fits into the broader planning hierarchy

This information helps business analysts understand the semantics of the data they are exploring, enabling them to ask more informed questions and interpret results correctly.

#### Domain Definition and Procedures

When defined through the administration endpoint of the Metrics Service, the server provides descriptions and terms used in a namespace family. These definitions are stored in the object store and presented as a resource that explains:

- Business terminology specific to the planning domain
- Procedures and conventions used in the planning process
- Organizational context for interpreting the data

This resource is essential for users who are new to a particular planning domain or who need to understand the business context behind the numbers.

### Prompts Interface

The Prompts interface provides PDD (Prompt Driven Development) scripts that guide users through common analysis workflows. These scripts offer structured approaches to typical business analysis tasks.

#### Available Prompts

**Exploring a New Planning Domain**: Guides users through discovering what metrics are available, understanding their relationships, and identifying key metrics for their analysis needs.

**Investigating Anomalies or Unexpected Changes**: Provides a systematic approach to understanding why a particular value changed unexpectedly, including tracing lineage and comparing versions.

**Comparing Forecasts Across Time Periods**: Helps users compare forecast accuracy over time and understand how predictions have evolved.

**Understanding Metric Relationships and Dependencies**: Guides users through exploring the directed acyclic graph of metric dependencies to understand how changes propagate through the planning system.

## Design Considerations

The design of the Business Analysis MCP Server must balance several competing concerns.

**Data volume management** is perhaps the most critical concern. The server must be judicious about what data it retrieves and how it presents that data to the AI assistant. This might involve summarizing data rather than returning raw values, filtering to relevant subsets before retrieval, or using aggregation to reduce the number of data points.

**Context preservation** is essential for multi-turn conversations. Users often ask follow-up questions that build on previous answers, and the server must maintain enough context to understand these references. However, maintaining too much context can lead to the same context rot problems that the server is designed to avoid.

**Accuracy and grounding** ensure that the server's responses are based on actual data rather than hallucinations. Every claim about metric values or relationships should be traceable to the underlying data sources. This grounding is essential for building user trust in the system.

## Related Documentation

- [v0 Architecture Overview](v0_architecture.md)
- [v0 Requirements - MCP Server Business Analysis Section](../requirements/v0_requirements/v0_requirements.md#33-mcp-server---business-analysis)
