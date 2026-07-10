// Package parser provides CPL parsing functionality.
//
// This file contains listener callbacks related to the CPLMetrics.g4 grammar file,
// which defines the metrics section and metric declarations.
//
// Grammar rules covered:
//   - metricsSection, metricDeclaration, inputDeclaration
//
// Symbol Table Driven Parsing:
// The parser builds a symbol table stack during parsing. Metric definitions are
// stored in MetricScope symbol tables. Detailed extraction and validation of
// metric definitions is deferred to the analysis phase.
package parser

import (
	"fmt"
	"strconv"

	antlrruntime "github.com/antlr4-go/antlr/v4"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser/antlr"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parsing"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// EnterMetricsSection is called when entering the metrics section.
// It pushes a new MetricsScope symbol table onto the stack.
//
// Error Handling:
// If the symbol stack operation fails, an error is added via AddError.
func (l *CPLListener) EnterMetricsSection(ctx *antlr.MetricsSectionContext) {
	l.logEnter("MetricsSection", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for the Metrics scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering Metrics section: %w", err))
		return
	}

	metricsTable := current.AddChild(symboltable.MetricsScope)
	l.symbolStack.Push(metricsTable)
}

// ExitMetricsSection is called when exiting the metrics section.
// It pops the Metrics scope symbol table from the stack.
//
// Error Handling:
// If the symbol stack operation fails, an error is added via AddError.
func (l *CPLListener) ExitMetricsSection(ctx *antlr.MetricsSectionContext) {
	_, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting Metrics section: %w", err))
	}
	l.logExit("MetricsSection", &ctx.BaseParserRuleContext)
}

// EnterMetricDeclaration is called when entering a metric declaration.
// It pushes a new MetricScope onto the symbol stack.
//
// Example metric declaration:
//
//	my_metric
//		inputs
//			metric input1
//			metric input2
//		calculate
//			0
//
// This corresponds to the 'metricDeclaration' rule in CPLMetrics.g4.
func (l *CPLListener) EnterMetricDeclaration(ctx *antlr.MetricDeclarationContext) {
	l.logEnter("MetricDeclaration", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for the Metric scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering metric declaration: %w", err))
		return
	}

	// Create a child scope for this metric definition and push it on the stack
	l.symbolStack.Push(current.AddChild(symboltable.MetricScope))
}

// ExitMetricDeclaration is called when exiting a metric declaration.
// It extracts the metric definition data, creates a ParserMetricDefinition,
// and stores it in the MetricScope's Data field before popping the scope.
//
// Symbol Table Driven Parsing:
// The metric name and inputs are collected from the symbol table definitions.
// The ParserMetricDefinition is stored in the MetricScope.Data field for later
// retrieval by the analysis phase.
func (l *CPLListener) ExitMetricDeclaration(ctx *antlr.MetricDeclarationContext) {
	l.logExit("MetricDeclaration", &ctx.BaseParserRuleContext)

	// Get the current MetricScope before popping
	metricScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting metric declaration: %w", err))
		return
	}

	// Get the metric name from the identifier
	identifier := ctx.Identifier()
	if identifier == nil {
		l.AddError(fmt.Errorf("metric declaration missing identifier"))
		_, _ = l.symbolStack.Pop()
		return
	}

	metricName := identifier.GetText()
	if metricName == "" {
		l.AddError(fmt.Errorf("metric declaration has empty identifier"))
		_, _ = l.symbolStack.Pop()
		return
	}

	// Collect input definitions from InputRefDefinition entries in Definitions
	var inputs []parsing.ParserInputDefinition
	for _, def := range metricScope.Definitions {
		if def.DefinitionType == symboltable.InputRefDefinition {
			if inputDef, ok := def.Data.(parsing.ParserInputDefinition); ok {
				inputs = append(inputs, inputDef)
			}
		}
	}

	// Collect calculate postfix from expression definitions in order
	calculatePostfix := l.collectCalculatePostfix(metricScope)

	// Get source location from the identifier token
	location := l.makeSourceLocation(identifier.GetStart())

	// Create the ParserMetricDefinition
	metricDef := &parsing.ParserMetricDefinition{
		Name:             metricName,
		Inputs:           inputs,
		CalculatePostfix: calculatePostfix,
		ShortForm:        false,
		SourceLocation:   location,
	}

	// Store the ParserMetricDefinition in the metric scope's Data field
	metricScope.Data = metricDef

	// Also push a MetricScopeDefinition entry into Definitions for analysis phase retrieval
	nameCopy := metricName
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, location, metricDef)

	// Pop the Metric scope from the symbol stack
	_, err = l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting metric declaration: %w", err))
	}
}

