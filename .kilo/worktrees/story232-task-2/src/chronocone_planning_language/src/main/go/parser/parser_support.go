// Package parser provides CPL parsing functionality.
//
// This file contains supporting definitions for the parser package including
// the SemanticError type, CPLListener struct, and utility methods used across
// multiple listener files.
package parser

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	antlrruntime "github.com/antlr4-go/antlr/v4"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser/antlr"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parsing"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// SemanticError represents a semantic error encountered during parsing.
// It contains detailed information about the error location and context.
type SemanticError struct {
	Message string // Error message describing what went wrong
	Rule    string // Name of the grammar rule where the error occurred
	File    string // Source file path where the error occurred
	Line    int    // Line number in the source file (1-based)
	Column  int    // Column number in the source file (0-based)
}

// Error returns the error message, implementing the error interface.
func (e *SemanticError) Error() string {
	if e.File != "" {
		return fmt.Sprintf("semantic error at %s:%d:%d in %s: %s", e.File, e.Line, e.Column, e.Rule, e.Message)
	}
	return fmt.Sprintf("semantic error at line %d:%d in %s: %s", e.Line, e.Column, e.Rule, e.Message)
}

// CPLListener implements the ANTLR listener interface to build the symbol table.
// It processes parse tree events and manages scope via the symbol table stack.
// This is the main listener corresponding to ChronoconePlanningLanguage.g4.
type CPLListener struct {
	*antlr.BaseChronoconePlanningLanguageListener
	logger      *slog.Logger
	errors      []error                       // Collection of semantic errors encountered during parsing
	symbolStack *symboltable.SymbolTableStack // Stack of symbol tables for scope management
	depth       int                           // Current parse tree depth for indentation
	lexer       *antlr.ChronoconePlanningLanguageLexer
	parser      *antlr.ChronoconePlanningLanguageParser
	fileName    string // Source file path for error reporting and source tracking

	// conditionalStates tracks state during conditional expression parsing.
	// Each nested conditional branch pushes a new state; the top of the slice
	// is the current branch being parsed. Empty when not inside a conditional.
	conditionalStates []*ConditionalParseState

	// currentTimeShift holds the time shift clause data when parsing a backward() or forward()
	// clause on an input declaration. It is set by ExitBackwardClause/ExitForwardClause and
	// consumed by ExitInputDeclaration. nil means no time shift was specified.
	currentTimeShift *parsing.ParserTimeShift

	// timeShiftExpressionScope holds a pointer to the TempScope pushed when entering
	// a shiftExpression (inside backward() or forward()). The scope is popped in
	// ExitShiftExpression and its Definitions are converted to PostfixElements.
	timeShiftExpressionScope *symboltable.SymbolTable

	// hasTimeShift tracks whether a time shift clause (backward or forward) has already
	// been processed on the current input declaration. Used to detect and reject inputs
	// that specify both backward() and forward() clauses, which are mutually exclusive.
	hasTimeShift bool
}

// NewCPLListener creates a new CPLListener.
// The logger parameter is optional and can be nil if logging is not needed.
// Initializes the symbol table stack with a global scope table.
func NewCPLListener(logger *slog.Logger,
	lexer *antlr.ChronoconePlanningLanguageLexer,
	parser *antlr.ChronoconePlanningLanguageParser) *CPLListener {
	return NewCPLListenerWithFile(logger, lexer, parser, "")
}

// NewCPLListenerWithFile creates a new CPLListener with an associated source file name.
// The fileName parameter is used for source location tracking in multi-file compilations.
// The logger parameter is optional and can be nil if logging is not needed.
// Initializes the symbol table stack with a global scope table.
func NewCPLListenerWithFile(logger *slog.Logger,
	lexer *antlr.ChronoconePlanningLanguageLexer,
	parser *antlr.ChronoconePlanningLanguageParser,
	fileName string) *CPLListener {
	// Create symbol table stack and push global scope
	stack := symboltable.NewSymbolTableStack()
	globalTable := symboltable.NewGlobalSymbolTable()
	stack.Push(globalTable)

	return &CPLListener{
		logger:      logger,
		errors:      make([]error, 0),
		symbolStack: stack,
		lexer:       lexer,
		parser:      parser,
		fileName:    fileName,
	}
}

