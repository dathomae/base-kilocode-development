// Package analysis provides the analysis phase data structures for the Chronocone Planning Language.
//
// This file contains the metrics analysis phase implementation.
// It processes metric definitions from the symbol table and builds the Metrics lookup map.
// See ERO005: Metrics Analysis Phase for the incremental implementation plan.
package analysis

import (
	"fmt"
	"math"

	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/operator"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parsing"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
)

// validTimeShiftFields is the set of valid Instant constraint field keywords
// that may appear in backward(...) and forward(...) time-shift clauses.
var validTimeShiftFields = map[string]bool{
	"year":        true,
	"month":       true,
	"day":         true,
	"hour":        true,
	"minute":      true,
	"second":      true,
	"day_of_week": true,
	"week_number": true,
}

// AnalyzeMetricsScope is called when entering a MetricsScope during the symbol table walk.
//
// For Increment 1, this function is a placeholder that allows the walker to continue
// to child MetricScope entries. The actual metric analysis is performed by
// AnalyzeMetricScope which is called for each individual MetricScope.
//
// Future increments may use this hook for section-level validation or setup.
//
// Parameters:
//   - table: The MetricsScope symbol table
//   - context: The analysis context (unused for metrics but required for handler signature)
//   - analyzer: The CPLAnalyzer containing the concepts and measures maps for validation
//
// Returns:
//   - nil to continue walking child scopes
func AnalyzeMetricsScope(table *symboltable.SymbolTable, context *AnalysisContext, analyzer *CPLAnalyzer) error {
	// For Increment 1, no section-level processing is needed.
	// Individual metric analysis is handled by AnalyzeMetricScope for each MetricScope.
	return nil
}

// AnalyzeMetricScope processes a single metric definition from a MetricScope.
//
// This function reads the ParserMetricDefinition from the MetricScope's Data field,
// validates input references against declared connectors and concepts, builds the
// CIL tree (for Increment 1: a single CILLiteral(0) node), and creates an
// AnalysisMetricDefinition which is stored in the analyzer's metrics map.
//
// Symbol Table Structure:
//
//	The parser attaches a *parsing.ParserMetricDefinition to table.Data before
//	popping the MetricScope. That structured record contains:
//	- Name: The metric identifier
//	- Inputs: Slice of ParserInputDefinition entries (metric/dimension references)
//	- CalculatePostfix: Postfix expression sequence (for Increment 1: single literal)
//	- SourceLocation: Location of the metric declaration
//
// Input Validation:
//
//	Each input reference is validated against the analyzer's declared entities:
//	- metric-type inputs: Must exist in the measures map (connector references are
//	  stored in measures during connector analysis)
//	- dimension-type inputs: Must exist in the concepts map
//
// CIL Tree Construction:
//
//	Two paths exist for building the CIL tree:
//	1. Flat postfix path (calculate-based metrics): buildCILTree uses a stack-based
//	   postfix evaluator to process the CalculatePostfix slice from
//	   ParserMetricDefinition. The postfix may contain arithmetic expressions
//	   (binary: + - * /, unary: negate), comparison and logical operators,
//	   literal operands (int, int64, float64, bool, nil), and identifier
//	   references (string names).
//	2. Conditional assembly path (if/when metrics): When the metric scope contains
//	   conditional clause definitions (IfClauseDefinition, WhenClauseDefinition, etc.),
//	   assembleConditional walks the clause definitions, builds condition and body
//	   CIL trees from their postfix sequences, and assembles a CILConditional or
//	   CILWhenClause node.
//
//	The resulting CIL tree is rooted at a CILMetricNode with:
//	- CILNode: the root of the expression tree (CILLiteral, CILBinaryOp, CILConditional, etc.)
//	- InputRefs: collected from CILInputRef nodes found in the expression tree
//
// CIL Tree Validation:
//
//	After construction, the CIL tree is validated with a single call to
//	validateCILTree. The consolidated walker performs type, truthiness,
//	and completeness checks in one traversal, replacing the previous
//	three-walker design. Validation errors are returned with the format:
//	  "validation error in metric %s: %w"
//	where the wrapped error carries a category prefix:
//	  "type error:"         — arithmetic operand is not numeric
//	  "truthiness error:"   — nil literal appears in an arithmetic context
//	  "completeness error:" — conditional branch is missing or has a nil body
//
// Parameters:
//   - table: The MetricScope symbol table whose Data field contains the structured metric definition
//   - context: The analysis context (unused for metrics but required for handler signature)
//   - analyzer: The CPLAnalyzer containing the concepts and measures maps for validation
//
// Returns:
//   - nil if the metric definition is successfully processed
//   - error if structured data validation or reference validation fails
//
// Error Cases:
//   - "metric scope has no MetricScopeDefinition in its Definitions" - No MetricScopeDefinition entry found in Definitions
//   - "duplicate metric definition: <name>" - Metric already defined in the analyzer-wide metrics map
//   - "metric <name> references undefined metric input: <input>" - Metric input not found in measures
//   - "metric <name> references undefined dimension input: <input>" - Dimension input not found in concepts
//   - "no conditional clause definitions found in metric <name>" - conditional assembly called but no clauses found
//   - "mixed if and when clause definitions in metric <name>" - both if and when clauses present
//   - "clause definition in metric <name> has nil Data" - clause definition Data is nil
//   - "clause definition in metric <name> has nil BodyScope" - clause BodyScope is nil
//   - "clause definition in metric <name> has unexpected Data type" - clause Data is not ClauseData
//   - "validation error in metric <name>: <wrapped>" - CIL tree validation failed;
//     the wrapped error carries a category prefix ("type error:",
//     "truthiness error:", or "completeness error:")
func AnalyzeMetricScope(table *symboltable.SymbolTable, context *AnalysisContext, analyzer *CPLAnalyzer) error {
	// Retrieve structured metric definition from Definitions
	var metricDef *parsing.ParserMetricDefinition
	for _, def := range table.Definitions {
		if def.DefinitionType == symboltable.MetricScopeDefinition {
			if md, ok := def.Data.(*parsing.ParserMetricDefinition); ok {
				metricDef = md
				break
			}
		}
	}
	if metricDef == nil {
		return fmt.Errorf("metric scope has no MetricScopeDefinition in its Definitions")
	}

	// Validate metric name
	if metricDef.Name == "" {
		return fmt.Errorf("metric declaration has no name")
	}

	metricName := metricDef.Name

	// Check for duplicate metric definitions
	if analyzer.metricAnalyzer.HasMetric(metricName) {
		return fmt.Errorf("duplicate metric definition: %s", metricName)
	}

	// Build CIL tree: use conditional assembly path for if/when metrics,
	// flat postfix path for calculate-based metrics.
	var cilRoot *cil.CILMetricNode
	var err error
	if hasConditionalClauses(table) {
		cilRoot, err = assembleConditional(table, metricName, analyzer)
	} else {
		cilRoot, err = buildCILTree(metricDef, analyzer)
	}
	if err != nil {
		return fmt.Errorf("failed to build CIL tree for metric %s: %w", metricName, err)
	}

	// Validate the CIL tree using the consolidated walker (type, truthiness, completeness).
	// Per constraint 8, types are initially float64; detailed type selection is deferred.
	if err := validateCILTree(cilRoot.CILNode); err != nil {
		return fmt.Errorf("validation error in metric %s: %w", metricName, err)
	}

	// Build inputs: for short form, derive from CIL tree's collected InputRefs;
	// for long form, use explicitly declared inputs.
	var inputs []AnalysisInputDefinition
	if metricDef.ShortForm {
		// Short form: build inputs from CIL tree's collected InputRefs.
		// The InputRefs were resolved during buildCILTree via resolveInputRefKind
		// with proper precedence (metrics → dimensions) and collision detection.
		//
		// Kind normalization: the CIL tree uses "connector" internally to distinguish
		// leaf data sources (measures, connectors) from intermediate metric computations.
		// The user-facing AnalysisInputDefinition uses the same keyword taxonomy as
		// long form — "metric" encompasses measures, connectors, and other metrics.
		// "dimension" matches concepts. We normalize here so short form produces
		// identical AnalysisInputDefinition structures to equivalent long form.
		for _, ref := range cilRoot.InputRefs {
			kind := ref.Kind
			if kind == "connector" {
				kind = "metric"
			}
			inputs = append(inputs, AnalysisInputDefinition{
				Kind:  kind,
				Name:  ref.Name,
				Alias: ref.Name,
			})
		}
	} else {
		// Long form: resolve aliases, validate, and collect explicitly declared inputs
		resolved, err := resolveMetricInputs(metricDef.Inputs, metricName)
		if err != nil {
			return err
		}
		for _, resolvedInput := range resolved {
			// Self-reference with a time shift is a valid time-series dependency;
			// skip entity validation for these inputs.
			if resolvedInput.Name == metricName && resolvedInput.Kind == "metric" && resolvedInput.TimeShift != nil {
				inputs = append(inputs, resolvedInput)
				continue
			}

			// Validate the underlying input name (not alias) against declared entities
			if err := validateInput(parsing.ParserInputDefinition{
				Kind: resolvedInput.Kind,
				Name: resolvedInput.Name,
			}, metricName, analyzer); err != nil {
				return err
			}
			inputs = append(inputs, resolvedInput)
		}
	}

	// Create analysis-level metric definition
	analysisDef := AnalysisMetricDefinition{
		Name:    metricName,
		Inputs:  inputs,
		CILRoot: cilRoot,
	}

	// Store the metric definition
	analyzer.metricAnalyzer.AddMetric(metricName, analysisDef)

	return nil
}

