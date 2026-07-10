// Copyright (c) 2026 Chronocone Contributors
// SPDX-License-Identifier: MIT

package synthesis

import (
	"fmt"
	"go/parser"
	"strings"
	"testing"

	antlrruntime "github.com/antlr4-go/antlr/v4"
	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	cplparser "github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser/antlr"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// syntaxErrorListener captures syntax errors from the ANTLR lexer/parser.
type syntaxErrorListener struct {
	*antlrruntime.DefaultErrorListener
	errors []string
}

func (s *syntaxErrorListener) SyntaxError(recognizer antlrruntime.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlrruntime.RecognitionException) {
	s.errors = append(s.errors, fmt.Sprintf("line %d:%d %s", line, column, msg))
}

// parseArithmeticAndBuildTree parses an arithmetic expression string and
// builds an analysis.Expression tree from it.
func parseArithmeticAndBuildTree(input string, units map[string]analysis.AnalysisUnitDefinition) (analysis.Expression, error) {
	inputStream := antlrruntime.NewInputStream(input)
	lexer := antlr.NewChronoconePlanningLanguageLexer(inputStream)

	lexer.RemoveErrorListeners()
	lexerErrListener := &syntaxErrorListener{}
	lexer.AddErrorListener(lexerErrListener)

	tokenStream := antlrruntime.NewCommonTokenStream(lexer, antlrruntime.TokenDefaultChannel)
	cplAntlrParser := antlr.NewChronoconePlanningLanguageParser(tokenStream)

	cplAntlrParser.RemoveErrorListeners()
	parserErrListener := &syntaxErrorListener{}
	cplAntlrParser.AddErrorListener(parserErrListener)

	listener := cplparser.NewCPLListener(nil, lexer, cplAntlrParser)
	stack := listener.GetSymbolStack()
	global, _ := stack.Peek()
	tempScope := global.AddChild(symboltable.TempScope)
	stack.Push(tempScope)

	tree := cplAntlrParser.ArithmeticExpr()
	antlrruntime.ParseTreeWalkerDefault.Walk(listener, tree)

	var allErrors []string
	allErrors = append(allErrors, lexerErrListener.errors...)
	allErrors = append(allErrors, parserErrListener.errors...)
	for _, err := range listener.GetErrors() {
		allErrors = append(allErrors, err.Error())
	}
	if len(allErrors) > 0 {
		return nil, fmt.Errorf("parse errors: %s", strings.Join(allErrors, "; "))
	}

	return analysis.BuildExpressionTree(tempScope.Definitions, units, nil)
}

// defaultUnits returns a units map suitable for testing.
func defaultUnits() map[string]analysis.AnalysisUnitDefinition {
	return map[string]analysis.AnalysisUnitDefinition{
		"Length": {Name: "Length", ConceptName: "DistanceUnits"},
		"Time":   {Name: "Time", ConceptName: "TimeUnits"},
		"Mass":   {Name: "Mass", ConceptName: "MassUnits"},
	}
}

// runPipeline executes the full parse -> analyze -> convert -> synthesize pipeline
// for a single expression string and returns the generated Go code.
func runPipeline(t *testing.T, expr string) string {
	t.Helper()

	exprTree, err := parseArithmeticAndBuildTree(expr, defaultUnits())
	require.NoError(t, err, "failed to parse/build expression %q", expr)

	cilNode, err := analysis.ExpressionToCILNode(exprTree)
	require.NoError(t, err, "failed to convert expression %q to CIL", expr)

	code, err := SynthesizeExpression(cilNode)
	require.NoError(t, err, "failed to synthesize expression %q", expr)
	require.NotEmpty(t, code, "synthesized code for %q is empty", expr)

	return code
}

func TestPipeline_SimpleUnitRef(t *testing.T) {
	t.Parallel()

	code := runPipeline(t, "Length")
	assert.Equal(t, "Length", code, "expected synthesized code to match")
}

func TestPipeline_Multiplication(t *testing.T) {
	t.Parallel()

	code := runPipeline(t, "Length*Time")
	assert.Equal(t, "(Length * Time)", code, "expected synthesized code to match")
}

func TestPipeline_Division(t *testing.T) {
	t.Parallel()

	code := runPipeline(t, "Length/Time")
	assert.Equal(t, "(Length / Time)", code, "expected synthesized code to match")
}

func TestPipeline_Power(t *testing.T) {
	t.Parallel()

	code := runPipeline(t, "Length^2")
	assert.Equal(t, "(Length * Length)", code, "expected synthesized code to match")
}

func TestPipeline_Addition(t *testing.T) {
	t.Parallel()

	code := runPipeline(t, "Length+Time")
	assert.Equal(t, "(Length + Time)", code, "expected synthesized code to match")
}

func TestPipeline_Subtraction(t *testing.T) {
	t.Parallel()

	code := runPipeline(t, "Length-Time")
	assert.Equal(t, "(Length - Time)", code, "expected synthesized code to match")
}

func TestPipeline_ComplexExpression(t *testing.T) {
	t.Parallel()

	code := runPipeline(t, "Length/Time^2")
	assert.Equal(t, "(Length / (Time * Time))", code, "expected synthesized code to match")
}

func TestPipeline_NumberLiteral(t *testing.T) {
	t.Parallel()

	code := runPipeline(t, "5*Length")
	assert.Equal(t, "(5 * Length)", code, "expected synthesized code to match")
}

func TestPipeline_AllOutputsCompilable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		expr string
	}{
		{"simple unit ref", "Length"},
		{"multiplication", "Length*Time"},
		{"division", "Length/Time"},
		{"power", "Length^2"},
		{"addition", "Length+Time"},
		{"subtraction", "Length-Time"},
		{"complex expression", "Length/Time^2"},
		{"number literal", "5*Length"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			code := runPipeline(t, tt.expr)
			_, err := parser.ParseExpr(code)
			assert.NoError(t, err, "synthesized code for %q does not parse\nCode: %s", tt.expr, code)
		})
	}
}

