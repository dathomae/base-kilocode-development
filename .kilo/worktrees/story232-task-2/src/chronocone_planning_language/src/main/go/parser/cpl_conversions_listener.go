// Package parser provides CPL parsing functionality.
//
// This file contains listener callbacks related to conversion clauses in both
// unit declarations (CPLUnits.g4) and measure declarations (CPLMeasures.g4).
//
// Grammar rules covered:
//   - conversionClause, patternClause, calculationClause, patternBoundVariable
//   - measureConversionClause, measureCalculationClause
//
// Symbol Table Driven Parsing:
// The parser builds a symbol table stack during parsing. Expression identifiers
// and operators are pushed to the current scope by existing listener methods
// (ExitIdentifier, ExitArithmeticFactor, ExitArithmeticTerm, ExitArithmeticExpr)
// when the current scope is in shouldPushExpressionDefs(). PatternScope and
// CalculationScope were added to shouldPushExpressionDefs() so that expressions
// inside pattern and calculation clauses accumulate definitions in postfix order.
//
// Detailed extraction and validation of conversion definitions is deferred to
// the analysis phase (subsequent tasks).
package parser

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser/antlr"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// ============================================================================
// Unit Conversion Clause Listeners
// ============================================================================

// EnterConversionClause is called when entering a conversion clause within a
// unit declaration. It pushes a new ConversionScope onto the symbol stack.
func (l *CPLListener) EnterConversionClause(ctx *antlr.ConversionClauseContext) {
	l.logEnter("ConversionClause", &ctx.BaseParserRuleContext)

	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering conversion clause: %w", err))
		return
	}

	l.symbolStack.Push(current.AddChild(symboltable.ConversionScope))
}

// ExitConversionClause is called when exiting a conversion clause within a
// unit declaration. It pops the ConversionScope from the symbol stack.
func (l *CPLListener) ExitConversionClause(ctx *antlr.ConversionClauseContext) {
	l.logExit("ConversionClause", &ctx.BaseParserRuleContext)

	_, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting conversion clause: %w", err))
	}
}

// EnterPatternClause is called when entering a pattern clause (inside either a
// unit conversion pair or a measure conversion pair). It pushes a new
// PatternScope onto the symbol stack.
func (l *CPLListener) EnterPatternClause(ctx *antlr.PatternClauseContext) {
	l.logEnter("PatternClause", &ctx.BaseParserRuleContext)

	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering pattern clause: %w", err))
		return
	}

	l.symbolStack.Push(current.AddChild(symboltable.PatternScope))
}

// ExitPatternClause is called when exiting a pattern clause. It pops the
// PatternScope from the symbol stack.
func (l *CPLListener) ExitPatternClause(ctx *antlr.PatternClauseContext) {
	l.logExit("PatternClause", &ctx.BaseParserRuleContext)

	_, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting pattern clause: %w", err))
	}
}

// EnterCalculationClause is called when entering a calculation clause within a
// unit conversion pair. It pushes a new CalculationScope onto the symbol stack.
func (l *CPLListener) EnterCalculationClause(ctx *antlr.CalculationClauseContext) {
	l.logEnter("CalculationClause", &ctx.BaseParserRuleContext)

	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering calculation clause: %w", err))
		return
	}

	l.symbolStack.Push(current.AddChild(symboltable.CalculationScope))
}

// ExitCalculationClause is called when exiting a calculation clause within a
// unit conversion pair. It pops the CalculationScope from the symbol stack.
func (l *CPLListener) ExitCalculationClause(ctx *antlr.CalculationClauseContext) {
	l.logExit("CalculationClause", &ctx.BaseParserRuleContext)

	_, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting calculation clause: %w", err))
	}
}

// ============================================================================
// Measure Conversion Clause Listeners
// ============================================================================

// EnterMeasureConversionClause is called when entering a conversion clause
// within a measure declaration. It pushes a new MeasureConversionScope onto
// the symbol stack.
func (l *CPLListener) EnterMeasureConversionClause(ctx *antlr.MeasureConversionClauseContext) {
	l.logEnter("MeasureConversionClause", &ctx.BaseParserRuleContext)

	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering measure conversion clause: %w", err))
		return
	}

	l.symbolStack.Push(current.AddChild(symboltable.MeasureConversionScope))
}

// ExitMeasureConversionClause is called when exiting a conversion clause within
// a measure declaration. It pops the MeasureConversionScope from the symbol
// stack.
func (l *CPLListener) ExitMeasureConversionClause(ctx *antlr.MeasureConversionClauseContext) {
	l.logExit("MeasureConversionClause", &ctx.BaseParserRuleContext)

	_, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting measure conversion clause: %w", err))
	}
}

// EnterMeasureCalculationClause is called when entering a calculation clause
// within a measure conversion pair. It pushes a new CalculationScope onto the
// symbol stack.
func (l *CPLListener) EnterMeasureCalculationClause(ctx *antlr.MeasureCalculationClauseContext) {
	l.logEnter("MeasureCalculationClause", &ctx.BaseParserRuleContext)

	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering measure calculation clause: %w", err))
		return
	}

	l.symbolStack.Push(current.AddChild(symboltable.CalculationScope))
}

// ExitMeasureCalculationClause is called when exiting a calculation clause
// within a measure conversion pair. It pops the CalculationScope from the
// symbol stack.
func (l *CPLListener) ExitMeasureCalculationClause(ctx *antlr.MeasureCalculationClauseContext) {
	l.logExit("MeasureCalculationClause", &ctx.BaseParserRuleContext)

	_, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting measure calculation clause: %w", err))
	}
}

// ============================================================================
// Pattern Bound Variable Listener
// ============================================================================

// ExitPatternBoundVariable is called when exiting a pattern-bound variable
// (LPAREN identifier RPAREN). It captures the bound variable name and adds it
// to the current PatternScope's definitions as an Identifier.
func (l *CPLListener) ExitPatternBoundVariable(ctx *antlr.PatternBoundVariableContext) {
	l.logExit("PatternBoundVariable", &ctx.BaseParserRuleContext)

	identifierCtx := ctx.Identifier()
	if identifierCtx == nil || identifierCtx.GetStart() == nil {
		l.AddError(fmt.Errorf("pattern bound variable missing identifier"))
		return
	}

	varName := identifierCtx.GetStart().GetText()
	currentScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error in ExitPatternBoundVariable: %w", err))
		return
	}

	err = currentScope.DefineWithData(
		&varName,
		symboltable.Identifier,
		l.makeSourceLocation(identifierCtx.GetStart()),
		nil,
	)
	if err != nil {
		l.AddError(fmt.Errorf("error defining pattern bound variable: %w", err))
	}
}
