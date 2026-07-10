# ChronoCone Backend

## Quick Usage

### Run init from command line

```bash
python -m src.database.chronocone_db_manager db init
python -m src.database.chronocone_db_manager db init --force
python -m src.database.chronocone_db_manager db init --skip-container
```

### Run status from command line

```bash
python -m src.database.chronocone_db_manager db status
```

### Run using the PDD script

See [`scripts/initialize-database.pdd.script.md`](../../scripts/initialize-database.pdd.script.md) for automated execution with parameter handling and error checking.

---

## Overview

This is the Python FastAPI backend for the ChronoCone application.

## Overview

The backend provides REST API endpoints and WebSocket support for time-series data management, metrics calculation, and policy enforcement. It uses FastAPI as the web framework and SQLAlchemy for database interactions.

## Project Structure

```
backend/
├── src/                    # Application source code
│   ├── __init__.py
│   └── main.py            # FastAPI application entry point
├── tests/                  # Test files
│   ├── __init__.py
│   ├── conftest.py        # Shared pytest fixtures and configuration
│   └── test_*.py          # Test modules
├── serde/rdbms/           # Database schema files
│   └── schema.sql         # Consolidated schema
├── pytest.ini             # Pytest configuration
└── requirements.txt       # Python dependencies
```

## Setup

### Prerequisites

- Python 3.11 or higher
- pip (Python package manager)

### Installation

1. Navigate to the backend directory:

   ```bash
   cd src/backend
   ```

2. Create and activate a virtual environment (recommended):

   ```bash
   python -m venv venv
   source venv/bin/activate  # On Windows: venv\Scripts\activate
   ```

3. Install dependencies:

   ```bash
   pip install -r requirements.txt
   ```

4. Run the application:
   ```bash
   uvicorn src.main:app --reload
   ```

## Testing

### Running Tests

All test commands should be executed from the `src/backend` directory.

#### Run All Tests

```bash
python -m pytest
```

#### Run Tests with Verbose Output

```bash
python -m pytest -v
```

#### Run Tests with Coverage

```bash
python -m pytest --cov=src --cov-report=term-missing
```

#### Run a Specific Test File

```bash
python -m pytest tests/test_sample.py
```

#### Run a Specific Test Function

```bash
python -m pytest tests/test_sample.py::TestSample::test_placeholder
```

#### Run Tests Matching a Pattern

```bash
python -m pytest -k "test_health"
```

### Test Configuration

The `pytest.ini` file contains the test configuration:

- **Test discovery**: Tests are discovered in the `tests/` directory
- **Test files**: Files matching `test_*.py` pattern
- **Test classes**: Classes matching `Test*` pattern
- **Test functions**: Functions matching `test_*` pattern
- **Async support**: pytest-asyncio is configured with `asyncio_mode = auto`

### Available Fixtures

The `conftest.py` file provides shared fixtures for all tests:

#### `client`

A synchronous FastAPI TestClient for making HTTP requests without a live server.

```python
def test_root_endpoint(client):
    response = client.get("/")
    assert response.status_code == 200
```

#### `async_client`

An asynchronous test client for testing async endpoints.

```python
async def test_async_endpoint(async_client):
    response = await async_client.get("/")
    assert response.status_code == 200
```

#### `sample_health_response`

A sample health check response structure for validation.

```python
def test_health_response_structure(client, sample_health_response):
    response = client.get("/health")
    data = response.json()
    assert data.keys() == sample_health_response.keys()
```

#### `test_settings`

Test-specific configuration settings (session-scoped).

```python
def test_with_settings(client, test_settings):
    timeout = test_settings["request_timeout"]
    # Use timeout in test logic
```

## Testing Conventions and Best Practices

### File Naming

- Test files should be named `test_*.py` (e.g., `test_users.py`, `test_metrics.py`)
- This pattern ensures pytest automatically discovers the files

### Test Function Naming

