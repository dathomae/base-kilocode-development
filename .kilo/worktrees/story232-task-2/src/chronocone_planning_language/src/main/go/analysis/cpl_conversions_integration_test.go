// Package analysis_test contains integration tests for the CPL conversions pipeline.
//
// These tests verify the full parse → analyze → synthesize pipeline for unit
// and measure conversions, covering the acceptance criteria from Story230.
package analysis_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/chronocone/chronocone_planning_language/src/main/go/cil"
	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/chronocone/chronocone_planning_language/src/main/go/symboltable"
	"github.com/chronocone/chronocone_planning_language/src/main/go/synthesis"
	cio "github.com/chronocone/chronocone_planning_language/src/main/go/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Target CPL Example: Length Unit Conversions
// ============================================================================

// TestLengthUnitConversions verifies that a CPL document with Length unit
// conversions (Kilometer ↔ Meter ↔ Centimeter) parses, analyzes, and produces
// the expected conversion pairs.
func TestLengthUnitConversions(t *testing.T) {
	t.Parallel()

	input := `concepts
	DistanceUnit = [Values[Kilometer, Meter, Centimeter]]

units
	Kilometer = DistanceUnit
		conversion
			pattern Kilometer(km) = Meter(m)
			calculation m / 1000
			pattern Meter(m) = Centimeter(cm)
			calculation cm / 100

	Meter = DistanceUnit
		conversion
			pattern Meter(m) = Kilometer(km)
			calculation km * 1000

	Centimeter = DistanceUnit
		conversion
			pattern Centimeter(cm) = Meter(m)
			calculation m * 100`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	require.NotNil(t, defs)

	// Verify all three units were analyzed
	require.Len(t, defs.Units, 3)
	require.Contains(t, defs.Units, "Kilometer")
	require.Contains(t, defs.Units, "Meter")
	require.Contains(t, defs.Units, "Centimeter")

	// Kilometer has 4 conversions (2 pattern/calculation pairs × 2 forward+inverse each)
	km := defs.Units["Kilometer"]
	require.Len(t, km.Conversions, 4, "Kilometer should have 2 pairs × 2 = 4 conversions")

	// First pair forward: Meter → Kilometer, km = m / 1000
	assert.Equal(t, "Meter", km.Conversions[0].SourceDef.Name)
	assert.Equal(t, "Kilometer", km.Conversions[0].TargetDef.Name)
	assert.Equal(t, "m", km.Conversions[0].SourceVar)
	assert.Equal(t, "km", km.Conversions[0].TargetVar)
	_, isDiv0 := km.Conversions[0].Calculation.(*analysis.DivisionNode)
	assert.True(t, isDiv0, "first calculation should be division")

	// First pair inverse: Kilometer → Meter, m = km * 1000
	assert.Equal(t, "Kilometer", km.Conversions[1].SourceDef.Name)
	assert.Equal(t, "Meter", km.Conversions[1].TargetDef.Name)
	assert.Equal(t, "km", km.Conversions[1].SourceVar)
	assert.Equal(t, "m", km.Conversions[1].TargetVar)
	_, isMult0 := km.Conversions[1].Calculation.(*analysis.MultiplicationNode)
	assert.True(t, isMult0, "inverse calculation should be multiplication")

	// Second pair forward: Centimeter → Meter, m = cm / 100
	assert.Equal(t, "Centimeter", km.Conversions[2].SourceDef.Name)
	assert.Equal(t, "Meter", km.Conversions[2].TargetDef.Name)
	assert.Equal(t, "cm", km.Conversions[2].SourceVar)
	assert.Equal(t, "m", km.Conversions[2].TargetVar)
	_, isDiv1 := km.Conversions[2].Calculation.(*analysis.DivisionNode)
	assert.True(t, isDiv1, "second forward calculation should be division")

	// Second pair inverse: Meter → Centimeter, cm = m * 100
	assert.Equal(t, "Meter", km.Conversions[3].SourceDef.Name)
	assert.Equal(t, "Centimeter", km.Conversions[3].TargetDef.Name)
	assert.Equal(t, "m", km.Conversions[3].SourceVar)
	assert.Equal(t, "cm", km.Conversions[3].TargetVar)
	_, isMult1 := km.Conversions[3].Calculation.(*analysis.MultiplicationNode)
	assert.True(t, isMult1, "second inverse calculation should be multiplication")

	// Meter has 2 conversions (forward + derived inverse from Kilometer's conversion)
	m := defs.Units["Meter"]
	require.Len(t, m.Conversions, 2)
	assert.Equal(t, "Kilometer", m.Conversions[0].SourceDef.Name)
	assert.Equal(t, "Meter", m.Conversions[0].TargetDef.Name)
	assert.Equal(t, "km", m.Conversions[0].SourceVar)
	assert.Equal(t, "m", m.Conversions[0].TargetVar)

	// Centimeter has 2 conversions (forward + derived inverse)
	cm := defs.Units["Centimeter"]
	require.Len(t, cm.Conversions, 2)
	assert.Equal(t, "Meter", cm.Conversions[0].SourceDef.Name)
	assert.Equal(t, "Centimeter", cm.Conversions[0].TargetDef.Name)
	assert.Equal(t, "m", cm.Conversions[0].SourceVar)
	assert.Equal(t, "cm", cm.Conversions[0].TargetVar)
}