// evaluatePostfixToFloat64 evaluates a postfix expression sequence containing
// only literals and arithmetic operators, returning a single float64 value.
//
// Supported operators: +, -, *, /, negate (unary minus).
//
// Parameters:
//   - postfix: The typed postfix element sequence
//   - metricName: The metric name for error messages
//
// Returns:
//   - The evaluated float64 value
//   - error if the expression contains unsupported elements or stack underflow
func evaluatePostfixToFloat64(postfix []parsing.PostfixElement, metricName string) (float64, error) {
	stack := make([]float64, 0)
	for _, elem := range postfix {
		switch elem := elem.(type) {
		case parsing.PostfixLiteral:
			stack = append(stack, elem.Value)
		case parsing.PostfixOperator:
			switch elem.DefType {
			case symboltable.PlusOperatorDefinition:
				if len(stack) < 2 {
					return 0, fmt.Errorf("postfix stack underflow in time-shift expression of metric %q", metricName)
				}
				right := stack[len(stack)-1]
				left := stack[len(stack)-2]
				stack = stack[:len(stack)-2]
				stack = append(stack, left+right)
			case symboltable.MinusOperatorDefinition:
				if len(stack) < 2 {
					return 0, fmt.Errorf("postfix stack underflow in time-shift expression of metric %q", metricName)
				}
				right := stack[len(stack)-1]
				left := stack[len(stack)-2]
				stack = stack[:len(stack)-2]
				stack = append(stack, left-right)
			case symboltable.MultiplyOperatorDefinition:
				if len(stack) < 2 {
					return 0, fmt.Errorf("postfix stack underflow in time-shift expression of metric %q", metricName)
				}
				right := stack[len(stack)-1]
				left := stack[len(stack)-2]
				stack = stack[:len(stack)-2]
				stack = append(stack, left*right)
			case symboltable.DivideOperatorDefinition:
				if len(stack) < 2 {
					return 0, fmt.Errorf("postfix stack underflow in time-shift expression of metric %q", metricName)
				}
				right := stack[len(stack)-1]
				if right == 0 {
					return 0, fmt.Errorf("division by zero in time-shift expression of metric %q", metricName)
				}
				left := stack[len(stack)-2]
				stack = stack[:len(stack)-2]
				stack = append(stack, left/right)
			case symboltable.PowerOperatorDefinition:
				if len(stack) < 2 {
					return 0, fmt.Errorf("postfix stack underflow in time-shift expression of metric %q", metricName)
				}
				right := stack[len(stack)-1]
				left := stack[len(stack)-2]
				stack = stack[:len(stack)-2]
				stack = append(stack, math.Pow(left, right))
			case symboltable.NegateOperatorDefinition:
				if len(stack) < 1 {
					return 0, fmt.Errorf("postfix stack underflow in time-shift expression of metric %q", metricName)
				}
				operand := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				stack = append(stack, -operand)
			default:
				return 0, fmt.Errorf("unsupported operator in time-shift expression of metric %q", metricName)
			}
		case parsing.PostfixIdentifier:
			return 0, fmt.Errorf("identifier references not supported in time-shift expression of metric %q", metricName)
		default:
			return 0, fmt.Errorf("unsupported element in time-shift expression of metric %q", metricName)
		}
	}
	if len(stack) != 1 {
		return 0, fmt.Errorf("time-shift expression of metric %q did not evaluate to a single value", metricName)
	}
	return stack[0], nil
}

// resolveTimeShift translates a ParserTimeShift to an AnalysisTimeShift,
// validating that the constraint field keyword is recognized and evaluating
// the shift amount expression if present.
//
// Parameters:
//   - ps: The ParserTimeShift from the parser input definition, or nil
//   - metricName: The name of the metric containing this time shift (for error messages)
//
// Returns:
//   - nil if ps is nil
//   - A pointer to AnalysisTimeShift with the translated values
//   - error if the field keyword is not in validTimeShiftFields or expression evaluation fails
func resolveTimeShift(ps *parsing.ParserTimeShift, metricName string) (*AnalysisTimeShift, error) {
	if ps == nil {
		return nil, nil
	}
	if !validTimeShiftFields[ps.Field] {
		return nil, fmt.Errorf("invalid time-shift field '%s' at line %d, column %d",
			ps.Field, ps.SourceLocation.Line, ps.SourceLocation.Column)
	}

	amount := ps.Amount
	if len(ps.Expression) > 0 {
		exprValue, err := evaluatePostfixToFloat64(ps.Expression, metricName)
		if err != nil {
			return nil, err
		}
		amount = exprValue
	}

	return &AnalysisTimeShift{
		Direction: ps.Direction,
		Field:     ps.Field,
		Amount:    amount,
	}, nil
}

// resolveMetricInputs converts parser input definitions to analysis input definitions,
// resolving aliases, where clauses, and time shifts. It validates alias uniqueness
// within the metric and detects prohibited self-references without time shifts.
//
// Parameters:
//   - parserInputs: The slice of ParserInputDefinition from the metric definition
//   - metricName: The name of the metric being resolved (for error messages)
//
// Returns:
//   - A slice of AnalysisInputDefinition with resolved aliases, where clauses, and time shifts
//   - error if duplicate aliases, invalid time-shift, or self-reference without time shift are found
func resolveMetricInputs(parserInputs []parsing.ParserInputDefinition, metricName string) ([]AnalysisInputDefinition, error) {
	seenAliases := make(map[string]bool)
	resolved := make([]AnalysisInputDefinition, 0, len(parserInputs))

	for _, input := range parserInputs {
		alias := input.Alias
		if alias == "" {
			alias = input.Name
		}

		if seenAliases[alias] {
			return nil, fmt.Errorf("duplicate alias '%s' in metric '%s'", alias, metricName)
		}
		seenAliases[alias] = true

		timeShift, err := resolveTimeShift(input.TimeShift, metricName)
		if err != nil {
			loc := input.NameLocation
			if input.TimeShift != nil {
				loc = input.TimeShift.SourceLocation
			}
			return nil, fmt.Errorf("metric '%s' has input '%s' with invalid time-shift at line %d, column %d: %w",
				metricName, input.Name, loc.Line, loc.Column, err)
		}

		resolved = append(resolved, AnalysisInputDefinition{
			Kind:           input.Kind,
			Name:           input.Name,
			Alias:          alias,
			WhereClause:    resolveWhereClause(input.WhereClause),
			TimeShift:      timeShift,
			SourceLocation: input.NameLocation,
		})
	}

	for _, input := range resolved {
		if input.Name == metricName && input.Kind == "metric" && input.TimeShift == nil {
			return nil, fmt.Errorf("metric '%s' references itself without backward/forward time shift at line %d, column %d",
				metricName, input.SourceLocation.Line, input.SourceLocation.Column)
		}
	}

	return resolved, nil
}

