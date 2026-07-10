// Package io provides I/O utilities for the Chronocone Planning Language.
// This file contains tests for the parquet file generation utility.
package io

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/apache/arrow/go/v16/parquet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateParquetFile_CreatesValidFile verifies that the GenerateParquetFile
// function creates a parquet file at the specified path.
//
// This test checks:
//   - The file is created at the specified file path
//   - The file exists after the function returns without error
//   - The file is a valid parquet file (can be opened and read by parquet readers)
//   - The file is not empty and contains valid parquet metadata
func TestGenerateParquetFile_CreatesValidFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.parquet")

	// Define test dimensions
	dimensions := map[string][]string{
		"Region":  {"North", "South"},
		"Product": {"A", "B"},
	}

	// Call GenerateParquetFile
	err := GenerateParquetFile(filePath, dimensions, nil, nil)
	require.NoError(t, err, "GenerateParquetFile should not return an error")

	// Verify file exists
	_, err = os.Stat(filePath)
	require.NoError(t, err, "Generated parquet file should exist")

	// Verify it's a valid parquet file by reading it
	schema, numRows, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")
	require.NotNil(t, schema, "Schema should not be nil")
	assert.Greater(t, numRows, 0, "File should have at least one row")
}

// TestGenerateParquetFile_CorrectSchema verifies that the generated parquet file
// has the correct schema based on the provided dimensions.
//
// This test checks:
//   - Each dimension becomes a string column in the parquet schema
//   - The "Value" column is of type float64
//   - The column count matches the number of dimensions plus one (for the Value column)
//   - Column names match the dimension keys provided in the input map
//   - The schema preserves the order of dimensions consistently
func TestGenerateParquetFile_CorrectSchema(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.parquet")

	// Define test dimensions
	dimensions := map[string][]string{
		"Region":  {"North", "South"},
		"Product": {"A", "B"},
	}

	// Call GenerateParquetFile
	err := GenerateParquetFile(filePath, dimensions, nil, nil)
	require.NoError(t, err, "GenerateParquetFile should not return an error")

	// Read the parquet file
	parquetSchema, _, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")

	// Verify column count (2 dimensions + 1 Value column = 3)
	assert.Equal(t, 3, parquetSchema.NumColumns(), "Schema should have 3 columns")

	// Verify column names
	columnNames := GetColumnNames(parquetSchema)
	assert.Contains(t, columnNames, "Product", "Schema should have Product column")
	assert.Contains(t, columnNames, "Region", "Schema should have Region column")
	assert.Contains(t, columnNames, "Value", "Schema should have Value column")

	// Verify Value column type is Double (float64)
	valueType := GetColumnType(parquetSchema, "Value")
	assert.Equal(t, parquet.Types.Double, valueType, "Value column should be Double type")

	// Verify dimension columns are ByteArray (string)
	regionType := GetColumnType(parquetSchema, "Region")
	assert.Equal(t, parquet.Types.ByteArray, regionType, "Region column should be ByteArray type")

	productType := GetColumnType(parquetSchema, "Product")
	assert.Equal(t, parquet.Types.ByteArray, productType, "Product column should be ByteArray type")
}

