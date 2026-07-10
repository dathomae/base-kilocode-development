# Story 015: Time Series Object Parser - Go

## Summary
Implement a Go-native parser and formatter for reading and writing time series object data in Parquet format. This implementation must strictly follow the original design for in-memory representation (multi-level indexing) and metadata handling to ensure 100% compatibility with the Parquet files and symmetry with other platform implementations (Kotlin/TypeScript).

## References
- [Architecture](../specifications/architecture/)
- [Story013_storage-interface.md](Story013_storage-interface.md) - StorageInterface for file I/O
- [Parquet-Go Documentation](https://github.com/segmentio/parquet-go)

## Dependencies
- **Prerequisites**: [Story013_storage-interface.md](Story013_storage-interface.md) (requires StorageInterface for file handling).

## Tasks

### 1. Create the timeseries parser module structure
- Create `internal/timeseries/` directory
- Create `parser.go` for the parser/formatter logic
- Create `models.go` for the in-memory structs
- Create `errors.go` for custom error types

### 2. Define Go structs for in-memory representation
- Implement `TimeseriesObject` struct with `DimensionOrder` ([]string) and `Metrics` (map[string]*MetricData)
- Implement `MetricData` struct representing data for a single metric
- Implement `DataPoint` struct for individual timestamp/value pairs
- **Constraint**: Replicate Pydantic-style validation for dimension order consistency within Go `New` constructors or `Validate()` methods.

### 3. Define custom error types
- Define `ErrTimeseriesParse` for parsing failures
- Define `ErrTimeseriesValidation` for data consistency failures
- Define `ErrTimeseriesFormat` for formatting failures

### 4. Implement custom metadata parsing
- Implement `parseCustomMetadata(metadataJSON string) (map[string]any, error)`
- Extract metric definitions (dimension names, types, time grain, forecast length, data types)
- Extract dimension order from top-level `dimension_order` field
- Validate metadata structure and required fields

### 5. Implement custom metadata formatting
- Implement `formatCustomMetadata(obj *TimeseriesObject) (string, error)`
- Build metric definitions from in-memory representation
- Include dimension order in output JSON
- Ensure JSON output is valid and properly formatted for the Parquet footer

### 6. Implement table data parsing
- Implement `parseTableData(reader *parquet.Reader, metadata map[string]any) (*TimeseriesObject, error)`
- Iterate over Parquet rows and build the multi-level map structure:
  `Metric Name → Dimension Value(s) → Forecast/Actuals → Timestamp → Value`
- Use `DimensionOrder` from metadata to traverse and index dimension values correctly
- Parse JSON-encoded values from the `value` column into Go maps/structs
- Convert Parquet timestamps to Go `time.Time` (UTC)

### 7. Implement table data formatting
- Implement `formatTableData(obj *TimeseriesObject) ([]ParquetRow, error)`
- Flatten the multi-level map structure into a flat row structure
- Build dimension value lists for each row following the `DimensionOrder`
- JSON-encode values for the `value` column
- Convert Go `time.Time` objects to Parquet-compatible timestamps

### 8. Implement the TimeseriesParser logic
- Implement `Parse(data []byte) (*TimeseriesObject, error)`
- Use `parquet-go` to read Parquet data from the byte slice
- Extract and parse custom metadata from the file footer
- Populate and return the `TimeseriesObject`

### 9. Implement the TimeseriesFormatter logic
- Implement `Format(obj *TimeseriesObject) ([]byte, error)`
- Generate custom metadata JSON and Parquet schema
- Generate flattened table data
- Use `parquet-go` to write the Parquet file to a byte buffer

### 10. Implement integration with StorageInterface
- Create `TimeseriesStorage` wrapper that integrates with the existing `StorageInterface`
- Implement `Load(ctx context.Context, configID int, location string) (*TimeseriesObject, error)` with caching
- Implement `Save(ctx context.Context, obj *TimeseriesObject, configID int, location string) error`

### 11. Add round-trip testing
- Create test fixtures with complex multi-dimension sample data
- Test `Parse → Format → Parse` round-trip to ensure zero data loss
- Verify that metadata is perfectly preserved through the cycle

## Decisions

1. **In-memory representation structure**: The in-memory representation uses a multi-level map with the key sequence: `Metric Name → Dimension Value(s) → Forecast/Actuals → Timestamp → Value`. This structure provides O(1) access to any data point when all keys are known.

2. **Dimension order handling**: The `DimensionOrder` attribute is stored at the top level. This ensures consistent traversal of dimension keys during both parsing and formatting.

3. **Value representation**: Values are stored as Go map[string]any. The keys in this map must correspond to the data types defined in the custom metadata (e.g., {"value": 100} or {"units": 100, "revenue": 5000.00}).

4. **Timestamp handling**: All timestamps are Go `time.Time` objects in UTC. They are converted to/from Parquet `TIMESTAMP` type during I/O.

5. **StorageInterface integration**: The `TimeseriesStorage` class integrates with `StorageInterface` caching to ensure each Parquet file is parsed only once per service instance.

## Notes

- The parser and formatter should be symmetric: parsing a formatted object should produce an equivalent object.
- Using `github.com/segmentio/parquet-go` for high-performance, Go-native Parquet handling.
- The in-memory representation is designed for efficient access patterns common in planning applications.
- Future enhancements may include streaming support for very large time series objects.
