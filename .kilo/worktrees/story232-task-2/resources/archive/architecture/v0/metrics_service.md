Directives for agentic or AI writers and editors:

1. You MUST leave this list of directives in place as the first item in the document and unaltered as changes are made to this document.
2. When editing this document you MUST use either the Technical Writer or Technical Editor roles. If you are not in one of these modes you MUST NOT edit this document and inform the user of the requirement.
3. This document contains annotations that are labelled as "Notes:" or "Questions:" and you MUST NOT edit these annotations. You should move them to keep them with the content they relate to.

--- Begin Document Content ---

# Metrics Service

## Purpose

The Metrics Service is the heart of ChronoCone's data management capabilities. It is responsible for storing, maintaining, and updating metrics that are used by the ChronoCone system, as well as the metadata associated with them (e.g., categories, namespaces). This service sits at the center of the architecture, providing the data foundation that all other components depend on.

The Metrics Service addresses a fundamental challenge in planning systems: how to manage large volumes of time series data while maintaining rich metadata about what that data means and how it relates to other data. Traditional approaches often separate these concerns into different systems, but ChronoCone integrates them to ensure consistency and enable sophisticated query patterns.

## Responsibilities

The Metrics Service has three primary responsibilities that together provide comprehensive data management for the planning system.

**Retrieving metrics, including actuals and forecasts** involves providing efficient access to time series data. Metrics can be large—potentially millions of data points across multiple dimensions and time periods. The service must provide this data efficiently while supporting various query patterns, from retrieving a single time series to exploring relationships across many metrics.

**Storing and maintaining metrics and associated metadata** encompasses the full lifecycle of data management. This includes creating new metrics, updating existing ones, and managing the metadata that gives meaning to the raw numbers. The metadata includes categories that define how metrics are dimensioned, namespaces that provide isolation boundaries, and the relationships between metrics.

**Implementing the compute engine that updates metrics when new data arrives** is perhaps the most distinctive responsibility of the Metrics Service. The compute engine uses transformers to calculate metric values, shifters to align metrics with different dimensional structures, and filters to split metrics into constituent parts. This event-driven approach—triggering computation when data arrives rather than requiring explicit update commands—accommodates a wide variety of usage scenarios, from batch updates to real-time data feeds.

## Interfaces

The Metrics Service exposes several interfaces that provide access to its capabilities. These interfaces are designed to be simple and consistent, following RESTful principles while accommodating the specific needs of time series data management.

### /metrics Interface

The `/metrics` interface retrieves metrics, including actuals and forecasts. Since metrics can be large, the response provides a reference to the metric data in an object store rather than returning the data directly. This approach allows the service to handle metrics of any size without running into response size limits or timeout issues.

In the v0 prototype, the object store is the local filesystem. In more mature versions, the service will likely use mechanisms like presigned S3 URLs, which provide secure, time-limited access to cloud storage without requiring the data to flow through the service itself.

### /metadata Interface

The `/metadata` interface supports synchronous HTTP requests for metadata operations against the PostgreSQL database. This interface provides the standard CRUD operations for managing the metadata that defines the structure and meaning of metrics.

### /list Interface

The `/list` interface accepts JSON input and returns a list of metadata. If the input specifies a metadata type, returns metadata of that type. If no type is specified, returns all metadata. This interface is designed for discovery and exploration, allowing users to understand what data is available in the system.

### /types Interface

The `/types` interface accepts any JSON input (which is ignored) and returns a list of all supported metadata types. This interface provides a simple way for clients to discover what types of metadata the system manages.

### /definitions Interface

The `/definitions` interface provides operations for managing transformer, shifter, and filter definitions within a specific namespace family and namespace combination. This interface supports the pull-modify-push workflow used by IDE plugins, enabling planners to retrieve, modify, and update calculation logic.

The interface addresses the push and pull requirements for transformers, shifters, and filters. When definitions are retrieved, they include an indication of whether each definition is published or unpublished. The interface also supports optional inclusion of test data alongside the function code.

#### GET /definitions

Retrieves transformer, shifter, or filter definitions for a specific namespace family and namespace combination.

**Query Parameters:**

| Parameter           | Type    | Description                                                                                     |
| ------------------- | ------- | ----------------------------------------------------------------------------------------------- |
| `namespace_family`  | string  | Required. The namespace family name                                                             |
| `namespace`         | string  | Required. The namespace name within the family                                                  |
| `definition_type`   | string  | Required. One of: `transformer`, `shifter`, `filter`                                            |
| `name`              | string  | Optional. Specific definition name to retrieve. If omitted, returns all definitions of the type |
| `include_dirty`     | boolean | Optional. If true, includes unpublished definitions. Default is false (published only)          |
| `include_test_data` | boolean | Optional. If true, includes test data with each definition. Default is false                    |

