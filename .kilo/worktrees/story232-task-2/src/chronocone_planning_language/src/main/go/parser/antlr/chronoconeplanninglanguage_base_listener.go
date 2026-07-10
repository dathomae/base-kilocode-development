// Code generated from ChronoconePlanningLanguage.g4 by ANTLR 4.13.1. DO NOT EDIT.

package antlr // ChronoconePlanningLanguage
import "github.com/antlr4-go/antlr/v4"

// BaseChronoconePlanningLanguageListener is a complete listener for a parse tree produced by ChronoconePlanningLanguageParser.
type BaseChronoconePlanningLanguageListener struct{}

var _ ChronoconePlanningLanguageListener = &BaseChronoconePlanningLanguageListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseChronoconePlanningLanguageListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseChronoconePlanningLanguageListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterDocument is called when production document is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterDocument(ctx *DocumentContext) {}

// ExitDocument is called when production document is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitDocument(ctx *DocumentContext) {}

// EnterStatement is called when production statement is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterStatement(ctx *StatementContext) {}

// ExitStatement is called when production statement is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitStatement(ctx *StatementContext) {}

// EnterIdentifier is called when production identifier is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterIdentifier(ctx *IdentifierContext) {}

// ExitIdentifier is called when production identifier is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitIdentifier(ctx *IdentifierContext) {}

// EnterValue is called when production value is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterValue(ctx *ValueContext) {}

// ExitValue is called when production value is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitValue(ctx *ValueContext) {}

// EnterTypeReference is called when production typeReference is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterTypeReference(ctx *TypeReferenceContext) {}

// ExitTypeReference is called when production typeReference is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitTypeReference(ctx *TypeReferenceContext) {}

// EnterParameterList is called when production parameterList is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterParameterList(ctx *ParameterListContext) {}

// ExitParameterList is called when production parameterList is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitParameterList(ctx *ParameterListContext) {}

// EnterParameter is called when production parameter is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterParameter(ctx *ParameterContext) {}

// ExitParameter is called when production parameter is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitParameter(ctx *ParameterContext) {}

// EnterParameterValue is called when production parameterValue is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterParameterValue(ctx *ParameterValueContext) {}

// ExitParameterValue is called when production parameterValue is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitParameterValue(ctx *ParameterValueContext) {}

// EnterIndent is called when production indent is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterIndent(ctx *IndentContext) {}

// ExitIndent is called when production indent is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitIndent(ctx *IndentContext) {}

// EnterIdentifierList is called when production identifierList is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterIdentifierList(ctx *IdentifierListContext) {}

// ExitIdentifierList is called when production identifierList is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitIdentifierList(ctx *IdentifierListContext) {}

// EnterExpression is called when production expression is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterExpression(ctx *ExpressionContext) {}

// ExitExpression is called when production expression is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitExpression(ctx *ExpressionContext) {}

// EnterConditionalExpr is called when production conditionalExpr is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterConditionalExpr(ctx *ConditionalExprContext) {}

// ExitConditionalExpr is called when production conditionalExpr is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitConditionalExpr(ctx *ConditionalExprContext) {}

// EnterIfBranch is called when production ifBranch is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterIfBranch(ctx *IfBranchContext) {}

// ExitIfBranch is called when production ifBranch is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitIfBranch(ctx *IfBranchContext) {}

// EnterElseIfBranch is called when production elseIfBranch is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterElseIfBranch(ctx *ElseIfBranchContext) {}

// ExitElseIfBranch is called when production elseIfBranch is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitElseIfBranch(ctx *ElseIfBranchContext) {}

// EnterElseBranch is called when production elseBranch is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterElseBranch(ctx *ElseBranchContext) {}

// ExitElseBranch is called when production elseBranch is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitElseBranch(ctx *ElseBranchContext) {}

// EnterWhenExpr is called when production whenExpr is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterWhenExpr(ctx *WhenExprContext) {}

// ExitWhenExpr is called when production whenExpr is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitWhenExpr(ctx *WhenExprContext) {}

// EnterWhenClause is called when production whenClause is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterWhenClause(ctx *WhenClauseContext) {}

