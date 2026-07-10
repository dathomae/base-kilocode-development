// Package metrics_conditionals contains structural definitions for IO specification
// JSON files that correspond to the conditional example CPL files.
//
// Story: Story188 - CPL Metrics Conditionals Synthesize
package metrics_conditionals

import (
	"github.com/chronocone/chronocone_planning_language/src/main/go/io"
)

// IOSpecDirectoryPath is the location where all conditional example subdirectories
// are stored. Each subdirectory contains an iospec.json file.
const IOSpecDirectoryPath = "src/chronocone_planning_language/examples/with_data/without_errors/metrics_conditionals"

// DefaultArchitecture is the standard compute architecture specification used
// for all conditional examples.
var DefaultArchitecture = io.ArchitectureSpecification{
	CPUCoreCount:    4,
	CacheSizeBytes:  8388608,
	MainMemoryBytes: 8589934592,
}

func fileDriverPath(path string) []map[string]interface{} {
	return []map[string]interface{}{{"path": path}}
}

func inputSpec(connectorName, parquetPath string) io.ConnectorSpecification {
	return io.ConnectorSpecification{
		ConnectorName:    connectorName,
		DataFormat:       "parquet",
		Driver:           "file",
		DriverParameters: fileDriverPath(parquetPath),
	}
}

func outputSpec(metricName string, grain []string) io.OutputSpecification {
	return io.OutputSpecification{
		MetricName: metricName,
		Dimensions: []string{"Region"},
		CellType:   "packages",
		Grain:      grain,
	}
}

func ConditionalIfElseArithmeticIOSpec() *io.IOSpecification {
	return &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			inputSpec("data_a", "data_a.parquet"),
			inputSpec("data_b", "data_b.parquet"),
		},
		Outputs: []io.OutputSpecification{
			outputSpec("conditional_metric", []string{"Daily"}),
		},
		Architecture: DefaultArchitecture,
	}
}

func ConditionalIfElseifElseIOSpec() *io.IOSpecification {
	return &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			inputSpec("data_a", "data_a.parquet"),
			inputSpec("data_b", "data_b.parquet"),
		},
		Outputs: []io.OutputSpecification{
			outputSpec("conditional_metric", []string{"Daily"}),
		},
		Architecture: DefaultArchitecture,
	}
}

func ConditionalWhenDefaultIOSpec() *io.IOSpecification {
	return &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			inputSpec("data_a", "data_a.parquet"),
			inputSpec("data_b", "data_b.parquet"),
		},
		Outputs: []io.OutputSpecification{
			outputSpec("conditional_metric", []string{"Daily"}),
		},
		Architecture: DefaultArchitecture,
	}
}

func ConditionalComparisonOperatorsIOSpec() *io.IOSpecification {
	return &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			inputSpec("data_a", "data_a.parquet"),
			inputSpec("data_b", "data_b.parquet"),
		},
		Outputs: []io.OutputSpecification{
			outputSpec("comparison_metric", []string{"Daily"}),
		},
		Architecture: DefaultArchitecture,
	}
}

func ConditionalLogicalOperatorsIOSpec() *io.IOSpecification {
	return &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			inputSpec("data_a", "data_a.parquet"),
			inputSpec("data_b", "data_b.parquet"),
		},
		Outputs: []io.OutputSpecification{
			outputSpec("conditional_metric", []string{"Daily"}),
		},
		Architecture: DefaultArchitecture,
	}
}

func ConditionalUnaryNegationIOSpec() *io.IOSpecification {
	return &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			inputSpec("data_a", "data_a.parquet"),
			inputSpec("data_b", "data_b.parquet"),
		},
		Outputs: []io.OutputSpecification{
			outputSpec("unary_metric", []string{"Daily"}),
		},
		Architecture: DefaultArchitecture,
	}
}

func ConditionalShortFormIOSpec() *io.IOSpecification {
	return &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			inputSpec("data_a", "data_a.parquet"),
			inputSpec("data_b", "data_b.parquet"),
		},
		Outputs: []io.OutputSpecification{
			outputSpec("short_cond_metric", []string{"Daily"}),
		},
		Architecture: DefaultArchitecture,
	}
}

func ConditionalTruthinessIOSpec() *io.IOSpecification {
	return &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			inputSpec("data_a", "data_a.parquet"),
			inputSpec("data_b", "data_b.parquet"),
		},
		Outputs: []io.OutputSpecification{
			outputSpec("truthiness_metric", []string{"Daily"}),
		},
		Architecture: DefaultArchitecture,
	}
}

func ConditionalTautologicalIfOnlyIOSpec() *io.IOSpecification {
	return &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			inputSpec("data_a", "data_a.parquet"),
			inputSpec("data_b", "data_b.parquet"),
		},
		Outputs: []io.OutputSpecification{
			outputSpec("tautology_metric", []string{"Daily"}),
		},
		Architecture: DefaultArchitecture,
	}
}

func ConditionalNestedArithmeticIOSpec() *io.IOSpecification {
	return &io.IOSpecification{
		Connectors: []io.ConnectorSpecification{
			inputSpec("data_a", "data_a.parquet"),
			inputSpec("data_b", "data_b.parquet"),
		},
		Outputs: []io.OutputSpecification{
			outputSpec("conditional_nested_arithmetic_metric", []string{"Daily"}),
		},
		Architecture: DefaultArchitecture,
	}
}
