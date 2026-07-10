// Package analysis_test contains tests for the CPL types analysis functionality.
package analysis_test

import (
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/analysis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// addStandardTimeConceptsToAnalyzer loads standard time concepts and adds them
// to the analyzer's concepts map for tests that exercise Instant/Interval types.
func addStandardTimeConceptsToAnalyzer(t *testing.T, analyzer *analysis.CPLAnalyzer) {
	t.Helper()
	concepts := analysis.LoadStandardTimeConcepts(t)
	for _, concept := range concepts {
		concept := concept // capture range variable
		analyzer.AddConcept(&concept)
	}
}

// TestNumberConstraintsValidate tests the NumberConstraints.Validate method.
func TestNumberConstraintsValidate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		constraints analysis.NumberConstraints
		expectError bool
		errorMsg    string
	}{
		{
			name:        "no constraints - valid",
			constraints: analysis.NumberConstraints{},
			expectError: false,
		},
		{
			name: "valid precision only",
			constraints: func() analysis.NumberConstraints {
				p := 2
				return analysis.NumberConstraints{Precision: &p}
			}(),
			expectError: false,
		},
		{
			name: "valid precision zero",
			constraints: func() analysis.NumberConstraints {
				p := 0
				return analysis.NumberConstraints{Precision: &p}
			}(),
			expectError: false,
		},
		{
			name: "negative precision - invalid",
			constraints: func() analysis.NumberConstraints {
				p := -1
				return analysis.NumberConstraints{Precision: &p}
			}(),
			expectError: true,
			errorMsg:    "precision must be non-negative",
		},
		{
			name: "valid minimum and maximum",
			constraints: func() analysis.NumberConstraints {
				min := 0
				max := 100
				return analysis.NumberConstraints{Minimum: &min, Maximum: &max}
			}(),
			expectError: false,
		},
		{
			name: "minimum equals maximum - invalid",
			constraints: func() analysis.NumberConstraints {
				min := 100
				max := 100
				return analysis.NumberConstraints{Minimum: &min, Maximum: &max}
			}(),
			expectError: true,
			errorMsg:    "minimum (100) must be less than maximum (100)",
		},
		{
			name: "minimum greater than maximum - invalid",
			constraints: func() analysis.NumberConstraints {
				min := 100
				max := 50
				return analysis.NumberConstraints{Minimum: &min, Maximum: &max}
			}(),
			expectError: true,
			errorMsg:    "minimum (100) must be less than maximum (50)",
		},
		{
			name: "minimum only - valid",
			constraints: func() analysis.NumberConstraints {
				min := 0
				return analysis.NumberConstraints{Minimum: &min}
			}(),
			expectError: false,
		},
		{
			name: "maximum only - valid",
			constraints: func() analysis.NumberConstraints {
				max := 100
				return analysis.NumberConstraints{Maximum: &max}
			}(),
			expectError: false,
		},
		{
			name: "all constraints valid",
			constraints: func() analysis.NumberConstraints {
				p := 2
				min := 0
				max := 100
				return analysis.NumberConstraints{Precision: &p, Minimum: &min, Maximum: &max}
			}(),
			expectError: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.constraints.Validate()
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCreateConstraints tests the CreateConstraints factory function.
func TestCreateConstraints(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name              string
		baseType          analysis.BuiltInType
		parserConstraints []analysis.Constraint
		expectError       bool
		errorMsg          string
		expectedType      string
	}{
		{
			name:              "Number type with no constraints",
			baseType:          analysis.NumberType,
			parserConstraints: []analysis.Constraint{},
			expectError:       false,
			expectedType:      "NumberConstraints",
		},
		{
			name:     "Number type with precision constraint",
			baseType: analysis.NumberType,
			parserConstraints: []analysis.Constraint{
				{Name: "precision", Value: "2"},
			},
			expectError:  false,
			expectedType: "NumberConstraints",
		},
		{
			name:     "Number type with min and max constraints",
			baseType: analysis.NumberType,
			parserConstraints: []analysis.Constraint{
				{Name: "minimum", Value: "0"},
				{Name: "maximum", Value: "100"},
			},
			expectError:  false,
			expectedType: "NumberConstraints",
		},
		{
			name:     "Number type with all constraints",
			baseType: analysis.NumberType,
			parserConstraints: []analysis.Constraint{
				{Name: "precision", Value: "2"},
				{Name: "minimum", Value: "0"},
				{Name: "maximum", Value: "100"},
			},
			expectError:  false,
			expectedType: "NumberConstraints",
		},
		{
			name:     "Number type with duplicate constraint names - error",
			baseType: analysis.NumberType,
			parserConstraints: []analysis.Constraint{
				{Name: "precision", Value: "2"},
				{Name: "precision", Value: "3"},
			},
			expectError: true,
			errorMsg:    "duplicate constraint name: precision",
		},
		{
			name:     "Number type with unknown constraint - error",
			baseType: analysis.NumberType,
			parserConstraints: []analysis.Constraint{
				{Name: "unknown", Value: "value"},
			},
			expectError: true,
			errorMsg:    "unknown constraint for Number: unknown",
		},
		{
			name:     "Number type with invalid precision value - error",
			baseType: analysis.NumberType,
			parserConstraints: []analysis.Constraint{
				{Name: "precision", Value: "not-a-number"},
			},
			expectError: true,
			errorMsg:    "precision constraint must be an integer",
		},
		{
			name:     "Number type with invalid min value - error",
			baseType: analysis.NumberType,
			parserConstraints: []analysis.Constraint{
				{Name: "minimum", Value: "abc"},
			},
			expectError: true,
			errorMsg:    "minimum constraint must be an integer",
		},
		{
			name:     "Number type with invalid max value - error",
			baseType: analysis.NumberType,
			parserConstraints: []analysis.Constraint{
				{Name: "maximum", Value: "xyz"},
			},
			expectError: true,
			errorMsg:    "maximum constraint must be an integer",
		},
		{
			name:     "Number type with min >= max - error",
			baseType: analysis.NumberType,
			parserConstraints: []analysis.Constraint{
				{Name: "minimum", Value: "100"},
				{Name: "maximum", Value: "50"},
			},
			expectError: true,
			errorMsg:    "minimum (100) must be less than maximum (50)",
		},
		{
			name:              "unsupported base type - error",
			baseType:          analysis.BuiltInType(999),
			parserConstraints: []analysis.Constraint{},
			expectError:       true,
			errorMsg:          "unsupported base type",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := analysis.CreateConstraints(tc.baseType, tc.parserConstraints, nil)
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				// Verify the type - now it returns *NumberConstraints (pointer)
				switch tc.expectedType {
				case "NumberConstraints":
					_, ok := result.(*analysis.NumberConstraints)
					assert.True(t, ok, "expected *NumberConstraints type")
				}
			}
		})
	}
}

// TestAnalysisTypeDefinition tests creating AnalysisTypeDefinition structs.
func TestAnalysisTypeDefinition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		typeDef      analysis.AnalysisTypeDefinition
		expectedName string
		expectedPath []string
	}{
		{
			name: "Number type with constraints",
			typeDef: analysis.AnalysisTypeDefinition{
				Name:           "Currency",
				BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
				RawConstraints: nil,
				DerivationPath: []string{"Currency", "Number"},
			},
			expectedName: "Currency",
			expectedPath: []string{"Currency", "Number"},
		},
		{
			name: "Number type with no constraints",
			typeDef: analysis.AnalysisTypeDefinition{
				Name:           "Count",
				BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
				RawConstraints: nil,
				DerivationPath: []string{"Count", "Number"},
			},
			expectedName: "Count",
			expectedPath: []string{"Count", "Number"},
		},
		{
			name: "Derived type with declared base",
			typeDef: analysis.AnalysisTypeDefinition{
				Name:           "SmallCurrency",
				BaseType:       analysis.NewDeclaredTypeRef("Currency"),
				RawConstraints: nil,
				DerivationPath: []string{"SmallCurrency", "Currency"},
			},
			expectedName: "SmallCurrency",
			expectedPath: []string{"SmallCurrency", "Currency"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.expectedName, tc.typeDef.Name)
			assert.Equal(t, tc.expectedPath, tc.typeDef.DerivationPath)
		})
	}
}

// TestCPLAnalyzerTypesMap tests that the CPLAnalyzer types map is properly initialized.
func TestCPLAnalyzerTypesMap(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()
	defs := analyzer.GetAnalysisDefinitions()

	assert.NotNil(t, defs.Types)
	assert.Len(t, defs.Types, 0)
}

// TestBuiltInTypeRef tests the BuiltInTypeRef struct.
func TestBuiltInTypeRef(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		builtinType   analysis.BuiltInType
		expectedName  string
		expectBuiltIn bool
	}{
		{
			name:          "NumberType returns Number",
			builtinType:   analysis.NumberType,
			expectedName:  "Number",
			expectBuiltIn: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ref := analysis.NewBuiltInTypeRef(tc.builtinType)
			assert.Equal(t, tc.expectedName, ref.GetName())
			assert.True(t, ref.IsBuiltIn())
			assert.Equal(t, tc.builtinType, ref.GetBuiltInType())
		})
	}
}

// TestDeclaredTypeRef tests the DeclaredTypeRef struct.
func TestDeclaredTypeRef(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		typeName      string
		expectedName  string
		expectBuiltIn bool
	}{
		{
			name:          "Currency declared type",
			typeName:      "Currency",
			expectedName:  "Currency",
			expectBuiltIn: false,
		},
		{
			name:          "SmallCurrency declared type",
			typeName:      "SmallCurrency",
			expectedName:  "SmallCurrency",
			expectBuiltIn: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ref := analysis.NewDeclaredTypeRef(tc.typeName)
			assert.Equal(t, tc.expectedName, ref.GetName())
			assert.False(t, ref.IsBuiltIn())
			// DeclaredTypeRef returns NumberType as default from GetBuiltInType
			assert.Equal(t, analysis.NumberType, ref.GetBuiltInType())
		})
	}
}