**Response:**

```json
{
  "namespace_family": "retail",
  "namespace": "dev",
  "definition_type": "transformer",
  "items": [
    {
      "name": "calculate_demand",
      "code_location": "transformers/calculate_demand.py",
      "code": "def calculate_demand(inputs):\n    # calculation logic\n    return result",
      "is_published": true,
      "test_data": [
        {
          "test_case_id": 1,
          "input_values": { "demand": 100 },
          "expected_output": { "result": 105 }
        }
      ]
    },
    {
      "name": "calculate_labor",
      "code_location": "transformers/calculate_labor.py",
      "code": "def calculate_labor(inputs):\n    # calculation logic\n    return result",
      "is_published": false,
      "test_data": []
    }
  ]
}
```

The `is_published` field indicates whether the definition is published (`is_dirty` is false in the database) or unpublished (`is_dirty` is true). When `include_test_data` is false, the `test_data` field is omitted from the response.

#### POST /definitions

Creates or updates a transformer, shifter, or filter definition. When updating an existing definition, the operation creates a new version rather than modifying the existing one, preserving immutability.

**Request:**

```json
{
  "namespace_family": "retail",
  "namespace": "dev",
  "definition_type": "transformer",
  "name": "calculate_demand",
  "action": "update",
  "data": {
    "code": "def calculate_demand(inputs):\n    # updated calculation logic\n    return result",
    "test_data": [
      {
        "test_case_id": 1,
        "input_values": { "demand": 100 },
        "expected_output": { "result": 110 }
      }
    ]
  }
}
```

**Response:**

```json
{
  "namespace_family": "retail",
  "namespace": "dev",
  "definition_type": "transformer",
  "name": "calculate_demand",
  "status": "updated",
  "is_published": false,
  "code_location": "transformers/calculate_demand_v2.py"
}
```

When a definition is created or updated through this endpoint, it is stored with `is_dirty` set to `true`, meaning it is unpublished. The definition must be published through the `/admin` endpoint before it will be used in production calculations.

### /admin Interface

The `/admin` interface provides administrative operations for the Metrics Service. This interface is intended for users with administrative privileges and supports operations that affect the overall state of the system.

#### POST /admin

Accepts administrative operations through a unified endpoint. Each operation is specified by an `action` field in the request body.

**Publish Action**

Publishes all unpublished (dirty) transformers, shifters, and filters for a specific namespace family and namespace combination. This operation makes the unpublished definitions available for production calculations by setting their `is_dirty` flag to `false`.

**Request:**

```json
{
  "action": "publish",
  "namespace_family": "retail",
  "namespace": "dev"
}
```

**Response:**

```json
{
  "action": "publish",
  "namespace_family": "retail",
  "namespace": "dev",
  "status": "completed",
  "published": {
    "transformers": ["calculate_demand", "calculate_labor"],
    "shifters": ["aggregate_region"],
    "filters": []
  }
}
```

The response lists all definitions that were transitioned from unpublished to published state. If no unpublished definitions exist for the specified namespace family and namespace combination, the response returns empty lists.

**Error Conditions:**

- If the namespace family or namespace does not exist, the response returns a 404 error.
- If the user lacks administrative privileges for the namespace, the response returns a 403 error.

## Metadata Types

The metadata endpoint manages several types of metadata, each serving a specific purpose in the overall data model. Within each type, the `name` field must be unique, ensuring that metadata can be unambiguously referenced.

| Type               | Description                                                                                                                                          |
| ------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| `namespace`        | Isolation boundaries for taxonomies, metrics, and time series data. Each namespace has a type that determines its capabilities through policy rules. |
| `namespace_family` | Groups related namespaces under a common owner and type (application-level, personal, or team-based).                                                |
| `category`         | Classification dimensions or groupings that form hierarchies; used as the basis for metric dimensions.                                               |
| `element`          | Specific members within a category; the leaf nodes used to tag data or define dimensions.                                                            |
| `metric`           | Measurable quantities with defined dimensions (categories that index the metric), referencing a concept that provides semantic meaning.              |
| `transform`        | Executable code that transforms input metrics into output metrics, stored in a storage location and versioned.                                       |
| `shifter`          | Functions that align metrics with different dimensional structures by aggregating, selecting, or expanding across dimensions.                        |
| `filter`           | Functions that split a single metric into multiple metrics whose values sum to the input.                                                            |