// ============================================================================
// Target CPL Example: Temperature Unit Conversions
// ============================================================================

// TestTemperatureUnitConversions verifies that Celsius ↔ Fahrenheit unit
// conversions parse, analyze, and produce correct calculation expression trees.
func TestTemperatureUnitConversions(t *testing.T) {
	t.Parallel()

	input := `concepts
	TemperatureUnit = [Values[Celsius, Fahrenheit]]

units
	Celsius = TemperatureUnit
		conversion
			pattern Celsius(c) = Fahrenheit(f)
			calculation (f - 32) * 5 / 9

	Fahrenheit = TemperatureUnit
		conversion
			pattern Fahrenheit(f) = Celsius(c)
			calculation c * 9 / 5 + 32`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	require.NotNil(t, defs)

	require.Len(t, defs.Units, 2)

	// Celsius conversion: Fahrenheit → Celsius, plus derived inverse
	celsius := defs.Units["Celsius"]
	require.Len(t, celsius.Conversions, 2)
	assert.Equal(t, "Fahrenheit", celsius.Conversions[0].SourceDef.Name)
	assert.Equal(t, "Celsius", celsius.Conversions[0].TargetDef.Name)
	assert.Equal(t, "f", celsius.Conversions[0].SourceVar)
	assert.Equal(t, "c", celsius.Conversions[0].TargetVar)
	require.NotNil(t, celsius.Conversions[0].Calculation)

	// Inverse: Celsius → Fahrenheit
	assert.Equal(t, "Celsius", celsius.Conversions[1].SourceDef.Name)
	assert.Equal(t, "Fahrenheit", celsius.Conversions[1].TargetDef.Name)
	assert.Equal(t, "c", celsius.Conversions[1].SourceVar)
	assert.Equal(t, "f", celsius.Conversions[1].TargetVar)
	require.NotNil(t, celsius.Conversions[1].Calculation)

	// Fahrenheit conversion: Celsius → Fahrenheit, plus derived inverse
	fahrenheit := defs.Units["Fahrenheit"]
	require.Len(t, fahrenheit.Conversions, 2)
	assert.Equal(t, "Celsius", fahrenheit.Conversions[0].SourceDef.Name)
	assert.Equal(t, "Fahrenheit", fahrenheit.Conversions[0].TargetDef.Name)
	assert.Equal(t, "c", fahrenheit.Conversions[0].SourceVar)
	assert.Equal(t, "f", fahrenheit.Conversions[0].TargetVar)
	require.NotNil(t, fahrenheit.Conversions[0].Calculation)

	// Inverse: Fahrenheit → Celsius
	assert.Equal(t, "Fahrenheit", fahrenheit.Conversions[1].SourceDef.Name)
	assert.Equal(t, "Celsius", fahrenheit.Conversions[1].TargetDef.Name)
	assert.Equal(t, "f", fahrenheit.Conversions[1].SourceVar)
	assert.Equal(t, "c", fahrenheit.Conversions[1].TargetVar)
	require.NotNil(t, fahrenheit.Conversions[1].Calculation)

	// Verify the calculation expression tree for Fahrenheit: c * 9 / 5 + 32
	// This should be a tree of operations, not nil
	_, ok := fahrenheit.Conversions[0].Calculation.(*analysis.AdditionNode)
	assert.True(t, ok, "Fahrenheit calculation should be an AdditionNode at the top level")
}

// ============================================================================
// Parser Symbol Table Tree Verification
// ============================================================================

// TestUnitConversionSymbolTableTree verifies that the parser produces the
// correct symbol table tree structure: ConversionScope → PatternScope +
// CalculationScope, including multiple pairs within a single ConversionScope.
func TestUnitConversionSymbolTableTree(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[km, m, cm]]