// resolveWhereClause builds an AnalysisWhereClause from a parser where clause.
// Returns nil if the parser clause is nil or has no and-groups.
//
// Parameters:
//   - parserClause: The ParserWhereClause from the input definition
//
// Returns:
//   - A pointer to AnalysisWhereClause populated by walking the parser tree, or nil
func resolveWhereClause(parserClause *parsing.ParserWhereClause) *AnalysisWhereClause {
	if parserClause == nil || len(parserClause.AndGroups) == 0 {
		return nil
	}
	whereClause := &AnalysisWhereClause{
		AndGroups: make([]AnalysisWhereAndGroup, len(parserClause.AndGroups)),
	}
	for i, parserGroup := range parserClause.AndGroups {
		group := AnalysisWhereAndGroup{
			Conditions: make([]AnalysisWhereCondition, len(parserGroup.Conditions)),
		}
		for j, parserCond := range parserGroup.Conditions {
			group.Conditions[j] = AnalysisWhereCondition{
				Negated: parserCond.Negated,
				Aspect:  parserCond.Aspect,
			}
		}
		whereClause.AndGroups[i] = group
	}
	return whereClause
}

// validateInput validates a single input reference against declared entities.
//
// For metric-type inputs: checks if the referenced name exists in the measures map
// or the connectors map.
// For dimension-type inputs: checks if the referenced name exists in the concepts map.
//
// Parameters:
//   - input: The ParserInputDefinition to validate
//   - metricName: The name of the metric being validated (for error messages)
//   - analyzer: The CPLAnalyzer containing the concepts and measures maps
//
// Returns:
//   - nil if the input reference is valid
//   - error if the input reference cannot be resolved
func validateInput(input parsing.ParserInputDefinition, metricName string, analyzer *CPLAnalyzer) error {
	switch input.Kind {
	case "metric":
		// Metric inputs reference measures or connectors
		if _, exists := analyzer.measures[input.Name]; !exists {
			if !analyzer.connectorAnalyzer.HasConnector(input.Name) {
				return fmt.Errorf("metric %s references undefined metric input: %s", metricName, input.Name)
			}
		}
	case "dimension":
		// Dimension inputs reference concepts
		if !analyzer.conceptAnalyzer.HasConcept(input.Name) {
			return fmt.Errorf("metric %s references undefined dimension input: %s", metricName, input.Name)
		}
	default:
		return fmt.Errorf("metric %s has unknown input kind: %s", metricName, input.Kind)
	}
	return nil
}

// buildCILTree constructs the CIL tree for a metric definition by evaluating
// its postfix expression sequence from ParserMetricDefinition.CalculatePostfix.
//
// This function implements a stack-based postfix evaluator. Each element in the
// CalculatePostfix slice ([]parsing.PostfixElement) is either an operand (pushed onto the
// stack) or an operator (pops operands from the stack and pushes the result).
// After processing all elements, the stack must contain exactly one node, which
// becomes the root of the CIL expression tree.
//
// Postfix Element Types and Their CIL Mappings:
//
// Operands (push onto stack):
//   - int    → CILLiteral{Value: float64(v)}
//   - int64  → CILLiteral{Value: float64(v)}
//   - float64 → CILLiteral{Value: v}
//   - string (identifier name) → CILInputRef{Kind: "metric", Name: v}
//
// Binary Operators (pop two from stack, push result):
//
//   - symboltable.PlusOperatorDefinition     → CILBinaryOp{Op: operator.Add, Left, Right}
//
//   - symboltable.MinusOperatorDefinition    → CILBinaryOp{Op: operator.Subtract, Left, Right}
//
//   - symboltable.MultiplyOperatorDefinition → CILBinaryOp{Op: operator.Multiply, Left, Right}
//
//   - symboltable.DivideOperatorDefinition   → CILBinaryOp{Op: operator.Divide, Left, Right}
//
//     For binary operators, the first popped value is the right operand and
//     the second popped value is the left operand (standard postfix convention).
//
// Unary Operators (pop one from stack, push result):
//   - symboltable.NegateOperatorDefinition → CILUnaryOp{Op: operator.Negate, Operand}
//
// Stack Evaluation Algorithm:
//
//  1. Initialize an empty []cil.CILNode stack.
//  2. Iterate over metricDef.CalculatePostfix in order.
//  3. For each element, type-switch to determine if it's an operand or operator.
//  4. Operands are pushed directly onto the stack.
//  5. Binary operators pop two nodes, construct the CILBinaryOp, and push it.
//  6. Unary operators pop one node, construct the CILUnaryOp, and push it.
//  7. After all elements, verify the stack has exactly one element.
//  8. Walk the resulting CIL tree to collect all CILInputRef nodes.
//  9. Build CILMetricNode with the root CILNode and collected InputRefs.
//
// InputRefs Collection:
//
//	After the expression tree is built, the tree is walked to collect all
//	CILInputRef nodes. Each CILInputRef is converted to a cil.CILInputRef
//	{Kind: "metric", Name: identifier} entry and stored in CILMetricNode.InputRefs.
//	This enables dependency analysis and validation of expression identifiers.
//
// Error Cases:
//   - "postfix stack underflow: need 2 operands for binary operator %s" -
//     Less than two nodes on the stack when a binary operator is encountered.
//   - "postfix stack underflow: need 1 operand for unary operator %s" -
//     Stack is empty when a unary operator is encountered.
//   - "postfix stack has %d elements after evaluation, expected exactly 1" -
//     Stack does not have exactly one element after processing all postfix.
//   - "unexpected postfix element type %T" -
//     An element in CalculatePostfix has an unrecognized type.
//
// Increment 1 Backwards Compatibility:
//
//	When CalculatePostfix contains a single literal (int/int64/float64), the
//	function produces a simple CILMetricNode with a CILLiteral root node and
//	empty InputRefs. This behavior is identical to the original Increment 1
//	implementation.
//
// Parameters:
//   - metricDef: The ParserMetricDefinition containing the calculate postfix
//
// Returns:
//   - A pointer to the constructed CILMetricNode with the expression tree and InputRefs
//   - An error if the postfix evaluation fails
//
// resolveInputRefKind resolves a name reference to its input kind using the
// precedence: metrics first, then dimensions (concepts), then lookups.
//
// Precedence order:
//  1. metrics (measures, connectors, or other metrics)
//  2. dimensions (concepts)
//  3. lookups
//
// Collision detection: if a name matches both a metric and a dimension,
// a collision error is returned.
func resolveInputRefKind(name string, metricName string, analyzer *CPLAnalyzer) (string, error) {
	// Track which entity types the name matches for collision detection
	matchesMeasure := false
	matchesConnector := false
	matchesMetric := false
	matchesConcept := false
	matchesLookup := false

	if _, exists := analyzer.measures[name]; exists {
		matchesMeasure = true
	}
	if analyzer.connectorAnalyzer.HasConnector(name) {
		matchesConnector = true
	}
	if analyzer.metricAnalyzer.HasMetric(name) {
		matchesMetric = true
	}
	if analyzer.conceptAnalyzer.HasConcept(name) {
		matchesConcept = true
	}
	if _, exists := analyzer.lookups[name]; exists {
		matchesLookup = true
	}

	// Determine kind with precedence:
	//  1. Connectors/measures (leaf data sources) → "connector"
	//  2. Other metrics → "metric"
	//  3. Concepts/dimensions → "dimension"
	//  4. Lookups → "lookup"
	//
	// Collision detection: a name matching categories from different precedence
	// levels (metric-group vs dimension-group) is an error.

	matchesMetricGroup := matchesMeasure || matchesConnector || matchesMetric
	matchesDimensionGroup := matchesConcept

	// Collision: name matches both a metric-type and a dimension-type entity
	if matchesMetricGroup && matchesDimensionGroup {
		return "", fmt.Errorf("name collision: %q matches both a metric and a dimension in metric %q", name, metricName)
	}

	if matchesMeasure || matchesConnector {
		return "connector", nil
	}
	if matchesMetric {
		return "metric", nil
	}
	if matchesConcept {
		return "dimension", nil
	}
	if matchesLookup {
		return "lookup", nil
	}
	return "", fmt.Errorf("undefined reference %q in metric %q", name, metricName)
}

