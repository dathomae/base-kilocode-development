# Initialize Database

This script initializes the ChronoCone PostgreSQL database by executing the schema SQL files. It assumes a PostgreSQL database is already running and accessible.

## Parameters

| Name           | Type    | Description                                           | Required | Default Value |
| -------------- | ------- | ----------------------------------------------------- | -------- | ------------- |
| environment    | string  | Target environment: `dev`, `test`, or `prod`          | No       | `dev`         |
| verbose        | boolean | Display detailed output during initialization         | No       | `true`        |

## Prerequisites

Before running this script, ensure the following prerequisites are met:

1. A PostgreSQL database is running and accessible.
2. The `CHRONOCONE_DB_HOST`, `CHRONOCONE_DB_PORT`, `CHRONOCONE_DB_USER`, `CHRONOCONE_DB_PASSWORD`, and `CHRONOCONE_DB_NAME` environment variables are set.
3. Port 5432 is available or the relevant environment variables point to the correct port.

## Operations

1. Validate that all required parameters are provided with acceptable values.
2. Display the initialization configuration, including environment and flags.
3. Locate SQL schema files in `src/backend/schema/`.
4. Execute each SQL schema file against the database in dependency order:
   - connectors.sql
   - taxonomy.sql
   - policy.sql
   - morphisms.sql
   - timeseries.sql
5. Verify database connectivity by querying table names after initialization.
6. Display a summary of the initialization results.

## Output

- **Console output**: Real-time progress messages during initialization.
- **Initialization summary**: Tables created and database connectivity status.
- **Return code**: 0 on success, 1 on failure.

## Error Handling

- **Database connection refused**: Indicates PostgreSQL is not accessible.
- **SQL execution errors**: Script halts without partial changes.
- **Missing schema files**: Reports condition and exits if no SQL files found in `src/backend/schema/`.