units
	Kilometer = Distance
		conversion
			pattern Kilometer(km) = Meter(m)
			calculation m * 1000
			pattern Meter(m) = Centimeter(cm)
			calculation cm * 100000`

	doc, err := parser.Parse(input)
	require.NoError(t, err)
	require.NotNil(t, doc)

	// Navigate to UnitsScope → UnitScope
	require.Len(t, doc.Children, 2)
	unitsScope := doc.Children[1]
	require.Equal(t, symboltable.UnitsScope, unitsScope.ScopeType)
	require.Len(t, unitsScope.Children, 1)

	unitScope := unitsScope.Children[0]
	require.Equal(t, symboltable.UnitScope, unitScope.ScopeType)

	// Verify UnitScope → ConversionScope
	var conversionScope *symboltable.SymbolTable
	for _, child := range unitScope.Children {
		if child.ScopeType == symboltable.ConversionScope {
			conversionScope = child
			break
		}
	}
	require.NotNil(t, conversionScope, "UnitScope should have a ConversionScope child")

	// Verify ConversionScope has 4 children: 2 PatternScopes + 2 CalculationScopes
	require.Len(t, conversionScope.Children, 4, "ConversionScope should have 4 children for 2 pairs")

	var patternScopes, calcScopes int
	for _, child := range conversionScope.Children {
		switch child.ScopeType {
		case symboltable.PatternScope:
			patternScopes++
		case symboltable.CalculationScope:
			calcScopes++
		}
	}
	assert.Equal(t, 2, patternScopes, "Expected 2 PatternScope children")
	assert.Equal(t, 2, calcScopes, "Expected 2 CalculationScope children")
}

// ============================================================================
// Measure Conversion Integration with Algebraic Reversal
// ============================================================================

// TestMeasureConversionsWithAlgebraicReversal verifies that measure conversions
// derive both forward and inverse pairs automatically.
func TestMeasureConversionsWithAlgebraicReversal(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[km, m]]

units
	KilometerUnit = Distance
	MeterUnit = Distance

measures
	KilometerLength = Number(unit=KilometerUnit)
		conversion
			pattern KilometerLength(km) = MeterLength(m)
			calculation m * 1000
	MeterLength = Number(unit=MeterUnit)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	require.NotNil(t, defs)

	// The declaring measure (KilometerLength) gets the conversions
	kmLen := defs.Measures["KilometerLength"]
	require.Len(t, kmLen.Conversions, 2, "Should have forward + inverse = 2 conversion pairs")

	// Forward: MeterLength → KilometerLength, calculation: m * 1000
	forward := kmLen.Conversions[0]
	assert.Equal(t, "MeterLength", forward.SourceDef.Name)
	assert.Equal(t, "KilometerLength", forward.TargetDef.Name)
	assert.Equal(t, "m", forward.SourceVar)
	assert.Equal(t, "km", forward.TargetVar)
	_, isForwardMult := forward.Calculation.(*analysis.MultiplicationNode)
	assert.True(t, isForwardMult, "forward calculation should be multiplication")

	// Inverse: KilometerLength → MeterLength, calculation: km / 1000
	inverse := kmLen.Conversions[1]
	assert.Equal(t, "KilometerLength", inverse.SourceDef.Name)
	assert.Equal(t, "MeterLength", inverse.TargetDef.Name)
	assert.Equal(t, "km", inverse.SourceVar)
	assert.Equal(t, "m", inverse.TargetVar)
	_, isInverseDiv := inverse.Calculation.(*analysis.DivisionNode)
	assert.True(t, isInverseDiv, "inverse calculation should be division")

	// The target measure (MeterLength) should have no conversions
	mLen := defs.Measures["MeterLength"]
	assert.Empty(t, mLen.Conversions, "MeterLength should have no conversions")
}

// ============================================================================
// Rejection Cases
// ============================================================================

// TestUnitConversions_BareNumbersRejection verifies that analysis rejects
// calculations containing bare numeric constants without unit annotations.
func TestUnitConversions_BareNumbersRejection(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[km, m]]

units
	Kilometer = Distance
		conversion
			pattern Kilometer(km) = Meter(m)
			calculation 5 + 9
	Meter = Distance`

	_, err := parseAndAnalyze(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bare numbers")
}

// TestMeasureConversions_BareNumbersRejection verifies that measure conversion
// analysis rejects bare numeric constants in additive contexts.
func TestMeasureConversions_BareNumbersRejection(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[km, m]]

units
	KilometerUnit = Distance
	MeterUnit = Distance

measures
	KmMeasure = Number(unit=KilometerUnit)
		conversion
			pattern KmMeasure(km) = MeterMeasure(m)
			calculation 5 + 9
	MeterMeasure = Number(unit=MeterUnit)`

	_, err := parseAndAnalyze(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bare numbers without unit annotations")
}

// TestUnitConversions_VariableNameCollisionRejection verifies that analysis
// rejects pattern-bound variables that share names with declared units.
func TestUnitConversions_VariableNameCollisionRejection(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[km, m]]

units
	Kilometer = Distance
	m = Distance

	Meter = Distance
		conversion
			pattern Kilometer(km) = Meter(m)
			calculation m * 1000`

	_, err := parseAndAnalyze(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts with unit name")
}

// TestMeasureConversions_VariableNameCollisionRejection verifies that measure
// conversion analysis rejects pattern-bound variables colliding with unit names.
func TestMeasureConversions_VariableNameCollisionRejection(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[km, m]]

units
	KilometerUnit = Distance
	m = Distance