- Test functions should start with `test_` prefix (e.g., `test_create_user`, `test_calculate_metrics`)
- Use descriptive names that explain what is being tested
- Follow the pattern: `test_<action>_<expected_result>` or `test_<scenario>_<expected_behavior>`

### Test Class Naming

- Test classes should start with `Test` prefix (e.g., `TestUserService`, `TestMetricsEndpoints`)
- Use classes to group related tests

### Test Structure

Follow the Arrange-Act-Assert pattern:

```python
def test_example():
    # Arrange: Set up test data and conditions
    input_data = {"name": "test", "value": 42}

    # Act: Execute the code being tested
    result = process_data(input_data)

    # Assert: Verify the expected outcome
    assert result["status"] == "success"
    assert result["value"] == 42
```

### Using Fixtures

- Define shared setup/teardown code in `conftest.py`
- Use fixtures to provide test data, mock objects, or configured clients
- Leverage fixture scopes (`function`, `class`, `module`, `session`) for performance optimization

```python
@pytest.fixture
def sample_user():
    return {"id": 1, "name": "Test User", "email": "test@example.com"}

def test_user_creation(client, sample_user):
    response = client.post("/users", json=sample_user)
    assert response.status_code == 201
```

### Test Markers

Use markers to categorize tests:

- `@pytest.mark.unit` - Unit tests for isolated functionality
- `@pytest.mark.integration` - Integration tests requiring external services
- `@pytest.mark.slow` - Tests that take a long time to run
- `@pytest.mark.api` - Tests for API endpoints

Run tests by marker:

```bash
python -m pytest -m unit
python -m pytest -m "not slow"
```

### Async Testing

For testing async code:

```python
import pytest

@pytest.mark.asyncio
async def test_async_function():
    result = await async_operation()
    assert result is not None
```

Or use the `async_client` fixture for API testing:

```python
async def test_async_endpoint(async_client):
    response = await async_client.get("/async-endpoint")
    assert response.status_code == 200
```

### Best Practices

1. **Keep tests isolated**: Each test should be independent and not rely on the state from other tests
2. **Use fixtures for setup**: Avoid duplicating setup code; use fixtures instead
3. **Test one thing at a time**: Each test should verify a single behavior or concept
4. **Use descriptive names**: Test names should clearly describe what is being tested
5. **Assert precisely**: Make specific assertions rather than broad ones
6. **Clean up resources**: Use fixture cleanup (yield pattern) for resource management
7. **Mock external dependencies**: Use mocks for external services to keep tests fast and isolated

### Example Test Patterns

#### Testing API Endpoints

```python
def test_create_item(client):
    """Test creating a new item via API."""
    response = client.post("/items", json={"name": "Test Item", "value": 100})
    assert response.status_code == 201
    data = response.json()
    assert data["name"] == "Test Item"
    assert data["value"] == 100

def test_get_item_not_found(client):
    """Test retrieving a non-existent item returns 404."""
    response = client.get("/items/99999")
    assert response.status_code == 404
```

#### Testing with Fixtures

```python
@pytest.fixture
def existing_item(client):
    """Create an item and return it for use in tests."""
    response = client.post("/items", json={"name": "Existing", "value": 50})
    return response.json()

def test_update_item(client, existing_item):
    """Test updating an existing item."""
    item_id = existing_item["id"]
    response = client.put(f"/items/{item_id}", json={"value": 75})
    assert response.status_code == 200
    assert response.json()["value"] == 75
```

## Core Service

The Core Service provides service registry and routing for ChronoCone backend services.

### Starting the Core Service

```bash
cd src/backend
python -m src.core.core_service start [OPTIONS]
```

**Options:**

| Option                           | Default   | Description                            |
| -------------------------------- | --------- | -------------------------------------- |
| `--host`                         | `0.0.0.0` | Host to bind the server to             |
| `--port`                         | `8000`    | Port to bind the server to             |
| `--autostart` / `--no-autostart` | `True`    | Automatically start dependent services |

**Examples:**

