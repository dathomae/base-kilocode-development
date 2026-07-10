// CPLUnits.g4 Parser grammar for the Chronocone Planning Language Units section. This file contains
// parser rules for unit definitions.
//
// Example CPL units section:
// units
//     Currency = Country
//     Weight = WeightUnit
//     Length = DistanceUnit
//
// Grammar supports:
// - Unit definitions linking a unit name to a concept name (unit group)
// - Multiple units per section
//
// Units are defined by referencing a concept that serves as the unit group.
// The concept must be defined in the concepts section and contains the
// valid values for that unit.

parser grammar CPLUnits;

options {
	tokenVocab = ChronoconePlanningLanguageLexer;
}

// ============================================================================ UNITS SECTION
// ============================================================================ The units section declares named units that reference concept groups.
// Each unit is defined by linking a unit name to a concept name (the unit group).

// A units section contains one or more unit declarations
// Example: units Currency = Country Weight = Mass
unitsSection: UNITS unitDeclaration+;

// A unit declaration associates a unit name with a concept name (unit group)
// Example: Currency = Country
// This means the Currency unit accepts values from the Country concept
unitDeclaration: indent (identifier EQUALS identifier conversionClause*)*;

// A conversion clause contains one or more conversion pairs
// Example: conversion pattern Kilometer(km) = Meter(m) calculation km * 1000
conversionClause: indent CONVERSION_KEYWORD conversionPair+;

// A conversion pair consists of a pattern clause and a calculation clause
conversionPair: patternClause calculationClause;

// A pattern clause defines the equation-form pattern with bound variables
// Example: pattern Kilometer(km) = Meter(m)
patternClause: indent PATTERN_KEYWORD pattern;

// Pattern equation form: unit_expr(var) = unit_expr(var)
pattern
    : arithmeticExpr patternBoundVariable EQUALS arithmeticExpr patternBoundVariable
    ;

// Pattern-bound variable syntax: (identifier)
patternBoundVariable: LPAREN identifier RPAREN;

// A calculation clause defines the conversion expression
// Example: calculation km * 1000
calculationClause: indent CALCULATION_KEYWORD arithmeticExpr;