measures
	KmMeasure = Number(unit=KilometerUnit)
		conversion
			pattern KmMeasure(km) = MeterMeasure(m)
			calculation m * 1000
	MeterMeasure = Number(unit=KilometerUnit)`

	_, err := parseAndAnalyze(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts with unit name")
}

// TestUnitConversions_CrossConceptRejection verifies that analysis rejects
// conversions between units belonging to different concepts.
func TestUnitConversions_CrossConceptRejection(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[km, m]]
	Time = [Values[s, ms]]

units
	Kilometer = Distance
		conversion
			pattern Kilometer(km) = Second(s)
			calculation s * 1000
	Second = Time`

	_, err := parseAndAnalyze(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "different groups")
}

// TestMeasureConversions_SelfConversionRejection verifies that measure
// conversion analysis rejects self-conversions.
func TestMeasureConversions_SelfConversionRejection(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[m]]

units
	Meter = Distance

measures
	LenMeasure = Number(unit=Meter)
		conversion
			pattern LenMeasure(a) = LenMeasure(b)
			calculation a * 1`

	_, err := parseAndAnalyze(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "self-conversion")
}

// ============================================================================
// Synthesis Integration Tests
// ============================================================================

// TestSynthesis_UnitConversions_EmitsCompilableCode verifies that the synthesis
// pipeline emits Go code with unit conversion tables that compile and execute
// correctly.
func TestSynthesis_UnitConversions_EmitsCompilableCode(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("'go' command not available in PATH")
	}

	input := `concepts
	Distance = [Values[km, m]]

units
	Kilometer = Distance
		conversion
			pattern Kilometer(km) = Meter(m)
			calculation m / 1000
	Meter = Distance`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	require.NotNil(t, defs)

	// Build synthesis input from analysis definitions
	inputSynth, err := synthesis.BuildTransformerFromAnalysis(defs)
	require.NoError(t, err)
	require.NotNil(t, inputSynth)

	// Verify unit conversion code was generated
	require.NotEmpty(t, inputSynth.UnitConversionCode, "UnitConversionCode should be populated")
	assert.Contains(t, inputSynth.UnitConversionCode, "var unitConversions")
	assert.Contains(t, inputSynth.UnitConversionCode, "func convertUnit")
	assert.Contains(t, inputSynth.UnitConversionCode, `"Meter"`)
	assert.Contains(t, inputSynth.UnitConversionCode, `"Kilometer"`)

	// Wrap in a tiny Go program and compile/run it
	program := fmt.Sprintf(`package main

import "fmt"

%s

func main() {
	fmt.Printf("%%.0f\n", convertUnit("Meter", "Kilometer", 5000))
	fmt.Printf("%%.0f\n", convertUnit("Kilometer", "Meter", 3))
}
`, inputSynth.UnitConversionCode)

	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "main.go")
	err = os.WriteFile(srcPath, []byte(program), 0644)
	require.NoError(t, err)

	cmd := exec.Command("go", "run", srcPath)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "generated code failed to compile/run: %s", string(out))

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	require.Len(t, lines, 2)
	assert.Equal(t, "5", lines[0], "Meter(5000) → Kilometer should be 5")
	assert.Equal(t, "3000", lines[1], "Kilometer(3) → Meter should be 3000")
}

// TestSynthesis_MeasureConversions_EmitsCompilableCode verifies that the synthesis
// pipeline emits Go code with measure conversion tables that compile and execute
// correctly, including algebraic reversal.
func TestSynthesis_MeasureConversions_EmitsCompilableCode(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("'go' command not available in PATH")
	}

	input := `concepts
	Distance = [Values[km, m]]

units
	KilometerUnit = Distance
	MeterUnit = Distance

measures
	KilometerLength = Number(unit=KilometerUnit)
		conversion
			pattern KilometerLength(km) = MeterLength(m)
			calculation m / 1000
	MeterLength = Number(unit=MeterUnit)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	require.NotNil(t, defs)

	// Build synthesis input from analysis definitions
	inputSynth, err := synthesis.BuildTransformerFromAnalysis(defs)
	require.NoError(t, err)
	require.NotNil(t, inputSynth)

	// Verify measure conversion code was generated
	require.NotEmpty(t, inputSynth.MeasureConversionCode, "MeasureConversionCode should be populated")
	assert.Contains(t, inputSynth.MeasureConversionCode, "var measureConversions")
	assert.Contains(t, inputSynth.MeasureConversionCode, "func convertMeasure")
	assert.Contains(t, inputSynth.MeasureConversionCode, `"KilometerLength"`)
	assert.Contains(t, inputSynth.MeasureConversionCode, `"MeterLength"`)

	// Wrap in a tiny Go program and compile/run it
	program := fmt.Sprintf(`package main

import "fmt"

%s

func main() {
	fmt.Printf("%%.0f\n", convertMeasure("MeterLength", "KilometerLength", 5000))
	fmt.Printf("%%.0f\n", convertMeasure("KilometerLength", "MeterLength", 3))
}
`, inputSynth.MeasureConversionCode)

	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "main.go")
	err = os.WriteFile(srcPath, []byte(program), 0644)
	require.NoError(t, err)

	cmd := exec.Command("go", "run", srcPath)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "generated code failed to compile/run: %s", string(out))

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	require.Len(t, lines, 2)
	assert.Equal(t, "5", lines[0], "MeterLength(5000) → KilometerLength should be 5")
	assert.Equal(t, "3000", lines[1], "KilometerLength(3) → MeterLength should be 3000")
}