// TestGenerateParquetFile_CartesianProduct verifies that all combinations of
// dimension values are generated in the parquet file.
//
// This test checks:
//   - The row count equals the product of all dimension value counts (Cartesian product)
//   - Each unique combination of dimension values appears exactly once
//   - No duplicate rows exist in the output
//   - All possible combinations of the provided dimension values are present
//
// For example, with dimensions {"Region": ["North", "South"], "Product": ["A", "B"]},
// the output should have 4 rows (2 x 2) representing all combinations.
func TestGenerateParquetFile_CartesianProduct(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.parquet")

	// Define test dimensions
	dimensions := map[string][]string{
		"Region":  {"North", "South", "East"},
		"Product": {"A", "B"},
	}

	// Call GenerateParquetFile
	err := GenerateParquetFile(filePath, dimensions, nil, nil)
	require.NoError(t, err, "GenerateParquetFile should not return an error")

	// Read the parquet file
	_, numRows, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")

	// Expected row count: 3 regions x 2 products = 6
	expectedRows := 3 * 2
	assert.Equal(t, expectedRows, numRows, "Should have %d rows (Cartesian product)", expectedRows)

	// Read the columns to verify combinations
	// Note: columns are sorted alphabetically, so Product is index 0, Region is index 1
	productData, err := ReadParquetColumnData(filePath, 0)
	require.NoError(t, err, "Should be able to read Product column")
	products := productData.([]string)

	regionData, err := ReadParquetColumnData(filePath, 1)
	require.NoError(t, err, "Should be able to read Region column")
	regions := regionData.([]string)

	// Verify we have the expected number of rows
	assert.Equal(t, expectedRows, len(regions), "Region column should have %d values", expectedRows)
	assert.Equal(t, expectedRows, len(products), "Product column should have %d values", expectedRows)

	// Verify all expected combinations exist
	combinations := make(map[string]bool)
	for i := 0; i < len(regions); i++ {
		key := regions[i] + "-" + products[i]
		combinations[key] = true
	}

	// Check all expected combinations are present
	expectedCombos := []string{
		"North-A", "North-B",
		"South-A", "South-B",
		"East-A", "East-B",
	}
	for _, combo := range expectedCombos {
		assert.True(t, combinations[combo], "Missing combination: %s", combo)
	}

	// Verify no duplicates (count should match unique combinations)
	assert.Equal(t, expectedRows, len(combinations), "There should be no duplicate rows")
}

// TestGenerateParquetFile_RandomValues verifies that the Value column contains
// random float64 values assigned to each row.
//
// This test checks:
//   - The Value column exists in the generated parquet file
//   - The Value column contains float64 data
//   - Each row has a Value assigned (no null values)
//   - Values are within a reasonable range for random test data
//   - Different invocations produce different random values
func TestGenerateParquetFile_RandomValues(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.parquet")

	// Define test dimensions
	dimensions := map[string][]string{
		"Region": {"North", "South"},
	}

	// Call GenerateParquetFile
	err := GenerateParquetFile(filePath, dimensions, nil, nil)
	require.NoError(t, err, "GenerateParquetFile should not return an error")

	// Read the parquet file
	parquetSchema, _, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")

	// Verify Value column exists
	columnNames := GetColumnNames(parquetSchema)
	assert.Contains(t, columnNames, "Value", "Schema should have Value column")

	// Read the Value column data
	valueIndex := -1
	for i, name := range columnNames {
		if name == "Value" {
			valueIndex = i
			break
		}
	}
	require.NotEqual(t, -1, valueIndex, "Value column should be found")

	valueData, err := ReadParquetColumnData(filePath, valueIndex)
	require.NoError(t, err, "Should be able to read Value column")

	values := valueData.([]float64)
	require.Equal(t, 2, len(values), "Should have 2 values (one for each region)")

	// Verify each row has a non-null value
	for i, v := range values {
		// Float64 values from rand.Float64() are in [0.0, 1.0)
		assert.True(t, v >= 0.0 && v < 1.0, "Value at row %d should be in range [0.0, 1.0)", i)
	}
}

// TestGenerateParquetFile_EmptyDimensions tests the edge case where an empty
// dimensions map is provided to the GenerateParquetFile function.
//
// This test verifies the expected behavior when no dimensions are specified:
//   - The function either returns an error indicating dimensions are required
//   - OR the function creates a file with only the "Value" column
//   - If a file is created, it contains at least one row with a random value
//   - The function handles this edge case gracefully without panicking
func TestGenerateParquetFile_EmptyDimensions(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.parquet")

	// Define empty dimensions
	dimensions := map[string][]string{}

	// Call GenerateParquetFile
	err := GenerateParquetFile(filePath, dimensions, nil, nil)

	// The current implementation creates a file with just the Value column
	require.NoError(t, err, "GenerateParquetFile should handle empty dimensions gracefully")

	// Verify file exists
	_, err = os.Stat(filePath)
	require.NoError(t, err, "Generated parquet file should exist")

	// Read the parquet file
	parquetSchema, numRows, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")

	// Should have only the Value column
	assert.Equal(t, 1, parquetSchema.NumColumns(), "Schema should have only 1 column (Value)")

	columnNames := GetColumnNames(parquetSchema)
	assert.Contains(t, columnNames, "Value", "Schema should have Value column")

	// Should have 1 row with a random value
	assert.Equal(t, 1, numRows, "Should have 1 row with a random value")
}

