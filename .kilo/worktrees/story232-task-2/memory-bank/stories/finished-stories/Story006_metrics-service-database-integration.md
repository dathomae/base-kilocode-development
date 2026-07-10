# Story006: Metrics Service Database Integration

## Goal

Modify the Metrics Service `/metadata` and `/list` endpoints to read from and write to the actual PostgreSQL database tables instead of using stub implementations with print statements. The endpoints will use the database schema defined in the schema folder to persist and retrieve metadata definitions.

## References

- [Architecture](../specifications/architecture/)
- [Requirements](../specifications/requirements/)
- [Story004_core-and-metrics-services.md](Story004_core-and-metrics-services.md)
- [Story005_database-initialization-cli.md](Story005_database-initialization-cli.md)
- [Terms](../terms.md)

## Dependencies

- [Story004_core-and-metrics-services.md](Story004_core-and-metrics-services.md) - Not Started (requires basic Metrics Service structure and endpoints)
- [Story005_database-initialization-cli.md](Story005_database-initialization-cli.md) - Not Started (requires database initialization and running PostgreSQL container)

## Dependent Stories

- [Story012_openapi-compliance.md](Story012_openapi-compliance.md) - Not Started

## Tasks

1. Set up database connection module - Completed
   a. Create `src/backend/src/metrics/database.py` for database connection handling - Completed
   b. Implement async connection pool using asyncpg - Completed
   c. Add environment variable configuration for database credentials - Completed
   d. Add connection retry logic with exponential backoff - Completed
   e. Create context manager for database connections - Completed

2. Define database models for metadata - Completed
   a. Create `src/backend/src/metrics/models.py` for database models - Completed
   b. Define Pydantic models for each metadata type to work with asyncpg - Completed
   c. Map models to schema tables: connectors, taxonomy, policy, morphisms - Completed
   d. Define relationships between metadata entities - Completed
   e. Add validation methods for model data - Completed

3. Implement metadata retrieval from database - Completed
   a. Create `src/backend/src/metrics/repository.py` for data access - Completed
   b. Implement `get_metadata_by_type(metadata_type)` function - Completed
   c. Implement `get_metadata_by_id(metadata_type, id)` function - Completed
   d. Implement `list_all_metadata(metadata_type=None)` function - Completed
   e. Add filtering support for metadata queries - Completed
   f. Handle database query errors with appropriate logging - Completed

4. Implement metadata persistence to database - Completed
   a. Implement `create_metadata(metadata_type, data)` function - Completed
   b. Implement `update_metadata(metadata_type, id, data)` function - Completed
   c. Implement `delete_metadata(metadata_type, id)` function - Completed
   d. Add transaction handling for data integrity - Completed
   e. Handle constraint violations with meaningful error messages - Completed

5. Update /list endpoint to query database - Completed
   a. Modify `POST /list` endpoint to call repository layer - Completed
   b. Query database for metadata based on optional `type` filter - Completed
   c. Return actual database records instead of placeholder data - Completed
   d. Handle empty results with appropriate response - Completed
   e. Add pagination support for large result sets - Completed
   f. Test endpoint returns correct data from database - Completed

6. Update /types endpoint to query database - Completed
   a. Modify `POST /types` endpoint to query available metadata types from database - Completed
   b. Query distinct metadata types from the schema - Completed
   c. Return actual types instead of hardcoded list - Completed
   d. Cache type list to reduce database queries - Completed
   e. Test endpoint returns correct types - Completed

7. Update /metadata endpoint to write to database - Completed
   a. Modify `POST /metadata` endpoint to persist data to database - Completed
   b. Validate metadata type against database schema - Completed
   c. Insert new metadata records into appropriate tables - Completed
   d. Return created record ID in response - Completed
   e. Handle duplicate entries with appropriate HTTP status - Completed
   f. Test endpoint creates records in database correctly - Completed

8. Add database health check endpoint - Completed
   a. Implement `GET /health/db` endpoint for database connectivity check - Completed
   b. Return status: connected, error, or degraded - Completed
   c. Include connection pool statistics in response - Completed
   d. Test endpoint reports correct database status - Completed

9. Update tests for database integration - Completed
   a. Modify existing tests to use test database - Completed
   b. Create test fixtures for database setup and teardown - Completed
   c. Add tests for database connection failures - Completed
   d. Add tests for constraint violations - Completed
   e. Add integration tests for full CRUD operations - Completed
   f. Run all tests and verify they pass - Completed

10. Document database integration - Completed
    a. Update metrics service README with database configuration - Completed
    b. Document the database schema and table relationships - Completed
    c. Update API documentation for `/list`, `/types`, and `/metadata` endpoints - Completed
    d. Document environment variables for database connection - Completed
    e. Add troubleshooting guide for common database issues - Completed

## Notes

- The metrics service currently uses stub implementations that print to console
- Database tables are defined in `src/backend/schema/` directory
- Supported metadata types: category, element, namespace, namespace_family, metric, transform, ts_doc_type
- Each metadata type maps to specific tables in the schema
- Connection pooling is essential for handling concurrent requests
- Database credentials should be configurable via environment variables
- Use asyncpg for async database operations with raw SQL queries
- The service should handle database unavailability gracefully
- Test database should be isolated from development/production databases
- Migration strategy should be considered for future schema changes