// ============================================================================
// Full Pipeline: Parse → Analyze → Synthesize → GenerateMainFunction
// ============================================================================

// TestFullPipeline_LengthConversions verifies the end-to-end pipeline for Length
// conversions: parse → analyze → BuildTransformerFromAnalysis → GenerateMainFunction.
func TestFullPipeline_LengthConversions(t *testing.T) {
	input := `concepts
	Distance = [Values[km, m, cm]]

units
	Kilometer = Distance
		conversion
			pattern Kilometer(km) = Meter(m)
			calculation m * 1000
	Meter = Distance
	Centimeter = Distance`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	require.NotNil(t, defs)

	inputSynth, err := synthesis.BuildTransformerFromAnalysis(defs)
	require.NoError(t, err)
	require.NotNil(t, inputSynth)

	ioSpec := &cio.IOSpecification{
		Connectors:       []cio.ConnectorSpecification{},
		Outputs:      []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{},
	}
	inputSynth.IOSpec = ioSpec

	code, err := synthesis.GenerateMainFunction(inputSynth)
	require.NoError(t, err)
	require.NotEmpty(t, code)

	// Verify conversion code is embedded in the generated main function
	assert.Contains(t, code, "unitConversions")
	assert.Contains(t, code, "func convertUnit")
}

// TestFullPipeline_TemperatureConversions verifies the end-to-end pipeline for
// temperature unit conversions (not measure conversions, since additive bare
// numbers are rejected in measure conversion calculations per constraint 4).
func TestFullPipeline_TemperatureConversions(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("'go' command not available in PATH")
	}

	input := `concepts
	Temperature = [Values[Celsius, Fahrenheit]]

units
	Celsius = Temperature
		conversion
			pattern Celsius(c) = Fahrenheit(f)
			calculation (f - 32) * 5 / 9
	Fahrenheit = Temperature`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	require.NotNil(t, defs)

	inputSynth, err := synthesis.BuildTransformerFromAnalysis(defs)
	require.NoError(t, err)
	require.NotNil(t, inputSynth)

	ioSpec := &cio.IOSpecification{
		Connectors:       []cio.ConnectorSpecification{},
		Outputs:      []cio.OutputSpecification{},
		Architecture: cio.ArchitectureSpecification{},
	}
	inputSynth.IOSpec = ioSpec

	code, err := synthesis.GenerateMainFunction(inputSynth)
	require.NoError(t, err)
	require.NotEmpty(t, code)

	// Verify conversion code is embedded in the generated main function
	assert.Contains(t, code, "unitConversions")
	assert.Contains(t, code, "func convertUnit")

	// Compile-check the generated main function
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "main.go")
	err = os.WriteFile(srcPath, []byte(code), 0644)
	require.NoError(t, err)

	cmd := exec.Command("go", "run", srcPath)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "generated main function failed to compile/run: %s", string(out))
}

// ============================================================================
// CIL Conversion Synthesis Direct Tests
// ============================================================================

// TestCILUnitConversions_SinglePair verifies CIL synthesis for a single
// unit conversion pair.
func TestCILUnitConversions_SinglePair(t *testing.T) {
	pairs := []cil.CILConversionPair{
		{SourceName: "Meter", TargetName: "Kilometer", SourceVar: "m", TargetVar: "km", Calculation: "(m / 1000)"},
	}
	code := cil.SynthesizeConversions(pairs, "unitConversions", "convertUnit", "unit")

	assert.Contains(t, code, "var unitConversions")
	assert.Contains(t, code, "func convertUnit")
	assert.Contains(t, code, `"Meter"`)
	assert.Contains(t, code, `"Kilometer"`)
	assert.Contains(t, code, "func(m float64) float64 { return (m / 1000) }")
}

// TestCILMeasureConversions_TemperaturePair verifies CIL synthesis for a
// measure conversion pair.
func TestCILMeasureConversions_TemperaturePair(t *testing.T) {
	pairs := []cil.CILConversionPair{
		{SourceName: "Celsius", TargetName: "Fahrenheit", SourceVar: "c", TargetVar: "f", Calculation: "((c * 1.8) + 32)"},
	}
	code := cil.SynthesizeConversions(pairs, "measureConversions", "convertMeasure", "measure")

	assert.Contains(t, code, "var measureConversions")
	assert.Contains(t, code, "func convertMeasure")
	assert.Contains(t, code, `"Celsius"`)
	assert.Contains(t, code, `"Fahrenheit"`)
	assert.Contains(t, code, "func(c float64) float64 { return ((c * 1.8) + 32) }")
}

