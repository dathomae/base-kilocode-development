// Package metrics_arithmetic contains structural definitions for IO specification
// JSON files that correspond to the arithmetic example CPL files.
//
// Story: Story202 - CPL Metrics Arithmetic Examples Revamp
//
// Each example lives in its own subdirectory with an iospec.json that maps
// connector names to parquet file paths and declares output metrics.
// The shared infrastructure (concepts, types, units, measures)
// is defined in core.cpl, so these IO specs only reference the connectors
// that the example-specific CPL file provides.
package metrics_arithmetic

import (
	"github.com/chronocone/chronocone_planning_language/src/main/go/io"
)

// IOSpecDirectoryPath is the location where all arithmetic example subdirectories
// are stored. Each subdirectory contains an iospec.json file.
const IOSpecDirectoryPath = "src/chronocone_planning_language/examples/with_data/without_errors/metrics_arithmetic"

// ============================================================================
// Shared architecture defaults across all arithmetic example IO specs
// ============================================================================

// DefaultArchitecture is the standard compute architecture specification used
// for all arithmetic examples. These values represent modest defaults suitable
// for small self-contained metric calculations.
//
// Fields:
//   - CPUCoreCount: 4 cores
//   - CacheSizeBytes: 8 MiB (8,388,608 bytes)
//   - MainMemoryBytes: 8 GiB (8,589,934,592 bytes)
var DefaultArchitecture = io.ArchitectureSpecification{
	CPUCoreCount:    4,
	CacheSizeBytes:  8388608,
	MainMemoryBytes: 8589934592,
}

// fileDriverPath creates a single driver parameter map with a file path.
func fileDriverPath(path string) []map[string]interface{} {
	return []map[string]interface{}{{"path": path}}
}

// inputSpec creates a ConnectorSpecification for a parquet file connector.
func inputSpec(connectorName, parquetPath string) io.ConnectorSpecification {
	return io.ConnectorSpecification{
		ConnectorName:    connectorName,
		DataFormat:       "parquet",
		Driver:           "file",
		DriverParameters: fileDriverPath(parquetPath),
	}
}

// outputSpec creates an OutputSpecification with the standard arithmetic
// example shape (dimensions: [Region], celltype: packages).
// The grain parameter accepts a slice of grain measure names so each
// example can specify its own grain (e.g., []string{"Daily"}) instead of
// hardcoding a single value.
func outputSpec(metricName string, grain []string) io.OutputSpecification {
	return io.OutputSpecification{
		MetricName: metricName,
		Dimensions: []string{"Region"},
		CellType:   "packages",
		Grain:      grain,
	}
}

// ============================================================================
// File 1: simple_addition/iospec.json
// ============================================================================
// Purpose: IO specification for the simple addition CPL example.
// Corresponding CPL: simple_addition/simple_addition.cpl
// Output metric: sum_metric
// Input connectors: input_a, input_b (defined in the example .cpl file)

// SimpleAdditionIOSpecJSON is the relative path to the simple addition
// IO spec file, from the metrics_arithmetic package directory.
const SimpleAdditionIOSpecJSON = "simple_addition/iospec.json"

// SimpleAdditionIOSpec returns the IO specification for the simple addition
// example. Two input connectors (input_a, input_b) with parquet file inputs,
// one output metric (sum_metric).
func SimpleAdditionIOSpec() *io.IOSpecification {
	return &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			inputSpec("input_a", "data_a.parquet"),
			inputSpec("input_b", "data_b.parquet"),
		},
		Outputs: []io.OutputSpecification{
			outputSpec("sum_metric", []string{"Daily"}),
		},
		Architecture: DefaultArchitecture,
	}
}

// ============================================================================
// File 2: multiplication_precedence/iospec.json
// ============================================================================
// Purpose: IO specification for the multiplication precedence CPL example.
// Corresponding CPL: multiplication_precedence/multiplication_precedence.cpl
// Output metric: precedence_metric
// Input connectors: input_a, input_b, input_c (defined in the example .cpl file)

// MultiplicationPrecedenceIOSpecJSON is the relative path to the
// multiplication precedence IO spec file.
const MultiplicationPrecedenceIOSpecJSON = "multiplication_precedence/iospec.json"

// MultiplicationPrecedenceIOSpec returns the IO specification for the
// multiplication precedence example (3 inputs: input_a, input_b, input_c).
func MultiplicationPrecedenceIOSpec() *io.IOSpecification {
	return &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			inputSpec("input_a", "data_a.parquet"),
			inputSpec("input_b", "data_b.parquet"),
			inputSpec("input_c", "data_c.parquet"),
		},
		Outputs: []io.OutputSpecification{
			outputSpec("precedence_metric", []string{"Daily"}),
		},
		Architecture: DefaultArchitecture,
	}
}

// ============================================================================
// File 3: subtraction_division/iospec.json
// ============================================================================
// Purpose: IO specification for the subtraction and division CPL example.
// Corresponding CPL: subtraction_division/subtraction_division.cpl
// Output metric: sub_div_metric
// Input connectors: input_a, input_b, input_c (defined in the example .cpl file)

// SubtractionDivisionIOSpecJSON is the relative path to the subtraction
// and division IO spec file.
const SubtractionDivisionIOSpecJSON = "subtraction_division/iospec.json"

