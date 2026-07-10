// Package metrics_arithmetic contains structural definitions for example CPL files
// that demonstrate arithmetic expressions in metric calculate clauses.
//
// Story: Story202 - CPL Metrics Arithmetic Examples Revamp
//
// The examples are organized into subdirectories, each containing:
//   - A metric-only .cpl file (no shared infrastructure boilerplate)
//   - An iospec.json with connector inputs and metric output
//   - parquet_gen_spec files for generating deterministic input data
//   - expected_output.json for automated verification
//
// The shared infrastructure (concepts, types, units, measures)
// lives in core.cpl at the top level of the metrics_arithmetic directory.
// core.cpl must be passed to the compiler alongside each example's .cpl file.
package metrics_arithmetic

// DirectoryPath is the location where all arithmetic example CPL files are stored.
const DirectoryPath = "src/chronocone_planning_language/examples/with_data/without_errors/metrics_arithmetic"

// CoreCPL is the shared infrastructure file defining concepts, types, units,
// and measures used by all arithmetic examples.
// This file must be passed to the compiler alongside each example-specific .cpl file.
const CoreCPL = "core.cpl"

// ============================================================================
// File 1: simple_addition.cpl
// ============================================================================
// Purpose: Demonstrates a metric that adds two input measures using binary '+'.
// Calculate clause: calculate input_a + input_b
// Expected CIL: CILBinaryOp{Op: "+", Left: CILInputRef{input_a}, Right: CILInputRef{input_b}}
//
// concepts section:
//   - Region: a Values concept with North, South, East, West.
//   - Day: a Values concept with Monday (used as grain).
//   - Unit: a Values concept with Each (used for measure units).
//
// types section:
//   - NumberType = Number() (numeric base type for measures).
//
// units section:
//   - Unit mapped to the Unit concept.
//   - Region mapped to the Region concept.
//   - Day mapped to the Day concept.
//
// measures section:
//   - input_a = NumberType(unit=Unit) (first numeric input measure).
//   - input_b = NumberType(unit=Unit) (second numeric input measure).
//   - Daily = Interval(unit=Day) (grain measure for the connector).
//
// connectors section:
//   - input_a = dimensions [Region] grain Daily celltype packages.
//   - input_b = dimensions [Region] grain Daily celltype packages.
//
// metrics section:
//   - sum_metric:
//       inputs: metric input_a, metric input_b.
//       calculate: input_a + input_b.

// SimpleAdditionCPL is the relative path (from metrics_arithmetic/) for the
// simple addition example file. Uses only two inputs (input_a, input_b).
const SimpleAdditionCPL = "simple_addition/simple_addition.cpl"

// ============================================================================
// File 2: multiplication_precedence.cpl
// ============================================================================
// Purpose: Demonstrates operator precedence where '*' binds tighter than '+'.
// Calculate clause: calculate input_a * input_b + input_c
// Expected parse: (input_a * input_b) + input_c
// Expected CIL: CILBinaryOp{Op: "+", Left: CILBinaryOp{Op: "*", ...}, Right: CILInputRef{input_c}}
//
// concepts section:
//   - Region: a Values concept with North, South.
//   - Day: a Values concept with Monday (used as grain).
//   - Unit: a Values concept with Each (used for measure units).
//
// types section:
//   - NumberType = Number() (numeric base type for measures).
//
// units section:
//   - Unit mapped to the Unit concept.
//   - Region mapped to the Region concept.
//   - Day mapped to the Day concept.
//
// measures section:
//   - input_a = NumberType(unit=Unit) (first numeric input).
//   - input_b = NumberType(unit=Unit) (second numeric input).
//   - input_c = NumberType(unit=Unit) (third numeric input).
//   - Daily = Interval(unit=Day) (grain measure).
//
// connectors section:
//   - input_a = dimensions [Region] grain Daily celltype packages.
//   - input_b = dimensions [Region] grain Daily celltype packages.
//   - input_c = dimensions [Region] grain Daily celltype packages.
//
// metrics section:
//   - precedence_metric:
//       inputs: metric input_a, metric input_b, metric input_c.
//       calculate: input_a * input_b + input_c.

// MultiplicationPrecedenceCPL is the relative path for the operator
// precedence example (3 inputs: input_a, input_b, input_c).
const MultiplicationPrecedenceCPL = "multiplication_precedence/multiplication_precedence.cpl"

