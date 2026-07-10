package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	gen "github.com/chronocone/chronocone_planning_language/src/main/go/cli/expected_xls_generator"
	"github.com/chronocone/chronocone_planning_language/src/main/go/io"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
)

func main() {
	var exampleDir string
	var outputPath string
	flag.StringVar(&exampleDir, "example", "", "Path to example directory (required)")
	flag.StringVar(&outputPath, "output", "", "Path to output XLSX file (default: <example>/expected_output.xlsx)")
	flag.Parse()

	if exampleDir == "" {
		fmt.Fprintf(os.Stderr, "Error: -example flag is required\n")
		flag.Usage()
		os.Exit(2)
	}

	if outputPath == "" {
		outputPath = filepath.Join(exampleDir, "expected_output.xlsx")
	}

	// 1. Collect .cpl files from the example directory
	cplPattern := filepath.Join(exampleDir, "*.cpl")
	cplFiles, err := filepath.Glob(cplPattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding CPL files: %v\n", err)
		os.Exit(2)
	}

	if len(cplFiles) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no .cpl files found in %s\n", exampleDir)
		os.Exit(2)
	}
	sort.Strings(cplFiles)

	// Check for core.cpl in parent directory (shared infrastructure)
	parentDir := filepath.Dir(exampleDir)
	parentCore := filepath.Join(parentDir, "core.cpl")
	if _, err := os.Stat(parentCore); err == nil {
		cplFiles = append([]string{parentCore}, cplFiles...)
		fmt.Printf("Found core.cpl: %s\n", parentCore)
	} else {
		fmt.Printf("Warning: core.cpl not found in %s. Proceeding without shared infrastructure.\n", parentDir)
	}

	// Prepend standard definitions
	stdDefEntries, err := os.ReadDir("standard_definitions")
	if err == nil {
		var stdDefPaths []string
		for _, entry := range stdDefEntries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".cpl") {
				stdDefPaths = append(stdDefPaths, filepath.Join("standard_definitions", entry.Name()))
			}
		}
		sort.Strings(stdDefPaths)
		cplFiles = append(stdDefPaths, cplFiles...)
	}

	// 2. Parse CPL files with logging config
	loggingConfig := &parser.ParseLoggingConfig{
		LogLevel:   slog.LevelInfo,
		OutputFile: "", // logs to stdout
	}
	rootTable, err := parser.ParseFiles(cplFiles, loggingConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing CPL files: %v\n", err)
		os.Exit(2)
	}

	// 3. Analyze
	defs, err := analysis.Analyze(rootTable)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error analyzing CPL: %v\n", err)
		os.Exit(2)
	}

	// 4. Load IO spec
	ioSpecPath := filepath.Join(exampleDir, "iospec.json")
	iospec, err := io.LoadIOSpecification(ioSpecPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading IO spec: %v\n", err)
		os.Exit(2)
	}

	// 5. Load input data files
	inputData := make(map[string][]map[string]interface{})
	dataPattern := filepath.Join(exampleDir, "data_*.json")
	dataFiles, err := filepath.Glob(dataPattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding data files: %v\n", err)
		os.Exit(2)
	}
	sort.Strings(dataFiles)
	if len(dataFiles) != len(iospec.Connectors) {
		fmt.Fprintf(os.Stderr, "Error: found %d data_*.json files but expected %d (matching iospec connectors)\n", len(dataFiles), len(iospec.Connectors))
		os.Exit(2)
	}
	for i, df := range dataFiles {
		raw, err := os.ReadFile(df)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", df, err)
			os.Exit(2)
		}
		var rows []map[string]interface{}
		if err := json.Unmarshal(raw, &rows); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", df, err)
			os.Exit(2)
		}
		connectorName := iospec.Connectors[i].ConnectorName
		inputData[connectorName] = rows
	}

	// 5b. Load lookup data files
	lookupData := make(map[string][]map[string]interface{})
	for _, lookupSpec := range iospec.Lookups {
		lookupFile := filepath.Join(exampleDir, lookupSpec.LookupName+".json")
		raw, err := os.ReadFile(lookupFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading lookup file %s: %v\n", lookupFile, err)
			os.Exit(2)
		}
		var rows []map[string]interface{}
		if err := json.Unmarshal(raw, &rows); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing lookup file %s: %v\n", lookupFile, err)
			os.Exit(2)
		}
		lookupData[lookupSpec.LookupName] = rows
	}

	// 6. Generate XLSX
	f, inputColumns, err := gen.GenerateXLSX(defs, iospec, inputData, lookupData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating XLSX: %v\n", err)
		os.Exit(2)
	}

	// 7. Save XLSX
	if err := f.SaveAs(outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing XLSX: %v\n", err)
		os.Exit(2)
	}
	f.Close()

	// 8. Write expected_output_schema.json
	if len(iospec.Outputs) > 0 {
		output := iospec.Outputs[0]
		keyColumns := make([]string, 0, len(output.Dimensions)+len(output.Grain))
		keyColumns = append(keyColumns, output.Dimensions...)
		keyColumns = append(keyColumns, output.Grain...)
		schema := gen.SchemaMeta{
			XLSXFile:     filepath.Base(outputPath),
			KeyColumns:   keyColumns,
			InputColumns: inputColumns,
			MetricColumn: len(keyColumns) + len(inputColumns),
			ValueColumn:  len(keyColumns) + len(inputColumns) + 1,
		}
		schemaPath := filepath.Join(exampleDir, "expected_output_schema.json")
		schemaData, err := json.MarshalIndent(schema, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling schema: %v\n", err)
			os.Exit(2)
		}
		if err := os.WriteFile(schemaPath, schemaData, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing schema: %v\n", err)
			os.Exit(2)
		}
	}

	fmt.Println("XLSX generation complete:", outputPath)
}
