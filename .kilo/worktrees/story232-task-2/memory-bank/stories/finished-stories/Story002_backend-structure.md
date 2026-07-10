# Story002: Backend Python Project Structure Setup

## Goal

Establish the proper directory structure and testing infrastructure for the backend workspace. This includes creating `src` and `test` directories, configuring pytest, migrating existing code, and verifying that the test framework is working correctly.

## References

- [Architecture](../specifications/architecture/)
- [Brief](../brief.md)

## Dependencies

- [Story001_workspace-setup.md](Story001_workspace-setup.md) - Done (must complete basic environment setup first)

## Dependent Stories

- [Story004_core-and-metrics-services.md](Story004_core-and-metrics-services.md) - Not Started (requires backend project structure and testing infrastructure)

## Tasks

1. Create backend directory structure - Completed
   a. Navigate to `src/backend` directory - Completed
   b. Create `src` subdirectory for application source code - Completed
   c. Create `tests` subdirectory for test files - Completed
   d. Create `__init__.py` files in both `src` and `tests` directories to make them proper Python packages - Completed
   e. Move existing main.py into the `src/` subdirectory - Completed

2. Configure pytest and testing infrastructure - Completed
   a. Update `src/backend/requirements.txt` to include pytest and pytest-asyncio - Completed
   b. Create `pytest.ini` or `pyproject.toml` configuration file in the backend root to configure pytest settings (test discovery paths, etc.) - Completed
   c. Create `tests/conftest.py` for shared pytest fixtures and test configuration - Completed
   d. Add basic test fixtures in conftest.py (e.g., a simple fixture that can be reused across tests) - Completed
   e. Install the testing dependencies in the virtual environment - Completed

3. Create initial test file and verify test execution - Completed
   a. Create `tests/test_sample.py` with a simple dummy test function (e.g., `test_placeholder` that always passes) - Completed
   b. Run pytest from the backend directory to verify test discovery and execution works - Completed
   c. Verify that the test output shows the test was discovered and passed - Completed
   d. Create a second dummy test that intentionally fails to verify pytest properly reports failures - Completed
   e. Fix the failing test to ensure baseline is all tests passing - Completed

4. Set up test coverage reporting (optional but recommended) - Completed
   a. Add pytest-cov to `requirements.txt` - Completed
   b. Configure coverage settings in pytest configuration - Completed
   c. Run tests with coverage to verify it's working - Completed
   d. Review coverage report to ensure it's capturing test execution data - Completed

5. Update top-level CI configuration - Completed
   a. Update `.github/workflows/ci.yml` to run pytest from the correct directory (`src/backend`) - Completed
   b. Configure CI to use the correct Python path for imports - Completed
   c. Push changes to trigger CI workflow - Completed
   d. Verify that CI workflow runs successfully and all tests pass - Completed
   e. If CI fails, diagnose and fix issues, then verify CI passes - Completed

6. Document testing procedures - Completed
   a. Create or update `src/backend/README.md` with testing instructions - Completed
   b. Document how to run tests locally (basic pytest commands) - Completed
   c. Document testing conventions and best practices for the project - Completed

## Notes

- Follow Python testing best practices: test files should be named `test_*.py` or `*_test.py`
- Test functions should be named with `test_` prefix for pytest auto-discovery
- Use pytest fixtures in conftest.py for common test setup/teardown
- The initial test should be simple but demonstrate that the testing infrastructure is properly configured
- conftest.py provides a central location for shared fixtures and test configuration across all test files
