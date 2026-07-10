package symboltable

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDefinitionData_InterfaceSatisfaction verifies that all concrete carrier
// types satisfy the DefinitionData interface.
func TestDefinitionData_InterfaceSatisfaction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data DefinitionData
	}{
		{"StringData", StringData{Value: "hello"}},
		{"IntData", IntData{Value: 42}},
		{"FloatData", FloatData{Value: 3.14}},
		{"BoolData", BoolData{Value: true}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.NotNil(t, tc.data)
		})
	}
}

// TestDefinitionData_ConstructionAndFieldAccess verifies that each carrier
// can be constructed and its field accessed correctly.
func TestDefinitionData_ConstructionAndFieldAccess(t *testing.T) {
	t.Parallel()

	t.Run("StringData", func(t *testing.T) {
		d := StringData{Value: "test"}
		assert.Equal(t, "test", d.Value)
	})

	t.Run("IntData", func(t *testing.T) {
		d := IntData{Value: 123}
		assert.Equal(t, 123, d.Value)
	})

	t.Run("FloatData", func(t *testing.T) {
		d := FloatData{Value: 2.718}
		assert.Equal(t, 2.718, d.Value)
	})

	t.Run("BoolData", func(t *testing.T) {
		d := BoolData{Value: false}
		assert.Equal(t, false, d.Value)
	})
}

// TestDefinitionData_NilSatisfiesInterface verifies that a nil value can be
// assigned to the DefinitionData interface (since it's an interface type).
func TestDefinitionData_NilSatisfiesInterface(t *testing.T) {
	t.Parallel()

	var d DefinitionData
	assert.Nil(t, d)
}
