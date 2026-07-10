// ChronoconePlanningLanguageLexer.g4 Lexer grammar for the Chronocone Planning Language. This file
// contains all lexical definitions (tokens) for CPL.

lexer grammar ChronoconePlanningLanguageLexer;

// ============================================================================ RESERVED KEYWORDS
// FOR SECTION MARKERS ============================================================================

CONCEPTS: 'concepts';
MEASURES: 'measures';
TYPES: 'types';
UNIT: 'unit';
UNITS: 'units';
VALUES: 'Values';
INT: 'Int';

// Concept Range generator token
DOTDOT: '..';

// Before/After keywords for concept extension insertion points.
// Used to specify where new values are inserted relative to existing values
// when extending a concept (e.g., Priority=[Values[Critical]] before Low).
BEFORE: 'before';
AFTER: 'after';

// Connector keywords - to be used in Phase C
CONNECTORS: 'connectors';
DIMENSIONS: 'dimensions';

// Metrics keywords - to be used in Phase E (Metrics implementation)
METRICS: 'metrics';
INPUTS: 'inputs';
CALCULATE: 'calculate';
METRIC: 'metric';
DIMENSION: 'dimension';
LOOKUP: 'lookup';  // Recognized but lookups are deferred to future design
LOOKUPS: 'lookups';  // Section keyword for lookups (plural), distinct from LOOKUP (singular, future input kind)

// Aggregators keywords - to be used in aggregators section
AGGREGATORS: 'aggregators';
FROM: 'from';
TO: 'to';

// Spreaders keywords - to be used in spreaders section
SPREADERS: 'spreaders';

// For the moment the cell type is just 'celltype', but it doesn't
// actually mean anything.
CELLTYPE: 'celltype';

// Time grain keyword for connector declarations
GRAIN: 'grain';

// Conditional expression keywords
IF_KEYWORD       : 'if';
ELSE_KEYWORD     : 'else';
END_KEYWORD      : 'end';
WHEN_KEYWORD     : 'when';
DEFAULT_KEYWORD  : 'default';
AND_KEYWORD      : 'and';
OR_KEYWORD       : 'or';
NOT_KEYWORD      : 'not';

// Alias and where clause keywords for input declarations
AS_KEYWORD       : 'as';
WHERE_KEYWORD    : 'where';
ASPECT_KEYWORD   : 'aspect';
ACTUALS_KEYWORD  : 'actuals';
FORECAST_KEYWORD : 'forecast';

// Time shift keywords for input declarations
BACKWARD_KEYWORD : 'backward';
FORWARD_KEYWORD  : 'forward';

// Conversion keywords
CONVERSION_KEYWORD : 'conversion';
PATTERN_KEYWORD    : 'pattern';
CALCULATION_KEYWORD : 'calculation';

// ============================================================================ PUNCTUATION TOKENS
// ============================================================================

EQUALS: '=';
LPAREN: '(';
RPAREN: ')';
LBRACKET: '[';
RBRACKET: ']';
COMMA: ',';
COLON: ':';
PLUS: '+';
MINUS: '-';
MULTIPLY: '*';
DIVIDE: '/';
ARROW            : '=>';
EQEQ             : '==';
NOTEQ            : '!=';
LT               : '<';
GT               : '>';
LTEQ             : '<=';
GTEQ             : '>=';

INDENT: [\r\n][ \t]*;
NEWLINE: [\r\n];

// ============================================================================ LITERALS
// ============================================================================

// Boolean and Nil literals
TRUE: 'True';
FALSE: 'False';
NIL: 'Nil';

// Exponentiation operator (for arithmetic expressions)
POW: '^';

// NUMBER must appear before INTEGER so that digit sequences are consistently tokenized as NUMBER.
NUMBER: [0-9]+ ('.' [0-9]+)?;

// Integer literals (for constraint values and other numeric uses)
INTEGER: [0-9]+;

// String literals (single or double quoted)
STRING_LITERAL: '"' (~[\r\n])* '"' | '\'' (~['\r\n])* '\'';

// ============================================================================ BUILT-IN TYPE KEYWORDS
// ============================================================================

// ============================================================================ IDENTIFIERS
// ============================================================================

// Identifiers: must start with letter or underscore, followed by letters, digits, or underscores
IDENTIFIER: [a-zA-Z_][a-zA-Z0-9_]*;
VALUE: [a-zA-Z0-9_]+;

// ============================================================================ WHITESPACE AND
// COMMENTS ============================================================================

// Single-line comments starting with #
LINE_COMMENT: [ \t]* '#' ~[\r\n]* [\r\n] -> skip;

// Whitespace (spaces, tabs, newlines)
WS: [ \t]+ -> skip;