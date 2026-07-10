// Package symboltable provides symbol table data structures for the CPL parser.
//
// This file contains unit tests for the symbol table types and operations.
// Tests follow table-driven patterns with testify assertions and parallel execution.
package symboltable

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScopeTypeString verifies the String() method for ScopeType.
func TestScopeTypeString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		scope    ScopeType
		expected string
	}{
		{
			name:     "GlobalScope returns GlobalScope",
			scope:    GlobalScope,
			expected: "GlobalScope",
		},
		{
			name:     "UnitsScope returns UnitsScope",
			scope:    UnitsScope,
			expected: "UnitsScope",
		},
		{
			name:     "ConnectorsScope returns ConnectorsScope",
			scope:    ConnectorsScope,
			expected: "ConnectorsScope",
		},
		{
			name:     "ConnectorScope returns ConnectorScope",
			scope:    ConnectorScope,
			expected: "ConnectorScope",
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := tc.scope.String()
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestDefinitionTypeString verifies the String() method for DefinitionType.
func TestDefinitionTypeString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		defType  DefinitionType
		expected string
	}{
		{
			name:     "UnitDefinition returns UnitDefinition",
			defType:  UnitDefinition,
			expected: "UnitDefinition",
		},
		{
			name:     "ConnectorsScopeDefinition returns ConnectorsScopeDefinition",
			defType:  ConnectorsScopeDefinition,
			expected: "ConnectorsScopeDefinition",
		},
		{
			name:     "ConnectorScopeDefinition returns ConnectorScopeDefinition",
			defType:  ConnectorScopeDefinition,
			expected: "ConnectorScopeDefinition",
		},
		{
			name:     "GrainDefinition returns GrainDefinition",
			defType:  GrainDefinition,
			expected: "GrainDefinition",
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := tc.defType.String()
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestSourceLocationCreation verifies SourceLocation struct initialization.
func TestSourceLocationCreation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		line         int
		column       int
		expectedLine int
		expectedCol  int
	}{
		{
			name:         "Valid line and column",
			line:         10,
			column:       5,
			expectedLine: 10,
			expectedCol:  5,
		},
		{
			name:         "Zero values",
			line:         0,
			column:       0,
			expectedLine: 0,
			expectedCol:  0,
		},
		{
			name:         "Large values",
			line:         999999,
			column:       888888,
			expectedLine: 999999,
			expectedCol:  888888,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			loc := SourceLocation{
				Line:   tc.line,
				Column: tc.column,
			}
			assert.Equal(t, tc.expectedLine, loc.Line)
			assert.Equal(t, tc.expectedCol, loc.Column)
		})
	}
}

// TestSymbolTableErrorCreation verifies SymbolTableError struct and Error() method.
func TestSymbolTableErrorCreation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		code            ErrorCode
		message         string
		symbolName      string
		expectedMessage string
		expectedCode    ErrorCode
		expectedName    string
	}{
		{
			name:            "Duplicate definition error",
			code:            ErrDuplicateDefinition,
			message:         "symbol 'meter' already defined",
			symbolName:      "meter",
			expectedMessage: "symbol 'meter' already defined",
			expectedCode:    ErrDuplicateDefinition,
			expectedName:    "meter",
		},
		{
			name:            "Symbol not found error",
			code:            ErrSymbolNotFound,
			message:         "symbol 'kilogram' not found",
			symbolName:      "kilogram",
			expectedMessage: "symbol 'kilogram' not found",
			expectedCode:    ErrSymbolNotFound,
			expectedName:    "kilogram",
		},
		{
			name:            "Empty stack error",
			code:            ErrEmptyStack,
			message:         "stack is empty",
			symbolName:      "",
			expectedMessage: "stack is empty",
			expectedCode:    ErrEmptyStack,
			expectedName:    "",
		},
		{
			name:            "Unbalanced stack error",
			code:            ErrUnbalancedStack,
			message:         "unbalanced stack: 3 tables remaining",
			symbolName:      "",
			expectedMessage: "unbalanced stack: 3 tables remaining",
			expectedCode:    ErrUnbalancedStack,
			expectedName:    "",
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := &SymbolTableError{
				Code:    tc.code,
				Message: tc.message,
				Name:    tc.symbolName,
			}

			assert.Equal(t, tc.expectedMessage, err.Error())
			assert.Equal(t, tc.expectedCode, err.Code)
			assert.Equal(t, tc.expectedName, err.Name)

			// Verify it implements the error interface
			var _ error = err
		})
	}
}

// TestSymbolTableErrorCodeUniqueness verifies that error codes are distinct.
func TestSymbolTableErrorCodeUniqueness(t *testing.T) {
	t.Parallel()

	codes := []ErrorCode{
		ErrDuplicateDefinition,
		ErrSymbolNotFound,
		ErrUnbalancedStack,
		ErrEmptyStack,
	}

	seen := make(map[ErrorCode]bool)
	for _, code := range codes {
		_, exists := seen[code]
		assert.False(t, exists, "Error code %d should be unique", code)
		seen[code] = true
	}

	assert.Equal(t, 4, len(seen), "Should have 4 distinct error codes")
}