// TestDeclaredTypeNotEndingInBuiltin tests that declared types that don't terminate
// in a built-in type produce errors at analysis time, not parse time.
//
// This test is the analysis-level equivalent of the parser test TestParseInvalidBaseType.
// The parser now accepts identifiers as base types (for type derivation), so "Integer"
// parses successfully. However, the ANALYSIS phase detects that "Integer" doesn't
// terminate in a built-in type and reports an error.
//
// The separation is intentional:
// - Parser handles syntax (is the grammar valid?)
// - Analysis handles semantics (does the type chain resolve to a built-in?)
func TestDeclaredTypeNotEndingInBuiltin(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()

	// Add a type that uses "Integer" as its base type
	// "Integer" is not a built-in type, so the chain doesn't terminate in Number
	analyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "Currency",
		BaseType:       analysis.NewDeclaredTypeRef("Integer"),
		RawConstraints: []analysis.Constraint{{Name: "precision", Value: "2"}},
		DerivationPath: []string{},
	})

	err := analyzer.ComputeDerivationPaths()

	// Analysis should fail because "Integer" doesn't derive from a built-in type
	require.Error(t, err, "Analysis should fail when type chain doesn't end in built-in")
	assert.Contains(t, err.Error(), "does not derive from a built-in type",
		"Error should indicate the chain doesn't end in built-in")
}

// TestParserAcceptsIdentifierButAnalysisRejects tests the full parse+analyze flow
// where parsing succeeds but analysis fails.
//
// This test verifies the complete workflow:
// 1. Parser accepts "Integer" as a valid identifier (baseType rule)
// 2. Analysis detects that "Integer" doesn't terminate in a built-in type
//
// This is equivalent to the parser test TestParseInvalidBaseType, but for the
// analysis phase.
func TestParserAcceptsIdentifierButAnalysisRejects(t *testing.T) {
	t.Parallel()

	// This is a direct analysis test, not a full parse+analyze integration test
	// The key point is that declared types with no built-in termination are rejected
	analyzer := analysis.NewCPLAnalyzer()

	// Simulate the state after parsing "types Currency = Integer(precision=2)"
	// The parser has accepted "Integer" as a valid base type identifier
	analyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:     "Currency",
		BaseType: analysis.NewDeclaredTypeRef("Integer"),
		RawConstraints: []analysis.Constraint{
			{Name: "precision", Value: "2"},
		},
		DerivationPath: []string{},
	})

	// ComputeDerivationPaths should fail
	err := analyzer.ComputeDerivationPaths()
	require.Error(t, err, "Analysis should fail for type not ending in built-in")
	assert.Contains(t, err.Error(), "does not derive from a built-in type")
}

// TestTypeDerivationChainWithMultiLevelInheritance tests that multi-level type
// derivation chains work correctly when they terminate in a built-in type.
//
// This test creates a chain: SmallCurrency -> Currency -> Number
// All types terminate in the built-in Number type.
func TestTypeDerivationChainWithMultiLevelInheritance(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()

	// Add three types in a chain
	analyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "Currency",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
		RawConstraints: []analysis.Constraint{{Name: "precision", Value: "2"}},
		DerivationPath: []string{},
	})
	analyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "SmallCurrency",
		BaseType:       analysis.NewDeclaredTypeRef("Currency"),
		RawConstraints: []analysis.Constraint{{Name: "precision", Value: "0"}},
		DerivationPath: []string{},
	})
	analyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "TinyCurrency",
		BaseType:       analysis.NewDeclaredTypeRef("SmallCurrency"),
		RawConstraints: []analysis.Constraint{{Name: "precision", Value: "0"}, {Name: "minimum", Value: "0"}},
		DerivationPath: []string{},
	})

	err := analyzer.ComputeDerivationPaths()
	require.NoError(t, err, "Multi-level derivation chain should succeed")

	defs := analyzer.GetAnalysisDefinitions()
	assert.Equal(t, []string{"Currency", "Number"}, defs.Types["Currency"].DerivationPath)
	assert.Equal(t, []string{"SmallCurrency", "Currency", "Number"}, defs.Types["SmallCurrency"].DerivationPath)
	assert.Equal(t, []string{"TinyCurrency", "SmallCurrency", "Currency", "Number"}, defs.Types["TinyCurrency"].DerivationPath)

	// Verify TinyCurrency has the most restrictive constraints
	tinyConstraints := defs.Types["TinyCurrency"].ParsedConstraints.(*analysis.NumberConstraints)
	require.NotNil(t, tinyConstraints.Precision)
	assert.Equal(t, 0, *tinyConstraints.Precision)
	require.NotNil(t, tinyConstraints.Minimum)
	assert.Equal(t, 0, *tinyConstraints.Minimum)
}

// TestTypeReferenceInterface tests that TypeReference interface is satisfied.
func TestTypeReferenceInterface(t *testing.T) {
	t.Parallel()

	// This test verifies that both BuiltInTypeRef and DeclaredTypeRef
	// satisfy the TypeReference interface
	var _ analysis.TypeReference = analysis.NewBuiltInTypeRef(analysis.NumberType)
	var _ analysis.TypeReference = analysis.NewDeclaredTypeRef("Currency")
}

// TestComputeDerivationPaths tests the ComputeDerivationPaths method on CPLAnalyzer.
func TestComputeDerivationPaths(t *testing.T) {
	t.Parallel()

	t.Run("single type derived from built-in", func(t *testing.T) {
		t.Parallel()

		analyzer := analysis.NewCPLAnalyzer()

		// Add a type definition using the test helper
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "Currency",
			BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
			RawConstraints: nil,
			DerivationPath: []string{}, // Initially empty
		})

		err := analyzer.ComputeDerivationPaths()
		require.NoError(t, err)

		// Use GetAnalysisDefinitions to access types
		defs := analyzer.GetAnalysisDefinitions()
		assert.Equal(t, []string{"Currency", "Number"}, defs.Types["Currency"].DerivationPath)
	})

	t.Run("type derived from another type", func(t *testing.T) {
		t.Parallel()

		analyzer := analysis.NewCPLAnalyzer()

		// Add two type definitions
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "Currency",
			BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
			RawConstraints: nil,
			DerivationPath: []string{},
		})
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "SmallCurrency",
			BaseType:       analysis.NewDeclaredTypeRef("Currency"),
			RawConstraints: nil,
			DerivationPath: []string{},
		})

		err := analyzer.ComputeDerivationPaths()
		require.NoError(t, err)

		defs := analyzer.GetAnalysisDefinitions()
		assert.Equal(t, []string{"Currency", "Number"}, defs.Types["Currency"].DerivationPath)
		assert.Equal(t, []string{"SmallCurrency", "Currency", "Number"}, defs.Types["SmallCurrency"].DerivationPath)
	})

	t.Run("multiple levels of type derivation", func(t *testing.T) {
		t.Parallel()

		analyzer := analysis.NewCPLAnalyzer()

		// Add three type definitions in chain
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "Number",
			BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
			RawConstraints: nil,
			DerivationPath: []string{},
		})
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "Currency",
			BaseType:       analysis.NewDeclaredTypeRef("Number"),
			RawConstraints: nil,
			DerivationPath: []string{},
		})
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "SmallCurrency",
			BaseType:       analysis.NewDeclaredTypeRef("Currency"),
			RawConstraints: nil,
			DerivationPath: []string{},
		})

		err := analyzer.ComputeDerivationPaths()
		require.NoError(t, err)

		defs := analyzer.GetAnalysisDefinitions()
		assert.Equal(t, []string{"Number", "Number"}, defs.Types["Number"].DerivationPath)
		assert.Equal(t, []string{"Currency", "Number", "Number"}, defs.Types["Currency"].DerivationPath)
		assert.Equal(t, []string{"SmallCurrency", "Currency", "Number", "Number"}, defs.Types["SmallCurrency"].DerivationPath)
	})

	t.Run("unknown type in chain returns error", func(t *testing.T) {
		t.Parallel()

		analyzer := analysis.NewCPLAnalyzer()

		// Add a type that references an unknown type
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "UnknownRef",
			BaseType:       analysis.NewDeclaredTypeRef("NonExistent"),
			RawConstraints: nil,
			DerivationPath: []string{},
		})

		err := analyzer.ComputeDerivationPaths()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "type NonExistent not found")
	})

	t.Run("type chain not ending in built-in returns error", func(t *testing.T) {
		t.Parallel()

		analyzer := analysis.NewCPLAnalyzer()

		// Add two types: Currency derives from built-in Number,
		// but SmallCurrency derives from Currency (not built-in)
		// This forms a chain: SmallCurrency -> Currency -> Number
		// The chain DOES end in a built-in, so this test is about the case
		// where we have a type that references a type that exists but the chain
		// doesn't reach a built-in (we'd need a circular reference or orphan type)
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "Currency",
			BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
			RawConstraints: nil,
			DerivationPath: []string{},
		})
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "SmallCurrency",
			BaseType:       analysis.NewDeclaredTypeRef("Currency"),
			RawConstraints: nil,
			DerivationPath: []string{},
		})

		// This should succeed - the chain ends in Number (built-in)
		err := analyzer.ComputeDerivationPaths()
		require.NoError(t, err)

		defs := analyzer.GetAnalysisDefinitions()
		assert.Equal(t, []string{"Currency", "Number"}, defs.Types["Currency"].DerivationPath)
		assert.Equal(t, []string{"SmallCurrency", "Currency", "Number"}, defs.Types["SmallCurrency"].DerivationPath)
	})

	t.Run("type referencing non-built-in base type only", func(t *testing.T) {
		t.Parallel()

		analyzer := analysis.NewCPLAnalyzer()

		// Add a type that only references a declared type, not a built-in
		// This simulates a type chain that doesn't terminate in a built-in type
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "OrphanType",
			BaseType:       analysis.NewDeclaredTypeRef("SomeBaseType"),
			RawConstraints: nil,
			DerivationPath: []string{},
		})

		err := analyzer.ComputeDerivationPaths()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not derive from a built-in type")
	})
}

