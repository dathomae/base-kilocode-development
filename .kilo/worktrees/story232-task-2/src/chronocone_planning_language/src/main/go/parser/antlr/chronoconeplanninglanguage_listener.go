// Code generated from ChronoconePlanningLanguage.g4 by ANTLR 4.13.1. DO NOT EDIT.

package antlr // ChronoconePlanningLanguage
import "github.com/antlr4-go/antlr/v4"

// ChronoconePlanningLanguageListener is a complete listener for a parse tree produced by ChronoconePlanningLanguageParser.
type ChronoconePlanningLanguageListener interface {
	antlr.ParseTreeListener

	// EnterDocument is called when entering the document production.
	EnterDocument(c *DocumentContext)

	// EnterStatement is called when entering the statement production.
	EnterStatement(c *StatementContext)

	// EnterIdentifier is called when entering the identifier production.
	EnterIdentifier(c *IdentifierContext)

	// EnterValue is called when entering the value production.
	EnterValue(c *ValueContext)

	// EnterTypeReference is called when entering the typeReference production.
	EnterTypeReference(c *TypeReferenceContext)

	// EnterParameterList is called when entering the parameterList production.
	EnterParameterList(c *ParameterListContext)

	// EnterParameter is called when entering the parameter production.
	EnterParameter(c *ParameterContext)

	// EnterParameterValue is called when entering the parameterValue production.
	EnterParameterValue(c *ParameterValueContext)

	// EnterIndent is called when entering the indent production.
	EnterIndent(c *IndentContext)

	// EnterIdentifierList is called when entering the identifierList production.
	EnterIdentifierList(c *IdentifierListContext)

	// EnterExpression is called when entering the expression production.
	EnterExpression(c *ExpressionContext)

	// EnterConditionalExpr is called when entering the conditionalExpr production.
	EnterConditionalExpr(c *ConditionalExprContext)

	// EnterIfBranch is called when entering the ifBranch production.
	EnterIfBranch(c *IfBranchContext)

	// EnterElseIfBranch is called when entering the elseIfBranch production.
	EnterElseIfBranch(c *ElseIfBranchContext)

	// EnterElseBranch is called when entering the elseBranch production.
	EnterElseBranch(c *ElseBranchContext)

	// EnterWhenExpr is called when entering the whenExpr production.
	EnterWhenExpr(c *WhenExprContext)

	// EnterWhenClause is called when entering the whenClause production.
	EnterWhenClause(c *WhenClauseContext)

	// EnterDefaultClause is called when entering the defaultClause production.
	EnterDefaultClause(c *DefaultClauseContext)

	// EnterLogicalOrExpr is called when entering the logicalOrExpr production.
	EnterLogicalOrExpr(c *LogicalOrExprContext)

	// EnterLogicalAndExpr is called when entering the logicalAndExpr production.
	EnterLogicalAndExpr(c *LogicalAndExprContext)

	// EnterNotExpr is called when entering the notExpr production.
	EnterNotExpr(c *NotExprContext)

	// EnterEqualityExpr is called when entering the equalityExpr production.
	EnterEqualityExpr(c *EqualityExprContext)

	// EnterRelationalExpr is called when entering the relationalExpr production.
	EnterRelationalExpr(c *RelationalExprContext)

	// EnterAdditionExpr is called when entering the additionExpr production.
	EnterAdditionExpr(c *AdditionExprContext)

	// EnterMultiplicationExpr is called when entering the multiplicationExpr production.
	EnterMultiplicationExpr(c *MultiplicationExprContext)

	// EnterUnaryExpr is called when entering the unaryExpr production.
	EnterUnaryExpr(c *UnaryExprContext)

	// EnterPrimaryExpr is called when entering the primaryExpr production.
	EnterPrimaryExpr(c *PrimaryExprContext)

	// EnterLiteral is called when entering the literal production.
	EnterLiteral(c *LiteralContext)

	// EnterConstraintList is called when entering the constraintList production.
	EnterConstraintList(c *ConstraintListContext)

	// EnterConstraint is called when entering the constraint production.
	EnterConstraint(c *ConstraintContext)

	// EnterConstraintValue is called when entering the constraintValue production.
	EnterConstraintValue(c *ConstraintValueContext)

	// EnterNestedConstraintValue is called when entering the nestedConstraintValue production.
	EnterNestedConstraintValue(c *NestedConstraintValueContext)

	// EnterArithmeticExpr is called when entering the arithmeticExpr production.
	EnterArithmeticExpr(c *ArithmeticExprContext)

	// EnterArithmeticTerm is called when entering the arithmeticTerm production.
	EnterArithmeticTerm(c *ArithmeticTermContext)

	// EnterArithmeticFactor is called when entering the arithmeticFactor production.
	EnterArithmeticFactor(c *ArithmeticFactorContext)

	// EnterMultOperator is called when entering the multOperator production.
	EnterMultOperator(c *MultOperatorContext)

	// EnterDivOperator is called when entering the divOperator production.
	EnterDivOperator(c *DivOperatorContext)

	// EnterPowOperator is called when entering the powOperator production.
	EnterPowOperator(c *PowOperatorContext)

	// EnterPlusOperator is called when entering the plusOperator production.
	EnterPlusOperator(c *PlusOperatorContext)

	// EnterMinusOperator is called when entering the minusOperator production.
	EnterMinusOperator(c *MinusOperatorContext)

	// EnterConceptsSection is called when entering the conceptsSection production.
	EnterConceptsSection(c *ConceptsSectionContext)

	// EnterConceptDefinition is called when entering the conceptDefinition production.
	EnterConceptDefinition(c *ConceptDefinitionContext)

	// EnterConceptLine is called when entering the conceptLine production.
	EnterConceptLine(c *ConceptLineContext)

	// EnterConditionalSpec is called when entering the conditionalSpec production.
	EnterConditionalSpec(c *ConditionalSpecContext)

	// EnterConceptExtension is called when entering the conceptExtension production.
	EnterConceptExtension(c *ConceptExtensionContext)

	// EnterInsertionPoint is called when entering the insertionPoint production.
	EnterInsertionPoint(c *InsertionPointContext)

	// EnterGeneratorPatternList is called when entering the generatorPatternList production.
	EnterGeneratorPatternList(c *GeneratorPatternListContext)

	// EnterGeneratorPattern is called when entering the generatorPattern production.
	EnterGeneratorPattern(c *GeneratorPatternContext)

	// EnterValuesPatternGenerator is called when entering the valuesPatternGenerator production.
	EnterValuesPatternGenerator(c *ValuesPatternGeneratorContext)

	// EnterIntRangeGenerator is called when entering the intRangeGenerator production.
	EnterIntRangeGenerator(c *IntRangeGeneratorContext)

	// EnterIntRangeParams is called when entering the intRangeParams production.
	EnterIntRangeParams(c *IntRangeParamsContext)

	// EnterConceptPatternGenerator is called when entering the conceptPatternGenerator production.
	EnterConceptPatternGenerator(c *ConceptPatternGeneratorContext)

	// EnterConceptRange is called when entering the conceptRange production.
	EnterConceptRange(c *ConceptRangeContext)

	// EnterConceptValueList is called when entering the conceptValueList production.
	EnterConceptValueList(c *ConceptValueListContext)

	// EnterMeasuresSection is called when entering the measuresSection production.
	EnterMeasuresSection(c *MeasuresSectionContext)

	// EnterMeasureDeclaration is called when entering the measureDeclaration production.
	EnterMeasureDeclaration(c *MeasureDeclarationContext)

	// EnterMeasureParameterList is called when entering the measureParameterList production.
	EnterMeasureParameterList(c *MeasureParameterListContext)

	// EnterMeasureParameter is called when entering the measureParameter production.
	EnterMeasureParameter(c *MeasureParameterContext)

	// EnterUnitParameter is called when entering the unitParameter production.
	EnterUnitParameter(c *UnitParameterContext)

	// EnterMeasureConversionClause is called when entering the measureConversionClause production.
	EnterMeasureConversionClause(c *MeasureConversionClauseContext)

	// EnterMeasureConversionPair is called when entering the measureConversionPair production.
	EnterMeasureConversionPair(c *MeasureConversionPairContext)

	// EnterMeasureCalculationClause is called when entering the measureCalculationClause production.
	EnterMeasureCalculationClause(c *MeasureCalculationClauseContext)

	// EnterTypesSection is called when entering the typesSection production.
	EnterTypesSection(c *TypesSectionContext)

	// EnterTypeDefinition is called when entering the typeDefinition production.
	EnterTypeDefinition(c *TypeDefinitionContext)

	// EnterBaseType is called when entering the baseType production.
	EnterBaseType(c *BaseTypeContext)

	// EnterUnitsSection is called when entering the unitsSection production.
	EnterUnitsSection(c *UnitsSectionContext)

	// EnterUnitDeclaration is called when entering the unitDeclaration production.
	EnterUnitDeclaration(c *UnitDeclarationContext)

	// EnterConversionClause is called when entering the conversionClause production.
	EnterConversionClause(c *ConversionClauseContext)

	// EnterConversionPair is called when entering the conversionPair production.
	EnterConversionPair(c *ConversionPairContext)

	// EnterPatternClause is called when entering the patternClause production.
	EnterPatternClause(c *PatternClauseContext)

	// EnterPattern is called when entering the pattern production.
	EnterPattern(c *PatternContext)

	// EnterPatternBoundVariable is called when entering the patternBoundVariable production.
	EnterPatternBoundVariable(c *PatternBoundVariableContext)

	// EnterCalculationClause is called when entering the calculationClause production.
	EnterCalculationClause(c *CalculationClauseContext)

	// EnterConnectorsSection is called when entering the connectorsSection production.
	EnterConnectorsSection(c *ConnectorsSectionContext)

	// EnterConnectorDeclaration is called when entering the connectorDeclaration production.
	EnterConnectorDeclaration(c *ConnectorDeclarationContext)

	// EnterMetricsSection is called when entering the metricsSection production.
	EnterMetricsSection(c *MetricsSectionContext)

	// EnterShortFormDeclaration is called when entering the shortFormDeclaration production.
	EnterShortFormDeclaration(c *ShortFormDeclarationContext)

	// EnterMetricDeclaration is called when entering the metricDeclaration production.
	EnterMetricDeclaration(c *MetricDeclarationContext)

	// EnterInputsBlock is called when entering the inputsBlock production.
	EnterInputsBlock(c *InputsBlockContext)

	// EnterInputDeclaration is called when entering the inputDeclaration production.
	EnterInputDeclaration(c *InputDeclarationContext)

	// EnterInputKind is called when entering the inputKind production.
	EnterInputKind(c *InputKindContext)

	// EnterInputAsClause is called when entering the inputAsClause production.
	EnterInputAsClause(c *InputAsClauseContext)

	// EnterWhereExpression is called when entering the whereExpression production.
	EnterWhereExpression(c *WhereExpressionContext)

	// EnterOrExpression is called when entering the orExpression production.
	EnterOrExpression(c *OrExpressionContext)

	// EnterAndExpression is called when entering the andExpression production.
	EnterAndExpression(c *AndExpressionContext)

	// EnterAspectExpression is called when entering the aspectExpression production.
	EnterAspectExpression(c *AspectExpressionContext)

	// EnterBackwardClause is called when entering the backwardClause production.
	EnterBackwardClause(c *BackwardClauseContext)

	// EnterForwardClause is called when entering the forwardClause production.
	EnterForwardClause(c *ForwardClauseContext)

	// EnterShiftExpression is called when entering the shiftExpression production.
	EnterShiftExpression(c *ShiftExpressionContext)

	// EnterCalculateClause is called when entering the calculateClause production.
	EnterCalculateClause(c *CalculateClauseContext)

	// EnterLookupsSection is called when entering the lookupsSection production.
	EnterLookupsSection(c *LookupsSectionContext)

	// EnterLookupDeclaration is called when entering the lookupDeclaration production.
	EnterLookupDeclaration(c *LookupDeclarationContext)

	// EnterDimensionsClause is called when entering the dimensionsClause production.
	EnterDimensionsClause(c *DimensionsClauseContext)

	// EnterCelltypeClause is called when entering the celltypeClause production.
	EnterCelltypeClause(c *CelltypeClauseContext)

	// EnterAggregatorsSection is called when entering the aggregatorsSection production.
	EnterAggregatorsSection(c *AggregatorsSectionContext)

	// EnterAggregatorDeclaration is called when entering the aggregatorDeclaration production.
	EnterAggregatorDeclaration(c *AggregatorDeclarationContext)

	// EnterToTarget is called when entering the toTarget production.
	EnterToTarget(c *ToTargetContext)

	// EnterFnCall is called when entering the fnCall production.
	EnterFnCall(c *FnCallContext)

	// EnterSpreadersSection is called when entering the spreadersSection production.
	EnterSpreadersSection(c *SpreadersSectionContext)

	// EnterSpreaderDeclaration is called when entering the spreaderDeclaration production.
	EnterSpreaderDeclaration(c *SpreaderDeclarationContext)

	// ExitDocument is called when exiting the document production.
	ExitDocument(c *DocumentContext)

	// ExitStatement is called when exiting the statement production.
	ExitStatement(c *StatementContext)

	// ExitIdentifier is called when exiting the identifier production.
	ExitIdentifier(c *IdentifierContext)

	// ExitValue is called when exiting the value production.
	ExitValue(c *ValueContext)

	// ExitTypeReference is called when exiting the typeReference production.
	ExitTypeReference(c *TypeReferenceContext)

	// ExitParameterList is called when exiting the parameterList production.
	ExitParameterList(c *ParameterListContext)

	// ExitParameter is called when exiting the parameter production.
	ExitParameter(c *ParameterContext)

	// ExitParameterValue is called when exiting the parameterValue production.
	ExitParameterValue(c *ParameterValueContext)

	// ExitIndent is called when exiting the indent production.
	ExitIndent(c *IndentContext)

	// ExitIdentifierList is called when exiting the identifierList production.
	ExitIdentifierList(c *IdentifierListContext)

	// ExitExpression is called when exiting the expression production.
	ExitExpression(c *ExpressionContext)

	// ExitConditionalExpr is called when exiting the conditionalExpr production.
	ExitConditionalExpr(c *ConditionalExprContext)

	// ExitIfBranch is called when exiting the ifBranch production.
	ExitIfBranch(c *IfBranchContext)

	// ExitElseIfBranch is called when exiting the elseIfBranch production.
	ExitElseIfBranch(c *ElseIfBranchContext)

	// ExitElseBranch is called when exiting the elseBranch production.
	ExitElseBranch(c *ElseBranchContext)

	// ExitWhenExpr is called when exiting the whenExpr production.
	ExitWhenExpr(c *WhenExprContext)

	// ExitWhenClause is called when exiting the whenClause production.
	ExitWhenClause(c *WhenClauseContext)

	// ExitDefaultClause is called when exiting the defaultClause production.
	ExitDefaultClause(c *DefaultClauseContext)

	// ExitLogicalOrExpr is called when exiting the logicalOrExpr production.
	ExitLogicalOrExpr(c *LogicalOrExprContext)

	// ExitLogicalAndExpr is called when exiting the logicalAndExpr production.
	ExitLogicalAndExpr(c *LogicalAndExprContext)

	// ExitNotExpr is called when exiting the notExpr production.
	ExitNotExpr(c *NotExprContext)

	// ExitEqualityExpr is called when exiting the equalityExpr production.
	ExitEqualityExpr(c *EqualityExprContext)

	// ExitRelationalExpr is called when exiting the relationalExpr production.
	ExitRelationalExpr(c *RelationalExprContext)

	// ExitAdditionExpr is called when exiting the additionExpr production.
	ExitAdditionExpr(c *AdditionExprContext)

	// ExitMultiplicationExpr is called when exiting the multiplicationExpr production.
	ExitMultiplicationExpr(c *MultiplicationExprContext)

	// ExitUnaryExpr is called when exiting the unaryExpr production.
	ExitUnaryExpr(c *UnaryExprContext)

	// ExitPrimaryExpr is called when exiting the primaryExpr production.
	ExitPrimaryExpr(c *PrimaryExprContext)

	// ExitLiteral is called when exiting the literal production.
	ExitLiteral(c *LiteralContext)

	// ExitConstraintList is called when exiting the constraintList production.
	ExitConstraintList(c *ConstraintListContext)

	// ExitConstraint is called when exiting the constraint production.
	ExitConstraint(c *ConstraintContext)

	// ExitConstraintValue is called when exiting the constraintValue production.
	ExitConstraintValue(c *ConstraintValueContext)

	// ExitNestedConstraintValue is called when exiting the nestedConstraintValue production.
	ExitNestedConstraintValue(c *NestedConstraintValueContext)

	// ExitArithmeticExpr is called when exiting the arithmeticExpr production.
	ExitArithmeticExpr(c *ArithmeticExprContext)

	// ExitArithmeticTerm is called when exiting the arithmeticTerm production.
	ExitArithmeticTerm(c *ArithmeticTermContext)

	// ExitArithmeticFactor is called when exiting the arithmeticFactor production.
	ExitArithmeticFactor(c *ArithmeticFactorContext)

	// ExitMultOperator is called when exiting the multOperator production.
	ExitMultOperator(c *MultOperatorContext)

	// ExitDivOperator is called when exiting the divOperator production.
	ExitDivOperator(c *DivOperatorContext)

	// ExitPowOperator is called when exiting the powOperator production.
	ExitPowOperator(c *PowOperatorContext)

	// ExitPlusOperator is called when exiting the plusOperator production.
	ExitPlusOperator(c *PlusOperatorContext)

	// ExitMinusOperator is called when exiting the minusOperator production.
	ExitMinusOperator(c *MinusOperatorContext)

	// ExitConceptsSection is called when exiting the conceptsSection production.
	ExitConceptsSection(c *ConceptsSectionContext)

	// ExitConceptDefinition is called when exiting the conceptDefinition production.
	ExitConceptDefinition(c *ConceptDefinitionContext)

	// ExitConceptLine is called when exiting the conceptLine production.
	ExitConceptLine(c *ConceptLineContext)

	// ExitConditionalSpec is called when exiting the conditionalSpec production.
	ExitConditionalSpec(c *ConditionalSpecContext)

	// ExitConceptExtension is called when exiting the conceptExtension production.
	ExitConceptExtension(c *ConceptExtensionContext)

	// ExitInsertionPoint is called when exiting the insertionPoint production.
	ExitInsertionPoint(c *InsertionPointContext)

	// ExitGeneratorPatternList is called when exiting the generatorPatternList production.
	ExitGeneratorPatternList(c *GeneratorPatternListContext)

	// ExitGeneratorPattern is called when exiting the generatorPattern production.
	ExitGeneratorPattern(c *GeneratorPatternContext)

	// ExitValuesPatternGenerator is called when exiting the valuesPatternGenerator production.
	ExitValuesPatternGenerator(c *ValuesPatternGeneratorContext)

	// ExitIntRangeGenerator is called when exiting the intRangeGenerator production.
	ExitIntRangeGenerator(c *IntRangeGeneratorContext)

	// ExitIntRangeParams is called when exiting the intRangeParams production.
	ExitIntRangeParams(c *IntRangeParamsContext)

	// ExitConceptPatternGenerator is called when exiting the conceptPatternGenerator production.
	ExitConceptPatternGenerator(c *ConceptPatternGeneratorContext)

	// ExitConceptRange is called when exiting the conceptRange production.
	ExitConceptRange(c *ConceptRangeContext)

	// ExitConceptValueList is called when exiting the conceptValueList production.
	ExitConceptValueList(c *ConceptValueListContext)

	// ExitMeasuresSection is called when exiting the measuresSection production.
	ExitMeasuresSection(c *MeasuresSectionContext)

	// ExitMeasureDeclaration is called when exiting the measureDeclaration production.
	ExitMeasureDeclaration(c *MeasureDeclarationContext)

	// ExitMeasureParameterList is called when exiting the measureParameterList production.
	ExitMeasureParameterList(c *MeasureParameterListContext)

	// ExitMeasureParameter is called when exiting the measureParameter production.
	ExitMeasureParameter(c *MeasureParameterContext)

	// ExitUnitParameter is called when exiting the unitParameter production.
	ExitUnitParameter(c *UnitParameterContext)

	// ExitMeasureConversionClause is called when exiting the measureConversionClause production.
	ExitMeasureConversionClause(c *MeasureConversionClauseContext)

	// ExitMeasureConversionPair is called when exiting the measureConversionPair production.
	ExitMeasureConversionPair(c *MeasureConversionPairContext)

	// ExitMeasureCalculationClause is called when exiting the measureCalculationClause production.
	ExitMeasureCalculationClause(c *MeasureCalculationClauseContext)

	// ExitTypesSection is called when exiting the typesSection production.
	ExitTypesSection(c *TypesSectionContext)

	// ExitTypeDefinition is called when exiting the typeDefinition production.
	ExitTypeDefinition(c *TypeDefinitionContext)

	// ExitBaseType is called when exiting the baseType production.
	ExitBaseType(c *BaseTypeContext)

	// ExitUnitsSection is called when exiting the unitsSection production.
	ExitUnitsSection(c *UnitsSectionContext)

	// ExitUnitDeclaration is called when exiting the unitDeclaration production.
	ExitUnitDeclaration(c *UnitDeclarationContext)

	// ExitConversionClause is called when exiting the conversionClause production.
	ExitConversionClause(c *ConversionClauseContext)

	// ExitConversionPair is called when exiting the conversionPair production.
	ExitConversionPair(c *ConversionPairContext)

	// ExitPatternClause is called when exiting the patternClause production.
	ExitPatternClause(c *PatternClauseContext)

	// ExitPattern is called when exiting the pattern production.
	ExitPattern(c *PatternContext)

	// ExitPatternBoundVariable is called when exiting the patternBoundVariable production.
	ExitPatternBoundVariable(c *PatternBoundVariableContext)

	// ExitCalculationClause is called when exiting the calculationClause production.
	ExitCalculationClause(c *CalculationClauseContext)

	// ExitConnectorsSection is called when exiting the connectorsSection production.
	ExitConnectorsSection(c *ConnectorsSectionContext)

	// ExitConnectorDeclaration is called when exiting the connectorDeclaration production.
	ExitConnectorDeclaration(c *ConnectorDeclarationContext)

	// ExitMetricsSection is called when exiting the metricsSection production.
	ExitMetricsSection(c *MetricsSectionContext)

	// ExitShortFormDeclaration is called when exiting the shortFormDeclaration production.
	ExitShortFormDeclaration(c *ShortFormDeclarationContext)

	// ExitMetricDeclaration is called when exiting the metricDeclaration production.
	ExitMetricDeclaration(c *MetricDeclarationContext)

	// ExitInputsBlock is called when exiting the inputsBlock production.
	ExitInputsBlock(c *InputsBlockContext)

	// ExitInputDeclaration is called when exiting the inputDeclaration production.
	ExitInputDeclaration(c *InputDeclarationContext)

	// ExitInputKind is called when exiting the inputKind production.
	ExitInputKind(c *InputKindContext)

	// ExitInputAsClause is called when exiting the inputAsClause production.
	ExitInputAsClause(c *InputAsClauseContext)

	// ExitWhereExpression is called when exiting the whereExpression production.
	ExitWhereExpression(c *WhereExpressionContext)

	// ExitOrExpression is called when exiting the orExpression production.
	ExitOrExpression(c *OrExpressionContext)

	// ExitAndExpression is called when exiting the andExpression production.
	ExitAndExpression(c *AndExpressionContext)

	// ExitAspectExpression is called when exiting the aspectExpression production.
	ExitAspectExpression(c *AspectExpressionContext)

	// ExitBackwardClause is called when exiting the backwardClause production.
	ExitBackwardClause(c *BackwardClauseContext)

	// ExitForwardClause is called when exiting the forwardClause production.
	ExitForwardClause(c *ForwardClauseContext)

	// ExitShiftExpression is called when exiting the shiftExpression production.
	ExitShiftExpression(c *ShiftExpressionContext)

	// ExitCalculateClause is called when exiting the calculateClause production.
	ExitCalculateClause(c *CalculateClauseContext)

	// ExitLookupsSection is called when exiting the lookupsSection production.
	ExitLookupsSection(c *LookupsSectionContext)

	// ExitLookupDeclaration is called when exiting the lookupDeclaration production.
	ExitLookupDeclaration(c *LookupDeclarationContext)

	// ExitDimensionsClause is called when exiting the dimensionsClause production.
	ExitDimensionsClause(c *DimensionsClauseContext)

	// ExitCelltypeClause is called when exiting the celltypeClause production.
	ExitCelltypeClause(c *CelltypeClauseContext)

	// ExitAggregatorsSection is called when exiting the aggregatorsSection production.
	ExitAggregatorsSection(c *AggregatorsSectionContext)

	// ExitAggregatorDeclaration is called when exiting the aggregatorDeclaration production.
	ExitAggregatorDeclaration(c *AggregatorDeclarationContext)

	// ExitToTarget is called when exiting the toTarget production.
	ExitToTarget(c *ToTargetContext)

	// ExitFnCall is called when exiting the fnCall production.
	ExitFnCall(c *FnCallContext)

	// ExitSpreadersSection is called when exiting the spreadersSection production.
	ExitSpreadersSection(c *SpreadersSectionContext)

	// ExitSpreaderDeclaration is called when exiting the spreaderDeclaration production.
	ExitSpreaderDeclaration(c *SpreaderDeclarationContext)
}
