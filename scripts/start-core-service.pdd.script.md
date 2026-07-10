# Start Backend Service

This script starts the ChronoCone Go backend service. The service provides the REST API for namespace families, namespaces, categories, elements, and related metadata.

## Parameters

| Name    | Type    | Description                          | Required | Default Value |
| ------- | ------- | ------------------------------------ | -------- | ------------- |
| host    | string  | Host to bind the backend to          | No       | `0.0.0.0`     |
| port    | integer | Port to bind the backend to          | No       | `8080`        |
| verbose | boolean | Display detailed output during start | No       | `true`        |

## Prerequisites

Before running this script, ensure the following prerequisites are met:

1. Go 1.22 or higher is installed. Verify with `go version`.
2. Go module dependencies are downloaded. Run `go mod download` in `src/backend/` if needed.
3. Port 8080 is available (or configure a different port).

## Operations

1. Validate that all required parameters are provided with acceptable values.
2. Display the startup configuration, including host and port.
3. Navigate to the backend directory at `src/backend/`.
4. Execute: `go run main.go`
5. Verify the service is running by making an HTTP request to `http://localhost:8080/families`.
6. Display a summary of the startup results.

## Output

- **Console output**: Real-time progress messages during startup.
- **Startup summary**: Service status and URL.
- **API docs**: Available at `http://localhost:8080/docs` once running.
- **Return code**: 0 if service starts successfully, 1 on failure.

## Error Handling

- **Port already in use**: Detect if port 8080 is in use and display an error.
- **Missing go binary**: Report if Go is not installed.
- **Build errors**: Display compiler output and halt.
