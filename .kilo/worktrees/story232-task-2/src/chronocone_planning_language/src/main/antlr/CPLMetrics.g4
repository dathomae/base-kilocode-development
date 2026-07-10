// CPLMetrics.g4 - Grammar for CPL Metrics section (Increment 1)
//
// This file contains parser rules for metric definitions.
// Metrics define calculations with inputs and calculate clauses.
//
// Example CPL metrics section:
// metrics
//     revenue
//         inputs
//             metric demand_forecast
//             dimension Price
//         calculate
//             0
//
// Grammar supports (Increment 1):
// - Metric declarations with inputs block and calculate clause
// - Input declarations with metric or dimension kind
// - Calculate clause accepting literal 0 (identity calculation)

parser grammar CPLMetrics;

import ChronoconePlanningLanguage;  // For identifier, literal, expression rules

options {
    tokenVocab = ChronoconePlanningLanguageLexer;
}

// ============================================================================ METRICS SECTION
// ============================================================================
// The metrics section declares named metrics that define calculations
// with input references and calculate expressions.

// A metrics section contains one or more metric declarations and/or short form definitions
metricsSection
    : METRICS (metricDeclaration | shortFormDeclaration)+
    ;

// Short form: <name> = <calculation>
// Uses the shared 'expression' rule from the root grammar.
// Short form declarations are converted to equivalent long form ParserMetricDefinition
// by the parser listener; downstream code (analysis, synthesis) treats them identically.
shortFormDeclaration
    : indent identifier EQUALS expression
    ;

// A metric declaration: name, inputs block, calculate clause
// Each component starts on an indented line
metricDeclaration
    : indent identifier inputsBlock calculateClause
    ;

// Inputs block: 'inputs' keyword followed by input declarations
inputsBlock
    : indent INPUTS inputDeclaration+
    ;

// An input declaration: metric|dimension <name> [as <alias>] [where <expression>] [backward(...)] [forward(...)]
inputDeclaration
    : indent inputKind identifier inputAsClause? (WHERE_KEYWORD whereExpression)? backwardClause? forwardClause?
    ;

// Input kind discriminator
inputKind
    : METRIC
    | DIMENSION
    // LOOKUP is recognized by the lexer but lookups are deferred
    ;

// Alias clause for input declarations
inputAsClause
    : AS_KEYWORD identifier
    ;

whereExpression : orExpression ;

orExpression : andExpression (OR_KEYWORD andExpression)* ;

andExpression : aspectExpression (AND_KEYWORD aspectExpression)* ;

aspectExpression : NOT_KEYWORD? ASPECT_KEYWORD EQUALS (ACTUALS_KEYWORD | FORECAST_KEYWORD) ;

// Backward time shift clause for input declarations
backwardClause
    : BACKWARD_KEYWORD LPAREN shiftExpression RPAREN
    ;

// Forward time shift clause for input declarations
forwardClause
    : FORWARD_KEYWORD LPAREN shiftExpression RPAREN
    ;

// Shift expression: constraint field name with optional expression offset
shiftExpression
    : identifier (COMMA expression)?
    ;

// Calculate clause: 'calculate' keyword followed by expression.
// Supports arithmetic expressions: +, -, *, /, unary negation, parentheses.
calculateClause
    : indent CALCULATE indent expression
    ;