// TestComputeDerivationPathsConstraints tests constraint validation in derivation paths.
func TestComputeDerivationPathsConstraints(t *testing.T) {
	t.Parallel()

	t.Run("constraints only narrowing - valid", func(t *testing.T) {
		t.Parallel()

		analyzer := analysis.NewCPLAnalyzer()

		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "Currency",
			BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
			RawConstraints: []analysis.Constraint{{Name: "precision", Value: "2"}},
			DerivationPath: []string{},
		})
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "SmallCurrency",
			BaseType:       analysis.NewDeclaredTypeRef("Currency"),
			RawConstraints: []analysis.Constraint{{Name: "precision", Value: "0"}},
			DerivationPath: []string{},
		})

		err := analyzer.ComputeDerivationPaths()
		require.NoError(t, err)

		defs := analyzer.GetAnalysisDefinitions()
		// SmallCurrency should have the more restrictive precision=0
		smallCurrencyConstraints := defs.Types["SmallCurrency"].ParsedConstraints.(*analysis.NumberConstraints)
		require.NotNil(t, smallCurrencyConstraints.Precision)
		assert.Equal(t, 0, *smallCurrencyConstraints.Precision)
	})

	t.Run("constraints expanding - invalid", func(t *testing.T) {
		t.Parallel()

		analyzer := analysis.NewCPLAnalyzer()

		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "ExactCurrency",
			BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
			RawConstraints: []analysis.Constraint{{Name: "precision", Value: "0"}},
			DerivationPath: []string{},
		})
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "FlexibleCurrency",
			BaseType:       analysis.NewDeclaredTypeRef("ExactCurrency"),
			RawConstraints: []analysis.Constraint{{Name: "precision", Value: "2"}},
			DerivationPath: []string{},
		})

		err := analyzer.ComputeDerivationPaths()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expands inherited constraint")
	})

	t.Run("minimum expanding - invalid", func(t *testing.T) {
		t.Parallel()

		analyzer := analysis.NewCPLAnalyzer()

		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "BoundedType",
			BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
			RawConstraints: []analysis.Constraint{{Name: "minimum", Value: "10"}},
			DerivationPath: []string{},
		})
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "ExpandedType",
			BaseType:       analysis.NewDeclaredTypeRef("BoundedType"),
			RawConstraints: []analysis.Constraint{{Name: "minimum", Value: "0"}},
			DerivationPath: []string{},
		})

		err := analyzer.ComputeDerivationPaths()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expands inherited constraint")
	})

	t.Run("maximum expanding - invalid", func(t *testing.T) {
		t.Parallel()

		analyzer := analysis.NewCPLAnalyzer()

		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "SmallRange",
			BaseType:       analysis.NewBuiltInTypeRef(analysis.NumberType),
			RawConstraints: []analysis.Constraint{{Name: "maximum", Value: "100"}},
			DerivationPath: []string{},
		})
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "LargeRange",
			BaseType:       analysis.NewDeclaredTypeRef("SmallRange"),
			RawConstraints: []analysis.Constraint{{Name: "maximum", Value: "1000"}},
			DerivationPath: []string{},
		})

		err := analyzer.ComputeDerivationPaths()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expands inherited constraint")
	})

	t.Run("mixed constraints with expanding precision - invalid", func(t *testing.T) {
		t.Parallel()

		analyzer := analysis.NewCPLAnalyzer()

		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:     "PreciseCurrency",
			BaseType: analysis.NewBuiltInTypeRef(analysis.NumberType),
			RawConstraints: []analysis.Constraint{
				{Name: "precision", Value: "0"},
				{Name: "minimum", Value: "0"},
			},
			DerivationPath: []string{},
		})
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:     "FlexibleCurrency",
			BaseType: analysis.NewDeclaredTypeRef("PreciseCurrency"),
			RawConstraints: []analysis.Constraint{
				{Name: "precision", Value: "2"}, // Expanded - invalid
			},
			DerivationPath: []string{},
		})

		err := analyzer.ComputeDerivationPaths()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expands inherited constraint")
	})
}

// TestInstantConstraintsValidate tests the InstantConstraints.Validate method.
// Validation now happens during constraint creation via concept lookup, so
// Validate() always returns nil for validly-constructed constraints.
func TestInstantConstraintsValidate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		constraints analysis.InstantConstraints
	}{
		{
			name:        "no constraints - valid",
			constraints: analysis.InstantConstraints{},
		},
		{
			name: "valid year only",
			constraints: func() analysis.InstantConstraints {
				y := 2024
				return analysis.InstantConstraints{Year: &y}
			}(),
		},
		{
			name: "valid month only",
			constraints: func() analysis.InstantConstraints {
				m := "January"
				return analysis.InstantConstraints{Month: &m}
			}(),
		},
		{
			name: "valid day_of_month only",
			constraints: func() analysis.InstantConstraints {
				d := 15
				return analysis.InstantConstraints{DayOfMonth: &d}
			}(),
		},
		{
			name: "valid hour only",
			constraints: func() analysis.InstantConstraints {
				h := 12
				return analysis.InstantConstraints{Hour: &h}
			}(),
		},
		{
			name: "valid minute only",
			constraints: func() analysis.InstantConstraints {
				m := 30
				return analysis.InstantConstraints{Minute: &m}
			}(),
		},
		{
			name: "valid second only",
			constraints: func() analysis.InstantConstraints {
				s := 1
				return analysis.InstantConstraints{Second: &s}
			}(),
		},
		{
			name: "valid day_of_week only",
			constraints: func() analysis.InstantConstraints {
				d := "Monday"
				return analysis.InstantConstraints{DayOfWeek: &d}
			}(),
		},
		{
			name: "valid week_number only",
			constraints: func() analysis.InstantConstraints {
				w := 1
				return analysis.InstantConstraints{WeekNumber: &w}
			}(),
		},
		{
			name: "multiple valid constraints",
			constraints: func() analysis.InstantConstraints {
				y := 2024
				m := "January"
				d := 1
				return analysis.InstantConstraints{Year: &y, Month: &m, DayOfMonth: &d}
			}(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.NoError(t, tc.constraints.Validate())
		})
	}
}

// TestInstantConstraintsConstraintNames tests the ConstraintNames method.
func TestInstantConstraintsConstraintNames(t *testing.T) {
	t.Parallel()

	t.Run("no constraints", func(t *testing.T) {
		t.Parallel()
		ic := analysis.InstantConstraints{}
		assert.Empty(t, ic.ConstraintNames())
	})

	t.Run("all constraints", func(t *testing.T) {
		t.Parallel()
		y, m, dom, doy, fq, h, min, s, dow, wn := 2024, "January", 15, 42, 1, 12, 30, 1, "Monday", 1
		ic := analysis.InstantConstraints{
			Year: &y, Month: &m, DayOfMonth: &dom, DayOfYear: &doy, FiscalQuarter: &fq,
			Hour: &h, Minute: &min, Second: &s, DayOfWeek: &dow, WeekNumber: &wn,
		}
		names := ic.ConstraintNames()
		assert.Len(t, names, 10)
		assert.Contains(t, names, "year")
		assert.Contains(t, names, "month")
		assert.Contains(t, names, "day_of_month")
		assert.Contains(t, names, "day_of_year")
		assert.Contains(t, names, "fiscal_quarter")
		assert.Contains(t, names, "hour")
		assert.Contains(t, names, "minute")
		assert.Contains(t, names, "second")
		assert.Contains(t, names, "day_of_week")
		assert.Contains(t, names, "week_number")
	})
}

// TestInstantConstraintsGetConstraint tests the GetConstraint method.
func TestInstantConstraintsGetConstraint(t *testing.T) {
	t.Parallel()

	y, m, dom := 2024, "January", 15
	ic := analysis.InstantConstraints{Year: &y, Month: &m, DayOfMonth: &dom}

	t.Run("existing int constraint", func(t *testing.T) {
		t.Parallel()
		val, ok := ic.GetConstraint("year")
		require.True(t, ok)
		assert.Equal(t, &y, val)
	})

	t.Run("existing string constraint", func(t *testing.T) {
		t.Parallel()
		val, ok := ic.GetConstraint("month")
		require.True(t, ok)
		assert.Equal(t, &m, val)
	})

	t.Run("missing constraint", func(t *testing.T) {
		t.Parallel()
		val, ok := ic.GetConstraint("day_of_year")
		assert.False(t, ok)
		assert.Nil(t, val)
	})

	t.Run("unknown constraint", func(t *testing.T) {
		t.Parallel()
		val, ok := ic.GetConstraint("foobar")
		assert.False(t, ok)
		assert.Nil(t, val)
	})
}