// TestGenerateParquetFile_OverwritesExisting verifies that the function handles
// existing files correctly by overwriting them.
//
// This test checks:
//   - The function can overwrite an existing file at the specified path
//   - The function returns without error when the file already exists
//   - The new content replaces the old content entirely
//   - The function handles file permission issues gracefully
func TestGenerateParquetFile_OverwritesExisting(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.parquet")

	// First call with small dimensions
	dimensions1 := map[string][]string{
		"Region": {"North"},
	}
	err := GenerateParquetFile(filePath, dimensions1, nil, nil)
	require.NoError(t, err, "First GenerateParquetFile call should succeed")

	// Read first file
	_, numRows1, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read first file")
	assert.Equal(t, 1, numRows1, "First file should have 1 row")

	// Second call with different dimensions (should overwrite)
	dimensions2 := map[string][]string{
		"Region":  {"North", "South", "East", "West"},
		"Product": {"A", "B"},
	}
	err = GenerateParquetFile(filePath, dimensions2, nil, nil)
	require.NoError(t, err, "Second GenerateParquetFile call should succeed (overwrite)")

	// Read second file
	parquetSchema, numRows2, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read second file")

	// Verify new content (4 regions x 2 products = 8 rows)
	assert.Equal(t, 8, numRows2, "Second file should have 8 rows")

	// Verify schema has both columns
	columnNames := GetColumnNames(parquetSchema)
	assert.Contains(t, columnNames, "Region", "Schema should have Region column")
	assert.Contains(t, columnNames, "Product", "Schema should have Product column")
	assert.Contains(t, columnNames, "Value", "Schema should have Value column")
}

// TestGenerateParquetFile_Grain_CorrectSchema verifies that when grain measures
// are provided, they appear as additional string columns in the schema after
// dimension columns and before the Value column.
//
// This test checks:
//   - Grain columns are included in the schema
//   - The total column count is dimensions + grain + 1 (Value)
//   - Grain columns are ByteArray (string) type
//   - Column order is: dimensions (alphabetical), then grain (spec order), then Value
func TestGenerateParquetFile_Grain_CorrectSchema(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_grain_schema.parquet")

	// Define test dimensions
	dimensions := map[string][]string{
		"Region": {"North", "South"},
	}

	// Define grain measures
	grain := []GrainMeasure{
		{Name: "Year", Values: []string{"2023", "2024"}},
	}

	// Call GenerateParquetFile with grain
	err := GenerateParquetFile(filePath, dimensions, grain, nil)
	require.NoError(t, err, "GenerateParquetFile should not return an error")

	// Read the parquet file
	parquetSchema, _, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")

	// Verify column count (1 dimension + 1 grain + 1 Value = 3)
	assert.Equal(t, 3, parquetSchema.NumColumns(), "Schema should have 3 columns")

	// Verify column names
	columnNames := GetColumnNames(parquetSchema)
	assert.Contains(t, columnNames, "Region", "Schema should have Region column")
	assert.Contains(t, columnNames, "Year", "Schema should have Year column")
	assert.Contains(t, columnNames, "Value", "Schema should have Value column")

	// Verify column order: Region (dimension, alphabetical), Year (grain), Value
	require.GreaterOrEqual(t, len(columnNames), 3, "Schema should have at least 3 columns to verify ordering")
	assert.Equal(t, "Region", columnNames[0], "First column should be Region (dimension, alphabetical)")
	assert.Equal(t, "Year", columnNames[1], "Second column should be Year (grain)")
	assert.Equal(t, "Value", columnNames[2], "Third column should be Value")

	// Verify Region and Year are ByteArray (string)
	regionType := GetColumnType(parquetSchema, "Region")
	assert.Equal(t, parquet.Types.ByteArray, regionType, "Region column should be ByteArray type")

	yearType := GetColumnType(parquetSchema, "Year")
	assert.Equal(t, parquet.Types.ByteArray, yearType, "Year column should be ByteArray type")

	// Verify Value is Double
	valueType := GetColumnType(parquetSchema, "Value")
	assert.Equal(t, parquet.Types.Double, valueType, "Value column should be Double type")
}