```bash
# Start with defaults (port 8000, autostart enabled)
python -m src.core.core_service start

# Start on custom port without autostart
python -m src.core.core_service start --port 9000 --no-autostart
```

### Core Service Endpoints

#### POST /endpoint

Lookup endpoint URL for a given service type.

**Request Body:**

```json
{
  "service_type": "metrics"
}
```

**Response (200 OK):**

```json
{
  "url": "http://localhost:8001"
}
```

**Error Response (404 Not Found):**

```json
{
  "detail": "Service type 'unknown' not found"
}
```

**Supported Service Types:**

- `metrics` - Metrics Service (default: http://localhost:8001)
- `admin` - Admin Service (default: http://localhost:8002)

## Metrics Service

The Metrics Service provides metadata management for ChronoCone.

### Starting the Metrics Service

```bash
cd src/backend
python -m src.metrics.metrics_service start [OPTIONS]
```

**Options:**

| Option   | Default   | Description                |
| -------- | --------- | -------------------------- |
| `--host` | `0.0.0.0` | Host to bind the server to |
| `--port` | `8001`    | Port to bind the server to |

**Examples:**

```bash
# Start with defaults (port 8001)
python -m src.metrics.metrics_service start

# Start on custom port
python -m src.metrics.metrics_service start --port 9001
```

### Metrics Service Database Configuration

The Metrics Service connects to PostgreSQL using the following environment variables:

| Variable            | Default Value | Description                                                |
| ------------------- | ------------- | ---------------------------------------------------------- |
| `POSTGRES_HOST`     | `localhost`   | PostgreSQL server hostname                                 |
| `POSTGRES_PORT`     | `5432`        | PostgreSQL server port                                     |
| `POSTGRES_DB`       | `chronocone`  | Database name                                              |
| `POSTGRES_USER`     | `chronocone`  | Database user                                              |
| `POSTGRES_PASSWORD` | `chronocone`  | Database password                                          |
| `DATABASE_URL`      | (optional)    | Full PostgreSQL connection URL (overrides individual vars) |

These variables are shared with the Database CLI. Set them in your environment or a `.env` file.

Example:

```bash
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_DB=chronocone
export POSTGRES_USER=chronocone
export POSTGRES_PASSWORD=chronocone
```

Or using DATABASE_URL:

```bash
export DATABASE_URL="postgresql://chronocone:chronocone@localhost:5432/chronocone"
```

### Supported Metadata Types

The Metrics Service supports the following metadata types:

- `category` - Metric categories
- `element` - Metric elements
- `namespace` - Namespaces
- `namespace_family` - Namespace families
- `metric` - Metric definitions
- `transform` - Data transformation definitions
- `ts_doc_type` - Time-series document types

### Metrics Service Endpoints

#### POST /list

List metadata definitions, optionally filtered by type.

**Request Body:**

```json
{
  "type": "category"
}
```

- `type` (optional): Filter by metadata type. If omitted, returns all types.

**Response (200 OK):**

```json
{
  "items": []
}
```

**Error Response (400 Bad Request):**

```json
{
  "detail": "Invalid metadata type: invalid_type"
}
```

#### POST /types

Get all supported metadata types.

**Request Body:**
Accepts any JSON input (body is ignored).

**Response (200 OK):**

```json
{
  "types": [
    "category",
    "element",
    "namespace",
    "namespace_family",
    "metric",
    "transform",
    "ts_doc_type"
  ]
}
```

#### POST /metadata

Create or update metadata.

**Request Body:**

```json
{
  "metadata_type": "category",
  "name": "Sales",
  "description": "Sales-related metrics",
  "attributes": {}
}
```

- `metadata_type` (required): Must be one of the supported types
- `name` (required): Name of the metadata
- `description` (required): Description of the metadata
- `attributes` (optional): Additional attributes object

**Response (200 OK):**

```json
{
  "id": "1",
  "message": "Metadata 'Sales' of type 'category' created successfully with ID 1"
}
```

**Error Response (400 Bad Request):**

```json
{
  "detail": "Invalid metadata_type: invalid_type. Must be one of: ['category', 'element', 'namespace', 'namespace_family', 'metric', 'transform', 'ts_doc_type']"
}
```

**Note:** The Metrics Service uses PostgreSQL for persistent storage. Ensure the database is initialized before starting the service.

#### GET /health/db

Check database connectivity status.

**Response (200 OK) - Connected:**

```json
{
  "status": "connected",
  "message": "Database connection is healthy"
}
```

**Response (200 OK) - Error:**

```json
{
  "status": "error",
  "message": "Database connection test failed"
}
```

**Usage:**

```bash
# Check database health
curl http://localhost:8001/health/db
```

## Database CLI

The ChronoCone backend includes a Click-based CLI for managing PostgreSQL Docker containers and database schema initialization.

### Prerequisites

Before using the database CLI, ensure you have:

- **Docker** installed and running on your system
  - [Install Docker](https://docs.docker.com/get-docker/) for your platform
  - Verify Docker is running: `docker ps`
- **Python environment** activated with dependencies installed
- **Sufficient disk space** for PostgreSQL data (at least 1GB recommended)

### Environment Variables

The CLI uses the following environment variables for database configuration:

| Variable                  | Default Value         | Description                |
| ------------------------- | --------------------- | -------------------------- |
| `POSTGRES_HOST`           | `localhost`           | PostgreSQL server hostname |
| `POSTGRES_PORT`           | `5432`                | PostgreSQL server port     |
| `POSTGRES_DB`             | `chronocone`          | Database name              |
| `POSTGRES_USER`           | `chronocone`          | Database user              |
| `POSTGRES_PASSWORD`       | `chronocone_secret`   | Database password          |
| `POSTGRES_CONTAINER_NAME` | `chronocone-postgres` | Docker container name      |

Set these variables in your environment or in a `.env` file in the backend directory.

### Using the Database CLI

The CLI is accessed through the `chronocone_db_manager.py` module:

```bash
python -m src.database.chronocone_db_manager [command] [options]
```

Or directly:

```bash
python src/database/chronocone_db_manager.py [command] [options]
```

### Commands

#### `db init` - Initialize the Database

The `init` command sets up the PostgreSQL Docker container and initializes the database schema.

**Usage:**

```bash
python -m src.database.chronocone_db_manager db init [OPTIONS]
```

**Options:**

| Option             | Short | Description                                                            |
| ------------------ | ----- | ---------------------------------------------------------------------- |
| `--force`          | `-f`  | Force recreation of existing tables without prompting                  |
| `--skip-container` |       | Skip Docker container management, assume PostgreSQL is already running |

**Examples:**

```bash
# Initialize database with interactive prompts
python -m src.database.chronocone_db_manager db init

# Force recreate tables without prompting
python -m src.database.chronocone_db_manager db init --force

# Use an existing PostgreSQL instance (skip container management)
python -m src.database.chronocone_db_manager db init --skip-container
```

**What it does:**

1. Checks if Docker daemon is running
2. Starts or creates the `chronocone-postgres` container
3. Waits for PostgreSQL to be ready
4. Discovers SQL schema files from `schema/` directory
5. Checks for existing tables
6. Prompts for confirmation if tables exist (unless `--force` is used)
7. Executes SQL file:
   - schema.sql

#### `db status` - Check Database Status

The `status` command displays the current state of the Docker container and database.

**Usage:**

```bash
python -m src.database.chronocone_db_manager db status
```

**What it displays:**

- Docker daemon status
- Chronocone container status (running, stopped, or not found)
- List of all PostgreSQL containers
- Database connectivity status
- Database version and connection details
- List of existing tables with row counts

**Example output:**

```
============================================================
Chronocone Database Status
============================================================

Docker Status: Running

Chronocone Container: running
  The chronocone-postgres container is running.

All PostgreSQL Containers:
  - chronocone-postgres (abc123def456)
    Status: running
    Image: postgres:15

------------------------------------------------------------
Database Connectivity
------------------------------------------------------------
Database Status: Connected
  Database: chronocone
  User: chronocone
  PostgreSQL Version: PostgreSQL 15.3

------------------------------------------------------------
Tables
------------------------------------------------------------
Table Name                                    Row Count
------------------------------------------------------------
connectors                                       5
policies                                         12
taxonomy_items                                   25
------------------------------------------------------------
Total                                            42

============================================================
Status check complete.
============================================================
```

### Typical Workflows

#### First-Time Setup

```bash
# 1. Navigate to the backend directory
cd src/backend

# 2. Activate your virtual environment
source venv/bin/activate  # On Windows: venv\Scripts\activate

# 3. Install dependencies (if not already done)
pip install -r requirements.txt

# 4. Initialize the database
python -m src.database.chronocone_db_manager db init

# 5. Verify the database is running
python -m src.database.chronocone_db_manager db status
```

#### Reset Database (Development)

```bash
# Force recreate all tables (destroys existing data)
python -m src.database.chronocone_db_manager db init --force
```

#### Using External PostgreSQL

If you have your own PostgreSQL instance:

```bash
# Set environment variables
export POSTGRES_HOST=my-db-server
export POSTGRES_PORT=5432
export POSTGRES_DB=chronocone
export POSTGRES_USER=myuser
export POSTGRES_PASSWORD=mypassword

# Initialize without Docker container management
python -m src.database.chronocone_db_manager db init --skip-container
```

#### Daily Development Workflow

```bash
# Check database status before starting development
python -m src.database.chronocone_db_manager db status

# If container is stopped, start it
python -m src.database.chronocone_db_manager db init

# Run the application
uvicorn src.main:app --reload
```

#### Metrics Service Database Connection Issues

```
Database error: Failed to connect to database after 5 attempts
```

**Solution:**

1. Verify PostgreSQL is running:

   ```bash
   python -m src.database.chronocone_db_manager db status
   ```

2. Check environment variables are set correctly:

   ```bash
   echo $POSTGRES_HOST $POSTGRES_PORT $POSTGRES_DB
   ```

3. Initialize the database if not already done:

   ```bash
   python -m src.database.chronocone_db_manager db init
   ```

4. Check the Metrics Service can connect:
   ```bash
   curl http://localhost:8001/health/db
   ```

#### Invalid Metadata Type Errors

```
Invalid metadata_type: custom_type. Must be one of: ['category', 'element', ...]
```

**Solution:** Use one of the supported metadata types: category, element, namespace, namespace_family, metric, transform, ts_doc_type.

SQL schema files are located in `src/backend/serde/rdbms/`:

```
rdbms/
└── schema.sql         # Consolidated schema
```

Files are executed in dependency order to satisfy foreign key constraints.

## Continuous Integration

The CI pipeline automatically runs tests on every push. The workflow:

1. Sets up Python 3.11
2. Installs backend dependencies from `requirements.txt`
3. Runs pytest with coverage reporting

To check CI status locally after pushing:

```bash
gh run list --branch <branch-name>
```

## Dependencies

Key dependencies:

- **FastAPI** (>=0.104.0) - Web framework
- **Uvicorn** (>=0.24.0) - ASGI server
- **SQLAlchemy** (>=2.0.23) - ORM for database operations
- **psycopg2-binary** (>=2.9.9) - PostgreSQL adapter
- **asyncpg** (>=0.29.0) - Async PostgreSQL driver for database operations
- **pydantic** (>=2.5.0) - Data validation
- **pytest** (>=7.4.3) - Testing framework
- **pytest-asyncio** (>=0.21.1) - Async testing support
- **pytest-cov** (>=4.1.0) - Coverage reporting
- **httpx** (>=0.25.2) - HTTP client for testing

## License

[Add license information here]