// EnterInputDeclaration is called when entering an input declaration.
// It resets the current time shift state to ensure no stale data leaks
// between input declarations.
func (l *CPLListener) EnterInputDeclaration(ctx *antlr.InputDeclarationContext) {
	l.currentTimeShift = nil
	l.hasTimeShift = false
	l.timeShiftExpressionScope = nil
}

// ExitInputDeclaration is called when exiting an input declaration.
// It extracts the input kind (metric or dimension) and name, creates a
// ParserInputDefinition, and defines it in the current metric scope.
//
// Example input declarations:
//
//	metric some_metric
//	dimension some_dimension
func (l *CPLListener) ExitInputDeclaration(ctx *antlr.InputDeclarationContext) {
	// Peek the metric scope
	metricScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error in input declaration: %w", err))
		return
	}

	// Get input kind from ctx.InputKind().GetText() ("metric" or "dimension")
	inputKindCtx := ctx.InputKind()
	if inputKindCtx == nil {
		l.AddError(fmt.Errorf("input declaration missing input kind"))
		return
	}
	inputKind := inputKindCtx.GetText()

	// Get input name from ctx.Identifier().GetText()
	identifier := ctx.Identifier()
	if identifier == nil {
		l.AddError(fmt.Errorf("input declaration missing identifier"))
		return
	}
	inputName := identifier.GetText()

	// Get source location
	location := l.makeSourceLocation(identifier.GetStart())

	// Extract optional alias from inputAsClause
	var alias string
	if inputAsClause := ctx.InputAsClause(); inputAsClause != nil {
		if aliasIdentifier := inputAsClause.Identifier(); aliasIdentifier != nil {
			alias = aliasIdentifier.GetText()
		}
	}

	// Extract optional where conditions from whereExpression
	var whereClause *parsing.ParserWhereClause
	if whereExpr := ctx.WhereExpression(); whereExpr != nil {
		if orExpr := whereExpr.OrExpression(); orExpr != nil {
			whereClause = &parsing.ParserWhereClause{}
			for _, andExpr := range orExpr.AllAndExpression() {
				andGroup := parsing.ParserWhereAndGroup{}
				for _, aspectExpr := range andExpr.AllAspectExpression() {
					cond := parsing.ParserWhereCondition{}
					if aspectExpr.ACTUALS_KEYWORD() != nil {
						cond.Aspect = "ACTUALS"
					} else {
						cond.Aspect = "FORECAST"
					}
					cond.Negated = aspectExpr.NOT_KEYWORD() != nil
					andGroup.Conditions = append(andGroup.Conditions, cond)
				}
				whereClause.AndGroups = append(whereClause.AndGroups, andGroup)
			}
		}
	}

	// Create ParserInputDefinition
	inputDef := parsing.ParserInputDefinition{
		Kind:         inputKind,
		Name:         inputName,
		Alias:        alias,
		WhereClause:  whereClause,
		NameLocation: location,
	}

	// Attach time shift if present
	if l.currentTimeShift != nil {
		inputDef.TimeShift = l.currentTimeShift
		l.currentTimeShift = nil // Reset for next input declaration
	}

	// Define as InputRefDefinition in the current metric scope
	// Store the ParserInputDefinition in the definition's Data field
	err = metricScope.DefineWithData(
		&inputName,
		symboltable.InputRefDefinition,
		location,
		inputDef,
	)
	if err != nil {
		l.AddError(fmt.Errorf("error defining input reference: %w", err))
	}
}

// EnterShiftExpression is called when entering a shift expression inside
// backward() or forward(). It pushes a TempScope so that expression
// definitions (operands and operators) are accumulated in postfix order.
// This scope is popped in ExitShiftExpression and its definitions are
// converted to PostfixElements for the time shift.
func (l *CPLListener) EnterShiftExpression(ctx *antlr.ShiftExpressionContext) {
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering shift expression: %w", err))
		return
	}
	l.timeShiftExpressionScope = current.AddChild(symboltable.TempScope)
	l.symbolStack.Push(l.timeShiftExpressionScope)
}