func buildCILTree(metricDef *parsing.ParserMetricDefinition, analyzer *CPLAnalyzer) (*cil.CILMetricNode, error) {
	return buildCILTreeFromPostfix(metricDef.CalculatePostfix, metricDef.Name, analyzer)
}

// buildCILTreeFromPostfix constructs a CILMetricNode from a raw postfix element sequence.
//
// This is the core postfix evaluator shared by both the flat postfix path (used
// for calculate clauses and conditional expressions) and the conditional assembly
// path. It takes a []parsing.PostfixElement sequence instead of a ParserMetricDefinition,
// allowing it to be called from both buildCILTree and assembleConditional.
//
// Parameters:
//   - postfix: The typed postfix element sequence (PostfixLiteral, PostfixIdentifier, PostfixOperator)
//   - metricName: The metric name for error messages and input ref resolution
//   - analyzer: The CPLAnalyzer for resolving input reference kinds
//
// Returns:
//   - A pointer to the constructed CILMetricNode with the expression tree and InputRefs
//   - An error if the postfix evaluation fails
func buildCILTreeFromPostfix(postfix []parsing.PostfixElement, metricName string, analyzer *CPLAnalyzer) (*cil.CILMetricNode, error) {
	// 1. Initialize empty stack for postfix evaluation.
	stack := make([]cil.CILNode, 0)

	// 2. Iterate over postfix elements.
	for _, elem := range postfix {
		switch elem := elem.(type) {
		// Numeric literal
		case parsing.PostfixLiteral:
			stack = append(stack, &cil.CILLiteral{Value: elem.Value})

		// Identifier reference
		case parsing.PostfixIdentifier:
			kind, err := resolveInputRefKind(elem.Name, metricName, analyzer)
			if err != nil {
				return nil, err
			}
			if kind == "lookup" {
				// Create a CILLookupNode for lookup references
				lookupInfo := analyzer.lookups[elem.Name]
				dimNames := make([]string, len(lookupInfo.Dimensions))
				for i, dim := range lookupInfo.Dimensions {
					dimNames[i] = dim.ConceptName
				}
				stack = append(stack, &cil.CILLookupNode{
					LookupName: elem.Name,
					Dimensions: dimNames,
					Celltype:   lookupInfo.CelltypeMeasure.Name,
				})
			} else {
				ref := &cil.CILInputRef{Kind: kind, Name: elem.Name}
				// For connector references, look up the connector's celltype measure
				// and propagate its unit name so the CIL unit checker can verify
				// conversions when the IO spec declares a different output unit.
				if kind == "connector" {
					conn := analyzer.connectorAnalyzer.GetConnector(elem.Name)
					if conn.CellType != nil {
						ref.UnitName = GetUnitNameFromExpression(conn.CellType.UnitExpression)
					}
				}
				stack = append(stack, ref)
			}

		// Operator or boolean/nil literal carried as DefinitionType
		case parsing.PostfixOperator:
			op := operatorSymbol(elem.DefType)
			switch elem.DefType {
			case symboltable.PlusOperatorDefinition, symboltable.MinusOperatorDefinition,
				symboltable.MultiplyOperatorDefinition, symboltable.DivideOperatorDefinition:
				if len(stack) < 2 {
					return nil, fmt.Errorf("postfix stack underflow: need 2 operands for binary operator %s", op)
				}
				right := stack[len(stack)-1]
				left := stack[len(stack)-2]
				stack = stack[:len(stack)-2]
				stack = append(stack, &cil.CILBinaryOp{Op: op, Left: left, Right: right})

			case symboltable.NegateOperatorDefinition:
				if len(stack) < 1 {
					return nil, fmt.Errorf("postfix stack underflow: need 1 operand for unary operator %s", op)
				}
				operand := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				stack = append(stack, &cil.CILUnaryOp{Op: op, Operand: operand})

			// Comparison operators (pop 2 nodes)
			case symboltable.EqOperatorDefinition, symboltable.NotEqOperatorDefinition,
				symboltable.LessThanOperatorDefinition, symboltable.GreaterThanOperatorDefinition,
				symboltable.LessThanEqOperatorDefinition, symboltable.GreaterThanEqOperatorDefinition:
				if len(stack) < 2 {
					return nil, fmt.Errorf("postfix stack underflow: need 2 operands for comparison operator %s", op)
				}
				right := stack[len(stack)-1]
				left := stack[len(stack)-2]
				stack = stack[:len(stack)-2]
				stack = append(stack, &cil.CILComparisonOp{Op: op, Left: left, Right: right})

			// Logical AND/OR (pop 2 nodes)
			case symboltable.AndOperatorDefinition, symboltable.OrOperatorDefinition:
				if len(stack) < 2 {
					return nil, fmt.Errorf("postfix stack underflow: need 2 operands for logical operator %s", op)
				}
				right := stack[len(stack)-1]
				left := stack[len(stack)-2]
				stack = stack[:len(stack)-2]
				stack = append(stack, &cil.CILLogicalOp{Op: op, Left: left, Right: right})

			// Logical NOT (pop 1 node)
			case symboltable.NotOperatorDefinition:
				if len(stack) < 1 {
					return nil, fmt.Errorf("postfix stack underflow: need 1 operand for logical operator %s", op)
				}
				operand := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				stack = append(stack, &cil.CILNotOp{Operand: operand})

			// Boolean literals (no pop)
			case symboltable.TrueLiteralDefinition:
				stack = append(stack, &cil.CILBoolLiteral{Value: true})
			case symboltable.FalseLiteralDefinition:
				stack = append(stack, &cil.CILBoolLiteral{Value: false})

			// Nil literal (no pop)
			case symboltable.NilLiteralDefinition:
				stack = append(stack, &cil.CILNilLiteral{})

			default:
				return nil, fmt.Errorf("unexpected postfix operator type %v", elem.DefType)
			}

		default:
			return nil, fmt.Errorf("unexpected postfix element type %T", elem)
		}
	}

	// 4. After loop, verify stack has exactly one element.
	if len(stack) != 1 {
		return nil, fmt.Errorf("postfix stack has %d elements after evaluation, expected exactly 1", len(stack))
	}

	// 5. Walk the resulting CIL tree to collect InputRefs from CILInputRef nodes.
	inputRefs, err := collectInputRefs(stack[0])
	if err != nil {
		return nil, fmt.Errorf("failed to collect input refs: %w", err)
	}

	// 6. Return CILMetricNode with the root CILNode and collected InputRefs.
	return &cil.CILMetricNode{
		CILNode:   stack[0],
		InputRefs: inputRefs,
	}, nil
}

