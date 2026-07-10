# Story041: Core Service Layered Architecture Compliance

## Goal

Update the Core Service implementation to comply with the updated Layered Architecture and Service Organization patterns, specifically enforcing separation of concerns between model layers and adding comprehensive Pydantic constraints to API inputs.

## Context

The Layered Architecture pattern (`specifications/patterns/layered-architecture.md`) and Service Organization pattern (`specifications/patterns/service-organization.md`) have been updated to require:

1. **Pydantic models ONLY in service files** (`*_service.py`) for API input/output validation
2. **Dataclasses in layer files** (`presentation.py`, `domain.py`, `repository.py`) for inter-layer communication
3. **Comprehensive Field constraints** on all Pydantic models (`min_length`, `max_length`, `pattern`, `ge`, `le`, etc.)
4. **Explicit model conversion** at layer boundaries

The current `core_service` implementation has Pydantic models in `presentation.py` and lacks comprehensive validation constraints.

## References

- [Layered Architecture Pattern](../specifications/patterns/layered-architecture.md)
- [Service Organization Pattern](../specifications/patterns/service-organization.md)
- [OpenAPI Usage Pattern](../specifications/patterns/openapi-usage.md)
- [Story012A: Core Service Patterns Compliance](finished-stories/Story012A_core-service-patterns-compliance.md)

## Dependencies

- [Story012A_core-service-patterns-compliance.md](finished-stories/Story012A_core-service-patterns-compliance.md) - Completed

## Dependent Stories

None

## Tasks

1. **Update `core_service.py` with Pydantic models and constraints** - Completed
   a. Define `EndpointRequest` Pydantic model with Field constraints for `/endpoint` endpoint - Completed
   b. Define `EndpointResponse` Pydantic model with Field constraints for response serialization - Completed
   c. Define `EndpointListResponse` Pydantic model with Field constraints for `/list_endpoints` response - Completed
   d. Add conversion logic to transform Pydantic models to presentation dataclasses - Completed
   e. Update endpoint handlers to use Pydantic models for input/output - Completed

2. **Refactor `presentation.py` to use dataclasses** - Completed
   a. Create `ServiceEndpoint` dataclass for presentation layer input - Completed
   b. Create `EndpointResult` dataclass for presentation layer output - Completed
   c. Create `EndpointListResult` dataclass for list endpoint output - Completed
   d. Update `get_endpoint()` function to accept/return dataclasses - Completed
   e. Update `get_endpoint_list()` function to return dataclass - Completed
   f. Remove all Pydantic imports and model definitions from `presentation.py` - Completed

3. **Verify domain layer compatibility** - Completed
   a. Confirm `domain.py` already uses appropriate dataclass patterns (or simple functions) - Completed
   b. Update any domain layer interfaces if needed to align with presentation dataclasses - Completed

4. **Run tests and verify compliance** - Completed
   a. Run existing tests to ensure API contract remains unchanged - Completed
   b. Verify all tests pass with new implementation - Completed (371 tests pass)
   c. Verify OpenAPI documentation reflects proper constraints - Completed

## Constraints

1. **No Pydantic in Layer Files**: `presentation.py`, `domain.py`, and `repository.py` must use dataclasses only
2. **Pydantic Only in Service File**: All Pydantic models must be defined in `core_service.py`
3. **Field Constraints Required**: All Pydantic string fields must have `min_length`, `max_length`, and `pattern` constraints where applicable
4. **Model Conversion**: Service file must convert Pydantic input to presentation dataclasses before calling presentation layer
5. **Backward Compatibility**: API interface must remain unchanged - same endpoints, same response structure

## Acceptance Criteria

- `core_service.py` contains all Pydantic models with comprehensive Field constraints
- `presentation.py` uses only dataclasses (no Pydantic) for inter-layer communication
- `domain.py` requires no Pydantic dependencies
- All existing tests pass without modification
- OpenAPI documentation shows validation constraints for all input parameters
- Code follows the inter-layer communication flow: Pydantic (service) → dataclass (presentation) → dataclass (domain) → dataclass (presentation) → Pydantic (service)

## Notes

- The service type pattern should be `^[a-z][a-z0-9_]*$` (snake_case starting with lowercase letter)
- URL pattern should validate HTTP/HTTPS URLs: `^https?://.+`
- The `EndpointRequest` model in `presentation.py` is currently unused (endpoint uses query params); this refactor will properly align the models with actual usage