// ExitWhenClause is called when production whenClause is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitWhenClause(ctx *WhenClauseContext) {}

// EnterDefaultClause is called when production defaultClause is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterDefaultClause(ctx *DefaultClauseContext) {}

// ExitDefaultClause is called when production defaultClause is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitDefaultClause(ctx *DefaultClauseContext) {}

// EnterLogicalOrExpr is called when production logicalOrExpr is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterLogicalOrExpr(ctx *LogicalOrExprContext) {}

// ExitLogicalOrExpr is called when production logicalOrExpr is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitLogicalOrExpr(ctx *LogicalOrExprContext) {}

// EnterLogicalAndExpr is called when production logicalAndExpr is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterLogicalAndExpr(ctx *LogicalAndExprContext) {}

// ExitLogicalAndExpr is called when production logicalAndExpr is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitLogicalAndExpr(ctx *LogicalAndExprContext) {}

// EnterNotExpr is called when production notExpr is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterNotExpr(ctx *NotExprContext) {}

// ExitNotExpr is called when production notExpr is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitNotExpr(ctx *NotExprContext) {}

// EnterEqualityExpr is called when production equalityExpr is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterEqualityExpr(ctx *EqualityExprContext) {}

// ExitEqualityExpr is called when production equalityExpr is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitEqualityExpr(ctx *EqualityExprContext) {}

// EnterRelationalExpr is called when production relationalExpr is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterRelationalExpr(ctx *RelationalExprContext) {}

// ExitRelationalExpr is called when production relationalExpr is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitRelationalExpr(ctx *RelationalExprContext) {}

// EnterAdditionExpr is called when production additionExpr is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterAdditionExpr(ctx *AdditionExprContext) {}

// ExitAdditionExpr is called when production additionExpr is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitAdditionExpr(ctx *AdditionExprContext) {}

// EnterMultiplicationExpr is called when production multiplicationExpr is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterMultiplicationExpr(ctx *MultiplicationExprContext) {
}

// ExitMultiplicationExpr is called when production multiplicationExpr is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitMultiplicationExpr(ctx *MultiplicationExprContext) {
}

// EnterUnaryExpr is called when production unaryExpr is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterUnaryExpr(ctx *UnaryExprContext) {}

// ExitUnaryExpr is called when production unaryExpr is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitUnaryExpr(ctx *UnaryExprContext) {}

// EnterPrimaryExpr is called when production primaryExpr is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterPrimaryExpr(ctx *PrimaryExprContext) {}

// ExitPrimaryExpr is called when production primaryExpr is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitPrimaryExpr(ctx *PrimaryExprContext) {}

// EnterLiteral is called when production literal is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterLiteral(ctx *LiteralContext) {}

// ExitLiteral is called when production literal is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitLiteral(ctx *LiteralContext) {}

// EnterConstraintList is called when production constraintList is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterConstraintList(ctx *ConstraintListContext) {}

// ExitConstraintList is called when production constraintList is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitConstraintList(ctx *ConstraintListContext) {}

// EnterConstraint is called when production constraint is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterConstraint(ctx *ConstraintContext) {}

// ExitConstraint is called when production constraint is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitConstraint(ctx *ConstraintContext) {}

// EnterConstraintValue is called when production constraintValue is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterConstraintValue(ctx *ConstraintValueContext) {}

// ExitConstraintValue is called when production constraintValue is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitConstraintValue(ctx *ConstraintValueContext) {}

// EnterNestedConstraintValue is called when production nestedConstraintValue is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterNestedConstraintValue(ctx *NestedConstraintValueContext) {
}

// ExitNestedConstraintValue is called when production nestedConstraintValue is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitNestedConstraintValue(ctx *NestedConstraintValueContext) {
}

// EnterArithmeticExpr is called when production arithmeticExpr is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterArithmeticExpr(ctx *ArithmeticExprContext) {}

// ExitArithmeticExpr is called when production arithmeticExpr is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitArithmeticExpr(ctx *ArithmeticExprContext) {}

// EnterArithmeticTerm is called when production arithmeticTerm is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterArithmeticTerm(ctx *ArithmeticTermContext) {}

// ExitArithmeticTerm is called when production arithmeticTerm is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitArithmeticTerm(ctx *ArithmeticTermContext) {}

