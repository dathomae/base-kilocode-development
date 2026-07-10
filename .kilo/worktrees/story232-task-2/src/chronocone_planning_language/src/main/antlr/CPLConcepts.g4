// CPLConcepts.g4 Parser grammar for the Chronocone Planning Language Concepts section. This file
// contains parser rules for concept definitions using the Values generator pattern.
// 
// Example CPL concepts section: concepts Country=[Values[USA, Canada, Mexico]]
// Priority=[Values[Low, Medium, High]]
// 
// Grammar supports: - Concept definitions with Values generators - Multiple values per concept -
// Multiple concepts per section

parser grammar CPLConcepts;

options {
	tokenVocab = ChronoconePlanningLanguageLexer;
}

// ============================================================================ CONCEPTS SECTION
// ============================================================================ The concepts section
// declares named concepts with value generators. Each concept defines a set of possible values that
// can be used in planning and metrics.

// A concepts section contains one or more concept definitions Example: concepts
// Country=[Values[USA, Canada]] Priority=[Values[Low, Medium, High]]
conceptsSection: CONCEPTS conceptDefinition+;

// A concept definition associates a name with a list of generator patterns Example:
// Country=[Values[USA, Canada, Mexico]]
conceptDefinition:
	indent (
		conceptLine
		| conceptExtension
		| conditionalSpec
	)*;

// Base concept line: ConceptName=[Values[...]] or ConceptName=[Int[1,10]]
conceptLine
	: identifier EQUALS LBRACKET generatorPatternList RBRACKET
	;

// Conditional specification: Parent[Value] => Child=[Values[...]]
// The condition (left of ARROW) specifies a parent concept and value.
// The result (right of ARROW) is a conceptLine for the child concept.
conditionalSpec
	: identifier LBRACKET value RBRACKET ARROW conceptLine
	;

// A concept extension adds new values to an existing concept with an insertion point.
// Example: Priority=[Values[Critical]] before Low
conceptExtension:
	identifier EQUALS LBRACKET generatorPatternList RBRACKET insertionPoint;

// Insertion point for concept extensions: specifies where new values are placed
// relative to an existing value. Example: before Low  or  after High
insertionPoint:
	BEFORE identifier
	| AFTER identifier;

// One or more generator patterns, separated by commas In Increment 1, only Values generators are
// supported Example: Values[USA, Canada] or Values[A, B], Values[C, D]
generatorPatternList:
	generatorPattern (COMMA generatorPattern)*;

// Generator pattern - supports Values generator, Int range generator, and Concept range generator
generatorPattern: valuesPatternGenerator | intRangeGenerator | conceptPatternGenerator;

// Values generator: produces an explicit list of concept values Example: Values[USA, Canada,
// Mexico]
valuesPatternGenerator:
	VALUES LBRACKET conceptValueList RBRACKET;

// Int range generator: produces a range of integer values
// Example: Int[1, 10] or Int[1, 10, 2]
intRangeGenerator:
	INT LBRACKET intRangeParams RBRACKET;

// Parameters for Int range generator: Start, End, optional Step
// Examples: 1, 10 or 1, 10, 2
intRangeParams:
	(NUMBER | INTEGER) COMMA (NUMBER | INTEGER) (COMMA (NUMBER | INTEGER))?;

// Concept range generator: produces a range of values from another concept
// Example: Quarter[Q1..Q4] or Month[Jan..Jun]
conceptPatternGenerator:
	identifier LBRACKET conceptRange RBRACKET;

// Concept range: StartValue..EndValue (references values in another concept)
conceptRange:
	identifier DOTDOT identifier;

// Comma-separated list of concept values (values) Example: USA, Canada, Mexico or Low, Medium,
// High
conceptValueList: value (COMMA value)*;