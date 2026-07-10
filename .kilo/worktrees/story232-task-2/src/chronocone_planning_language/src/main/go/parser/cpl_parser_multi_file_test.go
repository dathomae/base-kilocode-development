// Package parser_test contains tests for multi-file CPL parsing.
package parser_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chronocone/chronocone_planning_language/src/main/go/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseFiles_MultipleValidFiles parses multiple valid CPL files
// and verifies they are merged into a single compilation unit.
func TestParseFiles_MultipleValidFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create first CPL file with types
	file1 := filepath.Join(tempDir, "types.cpl")
	content1 := "types\n\tCurrency = Number(precision=2)\n"
	require.NoError(t, os.WriteFile(file1, []byte(content1), 0644))

	// Create second CPL file with concepts
	file2 := filepath.Join(tempDir, "concepts.cpl")
	content2 := "concepts\n\tCountry=[Values[USA, Canada]]\n"
	require.NoError(t, os.WriteFile(file2, []byte(content2), 0644))

	config := &parser.ParseLoggingConfig{
		LogLevel:   0,
		OutputFile: "",
	}

	rootTable, err := parser.ParseFiles([]string{file1, file2}, config)
	require.NoError(t, err)
	require.NotNil(t, rootTable)

	// Verify both types and concepts sections are present
	require.Len(t, rootTable.Children, 2, "expected two children (types and concepts)")
}

// TestParseFiles_SingleFile behaves like the original ParseWithLogLevel.
func TestParseFiles_SingleFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	file := filepath.Join(tempDir, "single.cpl")
	content := "types\n\tSales = Number()\n"
	require.NoError(t, os.WriteFile(file, []byte(content), 0644))

	config := &parser.ParseLoggingConfig{
		LogLevel:   0,
		OutputFile: "",
	}

	rootTable, err := parser.ParseFiles([]string{file}, config)
	require.NoError(t, err)
	require.NotNil(t, rootTable)
}

// TestParseFiles_NilConfig returns an error.
func TestParseFiles_NilConfig(t *testing.T) {
	t.Parallel()

	_, err := parser.ParseFiles([]string{"some.cpl"}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "logging config cannot be nil")
}