// ExitShiftExpression is called when exiting a shift expression inside
// backward() or forward(). It pops the TempScope so that the accumulated
// expression definitions are preserved in timeShiftExpressionScope for
// later conversion to PostfixElements by resolveTimeShift.
func (l *CPLListener) ExitShiftExpression(ctx *antlr.ShiftExpressionContext) {
	_, err := l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting shift expression: %w", err))
	}
}

// resolveTimeShift extracts field and amount/expression from a ShiftExpressionContext
// and builds a ParserTimeShift. Used by both ExitBackwardClause and ExitForwardClause.
//
// Parameters:
//   - shiftExpr: The ANTLR shift expression context
//   - direction: "backward" or "forward"
//   - ctxStart: The start token for error location reporting
//
// Returns:
//   - A populated ParserTimeShift (even on partial failure, so nil-checking is unnecessary)
func (l *CPLListener) resolveTimeShift(
	shiftExpr antlr.IShiftExpressionContext,
	direction string,
	ctxStart antlrruntime.Token,
) *parsing.ParserTimeShift {
	if shiftExpr == nil {
		l.AddError(fmt.Errorf("%s clause missing shift expression", direction))
		return &parsing.ParserTimeShift{Direction: direction, Field: "", Amount: 1.0}
	}

	identifier := shiftExpr.Identifier()
	if identifier == nil {
		l.AddError(fmt.Errorf("%s clause missing field identifier", direction))
		return &parsing.ParserTimeShift{Direction: direction, Field: "", Amount: 1.0}
	}
	field := identifier.GetText()

	// If the grammar produced an expression after the comma, the listener's
	// EnterShiftExpression/ExitShiftExpression pair pushed a TempScope and
	// the expression definitions accumulated there.
	// Only build Expression postfix when definitions were actually accumulated;
	// an empty TempScope (e.g., from backward(year)) means no expression was provided.
	location := l.makeSourceLocation(identifier.GetStart())
	if l.timeShiftExpressionScope != nil && len(l.timeShiftExpressionScope.Definitions) > 0 {
		exprPostfix := l.symbolDefsToPostfixElements(l.timeShiftExpressionScope.Definitions)
		l.timeShiftExpressionScope = nil // Reset for next shift expression
		return &parsing.ParserTimeShift{
			Direction:      direction,
			Field:          field,
			Expression:     exprPostfix,
			Amount:         1.0,
			SourceLocation: location,
		}
	}
	// Reset scope reference even when no expression definitions were accumulated
	if l.timeShiftExpressionScope != nil {
		l.timeShiftExpressionScope = nil
	}

	// No expression provided — default amount 1.0
	return &parsing.ParserTimeShift{
		Direction:      direction,
		Field:          field,
		Amount:         1.0,
		SourceLocation: location,
	}
}

// ExitBackwardClause is called when exiting a backward() clause on an input declaration.
// It extracts the field name and optional shift expression from the shift expression
// and stores them in currentTimeShift for attachment to the input definition.
// If a time shift clause has already been processed on the current input declaration
// (i.e., both backward() and forward() are specified), an error is emitted.
func (l *CPLListener) ExitBackwardClause(ctx *antlr.BackwardClauseContext) {
	if l.hasTimeShift {
		l.AddError(fmt.Errorf("input declaration at line %d, column %d cannot have both backward() and forward() time shifts",
			ctx.GetStart().GetLine(), ctx.GetStart().GetColumn()))
		return
	}
	l.hasTimeShift = true
	l.currentTimeShift = l.resolveTimeShift(ctx.ShiftExpression(), "backward", ctx.GetStart())
}

// ExitForwardClause is called when exiting a forward() clause on an input declaration.
// It extracts the field name and optional shift expression from the shift expression
// and stores them in currentTimeShift for attachment to the input definition.
// If a time shift clause has already been processed on the current input declaration
// (i.e., both backward() and forward() are specified), an error is emitted.
func (l *CPLListener) ExitForwardClause(ctx *antlr.ForwardClauseContext) {
	if l.hasTimeShift {
		l.AddError(fmt.Errorf("input declaration at line %d, column %d cannot have both backward() and forward() time shifts",
			ctx.GetStart().GetLine(), ctx.GetStart().GetColumn()))
		return
	}
	l.hasTimeShift = true
	l.currentTimeShift = l.resolveTimeShift(ctx.ShiftExpression(), "forward", ctx.GetStart())
}