// EnterArithmeticFactor is called when production arithmeticFactor is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterArithmeticFactor(ctx *ArithmeticFactorContext) {
}

// ExitArithmeticFactor is called when production arithmeticFactor is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitArithmeticFactor(ctx *ArithmeticFactorContext) {}

// EnterMultOperator is called when production multOperator is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterMultOperator(ctx *MultOperatorContext) {}

// ExitMultOperator is called when production multOperator is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitMultOperator(ctx *MultOperatorContext) {}

// EnterDivOperator is called when production divOperator is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterDivOperator(ctx *DivOperatorContext) {}

// ExitDivOperator is called when production divOperator is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitDivOperator(ctx *DivOperatorContext) {}

// EnterPowOperator is called when production powOperator is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterPowOperator(ctx *PowOperatorContext) {}

// ExitPowOperator is called when production powOperator is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitPowOperator(ctx *PowOperatorContext) {}

// EnterPlusOperator is called when production plusOperator is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterPlusOperator(ctx *PlusOperatorContext) {}

// ExitPlusOperator is called when production plusOperator is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitPlusOperator(ctx *PlusOperatorContext) {}

// EnterMinusOperator is called when production minusOperator is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterMinusOperator(ctx *MinusOperatorContext) {}

// ExitMinusOperator is called when production minusOperator is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitMinusOperator(ctx *MinusOperatorContext) {}

// EnterConceptsSection is called when production conceptsSection is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterConceptsSection(ctx *ConceptsSectionContext) {}

// ExitConceptsSection is called when production conceptsSection is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitConceptsSection(ctx *ConceptsSectionContext) {}

// EnterConceptDefinition is called when production conceptDefinition is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterConceptDefinition(ctx *ConceptDefinitionContext) {
}

// ExitConceptDefinition is called when production conceptDefinition is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitConceptDefinition(ctx *ConceptDefinitionContext) {
}

// EnterConceptLine is called when production conceptLine is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterConceptLine(ctx *ConceptLineContext) {}

// ExitConceptLine is called when production conceptLine is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitConceptLine(ctx *ConceptLineContext) {}

// EnterConditionalSpec is called when production conditionalSpec is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterConditionalSpec(ctx *ConditionalSpecContext) {}

// ExitConditionalSpec is called when production conditionalSpec is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitConditionalSpec(ctx *ConditionalSpecContext) {}

// EnterConceptExtension is called when production conceptExtension is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterConceptExtension(ctx *ConceptExtensionContext) {
}

// ExitConceptExtension is called when production conceptExtension is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitConceptExtension(ctx *ConceptExtensionContext) {}

// EnterInsertionPoint is called when production insertionPoint is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterInsertionPoint(ctx *InsertionPointContext) {}

// ExitInsertionPoint is called when production insertionPoint is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitInsertionPoint(ctx *InsertionPointContext) {}

// EnterGeneratorPatternList is called when production generatorPatternList is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterGeneratorPatternList(ctx *GeneratorPatternListContext) {
}

// ExitGeneratorPatternList is called when production generatorPatternList is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitGeneratorPatternList(ctx *GeneratorPatternListContext) {
}

// EnterGeneratorPattern is called when production generatorPattern is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterGeneratorPattern(ctx *GeneratorPatternContext) {
}

// ExitGeneratorPattern is called when production generatorPattern is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitGeneratorPattern(ctx *GeneratorPatternContext) {}

// EnterValuesPatternGenerator is called when production valuesPatternGenerator is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterValuesPatternGenerator(ctx *ValuesPatternGeneratorContext) {
}

// ExitValuesPatternGenerator is called when production valuesPatternGenerator is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitValuesPatternGenerator(ctx *ValuesPatternGeneratorContext) {
}

// EnterIntRangeGenerator is called when production intRangeGenerator is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterIntRangeGenerator(ctx *IntRangeGeneratorContext) {
}

// ExitIntRangeGenerator is called when production intRangeGenerator is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitIntRangeGenerator(ctx *IntRangeGeneratorContext) {
}

// EnterIntRangeParams is called when production intRangeParams is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterIntRangeParams(ctx *IntRangeParamsContext) {}