// TestNewSymbolTable verifies the NewSymbolTable constructor.
func TestNewSymbolTable(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		scopeType ScopeType
	}{
		{
			name:      "Create global scope table",
			scopeType: GlobalScope,
		},
		{
			name:      "Create units scope table",
			scopeType: UnitsScope,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			table := NewSymbolTable(tc.scopeType)

			require.NotNil(t, table)
			assert.Equal(t, tc.scopeType, table.ScopeType)
			assert.Nil(t, table.Parent)
			assert.Empty(t, table.Children)
			assert.Empty(t, table.Definitions)
		})
	}
}

// TestNewGlobalSymbolTable verifies the NewGlobalSymbolTable convenience constructor.
func TestNewGlobalSymbolTable(t *testing.T) {
	t.Parallel()

	table := NewGlobalSymbolTable()

	require.NotNil(t, table)
	assert.Equal(t, GlobalScope, table.ScopeType)
	assert.Nil(t, table.Parent)
	assert.Empty(t, table.Children)
	assert.Empty(t, table.Definitions)
}

// TestSymbolDefinitionCreation verifies SymbolDefinition struct initialization.
func TestSymbolDefinitionCreation(t *testing.T) {
	t.Parallel()

	name := "meter"
	location := SourceLocation{Line: 5, Column: 10}

	testCases := []struct {
		name       string
		symbolName *string
		defType    DefinitionType
		location   SourceLocation
		data       DefinitionData
	}{
		{
			name:       "Named definition with data",
			symbolName: &name,
			defType:    UnitDefinition,
			location:   location,
			data:       StringData{Value: "some data"},
		},
		{
			name:       "Named definition without data",
			symbolName: &name,
			defType:    UnitDefinition,
			location:   location,
			data:       nil,
		},
		{
			name:       "Anonymous definition",
			symbolName: nil,
			defType:    UnitDefinition,
			location:   location,
			data:       nil,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			def := SymbolDefinition{
				Name:           tc.symbolName,
				DefinitionType: tc.defType,
				Location:       tc.location,
				Data:           tc.data,
			}

			assert.Equal(t, tc.symbolName, def.Name)
			assert.Equal(t, tc.defType, def.DefinitionType)
			assert.Equal(t, tc.location, def.Location)
			assert.Equal(t, tc.data, def.Data)
		})
	}
}

// TestSymbolTableAddChild verifies the AddChild method creates proper parent-child relationships.
func TestSymbolTableAddChild(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		parentScope      ScopeType
		childScope       ScopeType
		expectedChildren int
	}{
		{
			name:             "Global parent with units child",
			parentScope:      GlobalScope,
			childScope:       UnitsScope,
			expectedChildren: 1,
		},
		{
			name:             "Global parent with global child",
			parentScope:      GlobalScope,
			childScope:       GlobalScope,
			expectedChildren: 1,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			parent := NewSymbolTable(tc.parentScope)
			child := parent.AddChild(tc.childScope)

			require.NotNil(t, child)
			assert.Equal(t, tc.childScope, child.ScopeType)
			assert.Equal(t, parent, child.Parent)
			assert.Len(t, parent.Children, tc.expectedChildren)
			assert.Contains(t, parent.Children, child)
		})
	}
}

// TestSymbolTableDefine verifies the Define method for adding symbols.
func TestSymbolTableDefine(t *testing.T) {
	t.Parallel()

	name := "meter"
	location := SourceLocation{Line: 5, Column: 10}

	testCases := []struct {
		name      string
		symName   *string
		defType   DefinitionType
		location  SourceLocation
		expectErr bool
		errCode   ErrorCode
	}{
		{
			name:      "Define named symbol",
			symName:   &name,
			defType:   UnitDefinition,
			location:  location,
			expectErr: false,
		},
		{
			name:      "Define anonymous symbol",
			symName:   nil,
			defType:   UnitDefinition,
			location:  location,
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			table := NewSymbolTable(GlobalScope)
			err := table.Define(tc.symName, tc.defType, tc.location)

			if tc.expectErr {
				require.Error(t, err)
				var symErr *SymbolTableError
				require.True(t, errors.As(err, &symErr))
				assert.Equal(t, tc.errCode, symErr.Code)
			} else {
				require.NoError(t, err)
				require.Len(t, table.Definitions, 1)
				assert.Equal(t, tc.symName, table.Definitions[0].Name)
				assert.Equal(t, tc.defType, table.Definitions[0].DefinitionType)
				assert.Equal(t, tc.location, table.Definitions[0].Location)
			}
		})
	}
}