These metadata types work together to create a rich data model. Namespaces and namespace families provide isolation and organization. Categories and elements define the dimensional structure of metrics. Transforms, shifters, and filters define how metrics are calculated and manipulated.

### /list Request/Response

**Request:**

```json
{
  "type": "namespace"
}
```

When `type` is specified, returns metadata of that type. When omitted, returns all metadata.

**Response:**

```json
{
  "items": [
    {
      "type": "namespace",
      "name": "namespace-name",
      "namespace_type": "Production"
    }
  ]
}
```

### /types Request/Response

**Request:**

```json
{}
```

Any JSON input is accepted and ignored.

**Response:**

```json
{
  "types": [
    "namespace",
    "namespace_family",
    "category",
    "element",
    "metric",
    "transform",
    "shifter",
    "filter"
  ]
}
```

### /metadata GET Request/Response

**Request:**

```json
{
  "type": "namespace",
  "name": "optional-specific-name"
}
```

**Response:**

```json
{
  "type": "namespace",
  "items": [
    {
      "name": "namespace-name",
      "namespace_type": "Production"
    }
  ]
}
```

### /metadata POST Request/Response

**Request:**

```json
{
  "type": "namespace",
  "action": "create",
  "data": {
    "name": "new-namespace",
    "namespace_type": "Production"
  }
}
```

**Response:**

```json
{
  "type": "namespace",
  "status": "created",
  "item": {
    "name": "new-namespace",
    "namespace_type": "Production"
  }
}
```

## Data Stores

The Metrics Service uses two distinct data stores, each chosen for its specific strengths in supporting the service's responsibilities.

**PostgreSQL database** runs in a Docker container and stores metadata for metrics, including categories and namespaces. PostgreSQL was chosen for its robust support for complex queries, JSONB data types for flexible metadata storage, and native support for temporal range types that enable bitemporal data modeling. The Metrics Service uses asyncpg for async database access, which provides non-blocking operations essential for maintaining responsiveness under load.

**Object store** stores metric data files (actuals and forecasts). The v0 prototype uses the local filesystem for simplicity. Later versions may use cloud storage with presigned URLs, which would enable the service to scale independently of data storage and provide efficient direct access to large data files.

## Database Access

The Metrics Service uses asyncpg for asynchronous PostgreSQL database operations. This choice reflects the service's need for high concurrency and responsiveness. asyncpg provides several key benefits that make it well-suited for this use case.

Async/await support enables non-blocking database operations, allowing the service to handle many concurrent requests without tying up threads. Connection pooling provides efficient resource utilization by reusing database connections across requests. Native support for PostgreSQL types including JSONB and tstzrange enables the service to work directly with the rich data types that PostgreSQL provides. Prepared statements enable query optimization by allowing the database to cache query plans.

### Connection Configuration

Database connections are configured via environment variables. This approach allows the service to be deployed in different environments without code changes, following the twelve-factor app methodology.

- `METRICS_DB_HOST` - PostgreSQL host (default: localhost)
- `METRICS_DB_PORT` - PostgreSQL port (default: 5432)
- `METRICS_DB_NAME` - Database name
- `METRICS_DB_USER` - Database user
- `METRICS_DB_PASSWORD` - Database password

### Connection Pool

The service maintains an asyncpg connection pool that is initialized at startup. This pool manages a set of database connections that can be reused across requests, avoiding the overhead of establishing a new connection for each request.

```python
pool = await asyncpg.create_pool(
    host=host,
    port=port,
    database=database,
    user=user,
    password=password,
    min_size=5,
    max_size=20
)
```

The pool is configured with a minimum of 5 connections and a maximum of 20. The minimum ensures that connections are available immediately when requests arrive, while the maximum prevents the service from overwhelming the database with too many concurrent connections.

## Pydantic Models

The Metrics Service uses Pydantic models for data validation and serialization. These models map to the database schema and support the bitemporal data model with validity ranges. Pydantic provides automatic validation, serialization, and documentation generation, which reduces boilerplate code and helps catch errors early in the request processing pipeline.

### Base Models