// ExitIntRangeParams is called when production intRangeParams is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitIntRangeParams(ctx *IntRangeParamsContext) {}

// EnterConceptPatternGenerator is called when production conceptPatternGenerator is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterConceptPatternGenerator(ctx *ConceptPatternGeneratorContext) {
}

// ExitConceptPatternGenerator is called when production conceptPatternGenerator is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitConceptPatternGenerator(ctx *ConceptPatternGeneratorContext) {
}

// EnterConceptRange is called when production conceptRange is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterConceptRange(ctx *ConceptRangeContext) {}

// ExitConceptRange is called when production conceptRange is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitConceptRange(ctx *ConceptRangeContext) {}

// EnterConceptValueList is called when production conceptValueList is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterConceptValueList(ctx *ConceptValueListContext) {
}

// ExitConceptValueList is called when production conceptValueList is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitConceptValueList(ctx *ConceptValueListContext) {}

// EnterMeasuresSection is called when production measuresSection is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterMeasuresSection(ctx *MeasuresSectionContext) {}

// ExitMeasuresSection is called when production measuresSection is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitMeasuresSection(ctx *MeasuresSectionContext) {}

// EnterMeasureDeclaration is called when production measureDeclaration is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterMeasureDeclaration(ctx *MeasureDeclarationContext) {
}

// ExitMeasureDeclaration is called when production measureDeclaration is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitMeasureDeclaration(ctx *MeasureDeclarationContext) {
}

// EnterMeasureParameterList is called when production measureParameterList is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterMeasureParameterList(ctx *MeasureParameterListContext) {
}

// ExitMeasureParameterList is called when production measureParameterList is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitMeasureParameterList(ctx *MeasureParameterListContext) {
}

// EnterMeasureParameter is called when production measureParameter is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterMeasureParameter(ctx *MeasureParameterContext) {
}

// ExitMeasureParameter is called when production measureParameter is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitMeasureParameter(ctx *MeasureParameterContext) {}

// EnterUnitParameter is called when production unitParameter is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterUnitParameter(ctx *UnitParameterContext) {}

// ExitUnitParameter is called when production unitParameter is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitUnitParameter(ctx *UnitParameterContext) {}

// EnterMeasureConversionClause is called when production measureConversionClause is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterMeasureConversionClause(ctx *MeasureConversionClauseContext) {
}

// ExitMeasureConversionClause is called when production measureConversionClause is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitMeasureConversionClause(ctx *MeasureConversionClauseContext) {
}

// EnterMeasureConversionPair is called when production measureConversionPair is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterMeasureConversionPair(ctx *MeasureConversionPairContext) {
}

// ExitMeasureConversionPair is called when production measureConversionPair is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitMeasureConversionPair(ctx *MeasureConversionPairContext) {
}

// EnterMeasureCalculationClause is called when production measureCalculationClause is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterMeasureCalculationClause(ctx *MeasureCalculationClauseContext) {
}

// ExitMeasureCalculationClause is called when production measureCalculationClause is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitMeasureCalculationClause(ctx *MeasureCalculationClauseContext) {
}

// EnterTypesSection is called when production typesSection is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterTypesSection(ctx *TypesSectionContext) {}

// ExitTypesSection is called when production typesSection is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitTypesSection(ctx *TypesSectionContext) {}

// EnterTypeDefinition is called when production typeDefinition is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterTypeDefinition(ctx *TypeDefinitionContext) {}

// ExitTypeDefinition is called when production typeDefinition is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitTypeDefinition(ctx *TypeDefinitionContext) {}

// EnterBaseType is called when production baseType is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterBaseType(ctx *BaseTypeContext) {}

// ExitBaseType is called when production baseType is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitBaseType(ctx *BaseTypeContext) {}

// EnterUnitsSection is called when production unitsSection is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterUnitsSection(ctx *UnitsSectionContext) {}

// ExitUnitsSection is called when production unitsSection is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitUnitsSection(ctx *UnitsSectionContext) {}

// EnterUnitDeclaration is called when production unitDeclaration is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterUnitDeclaration(ctx *UnitDeclarationContext) {}

