// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package synthesis

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	cio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
)

// generateLookupLoading emits Go source code for reading lookup data from
// parquet files and registering each lookup as a MetricComputation on the
// ChronoconeComputationController.
//
// For each lookup registered in the CIL transformer's registry, the function
// finds the matching entry in the lookups slice to obtain the parquet file
// path, then emits code that:
//   - Calls cplio.ReadParquetToTensor with the file path and dimension names
//   - Constructs a calculation.ChronoconeTensor from the returned data
//   - Creates a calculation.MetricComputation, sets the tensor, and registers
//     it on the controller via RegisterMetricComputation
//
// When no lookups are declared in the registry, a no-op loadLookupData
// function that returns nil is emitted.
//
// Parameters:
//   - transformer: The CIL abstract transformer containing lookup metadata
//     in its embedded CILRegistry.
//   - lookups: The slice of lookup specifications containing their source
//     file paths in DriverParameters.
//   - w: The writer to emit generated code to.
//
// Returns:
//   - error: An error if a lookup metadata entry has no matching IO spec entry,
//     or if writing to w fails.
func generateLookupLoading(transformer *cil.CILAbstractTransformer, lookups []cio.LookupSpecification, w io.Writer) error {
	if transformer == nil {
		return fmt.Errorf("transformer is nil")
	}

	// Build a map from lookup name to source file path for fast lookup.
	lookupPaths := make(map[string]string)
	for _, lookup := range lookups {
		if len(lookup.DriverParameters) > 0 {
			if path, ok := lookup.DriverParameters[0]["path"].(string); ok {
				lookupPaths[lookup.LookupName] = path
			}
		}
	}

	allMetadata := transformer.AllLookupMetadata()

	var sb strings.Builder

	sb.WriteString("// loadLookupData reads lookup data from parquet files and registers\n")
	sb.WriteString("// each lookup as a MetricComputation on the controller.\n")
	sb.WriteString("func loadLookupData(controller *calculation.ChronoconeComputationController, dataDir *string) error {\n")

	if len(allMetadata) == 0 {
		sb.WriteString("\treturn nil\n")
		sb.WriteString("}\n")
		_, err := w.Write([]byte(sb.String()))
		return err
	}

	// Sort lookup names for deterministic output.
	lookupNames := make([]string, 0, len(allMetadata))
	for name := range allMetadata {
		lookupNames = append(lookupNames, name)
	}
	sort.Strings(lookupNames)

	for _, name := range lookupNames {
		metadata := allMetadata[name]
		path, ok := lookupPaths[name]
		if !ok {
			return fmt.Errorf("no IO spec entry found for lookup %q", name)
		}

		dimsLiteral := formatStringSliceLiteral(metadata.Dimensions)

		sb.WriteString(fmt.Sprintf("\t// Load %s lookup data\n", name))
		sb.WriteString(fmt.Sprintf("\tdata, _, shape, err := cplio.ReadParquetToTensor(filepath.Join(*dataDir, %q), %s, nil, conceptValues)\n", path, dimsLiteral))
		sb.WriteString("\tif err != nil {\n")
		sb.WriteString(fmt.Sprintf("\t\treturn fmt.Errorf(\"failed to load lookup '%s': %%w\", err)\n", name))
		sb.WriteString("\t}\n")

		// ReadParquetToTensor returns []float64. Convert to []any so that
		// NewChronoconeTensor creates *ChronoconeTensor[any], which is the
		// type expected by MetricComputation.SetTensor.
		sb.WriteString("\ttypedData := make([]any, len(data))\n")
		sb.WriteString("\tfor i, v := range data {\n")
		sb.WriteString("\t\ttypedData[i] = v\n")
		sb.WriteString("\t}\n")

		sb.WriteString(fmt.Sprintf("\ttensor := calculation.NewChronoconeTensor(typedData, %s, shape)\n", dimsLiteral))
		sb.WriteString("\tmc := calculation.NewMetricComputation(0)\n")
		sb.WriteString("\tif err := mc.SetTensor(tensor); err != nil {\n")
		sb.WriteString(fmt.Sprintf("\t\treturn fmt.Errorf(\"failed to set tensor for lookup '%s': %%w\", err)\n", name))
		sb.WriteString("\t}\n")
		sb.WriteString(fmt.Sprintf("\tif err := controller.RegisterMetricComputation(%q, mc); err != nil {\n", name))
		sb.WriteString(fmt.Sprintf("\t\treturn fmt.Errorf(\"failed to register lookup '%s': %%w\", err)\n", name))
		sb.WriteString("\t}\n")
	}

	sb.WriteString("\treturn nil\n")
	sb.WriteString("}\n")

	_, err := w.Write([]byte(sb.String()))
	return err
}

// generateLookupIntegration generates the lookup loading function code for
// inclusion in the synthesized program's RawSourceBeforeMain.
//
// It returns an empty string when no lookups are registered in the
// transformer's CILRegistry, allowing the caller to skip emitting the
// loadLookupData function entirely in that case.
//
// Parameters:
//   - input: The synthesis input containing the transformer and IO
//     specification with lookups.
//
// Returns:
//   - string: The generated loadLookupData function code, or empty string
//     if no lookups are registered.
//   - error: An error if lookup code generation fails (e.g., missing IO spec
//     entry for a registered lookup).
func generateLookupIntegration(input *SynthesisInput) (string, error) {
	if input == nil || input.Transformer == nil || input.IOSpec == nil {
		return "", nil
	}
	if len(input.Transformer.AllLookupMetadata()) == 0 {
		return "", nil
	}
	if len(input.IOSpec.Lookups) == 0 {
		return "", fmt.Errorf("lookup IO spec is required when lookups are registered")
	}
	var buf strings.Builder
	if err := generateLookupLoading(input.Transformer, input.IOSpec.Lookups, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// formatStringSliceLiteral formats a slice of strings as a Go slice literal.
// For example, []string{"a", "b"}.
func formatStringSliceLiteral(values []string) string {
	var sb strings.Builder
	sb.WriteString("[]string{")
	for i, v := range values {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%q", v))
	}
	sb.WriteString("}")
	return sb.String()
}
