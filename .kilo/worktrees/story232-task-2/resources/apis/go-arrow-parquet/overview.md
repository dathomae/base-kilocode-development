# Go Arrow Parquet API (pqarrow)

The `pqarrow` package (`github.com/apache/arrow/go/v16/parquet/pqarrow`) provides the implementation for connecting Apache Arrow directly with the Parquet implementation in Go. It allows isolation of all explicitly Arrow-related code to this package, which has the interfaces for reading and writing directly to and from Arrow Arrays, Tables, and Records.

## Key Concepts

- **Arrow Table**: A two-dimensional dataset of a number of columns and rows.
- **Arrow Record**: A batch of rows (Record Batch) that conforms to a specific schema.
- **Parquet File**: A columnar storage file format.
- **Schema Manifest**: A mapping between a Parquet schema and an Arrow schema.

## Core Components

### Reading

- `ReadTable`: A convenience function to quickly and easily read a Parquet file into an Arrow Table.
- `FileReader`: The base object for reading a Parquet file into Arrow object types. It provides utility functions for reading record batches, a table, subsets of columns/rowgroups, etc.
- `ColumnReader`: Used for reading batches of data from a specific column across multiple row groups to return a chunked Arrow array.
- `RecordReader`: A Record Batch Reader that meets the interfaces for both `array.RecordReader` and `arrio.Reader` to allow easy progressive reading of record batches from the Parquet file. Ideal for streaming.

### Writing

- `WriteTable`: A convenience function to create and write a full `array.Table` to a Parquet file.
- `FileWriter`: An object for writing Arrow directly to a Parquet file. It allows progressive building of the file by columns or records without needing a full record or table to write.

### Properties

- `ArrowReadProperties`: Defines how to read a Parquet file into Arrow arrays (e.g., parallel reading, batch size).
- `ArrowWriterProperties`: Determines how to manipulate the Arrow data when writing it to a Parquet file (e.g., coercing timestamps, storing schema).

## Schema Translation

- `FromParquet`: Generates an Arrow Schema from a provided Parquet Schema.
- `ToParquet`: Generates a Parquet Schema from an Arrow Schema using the given properties.
- `SchemaManifest`: Represents a full manifest for mapping a Parquet schema to an Arrow Schema.
