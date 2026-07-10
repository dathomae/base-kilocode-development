// Package parser provides CPL parsing functionality.
// This package wraps the ANTLR-generated lexer and parser to provide
// a clean API for parsing Chronocone Planning Language documents.
//
// Error Handling Strategy:
// The parser collects three types of errors during parsing:
//  1. Lexer syntax errors - tokenization failures
//  2. Parser syntax errors - grammar rule violations
//  3. Semantic errors - validation errors from the CPLListener
//
// Error Combination Approach:
// After walking the parse tree, errors are combined in the following order:
//  1. Retrieve lexer errors from lexerErrorListener.errors
//  2. Retrieve parser errors from parserErrorListener.errors
//  3. Retrieve semantic errors from baseListener.GetErrors()
//
// All errors are concatenated with "; " separator and returned as a ParseError.
//
// Example combined error format:
//
//	"parse error: syntax error at line 1:5 missing IDENTIFIER; semantic error at line 2:10 in unit_declaration: unit 'MyUnit' missing type"
package parser

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	antlrruntime "github.com/antlr4-go/antlr/v4"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser/antlr"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// ParseError represents an error that occurred during parsing.
type ParseError struct {
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error: %s", e.Message)
}

// ParseLoggingConfig configures logging behavior for the parser.
// It controls the log level and output destination for parse operations.
type ParseLoggingConfig struct {
	LogLevel   slog.Level
	OutputFile string // If empty, logs to stdout; otherwise to specified file
}

// ParseWithLogLevel parses a CPL document with configurable logging.
// It takes an input string and a logging configuration, and returns
// the root symbol table if parsing succeeds, or an error if parsing fails.
// The config parameter must not be nil.
// Logs are written at the configured level, with parse entry/exit logged
// at the specified level and rule entry/exit logged at Debug level.
// Parsing populates the symbol table for subsequent analysis phases.
func ParseWithLogLevel(input string, config *ParseLoggingConfig) (*symboltable.SymbolTable, error) {
	return parseStringWithFile(input, "", config)
}

// ParseFiles parses multiple CPL files as a single compilation unit.
// It discovers all .cpl files if a directory path is provided, or parses
// the specified file paths directly. The resulting symbol table contains
// definitions from all files merged into a single global scope.
//
// Parameters:
//   - filePaths: list of paths to .cpl files or directories
//   - config: logging configuration for parse operations
//
// Returns:
//   - *symboltable.SymbolTable: the merged root symbol table
//   - error: nil if parsing succeeds, or a combined error of all parse failures
func ParseFiles(filePaths []string, config *ParseLoggingConfig) (*symboltable.SymbolTable, error) {
	if config == nil {
		return nil, fmt.Errorf("logging config cannot be nil")
	}

	// Expand directories to individual .cpl files
	var expandedPaths []string
	for _, path := range filePaths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("failed to access path %q: %w", path, err)
		}
		if info.IsDir() {
			files, err := discoverCPLFiles(path)
			if err != nil {
				return nil, fmt.Errorf("failed to discover files in directory %q: %w", path, err)
			}
			expandedPaths = append(expandedPaths, files...)
		} else {
			expandedPaths = append(expandedPaths, path)
		}
	}

	if len(expandedPaths) == 0 {
		return nil, fmt.Errorf("no .cpl files found to parse")
	}

	// Parse each file individually and collect results
	var allErrors []string
	var rootTables []*symboltable.SymbolTable

	for _, filePath := range expandedPaths {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %q: %w", filePath, err)
		}

		rootTable, err := parseStringWithFile(string(content), filePath, config)
		if err != nil {
			allErrors = append(allErrors, err.Error())
			continue
		}
		rootTables = append(rootTables, rootTable)
	}

	// Merge all symbol tables into a single compilation unit
	mergedTable, mergeErr := mergeSymbolTables(rootTables)
	if mergeErr != nil {
		return nil, mergeErr
	}

	// If there were parse errors, return them along with the partial result
	if len(allErrors) > 0 {
		return mergedTable, &ParseError{Message: strings.Join(allErrors, "; ")}
	}

	return mergedTable, nil
}

// ParseDirectory discovers all .cpl files in the given directory (non-recursive)
// and parses them as a single compilation unit.
//
// Parameters:
//   - dir: path to the directory containing .cpl files
//   - config: logging configuration for parse operations
//
// Returns:
//   - *symboltable.SymbolTable: the merged root symbol table
//   - error: nil if parsing succeeds, or a combined error of all parse failures
func ParseDirectory(dir string, config *ParseLoggingConfig) (*symboltable.SymbolTable, error) {
	return ParseFiles([]string{dir}, config)
}