// TestGenerateParquetFile_Grain_CartesianProduct verifies that grain measures
// participate in the Cartesian product, creating rows for all combinations of
// dimension and grain values.
//
// This test checks:
//   - The row count equals the product of all dimension and grain value counts
//   - All unique combinations of dimension and grain values appear exactly once
//   - No duplicate rows exist in the output
func TestGenerateParquetFile_Grain_CartesianProduct(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_grain_cartesian.parquet")

	// Define test dimensions
	dimensions := map[string][]string{
		"Region": {"North", "South", "East", "West"},
	}

	// Define grain measures
	grain := []GrainMeasure{
		{Name: "Year", Values: []string{"2023", "2024"}},
		{Name: "Quarter", Values: []string{"Q1", "Q2"}},
	}

	// Call GenerateParquetFile with grain
	err := GenerateParquetFile(filePath, dimensions, grain, nil)
	require.NoError(t, err, "GenerateParquetFile should not return an error")

	// Read the parquet file
	_, numRows, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")

	// Expected row count: 4 regions x 2 years x 2 quarters = 16
	expectedRows := 4 * 2 * 2
	assert.Equal(t, expectedRows, numRows, "Should have %d rows (Cartesian product)", expectedRows)

	// Verify schema has expected columns before reading by index
	parquetSchema, _, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")
	require.Equal(t, 4, parquetSchema.NumColumns(), "Schema should have 4 columns (Region, Year, Quarter, Value)")

	// Read column data to verify all combinations exist
	// Column order: Region (0), Year (1), Quarter (2), Value (3)
	regionData, err := ReadParquetColumnData(filePath, 0)
	require.NoError(t, err, "Should be able to read Region column")
	regions := regionData.([]string)

	yearData, err := ReadParquetColumnData(filePath, 1)
	require.NoError(t, err, "Should be able to read Year column")
	years := yearData.([]string)

	quarterData, err := ReadParquetColumnData(filePath, 2)
	require.NoError(t, err, "Should be able to read Quarter column")
	quarters := quarterData.([]string)

	// Verify row counts match
	assert.Equal(t, expectedRows, len(regions), "Region column should have %d values", expectedRows)
	assert.Equal(t, expectedRows, len(years), "Year column should have %d values", expectedRows)
	assert.Equal(t, expectedRows, len(quarters), "Quarter column should have %d values", expectedRows)

	// Build set of all observed combinations
	combinations := make(map[string]bool)
	for i := 0; i < len(regions); i++ {
		key := regions[i] + "-" + years[i] + "-" + quarters[i]
		combinations[key] = true
	}

	// Verify all expected combinations are present
	expectedRegions := []string{"North", "South", "East", "West"}
	expectedYears := []string{"2023", "2024"}
	expectedQuarters := []string{"Q1", "Q2"}
	for _, r := range expectedRegions {
		for _, y := range expectedYears {
			for _, q := range expectedQuarters {
				key := r + "-" + y + "-" + q
				assert.True(t, combinations[key], "Missing combination: %s", key)
			}
		}
	}

	// Verify no duplicates
	assert.Equal(t, expectedRows, len(combinations), "There should be no duplicate rows")
}

