// CPLTypes.g4 Parser grammar for the Chronocone Planning Language Types section. This file contains
// parser rules for type definitions.
// 
// Example CPL types section: types Currency = Number(precision=2) types Percentage = Number(min=0,
// max=100)
// 
// Grammar supports: - Type definitions with Number base type - Constraint specifications for Number
// type

parser grammar CPLTypes;

options {
	tokenVocab = ChronoconePlanningLanguageLexer;
}

// ============================================================================ TYPES SECTION
// ============================================================================ The types section
// declares named types based on built-in types. Each type can have constraints for validation.

// A types section contains one or more type definitions Example: types Currency =
// Number(precision=2)
typesSection: TYPES typeDefinition+;

// A type definition associates a name with a base type and optional constraints Example: Currency =
// Number(precision=2)
typeDefinition:
	indent (
		identifier EQUALS baseType LPAREN constraintList? RPAREN
	)*;

baseType: identifier;