```python
from datetime import datetime
from typing import Optional, Dict, Any
from pydantic import BaseModel, Field

class MetadataBase(BaseModel):
    """Base model for all metadata types with bitemporal support."""
    validity_start: datetime = Field(..., description="Start of validity period")
    validity_end: Optional[datetime] = Field(None, description="End of validity period (null if current)")
    insertion_metadata: Dict[str, Any] = Field(default_factory=dict, description="Metadata about the insertion")

class MetadataResponse(BaseModel):
    """Base response model for metadata operations."""
    type: str = Field(..., description="Metadata type")
    items: list = Field(default_factory=list, description="List of metadata items")
```

The `MetadataBase` model provides the foundation for all metadata types, including the bitemporal fields that track when data is valid and when it was inserted. This bitemporal model is essential for understanding how data has changed over time and for reproducing historical calculations.

### Namespace Models

```python
class Namespace(BaseModel):
    """Represents a namespace for isolating taxonomies, metrics, and time series data."""
    ns_id: Optional[int] = Field(None, description="Unique namespace identifier")
    owner_subject_id: int = Field(..., description="ID of the subject that owns this namespace")
    ns_fam_id: int = Field(..., description="Namespace family ID")
    ns_type_id: int = Field(..., description="Namespace type ID (1-8)")
    name: str = Field(..., max_length=64, description="Unique name within the family")
    tooltip: str = Field(..., max_length=255, description="Short description")
    description: str = Field(..., description="Full description")
    insertion_metadata: Dict[str, Any] = Field(default_factory=dict)

class NamespaceFamily(BaseModel):
    """Groups related namespaces under a common owner and type."""
    ns_fam_id: Optional[int] = Field(None, description="Unique family identifier")
    customer_subject_id: int = Field(..., description="Customer subject ID")
    owner_subject_id: int = Field(..., description="Owner subject ID")
    ns_fam_type_id: int = Field(..., description="Family type (1=application, 2=personal, 3=team)")
    name: str = Field(..., max_length=64, description="Unique name within customer")
    tooltip: str = Field(..., max_length=255, description="Short description")
    description: str = Field(..., description="Full description")
    insertion_metadata: Dict[str, Any] = Field(default_factory=dict)
```

Namespaces provide isolation boundaries that allow different teams, organizations, or environments to maintain their own data within the same ChronoCone instance. Namespace families group related namespaces together, typically representing a customer or organization.

### Taxonomy Models

```python
class Category(BaseModel):
    """Classification dimensions or groupings that form hierarchies."""
    category_id: Optional[int] = Field(None, description="Unique category identifier")
    ns_id: int = Field(..., description="Namespace ID")
    taxonomy_parent_id: Optional[int] = Field(None, description="Parent element ID if this is a sub-category")
    name: str = Field(..., max_length=64, description="Category name (case-insensitive, unique within namespace)")
    tooltip_description: str = Field(..., max_length=255, description="Short description")
    long_description: str = Field(..., description="Full description")
    insertion_metadata: Dict[str, Any] = Field(default_factory=dict)

class Element(BaseModel):
    """Specific members within a category; leaf nodes used to tag data."""
    element_id: Optional[int] = Field(None, description="Unique element identifier")
    category_id: int = Field(..., description="Parent category ID")
    name: str = Field(..., max_length=64, description="Element name (case-insensitive, unique within category)")
    tooltip_description: str = Field(..., max_length=255, description="Short description")
    long_description: str = Field(..., description="Full description")
    insertion_metadata: Dict[str, Any] = Field(default_factory=dict)

class CategoryAttribute(BaseModel):
    """Attributes associated with a category."""
    category_id: int = Field(..., description="Category ID")
    attr_owner_ns_id: int = Field(..., description="Namespace ID that owns this attribute")
    name: str = Field(..., max_length=64, description="Attribute name")
    value: Dict[str, Any] = Field(..., description="Attribute value as JSONB")
    insertion_metadata: Dict[str, Any] = Field(default_factory=dict)

class ElementAttribute(BaseModel):
    """Attributes associated with an element."""
    element_id: int = Field(..., description="Element ID")
    attr_owner_ns_id: int = Field(..., description="Namespace ID that owns this attribute")
    name: str = Field(..., max_length=64, description="Attribute name")
    value: Dict[str, Any] = Field(..., description="Attribute value as JSONB")
    insertion_metadata: Dict[str, Any] = Field(default_factory=dict)
```

Categories and elements form the taxonomy that defines how metrics are dimensioned. A category represents a dimension (such as region or product category), and elements are the specific values within that dimension (such as "West" or "Electronics"). Attributes allow additional metadata to be attached to categories and elements.

### Metric Models