// ============================================================================
// File 3: subtraction_division.cpl
// ============================================================================
// Purpose: Demonstrates subtraction and division with precedence.
// Calculate clause: calculate input_a - input_b / input_c
// Expected parse: input_a - (input_b / input_c)
// Expected CIL: CILBinaryOp{Op: "-", Left: CILInputRef{input_a}, Right: CILBinaryOp{Op: "/", ...}}
//
// concepts section:
//   - Region: a Values concept with North, South.
//   - Day: a Values concept with Monday (used as grain).
//   - Unit: a Values concept with Each (used for measure units).
//
// types section:
//   - NumberType = Number() (numeric base type for measures).
//
// units section:
//   - Unit mapped to the Unit concept.
//   - Region mapped to the Region concept.
//   - Day mapped to the Day concept.
//
// measures section:
//   - input_a = NumberType(unit=Unit) (first numeric input).
//   - input_b = NumberType(unit=Unit) (second numeric input).
//   - input_c = NumberType(unit=Unit) (third numeric input, divisor).
//   - Daily = Interval(unit=Day) (grain measure).
//
// connectors section:
//   - input_a = dimensions [Region] grain Daily celltype packages.
//   - input_b = dimensions [Region] grain Daily celltype packages.
//   - input_c = dimensions [Region] grain Daily celltype packages.
//
// metrics section:
//   - sub_div_metric:
//       inputs: metric input_a, metric input_b, metric input_c.
//       calculate: input_a - input_b / input_c.

// SubtractionDivisionCPL is the relative path for the subtraction and
// division example (3 inputs: input_a, input_b, input_c with input_b/input_c computed first).
const SubtractionDivisionCPL = "subtraction_division/subtraction_division.cpl"

// ============================================================================
// File 4: unary_negation.cpl
// ============================================================================
// Purpose: Demonstrates unary negation applied to an input measure.
// Calculate clause: calculate -input_a + input_b
// Expected parse: (-input_a) + input_b
// Expected CIL: CILBinaryOp{Op: "+", Left: CILUnaryOp{Op: "negate", Operand: CILInputRef{input_a}}, Right: CILInputRef{input_b}}
//
// concepts section:
//   - Region: a Values concept with North, South.
//   - Day: a Values concept with Monday (used as grain).
//   - Unit: a Values concept with Each (used for measure units).
//
// types section:
//   - NumberType = Number() (numeric base type for measures).
//
// units section:
//   - Unit mapped to the Unit concept.
//   - Region mapped to the Region concept.
//   - Day mapped to the Day concept.
//
// measures section:
//   - input_a = NumberType(unit=Unit) (numeric input to be negated).
//   - input_b = NumberType(unit=Unit) (second numeric input).
//   - Daily = Interval(unit=Day) (grain measure).
//
// connectors section:
//   - input_a = dimensions [Region] grain Daily celltype packages.
//   - input_b = dimensions [Region] grain Daily celltype packages.
//
// metrics section:
//   - negate_metric:
//       inputs: metric input_a, metric input_b.
//       calculate: -input_a + input_b.

// UnaryNegationCPL is the relative path for the unary negation example
// (2 inputs: input_a, input_b with -input_a computed before + input_b).
const UnaryNegationCPL = "unary_negation/unary_negation.cpl"

// ============================================================================
// File 5: parenthesized.cpl
// ============================================================================
// Purpose: Demonstrates explicit parentheses overriding natural precedence.
// Calculate clause: calculate (input_a + input_b) * (input_c - input_d)
// Expected parse: (input_a + input_b) * (input_c - input_d) — parentheses force '+' and '-' before '*'
// Expected CIL: CILBinaryOp{Op: "*", Left: CILBinaryOp{Op: "+", ...}, Right: CILBinaryOp{Op: "-", ...}}
//
// concepts section:
//   - Region: a Values concept with North, South.
//   - Day: a Values concept with Monday (used as grain).
//   - Unit: a Values concept with Each (used for measure units).
//
// types section:
//   - NumberType = Number() (numeric base type for measures).
//
// units section:
//   - Unit mapped to the Unit concept.
//   - Region mapped to the Region concept.
//   - Day mapped to the Day concept.
//
// measures section:
//   - input_a = NumberType(unit=Unit) (first numeric input).
//   - input_b = NumberType(unit=Unit) (second numeric input).
//   - input_c = NumberType(unit=Unit) (third numeric input).
//   - input_d = NumberType(unit=Unit) (fourth numeric input).
//   - Daily = Interval(unit=Day) (grain measure).
//
// connectors section:
//   - input_a = dimensions [Region] grain Daily celltype packages.
//   - input_b = dimensions [Region] grain Daily celltype packages.
//   - input_c = dimensions [Region] grain Daily celltype packages.
//   - input_d = dimensions [Region] grain Daily celltype packages.
//
// metrics section:
//   - paren_metric:
//       inputs: metric input_a, metric input_b, metric input_c, metric input_d.
//       calculate: (input_a + input_b) * (input_c - input_d).