// operatorSymbol returns the operator.Operator for a postfix operator definition.
//
// Binary operators return their corresponding Operator constant (e.g., Add, Subtract).
// The negation operator returns Negate. Unknown definitions return Unknown.
//
// This helper centralizes operator-to-typed mapping so that error messages,
// CILBinaryOp.Op, and CILUnaryOp.Op all derive from the same source.
func operatorSymbol(def symboltable.DefinitionType) operator.Operator {
	switch def {
	case symboltable.PlusOperatorDefinition:
		return operator.Add
	case symboltable.MinusOperatorDefinition:
		return operator.Subtract
	case symboltable.MultiplyOperatorDefinition:
		return operator.Multiply
	case symboltable.DivideOperatorDefinition:
		return operator.Divide
	case symboltable.NegateOperatorDefinition:
		return operator.Negate
	case symboltable.EqOperatorDefinition:
		return operator.Equal
	case symboltable.NotEqOperatorDefinition:
		return operator.NotEqual
	case symboltable.LessThanOperatorDefinition:
		return operator.LessThan
	case symboltable.GreaterThanOperatorDefinition:
		return operator.GreaterThan
	case symboltable.LessThanEqOperatorDefinition:
		return operator.LessThanOrEqual
	case symboltable.GreaterThanEqOperatorDefinition:
		return operator.GreaterThanOrEqual
	case symboltable.AndOperatorDefinition:
		return operator.And
	case symboltable.OrOperatorDefinition:
		return operator.Or
	case symboltable.NotOperatorDefinition:
		return operator.Not
	default:
		return operator.Unknown
	}
}

// collectInputRefs walks the CIL expression tree and collects all unique
// CILInputRef nodes, preserving order of first occurrence and deduplicating
// by name.
func collectInputRefs(node cil.CILNode) ([]cil.CILInputRef, error) {
	seen := make(map[string]bool)
	var refs []cil.CILInputRef
	if err := collectInputRefsRecursive(node, seen, &refs); err != nil {
		return nil, err
	}
	return refs, nil
}

// collectInputRefsRecursive recursively walks the CIL tree to find CILInputRef nodes.
func collectInputRefsRecursive(node cil.CILNode, seen map[string]bool, refs *[]cil.CILInputRef) error {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *cil.CILInputRef:
		if !seen[n.Name] {
			seen[n.Name] = true
			*refs = append(*refs, cil.CILInputRef{Kind: n.Kind, Name: n.Name})
		}

	case *cil.CILLookupNode:
		// Lookup nodes are leaf references; collect them as input refs
		if !seen[n.LookupName] {
			seen[n.LookupName] = true
			*refs = append(*refs, cil.CILInputRef{Kind: "lookup", Name: n.LookupName})
		}

	case *cil.CILBinaryOp:
		if err := collectInputRefsRecursive(n.Left, seen, refs); err != nil {
			return err
		}
		if err := collectInputRefsRecursive(n.Right, seen, refs); err != nil {
			return err
		}

	case *cil.CILUnaryOp:
		if err := collectInputRefsRecursive(n.Operand, seen, refs); err != nil {
			return err
		}

	case *cil.CILLiteral:
		// No children — nothing to collect.

	case *cil.CILComparisonOp:
		if err := collectInputRefsRecursive(n.Left, seen, refs); err != nil {
			return err
		}
		if err := collectInputRefsRecursive(n.Right, seen, refs); err != nil {
			return err
		}

	case *cil.CILLogicalOp:
		if err := collectInputRefsRecursive(n.Left, seen, refs); err != nil {
			return err
		}
		if err := collectInputRefsRecursive(n.Right, seen, refs); err != nil {
			return err
		}

	case *cil.CILNotOp:
		if err := collectInputRefsRecursive(n.Operand, seen, refs); err != nil {
			return err
		}

	case *cil.CILPower:
		if err := collectInputRefsRecursive(n.Base, seen, refs); err != nil {
			return err
		}

	case *cil.CILBoolLiteral, *cil.CILNilLiteral:
		// No children — nothing to collect.

	case *cil.CILConditional:
		for _, branch := range n.Branches {
			if branch.Condition != nil {
				if err := collectInputRefsRecursive(branch.Condition, seen, refs); err != nil {
					return err
				}
			}
			if err := collectInputRefsRecursive(branch.Body, seen, refs); err != nil {
				return err
			}
		}

	case *cil.CILWhenClause:
		for _, clause := range n.Clauses {
			if clause.Condition != nil {
				if err := collectInputRefsRecursive(clause.Condition, seen, refs); err != nil {
					return err
				}
			}
			if err := collectInputRefsRecursive(clause.Body, seen, refs); err != nil {
				return err
			}
		}
		if n.Default != nil {
			if err := collectInputRefsRecursive(n.Default, seen, refs); err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("collectInputRefsRecursive: unknown CILNode type %T", node)
	}
	return nil
}

// isNumericNode returns true if the given CILNode is a numeric-producing node
// in the CIL expression tree.
//
// Numeric nodes are those that produce a numeric (float64) value when evaluated:
//   - CILLiteral: A literal numeric value
//   - CILInputRef: A reference to an input metric (assumed numeric; actual type
//     resolution deferred to DAG analysis phase)
//   - CILBinaryOp: A binary arithmetic operation (+, -, *, /) producing a numeric result
//   - CILUnaryOp: A unary arithmetic operation (negate) producing a numeric result
//
// NOTE: In Increment 2, all existing CIL node types are numeric, so this function
// currently returns true for any valid expression tree. This is aspirational
// scaffolding; it will become meaningful when non-numeric node types (e.g.,
// boolean comparisons in Increment 4) are introduced. At that point, this
// function must be updated to return false for non-numeric-producing node types.
//
// Parameters:
//   - node: The CILNode to check for numeric type
//
// Returns:
//   - true if the node is a numeric-producing node
//   - false otherwise
func isNumericNode(node cil.CILNode) (bool, error) {
	switch node.(type) {
	case *cil.CILLiteral, *cil.CILInputRef, *cil.CILBinaryOp, *cil.CILUnaryOp, *cil.CILLookupNode:
		return true, nil
	case *cil.CILComparisonOp, *cil.CILLogicalOp, *cil.CILNotOp,
		*cil.CILBoolLiteral, *cil.CILNilLiteral,
		*cil.CILConditional, *cil.CILWhenClause:
		return false, nil
	default:
		return false, fmt.Errorf("isNumericNode: unknown CILNode type %T", node)
	}
}

// ============================================================================
// Consolidated CIL Tree Validation Walker — Story213
// ============================================================================
//
// validateCILTree replaces the three separate walkers (validateExpressionTypes,
// validateTruthiness, validateBranchCompleteness) with one consolidated
// traversal. See validateCILTreeWalk for the recursive implementation.

// validationState carries contextual flags shared across the consolidated CIL tree
// validation walk. It is passed by pointer through recursive calls so that child
// subtrees can inherit and override validation context.
//
// Fields:
//   - inArithmetic: true when the current subtree is inside a CILBinaryOp or
//     CILUnaryOp. When inArithmetic is true, CILNilLiteral nodes produce a
//     truthiness error because nil is not a valid arithmetic operand. The flag
//     is set to true when recursing into arithmetic operators and reset to
//     false when recursing into conditional conditions. For comparison, logical,
//     and conditional body subtrees, the parent state's inArithmetic value is
//     preserved (passed through unchanged).
type validationState struct {
	inArithmetic bool
}

// validateCILTree is the single entry point for CIL tree validation.
// It replaces the three separate walkers (validateExpressionTypes,
// validateTruthiness, validateBranchCompleteness) with one consolidated
// traversal that applies type, truthiness, and completeness checks per node.
//
// This function initializes a validationState with inArithmetic set to false
// and delegates to validateCILTreeWalk for the recursive traversal.
//
// Parameters:
//   - node: The root CILNode of the expression tree to validate
//
// Returns:
//   - nil if all validation rules pass
//   - error describing the first validation violation encountered, with a
//     category prefix ("type error:", "truthiness error:", or
//     "completeness error:") identifying the rule that failed
//
// Error Cases:
//   - "type error: ..." — arithmetic operand is not numeric
//   - "truthiness error: ..." — nil literal appears in an arithmetic context
//   - "completeness error: ..." — conditional branch is missing or has a nil body
//   - "nil node in CIL tree validation" — the root node or a child node is nil
//   - "unknown node type %T in CIL tree validation" — unrecognized CILNode type
func validateCILTree(node cil.CILNode) error {
	state := &validationState{inArithmetic: false}
	return validateCILTreeWalk(node, state)
}

