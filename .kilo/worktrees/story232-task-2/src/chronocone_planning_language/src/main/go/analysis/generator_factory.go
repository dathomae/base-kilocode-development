// Package analysis provides the analysis phase data structures for the Chronocone Planning Language.
//
// This file contains the CreateGenerator factory function that converts
// parser-level generator definitions to analysis-level generators.
//
// Supports ValuesGeneratorDef, IntRangeGeneratorDef, and ConceptRangeGeneratorDef.
package analysis

import (
	"fmt"

	"github.com/chronocone/chronocone_planning_language/src/main/go/ast"
)

// CreateGenerator converts a parser generator definition to an analysis generator.
//
// This factory function bridges the parser and analysis phases by transforming
// parser-level generator definitions into analysis-level generator implementations.
// The analysis generators are responsible for producing the actual string values
// that make up a concept's value list.
//
// Supported generator types:
//   - ValuesGeneratorDef
//   - IntRangeGeneratorDef
//   - ConceptRangeGeneratorDef
//
// Parameters:
//   - parserGen: A parser generator definition (e.g., ast.ValuesGeneratorDef).
//     Uses interface{} type to allow for different generator definition types.
//
// Returns:
//   - A Generator implementation ready for use in analysis.
//   - An error if the generator type is unknown or conversion fails.
//
// Example:
//
//	parserGen := ast.ValuesGeneratorDef{Values: []string{"USA", "Canada"}}
//	gen, err := CreateGenerator(parserGen)
//	if err != nil {
//	    return err
//	}
//	values, err := gen.GenerateValues(nil) // Returns ["USA", "Canada"]
func CreateGenerator(parserGen interface{}) (Generator, error) {
	switch gen := parserGen.(type) {
	case ast.ValuesGeneratorDef:
		return &ValuesGenerator{Values: gen.Values}, nil
	case ast.IntRangeGeneratorDef:
		return &IntRangeGenerator{Start: gen.Start, End: gen.End, Step: gen.Step}, nil
	case ast.ConceptRangeGeneratorDef:
		return &ConceptRangeGenerator{ConceptName: gen.ConceptName, StartValue: gen.StartValue, EndValue: gen.EndValue}, nil
	default:
		return nil, fmt.Errorf("unknown generator type: %T", parserGen)
	}
}
