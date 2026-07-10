// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package connector

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/chronocone/chronocone_planning_language/src/main/go/io"
)

// LoaderFuncDecl holds the structured representation of a generated connector
// loader function. This replaces the raw strings.Builder approach for building
// individual connector loader functions, enabling:
//   - Independent testing of function name, params, and body
//   - Cleaner separation of function structure from Go source rendering
//   - Deterministic output through typed field values
type LoaderFuncDecl struct {
	FuncName            string   // Generated function name (e.g., "loadSalesData")
	ConnectorName       string   // Original connector name
	Driver              string   // Driver type (e.g., "file")
	Dimensions          []string // Connector dimension names
	HasConditionalMasks bool     // Whether to include validity mask application
}

// Render generates the Go source code for the connector loader function.
// The output is byte-identical to the pre-refactoring GenerateSingleConnectorLoader
// for the same inputs.
//
// Returns:
//   - string: The rendered Go source code for the loader function.
func (l *LoaderFuncDecl) Render() string {
	var sb strings.Builder

	// Comment and function signature
	sb.WriteString(fmt.Sprintf("// %s loads data for connector %q.\n", l.FuncName, l.ConnectorName))
	sb.WriteString(fmt.Sprintf("func %s(connectorSpec cplio.ConnectorSpecification, conceptValues map[string][]string) (*calculation.ChronoconeTensor[any], error) {\n",
		l.FuncName))

	// Function body
	sb.WriteString(fmt.Sprintf("\t// Load data for connector %q\n", l.ConnectorName))
	sb.WriteString(fmt.Sprintf("\t// Driver: %s\n", l.Driver))
	sb.WriteString(fmt.Sprintf("\t// Dimensions: %v\n", l.Dimensions))
	sb.WriteString(fmt.Sprintf("\tmetadata := connectorMetadata[%q]\n", l.ConnectorName))
	sb.WriteString("\ttensor, err := loadParquetConnector[any](metadata, connectorSpec, conceptValues)\n")
	sb.WriteString("\tif err != nil {\n")
	sb.WriteString("\t\treturn nil, err\n")
	sb.WriteString("\t}\n")
	if l.HasConditionalMasks {
		sb.WriteString(fmt.Sprintf("\tif mask, ok := connectorValidityMasks[%q]; ok {\n", l.ConnectorName))
		sb.WriteString("\t\ttensor.SetMask(mask)\n")
		sb.WriteString("\t}\n")
	}
	sb.WriteString("\treturn tensor, nil\n")
	sb.WriteString("}")

	return sb.String()
}

// GenerateConnectorLoaders generates Go code for loading all connector data.
// This function orchestrates the generation of individual loader functions
// for each connector defined in the CIL trees.
//
// Parameters:
//   - provider: A ConnectorProvider containing connector metadata.
//   - ioSpec: The IO specification with input data sources.
//   - hasConditionalMasks: Whether connectors with conditional grain dimensions
//     should have validity masks applied from connectorValidityMasks after loading.
//
// Returns:
//   - string: The generated Go code for all connector loaders.
//   - error: An error if code generation fails.
func GenerateConnectorLoaders(provider ConnectorProvider, ioSpec *io.IOSpecification, hasConditionalMasks bool) (string, error) {
	if provider == nil {
		return "", fmt.Errorf("provider is nil")
	}
	if ioSpec == nil {
		return "", fmt.Errorf("ioSpec is nil")
	}

	var sb strings.Builder

	// Generate the reusable parquet loader function
	sb.WriteString(GenerateLoadParquetConnectorFunction())
	sb.WriteString("\n\n")

	// Generate individual loader functions for each connector
	for _, connectorSpec := range ioSpec.Connectors {
		metadata := provider.GetConnectorMetadata(connectorSpec.ConnectorName)
		if metadata == nil {
			return "", fmt.Errorf("no metadata registered for connector %q", connectorSpec.ConnectorName)
		}

		loaderCode, err := GenerateSingleConnectorLoader(connectorSpec.ConnectorName, metadata, &connectorSpec, hasConditionalMasks)
		if err != nil {
			return "", err
		}
		sb.WriteString(loaderCode)
		sb.WriteString("\n\n")
	}

	return sb.String(), nil
}

