// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package cil

import "github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"

// CILLookupNode represents a reference to a lookup by name in the CIL expression tree.
// It implements the CILNode interface and serves as a leaf node
// in CIL expression trees, representing a pre-computed value map defined
// by a lookup declaration.
//
// Unlike CILConnectorNode, CILLookupNode carries its resolved dimension names
// and celltype name as additional CIL metadata, enabling code generation to
// produce lookup-specific code without referring back to the analysis-level
// LookupInfo.
//
// Example CIL tree usage:
//
//	// For lookup: CountryProductIndex = dimensions [Country, Product] celltype UnitsSold
//	node := &CILLookupNode{
//	    LookupName: "CountryProductIndex",
//	    Dimensions: []string{"Country", "Product"},
//	    Celltype:   "UnitsSold",
//	}
//	cilTrees["CountryProductIndex"] = node
type CILLookupNode struct {
	// LookupName is the identifier for this lookup in the CPL document.
	LookupName string

	// Dimensions is the slice of resolved concept names for the lookup's dimensions.
	// Each entry is the name of a concept that has been validated during analysis.
	// The order of dimensions matches the order declared in the CPL source.
	Dimensions []string

	// Celltype is the resolved measure name for the lookup's cell values.
	Celltype string

	// SourceLocation records where this lookup was defined in the source CPL.
	SourceLocation symboltable.SourceLocation
}

// cilNode is a marker method that satisfies the CILNode interface.
// It has no behavior and exists solely to allow CILLookupNode to be stored
// in CIL expression trees alongside other CIL node types.
func (n *CILLookupNode) cilNode() {}