// logEnter logs entry into a rule if logger is configured.
// Includes indentation based on current parse tree depth.
func (l *CPLListener) logEnter(rule string, ctx *antlrruntime.BaseParserRuleContext) {
	l.depth++
	if l.logger == nil || l.lexer == nil {
		return
	}

	// Get start token and validate it
	startToken := ctx.GetStart()
	if startToken == nil {
		return
	}

	tokenType := startToken.GetTokenType()
	// Check for valid token type (EOF = -1 is invalid for indexing)
	if tokenType < 0 || tokenType >= len(l.lexer.SymbolicNames) {
		return
	}

	indent := strings.Repeat("  ", (l.depth-1)*2)
	line := strconv.Itoa(startToken.GetLine())
	column := strconv.Itoa(startToken.GetColumn())
	tokenName := l.lexer.SymbolicNames[tokenType]
	tokenText := startToken.GetText()

	l.logger.Debug(indent + line + ":" + column + " enter " + rule + "(" + tokenName + " " + tokenText + ")")
	l.logger.Debug(indent + "   start->" + "(" + tokenName + " " + tokenText + ")")
	l.logger.Debug(" ")
}

// logExit logs exit from a rule if logger is configured.
// Includes indentation based on current parse tree depth.
func (l *CPLListener) logExit(rule string, ctx *antlrruntime.BaseParserRuleContext) {
	l.depth--
	if l.logger == nil || l.lexer == nil {
		return
	}

	// Get start token and validate it
	startToken := ctx.GetStart()
	if startToken == nil {
		return
	}

	tokenType := startToken.GetTokenType()
	// Check for valid token type (EOF = -1 is invalid for indexing)
	if tokenType < 0 || tokenType >= len(l.lexer.SymbolicNames) {
		return
	}

	// Guard against negative depth
	if l.depth < 0 {
		l.depth = 0
	}
	indent := strings.Repeat("  ", l.depth*2)
	line := strconv.Itoa(startToken.GetLine())
	column := strconv.Itoa(startToken.GetColumn())
	tokenName := l.lexer.SymbolicNames[tokenType]
	tokenText := startToken.GetText()

	l.logger.Debug(indent + line + ":" + column + " exit" + rule + "(" + tokenName + " " + tokenText + ")")
	l.logger.Debug(indent + "   start->" + "(" + tokenName + " " + tokenText + ")")
	l.logger.Debug(" ")
}

// AddError adds a semantic error to the listener's error collection.
// This method is called during parsing when semantic errors are encountered.
func (l *CPLListener) AddError(err error) {
	if err == nil {
		return
	}
	l.errors = append(l.errors, err)
	if l.logger != nil {
		l.logger.Error("semantic_error", "error", err)
	}
}

// GetErrors returns all semantic errors collected during parsing.
// Returns an empty slice if no errors were encountered.
func (l *CPLListener) GetErrors() []error {
	result := make([]error, len(l.errors))
	copy(result, l.errors)
	return result
}

// HasErrors returns true if any semantic errors were collected during parsing.
func (l *CPLListener) HasErrors() bool {
	return len(l.errors) > 0
}

// GetSymbolStack returns the symbol table stack used during parsing.
// This allows access to the symbol table tree after parsing is complete.
func (l *CPLListener) GetSymbolStack() *symboltable.SymbolTableStack {
	return l.symbolStack
}

// shouldPushExpressionDefs returns true if the current scope is one where
// expression definitions (operands and operators) should be accumulated
// in postfix order. Returns true for MetricScope, AggregatorScope, TempScope,
// and all conditional body scopes (IfBodyScope, ElseIfBodyScope, ElseBodyScope,
// WhenBodyScope, DefaultBodyScope).
func (l *CPLListener) shouldPushExpressionDefs() bool {
	currentScope, err := l.symbolStack.Peek()
	if err != nil {
		return false
	}
	switch currentScope.ScopeType {
	case symboltable.MetricScope,
		symboltable.AggregatorScope,
		symboltable.SpreaderScope,
		symboltable.MeasureScope,
		symboltable.TempScope,
		symboltable.IfBodyScope,
		symboltable.ElseIfBodyScope,
		symboltable.ElseBodyScope,
		symboltable.WhenBodyScope,
		symboltable.DefaultBodyScope,
		symboltable.PatternScope,
		symboltable.CalculationScope:
		return true
	default:
		return false
	}
}

// pushOperatorDef pushes a ParserOperator definition onto the current scope.
func (l *CPLListener) pushOperatorDef(token antlrruntime.Token, opType OperatorType,
	defType symboltable.DefinitionType, loc symboltable.SourceLocation) {
	currentScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error pushing operator: %w", err))
		return
	}
	err = currentScope.DefineWithData(nil, defType, loc, ParserOperator{Operator: opType})
	if err != nil {
		l.AddError(fmt.Errorf("error defining operator: %w", err))
	}
}

// makeSourceLocation creates a SourceLocation from a token, including the source file name.
func (l *CPLListener) makeSourceLocation(token antlrruntime.Token) symboltable.SourceLocation {
	return symboltable.SourceLocation{
		File:   l.fileName,
		Line:   token.GetLine(),
		Column: token.GetColumn(),
	}
}