// discoverCPLFiles finds all .cpl files in the given directory (non-recursive).
func discoverCPLFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".cpl") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	return files, nil
}

// mergeSymbolTables merges multiple root symbol tables into a single global scope.
// Definitions and children from each table are copied into the merged table.
// Duplicate definitions across files are preserved and can be detected during analysis.
func mergeSymbolTables(tables []*symboltable.SymbolTable) (*symboltable.SymbolTable, error) {
	if len(tables) == 0 {
		return symboltable.NewGlobalSymbolTable(), nil
	}
	if len(tables) == 1 {
		return tables[0], nil
	}

	merged := symboltable.NewGlobalSymbolTable()

	for _, table := range tables {
		// Copy definitions from the root table
		for _, def := range table.Definitions {
			merged.Definitions = append(merged.Definitions, def)
		}
		// Copy children scopes
		for _, child := range table.Children {
			merged.Children = append(merged.Children, child)
			child.Parent = merged
		}
	}

	return merged, nil
}

// parseStringWithFile parses a CPL string with an associated source file name.
// It is the file-aware equivalent of ParseWithLogLevel.
func parseStringWithFile(input string, fileName string, config *ParseLoggingConfig) (*symboltable.SymbolTable, error) {
	if config == nil {
		return nil, fmt.Errorf("logging config cannot be nil")
	}

	startTime := time.Now()

	var handler slog.Handler
	var logFile *os.File

	if config.OutputFile == "" {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: config.LogLevel})
	} else {
		parentDir := filepath.Dir(config.OutputFile)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		var err error
		logFile, err = os.OpenFile(config.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		defer logFile.Close()
		handler = slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: config.LogLevel})
	}

	logger := slog.New(handler)

	logger.Log(nil, config.LogLevel, "parse_entry", "input_length", len(input), "file", fileName)

	inputStream := antlrruntime.NewInputStream(input)
	lexer := antlr.NewChronoconePlanningLanguageLexer(inputStream)

	lexerErrorListener := &syntaxErrorListener{fileName: fileName}
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(lexerErrorListener)

	tokenStream := antlrruntime.NewCommonTokenStream(lexer, antlrruntime.TokenDefaultChannel)
	parser := antlr.NewChronoconePlanningLanguageParser(tokenStream)

	parserErrorListener := &syntaxErrorListener{fileName: fileName}
	parser.RemoveErrorListeners()
	parser.AddErrorListener(parserErrorListener)

	baseListener := NewCPLListenerWithFile(logger, lexer, parser, fileName)

	antlrruntime.ParseTreeWalkerDefault.Walk(baseListener, parser.Document())

	var allErrors []string
	allErrors = append(allErrors, lexerErrorListener.errors...)
	allErrors = append(allErrors, parserErrorListener.errors...)

	semanticErrors := baseListener.GetErrors()
	for _, err := range semanticErrors {
		allErrors = append(allErrors, err.Error())
	}

	var parseErr error
	if len(allErrors) > 0 {
		parseErr = &ParseError{Message: strings.Join(allErrors, "; ")}
		logger.Error("parse_completed_with_errors", "error_count", len(allErrors), "file", fileName)
	}

	rootTable, err := baseListener.GetSymbolStack().Root()
	if err != nil {
		return nil, fmt.Errorf("failed to get root symbol table: %w", err)
	}

	duration := time.Since(startTime)

	var errStr string
	if parseErr != nil {
		errStr = parseErr.Error()
	} else {
		errStr = "null"
	}
	logger.Log(nil, config.LogLevel, "parse_exit", "duration_ms", duration.Milliseconds(), "error", errStr, "file", fileName)

	return rootTable, parseErr
}

// syntaxErrorListener captures syntax errors from the lexer and parser.
type syntaxErrorListener struct {
	*antlrruntime.DefaultErrorListener
	errors   []string
	fileName string
}

func (s *syntaxErrorListener) SyntaxError(recognizer antlrruntime.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlrruntime.RecognitionException) {
	if s.fileName != "" {
		s.errors = append(s.errors, fmt.Sprintf("%s:%d:%d %s", s.fileName, line, column, msg))
	} else {
		s.errors = append(s.errors, fmt.Sprintf("line %d:%d %s", line, column, msg))
	}
}

// Parse takes an input string and parses it as a CPL document.
// It returns the root symbol table if parsing succeeds, or an error if parsing fails.
// Parsing populates the symbol table for subsequent analysis phases.
// This function uses Info level logging to stdout by default.
func Parse(input string) (*symboltable.SymbolTable, error) {
	config := &ParseLoggingConfig{
		LogLevel:   slog.LevelInfo,
		OutputFile: "", // Empty means stdout
	}
	return ParseWithLogLevel(input, config)
}
