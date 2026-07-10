# Story040: Core Service Architecture Compliance

## Goal

Bring the Core Service implementation into full compliance with the demonstration environment architecture documentation. This involves simplifying the CLI interface to remove the `start` subcommand and removing the `/admin` endpoint from the service registry.

## References

- [Architecture](../specifications/architecture/demo-environment.md)
- [Requirements](../specifications/requirements/demo-environment.md)

## Dependencies

- [Story004_core-and-metrics-services.md](finished-stories/Story004_core-and-metrics-services.md) - Done (requires existing service implementation)

## Dependent Stories

- [Story012A_core-service-patterns-compliance.md](Story012A_core-service-patterns-compliance.md)

## Tasks

1. Create tests for desired CLI behavior - Completed
   a. Create `tests/services/core_service/test_cli_compliance.py` - Completed
   b. Add test case verifying that `python -m src.core.core_service` (without `start`) invokes the server startup - Completed
   c. Add test case verifying that `--host`, `--port`, and `--autostart` arguments are accepted by the top-level command - Completed

2. Refactor Core Service CLI - Completed
   a. Modified `src/backend/src/services/core_service/core_service.py` to remove the `start` subcommand and make the main command directly executable - Completed
   b. Ensured `if __name__ == "__main__":` invokes the main command directly - Completed

3. Update Service Registry - Completed
   a. The `admin` entry does not exist in `SERVICE_REGISTRY` in `src/backend/src/services/core_service/domain.py` - Already Compliant

4. Verify Compliance - Completed
   a. Created `tests/services/core_service/test_cli_compliance.py` with comprehensive CLI tests - Completed
   b. Existing tests remain unaffected by CLI changes - Completed