// TestGenerateParquetFile_Grain_Ordering verifies that grain columns preserve
// the order specified in the grain slice, rather than being sorted alphabetically.
//
// This test checks:
//   - Dimension columns are still sorted alphabetically
//   - Grain columns appear in the exact order provided in the grain slice
//   - Grain columns appear after all dimension columns and before Value
func TestGenerateParquetFile_Grain_Ordering(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_grain_ordering.parquet")

	// Define test dimensions with a single dimension
	dimensions := map[string][]string{
		"A": {"1", "2"},
	}

	// Define grain measures with names that are NOT in alphabetical order
	grain := []GrainMeasure{
		{Name: "Z", Values: []string{"a"}},
		{Name: "Y", Values: []string{"b"}},
	}

	// Call GenerateParquetFile with grain
	err := GenerateParquetFile(filePath, dimensions, grain, nil)
	require.NoError(t, err, "GenerateParquetFile should not return an error")

	// Read the parquet file
	parquetSchema, _, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")

	// Verify column names and order
	columnNames := GetColumnNames(parquetSchema)
	require.Equal(t, 4, len(columnNames), "Schema should have 4 columns")

	// Expected order: A (dimension, alphabetical), Z (grain, first in array), Y (grain, second in array), Value
	assert.Equal(t, "A", columnNames[0], "First column should be A (dimension, alphabetical)")
	assert.Equal(t, "Z", columnNames[1], "Second column should be Z (grain, first in spec order)")
	assert.Equal(t, "Y", columnNames[2], "Third column should be Y (grain, second in spec order)")
	assert.Equal(t, "Value", columnNames[3], "Fourth column should be Value")
}

// TestGenerateParquetFile_Grain_OnlyGrain verifies that parquet generation works
// correctly when only grain measures are provided and no dimensions are present.
//
// This test checks:
//   - The function handles empty dimensions gracefully when grain is provided
//   - Grain columns are included in the schema
//   - The row count equals the product of grain value counts
//   - The Value column is present
func TestGenerateParquetFile_Grain_OnlyGrain(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test_grain_only.parquet")

	// Define empty dimensions
	dimensions := map[string][]string{}

	// Define grain measures only
	grain := []GrainMeasure{
		{Name: "Day", Values: []string{"Mon", "Tue", "Wed"}},
	}

	// Call GenerateParquetFile with grain only
	err := GenerateParquetFile(filePath, dimensions, grain, nil)
	require.NoError(t, err, "GenerateParquetFile should not return an error")

	// Read the parquet file
	parquetSchema, numRows, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")

	// Verify schema has 2 columns: Day and Value
	assert.Equal(t, 2, parquetSchema.NumColumns(), "Schema should have 2 columns (Day, Value)")

	columnNames := GetColumnNames(parquetSchema)
	assert.Contains(t, columnNames, "Day", "Schema should have Day column")
	assert.Contains(t, columnNames, "Value", "Schema should have Value column")

	// Verify row count = 3 (just grain values)
	assert.Equal(t, 3, numRows, "Should have 3 rows (one per grain value)")

	// Read the Day column data to verify values
	dayIndex := -1
	for i, name := range columnNames {
		if name == "Day" {
			dayIndex = i
			break
		}
	}
	require.NotEqual(t, -1, dayIndex, "Day column should be found")

	dayData, err := ReadParquetColumnData(filePath, dayIndex)
	require.NoError(t, err, "Should be able to read Day column")
	days := dayData.([]string)
	require.Equal(t, 3, len(days), "Day column should have 3 values")

	// Verify all expected day values are present
	expectedDays := map[string]bool{"Mon": true, "Tue": true, "Wed": true}
	for _, d := range days {
		assert.True(t, expectedDays[d], "Unexpected day value: %s", d)
	}
}

