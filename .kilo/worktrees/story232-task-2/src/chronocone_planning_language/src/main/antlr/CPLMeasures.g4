// CPLMeasures.g4 Parser grammar for the Chronocone Planning Language Units section. This file contains
// parser rules for unit definitions.
//
// Example CPL measures section:
// measures
//     CurrencyMeasure = CurrencyType(unit=CurrencyUnit)
//     Currency = Currency(unit=Currency)
//
// Units are defined by referencing a measure group name to a type and a unit group name.

parser grammar CPLMeasures;

options {
	tokenVocab = ChronoconePlanningLanguageLexer;
}

// =============== MEASURES SECTION

// A measures section contains one or more measure declarations
measuresSection: MEASURES measureDeclaration+;

// A measure declaration associates a measure name with a type name and optional parameters.
// Example: Currency = Currency(unit=Currency)
// Example: Daily = Interval(starts_at=Midnight, unit=DurationUnit)
// Parameter interpretation (e.g., unit, starts_at) is deferred to the analysis phase.
measureDeclaration: indent (identifier EQUALS identifier LPAREN measureParameterList? RPAREN measureConversionClause*)*;

// Comma-separated list of measure parameters
measureParameterList: measureParameter (COMMA measureParameter)*;

// A measure parameter can be a unit expression or a generic constraint
measureParameter: unitParameter | constraint;

// Unit parameter using arithmetic expressions (Story229)
// Example: unit=Meter, unit=Meter/Second, unit=Currency*ExchangeRate
unitParameter: UNIT EQUALS arithmeticExpr;

// A measure conversion clause contains one or more measure conversion pairs
measureConversionClause: indent CONVERSION_KEYWORD measureConversionPair+;

// A measure conversion pair: pattern clause followed by a calculation clause
measureConversionPair: patternClause measureCalculationClause;

// A measure calculation clause defines the conversion expression
measureCalculationClause: indent CALCULATION_KEYWORD arithmeticExpr;