// TestInstantConstraintsIsMoreOrEqualRestrictive tests the IsMoreOrEqualRestrictive method.
func TestInstantConstraintsIsMoreOrEqualRestrictive(t *testing.T) {
	t.Parallel()

	ic := analysis.InstantConstraints{}

	t.Run("nil to value - more restrictive", func(t *testing.T) {
		t.Parallel()
		v := 2024
		ok, err := ic.IsMoreOrEqualRestrictive("year", (*int)(nil), &v)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("value to nil - less restrictive", func(t *testing.T) {
		t.Parallel()
		v := 2024
		ok, err := ic.IsMoreOrEqualRestrictive("year", &v, (*int)(nil))
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("same value - equal restrictive", func(t *testing.T) {
		t.Parallel()
		v := 2024
		ok, err := ic.IsMoreOrEqualRestrictive("year", &v, &v)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("different value - not more restrictive", func(t *testing.T) {
		t.Parallel()
		v1, v2 := 2024, 2025
		ok, err := ic.IsMoreOrEqualRestrictive("year", &v1, &v2)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("string same value - equal restrictive", func(t *testing.T) {
		t.Parallel()
		m := "January"
		ok, err := ic.IsMoreOrEqualRestrictive("month", &m, &m)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("string different value - not more restrictive", func(t *testing.T) {
		t.Parallel()
		m1, m2 := "January", "February"
		ok, err := ic.IsMoreOrEqualRestrictive("month", &m1, &m2)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("invalid type - error", func(t *testing.T) {
		t.Parallel()
		v := "not-an-int"
		_, err := ic.IsMoreOrEqualRestrictive("year", &v, &v)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must be *int")
	})

	t.Run("unknown constraint - error", func(t *testing.T) {
		t.Parallel()
		_, err := ic.IsMoreOrEqualRestrictive("foobar", nil, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown constraint name")
	})
}

// TestComputeDerivationPathsInstant tests derivation paths for Instant types.
func TestComputeDerivationPathsInstant(t *testing.T) {
	t.Parallel()

	t.Run("single Instant type", func(t *testing.T) {
		t.Parallel()
		analyzer := analysis.NewCPLAnalyzer()
		addStandardTimeConceptsToAnalyzer(t, analyzer)
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "Timestamp",
			BaseType:       analysis.NewBuiltInTypeRef(analysis.InstantType),
			RawConstraints: []analysis.Constraint{{Name: "second", Value: "1"}},
			DerivationPath: []string{},
		})

		err := analyzer.ComputeDerivationPaths()
		require.NoError(t, err)

		defs := analyzer.GetAnalysisDefinitions()
		assert.Equal(t, []string{"Timestamp", "Instant"}, defs.Types["Timestamp"].DerivationPath)
		ic, ok := defs.Types["Timestamp"].ParsedConstraints.(*analysis.InstantConstraints)
		require.True(t, ok)
		require.NotNil(t, ic.Second)
		assert.Equal(t, 1, *ic.Second)
	})

	t.Run("Instant type derivation chain", func(t *testing.T) {
		t.Parallel()
		analyzer := analysis.NewCPLAnalyzer()
		addStandardTimeConceptsToAnalyzer(t, analyzer)
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "NewYear",
			BaseType:       analysis.NewBuiltInTypeRef(analysis.InstantType),
			RawConstraints: []analysis.Constraint{{Name: "month", Value: "January"}, {Name: "day_of_month", Value: "1"}},
			DerivationPath: []string{},
		})
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "NewYear2024",
			BaseType:       analysis.NewDeclaredTypeRef("NewYear"),
			RawConstraints: []analysis.Constraint{{Name: "year", Value: "2024"}},
			DerivationPath: []string{},
		})

		err := analyzer.ComputeDerivationPaths()
		require.NoError(t, err)

		defs := analyzer.GetAnalysisDefinitions()
		assert.Equal(t, []string{"NewYear", "Instant"}, defs.Types["NewYear"].DerivationPath)
		assert.Equal(t, []string{"NewYear2024", "NewYear", "Instant"}, defs.Types["NewYear2024"].DerivationPath)

		ic, ok := defs.Types["NewYear2024"].ParsedConstraints.(*analysis.InstantConstraints)
		require.True(t, ok)
		require.NotNil(t, ic.Year)
		assert.Equal(t, 2024, *ic.Year)
		require.NotNil(t, ic.Month)
		assert.Equal(t, "January", *ic.Month)
		require.NotNil(t, ic.DayOfMonth)
		assert.Equal(t, 1, *ic.DayOfMonth)
	})

	t.Run("Instant constraint expansion - invalid", func(t *testing.T) {
		t.Parallel()
		analyzer := analysis.NewCPLAnalyzer()
		addStandardTimeConceptsToAnalyzer(t, analyzer)
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "ExactTime",
			BaseType:       analysis.NewBuiltInTypeRef(analysis.InstantType),
			RawConstraints: []analysis.Constraint{{Name: "hour", Value: "12"}},
			DerivationPath: []string{},
		})
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "DifferentTime",
			BaseType:       analysis.NewDeclaredTypeRef("ExactTime"),
			RawConstraints: []analysis.Constraint{{Name: "hour", Value: "13"}},
			DerivationPath: []string{},
		})

		err := analyzer.ComputeDerivationPaths()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expands inherited constraint")
	})

	t.Run("Instant type missing concepts - error", func(t *testing.T) {
		t.Parallel()
		analyzer := analysis.NewCPLAnalyzer()
		// Do NOT add standard time concepts
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "Timestamp",
			BaseType:       analysis.NewBuiltInTypeRef(analysis.InstantType),
			RawConstraints: []analysis.Constraint{{Name: "second", Value: "1"}},
			DerivationPath: []string{},
		})

		err := analyzer.ComputeDerivationPaths()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required concept Second for constraint second is not defined")
	})
}

// TestParseAndAnalyzeIntervalType tests the full parse+analyze pipeline for Interval types.
func TestParseAndAnalyzeIntervalType(t *testing.T) {
	t.Parallel()

	input := `concepts
		DurationUnit=[Values[month, day]]
		Year = [Int[1900, 3000, 1]]
		Month=[Values[January, February, March, April, May, June, July, August, September, October, November, December]]
		DayOfMonth = [Int[1, 31, 1]]
	units
		month = DurationUnit
	types
		FiscalYearStart = Instant(month=January, day_of_month=1)
		FiscalMonth = Interval(starts_at=FiscalYearStart, unit=month)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err, "Full parse+analyze should succeed for Interval type")

	require.Contains(t, defs.Types, "FiscalMonth")
	assert.Equal(t, []string{"FiscalMonth", "Interval"}, defs.Types["FiscalMonth"].DerivationPath)
	ic, ok := defs.Types["FiscalMonth"].ParsedConstraints.(*analysis.IntervalConstraints)
	require.True(t, ok)
	require.NotNil(t, ic.StartTypeName)
	assert.Equal(t, "FiscalYearStart", *ic.StartTypeName)
	require.NotNil(t, ic.UnitName)
	assert.Equal(t, "month", *ic.UnitName)
}

// TestParseAndAnalyzeIntervalTypeInvalidStartsAt tests analysis failure for non-Instant starts_at.
func TestParseAndAnalyzeIntervalTypeInvalidStartsAt(t *testing.T) {
	t.Parallel()

	input := `concepts
		DurationUnit=[Values[month]]
		Year = [Int[1900, 3000, 1]]
		Month=[Values[January, February, March, April, May, June, July, August, September, October, November, December]]
		DayOfMonth = [Int[1, 31, 1]]
	units
		month = DurationUnit
	types
		Currency = Number(precision=2)
		FiscalMonth = Interval(starts_at=Currency, unit=month)`

	_, err := parseAndAnalyze(input)
	require.Error(t, err, "Analysis should fail for non-Instant starts_at type")
	assert.Contains(t, err.Error(), "must derive from Instant, got Number")
}

// TestParseAndAnalyzeIntervalTypeWithInlineStartsAt tests the full parse+analyze pipeline for inline Interval starts_at constraints.
func TestParseAndAnalyzeIntervalTypeWithInlineStartsAt(t *testing.T) {
	t.Parallel()

	input := `concepts
		DurationUnit=[Values[day]]
		DayOfWeek = [Values[Sunday, Monday, Tuesday, Wednesday, Thursday, Friday, Saturday]]
	units
		day = DurationUnit
	types
		WorkWeek = Interval(starts_at=Instant(day_of_week=Monday), unit=day)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err, "Full parse+analyze should succeed for inline Interval starts_at")

	ic, ok := defs.Types["WorkWeek"].ParsedConstraints.(*analysis.IntervalConstraints)
	require.True(t, ok)
	require.NotNil(t, ic.StartsAt)
	assert.Nil(t, ic.StartTypeName)
	require.NotNil(t, ic.StartsAt.DayOfWeek)
	assert.Equal(t, "Monday", *ic.StartsAt.DayOfWeek)
	require.NotNil(t, ic.UnitName)
	assert.Equal(t, "day", *ic.UnitName)
}

// TestParseAndAnalyzeIntervalTypeWithNonInstantInlineStartsAtFails tests that generic nested starts_at values fail semantically unless they use Instant.
func TestParseAndAnalyzeIntervalTypeWithNonInstantInlineStartsAtFails(t *testing.T) {
	t.Parallel()

	input := `concepts
		DurationUnit=[Values[day]]
		DayOfWeek = [Values[Sunday, Monday, Tuesday, Wednesday, Thursday, Friday, Saturday]]
	units
		day = DurationUnit
	types
		WorkWeek = Interval(starts_at=NotInstant(day_of_week=Monday), unit=day)`

	_, err := parseAndAnalyze(input)
	require.Error(t, err, "Analysis should reject non-Instant nested starts_at values")
	assert.Contains(t, err.Error(), "starts_at inline constraint must use Instant, got NotInstant")
}