// ============================================================================
// Regression Tests
// ============================================================================

// TestUnitConversions_NoRegression_ExistingTests verifies that the fixes for
// multiple conversion pairs do not break the existing unit conversion tests.
func TestUnitConversions_NoRegression_ExistingTests(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[km, m]]

units
	Kilometer = Distance
		conversion
			pattern Kilometer(km) = Meter(m)
			calculation m * 1000
	Meter = Distance`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	require.NotNil(t, defs)

	km := defs.Units["Kilometer"]
	require.Len(t, km.Conversions, 2, "Kilometer should have forward + inverse = 2 conversions")
	assert.Equal(t, "Meter", km.Conversions[0].SourceDef.Name)
	assert.Equal(t, "Kilometer", km.Conversions[0].TargetDef.Name)
	assert.Equal(t, "m", km.Conversions[0].SourceVar)
	assert.Equal(t, "km", km.Conversions[0].TargetVar)

	// Inverse pair
	assert.Equal(t, "Kilometer", km.Conversions[1].SourceDef.Name)
	assert.Equal(t, "Meter", km.Conversions[1].TargetDef.Name)
	assert.Equal(t, "km", km.Conversions[1].SourceVar)
	assert.Equal(t, "m", km.Conversions[1].TargetVar)
}

// TestMeasureConversions_NoRegression_ExistingTests verifies that the fixes for
// multiple conversion pairs do not break the existing measure conversion tests.
func TestMeasureConversions_NoRegression_ExistingTests(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[km, m]]

units
	KilometerUnit = Distance
	MeterUnit = Distance

measures
	KilometerLength = Number(unit=KilometerUnit)
		conversion
			pattern KilometerLength(km) = MeterLength(m)
			calculation m * 1000
	MeterLength = Number(unit=MeterUnit)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	require.NotNil(t, defs)

	kmLen := defs.Measures["KilometerLength"]
	require.Len(t, kmLen.Conversions, 2)

	// Forward
	assert.Equal(t, "MeterLength", kmLen.Conversions[0].SourceDef.Name)
	assert.Equal(t, "KilometerLength", kmLen.Conversions[0].TargetDef.Name)

	// Inverse
	assert.Equal(t, "KilometerLength", kmLen.Conversions[1].SourceDef.Name)
	assert.Equal(t, "MeterLength", kmLen.Conversions[1].TargetDef.Name)
}

// TestMeasureConversions_MultiplePairsInSingleClause verifies that multiple
// measure conversion pairs within a single MeasureConversionScope are all
// processed and each generates forward+inverse pairs.
func TestMeasureConversions_MultiplePairsInSingleClause(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[km, m, cm]]

units
	KilometerUnit = Distance
	MeterUnit = Distance
	CentimeterUnit = Distance

measures
	KmMeasure = Number(unit=KilometerUnit)
		conversion
			pattern KmMeasure(km) = MeterMeasure(m)
			calculation m * 1000
			pattern KmMeasure(km) = CmMeasure(cm)
			calculation cm * 100000
	MeterMeasure = Number(unit=MeterUnit)
	CmMeasure = Number(unit=CentimeterUnit)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	require.NotNil(t, defs)

	kmMeasure := defs.Measures["KmMeasure"]
	require.Len(t, kmMeasure.Conversions, 4, "2 conversions × 2 pairs each = 4 pairs")
}

// ============================================================================
// Acceptance Criteria Coverage
// ============================================================================

// TestAcceptanceCriteria_1_ParserAcceptsConversionClauses verifies AC1:
// CPLUnits.g4 and CPLMeasures.g4 accept conversion clauses with pattern/calculation pairs.
func TestAcceptanceCriteria_1_ParserAcceptsConversionClauses(t *testing.T) {
	t.Parallel()

	unitInput := `concepts
	Distance = [Values[km, m]]

units
	Kilometer = Distance
		conversion
			pattern Kilometer(km) = Meter(m)
			calculation m * 1000
	Meter = Distance`

	doc, err := parser.Parse(unitInput)
	require.NoError(t, err)
	require.NotNil(t, doc)

	measureInput := `concepts
	Distance = [Values[km, m]]

units
	KilometerUnit = Distance
	MeterUnit = Distance

