# Story 019: Time Series Object Parser - TypeScript (VSCode Plugin)

## Overview

Implement a TypeScript parser and formatter for reading and writing time series object data in Parquet format, for use in the VSCode plugin. The parser uses a Node.js Parquet library to handle Parquet file I/O and produces the same in-memory representation as the Go implementation (multi-level object structure).

## References

- [Architecture](../specifications/architecture/)
- [Story015_timeseries-parser-go.md](Story015_timeseries-parser-go.md) - Go implementation for reference
- [Story017_timeseries-parser-kotlin.md](Story017_timeseries-parser-kotlin.md) - Kotlin implementation for reference
- [Story018_vscode-workspace-setup.md](Story018_vscode-workspace-setup.md) - VSCode workspace setup

## Dependencies

### Prerequisites

- [Story018_vscode-workspace-setup.md](Story018_vscode-workspace-setup.md) - Not Started (requires VSCode extension project structure)
- Parquet library dependency configured in the VSCode extension project

### Dependents

None yet. Future stories for VSCode plugin time series operations will depend on this story.

## Tasks

1. Add Parquet library dependency to VSCode extension project - Not Started
   a. Evaluate available Node.js Parquet libraries (parquetjs, apache-arrow, etc.)
   b. Add chosen library to package.json dependencies
   c. Install type definitions if available (or create custom declarations)
   d. Verify dependency resolution
   e. Document required library version in extension documentation

2. Create the timeseries parser module structure - Not Started
   a. Create `src/timeseries/` directory in extension source
   b. Create `parser.ts` for the parser class
   c. Create `formatter.ts` for the formatter class
   d. Create `models/` subdirectory for TypeScript interfaces and types
   e. Create `exceptions.ts` for custom error classes
   f. Create `index.ts` for module exports

3. Define TypeScript interfaces for in-memory representation - Not Started
   a. Create `TimeseriesObject` interface with `dimensionOrder` and `metrics` properties
   b. Create `MetricData` interface representing data for a single metric
   c. Create `DataPoint` interface for individual timestamp/value pairs
   d. Create `MetricMetadata` interface for metric definition information
   e. Create `TimeseriesMetadata` interface for the complete metadata structure
   f. Write unit tests for type guards and validation

4. Define custom error classes - Not Started
   a. Create `TimeseriesParseError` class for parsing failures
   b. Create `TimeseriesValidationError` class for data validation failures
   c. Create `TimeseriesFormatError` class for formatting failures
   d. Extend native Error class properly with correct prototype chain
   e. Write unit tests for error classes

5. Implement custom metadata parsing - Not Started
   a. Create `parseCustomMetadata(metadataJson: string): TimeseriesMetadata` function
   b. Use native JSON.parse with validation
   c. Extract metric definitions (dimension names, types, time grain, forecast length, data types)
   d. Extract dimension order from top-level `dimension_order` field
   e. Validate metadata structure and required fields
   f. Write unit tests for metadata parsing

6. Implement custom metadata formatting - Not Started
   a. Create `formatCustomMetadata(obj: TimeseriesObject): string` function
   b. Build metric definitions from in-memory representation
   c. Include dimension order in output
   d. Use JSON.stringify with proper formatting
   e. Write unit tests for metadata formatting

7. Implement table data parsing - Not Started
   a. Create `parseTableData(parquetData: Buffer, metadata: TimeseriesMetadata): TimeseriesObject` function
   b. Use chosen Parquet library to read Parquet data
   c. Iterate over rows and build multi-level object structure
   d. Use dimension order from metadata to traverse dimension values correctly
   e. Parse JSON-encoded values from the `value` column
   f. Convert timestamps to JavaScript Date objects
   g. Write unit tests for table parsing

8. Implement table data formatting - Not Started
   a. Create `formatTableData(obj: TimeseriesObject): RowData[]` function
   b. Flatten multi-level object to row structure
   c. Build dimension_values array for each row following dimensionOrder
   d. JSON-encode values for the `value` column
   e. Convert Date objects to timestamps
   f. Write unit tests for table formatting

9. Implement the TimeseriesParser class - Not Started
   a. Create `TimeseriesParser` class with `parse(filePath: string): Promise<TimeseriesObject>` method
   b. Create `parseFromBuffer(data: Buffer): Promise<TimeseriesObject>` overload for in-memory data
   c. Use Parquet library to read Parquet data
   d. Extract and parse custom metadata
   e. Parse table data using metadata
   f. Return populated `TimeseriesObject`
   g. Write unit tests for parser class

10. Implement the TimeseriesFormatter class - Not Started
    a. Create `TimeseriesFormatter` class with `format(obj: TimeseriesObject): Promise<Buffer>` method
    b. Create `formatToFile(obj: TimeseriesObject, filePath: string): Promise<void>` method
    c. Generate custom metadata JSON
    d. Generate table row data
    e. Use Parquet library to write Parquet data
    f. Write unit tests for formatter class

11. Add round-trip testing - Not Started
    a. Create test fixtures with sample time series data
    b. Test parse → format → parse round-trip preserves data
    c. Test format → parse → format round-trip preserves data
    d. Verify metadata is preserved through round-trips
    e. Test with multiple metrics and dimensions

12. Integration with VSCode extension - Not Started
    a. Create service class for time series operations
    b. Register parser/formatter with extension services
    c. Add file type recognition for .parquet files
    d. Create editor provider for viewing time series objects
    e. Implement tree view for time series object structure
    f. Write integration tests

## Decisions

1. **Parquet library selection**: Evaluate and select an appropriate Node.js Parquet library. Options include:
   - `parquetjs` - Pure JavaScript implementation
   - `apache-arrow` - Apache Arrow with Parquet support
   - Other libraries based on maturity, maintenance, and feature completeness

2. **In-memory representation structure**: The in-memory representation uses a multi-level object with the same key sequence as the Go implementation: Metric Name → Dimension Value(s) → Forecast/Actuals → Timestamp → Value. This ensures consistency across all ChronoCone components.

3. **Timestamp handling**: Timestamps in the in-memory representation are JavaScript `Date` objects. They are converted to/from Parquet TIMESTAMP type during formatting/parsing.

4. **Async API**: All parser and formatter methods should be async to support efficient file I/O in the VSCode extension host environment.

5. **Type safety**: Use strict TypeScript typing throughout. Define interfaces for all data structures and use type guards for runtime validation.

## Notes

- The parser and formatter should be symmetric with the Go and Kotlin implementations: parsing a file created by another implementation should produce an equivalent object.
- Consider using Web Workers for large file parsing to avoid blocking the VSCode extension host.
- Consider caching parsed objects in the extension to avoid redundant file reads during editing sessions.
- Future enhancements may include streaming support for very large time series objects.
- The VSCode extension API should be consulted for best practices on file handling and editor integration.