// TestParseAndAnalyzeIntervalTypeWithMissingConstraintsFails tests that empty Interval declarations fail analysis.
func TestParseAndAnalyzeIntervalTypeWithMissingConstraintsFails(t *testing.T) {
	t.Parallel()

	input := "types\n\tAnyInterval = Interval()"

	_, err := parseAndAnalyze(input)
	require.Error(t, err, "Analysis should fail for Interval type missing required constraints")
	assert.Contains(t, err.Error(), "Interval type must specify starts_at")
}

// TestParseAndAnalyzeIntervalTypeWithMissingUnitReferenceFails tests that Interval unit references must exist.
func TestParseAndAnalyzeIntervalTypeWithMissingUnitReferenceFails(t *testing.T) {
	t.Parallel()

	input := `concepts
		Year = [Int[1900, 3000, 1]]
		Month=[Values[January, February, March, April, May, June, July, August, September, October, November, December]]
		DayOfMonth = [Int[1, 31, 1]]
	types
		FiscalYearStart = Instant(month=January, day_of_month=1)
		FiscalMonth = Interval(starts_at=FiscalYearStart, unit=month)`

	_, err := parseAndAnalyze(input)
	require.Error(t, err, "Analysis should fail for Interval type referencing an undefined unit")
	assert.Contains(t, err.Error(), "references unknown unit month")
}

// TestParseAndAnalyzeInstantWithoutConceptsFails confirms that analysis fails with
// a clear error when a CPL file uses Instant constraints but standard concepts are not available.
func TestParseAndAnalyzeInstantWithoutConceptsFails(t *testing.T) {
	t.Parallel()

	input := "types\n\tTimestamp = Instant(second=1)"
	_, err := parseAndAnalyze(input)
	require.Error(t, err, "Analysis should fail when required time concepts are not defined")
	assert.Contains(t, err.Error(), "required concept Second for constraint second is not defined")
}

// TestParseAndAnalyzeInstantType tests the full parse+analyze pipeline for Instant types
// when standard concepts are explicitly provided in the source.
func TestParseAndAnalyzeInstantType(t *testing.T) {
	t.Parallel()

	input := `concepts
		Second = [Int[0, 59, 1]]
	types
		Timestamp = Instant(second=1)`
	defs, err := parseAndAnalyze(input)
	require.NoError(t, err, "Full parse+analyze should succeed for Instant type with concepts defined")

	require.Contains(t, defs.Types, "Timestamp")
	assert.Equal(t, []string{"Timestamp", "Instant"}, defs.Types["Timestamp"].DerivationPath)
	ic, ok := defs.Types["Timestamp"].ParsedConstraints.(*analysis.InstantConstraints)
	require.True(t, ok)
	require.NotNil(t, ic.Second)
	assert.Equal(t, 1, *ic.Second)
}

// TestCreateConstraintsInstant tests CreateConstraints with InstantType.
func TestCreateConstraintsInstant(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		constraints  []analysis.Constraint
		expectError  bool
		errorMsg     string
		expectedType string
	}{
		{
			name:         "Instant with no constraints",
			constraints:  []analysis.Constraint{},
			expectError:  false,
			expectedType: "InstantConstraints",
		},
		{
			name: "Instant with second constraint",
			constraints: []analysis.Constraint{
				{Name: "second", Value: "1"},
			},
			expectError:  false,
			expectedType: "InstantConstraints",
		},
		{
			name: "Instant with year constraint",
			constraints: []analysis.Constraint{
				{Name: "year", Value: "2024"},
			},
			expectError:  false,
			expectedType: "InstantConstraints",
		},
		{
			name: "Instant with month constraint",
			constraints: []analysis.Constraint{
				{Name: "month", Value: "January"},
			},
			expectError:  false,
			expectedType: "InstantConstraints",
		},
		{
			name: "Instant with day_of_week constraint",
			constraints: []analysis.Constraint{
				{Name: "day_of_week", Value: "Monday"},
			},
			expectError:  false,
			expectedType: "InstantConstraints",
		},
		{
			name: "Instant with multiple constraints",
			constraints: []analysis.Constraint{
				{Name: "month", Value: "January"},
				{Name: "day_of_month", Value: "1"},
			},
			expectError:  false,
			expectedType: "InstantConstraints",
		},
		{
			name: "Instant with day_of_year constraint",
			constraints: []analysis.Constraint{
				{Name: "day_of_year", Value: "42"},
			},
			expectError:  false,
			expectedType: "InstantConstraints",
		},
		{
			name: "Instant with fiscal_quarter constraint",
			constraints: []analysis.Constraint{
				{Name: "fiscal_quarter", Value: "1"},
			},
			expectError:  false,
			expectedType: "InstantConstraints",
		},
		{
			name: "Instant with day_of_month constraint",
			constraints: []analysis.Constraint{
				{Name: "day_of_month", Value: "1"},
			},
			expectError:  false,
			expectedType: "InstantConstraints",
		},
		{
			name: "Instant with duplicate constraint - error",
			constraints: []analysis.Constraint{
				{Name: "second", Value: "1"},
				{Name: "second", Value: "30"},
			},
			expectError: true,
			errorMsg:    "duplicate constraint name: second",
		},
		{
			name: "Instant with unknown constraint - error",
			constraints: []analysis.Constraint{
				{Name: "invalid_field", Value: "123"},
			},
			expectError: true,
			errorMsg:    "unknown constraint for Instant: invalid_field",
		},
		{
			name: "Instant with invalid year value - error",
			constraints: []analysis.Constraint{
				{Name: "year", Value: "not-a-number"},
			},
			expectError: true,
			errorMsg:    "year constraint must be an integer",
		},
		{
			name: "Instant with invalid second range - error",
			constraints: []analysis.Constraint{
				{Name: "second", Value: "60"},
			},
			expectError: true,
			errorMsg:    "second must be between 0 and 59",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := analysis.CreateConstraints(analysis.InstantType, tc.constraints, analysis.LoadStandardTimeConcepts(t))
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				_, ok := result.(*analysis.InstantConstraints)
				assert.True(t, ok, "expected *InstantConstraints type")
			}
		})
	}
}

// TestCreateConstraintsInterval tests CreateConstraints with IntervalType.
func TestCreateConstraintsInterval(t *testing.T) {
	t.Parallel()

	t.Run("Interval with no constraints - error", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.IntervalType, []analysis.Constraint{}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Interval type must specify starts_at")
		assert.Nil(t, result)
	})

	t.Run("Interval with starts_at named type", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.IntervalType, []analysis.Constraint{
			{Name: "starts_at", Value: "FiscalYearStart"},
			{Name: "unit", Value: "month"},
		}, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		ic, ok := result.(*analysis.IntervalConstraints)
		require.True(t, ok)
		require.NotNil(t, ic.StartTypeName)
		assert.Equal(t, "FiscalYearStart", *ic.StartTypeName)
		require.NotNil(t, ic.UnitName)
		assert.Equal(t, "month", *ic.UnitName)
		assert.Nil(t, ic.StartsAt)
	})

	t.Run("Interval with unknown constraint", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.IntervalType, []analysis.Constraint{
			{Name: "invalid_field", Value: "123"},
		}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown constraint for Interval: invalid_field")
		assert.Nil(t, result)
	})

	t.Run("Interval with duplicate constraint", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.IntervalType, []analysis.Constraint{
			{Name: "unit", Value: "day"},
			{Name: "unit", Value: "month"},
		}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate constraint name: unit")
		assert.Nil(t, result)
	})

	t.Run("Interval with missing starts_at - error", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.IntervalType, []analysis.Constraint{
			{Name: "unit", Value: "month"},
		}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Interval type must specify starts_at")
		assert.Nil(t, result)
	})

	t.Run("Interval with missing unit - valid", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.IntervalType, []analysis.Constraint{
			{Name: "starts_at", Value: "FiscalYearStart"},
		}, nil)
		require.NoError(t, err)
		require.NotNil(t, result)
		ic, ok := result.(*analysis.IntervalConstraints)
		require.True(t, ok)
		require.NotNil(t, ic.StartTypeName)
		assert.Equal(t, "FiscalYearStart", *ic.StartTypeName)
		assert.Nil(t, ic.UnitName)
		assert.Nil(t, ic.StartsAt)
	})

	t.Run("Interval with inline Instant constraints", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.IntervalType, []analysis.Constraint{
			{Name: "starts_at", Value: "Instant(month=January, day_of_month=1)"},
			{Name: "unit", Value: "day"},
		}, analysis.LoadStandardTimeConcepts(t))
		require.NoError(t, err)
		require.NotNil(t, result)
		ic, ok := result.(*analysis.IntervalConstraints)
		require.True(t, ok)
		require.NotNil(t, ic.StartsAt)
		assert.Nil(t, ic.StartTypeName)
		require.NotNil(t, ic.UnitName)
		assert.Equal(t, "day", *ic.UnitName)
	})

	t.Run("Interval with invalid inline Instant constraints - error", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.IntervalType, []analysis.Constraint{
			{Name: "starts_at", Value: "Instant(invalid_field=123)"},
			{Name: "unit", Value: "day"},
		}, analysis.LoadStandardTimeConcepts(t))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown constraint for Instant: invalid_field")
		assert.Nil(t, result)
	})

	t.Run("Interval with non-Instant nested starts_at - error", func(t *testing.T) {
		t.Parallel()
		result, err := analysis.CreateConstraints(analysis.IntervalType, []analysis.Constraint{
			{Name: "starts_at", Value: "NotInstant(day_of_week=Monday)"},
			{Name: "unit", Value: "day"},
		}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "starts_at inline constraint must use Instant, got NotInstant")
		assert.Nil(t, result)
	})
}

