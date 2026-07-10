// Code generated from ChronoconePlanningLanguage.g4 by ANTLR 4.13.1. DO NOT EDIT.

package antlr // ChronoconePlanningLanguage
import "github.com/antlr4-go/antlr/v4"

type BaseChronoconePlanningLanguageVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitDocument(ctx *DocumentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitStatement(ctx *StatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitIdentifier(ctx *IdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitValue(ctx *ValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitTypeReference(ctx *TypeReferenceContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitParameterList(ctx *ParameterListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitParameter(ctx *ParameterContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitParameterValue(ctx *ParameterValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitIndent(ctx *IndentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitIdentifierList(ctx *IdentifierListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitExpression(ctx *ExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitConditionalExpr(ctx *ConditionalExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitIfBranch(ctx *IfBranchContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitElseIfBranch(ctx *ElseIfBranchContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitElseBranch(ctx *ElseBranchContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitWhenExpr(ctx *WhenExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitWhenClause(ctx *WhenClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitDefaultClause(ctx *DefaultClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitLogicalOrExpr(ctx *LogicalOrExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitLogicalAndExpr(ctx *LogicalAndExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitNotExpr(ctx *NotExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitEqualityExpr(ctx *EqualityExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitRelationalExpr(ctx *RelationalExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitAdditionExpr(ctx *AdditionExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitMultiplicationExpr(ctx *MultiplicationExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitUnaryExpr(ctx *UnaryExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitPrimaryExpr(ctx *PrimaryExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitLiteral(ctx *LiteralContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitConstraintList(ctx *ConstraintListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitConstraint(ctx *ConstraintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitConstraintValue(ctx *ConstraintValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitNestedConstraintValue(ctx *NestedConstraintValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitArithmeticExpr(ctx *ArithmeticExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitArithmeticTerm(ctx *ArithmeticTermContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitArithmeticFactor(ctx *ArithmeticFactorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitMultOperator(ctx *MultOperatorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitDivOperator(ctx *DivOperatorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitPowOperator(ctx *PowOperatorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitPlusOperator(ctx *PlusOperatorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitMinusOperator(ctx *MinusOperatorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitConceptsSection(ctx *ConceptsSectionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitConceptDefinition(ctx *ConceptDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitConceptLine(ctx *ConceptLineContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitConditionalSpec(ctx *ConditionalSpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitConceptExtension(ctx *ConceptExtensionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitInsertionPoint(ctx *InsertionPointContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitGeneratorPatternList(ctx *GeneratorPatternListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitGeneratorPattern(ctx *GeneratorPatternContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitValuesPatternGenerator(ctx *ValuesPatternGeneratorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitIntRangeGenerator(ctx *IntRangeGeneratorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitIntRangeParams(ctx *IntRangeParamsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitConceptPatternGenerator(ctx *ConceptPatternGeneratorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitConceptRange(ctx *ConceptRangeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitConceptValueList(ctx *ConceptValueListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitMeasuresSection(ctx *MeasuresSectionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitMeasureDeclaration(ctx *MeasureDeclarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitMeasureParameterList(ctx *MeasureParameterListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitMeasureParameter(ctx *MeasureParameterContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitUnitParameter(ctx *UnitParameterContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitMeasureConversionClause(ctx *MeasureConversionClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitMeasureConversionPair(ctx *MeasureConversionPairContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitMeasureCalculationClause(ctx *MeasureCalculationClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitTypesSection(ctx *TypesSectionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitTypeDefinition(ctx *TypeDefinitionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitBaseType(ctx *BaseTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitUnitsSection(ctx *UnitsSectionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitUnitDeclaration(ctx *UnitDeclarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitConversionClause(ctx *ConversionClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitConversionPair(ctx *ConversionPairContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitPatternClause(ctx *PatternClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitPattern(ctx *PatternContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitPatternBoundVariable(ctx *PatternBoundVariableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitCalculationClause(ctx *CalculationClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitConnectorsSection(ctx *ConnectorsSectionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitConnectorDeclaration(ctx *ConnectorDeclarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitMetricsSection(ctx *MetricsSectionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitShortFormDeclaration(ctx *ShortFormDeclarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitMetricDeclaration(ctx *MetricDeclarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitInputsBlock(ctx *InputsBlockContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitInputDeclaration(ctx *InputDeclarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitInputKind(ctx *InputKindContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitInputAsClause(ctx *InputAsClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitWhereExpression(ctx *WhereExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitOrExpression(ctx *OrExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitAndExpression(ctx *AndExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitAspectExpression(ctx *AspectExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitBackwardClause(ctx *BackwardClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitForwardClause(ctx *ForwardClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitShiftExpression(ctx *ShiftExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitCalculateClause(ctx *CalculateClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitLookupsSection(ctx *LookupsSectionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitLookupDeclaration(ctx *LookupDeclarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitDimensionsClause(ctx *DimensionsClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitCelltypeClause(ctx *CelltypeClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitAggregatorsSection(ctx *AggregatorsSectionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitAggregatorDeclaration(ctx *AggregatorDeclarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitToTarget(ctx *ToTargetContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitFnCall(ctx *FnCallContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitSpreadersSection(ctx *SpreadersSectionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseChronoconePlanningLanguageVisitor) VisitSpreaderDeclaration(ctx *SpreaderDeclarationContext) interface{} {
	return v.VisitChildren(ctx)
}