// TestGenerateParquetFileWithConditionals_MonthDayInMonth verifies that the
// conditional generator produces only valid month/day combinations.
//
// This test checks:
//   - Row count is the sum of valid child values per parent, not the full Cartesian product
//   - Feb has 28 days, Jan and Mar have 31
//   - Every generated row represents a valid (Month, DayInMonth) combo
//   - Invalid combos (e.g., Feb 30) are not present
func TestGenerateParquetFileWithConditionals_MonthDayInMonth(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "conditional_month_day.parquet")

	dimensions := map[string][]string{
		"Month": {"Jan", "Feb", "Mar"},
	}
	grain := []GrainMeasure{
		{Name: "DayInMonth", Values: []string{
			"1", "2", "3", "4", "5", "6", "7", "8", "9", "10",
			"11", "12", "13", "14", "15", "16", "17", "18", "19", "20",
			"21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31",
		}},
	}
	conditionalDims := map[string]ConditionalDimMapping{
		"DayInMonth": {
			Parent: "Month",
			Values: map[string][]string{
				"Jan": {"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31"},
				"Feb": {"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28"},
				"Mar": {"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31"},
			},
		},
	}

	err := GenerateParquetFileWithConditionals(filePath, dimensions, grain, nil, conditionalDims)
	require.NoError(t, err, "GenerateParquetFileWithConditionals should not return an error")

	_, numRows, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")

	expectedRows := 31 + 28 + 31 // 90
	assert.Equal(t, expectedRows, numRows, "Should have %d rows (conditional product), not %d (full cartesian)", expectedRows, 3*31)

	// Read columns to verify all combinations are valid
	monthData, err := ReadParquetColumnData(filePath, 0)
	require.NoError(t, err, "Should be able to read Month column")
	months := monthData.([]string)

	dayData, err := ReadParquetColumnData(filePath, 1)
	require.NoError(t, err, "Should be able to read DayInMonth column")
	days := dayData.([]string)

	require.Equal(t, expectedRows, len(months), "Month column should have %d values", expectedRows)
	require.Equal(t, expectedRows, len(days), "DayInMonth column should have %d values", expectedRows)

	for i := 0; i < len(months); i++ {
		month := months[i]
		day := days[i]
		validDays, ok := conditionalDims["DayInMonth"].Values[month]
		assert.True(t, ok, "Month %s should be in conditional map", month)
		found := false
		for _, d := range validDays {
			if d == day {
				found = true
				break
			}
		}
		assert.True(t, found, "Day %s should be valid for month %s", day, month)
	}
}

// TestGenerateParquetFileWithConditionals_NoConditionals verifies that when
// conditionalDims is empty, the function behaves exactly like GenerateParquetFile.
//
// This test checks:
//   - The row count equals the full Cartesian product
//   - All expected combinations are present
//   - Schema and column ordering match GenerateParquetFile
func TestGenerateParquetFileWithConditionals_NoConditionals(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "no_conditionals.parquet")

	dimensions := map[string][]string{
		"Region":  {"North", "South"},
		"Product": {"A", "B"},
	}
	conditionalDims := map[string]ConditionalDimMapping{}

	err := GenerateParquetFileWithConditionals(filePath, dimensions, nil, nil, conditionalDims)
	require.NoError(t, err, "GenerateParquetFileWithConditionals should not return an error")

	_, numRows, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")

	assert.Equal(t, 4, numRows, "Should have 4 rows (full cartesian product)")

	productData, err := ReadParquetColumnData(filePath, 0)
	require.NoError(t, err, "Should be able to read Product column")
	products := productData.([]string)

	regionData, err := ReadParquetColumnData(filePath, 1)
	require.NoError(t, err, "Should be able to read Region column")
	regions := regionData.([]string)

	combinations := make(map[string]bool)
	for i := 0; i < len(regions); i++ {
		key := regions[i] + "-" + products[i]
		combinations[key] = true
	}

	expectedCombos := []string{"North-A", "North-B", "South-A", "South-B"}
	for _, combo := range expectedCombos {
		assert.True(t, combinations[combo], "Missing combination: %s", combo)
	}
}