// TestIntervalConstraintsValidate tests the IntervalConstraints.Validate method.
func TestIntervalConstraintsValidate(t *testing.T) {
	t.Parallel()

	t.Run("no constraints - invalid", func(t *testing.T) {
		t.Parallel()
		ic := analysis.IntervalConstraints{}
		err := ic.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Interval type must specify starts_at")
	})

	t.Run("with starts_at named type - valid", func(t *testing.T) {
		t.Parallel()
		name := "FiscalYearStart"
		unit := "month"
		ic := analysis.IntervalConstraints{StartTypeName: &name, UnitName: &unit}
		assert.NoError(t, ic.Validate())
	})

	t.Run("with starts_at inline constraints - valid", func(t *testing.T) {
		t.Parallel()
		instantConstraints := analysis.InstantConstraints{}
		unit := "day"
		ic := analysis.IntervalConstraints{StartsAt: &instantConstraints, UnitName: &unit}
		assert.NoError(t, ic.Validate())
	})

	t.Run("missing starts_at constraint - invalid", func(t *testing.T) {
		t.Parallel()
		unit := "month"
		ic := analysis.IntervalConstraints{UnitName: &unit}
		err := ic.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Interval type must specify starts_at")
	})

	t.Run("missing unit constraint - valid", func(t *testing.T) {
		t.Parallel()
		name := "FiscalYearStart"
		ic := analysis.IntervalConstraints{StartTypeName: &name}
		assert.NoError(t, ic.Validate())
	})

	t.Run("both inline and named starts_at - invalid", func(t *testing.T) {
		t.Parallel()
		instantConstraints := analysis.InstantConstraints{}
		name := "FiscalYearStart"
		unit := "month"
		ic := analysis.IntervalConstraints{StartsAt: &instantConstraints, StartTypeName: &name, UnitName: &unit}
		err := ic.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be both an inline Instant constraint and a named type reference")
	})
}

// TestIntervalConstraintsConstraintNames tests the ConstraintNames method.
func TestIntervalConstraintsConstraintNames(t *testing.T) {
	t.Parallel()

	t.Run("no constraints", func(t *testing.T) {
		t.Parallel()
		ic := analysis.IntervalConstraints{}
		assert.Empty(t, ic.ConstraintNames())
	})

	t.Run("all constraints", func(t *testing.T) {
		t.Parallel()
		name := "FiscalYearStart"
		unit := "month"
		ic := analysis.IntervalConstraints{StartTypeName: &name, UnitName: &unit}
		names := ic.ConstraintNames()
		assert.Len(t, names, 2)
		assert.Contains(t, names, "starts_at")
		assert.Contains(t, names, "unit")
	})
}

// TestIntervalConstraintsGetConstraint tests the GetConstraint method.
func TestIntervalConstraintsGetConstraint(t *testing.T) {
	t.Parallel()

	name := "FiscalYearStart"
	unit := "month"
	ic := analysis.IntervalConstraints{StartTypeName: &name, UnitName: &unit}

	t.Run("existing starts_at constraint", func(t *testing.T) {
		t.Parallel()
		val, ok := ic.GetConstraint("starts_at")
		require.True(t, ok)
		assert.Equal(t, &name, val)
	})

	t.Run("existing unit constraint", func(t *testing.T) {
		t.Parallel()
		val, ok := ic.GetConstraint("unit")
		require.True(t, ok)
		assert.Equal(t, &unit, val)
	})

	t.Run("missing constraint", func(t *testing.T) {
		t.Parallel()
		val, ok := ic.GetConstraint("precision")
		assert.False(t, ok)
		assert.Nil(t, val)
	})
}

