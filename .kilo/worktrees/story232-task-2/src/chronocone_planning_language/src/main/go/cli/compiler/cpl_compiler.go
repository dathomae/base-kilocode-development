// Package main provides the CPL compiler command-line interface.
// This compiler takes CPL files as input and produces representations
// of the input through parsing, analysis, and synthesis phases.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	cio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/chronocone/chronocone_planning_language/src/main/go/synthesis"
)

// levelMap converts string level names to slog.Level values.
var levelMap = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

// parseResult holds the results from the parse phase.
// It contains the root symbol table for analysis.
type parseResult struct {
	rootTable *symboltable.SymbolTable
}

// CPLCompiler orchestrates the CPL compilation pipeline from parsing through
// analysis, and optionally through code synthesis.
type CPLCompiler struct {
	inputPaths      []string
	outputPath      string
	ioSpecPath      string
	compileOnly     bool
	logger          *slog.Logger
	logLevel        slog.Level
	standardDefsDir string
}

// NewCPLCompiler creates a new CPLCompiler instance with the specified
// input path, output path, IO spec path, logger, and log level.
//
// Parameters:
//   - inputPaths: paths to the input CPL files or directories
//   - outputPath: path where the generated Go file will be written
//   - ioSpecPath: path to the IO specification JSON file
//   - logger: logger for compilation messages
//   - logLevel: minimum log level to output (for parse phase logging)
//
// Returns:
//   - *CPLCompiler: a new compiler instance configured with the provided parameters
func NewCPLCompiler(inputPaths []string, outputPath string, ioSpecPath string, logger *slog.Logger, logLevel slog.Level) *CPLCompiler {
	return &CPLCompiler{
		inputPaths:      inputPaths,
		outputPath:      outputPath,
		ioSpecPath:      ioSpecPath,
		logger:          logger,
		logLevel:        logLevel,
		standardDefsDir: "standard_definitions",
	}
}

// Compile orchestrates the CPL compilation pipeline through parse and analyze
// phases, followed by synthesis unless compile-only mode is enabled. Each phase
// feeds its output to the next phase. If any phase encounters an error,
// compilation stops and the error is returned.
//
// Returns:
//   - error: nil if compilation succeeds, or an error describing the failure
func (c *CPLCompiler) Compile() error {
	c.logger.Info("compile_entry", "input", c.inputPaths, "output", c.outputPath, "compile_only", c.compileOnly)

	// Phase 1: Parse
	parseResult, err := c.parsePhase()
	if err != nil {
		return fmt.Errorf("parse phase failed: %w", err)
	}

	// Phase 2: Analyze
	defs, err := c.analyzePhase(parseResult)
	if err != nil {
		return fmt.Errorf("analysis phase failed: %w", err)
	}

	if c.compileOnly {
		c.logger.Info("compile_exit", "status", "success", "mode", "compile_only")
		return nil
	}

	// Phase 3: Synthesize
	if err := c.synthesizePhase(defs); err != nil {
		return fmt.Errorf("synthesis phase failed: %w", err)
	}

	c.logger.Info("compile_exit", "status", "success")
	return nil
}

// discoverStandardDefinitions finds all .cpl files in the standard definitions
// directory. Returns an empty slice if the directory does not exist or contains
// no .cpl files.
func (c *CPLCompiler) discoverStandardDefinitions() ([]string, error) {
	if c.standardDefsDir == "" {
		return nil, nil
	}

	info, err := os.Stat(c.standardDefsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to stat standard definitions directory: %w", err)
	}

	if !info.IsDir() {
		return nil, nil
	}

	entries, err := os.ReadDir(c.standardDefsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read standard definitions directory: %w", err)
	}

	var paths []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".cpl") {
			paths = append(paths, filepath.Join(c.standardDefsDir, entry.Name()))
		}
	}

	return paths, nil
}