// ParenthesizedCPL is the relative path for the parenthesized expression
// example (4 inputs: input_a, input_b, input_c, input_d).
const ParenthesizedCPL = "parenthesized/parenthesized.cpl"

// ============================================================================
// File 6: complex_expression.cpl (in complex_expression/ subdirectory)
// ============================================================================
// Purpose: Demonstrates a deeply nested expression with multiple operators.
// Calculate clause: calculate input_a * input_b + input_c / input_d - input_e
// Expected parse: ((input_a * input_b) + (input_c / input_d)) - input_e
// Expected CIL: nested CILBinaryOp tree with three binary ops at three levels.
//
// concepts section:
//   - Region: a Values concept with North, South.
//   - Day: a Values concept with Monday (used as grain).
//   - Unit: a Values concept with Each (used for measure units).
//
// types section:
//   - NumberType = Number() (numeric base type for measures).
//
// units section:
//   - Unit mapped to the Unit concept.
//   - Region mapped to the Region concept.
//   - Day mapped to the Day concept.
//
// measures section:
//   - input_a = NumberType(unit=Unit) (first numeric input).
//   - input_b = NumberType(unit=Unit) (second numeric input).
//   - input_c = NumberType(unit=Unit) (third numeric input).
//   - input_d = NumberType(unit=Unit) (fourth numeric input, divisor).
//   - input_e = NumberType(unit=Unit) (fifth numeric input).
//   - Daily = Interval(unit=Day) (grain measure).
//
// connectors section:
//   - input_a = dimensions [Region] grain Daily celltype packages.
//   - input_b = dimensions [Region] grain Daily celltype packages.
//   - input_c = dimensions [Region] grain Daily celltype packages.
//   - input_d = dimensions [Region] grain Daily celltype packages.
//   - input_e = dimensions [Region] grain Daily celltype packages.
//
// metrics section:
//   - complex_metric:
//       inputs: metric input_a, metric input_b, metric input_c, metric input_d, metric input_e.
//       calculate: input_a * input_b + input_c / input_d - input_e.

// ComplexExpressionCPL is the relative path for the deeply nested expression
// example (5 inputs: input_a, input_b, input_c, input_d, input_e).
const ComplexExpressionCPL = "complex_expression/complex_expression.cpl"

// ============================================================================
// File 7: short_form_addition.cpl
// ============================================================================
// Purpose: Demonstrates short form metric syntax with binary '+'.
// Inputs are auto-derived by the analyzer from the calculate expression.
// Calculate clause: input_a + input_b (short form: sum_metric = input_a + input_b)
// Expected CIL: CILBinaryOp{Op: "+", Left: CILInputRef{input_a}, Right: CILInputRef{input_b}}
//
// concepts section:
//   - Region: a Values concept with North, South, East, West.
//   - Day: a Values concept with Monday (used as grain).
//   - Unit: a Values concept with Each (used for measure units).
//
// types section:
//   - NumberType = Number() (numeric base type for measures).
//
// units section:
//   - Unit mapped to the Unit concept.
//   - Region mapped to the Region concept.
//   - Day mapped to the Day concept.
//
// measures section:
//   - input_a = NumberType(unit=Unit) (first numeric input measure).
//   - input_b = NumberType(unit=Unit) (second numeric input measure).
//   - Daily = Interval(unit=Day) (grain measure for the connector).
//
// connectors section:
//   - input_a = dimensions [Region] grain Daily celltype packages.
//   - input_b = dimensions [Region] grain Daily celltype packages.
//
// metrics section:
//   - sum_metric = input_a + input_b.
//       (short form: no explicit inputs or calculate keyword; analyzer auto-detects inputs).

// ShortFormAdditionCPL is the relative path (from metrics_arithmetic/) for the
// short form addition example file. Uses short form syntax (sum_metric = input_a + input_b)
// where inputs are auto-derived by the analyzer.
const ShortFormAdditionCPL = "short_form_addition/short_form_addition.cpl"

// ============================================================================
// Shared structural patterns across all example files
// ============================================================================
//
// Each example CPL file follows the same structure:
//   1. measures — all input measures (input_a, input_b, ...) and the grain measure Daily.
//   2. connectors — input connectors mapping to the example's measures.
//   3. metrics — declares a single metric with inputs and a calculate clause.
//
// The shared infrastructure (concepts, types, units) is defined in core.cpl
// and must be passed to the compiler alongside each example file.
//
// The calculate clause in each file is the only varying section; it exercises
// a specific arithmetic expression pattern as documented above.