// TestSymbolTableLookupLocal verifies local-only symbol lookup.
func TestSymbolTableLookupLocal(t *testing.T) {
	t.Parallel()

	name := "meter"
	location := SourceLocation{Line: 5, Column: 10}

	testCases := []struct {
		name        string
		lookupName  string
		expectFound bool
		expectErr   bool
	}{
		{
			name:        "Lookup existing symbol",
			lookupName:  "meter",
			expectFound: true,
			expectErr:   false,
		},
		{
			name:        "Lookup non-existent symbol",
			lookupName:  "kilogram",
			expectFound: false,
			expectErr:   true,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			table := NewSymbolTable(GlobalScope)
			err := table.Define(&name, UnitDefinition, location)
			require.NoError(t, err)

			def, err := table.LookupLocal(tc.lookupName)

			if tc.expectErr {
				require.Error(t, err)
				var symErr *SymbolTableError
				require.True(t, errors.As(err, &symErr))
				assert.Equal(t, ErrSymbolNotFound, symErr.Code)
				assert.Nil(t, def)
			} else {
				require.NoError(t, err)
				require.NotNil(t, def)
				assert.Equal(t, &name, def.Name)
			}
		})
	}
}

// TestSymbolTableLookup verifies hierarchical symbol lookup.
func TestSymbolTableLookup(t *testing.T) {
	t.Parallel()

	parentName := "meter"
	childName := "kilogram"
	parentLoc := SourceLocation{Line: 5, Column: 10}
	childLoc := SourceLocation{Line: 10, Column: 15}

	testCases := []struct {
		name        string
		lookupName  string
		expectFound bool
		expectErr   bool
		expectedDef *string // pointer to expected name
	}{
		{
			name:        "Lookup in child scope finds child symbol",
			lookupName:  "kilogram",
			expectFound: true,
			expectErr:   false,
			expectedDef: &childName,
		},
		{
			name:        "Lookup in child scope finds parent symbol",
			lookupName:  "meter",
			expectFound: true,
			expectErr:   false,
			expectedDef: &parentName,
		},
		{
			name:        "Lookup non-existent symbol",
			lookupName:  "second",
			expectFound: false,
			expectErr:   true,
			expectedDef: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			parent := NewSymbolTable(GlobalScope)
			err := parent.Define(&parentName, UnitDefinition, parentLoc)
			require.NoError(t, err)

			child := parent.AddChild(UnitsScope)
			err = child.Define(&childName, UnitDefinition, childLoc)
			require.NoError(t, err)

			def, err := child.Lookup(tc.lookupName)

			if tc.expectErr {
				require.Error(t, err)
				var symErr *SymbolTableError
				require.True(t, errors.As(err, &symErr))
				assert.Equal(t, ErrSymbolNotFound, symErr.Code)
				assert.Nil(t, def)
			} else {
				require.NoError(t, err)
				require.NotNil(t, def)
				if tc.expectedDef != nil {
					assert.Equal(t, *tc.expectedDef, *def.Name)
				}
			}
		})
	}
}

// TestSymbolTableDefineWithData_GrainDefinition verifies that DefineWithData accepts
// GrainDefinition and stores it correctly in the Definitions slice.
func TestSymbolTableDefineWithData_GrainDefinition(t *testing.T) {
	t.Parallel()

	name := "Day"
	location := SourceLocation{Line: 5, Column: 10}
	data := StringData{Value: "grain-data"}

	table := NewSymbolTable(ConnectorScope)
	err := table.DefineWithData(&name, GrainDefinition, location, data)

	require.NoError(t, err)
	require.Len(t, table.Definitions, 1)
	assert.Equal(t, &name, table.Definitions[0].Name)
	assert.Equal(t, GrainDefinition, table.Definitions[0].DefinitionType)
	assert.Equal(t, location, table.Definitions[0].Location)
	assert.Equal(t, data, table.Definitions[0].Data)
}

// TestSymbolTableLookupShadowing verifies that child definitions shadow parent definitions.
func TestSymbolTableLookupShadowing(t *testing.T) {
	t.Parallel()

	sharedName := "unit"
	parentLoc := SourceLocation{Line: 5, Column: 10}
	childLoc := SourceLocation{Line: 10, Column: 15}

	parent := NewSymbolTable(GlobalScope)
	err := parent.Define(&sharedName, UnitDefinition, parentLoc)
	require.NoError(t, err)

	child := parent.AddChild(UnitsScope)
	err = child.Define(&sharedName, UnitDefinition, childLoc)
	require.NoError(t, err)

	// Lookup from child should find child's definition
	def, err := child.Lookup(sharedName)
	require.NoError(t, err)
	require.NotNil(t, def)
	assert.Equal(t, childLoc.Line, def.Location.Line)
	assert.Equal(t, childLoc.Column, def.Location.Column)
}
