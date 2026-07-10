package symboltable

// DefinitionData is the interface for type-specific data attached to SymbolDefinition.Data.
type DefinitionData interface {
	IsDefinitionData()
}

// StringData holds a single string value.
type StringData struct{ Value string }

func (StringData) IsDefinitionData() {}

// IntData holds an integer literal value.
type IntData struct{ Value int }

func (IntData) IsDefinitionData() {}

// FloatData holds a float literal value.
type FloatData struct{ Value float64 }

func (FloatData) IsDefinitionData() {}

// BoolData holds a boolean literal value.
type BoolData struct{ Value bool }

func (BoolData) IsDefinitionData() {}
