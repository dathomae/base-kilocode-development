// CPLSpreaders.g4 - Grammar for CPL Spreaders section
//
// This file contains parser rules for spreader definitions.
//
// Example CPL spreaders section:
// spreaders
//     from [Demand] to [DemandA, DemandB] = split(Demand)
//
// Spreaders define how to expand values from source identifiers
// to multiple target identifiers using function calls.

parser grammar CPLSpreaders;

options {
	tokenVocab = ChronoconePlanningLanguageLexer;
}

// ============================================================================ SPREADERS SECTION
// ============================================================================

// A spreaders section contains zero or more spreader declarations
spreadersSection
	: SPREADERS spreaderDeclaration*
	;

// A spreader declaration: from [sources] to [targets] = expression
// Example: from [Demand] to [DemandA, DemandB] = split(Demand)
// Supports both function calls (e.g., split(Demand)) and arithmetic expressions (e.g., A + B).
// Both from and to clauses use identifierList? (both may be empty or contain identifiers).
spreaderDeclaration
	: indent FROM LBRACKET identifierList? RBRACKET TO LBRACKET identifierList? RBRACKET EQUALS (fnCall | expression)
	;