// ExitUnitDeclaration is called when production unitDeclaration is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitUnitDeclaration(ctx *UnitDeclarationContext) {}

// EnterConversionClause is called when production conversionClause is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterConversionClause(ctx *ConversionClauseContext) {
}

// ExitConversionClause is called when production conversionClause is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitConversionClause(ctx *ConversionClauseContext) {}

// EnterConversionPair is called when production conversionPair is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterConversionPair(ctx *ConversionPairContext) {}

// ExitConversionPair is called when production conversionPair is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitConversionPair(ctx *ConversionPairContext) {}

// EnterPatternClause is called when production patternClause is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterPatternClause(ctx *PatternClauseContext) {}

// ExitPatternClause is called when production patternClause is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitPatternClause(ctx *PatternClauseContext) {}

// EnterPattern is called when production pattern is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterPattern(ctx *PatternContext) {}

// ExitPattern is called when production pattern is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitPattern(ctx *PatternContext) {}

// EnterPatternBoundVariable is called when production patternBoundVariable is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterPatternBoundVariable(ctx *PatternBoundVariableContext) {
}

// ExitPatternBoundVariable is called when production patternBoundVariable is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitPatternBoundVariable(ctx *PatternBoundVariableContext) {
}

// EnterCalculationClause is called when production calculationClause is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterCalculationClause(ctx *CalculationClauseContext) {
}

// ExitCalculationClause is called when production calculationClause is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitCalculationClause(ctx *CalculationClauseContext) {
}

// EnterConnectorsSection is called when production connectorsSection is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterConnectorsSection(ctx *ConnectorsSectionContext) {
}

// ExitConnectorsSection is called when production connectorsSection is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitConnectorsSection(ctx *ConnectorsSectionContext) {
}

// EnterConnectorDeclaration is called when production connectorDeclaration is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterConnectorDeclaration(ctx *ConnectorDeclarationContext) {
}

// ExitConnectorDeclaration is called when production connectorDeclaration is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitConnectorDeclaration(ctx *ConnectorDeclarationContext) {
}

// EnterMetricsSection is called when production metricsSection is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterMetricsSection(ctx *MetricsSectionContext) {}

// ExitMetricsSection is called when production metricsSection is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitMetricsSection(ctx *MetricsSectionContext) {}

// EnterShortFormDeclaration is called when production shortFormDeclaration is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterShortFormDeclaration(ctx *ShortFormDeclarationContext) {
}

// ExitShortFormDeclaration is called when production shortFormDeclaration is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitShortFormDeclaration(ctx *ShortFormDeclarationContext) {
}

// EnterMetricDeclaration is called when production metricDeclaration is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterMetricDeclaration(ctx *MetricDeclarationContext) {
}

// ExitMetricDeclaration is called when production metricDeclaration is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitMetricDeclaration(ctx *MetricDeclarationContext) {
}

// EnterInputsBlock is called when production inputsBlock is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterInputsBlock(ctx *InputsBlockContext) {}

// ExitInputsBlock is called when production inputsBlock is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitInputsBlock(ctx *InputsBlockContext) {}

// EnterInputDeclaration is called when production inputDeclaration is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterInputDeclaration(ctx *InputDeclarationContext) {
}

// ExitInputDeclaration is called when production inputDeclaration is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitInputDeclaration(ctx *InputDeclarationContext) {}

// EnterInputKind is called when production inputKind is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterInputKind(ctx *InputKindContext) {}

// ExitInputKind is called when production inputKind is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitInputKind(ctx *InputKindContext) {}

// EnterInputAsClause is called when production inputAsClause is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterInputAsClause(ctx *InputAsClauseContext) {}

// ExitInputAsClause is called when production inputAsClause is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitInputAsClause(ctx *InputAsClauseContext) {}

// EnterWhereExpression is called when production whereExpression is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterWhereExpression(ctx *WhereExpressionContext) {}

// ExitWhereExpression is called when production whereExpression is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitWhereExpression(ctx *WhereExpressionContext) {}

// EnterOrExpression is called when production orExpression is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterOrExpression(ctx *OrExpressionContext) {}