// TestParseFiles_EmptyPaths returns an error when no files are found.
func TestParseFiles_EmptyPaths(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	config := &parser.ParseLoggingConfig{
		LogLevel:   0,
		OutputFile: "",
	}

	_, err := parser.ParseFiles([]string{tempDir}, config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no .cpl files found")
}

// TestParseFiles_NonExistentPath returns an error.
func TestParseFiles_NonExistentPath(t *testing.T) {
	t.Parallel()

	config := &parser.ParseLoggingConfig{
		LogLevel:   0,
		OutputFile: "",
	}

	_, err := parser.ParseFiles([]string{"/nonexistent/path/file.cpl"}, config)
	require.Error(t, err)
}

// TestParseFiles_SyntaxErrorIncludesFileName verifies that syntax errors
// in multi-file parsing include the source file name.
func TestParseFiles_SyntaxErrorIncludesFileName(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Valid file
	file1 := filepath.Join(tempDir, "valid.cpl")
	require.NoError(t, os.WriteFile(file1, []byte("types\n\tSales = Number()\n"), 0644))

	// Invalid file
	file2 := filepath.Join(tempDir, "invalid.cpl")
	require.NoError(t, os.WriteFile(file2, []byte("types\n\t= Number()\n"), 0644))

	config := &parser.ParseLoggingConfig{
		LogLevel:   0,
		OutputFile: "",
	}

	_, err := parser.ParseFiles([]string{file1, file2}, config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid.cpl", "error should contain the file name")
}

// TestParseDirectory_DiscoversAndParsesFiles verifies that ParseDirectory
// discovers all .cpl files in a directory and parses them.
func TestParseDirectory_DiscoversAndParsesFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create multiple CPL files
	file1 := filepath.Join(tempDir, "a.cpl")
	require.NoError(t, os.WriteFile(file1, []byte("types\n\tA = Number()\n"), 0644))

	file2 := filepath.Join(tempDir, "b.cpl")
	require.NoError(t, os.WriteFile(file2, []byte("concepts\n\tB=[Values[X]]\n"), 0644))

	// Create a non-.cpl file that should be ignored
	nonCPL := filepath.Join(tempDir, "readme.txt")
	require.NoError(t, os.WriteFile(nonCPL, []byte("not a cpl file"), 0644))

	config := &parser.ParseLoggingConfig{
		LogLevel:   0,
		OutputFile: "",
	}

	rootTable, err := parser.ParseDirectory(tempDir, config)
	require.NoError(t, err)
	require.NotNil(t, rootTable)
	require.Len(t, rootTable.Children, 2, "expected two children from discovered .cpl files")
}

// TestParseFiles_SourceLocationIncludesFileName verifies that source locations
// in the merged symbol table include the originating file name.
func TestParseFiles_SourceLocationIncludesFileName(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	file1 := filepath.Join(tempDir, "file1.cpl")
	content1 := "types\n\tCurrency = Number(precision=2)\n"
	require.NoError(t, os.WriteFile(file1, []byte(content1), 0644))

	config := &parser.ParseLoggingConfig{
		LogLevel:   0,
		OutputFile: "",
	}

	rootTable, err := parser.ParseFiles([]string{file1}, config)
	require.NoError(t, err)
	require.NotNil(t, rootTable)

	// Find the types scope and verify definitions in leaf scopes have file names
	require.NotEmpty(t, rootTable.Children)
	typesScope := rootTable.Children[0]
	require.NotEmpty(t, typesScope.Children, "expected type definitions in children of TypesScope")

	// At least one definition in a leaf TypeScope should have the file name set
	found := false
	for _, typeScope := range typesScope.Children {
		for _, def := range typeScope.Definitions {
			if def.Location.File == file1 {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	assert.True(t, found, "expected at least one definition to have the file name in its source location")
}

// TestParseFiles_CrossFileReferences verifies that definitions from one file
// can be referenced from another when parsed together.
func TestParseFiles_CrossFileReferences(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// File 1 defines a concept
	file1 := filepath.Join(tempDir, "concepts.cpl")
	content1 := "concepts\n\tCountry=[Values[USA, Canada, Mexico]]\n"
	require.NoError(t, os.WriteFile(file1, []byte(content1), 0644))

	// File 2 defines a connector that references the concept
	file2 := filepath.Join(tempDir, "connectors.cpl")
	content2 := "connectors\n\tSalesData = dimensions [Country] grain Day celltype Number\n"
	require.NoError(t, os.WriteFile(file2, []byte(content2), 0644))

	config := &parser.ParseLoggingConfig{
		LogLevel:   0,
		OutputFile: "",
	}

	rootTable, err := parser.ParseFiles([]string{file1, file2}, config)
	require.NoError(t, err)
	require.NotNil(t, rootTable)

	// Should have concepts and connectors scopes
	require.Len(t, rootTable.Children, 2, "expected concepts and connectors scopes")
}

// TestParseFiles_DuplicateDefinitionsAcrossFiles verifies that duplicate
// definitions across different files are preserved in the merged symbol table
// for the analysis phase to detect.
func TestParseFiles_DuplicateDefinitionsAcrossFiles(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Both files define the same concept
	file1 := filepath.Join(tempDir, "file1.cpl")
	require.NoError(t, os.WriteFile(file1, []byte("concepts\n\tCountry=[Values[USA]]\n"), 0644))

	file2 := filepath.Join(tempDir, "file2.cpl")
	require.NoError(t, os.WriteFile(file2, []byte("concepts\n\tCountry=[Values[Canada]]\n"), 0644))

	config := &parser.ParseLoggingConfig{
		LogLevel:   0,
		OutputFile: "",
	}

	rootTable, err := parser.ParseFiles([]string{file1, file2}, config)
	require.NoError(t, err)
	require.NotNil(t, rootTable)

	// Both definitions should be present in the merged table.
	// When multiple files define the same section, each file's section scope
	// becomes a child of the merged root.
	countryCount := 0
	for _, childScope := range rootTable.Children {
		for _, conceptScope := range childScope.Children {
			for _, def := range conceptScope.Definitions {
				if def.Name != nil && *def.Name == "Country" {
					countryCount++
				}
			}
		}
	}
	assert.Equal(t, 2, countryCount, "expected both Country definitions to be present")
}

// TestParseDirectory_NonExistentDirectory returns an error.
func TestParseDirectory_NonExistentDirectory(t *testing.T) {
	t.Parallel()

	config := &parser.ParseLoggingConfig{
		LogLevel:   0,
		OutputFile: "",
	}

	_, err := parser.ParseDirectory("/nonexistent/directory", config)
	require.Error(t, err)
}

// TestParseFiles_DirectoryInput verifies that directories in the filePaths
// are expanded and their .cpl files are included.
func TestParseFiles_DirectoryInput(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create a CPL file in the directory
	file1 := filepath.Join(tempDir, "test.cpl")
	require.NoError(t, os.WriteFile(file1, []byte("types\n\tX = Number()\n"), 0644))

	// Also pass a direct file path
	directFile := filepath.Join(t.TempDir(), "direct.cpl")
	require.NoError(t, os.WriteFile(directFile, []byte("concepts\n\tY=[Values[A]]\n"), 0644))

	config := &parser.ParseLoggingConfig{
		LogLevel:   0,
		OutputFile: "",
	}

	rootTable, err := parser.ParseFiles([]string{tempDir, directFile}, config)
	require.NoError(t, err)
	require.NotNil(t, rootTable)
	require.Len(t, rootTable.Children, 2, "expected two children from directory + direct file")
}

// TestParseFiles_ErrorMessageFormat verifies that combined errors from multiple
// files have a readable format that includes file names, lines, and columns.
func TestParseFiles_ErrorMessageFormat(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	file1 := filepath.Join(tempDir, "bad1.cpl")
	require.NoError(t, os.WriteFile(file1, []byte("types\n\t= Number()\n"), 0644))

	file2 := filepath.Join(tempDir, "bad2.cpl")
	require.NoError(t, os.WriteFile(file2, []byte("types\n\tSales =\n"), 0644))

	config := &parser.ParseLoggingConfig{
		LogLevel:   0,
		OutputFile: "",
	}

	_, err := parser.ParseFiles([]string{file1, file2}, config)
	require.Error(t, err)

	errMsg := err.Error()
	assert.True(t, strings.HasPrefix(errMsg, "parse error:"), "error should start with 'parse error:'")
	assert.Contains(t, errMsg, "bad1.cpl", "error should mention bad1.cpl")
	assert.Contains(t, errMsg, "bad2.cpl", "error should mention bad2.cpl")
}