// validateCILTreeWalk recursively walks the CIL expression tree applying
// type, truthiness, and completeness validation rules at each node.
// It uses a type-switch on cil.CILNode to dispatch to per-node-type
// validation logic, recursing into children with appropriate
// validationState modifications.
//
// Per-node-type rules consolidate the three previous walkers
// (validateExpressionTypes, validateTruthiness, validateBranchCompleteness):
//
//   - CILLiteral, CILInputRef, CILBoolLiteral: leaf nodes, always valid.
//   - CILNilLiteral: returns "truthiness error: ..." if state.inArithmetic is true.
//   - CILBinaryOp, CILUnaryOp: type-check children via isNumericNode; return
//     "type error: ..." for non-numeric operands; recurse with inArithmetic=true.
//   - CILComparisonOp, CILLogicalOp, CILNotOp: recurse children, preserve
//     parent inArithmetic.
//   - CILConditional, CILWhenClause: validate branch/clause completeness;
//     return "completeness error: ..." for missing branches or nil bodies;
//     conditions recurse with inArithmetic=false; bodies inherit parent state.
//
// validationState propagation semantics:
//
//   - Arithmetic operators (CILBinaryOp, CILUnaryOp) create a new state with
//     inArithmetic=true for their children, because nil literals are invalid
//     inside arithmetic expressions.
//   - Conditional conditions (CILConditional branch conditions, CILWhenClause
//     clause conditions) always recurse with inArithmetic=false, because
//     conditions are boolean contexts where nil is permitted.
//   - Conditional bodies and all other children preserve the parent's
//     inArithmetic value, passing the existing state through unchanged.
//
// Validation errors raised by the walker's own per-node checks carry a
// category prefix ("type error:", "truthiness error:", or
// "completeness error:") to distinguish the validation rule that failed.
// Structural errors (e.g., nil node) and errors from helper functions
// such as isNumericNode propagate without modification.
//
// Parameters:
//   - node: The CILNode to validate
//   - state: Validation context carrying flags like inArithmetic
//
// Returns:
//   - nil if the subtree satisfies all validation rules
//   - error describing the first violation
//
// Error Cases:
//   - "nil node in CIL tree validation" — node is nil
//   - "type error: binary operator %s has non-numeric left operand"
//   - "type error: binary operator %s has non-numeric right operand"
//   - "type error: unary operator %s has non-numeric operand"
//   - "truthiness error: nil literal not allowed in arithmetic context"
//   - "completeness error: conditional expression has no branches"
//   - "completeness error: branch %d of conditional expression has nil body"
//   - "completeness error: when clause must have at least one clause"
//   - "completeness error: when clause %d has nil body"
//   - "completeness error: when clause must have a default case"
//   - "unknown node type %T in CIL tree validation" — unrecognized node type
func validateCILTreeWalk(node cil.CILNode, state *validationState) error {
	if node == nil {
		return fmt.Errorf("nil node in CIL tree validation")
	}

	switch n := node.(type) {
	// --- Leaf nodes: always valid (nil OK except in arithmetic context) ---
	case *cil.CILLiteral:
		return nil

	case *cil.CILInputRef:
		return nil

	case *cil.CILLookupNode:
		// Lookup nodes are leaf references that produce numeric values; always valid.
		return nil

	case *cil.CILBoolLiteral:
		return nil

	case *cil.CILNilLiteral:
		if state.inArithmetic {
			return fmt.Errorf("truthiness error: nil literal not allowed in arithmetic context")
		}
		return nil

	// --- Arithmetic operators: type-check children, recurse with inArithmetic=true ---
	case *cil.CILBinaryOp:
		if numeric, err := isNumericNode(n.Left); err != nil {
			return err
		} else if !numeric {
			return fmt.Errorf("type error: binary operator %s has non-numeric left operand", n.Op.Symbol())
		}
		if numeric, err := isNumericNode(n.Right); err != nil {
			return err
		} else if !numeric {
			return fmt.Errorf("type error: binary operator %s has non-numeric right operand", n.Op.Symbol())
		}
		if err := validateCILTreeWalk(n.Left, &validationState{inArithmetic: true}); err != nil {
			return err
		}
		if err := validateCILTreeWalk(n.Right, &validationState{inArithmetic: true}); err != nil {
			return err
		}
		return nil

	case *cil.CILUnaryOp:
		if numeric, err := isNumericNode(n.Operand); err != nil {
			return err
		} else if !numeric {
			return fmt.Errorf("type error: unary operator %s has non-numeric operand", n.Op.Symbol())
		}
		return validateCILTreeWalk(n.Operand, &validationState{inArithmetic: true})

	// --- Comparison / logical operators: recurse children, preserve parent state ---
	case *cil.CILComparisonOp:
		if err := validateCILTreeWalk(n.Left, state); err != nil {
			return err
		}
		if err := validateCILTreeWalk(n.Right, state); err != nil {
			return err
		}
		return nil

	case *cil.CILLogicalOp:
		if err := validateCILTreeWalk(n.Left, state); err != nil {
			return err
		}
		if err := validateCILTreeWalk(n.Right, state); err != nil {
			return err
		}
		return nil

	case *cil.CILNotOp:
		return validateCILTreeWalk(n.Operand, state)

	// --- Conditional: completeness check + condition=false, body=inherit ---
	case *cil.CILConditional:
		if len(n.Branches) == 0 {
			return fmt.Errorf("completeness error: conditional expression has no branches")
		}
		for i, branch := range n.Branches {
			if branch.Body == nil {
				return fmt.Errorf("completeness error: branch %d of conditional expression has nil body", i)
			}
			if branch.Condition != nil {
				if err := validateCILTreeWalk(branch.Condition, &validationState{inArithmetic: false}); err != nil {
					return err
				}
			}
			if err := validateCILTreeWalk(branch.Body, state); err != nil {
				return err
			}
		}
		return nil

	// --- When clause: completeness check + condition=false, body=inherit ---
	case *cil.CILWhenClause:
		if len(n.Clauses) == 0 {
			return fmt.Errorf("completeness error: when clause must have at least one clause")
		}
		for i, clause := range n.Clauses {
			if clause.Body == nil {
				return fmt.Errorf("completeness error: when clause %d has nil body", i)
			}
			if err := validateCILTreeWalk(clause.Condition, &validationState{inArithmetic: false}); err != nil {
				return err
			}
			if err := validateCILTreeWalk(clause.Body, state); err != nil {
				return err
			}
		}
		if n.Default == nil {
			return fmt.Errorf("completeness error: when clause must have a default case")
		}
		if err := validateCILTreeWalk(n.Default, state); err != nil {
			return err
		}
		return nil

	default:
		return fmt.Errorf("unknown node type %T in CIL tree validation", node)
	}
}

// ============================================================================
// Conditional Scope Detection
// ============================================================================

// hasConditionalClauses checks whether a MetricScope contains conditional
// expressions (if/when clauses) by scanning the scope's Definitions for
// clause definition types added during conditional parsing.
//
// When the parser encounters if/when conditional expressions, each branch
// (if-branch, else-if-branch, else-branch, when-clause, default-clause) adds
// a clause definition (IfClauseDefinition, ElseIfClauseDefinition, etc.) to
// the parent MetricScope.Definitions via DefineWithData. This function detects
// these definitions to determine whether the analysis should use the conditional
// assembly path (assembleConditional) instead of the flat postfix path (buildCILTree).
//
// Parameters:
//   - scope: The MetricScope symbol table to check for conditional clauses
//
// Returns:
//   - true if the scope contains any clause definition types
//   - false otherwise
func hasConditionalClauses(scope *symboltable.SymbolTable) bool {
	for _, def := range scope.Definitions {
		switch def.DefinitionType {
		case symboltable.IfClauseDefinition,
			symboltable.ElseIfClauseDefinition,
			symboltable.ElseClauseDefinition,
			symboltable.WhenClauseDefinition,
			symboltable.DefaultClauseDefinition:
			return true
		}
	}
	return false
}

