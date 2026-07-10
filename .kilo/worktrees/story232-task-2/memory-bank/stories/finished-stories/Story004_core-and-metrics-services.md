# Story004: Core Service and Basic Metrics Service Implementation

## Goal

Implement the Core Service with an `/endpoint` endpoint for service discovery and create the basic Metrics Service with a `/metadata` endpoint for retrieving and managing metadata. Both services must have independent CLI startup capabilities. The Core Service CLI includes an autostart parameter to optionally launch dependent services automatically.

## References

- [Architecture](../specifications/architecture/)
- [Requirements](../specifications/requirements/)
- [Terms](../terms.md)

## Dependencies

- [Story002_backend-structure.md](Story002_backend-structure.md) - Done (requires backend Python project structure and testing infrastructure)

## Dependent Stories

- [Story012_openapi-compliance.md](Story012_openapi-compliance.md) - Not Started

## Tasks

1. Create core service structure - Completed
   a. Create `src/backend/src/core/` directory for core service code - Completed
   b. Create `src/backend/src/core/__init__.py` - Completed
   c. Create `src/backend/src/core/core_service.py` with FastAPI application - Completed
   d. Define Pydantic models for request/response: `EndpointRequest` (service_type: str) and `EndpointResponse` (url: str) - Completed

2. Implement core service /endpoint endpoint - Completed
   a. Implement POST `/endpoint` endpoint that accepts JSON with `service_type` field - Completed
   b. Create in-memory service registry mapping service types to URLs (metrics, admin, etc.) - Completed
   c. Return JSON response with `url` field for the requested service type - Completed
   d. Handle unknown service types with appropriate HTTP 404 error - Completed
   e. Add basic logging for endpoint requests - Completed

3. Create core service CLI - Completed
   a. Create `src/backend/src/core/core_service.py` with CLI using Click - Completed
   b. Implement `start` command that launches the FastAPI server with uvicorn - Completed
   c. Add configurable host and port options (defaults: 0.0.0.0:8000) - Completed
   d. Add `--autostart` option (default: True) to automatically start dependent services - Completed
   e. When autostart=True, launch Metrics Service on startup - Completed
   f. Test CLI manually: `python -m core.core_service start` and `python -m core.core_service start --autostart=False` - Completed

4. Test core service - Completed
   a. Create `src/backend/tests/core/` directory - Completed
   b. Create `tests/core/test_endpoint.py` with tests for `/endpoint` endpoint - Completed
   c. Test successful service lookup returns correct URL - Completed
   d. Test unknown service type returns 404 - Completed
   e. Test invalid request format returns 422 - Completed
   f. Run tests and verify all pass - Completed

5. Create metrics service structure - Completed
   a. Create `src/backend/src/metrics/` directory for metrics service code - Completed
   b. Create `src/backend/src/metrics/__init__.py` - Completed
   c. Create `src/backend/src/metrics/metrics_service.py` with FastAPI application - Completed
   d. Define Pydantic models: `MetadataDefinition` (type, name, description), and `MetadataResponse` - Completed

6. Implement metrics service /list endpoint (POST) - Completed
   a. Implement POST `/list` endpoint that accepts JSON body with optional `type` field - Completed
   b. Define supported metadata types: category, element, namespace, namespace_family, metric, transform, ts_doc_type - Completed
   c. Parse request JSON to extract optional `type` field - Completed
   d. Print received request to console for debugging (stub implementation) - Completed
   e. Return placeholder JSON response with `items` array containing metadata definitions - Completed

7. Implement metrics service /types endpoint (POST) - Completed
   a. Implement POST `/types` endpoint that accepts any JSON input - Completed
   b. Return JSON response with `types` array containing all supported metadata types - Completed
   c. Add basic logging for requests - Completed

8. Implement metrics service /metadata endpoint (POST) - Completed
   a. Implement POST `/metadata` endpoint that accepts JSON metadata definition - Completed
   b. Accept fields: metadata_type, name, description, and optional attributes - Completed
   c. Print received data to console for debugging (stub implementation) - Completed
   d. Return placeholder JSON response with created/updated metadata ID - Completed
   e. Validate that metadata_type is one of the supported types - Completed

9. Create metrics service CLI - Completed
   a. Create `src/backend/src/metrics/metrics_service.py` with CLI using Click - Completed
   b. Implement `start` command that launches the FastAPI server with uvicorn - Completed
   c. Add configurable host and port options (defaults: 0.0.0.0:8001) - Completed
   d. Test CLI manually: `python -m metrics.metrics_service start` or equivalent - Completed

10. Test metrics service - In Progress
    a. Create `src/backend/tests/metrics/` directory - Completed
    b. Create `tests/metrics/test_list.py` with tests for `/list` endpoint - Completed
    c. Test POST without type returns all metadata - Completed
    d. Test POST with valid type returns filtered metadata - Completed
    e. Test POST with invalid type returns 400 - Completed
    f. Create `tests/metrics/test_types.py` with tests for `/types` endpoint - Completed
    g. Test POST returns list of supported metadata types - Completed
    h. Create `tests/metrics/test_metadata.py` with tests for `/metadata` POST endpoint - Completed
    i. Test POST with valid metadata definition returns success - Completed
    j. Test POST with invalid metadata_type returns 400 - Completed
    k. Run tests and verify all pass - Completed

11. Update CI configuration - Completed
    a. Add CI steps to run core service tests - Completed
    b. Add CI steps to run metrics service tests - Completed
    c. Verify all tests pass in CI pipeline - Completed

12. Document services - Completed
    a. Update backend README with instructions for running core service - Completed
    b. Update backend README with instructions for running metrics service - Completed
    c. Document the `/endpoint` endpoint interface (request/response format) - Completed
    d. Document the `/list` endpoint interface (POST request/response format) - Completed
    e. Document the `/types` endpoint interface (POST request/response format) - Completed
    f. Document the `/metadata` endpoint interface (POST format) - Completed
    g. Document supported metadata types - Completed

## Notes

- Core service runs on port 8000 by default, metrics service on port 8001 to avoid conflicts
- The /admin endpoint for core service is explicitly out of scope for this story
- Both services use FastAPI for consistency
- The metrics service currently uses stub implementations (print statements) for actual database operations
- Metadata types supported: category, element, namespace, namespace_family, metric, transform, ts_doc_type
- Services can be started independently for debugging using `--autostart=False`
- Future stories will implement actual database connectivity for the metrics service
- The core service registry is in-memory only for now; future versions may use configuration files