// TestMeasureUnitExpression_SimpleUnitRef verifies that a simple unit reference
// measure expression synthesizes to the unit name directly.
func TestMeasureUnitExpression_SimpleUnitRef(t *testing.T) {
	t.Parallel()

	expr := &analysis.UnitReference{Name: "Length"}
	cilNode, err := analysis.ExpressionToCILNode(expr)
	require.NoError(t, err, "failed to convert unit reference to CIL")

	code, err := SynthesizeExpression(cilNode)
	require.NoError(t, err, "failed to synthesize unit reference")
	assert.Equal(t, "Length", code, "expected synthesized code to match")
}

// TestMeasureUnitExpression_CompoundDivision verifies that a division of two
// unit references synthesizes to a parenthesized Go expression.
func TestMeasureUnitExpression_CompoundDivision(t *testing.T) {
	t.Parallel()

	expr := &analysis.DivisionNode{
		Left:  &analysis.UnitReference{Name: "Length"},
		Right: &analysis.UnitReference{Name: "Time"},
	}
	cilNode, err := analysis.ExpressionToCILNode(expr)
	require.NoError(t, err, "failed to convert division expression to CIL")

	code, err := SynthesizeExpression(cilNode)
	require.NoError(t, err, "failed to synthesize division expression")
	assert.Equal(t, "(Length / Time)", code, "expected synthesized code to match")
}

// TestMeasureUnitExpression_DivisionWithPower verifies that a division with a
// power in the denominator synthesizes correctly using multiplication.
func TestMeasureUnitExpression_DivisionWithPower(t *testing.T) {
	t.Parallel()

	expr := &analysis.DivisionNode{
		Left: &analysis.UnitReference{Name: "Length"},
		Right: &analysis.PowerNode{
			Base:     &analysis.UnitReference{Name: "Time"},
			Exponent: 2,
		},
	}
	cilNode, err := analysis.ExpressionToCILNode(expr)
	require.NoError(t, err, "failed to convert power division expression to CIL")

	code, err := SynthesizeExpression(cilNode)
	require.NoError(t, err, "failed to synthesize power division expression")
	assert.Equal(t, "(Length / (Time * Time))", code, "expected synthesized code to match")
}
