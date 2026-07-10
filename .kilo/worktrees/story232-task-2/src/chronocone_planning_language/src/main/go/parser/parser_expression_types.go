package parser

// ParserUnitReference captures a unit name reference in an arithmetic expression.
// This is pushed to the symbol table's Definitions slice by ExitIdentifier when
// parsing an arithmetic expression (e.g., "Length", "Time", "Mass").
type ParserUnitReference struct {
	Name string
}

func (ParserUnitReference) IsDefinitionData() {}

// OperatorType identifies an arithmetic operator in an expression.
type OperatorType int

const (
	OpMultiply OperatorType = iota
	OpDivide
	OpPower
	OpPlus
	OpMinus
)

// ParserOperator captures an arithmetic operator token in an expression.
// This is pushed to the symbol table's Definitions slice by the operator
// callback Exit methods (ExitMultOperator, ExitDivOperator, etc.) and by
// ExitArithmeticExpr, ExitArithmeticTerm, and ExitArithmeticFactor.
type ParserOperator struct {
	Operator OperatorType
}

func (ParserOperator) IsDefinitionData() {}