// GenerateSingleConnectorLoader generates Go code for loading a specific connector.
// This creates a function that reads data from a parquet file and populates
// a calculation.ChronoconeTensor based on the connector's metadata.
// Generated code uses package-qualified type names (e.g., cplio.ConnectorSpecification,
// calculation.ChronoconeTensor[T]).
//
// The implementation builds a LoaderFuncDecl struct and renders it, replacing
// the previous raw strings.Builder approach with a structured intermediate
// representation.
//
// Parameters:
//   - connectorName: The name of the connector to generate a loader for.
//   - metadata: The runtime metadata for the connector.
//   - connectorSpec: The connector specification with data source details.
//   - hasConditionalMasks: Whether the generated loader should look up
//     connectorValidityMasks and apply the mask via tensor.SetMask.
//
// Returns:
//   - string: The generated Go code for the connector loader.
//   - error: An error if code generation fails.
func GenerateSingleConnectorLoader(
	connectorName string,
	metadata *ConnectorMetadata,
	connectorSpec *io.ConnectorSpecification,
	hasConditionalMasks bool,
) (string, error) {
	if metadata == nil {
		return "", fmt.Errorf("metadata is nil")
	}
	if connectorSpec == nil {
		return "", fmt.Errorf("connectorSpec is nil")
	}

	decl := &LoaderFuncDecl{
		FuncName:            GenerateConnectorLoaderFunctionName(connectorName),
		ConnectorName:       connectorName,
		Driver:              connectorSpec.Driver,
		Dimensions:          metadata.Dimensions,
		HasConditionalMasks: hasConditionalMasks,
	}

	return decl.Render(), nil
}

// GenerateLoadParquetConnectorFunction generates the reusable loadParquetConnector function.
// Generated code uses package-qualified type names (e.g., connector.ConnectorMetadata,
// cplio.ConnectorSpecification, calculation.ChronoconeTensor[T], calculation.NewChronoconeTensor).
//
// The generated function delegates to cplio.ReadParquetToTensor to read parquet data and
// returns a *calculation.ChronoconeTensor[T]. It performs the following steps:
//  1. Extracts filePath from connectorSpec.DriverParameters.
//  2. Calls cplio.ReadParquetToTensor(filePath, columnNames, columnToConcept, conceptValues)
//     which returns ([]float64, []string, []int, error).
//  3. Handles any error from the parquet read.
//  4. Converts []float64 to []T using a type assertion loop over each value.
//  5. Returns calculation.NewChronoconeTensor(typedData, dims, shape).
//
// Returns:
//   - string: The generated Go code for the function.
func GenerateLoadParquetConnectorFunction() string {
	return `// loadParquetConnector loads data from a parquet file based on connector metadata.
// This generic function handles parquet reading for any connector type.
// It builds a columnToConcept mapping and delegates to cplio.ReadParquetToTensor
// to read the parquet file and construct the raw tensor data, then converts
// []float64 to []T via type assertion.
func loadParquetConnector[T any](metadata connector.ConnectorMetadata, connectorSpec cplio.ConnectorSpecification, conceptValues map[string][]string) (*calculation.ChronoconeTensor[T], error) {
	var filePath string
	for _, params := range connectorSpec.DriverParameters {
		if path, ok := params["path"].(string); ok && path != "" {
			filePath = path
			break
		}
	}
	if filePath == "" {
		return nil, fmt.Errorf("no file path specified for connector %s", metadata.Name)
	}
	
	// Build column-to-concept mapping.
	// Regular dimensions map to themselves (identity).
	// Grain measures default to identity mapping (no entry needed).
	columnNames := make([]string, 0, len(metadata.Dimensions)+len(metadata.Grain))
	columnNames = append(columnNames, metadata.Dimensions...)
	columnNames = append(columnNames, metadata.Grain...)
	columnToConcept := make(map[string]string, len(columnNames))
	for _, dim := range metadata.Dimensions {
		columnToConcept[dim] = dim
	}
	
	data, dims, shape, err := cplio.ReadParquetToTensor(filePath, columnNames, columnToConcept, conceptValues)
	if err != nil {
		return nil, fmt.Errorf("failed to read parquet: %w", err)
	}
	
	typedData := make([]T, len(data))
	for i, v := range data {
		if t, ok := any(v).(T); ok {
			typedData[i] = t
		} else {
			return nil, fmt.Errorf("value %v cannot be converted to type %T", v, *new(T))
		}
	}
	
	return calculation.NewChronoconeTensor(typedData, dims, shape), nil
}`
}

// GenerateConnectorLoaderFunctionName generates the function name for a connector loader.
//
// Parameters:
//   - connectorName: The name of the connector.
//
// Returns:
//   - string: The function name (e.g., "loadSalesData").
func GenerateConnectorLoaderFunctionName(connectorName string) string {
	if connectorName == "" {
		return "loadConnector"
	}

	// Split on underscores and convert to camelCase
	parts := strings.Split(connectorName, "_")
	var sb strings.Builder
	sb.WriteString("load")

	for _, part := range parts {
		if part == "" {
			continue
		}
		// Capitalize first letter of each part
		runes := []rune(part)
		if len(runes) > 0 {
			sb.WriteRune(unicode.ToUpper(runes[0]))
			if len(runes) > 1 {
				sb.WriteString(string(runes[1:]))
			}
		}
	}

	return sb.String()
}