// EnterShortFormDeclaration is called when entering a short form metric declaration.
// It pushes a new MetricScope onto the symbol stack so that expression definitions
// (operands and operators) are accumulated in postfix order by the root listener's
// expression handler methods (ExitPrimaryExpr, ExitAdditionExpr, etc.).
//
// Short form syntax: <name> = <calculation>
//
// Example:
//
//	revenue = demand_forecast * Price
//
// This corresponds to the 'shortFormDeclaration' rule in CPLMetrics.g4.
//
// The MetricScope is popped and the ParserMetricDefinition is assembled in
// ExitShortFormDeclaration.
func (l *CPLListener) EnterShortFormDeclaration(ctx *antlr.ShortFormDeclarationContext) {
	l.logEnter("ShortFormDeclaration", &ctx.BaseParserRuleContext)

	// Get the current table and create a child for this Metric scope
	current, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error entering short form declaration: %w", err))
		return
	}

	// Push a MetricScope so that expression definitions accumulate here
	l.symbolStack.Push(current.AddChild(symboltable.MetricScope))
}

// ExitShortFormDeclaration is called when exiting a short form metric declaration.
// It extracts the metric name, collects the calculate postfix from expression
// definitions accumulated in the MetricScope, creates a ParserMetricDefinition
// with ShortForm: true, and stores it in the MetricScope's Data field.
//
// The resulting ParserMetricDefinition is identical to long form except for
// the ShortForm flag. Inputs are derived during analysis by scanning the CIL
// tree for CILInputRef nodes.
//
// Scope handling: the MetricScope was pushed in EnterShortFormDeclaration,
// ensuring expression definitions accumulated there via shouldPushExpressionDefs.
func (l *CPLListener) ExitShortFormDeclaration(ctx *antlr.ShortFormDeclarationContext) {
	l.logExit("ShortFormDeclaration", &ctx.BaseParserRuleContext)

	// Get the MetricScope before popping
	metricScope, err := l.symbolStack.Peek()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting short form declaration: %w", err))
		return
	}

	// Get the metric name from the identifier
	identifier := ctx.Identifier()
	if identifier == nil {
		l.AddError(fmt.Errorf("short form declaration missing identifier"))
		_, _ = l.symbolStack.Pop()
		return
	}

	metricName := identifier.GetText()
	if metricName == "" {
		l.AddError(fmt.Errorf("short form declaration has empty identifier"))
		_, _ = l.symbolStack.Pop()
		return
	}

	// Collect calculate postfix from expression definitions in order
	calculatePostfix := l.collectCalculatePostfix(metricScope)

	// Get source location from the identifier token
	location := l.makeSourceLocation(identifier.GetStart())

	// Create ParserMetricDefinition with ShortForm: true and no explicit inputs
	// (inputs are derived during analysis by scanning the CIL tree)
	metricDef := &parsing.ParserMetricDefinition{
		Name:             metricName,
		Inputs:           nil, // Derived during analysis from CILInputRef nodes
		CalculatePostfix: calculatePostfix,
		ShortForm:        true,
		SourceLocation:   location,
	}

	// Store the ParserMetricDefinition in the MetricScope's Data field
	metricScope.Data = metricDef

	// Also push a MetricScopeDefinition entry into Definitions for analysis phase retrieval
	nameCopy := metricName
	_ = metricScope.DefineWithData(&nameCopy, symboltable.MetricScopeDefinition, location, metricDef)

	// Pop the MetricScope from the symbol stack
	_, err = l.symbolStack.Pop()
	if err != nil {
		l.AddError(fmt.Errorf("symbol stack error exiting short form declaration: %w", err))
	}
}