// SubtractionDivisionIOSpec returns the IO specification for the subtraction
// and division example (3 inputs: input_a, input_b, input_c).
func SubtractionDivisionIOSpec() *io.IOSpecification {
	return &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			inputSpec("input_a", "data_a.parquet"),
			inputSpec("input_b", "data_b.parquet"),
			inputSpec("input_c", "data_c.parquet"),
		},
		Outputs: []io.OutputSpecification{
			outputSpec("sub_div_metric", []string{"Daily"}),
		},
		Architecture: DefaultArchitecture,
	}
}

// ============================================================================
// File 4: unary_negation/iospec.json
// ============================================================================
// Purpose: IO specification for the unary negation CPL example.
// Corresponding CPL: unary_negation/unary_negation.cpl
// Output metric: negate_metric
// Input connectors: input_a, input_b (defined in the example .cpl file)

// UnaryNegationIOSpecJSON is the relative path to the unary negation
// IO spec file.
const UnaryNegationIOSpecJSON = "unary_negation/iospec.json"

// UnaryNegationIOSpec returns the IO specification for the unary negation
// example (2 inputs: input_a, input_b).
func UnaryNegationIOSpec() *io.IOSpecification {
	return &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			inputSpec("input_a", "data_a.parquet"),
			inputSpec("input_b", "data_b.parquet"),
		},
		Outputs: []io.OutputSpecification{
			outputSpec("negate_metric", []string{"Daily"}),
		},
		Architecture: DefaultArchitecture,
	}
}

// ============================================================================
// File 5: parenthesized/iospec.json
// ============================================================================
// Purpose: IO specification for the parenthesized expression CPL example.
// Corresponding CPL: parenthesized/parenthesized.cpl
// Output metric: paren_metric
// Input connectors: input_a, input_b, input_c, input_d (defined in the example .cpl file)

// ParenthesizedIOSpecJSON is the relative path to the parenthesized
// expression IO spec file.
const ParenthesizedIOSpecJSON = "parenthesized/iospec.json"

// ParenthesizedIOSpec returns the IO specification for the parenthesized
// expression example (4 inputs: input_a, input_b, input_c, input_d).
func ParenthesizedIOSpec() *io.IOSpecification {
	return &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			inputSpec("input_a", "data_a.parquet"),
			inputSpec("input_b", "data_b.parquet"),
			inputSpec("input_c", "data_c.parquet"),
			inputSpec("input_d", "data_d.parquet"),
		},
		Outputs: []io.OutputSpecification{
			outputSpec("paren_metric", []string{"Daily"}),
		},
		Architecture: DefaultArchitecture,
	}
}

// ============================================================================
// File 6: complex_expression/iospec.json
// ============================================================================
// Purpose: IO specification for the complex nested expression CPL example.
// Corresponding CPL: complex_expression/complex_expression.cpl
// Output metric: complex_metric
// Input connectors: input_a, input_b, input_c, input_d, input_e
//   (defined in the example .cpl file)

// ComplexExpressionIOSpecJSON is the relative path to the complex
// expression IO spec file.
const ComplexExpressionIOSpecJSON = "complex_expression/iospec.json"

// ComplexExpressionIOSpec returns the IO specification for the complex
// expression example (5 inputs: input_a, input_b, input_c, input_d, input_e).
func ComplexExpressionIOSpec() *io.IOSpecification {
	return &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			inputSpec("input_a", "data_a.parquet"),
			inputSpec("input_b", "data_b.parquet"),
			inputSpec("input_c", "data_c.parquet"),
			inputSpec("input_d", "data_d.parquet"),
			inputSpec("input_e", "data_e.parquet"),
		},
		Outputs: []io.OutputSpecification{
			outputSpec("complex_metric", []string{"Daily"}),
		},
		Architecture: DefaultArchitecture,
	}
}

// ============================================================================
// File 7: short_form_addition/iospec.json
// ============================================================================
// Purpose: IO specification for the short form addition CPL example.
// Corresponding CPL: short_form_addition/short_form_addition.cpl
// Output metric: sum_metric
// Input connectors: input_a, input_b (defined in the example .cpl file)

// ShortFormAdditionIOSpecJSON is the relative path to the short form addition
// IO spec file, from the metrics_arithmetic package directory.
const ShortFormAdditionIOSpecJSON = "short_form_addition/iospec.json"

// ShortFormAdditionIOSpec returns the IO specification for the short form addition
// example. Two input connectors (input_a, input_b) with parquet file inputs,
// one output metric (sum_metric).
func ShortFormAdditionIOSpec() *io.IOSpecification {
	return &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			inputSpec("input_a", "data_a.parquet"),
			inputSpec("input_b", "data_b.parquet"),
		},
		Outputs: []io.OutputSpecification{
			outputSpec("sum_metric", []string{"Daily"}),
		},
		Architecture: DefaultArchitecture,
	}
}

// ============================================================================
// Shared structural patterns across all IO spec definitions
// ============================================================================
//
// Each IO spec follows the same overall structure:
//   1. inputs — one InputSpecification per connector, mapping it to a
//      parquet file via the file driver. Connectors are defined in the
//      example-specific CPL file.
//   2. outputs — a single OutputSpecification with the metric name,
//      dimensions ["Region"], celltype "packages", and grain ["Daily"].
//   3. architecture — DefaultArchitecture with 4 CPU cores, 8 MiB cache,
//      and 8 GiB main memory.
//
// The only varying fields across definitions are the output metric name
// (which must match the metric declared in the corresponding CPL file)
// and the number of input connectors.
