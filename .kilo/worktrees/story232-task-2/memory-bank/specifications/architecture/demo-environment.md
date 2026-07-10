# Demonstration Environment Architecture

This document details the architecture of the ChronoCone demonstration environment, covering the deployment architecture, the Core Server, and the External API Simulator.

## Deployment Architecture

Demonstrations will be made from the containerized development environment, which consists of two docker containers, one for the services and one for the database. The services container will be configured to automatically start the Core server with the `--autostart` flag to launch other necessary services within the environment or coordinate with them.

## Core Server Architecture

The Core Server acts as the central orchestrator and entry point for the demonstration environment. It is implemented as a **Go-native service** utilizing standard library handlers and the Go-Jet persistence layer.

### CLI Interface

The Core Server is invoked directly via the command line to start the system. There are no interactive commands after startup; the service simply runs until terminated.

**Usage:** `./bin/core-service [OPTIONS]`

**Options:**

- `--host`: The host address to bind the server to. Default: `0.0.0.0`.
- `--port`: The port to bind the server to. Default: `8000`.
- `--autostart` / `--no-autostart`: Controls whether dependent services (such as the Metrics Service) are automatically started by the Core Server. Default: `True`.

### API Interface

The Core Server exposes a REST API for service discovery, allowing other components (e.g., MCP servers) to locate running services. This endpoint adheres to the OpenAPI specification.

**Endpoint:** `GET /endpoint`

**Summary:** Retrieve the URL for a specific service type.
**Operation ID:** `getEndpoint`

**Request:**

- **Query Parameters:**
  - `service_type` (string, required): The type of service to lookup (e.g., "metrics").

**Responses:**

- **200 OK (Success):**
  - **Description:** Successfully retrieved the service URL.
  - **Content-Type:** `application/json`
  - **Example Body:**
    ```json
    {
      "url": "http://localhost:8002/metrics"
    }
    ```

- **404 Not Found (Error):**
  - **Description:** The requested service type was not found in the registry.
  - **Content-Type:** `application/json`
  - **Example Body:**
    ```json
    {
      "detail": "Service type 'unknown_service' not found"
    }
    ```

## External API Simulator

_Not Yet Written_
