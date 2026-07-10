// ChronoconePlanningLanguage.g4 Chronocone Planning Language Grammar This is the root grammar that
// imports specialized grammars for different sections.

grammar ChronoconePlanningLanguage;

import CPLConcepts, CPLMeasures, CPLTypes, CPLUnits, CPLConnectors, CPLMetrics, CPLLookups, CPLAggregators, CPLSpreaders;

options {
	tokenVocab = ChronoconePlanningLanguageLexer;
}

// ============================================================================ TOP-LEVEL PARSER
// RULES ============================================================================ This section
// defines the syntactic structure of the language.

// A CPL document consists of zero or more statements
document: (indent | statement)* EOF;

// Statements are the top-level elements of a CPL file
statement
	: conceptsSection
	| typesSection
	| unitsSection
	| measuresSection
	| connectorsSection
	| lookupsSection
	| metricsSection
	| aggregatorsSection
	| spreadersSection;



// These are common rule that are used across multiple sections in the language
// Basic identifier rule - shared across all imported grammars
// Built-in type keywords, 'unit', and boolean/Nil literals are also valid identifiers
// in most contexts (e.g., as type references in connectors, measures, constraint names,
// and as metric/concept/connector names). True/False/Nil are included so that existing
// CPL sources using these words as names do not break — the analysis phase resolves
// semantic meaning (literal vs identifier) based on context.
identifier: IDENTIFIER | UNIT | TRUE | FALSE | NIL | CONVERSION_KEYWORD | PATTERN_KEYWORD | CALCULATION_KEYWORD;

value: IDENTIFIER | VALUE | NUMBER | INTEGER;

// Type references can be simple identifiers or parameterized types Examples: Count, Currency,
// Number(decimal_separator: '.')
typeReference: identifier parameterList?;

// Parameter list for type parameters: (param1: value1, param2: value2)
parameterList: LPAREN parameter (COMMA parameter)* RPAREN;

// Single parameter with name and value
parameter: identifier COLON parameterValue;

// Parameter values can be identifiers or string literals
parameterValue: identifier | STRING_LITERAL;

indent: INDENT;

// Comma-separated list of identifiers - shared across connectors, aggregators, and other sections
identifierList: identifier (COMMA identifier)*;

expression
    : conditionalExpr
    ;

// ============================================================================
// CONDITIONAL EXPRESSION RULES (Increment 4 — Phase C implementation)
// ============================================================================
// conditionalExpr supports if/else if/else conditionals.
// Falls through to whenExpr when no if keyword is present.
// whenExpr itself falls through to logicalOrExpr for non-conditional expressions.
// The ifBranch rule handles both inline (`if <cond>`) and separate-line
// (`if \n <cond>`) forms; the indent before the condition is optional.
conditionalExpr
    : IF_KEYWORD
      ifBranch
      elseIfBranch*
      elseBranch?
      (END_KEYWORD IF_KEYWORD | indent END_KEYWORD IF_KEYWORD)
    | whenExpr
    ;

// ifBranch: supports both inline and separate-line condition placement
//   Inline form:    if <cond> \n <body>      (expression indent expression)
//   Separate form:  if \n <cond> \n <body>   (indent expression indent expression)
ifBranch
    : expression indent expression          // Condition at same indent, then body
    | indent expression indent expression   // Condition indented, then body
    ;

// elseIfBranch: else if with condition and indented body
// Supports both inline (else if on same line as preceding body) and
// separate-line (else if on its own line) placement.
elseIfBranch
    : ELSE_KEYWORD IF_KEYWORD expression
      indent expression
    | indent ELSE_KEYWORD IF_KEYWORD expression
      indent expression
    ;

// elseBranch: else with indented body
// Supports both inline (else on same line as preceding body) and
// separate-line (else on its own line) placement.
elseBranch
    : ELSE_KEYWORD
      indent expression
    | indent ELSE_KEYWORD
      indent expression
    ;

