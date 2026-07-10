Directives for agentic or AI writers and editors:

1. You MUST leave this list of directives in place as the first item in the document and unaltered as changes are made to this document.
2. When editing this document you MUST use either the Technical Writer or Technical Editor roles. If you are not in one of these modes you MUST NOT edit this document and inform the user of the requirement.
3. This document contains annotations that are labelled as "Notes:" or "Questions:" and you MUST NOT edit these annotations. You should move them to keep them with the content they relate to.

--- Begin Document Content ---

# Core Service

## Purpose

The Core Service serves as the central coordinator for the ChronoCone system. It handles system startup and provides a mechanism for discovering endpoints for services, queues, or other functionality. This centralized coordination role is essential for maintaining loose coupling between components: rather than hard-coding connection information throughout the system, components ask the Core Service where to find the resources they need.

The Core Service addresses a fundamental challenge in distributed systems: how do components find each other without creating tight coupling? In a traditional monolithic application, components can call each other directly because they all exist in the same process. In a distributed system like ChronoCone, components run as separate services that need to communicate over the network. The Core Service provides a service registry that allows components to discover each other dynamically, enabling the system to be deployed in different configurations without code changes.

## Responsibilities

The Core Service has three primary responsibilities that support the overall system architecture.

**System startup and initialization** involves orchestrating the startup sequence for the entire system. When the Core Service starts, it can optionally start dependent services, ensuring that the system comes up in a consistent state. This orchestration simplifies deployment and reduces the chance of startup order problems.

**Service discovery for other system components** provides a registry where services can register themselves and where clients can look up service endpoints. This registry is the foundation for loose coupling between components, allowing services to be moved, scaled, or replaced without requiring changes to the clients that use them.

**Optional automatic startup of dependent services** allows the Core Service to function as a simple process manager in development and demonstration environments. This capability is particularly valuable for the v0 release, where the entire system runs on a single device.

## Interfaces

The Core Service provides two major top-level interfaces that serve different audiences and purposes.

### /endpoint Interface

The `/endpoint` interface is the primary service discovery mechanism. When supplied with an endpoint type, the Core Service replies with a datatype that is appropriate for the endpoint type requested. This simple request-response pattern allows any component to find any other component by type, without needing to know the specific host or port where that component is running.

**Request:**

```json
{
  "service_type": "metrics"
}
```

**Response:**

```json
{
  "url": "http://localhost:8001"
}
```

The request specifies the type of service needed (in this example, "metrics"), and the response provides the URL where that service can be reached. This abstraction allows the system to be deployed in different configurations—single host, multiple hosts, containerized, or cloud-based—without changing the code that uses the service discovery mechanism.

### /admin Interface

The `/admin` interface provides functionality primarily intended for system administrators. This interface supports setup, configuration, and other administrative tasks that are distinct from the normal operational use of the system. The admin interface is designed to be used by operators and deployment tools rather than by end users or business applications.

The admin interface provides two sub-endpoints: `/admin/health` for monitoring system status and `/admin/config` for managing configuration parameters.

#### /admin/health Endpoint

The `/admin/health` endpoint provides a way to monitor the operational status of the ChronoCone system and its dependent services. This endpoint is essential for operations teams who need to verify that the system is running correctly and for automated monitoring systems that track system health.

**GET /admin/health**

Returns the current health status of the Core Service and all registered services.

**Response:**

```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "services": [
    {
      "name": "core",
      "status": "healthy",
      "url": "http://localhost:8000"
    },
    {
      "name": "metrics",
      "status": "healthy",
      "url": "http://localhost:8001"
    },
    {
      "name": "database",
      "status": "healthy",
      "details": "PostgreSQL connection pool: 5/10 connections in use"
    },
    {
      "name": "object_store",
      "status": "healthy",
      "details": "Local filesystem at /var/chronocone/data"
    }
  ]
}
```

The response includes an overall `status` field that indicates whether the entire system is operational. The `timestamp` field records when the health check was performed. The `services` array contains individual health status entries for each component.

Each service entry includes a `name` identifying the service, a `status` field indicating its health, and optional additional details. The `status` field can have one of three values:

- `healthy`: The service is operating normally.
- `degraded`: The service is running but with reduced functionality or performance.
- `unhealthy`: The service is not responding or has failed.

The Core Service determines the health of dependent services by making HTTP requests to their health endpoints or by checking their registration status in the service registry. For the database and object store, the Core Service performs connectivity checks to verify that these resources are accessible.