// collectCalculatePostfix extracts the calculate postfix sequence from a
// MetricScope's Definitions in insertion order. This is used by both
// ExitMetricDeclaration (long form) and ExitShortFormDeclaration (short form)
// to produce the postfix expression consumed by buildCILTree during analysis.
//
// Postfix Element Types:
//   - integer literals → PostfixLiteral{Value: float64(v)}
//   - string literal data that parses as integer → PostfixLiteral{Value: float64(v)}
//   - identifier names → PostfixIdentifier{Name: name}
//   - operator definitions → PostfixOperator{DefType: definitionType}
//
// Errors for unparseable literals are accumulated via AddError and the
// offending entry is skipped.
func (l *CPLListener) collectCalculatePostfix(metricScope *symboltable.SymbolTable) []parsing.PostfixElement {
	var postfix []parsing.PostfixElement
	for _, def := range metricScope.Definitions {
		switch def.DefinitionType {
		case symboltable.LiteralDefinition:
			if literalValue, ok := def.Data.(symboltable.IntData); ok {
				postfix = append(postfix, parsing.PostfixLiteral{Value: float64(literalValue.Value)})
			} else if literalStr, ok := def.Data.(symboltable.StringData); ok {
				if intVal, err := strconv.Atoi(literalStr.Value); err == nil {
					postfix = append(postfix, parsing.PostfixLiteral{Value: float64(intVal)})
				} else {
					l.AddError(fmt.Errorf("literal value %q cannot be parsed as integer", literalStr.Value))
				}
			}
		case symboltable.Identifier:
			if identifierName, ok := def.Data.(symboltable.StringData); ok {
				postfix = append(postfix, parsing.PostfixIdentifier{Name: identifierName.Value})
			}
		case symboltable.UnitReferenceDefinition:
			if unitRef, ok := def.Data.(ParserUnitReference); ok {
				postfix = append(postfix, parsing.PostfixIdentifier{Name: unitRef.Name})
			}
		case symboltable.NumberDefinition:
			if floatData, ok := def.Data.(symboltable.FloatData); ok {
				postfix = append(postfix, parsing.PostfixLiteral{Value: floatData.Value})
			} else if intData, ok := def.Data.(symboltable.IntData); ok {
				postfix = append(postfix, parsing.PostfixLiteral{Value: float64(intData.Value)})
			}
		case symboltable.PlusOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.PlusOperatorDefinition})
		case symboltable.MinusOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.MinusOperatorDefinition})
		case symboltable.MultiplyOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.MultiplyOperatorDefinition})
		case symboltable.DivideOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.DivideOperatorDefinition})
		case symboltable.PowerOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.PowerOperatorDefinition})
		case symboltable.NegateOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.NegateOperatorDefinition})
		case symboltable.AndOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.AndOperatorDefinition})
		case symboltable.OrOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.OrOperatorDefinition})
		case symboltable.NotOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.NotOperatorDefinition})
		case symboltable.EqOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.EqOperatorDefinition})
		case symboltable.NotEqOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.NotEqOperatorDefinition})
		case symboltable.LessThanOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.LessThanOperatorDefinition})
		case symboltable.GreaterThanOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.GreaterThanOperatorDefinition})
		case symboltable.LessThanEqOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.LessThanEqOperatorDefinition})
		case symboltable.GreaterThanEqOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.GreaterThanEqOperatorDefinition})
		case symboltable.TrueLiteralDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.TrueLiteralDefinition})
		case symboltable.FalseLiteralDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.FalseLiteralDefinition})
		case symboltable.NilLiteralDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.NilLiteralDefinition})
		}
	}
	return postfix
}

