// Package analysis provides the analysis phase for the Chronocone Planning Language.
//
// This file defines AggregatorFunctionInfo and FunctionRegistry, which map
// aggregator function names to their CIL node types and dimensional unit effects.
package analysis

// AggregatorFunctionInfo defines a registered aggregator function.
type AggregatorFunctionInfo struct {
	Name        string
	CILNodeType string
	Arity       int
	DimUnits    func(params []string) *DimensionalUnitExpression
	Description string
}

// FunctionRegistry maps function names to their predefined definitions.
// The registry is populated during analysis initialization.
type FunctionRegistry struct {
	Functions map[string]*AggregatorFunctionInfo
}

// NewAggregatorFunctionRegistry creates a FunctionRegistry with the predefined
// aggregator functions. In Increment 2, only "sum" is registered.
func NewAggregatorFunctionRegistry() *FunctionRegistry {
	reg := &FunctionRegistry{
		Functions: make(map[string]*AggregatorFunctionInfo),
	}

	// Increment 2: only sum is supported
	reg.Functions["sum"] = &AggregatorFunctionInfo{
		Name:        "sum",
		CILNodeType: "CILAggregateNode",
		Arity:       1,
		DimUnits: func(params []string) *DimensionalUnitExpression {
			if len(params) != 1 {
				return nil
			}
			return &DimensionalUnitExpression{
				Numerator: []string{params[0]},
			}
		},
		Description: "Sums values across a concept dimension",
	}

	return reg
}