// ExitOrExpression is called when production orExpression is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitOrExpression(ctx *OrExpressionContext) {}

// EnterAndExpression is called when production andExpression is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterAndExpression(ctx *AndExpressionContext) {}

// ExitAndExpression is called when production andExpression is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitAndExpression(ctx *AndExpressionContext) {}

// EnterAspectExpression is called when production aspectExpression is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterAspectExpression(ctx *AspectExpressionContext) {
}

// ExitAspectExpression is called when production aspectExpression is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitAspectExpression(ctx *AspectExpressionContext) {}

// EnterBackwardClause is called when production backwardClause is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterBackwardClause(ctx *BackwardClauseContext) {}

// ExitBackwardClause is called when production backwardClause is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitBackwardClause(ctx *BackwardClauseContext) {}

// EnterForwardClause is called when production forwardClause is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterForwardClause(ctx *ForwardClauseContext) {}

// ExitForwardClause is called when production forwardClause is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitForwardClause(ctx *ForwardClauseContext) {}

// EnterShiftExpression is called when production shiftExpression is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterShiftExpression(ctx *ShiftExpressionContext) {}

// ExitShiftExpression is called when production shiftExpression is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitShiftExpression(ctx *ShiftExpressionContext) {}

// EnterCalculateClause is called when production calculateClause is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterCalculateClause(ctx *CalculateClauseContext) {}

// ExitCalculateClause is called when production calculateClause is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitCalculateClause(ctx *CalculateClauseContext) {}

// EnterLookupsSection is called when production lookupsSection is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterLookupsSection(ctx *LookupsSectionContext) {}

// ExitLookupsSection is called when production lookupsSection is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitLookupsSection(ctx *LookupsSectionContext) {}

// EnterLookupDeclaration is called when production lookupDeclaration is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterLookupDeclaration(ctx *LookupDeclarationContext) {
}

// ExitLookupDeclaration is called when production lookupDeclaration is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitLookupDeclaration(ctx *LookupDeclarationContext) {
}

// EnterDimensionsClause is called when production dimensionsClause is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterDimensionsClause(ctx *DimensionsClauseContext) {
}

// ExitDimensionsClause is called when production dimensionsClause is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitDimensionsClause(ctx *DimensionsClauseContext) {}

// EnterCelltypeClause is called when production celltypeClause is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterCelltypeClause(ctx *CelltypeClauseContext) {}

// ExitCelltypeClause is called when production celltypeClause is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitCelltypeClause(ctx *CelltypeClauseContext) {}

// EnterAggregatorsSection is called when production aggregatorsSection is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterAggregatorsSection(ctx *AggregatorsSectionContext) {
}

// ExitAggregatorsSection is called when production aggregatorsSection is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitAggregatorsSection(ctx *AggregatorsSectionContext) {
}

// EnterAggregatorDeclaration is called when production aggregatorDeclaration is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterAggregatorDeclaration(ctx *AggregatorDeclarationContext) {
}

// ExitAggregatorDeclaration is called when production aggregatorDeclaration is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitAggregatorDeclaration(ctx *AggregatorDeclarationContext) {
}

// EnterToTarget is called when production toTarget is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterToTarget(ctx *ToTargetContext) {}

// ExitToTarget is called when production toTarget is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitToTarget(ctx *ToTargetContext) {}

// EnterFnCall is called when production fnCall is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterFnCall(ctx *FnCallContext) {}

// ExitFnCall is called when production fnCall is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitFnCall(ctx *FnCallContext) {}

// EnterSpreadersSection is called when production spreadersSection is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterSpreadersSection(ctx *SpreadersSectionContext) {
}

// ExitSpreadersSection is called when production spreadersSection is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitSpreadersSection(ctx *SpreadersSectionContext) {}

// EnterSpreaderDeclaration is called when production spreaderDeclaration is entered.
func (s *BaseChronoconePlanningLanguageListener) EnterSpreaderDeclaration(ctx *SpreaderDeclarationContext) {
}

// ExitSpreaderDeclaration is called when production spreaderDeclaration is exited.
func (s *BaseChronoconePlanningLanguageListener) ExitSpreaderDeclaration(ctx *SpreaderDeclarationContext) {
}