// symbolDefsToPostfixElements converts a slice of SymbolDefinition entries to
// typed PostfixElements. This is a zero-failure variant used for time-shift
// expression definitions: every SymbolDefinition maps to a PostfixElement
// (errors are silently skipped for literal fallback).
func (l *CPLListener) symbolDefsToPostfixElements(defs []symboltable.SymbolDefinition) []parsing.PostfixElement {
	postfix := make([]parsing.PostfixElement, 0, len(defs))
	for _, def := range defs {
		switch def.DefinitionType {
		case symboltable.Identifier:
			if identifierName, ok := def.Data.(symboltable.StringData); ok {
				postfix = append(postfix, parsing.PostfixIdentifier{Name: identifierName.Value})
			}
		case symboltable.LiteralDefinition:
			if literalValue, ok := def.Data.(symboltable.IntData); ok {
				postfix = append(postfix, parsing.PostfixLiteral{Value: float64(literalValue.Value)})
			} else if literalStr, ok := def.Data.(symboltable.StringData); ok {
				if intVal, err := strconv.Atoi(literalStr.Value); err == nil {
					postfix = append(postfix, parsing.PostfixLiteral{Value: float64(intVal)})
				}
			}
		case symboltable.NumberDefinition:
			if floatData, ok := def.Data.(symboltable.FloatData); ok {
				postfix = append(postfix, parsing.PostfixLiteral{Value: floatData.Value})
			} else if intData, ok := def.Data.(symboltable.IntData); ok {
				postfix = append(postfix, parsing.PostfixLiteral{Value: float64(intData.Value)})
			}
		case symboltable.PlusOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.PlusOperatorDefinition})
		case symboltable.MinusOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.MinusOperatorDefinition})
		case symboltable.MultiplyOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.MultiplyOperatorDefinition})
		case symboltable.DivideOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.DivideOperatorDefinition})
		case symboltable.PowerOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.PowerOperatorDefinition})
		case symboltable.NegateOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.NegateOperatorDefinition})
		case symboltable.AndOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.AndOperatorDefinition})
		case symboltable.OrOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.OrOperatorDefinition})
		case symboltable.NotOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.NotOperatorDefinition})
		case symboltable.EqOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.EqOperatorDefinition})
		case symboltable.NotEqOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.NotEqOperatorDefinition})
		case symboltable.LessThanOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.LessThanOperatorDefinition})
		case symboltable.GreaterThanOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.GreaterThanOperatorDefinition})
		case symboltable.LessThanEqOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.LessThanEqOperatorDefinition})
		case symboltable.GreaterThanEqOperatorDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.GreaterThanEqOperatorDefinition})
		case symboltable.TrueLiteralDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.TrueLiteralDefinition})
		case symboltable.FalseLiteralDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.FalseLiteralDefinition})
		case symboltable.NilLiteralDefinition:
			postfix = append(postfix, parsing.PostfixOperator{DefType: symboltable.NilLiteralDefinition})
		}
	}
	return postfix
}

// ExitLiteral is called when exiting a literal node.
// If the current scope is an expression-accumulating scope (MetricScope,
// TempScope, or conditional body scopes), the literal value is added
// as a LiteralDefinition to the scope's Definitions.
//
// This supports building the postfix expression for calculate clauses
// and conditional expressions.
//
// Supported literal types:
//   - INTEGER: stored as int
//   - TRUE/FALSE: stored as symboltable.TrueLiteralDefinition/FalseLiteralDefinition
//   - NIL: stored as symboltable.NilLiteralDefinition
//   - Non-integer: stored as string (fallback)
func (l *CPLListener) ExitLiteral(ctx *antlr.LiteralContext) {
	// Peek the current scope
	currentScope, err := l.symbolStack.Peek()
	if err != nil {
		return
	}

	// Check if current scope is an expression-accumulating scope
	if !l.shouldPushExpressionDefs() {
		return
	}

	// Get the literal text
	literalText := ctx.GetText()
	if literalText == "" {
		return
	}

	// Handle boolean and Nil literals for conditional expressions
	switch literalText {
	case "True":
		err = currentScope.DefineWithData(
			nil,
			symboltable.TrueLiteralDefinition,
			l.makeSourceLocation(ctx.GetStart()),
			symboltable.BoolData{Value: true},
		)
		if err != nil {
			l.AddError(fmt.Errorf("error defining true literal: %w", err))
		}
		return
	case "False":
		err = currentScope.DefineWithData(
			nil,
			symboltable.FalseLiteralDefinition,
			l.makeSourceLocation(ctx.GetStart()),
			symboltable.BoolData{Value: false},
		)
		if err != nil {
			l.AddError(fmt.Errorf("error defining false literal: %w", err))
		}
		return
	case "Nil":
		err = currentScope.DefineWithData(
			nil,
			symboltable.NilLiteralDefinition,
			l.makeSourceLocation(ctx.GetStart()),
			nil,
		)
		if err != nil {
			l.AddError(fmt.Errorf("error defining nil literal: %w", err))
		}
		return
	}

	// Try to parse as integer
	intValue, err := strconv.Atoi(literalText)
	if err != nil {
		// If not an integer, store as string
		err = currentScope.DefineWithData(
			nil,
			symboltable.LiteralDefinition,
			l.makeSourceLocation(ctx.GetStart()),
			symboltable.StringData{Value: literalText},
		)
		if err != nil {
			l.AddError(fmt.Errorf("error defining literal: %w", err))
		}
		return
	}

	// Store as integer
	err = currentScope.DefineWithData(
		nil,
		symboltable.LiteralDefinition,
		l.makeSourceLocation(ctx.GetStart()),
		symboltable.IntData{Value: intValue},
	)
	if err != nil {
		l.AddError(fmt.Errorf("error defining literal: %w", err))
	}
}
