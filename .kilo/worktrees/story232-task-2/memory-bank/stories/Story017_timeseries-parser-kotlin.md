# Story 017: Time Series Object Parser - Kotlin (IntelliJ Plugin)

## Overview

Implement a Kotlin parser and formatter for reading and writing time series object data in Parquet format, for use in the IntelliJ IDEA plugin. The parser uses the ParquetKt library to handle Parquet file I/O and produces the same in-memory representation as the Go implementation (multi-level map structure).

## References

- [Architecture](../specifications/architecture/)
- [Story015_timeseries-parser-go.md](Story015_timeseries-parser-go.md) - Go implementation for reference
- [Story016_intellij-workspace-setup.md](Story016_intellij-workspace-setup.md) - IntelliJ workspace setup

## Dependencies

### Prerequisites

- [Story016_intellij-workspace-setup.md](Story016_intellij-workspace-setup.md) - Not Started (requires IntelliJ plugin project structure)
- ParquetKt library dependency configured in the IntelliJ plugin project

### Dependents

None yet. Future stories for IntelliJ plugin time series operations will depend on this story.

## Tasks

1. Add ParquetKt dependency to IntelliJ plugin project - Not Started
   a. Add ParquetKt library to build.gradle.kts dependencies
   b. Verify dependency resolution and download
   c. Document required ParquetKt version in plugin documentation
   d. Write build verification test

2. Create the timeseries parser package structure - Not Started
   a. Create `src/main/kotlin/com/chronocone/plugin/timeseries/` package
   b. Create `TimeseriesParser.kt` for the parser class
   c. Create `TimeseriesFormatter.kt` for the formatter class
   d. Create `model/` subdirectory for data classes
   e. Create `exception/` subdirectory for custom exceptions

3. Define data classes for in-memory representation - Not Started
   a. Create `TimeseriesObject.kt` data class with `dimensionOrder` and `metrics` properties
   b. Create `MetricData.kt` data class representing data for a single metric
   c. Create `DataPoint.kt` data class for individual timestamp/value pairs
   d. Create `MetricMetadata.kt` data class for metric definition information
   e. Write unit tests for data class validation

4. Define custom exception classes - Not Started
   a. Create `TimeseriesParseException.kt` for parsing failures
   b. Create `TimeseriesValidationException.kt` for data validation failures
   c. Create `TimeseriesFormatException.kt` for formatting failures
   d. Write unit tests for exception classes

5. Implement custom metadata parsing - Not Started
   a. Create `parseCustomMetadata(metadataJson: String): TimeseriesMetadata` function
   b. Use kotlinx.serialization or Gson for JSON parsing
   c. Extract metric definitions (dimension names, types, time grain, forecast length, data types)
   d. Extract dimension order from top-level `dimension_order` field
   e. Validate metadata structure and required fields
   f. Write unit tests for metadata parsing

6. Implement custom metadata formatting - Not Started
   a. Create `formatCustomMetadata(obj: TimeseriesObject): String` function
   b. Build metric definitions from in-memory representation
   c. Include dimension order in output
   d. Ensure JSON output is valid and properly formatted
   e. Write unit tests for metadata formatting

7. Implement table data parsing with ParquetKt - Not Started
   a. Create `parseTableData(parquetFile: ParquetReader, metadata: TimeseriesMetadata): TimeseriesObject` function
   b. Use ParquetKt to read Parquet file
   c. Iterate over rows and build multi-level map structure
   d. Use dimension order from metadata to traverse dimension values correctly
   e. Parse JSON-encoded values from the `value` column using kotlinx.serialization
   f. Convert timestamps to Java LocalDateTime
   g. Write unit tests for table parsing

8. Implement table data formatting with ParquetKt - Not Started
   a. Create `formatTableData(obj: TimeseriesObject): List<ParquetRecord>` function
   b. Flatten multi-level map to row structure
   c. Build dimension_values list for each row following dimensionOrder
   d. JSON-encode values for the `value` column
   e. Convert LocalDateTime to timestamps
   f. Write unit tests for table formatting

9. Implement the TimeseriesParser class - Not Started
   a. Create `TimeseriesParser` class with `parse(file: File): TimeseriesObject` method
   b. Create `parse(bytes: ByteArray): TimeseriesObject` overload for in-memory data
   c. Use ParquetKt to read Parquet data
   d. Extract and parse custom metadata
   e. Parse table data using metadata
   f. Return populated `TimeseriesObject`
   g. Write unit tests for parser class

10. Implement the TimeseriesFormatter class - Not Started
    a. Create `TimeseriesFormatter` class with `format(obj: TimeseriesObject): ByteArray` method
    b. Create `formatToFile(obj: TimeseriesObject, file: File)` method
    c. Generate custom metadata JSON
    d. Generate table records
    e. Use ParquetKt to write Parquet data
    f. Write unit tests for formatter class

11. Add round-trip testing - Not Started
    a. Create test fixtures with sample time series data
    b. Test parse → format → parse round-trip preserves data
    c. Test format → parse → format round-trip preserves data
    d. Verify metadata is preserved through round-trips
    e. Test with multiple metrics and dimensions

12. Integration with IntelliJ plugin - Not Started
    a. Create service class for time series operations
    b. Register parser/formatter with plugin service registry
    c. Add file type recognition for .parquet files
    d. Create editor support for viewing time series objects
    e. Write integration tests

## Decisions

1. **ParquetKt library**: Use the ParquetKt library for Parquet I/O in Kotlin. This is a pure Kotlin implementation that integrates well with the IntelliJ platform.

2. **In-memory representation structure**: The in-memory representation uses a multi-level map with the same key sequence as the Go implementation: Metric Name → Dimension Value(s) → Forecast/Actuals → Timestamp → Value. This ensures consistency across all ChronoCone components.

3. **JSON serialization**: Use kotlinx.serialization for JSON parsing/formatting of the value column and custom metadata. This is the idiomatic Kotlin approach and integrates well with the IntelliJ platform.

4. **Timestamp handling**: Timestamps in the in-memory representation are Java `LocalDateTime` objects. They are converted to/from Parquet TIMESTAMP type during formatting/parsing.

5. **Thread safety**: The parser and formatter classes should be stateless and thread-safe to support concurrent use in the IDE.

## Notes

- The parser and formatter should be symmetric with the Go and Kotlin implementations: parsing a file created by another implementation should produce an equivalent object.
- The ParquetKt library documentation should be consulted for API details and best practices.
- Consider caching parsed objects in the IDE to avoid redundant file reads during editing sessions.
- Future enhancements may include streaming support for very large time series objects.
