// Package analysis provides the analysis phase for the Chronocone Planning Language.
//
// This file contains the ConceptAnalyzer which handles concept-related analysis
// operations including concept registration, merging, and range resolution.
package analysis

import (
	"fmt"
)

// ConceptAnalyzer handles concept-related analysis operations.
//
// It maintains the concepts map and provides methods for concept registration,
// extension merging, and concept range generator resolution.
type ConceptAnalyzer struct {
	// concepts is a map from concept name to its analyzed definition.
	// This map is built during analysis by processing concept definitions
	// from the symbol table. It provides O(1) access to concept information
	// for validation and code generation.
	concepts map[string]AnalysisConceptDefinition

	// lastConceptName tracks the most recently added concept for hierarchy
	// (parent/child relationship) detection during analysis.
	lastConceptName *string
}

// NewConceptAnalyzer creates a new ConceptAnalyzer with initialized data structures.
//
// Returns:
//   - A pointer to a new, initialized ConceptAnalyzer.
func NewConceptAnalyzer() *ConceptAnalyzer {
	return &ConceptAnalyzer{
		concepts:        make(map[string]AnalysisConceptDefinition),
		lastConceptName: nil,
	}
}

// AddConcept adds a concept definition to the analyzer's concepts map.
//
// It detects parent-child relationships based on indentation levels and
// returns an error if a duplicate concept definition is encountered.
func (ca *ConceptAnalyzer) AddConcept(concept *AnalysisConceptDefinition) error {
	if existingConcept, exists := ca.concepts[concept.ConceptName]; exists {
		return fmt.Errorf("duplicate concept definition: %s (first at line %d, column %d; second at line %d, column %d)",
			concept.ConceptName,
			existingConcept.SourceLocation.Line,
			existingConcept.SourceLocation.Column,
			concept.SourceLocation.Line,
			concept.SourceLocation.Column)
	}

	if ca.lastConceptName != nil {
		previousConcept := ca.concepts[*ca.lastConceptName]
		if previousConcept.IndentAsSpaces < concept.IndentAsSpaces {
			// This concept is nested inside the previous concept
			concept.ParentConceptName = &previousConcept.ConceptName
		}
	}

	ca.concepts[concept.ConceptName] = *concept
	ca.lastConceptName = &concept.ConceptName

	return nil
}

// AddOrMergeConcept adds a new concept or merges a concept extension.
//
// When a concept definition is encountered with a non-zero InsertionPoint
// (i.e., a concept extension), this method merges the extension's generators
// and values into the existing concept instead of treating it as a duplicate.
//
// Decision logic:
//   - If no existing concept with the same name: add as a new concept (same as AddConcept).
//     If the new concept has a non-zero InsertionPoint (extension with no base), error.
//   - If an existing concept exists AND the new concept has a non-zero InsertionPoint:
//     treat as an extension. Call mergeConceptValues to merge the extension into the
//     existing concept, write the merged result back to the concepts map.
//   - If an existing concept exists AND the new concept has a zero InsertionPoint:
//     return an error (true duplicate definition, not an extension).
//
// Parameters:
//   - concept: The concept definition to add or merge. If InsertionPoint is non-zero,
//     this is treated as an extension of an existing concept.
//
// Returns:
//   - error if the concept cannot be added or merged (e.g., true duplicate,
//     extension with no base concept, invalid insertion point, merge failure).
func (ca *ConceptAnalyzer) AddOrMergeConcept(concept *AnalysisConceptDefinition) error {
	existingConcept, exists := ca.concepts[concept.ConceptName]

	if !exists {
		// No existing concept with this name
		if concept.InsertionPoint.BeforeValue != "" || concept.InsertionPoint.AfterValue != "" {
			return fmt.Errorf("concept extension '%s' has no initial concept definition to extend (extension line %d, column %d)",
				concept.ConceptName,
				concept.SourceLocation.Line,
				concept.SourceLocation.Column)
		}

		// Add as new concept (same as AddConcept)
		if ca.lastConceptName != nil {
			previousConcept := ca.concepts[*ca.lastConceptName]
			if previousConcept.IndentAsSpaces < concept.IndentAsSpaces {
				concept.ParentConceptName = &previousConcept.ConceptName
			}
		}
		// Initialize usedInsertionAnchors for duplicate insertion detection
		if concept.usedInsertionAnchors == nil {
			concept.usedInsertionAnchors = make(map[string]bool)
		}
		ca.concepts[concept.ConceptName] = *concept
		ca.lastConceptName = &concept.ConceptName
		return nil
	}

	// Existing concept found
	if concept.InsertionPoint.BeforeValue == "" && concept.InsertionPoint.AfterValue == "" {
		// True duplicate - no insertion point, not an extension
		return fmt.Errorf("duplicate concept definition: %s (first at line %d, column %d; second at line %d, column %d)",
			concept.ConceptName,
			existingConcept.SourceLocation.Line,
			existingConcept.SourceLocation.Column,
			concept.SourceLocation.Line,
			concept.SourceLocation.Column)
	}

	// Extension with insertion point - merge into existing concept
	if err := mergeConceptValues(&existingConcept, concept); err != nil {
		return fmt.Errorf("failed to merge concept extension '%s': %w", concept.ConceptName, err)
	}
	ca.concepts[concept.ConceptName] = existingConcept
	return nil
}