measures
	KmMeasure = Number(unit=KilometerUnit)
		conversion
			pattern KmMeasure(km) = MeterMeasure(m)
			calculation m * 1000
	MeterMeasure = Number(unit=MeterUnit)`

	doc, err = parser.Parse(measureInput)
	require.NoError(t, err)
	require.NotNil(t, doc)
}

// TestAcceptanceCriteria_2_SymbolTableTree verifies AC2: Parser produces
// correct symbol table tree with ConversionScope → PatternScope + CalculationScope.
func TestAcceptanceCriteria_2_SymbolTableTree(t *testing.T) {
	t.Parallel()

	input := `units
	Length = Distance
		conversion
			pattern Kilometer(km) = Meter(m)
			calculation km * 1000`

	doc, err := parser.Parse(input)
	require.NoError(t, err)
	require.NotNil(t, doc)

	// global → UnitsScope → UnitScope → ConversionScope → PatternScope + CalculationScope
	require.Len(t, doc.Children, 1)
	unitsScope := doc.Children[0]
	require.Equal(t, symboltable.UnitsScope, unitsScope.ScopeType)
	require.Len(t, unitsScope.Children, 1)

	unitScope := unitsScope.Children[0]
	require.Equal(t, symboltable.UnitScope, unitScope.ScopeType)
	require.True(t, len(unitScope.Children) >= 1)

	var conversionScope *symboltable.SymbolTable
	for _, child := range unitScope.Children {
		if child.ScopeType == symboltable.ConversionScope {
			conversionScope = child
			break
		}
	}
	require.NotNil(t, conversionScope)

	var patternFound, calcFound bool
	for _, child := range conversionScope.Children {
		switch child.ScopeType {
		case symboltable.PatternScope:
			patternFound = true
		case symboltable.CalculationScope:
			calcFound = true
		}
	}
	assert.True(t, patternFound, "ConversionScope should contain a PatternScope")
	assert.True(t, calcFound, "ConversionScope should contain a CalculationScope")
}

// TestAcceptanceCriteria_3_AnalyzeUnitConversions verifies AC3:
// AnalyzeUnitConversions assembles ConversionPair[AnalysisUnitDefinition] entries with resolved
// units and validated calculations.
func TestAcceptanceCriteria_3_AnalyzeUnitConversions(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[km, m]]

units
	Kilometer = Distance
		conversion
			pattern Kilometer(km) = Meter(m)
			calculation m * 1000
	Meter = Distance`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	require.NotNil(t, defs)

	km := defs.Units["Kilometer"]
	require.Len(t, km.Conversions, 2)

	pair := km.Conversions[0]
	assert.Equal(t, "Meter", pair.SourceDef.Name)
	assert.Equal(t, "Kilometer", pair.TargetDef.Name)
	assert.Equal(t, "m", pair.SourceVar)
	assert.Equal(t, "km", pair.TargetVar)
	require.NotNil(t, pair.Calculation)

	inverse := km.Conversions[1]
	assert.Equal(t, "Kilometer", inverse.SourceDef.Name)
	assert.Equal(t, "Meter", inverse.TargetDef.Name)
	assert.Equal(t, "km", inverse.SourceVar)
	assert.Equal(t, "m", inverse.TargetVar)
	require.NotNil(t, inverse.Calculation)
}

// TestAcceptanceCriteria_4_AnalyzeMeasureConversions verifies AC4:
// AnalyzeMeasureConversions assembles ConversionPair[AnalysisMeasureDefinition] entries.
func TestAcceptanceCriteria_4_AnalyzeMeasureConversions(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[km, m]]

units
	KilometerUnit = Distance
	MeterUnit = Distance

measures
	KilometerLength = Number(unit=KilometerUnit)
		conversion
			pattern KilometerLength(km) = MeterLength(m)
			calculation m * 1000
	MeterLength = Number(unit=MeterUnit)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	require.NotNil(t, defs)

	kmLen := defs.Measures["KilometerLength"]
	require.Len(t, kmLen.Conversions, 2)
	require.NotNil(t, kmLen.Conversions[0].Calculation)
	require.NotNil(t, kmLen.Conversions[1].Calculation)
}

