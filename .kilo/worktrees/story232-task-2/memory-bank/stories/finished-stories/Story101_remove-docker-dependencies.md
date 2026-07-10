# Story 101: Remove Docker Dependencies from Database Code

## Goal

Remove all Docker container management code from the database module and initialization scripts. The environment always uses a PostgreSQL instance provided by the devcontainer service, so explicit Docker container management is no longer needed. The code should assume PostgreSQL is already running at host `db` on port `5432`.

## References

- [Docker Utils](../../src/backend/src/database/docker_utils.py)
- [Database Manager CLI](../../src/backend/src/database/chronocone_db_manager.py)
- [Database Module Init](../../src/backend/src/database/__init__.py)
- [Table Manager](../../src/backend/src/database/table_manager.py)
- [Initialize Database Script](../../scripts/initialize-database.pdd.script.md)

## Dependencies

None - This is a cleanup/refactoring story that removes obsolete functionality.

## Dependent Stories

None

## Tasks

1. Remove Docker utility functions and exports - Completed
   a. Delete `src/backend/src/database/docker_utils.py` file entirely - Completed
   b. Remove all Docker-related imports from `src/backend/src/database/__init__.py` - Completed
   c. Remove all Docker-related exports from `src/backend/src/database/__init__.py` `__all__` list - Completed
   d. Verify no other modules import from `docker_utils` - Completed

2. Refactor database CLI to remove Docker dependency - Completed
   a. Remove Docker-related imports from `src/backend/src/database/chronocone_db_manager.py` - Completed
   b. Remove Docker daemon check at start of `init_db` command - Completed
   c. Remove container status check in `init_db` command - Completed
   d. Remove `wait_for_postgres` call in `init_db` command - Completed
   e. Simplify `init_db` to assume PostgreSQL is already running - Completed
   f. Refactor `db_status` command to remove Docker status checks - Completed
   g. Update `db_status` to show only database connectivity info - Completed

3. Delete Docker-related tests - Completed
   a. Delete `src/backend/tests/database/test_docker_utils.py` file entirely - Completed
   b. Update `src/backend/tests/database/test_cli.py` to remove Docker mocks - Completed
   c. Update test cases in `test_cli.py` to work without Docker checks - Completed
   d. Run tests to verify they pass after changes - Completed

4. Update requirements and configuration - Completed
   a. Check `src/backend/requirements.txt` for Docker SDK dependency - Completed
   b. Remove `docker` package from requirements if present - Completed
   c. Update `table_manager.py` to use `db` as default host instead of `localhost` - Completed

5. Update initialize-database script - Completed
   a. Remove Docker-related prerequisites from `scripts/initialize-database.pdd.script.md` - Completed
   b. Remove Docker daemon check step from Operations section - Completed
   c. Remove container status check step from Operations section - Completed
   d. Update script to assume PostgreSQL is already running - Completed
   e. Update error handling section to remove Docker-specific errors - Completed

## Constraints

- The database connection parameters must work via environment variables (`CHRONOCONE_DB_HOST`, `CHRONOCONE_DB_PORT`, `CHRONOCONE_DB_NAME`, `CHRONOCONE_DB_USER`, `CHRONOCONE_DB_PASSWORD`, `DATABASE_URL`)
- The default host should be `db` (the devcontainer PostgreSQL service name), not `localhost`
- All existing database functionality (schema initialization, table management) must continue to work
- The `skip_container` flag should be removed as it's no longer relevant

## Acceptance Criteria

1. **Docker code removed**: The `docker_utils.py` file is completely deleted and no imports reference it
2. **CLI works without Docker**: The `db init` and `db status` commands work without any Docker checks or dependencies
3. **Tests pass**: All tests pass after removing Docker-related tests and updating remaining tests
4. **Script updated**: The `initialize-database.pdd.script.md` no longer references Docker
5. **Dependencies clean**: No Docker SDK dependency in requirements.txt
6. **Default host updated**: The default database host is `db` instead of `localhost`
7. **Functionality preserved**: Database initialization, schema creation, and status checking work correctly assuming PostgreSQL is running

## Notes

- The `table_manager.py` needs to be updated to use `db` as the default host instead of `localhost`
- The devcontainer environment provides PostgreSQL at host `db` on port `5432`
- The `--skip-container` CLI option should be removed as it's no longer relevant
- A simple connection check at the start of `db init` should verify PostgreSQL connectivity and report a helpful error if not available