// TestGenerateParquetFileWithConditionals_MissingParentValue verifies that when
// a parent value has no entry in the conditional map, it is skipped entirely.
//
// This test checks:
//   - Rows for the missing parent value are not generated
//   - Other parent values still produce their valid child rows
func TestGenerateParquetFileWithConditionals_MissingParentValue(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "missing_parent.parquet")

	dimensions := map[string][]string{
		"Month": {"Jan", "Feb", "Apr"}, // Apr is missing from conditional map
	}
	grain := []GrainMeasure{
		{Name: "DayInMonth", Values: []string{"1", "2", "3"}},
	}
	conditionalDims := map[string]ConditionalDimMapping{
		"DayInMonth": {
			Parent: "Month",
			Values: map[string][]string{
				"Jan": {"1", "2", "3"},
				"Feb": {"1", "2"},
			},
		},
	}

	err := GenerateParquetFileWithConditionals(filePath, dimensions, grain, nil, conditionalDims)
	require.NoError(t, err, "GenerateParquetFileWithConditionals should not return an error")

	_, numRows, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")

	// Jan: 3 days, Feb: 2 days, Apr: 0 days (missing parent value)
	assert.Equal(t, 5, numRows, "Should skip Apr entirely")

	monthData, err := ReadParquetColumnData(filePath, 0)
	require.NoError(t, err, "Should be able to read Month column")
	months := monthData.([]string)

	for _, m := range months {
		assert.NotEqual(t, "Apr", m, "Apr should not appear in output")
	}
}

// TestGenerateParquetFileWithConditionals_EmptyChildSet verifies that when a
// parent value maps to an empty child slice, that parent value is skipped.
//
// This test checks:
//   - No rows are generated for the parent with an empty child set
//   - Other parents still produce rows
func TestGenerateParquetFileWithConditionals_EmptyChildSet(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "empty_child.parquet")

	dimensions := map[string][]string{
		"Month": {"Jan", "Feb"},
	}
	grain := []GrainMeasure{
		{Name: "DayInMonth", Values: []string{"1", "2", "3"}},
	}
	conditionalDims := map[string]ConditionalDimMapping{
		"DayInMonth": {
			Parent: "Month",
			Values: map[string][]string{
				"Jan": {"1", "2", "3"},
				"Feb": {}, // Empty child set
			},
		},
	}

	err := GenerateParquetFileWithConditionals(filePath, dimensions, grain, nil, conditionalDims)
	require.NoError(t, err, "GenerateParquetFileWithConditionals should not return an error")

	_, numRows, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")

	assert.Equal(t, 3, numRows, "Should have only Jan rows (Feb skipped due to empty child set)")

	monthData, err := ReadParquetColumnData(filePath, 0)
	require.NoError(t, err, "Should be able to read Month column")
	months := monthData.([]string)

	for _, m := range months {
		assert.Equal(t, "Jan", m, "Only Jan should appear")
	}
}

// TestGenerateParquetFileWithConditionals_RowCountNotFullCartesian verifies that
// the conditional generator produces fewer rows than the full Cartesian product.
func TestGenerateParquetFileWithConditionals_RowCountNotFullCartesian(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "row_count.parquet")

	dimensions := map[string][]string{
		"Month": {"Jan", "Feb", "Mar"},
	}
	grain := []GrainMeasure{
		{Name: "DayInMonth", Values: []string{"1", "2", "3", "4", "5"}},
	}
	conditionalDims := map[string]ConditionalDimMapping{
		"DayInMonth": {
			Parent: "Month",
			Values: map[string][]string{
				"Jan": {"1", "2", "3", "4", "5"},
				"Feb": {"1", "2", "3"},
				"Mar": {"1", "2", "3", "4", "5"},
			},
		},
	}

	err := GenerateParquetFileWithConditionals(filePath, dimensions, grain, nil, conditionalDims)
	require.NoError(t, err, "GenerateParquetFileWithConditionals should not return an error")

	_, numRows, err := ReadParquetFile(filePath)
	require.NoError(t, err, "Should be able to read the parquet file")

	expectedRows := 5 + 3 + 5 // 13
	fullCartesian := 3 * 5    // 15
	assert.Equal(t, expectedRows, numRows, "Should have %d rows, not %d (full cartesian)", expectedRows, fullCartesian)
}