**Error Response:**

If the Core Service itself is experiencing problems but can still respond, it returns a response with `status: "unhealthy"` and includes details about which services are failing:

```json
{
  "status": "unhealthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "services": [
    {
      "name": "core",
      "status": "healthy",
      "url": "http://localhost:8000"
    },
    {
      "name": "metrics",
      "status": "unhealthy",
      "url": "http://localhost:8001",
      "error": "Connection refused"
    },
    {
      "name": "database",
      "status": "healthy",
      "details": "PostgreSQL connection pool: 3/10 connections in use"
    },
    {
      "name": "object_store",
      "status": "healthy",
      "details": "Local filesystem at /var/chronocone/data"
    }
  ]
}
```

#### /admin/config Endpoint

The `/admin/config` endpoint allows administrators to view and update system configuration parameters. This endpoint supports both GET and POST methods, providing read and write access to configuration values.

**GET /admin/config**

Returns the current values of all configuration parameters.

**Response:**

```json
{
  "config": {
    "database": {
      "host": "localhost",
      "port": 5432
    },
    "object_store": {
      "path": "/var/chronocone/data"
    }
  }
}
```

The response contains a `config` object with nested objects for each configuration category. The initial configuration categories are:

- `database`: Connection parameters for the PostgreSQL database
  - `host`: The hostname or IP address of the database server
  - `port`: The port number for database connections

- `object_store`: Configuration for the time series data storage
  - `path`: The local filesystem path where time series data files are stored

**POST /admin/config**

Updates one or more configuration parameters. The request body contains only the parameters to be updated; parameters not included in the request retain their current values.

**Request:**

```json
{
  "database": {
    "host": "db.example.com",
    "port": 5433
  },
  "object_store": {
    "path": "/data/chronocone/metrics"
  }
}
```

**Response:**

```json
{
  "status": "updated",
  "config": {
    "database": {
      "host": "db.example.com",
      "port": 5433
    },
    "object_store": {
      "path": "/data/chronocone/metrics"
    }
  }
}
```

The response confirms the update with a `status` field and returns the complete updated configuration in the `config` field.

**Configuration Update Behavior:**

Configuration updates take effect immediately for new connections and requests. Existing connections to the database continue to use the previous configuration until they are closed and reestablished. This approach ensures that configuration changes do not disrupt ongoing operations.

For the object store path, the Core Service validates that the new path exists and is writable before applying the change. If validation fails, the configuration is not updated and an error response is returned:

```json
{
  "status": "error",
  "error": "Path validation failed",
  "details": "Directory /data/chronocone/metrics does not exist or is not writable"
}
```

**Partial Updates:**

The POST endpoint supports partial updates, allowing administrators to change individual parameters without specifying the entire configuration. For example, to update only the database port:

```json
{
  "database": {
    "port": 5433
  }
}
```

This request updates only the `database.port` parameter while preserving `database.host` and all `object_store` parameters.

### /queue Interface

The `/queue` interface provides a message queuing mechanism for asynchronous communication between services and between services and clients. This interface addresses a fundamental challenge in distributed systems: some operations, such as metric updates, may take longer than can be reasonably accommodated in a synchronous request-response pattern. The queuing mechanism allows these longer-running operations to proceed without blocking the caller.

#### Motivation

The queuing system serves two primary purposes in the ChronoCone architecture. First, it enables asynchronous processing of operations that exceed reasonable synchronous request timeouts. When a metric update requires extensive calculation or data retrieval, the operation can be initiated through the queue and the results delivered when complete. Second, it provides a publish-subscribe pattern where multiple recipients can subscribe to the same queue and receive copies of all messages sent to that queue.

#### Queue Model

The queuing system follows a named-queue model. Recipients subscribe to a queue by name, and senders send messages to that same queue name. When a message is sent to a queue, the Core Service delivers the message to all registered subscribers.

In v0, the queue implementation uses a simple in-memory mechanism. This approach is appropriate for the single-device deployment model of v0, where all services run on the same host. In post-v0 releases, the implementation will be replaced with production-grade message brokers such as AWS SNS/SQS for cloud deployments or RabbitMQ for local development.

#### Queue Operations

The `/queue` interface supports three primary operations: sending messages, registering subscribers, and unregistering subscribers.

**Send Message to Queue**

To send a message to a queue, a client posts to the `/queue` endpoint with the queue name and message payload.