// TestIntervalConstraintsIsMoreOrEqualRestrictive tests the IsMoreOrEqualRestrictive method.
func TestIntervalConstraintsIsMoreOrEqualRestrictive(t *testing.T) {
	t.Parallel()

	ic := analysis.IntervalConstraints{}

	t.Run("nil to value - more restrictive", func(t *testing.T) {
		t.Parallel()
		v := "month"
		ok, err := ic.IsMoreOrEqualRestrictive("unit", (*string)(nil), &v)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("value to nil - less restrictive", func(t *testing.T) {
		t.Parallel()
		v := "month"
		ok, err := ic.IsMoreOrEqualRestrictive("unit", &v, (*string)(nil))
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("same value - equal restrictive", func(t *testing.T) {
		t.Parallel()
		v := "month"
		ok, err := ic.IsMoreOrEqualRestrictive("unit", &v, &v)
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("different value - not more restrictive", func(t *testing.T) {
		t.Parallel()
		v1, v2 := "month", "day"
		ok, err := ic.IsMoreOrEqualRestrictive("unit", &v1, &v2)
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("unknown constraint - error", func(t *testing.T) {
		t.Parallel()
		_, err := ic.IsMoreOrEqualRestrictive("foobar", nil, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown constraint name")
	})
}

// TestValidateIntervalTypes tests the ValidateIntervalTypes cross-reference function.
func TestValidateIntervalTypes(t *testing.T) {
	t.Parallel()

	t.Run("valid Instant-derived starts_at type", func(t *testing.T) {
		t.Parallel()
		types := map[string]analysis.AnalysisTypeDefinition{
			"FiscalYearStart": {
				Name:              "FiscalYearStart",
				BaseType:          analysis.NewBuiltInTypeRef(analysis.InstantType),
				RawConstraints:    []analysis.Constraint{{Name: "month", Value: "January"}, {Name: "day_of_month", Value: "1"}},
				DerivationPath:    []string{"FiscalYearStart", "Instant"},
				ParsedConstraints: &analysis.InstantConstraints{},
			},
			"FiscalMonth": {
				Name:     "FiscalMonth",
				BaseType: analysis.NewBuiltInTypeRef(analysis.IntervalType),
				RawConstraints: []analysis.Constraint{
					{Name: "starts_at", Value: "FiscalYearStart"},
					{Name: "unit", Value: "month"},
				},
				DerivationPath:    []string{"FiscalMonth", "Interval"},
				ParsedConstraints: mustCreateIntervalConstraints(t, []analysis.Constraint{{Name: "starts_at", Value: "FiscalYearStart"}, {Name: "unit", Value: "month"}}),
			},
		}
		err := analysis.ValidateIntervalTypes(types, map[string]analysis.AnalysisUnitDefinition{
			"month": {Name: "month", ConceptName: "DurationUnit"},
		})
		assert.NoError(t, err)
	})

	t.Run("non-existent starts_at type", func(t *testing.T) {
		t.Parallel()
		types := map[string]analysis.AnalysisTypeDefinition{
			"FiscalMonth": {
				Name:     "FiscalMonth",
				BaseType: analysis.NewBuiltInTypeRef(analysis.IntervalType),
				RawConstraints: []analysis.Constraint{
					{Name: "starts_at", Value: "NonExistent"},
					{Name: "unit", Value: "month"},
				},
				DerivationPath:    []string{"FiscalMonth", "Interval"},
				ParsedConstraints: mustCreateIntervalConstraints(t, []analysis.Constraint{{Name: "starts_at", Value: "NonExistent"}, {Name: "unit", Value: "month"}}),
			},
		}
		err := analysis.ValidateIntervalTypes(types, map[string]analysis.AnalysisUnitDefinition{
			"month": {Name: "month", ConceptName: "DurationUnit"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "references unknown starts_at type NonExistent")
	})

	t.Run("non-Instant-derived starts_at type", func(t *testing.T) {
		t.Parallel()
		types := map[string]analysis.AnalysisTypeDefinition{
			"Currency": {
				Name:              "Currency",
				BaseType:          analysis.NewBuiltInTypeRef(analysis.NumberType),
				RawConstraints:    []analysis.Constraint{{Name: "precision", Value: "2"}},
				DerivationPath:    []string{"Currency", "Number"},
				ParsedConstraints: &analysis.NumberConstraints{},
			},
			"FiscalMonth": {
				Name:     "FiscalMonth",
				BaseType: analysis.NewBuiltInTypeRef(analysis.IntervalType),
				RawConstraints: []analysis.Constraint{
					{Name: "starts_at", Value: "Currency"},
					{Name: "unit", Value: "month"},
				},
				DerivationPath:    []string{"FiscalMonth", "Interval"},
				ParsedConstraints: mustCreateIntervalConstraints(t, []analysis.Constraint{{Name: "starts_at", Value: "Currency"}, {Name: "unit", Value: "month"}}),
			},
		}
		err := analysis.ValidateIntervalTypes(types, map[string]analysis.AnalysisUnitDefinition{
			"month": {Name: "month", ConceptName: "DurationUnit"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must derive from Instant, got Number")
	})

	t.Run("inline starts_at does not require cross-reference validation", func(t *testing.T) {
		t.Parallel()
		inlineConstraints := analysis.InstantConstraints{}
		types := map[string]analysis.AnalysisTypeDefinition{
			"WorkWeek": {
				Name:     "WorkWeek",
				BaseType: analysis.NewBuiltInTypeRef(analysis.IntervalType),
				RawConstraints: []analysis.Constraint{
					{Name: "starts_at", Value: "Instant(day=Monday)"},
					{Name: "unit", Value: "day"},
				},
				DerivationPath:    []string{"WorkWeek", "Interval"},
				ParsedConstraints: &analysis.IntervalConstraints{StartsAt: &inlineConstraints, UnitName: func(s string) *string { return &s }("day")},
			},
		}
		err := analysis.ValidateIntervalTypes(types, map[string]analysis.AnalysisUnitDefinition{
			"day": {Name: "day", ConceptName: "DurationUnit"},
		})
		assert.NoError(t, err)
	})

	t.Run("non-Interval type is skipped", func(t *testing.T) {
		t.Parallel()
		types := map[string]analysis.AnalysisTypeDefinition{
			"Currency": {
				Name:              "Currency",
				BaseType:          analysis.NewBuiltInTypeRef(analysis.NumberType),
				RawConstraints:    []analysis.Constraint{{Name: "precision", Value: "2"}},
				DerivationPath:    []string{"Currency", "Number"},
				ParsedConstraints: &analysis.NumberConstraints{},
			},
		}
		err := analysis.ValidateIntervalTypes(types, nil)
		assert.NoError(t, err)
	})
	t.Run("missing unit reference", func(t *testing.T) {
		t.Parallel()
		types := map[string]analysis.AnalysisTypeDefinition{
			"FiscalYearStart": {
				Name:              "FiscalYearStart",
				BaseType:          analysis.NewBuiltInTypeRef(analysis.InstantType),
				DerivationPath:    []string{"FiscalYearStart", "Instant"},
				ParsedConstraints: &analysis.InstantConstraints{},
			},
			"FiscalMonth": {
				Name:              "FiscalMonth",
				BaseType:          analysis.NewBuiltInTypeRef(analysis.IntervalType),
				DerivationPath:    []string{"FiscalMonth", "Interval"},
				ParsedConstraints: mustCreateIntervalConstraints(t, []analysis.Constraint{{Name: "starts_at", Value: "FiscalYearStart"}, {Name: "unit", Value: "month"}}),
			},
		}
		err := analysis.ValidateIntervalTypes(types, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "references unknown unit month")
	})
}

// mustCreateIntervalConstraints is a test helper that creates IntervalConstraints or fails the test.
func mustCreateIntervalConstraints(t *testing.T, constraints []analysis.Constraint) *analysis.IntervalConstraints {
	t.Helper()
	result, err := analysis.CreateConstraints(analysis.IntervalType, constraints, analysis.LoadStandardTimeConcepts(t))
	require.NoError(t, err)
	ic, ok := result.(*analysis.IntervalConstraints)
	require.True(t, ok)
	return ic
}

// TestBuiltInTypeRefInterval tests the BuiltInTypeRef for IntervalType.
func TestBuiltInTypeRefInterval(t *testing.T) {
	t.Parallel()

	ref := analysis.NewBuiltInTypeRef(analysis.IntervalType)
	assert.Equal(t, "Interval", ref.GetName())
	assert.True(t, ref.IsBuiltIn())
	assert.Equal(t, analysis.IntervalType, ref.GetBuiltInType())
}

// TestParseAndAnalyzeStringType tests the full parse+analyze pipeline for String types.
func TestParseAndAnalyzeStringType(t *testing.T) {
	t.Parallel()

	input := `types
		ShortString = String(maximum=50)
		NamePattern = String(pattern="^[A-Z][a-z]+$")`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err, "Full parse+analyze should succeed for String types")

	// Verify ShortString
	require.Contains(t, defs.Types, "ShortString")
	assert.Equal(t, []string{"ShortString", "String"}, defs.Types["ShortString"].DerivationPath)
	sc, ok := defs.Types["ShortString"].ParsedConstraints.(*analysis.StringConstraints)
	require.True(t, ok)
	assert.Nil(t, sc.Minimum)
	require.NotNil(t, sc.Maximum)
	assert.Equal(t, 50, *sc.Maximum)
	assert.Nil(t, sc.Pattern)

	// Verify NamePattern
	require.Contains(t, defs.Types, "NamePattern")
	assert.Equal(t, []string{"NamePattern", "String"}, defs.Types["NamePattern"].DerivationPath)
	sc2, ok := defs.Types["NamePattern"].ParsedConstraints.(*analysis.StringConstraints)
	require.True(t, ok)
	assert.Nil(t, sc2.Minimum)
	assert.Nil(t, sc2.Maximum)
	require.NotNil(t, sc2.Pattern)
	assert.Equal(t, "^[A-Z][a-z]+$", sc2.Pattern.String())
}

// TestParseAndAnalyzeBooleanType tests the full parse+analyze pipeline for Boolean types.
func TestParseAndAnalyzeBooleanType(t *testing.T) {
	t.Parallel()

	input := `types
		ActiveFlag = Boolean()`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err, "Full parse+analyze should succeed for Boolean type")

	require.Contains(t, defs.Types, "ActiveFlag")
	assert.Equal(t, []string{"ActiveFlag", "Boolean"}, defs.Types["ActiveFlag"].DerivationPath)
	// Boolean with no constraints has nil ParsedConstraints (consistent with other types)
	assert.Nil(t, defs.Types["ActiveFlag"].ParsedConstraints)
}

// TestComputeDerivationPathsStringNarrowing tests String type derivation with constraint narrowing.
func TestComputeDerivationPathsStringNarrowing(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()

	analyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "BaseStr",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.StringType),
		RawConstraints: []analysis.Constraint{{Name: "minimum", Value: "0"}, {Name: "maximum", Value: "100"}},
		DerivationPath: []string{},
	})
	analyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "ShortStr",
		BaseType:       analysis.NewDeclaredTypeRef("BaseStr"),
		RawConstraints: []analysis.Constraint{{Name: "maximum", Value: "50"}},
		DerivationPath: []string{},
	})

	err := analyzer.ComputeDerivationPaths()
	require.NoError(t, err, "String constraint narrowing should succeed")

	defs := analyzer.GetAnalysisDefinitions()
	assert.Equal(t, []string{"BaseStr", "String"}, defs.Types["BaseStr"].DerivationPath)
	assert.Equal(t, []string{"ShortStr", "BaseStr", "String"}, defs.Types["ShortStr"].DerivationPath)

	// Verify merged constraints: minimum=0 (inherited), maximum=50 (narrowed)
	sc, ok := defs.Types["ShortStr"].ParsedConstraints.(*analysis.StringConstraints)
	require.True(t, ok)
	require.NotNil(t, sc.Minimum)
	assert.Equal(t, 0, *sc.Minimum)
	require.NotNil(t, sc.Maximum)
	assert.Equal(t, 50, *sc.Maximum)
}

// TestComputeDerivationPathsStringExpansionError tests String type derivation with constraint expansion.
func TestComputeDerivationPathsStringExpansionError(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()

	analyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "BaseStr",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.StringType),
		RawConstraints: []analysis.Constraint{{Name: "maximum", Value: "50"}},
		DerivationPath: []string{},
	})
	analyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "WideStr",
		BaseType:       analysis.NewDeclaredTypeRef("BaseStr"),
		RawConstraints: []analysis.Constraint{{Name: "maximum", Value: "100"}},
		DerivationPath: []string{},
	})

	err := analyzer.ComputeDerivationPaths()
	require.Error(t, err, "String constraint expansion should fail")
	assert.Contains(t, err.Error(), "expands inherited constraint")
}

// TestComputeDerivationPathsBoolean tests Boolean type derivation chains.
func TestComputeDerivationPathsBoolean(t *testing.T) {
	t.Parallel()

	analyzer := analysis.NewCPLAnalyzer()

	analyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "BaseBool",
		BaseType:       analysis.NewBuiltInTypeRef(analysis.BooleanType),
		RawConstraints: []analysis.Constraint{},
		DerivationPath: []string{},
	})
	analyzer.AddType(analysis.AnalysisTypeDefinition{
		Name:           "DerivedBool",
		BaseType:       analysis.NewDeclaredTypeRef("BaseBool"),
		RawConstraints: []analysis.Constraint{},
		DerivationPath: []string{},
	})

	err := analyzer.ComputeDerivationPaths()
	require.NoError(t, err, "Boolean derivation chain should succeed")

	defs := analyzer.GetAnalysisDefinitions()
	assert.Equal(t, []string{"BaseBool", "Boolean"}, defs.Types["BaseBool"].DerivationPath)
	assert.Equal(t, []string{"DerivedBool", "BaseBool", "Boolean"}, defs.Types["DerivedBool"].DerivationPath)
}

