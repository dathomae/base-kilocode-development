// Package metrics_conditionals contains structural definitions for example CPL files
// that demonstrate conditional expressions in metric calculate clauses.
//
// Story: Story188 - CPL Metrics Conditionals Synthesize
package metrics_conditionals

// DirectoryPath is the location where all conditional example CPL files are stored.
const DirectoryPath = "src/chronocone_planning_language/examples/with_data/without_errors/metrics_conditionals"

// CoreCPL is the shared infrastructure file defining concepts, types, units,
// and measures used by all conditional examples.
// This file must be passed to the compiler alongside each example-specific .cpl file.
const CoreCPL = "core.cpl"

// ConditionalIfElseArithmeticCPL demonstrates basic if/else with arithmetic in both branches.
const ConditionalIfElseArithmeticCPL = "conditional_if_else_arithmetic/conditional_if_else_arithmetic.cpl"

// ConditionalIfElseifElseCPL demonstrates three-branch if/elseif/else.
const ConditionalIfElseifElseCPL = "conditional_if_elseif_else/conditional_if_elseif_else.cpl"

// ConditionalWhenDefaultCPL demonstrates when/default syntax.
const ConditionalWhenDefaultCPL = "conditional_when_default/conditional_when_default.cpl"

// ConditionalComparisonOperatorsCPL exercises all six comparison operators.
const ConditionalComparisonOperatorsCPL = "conditional_comparison_operators/conditional_comparison_operators.cpl"

// ConditionalLogicalOperatorsCPL demonstrates logical and, or, and not.
const ConditionalLogicalOperatorsCPL = "conditional_logical_operators/conditional_logical_operators.cpl"

// ConditionalUnaryNegationCPL demonstrates unary negation in conditions and branch bodies.
const ConditionalUnaryNegationCPL = "conditional_unary_negation/conditional_unary_negation.cpl"

// ConditionalShortFormCPL demonstrates short-form metric with conditional expression.
const ConditionalShortFormCPL = "conditional_short_form/conditional_short_form.cpl"

// ConditionalTruthinessCPL demonstrates numeric truthiness and True/False/Nil literals.
const ConditionalTruthinessCPL = "conditional_truthiness/conditional_truthiness.cpl"

// ConditionalTautologicalIfOnlyCPL demonstrates tautological if without else.
const ConditionalTautologicalIfOnlyCPL = "conditional_tautological_if_only/conditional_tautological_if_only.cpl"

// ConditionalNestedArithmeticCPL demonstrates complex parenthesized arithmetic inside conditional branches.
const ConditionalNestedArithmeticCPL = "conditional_nested_arithmetic/conditional_nested_arithmetic.cpl"