// ============================================================================
// Conditional Assembly Functions — Task 3d
// ============================================================================
//
// These functions implement the conditional assembly algorithm that walks child
// scopes of a MetricScope, pairs clause definitions with their associated body
// scopes, evaluates condition and body postfix sequences, and assembles
// CILConditional or CILWhenClause nodes.
//
// The mapping between body scope types and clause definition types:
//   - IfBodyScope       ←parent→  IfClauseDefinition
//   - ElseIfBodyScope   ←parent→  ElseIfClauseDefinition
//   - ElseBodyScope     ←parent→  ElseClauseDefinition
//   - WhenBodyScope     ←parent→  WhenClauseDefinition
//   - DefaultBodyScope  ←parent→  DefaultClauseDefinition
//
// Story: Story187 - CPL Metrics Conditionals Parse and Analyze
// Task: Task 3d - Implement conditional assembly algorithm

// symbolDefToPostfixElem converts a single SymbolDefinition to the
// parsing.PostfixElement format expected by buildCILTreeFromPostfix.
//
// Conversion rules:
//   - Identifier definitions (Name != nil) → PostfixIdentifier{Name: *Name}
//   - LiteralDefinition → PostfixLiteral{Value: float64 of the int/float64 data}
//   - All other definitions (operator types, boolean/nil literals) → PostfixOperator{DefType: def.DefinitionType}
//
// Parameters:
//   - def: The SymbolDefinition to convert
//   - metricName: The metric name for error messages
//
// Returns:
//   - The converted PostfixElement value suitable for postfix evaluation
//   - An error if the definition type cannot be mapped to a postfix element
func symbolDefToPostfixElem(def symboltable.SymbolDefinition, metricName string) (parsing.PostfixElement, error) {
	switch def.DefinitionType {
	case symboltable.Identifier:
		// Identifier text is stored in Data when pushed by ExitPrimaryExpr
		// (DefineWithData(nil, Identifier, loc, identifierText) sets Name=nil, Data=text).
		// When pushed by EnterIdentifier (non-expression scopes), Name holds the text.
		if def.Name != nil {
			return parsing.PostfixIdentifier{Name: *def.Name}, nil
		}
		if def.Data != nil {
			if sd, ok := def.Data.(symboltable.StringData); ok {
				return parsing.PostfixIdentifier{Name: sd.Value}, nil
			}
		}
		return nil, fmt.Errorf("identifier in metric %q has neither name nor string data", metricName)

	case symboltable.LiteralDefinition:
		if def.Data == nil {
			return nil, fmt.Errorf("literal definition in metric %q has nil data", metricName)
		}
		switch d := def.Data.(type) {
		case symboltable.IntData:
			return parsing.PostfixLiteral{Value: float64(d.Value)}, nil
		case symboltable.FloatData:
			return parsing.PostfixLiteral{Value: d.Value}, nil
		case symboltable.StringData:
			return parsing.PostfixIdentifier{Name: d.Value}, nil
		default:
			return nil, fmt.Errorf("literal definition in metric %q has unexpected data type %T", metricName, def.Data)
		}

	default:
		// All operator definitions and boolean/nil literal definitions
		// use their DefinitionType as the postfix element.
		return parsing.PostfixOperator{DefType: def.DefinitionType}, nil
	}
}

// buildCILTreeFromSymbolDefs constructs a CILMetricNode from a slice of
// SymbolDefinition entries by converting them to the raw postfix format and
// delegating to buildCILTreeFromPostfix.
//
// This is the bridge between the symbol-table-based representation (used by
// ClauseData.ConditionPostfix and body scope Definitions) and the postfix
// evaluator. Each SymbolDefinition is converted via symbolDefToPostfixElem.
//
// Parameters:
//   - defs: The slice of SymbolDefinition entries representing the postfix sequence
//   - metricName: The metric name for error messages and input ref resolution
//   - analyzer: The CPLAnalyzer for resolving input reference kinds
//
// Returns:
//   - A pointer to the constructed CILMetricNode with the expression tree and InputRefs
//   - An error if conversion or postfix evaluation fails
func buildCILTreeFromSymbolDefs(defs []symboltable.SymbolDefinition, metricName string, analyzer *CPLAnalyzer) (*cil.CILMetricNode, error) {
	postfix := make([]parsing.PostfixElement, 0, len(defs))
	for _, def := range defs {
		elem, err := symbolDefToPostfixElem(def, metricName)
		if err != nil {
			return nil, err
		}
		postfix = append(postfix, elem)
	}
	return buildCILTreeFromPostfix(postfix, metricName, analyzer)
}

// assembleConditional processes a MetricScope's children that form a conditional
// expression (if/elseif/else or when/default). It walks the clause definitions
// in the scope's Definitions, extracts ConditionPostfix from ClauseData and
// body postfix from the associated body scope, evaluates both via
// buildCILTreeFromSymbolDefs, and assembles the resulting CILConditional or
// CILWhenClause node as a CILMetricNode with merged InputRefs.
//
// Algorithm:
//
//  1. Scan the scope's Definitions for clause definition types to determine
//     whether this is an if/else chain or a when/default chain.
//  2. For each clause definition in order:
//     a. Extract ClauseData from the definition's Data field.
//     b. Build condition CIL tree from ClauseData.ConditionPostfix (if the
//     clause has a condition — ElseClauseDefinition and
//     DefaultClauseDefinition have no condition).
//     c. Build body CIL tree from ClauseData.BodyScope.Definitions.
//     d. Merge InputRefs from both condition and body trees.
//  3. Assemble the CILConditional (with CILConditionalBranch entries) or
//     CILWhenClause (with CILWhenCase entries and Default field).
//  4. Return a CILMetricNode with the assembled node as CILNode and merged
//     InputRefs.
//
// Scope-to-Clause Mapping:
//
//	IfBodyScope       ←parent→  IfClauseDefinition
//	ElseIfBodyScope   ←parent→  ElseIfClauseDefinition
//	ElseBodyScope     ←parent→  ElseClauseDefinition
//	WhenBodyScope     ←parent→  WhenClauseDefinition
//	DefaultBodyScope  ←parent→  DefaultClauseDefinition
//
// Parameters:
//   - scope: The MetricScope symbol table containing clause definitions
//   - metricName: The metric name for error messages and input ref resolution
//   - analyzer: The CPLAnalyzer for resolving input reference kinds
//
// Returns:
//   - A pointer to the constructed CILMetricNode with the conditional node and merged InputRefs
//   - An error if the conditional structure is malformed or postfix evaluation fails
//
// Error Cases:
//   - "no conditional clause definitions found in scope" — no if/when clauses present
//   - "mixed if and when clause definitions in same scope" — both if and when clauses found
//   - "clause definition missing ClauseData" — a clause definition has nil or wrong-type Data
//   - "clause definition missing BodyScope" — ClauseData has nil BodyScope
//   - Errors from buildCILTreeFromSymbolDefs for condition or body evaluation failures
func assembleConditional(scope *symboltable.SymbolTable, metricName string, analyzer *CPLAnalyzer) (*cil.CILMetricNode, error) {
	// Determine the type of conditional by scanning clause definitions.
	hasIf := false
	hasWhen := false

	for _, def := range scope.Definitions {
		switch def.DefinitionType {
		case symboltable.IfClauseDefinition, symboltable.ElseIfClauseDefinition, symboltable.ElseClauseDefinition:
			hasIf = true
		case symboltable.WhenClauseDefinition, symboltable.DefaultClauseDefinition:
			hasWhen = true
		}
	}

	if hasIf && hasWhen {
		return nil, fmt.Errorf("mixed if and when clause definitions in metric %q", metricName)
	}
	if !hasIf && !hasWhen {
		return nil, fmt.Errorf("no conditional clause definitions found in metric %q", metricName)
	}

	if hasIf {
		return assembleIfConditional(scope, metricName, analyzer)
	}
	return assembleWhenConditional(scope, metricName, analyzer)
}