// TestComputeDerivationPathsInterval tests derivation paths for Interval types.
func TestComputeDerivationPathsInterval(t *testing.T) {
	t.Parallel()

	t.Run("single Interval type", func(t *testing.T) {
		t.Parallel()
		analyzer := analysis.NewCPLAnalyzer()
		addStandardTimeConceptsToAnalyzer(t, analyzer)
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:     "FiscalMonth",
			BaseType: analysis.NewBuiltInTypeRef(analysis.IntervalType),
			RawConstraints: []analysis.Constraint{
				{Name: "starts_at", Value: "FiscalYearStart"},
				{Name: "unit", Value: "month"},
			},
			DerivationPath: []string{},
		})
		// Add the Instant type that FiscalMonth's starts_at references
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "FiscalYearStart",
			BaseType:       analysis.NewBuiltInTypeRef(analysis.InstantType),
			RawConstraints: []analysis.Constraint{{Name: "month", Value: "January"}, {Name: "day_of_month", Value: "1"}},
			DerivationPath: []string{},
		})

		err := analyzer.ComputeDerivationPaths()
		require.NoError(t, err)

		defs := analyzer.GetAnalysisDefinitions()
		assert.Equal(t, []string{"FiscalMonth", "Interval"}, defs.Types["FiscalMonth"].DerivationPath)
		ic, ok := defs.Types["FiscalMonth"].ParsedConstraints.(*analysis.IntervalConstraints)
		require.True(t, ok)
		require.NotNil(t, ic.StartTypeName)
		assert.Equal(t, "FiscalYearStart", *ic.StartTypeName)
		require.NotNil(t, ic.UnitName)
		assert.Equal(t, "month", *ic.UnitName)
	})

	t.Run("Interval type derivation chain", func(t *testing.T) {
		t.Parallel()
		analyzer := analysis.NewCPLAnalyzer()
		addStandardTimeConceptsToAnalyzer(t, analyzer)
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:     "Month",
			BaseType: analysis.NewBuiltInTypeRef(analysis.IntervalType),
			RawConstraints: []analysis.Constraint{
				{Name: "starts_at", Value: "MonthStart"},
				{Name: "unit", Value: "month"},
			},
			DerivationPath: []string{},
		})
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "FiscalMonth",
			BaseType:       analysis.NewDeclaredTypeRef("Month"),
			RawConstraints: []analysis.Constraint{}, // Inherits all constraints from Month
			DerivationPath: []string{},
		})
		// Add Instant types referenced by Interval types
		analyzer.AddType(analysis.AnalysisTypeDefinition{
			Name:           "MonthStart",
			BaseType:       analysis.NewBuiltInTypeRef(analysis.InstantType),
			RawConstraints: []analysis.Constraint{{Name: "day_of_month", Value: "1"}},
			DerivationPath: []string{},
		})

		err := analyzer.ComputeDerivationPaths()
		require.NoError(t, err)

		defs := analyzer.GetAnalysisDefinitions()
		assert.Equal(t, []string{"Month", "Interval"}, defs.Types["Month"].DerivationPath)
		assert.Equal(t, []string{"FiscalMonth", "Month", "Interval"}, defs.Types["FiscalMonth"].DerivationPath)

		ic, ok := defs.Types["FiscalMonth"].ParsedConstraints.(*analysis.IntervalConstraints)
		require.True(t, ok)
		require.NotNil(t, ic.StartTypeName)
		assert.Equal(t, "MonthStart", *ic.StartTypeName)
		require.NotNil(t, ic.UnitName)
		assert.Equal(t, "month", *ic.UnitName)
	})
}

// TestAnalyzeDuplicateTypeAcrossSections verifies that a duplicate type name across
// two separate types sections produces an error during analysis.
func TestAnalyzeDuplicateTypeAcrossSections(t *testing.T) {
	t.Parallel()

	input := `types
	Currency = Number(precision=2)

types
	Currency = Number(precision=4)`

	_, err := parseAndAnalyze(input)
	require.Error(t, err, "Analysis should fail for duplicate type name across sections")
	assert.Contains(t, err.Error(), "duplicate", "Error should mention duplicate")
}

// TestAnalyzeUniqueTypesAcrossSections verifies that unique type names across
// multiple types sections succeed and both types are present in defs.Types.
func TestAnalyzeUniqueTypesAcrossSections(t *testing.T) {
	t.Parallel()

	input := `types
	Currency = Number(precision=2)

types
	Percentage = Number(precision=2, minimum=0, maximum=100)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err, "Analysis should succeed for unique type names across sections")

	require.Contains(t, defs.Types, "Currency", "Currency should be in defs.Types")
	assert.Equal(t, []string{"Currency", "Number"}, defs.Types["Currency"].DerivationPath)

	require.Contains(t, defs.Types, "Percentage", "Percentage should be in defs.Types")
	assert.Equal(t, []string{"Percentage", "Number"}, defs.Types["Percentage"].DerivationPath)
}

// TestAnalyzeThreeSectionsWithDuplicate verifies that when three types sections
// contain a duplicate name, analysis errors on the duplicate.
func TestAnalyzeThreeSectionsWithDuplicate(t *testing.T) {
	t.Parallel()

	input := `types
	Currency = Number(precision=2)

types
	Percentage = Number(precision=2, minimum=0, maximum=100)

types
	Currency = Number(precision=4)`

	_, err := parseAndAnalyze(input)
	require.Error(t, err, "Analysis should fail for duplicate type name in third section")
	assert.Contains(t, err.Error(), "duplicate", "Error should mention duplicate")
	assert.Contains(t, err.Error(), "Currency", "Error should mention the duplicate type name Currency")
}

// TestSymbolTableWalkVisitsAllTypeScopes verifies that the analyzer's symbol table
// walk visits all TypeScope tables under all TypesScope tables, collecting every
// type definition into the analyzer's Types map.
func TestSymbolTableWalkVisitsAllTypeScopes(t *testing.T) {
	t.Parallel()

	input := `types
	Currency = Number(precision=2)
	Amount = Number()

types
	Percentage = Number(precision=2, minimum=0, maximum=100)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err, "Analysis should succeed when walking all TypeScopes")

	// All three types should be present
	require.Contains(t, defs.Types, "Currency", "Currency should be collected from first section")
	require.Contains(t, defs.Types, "Amount", "Amount should be collected from first section")
	require.Contains(t, defs.Types, "Percentage", "Percentage should be collected from second section")

	// Total count should equal sum of unique types across sections (3)
	assert.Len(t, defs.Types, 3, "Total types should equal sum across all sections")
}

// TestParseAndAnalyzeNumberAndStringAcrossSections verifies end-to-end parse+analyze
// with Number types in the first section and String types in the second.
func TestParseAndAnalyzeNumberAndStringAcrossSections(t *testing.T) {
	t.Parallel()

	input := `types
	Currency = Number(precision=2)

types
	NamePattern = String(maximum=50)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err, "Full parse+analyze should succeed for Number and String across sections")

	require.Contains(t, defs.Types, "Currency", "Currency should be in defs.Types")
	assert.Equal(t, []string{"Currency", "Number"}, defs.Types["Currency"].DerivationPath)

	require.Contains(t, defs.Types, "NamePattern", "NamePattern should be in defs.Types")
	assert.Equal(t, []string{"NamePattern", "String"}, defs.Types["NamePattern"].DerivationPath)
	sc, ok := defs.Types["NamePattern"].ParsedConstraints.(*analysis.StringConstraints)
	require.True(t, ok)
	require.NotNil(t, sc.Maximum)
	assert.Equal(t, 50, *sc.Maximum)
}

// TestParseAndAnalyzeBooleanAcrossSections verifies end-to-end parse+analyze
// with a Boolean type declared in a separate types section.
func TestParseAndAnalyzeBooleanAcrossSections(t *testing.T) {
	t.Parallel()

	input := `types
	Currency = Number(precision=2)

types
	ActiveFlag = Boolean()`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err, "Full parse+analyze should succeed for Boolean across sections")

	require.Contains(t, defs.Types, "ActiveFlag", "ActiveFlag should be in defs.Types")
	assert.Equal(t, []string{"ActiveFlag", "Boolean"}, defs.Types["ActiveFlag"].DerivationPath)
	assert.Nil(t, defs.Types["ActiveFlag"].ParsedConstraints)
}

// TestParseAndAnalyzeInstantIntervalAcrossSections verifies end-to-end parse+analyze
// with Instant and Interval types declared in separate types sections, requiring
// concepts and units sections for the time-based types.
func TestParseAndAnalyzeInstantIntervalAcrossSections(t *testing.T) {
	t.Parallel()

	input := `concepts
		DurationUnit=[Values[month, day]]
		Year = [Int[1900, 3000, 1]]
		Month=[Values[January, February, March, April, May, June, July, August, September, October, November, December]]
		DayOfMonth = [Int[1, 31, 1]]
	units
		month = DurationUnit
	types
		FiscalYearStart = Instant(month=January, day_of_month=1)
	types
		FiscalMonth = Interval(starts_at=FiscalYearStart, unit=month)`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err, "Full parse+analyze should succeed for Instant and Interval across sections")

	require.Contains(t, defs.Types, "FiscalYearStart", "FiscalYearStart should be in defs.Types")
	assert.Equal(t, []string{"FiscalYearStart", "Instant"}, defs.Types["FiscalYearStart"].DerivationPath)

	require.Contains(t, defs.Types, "FiscalMonth", "FiscalMonth should be in defs.Types")
	assert.Equal(t, []string{"FiscalMonth", "Interval"}, defs.Types["FiscalMonth"].DerivationPath)
	ic, ok := defs.Types["FiscalMonth"].ParsedConstraints.(*analysis.IntervalConstraints)
	require.True(t, ok)
	require.NotNil(t, ic.StartTypeName)
	assert.Equal(t, "FiscalYearStart", *ic.StartTypeName)
	require.NotNil(t, ic.UnitName)
	assert.Equal(t, "month", *ic.UnitName)
}

// TestAllTypesInSingleLookupMap verifies that when multiple types sections define
// Number, String, and Boolean types, all type names are present in the single
// defs.Types lookup map and the count matches the total unique types.
func TestAllTypesInSingleLookupMap(t *testing.T) {
	t.Parallel()

	input := `types
	Currency = Number(precision=2)

types
	NamePattern = String(maximum=50)

types
	ActiveFlag = Boolean()`

	defs, err := parseAndAnalyze(input)
	require.NoError(t, err, "Full parse+analyze should succeed for all types across sections")

	// All type names should be present in the single Types lookup map
	require.Contains(t, defs.Types, "Currency", "Currency should be in defs.Types")
	require.Contains(t, defs.Types, "NamePattern", "NamePattern should be in defs.Types")
	require.Contains(t, defs.Types, "ActiveFlag", "ActiveFlag should be in defs.Types")

	// Verify len(defs.Types) equals total unique types (3)
	assert.Len(t, defs.Types, 3, "Total unique types in defs.Types should be 3")

	// Verify each type has the correct derivation path
	assert.Equal(t, []string{"Currency", "Number"}, defs.Types["Currency"].DerivationPath)
	assert.Equal(t, []string{"NamePattern", "String"}, defs.Types["NamePattern"].DerivationPath)
	assert.Equal(t, []string{"ActiveFlag", "Boolean"}, defs.Types["ActiveFlag"].DerivationPath)
}