// parsePhase performs the parsing phase of compilation, converting the input
// CPL files into an AST representation and building the symbol table.
//
// Standard definition files are automatically discovered and prepended to the
// user-provided input paths so that standard types and concepts are available
// for reference in user code.
//
// Returns:
//   - *parseResult: contains both the parsed AST document and the root symbol table
//   - error: nil if parsing succeeds, or an error if parsing fails
func (c *CPLCompiler) parsePhase() (*parseResult, error) {
	c.logger.Info("parse_phase_entry", "input_paths", c.inputPaths)

	// Discover and include standard definition files
	stdDefPaths, err := c.discoverStandardDefinitions()
	if err != nil {
		return nil, fmt.Errorf("failed to discover standard definitions: %w", err)
	}

	allPaths := append(stdDefPaths, c.inputPaths...)

	// Use ParseFiles from the parser package to handle multiple input files
	config := &parser.ParseLoggingConfig{
		LogLevel:   c.logLevel,
		OutputFile: "", // Empty means stdout
	}

	rootTable, err := parser.ParseFiles(allPaths, config)
	if err != nil {
		return nil, fmt.Errorf("parse errors: %w", err)
	}

	c.logger.Info("parse_phase_exit", "status", "success")
	return &parseResult{
		rootTable: rootTable,
	}, nil
}

// analyzePhase performs semantic analysis on the parsed AST, validating
// concepts and extracting analysis definitions. When an IO specification path
// is configured, the IO spec is loaded and passed to analysis so that dead-code
// elimination can prune unreachable metrics before synthesis.
//
// Parameters:
//   - result: the parse result containing both document and symbol table
//
// Returns:
//   - *analysis.AnalysisDefinitions: the analysis results
//   - error: nil if analysis succeeds, or an error if analysis fails
func (c *CPLCompiler) analyzePhase(result *parseResult) (*analysis.AnalysisDefinitions, error) {
	c.logger.Info("analyze_phase_entry")

	if result.rootTable == nil {
		return nil, fmt.Errorf("root symbol table is nil")
	}

	if c.ioSpecPath != "" {
		ioSpec, err := cio.LoadIOSpecification(c.ioSpecPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load IO specification: %w", err)
		}
		return analysis.AnalyzeWithIO(result.rootTable, ioSpec)
	}

	return analysis.Analyze(result.rootTable)
}

// synthesizePhase generates the output Go code from the analysis results.
// It loads the IO specification and uses Synthesize to produce a full
// computation lifecycle program.
//
// Parameters:
//   - defs: the analysis definitions to synthesize into Go code
//
// Returns:
//   - error: nil if synthesis succeeds, or an error if synthesis fails
func (c *CPLCompiler) synthesizePhase(defs *analysis.AnalysisDefinitions) error {
	c.logger.Info("synthesize_phase_entry", "output_path", c.outputPath, "io_spec_path", c.ioSpecPath)

	ioSpec, err := cio.LoadIOSpecification(c.ioSpecPath)
	if err != nil {
		return fmt.Errorf("failed to load IO specification: %w", err)
	}

	if err := synthesis.Synthesize(defs, ioSpec, c.outputPath); err != nil {
		return fmt.Errorf("synthesis failed: %w", err)
	}

	c.logger.Info("synthesize_phase_exit", "status", "success")
	return nil
}

// validateOutputPath validates that the specified output path is valid and writable.
// It performs the following checks:
//   - Path is not empty
//   - Parent directory exists or can be created
//   - Path is writable (by testing file creation)
//
// Parameters:
//   - path: the file path to validate for output
//
// Returns:
//   - error: nil if valid, or an error describing why the path is invalid
func validateOutputPath(path string) error {
	if path == "" {
		return fmt.Errorf("output path cannot be empty")
	}

	// Get the parent directory of the output path
	dir := filepath.Dir(path)

	// Check if parent directory exists, create if it doesn't
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory %q: %w", dir, err)
		}
	}

	// Check if the directory is writable by attempting to create a temp file
	tmpFile, err := os.CreateTemp(dir, ".write_test_*")
	if err != nil {
		return fmt.Errorf("output path %q is not writable: %w", path, err)
	}
	// Clean up the temp file
	tmpFile.Close()
	os.Remove(tmpFile.Name())

	return nil
}

