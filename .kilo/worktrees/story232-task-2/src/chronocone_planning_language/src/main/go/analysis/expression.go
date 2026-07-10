package analysis

// Expression is the interface for all analysis-time expression nodes.
type Expression interface {
	isExpression()
}

// UnitReference represents a resolved unit reference.
type UnitReference struct {
	Name      string
	UnitGroup *AnalysisUnitDefinition
}

func (UnitReference) isExpression() {}

// MultiplicationNode represents multiplication (e.g., Length * Time).
type MultiplicationNode struct {
	Left  Expression
	Right Expression
}

func (MultiplicationNode) isExpression() {}

// DivisionNode represents division (e.g., Length / Time).
type DivisionNode struct {
	Left  Expression
	Right Expression
}

func (DivisionNode) isExpression() {}

// PowerNode represents exponentiation (e.g., Meter^2).
type PowerNode struct {
	Base     Expression
	Exponent int
}

func (PowerNode) isExpression() {}

// AdditionNode represents addition.
type AdditionNode struct {
	Left  Expression
	Right Expression
}

func (AdditionNode) isExpression() {}

// SubtractionNode represents subtraction.
type SubtractionNode struct {
	Left  Expression
	Right Expression
}

func (SubtractionNode) isExpression() {}

// NumberNode represents a numeric literal.
type NumberNode struct {
	Value float64
}

func (NumberNode) isExpression() {}

// VariableRef represents a pattern-bound variable reference in a conversion
// calculation expression.
type VariableRef struct {
	Name string
}

func (VariableRef) isExpression() {}