// ============================================================================
// WHEN EXPRESSION RULES (Increment 4 — Phase C implementation)
// ============================================================================
// whenExpr supports when/default switch-like conditionals.
whenExpr
    : WHEN_KEYWORD
      indent whenClause+
      defaultClause
      END_KEYWORD WHEN_KEYWORD
    | logicalOrExpr  // Fall-through when no when keyword
    ;

// whenClause: condition => indented body
whenClause
    : expression ARROW
      indent expression
    ;

// defaultClause: default => indented body
defaultClause
    : DEFAULT_KEYWORD ARROW
      indent expression
    ;

// ============================================================================
// LOGICAL EXPRESSION RULES (Increment 4 — Phase C implementation)
// ============================================================================
// logicalOrExpr: logical OR — lowest logical precedence
logicalOrExpr
    : logicalAndExpr (OR_KEYWORD logicalAndExpr)*
    ;

// logicalAndExpr: logical AND — higher precedence than OR
logicalAndExpr
    : notExpr (AND_KEYWORD notExpr)*
    ;

// notExpr: logical NOT — lower precedence than comparisons, higher than AND/OR
// Right-associative via recursion so that "not not a > 0" stacks correctly.
notExpr
    : NOT_KEYWORD notExpr
    | equalityExpr
    ;

// ============================================================================
// COMPARISON EXPRESSION RULES (Increment 4 — Phase C implementation)
// ============================================================================
// equalityExpr: equality and inequality comparisons
equalityExpr
    : relationalExpr ((EQEQ | NOTEQ) relationalExpr)*
    ;

// relationalExpr: ordering comparisons
relationalExpr
    : additionExpr ((LT | GT | LTEQ | GTEQ) additionExpr)*
    ;

// Left-associative addition/subtraction (lowest precedence after expression root)
// PLUS and MINUS tokens defined in ChronoconePlanningLanguageLexer
additionExpr
    : additionExpr PLUS multiplicationExpr
    | additionExpr MINUS multiplicationExpr
    | multiplicationExpr
    ;

// Left-associative multiplication/division (higher precedence than addition)
// MULTIPLY and DIVIDE tokens defined in ChronoconePlanningLanguageLexer
multiplicationExpr
    : multiplicationExpr MULTIPLY unaryExpr
    | multiplicationExpr DIVIDE unaryExpr
    | unaryExpr
    ;

// Unary operators: negation (logical not moved to notExpr)
unaryExpr
    : (MINUS)? primaryExpr
    ;

// Primary expressions: literals, identifiers, parenthesized sub-expressions
primaryExpr
    : literal
    | identifier
    | LPAREN expression RPAREN
    ;

// Literal values - integer, boolean, and Nil
literal: NUMBER | INTEGER | TRUE | FALSE | NIL;

// ============================================================================
// CONSTRAINT RULES (shared across types and measures)
// ============================================================================
// Comma-separated list of constraints
constraintList: constraint (COMMA constraint)*;

// A constraint specification: name = value
constraint: identifier EQUALS constraintValue;

// Constraint value - supports integer literals, string literals, identifiers, and nested constraint calls.
constraintValue: nestedConstraintValue | NUMBER | INTEGER | STRING_LITERAL | identifier;

// Nested constraint values are parsed as a generic call shape and semantically validated in analysis.
nestedConstraintValue: identifier LPAREN constraintList? RPAREN;

// ============================================================================
// ARITHMETIC EXPRESSION RULES (shared across units, measures, and metrics)
// ============================================================================

arithmeticExpr: arithmeticExpr PLUS arithmeticTerm
              | arithmeticExpr MINUS arithmeticTerm
              | arithmeticTerm
              ;

arithmeticTerm: arithmeticTerm MULTIPLY arithmeticFactor
              | arithmeticTerm DIVIDE arithmeticFactor
              | arithmeticFactor
              ;

arithmeticFactor: arithmeticFactor POW (NUMBER | INTEGER)
                | identifier
                | NUMBER
                | LPAREN arithmeticExpr RPAREN
                ;

// Explicit operator callback rules for listener method dispatch
multOperator: MULTIPLY;
divOperator: DIVIDE;
powOperator: POW;
plusOperator: PLUS;
minusOperator: MINUS;