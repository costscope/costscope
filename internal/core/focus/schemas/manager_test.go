package schemas

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSchema_Success(t *testing.T) {
	m := NewManager()

	schema, err := m.GetSchema("focus-1.2")
	assert.NoError(t, err)
	assert.NotNil(t, schema)
	assert.Equal(t, "1.2", schema.Version)

	// spot check a couple of required columns exist
	if assert.Contains(t, schema.RequiredColumns, "BillingAccountId") {
		assert.Equal(t, "string", schema.RequiredColumns["BillingAccountId"].Type)
		assert.True(t, schema.RequiredColumns["BillingAccountId"].Required)
	}
	if assert.Contains(t, schema.RequiredColumns, "EffectiveCost") {
		assert.Equal(t, "decimal", schema.RequiredColumns["EffectiveCost"].Type)
		assert.True(t, schema.RequiredColumns["EffectiveCost"].Required)
	}
}

func TestGetSchema_NotFound(t *testing.T) {
	m := NewManager()
	schema, err := m.GetSchema("does-not-exist")
	assert.Error(t, err)
	assert.Nil(t, schema)
}

func TestGetAvailableSchemas(t *testing.T) {
	m := NewManager()
	names := m.GetAvailableSchemas()

	// Order is not guaranteed; assert set membership and count
	assert.Len(t, names, 3)
	assert.Contains(t, names, "focus-1.2")
	assert.Contains(t, names, "focus-1.1")
	assert.Contains(t, names, "focus-1.0")
}

func TestValidateSchemaCompatibility_AllMatch_NoIssues(t *testing.T) {
	m := NewManager()
	base, err := m.GetSchema("focus-1.2")
	assert.NoError(t, err)

	// Build a schema that mirrors all required columns from base
	required := make(map[string]ColumnSpec, len(base.RequiredColumns))
	for k, v := range base.RequiredColumns {
		required[k] = v // copy preserves type/name/required flags
	}

	s := &FOCUSSchema{
		Version:         "custom",
		RequiredColumns: required,
		OptionalColumns: map[string]ColumnSpec{},
		Dimensions:      nil,
		Metrics:         nil,
		Metadata:        SchemaMetadata{Title: "custom"},
	}

	issues, err := m.ValidateSchemaCompatibility(s, "focus-1.2")
	assert.NoError(t, err)
	assert.Empty(t, issues)
}

func TestValidateSchemaCompatibility_MissingAndTypeMismatch(t *testing.T) {
	m := NewManager()
	base, err := m.GetSchema("focus-1.2")
	assert.NoError(t, err)

	// Start with a full copy then introduce problems
	required := make(map[string]ColumnSpec, len(base.RequiredColumns))
	for k, v := range base.RequiredColumns {
		required[k] = v
	}

	// 1) Remove a required column
	delete(required, "BillingAccountId")

	// 2) Type mismatch for an existing required column
	if col, ok := required["EffectiveCost"]; ok {
		col.Type = "string" // should be decimal
		required["EffectiveCost"] = col
	}

	s := &FOCUSSchema{
		Version:         "broken",
		RequiredColumns: required,
		OptionalColumns: map[string]ColumnSpec{},
		Metadata:        SchemaMetadata{Title: "broken"},
	}

	issues, err := m.ValidateSchemaCompatibility(s, "focus-1.2")
	assert.NoError(t, err)
	// Expect at least two issues: missing required column and type mismatch
	assert.GreaterOrEqual(t, len(issues), 2)
}