```python
class Metric(BaseModel):
    """Measurable quantities with defined dimensions."""
    metric_id: Optional[int] = Field(None, description="Unique metric identifier")
    concept_id: int = Field(..., description="Concept element ID that defines what is being measured")
    insertion_metadata: Dict[str, Any] = Field(default_factory=dict)

class MetricDimension(BaseModel):
    """Dimensions associated with a metric."""
    metric_id: int = Field(..., description="Metric ID")
    dimension_id: int = Field(..., description="Category ID serving as dimension")
    dimension_order: int = Field(..., description="Order of the dimension")
    insertion_metadata: Dict[str, Any] = Field(default_factory=dict)
```

Metrics represent the measurable quantities that the planning system tracks. Each metric is associated with a concept that defines what is being measured, and has dimensions that determine how the metric is indexed.

### Transform Models

```python
class Transform(BaseModel):
    """Executable code that transforms input metrics into output metrics."""
    transform_id: Optional[int] = Field(None, description="Unique transform identifier")
    storage_config_id: int = Field(..., description="Storage configuration ID for code location")
    code_location: str = Field(..., description="Path to transform code")
    insertion_metadata: Dict[str, Any] = Field(default_factory=dict)

class TransformInput(BaseModel):
    """Input metrics for a transform."""
    transform_id: int = Field(..., description="Transform ID")
    metric_id: int = Field(..., description="Input metric ID")
    input_order: int = Field(..., description="Order of this input")

class TransformOutput(BaseModel):
    """Output metrics produced by a transform."""
    transform_id: int = Field(..., description="Transform ID")
    metric_id: int = Field(..., description="Output metric ID")
```

Transforms define the calculation logic that produces metric values. The transform model stores the location of the code, while the input and output models define the data flow relationships.

### Time Series Document Models

```python
class TSDocumentType(BaseModel):
    """Types of time series documents."""
    ts_doc_type_id: Optional[int] = Field(None, description="Unique document type identifier")
    ns_id: int = Field(..., description="Namespace ID")
    name: str = Field(..., max_length=64, description="Document type name")
    tooltip_description: str = Field(..., max_length=255, description="Short description")
    long_description: str = Field(..., description="Full description")
    insertion_metadata: Dict[str, Any] = Field(default_factory=dict)

class TSDocumentRowType(BaseModel):
    """Row types within a time series document."""
    ts_doc_row_type_id: Optional[int] = Field(None, description="Unique row type identifier")
    ts_doc_type_id: int = Field(..., description="Parent document type ID")
    metric_id: int = Field(..., description="Metric for this row type")
    name: str = Field(..., max_length=64, description="Row type name")
    tooltip_description: str = Field(..., max_length=255, description="Short description")
    insertion_metadata: Dict[str, Any] = Field(default_factory=dict)
```

Time series document types define the structure of planning documents, including which metrics they contain and how those metrics are organized.

### Request/Response Models

```python
from typing import Literal, Optional

class MetadataFilter(BaseModel):
    """Filter for listing metadata."""
    type: Optional[str] = Field(None, description="Metadata type to filter by")
    namespace_id: Optional[int] = Field(None, description="Filter by namespace")
    name_pattern: Optional[str] = Field(None, description="Pattern to match names")

class MetadataCreateRequest(BaseModel):
    """Request to create new metadata."""
    type: str = Field(..., description="Metadata type")
    data: Dict[str, Any] = Field(..., description="Metadata fields")

class MetadataCreateResponse(BaseModel):
    """Response from metadata creation."""
    type: str = Field(..., description="Metadata type")
    status: Literal["created", "updated", "error"] = Field(..., description="Operation status")
    item: Optional[Dict[str, Any]] = Field(None, description="Created item data")
    error: Optional[str] = Field(None, description="Error message if status is error")

class ListResponse(BaseModel):
    """Response from /list endpoint."""
    items: list = Field(default_factory=list, description="List of metadata items")
    total_count: int = Field(0, description="Total number of items")
    page: Optional[int] = Field(None, description="Current page number")
    page_size: Optional[int] = Field(None, description="Items per page")

class TypesResponse(BaseModel):
    """Response from /types endpoint."""
    types: list[str] = Field(default_factory=list, description="List of supported metadata types")
```

These request and response models define the API contract for the metadata operations, ensuring type safety and automatic documentation.

## Related Documentation

- [v0 Architecture Overview](v0_architecture.md)
- [v0 Requirements - Metrics Service Section](../requirements/v0_requirements/v0_requirements.md#32-metric-service)
