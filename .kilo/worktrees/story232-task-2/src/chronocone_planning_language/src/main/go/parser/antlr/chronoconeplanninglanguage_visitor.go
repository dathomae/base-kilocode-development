// Code generated from ChronoconePlanningLanguage.g4 by ANTLR 4.13.1. DO NOT EDIT.

package antlr // ChronoconePlanningLanguage
import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by ChronoconePlanningLanguageParser.
type ChronoconePlanningLanguageVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#document.
	VisitDocument(ctx *DocumentContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#statement.
	VisitStatement(ctx *StatementContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#identifier.
	VisitIdentifier(ctx *IdentifierContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#value.
	VisitValue(ctx *ValueContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#typeReference.
	VisitTypeReference(ctx *TypeReferenceContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#parameterList.
	VisitParameterList(ctx *ParameterListContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#parameter.
	VisitParameter(ctx *ParameterContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#parameterValue.
	VisitParameterValue(ctx *ParameterValueContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#indent.
	VisitIndent(ctx *IndentContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#identifierList.
	VisitIdentifierList(ctx *IdentifierListContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#expression.
	VisitExpression(ctx *ExpressionContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#conditionalExpr.
	VisitConditionalExpr(ctx *ConditionalExprContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#ifBranch.
	VisitIfBranch(ctx *IfBranchContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#elseIfBranch.
	VisitElseIfBranch(ctx *ElseIfBranchContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#elseBranch.
	VisitElseBranch(ctx *ElseBranchContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#whenExpr.
	VisitWhenExpr(ctx *WhenExprContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#whenClause.
	VisitWhenClause(ctx *WhenClauseContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#defaultClause.
	VisitDefaultClause(ctx *DefaultClauseContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#logicalOrExpr.
	VisitLogicalOrExpr(ctx *LogicalOrExprContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#logicalAndExpr.
	VisitLogicalAndExpr(ctx *LogicalAndExprContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#notExpr.
	VisitNotExpr(ctx *NotExprContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#equalityExpr.
	VisitEqualityExpr(ctx *EqualityExprContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#relationalExpr.
	VisitRelationalExpr(ctx *RelationalExprContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#additionExpr.
	VisitAdditionExpr(ctx *AdditionExprContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#multiplicationExpr.
	VisitMultiplicationExpr(ctx *MultiplicationExprContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#unaryExpr.
	VisitUnaryExpr(ctx *UnaryExprContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#primaryExpr.
	VisitPrimaryExpr(ctx *PrimaryExprContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#literal.
	VisitLiteral(ctx *LiteralContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#constraintList.
	VisitConstraintList(ctx *ConstraintListContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#constraint.
	VisitConstraint(ctx *ConstraintContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#constraintValue.
	VisitConstraintValue(ctx *ConstraintValueContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#nestedConstraintValue.
	VisitNestedConstraintValue(ctx *NestedConstraintValueContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#arithmeticExpr.
	VisitArithmeticExpr(ctx *ArithmeticExprContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#arithmeticTerm.
	VisitArithmeticTerm(ctx *ArithmeticTermContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#arithmeticFactor.
	VisitArithmeticFactor(ctx *ArithmeticFactorContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#multOperator.
	VisitMultOperator(ctx *MultOperatorContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#divOperator.
	VisitDivOperator(ctx *DivOperatorContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#powOperator.
	VisitPowOperator(ctx *PowOperatorContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#plusOperator.
	VisitPlusOperator(ctx *PlusOperatorContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#minusOperator.
	VisitMinusOperator(ctx *MinusOperatorContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#conceptsSection.
	VisitConceptsSection(ctx *ConceptsSectionContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#conceptDefinition.
	VisitConceptDefinition(ctx *ConceptDefinitionContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#conceptLine.
	VisitConceptLine(ctx *ConceptLineContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#conditionalSpec.
	VisitConditionalSpec(ctx *ConditionalSpecContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#conceptExtension.
	VisitConceptExtension(ctx *ConceptExtensionContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#insertionPoint.
	VisitInsertionPoint(ctx *InsertionPointContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#generatorPatternList.
	VisitGeneratorPatternList(ctx *GeneratorPatternListContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#generatorPattern.
	VisitGeneratorPattern(ctx *GeneratorPatternContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#valuesPatternGenerator.
	VisitValuesPatternGenerator(ctx *ValuesPatternGeneratorContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#intRangeGenerator.
	VisitIntRangeGenerator(ctx *IntRangeGeneratorContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#intRangeParams.
	VisitIntRangeParams(ctx *IntRangeParamsContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#conceptPatternGenerator.
	VisitConceptPatternGenerator(ctx *ConceptPatternGeneratorContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#conceptRange.
	VisitConceptRange(ctx *ConceptRangeContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#conceptValueList.
	VisitConceptValueList(ctx *ConceptValueListContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#measuresSection.
	VisitMeasuresSection(ctx *MeasuresSectionContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#measureDeclaration.
	VisitMeasureDeclaration(ctx *MeasureDeclarationContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#measureParameterList.
	VisitMeasureParameterList(ctx *MeasureParameterListContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#measureParameter.
	VisitMeasureParameter(ctx *MeasureParameterContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#unitParameter.
	VisitUnitParameter(ctx *UnitParameterContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#measureConversionClause.
	VisitMeasureConversionClause(ctx *MeasureConversionClauseContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#measureConversionPair.
	VisitMeasureConversionPair(ctx *MeasureConversionPairContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#measureCalculationClause.
	VisitMeasureCalculationClause(ctx *MeasureCalculationClauseContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#typesSection.
	VisitTypesSection(ctx *TypesSectionContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#typeDefinition.
	VisitTypeDefinition(ctx *TypeDefinitionContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#baseType.
	VisitBaseType(ctx *BaseTypeContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#unitsSection.
	VisitUnitsSection(ctx *UnitsSectionContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#unitDeclaration.
	VisitUnitDeclaration(ctx *UnitDeclarationContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#conversionClause.
	VisitConversionClause(ctx *ConversionClauseContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#conversionPair.
	VisitConversionPair(ctx *ConversionPairContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#patternClause.
	VisitPatternClause(ctx *PatternClauseContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#pattern.
	VisitPattern(ctx *PatternContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#patternBoundVariable.
	VisitPatternBoundVariable(ctx *PatternBoundVariableContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#calculationClause.
	VisitCalculationClause(ctx *CalculationClauseContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#connectorsSection.
	VisitConnectorsSection(ctx *ConnectorsSectionContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#connectorDeclaration.
	VisitConnectorDeclaration(ctx *ConnectorDeclarationContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#metricsSection.
	VisitMetricsSection(ctx *MetricsSectionContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#shortFormDeclaration.
	VisitShortFormDeclaration(ctx *ShortFormDeclarationContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#metricDeclaration.
	VisitMetricDeclaration(ctx *MetricDeclarationContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#inputsBlock.
	VisitInputsBlock(ctx *InputsBlockContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#inputDeclaration.
	VisitInputDeclaration(ctx *InputDeclarationContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#inputKind.
	VisitInputKind(ctx *InputKindContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#inputAsClause.
	VisitInputAsClause(ctx *InputAsClauseContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#whereExpression.
	VisitWhereExpression(ctx *WhereExpressionContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#orExpression.
	VisitOrExpression(ctx *OrExpressionContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#andExpression.
	VisitAndExpression(ctx *AndExpressionContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#aspectExpression.
	VisitAspectExpression(ctx *AspectExpressionContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#backwardClause.
	VisitBackwardClause(ctx *BackwardClauseContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#forwardClause.
	VisitForwardClause(ctx *ForwardClauseContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#shiftExpression.
	VisitShiftExpression(ctx *ShiftExpressionContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#calculateClause.
	VisitCalculateClause(ctx *CalculateClauseContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#lookupsSection.
	VisitLookupsSection(ctx *LookupsSectionContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#lookupDeclaration.
	VisitLookupDeclaration(ctx *LookupDeclarationContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#dimensionsClause.
	VisitDimensionsClause(ctx *DimensionsClauseContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#celltypeClause.
	VisitCelltypeClause(ctx *CelltypeClauseContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#aggregatorsSection.
	VisitAggregatorsSection(ctx *AggregatorsSectionContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#aggregatorDeclaration.
	VisitAggregatorDeclaration(ctx *AggregatorDeclarationContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#toTarget.
	VisitToTarget(ctx *ToTargetContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#fnCall.
	VisitFnCall(ctx *FnCallContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#spreadersSection.
	VisitSpreadersSection(ctx *SpreadersSectionContext) interface{}

	// Visit a parse tree produced by ChronoconePlanningLanguageParser#spreaderDeclaration.
	VisitSpreaderDeclaration(ctx *SpreaderDeclarationContext) interface{}
}