// ResolveConceptRanges processes ConceptRangeGenerators in dependency order.
//
// It resolves ranges whose source concepts are fully resolved, deferring those
// whose sources still have pending ConceptRangeGenerators. Returns an error if
// circular dependencies are detected or if a referenced concept does not exist.
func (ca *ConceptAnalyzer) ResolveConceptRanges() error {
	type pendingRange struct {
		conceptName  string
		generator    *ConceptRangeGenerator
		generatorIdx int
	}
	var pending []pendingRange

	// Collect all ConceptRangeGenerators from all concepts
	for conceptName, concept := range ca.concepts {
		for idx, gen := range concept.Generators {
			if crg, ok := gen.(*ConceptRangeGenerator); ok {
				pending = append(pending, pendingRange{
					conceptName:  conceptName,
					generator:    crg,
					generatorIdx: idx,
				})
			}
		}
	}

	if len(pending) == 0 {
		return nil // No concept ranges to resolve
	}

	// Process in rounds until all resolved or circular dependency detected
	maxIterations := len(pending) + 1
	for iteration := 0; iteration < maxIterations && len(pending) > 0; iteration++ {
		var stillPending []pendingRange
		anyResolved := false

		for _, p := range pending {
			// Get the concept containing this generator
			concept := ca.concepts[p.conceptName]

			// Check if the referenced source concept exists
			_, exists := ca.concepts[p.generator.ConceptName]
			if !exists {
				return fmt.Errorf("concept '%s' references unknown concept '%s'",
					p.conceptName, p.generator.ConceptName)
			}

			// Check if the source concept has any unresolved ConceptRangeGenerators.
			// A self-reference is always considered pending because the source concept
			// (itself) has this very CRG that is not yet resolved.
			sourceHasPending := false
			if p.conceptName == p.generator.ConceptName {
				sourceHasPending = true
			} else {
				for _, stillP := range stillPending {
					if stillP.conceptName == p.generator.ConceptName {
						sourceHasPending = true
						break
					}
				}
				// Also check initial pending that will become stillPending
				if !sourceHasPending {
					for _, otherP := range pending {
						if otherP != p && otherP.conceptName == p.generator.ConceptName {
							sourceHasPending = true
							break
						}
					}
				}
			}

			if sourceHasPending {
				stillPending = append(stillPending, p)
				continue
			}

			// Source concept is fully resolved — resolve this range
			values, err := p.generator.GenerateValues(ca.concepts)
			if err != nil {
				return fmt.Errorf("failed to resolve concept range for '%s': %w", p.conceptName, err)
			}

			// Ensure index map is initialized
			if concept.index == nil {
				concept.index = make(map[string]int)
			}

			// Add the generated values to the concept
			for _, val := range values {
				if _, exists := concept.index[val]; exists {
					return fmt.Errorf("duplicate value '%s' in concept '%s' from concept range",
						val, p.conceptName)
				}
				pos := len(concept.ResolvedValues)
				concept.ResolvedValues = append(concept.ResolvedValues, AnalysisConceptValueDefinition{
					Name:     val,
					Position: pos,
				})
				concept.index[val] = pos
			}

			// Update the concept in the map (since it's a value type)
			ca.concepts[p.conceptName] = concept
			anyResolved = true
		}

		if !anyResolved && len(stillPending) > 0 {
			// No progress made — circular dependency
			var names []string
			for _, p := range stillPending {
				names = append(names, fmt.Sprintf("%s->%s", p.conceptName, p.generator.ConceptName))
			}
			return fmt.Errorf("circular dependency detected in concept ranges: %v", names)
		}

		pending = stillPending
	}

	if len(pending) > 0 {
		var names []string
		for _, p := range pending {
			names = append(names, fmt.Sprintf("%s->%s", p.conceptName, p.generator.ConceptName))
		}
		return fmt.Errorf("circular dependency detected in concept ranges: %v", names)
	}

	return nil
}

// AddNewConcept adds a concept definition directly to the concepts map.
// This is primarily useful for testing, allowing direct population of the concepts map.
func (ca *ConceptAnalyzer) AddNewConcept(concept *AnalysisConceptDefinition) {
	ca.concepts[concept.ConceptName] = *concept
}

// GetConcept returns the concept definition with the given name.
// Returns a zero-value AnalysisConceptDefinition if the concept is not found.
func (ca *ConceptAnalyzer) GetConcept(Name string) AnalysisConceptDefinition {
	return ca.concepts[Name]
}

// GetConcepts returns the concepts map.
func (ca *ConceptAnalyzer) GetConcepts() map[string]AnalysisConceptDefinition {
	return ca.concepts
}

// HasConcept returns true if a concept with the given name exists.
func (ca *ConceptAnalyzer) HasConcept(name string) bool {
	_, exists := ca.concepts[name]
	return exists
}

// SetConcept updates a concept definition in the concepts map.
func (ca *ConceptAnalyzer) SetConcept(name string, concept AnalysisConceptDefinition) {
	ca.concepts[name] = concept
}

// ResetLastConceptName clears the last concept name tracker.
//
// This is used when entering a new concepts scope to ensure that
// parent-child relationships do not incorrectly span across multiple
// concepts sections.
func (ca *ConceptAnalyzer) ResetLastConceptName() {
	ca.lastConceptName = nil
}