func main() {
	// Define command-line flags
	logLevelFlag := flag.String("loglevel", "info", "Set logging level: debug, info, warn, error")
	logFileFlag := flag.String("logfile", "", "Write logs to file (default: stdout)")
	outputFlag := flag.String("o", "", "Output file path for generated Go CLI (required unless --compile_only is set)")
	flag.StringVar(outputFlag, "output", "", "Output file path for generated Go CLI (required unless --compile_only is set)")
	ioSpecFlag := flag.String("io", "", "IO specification JSON file path (required unless --compile_only is set)")
	compileOnlyFlag := flag.Bool("compile_only", false, "Run parse and analyze phases only; skip code generation")

	// Set custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `CPL Compiler - Chronocone Planning Language

Compiles CPL source files into Go CLI executables.

COMPILATION PIPELINE:
	 1. Parse Phase: Reads CPL source and builds symbol table
	 2. Analyze Phase: Validates semantics and builds AnalysisDefinitions
	 3. Synthesize Phase: Generates Go CLI executable unless --compile_only is set

OPTIONS:
`)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
USAGE:
	 %s -io <io_spec.json> -o <output.go> <input.cpl>...
	 %s --compile_only <input.cpl>...

EXAMPLES:
	 # Compile with custom output path
	 %s -io ./io_spec.json -o ./my_output/generated.go input.cpl

	 # Compile with long-form output flag
	 %s -io ./io_spec.json --output ./my_output/generated.go input.cpl

	 # Compile with debug logging
	 %s -loglevel=debug -io ./io_spec.json -o ./output.go input.cpl

	 # Compile multiple files
	 %s -io ./io_spec.json -o ./output.go file1.cpl file2.cpl

	 # Compile with IO specification for connector-aware synthesis
	 %s -io ./io_spec.json -o ./output.go input.cpl

	 # Compile only: parse and analyze without code generation or IO specification
	 %s --compile_only input.cpl
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
	}

	// Parse flags
	flag.Parse()

	// Get remaining arguments (input files)
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	inputFiles := args

	// Extract and validate output path from flag when synthesis will run.
	outputPath := *outputFlag
	if !*compileOnlyFlag {
		if outputPath == "" {
			fmt.Fprintf(os.Stderr, "Error: no output path specified. Use -o or --output to specify the output file, or use --compile_only to skip code generation.\n\n")
			flag.Usage()
			os.Exit(1)
		}
		if err := validateOutputPath(outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid output path: %v\n", err)
			os.Exit(1)
		}
	}

	// Validate IO spec path when synthesis will run.
	ioSpecPath := *ioSpecFlag
	if !*compileOnlyFlag && ioSpecPath == "" {
		fmt.Fprintf(os.Stderr, "Error: no IO specification path specified. Use -io to specify the IO spec file, or use --compile_only to skip code generation.\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Parse and validate log level
	levelStr := strings.ToLower(*logLevelFlag)
	_, ok := levelMap[levelStr]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: invalid log level '%s'\n", *logLevelFlag)
		fmt.Fprintf(os.Stderr, "Valid log levels: debug, info, warn, error\n")
		os.Exit(1)
	}

	// Get log level from map
	logLevel := levelMap[levelStr]

	// Note: logFileFlag is parsed for CLI compatibility
	// It will be used when CPLCompiler.Compile() is fully implemented
	_ = *logFileFlag

	// Create logger
	logger := slog.Default()

	// Create CPLCompiler instance
	compiler := NewCPLCompiler(inputFiles, outputPath, ioSpecPath, logger, logLevel)
	compiler.compileOnly = *compileOnlyFlag

	// Run the compilation pipeline
	if err := compiler.Compile(); err != nil {
		fmt.Fprintf(os.Stderr, "Compilation error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s successfully processed\n", strings.Join(inputFiles, ", "))
}
