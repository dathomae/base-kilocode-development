# Story 021: Metrics Service Refactor
**Status**: Done
**Epic**: API Layer Foundation
**Owner**: @horatiar

## Overview
This story covers the architectural refactoring of the FastAPI application (`src/metrics/metrics_service.py`). The existing controller routes directly imported repository functions and duplicated business logic (e.g., ID resolution) that violates strict layering. We are migrating `metrics_service.py` to exclusively depend on the newly established **Ontology Service** layer (`services.py`).

## References
* `Story020_ontology-services.md`
* `src/metrics/services.py`
* `src/metrics/metrics_service.py`

## Dependencies
* `src/metrics/services.py` (Ontology Service)
* `src/metrics/models.py` (Pydantic schemas and DTOs)

## Tasks
- [x] Remove `src.metrics.repository` imports from `metrics_service.py`.
- [x] Delete `_resolve_foreign_keys` from `metrics_service.py` and rely on `services.py` natural key resolution.
- [x] Route `/metadata` POST requests to the corresponding `upsert_*` function in `services.py` based on `metadata_type`.
- [x] Route `/list` and `/list/hierarchy` GET requests through `services.py` wrappers to completely break the `repository.py` dependency.
- [x] Ensure all existing tests pass without modification.
