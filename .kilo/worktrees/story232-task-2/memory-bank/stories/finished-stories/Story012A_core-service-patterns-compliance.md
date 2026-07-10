# Story012A: Core Service Patterns Compliance

## Goal

Bring the Core Service implementation into full compliance with all architectural and design patterns defined in `specifications/patterns`, including Layered Architecture, Service Organization, and OpenAPI Usage.

## Context

The Core Service currently resides in `src/backend/src/core` but does not follow the standard directory structure or layered architecture pattern. Additionally, its API documentation needs to be enhanced to meet the OpenAPI Usage pattern.

The Service Organization pattern requires services to be located in `src/services/<service_name>` and shared layers in `src/layers`.

## Prerequisites

- None

## Plan

### 1. Directory Structure Setup

- [x] Create directory `src/backend/src/services`.
- [x] Create directory `src/backend/src/layers`.
- [x] Move `src/backend/src/core` to `src/backend/src/services/core_service`.

### 2. TDD: Update Tests

- [x] Update existing tests in `src/backend/tests/core/` to reflect the new module structure (imports from `services.core_service.presentation` and `services.core_service.domain` instead of `core.core_service` or `core.models`).
- [x] Add tests to verify OpenAPI compliance (e.g., checking `/openapi.json` content for required fields like `summary`, `description`, `examples`).
- [x] Run tests and confirm they fail (due to missing modules and unimplemented OpenAPI features).

### 3. Service Organization and Layered Architecture Refactor

- [x] Create `src/backend/src/services/core_service/presentation.py`.
- [x] Create `src/backend/src/services/core_service/domain.py`.
- [x] Move Pydantic models from `models.py` to `presentation.py` (for DTOs/API models) and `domain.py` (for business logic/entities).
- [x] Implement the Presentation Layer in `presentation.py`:
  - Implement functions to handle data transformation and validation.
  - Call Domain Layer functions.
- [x] Implement the Domain Layer in `domain.py`:
  - Implement business logic.
- [x] Update `src/backend/src/services/core_service/core_service.py` to serve as the main entry point:
  - Define FastAPI app and routes using `@app` or `@router` decorators.
  - Call functions from `presentation.py` within the route handlers.
- [x] Delete `src/backend/src/services/core_service/models.py`.

### 4. Shared Components

- [x] Create `src/backend/src/layers/presentation/schemas.py`.
- [x] Define reusable error response schemas (e.g., `ErrorResponse`, `ValidationErrorResponse`) in `schemas.py` for use by Core and Metrics services.

### 5. OpenAPI Compliance Implementation

- [x] Enhance Pydantic models in `presentation.py` (and `domain.py` if exposed) with `Field()`:
  - Add `description`.
  - Add `examples`.
  - Add constraints (e.g., `min_length`).
- [x] Enhance API routes in `core_service.py` with decorators:
  - Add `summary`.
  - Add `description`.
  - Add `tags`.
  - Add `responses` (including error codes using shared schemas).

### 6. Verification

- [x] Run tests and confirm they pass.
- [x] Manually verify the OpenAPI documentation in the browser (or via curl/test).

## Acceptance Criteria

- Core Service code is moved to `src/backend/src/services/core_service`.
- Core Service code is organized into `core_service.py`, `presentation.py`, and `domain.py`.
- `models.py` is removed.
- Shared components are located in `src/backend/src/layers`.
- API endpoints are fully documented in the generated OpenAPI spec.
- All tests pass.
