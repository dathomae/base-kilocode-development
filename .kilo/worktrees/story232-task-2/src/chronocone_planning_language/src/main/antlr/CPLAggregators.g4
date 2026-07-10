// CPLAggregators.g4 - Grammar for CPL Aggregators section
//
// This file contains parser rules for aggregator definitions.
//
// Example CPL aggregators section:
// aggregators
//     from [Demand] to [TotalDemand] = sum(Demand)
//
// Aggregators define how to roll up values from source identifiers
// to a target identifier using function calls.

parser grammar CPLAggregators;

options {
	tokenVocab = ChronoconePlanningLanguageLexer;
}

// ============================================================================ AGGREGATORS SECTION
// ============================================================================

// An aggregators section contains zero or more aggregator declarations
aggregatorsSection
	: AGGREGATORS aggregatorDeclaration*
	;

// An aggregator declaration: from [sources] to [target] = expression
// Example: from [Demand] to [TotalDemand] = sum(Demand)
// Supports both function calls (e.g., sum(Demand)) and arithmetic expressions (e.g., A + B).
aggregatorDeclaration
	: indent FROM LBRACKET identifierList RBRACKET TO LBRACKET toTarget RBRACKET EQUALS (fnCall | expression)
	;

// Target identifier for the aggregation (can be empty)
toTarget
	: identifier
	| /* empty */
	;

// Function call expression for aggregation
// Example: sum(Demand), count(Orders), size(Items)
fnCall
	: identifier LPAREN identifierList RPAREN
	;
