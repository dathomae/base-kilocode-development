// CPLLookups.g4 - Grammar for CPL Lookups section
//
// This file contains parser rules for lookup definitions.
// Lookups define dimensional tables with a celltype measure.
//
// Example CPL lookups section:
//   lookups
//     SalesLookup
//         dimensions[Store]
//         celltype RevenueMeasure
//
// Grammar supports:
// - Lookup declarations with dimensions list and celltype
// - Zero or more lookup declarations within the lookups section

parser grammar CPLLookups;

options {
    tokenVocab = ChronoconePlanningLanguageLexer;
}

// ============================================================================ LOOKUPS SECTION
// ============================================================================
// The lookups section declares named lookups that define dimensional
// relationships. Each lookup specifies dimensions and a celltype.

// lookupsSection: LOOKUPS lookupDeclaration* — permits zero or more declarations
lookupsSection: LOOKUPS lookupDeclaration*;

// lookupDeclaration: indent identifier dimensionsClause celltypeClause — all clauses required
lookupDeclaration: indent identifier dimensionsClause celltypeClause;

// dimensionsClause: indent DIMENSIONS LBRACKET identifier (COMMA identifier)* RBRACKET — comma-separated concept names
dimensionsClause: indent DIMENSIONS LBRACKET identifier (COMMA identifier)* RBRACKET;

// celltypeClause: indent CELLTYPE identifier — a single measure name
celltypeClause: indent CELLTYPE identifier;