// assembleIfConditional builds a CILConditional from if/elseif/else clause definitions.
//
// It iterates over the scope's Definitions in order, processing each
// IfClauseDefinition, ElseIfClauseDefinition, and ElseClauseDefinition to
// build CILConditionalBranch entries. An ElseClauseDefinition produces a
// branch with a nil Condition (the default/else case).
//
// Parameters:
//   - scope: The MetricScope symbol table containing clause definitions
//   - metricName: The metric name for error messages
//   - analyzer: The CPLAnalyzer for input ref resolution
//
// Returns:
//   - A CILMetricNode with CILConditional as CILNode and merged InputRefs
//   - An error if clause data is malformed or evaluation fails
func assembleIfConditional(scope *symboltable.SymbolTable, metricName string, analyzer *CPLAnalyzer) (*cil.CILMetricNode, error) {
	var branches []cil.CILConditionalBranch
	var allInputRefs []cil.CILInputRef
	seen := make(map[string]bool)

	for _, def := range scope.Definitions {
		switch def.DefinitionType {
		case symboltable.IfClauseDefinition, symboltable.ElseIfClauseDefinition:
			clauseData, err := extractClauseData(&def, metricName)
			if err != nil {
				return nil, err
			}

			// Build condition CIL tree
			condNode, err := buildCILTreeFromSymbolDefs(clauseData.ConditionPostfix, metricName, analyzer)
			if err != nil {
				return nil, fmt.Errorf("failed to build condition CIL tree in metric %q: %w", metricName, err)
			}

			// Build body CIL tree
			bodyNode, err := buildCILTreeFromSymbolDefs(clauseData.BodyScope.Definitions, metricName, analyzer)
			if err != nil {
				return nil, fmt.Errorf("failed to build body CIL tree in metric %q: %w", metricName, err)
			}

			branches = append(branches, cil.CILConditionalBranch{
				Condition: condNode.CILNode,
				Body:      bodyNode.CILNode,
			})

			// Merge InputRefs
			allInputRefs = mergeInputRefs(allInputRefs, condNode.InputRefs, seen)
			allInputRefs = mergeInputRefs(allInputRefs, bodyNode.InputRefs, seen)

		case symboltable.ElseClauseDefinition:
			clauseData, err := extractClauseData(&def, metricName)
			if err != nil {
				return nil, err
			}

			// Else has no condition
			bodyNode, err := buildCILTreeFromSymbolDefs(clauseData.BodyScope.Definitions, metricName, analyzer)
			if err != nil {
				return nil, fmt.Errorf("failed to build else body CIL tree in metric %q: %w", metricName, err)
			}

			branches = append(branches, cil.CILConditionalBranch{
				Condition: nil, // else has no condition
				Body:      bodyNode.CILNode,
			})

			// Merge InputRefs from else body
			allInputRefs = mergeInputRefs(allInputRefs, bodyNode.InputRefs, seen)
		}
	}

	return &cil.CILMetricNode{
		CILNode: &cil.CILConditional{
			Branches: branches,
		},
		InputRefs: allInputRefs,
	}, nil
}

// assembleWhenConditional builds a CILWhenClause from when/default clause definitions.
//
// It iterates over the scope's Definitions in order, processing each
// WhenClauseDefinition to build CILWhenCase entries, and optionally a
// DefaultClauseDefinition for the default body.
//
// Parameters:
//   - scope: The MetricScope symbol table containing clause definitions
//   - metricName: The metric name for error messages
//   - analyzer: The CPLAnalyzer for input ref resolution
//
// Returns:
//   - A CILMetricNode with CILWhenClause as CILNode and merged InputRefs
//   - An error if clause data is malformed or evaluation fails
func assembleWhenConditional(scope *symboltable.SymbolTable, metricName string, analyzer *CPLAnalyzer) (*cil.CILMetricNode, error) {
	var cases []cil.CILWhenCase
	var defaultBody cil.CILNode
	var allInputRefs []cil.CILInputRef
	seen := make(map[string]bool)

	for _, def := range scope.Definitions {
		switch def.DefinitionType {
		case symboltable.WhenClauseDefinition:
			clauseData, err := extractClauseData(&def, metricName)
			if err != nil {
				return nil, err
			}

			// Build condition CIL tree
			condNode, err := buildCILTreeFromSymbolDefs(clauseData.ConditionPostfix, metricName, analyzer)
			if err != nil {
				return nil, fmt.Errorf("failed to build when condition CIL tree in metric %q: %w", metricName, err)
			}

			// Build body CIL tree
			bodyNode, err := buildCILTreeFromSymbolDefs(clauseData.BodyScope.Definitions, metricName, analyzer)
			if err != nil {
				return nil, fmt.Errorf("failed to build when body CIL tree in metric %q: %w", metricName, err)
			}

			cases = append(cases, cil.CILWhenCase{
				Condition: condNode.CILNode,
				Body:      bodyNode.CILNode,
			})

			// Merge InputRefs
			allInputRefs = mergeInputRefs(allInputRefs, condNode.InputRefs, seen)
			allInputRefs = mergeInputRefs(allInputRefs, bodyNode.InputRefs, seen)

		case symboltable.DefaultClauseDefinition:
			clauseData, err := extractClauseData(&def, metricName)
			if err != nil {
				return nil, err
			}

			// Default has no condition
			bodyNode, err := buildCILTreeFromSymbolDefs(clauseData.BodyScope.Definitions, metricName, analyzer)
			if err != nil {
				return nil, fmt.Errorf("failed to build default body CIL tree in metric %q: %w", metricName, err)
			}

			defaultBody = bodyNode.CILNode
			allInputRefs = mergeInputRefs(allInputRefs, bodyNode.InputRefs, seen)
		}
	}

	return &cil.CILMetricNode{
		CILNode: &cil.CILWhenClause{
			Clauses: cases,
			Default: defaultBody,
		},
		InputRefs: allInputRefs,
	}, nil
}

// extractClauseData extracts *ClauseData from a clause SymbolDefinition's Data field.
//
// Parameters:
//   - def: The SymbolDefinition to extract ClauseData from
//   - metricName: The metric name for error messages
//
// Returns:
//   - The extracted ClauseData
//   - An error if Data is nil, wrong type, or BodyScope is nil
func extractClauseData(def *symboltable.SymbolDefinition, metricName string) (*symboltable.ClauseData, error) {
	if def.Data == nil {
		return nil, fmt.Errorf("clause definition in metric %q has nil Data", metricName)
	}
	// Accept both *ClauseData (from unit tests) and ClauseData (from parser).
	// The parser passes ClauseData by value to DefineWithData, so def.Data
	// holds a symboltable.ClauseData value, not a pointer.
	clauseData, ok := def.Data.(*symboltable.ClauseData)
	if !ok {
		cd, ok2 := def.Data.(symboltable.ClauseData)
		if !ok2 {
			return nil, fmt.Errorf("clause definition in metric %q has unexpected Data type %T", metricName, def.Data)
		}
		clauseData = &cd
	}
	if clauseData.BodyScope == nil {
		return nil, fmt.Errorf("clause definition in metric %q has nil BodyScope", metricName)
	}
	return clauseData, nil
}

// mergeInputRefs merges a slice of CILInputRef entries into an accumulator
// slice, preserving order of first occurrence and deduplicating by name.
//
// Parameters:
//   - acc: The accumulated InputRefs slice (modified in place conceptually, new slice returned)
//   - incoming: The new InputRefs to merge
//   - seen: Map tracking already-seen names (mutated in place)
//
// Returns:
//   - The updated accumulated InputRefs slice
func mergeInputRefs(acc []cil.CILInputRef, incoming []cil.CILInputRef, seen map[string]bool) []cil.CILInputRef {
	for _, ref := range incoming {
		if !seen[ref.Name] {
			seen[ref.Name] = true
			acc = append(acc, ref)
		}
	}
	return acc
}
