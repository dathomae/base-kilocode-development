# Story 013: Storage Interface Implementation

## Overview

Implement a singleton `StorageInterface` class that provides a unified interface for reading and writing time series data files. The class populates its internal driver registry from the `storage_config` table and dispatches storage requests to the appropriate driver based on the configuration.

For v0, all drivers are known at compile time, and the `storage_config` table serves as a list of available storage configurations. The code may depend on known driver names (e.g., "local" for local filesystem access).

## References

- [Architecture](../specifications/architecture/)
- [Terms - connector](../terms.md)
- [Terms - driver](../terms.md)
- [Schema - storage_config](../../src/backend/serde/rdbms/schema.sql)

## Dependencies

### Prerequisites

- [Story005_database-initialization-cli.md](Story005_database-initialization-cli.md) - Done (database and schema must be available)
- [Story006_metrics-service-database-integration.md](Story006_metrics-service-database-integration.md) - Done (database access patterns established)

### Dependents

- [Story015_timeseries-object-parser.md](Story015_timeseries-object-parser.md) - Not Started (requires StorageInterface for reading/writing Parquet files)

## Tasks

1. Create the storage module structure - Not Started
   a. Create `src/backend/src/storage/` directory
   b. Create `__init__.py` with module exports
   c. Create `storage_interface.py` for the main singleton class
   d. Create `drivers/` subdirectory for driver implementations
   e. Create `exceptions.py` for custom exception classes

2. Define custom exception classes - Not Started
   a. Create `StorageOperationError` as the base exception class
   b. Create `StorageNotFoundError` for missing files/locations
   c. Create `StoragePermissionError` for access denied errors
   d. Create `StorageConnectionError` for backend connection failures
   e. Write unit tests for exception classes

3. Define the base driver interface - Not Started
   a. Create `drivers/base.py` with an abstract `StorageDriver` base class
   b. Define abstract async methods: `read(location: str) -> bytes`, `write(location: str, data: bytes) -> None`, `exists(location: str) -> bool`
   c. Define `delete(location: str) -> None` async method
   d. Include type hints and docstrings for all methods
   e. All methods must be async; no synchronous operations

4. Implement the local filesystem driver - Not Started
   a. Create `drivers/local_driver.py` with `LocalStorageDriver` class
   b. Implement async `read()` using `aiofiles` or `asyncio.to_thread()` with standard file I/O
   c. Implement async `write()` with directory creation if needed
   d. Implement async `exists()` using `asyncio.to_thread()` with `os.path.exists()`
   e. Implement async `delete()` with error handling for missing files
   f. Raise appropriate custom exceptions for error conditions
   g. Write unit tests for `LocalStorageDriver` in `tests/storage/test_local_driver.py`

5. Implement the StorageInterface singleton - Not Started
   a. Create `StorageInterface` class with singleton pattern (using `__new__` or module-level instance)
   b. Implement `_load_drivers()` method that queries `storage_config` table
   c. Create driver registry as a dict mapping `storage_config_id` to driver instances
   d. Implement `get_driver(storage_config_id: int) -> StorageDriver` method
   e. Implement `get_storage_config_by_name(driver_name: str) -> Optional[StorageConfig]` helper
   f. Write unit tests for `StorageInterface` initialization and driver loading

6. Implement storage operations on the interface - Not Started
   a. Implement async `read(storage_config_id: int, location: str) -> bytes` that combines base path from config with relative location
   b. Implement async `write(storage_config_id: int, location: str, data: bytes) -> None`
   c. Implement async `exists(storage_config_id: int, location: str) -> bool`
   d. Implement async `delete(storage_config_id: int, location: str) -> None`
   e. Write integration tests that use a temporary directory as local storage

7. Add database integration for driver loading - Not Started
   a. Create `StorageConfigRepository` class for database access
   b. Implement `get_all_storage_configs()` method
   c. Implement `get_storage_config_by_id(storage_config_id: int)` method
   d. Integrate with existing database connection pool from Metrics Service
   e. Write tests using test database fixtures

8. Create Pydantic models for storage configuration - Not Started
   a. Create `models.py` in storage module
   b. Define `StorageConfig` model matching `storage_config` table schema
   c. Include validation for `direction` field ('I', 'O', 'IO')
   d. Write unit tests for model validation

9. Integration with Metrics Service - Not Started
   a. Initialize `StorageInterface` singleton in Metrics Service startup
   b. Add dependency injection for database connection
   c. Verify storage interface is available when service starts
   d. Write end-to-end test that writes and reads a test file

10. Implement object caching for Parquet files - Not Started
    a. Add internal cache dictionary to `StorageInterface` using storage location as key
    b. Implement `get_cached(location: str) -> Optional[Any]` method to check cache
    c. Implement `cache_object(location: str, obj: Any) -> None` method to store in cache
    d. Implement `get_or_load(storage_config_id: int, location: str, parser: Callable) -> Any` method that checks cache first, loads and caches if not present
    e. Add optional `invalidate(location: str) -> None` method for cache invalidation
    f. Write unit tests for caching behavior including cache hits and misses

## Decisions

1. **Async-only operations**: The `StorageInterface` should not support synchronous operations. All methods must be async to ensure consistency and prepare for cloud storage drivers that require async I/O.

2. **Connection pooling**: For v0, connection pooling is not required with the following condition: no information about which connection(s) are being used may escape the driver. Connection management must be entirely internal to the driver implementation.

3. **Custom exception classes**: Storage operations should use custom exception classes for error handling:
   - `StorageNotFoundError` - When a requested file or storage location does not exist
   - `StoragePermissionError` - When access is denied due to permissions
   - `StorageConnectionError` - When unable to connect to the storage backend
   - `StorageOperationError` - Base exception class for storage-related errors

4. **Object caching for Parquet files**: Multiple documents in the database may reference the same Parquet file in the object store. The `StorageInterface` must cache loaded objects to avoid redundant reads. When a time series object is requested, the interface checks its internal cache first; if not cached, it reads the file, caches the result, and returns it. This ensures each Parquet file is read at most once per service instance.

## Notes

- For v0, only the "local" driver is required. The driver registry should be extensible for future drivers (S3, GCS, etc.).
- The `storage_config.location` field contains the base path, and `ts_doc.location` contains the relative path within that storage.
- The singleton pattern ensures consistent driver instances and efficient resource usage.
- Driver names in the database must match the known driver names in code for v0.
