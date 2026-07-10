// Package analysis provides the analysis phase for the Chronocone Planning Language.
//
// This file contains the AnalysisContext stack used during symbol table walking
// and the AnalysisDefinition types it manages.
package analysis

import "errors"

// AnalysisDefinitionType identifies the kind of analysis definition on the context stack.
type AnalysisDefinitionType int

const (
	// conceptDefinition marks entries that hold an AnalysisConceptDefinition.
	conceptDefinition AnalysisDefinitionType = iota
)

// AnalysisDefinition is a heterogenous wrapper for definitions pushed onto the
// analysis context stack during the symbol table walk.
type AnalysisDefinition struct {
	DefinitionType AnalysisDefinitionType

	ConceptDefinition *AnalysisConceptDefinition
}

// NewAnalysisDefinition creates a new AnalysisDefinition with the given type.
func NewAnalysisDefinition(definitionType AnalysisDefinitionType) *AnalysisDefinition {
	return &AnalysisDefinition{
		DefinitionType: definitionType,
	}
}

// AnalysisContext is a stack of AnalysisDefinitions maintained during the
// analysis-phase symbol table walk.
type AnalysisContext struct {
	definitionStack []*AnalysisDefinition
}

// NewAnalysisContext creates a new empty analysis context stack.
func NewAnalysisContext() *AnalysisContext {
	return &AnalysisContext{
		definitionStack: make([]*AnalysisDefinition, 0),
	}
}

// IsEmpty returns true if the stack has no definitions.
func (ac *AnalysisContext) IsEmpty() bool {
	return len(ac.definitionStack) == 0
}

// Push adds a definition to the top of the stack.
func (ac *AnalysisContext) Push(ad *AnalysisDefinition) error {
	ac.definitionStack = append(ac.definitionStack, ad)
	return nil
}

// Pop removes and returns the top definition from the stack.
// Returns an error if the stack is empty.
func (ac *AnalysisContext) Pop() (*AnalysisDefinition, error) {
	if ac.IsEmpty() {
		return nil, errors.New("cannot pop from empty stack")
	}

	definition := ac.definitionStack[len(ac.definitionStack)-1]
	ac.definitionStack = ac.definitionStack[:len(ac.definitionStack)-1]
	return definition, nil
}

// Peek returns the top definition without removing it.
// Returns an error if the stack is empty.
func (ac *AnalysisContext) Peek() (*AnalysisDefinition, error) {
	if ac.IsEmpty() {
		return nil, errors.New("Attempt to peek at empty stack")
	}

	return ac.definitionStack[len(ac.definitionStack)-1], nil
}
