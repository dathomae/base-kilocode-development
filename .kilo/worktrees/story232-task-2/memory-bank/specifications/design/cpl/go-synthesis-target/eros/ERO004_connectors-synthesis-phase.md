# ERO004: Connectors Synthesis Phase

## Title

Connectors Synthesis Phase

## Purpose/Goal

Define the synthesis phase logic for generating Go code that reads data from a parquet file using Apache Arrow based on a connector declaration and an IO specification.

## Implementation Status

Not Yet Implemented

**Related Epics:**

- [Epic004_cpl-connectors-implementation.md](../../../../../epics/Epic004_cpl-connectors-implementation.md) (To be created)

## Context

### Situation

The synthesis phase receives the CIL representation of a connector (via `CILConnectorReferenceNode`) and an IO specification that details how to connect to the data source. For this increment, the data source is restricted to a parquet file, and the data must be read using the `go-arrow` library. The analysis phase now provides an ordered list of grain measures for each connector, and output IO specifications represent grain as an ordered array of grain names.

### Constraints

- The generated Go code must use the `github.com/apache/arrow/go/v16/parquet` (or appropriate version) library to read the parquet file.
- The generated code must map the parquet columns to the dimensions, celltype, and ordered grain measures specified in the connector declaration.
- The data must be loaded into Chronocone's internal tensor representation.

### Requirements

- Extract connection details (file path) from the IO specification.
- Generate Go code that opens the parquet file.
- Generate Go code that reads the parquet schema and validates it against the connector's dimensions, celltype, and ordered grain measures.
- Generate Go code that iterates over the parquet row groups/records and populates a Chronocone tensor.

## Evaluation Criteria

1. **Correctness** (High importance): The generated code must correctly read the parquet file and populate the tensor.
2. **Performance** (Medium importance): The generated code should read the data efficiently, leveraging Arrow's columnar format where possible.

## Decision

**Chosen Option:** Generate Go code that uses `go-arrow/parquet` to read data into tensors.

**Rationale:**
Apache Arrow provides efficient columnar data processing. By generating code that uses `go-arrow`, we can efficiently load parquet data into our internal tensor structures.

### Increment 1: Basic Parquet Reading

**Features Implemented:**

- **IO Specification Integration:**
  - The synthesizer will look up the connector name in the provided IO specification to find the parquet file path.
- **Code Generation:**
  - Generate imports for `github.com/apache/arrow/go/v16/parquet/file` and `github.com/apache/arrow/go/v16/parquet/pqarrow`.
  - Generate a single, reusable library function `loadParquetConnector(controller *ChronoconeComputationController, spec *InputSpecification, meta ConnectorMetadata) (*ChronoconeTensor[float64], error)` that reads any parquet file based on the provided metadata.
  - Generate a map declaration keyed by connector name containing the metadata (dimensions, cell type) for each connector defined in the CPL.

### Synthesis Types and Functions

To support generating the connector loading code, the synthesis phase will introduce the following types and functions:

#### ConnectorMetadata

This struct holds the information needed at runtime to load a connector. It is populated during the CIL tree traversal when a `CILConnectorReferenceNode` is encountered and synthesized into a map in the generated code. During the analysis phase, a CIL tree is built for each connector and metric, stored in a map keyed by name. For ERO004, the synthesizer will iterate through this map and generate code for everything it finds (which will be the connectors).

```go
// ConnectorMetadata holds the runtime information needed to load a connector
type ConnectorMetadata struct {
    Name       string
    Dimensions []string
    CellType   string // e.g., "float64", "int64"
    // Grain      string // Added in Increment 2
}
```

#### CILAbstractTransformer Updates

The `CILAbstractTransformer` will be updated to track connectors and generate their metadata map.

```go
type CILAbstractTransformer struct {
    // ... existing fields ...
    connectors map[string]*ConnectorMetadata
}

// RegisterConnector registers a connector's metadata.
// Called when visiting a CILConnectorReferenceNode.
func (t *CILAbstractTransformer) RegisterConnector(name string, dimensions []string, cellType string)

// GenerateConnectorMetadataMap generates the map declaration for all registered connectors.
func (t *CILAbstractTransformer) GenerateConnectorMetadataMap() string

// GenerateParquetLoaderLibrary generates the single reusable loadParquetConnector function.
func (t *CILAbstractTransformer) GenerateParquetLoaderLibrary() string
```

### Generated Go Code Structure

The generated code will include a metadata map, the reusable loader function, and the main function integration:

```go
// 1. Generated Connector Metadata Map
var connectorMetadata = map[string]ConnectorMetadata{
    "sales_data": {
        Name:       "sales_data",
        Dimensions: []string{"Product", "Region"},
        CellType:   "float64",
    },
    // ... other connectors ...
}

// 2. Reusable Parquet Loader Function
func loadParquetConnector(
    controller *ChronoconeComputationController,
    spec *InputSpecification,
    meta ConnectorMetadata,
) (*ChronoconeTensor[float64], error) {
    // Extract file path from spec.DriverParameters
    var filePath string
    for _, param := range spec.DriverParameters {
        if path, ok := param["path"].(string); ok {
            filePath = path
            break
        }
    }
    if filePath == "" {
        return nil, fmt.Errorf("missing 'path' parameter for file driver")
    }

    // Use the standard library function to read the parquet file
    // This function is provided by the Chronocone runtime library
    recordReader, err := parquetutil.ReadParquetFile(filePath)
    if err != nil {
        return nil, err
    }
    defer recordReader.Release()

    // Allocate Tensor based on metadata dimensions
    tensor, err := controller.AllocateTensor[float64](meta.Dimensions...)
    if err != nil {
        return nil, err
    }

    // Build value-to-index maps for dimensions to optimize flat index calculation
    dimMaps := make(map[string]map[string]int)
    for _, dim := range meta.Dimensions {
        dimMap := make(map[string]int)
        for i, val := range controller.conceptValues[dim] {
            dimMap[val.(string)] = i
        }
        dimMaps[dim] = dimMap
    }

    // Iterate over Arrow records and populate Tensor
    schema := recordReader.Schema()

    for recordReader.Next() {
        rec := recordReader.Record()

        // Find column indices dynamically based on metadata
        dimCols := make([]*array.String, len(meta.Dimensions))
        for i, dim := range meta.Dimensions {
            idx := schema.FieldIndices(dim)[0]
            dimCols[i] = rec.Column(idx).(*array.String)
        }

        // Assuming the value column is named "Value" or similar
        // In a real implementation, this would be mapped properly
        valIdx := schema.FieldIndices("Value")[0]
        valueCol := rec.Column(valIdx).(*array.Float64)

        for row := 0; row < int(rec.NumRows()); row++ {
            flatIndex := 0
            multiplier := 1
            validRow := true

            // Calculate flat index dynamically
            for i := len(meta.Dimensions) - 1; i >= 0; i-- {
                dim := meta.Dimensions[i]
                dimVal := dimCols[i].Value(row)

                dimIdx, ok := dimMaps[dim][dimVal]
                if !ok {
                    validRow = false
                    break
                }

                flatIndex += dimIdx * multiplier
                multiplier *= len(controller.conceptValues[dim])
            }

            if !validRow {
                continue
            }

            cellVal := valueCol.Value(row)
            tensor.data[flatIndex] = cellVal
        }
    }

    if err := recordReader.Err(); err != nil {
        return nil, err
    }

    return tensor, nil
}
```

### Main Function Integration

The generated `main` function will use the metadata map and the reusable loader function based on the `InputSpecification`:

```go
    // Load input tensors from specified connectors
    inputs := make(map[string]*ChronoconeTensor[float64])
    for _, inputSpec := range ioSpec.Inputs {
        meta, ok := connectorMetadata[inputSpec.ConnectorName]
        if !ok {
            log.Fatalf("Unknown connector in IO spec: %s", inputSpec.ConnectorName)
        }

        tensor, err := loadParquetConnector(controller, &inputSpec, meta)
        if err != nil {
            log.Fatalf("Failed to load input %s: %v", inputSpec.ConnectorName, err)
        }
        inputs[inputSpec.ConnectorName] = tensor
    }
```

### Output Writing

The generated `main` function also writes output results based on the `OutputSpecification`. Each output specification includes a `Grain` field that is an ordered slice of grain names (`[]string`).

> **Note:** Full synthesizer multi-grain output generation is not completed in Story176 and is deferred to [Story160_cpl-synthesizer-connectors-time.md](../../../../../../stories/Story160_cpl-synthesizer-connectors-time.md). As a temporary compatibility measure, the current synthesis code uses only the first element of the grain slice when generating output-writing text. This ensures compilation succeeds for single-grain and multi-grain declarations alike, but generated output-writing code does not yet fully support multiple grain dimensions.

**Limitations:**

- Only supports local file paths for parquet files (no S3/GCS yet).
- Assumes the parquet schema exactly matches the connector declaration (column names match dimension/grain/celltype names).
- Multi-grain output generation is not yet implemented. The synthesis code uses only the first grain element for output-writing text as a temporary compatibility measure; full support for multiple grain dimensions in generated output-writing code is deferred to [Story160_cpl-synthesizer-connectors-time.md](../../../../../../stories/Story160_cpl-synthesizer-connectors-time.md).
- Does not handle complex data type conversions (assumes parquet types map cleanly to Chronocone types).

**Value Provided:**

The synthesized program can ingest external data from parquet files, enabling the execution of plans that depend on external data.

## Implications

### Positive Implications

- Leverages industry-standard libraries (Arrow/Parquet) for efficient data ingestion.
- Provides a clear pattern for adding other data sources in the future.
- Generating a single reusable loader function and a metadata map reduces the size of the generated code and avoids duplicating parquet reading logic for every connector.

### Negative Implications / Trade-offs

- Adds a dependency on `go-arrow` to the synthesized code.
- The generic parquet reader might need to do a bit more dynamic work (e.g., looking up column indices by name dynamically) compared to a hardcoded reader, but Arrow handles this efficiently.

## Notes

**Related Documents:**

- [CPL Requirements - Connectors](../../../../requirements/cpl/connectors.md)
- [Go Arrow Parquet API Resources](../../../../../resources/apis/go-arrow-parquet/overview.md)
