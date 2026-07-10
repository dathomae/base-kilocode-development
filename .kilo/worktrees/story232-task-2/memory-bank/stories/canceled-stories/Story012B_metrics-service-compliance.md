# Story012B: Metrics Service Architecture and OpenAPI Compliance

## Goal

Refactor the Metrics Service to strictly adhere to the Layered Architecture and Service Organization patterns, and upgrade the endpoints to be fully compliant with the OpenAPI Specification (OAS) version 3.2.0. This involves separating concerns into distinct presentation, domain, and repository layers, and ensuring comprehensive API documentation.

## References

- [Layered Architecture](../specifications/patterns/layered-architecture.md)
- [Service Organization](../specifications/patterns/service-organization.md)
- [OpenAPI Usage](../specifications/patterns/openapi-usage.md)
- [Architecture](../specifications/architecture/)
- [OpenAPI Overview](../../resources/apis/openapi/overview.md)

## Dependencies

- [Story004_core-and-metrics-services.md](finished-stories/Story004_core-and-metrics-services.md) - Done
- [Story006_metrics-service-database-integration.md](finished-stories/Story006_metrics-service-database-integration.md) - Done
- [Story012A_core-service-patterns-compliance.md](Story012A_core-service-patterns-compliance.md) - Not Started (requires shared error schemas)

## Tasks

1.  **TDD: Prepare Tests** - Not Started
    a. Update existing tests in `tests/metrics/` to import from `src.metrics.presentation` and `src.metrics.domain` instead of `metrics_service` and `services` (anticipating the refactor) - Not Started
    b. Create `tests/metrics/test_openapi.py` for Metrics Service OpenAPI tests - Not Started
    c. Add test verifying `/openapi.json` returns valid JSON with required fields - Not Started
    d. Add test verifying all Metrics Service paths have complete operation documentation - Not Started
    e. Add test verifying all Metrics Service schemas have descriptions and examples - Not Started
    f. Run all tests and verify they fail (expecting `ImportError` due to missing modules and assertion errors due to missing OpenAPI fields) - Not Started

2.  **Refactor to Layered Architecture** - Not Started
    a. Rename `src/backend/src/metrics/services.py` to `src/backend/src/metrics/domain.py` - Not Started
    b. Create `src/backend/src/metrics/presentation.py` - Not Started
    c. Move API route handlers and API-specific Pydantic models (Request/Response bodies) from `metrics_service.py` to `presentation.py` - Not Started
    d. Update `metrics_service.py` to only contain FastAPI app initialization and router inclusion - Not Started
    e. Verify `repository.py` usage aligns with the layered pattern - Not Started
    f. Update all imports in the metrics service to reflect the new file structure - Not Started
    g. Run tests and verify `ImportError`s are resolved (OpenAPI tests should still fail) - Not Started

3.  **Add OpenAPI metadata to Metrics Service FastAPI application** - Not Started
    a. Add `summary`, `description`, `tags`, and `responses` to `/list` endpoint in `presentation.py` - Not Started
    b. Add `summary`, `description`, `tags`, and `responses` to `/list/hierarchy` endpoint in `presentation.py` - Not Started
    c. Add `summary`, `description`, `tags`, and `responses` to `/types` endpoint in `presentation.py` - Not Started
    d. Add `summary`, `description`, `tags`, and `responses` to `/metadata` endpoint in `presentation.py` - Not Started
    e. Add `summary`, `description`, `tags`, and `responses` to `/health/db` endpoint in `presentation.py` - Not Started

4.  **Enhance Metrics Service Pydantic models with OpenAPI annotations** - Not Started
    a. Update `MetadataDefinition` model fields with `Field(description=...)` - Not Started
    b. Add `model_config` with `json_schema_extra` example to `MetadataDefinition` - Not Started
    c. Update `MetadataResponse`, `ListResponse`, `TypesResponse` models with descriptions and examples - Not Started
    d. Update `MetadataCreateRequest` fields with descriptions and examples - Not Started
    e. Update `DatabaseHealthResponse` fields with descriptions - Not Started
    f. Update `MetadataHierarchyItem` and `HierarchyListResponse` fields with descriptions - Not Started

5.  **Use reusable error response schemas** - Not Started
    a. Create `src/backend/src/metrics/schemas.py` for Metrics-specific schemas (if needed) or use shared ones - Not Started
    b. Import and use error response schemas (from Core Service) in Metrics Service endpoint `responses` parameters - Not Started

6.  **Add request body examples to POST endpoints** - Not Started
    a. Add `openapi_extra` to `/metadata` POST endpoint with request body example showing namespace creation - Not Started
    b. Add `openapi_extra` to `/metadata` POST endpoint with request body example showing category creation - Not Started
    c. Add `openapi_extra` to `/metadata` POST endpoint with request body example showing element creation - Not Started

7.  **Add response examples to all endpoints** - Not Started
    a. Add response example to `/list` showing array of metadata items - Not Started
    b. Add response example to `/list/hierarchy` showing item with full hierarchy - Not Started
    c. Add response example to `/types` showing all supported types - Not Started
    d. Add response example to `/metadata` POST showing successful creation - Not Started
    e. Add response example to `/health/db` showing connected status - Not Started

8.  **Add OpenAPI Info and Server objects to FastAPI application** - Not Started
    a. Update Metrics Service `FastAPI()` constructor in `metrics_service.py` with `contact` and `license_info` - Not Started
    b. Update Metrics Service `FastAPI()` constructor in `metrics_service.py` with `servers` for development server on port 8001 - Not Started

9.  **Verify Compliance** - Not Started
    a. Run `tests/metrics/test_openapi.py` and confirm it passes - Not Started

## Notes

- The refactoring to `presentation.py` and `domain.py` is a prerequisite for the OpenAPI work to ensure annotations are placed in the correct files.
- FastAPI automatically generates OpenAPI specifications from the application structure.
- The `Field()` function in Pydantic models supports `description`, `examples`, and constraint parameters that enhance OpenAPI output.
- FastAPI's route decorators support `summary`, `description`, `response_description`, `responses`, `tags`, and `openapi_extra` parameters.