```json
{
  "queue_name": "metric-updates",
  "message": {
    "metric_id": "sales_revenue",
    "namespace": "retail",
    "status": "updated",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

The `queue_name` field identifies the target queue. The `message` field contains an arbitrary JSON object that will be delivered to all subscribers of that queue.

**Register Subscriber**

To receive messages from a queue, a service registers itself as a subscriber by posting to the `/queue/subscribe` endpoint.

```json
{
  "queue_name": "metric-updates",
  "subscriber_url": "http://localhost:8001/queue/receive"
}
```

The `queue_name` field specifies which queue to subscribe to. The `subscriber_url` field provides the URL where the subscriber will receive POST requests when messages arrive.

**Unregister Subscriber**

To stop receiving messages, a service unregisters by posting to the `/queue/unsubscribe` endpoint.

```json
{
  "queue_name": "metric-updates",
  "subscriber_url": "http://localhost:8001/queue/receive"
}
```

#### Message Delivery

When a message is sent to a queue, the Core Service delivers the message to all registered subscribers as a synchronous operation. The Core Service posts the message to each subscriber URL in sequence, waiting for each delivery to complete before proceeding to the next subscriber.

The message is delivered as a POST request to the subscriber's registered URL with the following body:

```json
{
  "queue_name": "metric-updates",
  "message": {
    "metric_id": "sales_revenue",
    "namespace": "retail",
    "status": "updated",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

If a subscriber fails to acknowledge receipt (returns a non-2xx status code or times out), the Core Service logs the failure but continues delivering to remaining subscribers. In v0, failed deliveries are not retried. Post-v0 implementations using production message brokers will provide configurable retry policies and dead-letter queues.

#### Queue Utility Object

Both senders and receivers interact with the queuing system through a utility `Queue` object that abstracts the HTTP communication details. This object provides a programmatic interface that is more convenient than making raw HTTP requests.

**Sender Usage**

Senders use the `Queue` object to send messages without managing HTTP connections:

```python
from core.queue import Queue

queue = Queue(core_service_url="http://localhost:8000")
queue.send("metric-updates", {
    "metric_id": "sales_revenue",
    "namespace": "retail",
    "status": "updated"
})
```

**Receiver Usage**

Receivers use the `Queue` object to manage subscriptions and handle incoming messages. The receiver registers a callback function that will be invoked when messages arrive:

```python
from core.queue import Queue

def handle_metric_update(message):
    """Process incoming metric update message."""
    metric_id = message["metric_id"]
    namespace = message["namespace"]
    # Process the update...

queue = Queue(core_service_url="http://localhost:8000")
queue.subscribe(
    queue_name="metric-updates",
    subscriber_url="http://localhost:8001/queue/receive",
    callback=handle_metric_update
)
```

The `Queue` object handles the internal mechanics of receiving messages. When the Core Service POSTs a message to the subscriber URL, the `Queue` object receives the HTTP request and queues the message internally. A background task periodically dispatches queued messages to the registered callback functions. This design decouples the HTTP receive operation from the message processing, allowing the subscriber to process messages at its own pace without blocking the HTTP response.

**Callback Dispatch Model**

The `Queue` object maintains an internal queue of received messages. When a message arrives via HTTP POST, it is added to this internal queue. A background task runs periodically to check for pending messages and dispatch them to the appropriate callback functions. This model provides several benefits:

1. **Non-blocking HTTP responses**: The HTTP endpoint can return immediately after receiving the message, without waiting for the callback to complete.

2. **Error isolation**: If a callback raises an exception, it does not affect the HTTP response or the delivery of other messages.

3. **Rate control**: The subscriber can control how frequently it processes messages by adjusting the dispatch interval.

4. **Message ordering**: Messages are processed in the order they were received, within each queue.

## Pydantic Models

The Core Service uses Pydantic models for request/response validation. These models ensure that requests are well-formed before processing and that responses conform to the expected structure. Pydantic provides automatic validation, serialization, and documentation generation, which reduces boilerplate code and helps catch errors early.

```python
from typing import Literal
from pydantic import BaseModel, Field

class EndpointRequest(BaseModel):
    """Request model for the /endpoint endpoint."""
    service_type: str = Field(
        ...,
        description="Type of service to discover (e.g., 'metrics', 'admin', 'ui')"
    )

class EndpointResponse(BaseModel):
    """Response model for the /endpoint endpoint."""
    url: str = Field(
        ...,
        description="URL of the requested service endpoint"
    )

class ServiceConfig(BaseModel):
    """Configuration for a discoverable service."""
    service_type: str = Field(..., description="Service type identifier")
    url: str = Field(..., description="Service URL")
    host: str = Field(default="localhost", description="Service host")
    port: int = Field(..., description="Service port")
    description: str = Field(default="", description="Service description")

# Queue-related models

class QueueSendMessage(BaseModel):
    """Request model for sending a message to a queue."""
    queue_name: str = Field(
        ...,
        description="Name of the queue to send the message to"
    )
    message: dict = Field(
        ...,
        description="Message payload as an arbitrary JSON object"
    )

class QueueSubscribe(BaseModel):
    """Request model for subscribing to a queue."""
    queue_name: str = Field(
        ...,
        description="Name of the queue to subscribe to"
    )
    subscriber_url: str = Field(
        ...,
        description="URL where the subscriber will receive POST requests for messages"
    )

class QueueUnsubscribe(BaseModel):
    """Request model for unsubscribing from a queue."""
    queue_name: str = Field(
        ...,
        description="Name of the queue to unsubscribe from"
    )
    subscriber_url: str = Field(
        ...,
        description="URL that was previously registered for this subscription"
    )

class QueueMessageDelivery(BaseModel):
    """Model for messages delivered to subscribers."""
    queue_name: str = Field(
        ...,
        description="Name of the queue the message was sent to"
    )
    message: dict = Field(
        ...,
        description="Message payload"
    )

# Admin-related models

class ServiceHealth(BaseModel):
    """Health status of an individual service."""
    name: str = Field(
        ...,
        description="Name of the service"
    )
    status: Literal["healthy", "degraded", "unhealthy"] = Field(
        ...,
        description="Current health status of the service"
    )
    url: str | None = Field(
        default=None,
        description="URL of the service, if applicable"
    )
    details: str | None = Field(
        default=None,
        description="Additional details about the service status"
    )
    error: str | None = Field(
        default=None,
        description="Error message if the service is unhealthy"
    )

class HealthResponse(BaseModel):
    """Response model for the /admin/health endpoint."""
    status: Literal["healthy", "degraded", "unhealthy"] = Field(
        ...,
        description="Overall health status of the system"
    )
    timestamp: str = Field(
        ...,
        description="ISO 8601 timestamp of when the health check was performed"
    )
    services: list[ServiceHealth] = Field(
        ...,
        description="Health status of individual services"
    )

class DatabaseConfig(BaseModel):
    """Database connection configuration."""
    host: str = Field(
        ...,
        description="Hostname or IP address of the database server"
    )
    port: int = Field(
        ...,
        description="Port number for database connections"
    )

class ObjectStoreConfig(BaseModel):
    """Object store configuration."""
    path: str = Field(
        ...,
        description="Local filesystem path for time series data storage"
    )

class SystemConfig(BaseModel):
    """Complete system configuration."""
    database: DatabaseConfig = Field(
        ...,
        description="Database connection parameters"
    )
    object_store: ObjectStoreConfig = Field(
        ...,
        description="Object store parameters"
    )

class ConfigUpdateRequest(BaseModel):
    """Request model for updating configuration parameters."""
    database: DatabaseConfig | None = Field(
        default=None,
        description="Updated database configuration (optional)"
    )
    object_store: ObjectStoreConfig | None = Field(
        default=None,
        description="Updated object store configuration (optional)"
    )

class ConfigResponse(BaseModel):
    """Response model for the /admin/config endpoint."""
    config: SystemConfig = Field(
        ...,
        description="Current system configuration"
    )

class ConfigUpdateResponse(BaseModel):
    """Response model for successful configuration updates."""
    status: Literal["updated"] = Field(
        ...,
        description="Status indicating the update was successful"
    )
    config: SystemConfig = Field(
        ...,
        description="The updated configuration"
    )

class ConfigErrorResponse(BaseModel):
    """Response model for configuration update failures."""
    status: Literal["error"] = Field(
        ...,
        description="Status indicating the update failed"
    )
    error: str = Field(
        ...,
        description="Brief description of the error"
    )
    details: str = Field(
        ...,
        description="Detailed error message"
    )
```

The `EndpointRequest` model validates that incoming requests contain a `service_type` field. The `EndpointResponse` model ensures that responses include the required `url` field. The `ServiceConfig` model captures the full configuration for a service, including its type, URL, host, port, and description.

The queue-related models support the `/queue` interface. `QueueSendMessage` validates messages being sent to a queue, ensuring both the queue name and message payload are present. `QueueSubscribe` and `QueueUnsubscribe` handle subscriber registration with the required queue name and subscriber URL fields. `QueueMessageDelivery` defines the structure of messages delivered to subscribers, including the queue name for routing and the message payload.

The admin-related models support the `/admin/health` and `/admin/config` endpoints. `ServiceHealth` represents the health status of an individual service, including its name, status, and optional URL, details, or error information. `HealthResponse` structures the complete health check response with an overall status, timestamp, and list of service health entries.

The configuration models provide a hierarchical structure for system configuration. `DatabaseConfig` captures database connection parameters (host and port). `ObjectStoreConfig` specifies the filesystem path for time series data storage. `SystemConfig` combines these into a complete configuration object. `ConfigUpdateRequest` supports partial updates by making both database and object_store fields optional. `ConfigResponse` wraps the system configuration for GET responses. `ConfigUpdateResponse` confirms successful updates with the updated configuration. `ConfigErrorResponse` provides detailed error information when configuration updates fail.

## Implementation

The Core Service is implemented using the FastAPI web framework. FastAPI was chosen for its performance, automatic OpenAPI documentation generation, and strong support for type hints through Pydantic integration. These characteristics make it well-suited for building APIs that need to be both performant and easy to understand.

### File Structure

The implementation is organized into two main files that separate concerns appropriately.

`src/backend/src/core/core_service.py` contains the main FastAPI application with `/endpoint` and `/admin` endpoints. This file defines the API routes, request handlers, and service registry logic.

`src/backend/src/core/cli.py` provides the CLI for starting the Core Service. This file handles command-line argument parsing and service initialization.

### CLI Usage

The Core Service CLI is invoked via `core_service.py`:

```bash
python -m core.cli start [options]
```

#### CLI Parameters

The CLI accepts several parameters that control how the service starts.

`--host` specifies the host to bind the server, with a default of 0.0.0.0. This default allows the service to accept connections from any network interface, which is appropriate for containerized deployments.

`--port` specifies the port to bind the server, with a default of 8000. This port was chosen as the default for the Core Service to leave port 8001 available for the Metrics Service.

`--autostart` controls whether to automatically start dependent services, with a default of True. This parameter is key to the development workflow.

#### Autostart Behavior

The `--autostart` parameter controls whether the Core Service automatically starts dependent services. This capability is particularly important for the v0 architecture, where the entire system is designed to run on a single device for development and demonstration purposes.

When `autostart=True` (the default), the Core Service automatically starts the Metrics Service and other dependent services on startup. This behavior provides a simple "one command" startup experience that is ideal for demonstrations and quick development iterations.

When `autostart=False`, the Core Service starts only itself. Developers must manually start dependent services using their respective CLIs. This mode facilitates debugging by allowing services to be started independently, making it easier to isolate problems to specific services.

Example for debugging:

```bash
# Terminal 1: Start Core Service without autostart
python -m core.cli start --autostart=False

# Terminal 2: Start Metrics Service manually
python -m metrics.cli start
```

This debugging workflow allows developers to restart individual services without affecting others, attach debuggers to specific services, and observe service interactions more clearly.

## Service Registry

The Core Service maintains an in-memory registry of discoverable services. This registry maps service types to their URLs, providing the information needed to respond to `/endpoint` requests.

| Service Type | Default URL           | Description                           |
| ------------ | --------------------- | ------------------------------------- |
| `metrics`    | http://localhost:8001 | Metrics Service for metadata and data |
| `admin`      | http://localhost:8002 | Administrative interface (future)     |
| `ui`         | http://localhost:3000 | Web UI frontend (future)              |

The registry is currently static, with service URLs determined at deployment time. In future versions, services may register themselves dynamically, allowing for more flexible deployment topologies and automatic failover.

## Future Migration Path

The Core Service is designed with a clear migration path toward cloud-native architectures. In a production deployment, the centralized service discovery function would likely be replaced by configuration or an API Gateway instance. API Gateway services like AWS API Gateway provide similar service discovery and routing capabilities with additional features like authentication, rate limiting, and request transformation.

The current design ensures that this migration can happen without requiring changes to the components that use the Core Service. Components will continue to make service discovery requests; only the implementation of the discovery mechanism will change. This forward-compatible design is a key benefit of the abstraction that the Core Service provides.

## Related Documentation

- [v0 Architecture Overview](v0_architecture.md)
- [v0 Requirements - Core Service Section](../requirements/v0_requirements/v0_requirements.md#31-core-server)
