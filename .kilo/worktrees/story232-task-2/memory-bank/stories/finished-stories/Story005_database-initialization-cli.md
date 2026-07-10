# Story005: Database Initialization CLI

## Goal

Create a Python CLI tool that manages PostgreSQL Docker container startup and database schema initialization. The tool checks for a running PostgreSQL container, starts one if needed, and executes SQL files from the schema folder to create tables. If tables already exist, the tool prompts the user before dropping and recreating them. Data is never deleted without explicit user consent.

## References

- [Design](../specifications/design/)
- [Requirements](../specifications/requirements/)
- [Terms](../terms.md)

## Dependencies

- [Story002_backend-structure.md](Story002_backend-structure.md) - Done (requires backend Python project structure)

## Dependent Stories

(To be populated as other stories are added)

## Tasks

1. Create CLI module structure - Completed
   a. Create `src/backend/src/database/` directory for database CLI code - Completed
   b. Create `src/backend/src/database/__init__.py` - Completed
   c. Create `src/backend/src/database/cli.py` with CLI using Click - Completed
   d. Define main command group `db` with subcommands `init` and `status` - Completed

2. Implement Docker container detection - Completed
   a. Add function to check if Docker daemon is running - Completed
   b. Add function to list running PostgreSQL containers - Completed
   c. Add function to check for existing chronocone PostgreSQL container by name - Completed
   d. Return container status: running, stopped, or not found - Completed

3. Implement Docker container startup - Completed
   a. Add function to start existing stopped container - Completed
   b. Add function to create and start new PostgreSQL container if none exists - Completed
   c. Use standard PostgreSQL image (postgres:15 or later) - Completed
   d. Configure container with environment variables for database name, user, and password - Completed
   e. Expose standard PostgreSQL port 5432 - Completed
   f. Add volume mount for data persistence at `datastores/postgres_data/` - Completed
   g. Wait for database to be ready before proceeding - Completed

4. Implement SQL file discovery - Completed
   a. Add function to locate the schema folder (`src/backend/schema/`) - Completed
   b. Add function to find all `.sql` files in the schema folder and subdirectories - Completed
   c. Order SQL files by dependencies: connectors, taxonomy, policy, morphisms, timeseries, then indexes - Completed
   d. Log discovered SQL files for user visibility - Completed

5. Implement table existence check - Completed
   a. Add function to connect to PostgreSQL using psycopg2 or asyncpg - Completed
   b. Query information_schema.tables to list existing tables - Completed
   c. Compare existing tables against tables defined in SQL files - Completed
   d. Return list of tables that already exist - Completed

6. Implement user prompt for table replacement - Completed
   a. Display list of existing tables to the user - Completed
   b. Prompt user with: "Tables already exist. Drop and recreate? (y/N):" - Completed
   c. Accept "y" or "Y" to proceed with drop and recreate - Completed
   d. Any other input defaults to "no" and exits without modification - Completed
   e. Require explicit confirmation; do not assume affirmative - Completed

7. Implement schema initialization - Completed
   a. Add function to execute SQL file against PostgreSQL database - Completed
   b. Read SQL file contents and execute using database connection - Completed
   c. Execute SQL files in dependency order - Completed
   d. If user confirmed drop: execute DROP TABLE statements before CREATE TABLE - Completed
   e. If user declined drop: skip existing tables, create only missing tables - Completed
   f. Log each SQL file as it is executed - Completed
   g. Handle SQL execution errors with clear error messages - Completed

8. Implement status command - Completed
   a. Add `db status` subcommand to check database state - Completed
   b. Display Docker container status (running, stopped, not found) - Completed
   c. Display database connectivity status - Completed
   d. Display list of existing tables with row counts - Completed
   e. Display schema version or last initialization timestamp if tracked - Completed

9. Test CLI functionality - Completed
   a. Create `src/backend/tests/database/` directory - Completed (exists)
   b. Create `tests/database/test_cli.py` with unit tests - Completed
   c. Test Docker detection functions with mocked Docker responses - Completed (in test_docker_utils.py)
   d. Test SQL file discovery returns correct file list - Completed (in test_sql_discovery.py)
   e. Test table existence check returns accurate results - Completed (in test_table_manager.py)
   f. Test user prompt logic with simulated input - Completed (in test_table_manager.py)
   g. Test SQL execution with temporary test database - Completed (in test_cli.py)
   h. Run tests and verify all pass - Completed (130 tests passing)

10. Document the CLI - Completed
    a. Update backend README with installation instructions - Completed
    b. Document `db init` command usage and options - Completed
    c. Document `db status` command usage - Completed
    d. Provide examples of typical workflows - Completed
    e. Document environment variables for database configuration - Completed
    f. Document Docker prerequisites - Completed

11. Create PDD script for database initialization - Completed
    a. Create `scripts/initialize-database.pdd.script.md` following the PDD script template - Completed
    b. Define script parameters: environment (dev/test/prod), force_recreate flag - Completed
    c. Document prerequisites: Docker installed, Python environment activated - Completed
    d. Define operations: check Docker status, run CLI `db init` command, verify tables created - Completed
    e. Include error handling for common failure scenarios - Completed
    f. Add verification steps to confirm database is ready for use - Completed

## Notes

- The CLI uses Click for command-line interface consistency
- Docker operations use the Docker SDK for Python (docker package) or subprocess calls to docker CLI
- PostgreSQL connection uses standard environment variables: POSTGRES_DB, POSTGRES_USER, POSTGRES_PASSWORD
- Default container name is `chronocone-postgres` to avoid conflicts with other PostgreSQL containers
- Data persistence is achieved through Docker volume mount to `datastores/postgres_data/`
- The schema folder is located at `src/backend/schema/`
- SQL files are executed in dependency order to satisfy foreign key constraints
- User confirmation is required before any destructive operations; silent execution is not allowed
- The CLI provides clear feedback at each step: container status, SQL file execution, and completion
- Future stories may add schema migration capabilities beyond initial table creation