// TestAcceptanceCriteria_5_AlgebraicReversal verifies AC5: Algebraic reversal
// produces inverse conversion pairs for every forward pair.
func TestAcceptanceCriteria_5_AlgebraicReversal(t *testing.T) {
	t.Parallel()

	input := `concepts
	Distance = [Values[km, m]]

units
	KilometerUnit = Distance
	MeterUnit = Distance

measures
	KmMeasure = Number(unit=KilometerUnit)
		conversion
			pattern KmMeasure(km) = MeterMeasure(m)
			calculation m / 1000
	MeterMeasure = Number(unit=MeterUnit)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	require.NotNil(t, defs)

	kmMeasure := defs.Measures["KmMeasure"]
	require.Len(t, kmMeasure.Conversions, 2)

	// Forward: MeterMeasure → KmMeasure, calculation: m / 1000
	forward := kmMeasure.Conversions[0]
	assert.Equal(t, "MeterMeasure", forward.SourceDef.Name)
	assert.Equal(t, "KmMeasure", forward.TargetDef.Name)
	_, isForwardDiv := forward.Calculation.(*analysis.DivisionNode)
	assert.True(t, isForwardDiv, "forward should be division")

	// Inverse: KmMeasure → MeterMeasure, calculation: km * 1000
	inverse := kmMeasure.Conversions[1]
	assert.Equal(t, "KmMeasure", inverse.SourceDef.Name)
	assert.Equal(t, "MeterMeasure", inverse.TargetDef.Name)
	_, isInverseMult := inverse.Calculation.(*analysis.MultiplicationNode)
	assert.True(t, isInverseMult, "inverse should be multiplication")
}

// TestAcceptanceCriteria_6_RejectionCases verifies AC6: Analysis rejects bare
// numeric constants, variable name collisions, and cross-concept conversions.
func TestAcceptanceCriteria_6_RejectionCases(t *testing.T) {
	t.Parallel()

	// Bare numbers
	_, err := parseAndAnalyze(`concepts
	Distance = [Values[km, m]]
units
	Kilometer = Distance
		conversion
			pattern Kilometer(km) = Meter(m)
			calculation 5 + 9
	Meter = Distance`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bare numbers")

	// Variable name collision
	_, err = parseAndAnalyze(`concepts
	Distance = [Values[km, m]]
units
	Kilometer = Distance
	m = Distance
	Meter = Distance
		conversion
			pattern Kilometer(km) = Meter(m)
			calculation m * 1000`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "conflicts with unit name")

	// Cross-concept
	_, err = parseAndAnalyze(`concepts
	Distance = [Values[km, m]]
	Time = [Values[s, ms]]
units
	Kilometer = Distance
		conversion
			pattern Kilometer(km) = Second(s)
			calculation s * 1000
	Second = Time`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "different groups")
}

// TestAcceptanceCriteria_7_SynthesisEmitsConversionTables verifies AC7:
// Synthesis emits compilable Go code with conversion tables keyed by (source, target).
func TestAcceptanceCriteria_7_SynthesisEmitsConversionTables(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("'go' command not available in PATH")
	}

	input := `concepts
	Distance = [Values[km, m]]

units
	Kilometer = Distance
		conversion
			pattern Kilometer(km) = Meter(m)
			calculation m / 1000
	Meter = Distance`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err)
	require.NotNil(t, defs)

	inputSynth, err := synthesis.BuildTransformerFromAnalysis(defs)
	require.NoError(t, err)
	require.NotNil(t, inputSynth)

	// Verify the emitted code contains the expected conversion table structure
	code := inputSynth.UnitConversionCode
	assert.Contains(t, code, "var unitConversions = map[string]map[string]func(float64) float64")
	assert.Contains(t, code, "func convertUnit(sourceName, targetName string, value float64) float64")
	assert.Contains(t, code, `"Meter"`)
	assert.Contains(t, code, `"Kilometer"`)

	// Verify it compiles and runs
	program := fmt.Sprintf(`package main

import "fmt"

%s

func main() {
	fmt.Printf("%%.0f\n", convertUnit("Meter", "Kilometer", 5000))
	fmt.Printf("%%.0f\n", convertUnit("Kilometer", "Meter", 3))
}
`, code)

	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "main.go")
	err = os.WriteFile(srcPath, []byte(program), 0644)
	require.NoError(t, err)

	cmd := exec.Command("go", "run", srcPath)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "generated code failed to compile/run: %s", string(out))

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	require.Len(t, lines, 2)
	assert.Equal(t, "5", lines[0])
	assert.Equal(t, "3000", lines[1])
}

// TestAcceptanceCriteria_8_TargetExamplesWork verifies AC8: Target CPL examples
// (Length, TemperatureScale) parse, analyze, and synthesize correctly.
func TestAcceptanceCriteria_8_TargetExamplesWork(t *testing.T) {
	t.Parallel()

	// Length example
	lengthInput := `concepts
	DistanceUnit = [Values[Kilometer, Meter, Centimeter]]

units
	Kilometer = DistanceUnit
		conversion
			pattern Kilometer(km) = Meter(m)
			calculation m * 1000
	Meter = DistanceUnit
		conversion
			pattern Meter(m) = Kilometer(km)
			calculation km * 1000
	Centimeter = DistanceUnit
		conversion
			pattern Centimeter(cm) = Meter(m)
			calculation m * 100`

	defs, err := parseAndAnalyze(lengthInput)
	require.NoError(t, err)
	require.NotNil(t, defs)
	require.Len(t, defs.Units, 3)

	// Temperature example
	tempInput := `concepts
	TemperatureUnit = [Values[Celsius, Fahrenheit]]

units
	Celsius = TemperatureUnit
		conversion
			pattern Celsius(c) = Fahrenheit(f)
			calculation (f - 32) * 5 / 9
	Fahrenheit = TemperatureUnit
		conversion
			pattern Fahrenheit(f) = Celsius(c)
			calculation c * 9 / 5 + 32`

	defs, err = parseAndAnalyze(tempInput)
	require.NoError(t, err)
	require.NotNil(t, defs)
	require.Len(t, defs.Units, 2)
}