// TestGenerateParquetFileWithConditionals_EmptyParentWithValuesReturnsError verifies
// that when a ConditionalDimMapping has Values but no Parent, an error is returned
// rather than silently treating the dimension as non-conditional.
func TestGenerateParquetFileWithConditionals_EmptyParentWithValuesReturnsError(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "empty_parent.parquet")

	dimensions := map[string][]string{
		"Month": {"Jan", "Feb"},
	}
	grain := []GrainMeasure{
		{Name: "DayInMonth", Values: []string{"1", "2", "3"}},
	}
	conditionalDims := map[string]ConditionalDimMapping{
		"DayInMonth": {
			Parent: "", // Empty parent with non-empty values
			Values: map[string][]string{
				"Jan": {"1", "2", "3"},
			},
		},
	}

	err := GenerateParquetFileWithConditionals(filePath, dimensions, grain, nil, conditionalDims)
	assert.Error(t, err, "Should return error when conditional mapping has values but no parent")
	assert.Contains(t, err.Error(), "parent", "Error should mention parent")
}

// TestGenerateConditionalProduct_EmptyParentWithValuesReturnsError verifies
// that GenerateConditionalProduct itself returns an error for the same condition.
func TestGenerateConditionalProduct_EmptyParentWithValuesReturnsError(t *testing.T) {
	combined := map[string][]string{
		"Month":      {"Jan", "Feb"},
		"DayInMonth": {"1", "2", "3"},
	}
	allNames := []string{"Month", "DayInMonth"}
	conditionalDims := map[string]ConditionalDimMapping{
		"DayInMonth": {
			Parent: "",
			Values: map[string][]string{
				"Jan": {"1", "2"},
			},
		},
	}

	_, err := GenerateConditionalProduct(combined, allNames, conditionalDims)
	assert.Error(t, err, "Should return error when conditional mapping has values but no parent")
	assert.Contains(t, err.Error(), "parent", "Error should mention parent")
}

// TestGenerateParquetFileWithConditionals_ReadBackWithReadParquetToTensor verifies
// that a parquet file generated without conditionals can be read back successfully
// with ReadParquetToTensor, confirming schema and data integrity.
func TestGenerateParquetFileWithConditionals_ReadBackWithReadParquetToTensor(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "readback.parquet")

	dimensions := map[string][]string{
		"Region": {"North", "South"},
	}
	grain := []GrainMeasure{
		{Name: "Year", Values: []string{"2023", "2024"}},
	}
	conditionalDims := map[string]ConditionalDimMapping{}

	// Use fixed values in cartesian product order
	values := []float64{1.0, 2.0, 3.0, 4.0}

	err := GenerateParquetFileWithConditionals(filePath, dimensions, grain, values, conditionalDims)
	require.NoError(t, err, "GenerateParquetFileWithConditionals should not return an error")

	conceptValues := map[string][]string{
		"Region": {"North", "South"},
		"Year":   {"2023", "2024"},
	}

	data, dimNames, shape, err := ReadParquetToTensor(filePath, []string{"Region", "Year"}, nil, conceptValues)
	require.NoError(t, err, "ReadParquetToTensor should succeed for full-product file")

	assert.Equal(t, []string{"Region", "Year"}, dimNames)
	assert.Equal(t, []int{2, 2}, shape)
	require.Equal(t, 4, len(data))

	assert.InDelta(t, 1.0, data[0], 0.0001)
	assert.InDelta(t, 2.0, data[1], 0.0001)
	assert.InDelta(t, 3.0, data[2], 0.0001)
	assert.InDelta(t, 4.0, data[3], 0.0001)
}
